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

// Plugin state
const loadedPlugins = new Map();
const registeredFunctions = new Map();
const registeredVisitors = [];
const registeredPreProcessors = [];
const registeredPostProcessors = [];
const registeredFileManagers = [];

// Shared memory state
const attachedBuffers = new Map(); // key -> { path, size, buffer }

// Less API mock (to be expanded)
const less = {
  version: [4, 0, 0],
  // Node constructors will be added later
};

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
    // Resolve the plugin path
    let resolvedPath;
    if (path.isAbsolute(pluginPath)) {
      resolvedPath = pluginPath;
    } else if (baseDir) {
      resolvedPath = path.resolve(baseDir, pluginPath);
    } else {
      resolvedPath = path.resolve(process.cwd(), pluginPath);
    }

    // Check if already loaded
    if (loadedPlugins.has(resolvedPath)) {
      const plugin = loadedPlugins.get(resolvedPath);
      sendResponse(id, true, {
        cached: true,
        functions: functionRegistry.getAll(),
      });
      return;
    }

    // Load the plugin using require
    const plugin = require(resolvedPath);

    // Call install if it exists
    if (plugin.install) {
      plugin.install(less, pluginManager, functionRegistry);
    }

    // Set options if provided
    if (options && plugin.setOptions) {
      plugin.setOptions(options);
    }

    // Cache the plugin
    loadedPlugins.set(resolvedPath, plugin);

    sendResponse(id, true, {
      cached: false,
      functions: functionRegistry.getAll(),
      visitors: registeredVisitors.length,
      preProcessors: registeredPreProcessors.length,
      postProcessors: registeredPostProcessors.length,
      fileManagers: registeredFileManagers.length,
    });
  } catch (err) {
    sendResponse(id, false, null, `Failed to load plugin: ${err.message}`);
  }
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
    const result = fn.apply(null, args || []);
    sendResponse(id, true, result);
  } catch (err) {
    sendResponse(id, false, null, `Function error: ${err.message}`);
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

// Export for testing
if (typeof module !== 'undefined') {
  module.exports = {
    parseFlatAST,
    getAttachedAST,
    attachedBuffers,
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
