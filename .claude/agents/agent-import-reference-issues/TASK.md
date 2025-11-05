# Agent: Fix import-reference-issues

## 🎯 Your Single Mission
Fix ONE test: `import-reference-issues` - make referenced mixins accessible

## 📊 Status
**Tests to Fix**: 1 (just import-reference-issues)
**Branch**: `claude/fix-import-reference-issues-<your-session-id>`
**Independence**: HIGH - Different files from other agents

## ❌ The Problem
```less
@import (reference) url("some-file.less");

// File contains:
#Namespace {
  .mixin() { color: red; }
}

// Try to use it:
.test {
  #Namespace > .mixin();  // Error: "#Namespace > .mixin is undefined"
}
```

**Issue**: When using `(reference)` option, the imported mixins should be accessible but currently they're not found.

## 🔍 Files You'll Modify
- `packages/less/src/less/less_go/import_visitor.go` - PRIMARY: Import processing
- `packages/less/src/less/less_go/import.go` - Reference flag handling
- `packages/less/src/less/less_go/set_tree_visibility_visitor.go` - Maybe: visibility control

**You will NOT conflict with other agents** - different focus.

## ✅ Success Criteria
- [ ] `import-reference-issues` test passes
- [ ] All unit tests pass: `pnpm -w test:go:unit`
- [ ] No regressions: `pnpm -w test:go:summary`

## 🧪 Test Commands
```bash
cd packages/less/src/less/less_go

go test -run "TestIntegrationSuite/main/import-reference-issues" -v

# With debug
LESS_GO_DEBUG=1 go test -run "TestIntegrationSuite/main/import-reference-issues" -v

# After fix
pnpm -w test:go:unit
pnpm -w test:go:summary
```

## 🔑 Key Concept
Reference imports should:
1. ✅ Make mixins/variables available in frames
2. ❌ NOT output the imported CSS by default
3. ✅ Only output CSS for explicitly used (extended/mixed) selectors

Currently: Mixins not being added to frames OR visibility is hiding them completely.

## 🔬 Debug Strategy
Add debug output:
```go
// In import_visitor.go
fmt.Printf("[IMPORT] Processing: %s, reference=%v\n",
    importPath, importNode.Options.Reference)
fmt.Printf("[IMPORT] Adding rulesets to frame: %d rulesets\n", len(rulesets))
```

## 📋 Likely Fix
In `import_visitor.go`, when processing referenced imports:
```go
if importNode.Options.Reference {
    // Add rulesets to frames so mixins are accessible
    // But mark them with special visibility flag
    for _, ruleset := range importedRulesets {
        // Make accessible
        context.Frames.Push(ruleset)
        // But don't output by default
        ruleset.SetReferenceOnly(true)
    }
}
```

## 📋 When Done
```bash
git add packages/less/src/less/less_go/*.go
git commit -m "Fix import-reference-issues: make referenced mixins accessible

Referenced imports now properly add mixins to frames while preventing
default CSS output. Mixins are accessible but only output when used.

Test fixed:
- import-reference-issues: ✅"

git push -u origin claude/fix-import-reference-issues-<your-session-id>
```

Report: "Fixed import-reference-issues test. Ready for PR."

## 📚 Reference
See `.claude/reference-issues/ISSUE_IMPORTS.md` for full analysis.
