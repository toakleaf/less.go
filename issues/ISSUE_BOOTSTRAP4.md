# Bootstrap 4 Integration Test - Agent Task

## 🎯 Mission
Fix 1 large integration test - Bootstrap 4 LESS compilation

## 📊 Status
- **Tests Failing**: 1 (large, real-world test)
- **Priority**: Medium (Wave 2 - Complex, might reveal multiple issues)
- **Complexity**: HIGH (large codebase, multiple dependencies)
- **Independence**: MEDIUM - Might reveal issues fixed by other agents

## ❌ Failing Test

### bootstrap4
**File**: `packages/test-data/less/3rd-party/bootstrap4.less`
**Error**: `open bootstrap-less-port/less/bootstrap: no such file or directory`

**Problem**: The test is trying to import the Bootstrap 4 LESS files, but:
1. The files might not be present in the test-data directory
2. The import path might be incorrect
3. The test might require special setup (npm install, submodule, etc.)

---

## 🔍 Root Cause Analysis

### Theory 1: Missing Dependencies
Bootstrap 4 files are not in the repository:
- Need to be downloaded/installed
- Might be a git submodule
- Might need npm install
- Might be ignored in .gitignore

### Theory 2: Path Resolution
The import path `bootstrap-less-port/less/bootstrap` might be:
- Relative to wrong directory
- Needing include path configuration
- Incorrect path structure

### Theory 3: Multiple Issues
Bootstrap 4 is a large, complex LESS codebase. It might:
- Trigger multiple bugs
- Use advanced features not yet working
- Combine issues that individually work but together fail

---

## 🔍 Investigation Steps

### Step 1: Find the Test File
```bash
# See what's in the test
cat packages/test-data/less/3rd-party/bootstrap4.less

# Check if bootstrap files exist
ls -la packages/test-data/less/3rd-party/bootstrap*
find packages/test-data -name "*bootstrap*" -type d
```

### Step 2: Check Git/Submodules
```bash
# Check for submodules
cat .gitmodules

# Check if bootstrap is a submodule
git submodule status

# Check git history for bootstrap
git log --all --oneline -- "*bootstrap*" | head -20
```

### Step 3: Check JavaScript Test Setup
```bash
# See how JavaScript tests handle bootstrap
grep -r "bootstrap4" packages/less/test/
grep -r "bootstrap" packages/less/test/index.js
```

---

## 🔍 Go Files to Investigate

This test is more about **setup** than code issues. But if it's a code issue:

### If File Resolution Issue
1. **`import_manager.go`** - Path resolution
2. **`file_manager.go`** - File finding
3. **`integration_suite_test.go`** - Test configuration

### If Multiple Bugs
Run the test and see what fails. It might reveal:
- Import issues (see ISSUE_IMPORTS.md)
- Mixin issues (see ISSUE_MIXINS_ARGS.md)
- URL issues (see ISSUE_URLS.md)
- New issues not yet discovered

---

## 📚 JavaScript Reference Files

Check how JavaScript tests set this up:

1. **`packages/less/test/index.js`**
   - Look for bootstrap4 test configuration
   - Check if special setup is needed

2. **`packages/test-data/less/3rd-party/`**
   - Check if there's a README or setup instructions

---

## ✅ Success Criteria

### Minimum Success
- Understand why test is failing
- Document if it's a setup issue vs code issue
- If setup: provide setup instructions
- If code: identify which existing issue(s) it's related to

### Target Success (1/1 test)
- `bootstrap4` - Compiles successfully
- Output matches expected CSS

---

## 🚫 Constraints

1. **NEVER modify any .js files**
2. **Must pass unit tests**: `pnpm -w test:go:unit`
3. **Must pass target test**: `pnpm -w test:go:filter -- "bootstrap4"`
4. **No regressions**: All currently passing tests must still pass

---

## 🧪 Testing Strategy

### Understand the Test
```bash
# Read the test file
cat packages/test-data/less/3rd-party/bootstrap4.less

# Check expected output
cat packages/test-data/css/3rd-party/bootstrap4.css 2>/dev/null || echo "No expected output file"

# Look for related files
find packages/test-data -path "*3rd-party*" -name "*bootstrap*"
```

### Setup Bootstrap (if needed)
```bash
# If it's a submodule
git submodule update --init --recursive

# If it needs to be downloaded
# Check if there's a script or instructions
cat packages/test-data/less/3rd-party/README* 2>/dev/null
```

### Run Test
```bash
# Try the test
go test -run "TestIntegrationSuite/3rd-party/bootstrap4" -v

# With debug
LESS_GO_DEBUG=1 go test -run "TestIntegrationSuite/3rd-party/bootstrap4" -v
```

---

## 📝 Expected Actions

### If It's a Setup Issue
1. **Document the setup**:
   - What files need to be present
   - How to obtain them
   - Where they should be placed

2. **Add setup to test infrastructure**:
   - Update test runner to check for required files
   - Add setup instructions to README
   - Consider adding bootstrap files to repo or as submodule

