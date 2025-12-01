package runtime

import (
	"encoding/binary"
	"fmt"
	"math"
	"sync"
)

// SharedMemoryProtocol implements a high-performance IPC protocol using
// a persistent shared memory region with atomic signaling. This eliminates
// JSON serialization overhead and reduces IPC latency to near-zero.
//
// Memory Layout (4MB default):
// ┌─────────────────────────────────────────────────────────────────┐
// │ Control Block (4KB)                                              │
// │   [0x000] request_ready:  uint32 (Go sets to 1 when ready)       │
// │   [0x004] response_ready: uint32 (JS sets to 1 when done)        │
// │   [0x008] function_id:    uint32 (function to call)              │
// │   [0x00C] arg_count:      uint32 (number of arguments)           │
// │   [0x010] arg_offsets:    [16]uint32 (offsets into args section) │
// │   [0x050] result_offset:  uint32                                 │
// │   [0x054] result_size:    uint32                                 │
// │   [0x058] error_flag:     uint32 (1 if error)                    │
// │   [0x05C] error_offset:   uint32                                 │
// │   [0x060] error_size:     uint32                                 │
// │   [0x064] shutdown:       uint32 (1 to signal shutdown)          │
// │   [0x068] js_ready:       uint32 (JS sets to 1 when initialized) │
// ├─────────────────────────────────────────────────────────────────┤
// │ Variables Section (1MB) - Pre-populated at compilation start     │
// │   Binary format as defined in binary_variables.go                │
// ├─────────────────────────────────────────────────────────────────┤
// │ Arguments Section (1MB) - Go writes function args here           │
// │   Binary format: [type:1][...data] for each arg                  │
// ├─────────────────────────────────────────────────────────────────┤
// │ Results Section (1MB) - JS writes results here                   │
// │   Binary format: [type:1][...data]                               │
// ├─────────────────────────────────────────────────────────────────┤
// │ Error Buffer (64KB) - For error messages                         │
// └─────────────────────────────────────────────────────────────────┘

const (
	// Section sizes
	ControlBlockSize  = 4 * 1024           // 4KB
	VariablesSectionSize = 1 * 1024 * 1024   // 1MB
	ArgsSectionSize      = 1 * 1024 * 1024   // 1MB
	ResultsSectionSize   = 1 * 1024 * 1024   // 1MB
	ErrorBufferSize      = 64 * 1024         // 64KB
	TotalSHMSize         = ControlBlockSize + VariablesSectionSize + ArgsSectionSize + ResultsSectionSize + ErrorBufferSize

	// Section offsets
	ControlBlockOffset  = 0
	VariablesSectionOffset = ControlBlockSize
	ArgsSectionOffset      = VariablesSectionOffset + VariablesSectionSize
	ResultsSectionOffset   = ArgsSectionOffset + ArgsSectionSize
	ErrorBufferOffset      = ResultsSectionOffset + ResultsSectionSize

	// Control block field offsets (relative to control block start)
	OffsetRequestReady  = 0x000
	OffsetResponseReady = 0x004
	OffsetFunctionID    = 0x008
	OffsetArgCount      = 0x00C
	OffsetArgOffsets    = 0x010 // Array of 16 uint32s
	OffsetResultOffset  = 0x050
	OffsetResultSize    = 0x054
	OffsetErrorFlag     = 0x058
	OffsetErrorOffset   = 0x05C
	OffsetErrorSize     = 0x060
	OffsetShutdown      = 0x064
	OffsetJSReady       = 0x068

	MaxArgs = 16

	// Argument types for binary serialization
	ArgTypeNull       = 0
	ArgTypeDimension  = 1
	ArgTypeColor      = 2
	ArgTypeQuoted     = 3
	ArgTypeKeyword    = 4
	ArgTypeExpression = 5
	ArgTypeAnonymous  = 6
	ArgTypeVariable   = 7
	ArgTypeNumber     = 8
	ArgTypeBoolean    = 9
	ArgTypeCall       = 10
)

