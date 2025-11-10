# Prompt: Investigate and Fix extend-chaining Test

## Context

You're working on **less.go**, a Go port of the less.js CSS preprocessor. The project is at **76.1% overall success rate** with **78/184 perfect CSS matches**.

There's a discrepancy with the `extend-chaining` integration test:
- **Documentation claims**: This was recently fixed (79 perfect matches total)
- **Current testing shows**: Test is failing with output differences (78 perfect matches)

Your task is to:
1. Determine if this is a **regression** (was passing, now broken) or **documentation error** (never actually fixed)
2. If it's broken, **fix it**
3. Ensure **zero regressions** in other tests

## Current Branch

Create a new branch for this work:
```bash
git checkout -b claude/fix-extend-chaining-<your-session-id>
```

## Investigation Steps

### Step 1: Check Git History

```bash
# Look for commits that mention extend-chaining
git log --all --grep="extend-chaining" --oneline

# Check recent commits on assessment branches
git log --all --grep="extend" --oneline | head -20

# Look at recent PRs/branches
git branch -a | grep -i extend
```

### Step 2: Check Historical Test Results

```bash
# Look at previous assessment reports
ls -la .claude/ASSESSMENT_*.md
cat .claude/ASSESSMENT_REPORT_2025-11-10.md | grep -A 5 "extend-chaining"
cat .claude/ASSESSMENT_REPORT_2025-11-09.md | grep -A 5 "extend-chaining"

# Check the work queue
cat .claude/AGENT_WORK_QUEUE.md | grep -i extend
```

### Step 3: Run the Test and Examine Output

```bash
# Run the specific test
pnpm -w test:go:filter -- "extend-chaining"

# Or run directly
cd packages/less/src/less/less_go
go test -v -run "TestIntegrationSuite/main/extend-chaining"
```

### Step 4: Compare Expected vs Actual

```bash
# View the test input
cat ../../../../test-data/less/_main/extend-chaining.less

# View expected output
cat ../../../../test-data/css/_main/extend-chaining.css

# Run test and capture actual output (you'll need to modify test or capture manually)
# The test runs through integration_suite_test.go
```

## Understanding the Test

The `extend-chaining` test covers:
- Simple extend chaining: `.a`, `.b:extend(.a)`, `.c:extend(.b)`
- Multi-level chaining with various orders
- Extend with `all` selector
- Self-referencing extends
- Circular references
- Extends inside rulesets
- Media query extends (shouldn't extend outside media query)

**Expected behavior**: When `.c:extend(.b)` and `.b:extend(.a)`, then `.c` should also get `.a`'s styles (transitive extending).

## Common Issues to Check

### Issue 1: Extend Chain Resolution
```go
// In packages/less/src/less/less_go/extend_visitor.go
// Check the visitRule method and how it processes extend chains
```

**Look for**:
- Does the extend visitor handle transitive extends?
- Are extend chains resolved iteratively until no more matches?
- Is there a loop detection mechanism for circular extends?

### Issue 2: Selector Matching
```go
// In packages/less/src/less/less_go/selector.go
// Check the Match method
```

**Look for**:
- Does extend matching work correctly for chained extends?
- Are extended selectors added to the match pool for subsequent extends?

### Issue 3: Media Query Isolation
```go
// In packages/less/src/less/less_go/media.go
// Check how extends inside media queries are isolated
```

**Look for**:
- Extends inside `@media` should not match selectors outside
- Extends outside `@media` should not match selectors inside

## Reference Implementation

Compare with JavaScript implementation:
```bash
# Look at the JS extend visitor
cat packages/less/src/less/visitors/extend-visitor.js | grep -A 20 "visitRule"
```

## Testing Strategy

### 1. Create a Minimal Reproduction

```bash
cat > /tmp/test-extend-chain.less << 'EOF'
.a { color: black; }
.b:extend(.a) {}
.c:extend(.b) {}
EOF

# Test with less.js (for comparison)
cd packages/less
npx lessc /tmp/test-extend-chain.less

# Expected output:
# .a, .b, .c { color: black; }
```

### 2. Run Full Test Suite After Fix

```bash
# Unit tests must pass
pnpm -w test:go:unit

# Integration tests
pnpm -w test:go

# Verify no regressions
pnpm -w test:go 2>&1 | grep "✅.*Perfect match!" | wc -l
# Expected: 79 (if this was the only issue) or same as before (if fix doesn't work yet)
```

## Success Criteria

- ✅ `extend-chaining` test shows "Perfect match!"
- ✅ All unit tests still pass (`pnpm -w test:go:unit`)
- ✅ Perfect match count increases to 79 (or stays at 78 if test was never actually passing)
- ✅ Zero regressions in other extend tests:
  - extend ✅
  - extend-clearfix ✅
  - extend-exact ✅
  - extend-media ✅
  - extend-nest ✅
  - extend-selector ✅

## Files to Focus On

1. **`packages/less/src/less/less_go/extend_visitor.go`** - Main extend processing logic
2. **`packages/less/src/less/less_go/selector.go`** - Selector matching
3. **`packages/less/src/less/less_go/ruleset.go`** - How extends are stored/processed
4. **`packages/less/src/less/less_go/media.go`** - Media query isolation

## Debugging Commands

```bash
# Enable debug logging (if available)
LESS_GO_DEBUG=1 pnpm -w test:go:filter -- "extend-chaining"

# Run with verbose output
go test -v -run "TestIntegrationSuite/main/extend-chaining" 2>&1 | less

# Compare with a working extend test
diff -u \
  <(pnpm -w test:go:filter -- "extend-exact" 2>&1) \
  <(pnpm -w test:go:filter -- "extend-chaining" 2>&1)
```

## When You're Done

### Report Your Findings

Create a summary file documenting what you found:

```bash
cat > .claude/EXTEND_CHAINING_INVESTIGATION.md << 'EOF'
# extend-chaining Investigation Results

## Was it a regression?
[YES/NO and explain]

## Root cause:
[Describe what was wrong]

## Fix applied:
[Describe what you changed]

## Test results:
- extend-chaining: [PASS/FAIL]
- Perfect match count: [number]
- Regressions: [number]
EOF
```

### Commit and Push

```bash
git add -A
git commit -m "Fix extend-chaining test [or: Document extend-chaining was never fixed]

[Describe what you found and what you fixed]

Test results:
- extend-chaining: [status]
- Perfect matches: [count]
- Zero regressions

[If it was a regression, explain how it broke]
[If it was never fixed, explain what still needs to be done]"

git push -u origin claude/fix-extend-chaining-<your-session-id>
```

## Additional Context

**Recent extend-related work**:
- All other extend tests (6/7) are passing
- Extend functionality was heavily worked on in recent sessions
- No obvious recent commits that would have broken this

**Project state**:
- 78 perfect matches currently
- 76.1% overall success rate
- 2,290+ unit tests passing
- Zero confirmed regressions in other areas

**Similar successful fixes**:
- extend-clearfix - Fixed selector matching
- extend-nest - Fixed nested extend resolution
- extend-media - Fixed media query extend isolation

Good luck! Start with the investigation steps to determine if this is a regression or just a documentation error, then proceed accordingly.
