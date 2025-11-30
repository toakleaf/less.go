/**
 * Unit tests for Visitor
 */

const { Visitor, VisitorContext, createVisitor } = require('./visitor');
const { NodeTypeID, Flags } = require('./node-facade');

// Mock AST data
function createMockAST() {
    return {
        version: 1,
        nodeCount: 4,
        rootIndex: 0,
        nodes: [
            // Node 0: Root Ruleset
            { typeID: NodeTypeID.Ruleset, flags: 0, childIndex: 1, nextIndex: 0, parentIndex: 0, propsOffset: 0, propsLength: 0 },
            // Node 1: Declaration (first child)
            { typeID: NodeTypeID.Declaration, flags: 0, childIndex: 3, nextIndex: 2, parentIndex: 0, propsOffset: 0, propsLength: 0 },
            // Node 2: Declaration (second child)
            { typeID: NodeTypeID.Declaration, flags: 0, childIndex: 0, nextIndex: 0, parentIndex: 0, propsOffset: 0, propsLength: 0 },
            // Node 3: Dimension (grandchild of node 1)
            { typeID: NodeTypeID.Dimension, flags: Flags.Parens, childIndex: 0, nextIndex: 0, parentIndex: 1, propsOffset: 0, propsLength: 27 },
        ],
        stringTable: ['10', 'px'],
        typeTable: [],
        propBuffer: Buffer.from('{"value":10,"unit":"px"}'),
    };
}

describe('Visitor', () => {
    describe('constructor', () => {
        it('should create a visitor with default options', () => {
            const visitor = new Visitor();
            expect(visitor.isPreEvalVisitor).toBe(false);
            expect(visitor.isReplacing).toBe(false);
        });
    });

    describe('visit', () => {
        it('should return node unchanged if no visit method exists', () => {
            const visitor = new Visitor();
            const node = { type: 'UnknownType', value: 'test' };
            const result = visitor.visit(node);
            expect(result).toBe(node);
        });

        it('should call visit method for specific type', () => {
            const visitor = new Visitor();
            const visited = [];

            visitor.visitDimension = function (node) {
                visited.push(node);
                return node;
            };

            const node = { type: 'Dimension', value: 10 };
            visitor.visit(node);

            expect(visited).toHaveLength(1);
            expect(visited[0]).toBe(node);
        });

        it('should return replacement from visit method', () => {
            const visitor = new Visitor();
            const replacement = { type: 'Dimension', value: 20 };

            visitor.visitDimension = function () {
                return replacement;
            };

            const node = { type: 'Dimension', value: 10 };
            const result = visitor.visit(node);

            expect(result).toBe(replacement);
        });

        it('should handle _type property', () => {
            const visitor = new Visitor();
            const visited = [];

            visitor.visitDimension = function (node) {
                visited.push(node);
                return node;
            };

            const node = { _type: 'Dimension', value: 10 };
            visitor.visit(node);

            expect(visited).toHaveLength(1);
        });

        it('should return null/undefined nodes unchanged', () => {
            const visitor = new Visitor();
            expect(visitor.visit(null)).toBeNull();
            expect(visitor.visit(undefined)).toBeUndefined();
        });
    });

    describe('visitChildren', () => {
        it('should visit all children', () => {
            const visitor = new Visitor();
            const visited = [];

            visitor.visitDimension = function (node) {
                visited.push(node);
                return node;
            };

            const parent = {
                type: 'Value',
                children: [
                    { type: 'Dimension', value: 1 },
                    { type: 'Dimension', value: 2 },
                ],
            };

            visitor.visitChildren(parent);

            expect(visited).toHaveLength(2);
        });

        it('should handle empty children', () => {
            const visitor = new Visitor();
            const parent = { type: 'Value', children: [] };

            expect(() => visitor.visitChildren(parent)).not.toThrow();
        });
    });

    describe('run', () => {
        it('should visit root node', () => {
            const visitor = new Visitor();
            let visitedRoot = null;

            visitor.visitRuleset = function (node) {
                visitedRoot = node;
                return node;
            };

            const root = { type: 'Ruleset', children: [] };
            visitor.run(root);

            expect(visitedRoot).toBe(root);
        });

        it('should return null for null input', () => {
            const visitor = new Visitor();
            expect(visitor.run(null)).toBeNull();
        });
    });

    describe('visitArray', () => {
        it('should visit all nodes in array', () => {
            const visitor = new Visitor();
            const nodes = [
                { type: 'Dimension', value: 1 },
                { type: 'Dimension', value: 2 },
            ];

            const result = visitor.visitArray(nodes);

            expect(result).toHaveLength(2);
        });

        it('should filter out null results', () => {
            const visitor = new Visitor();
            visitor.visitDimension = function (node) {
                if (node.value === 1) return null;
                return node;
            };

            const nodes = [
                { type: 'Dimension', value: 1 },
                { type: 'Dimension', value: 2 },
            ];

            const result = visitor.visitArray(nodes);

            expect(result).toHaveLength(1);
            expect(result[0].value).toBe(2);
        });

        it('should flatten array results', () => {
            const visitor = new Visitor();
            visitor.visitDimension = function (node) {
                return [node, { type: 'Dimension', value: node.value + 10 }];
            };

            const nodes = [{ type: 'Dimension', value: 1 }];
            const result = visitor.visitArray(nodes);

            expect(result).toHaveLength(2);
        });
    });

    describe('replacements', () => {
        it('should track replacements for replacing visitor', () => {
            const visitor = new Visitor();
            visitor.isReplacing = true;

            const replacement = { type: 'Dimension', value: 20 };
            visitor._storeReplacement({ index: 1 }, 0, replacement);

            const replacements = visitor.getReplacements();
            expect(replacements).toHaveLength(1);
            expect(replacements[0].parentIndex).toBe(1);
            expect(replacements[0].replacement).toBe(replacement);
        });

        it('should clear replacements', () => {
            const visitor = new Visitor();
            visitor.isReplacing = true;

            visitor._storeReplacement({ index: 1 }, 0, {});
            visitor.clearReplacements();

            expect(visitor.getReplacements()).toHaveLength(0);
        });
    });
});

