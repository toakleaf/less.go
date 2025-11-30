# AGENT 3: JavaScript Bindings & Node Constructors

**Status**: ⏸️ Blocked - Wait for Agent 1 to complete Phase 2 (AST Serialization)
**Dependencies**: Agent 1 Phase 2 (must have working FlattenAST/UnflattenAST)
**Estimated Time**: 5-7 days
**Blocks**: Agents 4, 5

---

You are implementing JavaScript bindings and node constructors for the less.go plugin system.

## Your Mission

Implement Phase 3 (JavaScript Bindings) and Phase 7 (Tree Node Constructors) from the strategy document.

## Prerequisites

✅ Verify Agent 1 has completed Phase 2:
- `FlattenAST()` and `UnflattenAST()` work
- Buffer serialization to shared memory works
- Roundtrip tests pass

Check: `go test ./runtime -run TestRoundtrip`

## Required Reading

BEFORE starting, read:
1. IMPLEMENTATION_STRATEGY.md - Focus on Phase 3 and Phase 7
2. `packages/less/src/less/tree/*.js` - JavaScript node implementations
3. `packages/less/src/less/less_go/tree/*.go` - Go node implementations
4. Plugin examples: `packages/test-data/plugin/plugin-tree-nodes.js`

## Your Tasks

### Phase 3: JavaScript Bindings

#### 1. Create NodeFacade (Node.js Side)

Create `packages/less/src/less/less_go/runtime/bindings/node-facade.js`:

```javascript
const shm = require('shm-typed-array');

class NodeFacade {
    constructor(buffer, index) {
        this._buffer = buffer;
        this._index = index;
    }

    get type() {
        const typeID = this._buffer.nodes[this._index].typeID;
        return this._buffer.typeTable[typeID];
    }

    get value() {
        const valueIndex = this._buffer.nodes[this._index].valueIndex;
        if (valueIndex === 0) return undefined;
        return this._buffer.stringTable[valueIndex];
    }

    get children() {
        const childIndex = this._buffer.nodes[this._index].childIndex;
        if (childIndex === 0) return [];
        return this._collectSiblings(childIndex);
    }

    _collectSiblings(startIndex) {
        const result = [];
        let index = startIndex;
        while (index !== 0) {
            result.push(new NodeFacade(this._buffer, index));
            index = this._buffer.nodes[index].nextIndex;
        }
        return result;
    }

    // Add methods for each node type
    // Dimension: get unit(), get value()
    // Color: get rgb(), get alpha()
    // etc.
}

module.exports = { NodeFacade };
```

#### 2. Create Visitor Support

Create `packages/less/src/less/less_go/runtime/bindings/visitor.js`:

```javascript
const { NodeFacade } = require('./node-facade');

class Visitor {
    constructor() {
        this.isReplacing = false;
        this.isPreEvalVisitor = false;
    }

    visit(root) {
        const node = new NodeFacade(this._buffer, root);
        return this._visit(node);
    }

    _visit(node) {
        const method = 'visit' + node.type;

        if (this[method]) {
            const result = this[method](node);
            if (result !== node) {
                return result; // Node was replaced
            }
        }

        // Visit children
        for (const child of node.children) {
            this._visit(child);
        }

        return node;
    }
}

module.exports = { Visitor };
```

#### 3. Integrate with plugin-host.js

Update `plugin-host.js` to attach buffer:

```javascript
const { NodeFacade } = require('./bindings/node-facade');
const { Visitor } = require('./bindings/visitor');

let sharedBuffer = null;

function handleAttachBuffer(cmd) {
    const shmKey = cmd.shmKey;
    sharedBuffer = shm.get(shmKey, 'Uint32Array');

    return { success: true };
}

// Expose to plugins
const less = {
    tree: {
        // Will be filled by Phase 7
    },
    visitors: {
        Visitor
    }
};
```

### Phase 7: Tree Node Constructors

#### 1. Generate Constructor Functions (Go Side)

Create `packages/less/src/less/less_go/runtime/codegen/generate_constructors.go`:

