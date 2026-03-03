package ui

import (
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/layout"
)

// VNodeAdapter 让 VNode 实现 layout.Node 接口
// 这样 App 可以从 VNode 树（包括 Inspector overlay）构建 HitMap
type VNodeAdapter struct {
	VNode VNode
}

// ID 实现 layout.Node
func (a *VNodeAdapter) ID() string {
	if key := a.VNode.Key(); key != "" {
		return key
	}
	if tagger, ok := a.VNode.(interface{ Tag() string }); ok {
		return tagger.Tag()
	}
	return a.VNode.Type().String()
}

// Type 实现 layout.Node
func (a *VNodeAdapter) Type() string {
	return a.VNode.Type().String()
}

// Children 实现 layout.Node
func (a *VNodeAdapter) Children() []layout.Node {
	children := a.VNode.Children()
	nodes := make([]layout.Node, 0, len(children))
	for _, child := range children {
		nodes = append(nodes, &VNodeAdapter{VNode: child})
	}
	return nodes
}

// GetPosition 实现 layout.Node
func (a *VNodeAdapter) GetPosition() (x, y int) {
	if boundsGetter, ok := a.VNode.(interface{ GetBounds() [4]int }); ok {
		bounds := boundsGetter.GetBounds()
		return bounds[0], bounds[1] // x, y
	}
	return 0, 0
}

// SetPosition 实现 layout.Node
func (a *VNodeAdapter) SetPosition(x, y int) {
	// VNode 的位置由布局引擎设置，不需要手动设置
}

// GetSize 实现 layout.Node
func (a *VNodeAdapter) GetSize() (width, height int) {
	if boundsGetter, ok := a.VNode.(interface{ GetBounds() [4]int }); ok {
		bounds := boundsGetter.GetBounds()
		return bounds[2], bounds[3] // width, height
	}
	return 0, 0
}

// SetSize 实现 layout.Node
func (a *VNodeAdapter) SetSize(w, h int) {
	// VNode 的大小由布局引擎设置，不需要手动设置
}

// GetWidth 实现 layout.Node
func (a *VNodeAdapter) GetWidth() int {
	if _, w := a.GetSize(); w > 0 {
		return w
	}
	return 0
}

// GetHeight 实现 layout.Node
func (a *VNodeAdapter) GetHeight() int {
	if _, h := a.GetSize(); h > 0 {
		return h
	}
	return 0
}

// Measure 实现 layout.Measurable 接口
// 这使得布局引擎能够调用VNode的Measure方法来计算正确的尺寸
func (a *VNodeAdapter) Measure(constraints layout.Constraints) layout.Size {
	// 获取runtime风格的约束
	runtimeConstraints := runtime.BoxConstraints{
		MinWidth:  constraints.MinWidth,
		MaxWidth:  constraints.MaxWidth,
		MinHeight: constraints.MinHeight,
		MaxHeight: constraints.MaxHeight,
	}

	// 尝试调用VNode的Measure方法（如果实现了）
	if measurable, ok := a.VNode.(interface {
		Measure(runtime.BoxConstraints) runtime.Size
	}); ok {
		size := measurable.Measure(runtimeConstraints)
		return layout.Size{Width: size.Width, Height: size.Height}
	}

	// 回退到使用GetSize
	if w, h := a.GetSize(); w > 0 || h > 0 {
		return layout.Size{Width: w, Height: h}
	}

	// 最后回退到约束的最小值
	return layout.Size{Width: runtimeConstraints.MinWidth, Height: runtimeConstraints.MinHeight}
}

// AsLayoutNode 将 VNode 转换为 layout.Node
func AsLayoutNode(vnode VNode) layout.Node {
	return &VNodeAdapter{VNode: vnode}
}
