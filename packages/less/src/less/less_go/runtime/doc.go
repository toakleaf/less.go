// Package runtime provides JavaScript execution capabilities for LESS plugins.
//
// This package implements a bridge between Go and Node.js, allowing LESS plugins
// written in JavaScript to be executed while maintaining performance through
// shared memory buffer transfer.
//
// Architecture Overview
//
// The runtime uses a hybrid approach combining Node.js for JavaScript execution
// with shared memory for efficient data transfer:
//
//   1. Go spawns a persistent Node.js process running plugin-host.js
//   2. Commands are sent via stdin/stdout (JSON protocol)
//   3. Large data (ASTs) is transferred via shared memory buffers
//   4. Node.js executes plugin JavaScript and writes results back
//
// This approach provides:
//   - Perfect JavaScript compatibility (real Node.js)
//   - Fast V8 execution (10x faster than embedded runtimes like goja)
//   - Zero-copy data transfer (no JSON serialization for ASTs)
//   - Simple plugin loading (native require())
//
// Usage
//
//	rt, err := runtime.NewNodeJSRuntime()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer rt.Stop()
//
//	// Send a command
//	response, err := rt.SendCommand(runtime.Command{
//	    Cmd: "ping",
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// IPC Protocol
//
// Commands are JSON objects sent via stdin:
//
//	{"id": 1, "cmd": "ping"}
//	{"id": 2, "cmd": "loadPlugin", "path": "./plugin.js"}
//	{"id": 3, "cmd": "callFunction", "functionID": "myFunc", "args": [...]}
//
// Responses are JSON objects received via stdout:
//
//	{"id": 1, "success": true, "result": "pong"}
//	{"id": 2, "success": true, "result": {"functions": ["myFunc"]}}
//	{"id": 3, "success": false, "error": "function not found"}
package runtime
