/**
 * Unit tests for BufferWriter
 */

const { BufferWriter, serializeToBuffer, FLAT_AST_MAGIC, FLAT_AST_VERSION } = require('./buffer-writer');
const constructors = require('./constructors');

describe('BufferWriter', () => {
  let writer;

  beforeEach(() => {
    writer = new BufferWriter();
  });

  describe('addString', () => {
    it('should add string to table', () => {
      const idx = writer.addString('hello');
      expect(idx).toBe(0);
      expect(writer.stringTable[0]).toBe('hello');
    });

    it('should deduplicate strings', () => {
      const idx1 = writer.addString('hello');
      const idx2 = writer.addString('hello');
      expect(idx1).toBe(idx2);
      expect(writer.stringTable.length).toBe(1);
    });

    it('should handle null/undefined', () => {
      const idx = writer.addString(null);
      expect(idx).toBe(0);
    });
  });

  describe('addProperties', () => {
    it('should add properties to buffer', () => {
      const { offset, length } = writer.addProperties({ key: 'value' });
      expect(offset).toBe(0);
      expect(length).toBeGreaterThan(0);
    });

    it('should return zero length for empty props', () => {
      const { offset, length } = writer.addProperties({});
      expect(length).toBe(0);
    });

    it('should return zero length for null props', () => {
      const { offset, length } = writer.addProperties(null);
      expect(length).toBe(0);
    });
  });

  describe('getTypeID', () => {
    it('should return correct type ID for Dimension', () => {
      expect(writer.getTypeID('Dimension')).toBe(13);
    });

    it('should return correct type ID for Color', () => {
      expect(writer.getTypeID('Color')).toBe(6);
    });

    it('should return Unknown for invalid type', () => {
      expect(writer.getTypeID('InvalidType')).toBe(0);
    });
  });

  describe('addNode', () => {
    it('should add a simple node', () => {
      const idx = writer.addNode(constructors.dimension(10, 'px'));
      expect(idx).toBe(0);
      expect(writer.nodes.length).toBe(1);
      expect(writer.nodes[0].typeID).toBe(13); // Dimension
    });

    it('should add node with children', () => {
      const parent = constructors.value([
        constructors.dimension(10, 'px'),
        constructors.dimension(20, 'em'),
      ]);

      writer.addNode(parent);

      expect(writer.nodes.length).toBe(3); // parent + 2 children
    });

    it('should link children correctly', () => {
      const parent = constructors.value([
        constructors.dimension(10, 'px'),
        constructors.dimension(20, 'em'),
      ]);

      writer.addNode(parent);

      // Parent should have childIndex pointing to first child
      expect(writer.nodes[0].childIndex).toBe(1);

      // First child should have nextIndex pointing to second child
      expect(writer.nodes[1].nextIndex).toBe(2);

      // Second child should have no next sibling
      expect(writer.nodes[2].nextIndex).toBe(0);
    });

    it('should set parent index for children', () => {
      const parent = constructors.value([constructors.dimension(10, 'px')]);

      writer.addNode(parent);

      expect(writer.nodes[1].parentIndex).toBe(0);
    });

    it('should handle flags', () => {
      const node = constructors.createNode('Dimension', {
        value: 10,
        parens: true,
        parensInOp: true,
      });

      writer.addNode(node);

      expect(writer.nodes[0].flags & 1).toBe(1); // Parens
      expect(writer.nodes[0].flags & 2).toBe(2); // ParensInOp
    });
  });

  describe('buildFromRoot', () => {
    it('should reset and build from root', () => {
      writer.addString('old');
      writer.addNode(constructors.keyword('old'));

      const rootIdx = writer.buildFromRoot(constructors.dimension(10, 'px'));

      expect(rootIdx).toBe(0);
      expect(writer.nodes.length).toBe(1);
    });
  });

  describe('toBuffer', () => {
    it('should create valid buffer', () => {
      writer.addNode(constructors.dimension(10, 'px'));
      const buffer = writer.toBuffer(0);

      expect(Buffer.isBuffer(buffer)).toBe(true);
      expect(buffer.length).toBeGreaterThan(28); // At least header size
    });

    it('should write correct magic number', () => {
      writer.addNode(constructors.dimension(10, 'px'));
      const buffer = writer.toBuffer(0);

      const magic = buffer.readUInt32LE(0);
      expect(magic).toBe(FLAT_AST_MAGIC);
    });

    it('should write correct version', () => {
      writer.addNode(constructors.dimension(10, 'px'));
      const buffer = writer.toBuffer(0);

      const version = buffer.readUInt32LE(4);
      expect(version).toBe(FLAT_AST_VERSION);
    });

    it('should write correct node count', () => {
      writer.addNode(constructors.value([
        constructors.dimension(10, 'px'),
        constructors.dimension(20, 'em'),
      ]));
      const buffer = writer.toBuffer(0);

      const nodeCount = buffer.readUInt32LE(8);
      expect(nodeCount).toBe(3);
    });

    it('should write correct root index', () => {
      writer.addNode(constructors.dimension(10, 'px'));
      const buffer = writer.toBuffer(0);

      const rootIndex = buffer.readUInt32LE(12);
      expect(rootIndex).toBe(0);
    });
  });

  describe('clear', () => {
    it('should reset all state', () => {
      writer.addString('test');
      writer.addNode(constructors.dimension(10, 'px'));

      writer.clear();

      expect(writer.nodes.length).toBe(0);
      expect(writer.stringTable.length).toBe(0);
      expect(writer.propBuffer.length).toBe(0);
    });
  });
});

