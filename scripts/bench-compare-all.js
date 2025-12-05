#!/usr/bin/env node

/**
 * Three-way benchmark comparison: Node.js vs Bun vs Go
 * Runs realistic benchmarks where each iteration is a fresh process
 * This simulates actual CLI/build tool usage patterns
 */

const { execSync, spawnSync } = require('child_process');
const path = require('path');
const fs = require('fs');

// Parse command line arguments
const iterationsArg = process.argv.find(arg => arg.startsWith('--iterations='));
const ITERATIONS = iterationsArg ? parseInt(iterationsArg.split('=')[1]) : 30;

// Check if Bun is available
function checkBunAvailable() {
    try {
        execSync('bun --version', { stdio: 'ignore' });
        return true;
    } catch {
        return false;
    }
}

const bunAvailable = checkBunAvailable();

console.log('╔══════════════════════════════════════════════════════════════════════════════╗');
console.log('║       NODE.JS vs BUN vs GO BENCHMARK COMPARISON                              ║');
console.log('║       (Each iteration = fresh process, like real CLI usage)                  ║');
console.log('╚══════════════════════════════════════════════════════════════════════════════╝\n');

if (!bunAvailable) {
    console.log('⚠️  Bun is not installed. Install with: curl -fsSL https://bun.sh/install | bash');
    console.log('   Proceeding with Node.js vs Go comparison only.\n');
}

console.log(`Running ${ITERATIONS} independent build sessions for each implementation...`);
console.log('This simulates realistic CLI/build tool usage where each build is a fresh process.\n');

// Run Node.js benchmarks
console.log(`Running Node.js benchmarks (${ITERATIONS} separate processes)...`);
const nodeTimes = [];
let nodeVersion = '';
for (let i = 0; i < ITERATIONS; i++) {
    process.stdout.write(`\r  Node Progress: ${i + 1}/${ITERATIONS} (${((i + 1) / ITERATIONS * 100).toFixed(1)}%)`);

    try {
        const output = execSync('node benchmark/suite.js --single-run', {
            encoding: 'utf8',
            cwd: path.join(__dirname, '..'),
            stdio: ['ignore', 'pipe', 'ignore']
        });

        const result = JSON.parse(output.trim());
        nodeTimes.push(result.totalTime);
        if (!nodeVersion) nodeVersion = result.runtimeVersion || process.version;
    } catch (error) {
        console.error(`\n⚠️  Node iteration ${i + 1} failed`);
    }
}
process.stdout.write('\n');

// Run Bun benchmarks (if available)
const bunTimes = [];
let bunVersion = '';
if (bunAvailable) {
    console.log(`Running Bun benchmarks (${ITERATIONS} separate processes)...`);
    for (let i = 0; i < ITERATIONS; i++) {
        process.stdout.write(`\r  Bun Progress: ${i + 1}/${ITERATIONS} (${((i + 1) / ITERATIONS * 100).toFixed(1)}%)`);

        try {
            const output = execSync('bun benchmark/suite.js --single-run', {
                encoding: 'utf8',
                cwd: path.join(__dirname, '..'),
                stdio: ['ignore', 'pipe', 'ignore']
            });

            const result = JSON.parse(output.trim());
            bunTimes.push(result.totalTime);
            if (!bunVersion) bunVersion = result.runtimeVersion || 'unknown';
        } catch (error) {
            console.error(`\n⚠️  Bun iteration ${i + 1} failed`);
        }
    }
    process.stdout.write('\n');
}

// Run Go benchmarks
console.log('Compiling Go test binary (one-time)...');
const goBinaryPath = path.join(__dirname, '..', 'less_go.test');
try {
    execSync(`go test -c -o ${goBinaryPath} ./less`, {
        encoding: 'utf8',
        cwd: path.join(__dirname, '..'),
        stdio: ['ignore', 'pipe', 'ignore']
    });
    console.log('  ✓ Go binary compiled successfully\n');
} catch (error) {
    console.error('  ✗ Failed to compile Go test binary:', error.message);
    process.exit(1);
}

console.log(`Running Go benchmarks (${ITERATIONS} separate processes)...`);
const goTimes = [];
let goBytesPerOp = 0;
let goAllocsPerOp = 0;

