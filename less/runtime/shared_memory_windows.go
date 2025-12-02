//go:build windows

package runtime

import (
	"fmt"
	"sync"
)

// ErrNotSupported is returned when shared memory operations are attempted on Windows.
var ErrNotSupported = fmt.Errorf("shared memory is not supported on Windows")

// SharedMemory represents a shared memory segment backed by a memory-mapped file.
// This provides zero-copy data transfer between Go and Node.js processes.
// Note: This implementation is a stub for Windows where mmap is not available.
type SharedMemory struct {
	key      string
	path     string
	size     int
	data     []byte
	readonly bool
	mu       sync.RWMutex
}

// SharedMemoryManager manages shared memory segments for a runtime.
type SharedMemoryManager struct {
	segments map[string]*SharedMemory
	tempDir  string
	mu       sync.RWMutex
}

// NewSharedMemoryManager creates a new shared memory manager.
// On Windows, this returns an error as shared memory is not supported.
func NewSharedMemoryManager() (*SharedMemoryManager, error) {
	return nil, ErrNotSupported
}

// Create creates a new shared memory segment of the specified size.
func (m *SharedMemoryManager) Create(size int) (*SharedMemory, error) {
	return nil, ErrNotSupported
}

// Open opens an existing shared memory segment by key (read-only).
func (m *SharedMemoryManager) Open(key string) (*SharedMemory, error) {
	return nil, ErrNotSupported
}

// Get retrieves a shared memory segment by key.
func (m *SharedMemoryManager) Get(key string) *SharedMemory {
	return nil
}

// Destroy destroys a shared memory segment.
func (m *SharedMemoryManager) Destroy(key string) error {
	return ErrNotSupported
}

// DestroyAll destroys all shared memory segments and cleans up.
func (m *SharedMemoryManager) DestroyAll() error {
	return ErrNotSupported
}

// Key returns the unique identifier for this segment.
func (s *SharedMemory) Key() string {
	return s.key
}

// Path returns the file path for this segment.
func (s *SharedMemory) Path() string {
	return s.path
}

// Size returns the size of the mapped region.
func (s *SharedMemory) Size() int {
	return s.size
}

// Data returns the underlying byte slice.
func (s *SharedMemory) Data() []byte {
	return nil
}

// Write writes data to the shared memory segment at the specified offset.
func (s *SharedMemory) Write(offset int, data []byte) error {
	return ErrNotSupported
}

// Read reads data from the shared memory segment.
func (s *SharedMemory) Read(offset, length int) ([]byte, error) {
	return nil, ErrNotSupported
}

// ReadAll reads all data from the shared memory segment.
func (s *SharedMemory) ReadAll() ([]byte, error) {
	return nil, ErrNotSupported
}

// WriteAll writes data to the beginning of the segment.
func (s *SharedMemory) WriteAll(data []byte) error {
	return ErrNotSupported
}

// Sync flushes any cached writes to the backing file.
func (s *SharedMemory) Sync() error {
	return ErrNotSupported
}

// Close unmaps the memory and closes the file.
func (s *SharedMemory) Close() error {
	return nil
}
