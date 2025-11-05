# Start Independent Agents - Quick Guide

## 🚀 Launch All 5 Agents in Parallel

Each agent has ONE focused task. Launch as many in parallel as you have capacity for!

### Setup (One Time)

```bash
# Clone repo 5 times (or use git worktrees)
cd ~/projects  # or wherever you want

git clone <your-repo-url> less.go-agent-urls
git clone <your-repo-url> less.go-agent-paths
git clone <your-repo-url> less.go-agent-namespacing-6
git clone <your-repo-url> less.go-agent-import-reference
git clone <your-repo-url> less.go-agent-import-reference-issues
```

---

## Agent 1: URLs (2 tests, same fix)

```bash
cd ~/projects/less.go-agent-urls
```

**Copy and paste this into Claude Code:**

```
Fix URL parsing to handle escaped characters in less.go

You're working on a Go port of less.js. URLs with escaped parentheses like \( cause parser errors.

YOUR SINGLE TASK:
- Fix 2 tests (same fix): urls (main + compression suites)
- Error: "expected ')' got '('" when parsing url(http://example.com?family=\"Font\":\(400\))
- Files to check: url.go, parser.go (line ~3270), parser_input.go
- Branch: claude/fix-urls-<your-session-id>

Read .claude/agents/agent-urls/TASK.md for full details.

Start by running:
cd packages/less/src/less/less_go
go test -run "TestIntegrationSuite/main/urls" -v

The regex at parser.go:3286 already tries to match escaped chars \\[()'"], but something's not working. Find and fix it.

CONSTRAINTS:
- Never modify .js files
- Must pass: pnpm -w test:go:unit
- Must pass both urls tests
- Create branch, commit, push when done

SUCCESS: Report "Fixed 2/2 URL tests. Ready for PR."
```

---

## Agent 2: Paths (1 test)

```bash
cd ~/projects/less.go-agent-paths
```

**Copy and paste this into Claude Code:**

```
Fix include path resolution for imports in less.go

You're working on a Go port of less.js. Imports aren't being resolved through configured include paths.

YOUR SINGLE TASK:
- Fix 1 test ONLY: include-path
- Error: "open import-test-e: no such file or directory"
- Files to check: integration_suite_test.go, import_manager.go, file_manager.go
- Branch: claude/fix-paths-<your-session-id>

Read .claude/agents/agent-paths/TASK.md for full details.

Start by finding where the file actually is:
find packages/test-data -name "*import-test-e*"

Then check if the test configures include paths:
grep -A10 "include-path" packages/less/src/less/less_go/integration_suite_test.go

Likely need to either:
A) Add include path config to test, OR
B) Fix import_manager.go to search include paths, OR
C) Both

CONSTRAINTS:
- Never modify .js files
- Must pass: pnpm -w test:go:unit
- Must pass include-path test ONLY
- Create branch, commit, push when done

SUCCESS: Report "Fixed include-path test. Ready for PR."
```

---

## Agent 3: Namespacing-6 (1 test)

```bash
cd ~/projects/less.go-agent-namespacing-6
```

**Copy and paste this into Claude Code:**

```
Fix namespacing-6 test in less.go

You're working on a Go port of less.js. Mixin calls assigned to variables fail when called.

YOUR SINGLE TASK:
- Fix 1 test ONLY: namespacing-6
- Error: "Could not evaluate variable call @alias"
- The code: @alias: .something(foo); @alias(); ← this fails
- Files to check: variable_call.go, variable.go, mixin_call.go
- Branch: claude/fix-namespacing-6-<your-session-id>

Read .claude/agents/agent-namespacing-6/TASK.md for full details.

THE PROBLEM:
When you assign a mixin call to a variable and then call it:
  @alias: .something(foo);  // Store mixin result
  @alias();                 // Call it - FAILS

Start with trace:
cd packages/less/src/less/less_go
LESS_GO_TRACE=1 go test -run "TestIntegrationSuite/namespacing/namespacing-6" -v 2>&1 | grep -i alias

Look in variable_call.go for "Could not evaluate variable call" error.

HINT: Similar to Issue #2 (detached-rulesets) which was fixed by checking Eval(any) (any, error) signature before Eval(any) any. You might need a similar pattern.

CONSTRAINTS:
- Never modify .js files
- Must pass: pnpm -w test:go:unit
- Must pass namespacing-6 test ONLY
- Create branch, commit, push when done

SUCCESS: Report "Fixed namespacing-6 test. Ready for PR."
```

---

## Agent 4: Import Reference (1 test)

```bash
cd ~/projects/less.go-agent-import-reference
```

**Copy and paste this into Claude Code:**

