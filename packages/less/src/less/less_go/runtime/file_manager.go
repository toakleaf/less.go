package runtime

import (
	"fmt"
)

// FileManagerInfo contains metadata about a registered JavaScript file manager.
type FileManagerInfo struct {
	Index int `json:"index"`
}

// LoadedFile represents the result of loading a file.
type LoadedFile struct {
	Filename string `json:"filename"`
	Contents string `json:"contents"`
}

// JSFileManager wraps a JavaScript file manager registered by a plugin.
// File managers provide custom import resolution logic for LESS files.
type JSFileManager struct {
	Index   int
	runtime *NodeJSRuntime
}

// NewJSFileManager creates a new JSFileManager wrapper.
func NewJSFileManager(runtime *NodeJSRuntime, index int) *JSFileManager {
	return &JSFileManager{
		Index:   index,
		runtime: runtime,
	}
}

// Supports checks if this file manager can handle the given file.
// It sends a request to Node.js to call the file manager's supports() method.
func (fm *JSFileManager) Supports(filename, currentDirectory string, options map[string]any) (bool, error) {
	if fm.runtime == nil {
		return false, fmt.Errorf("Node.js runtime not initialized")
	}

	resp, err := fm.runtime.SendCommand(Command{
		Cmd: "fileManagerSupports",
		Data: map[string]any{
			"managerIndex":     fm.Index,
			"filename":         filename,
			"currentDirectory": currentDirectory,
			"options":          options,
		},
	})
	if err != nil {
		return false, fmt.Errorf("file manager supports check failed: %w", err)
	}

	if !resp.Success {
		return false, fmt.Errorf("file manager error: %s", resp.Error)
	}

	// Parse the result
	resultMap, ok := resp.Result.(map[string]any)
	if !ok {
		return false, fmt.Errorf("unexpected result type: %T", resp.Result)
	}

	supports, ok := resultMap["supports"].(bool)
	if !ok {
		return false, nil
	}

	return supports, nil
}

// LoadFile loads a file using this file manager.
// It sends a request to Node.js to call the file manager's loadFile() method.
func (fm *JSFileManager) LoadFile(filename, currentDirectory string, options map[string]any) (*LoadedFile, error) {
	if fm.runtime == nil {
		return nil, fmt.Errorf("Node.js runtime not initialized")
	}

	resp, err := fm.runtime.SendCommand(Command{
		Cmd: "fileManagerLoad",
		Data: map[string]any{
			"managerIndex":     fm.Index,
			"filename":         filename,
			"currentDirectory": currentDirectory,
			"options":          options,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("file load failed: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("file manager error: %s", resp.Error)
	}

	// Parse the result
	resultMap, ok := resp.Result.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", resp.Result)
	}

	loadedFile := &LoadedFile{}
	if name, ok := resultMap["filename"].(string); ok {
		loadedFile.Filename = name
	}
	if contents, ok := resultMap["contents"].(string); ok {
		loadedFile.Contents = contents
	}

	return loadedFile, nil
}

// FileManagerCollection manages JavaScript file managers for a plugin loader.
type FileManagerCollection struct {
	runtime      *NodeJSRuntime
	fileManagers []*JSFileManager
}

// NewFileManagerCollection creates a new file manager collection.
func NewFileManagerCollection(runtime *NodeJSRuntime) *FileManagerCollection {
	return &FileManagerCollection{
		runtime:      runtime,
		fileManagers: make([]*JSFileManager, 0),
	}
}

// RefreshFileManagers fetches the current list of registered file managers from Node.js.
func (fmc *FileManagerCollection) RefreshFileManagers() error {
	resp, err := fmc.runtime.SendCommand(Command{
		Cmd: "getFileManagers",
	})
	if err != nil {
		return fmt.Errorf("failed to get file managers: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("get file managers failed: %s", resp.Error)
	}

	fmc.fileManagers = make([]*JSFileManager, 0)
	if managers, ok := resp.Result.([]any); ok {
		for _, m := range managers {
			if mMap, ok := m.(map[string]any); ok {
				index := 0
				if idx, ok := mMap["index"].(float64); ok {
					index = int(idx)
				}
				fmc.fileManagers = append(fmc.fileManagers, NewJSFileManager(fmc.runtime, index))
			}
		}
	}

	return nil
}

// GetFileManagers returns all registered file managers.
func (fmc *FileManagerCollection) GetFileManagers() []*JSFileManager {
	return fmc.fileManagers
}

// FileManagerCount returns the number of registered file managers.
func (fmc *FileManagerCollection) FileManagerCount() int {
	return len(fmc.fileManagers)
}

// FindSupportingManager finds the first file manager that supports the given file.
// Returns nil if no file manager supports the file.
func (fmc *FileManagerCollection) FindSupportingManager(filename, currentDirectory string, options map[string]any) *JSFileManager {
	for _, fm := range fmc.fileManagers {
		supports, err := fm.Supports(filename, currentDirectory, options)
		if err == nil && supports {
			return fm
		}
	}
	return nil
}

// LoadFile tries to load a file using the registered file managers.
// It tries each file manager in order and returns the first successful result.
// Returns an error if no file manager can load the file.
func (fmc *FileManagerCollection) LoadFile(filename, currentDirectory string, options map[string]any) (*LoadedFile, error) {
	for _, fm := range fmc.fileManagers {
		supports, err := fm.Supports(filename, currentDirectory, options)
		if err != nil || !supports {
			continue
		}

		file, err := fm.LoadFile(filename, currentDirectory, options)
		if err == nil {
			return file, nil
		}
	}

	return nil, fmt.Errorf("no file manager could load file: %s", filename)
}
