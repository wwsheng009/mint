package compute

import (
	"fmt"

	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/internal/reconciler"
	"github.com/wwsheng009/mint/internal/render"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/event"
	"github.com/wwsheng009/mint/runtime/layout"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Engine performs constraint-driven layout calculation
// It separates layout (position calculation) from paint (rendering)
type Engine struct {
	cache        *LayoutCache
	dirtyTracker *DirtyTracker
	debug        bool
	flexCache    map[string]*FlexDistributionInfo // Cache for flex distribution per parent
	traceDepth   int                              // Current depth for layout tracing
	validator    *BoundsValidator                 // Validates bounds consistency
}

// NewEngine creates a new layout engine
func NewEngine() *Engine {
	return &Engine{
		cache:        NewLayoutCache(),
		dirtyTracker: NewDirtyTracker(),
		debug:        log.LayoutLogger.Enabled(),
		validator:    NewBoundsValidator(),
	}
}

// SetDebug enables/disables debug output
func (e *Engine) SetDebug(debug bool) {
	e.debug = debug
}

// Layout performs layout calculation on a VNode tree
// Returns a ComputedLayout containing computed positions for all nodes
// AND a HitMap built from the final ComputedBox positions (including layer transforms)
//
// Parameters:
//
//	vnode: The VNode tree to layout
//	fiber: Optional Fiber node for passing NodeID to ComputedBox (Phase 3: Identity Refactoring)
//	       When provided, Fiber.NodeID is passed to ComputedBox for stable identity
//	       When nil, NodeID will be 0 (backward compatible with non-Fiber mode)
//	constraints: Box constraints for layout
func (e *Engine) Layout(vnode VNode, fiber *reconciler.Fiber, constraints runtime.BoxConstraints) (*ComputedLayout, error) {
	// Delegate to LayoutV3 (new layout engine)
	return e.LayoutV3(vnode, fiber, constraints)
}

// LayoutFiber lays out the entire Fiber tree
// This method now delegates to LayoutV3 which uses the new runtime/layout engine (fiber-first)
//
// Parameters:
//
//	root: The root Fiber node of the tree to layout
//	constraints: Box constraints for the entire tree layout
//
// Returns:
//
//	*ComputedLayout containing the root ComputedBox and HitMap
//	error if layout fails
//
// Note: This method now delegates to LayoutV3 (new layout engine with fiber-first support)
func (e *Engine) LayoutFiber(root *rtui.Fiber, constraints runtime.BoxConstraints) (*ComputedLayout, error) {
	// Delegate to LayoutV3 (new layout engine with fiber-first)
	// rtui.Fiber and reconciler.Fiber are the same type (type alias)
	return e.LayoutV3(nil, root, constraints)
}
// LayoutV3 performs layout calculation using the new layout engine (V3)
// Fiber-first: Uses the new layout.Engine from runtime/layout package
//
// This method bridges the legacy compute.Engine API with the new layout system.
// It converts inputs and outputs between the two systems while maintaining backward compatibility.
//
// Parameters:
//   - vnode: The VNode tree to layout (optional, used as fallback if fiber is nil)
//   - fiber: The Fiber tree for layout (Fiber-first: preferred over vnode)
//   - constraints: Box constraints for layout
//
// Returns:
//   - *ComputedLayout: Computed layout result with positions for all nodes
//   - error: Error if layout calculation fails
func (e *Engine) LayoutV3(vnode VNode, fiber *reconciler.Fiber, constraints runtime.BoxConstraints) (*ComputedLayout, error) {
	// Import render package adapters
	return layoutV3Impl(vnode, fiber, constraints)
}

// layoutV3Impl implements the LayoutV3 method using the new layout engine (V3)
//
// This function bridges the legacy compute.Engine API with the new layout system
// by calling render.LayoutV3 which performs the actual layout computation.
// The result is then converted to ComputedLayout format.
func layoutV3Impl(vnode VNode, fiber *reconciler.Fiber, constraints runtime.BoxConstraints) (*ComputedLayout, error) {
	// Validate inputs
	if vnode == nil && fiber == nil {
		return nil, fmt.Errorf("cannot layout: both vnode and fiber are nil")
	}

	// Call render.LayoutV3 to get layout result
	// Note: This returns interface{} to avoid circular import issues
	result, err := render.LayoutV3(vnode, fiber, constraints)
	if err != nil {
		return nil, err
	}

	// Type assert to *layout.LayoutResult
	layoutResult, ok := result.(*layout.LayoutResult)
	if !ok {
		return nil, fmt.Errorf("unexpected layout result type: %T", result)
	}

	// Convert to ComputedLayout (simplified, no Fiber references)
	computedLayout := convertLayoutResultToComputedBox(layoutResult, vnode, fiber)

	return computedLayout, nil
}

// convertLayoutResultToComputedBox converts layout.LayoutResult to compute.ComputedLayout
// This bridges the new layout engine (V3) with the legacy compute.Engine output format.
//
// Simplified approach: only preserves essential layout data (Box, NodeID, Layer, Children).
// VNode and ChildFiber references are removed to reduce coupling.
func convertLayoutResultToComputedBox(
	layoutResult *layout.LayoutResult,
	vnode VNode,
	fiber *rtui.Fiber,
) *ComputedLayout {
	if layoutResult == nil || layoutResult.Root == nil {
		return NewComputedLayout(nil)
	}

	// Convert LayoutBox tree to ComputedBox tree
	rootBox := convertLayoutBoxToComputedBox(layoutResult.Root, nil)

	// Create ComputedLayout
	computedLayout := NewComputedLayout(rootBox)

	// Convert HitMap (required for event routing)
	if layoutResult.HitMap != nil {
		computedLayout.HitMap = buildHitMapFromLayoutResult(layoutResult.HitMap, fiber)
	}

	return computedLayout
}

// convertLayoutBoxToComputedBox recursively converts layout.LayoutBox to compute.ComputedBox
// Simplified version: only preserves essential layout data (Box, NodeID, Layer, Children)
// VNode and ChildFiber references are removed as part of the migration to pure layout engine
func convertLayoutBoxToComputedBox(
	lbox *layout.LayoutBox,
	parent *ComputedBox,
) *ComputedBox {
	if lbox == nil {
		return nil
	}

	// Convert NodeID string to uint64
	nodeID := event.StringToNodeID(lbox.ID)

	// Create ComputedBox with minimal fields
	// VNode and ChildFiber are intentionally left nil to reduce dependencies
	cb := &ComputedBox{
		VNode:       nil,  // Deprecated: use fiberMap to look up Fiber by NodeID
		ChildFiber:  nil,  // Deprecated: stored only for legacy compatibility
		Parent:      parent,
		LayoutDirty: false,
		Box: runtime.Box{
			X:      lbox.X,
			Y:      lbox.Y,
			Width:  lbox.Width,
			Height: lbox.Height,
		},
		Children:     make([]*ComputedBox, 0, len(lbox.Children)),
		Layer:        rtui.Layer(lbox.Layer),
		NodeID:       nodeID,
		DiffKey:      "",     // Deprecated: not needed for pure layout results
		NaturalWidth: lbox.Width, // TODO: Extract natural width from layout metadata if available
	}

	// Recursively convert children
	for _, childLBox := range lbox.Children {
		childBox := convertLayoutBoxToComputedBox(childLBox, cb)
		if childBox != nil {
			cb.Children = append(cb.Children, childBox)
		}
	}

	return cb
}

// buildHitMapFromLayoutResult converts layout.HitMap to event.HitMap
// This bridges the new layout engine's hit map format to the event routing system
func buildHitMapFromLayoutResult(layoutHitMap *layout.HitMap, fiberRoot *reconciler.Fiber) *event.HitMap {
	if layoutHitMap == nil || layoutHitMap.Size() == 0 {
		return nil
	}

	allEntries := layoutHitMap.GetAll()
	if len(allEntries) == 0 {
		return nil
	}

	// Build entries for event.HitMap
	entries := make([]event.HitMapEntryInternal, 0, len(allEntries))
	for _, layoutEntry := range allEntries {
		// Convert NodeID string to uint64
		nodeID := event.StringToNodeID(layoutEntry.NodeID)

		// Look up the target Fiber by NodeID
		var targetFiber *rtui.Fiber
		if fiberRoot != nil {
			targetFiber = rtui.FindFiberByID(fiberRoot, nodeID)
		}

		entries = append(entries, event.HitMapEntryInternal{
			NodeID: nodeID,
			Bounds: layoutEntry.Rect,
			LocalXY: func(screenX, screenY int) (int, int) {
				return screenX - layoutEntry.Rect.X, screenY - layoutEntry.Rect.Y
			},
			ZOrder:      layoutEntry.ZIndex,
			TargetFiber: targetFiber, // Set TargetFiber for action routing
		})
	}

	return event.BuildHitMapFromEntries(entries)
}

// LayoutEngineV3 represents the new layout engine wrapper for compatibility
type LayoutEngineV3 struct {
	Engine *layout.Engine
}

// NewLayoutEngineV3 creates a new wrapper for the V3 layout engine
func NewLayoutEngineV3() *LayoutEngineV3 {
	return &LayoutEngineV3{
		Engine: layout.NewEngine(),
	}
}

// Layout performs layout using the V3 engine
func (e *LayoutEngineV3) Layout(vnode VNode, fiber *reconciler.Fiber, constraints runtime.BoxConstraints) (*ComputedLayout, error) {
	return layoutV3Impl(vnode, fiber, constraints)
}
