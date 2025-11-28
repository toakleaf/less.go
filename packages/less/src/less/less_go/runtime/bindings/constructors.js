/**
 * Constructors - Node constructor functions for JavaScript plugins.
 *
 * These constructors create simple JavaScript objects with the correct structure
 * for AST nodes. Plugins use these to create new nodes that can be returned
 * from visitor methods or custom functions.
 *
 * The created objects can be serialized to FlatAST format using BufferWriter.
 */

/**
 * Create a basic node object with type.
 *
 * @param {string} type - Node type name
 * @param {Object} props - Node properties
 * @returns {Object}
 */
function createNode(type, props = {}) {
  return {
    _type: type,
    ...props,
  };
}

// Node constructors - organized by category

// === Value Types ===

/**
 * Create a Dimension node (number with optional unit).
 *
 * @param {number} value - Numeric value
 * @param {string} [unit] - Unit string (e.g., 'px', 'em', '%')
 * @returns {Object}
 */
function dimension(value, unit) {
  return createNode('Dimension', {
    value: typeof value === 'number' ? value : parseFloat(value),
    unit: unit || '',
  });
}

/**
 * Create a Color node.
 *
 * @param {number[]} rgb - RGB values [r, g, b] (0-255)
 * @param {number} [alpha] - Alpha value (0-1)
 * @returns {Object}
 */
function color(rgb, alpha) {
  // Handle both [r,g,b] array and {r,g,b} object
  let rgbArray = rgb;
  if (!Array.isArray(rgb)) {
    rgbArray = [rgb.r || 0, rgb.g || 0, rgb.b || 0];
  }
  return createNode('Color', {
    rgb: rgbArray,
    alpha: alpha !== undefined ? alpha : 1,
  });
}

/**
 * Create a Quoted string node.
 *
 * @param {string} quote - Quote character ('"' or "'")
 * @param {string} value - String value
 * @param {boolean} [escaped] - Whether the value is escaped
 * @returns {Object}
 */
function quoted(quote, value, escaped) {
  return createNode('Quoted', {
    quote: quote || '"',
    value: value || '',
    escaped: escaped || false,
  });
}

/**
 * Create a Keyword node.
 *
 * @param {string} value - Keyword value
 * @returns {Object}
 */
function keyword(value) {
  return createNode('Keyword', {
    value: value || '',
  });
}

/**
 * Create an Anonymous node (raw CSS value).
 *
 * @param {string} value - Raw CSS value
 * @returns {Object}
 */
function anonymous(value) {
  return createNode('Anonymous', {
    value: value || '',
  });
}

/**
 * Create a URL node.
 *
 * @param {Object|string} value - URL value or node
 * @param {Object} [paths] - Path info
 * @returns {Object}
 */
function url(value, paths) {
  if (typeof value === 'string') {
    return createNode('URL', {
      value: quoted('"', value),
      paths: paths,
    });
  }
  return createNode('URL', {
    value: value,
    paths: paths,
  });
}

/**
 * Create a Variable node.
 *
 * @param {string} name - Variable name (with @)
 * @returns {Object}
 */
function variable(name) {
  return createNode('Variable', {
    name: name || '',
  });
}

/**
 * Create a Unit node.
 *
 * @param {string[]} [numerator] - Numerator units
 * @param {string[]} [denominator] - Denominator units
 * @returns {Object}
 */
function unit(numerator, denominator) {
  return createNode('Unit', {
    numerator: numerator || [],
    denominator: denominator || [],
  });
}

// === Structural Types ===

/**
 * Create a Value node (list of expressions).
 *
 * @param {Array|*} value - Expression(s)
 * @returns {Object}
 */
function value(val) {
  const children = Array.isArray(val) ? val : [val];
  return createNode('Value', {
    children: children,
  });
}

/**
 * Create an Expression node (space-separated values).
 *
 * @param {Array|*} value - Value(s)
 * @returns {Object}
 */
function expression(val) {
  const children = Array.isArray(val) ? val : [val];
  return createNode('Expression', {
    children: children,
  });
}

/**
 * Create a Paren node (parenthesized expression).
 *
 * @param {Object} node - Inner node
 * @returns {Object}
 */
function paren(node) {
  return createNode('Paren', {
    value: node,
    children: [node],
  });
}

/**
 * Create a Negative node.
 *
 * @param {Object} node - Node to negate
 * @returns {Object}
 */
function negative(node) {
  return createNode('Negative', {
    value: node,
    children: [node],
  });
}

// === Operations ===

/**
 * Create an Operation node.
 *
 * @param {string} op - Operator (+, -, *, /)
 * @param {Array} operands - Operand nodes
 * @returns {Object}
 */
function operation(op, operands) {
  return createNode('Operation', {
    op: op,
    operands: operands,
    children: operands,
  });
}

/**
 * Create a Condition node.
 *
 * @param {string} op - Comparison operator
 * @param {Object} lvalue - Left value
 * @param {Object} rvalue - Right value
 * @param {boolean} [negate] - Whether to negate the condition
 * @returns {Object}
 */
function condition(op, lvalue, rvalue, negate) {
  return createNode('Condition', {
    op: op,
    lvalue: lvalue,
    rvalue: rvalue,
    negate: negate || false,
    children: [lvalue, rvalue],
  });
}

