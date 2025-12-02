package runtime

import (
	"bytes"
	"os"
	"testing"
)

func TestSharedMemoryManager_Create(t *testing.T) {
	manager, err := NewSharedMemoryManager()
	if err != nil {
		t.Fatalf("NewSharedMemoryManager failed: %v", err)
	}
	defer manager.DestroyAll()

	// Create a segment
	shm, err := manager.Create(1024)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if shm == nil {
		t.Fatal("shared memory is nil")
	}

	if shm.Key() == "" {
		t.Error("key should not be empty")
	}

	if shm.Size() != 1024 {
		t.Errorf("size = %d, want 1024", shm.Size())
	}

	// Verify file exists
	if _, err := os.Stat(shm.Path()); err != nil {
		t.Errorf("backing file should exist: %v", err)
	}
}

func TestSharedMemoryManager_CreateMultiple(t *testing.T) {
	manager, err := NewSharedMemoryManager()
	if err != nil {
		t.Fatalf("NewSharedMemoryManager failed: %v", err)
	}
	defer manager.DestroyAll()

	// Create multiple segments
	segments := make([]*SharedMemory, 5)
	for i := 0; i < 5; i++ {
		shm, err := manager.Create(1024 * (i + 1))
		if err != nil {
			t.Fatalf("Create %d failed: %v", i, err)
		}
		segments[i] = shm
	}

	// Verify all have unique keys
	keys := make(map[string]bool)
	for _, shm := range segments {
		if keys[shm.Key()] {
			t.Errorf("duplicate key: %s", shm.Key())
		}
		keys[shm.Key()] = true
	}
}

func TestSharedMemory_WriteRead(t *testing.T) {
	manager, err := NewSharedMemoryManager()
	if err != nil {
		t.Fatalf("NewSharedMemoryManager failed: %v", err)
	}
	defer manager.DestroyAll()

	shm, err := manager.Create(1024)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Write data
	testData := []byte("Hello, shared memory!")
	if err := shm.Write(0, testData); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Read data back
	readData, err := shm.Read(0, len(testData))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if !bytes.Equal(readData, testData) {
		t.Errorf("data mismatch: got %q, want %q", readData, testData)
	}
}

func TestSharedMemory_WriteAll_ReadAll(t *testing.T) {
	manager, err := NewSharedMemoryManager()
	if err != nil {
		t.Fatalf("NewSharedMemoryManager failed: %v", err)
	}
	defer manager.DestroyAll()

	testData := []byte("This is test data for write all and read all operations.")
	shm, err := manager.Create(len(testData))
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Write all data
	if err := shm.WriteAll(testData); err != nil {
		t.Fatalf("WriteAll failed: %v", err)
	}

	// Sync to ensure visibility
	if err := shm.Sync(); err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	// Read all data back
	readData, err := shm.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	if !bytes.Equal(readData, testData) {
		t.Errorf("data mismatch: got %q, want %q", readData, testData)
	}
}

func TestSharedMemory_WriteOutOfBounds(t *testing.T) {
	manager, err := NewSharedMemoryManager()
	if err != nil {
		t.Fatalf("NewSharedMemoryManager failed: %v", err)
	}
	defer manager.DestroyAll()

	shm, err := manager.Create(100)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Try to write beyond bounds
	largeData := make([]byte, 150)
	err = shm.Write(0, largeData)
	if err == nil {
		t.Error("Write should fail for out-of-bounds data")
	}

	// Try to write at invalid offset
	smallData := []byte("test")
	err = shm.Write(99, smallData)
	if err == nil {
		t.Error("Write should fail for out-of-bounds offset")
	}
}

func TestSharedMemory_ReadOutOfBounds(t *testing.T) {
	manager, err := NewSharedMemoryManager()
	if err != nil {
		t.Fatalf("NewSharedMemoryManager failed: %v", err)
	}
	defer manager.DestroyAll()

	shm, err := manager.Create(100)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Try to read beyond bounds
	_, err = shm.Read(0, 150)
	if err == nil {
		t.Error("Read should fail for out-of-bounds length")
	}

	// Try to read at invalid offset
	_, err = shm.Read(99, 10)
	if err == nil {
		t.Error("Read should fail for out-of-bounds offset")
	}
}

