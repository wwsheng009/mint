package paint

import (
	"fmt"
	"strings"
)

// =============================================================================
// PaintableBox - Decoupled Layout Box for Paint Engine
// =============================================================================
// PaintableBox replaces compute.ComputedBox as the input to PaintEngine.
// It is a pure data structure with no VNode/Fiber dependencies.

// PaintableBox represents a computed layout box that can be painted.
// This is the primary input type for the PaintEngine after refactoring.
//
// Design Principle:
// - PaintableBox is a PURE PAINT data structure
// - It contains ALL data needed for painting, copied from Fiber/LayoutBox
// - No dependencies on layout or Fiber packages
type PaintableBox struct {
	// === PaintableNode Interface ===
	// Node provides the Paint() method and style information
	Node PaintableNode

	// === Position and Size (from LayoutBox) ===
	X, Y      int
	Width     int
	Height    int

	// === Identity (from Fiber) ===
	// NodeID for stable runtime identity (e.g., 95)
	NodeID uint64
	// DiffKey for dirty tracking (e.g., "item-1")
	DiffKey string

	// === Rendering Layer (from LayoutBox) ===
	// Layer: 0=Base, 1=Overlay, 2=Modal, 3=Tooltip, 4=Inspector
	Layer  int
	ZIndex int

	// === Tree Structure ===
	Children []*PaintableBox
	Parent   *PaintableBox

	// === Rendering Helpers ===
	// Text with alignment padding applied (computed during layout)
	RenderedText string
	// Natural width before stretching (for alignment)
	NaturalWidth int

	// === Border Information ===
	BorderStyle BorderStyle
	BorderColor string
	BorderLabel string

	// === Padding Information ===
	PaddingTop    int
	PaddingRight  int
	PaddingBottom int
	PaddingLeft   int

	// === Margin Information ===
	MarginTop    int
	MarginRight  int
	MarginBottom int
	MarginLeft   int

	// === State ===
	LayoutDirty bool
	LayoutHash  uint64
}

// NewPaintableBox creates a new PaintableBox with the given node.
func NewPaintableBox(node PaintableNode) *PaintableBox {
	return &PaintableBox{
		Node:    node,
		Layer:   0,
		ZIndex:  0,
		Children: make([]*PaintableBox, 0),
	}
}

// NewPaintableBoxWithBounds creates a new PaintableBox with specified bounds.
func NewPaintableBoxWithBounds(node PaintableNode, x, y, w, h int) *PaintableBox {
	return &PaintableBox{
		Node:     node,
		X:        x,
		Y:        y,
		Width:    w,
		Height:   h,
		Layer:    0,
		ZIndex:   0,
		Children: make([]*PaintableBox, 0),
	}
}

// Bounds returns the box bounds as separate values.
func (b *PaintableBox) Bounds() (x, y, w, h int) {
	return b.X, b.Y, b.Width, b.Height
}

// SetBounds sets the box bounds.
func (b *PaintableBox) SetBounds(x, y, w, h int) {
	b.X, b.Y, b.Width, b.Height = x, y, w, h
}

// Rect returns the box as a Rect.
func (b *PaintableBox) Rect() Rect {
	return Rect{X: b.X, Y: b.Y, Width: b.Width, Height: b.Height}
}

// AddChild adds a child box and sets its parent reference.
func (b *PaintableBox) AddChild(child *PaintableBox) {
	if child == nil {
		return
	}
	child.Parent = b
	b.Children = append(b.Children, child)
}

// MarkDirty marks this box and all ancestors as needing layout.
func (b *PaintableBox) MarkDirty() {
	b.LayoutDirty = true
	for parent := b.Parent; parent != nil; parent = parent.Parent {
		if parent.LayoutDirty {
			break // Already marked, stop traversing
		}
		parent.LayoutDirty = true
	}
}

// ClearDirty clears the dirty flag for this box and all descendants.
func (b *PaintableBox) ClearDirty() {
	b.LayoutDirty = false
	for _, child := range b.Children {
		child.ClearDirty()
	}
}

// Depth returns the depth of this box in the tree (root = 0).
func (b *PaintableBox) Depth() int {
	depth := 0
	for parent := b.Parent; parent != nil; parent = parent.Parent {
		depth++
	}
	return depth
}