// === Function Calls ===

/**
 * Create a Call node (function call).
 *
 * @param {string} name - Function name
 * @param {Array} [args] - Function arguments
 * @returns {Object}
 */
function call(name, args) {
  return createNode('Call', {
    name: name,
    args: args || [],
    children: args || [],
  });
}

// === Selectors ===

/**
 * Create a Combinator node.
 *
 * @param {string} value - Combinator value (' ', '>', '+', '~')
 * @returns {Object}
 */
function combinator(val) {
  return createNode('Combinator', {
    value: val || '',
  });
}

/**
 * Create an Element node.
 *
 * @param {Object|string} combinator - Combinator node or value
 * @param {string|Object} value - Element value
 * @returns {Object}
 */
function element(comb, val) {
  const combNode = typeof comb === 'string' ? combinator(comb) : comb;
  return createNode('Element', {
    combinator: combNode,
    value: val,
    children: combNode ? [combNode] : [],
  });
}

/**
 * Create a Selector node.
 *
 * @param {Array} elements - Element nodes
 * @returns {Object}
 */
function selector(elements) {
  const elems = Array.isArray(elements) ? elements : [elements];
  return createNode('Selector', {
    elements: elems,
    children: elems,
  });
}

// === Rules ===

/**
 * Create a Ruleset node.
 *
 * @param {Array|Object} selectors - Selector(s)
 * @param {Array} rules - Rule nodes
 * @returns {Object}
 */
function ruleset(selectors, rules) {
  const sels = Array.isArray(selectors) ? selectors : selectors ? [selectors] : [];
  const rs = rules || [];
  return createNode('Ruleset', {
    selectors: sels,
    rules: rs,
    children: [...sels, ...rs],
  });
}

/**
 * Create a Declaration node (property: value).
 *
 * @param {string} name - Property name
 * @param {Object} value - Value node
 * @param {string} [important] - Important flag
 * @param {boolean} [merge] - Merge flag
 * @param {boolean} [inline] - Inline flag
 * @param {boolean} [variable] - Is variable declaration
 * @returns {Object}
 */
function declaration(name, val, important, merge, inline, isVariable) {
  return createNode('Declaration', {
    name: name,
    value: val,
    important: important || '',
    merge: merge || false,
    inline: inline || false,
    variable: isVariable || false,
    children: [val],
  });
}

/**
 * Create a DetachedRuleset node.
 *
 * @param {Object} rulesetNode - Ruleset node
 * @returns {Object}
 */
function detachedruleset(rulesetNode) {
  return createNode('DetachedRuleset', {
    ruleset: rulesetNode,
    children: rulesetNode ? [rulesetNode] : [],
  });
}

// === At-Rules ===

/**
 * Create an AtRule node.
 *
 * @param {string} name - At-rule name (with @)
 * @param {Object} [value] - Value node
 * @param {Array} [rules] - Rules array
 * @param {number} [index] - Source index
 * @param {boolean} [isRooted] - Is rooted
 * @returns {Object}
 */
function atrule(name, val, rules, index, isRooted) {
  const children = [];
  if (val) children.push(val);
  if (rules) children.push(...rules);

  return createNode('AtRule', {
    name: name,
    value: val,
    rules: rules,
    isRooted: isRooted || false,
    children: children,
  });
}

// === Other ===

/**
 * Create an Assignment node (for mixin arguments).
 *
 * @param {string} key - Assignment key
 * @param {*} value - Assignment value
 * @returns {Object}
 */
function assignment(key, val) {
  return createNode('Assignment', {
    key: key,
    value: val,
  });
}

/**
 * Create an Attribute node (for attribute selectors).
 *
 * @param {string} key - Attribute name
 * @param {string} [op] - Operator (=, ~=, |=, etc.)
 * @param {*} [value] - Attribute value
 * @returns {Object}
 */
function attribute(key, op, val) {
  return createNode('Attribute', {
    key: key,
    op: op || '',
    value: val,
  });
}

/**
 * Create a Comment node.
 *
 * @param {string} value - Comment text
 * @param {boolean} [isLineComment] - Is a line comment
 * @returns {Object}
 */
function comment(val, isLineComment) {
  return createNode('Comment', {
    value: val,
    isLineComment: isLineComment || false,
  });
}

/**
 * Create a UnicodeDescriptor node.
 *
 * @param {string} value - Unicode range
 * @returns {Object}
 */
function unicodeDescriptor(val) {
  return createNode('UnicodeDescriptor', {
    value: val,
  });
}

// Export all constructors
module.exports = {
  // Core helper
  createNode,

  // Value types
  dimension,
  color,
  quoted,
  keyword,
  anonymous,
  url,
  variable,
  unit,

  // Structural types
  value,
  expression,
  paren,
  negative,

  // Operations
  operation,
  condition,

  // Function calls
  call,

  // Selectors
  combinator,
  element,
  selector,

  // Rules
  ruleset,
  declaration,
  detachedruleset,

  // At-rules
  atrule,

  // Other
  assignment,
  attribute,
  comment,
  unicodeDescriptor,
};
