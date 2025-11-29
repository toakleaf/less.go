#!/usr/bin/env node
/**
 * Plugin Host for less.go
 *
 * This Node.js script acts as a bridge between the Go runtime and JavaScript plugins.
 * It communicates via stdin/stdout using a JSON-based IPC protocol.
 *
 * Protocol:
 * - Commands are received as JSON objects on stdin (one per line)
 * - Responses are sent as JSON objects on stdout (one per line)
 * - Each command has an "id" field for request/response correlation
 * - Responses include "success" boolean and either "result" or "error"
 */

const readline = require('readline');
const path = require('path');
const fs = require('fs');
const vm = require('vm');

// ============================================================================
// Synchronous Callback Protocol for On-Demand Variable Lookup
// ============================================================================
//
// When Go sends a function call with `useOnDemandLookup: true`, JavaScript
// can request variable values from Go using synchronous callbacks:
//
// 1. JS sends callback request via stdout: {"id": N, "callback": "lookupVariable", "data": {...}}
// 2. Go reads the request and processes it
// 3. Go sends response via stdin: {"id": N, "success": true, "result": {...}}
// 4. JS reads the response synchronously and continues
//
// This avoids serializing all 283+ frames with all variables upfront.
// ============================================================================

let callbackIdCounter = 1000000; // Start high to avoid collision with command IDs
const pendingCallbacks = new Map(); // Map of callback ID -> {resolve, reject}

/**
 * Send a synchronous callback request to Go and wait for the response.
 * This uses a blocking read from stdin to wait for Go's response.
 * @param {string} callbackName - The callback type (e.g., "lookupVariable")
 * @param {any} data - The callback data
 * @returns {any} The callback result
 */
function sendCallbackSync(callbackName, data) {
  const id = callbackIdCounter++;

  // Send the callback request via stdout
  const request = JSON.stringify({ id, callback: callbackName, data });
  process.stdout.write(request + '\n');

  if (process.env.LESS_GO_DEBUG) {
    console.error(`[plugin-host] Sent callback request: ${callbackName} id=${id}`);
  }

  // Read response synchronously from stdin
  // We need to read line by line until we get our response
  const response = readResponseSync(id);

  if (!response.success) {
    throw new Error(`Callback failed: ${response.error}`);
  }

  return response.result;
}

/**
 * Read a response synchronously from stdin.
 * This blocks until the response with the matching ID is received.
 * @param {number} expectedId - The response ID to wait for
 * @returns {Object} The response object
 */
function readResponseSync(expectedId) {
  const BUFFER_SIZE = 1024 * 64; // 64KB chunks
  let buffer = '';

  // Read from stdin (fd 0) synchronously
  while (true) {
    const chunk = Buffer.alloc(BUFFER_SIZE);
    let bytesRead;

    try {
      bytesRead = fs.readSync(0, chunk, 0, BUFFER_SIZE);
    } catch (e) {
      if (e.code === 'EAGAIN' || e.code === 'EWOULDBLOCK') {
        // No data available, try again
        continue;
      }
      throw e;
    }

    if (bytesRead === 0) {
      // EOF - stdin closed
      throw new Error('stdin closed while waiting for callback response');
    }

    buffer += chunk.toString('utf8', 0, bytesRead);

    // Check for complete lines
    let newlineIdx;
    while ((newlineIdx = buffer.indexOf('\n')) !== -1) {
      const line = buffer.substring(0, newlineIdx);
      buffer = buffer.substring(newlineIdx + 1);

      if (!line.trim()) continue;

      try {
        const response = JSON.parse(line);

        if (response.id === expectedId) {
          if (process.env.LESS_GO_DEBUG) {
            console.error(`[plugin-host] Got callback response for id=${expectedId}`);
          }
          return response;
        } else {
          // This is a response for a different ID or a new command
          // Queue it for later processing
          if (process.env.LESS_GO_DEBUG) {
            console.error(`[plugin-host] Got response for different id=${response.id}, expected ${expectedId}`);
          }
          queuedResponses.push(response);
        }
      } catch (e) {
        console.error(`[plugin-host] Failed to parse response: ${e.message}`);
      }
    }
  }
}

// Queue for responses received while waiting for a specific callback
const queuedResponses = [];

// ============================================================================
// Shared Memory Variable Buffer
// ============================================================================
// This buffer is used to read variable data that Go writes in binary format.
// It avoids JSON serialization for variable values.

let varBuffer = null; // File descriptor for the memory-mapped file
let varBufferData = null; // Buffer view of the mmap'd data

/**
 * Attach a shared memory buffer for variable data.
 * @param {number} id - Command ID
 * @param {Object} data - Buffer info { key, path, size }
 */
function handleAttachVarBuffer(id, data) {
  try {
    const { path, size } = data;

    // Open the file
    const fd = fs.openSync(path, 'r');

    // Read the buffer - we can't truly mmap in Node.js without native modules,
    // but we can read the file content which Go updates
    varBuffer = { fd, path, size };
    varBufferData = Buffer.alloc(size);

    sendResponse(id, true, { attached: true });
  } catch (e) {
    sendResponse(id, false, null, `Failed to attach var buffer: ${e.message}`);
  }
}

/**
 * Detach the variable buffer.
 * @param {number} id - Command ID
 */
function handleDetachVarBuffer(id) {
  try {
    if (varBuffer) {
      fs.closeSync(varBuffer.fd);
      varBuffer = null;
      varBufferData = null;
    }
    sendResponse(id, true, { detached: true });
  } catch (e) {
    sendResponse(id, false, null, `Failed to detach var buffer: ${e.message}`);
  }
}

/**
 * Read a variable value from shared memory.
 * Binary format:
 * [1 byte: important flag]
 * [1 byte: type]
 * [value data based on type...]
 *
 * Types:
 * 1 = Dimension (8 bytes float64 value + 4 bytes unit length + unit string)
 * 2 = Color (8 bytes r, g, b, alpha as float64s = 32 bytes)
 * 3 = Quoted (4 bytes length + string + 1 byte quote char)
 * 4 = Keyword (4 bytes length + string)
 *
 * @param {number} offset - Offset in buffer
 * @param {number} length - Length of data
 * @returns {Object} The variable value as a JavaScript object
 */
function readVariableFromSharedMemory(offset, length) {
  if (!varBuffer || !varBufferData) {
    return null;
  }

  try {
    // Re-read the portion of the file that was updated
    const bytesRead = fs.readSync(varBuffer.fd, varBufferData, offset, length, offset);
    if (bytesRead < length) {
      return null;
    }

    let pos = offset;

    // Read important flag
    const important = varBufferData[pos++] === 1;

    // Read type
    const type = varBufferData[pos++];

    let value;
    switch (type) {
      case 1: // Dimension
        {
          const val = readFloat64(varBufferData, pos);
          pos += 8;
          const unitLen = varBufferData.readUInt32LE(pos);
          pos += 4;
          const unit = varBufferData.toString('utf8', pos, pos + unitLen);
          pos += unitLen;
          value = {
            _type: 'Dimension',
            value: val,
            unit: { numerator: [unit], denominator: [] },
          };
        }
        break;

      case 2: // Color
        {
          const r = readFloat64(varBufferData, pos);
          pos += 8;
          const g = readFloat64(varBufferData, pos);
          pos += 8;
          const b = readFloat64(varBufferData, pos);
          pos += 8;
          const alpha = readFloat64(varBufferData, pos);
          pos += 8;
          value = {
            _type: 'Color',
            rgb: [r, g, b],
            alpha: alpha,
          };
        }
        break;

      case 3: // Quoted
        {
          const strLen = varBufferData.readUInt32LE(pos);
          pos += 4;
          const str = varBufferData.toString('utf8', pos, pos + strLen);
          pos += strLen;
          const quote = String.fromCharCode(varBufferData[pos++]);
          value = {
            _type: 'Quoted',
            value: str,
            quote: quote,
          };
        }
        break;

      case 4: // Keyword
        {
          const strLen = varBufferData.readUInt32LE(pos);
          pos += 4;
          const str = varBufferData.toString('utf8', pos, pos + strLen);
          value = {
            _type: 'Keyword',
            value: str,
          };
        }
        break;

      default:
        return null;
    }

    return { value, important };
  } catch (e) {
    if (process.env.LESS_GO_DEBUG) {
      console.error(`[plugin-host] Error reading from shared memory: ${e.message}`);
    }
    return null;
  }
}

/**
 * Read a float64 from a buffer in little-endian format.
 */
function readFloat64(buf, offset) {
  const low = buf.readUInt32LE(offset);
  const high = buf.readUInt32LE(offset + 4);
  const dataView = new DataView(new ArrayBuffer(8));
  dataView.setUint32(0, low, true);
  dataView.setUint32(4, high, true);
  return dataView.getFloat64(0, true);
}

// Import bindings
let bindings;
try {
  bindings = require('./bindings');
} catch (e) {
  // Bindings not available - will use built-in constructors
  bindings = null;
}

// Plugin state
const loadedPlugins = new Map();
const registeredFunctions = new Map(); // Legacy global registry (kept for compatibility)
const registeredVisitors = [];
const registeredPreProcessors = [];
const registeredPostProcessors = [];
const registeredFileManagers = [];

// Function scope stack for proper scoping of plugin functions
// Each scope is a Map of function name -> function
// The stack represents nested scopes: [root, child1, child2, ...]
// Function lookup traverses from current scope (end of array) to root
const functionScopeStack = [new Map()]; // Start with root scope

/**
 * Enter a new function scope (called when entering a ruleset/mixin)
 */
function enterFunctionScope() {
  functionScopeStack.push(new Map());
  const depth = functionScopeStack.length - 1;
  if (process.env.LESS_GO_DEBUG) {
    console.error(`[plugin-host] enterFunctionScope -> depth=${depth}`);
  }
  return depth;
}

/**
 * Exit the current function scope (called when exiting a ruleset/mixin)
 */
function exitFunctionScope() {
  if (functionScopeStack.length > 1) {
    functionScopeStack.pop();
  }
  const depth = functionScopeStack.length - 1;
  if (process.env.LESS_GO_DEBUG) {
    console.error(`[plugin-host] exitFunctionScope -> depth=${depth}`);
  }
  return depth;
}

/**
 * Get the current scope (top of stack)
 */
function getCurrentScope() {
  return functionScopeStack[functionScopeStack.length - 1];
}

/**
 * Add a function to the current scope
 */
function addFunctionToScope(name, fn) {
  const depth = functionScopeStack.length - 1;
  getCurrentScope().set(name, fn);
  // Also add to legacy global registry for backwards compatibility
  registeredFunctions.set(name, fn);
  if (process.env.LESS_GO_DEBUG) {
    console.error(`[plugin-host] addFunctionToScope: ${name} at depth=${depth}`);
  }
}

/**
 * Look up a function by name, traversing from current scope to root
 * Returns the function if found, undefined otherwise
 */
function lookupFunction(name) {
  const currentDepth = functionScopeStack.length - 1;
  // Search from current scope (end) to root (beginning)
  for (let i = functionScopeStack.length - 1; i >= 0; i--) {
    const fn = functionScopeStack[i].get(name);
    if (fn !== undefined) {
      if (process.env.LESS_GO_DEBUG) {
        console.error(`[plugin-host] lookupFunction: ${name} found at depth=${i}, current depth=${currentDepth}`);
      }
      return fn;
    }
  }
  if (process.env.LESS_GO_DEBUG) {
    console.error(`[plugin-host] lookupFunction: ${name} NOT FOUND, current depth=${currentDepth}`);
  }
  return undefined;
}

/**
 * Check if a function exists in any scope
 */
function hasFunction(name) {
  return lookupFunction(name) !== undefined;
}

// Shared memory state
const attachedBuffers = new Map(); // key -> { path, size, buffer }

// Node constructor helpers
// These create simple object representations that will be serialized back to Go
function createNode(type, props = {}) {
  return {
    _type: type,
    ...props,
  };
}

