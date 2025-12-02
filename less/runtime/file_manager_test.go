//go:build !windows

package runtime

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileManagerCollection_RefreshFileManagers(t *testing.T) {
	pluginHostPath := findPluginHostJS()
	if pluginHostPath == "" {
		t.Skip("plugin-host.js not found")
	}

	rt, err := NewNodeJSRuntime(WithPluginHostPath(pluginHostPath))
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}

	if err := rt.Start(); err != nil {
		t.Fatalf("Failed to start runtime: %v", err)
	}
	defer rt.Stop()

	fmc := NewFileManagerCollection(rt)

	// Initially should have no file managers
	if err := fmc.RefreshFileManagers(); err != nil {
		t.Fatalf("RefreshFileManagers failed: %v", err)
	}

	if fmc.FileManagerCount() != 0 {
		t.Errorf("Expected 0 file managers, got %d", fmc.FileManagerCount())
	}
}

func TestFileManagerCollection_GetFileManagers(t *testing.T) {
	pluginHostPath := findPluginHostJS()
	if pluginHostPath == "" {
		t.Skip("plugin-host.js not found")
	}

	rt, err := NewNodeJSRuntime(WithPluginHostPath(pluginHostPath))
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}

	if err := rt.Start(); err != nil {
		t.Fatalf("Failed to start runtime: %v", err)
	}
	defer rt.Stop()

	fmc := NewFileManagerCollection(rt)

	// GetFileManagers should work even without refresh
	managers := fmc.GetFileManagers()
	if managers == nil {
		t.Error("GetFileManagers returned nil")
	}
}

func TestJSFileManager_NilRuntime(t *testing.T) {
	fm := NewJSFileManager(nil, 0)

	_, err := fm.Supports("test.less", "/current", nil)
	if err == nil {
		t.Error("Expected error for nil runtime in Supports")
	}

	_, err = fm.LoadFile("test.less", "/current", nil)
	if err == nil {
		t.Error("Expected error for nil runtime in LoadFile")
	}
}

func TestFileManagerCollection_WithPlugin(t *testing.T) {
	pluginHostPath := findPluginHostJS()
	if pluginHostPath == "" {
		t.Skip("plugin-host.js not found")
	}

	rt, err := NewNodeJSRuntime(WithPluginHostPath(pluginHostPath))
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}

	if err := rt.Start(); err != nil {
		t.Fatalf("Failed to start runtime: %v", err)
	}
	defer rt.Stop()

	// Create a temporary test plugin that adds a file manager
	tmpDir := t.TempDir()
	pluginPath := filepath.Join(tmpDir, "test-file-manager-plugin.js")
	pluginCode := `
module.exports = {
    install(less, pluginManager, functions) {
        // Add a file manager that handles custom:// URLs
        pluginManager.addFileManager({
            supports(filename, currentDirectory, options, environment) {
                return filename.startsWith('custom://');
            },
            loadFile(filename, currentDirectory, options, environment) {
                // Remove the custom:// prefix and return mock content
                const name = filename.replace('custom://', '');
                return {
                    filename: name,
                    contents: '/* loaded from custom:// */\n.custom { color: blue; }'
                };
            }
        });
    }
};
`
	if err := os.WriteFile(pluginPath, []byte(pluginCode), 0644); err != nil {
		t.Fatalf("Failed to write test plugin: %v", err)
	}

	// Load the plugin
	loader := NewJSPluginLoader(rt)
	result := loader.LoadPluginSync(pluginPath, tmpDir, nil, nil, nil)
	if err, ok := result.(error); ok {
		t.Fatalf("Failed to load plugin: %v", err)
	}

	// Create file manager collection and refresh
	fmc := NewFileManagerCollection(rt)
	if err := fmc.RefreshFileManagers(); err != nil {
		t.Fatalf("RefreshFileManagers failed: %v", err)
	}

	// Verify file manager was registered
	if fmc.FileManagerCount() != 1 {
		t.Errorf("Expected 1 file manager, got %d", fmc.FileManagerCount())
	}

	// Test supports check
	fm := fmc.GetFileManagers()[0]
	supports, err := fm.Supports("custom://test.less", "/current", nil)
	if err != nil {
		t.Fatalf("Supports check failed: %v", err)
	}
	if !supports {
		t.Error("Expected file manager to support custom:// URL")
	}

	// Test supports check for non-custom URL
	supports, err = fm.Supports("regular.less", "/current", nil)
	if err != nil {
		t.Fatalf("Supports check failed: %v", err)
	}
	if supports {
		t.Error("Expected file manager to not support regular URL")
	}

	// Test loadFile
	file, err := fm.LoadFile("custom://test.less", "/current", nil)
	if err != nil {
		t.Fatalf("LoadFile failed: %v", err)
	}

	if file.Filename != "test.less" {
		t.Errorf("Expected filename 'test.less', got %q", file.Filename)
	}

	expectedContents := "/* loaded from custom:// */\n.custom { color: blue; }"
	if file.Contents != expectedContents {
		t.Errorf("Contents mismatch:\nExpected: %q\nGot: %q", expectedContents, file.Contents)
	}
}

