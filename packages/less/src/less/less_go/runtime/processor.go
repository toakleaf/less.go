package runtime

import (
	"fmt"
)

// ProcessorInfo contains metadata about a registered JavaScript processor.
type ProcessorInfo struct {
	Index    int `json:"index"`
	Priority int `json:"priority"`
}

// ProcessorResult contains the result of running a processor.
type ProcessorResult struct {
	Success bool   `json:"success"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
}

// JSPreProcessor wraps a JavaScript pre-processor registered by a plugin.
// Pre-processors transform source code before parsing.
type JSPreProcessor struct {
	Index    int
	Priority int
	runtime  *NodeJSRuntime
}

// NewJSPreProcessor creates a new JSPreProcessor wrapper.
func NewJSPreProcessor(runtime *NodeJSRuntime, index, priority int) *JSPreProcessor {
	return &JSPreProcessor{
		Index:    index,
		Priority: priority,
		runtime:  runtime,
	}
}

// Process runs the pre-processor on the input source code.
// It sends the source to Node.js for processing and returns the transformed result.
func (p *JSPreProcessor) Process(input string, options map[string]any) (string, error) {
	if p.runtime == nil {
		return "", fmt.Errorf("Node.js runtime not initialized")
	}

	resp, err := p.runtime.SendCommand(Command{
		Cmd: "runPreProcessor",
		Data: map[string]any{
			"processorIndex": p.Index,
			"input":          input,
			"options":        options,
		},
	})
	if err != nil {
		return "", fmt.Errorf("pre-processor call failed: %w", err)
	}

	if !resp.Success {
		return "", fmt.Errorf("pre-processor error: %s", resp.Error)
	}

	// Parse the result
	resultMap, ok := resp.Result.(map[string]any)
	if !ok {
		// If result is directly a string, return it
		if output, ok := resp.Result.(string); ok {
			return output, nil
		}
		return "", fmt.Errorf("unexpected result type: %T", resp.Result)
	}

	if output, ok := resultMap["output"].(string); ok {
		return output, nil
	}

	return "", fmt.Errorf("no output in processor result")
}

// JSPostProcessor wraps a JavaScript post-processor registered by a plugin.
// Post-processors transform CSS output after compilation.
type JSPostProcessor struct {
	Index    int
	Priority int
	runtime  *NodeJSRuntime
}

// NewJSPostProcessor creates a new JSPostProcessor wrapper.
func NewJSPostProcessor(runtime *NodeJSRuntime, index, priority int) *JSPostProcessor {
	return &JSPostProcessor{
		Index:    index,
		Priority: priority,
		runtime:  runtime,
	}
}

// Process runs the post-processor on the CSS output.
// It sends the CSS to Node.js for processing and returns the transformed result.
func (p *JSPostProcessor) Process(css string, options map[string]any) (string, error) {
	if p.runtime == nil {
		return "", fmt.Errorf("Node.js runtime not initialized")
	}

	resp, err := p.runtime.SendCommand(Command{
		Cmd: "runPostProcessor",
		Data: map[string]any{
			"processorIndex": p.Index,
			"input":          css,
			"options":        options,
		},
	})
	if err != nil {
		return "", fmt.Errorf("post-processor call failed: %w", err)
	}

	if !resp.Success {
		return "", fmt.Errorf("post-processor error: %s", resp.Error)
	}

	// Parse the result
	resultMap, ok := resp.Result.(map[string]any)
	if !ok {
		// If result is directly a string, return it
		if output, ok := resp.Result.(string); ok {
			return output, nil
		}
		return "", fmt.Errorf("unexpected result type: %T", resp.Result)
	}

	if output, ok := resultMap["output"].(string); ok {
		return output, nil
	}

	return "", fmt.Errorf("no output in processor result")
}

// ProcessorManager manages JavaScript pre/post processors for a plugin loader.
type ProcessorManager struct {
	runtime        *NodeJSRuntime
	preProcessors  []*JSPreProcessor
	postProcessors []*JSPostProcessor
}

// NewProcessorManager creates a new processor manager.
func NewProcessorManager(runtime *NodeJSRuntime) *ProcessorManager {
	return &ProcessorManager{
		runtime:        runtime,
		preProcessors:  make([]*JSPreProcessor, 0),
		postProcessors: make([]*JSPostProcessor, 0),
	}
}

// RefreshProcessors fetches the current list of registered processors from Node.js.
func (pm *ProcessorManager) RefreshProcessors() error {
	// Get pre-processors
	preResp, err := pm.runtime.SendCommand(Command{
		Cmd: "getPreProcessors",
	})
	if err != nil {
		return fmt.Errorf("failed to get pre-processors: %w", err)
	}

	if preResp.Success {
		pm.preProcessors = make([]*JSPreProcessor, 0)
		if processors, ok := preResp.Result.([]any); ok {
			for i, p := range processors {
				priority := 1000 // default priority
				if pMap, ok := p.(map[string]any); ok {
					if pri, ok := pMap["priority"].(float64); ok {
						priority = int(pri)
					}
				}
				pm.preProcessors = append(pm.preProcessors, NewJSPreProcessor(pm.runtime, i, priority))
			}
		}
	}

	// Get post-processors
	postResp, err := pm.runtime.SendCommand(Command{
		Cmd: "getPostProcessors",
	})
	if err != nil {
		return fmt.Errorf("failed to get post-processors: %w", err)
	}

	if postResp.Success {
		pm.postProcessors = make([]*JSPostProcessor, 0)
		if processors, ok := postResp.Result.([]any); ok {
			for i, p := range processors {
				priority := 1000 // default priority
				if pMap, ok := p.(map[string]any); ok {
					if pri, ok := pMap["priority"].(float64); ok {
						priority = int(pri)
					}
				}
				pm.postProcessors = append(pm.postProcessors, NewJSPostProcessor(pm.runtime, i, priority))
			}
		}
	}

	return nil
}

// GetPreProcessors returns all registered pre-processors.
func (pm *ProcessorManager) GetPreProcessors() []*JSPreProcessor {
	return pm.preProcessors
}

// GetPostProcessors returns all registered post-processors.
func (pm *ProcessorManager) GetPostProcessors() []*JSPostProcessor {
	return pm.postProcessors
}

// RunPreProcessors runs all pre-processors on the input source code.
// Processors are run in order of their priority (lower priority runs first).
func (pm *ProcessorManager) RunPreProcessors(input string, options map[string]any) (string, error) {
	result := input
	for _, proc := range pm.preProcessors {
		var err error
		result, err = proc.Process(result, options)
		if err != nil {
			return "", fmt.Errorf("pre-processor %d failed: %w", proc.Index, err)
		}
	}
	return result, nil
}

// RunPostProcessors runs all post-processors on the CSS output.
// Processors are run in order of their priority (lower priority runs first).
func (pm *ProcessorManager) RunPostProcessors(css string, options map[string]any) (string, error) {
	result := css
	for _, proc := range pm.postProcessors {
		var err error
		result, err = proc.Process(result, options)
		if err != nil {
			return "", fmt.Errorf("post-processor %d failed: %w", proc.Index, err)
		}
	}
	return result, nil
}

// PreProcessorCount returns the number of registered pre-processors.
func (pm *ProcessorManager) PreProcessorCount() int {
	return len(pm.preProcessors)
}

// PostProcessorCount returns the number of registered post-processors.
func (pm *ProcessorManager) PostProcessorCount() int {
	return len(pm.postProcessors)
}
