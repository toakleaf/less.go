# Prompt: Create Task Specifications for Agent Distribution

## Context

The less.go project is a Go port of less.js. We have a comprehensive integration test suite that reveals which features are broken or incomplete. We need to create detailed task specifications so that multiple AI agents can work on fixes in parallel.

**IMPORTANT**: Before starting this task, check if there are any critical regressions that need to be fixed first. See `.claude/tracking/TEST_AUDIT_2025-11-05.md` for the latest test status.

## Your Task

**Create detailed task specification documents for the remaining test failures so independent agents can work on them in parallel.**

## Current Test Status

Run the tests first to get the current baseline:

```bash
# Get current status
pnpm -w test:go

# Get summary
pnpm -w test:go:summary

# Count perfect matches
pnpm -w test:go 2>&1 | grep "✅.*Perfect match" | wc -l

# Count compilation failures
pnpm -w test:go 2>&1 | grep "❌.*Compilation failed" | wc -l

# List all failures
pnpm -w test:go 2>&1 | grep "❌" | sed 's/.*❌ //' | sed 's/: Compilation failed.*//' | sort | uniq
```

## Existing Task Files

Check what already exists:
```bash
ls -la .claude/tasks/runtime-failures/
ls -la .claude/tasks/output-differences/
```

**Note**: Some existing task files may be outdated due to recent regressions. Verify test status before trusting existing task specs.

## Task Specification Template

For each failing test or group of related tests, create a task file following this structure:

```markdown
# Task: Fix [Feature Name]

**Status**: Available
**Priority**: [CRITICAL/HIGH/MEDIUM/LOW]
**Estimated Time**: [time estimate]
**Complexity**: [Low/Medium/High/Very High]

## Overview

[1-2 sentence description of what needs to be fixed]

## Failing Tests

- `test-name-1` ([suite name])
- `test-name-2` ([suite name])
...

## Current Behavior

**Error Message**:
```
[actual error message from test output]
```

**Test Command**:
```bash
pnpm -w test:go:filter -- "suite/test-name"
```

## Expected Behavior

[Describe what the test should do based on less.js documentation and test expectations]

## Investigation Starting Points

### Test Data

```bash
# Look at test input
cat packages/test-data/less/[suite]/[test-name].less

# Look at expected output
cat packages/test-data/css/[suite]/[test-name].css
```

### JavaScript Implementation

**Key files to examine**:
- `packages/less/src/less/tree/[relevant-file].js`

**Key logic**: [Brief description of what to look for]

### Go Implementation

**Files to check**:
- `packages/less/src/less/less_go/[relevant-file].go`

### Debugging Commands

```bash
# See differences
LESS_GO_DIFF=1 pnpm -w test:go:filter -- "test-name"

# Trace execution
LESS_GO_TRACE=1 pnpm -w test:go:filter -- "test-name"

# Compare with JavaScript
cd packages/less && npx lessc test-data/less/[suite]/[test-name].less -
```

## Likely Root Causes

**Hypothesis 1**: [Most likely cause]
- [Evidence]
- [What to check]

**Hypothesis 2**: [Alternative cause]
- [Evidence]
- [What to check]

## Implementation Hints

### If Hypothesis 1 is correct:

[Specific code guidance]

### If Hypothesis 2 is correct:

[Alternative approach]

## Success Criteria

- ✅ Test shows "Perfect match!" (or specific improvement expected)
- ✅ All unit tests still pass
- ✅ No regressions in other tests
- ✅ [Any specific criteria]

## Validation Checklist

```bash
# 1. Specific test passes
pnpm -w test:go:filter -- "test-name"
# Expected: [specific result]

# 2. Unit tests pass
pnpm -w test:go:unit
# Expected: All pass

# 3. Full integration suite
pnpm -w test:go
# Expected: No new failures
```

## Files Likely Modified

- `[file1.go]` - [what changes]
- `[file2.go]` - [what changes]

## Related Issues

[Any dependencies or related tasks]

## Notes

[Any additional context, warnings, or tips]
```

## Priority Guidelines

**CRITICAL**:
- Blocks multiple other tests
- Recent regression that broke working functionality
- Core feature that affects many use cases

