package runtime

import (
	"encoding/base64"
	"fmt"
)

// VisitorInfo contains metadata about a registered JavaScript visitor.
type VisitorInfo struct {
	Index           int  `json:"index"`
	IsPreEvalVisitor bool `json:"isPreEvalVisitor"`
	IsReplacing     bool `json:"isReplacing"`
}

// VisitorResult contains the result of running a visitor.
type VisitorResult struct {
	Success      bool                     `json:"success"`
	Replacements []VisitorReplacementSet  `json:"replacements,omitempty"`
	VisitorCount int                      `json:"visitorCount,omitempty"`
	ResultType   string                   `json:"resultType,omitempty"`
	Message      string                   `json:"message,omitempty"`
}

// VisitorReplacementSet contains replacements from a single visitor.
type VisitorReplacementSet struct {
	VisitorIndex int                 `json:"visitorIndex"`
	Replacements []NodeReplacement   `json:"replacements"`
}

// NodeReplacement represents a single node replacement in the AST.
type NodeReplacement struct {
	ParentIndex  int         `json:"parentIndex"`
	ChildIndex   int         `json:"childIndex"`
	Replacement  interface{} `json:"replacement"`
}

// JSVisitor wraps a JavaScript visitor registered by a plugin.
// It provides methods to invoke the visitor on Go AST nodes.
type JSVisitor struct {
	Index           int
	IsPreEvalVisitor bool
	IsReplacing     bool
	runtime         *NodeJSRuntime
}

// NewJSVisitor creates a new JSVisitor wrapper.
func NewJSVisitor(runtime *NodeJSRuntime, info VisitorInfo) *JSVisitor {
	return &JSVisitor{
		Index:            info.Index,
		IsPreEvalVisitor: info.IsPreEvalVisitor,
		IsReplacing:      info.IsReplacing,
		runtime:          runtime,
	}
}

