package layout

import (
	"github.com/wwsheng009/mint/framework/component"
	"github.com/wwsheng009/mint/runtime/layout"
	rtstyle "github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/runtime/paint"
)

// ==============================================================================
// Flex Layout Components (Column/Row)
// ==============================================================================
// These components wrap the runtime flex layout implementation at the framework level.

// FlexDirection 弹性方向
type FlexDirection = layout.FlexDirection

const (
	FlexRow         FlexDirection = layout.FlexRow         // 水平方向，从左到右
	FlexColumn      FlexDirection = layout.FlexColumn      // 垂直方向，从上到下
	FlexRowReverse  FlexDirection = layout.FlexRowReverse  // 水平方向，从右到左
	FlexColumnReverse FlexDirection = layout.FlexColumnReverse // 垂直方向，从下到上
)

// MainAxisAlignment 主轴对齐方式
type MainAxisAlignment = layout.MainAxisAlignment

const (
	MainStart    MainAxisAlignment = layout.MainStart    // 主轴起点对齐
	MainEnd      MainAxisAlignment = layout.MainEnd      // 主轴终点对齐
	MainCenter   MainAxisAlignment = layout.Center       // 主轴居中对齐
	MainSpaceBetween MainAxisAlignment = layout.SpaceBetween    // 两端对齐，间距平均分配
	MainSpaceAround  MainAxisAlignment = layout.SpaceAround     // 每个子元素两侧间距相等
	MainSpaceEvenly  MainAxisAlignment = layout.SpaceEvenly     // 所有间距相等
)

// CrossAxisAlignment 交叉轴对齐方式
type CrossAxisAlignment = layout.CrossAxisAlignment

const (
	CrossStart   CrossAxisAlignment = layout.CrossStart   // 交叉轴起点对齐
	CrossEnd     CrossAxisAlignment = layout.CrossEnd     // 交叉轴终点对齐
	CrossCenter  CrossAxisAlignment = layout.CrossCenter  // 交叉轴居中对齐
	CrossStretch CrossAxisAlignment = layout.Stretch      // 拉伸填满交叉轴
)

// Padding 内边距
type Padding struct {
	Top    int
	Right  int
	Bottom int
	Left   int
}

// FlexConfig 弹性配置
type FlexConfig = layout.Flex

// Flex 弹性布局容器
//
// Flex 包装 runtime flex 实现，提供 framework 层的组件接口。
// 它将 framework component.Node 转换为 runtime layout.Node 进行布局计算。
type Flex struct {
	*component.BaseComponent
	*component.StateHolder

	direction      FlexDirection
	mainAxis       MainAxisAlignment
	crossAxis      CrossAxisAlignment
	gap            int
	crossGap       int
	padding        Padding
	children       []component.Node
	flexChildren   map[int]*FlexConfig
	background     rtstyle.Color

	// runtime flex 布局引擎
	rtFlex *layout.FlexLayout
}

// NewFlex 创建一个新的弹性布局容器
func NewFlex() *Flex {
	f := &Flex{
		BaseComponent:  component.NewBaseComponent("flex"),
		StateHolder:    component.NewStateHolder(),
		direction:      FlexColumn,
		mainAxis:       MainStart,
		crossAxis:      CrossStart,
		gap:            0,
		crossGap:       0,
		padding:        Padding{},
		children:       make([]component.Node, 0),
		flexChildren:   make(map[int]*FlexConfig),
		background:     "",
	}
	// 初始化 runtime flex
	f.rtFlex = layout.NewFlexLayout(f.ID(), nil)
	return f
}

// NewColumn 创建一个垂直布局容器（从上到下）
func NewColumn() *Flex {
	return NewFlex().Direction(FlexColumn)
}

// NewRow 创建一个水平布局容器（从左到右）
func NewRow() *Flex {
	return NewFlex().Direction(FlexRow)
}

// ==============================================================================
// Chainable Configuration Methods
// ==============================================================================

// Direction 设置弹性方向
func (f *Flex) Direction(dir FlexDirection) *Flex {
	f.direction = dir
	f.rtFlex.SetDirection(dir)
	f.MarkDirty()
	return f
}

// MainAlign 设置主轴对齐方式
func (f *Flex) MainAlign(align MainAxisAlignment) *Flex {
	f.mainAxis = align
	f.rtFlex.SetMainAxis(align)
	f.MarkDirty()
	return f
}

