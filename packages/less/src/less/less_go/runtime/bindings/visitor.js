/**
 * Visitor - AST traversal and transformation support for plugin visitors.
 *
 * This module provides a Visitor base class that plugins can extend to
 * implement pre-eval and post-eval visitors. The visitor pattern allows
 * plugins to inspect and transform AST nodes during compilation.
 *
 * Visitors work with NodeFacade objects for reading from the shared memory
 * buffer, but can return new node objects for replacements.
 */

const { NodeFacade, createRootFacade, TypeNames } = require('./node-facade');

/**
 * Visitor base class for AST traversal.
 *
 * Plugins extend this class and implement visit* methods for specific node types.
 * The visitor calls these methods during traversal, passing NodeFacade objects.
 *
 * Example:
 *   class MyVisitor extends Visitor {
 *     constructor() {
 *       super();
 *       this.isPreEvalVisitor = true;
 *     }
 *
 *     visitDimension(node) {
 *       // Transform dimension nodes
 *       if (node.value === 10) {
 *         return less.dimension(20, node.unit);
 *       }
 *       return node;
 *     }
 *   }
 */
class Visitor {
  constructor() {
    /**
     * Set to true for pre-evaluation visitors (run before evaluation)
     * @type {boolean}
     */
    this.isPreEvalVisitor = false;

    /**
     * Set to true if this visitor can replace nodes
     * @type {boolean}
     */
    this.isReplacing = false;

    /**
     * Cache of visit function names to avoid repeated lookups
     * @private
     */
    this._visitFnCache = {};

    /**
     * Track visitor state during traversal
     * @private
     */
    this._visitStack = [];
  }

  /**
   * Run the visitor on an AST root node.
   * This is the main entry point called by the plugin system.
   *
   * @param {NodeFacade|Object} root - The AST root to visit
   * @returns {NodeFacade|Object} - The (possibly transformed) root
   */
  run(root) {
    if (!root) {
      return root;
    }
    return this.visit(root);
  }

  /**
   * Visit a single node, calling the appropriate visit* method.
   *
   * @param {NodeFacade|Object} node - The node to visit
   * @returns {NodeFacade|Object} - The (possibly transformed) node
   */
  visit(node) {
    if (!node) {
      return node;
    }

    // Get the node type - support both NodeFacade and plain objects
    const type = node.type || node._type;
    if (!type) {
      return node;
    }

    // Check for a visit method for this type
    const funcName = 'visit' + type;
    const visitFn = this[funcName];

    if (typeof visitFn === 'function') {
      // Call the visit method
      const result = visitFn.call(this, node);

      // If the visitor returned a different node, use it
      if (result !== undefined && result !== node) {
        return result;
      }
    }

    // Visit children if the node supports it
    if (typeof node.children !== 'undefined') {
      this.visitChildren(node);
    }

    // Call visitOut if it exists
    const visitOutFn = this[funcName + 'Out'];
    if (typeof visitOutFn === 'function') {
      visitOutFn.call(this, node);
    }

    return node;
  }

  /**
   * Visit all children of a node.
   *
   * @param {NodeFacade|Object} node - The parent node
   */
  visitChildren(node) {
    const children = node.children;
    if (!children || !Array.isArray(children)) {
      return;
    }

    for (let i = 0; i < children.length; i++) {
      const child = children[i];
      if (child) {
        const result = this.visit(child);
        // If this is a replacing visitor and result differs, update would be tracked
        // For now, we track replacements separately since facades are read-only
        if (this.isReplacing && result !== child) {
          // Store replacement for later application
          this._storeReplacement(node, i, result);
        }
      }
    }
  }

  /**
   * Store a node replacement for later application.
   * Since NodeFacade is read-only, replacements are tracked and applied later.
   *
   * @private
   * @param {NodeFacade|Object} parent - Parent node
   * @param {number} childIndex - Index in children array
   * @param {Object} replacement - The replacement node
   */
  _storeReplacement(parent, childIndex, replacement) {
    if (!this._replacements) {
      this._replacements = [];
    }
    this._replacements.push({
      parentIndex: parent.index !== undefined ? parent.index : -1,
      childIndex: childIndex,
      replacement: replacement,
    });
  }

  /**
   * Get all replacements made during traversal.
   *
   * @returns {Array} - Array of replacement records
   */
  getReplacements() {
    return this._replacements || [];
  }

  /**
   * Clear stored replacements.
   */
  clearReplacements() {
    this._replacements = [];
  }

