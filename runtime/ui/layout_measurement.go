// Package ui provides the single-pass layout measurement implementation for LayoutNode
package ui

import (
	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/runtime"
)

// =============================================================================
// LayoutNode implements LayoutMeasurer
// =============================================================================

func (l *LayoutNode) IsLayoutMeasurer() {}

// MeasureLayout implements the LayoutMeasurer interface for single-pass layout.
//
// This method:
// 1. Measures all children with appropriate constraints
// 2. Calculates the node's own size based on children
// 3. Returns the constraints used for each child
//
// The returned constraints can be reused by buildComputedBox, ensuring
// consistency between measurement and layout phases.
func (l *LayoutNode) MeasureLayout(
	measurer runtime.ChildMeasurer,
	constraints runtime.BoxConstraints,
) runtime.LayoutMeasurement {
	if l.direction == DirectionRow {
		return l.measureHStackLayout(measurer, constraints)
	}
	return l.measureVStackLayout(measurer, constraints)
}

// =============================================================================
// HStack Layout Measurement
// =============================================================================

func (l *LayoutNode) measureHStackLayout(
	measurer runtime.ChildMeasurer,
	constraints runtime.BoxConstraints,
) runtime.LayoutMeasurement {
	children := l.Children()
	if len(children) == 0 {
		return runtime.LayoutMeasurement{
			Size: runtime.Size{
				Width:  constraints.MinWidth,
				Height: constraints.MinHeight,
			},
			ChildConstraints: []runtime.BoxConstraints{},
		}
	}

	layoutInfo := GetLayoutInfo(l)
	gap := layoutInfo.Gap
	padding := layoutInfo.Padding // top, right, bottom, left
	paddingWidth := padding[1] + padding[3]
	paddingHeight := padding[0] + padding[2]

	// Calculate inner height constraint
	innerMaxHeight := runtime.Infinity
	if constraints.HasBoundedHeight() {
		innerMaxHeight = max(0, constraints.MaxHeight-paddingHeight)
	}

	// ⭐ Check for explicit width/height props and use them to constrain layout
	// This allows HStack().Width(n).Height(n) to properly constrain children
	if props := l.Props(); props != nil {
		if width, ok := props["width"].(int); ok && width > 0 {
			// Use explicit width as MaxWidth constraint
			constraints.MaxWidth = width
			// Ensure MinWidth doesn't exceed MaxWidth
			if constraints.MinWidth > width {
				constraints.MinWidth = width
			}
			log.LayoutLogger.IfEnabled().Debug("[HStack.MeasureLayout] Using width prop: %d", width)
		}
		// ⭐ CRITICAL FIX: Also check height prop for bounded height constraint
		// This ensures flex children receive bounded constraints
		if height, ok := props["height"].(int); ok && height > 0 {
			// Use explicit height as bounded constraint
			constraints.MaxHeight = height
			// Ensure MinHeight doesn't exceed MaxHeight
			if constraints.MinHeight > height {
				constraints.MinHeight = height
			}
			// Recalculate innerMaxHeight with new constraint
			innerMaxHeight = max(0, height-paddingHeight)
			log.RenderLogger.IfEnabled().Debug("[HStack.MeasureLayout] Using height prop: %d, innerMaxHeight=%d", height, innerMaxHeight)

		}
	}

	// First pass: identify flex children and measure non-flex children
	type flexChild struct {
		child  VNode
		index  int
		factor int
	}
	var flexChildren []flexChild
	var fixedWidth int
	flexTotalFactor := 0

	childConstraints := make([]runtime.BoxConstraints, len(children))
	childSizes := make([]runtime.Size, len(children))

	for i, child := range children {
		childInfo := GetLayoutInfo(child)
		log.RenderLogger.Debug("[HStack.MeasureLayout] child %d: GetLayoutInfo.Flex=%d, tag=%s",
			i, childInfo.Flex, child.Type().String())

		if childInfo.Flex > 0 {
			flexChildren = append(flexChildren, flexChild{
				child:  child,
				index:  i,
				factor: childInfo.Flex,
			})
			flexTotalFactor += childInfo.Flex
		} else {
			// Non-flex child: measure with natural width
			childConstraints[i] = runtime.BoxConstraints{
				MinWidth:  0,
				MaxWidth:  runtime.Infinity,
				MinHeight: 0,
				MaxHeight: innerMaxHeight,
			}
			childSizes[i] = measurer.MeasureChild(child, childConstraints[i])
			fixedWidth += childSizes[i].Width

			log.RenderLogger.Debug("[HStack.MeasureLayout] non-flex child %d: size=%v",
				i, childSizes[i])

		}
	}

	totalWidth := fixedWidth

	// Distribute space to flex children if bounded width
	if len(flexChildren) > 0 && constraints.HasBoundedWidth() {
		availableWidth := constraints.MaxWidth - paddingWidth - (len(children)-1)*gap
		remainingSpace := availableWidth - fixedWidth

		if remainingSpace > 0 {
			log.RenderLogger.Debug("[HStack.MeasureLayout] flex: available=%d, fixed=%d, remaining=%d",
				availableWidth, fixedWidth, remainingSpace)
		}

		for _, fc := range flexChildren {
			flexWidth := (remainingSpace * fc.factor) / flexTotalFactor
			if flexWidth < 0 {
				flexWidth = 0
			}

			childConstraints[fc.index] = runtime.BoxConstraints{
				MinWidth:  flexWidth,
				MaxWidth:  flexWidth,
				MinHeight: 0,
				MaxHeight: innerMaxHeight,
			}
			childSizes[fc.index] = measurer.MeasureChild(fc.child, childConstraints[fc.index])
			totalWidth += childSizes[fc.index].Width

			log.RenderLogger.Debug("[HStack.MeasureLayout] flex child %d: flexWidth=%d, size=%v",
				fc.index, flexWidth, childSizes[fc.index])

		}
	} else {
		// No bounded width: measure flex children naturally
		for _, fc := range flexChildren {
			childConstraints[fc.index] = runtime.BoxConstraints{
				MinWidth:  0,
				MaxWidth:  runtime.Infinity,
				MinHeight: 0,
				MaxHeight: innerMaxHeight,
			}
			childSizes[fc.index] = measurer.MeasureChild(fc.child, childConstraints[fc.index])
			totalWidth += childSizes[fc.index].Width
		}
	}

	// Calculate max height from children
	maxHeight := 0
	for _, size := range childSizes {
		if size.Height > maxHeight && size.Height < runtime.Infinity {
			maxHeight = size.Height
		}
	}

	// Add padding
	totalWidth += paddingWidth
	maxHeight += paddingHeight

	// Cross-axis filling: fill available height
	if constraints.HasBoundedHeight() && maxHeight < constraints.MaxHeight {
		maxHeight = constraints.MaxHeight
	}

	// Main-axis filling: fill available width (important for tight constraints)
	// When HStack is in VStack with tight width constraints, it should expand to fill
	if constraints.HasBoundedWidth() && totalWidth < constraints.MaxWidth {
		totalWidth = constraints.MaxWidth
	}

	// Apply MinWidth constraint
	if totalWidth < constraints.MinWidth {
		totalWidth = constraints.MinWidth
	}

	// Apply MinHeight constraint
	if maxHeight < constraints.MinHeight {
		maxHeight = constraints.MinHeight
	}

	// Clamp to MaxWidth
	if constraints.HasBoundedWidth() && totalWidth > constraints.MaxWidth {
		totalWidth = constraints.MaxWidth
	}

	// Clamp to MaxHeight
	if constraints.HasBoundedHeight() && maxHeight > constraints.MaxHeight {
		maxHeight = constraints.MaxHeight
	}

	log.RenderLogger.Debug("[HStack.MeasureLayout] RETURN: Size=%v",
		runtime.Size{Width: totalWidth, Height: maxHeight})

	return runtime.LayoutMeasurement{
		Size:             runtime.Size{Width: totalWidth, Height: maxHeight},
		ChildConstraints: childConstraints,
		ChildSizes:       childSizes,
	}
}