// Count returns the total number of boxes in this subtree.
func (b *PaintableBox) Count() int {
	count := 1
	for _, child := range b.Children {
		count += child.Count()
	}
	return count
}

// FindByPosition finds the innermost box containing the given position.
func (b *PaintableBox) FindByPosition(x, y int) *PaintableBox {
	// Check if point is in this box
	if x < b.X || x >= b.X+b.Width || y < b.Y || y >= b.Y+b.Height {
		return nil
	}

	// Check children in reverse order (topmost first)
	for i := len(b.Children) - 1; i >= 0; i-- {
		if found := b.Children[i].FindByPosition(x, y); found != nil {
			return found
		}
	}

	return b
}

// FindByID finds a box by its NodeID.
func (b *PaintableBox) FindByID(nodeID uint64) *PaintableBox {
	if b.NodeID == nodeID {
		return b
	}
	for _, child := range b.Children {
		if found := child.FindByID(nodeID); found != nil {
			return found
		}
	}
	return nil
}

// FindByNodeID finds a box by Node ID string (for compatibility).
func (b *PaintableBox) FindByNodeID(id string) *PaintableBox {
	if b.Node != nil && b.Node.ID() == id {
		return b
	}
	for _, child := range b.Children {
		if found := child.FindByNodeID(id); found != nil {
			return found
		}
	}
	return nil
}

// HasBorder returns true if this box has a border.
func (b *PaintableBox) HasBorder() bool {
	return b.BorderStyle != BorderStyleNone
}

// GetBorderInfo returns border information, preferring direct properties over Node interface.
func (b *PaintableBox) GetBorderInfo() (style BorderStyle, color, label string) {
	// Prefer direct properties
	if b.BorderStyle != BorderStyleNone {
		return b.BorderStyle, b.BorderColor, b.BorderLabel
	}
	// Try to get from Node interface
	if b.Node != nil {
		if bi, ok := b.Node.(BorderInfo); ok {
			return bi.GetBorderStyle(), bi.GetBorderColor(), bi.GetBorderLabel()
		}
	}
	return BorderStyleNone, "", ""
}

// Contains checks if a point is within this box.
func (b *PaintableBox) Contains(x, y int) bool {
	return x >= b.X && x < b.X+b.Width && y >= b.Y && y < b.Y+b.Height
}

// ContainsRect checks if a rectangle is fully contained within this box.
func (b *PaintableBox) ContainsRect(r Rect) bool {
	return r.X >= b.X && r.Y >= b.Y &&
		r.X+r.Width <= b.X+b.Width && r.Y+r.Height <= b.Y+b.Height
}

// Intersects checks if this box intersects with another box.
func (b *PaintableBox) Intersects(other *PaintableBox) bool {
	if other == nil {
		return false
	}
	return b.X < other.X+other.Width &&
		b.X+b.Width > other.X &&
		b.Y < other.Y+other.Height &&
		b.Y+b.Height > other.Y
}

// Clone creates a shallow copy of this box (children are not cloned).
func (b *PaintableBox) Clone() *PaintableBox {
	return &PaintableBox{
		Node:         b.Node,
		X:            b.X,
		Y:            b.Y,
		Width:        b.Width,
		Height:       b.Height,
		NodeID:       b.NodeID,
		DiffKey:      b.DiffKey,
		Layer:        b.Layer,
		ZIndex:       b.ZIndex,
		RenderedText: b.RenderedText,
		NaturalWidth: b.NaturalWidth,
		BorderStyle:  b.BorderStyle,
		BorderColor:  b.BorderColor,
		BorderLabel:  b.BorderLabel,
		LayoutDirty:  b.LayoutDirty,
		LayoutHash:   b.LayoutHash,
		Children:     make([]*PaintableBox, 0),
	}
}

// CloneDeep creates a deep copy of this box and all its children.
func (b *PaintableBox) CloneDeep() *PaintableBox {
	clone := b.Clone()
	for _, child := range b.Children {
		childClone := child.CloneDeep()
		childClone.Parent = clone
		clone.Children = append(clone.Children, childClone)
	}
	return clone
}