// Less API with node constructors
const less = {
  version: [4, 0, 0],

  // Node constructors (matching less.js tree API)
  dimension: (value, unit) => createNode('Dimension', { value, unit: unit || '' }),
  color: (rgb, alpha) => {
    // Handle both [r,g,b] array and {r,g,b} object
    if (Array.isArray(rgb)) {
      return createNode('Color', { rgb, alpha: alpha !== undefined ? alpha : 1 });
    }
    return createNode('Color', { rgb: [rgb.r || 0, rgb.g || 0, rgb.b || 0], alpha: alpha !== undefined ? alpha : 1 });
  },
  quoted: (quote, value, escaped) => createNode('Quoted', { quote, value, escaped: escaped || false }),
  keyword: (value) => createNode('Keyword', { value }),
  anonymous: (value) => createNode('Anonymous', { value }),
  url: (value, paths) => createNode('URL', { value, paths }),
  call: (name, args) => createNode('Call', { name, args: args || [] }),
  variable: (name) => createNode('Variable', { name }),
  value: (value) => createNode('Value', { value: Array.isArray(value) ? value : [value] }),
  expression: (value) => createNode('Expression', { value: Array.isArray(value) ? value : [value] }),
  operation: (op, operands) => createNode('Operation', { op, operands }),
  combinator: (value) => createNode('Combinator', { value }),
  element: (combinator, value) => createNode('Element', { combinator, value }),
  selector: (elements) => createNode('Selector', { elements: Array.isArray(elements) ? elements : [elements] }),
  ruleset: (selectors, rules) => createNode('Ruleset', { selectors, rules }),
  declaration: (name, value, important, merge, inline, variable) =>
    createNode('Declaration', { name, value, important, merge, inline, variable }),
  detachedruleset: (ruleset) => createNode('DetachedRuleset', { ruleset }),
  paren: (node) => createNode('Paren', { value: node }),
  negative: (node) => createNode('Negative', { value: node }),
  atrule: (name, value, rules, index, isRooted) =>
    createNode('AtRule', { name, value, rules, index, isRooted }),
  assignment: (key, val) => createNode('Assignment', { key, value: val }),
  attribute: (key, op, value) => createNode('Attribute', { key, op, value }),
  condition: (op, lvalue, rvalue, negate) => createNode('Condition', { op, lvalue, rvalue, negate }),

  // Visitor base class
  visitors: {
    Visitor: class Visitor {
      constructor(implementation) {
        this._implementation = implementation;
        this._visitFnCache = {};
      }

      visit(node) {
        if (!node) return node;

        const type = node._type || node.type;
        if (!type) return node;

        const funcName = 'visit' + type;
        if (this._implementation[funcName]) {
          return this._implementation[funcName](node);
        }
        return node;
      }
    },
  },

  // Tree namespace (for compatibility with plugins that use less.tree)
  tree: {
    Anonymous: function (value) {
      return createNode('Anonymous', { value });
    },
    Dimension: function (value, unit) {
      return createNode('Dimension', { value, unit: unit || '' });
    },
    Color: function (rgb, alpha) {
      if (Array.isArray(rgb)) {
        return createNode('Color', { rgb, alpha: alpha !== undefined ? alpha : 1 });
      }
      return createNode('Color', { rgb: [rgb.r || 0, rgb.g || 0, rgb.b || 0], alpha: alpha !== undefined ? alpha : 1 });
    },
    Quoted: function (quote, value, escaped) {
      return createNode('Quoted', { quote, value, escaped: escaped || false });
    },
    Keyword: function (value) {
      return createNode('Keyword', { value });
    },
    URL: function (value, paths) {
      return createNode('URL', { value, paths });
    },
    Call: function (name, args) {
      return createNode('Call', { name, args: args || [] });
    },
    Variable: function (name) {
      const node = createNode('Variable', { name });
      return node;
    },
    Value: function (value) {
      return createNode('Value', { value: Array.isArray(value) ? value : [value] });
    },
    Expression: function (value) {
      return createNode('Expression', { value: Array.isArray(value) ? value : [value] });
    },
    Operation: function (op, operands) {
      return createNode('Operation', { op, operands });
    },
    Combinator: function (value) {
      return createNode('Combinator', { value });
    },
    Element: function (combinator, value) {
      return createNode('Element', { combinator, value });
    },
    Selector: function (elements) {
      return createNode('Selector', { elements: Array.isArray(elements) ? elements : [elements] });
    },
    Ruleset: function (selectors, rules) {
      return createNode('Ruleset', { selectors, rules });
    },
    Declaration: function (name, value, important, merge, inline, variable) {
      return createNode('Declaration', { name, value, important, merge, inline, variable });
    },
    DetachedRuleset: function (ruleset) {
      return createNode('DetachedRuleset', { ruleset });
    },
    Paren: function (node) {
      return createNode('Paren', { value: node });
    },
    Negative: function (node) {
      return createNode('Negative', { value: node });
    },
    AtRule: function (name, value, rules, index, isRooted) {
      return createNode('AtRule', { name, value, rules, index, isRooted });
    },
    Assignment: function (key, val) {
      return createNode('Assignment', { key, value: val });
    },
    Attribute: function (key, op, value) {
      return createNode('Attribute', { key, op, value });
    },
    Condition: function (op, lvalue, rvalue, negate) {
      return createNode('Condition', { op, lvalue, rvalue, negate });
    },
  },
};

// Create global references for legacy plugins that use `tree` and `functions` directly
const tree = less.tree;

// Add Variable.prototype.find - used by bootstrap-less-port plugins to look up variables
// This iterates over frames and calls the callback until it finds a truthy result
tree.Variable.prototype.find = function (frames, callback) {
  for (let i = 0; i < frames.length; i++) {
    const result = callback(frames[i]);
    if (result) {
      return result;
    }
  }
  return null;
};

// Function registry - now uses scoped function management
const functionRegistry = {
  add(name, fn) {
    addFunctionToScope(name, fn);
  },
  addMultiple(functions) {
    for (const [name, fn] of Object.entries(functions)) {
      this.add(name, fn);
    }
  },
  get(name) {
    // Use scoped lookup for proper function resolution
    return lookupFunction(name);
  },
  getAll() {
    // Return all unique function names from all scopes
    const allNames = new Set();
    for (const scope of functionScopeStack) {
      for (const name of scope.keys()) {
        allNames.add(name);
      }
    }
    return Array.from(allNames);
  },
};

// Plugin manager mock
const pluginManager = {
  addVisitor(visitor) {
    registeredVisitors.push(visitor);
  },
  addPreProcessor(processor, priority = 1000) {
    registeredPreProcessors.push({ processor, priority });
    registeredPreProcessors.sort((a, b) => a.priority - b.priority);
  },
  addPostProcessor(processor, priority = 1000) {
    registeredPostProcessors.push({ processor, priority });
    registeredPostProcessors.sort((a, b) => a.priority - b.priority);
  },
  addFileManager(fileManager) {
    registeredFileManagers.push(fileManager);
  },
  getVisitors() {
    return registeredVisitors;
  },
  getPreProcessors() {
    return registeredPreProcessors.map((p) => p.processor);
  },
  getPostProcessors() {
    return registeredPostProcessors.map((p) => p.processor);
  },
  getFileManagers() {
    return registeredFileManagers;
  },
};

/**
 * Send a response to the Go runtime
 * @param {number} id - Command ID for correlation
 * @param {boolean} success - Whether the command succeeded
 * @param {*} result - Result data (if success)
 * @param {string} error - Error message (if !success)
 */
function sendResponse(id, success, result, error) {
  const response = { id, success };
  if (success) {
    response.result = result;
  } else {
    response.error = error;
  }
  process.stdout.write(JSON.stringify(response) + '\n');
}

/**
 * Handle incoming commands
 * @param {Object} cmd - Command object
 */
function handleCommand(cmd) {
  const { id, cmd: command, data } = cmd;

  try {
    switch (command) {
      case 'ping':
        sendResponse(id, true, 'pong');
        break;

      case 'echo':
        sendResponse(id, true, data);
        break;

      case 'shutdown':
        sendResponse(id, true, 'shutting down');
        // Exit after a brief delay to ensure response is sent
        setImmediate(() => process.exit(0));
        break;

      case 'loadPlugin':
        handleLoadPlugin(id, data);
        break;

      case 'callFunction':
        handleCallFunction(id, data);
        break;

      case 'getRegisteredFunctions':
        sendResponse(id, true, functionRegistry.getAll());
        break;

      case 'enterScope':
        // Enter a new function scope (for plugin function scoping)
        const scopeDepth = enterFunctionScope();
        sendResponse(id, true, { depth: scopeDepth });
        break;

      case 'exitScope':
        // Exit the current function scope
        const newDepth = exitFunctionScope();
        sendResponse(id, true, { depth: newDepth });
        break;

      case 'addFunctionToScope':
        // Add a function to the current scope (used for inheriting plugin functions from parent frames)
        handleAddFunctionToScope(id, data);
        break;

      case 'attachVarBuffer':
        // Attach a shared memory buffer for variable data
        handleAttachVarBuffer(id, data);
        break;

      case 'detachVarBuffer':
        // Detach the variable buffer
        handleDetachVarBuffer(id);
        break;

      case 'getVisitors':
        sendResponse(
          id,
          true,
          registeredVisitors.map((v, i) => ({
            index: i,
            isPreEvalVisitor: v.isPreEvalVisitor || false,
            isReplacing: v.isReplacing || false,
          }))
        );
        break;

      case 'attachBuffer':
        handleAttachBuffer(id, data);
        break;

      case 'detachBuffer':
        handleDetachBuffer(id, data);
        break;

      case 'readBuffer':
        handleReadBuffer(id, data);
        break;

      case 'getBufferInfo':
        handleGetBufferInfo(id, data);
        break;

      case 'runVisitor':
        handleRunVisitor(id, data);
        break;

      case 'runPreEvalVisitors':
        handleRunPreEvalVisitors(id, data);
        break;

      case 'runPreEvalVisitorsJSON':
        handleRunPreEvalVisitorsJSON(id, data);
        break;

      case 'checkVariableReplacements':
        handleCheckVariableReplacements(id, data);
        break;

      case 'runPostEvalVisitors':
        handleRunPostEvalVisitors(id, data);
        break;

      case 'parseASTBuffer':
        handleParseASTBuffer(id, data);
        break;

      case 'serializeNode':
        handleSerializeNode(id, data);
        break;

      case 'callFunctionSharedMem':
        handleCallFunctionSharedMem(id, data);
        break;

      case 'getPreProcessors':
        handleGetPreProcessors(id, data);
        break;

      case 'getPostProcessors':
        handleGetPostProcessors(id, data);
        break;

      case 'runPreProcessor':
        handleRunPreProcessor(id, data);
        break;

      case 'runPostProcessor':
        handleRunPostProcessor(id, data);
        break;

      case 'getFileManagers':
        handleGetFileManagers(id, data);
        break;

      case 'fileManagerSupports':
        handleFileManagerSupports(id, data);
        break;

      case 'fileManagerLoad':
        handleFileManagerLoad(id, data);
        break;

      default:
        sendResponse(id, false, null, `Unknown command: ${command}`);
    }
  } catch (err) {
    sendResponse(id, false, null, err.message || String(err));
  }
}

/**
 * Add a function to the current scope by name.
 * This is used when mixins inherit plugin functions from parent frames.
 * The function must already exist in the registry (loaded by a previous @plugin).
 * @param {number} id - Command ID
 * @param {Object} args - { name: string }
 */
function handleAddFunctionToScope(id, args) {
  const { name } = args || {};

  if (!name) {
    sendResponse(id, false, null, 'Function name is required');
    return;
  }

  // Look up the function from the legacy global registry
  // (functions are stored there when first loaded)
  const fn = registeredFunctions.get(name);
  if (!fn) {
    // Function not found in registry - this shouldn't happen normally
    // but we'll silently succeed since the function might be available at a different scope
    if (process.env.LESS_GO_DEBUG) {
      console.error(`[plugin-host] handleAddFunctionToScope: function '${name}' not found in registry`);
    }
    sendResponse(id, true, { added: false, name, reason: 'not found in registry' });
    return;
  }

  // Add the function to the current scope
  addFunctionToScope(name, fn);
  sendResponse(id, true, { added: true, name, depth: functionScopeStack.length - 1 });
}

/**
 * Load a plugin from a file path
 * @param {number} id - Command ID
 * @param {Object} data - Plugin data
 */
