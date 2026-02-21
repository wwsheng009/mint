package layout

// ==============================================================================
// Layout Types (V3)
// ==============================================================================
// 核心布局类型定义

// Node 布局节点接口
// 这是布局引擎操作的抽象节点
type Node interface {
	// ID 返回节点唯一标识
	ID() string

	// Type 返回节点类型
	Type() string

	// Children 返回子节点
	Children() []Node

	// GetPosition 获取位置
	GetPosition() (x, y int)

	// SetPosition 设置位置
	SetPosition(x, y int)

	// GetSize 获取尺寸
	GetSize() (width, height int)

	// SetSize 设置尺寸
	SetSize(width, height int)

	// GetWidth 获取宽度
	GetWidth() int

	// GetHeight 获取高度
	GetHeight() int
}

// Identifiable 可标识接口
// 节点可以实现此接口以提供稳定的标识符
type Identifiable interface {
	// GetStableID 返回节点的稳定标识符
	// 该标识符应该在节点的整个生命周期中保持不变
	GetStableID() uint64
}

// Versioned 可版本接口
// 节点可以实现此接口以跟踪版本信息
type Versioned interface {
	// GetVersion 返回节点的版本号
	// 版本号应该在节点内容发生变化时递增
	GetVersion() uint64
}

// LayoutInfoProvider 布局信息提供者接口
// 节点可以实现此接口以提供布局信息
type LayoutInfoProvider interface {
	// GetLayoutInfo 返回节点的布局信息
	GetLayoutInfo() *LayoutInfo
}

// LayoutInfo 布局信息
// 包含节点的标识符、版本和布局结果
type LayoutInfo struct {
	// ID 节点的稳定标识符
	ID uint64

	// Version 节点的版本号
	Version uint64

	// LayoutBox 布局结果盒子
	LayoutBox *LayoutBox
}

// Size 尺寸
type Size struct {
	Width  int
	Height int
}

// Point 位置
type Point struct {
	X int
	Y int
}

// Rect 矩形区域
type Rect struct {
	X      int
	Y      int
	Width  int
	Height int
}

// Contains 检查点 (x, y) 是否在矩形内
func (r Rect) Contains(x, y int) bool {
	return x >= r.X && x < r.X+r.Width &&
		y >= r.Y && y < r.Y+r.Height
}

// Intersects 检查两个矩形是否相交
func (r Rect) Intersects(other Rect) bool {
	return r.X < other.X+other.Width &&
		r.X+r.Width > other.X &&
		r.Y < other.Y+other.Height &&
		r.Y+r.Height > other.Y
}

// LayoutBox 布局结果盒子
// 表示一个节点在布局后的最终位置和尺寸
type LayoutBox struct {
	// ID 节点ID
	ID string

	// X, Y 位置（相对于父节点）
	X int
	Y int

	// Width, Height 尺寸
	Width  int
	Height int

	// Baseline 基线（用于文本对齐）
	Baseline int

	// Layer 渲染层（用于多层渲染）
	Layer Layer

	// ZIndex 层内排序索引
	ZIndex int

	// Border 边框信息（如果有）
	Border Border

	// Children 子节点布局结果
	Children []*LayoutBox
}

// LayoutResult 布局结果
// 一次布局计算的完整结果
type LayoutResult struct {
	// Boxes 所有节点的布局结果
	Boxes []LayoutBox

	// Root 根节点
	Root *LayoutBox

	// ContentSize 内容尺寸
	ContentSize Size

	// Dirty 脏标记
	Dirty bool

	// HitMap 命中映射表（可选）
	HitMap *HitMap
}

// Constraints 布局约束
type Constraints struct {
	// MinWidth 最小宽度
	MinWidth int

	// MaxWidth 最大宽度
	MaxWidth int

	// MinHeight 最小高度
	MinHeight int

	// MaxHeight 最大高度
	MaxHeight int
}

// NewConstraints 创建约束，并确保 Min <= Max 且值非负
func NewConstraints(minWidth, maxWidth, minHeight, maxHeight int) Constraints {
	if minWidth < 0 {
		minWidth = 0
	}
	if minHeight < 0 {
		minHeight = 0
	}
	if maxWidth < 0 {
		maxWidth = 0
	}
	if maxHeight < 0 {
		maxHeight = 0
	}
	if maxWidth < minWidth {
		minWidth = maxWidth
	}
	if maxHeight < minHeight {
		minHeight = maxHeight
	}
	return Constraints{
		MinWidth:  minWidth,
		MaxWidth:  maxWidth,
		MinHeight: minHeight,
		MaxHeight: maxHeight,
	}
}