func TestSharedMemoryManager_Destroy(t *testing.T) {
	manager, err := NewSharedMemoryManager()
	if err != nil {
		t.Fatalf("NewSharedMemoryManager failed: %v", err)
	}
	defer manager.DestroyAll()

	shm, err := manager.Create(1024)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	path := shm.Path()
	key := shm.Key()

	// Destroy the segment
	if err := manager.Destroy(key); err != nil {
		t.Fatalf("Destroy failed: %v", err)
	}

	// Verify file is removed
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("backing file should be removed after destroy")
	}

	// Verify Get returns nil
	if manager.Get(key) != nil {
		t.Error("Get should return nil after destroy")
	}

	// Destroying again should fail
	if err := manager.Destroy(key); err == nil {
		t.Error("Destroy should fail for non-existent segment")
	}
}

func TestSharedMemoryManager_DestroyAll(t *testing.T) {
	manager, err := NewSharedMemoryManager()
	if err != nil {
		t.Fatalf("NewSharedMemoryManager failed: %v", err)
	}

	// Create multiple segments
	paths := make([]string, 3)
	for i := 0; i < 3; i++ {
		shm, err := manager.Create(1024)
		if err != nil {
			t.Fatalf("Create %d failed: %v", i, err)
		}
		paths[i] = shm.Path()
	}

	tempDir := manager.tempDir

	// Destroy all
	if err := manager.DestroyAll(); err != nil {
		t.Fatalf("DestroyAll failed: %v", err)
	}

	// Verify all files are removed
	for i, path := range paths {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("backing file %d should be removed after DestroyAll", i)
		}
	}

	// Verify temp directory is removed
	if _, err := os.Stat(tempDir); !os.IsNotExist(err) {
		t.Error("temp directory should be removed after DestroyAll")
	}
}

func TestSharedMemory_Close(t *testing.T) {
	manager, err := NewSharedMemoryManager()
	if err != nil {
		t.Fatalf("NewSharedMemoryManager failed: %v", err)
	}
	defer manager.DestroyAll()

	shm, err := manager.Create(1024)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	path := shm.Path()

	// Write some data
	if err := shm.WriteAll([]byte("test data")); err != nil {
		t.Fatalf("WriteAll failed: %v", err)
	}

	// Close the segment directly
	if err := shm.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Verify file is removed
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("backing file should be removed after Close")
	}

	// Operations on closed segment should work (they just return empty/error)
	// since we nil out the data
	if shm.Data() != nil {
		t.Error("Data() should return nil after Close")
	}
}

func TestSharedMemory_CreateInvalidSize(t *testing.T) {
	manager, err := NewSharedMemoryManager()
	if err != nil {
		t.Fatalf("NewSharedMemoryManager failed: %v", err)
	}
	defer manager.DestroyAll()

	// Try to create with size 0
	_, err = manager.Create(0)
	if err == nil {
		t.Error("Create should fail with size 0")
	}

	// Try to create with negative size
	_, err = manager.Create(-1)
	if err == nil {
		t.Error("Create should fail with negative size")
	}
}

func TestSharedMemory_BinaryData(t *testing.T) {
	manager, err := NewSharedMemoryManager()
	if err != nil {
		t.Fatalf("NewSharedMemoryManager failed: %v", err)
	}
	defer manager.DestroyAll()

	// Create binary data with all byte values
	binaryData := make([]byte, 256)
	for i := 0; i < 256; i++ {
		binaryData[i] = byte(i)
	}

	shm, err := manager.Create(len(binaryData))
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Write binary data
	if err := shm.WriteAll(binaryData); err != nil {
		t.Fatalf("WriteAll failed: %v", err)
	}

	// Read back
	readData, err := shm.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	if !bytes.Equal(readData, binaryData) {
		t.Error("binary data mismatch")
	}
}

// Benchmarks

func BenchmarkSharedMemory_Create(b *testing.B) {
	manager, err := NewSharedMemoryManager()
	if err != nil {
		b.Fatalf("NewSharedMemoryManager failed: %v", err)
	}
	defer manager.DestroyAll()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		shm, err := manager.Create(4096)
		if err != nil {
			b.Fatalf("Create failed: %v", err)
		}
		manager.Destroy(shm.Key())
	}
}

func BenchmarkSharedMemory_WriteRead(b *testing.B) {
	manager, err := NewSharedMemoryManager()
	if err != nil {
		b.Fatalf("NewSharedMemoryManager failed: %v", err)
	}
	defer manager.DestroyAll()

	data := make([]byte, 64*1024) // 64KB
	for i := range data {
		data[i] = byte(i % 256)
	}

	shm, err := manager.Create(len(data))
	if err != nil {
		b.Fatalf("Create failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := shm.WriteAll(data); err != nil {
			b.Fatalf("WriteAll failed: %v", err)
		}
		if _, err := shm.ReadAll(); err != nil {
			b.Fatalf("ReadAll failed: %v", err)
		}
	}
}

