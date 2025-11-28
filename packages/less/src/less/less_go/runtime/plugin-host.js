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

// Plugin state
const loadedPlugins = new Map();
const registeredFunctions = new Map();
const registeredVisitors = [];
const registeredPreProcessors = [];
const registeredPostProcessors = [];
const registeredFileManagers = [];

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
