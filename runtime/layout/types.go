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

	//ContentSize 内容尺寸
	ContentSize Size

	// Dirty 脏标记
	Dirty bool
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

// =============================================================================
// Layout Engine (V3)
// =============================================================================

// Engine 布局引擎
// 负责计算节点树中所有节点的位置和尺寸
type Engine struct {
	// dirtyNodes 脏节点集合
	dirtyNodes map[string]bool

	// stats 布局统计
	stats LayoutStats

	// cache 布局缓存
	cache *Cache
}

// NewEngine 创建新的布局引擎
func NewEngine() *Engine {
	return &Engine{
		dirtyNodes: make(map[string]bool),
		stats:      LayoutStats{},
		cache: &Cache{
			entries: make(map[string]*CachedLayout),
			maxSize: 1000,
		},
	}
}

// Invalidate 使整个布局树失效
func (e *Engine) Invalidate() {
	e.dirtyNodes = make(map[string]bool)
}

// InvalidateNode 使单个节点失效
func (e *Engine) InvalidateNode(id string) {
	e.dirtyNodes[id] = true
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

	// 存入缓存
	if e.cache != nil {
		e.cache.Put(root, constraints, result)
	}

	return result
}

// layoutNode 递归布局单个节点
func (e *Engine) layoutNode(node Node, constraints Constraints, x, y int) *LayoutBox {
	if node == nil {
		return nil
	}

	// 获取节点尺寸
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
		childBox := e.layoutNode(child, constraints, childX, childY)
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
