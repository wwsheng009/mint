// Package compute provides constraint-driven layout engine for TUI components
package compute

import (
	"fmt"
	"os"
	"strings"

	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Engine performs constraint-driven layout calculation
// It separates layout (position calculation) from paint (rendering)
type Engine struct {
	cache        *LayoutCache
	dirtyTracker *DirtyTracker
	debug        bool
	flexCache    map[string]*FlexDistributionInfo // Cache for flex distribution per parent
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

	// Reset flex cache for each layout pass
	e.flexCache = make(map[string]*FlexDistributionInfo)

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
	padding := layoutInfo.Padding // top, right, bottom, left
	paddingWidth := padding[1] + padding[3]
	paddingHeight := padding[0] + padding[2]

	if layoutInfo.IsHorizontal {
		// HStack: sum widths, max height
		totalWidth := 0
		maxHeight := 0

		// Calculate inner height constraint
		// Use parent's MaxHeight only if it's bounded, otherwise use Infinity
		innerMaxHeight := runtime.Infinity
		if constraints.HasBoundedHeight() {
			innerMaxHeight = max(0, constraints.MaxHeight-paddingHeight)
		}

		if e.debug {
			fmt.Fprintf(os.Stderr, "[measureLayoutChildren.HStack] constraints=%v, paddingWidth=%d, paddingHeight=%d\n",
				constraints, paddingWidth, paddingHeight)
		}

		// First pass: identify flex children and measure non-flex children
		var flexChildren []struct {
			child  VNode
			index  int
			factor int
		}
		var fixedWidth int
		flexTotalFactor := 0

		for i, child := range children {
			childInfo := rtui.GetLayoutInfo(child)
			if childInfo.Flex > 0 {
				flexChildren = append(flexChildren, struct {
					child  VNode
					index  int
					factor int
				}{child, i, childInfo.Flex})
				flexTotalFactor += childInfo.Flex
			} else {
				// Non-flex child: measure with natural width
				childConstraints := runtime.BoxConstraints{
					MinWidth:  0,
					MaxWidth:  runtime.Infinity,
					MinHeight: 0,
					MaxHeight: innerMaxHeight,
				}
				childSize := e.measureVNode(child, childConstraints)
				fixedWidth += childSize.Width
				if childSize.Height > maxHeight && childSize.Height < runtime.Infinity {
					maxHeight = childSize.Height
				}
			}
			// Account for gap (except after last child)
			if i < len(children)-1 {
				fixedWidth += gap
			}
		}

		totalWidth = fixedWidth

		// If we have flex children and bounded width, distribute remaining space
		if len(flexChildren) > 0 && constraints.HasBoundedWidth() {
			availableWidth := constraints.MaxWidth - paddingWidth - (len(children)-1)*gap
			remainingSpace := availableWidth - fixedWidth

			if e.debug && remainingSpace > 0 {
				fmt.Fprintf(os.Stderr, "[measureLayoutChildren.HStack] flex distribution: available=%d, fixed=%d, remaining=%d, factors=%d\n",
					availableWidth, fixedWidth, remainingSpace, flexTotalFactor)
			}

			// Distribute remaining space to flex children
			for _, fc := range flexChildren {
				flexWidth := (remainingSpace * fc.factor) / flexTotalFactor
				if flexWidth < 0 {
					flexWidth = 0
				}

				childConstraints := runtime.BoxConstraints{
					MinWidth:  flexWidth,
					MaxWidth:  flexWidth,
					MinHeight: 0,
					MaxHeight: innerMaxHeight,
				}
				childSize := e.measureVNode(fc.child, childConstraints)
				totalWidth += childSize.Width
				if childSize.Height > maxHeight && childSize.Height < runtime.Infinity {
					maxHeight = childSize.Height
				}
			}
		} else {
			// No flex or unbounded width: measure flex children naturally
			for _, fc := range flexChildren {
				childConstraints := runtime.BoxConstraints{
					MinWidth:  0,
					MaxWidth:  runtime.Infinity,
					MinHeight: 0,
					MaxHeight: innerMaxHeight,
				}
				childSize := e.measureVNode(fc.child, childConstraints)
				totalWidth += childSize.Width
				if childSize.Height > maxHeight && childSize.Height < runtime.Infinity {
					maxHeight = childSize.Height
				}
			}
		}

		// Add padding to total size
		totalWidth += paddingWidth
		maxHeight += paddingHeight

		// IMPORTANT: Cross-axis filling for HStack
		// Fill available height so children can stretch vertically
		if constraints.HasBoundedHeight() && maxHeight < constraints.MaxHeight {
			maxHeight = constraints.MaxHeight
		}

		// Apply MinHeight constraint
		if maxHeight < constraints.MinHeight {
			maxHeight = constraints.MinHeight
		}

		// Clamp to MaxHeight if exceeded
		if constraints.HasBoundedHeight() && maxHeight > constraints.MaxHeight {
			maxHeight = constraints.MaxHeight
		}

		if e.debug {
			fmt.Fprintf(os.Stderr, "[measureLayoutChildren.HStack] RETURN: Width=%d, Height=%d\n",
				totalWidth, maxHeight)
		}

		return runtime.Size{Width: totalWidth, Height: maxHeight}
	} else {
		// VStack: max width, sum heights
		maxWidth := 0
		totalHeight := 0

		// Calculate inner width constraint
		// Use parent's MaxWidth only if it's bounded, otherwise use Infinity
		innerMaxWidth := runtime.Infinity
		if constraints.HasBoundedWidth() {
			innerMaxWidth = max(0, constraints.MaxWidth-paddingWidth)
		}

		if e.debug {
			fmt.Fprintf(os.Stderr, "[measureLayoutChildren.VStack] constraints=%v, innerMaxWidth=%d\n",
				constraints, innerMaxWidth)
		}

		// First pass: identify flex children and measure non-flex children
		var flexChildren []struct {
			child  VNode
			index  int
			factor int
		}
		var fixedHeight int
		flexTotalFactor := 0

		for i, child := range children {
			childInfo := rtui.GetLayoutInfo(child)
			if childInfo.Flex > 0 {
				flexChildren = append(flexChildren, struct {
					child  VNode
					index  int
					factor int
				}{child, i, childInfo.Flex})
				flexTotalFactor += childInfo.Flex
			} else {
				// Non-flex child: measure with natural height
				childConstraints := runtime.BoxConstraints{
					MinWidth:  0,
					MaxWidth:  innerMaxWidth,
					MinHeight: 0,
					MaxHeight: runtime.Infinity,
				}
				childSize := e.measureVNode(child, childConstraints)
				if childSize.Width > maxWidth && childSize.Width < runtime.Infinity {
					maxWidth = childSize.Width
				}
				if childSize.Height < runtime.Infinity {
					fixedHeight += childSize.Height
				}
			}
			// Account for gap (except after last child)
			if i < len(children)-1 {
				fixedHeight += gap
			}
		}

		totalHeight = fixedHeight

		// If we have flex children and bounded height, distribute remaining space
		if len(flexChildren) > 0 && constraints.HasBoundedHeight() {
			availableHeight := constraints.MaxHeight - paddingHeight - (len(children)-1)*gap
			remainingSpace := availableHeight - fixedHeight

			if e.debug && remainingSpace > 0 {
				fmt.Fprintf(os.Stderr, "[measureLayoutChildren.VStack] flex distribution: available=%d, fixed=%d, remaining=%d, factors=%d\n",
					availableHeight, fixedHeight, remainingSpace, flexTotalFactor)
			}

			// Distribute remaining space to flex children
			for _, fc := range flexChildren {
				flexHeight := (remainingSpace * fc.factor) / flexTotalFactor
				if flexHeight < 0 {
					flexHeight = 0
				}

				childConstraints := runtime.BoxConstraints{
					MinWidth:  0,
					MaxWidth:  innerMaxWidth,
					MinHeight: flexHeight,
					MaxHeight: flexHeight,
				}
				childSize := e.measureVNode(fc.child, childConstraints)
				if childSize.Width > maxWidth && childSize.Width < runtime.Infinity {
					maxWidth = childSize.Width
				}
				totalHeight += childSize.Height
			}
		} else {
			// No flex or unbounded height: measure flex children naturally
			for _, fc := range flexChildren {
				childConstraints := runtime.BoxConstraints{
					MinWidth:  0,
					MaxWidth:  innerMaxWidth,
					MinHeight: 0,
					MaxHeight: runtime.Infinity,
				}
				childSize := e.measureVNode(fc.child, childConstraints)
				if childSize.Width > maxWidth && childSize.Width < runtime.Infinity {
					maxWidth = childSize.Width
				}
				if childSize.Height < runtime.Infinity {
					totalHeight += childSize.Height
				}
			}
		}

		// Add padding to total size
		maxWidth += paddingWidth
		totalHeight += paddingHeight

		// IMPORTANT: Cross-axis filling for VStack
		// Fill available width so children can stretch horizontally
		if constraints.HasBoundedWidth() && maxWidth < constraints.MaxWidth {
			maxWidth = constraints.MaxWidth
		}

		// Apply MinWidth constraint
		if maxWidth < constraints.MinWidth {
			maxWidth = constraints.MinWidth
		}

		// Clamp to MaxWidth if exceeded
		if constraints.HasBoundedWidth() && maxWidth > constraints.MaxWidth {
			maxWidth = constraints.MaxWidth
		}

		if e.debug {
			fmt.Fprintf(os.Stderr, "[measureLayoutChildren.VStack] RETURN: Width=%d, Height=%d\n",
				maxWidth, totalHeight)
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
	// Use SubtractPadding helper to properly handle bounded/unbounded constraints
	innerConstraints := constraints.SubtractPadding(2, 2)
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

// measureTextWidth returns the display width of a text string
func (e *Engine) measureTextWidth(text string) int {
	return paint.StringWidth(text)
}

// getFlexDistribution calculates (or retrieves from cache) the flex distribution
// for a parent HStack or VStack. This avoids O(N²) re-measurement.
// isHorizontal: true for HStack, false for VStack
// maxSize: MaxHeight for HStack, MaxWidth for VStack
func (e *Engine) getFlexDistribution(parent VNode, isHorizontal bool, maxSize int) *FlexDistributionInfo {
	// Generate cache key from parent's key or address
	cacheKey := ""
	if key := parent.Key(); key != "" {
		cacheKey = key
	} else {
		// Fallback: use a unique identifier based on parent's address (not ideal but works)
		cacheKey = fmt.Sprintf("%p", parent)
	}

	// Check cache
	if cached, ok := e.flexCache[cacheKey]; ok && cached.Valid {
		return cached
	}

	// Calculate flex distribution
	info := &FlexDistributionInfo{
		TotalFlexFactor: 0,
		FixedSize:       0,
		ChildCount:      0,
		Valid:           true,
	}

	children := parent.Children()
	info.ChildCount = len(children)

	for _, child := range children {
		childInfo := rtui.GetLayoutInfo(child)
		if childInfo.Flex > 0 {
			info.TotalFlexFactor += childInfo.Flex
		} else {
			// Measure non-flex child
			var childSize runtime.Size
			if isHorizontal {
				// HStack: measure with unlimited width
				childConstraints := runtime.BoxConstraints{
					MinWidth:  0,
					MaxWidth:  runtime.Infinity,
					MinHeight: 0,
					MaxHeight: maxSize,
				}
				childSize = e.measureVNode(child, childConstraints)
				info.FixedSize += childSize.Width
			} else {
				// VStack: measure with unlimited height
				childConstraints := runtime.BoxConstraints{
					MinWidth:  0,
					MaxWidth:  maxSize,
					MinHeight: 0,
					MaxHeight: runtime.Infinity,
				}
				childSize = e.measureVNode(child, childConstraints)
				if childSize.Height < runtime.Infinity {
					info.FixedSize += childSize.Height
				}
			}
		}
	}

	// Store in cache
	e.flexCache[cacheKey] = info

	return info
}

// getChildConstraints calculates constraints for a child VNode
// IMPORTANT: Constraints should be passed based on parent's actual bounds, not Infinity
// Padding must be subtracted from constraints before passing to children
// Flex distribution is handled here for layout containers
func (e *Engine) getChildConstraints(parent, child VNode, parentConstraints runtime.BoxConstraints, parentSize runtime.Size) runtime.BoxConstraints {
	// Get parent layout info to check for padding
	layoutInfo := rtui.GetLayoutInfo(parent)
	padding := layoutInfo.Padding // top, right, bottom, left
	paddingWidth := padding[1] + padding[3]
	paddingHeight := padding[0] + padding[2]

	// Check if parent is a layout container
	if tagger, ok := parent.(interface{ Tag() string }); ok {
		switch tagger.Tag() {
		case "hstack":
			// HStack: calculate flex distribution if applicable
			childMaxHeight := runtime.Infinity
			if parentConstraints.HasBoundedHeight() {
				childMaxHeight = max(0, parentConstraints.MaxHeight-paddingHeight)
			}

			// Check if child has flex
			childInfo := rtui.GetLayoutInfo(child)

			// If child has flex and parent has bounded width, calculate flex width
			if childInfo.Flex > 0 && parentConstraints.HasBoundedWidth() {
				// Use cached flex distribution to avoid O(N²) re-measurement
				flexDist := e.getFlexDistribution(parent, true, childMaxHeight)

				// Calculate available space and this child's flex width
				gapSpace := 0
				if flexDist.ChildCount > 1 {
					gapSpace = (flexDist.ChildCount - 1) * layoutInfo.Gap
				}
				availableWidth := parentConstraints.MaxWidth - paddingWidth - gapSpace
				remainingSpace := availableWidth - flexDist.FixedSize
				flexWidth := (remainingSpace * childInfo.Flex) / flexDist.TotalFlexFactor
				if flexWidth < 0 {
					flexWidth = 0
				}

				if e.debug {
					fmt.Fprintf(os.Stderr, "[getChildConstraints.HStack] child flex=%d/%d, flexWidth=%d (cached)\n",
						childInfo.Flex, flexDist.TotalFlexFactor, flexWidth)
				}

				return runtime.BoxConstraints{
					MinWidth:  flexWidth,
					MaxWidth:  flexWidth,
					MinHeight: 0,
					MaxHeight: childMaxHeight,
				}
			}

			// Non-flex child: unconstrained width
			return runtime.BoxConstraints{
				MinWidth:  0,
				MaxWidth:  runtime.Infinity,
				MinHeight: 0,
				MaxHeight: childMaxHeight,
			}
		case "vstack":
			// VStack: calculate flex distribution if applicable
			childMaxWidth := runtime.Infinity
			if parentConstraints.HasBoundedWidth() {
				childMaxWidth = max(0, parentConstraints.MaxWidth-paddingWidth)
			}

			// Check if child has flex
			childInfo := rtui.GetLayoutInfo(child)

			// If child has flex and parent has bounded height, calculate flex height
			if childInfo.Flex > 0 && parentConstraints.HasBoundedHeight() {
				// Use cached flex distribution to avoid O(N²) re-measurement
				flexDist := e.getFlexDistribution(parent, false, childMaxWidth)

				// Calculate available space and this child's flex height
				gapSpace := 0
				if flexDist.ChildCount > 1 {
					gapSpace = (flexDist.ChildCount - 1) * layoutInfo.Gap
				}
				availableHeight := parentConstraints.MaxHeight - paddingHeight - gapSpace
				remainingSpace := availableHeight - flexDist.FixedSize
				flexHeight := (remainingSpace * childInfo.Flex) / flexDist.TotalFlexFactor
				if flexHeight < 0 {
					flexHeight = 0
				}

				return runtime.BoxConstraints{
					MinWidth:  0,
					MaxWidth:  childMaxWidth,
					MinHeight: flexHeight,
					MaxHeight: flexHeight,
				}
			}

			// Non-flex child: unconstrained height
			return runtime.BoxConstraints{
				MinWidth:  0,
				MaxWidth:  childMaxWidth,
				MinHeight: 0,
				MaxHeight: runtime.Infinity,
			}
		case "bordered":
			// Bordered: subtract border (2 units) from constraints
			// Use SubtractPadding helper to handle bounded/unbounded constraints
			return parentConstraints.SubtractPadding(2, 2)
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
	crossAlign := layoutInfo.CrossAlign
	stretchCross := layoutInfo.StretchCross

	childX := x
	for _, child := range box.Children {
		childInfo := rtui.GetLayoutInfo(child.VNode)

		// Stretch child to container height if:
		// 1. Child has flex > 0 (explicit flex), OR
	// 2. Container has StretchCross enabled (auto-stretch all children)
	// IMPORTANT: Only stretch if container height is finite (not Infinity)
	if (childInfo.Flex > 0 || stretchCross) && box.Box.Height < runtime.Infinity {
		child.Box.Height = box.Box.Height
	}

		// Calculate child Y position based on cross-axis alignment
		childY := y
		if child.Box.Height < box.Box.Height {
			switch crossAlign {
			case rtui.AlignCenter:
				childY = y + (box.Box.Height-child.Box.Height)/2
			case rtui.AlignEnd:
				childY = y + box.Box.Height - child.Box.Height
			case rtui.AlignStart, rtui.AlignSpaceBetween, rtui.AlignSpaceAround:
				// Default to top align
				childY = y
			}
		}

		e.calculatePositions(child, childX, childY)
		childX += child.Box.Width + gap
	}
}

// layoutVStack positions children vertically
func (e *Engine) layoutVStack(box *ComputedBox, x, y int) {
	layoutInfo := rtui.GetLayoutInfo(box.VNode)
	gap := layoutInfo.Gap
	crossAlign := layoutInfo.CrossAlign
	stretchCross := layoutInfo.StretchCross

	if os.Getenv("TUI_STRETCH_DEBUG") == "true" {
		fmt.Fprintf(os.Stderr, "[layoutVStack] box=(%d,%d,%dx%d) stretchCross=%v\n",
			box.Box.X, box.Box.Y, box.Box.Width, box.Box.Height, stretchCross)
	}

	childY := y
	for _, child := range box.Children {
		childInfo := rtui.GetLayoutInfo(child.VNode)
		oldWidth := child.Box.Width

		// Stretch child to container width if:
		// 1. Child has flex > 0 (explicit flex), OR
		// 2. Container has StretchCross enabled (auto-stretch all children)
		// IMPORTANT: Only stretch if container width is finite (not Infinity)
		if (childInfo.Flex > 0 || stretchCross) && box.Box.Width < runtime.Infinity {
			child.Box.Width = box.Box.Width
			if os.Getenv("TUI_STRETCH_DEBUG") == "true" {
				fmt.Fprintf(os.Stderr, "[layoutVStack]   stretch child: %d -> %d (text=%q)\n",
					oldWidth, child.Box.Width, rtui.GetTextContent(child.VNode))
			}
		}

		// If text node was stretched, calculate RenderedText with padding
		// IMPORTANT: Only do this if the width is reasonable (not Infinity)
		if child.Box.Width > oldWidth && child.Box.Width > 0 && child.Box.Width < runtime.Infinity {
			if text := rtui.GetTextContent(child.VNode); text != "" {
				// Calculate padding needed
				textWidth := e.measureTextWidth(text)
				padding := child.Box.Width - textWidth
				if padding > 0 && padding < 1000 { // Also check padding is reasonable
					// Get text alignment from props (default: left)
					textAlign := runtime.TextAlignLeft
					if props := child.VNode.Props(); props != nil {
						if align, ok := props["textAlign"].(runtime.TextAlign); ok {
							textAlign = align
						} else if alignStr, ok := props["textAlign"].(string); ok {
							// Support string values too: "left", "center", "right"
							switch alignStr {
							case "center":
								textAlign = runtime.TextAlignCenter
							case "right":
								textAlign = runtime.TextAlignRight
							}
						}
					}

					// Apply padding based on alignment
					var leftPad, rightPad int
					switch textAlign {
					case runtime.TextAlignCenter:
						leftPad = padding / 2
						rightPad = padding - leftPad
					case runtime.TextAlignRight:
						leftPad = padding
						rightPad = 0
					default: // TextAlignLeft
						leftPad = 0
						rightPad = padding
					}

					// Build rendered text with padding
					rendered := strings.Repeat(" ", leftPad) + text + strings.Repeat(" ", rightPad)
					child.RenderedText = rendered
					if os.Getenv("TUI_STRETCH_DEBUG") == "true" {
						fmt.Fprintf(os.Stderr, "[layoutVStack]   renderedText: %q (len=%d, align=%v)\n",
							child.RenderedText, len(child.RenderedText), textAlign)
					}
				}
			}
		}

		// Calculate child X position based on cross-axis alignment
		childX := x
		if child.Box.Width < box.Box.Width {
			switch crossAlign {
			case rtui.AlignCenter:
				childX = x + (box.Box.Width-child.Box.Width)/2
			case rtui.AlignEnd:
				childX = x + box.Box.Width - child.Box.Width
			case rtui.AlignStart, rtui.AlignSpaceBetween, rtui.AlignSpaceAround:
				// Default to left align
				childX = x
			}
		}

		e.calculatePositions(child, childX, childY)
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
//
// The cache key MUST include all properties that affect the VNode's size:
// - Text content for text nodes
// - Label for buttons/checkboxes
// - Value/placeholder for inputs
// - Border style/color/label for bordered nodes
// - Any other props that affect rendering
func (e *Engine) getCacheKey(vnode VNode, constraints runtime.BoxConstraints) LayoutCacheKey {
	key := LayoutCacheKey{
		VNodeType:   vnode.Type().String(),
		Constraints: constraints,
	}

	if keyNode := vnode.Key(); keyNode != "" {
		key.VNodeKey = keyNode
	}

	// Include content hash based on VNode type
	switch {
	// Text nodes - hash the text content
	case vnode.Type() == rtui.VNodeText:
		if text := rtui.GetTextContent(vnode); text != "" {
			key.ContentHash = hashString(text)
		}

	// Button nodes - hash the label (different labels = different sizes)
	case vnode.Type() == rtui.VNodeElement && vnode.Tag() == "button":
		if labeler, ok := vnode.(interface{ Label() string }); ok {
			if label := labeler.Label(); label != "" {
				key.ContentHash = hashString(label)
			}
		}

	// Input nodes - hash value/placeholder (affects displayed width)
	case vnode.Type() == rtui.VNodeElement && vnode.Tag() == "input":
		var content string
		if valuer, ok := vnode.(interface{ Value() string }); ok {
			content = valuer.Value()
		}
		if content == "" {
			if placer, ok := vnode.(interface{ Placeholder() string }); ok {
				content = placer.Placeholder()
			}
		}
		if content != "" {
			key.ContentHash = hashString(content)
		}

	// Checkbox nodes - hash label
	case vnode.Type() == rtui.VNodeElement && vnode.Tag() == "checkbox":
		if labeler, ok := vnode.(interface{ Label() string }); ok {
			if label := labeler.Label(); label != "" {
				key.ContentHash = hashString(label)
			}
		}

	// Bordered nodes - hash style, color, and label
	case vnode.Type() == rtui.VNodeElement && vnode.Tag() == "bordered":
		h := uint64(5381)
		if styler, ok := vnode.(interface{ GetBorderStyle() rtui.BorderStyle }); ok {
			h = h*31 + uint64(styler.GetBorderStyle())
		}
		if colorer, ok := vnode.(interface{ GetBorderColor() string }); ok {
			if color := colorer.GetBorderColor(); color != "" {
				h = h*31 + hashString(color)
			}
		}
		if labeler, ok := vnode.(interface{ GetBorderLabel() string }); ok {
			if label := labeler.GetBorderLabel(); label != "" {
				h = h*31 + hashString(label)
			}
		}
		key.ContentHash = h
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

// hashString computes a hash of a string for cache keys
func hashString(s string) uint64 {
	h := uint64(5381)
	for _, c := range s {
		h = h*31 + uint64(c)
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
