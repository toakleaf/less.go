#!/bin/bash
# Publish all npm packages

set -e

VERSION=${1:-$(node -p "require('./npm/lessgo/package.json').version")}

echo "Publishing version $VERSION..."

# Update versions in all packages
for pkg in npm/*/package.json; do
  node -e "
    const fs = require('fs');
    const pkg = JSON.parse(fs.readFileSync('$pkg'));
    pkg.version = '$VERSION';
    if (pkg.optionalDependencies) {
      for (const dep of Object.keys(pkg.optionalDependencies)) {
        pkg.optionalDependencies[dep] = '$VERSION';
      }
    }
    fs.writeFileSync('$pkg', JSON.stringify(pkg, null, 2) + '\n');
  "
done

# Publish platform packages first (scoped packages under @lessgo)
for dir in npm/darwin-* npm/linux-* npm/win32-*; do
  if [ -d "$dir" ]; then
    echo "Publishing $(basename $dir)..."
    (cd "$dir" && npm publish --access public --provenance)
  fi
done

# Publish main package last (unscoped 'lessgo')
echo "Publishing lessgo..."
(cd npm/lessgo && npm publish --access public --provenance)

echo "Done!"