func TestFileManagerCollection_FindSupportingManager(t *testing.T) {
	pluginHostPath := findPluginHostJS()
	if pluginHostPath == "" {
		t.Skip("plugin-host.js not found")
	}

	rt, err := NewNodeJSRuntime(WithPluginHostPath(pluginHostPath))
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}

	if err := rt.Start(); err != nil {
		t.Fatalf("Failed to start runtime: %v", err)
	}
	defer rt.Stop()

	// Create a temporary test plugin with multiple file managers
	tmpDir := t.TempDir()
	pluginPath := filepath.Join(tmpDir, "test-multi-file-manager-plugin.js")
	pluginCode := `
module.exports = {
    install(less, pluginManager, functions) {
        // Add file manager for http:// URLs
        pluginManager.addFileManager({
            supports(filename, currentDirectory, options, environment) {
                return filename.startsWith('http://');
            },
            loadFile(filename, currentDirectory, options, environment) {
                return { filename, contents: '/* http content */' };
            }
        });

        // Add file manager for https:// URLs
        pluginManager.addFileManager({
            supports(filename, currentDirectory, options, environment) {
                return filename.startsWith('https://');
            },
            loadFile(filename, currentDirectory, options, environment) {
                return { filename, contents: '/* https content */' };
            }
        });
    }
};
`
	if err := os.WriteFile(pluginPath, []byte(pluginCode), 0644); err != nil {
		t.Fatalf("Failed to write test plugin: %v", err)
	}

	// Load the plugin
	loader := NewJSPluginLoader(rt)
	result := loader.LoadPluginSync(pluginPath, tmpDir, nil, nil, nil)
	if err, ok := result.(error); ok {
		t.Fatalf("Failed to load plugin: %v", err)
	}

	// Create file manager collection and refresh
	fmc := NewFileManagerCollection(rt)
	if err := fmc.RefreshFileManagers(); err != nil {
		t.Fatalf("RefreshFileManagers failed: %v", err)
	}

	// Verify file managers were registered
	if fmc.FileManagerCount() != 2 {
		t.Errorf("Expected 2 file managers, got %d", fmc.FileManagerCount())
	}

	// Test FindSupportingManager for http://
	fm := fmc.FindSupportingManager("http://example.com/style.less", "/current", nil)
	if fm == nil {
		t.Error("Expected to find a supporting manager for http://")
	}

	// Test FindSupportingManager for https://
	fm = fmc.FindSupportingManager("https://example.com/style.less", "/current", nil)
	if fm == nil {
		t.Error("Expected to find a supporting manager for https://")
	}

	// Test FindSupportingManager for unsupported URL
	fm = fmc.FindSupportingManager("ftp://example.com/style.less", "/current", nil)
	if fm != nil {
		t.Error("Expected no supporting manager for ftp://")
	}
}

func TestFileManagerCollection_LoadFile(t *testing.T) {
	pluginHostPath := findPluginHostJS()
	if pluginHostPath == "" {
		t.Skip("plugin-host.js not found")
	}

	rt, err := NewNodeJSRuntime(WithPluginHostPath(pluginHostPath))
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}

	if err := rt.Start(); err != nil {
		t.Fatalf("Failed to start runtime: %v", err)
	}
	defer rt.Stop()

	// Create a temporary test plugin
	tmpDir := t.TempDir()
	pluginPath := filepath.Join(tmpDir, "test-load-plugin.js")
	pluginCode := `
module.exports = {
    install(less, pluginManager, functions) {
        pluginManager.addFileManager({
            supports(filename, currentDirectory, options, environment) {
                return filename.startsWith('virtual://');
            },
            loadFile(filename, currentDirectory, options, environment) {
                const name = filename.replace('virtual://', '');
                return {
                    filename: name,
                    contents: '.virtual { content: "' + name + '"; }'
                };
            }
        });
    }
};
`
	if err := os.WriteFile(pluginPath, []byte(pluginCode), 0644); err != nil {
		t.Fatalf("Failed to write test plugin: %v", err)
	}

	// Load the plugin
	loader := NewJSPluginLoader(rt)
	result := loader.LoadPluginSync(pluginPath, tmpDir, nil, nil, nil)
	if err, ok := result.(error); ok {
		t.Fatalf("Failed to load plugin: %v", err)
	}

	// Create file manager collection and refresh
	fmc := NewFileManagerCollection(rt)
	if err := fmc.RefreshFileManagers(); err != nil {
		t.Fatalf("RefreshFileManagers failed: %v", err)
	}

	// Test loading through the collection
	file, err := fmc.LoadFile("virtual://myfile.less", "/current", nil)
	if err != nil {
		t.Fatalf("LoadFile failed: %v", err)
	}

	if file.Filename != "myfile.less" {
		t.Errorf("Expected filename 'myfile.less', got %q", file.Filename)
	}

	expectedContents := `.virtual { content: "myfile.less"; }`
	if file.Contents != expectedContents {
		t.Errorf("Contents mismatch:\nExpected: %q\nGot: %q", expectedContents, file.Contents)
	}

	// Test loading unsupported file (should fail)
	_, err = fmc.LoadFile("unsupported.less", "/current", nil)
	if err == nil {
		t.Error("Expected error for unsupported file")
	}
}

func TestFileManagerCollection_LoadFile_NoManagers(t *testing.T) {
	pluginHostPath := findPluginHostJS()
	if pluginHostPath == "" {
		t.Skip("plugin-host.js not found")
	}

	rt, err := NewNodeJSRuntime(WithPluginHostPath(pluginHostPath))
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}

	if err := rt.Start(); err != nil {
		t.Fatalf("Failed to start runtime: %v", err)
	}
	defer rt.Stop()

	fmc := NewFileManagerCollection(rt)
	if err := fmc.RefreshFileManagers(); err != nil {
		t.Fatalf("RefreshFileManagers failed: %v", err)
	}

	// With no file managers, LoadFile should fail
	_, err = fmc.LoadFile("test.less", "/current", nil)
	if err == nil {
		t.Error("Expected error when no file managers are registered")
	}
}