func BenchmarkSharedMemory_LargeBuffer(b *testing.B) {
	manager, err := NewSharedMemoryManager()
	if err != nil {
		b.Fatalf("NewSharedMemoryManager failed: %v", err)
	}
	defer manager.DestroyAll()

	data := make([]byte, 1024*1024) // 1MB
	for i := range data {
		data[i] = byte(i % 256)
	}

	shm, err := manager.Create(len(data))
	if err != nil {
		b.Fatalf("Create failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := shm.WriteAll(data); err != nil {
			b.Fatalf("WriteAll failed: %v", err)
		}
		if _, err := shm.ReadAll(); err != nil {
			b.Fatalf("ReadAll failed: %v", err)
		}
	}
}

// Tests for AST buffer roundtrip via shared memory

func TestSharedMemory_ASTRoundtrip(t *testing.T) {
	manager, err := NewSharedMemoryManager()
	if err != nil {
		t.Fatalf("NewSharedMemoryManager failed: %v", err)
	}
	defer manager.DestroyAll()

	// Create a test AST
	flat := NewFlatAST()
	flat.AddString("hello")
	flat.AddString("world")
	flat.TypeTable = append(flat.TypeTable, "TestType")

	props := map[string]any{"key": "value", "num": 42.5}
	offset, length := flat.AddProperties(props)

	flat.AddNode(FlatNode{
		TypeID:      TypeDimension,
		Flags:       FlagParens,
		PropsOffset: offset,
		PropsLength: length,
	})
	flat.AddNode(FlatNode{
		TypeID:      TypeQuoted,
		ParentIndex: 0,
	})
	flat.RootIndex = 0

	// Serialize to bytes
	data, err := flat.ToBytes()
	if err != nil {
		t.Fatalf("ToBytes failed: %v", err)
	}

	// Write to shared memory
	shm, err := manager.Create(len(data))
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := shm.WriteAll(data); err != nil {
		t.Fatalf("WriteAll failed: %v", err)
	}

	if err := shm.Sync(); err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	// Read back from shared memory
	readData, err := shm.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	// Deserialize
	restored, err := FromBytes(readData)
	if err != nil {
		t.Fatalf("FromBytes failed: %v", err)
	}

	// Verify
	if restored.Version != flat.Version {
		t.Errorf("Version = %d, want %d", restored.Version, flat.Version)
	}
	if restored.NodeCount != flat.NodeCount {
		t.Errorf("NodeCount = %d, want %d", restored.NodeCount, flat.NodeCount)
	}
	if restored.RootIndex != flat.RootIndex {
		t.Errorf("RootIndex = %d, want %d", restored.RootIndex, flat.RootIndex)
	}
	if len(restored.StringTable) != len(flat.StringTable) {
		t.Errorf("StringTable length = %d, want %d", len(restored.StringTable), len(flat.StringTable))
	}
	if len(restored.Nodes) != len(flat.Nodes) {
		t.Errorf("Nodes length = %d, want %d", len(restored.Nodes), len(flat.Nodes))
	}
}

func TestNodeJSRuntime_WriteASTBuffer(t *testing.T) {
	path := getPluginHostPath(t)
	rt, err := NewNodeJSRuntime(WithPluginHostPath(path))
	if err != nil {
		t.Fatalf("NewNodeJSRuntime failed: %v", err)
	}

	if err := rt.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer rt.Stop()

	// Create a test AST
	flat := NewFlatAST()
	flat.AddString("test-value")
	flat.AddNode(FlatNode{
		TypeID: TypeDimension,
		Flags:  FlagParens,
	})
	flat.RootIndex = 0

	// Write to shared memory via runtime
	shm, err := rt.WriteASTBuffer(flat)
	if err != nil {
		t.Fatalf("WriteASTBuffer failed: %v", err)
	}

	if shm == nil {
		t.Fatal("shared memory is nil")
	}

	if shm.Key() == "" {
		t.Error("key should not be empty")
	}

	// Read back and verify
	restored, err := rt.ReadASTBuffer(shm)
	if err != nil {
		t.Fatalf("ReadASTBuffer failed: %v", err)
	}

	if restored.NodeCount != flat.NodeCount {
		t.Errorf("NodeCount = %d, want %d", restored.NodeCount, flat.NodeCount)
	}
}

