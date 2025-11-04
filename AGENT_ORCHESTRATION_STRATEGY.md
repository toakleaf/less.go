# Multi-Agent Orchestration Strategy for less.go

**Goal**: Get all 185 integration tests passing using parallel independent agents with minimal human oversight and budget-conscious execution.

**Budget**: $1,000 credit with checkpoints at each phase

---

## 🏗️ Architecture Overview

```
Master Orchestrator Agent (You monitoring via Claude Code)
    ├── Issue Catalog (this document + issue files)
    ├── Agent Pool (parallel workers)
    │   ├── Agent 1: Import Issues (4 tests)
    │   ├── Agent 2: Namespace Resolution (2 tests)
    │   ├── Agent 3: URL Processing (2 tests)
    │   ├── Agent 4: Mixin Args (2 tests)
    │   ├── Agent 5: Path Resolution (1 test)
    │   ├── Agent 6: Bootstrap4 (1 large test)
    │   └── Agent 7-N: Output Differences (102 tests, batched)
    └── Integration Manager
        ├── Merge successful fixes
        ├── Run regression tests
        └── Track metrics
```

---

## 📋 Phase 1: Issue Cataloging (Estimated: $10-20)

**Objective**: Create detailed, independent issue documents for each failing category

**Tasks**:
1. Run full test suite with detailed output capture
2. Group failures by root cause
3. Create issue documents with:
   - Test names affected
   - Error messages and stack traces
   - Expected vs actual output
   - Relevant Go files to investigate
   - Relevant JS reference files
   - Independence score (can be fixed in isolation?)

**Deliverables**:
- `ISSUE_IMPORTS.md` - Import reference/interpolation issues (4 tests)
- `ISSUE_NAMESPACING.md` - Namespace resolution (2 tests)
- `ISSUE_URLS.md` - URL processing (2 tests)
- `ISSUE_MIXINS_ARGS.md` - Mixin argument binding (2 tests)
- `ISSUE_PATHS.md` - Include path resolution (1 test)
- `ISSUE_BOOTSTRAP4.md` - Bootstrap4 integration (1 test)
- `ISSUE_OUTPUT_DIFFS.md` - Categorized output differences (102 tests)

**Output**: Clear roadmap showing which issues can be parallelized

---

## 📋 Phase 2: Parallel Agent Setup (Estimated: $5-10)

**Objective**: Prepare independent agent task descriptions

**Tasks**:
1. For each issue category, create:
   - Isolated task description
   - Success criteria (specific tests passing)
   - Constraint: "Never modify JS files"
   - Commit message template
   - Branch naming convention
2. Determine optimal parallelization:
   - Group truly independent issues
   - Identify potential conflicts (same files)
   - Create execution waves

**Wave 1 (Fully Independent)**:
- Import issues → `claude/fix-imports-<session-id>`
- Namespace resolution → `claude/fix-namespacing-<session-id>`
- URL processing → `claude/fix-urls-<session-id>`
- Path resolution → `claude/fix-paths-<session-id>`

**Wave 2 (May touch overlapping files)**:
- Mixin args → `claude/fix-mixin-args-<session-id>`
- Bootstrap4 → `claude/fix-bootstrap4-<session-id>`

**Wave 3 (Large, batched work)**:
- Output differences (grouped by category)

**Deliverables**:
- `AGENT_TASKS.md` - Task descriptions for each agent
- Execution wave plan
- Conflict matrix (which tasks can run in parallel)

---

## 📋 Phase 3: Parallel Execution (Estimated: $200-400)

**Objective**: Run independent agents in parallel to fix issues

**Process**:
1. Start Wave 1 agents simultaneously (4 agents in parallel)
2. Each agent:
   - Receives isolated task from `AGENT_TASKS.md`
   - Works on separate branch
   - Runs specific tests to verify fix
   - Commits with clear message
   - Reports success/failure
3. Master agent monitors:
   - Agent completion
   - Test results
   - Budget consumption
   - Conflicts/blockers

**Agent Task Template**:
```markdown
# Agent Task: Fix Import Reference Issues

## Your Mission
Fix 4 failing tests related to import reference functionality

## Tests to Fix
- import-reference
- import-reference-issues
- import-interpolation (architectural - may need to defer)
- import-module

## Constraints
- NEVER modify any .js files
- Only modify .go files
- All changes must pass unit tests: `pnpm -w test:go:unit`
- Target tests must pass: `pnpm -w test:go:filter -- "import-reference"`

## Investigation Starting Points
- import.go - Import option handling
- import_visitor.go - Import processing
- import_manager.go - Import resolution
- JavaScript reference: packages/less/src/less/import-manager.js

## Success Criteria
- At least 2 of 4 tests passing
- No regression in other tests
- Clear commit message explaining fix

## When Done
- Commit to branch: `claude/fix-imports-<your-session-id>`
- Push to remote
- Report: "Fixed X/4 import tests: [test names]"
```

**Budget Checkpoint**: After Wave 1, review:
- Cost consumed: ~$100-150
- Tests fixed: Expect 6-10 tests
- Value assessment: Continue to Wave 2?

---

## 📋 Phase 4: Integration & Verification (Estimated: $50-100)

**Objective**: Merge fixes and ensure no regressions

