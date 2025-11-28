# JavaScript Plugin Implementation for less.go

## 🎯 Mission

Implement JavaScript plugin support for the less.go project to achieve 100% feature parity with less.js, enabling the 8 quarantined plugin tests to pass.

## 📊 Current Status

- **Non-plugin tests**: ✅ 183/183 passing (100%)
- **Plugin tests**: ⏸️ 8 quarantined
  - `plugin` - Comprehensive plugin functionality
  - `plugin-simple` - Basic function registration
  - `plugin-preeval` - Pre-evaluation visitors
  - `plugin-module` - NPM module loading
  - `plugin-tree-nodes` - Node construction
  - Plus 3 more error/edge case tests

- **Implementation progress**: 🔴 Not started
  - Phase 1: Runtime Integration - 🔴 0/4 tasks
  - Phase 2: Serialization - 🔴 0/5 tasks
  - Phase 3: Bindings - 🔴 0/3 tasks
  - Phase 4: Loader - 🔴 0/5 tasks
  - Phase 5: Functions - 🔴 0/5 tasks
  - Phase 6: Visitors - 🔴 0/5 tasks
  - Phase 7: Constructors - 🔴 0/4 tasks
  - Phase 8: Processors - 🔴 0/5 tasks
  - Phase 9: File Managers - 🔴 0/4 tasks
  - Phase 10: Scoping - 🔴 0/5 tasks

## 📚 Documentation

### Start Here (in order)

1. **[IMPLEMENTATION_STRATEGY.md](IMPLEMENTATION_STRATEGY.md)** - Overall architecture, OXC approach, design decisions
2. **[TASK_BREAKDOWN.md](TASK_BREAKDOWN.md)** - Detailed task specifications with acceptance criteria
3. **[QUICKSTART.md](QUICKSTART.md)** - Development workflow, testing, debugging tips

### For Agents

4. **[AGENT_TEMPLATE.md](AGENT_TEMPLATE.md)** - Template for spawning task-specific agents

## 🚀 Quick Start

### For the first agent

```bash
# 1. Read the docs
cat .claude/tasks/js-plugins/IMPLEMENTATION_STRATEGY.md
cat .claude/tasks/js-plugins/TASK_BREAKDOWN.md
cat .claude/tasks/js-plugins/QUICKSTART.md

# 2. Claim Phase 1 in TASK_BREAKDOWN.md
# Update status: 🔴 → 🟡

# 3. Start coding
cd packages/less/src/less/less_go
mkdir runtime
cd runtime

# 4. Follow TDD workflow
# - Write tests
# - Implement functionality
# - Verify no regressions

# 5. Update status when done
# TASK_BREAKDOWN.md: 🟡 → 🟢
```

### For subsequent agents

```bash
# 1. Check TASK_BREAKDOWN.md for available tasks
#    - Look for 🔴 tasks with satisfied dependencies
#    - Claim by updating to 🟡

# 2. Read relevant docs
cat .claude/tasks/js-plugins/IMPLEMENTATION_STRATEGY.md  # Architecture
cat .claude/tasks/js-plugins/TASK_BREAKDOWN.md           # Your task spec

# 3. Implement your phase
#    - Follow QUICKSTART.md workflow
#    - Reference JavaScript implementation
#    - Write comprehensive tests

# 4. Mark complete
#    - Update TASK_BREAKDOWN.md: 🟡 → 🟢
#    - Document any gotchas
```

## 🏗️ Architecture Overview

### The OXC Approach

We're following the OXC project's strategy for running JavaScript from native code:

1. **Raw Transfer**: Pass memory buffers instead of JSON (eliminates 80% of overhead)
2. **Lazy Deserialization**: Proxy objects read buffers on-demand (reduces GC pressure)
3. **Flattened AST**: Array indices instead of pointers (enables buffer storage)

**Result**: 2-5x overhead instead of 10-20x with traditional approaches.

### High-Level Design

