package runtime

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Command represents a command sent to the Node.js process.
type Command struct {
	ID   int64  `json:"id"`
	Cmd  string `json:"cmd"`
	Data any    `json:"data,omitempty"`
}

// Response represents a response from the Node.js process.
type Response struct {
	ID      int64  `json:"id"`
	Success bool   `json:"success"`
	Result  any    `json:"result,omitempty"`
	Error   string `json:"error,omitempty"`
}

// CallbackRequest represents a callback request from Node.js to Go.
// This is used for on-demand variable lookup during function execution.
type CallbackRequest struct {
	ID       int64  `json:"id"`
	Callback string `json:"callback"`
	Data     any    `json:"data,omitempty"`
}

// CallbackHandler is a function that handles callbacks from Node.js.
type CallbackHandler func(data any) (any, error)

// NodeJSRuntime manages a Node.js process for JavaScript plugin execution.
type NodeJSRuntime struct {
	process *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
	stderr  io.ReadCloser

	// Response handling
	responses   map[int64]chan Response
	responsesMu sync.RWMutex

	// Callback handling for on-demand operations
	callbackHandlers   map[string]CallbackHandler
	callbackHandlersMu sync.RWMutex

	// Command ID counter
	cmdID atomic.Int64

	// State
	alive   atomic.Bool
	started atomic.Bool

	// Shutdown coordination
	done     chan struct{}
	wg       sync.WaitGroup
	stderrWg sync.WaitGroup

	// Error from background goroutines
	errMu sync.RWMutex
	err   error

	// Configuration
	pluginHostPath string
	nodeCommand    string

	// Shared memory for zero-copy AST transfer
	shmManager *SharedMemoryManager

	// Reusable prefetch buffer for plugin function calls
	prefetchShm   *SharedMemory
	prefetchShmMu sync.Mutex

	// Function result cache - shared across all JSFunctionDefinitions
	// Key format: "funcName:arg1|arg2|..."
	funcResultCache   map[string]any
	funcResultCacheMu sync.RWMutex

	// High-performance shared memory protocol (optional)
	shmProtocol   *SharedMemoryProtocol
	shmProtocolMu sync.Mutex
	useSHMProtocol bool // Whether to use the binary SHM protocol
}

// RuntimeOption configures a NodeJSRuntime.
type RuntimeOption func(*NodeJSRuntime)

// WithPluginHostPath sets the path to the plugin-host.js file.
func WithPluginHostPath(path string) RuntimeOption {
	return func(rt *NodeJSRuntime) {
		rt.pluginHostPath = path
	}
}

// WithNodeCommand sets the Node.js command to use (default: "node").
func WithNodeCommand(cmd string) RuntimeOption {
	return func(rt *NodeJSRuntime) {
		rt.nodeCommand = cmd
	}
}

// NewNodeJSRuntime creates a new Node.js runtime instance.
//
// The runtime is not started automatically. Call Start() to spawn the Node.js process.
func NewNodeJSRuntime(opts ...RuntimeOption) (*NodeJSRuntime, error) {
	rt := &NodeJSRuntime{
		responses:        make(map[int64]chan Response),
		callbackHandlers: make(map[string]CallbackHandler),
		done:             make(chan struct{}),
		nodeCommand:      "node",
		funcResultCache:  make(map[string]any),
	}

	// Apply options
	for _, opt := range opts {
		opt(rt)
	}

	// Find plugin-host.js if not specified
	if rt.pluginHostPath == "" {
		// Look for plugin-host.js relative to this package
		candidates := []string{
			"plugin-host.js",
			filepath.Join("runtime", "plugin-host.js"),
		}

		// Try to find it relative to the executable or working directory
		execPath, err := os.Executable()
		if err == nil {
			execDir := filepath.Dir(execPath)
			candidates = append(candidates,
				filepath.Join(execDir, "plugin-host.js"),
				filepath.Join(execDir, "runtime", "plugin-host.js"),
			)
		}

		// Also check relative to current working directory
		cwd, err := os.Getwd()
		if err == nil {
			candidates = append(candidates,
				filepath.Join(cwd, "packages", "less", "src", "less", "less_go", "runtime", "plugin-host.js"),
			)
		}

		// Find first existing file
		for _, candidate := range candidates {
			if _, err := os.Stat(candidate); err == nil {
				rt.pluginHostPath = candidate
				break
			}
		}

		if rt.pluginHostPath == "" {
			return nil, fmt.Errorf("plugin-host.js not found; tried: %v", candidates)
		}
	}

	return rt, nil
}

