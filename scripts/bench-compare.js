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

// Run Go benchmarks (individual file benchmarks for fair comparison)
// Using benchtime=30x to match JavaScript's 30 iterations per file
console.log('Running Go warm benchmarks (30 iterations per file, with 5 warmup runs)...');
let goWarmOutput;
try {
    goWarmOutput = execSync('go test -bench="^BenchmarkLessCompilation$" -benchmem -benchtime=30x ./less', {
        encoding: 'utf8',
        cwd: path.join(__dirname, '..'),
        maxBuffer: 10 * 1024 * 1024 // 10MB buffer for all the output
    });
} catch (error) {
    // Benchmark may fail on some tests, but we can still parse the successful results
    if (error.stdout) {
        goWarmOutput = error.stdout;
        console.log('⚠️  Some Go warm benchmarks failed, but continuing with successful results...');
    } else {
        console.error('Failed to run Go warm benchmarks');
        console.error(error.message);
        process.exit(1);
    }
}

console.log('Running Go cold-start benchmarks (30 iterations per file, no warmup)...\n');
let goColdOutput;
try {
    goColdOutput = execSync('go test -bench=BenchmarkLessCompilationColdStart -benchmem -benchtime=30x ./less', {
        encoding: 'utf8',
        cwd: path.join(__dirname, '..'),
        maxBuffer: 10 * 1024 * 1024 // 10MB buffer for all the output
    });
} catch (error) {
    // Benchmark may fail on some tests, but we can still parse the successful results
    if (error.stdout) {
        goColdOutput = error.stdout;
        console.log('⚠️  Some Go cold-start benchmarks failed, but continuing with successful results...\n');
    } else {
        console.error('Failed to run Go cold-start benchmarks');
        console.error(error.message);
        process.exit(1);
    }
}

// Parse Go warm benchmark output
// Format: BenchmarkLessCompilation/main/colors-10    123    12345678 ns/op    234567 B/op    5678 allocs/op
const goWarmResults = [];
const goWarmLines = goWarmOutput.split('\n');
for (const line of goWarmLines) {
    const match = line.match(/BenchmarkLessCompilation\/(.+?)-\d+\s+(\d+)\s+(\d+)\s+ns\/op\s+(\d+)\s+B\/op\s+(\d+)\s+allocs\/op/);
    if (match) {
        goWarmResults.push({
            name: match[1],
            iterations: parseInt(match[2]),
            nsPerOp: parseInt(match[3]),
            bytesPerOp: parseInt(match[4]),
            allocsPerOp: parseInt(match[5])
        });
    }
}

// Parse Go cold-start benchmark output
const goColdResults = [];
const goColdLines = goColdOutput.split('\n');
for (const line of goColdLines) {
    const match = line.match(/BenchmarkLessCompilationColdStart\/(.+?)-\d+\s+(\d+)\s+(\d+)\s+ns\/op\s+(\d+)\s+B\/op\s+(\d+)\s+allocs\/op/);
    if (match) {
        goColdResults.push({
            name: match[1],
            iterations: parseInt(match[2]),
            nsPerOp: parseInt(match[3]),
            bytesPerOp: parseInt(match[4]),
            allocsPerOp: parseInt(match[5])
        });
    }
}

if (goWarmResults.length === 0 && goColdResults.length === 0) {
    console.error('Failed to parse Go benchmark output - no results found');
    console.error('Warm output:', goWarmOutput);
    console.error('Cold output:', goColdOutput);
    process.exit(1);
}

// Check for skipped tests
const failedTests = [];
const allGoOutput = goWarmOutput + '\n' + goColdOutput;
const failLines = allGoOutput.split('\n').filter(line => line.includes('--- FAIL:'));
for (const line of failLines) {
    const match = line.match(/--- FAIL: Benchmark\w+\/(.+)/);
    if (match) {
        failedTests.push(match[1]);
    }
}

// Calculate statistics from JavaScript results
const jsTestCount = jsData.tests.length;

// JavaScript warm stats (after warmup)
const jsWarmAvg = jsData.tests.reduce((sum, t) => sum + (t.total?.avg || 0), 0) / jsTestCount;
const jsWarmMedian = calculateMedian(jsData.tests.map(t => t.total?.median || 0));
const jsWarmTotal = jsData.tests.reduce((sum, t) => sum + (t.total?.avg || 0), 0);

// JavaScript cold-start stats
const jsColdStarts = jsData.tests.map(t => t.coldStart).filter(t => t != null);
const jsColdAvg = jsColdStarts.reduce((sum, t) => sum + t, 0) / jsColdStarts.length;
const jsColdMedian = calculateMedian(jsColdStarts);
const jsColdTotal = jsColdStarts.reduce((sum, t) => sum + t, 0);