  /**
   * Visit an array of nodes.
   *
   * @param {Array} nodes - Array of nodes to visit
   * @returns {Array} - Array of (possibly transformed) nodes
   */
  visitArray(nodes) {
    if (!Array.isArray(nodes)) {
      return nodes;
    }

    const result = [];
    for (const node of nodes) {
      const visited = this.visit(node);
      if (visited !== null && visited !== undefined) {
        if (Array.isArray(visited)) {
          result.push(...visited);
        } else {
          result.push(visited);
        }
      }
    }
    return result;
  }
}

/**
 * VisitorContext manages visitor execution on a buffer-based AST.
 *
 * This is used by the plugin host to run visitors registered by plugins
 * on the shared memory AST buffer.
 */
class VisitorContext {
  /**
   * Create a new VisitorContext.
   *
   * @param {Object} ast - Parsed FlatAST from parseFlatAST()
   */
  constructor(ast) {
    this._ast = ast;
    this._visitors = [];
    this._preEvalVisitors = [];
    this._postEvalVisitors = [];
  }

  /**
   * Register a visitor.
   *
   * @param {Visitor} visitor - The visitor to register
   */
  addVisitor(visitor) {
    this._visitors.push(visitor);

    if (visitor.isPreEvalVisitor) {
      this._preEvalVisitors.push(visitor);
    } else {
      this._postEvalVisitors.push(visitor);
    }
  }

  /**
   * Run all pre-evaluation visitors on the AST.
   *
   * @returns {Object} - Results including any replacements
   */
  runPreEvalVisitors() {
    return this._runVisitors(this._preEvalVisitors);
  }

  /**
   * Run all post-evaluation visitors on the AST.
   *
   * @returns {Object} - Results including any replacements
   */
  runPostEvalVisitors() {
    return this._runVisitors(this._postEvalVisitors);
  }

  /**
   * Run a list of visitors on the AST.
   *
   * @private
   * @param {Visitor[]} visitors - Visitors to run
   * @returns {Object} - Results
   */
  _runVisitors(visitors) {
    const root = createRootFacade(this._ast);
    const allReplacements = [];

    for (const visitor of visitors) {
      visitor.clearReplacements();
      visitor.run(root);

      const replacements = visitor.getReplacements();
      if (replacements.length > 0) {
        allReplacements.push({
          visitorIndex: this._visitors.indexOf(visitor),
          replacements: replacements,
        });
      }
    }

    return {
      success: true,
      replacements: allReplacements,
      visitorCount: visitors.length,
    };
  }

  /**
   * Run a specific visitor by index.
   *
   * @param {number} visitorIndex - Index of visitor in the visitors array
   * @returns {Object} - Results
   */
  runVisitor(visitorIndex) {
    if (visitorIndex < 0 || visitorIndex >= this._visitors.length) {
      return {
        success: false,
        error: `Invalid visitor index: ${visitorIndex}`,
      };
    }

    const visitor = this._visitors[visitorIndex];
    const root = createRootFacade(this._ast);

    visitor.clearReplacements();
    const result = visitor.run(root);

    return {
      success: true,
      replacements: visitor.getReplacements(),
      resultIndex: result ? result.index : null,
    };
  }

  /**
   * Get visitor info for all registered visitors.
   *
   * @returns {Array} - Array of visitor metadata
   */
  getVisitorInfo() {
    return this._visitors.map((v, i) => ({
      index: i,
      isPreEvalVisitor: v.isPreEvalVisitor || false,
      isReplacing: v.isReplacing || false,
    }));
  }
}

/**
 * Create a visitor implementation from a plain object.
 *
 * This allows plugins to use simpler syntax:
 *   const myVisitor = createVisitor({
 *     isPreEvalVisitor: true,
 *     visitDimension(node) { ... }
 *   });
 *
 * @param {Object} impl - Visitor implementation object
 * @returns {Visitor} - A Visitor instance with the implementation
 */
function createVisitor(impl) {
  const visitor = new Visitor();

  // Copy flags
  if (impl.isPreEvalVisitor !== undefined) {
    visitor.isPreEvalVisitor = impl.isPreEvalVisitor;
  }
  if (impl.isReplacing !== undefined) {
    visitor.isReplacing = impl.isReplacing;
  }

  // Copy all visit* methods and other functions
  for (const key of Object.keys(impl)) {
    if (typeof impl[key] === 'function') {
      visitor[key] = impl[key].bind(visitor);
    }
  }

  return visitor;
}

module.exports = {
  Visitor,
  VisitorContext,
  createVisitor,
};