// SharedMemoryProtocol manages a persistent shared memory region for
// high-performance IPC between Go and Node.js.
type SharedMemoryProtocol struct {
	shm          *SharedMemory
	functionMap  map[string]uint32 // function name -> ID
	functionList []string          // ID -> function name
	mu           sync.RWMutex

	// Pointers to sections (for convenience)
	controlBlock    []byte
	variablesSection []byte
	argsSection     []byte
	resultsSection  []byte
	errorBuffer     []byte

	// Variable index for fast lookups
	variableIndex map[string]uint32 // variable name -> offset in variables section
	varIndexMu    sync.RWMutex

	// Write position tracking
	argsWritePos uint32
}

// NewSharedMemoryProtocol creates a new shared memory protocol instance.
func NewSharedMemoryProtocol(shmManager *SharedMemoryManager) (*SharedMemoryProtocol, error) {
	if shmManager == nil {
		return nil, fmt.Errorf("shared memory manager is nil")
	}

	// Create the shared memory segment
	shm, err := shmManager.Create(TotalSHMSize)
	if err != nil {
		return nil, fmt.Errorf("failed to create shared memory: %w", err)
	}

	// Get direct access to the data
	data := shm.Data()

	protocol := &SharedMemoryProtocol{
		shm:              shm,
		functionMap:      make(map[string]uint32),
		functionList:     make([]string, 0),
		controlBlock:     data[ControlBlockOffset : ControlBlockOffset+ControlBlockSize],
		variablesSection: data[VariablesSectionOffset : VariablesSectionOffset+VariablesSectionSize],
		argsSection:      data[ArgsSectionOffset : ArgsSectionOffset+ArgsSectionSize],
		resultsSection:   data[ResultsSectionOffset : ResultsSectionOffset+ResultsSectionSize],
		errorBuffer:      data[ErrorBufferOffset : ErrorBufferOffset+ErrorBufferSize],
		variableIndex:    make(map[string]uint32),
	}

	// Initialize control block
	protocol.clearControlBlock()

	return protocol, nil
}

// Path returns the path to the shared memory file.
func (p *SharedMemoryProtocol) Path() string {
	return p.shm.Path()
}

// Key returns the unique key for this shared memory segment.
func (p *SharedMemoryProtocol) Key() string {
	return p.shm.Key()
}

// Close releases the shared memory resources.
func (p *SharedMemoryProtocol) Close() error {
	// Signal shutdown
	p.setUint32(OffsetShutdown, 1)
	p.shm.Sync()

	return p.shm.Close()
}

// clearControlBlock initializes the control block to zero state.
func (p *SharedMemoryProtocol) clearControlBlock() {
	for i := range p.controlBlock {
		p.controlBlock[i] = 0
	}
}

// RegisterFunction registers a function and returns its ID.
func (p *SharedMemoryProtocol) RegisterFunction(name string) uint32 {
	p.mu.Lock()
	defer p.mu.Unlock()

	if id, exists := p.functionMap[name]; exists {
		return id
	}

	id := uint32(len(p.functionList))
	p.functionMap[name] = id
	p.functionList = append(p.functionList, name)
	return id
}

// GetFunctionID returns the ID for a function name.
func (p *SharedMemoryProtocol) GetFunctionID(name string) (uint32, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	id, ok := p.functionMap[name]
	return id, ok
}

// GetFunctionName returns the name for a function ID.
func (p *SharedMemoryProtocol) GetFunctionName(id uint32) (string, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if int(id) >= len(p.functionList) {
		return "", false
	}
	return p.functionList[id], true
}