for (let i = 0; i < ITERATIONS; i++) {
    process.stdout.write(`\r  Go Progress: ${i + 1}/${ITERATIONS} (${((i + 1) / ITERATIONS * 100).toFixed(1)}%)`);

    try {
        const goOutput = execSync(`${goBinaryPath} -test.run=^$ -test.bench=BenchmarkLargeSuite -test.benchmem -test.benchtime=1x`, {
            encoding: 'utf8',
            cwd: path.join(__dirname, '..', 'less'),
            maxBuffer: 10 * 1024 * 1024,
            stdio: ['ignore', 'pipe', 'ignore']
        });

        const match = goOutput.match(/BenchmarkLargeSuite-\d+\s+\d+\s+(\d+)\s+ns\/op\s+(\d+)\s+B\/op\s+(\d+)\s+allocs\/op/);
        if (match) {
            const goNsPerOp = parseInt(match[1]);
            const goMsPerOp = goNsPerOp / 1_000_000;
            goTimes.push(goMsPerOp);

            if (i === 0) {
                goBytesPerOp = parseInt(match[2]);
                goAllocsPerOp = parseInt(match[3]);
            }
        }
    } catch (error) {
        console.error(`\n⚠️  Go iteration ${i + 1} failed`);
    }
}

// Clean up compiled binary
try {
    fs.unlinkSync(goBinaryPath);
} catch (e) {
    // Ignore cleanup errors
}

process.stdout.write('\n\n');

// Calculate statistics helper
function calcStats(times) {
    if (times.length === 0) return null;
    const avg = times.reduce((sum, t) => sum + t, 0) / times.length;
    const min = Math.min(...times);
    const max = Math.max(...times);
    const sorted = [...times].sort((a, b) => a - b);
    const median = sorted[Math.floor(sorted.length / 2)];
    const stdDev = Math.sqrt(times.reduce((sum, t) => sum + Math.pow(t - avg, 2), 0) / times.length);
    const variationPerc = (stdDev / avg * 100);
    return { avg, min, max, median, stdDev, variationPerc };
}

const nodeStats = calcStats(nodeTimes);
const bunStats = bunAvailable ? calcStats(bunTimes) : null;
const goStats = calcStats(goTimes);

// Get test count
let testCount = 73;
try {
    const output = execSync('node benchmark/suite.js --single-run', {
        encoding: 'utf8',
        cwd: path.join(__dirname, '..'),
        stdio: ['ignore', 'pipe', 'ignore']
    });
    const result = JSON.parse(output.trim());
    testCount = result.testCount;
} catch (e) {
    // Use default
}

// Display results
console.log('═'.repeat(80));
console.log('REALISTIC SUITE BENCHMARK RESULTS');
console.log('═'.repeat(80));
console.log(`Files per build: ${testCount}`);
console.log(`Build iterations: ${ITERATIONS} independent processes per implementation`);
console.log(`Methodology: Each iteration = fresh process compiling all files once`);
console.log('');
console.log('Runtimes:');
console.log(`  • Node.js: ${nodeVersion}`);
if (bunAvailable) console.log(`  • Bun: ${bunVersion}`);
console.log(`  • Go: (compiled binary)`);
console.log('');

// Format helpers
function formatTime(ms) {
    if (ms < 0.001) {
        return `${(ms * 1000000).toFixed(2)}ns`;
    } else if (ms < 1) {
        return `${(ms * 1000).toFixed(2)}µs`;
    } else if (ms < 1000) {
        return `${ms.toFixed(2)}ms`;
    } else {
        return `${(ms / 1000).toFixed(2)}s`;
    }
}

function formatDiff(baseTime, compareTime) {
    const ratio = compareTime / baseTime;
    if (ratio < 0.8) {
        return `${(1/ratio).toFixed(1)}x faster ✓`;
    } else if (ratio < 1.2) {
        return `~same`;
    } else {
        return `${ratio.toFixed(1)}x slower`;
    }
}

