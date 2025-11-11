/**
 * Comprehensive benchmark suite for less.js
 * Tests the same files as the Go benchmark for fair comparison
 */

const path = require('path');
const fs = require('fs');
const less = require('../.');

// Configuration
const TOTAL_RUNS = 30;
const WARMUP_RUNS = 5;

// Test suites matching the Go benchmarks
const benchmarkTestFiles = [
	{
		suite: 'main',
		folder: '_main/',
		options: {
			relativeUrls: true,
			silent: true,
			javascriptEnabled: true
		},
		files: [
			'calc',
			'charsets',
			'colors',
			'colors2',
			'comments',
			'css-escapes',
			'css-grid',
			'css-guards',
			'empty',
			'extend-chaining',
			'extend-clearfix',
			'extend-exact',
			'extend-media',
			'extend-nest',
			'extend-selector',
			'extend',
			'extract-and-length',
			'functions-each',
			'ie-filters',
			'import-inline',
			'import-interpolation',
			'import-once',
			'lazy-eval',
			'merge',
			'mixin-noparens',
			'mixins-closure',
			'mixins-guards-default-func',
			'mixins-guards',
			'mixins-important',
			'mixins-interpolated',
			'mixins-named-args',
			'mixins-nested',
			'mixins-pattern',
			'mixins',
			'no-output',
			'operations',
			'parse-interpolation',
			'permissive-parse',
			'property-accessors',
			'property-name-interp',
			'rulesets',
			'scope',
			'selectors',
			'strings',
			'variables-in-at-rules',
			'variables',
			'whitespace'
		]
	},
	{
		suite: 'namespacing',
		folder: 'namespacing/',
		options: {},
		files: [
			'namespacing-1',
			'namespacing-2',
			'namespacing-3',
			'namespacing-4',
			'namespacing-5',
			'namespacing-6',
			'namespacing-7',
			'namespacing-8',
			'namespacing-functions',
			'namespacing-media',
			'namespacing-operations'
		]
	},
	{
		suite: 'math-parens',
		folder: 'math/strict/',
		options: {
			math: 'parens'
		},
		files: [
			'css',
			'media-math',
			'mixins-args',
			'parens'
		]
	},
	{
		suite: 'math-parens-division',
		folder: 'math/parens-division/',
		options: {
			math: 'parens-division'
		},
		files: [
			'media-math',
			'mixins-args',
			'new-division',
			'parens'
		]
	},
	{
		suite: 'math-always',
		folder: 'math/always/',
		options: {
			math: 'always'
		},
		files: [
			'mixins-guards',
			'no-sm-operations'
		]
	},
	{
		suite: 'compression',
		folder: 'compression/',
		options: {
			math: 'strict',
			compress: true
		},
		files: [
			'compression'
		]
	},
	{
		suite: 'units-strict',
		folder: 'units/strict/',
		options: {
			math: 0,
			strictUnits: true
		},
		files: [
			'strict-units'
		]
	},
	{
		suite: 'units-no-strict',
		folder: 'units/no-strict/',
		options: {
			math: 0,
			strictUnits: false
		},
		files: [
			'no-strict'
		]
	},
	{
		suite: 'rewrite-urls',
		folder: 'rewrite-urls-all/',
		options: {
			rewriteUrls: 'all'
		},
		files: [
			'rewrite-urls-all'
		]
	},
	{
		suite: 'include-path',
		folder: 'include-path/',
		options: {
			paths: ['data/', '_main/import/']
		},
		files: [
			'include-path'
		]
	}
];

// Prepare test data
function prepareTests() {
	const testDataRoot = path.join(__dirname, '../../test-data');
	const lessRoot = path.join(testDataRoot, 'less');
	const tests = [];

	for (const suite of benchmarkTestFiles) {
		for (const fileName of suite.files) {
			const lessFile = path.join(lessRoot, suite.folder, fileName + '.less');
			try {
				const content = fs.readFileSync(lessFile, 'utf8');
				tests.push({
					name: `${suite.suite}/${fileName}`,
					content,
					options: { ...suite.options, filename: lessFile },
					file: lessFile
				});
			} catch (err) {
				console.warn(`Warning: Could not read ${lessFile}`);
			}
		}
	}

	return tests;
}