// WriteVariables writes all variables to the variables section in binary format.
// This should be called once at the start of compilation.
func (p *SharedMemoryProtocol) WriteVariables(variables map[string]any) error {
	p.varIndexMu.Lock()
	defer p.varIndexMu.Unlock()

	// Clear the variables index
	p.variableIndex = make(map[string]uint32)

	// Write variables using the existing binary format
	data := WritePrefetchedVariables(variables)

	if len(data) > VariablesSectionSize {
		return fmt.Errorf("variables exceed section size: %d > %d", len(data), VariablesSectionSize)
	}

	copy(p.variablesSection, data)

	// Build the index by parsing the header
	if len(data) >= 12 {
		magic := binary.LittleEndian.Uint32(data[0:4])
		if magic == PrefetchMagic {
			varCount := binary.LittleEndian.Uint32(data[8:12])
			offset := uint32(12)

			for i := uint32(0); i < varCount && offset < uint32(len(data)); i++ {
				if offset+4 > uint32(len(data)) {
					break
				}
				nameLen := binary.LittleEndian.Uint32(data[offset : offset+4])
				offset += 4

				if offset+nameLen > uint32(len(data)) {
					break
				}
				name := string(data[offset : offset+nameLen])
				p.variableIndex[name] = offset - 4 // Store offset to the name length field
				offset += nameLen

				// Skip important flag and type
				offset += 2

				// Skip value data based on type (simplified - just mark position)
				// The actual parsing happens on the JS side
			}
		}
	}

	return p.shm.Sync()
}

// PrepareCall prepares a function call in shared memory.
// Returns the offset where arguments should be written.
func (p *SharedMemoryProtocol) PrepareCall(functionID uint32, argCount int) error {
	if argCount > MaxArgs {
		return fmt.Errorf("too many arguments: %d > %d", argCount, MaxArgs)
	}

	// Clear request/response flags
	p.setUint32(OffsetRequestReady, 0)
	p.setUint32(OffsetResponseReady, 0)
	p.setUint32(OffsetErrorFlag, 0)

	// Set function ID and arg count
	p.setUint32(OffsetFunctionID, functionID)
	p.setUint32(OffsetArgCount, uint32(argCount))

	// Reset args write position
	p.argsWritePos = 0

	return nil
}

// WriteArg writes an argument to the arguments section.
// Returns the offset where the argument was written.
func (p *SharedMemoryProtocol) WriteArg(argIndex int, value any) (uint32, error) {
	if argIndex >= MaxArgs {
		return 0, fmt.Errorf("argument index out of range: %d >= %d", argIndex, MaxArgs)
	}

	startOffset := p.argsWritePos

	// Write the argument value
	written, err := p.writeValue(p.argsSection, p.argsWritePos, value)
	if err != nil {
		return 0, err
	}

	p.argsWritePos += written

	// Store the offset in the control block
	p.setUint32(OffsetArgOffsets+uint32(argIndex*4), startOffset)

	return startOffset, nil
}

// SignalRequest signals that the request is ready.
func (p *SharedMemoryProtocol) SignalRequest() error {
	// Memory barrier before signaling
	p.setUint32(OffsetRequestReady, 1)
	return p.shm.Sync()
}

// WaitForResponse waits for the JavaScript side to signal completion.
// This uses busy-waiting with file re-reads to detect changes from Node.js.
func (p *SharedMemoryProtocol) WaitForResponse(maxWaitMs int) (bool, error) {
	for i := 0; i < maxWaitMs*10; i++ { // Check every 100µs
		// Re-read the control block from the file
		// This is needed because Node.js uses file I/O, not mmap
		p.shm.Sync() // Ensure mmap is synced

		if p.getUint32(OffsetResponseReady) == 1 {
			return true, nil
		}

		// Small sleep to avoid spinning too hard
		if i%10 == 0 {
			// Sleep 100µs every 10 iterations
		}
	}
	return false, fmt.Errorf("timeout waiting for response")
}

