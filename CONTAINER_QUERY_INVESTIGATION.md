# Container Query Output Investigation

## Issue
Container queries nested in rulesets are not bubbling out correctly. They should behave like media queries.

## Current Behavior
```css
.widget {
  container-type: inline-size;
  @container (max-width: 350px) {    /* ❌ Stays nested */
    .cite {
      .wdr-authors {
        display: none;
      }
    }
  }
}
```

## Expected Behavior (like @media)
```css
.widget {
  container-type: inline-size;
}
@container (max-width: 350px) {    /* ✅ Bubbled to root */
  .widget .cite .wdr-authors {     /* ✅ Selectors joined */
    display: none;
  }
}
```

## Investigation Findings

### Media Queries Work Correctly
- `@media` nested in rulesets correctly bubble out and selectors are joined
- Testing shows: `.widget { @media { .cite } }` → `@media { .widget .cite }`

### Container vs Media Comparison
Both Container and Media:
1. Implement identical bubbling interfaces (BubbleSelectors)
2. Use same evaluation flow (Eval → evalTop/evalNested)
3. Are registered in type system
4. Create inner rulesets with Root=false

### The Mystery
**Media.evalTop() and Container.evalTop() both:**
- Clear `mediaBlocks` and `mediaPath` from context
- Return `this` (the media/container) when there's 1 block
- Return wrapped Ruleset when there are multiple blocks

**Yet media queries work but containers don't!**

This suggests there's a subtle difference in how they're processed that I haven't identified yet.

### Attempted Fixes (Unsuccessful)
1. **Don't clear mediaBlocks in evalTop** - Broke media queries entirely
2. **Return empty Ruleset for single containers** - Not tested fully

### Ruleset Bubbling Logic (lines 765-798 in ruleset.go)
```go
// After evaluation, bubble selectors to new mediaBlocks
for i := mediaBlockCount; i < len(mediaBlocks); i++ {
    if mb, ok := mediaBlocks[i].(interface{ BubbleSelectors(any) }); ok {
        mb.BubbleSelectors(selectors)
    }
}

// If root ruleset and mediaPath empty, append mediaBlocks to rules
if ruleset.Root && len(mediaPath) == 0 && mediaBlocks != nil && len(mediaBlocks) > 0 {
    ruleset.Rules = append(ruleset.Rules, mediaBlocks...)
    ...
}
```

### Key Questions for Future Investigation
1. Why does Media work correctly when it also clears mediaBlocks in evalTop?
2. Is there timing difference in when Media vs Container eval happens?
3. Is there post-processing that removes Media from parent rules?
4. Do we need to debug the actual runtime with trace logging?

### Related Commits
- `fe3a454` - Previous attempt, noted "Root flag discrepancy" but issue persists
- `8f3ca40` - Media query fixes (study this for clues)
- `a2387ae` - "Refactor Container evaluation to match Media pattern"

### Next Steps
1. Add trace logging to Media.Eval and Container.Eval to compare runtime behavior
2. Check if visitor pattern handles them differently
3. Compare GenCSS output paths
4. May need to step through JavaScript execution for clarity

### Test Status
- ✅ Unit tests: 2,290+ passing, no regressions
- ✅ Integration baseline: 79 perfect matches maintained
- ❌ Container test: Still failing (expected)
