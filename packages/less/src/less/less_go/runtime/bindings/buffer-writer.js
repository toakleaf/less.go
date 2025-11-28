/**
 * BufferWriter - Writes AST nodes to FlatAST binary format.
 *
 * This module allows JavaScript plugins to create new AST nodes that can be
 * serialized back to the shared memory buffer format that Go can read.
 *
 * The format matches Go's FlatAST structure exactly:
 * - Header: magic (4), version (4), nodeCount (4), rootIndex (4), offsets (3x4)
 * - Nodes: array of FlatNode structs (24 bytes each)
 * - StringTable: length-prefixed strings
 * - TypeTable: length-prefixed type names
 * - PropBuffer: JSON-encoded properties
 */

const { NodeTypeID, TypeNames, Flags, FLAT_NODE_SIZE } = require('./node-facade');

// FlatAST format constants - must match Go
const FLAT_AST_MAGIC = 0x4c455353; // "LESS"
const FLAT_AST_VERSION = 1;
const HEADER_SIZE = 28; // 7 x 4 bytes

/**
 * BufferWriter creates a FlatAST binary buffer from JavaScript node objects.
 */
class BufferWriter {
  constructor() {
    // Node storage
    this.nodes = [];
    this.stringTable = [];
    this.stringIndex = new Map(); // string -> index for deduplication
    this.typeTable = [];
    this.propBuffer = Buffer.alloc(0);

    // Current node index
    this._nextIndex = 0;
  }

  /**
   * Add a string to the string table with deduplication.
   *
   * @param {string} str - String to add
   * @returns {number} - Index in string table
   */
  addString(str) {
    if (str === undefined || str === null) {
      return 0;
    }
    str = String(str);

    if (this.stringIndex.has(str)) {
      return this.stringIndex.get(str);
    }

    const index = this.stringTable.length;
    this.stringTable.push(str);
    this.stringIndex.set(str, index);
    return index;
  }

  /**
   * Add properties to the property buffer.
   *
   * @param {Object} props - Properties object
   * @returns {{offset: number, length: number}}
   */
  addProperties(props) {
    if (!props || Object.keys(props).length === 0) {
      return { offset: 0, length: 0 };
    }

    const json = JSON.stringify(props);
    const propBytes = Buffer.from(json, 'utf8');
    const offset = this.propBuffer.length;

    this.propBuffer = Buffer.concat([this.propBuffer, propBytes]);

    return { offset, length: propBytes.length };
  }

  /**
   * Get type ID from type name.
   *
   * @param {string} typeName - Node type name
   * @returns {number} - Type ID
   */
  getTypeID(typeName) {
    // First check if it's in our TypeNames reverse lookup
    for (const [id, name] of Object.entries(TypeNames)) {
      if (name === typeName) {
        return parseInt(id, 10);
      }
    }
    return NodeTypeID.Unknown;
  }

  /**
   * Add a node from a JavaScript object representation.
   *
   * @param {Object} node - Node object with _type or type property
   * @param {number} parentIndex - Parent node index (0 for root)
   * @returns {number} - Index of the added node
   */
  addNode(node, parentIndex = 0) {
    if (!node) {
      return 0;
    }

    const typeName = node._type || node.type || 'Unknown';
    const typeID = this.getTypeID(typeName);

    // Build flags
    let flags = 0;
    if (node.parens) {
      flags |= Flags.Parens;
    }
    if (node.parensInOp) {
      flags |= Flags.ParensInOp;
    }
    if (node.nodeVisible !== undefined) {
      flags |= Flags.VisibleSet;
      if (node.nodeVisible) {
        flags |= Flags.Visible;
      } else {
        flags |= Flags.Invisible;
      }
    }

    // Extract properties (excluding special fields)
    const props = {};
    const skipFields = new Set([
      '_type',
      'type',
      'parens',
      'parensInOp',
      'nodeVisible',
      'children',
      'parent',
    ]);

    for (const [key, value] of Object.entries(node)) {
      if (skipFields.has(key)) continue;

      // Handle different value types
      if (typeof value === 'string') {
        props[key] = this.addString(value);
      } else if (typeof value === 'number' || typeof value === 'boolean') {
        props[key] = value;
      } else if (Array.isArray(value) && !value.some((v) => typeof v === 'object')) {
        // Primitive arrays (like rgb values)
        props[key] = value;
      }
      // Skip complex objects/nodes - they should be added as children
    }

    const { offset: propsOffset, length: propsLength } = this.addProperties(props);

    // Create the flat node
    const nodeIndex = this._nextIndex++;
    const flatNode = {
      typeID: typeID,
      flags: flags,
      childIndex: 0, // Will be set when children are added
      nextIndex: 0, // Will be set for siblings
      parentIndex: parentIndex,
      propsOffset: propsOffset,
      propsLength: propsLength,
    };

    this.nodes.push(flatNode);

    // Add children
    const children = node.children || node.value;
    if (Array.isArray(children)) {
      let prevChildIndex = 0;
      let firstChildIndex = 0;

      for (let i = 0; i < children.length; i++) {
        const child = children[i];
        if (child && typeof child === 'object' && (child._type || child.type)) {
          const childIndex = this.addNode(child, nodeIndex);
          if (childIndex > 0) {
            if (firstChildIndex === 0) {
              firstChildIndex = childIndex;
            }
            if (prevChildIndex > 0) {
              this.nodes[prevChildIndex].nextIndex = childIndex;
            }
            prevChildIndex = childIndex;
          }
        }
      }

      if (firstChildIndex > 0) {
        this.nodes[nodeIndex].childIndex = firstChildIndex;
      }
    }

    return nodeIndex;
  }