describe('createVisitor', () => {
    it('should create visitor from plain object', () => {
        const visitor = createVisitor({
            isPreEvalVisitor: true,
            isReplacing: true,
            visitDimension(node) {
                return node;
            },
        });

        expect(visitor).toBeInstanceOf(Visitor);
        expect(visitor.isPreEvalVisitor).toBe(true);
        expect(visitor.isReplacing).toBe(true);
        expect(typeof visitor.visitDimension).toBe('function');
    });

    it('should bind methods to visitor', () => {
        let capturedThis = null;

        const visitor = createVisitor({
            visitDimension(node) {
                capturedThis = this;
                return node;
            },
        });

        visitor.visitDimension({});

        expect(capturedThis).toBe(visitor);
    });
});

describe('VisitorContext', () => {
    let mockAST;

    beforeEach(() => {
        mockAST = createMockAST();
    });

    describe('addVisitor', () => {
        it('should add visitor to list', () => {
            const ctx = new VisitorContext(mockAST);
            const visitor = new Visitor();

            ctx.addVisitor(visitor);

            expect(ctx._visitors).toHaveLength(1);
        });

        it('should categorize pre-eval visitors', () => {
            const ctx = new VisitorContext(mockAST);

            const preEval = new Visitor();
            preEval.isPreEvalVisitor = true;

            const postEval = new Visitor();
            postEval.isPreEvalVisitor = false;

            ctx.addVisitor(preEval);
            ctx.addVisitor(postEval);

            expect(ctx._preEvalVisitors).toHaveLength(1);
            expect(ctx._postEvalVisitors).toHaveLength(1);
        });
    });

    describe('runPreEvalVisitors', () => {
        it('should run all pre-eval visitors', () => {
            const ctx = new VisitorContext(mockAST);
            const visited = [];

            const visitor = new Visitor();
            visitor.isPreEvalVisitor = true;
            visitor.visitRuleset = function (node) {
                visited.push(node.type);
                return node;
            };

            ctx.addVisitor(visitor);
            const result = ctx.runPreEvalVisitors();

            expect(result.success).toBe(true);
            expect(result.visitorCount).toBe(1);
        });
    });

    describe('runPostEvalVisitors', () => {
        it('should run all post-eval visitors', () => {
            const ctx = new VisitorContext(mockAST);

            const visitor = new Visitor();
            visitor.isPreEvalVisitor = false;

            ctx.addVisitor(visitor);
            const result = ctx.runPostEvalVisitors();

            expect(result.success).toBe(true);
            expect(result.visitorCount).toBe(1);
        });
    });

    describe('getVisitorInfo', () => {
        it('should return visitor metadata', () => {
            const ctx = new VisitorContext(mockAST);

            const visitor = new Visitor();
            visitor.isPreEvalVisitor = true;
            visitor.isReplacing = true;

            ctx.addVisitor(visitor);

            const info = ctx.getVisitorInfo();
            expect(info).toHaveLength(1);
            expect(info[0].index).toBe(0);
            expect(info[0].isPreEvalVisitor).toBe(true);
            expect(info[0].isReplacing).toBe(true);
        });
    });
});
