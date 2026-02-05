// Package compute provides constraint-driven layout engine for TUI components
package compute

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/runtime"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Engine performs constraint-driven layout calculation
// It separates layout (position calculation) from paint (rendering)
type Engine struct {
	cache        *LayoutCache
	dirtyTracker *DirtyTracker
	debug        bool
}

// NewEngine creates a new layout engine
func NewEngine() *Engine {
	return &Engine{
		cache:        NewLayoutCache(),
		dirtyTracker: NewDirtyTracker(),
		debug:        os.Getenv("TUI_LAYOUT_DEBUG") == "true",
	}
}

// SetDebug enables/disables debug output
func (e *Engine) SetDebug(debug bool) {
	e.debug = debug
}

// Layout performs layout calculation on a VNode tree
// Returns a ComputedLayout containing computed positions for all nodes
func (e *Engine) Layout(vnode VNode, constraints runtime.BoxConstraints) (*ComputedLayout, error) {
	if vnode == nil {
		return nil, fmt.Errorf("cannot layout nil VNode")
	}

	// Build layout tree and measure
	root := e.buildComputedBox(vnode, nil, constraints)
	if root == nil {
		return NewComputedLayout(nil), nil
	}

	// Calculate positions (second pass)
	e.calculatePositions(root, 0, 0)

	// Clear dirty flags after layout
	root.ClearDirty()

	return NewComputedLayout(root), nil
}

// =============================================================================
// Layout Box Building (First Pass: Measurement)
// =============================================================================

// buildComputedBox creates a computed box for a VNode with its children
// This is the first pass - measuring sizes
//
// Caching strategy: Only cache leaf nodes (nodes without vnode children).
// This avoids the complexity of caching entire subtrees while still
// providing performance benefits for simple nodes like text.
func (e *Engine) buildComputedBox(vnode VNode, parent *ComputedBox, constraints runtime.BoxConstraints) *ComputedBox {
	if vnode == nil {
		return nil
	}

	box := &ComputedBox{
		VNode:  vnode,
		Parent: parent,
		Box:    runtime.Box{X: 0, Y: 0, Width: 0, Height: 0},
	}

	// Get vnode children to determine if this is a leaf node
	vnodeChildren := vnode.Children()
	isLeaf := len(vnodeChildren) == 0

	// Check cache (only for leaf nodes or nodes with explicit keys)
	cacheKey := e.getCacheKey(vnode, constraints)
	if isLeaf || vnode.Key() != "" {
		if cached, ok := e.cache.Get(cacheKey); ok && !e.dirtyTracker.NeedLayoutBox(box) {
			// Cache hit - use cached size
			box.Box = cached.Box
			if e.debug {
				fmt.Fprintf(os.Stderr, "[Layout.CacheHit] %s (key=%s): %v\n",
					vnode.Type().String(), vnode.Key(), cached.Box)
			}
			// For leaf nodes, we're done
			if isLeaf {
				box.Children = nil // Leaf nodes have no children
				return box
			}
			// For keyed nodes with children, we still need to build children
			// but can use the cached size for the parent
		}
	}

	// Measure the vnode
	size := e.measureVNode(vnode, constraints)
	box.Box.Width = size.Width
	box.Box.Height = size.Height

	// Build children layout boxes
	box.Children = make([]*ComputedBox, 0, len(vnodeChildren))

	for _, child := range vnodeChildren {
		// Calculate child constraints based on layout type
		childConstraints := e.getChildConstraints(vnode, child, constraints, size)

		childBox := e.buildComputedBox(child, box, childConstraints)
		if childBox != nil {
			box.Children = append(box.Children, childBox)
		}
	}

	// Cache the result for leaf nodes
	// Container nodes are not cached since their children may change
	if isLeaf {
		e.cache.Set(cacheKey, LayoutCacheEntry{
			Box:     box.Box,
			Size:    size,
			Hash:    e.computeHash(vnode, constraints),
			IsLeaf:  true,
			VNodeID: vnode.Key(),
		})
		if e.debug {
			fmt.Fprintf(os.Stderr, "[Layout.CacheSet] %s: %v\n",
				vnode.Type().String(), box.Box)
		}
	}

	return box
}

