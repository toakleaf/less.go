# Custom Integration Tests

Add your own LESS integration tests here. Each test requires two files:

1. **Input file**: `<name>.less` in this directory (`testdata/less/custom/`)
2. **Expected output**: `<name>.css` in `testdata/css/custom/`

## Example

To create a test named `my-feature`:

1. Create `testdata/less/custom/my-feature.less`:
```less
@color: #333;
.example {
  color: @color;
}
```

2. Create `testdata/css/custom/my-feature.css`:
```css
.example {
  color: #333;
}
```

## Running Custom Tests

```bash
# Run only custom tests
LESS_GO_CUSTOM_ONLY=1 go test ./less -v -run TestIntegrationSuite -timeout 5m

# Run all tests including custom
pnpm test:go:all

# Skip custom tests
LESS_GO_SKIP_CUSTOM=1 pnpm test:go:all

# Debug a specific custom test
LESS_GO_DEBUG=1 go test -v -run TestIntegrationSuite/custom/<testname> ./less
```

## Options

Custom tests run with these default options:
- `relativeUrls: true`
- `silent: true`
- `javascriptEnabled: true`