function handleLoadPlugin(id, data) {
  const { path: pluginPath, options, baseDir } = data || {};

  if (!pluginPath) {
    sendResponse(id, false, null, 'Plugin path is required');
    return;
  }

  try {
    // Resolve the plugin path using Node.js require resolution
    let resolvedPath;
    let plugin;

    // Try to resolve the plugin path
    const isRelative = pluginPath.startsWith('.') || pluginPath.startsWith('/');
    const isAbsolute = path.isAbsolute(pluginPath);

    if (isAbsolute) {
      resolvedPath = pluginPath;
      // Try with .js extension if file doesn't exist
      if (!fs.existsSync(resolvedPath) && fs.existsSync(resolvedPath + '.js')) {
        resolvedPath = resolvedPath + '.js';
      }
    } else if (isRelative) {
      const basePath = baseDir || process.cwd();
      resolvedPath = path.resolve(basePath, pluginPath);
      // Try with .js extension if file doesn't exist
      if (!fs.existsSync(resolvedPath) && fs.existsSync(resolvedPath + '.js')) {
        resolvedPath = resolvedPath + '.js';
      }
    } else {
      // Not an explicit relative or absolute path
      // First, try to find in baseDir (common for @plugin "name" without ./)
      let foundInBaseDir = false;
      if (baseDir) {
        const baseDirPath = path.resolve(baseDir, pluginPath);
        if (fs.existsSync(baseDirPath)) {
          resolvedPath = baseDirPath;
          foundInBaseDir = true;
        } else if (fs.existsSync(baseDirPath + '.js')) {
          resolvedPath = baseDirPath + '.js';
          foundInBaseDir = true;
        }
      }

      // If not found in baseDir, try NPM module resolution
      if (!foundInBaseDir) {
        const searchPaths = [];
        if (baseDir) {
          searchPaths.push(baseDir);
          searchPaths.push(path.join(baseDir, 'node_modules'));
        }
        searchPaths.push(process.cwd());
        searchPaths.push(path.join(process.cwd(), 'node_modules'));

        try {
          resolvedPath = require.resolve(pluginPath, { paths: searchPaths });
        } catch (resolveErr) {
          throw new Error(`Cannot find module '${pluginPath}': ${resolveErr.message}`);
        }
      }
    }

    console.error(`[plugin-host] Checking cache for resolvedPath: ${resolvedPath}, has=${loadedPlugins.has(resolvedPath)}`);
    // Check if already loaded (by resolved path)
    if (loadedPlugins.has(resolvedPath)) {
      // IMPORTANT: Even when cached, we need to register the plugin's functions
      // in the CURRENT scope. This is essential for proper plugin scoping -
      // when the same plugin is loaded in different scopes (e.g., in mixins),
      // its functions should be available in that scope.
      const cachedPlugin = loadedPlugins.get(resolvedPath);
      console.error(`[plugin-host] Cache hit for ${resolvedPath}, functions: ${cachedPlugin && cachedPlugin.functions ? Object.keys(cachedPlugin.functions).join(', ') : 'none'}`);
      if (cachedPlugin && cachedPlugin.functions) {
        // Re-register all functions from the cached plugin in the current scope
        const funcNames = Object.keys(cachedPlugin.functions);
        for (const name of funcNames) {
          addFunctionToScope(name, cachedPlugin.functions[name]);
        }
      }

      sendResponse(id, true, {
        cached: true,
        path: resolvedPath,
        functions: functionRegistry.getAll(),
        visitors: registeredVisitors.length,
        preProcessors: registeredPreProcessors.length,
        postProcessors: registeredPostProcessors.length,
        fileManagers: registeredFileManagers.length,
      });
      return;
    }

    // Count functions before loading to track new ones
    const functionsBefore = registeredFunctions.size;
    const visitorsBefore = registeredVisitors.length;
    const preProcessorsBefore = registeredPreProcessors.length;
    const postProcessorsBefore = registeredPostProcessors.length;
    const fileManagersBefore = registeredFileManagers.length;

    // Track which functions were registered during this plugin load
    // This is needed for proper scope-aware caching
    const functionsBeforeLoad = new Map();
    for (const scope of functionScopeStack) {
      for (const [name, fn] of scope.entries()) {
        if (!functionsBeforeLoad.has(name)) {
          functionsBeforeLoad.set(name, fn);
        }
      }
    }

    // Set global references for legacy plugins that use `functions` and `tree` directly
    // These need to be set before require() is called so legacy plugins can use them
    global.functions = functionRegistry;
    global.tree = tree;
    global.less = less;

    // Try to load via require() first
    try {
      plugin = require(resolvedPath);
    } catch (requireErr) {
      // If require fails, try loading as raw JS code with vm
      const fileContent = fs.readFileSync(resolvedPath, 'utf8');

      // Create a registerPlugin function to capture plugins using the v3+ API
      let registeredPlugin = null;
      const registerPlugin = (pluginObj) => {
        registeredPlugin = pluginObj;
      };

      const context = vm.createContext({
        functions: functionRegistry,
        tree: tree,
        less: less,
        module: { exports: {} },
        exports: {},
        require: require,
        console: console,
        __filename: resolvedPath,
        __dirname: path.dirname(resolvedPath),
        registerPlugin: registerPlugin, // Add registerPlugin for v3+ API
      });

      vm.runInContext(fileContent, context, { filename: resolvedPath });

      // Check if plugin was registered via registerPlugin() API
      if (registeredPlugin) {
        plugin = registeredPlugin;
      } else {
        plugin = context.module.exports;
      }
    }

    // Handle plugin as either an object or a function (constructor)
    if (typeof plugin === 'function') {
      plugin = new plugin();
    }

    // Check if the plugin registered anything already (legacy format)
    // Legacy plugins call functions.add() directly during require()
    const hasLegacyRegistrations = registeredFunctions.size > functionsBefore;

    // Determine plugin version (used to decide order of setOptions vs install)
    // In less.js, the default behavior when options are provided is:
    // - If minVersion is specified as 3.x+: setOptions AFTER install
    // - Otherwise (no minVersion or minVersion < 3): setOptions BEFORE install
    let isV3Plus = false; // Default to pre-v3 behavior (setOptions before install)
    if (plugin && plugin.minVersion) {
      const minVer = Array.isArray(plugin.minVersion) ? plugin.minVersion : plugin.minVersion.split('.').map(Number);
      isV3Plus = (minVer[0] || 0) >= 3;

      // Validate minVersion
      const lessVer = less.version;
      for (let i = 0; i < minVer.length; i++) {
        if ((minVer[i] || 0) > (lessVer[i] || 0)) {
          throw new Error(`Plugin requires Less.js version ${minVer.join('.')} or higher`);
        } else if ((minVer[i] || 0) < (lessVer[i] || 0)) {
          break;
        }
      }
    }

    // For pre-v3 plugins (or no minVersion): setOptions is called BEFORE install
    if (!isV3Plus && options && plugin && plugin.setOptions) {
      plugin.setOptions(options);
    }

    // If no legacy registrations and plugin has an install method, call it
    if (!hasLegacyRegistrations && plugin && plugin.install) {
      plugin.install(less, pluginManager, functionRegistry);
    }

    // For v3+ plugins: setOptions is called AFTER install
    if (isV3Plus && options && plugin && plugin.setOptions) {
      plugin.setOptions(options);
    }

    // Call use() if it exists (called every time plugin is loaded)
    if (plugin && plugin.use) {
      plugin.use(plugin);
    }

    // Capture which functions were newly registered by this plugin
    // This enables proper re-registration when loading from cache in different scopes
    const pluginFunctions = {};
    const currentScope = functionScopeStack[functionScopeStack.length - 1];
    if (currentScope) {
      for (const [name, fn] of currentScope.entries()) {
        // Only capture functions that weren't present before loading
        if (!functionsBeforeLoad.has(name) || functionsBeforeLoad.get(name) !== fn) {
          pluginFunctions[name] = fn;
        }
      }
    }

    // Cache the plugin with its registered functions
    loadedPlugins.set(resolvedPath, {
      plugin: plugin || {},
      functions: pluginFunctions,
    });

    // Count new registrations
    const newFunctions = registeredFunctions.size - functionsBefore;
    const newVisitors = registeredVisitors.length - visitorsBefore;
    const newPreProcessors = registeredPreProcessors.length - preProcessorsBefore;
    const newPostProcessors = registeredPostProcessors.length - postProcessorsBefore;
    const newFileManagers = registeredFileManagers.length - fileManagersBefore;

    sendResponse(id, true, {
      cached: false,
      path: resolvedPath,
      functions: functionRegistry.getAll(),
      visitors: registeredVisitors.length,
      preProcessors: registeredPreProcessors.length,
      postProcessors: registeredPostProcessors.length,
      fileManagers: registeredFileManagers.length,
      newFunctions,
      newVisitors,
      newPreProcessors,
      newPostProcessors,
      newFileManagers,
    });
  } catch (err) {
    sendResponse(id, false, null, `Failed to load plugin: ${err.message}\n${err.stack || ''}`);
  }
}

/**
 * Convert a Go-serialized node argument to a JavaScript node object.
 * Go sends nodes as maps with _type and properties.
 * @param {*} arg - The argument from Go
 * @returns {*} The converted argument
 */
function convertGoNodeToJS(arg) {
  if (arg === null || arg === undefined) {
    return arg;
  }

  // Primitive types pass through
  if (typeof arg !== 'object') {
    return arg;
  }

  // Arrays need recursive conversion
  if (Array.isArray(arg)) {
    return arg.map(convertGoNodeToJS);
  }

  // Check if it's a typed node from Go
  const nodeType = arg._type;
  if (!nodeType) {
    // Plain object, convert values recursively
    const result = {};
    for (const [key, value] of Object.entries(arg)) {
      result[key] = convertGoNodeToJS(value);
    }
    return result;
  }

  // Create a node-like object that plugin functions can use
  // This matches the structure that less.js plugins expect
  const node = {
    _type: nodeType,
    type: nodeType,
  };

  // Copy all properties from the Go node
  for (const [key, value] of Object.entries(arg)) {
    if (key !== '_type') {
      node[key] = convertGoNodeToJS(value);
    }
  }

  // Add common getter methods that plugins might use
  Object.defineProperty(node, 'value', {
    get() {
      return this._value !== undefined ? this._value : (this.val !== undefined ? this.val : this._internalValue);
    },
    set(v) {
      this._internalValue = v;
    },
    enumerable: true,
    configurable: true,
  });

  // Initialize _internalValue from existing value property
  if (arg.value !== undefined) {
    node._internalValue = convertGoNodeToJS(arg.value);
  }

  return node;
}

/**
 * Convert a JavaScript result node to Go-compatible format.
 * @param {*} result - The JavaScript result
 * @returns {*} The Go-compatible result
 */
