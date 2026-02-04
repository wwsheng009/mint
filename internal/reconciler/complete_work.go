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