// Start spawns the Node.js process and begins handling IPC.
func (rt *NodeJSRuntime) Start() error {
	if rt.started.Load() {
		return fmt.Errorf("runtime already started")
	}

	// Initialize shared memory manager
	shmManager, err := NewSharedMemoryManager()
	if err != nil {
		return fmt.Errorf("failed to create shared memory manager: %w", err)
	}
	rt.shmManager = shmManager

	// Create the Node.js process
	rt.process = exec.Command(rt.nodeCommand, rt.pluginHostPath)
	rt.process.Env = append(os.Environ(), "LESS_PLUGIN_HOST=1")

	// Set up stdio pipes
	rt.stdin, err = rt.process.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	rt.stdout, err = rt.process.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	rt.stderr, err = rt.process.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the process
	if err := rt.process.Start(); err != nil {
		return fmt.Errorf("failed to start Node.js process: %w", err)
	}

	rt.started.Store(true)
	rt.alive.Store(true)

	// Start response reader goroutine
	rt.wg.Add(1)
	go rt.readResponses()

	// Start stderr reader goroutine
	rt.stderrWg.Add(1)
	go rt.readStderr()

	// Wait for process to be ready with a ping
	ctx, cancel := contextWithTimeout(5 * time.Second)
	defer cancel()

	if err := rt.waitReady(ctx); err != nil {
		rt.Stop()
		return fmt.Errorf("Node.js process failed to start: %w", err)
	}

	return nil
}

// waitReady sends a ping and waits for a response to verify the process is ready.
func (rt *NodeJSRuntime) waitReady(ctx context) error {
	resp, err := rt.SendCommandWithContext(ctx, Command{Cmd: "ping"})
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("ping failed: %s", resp.Error)
	}
	return nil
}

// Stop gracefully shuts down the Node.js process.
func (rt *NodeJSRuntime) Stop() error {
	if !rt.started.Load() {
		return nil
	}

	rt.alive.Store(false)
	close(rt.done)

	// Clean up shared memory
	if rt.shmManager != nil {
		rt.shmManager.DestroyAll()
		rt.shmManager = nil
	}

	// Send shutdown command (best effort)
	if rt.stdin != nil {
		cmd := Command{ID: rt.cmdID.Add(1), Cmd: "shutdown"}
		data, _ := json.Marshal(cmd)
		rt.stdin.Write(append(data, '\n'))
	}

	// Close stdin to signal EOF
	if rt.stdin != nil {
		rt.stdin.Close()
	}

	// Wait for response reader to finish
	rt.wg.Wait()

	// Wait for stderr reader to finish
	rt.stderrWg.Wait()

	// Wait for process to exit (with timeout)
	done := make(chan error, 1)
	go func() {
		done <- rt.process.Wait()
	}()

	select {
	case err := <-done:
		// Process exited normally or with error
		if err != nil {
			// Check if it's just a non-zero exit code (normal for shutdown)
			if _, ok := err.(*exec.ExitError); ok {
				return nil
			}
			return err
		}
		return nil
	case <-time.After(5 * time.Second):
		// Force kill if it doesn't exit gracefully
		rt.process.Process.Kill()
		return fmt.Errorf("Node.js process did not exit gracefully, killed")
	}
}

// IsAlive returns true if the Node.js process is running.
func (rt *NodeJSRuntime) IsAlive() bool {
	return rt.alive.Load()
}

