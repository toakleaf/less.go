# Agent: URL Parsing Fix

## 🎯 Your Mission
Fix URL parsing to handle escaped characters like `\(` and `\)` correctly.

## 📊 Status
**Tests to Fix**: 2 (same test in 2 suites)
**Branch**: `claude/fix-urls-<your-session-id>`
**Independence**: HIGH - No conflicts with other agents

## ❌ The Problem
```less
.test {
  background: url(http://fonts.googleapis.com/css?family=\"Rokkitt\":\(400\),700);
}
```
**Error**: `expected ')' got '('`

The parser sees `\(` as an opening paren instead of an escaped character.

## 🔍 Files You'll Modify
- `packages/less/src/less/less_go/url.go` - Primary: URL parsing logic
- `packages/less/src/less/less_go/parser.go` - Likely: url() function around line 3270
- `packages/less/src/less/less_go/parser_input.go` - Maybe: escape handling

**You will NOT conflict with other agents** - they work on different files.

## ✅ Success Criteria
- [ ] `urls` test (main suite) passes
- [ ] `urls` test (compression suite) passes
- [ ] All unit tests pass: `pnpm -w test:go:unit`
- [ ] No regressions: `pnpm -w test:go:summary`

## 🧪 Test Commands
```bash
# Minimal test case
cat > /tmp/test-url.less << 'EOF'
.test { background: url(http://example.com/css?family=\"Font\":\(400\),700); }
EOF
go run cmd/lessc/lessc.go /tmp/test-url.less

# Run actual tests
cd packages/less/src/less/less_go
go test -run "TestIntegrationSuite/main/urls" -v
go test -run "TestIntegrationSuite/compression/urls" -v

# Verify no regressions
pnpm -w test:go:unit
pnpm -w test:go:summary
```

## 🔑 Key Insight
Look at the regex in `parser.go` around line 3286:
```go
urlMatch := e.parsers.parser.parserInput.Re(regexp.MustCompile(`^(?:(?:\\[()'""])|[^()'""])+`))
```

This regex is supposed to match escaped characters `\\[()'"]`, but the parser might not be handling the result correctly. The backslash should escape the next character, making it literal.

## 📋 When Done
```bash
git add packages/less/src/less/less_go/*.go
git commit -m "Fix URL parsing with escaped characters

URLs containing escaped parentheses like \( were causing parser errors.
Fixed escape handling to treat backslash-escaped characters as literals.

Tests fixed:
- urls (main suite): ✅
- urls (compression suite): ✅"

git push -u origin claude/fix-urls-<your-session-id>
```

Report: "Fixed 2/2 URL tests. Ready for PR."