```
┌─────────────────────────────────────────────────┐
│              Go LESS Compiler                    │
│                                                  │
│  ┌──────────────────────────────────────────┐  │
│  │ Parse & Evaluate LESS                     │  │
│  │ ┌────────────┐         ┌──────────────┐  │  │
│  │ │   Parser   │  ──────>│  Evaluator   │  │  │
│  │ └────────────┘         └──────┬───────┘  │  │
│  │                                │          │  │
│  │                                │          │  │
│  └────────────────────────────────┼──────────┘  │
│                                   │              │
│  ┌────────────────────────────────▼──────────┐  │
│  │ Plugin System                             │  │
│  │                                           │  │
│  │  ┌─────────────┐    ┌──────────────┐    │  │
│  │  │ AST         │───>│ Flat Buffer  │    │  │
│  │  │ Serializer  │<───│ (Raw Transfer)│   │  │
│  │  └─────────────┘    └──────┬───────┘    │  │
│  │                             │            │  │
│  │  ┌──────────────────────────▼─────────┐ │  │
│  │  │ JavaScript Runtime (goja/v8go)     │ │  │
│  │  │                                    │ │  │
│  │  │  ┌──────────────────────────────┐ │ │  │
│  │  │  │ Plugin Code                  │ │ │  │
│  │  │  │                              │ │ │  │
│  │  │  │ - Add functions              │ │ │  │
│  │  │  │ - Add visitors               │ │ │  │
│  │  │  │ - Transform AST              │ │ │  │
│  │  │  │ - Create nodes               │ │ │  │
│  │  │  └──────────────────────────────┘ │ │  │
│  │  │                                    │ │  │
│  │  │  NodeFacade (Lazy Deser)          │ │  │
│  │  └────────────────────────────────────┘ │  │
│  │                                           │  │
│  └───────────────────────────────────────────┘  │
│                                                  │
└─────────────────────────────────────────────────┘
```

### Plugin API Surface

```javascript
// Plugin structure
module.exports = {
    install(less, pluginManager, functionRegistry) { },
    use(plugin) { },
    setOptions(options) { },
    minVersion: [3, 0, 0]
};

// Capabilities
functionRegistry.add('myFunc', fn);          // Custom functions
pluginManager.addVisitor(visitor);           // AST transformation
pluginManager.addPreProcessor(fn, priority); // Source transformation
pluginManager.addPostProcessor(fn, priority);// CSS transformation
pluginManager.addFileManager(fm);            // Custom imports

// Node constructors
less.dimension(1, 'px');
less.color([255, 0, 0]);
less.quoted('"', 'value');
// ... all node types
```

## 🗺️ Implementation Roadmap

### Phase Dependencies

```
Phase 1 (Runtime)
  ↓
Phase 2 (Serialization) ← Can develop in parallel with Phase 4
  ↓
Phase 3 (Bindings)
  ↓
Phase 4 (Loader) ← Can start after Phase 1
  ↓
Phase 5 (Functions) ←┐
Phase 6 (Visitors)   ├── Can be done in parallel
Phase 7 (Constructors)←┘
  ↓
Phase 8 (Processors) ←┐
Phase 9 (File Managers)├── Can be done in parallel
Phase 10 (Scoping)    ←┘
  ↓
Integration Testing
  ↓
Performance Optimization
```

### Suggested Agent Assignment

- **Agent 1**: Phase 1 + Phase 2 (Runtime + Serialization) - Critical path
- **Agent 2**: Phase 4 (Plugin Loader) - Can start early, parallel to Agent 1
- **Agent 3**: Phase 3 + Phase 7 (Bindings + Constructors) - Needs Phase 2
- **Agent 4**: Phase 5 (Functions) - Needs Phase 3
- **Agent 5**: Phase 6 (Visitors) - Needs Phase 3
- **Agent 6**: Phase 8 + Phase 9 (Processors + File Managers) - Can be parallel
- **Agent 7**: Phase 10 + Integration Testing - Needs most phases complete

### Timeline Estimate

With parallel agents: **6-8 weeks**
Sequential: **12-16 weeks**

## ✅ Success Criteria

### Minimum Viable Product (MVP)

- ✅ All 8 quarantined plugin tests pass
- ✅ Perfect CSS output match with less.js
- ✅ No regressions in existing 183 tests
- ✅ Basic plugin scoping (global, local)
- ✅ All plugin capabilities work:
  - Custom functions
  - Pre/post-eval visitors
  - Pre/post-processors
  - File managers

### Stretch Goals

- 🎯 Performance overhead < 3x vs. native Go implementation
- 🎯 Support for real-world plugins (bootstrap4 test)
- 🎯 Migration to v8go for better performance
- 🎯 Plugin development documentation and examples

## 🔬 Testing Strategy

### Unit Tests (per phase)

```bash
# Test specific package
go test ./packages/less/src/less/less_go/runtime

# With coverage
go test -cover ./packages/less/src/less/less_go/runtime

# Verbose
go test -v ./packages/less/src/less/less_go/runtime
```

### Integration Tests

```bash
# All tests (including non-plugin)
pnpm -w test:go

# Unit tests only
pnpm -w test:go:unit

# Specific test
go test -v -run TestIntegrationSuite/plugin-simple

# With debug output
LESS_GO_DEBUG=1 pnpm -w test:go

# With CSS diff
LESS_GO_DIFF=1 pnpm -w test:go
```

