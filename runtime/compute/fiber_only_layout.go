// Package compute provides Fiber-first layout computation.
// This file contains ONLY Fiber-based layout computation - NO VNode access.
//
// Fiber-First Architecture Principles:
// 1. Build ComputedBox tree from ONLY Fiber tree structure
// 2. Use Fiber.MeasureLayout() for layout computation (Fiber implements LayoutMeasurer)
// 3. Use runtime.ChildMeasurer to bridge between compute and ui packages
// 4. This is a PURE FUNCTION - no side effects on Fiber
//
// During transition:
// - OLD: buildComputedBox() uses VNode (marked deprecated)
// - NEW: BuildComputedBoxFiberOnly() uses ONLY Fiber
//
// See: docs/plan/fiber/fiber_first.md
// See: docs/plan/fiber/FIBER_FIRST_LAYOUT_PLAN.md
package compute

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/runtime"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Fiber-Only Layout Entry Point
// =============================================================================

// BuildComputedBoxFiberOnly performs layout on a Fiber tree ONLY.
// This is the main entry point for Fiber-first layout computation.
//
// Parameters:
//
//	root: The root Fiber node of the tree to layout
//	constraints: Box constraints for the entire tree
//
// Returns:
//
//	*ComputedLayout containing root ComputedBox and HitMap
//	error if layout fails
//
// Fiber-First Guarantees:
// 1. NEVER accesses VNode for layout properties
// 2. Uses ONLY Fiber tree structure (Child->Sibling)
// 3. Uses Fiber.MeasureLayout() for flex layout computation
// 4. Builds ComputedBox tree with NodeID from Fiber
func (e *Engine) BuildComputedBoxFiberOnly(root *rtui.Fiber, constraints runtime.BoxConstraints) (*ComputedLayout, error) {
	if root == nil {
		return nil, fmt.Errorf("cannot layout nil Fiber tree")
	}

	if e.debug {
		log.EngineLogger.Debug("[BuildComputedBoxFiberOnly] START NodeID=%d Tag=%s Constraints=%v",
			root.NodeID, root.Tag, constraints)
	}

	// Reset flex cache for each layout pass
	e.flexCache = make(map[string]*FlexDistributionInfo)

	// Build ComputedBox tree from Fiber (first pass: measurement)
	rootBox := e.buildFiberOnlyBox(root, nil, constraints)
	if rootBox == nil {
		layout := NewComputedLayout(nil)
		layout.HitMap = nil // Empty HitMap
		return layout, nil
	}

	// Calculate positions (second pass: layout)
	e.calculateFiberOnlyPositions(rootBox, 0, 0)

	// Clear dirty flags after layout
	rootBox.ClearDirty()

	// Build HitMap from ComputedBox tree
	hitMap := e.buildHitMapFromComputedBoxes(rootBox)

	layout := NewComputedLayout(rootBox)
	layout.HitMap = hitMap

	if e.debug {
		log.EngineLogger.Debug("[BuildComputedBoxFiberOnly] COMPLETE RootBox=%v HitMapSize=%d",
			rootBox.Box, hitMap.Size())
	}

	return layout, nil
}

// =============================================================================
// Fiber-Only Box Building
// =============================================================================

