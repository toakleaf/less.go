# URL Processing Issues - Agent Task

## 🎯 Mission
Fix 2 failing tests related to URL parsing and processing

## 📊 Status
- **Tests Failing**: 2 (same test in different suites)
- **Priority**: Medium (Wave 1 - Independent)
- **Complexity**: Low-Medium
- **Independence**: HIGH - Can be fixed in parallel with other issues

## ❌ Failing Tests

### 1. urls (main suite)
**File**: `packages/test-data/less/_main/urls.less`
**Error**: `expected ')' got '(' in ../../../../test-data/less/_main/urls.less`

### 2. urls (compression suite)
**File**: `packages/test-data/less/compression/urls.less`
**Error**: Same as above (likely same content or similar)

**Test Content** (first 50 lines):
```less
@import "nested-gradient-with-svg-gradient/mixin-consumer.less";

@font-face {
  src: url("/fonts/garamond-pro.ttf");
  src: local(Futura-Medium),
       url(fonts.svg#MyGeometricModern) format("svg");
  not-a-comment: url(//z);
}

#shorthands {
  background: url("http://www.lesscss.org/spec.html") no-repeat 0 4px;
  background: url("img.jpg") center / 100px;
  background: #fff url(image.png) center / 1px 100px repeat-x scroll content-box padding-box;
}

#misc {
  background-image: url(images/image.jpg);
}

#data-uri {
  background: url(data:image/png;charset=utf-8;base64,
    kiVBORw0KGgoAAAANSUhEUgAAABAAAAAQAQMAAAAlPW0iAAAABlBMVEUAAAD/
    k//+l2Z/dAAAAM0lEQVR4nGP4/5/h/1+G/58ZDrAz3D/McH8yw83NDDeNGe4U
    kg9C9zwz3gVLMDA/A6P9/AFGGFyjOXZtQAAAAAElFTkSuQmCC);
  background-image: url(data:image/x-png,f9difSSFIIGFIFJD1f982FSDKAA9==);
  background-image: url(http://fonts.googleapis.com/css?family=\"Rokkitt\":\(400\),700);
  background-image: url("http://fonts.googleapis.com/css?family=\"Rokkitt\":\(400\),700");
}

#svg-data-uri {
  background: transparent url('data:image/svg+xml, <svg version="1.1"><g></g></svg>');
}
```

**Problem**: Parser error "expected ')' got '('" suggests the URL parser is failing on:
- Escaped parentheses in URLs: `\(400\)`
- Complex URLs with query parameters
- URLs containing parentheses that should be treated as part of the URL, not as syntax

**Most Likely Culprit**: Line 23
```less
background-image: url(http://fonts.googleapis.com/css?family=\"Rokkitt\":\(400\),700);
```

This URL contains:
- Escaped quotes: `\"`
- Escaped parentheses: `\(` and `\)`
- The parser might be seeing the `\(` as a syntax element instead of part of the URL string

---

## 🔍 Root Cause Analysis

### Theory 1: Escaped Character Handling
The URL parser doesn't properly handle escaped characters inside URLs. When it encounters `\(`, it might be:
- Treating it as a new opening parenthesis
- Not recognizing it as an escaped character
- Tokenizing it incorrectly

### Theory 2: URL Content Parsing
URLs can contain almost any character when escaped. The parser might be:
- Too strict about what characters are allowed
- Not handling the escaping correctly
- Confusing URL content with LESS syntax

### Theory 3: Quote Handling
The URL contains both escaped quotes (`\"`) and escaped parens (`\(`). The combination might be:
- Causing the parser to lose track of string boundaries
- Mixing up URL content with LESS expressions

---

## 🔍 Go Files to Investigate

### Primary Files
1. **`parser.go`** - Main parser implementation
   - Look for URL parsing logic
   - Check how `url()` is parsed
   - Find the paren matching logic

2. **`url.go`** - URL node implementation
   - Check URL parsing method
   - Verify escape character handling
   - Look at quote handling

3. **`parser_input.go`** - Input tokenization
   - Check character escaping
   - Verify paren/quote tracking
   - Look for escape sequence handling

### Secondary Files
4. **`quoted.go`** - Quoted string handling
   - URLs can contain quoted strings
   - Check if escaping is handled

5. **`chunker.go`** - Tokenization
   - Might handle character escaping
   - Check paren/bracket matching

---

## 📚 JavaScript Reference Files

Study these to understand correct behavior:

1. **`packages/less/src/less/parser/parser.js`**
   - Look for `url()` parsing
   - Check escape character handling
   - Find paren matching logic

2. **`packages/less/src/less/tree/url.js`**
   - URL node implementation
   - Value parsing and escaping

3. **`packages/less/src/less/parser/parser-input.js`**
   - Input character handling
   - Escape sequences

---

## ✅ Success Criteria

### Minimum Success (1/2 tests)
- `urls` (main suite) - URL with escaped characters parses correctly

### Target Success (2/2 tests)
- `urls` (main suite) - All URL formats work
- `urls` (compression suite) - Same fixes apply

---

## 🚫 Constraints