// =============================================================================
// PaintableLayout - Root Layout Container
// =============================================================================

// PaintableLayout represents a complete layout tree for painting.
// This replaces compute.ComputedLayout as the input to PaintEngine.
type PaintableLayout struct {
	Root   *PaintableBox
	HitMap interface{} // Using interface{} to avoid importing event package; will be *event.HitMap
}

// NewPaintableLayout creates a new PaintableLayout with the given root box.
func NewPaintableLayout(root *PaintableBox) *PaintableLayout {
	return &PaintableLayout{Root: root}
}

// NewPaintableLayoutWithHitMap creates a new PaintableLayout with a HitMap.
func NewPaintableLayoutWithHitMap(root *PaintableBox, hitMap interface{}) *PaintableLayout {
	return &PaintableLayout{
		Root:   root,
		HitMap: hitMap,
	}
}

// FindByPosition finds the innermost box containing the given position.
func (l *PaintableLayout) FindByPosition(x, y int) *PaintableBox {
	if l.Root == nil {
		return nil
	}
	return l.Root.FindByPosition(x, y)
}

// FindByID finds a box by its NodeID.
func (l *PaintableLayout) FindByID(nodeID uint64) *PaintableBox {
	if l.Root == nil {
		return nil
	}
	return l.Root.FindByID(nodeID)
}

// Count returns the total number of boxes in the layout.
func (l *PaintableLayout) Count() int {
	if l.Root == nil {
		return 0
	}
	return l.Root.Count()
}

// IsEmpty returns true if the layout has no root.
func (l *PaintableLayout) IsEmpty() bool {
	return l.Root == nil
}

// Bounds returns the bounds of the root box.
func (l *PaintableLayout) Bounds() (x, y, w, h int) {
	if l.Root == nil {
		return 0, 0, 0, 0
	}
	return l.Root.Bounds()
}

// =============================================================================
// PaintableBox Builder (Fluent API)
// =============================================================================

// PaintableBoxBuilder provides a fluent API for constructing PaintableBox.
type PaintableBoxBuilder struct {
	box *PaintableBox
}

// NewPaintableBoxBuilder creates a new builder.
func NewPaintableBoxBuilder() *PaintableBoxBuilder {
	return &PaintableBoxBuilder{
		box: &PaintableBox{Children: make([]*PaintableBox, 0)},
	}
}

// WithNode sets the paintable node.
func (b *PaintableBoxBuilder) WithNode(node PaintableNode) *PaintableBoxBuilder {
	b.box.Node = node
	return b
}

// WithBounds sets the position and size.
func (b *PaintableBoxBuilder) WithBounds(x, y, w, h int) *PaintableBoxBuilder {
	b.box.X, b.box.Y, b.box.Width, b.box.Height = x, y, w, h
	return b
}

// WithLayer sets the layer.
func (b *PaintableBoxBuilder) WithLayer(layer int) *PaintableBoxBuilder {
	b.box.Layer = layer
	return b
}

// WithZIndex sets the Z-index.
func (b *PaintableBoxBuilder) WithZIndex(zIndex int) *PaintableBoxBuilder {
	b.box.ZIndex = zIndex
	return b
}

// WithNodeID sets the node ID.
func (b *PaintableBoxBuilder) WithNodeID(nodeID uint64) *PaintableBoxBuilder {
	b.box.NodeID = nodeID
	return b
}

// WithDiffKey sets the diff key for dirty tracking.
func (b *PaintableBoxBuilder) WithDiffKey(diffKey string) *PaintableBoxBuilder {
	b.box.DiffKey = diffKey
	return b
}

// WithBorder sets the border properties.
func (b *PaintableBoxBuilder) WithBorder(style BorderStyle, color, label string) *PaintableBoxBuilder {
	b.box.BorderStyle = style
	b.box.BorderColor = color
	b.box.BorderLabel = label
	return b
}

// WithRenderedText sets the rendered text.
func (b *PaintableBoxBuilder) WithRenderedText(text string) *PaintableBoxBuilder {
	b.box.RenderedText = text
	return b
}

// AddChild adds a child box.
func (b *PaintableBoxBuilder) AddChild(child *PaintableBox) *PaintableBoxBuilder {
	b.box.AddChild(child)
	return b
}