function convertJSResultToGo(result) {
  if (result === null || result === undefined) {
    return result;
  }

  // Primitive types pass through
  if (typeof result !== 'object') {
    return result;
  }

  // Arrays need recursive conversion
  if (Array.isArray(result)) {
    return result.map(convertJSResultToGo);
  }

  // Get the type from _type or type property
  const nodeType = result._type || result.type;
  if (!nodeType) {
    // Plain object, convert values recursively
    const converted = {};
    for (const [key, value] of Object.entries(result)) {
      converted[key] = convertJSResultToGo(value);
    }
    return converted;
  }

  // Create a clean node representation for Go
  const goNode = {
    _type: nodeType,
  };

  // Copy relevant properties based on node type
  switch (nodeType) {
    case 'Dimension':
      goNode.value = typeof result.value === 'number' ? result.value : parseFloat(result.value) || 0;
      goNode.unit = result.unit || '';
      break;

    case 'Color':
      goNode.rgb = result.rgb || [0, 0, 0];
      goNode.alpha = result.alpha !== undefined ? result.alpha : 1;
      break;

    case 'Quoted':
      goNode.value = result.value || '';
      goNode.quote = result.quote || '"';
      goNode.escaped = result.escaped || false;
      break;

    case 'Keyword':
      goNode.value = result.value || '';
      break;

    case 'Anonymous':
      goNode.value = result.value !== undefined ? String(result.value) : '';
      break;

    case 'URL':
      goNode.value = convertJSResultToGo(result.value);
      if (result.paths) {
        goNode.paths = result.paths;
      }
      break;

    case 'Expression':
    case 'Value':
      if (result.value) {
        goNode.value = Array.isArray(result.value)
          ? result.value.map(convertJSResultToGo)
          : [convertJSResultToGo(result.value)];
      }
      break;

    case 'Call':
      goNode.name = result.name || '';
      goNode.args = result.args ? result.args.map(convertJSResultToGo) : [];
      break;

    case 'Combinator':
      goNode.value = result.value || '';
      break;

    case 'Element':
      goNode.combinator = convertJSResultToGo(result.combinator);
      goNode.value = result.value || '';
      break;

    case 'Selector':
      goNode.elements = result.elements ? result.elements.map(convertJSResultToGo) : [];
      break;

    case 'Ruleset':
      goNode.selectors = result.selectors ? result.selectors.map(convertJSResultToGo) : [];
      goNode.rules = result.rules ? result.rules.map(convertJSResultToGo) : [];
      break;

    case 'Declaration':
      goNode.name = result.name || '';
      goNode.value = convertJSResultToGo(result.value);
      goNode.important = result.important || '';
      break;

    case 'DetachedRuleset':
      goNode.ruleset = convertJSResultToGo(result.ruleset);
      break;

    case 'AtRule':
      goNode.name = result.name || '';
      goNode.value = convertJSResultToGo(result.value);
      if (result.rules) {
        goNode.rules = result.rules.map(convertJSResultToGo);
      }
      break;

    case 'Operation':
      goNode.op = result.op || '';
      goNode.operands = result.operands ? result.operands.map(convertJSResultToGo) : [];
      break;

    case 'Condition':
      goNode.op = result.op || '';
      goNode.lvalue = convertJSResultToGo(result.lvalue);
      goNode.rvalue = convertJSResultToGo(result.rvalue);
      goNode.negate = result.negate || false;
      break;

    case 'Assignment':
      goNode.key = result.key || '';
      goNode.value = result.value;
      break;

    case 'Attribute':
      goNode.key = result.key || '';
      goNode.op = result.op || '';
      goNode.value = result.value;
      break;

    default:
      // For unknown types, copy all enumerable properties
      for (const [key, value] of Object.entries(result)) {
        if (key !== '_type' && key !== 'type') {
          goNode[key] = convertJSResultToGo(value);
        }
      }
  }

  return goNode;
}

/**
 * Add eval() and toCSS() methods to a value node.
 * This is needed because plugin functions like those in bootstrap-less-port
 * call value.eval(context) on variable values.
 * @param {*} value - The value to augment
 * @returns {*} The augmented value
 */
function augmentValueWithMethods(value) {
  if (value === null || value === undefined) {
    return value;
  }

  if (typeof value !== 'object') {
    return value;
  }

  // If value already has eval, return as-is (but add toCSS if missing)
  if (typeof value.eval === 'function' && typeof value.toCSS === 'function') {
    return value;
  }

  // Add eval method that returns the value itself (already evaluated from Go)
  if (typeof value.eval !== 'function') {
    value.eval = function(context) {
      return this;
    };
  }

  // Add toCSS method based on node type
  if (typeof value.toCSS !== 'function') {
    const nodeType = value._type || value.type;
    switch (nodeType) {
      case 'Color':
        value.toCSS = function() {
          const rgb = this.rgb || [0, 0, 0];
          const alpha = this.alpha !== undefined ? this.alpha : 1;
          if (alpha < 1) {
            return `rgba(${Math.round(rgb[0])}, ${Math.round(rgb[1])}, ${Math.round(rgb[2])}, ${alpha})`;
          }
          // Convert to hex
          const r = Math.round(rgb[0]).toString(16).padStart(2, '0');
          const g = Math.round(rgb[1]).toString(16).padStart(2, '0');
          const b = Math.round(rgb[2]).toString(16).padStart(2, '0');
          return `#${r}${g}${b}`;
        };
        break;

      case 'Dimension':
        value.toCSS = function() {
          return `${this.value}${this.unit || ''}`;
        };
        break;

      case 'Quoted':
        value.toCSS = function() {
          return this.escaped ? this.value : `${this.quote || '"'}${this.value}${this.quote || '"'}`;
        };
        break;

      case 'Keyword':
      case 'Anonymous':
        value.toCSS = function() {
          return String(this.value || '');
        };
        break;

      case 'Expression':
      case 'Value':
        value.toCSS = function() {
          if (!Array.isArray(this.value)) {
            return String(this.value || '');
          }
          return this.value.map(v => {
            if (v && typeof v.toCSS === 'function') {
              return v.toCSS();
            }
            return String(v || '');
          }).join(' ');
        };
        break;

      default:
        // Generic fallback - try to stringify the value
        value.toCSS = function() {
          if (this.value !== undefined) {
            if (typeof this.value === 'object' && this.value && typeof this.value.toCSS === 'function') {
              return this.value.toCSS();
            }
            return String(this.value);
          }
          return '';
        };
    }
  }

  // If value has nested values (like Expression or Value), augment them too
  if (Array.isArray(value.value)) {
    value.value = value.value.map(augmentValueWithMethods);
  } else if (value.value && typeof value.value === 'object') {
    value.value = augmentValueWithMethods(value.value);
  }

  return value;
}

/**
 * Create an evaluation context that plugin functions can use to look up variables.
 * This matches the structure that less.js plugins expect via `this.context`.
 *
 * Supports three modes:
 * 1. Pre-serialized: All frames and variables are sent upfront (legacy mode)
 * 2. Prefetch mode: Known variables are pre-sent, others fetched via callback
 * 3. On-demand lookup: Only frame count is sent; variables are fetched via callback
 *
 * @param {Object} contextData - Serialized context from Go
 * @returns {Object} An evaluation context object
 */
function createEvalContext(contextData) {
  if (!contextData) {
    return null;
  }

  // Check if we should use prefetch mode (best performance)
  if (contextData.usePrefetch) {
    return createPrefetchEvalContext(contextData);
  }

  // Check if we should use on-demand lookup mode
  if (contextData.useOnDemandLookup) {
    return createOnDemandEvalContext(contextData);
  }

  // Legacy mode: use pre-serialized frames
  if (!contextData.frames) {
    return null;
  }

  // The evalContext needs a reference back to itself for value.eval(context) calls
  const evalContext = {};

  // Create frame objects with variable() method
  const frames = (contextData.frames || []).map((frameData) => {
    const variables = frameData.variables || {};

    return {
      // The variable() method looks up a variable by name
      // Returns { value, important } or undefined
      variable(name) {
        const varDecl = variables[name];
        if (!varDecl) {
          return undefined;
        }
        // Convert the serialized value back to a node-like object
        let value = convertGoNodeToJS(varDecl.value);
        // Augment with eval() and toCSS() methods
        value = augmentValueWithMethods(value);
        return {
          value: value,
          important: varDecl.important || false,
        };
      },
      // For debugging
      _variables: variables,
    };
  });

  // Create importantScope array
  const importantScope = contextData.importantScope || [{}];

  // Set up the context object
  evalContext.frames = frames;
  evalContext.importantScope = importantScope;
  // Provide access to the plugin manager for functions like mix()
  evalContext.pluginManager = {
    less: {
      functions: {
        functionRegistry: builtinFunctionRegistry,
      },
    },
  };

  return evalContext;
}

/**
 * Create an evaluation context that uses on-demand variable lookup.
 * Supports two modes:
 * 1. Shared memory mode: Go writes binary data to mmap'd file, JS reads directly
 * 2. JSON fallback mode: Uses callbacks with JSON serialization
 *
 * @param {Object} contextData - Minimal context data with frameCount
 * @returns {Object} An evaluation context object
 */
function createOnDemandEvalContext(contextData) {
  const frameCount = contextData.frameCount || 0;
  const importantScope = contextData.importantScope || [{}];
  const useSharedMemory = contextData.useSharedMemory && varBuffer;

  // Cache for already-looked-up variables
  const variableCache = new Map();

  if (process.env.LESS_GO_DEBUG) {
    console.error(`[plugin-host] Creating on-demand eval context: ${frameCount} frames, shm=${useSharedMemory}`);
  }

  // Create frame objects that use on-demand lookup
  const frames = [];
  for (let i = 0; i < frameCount; i++) {
    const frameIndex = i;
    frames.push({
      // The variable() method looks up a variable by name via callback to Go
      variable(name) {
        // Check cache first
        const cacheKey = `${frameIndex}:${name}`;
        if (variableCache.has(cacheKey)) {
          return variableCache.get(cacheKey);
        }

        // Request the variable from Go
        try {
          const result = sendCallbackSync('lookupVariable', {
            name: name,
            frameIndex: frameIndex,
          });

          if (!result || !result.found) {
            variableCache.set(cacheKey, undefined);
            return undefined;
          }

          let value;
          let important = false;

          // Check if we should read from shared memory or use JSON
          if (result.shmOffset !== undefined && result.shmLength !== undefined) {
            // Read from shared memory (binary format)
            const shmData = readVariableFromSharedMemory(result.shmOffset, result.shmLength);
            if (shmData) {
              value = convertGoNodeToJS(shmData.value);
              important = shmData.important;
            } else {
              // Fallback to JSON if shared memory read fails
              if (result.value) {
                value = convertGoNodeToJS(result.value.value);
                important = result.value.important || false;
              }
            }
          } else if (result.value) {
            // JSON mode
            value = convertGoNodeToJS(result.value.value);
            important = result.value.important || false;
          }

          if (!value) {
            variableCache.set(cacheKey, undefined);
            return undefined;
          }

          // Augment with eval() and toCSS() methods
          value = augmentValueWithMethods(value);

          const varResult = {
            value: value,
            important: important,
          };

          // Cache at all frame indices from 0 to the found frame
          for (let j = 0; j <= result.frameIndex; j++) {
            variableCache.set(`${j}:${name}`, varResult);
          }

          return varResult;
        } catch (e) {
          if (process.env.LESS_GO_DEBUG) {
            console.error(`[plugin-host] Variable lookup failed for ${name}: ${e.message}`);
          }
          variableCache.set(cacheKey, undefined);
          return undefined;
        }
      },
      // For debugging
      _frameIndex: frameIndex,
      _onDemand: true,
    });
  }

  // The evalContext needs a reference back to itself for value.eval(context) calls
  const evalContext = {
    frames: frames,
    importantScope: importantScope,
    // Provide access to the plugin manager for functions like mix()
    pluginManager: {
      less: {
        functions: {
          functionRegistry: builtinFunctionRegistry,
        },
      },
    },
  };

  return evalContext;
}

/**
 * Create an evaluation context that uses prefetched variables.
 * This is the fastest mode - Go pre-serializes commonly needed variables
 * and sends them with the context, avoiding IPC round-trips.
 *
 * OPTIMIZED: When useSharedMemory is true, reads variables from binary
 * shared memory format using DataView - NO JSON parsing for variable data.
 *
 * For any variables not in the prefetch list, it falls back to on-demand callback.
 *
 * @param {Object} contextData - Context data with prefetchedVars or shared memory info
 * @returns {Object} An evaluation context object
 */