### Benchmarks

```bash
# Run benchmarks
go test -bench=. -benchmem ./packages/less/src/less/less_go/runtime

# Profile
go test -bench=. -cpuprofile=cpu.prof ./runtime
go tool pprof cpu.prof
```

## 📖 References

### External Resources

1. [OXC Plugin Proposal](https://github.com/oxc-project/oxc/issues/2409#issue-2133367176) - Core inspiration
2. [Speeding Up JS Ecosystem Part 11](https://marvinh.dev/blog/speeding-up-javascript-ecosystem-part-11/) - Flattened AST approach
3. [Oxlint JS Plugins - Raw Transfer](https://oxc.rs/blog/2025-10-09-oxlint-js-plugins.html#raw-transfer) - Implementation details
4. [less.js Plugin Documentation](https://lesscss.org/features/#plugin-atrules-feature) - Official plugin API
5. [goja](https://github.com/dop251/goja) - Pure Go JavaScript runtime
6. [v8go](https://github.com/rogchap/v8go) - V8 bindings for Go (performance upgrade)

### Internal Resources

- **JavaScript Implementation**: `packages/less/src/less/plugin-manager.js`
- **Plugin Loader**: `packages/less/src/less-node/plugin-loader.js`
- **Example Plugins**: `packages/test-data/plugin/*.js`
- **Test Files**: `packages/test-data/less/_main/plugin*.less`
- **Go Codebase**: `packages/less/src/less/less_go/`

## 🤝 Contributing

### For New Agents

1. Read **IMPLEMENTATION_STRATEGY.md** (30 min) - Understand the architecture
2. Read **TASK_BREAKDOWN.md** (20 min) - Find your task
3. Read **QUICKSTART.md** (5 min) - Learn the workflow
4. Claim a task (update TASK_BREAKDOWN.md)
5. Implement following TDD
6. Submit and update status

### Coordination

- **Update TASK_BREAKDOWN.md** when starting/finishing tasks
- **Don't overlap** - Coordinate on shared code areas
- **Test frequently** - Ensure no regressions
- **Document gotchas** - Help future agents
- **Review code** - Quality over speed

## 🐛 Common Issues

### "Tests are slow"

- Benchmark and profile first
- Optimize hot paths only
- Consider v8go migration if needed

### "JavaScript error unclear"

- Add console.log to plugin (temporarily)
- Check JavaScript implementation behavior
- Verify buffer serialization is correct

### "AST doesn't match"

- Use LESS_GO_DIFF=1 to see differences
- Compare with JavaScript AST
- Check node constructors

### "Import resolution fails"

- Verify pnpm install was run
- Check node_modules structure
- Test with absolute path first

## 📊 Progress Tracking

Update this section as phases complete:

- [ ] Phase 1: JavaScript Runtime Integration
- [ ] Phase 2: AST Serialization
- [ ] Phase 3: JavaScript Bindings
- [ ] Phase 4: Plugin Loader
- [ ] Phase 5: Function Registry Integration
- [ ] Phase 6: Visitor Integration
- [ ] Phase 7: Tree Node Constructors
- [ ] Phase 8: Pre/Post Processors
- [ ] Phase 9: File Manager Support
- [ ] Phase 10: Plugin Scope Management
- [ ] Integration Testing: plugin-simple
- [ ] Integration Testing: plugin-tree-nodes
- [ ] Integration Testing: plugin-preeval
- [ ] Integration Testing: plugin
- [ ] Integration Testing: plugin-module
- [ ] Integration Testing: bootstrap4 (stretch)
- [ ] Performance Optimization
- [ ] Documentation

## 📝 Notes

### Design Decisions

Document key decisions here as they're made:

- **Runtime choice**: [goja | v8go | other] - Rationale: ...
- **Buffer format version**: v1 - Rationale: ...
- **Visitor timing**: [pre-eval | post-eval | both] - Rationale: ...

### Performance Metrics

Track performance as you go:

- **Serialization**: ? ms per 1000 nodes (target: < 10ms)
- **Deserialization**: ? ms per 1000 nodes (target: < 5ms)
- **Function call overhead**: ? μs per call (target: < 100μs)
- **End-to-end**: ?x slower than native (target: < 3x)

### Lessons Learned

Document challenges and solutions:

- **Lesson 1**: ...
- **Lesson 2**: ...

---

**Last Updated**: 2025-11-28
**Status**: Planning Complete, Ready for Implementation
**Next Step**: Agent 1 should start Phase 1 (Runtime Integration)