// measureVNode measures a VNode's size using constraints
func (e *Engine) measureVNode(vnode VNode, constraints runtime.BoxConstraints) runtime.Size {
	// PRIORITY 1: Use Measurable interface (constraint-based measurement)
	if measurable, ok := vnode.(interface {
		Measure(runtime.BoxConstraints) runtime.Size
	}); ok {
		size := measurable.Measure(constraints)
		if e.debug {
			fmt.Fprintf(os.Stderr, "[Layout.Measure] %s: constraints=%v, size=%v\n",
				vnode.Type().String(), constraints, size)
		}
		return size
	}

	// PRIORITY 2: Check for known layout types
	if tagger, ok := vnode.(interface{ Tag() string }); ok {
		switch tagger.Tag() {
		case "hstack", "vstack":
			// LayoutNode types should implement Measurable
			// Fallback to measuring children
			return e.measureLayoutChildren(vnode, constraints)
		case "bordered":
			// BorderedNode should implement Measurable
			// Fallback to content size + 2
			return e.measureBordered(vnode, constraints)
		case "text":
			// Text node
			if text := rtui.GetTextContent(vnode); text != "" {
				runes := []rune(text)
				return runtime.Size{Width: len(runes), Height: 1}
			}
			return runtime.Size{Width: 0, Height: 1}
		case "table":
			return e.measureTable(vnode, constraints)
		}
	}

	// PRIORITY 3: Fallback to estimation
	return e.estimateSize(vnode, constraints)
}

// measureLayoutChildren measures layout containers (HStack/VStack)
func (e *Engine) measureLayoutChildren(vnode VNode, constraints runtime.BoxConstraints) runtime.Size {
	children := vnode.Children()
	if len(children) == 0 {
		return runtime.Size{Width: constraints.MinWidth, Height: constraints.MinHeight}
	}

	// Get layout info
	layoutInfo := rtui.GetLayoutInfo(vnode)
	gap := layoutInfo.Gap

	if layoutInfo.IsHorizontal {
		// HStack: sum widths, max height
		totalWidth := 0
		maxHeight := 0
		for i, child := range children {
			childSize := e.measureVNode(child, runtime.BoxConstraints{
				MinWidth:  0,
				MaxWidth:  runtime.Infinity,
				MinHeight: 0,
				MaxHeight: constraints.MaxHeight,
			})
			totalWidth += childSize.Width
			if childSize.Height > maxHeight {
				maxHeight = childSize.Height
			}
			// Add gap (except after last child)
			if i < len(children)-1 {
				totalWidth += gap
			}
		}
		return runtime.Size{Width: totalWidth, Height: maxHeight}
	} else {
		// VStack: max width, sum heights
		maxWidth := 0
		totalHeight := 0
		for i, child := range children {
			childSize := e.measureVNode(child, runtime.BoxConstraints{
				MinWidth:  0,
				MaxWidth:  constraints.MaxWidth,
				MinHeight: 0,
				MaxHeight: runtime.Infinity,
			})
			if childSize.Width > maxWidth {
				maxWidth = childSize.Width
			}
			totalHeight += childSize.Height
			// Add gap (except after last child)
			if i < len(children)-1 {
				totalHeight += gap
			}
		}
		return runtime.Size{Width: maxWidth, Height: totalHeight}
	}
}

// measureBordered measures a bordered container
func (e *Engine) measureBordered(vnode VNode, constraints runtime.BoxConstraints) runtime.Size {
	children := vnode.Children()
	if len(children) == 0 {
		// Empty border - minimum size
		return runtime.Size{Width: 2, Height: 2}
	}

	child := children[0]

	// Measure child with inner constraints (subtract border)
	innerConstraints := runtime.BoxConstraints{
		MinWidth:  max(0, constraints.MinWidth-2),
		MaxWidth:  max(0, constraints.MaxWidth-2),
		MinHeight: max(0, constraints.MinHeight-2),
		MaxHeight: max(0, constraints.MaxHeight-2),
	}
	childSize := e.measureVNode(child, innerConstraints)

	// Check for label
	if labeled, ok := vnode.(interface{ GetBorderLabel() string }); ok {
		label := labeled.GetBorderLabel()
		if label != "" {
			labelWidth := len(label) + 2 // +2 for spaces
			if labelWidth > childSize.Width {
				childSize.Width = labelWidth
			}
		}
	}

	// Border adds 2 to width and height
	return runtime.Size{
		Width:  childSize.Width + 2,
		Height: childSize.Height + 2,
	}
}

// measureTable measures a table
func (e *Engine) measureTable(vnode VNode, constraints runtime.BoxConstraints) runtime.Size {
	rows := vnode.Children()
	if len(rows) == 0 {
		return runtime.Size{Width: 0, Height: 0}
	}

	// Calculate column widths
	colWidths := e.calculateColumnWidths(rows)

	// Sum column widths
	totalWidth := 0
	for _, w := range colWidths {
		totalWidth += w
	}

	// Sum row heights
	totalHeight := 0
	for _, row := range rows {
		rowHeight := e.measureVNode(row, constraints).Height
		totalHeight += rowHeight
	}

	return runtime.Size{Width: totalWidth, Height: totalHeight}
}

