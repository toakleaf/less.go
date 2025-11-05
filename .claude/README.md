# Claude Agent Orchestration for less.go

This directory contains everything needed to orchestrate multiple AI agents to fix the remaining integration test failures in the less.go project.

---

## ⚠️ **CURRENT STATUS: REGRESSIONS DETECTED**

**As of 2025-11-05**, the codebase has critical regressions that must be fixed **before any new feature work**.

**Current Test Status**:
- **Perfect Matches**: 14 (DOWN from 15)
- **Compilation Failures**: 11 (UP from 6)
- **Output Differences**: ~100+
- **Net Result**: REGRESSION

### Critical Regressions
Recent namespace/variable work broke multiple tests:
- ❌ `namespacing-6`: Perfect match → Compilation failed
- ❌ `extend-clearfix`: Perfect match → Output differs
- ❌ `namespacing-functions`: Worse (now fails to compile)
- ❌ `import-reference`: Worse (now fails to compile)
- ❌ `import-reference-issues`: Worse (now fails to compile)

**See full audit**: `.claude/tracking/TEST_AUDIT_2025-11-05.md`

---

## 🚀 **Quick Start: What To Do Right Now**

### STEP 1: Fix Critical Regressions (URGENT)

**Prompt File**: `.claude/PROMPT_FIX_REGRESSIONS.md`

This must be done **first** before any other work:

```bash
# Open a new Claude Code session and paste contents of:
cat .claude/PROMPT_FIX_REGRESSIONS.md
```

This will guide an agent to:
- Fix the 5 critical test regressions
- Restore codebase to stable state
- Verify no new failures introduced

**Priority**: CRITICAL - Blocks all other work

---

### STEP 2: Create Agent Task Specifications (After Step 1)

**Prompt File**: `.claude/PROMPT_CREATE_AGENT_TASKS.md`

Once regressions are fixed and codebase is stable:

```bash
# Open a new Claude Code session and paste contents of:
cat .claude/PROMPT_CREATE_AGENT_TASKS.md
```

This will guide an agent to:
- Analyze remaining test failures
- Create detailed task specifications
- Organize tasks for parallel agent work
- Update work queue and tracking

---

## 📁 Directory Structure

```
.claude/
├── README.md                          ← You are here
├── README_AGENT_PROMPTS.md           ← Guide to using prompts
│
├── PROMPT_FIX_REGRESSIONS.md         ← START HERE: Fix regressions
├── PROMPT_CREATE_AGENT_TASKS.md      ← Then: Create task specs
│
├── AGENT_WORK_QUEUE.md               ← Summary of available work
├── VALIDATION_REQUIREMENTS.md        ← Test requirements for PRs
│
├── strategy/
│   ├── agent-workflow.md             ← Step-by-step agent workflow
│   └── MASTER_PLAN.md                ← Overall project strategy
│
├── tasks/
│   ├── runtime-failures/             ← High-priority failing tests
│   │   ├── mixin-args.md
│   │   ├── include-path.md
│   │   └── ... (more to be created)
│   └── output-differences/           ← Medium-priority output issues
│       ├── namespacing-output.md
│       ├── guards-conditionals.md
│       └── ... (more to be created)
│
├── tracking/
│   ├── assignments.json              ← Task status tracking
│   ├── TEST_AUDIT_2025-11-05.md     ← Latest test status audit
│   └── TEST_STATUS_REPORT.md        ← Detailed analysis
│
└── [OLD STRUCTURE - Can be archived]:
    ├── agents/                        ← Old agent task structure
    ├── reference-issues/              ← Old issue analysis
    └── LAUNCH-WEB-AGENTS.md          ← Old web agent approach
```

---

## 📊 Current Test Status (Verified 2025-11-05)

### Perfect Matches (14 tests)
Tests that produce exactly correct CSS output:
- charsets (NEW!)
- css-grid
- empty
- ie-filters
- impor
- lazy-eval
- mixin-noparens
- mixins
- mixins-closure
- mixins-interpolated
- mixins-pattern
- no-output
- plugi
- rulesets

### Compilation Failures (11 tests)
Tests that don't even compile:
1. **namespacing-6** ❌ REGRESSION (was working!)
2. **namespacing-functions** ❌ REGRESSION (worse)
3. **import-reference** ❌ REGRESSION (worse)
4. **import-reference-issues** ❌ REGRESSION (worse)
5. mixins-args (appears in 3 suites)
6. include-path
7. import-interpolation
8. import-module
9. bootstrap4
10. google (network issue)
11. urls

### Output Differences (~100+ tests)
Tests that compile but produce wrong CSS.

---

## 🎯 Work Phases