// SendCommand sends a command to the Node.js process and waits for a response.
func (rt *NodeJSRuntime) SendCommand(cmd Command) (Response, error) {
	ctx, cancel := contextWithTimeout(30 * time.Second)
	defer cancel()
	return rt.SendCommandWithContext(ctx, cmd)
}

// SendCommandWithContext sends a command with a context for timeout/cancellation.
func (rt *NodeJSRuntime) SendCommandWithContext(ctx context, cmd Command) (Response, error) {
	if !rt.alive.Load() {
		return Response{}, fmt.Errorf("runtime not alive")
	}

	// Assign command ID
	cmd.ID = rt.cmdID.Add(1)

	// Create response channel
	respChan := make(chan Response, 1)
	rt.responsesMu.Lock()
	rt.responses[cmd.ID] = respChan
	rt.responsesMu.Unlock()

	defer func() {
		rt.responsesMu.Lock()
		delete(rt.responses, cmd.ID)
		rt.responsesMu.Unlock()
	}()

	// Serialize and send command
	data, err := json.Marshal(cmd)
	if err != nil {
		return Response{}, fmt.Errorf("failed to marshal command: %w", err)
	}

	// Write command with newline delimiter
	if _, err := rt.stdin.Write(append(data, '\n')); err != nil {
		return Response{}, fmt.Errorf("failed to send command: %w", err)
	}

	// Wait for response
	select {
	case resp := <-respChan:
		return resp, nil
	case <-ctx.done():
		return Response{}, fmt.Errorf("command timed out")
	case <-rt.done:
		return Response{}, fmt.Errorf("runtime shutting down")
	}
}

// SendCommandFireAndForget sends a command without waiting for a response.
// This is useful for commands that don't need a response or when the response
// can be safely ignored. The command is still sent with an ID for logging purposes.
//
// CRITICAL: Only use this for idempotent operations where:
// 1. The response is not needed by the caller
// 2. Order of execution is not critical
// 3. Failure can be tolerated or will be detected later
//
// This dramatically reduces IPC latency for operations like scope management
// where we send hundreds of thousands of updates during Bootstrap4 compilation.
func (rt *NodeJSRuntime) SendCommandFireAndForget(cmd Command) error {
	if !rt.alive.Load() {
		return fmt.Errorf("runtime not alive")
	}

	// Assign command ID (for logging on Node.js side)
	cmd.ID = rt.cmdID.Add(1)

	// Serialize and send command
	data, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("failed to marshal command: %w", err)
	}

	// Write command with newline delimiter
	if _, err := rt.stdin.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to send command: %w", err)
	}

	return nil
}

// readResponses reads responses from stdout and dispatches them.
// It also handles callback requests from Node.js for on-demand operations.
func (rt *NodeJSRuntime) readResponses() {
	defer rt.wg.Done()

	scanner := bufio.NewScanner(rt.stdout)
	// Increase buffer size for large responses
	const maxScanTokenSize = 10 * 1024 * 1024 // 10MB
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, maxScanTokenSize)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		// First, try to parse as a generic message to check the type
		var msg map[string]any
		if err := json.Unmarshal(line, &msg); err != nil {
			rt.setError(fmt.Errorf("failed to parse message: %w", err))
			continue
		}

		// Check if this is a callback request (has "callback" field)
		if _, hasCallback := msg["callback"]; hasCallback {
			var cbReq CallbackRequest
			if err := json.Unmarshal(line, &cbReq); err != nil {
				rt.setError(fmt.Errorf("failed to parse callback request: %w", err))
				continue
			}
			rt.handleCallback(cbReq)
			continue
		}

		// Otherwise, it's a regular response
		var resp Response
		if err := json.Unmarshal(line, &resp); err != nil {
			rt.setError(fmt.Errorf("failed to parse response: %w", err))
			continue
		}

		// Dispatch response to waiting goroutine
		rt.responsesMu.RLock()
		if ch, ok := rt.responses[resp.ID]; ok {
			select {
			case ch <- resp:
			default:
				// Channel full, response was probably already timed out
			}
		}
		rt.responsesMu.RUnlock()
	}

	if err := scanner.Err(); err != nil && rt.alive.Load() {
		rt.setError(fmt.Errorf("stdout read error: %w", err))
	}

	rt.alive.Store(false)
}

