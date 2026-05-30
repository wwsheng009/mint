// Package render provides rendering pipeline components
package render

import (
	"strconv"

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
	fiberMap map[string]*reconciler.Fiber
}

// NewFiberToPaintableConverter creates a new converter.
func NewFiberToPaintableConverter(rootFiber *reconciler.Fiber) *FiberToPaintableConverter {
	c := &FiberToPaintableConverter{
		fiberMap: make(map[string]*reconciler.Fiber),
	}

	if rootFiber != nil {
		c.buildFiberMap(rootFiber)
	}

	return c
}

// buildFiberMap recursively indexes all Fibers by multiple keys
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
	// Index by NodeID string
	c.fiberMap[strconv.FormatUint(fiber.NodeID, 10)] = fiber

	// Recursively index children
	for child := fiber.Child; child != nil; child = child.Sibling {
		c.buildFiberMap(child)
	}
}

// Convert converts a LayoutBox tree to a PaintableBox tree.
func (c *FiberToPaintableConverter) Convert(
	lbox *layout.LayoutBox,
	parent *paint.PaintableBox,
) *paint.PaintableBox {
	if lbox == nil {
		return nil
	}

	pbox := &paint.PaintableBox{
		X:        lbox.X,
		Y:        lbox.Y,
		Width:    lbox.Width,
		Height:   lbox.Height,
		Layer:    convertLayoutLayerToInt(lbox.Layer),
		ZIndex:   lbox.ZIndex,
		Parent:   parent,
		Children: make([]*paint.PaintableBox, 0, len(lbox.Children)),
	}

	// Find matching Fiber and fill paint-specific data
	if fiber := c.findFiber(lbox.ID); fiber != nil {
		c.fillFromFiber(pbox, fiber)
	}

	// Copy BoxModel information (Padding, Margin) for debugging
	// This data comes from layout calculation
	if !lbox.BoxModel.IsEmpty() {
		pbox.PaddingTop = lbox.BoxModel.Padding.Top
		pbox.PaddingRight = lbox.BoxModel.Padding.Right
		pbox.PaddingBottom = lbox.BoxModel.Padding.Bottom
		pbox.PaddingLeft = lbox.BoxModel.Padding.Left

		pbox.MarginTop = lbox.BoxModel.Margin.Top
		pbox.MarginRight = lbox.BoxModel.Margin.Right
		pbox.MarginBottom = lbox.BoxModel.Margin.Bottom
		pbox.MarginLeft = lbox.BoxModel.Margin.Left
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

// findFiber finds a Fiber by ID
func (c *FiberToPaintableConverter) findFiber(id string) *reconciler.Fiber {
	if id == "" {
		return nil
	}

	// Strategy 1: Direct match by DiffKey
	if f, ok := c.fiberMap[id]; ok {
		return f
	}

	return nil
}

// fillFromFiber populates PaintableBox with data from Fiber
func (c *FiberToPaintableConverter) fillFromFiber(pbox *paint.PaintableBox, fiber *reconciler.Fiber) {
	pbox.NodeID = fiber.NodeID
	pbox.DiffKey = fiber.DiffKey
	pbox.LayoutDirty = fiber.IsLayoutDirty() ||
		fiber.IsPaintDirty() ||
		(fiber.Flags&(reconciler.EffectPlacement|reconciler.EffectUpdate|reconciler.EffectDeletion) != 0)

	// PaintableNode interface (wraps Fiber)
	pbox.Node = NewFiberPaintableNode(fiber)

	// Border info from Props
	// Note: Panel components don't render their own border - the internal VStack does.
	// This prevents double border rendering.
	if fiber.Props != nil && fiber.Tag != "panel" {
		pbox.BorderStyle = getBorderStyleFromProps(fiber.Props)
		pbox.BorderColor = getBorderProp(fiber.Props, "borderColor")
		pbox.BorderLabel = getBorderProp(fiber.Props, "borderLabel")
	}

	// Note: Padding and Margin from LayoutBox.BoxModel are NOT copied here.
	// They should be set during layout calculation if needed for debugging.
	// LayoutBox.BoxModel contains the complete box model information (margin, padding, border).
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
	if n.fiber.NodeID != 0 {
		return "fiber-node-" + strconv.FormatUint(n.fiber.NodeID, 10)
	}
	if n.fiber.DiffKey != "" {
		return n.fiber.DiffKey
	}

	return n.fiber.Key
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

// SetStyle sets the Fiber's style
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
	if content, ok := n.fiber.MemoizedState.(string); ok {
		return content
	}
	if n.fiber.Props != nil {
		if content, ok := n.fiber.Props["content"].(string); ok {
			return content
		}
	}
	return ""
}

// SetBounds sets the layout bounds for the Fiber's Instance.
// This is called by PaintEngine before Paint() to provide layout dimensions.
// Fiber-first Architecture: Passes layout-computed dimensions to Instance.
func (n *FiberPaintableNode) SetBounds(x, y, w, h int) {
	if n.fiber == nil || n.fiber.Instance == nil {
		return
	}

	// Pass bounds to Instance if it supports SetBounds
	if boundsSetter, ok := n.fiber.Instance.(interface{ SetBounds(int, int, int, int) }); ok {
		boundsSetter.SetBounds(x, y, w, h)
	}
}

func (n *FiberPaintableNode) SetViewportSize(width, height int) {
	if n.fiber == nil || n.fiber.Instance == nil {
		return
	}
	if viewportSetter, ok := n.fiber.Instance.(interface{ SetViewportSize(int, int) }); ok {
		viewportSetter.SetViewportSize(width, height)
	}
}

// Paint delegates to the Fiber's Instance.Paint() method.
// Fiber-first Architecture: Uses Instance ONLY, no VNode access.
func (n *FiberPaintableNode) Paint(x, y int) []paint.DrawCmd {
	if n.fiber == nil {
		return nil
	}

	// Primary Path: Use Fiber.Instance (Fiber-first)
	// The Instance persists across renders and holds all state.
	if n.fiber.Instance != nil {
		if inst, ok := n.fiber.Instance.(rtui.PaintableInstance); ok {
			return inst.Paint(x, y)
		}
	}

	// Fallback Path 1: Use PaintRegistry (simple stateless components)
	if fn := rtui.GetPaint(n.fiber.Tag); fn != nil {
		return fn(n.fiber.Props, n.fiber.Style, x, y)
	}

	// ⚠️ DEPRECATED: Fiber.FocusableVNode fallback removed.
	// Use Instance-based PaintableInstance interface instead.
	return nil
}

// =============================================================================
// Helper Functions
// =============================================================================

// convertLayoutLayerToInt 将 layout.Layer 转换为 int。
// 由于 layout.Layer 现在是 runtime.Layer 的别名，直接返回 int 值即可。
func convertLayoutLayerToInt(l layout.Layer) int {
	return int(l)
}

func getBorderStyleFromProps(props rtui.Props) paint.BorderStyle {
	if props == nil {
		return paint.BorderStyleNone
	}

	// Handle string style
	if s, ok := props["borderStyle"].(string); ok {
		switch s {
		case "single":
			return paint.BorderStyleSingle
		case "double":
			return paint.BorderStyleDouble
		case "rounded":
			return paint.BorderStyleRounded
		case "dashed":
			return paint.BorderStyleDashed
		}
		return paint.BorderStyleNone
	}

	// Handle layout.BorderStyle (unified type used by all packages)
	if bs, ok := props["borderStyle"].(layout.BorderStyle); ok {
		return paint.BorderStyle(bs)
	}

	return paint.BorderStyleNone
}

func getBorderProp(props rtui.Props, key string) string {
	if props == nil {
		return ""
	}
	if v, ok := props[key].(string); ok {
		return v
	}
	return ""
}