func TestNodeJSRuntime_AttachBuffer(t *testing.T) {
	path := getPluginHostPath(t)
	rt, err := NewNodeJSRuntime(WithPluginHostPath(path))
	if err != nil {
		t.Fatalf("NewNodeJSRuntime failed: %v", err)
	}

	if err := rt.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer rt.Stop()

	// Create a test AST
	flat := NewFlatAST()
	flat.AddString("attached-test")
	flat.AddNode(FlatNode{
		TypeID: TypeQuoted,
	})
	flat.RootIndex = 0

	// Write to shared memory
	shm, err := rt.WriteASTBuffer(flat)
	if err != nil {
		t.Fatalf("WriteASTBuffer failed: %v", err)
	}

	// Attach buffer in Node.js
	if err := rt.AttachBuffer(shm); err != nil {
		t.Fatalf("AttachBuffer failed: %v", err)
	}

	// Get buffer info from Node.js
	resp, err := rt.SendCommand(Command{
		Cmd: "getBufferInfo",
		Data: map[string]any{
			"key": shm.Key(),
		},
	})
	if err != nil {
		t.Fatalf("getBufferInfo failed: %v", err)
	}

	if !resp.Success {
		t.Fatalf("getBufferInfo not successful: %s", resp.Error)
	}

	result := resp.Result.(map[string]any)
	if result["key"] != shm.Key() {
		t.Errorf("key = %v, want %s", result["key"], shm.Key())
	}

	// Detach buffer
	if err := rt.DetachBuffer(shm.Key()); err != nil {
		t.Fatalf("DetachBuffer failed: %v", err)
	}

	// Verify buffer is detached
	resp, err = rt.SendCommand(Command{
		Cmd: "getBufferInfo",
		Data: map[string]any{
			"key": shm.Key(),
		},
	})
	if err != nil {
		t.Fatalf("getBufferInfo after detach failed: %v", err)
	}

	// Should fail because buffer is detached
	if resp.Success {
		t.Error("getBufferInfo should fail after detach")
	}
}

func TestNodeJSRuntime_SharedMemoryRoundtrip(t *testing.T) {
	path := getPluginHostPath(t)
	rt, err := NewNodeJSRuntime(WithPluginHostPath(path))
	if err != nil {
		t.Fatalf("NewNodeJSRuntime failed: %v", err)
	}

	if err := rt.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer rt.Stop()

	// Create a more complex test AST
	child := &TestNode{nodeType: "Keyword", Value: "inherit"}
	parent := &TestNode{
		nodeType: "Value",
		Children: []*TestNode{child},
		Parens:   true,
	}

	// Flatten the AST
	flat, err := FlattenAST(parent)
	if err != nil {
		t.Fatalf("FlattenAST failed: %v", err)
	}

	// Write to shared memory
	shm, err := rt.WriteASTBuffer(flat)
	if err != nil {
		t.Fatalf("WriteASTBuffer failed: %v", err)
	}

	// Attach buffer in Node.js
	if err := rt.AttachBuffer(shm); err != nil {
		t.Fatalf("AttachBuffer failed: %v", err)
	}

	// Read back and verify
	restored, err := rt.ReadASTBuffer(shm)
	if err != nil {
		t.Fatalf("ReadASTBuffer failed: %v", err)
	}

	// Unflatten
	root, err := UnflattenAST(restored)
	if err != nil {
		t.Fatalf("UnflattenAST failed: %v", err)
	}

	// Verify structure
	if root.Type != "Value" {
		t.Errorf("Root type = %q, want 'Value'", root.Type)
	}
	if !root.Parens {
		t.Error("Root parens should be true")
	}
	if len(root.Children) != 1 {
		t.Errorf("Children count = %d, want 1", len(root.Children))
	}
	if root.Children[0].Type != "Keyword" {
		t.Errorf("Child type = %q, want 'Keyword'", root.Children[0].Type)
	}

	// Clean up
	if err := rt.DetachBuffer(shm.Key()); err != nil {
		t.Fatalf("DetachBuffer failed: %v", err)
	}
}

// Benchmark: Compare JSON vs Shared Memory transfer

