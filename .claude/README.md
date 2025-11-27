# .claude/ Directory

This directory contains all project coordination and task management documentation for the less.go port.

**Last Updated**: 2025-11-27

## Current Status

- **Perfect CSS Matches**: 89 tests (48.4%)
- **Output Differences**: 3 tests (1.6%)
- **Overall Success Rate**: 96.7%
- **Unit Tests**: 3,012 passing (100%)

Only 3 output differences remain: `import-reference`, `import-reference-issues`, `urls`

## Directory Structure

```
.claude/
├── README.md                      # This file
├── AGENT_WORK_QUEUE.md            # Current work items (START HERE)
├── QUICK_START_AGENT_GUIDE.md     # Quick onboarding for agents
├── SESSION_SUMMARY_2025-11-27.md  # Latest session summary
├── INTEGRATION_TEST_GUIDE.md      # How to use integration tests
├── VALIDATION_REQUIREMENTS.md     # Required validation before PRs
├── strategy/                      # High-level planning
│   ├── MASTER_PLAN.md            # Overall strategy and status
│   └── agent-workflow.md         # Workflow for agents
├── tasks/                         # Task specifications
│   ├── runtime-failures/         # Active tasks (3 remaining)
│   │   └── import-reference.md   # Import reference fix
│   ├── error-handling/           # Error test documentation
│   ├── performance/              # Performance analysis
│   └── archived/                 # Completed task documentation
├── templates/                     # Templates for agents
│   └── AGENT_PROMPT.md           # Onboarding prompt
├── tracking/                      # Progress tracking
│   └── TEST_STATUS_REPORT.md     # Current test status
├── benchmarks/                    # Performance benchmarks
├── prompts/                       # Legacy prompts (archived)
└── archived-reports/              # Historical status reports
```

## Quick Start

### For New AI Agents

1. **Read** `AGENT_WORK_QUEUE.md` - See current work items
2. **Read** `QUICK_START_AGENT_GUIDE.md` - Quick onboarding
3. **Pick** a task (only 3 remaining!)
4. **Follow** the workflow in `strategy/agent-workflow.md`
5. **Test** thoroughly using commands in the task file
6. **Create PR** when tests pass

### Remaining Tasks

| Task | Tests | Priority |
|------|-------|----------|
| Import Reference | 2 tests | HIGH |
| URL Handling | 1 test | MEDIUM |

## Key Files

| File | Purpose |
|------|---------|
| `AGENT_WORK_QUEUE.md` | Current tasks and priorities |
| `strategy/MASTER_PLAN.md` | Overall strategy |
| `tracking/TEST_STATUS_REPORT.md` | Test metrics |
| `tasks/runtime-failures/import-reference.md` | Import reference task |

## Validation Commands

```bash
# Check current state
pnpm -w test:go:unit          # 3,012 tests passing
LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | tail -30  # 89 perfect matches
```

---

**Maintained By**: Project maintainers and contributing agents