// Tight 创建紧约束（固定尺寸）
func TightConstraints(width, height int) Constraints {
	return Constraints{
		MinWidth:  width,
		MaxWidth:  width,
		MinHeight: height,
		MaxHeight: height,
	}
}

// Loose 创建松约束（只有最小值）
func LooseConstraints(minWidth, minHeight int) Constraints {
	return Constraints{
		MinWidth:  minWidth,
		MaxWidth:  MaxInt,
		MinHeight: minHeight,
		MaxHeight: MaxInt,
	}
}

// Unbounded 创建无界约束
func UnboundedConstraints() Constraints {
	return Constraints{
		MinWidth:  0,
		MaxWidth:  MaxInt,
		MinHeight: 0,
		MaxHeight: MaxInt,
	}
}

// Width 创建宽度约束
func (c Constraints) Width(minWidth, maxWidth int) Constraints {
	return Constraints{
		MinWidth:  minWidth,
		MaxWidth:  maxWidth,
		MinHeight: c.MinHeight,
		MaxHeight: c.MaxHeight,
	}
}

// Height 创建高度约束
func (c Constraints) Height(minHeight, maxHeight int) Constraints {
	return Constraints{
		MinWidth:  c.MinWidth,
		MaxWidth:  c.MaxWidth,
		MinHeight: minHeight,
		MaxHeight: maxHeight,
	}
}

// IsTight 检查是否为紧约束
func (c Constraints) IsTight() bool {
	return c.MinWidth == c.MaxWidth && c.MinHeight == c.MaxHeight
}

// IsBounded 检查是否有界
func (c Constraints) IsBounded() bool {
	return c.MaxWidth < MaxInt || c.MaxHeight < MaxInt
}

// Constrain 约束尺寸到范围内
func (c Constraints) Constrain(width, height int) (int, int) {
	if width < c.MinWidth {
		width = c.MinWidth
	}
	if width > c.MaxWidth {
		width = c.MaxWidth
	}
	if height < c.MinHeight {
		height = c.MinHeight
	}
	if height > c.MaxHeight {
		height = c.MaxHeight
	}
	return width, height
}

// ConstrainWidth 约束宽度
func (c Constraints) ConstrainWidth(width int) int {
	if width < c.MinWidth {
		return c.MinWidth
	}
	if width > c.MaxWidth {
		return c.MaxWidth
	}
	return width
}

// ConstrainHeight 约束高度
func (c Constraints) ConstrainHeight(height int) int {
	if height < c.MinHeight {
		return c.MinHeight
	}
	if height > c.MaxHeight {
		return c.MaxHeight
	}
	return height
}

// MaxInt 最大整数值（表示无界）
const MaxInt = 1 << 30

// Measurable 可测量节点
type Measurable interface {
	Node
	// Measure 测量节点在给定约束下的理想尺寸
	Measure(constraints Constraints) Size
}

// Dirtyable 脏标记节点（可选接口）
// 节点可以实现此接口以支持增量布局优化
type Dirtyable interface {
	// IsLayoutDirty 返回节点是否需要重新布局
	IsLayoutDirty() bool
	// ClearLayoutDirty 清除布局脏标记
	ClearLayoutDirty()
	// MarkLayoutDirty 标记节点为需要布局
	MarkLayoutDirty()
}

// Marginal 外边距节点（可选接口）
// 节点可以实现此接口以提供外边距信息
type Marginal interface {
	Node
	// GetMargin 返回节点的外边距
	GetMargin() Margin
}

// MarginBox 带外边距的盒子信息
// 用于布局计算时跟踪外边距
type MarginBox struct {
	// Box 基础布局盒子
	Box *LayoutBox
	// Margin 外边距
	Margin Margin
}

// ContentBox 返回内容区域（去除 margin 后的区域）
func (mb *MarginBox) ContentBox() *LayoutBox {
	if mb.Box == nil {
		return nil
	}
	return &LayoutBox{
		ID:      mb.Box.ID,
		X:       mb.Box.X + mb.Margin.Left,
		Y:       mb.Box.Y + mb.Margin.Top,
		Width:   mb.Box.Width - mb.Margin.Horizontal(),
		Height:  mb.Box.Height - mb.Margin.Vertical(),
	}
}

// BorderBox 返回边框盒（包含 margin 的完整区域）
func (mb *MarginBox) BorderBox() *LayoutBox {
	return mb.Box
}

