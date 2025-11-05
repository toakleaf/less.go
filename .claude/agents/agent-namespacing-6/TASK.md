# Agent: Fix namespacing-6

## 🎯 Your Single Mission
Fix ONE test: `namespacing-6` - variable calls to mixin results

## 📊 Status
**Tests to Fix**: 1 (just namespacing-6)
**Branch**: `claude/fix-namespacing-6-<your-session-id>`
**Independence**: HIGH - No conflicts with other agents

## ❌ The Problem
```less
.something(foo) {
  width: 10px;
}

.rule-1 {
  @alias: .something(foo);  // Assign mixin call to variable
  @alias();                 // Call the variable - FAILS
}
```
**Error**: `Could not evaluate variable call @alias`

## 🔍 Files You'll Modify
- `packages/less/src/less/less_go/variable_call.go` - Where the error occurs
- `packages/less/src/less/less_go/variable.go` - Variable storage/evaluation
- `packages/less/src/less/less_go/mixin_call.go` - Mixin call return values

**You will NOT conflict with other agents** - they work on different files.

## ✅ Success Criteria
- [ ] `namespacing-6` test passes
- [ ] All unit tests pass: `pnpm -w test:go:unit`
- [ ] No regressions: `pnpm -w test:go:summary`

## 🧪 Test Commands
```bash
cd packages/less/src/less/less_go

# With trace (very helpful!)
LESS_GO_TRACE=1 go test -run "TestIntegrationSuite/namespacing/namespacing-6" -v 2>&1 | grep -A5 -B5 "@alias"

# Regular test
go test -run "TestIntegrationSuite/namespacing/namespacing-6" -v

# After fix
pnpm -w test:go:unit
pnpm -w test:go:summary
```

## 🔑 Key Insight
This is similar to **Issue #2** (detached-rulesets) which was FIXED. That fix involved checking `Eval(any) (any, error)` signature BEFORE `Eval(any) any`.

Your issue likely needs similar pattern - the mixin call result needs to be properly evaluated and stored as a callable type (DetachedRuleset).

## 🔬 Debug Strategy
Add trace output:
```go
// In variable_call.go
fmt.Printf("[VARCALL] Variable @%s, value type: %T\n", vc.Variable, value)

// In mixin_call.go
fmt.Printf("[MIXIN] Result type: %T\n", result)

// In variable.go
fmt.Printf("[VAR] Storing @%s, value type: %T\n", v.Name, evaluatedValue)
```

## 📋 When Done
```bash
git add packages/less/src/less/less_go/*.go
git commit -m "Fix namespacing-6: variable calls to mixin results

When mixin calls are assigned to variables and then called,
evaluation was failing. Fixed [describe what you fixed].

Test fixed:
- namespacing-6: ✅"

git push -u origin claude/fix-namespacing-6-<your-session-id>
```

Report: "Fixed namespacing-6 test. Ready for PR."

## 📚 Reference
See `.claude/reference-issues/ISSUE_NAMESPACING.md` for deep dive analysis.