// buildFiberOnlyBox creates ComputedBox from Fiber tree ONLY.
// This is the CORE Fiber-first layout method - NO VNode access.
//
// Parameters:
//
//	fiber: The Fiber node to build ComputedBox for
//	parent: Parent ComputedBox (for tree structure)
//	constraints: Box constraints for layout
//
// Returns:
//
//	*ComputedBox for this Fiber node
func (e *Engine) buildFiberOnlyBox(
	fiber *rtui.Fiber,
	parent *ComputedBox,
	constraints runtime.BoxConstraints) *ComputedBox {
	if fiber == nil {
		return nil
	}

	if e.debug {
		depth := e.getTraceDepth()
		e.incrementTraceDepth()
		defer e.decrementTraceDepth()

		indent := strings.Repeat("  ", depth)
		log.EngineLogger.Debug("%s[FiberOnly.ENTER] NodeID=%d Tag=%s Constraints=%v",
			indent, fiber.NodeID, fiber.Tag, constraints)
	}

	// Create base ComputedBox using ONLY Fiber properties
	box := &ComputedBox{
		Parent:     parent,
		NodeID:     fiber.NodeID,
		Layer:      fiber.Layer,
		ChildFiber: fiber, // Store Fiber reference for NodeID propagation
		Box: runtime.Box{
			X:      0,
			Y:      0,
			Width:  0,
			Height: 0,
		},
	}

	// Measure layout using Fiber.MeasureLayout (single-pass flex algorithm)
	measurement := fiber.MeasureLayout(e, constraints)
	box.Box.Width = measurement.Size.Width
	box.Box.Height = measurement.Size.Height

	// IMPORTANT: Ensure children are measured using Fiber-only path
	// The measurer passed to Fiber.MeasureLayout is 'e' which is *Engine,
	// and it calls buildFiberOnlyBox recursively for children.
	// This ensures pure Fiber data flow without VNode access.
	//
	// NOTE: For leaf nodes like Text that don't implement LayoutMeasurer,
	// we need to estimate their size based on content length.
	boxSize := estimateBoxSize(fiber, constraints)
	if boxSize.Width > 0 && boxSize.Height > 0 {
		box.Box.Width = boxSize.Width
		box.Box.Height = boxSize.Height
	}

	// Build children ComputedBox using Fiber tree traversal
	children := fiber.GetChildFibers()
	childCount := len(children)

	// Get child constraints from measurement
	var childConstraints []runtime.BoxConstraints
	if measurement.ChildConstraints != nil && len(measurement.ChildConstraints) == childCount {
		// Use measurement constraints for each child
		childConstraints = measurement.ChildConstraints
	} else {
		// Fallback: each child gets parent constraints
		childConstraints = make([]runtime.BoxConstraints, childCount)
		for i := range childConstraints {
			childConstraints[i] = constraints
		}
	}

	box.Children = make([]*ComputedBox, 0, childCount)

	for i, childFiber := range children {
		childBox := e.buildFiberOnlyBox(childFiber, box, childConstraints[i])
		if childBox != nil {
			box.Children = append(box.Children, childBox)
		}
	}

	if e.debug {
		depth := e.getTraceDepth()
		e.incrementTraceDepth()
		defer e.decrementTraceDepth()

		indent := strings.Repeat("  ", depth)
		log.EngineLogger.Debug("%s[FiberOnly.LEAVE] NodeID=%d Size=%dx%d Children=%d",
			indent, fiber.NodeID, box.Box.Width, box.Box.Height, len(box.Children))
	}

	return box
}

// estimateBoxSize estimates the size of a Fiber node that doesn't implement LayoutMeasurer.
// This is used for leaf nodes like Text that have no intrinsic size.
func estimateBoxSize(fiber *rtui.Fiber, constraints runtime.BoxConstraints) runtime.Size {
	if fiber == nil {
		return runtime.Size{}
	}

	tag := fiber.Tag
	if tag == "text" || tag == "unknown" {
		// Text node - estimate based on content
		// Try to get content from MemoizedState or Props
		content := ""
		if props := fiber.Props; props != nil {
			if c, ok := props["content"]; ok {
				if s, ok := c.(string); ok {
					content = s
				}
			}
		}
		// Each character is approximately 1 column wide in terminal
		width := len([]rune(content))
		if width > constraints.MaxWidth {
			width = constraints.MaxWidth
		}
		if width < constraints.MinWidth {
			width = constraints.MinWidth
		}
		height := 1
		if height > constraints.MaxHeight {
			height = constraints.MaxHeight
		}
		return runtime.Size{Width: width, Height: height}
	}

	return runtime.Size{}
}

// =============================================================================
// Fiber-Only Position Calculation
// =============================================================================

// calculateFiberOnlyPositions computes (x, y) positions for ComputedBox tree.
// This is the second pass of layout - after all sizes are known.
//
// Fiber-First: Uses ChildFiber.Tag() to determine layout type
func (e *Engine) calculateFiberOnlyPositions(box *ComputedBox, x, y int) {
	box.Box.X = x
	box.Box.Y = y

	if e.debug {
		tagStr := "no-fiber"
		if fiber := box.GetFiber(); fiber != nil {
			tagStr = fiber.Tag
		}
		log.EngineLogger.Debug("[FiberOnly.Position] %s at (%d,%d) size=%dx%d",
			tagStr, x, y, box.Box.Width, box.Box.Height)
	}

	// Get layout tag from Fiber
	var tag string
	if fiber := box.GetFiber(); fiber != nil {
		tag = fiber.Tag
	}

	// Dispatch based on tag
	switch tag {
	case "hstack", "row":
		e.layoutFiberOnlyHStack(box, x, y)
	case "vstack", "column":
		e.layoutFiberOnlyVStack(box, x, y)
	case "bordered":
		e.layoutFiberOnlyBordered(box, x, y)
	default:
		e.layoutFiberOnlyDefault(box, x, y)
	}
}

