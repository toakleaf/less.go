# Fix Detached Rulesets: The Correct Approach

## ⚠️ IMPORTANT: Previous Approach Was Wrong

**Branch:** `claude/fix-detached-rulesets-mediapath-011CUzHVDoVsRgMmWen7BWeb`
**Current Status:** 79/184 tests pass (0 regressions), but `detached-rulesets` still fails
**What's Wrong:** Previous session added ~300 lines of special-case logic that doesn't match JavaScript

## The Problem

When detached rulesets containing `@media` queries are called from within parent `@media` blocks, the merged media queries should appear in CSS output. Currently they're missing.

**Example:**
```less
@my-ruleset: {
  .my-selector {
    @media (tv) {
      background-color: black;
    }
  }
};
@media (orientation:portrait) {
  @my-ruleset();
}
```

**Expected:**
```css
@media (orientation: portrait) and (tv) {
  .my-selector {
    background-color: black;
  }
}
```

**Actual:** Media queries are missing entirely

## What Was Done (Wrong Approach)

The previous session added special-case handling for MultiMedia Rulesets:

1. **to_css_visitor.go (lines 637-724):** Added checks to skip extraction for `MultiMedia=true` rulesets
2. **ruleset.go (lines 1539-1554):** Added special GenCSS rendering for MultiMedia rulesets
3. **ruleset.go (lines 221-224):** Added `GetMultiMedia()` method

**Why this is wrong:**
- JavaScript doesn't special-case multiMedia in ToCSSVisitor.visitRuleset
- Added complexity without fixing the root cause
- Media nodes in MultiMedia Rulesets still have empty Rules arrays

## How JavaScript Actually Works

### Key Insight: JavaScript Uses `root` Flag, Not Special Cases

**JavaScript's approach** (join-selector-visitor.js:236):
```javascript
// In JoinSelectorVisitor.visitMedia:
mediaNode.rules[0].root = (context.length === 0 || context[0].multiMedia);
```

When a Media node is inside a MultiMedia Ruleset, JavaScript:
1. Sets the Media's inner ruleset's `root` property to `true`
2. ToCSSVisitor.visitRuleset checks `if (!rulesetNode.root)` (line 232)
3. Root rulesets skip the extraction logic (lines 232-259)
4. Non-root rulesets get their nested rulesets extracted

**The multiMedia flag is ONLY used in JoinSelectorVisitor to set root=true on inner rulesets**

### JavaScript Code References

**nested-at-rule.js (line 30):**
```javascript
evalTop(context) {
    if (context.mediaBlocks.length > 1) {
        result = new Ruleset(selectors, context.mediaBlocks);
        result.multiMedia = true;  // ← Only used by JoinSelectorVisitor
    }
    // ...
}
```

**join-selector-visitor.js (line 236):**
```javascript
visitMedia(mediaNode, visitArgs) {
    // Set inner ruleset to root if at top level OR inside multiMedia
    mediaNode.rules[0].root = (context.length === 0 || context[0].multiMedia);
    // ...
}
```

**to-css-visitor.js (line 232):**
```javascript
visitRuleset(rulesetNode, visitArgs) {
    if (!rulesetNode.root) {
        // Extract nested rulesets - this is where Media nodes get removed
        for (let i = 0; i < nodeRuleCnt; ) {
            rule = nodeRules[i];
            if (rule && rule.rules) {  // Media nodes have rules property
                rulesets.push(this._visitor.visit(rule));
                nodeRules.splice(i, 1);  // Remove from parent
                // ...
            }
        }
    }
    // Root rulesets skip extraction entirely
}
```

## The Correct Fix

### Step 1: Check Current JoinSelectorVisitor Implementation

**File:** `packages/less/src/less/less_go/join_selector_visitor.go`

Look for the `VisitMedia` method (around line 200-330). Check if it properly:
1. Detects when context has `multiMedia=true` (lines 236-321 seem to check this)
2. Sets `root=true` on the Media node's inner ruleset (media.Rules[0])

**Expected code pattern:**
```go
// In VisitMedia, when context[0] has multiMedia=true:
if len(media.Rules) > 0 {
    if ruleset, ok := media.Rules[0].(*Ruleset); ok {
        ruleset.Root = true  // ← This should happen
    }
}
```

### Step 2: Verify Context Passing

The JoinSelectorVisitor needs to know it's inside a MultiMedia Ruleset. Check:

**File:** `packages/less/src/less/less_go/join_selector_visitor.go` (VisitRuleset method)

When visiting a MultiMedia Ruleset, the visitor should:
```go
if ruleset.MultiMedia {
    // Add ruleset to context so child Media nodes know they're in multiMedia
    newContext := append([]any{ruleset}, context...)
    // Visit children with this context
}
```

### Step 3: Remove Wrong Special Cases (Optional)

The special-case logic in ToCSSVisitor isn't needed if root flag works correctly:
- `to_css_visitor.go` lines 637-696 (multiMedia checks)
- `ruleset.go` lines 1539-1554 (special GenCSS)

However, **test carefully** - the regression fix (visiting children) might still be needed.

### Step 4: Fix the Real Issue - Empty Media Rules

Even after the above, Media nodes may have empty Rules. The root cause:

**File:** `packages/less/src/less/less_go/media.go`

When `evalNested()` is called (line 206-332):
1. It modifies the Media node's features
2. Returns an empty placeholder Ruleset (line 332)
3. The Media node itself is added to mediaBlocks (line 556 in Eval)

**Problem:** The Media node added to mediaBlocks has been modified and may have lost its Rules.

**Possible fixes:**

