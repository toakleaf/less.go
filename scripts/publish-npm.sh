#!/bin/bash
# Publish all npm packages

set -e

VERSION=${1:-$(node -p "require('./npm/less.go/package.json').version")}

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

# Publish platform packages first
for dir in npm/less.go-*/; do
  echo "Publishing $(basename $dir)..."
  (cd "$dir" && npm publish --access public)
done

# Publish main package last
echo "Publishing less.go..."
(cd npm/less.go && npm publish --access public)

echo "Done!"
