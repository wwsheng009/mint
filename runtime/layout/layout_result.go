package layout

import (
	"fmt"
	"strings"
)

// String 返回布局树的字符串表示
func (lr *LayoutResult) String() string {
	return lr.TreeString()
}

// TreeString 返回布局树的字符串表示（层级结构）
func (lr *LayoutResult) TreeString() string {
	if lr == nil || lr.Root == nil {
		return "No layout tree found!"
	}

	var sb strings.Builder
	sb.WriteString("Layout Tree (hierarchical):\n")
	sb.WriteString(strings.Repeat("=", 70))
	sb.WriteString("\n")
	lr.buildTreeNodeString(lr.Root, 0, &sb)
	return sb.String()
}

// buildTreeNode 递归构建布局树节点的字符串表示
func (lr *LayoutResult) buildTreeNodeString(box *LayoutBox, depth int, sb *strings.Builder) {
	if box == nil {
		return
	}

	indent := strings.Repeat("  ", depth)
	propsID := box.PropsID
	if len(propsID) > 15 {
		propsID = propsID[:12] + "..."
	}
	if propsID == "" {
		propsID = "-"
	}

	// Append node with hierarchical relationship
	sb.WriteString(fmt.Sprintf("%s└─ [%s] %s (ID:%s, Size:%dx%d, Pos:%d,%d)\n",
		indent, box.Tag, propsID, box.ID, box.Width, box.Height, box.X, box.Y))

	// Append children recursively
	for _, child := range box.Children {
		lr.buildTreeNodeString(child, depth+1, sb)
	}
}