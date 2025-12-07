#!/usr/bin/env node
/**
 * Script to compile custom LESS test files using Less.js
 * Usage: node scripts/compile-custom-tests.js [testname]
 *
 * If testname is provided, only that test is compiled.
 * Otherwise, all .less files in testdata/less/custom/ are compiled.
 */

const less = require('less');
const fs = require('fs');
const path = require('path');

const lessDir = path.join(__dirname, '../testdata/less/custom');
const cssDir = path.join(__dirname, '../testdata/css/custom');

async function compileFile(lessFile) {
    const lessPath = path.join(lessDir, lessFile);
    const cssFile = lessFile.replace('.less', '.css');
    const cssPath = path.join(cssDir, cssFile);

    const lessContent = fs.readFileSync(lessPath, 'utf8');

    try {
        const result = await less.render(lessContent, {
            filename: lessPath,
            paths: [lessDir, path.join(__dirname, '../testdata/less/_main')],
            relativeUrls: true,
            javascriptEnabled: true,
            silent: true
        });

        fs.writeFileSync(cssPath, result.css);
        console.log(`✓ Compiled ${lessFile} -> ${cssFile}`);
        return true;
    } catch (error) {
        console.error(`✗ Error compiling ${lessFile}: ${error.message}`);
        return false;
    }
}

async function main() {
    const specificTest = process.argv[2];

    // Ensure output directory exists
    if (!fs.existsSync(cssDir)) {
        fs.mkdirSync(cssDir, { recursive: true });
    }

    let files;
    if (specificTest) {
        const lessFile = specificTest.endsWith('.less') ? specificTest : `${specificTest}.less`;
        if (!fs.existsSync(path.join(lessDir, lessFile))) {
            console.error(`File not found: ${lessFile}`);
            process.exit(1);
        }
        files = [lessFile];
    } else {
        files = fs.readdirSync(lessDir)
            .filter(f => f.endsWith('.less'));
    }

    console.log(`Compiling ${files.length} LESS file(s)...\n`);

    let success = 0;
    let failed = 0;

    for (const file of files) {
        const result = await compileFile(file);
        if (result) success++;
        else failed++;
    }

    console.log(`\nDone: ${success} succeeded, ${failed} failed`);
    process.exit(failed > 0 ? 1 : 0);
}

main().catch(console.error);
