/**
 * NodeFacade - Lazy deserialization facade for AST nodes from shared memory buffer.
 *
 * This class provides a JavaScript interface to AST nodes stored in Go's FlatAST
 * buffer format. Instead of eagerly deserializing all nodes, it reads properties
 * on-demand, providing significant performance benefits for visitors that only
 * need to inspect a subset of nodes.
 *
 * The facade implements the same interface as less.js tree nodes, allowing
 * plugins to use familiar patterns like node.type, node.value, and visitor methods.
 */

// Node type constants - must match Go's NodeTypeID constants in ast_serializer.go
const NodeTypeID = {
  Unknown: 0,
  Anonymous: 1,
  Assignment: 2,
  AtRule: 3,
  Attribute: 4,
  Call: 5,
  Color: 6,
  Combinator: 7,
  Comment: 8,
  Condition: 9,
  Container: 10,
  Declaration: 11,
  DetachedRuleset: 12,
  Dimension: 13,
  Element: 14,
  Expression: 15,
  Extend: 16,
  Import: 17,
  JavaScript: 18,
  Keyword: 19,
  Media: 20,
  MixinCall: 21,
  MixinDefinition: 22,
  NamespaceValue: 23,
  Negative: 24,
  Operation: 25,
  Paren: 26,
  Property: 27,
  QueryInParens: 28,
  Quoted: 29,
  Ruleset: 30,
  Selector: 31,
  SelectorList: 32,
  UnicodeDescriptor: 33,
  Unit: 34,
  URL: 35,
  Value: 36,
  Variable: 37,
  VariableCall: 38,
  Node: 39,
};

// Reverse mapping for type names
const TypeNames = Object.entries(NodeTypeID).reduce((acc, [name, id]) => {
  acc[id] = name;
  return acc;
}, {});

// FlatNode flags - must match Go's FlatNode flags
const Flags = {
  Parens: 1 << 0,
  ParensInOp: 1 << 1,
  Visible: 1 << 2,
  Invisible: 1 << 3,
  VisibleSet: 1 << 4,
  HasFileInfo: 1 << 5,
  HasIndex: 1 << 6,
};

// Size of a FlatNode in bytes (24 bytes in Go)
const FLAT_NODE_SIZE = 24;

/**
 * NodeFacade provides lazy access to AST nodes from a FlatAST buffer.
 */
class NodeFacade {
  /**
   * Create a new NodeFacade for a node at the given index.
   * @param {Object} ast - The parsed FlatAST structure
   * @param {number} index - The node index in the ast.nodes array
   */
  constructor(ast, index) {
    this._ast = ast;
    this._index = index;
    this._node = ast.nodes[index];
    this._cachedProps = null;
    this._cachedChildren = null;
    this._cachedParent = null;
    this._parentChecked = false;
  }

  /**
   * Get the node type name (e.g., 'Dimension', 'Color', 'Ruleset')
   * @returns {string}
   */
  get type() {
    return TypeNames[this._node.typeID] || 'Unknown';
  }

  /**
   * Get the node type ID
   * @returns {number}
   */
  get typeID() {
    return this._node.typeID;
  }

  /**
   * Get node flags
   * @returns {number}
   */
  get flags() {
    return this._node.flags;
  }

  /**
   * Check if this node has parens
   * @returns {boolean}
   */
  get parens() {
    return (this._node.flags & Flags.Parens) !== 0;
  }

  /**
   * Check if this node has parens in operation
   * @returns {boolean}
   */
  get parensInOp() {
    return (this._node.flags & Flags.ParensInOp) !== 0;
  }

  /**
   * Check if visibility is explicitly set
   * @returns {boolean}
   */
  get visibilitySet() {
    return (this._node.flags & Flags.VisibleSet) !== 0;
  }

  /**
   * Get visibility state (only meaningful if visibilitySet is true)
   * @returns {boolean|undefined}
   */
  get nodeVisible() {
    if (!this.visibilitySet) {
      return undefined;
    }
    return (this._node.flags & Flags.Visible) !== 0;
  }

  /**
   * Get the index in the buffer
   * @returns {number}
   */
  get index() {
    return this._index;
  }

  /**
   * Get the child index (index of first child, 0 if none)
   * @returns {number}
   */
  get childIndex() {
    return this._node.childIndex;
  }

  /**
   * Get the next sibling index (0 if none)
   * @returns {number}
   */
  get nextIndex() {
    return this._node.nextIndex;
  }

  /**
   * Get the parent index (0 if root)
   * @returns {number}
   */
  get parentIndex() {
    return this._node.parentIndex;
  }

