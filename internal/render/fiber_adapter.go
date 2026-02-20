// Package render provides Fiber to layout.Node adapter for the new layout engine
package render

import (
	"fmt"

	"github.com/wwsheng009/mint/internal/reconciler"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/compute"
	"github.com/wwsheng009/mint/runtime/layout"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// FiberToNodeAdapter - Adapts Fiber to layout.Node interface (Fiber-first)
// =============================================================================

// FiberToNodeAdapter wraps a Fiber tree to implement layout.Node interface
// This allows the new layout engine to work with Fiber trees
// Fiber-first: All data comes from Fiber fields, no VNode dependency
type FiberToNodeAdapter struct {
	fiber    *reconciler.Fiber
	children []layout.Node
}

// NewFiberToNodeAdapter creates a new adapter for a Fiber tree (Fiber-first)
func NewFiberToNodeAdapter(fiber *reconciler.Fiber, _ rtui.VNode) *FiberToNodeAdapter {
	adapter := &FiberToNodeAdapter{
		fiber: fiber,
	}
	adapter.initChildren()
	return adapter
}

// NewFiberToNodeAdapterPure creates a new adapter for a Fiber tree without VNode
func NewFiberToNodeAdapterPure(fiber *reconciler.Fiber) *FiberToNodeAdapter {
	adapter := &FiberToNodeAdapter{
		fiber: fiber,
	}
	adapter.initChildren()
	return adapter
}

// initChildren initializes children adapters from Fiber tree
func (a *FiberToNodeAdapter) initChildren() {
	if a.fiber == nil {
		return
	}

	// Build children from Fiber tree (Child -> Sibling linked list)
	childFibers := getFiberChildren(a.fiber)
	a.children = make([]layout.Node, len(childFibers))

	for i, childFiber := range childFibers {
		a.children[i] = NewFiberToNodeAdapterPure(childFiber)
	}
}

// getFiberChildren extracts children from a Fiber node
func getFiberChildren(fiber *rtui.Fiber) []*rtui.Fiber {
	if fiber == nil {
		return nil
	}

	var children []*rtui.Fiber

	// Fiber uses Child pointer for first child, Sibling for rest
	child := fiber.Child
	for child != nil {
		children = append(children, child)
		child = child.Sibling
	}

	return children
}

// ID returns the node identifier (from Fiber.NodeID)
func (a *FiberToNodeAdapter) ID() string {
	if a.fiber == nil {
		return ""
	}
	return fmt.Sprintf("%d", a.fiber.NodeID)
}

// Type returns the node type (from Fiber.Tag)
func (a *FiberToNodeAdapter) Type() string {
	if a.fiber == nil {
		return "unknown"
	}
	return string(a.fiber.Tag)
}

// Children returns child nodes
func (a *FiberToNodeAdapter) Children() []layout.Node {
	return a.children
}

// GetPosition returns the current position (from Fiber.ComputedBox)
func (a *FiberToNodeAdapter) GetPosition() (x, y int) {
	if a.fiber == nil {
		return 0, 0
	}
	// Try to get from computed box
	if a.fiber.ComputedBox != nil {
		if computedBox, ok := a.fiber.ComputedBox.(*compute.ComputedBox); ok {
			return computedBox.Box.X, computedBox.Box.Y
		}
	}
	return 0, 0
}

// SetPosition sets the position (stores in Fiber.ComputedBox)
func (a *FiberToNodeAdapter) SetPosition(x, y int) {
	if a.fiber == nil {
		return
	}
	// Store in fiber.ComputedBox if available
	if a.fiber.ComputedBox != nil {
		if computedBox, ok := a.fiber.ComputedBox.(*compute.ComputedBox); ok {
			computedBox.Box.X = x
			computedBox.Box.Y = y
		}
	}
}

// GetSize returns the current size (from Fiber fields)
// Fiber-first: All size data comes from Fiber.Instance or Fiber.Style
func (a *FiberToNodeAdapter) GetSize() (width, height int) {
	if a.fiber == nil {
		return 0, 0
	}

	// 1. Try Instance (preferred for migrated components)
	if a.fiber.Instance != nil {
		if sizable, ok := a.fiber.Instance.(interface{ GetSize() (int, int) }); ok {
			return sizable.GetSize()
		}
	}

	// 2. Try computed box (legacy compatibility)
	if a.fiber.ComputedBox != nil {
		if computedBox, ok := a.fiber.ComputedBox.(*compute.ComputedBox); ok {
			return computedBox.Box.Width, computedBox.Box.Height
		}
	}

	// 3. Try from Fiber.Style
	if a.fiber.Style.Width > 0 && a.fiber.Style.Height > 0 {
		return a.fiber.Style.Width, a.fiber.Style.Height
	}

	// 4. Try from Fiber.Props
	if a.fiber.Props != nil {
		if w, ok := a.fiber.Props["width"].(int); ok && w > 0 {
			if h, ok := a.fiber.Props["height"].(int); ok && h > 0 {
				return w, h
			}
		}
	}

	return 0, 0
}

// SetSize sets the size (stores in Fiber.ComputedBox)
func (a *FiberToNodeAdapter) SetSize(width, height int) {
	if a.fiber == nil {
		return
	}
	if a.fiber.ComputedBox != nil {
		if computedBox, ok := a.fiber.ComputedBox.(*compute.ComputedBox); ok {
			computedBox.Box.Width = width
			computedBox.Box.Height = height
		}
	}
}

// GetWidth returns the width
func (a *FiberToNodeAdapter) GetWidth() int {
	w, _ := a.GetSize()
	return w
}

// GetHeight returns the height
func (a *FiberToNodeAdapter) GetHeight() int {
	_, h := a.GetSize()
	return h
}

// Measure 实现 layout.Measurable 接口
// Fiber-first: 测量节点在给定约束下的理想尺寸
// 所有测量数据来自 Fiber.Instance（已迁移组件）或 Fiber.Style/Props
func (a *FiberToNodeAdapter) Measure(constraints layout.Constraints) layout.Size {
	if a.fiber == nil {
		return layout.Size{}
	}

	// 1. 从 Instance 获取尺寸（优先，用于已迁移组件）
	if a.fiber.Instance != nil {
		// 检查 Instance 是否实现 Measurable 接口
		if measurable, ok := a.fiber.Instance.(interface {
			Measure(layout.Constraints) layout.Size
		}); ok {
			return measurable.Measure(constraints)
		}

		// 检查 Instance 是否实现 Sizable 接口
		if sizable, ok := a.fiber.Instance.(interface {
			GetSize() (int, int)
		}); ok {
			w, h := sizable.GetSize()
			if w > 0 || h > 0 {
				return layout.Size{
					Width:  constraints.ConstrainWidth(w),
					Height: constraints.ConstrainHeight(h),
				}
			}
		}
	}

	// 2. 从 Style 获取固定尺寸
	if a.fiber.Style.Width > 0 || a.fiber.Style.Height > 0 {
		return layout.Size{
			Width:  constraints.ConstrainWidth(a.fiber.Style.Width),
			Height: constraints.ConstrainHeight(a.fiber.Style.Height),
		}
	}

	// 3. 从 Props 获取尺寸配置
	if a.fiber.Props != nil {
		if w, ok := a.fiber.Props["width"].(int); ok && w > 0 {
			if h, ok := a.fiber.Props["height"].(int); ok && h > 0 {
				return layout.Size{
					Width:  constraints.ConstrainWidth(w),
					Height: constraints.ConstrainHeight(h),
				}
			}
			return layout.Size{
				Width:  constraints.ConstrainWidth(w),
				Height: constraints.ConstrainHeight(0),
			}
		}
	}

	// 4. 默认值
	return layout.Size{Width: 0, Height: 0}
}

// GetMargin returns the margin from Fiber fields
// Implements layout.Marginal interface
func (a *FiberToNodeAdapter) GetMargin() layout.Margin {
	if a.fiber == nil {
		return layout.Margin{}
	}

	// Try Fiber.Props first
	if a.fiber.Props != nil {
		if m, ok := a.fiber.Props["margin"].([4]int); ok {
			return layout.Margin{
				Top:    m[0],
				Right:  m[1],
				Bottom: m[2],
				Left:   m[3],
			}
		}
	}

	return layout.Margin{}
}

// GetPositionType returns the position type from Fiber fields
// Implements layout.Positionable interface
func (a *FiberToNodeAdapter) GetPositionType() layout.Position {
	if a.fiber == nil {
		return layout.NewRelativePosition()
	}

	// Try Fiber.Props
	if a.fiber.Props != nil {
		if posType, ok := a.fiber.Props["position"].(string); ok {
			switch posType {
			case "absolute":
				pos := layout.NewAbsolutePosition()
				if top, ok := a.fiber.Props["top"].(int); ok {
					pos.Top = &top
				}
				if left, ok := a.fiber.Props["left"].(int); ok {
					pos.Left = &left
				}
				if right, ok := a.fiber.Props["right"].(int); ok {
					pos.Right = &right
				}
				if bottom, ok := a.fiber.Props["bottom"].(int); ok {
					pos.Bottom = &bottom
				}
				return pos
			}
		}
	}

	return layout.NewRelativePosition()
}

// GetFlexStyle returns the flex style from Fiber fields
// This implements FlexLayoutInfo interface for flex containers
func (a *FiberToNodeAdapter) GetFlexStyle() *layout.FlexStyle {
	if a.fiber == nil {
		return layout.DefaultFlexStyle()
	}

	style := layout.DefaultFlexStyle()

	// Map direction from Fiber.LayoutDirection
	switch a.fiber.LayoutDirection {
	case rtui.DirectionRow:
		style.Direction = layout.FlexRow
	case rtui.DirectionColumn:
		style.Direction = layout.FlexColumn
	}

	// Map alignment from Fiber.LayoutAlign
	switch a.fiber.LayoutAlign {
	case rtui.AlignStart:
		style.MainAxis = layout.MainStart
	case rtui.AlignCenter:
		style.MainAxis = layout.Center
	case rtui.AlignEnd:
		style.MainAxis = layout.MainEnd
	case rtui.AlignSpaceBetween:
		style.MainAxis = layout.SpaceBetween
	case rtui.AlignSpaceAround:
		style.MainAxis = layout.SpaceAround
	}

	// Map cross alignment from Fiber.LayoutCrossAlign
	switch a.fiber.LayoutCrossAlign {
	case rtui.AlignStart:
		style.CrossAxis = layout.CrossStart
	case rtui.AlignCenter:
		style.CrossAxis = layout.CrossCenter
	case rtui.AlignEnd:
		style.CrossAxis = layout.CrossEnd
	}

	// Gap from Fiber.LayoutGap
	style.Gap = a.fiber.LayoutGap

	// Padding from Fiber.LayoutPadding
	style.Padding = layout.Padding{
		Top:    a.fiber.LayoutPadding[0],
		Right:  a.fiber.LayoutPadding[1],
		Bottom: a.fiber.LayoutPadding[2],
		Left:   a.fiber.LayoutPadding[3],
	}

	return style
}

// GetBorder returns the border from Fiber fields
// Implements layout.Bordered interface
func (a *FiberToNodeAdapter) GetBorder() layout.Border {
	if a.fiber == nil {
		return layout.Border{Style: layout.BorderNone}
	}

	// Check tag for bordered container
	tag := a.fiber.Tag
	if tag == "bordered" || tag == "Bordered" || tag == "border" {
		// Extract border style from Fiber.Props
		if a.fiber.Props == nil {
			return layout.NewBorder(layout.BorderSingle)
		}

		borderStyle := layout.BorderSingle
		if s, ok := a.fiber.Props["borderStyle"].(string); ok {
			switch s {
			case "none":
				borderStyle = layout.BorderNone
			case "single":
				borderStyle = layout.BorderSingle
			case "double":
				borderStyle = layout.BorderDouble
			case "rounded":
				borderStyle = layout.BorderRounded
			case "dashed":
				borderStyle = layout.BorderDashed
			}
		}

		border := layout.NewBorder(borderStyle)

		// Extract label if present
		if label, ok := a.fiber.Props["borderLabel"].(string); ok {
			border.Label = label
		}

		return border
	}

	return layout.Border{Style: layout.BorderNone}
}

// GetStableID returns the stable node ID (from Fiber.NodeID)
// Implements layout.Identifiable interface
func (a *FiberToNodeAdapter) GetStableID() uint64 {
	if a.fiber == nil {
		return 0
	}
	return a.fiber.NodeID
}

// GetTextContent returns text content from Fiber (for text nodes)
func (a *FiberToNodeAdapter) GetTextContent() string {
	if a.fiber == nil {
		return ""
	}

	// Text content is stored in MemoizedState for text nodes
	if a.fiber.Type == rtui.VNodeText {
		if content, ok := a.fiber.MemoizedState.(string); ok {
			return content
		}
		// Fallback to Props
		if a.fiber.Props != nil {
			if content, ok := a.fiber.Props["content"].(string); ok {
				return content
			}
		}
	}
	return ""
}

// GetLayer returns the layer from Fiber
func (a *FiberToNodeAdapter) GetLayer() layout.Layer {
	if a.fiber == nil {
		return layout.LayerBase
	}

	switch a.fiber.Layer {
	case rtui.LayerBase:
		return layout.LayerBase
	case rtui.LayerOverlay:
		return layout.LayerDropdown
	case rtui.LayerModal:
		return layout.LayerModal
	case rtui.LayerTooltip:
		return layout.LayerTooltip
	default:
		return layout.LayerBase
	}
}

// GetZIndex returns the z-index from Fiber
// Implements layout.Layered interface
func (a *FiberToNodeAdapter) GetZIndex() int {
	if a.fiber == nil {
		return 0
	}
	// ZIndex could be stored in Fiber.Props or Fiber.Style
	// For now, return 0 as default
	return 0
}

// ========== layout.Dirtyable 接口实现 ==========

// IsLayoutDirty 检查是否需要重新布局
// 实现 layout.Dirtyable 接口
func (a *FiberToNodeAdapter) IsLayoutDirty() bool {
	if a.fiber == nil {
		return false
	}
	return a.fiber.IsLayoutDirty()
}

// ClearLayoutDirty 清除布局脏标记
// 实现 layout.Dirtyable 接口
func (a *FiberToNodeAdapter) ClearLayoutDirty() {
	if a.fiber == nil {
		return
	}
	a.fiber.ClearLayoutDirty()
}

// MarkLayoutDirty 标记需要重新布局
// 实现 layout.Dirtyable 接口
func (a *FiberToNodeAdapter) MarkLayoutDirty() {
	if a.fiber == nil {
		return
	}
	a.fiber.MarkLayoutDirty()
}

// =============================================================================
// VNodeToNodeAdapter - Adapts VNode to layout.Node interface
// =============================================================================

// VNodeToNodeAdapter wraps a VNode tree to implement layout.Node interface
type VNodeToNodeAdapter struct {
	vnode    rtui.VNode
	children []layout.Node
}

// NewVNodeToNodeAdapter creates a new adapter for a VNode tree
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
	return 0, 0
}