function createPrefetchEvalContext(contextData) {
  const frameCount = contextData.frameCount || 0;
  const importantScope = contextData.importantScope || [{}];

  // Cache for looked-up variables
  const variableCache = new Map();

  // Check if we should use shared memory (binary format)
  if (contextData.useSharedMemory) {
    // Read variables from binary shared memory
    const bufferPath = contextData.prefetchBufferPath;
    const bufferSize = contextData.prefetchBufferSize;

    if (bufferPath && bufferSize > 0) {
      try {
        // Read the binary data from the shared memory file
        const buffer = fs.readFileSync(bufferPath);

        // Parse the binary prefetch format
        const prefetchedVars = parseBinaryPrefetchBuffer(buffer, bufferSize);

        // Populate cache from binary data
        for (const [name, varData] of prefetchedVars) {
          let value = varData.value;
          value = augmentValueWithMethods(value);
          const varResult = {
            value: value,
            important: varData.important || false,
          };
          variableCache.set(`0:${name}`, varResult);
        }

        if (process.env.LESS_GO_DEBUG) {
          console.error(`[plugin-host] Creating prefetch eval context (BINARY): ${frameCount} frames, ${prefetchedVars.size} prefetched vars from shared memory`);
        }
      } catch (e) {
        if (process.env.LESS_GO_DEBUG) {
          console.error(`[plugin-host] Failed to read shared memory: ${e.message}, falling back to JSON`);
        }
        // Fall through to JSON path
      }
    }
  } else {
    // Use JSON prefetched variables (fallback path)
    const prefetchedVars = contextData.prefetchedVars || {};

    for (const [name, varDecl] of Object.entries(prefetchedVars)) {
      if (varDecl) {
        let value = convertGoNodeToJS(varDecl.value);
        value = augmentValueWithMethods(value);
        const varResult = {
          value: value,
          important: varDecl.important || false,
        };
        variableCache.set(`0:${name}`, varResult);
      }
    }

    if (process.env.LESS_GO_DEBUG) {
      console.error(`[plugin-host] Creating prefetch eval context (JSON): ${frameCount} frames, ${Object.keys(prefetchedVars).length} prefetched vars`);
    }
  }

  // Create frame objects that use cached values first, then callback
  const frames = [];
  for (let i = 0; i < frameCount; i++) {
    const frameIndex = i;
    frames.push({
      variable(name) {
        // Check cache first (including prefetched vars)
        const cacheKey = `${frameIndex}:${name}`;
        if (variableCache.has(cacheKey)) {
          return variableCache.get(cacheKey);
        }

        // For frame 0, check if we have it prefetched
        if (frameIndex === 0 && variableCache.has(`0:${name}`)) {
          return variableCache.get(`0:${name}`);
        }

        // Fall back to callback for non-prefetched variables
        try {
          const result = sendCallbackSync('lookupVariable', {
            name: name,
            frameIndex: frameIndex,
          });

          if (!result || !result.found) {
            variableCache.set(cacheKey, undefined);
            return undefined;
          }

          let value;
          let important = false;

          // Check if value is in shared memory or JSON
          if (result.shmOffset !== undefined && result.shmLength !== undefined) {
            // Read from shared memory - binary format
            const shmData = readVariableFromSharedMemory(result.shmOffset, result.shmLength);
            if (shmData) {
              value = shmData.value;
              important = shmData.important || false;
            }
          } else if (result.value) {
            // JSON fallback
            value = convertGoNodeToJS(result.value.value);
            important = result.value.important || false;
          }

          if (!value) {
            variableCache.set(cacheKey, undefined);
            return undefined;
          }

          value = augmentValueWithMethods(value);

          const varResult = {
            value: value,
            important: important,
          };

          // Cache at all frame indices from 0 to the found frame
          for (let j = 0; j <= result.frameIndex; j++) {
            variableCache.set(`${j}:${name}`, varResult);
          }

          return varResult;
        } catch (e) {
          if (process.env.LESS_GO_DEBUG) {
            console.error(`[plugin-host] Variable lookup failed for ${name}: ${e.message}`);
          }
          variableCache.set(cacheKey, undefined);
          return undefined;
        }
      },
      _frameIndex: frameIndex,
      _prefetch: true,
    });
  }

  const evalContext = {
    frames: frames,
    importantScope: importantScope,
    pluginManager: {
      less: {
        functions: {
          functionRegistry: builtinFunctionRegistry,
        },
      },
    },
  };

  return evalContext;
}

// ============================================================================
// Binary Prefetch Buffer Parser
// ============================================================================
//
// Binary format for prefetched variables:
//
// Header:
//   [4 bytes] magic: 0x50524546 ("PREF")
//   [4 bytes] version: 1
//   [4 bytes] variable count
//
// For each variable:
//   [4 bytes] name length
//   [N bytes] name (UTF-8)
//   [1 byte]  important flag (0 or 1)
//   [1 byte]  type (0=null, 1=Dimension, 2=Color, 3=Quoted, 4=Keyword, 5=Expression, 6=Anonymous)
//   [variable] value data (type-specific)
//
// Type-specific value encodings:
//   Dimension (type=1):
//     [8 bytes] value (float64 LE)
//     [4 bytes] unit length
//     [N bytes] unit string (UTF-8)
//
//   Color (type=2):
//     [8 bytes] R (float64 LE)
//     [8 bytes] G (float64 LE)
//     [8 bytes] B (float64 LE)
//     [8 bytes] alpha (float64 LE)
//
//   Quoted (type=3):
//     [4 bytes] string length
//     [N bytes] string value (UTF-8)
//     [1 byte]  quote character
//     [1 byte]  escaped flag (0 or 1)
//
//   Keyword (type=4):
//     [4 bytes] string length
//     [N bytes] string value (UTF-8)
//
//   Expression (type=5):
//     [4 bytes] value count
//     For each value:
//       [1 byte]  type
//       [variable] value data
//
//   Anonymous (type=6):
//     [4 bytes] string length
//     [N bytes] string value (UTF-8)
// ============================================================================

const PREFETCH_MAGIC = 0x50524546; // "PREF"
const PREFETCH_VERSION = 1;

const VAR_TYPE_NULL = 0;
const VAR_TYPE_DIMENSION = 1;
const VAR_TYPE_COLOR = 2;
const VAR_TYPE_QUOTED = 3;
const VAR_TYPE_KEYWORD = 4;
const VAR_TYPE_EXPRESSION = 5;
const VAR_TYPE_ANONYMOUS = 6;

/**
 * Parse the binary prefetch buffer and return a Map of variable name -> {value, important}.
 * Uses DataView for efficient binary reading - NO JSON.parse() involved.
 *
 * @param {Buffer} buffer - The binary buffer from shared memory
 * @param {number} dataSize - The actual size of valid data in the buffer
 * @returns {Map<string, {value: any, important: boolean}>} Map of variable names to values
 */
function parseBinaryPrefetchBuffer(buffer, dataSize) {
  const result = new Map();

  if (buffer.length < 12) {
    return result; // Buffer too small for header
  }

  let offset = 0;

  // Read and verify header
  const magic = buffer.readUInt32LE(offset);
  offset += 4;

  if (magic !== PREFETCH_MAGIC) {
    if (process.env.LESS_GO_DEBUG) {
      console.error(`[plugin-host] Invalid prefetch magic: expected 0x${PREFETCH_MAGIC.toString(16)}, got 0x${magic.toString(16)}`);
    }
    return result;
  }

  const version = buffer.readUInt32LE(offset);
  offset += 4;

  if (version !== PREFETCH_VERSION) {
    if (process.env.LESS_GO_DEBUG) {
      console.error(`[plugin-host] Unsupported prefetch version: ${version}`);
    }
    return result;
  }

  const varCount = buffer.readUInt32LE(offset);
  offset += 4;

  // Read each variable
  for (let i = 0; i < varCount && offset < dataSize; i++) {
    try {
      // Read variable name
      const nameLen = buffer.readUInt32LE(offset);
      offset += 4;

      const name = buffer.toString('utf8', offset, offset + nameLen);
      offset += nameLen;

      // Read important flag
      const important = buffer[offset++] === 1;

      // Read type
      const varType = buffer[offset++];

      // Read value based on type
      let value = null;

      switch (varType) {
        case VAR_TYPE_NULL:
          // Null value, nothing more to read
          break;

        case VAR_TYPE_DIMENSION:
          {
            const numValue = readFloat64LE(buffer, offset);
            offset += 8;
            const unitLen = buffer.readUInt32LE(offset);
            offset += 4;
            const unit = buffer.toString('utf8', offset, offset + unitLen);
            offset += unitLen;
            value = {
              _type: 'Dimension',
              value: numValue,
              unit: unit ? { numerator: [unit], denominator: [] } : { numerator: [], denominator: [] },
            };
          }
          break;

        case VAR_TYPE_COLOR:
          {
            const r = readFloat64LE(buffer, offset);
            offset += 8;
            const g = readFloat64LE(buffer, offset);
            offset += 8;
            const b = readFloat64LE(buffer, offset);
            offset += 8;
            const alpha = readFloat64LE(buffer, offset);
            offset += 8;
            value = {
              _type: 'Color',
              rgb: [r, g, b],
              alpha: alpha,
            };
          }
          break;

        case VAR_TYPE_QUOTED:
          {
            const strLen = buffer.readUInt32LE(offset);
            offset += 4;
            const str = buffer.toString('utf8', offset, offset + strLen);
            offset += strLen;
            const quote = String.fromCharCode(buffer[offset++]);
            const escaped = buffer[offset++] === 1;
            value = {
              _type: 'Quoted',
              value: str,
              quote: quote,
              escaped: escaped,
            };
          }
          break;

        case VAR_TYPE_KEYWORD:
          {
            const strLen = buffer.readUInt32LE(offset);
            offset += 4;
            const str = buffer.toString('utf8', offset, offset + strLen);
            offset += strLen;
            value = {
              _type: 'Keyword',
              value: str,
            };
          }
          break;

        case VAR_TYPE_EXPRESSION:
          {
            const valueCount = buffer.readUInt32LE(offset);
            offset += 4;
            const values = [];
            for (let j = 0; j < valueCount; j++) {
              const { value: subValue, newOffset } = readBinaryValue(buffer, offset);
              values.push(subValue);
              offset = newOffset;
            }
            value = {
              _type: 'Expression',
              value: values,
            };
          }
          break;

        case VAR_TYPE_ANONYMOUS:
          {
            const strLen = buffer.readUInt32LE(offset);
            offset += 4;
            const str = buffer.toString('utf8', offset, offset + strLen);
            offset += strLen;
            value = {
              _type: 'Anonymous',
              value: str,
            };
          }
          break;

        default:
          if (process.env.LESS_GO_DEBUG) {
            console.error(`[plugin-host] Unknown variable type: ${varType}`);
          }
          break;
      }

      if (value !== null) {
        result.set(name, { value, important });
      }
    } catch (e) {
      if (process.env.LESS_GO_DEBUG) {
        console.error(`[plugin-host] Error parsing variable at offset ${offset}: ${e.message}`);
      }
      break;
    }
  }

  return result;
}

/**
 * Read a single value from the binary buffer (for nested values in Expression).
 * @param {Buffer} buffer - The binary buffer
 * @param {number} offset - Current offset
 * @returns {{value: any, newOffset: number}} The parsed value and new offset
 */
function readBinaryValue(buffer, offset) {
  const varType = buffer[offset++];

  switch (varType) {
    case VAR_TYPE_NULL:
      return { value: null, newOffset: offset };

    case VAR_TYPE_DIMENSION:
      {
        const numValue = readFloat64LE(buffer, offset);
        offset += 8;
        const unitLen = buffer.readUInt32LE(offset);
        offset += 4;
        const unit = buffer.toString('utf8', offset, offset + unitLen);
        offset += unitLen;
        return {
          value: {
            _type: 'Dimension',
            value: numValue,
            unit: unit ? { numerator: [unit], denominator: [] } : { numerator: [], denominator: [] },
          },
          newOffset: offset,
        };
      }

    case VAR_TYPE_COLOR:
      {
        const r = readFloat64LE(buffer, offset);
        offset += 8;
        const g = readFloat64LE(buffer, offset);
        offset += 8;
        const b = readFloat64LE(buffer, offset);
        offset += 8;
        const alpha = readFloat64LE(buffer, offset);
        offset += 8;
        return {
          value: {
            _type: 'Color',
            rgb: [r, g, b],
            alpha: alpha,
          },
          newOffset: offset,
        };
      }

    case VAR_TYPE_QUOTED:
      {
        const strLen = buffer.readUInt32LE(offset);
        offset += 4;
        const str = buffer.toString('utf8', offset, offset + strLen);
        offset += strLen;
        const quote = String.fromCharCode(buffer[offset++]);
        const escaped = buffer[offset++] === 1;
        return {
          value: {
            _type: 'Quoted',
            value: str,
            quote: quote,
            escaped: escaped,
          },
          newOffset: offset,
        };
      }

    case VAR_TYPE_KEYWORD:
      {
        const strLen = buffer.readUInt32LE(offset);
        offset += 4;
        const str = buffer.toString('utf8', offset, offset + strLen);
        offset += strLen;
        return {
          value: {
            _type: 'Keyword',
            value: str,
          },
          newOffset: offset,
        };
      }

    case VAR_TYPE_ANONYMOUS:
      {
        const strLen = buffer.readUInt32LE(offset);
        offset += 4;
        const str = buffer.toString('utf8', offset, offset + strLen);
        offset += strLen;
        return {
          value: {
            _type: 'Anonymous',
            value: str,
          },
          newOffset: offset,
        };
      }

    default:
      return { value: null, newOffset: offset };
  }
}

/**
 * Read a float64 from buffer in little-endian format.
 * Uses DataView for proper IEEE-754 float parsing.
 * @param {Buffer} buf - The buffer
 * @param {number} offset - Offset to read from
 * @returns {number} The float64 value
 */