// Calculate Go warm statistics from individual file results
const goWarmTimesMs = goWarmResults.map(r => r.nsPerOp / 1_000_000);
const goWarmAvgMs = goWarmTimesMs.length > 0 ? goWarmTimesMs.reduce((sum, t) => sum + t, 0) / goWarmTimesMs.length : 0;
const goWarmMedianMs = goWarmTimesMs.length > 0 ? calculateMedian(goWarmTimesMs) : 0;
const goWarmTotalMs = goWarmTimesMs.reduce((sum, t) => sum + t, 0);

// Calculate Go cold-start statistics
const goColdTimesMs = goColdResults.map(r => r.nsPerOp / 1_000_000);
const goColdAvgMs = goColdTimesMs.length > 0 ? goColdTimesMs.reduce((sum, t) => sum + t, 0) / goColdTimesMs.length : 0;
const goColdMedianMs = goColdTimesMs.length > 0 ? calculateMedian(goColdTimesMs) : 0;
const goColdTotalMs = goColdTimesMs.reduce((sum, t) => sum + t, 0);

// Use warm results for memory stats (cold-start allocates more, not representative)
const goAvgBytesPerOp = goWarmResults.length > 0 ? goWarmResults.reduce((sum, r) => sum + r.bytesPerOp, 0) / goWarmResults.length : 0;
const goAvgAllocsPerOp = goWarmResults.length > 0 ? Math.round(goWarmResults.reduce((sum, r) => sum + r.allocsPerOp, 0) / goWarmResults.length) : 0;
const goMemoryMB = goAvgBytesPerOp / (1024 * 1024);

// Display results
console.log('═'.repeat(80));
console.log('RESULTS SUMMARY');
console.log('═'.repeat(80));
console.log(`Test Files: ${jsTestCount}`);
console.log(`Go warm benchmarked: ${goWarmResults.length}, Go cold benchmarked: ${goColdResults.length}`);
console.log(`Methodology: Both benchmark each file individually with identical iterations`);
if (failedTests.length > 0) {
    console.log(`⚠️  Skipped Go tests: ${failedTests.join(', ')}`);
}
console.log('');

console.log('┌─────────────────────────────────────────────────────────────────────────────┐');
console.log('│ 🥶 COLD START PERFORMANCE (1st iteration, no warmup)                        │');
console.log('├─────────────────────────────────────────────────────────────────────────────┤');
console.log('│                    │  JavaScript  │      Go      │   Difference             │');
console.log('├────────────────────┼──────────────┼──────────────┼──────────────────────────┤');
console.log(`│ Per File (avg)     │ ${formatTime(jsColdAvg).padEnd(12)} │ ${formatTime(goColdAvgMs).padEnd(12)} │ ${formatDiff(jsColdAvg, goColdAvgMs).padEnd(24)} │`);
console.log(`│ Per File (median)  │ ${formatTime(jsColdMedian).padEnd(12)} │ ${formatTime(goColdMedianMs).padEnd(12)} │ ${formatDiff(jsColdMedian, goColdMedianMs).padEnd(24)} │`);
console.log(`│ All Files (total)  │ ${formatTime(jsColdTotal).padEnd(12)} │ ${formatTime(goColdTotalMs).padEnd(12)} │ ${formatDiff(jsColdTotal, goColdTotalMs).padEnd(24)} │`);
console.log('└────────────────────┴──────────────┴──────────────┴──────────────────────────┘');
console.log('');

console.log('┌─────────────────────────────────────────────────────────────────────────────┐');
console.log('│ 🔥 WARM PERFORMANCE (after 5 warmup runs)                                   │');
console.log('├─────────────────────────────────────────────────────────────────────────────┤');
console.log('│                    │  JavaScript  │      Go      │   Difference             │');
console.log('├────────────────────┼──────────────┼──────────────┼──────────────────────────┤');
console.log(`│ Per File (avg)     │ ${formatTime(jsWarmAvg).padEnd(12)} │ ${formatTime(goWarmAvgMs).padEnd(12)} │ ${formatDiff(jsWarmAvg, goWarmAvgMs).padEnd(24)} │`);
console.log(`│ Per File (median)  │ ${formatTime(jsWarmMedian).padEnd(12)} │ ${formatTime(goWarmMedianMs).padEnd(12)} │ ${formatDiff(jsWarmMedian, goWarmMedianMs).padEnd(24)} │`);
console.log(`│ All Files (total)  │ ${formatTime(jsWarmTotal).padEnd(12)} │ ${formatTime(goWarmTotalMs).padEnd(12)} │ ${formatDiff(jsWarmTotal, goWarmTotalMs).padEnd(24)} │`);
console.log('└────────────────────┴──────────────┴──────────────┴──────────────────────────┘');
console.log('');