**Option A: Clone before modifying**
```go
// In Media.Eval(), line ~555:
// Clone the media node before adding to mediaBlocks
mediaForBlocks := &Media{
    Node:     m.Node,
    Features: media.Features,
    Rules:    media.Rules,  // Evaluated rules
    // ... copy other fields
}
evalCtx.MediaBlocks = append(evalCtx.MediaBlocks, mediaForBlocks)
```

**Option B: Don't clear Rules in nested media**
```go
// In evalNested(), the Media node modifies itself but should keep Rules
// Check if Rules need to be preserved when returning empty placeholder
```

## Implementation Steps

### 1. Debug Current State

```bash
cd packages/less/src/less/less_go

# Check if root flag is being set
LESS_GO_DEBUG=1 go test -v -run "TestIntegrationSuite/main/detached-rulesets" 2>&1 | \
  grep -E "root=true|Root=true|multiMedia" | head -20

# Check Media Rules
LESS_GO_DEBUG=1 go test -v -run "TestIntegrationSuite/main/detached-rulesets" 2>&1 | \
  grep -E "Media.*Rules|Rules.*Media" | head -20
```

### 2. Add Debug Logging

**In join_selector_visitor.go VisitMedia:**
```go
if os.Getenv("LESS_GO_DEBUG") == "1" {
    if len(media.Rules) > 0 {
        if rs, ok := media.Rules[0].(*Ruleset); ok {
            fmt.Fprintf(os.Stderr, "[JoinSelectorVisitor.VisitMedia] Setting inner ruleset Root=%v (context multiMedia check)\n",
                len(context) == 0 || hasMultiMediaInContext(context))
        }
    }
}
```

**In to_css_visitor.go VisitRuleset:**
```go
if os.Getenv("LESS_GO_DEBUG") == "1" {
    if _, isMedia := rule.(*Media); isMedia {
        fmt.Fprintf(os.Stderr, "[ToCSSVisitor] Found Media in ruleset (root=%v), extracting=%v\n",
            rulesetNode.GetRoot(), !rulesetNode.GetRoot())
    }
}
```

### 3. Test Incrementally

**After each change:**
```bash
# Test target
pnpm -w test:go:filter -- "detached-rulesets"

# Check regressions (should stay at 79)
go test -v -run TestIntegrationSuite 2>&1 | grep "✅.*Perfect match" | wc -l

# Specific regression checks
go test -v -run "TestIntegrationSuite/main/extend-(chaining|media)" 2>&1 | grep "✅"
```

### 4. Expected Results

After correct fix:
```bash
# Should see:
# ✅ detached-rulesets: Perfect match!

# Test count should increase:
# 80/184 perfect matches (up from 79)
```

## Key Files to Modify

1. **join_selector_visitor.go** (lines 200-330)
   - Ensure multiMedia context detection works
   - Set Root=true on Media inner rulesets in multiMedia context

2. **media.go** (lines 206-332, 499-606)
   - Fix evalNested to preserve Rules when adding to mediaBlocks
   - OR clone Media before modifying in evalNested

3. **to_css_visitor.go** (lines 637-696) - OPTIONAL
   - Consider removing special multiMedia logic if root flag works
   - Keep the regression fix (visiting children)

## Testing Strategy

### Minimal Test Case

Create a simple test file:
```less
@my-ruleset: {
  .test {
    @media (tv) {
      color: red;
    }
  }
};
@media (portrait) {
  @my-ruleset();
}
```

Expected output:
```css
@media (portrait) and (tv) {
  .test {
    color: red;
  }
}
```

Test with:
```bash
echo '@my-ruleset: { .test { @media (tv) { color: red; } } }; @media (portrait) { @my-ruleset(); }' | \
  go run ./cmd/lessc/lessc.go -
```

## Success Criteria

1. ✅ `detached-rulesets` test passes with perfect match
2. ✅ Test count increases to 80/184 (or stays at 79 if test was already passing)
3. ✅ Zero regressions (especially extend-chaining, extend-media)
4. ✅ All unit tests pass: `pnpm -w test:go:unit`
5. ✅ Code changes align with JavaScript implementation
6. ✅ No special-case logic needed in ToCSSVisitor

## Common Pitfalls

1. **Don't add more special cases** - Use the root flag mechanism
2. **Test regressions frequently** - extend tests are sensitive to visitor changes
3. **Check both evaluation AND visitor phases** - The fix likely needs both
4. **Verify Rules preservation** - Media nodes must have content when rendered

## References

- JavaScript evalTop: `packages/less/src/less/tree/nested-at-rule.js:23-38`
- JavaScript JoinSelectorVisitor: `packages/less/src/less/visitors/join-selector-visitor.js:229-254`
- JavaScript ToCSSVisitor: `packages/less/src/less/visitors/to-css-visitor.js:224-280`
- Go JoinSelectorVisitor: `packages/less/src/less/less_go/join_selector_visitor.go:200-330`
- Go ToCSSVisitor: `packages/less/src/less/less_go/to_css_visitor.go:584-742`

## Quick Start Command

```bash
cd /home/user/less.go
git checkout claude/fix-detached-rulesets-mediapath-011CUzHVDoVsRgMmWen7BWeb

# Read this file
cat .claude/tasks/runtime-failures/DETACHED-RULESETS-CORRECT-FIX.md

# Start debugging
cd packages/less/src/less/less_go
LESS_GO_DEBUG=1 go test -v -run "TestIntegrationSuite/main/detached-rulesets" 2>&1 | less
```

Good luck! The correct fix should be much simpler than what was attempted before.
