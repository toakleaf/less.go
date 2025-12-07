/**
 * Comprehensive benchmark suite for less.js
 * Tests the same files as the Go benchmark for fair comparison
 * Supports both Node.js and Bun runtimes
 */

const path = require('path');
const fs = require('fs');
const less = require('less');

// Runtime detection
const isBun = typeof Bun !== 'undefined';
const runtime = isBun ? 'bun' : 'node';
const runtimeVersion = isBun ? Bun.version : process.version;

// Configuration
const TOTAL_RUNS = 30;
const WARMUP_RUNS = 5;

// Test suites matching the Go benchmarks
// PURE GO TESTS: These tests do NOT require Node.js/plugin support
const benchmarkTestFiles = [
	{
		suite: 'main',
		folder: '_main/',
		options: {
			relativeUrls: true,
			silent: true,
			javascriptEnabled: true
		},
		// All passing _main tests EXCEPT those requiring plugins/Node.js or network
		// Excluded: import, import-module, javascript, plugin, plugin-module, plugin-preeval
		// Also excluded: import-remote (makes network requests to cdn.jsdelivr.net)
		files: [
			'calc',
			'charsets',
			'colors',
			'colors2',
			'comments',
			'comments2',
			'container',
			'css-3',
			'css-escapes',
			'css-grid',
			'css-guards',
			'detached-rulesets',
			'directives-bubling',
			'empty',
			'extend',
			'extend-chaining',
			'extend-clearfix',
			'extend-exact',
			'extend-media',
			'extend-nest',
			'extend-selector',
			'extract-and-length',
			'functions',
			'functions-each',
			'ie-filters',
			'impor',
			'import-inline',
			'import-interpolation',
			'import-once',
			'import-reference',
			'import-reference-issues',
			'lazy-eval',
			'media',
			'merge',
			'mixin-noparens',
			'mixins',
			'mixins-closure',
			'mixins-guards',
			'mixins-guards-default-func',
			'mixins-important',
			'mixins-interpolated',
			'mixins-named-args',
			'mixins-nested',
			'mixins-pattern',
			'no-output',
			'operations',
			'parse-interpolation',
			'permissive-parse',
			'plugi',
			'property-accessors',
			'property-name-interp',
			'rulesets',
			'scope',
			'selectors',
			'strings',
			'urls',
			'variables',
			'variables-in-at-rules',
			'whitespace'
		]
	},
	{
		suite: 'namespacing',
		folder: 'namespacing/',
		options: {},
		// Excluded: namespacing-3 (context bug), namespacing-media (undefined namespace)
		files: [
			'namespacing-1',
			'namespacing-2',
			'namespacing-4',
			'namespacing-5',
			'namespacing-6',
			'namespacing-7',
			'namespacing-8',
			'namespacing-functions',
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
		suite: 'static-urls',
		folder: 'static-urls/',
		options: {
			math: 'strict',
			relativeUrls: false,
			rootpath: 'folder (1)/'
		},
		files: [
			'urls'
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
		suite: 'url-args',
		folder: 'url-args/',
		options: {
			urlArgs: '424242'
		},
		files: [
			'urls'
		]
	},
	{
		suite: 'rewrite-urls-all',
		folder: 'rewrite-urls-all/',
		options: {
			rewriteUrls: 'all'
		},
		files: [
			'rewrite-urls-all'
		]
	},
	{
		suite: 'rewrite-urls-local',
		folder: 'rewrite-urls-local/',
		options: {
			rewriteUrls: 'local'
		},
		files: [
			'rewrite-urls-local'
		]
	},
	{
		suite: 'rootpath-rewrite-urls-all',
		folder: 'rootpath-rewrite-urls-all/',
		options: {
			rootpath: 'http://example.com/assets/css/',
			rewriteUrls: 'all'
		},
		files: [
			'rootpath-rewrite-urls-all'
		]
	},
	{
		suite: 'rootpath-rewrite-urls-local',
		folder: 'rootpath-rewrite-urls-local/',
		options: {
			rootpath: 'http://example.com/assets/css/',
			rewriteUrls: 'local'
		},
		files: [
			'rootpath-rewrite-urls-local'
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
	},
	{
		suite: 'include-path-string',
		folder: 'include-path-string/',
		options: {
			paths: 'data/'
		},
		files: [
			'include-path-string'
		]
	},
	{
		suite: 'process-imports',
		folder: 'process-imports/',
		options: {
			processImports: false
		},
		files: [
			'google'
		]
	},
	{
		suite: 'custom',
		folder: 'custom/',
		options: {
			relativeUrls: true,
			silent: true,
			javascriptEnabled: true
		},
		// All custom integration tests
		// Excluded: import-module-variation (requires npm modules), var-javascript (requires Node.js)
		files: [
			'calc-variations',
			'charsets-variations',
			'colors-variations',
			'colors2-variations',
			'comments-variations',
			'comments2-variations',
			'container-variations',
			'css-escapes-variations',
			'css-grid-variations',
			'css-guards-variation',
			'css-output-variation',
			'css3-variations',
			'detached-ruleset-preeval-variation',
			'detached-rulesets-variation',
			'directives-bubbling-variation',
			'empty-variation',
			'example-nesting',
			'example-variables',
			'extend-chaining-variation',
			'extend-clearfix-variation',
			'extend-exact-variation',
			'extend-media-variation',
			'extend-nest-variation',
			'extend-selector-variation',
			'extend-variation',
			'extract-and-length-variation',
			'fade-variable-alpha',
			'functions-each-variation',
			'functions-variation',
			'ie-filters-variation',
			'impor-variation',
			'import-inline-variation',
			'import-interpolation-variation',
			'import-once-variation',
			'import-reference-issues-variation',
			'math-always-mixins-guards-variation',
			'math-always-no-sm-operations-variation',
			'math-parens-division-media-math-variation',
			'math-parens-division-mixins-args-variation',
			'math-parens-division-new-division-variation',
			'math-parens-division-parens-variation',
			'math-strict-parens-variation',
			'media-math-variation',
			'mixins-args-variation',
			'mixins-interpolated-var',
			'mixins-named-args-var',
			'mixins-nested-var',
			'mixins-pattern-var',
			'mixins-var',
			'module-pattern-variation',
			'multi-target-extend',
			'multi-var-import',
			'namespacing-7-variation',
			'namespacing-8-variation',
			'namespacing-functions-var',
			'namespacing-media-var',
			'namespacing-operations-var',
			'namespacing-var-1',
			'namespacing-var-2',
			'namespacing-var-3',
			'namespacing-var-4',
			'namespacing-var-5',
			'namespacing-var-6',
			'operations-var',
			'operations-variations',
			'parenthesized-list-values',
			'parse-interpolation-var',
			'parse-interpolation-variations',
			'permissive-parse-var',
			'permissive-parse-variations',
			'plugin-functions-simulation',
			'pre-post-process-simulation',
			'property-accessors-variation',
			'property-accessors-variations',
			'property-name-interp-variation',
			'property-name-interp-variations',
			'property-name-interp-variations-v2',
			'rewrite-paths-variation',
			'rgba-fade-alpha',
			'rgba-variable-alpha',
			'root-variable-scope-variation',
			'rootpath-url-variation',
			'rulesets-variations',
			'rulesets-variations-v2',
			'scope-variations',
			'scope-variations-v2',
			'selectors-variations',
			'selectors-variations-v2',
			'starting-style-variations',
			'starting-style-variations-v2',
			'strings-variations',
			'strings-variations-v2',
			'url-handling-variation',
			'urls-variations',
			'urls-variations-v2',
			'var-detached-rulesets',
			'var-extend-variations',
			'var-import',
			'var-import-reference',
			'var-layer',
			'var-lazy-eval',
			'var-media',
			'var-merge',
			'var-mixin-guards',
			'var-mixin-noparens',
			'var-mixins',
			'var-mixins-closure',
			'var-mixins-guards',
			'var-mixins-guards-default',
			'var-mixins-important',
			'var-mixins-interpolated',
			'var-mixins-named-args',
			'var-mixins-nested',
			'var-mixins-pattern',
			'var-no-output',
			'var-operations',
			'variables-at-rules',
			'variables-at-rules-variations-v2',
			'variables-variations',
			'variables-variations-v2',
			'whitespace-variations',
			'whitespace-variations-v2'
		]
	}
];

// Prepare test data
function prepareTests() {
	const testDataRoot = path.join(__dirname, '../testdata');
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
// Use Bun.nanoseconds() for more precise timing in Bun, fallback to process.hrtime()
function getTime() {
	if (isBun) {
		return Bun.nanoseconds() / 1_000_000; // Convert to milliseconds
	}
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
	console.log(`LESS.JS BENCHMARK RESULTS (${runtime} ${runtimeVersion})`);
	console.log('='.repeat(80));
	console.log(`Runtime: ${runtime} ${runtimeVersion}`);
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

// Run suite benchmark - single pass through all files
// This represents a single build session
async function benchmarkSuiteSingleRun(tests) {
	const startTotal = getTime();

	// Compile all files in sequence (one build session)
	for (const test of tests) {
		try {
			await new Promise((resolve, reject) => {
				less.render(test.content, test.options, (err, result) => {
					if (err) {
						reject(err);
					} else {
						resolve(result);
					}
				});
			});
		} catch (err) {
			// Skip this file if there's an error
			continue;
		}
	}

	const endTotal = getTime();
	return endTotal - startTotal;
}

// Run suite benchmark (all files sequentially, repeated N times IN THE SAME PROCESS)
// Note: This is NOT realistic for CLI tools - use --single-run for realistic benchmarking
async function benchmarkSuite(tests, runCount) {
	const times = {
		total: [],
		coldStart: null
	};

	for (let i = 0; i < runCount; i++) {
		const totalTime = await benchmarkSuiteSingleRun(tests);
		times.total.push(totalTime);

		// Capture cold-start time (first iteration before any warmup)
		if (i === 0) {
			times.coldStart = totalTime;
		}
	}

	return times;
}

// Print suite results
function printSuiteResults(times, testCount, runCount) {
	console.log('\n' + '='.repeat(80));
	console.log(`LESS.JS SUITE BENCHMARK RESULTS (${runtime} ${runtimeVersion})`);
	console.log('='.repeat(80));
	console.log(`Runtime: ${runtime} ${runtimeVersion}`);
	console.log(`Total files: ${testCount}`);
	console.log(`Suite runs: ${runCount} (${WARMUP_RUNS} warmup)`);
	console.log(`Methodology: All ${testCount} files compiled sequentially per iteration`);
	console.log('='.repeat(80));

	const stats = calculateStats(times.total, WARMUP_RUNS);
	const coldStart = times.coldStart;

	console.log('\n📊 SUITE PERFORMANCE (all files compiled sequentially)');
	console.log('-'.repeat(80));

	if (coldStart) {
		console.log('\n🥶 Cold Start (1st iteration, no warmup):');
		console.log(`   Total time: ${formatTime(coldStart)}`);
		console.log(`   Per file:   ${formatTime(coldStart / testCount)}`);
	}

	if (stats) {
		console.log('\n🔥 Warm Performance (after warmup):');
		console.log(`   Average:    ${formatTime(stats.avg)} ± ${stats.variationPerc.toFixed(1)}%`);
		console.log(`   Median:     ${formatTime(stats.median)}`);
		console.log(`   Min:        ${formatTime(stats.min)}`);
		console.log(`   Max:        ${formatTime(stats.max)}`);
		console.log(`   Per file:   ${formatTime(stats.avg / testCount)}`);
	}

	if (coldStart && stats) {
		const warmupEffect = ((coldStart - stats.avg) / coldStart * 100);
		console.log(`\n📈 Warmup Effect: ${warmupEffect.toFixed(1)}% faster after warmup`);
	}

	console.log('\n' + '='.repeat(80));
}

// Main execution
async function main() {
	const showIndividual = process.argv.includes('--detailed') || process.argv.includes('-d');
	const suiteMode = process.argv.includes('--suite') || process.argv.includes('-s');
	const singleRun = process.argv.includes('--single-run');
	const customRuns = parseInt(process.argv.find(arg => arg.startsWith('--runs='))?.split('=')[1]);
	const runCount = customRuns || TOTAL_RUNS;

	const tests = prepareTests();

	if (singleRun) {
		// Single-run mode: compile all files once and exit
		// This simulates a single build session (realistic for CLI tools)
		const totalTime = await benchmarkSuiteSingleRun(tests);

		// Output JSON for easy parsing by wrapper scripts
		console.log(JSON.stringify({
			timestamp: new Date().toISOString(),
			runtime: runtime,
			runtimeVersion: runtimeVersion,
			mode: 'single-run',
			testCount: tests.length,
			totalTime: totalTime,
			perFileAvg: totalTime / tests.length
		}));
	} else if (suiteMode) {
		// Suite mode: compile all files sequentially, repeat N times
		console.log('🚀 Starting Less.js Suite Benchmark...');
		console.log(`   Suite runs: ${runCount} (${WARMUP_RUNS} warmup runs)`);
		console.log(`   Total files per suite: ${tests.length}`);
		console.log(`   Total compilations: ${tests.length * runCount}`);
		console.log('');

		process.stdout.write('Running suite benchmark...');
		const times = await benchmarkSuite(tests, runCount);
		process.stdout.write(' done!\n');

		printSuiteResults(times, tests.length, runCount);

		// Output JSON format if requested
		if (process.argv.includes('--json')) {
			const stats = calculateStats(times.total, WARMUP_RUNS);
			const jsonOutput = {
				timestamp: new Date().toISOString(),
				runtime: runtime,
				runtimeVersion: runtimeVersion,
				mode: 'suite',
				runs: runCount,
				warmupRuns: WARMUP_RUNS,
				testCount: tests.length,
				coldStart: times.coldStart,
				warm: stats
			};
			console.log('\n\nJSON OUTPUT:');
			console.log(JSON.stringify(jsonOutput, null, 2));
		}
	} else {
		// Individual file mode: compile each file N times
		console.log('🚀 Starting Less.js Benchmark Suite...');
		console.log(`   Runs per test: ${runCount} (${WARMUP_RUNS} warmup runs)`);
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
				runtime: runtime,
				runtimeVersion: runtimeVersion,
				mode: 'individual',
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
}

// Run the benchmark
main().catch(err => {
	console.error('Benchmark failed:', err);
	process.exit(1);
});