describe('serializeToBuffer', () => {
  it('should serialize a node tree to buffer', () => {
    const root = constructors.ruleset(
      [constructors.selector([constructors.element('', 'div')])],
      [constructors.declaration('color', constructors.keyword('red'))]
    );

    const buffer = serializeToBuffer(root);

    expect(Buffer.isBuffer(buffer)).toBe(true);

    // Verify header
    expect(buffer.readUInt32LE(0)).toBe(FLAT_AST_MAGIC);
    expect(buffer.readUInt32LE(4)).toBe(FLAT_AST_VERSION);
  });

  it('should serialize simple node', () => {
    const buffer = serializeToBuffer(constructors.dimension(10, 'px'));
    expect(buffer.length).toBeGreaterThan(28);
  });

  it('should serialize complex tree', () => {
    const tree = constructors.ruleset(
      [],
      [
        constructors.declaration('width', constructors.dimension(100, '%')),
        constructors.declaration('height', constructors.dimension(50, 'vh')),
        constructors.declaration(
          'margin',
          constructors.expression([
            constructors.dimension(10, 'px'),
            constructors.dimension(20, 'px'),
          ])
        ),
      ]
    );

    const buffer = serializeToBuffer(tree);
    expect(buffer.length).toBeGreaterThan(100);
  });
});

describe('Buffer format compatibility', () => {
  it('should produce buffer that can be parsed back', () => {
    const { parseFlatAST } = require('../plugin-host');

    const original = constructors.value([
      constructors.dimension(10, 'px'),
      constructors.keyword('auto'),
    ]);

    const buffer = serializeToBuffer(original);
    const parsed = parseFlatAST(buffer);

    expect(parsed.version).toBe(FLAT_AST_VERSION);
    expect(parsed.nodeCount).toBe(3);
    expect(parsed.rootIndex).toBe(0);
  });

  it('should preserve node types through roundtrip', () => {
    const { parseFlatAST } = require('../plugin-host');
    const { NodeTypeID } = require('./node-facade');

    const buffer = serializeToBuffer(constructors.dimension(10, 'px'));
    const parsed = parseFlatAST(buffer);

    expect(parsed.nodes[0].typeID).toBe(NodeTypeID.Dimension);
  });

  it('should preserve parent-child relationships', () => {
    const { parseFlatAST } = require('../plugin-host');

    const parent = constructors.expression([
      constructors.dimension(10, 'px'),
      constructors.dimension(20, 'px'),
    ]);

    const buffer = serializeToBuffer(parent);
    const parsed = parseFlatAST(buffer);

    // Parent node should have childIndex pointing to first child
    expect(parsed.nodes[0].childIndex).toBe(1);

    // Children should have parentIndex pointing to parent
    expect(parsed.nodes[1].parentIndex).toBe(0);
    expect(parsed.nodes[2].parentIndex).toBe(0);
  });
});