// CrossAlign 设置交叉轴对齐方式
func (f *Flex) CrossAlign(align CrossAxisAlignment) *Flex {
	f.crossAxis = align
	f.rtFlex.SetCrossAxis(align)
	f.MarkDirty()
	return f
}

// Gap 设置主轴间距
func (f *Flex) Gap(gap int) *Flex {
	f.gap = gap
	f.rtFlex.SetGap(gap)
	f.MarkDirty()
	return f
}

// CrossGap 设置交叉轴间距
func (f *Flex) CrossGap(gap int) *Flex {
	f.crossGap = gap
	f.rtFlex.SetCrossGap(gap)
	f.MarkDirty()
	return f
}

// Padding 设置内边距（全部）
func (f *Flex) Padding(all int) *Flex {
	f.padding = Padding{
		Top:    all,
		Right:  all,
		Bottom: all,
		Left:   all,
	}
	f.rtFlex.SetPadding(all, all, all, all)
	f.MarkDirty()
	return f
}

// PaddingV 设置垂直内边距
func (f *Flex) PaddingV(vertical int) *Flex {
	f.padding.Top = vertical
	f.padding.Bottom = vertical
	f.rtFlex.SetPadding(f.padding.Left, f.padding.Right, vertical, vertical)
	f.MarkDirty()
	return f
}

// PaddingH 设置水平内边距
func (f *Flex) PaddingH(horizontal int) *Flex {
	f.padding.Left = horizontal
	f.padding.Right = horizontal
	f.rtFlex.SetPadding(horizontal, horizontal, f.padding.Top, f.padding.Bottom)
	f.MarkDirty()
	return f
}

// PaddingLTRB 分别设置四个方向的内边距
func (f *Flex) PaddingLTRB(left, top, right, bottom int) *Flex {
	f.padding = Padding{
		Left:   left,
		Top:    top,
		Right:  right,
		Bottom: bottom,
	}
	f.rtFlex.SetPadding(left, right, top, bottom)
	f.MarkDirty()
	return f
}

// Background 设置背景色
func (f *Flex) Background(color rtstyle.Color) *Flex {
	f.background = color
	f.MarkDirty()
	return f
}

// AddChild 添加子组件
func (f *Flex) AddChild(child component.Node) *Flex {
	f.children = append(f.children, child)
	f.syncRuntimeChildren()
	f.MarkDirty()
	return f
}

// AddChildren 添加多个子组件
func (f *Flex) AddChildren(children ...component.Node) *Flex {
	f.children = append(f.children, children...)
	f.syncRuntimeChildren()
	f.MarkDirty()
	return f
}

// Flex 设置子组件的弹性配置
func (f *Flex) Flex(index int, config FlexConfig) *Flex {
	f.flexChildren[index] = &config
	f.rtFlex.SetFlex(index, config.Grow, config.Shrink, config.Basis)
	f.MarkDirty()
	return f
}

// FlexGrow 设置子组件的放大比例
func (f *Flex) FlexGrow(index, grow int) *Flex {
	if config, ok := f.flexChildren[index]; ok {
		config.Grow = grow
	} else {
		f.flexChildren[index] = &FlexConfig{Grow: grow, Shrink: 1, Basis: 0}
	}
	f.rtFlex.SetFlex(index, grow, 1, 0)
	f.MarkDirty()
	return f
}

// FlexBasis 设置子组件的基础尺寸
func (f *Flex) FlexBasis(index, basis int) *Flex {
	if config, ok := f.flexChildren[index]; ok {
		config.Basis = basis
	} else {
		f.flexChildren[index] = &FlexConfig{Grow: 0, Shrink: 1, Basis: basis}
	}
	f.rtFlex.SetFlex(index, 0, 1, basis)
	f.MarkDirty()
	return f
}

// syncRuntimeChildren 同步子组件到 runtime flex
// 将 framework component.Node 转换为 runtime layout.Node
func (f *Flex) syncRuntimeChildren() {
	rtChildren := make([]layout.Node, len(f.children))
	for i, child := range f.children {
		rtChildren[i] = &componentNodeAdapter{component: child}
	}
	f.rtFlex = layout.NewFlexLayout(f.ID(), rtChildren)

	// 重新应用配置
	f.rtFlex.SetDirection(f.direction)
	f.rtFlex.SetMainAxis(f.mainAxis)
	f.rtFlex.SetCrossAxis(f.crossAxis)
	f.rtFlex.SetGap(f.gap)
	f.rtFlex.SetCrossGap(f.crossGap)
	f.rtFlex.SetPadding(f.padding.Left, f.padding.Right, f.padding.Top, f.padding.Bottom)

	// 重新应用 flex 配置
	for idx, config := range f.flexChildren {
		f.rtFlex.SetFlex(idx, config.Grow, config.Shrink, config.Basis)
	}
}

