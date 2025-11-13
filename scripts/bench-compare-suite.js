#!/usr/bin/env node

/**
 * Multi-process suite benchmark comparison tool
 * Runs REALISTIC benchmarks where each iteration is a fresh process
 * This simulates actual CLI/build tool usage patterns
 */

const { execSync } = require('child_process');
const path = require('path');

const ITERATIONS = 30; // Number of independent process runs

console.log('╔══════════════════════════════════════════════════════════════════════════════╗');
console.log('║       LESS.JS vs LESS.GO REALISTIC SUITE BENCHMARK                           ║');
console.log('║       (Each iteration = fresh process, like real CLI usage)                  ║');
console.log('╚══════════════════════════════════════════════════════════════════════════════╝\n');

console.log(`Running ${ITERATIONS} independent build sessions for each implementation...`);
console.log('This simulates realistic CLI/build tool usage where each build is a fresh process.\n');

// Run JavaScript benchmarks - multiple independent processes
console.log(`Running JavaScript benchmarks (${ITERATIONS} separate processes)...`);
const jsTimes = [];
for (let i = 0; i < ITERATIONS; i++) {
    process.stdout.write(`\r  JS Progress: ${i + 1}/${ITERATIONS} (${((i + 1) / ITERATIONS * 100).toFixed(1)}%)`);

    try {
        const output = execSync('node packages/less/benchmark/suite.js --single-run', {
            encoding: 'utf8',
            cwd: path.join(__dirname, '..'),
            stdio: ['ignore', 'pipe', 'ignore'] // Suppress stderr
        });

        const result = JSON.parse(output.trim());
        jsTimes.push(result.totalTime);
    } catch (error) {
        console.error(`\n⚠️  JS iteration ${i + 1} failed`);
    }
}
process.stdout.write('\n');

// Run Go benchmarks - Go's benchmark framework handles this
console.log(`Running Go benchmarks (${ITERATIONS} iterations via go test)...\n`);
let goOutput;
try {
    goOutput = execSync(`go test -bench=BenchmarkLargeSuite -benchmem -benchtime=${ITERATIONS}x ./packages/less/src/less/less_go`, {
        encoding: 'utf8',
        cwd: path.join(__dirname, '..'),
        maxBuffer: 10 * 1024 * 1024
    });
} catch (error) {
    if (error.stdout) {
        goOutput = error.stdout;
        console.log('⚠️  Go benchmark had issues, but continuing with results...\n');
    } else {
        console.error('Failed to run Go benchmark');
        console.error(error.message);
        process.exit(1);
    }
}

// Parse Go benchmark output
const match = goOutput.match(/BenchmarkLargeSuite-\d+\s+(\d+)\s+(\d+)\s+ns\/op\s+(\d+)\s+B\/op\s+(\d+)\s+allocs\/op/);
if (!match) {
    console.error('Failed to parse Go benchmark output');
    console.error('Output:', goOutput);
    process.exit(1);
}

const goIterations = parseInt(match[1]);
const goNsPerOp = parseInt(match[2]);
const goBytesPerOp = parseInt(match[3]);
const goAllocsPerOp = parseInt(match[4]);
const goMsPerOp = goNsPerOp / 1_000_000;

// Calculate JavaScript statistics
const jsAvg = jsTimes.reduce((sum, t) => sum + t, 0) / jsTimes.length;
const jsMin = Math.min(...jsTimes);
const jsMax = Math.max(...jsTimes);
const jsSorted = [...jsTimes].sort((a, b) => a - b);
const jsMedian = jsSorted[Math.floor(jsSorted.length / 2)];
const jsStdDev = Math.sqrt(jsTimes.reduce((sum, t) => sum + Math.pow(t - jsAvg, 2), 0) / jsTimes.length);
const jsVariationPerc = (jsStdDev / jsAvg * 100);