// handleCallback handles a callback request from Node.js.
func (rt *NodeJSRuntime) handleCallback(req CallbackRequest) {
	rt.callbackHandlersMu.RLock()
	handler, ok := rt.callbackHandlers[req.Callback]
	rt.callbackHandlersMu.RUnlock()

	if !ok {
		rt.sendCallbackResponse(req.ID, false, nil, fmt.Sprintf("no handler for callback: %s", req.Callback))
		return
	}

	result, err := handler(req.Data)
	if err != nil {
		rt.sendCallbackResponse(req.ID, false, nil, err.Error())
		return
	}

	rt.sendCallbackResponse(req.ID, true, result, "")
}

// sendCallbackResponse sends a response to a callback request.
func (rt *NodeJSRuntime) sendCallbackResponse(id int64, success bool, result any, errMsg string) {
	resp := Response{
		ID:      id,
		Success: success,
		Result:  result,
		Error:   errMsg,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		rt.setError(fmt.Errorf("failed to marshal callback response: %w", err))
		return
	}

	if _, err := rt.stdin.Write(append(data, '\n')); err != nil {
		rt.setError(fmt.Errorf("failed to send callback response: %w", err))
	}
}

// RegisterCallback registers a callback handler for a specific callback type.
func (rt *NodeJSRuntime) RegisterCallback(name string, handler CallbackHandler) {
	rt.callbackHandlersMu.Lock()
	rt.callbackHandlers[name] = handler
	rt.callbackHandlersMu.Unlock()
}

// UnregisterCallback removes a callback handler.
func (rt *NodeJSRuntime) UnregisterCallback(name string) {
	rt.callbackHandlersMu.Lock()
	delete(rt.callbackHandlers, name)
	rt.callbackHandlersMu.Unlock()
}

// readStderr reads and logs stderr output from the Node.js process.
func (rt *NodeJSRuntime) readStderr() {
	defer rt.stderrWg.Done()

	scanner := bufio.NewScanner(rt.stderr)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			// Print debug output if LESS_GO_DEBUG is set
			if os.Getenv("LESS_GO_DEBUG") == "1" {
				fmt.Fprintf(os.Stderr, "%s\n", line)
			}
			// Store errors (non-debug lines)
			if !strings.HasPrefix(line, "[plugin-host]") && !strings.HasPrefix(line, "[DEBUG") {
				rt.setError(fmt.Errorf("Node.js stderr: %s", line))
			}
		}
	}
}

// setError sets the runtime error in a thread-safe manner.
func (rt *NodeJSRuntime) setError(err error) {
	rt.errMu.Lock()
	defer rt.errMu.Unlock()
	if rt.err == nil {
		rt.err = err
	}
}

// Error returns any error from the runtime's background operations.
func (rt *NodeJSRuntime) Error() error {
	rt.errMu.RLock()
	defer rt.errMu.RUnlock()
	return rt.err
}

// Ping sends a ping command to verify the Node.js process is responsive.
func (rt *NodeJSRuntime) Ping() error {
	resp, err := rt.SendCommand(Command{Cmd: "ping"})
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("ping failed: %s", resp.Error)
	}
	return nil
}

// Echo sends a value to Node.js and expects it back (for testing).
func (rt *NodeJSRuntime) Echo(value any) (any, error) {
	resp, err := rt.SendCommand(Command{
		Cmd:  "echo",
		Data: value,
	})
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf("echo failed: %s", resp.Error)
	}
	return resp.Result, nil
}

// Simple context interface to avoid importing context package
// (for environments where context is not available)
type context interface {
	done() <-chan struct{}
}

