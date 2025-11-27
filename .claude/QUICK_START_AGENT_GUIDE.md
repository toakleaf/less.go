# Quick Start Guide for Independent Agents
## less.go Port - 2025-11-27

---

## Current Status - Only 3 Tests Remaining!

### Baseline Metrics (MUST MAINTAIN)
- **Unit Tests**: 3,012 tests passing (100%)
- **Perfect Matches**: 89 tests (48.4%)
- **Output Differences**: 3 tests (1.6%)
- **Error Tests**: 89 tests (48.4%)
- **Overall Success**: 96.7%
- **NO REGRESSIONS**: Maintaining all progress

---

## Only 3 Tasks Left!

| # | Task | Impact | Difficulty |
|---|------|--------|------------|
| 1 | Import Reference | +2 tests | Medium |
| 2 | URL Handling | +1 test | Low-Medium |

### Task 1: Import Reference (HIGH PRIORITY)
**Files**: `import-reference`, `import-reference-issues`
**Details**: `.claude/tasks/runtime-failures/import-reference.md`

Reference imports outputting CSS when they shouldn't. Files imported with `(reference)` option should not output CSS, but selectors/mixins should be available.

### Task 2: URL Handling (MEDIUM PRIORITY)
**Files**: `urls` (main suite only)
**Note**: Other URL tests (static-urls, url-args) are now passing

URL edge cases in the main urls test.

---

## The Work Cycle

### 1. Before You Start (5 minutes)
```bash
cd /home/user/less.go
git fetch origin
git checkout -b claude/fix-{taskname}-{yourID}

# Get baseline numbers
pnpm -w test:go:unit | tail -5          # Should see: PASS
LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | grep "Perfect CSS"  # Should see: 89
```

### 2. Make Your Changes
- Edit Go files in `packages/less/src/less/less_go/`
- Follow patterns from working code
- Compare with JavaScript implementation
- Use LESS_GO_DIFF to see differences

### 3. Test Incrementally
```bash
# Test your specific fix
LESS_GO_DIFF=1 pnpm -w test:go 2>&1 | grep -A 20 "your-test-name"

# Check for unit test regressions
pnpm -w test:go:unit

# See full impact
LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | tail -30
```

### 4. Verify No Regressions (CRITICAL!)
```bash
pnpm -w test:go:unit          # MUST: 3,012 passing
LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | grep "Perfect CSS"  # MUST: >= 89
```

### 5. Commit & Push
```bash
git add -A
git commit -m "Fix {feature}: Brief description"
git push origin claude/fix-{taskname}-{yourID}
```

---

## Files You'll Need

```
less.go/
├── .claude/
│   ├── AGENT_WORK_QUEUE.md        ← Current work items
│   ├── strategy/MASTER_PLAN.md    ← Overall strategy
│   └── tasks/runtime-failures/    ← Task details
│
├── packages/less/src/less/less_go/  ← WHERE YOU MAKE CHANGES
│   ├── import.go, import_visitor.go ← For import-reference
│   ├── url.go, ruleset.go           ← For urls
│   └── (other files as needed)
│
└── packages/test-data/
    ├── less/_main/                  ← Test input files
    └── css/_main/                   ← Expected output files
```

---

## Golden Rules

1. **ALWAYS check baseline before starting**
2. **NEVER let unit tests fail**
3. **NEVER reduce perfect match count** (currently 89)
4. **ALWAYS test incrementally**
5. **ALWAYS read the JavaScript version when confused**
6. **ALWAYS verify no regressions before committing**

---

## If Things Go Wrong

### Unit tests failing?
```bash
pnpm -w test:go:unit 2>&1 | grep -A 5 "FAIL"
# Review your changes, understand what you modified
```

### Perfect match count dropped?
```bash
LESS_GO_DIFF=1 pnpm -w test:go 2>&1 | grep -B 3 "Output Differs"
# Revert your last change, understand why it broke something
```

---

## Debugging Commands

```bash
# See test data
cat packages/test-data/less/_main/import-reference.less  # Input
cat packages/test-data/css/_main/import-reference.css    # Expected

# Compare with actual output
LESS_GO_DIFF=1 pnpm -w test:go 2>&1 | grep -A 30 "import-reference"

# Full debug mode
LESS_GO_DEBUG=1 LESS_GO_TRACE=1 LESS_GO_DIFF=1 pnpm -w test:go
```

---

## What Success Looks Like

When you fix both import-reference tests:
```
✅ Perfect CSS Matches: 91 (49.5%)  # Was 89, now 91 (+2!)
```

When all 3 remaining tests are fixed:
```
✅ Perfect CSS Matches: 92 (50.0%)
✅ Success Rate: 98.4% (181/184 tests)
# Only 3 external dependency tests remain as expected failures
```

---

**You've got this!** Only 3 tests left to fix.

---
Last Updated: 2025-11-27
