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

// generateDebugPathForFiber generates a debug path for a fiber
// The path is ONLY for debugging (inspector, hit map, logging)
// It does NOT participate in diffing - DiffKey handles that
// Fiber-first version: uses Fiber fields instead of VNode
func generateDebugPathForFiber(fiber *Fiber, returnFiber *Fiber, siblingIndex int, typeIndex int) {
	if pathGenerator == nil {
		pathGenerator = NewPathGenerator()
	}

	userKey := fiber.Key

	// Check for layer-based path generation
	isRootChild := returnFiber != nil && returnFiber.Key == "root" && returnFiber.Path == "/root"
	isLayerNode := fiber.Layer != rtui.LayerBase && fiber.Layer.IsValid()
	useLayerBasedPath := isRootChild || isLayerNode

	var typePath string
	if useLayerBasedPath {
		typePath = pathGenerator.generateRootPathFromFiber(fiber)
	} else if typeIndex >= 0 {
		typePath = pathGenerator.GeneratePathWithIndexFromFiber(returnFiber, fiber, siblingIndex, typeIndex)
	} else {
		typePath = pathGenerator.GeneratePathFromFiber(returnFiber, fiber, siblingIndex)
	}

	// Append user key if present
	if userKey != "" {
		fiber.Path = typePath + "/key[" + userKey + "]"
	} else {
		fiber.Path = typePath
	}

	// Extract path segment from typePath (before user key append)
	fiber.PathSegment = extractPathSegment(typePath)
}

