# Agent Task Prompts

This directory contains prompts for spinning up AI agents to work on specific aspects of the less.go project.

## Current Status

⚠️ **IMPORTANT**: As of 2025-11-05, there are critical regressions in the codebase. See the audit report at `.claude/tracking/TEST_AUDIT_2025-11-05.md` for details.

## Available Prompts

### 1. Fix Critical Regressions (URGENT)

**File**: `.claude/PROMPT_FIX_REGRESSIONS.md`

**When to use**: RIGHT NOW - Before any other work

**What it does**: Fixes the 5 critical test regressions that were introduced by recent work:
- namespacing-6 (perfect match → compilation failure)
- extend-clearfix (perfect match → output differs)
- namespacing-functions (worse)
- import-reference (worse)
- import-reference-issues (worse)

**Priority**: CRITICAL - Blocks all other work

**Usage**:
```
Start a new Claude session and paste the contents of:
.claude/PROMPT_FIX_REGRESSIONS.md
```

---

### 2. Create Agent Task Specifications

**File**: `.claude/PROMPT_CREATE_AGENT_TASKS.md`

**When to use**: After regressions are fixed and codebase is stable

**What it does**: Creates detailed task specification documents for the remaining test failures so multiple agents can work in parallel

**Priority**: HIGH - But must wait for regressions to be fixed first

**Usage**:
```
1. First verify no critical regressions exist
2. Run test suite to get current baseline
3. Start a new Claude session and paste the contents of:
   .claude/PROMPT_CREATE_AGENT_TASKS.md
```

---

## Workflow

### Current Situation (2025-11-05)

```
┌─────────────────────────────────┐
│  CRITICAL REGRESSIONS EXIST     │
│  Must fix before new work       │
└─────────────────────────────────┘
              │
              ▼
┌─────────────────────────────────┐
│  Use: PROMPT_FIX_REGRESSIONS    │
│  Fix the 5 broken tests         │
└─────────────────────────────────┘
              │
              ▼
┌─────────────────────────────────┐
│  Verify: Run full test suite   │
│  Confirm regressions fixed      │
└─────────────────────────────────┘
              │
              ▼
┌─────────────────────────────────┐
│  Use: PROMPT_CREATE_AGENT_TASKS │
│  Create task specs for agents   │
└─────────────────────────────────┘
              │
              ▼
┌─────────────────────────────────┐
│  Distribute tasks to agents     │
│  Parallel work can begin        │
└─────────────────────────────────┘
```

### Normal Workflow (After Stable)

1. Agent sees available tasks in `.claude/AGENT_WORK_QUEUE.md`
2. Agent claims a task in `.claude/tracking/assignments.json`
3. Agent follows `.claude/strategy/agent-workflow.md`
4. Agent completes task and updates tracking
5. Agent creates PR with validation proof

## Documentation Structure

```
.claude/
├── README_AGENT_PROMPTS.md          ← You are here
├── PROMPT_FIX_REGRESSIONS.md        ← Regression fix prompt
├── PROMPT_CREATE_AGENT_TASKS.md     ← Task creation prompt
├── AGENT_WORK_QUEUE.md              ← Summary of available work
├── VALIDATION_REQUIREMENTS.md       ← Test requirements for PRs
│
├── strategy/
│   ├── agent-workflow.md            ← Step-by-step workflow
│   └── MASTER_PLAN.md              ← Overall project strategy
│
├── tasks/
│   ├── runtime-failures/           ← High priority tasks
│   │   ├── mixin-args.md
│   │   ├── include-path.md
│   │   └── ...
│   └── output-differences/         ← Medium priority tasks
│       ├── namespacing-output.md
│       ├── guards-conditionals.md
│       └── ...
│
└── tracking/
    ├── assignments.json            ← Task tracking
    ├── TEST_AUDIT_2025-11-05.md   ← Latest audit
    └── TEST_STATUS_REPORT.md      ← Detailed analysis
```

## Quick Reference

### Check Current Test Status
```bash
# Run full test suite
pnpm -w test:go

# Get summary
pnpm -w test:go:summary

# Count perfect matches
pnpm -w test:go 2>&1 | grep "✅.*Perfect match" | wc -l

# Count failures
pnpm -w test:go 2>&1 | grep "❌.*Compilation failed" | wc -l
```

### Before Starting Any Work
```bash
# 1. Check for regressions
cat .claude/tracking/TEST_AUDIT_2025-11-05.md

# 2. Run tests yourself
pnpm -w test:go

# 3. Verify baseline numbers match documentation
```

### Creating a New Task Prompt

If you need to create a new specialized prompt:

1. Copy an existing prompt as template
2. Include clear context about current state
3. Specify exact validation requirements
4. List files to check/modify
5. Provide test commands
6. Define success criteria
7. Reference relevant documentation

## Notes

- **Always verify test status** before starting work
- **Never trust outdated baselines** - run tests yourself
- **Document regressions immediately** if you find any
- **Update tracking files** when status changes
- **Follow validation requirements** for all PRs

## Getting Help

- Review `.claude/strategy/MASTER_PLAN.md` for big picture
- Check `.claude/VALIDATION_REQUIREMENTS.md` for PR requirements
- See `.claude/strategy/agent-workflow.md` for step-by-step process
- Read specific task files in `.claude/tasks/` for detailed guidance
