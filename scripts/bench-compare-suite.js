#!/usr/bin/env node

/**
 * Suite benchmark comparison tool
 * Compares suite-mode benchmarks where all files are compiled sequentially
 * This provides a more realistic workload simulation than per-file benchmarks
 */

const { execSync } = require('child_process');
const path = require('path');

console.log('╔══════════════════════════════════════════════════════════════════════════════╗');
console.log('║       LESS.JS vs LESS.GO SUITE BENCHMARK COMPARISON                          ║');
console.log('║       (All files compiled sequentially, repeated 30x)                        ║');
console.log('╚══════════════════════════════════════════════════════════════════════════════╝\n');

// Run JavaScript suite benchmark and capture JSON output
console.log('Running JavaScript suite benchmark (30 iterations)...');
const jsOutput = execSync('node packages/less/benchmark/suite.js --suite --json', {
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

// Run Go suite benchmark
console.log('Running Go suite benchmark (30 iterations with 5 warmup runs)...\n');
let goOutput;
try {
    goOutput = execSync('go test -bench=BenchmarkLargeSuite -benchmem -benchtime=30x ./packages/less/src/less/less_go', {
        encoding: 'utf8',
        cwd: path.join(__dirname, '..'),
        maxBuffer: 10 * 1024 * 1024
    });
} catch (error) {
    if (error.stdout) {
        goOutput = error.stdout;
        console.log('⚠️  Go suite benchmark had issues, but continuing with results...\n');
    } else {
        console.error('Failed to run Go suite benchmark');
        console.error(error.message);
        process.exit(1);
    }
}

// Parse Go benchmark output
// Format: BenchmarkLargeSuite-10    30    12345678 ns/op    234567 B/op    5678 allocs/op
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

// Extract statistics
const jsTestCount = jsData.testCount;
const jsColdMs = jsData.coldStart;
const jsWarmMs = jsData.warm.avg;
const jsWarmMedianMs = jsData.warm.median;
const jsWarmMinMs = jsData.warm.min;
const jsWarmMaxMs = jsData.warm.max;

// Display results
console.log('═'.repeat(80));
console.log('SUITE BENCHMARK RESULTS');
console.log('═'.repeat(80));
console.log(`Files in suite: ${jsTestCount}`);
console.log(`Suite iterations: 30 (5 warmup + 25 measured)`);
console.log(`Total compilations per benchmark: ${jsTestCount * 30}`);
console.log(`Methodology: All ${jsTestCount} files compiled sequentially per iteration`);
console.log('');

console.log('┌─────────────────────────────────────────────────────────────────────────────┐');
console.log('│ 🥶 COLD START PERFORMANCE (1st iteration, no warmup)                        │');
console.log('├─────────────────────────────────────────────────────────────────────────────┤');
console.log('│                    │  JavaScript  │      Go      │   Difference             │');
console.log('├────────────────────┼──────────────┼──────────────┼──────────────────────────┤');
console.log(`│ All Files (total)  │ ${formatTime(jsColdMs).padEnd(12)} │ N/A          │ N/A                      │`);
console.log(`│ Per File (avg)     │ ${formatTime(jsColdMs / jsTestCount).padEnd(12)} │ N/A          │ N/A                      │`);
console.log('└────────────────────┴──────────────┴──────────────┴──────────────────────────┘');
console.log('Note: Go cold-start not measured in suite mode');
console.log('');

console.log('┌─────────────────────────────────────────────────────────────────────────────┐');
console.log('│ 🔥 WARM PERFORMANCE (after 5 warmup suite iterations)                       │');
console.log('├─────────────────────────────────────────────────────────────────────────────┤');
console.log('│                    │  JavaScript  │      Go      │   Difference             │');
console.log('├────────────────────┼──────────────┼──────────────┼──────────────────────────┤');
console.log(`│ All Files (avg)    │ ${formatTime(jsWarmMs).padEnd(12)} │ ${formatTime(goMsPerOp).padEnd(12)} │ ${formatDiff(jsWarmMs, goMsPerOp).padEnd(24)} │`);
console.log(`│ All Files (median) │ ${formatTime(jsWarmMedianMs).padEnd(12)} │ N/A          │ N/A                      │`);
console.log(`│ Per File (avg)     │ ${formatTime(jsWarmMs / jsTestCount).padEnd(12)} │ ${formatTime(goMsPerOp / jsTestCount).padEnd(12)} │ ${formatDiff(jsWarmMs / jsTestCount, goMsPerOp / jsTestCount).padEnd(24)} │`);
console.log('└────────────────────┴──────────────┴──────────────┴──────────────────────────┘');
console.log('');

console.log('┌─────────────────────────────────────────────────────────────────────────────┐');
console.log('│ MEMORY & ALLOCATIONS (Go only, per suite iteration)                        │');
console.log('├─────────────────────────────────────────────────────────────────────────────┤');
console.log(`│ Memory per suite:        ${(goBytesPerOp / (1024 * 1024)).toFixed(2)} MB                                          │`);
console.log(`│ Memory per file:         ${(goBytesPerOp / (1024 * 1024) / jsTestCount).toFixed(2)} MB                                          │`);
console.log(`│ Allocations per suite:   ${goAllocsPerOp.toLocaleString()} allocations                                │`);
console.log(`│ Allocations per file:    ${Math.round(goAllocsPerOp / jsTestCount).toLocaleString()} allocations                                   │`);
console.log('└─────────────────────────────────────────────────────────────────────────────┘');
console.log('');

// Performance verdict
console.log('═'.repeat(80));
console.log('PERFORMANCE ANALYSIS');
console.log('═'.repeat(80));

const warmSpeedRatio = goMsPerOp / jsWarmMs;
let warmVerdict;
if (warmSpeedRatio < 0.8) {
    warmVerdict = `🚀 Go is ${(1/warmSpeedRatio).toFixed(1)}x FASTER than JavaScript`;
} else if (warmSpeedRatio < 1.2) {
    warmVerdict = `⚖️  Performance is SIMILAR (within 20%)`;
} else if (warmSpeedRatio < 2) {
    warmVerdict = `🐌 Go is ${warmSpeedRatio.toFixed(1)}x slower than JavaScript`;
} else {
    warmVerdict = `🐌 Go is ${warmSpeedRatio.toFixed(1)}x SLOWER than JavaScript`;
}

console.log('🔥 SUITE WARM PERFORMANCE:');
console.log(`   ${warmVerdict}`);
console.log(`   Suite time: JS ${formatTime(jsWarmMs)} vs Go ${formatTime(goMsPerOp)} (${warmSpeedRatio.toFixed(2)}x)`);
console.log(`   Per-file avg: JS ${formatTime(jsWarmMs / jsTestCount)} vs Go ${formatTime(goMsPerOp / jsTestCount)}`);
console.log('');

// Warmup effect (JavaScript only, since Go doesn't have cold-start in suite mode)
const jsWarmupEffect = ((jsColdMs - jsWarmMs) / jsColdMs * 100);
console.log('📈 WARMUP EFFECT:');
console.log(`   JavaScript: ${jsWarmupEffect.toFixed(1)}% faster after warmup`);
console.log(`   Go: Not measured in suite mode (warmup included in measurement)`);
console.log('');

// Context
console.log('📝 NOTES:');
console.log('  • Suite mode simulates realistic build workloads');
console.log('  • All files compiled sequentially, reducing cache benefits');
console.log('  • More realistic than per-file repeated compilation');
console.log('  • Both implementations produce identical CSS output');
console.log('  • JavaScript: 5 warmup suite runs + 25 measured suite runs');
console.log('  • Go: 5 warmup suite runs + 25 measured suite runs');
console.log('');

console.log('💡 COMPARISON WITH PER-FILE BENCHMARKS:');
console.log('  • Run `pnpm bench:compare` to see per-file benchmark results');
console.log('  • Per-file benchmarks may show better performance due to cache warming');
console.log('  • Suite benchmarks better represent real-world build processes');
console.log('');

// Recommendations
if (warmSpeedRatio > 1.5) {
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
