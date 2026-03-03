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

	// 构建边框详细信息（包括坐标）
	borderInfo := ""
	if box.BoxModel.Border.Style != BorderNone {
		// 边框宽度（单字符）
		borderWidth := box.BoxModel.Border.Width
		if borderWidth == 0 {
			borderWidth = 1
		}

		// 基本边框信息
		borderInfo = ", Border:" + box.BoxModel.Border.Style.String()
		if box.BoxModel.Border.Label != "" {
			borderInfo += fmt.Sprintf("(%q)", box.BoxModel.Border.Label)
		}

		// 边框绘制区域（与 box 相同）
		borderInfo += fmt.Sprintf(" [borderarea:%d,%d,%dx%d]",
			box.X, box.Y, box.Width, box.Height)

		// 内容区域坐标（边框内 + padding内）
		borderX := box.X + borderWidth
		borderY := box.Y + borderWidth
		borderW := box.Width - borderWidth*2
		borderH := box.Height - borderWidth*2
		paddingLeft := box.BoxModel.Padding.Left
		paddingRight := box.BoxModel.Padding.Right
		paddingTop := box.BoxModel.Padding.Top
		paddingBottom := box.BoxModel.Padding.Bottom

		contentX := borderX + paddingLeft
		contentY := borderY + paddingTop
		contentW := borderW - paddingLeft - paddingRight
		contentH := borderH - paddingTop - paddingBottom

		borderInfo += fmt.Sprintf(" [padding_inner:%d,%d,%d,%d]", paddingTop, paddingRight, paddingBottom, paddingLeft)
		borderInfo += fmt.Sprintf(" [content:%d,%d,%dx%d]", contentX, contentY, contentW, contentH)

		// 标签额外占用空间
		if box.BoxModel.Border.Label != "" {
			labelWidth := len(box.BoxModel.Border.Label) + 2 // 标签+2个空格
			borderInfo += fmt.Sprintf(" [label_w:%d]", labelWidth)
		}
	}

	// 构建 Margin 信息（从 BoxModel 读取）
	marginInfo := ""
	if box.BoxModel.Margin.Left != 0 || box.BoxModel.Margin.Right != 0 ||
		box.BoxModel.Margin.Top != 0 || box.BoxModel.Margin.Bottom != 0 {
		marginInfo = fmt.Sprintf(" [margin:%d,%d,%d,%d]",
			box.BoxModel.Margin.Top, box.BoxModel.Margin.Right,
			box.BoxModel.Margin.Bottom, box.BoxModel.Margin.Left)
	}

	// Append node with hierarchical relationship
	sb.WriteString(fmt.Sprintf("%s└─ [%s] %s (ID:%s, Size:%dx%d, Pos:%d,%d%s%s)\n",
		indent, box.Tag, propsID, box.ID, box.Width, box.Height, box.X, box.Y, borderInfo, marginInfo))

	// Append children recursively
	for _, child := range box.Children {
		lr.buildTreeNodeString(child, depth+1, sb)
	}
}