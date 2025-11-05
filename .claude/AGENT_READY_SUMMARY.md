# Agent Task Distribution - Ready for Assignment

**Date**: 2025-11-05
**Session**: claude/complete-tasks-review-011CUqB6UCK6xP5sMjcnxdpR

---

## 📊 Current Status

Ran full integration test suite and audited all existing agent tasks against current test results.

### Test Results:
- ✅ **Perfect Matches**: 14 tests
- ❌ **Compilation Failures**: 11 unique tests (12 instances)
- ⚠️ **Output Differences**: ~100+ tests
- 🚨 **Critical**: 5 regressions detected

### Task Review Results:
**ALL 5 existing agent tasks are STILL FAILING** - none have been completed yet:
- ❌ agent-namespacing-6 (now part of regression task)
- ❌ agent-import-reference (now part of regression task)
- ❌ agent-import-reference-issues (now part of regression task)
- ❌ agent-urls (still valid, can be used as-is)
- ❌ agent-paths (still valid, can be used as-is)

---

## 📦 What I've Created

### New Task Specifications (7 tasks)

#### Phase 1: CRITICAL Regressions (Fix First!)

1. **`.claude/tasks/regressions/namespacing-variable-calls.md`**
   - Tests: namespacing-6, namespacing-functions
   - Impact: Restore 1-2 perfect matches
   - Issue: "Could not evaluate variable call @alias"

2. **`.claude/tasks/regressions/extend-clearfix.md`**
   - Tests: extend-clearfix
   - Impact: Restore 1 perfect match
   - Issue: `:extend(.clearfix all)` not extending nested selectors

3. **`.claude/tasks/regressions/import-reference.md`**
   - Tests: import-reference, import-reference-issues
   - Impact: Restore compilation for 2 tests
   - Issue: CSS files loaded as LESS, referenced mixins not accessible

#### Phase 2: High/Medium Priority Compilation Failures

4. **`.claude/tasks/runtime-failures/mixins-args.md`**
   - Tests: mixins-args (2 instances)
   - Issue: Mixins can't be called before definition (forward reference problem)

5. **`.claude/tasks/runtime-failures/import-interpolation.md`**
   - Tests: import-interpolation
   - Issue: Import paths with `@{variable}` not evaluated

6. **`.claude/tasks/runtime-failures/import-module.md`**
   - Tests: import-module
   - Issue: Module-style imports (`@less/module-name`) not supported

7. **Referenced Existing Tasks**:
   - `.claude/agents/agent-urls/TASK.md` - Still valid
   - `.claude/agents/agent-paths/TASK.md` - Still valid

### Organization Documents

- **`.claude/WORK_QUEUE.md`** - Master work queue with priorities and assignment strategy
- **`.claude/tracking/assignments.json`** - Structured task tracking (JSON format)
- **`.claude/tracking/TEST_AUDIT_2025-11-05.md`** - Already existed, validated it's current

---

## 🎯 How to Use This

### Option 1: Assign to Multiple Agents in Parallel

**Phase 1 (CRITICAL - Do First):**
```bash
# Agent 1
cat .claude/tasks/regressions/namespacing-variable-calls.md
# Give this to agent 1 with instructions to fix the regression

# Agent 2
cat .claude/tasks/regressions/extend-clearfix.md
# Give this to agent 2 with instructions to fix the regression

# Agent 3
cat .claude/tasks/regressions/import-reference.md
# Give this to agent 3 with instructions to fix the regression
```

These can work in parallel - they touch different files.

**Phase 2 (After Phase 1 Complete):**
```bash
# Agent 4
cat .claude/tasks/runtime-failures/mixins-args.md

# Agent 5
cat .claude/agents/agent-urls/TASK.md

# Agent 6
cat .claude/agents/agent-paths/TASK.md

# Etc.
```

### Option 2: Use WORK_QUEUE as a Menu

```bash
# Show the work queue
cat .claude/WORK_QUEUE.md

# Pick any available Phase 1 task and assign to an agent
# After Phase 1 complete, pick from Phase 2 tasks
```

### Option 3: Automated Assignment

Use the JSON tracking file:

```bash
cat .claude/tracking/assignments.json

# Parse JSON to find available tasks
# Filter by priority (CRITICAL first)
# Assign to agents programmatically
```

---

## ✅ Task Specifications Include

Each task specification contains:

1. **Overview** - What's broken and why it matters
2. **Current vs Expected Behavior** - Clear examples
3. **Investigation Starting Points** - Files to examine, debug commands
4. **Root Cause Hypothesis** - Where the problem likely is
5. **Success Criteria** - How to know when it's fixed
6. **Validation Checklist** - Required test commands before PR
7. **Additional Context** - Related tests, JavaScript comparison, notes

---

## 📋 Validation Requirements

Every agent MUST:

1. ✅ Fix their specific test(s)
2. ✅ Run ALL unit tests: `pnpm -w test:go:unit`
3. ✅ Run FULL integration suite: `pnpm -w test:go`
4. ✅ Verify no new failures (zero regression tolerance)
5. ✅ Document test results in PR

See `.claude/VALIDATION_REQUIREMENTS.md` for full details.

---

## 🎯 Expected Outcomes

### After Phase 1 (Regressions Fixed):
- ✅ Perfect matches: 16-17 (up from 14)
- ❌ Compilation failures: 9 (down from 11)
- ✅ Regressions: 0 (goal!)
- 📈 Restored to stable state

### After Phase 2 (Compilation Failures Fixed):
- ✅ Perfect matches: 17-18
- ❌ Compilation failures: 5-6 (down from 11)
- ✅ All high-priority tests compile
- 📈 Ready for output difference work

### Phase 3 (Future):
- Need to analyze ~100+ tests with output differences
- Create more task specifications
- Continue incremental improvements

---

## 📁 File Structure

```
.claude/
├── WORK_QUEUE.md                        ← Start here: Master work queue
├── AGENT_READY_SUMMARY.md              ← This file: What's ready
├── VALIDATION_REQUIREMENTS.md           ← Required testing
├── README.md                            ← Project overview
│
├── tasks/
│   ├── regressions/                     ← CRITICAL priority
│   │   ├── namespacing-variable-calls.md
│   │   ├── extend-clearfix.md
│   │   └── import-reference.md
│   │
│   └── runtime-failures/                ← HIGH/MEDIUM priority
│       ├── mixins-args.md
│       ├── import-interpolation.md
│       └── import-module.md
│
├── agents/                              ← OLD structure, some still valid
│   ├── agent-urls/TASK.md              ← STILL VALID - use this
│   └── agent-paths/TASK.md             ← STILL VALID - use this
│
└── tracking/
    ├── assignments.json                 ← Task tracking (JSON)
    └── TEST_AUDIT_2025-11-05.md        ← Latest audit
```

---

## 🚀 Quick Start Commands

### Check Current Status
```bash
# Run full test suite
pnpm -w test:go

# Get summary
pnpm -w test:go:summary

# Count perfect matches
pnpm -w test:go 2>&1 | grep "✅.*Perfect match" | wc -l
```

### Assign First Task
```bash
# Read the highest priority task
cat .claude/tasks/regressions/namespacing-variable-calls.md

# Copy contents and give to an agent with:
# "Fix this regression following the task specification"
```

### Track Progress
```bash
# View work queue
cat .claude/WORK_QUEUE.md

# View assignments
cat .claude/tracking/assignments.json

# Update when task claimed/completed
# Edit assignments.json to mark status
```

---

## ⚠️ Critical Notes

1. **Regressions MUST be fixed first** - These broke previously working tests
2. **Zero regression tolerance** - Any fix that breaks a passing test must be fixed
3. **All agents must run full test suite** - Both unit and integration
4. **Document test results** - Include counts in PRs
5. **Import tasks may conflict** - Be careful with parallel import work

---

## 📞 Need More Tasks?

Once these 8 tasks are complete (or in progress), I can:
1. Analyze the ~100+ tests with output differences
2. Create more task specifications
3. Prioritize based on impact and dependencies
4. Organize into more phases

For now, these 8 tasks should keep multiple agents busy and will significantly improve the test pass rate.

---

## 🎓 Related Documentation

- **Main Project Context**: `/home/user/less.go/CLAUDE.md`
- **Agent Workflow Guide**: `.claude/strategy/agent-workflow.md`
- **Master Plan**: `.claude/strategy/MASTER_PLAN.md`
- **Test Audit**: `.claude/tracking/TEST_AUDIT_2025-11-05.md`

---

**Status**: ✅ Ready for agent assignment
**Next Step**: Assign Phase 1 regression tasks to agents
**Goal**: Restore codebase to stable state, then systematically fix remaining failures

🚀 Let's get these tests passing!