// ReadResult reads the result from the results section.
func (p *SharedMemoryProtocol) ReadResult() (any, error) {
	// Check for error
	if p.getUint32(OffsetErrorFlag) == 1 {
		errorOffset := p.getUint32(OffsetErrorOffset)
		errorSize := p.getUint32(OffsetErrorSize)
		if errorSize > 0 && errorOffset+errorSize <= ErrorBufferSize {
			errorMsg := string(p.errorBuffer[errorOffset : errorOffset+errorSize])
			return nil, fmt.Errorf("JavaScript error: %s", errorMsg)
		}
		return nil, fmt.Errorf("JavaScript error (no message)")
	}

	resultOffset := p.getUint32(OffsetResultOffset)
	resultSize := p.getUint32(OffsetResultSize)

	if resultSize == 0 {
		return nil, nil
	}

	return p.readValue(p.resultsSection, resultOffset, resultSize)
}

// ClearResponse clears the response ready flag for the next call.
func (p *SharedMemoryProtocol) ClearResponse() {
	p.setUint32(OffsetResponseReady, 0)
	p.setUint32(OffsetRequestReady, 0)
}

// IsJSReady checks if JavaScript has signaled it's ready.
func (p *SharedMemoryProtocol) IsJSReady() bool {
	return p.getUint32(OffsetJSReady) == 1
}

// Helper methods for control block access

func (p *SharedMemoryProtocol) getUint32(offset uint32) uint32 {
	return binary.LittleEndian.Uint32(p.controlBlock[offset:])
}

func (p *SharedMemoryProtocol) setUint32(offset uint32, value uint32) {
	binary.LittleEndian.PutUint32(p.controlBlock[offset:], value)
}