type timeoutContext struct {
	ch      chan struct{}
	timeout time.Duration
}

func contextWithTimeout(timeout time.Duration) (*timeoutContext, func()) {
	ctx := &timeoutContext{
		ch:      make(chan struct{}),
		timeout: timeout,
	}

	timer := time.AfterFunc(timeout, func() {
		close(ctx.ch)
	})

	cancel := func() {
		timer.Stop()
	}

	return ctx, cancel
}

func (c *timeoutContext) done() <-chan struct{} {
	return c.ch
}

// SharedMemoryManager returns the shared memory manager for this runtime.
func (rt *NodeJSRuntime) SharedMemoryManager() *SharedMemoryManager {
	return rt.shmManager
}

// CreateSharedMemory creates a new shared memory segment of the specified size.
func (rt *NodeJSRuntime) CreateSharedMemory(size int) (*SharedMemory, error) {
	if rt.shmManager == nil {
		return nil, fmt.Errorf("shared memory manager not initialized")
	}
	return rt.shmManager.Create(size)
}

// DestroySharedMemory destroys a shared memory segment by key.
func (rt *NodeJSRuntime) DestroySharedMemory(shm *SharedMemory) error {
	if rt.shmManager == nil {
		return fmt.Errorf("shared memory manager not initialized")
	}
	return rt.shmManager.Destroy(shm.Key())
}

// GetPrefetchBuffer returns a reusable shared memory buffer for prefetch data.
// The buffer is created on first use and reused across calls.
// If the required size is larger than the current buffer, it's resized.
func (rt *NodeJSRuntime) GetPrefetchBuffer(size int) (*SharedMemory, error) {
	rt.prefetchShmMu.Lock()
	defer rt.prefetchShmMu.Unlock()

	if rt.shmManager == nil {
		return nil, fmt.Errorf("shared memory manager not initialized")
	}

	// Minimum size of 4KB to reduce resizing
	if size < 4096 {
		size = 4096
	}

	// If we have a buffer and it's big enough, reuse it
	if rt.prefetchShm != nil && rt.prefetchShm.Size() >= size {
		return rt.prefetchShm, nil
	}

	// Need to create a new buffer (first use or resize)
	if rt.prefetchShm != nil {
		// Destroy old buffer
		rt.shmManager.Destroy(rt.prefetchShm.Key())
	}

	// Create new buffer with some extra room for growth
	bufferSize := size + 1024
	shm, err := rt.shmManager.Create(bufferSize)
	if err != nil {
		return nil, err
	}

	rt.prefetchShm = shm
	return shm, nil
}

// GetCachedResult retrieves a cached function result by key.
// Returns the cached value and true if found, or nil and false if not cached.
func (rt *NodeJSRuntime) GetCachedResult(key string) (any, bool) {
	rt.funcResultCacheMu.RLock()
	result, ok := rt.funcResultCache[key]
	rt.funcResultCacheMu.RUnlock()
	return result, ok
}

// SetCachedResult stores a function result in the cache.
func (rt *NodeJSRuntime) SetCachedResult(key string, value any) {
	rt.funcResultCacheMu.Lock()
	rt.funcResultCache[key] = value
	rt.funcResultCacheMu.Unlock()
}

// ClearFunctionCache clears all cached function results.
// Call this between compilations to ensure fresh results.
func (rt *NodeJSRuntime) ClearFunctionCache() {
	rt.funcResultCacheMu.Lock()
	rt.funcResultCache = make(map[string]any)
	rt.funcResultCacheMu.Unlock()
}

// ClearCachedResultsForFunction clears cached results for a specific function.
// The cache key format is "funcName:arg1|arg2|...", so this deletes all entries
// that start with the function name followed by a colon.
func (rt *NodeJSRuntime) ClearCachedResultsForFunction(funcName string) {
	prefix := funcName + ":"
	rt.funcResultCacheMu.Lock()
	defer rt.funcResultCacheMu.Unlock()
	for key := range rt.funcResultCache {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			delete(rt.funcResultCache, key)
		}
	}
}

