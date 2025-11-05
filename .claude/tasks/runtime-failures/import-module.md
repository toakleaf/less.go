# Task: Fix Import Module Syntax

**Status**: Available
**Priority**: MEDIUM
**Estimated Time**: 3-4 hours
**Complexity**: Medium-High
**Tests Affected**: 1 (import-module)

## Overview

The `import-module` test fails because the import system doesn't support module-style import syntax using the `@` prefix (like `@less/module-name`). This is a special syntax introduced in LESS for importing from "module" directories.

## Failing Test

### import-module
- **Status**: ❌ Compilation failed
- **Error**: `Syntax: : open @less/test-import-module/one/1.less: no such file or directory`
- **Issue**: Module path syntax (`@less/...`) not resolved to actual file path
- **Test Command**:
  ```bash
  pnpm -w test:go:filter -- "main/import-module"
  ```

## Current Behavior

```less
// Example from test
@import "@less/test-import-module/one/1.less";

// Current behavior:
// Tries to open: "@less/test-import-module/one/1.less" ❌
// (Treating @ as a literal directory name)
```

**Error**: `open @less/test-import-module/one/1.less: no such file or directory`

The system is treating `@less/test-import-module/` as a literal path including the `@` character, rather than recognizing it as a module path that needs special resolution.

## Expected Behavior

Module paths should be resolved to actual file system paths:

**Pattern**: `@module-name/path/to/file`

**Resolution Strategy**:
1. Recognize `@` prefix as module syntax
2. Look up module in configured module directories
3. Common locations:
   - `node_modules/@module-name/`
   - Configured module paths
4. Resolve to actual file path

**Example**:
```
@import "@less/test-import-module/one/1.less"
        ↓
Resolve to: node_modules/@less/test-import-module/one/1.less
        ↓
Load file from file system
```

## Investigation Starting Points

### Files to Examine

1. **`import_manager.go`** (PRIMARY) - Lines 100-350
   - File resolution logic
   - Check for module path detection (paths starting with `@`)
   - Module directory search paths

2. **`import.go`** - Lines 50-150
   - Import path parsing
   - May need to identify module syntax early

3. **`file_manager.go`** - Lines 50-200
   - File resolution at file system level
   - Path normalization

4. **Integration test configuration** - `integration_suite_test.go`
   - Check if test specifies module paths or `node_modules` locations
   - Might need to configure module directories for this test

### Debug Commands

```bash
# See the exact error
pnpm -w test:go:filter -- "main/import-module" 2>&1 | grep -A5 "import-module"

# Look at the test file
cat packages/test-data/less/_main/import-module.less

# Find where the target files actually are
find packages/test-data -name "*test-import-module*" -type d
find packages/test-data -name "1.less" | grep import-module

# Check if there's a node_modules or similar structure
ls -la packages/test-data/less/_main/ | grep -E "(@|node_modules)"

# Run with debug
LESS_GO_DEBUG=1 pnpm -w test:go:filter -- "main/import-module" 2>&1 | grep -E "(import|module|@less)"
```

### Compare with JavaScript

```bash
# See how less.js handles this
cd packages/less
npx lessc test-data/less/_main/import-module.less /tmp/js-output.css
echo "Exit code: $?"
```

## Root Cause Hypothesis

**Most Likely**: The import manager doesn't have logic to detect and handle module-style paths (`@prefix/path`).

**Needed Features**:
1. **Module Path Detection**: Recognize paths starting with `@` as module references
2. **Module Resolution**: Search for modules in:
   - `node_modules/@module-name/`
   - Configured module paths
   - Test data directory structure
3. **Fallback Handling**: If module not found, try literal path

### Likely Fix

In `import_manager.go`, add module path detection:

```go
func (im *ImportManager) resolveImportPath(importPath string, fromFile string) (string, error) {
    // Detect module syntax: @module-name/path
    if strings.HasPrefix(importPath, "@") {
        return im.resolveModulePath(importPath)
    }

    // Regular path resolution
    return im.resolveRegularPath(importPath, fromFile)
}

func (im *ImportManager) resolveModulePath(modulePath string) (string, error) {
    // Module path: @module-name/rest/of/path

    // Try node_modules
    nodeModulesPath := filepath.Join("node_modules", modulePath)
    if im.fileExists(nodeModulesPath) {
        return nodeModulesPath, nil
    }

    // Try configured module paths
    for _, modPath := range im.options.ModulePaths {
        fullPath := filepath.Join(modPath, modulePath)
        if im.fileExists(fullPath) {
            return fullPath, nil
        }
    }

    return "", fmt.Errorf("Module not found: %s", modulePath)
}
```

## Special Considerations

### Test Data Location

The test data might be organized to simulate a module structure:
```
packages/test-data/less/_main/
├── node_modules/
│   └── @less/
│       └── test-import-module/
│           └── one/
│               └── 1.less
```

Or it might be elsewhere:
```
packages/test-data/less/
└── @less/
    └── test-import-module/
        └── one/
            └── 1.less
```

You'll need to:
1. Find where the actual files are located
2. Configure the test to know where to look for modules
3. Or implement path resolution that searches multiple locations

### Test Configuration

In `integration_suite_test.go`, the test might need module path configuration:

```go
{
    Name: "import-module",
    Options: map[string]any{
        "modulePaths": []string{
            "./test-data/less",  // Or wherever @less/ directory is
        },
    },
}
```

## Success Criteria

- [ ] `import-module` test compiles successfully
- [ ] Module-style imports (`@module/path`) are resolved correctly
- [ ] Regular imports still work (no regressions)
- [ ] All unit tests pass: `pnpm -w test:go:unit`
- [ ] FULL integration test suite shows no regressions: `pnpm -w test:go`

## Validation Checklist

Before creating PR:

```bash
# 1. Verify specific test passes
pnpm -w test:go:filter -- "main/import-module"
# Expected: ✅ Compilation succeeds, output matches expected

# 2. Verify other import tests still work (NO REGRESSIONS)
pnpm -w test:go 2>&1 | grep "import" | grep -E "(✅|❌)"
# Expected: All currently passing import tests still pass
# Specifically: "impor" (yes, typo) should remain ✅ perfect match

# 3. Run ALL unit tests (catch any regressions) - REQUIRED
pnpm -w test:go:unit
# Expected: ✅ All unit tests pass (no failures)

# 4. Run FULL integration test suite - REQUIRED
pnpm -w test:go
# Expected:
# - ✅ Compilation failure count drops by 1
# - ✅ No new compilation failures
# - ✅ Perfect matches remain at 14+
```

**If any test fails that was passing before**: STOP and fix the regression before proceeding.

## Additional Context

### Module Syntax in LESS

The `@module/path` syntax is relatively recent in LESS and allows:
- Importing from npm packages: `@import "@less/plugin-advanced/index.less"`
- Organizing code in module-style: `@import "@company/styles/theme.less"`
- Avoiding path traversal: `@import "../../../node_modules/@less/..."`

### Similar to Node.js Module Resolution

This is similar to how Node.js resolves module imports:
1. Check for `@scope/package` in `node_modules/@scope/package/`
2. Walk up directory tree looking in each `node_modules/`
3. Use configured paths

We likely don't need full Node.js resolution - just basic `@module/path` → `<search-paths>/@module/path` lookup.

### Related Tests

- ✅ `impor` - Basic imports (perfect match)
- ❌ `import-interpolation` - Interpolated paths
- ❌ `import-reference` - Reference option with CSS
- ❌ `import-reference-issues` - Reference option with mixins
- ⚠️ Other import tests with output differences

## Notes

This is a feature that needs to be implemented, not a regression. The implementation should:
1. Be minimal - just enough to pass the test
2. Not break existing import functionality
3. Be extensible for future module features

Focus on making the test pass rather than implementing a complete module system. A simple path prefix replacement might be sufficient:
- Detect `@` prefix
- Search in known module locations
- Fall back to regular path resolution
