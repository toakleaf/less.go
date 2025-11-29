/**
 * AST Bindings for less.go plugin-host
 *
 * This module provides functions to create tree-like facades over the
 * binary FlatAST format, enabling JavaScript visitors to walk and
 * transform the AST.
 */

/**
 * Type ID to type name mapping (must match ast_serializer.go)
 */
const typeNames = {
  0: 'Unknown',
  1: 'Anonymous',
  2: 'Assignment',
  3: 'AtRule',
  4: 'Attribute',
  5: 'Call',
  6: 'Color',
  7: 'Combinator',
  8: 'Comment',
  9: 'Condition',
  10: 'Container',
  11: 'Declaration',
  12: 'DetachedRuleset',
  13: 'Dimension',
  14: 'Element',
  15: 'Expression',
  16: 'Extend',
  17: 'Import',
  18: 'JavaScript',
  19: 'Keyword',
  20: 'Media',
  21: 'MixinCall',
  22: 'MixinDefinition',
  23: 'NamespaceValue',
  24: 'Negative',
  25: 'Operation',
  26: 'Paren',
  27: 'Property',
  28: 'QueryInParens',
  29: 'Quoted',
  30: 'Ruleset',
  31: 'Selector',
  32: 'SelectorList',
  33: 'UnicodeDescriptor',
  34: 'Unit',
  35: 'URL',
  36: 'Value',
  37: 'Variable',
  38: 'VariableCall',
  39: 'Node',
};

/**
 * NodeFacade wraps a flat AST node to provide a tree-like interface.
 * It lazily reconstructs children on access.
 */
class NodeFacade {
  constructor(ast, nodeIndex) {
    this._ast = ast;
    this._nodeIndex = nodeIndex;
    this._node = ast.nodes[nodeIndex];
    this._childrenCache = null;
    this._propertiesCache = null;

    // Set type
    this._type = typeNames[this._node.typeID] || 'Unknown';
    this.type = this._type;

    // Extract properties from prop buffer
    this._loadProperties();
  }

  /**
   * Load properties from the AST's prop buffer
   */
  _loadProperties() {
    const node = this._node;
    if (node.propsLength > 0 &&
        node.propsOffset + node.propsLength <= this._ast.propBuffer.length) {
      try {
        const propData = this._ast.propBuffer.slice(
          node.propsOffset,
          node.propsOffset + node.propsLength
        );
        this._propertiesCache = JSON.parse(propData.toString('utf8'));

        // Resolve string table references and set properties directly on this
        for (const [key, value] of Object.entries(this._propertiesCache)) {
          if (typeof value === 'number' &&
              value < this._ast.stringTable.length &&
              key !== 'value' &&
              !key.includes('Index')) {
            // This is likely a string table index
            this[key] = this._ast.stringTable[value];
          } else {
            this[key] = value;
          }
        }
      } catch (e) {
        this._propertiesCache = {};
      }
    } else {
      this._propertiesCache = {};
    }
  }

  /**
   * Get child nodes lazily
   */
  get children() {
    if (this._childrenCache !== null) {
      return this._childrenCache;
    }

    this._childrenCache = [];
    if (this._node.childIndex > 0) {
      let childIdx = this._node.childIndex;
      while (childIdx > 0 && childIdx < this._ast.nodes.length) {
        this._childrenCache.push(new NodeFacade(this._ast, childIdx));
        childIdx = this._ast.nodes[childIdx].nextIndex;
      }
    }

    return this._childrenCache;
  }

  /**
   * Get the node type
   */
  getType() {
    return this._type;
  }

  /**
   * Check if node has a specific type
   */
  isType(typeName) {
    return this._type === typeName;
  }

  /**
   * Get a child-like property (rules, value, etc.)
   * Less.js nodes have different properties that contain children
   */
  getRules() { return this._getChildProperty('rules'); }
  getValue() { return this._getChildProperty('value'); }
  getElements() { return this._getChildProperty('elements'); }
  getArgs() { return this._getChildProperty('args'); }
  getSelectors() { return this._getChildProperty('selectors'); }

  _getChildProperty(name) {
    // First check if it's a direct property
    if (this[name] !== undefined) {
      return this[name];
    }
    // Otherwise, return children as the default "rules" or "value"
    // depending on node type
    return this.children;
  }

