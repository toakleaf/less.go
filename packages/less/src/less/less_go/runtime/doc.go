// Package runtime provides JavaScript execution capabilities for LESS plugins.
//
// This package implements the bridge between Go and JavaScript, allowing
// LESS plugins written in JavaScript to be executed from the Go runtime.
//
// The runtime uses a Node.js process with shared memory to achieve:
//   - Perfect JavaScript compatibility (real Node.js with V8)
//   - Zero-copy data transfer (shared memory buffers)
//   - Fast execution (V8 engine)
//   - Simple plugin loading (native require())
//
// Architecture:
//   - Go spawns a Node.js process running plugin-host.js
//   - Commands sent over stdin/stdout (JSON)
//   - Large data (AST buffers) passed via shared memory
//   - Node.js process kept alive across compilations
//
// This approach is inspired by OXC's buffer-based serialization combined
// with the reliability and compatibility of Node.js.
package runtime