// =============================================================================
// Layout Engine (V3)
// =============================================================================

// Engine 布局引擎
// 负责计算节点树中所有节点的位置和尺寸
type Engine struct {
	// dirty 脏标记跟踪器
	dirty *DirtyTracker

	// stats 布局统计
	stats LayoutStats

	// cache 布局缓存
	cache *Cache

	// flexCache Flex 分布缓存
	flexCache *FlexCache

	// hitMap 命中映射表
	hitMap *HitMap
}

// NewEngine 创建新的布局引擎
func NewEngine() *Engine {
	return &Engine{
		dirty: NewDirtyTracker(),
		stats: LayoutStats{},
		cache: &Cache{
			entries: make(map[string]*CachedLayout),
			maxSize: 1000,
		},
		flexCache: NewFlexCache(),
		hitMap:   NewHitMap(),
	}
}

// Invalidate 使整个布局树失效
func (e *Engine) Invalidate() {
	e.dirty.Clear()
}

// InvalidateNode 使单个节点失效
func (e *Engine) InvalidateNode(id string) {
	e.dirty.MarkLayoutDirty(id)
}

// Layout 执行布局计算
// 输入根节点和约束，返回布局结果
func (e *Engine) Layout(root Node, constraints Constraints) *LayoutResult {
	if root == nil {
		return &LayoutResult{}
	}

	// 检查缓存
	if e.cache != nil {
		if cached := e.cache.Get(root, constraints); cached != nil {
			e.stats.CacheHits++
			return cached
		}
		e.stats.CacheMisses++
	}

	result := &LayoutResult{
		Boxes: make([]LayoutBox, 0),
		Dirty: true,
	}

	// 递归布局节点
	box := e.layoutNode(root, constraints, 0, 0)
	result.Root = box
	result.Boxes = e.collectBoxes(box)

	// 构建命中映射表
	if e.hitMap != nil {
		result.HitMap = e.hitMap
		e.hitMap.BuildFromLayoutBox(box)
	}

	// 存入缓存
	if e.cache != nil {
		e.cache.Put(root, constraints, result)
	}

	return result
}

// LayoutIncremental 执行增量布局计算
// 使用脏标记跳过干净的节点
func (e *Engine) LayoutIncremental(root Node, constraints Constraints) *LayoutResult {
	if root == nil {
		return &LayoutResult{}
	}

	// 检查缓存
	if e.cache != nil {
		if cached := e.cache.Get(root, constraints); cached != nil {
			e.stats.CacheHits++
			return cached
		}
		e.stats.CacheMisses++
	}

	result := &LayoutResult{
		Boxes: make([]LayoutBox, 0),
		Dirty: true,
	}

	// 使用增量布局
	box := e.layoutNodeIncremental(root, constraints, 0, 0)
	result.Root = box
	result.Boxes = e.collectBoxes(box)

	// 构建命中映射表
	if e.hitMap != nil {
		result.HitMap = e.hitMap
		e.hitMap.BuildFromLayoutBox(box)
	}

	// 存入缓存
	if e.cache != nil {
		e.cache.Put(root, constraints, result)
	}

	// 清除脏标记
	e.clearDirtyMarkers(root)

	return result
}

// MaxLayoutDepth is the maximum depth for layout recursion
const MaxLayoutDepth = 500

// layoutNode 递归布局单个节点
func (e *Engine) layoutNode(node Node, constraints Constraints, x, y int) *LayoutBox {
	return e.layoutNodeWithDepth(node, constraints, x, y, 0, make(map[string]bool))
}

