// Package render provides layout engine adapters for the rendering pipeline.
package render

import (
	"fmt"

	"github.com/wwsheng009/mint/internal/reconciler"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/event"
	"github.com/wwsheng009/mint/runtime/layer"
	"github.com/wwsheng009/mint/runtime/layout"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// LayoutResult - Unified layout result interface
// =============================================================================

// LayoutResult represents the output of a layout computation.
type LayoutResult interface {
	// GetRootBox returns the root layout box for painting
	GetRootBox() PaintableBox

	// GetHitMap returns the hit map for event routing
	GetHitMap() *event.HitMap

	// GetRenderPlanes returns layered boxes for multi-layer rendering
	GetRenderPlanes() *layer.RenderPlanes
}

// PaintableBox represents a box that can be painted.
type PaintableBox interface {
	// GetBounds returns the box bounds (x, y, width, height)
	GetBounds() (x, y, width, height int)

	// GetChildren returns child boxes
	GetChildren() []PaintableBox
}

// CacheStats represents cache statistics.
type CacheStats struct {
	Hits   int
	Misses int
	Size   int
}

// =============================================================================
// NewLayoutEngineAdapter - Adapts layout.Engine to LayoutEngine interface
// =============================================================================

// NewLayoutEngineAdapter wraps layout.Engine to implement the layout interface.
type NewLayoutEngineAdapter struct {
	engine *layout.Engine
}

// NewNewLayoutEngineAdapter creates a new adapter for the layout engine.
func NewNewLayoutEngineAdapter() *NewLayoutEngineAdapter {
	return &NewLayoutEngineAdapter{
		engine: layout.NewEngine(),
	}
}

// Layout performs layout using the new layout engine.
// Fiber-first: Uses Fiber data when available, falls back to VNode.
func (a *NewLayoutEngineAdapter) Layout(vnode rtui.VNode, fiber *reconciler.Fiber, constraints runtime.BoxConstraints) (LayoutResult, error) {
	// Fiber-first: Prefer Fiber over VNode
	var node layout.Node
	if fiber != nil {
		// Use Fiber-only adapter (no VNode dependency)
		node = NewFiberToNodeAdapterPure(fiber)
	} else if vnode != nil {
		// Fallback to VNode adapter for legacy support
		node = NewVNodeToNodeAdapter(vnode)
	} else {
		return nil, fmt.Errorf("both fiber and vnode are nil")
	}

	// Convert constraints
	layoutConstraints := layout.Constraints{
		MinWidth:  constraints.MinWidth,
		MaxWidth:  constraints.MaxWidth,
		MinHeight: constraints.MinHeight,
		MaxHeight: constraints.MaxHeight,
	}

	// Perform layout
	result := a.engine.Layout(node, layoutConstraints)

	return &newLayoutResultAdapter{result: result}, nil
}

// LayoutFiber performs layout on a Fiber tree using the new layout engine (Fiber-first).
func (a *NewLayoutEngineAdapter) LayoutFiber(fiber *reconciler.Fiber, constraints runtime.BoxConstraints) (LayoutResult, error) {
	if fiber == nil {
		return nil, fmt.Errorf("fiber is nil")
	}

	// Use Fiber-only adapter (no VNode dependency)
	node := NewFiberToNodeAdapterPure(fiber)

	// Convert constraints
	layoutConstraints := layout.Constraints{
		MinWidth:  constraints.MinWidth,
		MaxWidth:  constraints.MaxWidth,
		MinHeight: constraints.MinHeight,
		MaxHeight: constraints.MaxHeight,
	}

	// Perform layout
	result := a.engine.Layout(node, layoutConstraints)

	return &newLayoutResultAdapter{result: result, fiberRoot: fiber}, nil
}

// GetStats returns cache statistics.
func (a *NewLayoutEngineAdapter) GetStats() CacheStats {
	stats := a.engine.GetStats()
	return CacheStats{
		Hits:   int(stats.CacheHits),
		Misses: int(stats.CacheMisses),
		Size:   0, // Not tracked in new engine
	}
}

// ClearCache clears the layout cache.
func (a *NewLayoutEngineAdapter) ClearCache() {
	a.engine.Invalidate()
}

// SetDebug enables/disables debug output.
func (a *NewLayoutEngineAdapter) SetDebug(debug bool) {
	// Layout engine doesn't have debug mode yet
}

// GetEngine returns the underlying layout engine.
func (a *NewLayoutEngineAdapter) GetEngine() *layout.Engine {
	return a.engine
}

// =============================================================================
// newLayoutResultAdapter - Adapts layout.LayoutResult to LayoutResult
// =============================================================================

// newLayoutResultAdapter adapts layout.LayoutResult to LayoutResult.
type newLayoutResultAdapter struct {
	result    *layout.LayoutResult
	fiberRoot *rtui.Fiber // For HitMap enrichment with TargetFiber
}

// GetRootBox returns the root box for painting.
func (a *newLayoutResultAdapter) GetRootBox() PaintableBox {
	if a.result == nil || a.result.Root == nil {
		return nil
	}
	return &layoutBoxAdapter{LayoutBox: a.result.Root}
}

// GetHitMap returns the hit map (converted from layout.HitMap).
func (a *newLayoutResultAdapter) GetHitMap() *event.HitMap {
	if a.result == nil || a.result.HitMap == nil {
		return nil
	}
	// Convert layout.HitMap to event.HitMap with TargetFiber enrichment
	return convertLayoutHitMap(a.result.HitMap, a.fiberRoot)
}

// GetRenderPlanes returns render planes built from layout.LayoutBox tree.
// NOTE: Currently returns nil because layout.LayoutBox doesn't contain VNode references
// needed for painting. The RenderLayers will fallback to Render() method which works correctly.
func (a *newLayoutResultAdapter) GetRenderPlanes() *layer.RenderPlanes {
	// Return nil to force fallback to Render() method
	// The Render() method handles Modal centering via applyLayerTransforms
	// TODO: In future, we could build proper RenderPlanes if we store VNode references in LayoutBox
	return nil
}

// GetLayoutResult returns the underlying layout.LayoutResult.
// This is needed for applying layer transforms (Modal centering).
func (a *newLayoutResultAdapter) GetLayoutResult() *layout.LayoutResult {
	return a.result
}

// =============================================================================
// layoutBoxAdapter - Adapts layout.LayoutBox to PaintableBox
// =============================================================================

// layoutBoxAdapter adapts layout.LayoutBox to PaintableBox.
type layoutBoxAdapter struct {
	*layout.LayoutBox
}

// GetBounds returns the box bounds.
func (a *layoutBoxAdapter) GetBounds() (x, y, width, height int) {
	if a.LayoutBox == nil {
		return 0, 0, 0, 0
	}
	return a.LayoutBox.X, a.LayoutBox.Y, a.LayoutBox.Width, a.LayoutBox.Height
}

// GetChildren returns child boxes.
func (a *layoutBoxAdapter) GetChildren() []PaintableBox {
	if a.LayoutBox == nil {
		return nil
	}
	children := make([]PaintableBox, len(a.LayoutBox.Children))
	for i, child := range a.LayoutBox.Children {
		children[i] = &layoutBoxAdapter{LayoutBox: child}
	}
	return children
}

// convertLayoutHitMap converts layout.HitMap to event.HitMap.
// This converts from layout.HitMap (NodeID=string, ZIndex=int, Rect) to
// event.HitMap (NodeID=uint64, ZOrder=int, LocalXY function, TargetFiber).
//
// The fiberRoot parameter is used to look up Fiber nodes by NodeID and set TargetFiber.
func convertLayoutHitMap(hm *layout.HitMap, fiberRoot *rtui.Fiber) *event.HitMap {
	if hm == nil || hm.Size() == 0 {
		return nil
	}

	allEntries := hm.GetAll()
	if len(allEntries) == 0 {
		return nil
	}

	// Build entries for event.HitMap
	entries := make([]event.HitMapEntryInternal, 0, len(allEntries))
	for _, layoutEntry := range allEntries {
		// Convert NodeID string to uint64
		// Since adapter_fiber.go now returns NodeID as plain string (e.g., "123"),
		// StringToNodeID("123") will return 123 as uint64
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
