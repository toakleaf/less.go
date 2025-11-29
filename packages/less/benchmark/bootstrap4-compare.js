/**
 * Bootstrap4 benchmark - compares less.js compilation time
 * Usage: node bootstrap4-compare.js [--runs=N]
 */

const path = require('path');
const fs = require('fs');
const less = require('../.');

// Configuration
const DEFAULT_RUNS = 10;
const WARMUP_RUNS = 3;

// High-resolution timing
function getTime() {
	const hrtime = process.hrtime();
	return hrtime[0] * 1000 + hrtime[1] / 1000000;
}

function formatTime(ms) {
	if (ms < 1) {
		return `${(ms * 1000).toFixed(2)}µs`;
	} else if (ms < 1000) {
		return `${ms.toFixed(2)}ms`;
	} else {
		return `${(ms / 1000).toFixed(2)}s`;
	}
}

async function runBenchmark() {
	const customRuns = parseInt(process.argv.find(arg => arg.startsWith('--runs='))?.split('=')[1]);
	const runCount = customRuns || DEFAULT_RUNS;
	const jsonMode = process.argv.includes('--json');

	// Locate the bootstrap4 test file
	const testDataRoot = path.join(__dirname, '../../test-data');
	const lessFile = path.join(testDataRoot, 'less/3rd-party/bootstrap4.less');
	const lessContent = fs.readFileSync(lessFile, 'utf8');

	const options = {
		filename: lessFile,
		math: 0, // strict math
		javascriptEnabled: true,
	};

	const times = [];
	let coldStartTime = null;

	if (!jsonMode) {
		console.log('Bootstrap4 Less.js Benchmark');
		console.log('='.repeat(50));
		console.log(`Runs: ${runCount} (${WARMUP_RUNS} warmup)`);
		console.log('');
	}

	for (let i = 0; i < runCount + WARMUP_RUNS; i++) {
		const startTime = getTime();

		try {
			await new Promise((resolve, reject) => {
				less.render(lessContent, options, (err, result) => {
					if (err) reject(err);
					else resolve(result);
				});
			});
		} catch (err) {
			console.error('Compilation error:', err.message);
			process.exit(1);
		}

		const endTime = getTime();
		const elapsed = endTime - startTime;

		if (i === 0) {
			coldStartTime = elapsed;
		}

		if (i >= WARMUP_RUNS) {
			times.push(elapsed);
		}

		if (!jsonMode) {
			const status = i < WARMUP_RUNS ? '(warmup)' : '';
			process.stdout.write(`\rRun ${i + 1}/${runCount + WARMUP_RUNS} ${status}   `);
		}
	}

	if (!jsonMode) {
		process.stdout.write('\n\n');
	}

	// Calculate stats
	const sum = times.reduce((a, b) => a + b, 0);
	const avg = sum / times.length;
	const min = Math.min(...times);
	const max = Math.max(...times);
	const sorted = [...times].sort((a, b) => a - b);
	const median = sorted[Math.floor(sorted.length / 2)];
	const variationPerc = ((max - min) / avg) * 100;

	if (jsonMode) {
		console.log(JSON.stringify({
			test: 'bootstrap4',
			engine: 'less.js',
			runs: runCount,
			warmupRuns: WARMUP_RUNS,
			coldStart: coldStartTime,
			avg: avg,
			min: min,
			max: max,
			median: median,
			variationPerc: variationPerc
		}));
	} else {
		console.log('Results (after warmup):');
		console.log('-'.repeat(50));
		console.log(`  Cold start:  ${formatTime(coldStartTime)}`);
		console.log(`  Average:     ${formatTime(avg)} ± ${variationPerc.toFixed(1)}%`);
		console.log(`  Median:      ${formatTime(median)}`);
		console.log(`  Min:         ${formatTime(min)}`);
		console.log(`  Max:         ${formatTime(max)}`);
		console.log('');
	}
}

runBenchmark().catch(err => {
	console.error('Benchmark failed:', err);
	process.exit(1);
});
