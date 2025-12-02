package runtime

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

// TestIPCLatencyProfile profiles the IPC latency breakdown.
// Run with: go test -v -run TestIPCLatencyProfile ./less/runtime/...
func TestIPCLatencyProfile(t *testing.T) {
	rt, err := NewNodeJSRuntime()
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}
	if err := rt.Start(); err != nil {
		t.Fatalf("Failed to start runtime: %v", err)
	}
	defer rt.Stop()

	// Warmup
	for i := 0; i < 10; i++ {
		rt.Ping()
	}

	const iterations = 100

	// Test 1: Measure JSON marshalling time
	t.Run("JSONMarshal", func(t *testing.T) {
		cmd := Command{
			ID:  1,
			Cmd: "ping",
		}

		start := time.Now()
		for i := 0; i < iterations; i++ {
			cmd.ID = int64(i)
			_, err := json.Marshal(cmd)
			if err != nil {
				t.Fatal(err)
			}
		}
		elapsed := time.Since(start)
		t.Logf("JSON Marshal: %v per call (%v total for %d calls)",
			elapsed/time.Duration(iterations), elapsed, iterations)
	})

	// Test 2: Full round-trip latency
	t.Run("FullRoundTrip", func(t *testing.T) {
		start := time.Now()
		for i := 0; i < iterations; i++ {
			err := rt.Ping()
			if err != nil {
				t.Fatal(err)
			}
		}
		elapsed := time.Since(start)
		t.Logf("Full round-trip: %v per call (%v total for %d calls)",
			elapsed/time.Duration(iterations), elapsed, iterations)
	})

	// Test 3: Echo with data payload
	t.Run("EchoSmallPayload", func(t *testing.T) {
		start := time.Now()
		for i := 0; i < iterations; i++ {
			_, err := rt.Echo("hello")
			if err != nil {
				t.Fatal(err)
			}
		}
		elapsed := time.Since(start)
		t.Logf("Echo (small): %v per call (%v total for %d calls)",
			elapsed/time.Duration(iterations), elapsed, iterations)
	})

	// Test 4: Echo with larger payload
	t.Run("EchoLargePayload", func(t *testing.T) {
		largeData := make(map[string]any)
		for i := 0; i < 50; i++ {
			largeData[fmt.Sprintf("key%d", i)] = fmt.Sprintf("value%d", i)
		}

		start := time.Now()
		for i := 0; i < iterations; i++ {
			_, err := rt.Echo(largeData)
			if err != nil {
				t.Fatal(err)
			}
		}
		elapsed := time.Since(start)
		t.Logf("Echo (large ~50 keys): %v per call (%v total for %d calls)",
			elapsed/time.Duration(iterations), elapsed, iterations)
	})
}

// TestIPCLatencyDetailed provides detailed timing breakdown for a single IPC call.
func TestIPCLatencyDetailed(t *testing.T) {
	rt, err := NewNodeJSRuntime()
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}
	if err := rt.Start(); err != nil {
		t.Fatalf("Failed to start runtime: %v", err)
	}
	defer rt.Stop()

	// Warmup
	for i := 0; i < 10; i++ {
		rt.Ping()
	}

	const iterations = 100

	var (
		totalMarshal time.Duration
		totalWrite   time.Duration
		totalWait    time.Duration
	)

	for i := 0; i < iterations; i++ {
		cmd := Command{
			ID:   int64(i + 1000),
			Cmd:  "echo",
			Data: "test",
		}

		// Measure marshal
		t1 := time.Now()
		data, _ := json.Marshal(cmd)
		t2 := time.Now()
		totalMarshal += t2.Sub(t1)

		// Create response channel
		respChan := make(chan Response, 1)
		rt.responsesMu.Lock()
		rt.responses[cmd.ID] = respChan
		rt.responsesMu.Unlock()

		// Measure write
		t3 := time.Now()
		rt.stdin.Write(append(data, '\n'))
		t4 := time.Now()
		totalWrite += t4.Sub(t3)

		// Measure wait for response
		t5 := time.Now()
		select {
		case <-respChan:
		case <-time.After(5 * time.Second):
			t.Fatal("timeout")
		}
		t6 := time.Now()
		totalWait += t6.Sub(t5)

		// Cleanup
		rt.responsesMu.Lock()
		delete(rt.responses, cmd.ID)
		rt.responsesMu.Unlock()
	}

	t.Logf("=== IPC Latency Breakdown (avg over %d iterations) ===", iterations)
	t.Logf("  JSON Marshal:    %v", totalMarshal/time.Duration(iterations))
	t.Logf("  Stdin Write:     %v", totalWrite/time.Duration(iterations))
	t.Logf("  Wait (IPC+JS):   %v", totalWait/time.Duration(iterations))
	t.Logf("  TOTAL:           %v", (totalMarshal+totalWrite+totalWait)/time.Duration(iterations))
}

