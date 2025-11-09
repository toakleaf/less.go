# Fix Comment Handling Between Property Names and Colons

## Problem Statement

Comments that appear between CSS property names and colons are not being properly filtered out during parsing or CSS generation in the Go port of less.js.

## Expected Behavior

**Input LESS:**
```less
.selector {
  color/* survive */ /* me too */: grey, /* blue */ orange;
  -webkit-border-radius: 2px /* webkit only */;
  -moz-border-radius: (2px * 4) /* moz only with operation */;
}
```

**Expected CSS Output:**
```css
.selector {
  color: grey, /* blue */ orange;
  -webkit-border-radius: 2px /* webkit only */;
  -moz-border-radius: 8px /* moz only with operation */;
}
```

**Note**: Comments BEFORE the colon (between property name and `:`) should be removed, but comments AFTER the colon should be preserved.

## Test Case

- **Input file**: `packages/test-data/less/_main/comments.less` (see line 65)
- **Expected output**: `packages/test-data/css/_main/comments.css` (see line 51)
- **Test command**: `go test -v -run "TestIntegrationSuite/main/comments$" ./packages/less/src/less/less_go/...`
- **Current status**: Test shows "Output differs" but specific differences need to be extracted

## Key Insight from JavaScript Implementation

The JavaScript `parser.js` `ruleProperty()` function uses a regex that matches whitespace but NOT comments:

```javascript
// packages/less/src/less/parser/parser.js
ruleProperty: function () {
    const simpleProperty = parserInput.$re(/^([_a-zA-Z0-9-]+)\s*:/);
    // Note: \s* matches WHITESPACE ONLY, not comments!
    // The regex is atomic - it either matches or doesn't
    // Comments are handled separately by the parser infrastructure
}
```

**Critical Point**: The regex `/^([_a-zA-Z0-9-]+)\s*:/` matches:
- Property name: `[_a-zA-Z0-9-]+`
- Optional whitespace: `\s*`
- Colon: `:`

Comments are NOT matched by `\s*`, so they cause the simple property match to fail. The parser must have a fallback mechanism or comment stripping phase.

## Key Files to Investigate

### Go Implementation:
- **`packages/less/src/less/less_go/parser.go`** - `RuleProperty()` method
  - Search for: `func.*RuleProperty`
  - This is where property parsing happens

- **`packages/less/src/less/less_go/parser.go`** - Comment handling
  - Search for: `Comment\(\)` method
  - Check: `CommentsReset()` or similar comment management

- **`packages/less/src/less/less_go/declaration.go`** - Declaration GenCSS
  - How declarations output their names and values
  - May filter comments during output phase

### JavaScript Reference:
- **`packages/less/src/less/parser/parser.js`** - Lines ~750-850 for `ruleProperty()`
- **`packages/less/src/less/parser/parser.js`** - Comment handling infrastructure
- **`packages/less/src/less/tree/declaration.js`** - How declarations are output

## Investigation Steps

### Step 1: Understand Current Parsing Behavior

```bash
# Run the test and capture full output
go test -v -run "TestIntegrationSuite/main/comments$" ./packages/less/src/less/less_go/... 2>&1 > comments_test_output.txt

# Look for the specific property with comments
grep -A 5 -B 5 "color.*grey.*orange" comments_test_output.txt
```

### Step 2: Find RuleProperty Implementation

```bash
# Find the RuleProperty function in Go
grep -n "func.*RuleProperty" packages/less/src/less/less_go/parser.go

# Find the JavaScript equivalent
grep -n "ruleProperty.*function" packages/less/src/less/parser/parser.js
```

### Step 3: Compare Implementations

Look for differences in:
1. **Regex patterns**: Does Go's pattern accidentally match comments?
2. **Comment stripping**: Does JavaScript strip comments before property matching?
3. **Fallback logic**: If simple property match fails, what happens next?

### Step 4: Check Comment Infrastructure

```bash
# Find comment-related methods in Go parser
grep -n "Comment\|comment" packages/less/src/less/less_go/parser.go | head -30

# Check if there's a comment buffer or comment stripping phase
grep -n "CommentsReset\|commentStore\|getComments" packages/less/src/less/less_go/parser.go
```

## Hypotheses to Test

### Hypothesis 1: Parser Regex Issue
The Go `RuleProperty()` regex might be matching comments as part of whitespace or the property name.

**Check**:
```go
// Look for the regex pattern in RuleProperty
// It should be something like: /^([_a-zA-Z0-9-]+)\s*:/
// Make sure it doesn't include comment characters
```

