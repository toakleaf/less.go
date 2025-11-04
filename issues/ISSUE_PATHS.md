# Include Path Resolution Issues - Agent Task

## 🎯 Mission
Fix 1 failing test related to include path resolution for imports

## 📊 Status
- **Tests Failing**: 1
- **Priority**: Low (Wave 1 - Independent, simple issue)
- **Complexity**: Low
- **Independence**: HIGH - Can be fixed in parallel with other issues

## ❌ Failing Test

### include-path
**File**: `packages/test-data/less/include-path/include-path.less`
**Error**: `open import-test-e: no such file or directory`

**Test Content**:
```less
@import "import-test-e";

data-uri {
  property: data-uri('image.svg');
}
image-size {
  property: image-size('image.svg');
}
```

**Problem**: The import `@import "import-test-e";` is failing because the file is not found. The file is likely in a different directory that should be searched via the `includePaths` option.

**Expected Behavior**:
- The test suite should configure include paths
- The import manager should search those paths
- The file should be found and imported successfully

---

## 🔍 Root Cause Analysis

### Theory 1: Include Paths Not Configured
The test expects the include path to be configured (pointing to a directory containing `import-test-e.less`), but:
- The test suite might not be setting the option
- OR the option is set but not being used

### Theory 2: Import Manager Not Using Include Paths
The import manager might:
- Not have access to the include paths option
- Not be searching those paths
- Not be implementing the search correctly

### Theory 3: Path Resolution Order
Include paths should be searched:
1. Current directory (relative to importing file)
2. Each include path in order
3. Fail if not found in any location

The search might not be happening in the right order or at all.

---

## 🔍 Test Setup Investigation

Looking at `integration_suite_test.go`, find the include-path test configuration:

```bash
grep -A10 "include-path" packages/less/src/less/less_go/integration_suite_test.go
```

This will show:
- What options are set for this test
- What the include path is supposed to be
- Where `import-test-e.less` file actually is

---

## 🔍 Go Files to Investigate

### Primary Files
1. **`integration_suite_test.go`** - Test configuration
   - Find include-path test suite definition
   - Check what `includePaths` or `paths` option is set
   - Verify the option is being passed to the renderer

2. **`import_manager.go`** - Import resolution
   - Check if it reads `includePaths` option
   - Verify it searches those paths
   - Look for the file resolution logic

3. **`options.go`** or **`environment.go`** - Options handling
   - Check if `includePaths` option is supported
   - Verify it's being passed through correctly

4. **`file_manager.go`** - File loading
   - Check how file paths are resolved
   - Verify include path searching

### Secondary Files
5. **`render.go`** - Rendering pipeline
   - Check how options are passed to import manager

---

## 📚 JavaScript Reference Files

Study these to understand correct behavior:

1. **`packages/less/src/less/import-manager.js`**
   - Look for `paths` or `includePaths` handling
   - Check file resolution logic with include paths

2. **`packages/less/test/index.js`** (the JavaScript test runner)
   - See how include-path test is configured
   - Check what options are set

---

## ✅ Success Criteria

### Target Success (1/1 test)
- `include-path` - File found via include path and imported successfully

---

## 🚫 Constraints

1. **NEVER modify any .js files**
2. **Must pass unit tests**: `pnpm -w test:go:unit`
3. **Must pass target test**: `pnpm -w test:go:filter -- "include-path"`
4. **No regressions**: All currently passing tests must still pass

---

## 🧪 Testing Strategy

### Find the File
```bash
# Find where import-test-e.less actually is
find packages/test-data -name "import-test-e.less" -o -name "import-test-e"

# Check the test directory structure
ls -la packages/test-data/less/include-path/
```

### Check Test Configuration
```bash
# See how the test is configured
grep -B5 -A10 "include-path" packages/less/src/less/less_go/integration_suite_test.go
```

### Run Specific Test
```bash
# Test the case
go test -run "TestIntegrationSuite.*include-path" -v

# With debug output
LESS_GO_DEBUG=1 go test -run "TestIntegrationSuite.*include-path" -v
```

