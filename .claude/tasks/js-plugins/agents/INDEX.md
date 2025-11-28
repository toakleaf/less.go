# Agent Execution Guide

This directory contains prompts for spawning parallel agents to implement JavaScript plugin support.

## 📋 Agent Overview

| Agent | Task | Time | Dependencies | Can Start When |
|-------|------|------|--------------|----------------|
| **Agent 1** | Node.js Process & Serialization | 1 week | None | ✅ Immediately |
| **Agent 2** | Plugin Loader | 3-4 days | Agent 1 Phase 1 | After basic IPC works |
| **Agent 3** | Bindings & Constructors | 5-7 days | Agent 1 Phase 2 | After serialization complete |
| **Agent 4** | Function Registry | 3-4 days | Agent 3 | After bindings complete |
| **Agent 5** | Visitor Integration | 3-4 days | Agent 3 | After bindings complete (parallel with 4) |
| **Agent 6** | Processors & File Managers | 3-4 days | Agent 2 | After plugin loading (parallel with 4, 5) |
| **Agent 7** | Scoping & Integration | 4-5 days | Agents 2, 4, 5 | Final phase |

**Total Timeline**: 4-6 weeks with parallel execution

## 🚀 Execution Order

### Week 1: Foundation

**Start immediately:**
```bash
# Spawn Agent 1
# Use prompt: AGENT_1_PROMPT.md
```

**After Agent 1 has basic IPC (Phase 1 Tasks 1-3 complete):**
```bash
# Spawn Agent 2 (can work in parallel with Agent 1's Phase 2)
# Use prompt: AGENT_2_PROMPT.md
```

### Week 2: Core Components

**After Agent 1 Phase 2 complete (serialization works):**
```bash
# Spawn Agent 3
# Use prompt: AGENT_3_PROMPT.md
```

### Week 3-4: Plugin Capabilities

**After Agent 3 complete (spawn these in parallel):**
```bash
# Spawn Agent 4
# Use prompt: AGENT_4_PROMPT.md

# Spawn Agent 5 (can work in parallel with Agent 4)
# Use prompt: AGENT_5_PROMPT.md
```

**After Agent 2 complete (can work in parallel with Agents 4 & 5):**
```bash
# Spawn Agent 6
# Use prompt: AGENT_6_PROMPT.md
```

### Week 5-6: Final Integration

**After Agents 2, 4, 5 complete:**
```bash
# Spawn Agent 7 (final phase)
# Use prompt: AGENT_7_PROMPT.md
```

## 📊 Dependency Graph

```
Agent 1 (Node.js + Serialization)
  ├─ Phase 1 (IPC) ──────────┐
  │                          │
  │  ┌───────────────────────┘
  │  ▼
  │  Agent 2 (Plugin Loader) ─────────────┐
  │                                       │
  └─ Phase 2 (Serialization) ────┐       │
                                  │       │
          ┌───────────────────────┘       │
          ▼                               │
     Agent 3 (Bindings) ──────┬───────────┼────────┐
                              │           │        │
          ┌───────────────────┼───────────┘        │
          │                   │                    │
          ▼                   ▼                    ▼
     Agent 4 (Functions)  Agent 5 (Visitors)  Agent 6 (Processors)
          │                   │                    │
          └───────────────────┼────────────────────┘
                              │
                              ▼
                         Agent 7 (Scoping + Tests)
                              │
                              ▼
                         🎉 SUCCESS! 🎉
```

## ✅ Success Criteria Per Agent

### Agent 1
- ✅ Can spawn Node.js process
- ✅ IPC works (stdin/stdout)
- ✅ Shared memory works
- ✅ AST serialization roundtrip works
- ✅ No regressions: `pnpm -w test:go:unit` and `pnpm -w test:go`

### Agent 2
- ✅ Can load plugins via `require()`
- ✅ NPM module resolution works
- ✅ Plugin caching works
- ✅ Can load all test plugins
- ✅ No regressions

### Agent 3
- ✅ NodeFacade reads from shared memory
- ✅ Visitor pattern works
- ✅ All node constructors work
- ✅ Can create nodes from JavaScript
- ✅ No regressions

### Agent 4
- ✅ Can call JS functions from Go
- ✅ Args/results via shared memory
- ✅ Error handling works
- ✅ plugin-simple.js functions work
- ✅ No regressions

### Agent 5
- ✅ Pre-eval visitors work
- ✅ Post-eval visitors work
- ✅ Node replacement works
- ✅ plugin-preeval.js works
- ✅ No regressions

### Agent 6
- ✅ Pre/post processors work
- ✅ File managers work
- ✅ Priority ordering works
- ✅ No regressions

### Agent 7
- ✅ Plugin scoping works
- ✅ Function shadowing works
- ✅ All 5+ plugin integration tests pass
- ✅ No regressions
- ✅ Test count increases from 183 → 188+

## 🎯 Final Success

When all agents complete:

```bash
# Run full test suite
pnpm -w test:go:unit  # Should be 100%
pnpm -w test:go       # Should show 188+/191 passing

# Check summary
LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | tail -100
```

Expected results:
- ✅ Perfect CSS Matches: 99+ (was 94)
- ✅ Overall Success Rate: 100%
- ✅ Compilation Rate: 100%
- ✅ All plugin tests passing
- ✅ Zero regressions

## 📝 Communication Between Agents

Agents should coordinate by:

1. **Updating status** in `TASK_BREAKDOWN.md`
2. **Documenting gotchas** in their deliverable summary
3. **Running tests** before marking complete
4. **Checking for regressions** with every commit

## 🆘 If Something Goes Wrong

If an agent gets stuck:

1. Check the dependencies are actually complete
2. Review the JavaScript implementation for guidance
3. Run tests incrementally (don't wait until the end)
4. Ask for help with specific error messages
5. Consider if the approach needs adjustment

## 📖 Additional Resources

- **Strategy**: `../IMPLEMENTATION_STRATEGY.md`
- **Quick Start**: `../QUICKSTART.md`
- **Changes Log**: `../CHANGES.md`
- **Task Breakdown**: `../TASK_BREAKDOWN.md`

---

**Ready to start?** Begin with Agent 1! 🚀

Copy the prompt from `AGENT_1_PROMPT.md` and spawn your first agent.
