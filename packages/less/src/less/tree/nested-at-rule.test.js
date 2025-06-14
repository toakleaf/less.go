import { describe, it, expect, beforeEach, vi } from 'vitest';
import NestableAtRulePrototype from './nested-at-rule';
import Ruleset from './ruleset';
import Value from './value';
import Selector from './selector';
import Anonymous from './anonymous';
import Expression from './expression';
import * as utils from '../utils';

// Mock the dependencies
vi.mock('./ruleset');
vi.mock('./value');
vi.mock('./selector');
vi.mock('./anonymous');
vi.mock('./expression');
vi.mock('../utils');

describe('NestableAtRulePrototype', () => {
    let atRule;
    let mockContext;
    let mockVisitor;

    beforeEach(() => {
        // Create a test object that implements the prototype
        atRule = Object.create(NestableAtRulePrototype);
        atRule.type = 'Media';
        atRule.features = null;
        atRule.rules = null;
        atRule._index = 0;
        atRule._fileInfo = { filename: 'test.less' };
        atRule.visibilityInfo = () => ({ visibilityBlocks: 0, nodeVisible: true });
        atRule.getIndex = () => atRule._index;
        atRule.fileInfo = () => atRule._fileInfo;
        atRule.setParent = vi.fn();
        atRule.copyVisibilityInfo = vi.fn();

        // Mock context
        mockContext = {
            mediaBlocks: [],
            mediaPath: []
        };

        // Mock visitor
        mockVisitor = {
            visit: vi.fn(),
            visitArray: vi.fn()
        };

        // Reset all mocks
        vi.clearAllMocks();
    });

    describe('isRulesetLike', () => {
        it('should return true', () => {
            expect(atRule.isRulesetLike()).toBe(true);
        });
    });

    describe('accept', () => {
        it('should visit features if they exist', () => {
            const mockFeatures = { type: 'Value' };
            const visitedFeatures = { type: 'Value', visited: true };
            
            atRule.features = mockFeatures;
            mockVisitor.visit.mockReturnValue(visitedFeatures);

            atRule.accept(mockVisitor);

            expect(mockVisitor.visit).toHaveBeenCalledWith(mockFeatures);
            expect(atRule.features).toBe(visitedFeatures);
        });

        it('should visit rules if they exist', () => {
            const mockRules = [{ type: 'Rule' }];
            const visitedRules = [{ type: 'Rule', visited: true }];
            
            atRule.rules = mockRules;
            mockVisitor.visitArray.mockReturnValue(visitedRules);

            atRule.accept(mockVisitor);

            expect(mockVisitor.visitArray).toHaveBeenCalledWith(mockRules);
            expect(atRule.rules).toBe(visitedRules);
        });

        it('should handle null features and rules', () => {
            atRule.features = null;
            atRule.rules = null;

            atRule.accept(mockVisitor);

            expect(mockVisitor.visit).not.toHaveBeenCalled();
            expect(mockVisitor.visitArray).not.toHaveBeenCalled();
        });

        it('should visit both features and rules when both exist', () => {
            const mockFeatures = { type: 'Value' };
            const mockRules = [{ type: 'Rule' }];
            const visitedFeatures = { type: 'Value', visited: true };
            const visitedRules = [{ type: 'Rule', visited: true }];
            
            atRule.features = mockFeatures;
            atRule.rules = mockRules;
            mockVisitor.visit.mockReturnValue(visitedFeatures);
            mockVisitor.visitArray.mockReturnValue(visitedRules);

            atRule.accept(mockVisitor);

            expect(mockVisitor.visit).toHaveBeenCalledWith(mockFeatures);
            expect(mockVisitor.visitArray).toHaveBeenCalledWith(visitedRules);
            expect(atRule.features).toBe(visitedFeatures);
            expect(atRule.rules).toBe(visitedRules);
        });
    });

    describe('evalTop', () => {
        it('should return itself when mediaBlocks length is 1 or less', () => {
            mockContext.mediaBlocks = [];
            const result = atRule.evalTop(mockContext);
            expect(result).toBe(atRule);
        });

        it('should return itself when mediaBlocks length is exactly 1', () => {
            mockContext.mediaBlocks = [{ type: 'Media' }];
            const result = atRule.evalTop(mockContext);
            expect(result).toBe(atRule);
        });

        it('should create new Ruleset when mediaBlocks length > 1', () => {
            const mockMediaBlocks = [
                { type: 'Media', name: 'block1' },
                { type: 'Media', name: 'block2' }
            ];
            mockContext.mediaBlocks = mockMediaBlocks;

            const mockSelectors = [{ type: 'Selector' }];
            const mockRuleset = {
                multiMedia: false,
                copyVisibilityInfo: vi.fn(),
                type: 'Ruleset'
            };

            // Mock Selector constructor and createEmptySelectors
            const MockSelector = vi.fn().mockImplementation(() => ({
                createEmptySelectors: () => mockSelectors
            }));
            vi.mocked(Selector).mockImplementation(MockSelector);

            // Mock Ruleset constructor
            vi.mocked(Ruleset).mockImplementation(() => mockRuleset);

            const result = atRule.evalTop(mockContext);

            expect(MockSelector).toHaveBeenCalledWith([], null, null, atRule._index, atRule._fileInfo);
            expect(Ruleset).toHaveBeenCalledWith(mockSelectors, mockMediaBlocks);
            expect(mockRuleset.multiMedia).toBe(true);
            expect(mockRuleset.copyVisibilityInfo).toHaveBeenCalledWith(atRule.visibilityInfo());
            expect(atRule.setParent).toHaveBeenCalledWith(mockRuleset, atRule);
            expect(result).toBe(mockRuleset);
        });

        it('should delete mediaBlocks and mediaPath from context', () => {
            mockContext.mediaBlocks = [];
            mockContext.mediaPath = [];

            atRule.evalTop(mockContext);

            expect(mockContext.mediaBlocks).toBeUndefined();
            expect(mockContext.mediaPath).toBeUndefined();
        });
    });

    describe('evalNested', () => {
        beforeEach(() => {
            mockContext.mediaPath = [];
        });

        it('should return this when path contains different type', () => {
            const otherTypeNode = { type: 'Import' };
            mockContext.mediaPath = [otherTypeNode];
            mockContext.mediaBlocks = [{ type: 'Media' }];

            const result = atRule.evalNested(mockContext);

            expect(mockContext.mediaBlocks).toEqual([]);
            expect(result).toBe(atRule);
        });

        it('should process features and create new Value with permutations', () => {
            const mockFeatures = {
                value: ['screen', 'print']
            };
            atRule.features = mockFeatures;
            atRule.type = 'Media';

            mockContext.mediaPath = [atRule];
            mockContext.mediaBlocks = [];

            const mockPermutationResult = [['screen'], ['print']];
            const mockValue = { type: 'Value' };
            const mockRuleset = { type: 'Ruleset' };

            // Mock permute method
            atRule.permute = vi.fn().mockReturnValue(mockPermutationResult);

            // Mock Anonymous constructor
            vi.mocked(Anonymous).mockImplementation((val) => ({ 
                type: 'Anonymous', 
                value: val,
                toCSS: null
            }));

            // Mock Expression constructor
            vi.mocked(Expression).mockImplementation((path) => ({ 
                type: 'Expression', 
                path 
            }));

            // Mock Value constructor
            vi.mocked(Value).mockImplementation(() => mockValue);

            // Mock Ruleset constructor
            vi.mocked(Ruleset).mockImplementation(() => mockRuleset);

            const result = atRule.evalNested(mockContext);

            expect(atRule.permute).toHaveBeenCalledWith([['screen', 'print']]);
            expect(Value).toHaveBeenCalled();
            expect(atRule.features).toBe(mockValue);
            expect(atRule.setParent).toHaveBeenCalledWith(mockValue, atRule);
            expect(Ruleset).toHaveBeenCalledWith([], []);
            expect(result).toBe(mockRuleset);
        });

        it('should handle features as Value object', () => {
            const mockValue = { value: ['screen'] };
            const mockFeatures = mockValue;
            atRule.features = mockFeatures;
            atRule.type = 'Media';

            mockContext.mediaPath = [atRule];
            mockContext.mediaBlocks = [];

            const mockPermutationResult = [['screen']];
            const mockResultValue = { type: 'Value' };
            const mockRuleset = { type: 'Ruleset' };

            atRule.permute = vi.fn().mockReturnValue(mockPermutationResult);
            vi.mocked(Value).mockImplementation(() => mockResultValue);
            vi.mocked(Ruleset).mockImplementation(() => mockRuleset);
            vi.mocked(Anonymous).mockImplementation((val) => ({ 
                type: 'Anonymous', 
                value: val,
                toCSS: null
            }));
            vi.mocked(Expression).mockImplementation((path) => ({ 
                type: 'Expression', 
                path 
            }));

            const result = atRule.evalNested(mockContext);

            expect(atRule.permute).toHaveBeenCalledWith([['screen']]);
            expect(result).toBe(mockRuleset);
        });

        it('should handle single feature value', () => {
            atRule.features = 'screen';
            atRule.type = 'Media';

            mockContext.mediaPath = [atRule];
            mockContext.mediaBlocks = [];

            const mockPermutationResult = [['screen']];
            const mockResultValue = { type: 'Value' };
            const mockRuleset = { type: 'Ruleset' };

            atRule.permute = vi.fn().mockReturnValue(mockPermutationResult);
            vi.mocked(Value).mockImplementation(() => mockResultValue);
            vi.mocked(Ruleset).mockImplementation(() => mockRuleset);
            vi.mocked(Anonymous).mockImplementation((val) => ({ 
                type: 'Anonymous', 
                value: val,
                toCSS: null
            }));
            vi.mocked(Expression).mockImplementation((path) => ({ 
                type: 'Expression', 
                path 
            }));

            const result = atRule.evalNested(mockContext);

            expect(atRule.permute).toHaveBeenCalledWith([['screen']]);
            expect(result).toBe(mockRuleset);
        });

        it('should insert "and" between path fragments', () => {
            atRule.features = 'screen';
            atRule.type = 'Media';

            mockContext.mediaPath = [atRule];
            mockContext.mediaBlocks = [];

            const mockPermutationResult = [['screen', 'color']];
            const mockResultValue = { type: 'Value' };
            const mockRuleset = { type: 'Ruleset' };
            const mockAndAnonymous = { type: 'Anonymous', value: 'and' };

            atRule.permute = vi.fn().mockReturnValue(mockPermutationResult);
            vi.mocked(Value).mockImplementation(() => mockResultValue);
            vi.mocked(Ruleset).mockImplementation(() => mockRuleset);
            
            vi.mocked(Anonymous).mockImplementation((val) => {
                if (val === 'and') {
                    return mockAndAnonymous;
                }
                return { 
                    type: 'Anonymous', 
                    value: val,
                    toCSS: null 
                };
            });

            vi.mocked(Expression).mockImplementation((path) => ({ 
                type: 'Expression', 
                path 
            }));

            const result = atRule.evalNested(mockContext);

            expect(result).toBe(mockRuleset);
            // Should have called Anonymous with 'and' for inserting between fragments
            expect(Anonymous).toHaveBeenCalledWith('and');
        });
    });

    describe('permute', () => {
        it('should return empty array for empty input', () => {
            const result = atRule.permute([]);
            expect(result).toEqual([]);
        });

        it('should return first element for single element array', () => {
            const input = [['a', 'b']];
            const result = atRule.permute(input);
            expect(result).toEqual(['a', 'b']);
        });

        it('should create permutations for multiple arrays', () => {
            const input = [['a', 'b'], ['c', 'd']];
            const result = atRule.permute(input);
            expect(result).toEqual([
                ['a', 'c'],
                ['a', 'd'],
                ['b', 'c'],
                ['b', 'd']
            ]);
        });

        it('should handle three arrays', () => {
            const input = [['a'], ['b', 'c'], ['d']];
            const result = atRule.permute(input);
            expect(result).toEqual([
                ['a', 'b', 'd'],
                ['a', 'c', 'd']
            ]);
        });

        it('should handle nested recursive calls', () => {
            const input = [['1', '2'], ['3'], ['4', '5']];
            const result = atRule.permute(input);
            expect(result).toEqual([
                ['1', '3', '4'],
                ['1', '3', '5'],
                ['2', '3', '4'],
                ['2', '3', '5']
            ]);
        });
    });

    describe('bubbleSelectors', () => {
        it('should return early if selectors is null', () => {
            atRule.rules = [{ type: 'Rule' }];
            atRule.bubbleSelectors(null);
            expect(atRule.rules).toEqual([{ type: 'Rule' }]);
        });

        it('should return early if selectors is undefined', () => {
            atRule.rules = [{ type: 'Rule' }];
            atRule.bubbleSelectors(undefined);
            expect(atRule.rules).toEqual([{ type: 'Rule' }]);
        });

        it('should create new ruleset with copied selectors and first rule', () => {
            const mockSelectors = [{ type: 'Selector' }];
            const mockFirstRule = { type: 'Rule' };
            const mockCopiedSelectors = [{ type: 'Selector', copied: true }];
            const mockRuleset = { type: 'Ruleset' };

            atRule.rules = [mockFirstRule];

            // Mock utils.copyArray
            vi.mocked(utils.copyArray).mockReturnValue(mockCopiedSelectors);

            // Mock Ruleset constructor
            vi.mocked(Ruleset).mockImplementation(() => mockRuleset);

            atRule.bubbleSelectors(mockSelectors);

            expect(utils.copyArray).toHaveBeenCalledWith(mockSelectors);
            expect(Ruleset).toHaveBeenCalledWith(mockCopiedSelectors, [mockFirstRule]);
            expect(atRule.rules).toEqual([mockRuleset]);
            expect(atRule.setParent).toHaveBeenCalledWith([mockRuleset], atRule);
        });

        it('should handle empty rules array', () => {
            const mockSelectors = [{ type: 'Selector' }];
            const mockCopiedSelectors = [{ type: 'Selector', copied: true }];
            const mockRuleset = { type: 'Ruleset' };

            atRule.rules = [];

            vi.mocked(utils.copyArray).mockReturnValue(mockCopiedSelectors);
            vi.mocked(Ruleset).mockImplementation(() => mockRuleset);

            atRule.bubbleSelectors(mockSelectors);

            expect(Ruleset).toHaveBeenCalledWith(mockCopiedSelectors, [undefined]);
            expect(atRule.rules).toEqual([mockRuleset]);
        });
    });

    describe('Integration tests', () => {
        it('should work with all methods in sequence', () => {
            const mockVisitor = {
                visit: vi.fn().mockReturnValue({ visited: true }),
                visitArray: vi.fn().mockReturnValue([{ visited: true }])
            };

            atRule.features = { type: 'Value' };
            atRule.rules = [{ type: 'Rule' }];

            // Test accept
            atRule.accept(mockVisitor);
            expect(atRule.features.visited).toBe(true);
            expect(atRule.rules[0].visited).toBe(true);

            // Test isRulesetLike
            expect(atRule.isRulesetLike()).toBe(true);

            // Test evalTop with single mediaBlock
            mockContext.mediaBlocks = [{ type: 'Media' }];
            const topResult = atRule.evalTop(mockContext);
            expect(topResult).toBe(atRule);

            // Test permute
            const permuteResult = atRule.permute([['a'], ['b']]);
            expect(permuteResult).toEqual([['a', 'b']]);
        });

        it('should handle complex nested evaluation scenario', () => {
            atRule.type = 'Media';
            atRule.features = { value: ['screen', 'print'] };
            
            const anotherRule = Object.create(NestableAtRulePrototype);
            anotherRule.type = 'Media';
            anotherRule.features = 'color';

            mockContext.mediaPath = [atRule, anotherRule];
            mockContext.mediaBlocks = [];

            const mockValue = { type: 'Value' };
            const mockRuleset = { type: 'Ruleset' };

            atRule.permute = vi.fn().mockReturnValue([['screen', 'color'], ['print', 'color']]);
            vi.mocked(Value).mockImplementation(() => mockValue);
            vi.mocked(Ruleset).mockImplementation(() => mockRuleset);
            vi.mocked(Anonymous).mockImplementation((val) => ({ 
                type: 'Anonymous', 
                value: val,
                toCSS: null
            }));
            vi.mocked(Expression).mockImplementation((path) => ({ 
                type: 'Expression', 
                path 
            }));

            const result = atRule.evalNested(mockContext);

            expect(atRule.permute).toHaveBeenCalled();
            expect(result).toBe(mockRuleset);
        });
    });

    describe('Error handling', () => {
        it('should handle null context in evalTop', () => {
            expect(() => atRule.evalTop(null)).toThrow();
        });

        it('should handle null context in evalNested', () => {
            expect(() => atRule.evalNested(null)).toThrow();
        });

        it('should handle malformed mediaPath in evalNested', () => {
            mockContext.mediaPath = null;
            expect(() => atRule.evalNested(mockContext)).toThrow();
        });

        it('should handle permute with null input', () => {
            expect(() => atRule.permute(null)).toThrow();
        });

        it('should handle bubbleSelectors with empty rules', () => {
            atRule.rules = null;
            const mockSelectors = [{ type: 'Selector' }];
            
            expect(() => atRule.bubbleSelectors(mockSelectors)).toThrow();
        });
    });
});