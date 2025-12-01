package less_go

// UnicodeDescriptor represents a unicode descriptor node in the Less AST
type UnicodeDescriptor struct {
	*Node
	value any
}

func NewUnicodeDescriptor(value any) *UnicodeDescriptor {
	u := &UnicodeDescriptor{
		Node:  NewNode(),
		value: value,
	}
	u.Node.Value = value // Set the Node's Value field as well for consistency
	return u
}

func (u *UnicodeDescriptor) Type() string {
	return "UnicodeDescriptor"
}

func (u *UnicodeDescriptor) GetValue() any {
	return u.value
}

func (u *UnicodeDescriptor) SetValue(value any) {
	u.value = value
	u.Node.Value = value // Keep Node.Value in sync
}

// Accept implements the Visitor pattern
func (u *UnicodeDescriptor) Accept(visitor any) {
	if v, ok := visitor.(interface{ Visit(any) any }); ok {
		u.value = v.Visit(u.value)
		u.Node.Value = u.value // Keep Node.Value in sync
	}
}

func (u *UnicodeDescriptor) GenCSS(context any, output *CSSOutput) {
	if u.value != nil {
		output.Add(u.value, nil, nil)
	}
}

func (u *UnicodeDescriptor) ToCSS(context any) string {
	return u.Node.ToCSS(context)
}

// Eval returns the UnicodeDescriptor itself (matches JavaScript behavior)
func (u *UnicodeDescriptor) Eval() *UnicodeDescriptor {
	return u
}

func (u *UnicodeDescriptor) SetParent(nodes any, parent *Node) {
	u.Node.SetParent(nodes, parent)
}

func (u *UnicodeDescriptor) GetIndex() int {
	return u.Node.GetIndex()
}

func (u *UnicodeDescriptor) FileInfo() map[string]any {
	return u.Node.FileInfo()
}

func (u *UnicodeDescriptor) IsRulesetLike() bool {
	return u.Node.IsRulesetLike()
}

// Operate performs basic arithmetic operations (inherited from Node)
func (u *UnicodeDescriptor) Operate(context any, op string, a, b float64) float64 {
	return u.Node.Operate(context, op, a, b)
}

// Fround rounds numbers based on precision (inherited from Node)
func (u *UnicodeDescriptor) Fround(context any, value float64) float64 {
	return u.Node.Fround(context, value)
}

func (u *UnicodeDescriptor) BlocksVisibility() bool {
	return u.Node.BlocksVisibility()
}

func (u *UnicodeDescriptor) AddVisibilityBlock() {
	u.Node.AddVisibilityBlock()
}

func (u *UnicodeDescriptor) RemoveVisibilityBlock() {
	u.Node.RemoveVisibilityBlock()
}

func (u *UnicodeDescriptor) EnsureVisibility() {
	u.Node.EnsureVisibility()
}

func (u *UnicodeDescriptor) EnsureInvisibility() {
	u.Node.EnsureInvisibility()
}

func (u *UnicodeDescriptor) IsVisible() *bool {
	return u.Node.IsVisible()
}

func (u *UnicodeDescriptor) VisibilityInfo() map[string]any {
	return u.Node.VisibilityInfo()
}

func (u *UnicodeDescriptor) CopyVisibilityInfo(info map[string]any) {
	u.Node.CopyVisibilityInfo(info)
} 