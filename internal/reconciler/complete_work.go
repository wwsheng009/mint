package reconciler

// =============================================================================
// CompleteWork Phase
// =============================================================================
// CompleteWork is where we finalize the work on a Fiber node.
// For each Fiber, we:
// 1. Create/Update the DOM node (or prepare for rendering)
// 2. Collect child effects
// 3. Return the next Fiber to process
//
// This is the "completion" of processing a work unit.
// After CompleteWork, we move to the next work unit in the traversal.
// =============================================================================

import (
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// CompleteWork completes processing of a Fiber node during the render phase
// Returns the next Fiber to process (usually nil, since we traverse in workLoop)
func CompleteWork(current, workInProgress *Fiber) *Fiber {
	if workInProgress == nil {
		return nil
	}

	// Dispatch based on Fiber type
	switch workInProgress.Type {
	case rtui.VNodeComponent:
		return completeWorkComponent(current, workInProgress)

	case rtui.VNodeText:
		return completeWorkText(current, workInProgress)

	case rtui.VNodeElement:
		return completeWorkElement(current, workInProgress)

	case rtui.VNodeFragment:
		return completeWorkFragment(current, workInProgress)

	default:
		// Unknown type, skip
		return workInProgress
	}
}

// =============================================================================
// Component CompleteWork
// =============================================================================

// completeWorkComponent finalizes a component Fiber
func completeWorkComponent(current, workInProgress *Fiber) *Fiber {
	componentVNode, ok := workInProgress.VNode.(*rtui.ComponentVNode)
	if !ok {
		return workInProgress
	}

	// Store the rendered result for later use during commit
	workInProgress.MemoizedProps = workInProgress.Props

	// ✨ Copy Layer from VNode (in case it changed)
	workInProgress.Layer = workInProgress.VNode.GetLayer()

	// Components don't directly render to buffer
	// Their children are rendered recursively
	_ = componentVNode

	return workInProgress
}

// =============================================================================
// Text CompleteWork
// =============================================================================

// completeWorkText finalizes a text Fiber
func completeWorkText(current, workInProgress *Fiber) *Fiber {
	textVNode, ok := workInProgress.VNode.(*rtui.TextVNode)
	if !ok {
		return workInProgress
	}

	// Store the text content for rendering during commit
	workInProgress.MemoizedState = textVNode.Content()

	return workInProgress
}

// =============================================================================
// Element CompleteWork
// =============================================================================

// completeWorkElement finalizes an element Fiber
func completeWorkElement(current, workInProgress *Fiber) *Fiber {
	elementVNode, ok := workInProgress.VNode.(*rtui.ElementVNode)
	if !ok {
		return workInProgress
	}

	// Store element properties for rendering during commit
	workInProgress.MemoizedProps = workInProgress.Props

	// Store tag/component name for reference
	_ = elementVNode.Tag()

	// === Phase 1: Extract layout info to Fiber ===
	// This enables Fiber-first layout by avoiding VNode delegation
	// Use safe interface methods to check for layout properties
	extractLayoutInfoToFiber(workInProgress, elementVNode)

	return workInProgress
}

// =============================================================================
// Fragment CompleteWork
// =============================================================================

// completeWorkFragment finalizes a fragment Fiber
func completeWorkFragment(current, workInProgress *Fiber) *Fiber {
	// Fragments don't render anything themselves
	// They just group children

	return workInProgress
}

// =============================================================================
// Effect Collection
// =============================================================================

// collectChildEffects collects effect flags from children
// This bubbles up effect flags from descendant fibers
//
// SubtreeFlags Propagation Algorithm:
// - Bottom-up aggregation: child flags propagate to parent during render
// - For each child, we OR both child.Flags and child.SubtreeFlags into parent
// - This creates a complete picture of all effects in the subtree
//
// When called:
// - During performUnitOfWork, after CompleteWork for each fiber
// - Ensures parents know about all descendant effects before commit
//
// Example propagation:
//   Tree before collection:
//     Parent (SubtreeFlags: 0)
//       ├── ChildA (Flags: 2, SubtreeFlags: 4)
//       └── ChildB (Flags: 8, SubtreeFlags: 0)
//
//   After collection (Parent.SubtreeFlags = 2 | 4 | 8 = 14):
//     Parent (SubtreeFlags: 14) ← OR of all descendant flags
//
// Note: SubtreeFlags is NOT automatically propagated upward when flags change.
// The entire tree must be re-rendered to update SubtreeFlags. This is acceptable
// because flag changes trigger re-renders anyway.
func collectChildEffects(workInProgress *Fiber) {
	if workInProgress == nil {
		return
	}

	// Collect flags from all children
	child := workInProgress.Child
	for child != nil {
		// Merge child's flags into parent's subtree flags
		workInProgress.SubtreeFlags |= child.Flags
		workInProgress.SubtreeFlags |= child.SubtreeFlags

		child = child.Sibling
	}
}

// =============================================================================
// Layout Info Extraction (Phase 1)
// =============================================================================

// extractLayoutInfoToFiber extracts layout properties from VNode to Fiber
// This enables Fiber-first layout by storing layout info during completeWork
func extractLayoutInfoToFiber(fiber *Fiber, vnode rtui.VNode) {
	if fiber == nil || vnode == nil {
		return
	}

	// Use safe interface methods to extract layout info
	// Check for Direction method
	if dirGetter, ok := vnode.(interface{ Direction() rtui.Direction }); ok {
		fiber.LayoutDirection = dirGetter.Direction()
	}
	// Check for Align method
	if alignGetter, ok := vnode.(interface{ Align() rtui.Align }); ok {
		fiber.LayoutAlign = alignGetter.Align()
	}
	// Check for CrossAlign method
	if crossAlignGetter, ok := vnode.(interface{ CrossAlign() rtui.Align }); ok {
		fiber.LayoutCrossAlign = crossAlignGetter.CrossAlign()
	}
	// Check for Gap method
	if gapGetter, ok := vnode.(interface{ Gap() int }); ok {
		fiber.LayoutGap = gapGetter.Gap()
	}
	// Check for Padding method
	if paddingGetter, ok := vnode.(interface{ Padding() [4]int }); ok {
		fiber.LayoutPadding = paddingGetter.Padding()
	}
	// Check for Flex method
	if flexGetter, ok := vnode.(interface{ Flex() int }); ok {
		fiber.LayoutFlex = flexGetter.Flex()
	}
}