3. **Report**:
   - "Bootstrap4 test requires setup: [instructions]"
   - "Setup not yet automated - needs manual intervention"

### If It's a Code Issue
1. **Identify the bug**:
   - Run test with debug output
   - See what fails
   - Map to existing issue documents (imports, mixins, urls, etc.)

2. **Fix or defer**:
   - If it's covered by another issue: note in that issue
   - If it's a new issue: create ISSUE_BOOTSTRAP4_BUG.md
   - If too complex: defer for later

3. **Report**:
   - "Bootstrap4 fails due to: [issue name]"
   - "Will be fixed when [other issue] is resolved"

### If It's Multiple Issues
1. **Document all issues found**
2. **Fix what you can**
3. **Note dependencies** - "Requires fixes from: [list of issues]"
4. **Report**: "Bootstrap4 reveals X issues, Y already tracked, Z new"

---

## 🎯 Priority Guidance

### High Priority
- If it's a simple setup issue: Fix it

### Medium Priority
- If it's ONE bug covered by another issue: Note it and move on

### Low Priority
- If it's MULTIPLE bugs: Defer until other issues are fixed
- If it's a NEW complex bug: Document and defer

**Don't spend > 1 hour on this unless it's obviously simple**

---

## 📊 Estimated Impact

- **Tests Fixed**: 1 (if successful)
- **Complexity**: Very high - Bootstrap is large
- **Value**: High - Real-world validation
- **Risk**: Medium - Might reveal new issues

---

## 🔄 Iteration Strategy

### Round 1: Investigation (15 min)
1. Find what files are needed
2. Check if they exist
3. Understand the error

### Round 2: Setup or Fix (30 min)
- If setup: Get files in place
- If code bug: Identify which issue
- If complex: Document and defer

### Round 3: Verify or Defer (15 min)
- If fixed: Test thoroughly
- If not: Document for later
- Report findings

---

## 📋 Commit Message Templates

### If Setup Fixed
```
Fix bootstrap4 test setup

The bootstrap4 test was failing because required LESS files were not present.

Fix:
- [Downloaded/installed/configured bootstrap files]
- [Updated paths or test configuration]

Test fixed:
- bootstrap4: ✅
```

### If Code Issue Found
```
Document bootstrap4 test dependencies

The bootstrap4 test reveals the following issues:
- [Issue #1: description]
- [Issue #2: description]

This test will pass once those issues are resolved.

Documented in: ISSUE_BOOTSTRAP4_BUG.md
Test status: Deferred pending bug fixes
```

---

## 🚀 When Done

1. **If Fixed**:
   - Commit to branch: `claude/fix-bootstrap4-<your-session-id>`
   - Push to remote
   - Report: "Fixed bootstrap4 test - [what was wrong]"

2. **If Deferred**:
   - Commit documentation to branch: `claude/investigate-bootstrap4-<your-session-id>`
   - Report: "Bootstrap4 test requires: [list of issues] - deferred"

---

## 💡 Key Insights

1. **Large real-world test** - More valuable than synthetic tests
2. **Likely multiple issues** - Bootstrap uses many LESS features
3. **Don't over-invest** - If too complex, document and move on
4. **Setup vs code** - Distinguish between missing files and bugs
5. **Validation test** - Success here validates many fixes

---

## 🔗 Related Issues

Bootstrap might reveal issues from:
- **ISSUE_IMPORTS.md** - Complex import hierarchies
- **ISSUE_MIXINS_ARGS.md** - Advanced mixin patterns
- **ISSUE_URLS.md** - URL handling
- **ISSUE_NAMESPACING.md** - Namespace resolution
- New bugs not yet categorized

---

## ⚠️ Special Notes

1. **Time-box this** - Max 1 hour unless obviously simple
2. **Setup first** - Don't debug code if files are missing
3. **Validation value** - High value if it works, OK to defer if not
4. **Document well** - If deferred, make it easy for next person
5. **Check JavaScript** - See how JS tests handle it

---

## 🎓 Learning Opportunity

Bootstrap 4 is a **complex, real-world codebase**. It will:
- Stress-test your implementation
- Reveal edge cases
- Validate that fixes work in practice
- Show what features are still missing

Even if you don't fix it, **understanding what fails is valuable**.

---

## 📖 Quick Start

```bash
# Minute 1-5: Investigate
cat packages/test-data/less/3rd-party/bootstrap4.less
find packages/test-data -name "*bootstrap*"
git submodule status

# Minute 5-10: Identify Issue
go test -run "TestIntegrationSuite/3rd-party/bootstrap4" -v 2>&1 | head -50

# Minute 10-60: Fix or Document
# ... based on what you find ...

# If simple: Fix it
# If complex: Document it
# Either way: Report findings
```

---

## 🎯 Expected Outcome

**Most Likely**: "Bootstrap files not found - needs setup"
**Possible**: "Bootstrap reveals issues: [list]"
**Unlikely**: "Bootstrap works perfectly!"

Adjust your approach based on what you find.
