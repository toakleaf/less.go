# Quick Start Guide for Independent Agents
## less.go Port - 2025-12-02

---

## Current Status - Port Complete!

### Baseline Metrics (MUST MAINTAIN)
- **Unit Tests**: 3,012 tests passing (100%)
- **Perfect Matches**: 100 tests
- **Error Tests**: 91 tests (correctly failing)
- **Overall Success**: 100% (191/191 tests)
- **NO REGRESSIONS**: Maintaining all progress

---

## Project Structure

```
less.go/
├── .claude/
│   ├── AGENT_WORK_QUEUE.md        ← Current work items
│   ├── strategy/MASTER_PLAN.md    ← Overall strategy
│   └── tasks/                     ← Task details
│
├── less/                           ← Go implementation (WHERE YOU MAKE CHANGES)
│   ├── parser.go                  ← LESS parsing
│   ├── evaluator.go               ← Expression evaluation
│   ├── import.go                  ← Import handling
│   └── (other .go files)
│
└── testdata/
    ├── less/                       ← Test input LESS files
    └── css/                        ← Expected CSS output files
```

---

## The Work Cycle

### 1. Before You Start (5 minutes)
```bash
cd /home/user/less.go
git fetch origin
git checkout -b claude/fix-{taskname}-{yourID}

# Get baseline numbers
pnpm test:go:unit | tail -5          # Should see: PASS
LESS_GO_QUIET=1 pnpm test:go 2>&1 | grep "Perfect CSS"  # Should see: 100
```

### 2. Make Your Changes
- Edit Go files in `less/`
- Follow patterns from working code
- Compare with JavaScript implementation in `reference/less.js/`
- Use LESS_GO_DIFF to see differences

### 3. Test Incrementally
```bash
# Test your specific fix
LESS_GO_DIFF=1 pnpm test:go 2>&1 | grep -A 20 "your-test-name"

# Check for unit test regressions
pnpm test:go:unit

# See full impact
LESS_GO_QUIET=1 pnpm test:go 2>&1 | tail -30
```

### 4. Verify No Regressions (CRITICAL!)
```bash
pnpm test:go:unit          # MUST: 3,012 passing
LESS_GO_QUIET=1 pnpm test:go 2>&1 | grep "Perfect CSS"  # MUST: >= 100
```

### 5. Commit & Push
```bash
git add -A
git commit -m "Fix {feature}: Brief description"
git push origin claude/fix-{taskname}-{yourID}
```

---

## Golden Rules

1. **ALWAYS check baseline before starting**
2. **NEVER let unit tests fail**
3. **NEVER reduce perfect match count** (currently 100)
4. **ALWAYS test incrementally**
5. **ALWAYS read the JavaScript version when confused** (in `reference/less.js/`)
6. **ALWAYS verify no regressions before committing**

---

## If Things Go Wrong

### Unit tests failing?
```bash
pnpm test:go:unit 2>&1 | grep -A 5 "FAIL"
# Review your changes, understand what you modified
```

### Perfect match count dropped?
```bash
LESS_GO_DIFF=1 pnpm test:go 2>&1 | grep -B 3 "Output Differs"
# Revert your last change, understand why it broke something
```

---

## Debugging Commands

```bash
# See test data
cat testdata/less/_main/import-reference.less  # Input
cat testdata/css/_main/import-reference.css    # Expected

# Compare with actual output
LESS_GO_DIFF=1 pnpm test:go 2>&1 | grep -A 30 "import-reference"

# Full debug mode
LESS_GO_DEBUG=1 LESS_GO_TRACE=1 LESS_GO_DIFF=1 pnpm test:go
```

---

**You've got this!** The port is complete - now it's about maintenance and improvements.

---
Last Updated: 2025-12-02
