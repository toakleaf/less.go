# Less.go Work Queue - Ready for Agent Assignment

**Last Updated**: 2025-11-05
**Current Status**: 14 perfect matches, 11 unique compilation failures
**Critical**: 5 regressions must be fixed first

---

## 🚨 CRITICAL PRIORITY - Regressions (Fix First!)

These tests were working or better before recent changes. **Fix these before any other work.**

### 1. Namespacing Variable Calls (2 tests)
- **Task Spec**: `.claude/tasks/regressions/namespacing-variable-calls.md`
- **Tests**: `namespacing-6`, `namespacing-functions`
- **Impact**: Restore 1-2 perfect matches
- **Complexity**: High
- **Time**: 3-4 hours
- **Summary**: Variable calls to mixin results failing with "Could not evaluate variable call"
- **Old Task**: `.claude/agents/agent-namespacing-6/TASK.md` (outdated, use new spec)

### 2. Extend Clearfix (1 test)
- **Task Spec**: `.claude/tasks/regressions/extend-clearfix.md`
- **Tests**: `extend-clearfix`
- **Impact**: Restore 1 perfect match
- **Complexity**: Medium
- **Time**: 2-3 hours
- **Summary**: `:extend(.clearfix all)` not extending nested selectors like `:after`

### 3. Import Reference (2 tests)
- **Task Spec**: `.claude/tasks/regressions/import-reference.md`
- **Tests**: `import-reference`, `import-reference-issues`
- **Impact**: Restore compilation for 2 tests
- **Complexity**: High
- **Time**: 4-6 hours
- **Summary**: Reference imports broken - CSS files loaded as LESS, mixins not accessible
- **Old Tasks**:
  - `.claude/agents/agent-import-reference/TASK.md` (outdated)
  - `.claude/agents/agent-import-reference-issues/TASK.md` (outdated)

**Total Regressions**: 5 tests (2 will restore perfect matches, 2 will restore compilation)

---

## 🔥 HIGH PRIORITY - Compilation Failures

These tests don't compile. Fix after regressions are resolved.

### 4. Mixin Args - Forward References (2 test instances)
- **Task Spec**: `.claude/tasks/runtime-failures/mixins-args.md`
- **Tests**: `mixins-args` (appears in 2 suites: strict & parens-division)
- **Impact**: Fix 2 compilation failures
- **Complexity**: Medium-High
- **Time**: 3-4 hours
- **Summary**: Mixins can't be called before they're defined in same ruleset
- **Key Issue**: "No matching definition was found for `.m3()`" - forward reference problem

### 5. URLs - Escaped Characters (1 test)
- **Task Spec**: `.claude/agents/agent-urls/TASK.md` (existing, still valid)
- **Tests**: `urls` (appears in 2 suites but same issue)
- **Impact**: Fix 1-2 compilation failures
- **Complexity**: Medium
- **Time**: 2-3 hours
- **Summary**: URL parsing fails on escaped characters like `\(` and `\)`
- **Key Issue**: "expected ')' got '('"

### 6. Include Path (1 test)
- **Task Spec**: `.claude/agents/agent-paths/TASK.md` (existing, still valid)
- **Tests**: `include-path`
- **Impact**: Fix 1 compilation failure
- **Complexity**: Medium
- **Time**: 2-3 hours
- **Summary**: Imports don't search configured include paths
- **Key Issue**: "open import-test-e: no such file or directory"

---

## 📦 MEDIUM PRIORITY - Import System Features

These are import-related compilation failures. Can be worked in parallel with each other.

### 7. Import Interpolation (1 test)
- **Task Spec**: `.claude/tasks/runtime-failures/import-interpolation.md`
- **Tests**: `import-interpolation`
- **Impact**: Fix 1 compilation failure
- **Complexity**: Medium
- **Time**: 2-3 hours
- **Summary**: Import paths with `@{variable}` interpolation not evaluated
- **Key Issue**: Tries to load literal "import-@{in}@{terpolation}.less"

### 8. Import Module (1 test)
- **Task Spec**: `.claude/tasks/runtime-failures/import-module.md`
- **Tests**: `import-module`
- **Impact**: Fix 1 compilation failure
- **Complexity**: Medium-High
- **Time**: 3-4 hours
- **Summary**: Module-style imports (`@less/module-name`) not supported
- **Key Issue**: "open @less/test-import-module/one/1.less: no such file"

---

## 📊 Current Test Metrics

### Before Any Fixes:
- ✅ **Perfect Matches**: 14
- ❌ **Compilation Failures**: 11 unique tests (12 instances)
- ⚠️ **Output Differences**: ~100+ tests
- ❌ **Regressions**: 5 tests broken

### After Regression Fixes (Target):
- ✅ **Perfect Matches**: 16-17 (restore 2-3 tests)
- ❌ **Compilation Failures**: 9 unique tests
- ⚠️ **Output Differences**: ~100+ tests
- ✅ **Regressions**: 0 (goal!)

