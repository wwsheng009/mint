package render

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/layout"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// VNodeToNodeAdapter - Adapts VNode to layout.Node interface
// =============================================================================

// VNodeToNodeAdapter wraps a VNode tree to implement layout.Node interface
//
// Deprecated: Use FiberToNodeAdapterPure with Fiber-first architecture instead.
// In Fiber-first architecture, VNode is discarded after Fiber creation.
// This adapter is only kept for legacy compatibility.
type VNodeToNodeAdapter struct {
	vnode    rtui.VNode
	children []layout.Node
	// Layout result storage (set by layout.Engine)
	x, y          int
	width, height int
}

// NewVNodeToNodeAdapter creates a new adapter for a VNode tree
//
// Deprecated: Use NewFiberToNodeAdapterPure with Fiber-first architecture instead.
func NewVNodeToNodeAdapter(vnode rtui.VNode) *VNodeToNodeAdapter {
	adapter := &VNodeToNodeAdapter{
		vnode: vnode,
	}
	adapter.initChildren()
	return adapter
}

// initChildren initializes children adapters
func (a *VNodeToNodeAdapter) initChildren() {
	if a.vnode == nil {
		return
	}

	vnodeChildren := a.vnode.Children()
	a.children = make([]layout.Node, len(vnodeChildren))

	for i, child := range vnodeChildren {
		a.children[i] = NewVNodeToNodeAdapter(child)
	}
}

// ID returns the node identifier
func (a *VNodeToNodeAdapter) ID() string {
	if a.vnode == nil {
		return ""
	}
	key := a.vnode.Key()
	if key != "" {
		return key
	}
	// Use type as fallback
	return fmt.Sprintf("%d", a.vnode.Type())
}

// Type returns the node type
func (a *VNodeToNodeAdapter) Type() string {
	if a.vnode == nil {
		return "unknown"
	}
	return fmt.Sprintf("%d", a.vnode.Type())
}

// Children returns child nodes
func (a *VNodeToNodeAdapter) Children() []layout.Node {
	return a.children
}

// GetPosition returns the current position
func (a *VNodeToNodeAdapter) GetPosition() (x, y int) {
	return a.x, a.y
}

// SetPosition sets the position
func (a *VNodeToNodeAdapter) SetPosition(x, y int) {
	a.x, a.y = x, y
}

// GetSize returns the current size
func (a *VNodeToNodeAdapter) GetSize() (width, height int) {
	// First check if we have layout-computed size
	if a.width > 0 && a.height > 0 {
		return a.width, a.height
	}

	// Try to get from props as fallback
	if a.vnode != nil {
		props := a.vnode.Props()
		if props != nil {
			if w, ok := props["width"].(int); ok && w > 0 {
				if h, ok := props["height"].(int); ok && h > 0 {
					return w, h
				}
			}
		}
	}

	return a.width, a.height
}

// SetSize sets the size
func (a *VNodeToNodeAdapter) SetSize(width, height int) {
	a.width, a.height = width, height
}

// Measure 实现 layout.Measurable 接口
// 测量 VNode 在给定约束下的理想尺寸
func (a *VNodeToNodeAdapter) Measure(constraints layout.Constraints) layout.Size {
	if a.vnode == nil {
		return layout.Size{}
	}

	// 1. 检查是否有明确的尺寸约束
	props := a.vnode.Props()
	if props != nil {
		if w, ok := props["width"].(int); ok && w > 0 {
			if h, ok := props["height"].(int); ok && h > 0 {
				return layout.Size{
					Width:  constraints.ConstrainWidth(w),
					Height: constraints.ConstrainHeight(h),
				}
			}
		}
	}

	// 2. 检查 VNode 是否实现了 Measurable 接口
	// 这允许已经迁移的组件提供自己的测量逻辑
	type vnodeMeasurable interface {
		Measure(layout.Constraints) layout.Size
	}
	if m, ok := a.vnode.(vnodeMeasurable); ok {
		return m.Measure(constraints)
	}

	// 3. 对于文本节点，测量文本内容
	if content := rtui.GetTextContent(a.vnode); content != "" {
		width := len(content)
		// 如果有最大宽度约束，文本可能需要换行（简化处理，假设不换行）
		if constraints.MaxWidth > 0 && width > constraints.MaxWidth {
			width = constraints.MaxWidth
		}
		height := 1
		return layout.Size{
			Width:  constraints.ConstrainWidth(width),
			Height: constraints.ConstrainHeight(height),
		}
	}

	// 4. 对于 button 元素（有 label prop），测量标签长度
	if label, ok := props["label"].(string); ok && label != "" {
		width := len(label) + 2 // 添加 [] 括号
		height := 1
		return layout.Size{
			Width:  constraints.ConstrainWidth(width),
			Height: constraints.ConstrainHeight(height),
		}
	}

	// 5. Fragment 节点：累加子节点高度，但宽度保持为 0（兼容旧测试）
	// Note: Fragments don't have intrinsic width calculation
	// 这是已知的限制，为了兼容性保持这个行为
	// 注意：这个检查必须在 children 逻辑之前，因为 Fragment 也有 children
	if a.vnode != nil && a.vnode.Type() == rtui.VNodeFragment {
		children := a.vnode.Children()
		var totalHeight int
		for _, child := range children {
			childAdapter := NewVNodeToNodeAdapter(child)
			childSize := childAdapter.Measure(constraints)
			totalHeight += childSize.Height
		}
		return layout.Size{
			Width:  0, // Fragment 的宽度保持为 0（已知限制）
			Height: totalHeight,
		}
	}

	// 6. 对于容器节点（HStack, VStack 等），从布局信息获取 gap
	layoutInfo := rtui.GetLayoutInfo(a.vnode)
	children := a.vnode.Children()
	if len(children) > 0 {
		// 简化处理：累加子节点尺寸
		var totalWidth, totalHeight int
		for i, child := range children {
			childAdapter := NewVNodeToNodeAdapter(child)
			childSize := childAdapter.Measure(constraints)
			if layoutInfo.IsHorizontal {
				totalWidth += childSize.Width
				totalHeight = max(totalHeight, childSize.Height)
				if i > 0 && layoutInfo.Gap > 0 {
					totalWidth += layoutInfo.Gap
				}
			} else {
				totalHeight += childSize.Height
				totalWidth = max(totalWidth, childSize.Width)
				if i > 0 && layoutInfo.Gap > 0 {
					totalHeight += layoutInfo.Gap
				}
			}
		}
		return layout.Size{
			Width:  constraints.ConstrainWidth(totalWidth),
			Height: constraints.ConstrainHeight(totalHeight),
		}
	}

	// 7. 默认尺寸
	return layout.Size{
		Width:  1,
		Height: 1,
	}
}


