# Quick Start Guide for Independent Agents
## less.go Port - 2025-11-26

---

## 🚀 Your 10 Independent Tasks (Pick Any!)

### 📋 Task Priority List

| # | Task | Impact | Time | Difficulty | Best For |
|---|------|--------|------|------------|----------|
| 1 | Import Reference CSS Suppression | +2 tests | 2-3h | Medium | Feature fix |
| 2 | Detached Ruleset Media Queries | +1 test | 2-3h | Medium-High | Complex feature |
| 3 | URL Handling Fix | +2 tests | 2-3h | Medium | Feature fix |
| 4 | Media Query Formatting | +1 test | 2h | Medium | Formatting |
| 5 | Container Queries Support | +1 test | 2-3h | Medium | Feature add |
| 6 | Directive Bubbling Order | +1 test | 2-3h | Medium-High | Complex |
| 7 | Analysis: Root Causes | +Plan | 1-2h | Low | Knowledge |
| 8 | Error Handling: JS Undefined | +1 test | 1-2h | Low | Error check |
| 9 | Performance Analysis | +Plan | 1-2h | Low | Knowledge |
| 10 | Documentation Cleanup | +Org | 1h | Low | Admin |

---

## ⚡ Quick Stats You Need to Know

### Current Baseline (MUST MAINTAIN)
- ✅ **Unit Tests**: 2,304 tests passing (100%)
- ✅ **Perfect Matches**: 84 tests (45.7%)
- ✅ **Error Tests**: 88 tests (98.9% correct)
- ✅ **Overall Success**: 93.5%
- ✅ **NO REGRESSIONS**: Maintaining all progress

### What That Means
- If you fix something, these numbers must not go DOWN
- If unit tests start failing → you've introduced a bug
- If perfect matches decrease → you've broken something
- These are your success metrics

---

## 🔄 The Work Cycle (Super Simple)

### 1. **Before You Start** (5 minutes)
```bash
cd /home/user/less.go
git fetch origin
git checkout -b claude/fix-{taskname}-{yourID}

# Get baseline numbers
pnpm -w test:go:unit | tail -5          # Should see: PASS
LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | grep "Perfect CSS"  # Should see: 84
```

### 2. **Make Your Changes** (varies)
- Edit Go files in `packages/less/src/less/less_go/`
- Follow patterns from working code
- Compare with JavaScript implementation
- Use LESS_GO_DIFF to see differences: `LESS_GO_DIFF=1 pnpm -w test:go`

### 3. **Test Incrementally**
```bash
# Test your specific fix
pnpm -w test:go:filter -- "your-test-name"

# Check for unit test regressions
pnpm -w test:go:unit

# See full impact
LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | tail -30
```

### 4. **Verify No Regressions** (CRITICAL!)
```bash
# These MUST pass before committing:
pnpm -w test:go:unit          # MUST: 2,304 passing
LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | grep "Perfect CSS"  # MUST: >= 84
```

### 5. **Commit & Push**
```bash
git add -A
git commit -m "Fix {feature}: Brief description of what changed"
git push origin claude/fix-{taskname}-{yourID}
```

### 6. **Create PR**
- Use clear title: "Fix {feature}: What was broken and how you fixed it"
- Reference the baseline metrics in PR description
- Link to original prompt if applicable

---

## 📍 Files You'll Need to Know

```
less.go/
├── .claude/
│   ├── AGENT_PROMPTS_2025-11-26.md  ← READ THIS! Full task descriptions
│   ├── SESSION_SUMMARY_2025-11-26.md ← Context for current state
│   ├── QUICK_START_AGENT_GUIDE.md    ← You are here
│   ├── strategy/
│   │   └── agent-workflow.md          ← Detailed workflow
│   └── tasks/
│       ├── runtime-failures/          ← Info on failing tests
│       └── error-handling/            ← Info on error tests
│
├── packages/less/src/less/less_go/    ← WHERE YOU MAKE CHANGES
│   ├── import.go, import_visitor.go   ← For task 1 (import-reference)
│   ├── detached_ruleset.go, media.go  ← For task 2 (detached-rulesets)
│   ├── url.go, ruleset.go             ← For task 3 (urls)
│   ├── media.go                       ← For task 4 (media formatting)
│   ├── at_rule.go                     ← For task 5 (container queries)
│   └── (other files as needed)
│
└── packages/test-data/
    ├── less/_main/                    ← Test input files
    └── css/_main/                     ← Expected output files
```

---

## 🎯 How to Pick Your Task

### If you like feature work:
→ Pick **Task 1-6** (feature/fix tasks)

