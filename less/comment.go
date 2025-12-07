package less_go

type Comment struct {
	*Node
	Value        string
	IsLineComment bool
	AllowRoot    bool
	DebugInfo    map[string]any
}

func NewComment(value string, isLineComment bool, index int, currentFileInfo map[string]any) *Comment {
	node := NewNode()
	node.TypeIndex = GetTypeIndexForNodeType("Comment")

	comment := &Comment{
		Node:          node,
		Value:         value,
		IsLineComment: isLineComment,
	}
	comment.Index = index
	comment.SetFileInfo(currentFileInfo)
	comment.AllowRoot = true
	return comment
}

func (c *Comment) GetType() string {
	return "Comment"
}

func (c *Comment) Accept(visitor any) {
	if v, ok := visitor.(interface{ VisitComment(any, any) any }); ok {
		v.VisitComment(c, nil)
	}
}

func (c *Comment) GenCSS(context any, output *CSSOutput) {
	if c.Node != nil && c.Node.BlocksVisibility() {
		nodeVisible := c.Node.IsVisible()
		if nodeVisible == nil || !*nodeVisible {
			return
		}
	}

	if c.DebugInfo != nil {
		var ctx map[string]any
		if ctxMap, ok := context.(map[string]any); ok {
			ctx = ctxMap
		}
		if ctx != nil {
			output.Add(DebugInfo(ctx, c, ""), c.FileInfo(), c.GetIndex())
		}
	}
	output.Add(c.Value, nil, nil)
}

func (c *Comment) IsSilent(context any) bool {
	var compress bool
	if ctxMap, ok := context.(map[string]any); ok {
		if compressVal, exists := ctxMap["compress"]; exists {
			compress, _ = compressVal.(bool)
		}
	}

	isCompressed := compress && len(c.Value) > 2 && c.Value[2] != '!'
	return c.IsLineComment || isCompressed
}

func (c *Comment) SetParent(node any, parent *Node) {
	if parent != nil {
		c.Parent = parent
	}
}

func (c *Comment) GetDebugInfo() map[string]any {
	return c.DebugInfo
}

func (c *Comment) Eval(context any) any {
	return c
}

func (c *Comment) IsVisible() bool {
	return true
}

func DebugInfo(context map[string]any, node any, separator string) string {
	if context == nil || node == nil {
		return ""
	}

	dumpLineNumbers, ok := context["dumpLineNumbers"].(string)
	if !ok || dumpLineNumbers == "" {
		return ""
	}

	compress, _ := context["compress"].(bool)
	if compress && dumpLineNumbers != "all" {
		return ""
	}

	var debugInfo map[string]any
	if comment, ok := node.(*Comment); ok && comment.DebugInfo != nil {
		debugInfo = comment.DebugInfo
	} else {
		return ""
	}

	lineNumber, ok := debugInfo["lineNumber"].(int)
	if !ok {
		return ""
	}
	
	fileName, ok := debugInfo["fileName"].(string)
	if !ok {
		return ""
	}

	var result string
	switch dumpLineNumbers {
	case "comments":
		result = asComment(lineNumber, fileName)
	case "mediaquery":
		result = asMediaQuery(lineNumber, fileName)
	case "all":
		result = asComment(lineNumber, fileName)
		if separator != "" {
			result += separator
		}
		result += asMediaQuery(lineNumber, fileName)
	}
	
	return result
} 