// High-resolution timing
function getTime() {
	const hrtime = process.hrtime();
	return hrtime[0] * 1000 + hrtime[1] / 1000000;
}

// Run benchmark for a single test
async function benchmarkTest(test, runCount) {
	const times = {
		total: [],
		parse: [],
		eval: [],
		coldStart: null // Track the very first iteration
	};

	for (let i = 0; i < runCount; i++) {
		const startTotal = getTime();

		try {
			// Use less.render() which handles both parse and eval
			// We can't easily separate parse/eval with the public API, so we'll time the full render
			await new Promise((resolve, reject) => {
				less.render(test.content, test.options, (err, result) => {
					if (err) {
						reject(err);
					} else {
						resolve(result);
					}
				});
			});
			const endTotal = getTime();

			const totalTime = endTotal - startTotal;
			times.total.push(totalTime);

			// Capture cold-start time (first iteration before any warmup)
			if (i === 0) {
				times.coldStart = totalTime;
			}

			// For now, we can't easily separate parse/eval without using internal APIs
			// So we'll just record the total time
			times.parse.push(0);
			times.eval.push(totalTime);
		} catch (err) {
			// Skip this run if there's an error
			continue;
		}
	}

	return times;
}

// Calculate statistics
function calculateStats(times, warmupRuns) {
	// Remove warmup runs
	const validTimes = times.slice(warmupRuns);

	if (validTimes.length === 0) {
		return null;
	}

	const sum = validTimes.reduce((a, b) => a + b, 0);
	const avg = sum / validTimes.length;
	const min = Math.min(...validTimes);
	const max = Math.max(...validTimes);
	const variation = max - min;
	const variationPerc = (variation / avg) * 100;

	// Calculate median
	const sorted = [...validTimes].sort((a, b) => a - b);
	const median = sorted[Math.floor(sorted.length / 2)];

	return {
		avg: avg,
		min: min,
		max: max,
		median: median,
		variation: variation,
		variationPerc: variationPerc,
		count: validTimes.length
	};
}

// Format time for display
function formatTime(ms) {
	if (ms < 1) {
		return `${(ms * 1000).toFixed(2)}µs`;
	} else if (ms < 1000) {
		return `${ms.toFixed(2)}ms`;
	} else {
		return `${(ms / 1000).toFixed(2)}s`;
	}
}

