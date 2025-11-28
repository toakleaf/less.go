/**
 * JavaScript Bindings for less.go Plugin System
 *
 * This module provides the JavaScript API for interacting with the Go LESS compiler
 * through the shared memory buffer system. It includes:
 *
 * - NodeFacade: Lazy deserialization of AST nodes from Go's FlatAST buffer
 * - Visitor: Base class for AST traversal and transformation
 * - Constructors: Functions to create new AST nodes
 * - BufferWriter: Serialize JavaScript nodes back to FlatAST format
 *
 * Usage in plugins:
 *   const { dimension, color, Visitor } = require('./bindings');
 *
 *   class MyVisitor extends Visitor {
 *     visitDimension(node) {
 *       return dimension(node.value * 2, node.unit);
 *     }
 *   }
 */

const { NodeFacade, createRootFacade, createFacadeAt, NodeTypeID, TypeNames, Flags } = require('./node-facade');
const { Visitor, VisitorContext, createVisitor } = require('./visitor');
const { BufferWriter, serializeToBuffer, FLAT_AST_MAGIC, FLAT_AST_VERSION } = require('./buffer-writer');
const constructors = require('./constructors');

// Re-export all constructors at top level for convenience
const {
  createNode,
  dimension,
  color,
  quoted,
  keyword,
  anonymous,
  url,
  variable,
  unit,
  value,
  expression,
  paren,
  negative,
  operation,
  condition,
  call,
  combinator,
  element,
  selector,
  ruleset,
  declaration,
  detachedruleset,
  atrule,
  assignment,
  attribute,
  comment,
  unicodeDescriptor,
} = constructors;

// Tree namespace for plugins that use less.tree.* syntax
const tree = {
  Anonymous: anonymous,
  Assignment: assignment,
  AtRule: atrule,
  Attribute: attribute,
  Call: call,
  Color: color,
  Combinator: combinator,
  Comment: comment,
  Condition: condition,
  Declaration: declaration,
  DetachedRuleset: detachedruleset,
  Dimension: dimension,
  Element: element,
  Expression: expression,
  Keyword: keyword,
  Negative: negative,
  Operation: operation,
  Paren: paren,
  Quoted: quoted,
  Ruleset: ruleset,
  Selector: selector,
  UnicodeDescriptor: unicodeDescriptor,
  Unit: unit,
  URL: url,
  Value: value,
  Variable: variable,
};

// Visitors namespace
const visitors = {
  Visitor,
  VisitorContext,
  createVisitor,
};

module.exports = {
  // Core classes
  NodeFacade,
  Visitor,
  VisitorContext,
  BufferWriter,

  // Factory functions
  createRootFacade,
  createFacadeAt,
  createVisitor,
  serializeToBuffer,

  // Node constructors - flat
  createNode,
  dimension,
  color,
  quoted,
  keyword,
  anonymous,
  url,
  variable,
  unit,
  value,
  expression,
  paren,
  negative,
  operation,
  condition,
  call,
  combinator,
  element,
  selector,
  ruleset,
  declaration,
  detachedruleset,
  atrule,
  assignment,
  attribute,
  comment,
  unicodeDescriptor,

  // Namespaces (for less.tree.* and less.visitors.* compatibility)
  tree,
  visitors,

  // Constants
  NodeTypeID,
  TypeNames,
  Flags,
  FLAT_AST_MAGIC,
  FLAT_AST_VERSION,
};
