# Agent: Namespace Variable Call Fix

## 🎯 Your Mission
Fix variable calls to mixin results - when mixins are assigned to variables and then called.

## 📊 Status
**Tests to Fix**: 2
**Branch**: `claude/fix-namespacing-<your-session-id>`
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

The mixin call result should be stored as a callable DetachedRuleset, but something's wrong with how it's evaluated or stored.

## 🔍 Files You'll Modify
- `packages/less/src/less/less_go/variable_call.go` - Variable call evaluation (error is here)
- `packages/less/src/less/less_go/variable.go` - Variable storage/evaluation
- `packages/less/src/less/less_go/mixin_call.go` - Mixin call return values
- `packages/less/src/less/less_go/detached_ruleset.go` - Maybe: callable interface

**You will NOT conflict with other agents** - they work on different files.

## 🔑 Key Insight
This is similar to **Issue #2** (detached-rulesets) which was FIXED. That fix involved:
- Checking `Eval(any) (any, error)` signature BEFORE `Eval(any) any`
- Ensuring Expression nodes are evaluated properly

Your issue might need a similar pattern, or the mixin call needs to return a proper DetachedRuleset.

## ✅ Success Criteria
- [ ] `namespacing-6` test passes
- [ ] `namespacing-functions` test passes
- [ ] All unit tests pass: `pnpm -w test:go:unit`
- [ ] No regressions: `pnpm -w test:go:summary`

## 🧪 Test Commands
```bash
cd packages/less/src/less/less_go

# With trace (very helpful!)
LESS_GO_TRACE=1 go test -run "TestIntegrationSuite/namespacing/namespacing-6" -v 2>&1 | grep -A5 -B5 "@alias"

# Regular test
go test -run "TestIntegrationSuite/namespacing/namespacing-6" -v
go test -run "TestIntegrationSuite/namespacing/namespacing-functions" -v

# After fix
pnpm -w test:go:unit
pnpm -w test:go:summary
```

## 🔬 Debug Strategy

### 1. Add trace output to understand the flow:
```go
// In variable_call.go
fmt.Printf("[VARCALL] Calling variable @%s, value type: %T\n", vc.Variable, value)

// In mixin_call.go
fmt.Printf("[MIXIN] Mixin call result type: %T\n", result)

// In variable.go
fmt.Printf("[VAR] Variable %s assigned value type: %T\n", v.Name, evaluatedValue)
```

### 2. Questions to answer:
- What type is returned when `.something(foo)` is evaluated?
- What type is stored in the `@alias` variable?
- What type does `variable_call.go` expect to call?
- Is there a type mismatch or missing evaluation?

### 3. Compare with JavaScript
Look at `packages/less/src/less/tree/variable-call.js` to see how JS handles this.

## 📋 When Done
```bash
git add packages/less/src/less/less_go/*.go
git commit -m "Fix namespace variable call evaluation

When mixin calls like .something(foo) are assigned to variables and then
called as @alias(), the evaluation was failing with 'Could not evaluate
variable call @alias'.

Root cause: [Describe what was wrong]

Fix: [Describe what was changed]

Tests fixed:
- namespacing-6: ✅
- namespacing-functions: ✅"

git push -u origin claude/fix-namespacing-<your-session-id>
```

Report: "Fixed 2/2 namespacing tests. Ready for PR."
