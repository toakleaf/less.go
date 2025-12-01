package less_go

import (
	"os"
	"runtime"
	"sync"
)

// ParallelEvalConfig controls parallel evaluation behavior
type ParallelEvalConfig struct {
	// Enabled controls whether parallel evaluation is used
	Enabled bool

	// MinSelectorsForParallel is the minimum number of selectors required
	// to trigger parallel evaluation. Below this threshold, sequential
	// evaluation is faster due to goroutine overhead.
	MinSelectorsForParallel int

	// MaxWorkers is the maximum number of parallel workers.
	// Defaults to runtime.GOMAXPROCS(0) if set to 0.
	MaxWorkers int
}

// DefaultParallelEvalConfig returns the default configuration
func DefaultParallelEvalConfig() *ParallelEvalConfig {
	return &ParallelEvalConfig{
		Enabled:                 true,
		MinSelectorsForParallel: 8, // Only parallelize if 8+ selectors
		MaxWorkers:              0, // Use GOMAXPROCS
	}
}

// globalParallelConfig is the singleton configuration
var globalParallelConfig = DefaultParallelEvalConfig()
var parallelConfigMu sync.RWMutex

// GetParallelEvalConfig returns the current parallel evaluation configuration
func GetParallelEvalConfig() *ParallelEvalConfig {
	parallelConfigMu.RLock()
	defer parallelConfigMu.RUnlock()
	// Return a copy to prevent races
	return &ParallelEvalConfig{
		Enabled:                 globalParallelConfig.Enabled,
		MinSelectorsForParallel: globalParallelConfig.MinSelectorsForParallel,
		MaxWorkers:              globalParallelConfig.MaxWorkers,
	}
}

// SetParallelEvalEnabled enables or disables parallel evaluation
func SetParallelEvalEnabled(enabled bool) {
	parallelConfigMu.Lock()
	globalParallelConfig.Enabled = enabled
	parallelConfigMu.Unlock()
}

// SetParallelEvalMinSelectors sets the minimum selectors threshold
func SetParallelEvalMinSelectors(min int) {
	parallelConfigMu.Lock()
	globalParallelConfig.MinSelectorsForParallel = min
	parallelConfigMu.Unlock()
}

// SetParallelEvalMaxWorkers sets the maximum number of workers
func SetParallelEvalMaxWorkers(max int) {
	parallelConfigMu.Lock()
	globalParallelConfig.MaxWorkers = max
	parallelConfigMu.Unlock()
}

// getWorkerCount returns the number of workers to use
func getWorkerCount(cfg *ParallelEvalConfig) int {
	if cfg.MaxWorkers > 0 {
		return cfg.MaxWorkers
	}
	return runtime.GOMAXPROCS(0)
}

// SelectorEvalResult holds the result of a parallel selector evaluation
type SelectorEvalResult struct {
	Index                 int
	Selector              *Selector
	Error                 error
	HasVariable           bool
	HasOnePassingSelector bool
}

// SelectorEvalResults holds the aggregated results of selector evaluation
type SelectorEvalResults struct {
	Selectors             []*Selector
	HasVariable           bool
	HasOnePassingSelector bool
}

// EvalSelectorsParallel evaluates selectors in parallel if beneficial
// Returns the evaluated selectors in order with flag information
func EvalSelectorsParallel(selectors []any, context any) (*SelectorEvalResults, error) {
	cfg := GetParallelEvalConfig()

	// Check if parallel evaluation should be used
	if !cfg.Enabled || len(selectors) < cfg.MinSelectorsForParallel {
		return evalSelectorsSequential(selectors, context)
	}

	// Check environment override
	if os.Getenv("LESS_GO_PARALLEL") == "0" {
		return evalSelectorsSequential(selectors, context)
	}

	return evalSelectorsParallelImpl(selectors, context, cfg)
}

