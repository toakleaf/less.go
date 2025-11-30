# Agent 3: Integration & Testing

**Status**: ⏸️ Blocked - Waiting for Agents 1 and 2
**Dependencies**: Agent 1 (JS), Agent 2 (Go)
**Estimated Time**: 2-3 hours

---

## Your Mission

Un-quarantine the JavaScript tests, verify everything works, fix any issues, and ensure no regressions.

## Prerequisites

Before starting, verify Agents 1 and 2 have completed:

1. **Agent 1 Complete**: `plugin-host.js` has `evalJS` command handler
2. **Agent 2 Complete**: `js_eval_node.go` calls Node.js runtime

Quick verification:
```bash
# Check plugin-host.js has evalJS
grep -n "case 'evalJS'" packages/less/src/less/less_go/runtime/plugin-host.js

# Check js_eval_node.go has runtime call
grep -n "evalJS" packages/less/src/less/less_go/js_eval_node.go

# Ensure code compiles
go build ./packages/less/src/less/less_go/...
```

## Required Reading

1. `.claude/tasks/inline-js/README.md` - Overview (~5 min)
2. `.claude/tasks/inline-js/TASK_BREAKDOWN.md` - Your tasks are 3.1-3.7 (~10 min)
3. `packages/test-data/less/_main/javascript.less` - Test input (~5 min)
4. `packages/test-data/css/_main/javascript.css` - Expected output (~5 min)

---

## Your Tasks

### Task 3.1: Un-quarantine JavaScript Test

**File**: `packages/less/src/less/less_go/integration_suite_test.go`

Find the quarantine configuration (around line 295-317) and remove `javascript`:

```go
// Find this section and remove "javascript" from quarantine
// Before:
quarantinedTests := map[string]bool{
    "javascript":     true,  // <-- REMOVE THIS LINE
    "plugin":         true,
    "plugin-module":  true,
    "plugin-preeval": true,
    "bootstrap4":     true,
}

// After:
quarantinedTests := map[string]bool{
    "plugin":         true,
    "plugin-module":  true,
    "plugin-preeval": true,
    "bootstrap4":     true,
}
```

**Run initial test**:
```bash
LESS_GO_DEBUG=1 go test -v -run "TestIntegrationSuite/_main/javascript" ./packages/less/src/less/less_go
```

---

### Task 3.2: Debug JavaScript Test Issues

If the test fails, use these debugging techniques:

**See the CSS diff**:
```bash
LESS_GO_DIFF=1 go test -v -run "TestIntegrationSuite/_main/javascript" ./packages/less/src/less/less_go
```

**Enable tracing**:
```bash
LESS_GO_TRACE=1 go test -v -run "TestIntegrationSuite/_main/javascript" ./packages/less/src/less/less_go
```

**Common issues to check**:

1. **Variable interpolation**: `@{world}` should be replaced with variable value
2. **this.foo.toJS()**: Should return the CSS string of the variable
3. **Array joining**: `[1, 2, 3].join(', ')` should produce `"1, 2, 3"`
4. **Escaped output**: `~` prefix should produce unquoted output
5. **process.title**: Should be a string in Node.js (type check)

**Expected test file mapping**:
| Less Input | Expected CSS Output |
|------------|---------------------|
| `` `42` `` | `42` |
| `` `1 + 1` `` | `2` |
| `` `"hello world"` `` | `"hello world"` |
| `` `[1, 2, 3]` `` | `1, 2, 3` |
| `` `typeof process.title` `` | `"string"` |
| `` `parseInt(this.foo.toJS())` `` with `@foo: 42` | `42` |
| `` ~`2 + 5 + 'px'` `` | `7px` (unquoted) |

---

### Task 3.3: Un-quarantine Error Tests

**Remove js-type-errors from quarantine**:

Look for wildcard patterns in the quarantine config:
```go
// Find patterns like:
quarantinedErrorSuites := []string{
    "js-type-errors",  // <-- REMOVE
    "no-js-errors",    // <-- May need to check this too
}
```

**Or individual test patterns**:
```go
if strings.HasPrefix(testName, "js-type-errors/") {
    t.Skip("Quarantined: JS type errors")  // <-- Remove this
}
```

**Run error tests**:
```bash
go test -v -run "TestIntegrationSuite/js-type-errors" ./packages/less/src/less/less_go
```

---

### Task 3.4: Verify Error Messages

The error tests check that JavaScript errors are properly detected and reported.

**Test file**: `packages/test-data/less/js-type-errors/js-type-error.less`
```less
.scope {
  var: `this.foo.toJS`;  // Error: foo is undefined
}
```

**Expected error** (from `js-type-error.txt`):
```
SyntaxError: JavaScript evaluation error: 'TypeError: Cannot read property 'toJS' of undefined' in {path}js-type-error.less on line 2, column 8:
```

**If exact match isn't possible**, you can modify the expected error file to match Go's output. The key requirements are:
1. Error type is detected (TypeError)
2. Relevant error message is included
3. File and line info is present