function readFloat64LE(buf, offset) {
  // Create a DataView to properly read the float64
  const arrayBuffer = buf.buffer.slice(buf.byteOffset + offset, buf.byteOffset + offset + 8);
  const view = new DataView(arrayBuffer);
  return view.getFloat64(0, true); // true = little-endian
}

/**
 * Built-in function registry that includes both plugin functions and built-in Less functions.
 * Plugins access this via context.pluginManager.less.functions.functionRegistry
 */
const builtinFunctionRegistry = {
  get(name) {
    // Check if it's a built-in function first
    const builtin = builtinFunctions[name];
    if (builtin) {
      return builtin;
    }
    // Fall back to plugin-registered functions
    return lookupFunction(name);
  },
};

/**
 * Built-in Less functions implemented in JavaScript for use by plugins.
 * These match the behavior of the corresponding Go/Less.js implementations.
 */
const builtinFunctions = {
  /**
   * Mix two colors together in variable proportion.
   * Opacity is included in the calculations.
   * @param {Color} color1 - First color
   * @param {Color} color2 - Second color
   * @param {Dimension} weight - Optional weight (0-100%), default 50%
   * @returns {Color} Mixed color
   */
  mix: function(color1, color2, weight) {
    // Get weight as a decimal (0-1)
    let w = 0.5; // default 50%
    if (weight) {
      if (typeof weight.value === 'number') {
        w = weight.value / 100;
      } else if (typeof weight === 'number') {
        w = weight / 100;
      }
    }

    // Get RGB values
    const rgb1 = color1.rgb || [0, 0, 0];
    const rgb2 = color2.rgb || [0, 0, 0];
    const alpha1 = color1.alpha !== undefined ? color1.alpha : 1;
    const alpha2 = color2.alpha !== undefined ? color2.alpha : 1;

    // Mix the colors
    // The algorithm matches less.js:
    // w1 and w2 are the weights for each color
    const w1 = w;
    const w2 = 1 - w;

    const rgb = [
      Math.round(rgb1[0] * w1 + rgb2[0] * w2),
      Math.round(rgb1[1] * w1 + rgb2[1] * w2),
      Math.round(rgb1[2] * w1 + rgb2[2] * w2),
    ];

    // Mix alpha
    const alpha = alpha1 * w1 + alpha2 * w2;

    const result = createNode('Color', { rgb, alpha });
    // Add toCSS method for use by plugins
    result.toCSS = function() {
      if (this.alpha < 1) {
        return `rgba(${this.rgb[0]}, ${this.rgb[1]}, ${this.rgb[2]}, ${this.alpha})`;
      }
      const r = this.rgb[0].toString(16).padStart(2, '0');
      const g = this.rgb[1].toString(16).padStart(2, '0');
      const b = this.rgb[2].toString(16).padStart(2, '0');
      return `#${r}${g}${b}`;
    };
    return result;
  },
};

/**
 * Call a registered function
 * @param {number} id - Command ID
 * @param {Object} data - Function call data
 */
function handleCallFunction(id, data) {
  const { name, args, context } = data || {};

  if (!name) {
    sendResponse(id, false, null, 'Function name is required');
    return;
  }

  // Use scoped lookup for proper function resolution
  const fn = lookupFunction(name);
  if (!fn) {
    sendResponse(id, false, null, `Function not found: ${name}`);
    return;
  }

  try {
    // Convert Go node arguments to JavaScript format
    const convertedArgs = (args || []).map(convertGoNodeToJS);

    // Create evaluation context if provided
    const evalContext = createEvalContext(context);

    // Create the `this` binding with context
    const thisBinding = {
      context: evalContext,
    };

    // Call the function with context as `this` and converted arguments
    const result = fn.apply(thisBinding, convertedArgs);

    // Debug: log the raw result for detached-ruleset
    if (name === 'test-detached-ruleset' && process.env.LESS_GO_DEBUG) {
      console.error('[DEBUG plugin-host] test-detached-ruleset raw result:', JSON.stringify(result, null, 2));
    }

    // Convert the result back to Go-compatible format
    const goResult = convertJSResultToGo(result);

    // Debug: log the converted result
    if (name === 'test-detached-ruleset' && process.env.LESS_GO_DEBUG) {
      console.error('[DEBUG plugin-host] test-detached-ruleset converted result:', JSON.stringify(goResult, null, 2));
    }

    sendResponse(id, true, goResult);
  } catch (err) {
    sendResponse(id, false, null, `Function error: ${err.message}\n${err.stack || ''}`);
  }
}

/**
 * Attach to a shared memory buffer from Go
 * @param {number} id - Command ID
 * @param {Object} data - Buffer data { key, path, size }
 */
function handleAttachBuffer(id, data) {
  const { key, path: bufferPath, size } = data || {};

  if (!key || !bufferPath) {
    sendResponse(id, false, null, 'Buffer key and path are required');
    return;
  }

  try {
    // Check if already attached
    if (attachedBuffers.has(key)) {
      sendResponse(id, true, { cached: true, key, size });
      return;
    }

    // Read the file into a buffer
    // Note: We read the entire file for now. For very large files,
    // we could use memory mapping via native modules if needed.
    const buffer = fs.readFileSync(bufferPath);

    attachedBuffers.set(key, {
      path: bufferPath,
      size: buffer.length,
      buffer,
    });

    sendResponse(id, true, {
      cached: false,
      key,
      size: buffer.length,
    });
  } catch (err) {
    sendResponse(id, false, null, `Failed to attach buffer: ${err.message}`);
  }
}

/**
 * Detach from a shared memory buffer
 * @param {number} id - Command ID
 * @param {Object} data - Buffer data { key }
 */
function handleDetachBuffer(id, data) {
  const { key } = data || {};

  if (!key) {
    sendResponse(id, false, null, 'Buffer key is required');
    return;
  }

  if (!attachedBuffers.has(key)) {
    sendResponse(id, false, null, `Buffer not found: ${key}`);
    return;
  }

  attachedBuffers.delete(key);
  sendResponse(id, true, { detached: true, key });
}

/**
 * Read data from an attached buffer
 * @param {number} id - Command ID
 * @param {Object} data - Read data { key, offset, length }
 */
function handleReadBuffer(id, data) {
  const { key, offset = 0, length } = data || {};

  if (!key) {
    sendResponse(id, false, null, 'Buffer key is required');
    return;
  }

  const bufInfo = attachedBuffers.get(key);
  if (!bufInfo) {
    sendResponse(id, false, null, `Buffer not found: ${key}`);
    return;
  }

  const { buffer } = bufInfo;
  const readLength = length || buffer.length - offset;

  if (offset < 0 || offset + readLength > buffer.length) {
    sendResponse(id, false, null, `Read out of bounds: offset=${offset}, length=${readLength}, size=${buffer.length}`);
    return;
  }

  // Return the data as base64 for JSON transport
  const slice = buffer.slice(offset, offset + readLength);
  sendResponse(id, true, {
    data: slice.toString('base64'),
    offset,
    length: readLength,
  });
}

/**
 * Get info about an attached buffer
 * @param {number} id - Command ID
 * @param {Object} data - Buffer data { key }
 */
function handleGetBufferInfo(id, data) {
  const { key } = data || {};

  if (!key) {
    sendResponse(id, false, null, 'Buffer key is required');
    return;
  }

  const bufInfo = attachedBuffers.get(key);
  if (!bufInfo) {
    sendResponse(id, false, null, `Buffer not found: ${key}`);
    return;
  }

  sendResponse(id, true, {
    key,
    path: bufInfo.path,
    size: bufInfo.size,
  });
}

/**
 * Parse the FlatAST binary format from a buffer
 * This is the JavaScript equivalent of FromBytes in Go
 */
function parseFlatAST(buffer) {
  if (buffer.length < 28) {
    throw new Error('Buffer too small for FlatAST header');
  }

  let offset = 0;

  // Read and verify magic
  const magic = buffer.readUInt32LE(offset);
  offset += 4;
  if (magic !== 0x4C455353) { // "LESS"
    throw new Error(`Invalid magic: expected 0x4C455353, got 0x${magic.toString(16)}`);
  }

  // Read header
  const version = buffer.readUInt32LE(offset);
  offset += 4;
  const nodeCount = buffer.readUInt32LE(offset);
  offset += 4;
  const rootIndex = buffer.readUInt32LE(offset);
  offset += 4;
  const nodesOffset = buffer.readUInt32LE(offset);
  offset += 4;
  const stringTableOffset = buffer.readUInt32LE(offset);
  offset += 4;
  const typeTableOffset = buffer.readUInt32LE(offset);
  offset += 4;

  // Read nodes (each node is 24 bytes)
  const nodes = [];
  offset = nodesOffset;
  for (let i = 0; i < nodeCount; i++) {
    const node = {
      typeID: buffer.readUInt16LE(offset),
      flags: buffer.readUInt16LE(offset + 2),
      childIndex: buffer.readUInt32LE(offset + 4),
      nextIndex: buffer.readUInt32LE(offset + 8),
      parentIndex: buffer.readUInt32LE(offset + 12),
      propsOffset: buffer.readUInt32LE(offset + 16),
      propsLength: buffer.readUInt32LE(offset + 20),
    };
    nodes.push(node);
    offset += 24;
  }

  // Read string table
  offset = stringTableOffset;
  const stringCount = buffer.readUInt32LE(offset);
  offset += 4;
  const stringTable = [];
  for (let i = 0; i < stringCount; i++) {
    const strLen = buffer.readUInt32LE(offset);
    offset += 4;
    const str = buffer.slice(offset, offset + strLen).toString('utf8');
    stringTable.push(str);
    offset += strLen;
  }

  // Read type table
  offset = typeTableOffset;
  const typeCount = buffer.readUInt32LE(offset);
  offset += 4;
  const typeTable = [];
  for (let i = 0; i < typeCount; i++) {
    const strLen = buffer.readUInt32LE(offset);
    offset += 4;
    const str = buffer.slice(offset, offset + strLen).toString('utf8');
    typeTable.push(str);
    offset += strLen;
  }

  // Read prop buffer
  const propLen = buffer.readUInt32LE(offset);
  offset += 4;
  const propBuffer = buffer.slice(offset, offset + propLen);

  return {
    version,
    nodeCount,
    rootIndex,
    nodes,
    stringTable,
    typeTable,
    propBuffer,
  };
}

/**
 * Get an attached buffer's parsed AST
 * @param {string} key - Buffer key
 * @returns {Object} Parsed FlatAST
 */
function getAttachedAST(key) {
  const bufInfo = attachedBuffers.get(key);
  if (!bufInfo) {
    throw new Error(`Buffer not found: ${key}`);
  }

  // Cache the parsed AST
  if (!bufInfo.ast) {
    bufInfo.ast = parseFlatAST(bufInfo.buffer);
  }

  return bufInfo.ast;
}

/**
 * Run a specific visitor on the AST buffer
 * @param {number} id - Command ID
 * @param {Object} data - { bufferKey, visitorIndex }
 */
function handleRunVisitor(id, data) {
  const { bufferKey, visitorIndex } = data || {};

  if (!bufferKey) {
    sendResponse(id, false, null, 'Buffer key is required');
    return;
  }

  if (visitorIndex === undefined || visitorIndex < 0 || visitorIndex >= registeredVisitors.length) {
    sendResponse(id, false, null, `Invalid visitor index: ${visitorIndex}`);
    return;
  }

  try {
    const ast = getAttachedAST(bufferKey);
    const visitor = registeredVisitors[visitorIndex];

    // Use bindings if available for better performance
    if (bindings && bindings.createRootFacade) {
      const root = bindings.createRootFacade(ast);
      visitor._replacements = [];

      const result = visitor.run ? visitor.run(root) : visitor.visit ? visitor.visit(root) : root;

      sendResponse(id, true, {
        success: true,
        replacements: visitor._replacements || [],
        resultType: result ? (result.type || result._type) : null,
      });
    } else {
      // Fallback: run visitor directly on parsed AST structure
      sendResponse(id, true, {
        success: true,
        message: 'Visitor executed (bindings not available)',
        replacements: [],
      });
    }
  } catch (err) {
    sendResponse(id, false, null, `Visitor error: ${err.message}\n${err.stack || ''}`);
  }
}