func BenchmarkASTTransfer_JSON(b *testing.B) {
	path := "plugin-host.js"
	if _, err := os.Stat(path); err != nil {
		b.Skip("plugin-host.js not found")
	}

	rt, err := NewNodeJSRuntime(WithPluginHostPath(path))
	if err != nil {
		b.Fatalf("NewNodeJSRuntime failed: %v", err)
	}

	if err := rt.Start(); err != nil {
		b.Fatalf("Start failed: %v", err)
	}
	defer rt.Stop()

	// Create test data
	flat := NewFlatAST()
	for i := 0; i < 100; i++ {
		flat.AddString("test string " + string(rune(i)))
		flat.AddNode(FlatNode{
			TypeID: TypeDimension,
			Flags:  FlagParens,
		})
	}

	// Serialize for JSON transfer
	data, err := flat.ToBytes()
	if err != nil {
		b.Fatalf("ToBytes failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Echo the data through JSON (simulates JSON-based transfer)
		_, err := rt.Echo(data)
		if err != nil {
			b.Fatalf("Echo failed: %v", err)
		}
	}
}

func BenchmarkASTTransfer_SharedMemory(b *testing.B) {
	path := "plugin-host.js"
	if _, err := os.Stat(path); err != nil {
		b.Skip("plugin-host.js not found")
	}

	rt, err := NewNodeJSRuntime(WithPluginHostPath(path))
	if err != nil {
		b.Fatalf("NewNodeJSRuntime failed: %v", err)
	}

	if err := rt.Start(); err != nil {
		b.Fatalf("Start failed: %v", err)
	}
	defer rt.Stop()

	// Create test data
	flat := NewFlatAST()
	for i := 0; i < 100; i++ {
		flat.AddString("test string " + string(rune(i)))
		flat.AddNode(FlatNode{
			TypeID: TypeDimension,
			Flags:  FlagParens,
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Write to shared memory
		shm, err := rt.WriteASTBuffer(flat)
		if err != nil {
			b.Fatalf("WriteASTBuffer failed: %v", err)
		}

		// Attach in Node.js
		if err := rt.AttachBuffer(shm); err != nil {
			b.Fatalf("AttachBuffer failed: %v", err)
		}

		// Detach and cleanup
		if err := rt.DetachBuffer(shm.Key()); err != nil {
			b.Fatalf("DetachBuffer failed: %v", err)
		}

		if err := rt.DestroySharedMemory(shm); err != nil {
			b.Fatalf("DestroySharedMemory failed: %v", err)
		}
	}
}

// BenchmarkASTTransfer_SharedMemory_Reuse measures shared memory with reused segment.
// This is more representative of real usage where we'd reuse the same segment.
func BenchmarkASTTransfer_SharedMemory_Reuse(b *testing.B) {
	path := "plugin-host.js"
	if _, err := os.Stat(path); err != nil {
		b.Skip("plugin-host.js not found")
	}

	rt, err := NewNodeJSRuntime(WithPluginHostPath(path))
	if err != nil {
		b.Fatalf("NewNodeJSRuntime failed: %v", err)
	}

	if err := rt.Start(); err != nil {
		b.Fatalf("Start failed: %v", err)
	}
	defer rt.Stop()

	// Create test data
	flat := NewFlatAST()
	for i := 0; i < 100; i++ {
		flat.AddString("test string " + string(rune(i)))
		flat.AddNode(FlatNode{
			TypeID: TypeDimension,
			Flags:  FlagParens,
		})
	}

	// Serialize once
	data, err := flat.ToBytes()
	if err != nil {
		b.Fatalf("ToBytes failed: %v", err)
	}

	// Create shared memory segment once (oversized for safety)
	shm, err := rt.CreateSharedMemory(len(data) * 2)
	if err != nil {
		b.Fatalf("CreateSharedMemory failed: %v", err)
	}
	defer rt.DestroySharedMemory(shm)

	// Attach once
	if err := rt.AttachBuffer(shm); err != nil {
		b.Fatalf("AttachBuffer failed: %v", err)
	}
	defer rt.DetachBuffer(shm.Key())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Write to existing shared memory
		if err := shm.WriteAll(data); err != nil {
			b.Fatalf("WriteAll failed: %v", err)
		}

		// Ping to force roundtrip (simulates Node.js reading the buffer)
		if err := rt.Ping(); err != nil {
			b.Fatalf("Ping failed: %v", err)
		}
	}
}

// BenchmarkSharedMemory_WriteOnly measures just the write to shared memory (no IPC).
func BenchmarkSharedMemory_WriteOnly(b *testing.B) {
	manager, err := NewSharedMemoryManager()
	if err != nil {
		b.Fatalf("NewSharedMemoryManager failed: %v", err)
	}
	defer manager.DestroyAll()

	// Create test data
	flat := NewFlatAST()
	for i := 0; i < 100; i++ {
		flat.AddString("test string " + string(rune(i)))
		flat.AddNode(FlatNode{
			TypeID: TypeDimension,
			Flags:  FlagParens,
		})
	}

	data, err := flat.ToBytes()
	if err != nil {
		b.Fatalf("ToBytes failed: %v", err)
	}

	shm, err := manager.Create(len(data))
	if err != nil {
		b.Fatalf("Create failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := shm.WriteAll(data); err != nil {
			b.Fatalf("WriteAll failed: %v", err)
		}
	}
}