// layoutNodeWithDepth 递归布局单个节点，带深度限制和循环检测
func (e *Engine) layoutNodeWithDepth(node Node, constraints Constraints, x, y int, depth int, visited map[string]bool) *LayoutBox {
	if node == nil {
		return nil
	}

	// 深度限制检查
	if depth > MaxLayoutDepth {
		// 达到最大深度，返回最小尺寸的盒子
		return &LayoutBox{
			ID:      node.ID(),
			X:       x,
			Y:       y,
			Width:   0,
			Height:  0,
			Children: make([]*LayoutBox, 0),
		}
	}

	// 循环检测
	nodeID := node.ID()
	if nodeID != "" {
		if visited[nodeID] {
			// 检测到循环，返回空盒子避免无限递归
			return &LayoutBox{
				ID:      node.ID() + "_cycle",
				X:       x,
				Y:       y,
				Width:   0,
				Height:  0,
				Children: make([]*LayoutBox, 0),
			}
		}
		visited[nodeID] = true
		defer delete(visited, nodeID) // 离开时清除标记
	}

	// 获取节点尺寸
	width, height := node.GetSize()

	// 如果节点实现了 Measurable 接口，测量其尺寸
	if measurable, ok := node.(Measurable); ok {
		size := measurable.Measure(constraints)
		width, height = size.Width, size.Height
	}

	// Get Layer and ZIndex from node if it implements Layered interface
	layer := GetLayerFromNode(node)
	zIndex := GetZIndexFromNode(node)

	// Get Border from node if it implements Bordered interface
	nodeBorder := GetBorderFromNode(node)

	box := &LayoutBox{
		ID:       node.ID(),
		X:        x,
		Y:        y,
		Width:    width,
		Height:   height,
		Layer:    layer,
		ZIndex:   zIndex,
		Border:   nodeBorder,
		Children: make([]*LayoutBox, 0),
	}

	// 设置节点位置和尺寸
	node.SetPosition(x, y)
	node.SetSize(width, height)

	// 计算边框偏移（用于布局子节点）
	borderOffsetX, borderOffsetY := nodeBorder.ContentOffset()

	// 检查节点是否实现了 FlexStyleProvider 接口
	// 如果是，使用 FlexLayout 进行子节点布局
	if flexProvider, ok := node.(FlexStyleProvider); ok {
		flexStyle := flexProvider.GetFlexStyle()
		if flexStyle != nil && len(node.Children()) > 0 {
			// 使用 FlexLayout 进行布局
			flex := NewFlexLayout(node.ID(), node.Children())
			flex.SetDirection(flexStyle.Direction)
			flex.SetMainAxis(flexStyle.MainAxis)
			flex.SetCrossAxis(flexStyle.CrossAxis)
			flex.SetGap(flexStyle.Gap)
			flex.SetPadding(flexStyle.Padding.Left, flexStyle.Padding.Right, flexStyle.Padding.Top, flexStyle.Padding.Bottom)

			// 设置子节点的 flex 属性
			children := node.Children()
			for i, child := range children {
				if flexChild, ok := child.(FlexChildProvider); ok {
					childFlex := flexChild.GetFlex()
					if childFlex > 0 {
						flex.SetFlex(i, childFlex, 0, 0)
					}
				}
			}

			// 布局子节点
			childBoxes := flex.LayoutChildren(width, height)
			for i, childBox := range childBoxes {
				// 递归布局子节点的子节点
				child := node.Children()[i]
				if child != nil {
					// 子节点的位置相对于父节点，并考虑边框偏移
					childConstraints := constraints
					// 应用边框偏移
					childX := x + childBox.X + borderOffsetX
					childY := y + childBox.Y + borderOffsetY
					subBox := e.layoutNodeWithDepth(child, childConstraints, childX, childY, depth+1, visited)
					if subBox != nil {
						// 使用 FlexLayout 计算的位置和尺寸，加上边框偏移
						subBox.X = childX
						subBox.Y = childY
						box.Children = append(box.Children, subBox)
					}
				}
			}
			return box
		}
	}

	// 检查节点是否实现了 GridStyleProvider 接口
	// 如果是，使用 GridLayout 进行子节点布局
	if gridProvider, ok := node.(GridStyleProvider); ok {
		gridStyle := gridProvider.GetGridStyle()
		if gridStyle != nil && (len(gridStyle.Cells) > 0 || len(node.Children()) > 0) {
			// 使用 GridLayout 进行布局
			grid := NewGridLayout(node.ID(), gridStyle)
			grid.SetChildren(node.Children())

			// 布局子节点
			childBoxes := grid.LayoutChildren(width, height)
			for i, childBox := range childBoxes {
				// 递归布局子节点的子节点
				// 需要找到对应的 child 节点
				var child Node
				if len(gridStyle.Cells) > 0 && i < len(gridStyle.Cells) {
					child = gridStyle.Cells[i].Child
				} else if i < len(node.Children()) {
					child = node.Children()[i]
				}
				if child != nil {
					childX := x + childBox.X + borderOffsetX
					childY := y + childBox.Y + borderOffsetY
					subBox := e.layoutNodeWithDepth(child, constraints, childX, childY, depth+1, visited)
					if subBox != nil {
						subBox.X = childX
						subBox.Y = childY
						box.Children = append(box.Children, subBox)
					}
				}
			}
			return box
		}
	}

	// 检查节点是否实现了 AbsoluteStyleProvider 接口
	// 如果是，使用绝对定位进行子节点布局
	if absProvider, ok := node.(AbsoluteStyleProvider); ok {
		absStyle := absProvider.GetAbsoluteStyle()
		if absStyle != nil {
			// 绝对定位容器：子元素相对于容器定位
			// 使用 absolute 节点的尺寸作为容器尺寸
			// 如果尺寸为 0，使用约束的最大值
			containerWidth := width
			containerHeight := height
			if containerWidth <= 0 {
				containerWidth = constraints.MaxWidth
			}
			if containerHeight <= 0 {
				containerHeight = constraints.MaxHeight
			}
			for _, child := range node.Children() {
				// 获取子元素尺寸
				childWidth, childHeight := child.GetSize()

				// 如果子元素实现了 Measurable，测量其尺寸
				if measurable, ok := child.(Measurable); ok {
					childConstraints := Constraints{
						MinWidth:  0,
						MaxWidth:  containerWidth,
						MinHeight: 0,
						MaxHeight: containerHeight,
					}
					size := measurable.Measure(childConstraints)
					childWidth = size.Width
					childHeight = size.Height
				}

				// 使用 AbsoluteStyle 计算子元素位置
				childX, childY := absStyle.CalculatePosition(containerWidth, containerHeight, childWidth, childHeight)

				// 递归布局子节点
				subBox := e.layoutNodeWithDepth(child, constraints, x+childX+borderOffsetX, y+childY+borderOffsetY, depth+1, visited)
				if subBox != nil {
					subBox.X = x + childX + borderOffsetX
					subBox.Y = y + childY + borderOffsetY
					box.Children = append(box.Children, subBox)
				}
			}
			return box
		}
	}

	// 默认布局：递归布局子节点（垂直方向），考虑边框偏移
	childX := x + borderOffsetX
	childY := y + borderOffsetY
	// 为子节点创建新的约束，使用节点的实际尺寸减去边框偏移
	// 这样 absolute 子节点可以使用正确的内容区域尺寸
	childConstraints := constraints
	if width > 0 && height > 0 {
		// 计算内容区域尺寸
		contentWidth := width - 2*borderOffsetX
		contentHeight := height - 2*borderOffsetY
		if contentWidth > 0 && contentHeight > 0 {
			childConstraints = Constraints{
				MinWidth:  0,
				MaxWidth:  min(constraints.MaxWidth, contentWidth),
				MinHeight: 0,
				MaxHeight: min(constraints.MaxHeight, contentHeight),
			}
		}
	}
	for _, child := range node.Children() {
		childBox := e.layoutNodeWithDepth(child, childConstraints, childX, childY, depth+1, visited)
		if childBox != nil {
			box.Children = append(box.Children, childBox)
			childY += childBox.Height
		}
	}

	return box
}

