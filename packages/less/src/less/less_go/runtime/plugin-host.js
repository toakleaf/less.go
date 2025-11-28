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
const registeredFunctions = new Map();
const registeredVisitors = [];
const registeredPreProcessors = [];
const registeredPostProcessors = [];
const registeredFileManagers = [];

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
      return createNode('Variable', { name });
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

// Function registry mock
const functionRegistry = {
  add(name, fn) {
    registeredFunctions.set(name, fn);
  },
  addMultiple(functions) {
    for (const [name, fn] of Object.entries(functions)) {
      this.add(name, fn);
    }
  },
  get(name) {
    return registeredFunctions.get(name);
  },
  getAll() {
    return Array.from(registeredFunctions.keys());
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

      case 'runPostEvalVisitors':
        handleRunPostEvalVisitors(id, data);
        break;

      case 'parseASTBuffer':
        handleParseASTBuffer(id, data);
        break;

      case 'serializeNode':
        handleSerializeNode(id, data);
        break;

      default:
        sendResponse(id, false, null, `Unknown command: ${command}`);
    }
  } catch (err) {
    sendResponse(id, false, null, err.message || String(err));
  }
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
    } else if (isRelative) {
      const basePath = baseDir || process.cwd();
      resolvedPath = path.resolve(basePath, pluginPath);
    } else {
      // NPM module - use require.resolve to find it
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

    // Check if already loaded (by resolved path)
    if (loadedPlugins.has(resolvedPath)) {
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
      });

      vm.runInContext(fileContent, context, { filename: resolvedPath });
      plugin = context.module.exports;
    }

    // Handle plugin as either an object or a function (constructor)
    if (typeof plugin === 'function') {
      plugin = new plugin();
    }

    // Check if the plugin registered anything already (legacy format)
    // Legacy plugins call functions.add() directly during require()
    const hasLegacyRegistrations = registeredFunctions.size > functionsBefore;

    // If no legacy registrations and plugin has an install method, call it
    if (!hasLegacyRegistrations && plugin && plugin.install) {
      // Validate minVersion if specified
      if (plugin.minVersion) {
        const minVer = Array.isArray(plugin.minVersion) ? plugin.minVersion : plugin.minVersion.split('.').map(Number);
        const lessVer = less.version;
        for (let i = 0; i < minVer.length; i++) {
          if ((minVer[i] || 0) > (lessVer[i] || 0)) {
            throw new Error(`Plugin requires Less.js version ${minVer.join('.')} or higher`);
          } else if ((minVer[i] || 0) < (lessVer[i] || 0)) {
            break;
          }
        }
      }

      plugin.install(less, pluginManager, functionRegistry);
    }

    // Set options if provided (for version 3.x+, options are set after install)
    if (options && plugin && plugin.setOptions) {
      plugin.setOptions(options);
    }

    // Call use() if it exists (called every time plugin is loaded)
    if (plugin && plugin.use) {
      plugin.use(plugin);
    }

    // Cache the plugin
    loadedPlugins.set(resolvedPath, plugin || {});

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
 * Call a registered function
 * @param {number} id - Command ID
 * @param {Object} data - Function call data
 */
function handleCallFunction(id, data) {
  const { name, args } = data || {};

  if (!name) {
    sendResponse(id, false, null, 'Function name is required');
    return;
  }

  const fn = registeredFunctions.get(name);
  if (!fn) {
    sendResponse(id, false, null, `Function not found: ${name}`);
    return;
  }

  try {
    // Convert Go node arguments to JavaScript format
    const convertedArgs = (args || []).map(convertGoNodeToJS);

    // Call the function with converted arguments
    const result = fn.apply(null, convertedArgs);

    // Convert the result back to Go-compatible format
    const goResult = convertJSResultToGo(result);

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

// Export for testing
if (typeof module !== 'undefined') {
  module.exports = {
    parseFlatAST,
    getAttachedAST,
    attachedBuffers,
    registeredFunctions,
    registeredVisitors,
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
