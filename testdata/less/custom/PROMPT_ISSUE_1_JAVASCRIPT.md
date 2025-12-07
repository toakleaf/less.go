# Fix Issue: JavaScript Evaluation Not Enabled for Custom Tests

## Problem Description

When running custom integration tests with `javascriptEnabled: true` in the options, the JavaScript evaluation context is not receiving the option. The test fails with:

```
Syntax: JavaScript: inline JavaScript is not enabled. Is it set in your options?
```

Debug output shows the option is set in the initial options but not propagated to the evaluation context:

```
[DEBUG] Starting compilation with options: map[... javascriptEnabled:true ...]
[DEBUG JsEvalNode] *Eval context, JavascriptEnabled=false, PluginBridge=<nil>
```

## Reproduction Steps

1. Create a test file `testdata/less/custom/var-javascript.less` with inline JavaScript:
```less
.js-basic {
  answer: `42`;
}
```

2. Run the custom tests:
```bash
LESS_GO_DEBUG=1 LESS_GO_CUSTOM_ONLY=1 go test ./less -v -run TestIntegrationSuite/custom/var-javascript
```

3. Observe the error: "JavaScript: inline JavaScript is not enabled"

## Expected Behavior

The `javascriptEnabled: true` option should be properly propagated through the compilation pipeline to the evaluation context, allowing inline JavaScript to be evaluated.

## Investigation Areas

1. **Check how custom tests set options:**
   - Look in `less/integration_suite_test.go` for how custom tests configure compilation options
   - Verify the `javascriptEnabled` option is being set correctly

2. **Trace the options flow:**
   - From `less.Render()` or `less.Parse()` entry points
   - Through `createParse` and `createRender` functions
   - To the `Eval` context creation

3. **Look at the JavaScript node evaluation:**
   - File: `less/tree_javascript.go` (or similar)
   - Check how `JsEvalNode.Eval()` accesses `context.JavascriptEnabled`

4. **Compare with how main integration tests handle JavaScript:**
   - The `javascript.less` test in `_main` suite should work
   - Check if there's a difference in how options are passed for main vs custom tests

## Files to Examine

- `less/integration_suite_test.go` - Test setup and option configuration
- `less/render.go` - Render function and option handling
- `less/parse.go` - Parse function and option handling
- `less/contexts.go` - Eval context definition
- `less/tree_javascript.go` - JavaScript node evaluation

## Verification

After fixing, run:
```bash
LESS_GO_CUSTOM_ONLY=1 go test ./less -v -run TestIntegrationSuite/custom/var-javascript
```

The test should compile successfully and produce output matching the expected CSS.
