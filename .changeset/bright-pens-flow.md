---
"lessgo": minor
"@lessgo/darwin-arm64": minor
"@lessgo/darwin-x64": minor
"@lessgo/linux-x64": minor
"@lessgo/linux-arm64": minor
"@lessgo/win32-x64": minor
"@lessgo/win32-arm64": minor
---

## Bug Fixes

- Fix `!important` handling in variable declarations, merge declarations, and anonymous values
- Fix Anonymous color values in `contrast()` and other color functions
- Preserve `calc()` expressions in mixin default values
- Fix `@page` pseudo-selectors and variable hoisting in imports
- Fix parser handling of unmatched closing braces in `ParseUntil` to prevent over-consumption
- Fix `Sub()` parser to match Less.js behavior for parens-error tests
- Preserve parentheses in parenthesized list values with comma separation
- Fix extend property formatting to match Less.js behavior
- Fix namespace property access evaluation to Dimension for math operations
- Prevent property merge duplication when using property accessors
- Fix arrow function syntax handling with parameters in Less variables
- Preserve frames for variable imports during path interpolation
- Fix Quoted node evaluation inside URL for variable interpolation
- Prevent incorrect deduplication of multi-target extends
- Remove extra whitespace when extend-only rulesets are inside media queries
- Stabilize import-remote test by pinning `@less/test-data@4.4.2`

## Performance Improvements

- Add `ReleaseTree` function for improved object pool utilization
- Convert CSS output methods from slice concatenation to `strings.Builder` for reduced allocations
- Add context map pool for evaluation phase allocations
- Optimize `Selector.CreateDerived` to reduce map allocations
- Optimize `EvalCall` memory allocations in `MixinDefinition`
- Reduce unnecessary map allocations in CSS generation phase

## Testing

- Expand benchmark test suite to include custom integration tests
- Add extensive test variations for overfitting detection across math modes, mixins, and other features
