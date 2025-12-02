# .claude/ Directory

This directory contains all project coordination and task management documentation for the less.go port.

**Last Updated**: 2025-12-02

## Current Status

**Port Complete!**

- **Perfect CSS Matches**: 100 tests
- **Correctly Failed (Error Tests)**: 91 tests
- **Overall**: 191/191 tests passing (100%)
- **Unit Tests**: 3,012 passing (100%)

## Project Structure

```
less.go/
├── less/               # Go implementation (core library)
├── cmd/lessc-go/       # CLI tool
├── testdata/           # Test fixtures (LESS files, expected CSS)
├── test/js/            # JavaScript unit tests
├── npm/                # NPM package templates (platform-specific)
├── reference/less.js/  # Original Less.js (git submodule, reference only)
├── examples/           # Usage examples
├── scripts/            # Build and test scripts
└── .claude/            # Claude Code configuration and documentation
```

## Directory Structure

```
.claude/
├── README.md                      # This file
├── AGENT_WORK_QUEUE.md            # Current work items
├── QUICK_START_AGENT_GUIDE.md     # Quick onboarding for agents
├── INTEGRATION_TEST_GUIDE.md      # How to use integration tests
├── VALIDATION_REQUIREMENTS.md     # Required validation before PRs
├── strategy/                      # High-level planning
│   ├── MASTER_PLAN.md            # Overall strategy and status
│   └── agent-workflow.md         # Workflow for agents
├── tasks/                         # Task specifications
│   └── archived/                 # Completed task documentation
├── templates/                     # Templates for agents
│   └── AGENT_PROMPT.md           # Onboarding prompt
├── tracking/                      # Progress tracking
│   └── TEST_STATUS_REPORT.md     # Current test status
├── benchmarks/                    # Performance benchmarks
└── archived-reports/              # Historical status reports
```

## Quick Start

### For New AI Agents

1. **Read** `AGENT_WORK_QUEUE.md` - See current work items
2. **Read** `QUICK_START_AGENT_GUIDE.md` - Quick onboarding
3. **Follow** the workflow in `strategy/agent-workflow.md`
4. **Test** thoroughly using commands in the task file
5. **Create PR** when tests pass

## Key Files

| File | Purpose |
|------|---------|
| `AGENT_WORK_QUEUE.md` | Current tasks and priorities |
| `strategy/MASTER_PLAN.md` | Overall strategy |
| `tracking/TEST_STATUS_REPORT.md` | Test metrics |
| `INTEGRATION_TEST_GUIDE.md` | Test usage guide |

## Validation Commands

```bash
# Run unit tests
pnpm test:go:unit          # 3,012 tests passing

# Run integration tests (quick summary)
LESS_GO_QUIET=1 pnpm test:go 2>&1 | tail -30

# Run all tests
pnpm test
```

## Documentation Links

- [README.md](../README.md) - Project overview
- [CONTRIBUTING.md](../CONTRIBUTING.md) - Contribution guidelines
- [CLAUDE.md](../CLAUDE.md) - Claude Code context
- [BENCHMARKS.md](../BENCHMARKS.md) - Performance benchmarks

---

**Maintained By**: Project maintainers and contributing agents
