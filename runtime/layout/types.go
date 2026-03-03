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

// PropsIDProvider 业务标识提供者接口
// 节点可以实现此接口以提供业务标识符 (PropsID)
type PropsIDProvider interface {
	// GetPropsID 返回节点的业务标识符
	// 该标识符来自 Fiber.ID，用于业务引用和定位
	GetPropsID() string
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
	// ID 节点ID (运行时标识，fmt.Sprintf("%d", Fiber.NodeID))
	ID string

	// Tag 组件标签 (来自 Fiber.Tag，如 "vstack", "panel", "text")
	// 用于调试和类型识别
	Tag string

	// PropsID 业务标识 (来自 Fiber.ID，由 SetID() 设置)
	// 用于业务引用和定位，如 Portal.anchorId
	PropsID string

	// X, Y 位置（相对于父节点）
	X int
	Y int

	// ✨ Phase 1.2: AbsX, AbsY 全局坐标（相对于屏幕/Root）
	AbsX int
	AbsY int

	// Width, Height 尺寸
	Width  int
	Height int

	// Baseline 基线（用于文本对齐）
	Baseline int

	// Layer 渲染层（用于多层渲染）
	Layer Layer

	// ZIndex 层内排序索引
	ZIndex int

	// ✨ Phase 1.1: ShouldCenter 是否需要居中（用于 Modal）
	// 检测到 Modal 且未设置明确位置时为 true
	ShouldCenter bool

	// BoxModel 统一的盒模型信息（包含 margin, padding, border）
	// 注意：BoxModel 主要在布局计算过程中使用，不总是需要存储在最终结果中
	BoxModel BoxModel

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

	// LayerManager 管理 layer 坐标和位置
	// 负责收集各个layer的信息，统一把非base的layer的坐标对齐到0,0
	LayerManager *LayerManager
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

	// layerManager 管理 layer 坐标和位置
	layerManager *LayerManager

	// ✨ Phase 3.3: overlayManager 管理 Portal 跨树挂载
	// 用于处理 Modal/Tooltip 等需要挂载到不同位置的组件
	overlayManager *OverlayManager

	// ✨ viewportConstraints 保存根节点的原始约束
	// 用于 Fixed 定位节点计算正确的 viewport 参考系
	viewportConstraints Constraints
}


// MaxLayoutDepth is the maximum depth for layout recursion
const MaxLayoutDepth = 500


// LayoutStats 布局统计
type LayoutStats struct {
	CacheHits   int64
	CacheMisses int64
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