// generateDebugPathForFiberFromVNode generates path from VNode (used during initial creation)
func generateDebugPathForFiberFromVNode(fiber *Fiber, returnFiber *Fiber, vnode rtui.VNode, siblingIndex int, typeIndex int) {
	if pathGenerator == nil {
		pathGenerator = NewPathGenerator()
	}

	userKey := vnode.Key()

	// Check for layer-based path generation
	isRootChild := returnFiber != nil && returnFiber.Key == "root" && returnFiber.Path == "/root"
	isLayerNode := vnode.GetLayer() != rtui.LayerBase && vnode.GetLayer().IsValid()
	useLayerBasedPath := isRootChild || isLayerNode

	var typePath string
	if useLayerBasedPath {
		typePath = pathGenerator.generateRootPath(vnode)
	} else if typeIndex >= 0 {
		typePath = pathGenerator.GeneratePathWithIndex(returnFiber, vnode, siblingIndex, typeIndex)
	} else {
		typePath = pathGenerator.GeneratePath(returnFiber, vnode, siblingIndex)
	}

	// Append user key if present
	if userKey != "" {
		fiber.Path = typePath + "/key[" + userKey + "]"
	} else {
		fiber.Path = typePath
	}

	// Extract path segment from typePath
	fiber.PathSegment = extractPathSegment(typePath)
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
	// Note: In Fiber-first, we pass nil for parentVNode since we don't store VNode
	if currentReconciler != nil && currentReconciler.keyValidator != nil {
		currentReconciler.keyValidator.ValidateChildren(nil, newChildren)
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
		// ⚠️ Skip nil VNodes - they should not be in the Fiber tree
		if childVNode == nil {
			if log.HitMapLogger.Enabled() {
				log.UILogger.Debug("[createAllNewChildren] Skipping nil VNode at index %d", i)
			}
			continue
		}

		// Determine type index BEFORE creating the node
		var typeIndex int
		if pathGenerator != nil {
			typeID := pathGenerator.getTypeIdentifier(childVNode)
			typeIndex = typeCounts[typeID]
			typeCounts[typeID]++
		}

		// Create child with the pre-calculated index
		child := createChildFiberWithIndex(returnFiber, childVNode, lanes, i, typeIndex)
		if child == nil {
			// CreateFiber may return nil for nil VNode (though we check above)
			if log.HitMapLogger.Enabled() {
				log.UILogger.Debug("[createAllNewChildren] Skipping nil Fiber at index %d", i)
			}
			continue
		}

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
		matchedChild := findMatchingChild(currentChild, childVNode, i)

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
// newSiblingIndex is the position this VNode should have in the new tree (for correct DiffKey calculation)
func findMatchingChild(currentChild *Fiber, vnode rtui.VNode, newSiblingIndex int) *Fiber {
	for child := currentChild; child != nil; child = child.Sibling {
		if shouldUpdate(child, vnode, newSiblingIndex) {
			return child
		}
	}
	return nil
}

// =============================================================================
// DiffKey Normalization
// =============================================================================

// normalizeDiffKey generates a normalized DiffKey from VNode.
// This ensures consistent key generation between creation and comparison.
// Priority: user key > index fallback
func normalizeDiffKey(vnode rtui.VNode, siblingIndex int) string {
	userKey := vnode.Key()
	if userKey != "" {
		return userKey
	}
	return fmt.Sprintf("_idx_%d", siblingIndex)
}

// =============================================================================
// Fiber Comparison (Fiber-first)
// =============================================================================

// shouldUpdate checks if a current fiber can be updated with new VNode.
// ✨ Fiber-first: Creates a temporary newFiber for comparison, then compares Fiber vs Fiber.
// This ensures DiffKey normalization is consistent between creation and comparison.
//
// newSiblingIndex is the position the VNode should have in the new tree.
// This is CRITICAL for correct DiffKey comparison when children are reordered/removed.
//
// Per docs/fiber/diff_key_detail.md:
// - DiffKey is runtime concept (normalized)
// - VNode.key is declaration concept (raw)
// - Should compare Fiber.DiffKey vs Fiber.DiffKey, not Fiber.DiffKey vs VNode.key
func shouldUpdate(current *Fiber, vnode rtui.VNode, newSiblingIndex int) bool {
	if current == nil || vnode == nil {
		return false
	}

	// ✨ BUG FIX: Use newSiblingIndex (position in new tree) instead of current.SiblingIndex
	// This prevents incorrect matches when children are reordered or removed
	newDiffKey := normalizeDiffKey(vnode, newSiblingIndex)

	if current.DiffKey != newDiffKey {
		return false
	}

	if current.Type != vnode.Type() {
		return false
	}

	if current.Type == rtui.VNodeComponent {
		newComp, ok := vnode.(*rtui.ComponentVNode)
		if ok {
			return current.ComponentName == newComp.Name()
		}
		if tagger, ok := vnode.(interface{ Tag() string }); ok {
			return current.Tag == tagger.Tag()
		}
		return false
	}

	if current.Type == rtui.VNodeElement {
		if tagger, ok := vnode.(interface{ Tag() string }); ok {
			return current.Tag == tagger.Tag()
		}
		return false
	}

	return true
}

// shouldReuseFiber checks if two Fiber nodes can be reused.
// This is the pure Fiber-to-Fiber comparison (no VNode dependency).
// Use this when comparing existing Fibers during reconciliation.
func shouldReuseFiber(current *Fiber, newFiber *Fiber) bool {
	if current == nil || newFiber == nil {
		return false
	}

	if current.DiffKey != newFiber.DiffKey {
		return false
	}

	if current.Type != newFiber.Type {
		return false
	}

	if current.Type == rtui.VNodeComponent {
		return current.ComponentName == newFiber.ComponentName
	}

	if current.Type == rtui.VNodeElement {
		return current.Tag == newFiber.Tag
	}

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
//
// ✨ DiffKey normalization: Uses normalizeDiffKey for consistent key generation
// This ensures DiffKey matches between creation and shouldUpdate comparison
func createChildFiberWithIndex(returnFiber *Fiber, vnode rtui.VNode, lanes Lane, siblingIndex int, typeIndex int) *Fiber {
	fiber := CreateFiber(vnode)
	if fiber == nil {
		return nil
	}

	fiber.Return = returnFiber
	fiber.Lanes = lanes
	fiber.SiblingIndex = siblingIndex

	userKey := vnode.Key()
	if userKey == "" && isDynamicList(returnFiber) {
		requireKeyPanic(returnFiber, vnode, siblingIndex)
	}

	fiber.DiffKey = normalizeDiffKey(vnode, siblingIndex)
	fiber.Key = fiber.DiffKey

	generateDebugPathForFiberFromVNode(fiber, returnFiber, vnode, siblingIndex, typeIndex)

	return fiber
}

// cloneExistingFiber clones an existing fiber with new VNode data
// ✨ Optimized: Preserves DiffKey for stable diffing
// Path is only updated for debugging (inspector, hit map)
// ✨ Fiber-first: Extracts data from VNode without storing reference
func cloneExistingFiber(returnFiber *Fiber, current *Fiber, vnode rtui.VNode, siblingIndex int) *Fiber {
	fiber := CloneFiber(current)
	fiber.Return = returnFiber
	// Extract data from vnode instead of storing reference
	// IMPORTANT: For elements/fragments, ensure children are stored in Props
	props := vnode.Props()
	if props == nil {
		props = make(rtui.Props)
	}
	vnodeType := vnode.Type()
	children := vnode.Children()
	if vnodeType == rtui.VNodeElement || vnodeType == rtui.VNodeFragment {
		if len(children) > 0 {
			newProps := make(rtui.Props)
			for k, v := range props {
				newProps[k] = v
			}
			newProps["children"] = children
			props = newProps
		}
	}
	fiber.Props = props
	fiber.Style = vnode.Style()
	fiber.Layer = vnode.GetLayer()
	// Update tag if available
	if tagger, ok := vnode.(interface{ Tag() string }); ok {
		fiber.Tag = tagger.Tag()
	}
	// Update FocusableVNode from new VNode (Fiber-first)
	// if f, ok := vnode.(rtui.FocusableVNode); ok && f.IsFocusable() {
	// 	fiber.FocusableVNode = f
	// } else {
	// 	fiber.FocusableVNode = nil
	// }
	fiber.Lanes = LaneNoLane
	fiber.Flags = EffectNoEffect

	// ✨ CRITICAL: DiffKey is preserved by CloneFiber
	// The DiffKey remains stable across renders as long as vnode.Key() doesn't change
	// This ensures diffing works correctly

	// Path is only for debugging - update it for inspector/hit map
	generateDebugPathForFiber(fiber, returnFiber, siblingIndex, -1)

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
		log.UILogger.Debug("[markForDeletion] Marking key=%q, current flags=%d",
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