// Print detailed results
function printResults(results, runCount, showIndividual = false) {
	console.log('\n' + '='.repeat(80));
	console.log('LESS.JS BENCHMARK RESULTS');
	console.log('='.repeat(80));
	console.log(`Total tests: ${results.length}`);
	console.log(`Runs per test: ${runCount} (${WARMUP_RUNS} warmup)`);
	console.log('='.repeat(80));

	// Calculate overall statistics
	const allTotalTimes = results.flatMap(r => r.times.total.slice(WARMUP_RUNS));
	const allParseTimes = results.flatMap(r => r.times.parse.slice(WARMUP_RUNS));
	const allEvalTimes = results.flatMap(r => r.times.eval.slice(WARMUP_RUNS));
	const allColdStarts = results.map(r => r.times.coldStart).filter(t => t != null);

	const totalStats = calculateStats(allTotalTimes, 0);
	const parseStats = calculateStats(allParseTimes, 0);
	const evalStats = calculateStats(allEvalTimes, 0);
	const coldStartStats = calculateStats(allColdStarts, 0);

	console.log('\n📊 OVERALL STATISTICS (all tests combined)');
	console.log('-'.repeat(80));

	if (coldStartStats) {
		console.log('\n🥶 Cold Start (1st iteration, no warmup):');
		console.log(`   Average: ${formatTime(coldStartStats.avg)} ± ${coldStartStats.variationPerc.toFixed(1)}%`);
		console.log(`   Median:  ${formatTime(coldStartStats.median)}`);
		console.log(`   Min:     ${formatTime(coldStartStats.min)}`);
		console.log(`   Max:     ${formatTime(coldStartStats.max)}`);
	}

	if (totalStats) {
		console.log('\n🔥 Warm Performance (after warmup):');
		console.log(`   Average: ${formatTime(totalStats.avg)} ± ${totalStats.variationPerc.toFixed(1)}%`);
		console.log(`   Median:  ${formatTime(totalStats.median)}`);
		console.log(`   Min:     ${formatTime(totalStats.min)}`);
		console.log(`   Max:     ${formatTime(totalStats.max)}`);
	}

	if (coldStartStats && totalStats) {
		const warmupEffect = ((coldStartStats.avg - totalStats.avg) / coldStartStats.avg * 100);
		console.log(`\n📈 Warmup Effect: ${warmupEffect.toFixed(1)}% faster after warmup`);
	}

	if (parseStats) {
		console.log('\n📝 Parse Time:');
		console.log(`   Average: ${formatTime(parseStats.avg)} ± ${parseStats.variationPerc.toFixed(1)}%`);
		console.log(`   Median:  ${formatTime(parseStats.median)}`);
		console.log(`   Min:     ${formatTime(parseStats.min)}`);
		console.log(`   Max:     ${formatTime(parseStats.max)}`);
	}

	if (evalStats) {
		console.log('\n⚡ Eval Time:');
		console.log(`   Average: ${formatTime(evalStats.avg)} ± ${evalStats.variationPerc.toFixed(1)}%`);
		console.log(`   Median:  ${formatTime(evalStats.median)}`);
		console.log(`   Min:     ${formatTime(evalStats.min)}`);
		console.log(`   Max:     ${formatTime(evalStats.max)}`);
	}

	if (showIndividual) {
		console.log('\n\n📋 INDIVIDUAL TEST RESULTS');
		console.log('='.repeat(80));

		for (const result of results) {
			const stats = calculateStats(result.times.total, WARMUP_RUNS);
			if (stats) {
				console.log(`\n${result.name}:`);
				console.log(`   Cold: ${formatTime(result.times.coldStart)}, Warm Avg: ${formatTime(stats.avg)}, Med: ${formatTime(stats.median)}`);
			}
		}
	}

	console.log('\n' + '='.repeat(80));
}

// Main execution
async function main() {
	const showIndividual = process.argv.includes('--detailed') || process.argv.includes('-d');
	const customRuns = parseInt(process.argv.find(arg => arg.startsWith('--runs='))?.split('=')[1]);
	const runCount = customRuns || TOTAL_RUNS;

	console.log('🚀 Starting Less.js Benchmark Suite...');
	console.log(`   Runs per test: ${runCount} (${WARMUP_RUNS} warmup runs)`);

	const tests = prepareTests();
	console.log(`   Total tests: ${tests.length}`);
	console.log('');

	const results = [];
	let completed = 0;

	for (const test of tests) {
		process.stdout.write(`\rRunning benchmarks... ${++completed}/${tests.length} (${((completed / tests.length) * 100).toFixed(1)}%)`);

		const times = await benchmarkTest(test, runCount);
		results.push({
			name: test.name,
			times: times
		});
	}

	process.stdout.write('\n');
	printResults(results, runCount, showIndividual);

	// Output JSON format if requested
	if (process.argv.includes('--json')) {
		const jsonOutput = {
			timestamp: new Date().toISOString(),
			runs: runCount,
			warmupRuns: WARMUP_RUNS,
			tests: results.map(r => ({
				name: r.name,
				coldStart: r.times.coldStart,
				total: calculateStats(r.times.total, WARMUP_RUNS),
				parse: calculateStats(r.times.parse, WARMUP_RUNS),
				eval: calculateStats(r.times.eval, WARMUP_RUNS)
			}))
		};
		console.log('\n\nJSON OUTPUT:');
		console.log(JSON.stringify(jsonOutput, null, 2));
	}
}

// Run the benchmark
main().catch(err => {
	console.error('Benchmark failed:', err);
	process.exit(1);
});