// collectBoxes 收集所有布局盒子
func (e *Engine) collectBoxes(root *LayoutBox) []LayoutBox {
	if root == nil {
		return nil
	}

	boxes := make([]LayoutBox, 0)
	e.collectBoxesRecursive(root, &boxes)
	return boxes
}

// collectBoxesRecursive 递归收集布局盒子
func (e *Engine) collectBoxesRecursive(box *LayoutBox, boxes *[]LayoutBox) {
	*boxes = append(*boxes, *box)
	for _, child := range box.Children {
		e.collectBoxesRecursive(child, boxes)
	}
}

// layoutNodeIncremental 递归布局单个节点（使用脏标记）
func (e *Engine) layoutNodeIncremental(node Node, constraints Constraints, x, y int) *LayoutBox {
	return e.layoutNodeIncrementalWithDepth(node, constraints, x, y, 0, make(map[string]bool))
}

// layoutNodeIncrementalWithDepth 递归布局单个节点，带深度限制和循环检测
func (e *Engine) layoutNodeIncrementalWithDepth(node Node, constraints Constraints, x, y int, depth int, visited map[string]bool) *LayoutBox {
	if node == nil {
		return nil
	}

	// 深度限制检查
	if depth > MaxLayoutDepth {
		return &LayoutBox{
			ID:      node.ID(),
			X:       x,
			Y:       y,
			Width:   0,
			Height:  0,
			Children: make([]*LayoutBox, 0),
		}
	}

	// 循环检测
	nodeID := node.ID()
	if nodeID != "" {
		if visited[nodeID] {
			return &LayoutBox{
				ID:      node.ID() + "_cycle",
				X:       x,
				Y:       y,
				Width:   0,
				Height:  0,
				Children: make([]*LayoutBox, 0),
			}
		}
		visited[nodeID] = true
		defer delete(visited, nodeID)
	}

	// 检查节点是否是脏的
	if !e.dirty.IsLayoutDirty(node.ID()) {
		// 节点是干净的，使用当前尺寸和位置
		width, height := node.GetSize()
		curX, curY := node.GetPosition()

		box := &LayoutBox{
			ID:       node.ID(),
			X:        curX,
			Y:        curY,
			Width:    width,
			Height:   height,
			Children: make([]*LayoutBox, 0),
		}

		// 递归处理子节点（仍然检查脏标记）
		childX := curX
		childY := curY
		for _, child := range node.Children() {
			childBox := e.layoutNodeIncrementalWithDepth(child, constraints, childX, childY, depth+1, visited)
			if childBox != nil {
				box.Children = append(box.Children, childBox)
				childY += childBox.Height
			}
		}

		return box
	}

	// 节点是脏的，需要重新布局
	width, height := node.GetSize()

	// 如果节点实现了 Measurable 接口，测量其尺寸
	if measurable, ok := node.(Measurable); ok {
		size := measurable.Measure(constraints)
		width, height = size.Width, size.Height
	}

	box := &LayoutBox{
		ID:      node.ID(),
		X:       x,
		Y:       y,
		Width:   width,
		Height:  height,
		Children: make([]*LayoutBox, 0),
	}

	// 设置节点位置和尺寸
	node.SetPosition(x, y)
	node.SetSize(width, height)

	// 递归布局子节点
	childX := x
	childY := y
	for _, child := range node.Children() {
		childBox := e.layoutNodeIncrementalWithDepth(child, constraints, childX, childY, depth+1, visited)
		if childBox != nil {
			box.Children = append(box.Children, childBox)
			childY += childBox.Height
		}
	}

	return box
}

