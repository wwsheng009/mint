package ui

import (
	"github.com/wwsheng009/mint/runtime/style"
)

// FiberVNode wraps a Fiber to implement VNode interface.
// This allows Fiber-first layout/render pipeline to work with existing PaintEngine.
type FiberVNode struct {
	fiber *Fiber
}

// NewFiberVNode creates a VNode wrapper around a Fiber.
func NewFiberVNode(fiber *Fiber) *FiberVNode {
	return &FiberVNode{fiber: fiber}
}

// Fiber returns the underlying Fiber.
func (f *FiberVNode) Fiber() *Fiber {
	return f.fiber
}

// VNode interface implementation

func (f *FiberVNode) Type() VNodeType {
	if f.fiber == nil {
		return VNodeElement
	}
	return f.fiber.Type
}

func (f *FiberVNode) Tag() string {
	if f.fiber == nil {
		return ""
	}
	return f.fiber.Tag
}

func (f *FiberVNode) Key() string {
	if f.fiber == nil {
		return ""
	}
	return f.fiber.DiffKey
}

func (f *FiberVNode) Props() Props {
	if f.fiber == nil {
		return nil
	}
	return f.fiber.Props
}

func (f *FiberVNode) SetProps(props Props) {
	if f.fiber != nil {
		f.fiber.Props = props
	}
}

func (f *FiberVNode) Style() style.Style {
	if f.fiber == nil {
		return style.Style{}
	}
	return f.fiber.Style
}

func (f *FiberVNode) SetStyle(s style.Style) {
	if f.fiber != nil {
		f.fiber.Style = s
	}
}

func (f *FiberVNode) SetKey(key string) {
	if f.fiber != nil {
		f.fiber.DiffKey = key
		f.fiber.Key = key
	}
}

func (f *FiberVNode) Children() []VNode {
	if f.fiber == nil {
		return nil
	}
	return GetFiberChildrenAsVNodes(f.fiber)
}

func (f *FiberVNode) SetChildren(children []VNode) {
}

func (f *FiberVNode) GetLayer() Layer {
	if f.fiber == nil {
		return LayerBase
	}
	return f.fiber.Layer
}

func (f *FiberVNode) SetLayer(layer Layer) VNode {
	if f.fiber != nil {
		f.fiber.Layer = layer
	}
	return f
}

func (f *FiberVNode) Clone() VNode {
	return NewFiberVNode(f.fiber)
}

// GetFiberChildrenAsVNodes converts Fiber children to VNode slice.
func GetFiberChildrenAsVNodes(fiber *Fiber) []VNode {
	if fiber == nil {
		return nil
	}

	children := fiber.GetChildFibers()
	if len(children) == 0 {
		return nil
	}

	vnodes := make([]VNode, len(children))
	for i, child := range children {
		vnodes[i] = NewFiberVNode(child)
	}
	return vnodes
}
