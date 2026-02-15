// Package compute provides constraint-driven layout engine for TUI components
package compute

import (
	"fmt"
	"sort"
	"strings"

	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/internal/reconciler"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/event"
	runtimelayout "github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
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

func (e *Engine) getTraceDepth() int {
	return e.traceDepth
}

func (e *Engine) incrementTraceDepth() {
	e.traceDepth++
}

func (e *Engine) decrementTraceDepth() {
	e.traceDepth--
	if e.traceDepth < 0 {
		e.traceDepth = 0
	}
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
//   vnode: The VNode tree to layout
//   fiber: Optional Fiber node for passing NodeID to ComputedBox (Phase 3: Identity Refactoring)
//          When provided, Fiber.NodeID is passed to ComputedBox for stable identity
//          When nil, NodeID will be 0 (backward compatible with non-Fiber mode)
//   constraints: Box constraints for layout
func (e *Engine) Layout(vnode VNode, fiber *reconciler.Fiber, constraints runtime.BoxConstraints) (*ComputedLayout, error) {
	if vnode == nil {
		return nil, fmt.Errorf("cannot layout nil VNode")
	}

	// Reset flex cache for each layout pass
	e.flexCache = make(map[string]*FlexDistributionInfo)

	// Build layout tree and measure
	// Pass Fiber to buildComputedBox so it can extract NodeID
	root := e.buildComputedBox(vnode, fiber, nil, constraints)
	if root == nil {
		layout := NewComputedLayout(nil)
		layout.HitMap = event.NewHitMap()
		return layout, nil
	}

	// Calculate positions (second pass)
	e.calculatePositions(root, 0, 0)

	// Clear dirty flags after layout
	root.ClearDirty()

	// Build HitMap directly from ComputedBox tree
	// This captures the FINAL positions after all transforms (including layer centering)
	hitMap := e.buildHitMapFromComputedBoxes(root)

	layout := NewComputedLayout(root)
	layout.HitMap = hitMap

	log.HitMapLogger.Debug("[Engine.Layout] Built HitMap with %d entries", hitMap.Size())

	// Validate bounds consistency (only in debug mode)
	if err := e.validator.ValidateLayout(layout); err != nil {
		// Log validation error but don't fail the layout
		// This helps catch bugs during development
		log.RenderLogger.Debug("[Engine.Layout] ⚠️ Validation warning: %v", err)
	}

	return layout, nil
}

// =============================================================================
// Layout Box Building (First Pass: Measurement)
// =============================================================================

// buildComputedBoxWithSize creates a computed box for a VNode with its children,
// using a pre-measured size if provided (to avoid re-measurement in single-pass layout).
// This is the first pass - measuring sizes
//
// Caching strategy: Only cache leaf nodes (nodes without vnode children).
// This avoids the complexity of caching entire subtrees while still
// providing performance benefits for simple nodes like text.
//
// Parameters:
//   vnode: The VNode to build ComputedBox for
//   fiber: Optional Fiber node for passing NodeID (Phase 3: Identity Refactoring)
//          When provided, Fiber.NodeID is set in ComputedBox for stable identity
//   parent: Parent ComputedBox (for tree structure)
//   constraints: Box constraints for layout
//   preMeasuredSize: Optional pre-measured size to avoid re-measurement
func (e *Engine) buildComputedBoxWithSize(vnode VNode, fiber *reconciler.Fiber, parent *ComputedBox, constraints runtime.BoxConstraints, preMeasuredSize *runtime.Size) *ComputedBox {
	if vnode == nil {
		return nil
	}

	box := &ComputedBox{
		VNode:        vnode,
		Parent:       parent,
		Box:          runtime.Box{X: 0, Y: 0, Width: 0, Height: 0},
		NaturalWidth:  0, // Will be measured below
		NodeID:       0, // Will be set below
		ChildFiber:   fiber, // Set Fiber if provided (for NodeID propagation)
	}

	// Phase 3: Set NodeID from Fiber for stable identity
	// This provides runtime identity independent of VNode keys and paths
	// See: docs/render/fiber/IDENTITY_REFACTORING_PLAN.md
	// Priority: Use own Fiber.NodeID first, then inherit from parent
	if fiber != nil {
		box.NodeID = fiber.NodeID
		box.ChildFiber = fiber // Set ChildFiber for NodeID propagation to children
		if e.debug {
			log.EngineLogger.Debug("[buildComputedBoxWithSize] Set NodeID=%d from Fiber (type=%s key=%s)",
				box.NodeID, vnode.Type().String(), vnode.Key())
		}
	} else if parent != nil && parent.NodeID != 0 {
		box.NodeID = parent.NodeID
		if e.debug {
			log.EngineLogger.Debug("[buildComputedBoxWithSize] Inherited NodeID=%d from parent (type=%s key=%s)",
				box.NodeID, vnode.Type().String(), vnode.Key())
		}
	} else {
		if e.debug {
			log.EngineLogger.Debug("[buildComputedBoxWithSize] ⚠️  NodeID=0 (no fiber, no parent with NodeID)")
		}
	}

	// Measure natural width (unconstrained) for alignment calculations
	// This is needed for proper centering when element is stretched by flex
	if measurable, ok := vnode.(interface {
		Measure(runtime.BoxConstraints) runtime.Size
	}); ok {
		naturalSize := measurable.Measure(runtime.BoxConstraints{
			MinWidth:  0,
			MaxWidth:  runtime.Infinity,
			MinHeight: 0,
			MaxHeight: runtime.Infinity,
		})
		box.NaturalWidth = naturalSize.Width
	}

	// Get vnode children to determine if this is a leaf node
	vnodeChildren := vnode.Children()
	isLeaf := len(vnodeChildren) == 0

	// Check cache (only for leaf nodes or nodes with explicit keys)
	cacheKey := e.getCacheKey(vnode, constraints)
	if isLeaf || vnode.Key() != "" {
		if cached, ok := e.cache.Get(cacheKey); ok && !e.dirtyTracker.NeedLayoutBox(box) {
			// Cache hit - use cached size
			box.Box = cached.Box
			if e.debug {
				log.EngineLogger.Debug("[Layout.CacheHit] %s (key=%s): %v\n",
					vnode.Type().String(), vnode.Key(), cached.Box)
			}
			// For leaf nodes, we're done
			if isLeaf {
				box.Children = nil // Leaf nodes have no children
				return box
			}
			// For keyed nodes with children, we still need to build children
			// but can use the cached size for the parent
		}
	}

	// Try single-pass measurement if LayoutMeasurer is implemented
	measurement := e.TryMeasureLayout(vnode, constraints)
	if e.debug {
		tag := "none"
		if tagger, ok := vnode.(interface{ Tag() string }); ok {
			tag = tagger.Tag()
		}
		log.LayoutLogger.Debug("[buildComputedBox] tag=%s, childConstraints=%d, using single-pass=%v",
			tag, len(measurement.ChildConstraints), len(measurement.ChildConstraints) > 0)
	}
	// Check if we got a valid measurement (has child constraints)
	if len(measurement.ChildConstraints) > 0 {
		// Use the new single-pass approach
		box.Box.Width = measurement.Size.Width
		box.Box.Height = measurement.Size.Height

		// Build children using pre-calculated constraints
		// Note: We don't pass pre-measured sizes because:
		// 1. LayoutNodes will use their own MeasureLayout with the correct constraints
		// 2. Leaf nodes will be measured with the correct constraints
		box.Children = make([]*ComputedBox, 0, len(vnodeChildren))
		for i, child := range vnode.Children() {
			childConstraints := measurement.ChildConstraints[i]

			// Set childFiber before building child box
			// Find the Fiber node that corresponds to this child VNode
			var childFiber *rtui.Fiber
			if fiber != nil {
				// Debug: Check if fiber.Child exists
				if e.debug && fiber.Child == nil {
					log.EngineLogger.Debug("[buildComputedBoxWithSize] ⚠️  fiber.Child is NIL! fiber.NodeID=%d Type=%s",
						fiber.NodeID, fiber.Type.String())
				}

				// IMPORTANT: Match by DiffKey (vnode.Key()), not by index
				// This is critical because the Fiber tree maintains stable NodeIDs across renders
				// while the VNode order may change (e.g., reordering, filtering)
				childKey := child.Key()
				if e.debug {
					log.EngineLogger.Debug("[buildComputedBoxWithSize] Looking for child[%d] with key=%q",
						i, childKey)
				}

				// Strategy:
				// 1. If child has an explicit key, match by DiffKey in the Fiber tree
				// 2. If child has no key, fall back to index matching (for static UI)
				if childKey != "" {
					// Match by DiffKey - this ensures we find the correct Fiber node
					// regardless of VNode order changes
					for f := fiber.Child; f != nil; f = f.Sibling {
						if f.DiffKey == childKey {
							childFiber = f
							if e.debug {
								log.EngineLogger.Debug("[buildComputedBoxWithSize] ✅ Matched child %d by DiffKey: childKey=%q, childFiber.NodeID=%d Type=%s DiffKey=%q",
									i, childKey, f.NodeID, f.Type.String(), f.DiffKey)
							}
							break
						}
					}

					if e.debug && childFiber == nil {
						log.EngineLogger.Debug("[buildComputedBoxWithSize] ⚠️  No Fiber with DiffKey=%q found (expected for newly inserted nodes)",
							childKey)
					}
				} else {
					// No key: fall back to index matching (for static UI components)
					// This is less stable but is the fallback for components without keys
					childIndex := 0
					for f := fiber.Child; f != nil; f = f.Sibling {
						if childIndex == i {
							childFiber = f
							if e.debug {
								log.EngineLogger.Debug("[buildComputedBoxWithSize] ✅ Matched child %d by index (no key): childFiber.NodeID=%d Type=%s",
									i, f.NodeID, f.Type.String())
							}
							break
						}
						childIndex++
					}

					if e.debug && childFiber == nil {
						log.EngineLogger.Debug("[buildComputedBoxWithSize] ⚠️  No matching Fiber for index %d (child has no key)",
							i)
					}
				}
			}

			childBox := e.buildComputedBox(child, childFiber, box, childConstraints)
			if childBox != nil {
				box.Children = append(box.Children, childBox)

				// Debug: Verify childFiber was passed
				if e.debug && childBox.ChildFiber != nil {
					log.EngineLogger.Debug("[buildComputedBoxWithSize] childBox[%d] ChildFiber.NodeID=%d (parent Fiber.NodeID=%d)",
						i, childBox.ChildFiber.NodeID, box.ChildFiber.NodeID)
				}
			}
		}

		// Cache the result for leaf nodes
		if isLeaf {
			e.cache.Set(cacheKey, LayoutCacheEntry{
				Box:     box.Box,
				Size:    runtime.Size{Width: box.Box.Width, Height: box.Box.Height},
				Hash:    e.computeHash(vnode, constraints),
				IsLeaf:  true,
				VNodeID: vnode.Key(),
			})
			if e.debug {
				log.EngineLogger.Debug("[Layout.CacheSet] %s: %v\n",
					vnode.Type().String(), box.Box)
			}
		}

		return box
	}

	// FALLBACK: Use the legacy two-pass approach
	size := e.measureVNode(vnode, constraints)
	box.Box.Width = size.Width
	box.Box.Height = size.Height

	// REFRESH children: Measure() might have updated the children list (e.g. TreeView virtual scrolling)
	vnodeChildren = vnode.Children()

	// Build children layout boxes
	box.Children = make([]*ComputedBox, 0, len(vnodeChildren))

	for i, child := range vnodeChildren {
		// Calculate child constraints based on layout type
		childConstraints := e.getChildConstraints(vnode, child, constraints, size)

		// Set childFiber before building child box
		// Find the Fiber node that corresponds to this child VNode
		var childFiber *rtui.Fiber
		if fiber != nil {
			// IMPORTANT: Match by DiffKey (vnode.Key()), not by index
			// See comment in single-pass path above for rationale
			childKey := child.Key()
			if e.debug {
				log.EngineLogger.Debug("[buildComputedBoxWithSize.FALLBACK] Looking for child[%d] with key=%q",
					i, childKey)
			}

			if childKey != "" {
				// Match by DiffKey - ensures correct Fiber node regardless of VNode order
				for f := fiber.Child; f != nil; f = f.Sibling {
					if f.DiffKey == childKey {
						childFiber = f
						if e.debug {
							log.EngineLogger.Debug("[buildComputedBoxWithSize.FALLBACK] ✅ Matched child %d by DiffKey: childKey=%q, childFiber.NodeID=%d",
								i, childKey, f.NodeID)
						}
						break
					}
				}

				if e.debug && childFiber == nil {
					log.EngineLogger.Debug("[buildComputedBoxWithSize.FALLBACK] ⚠️  No Fiber with DiffKey=%q found (expected for newly inserted nodes)",
						childKey)
				}
			} else {
				// No key: fall back to index matching (for static UI)
				childIndex := 0
				for f := fiber.Child; f != nil; f = f.Sibling {
					if childIndex == i {
						childFiber = f
						if e.debug {
							log.EngineLogger.Debug("[buildComputedBoxWithSize.FALLBACK] ✅ Matched child %d by index (no key): childFiber.NodeID=%d",
								i, f.NodeID)
						}
						break
					}
					childIndex++
				}

				if e.debug && childFiber == nil {
					log.EngineLogger.Debug("[buildComputedBoxWithSize.FALLBACK] ⚠️  No matching Fiber for index %d (child has no key)",
						i)
				}
			}
		}

		childBox := e.buildComputedBox(child, childFiber, box, childConstraints)
		if childBox != nil {
			box.Children = append(box.Children, childBox)
		}
	}

	// Cache the result for leaf nodes
	// Container nodes are not cached since their children may change
	if isLeaf {
		e.cache.Set(cacheKey, LayoutCacheEntry{
			Box:     box.Box,
			Size:    runtime.Size{Width: box.Box.Width, Height: box.Box.Height},
			Hash:    e.computeHash(vnode, constraints),
			IsLeaf:  true,
			VNodeID: vnode.Key(),
		})
		if e.debug {
			log.EngineLogger.Debug("[Layout.CacheSet] %s: %v\n",
				vnode.Type().String(), box.Box)
		}
	}

	return box
}

// buildComputedBox creates a computed box for a VNode with its children.
// This is a convenience wrapper for buildComputedBoxWithSize without pre-measurement.
//
// Parameters:
//   vnode: The VNode to build ComputedBox for
//   fiber: Optional Fiber node for passing NodeID (Phase 3: Identity Refactoring)
//   parent: Parent ComputedBox (for tree structure)
//   constraints: Box constraints for layout
func (e *Engine) buildComputedBox(vnode VNode, fiber *reconciler.Fiber, parent *ComputedBox, constraints runtime.BoxConstraints) *ComputedBox {
	return e.buildComputedBoxWithSize(vnode, fiber, parent, constraints, nil)
}

// =============================================================================
// Fiber-First Layout (Phase 2)
// =============================================================================
// buildComputedBoxFromFiber creates ComputedBox from Fiber tree only
// This is Phase 6 implementation that removes VNode dependency
//
// Parameters:
//   fiber: The Fiber node to build ComputedBox for
//   parent: Parent ComputedBox (for tree structure)
//   constraints: Box constraints for layout
//
// Returns:
//   *ComputedBox for this Fiber node
func (e *Engine) buildComputedBoxFromFiber(fiber *rtui.Fiber, parent *ComputedBox, constraints runtime.BoxConstraints) *ComputedBox {
	if fiber == nil {
		return nil
	}

	// Get layout info from Fiber (Phase 1: stored in completeWork)
	info := e.getLayoutInfoFromFiber(fiber)

	// Create base box using Fiber properties
	box := &ComputedBox{
		Parent:  parent,
		NodeID:  fiber.NodeID,
		Layer:   fiber.Layer,
		Box: runtime.Box{
			X:      0,
			Y:      0,
			Width:  0,
			Height: 0,
		},
		// Note: VNode field kept during transition, removed in Phase 4
	}

	// Measure layout using Fiber layout info
	measurement := e.measureFiberLayout(fiber, info, constraints)
	box.Box.Width = measurement.Size.Width
	box.Box.Height = measurement.Size.Height

	// Build children using Fiber tree traversal (Child->Sibling)
	children := e.getChildFibers(fiber)
	box.Children = make([]*ComputedBox, 0, len(children))

	for i, childFiber := range children {
		var childConstraints runtime.BoxConstraints
		if i < len(measurement.ChildConstraints) {
			childConstraints = measurement.ChildConstraints[i]
		} else {
			childConstraints = constraints
		}

		childBox := e.buildComputedBoxFromFiber(childFiber, box, childConstraints)
		if childBox != nil {
			box.Children = append(box.Children, childBox)
		}
	}

	return box
}

// getLayoutInfoFromFiber extracts layout info from Fiber using Fiber-first methods
func (e *Engine) getLayoutInfoFromFiber(fiber *rtui.Fiber) rtui.LayoutInfo {
	return rtui.LayoutInfo{
		IsHorizontal: fiber.GetDirection() == rtui.DirectionRow,
		Gap:         fiber.GetGap(),
		Flex:        fiber.GetFlex(),
		Align:        fiber.GetAlign(),
		CrossAlign:   fiber.GetCrossAlign(),
		Padding:      fiber.GetPadding(),
	}
}

// measureFiberLayout measures layout size from Fiber
func (e *Engine) measureFiberLayout(fiber *rtui.Fiber, info rtui.LayoutInfo, constraints runtime.BoxConstraints) runtime.LayoutMeasurement {
	// Try single-pass measurement via LayoutMeasurer interface
	if lm, ok := fiber.VNode.(runtime.LayoutMeasurer); ok {
		// Create a ChildMeasurer that uses Fiber children
		measurer := &fiberChildMeasurer{
			engine: e,
			fiber:  fiber,
		}
		return lm.MeasureLayout(measurer, constraints)
	}
	// Fallback to two-pass
	return runtime.LayoutMeasurement{
		Size: runtime.Size{Width: 0, Height: 0},
	}
}

// getChildFibers returns all child fibers as a slice
func (e *Engine) getChildFibers(fiber *rtui.Fiber) []*rtui.Fiber {
	return fiber.GetChildFibers()
}

// fiberChildMeasurer implements runtime.ChildMeasurer for Fiber-based layout
type fiberChildMeasurer struct {
	engine *Engine
	fiber  *rtui.Fiber
}

// MeasureChild measures a child Fiber's size with given constraints
func (m *fiberChildMeasurer) MeasureChild(child interface{}, constraints runtime.BoxConstraints) runtime.Size {
	childFiber, ok := child.(*rtui.Fiber)
	if !ok || childFiber == nil {
		return runtime.Size{Width: 0, Height: 0}
	}

	// Build child's ComputedBox to get its size
	box := m.engine.buildComputedBoxFromFiber(childFiber, nil, constraints)
	if box == nil {
		return runtime.Size{Width: 0, Height: 0}
	}

	return runtime.Size{
		Width:  box.Box.Width,
		Height: box.Box.Height,
	}
}

// measureVNode measures a VNode's size using constraints
func (e *Engine) measureVNode(vnode VNode, constraints runtime.BoxConstraints) runtime.Size {
	// Add layout tracing for debugging if enabled
	if e.debug {
		depth := e.getTraceDepth()
		e.incrementTraceDepth()
		defer e.decrementTraceDepth()

		indent := strings.Repeat("  ", depth)
		log.LayoutLogger.Debug("%s[Layout.ENTER] Type=%T %s Props:%v Constraints:%v",
			indent, vnode, vnode.Type().String(), vnode.Props(), constraints)

		size := e.doMeasureVNode(vnode, constraints)

		log.LayoutLogger.Debug("%s[Layout.LEAVE] %s Size:%v",
			indent, vnode.Type().String(), size)
		return size
	}

	return e.doMeasureVNode(vnode, constraints)
}

func (e *Engine) doMeasureVNode(vnode VNode, constraints runtime.BoxConstraints) runtime.Size {
	// SPECIAL CASE: Bordered nodes use Engine's measureBordered to handle
	// explicit width/height props correctly through constraints.
	// This bypasses the node's own Measure() method which doesn't check props.
	if isBordered(vnode) {
		return e.measureBordered(vnode, constraints)
	}

	// PRIORITY 1: Use Measurable interface (constraint-based measurement)
	if measurable, ok := vnode.(runtime.Measurable); ok {
		return measurable.Measure(constraints)
	}

	// PRIORITY 2: Check for known layout types
	if tagger, ok := vnode.(interface{ Tag() string }); ok {
		switch tagger.Tag() {
		case "hstack", "vstack":
			// LayoutNode types should implement Measurable
			// Fallback to measuring children
			return e.measureLayoutChildren(vnode, constraints)
		case "text":
			// Text node
			if text := rtui.GetTextContent(vnode); text != "" {
				runes := []rune(text)
				return runtime.Size{Width: len(runes), Height: 1}
			}
			return runtime.Size{Width: 0, Height: 1}
		case "table":
			return e.measureTable(vnode, constraints)
		}
	}

	// PRIORITY 3: Fallback to estimation
	return e.estimateSize(vnode, constraints)
}

// measureLayoutChildren measures layout containers (HStack/VStack)
func (e *Engine) measureLayoutChildren(vnode VNode, constraints runtime.BoxConstraints) runtime.Size {
	children := vnode.Children()
	if len(children) == 0 {
		return runtime.Size{Width: constraints.MinWidth, Height: constraints.MinHeight}
	}

	// Get layout info
	layoutInfo := rtui.GetLayoutInfo(vnode)
	gap := layoutInfo.Gap
	padding := layoutInfo.Padding // top, right, bottom, left
	paddingWidth := padding[1] + padding[3]
	paddingHeight := padding[0] + padding[2]

	// Get props for explicit width/height
	props := vnode.Props()

	if layoutInfo.IsHorizontal {
		// HStack: sum widths, max height
		totalWidth := 0
		maxHeight := 0

		// Check for explicit height prop - this creates a bounded constraint
		if height, ok := props["height"].(int); ok && height > 0 {
			constraints.MaxHeight = height
			if constraints.MinHeight > height {
				constraints.MinHeight = height
			}
		}

		// Calculate inner height constraint
		// Use parent's MaxHeight only if it's bounded, otherwise use Infinity
		innerMaxHeight := runtime.Infinity
		if constraints.HasBoundedHeight() {
			innerMaxHeight = max(0, constraints.MaxHeight-paddingHeight)
		}

		if e.debug {
			log.EngineLogger.Debug("[measureLayoutChildren.HStack] constraints=%v, paddingWidth=%d, paddingHeight=%d\n",
				constraints, paddingWidth, paddingHeight)
		}

		// First pass: identify flex children and measure non-flex children
		var flexChildren []struct {
			child  VNode
			index  int
			factor int
		}
		var fixedWidth int
		flexTotalFactor := 0

		for i, child := range children {
			childInfo := rtui.GetLayoutInfo(child)
			if childInfo.Flex > 0 {
				flexChildren = append(flexChildren, struct {
					child  VNode
					index  int
					factor int
				}{child, i, childInfo.Flex})
				flexTotalFactor += childInfo.Flex
			} else {
				// Non-flex child: measure with natural width
				childConstraints := runtime.BoxConstraints{
					MinWidth:  0,
					MaxWidth:  runtime.Infinity,
					MinHeight: 0,
					MaxHeight: innerMaxHeight,
				}
				childSize := e.measureVNode(child, childConstraints)
				fixedWidth += childSize.Width
				if childSize.Height > maxHeight && childSize.Height < runtime.Infinity {
					maxHeight = childSize.Height
				}
			}
			// Account for gap (except after last child)
			if i < len(children)-1 {
				fixedWidth += gap
			}
		}

		totalWidth = fixedWidth

		// If we have flex children, distribute space
		if len(flexChildren) > 0 {
			if constraints.HasBoundedWidth() {
				availableWidth := constraints.MaxWidth - paddingWidth - (len(children)-1)*gap
				remainingSpace := availableWidth - fixedWidth

				// Distribute remaining space to flex children
				for _, fc := range flexChildren {
					flexWidth := (remainingSpace * fc.factor) / flexTotalFactor
					if flexWidth < 0 {
						flexWidth = 0
					}

					childConstraints := runtime.BoxConstraints{
						MinWidth:  flexWidth,
						MaxWidth:  flexWidth,
						MinHeight: 0,
						MaxHeight: innerMaxHeight,
					}
					childSize := e.measureVNode(fc.child, childConstraints)
					totalWidth += childSize.Width
					if childSize.Height > maxHeight && childSize.Height < runtime.Infinity {
						maxHeight = childSize.Height
					}
				}
			} else {
				// No bounded width: measure flex children naturally
				for _, fc := range flexChildren {
					childConstraints := runtime.BoxConstraints{
						MinWidth:  0,
						MaxWidth:  runtime.Infinity,
						MinHeight: 0,
						MaxHeight: innerMaxHeight,
					}
					childSize := e.measureVNode(fc.child, childConstraints)
					totalWidth += childSize.Width
					if childSize.Height > maxHeight && childSize.Height < runtime.Infinity {
						maxHeight = childSize.Height
					}
				}
			}
		}

		// Add padding to total size
		totalWidth += paddingWidth
		maxHeight += paddingHeight

		// IMPORTANT: Cross-axis filling for HStack
		// Fill available height so children can stretch vertically
		if constraints.HasBoundedHeight() && maxHeight < constraints.MaxHeight {
			maxHeight = constraints.MaxHeight
		}

		// Apply MinHeight constraint
		if maxHeight < constraints.MinHeight {
			maxHeight = constraints.MinHeight
		}

		// Clamp to MaxHeight if exceeded
		if constraints.HasBoundedHeight() && maxHeight > constraints.MaxHeight {
			maxHeight = constraints.MaxHeight
		}

		if e.debug {
			log.EngineLogger.Debug("[measureLayoutChildren.HStack] RETURN: Width=%d, Height=%d\n",
				totalWidth, maxHeight)
		}

		return runtime.Size{Width: totalWidth, Height: maxHeight}
	} else {
		// VStack: max width, sum heights
		maxWidth := 0
		totalHeight := 0

		// Check for explicit height prop - this creates a bounded constraint
		if height, ok := props["height"].(int); ok && height > 0 {
			constraints.MaxHeight = height
			if constraints.MinHeight > height {
				constraints.MinHeight = height
			}
		}

		// Calculate inner width constraint
		// Use parent's MaxWidth only if it's bounded, otherwise use Infinity
		innerMaxWidth := runtime.Infinity
		if constraints.HasBoundedWidth() {
			innerMaxWidth = max(0, constraints.MaxWidth-paddingWidth)
		}

		if e.debug {
			log.EngineLogger.Debug("[measureLayoutChildren.VStack] constraints=%v, innerMaxWidth=%d\n",
				constraints, innerMaxWidth)
		}

		// First pass: identify flex children and measure non-flex children
		var flexChildren []struct {
			child  VNode
			index  int
			factor int
		}
		var fixedHeight int
		flexTotalFactor := 0

		for i, child := range children {
			childInfo := rtui.GetLayoutInfo(child)
			if childInfo.Flex > 0 {
				flexChildren = append(flexChildren, struct {
					child  VNode
					index  int
					factor int
				}{child, i, childInfo.Flex})
				flexTotalFactor += childInfo.Flex
			} else {
				// Non-flex child: measure with natural height
				// Special case: if child is HStack and parent has bounded width,
				// make it fill full width for main-axis alignment to work
				childMinWidth := 0
				if innerMaxWidth != runtime.Infinity && isHStack(child) {
					childMinWidth = innerMaxWidth // HStack in VStack fills width for alignment
				}
				childConstraints := runtime.BoxConstraints{
					MinWidth:  childMinWidth,
					MaxWidth:  innerMaxWidth,
					MinHeight: 0,
					MaxHeight: runtime.Infinity,
				}
				childSize := e.measureVNode(child, childConstraints)
				if childSize.Width > maxWidth && childSize.Width < runtime.Infinity {
					maxWidth = childSize.Width
				}
				if childSize.Height < runtime.Infinity {
					fixedHeight += childSize.Height
				}
			}
			// Account for gap (except after last child)
			if i < len(children)-1 {
				fixedHeight += gap
			}
		}

		totalHeight = fixedHeight

		// If we have flex children and bounded height, distribute remaining space
		if len(flexChildren) > 0 && constraints.HasBoundedHeight() {
			availableHeight := constraints.MaxHeight - paddingHeight - (len(children)-1)*gap
			remainingSpace := availableHeight - fixedHeight

			if e.debug && remainingSpace > 0 {
				log.EngineLogger.Debug("[measureLayoutChildren.VStack] flex distribution: available=%d, fixed=%d, remaining=%d, factors=%d\n",
					availableHeight, fixedHeight, remainingSpace, flexTotalFactor)
			}

			// Distribute remaining space to flex children
			for _, fc := range flexChildren {
				flexHeight := (remainingSpace * fc.factor) / flexTotalFactor
				if flexHeight < 0 {
					flexHeight = 0
				}

				// Special case: if child is HStack and parent has bounded width,
				// make it fill full width for main-axis alignment to work
				childMinWidth := 0
				if innerMaxWidth != runtime.Infinity && isHStack(fc.child) {
					childMinWidth = innerMaxWidth // HStack in VStack fills width
				}
				childConstraints := runtime.BoxConstraints{
					MinWidth:  childMinWidth,
					MaxWidth:  innerMaxWidth,
					MinHeight: flexHeight,
					MaxHeight: flexHeight,
				}
				childSize := e.measureVNode(fc.child, childConstraints)
				if childSize.Width > maxWidth && childSize.Width < runtime.Infinity {
					maxWidth = childSize.Width
				}
				totalHeight += childSize.Height
			}
		} else {
			// No flex or unbounded height: measure flex children naturally
			for _, fc := range flexChildren {
				// Special case: if child is HStack and parent has bounded width,
				// make it fill full width for main-axis alignment to work
				childMinWidth := 0
				if innerMaxWidth != runtime.Infinity && isHStack(fc.child) {
					childMinWidth = innerMaxWidth // HStack in VStack fills width
				}
				childConstraints := runtime.BoxConstraints{
					MinWidth:  childMinWidth,
					MaxWidth:  innerMaxWidth,
					MinHeight: 0,
					MaxHeight: runtime.Infinity,
				}
				childSize := e.measureVNode(fc.child, childConstraints)
				if childSize.Width > maxWidth && childSize.Width < runtime.Infinity {
					maxWidth = childSize.Width
				}
				if childSize.Height < runtime.Infinity {
					totalHeight += childSize.Height
				}
			}
		}

		// Add padding to total size
		maxWidth += paddingWidth
		totalHeight += paddingHeight

		// IMPORTANT: Cross-axis filling for VStack
		// Fill available width so children can stretch horizontally
		if constraints.HasBoundedWidth() && maxWidth < constraints.MaxWidth {
			maxWidth = constraints.MaxWidth
		}

		// Apply MinWidth constraint
		if maxWidth < constraints.MinWidth {
			maxWidth = constraints.MinWidth
		}

		// Clamp to MaxWidth if exceeded
		if constraints.HasBoundedWidth() && maxWidth > constraints.MaxWidth {
			maxWidth = constraints.MaxWidth
		}

		if e.debug {
			log.EngineLogger.Debug("[measureLayoutChildren.VStack] RETURN: Width=%d, Height=%d\n",
				maxWidth, totalHeight)
		}

		return runtime.Size{Width: maxWidth, Height: totalHeight}
	}
}

// measureBordered measures a bordered container
func (e *Engine) measureBordered(vnode VNode, constraints runtime.BoxConstraints) runtime.Size {
	children := vnode.Children()
	if len(children) == 0 {
		// Empty border - minimum size
		return runtime.Size{Width: 2, Height: 2}
	}

	// Check for explicit width/height props
	props := vnode.Props()
	explicitWidth := 0
	explicitHeight := 0
	hasWidthConstraint := false
	hasHeightConstraint := false

	if props != nil {
		if w, ok := props["width"].(int); ok && w > 0 {
			explicitWidth = w
			hasWidthConstraint = true
		}
		if h, ok := props["height"].(int); ok && h > 0 {
			explicitHeight = h
			hasHeightConstraint = true
		}
	}

	child := children[0]

	// Measure child with inner constraints (subtract border)
	// Use SubtractPadding helper to properly handle bounded/unbounded constraints
	innerConstraints := constraints.SubtractPadding(2, 2)

	// If explicit width is set, constrain to that width (minus border)
	if hasWidthConstraint {
		innerConstraints = runtime.NewBoxConstraints(
			explicitWidth-2, // MinWidth = MaxWidth to enforce fixed width
			explicitWidth-2, // MaxWidth
			0,
			innerConstraints.MaxHeight,
		)
	}

	// If explicit height is set, constrain to that height (minus border)
	if hasHeightConstraint {
		innerConstraints = runtime.NewBoxConstraints(
			innerConstraints.MinWidth,
			innerConstraints.MaxWidth,
			0,
			explicitHeight-2,
		)
	}

	childSize := e.measureVNode(child, innerConstraints)

	// Check for label
	if labeled, ok := vnode.(interface{ GetBorderLabel() string }); ok {
		label := labeled.GetBorderLabel()
		if label != "" {
			labelWidth := len(label) + 2 // +2 for spaces
			if labelWidth > childSize.Width {
				childSize.Width = labelWidth
			}
		}
	}

	// Calculate final size
	contentWidth := childSize.Width
	contentHeight := childSize.Height

	// Use explicit constraints if set
	if hasWidthConstraint {
		contentWidth = explicitWidth - 2 // Subtract border
	}
	if hasHeightConstraint {
		contentHeight = explicitHeight - 2 // Subtract border
	}

	// Border adds 2 to width and height
	result := runtime.Size{
		Width:  contentWidth + 2,
		Height: contentHeight + 2,
	}
	return result
}

// measureTable measures a table
func (e *Engine) measureTable(vnode VNode, constraints runtime.BoxConstraints) runtime.Size {
	rows := vnode.Children()
	if len(rows) == 0 {
		return runtime.Size{Width: 0, Height: 0}
	}

	// Calculate column widths
	colWidths := e.calculateColumnWidths(rows)

	// Sum column widths
	totalWidth := 0
	for _, w := range colWidths {
		totalWidth += w
	}

	// Sum row heights
	totalHeight := 0
	for _, row := range rows {
		rowHeight := e.measureVNode(row, constraints).Height
		totalHeight += rowHeight
	}

	return runtime.Size{Width: totalWidth, Height: totalHeight}
}

// calculateColumnWidths calculates the width of each column in a table
func (e *Engine) calculateColumnWidths(rows []VNode) []int {
	// Find max columns
	maxCols := 0
	for _, row := range rows {
		if tagger, ok := row.(interface{ Tag() string }); ok && tagger.Tag() == "tr" {
			cells := row.Children()
			if len(cells) > maxCols {
				maxCols = len(cells)
			}
		}
	}

	colWidths := make([]int, maxCols)

	// Find max width for each column
	for _, row := range rows {
		if tagger, ok := row.(interface{ Tag() string }); ok && tagger.Tag() == "tr" {
			cells := row.Children()
			for colIdx, cell := range cells {
				if colIdx < maxCols {
					cellWidth := e.measureVNode(cell, runtime.UnboundedConstraints()).Width
					if cellWidth > colWidths[colIdx] {
						colWidths[colIdx] = cellWidth
					}
				}
			}
		}
	}

	return colWidths
}

// estimateSize estimates size for unknown VNode types
func (e *Engine) estimateSize(vnode VNode, constraints runtime.BoxConstraints) runtime.Size {
	// Check for explicit width/height props
	if props := vnode.Props(); props != nil {
		if w := props.GetInt("width"); w > 0 {
			if h := props.GetInt("height"); h > 0 {
				return runtime.Size{Width: w, Height: h}
			}
			return runtime.Size{Width: w, Height: 1}
		}
	}

	// Check for text content
	if text := rtui.GetTextContent(vnode); text != "" {
		runes := []rune(text)
		return runtime.Size{Width: len(runes), Height: 1}
	}

	// Default minimum size
	return runtime.Size{Width: 10, Height: 1}
}

// measureTextWidth returns the display width of a text string
func (e *Engine) measureTextWidth(text string) int {
	return paint.StringWidth(text)
}

// getFlexDistribution calculates (or retrieves from cache) the flex distribution
// for a parent HStack or VStack. This avoids O(N²) re-measurement.
// isHorizontal: true for HStack, false for VStack
// maxSize: MaxHeight for HStack, MaxWidth for VStack
func (e *Engine) getFlexDistribution(parent VNode, isHorizontal bool, maxSize int) *FlexDistributionInfo {
	// Generate cache key from parent's key or address
	cacheKey := ""
	if key := parent.Key(); key != "" {
		cacheKey = key
	} else {
		// Fallback: use a unique identifier based on parent's address (not ideal but works)
		cacheKey = fmt.Sprintf("%p", parent)
	}

	// Check cache
	if cached, ok := e.flexCache[cacheKey]; ok && cached.Valid {
		return cached
	}

	// Calculate flex distribution
	info := &FlexDistributionInfo{
		TotalFlexFactor: 0,
		FixedSize:       0,
		ChildCount:      0,
		Valid:           true,
	}

	children := parent.Children()
	info.ChildCount = len(children)

	for _, child := range children {
		childInfo := rtui.GetLayoutInfo(child)
		if childInfo.Flex > 0 {
			info.TotalFlexFactor += childInfo.Flex
		} else {
			// Measure non-flex child
			var childSize runtime.Size
			if isHorizontal {
				// HStack: measure with unlimited width
				childConstraints := runtime.BoxConstraints{
					MinWidth:  0,
					MaxWidth:  runtime.Infinity,
					MinHeight: 0,
					MaxHeight: maxSize,
				}
				childSize = e.measureVNode(child, childConstraints)
				info.FixedSize += childSize.Width
			} else {
				// VStack: measure with unlimited height
				childConstraints := runtime.BoxConstraints{
					MinWidth:  0,
					MaxWidth:  maxSize,
					MinHeight: 0,
					MaxHeight: runtime.Infinity,
				}
				childSize = e.measureVNode(child, childConstraints)
				if childSize.Height < runtime.Infinity {
					info.FixedSize += childSize.Height
				}
			}
		}
	}

	// Store in cache
	e.flexCache[cacheKey] = info

	return info
}

// getChildConstraints calculates constraints for a child VNode
// IMPORTANT: Constraints should be passed based on parent's actual bounds, not Infinity
//
// NOTE: HStack/VStack nodes use the single-pass LayoutMeasurer path, so they are
// handled by MeasureLayout rather than this function. This function is only for
// nodes that don't implement LayoutMeasurer (fallback path).
func (e *Engine) getChildConstraints(parent, child VNode, parentConstraints runtime.BoxConstraints, parentSize runtime.Size) runtime.BoxConstraints {
	// Check if parent is a special node type
	if tagger, ok := parent.(interface{ Tag() string }); ok {
		switch tagger.Tag() {
		case "bordered":
			// Bordered: subtract border (2 units) from constraints
			// But check for explicit width/height props first
			props := parent.Props()
			explicitWidth, hasExplicitWidth := 0, false
			explicitHeight, hasExplicitHeight := 0, false

			if props != nil {
				if w, ok := props["width"].(int); ok && w > 0 {
					explicitWidth = w
					hasExplicitWidth = true
				}
				if h, ok := props["height"].(int); ok && h > 0 {
					explicitHeight = h
					hasExplicitHeight = true
				}
			}

			// If explicit width is set, create tight constraint for child
			if hasExplicitWidth {
				contentWidth := explicitWidth - 2 // Subtract border
				if contentWidth < 0 {
					contentWidth = 0
				}
				if hasExplicitHeight {
					contentHeight := explicitHeight - 2
					if contentHeight < 0 {
						contentHeight = 0
					}
					result := runtime.BoxConstraints{
						MinWidth:  contentWidth,
						MaxWidth:  contentWidth,
						MinHeight: contentHeight,
						MaxHeight: contentHeight,
					}
					return result
				}
				// Only width is explicit
				result := runtime.BoxConstraints{
					MinWidth:  contentWidth,
					MaxWidth:  contentWidth,
					MinHeight: 0,
					MaxHeight: runtime.Infinity,
				}
				return result
			}

			// No explicit width, use default behavior
			result := parentConstraints.SubtractPadding(2, 2)
			return result
		}
	}

	// Default: pass through parent constraints
	return parentConstraints
}

// =============================================================================
// Position Calculation (Second Pass: Layout)
// =============================================================================

// calculatePositions computes the (x, y) position for each computed box
// This is the second pass - after all sizes are known
func (e *Engine) calculatePositions(box *ComputedBox, x, y int) {
	box.Box.X = x
	box.Box.Y = y

	// Store bounds in VNode if it supports SetBounds (for Paint methods)
	if box.VNode != nil {
		// Store bounds in VNode if it supports SetBounds
		if boundsAware, ok := box.VNode.(interface{ SetBounds(int, int, int, int) }); ok {
			boundsAware.SetBounds(x, y, box.Box.Width, box.Box.Height)
		}
	}

	if e.debug {
		log.EngineLogger.Debug("[Layout.Position] %s at %s\n",
			box.VNode.Type(), box.Box.String())
	}

	// Layout children based on parent type
	if box.VNode != nil {
		if tagger, ok := box.VNode.(interface{ Tag() string }); ok {
			switch tagger.Tag() {
			case "hstack":
				e.layoutHStack(box, x, y)
				return
			case "vstack":
				e.layoutVStack(box, x, y)
				return
			case "bordered":
				e.layoutBordered(box, x, y)
				return
			case "table":
				e.layoutTable(box, x, y)
				return
			}
		}
	}

	// Default: stack children vertically
	e.layoutDefault(box, x, y)
}

// layoutHStack positions children horizontally
func (e *Engine) layoutHStack(box *ComputedBox, x, y int) {
	layoutInfo := rtui.GetLayoutInfo(box.VNode)
	gap := layoutInfo.Gap
	crossAlign := layoutInfo.CrossAlign
	mainAlign := layoutInfo.Align
	stretchCross := layoutInfo.StretchCross

	// Calculate total width of all children (for main-axis alignment)
	totalChildWidth := 0
	for _, child := range box.Children {
		totalChildWidth += child.Box.Width
	}
	if len(box.Children) > 0 {
		totalChildWidth += (len(box.Children) - 1) * gap
	}

	// Calculate starting X position based on main-axis alignment
	childX := x
	switch mainAlign {
	case rtui.AlignCenter:
		if totalChildWidth < box.Box.Width {
			childX = x + (box.Box.Width-totalChildWidth)/2
		}
	case rtui.AlignEnd:
		if totalChildWidth < box.Box.Width {
			childX = x + box.Box.Width - totalChildWidth
		}
	case rtui.AlignSpaceBetween:
		// For SpaceBetween, recalculate gap without including initial gap
		// SpaceBetween distributes ALL available space as gaps between items
		if len(box.Children) > 1 {
			// Calculate total button width only (without gaps)
			totalButtonWidth := 0
			for _, child := range box.Children {
				totalButtonWidth += child.Box.Width
			}
			// Distribute remaining space as gaps
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

	for i, child := range box.Children {
		childInfo := rtui.GetLayoutInfo(child.VNode)

		// Calculate child X position with individual alignment
		// If child was stretched by flex (allocated width > natural width),
		// apply mainAlign to position content within allocated space
		alignedChildX := childX
		if child.NaturalWidth > 0 && child.Box.Width > child.NaturalWidth {
			// Child was stretched (flex layout), apply alignment
			switch mainAlign {
			case rtui.AlignCenter:
				// Center: left padding = (allocated - natural) / 2
				padding := (child.Box.Width - child.NaturalWidth) / 2
				alignedChildX = childX + padding
			case rtui.AlignEnd:
				// Right align: left padding = allocated - natural
				padding := child.Box.Width - child.NaturalWidth
				alignedChildX = childX + padding
			case rtui.AlignStart, rtui.AlignSpaceBetween, rtui.AlignSpaceAround:
				// Left align (default): no adjustment
				alignedChildX = childX
			}
		}

		// Stretch child to container height if:
		// 1. Child has flex > 0 (explicit flex), OR
		// 2. Container has StretchCross enabled (auto-stretch all children), OR
		// 3. Child has FillHeight enabled (stretch this specific child)
		// IMPORTANT: Only stretch if container height is finite (not Infinity)
		if (childInfo.Flex > 0 || stretchCross || childInfo.FillHeight) && box.Box.Height < runtime.Infinity {
			child.Box.Height = box.Box.Height
		}

		// Calculate child Y position based on cross-axis alignment
		childY := y
		if child.Box.Height < box.Box.Height {
			switch crossAlign {
			case rtui.AlignCenter:
				childY = y + (box.Box.Height-child.Box.Height)/2
			case rtui.AlignEnd:
				childY = y + box.Box.Height - child.Box.Height
			case rtui.AlignStart, rtui.AlignSpaceBetween, rtui.AlignSpaceAround:
				// Default to top align
				childY = y
			}
		}

		e.calculatePositions(child, alignedChildX, childY)
		// Calculate next child position based on alignment mode
		if mainAlign == rtui.AlignSpaceAround {
			childX += child.Box.Width + gap
		} else if mainAlign == rtui.AlignSpaceBetween && i < len(box.Children)-1 {
			childX += child.Box.Width + gap
		} else {
			childX += child.Box.Width + layoutInfo.Gap
		}
	}
}

// layoutVStack positions children vertically
func (e *Engine) layoutVStack(box *ComputedBox, x, y int) {
	layoutInfo := rtui.GetLayoutInfo(box.VNode)
	gap := layoutInfo.Gap
	crossAlign := layoutInfo.CrossAlign
	stretchCross := layoutInfo.StretchCross

	childY := y
	for _, child := range box.Children {
		childInfo := rtui.GetLayoutInfo(child.VNode)
		oldWidth := child.Box.Width

		// Stretch child to container width if:
		// 1. Child has flex > 0 (explicit flex), OR
		// 2. Container has StretchCross enabled (auto-stretch all children)
		// 3. Child has FillWidth enabled (stretch this specific child)
		// IMPORTANT: Only stretch if container width is finite (not Infinity)
		if (childInfo.Flex > 0 || stretchCross || childInfo.FillWidth) && box.Box.Width < runtime.Infinity {
			child.Box.Width = box.Box.Width
		}

		// If text node was stretched, calculate RenderedText with padding
		// IMPORTANT: Only do this if the width is reasonable (not Infinity)
		if child.Box.Width > oldWidth && child.Box.Width > 0 && child.Box.Width < runtime.Infinity {
			if text := rtui.GetTextContent(child.VNode); text != "" {
				// Calculate padding needed
				textWidth := e.measureTextWidth(text)
				padding := child.Box.Width - textWidth
				if padding > 0 && padding < 1000 { // Also check padding is reasonable
					// Get text alignment from props (default: left)
					textAlign := runtime.TextAlignLeft
					if props := child.VNode.Props(); props != nil {
						if align, ok := props["textAlign"].(runtime.TextAlign); ok {
							textAlign = align
						} else if alignStr, ok := props["textAlign"].(string); ok {
							// Support string values too: "left", "center", "right"
							switch alignStr {
							case "center":
								textAlign = runtime.TextAlignCenter
							case "right":
								textAlign = runtime.TextAlignRight
							}
						}
					}

					// Apply padding based on alignment
					var leftPad, rightPad int
					switch textAlign {
					case runtime.TextAlignCenter:
						leftPad = padding / 2
						rightPad = padding - leftPad
					case runtime.TextAlignRight:
						leftPad = padding
						rightPad = 0
					default: // TextAlignLeft
						leftPad = 0
						rightPad = padding
					}

					// Build rendered text with padding
					rendered := strings.Repeat(" ", leftPad) + text + strings.Repeat(" ", rightPad)
					child.RenderedText = rendered
				}
			}
		}

		// Calculate child X position based on cross-axis alignment
		childX := x
		if child.Box.Width < box.Box.Width {
			switch crossAlign {
			case rtui.AlignCenter:
				childX = x + (box.Box.Width-child.Box.Width)/2
			case rtui.AlignEnd:
				childX = x + box.Box.Width - child.Box.Width
			case rtui.AlignStart, rtui.AlignSpaceBetween, rtui.AlignSpaceAround:
				// Default to left align
				childX = x
			}
		}

		e.calculatePositions(child, childX, childY)
		childY += child.Box.Height + gap
	}
}

// layoutBordered positions content inside border
func (e *Engine) layoutBordered(box *ComputedBox, x, y int) {
	// Content starts at (x+1, y+1) - inside the border
	for _, child := range box.Children {
		e.calculatePositions(child, x+1, y+1)
	}
}

// layoutTable positions table rows and cells
func (e *Engine) layoutTable(box *ComputedBox, x, y int) {
	// Calculate column widths
	colWidths := e.calculateColumnWidthsFromBoxes(box)

	rowY := y
	for _, child := range box.Children {
		if child.VNode != nil {
			if tagger, ok := child.VNode.(interface{ Tag() string }); ok && tagger.Tag() == "tr" {
				e.layoutTableRow(child, x, rowY, colWidths)
				rowY += child.Box.Height
			} else {
				e.calculatePositions(child, x, rowY)
				rowY += child.Box.Height
			}
		}
	}
}

// layoutTableRow positions cells in a table row
func (e *Engine) layoutTableRow(box *ComputedBox, x, y int, colWidths []int) {
	cellX := x
	for colIdx, child := range box.Children {
		e.calculatePositions(child, cellX, y)
		if colIdx < len(colWidths) {
			cellX += colWidths[colIdx]
		} else {
			cellX += child.Box.Width
		}
	}
}

// layoutDefault positions children vertically (default layout)
func (e *Engine) layoutDefault(box *ComputedBox, x, y int) {
	childY := y
	for _, child := range box.Children {
		e.calculatePositions(child, x, childY)
		childY += child.Box.Height
	}
}

// calculateColumnWidthsFromBoxes calculates column widths from computed boxes
func (e *Engine) calculateColumnWidthsFromBoxes(box *ComputedBox) []int {
	// Find max columns
	maxCols := 0
	for _, child := range box.Children {
		if child.VNode != nil {
			if tagger, ok := child.VNode.(interface{ Tag() string }); ok && tagger.Tag() == "tr" {
				if len(child.Children) > maxCols {
					maxCols = len(child.Children)
				}
			}
		}
	}

	colWidths := make([]int, maxCols)
	for _, rowBox := range box.Children {
		for colIdx, cellBox := range rowBox.Children {
			if colIdx < maxCols && cellBox.Box.Width > colWidths[colIdx] {
				colWidths[colIdx] = cellBox.Box.Width
			}
		}
	}

	return colWidths
}

// =============================================================================
// Cache Keys and Helpers
// =============================================================================

// getCacheKey generates a cache key for a VNode
//
// The cache key MUST include all properties that affect the VNode's size:
// - Text content for text nodes
// - Label for buttons/checkboxes
// - Value/placeholder for inputs
// - Border style/color/label for bordered nodes
// - Any other props that affect rendering
func (e *Engine) getCacheKey(vnode VNode, constraints runtime.BoxConstraints) LayoutCacheKey {
	key := LayoutCacheKey{
		VNodeType:   vnode.Type().String(),
		Constraints: constraints,
	}

	if keyNode := vnode.Key(); keyNode != "" {
		key.VNodeKey = keyNode
	}

	// Include content hash based on VNode type
	switch {
	// Text nodes - hash the text content
	case vnode.Type() == rtui.VNodeText:
		if text := rtui.GetTextContent(vnode); text != "" {
			key.ContentHash = hashString(text)
		}

	// Button nodes - hash the label (different labels = different sizes)
	case vnode.Type() == rtui.VNodeElement && vnode.Tag() == "button":
		if labeler, ok := vnode.(interface{ Label() string }); ok {
			if label := labeler.Label(); label != "" {
				key.ContentHash = hashString(label)
			}
		}

	// Input nodes - hash value/placeholder (affects displayed width)
	case vnode.Type() == rtui.VNodeElement && vnode.Tag() == "input":
		var content string
		if valuer, ok := vnode.(interface{ Value() string }); ok {
			content = valuer.Value()
		}
		if content == "" {
			if placer, ok := vnode.(interface{ Placeholder() string }); ok {
				content = placer.Placeholder()
			}
		}
		if content != "" {
			key.ContentHash = hashString(content)
		}

	// Checkbox nodes - hash label
	case vnode.Type() == rtui.VNodeElement && vnode.Tag() == "checkbox":
		if labeler, ok := vnode.(interface{ Label() string }); ok {
			if label := labeler.Label(); label != "" {
				key.ContentHash = hashString(label)
			}
		}

	// Bordered nodes - hash style, color, and label
	case vnode.Type() == rtui.VNodeElement && vnode.Tag() == "bordered":
		h := uint64(5381)
		if styler, ok := vnode.(interface{ GetBorderStyle() rtui.BorderStyle }); ok {
			h = h*31 + uint64(styler.GetBorderStyle())
		}
		if colorer, ok := vnode.(interface{ GetBorderColor() string }); ok {
			if color := colorer.GetBorderColor(); color != "" {
				h = h*31 + hashString(color)
			}
		}
		if labeler, ok := vnode.(interface{ GetBorderLabel() string }); ok {
			if label := labeler.GetBorderLabel(); label != "" {
				h = h*31 + hashString(label)
			}
		}
		key.ContentHash = h
	}

	// Include relevant props in the key
	if props := vnode.Props(); props != nil {
		if w := props.GetInt("width"); w > 0 {
			key.PropsHash = key.PropsHash*31 + uint64(w)
		}
		if h := props.GetInt("height"); h > 0 {
			key.PropsHash = key.PropsHash*31 + uint64(h)
		}
	}

	return key
}

// computeHash computes a hash for VNode + constraints
func (e *Engine) computeHash(vnode VNode, constraints runtime.BoxConstraints) uint64 {
	h := uint64(5381)

	// Hash type
	h = h*31 + uint64(vnode.Type())

	// Hash key
	if key := vnode.Key(); key != "" {
		h = h*31 + uint64(hashString(key))
	}

	// Hash constraints
	h = h*31 + uint64(constraints.MinWidth)
	h = h*31 + uint64(constraints.MaxWidth)
	h = h*31 + uint64(constraints.MinHeight)
	h = h*31 + uint64(constraints.MaxHeight)

	return h
}

// hashString computes a hash of a string for cache keys
func hashString(s string) uint64 {
	h := uint64(5381)
	for _, c := range s {
		h = h*31 + uint64(c)
	}
	return h
}

// =============================================================================
// Helper Functions
// =============================================================================

// =============================================================================
// Single-Pass Layout Support
// =============================================================================

// MeasureChild implements runtime.ChildMeasurer interface.
// This allows LayoutNode to call this without importing compute package.
func (e *Engine) MeasureChild(child interface{}, constraints runtime.BoxConstraints) runtime.Size {
	// Type assert to VNode
	if vnode, ok := child.(VNode); ok {
		return e.measureVNode(vnode, constraints)
	}
	// Fallback for non-VNode types
	return runtime.Size{Width: 0, Height: 0}
}

// TryMeasureLayout attempts to use the new LayoutMeasurer interface if available.
// Returns a valid LayoutMeasurement if the node implements LayoutMeasurer.
// Returns a zero LayoutMeasurement if the node should use the fallback path.
func (e *Engine) TryMeasureLayout(vnode VNode, constraints runtime.BoxConstraints) runtime.LayoutMeasurement {
	// Check if vnode implements LayoutMeasurer via the marker method
	if _, ok := vnode.(interface{ IsLayoutMeasurer() }); !ok {
		return runtime.LayoutMeasurement{}
	}

	// Type assert to LayoutMeasurer
	measurer, ok := vnode.(runtime.LayoutMeasurer)
	if !ok {
		return runtime.LayoutMeasurement{}
	}

	// Call MeasureLayout with the engine as the ChildMeasurer
	return measurer.MeasureLayout(e, constraints)
}

// =============================================================================
// Helper Functions
// =============================================================================

// isHStack checks if a VNode is an HStack (horizontal layout)
// Used to determine if a child should fill the available width for alignment
func isHStack(vnode rtui.VNode) bool {
	if vnode == nil {
		return false
	}
	// Check by tag
	if tagger, ok := vnode.(interface{ Tag() string }); ok {
		return tagger.Tag() == "hstack" || tagger.Tag() == "row"
	}
	return false
}

// isBordered checks if a VNode is a Bordered container
func isBordered(vnode VNode) bool {
	if vnode == nil {
		return false
	}
	if tagger, ok := vnode.(interface{ Tag() string }); ok {
		return tagger.Tag() == "bordered"
	}
	return false
}

// getTag returns the tag of a VNode for debugging
func getTag(vnode rtui.VNode) string {
	if vnode == nil {
		return "nil"
	}
	if tagger, ok := vnode.(interface{ Tag() string }); ok {
		return tagger.Tag()
	}
	return vnode.Type().String()
}

// =============================================================================
// Cache Management
// =============================================================================

// GetCacheStats returns statistics about the layout cache
func (e *Engine) GetCacheStats() CacheStats {
	return e.cache.Stats()
}

// ResetCacheStats resets cache hit/miss counters without clearing the cache
func (e *Engine) ResetCacheStats() {
	e.cache.ResetStats()
}

// ClearCache clears all cached layout results
func (e *Engine) ClearCache() {
	e.cache.Clear()
}

// InvalidateCacheByType removes all cached entries for a specific VNode type
func (e *Engine) InvalidateCacheByType(vNodeType string) {
	e.cache.InvalidateByType(vNodeType)
}

// InvalidateCacheByKey removes all cached entries for a specific VNode key
func (e *Engine) InvalidateCacheByKey(vnodeKey string) {
	e.cache.InvalidateByKey(vnodeKey)
}

// CacheSize returns the number of entries in the cache
func (e *Engine) CacheSize() int {
	return e.cache.Size()
}

// =============================================================================
// HitMap Building (Phase 1: Build HitMap during Layout)
// =============================================================================

// buildHitMapFromComputedBoxes builds a HitMap directly from the ComputedBox tree
// This is called AFTER layout calculations, including layer transforms like centering.
// The HitMap captures the FINAL positions that will be used for painting.
func (e *Engine) buildHitMapFromComputedBoxes(root *ComputedBox) *event.HitMap {
	if root == nil {
		return event.NewHitMap() // Return empty HitMap for nil root
	}

	var entries []event.HitMapEntryInternal

	// Recursively walk the ComputedBox tree and build HitMap entries
	var walk func(box *ComputedBox, zOrder int)
	walk = func(box *ComputedBox, zOrder int) {
		if box == nil {
			return
		}

		// Skip nodes with zero size (not rendered)
		if box.Box.Width <= 0 || box.Box.Height <= 0 {
			// Continue walking children (they might be visible)
			for _, child := range box.Children {
				walk(child, zOrder+1)
			}
			return
		}

		// Get NodeID from ComputedBox (now has uint64 NodeID field)
		// Fallback to string key conversion for compatibility during transition
		nodeID := box.NodeID
		if nodeID == 0 && box.VNode != nil {
			// Convert VNode key to NodeID using hash for compatibility
			if key := box.VNode.Key(); key != "" {
				nodeID = event.StringToNodeID(key)
			}
		}

		// Debug: Log NodeID collection
		if e.debug {
			var nodeInfo string
			if box.VNode != nil {
				nodeInfo = fmt.Sprintf("type=%s key=%s", box.VNode.Type().String(), box.VNode.Key())
			} else {
				nodeInfo = "VNode=nil"
			}
			log.EngineLogger.Debug("[buildHitMapFromComputedBoxes] Adding entry: NodeID=%d %s bounds=(%d,%d,%dx%d)",
				nodeID, nodeInfo, box.Box.X, box.Box.Y, box.Box.Width, box.Box.Height)
		}

		// Create entry using ComputedBox positions
		// ✅ These positions include ALL transforms (layout, layer centering, etc.)
		entry := event.HitMapEntryInternal{
			NodeID: nodeID,
			Node:   rtui.AsLayoutNode(box.VNode),
			Bounds: runtimelayout.Rect{
				X:      box.Box.X, // ✅ Final position after layer centering
				Y:      box.Box.Y,
				Width:  box.Box.Width,
				Height: box.Box.Height,
			},
			LocalXY: func(screenX, screenY int) (int, int) {
				// Convert screen coordinates to local coordinates relative to this node
				return screenX - box.Box.X, screenY - box.Box.Y
			},
			ZOrder: zOrder,
		}

		entries = append(entries, entry)

		// Recursively process children (higher Z-order)
		for _, child := range box.Children {
			walk(child, zOrder+1)
		}
	}

	walk(root, 0)

	// Sort by Z-order (ascending - lower Z first)
	// HitTest will iterate backwards to find upper layers first
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].ZOrder < entries[j].ZOrder
	})

	log.HitMapLogger.Debug("[Engine.buildHitMap] Built HitMap with %d entries", len(entries))

	// Build HitMap from entries using BuildFromEntries helper
	return event.BuildHitMapFromEntries(entries)
}

// =============================================================================
// Fiber-Based Layout (Phase 2: Layout 基于 Fiber)
// =============================================================================

// layoutFiber builds a ComputedBox for a Fiber node
// This is the new fiber-based layout implementation that uses Fiber as the primary source
// of layout information, with VNode only as backing data
//
// Parameters:
//   fiber: The Fiber node to layout
//   constraints: Box constraints for layout
//   depth: Tree depth (for debug logging)
//
// Returns:
//   *ComputedBox containing layout result
func (e *Engine) layoutFiber(fiber *rtui.Fiber, constraints runtime.BoxConstraints, depth int) *ComputedBox {
	if fiber == nil {
		return nil
	}

	// Check if Fiber already has a cached ComputedBox from previous layout
	// IMPORTANT: This cache is per-reconciliation cycle, not permanent cache
	if fiber.ComputedBox != nil {
		cachedBox, ok := fiber.ComputedBox.(*ComputedBox)
		if ok && cachedBox != nil {
			// Check if the cached box is still valid
			if !cachedBox.LayoutDirty {
				if e.debug {
					log.EngineLogger.Debug("[layoutFiber] Using cached ComputedBox for NodeID=%d", fiber.NodeID)
				}
				return cachedBox
			}
		}
	}

	// Build ComputedBox using fiber.VNode (still needed for content)
	// but associate it with fiber.NodeID and fiber.Layer
	box := e.buildComputedBox(fiber.VNode, fiber, nil, constraints)

	if box == nil {
		if e.debug {
			log.EngineLogger.Debug("[layoutFiber] ⚠️ nil ComputedBox for NodeID=%d", fiber.NodeID)
		}
		return nil
	}

	// Copy NodeID from Fiber
	box.NodeID = fiber.NodeID

	// Copy Layer from Fiber
	box.Layer = fiber.Layer

	// Store ComputedBox back in Fiber (for potential caching)
	fiber.ComputedBox = box

	if e.debug {
		log.EngineLogger.Debug("[layoutFiber] Created ComputedBox: NodeID=%d Layer=%d bounds=(%d,%d,%dx%d)",
			box.NodeID, box.Layer, box.Box.X, box.Box.Y, box.Box.Width, box.Box.Height)
	}

	return box
}

// LayoutFiber lays out the entire Fiber tree and populates each fiber.ComputedBox
// This is the new public API for fiber-based layout that replaces VNode-based layout
// It traverses the Fiber tree recursively and creates ComputedBox for each node
//
// IMPORTANT: This method sets fiber.ComputedBox for each Fiber node, which is required
// for BuildHitMapFromFiber() to work correctly in the new unified architecture
//
// Parameters:
//   root: The root Fiber node of the tree to layout
//   constraints: Box constraints for the entire tree layout
//
// Returns:
//   *ComputedLayout containing the root ComputedBox and HitMap
//   error if layout fails
//
// Usage Phase 5+:
//   layout := engine.LayoutFiber(fiberRoot, constraints)
//   renderPlanes := layer.BuildFromFiber(fiberRoot)
//   hitMap := ui.BuildHitMapFromFiber(fiberRoot)
func (e *Engine) LayoutFiber(root *rtui.Fiber, constraints runtime.BoxConstraints) (*ComputedLayout, error) {
	if root == nil {
		return nil, fmt.Errorf("cannot layout nil Fiber tree")
	}

	// Reset flex cache for each layout pass
	e.flexCache = make(map[string]*FlexDistributionInfo)

	// Layout the root Fiber node
	// layoutFiber() will recursively call buildComputedBox() which handles VNode children
	// But it won't traverse the Fiber tree siblings, so we need to ensure all Fiber nodes are processed
	rootBox := e.layoutFiber(root, constraints, 0)

	if rootBox == nil {
		layout := NewComputedLayout(nil)
		layout.HitMap = event.NewHitMap()
		return layout, nil
	}

	// Recursively ensure all Fiber nodes in the tree have ComputedBox set
	// This is needed because buildComputedBox() only processes VNode tree,
	// and Fiber tree might have different structure (e.g., unsorted siblings)
	e.ensureFiberComputedBox(root, constraints, 0)

	// Calculate positions (second pass)
	e.calculatePositions(rootBox, 0, 0)

	// Clear dirty flags after layout
	rootBox.ClearDirty()

	// Build HitMap directly from ComputedBox tree
	hitMap := e.buildHitMapFromComputedBoxes(rootBox)

	layout := NewComputedLayout(rootBox)
	layout.HitMap = hitMap

	log.HitMapLogger.Debug("[Engine.LayoutFiber] Built HitMap with %d entries", hitMap.Size())

	return layout, nil
}

// ensureFiberComputedBox traverses the Fiber tree and ensures each node has ComputedBox set
// This is a safety measure to ensure BuildHitMapFromFiber() has complete data
//
// Parameters:
//   fiber: The current Fiber node
//   constraints: Box constraints for this node
//   depth: Tree depth for debugging
func (e *Engine) ensureFiberComputedBox(fiber *rtui.Fiber, constraints runtime.BoxConstraints, depth int) {
	if fiber == nil {
		return
	}

	// If this Fiber doesn't have ComputedBox, create it
	if fiber.ComputedBox == nil {
		box := e.buildComputedBox(fiber.VNode, fiber, nil, constraints)
		if box != nil {
			box.NodeID = fiber.NodeID
			box.Layer = fiber.Layer
			fiber.ComputedBox = box
		}
	}

	// Process all siblings at this level
	for sibling := fiber.Sibling; sibling != nil; sibling = sibling.Sibling {
		e.ensureFiberComputedBox(sibling, constraints, depth)
	}

	// Process children (next level)
	if fiber.Child != nil {
		e.ensureFiberComputedBox(fiber.Child, constraints, depth+1)
	}
}

// layoutFiberTree recursively traverses the Fiber tree and creates ComputedBox for each node
// This uses depth-first traversal following Child → Sibling chains
//
// Parameters:
//   fiber: The current Fiber node to layout
//   constraints: Box constraints for this node
//   depth: Tree depth for debugging
//
// Returns:
//   *ComputedBox containing layout result for this node and its children
func (e *Engine) layoutFiberTree(fiber *rtui.Fiber, constraints runtime.BoxConstraints, depth int) *ComputedBox {
	if fiber == nil {
		return nil
	}

	// Layout current Fiber node
	box := e.layoutFiber(fiber, constraints, depth)
	if box == nil {
		return nil
	}

	// Layout children (first child)
	if fiber.Child != nil {
		childBox := e.layoutFiberTree(fiber.Child, constraints, depth+1)
		if childBox != nil {
			box.Children = []*ComputedBox{childBox}
		}
	}

	// Layout siblings (next sibling)
	// NOTE: In Fiber tree, siblings are linked via Sibling field, not as children in ComputedBox
	// We only layout this subtree, siblings will be handled by parent

	return box
}

// measureFiber measures the size of a Fiber node
// This is the new fiber-based measure implementation
//
// Parameters:
//   fiber: The Fiber node to measure
//   constraints: Box constraints for measurement
//
// Returns:
//   runtime.Box containing measured size
func (e *Engine) measureFiber(fiber *rtui.Fiber, constraints runtime.BoxConstraints) runtime.Box {
	if fiber == nil {
		return runtime.Box{}
	}

	// Measure using fiber.VNode (still needed for content)
	size := e.measureVNode(fiber.VNode, constraints)

	return runtime.Box{
		X:      0,
		Y:      0,
		Width:  size.Width,
		Height: size.Height,
	}
}

// layoutFiberChildren builds ComputedBox children from Fiber children
// This traverses the Fiber tree (Child -> Sibling) and builds ComputedBoxes for all children
//
// Parameters:
//   fiber: The parent Fiber node
//   constraints: Box constraints for layout
//   depth: Tree depth (for debug logging)
//
// Returns:
//   []*ComputedBox containing layout results for all children
func (e *Engine) layoutFiberChildren(fiber *rtui.Fiber, constraints runtime.BoxConstraints, depth int) []*ComputedBox {
	if fiber == nil || fiber.Child == nil {
		return nil
	}

	var children []*ComputedBox

	// Traverse Fiber children: Child -> Sibling chain
	for childFiber := fiber.Child; childFiber != nil; childFiber = childFiber.Sibling {
		childBox := e.layoutFiber(childFiber, constraints, depth+1)
		if childBox != nil {
			children = append(children, childBox)
		}
	}

	if e.debug {
		log.EngineLogger.Debug("[layoutFiberChildren] Processed %d children for NodeID=%d",
			len(children), fiber.NodeID)
	}

	return children
}

// buildHitMapFromFiber builds a HitMap directly from the Fiber tree
// This traverses the Fiber and collects ComputedBox entries from each Fiber
//
// Parameters:
//   root: The root Fiber node
//
// Returns:
//   *event.HitMap containing all hit test entries from the Fiber tree
func (e *Engine) buildHitMapFromFiber(root *rtui.Fiber) *event.HitMap {
	if root == nil {
		return event.NewHitMap()
	}

	var entries []event.HitMapEntryInternal

	// Recursively walk the Fiber tree and build HitMap entries
	var walk func(fiber *rtui.Fiber, treeDepth int)
	walk = func(fiber *rtui.Fiber, treeDepth int) {
		if fiber == nil {
			return
		}

		// Get ComputedBox from Fiber
		box, ok := fiber.ComputedBox.(*ComputedBox)
		if !ok || box == nil {
			// No ComputedBox - continue walking children
			for childFiber := fiber.Child; childFiber != nil; childFiber = childFiber.Sibling {
				walk(childFiber, treeDepth+1)
			}
			return
		}

		// Skip nodes with zero size
		if box.Box.Width <= 0 || box.Box.Height <= 0 {
			for childFiber := fiber.Child; childFiber != nil; childFiber = childFiber.Sibling {
				walk(childFiber, treeDepth+1)
			}
			return
		}

		// Calculate Z-order: int(fiber.Layer) * 10000 + treeDepth
		// This ensures proper layer priority: Overlay > Base, Modal > Overlay, etc.
		zOrder := int(fiber.Layer)*10000 + treeDepth

		if e.debug {
			log.EngineLogger.Debug("[buildHitMapFromFiber] Adding entry: NodeID=%d Layer=%d treeDepth=%d ZOrder=%d bounds=(%d,%d,%dx%d)",
				box.NodeID, fiber.Layer, treeDepth, zOrder, box.Box.X, box.Box.Y, box.Box.Width, box.Box.Height)
		}

		// Create HitMap entry
		entry := event.HitMapEntryInternal{
			NodeID: box.NodeID,
			Node:   rtui.AsLayoutNode(box.VNode),
			Bounds: runtimelayout.Rect{
				X:      box.Box.X,
				Y:      box.Box.Y,
				Width:  box.Box.Width,
				Height: box.Box.Height,
			},
			LocalXY: func(screenX, screenY int) (int, int) {
				return screenX - box.Box.X, screenY - box.Box.Y
			},
			ZOrder: zOrder,
		}

		entries = append(entries, entry)

		// Recursively process children (via Fiber tree: Child -> Sibling chain)
		for childFiber := fiber.Child; childFiber != nil; childFiber = childFiber.Sibling {
			walk(childFiber, treeDepth+1)
		}
	}

	walk(root, 0)

	// Sort by Z-order (ascending - lower Z first)
	// HitTest will iterate backwards to find upper layers first
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].ZOrder < entries[j].ZOrder
	})

	if e.debug {
		log.HitMapLogger.Debug("[buildHitMapFromFiber] Built HitMap with %d entries", len(entries))
	}

	// Build HitMap from entries
	return event.BuildHitMapFromEntries(entries)
}