**Process**:
1. Master agent creates integration branch
2. Cherry-pick successful fixes from agent branches
3. Run full test suite after each merge
4. If conflicts/regressions:
   - Isolate the conflict
   - Spawn conflict-resolution agent
5. Update metrics and documentation

**Deliverables**:
- All fixes merged to main development branch
- Updated RUNTIME_ISSUES.md
- Updated CLAUDE.md with new test counts
- Regression report

---

## 📋 Phase 5: Output Differences (Estimated: $300-400)

**Objective**: Fix 102 tests that compile but produce wrong CSS

**Process**:
1. Categorize by issue type:
   - Selector generation (guards, extends, etc.)
   - Math/operator precedence
   - Import content insertion
   - Source maps
   - Compression/minification
   - Custom properties
   - URL rewriting
2. Create batches of similar issues
3. Run wave of agents (5-10 agents on different categories)
4. Iterate until all output matches

**Budget Checkpoint**: Review every 20 tests fixed

---

## 💰 Budget Management Strategy

**Cost Estimates** (conservative):
- Phase 1 (Cataloging): $10-20
- Phase 2 (Setup): $5-10
- Phase 3 (Wave 1): $100-150
- **CHECKPOINT 1**: Review progress vs cost
- Phase 3 (Wave 2): $50-100
- **CHECKPOINT 2**: Review progress vs cost
- Phase 4 (Integration): $50-100
- **CHECKPOINT 3**: Assess remaining budget
- Phase 5 (Output diffs): $300-400
- **Final**: Reserve $100 for unexpected issues

**Total Estimated**: $615-880 (within $1,000 budget)

**Cost Control Measures**:
1. Use focused agents (specific file reads, targeted changes)
2. Avoid exploratory agents when issue is clear
3. Batch similar issues
4. Human checkpoint before major waves
5. Stop and reassess if single wave exceeds $150

---

## 🎯 Success Metrics

**Phase Goals**:
- Phase 1-2: Issue clarity (success = clear task definitions)
- Phase 3 Wave 1: 6-10 tests fixed (50-83% of wave targets)
- Phase 3 Wave 2: 2-3 tests fixed (33-50% of wave targets)
- Phase 4: No regressions, clean merge
- Phase 5: 70+ output diffs fixed (68% of remaining)

**Final Success**:
- 95%+ tests passing (176/185)
- 90%+ perfect CSS matches (166/185)
- All runtime errors resolved
- Bootstrap4 test passing (key milestone)

---

## 🚀 Immediate Next Steps

1. **YOU SAY**: "Start Phase 1"
   → I create detailed issue documents by analyzing failing tests

2. **CHECKPOINT**: Review issue docs, adjust plan if needed

3. **YOU SAY**: "Start Phase 3 Wave 1"
   → I spawn 4 parallel agents with isolated tasks

4. **CHECKPOINT**: Review Wave 1 results and cost (~$150)

5. **Continue or Adjust** based on results

---

## 📊 Progress Tracking

**Master Dashboard** (to be updated after each phase):

```
Phase 1: [ ] Complete - Issue docs created
Phase 2: [ ] Complete - Agent tasks defined
Phase 3 Wave 1: [ ] Complete - _/_ tests fixed, $____ spent
Phase 3 Wave 2: [ ] Complete - _/_ tests fixed, $____ spent
Phase 4: [ ] Complete - All merged, 0 regressions
Phase 5: [ ] In Progress - __/102 output diffs fixed

Current Status:
- Tests Passing: 71/185 (38.4%)
- Perfect CSS: 15/185 (8.1%)
- Budget Used: $____
- Budget Remaining: $____
```

---

## ⚠️ Risk Mitigation

**Risk**: Agents create conflicts by modifying same files
**Mitigation**: Wave-based execution, conflict matrix, integration testing

**Risk**: Agent makes changes that break other tests
**Mitigation**: Each agent must run full unit test suite before committing

**Risk**: Budget runs out before completion
**Mitigation**: Checkpoints every $150, prioritize high-impact fixes first

**Risk**: Issue is more complex than expected
**Mitigation**: Agents report blockers, can defer to next iteration

**Risk**: Agents don't understand context
**Mitigation**: Detailed issue docs with JS reference files and Go investigation points

---

## 🔄 Iteration Strategy

After first pass:
1. Assess remaining failures
2. Group by new patterns discovered
3. Run second wave with refined task descriptions
4. Target: 95% pass rate after 2-3 iterations

---

## 📝 Documentation Updates

After each phase, update:
- `CLAUDE.md` - Current test counts
- `RUNTIME_ISSUES.md` - Remove fixed issues
- `AGENT_ORCHESTRATION_STRATEGY.md` - Progress tracking
- Commit messages should reference issue docs

---

## 🎓 Learning & Adaptation

**If an agent fails**:
1. Review its approach
2. Refine task description
3. Add more context about JS implementation
4. Try again with improved instructions

**If an agent succeeds**:
1. Document the successful pattern
2. Apply pattern to similar issues
3. Create template for similar fixes

---

## ✅ Ready to Begin?

**Say "Start Phase 1"** and I'll:
1. Run full test suite with detailed output
2. Analyze each failing test
3. Create comprehensive issue documents
4. Present you with the catalog for review

**Estimated time**: 10-15 minutes
**Estimated cost**: $10-20

After review, we'll proceed to parallel agent execution at your command.
