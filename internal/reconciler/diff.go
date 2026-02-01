package reconciler

// =============================================================================
// Fiber Reconciliation Algorithm
// =============================================================================
// reconcileChildren implements the Fiber tree diffing algorithm.
// This operates on Fiber nodes to reconcile old and new children.
// =============================================================================

import (
	"github.com/wwsheng009/mint/ui"
)

// reconcileChildren reconciles the current children with new children
// Returns the first child of the reconciled Fiber tree
func reconcileChildren(
	returnFiber *Fiber,
	currentFirstChild *Fiber,
	newChildren []ui.VNode,
	lanes Lane,
) *Fiber {
	// Validate keys for list children (React-style warning)
	if currentReconciler != nil && currentReconciler.keyValidator != nil {
		var parentVNode ui.VNode
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
func createAllNewChildren(returnFiber *Fiber, children []ui.VNode, lanes Lane) *Fiber {
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
// This is a simplified position-based reconciliation
func reconcileExistingChildren(
	returnFiber *Fiber,
	currentFirstChild *Fiber,
	newChildren []ui.VNode,
	lanes Lane,
) *Fiber {
	var firstChild *Fiber
	var previousChild *Fiber
	currentChild := currentFirstChild

	for _, childVNode := range newChildren {
		var child *Fiber

		// Try to match with current child
		if currentChild != nil && shouldUpdate(currentChild, childVNode) {
			// Reuse existing fiber
			child = cloneExistingFiber(returnFiber, currentChild, childVNode)
			currentChild = currentChild.Sibling
		} else {
			// Create new fiber
			child = createChildFiber(returnFiber, childVNode, lanes)
			// Remaining currentChildren will be deleted
			_ = currentChild // TODO: Schedule deletion in Phase 2
		}

		if firstChild == nil {
			firstChild = child
		} else {
			previousChild.Sibling = child
		}

		previousChild = child
	}

	// Delete remaining current children
	// TODO: Schedule deletion in Phase 2
	_ = currentChild

	return firstChild
}

// shouldUpdate checks if a current fiber can be updated with new VNode
// This follows React's reconciliation logic:
// 1. Key is primary - different keys mean different elements
// 2. Type is secondary - same key but different type means replace
func shouldUpdate(current *Fiber, vnode ui.VNode) bool {
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
	if current.Type == ui.VNodeComponent {
		currentComp, ok1 := current.VNode.(*ui.ComponentVNode)
		newComp, ok2 := vnode.(*ui.ComponentVNode)
		if ok1 && ok2 {
			// Compare component names since functions cannot be directly compared
			// Same key + same name = same component
			return currentComp.Name() == newComp.Name()
		}
	}

	// For elements, check if tag is the same
	if current.Type == ui.VNodeElement {
		currentElem, ok1 := current.VNode.(*ui.ElementVNode)
		newElem, ok2 := vnode.(*ui.ElementVNode)
		if ok1 && ok2 {
			return currentElem.Tag() == newElem.Tag()
		}
	}

	// For text and fragments, type match is sufficient
	return true
}

// createChildFiber creates a new Fiber for a child VNode
func createChildFiber(returnFiber *Fiber, vnode ui.VNode, lanes Lane) *Fiber {
	fiber := CreateFiberFromVNode(vnode)
	fiber.Return = returnFiber
	fiber.Lanes = lanes
	fiber.Props = vnode.Props()
	return fiber
}

// cloneExistingFiber clones an existing fiber with new VNode data
func cloneExistingFiber(returnFiber *Fiber, current *Fiber, vnode ui.VNode) *Fiber {
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
