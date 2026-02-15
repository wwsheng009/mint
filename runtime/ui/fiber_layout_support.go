// Package ui provides Fiber-first layout support.
// This file enables Fiber to implement runtime.LayoutMeasurer interface.
//
// During transition phase, BOTH paths exist:
// - OLD: VNode-based layout (deprecated)
// - NEW: Fiber-only layout (preferred)
//
// The key principle is SEPARATION - no mixing of VNode and Fiber access
// within the same layout path.
package ui

import (
	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/runtime"
)

// =============================================================================
// Fiber LayoutMeasurer Implementation
// =============================================================================

// IsLayoutMeasurer marks Fiber as implementing LayoutMeasurer.
// This enables single-pass layout on Fiber trees.
func (f *Fiber) IsLayoutMeasurer() {
	if log.LayoutLogger.Enabled() {
		log.LayoutLogger.Debug("[Fiber.IsLayoutMeasurer] NodeID=%d Tag=%s",
			f.NodeID, f.Tag)
	}
}

// MeasureLayout implements runtime.LayoutMeasurer for Fiber nodes.
// This is the CORE Fiber-first layout algorithm - NO VNode access.
//
// Parameters:
//   measurer: runtime.ChildMeasurer for measuring children
//   constraints: Box constraints for this node
//
// Returns:
//   LayoutMeasurement: Size and child constraints
func (f *Fiber) MeasureLayout(
	measurer runtime.ChildMeasurer,
	constraints runtime.BoxConstraints,
) runtime.LayoutMeasurement {
	if f == nil {
		return runtime.LayoutMeasurement{}
	}

	if log.LayoutLogger.Enabled() {
		log.LayoutLogger.Debug("[Fiber.MeasureLayout] NodeID=%d Tag=%s Constraints=%v",
			f.NodeID, f.Tag, constraints)
	}

	// Get layout properties from Fiber fields (NOT from VNode)
	direction := f.GetDirection()
	gap := f.GetGap()
	padding := f.GetPadding()

	// Calculate inner constraints (subtract padding)
	innerConstraints := f.calculateInnerConstraints(constraints, padding)

	// Get children as Fiber slice
	children := f.GetChildFibers()

	if len(children) == 0 {
		// Empty container - just padding size
		size := runtime.Size{
			Width:  padding[1] + padding[3],
			Height: padding[0] + padding[2],
		}
		return runtime.LayoutMeasurement{
			Size:           size,
			ChildConstraints: []runtime.BoxConstraints{},
		}
	}

	// Dispatch based on direction
	if direction == DirectionRow {
		return f.measureHStackLayoutFiber(measurer, innerConstraints, gap, padding)
	}
	return f.measureVStackLayoutFiber(measurer, innerConstraints, gap, padding)
}

// calculateInnerConstraints calculates constraints for content area (subtracting padding)
func (f *Fiber) calculateInnerConstraints(constraints runtime.BoxConstraints, padding [4]int) runtime.BoxConstraints {
	paddingWidth := padding[1] + padding[3]
	paddingHeight := padding[0] + padding[2]

	innerConstraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  runtime.Infinity,
		MinHeight: 0,
		MaxHeight: runtime.Infinity,
	}

	// Adjust for bounded constraints
	if constraints.HasBoundedWidth() {
		innerMaxWidth := constraints.MaxWidth - paddingWidth
		if innerMaxWidth < 0 {
			innerMaxWidth = 0
		}
		innerConstraints.MaxWidth = innerMaxWidth
	}

	if constraints.HasBoundedHeight() {
		innerMaxHeight := constraints.MaxHeight - paddingHeight
		if innerMaxHeight < 0 {
			innerMaxHeight = 0
		}
		innerConstraints.MaxHeight = innerMaxHeight
	}

	// Apply minimum constraints
	if innerConstraints.MinWidth < constraints.MinWidth {
		innerConstraints.MinWidth = constraints.MinWidth
	}
	if innerConstraints.MinHeight < constraints.MinHeight {
		innerConstraints.MinHeight = constraints.MinHeight
	}

	return innerConstraints
}