// Visit runs the visitor on a Go AST node.
// It serializes the AST to a buffer, sends it to Node.js, runs the visitor,
// and returns any modifications.
func (v *JSVisitor) Visit(node interface{}) (*VisitorResult, error) {
	if v.runtime == nil {
		return nil, fmt.Errorf("Node.js runtime not initialized")
	}

	// Flatten the AST to buffer format
	flat, err := FlattenAST(node)
	if err != nil {
		return nil, fmt.Errorf("failed to flatten AST: %w", err)
	}

	// Write to shared memory
	shm, err := v.runtime.WriteASTBuffer(flat)
	if err != nil {
		return nil, fmt.Errorf("failed to write AST buffer: %w", err)
	}
	defer v.runtime.DestroySharedMemory(shm)

	// Attach buffer in Node.js
	if err := v.runtime.AttachBuffer(shm); err != nil {
		return nil, fmt.Errorf("failed to attach buffer: %w", err)
	}
	defer v.runtime.DetachBuffer(shm.Key())

	// Run the visitor
	resp, err := v.runtime.SendCommand(Command{
		Cmd: "runVisitor",
		Data: map[string]interface{}{
			"bufferKey":    shm.Key(),
			"visitorIndex": v.Index,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to run visitor: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("visitor error: %s", resp.Error)
	}

	// Parse result
	result := &VisitorResult{Success: true}
	if resultMap, ok := resp.Result.(map[string]interface{}); ok {
		if replacements, ok := resultMap["replacements"].([]interface{}); ok {
			result.Replacements = parseReplacements(replacements)
		}
		if resultType, ok := resultMap["resultType"].(string); ok {
			result.ResultType = resultType
		}
	}

	return result, nil
}

// VisitorManager manages JavaScript visitors for a plugin loader.
type VisitorManager struct {
	runtime  *NodeJSRuntime
	visitors []*JSVisitor
}

// NewVisitorManager creates a new visitor manager.
func NewVisitorManager(runtime *NodeJSRuntime) *VisitorManager {
	return &VisitorManager{
		runtime:  runtime,
		visitors: make([]*JSVisitor, 0),
	}
}

// RefreshVisitors fetches the current list of registered visitors from Node.js.
func (vm *VisitorManager) RefreshVisitors() error {
	resp, err := vm.runtime.SendCommand(Command{
		Cmd: "getVisitors",
	})
	if err != nil {
		return fmt.Errorf("failed to get visitors: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("get visitors failed: %s", resp.Error)
	}

	// Parse visitor info
	vm.visitors = make([]*JSVisitor, 0)
	if visitors, ok := resp.Result.([]interface{}); ok {
		for _, v := range visitors {
			if vMap, ok := v.(map[string]interface{}); ok {
				info := VisitorInfo{}
				if idx, ok := vMap["index"].(float64); ok {
					info.Index = int(idx)
				}
				if isPreEval, ok := vMap["isPreEvalVisitor"].(bool); ok {
					info.IsPreEvalVisitor = isPreEval
				}
				if isReplacing, ok := vMap["isReplacing"].(bool); ok {
					info.IsReplacing = isReplacing
				}
				vm.visitors = append(vm.visitors, NewJSVisitor(vm.runtime, info))
			}
		}
	}

	return nil
}

// GetPreEvalVisitors returns all pre-evaluation visitors.
func (vm *VisitorManager) GetPreEvalVisitors() []*JSVisitor {
	result := make([]*JSVisitor, 0)
	for _, v := range vm.visitors {
		if v.IsPreEvalVisitor {
			result = append(result, v)
		}
	}
	return result
}

// GetPostEvalVisitors returns all post-evaluation visitors.
func (vm *VisitorManager) GetPostEvalVisitors() []*JSVisitor {
	result := make([]*JSVisitor, 0)
	for _, v := range vm.visitors {
		if !v.IsPreEvalVisitor {
			result = append(result, v)
		}
	}
	return result
}

// RunPreEvalVisitors runs all pre-evaluation visitors on an AST.
func (vm *VisitorManager) RunPreEvalVisitors(node interface{}) (*VisitorResult, error) {
	if vm.runtime == nil {
		return nil, fmt.Errorf("Node.js runtime not initialized")
	}

	// Flatten the AST
	flat, err := FlattenAST(node)
	if err != nil {
		return nil, fmt.Errorf("failed to flatten AST: %w", err)
	}

	// Write to shared memory
	shm, err := vm.runtime.WriteASTBuffer(flat)
	if err != nil {
		return nil, fmt.Errorf("failed to write AST buffer: %w", err)
	}
	defer vm.runtime.DestroySharedMemory(shm)

	// Attach buffer
	if err := vm.runtime.AttachBuffer(shm); err != nil {
		return nil, fmt.Errorf("failed to attach buffer: %w", err)
	}
	defer vm.runtime.DetachBuffer(shm.Key())

	// Run pre-eval visitors
	resp, err := vm.runtime.SendCommand(Command{
		Cmd: "runPreEvalVisitors",
		Data: map[string]interface{}{
			"bufferKey": shm.Key(),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to run pre-eval visitors: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("pre-eval visitors error: %s", resp.Error)
	}

	return parseVisitorResult(resp.Result)
}

// RunPostEvalVisitors runs all post-evaluation visitors on an AST.
func (vm *VisitorManager) RunPostEvalVisitors(node interface{}) (*VisitorResult, error) {
	if vm.runtime == nil {
		return nil, fmt.Errorf("Node.js runtime not initialized")
	}

	// Flatten the AST
	flat, err := FlattenAST(node)
	if err != nil {
		return nil, fmt.Errorf("failed to flatten AST: %w", err)
	}

	// Write to shared memory
	shm, err := vm.runtime.WriteASTBuffer(flat)
	if err != nil {
		return nil, fmt.Errorf("failed to write AST buffer: %w", err)
	}
	defer vm.runtime.DestroySharedMemory(shm)

	// Attach buffer
	if err := vm.runtime.AttachBuffer(shm); err != nil {
		return nil, fmt.Errorf("failed to attach buffer: %w", err)
	}
	defer vm.runtime.DetachBuffer(shm.Key())

	// Run post-eval visitors
	resp, err := vm.runtime.SendCommand(Command{
		Cmd: "runPostEvalVisitors",
		Data: map[string]interface{}{
			"bufferKey": shm.Key(),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to run post-eval visitors: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("post-eval visitors error: %s", resp.Error)
	}

	return parseVisitorResult(resp.Result)
}

// parseVisitorResult parses a visitor result from the JSON response.
func parseVisitorResult(result interface{}) (*VisitorResult, error) {
	vr := &VisitorResult{Success: true}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return vr, nil
	}

	if count, ok := resultMap["visitorCount"].(float64); ok {
		vr.VisitorCount = int(count)
	}

	if replacements, ok := resultMap["replacements"].([]interface{}); ok {
		vr.Replacements = parseReplacements(replacements)
	}

	return vr, nil
}

// parseReplacements parses replacement data from JSON.
func parseReplacements(data []interface{}) []VisitorReplacementSet {
	result := make([]VisitorReplacementSet, 0, len(data))

	for _, item := range data {
		if setMap, ok := item.(map[string]interface{}); ok {
			set := VisitorReplacementSet{}

			if idx, ok := setMap["visitorIndex"].(float64); ok {
				set.VisitorIndex = int(idx)
			}

			if replacements, ok := setMap["replacements"].([]interface{}); ok {
				for _, r := range replacements {
					if rMap, ok := r.(map[string]interface{}); ok {
						replacement := NodeReplacement{}
						if pi, ok := rMap["parentIndex"].(float64); ok {
							replacement.ParentIndex = int(pi)
						}
						if ci, ok := rMap["childIndex"].(float64); ok {
							replacement.ChildIndex = int(ci)
						}
						replacement.Replacement = rMap["replacement"]
						set.Replacements = append(set.Replacements, replacement)
					}
				}
			}

			result = append(result, set)
		}
	}

	return result
}

// SerializeNodeResult contains the result of serializing a node.
type SerializeNodeResult struct {
	Buffer []byte
	JSON   string
	Size   int
}

// SerializeNode serializes a JavaScript node to buffer format.
func (vm *VisitorManager) SerializeNode(node interface{}) (*SerializeNodeResult, error) {
	resp, err := vm.runtime.SendCommand(Command{
		Cmd: "serializeNode",
		Data: map[string]interface{}{
			"node": node,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to serialize node: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("serialize error: %s", resp.Error)
	}

	result := &SerializeNodeResult{}
	if resultMap, ok := resp.Result.(map[string]interface{}); ok {
		if bufStr, ok := resultMap["buffer"].(string); ok {
			result.Buffer, _ = base64.StdEncoding.DecodeString(bufStr)
		}
		if jsonStr, ok := resultMap["json"].(string); ok {
			result.JSON = jsonStr
		}
		if size, ok := resultMap["size"].(float64); ok {
			result.Size = int(size)
		}
	}

	return result, nil
}

// ParseASTBufferResult contains the result of parsing an AST buffer.
type ParseASTBufferResult struct {
	Version         uint32
	NodeCount       uint32
	RootIndex       uint32
	StringTableSize int
	TypeTableSize   int
}

// ParseASTBuffer parses an AST buffer and returns metadata.
func (vm *VisitorManager) ParseASTBuffer(shm *SharedMemory) (*ParseASTBufferResult, error) {
	if err := vm.runtime.AttachBuffer(shm); err != nil {
		return nil, fmt.Errorf("failed to attach buffer: %w", err)
	}
	defer vm.runtime.DetachBuffer(shm.Key())

	resp, err := vm.runtime.SendCommand(Command{
		Cmd: "parseASTBuffer",
		Data: map[string]interface{}{
			"bufferKey": shm.Key(),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse AST buffer: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("parse error: %s", resp.Error)
	}

	result := &ParseASTBufferResult{}
	if resultMap, ok := resp.Result.(map[string]interface{}); ok {
		if v, ok := resultMap["version"].(float64); ok {
			result.Version = uint32(v)
		}
		if nc, ok := resultMap["nodeCount"].(float64); ok {
			result.NodeCount = uint32(nc)
		}
		if ri, ok := resultMap["rootIndex"].(float64); ok {
			result.RootIndex = uint32(ri)
		}
		if sts, ok := resultMap["stringTableSize"].(float64); ok {
			result.StringTableSize = int(sts)
		}
		if tts, ok := resultMap["typeTableSize"].(float64); ok {
			result.TypeTableSize = int(tts)
		}
	}

	return result, nil
}