  /**
   * Get the parent node facade
   * @returns {NodeFacade|null}
   */
  get parent() {
    if (this._parentChecked) {
      return this._cachedParent;
    }
    this._parentChecked = true;

    // If this is the root or parent index is invalid
    if (this._index === this._ast.rootIndex || this._node.parentIndex === this._index) {
      this._cachedParent = null;
    } else if (this._node.parentIndex >= 0 && this._node.parentIndex < this._ast.nodes.length) {
      this._cachedParent = new NodeFacade(this._ast, this._node.parentIndex);
    }

    return this._cachedParent;
  }

  /**
   * Get all child nodes
   * @returns {NodeFacade[]}
   */
  get children() {
    if (this._cachedChildren !== null) {
      return this._cachedChildren;
    }

    this._cachedChildren = [];
    if (this._node.childIndex === 0) {
      return this._cachedChildren;
    }

    // Collect all siblings starting from first child
    let childIdx = this._node.childIndex;
    while (childIdx !== 0 && childIdx < this._ast.nodes.length) {
      this._cachedChildren.push(new NodeFacade(this._ast, childIdx));
      childIdx = this._ast.nodes[childIdx].nextIndex;
    }

    return this._cachedChildren;
  }

  /**
   * Check if this node has children
   * @returns {boolean}
   */
  hasChildren() {
    return this._node.childIndex !== 0;
  }

  /**
   * Get the next sibling node
   * @returns {NodeFacade|null}
   */
  get nextSibling() {
    if (this._node.nextIndex === 0) {
      return null;
    }
    return new NodeFacade(this._ast, this._node.nextIndex);
  }

  /**
   * Get all siblings (including this node)
   * @returns {NodeFacade[]}
   */
  getSiblings() {
    if (!this.parent) {
      return [this];
    }
    return this.parent.children;
  }

  /**
   * Get node-specific properties (lazily parsed from JSON)
   * @returns {Object}
   */
  get properties() {
    if (this._cachedProps !== null) {
      return this._cachedProps;
    }

    const { propsOffset, propsLength } = this._node;
    if (propsLength === 0) {
      this._cachedProps = {};
      return this._cachedProps;
    }

    try {
      const propData = this._ast.propBuffer.slice(propsOffset, propsOffset + propsLength);
      this._cachedProps = JSON.parse(propData.toString('utf8'));
    } catch (e) {
      this._cachedProps = {};
    }

    return this._cachedProps;
  }

  /**
   * Get a string value from the string table
   * @param {string} propName - Property name containing string index
   * @returns {string}
   */
  getString(propName) {
    const props = this.properties;
    const idx = props[propName];
    if (idx === undefined || idx === null) {
      return '';
    }
    if (typeof idx === 'string') {
      return idx;
    }
    if (typeof idx === 'number' && idx >= 0 && idx < this._ast.stringTable.length) {
      return this._ast.stringTable[idx];
    }
    return '';
  }

  /**
   * Get a number value from properties
   * @param {string} propName - Property name
   * @param {number} defaultValue - Default value if not found
   * @returns {number}
   */
  getNumber(propName, defaultValue = 0) {
    const props = this.properties;
    const val = props[propName];
    if (typeof val === 'number') {
      return val;
    }
    return defaultValue;
  }

  /**
   * Get a boolean value from properties
   * @param {string} propName - Property name
   * @param {boolean} defaultValue - Default value if not found
   * @returns {boolean}
   */
  getBool(propName, defaultValue = false) {
    const props = this.properties;
    const val = props[propName];
    if (typeof val === 'boolean') {
      return val;
    }
    return defaultValue;
  }

  // Type-specific getters (commonly used properties)

  /**
   * Get the value property - common across many node types
   * @returns {*}
   */
  get value() {
    const props = this.properties;

    // Check if value is a string index
    if (typeof props.Value === 'number') {
      return this._ast.stringTable[props.Value] || props.Value;
    }
    if (props.Value !== undefined) {
      return props.Value;
    }

    // Some nodes use lowercase 'value'
    if (typeof props.value === 'number') {
      return this._ast.stringTable[props.value] || props.value;
    }
    if (props.value !== undefined) {
      return props.value;
    }

    // For nodes with children as their "value", return children
    if (this.hasChildren()) {
      return this.children;
    }

    return undefined;
  }

  /**
   * Get the name property (for Variables, Calls, etc.)
   * @returns {string}
   */
  get name() {
    return this.getString('Name') || this.getString('name');
  }

  /**
   * Get the unit property (for Dimensions)
   * @returns {string}
   */
  get unit() {
    return this.getString('Unit') || this.getString('unit');
  }

