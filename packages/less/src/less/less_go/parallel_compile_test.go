package less_go

import (
	"fmt"
	"strings"
	"testing"
)

func TestBatchCompile(t *testing.T) {
	factory := Factory(nil, nil)

	tests := []struct {
		name        string
		jobs        []CompileJob
		parallel    bool
		expectError bool
	}{
		{
			name: "Sequential - Single file",
			jobs: []CompileJob{
				{
					Input: `.test { color: red; }`,
					Options: map[string]any{
						"filename": "test.less",
					},
					ID: "test.less",
				},
			},
			parallel:    false,
			expectError: false,
		},
		{
			name: "Sequential - Multiple files",
			jobs: []CompileJob{
				{
					Input: `.test1 { color: red; }`,
					Options: map[string]any{
						"filename": "test1.less",
					},
					ID: "test1.less",
				},
				{
					Input: `.test2 { color: blue; }`,
					Options: map[string]any{
						"filename": "test2.less",
					},
					ID: "test2.less",
				},
				{
					Input: `.test3 { color: green; }`,
					Options: map[string]any{
						"filename": "test3.less",
					},
					ID: "test3.less",
				},
			},
			parallel:    false,
			expectError: false,
		},
		{
			name: "Parallel - Multiple files",
			jobs: []CompileJob{
				{
					Input: `.test1 { color: red; }`,
					Options: map[string]any{
						"filename": "test1.less",
					},
					ID: "test1.less",
				},
				{
					Input: `.test2 { color: blue; }`,
					Options: map[string]any{
						"filename": "test2.less",
					},
					ID: "test2.less",
				},
				{
					Input: `.test3 { color: green; }`,
					Options: map[string]any{
						"filename": "test3.less",
					},
					ID: "test3.less",
				},
			},
			parallel:    true,
			expectError: false,
		},
		{
			name: "Parallel - With variables",
			jobs: []CompileJob{
				{
					Input: `@color: red; .test1 { color: @color; }`,
					Options: map[string]any{
						"filename": "test1.less",
					},
					ID: "test1.less",
				},
				{
					Input: `@color: blue; .test2 { color: @color; }`,
					Options: map[string]any{
						"filename": "test2.less",
					},
					ID: "test2.less",
				},
			},
			parallel:    true,
			expectError: false,
		},
		{
			name: "Parallel - With operations",
			jobs: []CompileJob{
				{
					Input: `.test1 { width: 10px + 5px; }`,
					Options: map[string]any{
						"filename": "test1.less",
					},
					ID: "test1.less",
				},
				{
					Input: `.test2 { width: 20px * 2; }`,
					Options: map[string]any{
						"filename": "test2.less",
					},
					ID: "test2.less",
				},
			},
			parallel:    true,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &ParallelCompileOptions{
				Enable:      tt.parallel,
				MaxWorkers:  4,
				StopOnError: false,
			}

			results := BatchCompile(factory, tt.jobs, opts)

			// Check we got the right number of results
			if len(results) != len(tt.jobs) {
				t.Errorf("Expected %d results, got %d", len(tt.jobs), len(results))
			}

			// Check each result
			for i, result := range results {
				// Check error
				if tt.expectError {
					if result.Error == nil {
						t.Errorf("Job %d: expected error but got none", i)
					}
				} else {
					if result.Error != nil {
						t.Errorf("Job %d: unexpected error: %v", i, result.Error)
					}
				}

				// Check CSS is not empty for successful compilations
				if result.Error == nil && result.CSS == "" {
					t.Errorf("Job %d: expected non-empty CSS", i)
				}

				// Check ID matches
				if result.ID != tt.jobs[i].ID {
					t.Errorf("Job %d: expected ID %s, got %s", i, tt.jobs[i].ID, result.ID)
				}

				// Check Index matches
				if result.Index != i {
					t.Errorf("Job %d: expected Index %d, got %d", i, i, result.Index)
				}
			}
		})
	}
}

func TestBatchCompileOrdering(t *testing.T) {
	factory := Factory(nil, nil)

	// Create jobs with distinct output
	jobs := make([]CompileJob, 10)
	for i := 0; i < 10; i++ {
		jobs[i] = CompileJob{
			Input: fmt.Sprintf(`.test%d { content: "%d"; }`, i, i),
			Options: map[string]any{
				"filename": fmt.Sprintf("test%d.less", i),
			},
			ID: fmt.Sprintf("test%d", i),
		}
	}

	// Test both sequential and parallel
	for _, parallel := range []bool{false, true} {
		name := "Sequential"
		if parallel {
			name = "Parallel"
		}

		t.Run(name, func(t *testing.T) {
			opts := &ParallelCompileOptions{
				Enable:      parallel,
				MaxWorkers:  4,
				StopOnError: false,
			}

			results := BatchCompile(factory, jobs, opts)

			// Verify results are in the same order as jobs
			for i, result := range results {
				if result.ID != jobs[i].ID {
					t.Errorf("Result %d: expected ID %s, got %s", i, jobs[i].ID, result.ID)
				}
				if result.Index != i {
					t.Errorf("Result %d: expected Index %d, got %d", i, i, result.Index)
				}

				// Verify the CSS contains the correct content
				expectedContent := fmt.Sprintf(`"%d"`, i)
				if !strings.Contains(result.CSS, expectedContent) {
					t.Errorf("Result %d: expected CSS to contain %s, got:\n%s", i, expectedContent, result.CSS)
				}
			}
		})
	}
}

