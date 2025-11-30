# Agent 1: JavaScript Side - plugin-host.js

**Status**: 🔴 Can start immediately
**Can run in parallel with**: Agent 2
**Estimated Time**: 2-3 hours
**Blocks**: Agent 3

---

## Your Mission

Add inline JavaScript evaluation support to `plugin-host.js` by implementing the `evalJS` command handler. This enables Less.js backtick expressions like `` `1 + 1` `` to execute via the existing Node.js runtime.

## Required Reading

Before starting, read these files:
1. `.claude/tasks/inline-js/README.md` - Overview (~5 min)
2. `.claude/tasks/inline-js/TASK_BREAKDOWN.md` - Your tasks are 1.1-1.5 (~10 min)
3. `packages/less/src/less/tree/javascript.js` - Reference implementation (~10 min)
4. `packages/less/src/less/tree/js-eval-node.js` - How variables are exposed (~10 min)

## Background Context

### What We're Implementing

Less.js supports inline JavaScript:

```less
.example {
    @foo: 42;
    width: `1 + 1`;                      // → 2
    height: `parseInt(this.foo.toJS())`; // → 42 (access Less variable)
    content: ~`"hello" + " world"`;      // → hello world (escaped)
}
```

### Existing Infrastructure

- `plugin-host.js` already handles plugin loading, function calls, visitors
- Go side will send an `evalJS` command with expression and variable context
- You just need to add the handler

## Your Tasks

### Task 1.1: Add evalJS Command Handler

**File**: `packages/less/src/less/less_go/runtime/plugin-host.js`

Add to the `handleCommand` switch statement (around line 350):

```javascript
case 'evalJS':
  handleEvalJS(id, data);
  break;
```

Then implement the handler function:

```javascript
/**
 * Evaluate an inline JavaScript expression
 * @param {number} id - Command ID
 * @param {Object} data - { expression: string, variables: { name: { value: string } } }
 */
function handleEvalJS(id, data) {
  const { expression, variables } = data || {};

  if (expression === undefined) {
    sendResponse(id, false, null, 'Expression is required');
    return;
  }

  try {
    // Build evaluation context with variables accessible as this.varName
    const evalContext = buildEvalContext(variables);

    // Create function and execute with context as 'this'
    // This matches less.js: new Function(`return (${expression})`)
    const fn = new Function(`return (${expression})`);
    const result = fn.call(evalContext);

    // Convert result to serializable format
    const converted = convertJSResult(result);

    sendResponse(id, true, converted);
  } catch (err) {
    // Format error for Less error reporting
    sendResponse(id, false, null, formatJSError(err));
  }
}
```

---

### Task 1.2: Build Evaluation Context

Less variables are exposed as `this.varName` with a `toJS()` method.

```javascript
/**
 * Build evaluation context from Less variables
 * Variables are exposed as this.varName with a toJS() method
 * @param {Object} variables - { name: { value: string } }
 * @returns {Object} Context object for function.call()
 */
function buildEvalContext(variables) {
  const context = {};

  if (!variables) {
    return context;
  }

  for (const [name, info] of Object.entries(variables)) {
    // Remove @ prefix if present (Go may send "@foo" or "foo")
    const cleanName = name.startsWith('@') ? name.slice(1) : name;

    // Create variable object with toJS() method
    // This matches less.js behavior in js-eval-node.js
    context[cleanName] = {
      value: info.value,
      toJS: function() {
        return this.value;
      }
    };
  }

  return context;
}
```

---

### Task 1.3: Convert JavaScript Results

Convert JavaScript results to a serializable format that Go can process.

```javascript
/**
 * Convert JavaScript result to serializable format
 * Matches less.js behavior in tree/javascript.js eval()
 * @param {*} result - The JavaScript result
 * @returns {Object} Serializable result with type info
 */
function convertJSResult(result) {
  const type = typeof result;

  // Number (not NaN)
  if (type === 'number' && !isNaN(result)) {
    return { type: 'number', value: result };
  }

  // String
  if (type === 'string') {
    return { type: 'string', value: result };
  }

  // Array - join with comma+space (matches less.js)
  if (Array.isArray(result)) {
    return { type: 'array', value: result.join(', ') };
  }

  // Boolean
  if (type === 'boolean') {
    return { type: 'boolean', value: result };
  }

  // Null/undefined/NaN - treat as empty
  if (result === null || result === undefined || (type === 'number' && isNaN(result))) {
    return { type: 'empty', value: '' };
  }

  // Object or other - convert to string
  return { type: 'other', value: String(result) };
}
```

---

### Task 1.4: Format JavaScript Errors

Format errors to match less.js error message format.