  /**
   * Get rgb values (for Colors)
   * @returns {number[]}
   */
  get rgb() {
    const props = this.properties;
    if (Array.isArray(props.rgb)) {
      return props.rgb;
    }
    if (Array.isArray(props.Rgb)) {
      return props.Rgb;
    }
    return [0, 0, 0];
  }

  /**
   * Get alpha value (for Colors)
   * @returns {number}
   */
  get alpha() {
    return this.getNumber('Alpha', 1) || this.getNumber('alpha', 1);
  }

  /**
   * Get the quote character (for Quoted)
   * @returns {string}
   */
  get quote() {
    return this.getString('Quote') || this.getString('quote') || '"';
  }

  /**
   * Check if the value is escaped (for Quoted)
   * @returns {boolean}
   */
  get escaped() {
    return this.getBool('Escaped') || this.getBool('escaped');
  }

  /**
   * Get the operator (for Operations, Conditions)
   * @returns {string}
   */
  get op() {
    return this.getString('Op') || this.getString('op');
  }

  /**
   * Get operands (for Operations)
   * @returns {NodeFacade[]}
   */
  get operands() {
    return this.children;
  }

  /**
   * Get selectors (for Rulesets)
   * @returns {NodeFacade[]}
   */
  get selectors() {
    // First children are usually selectors
    const children = this.children;
    // This is a simplification - may need refinement based on actual structure
    return children.filter(c => c.type === 'Selector' || c.type === 'SelectorList');
  }

  /**
   * Get rules (for Rulesets)
   * @returns {NodeFacade[]}
   */
  get rules() {
    const children = this.children;
    return children.filter(c => c.type !== 'Selector' && c.type !== 'SelectorList');
  }

  /**
   * Get elements (for Selectors)
   * @returns {NodeFacade[]}
   */
  get elements() {
    return this.children.filter(c => c.type === 'Element');
  }

  // Compatibility methods to match less.js Node interface

  /**
   * Check if this is a ruleset-like node
   * @returns {boolean}
   */
  isRulesetLike() {
    return this.type === 'Ruleset' || this.type === 'MixinDefinition' || this.type === 'AtRule';
  }

  /**
   * Get visibility info
   * @returns {Object}
   */
  visibilityInfo() {
    return {
      visibilityBlocks: 0, // Not stored in flat format
      nodeVisible: this.nodeVisible,
    };
  }

  /**
   * Accept a visitor (visitor pattern)
   * @param {Object} visitor - Visitor object with visit methods
   * @returns {*}
   */
  accept(visitor) {
    return visitor.visit(this);
  }

  /**
   * Simple eval - facades are already "evaluated" from Go's perspective
   * @returns {NodeFacade}
   */
  eval() {
    return this;
  }

  /**
   * Convert to a plain JavaScript object for serialization
   * @returns {Object}
   */
  toJSON() {
    const result = {
      type: this.type,
      index: this._index,
    };

    if (this.parens) {
      result.parens = true;
    }
    if (this.parensInOp) {
      result.parensInOp = true;
    }

    const props = this.properties;
    if (Object.keys(props).length > 0) {
      result.properties = props;
    }

    if (this.hasChildren()) {
      result.children = this.children.map(c => c.toJSON());
    }

    return result;
  }

  /**
   * Create a string representation for debugging
   * @returns {string}
   */
  toString() {
    const props = this.properties;
    const propStr = Object.keys(props).length > 0 ? JSON.stringify(props) : '';
    return `NodeFacade(${this.type}@${this._index}${propStr ? ', ' + propStr : ''})`;
  }
}

/**
 * Create a NodeFacade from a parsed FlatAST structure.
 * @param {Object} ast - Parsed FlatAST from parseFlatAST()
 * @returns {NodeFacade} - The root node facade
 */
function createRootFacade(ast) {
  if (!ast || !ast.nodes || ast.nodes.length === 0) {
    throw new Error('Invalid or empty AST');
  }
  return new NodeFacade(ast, ast.rootIndex);
}

/**
 * Create a NodeFacade at a specific index
 * @param {Object} ast - Parsed FlatAST
 * @param {number} index - Node index
 * @returns {NodeFacade}
 */
function createFacadeAt(ast, index) {
  if (index < 0 || index >= ast.nodes.length) {
    throw new Error(`Invalid node index: ${index}`);
  }
  return new NodeFacade(ast, index);
}

module.exports = {
  NodeFacade,
  createRootFacade,
  createFacadeAt,
  NodeTypeID,
  TypeNames,
  Flags,
  FLAT_NODE_SIZE,
};