// FunctionCacheSize returns the number of cached function results.
func (rt *NodeJSRuntime) FunctionCacheSize() int {
	rt.funcResultCacheMu.RLock()
	defer rt.funcResultCacheMu.RUnlock()
	return len(rt.funcResultCache)
}

// BatchCall represents a single function call in a batch.
// This is used by BatchCallFunctions to reduce IPC overhead.
type BatchCall struct {
	// Key is a unique identifier for this call (e.g., "funcName:arg1|arg2")
	// Used to correlate requests with responses in the batch result map.
	Key string `json:"key"`
	// Name is the function name to call
	Name string `json:"name"`
	// Args are the serialized arguments for the function
	Args []any `json:"args"`
	// Context is the optional evaluation context for variable lookups
	Context map[string]any `json:"context,omitempty"`
}

// BatchCallResult represents the result of a single call in a batch.
type BatchCallResult struct {
	// Success indicates if the call succeeded
	Success bool `json:"success"`
	// Result is the function return value (if successful)
	Result any `json:"result,omitempty"`
	// Error is the error message (if failed)
	Error string `json:"error,omitempty"`
}

// BatchCallFunctions sends multiple function calls to Node.js in a single IPC request.
// This reduces the overhead of multiple round-trips for plugin function calls.
//
// The function returns a map of results keyed by the BatchCall.Key field.
// Each result contains Success, Result, and Error fields.
//
// Example usage:
//
//	calls := []BatchCall{
//	    {Key: "map-get:colors|primary", Name: "map-get", Args: [...]},
//	    {Key: "color-yiq:#fff", Name: "color-yiq", Args: [...]},
//	}
//	results, err := rt.BatchCallFunctions(calls)
//	// results["map-get:colors|primary"].Result contains the result
func (rt *NodeJSRuntime) BatchCallFunctions(calls []BatchCall) (map[string]BatchCallResult, error) {
	if !rt.alive.Load() {
		return nil, fmt.Errorf("runtime not alive")
	}

	if len(calls) == 0 {
		return make(map[string]BatchCallResult), nil
	}

	// Send batch request to Node.js
	resp, err := rt.SendCommand(Command{
		Cmd: "batchCallFunctions",
		Data: map[string]any{
			"calls": calls,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("batch call failed: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("batch call error: %s", resp.Error)
	}

	// Parse the results map
	resultsMap, ok := resp.Result.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unexpected batch result type: %T", resp.Result)
	}

	// Convert to typed results
	results := make(map[string]BatchCallResult, len(resultsMap))
	for key, val := range resultsMap {
		if resultMap, ok := val.(map[string]any); ok {
			result := BatchCallResult{}
			if success, ok := resultMap["success"].(bool); ok {
				result.Success = success
			}
			if resultVal, ok := resultMap["result"]; ok {
				result.Result = resultVal
			}
			if errMsg, ok := resultMap["error"].(string); ok {
				result.Error = errMsg
			}
			results[key] = result
		}
	}

	return results, nil
}

// BatchCallFunctionsAndCache sends multiple function calls to Node.js in a single IPC request
// and caches all successful results. This is the most efficient way to warm up the function
// result cache for plugin functions.
//
// Returns the number of successfully cached results and any error.
func (rt *NodeJSRuntime) BatchCallFunctionsAndCache(calls []BatchCall) (int, error) {
	results, err := rt.BatchCallFunctions(calls)
	if err != nil {
		return 0, err
	}

	cached := 0
	rt.funcResultCacheMu.Lock()
	for key, result := range results {
		if result.Success && result.Result != nil {
			rt.funcResultCache[key] = result.Result
			cached++
		}
	}
	rt.funcResultCacheMu.Unlock()

	return cached, nil
}

// WriteASTBuffer writes a FlatAST to shared memory and returns the segment.
// This enables zero-copy transfer of AST data to Node.js.
func (rt *NodeJSRuntime) WriteASTBuffer(flat *FlatAST) (*SharedMemory, error) {
	if rt.shmManager == nil {
		return nil, fmt.Errorf("shared memory manager not initialized")
	}

	// Serialize the AST to bytes
	data, err := flat.ToBytes()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize AST: %w", err)
	}

	// Create shared memory segment
	shm, err := rt.shmManager.Create(len(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create shared memory: %w", err)
	}

	// Write data to shared memory
	if err := shm.WriteAll(data); err != nil {
		rt.shmManager.Destroy(shm.Key())
		return nil, fmt.Errorf("failed to write to shared memory: %w", err)
	}

	// Sync to ensure data is visible to other processes
	if err := shm.Sync(); err != nil {
		rt.shmManager.Destroy(shm.Key())
		return nil, fmt.Errorf("failed to sync shared memory: %w", err)
	}

	return shm, nil
}

// ReadASTBuffer reads a FlatAST from a shared memory segment.
// This reads the AST data that was written by Node.js.
func (rt *NodeJSRuntime) ReadASTBuffer(shm *SharedMemory) (*FlatAST, error) {
	// Read all data from shared memory
	data, err := shm.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read from shared memory: %w", err)
	}

	// Deserialize the AST
	flat, err := FromBytes(data)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize AST: %w", err)
	}

	return flat, nil
}