// writeValue writes a value to a buffer and returns bytes written.
func (p *SharedMemoryProtocol) writeValue(buf []byte, offset uint32, value any) (uint32, error) {
	if value == nil {
		buf[offset] = ArgTypeNull
		return 1, nil
	}

	pos := offset

	// Try to get type from the value
	typer, ok := value.(interface{ GetType() string })
	if !ok {
		// Handle primitive types
		switch v := value.(type) {
		case string:
			buf[pos] = ArgTypeKeyword
			pos++
			pos += p.writeString(buf, pos, v)
			return pos - offset, nil
		case float64:
			buf[pos] = ArgTypeNumber
			pos++
			binary.LittleEndian.PutUint64(buf[pos:], math.Float64bits(v))
			pos += 8
			return pos - offset, nil
		case int:
			buf[pos] = ArgTypeNumber
			pos++
			binary.LittleEndian.PutUint64(buf[pos:], math.Float64bits(float64(v)))
			pos += 8
			return pos - offset, nil
		case bool:
			buf[pos] = ArgTypeBoolean
			pos++
			if v {
				buf[pos] = 1
			} else {
				buf[pos] = 0
			}
			pos++
			return pos - offset, nil
		default:
			buf[pos] = ArgTypeNull
			return 1, nil
		}
	}

	nodeType := typer.GetType()
	switch nodeType {
	case "Dimension":
		buf[pos] = ArgTypeDimension
		pos++
		// Write value
		if getter, ok := value.(interface{ GetValue() float64 }); ok {
			binary.LittleEndian.PutUint64(buf[pos:], math.Float64bits(getter.GetValue()))
		}
		pos += 8
		// Write unit
		unitStr := ""
		if unitGetter, ok := value.(interface{ GetUnit() any }); ok {
			unit := unitGetter.GetUnit()
			if unit != nil {
				if s, ok := unit.(fmt.Stringer); ok {
					unitStr = s.String()
				} else if s, ok := unit.(interface{ ToString() string }); ok {
					unitStr = s.ToString()
				}
			}
		}
		pos += p.writeString(buf, pos, unitStr)

	case "Color":
		buf[pos] = ArgTypeColor
		pos++
		rgb := []float64{0, 0, 0}
		if rgbGetter, ok := value.(interface{ GetRGB() []float64 }); ok {
			rgb = rgbGetter.GetRGB()
		}
		for i := 0; i < 3; i++ {
			if i < len(rgb) {
				binary.LittleEndian.PutUint64(buf[pos:], math.Float64bits(rgb[i]))
			} else {
				binary.LittleEndian.PutUint64(buf[pos:], 0)
			}
			pos += 8
		}
		alpha := 1.0
		if alphaGetter, ok := value.(interface{ GetAlpha() float64 }); ok {
			alpha = alphaGetter.GetAlpha()
		}
		binary.LittleEndian.PutUint64(buf[pos:], math.Float64bits(alpha))
		pos += 8

	case "Quoted":
		buf[pos] = ArgTypeQuoted
		pos++
		strVal := ""
		if getter, ok := value.(interface{ GetValue() string }); ok {
			strVal = getter.GetValue()
		}
		pos += p.writeString(buf, pos, strVal)
		quote := byte('"')
		if getter, ok := value.(interface{ GetQuote() string }); ok {
			q := getter.GetQuote()
			if len(q) > 0 {
				quote = q[0]
			}
		}
		buf[pos] = quote
		pos++
		escaped := byte(0)
		if getter, ok := value.(interface{ GetEscaped() bool }); ok && getter.GetEscaped() {
			escaped = 1
		}
		buf[pos] = escaped
		pos++

	case "Keyword":
		buf[pos] = ArgTypeKeyword
		pos++
		strVal := ""
		if getter, ok := value.(interface{ GetValue() string }); ok {
			strVal = getter.GetValue()
		}
		pos += p.writeString(buf, pos, strVal)

	case "Anonymous":
		buf[pos] = ArgTypeAnonymous
		pos++
		if getter, ok := value.(interface{ GetValue() any }); ok {
			v := getter.GetValue()
			switch val := v.(type) {
			case string:
				pos += p.writeString(buf, pos, val)
			case float64:
				pos += p.writeString(buf, pos, fmt.Sprintf("%g", val))
			case int:
				pos += p.writeString(buf, pos, fmt.Sprintf("%d", val))
			default:
				pos += p.writeString(buf, pos, fmt.Sprintf("%v", v))
			}
		} else {
			pos += p.writeString(buf, pos, "")
		}

	case "Expression", "Value":
		buf[pos] = ArgTypeExpression
		pos++
		if getter, ok := value.(interface{ GetValue() []any }); ok {
			vals := getter.GetValue()
			binary.LittleEndian.PutUint32(buf[pos:], uint32(len(vals)))
			pos += 4
			for _, v := range vals {
				written, _ := p.writeValue(buf, pos, v)
				pos += written
			}
		} else {
			binary.LittleEndian.PutUint32(buf[pos:], 0)
			pos += 4
		}

	default:
		buf[pos] = ArgTypeNull
		pos++
	}

	return pos - offset, nil
}

// writeString writes a length-prefixed string to a buffer.
func (p *SharedMemoryProtocol) writeString(buf []byte, offset uint32, s string) uint32 {
	binary.LittleEndian.PutUint32(buf[offset:], uint32(len(s)))
	copy(buf[offset+4:], s)
	return 4 + uint32(len(s))
}

