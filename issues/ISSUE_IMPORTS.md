# Import Issues - Agent Task

## 🎯 Mission
Fix 4-5 failing tests related to import functionality (interpolation, reference, module imports)

## 📊 Status
- **Tests Failing**: 4-5
- **Priority**: High (Wave 1 - Independent)
- **Complexity**: High
- **Independence**: HIGH - Can be fixed in parallel with other issues

## ❌ Failing Tests

### 1. import-interpolation
**File**: `packages/test-data/less/_main/import-interpolation.less`
**Error**: `open import/import-@{in}@{terpolation}.less: no such file or directory`
**Status**: DEFERRED (See note below)

**Test Content**:
```less
@my_theme: "test";
@import "import/import-@{my_theme}-e.less";  // Variable in import path
@import "import/import-@{in}@{terpolation}.less";  // Multiple variables
```

**Problem**: Import paths with variable interpolation (`@{var}`) are being evaluated at parse time instead of runtime. The parser tries to open `import-@{in}@{terpolation}.less` literally before evaluating the variables.

**Root Cause**: Architectural - the import resolution happens during parsing, but variable interpolation requires evaluation context. JavaScript less.js has the same limitation and handles this during the import visitor phase.

**NOTE**: This issue was documented in `IMPORT_INTERPOLATION_INVESTIGATION.md` and marked as deferred pending architecture refactor. **Recommend DEFER this test for now.**

---

### 2. import-reference
**File**: `packages/test-data/less/_main/import-reference.less`
**Error**: `open test.css: no such file or directory`

**Test Content**:
```less
@import (reference) url("import-once.less");
@import (reference) url("css-3.less");
@import (reference) url("media.less");
@import (reference) url("import/import-reference.less");
@import (reference) url("import/css-import.less");

.b { .z(); }  // Uses mixin from referenced import
```

**Problem**: The `(reference)` option should prevent CSS output for the imported file, but currently:
1. CSS files (`test.css`) are being treated as file imports instead of staying as `@import` statements
2. The reference flag might not be properly preserved during import processing

**Investigation Points**:
- How is `(reference)` option stored and checked?
- Is the visibility flag being set correctly?
- Are CSS imports being handled differently than LESS imports?

---

### 3. import-reference-issues
**File**: `packages/test-data/less/_main/import-reference-issues.less`
**Error**: `#Namespace > .mixin is undefined`

**Problem**: When importing with `(reference)`, mixins/rulesets from the imported file are not accessible. This suggests the reference flag is making the imports completely invisible instead of "reference-only visible".

**Expected Behavior**:
- Referenced imports should NOT output CSS by default
- But mixins from referenced imports SHOULD be usable via extend or direct calls
- Only output CSS for referenced selectors when explicitly used

**Investigation Points**:
- Check if imported rulesets are being marked as invisible when `(reference)` is used
- Verify namespace resolution works with referenced imports
- Compare with JS implementation of reference imports

---

### 4. import-module
**File**: `packages/test-data/less/_main/import-module.less`
**Error**: `open @less/test-import-module/one/1.less: no such file or directory`

**Problem**: Module imports (using `@less/module-name` syntax) are not being resolved through the module resolution system.

**Expected Behavior**:
- Imports starting with `@less/` should be resolved as module imports
- Should look in node_modules or configured module paths
- Similar to Node.js module resolution

**Investigation Points**:
- Is module import syntax being recognized?
- Is the module resolution path configured in import manager?
- Check if there's a module resolver that needs to be implemented

---

### 5. google (Remote Import)
**File**: `packages/test-data/less/process-imports/google.less`
**Error**: `Failed to fetch remote file: Get "https://fonts.googleapis.com/css?family=Open+Sans:400,700": dial tcp: lookup fonts.googleapis.com on [::1]:53`

**Problem**: Remote imports (HTTP/HTTPS URLs) are failing due to network issues or not being properly handled.

**Note**: This might be a test environment issue (DNS resolution failing). May need to:
- Skip this test in environments without internet
- Mock the HTTP request
- Or implement proper remote import handling

**Recommendation**: LOW PRIORITY - likely environment issue, not code issue.

---

## 🔍 Go Files to Investigate

### Primary Files
1. **`import.go`** - Import node and option handling
   - Check how `reference` option is stored
   - Verify option flags are preserved

2. **`import_visitor.go`** - Import processing during evaluation
   - Check reference import visibility logic
   - Verify CSS vs LESS import handling
   - Look for where imports are processed and inserted

3. **`import_manager.go`** - Import resolution and file loading
   - Check file resolution logic
   - Verify module import path resolution
   - Check if CSS files are handled specially

4. **`file_manager.go`** - File system operations
   - Verify path resolution
   - Check how relative paths are handled

### Secondary Files
5. **`set_tree_visibility_visitor.go`** - Controls visibility of nodes
   - Might be relevant for reference imports
   - Check if reference flag affects visibility

6. **`render.go`** - Main rendering pipeline
   - Understand how imports are processed in the pipeline

---

## 📚 JavaScript Reference Files

Study these to understand correct behavior:

1. **`packages/less/src/less/import-manager.js`**
   - How reference imports work
   - Module resolution logic
   - CSS import handling

2. **`packages/less/src/less/visitors/import-visitor.js`**
   - How imports are processed during evaluation
   - Reference flag handling
   - Visibility control

3. **`packages/less/src/less/tree/import.js`**
   - Import node structure
   - Option handling

---

## ✅ Success Criteria

### Minimum Success (2/5 tests)
- `import-reference` - CSS imports handled correctly
- `import-reference-issues` - Referenced mixins accessible

### Target Success (3/5 tests)
- Above + `import-module` - Module resolution working

### Stretch Goal (4/5 tests)
- Above + `google` - Remote imports (if feasible)
- `import-interpolation` - DEFER (architectural issue)

---

## 🚫 Constraints

1. **NEVER modify any .js files**
2. **Must pass unit tests**: `pnpm -w test:go:unit`
3. **Must pass target tests**: `pnpm -w test:go:filter -- "import-reference"`
4. **No regressions**: All currently passing tests must still pass

---

## 🧪 Testing Strategy

### Run Specific Tests
```bash
# Test individual cases
go test -run "TestIntegrationSuite/main/import-reference" -v
go test -run "TestIntegrationSuite/main/import-reference-issues" -v
go test -run "TestIntegrationSuite/main/import-module" -v

# With debug output
LESS_GO_DEBUG=1 go test -run "TestIntegrationSuite/main/import-reference" -v

# With trace
LESS_GO_TRACE=1 go test -run "TestIntegrationSuite/main/import-reference" -v
```

### Verify No Regressions
```bash
# Run all unit tests
pnpm -w test:go:unit

# Run full integration suite
pnpm -w test:go:summary
```

---

## 📝 Expected Changes

### Likely Changes Needed

1. **import.go** - Ensure reference flag is properly stored and accessible
2. **import_visitor.go** - Fix reference import visibility logic:
   - Don't output CSS for referenced imports
   - But make mixins/variables accessible
3. **import_manager.go** - Fix CSS import handling:
   - CSS files should remain as `@import` statements
   - Not be read and processed as LESS files
4. **Module resolution** - Implement or fix module path resolution

### Testing Pattern

For each fix:
1. Make minimal change
2. Run specific test
3. If passing, run unit tests
4. If unit tests pass, run full integration suite
5. Commit with clear message

---

## 🎯 Debugging Hints

### Reference Import Issue
```bash
# Add debug output in import_visitor.go
fmt.Printf("[IMPORT-DEBUG] Processing import: %s, reference=%v\n",
    importPath, importNode.Options.Reference)

# Check if referenced rulesets are being added to frame
fmt.Printf("[IMPORT-DEBUG] Adding ruleset to frame: visible=%v\n",
    ruleset.GetVisibility())
```

### CSS Import Issue
```bash
# Check in import_manager.go
fmt.Printf("[IMPORT-DEBUG] File extension: %s, isCSSImport=%v\n",
    filepath.Ext(path), isCSSImport)
```

---

## 📊 Estimated Impact

- **Tests Fixed**: 2-3 (out of 4-5)
- **Other Tests Potentially Improved**: 5-10 tests that use imports
- **Risk Level**: Medium - imports are used throughout, but well isolated

---

## 🔄 Iteration Strategy

### Round 1: Reference Imports (Highest Value)
1. Fix `import-reference` - CSS import handling
2. Fix `import-reference-issues` - Referenced mixin visibility
3. Commit and push

### Round 2: Module Imports (If time permits)
1. Fix `import-module` - Module path resolution
2. Commit and push

### Round 3: Remote Imports (Low priority)
1. Investigate `google` test failure
2. May be environment issue, not code issue

---

## 📋 Commit Message Template

```
Fix import reference functionality

- Preserve reference flag during import processing
- Make referenced mixins accessible while preventing CSS output
- Handle CSS imports correctly (keep as @import statements)
- [Optional] Fix module import path resolution

Tests fixed:
- import-reference: ✅
- import-reference-issues: ✅
- [Optional] import-module: ✅

Tests deferred:
- import-interpolation: Architectural limitation (see IMPORT_INTERPOLATION_INVESTIGATION.md)
```

---

## 🚀 When Done

1. **Commit** to branch: `claude/fix-imports-<your-session-id>`
2. **Push** to remote: `git push -u origin claude/fix-imports-<your-session-id>`
3. **Report**: "Fixed X/5 import tests: [test names]. Deferred import-interpolation (architectural)."

---

## 💡 Key Insights from RUNTIME_ISSUES.md

The parser is working correctly! All issues are in runtime evaluation and import processing:
- 92.4% compilation rate means parser handles import syntax
- Issues are in how imports are resolved and processed
- Focus on import_visitor.go and import_manager.go logic

---

## ⚠️ Special Notes

1. **import-interpolation is DEFERRED** - Don't spend time on this, it requires architectural changes
2. **google test** might be environment-specific - Low priority
3. **Focus on reference imports** - Highest value, most likely to fix multiple tests
4. **CSS imports are special** - Should NOT be processed as LESS, should remain as @import