**To update expected error**:
```bash
# Run test and capture actual error
LESS_GO_DEBUG=1 go test -v -run "TestIntegrationSuite/js-type-errors/js-type-error" ./packages/less/src/less/less_go 2>&1

# If error format differs, update the .txt file:
# packages/test-data/less/js-type-errors/js-type-error.txt
```

---

### Task 3.5: Verify no-js-errors Test

This test should already work since the `javascriptEnabled: false` check was already implemented.

**Run test**:
```bash
go test -v -run "TestIntegrationSuite/no-js-errors" ./packages/less/src/less/less_go
```

**Expected behavior**: Test should error with:
```
SyntaxError: Inline JavaScript is not enabled. Is it set in your options? in {path}no-js-errors.less on line 2, column 6:
```

**If quarantined**, remove from quarantine and verify it passes.

---

### Task 3.6: Verify No Regressions

Run the full test suite to ensure nothing broke:

```bash
# Unit tests - must be 100%
pnpm -w test:go:unit

# Integration tests - should maintain 183/183 or improve
LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | tail -100
```

**Expected results**:
- Unit tests: 3,012+ passing (100%)
- Integration tests: 183/183 or more (100%)
- New JS tests: +1 to +3 more passing tests

**If regressions**:
1. Identify which test regressed
2. Check if it's related to JavaScript changes
3. Debug and fix before proceeding

---

### Task 3.7: Update Documentation

Update `CLAUDE.md` with new status:

**Location**: `/home/user/less.go/CLAUDE.md`

**Find the quarantined tests section** (around line 9 in section 9):
```markdown
9. **Quarantined Features** (for future implementation):
   - Plugin system tests (`plugin`, `plugin-module`, `plugin-preeval`)
   - JavaScript execution tests (`javascript`, `js-type-errors/*`, `no-js-errors/*`)  // <-- Update
   ...
```

**Update to**:
```markdown
9. **Quarantined Features** (for future implementation):
   - Plugin system tests (`plugin`, `plugin-module`, `plugin-preeval`)
   - ~~JavaScript execution tests~~ ✅ Now implemented!
   ...
```

**Update test counts** if they've changed:
- Find "Current Integration Test Status" section
- Update perfect CSS match count
- Update any other relevant numbers

---

## Expected Final State

After completing all tasks:

| Test Suite | Status |
|------------|--------|
| `javascript` | ✅ Passing (perfect CSS match) |
| `js-type-errors/*` | ✅ Passing (correct error detection) |
| `no-js-errors/*` | ✅ Passing (correct "not enabled" error) |
| All unit tests | ✅ No regressions |
| All integration tests | ✅ No regressions |

---

## Troubleshooting Guide

### Issue: "JavaScript runtime not available"

The Node.js runtime isn't being passed to the JavaScript node evaluation.

**Check**:
1. Is `javascriptEnabled: true` in test options?
2. Is `PluginBridge` or `LazyPluginBridge` set on the Eval context?
3. Is the runtime started?

**Solution**: The integration test may need to enable plugins for JavaScript tests. Check how plugin tests are configured.

### Issue: "undefined is not an object" or similar

Variable context isn't being passed correctly.

**Check**:
1. Is `buildVariableContext` returning variables?
2. Are variable names stripped of `@` prefix?
3. Is `toJS()` working in Node.js?

### Issue: Output differs from expected

**Check**:
1. Are numbers coming through as numbers (not strings)?
2. Are strings properly quoted (or unquoted for escaped)?
3. Are arrays joined with correct separator?

### Issue: Tests skip with "quarantined"

**Check**:
1. Did you remove all quarantine entries?
2. Are there multiple quarantine locations?
3. Is there a wildcard pattern matching?

---

## Verification Checklist

Before marking complete:

- [ ] `javascript` test un-quarantined and passing
- [ ] `js-type-errors` tests un-quarantined and passing
- [ ] `no-js-errors` tests un-quarantined and passing
- [ ] Unit tests: 100% passing (no regressions)
- [ ] Integration tests: 183/183+ (no regressions)
- [ ] CLAUDE.md updated with new status
- [ ] All changes committed

---

## Deliverables

When complete, provide:

1. **Summary**: Test results (2-3 sentences)
2. **Test counts**: Before and after (e.g., "183 → 186 perfect matches")
3. **Files modified**: List of changed files
4. **Issues fixed**: Any problems encountered and how resolved
5. **Documentation updates**: What was updated in CLAUDE.md

---

## Commit Message Template

```
feat(inline-js): enable inline JavaScript expression evaluation

- Add evalJS command to plugin-host.js
- Modify js_eval_node.go to use Node.js runtime
- Un-quarantine javascript, js-type-errors, no-js-errors tests
- All inline JS tests now passing

Test results:
- javascript: perfect CSS match
- js-type-errors: correct error detection
- no-js-errors: correct "not enabled" error
- No regressions in existing tests
```

Good luck! You're bringing the final JavaScript features online. 🚀