// AttachBuffer sends a command to Node.js to attach to a shared memory buffer.
// Returns the path to the shared memory file for Node.js to map.
func (rt *NodeJSRuntime) AttachBuffer(shm *SharedMemory) error {
	resp, err := rt.SendCommand(Command{
		Cmd: "attachBuffer",
		Data: map[string]any{
			"key":  shm.Key(),
			"path": shm.Path(),
			"size": shm.Size(),
		},
	})
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("attachBuffer failed: %s", resp.Error)
	}
	return nil
}

// DetachBuffer sends a command to Node.js to detach from a shared memory buffer.
func (rt *NodeJSRuntime) DetachBuffer(key string) error {
	resp, err := rt.SendCommand(Command{
		Cmd: "detachBuffer",
		Data: map[string]any{
			"key": key,
		},
	})
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("detachBuffer failed: %s", resp.Error)
	}
	return nil
}

// ============================================================================
// High-Performance Shared Memory Protocol
// ============================================================================

// InitSHMProtocol initializes the high-performance shared memory protocol.
// This creates a persistent 4MB shared memory region for binary IPC.
// Call this once at the start of compilation for maximum performance.
func (rt *NodeJSRuntime) InitSHMProtocol() error {
	rt.shmProtocolMu.Lock()
	defer rt.shmProtocolMu.Unlock()

	if rt.shmProtocol != nil {
		return nil // Already initialized
	}

	if rt.shmManager == nil {
		return fmt.Errorf("shared memory manager not initialized")
	}

	protocol, err := NewSharedMemoryProtocol(rt.shmManager)
	if err != nil {
		return fmt.Errorf("failed to create SHM protocol: %w", err)
	}

	rt.shmProtocol = protocol

	// Notify JavaScript about the protocol
	resp, err := rt.SendCommand(Command{
		Cmd: "initSHMProtocol",
		Data: map[string]any{
			"path":           protocol.Path(),
			"key":            protocol.Key(),
			"totalSize":      TotalSHMSize,
			"controlLayout":  protocol.GetControlBlockLayout(),
			"sectionOffsets": protocol.GetSectionOffsets(),
		},
	})
	if err != nil {
		protocol.Close()
		rt.shmProtocol = nil
		return fmt.Errorf("failed to initialize SHM protocol on JS side: %w", err)
	}
	if !resp.Success {
		protocol.Close()
		rt.shmProtocol = nil
		return fmt.Errorf("JS failed to init SHM protocol: %s", resp.Error)
	}

	rt.useSHMProtocol = true
	return nil
}