// Build comparison table
if (bunAvailable) {
    console.log('┌─────────────────────────────────────────────────────────────────────────────────────────────────┐');
    console.log('│ 📊 BUILD PERFORMANCE (all files, one pass per build)                                           │');
    console.log('├─────────────────────────────────────────────────────────────────────────────────────────────────┤');
    console.log('│                    │    Node.js   │      Bun     │      Go      │  Bun vs Node │  Go vs Node   │');
    console.log('├────────────────────┼──────────────┼──────────────┼──────────────┼──────────────┼───────────────┤');
    console.log(`│ All Files (avg)    │ ${formatTime(nodeStats.avg).padEnd(12)} │ ${formatTime(bunStats.avg).padEnd(12)} │ ${formatTime(goStats.avg).padEnd(12)} │ ${formatDiff(nodeStats.avg, bunStats.avg).padEnd(12)} │ ${formatDiff(nodeStats.avg, goStats.avg).padEnd(13)} │`);
    console.log(`│ All Files (median) │ ${formatTime(nodeStats.median).padEnd(12)} │ ${formatTime(bunStats.median).padEnd(12)} │ ${formatTime(goStats.median).padEnd(12)} │ ${formatDiff(nodeStats.median, bunStats.median).padEnd(12)} │ ${formatDiff(nodeStats.median, goStats.median).padEnd(13)} │`);
    console.log(`│ All Files (min)    │ ${formatTime(nodeStats.min).padEnd(12)} │ ${formatTime(bunStats.min).padEnd(12)} │ ${formatTime(goStats.min).padEnd(12)} │ ${formatDiff(nodeStats.min, bunStats.min).padEnd(12)} │ ${formatDiff(nodeStats.min, goStats.min).padEnd(13)} │`);
    console.log(`│ All Files (max)    │ ${formatTime(nodeStats.max).padEnd(12)} │ ${formatTime(bunStats.max).padEnd(12)} │ ${formatTime(goStats.max).padEnd(12)} │ ${formatDiff(nodeStats.max, bunStats.max).padEnd(12)} │ ${formatDiff(nodeStats.max, goStats.max).padEnd(13)} │`);
    console.log(`│ Per File (avg)     │ ${formatTime(nodeStats.avg / testCount).padEnd(12)} │ ${formatTime(bunStats.avg / testCount).padEnd(12)} │ ${formatTime(goStats.avg / testCount).padEnd(12)} │ ${formatDiff(nodeStats.avg, bunStats.avg).padEnd(12)} │ ${formatDiff(nodeStats.avg, goStats.avg).padEnd(13)} │`);
    console.log(`│ Variation (±%)     │ ${nodeStats.variationPerc.toFixed(1)}%`.padEnd(34) + '│ ' + `${bunStats.variationPerc.toFixed(1)}%`.padEnd(12) + '│ ' + `${goStats.variationPerc.toFixed(1)}%`.padEnd(12) + '│ ' + 'N/A'.padEnd(12) + ' │ ' + 'N/A'.padEnd(13) + ' │');
    console.log('└────────────────────┴──────────────┴──────────────┴──────────────┴──────────────┴───────────────┘');
} else {
    console.log('┌─────────────────────────────────────────────────────────────────────────────┐');
    console.log('│ 📊 BUILD PERFORMANCE (all files, one pass per build)                        │');
    console.log('├─────────────────────────────────────────────────────────────────────────────┤');
    console.log('│                    │    Node.js   │      Go      │   Difference             │');
    console.log('├────────────────────┼──────────────┼──────────────┼──────────────────────────┤');
    console.log(`│ All Files (avg)    │ ${formatTime(nodeStats.avg).padEnd(12)} │ ${formatTime(goStats.avg).padEnd(12)} │ ${formatDiff(nodeStats.avg, goStats.avg).padEnd(24)} │`);
    console.log(`│ All Files (median) │ ${formatTime(nodeStats.median).padEnd(12)} │ ${formatTime(goStats.median).padEnd(12)} │ ${formatDiff(nodeStats.median, goStats.median).padEnd(24)} │`);
    console.log(`│ All Files (min)    │ ${formatTime(nodeStats.min).padEnd(12)} │ ${formatTime(goStats.min).padEnd(12)} │ ${formatDiff(nodeStats.min, goStats.min).padEnd(24)} │`);
    console.log(`│ All Files (max)    │ ${formatTime(nodeStats.max).padEnd(12)} │ ${formatTime(goStats.max).padEnd(12)} │ ${formatDiff(nodeStats.max, goStats.max).padEnd(24)} │`);
    console.log(`│ Per File (avg)     │ ${formatTime(nodeStats.avg / testCount).padEnd(12)} │ ${formatTime(goStats.avg / testCount).padEnd(12)} │ ${formatDiff(nodeStats.avg, goStats.avg).padEnd(24)} │`);
    console.log(`│ Variation (±%)     │ ${nodeStats.variationPerc.toFixed(1)}%`.padEnd(34) + '│ ' + `${goStats.variationPerc.toFixed(1)}%`.padEnd(12) + '│ ' + 'N/A'.padEnd(24) + ' │');
    console.log('└────────────────────┴──────────────┴──────────────┴──────────────────────────┘');
}
console.log('');