// ==============================================================================
// Container Interface Implementation
// ==============================================================================

// Children 返回子节点
func (f *Flex) Children() []component.Node {
	return f.children
}

// Add 添加子节点（实现 Container 接口）
func (f *Flex) Add(child component.Node) {
	f.AddChild(child)
}

// Remove 移除子节点
func (f *Flex) Remove(child component.Node) {
	for i, c := range f.children {
		if c == child {
			f.children = append(f.children[:i], f.children[i+1:]...)
			f.updateFlexIndices(i, -1)
			f.syncRuntimeChildren()
			f.MarkDirty()
			break
		}
	}
}

// RemoveAt 移除指定位置的子节点
func (f *Flex) RemoveAt(index int) {
	if index >= 0 && index < len(f.children) {
		f.children = append(f.children[:index], f.children[index+1:]...)
		f.updateFlexIndices(index, -1)
		f.syncRuntimeChildren()
		f.MarkDirty()
	}
}

// GetChildren 返回子节点（实现 Container 接口）
func (f *Flex) GetChildren() []component.Node {
	return f.children
}

// GetChild 返回指定位置的子节点
func (f *Flex) GetChild(index int) component.Node {
	if index >= 0 && index < len(f.children) {
		return f.children[index]
	}
	return nil
}

// ChildCount 返回子节点数量
func (f *Flex) ChildCount() int {
	return len(f.children)
}

// ClearChildren 清空所有子节点
func (f *Flex) ClearChildren() {
	f.children = make([]component.Node, 0)
	f.flexChildren = make(map[int]*FlexConfig)
	f.syncRuntimeChildren()
	f.MarkDirty()
}

// updateFlexIndices 更新弹性配置索引
func (f *Flex) updateFlexIndices(startIndex int, delta int) {
	newFlexChildren := make(map[int]*FlexConfig)
	for i, config := range f.flexChildren {
		newIndex := i
		if i >= startIndex {
			newIndex = i + delta
		}
		if newIndex >= 0 {
			newFlexChildren[newIndex] = config
		}
	}
	f.flexChildren = newFlexChildren
}

// ==============================================================================
// Measurable 接口实现
// ==============================================================================

// Measure 测量理想尺寸
// 委托给 runtime flex 进行计算
func (f *Flex) Measure(maxWidth, maxHeight int) (width, height int) {
	if len(f.children) == 0 {
		width = f.padding.Left + f.padding.Right
		height = f.padding.Top + f.padding.Bottom
		return
	}

	// 使用 runtime flex 进行测量
	constraints := layout.Constraints{
		MinWidth:  0,
		MaxWidth:  maxWidth,
		MinHeight: 0,
		MaxHeight: maxHeight,
	}

	size := f.rtFlex.Measure(constraints)
	return size.Width, size.Height
}

// ==============================================================================
// Paintable 接口实现
// ==============================================================================

// Paint 绘制组件
func (f *Flex) Paint(ctx component.PaintContext, buf *paint.Buffer) {
	if !f.IsVisible() {
		return
	}

	width := ctx.AvailableWidth
	height := ctx.AvailableHeight

	if width <= 0 || height <= 0 {
		return
	}

	// 绘制背景
	if f.background != "" {
		bgStyle := rtstyle.Style{}.Background(f.background)
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				buf.SetCell(ctx.X+x, ctx.Y+y, ' ', bgStyle)
			}
		}
	}

	// 计算内容区域
	contentX := ctx.X + f.padding.Left
	contentY := ctx.Y + f.padding.Top
	contentW := width - f.padding.Left - f.padding.Right
	contentH := height - f.padding.Top - f.padding.Bottom

	if contentW <= 0 || contentH <= 0 {
		return
	}

	// 使用 runtime flex 进行子组件布局
	// 然后将布局结果同步回 framework 组件
	boxes := f.rtFlex.LayoutChildren(contentW, contentH)

	for i, box := range boxes {
		if i < len(f.children) {
			child := f.children[i]
			if pos, ok := child.(component.Positionable); ok {
				pos.SetPosition(contentX+box.X, contentY+box.Y)
			}
			if sz, ok := child.(component.Sizable); ok {
				sz.SetSize(box.Width, box.Height)
			}

			// 递归绘制子组件
			if paintable, ok := child.(component.Paintable); ok {
				childCtx := component.PaintContext{
					Buffer:          buf,
					AvailableWidth:  box.Width,
					AvailableHeight: box.Height,
					X:               contentX + box.X,
					Y:               contentY + box.Y,
				}
				paintable.Paint(childCtx, buf)
			}
		}
	}
}

