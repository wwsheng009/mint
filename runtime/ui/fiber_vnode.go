package ui

import (
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/runtime/types"
)

// FiberVNode wraps a Fiber to implement VNode interface.
// This allows Fiber-first layout/render pipeline to work with existing PaintEngine.
type FiberVNode struct {
	fiber *Fiber
	id    string // Business identifier for fiber reference/positioning
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

func (f *FiberVNode) SetProps(props Props) VNode {
	if f.fiber != nil {
		f.fiber.Props = props
	}
	return f
}

// SetProp sets a single property - returns VNode for chaining (implements VNode interface)
func (f *FiberVNode) SetProp(key string, value interface{}) VNode {
	if f.fiber != nil {
		if f.fiber.Props == nil {
			f.fiber.Props = make(Props)
		}
		f.fiber.Props[key] = value
	}
	return f
}

func (f *FiberVNode) Style() style.Style {
	if f.fiber == nil {
		return style.Style{}
	}
	return f.fiber.Style
}

func (f *FiberVNode) SetStyle(s style.Style) VNode {
	if f.fiber != nil {
		f.fiber.Style = s
	}
	return f
}

func (f *FiberVNode) SetKey(key string) VNode {
	if f.fiber != nil {
		f.fiber.DiffKey = key
		f.fiber.Key = key
	}
	return f
}

// ID implements VNode - returns the business identifier for fiber reference/positioning
func (f *FiberVNode) ID() string {
	return f.id
}

// SetID implements VNode - sets the business identifier and returns VNode for chaining
func (f *FiberVNode) SetID(id string) VNode {
	f.id = id
	return f
}

func (f *FiberVNode) Children() []VNode {
	if f.fiber == nil {
		return nil
	}
	return GetFiberChildrenAsVNodes(f.fiber)
}

func (f *FiberVNode) SetChildren(children []VNode) VNode {
	return f
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

// =============================================================================
// Portal Methods - Chainable methods for Portal configuration
// =============================================================================

// SetPortalRoot implements VNode - sets the portalRoot property
func (f *FiberVNode) SetPortalRoot(portalRootID string) VNode {
	if f.fiber == nil {
		return f
	}
	if f.fiber.Props == nil {
		f.fiber.Props = make(Props)
	}
	f.fiber.Props["portalRoot"] = portalRootID
	f.fiber.Props["_portal"] = true
	return f
}

// SetAnchorTo implements VNode - sets anchorId and anchor properties
func (f *FiberVNode) SetAnchorTo(anchorID string, anchor types.Anchor) VNode {
	if f.fiber == nil {
		return f
	}
	if f.fiber.Props == nil {
		f.fiber.Props = make(Props)
	}
	f.fiber.Props["anchorId"] = anchorID
	f.fiber.Props["anchor"] = anchor
	return f
}

// SetPortalPosition implements VNode - sets the position property
func (f *FiberVNode) SetPortalPosition(position types.PositionType) VNode {
	if f.fiber == nil {
		return f
	}
	if f.fiber.Props == nil {
		f.fiber.Props = make(Props)
	}
	f.fiber.Props["position"] = position
	return f
}

// SetPortalPriority implements VNode - sets the priority property
func (f *FiberVNode) SetPortalPriority(priority int) VNode {
	if f.fiber == nil {
		return f
	}
	if f.fiber.Props == nil {
		f.fiber.Props = make(Props)
	}
	f.fiber.Props["priority"] = priority
	return f
}

// SetPortalRootId implements VNode - sets the portalRootId property
func (f *FiberVNode) SetPortalRootId(portalRootId string) VNode {
	if f.fiber == nil {
		return f
	}
	if f.fiber.Props == nil {
		f.fiber.Props = make(Props)
	}
	f.fiber.Props["portalRootId"] = portalRootId
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
