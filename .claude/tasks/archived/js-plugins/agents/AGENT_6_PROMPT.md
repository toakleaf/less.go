# AGENT 6: Processors & File Managers

**Status**: ⏸️ Blocked - Wait for Agent 2 (Plugin Loader)
**Dependencies**: Agent 2 (need plugin loading to work)
**Estimated Time**: 3-4 days
**Can work in parallel with**: Agents 4, 5

---

You are implementing pre/post processors and file managers for less.go plugins.

## Your Mission

Implement Phase 8 (Pre/Post Processors) and Phase 9 (File Managers) from the strategy document.

## Prerequisites

✅ Verify Agent 2 has completed:
- Plugin loading works
- Plugins can register items via pluginManager
- IPC command/response works

Check: `go test ./runtime -run TestPluginLoader`

## Required Reading

BEFORE starting, read:
1. IMPLEMENTATION_STRATEGY.md - Focus on Phase 8 and Phase 9
2. `packages/less/src/less/plugin-manager.js` - JavaScript implementation
3. Less.js file manager implementation

## Your Tasks

### Phase 8: Pre/Post Processors

#### 1. Create Processor Types (Go Side)

```go
// runtime/processor.go

type JSPreProcessor struct {
    processorID string
    priority    int
    runtime     *NodeJSRuntime
}

func (jpp *JSPreProcessor) Process(input string, options map[string]interface{}) (string, error) {
    resp, err := jpp.runtime.SendCommand(Command{
        Cmd:         "runPreProcessor",
        ProcessorID: jpp.processorID,
        Input:       input,
        Options:     options,
    })

    if err != nil {
        return "", err
    }

    return resp.Output, nil
}

type JSPostProcessor struct {
    processorID string
    priority    int
    runtime     *NodeJSRuntime
}

func (jpo *JSPostProcessor) Process(css string, options map[string]interface{}) (string, error) {
    // Similar to pre-processor
}
```

#### 2. Add to PluginManager

```go
// runtime/plugin_manager.go

type PluginManager struct {
    // ... existing fields
    preProcessors  []*JSPreProcessor
    postProcessors []*JSPostProcessor
}

func (pm *PluginManager) AddPreProcessor(proc *JSPreProcessor) {
    // Insert by priority
    pm.preProcessors = insertByPriority(pm.preProcessors, proc)
}

func (pm *PluginManager) AddPostProcessor(proc *JSPostProcessor) {
    pm.postProcessors = insertByPriority(pm.postProcessors, proc)
}
```

#### 3. Implement in Node.js

```javascript
// plugin-host.js

const registeredPreProcessors = [];
const registeredPostProcessors = [];

const pluginManager = {
    // ... existing methods

    addPreProcessor(processor, priority = 1000) {
        const id = `preproc_${Date.now()}`;
        registeredPreProcessors.push({
            id,
            processor,
            priority
        });
        registeredPreProcessors.sort((a, b) => a.priority - b.priority);
        return { id, priority };
    },

    addPostProcessor(processor, priority = 1000) {
        const id = `postproc_${Date.now()}`;
        registeredPostProcessors.push({
            id,
            processor,
            priority
        });
        registeredPostProcessors.sort((a, b) => a.priority - b.priority);
        return { id, priority };
    }
};

function handleRunPreProcessor(cmd) {
    const { processorID, input, options } = cmd;

    const proc = registeredPreProcessors.find(p => p.id === processorID);
    if (!proc) {
        return { success: false, error: 'Processor not found' };
    }

    try {
        const output = proc.processor.process(input, options);
        return { success: true, output };
    } catch (error) {
        return { success: false, error: error.message };
    }
}
```

#### 4. Integrate in Parse/Compile Pipeline

```go
// In parse.go or similar

func Parse(input string, options *Options) (Node, error) {
    // Run pre-processors
    processedInput := input
    for _, proc := range options.PluginManager.GetPreProcessors() {
        var err error
        processedInput, err = proc.Process(processedInput, options.ToMap())
        if err != nil {
            return nil, err
        }
    }

    // Parse
    ast, err := parser.Parse(processedInput, options)
    if err != nil {
        return nil, err
    }

    return ast, nil
}

// In output generation

func ToCSS(root Node, options *Options) (string, error) {
    // Generate CSS
    css := root.ToCSS(options)

    // Run post-processors
    processedCSS := css
    for _, proc := range options.PluginManager.GetPostProcessors() {
        var err error
        processedCSS, err = proc.Process(processedCSS, options.ToMap())
        if err != nil {
            return "", err
        }
    }

    return processedCSS, nil
}
```

