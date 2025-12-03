---
'@lessgo/darwin-arm64': patch
'@lessgo/darwin-x64': patch
'lessgo': patch
'@lessgo/linux-arm64': patch
'@lessgo/linux-x64': patch
'@lessgo/win32-arm64': patch
'@lessgo/win32-x64': patch
'@lessgo/plugin-vite': patch
---

Fix release script to resolve workspace:* protocol for @lessgo/plugin-vite

The npm publish command doesn't understand pnpm's workspace:* protocol, which caused
@lessgo/plugin-vite to be published with "lessgo": "workspace:*" as a dependency.
Now using pnpm pack (which resolves workspace:* to actual versions) before publishing.
