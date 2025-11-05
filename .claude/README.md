# Independent Multi-Agent Orchestration

This directory contains everything needed to run **multiple independent Claude Code agents in parallel** to fix the remaining integration test failures.

---

## 🌐 **Quick Start: Web Claude Code** (Recommended)

**See [`LAUNCH-WEB-AGENTS.md`](LAUNCH-WEB-AGENTS.md)** for the simplest approach:
1. Open 5 tabs of Claude Code web
2. Paste kickoff prompts
3. Watch agents create PRs
4. Done!

---

## 📊 Current Status

- **Tests Passing**: 71/185 (38.4%)
- **Tests to Fix**: 114 (12-13 runtime failures + 102 output differences)
- **Parser Status**: ✅ 92.4% compilation rate - parser is working!
- **Focus**: Runtime evaluation and CSS output generation

## 📁 Directory Structure

```
.claude/
├── README.md                          ← You are here
├── LAUNCH-WEB-AGENTS.md              ← START HERE for web Claude Code
├── agents/                            ← Independent agent tasks
│   ├── agent-urls/
│   │   ├── TASK.md                   ← Detailed task description
│   │   └── KICKOFF.txt               ← Quick start prompt
│   ├── agent-paths/
│   ├── agent-namespacing-6/
│   ├── agent-import-reference/
│   └── agent-import-reference-issues/
└── reference-issues/                  ← Original detailed issue analysis
    ├── ISSUE_IMPORTS.md
    ├── ISSUE_NAMESPACING.md
    ├── ISSUE_URLS.md
    └── ...
```

## 🚀 How to Run Multiple Agents in Parallel

### Method 1: Multiple Claude Code Sessions (Recommended)

1. **Open 5 separate terminals** (or 5 Claude Code sessions)

2. **In each terminal, checkout the repo in a separate location**:
   ```bash
   # Terminal 1
   git clone <repo-url> less.go-agent-urls
   cd less.go-agent-urls

   # Terminal 2
   git clone <repo-url> less.go-agent-paths
   cd less.go-agent-paths

   # Terminal 3
   git clone <repo-url> less.go-agent-namespacing-6
   cd less.go-agent-namespacing-6

   # Terminal 4
   git clone <repo-url> less.go-agent-import-reference
   cd less.go-agent-import-reference

   # Terminal 5
   git clone <repo-url> less.go-agent-import-reference-issues
   cd less.go-agent-import-reference-issues
   ```

3. **Start Claude Code in each** and give each the KICKOFF prompt:
   ```bash
   # Terminal 1 - Copy/paste content of:
   cat .claude/agents/agent-urls/KICKOFF.txt

   # Terminal 2 - Copy/paste content of:
   cat .claude/agents/agent-paths/KICKOFF.txt

   # Terminal 3 - Copy/paste content of:
   cat .claude/agents/agent-namespacing-6/KICKOFF.txt

   # Terminal 4 - Copy/paste content of:
   cat .claude/agents/agent-import-reference/KICKOFF.txt

   # Terminal 5 - Copy/paste content of:
   cat .claude/agents/agent-import-reference-issues/KICKOFF.txt
   ```

4. **Each agent will**:
   - Read their TASK.md for full details
   - Work on their **single focused issue**
   - Create a branch: `claude/fix-<issue>-<session-id>`
   - Commit their fix
   - Push their branch
   - Report completion

5. **After all agents complete**, review branches and create PRs or merge to main.

### Method 2: Git Worktrees (Advanced)

If you want to use a single repo with worktrees:

```bash
# Create worktrees for each agent
git worktree add ../less.go-urls claude/fix-urls-temp
git worktree add ../less.go-paths claude/fix-paths-temp
git worktree add ../less.go-namespacing-6 claude/fix-namespacing-6-temp
git worktree add ../less.go-import-reference claude/fix-import-reference-temp
git worktree add ../less.go-import-reference-issues claude/fix-import-reference-issues-temp

# Start Claude Code in each worktree
# Give each agent their KICKOFF prompt
```

## 🎯 Agent Overview

| Agent | Tests | Files Modified | Independence | Priority |
|-------|-------|----------------|--------------|----------|
| **agent-urls** | 2 (same fix) | url.go, parser.go | HIGH ✅ | High |
| **agent-paths** | 1 | integration_suite_test.go, import_manager.go | MEDIUM ⚠️ | Medium |
| **agent-namespacing-6** | 1 | variable_call.go, variable.go, mixin_call.go | HIGH ✅ | High |
| **agent-import-reference** | 1 | import_manager.go, import_visitor.go | MEDIUM ⚠️ | High |
| **agent-import-reference-issues** | 1 | import_visitor.go, import.go | HIGH ✅ | High |

### Independence Notes