### Phase 9: File Managers

#### 1. Create JSFileManager (Go Side)

```go
// runtime/file_manager.go

type JSFileManager struct {
    managerID string
    runtime   *NodeJSRuntime
}

func (jfm *JSFileManager) Supports(filename string, currentDirectory string) bool {
    resp, err := jfm.runtime.SendCommand(Command{
        Cmd:              "fileManagerSupports",
        ManagerID:        jfm.managerID,
        Filename:         filename,
        CurrentDirectory: currentDirectory,
    })

    if err != nil {
        return false
    }

    return resp.Supports
}

func (jfm *JSFileManager) LoadFile(filename string, currentDirectory string) (*LoadedFile, error) {
    resp, err := jfm.runtime.SendCommand(Command{
        Cmd:              "fileManagerLoad",
        ManagerID:        jfm.managerID,
        Filename:         filename,
        CurrentDirectory: currentDirectory,
    })

    if err != nil {
        return nil, err
    }

    return &LoadedFile{
        Filename: resp.Filename,
        Contents: resp.Contents,
    }, nil
}
```

#### 2. Implement in Node.js

```javascript
const registeredFileManagers = [];

const pluginManager = {
    // ... existing methods

    addFileManager(fileManager) {
        const id = `filemgr_${Date.now()}`;
        registeredFileManagers.push({
            id,
            fileManager
        });
        return { id };
    }
};

function handleFileManagerSupports(cmd) {
    const { managerID, filename, currentDirectory } = cmd;

    const mgr = registeredFileManagers.find(m => m.id === managerID);
    if (!mgr) {
        return { success: false, error: 'File manager not found' };
    }

    try {
        const supports = mgr.fileManager.supports(
            filename,
            currentDirectory,
            {},  // options
            {}   // environment
        );
        return { success: true, supports };
    } catch (error) {
        return { success: false, error: error.message };
    }
}

function handleFileManagerLoad(cmd) {
    const { managerID, filename, currentDirectory } = cmd;

    const mgr = registeredFileManagers.find(m => m.id === managerID);
    if (!mgr) {
        return { success: false, error: 'File manager not found' };
    }

    try {
        const result = mgr.fileManager.loadFile(
            filename,
            currentDirectory,
            {},  // options
            {}   // environment
        );

        // Handle promise or direct return
        if (result && result.then) {
            return result.then(data => ({
                success: true,
                filename: data.filename,
                contents: data.contents
            }));
        }

        return {
            success: true,
            filename: result.filename,
            contents: result.contents
        };

    } catch (error) {
        return { success: false, error: error.message };
    }
}
```

#### 3. Integrate in Import Resolution

```go
// In import resolution code

func (im *ImportManager) ResolveImport(path string, currentDir string) (*LoadedFile, error) {
    // Try plugin file managers first
    for _, fm := range im.PluginManager.GetFileManagers() {
        if fm.Supports(path, currentDir) {
            return fm.LoadFile(path, currentDir)
        }
    }

    // Fall back to default file manager
    return im.defaultFileManager.LoadFile(path, currentDir)
}
```

## Success Criteria

✅ **Phase 8 Complete When**:
- Pre-processors transform source before parsing
- Post-processors transform CSS after generation
- Priority-based ordering works
- Processors can chain (output of one is input of next)
- Unit tests pass

✅ **Phase 9 Complete When**:
- File managers can provide custom import resolution
- `supports()` check works correctly
- `loadFile()` returns correct content
- Falls back to default file manager if no plugin supports
- Unit tests pass

✅ **No Regressions**:
- ALL existing tests still pass: `pnpm -w test:go:unit` (100%)
- NO integration test regressions: `pnpm -w test:go` (183/183)

## Test Requirements

```go
func TestPreProcessor_Transform(t *testing.T)
func TestPostProcessor_Transform(t *testing.T)
func TestProcessors_Priority(t *testing.T)
func TestFileManager_Supports(t *testing.T)
func TestFileManager_LoadFile(t *testing.T)
func TestFileManager_Fallback(t *testing.T)
```

## Deliverables

1. Working pre/post processors
2. Priority-based ordering
3. Working file managers
4. Import resolution integration
5. All unit tests passing
6. No regressions
7. Brief summary

You're adding text transformation and custom imports! ⚙️