// measureHStackLayoutFiber measures horizontal stack layout.
// Fiber-only version using runtime.ChildMeasurer interface.
func (f *Fiber) measureHStackLayoutFiber(
	measurer runtime.ChildMeasurer,
	constraints runtime.BoxConstraints,
	gap int,
	padding [4]int,
) runtime.LayoutMeasurement {
	children := f.GetChildFibers()
	childCount := len(children)

	// Prepare child constraints and sizes
	childConstraints := make([]runtime.BoxConstraints, childCount)
	childSizes := make([]runtime.Size, childCount)

	// Calculate cross-axis (height) constraint
	innerMaxHeight := runtime.Infinity
	if constraints.HasBoundedHeight() {
		innerMaxHeight = max(0, constraints.MaxHeight)
	}

	// Phase 1: Identify flex vs non-flex children
	type flexChild struct {
		fiber  *Fiber
		index  int
		factor int
	}
	var flexChildren []flexChild
	var fixedWidth int
	flexTotalFactor := 0

	for i, child := range children {
		childFlex := child.GetFlex()
		if childFlex > 0 {
			flexChildren = append(flexChildren, flexChild{
				fiber:  child,
				index:  i,
				factor: childFlex,
			})
			flexTotalFactor += childFlex
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
		}
	}

	totalWidth := fixedWidth

	// Phase 2: Distribute space to flex children
	if len(flexChildren) > 0 && constraints.HasBoundedWidth() {
		availableWidth := constraints.MaxWidth
		remainingSpace := availableWidth - fixedWidth

		if remainingSpace > 0 {
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
				childSizes[fc.index] = measurer.MeasureChild(fc.fiber, childConstraints[fc.index])
				totalWidth += childSizes[fc.index].Width
			}
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
			childSizes[fc.index] = measurer.MeasureChild(fc.fiber, childConstraints[fc.index])
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
	paddingWidth := padding[1] + padding[3]
	paddingHeight := padding[0] + padding[2]
	totalWidth += paddingWidth
	totalHeight := maxHeight + paddingHeight

	// Apply constraints
	if totalWidth < constraints.MinWidth {
		totalWidth = constraints.MinWidth
	}
	if totalHeight < constraints.MinHeight {
		totalHeight = constraints.MinHeight
	}

	// Clamp to max
	if constraints.HasBoundedWidth() && totalWidth > constraints.MaxWidth {
		totalWidth = constraints.MaxWidth
	}
	if constraints.HasBoundedHeight() && totalHeight > constraints.MaxHeight {
		totalHeight = constraints.MaxHeight
	}

	return runtime.LayoutMeasurement{
		Size:           runtime.Size{Width: totalWidth, Height: totalHeight},
		ChildConstraints: childConstraints,
		ChildSizes:      childSizes,
	}
}

// measureVStackLayoutFiber measures vertical stack layout.
// Fiber-only version using runtime.ChildMeasurer interface.
func (f *Fiber) measureVStackLayoutFiber(
	measurer runtime.ChildMeasurer,
	constraints runtime.BoxConstraints,
	gap int,
	padding [4]int,
) runtime.LayoutMeasurement {
	children := f.GetChildFibers()
	childCount := len(children)

	// Prepare child constraints and sizes
	childConstraints := make([]runtime.BoxConstraints, childCount)
	childSizes := make([]runtime.Size, childCount)

	// Calculate cross-axis (width) constraint
	innerMaxWidth := runtime.Infinity
	if constraints.HasBoundedWidth() {
		innerMaxWidth = max(0, constraints.MaxWidth)
	}

	// Phase 1: Identify flex vs non-flex children
	type flexChild struct {
		fiber  *Fiber
		index  int
		factor int
	}
	var flexChildren []flexChild
	var fixedHeight int
	flexTotalFactor := 0

	for i, child := range children {
		childFlex := child.GetFlex()
		if childFlex > 0 {
			flexChildren = append(flexChildren, flexChild{
				fiber:  child,
				index:  i,
				factor: childFlex,
			})
			flexTotalFactor += childFlex
		} else {
			// Non-flex child: measure with natural height
			childConstraints[i] = runtime.BoxConstraints{
				MinWidth:  0,
				MaxWidth:  innerMaxWidth,
				MinHeight: 0,
				MaxHeight: runtime.Infinity,
			}
			childSizes[i] = measurer.MeasureChild(child, childConstraints[i])
			if childSizes[i].Height < runtime.Infinity {
				fixedHeight += childSizes[i].Height
			}
		}
	}

	totalHeight := fixedHeight

	// Phase 2: Distribute space to flex children
	if len(flexChildren) > 0 && constraints.HasBoundedHeight() {
		availableHeight := constraints.MaxHeight
		remainingSpace := availableHeight - fixedHeight

		if remainingSpace > 0 {
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
				childSizes[fc.index] = measurer.MeasureChild(fc.fiber, childConstraints[fc.index])
				totalHeight += childSizes[fc.index].Height
			}
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
			childSizes[fc.index] = measurer.MeasureChild(fc.fiber, childConstraints[fc.index])
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
	paddingWidth := padding[1] + padding[3]
	paddingHeight := padding[0] + padding[2]
	totalWidth := maxWidth + paddingWidth
	totalHeight += paddingHeight

	// Apply constraints
	if totalWidth < constraints.MinWidth {
		totalWidth = constraints.MinWidth
	}
	if totalHeight < constraints.MinHeight {
		totalHeight = constraints.MinHeight
	}

	// Clamp to max
	if constraints.HasBoundedWidth() && totalWidth > constraints.MaxWidth {
		totalWidth = constraints.MaxWidth
	}
	if constraints.HasBoundedHeight() && totalHeight > constraints.MaxHeight {
		totalHeight = constraints.MaxHeight
	}

	return runtime.LayoutMeasurement{
		Size:           runtime.Size{Width: totalWidth, Height: totalHeight},
		ChildConstraints: childConstraints,
		ChildSizes:      childSizes,
	}
}