### Phase 1: CRITICAL - Fix Regressions ⚠️
**Status**: Must do NOW
**Prompt**: `.claude/PROMPT_FIX_REGRESSIONS.md`
**Impact**: Restores 1-2 perfect matches, fixes 5 tests
**Time**: 2-4 hours

### Phase 2: Create Task Specifications
**Status**: After Phase 1 complete
**Prompt**: `.claude/PROMPT_CREATE_AGENT_TASKS.md`
**Impact**: Enables parallel agent work
**Time**: 3-4 hours

### Phase 3: Parallel Agent Work
**Status**: After Phase 2 complete
**Approach**: Multiple agents work on independent tasks
**Reference**: `.claude/AGENT_WORK_QUEUE.md` (will be updated in Phase 2)
**Impact**: Fix remaining failures in parallel

---

## 📚 Documentation Guide

### For Understanding Current State
- `.claude/tracking/TEST_AUDIT_2025-11-05.md` - Latest test audit
- `.claude/tracking/TEST_STATUS_REPORT.md` - Detailed analysis
- `CLAUDE.md` - Project overview and context

### For Working on Tasks
- `.claude/README_AGENT_PROMPTS.md` - Guide to prompts
- `.claude/strategy/agent-workflow.md` - Step-by-step workflow
- `.claude/VALIDATION_REQUIREMENTS.md` - Test validation rules
- `.claude/tasks/*/` - Individual task specifications

### For Tracking Progress
- `.claude/tracking/assignments.json` - Task assignments
- `.claude/AGENT_WORK_QUEUE.md` - Available work summary

---

## 🔍 Useful Commands

### Check Current Test Status
```bash
# Run full integration test suite
pnpm -w test:go

# Get summary
pnpm -w test:go:summary

# Count perfect matches
pnpm -w test:go 2>&1 | grep "✅.*Perfect match" | wc -l

# Count compilation failures
pnpm -w test:go 2>&1 | grep "❌.*Compilation failed" | wc -l

# List all failures
pnpm -w test:go 2>&1 | grep "❌" | sed 's/.*❌ //' | sed 's/: Compilation failed.*//' | sort | uniq
```

### Run Specific Test
```bash
# Run a specific test
pnpm -w test:go:filter -- "suite/test-name"

# Example
pnpm -w test:go:filter -- "namespacing/namespacing-6"
pnpm -w test:go:filter -- "main/extend-clearfix"
```

### Debug Tests
```bash
# Trace execution
LESS_GO_TRACE=1 pnpm -w test:go:filter -- "test-name"

# Show differences
LESS_GO_DIFF=1 pnpm -w test:go:filter -- "test-name"

# Full debug
LESS_GO_DEBUG=1 LESS_GO_TRACE=1 LESS_GO_DIFF=1 pnpm -w test:go:filter -- "test-name"
```

---

## ⚡ Quick Start Checklist

- [ ] Read `.claude/tracking/TEST_AUDIT_2025-11-05.md` to understand regressions
- [ ] Run test suite yourself to verify current state
- [ ] Open new Claude session with `.claude/PROMPT_FIX_REGRESSIONS.md`
- [ ] Wait for regressions to be fixed
- [ ] Open new Claude session with `.claude/PROMPT_CREATE_AGENT_TASKS.md`
- [ ] Review created task specifications
- [ ] Spin up agents for parallel work

---

## 🎓 Learning from Past Work

### What Worked Well ✅
- Fixing one focused issue at a time
- Comprehensive debugging with trace mode
- Comparing with JavaScript implementation
- Thorough validation before merging

### What Went Wrong ❌
- Making changes without running full test suite
- Not detecting regressions before merge
- Breaking previously working tests while fixing others
- Not having baseline requirements

### Going Forward 📈
- **Always run full test suite** before creating PR
- **Zero regressions tolerance** - must fix before merge
- **Require validation proof** in all PRs
- **Document baseline** test numbers clearly

---

## 🔗 Related Files

- `CLAUDE.md` - Main project context (at repository root)
- `AGENT_ORCHESTRATION_STRATEGY.md` - Old orchestration strategy (root)
- `RUNTIME_ISSUES.md` - Old runtime tracking (root)
- `NEXT_SESSION.md` - Old session prompt (root)

**Note**: Root-level files may be outdated. Trust `.claude/` directory for current info.

---

## 📞 Need Help?

1. **Check documentation**: Start with `.claude/README_AGENT_PROMPTS.md`
2. **Review audit report**: `.claude/tracking/TEST_AUDIT_2025-11-05.md`
3. **Read task specs**: `.claude/tasks/` for detailed guidance
4. **Follow workflow**: `.claude/strategy/agent-workflow.md`

---

**Current Priority**: Fix regressions using `.claude/PROMPT_FIX_REGRESSIONS.md` 🚨
