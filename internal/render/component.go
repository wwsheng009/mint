package render

import (
	"github.com/wwsheng009/mint/runtime/paint"
)

// Component 组件标准接口
//
// 参考: framework/docs/ui/idea/idea4_comp.md
// 所有可渲染组件都应实现此接口
type Component interface {
	// ID 返回组件唯一标识符
	ID() string

	// Type 返回组件类型标识
	Type() string

	// Mount 挂载组件
	Mount(ctx Context) error

	// Update 更新组件
	Update(ctx Context) error

	// Unmount 卸载组件
	Unmount(ctx Context) error

	// Measure 测量组件尺寸
	// 给定约束条件，返回组件的期望尺寸
	Measure(constraints Constraints) Size

	// Paint 绘制组件
	// 使用提供的 PaintContext 绘制组件内容
	Paint(ctx PaintContext)
}

// Context 组件上下文接口
type Context interface {
	// Buffer 获取绘制缓冲区
	Buffer() *paint.Buffer

	// Bounds 获取可用边界
	Bounds() paint.Rect

	// SetValue 设置上下文值
	SetValue(key string, value any)

	// GetValue 获取上下文值
	GetValue(key string) (any, bool)
}

// PaintContext 绘制上下文接口
// 扩展自 runtime/paint.PaintContext
type PaintContext interface {
	Context

	// Painter 获取绘制器
	Painter() *paint.Painter
}

// SimplePaintContext 简单的绘制上下文实现
type SimplePaintContext struct {
	buf    *paint.Buffer
	bounds paint.Rect
	values map[string]any
}

// NewSimplePaintContext 创建简单绘制上下文
func NewSimplePaintContext(buf *paint.Buffer, bounds paint.Rect) *SimplePaintContext {
	return &SimplePaintContext{
		buf:    buf,
		bounds: bounds,
		values: make(map[string]any),
	}
}

// Buffer 返回缓冲区
func (c *SimplePaintContext) Buffer() *paint.Buffer {
	return c.buf
}

// Bounds 返回边界
func (c *SimplePaintContext) Bounds() paint.Rect {
	return c.bounds
}

// SetValue 设置值
func (c *SimplePaintContext) SetValue(key string, value any) {
	c.values[key] = value
}

// GetValue 获取值
func (c *SimplePaintContext) GetValue(key string) (any, bool) {
	v, ok := c.values[key]
	return v, ok
}

// Painter 返回绘制器
func (c *SimplePaintContext) Painter() *paint.Painter {
	return paint.NewPainter(paint.NewPaintContext(c.buf, c.bounds))
}

// Constraints 布局约束
type Constraints struct {
	MinWidth  int
	MaxWidth  int
	MinHeight int
	MaxHeight int
}

// NewConstraints 创建约束
func NewConstraints(minW, maxW, minH, maxH int) Constraints {
	return Constraints{
		MinWidth:  minW,
		MaxWidth:  maxW,
		MinHeight: minH,
		MaxHeight: maxH,
	}
}

// Unbounded 创建无边界约束
func Unbounded() Constraints {
	const maxInt = int(^uint(0) >> 1)
	return Constraints{
		MaxWidth:  maxInt,
		MaxHeight: maxInt,
	}
}

// Tight 创建紧约束 (固定尺寸)
func Tight(width, height int) Constraints {
	return Constraints{
		MinWidth:  width,
		MaxWidth:  width,
		MinHeight: height,
		MaxHeight: height,
	}
}

// Width 创建宽度约束
func (c Constraints) Width(minW, maxW int) Constraints {
	return Constraints{
		MinWidth:  minW,
		MaxWidth:  maxW,
		MinHeight: c.MinHeight,
		MaxHeight: c.MaxHeight,
	}
}

// Height 创建高度约束
func (c Constraints) Height(minH, maxH int) Constraints {
	return Constraints{
		MinWidth:  c.MinWidth,
		MaxWidth:  c.MaxWidth,
		MinHeight: minH,
		MaxHeight: maxH,
	}
}

// Size 组件尺寸
type Size struct {
	Width  int
	Height int
}

// Zero 零尺寸
func ZeroSize() Size {
	return Size{}
}

// Infinite 无限尺寸
func InfiniteSize() Size {
	const maxInt = int(^uint(0) >> 1)
	return Size{Width: maxInt, Height: maxInt}
}

// NewSize 创建尺寸
func NewSize(width, height int) Size {
	return Size{Width: width, Height: height}
}
