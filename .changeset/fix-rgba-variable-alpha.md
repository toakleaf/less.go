---
"lessgo": patch
"@lessgo/darwin-arm64": patch
"@lessgo/darwin-x64": patch
"@lessgo/linux-x64": patch
"@lessgo/linux-arm64": patch
"@lessgo/win32-x64": patch
"@lessgo/win32-arm64": patch
"@lessgo/plugin-vite": patch
---

Fix rgba() ignoring variable alpha values

The rgba() function (and rgb(), hsl(), hsla()) was ignoring alpha values when passed as variables. For example:

```less
@alpha: 0.5;
color: rgba(255, 0, 0, @alpha); // was producing #ff0000 instead of rgba(255, 0, 0, 0.5)
```

This fix ensures that variable arguments are properly evaluated before being passed to color functions.
