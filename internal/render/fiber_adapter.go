// Package render provides Fiber to layout.Node adapter for the new layout engine
package render

import (
	"fmt"
	"reflect"

	"github.com/wwsheng009/mint/internal/reconciler"
	"github.com/wwsheng009/mint/runtime"
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

// GetPropsID returns the business identifier (from Fiber.ID)
// Implements layout.PropsIDProvider interface
func (a *FiberToNodeAdapter) GetPropsID() string {
	if a.fiber == nil {
		return ""
	}
	return a.fiber.ID
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

// GetPosition returns the current position
// Note: In MINT_FIBER_FIRST architecture, position is stored in layout.LayoutBox tree,
// not in Fiber. This method returns (0, 0) as Fiber doesn't track position.
// Use DeclarativeNode.lastLayoutResult to get layout positions.
func (a *FiberToNodeAdapter) GetPosition() (x, y int) {
	return 0, 0
}

// SetPosition sets the position
// Note: In MINT_FIBER_FIRST architecture, position is set by layout.Engine on LayoutBox,
// not on Fiber. This method syncs to Instance.bounds for component compatibility.
func (a *FiberToNodeAdapter) SetPosition(x, y int) {
	if a.fiber == nil {
		return
	}

	// Sync to Instance.bounds (component expects bounds for painting)
	if a.fiber.Instance != nil {
		if positionable, ok := a.fiber.Instance.(interface{ SetPosition(x, y int) }); ok {
			positionable.SetPosition(x, y)
		}
		// Try SetBounds for full bounds sync
		if boundsHaver, ok := a.fiber.Instance.(interface{ SetBounds(x, y, w, h int) }); ok {
			boundsHaver.SetBounds(x, y, 0, 0)
		}
	}
}

// GetSize returns the current size
// Fiber-first: Size data comes from Fiber.Instance, Fiber.Style, or Fibner.Props
// Note: Layout size is stored in layout.LayoutBox, not in Fiber.
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

	// 2. Try from Fiber.Style
	if a.fiber.Style.Width > 0 && a.fiber.Style.Height > 0 {
		return a.fiber.Style.Width, a.fiber.Style.Height
	}

	// 3. Try from Fiber.Props
	if a.fiber.Props != nil {
		if w, ok := a.fiber.Props["width"].(int); ok && w > 0 {
			if h, ok := a.fiber.Props["height"].(int); ok && h > 0 {
				return w, h
			}
		}
	}

	return 0, 0
}

// SetSize sets the size
// Note: In MINT_FIBER_FIRST architecture, size is set by layout.Engine on LayoutBox,
// not on Fiber. This method syncs to Instance.bounds for component compatibility.
func (a *FiberToNodeAdapter) SetSize(width, height int) {
	if a.fiber == nil {
		return
	}

	// Sync to Instance.bounds (component expects bounds for painting)
	if a.fiber.Instance != nil {
		if sizable, ok := a.fiber.Instance.(interface{ SetSize(width, height int) }); ok {
			sizable.SetSize(width, height)
		}
		// Try SetBounds for full bounds sync
		if boundsHaver, ok := a.fiber.Instance.(interface{ SetBounds(x, y, w, h int) }); ok {
			boundsHaver.SetBounds(0, 0, width, height)
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
//
// 边框处理（方案 A）：
// - width/height 是容器总尺寸（包含边框）
// - 子节点可用空间 = 总尺寸 - 边框占用
func (a *FiberToNodeAdapter) Measure(constraints layout.Constraints) layout.Size {
	if a.fiber == nil {
		return layout.Size{}
	}

	// ✨ 获取边框配置
	border := a.GetBorder()

	// 特殊处理：absolute 组件应该填充父容器的可用空间
	// 因为 absolute 是一个定位容器，它的子元素相对于它定位
	// 所以它应该使用约束的最大尺寸，但考虑父容器的边框
	if a.fiber.Tag == "absolute" {
		// 使用约束的最大尺寸，但要确保是非负值
		w := constraints.MaxWidth
		h := constraints.MaxHeight
		if w < 0 {
			w = 0
		}
		if h < 0 {
			h = 0
		}
		return layout.Size{
			Width:  w,
			Height: h,
		}
	}

	// 1. 从 Instance 获取尺寸（优先，用于已迁移组件）
	// Instance 的 Measure() 方法返回的是内容尺寸（不含容器边框）
	// ✨ 辅助函数：测量 Instance（内容尺寸），然后加上边框
	measureInstanceWithBorder := func(measurable interface {
		Measure(layout.Constraints) layout.Size
	}) layout.Size {
		// ✨ 调整约束：如果容器有边框， Instance 只能使用剩下空间
		instanceConstraints := constraints
		if border.HasBorder() {
			instanceConstraints = layout.Constraints{
				MinWidth:  max(0, constraints.MinWidth-border.HorizontalPadding()),
				MaxWidth:  max(0, constraints.MaxWidth-border.TotalHorizontalPadding()),
				MinHeight: max(0, constraints.MinHeight-border.VerticalPadding()),
				MaxHeight: max(0, constraints.MaxHeight-border.VerticalPadding()),
			}
		}

		// 测量 Instance（返回内容尺寸）
		size := measurable.Measure(instanceConstraints)

		// ✨ 边框：加上容器边框，返回总尺寸
		if border.HasBorder() {
			innerWidth := size.Width
			// Handle label width - the border must be wide enough for the label
			if border.Label != "" {
				labelWidth := len(" " + border.Label + " ")
				if labelWidth > innerWidth {
					innerWidth = labelWidth
				}
			}
			return layout.Size{
				Width:  innerWidth + border.TotalHorizontalPadding(),
				Height: size.Height + border.VerticalPadding(),
			}
		}
		return size
	}

	if a.fiber.Instance != nil {
		// 检查 Instance 是否实现 Measurable 接口
		if measurable, ok := a.fiber.Instance.(interface {
			Measure(layout.Constraints) layout.Size
		}); ok {
			return measureInstanceWithBorder(measurable)
		}

		// 检查 Instance 是否实现 Sizable 接口
		if sizable, ok := a.fiber.Instance.(interface {
			GetSize() (int, int)
		}); ok {
			w, h := sizable.GetSize()
			if w > 0 || h > 0 {
				// ✨ 边框：如果有边框，确保尺寸包含边框
				if border.HasBorder() {
					return layout.Size{
						Width:  constraints.ConstrainWidth(w + border.HorizontalPadding()),
						Height: constraints.ConstrainHeight(h + border.VerticalPadding()),
					}
				}
				return layout.Size{
					Width:  constraints.ConstrainWidth(w),
					Height: constraints.ConstrainHeight(h),
				}
			}
		}
	}

	// 2. 从 Style 获取固定尺寸
	// Style 的 width/height 是总尺寸（包含边框）
	if a.fiber.Style.Width > 0 || a.fiber.Style.Height > 0 {
		w := a.fiber.Style.Width
		h := a.fiber.Style.Height
		// ✨ 边框：如果有边框，尺寸已经是总尺寸（无需调整）
		// 因为用户设置的是"容器总宽度"，不是"内容宽度"
		return layout.Size{
			Width:  constraints.ConstrainWidth(w),
			Height: constraints.ConstrainHeight(h),
		}
	}

	// 3. 从 Props 获取尺寸配置
	if a.fiber.Props != nil {
		if w, ok := a.fiber.Props["width"].(int); ok && w > 0 {
			if h, ok := a.fiber.Props["height"].(int); ok && h > 0 {
				// ✨ 边框：Props 的 width/height 是总尺寸（包含边框）
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

	// 4. 测量子节点（自动尺寸）
	// 对于没有固定尺寸的容器，测量子节点以确定容器尺寸
	children := a.children
	if len(children) > 0 {
		// 获取 Flex 样式以判断布局方向
		flexStyle := a.GetFlexStyle()

		// 判断布局方向
		isFlexRow := flexStyle != nil && flexStyle.Direction == layout.FlexRow
		isFlexColumn := flexStyle != nil && flexStyle.Direction == layout.FlexColumn

		// 为子节点计算约束
		// 关键修复：容器应该将父容器的约束传递给子节点
		// 特别是跨轴约束必须传递，以确保子节点正确响应约束
		// HStack: 跨轴是高度，应该限制在父容器的高度内
		// VStack: 跨轴是宽度，应该限制在父容器的宽度内
		innerConstraints := layout.Constraints{
			MinWidth:  0,
			MaxWidth:  layout.MaxInt,  // 主轴方向：暂时无界（内容由布局引擎分配）
			MinHeight: 0,
			MaxHeight: layout.MaxInt, // 主轴方向：暂时无界（内容由布局引擎分配）
		}

		// ✅ 重要：传递跨轴约束
		// 这确保子节点能够正确响应父容器的尺寸限制
		if isFlexColumn {
			// VStack: 传递宽度约束（跨轴）
			if constraints.MaxWidth < layout.MaxInt {
				// 使用 TotalHorizontalPadding 而不是 HorizontalPadding
				// 因为边框可能包含标签，需要考虑标签额外的 2 像素填充
				innerConstraints.MinWidth = max(0, constraints.MinWidth-border.TotalHorizontalPadding())
				innerConstraints.MaxWidth = max(0, constraints.MaxWidth-border.TotalHorizontalPadding())
			}
		} else if isFlexRow {
			// HStack: 传递高度约束（跨轴）
			if constraints.MaxHeight < layout.MaxInt {
				innerConstraints.MinHeight = max(0, constraints.MinHeight-border.VerticalPadding())
				innerConstraints.MaxHeight = max(0, constraints.MaxHeight-border.VerticalPadding())
			}
		}

		// ⭐ OPTIMIZATION: 检查是否有 flex 子节点
		// 如果有 flex 子节点且父容器主轴方向有界限，自动填充可用宽度
		hasFlexChildren := false
		for _, child := range children {
			if flexChild, ok := child.(layout.FlexChildProvider); ok {
				if flexChild.GetFlex() > 0 {
					hasFlexChildren = true
					break
				}
			}
		}

		// 测量子节点 - 根据布局方向确定累加方式
		totalWidth := 0
		totalHeight := 0

		for _, child := range children {
			if measurable, ok := child.(layout.Measurable); ok {
				size := measurable.Measure(innerConstraints)

				if isFlexColumn {
					// VStack (FlexColumn): 宽度取最大，高度累加
					if size.Width > totalWidth {
						totalWidth = size.Width
					}
					totalHeight += size.Height
				} else if isFlexRow {
					// HStack (FlexRow): 宽度累加，高度取最大
					totalWidth += size.Width
					if size.Height > totalHeight {
						totalHeight = size.Height
					}
				} else {
					// 其他布局：取最大（保守处理）
					if size.Width > totalWidth {
						totalWidth = size.Width
					}
					if size.Height > totalHeight {
						totalHeight = size.Height
					}
				}
			}
		}

		// 添加 Gap（Flex 布局）
		if flexStyle != nil && flexStyle.Gap > 0 && len(children) > 1 {
			gapCount := len(children) - 1
			if isFlexRow {
				totalWidth += flexStyle.Gap * gapCount
			} else if isFlexColumn {
				totalHeight += flexStyle.Gap * gapCount
			}
		}

		// ⭐ OPTIMIZATION: Flex 容器自动填充可用宽度
		// 如果有 flex 子节点且父容器主轴方向有界限，使用父容器最大宽度
		if hasFlexChildren {
			if isFlexRow && constraints.MaxWidth < layout.MaxInt {
				// HStack: 使用父容器可用宽度（减去边框）
				innerWidth := max(0, constraints.MaxWidth-border.TotalHorizontalPadding())
				if innerWidth > totalWidth {
					totalWidth = innerWidth
				}
			} else if isFlexColumn && constraints.MaxHeight < layout.MaxInt {
				// VStack: 使用父容器可用高度（减去边框）
				innerHeight := max(0, constraints.MaxHeight-border.VerticalPadding())
				if innerHeight > totalHeight {
					totalHeight = innerHeight
				}
			}
		}

		// ✨ 边框：加上边框占用
		if border.HasBorder() {
			// Handle label width - the border must be wide enough for the label
			if border.Label != "" {
				labelWidth := len(" " + border.Label + " ")
				if labelWidth > totalWidth {
					totalWidth = labelWidth
				}
			}
			// 应用父容器的约束，确保不超出可用宽度
			return layout.Size{
				Width:  constraints.ConstrainWidth(totalWidth + border.TotalHorizontalPadding()),
				Height: constraints.ConstrainHeight(totalHeight + border.VerticalPadding()),
			}
		}
		// 应用父容器的约束（无边框情况）
		return layout.Size{
			Width:  constraints.ConstrainWidth(totalWidth),
			Height: constraints.ConstrainHeight(totalHeight),
		}
	}

	// 5. Text 元素特殊处理
	// 对于 text 元素，从 Props["content"] 或 MemoizedState 读取文本内容
	if a.fiber.Tag == "text" || a.fiber.Type == rtui.VNodeText {
		// 优先从 MemoizedState 读取（运行时状态）
		content := ""
		if a.fiber.MemoizedState != nil {
			if s, ok := a.fiber.MemoizedState.(string); ok {
				content = s
			}
		}
		// 回退到 Props
		if content == "" && a.fiber.Props != nil {
			if s, ok := a.fiber.Props["content"].(string); ok {
				content = s
			}
		}
		// 计算文本宽度（简单使用字符串长度）
		textWidth := len([]rune(content)) // 使用 rune 数来支持 multibyte 字符
		textHeight := 1 // 文本默认高度为 1

		// 处理带有边框的 text 元素（不常见，但可能）
		if border.HasBorder() {
			return layout.Size{
				Width:  constraints.ConstrainWidth(textWidth + border.TotalHorizontalPadding()),
				Height: constraints.ConstrainHeight(textHeight + border.VerticalPadding()),
			}
		}
		// 应用父容器的约束
		return layout.Size{
			Width:  constraints.ConstrainWidth(textWidth),
			Height: constraints.ConstrainHeight(textHeight),
		}
	}

	// 6. 默认值
	return layout.Size{Width: 0, Height: 0}
}

// GetMargin returns the margin from Fiber fields
// Implements layout.Marginal interface
func (a *FiberToNodeAdapter) GetMargin() layout.Margin {
	if a.fiber == nil {
		return layout.Margin{}
	}

	// ✨ 优先使用 Fiber.LayoutMargin 字段（在 completeWork 中设置）
	// LayoutMargin[4]int = [top, right, bottom, left]
	if a.fiber.LayoutMargin[0] != 0 || a.fiber.LayoutMargin[1] != 0 ||
		a.fiber.LayoutMargin[2] != 0 || a.fiber.LayoutMargin[3] != 0 {
		return layout.Margin{
			Top:    a.fiber.LayoutMargin[0],
			Right:  a.fiber.LayoutMargin[1],
			Bottom: a.fiber.LayoutMargin[2],
			Left:   a.fiber.LayoutMargin[3],
		}
	}

	// 向后兼容：Fiber.Props["margin"]（旧的方式）
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

// GetAbsolutePosition returns the absolute position from Fiber fields
// Implements layout.Positionable interface
func (a *FiberToNodeAdapter) GetAbsolutePosition() layout.Position {
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
// This implements FlexStyleProvider interface for flex containers
// Returns nil for non-flex containers (grid, absolute, etc.)
func (a *FiberToNodeAdapter) GetFlexStyle() *layout.FlexStyle {
	if a.fiber == nil {
		return nil
	}

	// Only return flex style for flex containers (vstack, hstack)
	if a.fiber.Tag != "vstack" && a.fiber.Tag != "hstack" {
		return nil
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

// GetFlex returns the flex factor from Fiber.LayoutFlex
// Implements layout.FlexChildProvider interface
func (a *FiberToNodeAdapter) GetFlex() int {
	if a.fiber == nil {
		return 0
	}
	return a.fiber.LayoutFlex
}

// GetGridStyle returns the grid style from Fiber fields
// Implements layout.GridStyleProvider interface
func (a *FiberToNodeAdapter) GetGridStyle() *layout.GridStyle {
	if a.fiber == nil {
		return nil
	}

	// Only for grid containers
	if a.fiber.Tag != "grid" {
		return nil
	}

	style := layout.DefaultGridStyle()

	// Extract from Props
	if a.fiber.Props != nil {
		// Columns - convert from component types to layout types
		style.Columns = convertGridDimensions(a.fiber.Props["columns"])
		// Rows
		style.Rows = convertGridDimensions(a.fiber.Props["rows"])
		// Cells - convert from component types to layout types
		style.Cells = convertGridCells(a.fiber.Props["cells"], a.children)

		// Gaps
		if gap, ok := a.fiber.Props["columnGap"].(int); ok {
			style.ColumnGap = gap
		}
		if gap, ok := a.fiber.Props["rowGap"].(int); ok {
			style.RowGap = gap
		}
		// Padding
		if pad, ok := a.fiber.Props["padding"].([4]int); ok {
			style.Padding = layout.Padding{
				Top:    pad[0],
				Right:  pad[1],
				Bottom: pad[2],
				Left:   pad[3],
			}
		}
		// Explicit size
		if w, ok := a.fiber.Props["width"].(int); ok {
			style.Width = w
		}
		if h, ok := a.fiber.Props["height"].(int); ok {
			style.Height = h
		}
		// ✨ Cell Borders
		if showBorders, ok := a.fiber.Props["showCellBorders"].(bool); ok {
			style.ShowCellBorders = showBorders
		}
		if style.ShowCellBorders {
			style.CellBorderWidth = 1
			style.CellBorderHeight = 1
		}
	}

	return style
}

// GetAbsoluteStyle returns the absolute positioning style from Fiber fields
// Implements layout.AbsoluteStyleProvider interface
func (a *FiberToNodeAdapter) GetAbsoluteStyle() *layout.AbsoluteStyle {
	if a.fiber == nil {
		return nil
	}

	// Only for absolute positioned nodes
	if a.fiber.Tag != "absolute" {
		return nil
	}

	style := layout.NewAbsoluteStyle()

	// Extract from Props
	if a.fiber.Props != nil {
		// Position values - convert from component types to layout types
		style.Left = convertPositionValue(a.fiber.Props["left"])
		style.Top = convertPositionValue(a.fiber.Props["top"])
		style.Right = convertPositionValue(a.fiber.Props["right"])
		style.Bottom = convertPositionValue(a.fiber.Props["bottom"])
		// Anchor - handle both int and component-specific Anchor types
		style.Anchor = convertAnchorValue(a.fiber.Props["anchor"])
		// Explicit size
		if w, ok := a.fiber.Props["width"].(int); ok {
			style.Width = w
		}
		if h, ok := a.fiber.Props["height"].(int); ok {
			style.Height = h
		}
		// ZIndex
		if z, ok := a.fiber.Props["zIndex"].(int); ok {
			style.ZIndex = z
		}
	}

	return style
}

// GetWrapStyle returns the wrap layout style from Fiber fields
// Implements layout.WrapStyleProvider interface
func (a *FiberToNodeAdapter) GetWrapStyle() *layout.WrapStyle {
	if a.fiber == nil {
		return nil
	}

	// Only for wrap containers
	if a.fiber.Tag != "wrap" {
		return nil
	}

	style := layout.DefaultWrapStyle()

	// Extract from Props
	if a.fiber.Props != nil {
		// Width - container width for wrap calculation
		if w, ok := a.fiber.Props["width"].(int); ok {
			style.Width = w
		}
		// Gap - spacing between items in the same row
		if gap, ok := a.fiber.Props["gap"].(int); ok {
			style.Gap = gap
		}
		// RowGap - spacing between rows (0 = use gap)
		if rowGap, ok := a.fiber.Props["rowGap"].(int); ok {
			style.RowGap = rowGap
		}
		// Align - row alignment
		if align, ok := a.fiber.Props["align"].(rtui.Align); ok {
			switch align {
			case rtui.AlignStart:
				style.Align = layout.WrapAlignStart
			case rtui.AlignCenter:
				style.Align = layout.WrapAlignCenter
			case rtui.AlignEnd:
				style.Align = layout.WrapAlignEnd
			}
		}
		// Padding
		if pad, ok := a.fiber.Props["padding"].([4]int); ok {
			style.Padding = layout.Padding{
				Top:    pad[0],
				Right:  pad[1],
				Bottom: pad[2],
				Left:   pad[3],
			}
		}
		// FillWidth
		if fill, ok := a.fiber.Props["fillWidth"].(bool); ok {
			style.FillWidth = fill
		}
		// FillHeight
		if fill, ok := a.fiber.Props["fillHeight"].(bool); ok {
			style.FillHeight = fill
		}
	}

	return style
}

// GetBorder returns the border from Fiber fields
// Implements layout.Bordered interface
func (a *FiberToNodeAdapter) GetBorder() layout.Border {
	if a.fiber == nil {
		return layout.Border{Style: layout.BorderNone}
	}

	// ✨ 方案 A: 优先使用 Fiber.BorderStyle 字段（对所有容器有效）
	if a.fiber.BorderStyle != "" && a.fiber.BorderStyle != "none" {
		borderStyle := parseBorderStyleString(a.fiber.BorderStyle)
		return layout.Border{
			Style: borderStyle,
			Label: a.fiber.BorderLabel,
		}
	}

	// 向后兼容：使用旧的 Props 方式
	props := a.fiber.Props
	if props == nil {
		return layout.Border{Style: layout.BorderNone}
	}

	tag := a.fiber.Tag

	// 处理 bordered 容器（旧组件）
	if tag == "bordered" || tag == "Bordered" || tag == "border" {
		borderStyle := layout.BorderSingle

		// 从 Props 中读取边框样式
		if s, ok := props["borderStyle"].(string); ok {
			borderStyle = parseBorderStyleString(s)
		} else if bs, ok := props["borderStyle"].(layout.BorderStyle); ok {
			borderStyle = bs
		} else if s, ok := props["style"].(string); ok {
			// 边框组件也可能使用 "style" 属性
			borderStyle = parseBorderStyleString(s)
		}

		// 读取标签
		label := ""
		if l, ok := props["borderLabel"].(string); ok {
			label = l
		} else if l, ok := props["label"].(string); ok {
			label = l
		}

		return layout.Border{
			Style: borderStyle,
			Label: label,
		}
	}

	// 处理 modal（旧组件的 props 方式）
	if tag == "modal" {
		borderStyle := layout.BorderDouble // Modal 默认双线边框

		if s, ok := props["borderStyle"].(string); ok {
			borderStyle = parseBorderStyleString(s)
		}

		label := ""
		if l, ok := props["title"].(string); ok {
			label = l
		} else if l, ok := props["label"].(string); ok {
			label = l
		}

		return layout.Border{
			Style: borderStyle,
			Label: label,
		}
	}

	return layout.Border{Style: layout.BorderNone}
}

// parseBorderStyleString 将字符串边框样式转换为 layout.BorderStyle
func parseBorderStyleString(s string) layout.BorderStyle {
	switch s {
	case "none":
		return layout.BorderNone
	case "single":
		return layout.BorderSingle
	case "double":
		return layout.BorderDouble
	case "rounded":
		return layout.BorderRounded
	case "dashed":
		return layout.BorderDashed
	default:
		return layout.BorderSingle // 默认单线边框
	}
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
// 由于 rtui.Layer 和 layout.Layer 都是 runtime.Layer 的别名，直接返回
func (a *FiberToNodeAdapter) GetLayer() layout.Layer {
	if a.fiber == nil {
		return layout.LayerBase
	}
	return layout.Layer(a.fiber.Layer)
}

// GetZIndex returns the z-index from Fiber
// Implements layout.Layered interface
func (a *FiberToNodeAdapter) GetZIndex() int {
	if a.fiber == nil {
		return 0
	}

	// Check Props["zIndex"] (for absolute components)
	if a.fiber.Props != nil {
		if z, ok := a.fiber.Props["zIndex"].(int); ok {
			return z
		}
	}

	// Check Instance for ZIndex (for absolute components)
	if a.fiber.Instance != nil {
		if layered, ok := a.fiber.Instance.(layout.Layered); ok {
			return layered.GetZIndex()
		}
	}

	// Default return 0
	return 0
}

// ✨ ShouldCenter returns whether a Modal should be centered (Phase 1.4)
// This is modal-specific centering logic controlled by Props["centered"]
func (a *FiberToNodeAdapter) ShouldCenter() bool {
	if a.fiber == nil {
		return false
	}

	// Only Modal components support centering
	if a.fiber.Tag != "modal" {
		return false
	}

	// Return the centered property from Fiber (synced from VNode props)
	return a.fiber.Centered
}

// GetPositionType returns the positioning scheme (Phase 2.3)
// Implements layout.PositionProvider interface
func (a *FiberToNodeAdapter) GetPositionType() layout.PositionType {
	if a.fiber == nil {
		return layout.PositionRelative
	}
	return a.fiber.Position
}

// GetAnchor returns the anchor point for position calculation (Phase 2.3)
// Implements layout.PositionProvider interface
func (a *FiberToNodeAdapter) GetAnchor() layout.Anchor {
	if a.fiber == nil {
		return layout.AnchorTopLeft
	}
	return a.fiber.Anchor
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

// extractFlexStyle extracts flex style from Fiber/VNode
//
// Deprecated: This function mixes Fiber and VNode data. Use Fiber-only data extraction.
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


// =============================================================================
// Type Conversion Helpers
// =============================================================================

// convertGridDimensions converts component dimension slice to layout dimension slice
func convertGridDimensions(val any) []layout.GridDimension {
	if val == nil {
		return nil
	}

	// Try as []layout.GridDimension first (direct layout type)
	if dims, ok := val.([]layout.GridDimension); ok {
		return dims
	}

	// Use reflection to handle different slice types
	rv := reflect.ValueOf(val)
	if rv.Kind() != reflect.Slice {
		return nil
	}

	result := make([]layout.GridDimension, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		result[i] = convertSingleDimension(rv.Index(i).Interface())
	}
	return result
}

// convertSingleDimension converts a single dimension value using reflection
func convertSingleDimension(val any) layout.GridDimension {
	if val == nil {
		return layout.GridFlex{Factor: 1}
	}

	// First check for layout types directly
	switch v := val.(type) {
	case layout.GridDimension:
		return v
	case int:
		return layout.GridFixed(v)
	}

	// Use reflection to detect type by name
	rv := reflect.ValueOf(val)
	rt := rv.Type()

	// Check for Fixed type (int-based)
	if rt.Name() == "Fixed" || rt.Name() == "GridFixed" {
		return layout.GridFixed(int(rv.Int()))
	}

	// Check for Flex type (struct with Factor field)
	if rt.Name() == "Flex" || rt.Name() == "GridFlex" {
		factor := 1
		if f := rv.FieldByName("Factor"); f.IsValid() {
			factor = int(f.Int())
		}
		return layout.GridFlex{Factor: factor}
	}

	// Check for Auto type
	if rt.Name() == "Auto" || rt.Name() == "GridAuto" {
		return layout.GridAuto{}
	}

	// Check for Min type
	if rt.Name() == "Min" || rt.Name() == "GridMin" {
		minVal := 0
		if m := rv.FieldByName("Min"); m.IsValid() {
			minVal = int(m.Int())
		}
		return layout.GridMin{Min: minVal}
	}

	// Check for Max type
	if rt.Name() == "Max" || rt.Name() == "GridMax" {
		maxVal := 0
		if m := rv.FieldByName("Max"); m.IsValid() {
			maxVal = int(m.Int())
		}
		return layout.GridMax{Max: maxVal}
	}

	// Default to flex
	return layout.GridFlex{Factor: 1}
}

// convertGridCells converts component cell slice to layout cell slice
func convertGridCells(val any, children []layout.Node) []layout.GridCell {
	if val == nil {
		// Auto-generate cells from children
		if len(children) > 0 {
			cells := make([]layout.GridCell, len(children))
			for i, child := range children {
				cells[i] = layout.GridCell{
					Child:   child,
					Row:     0,
					Col:     i,
					RowSpan: 1,
					ColSpan: 1,
				}
			}
			return cells
		}
		return nil
	}

	// Try as []layout.GridCell first
	if cells, ok := val.([]layout.GridCell); ok {
		return cells
	}

	// Use reflection to handle different slice types
	rv := reflect.ValueOf(val)
	if rv.Kind() != reflect.Slice {
		return nil
	}

	result := make([]layout.GridCell, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		result[i] = convertSingleCell(rv.Index(i).Interface(), i, children)
	}
	return result
}

// convertSingleCell converts a single cell value using reflection
func convertSingleCell(val any, index int, children []layout.Node) layout.GridCell {
	if val == nil {
		if index < len(children) {
			return layout.GridCell{
				Child:   children[index],
				Row:     0,
				Col:     index,
				RowSpan: 1,
				ColSpan: 1,
			}
		}
		return layout.GridCell{}
	}

	// Try as layout.GridCell
	if cell, ok := val.(layout.GridCell); ok {
		return cell
	}

	// Use reflection to extract cell fields
	rv := reflect.ValueOf(val)
	if rv.Kind() == reflect.Struct {
		cell := layout.GridCell{
			Child:   nil, // Will be set below
			Row:     0,
			Col:     index,
			RowSpan: 1,
			ColSpan: 1,
		}

		// Extract Row
		if f := rv.FieldByName("Row"); f.IsValid() {
			cell.Row = int(f.Int())
		}
		// Extract Col
		if f := rv.FieldByName("Col"); f.IsValid() {
			cell.Col = int(f.Int())
		}
		// Extract RowSpan
		if f := rv.FieldByName("RowSpan"); f.IsValid() {
			cell.RowSpan = int(f.Int())
		}
		// Extract ColSpan
		if f := rv.FieldByName("ColSpan"); f.IsValid() {
			cell.ColSpan = int(f.Int())
		}

		// IMPORTANT: cells[i].Child (VNode) corresponds to children[i] (layout.Node adapter)
		// The children array is built from Fiber's child chain, which is populated
		// from VNode.Children() - and grid cells are populated in the same order
		if index < len(children) {
			cell.Child = children[index]
		}

		return cell
	}

	// Create default cell
	cell := layout.GridCell{
		Row:     0,
		Col:     index,
		RowSpan: 1,
		ColSpan: 1,
	}

	// Use index for child if available
	if index < len(children) {
		cell.Child = children[index]
	}

	return cell
}

// convertPositionValue converts component position value to layout position value
func convertPositionValue(val any) layout.PositionValue {
	if val == nil {
		return nil
	}

	// Direct type checks
	switch v := val.(type) {
	case layout.PositionValue:
		return v
	case int:
		return layout.AbsolutePos(v)
	case layout.AbsolutePos:
		return v
	case layout.RelativePos:
		return v
	}

	// Use reflection for component types
	rv := reflect.ValueOf(val)
	rt := rv.Type()

	// Check for AbsolutePos type (int-based)
	if rt.Name() == "AbsolutePos" {
		return layout.AbsolutePos(int(rv.Int()))
	}

	// Check for RelativePos type (int-based)
	if rt.Name() == "RelativePos" {
		return layout.RelativePos(int(rv.Int()))
	}

	return nil
}

// convertAnchorValue converts anchor value to layout.Anchor
// Since absolute.Anchor is now an alias for layout.Anchor, direct type assertion works
func convertAnchorValue(val any) layout.Anchor {
	if val == nil {
		return layout.AnchorTopLeft // default
	}

	switch v := val.(type) {
	case layout.Anchor:
		return v
	case int:
		return layout.Anchor(v)
	}

	return layout.AnchorTopLeft // default
}
