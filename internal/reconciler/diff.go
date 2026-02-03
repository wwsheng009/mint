package reconciler

// =============================================================================
// Fiber Reconciliation Algorithm
// =============================================================================
// reconcileChildren implements the Fiber tree diffing algorithm.
// This operates on Fiber nodes to reconcile old and new children.
// =============================================================================

import (
	"fmt"
	"os"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// reconcileChildren reconciles the current children with new children
// Returns the first child of the reconciled Fiber tree
func reconcileChildren(
	returnFiber *Fiber,
	currentFirstChild *Fiber,
	newChildren []rtui.VNode,
	lanes Lane,
) *Fiber {
	// Validate keys for list children (React-style warning)
	if currentReconciler != nil && currentReconciler.keyValidator != nil {
		var parentVNode rtui.VNode
		if returnFiber != nil {
			parentVNode = returnFiber.VNode
		}
		currentReconciler.keyValidator.ValidateChildren(parentVNode, newChildren)
	}

	// If no new children, delete all existing children
	if len(newChildren) == 0 {
		return nil
	}

	// If no existing children, create all new children
	if currentFirstChild == nil {
		return createAllNewChildren(returnFiber, newChildren, lanes)
	}

	// Both old and new children exist - reconcile
	return reconcileExistingChildren(returnFiber, currentFirstChild, newChildren, lanes)
}

// createAllNewChildren creates Fiber nodes for all new children
func createAllNewChildren(returnFiber *Fiber, children []rtui.VNode, lanes Lane) *Fiber {
	var firstChild *Fiber
	var previousChild *Fiber

	for _, childVNode := range children {
		child := createChildFiber(returnFiber, childVNode, lanes)

		if firstChild == nil {
			firstChild = child
		} else {
			previousChild.Sibling = child
		}

		previousChild = child
	}

	return firstChild
}

// reconcileExistingChildren reconciles existing children with new children
// This is a simplified position-based reconciliation with key-based matching
func reconcileExistingChildren(
	returnFiber *Fiber,
	currentFirstChild *Fiber,
	newChildren []rtui.VNode,
	lanes Lane,
) *Fiber {
	var firstChild *Fiber
	var previousChild *Fiber
	currentChild := currentFirstChild

	for _, childVNode := range newChildren {
		var child *Fiber

		// Try to match with current child or any of its siblings
		// This handles cases where a child later in the list matches
		matchedChild := findMatchingChild(currentChild, childVNode)

		if matchedChild != nil {
			// Found a match - reuse existing fiber
			child = cloneExistingFiber(returnFiber, matchedChild, childVNode)

			// Mark all children between currentChild and matchedChild for deletion
			// (they were skipped over and are no longer in the tree)
			for currentChild != nil && currentChild != matchedChild {
				markForDeletion(currentChild)
				currentChild = currentChild.Sibling
			}

			// Advance past the matched child
			currentChild = matchedChild.Sibling
		} else {
			// No match found - create new fiber
			child = createChildFiber(returnFiber, childVNode, lanes)
			// The currentChild remains unchanged (will be processed in next iteration or deleted)
		}

		if firstChild == nil {
			firstChild = child
		} else {
			previousChild.Sibling = child
		}

		previousChild = child
	}

	// Delete remaining current children that weren't matched
	// These nodes are being removed from the tree
	for currentChild != nil {
		markForDeletion(currentChild)
		currentChild = currentChild.Sibling
	}

	return firstChild
}

// findMatchingChild searches for a child fiber that matches the given VNode
// It checks currentChild and all its siblings for a match based on key and type
func findMatchingChild(currentChild *Fiber, vnode rtui.VNode) *Fiber {
	for child := currentChild; child != nil; child = child.Sibling {
		if shouldUpdate(child, vnode) {
			return child
		}
	}
	return nil
}

// shouldUpdate checks if a current fiber can be updated with new VNode
// This follows React's reconciliation logic:
// 1. Key is primary - different keys mean different elements
// 2. Type is secondary - same key but different type means replace
func shouldUpdate(current *Fiber, vnode rtui.VNode) bool {
	if current == nil || vnode == nil {
		return false
	}

	// Get the keys for comparison
	currentKey := current.Key
	newKey := vnode.Key()

	// If keys differ, this is definitely not the same element
	if currentKey != newKey {
		return false
	}

	// Check if types match
	if current.Type != vnode.Type() {
		return false
	}

	// For components, check if the component function is the same
	if current.Type == rtui.VNodeComponent {
		currentComp, ok1 := current.VNode.(*rtui.ComponentVNode)
		newComp, ok2 := vnode.(*rtui.ComponentVNode)
		if ok1 && ok2 {
			// Compare component names since functions cannot be directly compared
			// Same key + same name = same component
			return currentComp.Name() == newComp.Name()
		}
	}

	// For elements, check if tag is the same
	if current.Type == rtui.VNodeElement {
		currentElem, ok1 := current.VNode.(*rtui.ElementVNode)
		newElem, ok2 := vnode.(*rtui.ElementVNode)
		if ok1 && ok2 {
			return currentElem.Tag() == newElem.Tag()
		}
	}

	// For text and fragments, type match is sufficient
	return true
}

// createChildFiber creates a new Fiber for a child VNode
func createChildFiber(returnFiber *Fiber, vnode rtui.VNode, lanes Lane) *Fiber {
	fiber := CreateFiberFromVNode(vnode)
	fiber.Return = returnFiber
	fiber.Lanes = lanes
	fiber.Props = vnode.Props()
	return fiber
}

// cloneExistingFiber clones an existing fiber with new VNode data
func cloneExistingFiber(returnFiber *Fiber, current *Fiber, vnode rtui.VNode) *Fiber {
	fiber := CloneFiber(current)
	fiber.Return = returnFiber
	fiber.VNode = vnode
	fiber.Props = vnode.Props()
	fiber.Lanes = LaneNoLane
	fiber.Flags = EffectNoEffect

	// Link to alternate
	fiber.Alternate = current
	if current.Alternate != nil {
		current.Alternate.Alternate = nil
	}

	return fiber
}

// =============================================================================
// Node Deletion
// =============================================================================

// markForDeletion marks a fiber and all its descendants for deletion
// This recursively traverses the subtree and sets the EffectDeletion flag
// The actual cleanup happens during the commit phase
//
// IMPORTANT: This only marks the fiber and its CHILD descendants, NOT siblings.
// Siblings are separate tree branches that should be processed independently.
func markForDeletion(fiber *Fiber) {
	if fiber == nil {
		return
	}

	// Debug logging
	if os.Getenv("TUI_DEBUG_DELETION") == "true" {
		fmt.Fprintf(os.Stderr, "[markForDeletion] Marking key=%q, current flags=%d\n",
			fiber.Key, fiber.Flags)
	}

	// Mark this fiber for deletion
	fiber.Flags |= EffectDeletion

	// Recursively mark all descendants (children only, not siblings)
	if child := fiber.Child; child != nil {
		markForDeletion(child)
	}

	// Trigger cleanup for component instances (e.g., useEffect cleanup)
	// This will be called during commit phase
	_ = fiber
}