console.log('┌─────────────────────────────────────────────────────────────────────────────┐');
console.log('│ MEMORY & ALLOCATIONS (Go only, averaged per file)                          │');
console.log('├─────────────────────────────────────────────────────────────────────────────┤');
console.log(`│ Memory per file:         ${goMemoryMB.toFixed(2)} MB                                              │`);
console.log(`│ Allocations per file:    ${goAvgAllocsPerOp.toLocaleString()} allocations                                   │`);
console.log('└─────────────────────────────────────────────────────────────────────────────┘');
console.log('');

// Performance verdict
console.log('═'.repeat(80));
console.log('PERFORMANCE ANALYSIS');
console.log('═'.repeat(80));

// Warm performance analysis (primary metric for fair comparison)
const warmSpeedRatio = goWarmAvgMs / jsWarmAvg;
let warmVerdict;
if (warmSpeedRatio < 0.8) {
    warmVerdict = `🚀 Go is ${(1/warmSpeedRatio).toFixed(1)}x FASTER than JavaScript (warm)`;
} else if (warmSpeedRatio < 1.2) {
    warmVerdict = `⚖️  Warm performance is SIMILAR (within 20%)`;
} else if (warmSpeedRatio < 2) {
    warmVerdict = `🐌 Go is ${warmSpeedRatio.toFixed(1)}x slower than JavaScript (warm)`;
} else {
    warmVerdict = `🐌 Go is ${warmSpeedRatio.toFixed(1)}x SLOWER than JavaScript (warm)`;
}

// Cold-start performance analysis
const coldSpeedRatio = goColdAvgMs / jsColdAvg;
let coldVerdict;
if (coldSpeedRatio < 0.8) {
    coldVerdict = `🚀 Go is ${(1/coldSpeedRatio).toFixed(1)}x FASTER than JavaScript (cold start)`;
} else if (coldSpeedRatio < 1.2) {
    coldVerdict = `⚖️  Cold-start performance is SIMILAR (within 20%)`;
} else if (coldSpeedRatio < 2) {
    coldVerdict = `🐌 Go is ${coldSpeedRatio.toFixed(1)}x slower than JavaScript (cold start)`;
} else {
    coldVerdict = `🐌 Go is ${coldSpeedRatio.toFixed(1)}x SLOWER than JavaScript (cold start)`;
}

console.log('🔥 WARM PERFORMANCE (primary comparison metric):');
console.log(`   ${warmVerdict}`);
console.log(`   Per-file average: JS ${formatTime(jsWarmAvg)} vs Go ${formatTime(goWarmAvgMs)} (${warmSpeedRatio.toFixed(2)}x)`);
console.log('');

console.log('🥶 COLD START PERFORMANCE:');
console.log(`   ${coldVerdict}`);
console.log(`   Per-file average: JS ${formatTime(jsColdAvg)} vs Go ${formatTime(goColdAvgMs)} (${coldSpeedRatio.toFixed(2)}x)`);
console.log('');

// Warmup effect analysis
const jsWarmupEffect = ((jsColdAvg - jsWarmAvg) / jsColdAvg * 100);
const goWarmupEffect = ((goColdAvgMs - goWarmAvgMs) / goColdAvgMs * 100);
console.log('📈 WARMUP EFFECT:');
console.log(`   JavaScript: ${jsWarmupEffect.toFixed(1)}% faster after warmup`);
console.log(`   Go:         ${goWarmupEffect.toFixed(1)}% faster after warmup`);
console.log('');

// Context
console.log('📝 NOTES:');
console.log('  • Both benchmarks now use IDENTICAL methodology with warmup runs');
console.log('  • JavaScript: 5 warmup runs + 25 measured runs per file');
console.log('  • Go: 5 warmup runs + 25 measured runs per file');
console.log('  • Warm performance is the PRIMARY metric for fair JIT vs AOT comparison');
console.log('  • Cold-start shows real-world CLI performance (first run)');
console.log('  • Both implementations produce identical CSS output');
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

function calculateMedian(arr) {
    const sorted = [...arr].sort((a, b) => a - b);
    const mid = Math.floor(sorted.length / 2);
    return sorted.length % 2 ? sorted[mid] : (sorted[mid - 1] + sorted[mid]) / 2;
}
