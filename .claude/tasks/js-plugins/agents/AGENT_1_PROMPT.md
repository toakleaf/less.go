# AGENT 1: Node.js Process & AST Serialization

**Status**: ✅ Can start immediately
**Dependencies**: None
**Estimated Time**: 1 week
**Blocks**: Agents 3, 4, 5

---

You are implementing the foundation for JavaScript plugin support in less.go using a Node.js + shared memory hybrid approach.

## Your Mission

Implement Phase 1 (Node.js Process Integration) and Phase 2 (AST Serialization) from the strategy document.

## Required Reading

BEFORE starting, read these files in `.claude/tasks/js-plugins/`:
1. IMPLEMENTATION_STRATEGY.md - Focus on Phase 1 and Phase 2
2. QUICKSTART.md - Development workflow
3. CHANGES.md - Why we chose Node.js + shared memory

## Your Tasks

### Phase 1: Node.js Process Integration

1. **Create package structure**:
   - Create `packages/less/src/less/less_go/runtime/` directory
   - Create `packages/less/src/less/less_go/runtime/nodejs_runtime.go`
   - Create `packages/less/src/less/less_go/runtime/nodejs_runtime_test.go`

2. **Implement Node.js process management**:
   ```go
   type NodeJSRuntime struct {
       process    *exec.Cmd
       stdin      io.WriteCloser
       stdout     io.ReadCloser
       stderr     io.ReadCloser
       alive      bool
       commands   chan Command
       responses  chan Response
   }

   func NewNodeJSRuntime() (*NodeJSRuntime, error)
   func (rt *NodeJSRuntime) Start() error
   func (rt *NodeJSRuntime) Stop() error
   func (rt *NodeJSRuntime) SendCommand(cmd Command) (Response, error)
   ```

3. **Create plugin-host.js**:
   - Create `packages/less/src/less/less_go/runtime/plugin-host.js`
   - Implement IPC protocol (read from stdin, write to stdout)
   - Handle basic commands: ping, echo (for testing)
   - Clean error handling

4. **Implement shared memory** (can defer to after basic IPC works):
   - Research Go shared memory libraries (consider `github.com/gen2brain/shm`)
   - Implement `SharedMemory` wrapper in Go
   - Implement Node.js shared memory access (use `shm-typed-array` or similar)
   - Create/destroy shared memory segments

### Phase 2: AST Serialization

1. **Design flat buffer format**:
   ```go
   type FlatNode struct {
       TypeID      uint16  // Index into type table
       ChildIndex  uint32  // First child (0 if none)
       NextIndex   uint32  // Next sibling (0 if none)
       ParentIndex uint32  // Parent node (0 if root)
       ValueIndex  uint32  // Index into string table
       PropsOffset uint32  // Offset into properties buffer
   }

   type FlatAST struct {
       Nodes       []FlatNode
       TypeTable   []string
       StringTable []string
       PropBuffer  []byte
   }
   ```

2. **Implement AST flattening**:
   ```go
   func FlattenAST(root Node) (*FlatAST, error)
   ```
   - Walk AST depth-first
   - Assign sequential indices
   - Build type and string tables
   - Handle all node types from `tree/` package

3. **Implement AST unflattening**:
   ```go
   func UnflattenAST(flat *FlatAST) (Node, error)
   ```
   - Reconstruct Go tree from flat buffer
   - Restore parent/child relationships

4. **Implement buffer serialization**:
   ```go
   func (flat *FlatAST) ToBytes() []byte
   func FromBytes(data []byte) (*FlatAST, error)
   ```

5. **Write to shared memory**:
   ```go
   func (rt *NodeJSRuntime) WriteASTBuffer(flat *FlatAST) (shmKey string, error)
   func (rt *NodeJSRuntime) ReadASTBuffer(shmKey string) (*FlatAST, error)
   ```

## Success Criteria

✅ **Phase 1 Complete When**:
- Can spawn Node.js process from Go
- Can send/receive JSON commands via stdin/stdout
- Process lifecycle managed (start, stop, error handling)
- Shared memory segments can be created and accessed from both sides
- Unit tests pass: `go test ./runtime`

✅ **Phase 2 Complete When**:
- Can flatten any AST node to buffer format
- Can unflatten buffer back to identical AST
- Roundtrip tests pass for all node types
- Buffer can be written to/read from shared memory
- Benchmarks show < 10ms flatten time for typical AST
- Unit tests pass: `go test ./runtime`

✅ **No Regressions**:
- ALL existing tests still pass: `pnpm -w test:go:unit` (100%)
- NO integration test regressions: `pnpm -w test:go` (183/183)

## Test Requirements

Write comprehensive tests:

```go
// Test Node.js process lifecycle
func TestNodeJSRuntime_StartStop(t *testing.T)
func TestNodeJSRuntime_Ping(t *testing.T)
func TestNodeJSRuntime_CommandResponse(t *testing.T)

// Test shared memory
func TestSharedMemory_CreateDestroy(t *testing.T)
func TestSharedMemory_ReadWrite(t *testing.T)

// Test AST serialization
func TestFlattenAST_SimpleRuleset(t *testing.T)
func TestFlattenAST_AllNodeTypes(t *testing.T)
func TestRoundtrip_PreservesStructure(t *testing.T)
func BenchmarkFlattenAST(b *testing.B)
```

Run tests frequently:
```bash
# Your tests
go test -v ./packages/less/src/less/less_go/runtime

# No regressions
pnpm -w test:go:unit
pnpm -w test:go
```

## Deliverables

When complete, provide:
1. Working Node.js process spawning and IPC
2. Working shared memory access from both Go and Node.js
3. Complete AST serialization/deserialization
4. All unit tests passing
5. No regressions in existing tests
6. Brief summary of implementation and any gotchas discovered

## Files to Reference

- JavaScript implementation: `packages/less/src/less/plugin-manager.js`
- Go AST nodes: `packages/less/src/less/less_go/tree/*.go`
- Integration tests: `packages/less/src/less/less_go/integration_suite_test.go`

## Important Notes

- Node.js is already available (used by pnpm workspace)
- Start simple: Get basic IPC working before shared memory
- Test incrementally: Don't try to flatten all node types at once
- Reference the JavaScript AST when designing flat format
- Keep Node.js process alive (don't spawn per-command)

Good luck! You're building the foundation for the entire plugin system. 🚀