// clearDirtyMarkers 清除节点树的脏标记
func (e *Engine) clearDirtyMarkers(node Node) {
	if node == nil {
		return
	}

	// 清除当前节点的脏标记
	e.dirty.ClearKey(node.ID())

	// 递归清除子节点的脏标记
	for _, child := range node.Children() {
		e.clearDirtyMarkers(child)
	}
}

// LayoutStats 布局统计
type LayoutStats struct {
	CacheHits   int64
	CacheMisses int64
}

// GetStats 获取布局统计
func (e *Engine) GetStats() LayoutStats {
	return e.stats
}

// Engine.Measure method should check if the node is Measurable, and if so, call its Measure method. 
// Otherwise, it should return the node's current size constrained by the input.
func (e *Engine) Measure(node Node, constraints Constraints) Size {
	if node == nil {
		return Size{}
	}

	if measurable, ok := node.(Measurable); ok {
		return measurable.Measure(constraints)
	}

	// Default: return current node size but respect the constraints.
	w, h := node.GetSize()
	w = constraints.ConstrainWidth(w)
	h = constraints.ConstrainHeight(h)
	return Size{Width: w, Height: h}
}

// =============================================================================
// Baseline Alignment Interface
// =============================================================================

// HasBaseline nodes can provide a baseline offset for alignment
type HasBaseline interface {
	Node

	// GetBaseline returns the baseline offset from the top of the node
	// For text nodes, this is typically the distance from top to the text baseline
	GetBaseline() int
}

// BaselineLayoutBox extends LayoutBox with baseline information
type BaselineLayoutBox struct {
	*LayoutBox

	// Baseline is the offset from top to the baseline
	Baseline int

	// HasBaselineInfo indicates if this box has valid baseline info
	HasBaselineInfo bool
}

// GetEffectiveBaseline returns the baseline, or estimated if not available
func (b *BaselineLayoutBox) GetEffectiveBaseline() int {
	if b.HasBaselineInfo {
		return b.Baseline
	}
	// Default baseline is 2/3 of height (typical for text)
	if b.Height > 0 {
		return b.Height * 2 / 3
	}
	return 0
}

// GetBaselineFromNode safely gets baseline from a node
func GetBaselineFromNode(node Node) int {
	if node == nil {
		return 0
	}
	if hasBaseline, ok := node.(HasBaseline); ok {
		return hasBaseline.GetBaseline()
	}
	// Default: estimate baseline as 2/3 of height
	w, h := node.GetSize()
	_ = w // unused
	if h > 0 {
		return h * 2 / 3
	}
	return 0
}