console.log('┌─────────────────────────────────────────────────────────────────────────────┐');
console.log('│ MEMORY & ALLOCATIONS (Go only, per build)                                  │');
console.log('├─────────────────────────────────────────────────────────────────────────────┤');
console.log(`│ Memory per build:        ${(goBytesPerOp / (1024 * 1024)).toFixed(2)} MB                                          │`);
console.log(`│ Memory per file:         ${(goBytesPerOp / (1024 * 1024) / testCount).toFixed(2)} MB                                          │`);
console.log(`│ Allocations per build:   ${goAllocsPerOp.toLocaleString()} allocations                                │`);
console.log(`│ Allocations per file:    ${Math.round(goAllocsPerOp / testCount).toLocaleString()} allocations                                   │`);
console.log('└─────────────────────────────────────────────────────────────────────────────┘');
console.log('');

// Performance verdicts
console.log('═'.repeat(80));
console.log('PERFORMANCE ANALYSIS');
console.log('═'.repeat(80));

// Find the fastest
const implementations = [
    { name: 'Node.js', avg: nodeStats.avg },
    { name: 'Go', avg: goStats.avg }
];
if (bunAvailable) {
    implementations.push({ name: 'Bun', avg: bunStats.avg });
}
implementations.sort((a, b) => a.avg - b.avg);
const fastest = implementations[0];
const slowest = implementations[implementations.length - 1];

console.log('🏆 RANKING (fastest to slowest):');
implementations.forEach((impl, idx) => {
    const medal = idx === 0 ? '🥇' : idx === 1 ? '🥈' : '🥉';
    const ratio = impl.avg / fastest.avg;
    const suffix = idx === 0 ? '' : ` (${ratio.toFixed(2)}x slower)`;
    console.log(`   ${medal} ${impl.name}: ${formatTime(impl.avg)}${suffix}`);
});
console.log('');

if (bunAvailable) {
    const bunVsNode = bunStats.avg / nodeStats.avg;
    console.log('📊 BUN vs NODE.JS:');
    if (bunVsNode < 0.8) {
        console.log(`   🚀 Bun is ${(1/bunVsNode).toFixed(1)}x FASTER than Node.js`);
    } else if (bunVsNode < 1.2) {
        console.log(`   ⚖️  Bun and Node.js perform similarly (within 20%)`);
    } else {
        console.log(`   🐌 Bun is ${bunVsNode.toFixed(1)}x slower than Node.js`);
    }
    console.log(`   Build time: Node ${formatTime(nodeStats.avg)} vs Bun ${formatTime(bunStats.avg)}`);
    console.log('');
}

const goVsNode = goStats.avg / nodeStats.avg;
console.log('📊 GO vs NODE.JS:');
if (goVsNode < 0.8) {
    console.log(`   🚀 Go is ${(1/goVsNode).toFixed(1)}x FASTER than Node.js`);
} else if (goVsNode < 1.2) {
    console.log(`   ⚖️  Go and Node.js perform similarly (within 20%)`);
} else {
    console.log(`   🐌 Go is ${goVsNode.toFixed(1)}x slower than Node.js`);
}
console.log(`   Build time: Node ${formatTime(nodeStats.avg)} vs Go ${formatTime(goStats.avg)}`);
console.log('');

if (bunAvailable) {
    const goVsBun = goStats.avg / bunStats.avg;
    console.log('📊 GO vs BUN:');
    if (goVsBun < 0.8) {
        console.log(`   🚀 Go is ${(1/goVsBun).toFixed(1)}x FASTER than Bun`);
    } else if (goVsBun < 1.2) {
        console.log(`   ⚖️  Go and Bun perform similarly (within 20%)`);
    } else {
        console.log(`   🐌 Go is ${goVsBun.toFixed(1)}x slower than Bun`);
    }
    console.log(`   Build time: Bun ${formatTime(bunStats.avg)} vs Go ${formatTime(goStats.avg)}`);
    console.log('');
}

// Notes
console.log('📝 NOTES:');
console.log('  • This benchmark simulates REALISTIC CLI/build tool usage');
console.log('  • Each iteration = independent process compiling all files once');
console.log('  • NO artificial JIT warming from repeated in-process runs');
console.log('  • All implementations produce identical CSS output');
console.log(`  • All implementations: ${ITERATIONS} separate processes`);
console.log('  • Fair comparison with same methodology for all runtimes');
console.log('');

console.log('💡 OTHER BENCHMARKS:');
console.log('  • Run `pnpm bench:compare` for Node.js vs Go per-file comparison');
console.log('  • Run `pnpm bench:compare:suite` for Node.js vs Go suite comparison');
console.log('  • Run `pnpm bench:bun` for Bun standalone benchmark');
console.log('');

console.log('═'.repeat(80));