```javascript
/**
 * Format JavaScript error for Less error reporting
 * Matches less.js format: "JavaScript evaluation error: 'TypeError: ...'"
 * @param {Error} err - The JavaScript error
 * @returns {string} Formatted error message
 */
function formatJSError(err) {
  // Match less.js format from tree/js-eval-node.js line 40
  const errName = err.name || 'Error';
  const errMsg = err.message || String(err);
  // Replace double quotes with single quotes in message
  return `JavaScript evaluation error: '${errName}: ${errMsg.replace(/["]/g, "'")}'`;
}
```

---

### Task 1.5: Handle Edge Cases

The test file `packages/test-data/less/_main/javascript.less` has several edge cases:

1. **Empty function result**: `` `+function(){}` `` returns empty string
2. **Multiline expressions**: Expressions can span multiple lines
3. **IIFE**: `` `(function(){ var x = 1; return x })()` ``
4. **typeof**: `` `typeof process.title` `` (Node.js environment)

The current implementation should handle these, but verify:

```javascript
// In handleEvalJS, these edge cases should work:
// 1. Empty result → { type: 'empty', value: '' }
// 2. Multiline → new Function handles this
// 3. IIFE → returns result of function call
// 4. typeof → returns 'string' or 'undefined'
```

**Special case for `process.title`**: Since we're in Node.js, `process.title` exists. The test expects `typeof process.title` to return `"string"`.

---

## Test Your Implementation

Create a simple test script:

```bash
# Save this as test-evaljs.js and run from the runtime directory
cd packages/less/src/less/less_go/runtime

cat > test-evaljs.js << 'EOF'
// Mock the sendResponse and test handleEvalJS
const { handleEvalJS } = require('./test-helpers');

// Or inline test:
function buildEvalContext(variables) {
  const context = {};
  if (!variables) return context;
  for (const [name, info] of Object.entries(variables)) {
    const cleanName = name.startsWith('@') ? name.slice(1) : name;
    context[cleanName] = {
      value: info.value,
      toJS: function() { return this.value; }
    };
  }
  return context;
}

function convertJSResult(result) {
  const type = typeof result;
  if (type === 'number' && !isNaN(result)) return { type: 'number', value: result };
  if (type === 'string') return { type: 'string', value: result };
  if (Array.isArray(result)) return { type: 'array', value: result.join(', ') };
  if (type === 'boolean') return { type: 'boolean', value: result };
  if (result === null || result === undefined) return { type: 'empty', value: '' };
  return { type: 'other', value: String(result) };
}

// Test cases
const tests = [
  { expr: '1 + 1', expected: { type: 'number', value: 2 } },
  { expr: '"hello"', expected: { type: 'string', value: 'hello' } },
  { expr: '[1, 2, 3]', expected: { type: 'array', value: '1, 2, 3' } },
  { expr: 'true', expected: { type: 'boolean', value: true } },
  { expr: '+function(){}', expected: { type: 'empty', value: '' } },  // NaN
  { expr: 'typeof process.title', expected: { type: 'string', value: 'string' } },
];

for (const test of tests) {
  const fn = new Function(`return (${test.expr})`);
  const result = fn.call({});
  const converted = convertJSResult(result);
  const pass = JSON.stringify(converted) === JSON.stringify(test.expected);
  console.log(`${pass ? '✓' : '✗'} ${test.expr} → ${JSON.stringify(converted)}`);
}

// Test variable access
const ctx = buildEvalContext({ '@foo': { value: '42' } });
const fn = new Function(`return parseInt(this.foo.toJS())`);
const result = fn.call(ctx);
console.log(`✓ this.foo.toJS() → ${result} (expected 42)`);
EOF

node test-evaljs.js
rm test-evaljs.js
```

---

## Verification Checklist

Before marking complete:

- [ ] Added `evalJS` case to handleCommand switch
- [ ] Implemented `handleEvalJS` function
- [ ] Implemented `buildEvalContext` function
- [ ] Implemented `convertJSResult` function
- [ ] Implemented `formatJSError` function
- [ ] Tested with basic expressions: `1 + 1`, `"hello"`, `[1,2,3]`
- [ ] Tested with variable access: `this.foo.toJS()`
- [ ] Tested edge cases: empty function, multiline, typeof
- [ ] Tested error case: `this.undefined.property`
- [ ] No syntax errors in plugin-host.js

---

## Deliverables

When complete, provide:

1. **Summary**: What you implemented (2-3 sentences)
2. **Files modified**: List with brief description
3. **Test results**: Output of your test script
4. **Issues encountered**: Any blockers or surprises

---

## Reference: less.js Implementation

**tree/js-eval-node.js** (evaluation context):
```javascript
const variables = context.frames[0].variables();
for (const k in variables) {
    evalContext[k.slice(1)] = {  // Remove @ prefix
        value: variables[k].value,
        toJS: function () {
            return this.value.eval(context).toCSS();
        }
    };
}
```

**tree/javascript.js** (result conversion):
```javascript
eval(context) {
    const result = this.evaluateJavaScript(this.expression, context);
    const type = typeof result;

    if (type === 'number' && !isNaN(result)) {
        return new Dimension(result);
    } else if (type === 'string') {
        return new Quoted(`"${result}"`, result, this.escaped, this._index);
    } else if (Array.isArray(result)) {
        return new Anonymous(result.join(', '));
    } else {
        return new Anonymous(result);
    }
}
```

Good luck! You're enabling a key Less.js feature. 🚀
