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
// Fiber-first Architecture:
// 1. Primary: Use Fiber.PaintFunc (the recommended Fiber-first approach)
// 2. Fallback: Use Fiber.FocusableVNode (deprecated, transition only)
// 3. Last resort: Check Fiber.ComponentInstance
func (a *FiberPaintableAdapter) Paint(x, y int) []paint.DrawCmd {
	if a.Fiber == nil {
		return nil
	}

	// Primary Path: Use Fiber.PaintFunc (Fiber-first)
	// PaintFunc stores the VNode reference that has Paint method
	// This is extracted during CreateFiber from VNode
	if a.Fiber.PaintFunc != nil {
		// PaintFunc stores the VNode itself (transition approach)
		// Type assert to Paintable interface
		if paintable, ok := a.Fiber.PaintFunc.(interface {
			Paint(int, int) []paint.DrawCmd
		}); ok {
			return paintable.Paint(x, y)
		}
	}

	// Fallback Path 1: Check if FocusableVNode has Paint method
	// DEPRECATED: This field is marked deprecated in Fiber struct
	// Will be removed once Fiber-first migration is complete
	if a.Fiber.FocusableVNode != nil {
		if paintable, ok := a.Fiber.FocusableVNode.(interface {
			Paint(int, int) []paint.DrawCmd
		}); ok {
			return paintable.Paint(x, y)
		}
	}

	// Fallback Path 2: Check if ComponentInstance has Paint method
	// Future path for components with state
	if a.Fiber.ComponentInstance != nil {
		if paintable, ok := a.Fiber.ComponentInstance.(interface {
			Paint(int, int) []paint.DrawCmd
		}); ok {
			return paintable.Paint(x, y)
		}
	}

	// No paint logic available for this Fiber node
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