// =============================================================================
// VStack Layout Measurement
// =============================================================================

func (l *LayoutNode) measureVStackLayout(
	measurer runtime.ChildMeasurer,
	constraints runtime.BoxConstraints,
) runtime.LayoutMeasurement {
	children := l.Children()
	if len(children) == 0 {
		return runtime.LayoutMeasurement{
			Size: runtime.Size{
				Width:  constraints.MinWidth,
				Height: constraints.MinHeight,
			},
			ChildConstraints: []runtime.BoxConstraints{},
		}
	}

	layoutInfo := GetLayoutInfo(l)
	gap := layoutInfo.Gap
	padding := layoutInfo.Padding // top, right, bottom, left
	paddingWidth := padding[1] + padding[3]
	paddingHeight := padding[0] + padding[2]

	// Calculate inner width constraint
	innerMaxWidth := runtime.Infinity
	if constraints.HasBoundedWidth() {
		innerMaxWidth = max(0, constraints.MaxWidth-paddingWidth)
	}

	// ⭐ Check for explicit width/height props and use them to constrain layout
	// This allows VStack().Width(n).Height(n) to properly constrain children
	if props := l.Props(); props != nil {
		if width, ok := props["width"].(int); ok && width > 0 {
			// Use explicit width as MaxWidth constraint
			constraints.MaxWidth = width
			// Ensure MinWidth doesn't exceed MaxWidth
			if constraints.MinWidth > width {
				constraints.MinWidth = width
			}
			// Recalculate innerMaxWidth with new constraint
			innerMaxWidth = max(0, width-paddingWidth)
			log.LayoutLogger.IfEnabled().Debug("[VStack.MeasureLayout] Using width prop: %d, innerMaxWidth=%d", width, innerMaxWidth)
		}
		// ⭐ CRITICAL FIX: Also check height prop for bounded height constraint
		// This ensures flex children receive bounded constraints and can properly
		// distribute remaining space. Without this, TreeView and other flex children
		// would receive unbounded height and render all content.
		if height, ok := props["height"].(int); ok && height > 0 {
			// Use explicit height as bounded constraint
			constraints.MaxHeight = height
			// Ensure MinHeight doesn't exceed MaxHeight
			if constraints.MinHeight > height {
				constraints.MinHeight = height
			}
			log.RenderLogger.IfEnabled().Debug("[VStack.MeasureLayout] Using height prop: %d", height)

		}
	}

	// First pass: identify flex children and measure non-flex children
	type flexChild struct {
		child  VNode
		index  int
		factor int
	}
	var flexChildren []flexChild
	var fixedHeight int
	flexTotalFactor := 0

	childConstraints := make([]runtime.BoxConstraints, len(children))
	childSizes := make([]runtime.Size, len(children))

	for i, child := range children {
		childInfo := GetLayoutInfo(child)
		if childInfo.Flex > 0 {
			flexChildren = append(flexChildren, flexChild{
				child:  child,
				index:  i,
				factor: childInfo.Flex,
			})
			flexTotalFactor += childInfo.Flex
		} else {
			// Non-flex child: measure with natural height
			// SPECIAL CASE: HStack in VStack may fill width if needed for alignment
			childMinWidth := 0
			childTag := ""
			if tagger, ok := child.(interface{ Tag() string }); ok {
				childTag = tagger.Tag()
			}
			// Only force-stretch HStack if it needs width for alignment and doesn't have explicit width
			if innerMaxWidth != runtime.Infinity && (childTag == "hstack" || childTag == "row") {
				// Check if child has explicit width constraint - if so, respect it
				childInfo := GetLayoutInfo(child)
				hasExplicitWidth := false
				if props := child.Props(); props != nil {
					if w, ok := props["width"].(int); ok && w > 0 {
						hasExplicitWidth = true
					}
				}
				// HStack only fills VStack width if:
				// 1. It needs width for alignment (AlignCenter/AlignEnd/AlignSpace*)
				// 2. AND doesn't have explicit width set by user
				// 3. AND has flex or FillWidth enabled
				mainAlign := childInfo.Align
				needsAlignmentWidth := mainAlign == AlignCenter || mainAlign == AlignEnd ||
					mainAlign == AlignSpaceBetween || mainAlign == AlignSpaceAround
				if needsAlignmentWidth && !hasExplicitWidth && (childInfo.Flex > 0 || childInfo.FillWidth) {
					childMinWidth = innerMaxWidth
				}
			}

			childConstraints[i] = runtime.BoxConstraints{
				MinWidth:  childMinWidth,
				MaxWidth:  innerMaxWidth,
				MinHeight: 0,
				MaxHeight: runtime.Infinity,
			}
			childSizes[i] = measurer.MeasureChild(child, childConstraints[i])

			log.RenderLogger.Debug("[VStack.MeasureLayout] non-flex child %d (tag=%s): constraints=%v, size=%v",
				i, childTag, childConstraints[i], childSizes[i])

			if childSizes[i].Height < runtime.Infinity {
				fixedHeight += childSizes[i].Height
			}
		}
	}

	totalHeight := fixedHeight

	// Distribute space to flex children if bounded height
	if len(flexChildren) > 0 && constraints.HasBoundedHeight() {
		availableHeight := constraints.MaxHeight - paddingHeight - (len(children)-1)*gap
		remainingSpace := availableHeight - fixedHeight

		if remainingSpace > 0 {
			log.RenderLogger.Debug("[VStack.MeasureLayout] flex: available=%d, fixed=%d, remaining=%d",
				availableHeight, fixedHeight, remainingSpace)
		}

		for _, fc := range flexChildren {
			flexHeight := (remainingSpace * fc.factor) / flexTotalFactor
			if flexHeight < 0 {
				flexHeight = 0
			}

			childConstraints[fc.index] = runtime.BoxConstraints{
				MinWidth:  0,
				MaxWidth:  innerMaxWidth,
				MinHeight: flexHeight,
				MaxHeight: flexHeight,
			}
			childSizes[fc.index] = measurer.MeasureChild(fc.child, childConstraints[fc.index])
			totalHeight += childSizes[fc.index].Height

			log.RenderLogger.Debug("[VStack.MeasureLayout] flex child %d: flexHeight=%d, size=%v",
				fc.index, flexHeight, childSizes[fc.index])

		}
	} else {
		// No bounded height: measure flex children naturally
		for _, fc := range flexChildren {
			childConstraints[fc.index] = runtime.BoxConstraints{
				MinWidth:  0,
				MaxWidth:  innerMaxWidth,
				MinHeight: 0,
				MaxHeight: runtime.Infinity,
			}
			childSizes[fc.index] = measurer.MeasureChild(fc.child, childConstraints[fc.index])
			if childSizes[fc.index].Height < runtime.Infinity {
				totalHeight += childSizes[fc.index].Height
			}
		}
	}

	// Calculate max width from children
	maxWidth := 0
	for _, size := range childSizes {
		if size.Width > maxWidth && size.Width < runtime.Infinity {
			maxWidth = size.Width
		}
	}

	// Add padding
	maxWidth += paddingWidth
	totalHeight += paddingHeight

	// Cross-axis filling: fill available width
	if constraints.HasBoundedWidth() && maxWidth < constraints.MaxWidth {
		maxWidth = constraints.MaxWidth
	}

	// Apply MinWidth constraint
	if maxWidth < constraints.MinWidth {
		maxWidth = constraints.MinWidth
	}

	// Clamp to MaxWidth
	if constraints.HasBoundedWidth() && maxWidth > constraints.MaxWidth {
		maxWidth = constraints.MaxWidth
	}

	log.RenderLogger.Debug("[VStack.MeasureLayout] RETURN: Size=%v",
		runtime.Size{Width: maxWidth, Height: totalHeight})

	return runtime.LayoutMeasurement{
		Size:             runtime.Size{Width: maxWidth, Height: totalHeight},
		ChildConstraints: childConstraints,
		ChildSizes:       childSizes,
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

// getTagForDebug returns the tag of a VNode for debugging
func getTagForDebug(vnode VNode) string {
	if vnode == nil {
		return "nil"
	}
	if tagger, ok := vnode.(interface{ Tag() string }); ok {
		return tagger.Tag()
	}
	return vnode.Type().String()
}
