# Investigation Prompt: Fix import-inline Media Query Bug

## Your Mission

Fix the `import-inline` test where `@import (inline) url("file.css") (min-width:600px);` doesn't output the media query block.

**Run this test to see the failure:**
```bash
cd /home/user/less.go
pnpm -w test:go -- -run "TestIntegrationSuite/main/import-inline$"
```

**Expected output (missing):**
```css
@media (min-width: 600px) {
  #css { color: yellow; }
}
```

## The Problem in a Nutshell

In `packages/less/src/less/less_go/import.go` around line 447, when creating a Media node for inline imports with features:

```go
if features != nil {
    var featuresValue any = features
    if featuresMap, ok := features.(map[string]any); ok {
        if val, exists := featuresMap["value"]; exists {
            featuresValue = val
        }
    } else if featuresWithValue, ok := features.(interface{ GetValue() any }); ok {
        featuresValue = featuresWithValue.GetValue()
    }
    return NewMedia([]any{contents}, featuresValue, i._index, i._fileInfo, nil), nil
}
```

**The Issue**: After evaluation, `features` is a `*Paren` object, but the current extraction logic doesn't handle `*Paren`. So it passes the `*Paren` directly to `NewMedia`, which wraps it in another `*Value`, creating incorrect nesting.

## What We Know

1. Media.GenCSS() IS being called - the node exists
2. The evaluated `features` is `*Paren` type (confirmed via debug output)
3. `*Paren.Value` contains a `*Value` object
4. `*Value.Value` contains the actual array of expressions
5. JavaScript does: `new Media([contents], this.features.value)` - extracts `.value` property

## Your Action Plan

### Step 1: Add Debug Output (5 minutes)
Add this to `import.go` line 447 area:
```go
if features != nil {
    fmt.Printf("=== INLINE IMPORT DEBUG ===\n")
    fmt.Printf("features type: %T\n", features)

    if paren, ok := features.(*Paren); ok {
        fmt.Printf("Paren.Value type: %T\n", paren.Value)
        if val, ok := paren.Value.(*Value); ok {
            fmt.Printf("Value.Value type: %T, content: %+v\n", val.Value, val.Value)
        }
    }
```

Run the test and examine the exact type chain.

### Step 2: Fix the Extraction Logic (15 minutes)
Based on what you see, implement proper extraction:
```go
if features != nil {
    var featuresValue any = features

    // Extract from *Paren if needed
    if paren, ok := features.(*Paren); ok {
        featuresValue = paren.Value
    }

    // Extract from *Value if needed
    if val, ok := featuresValue.(*Value); ok {
        featuresValue = val.Value
    }

    return NewMedia([]any{contents}, featuresValue, i._index, i._fileInfo, nil), nil
}
```

Apply the same fix around line 495 for regular imports.

### Step 3: Verify the Fix
```bash
pnpm -w test:go -- -run "TestIntegrationSuite/main/import-inline$"
```

Should show: `✅ import-inline: Perfect match!`

### Step 4: Check for Regressions
```bash
pnpm -w test:go:unit && pnpm -w test:go
```

All tests must pass.

### Step 5: Commit and Push
```bash
git add packages/less/src/less/less_go/import.go
git commit -m "Fix import-inline media query output

Extract features from *Paren objects before passing to NewMedia to avoid
double-wrapping in Value objects. Handle both *Paren and *Value types to
properly unwrap the media query features.

Fixes: import-inline test - media query block now appears in output"

git push -u origin claude/fix-inline-import-media-queries-011CUuVZ9TJewC5KgGhR5zaq
```

## If It Still Doesn't Work

Check these alternative root causes:

1. **The Media node isn't in the parent's rules** - trace how `Import.Eval()` returns the Media and how `Ruleset.EvalImports()` adds it to the tree

2. **Media.Eval() is transforming it wrong** - check `packages/less/src/less/less_go/media.go` lines 410-507

3. **Visibility is blocking it** - check Media.GenCSS() line 391-397 for visibility checks

4. **Compare with working @media** - create a regular `@media (min-width: 600px) { ... }` rule and compare its features structure

## Full Investigation Document

Read `.claude/tasks/IMPORT_INLINE_INVESTIGATION.md` for complete context, debug findings, and all attempted approaches.

## Quick Reference

**Test file:** `packages/test-data/less/_main/import-inline.less`
**Main code:** `packages/less/src/less/less_go/import.go` lines 440-460, 479-495
**Media impl:** `packages/less/src/less/less_go/media.go`
**JS reference:** `packages/less/src/less/tree/import.js` line 167
**Branch:** `claude/fix-inline-import-media-queries-011CUuVZ9TJewC5KgGhR5zaq`

## Success Criteria

When you run the test, you should see:
```
=== RUN   TestIntegrationSuite/main/import-inline
    integration_suite_test.go:460: ✅ import-inline: Perfect match!
```

**Go fix it!** 🚀
