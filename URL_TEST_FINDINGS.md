# URL Test Failures Analysis

## Summary
Three URL-related test suites are failing with output differences:
1. **main/urls** - Main URL handling test (largest, most comprehensive)
2. **static-urls/urls** - URL handling with static paths and rootpath option
3. **url-args/urls** - URL handling with urlArgs appended to URLs

**Key Finding**: All three tests have different issues, suggesting the problem is in different parts of the URL processing pipeline.

---

## Test 1: main/urls

### Issue: Import statements are missing from output
**Expected**: Output starts with import statements
```css
@import "css/background.css";
@import "import/import-test-d.css";
@import "file.css";
```

**Actual**: Output starts directly with @font-face (imports are skipped)
```css
@font-face {
  src: url("/fonts/garamond-pro.ttf");
```

### Root Path Rewriting Issues
1. **Relative URL issue** (line 42):
   - Expected: `url(../data/image.jpg);`
   - Actual: `url(../../data/image.jpg);`
   - The path is being rewritten with one too many `../` segments

2. **Quoted relative URL issue** (line 43):
   - Expected: `url('../data/image.jpg');`
   - Actual: `url('../../data/image.jpg');`
   - Same path rewriting problem

### Common Issues in all three:
- URL path rewriting logic is adding extra parent directory traversals
- Data URI URLs appear to be correctly handled (no changes)
- Absolute URLs are preserved correctly
- Variable substitution in URLs works

---

## Test 2: static-urls/urls

### Issue 1: Missing Import Statements
**Expected**: First lines are imports:
```css
@import "css/background.css";
@import "folder (1)/import-test-d.css";
```

**Actual**: Import statements are completely missing; output starts with @font-face

### Issue 2: Path Escaping
With `rootpath: "folder (1)/"` option, spaces in paths should be escaped:
- Expected: `url(folder\ \(1\)/fonts.svg#MyGeometricModern)`
- Actual: Same (this part works correctly)

### Issue 3: Relative Path Rewriting
- Expected: `url(../data/image.jpg);` and `url('../data/image.jpg');`
- Actual: `url(../../data/image.jpg);` and `url('../../data/image.jpg');`
- Same "too many parent directories" bug

**Root cause**: When processing relative URLs, the system is adding an extra `../` when it should not.

---

## Test 3: url-args/urls

### Issue 1: Missing import statement pattern
Unlike the other tests, this one doesn't show import statements at the start, but that's expected from this test file.

### Issue 2: URL argument appending (urlArgs: "424242")
**Expected behavior**: Query string `?424242` appended to non-data-URI URLs

**Example - line 2**:
- Expected: `url("/fonts/garamond-pro.ttf?424242");`
- Actual: `url("/fonts/garamond-pro.ttf");`
- **urlArgs NOT being appended**

**Example - line 6**:
- Expected: `url("http://www.lesscss.org/spec.html?424242")`
- Actual: `url("http://www.lesscss.org/spec.html")`
- **urlArgs NOT being appended**

**Example - line 19-20**:
- Expected: URLs with existing query strings get `&424242`:
  ```
  url(http://fonts.googleapis.com/css?family=\"Rokkitt\":\(400\),700&424242)
  ```
- Actual: Query string is not modified

**Data URIs correctly NOT modified**:
- `url(data:image/png;charset=utf-8;base64,...)` - NOT modified (correct)
- `url(data:image/x-png,...)` - NOT modified (correct)

---

## Common Patterns Across All Tests

### What's Working ✅
1. Data URI URLs - completely preserved
2. Absolute URLs (http/https) - correctly handled
3. URL syntax (parentheses, quotes) - preserved
4. Variable substitution in URLs - works
5. Multiple URLs in same property - all processed
6. Hash fragments in URLs - preserved

### What's NOT Working ❌

1. **Import Statement Processing**
   - `@import` directives disappear from output
   - Should appear at the top of CSS output
   - Root cause: Likely in import collection/output ordering

2. **Relative Path Rewriting** 
   - When rootpath is relative (like `../data/`), paths get one extra `../`
   - Problem: `../data/image.jpg` becomes `../../data/image.jpg`
   - This affects relative URLs in imported files
   - Root cause: Likely in path resolution logic when handling relative includes

3. **urlArgs Appending**
   - `urlArgs` option value not being appended to URLs
   - Should add query string to all non-data-URI URLs
   - Root cause: urlArgs handling not implemented or broken in URL processing

---

## Technical Investigation Needed

### For Import Issue:
- Check: How imports are collected during parsing
- Check: CSS output generation order (imports should be first)
- Files to examine: Import handling in render pipeline

### For Relative Path Issue:
- Check: Path normalization in relative URL rewriting
- The bug manifests as: `../` + `../data/` = `../../data/` (one too many)
- Should be: `../data/` remains `../data/`
- Files to examine: URL rewriting logic, path resolution

### For urlArgs Issue:
- Check: Where urlArgs option is processed
- Check: URL processing pipeline for query string appending
- Should append `?value` or `&value` (if query string exists) to all URLs except data URIs
- Files to examine: URL processing functions, options handling

---

## Test Execution Details

### Command Used:
```bash
LESS_GO_DIFF=1 go test -run "TestIntegrationSuite/[suite]/urls$" -v
```

### Test Options by Suite:
1. **main/urls**:
   - `relativeUrls: true`
   - `silent: true`
   - `javascriptEnabled: true`

2. **static-urls/urls**:
   - `math: "strict"`
   - `relativeUrls: false`
   - `rootpath: "folder (1)/"`

3. **url-args/urls**:
   - `urlArgs: "424242"`