**HIGH**:
- Runtime/compilation failures (tests don't even run)
- High-impact features (affects 5+ tests)
- Common LESS features

**MEDIUM**:
- Output differences (tests run but CSS doesn't match)
- Moderate impact (affects 1-4 tests)
- Standard LESS features

**LOW**:
- Polish/formatting issues
- Edge cases
- Advanced/rarely used features

## Categorization

### Runtime Failures (High Priority)
Place in: `.claude/tasks/runtime-failures/`

These are tests that don't even compile or crash during evaluation:
- import failures
- mixin matching errors
- namespace evaluation errors
- parser errors

### Output Differences (Medium Priority)
Place in: `.claude/tasks/output-differences/`

These tests compile but produce wrong CSS:
- Math operations
- Color functions
- Guards/conditionals
- Extend functionality
- Formatting issues

## Task Grouping

Group related tests together:

**Good grouping**:
- All `namespacing-*` tests that fail for the same reason
- All `extend-*` tests
- All `math-*` tests in a suite

**Bad grouping**:
- Random unrelated tests
- Tests that fail for completely different reasons

## Research Each Test

For each test or test group:

1. **Run the test** to see current error
2. **Read the test file** to understand what it's testing
3. **Compare with JavaScript** to understand expected behavior
4. **Identify root cause** (or hypotheses)
5. **Estimate complexity** and time
6. **Check dependencies** (does it require another fix first?)

## Update Tracking

After creating task files, update `.claude/tracking/assignments.json`:

```json
{
  "task-id": "fix-[name]",
  "status": "available",
  "priority": "high",
  "category": "runtime-failures",
  "tests_affected": ["test1", "test2"],
  "estimated_time": "2-3 hours",
  "complexity": "medium",
  "files_likely_modified": ["file1.go", "file2.go"],
  "task_file": ".claude/tasks/[category]/[name].md"
}
```

## Create Work Queue Summary

After creating all task files, create/update `.claude/AGENT_WORK_QUEUE.md`:

- List all high-priority tasks ready for assignment
- Group by priority
- Note which tasks are independent (can be done in parallel)
- Note which tasks have dependencies
- Provide brief description of each
- Include "How to Claim a Task" instructions

## Validation

Before finalizing your work:

```bash
# Verify all task files are valid markdown
find .claude/tasks -name "*.md" -type f

# Check that each task file has:
# - Clear test command
# - Current error message
# - Investigation steps
# - Validation checklist
# - Success criteria

# Ensure assignments.json is valid JSON
cat .claude/tracking/assignments.json | jq '.'
```

## Branch Setup

```bash
# Work on the base branch (or create from latest merged work)
git checkout claude/port-lessjs-golang-011CUoW25gKApW3ByzxRpUVk
git pull origin claude/port-lessjs-golang-011CUoW25gKApW3ByzxRpUVk

# Create your branch
git checkout -b claude/create-agent-tasks-YOUR_SESSION_ID
```

## Deliverables

Your PR should include:

1. **Task specification files** (`.claude/tasks/*/[name].md`)
   - At least 5-10 new task specs
   - Cover high-priority failures
   - Well-researched and detailed

2. **Updated assignments.json** (`.claude/tracking/assignments.json`)
   - All new tasks listed
   - Accurate metadata
   - Valid JSON

3. **Updated work queue** (`.claude/AGENT_WORK_QUEUE.md`)
   - Summary of available work
   - Prioritized list
   - Parallelization recommendations

4. **Test status report** (if things have changed)
   - Current baseline numbers
   - What tasks were created
   - Priority recommendations

## Tips

- **Don't assume** existing task files are correct - verify test status first
- **Be specific** in error messages and investigation steps
- **Include actual commands** that agents can copy-paste
- **Test your instructions** by following them yourself
- **Group wisely** - related tests together, but not too many
- **Estimate realistically** - check similar past fixes for time estimates
- **Note blockers** - if a task depends on another fix, document it

## Success Criteria

Your work is successful when:

- ✅ At least 5-10 detailed task specifications created
- ✅ Each task is well-researched with clear investigation steps
- ✅ High-priority failures are covered
- ✅ Tasks are properly categorized and prioritized
- ✅ assignments.json is updated and valid
- ✅ AGENT_WORK_QUEUE.md provides clear summary
- ✅ Independent tasks are identified for parallel work
- ✅ All documentation is clear and actionable

## Reference Documents

- Existing task template: `.claude/tasks/runtime-failures/mixin-args.md`
- Agent workflow: `.claude/strategy/agent-workflow.md`
- Validation requirements: `.claude/VALIDATION_REQUIREMENTS.md`
- Latest test audit: `.claude/tracking/TEST_AUDIT_2025-11-05.md`
