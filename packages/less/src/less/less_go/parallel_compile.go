package less_go

import (
	"fmt"
	"runtime"
	"sync"
)

// ParallelCompileOptions controls parallelization behavior
type ParallelCompileOptions struct {
	// Enable enables parallel compilation (default: false for safety)
	Enable bool

	// MaxWorkers limits the number of concurrent workers (default: runtime.NumCPU())
	// Set to 0 to use runtime.NumCPU()
	MaxWorkers int

	// StopOnError stops all compilation if any file fails (default: false)
	StopOnError bool
}

// CompileJob represents a single compilation job
type CompileJob struct {
	// Input is the LESS source code
	Input string

	// Options for this specific compilation
	Options map[string]any

	// ID is an optional identifier for this job (e.g., filename)
	ID string
}

// CompileResult represents the result of a compilation
type CompileResult struct {
	// ID matches the CompileJob.ID
	ID string

	// CSS is the compiled CSS output
	CSS string

	// Error is set if compilation failed
	Error error

	// Index is the original job index
	Index int
}

// BatchCompile compiles multiple LESS inputs, optionally in parallel
// Returns results in the same order as the input jobs
func BatchCompile(factory map[string]any, jobs []CompileJob, parallelOpts *ParallelCompileOptions) []CompileResult {
	if parallelOpts == nil {
		parallelOpts = &ParallelCompileOptions{
			Enable:      false,
			MaxWorkers:  0,
			StopOnError: false,
		}
	}

	// If parallel compilation is disabled, compile sequentially
	if !parallelOpts.Enable {
		return batchCompileSequential(factory, jobs)
	}

	// Parallel compilation
	return batchCompileParallel(factory, jobs, parallelOpts)
}

// batchCompileSequential compiles jobs one at a time
func batchCompileSequential(factory map[string]any, jobs []CompileJob) []CompileResult {
	results := make([]CompileResult, len(jobs))

	for i, job := range jobs {
		css, err := compileLessForTest(factory, job.Input, job.Options)
		results[i] = CompileResult{
			ID:    job.ID,
			CSS:   css,
			Error: err,
			Index: i,
		}
	}

	return results
}

// batchCompileParallel compiles jobs in parallel using worker pool
func batchCompileParallel(factory map[string]any, jobs []CompileJob, opts *ParallelCompileOptions) []CompileResult {
	numJobs := len(jobs)
	if numJobs == 0 {
		return []CompileResult{}
	}

	// Determine number of workers
	maxWorkers := opts.MaxWorkers
	if maxWorkers <= 0 {
		maxWorkers = runtime.NumCPU()
	}

	// Don't create more workers than jobs
	if maxWorkers > numJobs {
		maxWorkers = numJobs
	}

	// Create channels for work distribution
	jobChan := make(chan struct {
		job   CompileJob
		index int
	}, numJobs)
	resultChan := make(chan CompileResult, numJobs)

	// Error context for StopOnError mode
	var (
		stopMutex sync.Mutex
		stopped   bool
	)

	checkStopped := func() bool {
		stopMutex.Lock()
		defer stopMutex.Unlock()
		return stopped
	}

	setStopped := func() {
		stopMutex.Lock()
		defer stopMutex.Unlock()
		stopped = true
	}

	// Worker function
	worker := func(wg *sync.WaitGroup) {
		defer wg.Done()

		for work := range jobChan {
			// Check if we should stop
			if opts.StopOnError && checkStopped() {
				// Send empty result to maintain result count
				resultChan <- CompileResult{
					ID:    work.job.ID,
					Error: fmt.Errorf("compilation stopped due to error in another job"),
					Index: work.index,
				}
				continue
			}

			// Compile the LESS file
			css, err := compileLessForTest(factory, work.job.Input, work.job.Options)

			// If error occurred and StopOnError is enabled, signal stop
			if err != nil && opts.StopOnError {
				setStopped()
			}

			// Send result
			resultChan <- CompileResult{
				ID:    work.job.ID,
				CSS:   css,
				Error: err,
				Index: work.index,
			}
		}
	}

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go worker(&wg)
	}

	// Send jobs to workers
	go func() {
		for i, job := range jobs {
			jobChan <- struct {
				job   CompileJob
				index int
			}{job: job, index: i}
		}
		close(jobChan)
	}()

	// Collect results
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Gather all results
	results := make([]CompileResult, numJobs)
	for result := range resultChan {
		results[result.Index] = result
	}

	return results
}

// ParallelCompileMultipleFiles is a convenience function that compiles multiple LESS files in parallel
// This is useful for build tools and batch processors
func ParallelCompileMultipleFiles(inputs []struct {
	Content  string
	Options  map[string]any
	Filename string
}, enableParallel bool) []CompileResult {
	// Create factory once (can be reused across compilations)
	factory := Factory(nil, nil)

	// Convert inputs to jobs
	jobs := make([]CompileJob, len(inputs))
	for i, input := range inputs {
		// Ensure filename is set in options
		if input.Options == nil {
			input.Options = make(map[string]any)
		}
		if _, ok := input.Options["filename"]; !ok {
			input.Options["filename"] = input.Filename
		}

		jobs[i] = CompileJob{
			Input:   input.Content,
			Options: input.Options,
			ID:      input.Filename,
		}
	}

	// Compile with or without parallelization
	return BatchCompile(factory, jobs, &ParallelCompileOptions{
		Enable:      enableParallel,
		MaxWorkers:  0, // Use all CPUs
		StopOnError: false,
	})
}
