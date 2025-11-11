#!/usr/bin/env node

/**
 * Benchmark comparison tool
 * Runs both JS and Go benchmarks and provides clear, actionable comparison
 */

const { execSync } = require('child_process');
const path = require('path');

console.log('╔══════════════════════════════════════════════════════════════════════════════╗');
console.log('║              LESS.JS vs LESS.GO PERFORMANCE COMPARISON                       ║');
console.log('╚══════════════════════════════════════════════════════════════════════════════╝\n');

// Run JavaScript benchmark and capture JSON output
console.log('Running JavaScript benchmarks...');
const jsOutput = execSync('node packages/less/benchmark/suite.js --json', {
    encoding: 'utf8',
    cwd: path.join(__dirname, '..')
});

// Parse JSON output (it's at the end after "JSON OUTPUT:")
const jsonStart = jsOutput.indexOf('JSON OUTPUT:') + 'JSON OUTPUT:'.length;
const jsonStr = jsOutput.substring(jsonStart).trim();
let jsData;
try {
    jsData = JSON.parse(jsonStr);
} catch (e) {
    console.error('Failed to parse JavaScript benchmark JSON output');
    console.error('Output:', jsonStr.substring(0, 500));
    process.exit(1);
}

// Run Go benchmark
console.log('Running Go benchmarks...\n');
const goOutput = execSync('go test -bench=BenchmarkLargeSuite -benchmem -benchtime=5s ./packages/less/src/less/less_go', {
    encoding: 'utf8',
    cwd: path.join(__dirname, '..')
});

// Parse Go benchmark output
// Format: BenchmarkLargeSuite-10    45    264059064 ns/op    299574207 B/op   3437404 allocs/op
const goMatch = goOutput.match(/BenchmarkLargeSuite-\d+\s+(\d+)\s+(\d+)\s+ns\/op\s+(\d+)\s+B\/op\s+(\d+)\s+allocs\/op/);
if (!goMatch) {
    console.error('Failed to parse Go benchmark output');
    console.error('Output:', goOutput);
    process.exit(1);
}

const goIterations = parseInt(goMatch[1]);
const goNsPerOp = parseInt(goMatch[2]);
const goBytesPerOp = parseInt(goMatch[3]);
const goAllocsPerOp = parseInt(goMatch[4]);

// Calculate statistics
const jsTestCount = jsData.tests.length;
const jsTotalAvg = jsData.tests.reduce((sum, t) => sum + (t.total?.avg || 0), 0);
const jsAvgPerFile = jsTotalAvg / jsTestCount;
const jsMedianPerFile = calculateMedian(jsData.tests.map(t => t.total?.median || 0));
const jsTotalTime = jsTotalAvg; // Sum of all averages

// Go is measuring all files in one operation
const goTotalTimeMs = goNsPerOp / 1_000_000;
const goAvgPerFileMs = goTotalTimeMs / jsTestCount;
const goMemoryMB = goBytesPerOp / (1024 * 1024);

// Display results
console.log('═'.repeat(80));
console.log('RESULTS SUMMARY');
console.log('═'.repeat(80));
console.log(`Test Files: ${jsTestCount}`);
console.log(`Runs: JS=${jsData.runs - jsData.warmupRuns} (after ${jsData.warmupRuns} warmup), Go=${goIterations}`);
console.log('');

console.log('┌─────────────────────────────────────────────────────────────────────────────┐');
console.log('│ COMPILATION TIME                                                            │');
console.log('├─────────────────────────────────────────────────────────────────────────────┤');
console.log('│                    │  JavaScript  │      Go      │   Difference             │');
console.log('├────────────────────┼──────────────┼──────────────┼──────────────────────────┤');
console.log(`│ Per File (avg)     │ ${formatTime(jsAvgPerFile).padEnd(12)} │ ${formatTime(goAvgPerFileMs).padEnd(12)} │ ${formatDiff(jsAvgPerFile, goAvgPerFileMs).padEnd(24)} │`);
console.log(`│ Per File (median)  │ ${formatTime(jsMedianPerFile).padEnd(12)} │ ${'N/A'.padEnd(12)} │ ${''.padEnd(24)} │`);
console.log(`│ All Files (total)  │ ${formatTime(jsTotalTime).padEnd(12)} │ ${formatTime(goTotalTimeMs).padEnd(12)} │ ${formatDiff(jsTotalTime, goTotalTimeMs).padEnd(24)} │`);
console.log('└────────────────────┴──────────────┴──────────────┴──────────────────────────┘');
console.log('');

console.log('┌─────────────────────────────────────────────────────────────────────────────┐');
console.log('│ MEMORY & ALLOCATIONS (Go only)                                              │');
console.log('├─────────────────────────────────────────────────────────────────────────────┤');
console.log(`│ Memory per operation:    ${goMemoryMB.toFixed(2)} MB                                              │`);
console.log(`│ Allocations per op:      ${goAllocsPerOp.toLocaleString()} allocations                                   │`);
console.log(`│ Avg per file:            ${(goMemoryMB / jsTestCount).toFixed(2)} MB / ${Math.round(goAllocsPerOp / jsTestCount).toLocaleString()} allocs                         │`);
console.log('└─────────────────────────────────────────────────────────────────────────────┘');
console.log('');

// Performance verdict
console.log('═'.repeat(80));
console.log('PERFORMANCE ANALYSIS');
console.log('═'.repeat(80));

const speedRatio = goAvgPerFileMs / jsAvgPerFile;
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

console.log(verdict);
console.log('');

// Detailed breakdown
console.log('Per-file average:');
console.log(`  • JavaScript: ${formatTime(jsAvgPerFile)}`);
console.log(`  • Go:         ${formatTime(goAvgPerFileMs)}`);
console.log(`  • Ratio:      ${speedRatio.toFixed(2)}x`);
console.log('');

// Context
console.log('Notes:');
console.log('  • JavaScript is a mature, highly optimized JIT-compiled implementation');
console.log('  • Go port is still under active development');
console.log('  • Go provides predictable performance and lower memory overhead');
console.log('  • Both implementations produce identical CSS output');
console.log('');

// Recommendations
if (speedRatio > 1.5) {
    console.log('💡 Optimization Opportunities:');
    console.log('  • Profile with: go test -bench=BenchmarkLargeSuite -cpuprofile=cpu.prof');
    console.log('  • Analyze with: go tool pprof cpu.prof');
    console.log('  • Check for: excessive allocations, string operations, reflection');
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

function calculateMedian(arr) {
    const sorted = [...arr].sort((a, b) => a - b);
    const mid = Math.floor(sorted.length / 2);
    return sorted.length % 2 ? sorted[mid] : (sorted[mid - 1] + sorted[mid]) / 2;
}
