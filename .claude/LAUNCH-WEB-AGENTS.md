# Launch Web-Based Parallel Agents

## 🌐 For Claude Code Web Interface

This guide is optimized for running multiple agents using the **web version of Claude Code** in parallel.

---

## 🚀 Super Simple Launch Process

### Step 1: Open 5 Claude Code Tabs

Open **5 tabs** of Claude Code web interface pointing to your repo:
- https://claude.ai/code (or your Claude Code web URL)
- Each will automatically access the `toakleaf/less.go` repository

### Step 2: Copy-Paste Kickoff Prompts

In each tab, paste one of the prompts below. That's it!

---

## 📋 The 5 Kickoff Prompts

### Agent 1: URLs (2 tests) - Tab 1

```
Fix URL parsing with escaped characters in less.go

CONTEXT: You're working on a Go port of less.js at https://github.com/toakleaf/less.go. The parser is 92% working but URLs with escaped parentheses cause errors.

YOUR SINGLE TASK: Fix URL parsing for 2 tests (same fix applies to both)

FAILING TESTS:
- urls (main suite)
- urls (compression suite)

ERROR: "expected ')' got '('" when parsing:
  url(http://fonts.googleapis.com/css?family=\"Rokkitt\":\(400\),700)

THE BUG: The parser sees \( as an opening paren instead of escaped character.

WORKFLOW:
1. Read .claude/agents/agent-urls/TASK.md for details
2. Test the issue:
   cd packages/less/src/less/less_go
   go test -run "TestIntegrationSuite/main/urls" -v
3. Fix it (likely in url.go or parser.go around line 3270)
4. Verify fix:
   go test -run "TestIntegrationSuite/main/urls" -v
   go test -run "TestIntegrationSuite/compression/urls" -v
   pnpm -w test:go:unit
5. Create branch: claude/fix-urls-<your-session-id>
6. Commit with clear message
7. Push: git push -u origin claude/fix-urls-<your-session-id>
8. Create PR:
   gh pr create --title "Fix URL parsing with escaped characters" \
     --body "Fixes URL parsing to handle escaped parentheses like \( and \).

Tests fixed:
- urls (main suite)
- urls (compression suite)

The parser was treating escaped characters as syntax. Updated escape handling to treat backslash-escaped characters as literals.

Closes #[issue-number-if-exists]"

CONSTRAINTS:
- Never modify .js files
- Must pass: pnpm -w test:go:unit
- No regressions

SUCCESS: Report "✅ Fixed 2 URL tests. PR created: [URL]"
```

---

### Agent 2: Paths (1 test) - Tab 2

```
Fix include path resolution for imports in less.go

CONTEXT: You're working on a Go port of less.js at https://github.com/toakleaf/less.go. Imports aren't being resolved through configured include paths.

YOUR SINGLE TASK: Fix include path resolution for 1 test

FAILING TEST:
- include-path

ERROR: "open import-test-e: no such file or directory"

THE BUG: Files should be searched in configured include paths, but they're not.

WORKFLOW:
1. Read .claude/agents/agent-paths/TASK.md for details
2. Find where the file actually is:
   find packages/test-data -name "*import-test-e*"
3. Check test configuration:
   grep -A10 "include-path" packages/less/src/less/less_go/integration_suite_test.go
4. Fix (likely need to either):
   - Add include path config to test, OR
   - Fix import_manager.go to search include paths
5. Verify fix:
   go test -run "TestIntegrationSuite.*include-path" -v
   pnpm -w test:go:unit
6. Create branch: claude/fix-include-path-<your-session-id>
7. Commit with clear message
8. Push: git push -u origin claude/fix-include-path-<your-session-id>
9. Create PR:
   gh pr create --title "Fix include path resolution for imports" \
     --body "Imports now search configured include paths when files aren't found relative to current file.

Test fixed:
- include-path

Added include path configuration and/or fixed import manager to properly search configured paths.

Closes #[issue-number-if-exists]"

CONSTRAINTS:
- Never modify .js files
- Must pass: pnpm -w test:go:unit
- No regressions

SUCCESS: Report "✅ Fixed include-path test. PR created: [URL]"
```

---

### Agent 3: Namespacing-6 (1 test) - Tab 3

```
Fix namespacing-6 test in less.go

CONTEXT: You're working on a Go port of less.js at https://github.com/toakleaf/less.go. Mixin calls assigned to variables fail when called.

YOUR SINGLE TASK: Fix variable calls for 1 test

FAILING TEST:
- namespacing-6

ERROR: "Could not evaluate variable call @alias"

THE BUG: When you do @alias: .something(foo); @alias(); ← the call fails

WORKFLOW:
1. Read .claude/agents/agent-namespacing-6/TASK.md for details
2. Use trace to understand flow:
   cd packages/less/src/less/less_go
   LESS_GO_TRACE=1 go test -run "TestIntegrationSuite/namespacing/namespacing-6" -v 2>&1 | grep -i alias
3. Find "Could not evaluate variable call" error in variable_call.go
4. Fix it (similar to Issue #2 - check Eval(any) (any, error) before Eval(any) any)
5. Verify fix:
   go test -run "TestIntegrationSuite/namespacing/namespacing-6" -v
   pnpm -w test:go:unit
6. Create branch: claude/fix-namespacing-6-<your-session-id>
7. Commit with clear message
8. Push: git push -u origin claude/fix-namespacing-6-<your-session-id>
9. Create PR:
   gh pr create --title "Fix namespacing-6: variable calls to mixin results" \
     --body "Fixed evaluation of mixin calls assigned to variables.

Test fixed:
- namespacing-6

When mixin calls like .something(foo) are assigned to variables and then called as @alias(), the evaluation was failing. Fixed [describe your fix].

Closes #[issue-number-if-exists]"

CONSTRAINTS:
- Never modify .js files
- Must pass: pnpm -w test:go:unit
- No regressions

SUCCESS: Report "✅ Fixed namespacing-6 test. PR created: [URL]"
```