// GetWidth returns the width
func (a *VNodeToNodeAdapter) GetWidth() int {
	w, _ := a.GetSize()
	return w
}

// GetHeight returns the height
func (a *VNodeToNodeAdapter) GetHeight() int {
	_, h := a.GetSize()
	return h
}

// GetMargin returns the margin from VNode
// Implements layout.Marginal interface
func (a *VNodeToNodeAdapter) GetMargin() layout.Margin {
	if a.vnode == nil {
		return layout.Margin{}
	}

	// Try props first
	if props := a.vnode.Props(); props != nil {
		if m, ok := props["margin"].([4]int); ok {
			return layout.Margin{
				Top:    m[0],
				Right:  m[1],
				Bottom: m[2],
				Left:   m[3],
			}
		}
	}

	// Try layout info
	layoutInfo := rtui.GetLayoutInfo(a.vnode)
	return layout.Margin{
		Top:    layoutInfo.Margin[0],
		Right:  layoutInfo.Margin[1],
		Bottom: layoutInfo.Margin[2],
		Left:   layoutInfo.Margin[3],
	}
}

// GetPositionType returns the position type from VNode
// Implements layout.Positionable interface
func (a *VNodeToNodeAdapter) GetPositionType() layout.Position {
	if a.vnode == nil {
		return layout.NewRelativePosition()
	}

	// Try props
	if props := a.vnode.Props(); props != nil {
		if posType, ok := props["position"].(string); ok {
			switch posType {
			case "absolute":
				pos := layout.NewAbsolutePosition()
				if top, ok := props["top"].(int); ok {
					pos.Top = &top
				}
				if left, ok := props["left"].(int); ok {
					pos.Left = &left
				}
				if right, ok := props["right"].(int); ok {
					pos.Right = &right
				}
				if bottom, ok := props["bottom"].(int); ok {
					pos.Bottom = &bottom
				}
				return pos
			}
		}
	}

	return layout.NewRelativePosition()
}


// GetBorder returns the border from VNode
// Implements layout.Bordered interface for VNodeToNodeAdapter
func (a *VNodeToNodeAdapter) GetBorder() layout.Border {
	if a.vnode == nil {
		return layout.Border{Style: layout.BorderNone}
	}

	// Check tag for bordered container
	if a.vnode.Type() == rtui.VNodeElement {
		tag := a.vnode.Tag()
		if tag == "bordered" || tag == "Bordered" || tag == "border" {
			props := a.vnode.Props()
			if props == nil {
				return layout.NewBorder(layout.BorderSingle)
			}

			style := layout.BorderSingle
			if s, ok := props["borderStyle"].(string); ok {
				switch s {
				case "none":
					style = layout.BorderNone
				case "single":
					style = layout.BorderSingle
				case "double":
					style = layout.BorderDouble
				case "rounded":
					style = layout.BorderRounded
				case "dashed":
					style = layout.BorderDashed
				}
			}

			border := layout.NewBorder(style)
			if label, ok := props["borderLabel"].(string); ok {
				border.Label = label
			}

			return border
		}
	}

	return layout.Border{Style: layout.BorderNone}
}