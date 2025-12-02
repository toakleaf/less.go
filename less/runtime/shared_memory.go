package runtime

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"syscall"
)

// SharedMemory represents a shared memory segment backed by a memory-mapped file.
// This provides zero-copy data transfer between Go and Node.js processes.
type SharedMemory struct {
	key      string   // Unique identifier for this segment
	path     string   // Path to the backing file
	size     int      // Size of the mapped region
	data     []byte   // The mmap'd byte slice
	file     *os.File // The backing file
	readonly bool     // Whether this segment is read-only
	mu       sync.RWMutex
}

// SharedMemoryManager manages shared memory segments for a runtime.
type SharedMemoryManager struct {
	segments map[string]*SharedMemory
	tempDir  string
	mu       sync.RWMutex
}

// NewSharedMemoryManager creates a new shared memory manager.
func NewSharedMemoryManager() (*SharedMemoryManager, error) {
	// Create a temporary directory for shared memory files
	tempDir, err := os.MkdirTemp("", "less-go-shm-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	return &SharedMemoryManager{
		segments: make(map[string]*SharedMemory),
		tempDir:  tempDir,
	}, nil
}

// generateKey generates a unique key for a shared memory segment.
func generateKey() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Create creates a new shared memory segment of the specified size.
func (m *SharedMemoryManager) Create(size int) (*SharedMemory, error) {
	if size <= 0 {
		return nil, fmt.Errorf("size must be positive")
	}

	key, err := generateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	path := filepath.Join(m.tempDir, key)

	// Create and size the file
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}

	// Truncate to the desired size
	if err := file.Truncate(int64(size)); err != nil {
		file.Close()
		os.Remove(path)
		return nil, fmt.Errorf("failed to truncate file: %w", err)
	}

	// Memory map the file
	data, err := syscall.Mmap(int(file.Fd()), 0, size,
		syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		file.Close()
		os.Remove(path)
		return nil, fmt.Errorf("failed to mmap: %w", err)
	}

	shm := &SharedMemory{
		key:  key,
		path: path,
		size: size,
		data: data,
		file: file,
	}

	m.mu.Lock()
	m.segments[key] = shm
	m.mu.Unlock()

	return shm, nil
}

// Open opens an existing shared memory segment by key (read-only).
func (m *SharedMemoryManager) Open(key string) (*SharedMemory, error) {
	path := filepath.Join(m.tempDir, key)

	file, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	// Get file size
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}
	size := int(info.Size())

	// Memory map the file (read-only)
	data, err := syscall.Mmap(int(file.Fd()), 0, size,
		syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to mmap: %w", err)
	}

	shm := &SharedMemory{
		key:      key,
		path:     path,
		size:     size,
		data:     data,
		file:     file,
		readonly: true,
	}

	return shm, nil
}

// Get retrieves a shared memory segment by key.
func (m *SharedMemoryManager) Get(key string) *SharedMemory {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.segments[key]
}

// Destroy destroys a shared memory segment.
func (m *SharedMemoryManager) Destroy(key string) error {
	m.mu.Lock()
	shm, ok := m.segments[key]
	if ok {
		delete(m.segments, key)
	}
	m.mu.Unlock()

	if !ok {
		return fmt.Errorf("segment not found: %s", key)
	}

	return shm.Close()
}

// DestroyAll destroys all shared memory segments and cleans up.
func (m *SharedMemoryManager) DestroyAll() error {
	m.mu.Lock()
	segments := make([]*SharedMemory, 0, len(m.segments))
	for _, shm := range m.segments {
		segments = append(segments, shm)
	}
	m.segments = make(map[string]*SharedMemory)
	m.mu.Unlock()

	var lastErr error
	for _, shm := range segments {
		if err := shm.Close(); err != nil {
			lastErr = err
		}
	}

	// Remove temp directory
	if err := os.RemoveAll(m.tempDir); err != nil {
		lastErr = err
	}

	return lastErr
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

// Data returns the underlying byte slice (read-only copy).
func (s *SharedMemory) Data() []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data
}

// Write writes data to the shared memory segment at the specified offset.
func (s *SharedMemory) Write(offset int, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.readonly {
		return fmt.Errorf("segment is read-only")
	}

	if offset < 0 || offset+len(data) > s.size {
		return fmt.Errorf("write out of bounds: offset=%d, len=%d, size=%d",
			offset, len(data), s.size)
	}

	copy(s.data[offset:], data)
	return nil
}

// Read reads data from the shared memory segment.
func (s *SharedMemory) Read(offset, length int) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if offset < 0 || offset+length > s.size {
		return nil, fmt.Errorf("read out of bounds: offset=%d, len=%d, size=%d",
			offset, length, s.size)
	}

	result := make([]byte, length)
	copy(result, s.data[offset:offset+length])
	return result, nil
}

// ReadAll reads all data from the shared memory segment.
func (s *SharedMemory) ReadAll() ([]byte, error) {
	return s.Read(0, s.size)
}

// WriteAll writes data to the beginning of the segment.
func (s *SharedMemory) WriteAll(data []byte) error {
	return s.Write(0, data)
}

// Sync flushes any cached writes to the backing file.
// This is critical for IPC with processes that read the file directly
// instead of memory-mapping it (like Node.js using fs.readFileSync).
func (s *SharedMemory) Sync() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.readonly {
		return nil
	}

	// For cross-process visibility when the reader uses file I/O (not mmap),
	// we need to ensure the mmap'd data is flushed to the file.
	// Since Go's syscall package doesn't have Msync on all platforms,
	// we use file.Sync() to flush the file's data to disk.
	if s.file != nil {
		return s.file.Sync()
	}

	return nil
}

// Close unmaps the memory and closes the file.
func (s *SharedMemory) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var lastErr error

	if s.data != nil {
		if err := syscall.Munmap(s.data); err != nil {
			lastErr = fmt.Errorf("failed to munmap: %w", err)
		}
		s.data = nil
	}

	if s.file != nil {
		if err := s.file.Close(); err != nil && lastErr == nil {
			lastErr = fmt.Errorf("failed to close file: %w", err)
		}
		s.file = nil
	}

	// Remove the backing file (if we created it)
	if !s.readonly && s.path != "" {
		if err := os.Remove(s.path); err != nil && lastErr == nil {
			lastErr = fmt.Errorf("failed to remove file: %w", err)
		}
	}

	return lastErr
}