// evalSelectorsSequential evaluates selectors sequentially (baseline)
func evalSelectorsSequential(selectors []any, context any) (*SelectorEvalResults, error) {
	results := &SelectorEvalResults{
		Selectors: make([]*Selector, 0, len(selectors)),
	}

	for _, sel := range selectors {
		if sel == nil {
			continue
		}

		selector, ok := sel.(*Selector)
		if !ok {
			continue
		}

		evaluatedAny, err := selector.Eval(context)
		if err != nil {
			return nil, err
		}

		if evaluatedAny != nil {
			if evaluated, ok := evaluatedAny.(*Selector); ok {
				results.Selectors = append(results.Selectors, evaluated)

				// Check for variables in elements
				for _, elem := range evaluated.Elements {
					if elem.IsVariable {
						results.HasVariable = true
						break
					}
				}

				// Check for passing condition
				if evaluated.EvaldCondition {
					results.HasOnePassingSelector = true
				}
			}
		}
	}

	return results, nil
}

// evalSelectorsParallelImpl performs actual parallel evaluation
func evalSelectorsParallelImpl(selectors []any, context any, cfg *ParallelEvalConfig) (*SelectorEvalResults, error) {
	n := len(selectors)
	workers := getWorkerCount(cfg)
	if workers > n {
		workers = n
	}

	// Create channels for work distribution
	jobs := make(chan int, n)
	resultsChan := make(chan SelectorEvalResult, n)

	// Launch workers
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobs {
				sel := selectors[idx]
				if sel == nil {
					resultsChan <- SelectorEvalResult{Index: idx, Selector: nil, Error: nil}
					continue
				}

				selector, ok := sel.(*Selector)
				if !ok {
					resultsChan <- SelectorEvalResult{Index: idx, Selector: nil, Error: nil}
					continue
				}

				// Each goroutine gets a read-only view of the context
				// Selector.Eval should be safe for concurrent read access
				evaluatedAny, err := selector.Eval(context)
				var evaluated *Selector
				var hasVariable, hasOnePassingSelector bool

				if evaluatedAny != nil {
					if evalSel, ok := evaluatedAny.(*Selector); ok {
						evaluated = evalSel

						// Check for variables in elements
						for _, elem := range evalSel.Elements {
							if elem.IsVariable {
								hasVariable = true
								break
							}
						}

						// Check for passing condition
						if evalSel.EvaldCondition {
							hasOnePassingSelector = true
						}
					}
				}

				resultsChan <- SelectorEvalResult{
					Index:                 idx,
					Selector:              evaluated,
					Error:                 err,
					HasVariable:           hasVariable,
					HasOnePassingSelector: hasOnePassingSelector,
				}
			}
		}()
	}

	// Send jobs
	for i := 0; i < n; i++ {
		jobs <- i
	}
	close(jobs)

	// Wait for completion and collect results
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results in order
	evalResults := make([]*Selector, n)
	var firstError error
	var hasVariable, hasOnePassingSelector bool

	for result := range resultsChan {
		if result.Error != nil && firstError == nil {
			firstError = result.Error
		}
		evalResults[result.Index] = result.Selector
		if result.HasVariable {
			hasVariable = true
		}
		if result.HasOnePassingSelector {
			hasOnePassingSelector = true
		}
	}

	if firstError != nil {
		return nil, firstError
	}

	// Filter out nil selectors and compact the result
	compactedSelectors := make([]*Selector, 0, n)
	for _, sel := range evalResults {
		if sel != nil {
			compactedSelectors = append(compactedSelectors, sel)
		}
	}

	return &SelectorEvalResults{
		Selectors:             compactedSelectors,
		HasVariable:           hasVariable,
		HasOnePassingSelector: hasOnePassingSelector,
	}, nil
}

// RuleEvalResult holds the result of a parallel rule evaluation
type RuleEvalResult struct {
	Index int
	Rule  any
	Error error
}

// ParallelVisitorConfig controls parallel visitor behavior
type ParallelVisitorConfig struct {
	Enabled     bool
	MinNodes    int // Minimum nodes to trigger parallel processing
	MaxWorkers  int
}

// DefaultParallelVisitorConfig returns default visitor parallelization config
func DefaultParallelVisitorConfig() *ParallelVisitorConfig {
	return &ParallelVisitorConfig{
		Enabled:    false, // Disabled by default - visitors often modify state
		MinNodes:   16,
		MaxWorkers: 0,
	}
}
