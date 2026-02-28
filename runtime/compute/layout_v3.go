package compute

import (
	"fmt"

	"github.com/wwsheng009/mint/internal/render"
	"github.com/wwsheng009/mint/internal/reconciler"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/event"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

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
