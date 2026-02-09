package inspector

import (
	"fmt"
	"github.com/wwsheng009/mint/runtime"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"strings"
)

// LayoutNodeSnapshot 记录单个节点的布局快照
type LayoutNodeSnapshot struct {
	ID          string
	Type        string
	Constraints runtime.BoxConstraints
	Size        runtime.Size
	Props       map[string]interface{}
	Children    []*LayoutNodeSnapshot
	Depth       int
}

// LayoutAnalyzer 布局分析器工具
type LayoutAnalyzer struct {
	Snapshots map[uintptr]*LayoutNodeSnapshot
}

func NewLayoutAnalyzer() *LayoutAnalyzer {
	return &LayoutAnalyzer{
		Snapshots: make(map[uintptr]*LayoutNodeSnapshot),
	}
}

// FormatTree 将捕获到的布局信息格式化为可视化树
func (la *LayoutAnalyzer) FormatTree(node *LayoutNodeSnapshot) string {
	var sb strings.Builder
	la.formatNode(&sb, node, 0, true)
	return sb.String()
}

func (la *LayoutAnalyzer) formatNode(sb *strings.Builder, node *LayoutNodeSnapshot, depth int, isLast bool) {
	indent := strings.Repeat("  ", depth)
	prefix := "├── "
	if isLast {
		prefix = "└── "
	}
	if depth == 0 {
		prefix = ""
	}

	// 核心信息：类型 [宽 x 高]
	line := fmt.Sprintf("%s%s <%s> Size: %dx%d",
		indent, prefix, node.Type, node.Size.Width, node.Size.Height)

	// 约束信息
	constraints := fmt.Sprintf(" Constraints: [W:%v-%v, H:%v-%v]",
		la.fmtVal(node.Constraints.MinWidth), la.fmtVal(node.Constraints.MaxWidth),
		la.fmtVal(node.Constraints.MinHeight), la.fmtVal(node.Constraints.MaxHeight))

	sb.WriteString(line + constraints + "\n")

	// 打印关键属性 (Width, Height, Flex)
	if len(node.Props) > 0 {
		propStr := ""
		if w, ok := node.Props["width"]; ok {
			propStr += fmt.Sprintf(" width=%v", w)
		}
		if h, ok := node.Props["height"]; ok {
			propStr += fmt.Sprintf(" height=%v", h)
		}
		if f, ok := node.Props["flex"]; ok {
			propStr += fmt.Sprintf(" flex=%v", f)
		}

		if propStr != "" {
			sb.WriteString(fmt.Sprintf("%s    Props:%s\n", indent, propStr))
		}
	}

	for i, child := range node.Children {
		la.formatNode(sb, child, depth+1, i == len(node.Children)-1)
	}
}

func (la *LayoutAnalyzer) fmtVal(v int) string {
	if v >= 1000000 {
		return "∞"
	}
	return fmt.Sprintf("%d", v)
}

// Capture 从一个 VNode 递归生成布局快照树
// 注意：这假设 VNode 已经被测量过了
func (la *LayoutAnalyzer) Capture(vnode rtui.VNode, depth int) *LayoutNodeSnapshot {
	if vnode == nil {
		return nil
	}

	// 获取节点当前的尺寸
	// 注意：VNode 不直接存储尺寸信息，尺寸在布局计算后存储在 ComputedBox 中
	// 这里我们使用零尺寸，实际尺寸会在布局过程中计算
	size := runtime.Size{Width: 0, Height: 0}

	snapshot := &LayoutNodeSnapshot{
		Type: vnode.Type().String(),
		Size: size,
		// 注意：Constraints 通常在测量后不会保存在 VNode 中，
		// 这就是为什么我们需要在测量过程中捕获它们。
		// 这里我们先从 Props 中获取一些线索。
		Props:    vnode.Props(),
		Depth:    depth,
		Children: make([]*LayoutNodeSnapshot, 0),
	}

	for _, child := range vnode.Children() {
		childSnap := la.Capture(child, depth+1)
		if childSnap != nil {
			snapshot.Children = append(snapshot.Children, childSnap)
		}
	}

	return snapshot
}
