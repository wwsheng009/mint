// Package render provides rendering pipeline components
package render

import (
	"fmt"

	"github.com/wwsheng009/mint/internal/reconciler"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// FiberToPaintableConverter - Converts LayoutBox + Fiber to PaintableBox
// =============================================================================
// This is the SINGLE conversion step in the simplified pipeline:
//
//	Fiber → LayoutBox → PaintableBox → Paint
//	        (layout)    (paint)
//
// Design Principles:
// 1. LayoutBox is PURE layout data (no paint dependencies)
// 2. PaintableBox is PURE paint data (no layout dependencies)
// 3. This converter bridges them in ONE step

// FiberToPaintableConverter converts LayoutBox tree to PaintableBox tree
// by combining layout data with Fiber runtime data.
type FiberToPaintableConverter struct {
	// fiberMap provides quick Fiber lookup by key
	// Key: Fiber.DiffKey or Fiber.Key
	fiberMap map[string]*reconciler.Fiber
}

// NewFiberToPaintableConverter creates a new converter.
// It builds an index of all Fibers for efficient lookup during conversion.
func NewFiberToPaintableConverter(rootFiber *reconciler.Fiber) *FiberToPaintableConverter {
	c := &FiberToPaintableConverter{
		fiberMap: make(map[string]*reconciler.Fiber),
	}
	
	// Build fiber index
	if rootFiber != nil {
		c.buildFiberMap(rootFiber)
	}
	
	return c
}

// buildFiberMap recursively indexes all Fibers by multiple keys
// Keys used:
// 1. DiffKey (primary)
// 2. Key (alias for DiffKey)
// 3. NodeID as string (for matching with LayoutBox.ID from FiberToNodeAdapter)
func (c *FiberToPaintableConverter) buildFiberMap(fiber *reconciler.Fiber) {
	if fiber == nil {
		return
	}
	
	// Index by DiffKey (primary)
	if fiber.DiffKey != "" {
		c.fiberMap[fiber.DiffKey] = fiber
	}
	// Also index by Key (alias)
	if fiber.Key != "" && fiber.Key != fiber.DiffKey {
		c.fiberMap[fiber.Key] = fiber
	}
	// Index by NodeID string (for matching with FiberToNodeAdapter.ID())
	nodeIDKey := fmt.Sprintf("%d", fiber.NodeID)
	c.fiberMap[nodeIDKey] = fiber
	
	// Recursively index children
	for child := fiber.Child; child != nil; child = child.Sibling {
		c.buildFiberMap(child)
	}
}

// Convert converts a LayoutBox tree to a PaintableBox tree.
// This is the main entry point for the converter.
func (c *FiberToPaintableConverter) Convert(
	lbox *layout.LayoutBox,
	parent *paint.PaintableBox,
) *paint.PaintableBox {
	if lbox == nil {
		return nil
	}
	
	// Create PaintableBox with layout data
	pbox := &paint.PaintableBox{
		X:       lbox.X,
		Y:       lbox.Y,
		Width:   lbox.Width,
		Height:  lbox.Height,
		Layer:   convertLayoutLayerToInt(lbox.Layer),
		ZIndex:  lbox.ZIndex,
		Parent:  parent,
		Children: make([]*paint.PaintableBox, 0, len(lbox.Children)),
	}
	
	// Find matching Fiber and fill paint-specific data
	if fiber := c.findFiber(lbox.ID); fiber != nil {
		c.fillFromFiber(pbox, fiber)
	}
	
	// Recursively convert children
	for _, childLBox := range lbox.Children {
		childPBox := c.Convert(childLBox, pbox)
		if childPBox != nil {
			pbox.Children = append(pbox.Children, childPBox)
		}
	}
	
	return pbox
}

// ConvertToLayout wraps the result in a PaintableLayout
func (c *FiberToPaintableConverter) ConvertToLayout(lbox *layout.LayoutBox) *paint.PaintableLayout {
	root := c.Convert(lbox, nil)
	return paint.NewPaintableLayout(root)
}

// findFiber finds a Fiber by ID, supporting both DiffKey and NodeID formats
func (c *FiberToPaintableConverter) findFiber(id string) *reconciler.Fiber {
	if id == "" {
		return nil
	}
	
	// Strategy 1: Direct match by DiffKey
	if f, ok := c.fiberMap[id]; ok {
		return f
	}
	
	// Strategy 2: Match by NodeID format (e.g., "95")
	for _, f := range c.fiberMap {
		if fmt.Sprintf("%d", f.NodeID) == id {
			return f
		}
	}
	
	return nil
}

// fillFromFiber populates PaintableBox with data from Fiber
func (c *FiberToPaintableConverter) fillFromFiber(pbox *paint.PaintableBox, fiber *reconciler.Fiber) {
	// Identity
	pbox.NodeID = fiber.NodeID
	pbox.DiffKey = fiber.DiffKey
	
	// PaintableNode interface (wraps Fiber)
	pbox.Node = NewFiberPaintableNode(fiber)
	
	// Border info from Props
	if fiber.Props != nil {
		pbox.BorderStyle = getBorderStyleFromProps(fiber.Props)
		pbox.BorderColor = getBorderProp(fiber.Props, "borderColor")
		pbox.BorderLabel = getBorderProp(fiber.Props, "borderLabel")
	}
}

// =============================================================================
// FiberPaintableNode - Wraps Fiber to implement PaintableNode
// =============================================================================

// FiberPaintableNode wraps a Fiber to implement paint.PaintableNode interface
type FiberPaintableNode struct {
	fiber *reconciler.Fiber
}

// NewFiberPaintableNode creates a new PaintableNode wrapper for a Fiber
func NewFiberPaintableNode(fiber *reconciler.Fiber) *FiberPaintableNode {
	return &FiberPaintableNode{fiber: fiber}
}

// Ensure interface implementation
var _ paint.PaintableNode = (*FiberPaintableNode)(nil)

// ID returns the Fiber's DiffKey
func (n *FiberPaintableNode) ID() string {
	if n.fiber == nil {
		return ""
	}
	return n.fiber.DiffKey
}

// NodeType returns the paint node type based on Fiber type
func (n *FiberPaintableNode) NodeType() paint.NodeType {
	if n.fiber == nil {
		return paint.NodeTypeFragment
	}
	switch n.fiber.Type {
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

// Tag returns the Fiber's tag
func (n *FiberPaintableNode) Tag() string {
	if n.fiber == nil {
		return ""
	}
	return n.fiber.Tag
}

// Style returns the Fiber's style
func (n *FiberPaintableNode) Style() style.Style {
	if n.fiber == nil {
		return style.Style{}
	}
	return n.fiber.Style
}

// SetStyle sets the Fiber's style (for background inheritance)
func (n *FiberPaintableNode) SetStyle(s style.Style) {
	if n.fiber != nil {
		n.fiber.Style = s
	}
}

// TextContent returns text content from the Fiber
func (n *FiberPaintableNode) TextContent() string {
	if n.fiber == nil {
		return ""
	}
	// Check MemoizedState first (set for text nodes)
	if content, ok := n.fiber.MemoizedState.(string); ok {
		return content
	}
	// Fallback to Props
	if n.fiber.Props != nil {
		if content, ok := n.fiber.Props["content"].(string); ok {
			return content
		}
	}
	return ""
}

// Paint delegates to the Fiber's PaintFunc or returns nil
func (n *FiberPaintableNode) Paint(x, y int) []paint.DrawCmd {
	if n.fiber == nil || n.fiber.PaintFunc == nil {
		return nil
	}
	
	// PaintFunc stores the VNode reference (transitional approach)
	// Type assert to Paintable interface
	if paintable, ok := n.fiber.PaintFunc.(interface {
		Paint(int, int) []paint.DrawCmd
	}); ok {
		return paintable.Paint(x, y)
	}
	
	return nil
}

// =============================================================================
// Helper Functions
// =============================================================================

// convertLayoutLayerToInt converts layout.Layer to int
func convertLayoutLayerToInt(l layout.Layer) int {
	switch l {
	case layout.LayerBase:
		return 0
	case layout.LayerDropdown, layout.LayerSticky, layout.LayerFixed:
		return 1 // Overlay
	case layout.LayerModalBackdrop, layout.LayerModal:
		return 2 // Modal
	case layout.LayerPopover, layout.LayerTooltip:
		return 3 // Tooltip
	default:
		return 0
	}
}

// getBorderStyleFromProps extracts border style from Props
func getBorderStyleFromProps(props rtui.Props) paint.BorderStyle {
	if props == nil {
		return paint.BorderStyleNone
	}
	
	// Try string format
	if s, ok := props["borderStyle"].(string); ok {
		switch s {
		case "single":
			return paint.BorderStyleSingle
		case "double":
			return paint.BorderStyleDouble
		case "rounded":
			return paint.BorderStyleRounded
		}
	}
	
	// Try rtui.BorderStyle format
	if bs, ok := props["borderStyle"].(rtui.BorderStyle); ok {
		switch bs {
		case rtui.BorderSingle:
			return paint.BorderStyleSingle
		case rtui.BorderDouble:
			return paint.BorderStyleDouble
		case rtui.BorderRounded:
			return paint.BorderStyleRounded
		}
	}
	
	return paint.BorderStyleNone
}

// getBorderProp extracts a border property from Props
func getBorderProp(props rtui.Props, key string) string {
	if props == nil {
		return ""
	}
	if v, ok := props[key].(string); ok {
		return v
	}
	return ""
}