1. **NEVER modify any .js files**
2. **Must pass unit tests**: `pnpm -w test:go:unit`
3. **Must pass target tests**: `pnpm -w test:go:filter -- "urls"`
4. **No regressions**: All currently passing tests must still pass

---

## 🧪 Testing Strategy

### Create Minimal Test Case
```bash
cat > /tmp/test-url.less << 'EOF'
.test {
  /* Simple URL - should work */
  background: url(image.png);

  /* URL with escaped parens - likely fails */
  background: url(http://example.com/test\(400\).png);

  /* The actual failing case */
  background-image: url(http://fonts.googleapis.com/css?family=\"Rokkitt\":\(400\),700);
}
EOF

# Test it
go run cmd/lessc/lessc.go /tmp/test-url.less
```

### Run Specific Tests
```bash
# Test individual cases
go test -run "TestIntegrationSuite/main/urls" -v
go test -run "TestIntegrationSuite/compression/urls" -v

# With debug output
LESS_GO_DEBUG=1 go test -run "TestIntegrationSuite/main/urls" -v
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

1. **url.go** - Fix URL content parsing
   - Handle escaped characters: `\(`, `\)`, `\"`
   - Don't treat escaped parens as syntax
   - Properly parse URL content until closing `)`

2. **parser.go** - Fix URL parsing logic
   - When parsing `url(...)`, handle escapes
   - Track whether parens are escaped
   - Don't count escaped parens in bracket matching

3. **parser_input.go** - Ensure escape sequences work
   - Backslash should escape the next character
   - Escaped characters should not participate in syntax

### Testing Pattern

For each fix:
1. Test with minimal case
2. Test with full urls.less
3. Run unit tests
4. Run integration suite
5. Commit with clear message

---

## 🎯 Debugging Hints

### Find the Exact Line
```bash
# Get line number of failure
LESS_GO_DEBUG=1 go test -run "TestIntegrationSuite/main/urls" -v 2>&1 | grep "expected ')'"

# Look at that specific line in the test file
sed -n '23p' packages/test-data/less/_main/urls.less
```

### Add Parser Debug Output
```go
// In url.go or parser.go where URL is parsed
fmt.Printf("[URL-DEBUG] Parsing URL content: %q\n", currentChar)
fmt.Printf("[URL-DEBUG] Is escaped: %v\n", isEscaped)
fmt.Printf("[URL-DEBUG] Paren depth: %d\n", parenDepth)
```

### Check JavaScript Implementation
```bash
# Test the same URL in JavaScript less.js to verify expected behavior
node -e "
const less = require('less');
const input = '.test { background: url(http://fonts.googleapis.com/css?family=\"Rokkitt\":\\(400\\),700); }';
less.render(input).then(output => console.log(output.css));
"
```

---

## 📊 Estimated Impact

- **Tests Fixed**: 2 (same test in 2 suites)
- **Other Tests Potentially Improved**: 3-5 tests using complex URLs
- **Risk Level**: Low - URL parsing is isolated, unlikely to affect other features

---

## 🔄 Iteration Strategy

### Round 1: Identify Exact Failure Point
1. Find which line causes "expected ')' got '('"
2. Create minimal reproduction
3. Understand why parser fails

### Round 2: Fix Escape Handling
1. Implement proper escape character handling in URLs
2. Test with minimal case
3. Test with full urls.less

### Round 3: Verify and Commit
1. Run unit tests
2. Run integration tests
3. Verify both urls tests pass
4. Commit and push

---

## 📋 Commit Message Template

```
Fix URL parsing with escaped characters

URLs containing escaped parentheses and quotes were causing parser errors
like "expected ')' got '('". This affected URLs like:
  url(http://example.com/css?family=\"Font\":\(400\),700)

Root cause: [Describe issue - likely escape character handling]

Fix:
- Properly handle backslash-escaped characters in URL content
- Don't treat escaped parens as syntax elements
- Maintain proper paren/quote matching

Tests fixed:
- urls (main suite): ✅
- urls (compression suite): ✅
```

---

## 🚀 When Done

1. **Commit** to branch: `claude/fix-urls-<your-session-id>`
2. **Push** to remote: `git push -u origin claude/fix-urls-<your-session-id>`
3. **Report**: "Fixed 2/2 urls tests (main + compression suites)"

---

## 💡 Key Insights

1. **Parser error, not evaluation error** - This is a parsing issue, not runtime
2. **Isolated issue** - URL parsing is contained, low risk of regressions
3. **Well-defined problem** - "expected ')' got '('" is very specific
4. **Likely simple fix** - Probably just need to handle escape sequences
5. **Tests both suites** - Same fix should resolve both failing tests

---

## 🔗 Related Tests

Other tests that use URLs (currently passing, verify these don't break):
- `css-escapes` - CSS escape sequences
- `data-uri` - Data URI handling
- Various tests with `url()` function

---

## ⚠️ Special Notes

1. **This is a PARSER issue** - Different from the runtime issues that have been fixed
2. **Escape handling is critical** - Don't break existing escape sequences
3. **URLs are everywhere** - Be careful not to break other URL parsing
4. **Test with data URIs too** - They have complex content in URLs
5. **Both test suites will pass** - Same fix applies to main and compression suites