// calculateColumnWidths calculates the width of each column in a table
func (e *Engine) calculateColumnWidths(rows []VNode) []int {
	// Find max columns
	maxCols := 0
	for _, row := range rows {
		if tagger, ok := row.(interface{ Tag() string }); ok && tagger.Tag() == "tr" {
			cells := row.Children()
			if len(cells) > maxCols {
				maxCols = len(cells)
			}
		}
	}

	colWidths := make([]int, maxCols)

	// Find max width for each column
	for _, row := range rows {
		if tagger, ok := row.(interface{ Tag() string }); ok && tagger.Tag() == "tr" {
			cells := row.Children()
			for colIdx, cell := range cells {
				if colIdx < maxCols {
					cellWidth := e.measureVNode(cell, runtime.UnboundedConstraints()).Width
					if cellWidth > colWidths[colIdx] {
						colWidths[colIdx] = cellWidth
					}
				}
			}
		}
	}

	return colWidths
}

// estimateSize estimates size for unknown VNode types
func (e *Engine) estimateSize(vnode VNode, constraints runtime.BoxConstraints) runtime.Size {
	// Check for explicit width/height props
	if props := vnode.Props(); props != nil {
		if w := props.GetInt("width"); w > 0 {
			if h := props.GetInt("height"); h > 0 {
				return runtime.Size{Width: w, Height: h}
			}
			return runtime.Size{Width: w, Height: 1}
		}
	}

	// Check for text content
	if text := rtui.GetTextContent(vnode); text != "" {
		runes := []rune(text)
		return runtime.Size{Width: len(runes), Height: 1}
	}

	// Default minimum size
	return runtime.Size{Width: 10, Height: 1}
}

// getChildConstraints calculates constraints for a child VNode
func (e *Engine) getChildConstraints(parent, child VNode, parentConstraints runtime.BoxConstraints, parentSize runtime.Size) runtime.BoxConstraints {
	// Check if parent is a layout container
	if tagger, ok := parent.(interface{ Tag() string }); ok {
		switch tagger.Tag() {
		case "hstack":
			// HStack: children get unlimited width, constrained height
			return runtime.BoxConstraints{
				MinWidth:  0,
				MaxWidth:  runtime.Infinity,
				MinHeight: 0,
				MaxHeight: parentSize.Height,
			}
		case "vstack":
			// VStack: children get constrained width, unlimited height
			return runtime.BoxConstraints{
				MinWidth:  0,
				MaxWidth:  parentSize.Width,
				MinHeight: 0,
				MaxHeight: runtime.Infinity,
			}
		case "bordered":
			// Bordered: subtract border from constraints
			return runtime.BoxConstraints{
				MinWidth:  max(0, parentConstraints.MinWidth-2),
				MaxWidth:  max(0, parentConstraints.MaxWidth-2),
				MinHeight: max(0, parentConstraints.MinHeight-2),
				MaxHeight: max(0, parentConstraints.MaxHeight-2),
			}
		}
	}

	// Default: pass through parent constraints
	return parentConstraints
}

// =============================================================================
// Position Calculation (Second Pass: Layout)
// =============================================================================

// calculatePositions computes the (x, y) position for each computed box
// This is the second pass - after all sizes are known
func (e *Engine) calculatePositions(box *ComputedBox, x, y int) {
	box.Box.X = x
	box.Box.Y = y

	if e.debug {
		fmt.Fprintf(os.Stderr, "[Layout.Position] %s at %s\n",
			box.VNode.Type(), box.Box.String())
	}

	// Layout children based on parent type
	if box.VNode != nil {
		if tagger, ok := box.VNode.(interface{ Tag() string }); ok {
			switch tagger.Tag() {
			case "hstack":
				e.layoutHStack(box, x, y)
				return
			case "vstack":
				e.layoutVStack(box, x, y)
				return
			case "bordered":
				e.layoutBordered(box, x, y)
				return
			case "table":
				e.layoutTable(box, x, y)
				return
			}
		}
	}

	// Default: stack children vertically
	e.layoutDefault(box, x, y)
}

// layoutHStack positions children horizontally
func (e *Engine) layoutHStack(box *ComputedBox, x, y int) {
	layoutInfo := rtui.GetLayoutInfo(box.VNode)
	gap := layoutInfo.Gap

	childX := x
	for _, child := range box.Children {
		e.calculatePositions(child, childX, y)
		childX += child.Box.Width + gap
	}
}

// layoutVStack positions children vertically
func (e *Engine) layoutVStack(box *ComputedBox, x, y int) {
	layoutInfo := rtui.GetLayoutInfo(box.VNode)
	gap := layoutInfo.Gap

	childY := y
	for _, child := range box.Children {
		e.calculatePositions(child, x, childY)
		childY += child.Box.Height + gap
	}
}

// layoutBordered positions content inside border
func (e *Engine) layoutBordered(box *ComputedBox, x, y int) {
	// Content starts at (x+1, y+1) - inside the border
	for _, child := range box.Children {
		e.calculatePositions(child, x+1, y+1)
	}
}