// BenchmarkIPCPing benchmarks the raw ping round-trip.
func BenchmarkIPCPing(b *testing.B) {
	rt, err := NewNodeJSRuntime()
	if err != nil {
		b.Fatalf("Failed to create runtime: %v", err)
	}
	if err := rt.Start(); err != nil {
		b.Fatalf("Failed to start runtime: %v", err)
	}
	defer rt.Stop()

	// Warmup
	for i := 0; i < 10; i++ {
		rt.Ping()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := rt.Ping(); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkJSONMarshalCommand benchmarks JSON marshalling overhead.
func BenchmarkJSONMarshalCommand(b *testing.B) {
	cmd := Command{
		ID:  1,
		Cmd: "callFunction",
		Data: map[string]any{
			"name": "test-function",
			"args": []any{"arg1", "arg2"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd.ID = int64(i)
		_, err := json.Marshal(cmd)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkIPCWithPayload benchmarks IPC with different payload sizes.
func BenchmarkIPCWithPayload(b *testing.B) {
	rt, err := NewNodeJSRuntime()
	if err != nil {
		b.Fatalf("Failed to create runtime: %v", err)
	}
	if err := rt.Start(); err != nil {
		b.Fatalf("Failed to start runtime: %v", err)
	}
	defer rt.Stop()

	// Warmup
	for i := 0; i < 10; i++ {
		rt.Ping()
	}

	b.Run("SmallPayload", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			rt.Echo("hello")
		}
	})

	b.Run("MediumPayload", func(b *testing.B) {
		data := make(map[string]any)
		for i := 0; i < 20; i++ {
			data[fmt.Sprintf("key%d", i)] = fmt.Sprintf("value%d", i)
		}
		for i := 0; i < b.N; i++ {
			rt.Echo(data)
		}
	})

	b.Run("LargePayload", func(b *testing.B) {
		data := make(map[string]any)
		for i := 0; i < 100; i++ {
			data[fmt.Sprintf("key%d", i)] = fmt.Sprintf("value%d with more content to make it larger", i)
		}
		for i := 0; i < b.N; i++ {
			rt.Echo(data)
		}
	})
}

// TestIPCBufferingComparison tests buffered vs unbuffered writes.
func TestIPCBufferingComparison(t *testing.T) {
	rt, err := NewNodeJSRuntime()
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}
	if err := rt.Start(); err != nil {
		t.Fatalf("Failed to start runtime: %v", err)
	}
	defer rt.Stop()

	// Warmup
	for i := 0; i < 10; i++ {
		rt.Ping()
	}

	const iterations = 100

	// Test current unbuffered approach
	t.Run("CurrentUnbuffered", func(t *testing.T) {
		start := time.Now()
		for i := 0; i < iterations; i++ {
			rt.Ping()
		}
		elapsed := time.Since(start)
		t.Logf("Unbuffered: %v per call", elapsed/time.Duration(iterations))
	})

	// Note: To test buffered writes, we'd need to modify the runtime
	// For now, we just measure the baseline
}
