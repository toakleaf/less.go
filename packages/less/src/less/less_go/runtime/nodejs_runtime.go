package runtime

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
)

// NodeJSRuntime manages a Node.js process for executing JavaScript plugins.
type NodeJSRuntime struct {
	process   *exec.Cmd
	stdin     io.WriteCloser
	stdout    io.ReadCloser
	stderr    io.ReadCloser
	alive     bool
	mu        sync.Mutex
	responses chan *Response
	errors    chan error
	nextID    int
	pending   map[int]chan *Response
	stopChan  chan struct{} // Signal to stop goroutines
	wg        sync.WaitGroup // Wait for goroutines to finish
}

// Command represents a command sent to the Node.js process
type Command struct {
	ID      int                    `json:"id"`
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

// Response represents a response from the Node.js process
type Response struct {
	ID      int                    `json:"id"`
	Success bool                   `json:"success"`
	Result  map[string]interface{} `json:"result,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

// NewNodeJSRuntime creates a new Node.js runtime instance.
// The Node.js process is not started until Start() is called.
func NewNodeJSRuntime() (*NodeJSRuntime, error) {
	return &NodeJSRuntime{
		alive:     false,
		responses: make(chan *Response, 10),
		errors:    make(chan error, 10),
		pending:   make(map[int]chan *Response),
		stopChan:  make(chan struct{}),
	}, nil
}

// Start spawns the Node.js process and establishes communication.
func (rt *NodeJSRuntime) Start() error {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	if rt.alive {
		return fmt.Errorf("runtime already started")
	}

	// Find the plugin-host.js file path relative to this source file
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("failed to get current file path")
	}
	pluginHostPath := filepath.Join(filepath.Dir(filename), "plugin-host.js")

	// Create the Node.js command
	rt.process = exec.Command("node", pluginHostPath)

	// Set up stdin
	stdin, err := rt.process.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	rt.stdin = stdin

	// Set up stdout
	stdout, err := rt.process.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	rt.stdout = stdout

	// Set up stderr
	stderr, err := rt.process.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	rt.stderr = stderr

	// Start the process
	if err := rt.process.Start(); err != nil {
		return fmt.Errorf("failed to start Node.js process: %w", err)
	}

	rt.alive = true

	// Start goroutines to handle I/O
	rt.wg.Add(2)
	go rt.readResponses()
	go rt.readErrors()

	return nil
}

// Stop terminates the Node.js process and cleans up resources.
func (rt *NodeJSRuntime) Stop() error {
	rt.mu.Lock()

	if !rt.alive {
		rt.mu.Unlock()
		return nil
	}

	rt.alive = false

	// Close stdin to signal Node.js process to exit gracefully
	if rt.stdin != nil {
		rt.stdin.Close()
	}

	// Signal goroutines to stop
	close(rt.stopChan)

	rt.mu.Unlock()

	// Wait for process to exit
	if rt.process != nil {
		rt.process.Wait()
	}

	// Wait for goroutines to finish
	rt.wg.Wait()

	// Now safe to close channels
	close(rt.responses)
	close(rt.errors)

	return nil
}

// IsAlive returns whether the Node.js process is running.
func (rt *NodeJSRuntime) IsAlive() bool {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	return rt.alive
}

// SendCommand sends a command to the Node.js process and waits for a response.
func (rt *NodeJSRuntime) SendCommand(cmdType string, payload map[string]interface{}) (*Response, error) {
	rt.mu.Lock()

	if !rt.alive {
		rt.mu.Unlock()
		return nil, fmt.Errorf("runtime not started")
	}

	// Generate command ID
	rt.nextID++
	cmdID := rt.nextID

	// Create response channel for this command
	respChan := make(chan *Response, 1)
	rt.pending[cmdID] = respChan

	// Create command
	cmd := Command{
		ID:      cmdID,
		Type:    cmdType,
		Payload: payload,
	}

	rt.mu.Unlock()

	// Send command
	data, err := json.Marshal(cmd)
	if err != nil {
		rt.mu.Lock()
		delete(rt.pending, cmdID)
		rt.mu.Unlock()
		return nil, fmt.Errorf("failed to marshal command: %w", err)
	}

	// Write to stdin (one command per line)
	if _, err := rt.stdin.Write(append(data, '\n')); err != nil {
		rt.mu.Lock()
		delete(rt.pending, cmdID)
		rt.mu.Unlock()
		return nil, fmt.Errorf("failed to write command: %w", err)
	}

	// Wait for response
	resp := <-respChan

	// Clean up
	rt.mu.Lock()
	delete(rt.pending, cmdID)
	rt.mu.Unlock()

	if !resp.Success {
		return nil, fmt.Errorf("command failed: %s", resp.Error)
	}

	return resp, nil
}

// readResponses reads responses from the Node.js process stdout.
func (rt *NodeJSRuntime) readResponses() {
	defer rt.wg.Done()

	scanner := bufio.NewScanner(rt.stdout)
	for scanner.Scan() {
		line := scanner.Text()

		var resp Response
		if err := json.Unmarshal([]byte(line), &resp); err != nil {
			select {
			case rt.errors <- fmt.Errorf("failed to parse response: %w", err):
			case <-rt.stopChan:
				return
			}
			continue
		}

		// Route response to waiting command
		rt.mu.Lock()
		if respChan, ok := rt.pending[resp.ID]; ok {
			respChan <- &resp
		}
		rt.mu.Unlock()
	}

	if err := scanner.Err(); err != nil {
		select {
		case rt.errors <- fmt.Errorf("error reading stdout: %w", err):
		case <-rt.stopChan:
		}
	}
}

// readErrors reads errors from the Node.js process stderr.
func (rt *NodeJSRuntime) readErrors() {
	defer rt.wg.Done()

	scanner := bufio.NewScanner(rt.stderr)
	for scanner.Scan() {
		line := scanner.Text()
		select {
		case rt.errors <- fmt.Errorf("Node.js stderr: %s", line):
		case <-rt.stopChan:
			return
		}
	}
}

// Ping sends a ping command to verify the Node.js process is responding.
func (rt *NodeJSRuntime) Ping() error {
	_, err := rt.SendCommand("ping", nil)
	return err
}