// layoutTable positions table rows and cells
func (e *Engine) layoutTable(box *ComputedBox, x, y int) {
	// Calculate column widths
	colWidths := e.calculateColumnWidthsFromBoxes(box)

	rowY := y
	for _, child := range box.Children {
		if child.VNode != nil {
			if tagger, ok := child.VNode.(interface{ Tag() string }); ok && tagger.Tag() == "tr" {
				e.layoutTableRow(child, x, rowY, colWidths)
				rowY += child.Box.Height
			} else {
				e.calculatePositions(child, x, rowY)
				rowY += child.Box.Height
			}
		}
	}
}

// layoutTableRow positions cells in a table row
func (e *Engine) layoutTableRow(box *ComputedBox, x, y int, colWidths []int) {
	cellX := x
	for colIdx, child := range box.Children {
		e.calculatePositions(child, cellX, y)
		if colIdx < len(colWidths) {
			cellX += colWidths[colIdx]
		} else {
			cellX += child.Box.Width
		}
	}
}

// layoutDefault positions children vertically (default layout)
func (e *Engine) layoutDefault(box *ComputedBox, x, y int) {
	childY := y
	for _, child := range box.Children {
		e.calculatePositions(child, x, childY)
		childY += child.Box.Height
	}
}

// calculateColumnWidthsFromBoxes calculates column widths from computed boxes
func (e *Engine) calculateColumnWidthsFromBoxes(box *ComputedBox) []int {
	// Find max columns
	maxCols := 0
	for _, child := range box.Children {
		if child.VNode != nil {
			if tagger, ok := child.VNode.(interface{ Tag() string }); ok && tagger.Tag() == "tr" {
				if len(child.Children) > maxCols {
					maxCols = len(child.Children)
				}
			}
		}
	}

	colWidths := make([]int, maxCols)
	for _, rowBox := range box.Children {
		for colIdx, cellBox := range rowBox.Children {
			if colIdx < maxCols && cellBox.Box.Width > colWidths[colIdx] {
				colWidths[colIdx] = cellBox.Box.Width
			}
		}
	}

	return colWidths
}

// =============================================================================
// Cache Keys and Helpers
// =============================================================================

// getCacheKey generates a cache key for a VNode
func (e *Engine) getCacheKey(vnode VNode, constraints runtime.BoxConstraints) LayoutCacheKey {
	key := LayoutCacheKey{
		VNodeType:   vnode.Type().String(),
		Constraints: constraints,
	}

	if keyNode := vnode.Key(); keyNode != "" {
		key.VNodeKey = keyNode
	}

	// Include relevant props in the key
	if props := vnode.Props(); props != nil {
		if w := props.GetInt("width"); w > 0 {
			key.PropsHash = key.PropsHash*31 + uint64(w)
		}
		if h := props.GetInt("height"); h > 0 {
			key.PropsHash = key.PropsHash*31 + uint64(h)
		}
	}

	return key
}

// computeHash computes a hash for VNode + constraints
func (e *Engine) computeHash(vnode VNode, constraints runtime.BoxConstraints) uint64 {
	h := uint64(5381)

	// Hash type
	h = h*31 + uint64(vnode.Type())

	// Hash key
	if key := vnode.Key(); key != "" {
		h = h*31 + uint64(hashString(key))
	}

	// Hash constraints
	h = h*31 + uint64(constraints.MinWidth)
	h = h*31 + uint64(constraints.MaxWidth)
	h = h*31 + uint64(constraints.MinHeight)
	h = h*31 + uint64(constraints.MaxHeight)

	return h
}

func hashString(s string) uint32 {
	h := uint32(5381)
	for _, c := range s {
		h = h*32 + uint32(c)
	}
	return h
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// =============================================================================
// Cache Management
// =============================================================================

// GetCacheStats returns statistics about the layout cache
func (e *Engine) GetCacheStats() CacheStats {
	return e.cache.Stats()
}

// ResetCacheStats resets cache hit/miss counters without clearing the cache
func (e *Engine) ResetCacheStats() {
	e.cache.ResetStats()
}

// ClearCache clears all cached layout results
func (e *Engine) ClearCache() {
	e.cache.Clear()
}

// InvalidateCacheByType removes all cached entries for a specific VNode type
func (e *Engine) InvalidateCacheByType(vNodeType string) {
	e.cache.InvalidateByType(vNodeType)
}

// InvalidateCacheByKey removes all cached entries for a specific VNode key
func (e *Engine) InvalidateCacheByKey(vnodeKey string) {
	e.cache.InvalidateByKey(vnodeKey)
}

// CacheSize returns the number of entries in the cache
func (e *Engine) CacheSize() int {
	return e.cache.Size()
}