// GetSHMProtocol returns the shared memory protocol if initialized.
func (rt *NodeJSRuntime) GetSHMProtocol() *SharedMemoryProtocol {
	rt.shmProtocolMu.Lock()
	defer rt.shmProtocolMu.Unlock()
	return rt.shmProtocol
}

// UseSHMProtocol returns whether the binary SHM protocol is enabled.
func (rt *NodeJSRuntime) UseSHMProtocol() bool {
	return rt.useSHMProtocol
}

// PreloadVariables writes all variables from the evaluation context to shared memory.
// This should be called once at the start of compilation for best performance.
func (rt *NodeJSRuntime) PreloadVariables(frames []any) error {
	rt.shmProtocolMu.Lock()
	protocol := rt.shmProtocol
	rt.shmProtocolMu.Unlock()

	if protocol == nil {
		return fmt.Errorf("SHM protocol not initialized")
	}

	// Collect all variables from all frames
	allVars := make(map[string]any)
	for _, frame := range frames {
		if frame == nil {
			continue
		}
		variablesProvider, ok := frame.(interface{ Variables() map[string]any })
		if !ok {
			continue
		}
		variables := variablesProvider.Variables()
		if variables == nil {
			continue
		}
		for name, decl := range variables {
			if _, exists := allVars[name]; !exists {
				allVars[name] = decl
			}
		}
	}

	// Write to shared memory
	if err := protocol.WriteVariables(allVars); err != nil {
		return fmt.Errorf("failed to write variables: %w", err)
	}

	if os.Getenv("LESS_GO_DEBUG") == "1" {
		fmt.Printf("[PreloadVariables] Preloaded %d variables to shared memory\n", len(allVars))
	}

	return nil
}

// CallFunctionViaSHM calls a JavaScript function using the binary shared memory protocol.
// This is much faster than JSON-based IPC for repeated function calls.
func (rt *NodeJSRuntime) CallFunctionViaSHM(functionName string, args ...any) (any, error) {
	rt.shmProtocolMu.Lock()
	protocol := rt.shmProtocol
	rt.shmProtocolMu.Unlock()

	if protocol == nil {
		return nil, fmt.Errorf("SHM protocol not initialized")
	}

	// Get or register the function ID
	funcID, ok := protocol.GetFunctionID(functionName)
	if !ok {
		funcID = protocol.RegisterFunction(functionName)
		// Notify JS about the function mapping
		rt.SendCommand(Command{
			Cmd: "registerSHMFunction",
			Data: map[string]any{
				"name": functionName,
				"id":   funcID,
			},
		})
	}

	// Prepare the call
	if err := protocol.PrepareCall(funcID, len(args)); err != nil {
		return nil, fmt.Errorf("failed to prepare call: %w", err)
	}

	// Write arguments
	for i, arg := range args {
		if _, err := protocol.WriteArg(i, arg); err != nil {
			return nil, fmt.Errorf("failed to write arg %d: %w", i, err)
		}
	}

	// Signal the request
	if err := protocol.SignalRequest(); err != nil {
		return nil, fmt.Errorf("failed to signal request: %w", err)
	}

	// Wait for response (timeout: 30 seconds)
	ready, err := protocol.WaitForResponse(30000)
	if err != nil || !ready {
		return nil, fmt.Errorf("timeout or error waiting for response: %v", err)
	}

	// Read the result
	result, err := protocol.ReadResult()
	if err != nil {
		return nil, err
	}

	// Clear for next call
	protocol.ClearResponse()

	return result, nil
}

// CloseSHMProtocol closes the shared memory protocol and releases resources.
func (rt *NodeJSRuntime) CloseSHMProtocol() error {
	rt.shmProtocolMu.Lock()
	defer rt.shmProtocolMu.Unlock()

	if rt.shmProtocol == nil {
		return nil
	}

	err := rt.shmProtocol.Close()
	rt.shmProtocol = nil
	rt.useSHMProtocol = false
	return err
}
