# Agent: Include Path Resolution

## 🎯 Your Mission
Fix include path resolution so imports can find files in configured include directories.

## 📊 Status
**Tests to Fix**: 1
**Branch**: `claude/fix-paths-<your-session-id>`
**Independence**: MEDIUM - May touch import_manager.go (agent-imports also uses this, but different sections)

## ❌ The Problem
```less
@import "import-test-e";  // File not in current dir, should be found via include path
```
**Error**: `open import-test-e: no such file or directory`

The test expects include paths to be configured and searched, but they're not being used.

## 🔍 Files You'll Modify
- `packages/less/src/less/less_go/integration_suite_test.go` - Add include path config for test
- `packages/less/src/less/less_go/import_manager.go` - Ensure include paths are searched
- `packages/less/src/less/less_go/file_manager.go` - Maybe: file resolution

**Potential Conflict**: agent-imports also modifies import_manager.go, but different sections. If you focus on the "path search" logic and they focus on "reference flag" logic, you won't conflict.

## 🔬 Investigation Steps

### Step 1: Find the file
```bash
find packages/test-data -name "*import-test-e*"
```

### Step 2: Check test configuration
```bash
grep -B5 -A10 "include-path" packages/less/src/less/less_go/integration_suite_test.go
```

### Step 3: Understand what's needed
The test suite likely needs to configure an `includePaths` or `paths` option pointing to where `import-test-e.less` actually lives.

## ✅ Success Criteria
- [ ] `include-path` test passes
- [ ] All unit tests pass: `pnpm -w test:go:unit`
- [ ] No regressions: `pnpm -w test:go:summary`

## 🧪 Test Commands
```bash
cd packages/less/src/less/less_go
go test -run "TestIntegrationSuite.*include-path" -v

# After fix
pnpm -w test:go:unit
pnpm -w test:go:summary
```

## 🔑 Likely Fix

You'll probably need to:

1. **Add test configuration** in `integration_suite_test.go`:
```go
{
    Name: "include-path",
    Options: map[string]any{
        "paths": []string{"./path-to-include-dir/"}, // Where import-test-e.less is
    },
    Folder: "include-path/",
}
```

2. **Ensure import_manager.go searches include paths**:
```go
// Try relative path first
// Then try each path in options.Paths
for _, includePath := range im.options.Paths {
    fullPath := filepath.Join(includePath, importPath)
    if fileExists(fullPath) {
        return fullPath, nil
    }
}
```

## 📋 When Done
```bash
git add packages/less/src/less/less_go/*.go
git commit -m "Fix include path resolution for imports

Imports now search configured include paths when files aren't found
relative to the current file.

Test fixed:
- include-path: ✅"

git push -u origin claude/fix-paths-<your-session-id>
```

Report: "Fixed 1/1 include-path test. Ready for PR."