// SetPosition sets the position
func (a *VNodeToNodeAdapter) SetPosition(x, y int) {
	// VNode doesn't have position storage
}

// GetSize returns the current size
func (a *VNodeToNodeAdapter) GetSize() (width, height int) {
	if a.vnode == nil {
		return 0, 0
	}

	// Try to get from props
	props := a.vnode.Props()
	if props != nil {
		if w, ok := props["width"].(int); ok {
			if h, ok := props["height"].(int); ok {
				return w, h
			}
		}
	}

	return 0, 0
}

// SetSize sets the size
func (a *VNodeToNodeAdapter) SetSize(width, height int) {
	// VNode doesn't have size storage
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

// =============================================================================
// FlexLayoutAdapter - Creates FlexLayout from Fiber/VNode
// =============================================================================

// FlexLayoutAdapter creates layout.FlexLayout nodes from Fiber/VNode trees
// This handles the conversion of flex properties
type FlexLayoutAdapter struct {
	*FiberToNodeAdapter
	style *layout.FlexStyle
}

// NewFlexLayoutAdapter creates a flex layout adapter
func NewFlexLayoutAdapter(fiber *rtui.Fiber, vnode rtui.VNode) *FlexLayoutAdapter {
	adapter := &FlexLayoutAdapter{
		FiberToNodeAdapter: NewFiberToNodeAdapter(fiber, vnode),
		style:              extractFlexStyle(fiber, vnode),
	}
	return adapter
}

// GetFlexStyle returns the flex style
func (a *FlexLayoutAdapter) GetFlexStyle() *layout.FlexStyle {
	return a.style
}

// extractFlexStyle extracts flex style from Fiber/VNode
func extractFlexStyle(fiber *rtui.Fiber, vnode rtui.VNode) *layout.FlexStyle {
	style := layout.DefaultFlexStyle()

	// Extract from Fiber fields first (more reliable)
	if fiber != nil {
		// Map direction
		switch fiber.LayoutDirection {
		case rtui.DirectionRow:
			style.Direction = layout.FlexRow
		case rtui.DirectionColumn:
			style.Direction = layout.FlexColumn
		}

		// Map alignment
		switch fiber.LayoutAlign {
		case rtui.AlignStart:
			style.MainAxis = layout.MainStart
		case rtui.AlignCenter:
			style.MainAxis = layout.Center
		case rtui.AlignEnd:
			style.MainAxis = layout.MainEnd
		case rtui.AlignSpaceBetween:
			style.MainAxis = layout.SpaceBetween
		case rtui.AlignSpaceAround:
			style.MainAxis = layout.SpaceAround
		}

		// Map cross alignment
		switch fiber.LayoutCrossAlign {
		case rtui.AlignStart:
			style.CrossAxis = layout.CrossStart
		case rtui.AlignCenter:
			style.CrossAxis = layout.CrossCenter
		case rtui.AlignEnd:
			style.CrossAxis = layout.CrossEnd
		}

		// Stretch cross alignment (check StretchCross field)
		if fiber.LayoutFlex > 0 {
			// Has flex factor
		}

		// Gap
		style.Gap = fiber.LayoutGap

		// Padding
		style.Padding = layout.Padding{
			Top:    fiber.LayoutPadding[0],
			Right:  fiber.LayoutPadding[1],
			Bottom: fiber.LayoutPadding[2],
			Left:   fiber.LayoutPadding[3],
		}

		// Note: Margin is handled via Marginal interface, not FlexStyle
		// Margin is extracted in GetMargin() method of FiberToNodeAdapter
	}

	// Fallback to VNode layout info
	if vnode != nil {
		layoutInfo := rtui.GetLayoutInfo(vnode)

		// Map direction (only if not set from fiber)
		if fiber == nil || fiber.LayoutDirection == 0 {
			if layoutInfo.IsHorizontal {
				style.Direction = layout.FlexRow
			} else {
				style.Direction = layout.FlexColumn
			}
		}

		// Gap (only if not set from fiber)
		if fiber == nil || fiber.LayoutGap == 0 {
			style.Gap = layoutInfo.Gap
		}

		// Padding (only if not set from fiber)
		if fiber == nil || fiber.LayoutPadding == [4]int{} {
			style.Padding = layout.Padding{
				Top:    layoutInfo.Padding[0],
				Right:  layoutInfo.Padding[1],
				Bottom: layoutInfo.Padding[2],
				Left:   layoutInfo.Padding[3],
			}
		}

		// Alignment (only if not set from fiber)
		if fiber == nil || fiber.LayoutAlign == 0 {
			switch layoutInfo.Align {
			case rtui.AlignStart:
				style.MainAxis = layout.MainStart
			case rtui.AlignCenter:
				style.MainAxis = layout.Center
			case rtui.AlignEnd:
				style.MainAxis = layout.MainEnd
			case rtui.AlignSpaceBetween:
				style.MainAxis = layout.SpaceBetween
			case rtui.AlignSpaceAround:
				style.MainAxis = layout.SpaceAround
			}
		}

		// Cross alignment (only if not set from fiber)
		if fiber == nil || fiber.LayoutCrossAlign == 0 {
			switch layoutInfo.CrossAlign {
			case rtui.AlignStart:
				style.CrossAxis = layout.CrossStart
			case rtui.AlignCenter:
				style.CrossAxis = layout.CrossCenter
			case rtui.AlignEnd:
				style.CrossAxis = layout.CrossEnd
			}
			// Handle stretch cross
			if layoutInfo.StretchCross {
				style.CrossAxis = layout.Stretch
			}
		}
	}

	return style
}

// =============================================================================
// Constraint Conversion Utilities
// =============================================================================

// ConvertBoxConstraints converts runtime.BoxConstraints to layout.Constraints
func ConvertBoxConstraints(bc runtime.BoxConstraints) layout.Constraints {
	return layout.Constraints{
		MinWidth:  bc.MinWidth,
		MaxWidth:  bc.MaxWidth,
		MinHeight: bc.MinHeight,
		MaxHeight: bc.MaxHeight,
	}
}

// ConvertLayoutConstraints converts layout.Constraints to runtime.BoxConstraints
func ConvertLayoutConstraints(c layout.Constraints) runtime.BoxConstraints {
	return runtime.BoxConstraints{
		MinWidth:  c.MinWidth,
		MaxWidth:  c.MaxWidth,
		MinHeight: c.MinHeight,
		MaxHeight: c.MaxHeight,
	}
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
