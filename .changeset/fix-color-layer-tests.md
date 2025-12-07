---
"lessgo": patch
"@lessgo/plugin-vite": patch
---

Fix CSS relative color syntax and @layer parent selector issues

- Add ColorOperand parser for CSS relative color syntax (oklch, hsl, rgb with calc expressions using channel identifiers l, c, h, r, g, b, s)
- Fix @layer parent selector (&) resolution to properly join with parent selectors