  /**
   * Convert back to a plain object for serialization
   */
  toObject() {
    const obj = {
      _type: this._type,
      type: this._type,
    };

    // Copy all properties
    for (const [key, value] of Object.entries(this._propertiesCache || {})) {
      if (typeof value === 'number' && value < this._ast.stringTable.length) {
        obj[key] = this._ast.stringTable[value];
      } else {
        obj[key] = value;
      }
    }

    // Handle children
    if (this.children.length > 0) {
      obj.children = this.children.map(c => c.toObject());
    }

    return obj;
  }
}

/**
 * Creates a root facade for the entire AST tree.
 * This is used by visitors to walk the tree.
 */
function createRootFacade(ast) {
  if (!ast || !ast.nodes || ast.nodes.length === 0) {
    return null;
  }

  const rootIndex = ast.rootIndex || 0;
  return new NodeFacade(ast, rootIndex);
}

/**
 * Recursively visit a node and its children, applying visitor methods.
 * Returns the modified tree (or a replacement node).
 *
 * @param {Object} visitor - The visitor implementation with visitXxx methods
 * @param {NodeFacade|Object} node - The node to visit
 * @param {Array} replacements - Array to collect replacements
 * @param {number} parentIndex - Parent node index for replacement tracking
 * @param {number} childIndex - Child index within parent for replacement tracking
 */
function visitNode(visitor, node, replacements = [], parentIndex = -1, childIndex = -1) {
  if (!node) {
    return node;
  }

  // Get the type name
  const type = node._type || node.type;
  if (!type) {
    return node;
  }

  // Check for a visitor method for this type
  const funcName = 'visit' + type;
  let result = node;

  if (visitor[funcName]) {
    result = visitor[funcName](node);

    // If the visitor returned a different node, track the replacement
    if (result !== node && result !== undefined) {
      if (parentIndex >= 0 && childIndex >= 0) {
        replacements.push({
          parentIndex,
          childIndex,
          replacement: result,
        });
      }
      // Return the replacement - don't visit its children since it's new
      return result;
    }
  }

  // Visit children recursively
  const children = node.children || [];
  const nodeIndex = node._nodeIndex !== undefined ? node._nodeIndex : parentIndex;

  for (let i = 0; i < children.length; i++) {
    const child = children[i];
    const childResult = visitNode(visitor, child, replacements, nodeIndex, i);

    // If child was replaced, update the children array
    if (childResult !== child) {
      children[i] = childResult;
    }
  }

  // Check for visitXxxOut method (called after children are visited)
  const outFuncName = 'visit' + type + 'Out';
  if (visitor[outFuncName]) {
    visitor[outFuncName](result);
  }

  return result;
}

/**
 * Run a visitor on the AST tree.
 * This is used by the plugin-host to run pre-eval and post-eval visitors.
 *
 * @param {Object} visitor - The visitor with run() or visit methods
 * @param {Object} ast - The parsed FlatAST
 * @returns {Object} Result with replacements array
 */
function runVisitor(visitor, ast) {
  const root = createRootFacade(ast);
  if (!root) {
    return { success: true, replacements: [] };
  }

  const replacements = [];

  // If visitor has native wrapper (less.js style), use its visit method
  if (visitor.native && typeof visitor.native.visit === 'function') {
    // The native visitor wraps our recursive visiting
    const nativeVisitor = visitor.native;
    nativeVisitor._visitImpl = visitor;

    // Override the native visit to do recursive walking
    const originalVisit = nativeVisitor.visit.bind(nativeVisitor);
    nativeVisitor.visit = (node) => {
      return visitNode(visitor, node, replacements, -1, -1);
    };

    // Run the visitor
    if (visitor.run) {
      visitor.run(root);
    }
  } else if (visitor.run) {
    // Call run directly if defined
    visitor.run(root);
  } else {
    // Fall back to walking the tree with the visitor
    visitNode(visitor, root, replacements, -1, -1);
  }

  return {
    success: true,
    replacements: replacements,
  };
}

/**
 * Serialize a JavaScript node back to a format Go can understand.
 * This is a simplified version that returns JSON.
 */
function serializeToBuffer(node) {
  // For now, return JSON representation
  const json = JSON.stringify(node);
  return Buffer.from(json, 'utf8');
}

module.exports = {
  NodeFacade,
  createRootFacade,
  visitNode,
  runVisitor,
  serializeToBuffer,
  typeNames,
};