### Verify No Regressions
```bash
# Run all unit tests
pnpm -w test:go:unit

# Run full integration suite
pnpm -w test:go:summary
```

---

## 📝 Expected Changes

### Likely Changes Needed

1. **integration_suite_test.go** - Add include path configuration
   ```go
   {
       Name:   "include-path",
       Options: map[string]any{
           "paths": []string{"./include-path-data/"}, // Or wherever the file is
       },
       Folder: "include-path/",
   }
   ```

2. **import_manager.go** - Use include paths
   ```go
   // Pseudo-code
   func (im *ImportManager) resolveImport(path string) (string, error) {
       // Try relative to current file
       if exists(relativePath) {
           return relativePath
       }

       // Try each include path
       for _, includePath := range im.options.Paths {
           fullPath := filepath.Join(includePath, path)
           if exists(fullPath) {
               return fullPath
           }
       }

       return "", fmt.Errorf("file not found: %s", path)
   }
   ```

3. **options.go** - Ensure paths option is supported
   - Verify `Paths` or `IncludePaths` field exists
   - Ensure it's being passed through to import manager

---

## 🎯 Debugging Hints

### Add Debug Output
```go
// In import_manager.go
fmt.Printf("[IMPORT-DEBUG] Looking for: %s\n", importPath)
fmt.Printf("[IMPORT-DEBUG] Include paths: %v\n", im.options.Paths)
fmt.Printf("[IMPORT-DEBUG] Current dir: %s\n", currentDir)
```

### Check File Locations
```bash
# From the test directory
cd packages/test-data/less/include-path
ls -la
# See what files are there

# Check parent directories
ls -la ..
ls -la ../imports
```

---

## 📊 Estimated Impact

- **Tests Fixed**: 1
- **Other Tests Potentially Improved**: Any test using include paths
- **Risk Level**: Low - Include path is an isolated feature

---

## 🔄 Iteration Strategy

### Round 1: Understand Structure
1. Find where `import-test-e.less` actually is
2. Check what the test expects
3. Verify the test configuration

### Round 2: Implement Fix
1. Either add include path to test config
2. OR fix import manager to search paths correctly
3. OR both

### Round 3: Verify and Commit
1. Run specific test
2. Run unit tests
3. Run integration suite
4. Commit and push

---

## 📋 Commit Message Template

```
Fix include path resolution for imports

The include-path test was failing because imports were not being resolved
through configured include paths.

Root cause: [Describe issue - test config or import manager]

Fix:
- [Added include path configuration to test] OR
- [Fixed import manager to search include paths] OR
- [Both]

Test fixed:
- include-path: ✅
```

---

## 🚀 When Done

1. **Commit** to branch: `claude/fix-paths-<your-session-id>`
2. **Push** to remote: `git push -u origin claude/fix-paths-<your-session-id>`
3. **Report**: "Fixed 1/1 include-path test"

---

## 💡 Key Insights

1. **Simple issue** - Likely just missing configuration or path search
2. **Low risk** - Include paths are isolated feature
3. **Quick fix** - Probably 10-30 lines of code
4. **Test first** - Understand the file structure before coding

---

## 🔗 Related Issues

- **ISSUE_IMPORTS.md** - Related to import handling, but different issue
- This is about PATH resolution, not import options

---

## ⚠️ Special Notes

1. **Check both test config AND implementation** - Could be either or both
2. **File might have different name** - Could be `import-test-e.less` or just `import-test-e`
3. **Relative vs absolute paths** - Include paths might be relative to different locations
4. **Multiple include paths** - Should support array of paths, search in order
5. **Very low complexity** - Good warm-up issue for an agent

---

## 📖 Quick Start

```bash
# Step 1: Find the file
find packages/test-data -name "*import-test-e*"

# Step 2: See test config
grep -A5 "include-path" packages/less/src/less/less_go/integration_suite_test.go

# Step 3: Run test to see error
go test -run "TestIntegrationSuite.*include-path" -v

# Step 4: Fix based on what you find
# Either update test config or import manager

# Step 5: Verify
go test -run "TestIntegrationSuite.*include-path" -v
pnpm -w test:go:summary
```
