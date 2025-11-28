#!/usr/bin/env node

/**
 * Plugin Host - Node.js side of the JavaScript plugin runtime
 *
 * This script runs in a Node.js process spawned by the Go runtime.
 * It receives commands via stdin (JSON, one per line) and responds via stdout.
 *
 * Commands:
 * - ping: Test connectivity
 * - echo: Echo back a message (for testing)
 * - load_plugin: Load a JavaScript plugin file
 * - call_function: Call a plugin function
 *
 * Architecture:
 * - Commands are JSON objects with {id, type, payload}
 * - Responses are JSON objects with {id, success, result?, error?}
 * - Each command/response is on a single line
 * - Errors go to stderr for debugging
 */

'use strict';

const fs = require('fs');
const path = require('path');
const readline = require('readline');

// Global state
const loadedPlugins = new Map();
const registeredFunctions = new Map();
let nextPluginID = 1;

// Set up readline interface for stdin/stdout
const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout,
  terminal: false
});

/**
 * Send a response back to Go
 */
function sendResponse(id, success, result = null, error = null) {
  const response = {
    id,
    success,
    ...(result && { result }),
    ...(error && { error })
  };
  console.log(JSON.stringify(response));
}

/**
 * Send an error message to stderr (for debugging)
 */
function logError(message, error) {
  console.error(`[plugin-host] ${message}`, error || '');
}

/**
 * Handle ping command
 */
function handlePing(id, payload) {
  sendResponse(id, true, { message: 'pong' });
}

/**
 * Handle echo command (for testing)
 */
function handleEcho(id, payload) {
  if (!payload || !payload.message) {
    sendResponse(id, false, null, 'Missing message in payload');
    return;
  }
  sendResponse(id, true, { echo: payload.message });
}

/**
 * Handle load_plugin command
 */
function handleLoadPlugin(id, payload) {
  try {
    const { pluginPath, baseDir, options } = payload;

    if (!pluginPath) {
      sendResponse(id, false, null, 'Missing pluginPath in payload');
      return;
    }

    // Resolve the plugin path
    let resolvedPath;
    if (path.isAbsolute(pluginPath)) {
      resolvedPath = pluginPath;
    } else if (baseDir) {
      resolvedPath = path.resolve(baseDir, pluginPath);
    } else {
      resolvedPath = path.resolve(pluginPath);
    }

    // Check if already loaded
    if (loadedPlugins.has(resolvedPath)) {
      const pluginInfo = loadedPlugins.get(resolvedPath);
      sendResponse(id, true, {
        pluginID: pluginInfo.id,
        functions: Array.from(pluginInfo.functions.keys()),
        cached: true
      });
      return;
    }

    // Create plugin context
    const pluginID = `plugin-${nextPluginID++}`;
    const pluginFunctions = new Map();

    // Create a mock functions registry
    const functionsRegistry = {
      add: (name, fn) => {
        const fnID = `${pluginID}::${name}`;
        pluginFunctions.set(name, { id: fnID, fn });
        registeredFunctions.set(fnID, fn);
      },
      addMultiple: (fns) => {
        for (const [name, fn] of Object.entries(fns)) {
          functionsRegistry.add(name, fn);
        }
      }
    };

    // Create a mock less object with basic node constructors
    const less = {
      dimension: (value, unit) => ({ type: 'Dimension', value, unit }),
      color: (rgb, alpha) => ({ type: 'Color', rgb, alpha }),
      quoted: (quote, value, escaped) => ({ type: 'Quoted', quote, value, escaped }),
      keyword: (value) => ({ type: 'Keyword', value }),
      anonymous: (value) => ({ type: 'Anonymous', value }),
      // Add more constructors as needed
    };

    // Create mock plugin manager
    const pluginManager = {
      addVisitor: (visitor) => {
        // TODO: Implement visitor support
        logError('Visitor support not yet implemented:', visitor);
      },
      addPreProcessor: (processor, priority) => {
        // TODO: Implement preprocessor support
        logError('PreProcessor support not yet implemented');
      },
      addPostProcessor: (processor, priority) => {
        // TODO: Implement postprocessor support
        logError('PostProcessor support not yet implemented');
      },
      addFileManager: (manager) => {
        // TODO: Implement file manager support
        logError('FileManager support not yet implemented');
      }
    };

    // Set up global objects BEFORE requiring the plugin
    // Some plugins use global.functions directly at module level
    global.functions = functionsRegistry;
    global.less = less;

    // Load the plugin using require()
    // Note: Node.js require() handles:
    // - npm modules (node_modules resolution)
    // - relative paths (./plugin.js)
    // - absolute paths (/full/path.js)
    // - package.json main field
    // - .js extension inference
    const plugin = require(resolvedPath);

    // If the plugin has an install function, call it
    if (typeof plugin.install === 'function') {
      plugin.install(less, pluginManager, functionsRegistry);
    }

    // Call setOptions if provided
    if (options && typeof plugin.setOptions === 'function') {
      plugin.setOptions(options);
    }

    // Store plugin info
    loadedPlugins.set(resolvedPath, {
      id: pluginID,
      plugin,
      functions: pluginFunctions
    });

    // Return plugin info
    sendResponse(id, true, {
      pluginID,
      functions: Array.from(pluginFunctions.keys()),
      cached: false
    });

  } catch (error) {
    logError('Error loading plugin:', error);
    sendResponse(id, false, null, error.message);
  }
}

/**
 * Handle call_function command
 */
function handleCallFunction(id, payload) {
  try {
    const { functionID, args } = payload;

    if (!functionID) {
      sendResponse(id, false, null, 'Missing functionID in payload');
      return;
    }

    const fn = registeredFunctions.get(functionID);
    if (!fn) {
      sendResponse(id, false, null, `Function not found: ${functionID}`);
      return;
    }

    // Call the function with args
    const result = fn(...(args || []));

    sendResponse(id, true, { result });

  } catch (error) {
    logError('Error calling function:', error);
    sendResponse(id, false, null, error.message);
  }
}

/**
 * Main command handler
 */
function handleCommand(line) {
  try {
    const command = JSON.parse(line);
    const { id, type, payload } = command;

    if (!id || !type) {
      logError('Invalid command:', line);
      return;
    }

    switch (type) {
      case 'ping':
        handlePing(id, payload);
        break;
      case 'echo':
        handleEcho(id, payload);
        break;
      case 'load_plugin':
        handleLoadPlugin(id, payload);
        break;
      case 'call_function':
        handleCallFunction(id, payload);
        break;
      default:
        sendResponse(id, false, null, `Unknown command type: ${type}`);
    }

  } catch (error) {
    logError('Error handling command:', error);
  }
}

// Process commands from stdin
rl.on('line', handleCommand);

// Handle errors
rl.on('error', (error) => {
  logError('Readline error:', error);
});

// Handle close
rl.on('close', () => {
  process.exit(0);
});

// Log startup
logError('Plugin host started, waiting for commands...');
