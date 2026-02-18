package compute

import (
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// VNodePaintableAdapter - Adapts VNode to PaintableNode Interface
// =============================================================================
// This adapter allows the PaintEngine to work with VNode without direct coupling.

// VNodePaintableAdapter wraps a VNode to implement paint.PaintableNode.
type VNodePaintableAdapter struct {
	VNode rtui.VNode
}

// Ensure VNodePaintableAdapter implements paint.PaintableNode
var _ paint.PaintableNode = (*VNodePaintableAdapter)(nil)

// NewVNodePaintableAdapter creates a new adapter for a VNode.
func NewVNodePaintableAdapter(vnode rtui.VNode) *VNodePaintableAdapter {
	return &VNodePaintableAdapter{VNode: vnode}
}

// ID returns the VNode's key.
func (a *VNodePaintableAdapter) ID() string {
	if a.VNode == nil {
		return ""
	}
	return a.VNode.Key()
}

// NodeType returns the paint node type based on VNode type.
func (a *VNodePaintableAdapter) NodeType() paint.NodeType {
	if a.VNode == nil {
		return paint.NodeTypeFragment
	}
	switch a.VNode.Type() {
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

// Tag returns the VNode's tag.
func (a *VNodePaintableAdapter) Tag() string {
	if a.VNode == nil {
		return ""
	}
	return a.VNode.Tag()
}

// Style returns the VNode's style.
func (a *VNodePaintableAdapter) Style() style.Style {
	if a.VNode == nil {
		return style.Style{}
	}
	return a.VNode.Style()
}

// SetStyle sets the VNode's style.
func (a *VNodePaintableAdapter) SetStyle(s style.Style) {
	if a.VNode != nil {
		a.VNode.SetStyle(s)
	}
}

// TextContent returns the text content of the VNode.
func (a *VNodePaintableAdapter) TextContent() string {
	if a.VNode == nil {
		return ""
	}
	return rtui.GetTextContent(a.VNode)
}

// Paint delegates to the VNode's Paint method if it implements Paintable.
func (a *VNodePaintableAdapter) Paint(x, y int) []paint.DrawCmd {
	if a.VNode == nil {
		return nil
	}
	// Check if VNode implements the Paint method
	if paintable, ok := a.VNode.(interface {
		Paint(int, int) []paint.DrawCmd
	}); ok {
		return paintable.Paint(x, y)
	}
	return nil
}

// =============================================================================
// BorderInfo Implementation
// =============================================================================

// GetBorderStyle returns the border style from the VNode.
func (a *VNodePaintableAdapter) GetBorderStyle() paint.BorderStyle {
	if a.VNode == nil {
		return paint.BorderStyleNone
	}
	if bs, ok := a.VNode.(interface{ GetBorderStyle() rtui.BorderStyle }); ok {
		switch bs.GetBorderStyle() {
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

// GetBorderColor returns the border color from the VNode.
func (a *VNodePaintableAdapter) GetBorderColor() string {
	if a.VNode == nil {
		return ""
	}
	if bc, ok := a.VNode.(interface{ GetBorderColor() string }); ok {
		return bc.GetBorderColor()
	}
	return ""
}

// GetBorderLabel returns the border label from the VNode.
func (a *VNodePaintableAdapter) GetBorderLabel() string {
	if a.VNode == nil {
		return ""
	}
	if bl, ok := a.VNode.(interface{ GetBorderLabel() string }); ok {
		return bl.GetBorderLabel()
	}
	return ""
}
