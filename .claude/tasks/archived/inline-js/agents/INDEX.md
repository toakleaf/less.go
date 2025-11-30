# Inline JavaScript - Agent Index

## Overview

This implementation uses 3 agents:
- **Agents 1 & 2**: Can run in parallel
- **Agent 3**: Runs after 1 & 2 complete

## Agent Assignments

| Agent | Focus | Status | Dependencies |
|-------|-------|--------|--------------|
| Agent 1 | JavaScript Side (plugin-host.js) | 🔴 Ready | None |
| Agent 2 | Go Side (js_eval_node.go) | 🔴 Ready | None |
| Agent 3 | Integration & Testing | ⏸️ Blocked | Agents 1 & 2 |

## Execution Flow

```
┌─────────────────┐     ┌─────────────────┐
│    Agent 1      │     │    Agent 2      │
│   (JS Side)     │     │   (Go Side)     │
│   ~2-3 hours    │     │   ~2-3 hours    │
└────────┬────────┘     └────────┬────────┘
         │                       │
         │   Can run parallel    │
         │                       │
         └───────────┬───────────┘
                     │
                     ▼
         ┌─────────────────────┐
         │      Agent 3        │
         │    (Integration)    │
         │     ~2-3 hours      │
         └─────────────────────┘
```

## Quick Start

### For Agent 1
```bash
# Read the prompt
cat .claude/tasks/inline-js/agents/AGENT_1_PROMPT.md

# Start working
code packages/less/src/less/less_go/runtime/plugin-host.js
```

### For Agent 2
```bash
# Read the prompt
cat .claude/tasks/inline-js/agents/AGENT_2_PROMPT.md

# Start working
code packages/less/src/less/less_go/js_eval_node.go
```

### For Agent 3
```bash
# Verify prerequisites
grep -n "case 'evalJS'" packages/less/src/less/less_go/runtime/plugin-host.js
grep -n "evalJS" packages/less/src/less/less_go/js_eval_node.go

# Read the prompt
cat .claude/tasks/inline-js/agents/AGENT_3_PROMPT.md

# Start testing
go test -v -run "TestIntegrationSuite/_main/javascript" ./packages/less/src/less/less_go
```

## Status Tracking

Update this file as agents complete:

- [ ] Agent 1: JavaScript side complete
- [ ] Agent 2: Go side complete
- [ ] Agent 3: Integration testing complete
- [ ] All tests passing
- [ ] Documentation updated
- [ ] Changes committed and pushed