// layoutFiberOnlyHStack positions children horizontally.
// Fiber-First: Gets layout info from Fiber fields
func (e *Engine) layoutFiberOnlyHStack(box *ComputedBox, x, y int) {
	fiber := box.GetFiber()
	if fiber == nil {
		e.layoutFiberOnlyDefault(box, x, y)
		return
	}

	gap := fiber.GetGap()
	align := fiber.GetAlign()
	crossAlign := fiber.GetCrossAlign()

	// Calculate total child width
	totalChildWidth := 0
	for _, child := range box.Children {
		totalChildWidth += child.Box.Width
	}
	if len(box.Children) > 0 {
		totalChildWidth += (len(box.Children) - 1) * gap
	}

	// Calculate starting X based on main-axis alignment
	childX := x
	switch align {
	case rtui.AlignCenter:
		if totalChildWidth < box.Box.Width {
			childX = x + (box.Box.Width-totalChildWidth)/2
		}
	case rtui.AlignEnd:
		if totalChildWidth < box.Box.Width {
			childX = x + box.Box.Width - totalChildWidth
		}
	case rtui.AlignSpaceBetween:
		if len(box.Children) > 1 {
			totalButtonWidth := 0
			for _, child := range box.Children {
				totalButtonWidth += child.Box.Width
			}
			if totalButtonWidth < box.Box.Width {
				gap = (box.Box.Width - totalButtonWidth) / (len(box.Children) - 1)
			}
		}
	case rtui.AlignSpaceAround:
		if len(box.Children) > 0 && totalChildWidth < box.Box.Width {
			extraSpace := box.Box.Width - totalChildWidth
			gap = extraSpace / len(box.Children)
			childX = x + gap/2
		}
	}

	for _, child := range box.Children {
		// Calculate child Y based on cross-axis alignment
		childY := y
		if child.Box.Height < box.Box.Height {
			switch crossAlign {
			case rtui.AlignCenter:
				childY = y + (box.Box.Height-child.Box.Height)/2
			case rtui.AlignEnd:
				childY = y + box.Box.Height - child.Box.Height
			}
		}

		e.calculateFiberOnlyPositions(child, childX, childY)
		childX += child.Box.Width + gap
	}
}

// layoutFiberOnlyVStack positions children vertically.
// Fiber-First: Gets layout info from Fiber fields
func (e *Engine) layoutFiberOnlyVStack(box *ComputedBox, x, y int) {
	fiber := box.GetFiber()
	if fiber == nil {
		e.layoutFiberOnlyDefault(box, x, y)
		return
	}

	gap := fiber.GetGap()
	align := fiber.GetAlign()
	crossAlign := fiber.GetCrossAlign()

	// Calculate total child height
	totalChildHeight := 0
	for _, child := range box.Children {
		totalChildHeight += child.Box.Height
	}
	if len(box.Children) > 0 {
		totalChildHeight += (len(box.Children) - 1) * gap
	}

	// Calculate starting Y based on main-axis alignment
	childY := y
	switch align {
	case rtui.AlignCenter:
		if totalChildHeight < box.Box.Height {
			childY = y + (box.Box.Height-totalChildHeight)/2
		}
	case rtui.AlignEnd:
		if totalChildHeight < box.Box.Height {
			childY = y + box.Box.Height - totalChildHeight
		}
	case rtui.AlignSpaceBetween:
		if len(box.Children) > 1 {
			totalButtonHeight := 0
			for _, child := range box.Children {
				totalButtonHeight += child.Box.Height
			}
			if totalButtonHeight < box.Box.Height {
				gap = (box.Box.Height - totalButtonHeight) / (len(box.Children) - 1)
			}
		}
	case rtui.AlignSpaceAround:
		if len(box.Children) > 0 && totalChildHeight < box.Box.Height {
			extraSpace := box.Box.Height - totalChildHeight
			gap = extraSpace / len(box.Children)
			childY = y + gap/2
		}
	}

	for _, child := range box.Children {
		// Calculate child X based on cross-axis alignment
		childX := x
		if child.Box.Width < box.Box.Width {
			switch crossAlign {
			case rtui.AlignCenter:
				childX = x + (box.Box.Width-child.Box.Width)/2
			case rtui.AlignEnd:
				childX = x + box.Box.Width - child.Box.Width
			}
		}

		e.calculateFiberOnlyPositions(child, childX, childY)
		childY += child.Box.Height + gap
	}
}

// layoutFiberOnlyBordered positions content inside border.
func (e *Engine) layoutFiberOnlyBordered(box *ComputedBox, x, y int) {
	// Content starts at (x+1, y+1) - inside the border
	for _, child := range box.Children {
		e.calculateFiberOnlyPositions(child, x+1, y+1)
	}
}

// layoutFiberOnlyDefault positions children vertically (default).
func (e *Engine) layoutFiberOnlyDefault(box *ComputedBox, x, y int) {
	childY := y
	for _, child := range box.Children {
		e.calculateFiberOnlyPositions(child, x, childY)
		childY += child.Box.Height
	}
}
