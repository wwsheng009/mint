package reconciler

// =============================================================================
// Fiber Reconciliation Algorithm
// =============================================================================
// reconcileChildren implements the Fiber tree diffing algorithm.
// This operates on Fiber nodes to reconcile old and new children.
// =============================================================================

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/internal/log"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Path Generator (Global Instance)
// =============================================================================
// Global path generator instance for automatic key generation
// Will be initialized by the reconciler
var pathGenerator *PathGenerator

// =============================================================================
// Path Helper Functions
// =============================================================================

// extractPathSegment extracts the last segment from a full path
// For example: "/root/base[0]/vstack[0]/panel[1]" → "panel[1]"
func extractPathSegment(path string) string {
	if path == "" {
		return ""
	}

	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return ""
	}

	return parts[len(parts)-1]
}

// getTypeIDFromSegment extracts the type ID from a path segment
// For example: "button[2]" → "button"
func getTypeIDFromSegment(segment string) string {
	idx := strings.Index(segment, "[")
	if idx == -1 {
		return segment
	}
	return segment[:idx]
}


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

	// ✨ Track type counts for correct indexing
	typeCounts := make(map[string]int)

	if log.HitMapLogger.Enabled() {
		log.UILogger.Debug("[createAllNewChildren] Creating %d children for parent Key=%q, Tag=%q",
			len(children), returnFiber.Key, returnFiber.Tag)
	}

	for i, childVNode := range children {
		// Determine type index BEFORE creating the node
		var typeIndex int
		if pathGenerator != nil {
			typeID := pathGenerator.getTypeIdentifier(childVNode)
			typeIndex = typeCounts[typeID]
			typeCounts[typeID]++
		}

		// Create child with the pre-calculated index
		child := createChildFiberWithIndex(returnFiber, childVNode, lanes, i, typeIndex)

		if log.HitMapLogger.Enabled() {
			typeName := "UNKNOWN"
			switch child.Type {
			case rtui.VNodeComponent:
				typeName = "VNodeComponent"
			case rtui.VNodeText:
				typeName = "VNodeText"
			case rtui.VNodeElement:
				typeName = "VNodeElement"
			case rtui.VNodeFragment:
				typeName = "VNodeFragment"
			}
			log.UILogger.Debug("[createAllNewChildren] Created child %d: Type=%d(%s), Key=%q, Tag=%q, Path=%q",
				i, child.Type, typeName, child.Key, child.Tag, child.Path)
		}

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

	for i, childVNode := range newChildren {
		var child *Fiber

		// Try to match with current child or any of its siblings
		// This handles cases where a child later in the list matches
		matchedChild := findMatchingChild(currentChild, childVNode)

		if matchedChild != nil {
			// Found a match - reuse existing fiber
			child = cloneExistingFiber(returnFiber, matchedChild, childVNode, i)

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
			child = createChildFiber(returnFiber, childVNode, lanes, i)
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
// Implements the mixed key strategy:
// 1. User-provided key (highest priority)
// 2. Dynamic list → require key (panic if missing)
// 3. Static UI → auto-generate path key
func createChildFiber(returnFiber *Fiber, vnode rtui.VNode, lanes Lane, siblingIndex int) *Fiber {
	// Delegate to createChildFiberWithIndex with typeIndex=-1 (auto-calculate)
	return createChildFiberWithIndex(returnFiber, vnode, lanes, siblingIndex, -1)
}

// createChildFiberWithIndex creates a new Fiber with a pre-calculated type index
// Used by createAllNewChildren where type index is tracked externally
// If typeIndex is -1, it will be auto-calculated from parent's children
func createChildFiberWithIndex(returnFiber *Fiber, vnode rtui.VNode, lanes Lane, siblingIndex int, typeIndex int) *Fiber {
	fiber := CreateFiberFromVNode(vnode)
	fiber.Return = returnFiber
	fiber.Lanes = lanes
	fiber.Props = vnode.Props()
	fiber.SiblingIndex = siblingIndex

	// ✨ Mixed Key Strategy
	userKey := vnode.Key()

	// ✨ Special case: If parent is the root ComponentVNode (Key="root"),
	// this child is the actual app content and should get a layer-based path
	// FIX: Also check if vnode itself is a layer root node (modal, overlay, tooltip, inspector)
	// This ensures layer nodes get layer-based paths even when they're nested
	isRootChild := returnFiber != nil && returnFiber.Key == "root" && returnFiber.Path == "/root"
	isLayerNode := vnode.GetLayer() != rtui.LayerBase && vnode.GetLayer().IsValid()
	useLayerBasedPath := isRootChild || isLayerNode
	
	if !isLayerNode{
		fmt.Printf("layer s%\n", vnode.GetLayer().String())
	}

	if userKey != "" {
		// Priority 1: User provided a key
		// Use user key for reconciliation, but generate full path with type for Inspector
		fiber.Key = userKey

		// Generate path with type, then append user key for unique identification
		if pathGenerator == nil {
			pathGenerator = NewPathGenerator()
		}
		// Generate base path with type
		var typePath string
		if useLayerBasedPath {
			// Layer root nodes get layer-based path (e.g., /root/modal[0])
			// This is for the modal/tooltip/overlay itself, NOT its children
			typePath = pathGenerator.generateRootPath(vnode)
		} else {
			typePath = pathGenerator.GeneratePath(returnFiber, vnode, siblingIndex)
		}
		// Append user key to make it unique (e.g., .../hstack[0]/key[btn-event])
		fiber.Path = typePath + "/key[" + userKey + "]"
		// ✨ Sync full path to VNode so Inspector can access it
		fiber.Key = fiber.Path
		vnode.SetKey(fiber.Path)
	} else if isDynamicList(returnFiber) {
		// Priority 2: Dynamic list → require key (panic if missing)
		requireKeyPanic(returnFiber, vnode, siblingIndex)
	} else {
		// Priority 3: Static UI → auto-generate path key
		if pathGenerator == nil {
			pathGenerator = NewPathGenerator()
		}
		// Use provided typeIndex if available, otherwise auto-calculate
		if useLayerBasedPath {
			// Layer root nodes get layer-based path (e.g., /root/modal[0])
			// This is for the modal/tooltip/overlay itself, NOT its children
			fiber.Path = pathGenerator.generateRootPath(vnode)
		} else if typeIndex >= 0 {
			fiber.Path = pathGenerator.GeneratePathWithIndex(returnFiber, vnode, siblingIndex, typeIndex)
		} else {
			fiber.Path = pathGenerator.GeneratePath(returnFiber, vnode, siblingIndex)
		}
		fiber.Key = fiber.Path
		vnode.SetKey(fiber.Path)
	}

	// Extract path segment (last part of path)
	fiber.PathSegment = extractPathSegment(fiber.Path)

	return fiber
}

// cloneExistingFiber clones an existing fiber with new VNode data
func cloneExistingFiber(returnFiber *Fiber, current *Fiber, vnode rtui.VNode, siblingIndex int) *Fiber {
	fiber := CloneFiber(current)
	fiber.Return = returnFiber
	fiber.VNode = vnode
	fiber.Props = vnode.Props()
	fiber.Lanes = LaneNoLane
	fiber.Flags = EffectNoEffect

	// ✨ Keep path and key for Instance reuse
	userKey := vnode.Key()
	if userKey != "" && userKey != current.Key {
		// User changed the key, regenerate path with user key
		fiber.Key = userKey

		// Generate path with type, then append user key
		if pathGenerator == nil {
			pathGenerator = NewPathGenerator()
		}
		typePath := pathGenerator.GeneratePath(returnFiber, vnode, siblingIndex)
		fiber.Path = typePath + "/key[" + userKey + "]"
		// ✨ Sync full path to VNode so Inspector can access it
		vnode.SetKey(fiber.Path)
	} else {
		// Keep original path and key (critical for Instance reuse)
		fiber.Path = current.Path
		fiber.Key = current.Key
		// ✨ Sync path to VNode so Inspector and HitMap can access it
		// This ensures HitMap NodeID matches Instance key for event routing
		// FIX: Remove "/root/" prefix restriction to sync all paths
		if current.Path != "" {
			vnode.SetKey(current.Path)
		} else if current.Key != "" {
			// Fallback: use Key if Path is empty
			vnode.SetKey(current.Key)
		}
	}
	fiber.PathSegment = current.PathSegment
	fiber.SiblingIndex = siblingIndex

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
	if log.FiberLogger.Enabled() {
		log.UILogger.Debug("[markForDeletion] Marking key=%q, current flags=%d\n",
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
