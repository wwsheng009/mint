package compute

import (
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// ComputedBox to PaintableBox Conversion
// =============================================================================
// These methods convert compute.ComputedBox to paint.PaintableBox,
// enabling the PaintEngine to work with an abstracted interface.

// AsPaintable converts a ComputedBox to a PaintableBox.
// This is the main entry point for PaintEngine to consume layout results.
func (cb *ComputedBox) AsPaintable() *paint.PaintableBox {
	if cb == nil {
		return nil
	}

	box := &paint.PaintableBox{
		X:            cb.Box.X,
		Y:            cb.Box.Y,
		Width:        cb.Box.Width,
		Height:       cb.Box.Height,
		RenderedText: cb.RenderedText,
		NaturalWidth: cb.NaturalWidth,
		LayoutDirty:  cb.LayoutDirty,
		LayoutHash:   cb.LayoutHash,
		NodeID:       cb.NodeID,
		Layer:        int(cb.Layer),
		ZIndex:       0, // ZIndex can be set separately if needed
		Children:     make([]*paint.PaintableBox, 0, len(cb.Children)),
	}

	// Create the appropriate adapter for the node
	if cb.VNode != nil {
		box.Node = NewVNodePaintableAdapter(cb.VNode)
		// Get border info from VNode
		cb.copyBorderInfoToPaintableBox(box)
	} else if cb.ChildFiber != nil {
		box.Node = NewFiberPaintableAdapter(cb.ChildFiber)
		// Get border info from Fiber's Props
		cb.copyFiberBorderInfoToPaintableBox(box)
	}

	// Recursively convert children
	for _, child := range cb.Children {
		if childBox := child.AsPaintable(); childBox != nil {
			childBox.Parent = box
			box.Children = append(box.Children, childBox)
		}
	}

	return box
}

// copyBorderInfoToPaintableBox copies border info from VNode to PaintableBox.
func (cb *ComputedBox) copyBorderInfoToPaintableBox(box *paint.PaintableBox) {
	if cb.VNode == nil {
		return
	}

	// Get border style
	if bs, ok := cb.VNode.(interface{ GetBorderStyle() rtui.BorderStyle }); ok {
		switch bs.GetBorderStyle() {
		case rtui.BorderSingle:
			box.BorderStyle = paint.BorderStyleSingle
		case rtui.BorderDouble:
			box.BorderStyle = paint.BorderStyleDouble
		case rtui.BorderRounded:
			box.BorderStyle = paint.BorderStyleRounded
		}
	}

	// Get border color
	if bc, ok := cb.VNode.(interface{ GetBorderColor() string }); ok {
		box.BorderColor = bc.GetBorderColor()
	}

	// Get border label
	if bl, ok := cb.VNode.(interface{ GetBorderLabel() string }); ok {
		box.BorderLabel = bl.GetBorderLabel()
	}
}

// copyFiberBorderInfoToPaintableBox copies border info from Fiber.Props to PaintableBox.
func (cb *ComputedBox) copyFiberBorderInfoToPaintableBox(box *paint.PaintableBox) {
	if cb.ChildFiber == nil || cb.ChildFiber.Props == nil {
		return
	}

	// Get border style
	if bs, ok := cb.ChildFiber.Props["borderStyle"].(rtui.BorderStyle); ok {
		switch bs {
		case rtui.BorderSingle:
			box.BorderStyle = paint.BorderStyleSingle
		case rtui.BorderDouble:
			box.BorderStyle = paint.BorderStyleDouble
		case rtui.BorderRounded:
			box.BorderStyle = paint.BorderStyleRounded
		}
	}

	// Get border color
	if bc, ok := cb.ChildFiber.Props["borderColor"].(string); ok {
		box.BorderColor = bc
	}

	// Get border label
	if bl, ok := cb.ChildFiber.Props["borderLabel"].(string); ok {
		box.BorderLabel = bl
	}
}

// =============================================================================
// ComputedLayout to PaintableLayout Conversion
// =============================================================================

// AsPaintableLayout converts a ComputedLayout to a PaintableLayout.
// This is the main entry point for PaintEngine.Paint().
func (cl *ComputedLayout) AsPaintableLayout() *paint.PaintableLayout {
	if cl == nil {
		return nil
	}

	layout := paint.NewPaintableLayout(cl.Root.AsPaintable())

	// Preserve HitMap reference (HitMap is already decoupled)
	if cl.HitMap != nil {
		layout.HitMap = cl.HitMap
	}

	return layout
}

// =============================================================================
// Batch Conversion Utilities
// =============================================================================

// ConvertComputedBoxes converts a slice of ComputedBox to PaintableBox.
func ConvertComputedBoxes(boxes []*ComputedBox) []*paint.PaintableBox {
	if boxes == nil {
		return nil
	}
	result := make([]*paint.PaintableBox, 0, len(boxes))
	for _, box := range boxes {
		if pb := box.AsPaintable(); pb != nil {
			result = append(result, pb)
		}
	}
	return result
}

// ConvertComputedLayout converts a ComputedLayout pointer to PaintableLayout.
// Returns nil if the input is nil.
func ConvertComputedLayout(layout *ComputedLayout) *paint.PaintableLayout {
	if layout == nil {
		return nil
	}
	return layout.AsPaintableLayout()
}

// =============================================================================
// Convenience Methods for Testing
// =============================================================================

// NewTestPaintableBox creates a simple PaintableBox for testing purposes.
// This is useful for unit tests that don't need full VNode/Fiber setup.
func NewTestPaintableBox(id string, x, y, w, h int) *paint.PaintableBox {
	return paint.NewPaintableBoxBuilder().
		WithNode(&testPaintableNode{id: id}).
		WithBounds(x, y, w, h).
		Build()
}

// testPaintableNode is a minimal PaintableNode implementation for testing.
type testPaintableNode struct {
	id string
}

func (n *testPaintableNode) ID() string                          { return n.id }
func (n *testPaintableNode) NodeType() paint.NodeType            { return paint.NodeTypeElement }
func (n *testPaintableNode) Tag() string                         { return "test" }
func (n *testPaintableNode) Style() style.Style                  { return style.Style{} }
func (n *testPaintableNode) SetStyle(_ style.Style)              {}
func (n *testPaintableNode) TextContent() string                 { return "" }
func (n *testPaintableNode) Paint(_, _ int) []paint.DrawCmd      { return nil }