### If you like understanding code:
→ Pick **Task 7** (analysis) or **Task 9** (performance)

### If you like quick wins:
→ Pick **Task 4** (media formatting), **Task 8** (error), or **Task 10** (cleanup)

### If you like learning the codebase:
→ Pick **Task 3** (urls) or **Task 6** (directive bubbling) - most educational

---

## 🔍 Example: How Task 1 Works

### What It Is
Fix import-reference to suppress CSS output from imported files with `(reference)` option

### Get the Full Details
```bash
cat .claude/AGENT_PROMPTS_2025-11-26.md | grep -A 100 "Prompt 1:"
```

### What Success Looks Like
```bash
$ LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | grep -A 2 "Perfect CSS"
✅ Perfect CSS Matches: 86 (46.7%)  # ← Was 84, now 86 (+2!)
```

---

## ⚠️ Golden Rules (DON'T BREAK THESE!)

1. **ALWAYS check baseline before starting**
   ```bash
   pnpm -w test:go:unit          # Must pass
   LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | tail -30  # Note the numbers
   ```

2. **NEVER let unit tests fail**
   - If `pnpm -w test:go:unit` shows failures, you've broken something
   - Fix it before committing

3. **NEVER reduce perfect match count**
   - If 84 goes down to 83, you broke a previously working test
   - Don't commit until it's back to 84+

4. **ALWAYS test incrementally**
   - Don't make 10 changes and test once
   - Make 1-2 changes, test, verify, commit

5. **ALWAYS read the JavaScript version**
   - When confused about behavior, check JavaScript implementation
   - Match it exactly in Go

6. **ALWAYS include regression check in your process**
   - Before every commit: verify metrics haven't regressed
   - This is non-negotiable

---

## 🆘 If Things Go Wrong

### Unit tests failing?
```bash
# See what's failing
pnpm -w test:go:unit 2>&1 | grep -A 5 "FAIL"

# Most likely: You changed something that broke existing functionality
# Solution: Review your changes, understand what you modified
```

### Perfect match count dropped?
```bash
# See which test(s) broke
LESS_GO_DIFF=1 pnpm -w test:go 2>&1 | grep -B 3 "Output Differs"

# Solution: Revert your last change, understand why it broke something else
```

### Don't know what to fix?
```bash
# Look at the test data
cat packages/test-data/less/_main/import-reference.less  # Input
cat packages/test-data/css/_main/import-reference.css    # Expected output

# Compare with actual
LESS_GO_DIFF=1 pnpm -w test:go:filter -- "import-reference"
```

---

## 📚 Learning Resources in This Repo

### To understand how a feature works:
1. Read the test data: `packages/test-data/less/_main/{feature}.less`
2. See expected output: `packages/test-data/css/_main/{feature}.css`
3. Look at JavaScript: `packages/less/src/less/tree/{feature}.js`
4. Look at Go: `packages/less/src/less/less_go/{feature}.go`
5. Compare implementations line by line

### To debug a failing test:
1. Run with diffs: `LESS_GO_DIFF=1 pnpm -w test:go:filter -- "test-name"`
2. Understand what's different between expected and actual
3. Use LESS_GO_TRACE for detailed execution: `LESS_GO_TRACE=1 pnpm -w test:go:filter -- "test-name"`
4. Add temporary debug prints in Go code if needed

### To understand task context:
- Read the prompt in `AGENT_PROMPTS_2025-11-26.md`
- Check for related task files in `.claude/tasks/`
- Review JavaScript implementation for that feature

---

## ✨ Expected Outcomes

### For You (Agent)
- ✅ Learn how less.go port works
- ✅ Understand Go patterns for AST processing
- ✅ Get hands-on with test-driven development
- ✅ See your changes improve the overall project

### For the Project
- ✅ Fix 6-8 tests (if you do tasks 1-6)
- ✅ Increase success rate from 45.7% → ~50%
- ✅ Get one step closer to production readiness
- ✅ Maintain zero regressions

---

## 🚀 Ready to Start?

1. **Read full prompt**: `.claude/AGENT_PROMPTS_2025-11-26.md`
2. **Pick a task**: Any of prompts 1-10
3. **Follow the workflow**: Start → Fix → Test → Verify → Commit → PR
4. **Check baseline**: Always before and after
5. **Submit PR**: With clear description of changes

**You've got this!** The codebase is well-structured, tests are clear, and the prompts have everything you need. 🎉

---

Last Updated: 2025-11-26
Questions? Check the task-specific prompt or `.claude/strategy/agent-workflow.md`
