# Agent: Import Reference Functionality

## 🎯 Your Mission
Fix import reference functionality - making mixins from referenced imports accessible without outputting CSS.

## 📊 Status
**Tests to Fix**: 2-3 of 5 (defer 2)
**Branch**: `claude/fix-imports-<your-session-id>`
**Independence**: MEDIUM - Touches import_manager.go (agent-paths also uses this, but different sections)

## ❌ The Problems

### Test 1: import-reference
```less
@import (reference) url("import-once.less");
.b { .z(); }  // Use mixin from referenced import
```
**Error**: `open test.css: no such file or directory`

**Issue**: CSS files are being treated as file imports instead of staying as `@import` statements.

### Test 2: import-reference-issues
```less
@import (reference) url("some-file.less");
#Namespace > .mixin();  // Should be accessible
```
**Error**: `#Namespace > .mixin is undefined`

**Issue**: Referenced imports should make mixins accessible but not output CSS.

### Test 3: import-module (stretch goal)
**Error**: `open @less/test-import-module/one/1.less: no such file or directory`

**Issue**: Module imports with `@less/` prefix not being resolved.

### Test 4 & 5: DEFER
- `import-interpolation`: Architectural issue, documented as deferred
- `google`: Network issue with remote imports, low priority

## 🔍 Files You'll Modify
- `packages/less/src/less/less_go/import_visitor.go` - PRIMARY: Import processing and visibility
- `packages/less/src/less/less_go/import_manager.go` - Import resolution, CSS detection
- `packages/less/src/less/less_go/import.go` - Reference flag handling
- `packages/less/src/less/less_go/set_tree_visibility_visitor.go` - Maybe: visibility control

**Potential Conflict**: agent-paths also modifies import_manager.go. You focus on:
- Reference flag preservation
- CSS import detection (don't load CSS files, keep as @import)
- Making referenced mixins accessible

They focus on:
- Include path searching

Different sections, should be fine.

## 🔑 Key Concepts

### Reference Imports Should:
1. ✅ Make mixins/variables available for use
2. ✅ Allow extend to reference selectors
3. ❌ NOT output the imported CSS by default
4. ✅ Only output CSS for explicitly used (extended/mixed) selectors

### CSS Imports Should:
1. Stay as `@import` statements in output
2. NOT be loaded and processed as LESS files
3. File extension `.css` is the indicator

## ✅ Success Criteria
- [ ] `import-reference` test passes
- [ ] `import-reference-issues` test passes
- [ ] (Stretch) `import-module` test passes
- [ ] All unit tests pass: `pnpm -w test:go:unit`
- [ ] No regressions: `pnpm -w test:go:summary`

## 🧪 Test Commands
```bash
cd packages/less/src/less/less_go

go test -run "TestIntegrationSuite/main/import-reference" -v
go test -run "TestIntegrationSuite/main/import-reference-issues" -v
go test -run "TestIntegrationSuite/main/import-module" -v

# With debug
LESS_GO_DEBUG=1 go test -run "TestIntegrationSuite/main/import-reference" -v

# After fix
pnpm -w test:go:unit
pnpm -w test:go:summary
```

## 🔬 Debug Strategy

### 1. Add debug output:
```go
// In import_visitor.go
fmt.Printf("[IMPORT] Processing: %s, reference=%v, isCSS=%v\n",
    importPath, importNode.Options.Reference, isCSSImport)

// In import_manager.go
fmt.Printf("[IMPORT] File extension: %s, treating as CSS: %v\n",
    filepath.Ext(path), isCSSImport)
```

### 2. Questions to answer:
- Is the `reference` flag being preserved during import processing?
- Are CSS files (*.css) being detected and handled differently?
- Are referenced rulesets being added to frames but marked as non-output?
- Is the visibility flag being set correctly?

### 3. JavaScript reference:
- `packages/less/src/less/import-manager.js` - Reference import handling
- `packages/less/src/less/visitors/import-visitor.js` - Import processing

## 📋 Likely Fixes

### Fix 1: CSS Import Detection
```go
// In import_manager.go
func (im *ImportManager) shouldProcessAsLESS(path string) bool {
    // If it's a .css file, don't process it as LESS
    if filepath.Ext(path) == ".css" {
        return false
    }
    return true
}
```

### Fix 2: Reference Flag Handling
```go
// In import_visitor.go
if importNode.Options.Reference {
    // Don't output CSS by default
    ruleset.SetVisibility(VisibilityReferenceOnly)
    // But make mixins/variables accessible in frames
}
```

## 📋 When Done
```bash
git add packages/less/src/less/less_go/*.go
git commit -m "Fix import reference functionality

- Preserve reference flag during import processing
- Make referenced mixins accessible while preventing CSS output
- Handle CSS imports correctly (keep as @import statements)

Tests fixed:
- import-reference: ✅
- import-reference-issues: ✅
[- import-module: ✅]

Tests deferred:
- import-interpolation: Architectural limitation
- google: Network/environment issue"

git push -u origin claude/fix-imports-<your-session-id>
```

Report: "Fixed 2-3/5 import tests (2 deferred as documented). Ready for PR."