✅ **HIGH** = No file conflicts with other agents
⚠️ **MEDIUM** = Potential overlap but different sections:
- agent-paths: Include path searching in import_manager.go
- agent-import-reference: CSS detection in import_manager.go
- agent-import-reference-issues: Reference visibility in import_visitor.go

All agents are aware of each other and work on different aspects.

## 📋 Quick Start for Each Agent

### Agent 1: URLs (2 tests, same fix)
**What**: Fix URL parsing with escaped characters
**Files**: url.go, parser.go
**Tests**: 2 (urls in main + compression)
**Kickoff**: `.claude/agents/agent-urls/KICKOFF.txt`
**Branch**: `claude/fix-urls-<session-id>`

### Agent 2: Paths (1 test - Quick Win)
**What**: Fix include path resolution
**Files**: integration_suite_test.go, import_manager.go
**Tests**: 1 (include-path)
**Kickoff**: `.claude/agents/agent-paths/KICKOFF.txt`
**Branch**: `claude/fix-paths-<session-id>`

### Agent 3: Namespacing-6 (1 test)
**What**: Fix variable calls to mixin results
**Files**: variable_call.go, variable.go, mixin_call.go
**Tests**: 1 (namespacing-6)
**Kickoff**: `.claude/agents/agent-namespacing-6/KICKOFF.txt`
**Branch**: `claude/fix-namespacing-6-<session-id>`

### Agent 4: Import Reference (1 test)
**What**: Fix CSS import handling
**Files**: import_manager.go, import_visitor.go
**Tests**: 1 (import-reference)
**Kickoff**: `.claude/agents/agent-import-reference/KICKOFF.txt`
**Branch**: `claude/fix-import-reference-<session-id>`

### Agent 5: Import Reference Issues (1 test)
**What**: Make referenced mixins accessible
**Files**: import_visitor.go, import.go
**Tests**: 1 (import-reference-issues)
**Kickoff**: `.claude/agents/agent-import-reference-issues/KICKOFF.txt`
**Branch**: `claude/fix-import-reference-issues-<session-id>`

## ✅ Success Criteria

Each agent should:
- [ ] Fix their assigned test(s) - ONE focused task
- [ ] Pass all unit tests: `pnpm -w test:go:unit`
- [ ] Not break any currently passing tests
- [ ] Commit to their branch with clear message
- [ ] Push their branch
- [ ] Report: "Fixed <test-name> test. Ready for PR."

## 📊 Expected Results

After all 5 agents complete:
- **Tests Fixed**: 6 tests (2 URLs + 1 path + 1 namespacing + 2 imports)
- **Pass Rate**: 38.4% → 41-43%
- **Branches**: 5 independent branches ready for review/merge

## 🔄 After Wave 1

Once these 4 agents complete and their fixes are merged:

1. **Review results**: How many tests fixed? Any issues?
2. **Run full test suite**: `pnpm -w test:go:summary`
3. **Next wave**: Consider additional agents for:
   - Mixin argument expansion (2 tests)
   - Bootstrap4 investigation (1 test)
   - Output differences (102 tests in batches)

## 📚 Reference Material

- **Original issue analysis**: `.claude/reference-issues/`
- **Overall strategy**: `AGENT_ORCHESTRATION_STRATEGY.md`
- **Test results tracking**: `RUNTIME_ISSUES.md`

## 🐛 Troubleshooting

**If agents conflict**:
- agent-paths and agent-imports both touch import_manager.go
- Run agent-paths first, then agent-imports
- Or have them coordinate via git branches

**If tests fail**:
- Each TASK.md has debug strategies
- Use `LESS_GO_TRACE=1` for detailed execution tracing
- Use `LESS_GO_DEBUG=1` for error details
- Use `LESS_GO_DIFF=1` for CSS output differences

**If agent gets stuck**:
- Read the full TASK.md for context
- Check JavaScript reference implementation
- Ask agent to add debug output and trace the flow
- Move on if spending >30 minutes on one issue

## 💡 Tips for Success

1. **Start with agent-urls** - Simplest, good test of the approach
2. **Run agents truly in parallel** - Don't wait for one to finish
3. **Each agent is autonomous** - They have all the info they need
4. **Review branches** - Don't auto-merge without review
5. **Iterate** - If an agent fails, you can spawn a new one with refined instructions

## 🎓 What Happens After

Once runtime failures are fixed, we can tackle the 102 output difference tests in batches:
- Core features (26 tests)
- Extend functionality (6 tests)
- CSS standards (6 tests)
- Etc.

See `.claude/reference-issues/ISSUE_OUTPUT_DIFFS.md` for the batching strategy.

---

**Ready to start?** Pick an agent, copy their KICKOFF prompt, and let them work! 🚀
