# Task: Fix Import Reference Regressions

**Status**: Available
**Priority**: CRITICAL
**Estimated Time**: 4-6 hours
**Complexity**: High
**Tests Affected**: 2 (import-reference, import-reference-issues)

## Overview

Two import-related tests have **regressed** from their previous state and now fail to compile. Both involve the `(reference)` import option, which is supposed to:
1. Make mixins/variables available for use
2. NOT output the imported CSS by default
3. Only output CSS when explicitly used (via extend or mixin)

Recent changes appear to have broken reference import handling.

## Failing Tests

### 1. import-reference (REGRESSION)
- **Previous Status**: ⚠️ Output differs (was compiling)
- **Current Status**: ❌ Compilation failed
- **Error**: `Syntax: : open test.css: no such file or directory`
- **Issue**: CSS files are being loaded as LESS files
- **Test Command**:
  ```bash
  pnpm -w test:go:filter -- "main/import-reference"
  ```

### 2. import-reference-issues (REGRESSION - worse)
- **Previous Status**: ⚠️ Output differs (was compiling)
- **Current Status**: ❌ Compilation failed
- **Error**: `Syntax: #Namespace > .mixin is undefined`
- **Issue**: Referenced imports' mixins not accessible
- **Test Command**:
  ```bash
  pnpm -w test:go:filter -- "main/import-reference-issues"
  ```

## Current Behavior

### Test 1: import-reference

**Input LESS**:
```less
@import (reference) url("some-file.less");
@import (reference) url("test.css");  // CSS file!

.b { .z(); }  // Use mixin from referenced LESS import
```

**Current Behavior ❌**:
- Tries to open and load `test.css` as a LESS file
- Fails with "no such file or directory"
- CSS files should NOT be loaded/processed
- CSS `@import` statements should pass through to output unchanged

### Test 2: import-reference-issues

**Input LESS**:
```less
@import (reference) url("file-with-namespace.less");

// file-with-namespace.less contains:
// #Namespace {
//   .mixin() { color: red; }
// }

.test {
  #Namespace > .mixin();  // Should work! ❌ But fails
}
```

**Current Behavior ❌**:
- Referenced imports are processed
- BUT their mixins aren't being added to the evaluation frames
- Mixin lookups fail with "undefined"

## Expected Behavior

### Test 1: import-reference

CSS files should:
1. ✅ Stay as `@import` statements in the CSS output
2. ✅ NOT be loaded or parsed as LESS files
3. ✅ File extension `.css` is the indicator

**Expected**: Import statements referencing `.css` files should be kept as-is, not loaded.

### Test 2: import-reference-issues

Reference imports should:
1. ✅ Make mixins and variables available in evaluation frames
2. ✅ NOT output their CSS by default
3. ✅ Only output CSS when selectors are explicitly used (extended/mixed)

**Expected**: `#Namespace > .mixin()` should be found and callable.

## Investigation Starting Points

### Files to Examine

1. **`import_manager.go`** (PRIMARY) - Lines 100-300
   - File loading and resolution logic
   - CSS file detection (test 1)
   - Check if there's a `shouldProcessAsLESS()` function

2. **`import_visitor.go`** (PRIMARY) - Lines 150-400
   - Import processing and frame management
   - Reference option handling (test 2)
   - How rulesets from imports are added to frames

3. **`import.go`** - Lines 50-150
   - Import node and options
   - Check if `Options.Reference` flag is being set/read correctly

4. **`set_tree_visibility_visitor.go`** - Lines 100-200
   - Visibility control for referenced imports
   - Might be hiding things that should be accessible

### Debug Commands

```bash
# Test 1 - CSS file loading
LESS_GO_DEBUG=1 pnpm -w test:go:filter -- "main/import-reference" 2>&1 | grep -E "(import|css|test\.css)"

# Test 2 - Mixin accessibility
LESS_GO_TRACE=1 pnpm -w test:go:filter -- "main/import-reference-issues" 2>&1 | grep -E "(Namespace|mixin|reference)"

# Look at test input files
cat packages/test-data/less/_main/import-reference.less
cat packages/test-data/less/_main/import-reference-issues.less
```

