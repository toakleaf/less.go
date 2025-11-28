/**
 * Unit tests for NodeFacade
 */

const { NodeFacade, createRootFacade, NodeTypeID, TypeNames, Flags } = require('./node-facade');

// Mock AST data that matches FlatAST structure
function createMockAST() {
  return {
    version: 1,
    nodeCount: 3,
    rootIndex: 0,
    nodes: [
      // Node 0: Root Ruleset
      {
        typeID: NodeTypeID.Ruleset,
        flags: 0,
        childIndex: 1,
        nextIndex: 0,
        parentIndex: 0,
        propsOffset: 0,
        propsLength: 0,
      },
      // Node 1: Declaration (first child)
      {
        typeID: NodeTypeID.Declaration,
        flags: 0,
        childIndex: 2,
        nextIndex: 0,
        parentIndex: 0,
        propsOffset: 0,
        propsLength: 34,
      },
      // Node 2: Dimension (grandchild)
      {
        typeID: NodeTypeID.Dimension,
        flags: Flags.Parens,
        childIndex: 0,
        nextIndex: 0,
        parentIndex: 1,
        propsOffset: 34,
        propsLength: 27,
      },
    ],
    stringTable: ['color', '10', 'px'],
    typeTable: [],
    propBuffer: Buffer.from('{"name":"color","value":"red"}{"value":10,"unit":"px"}'),
  };
}

describe('NodeFacade', () => {
  let mockAST;

  beforeEach(() => {
    mockAST = createMockAST();
  });

  describe('constructor', () => {
    it('should create a facade for a node', () => {
      const facade = new NodeFacade(mockAST, 0);
      expect(facade).toBeInstanceOf(NodeFacade);
    });
  });

  describe('type property', () => {
    it('should return correct type name for Ruleset', () => {
      const facade = new NodeFacade(mockAST, 0);
      expect(facade.type).toBe('Ruleset');
    });

    it('should return correct type name for Declaration', () => {
      const facade = new NodeFacade(mockAST, 1);
      expect(facade.type).toBe('Declaration');
    });

    it('should return correct type name for Dimension', () => {
      const facade = new NodeFacade(mockAST, 2);
      expect(facade.type).toBe('Dimension');
    });
  });

  describe('flags', () => {
    it('should detect parens flag', () => {
      const facade = new NodeFacade(mockAST, 2);
      expect(facade.parens).toBe(true);
    });

    it('should not have parens when flag not set', () => {
      const facade = new NodeFacade(mockAST, 0);
      expect(facade.parens).toBe(false);
    });
  });

  describe('children', () => {
    it('should return children for parent node', () => {
      const facade = new NodeFacade(mockAST, 0);
      const children = facade.children;
      expect(children).toHaveLength(1);
      expect(children[0].type).toBe('Declaration');
    });

    it('should return empty array for leaf node', () => {
      const facade = new NodeFacade(mockAST, 2);
      expect(facade.children).toHaveLength(0);
    });

    it('should cache children', () => {
      const facade = new NodeFacade(mockAST, 0);
      const children1 = facade.children;
      const children2 = facade.children;
      expect(children1).toBe(children2);
    });
  });

  describe('parent', () => {
    it('should return null for root node', () => {
      const facade = new NodeFacade(mockAST, 0);
      // Root node has parentIndex pointing to itself
      expect(facade.parent).toBeNull();
    });

    it('should return parent facade for child node', () => {
      const facade = new NodeFacade(mockAST, 2);
      expect(facade.parent).not.toBeNull();
      expect(facade.parent.type).toBe('Declaration');
    });
  });

  describe('properties', () => {
    it('should parse properties from buffer', () => {
      const facade = new NodeFacade(mockAST, 1);
      const props = facade.properties;
      expect(props.name).toBe('color');
      expect(props.value).toBe('red');
    });

    it('should return empty object for node without props', () => {
      const facade = new NodeFacade(mockAST, 0);
      expect(facade.properties).toEqual({});
    });

    it('should cache properties', () => {
      const facade = new NodeFacade(mockAST, 1);
      const props1 = facade.properties;
      const props2 = facade.properties;
      expect(props1).toBe(props2);
    });
  });

  describe('hasChildren', () => {
    it('should return true for nodes with children', () => {
      const facade = new NodeFacade(mockAST, 0);
      expect(facade.hasChildren()).toBe(true);
    });

    it('should return false for leaf nodes', () => {
      const facade = new NodeFacade(mockAST, 2);
      expect(facade.hasChildren()).toBe(false);
    });
  });

  describe('isRulesetLike', () => {
    it('should return true for Ruleset', () => {
      const facade = new NodeFacade(mockAST, 0);
      expect(facade.isRulesetLike()).toBe(true);
    });

    it('should return false for Declaration', () => {
      const facade = new NodeFacade(mockAST, 1);
      expect(facade.isRulesetLike()).toBe(false);
    });
  });

  describe('toJSON', () => {
    it('should serialize to JSON object', () => {
      const facade = new NodeFacade(mockAST, 2);
      const json = facade.toJSON();

      expect(json.type).toBe('Dimension');
      expect(json.index).toBe(2);
      expect(json.parens).toBe(true);
    });
  });

  describe('toString', () => {
    it('should return string representation', () => {
      const facade = new NodeFacade(mockAST, 2);
      const str = facade.toString();
      expect(str).toContain('NodeFacade');
      expect(str).toContain('Dimension');
    });
  });
});

describe('createRootFacade', () => {
  it('should create facade for root node', () => {
    const mockAST = createMockAST();
    const root = createRootFacade(mockAST);

    expect(root).toBeInstanceOf(NodeFacade);
    expect(root.type).toBe('Ruleset');
    expect(root.index).toBe(0);
  });

  it('should throw for empty AST', () => {
    expect(() => createRootFacade({ nodes: [] })).toThrow('Invalid or empty AST');
  });

  it('should throw for null AST', () => {
    expect(() => createRootFacade(null)).toThrow();
  });
});

describe('NodeTypeID constants', () => {
  it('should have correct type IDs', () => {
    expect(NodeTypeID.Dimension).toBe(13);
    expect(NodeTypeID.Color).toBe(6);
    expect(NodeTypeID.Ruleset).toBe(30);
  });

  it('should have corresponding TypeNames', () => {
    expect(TypeNames[NodeTypeID.Dimension]).toBe('Dimension');
    expect(TypeNames[NodeTypeID.Color]).toBe('Color');
    expect(TypeNames[NodeTypeID.Ruleset]).toBe('Ruleset');
  });
});

describe('Flags constants', () => {
  it('should have correct flag values', () => {
    expect(Flags.Parens).toBe(1);
    expect(Flags.ParensInOp).toBe(2);
    expect(Flags.Visible).toBe(4);
  });
});