```
Fix import-reference test in less.go

You're working on a Go port of less.js. CSS imports are being processed instead of kept as @import statements.

YOUR SINGLE TASK:
- Fix 1 test ONLY: import-reference
- Error: "open test.css: no such file or directory"
- Files to check: import_visitor.go, import_manager.go, import.go
- Branch: claude/fix-import-reference-<your-session-id>

Read .claude/agents/agent-import-reference/TASK.md for full details.

THE PROBLEM:
When importing CSS files, they should remain as @import statements in output, NOT be loaded and processed as LESS files.

Test code has:
  @import (reference) url("import-once.less");
  @import (reference) url("css-3.less");

But it's trying to open ".css" files and process them.

Start by testing:
cd packages/less/src/less/less_go
go test -run "TestIntegrationSuite/main/import-reference" -v

KEY FIX NEEDED:
In import_manager.go, detect .css file extension and keep as @import statement instead of processing.

CONSTRAINTS:
- Never modify .js files
- Must pass: pnpm -w test:go:unit
- Must pass import-reference test ONLY
- Create branch, commit, push when done

SUCCESS: Report "Fixed import-reference test. Ready for PR."
```

---

## Agent 5: Import Reference Issues (1 test)

```bash
cd ~/projects/less.go-agent-import-reference-issues
```

**Copy and paste this into Claude Code:**

```
Fix import-reference-issues test in less.go

You're working on a Go port of less.js. Referenced imports aren't making mixins accessible.

YOUR SINGLE TASK:
- Fix 1 test ONLY: import-reference-issues
- Error: "#Namespace > .mixin is undefined"
- Files to check: import_visitor.go, import.go, set_tree_visibility_visitor.go
- Branch: claude/fix-import-reference-issues-<your-session-id>

Read .claude/agents/agent-import-reference-issues/TASK.md for full details.

THE PROBLEM:
@import (reference) should:
- Make mixins/variables available for use
- NOT output the imported CSS by default
- Only output CSS for explicitly used selectors

But currently referenced mixins are not accessible.

Start by testing:
cd packages/less/src/less/less_go
go test -run "TestIntegrationSuite/main/import-reference-issues" -v

KEY CONCEPT:
The `reference` flag should mark imports so their mixins are accessible but CSS isn't output unless explicitly used (via extend or mixin call).

CONSTRAINTS:
- Never modify .js files
- Must pass: pnpm -w test:go:unit
- Must pass import-reference-issues test ONLY
- Create branch, commit, push when done

SUCCESS: Report "Fixed import-reference-issues test. Ready for PR."
```

---

## ✅ What Happens Next

Each agent will:
1. Read their detailed TASK.md
2. Investigate their single issue
3. Make their fix
4. Test thoroughly
5. Create branch: `claude/fix-<issue>-<session-id>`
6. Commit with clear message
7. Push their branch
8. Report: "Fixed <test-name> test. Ready for PR."

---

## 📊 After All Agents Complete

```bash
# Check results from main repo
cd ~/projects/less.go  # your main repo

# See all branches
git fetch --all
git branch -r | grep fix-

# Review each branch
git log origin/claude/fix-urls-xxxxx --oneline -5
git log origin/claude/fix-paths-xxxxx --oneline -5
git log origin/claude/fix-namespacing-6-xxxxx --oneline -5
git log origin/claude/fix-import-reference-xxxxx --oneline -5
git log origin/claude/fix-import-reference-issues-xxxxx --oneline -5

# Merge or create PRs as desired
```

Expected results:
- **6 tests fixed** (2 URLs + 1 path + 1 namespacing + 2 imports)
- **Pass rate**: 38% → 41-43%
- **5 independent branches** ready for review

---

## 🎓 Alternative: Use Git Worktrees

If you prefer worktrees over clones:

```bash
cd ~/projects/less.go  # your main repo

git worktree add ../less.go-urls -b fix-urls-temp
git worktree add ../less.go-paths -b fix-paths-temp
git worktree add ../less.go-namespacing-6 -b fix-namespacing-6-temp
git worktree add ../less.go-import-reference -b fix-import-reference-temp
git worktree add ../less.go-import-reference-issues -b fix-import-reference-issues-temp

# Now open Claude Code in each worktree and use the prompts above
```

---

## 📋 Agent Independence Matrix

| Agent | Tests | Files | Conflicts With |
|-------|-------|-------|----------------|
| URLs | 2 | url.go, parser.go | None ✅ |
| Paths | 1 | integration_suite_test.go, import_manager.go | Import agents (different sections) |
| Namespacing-6 | 1 | variable_call.go, variable.go, mixin_call.go | None ✅ |
| Import Ref | 1 | import_manager.go, import_visitor.go | Paths (different sections) |
| Import Ref Issues | 1 | import_visitor.go, import.go | None ✅ |

**All agents can run in parallel!** The potential conflicts are in different sections of the same files.

---

## 🎯 Quick Start Tips

**Start with these 3 (Zero conflicts)**:
1. URLs
2. Namespacing-6
3. Import Reference Issues

**Then add these 2 (Minor overlaps)**:
4. Paths
5. Import Reference

Or just launch all 5 at once - they're aware of each other!

---

**Ready? Copy the 5 prompts above into 5 separate Claude Code sessions and watch them work! 🚀**