### After All High Priority Fixes (Target):
- ✅ **Perfect Matches**: 17-18
- ❌ **Compilation Failures**: 5-6 unique tests
- ⚠️ **Output Differences**: ~100+ tests

---

## 🎯 Work Assignment Strategy

### Phase 1: Fix Regressions (CRITICAL - Do First)
Assign these in order or in parallel:
1. One agent → Namespacing Variable Calls
2. One agent → Extend Clearfix
3. One agent → Import Reference (both tests)

**Goal**: Restore codebase to pre-regression state
**Timeline**: 1-2 days with 3 parallel agents

### Phase 2: Fix Compilation Failures (After Phase 1)
Can be done in parallel:
1. One agent → Mixin Args Forward References
2. One agent → URLs Escaped Characters
3. One agent → Include Path
4. One agent → Import Interpolation
5. One agent → Import Module

**Goal**: All tests at least compile (may still have output differences)
**Timeline**: 2-3 days with parallel agents

### Phase 3: Output Differences (After Phase 2)
Many tests compile but produce wrong CSS. These need individual analysis and task creation.

---

## 📋 Assignment Checklist

When claiming a task:
- [ ] Read the task specification thoroughly
- [ ] Check for conflicts with other agents (task specs note this)
- [ ] Create branch: `claude/<task-name>-<your-session-id>`
- [ ] Run the specific test(s) to confirm failure
- [ ] Make your fixes
- [ ] **REQUIRED**: Run ALL unit tests: `pnpm -w test:go:unit`
- [ ] **REQUIRED**: Run FULL integration suite: `pnpm -w test:go`
- [ ] Verify no new failures (zero regression tolerance)
- [ ] Create PR with test evidence
- [ ] See `.claude/VALIDATION_REQUIREMENTS.md` for full checklist

---

## 🔍 Test Commands Reference

```bash
# Run full integration suite
pnpm -w test:go

# Get summary counts
pnpm -w test:go:summary

# Run specific test
pnpm -w test:go:filter -- "suite/test-name"

# Run with debug/trace
LESS_GO_TRACE=1 pnpm -w test:go:filter -- "test-name"
LESS_GO_DIFF=1 pnpm -w test:go:filter -- "test-name"

# Run all unit tests (REQUIRED before PR)
pnpm -w test:go:unit
```

---

## 📚 Documentation Structure

```
.claude/
├── WORK_QUEUE.md                    ← You are here
├── VALIDATION_REQUIREMENTS.md       ← Required testing before PR
├── README_AGENT_PROMPTS.md         ← Guide to documentation
├── README.md                        ← Project overview
│
├── tasks/
│   ├── regressions/                 ← CRITICAL - Fix first
│   │   ├── namespacing-variable-calls.md
│   │   ├── extend-clearfix.md
│   │   └── import-reference.md
│   │
│   └── runtime-failures/            ← HIGH/MEDIUM priority
│       ├── mixins-args.md
│       ├── import-interpolation.md
│       └── import-module.md
│
├── agents/                          ← OLD task structure
│   ├── agent-urls/TASK.md          ← Still valid
│   └── agent-paths/TASK.md         ← Still valid
│
└── tracking/
    ├── TEST_AUDIT_2025-11-05.md    ← Latest audit
    └── TEST_STATUS_REPORT.md        ← Detailed analysis
```

---

## ⚠️ Critical Rules

1. **Fix regressions first** - Do not start Phase 2 work until Phase 1 complete
2. **Zero regression tolerance** - Any fix that breaks a passing test must be fixed before merging
3. **Always run full test suite** - Both unit and integration tests required
4. **Document test results** - Include before/after test counts in PR
5. **One task per agent** - Don't try to fix multiple unrelated issues in one PR

---

## 🤝 Agent Independence

### High Independence (Can work in parallel):
- Namespacing Variable Calls + Extend Clearfix + Import Reference
- Mixin Args + URLs + Include Path
- Import Interpolation + Import Module

### Potential Conflicts:
- Import Reference + Import Interpolation (both touch import_manager.go)
- Import Reference + Import Module (both touch import_manager.go)

**Recommendation**: Do import-related tasks sequentially or coordinate carefully.

---

## 📈 Progress Tracking

Update `.claude/tracking/assignments.json` when claiming or completing a task:

```json
{
  "namespacing-variable-calls": {
    "status": "in_progress",
    "agent": "session-id",
    "branch": "claude/fix-namespacing-variable-calls-session-id",
    "started": "2025-11-05"
  }
}
```

---

## 🎓 Learning Resources

- **Project Context**: `/home/user/less.go/CLAUDE.md`
- **Agent Workflow**: `.claude/strategy/agent-workflow.md`
- **Validation Requirements**: `.claude/VALIDATION_REQUIREMENTS.md`
- **Latest Audit**: `.claude/tracking/TEST_AUDIT_2025-11-05.md`
- **JavaScript Reference**: `packages/less/src/less-tree/` for comparing implementations

---

**Ready to start?** Pick a task from Phase 1 (regressions) and dive in! 🚀