### Hypothesis 2: Comment Stripping Phase
JavaScript might strip comments from the input stream before attempting property matching.

**Check**:
- Does JavaScript have a `CommentsReset()` or similar before property parsing?
- Does Go's parser need to skip comments explicitly?

### Hypothesis 3: Two-Phase Property Matching
JavaScript might try:
1. Simple property match (no comments allowed)
2. If fails, try complex property match that handles comments differently

**Check**:
- Look for fallback logic in JavaScript's `ruleProperty()`
- See if there's a `complexProperty()` or similar fallback

### Hypothesis 4: Output Phase Filtering
Comments might be stored in the AST but filtered out during CSS generation.

**Check**:
- `declaration.go` `GenCSS()` method
- See if property names can contain comment nodes
- Check if comments are stripped during output

## Recommended Fix Approaches

### Approach 1: Parser-Level Fix (Most Likely)
**Where**: `parser.go` `RuleProperty()` method

**Solution**: Match JavaScript behavior exactly:
1. Try simple regex match: `/^([_a-zA-Z0-9-]+)\s*:/`
2. If successful, return immediately (no comments in between)
3. If failed, either:
   - Skip comments and retry, OR
   - Fall back to complex property matching that handles comments

**Example structure**:
```go
func (p *Parsers) RuleProperty() any {
    // Try simple property match first
    simpleMatch := p.parser.parserInput.Re(regexp.MustCompile(`^([_a-zA-Z0-9-]+)\s*:`))
    if simpleMatch != nil {
        return simpleMatch // Success - no comments in between
    }

    // If failed, might need to handle comments specially
    // Check JavaScript implementation for fallback logic
    // ...
}
```

### Approach 2: Comment Stripping (Less Likely)
**Where**: Before property parsing

**Solution**: Strip or skip comments in the input stream before attempting property match

### Approach 3: AST-Level Fix (Least Likely)
**Where**: `declaration.go` during construction or output

**Solution**: Filter comment nodes from property names

## Important Constraints

⚠️ **DO NOT**:
- Skip comments during ALL parsing (breaks other features)
- Modify comment parsing itself (comments2 test passes perfectly)
- Remove comments globally (they're needed in values and other contexts)

✅ **DO**:
- Match JavaScript's exact behavior for property name parsing
- Preserve comments in property values (after the colon)
- Maintain all currently passing tests

## Verification Commands

```bash
# Run the specific test
go test -v -run "TestIntegrationSuite/main/comments$" ./packages/less/src/less/less_go/...

# Verify comments2 still passes (uses different comment handling)
go test -v -run "TestIntegrationSuite/main/comments2$" ./packages/less/src/less/less_go/...

# Run unit tests (must pass 100%)
pnpm -w test:go:unit

# Run full integration suite
pnpm -w test:go

# Check current perfect match count (should be >= 63)
pnpm -w test:go 2>&1 | grep -c "Perfect match"
```

## Success Criteria

✅ `comments` test achieves perfect match
✅ `comments2` test remains perfect match
✅ ALL previously passing tests remain passing (no regressions)
✅ Perfect matches count >= 63
✅ Unit tests pass 100%

## Test Input Details

From `packages/test-data/less/_main/comments.less`:
- **Line 65**: `color/* survive */ /* me too */: grey, /* blue */ orange;`
- **Line 66**: `-webkit-border-radius: 2px /* webkit only */;`
- **Line 67**: `-moz-border-radius: (2px * 4) /* moz only with operation */;`

Expected output should have comments between property and colon removed, but comments after colon preserved.

## Additional Context

- The codebase is a Go port of less.js maintaining 1:1 functionality
- Current status: 63/185 tests perfect matches (34.1%)
- Integration test suite: `packages/less/src/less/less_go/integration_suite_test.go`
- The `comments2` test passes perfectly, indicating general comment handling works
- This is a specific edge case for comments in property declarations

## Debugging Tip

Add debug output to see what's being parsed:

```go
// In RuleProperty() or wherever property parsing happens
if os.Getenv("LESS_GO_DEBUG") == "1" {
    fmt.Printf("[DEBUG] Property parsing at position %d\n", p.parser.parserInput.GetIndex())
    fmt.Printf("[DEBUG] Next 50 chars: %q\n", p.parser.parserInput.PeekString(50))
}
```

Run with: `LESS_GO_DEBUG=1 go test -v -run ...`

This will show exactly what the parser sees when it encounters properties with comments.