// Build returns the constructed PaintableBox.
func (b *PaintableBoxBuilder) Build() *PaintableBox {
	return b.box
}

// =============================================================================
// PaintableBox Tree Methods
// =============================================================================

// String returns the paintable tree as string representation.
func (b *PaintableBox) String() string {
	return b.TreeString()
}

// TreeString returns the paintable tree as string (hierarchical structure).
func (b *PaintableBox) TreeString() string {
	if b == nil {
		return "No paintable tree found!"
	}

	var sb strings.Builder
	sb.WriteString("Paintable Tree (hierarchical):\n")
	sb.WriteString(strings.Repeat("=", 70))
	sb.WriteString("\n")
	b.buildTreeNodeString(b, 0, &sb)
	return sb.String()
}

// buildTreeNodeString recursively builds the string representation of paintable tree nodes.
func (b *PaintableBox) buildTreeNodeString(box *PaintableBox, depth int, sb *strings.Builder) {
	if box == nil {
		return
	}

	indent := strings.Repeat("  ", depth)

	// Get node type
	nodeType := "unknown"
	tag := "-"
	if box.Node != nil {
		// Use Tag() for element names, fallback to NodeType for generic types
		nodeTag := box.Node.Tag()
		if nodeTag != "" {
			tag = nodeTag
		}
		nodeType = fmt.Sprintf("%s (type:%s)", tag, box.Node.NodeType())
	}

	// Build diff key info
	diffKey := box.DiffKey
	if len(diffKey) > 15 {
		diffKey = diffKey[:12] + "..."
	}
	if diffKey == "" {
		diffKey = "-"
	}

	// 构建边框详细信息
	borderInfo := ""
	if box.BorderStyle != BorderStyleNone {
		borderWidth := 1 // 边框宽度（单字符）
		borderInfo += fmt.Sprintf(", Border:%s", box.BorderStyle.String())
		if box.BorderLabel != "" {
			borderInfo += fmt.Sprintf("(%q)", box.BorderLabel)
		}
		// 边框绘制信息：边框占用2个字符的水平和垂直空间
		borderInfo += fmt.Sprintf(" [border_area:%d,%d,%dx%d]",
			box.X, box.Y, box.Width, box.Height)

		// 内容区域坐标（边框内 + padding内）
		borderX := box.X + borderWidth
		borderY := box.Y + borderWidth
		borderW := box.Width - borderWidth*2
		borderH := box.Height - borderWidth*2

		contentX := borderX + box.PaddingLeft
		contentY := borderY + box.PaddingTop
		contentW := borderW - box.PaddingLeft - box.PaddingRight
		contentH := borderH - box.PaddingTop - box.PaddingBottom

		borderInfo += fmt.Sprintf(" [padding_inner:%d,%d,%d,%d]",
			box.PaddingTop, box.PaddingRight, box.PaddingBottom, box.PaddingLeft)
		borderInfo += fmt.Sprintf(" [content:%d,%d,%dx%d]", contentX, contentY, contentW, contentH)

		// 标签额外占用空间
		if box.BorderLabel != "" {
			labelWidth := len(box.BorderLabel) + 2 // 标签+2个空格
			borderInfo += fmt.Sprintf(" [label_w:%d]", labelWidth)
		}
	}

	// 构建 Margin 信息
	marginInfo := ""
	if box.MarginTop != 0 || box.MarginRight != 0 || box.MarginBottom != 0 || box.MarginLeft != 0 {
		marginInfo = fmt.Sprintf(" [margin:%d,%d,%d,%d]",
			box.MarginTop, box.MarginRight, box.MarginBottom, box.MarginLeft)
	}

	// Append node with hierarchical relationship
	sb.WriteString(fmt.Sprintf("%s└─ [%s] ID:%d DiffKey:%s Size:%dx%d Pos:%d,%d Layer:%d%s%s\n",
		indent, nodeType, box.NodeID, diffKey,
		box.Width, box.Height, box.X, box.Y, box.Layer, borderInfo, marginInfo))

	// Append children recursively
	for _, child := range box.Children {
		b.buildTreeNodeString(child, depth+1, sb)
	}
}