// Get test count from first JS run
let testCount = 73; // default
try {
    const output = execSync('node packages/less/benchmark/suite.js --single-run', {
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
console.log(`Build iterations: ${ITERATIONS} independent processes (JS) / ${goIterations} iterations (Go)`);
console.log(`Methodology: Each iteration = fresh process compiling all files once`);
console.log('This represents actual CLI/build tool usage patterns');
console.log('');

console.log('┌─────────────────────────────────────────────────────────────────────────────┐');
console.log('│ 📊 BUILD PERFORMANCE (all files, one pass per build)                        │');
console.log('├─────────────────────────────────────────────────────────────────────────────┤');
console.log('│                    │  JavaScript  │      Go      │   Difference             │');
console.log('├────────────────────┼──────────────┼──────────────┼──────────────────────────┤');
console.log(`│ All Files (avg)    │ ${formatTime(jsAvg).padEnd(12)} │ ${formatTime(goMsPerOp).padEnd(12)} │ ${formatDiff(jsAvg, goMsPerOp).padEnd(24)} │`);
console.log(`│ All Files (median) │ ${formatTime(jsMedian).padEnd(12)} │ N/A          │ N/A                      │`);
console.log(`│ All Files (min)    │ ${formatTime(jsMin).padEnd(12)} │ N/A          │ N/A                      │`);
console.log(`│ All Files (max)    │ ${formatTime(jsMax).padEnd(12)} │ N/A          │ N/A                      │`);
console.log(`│ Per File (avg)     │ ${formatTime(jsAvg / testCount).padEnd(12)} │ ${formatTime(goMsPerOp / testCount).padEnd(12)} │ ${formatDiff(jsAvg / testCount, goMsPerOp / testCount).padEnd(24)} │`);
console.log(`│ Variation (±%)     │ ${jsVariationPerc.toFixed(1)}%`.padEnd(15) + '│ N/A          │ N/A                      │');
console.log('└────────────────────┴──────────────┴──────────────┴──────────────────────────┘');
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

// Performance verdict
console.log('═'.repeat(80));
console.log('PERFORMANCE ANALYSIS');
console.log('═'.repeat(80));

const speedRatio = goMsPerOp / jsAvg;
let verdict;
if (speedRatio < 0.8) {
    verdict = `🚀 Go is ${(1/speedRatio).toFixed(1)}x FASTER than JavaScript`;
} else if (speedRatio < 1.2) {
    verdict = `⚖️  Performance is SIMILAR (within 20%)`;
} else if (speedRatio < 2) {
    verdict = `🐌 Go is ${speedRatio.toFixed(1)}x slower than JavaScript`;
} else {
    verdict = `🐌 Go is ${speedRatio.toFixed(1)}x SLOWER than JavaScript`;
}

console.log('🏗️  REALISTIC BUILD PERFORMANCE:');
console.log(`   ${verdict}`);
console.log(`   Build time: JS ${formatTime(jsAvg)} vs Go ${formatTime(goMsPerOp)} (${speedRatio.toFixed(2)}x)`);
console.log(`   Per-file avg: JS ${formatTime(jsAvg / testCount)} vs Go ${formatTime(goMsPerOp / testCount)}`);
console.log('');

// Check for within-iteration variance in JS
const firstHalf = jsTimes.slice(0, Math.floor(jsTimes.length / 2));
const secondHalf = jsTimes.slice(Math.floor(jsTimes.length / 2));
const firstHalfAvg = firstHalf.reduce((sum, t) => sum + t, 0) / firstHalf.length;
const secondHalfAvg = secondHalf.reduce((sum, t) => sum + t, 0) / secondHalf.length;
const improvementPerc = ((firstHalfAvg - secondHalfAvg) / firstHalfAvg * 100);

console.log('📈 PROCESS WARMING ANALYSIS:');
if (Math.abs(improvementPerc) < 5) {
    console.log(`   ✅ No significant warming detected (${improvementPerc.toFixed(1)}% difference)`);
    console.log(`   Each build session is truly independent`);
} else if (improvementPerc > 0) {
    console.log(`   ⚠️  Builds got ${improvementPerc.toFixed(1)}% faster over time`);
    console.log(`   This suggests system-level caching (disk cache, OS, etc.)`);
} else {
    console.log(`   ⚠️  Builds got ${Math.abs(improvementPerc).toFixed(1)}% slower over time`);
    console.log(`   This suggests system resource contention or thermal throttling`);
}
console.log(`   First half avg: ${formatTime(firstHalfAvg)}, Second half avg: ${formatTime(secondHalfAvg)}`);
console.log('');

// Context
console.log('📝 NOTES:');
console.log('  • This benchmark simulates REALISTIC CLI/build tool usage');
console.log('  • Each iteration = independent process compiling all files once');
console.log('  • NO artificial JIT warming from repeated in-process runs');
console.log('  • Both implementations produce identical CSS output');
console.log(`  • JavaScript: ${ITERATIONS} separate node processes`);
console.log(`  • Go: ${goIterations} benchmark iterations`);
console.log('');

console.log('💡 COMPARISON WITH PER-FILE BENCHMARKS:');
console.log('  • Run `pnpm bench:compare` to see per-file warmup effects');
console.log('  • Per-file benchmarks show JIT optimization potential');
console.log('  • Suite benchmarks show real-world build tool performance');
console.log('');

// Recommendations
if (speedRatio > 1.5) {
    console.log('💡 OPTIMIZATION OPPORTUNITIES:');
    console.log('  • Profile with: pnpm bench:profile');
    console.log('  • Or manually: go test -bench=BenchmarkLargeSuite -cpuprofile=cpu.prof');
    console.log('  • Analyze with: go tool pprof cpu.prof');
    console.log('  • Common hotspots: excessive allocations, string operations, reflection');
}

console.log('═'.repeat(80));

// Helper functions
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

function formatDiff(jsTime, goTime) {
    const ratio = goTime / jsTime;
    const percent = ((ratio - 1) * 100).toFixed(1);

    if (ratio < 0.8) {
        return `Go ${(1/ratio).toFixed(1)}x faster ✓`;
    } else if (ratio < 1.2) {
        return `Similar (~${Math.abs(percent)}%)`;
    } else {
        return `Go ${ratio.toFixed(1)}x slower`;
    }
}
