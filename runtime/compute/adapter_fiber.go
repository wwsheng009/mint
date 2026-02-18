package compute

import (
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// FiberPaintableAdapter - Adapts Fiber to PaintableNode Interface
// =============================================================================
// This adapter allows the PaintEngine to work with Fiber without direct coupling.

// FiberPaintableAdapter wraps a Fiber to implement paint.PaintableNode.
type FiberPaintableAdapter struct {
	Fiber *rtui.Fiber
}

// Ensure FiberPaintableAdapter implements paint.PaintableNode
var _ paint.PaintableNode = (*FiberPaintableAdapter)(nil)

// NewFiberPaintableAdapter creates a new adapter for a Fiber.
func NewFiberPaintableAdapter(fiber *rtui.Fiber) *FiberPaintableAdapter {
	return &FiberPaintableAdapter{Fiber: fiber}
}

// ID returns the Fiber's DiffKey (or Key for compatibility).
func (a *FiberPaintableAdapter) ID() string {
	if a.Fiber == nil {
		return ""
	}
	if a.Fiber.DiffKey != "" {
		return a.Fiber.DiffKey
	}
	return a.Fiber.Key
}

// NodeType returns the paint node type based on Fiber type.
func (a *FiberPaintableAdapter) NodeType() paint.NodeType {
	if a.Fiber == nil {
		return paint.NodeTypeFragment
	}
	switch a.Fiber.Type {
	case rtui.VNodeText:
		return paint.NodeTypeText
	case rtui.VNodeElement:
		return paint.NodeTypeElement
	case rtui.VNodeComponent:
		return paint.NodeTypeComponent
	default:
		return paint.NodeTypeFragment
	}
}

// Tag returns the Fiber's tag.
func (a *FiberPaintableAdapter) Tag() string {
	if a.Fiber == nil {
		return ""
	}
	return a.Fiber.Tag
}

// Style returns the Fiber's style.
func (a *FiberPaintableAdapter) Style() style.Style {
	if a.Fiber == nil {
		return style.Style{}
	}
	// Fiber-first: use Style field directly
	return a.Fiber.Style
}

// SetStyle sets the Fiber's style.
func (a *FiberPaintableAdapter) SetStyle(s style.Style) {
	if a.Fiber != nil {
		a.Fiber.Style = s
	}
}

// TextContent returns the text content from the Fiber.
func (a *FiberPaintableAdapter) TextContent() string {
	if a.Fiber == nil {
		return ""
	}
	// Priority: MemoizedState (set by CreateFiber/completeWork)
	if a.Fiber.MemoizedState != nil {
		if s, ok := a.Fiber.MemoizedState.(string); ok {
			return s
		}
	}
	// Fallback: Props["content"]
	if a.Fiber.Props != nil {
		if c, ok := a.Fiber.Props["content"]; ok {
			if s, ok := c.(string); ok {
				return s
			}
		}
	}
	return ""
}

// Paint delegates to the Fiber's Paint method.
// Fiber-first Architecture (priority order):
// 1. Fiber.Instance.Paint() (PaintableInstance)
// 2. PaintRegistry by Tag (simple components)
// 3. Fiber.PaintFunc (legacy transition path)
func (a *FiberPaintableAdapter) Paint(x, y int) []paint.DrawCmd {
	if a.Fiber == nil {
		return nil
	}

	// Primary Path: Use Fiber.Instance (Fiber-first)
	// The Instance persists across renders and holds state
	if a.Fiber.Instance != nil {
		if inst, ok := rtui.AsPaintableInstance(a.Fiber.Instance); ok {
			return inst.Paint(x, y)
		}
	}

	// Fallback Path 1: Use PaintRegistry (simple components)
	if fn := rtui.GetPaint(a.Fiber.Tag); fn != nil {
		return fn(a.Fiber.Props, a.Fiber.Style, x, y)
	}

	// Fallback Path 2: Use Fiber.FocusableVNode (legacy)
	if a.Fiber.FocusableVNode != nil {
		if paintable, ok := a.Fiber.FocusableVNode.(interface {
			Paint(int, int) []paint.DrawCmd
		}); ok {
			return paintable.Paint(x, y)
		}
	}

	// Fallback Path 3: Use Fiber.PaintFunc (Legacy)
	if a.Fiber.PaintFunc != nil {
		if paintable, ok := a.Fiber.PaintFunc.(interface {
			Paint(int, int) []paint.DrawCmd
		}); ok {
			return paintable.Paint(x, y)
		}
	}

	return nil
}

// =============================================================================
// BorderInfo Implementation
// =============================================================================

// GetBorderStyle returns the border style from the Fiber's Props.
func (a *FiberPaintableAdapter) GetBorderStyle() paint.BorderStyle {
	if a.Fiber == nil || a.Fiber.Props == nil {
		return paint.BorderStyleNone
	}
	if bs, ok := a.Fiber.Props["borderStyle"].(rtui.BorderStyle); ok {
		switch bs {
		case rtui.BorderSingle:
			return paint.BorderStyleSingle
		case rtui.BorderDouble:
			return paint.BorderStyleDouble
		case rtui.BorderRounded:
			return paint.BorderStyleRounded
		default:
			return paint.BorderStyleNone
		}
	}
	return paint.BorderStyleNone
}

// GetBorderColor returns the border color from the Fiber's Props.
func (a *FiberPaintableAdapter) GetBorderColor() string {
	if a.Fiber == nil || a.Fiber.Props == nil {
		return ""
	}
	if bc, ok := a.Fiber.Props["borderColor"].(string); ok {
		return bc
	}
	return ""
}

// GetBorderLabel returns the border label from the Fiber's Props.
func (a *FiberPaintableAdapter) GetBorderLabel() string {
	if a.Fiber == nil || a.Fiber.Props == nil {
		return ""
	}
	if bl, ok := a.Fiber.Props["borderLabel"].(string); ok {
		return bl
	}
	return ""
}