// ==============================================================================
// Getters
// ==============================================================================

// GetDirection 获取弹性方向
func (f *Flex) GetDirection() FlexDirection {
	return f.direction
}

// GetMainAlign 获取主轴对齐方式
func (f *Flex) GetMainAlign() MainAxisAlignment {
	return f.mainAxis
}

// GetCrossAlign 获取交叉轴对齐方式
func (f *Flex) GetCrossAlign() CrossAxisAlignment {
	return f.crossAxis
}

// GetGap 获取主轴间距
func (f *Flex) GetGap() int {
	return f.gap
}

// GetCrossGap 获取交叉轴间距
func (f *Flex) GetCrossGap() int {
	return f.crossGap
}

// GetPadding 获取内边距
func (f *Flex) GetPadding() Padding {
	return f.padding
}

// GetBackground 获取背景色
func (f *Flex) GetBackground() rtstyle.Color {
	return f.background
}

// GetFlexChildren 获取弹性配置
func (f *Flex) GetFlexChildren() map[int]*FlexConfig {
	return f.flexChildren
}

// ==============================================================================
// componentNodeAdapter
// ==============================================================================

// componentNodeAdapter 将 framework component.Node 适配为 runtime layout.Node
type componentNodeAdapter struct {
	component component.Node
	x, y      int
	width     int
	height    int
}

// ID 返回节点ID
func (a *componentNodeAdapter) ID() string {
	return a.component.ID()
}

// Type 返回节点类型
func (a *componentNodeAdapter) Type() string {
	return a.component.Type()
}

// Children 返回子节点
func (a *componentNodeAdapter) Children() []layout.Node {
	// 使用 Container 接口获取子节点
	if container, ok := a.component.(component.Container); ok {
		children := container.GetChildren()
		if children != nil {
			result := make([]layout.Node, len(children))
			for i, child := range children {
				result[i] = &componentNodeAdapter{component: child}
			}
			return result
		}
	}
	return nil
}

// GetPosition 获取位置
func (a *componentNodeAdapter) GetPosition() (int, int) {
	if pos, ok := a.component.(component.Positionable); ok {
		return pos.GetPosition()
	}
	return a.x, a.y
}

// SetPosition 设置位置
func (a *componentNodeAdapter) SetPosition(x, y int) {
	a.x = x
	a.y = y
	if pos, ok := a.component.(component.Positionable); ok {
		pos.SetPosition(x, y)
	}
}

// GetSize 获取尺寸
func (a *componentNodeAdapter) GetSize() (int, int) {
	if sz, ok := a.component.(component.Sizable); ok {
		return sz.GetSize()
	}
	return a.width, a.height
}

// SetSize 设置尺寸
func (a *componentNodeAdapter) SetSize(width, height int) {
	a.width = width
	a.height = height
	if sz, ok := a.component.(component.Sizable); ok {
		sz.SetSize(width, height)
	}
}

// GetWidth 获取宽度
func (a *componentNodeAdapter) GetWidth() int {
	w, _ := a.GetSize()
	return w
}

// GetHeight 获取高度
func (a *componentNodeAdapter) GetHeight() int {
	_, h := a.GetSize()
	return h
}

// Measure 测量节点尺寸（实现 runtime.layout.Measurable）
//
// 注意：component.Measurable (runtime.Measurable) 和 runtime.layout.Measurable
// 签名不同，这里使用匿名类型断言来适配 framework 层的 Measure 方法。
func (a *componentNodeAdapter) Measure(constraints layout.Constraints) layout.Size {
	// 尝试使用 framework 层的 Measure(int, int) (int, int) 方法
	if measurable, ok := a.component.(interface{ Measure(int, int) (int, int) }); ok {
		w, h := measurable.Measure(constraints.MaxWidth, constraints.MaxHeight)
		return layout.Size{Width: w, Height: h}
	}
	// 默认尺寸
	return layout.Size{Width: 0, Height: 0}
}