// readValue reads a value from a buffer.
func (p *SharedMemoryProtocol) readValue(buf []byte, offset, size uint32) (any, error) {
	if size == 0 || offset >= uint32(len(buf)) {
		return nil, nil
	}

	valueType := buf[offset]
	pos := offset + 1

	switch valueType {
	case ArgTypeNull:
		return nil, nil

	case ArgTypeDimension:
		value := math.Float64frombits(binary.LittleEndian.Uint64(buf[pos:]))
		pos += 8
		unitLen := binary.LittleEndian.Uint32(buf[pos:])
		pos += 4
		unit := string(buf[pos : pos+unitLen])
		return &JSResultNode{
			NodeType: "Dimension",
			Properties: map[string]any{
				"value": value,
				"unit":  unit,
			},
		}, nil

	case ArgTypeColor:
		r := math.Float64frombits(binary.LittleEndian.Uint64(buf[pos:]))
		pos += 8
		g := math.Float64frombits(binary.LittleEndian.Uint64(buf[pos:]))
		pos += 8
		b := math.Float64frombits(binary.LittleEndian.Uint64(buf[pos:]))
		pos += 8
		alpha := math.Float64frombits(binary.LittleEndian.Uint64(buf[pos:]))
		return &JSResultNode{
			NodeType: "Color",
			Properties: map[string]any{
				"rgb":   []float64{r, g, b},
				"alpha": alpha,
			},
		}, nil

	case ArgTypeQuoted:
		strLen := binary.LittleEndian.Uint32(buf[pos:])
		pos += 4
		strVal := string(buf[pos : pos+strLen])
		pos += strLen
		quote := string(buf[pos])
		pos++
		escaped := buf[pos] == 1
		return &JSResultNode{
			NodeType: "Quoted",
			Properties: map[string]any{
				"value":   strVal,
				"quote":   quote,
				"escaped": escaped,
			},
		}, nil

	case ArgTypeKeyword:
		strLen := binary.LittleEndian.Uint32(buf[pos:])
		pos += 4
		strVal := string(buf[pos : pos+strLen])
		return &JSResultNode{
			NodeType: "Keyword",
			Properties: map[string]any{
				"value": strVal,
			},
		}, nil

	case ArgTypeAnonymous:
		strLen := binary.LittleEndian.Uint32(buf[pos:])
		pos += 4
		strVal := string(buf[pos : pos+strLen])
		return &JSResultNode{
			NodeType: "Anonymous",
			Properties: map[string]any{
				"value": strVal,
			},
		}, nil

	case ArgTypeNumber:
		value := math.Float64frombits(binary.LittleEndian.Uint64(buf[pos:]))
		return value, nil

	case ArgTypeBoolean:
		return buf[pos] == 1, nil

	default:
		return nil, fmt.Errorf("unknown value type: %d", valueType)
	}
}

// GetVariablesSectionInfo returns info about the variables section for JS.
func (p *SharedMemoryProtocol) GetVariablesSectionInfo() (offset, size uint32) {
	return VariablesSectionOffset, VariablesSectionSize
}

// GetArgsSectionInfo returns info about the args section for JS.
func (p *SharedMemoryProtocol) GetArgsSectionInfo() (offset, size uint32) {
	return ArgsSectionOffset, ArgsSectionSize
}

// GetResultsSectionInfo returns info about the results section for JS.
func (p *SharedMemoryProtocol) GetResultsSectionInfo() (offset, size uint32) {
	return ResultsSectionOffset, ResultsSectionSize
}

// GetControlBlockLayout returns the control block layout for JS initialization.
func (p *SharedMemoryProtocol) GetControlBlockLayout() map[string]uint32 {
	return map[string]uint32{
		"requestReady":  OffsetRequestReady,
		"responseReady": OffsetResponseReady,
		"functionId":    OffsetFunctionID,
		"argCount":      OffsetArgCount,
		"argOffsets":    OffsetArgOffsets,
		"resultOffset":  OffsetResultOffset,
		"resultSize":    OffsetResultSize,
		"errorFlag":     OffsetErrorFlag,
		"errorOffset":   OffsetErrorOffset,
		"errorSize":     OffsetErrorSize,
		"shutdown":      OffsetShutdown,
		"jsReady":       OffsetJSReady,
	}
}

// GetSectionOffsets returns all section offsets for JS.
func (p *SharedMemoryProtocol) GetSectionOffsets() map[string]uint32 {
	return map[string]uint32{
		"controlBlock":     ControlBlockOffset,
		"variablesSection": VariablesSectionOffset,
		"argsSection":      ArgsSectionOffset,
		"resultsSection":   ResultsSectionOffset,
		"errorBuffer":      ErrorBufferOffset,
	}
}
