# Start Independent Agents - Quick Guide

## 🚀 Launch All 4 Agents in Parallel

### Setup (One Time)

```bash
# Clone repo 4 times (or use git worktrees)
cd ~/projects  # or wherever you want

git clone <your-repo-url> less.go-agent-urls
git clone <your-repo-url> less.go-agent-paths
git clone <your-repo-url> less.go-agent-namespacing
git clone <your-repo-url> less.go-agent-imports
```

---

## Agent 1: URLs (Terminal 1)

```bash
cd ~/projects/less.go-agent-urls
```

**Copy and paste this into Claude Code:**

```
Fix URL parsing to handle escaped characters in less.go

You're working on a Go port of less.js. URLs with escaped parentheses like \( cause parser errors.

YOUR TASK:
- Fix 2 failing tests: urls (main + compression suites)
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
```

---

## Agent 2: Paths (Terminal 2)

```bash
cd ~/projects/less.go-agent-paths
```

**Copy and paste this into Claude Code:**

```
Fix include path resolution for imports in less.go

You're working on a Go port of less.js. Imports aren't being resolved through configured include paths.

YOUR TASK:
- Fix 1 failing test: include-path
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
- Create branch, commit, push when done

NOTE: agent-imports also touches import_manager.go but different sections. You focus on path searching, they focus on reference flags.
```

---

## Agent 3: Namespacing (Terminal 3)

```bash
cd ~/projects/less.go-agent-namespacing
```

**Copy and paste this into Claude Code:**

```
Fix namespace variable call evaluation in less.go

You're working on a Go port of less.js. Mixin calls assigned to variables fail when called.

YOUR TASK:
- Fix 2 failing tests: namespacing-6, namespacing-functions
- Error: "Could not evaluate variable call @alias"
- Code: @alias: .something(foo); @alias(); ← fails
- Files to check: variable_call.go, variable.go, mixin_call.go, detached_ruleset.go
- Branch: claude/fix-namespacing-<your-session-id>

Read .claude/agents/agent-namespacing/TASK.md for full details.

Start with trace to understand the flow:
cd packages/less/src/less/less_go
LESS_GO_TRACE=1 go test -run "TestIntegrationSuite/namespacing/namespacing-6" -v 2>&1 | grep -i alias

Look for "Could not evaluate variable call" in variable_call.go - that's where it fails.

Similar to Issue #2 (detached-rulesets) which was fixed by checking Eval(any) (any, error) before Eval(any) any. Your issue might need similar pattern.

CONSTRAINTS:
- Never modify .js files
- Must pass: pnpm -w test:go:unit
- Must pass both namespacing tests
- Create branch, commit, push when done
```

---

## Agent 4: Imports (Terminal 4)

```bash
cd ~/projects/less.go-agent-imports
```

**Copy and paste this into Claude Code:**

```
Fix import reference functionality in less.go

You're working on a Go port of less.js. Import reference handling has bugs.

YOUR TASK:
- Fix 2-3 of 5 import tests
- Target: import-reference, import-reference-issues, (stretch: import-module)
- DEFER: import-interpolation (architectural), google (network issue)
- Files to check: import_visitor.go, import_manager.go, import.go, set_tree_visibility_visitor.go
- Branch: claude/fix-imports-<your-session-id>

Read .claude/agents/agent-imports/TASK.md for full details.

KEY ISSUES:
1. import-reference: CSS files being processed instead of kept as @import statements
2. import-reference-issues: Referenced mixins not accessible (#Namespace > .mixin undefined)

THE CONCEPT:
@import (reference) should:
- Make mixins/variables available
- NOT output the imported CSS by default
- Only output explicitly used selectors

Start by testing:
cd packages/less/src/less/less_go
go test -run "TestIntegrationSuite/main/import-reference" -v

CONSTRAINTS:
- Never modify .js files
- Must pass: pnpm -w test:go:unit
- Create branch, commit, push when done

NOTE: agent-paths also touches import_manager.go but they focus on path searching. You focus on reference flags and CSS detection. Different sections.
```

---

## ✅ What Happens Next

Each agent will:
1. Read their detailed TASK.md
2. Investigate the issue
3. Make their fixes
4. Test thoroughly
5. Create branch: `claude/fix-<issue>-<session-id>`
6. Commit with clear message
7. Push their branch
8. Report: "Fixed X/Y tests. Ready for PR."

---

## 📊 After All Agents Complete

```bash
# Check results from main repo
cd ~/projects/less.go  # your main repo

# See all branches
git fetch --all
git branch -r | grep fix-

# Review each branch
git checkout origin/claude/fix-urls-xxxxx
git checkout origin/claude/fix-paths-xxxxx
git checkout origin/claude/fix-namespacing-xxxxx
git checkout origin/claude/fix-imports-xxxxx

# Merge or create PRs as desired
```

Expected results:
- **7-9 tests fixed** (of 12-13 runtime failures)
- **Pass rate**: 38% → 43-47%
- **4 independent branches** ready for review

---

## 🎓 Alternative: Use Git Worktrees

If you prefer worktrees over clones:

```bash
cd ~/projects/less.go  # your main repo

git worktree add ../less.go-urls -b fix-urls-temp
git worktree add ../less.go-paths -b fix-paths-temp
git worktree add ../less.go-namespacing -b fix-namespacing-temp
git worktree add ../less.go-imports -b fix-imports-temp

# Now open Claude Code in each worktree and use the prompts above
```

---

**Ready? Copy the 4 prompts above into 4 separate Claude Code sessions and watch them work! 🚀**
