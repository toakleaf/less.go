package runtime

import (
	"testing"
	"time"
)

// TestNodeJSRuntime_StartStop tests the basic lifecycle of the Node.js runtime
func TestNodeJSRuntime_StartStop(t *testing.T) {
	rt, err := NewNodeJSRuntime()
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}

	// Start the runtime
	err = rt.Start()
	if err != nil {
		t.Fatalf("Failed to start runtime: %v", err)
	}

	// Check that it's alive
	if !rt.IsAlive() {
		t.Fatal("Runtime should be alive after Start()")
	}

	// Stop the runtime
	err = rt.Stop()
	if err != nil {
		t.Fatalf("Failed to stop runtime: %v", err)
	}

	// Check that it's not alive
	if rt.IsAlive() {
		t.Fatal("Runtime should not be alive after Stop()")
	}
}

// TestNodeJSRuntime_Ping tests the ping command
func TestNodeJSRuntime_Ping(t *testing.T) {
	rt, err := NewNodeJSRuntime()
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}

	if err := rt.Start(); err != nil {
		t.Fatalf("Failed to start runtime: %v", err)
	}
	defer rt.Stop()

	// Give Node.js a moment to start up
	time.Sleep(100 * time.Millisecond)

	// Send ping command
	err = rt.Ping()
	if err != nil {
		t.Fatalf("Ping failed: %v", err)
	}
}

// TestNodeJSRuntime_Echo tests the echo command
func TestNodeJSRuntime_Echo(t *testing.T) {
	rt, err := NewNodeJSRuntime()
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}

	if err := rt.Start(); err != nil {
		t.Fatalf("Failed to start runtime: %v", err)
	}
	defer rt.Stop()

	// Give Node.js a moment to start up
	time.Sleep(100 * time.Millisecond)

	// Send echo command
	testMessage := "Hello from Go!"
	resp, err := rt.SendCommand("echo", map[string]interface{}{
		"message": testMessage,
	})

	if err != nil {
		t.Fatalf("Echo command failed: %v", err)
	}

	if resp.Result["echo"] != testMessage {
		t.Errorf("Expected echo '%s', got '%v'", testMessage, resp.Result["echo"])
	}
}

// TestNodeJSRuntime_LoadPlugin tests loading a JavaScript plugin
func TestNodeJSRuntime_LoadPlugin(t *testing.T) {
	rt, err := NewNodeJSRuntime()
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}

	if err := rt.Start(); err != nil {
		t.Fatalf("Failed to start runtime: %v", err)
	}
	defer rt.Stop()

	// Give Node.js a moment to start up
	time.Sleep(100 * time.Millisecond)

	// Path to the simple plugin - use absolute path
	pluginPath := "/home/user/less.go/packages/test-data/plugin/plugin-simple.js"

	resp, err := rt.SendCommand("load_plugin", map[string]interface{}{
		"pluginPath": pluginPath,
	})

	if err != nil {
		t.Fatalf("Load plugin failed: %v", err)
	}

	// Check that plugin was loaded
	pluginID, ok := resp.Result["pluginID"].(string)
	if !ok || pluginID == "" {
		t.Fatal("Expected pluginID in response")
	}

	// Check that functions were registered
	functions, ok := resp.Result["functions"].([]interface{})
	if !ok {
		t.Fatal("Expected functions array in response")
	}

	if len(functions) < 1 {
		t.Fatal("Expected at least one function to be registered")
	}

	t.Logf("Loaded plugin %s with %d functions: %v", pluginID, len(functions), functions)
}

// TestNodeJSRuntime_MultipleCommands tests sending multiple commands in sequence
func TestNodeJSRuntime_MultipleCommands(t *testing.T) {
	rt, err := NewNodeJSRuntime()
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}

	if err := rt.Start(); err != nil {
		t.Fatalf("Failed to start runtime: %v", err)
	}
	defer rt.Stop()

	// Give Node.js a moment to start up
	time.Sleep(100 * time.Millisecond)

	// Send multiple commands
	for i := 0; i < 5; i++ {
		err := rt.Ping()
		if err != nil {
			t.Fatalf("Ping %d failed: %v", i, err)
		}
	}
}

// TestNodeJSRuntime_CommandBeforeStart tests that commands fail before Start()
func TestNodeJSRuntime_CommandBeforeStart(t *testing.T) {
	rt, err := NewNodeJSRuntime()
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}

	// Try to send command before starting
	err = rt.Ping()
	if err == nil {
		t.Fatal("Expected error when sending command before Start()")
	}
}

// TestNodeJSRuntime_DoubleStart tests that starting twice fails
func TestNodeJSRuntime_DoubleStart(t *testing.T) {
	rt, err := NewNodeJSRuntime()
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}

	if err := rt.Start(); err != nil {
		t.Fatalf("Failed to start runtime: %v", err)
	}
	defer rt.Stop()

	// Try to start again
	err = rt.Start()
	if err == nil {
		t.Fatal("Expected error when starting twice")
	}
}

// TestNodeJSRuntime_InvalidCommand tests handling of invalid commands
func TestNodeJSRuntime_InvalidCommand(t *testing.T) {
	rt, err := NewNodeJSRuntime()
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}

	if err := rt.Start(); err != nil {
		t.Fatalf("Failed to start runtime: %v", err)
	}
	defer rt.Stop()

	// Give Node.js a moment to start up
	time.Sleep(100 * time.Millisecond)

	// Send invalid command type
	_, err = rt.SendCommand("invalid_command_type", nil)
	if err == nil {
		t.Fatal("Expected error for invalid command type")
	}
}

// TestNodeJSRuntime_EchoMissingPayload tests error handling for missing payload
func TestNodeJSRuntime_EchoMissingPayload(t *testing.T) {
	rt, err := NewNodeJSRuntime()
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}

	if err := rt.Start(); err != nil {
		t.Fatalf("Failed to start runtime: %v", err)
	}
	defer rt.Stop()

	// Give Node.js a moment to start up
	time.Sleep(100 * time.Millisecond)

	// Send echo without message
	_, err = rt.SendCommand("echo", map[string]interface{}{})
	if err == nil {
		t.Fatal("Expected error for echo without message")
	}
}