### Recent Changes

Both tests were compiling before. Check what changed:

```bash
# Check recent changes to import handling
git log -p --since="2025-11-01" -- packages/less/src/less/less_go/import*.go | head -500

# Check recent changes to visibility handling
git log -p --since="2025-11-01" -- packages/less/src/less/less_go/set_tree_visibility_visitor.go
```

## Root Cause Hypothesis

### Test 1 (import-reference)

**Most Likely**: Recent refactoring removed or broke CSS file detection logic.

**Needed Fix**: Add/restore logic to detect CSS files and skip loading them:

```go
// In import_manager.go
func (im *ImportManager) shouldProcessAsLESS(path string) bool {
    // CSS files should NOT be processed as LESS
    if filepath.Ext(path) == ".css" {
        return false
    }
    return true
}
```

Then use this check before attempting to load/parse the file.

### Test 2 (import-reference-issues)

**Most Likely**: Recent changes to import processing broke how referenced imports add their content to frames.

**Possibilities**:
1. Referenced import rulesets aren't being added to frames at all
2. They're being added but with wrong visibility (hidden when should be accessible)
3. The frame push/pop logic changed and broke access

**Needed Fix**: In `import_visitor.go`, ensure referenced imports:
```go
if importNode.Options.Reference {
    // Add rulesets to frames so mixins are accessible
    for _, ruleset := range importedRulesets {
        context.Frames.Push(ruleset)  // Make accessible
        ruleset.SetReferenceOnly(true)  // But don't output by default
    }
}
```

## Success Criteria

- [ ] `import-reference` test compiles successfully
- [ ] `import-reference-issues` test compiles successfully
- [ ] CSS files are NOT loaded as LESS files (test 1)
- [ ] Referenced mixins are accessible (test 2)
- [ ] All unit tests pass: `pnpm -w test:go:unit`
- [ ] FULL integration test suite shows no new regressions: `pnpm -w test:go`

## Validation Checklist

Before creating PR:

```bash
# 1. Verify both specific tests compile
pnpm -w test:go:filter -- "main/import-reference"
# Expected: ✅ Compilation succeeds (output may still differ, that's ok)

pnpm -w test:go:filter -- "main/import-reference-issues"
# Expected: ✅ Compilation succeeds (output may still differ, that's ok)

# 2. Run ALL unit tests (catch any regressions) - REQUIRED
pnpm -w test:go:unit
# Expected: ✅ All unit tests pass (no failures)

# 3. Run FULL integration test suite - REQUIRED
pnpm -w test:go
# Expected:
# - ✅ Compilation failure count drops by 2 (from 11 to 9)
# - ✅ No new compilation failures
# - ✅ Perfect matches remain at 14+
```

**If any test fails that was passing before**: STOP and fix the regression before proceeding.

## Additional Context

### Compare with JavaScript Implementation

The less.js implementation handles this in:
- `packages/less/src/less-node/import-manager.js` - File loading
- `packages/less/src/less-tree/import-visitor.js` - Import processing

Key JavaScript patterns:
```javascript
// CSS file detection
if (importPath.endsWith('.css')) {
    // Don't process, keep as @import statement
}

// Reference option handling
if (importNode.options.reference) {
    // Add to frames for lookups
    // But mark for no output unless used
}
```

### Related Documentation

- Previous agent tasks exist for these:
  - `.claude/agents/agent-import-reference/TASK.md`
  - `.claude/agents/agent-import-reference-issues/TASK.md`
- Deep analysis: `.claude/reference-issues/ISSUE_IMPORTS.md`

### Related Tests

Other import tests should keep working:
- `import-inline` - Inline import option
- `import-once` - Import once behavior
- `impor` (yes, typo in test name) - Basic imports (currently ✅ perfect match)

## Notes

Since these are regressions (were compiling before), the fix might involve:
- Reverting recent changes that broke functionality
- Re-adding deleted code
- Fixing conditional logic that changed

The goal is to restore compilation success for both tests. Getting the output exactly right can be a follow-up task if needed - the critical issue is that they currently fail to compile at all.