/**
 * Run all pre-eval visitors on the AST buffer
 * @param {number} id - Command ID
 * @param {Object} data - { bufferKey }
 */
function handleRunPreEvalVisitors(id, data) {
  const { bufferKey } = data || {};

  if (!bufferKey) {
    sendResponse(id, false, null, 'Buffer key is required');
    return;
  }

  try {
    const ast = getAttachedAST(bufferKey);
    const preEvalVisitors = registeredVisitors.filter(v => v.isPreEvalVisitor);
    const allReplacements = [];

    if (bindings && bindings.createRootFacade) {
      let root = bindings.createRootFacade(ast);

      for (let i = 0; i < preEvalVisitors.length; i++) {
        const visitor = preEvalVisitors[i];
        visitor._replacements = [];

        if (visitor.run) {
          root = visitor.run(root) || root;
        } else if (visitor.visit) {
          root = visitor.visit(root) || root;
        }

        if (visitor._replacements && visitor._replacements.length > 0) {
          allReplacements.push({
            visitorIndex: registeredVisitors.indexOf(visitor),
            replacements: visitor._replacements,
          });
        }
      }
    }

    sendResponse(id, true, {
      success: true,
      visitorCount: preEvalVisitors.length,
      replacements: allReplacements,
    });
  } catch (err) {
    sendResponse(id, false, null, `Pre-eval visitors error: ${err.message}\n${err.stack || ''}`);
  }
}

/**
 * Recursively visit a JSON AST node with a visitor.
 * @param {Object} node - The node to visit (JSON object)
 * @param {Object} visitor - The visitor with visitXxx methods
 * @param {string} parentPath - Path to this node for debugging
 * @returns {Object} The potentially modified node
 */
function visitNodeRecursive(node, visitor, parentPath = '') {
  if (!node || typeof node !== 'object') {
    return node;
  }

  // Handle arrays
  if (Array.isArray(node)) {
    return node.map((child, i) => visitNodeRecursive(child, visitor, `${parentPath}[${i}]`));
  }

  // Get node type
  const type = node._type || node.type || node.Type;
  if (!type) {
    // Not a typed node, but visit children anyway
    const result = { ...node };
    for (const [key, value] of Object.entries(node)) {
      if (value && typeof value === 'object') {
        result[key] = visitNodeRecursive(value, visitor, `${parentPath}.${key}`);
      }
    }
    return result;
  }

  // Check for visitor method
  const funcName = 'visit' + type;
  let result = node;

  if (visitor[funcName]) {
    result = visitor[funcName](node);
    // If visitor returned a different node, use it (replacement)
    if (result !== node && result !== undefined) {
      return result;
    }
    // If visitor returned undefined, keep original node
    if (result === undefined) {
      result = node;
    }
  }

  // Visit children recursively
  // Handle common child properties in Less AST nodes
  const childProps = ['value', 'rules', 'elements', 'selectors', 'args', 'operands',
                      'condition', 'lvalue', 'rvalue', 'params', 'mixinDefinitions',
                      'variables', 'properties', 'extendList', 'features', 'paths'];

  const newResult = { ...result };
  let modified = false;

  for (const prop of childProps) {
    if (result[prop] !== undefined && result[prop] !== null) {
      const visited = visitNodeRecursive(result[prop], visitor, `${parentPath}.${prop}`);
      if (visited !== result[prop]) {
        newResult[prop] = visited;
        modified = true;
      }
    }
  }

  // Also visit any array properties that might contain nodes
  for (const [key, value] of Object.entries(result)) {
    if (Array.isArray(value) && !childProps.includes(key)) {
      const visited = visitNodeRecursive(value, visitor, `${parentPath}.${key}`);
      if (visited !== value) {
        newResult[key] = visited;
        modified = true;
      }
    }
  }

  // Call visitXxxOut if defined
  const outFuncName = 'visit' + type + 'Out';
  if (visitor[outFuncName]) {
    visitor[outFuncName](modified ? newResult : result);
  }

  return modified ? newResult : result;
}

/**
 * Run pre-eval visitors on a JSON AST.
 * This is a simpler approach that works with JSON serialization.
 * @param {number} id - Command ID
 * @param {Object} data - { ast: JSON AST object }
 */
function handleRunPreEvalVisitorsJSON(id, data) {
  const { ast } = data || {};

  if (!ast) {
    sendResponse(id, false, null, 'AST is required');
    return;
  }

  try {
    const preEvalVisitors = registeredVisitors.filter(v => v.isPreEvalVisitor);

    if (preEvalVisitors.length === 0) {
      // No visitors, return unchanged
      sendResponse(id, true, {
        success: true,
        visitorCount: 0,
        modifiedAst: ast,
        modified: false,
      });
      return;
    }

    let currentAst = ast;
    let totalModified = false;

    for (const visitor of preEvalVisitors) {
      // For each visitor, walk the tree recursively
      const modifiedAst = visitNodeRecursive(currentAst, visitor, 'root');

      // Check if tree was modified
      if (JSON.stringify(modifiedAst) !== JSON.stringify(currentAst)) {
        currentAst = modifiedAst;
        totalModified = true;
      }
    }

    sendResponse(id, true, {
      success: true,
      visitorCount: preEvalVisitors.length,
      modifiedAst: currentAst,
      modified: totalModified,
    });
  } catch (err) {
    sendResponse(id, false, null, `Pre-eval visitors JSON error: ${err.message}\n${err.stack || ''}`);
  }
}

/**
 * Check which variables should be replaced by pre-eval visitors.
 * This is an optimization to avoid serializing the entire AST.
 * @param {number} id - Command ID
 * @param {Object} data - { variables: [{id, name}, ...] }
 */
function handleCheckVariableReplacements(id, data) {
  const { variables } = data || {};

  if (!variables || !Array.isArray(variables)) {
    sendResponse(id, false, null, 'Variables array is required');
    return;
  }

  try {
    const preEvalVisitors = registeredVisitors.filter(v => v.isPreEvalVisitor);

    if (preEvalVisitors.length === 0) {
      // No pre-eval visitors, no replacements
      sendResponse(id, true, {
        success: true,
        replacements: {},
      });
      return;
    }

    const replacements = {};

    for (const varInfo of variables) {
      // Create a mock Variable node for the visitor to check
      const mockNode = {
        _type: 'Variable',
        type: 'Variable',
        name: varInfo.name,
      };

      // Check each visitor
      for (const visitor of preEvalVisitors) {
        if (visitor.visitVariable) {
          const result = visitor.visitVariable(mockNode);
          // If visitor returned a different node, it's a replacement
          if (result !== mockNode && result !== undefined) {
            replacements[varInfo.id] = convertJSResultToGo(result);
            break; // First replacement wins
          }
        }
      }
    }

    sendResponse(id, true, {
      success: true,
      replacements: replacements,
    });
  } catch (err) {
    sendResponse(id, false, null, `Check variable replacements error: ${err.message}\n${err.stack || ''}`);
  }
}

/**
 * Run all post-eval visitors on the AST buffer
 * @param {number} id - Command ID
 * @param {Object} data - { bufferKey }
 */
function handleRunPostEvalVisitors(id, data) {
  const { bufferKey } = data || {};

  if (!bufferKey) {
    sendResponse(id, false, null, 'Buffer key is required');
    return;
  }

  try {
    const ast = getAttachedAST(bufferKey);
    const postEvalVisitors = registeredVisitors.filter(v => !v.isPreEvalVisitor);
    const allReplacements = [];

    if (bindings && bindings.createRootFacade) {
      let root = bindings.createRootFacade(ast);

      for (let i = 0; i < postEvalVisitors.length; i++) {
        const visitor = postEvalVisitors[i];
        visitor._replacements = [];

        if (visitor.run) {
          root = visitor.run(root) || root;
        } else if (visitor.visit) {
          root = visitor.visit(root) || root;
        }

        if (visitor._replacements && visitor._replacements.length > 0) {
          allReplacements.push({
            visitorIndex: registeredVisitors.indexOf(visitor),
            replacements: visitor._replacements,
          });
        }
      }
    }

    sendResponse(id, true, {
      success: true,
      visitorCount: postEvalVisitors.length,
      replacements: allReplacements,
    });
  } catch (err) {
    sendResponse(id, false, null, `Post-eval visitors error: ${err.message}\n${err.stack || ''}`);
  }
}

/**
 * Parse an AST buffer and return the structure
 * @param {number} id - Command ID
 * @param {Object} data - { bufferKey }
 */
function handleParseASTBuffer(id, data) {
  const { bufferKey } = data || {};

  if (!bufferKey) {
    sendResponse(id, false, null, 'Buffer key is required');
    return;
  }

  try {
    const ast = getAttachedAST(bufferKey);

    // Return a summary of the AST
    sendResponse(id, true, {
      version: ast.version,
      nodeCount: ast.nodeCount,
      rootIndex: ast.rootIndex,
      stringTableSize: ast.stringTable.length,
      typeTableSize: ast.typeTable.length,
    });
  } catch (err) {
    sendResponse(id, false, null, `Parse error: ${err.message}`);
  }
}

/**
 * Serialize a JavaScript node to buffer format
 * @param {number} id - Command ID
 * @param {Object} data - { node } - The node to serialize
 */
function handleSerializeNode(id, data) {
  const { node } = data || {};

  if (!node) {
    sendResponse(id, false, null, 'Node is required');
    return;
  }

  try {
    if (bindings && bindings.serializeToBuffer) {
      const buffer = bindings.serializeToBuffer(node);
      sendResponse(id, true, {
        buffer: buffer.toString('base64'),
        size: buffer.length,
      });
    } else {
      // Fallback: return the node as JSON
      sendResponse(id, true, {
        json: JSON.stringify(node),
        size: 0,
      });
    }
  } catch (err) {
    sendResponse(id, false, null, `Serialize error: ${err.message}`);
  }
}

/**
 * Call a registered function using shared memory for zero-copy arg/result transfer.
 * Arguments are read directly from the shared memory buffer using NodeFacade.
 * Results are written back to the buffer using BufferWriter.
 *
 * @param {number} id - Command ID
 * @param {Object} data - { name, bufferKey, argIndices, argsSize }
 */
function handleCallFunctionSharedMem(id, data) {
  const { name, bufferKey, argIndices, argsSize } = data || {};

  if (!name) {
    sendResponse(id, false, null, 'Function name is required');
    return;
  }

  if (!bufferKey) {
    sendResponse(id, false, null, 'Buffer key is required');
    return;
  }

  // Use scoped lookup for proper function resolution
  const fn = lookupFunction(name);
  if (!fn) {
    sendResponse(id, false, null, `Function not found: ${name}`);
    return;
  }

  try {
    // Get the attached buffer
    const bufferInfo = attachedBuffers.get(bufferKey);
    if (!bufferInfo) {
      sendResponse(id, false, null, `Buffer not attached: ${bufferKey}`);
      return;
    }

    const buffer = bufferInfo.buffer;
    let args = [];

    // If we have arguments and bindings available, use NodeFacade for zero-copy access
    if (argIndices && argIndices.length > 0 && bindings && bindings.NodeFacade) {
      try {
        // Parse the FlatAST from the buffer
        const ast = parseFlatAST(buffer);

        // Create NodeFacade for each argument
        args = argIndices.map(idx => {
          if (idx === 0 && ast.nodeCount === 0) {
            return null;
          }
          if (idx < ast.nodeCount) {
            return new bindings.NodeFacade(ast, idx);
          }
          return null;
        });
      } catch (parseErr) {
        // If parsing fails, fall back to no args
        // This can happen if the Go side sent primitives that can't be flattened
        args = [];
      }
    } else if (argIndices && argIndices.length > 0) {
      // Without bindings, try to parse the buffer and convert to simple objects
      try {
        const ast = parseFlatAST(buffer);
        args = argIndices.map(idx => {
          if (idx === 0 && ast.nodeCount === 0) {
            return null;
          }
          if (idx < ast.nodeCount) {
            // Convert FlatAST node to a simple object
            return convertFlatNodeToObject(ast, idx);
          }
          return null;
        });
      } catch (parseErr) {
        args = [];
      }
    }

    // Call the function with the arguments
    const result = fn.apply(null, args);

    // Try to write the result back to shared memory
    if (result !== null && result !== undefined && typeof result === 'object') {
      try {
        // Use bindings to serialize if available
        if (bindings && bindings.serializeToBuffer) {
          const resultBuffer = bindings.serializeToBuffer(result);

          // Check if the result fits in the remaining buffer space
          const resultOffset = argsSize || 0;
          if (resultOffset + resultBuffer.length <= buffer.length) {
            // Write the result to the buffer after the args
            resultBuffer.copy(buffer, resultOffset);

            // Write the updated buffer back to the file for Go to read
            fs.writeFileSync(bufferInfo.path, buffer);

            sendResponse(id, true, {
              resultOffset: resultOffset,
              resultSize: resultBuffer.length,
            });
            return;
          }
        }
      } catch (writeErr) {
        // Fall back to JSON if buffer writing fails
      }
    }

    // Fallback: return result as JSON
    const goResult = convertJSResultToGo(result);
    sendResponse(id, true, { jsonResult: goResult });

  } catch (err) {
    sendResponse(id, false, null, `Function error: ${err.message}\n${err.stack || ''}`);
  }
}