func TestBatchCompileEmptyJobs(t *testing.T) {
	factory := Factory(nil, nil)

	opts := &ParallelCompileOptions{
		Enable:     true,
		MaxWorkers: 4,
	}

	results := BatchCompile(factory, []CompileJob{}, opts)

	if len(results) != 0 {
		t.Errorf("Expected 0 results for empty jobs, got %d", len(results))
	}
}

func TestBatchCompileWithErrors(t *testing.T) {
	factory := Factory(nil, nil)

	jobs := []CompileJob{
		{
			Input: `.test1 { color: red; }`,
			Options: map[string]any{
				"filename": "test1.less",
			},
			ID: "test1.less",
		},
		{
			Input: `.test2 { color: @undefined`, // Syntax error - missing semicolon and closing brace
			Options: map[string]any{
				"filename": "test2.less",
			},
			ID: "test2.less",
		},
		{
			Input: `.test3 { color: blue; }`,
			Options: map[string]any{
				"filename": "test3.less",
			},
			ID: "test3.less",
		},
	}

	opts := &ParallelCompileOptions{
		Enable:      true,
		MaxWorkers:  4,
		StopOnError: false, // Continue on error
	}

	results := BatchCompile(factory, jobs, opts)

	// Should have all results
	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	// First should succeed
	if results[0].Error != nil {
		t.Errorf("Job 0: expected no error, got %v", results[0].Error)
	}

	// Second should fail (syntax error)
	if results[1].Error == nil {
		t.Errorf("Job 1: expected error for syntax error")
	}

	// Third should succeed
	if results[2].Error != nil {
		t.Errorf("Job 2: expected no error, got %v", results[2].Error)
	}
}

func TestParallelCompileMultipleFiles(t *testing.T) {
	inputs := []struct {
		Content  string
		Options  map[string]any
		Filename string
	}{
		{
			Content:  `.test1 { color: red; }`,
			Options:  map[string]any{},
			Filename: "test1.less",
		},
		{
			Content:  `.test2 { color: blue; }`,
			Options:  map[string]any{},
			Filename: "test2.less",
		},
		{
			Content:  `.test3 { color: green; }`,
			Options:  map[string]any{},
			Filename: "test3.less",
		},
	}

	// Test both parallel and sequential
	for _, enableParallel := range []bool{false, true} {
		name := "Sequential"
		if enableParallel {
			name = "Parallel"
		}

		t.Run(name, func(t *testing.T) {
			results := ParallelCompileMultipleFiles(inputs, enableParallel)

			if len(results) != len(inputs) {
				t.Errorf("Expected %d results, got %d", len(inputs), len(results))
			}

			for i, result := range results {
				if result.Error != nil {
					t.Errorf("Job %d: unexpected error: %v", i, result.Error)
				}
				if result.CSS == "" {
					t.Errorf("Job %d: expected non-empty CSS", i)
				}
				if result.ID != inputs[i].Filename {
					t.Errorf("Job %d: expected ID %s, got %s", i, inputs[i].Filename, result.ID)
				}
			}
		})
	}
}

func TestBatchCompileWorkerCounts(t *testing.T) {
	factory := Factory(nil, nil)

	// Create 20 jobs
	jobs := make([]CompileJob, 20)
	for i := 0; i < 20; i++ {
		jobs[i] = CompileJob{
			Input: fmt.Sprintf(`.test%d { color: red; }`, i),
			Options: map[string]any{
				"filename": fmt.Sprintf("test%d.less", i),
			},
			ID: fmt.Sprintf("test%d", i),
		}
	}

	// Test with different worker counts
	workerCounts := []int{1, 2, 4, 8, 0} // 0 = NumCPU
	for _, workers := range workerCounts {
		t.Run(fmt.Sprintf("Workers-%d", workers), func(t *testing.T) {
			opts := &ParallelCompileOptions{
				Enable:      true,
				MaxWorkers:  workers,
				StopOnError: false,
			}

			results := BatchCompile(factory, jobs, opts)

			// Verify all jobs completed
			if len(results) != len(jobs) {
				t.Errorf("Expected %d results, got %d", len(jobs), len(results))
			}

			// Verify all succeeded
			for i, result := range results {
				if result.Error != nil {
					t.Errorf("Job %d: unexpected error: %v", i, result.Error)
				}
			}
		})
	}
}