```go
package codegen

// Generate JavaScript constructor functions from Go node definitions
func GenerateConstructors() string {
    // Read all node types from tree/*.go
    // Generate JavaScript constructors
    // Return JavaScript code
}
```

Run generation:
```bash
go run packages/less/src/less/less_go/runtime/codegen/generate_constructors.go > \
  packages/less/src/less/less_go/runtime/bindings/constructors.js
```

#### 2. Implement Node Constructors (Node.js Side)

Create `packages/less/src/less/less_go/runtime/bindings/constructors.js`:

```javascript
const { NodeFacade } = require('./node-facade');

// Allocate node in buffer
function allocateNode(type, props) {
    // This will write to shared memory buffer
    // Go will read it back

    const nodeIndex = sharedBuffer.nodes.length;
    sharedBuffer.nodes[nodeIndex] = {
        typeID: getTypeID(type),
        childIndex: 0,
        nextIndex: 0,
        parentIndex: 0,
        valueIndex: props.value ? addString(props.value) : 0,
        propsOffset: 0
    };

    return new NodeFacade(sharedBuffer, nodeIndex);
}

// Constructor functions (generated or hand-written)
const constructors = {
    dimension(value, unit) {
        return allocateNode('Dimension', { value, unit });
    },

    color(rgb, alpha) {
        return allocateNode('Color', { rgb, alpha });
    },

    quoted(quote, value, escaped) {
        return allocateNode('Quoted', { quote, value, escaped });
    },

    keyword(value) {
        return allocateNode('Keyword', { value });
    },

    url(value) {
        return allocateNode('URL', { value });
    },

    call(name, args) {
        return allocateNode('Call', { name, args });
    },

    // Add all node types...
};

module.exports = constructors;
```

#### 3. Expose Constructors to Plugins

Update `plugin-host.js`:

```javascript
const constructors = require('./bindings/constructors');

const less = {
    // Expose constructors directly
    dimension: constructors.dimension,
    color: constructors.color,
    quoted: constructors.quoted,
    keyword: constructors.keyword,
    url: constructors.url,
    call: constructors.call,
    // ... all constructors

    tree: {
        // Also nest under tree for compatibility
        ...constructors
    },

    visitors: {
        Visitor
    }
};
```

## Success Criteria

✅ **Phase 3 Complete When**:
- NodeFacade can read from shared memory buffer
- Can access all node properties (type, value, children)
- Can navigate tree structure (parent, siblings, children)
- Visitor pattern works
- JavaScript unit tests pass: `node runtime/bindings/test.js`

✅ **Phase 7 Complete When**:
- All node constructors implemented
- Constructors create nodes in shared memory
- Go can read constructed nodes from buffer
- Plugin can use `less.dimension()`, `less.color()`, etc.
- Test plugin that creates nodes works

✅ **No Regressions**:
- ALL existing tests still pass: `pnpm -w test:go:unit` (100%)
- NO integration test regressions: `pnpm -w test:go` (183/183)

## Test Requirements

### JavaScript Side (can run with `node`!)

```javascript
// runtime/bindings/test.js
const { NodeFacade } = require('./node-facade');
const assert = require('assert');

// Test facade
function testNodeFacade() {
    // Create mock buffer
    const buffer = createMockBuffer();
    const node = new NodeFacade(buffer, 0);

    assert.equal(node.type, 'Dimension');
    assert.equal(node.value, 10);
}

// Test constructors
function testConstructors() {
    const dim = less.dimension(10, 'px');
    assert.equal(dim.value, 10);
    assert.equal(dim.unit, 'px');
}
```

Run: `node packages/less/src/less/less_go/runtime/bindings/test.js`

### Go Side

```go
func TestBindings_ReadNodeFromJS(t *testing.T) {
    // JS creates a node
    // Go reads it from shared memory
    // Verify properties
}

func TestConstructors_AllNodeTypes(t *testing.T) {
    // For each node type, test constructor
}
```

## Deliverables

1. Working NodeFacade for lazy deserialization
2. Visitor pattern support
3. All node constructors implemented
4. Shared memory node allocation
5. JavaScript unit tests passing
6. Go integration tests passing
7. No regressions
8. Brief summary

You're building the API plugins will use! 🎨
