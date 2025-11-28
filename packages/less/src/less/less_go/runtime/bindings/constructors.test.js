/**
 * Unit tests for Constructors
 */

const constructors = require('./constructors');

describe('Node Constructors', () => {
  describe('createNode', () => {
    it('should create a node with type', () => {
      const node = constructors.createNode('TestType', { value: 'test' });
      expect(node._type).toBe('TestType');
      expect(node.value).toBe('test');
    });

    it('should create a node without extra props', () => {
      const node = constructors.createNode('TestType');
      expect(node._type).toBe('TestType');
    });
  });

  describe('dimension', () => {
    it('should create dimension with value and unit', () => {
      const dim = constructors.dimension(10, 'px');
      expect(dim._type).toBe('Dimension');
      expect(dim.value).toBe(10);
      expect(dim.unit).toBe('px');
    });

    it('should create dimension without unit', () => {
      const dim = constructors.dimension(10);
      expect(dim._type).toBe('Dimension');
      expect(dim.value).toBe(10);
      expect(dim.unit).toBe('');
    });

    it('should parse string value', () => {
      const dim = constructors.dimension('10.5', 'em');
      expect(dim.value).toBe(10.5);
    });
  });

  describe('color', () => {
    it('should create color with rgb array', () => {
      const col = constructors.color([255, 128, 0]);
      expect(col._type).toBe('Color');
      expect(col.rgb).toEqual([255, 128, 0]);
      expect(col.alpha).toBe(1);
    });

    it('should create color with rgb object', () => {
      const col = constructors.color({ r: 255, g: 128, b: 0 });
      expect(col.rgb).toEqual([255, 128, 0]);
    });

    it('should create color with alpha', () => {
      const col = constructors.color([255, 128, 0], 0.5);
      expect(col.alpha).toBe(0.5);
    });
  });

  describe('quoted', () => {
    it('should create quoted string', () => {
      const q = constructors.quoted('"', 'hello');
      expect(q._type).toBe('Quoted');
      expect(q.quote).toBe('"');
      expect(q.value).toBe('hello');
      expect(q.escaped).toBe(false);
    });

    it('should create escaped quoted', () => {
      const q = constructors.quoted('"', 'hello', true);
      expect(q.escaped).toBe(true);
    });

    it('should use default quote', () => {
      const q = constructors.quoted(null, 'hello');
      expect(q.quote).toBe('"');
    });
  });

  describe('keyword', () => {
    it('should create keyword', () => {
      const k = constructors.keyword('auto');
      expect(k._type).toBe('Keyword');
      expect(k.value).toBe('auto');
    });
  });

  describe('anonymous', () => {
    it('should create anonymous value', () => {
      const a = constructors.anonymous('raw css');
      expect(a._type).toBe('Anonymous');
      expect(a.value).toBe('raw css');
    });
  });

  describe('url', () => {
    it('should create URL from string', () => {
      const u = constructors.url('image.png');
      expect(u._type).toBe('URL');
      expect(u.value._type).toBe('Quoted');
      expect(u.value.value).toBe('image.png');
    });

    it('should create URL from node', () => {
      const quoted = constructors.quoted('"', 'image.png');
      const u = constructors.url(quoted);
      expect(u._type).toBe('URL');
      expect(u.value).toBe(quoted);
    });
  });

  describe('variable', () => {
    it('should create variable', () => {
      const v = constructors.variable('@color');
      expect(v._type).toBe('Variable');
      expect(v.name).toBe('@color');
    });
  });

  describe('unit', () => {
    it('should create unit with numerator and denominator', () => {
      const u = constructors.unit(['px'], ['s']);
      expect(u._type).toBe('Unit');
      expect(u.numerator).toEqual(['px']);
      expect(u.denominator).toEqual(['s']);
    });

    it('should create unit with defaults', () => {
      const u = constructors.unit();
      expect(u.numerator).toEqual([]);
      expect(u.denominator).toEqual([]);
    });
  });

  describe('value', () => {
    it('should create value with array', () => {
      const v = constructors.value([constructors.dimension(10, 'px')]);
      expect(v._type).toBe('Value');
      expect(v.children).toHaveLength(1);
    });

    it('should wrap single item in array', () => {
      const v = constructors.value(constructors.dimension(10, 'px'));
      expect(v.children).toHaveLength(1);
    });
  });

  describe('expression', () => {
    it('should create expression', () => {
      const e = constructors.expression([
        constructors.dimension(10, 'px'),
        constructors.dimension(20, 'px'),
      ]);
      expect(e._type).toBe('Expression');
      expect(e.children).toHaveLength(2);
    });
  });

  describe('paren', () => {
    it('should create parenthesized node', () => {
      const inner = constructors.dimension(10, 'px');
      const p = constructors.paren(inner);
      expect(p._type).toBe('Paren');
      expect(p.value).toBe(inner);
      expect(p.children).toContain(inner);
    });
  });

  describe('negative', () => {
    it('should create negative node', () => {
      const inner = constructors.dimension(10, 'px');
      const n = constructors.negative(inner);
      expect(n._type).toBe('Negative');
      expect(n.value).toBe(inner);
    });
  });

  describe('operation', () => {
    it('should create operation', () => {
      const left = constructors.dimension(10, 'px');
      const right = constructors.dimension(5, 'px');
      const op = constructors.operation('+', [left, right]);

      expect(op._type).toBe('Operation');
      expect(op.op).toBe('+');
      expect(op.operands).toHaveLength(2);
      expect(op.children).toHaveLength(2);
    });
  });

  describe('condition', () => {
    it('should create condition', () => {
      const left = constructors.dimension(10, 'px');
      const right = constructors.dimension(5, 'px');
      const cond = constructors.condition('>', left, right);

      expect(cond._type).toBe('Condition');
      expect(cond.op).toBe('>');
      expect(cond.lvalue).toBe(left);
      expect(cond.rvalue).toBe(right);
      expect(cond.negate).toBe(false);
    });

    it('should create negated condition', () => {
      const cond = constructors.condition('=', {}, {}, true);
      expect(cond.negate).toBe(true);
    });
  });

  describe('call', () => {
    it('should create function call', () => {
      const args = [constructors.dimension(10, 'px')];
      const c = constructors.call('rgb', args);

      expect(c._type).toBe('Call');
      expect(c.name).toBe('rgb');
      expect(c.args).toEqual(args);
      expect(c.children).toEqual(args);
    });

    it('should create call without args', () => {
      const c = constructors.call('test');
      expect(c.args).toEqual([]);
    });
  });

  describe('combinator', () => {
    it('should create combinator', () => {
      const c = constructors.combinator('>');
      expect(c._type).toBe('Combinator');
      expect(c.value).toBe('>');
    });
  });

  describe('element', () => {
    it('should create element with combinator', () => {
      const e = constructors.element('>', 'div');
      expect(e._type).toBe('Element');
      expect(e.combinator._type).toBe('Combinator');
      expect(e.value).toBe('div');
    });

    it('should create element with combinator node', () => {
      const comb = constructors.combinator(' ');
      const e = constructors.element(comb, 'span');
      expect(e.combinator).toBe(comb);
    });
  });

  describe('selector', () => {
    it('should create selector with elements', () => {
      const elem1 = constructors.element('', 'div');
      const elem2 = constructors.element(' ', 'span');
      const s = constructors.selector([elem1, elem2]);

      expect(s._type).toBe('Selector');
      expect(s.elements).toHaveLength(2);
      expect(s.children).toHaveLength(2);
    });
  });

  describe('ruleset', () => {
    it('should create ruleset', () => {
      const sel = constructors.selector([constructors.element('', 'div')]);
      const decl = constructors.declaration('color', constructors.keyword('red'));
      const rs = constructors.ruleset([sel], [decl]);

      expect(rs._type).toBe('Ruleset');
      expect(rs.selectors).toHaveLength(1);
      expect(rs.rules).toHaveLength(1);
    });
  });

  describe('declaration', () => {
    it('should create declaration', () => {
      const val = constructors.keyword('red');
      const d = constructors.declaration('color', val);

      expect(d._type).toBe('Declaration');
      expect(d.name).toBe('color');
      expect(d.value).toBe(val);
      expect(d.important).toBe('');
    });

    it('should create declaration with important', () => {
      const d = constructors.declaration('color', {}, '!important');
      expect(d.important).toBe('!important');
    });
  });

  describe('detachedruleset', () => {
    it('should create detached ruleset', () => {
      const rs = constructors.ruleset([], []);
      const dr = constructors.detachedruleset(rs);

      expect(dr._type).toBe('DetachedRuleset');
      expect(dr.ruleset).toBe(rs);
    });
  });

  describe('atrule', () => {
    it('should create at-rule', () => {
      const ar = constructors.atrule('@media', constructors.keyword('screen'));

      expect(ar._type).toBe('AtRule');
      expect(ar.name).toBe('@media');
    });
  });

  describe('assignment', () => {
    it('should create assignment', () => {
      const a = constructors.assignment('key', 'value');
      expect(a._type).toBe('Assignment');
      expect(a.key).toBe('key');
      expect(a.value).toBe('value');
    });
  });

  describe('attribute', () => {
    it('should create attribute', () => {
      const a = constructors.attribute('href', '=', 'url');
      expect(a._type).toBe('Attribute');
      expect(a.key).toBe('href');
      expect(a.op).toBe('=');
      expect(a.value).toBe('url');
    });
  });

  describe('comment', () => {
    it('should create block comment', () => {
      const c = constructors.comment('comment text');
      expect(c._type).toBe('Comment');
      expect(c.value).toBe('comment text');
      expect(c.isLineComment).toBe(false);
    });

    it('should create line comment', () => {
      const c = constructors.comment('comment text', true);
      expect(c.isLineComment).toBe(true);
    });
  });

  describe('unicodeDescriptor', () => {
    it('should create unicode descriptor', () => {
      const ud = constructors.unicodeDescriptor('U+0-7F');
      expect(ud._type).toBe('UnicodeDescriptor');
      expect(ud.value).toBe('U+0-7F');
    });
  });
});