---

### Agent 4: Import Reference (1 test) - Tab 4

```
Fix import-reference test in less.go

CONTEXT: You're working on a Go port of less.js at https://github.com/toakleaf/less.go. CSS imports are being processed instead of kept as @import statements.

YOUR SINGLE TASK: Fix CSS import handling for 1 test

FAILING TEST:
- import-reference

ERROR: "open test.css: no such file or directory"

THE BUG: CSS files should remain as @import statements, NOT be loaded/processed as LESS.

WORKFLOW:
1. Read .claude/agents/agent-import-reference/TASK.md for details
2. Test the issue:
   cd packages/less/src/less/less_go
   go test -run "TestIntegrationSuite/main/import-reference" -v
3. Fix it (likely in import_manager.go - detect .css extension)
4. Verify fix:
   go test -run "TestIntegrationSuite/main/import-reference" -v
   pnpm -w test:go:unit
5. Create branch: claude/fix-import-reference-<your-session-id>
6. Commit with clear message
7. Push: git push -u origin claude/fix-import-reference-<your-session-id>
8. Create PR:
   gh pr create --title "Fix import-reference: handle CSS imports correctly" \
     --body "CSS files now remain as @import statements instead of being processed as LESS.

Test fixed:
- import-reference

Added CSS file detection in import manager to keep .css imports as-is in output.

Closes #[issue-number-if-exists]"

CONSTRAINTS:
- Never modify .js files
- Must pass: pnpm -w test:go:unit
- No regressions

SUCCESS: Report "✅ Fixed import-reference test. PR created: [URL]"
```

---

### Agent 5: Import Reference Issues (1 test) - Tab 5

```
Fix import-reference-issues test in less.go

CONTEXT: You're working on a Go port of less.js at https://github.com/toakleaf/less.go. Referenced imports aren't making mixins accessible.

YOUR SINGLE TASK: Fix referenced mixin accessibility for 1 test

FAILING TEST:
- import-reference-issues

ERROR: "#Namespace > .mixin is undefined"

THE BUG: @import (reference) should make mixins available but NOT output CSS by default.

WORKFLOW:
1. Read .claude/agents/agent-import-reference-issues/TASK.md for details
2. Test the issue:
   cd packages/less/src/less/less_go
   go test -run "TestIntegrationSuite/main/import-reference-issues" -v
3. Fix it (likely in import_visitor.go - preserve reference flag, add to frames)
4. Verify fix:
   go test -run "TestIntegrationSuite/main/import-reference-issues" -v
   pnpm -w test:go:unit
5. Create branch: claude/fix-import-reference-issues-<your-session-id>
6. Commit with clear message
7. Push: git push -u origin claude/fix-import-reference-issues-<your-session-id>
8. Create PR:
   gh pr create --title "Fix import-reference-issues: make referenced mixins accessible" \
     --body "Referenced imports now properly make mixins accessible while preventing default CSS output.

Test fixed:
- import-reference-issues

Fixed import visitor to add referenced rulesets to frames so mixins are accessible, while marking them to prevent CSS output unless explicitly used.

Closes #[issue-number-if-exists]"

CONSTRAINTS:
- Never modify .js files
- Must pass: pnpm -w test:go:unit
- No regressions

SUCCESS: Report "✅ Fixed import-reference-issues test. PR created: [URL]"
```

---

## ✅ What Happens Next

Each agent will:
1. Pull the repository (automatic in web Claude Code)
2. Read their TASK.md for detailed context
3. Investigate and fix their single issue
4. Run tests to verify
5. Create a branch: `claude/fix-<issue>-<session-id>`
6. Commit with clear message
7. Push their branch
8. **Create a PR using `gh` CLI**
9. Report completion with PR URL

---

## 📊 After All Agents Complete

You'll have **5 PRs** to review:
1. PR for URL parsing fix (2 tests)
2. PR for include path fix (1 test)
3. PR for namespacing-6 fix (1 test)
4. PR for import-reference fix (1 test)
5. PR for import-reference-issues fix (1 test)

**Total: 6 tests fixed**

---

## 🎯 Benefits of This Approach

✅ **No manual setup** - Web Claude Code handles repo access
✅ **True parallelism** - All 5 agents work simultaneously
✅ **Independent PRs** - Easy to review and merge individually
✅ **No conflicts** - Each agent works on different issues
✅ **Clean workflow** - Paste prompt → Agent works → PR created

---

## 💡 Pro Tips

**Start with 2-3 first** to validate the approach:
- Agent 1 (URLs) - Simplest
- Agent 3 (Namespacing-6) - Medium
- Agent 5 (Import Reference Issues) - Independent

Then launch the remaining 2 once you see it working!

**Monitor progress**: Each agent will report when they create their PR

**Review individually**: You can merge PRs as they complete, no need to wait for all 5

---

**Ready? Open 5 tabs and paste the prompts above! 🚀**