/**
 * Convert a FlatAST node to a simple JavaScript object.
 * Used when NodeFacade bindings are not available.
 *
 * @param {Object} ast - Parsed FlatAST
 * @param {number} idx - Node index
 * @returns {Object}
 */
function convertFlatNodeToObject(ast, idx) {
  if (idx >= ast.nodes.length) {
    return null;
  }

  const node = ast.nodes[idx];
  const typeName = getTypeNameFromID(node.typeID);

  // Get properties
  let props = {};
  if (node.propsLength > 0 && node.propsOffset + node.propsLength <= ast.propBuffer.length) {
    try {
      const propData = ast.propBuffer.slice(node.propsOffset, node.propsOffset + node.propsLength);
      props = JSON.parse(propData.toString('utf8'));
    } catch (e) {
      // Ignore parse errors
    }
  }

  // Resolve string indices in properties
  for (const [key, value] of Object.entries(props)) {
    if (typeof value === 'number' && key !== 'value' && !key.includes('Index')) {
      // Might be a string table index
      if (value < ast.stringTable.length) {
        props[key] = ast.stringTable[value];
      }
    }
  }

  const result = {
    _type: typeName,
    type: typeName,
    ...props,
  };

  // Handle flags
  if (node.flags & 0x01) result.parens = true;
  if (node.flags & 0x02) result.parensInOp = true;

  // Collect children
  if (node.childIndex > 0) {
    result.children = [];
    let childIdx = node.childIndex;
    while (childIdx > 0 && childIdx < ast.nodes.length) {
      const child = convertFlatNodeToObject(ast, childIdx);
      if (child) {
        result.children.push(child);
      }
      childIdx = ast.nodes[childIdx].nextIndex;
    }
  }

  return result;
}

/**
 * Get type name from type ID.
 * @param {number} typeID
 * @returns {string}
 */
function getTypeNameFromID(typeID) {
  const typeNames = {
    0: 'Unknown',
    1: 'Anonymous',
    2: 'Assignment',
    3: 'AtRule',
    4: 'Attribute',
    5: 'Call',
    6: 'Color',
    7: 'Combinator',
    8: 'Comment',
    9: 'Condition',
    10: 'Container',
    11: 'Declaration',
    12: 'DetachedRuleset',
    13: 'Dimension',
    14: 'Element',
    15: 'Expression',
    16: 'Extend',
    17: 'Import',
    18: 'JavaScript',
    19: 'Keyword',
    20: 'Media',
    21: 'MixinCall',
    22: 'MixinDefinition',
    23: 'NamespaceValue',
    24: 'Negative',
    25: 'Operation',
    26: 'Paren',
    27: 'Property',
    28: 'QueryInParens',
    29: 'Quoted',
    30: 'Ruleset',
    31: 'Selector',
    32: 'SelectorList',
    33: 'UnicodeDescriptor',
    34: 'Unit',
    35: 'URL',
    36: 'Value',
    37: 'Variable',
    38: 'VariableCall',
    39: 'Node',
  };
  return typeNames[typeID] || 'Unknown';
}

/**
 * Get the list of registered pre-processors
 * @param {number} id - Command ID
 * @param {Object} data - Unused
 */
function handleGetPreProcessors(id, data) {
  try {
    const processors = registeredPreProcessors.map((p, index) => ({
      index,
      priority: p.priority,
    }));
    sendResponse(id, true, processors);
  } catch (err) {
    sendResponse(id, false, null, `Failed to get pre-processors: ${err.message}`);
  }
}

/**
 * Get the list of registered post-processors
 * @param {number} id - Command ID
 * @param {Object} data - Unused
 */
function handleGetPostProcessors(id, data) {
  try {
    const processors = registeredPostProcessors.map((p, index) => ({
      index,
      priority: p.priority,
    }));
    sendResponse(id, true, processors);
  } catch (err) {
    sendResponse(id, false, null, `Failed to get post-processors: ${err.message}`);
  }
}

/**
 * Run a pre-processor on the input source
 * @param {number} id - Command ID
 * @param {Object} data - { processorIndex, input, options }
 */
function handleRunPreProcessor(id, data) {
  const { processorIndex, input, options } = data || {};

  if (processorIndex === undefined || processorIndex < 0) {
    sendResponse(id, false, null, 'Processor index is required');
    return;
  }

  if (processorIndex >= registeredPreProcessors.length) {
    sendResponse(id, false, null, `Pre-processor index out of range: ${processorIndex}`);
    return;
  }

  try {
    const proc = registeredPreProcessors[processorIndex];
    const processor = proc.processor;

    // Build extra context object (matching less.js behavior)
    const extra = {
      context: options || {},
      fileInfo: options?.fileInfo || {},
      imports: options?.imports || {},
    };

    // Call the processor's process method
    let output;
    if (typeof processor.process === 'function') {
      output = processor.process(input, extra);
    } else if (typeof processor === 'function') {
      output = processor(input, extra);
    } else {
      sendResponse(id, false, null, 'Pre-processor does not have a process method');
      return;
    }

    // Handle promise result
    if (output && typeof output.then === 'function') {
      output.then((result) => {
        sendResponse(id, true, { output: result });
      }).catch((err) => {
        sendResponse(id, false, null, `Pre-processor error: ${err.message}`);
      });
    } else {
      sendResponse(id, true, { output });
    }
  } catch (err) {
    sendResponse(id, false, null, `Pre-processor error: ${err.message}\n${err.stack || ''}`);
  }
}

/**
 * Run a post-processor on the CSS output
 * @param {number} id - Command ID
 * @param {Object} data - { processorIndex, input, options }
 */
function handleRunPostProcessor(id, data) {
  const { processorIndex, input, options } = data || {};

  if (processorIndex === undefined || processorIndex < 0) {
    sendResponse(id, false, null, 'Processor index is required');
    return;
  }

  if (processorIndex >= registeredPostProcessors.length) {
    sendResponse(id, false, null, `Post-processor index out of range: ${processorIndex}`);
    return;
  }

  try {
    const proc = registeredPostProcessors[processorIndex];
    const processor = proc.processor;

    // Build extra context object (matching less.js behavior)
    const extra = {
      context: options || {},
      fileInfo: options?.fileInfo || {},
      imports: options?.imports || {},
    };

    // Call the processor's process method
    let output;
    if (typeof processor.process === 'function') {
      output = processor.process(input, extra);
    } else if (typeof processor === 'function') {
      output = processor(input, extra);
    } else {
      sendResponse(id, false, null, 'Post-processor does not have a process method');
      return;
    }

    // Handle promise result
    if (output && typeof output.then === 'function') {
      output.then((result) => {
        sendResponse(id, true, { output: result });
      }).catch((err) => {
        sendResponse(id, false, null, `Post-processor error: ${err.message}`);
      });
    } else {
      sendResponse(id, true, { output });
    }
  } catch (err) {
    sendResponse(id, false, null, `Post-processor error: ${err.message}\n${err.stack || ''}`);
  }
}

/**
 * Get the list of registered file managers
 * @param {number} id - Command ID
 * @param {Object} data - Unused
 */
function handleGetFileManagers(id, data) {
  try {
    const managers = registeredFileManagers.map((fm, index) => ({
      index,
    }));
    sendResponse(id, true, managers);
  } catch (err) {
    sendResponse(id, false, null, `Failed to get file managers: ${err.message}`);
  }
}

/**
 * Check if a file manager supports a given file
 * @param {number} id - Command ID
 * @param {Object} data - { managerIndex, filename, currentDirectory, options }
 */
function handleFileManagerSupports(id, data) {
  const { managerIndex, filename, currentDirectory, options } = data || {};

  if (managerIndex === undefined || managerIndex < 0) {
    sendResponse(id, false, null, 'File manager index is required');
    return;
  }

  if (managerIndex >= registeredFileManagers.length) {
    sendResponse(id, false, null, `File manager index out of range: ${managerIndex}`);
    return;
  }

  try {
    const fileManager = registeredFileManagers[managerIndex];

    // Call the supports method if it exists
    let supports = false;
    if (typeof fileManager.supports === 'function') {
      supports = fileManager.supports(filename, currentDirectory, options || {}, {});
    } else if (typeof fileManager.supportsSync === 'function') {
      supports = fileManager.supportsSync(filename, currentDirectory, options || {}, {});
    } else {
      // If no supports method, assume it supports all files
      supports = true;
    }

    sendResponse(id, true, { supports: !!supports });
  } catch (err) {
    sendResponse(id, false, null, `File manager supports check error: ${err.message}`);
  }
}

/**
 * Load a file using a file manager
 * @param {number} id - Command ID
 * @param {Object} data - { managerIndex, filename, currentDirectory, options }
 */
function handleFileManagerLoad(id, data) {
  const { managerIndex, filename, currentDirectory, options } = data || {};

  if (managerIndex === undefined || managerIndex < 0) {
    sendResponse(id, false, null, 'File manager index is required');
    return;
  }

  if (managerIndex >= registeredFileManagers.length) {
    sendResponse(id, false, null, `File manager index out of range: ${managerIndex}`);
    return;
  }

  try {
    const fileManager = registeredFileManagers[managerIndex];

    // Try loadFile or loadFileSync
    let result;
    if (typeof fileManager.loadFile === 'function') {
      result = fileManager.loadFile(filename, currentDirectory, options || {}, {});
    } else if (typeof fileManager.loadFileSync === 'function') {
      result = fileManager.loadFileSync(filename, currentDirectory, options || {}, {});
    } else {
      sendResponse(id, false, null, 'File manager does not have a loadFile method');
      return;
    }

    // Handle promise result
    if (result && typeof result.then === 'function') {
      result.then((data) => {
        sendResponse(id, true, {
          filename: data.filename || filename,
          contents: data.contents || '',
        });
      }).catch((err) => {
        sendResponse(id, false, null, `File load error: ${err.message}`);
      });
    } else {
      sendResponse(id, true, {
        filename: result.filename || filename,
        contents: result.contents || '',
      });
    }
  } catch (err) {
    sendResponse(id, false, null, `File manager load error: ${err.message}\n${err.stack || ''}`);
  }
}

// Export for testing
if (typeof module !== 'undefined') {
  module.exports = {
    parseFlatAST,
    getAttachedAST,
    attachedBuffers,
    registeredFunctions,
    registeredVisitors,
    registeredPreProcessors,
    registeredPostProcessors,
    registeredFileManagers,
    bindings,
    less,
    tree,
    functionRegistry,
    pluginManager,
  };
}

// Set up readline interface for stdin
const rl = readline.createInterface({
  input: process.stdin,
  output: null, // Don't echo to stdout
  terminal: false,
});

// Handle incoming lines
rl.on('line', (line) => {
  if (!line.trim()) return;

  try {
    const cmd = JSON.parse(line);
    handleCommand(cmd);
  } catch (err) {
    // Can't send response without ID, log to stderr
    process.stderr.write(`Failed to parse command: ${err.message}\n`);
  }
});

// Handle stdin close (Go process closed the pipe)
rl.on('close', () => {
  process.exit(0);
});

// Handle uncaught errors
process.on('uncaughtException', (err) => {
  process.stderr.write(`Uncaught exception: ${err.message}\n${err.stack}\n`);
  process.exit(1);
});

process.on('unhandledRejection', (reason, promise) => {
  process.stderr.write(`Unhandled rejection: ${reason}\n`);
});

// Signal that we're ready by not doing anything special
// The Go side will send a ping to verify readiness