  /**
   * Build a complete FlatAST from a root node.
   *
   * @param {Object} root - Root node object
   * @returns {number} - Root index
   */
  buildFromRoot(root) {
    this.nodes = [];
    this.stringTable = [];
    this.stringIndex.clear();
    this.propBuffer = Buffer.alloc(0);
    this._nextIndex = 0;

    return this.addNode(root, 0);
  }

  /**
   * Calculate the size of the string table in bytes.
   *
   * @returns {number}
   */
  _stringTableSize() {
    let size = 4; // count
    for (const str of this.stringTable) {
      size += 4 + Buffer.byteLength(str, 'utf8');
    }
    return size;
  }

  /**
   * Calculate the size of the type table in bytes.
   *
   * @returns {number}
   */
  _typeTableSize() {
    let size = 4; // count
    for (const str of this.typeTable) {
      size += 4 + Buffer.byteLength(str, 'utf8');
    }
    return size;
  }

  /**
   * Serialize to a binary buffer.
   *
   * @param {number} rootIndex - Index of the root node
   * @returns {Buffer}
   */
  toBuffer(rootIndex = 0) {
    const nodesSize = this.nodes.length * FLAT_NODE_SIZE;
    const stringTableSize = this._stringTableSize();
    const typeTableSize = this._typeTableSize();
    const propBufferSize = 4 + this.propBuffer.length;

    const totalSize = HEADER_SIZE + nodesSize + stringTableSize + typeTableSize + propBufferSize;
    const buffer = Buffer.alloc(totalSize);

    let offset = 0;

    // Write header
    buffer.writeUInt32LE(FLAT_AST_MAGIC, offset);
    offset += 4;
    buffer.writeUInt32LE(FLAT_AST_VERSION, offset);
    offset += 4;
    buffer.writeUInt32LE(this.nodes.length, offset);
    offset += 4;
    buffer.writeUInt32LE(rootIndex, offset);
    offset += 4;

    // Calculate section offsets
    const nodesOffset = HEADER_SIZE;
    const stringTableOffset = nodesOffset + nodesSize;
    const typeTableOffset = stringTableOffset + stringTableSize;

    buffer.writeUInt32LE(nodesOffset, offset);
    offset += 4;
    buffer.writeUInt32LE(stringTableOffset, offset);
    offset += 4;
    buffer.writeUInt32LE(typeTableOffset, offset);
    offset += 4;

    // Write nodes
    offset = nodesOffset;
    for (const node of this.nodes) {
      buffer.writeUInt16LE(node.typeID, offset);
      offset += 2;
      buffer.writeUInt16LE(node.flags, offset);
      offset += 2;
      buffer.writeUInt32LE(node.childIndex, offset);
      offset += 4;
      buffer.writeUInt32LE(node.nextIndex, offset);
      offset += 4;
      buffer.writeUInt32LE(node.parentIndex, offset);
      offset += 4;
      buffer.writeUInt32LE(node.propsOffset, offset);
      offset += 4;
      buffer.writeUInt32LE(node.propsLength, offset);
      offset += 4;
    }

    // Write string table
    offset = stringTableOffset;
    buffer.writeUInt32LE(this.stringTable.length, offset);
    offset += 4;
    for (const str of this.stringTable) {
      const strBytes = Buffer.from(str, 'utf8');
      buffer.writeUInt32LE(strBytes.length, offset);
      offset += 4;
      strBytes.copy(buffer, offset);
      offset += strBytes.length;
    }

    // Write type table
    offset = typeTableOffset;
    buffer.writeUInt32LE(this.typeTable.length, offset);
    offset += 4;
    for (const str of this.typeTable) {
      const strBytes = Buffer.from(str, 'utf8');
      buffer.writeUInt32LE(strBytes.length, offset);
      offset += 4;
      strBytes.copy(buffer, offset);
      offset += strBytes.length;
    }

    // Write prop buffer
    buffer.writeUInt32LE(this.propBuffer.length, offset);
    offset += 4;
    this.propBuffer.copy(buffer, offset);

    return buffer;
  }

  /**
   * Clear the writer state for reuse.
   */
  clear() {
    this.nodes = [];
    this.stringTable = [];
    this.stringIndex.clear();
    this.typeTable = [];
    this.propBuffer = Buffer.alloc(0);
    this._nextIndex = 0;
  }
}

/**
 * Serialize a JavaScript node tree to FlatAST buffer.
 *
 * @param {Object} root - Root node object
 * @returns {Buffer}
 */
function serializeToBuffer(root) {
  const writer = new BufferWriter();
  const rootIndex = writer.buildFromRoot(root);
  return writer.toBuffer(rootIndex);
}

module.exports = {
  BufferWriter,
  serializeToBuffer,
  FLAT_AST_MAGIC,
  FLAT_AST_VERSION,
  HEADER_SIZE,
};
