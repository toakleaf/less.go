# Agent: Fix import-reference

## 🎯 Your Single Mission
Fix ONE test: `import-reference` - CSS imports should stay as @import statements

## 📊 Status
**Tests to Fix**: 1 (just import-reference)
**Branch**: `claude/fix-import-reference-<your-session-id>`
**Independence**: MEDIUM - May touch import_manager.go (agent-paths also uses this, different section)

## ❌ The Problem
```less
@import (reference) url("import-once.less");
@import (reference) url("css-3.less");
.b { .z(); }  // Use mixin from referenced import
```
**Error**: `open test.css: no such file or directory`

**Issue**: CSS files (*.css) are being treated as LESS files and being loaded/processed. They should remain as `@import` statements in the output.

## 🔍 Files You'll Modify
- `packages/less/src/less/less_go/import_manager.go` - PRIMARY: CSS file detection
- `packages/less/src/less/less_go/import_visitor.go` - Maybe: import processing
- `packages/less/src/less/less_go/import.go` - Maybe: import options

**Potential Conflict**: agent-paths also touches import_manager.go, but they focus on include path searching. You focus on CSS file detection. Different sections.

## ✅ Success Criteria
- [ ] `import-reference` test passes
- [ ] All unit tests pass: `pnpm -w test:go:unit`
- [ ] No regressions: `pnpm -w test:go:summary`

## 🧪 Test Commands
```bash
cd packages/less/src/less/less_go

go test -run "TestIntegrationSuite/main/import-reference" -v

# With debug
LESS_GO_DEBUG=1 go test -run "TestIntegrationSuite/main/import-reference" -v

# After fix
pnpm -w test:go:unit
pnpm -w test:go:summary
```

## 🔑 Key Insight
CSS files should:
1. Stay as `@import` statements in output
2. NOT be loaded and processed as LESS files
3. File extension `.css` is the indicator

## 🔬 Debug Strategy
Add debug output:
```go
// In import_manager.go
fmt.Printf("[IMPORT] File: %s, ext: %s, isCSS: %v\n",
    path, filepath.Ext(path), isCSSFile)
```

## 📋 Likely Fix
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

Then in the import loading logic, check this before trying to read/parse the file.

## 📋 When Done
```bash
git add packages/less/src/less/less_go/*.go
git commit -m "Fix import-reference: handle CSS imports correctly

CSS files should remain as @import statements, not be processed as LESS.
Added CSS file detection to keep them as-is in output.

Test fixed:
- import-reference: ✅"

git push -u origin claude/fix-import-reference-<your-session-id>
```

Report: "Fixed import-reference test. Ready for PR."

## 📚 Reference
See `.claude/reference-issues/ISSUE_IMPORTS.md` for full analysis (covers multiple import issues, you only fix import-reference).
