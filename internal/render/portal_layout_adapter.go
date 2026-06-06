// Package render provides Portal-aware layout engine adapters for two-phase layout
package render

import (
	"fmt"

	"github.com/wwsheng009/mint/internal/reconciler"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/types"
)

// =============================================================================// Two-Phase Layout System for Portals
// =============================================================================
//
// Portal System Architecture: "Fiber不动, Layout重建"
//
// Phase 1 - Main Tree Layout:
//   - Traverse the main Fiber tree
//   - Skip Portal nodes (collect them to a queue, don't layout children)
//   - Normal layout for all other nodes
//
// Phase 2 - Overlay Layout:
//   - Layout each Portal independently using Root coordinates
//   - Use PortalRoot as the anchor/positioning context
//   - Support PositionFixed and Anchor-based positioning
//   - Merge Portal layout results into final layout tree
//
// Portal vs Layer Distinction:
//   - Portal: Changes layout parent/coordinate system
//   - Layer: Controls render order (Z-index)
//   - These are orthogonal concepts

// =============================================================================
// PortalCollector - Collects Portal nodes during Phase 1
// =============================================================================

// PortalContext stores information about a Portal node
type PortalContext struct {
	PortalFiber  *reconciler.Fiber // The Portal node
	PortalRootID string            // Target PortalRoot ID
	Children     []layout.Node     // Portal children (to be laid out in Phase 2)
}

// PortalCollector collects Portal nodes during main tree layout
type PortalCollector struct {
	portals    []PortalContext
	portalRoot map[string]*reconciler.Fiber // portalRootID -> PortalRoot Fiber
}

func isPortalFiber(fiber *reconciler.Fiber) bool {
	if fiber == nil || fiber.Props == nil {
		return false
	}
	portalRootID, ok := fiber.Props["portalRoot"].(string)
	if !ok || portalRootID == "" {
		return false
	}
	if marked, ok := fiber.Props["_portal"].(bool); ok && marked {
		return true
	}
	if _, ok := fiber.Props["anchorId"].(string); ok {
		return true
	}
	if _, ok := fiber.Props["position"].(types.PositionType); ok {
		return true
	}
	if position, ok := fiber.Props["position"].(string); ok && position != "" {
		return true
	}
	if _, ok := fiber.Props["priority"].(int); ok {
		return true
	}
	return false
}

// NewPortalCollector creates a new PortalCollector
func NewPortalCollector() *PortalCollector {
	return &PortalCollector{
		portals:    make([]PortalContext, 0),
		portalRoot: make(map[string]*reconciler.Fiber),
	}
}

// Reset clears all per-frame collected portal state.
func (pc *PortalCollector) Reset() {
	if pc == nil {
		return
	}
	pc.portals = pc.portals[:0]
	for key := range pc.portalRoot {
		delete(pc.portalRoot, key)
	}
}

// CollectPortalRoots collects all PortalRoot nodes from the Fiber tree
func (pc *PortalCollector) CollectPortalRoots(fiber *reconciler.Fiber) {
	if fiber == nil {
		return
	}

	// Check if this fiber is a PortalRoot
	if fiber.Props != nil {
		if portalRootID, ok := fiber.Props["portalRootId"].(string); ok && portalRootID != "" {
			pc.portalRoot[portalRootID] = fiber
		}
	}

	// Recurse to children and siblings
	pc.CollectPortalRoots(fiber.Child)
	pc.CollectPortalRoots(fiber.Sibling)
}

// AddPortal adds a Portal node to the collection queue
func (pc *PortalCollector) AddPortal(portalFiber *reconciler.Fiber) error {
	if portalFiber == nil {
		return fmt.Errorf("portal fiber is nil")
	}

	if !isPortalFiber(portalFiber) {
		return nil
	}
	portalRootID, _ := portalFiber.Props["portalRoot"].(string)

	// Collect portal children
	children := make([]layout.Node, 0)
	childFiber := portalFiber.Child
	for childFiber != nil {
		children = append(children, NewFiberToNodeAdapterPure(childFiber))
		childFiber = childFiber.Sibling
	}

	context := PortalContext{
		PortalFiber:  portalFiber,
		PortalRootID: portalRootID,
		Children:     children,
	}

	pc.portals = append(pc.portals, context)

	return nil
}

// GetPortals returns all collected Portals
func (pc *PortalCollector) GetPortals() []PortalContext {
	return pc.portals
}

// GetPortalRoot returns the PortalRoot Fiber for a given portalRootID
func (pc *PortalCollector) GetPortalRoot(portalRootID string) (*reconciler.Fiber, bool) {
	root, exists := pc.portalRoot[portalRootID]
	return root, exists
}

// =============================================================================
// PortalAwareFiberToNodeAdapter - Adapts Fiber to layout.Node with Portal handling
// =============================================================================

// PortalAwareFiberToNodeAdapter wraps a Fiber tree to implement layout.Node interface
// During Phase 1 (main tree layout), it skips Portal nodes and collects them
type PortalAwareFiberToNodeAdapter struct {
	base       *FiberToNodeAdapter
	fiber      *reconciler.Fiber
	children   []layout.Node
	collector  *PortalCollector
	isPortal   bool
	portalRoot *reconciler.Fiber
}

// NewPortalAwareFiberToNodeAdapter creates a new adapter with Portal handling
func NewPortalAwareFiberToNodeAdapter(fiber *reconciler.Fiber, collector *PortalCollector) *PortalAwareFiberToNodeAdapter {
	adapter := &PortalAwareFiberToNodeAdapter{
		base:      NewFiberToNodeAdapterPure(fiber),
		fiber:     fiber,
		collector: collector,
		isPortal:  false,
	}

	// Check if this fiber is a Portal
	if isPortalFiber(fiber) {
		portalRootID, _ := fiber.Props["portalRoot"].(string)
		adapter.isPortal = true
		// Find the PortalRoot
		if root, exists := collector.GetPortalRoot(portalRootID); exists {
			adapter.portalRoot = root
		}
		// Add to collection queue
		collector.AddPortal(fiber)
	}

	// Initialize children (skip Portal children)
	adapter.initChildren()

	return adapter
}

// initChildren initializes children adapters from Fiber tree
// For Portal nodes, children are skipped (will be laid out in Phase 2)
func (a *PortalAwareFiberToNodeAdapter) initChildren() {
	if a.fiber == nil {
		return
	}

	// For Portal nodes, don't include children (they'll be laid out separately)
	if a.isPortal {
		a.children = make([]layout.Node, 0)
		return
	}

	// Build children from Fiber tree (Child -> Sibling linked list)
	childFibers := getFiberChildren(a.fiber)
	a.children = make([]layout.Node, len(childFibers))

	for i, childFiber := range childFibers {
		a.children[i] = NewPortalAwareFiberToNodeAdapter(childFiber, a.collector)
	}
}

// ID returns the node identifier
func (a *PortalAwareFiberToNodeAdapter) ID() string {
	if a.fiber == nil {
		return ""
	}
	return fmt.Sprintf("%d", a.fiber.NodeID)
}

// GetPropsID returns the business identifier (from Fiber.ID)
// Implements layout.PropsIDProvider interface
func (a *PortalAwareFiberToNodeAdapter) GetPropsID() string {
	if a.fiber == nil {
		return ""
	}
	return a.fiber.ID
}

// Type returns the node type
func (a *PortalAwareFiberToNodeAdapter) Type() string {
	if a.fiber == nil {
		return "unknown"
	}
	return string(a.fiber.Tag)
}

// Children returns child nodes
func (a *PortalAwareFiberToNodeAdapter) Children() []layout.Node {
	return a.children
}

// GetPosition returns the current position
func (a *PortalAwareFiberToNodeAdapter) GetPosition() (x, y int) {
	return 0, 0
}

// SetPosition sets the position
func (a *PortalAwareFiberToNodeAdapter) SetPosition(x, y int) {
	if a.fiber == nil {
		return
	}
	// Sync to Instance.bounds
	if a.fiber.Instance != nil {
		if positionable, ok := a.fiber.Instance.(interface{ SetPosition(x, y int) }); ok {
			positionable.SetPosition(x, y)
		}
		if boundsHaver, ok := a.fiber.Instance.(interface{ SetBounds(x, y, w, h int) }); ok {
			boundsHaver.SetBounds(x, y, 0, 0)
		}
	}
}

// GetSize returns the current size
func (a *PortalAwareFiberToNodeAdapter) GetSize() (width, height int) {
	if a.base == nil {
		return 0, 0
	}
	return a.base.GetSize()
}

// GetWidth returns the width
func (a *PortalAwareFiberToNodeAdapter) GetWidth() int {
	w, _ := a.GetSize()
	return w
}

// GetHeight returns the height
func (a *PortalAwareFiberToNodeAdapter) GetHeight() int {
	_, h := a.GetSize()
	return h
}

// SetSize sets the size
func (a *PortalAwareFiberToNodeAdapter) SetSize(width, height int) {
	if a.base == nil {
		return
	}
	a.base.SetSize(width, height)
}

// GetProp returns a property value
func (a *PortalAwareFiberToNodeAdapter) GetProp(key string) interface{} {
	if a.fiber == nil || a.fiber.Props == nil {
		return nil
	}
	return a.fiber.Props[key]
}

// GetInstancePosition returns the instance position from bounds
func (a *PortalAwareFiberToNodeAdapter) GetInstancePosition() (x, y int) {
	if a.fiber == nil || a.fiber.Instance == nil {
		return 0, 0
	}
	if boundsHaver, ok := a.fiber.Instance.(interface{ GetBounds() (int, int, int, int) }); ok {
		x, y, _, _ := boundsHaver.GetBounds()
		return x, y
	}
	return 0, 0
}

// IsPositionFixed returns true if this node uses fixed positioning
func (a *PortalAwareFiberToNodeAdapter) IsPositionFixed() bool {
	if a.fiber == nil {
		return false
	}
	// Check for PositionFixed prop
	if a.fiber.Props != nil {
		if pos, ok := a.fiber.Props["position"].(types.PositionType); ok {
			return pos == types.PositionFixed
		}
	}
	return false
}

func (a *PortalAwareFiberToNodeAdapter) Measure(constraints layout.Constraints) layout.Size {
	if a.base == nil {
		return layout.Size{}
	}
	return a.base.Measure(constraints)
}

func (a *PortalAwareFiberToNodeAdapter) GetMargin() layout.Margin {
	if a.base == nil {
		return layout.Margin{}
	}
	return a.base.GetMargin()
}

func (a *PortalAwareFiberToNodeAdapter) GetBoxModel() layout.BoxModel {
	if a.base == nil {
		return layout.BoxModel{}
	}
	return a.base.GetBoxModel()
}

func (a *PortalAwareFiberToNodeAdapter) GetFlexStyle() *layout.FlexStyle {
	if a.base == nil {
		return nil
	}
	return a.base.GetFlexStyle()
}

func (a *PortalAwareFiberToNodeAdapter) GetGridStyle() *layout.GridStyle {
	if a.base == nil {
		return nil
	}
	return a.base.GetGridStyle()
}

func (a *PortalAwareFiberToNodeAdapter) GetWrapStyle() *layout.WrapStyle {
	if a.base == nil {
		return nil
	}
	return a.base.GetWrapStyle()
}

func (a *PortalAwareFiberToNodeAdapter) GetAbsoluteStyle() *layout.AbsoluteStyle {
	if a.base == nil {
		return nil
	}
	return a.base.GetAbsoluteStyle()
}

func (a *PortalAwareFiberToNodeAdapter) GetBorder() layout.Border {
	if a.base == nil {
		return layout.Border{Style: layout.BorderNone}
	}
	return a.base.GetBorder()
}

func (a *PortalAwareFiberToNodeAdapter) ShouldCenter() bool {
	if a.base == nil {
		return false
	}
	return a.base.ShouldCenter()
}

func (a *PortalAwareFiberToNodeAdapter) GetPositionType() layout.PositionType {
	if a.base == nil {
		return layout.PositionRelative
	}
	return a.base.GetPositionType()
}

func (a *PortalAwareFiberToNodeAdapter) GetAnchor() layout.Anchor {
	if a.base == nil {
		return layout.AnchorTopLeft
	}
	return a.base.GetAnchor()
}

func (a *PortalAwareFiberToNodeAdapter) GetLayer() layout.Layer {
	if a.base == nil {
		return layout.LayerBase
	}
	return a.base.GetLayer()
}

func (a *PortalAwareFiberToNodeAdapter) GetZIndex() int {
	if a.base == nil {
		return 0
	}
	return a.base.GetZIndex()
}

func (a *PortalAwareFiberToNodeAdapter) GetScrollViewport() layout.ScrollViewport {
	if a.base == nil {
		return layout.ScrollViewport{}
	}
	return a.base.GetScrollViewport()
}

func (a *PortalAwareFiberToNodeAdapter) SetScrollViewportMetrics(contentWidth, contentHeight, viewportWidth, viewportHeight int) {
	if a.base == nil {
		return
	}
	a.base.SetScrollViewportMetrics(contentWidth, contentHeight, viewportWidth, viewportHeight)
}

// =============================================================================
// PortalAwareLayoutEngine - Two-phase layout engine for Portals
// =============================================================================

// Portal Z-index constants
// Portals use a separate Z-index range to ensure they are rendered above main tree
const (
	// PortalZIndexBase is the starting Z-index for Portals (1000)
	// This ensures Portals are above any main tree nodes (typically Z-index < 100)
	PortalZIndexBase = 1000
)

// PortalAwareLayoutEngine wraps layout.Engine to implement two-phase layout
type PortalAwareLayoutEngine struct {
	engine         *layout.Engine
	collector      *PortalCollector
	overlayManager *layout.OverlayManager // Overlay manager for portal management
}

// NewPortalAwareLayoutEngine creates a new Portal-aware layout engine
func NewPortalAwareLayoutEngine() *PortalAwareLayoutEngine {
	return &PortalAwareLayoutEngine{
		engine:         layout.NewEngine(),
		collector:      NewPortalCollector(),
		overlayManager: layout.NewOverlayManager(),
	}
}

// Layout performs two-phase layout: Main Tree + Overlays
func (e *PortalAwareLayoutEngine) Layout(fiber *reconciler.Fiber, constraints runtime.BoxConstraints) (*layout.LayoutResult, []layout.LayoutBox, error) {
	if fiber == nil {
		return nil, nil, fmt.Errorf("fiber is nil")
	}

	log := fmt.Sprintf

	// Reset per-frame portal state to avoid leaking stale portals across renders.
	e.collector.Reset()
	e.overlayManager.Clear()

	// Convert constraints
	layoutConstraints := layout.Constraints{
		MinWidth:  constraints.MinWidth,
		MaxWidth:  constraints.MaxWidth,
		MinHeight: constraints.MinHeight,
		MaxHeight: constraints.MaxHeight,
	}

	// === Phase 1: Collect PortalRoots ===
	e.collector.CollectPortalRoots(fiber)

	// === Phase 2: Main Tree Layout (skip Portals) ===
	mainNode := NewPortalAwareFiberToNodeAdapter(fiber, e.collector)
	mainResult := e.engine.Layout(mainNode, layoutConstraints)

	if mainResult == nil || mainResult.Root == nil {
		return nil, nil, fmt.Errorf("main tree layout failed")
	}

	// === Phase 3: Overlay Layout (each Portal independently) ===
	portalBoxes := make([]layout.LayoutBox, 0)
	portals := e.collector.GetPortals()

	ctxLog := fmt.Sprintf("[PortalAwareLayoutEngine] collected %d Portals", len(portals))
	log(ctxLog)

	for _, portal := range portals {
		portalBox := e.layoutPortal(portal, constraints, mainResult.Root)
		if portalBox != nil {
			// Register portal to OverlayManager
			// Use portal.PortalFiber.NodeID as unique ID
			portalID := fmt.Sprintf("portal-%d", portal.PortalFiber.NodeID)
			priority := e.getPortalPriority(portal.PortalFiber)

			e.overlayManager.Push(portalID, portalBox, portal.PortalRootID, priority)

			// Set portal box ID for HitMap identification
			portalBox.ID = portalID

			portalBoxes = append(portalBoxes, *portalBox)

			// Also append to root children for HitMap traversal
			// This ensures they are included in HitMap with higher zOrder
			mainResult.Root.Children = append(mainResult.Root.Children, portalBox)
		}
	}

	// Rebuild HitMap after portal boxes are merged into the main tree.
	// The initial layout engine hit map is computed before overlays are appended,
	// so without rebuilding it, portal content can be visible but not hittable.
	if mainResult.HitMap == nil {
		mainResult.HitMap = layout.NewHitMap()
	}
	mainResult.HitMap.BuildFromLayoutBox(mainResult.Root)

	return mainResult, portalBoxes, nil
}

// layoutPortal performs layout for a single Portal node
func (e *PortalAwareLayoutEngine) layoutPortal(
	portal PortalContext,
	constraints runtime.BoxConstraints,
	rootLayoutBox *layout.LayoutBox,
) *layout.LayoutBox {
	// Get PortalRoot (anchor for positioning)
	portalRoot, exists := e.collector.GetPortalRoot(portal.PortalRootID)
	if !exists {
		// PortalRoot doesn't exist, skip this portal
		return nil
	}

	// Get PortalRoot layout position from main tree
	rootPos := e.findPortalRootLayoutPosition(portalRoot, rootLayoutBox)
	if rootPos == nil {
		// PortalRoot not found in layout tree
		return nil
	}

	// Layout portal children to get portal size
	// Use unconstrained layout to measure portal content
	layoutConstraints := layout.Constraints{
		MinWidth:  0,
		MaxWidth:  constraints.MaxWidth, // Use viewport width as max
		MinHeight: 0,
		MaxHeight: constraints.MaxHeight, // Use viewport height as max
	}

	// Layout portal children as if they were at root level
	// Use a dummy root for layout calculation (coordinates are relative to Root)
	dummyRoot := &portalChildrenRootAdapter{
		children: portal.Children,
	}

	box := e.engine.Layout(dummyRoot, layoutConstraints)

	if box == nil || box.Root == nil {
		return nil
	}

	// Parse portal positioning configuration from Fiber props
	props := portal.PortalFiber.Props
	portalWidth, portalHeight := resolvePortalPositioningSize(props, box.Root)
	posConfig := layout.ParsePortalPositionConfig(
		props,
		constraints.MaxWidth,  // viewport width
		constraints.MaxHeight, // viewport height
		portalWidth,           // portal positioning width
		portalHeight,          // portal positioning height
	)

	// Check if using Anchor-based positioning
	if posConfig.AnchorID != "" {
		// Find anchor element position in main tree
		ax, ay, aw, ah, found := layout.FindAnchorPosition(rootLayoutBox, posConfig.AnchorID)
		if found {
			posConfig.SetAnchorPosition(ax, ay, aw, ah)
		}
	}

	// Calculate final portal position using PortalPositionCalculator
	finalX, finalY := 0, 0
	if popupResult, ok := layout.ResolveAnchoredPopupPositionFromProps(props, posConfig); ok {
		finalX, finalY = popupResult.X, popupResult.Y
	} else if popupResult, ok := layout.ResolveViewportClampedPopupPositionFromProps(props, posConfig); ok {
		finalX, finalY = popupResult.X, popupResult.Y
	} else {
		calculator := layout.NewPortalPositionCalculator()
		finalX, finalY = calculator.CalculatePosition(posConfig)
	}

	// Apply calculated position to the layout box
	e.applyPortalTransform(box.Root, finalX, finalY)

	// Set Portal Z-index to ensure it renders above main tree
	// Use separate Z-index range (PortalZIndexBase + priority)
	priority := e.getPortalPriority(portal.PortalFiber)
	e.setPortalZIndex(box.Root, PortalZIndexBase+priority)

	return box.Root
}

func resolvePortalPositioningSize(props map[string]interface{}, box *layout.LayoutBox) (width, height int) {
	if box != nil {
		width = box.Width
		height = box.Height
	}
	if props == nil {
		return width, height
	}
	if explicitWidth, ok := props["positioningWidth"].(int); ok && explicitWidth > 0 {
		width = explicitWidth
	}
	if explicitHeight, ok := props["positioningHeight"].(int); ok && explicitHeight > 0 {
		height = explicitHeight
	}
	return width, height
}

// findPortalRootLayoutPosition finds the layout position of a PortalRoot in the main tree
func (e *PortalAwareLayoutEngine) findPortalRootLayoutPosition(
	portalRoot *reconciler.Fiber,
	rootBox *layout.LayoutBox,
) *struct{ X, Y int } {
	// Traverse the layout tree to find the PortalRoot's position
	var result struct{ X, Y int }
	found := false

	var findNode func(box *layout.LayoutBox)
	findNode = func(box *layout.LayoutBox) {
		if box == nil {
			return
		}

		// Check if this layout box corresponds to the PortalRoot Fiber
		if box.ID != "" && box.ID == fmt.Sprintf("%d", portalRoot.NodeID) {
			result.X = box.X
			result.Y = box.Y
			found = true
			return
		}

		// Check children
		for _, child := range box.Children {
			findNode(child)
			if found {
				return
			}
		}
	}

	findNode(rootBox)

	if found {
		return &result
	}

	return nil
}

// applyPortalTransform applies calculated position to the layout box tree
func (e *PortalAwareLayoutEngine) applyPortalTransform(box *layout.LayoutBox, x, y int) {
	if box == nil {
		return
	}

	// Set the absolute position for this box
	box.X = x
	box.Y = y

	// Recursively apply to children (relative to parent)
	// Children in the layout tree have their own relative positions
	// which should be added to the parent's absolute position
	for _, child := range box.Children {
		childX := x + child.X
		childY := y + child.Y
		e.applyPortalTransform(child, childX, childY)
	}
}

// setPortalZIndex sets Z-index for a Portal and its children
// Uses a separate Z-index range to ensure Portals render above the main tree
func (e *PortalAwareLayoutEngine) setPortalZIndex(box *layout.LayoutBox, zIndex int) {
	if box == nil {
		return
	}

	// Use breadth-first traversal to assign ZIndex
	// This ensures siblings are numbered before their children
	// Portal: 1000, Child1: 1001, Child2: 1002, Grandchild: 1003
	queue := []struct {
		box    *layout.LayoutBox
		zIndex int
	}{
		{box, zIndex},
	}

	for len(queue) > 0 {
		// Dequeue
		current := queue[0]
		queue = queue[1:]

		// Set Z-index
		current.box.ZIndex = current.zIndex

		// Enqueue children with sequential ZIndex values
		for _, child := range current.box.Children {
			queue = append(queue, struct {
				box    *layout.LayoutBox
				zIndex int
			}{child, current.zIndex + len(queue) + 1})
		}
	}
}

// setPortalZIndexHelper is a helper that sets ZIndex and returns the next available ZIndex
// DEPRECATED: This was for depth-first, use setPortalZIndex (breadth-first) instead
func (e *PortalAwareLayoutEngine) setPortalZIndexHelper(box *layout.LayoutBox, zIndex int) int {
	if box == nil {
		return zIndex
	}

	// Set Z-index for this box
	box.ZIndex = zIndex

	// Process children and return next available ZIndex
	nextZIndex := zIndex + 1
	for _, child := range box.Children {
		nextZIndex = e.setPortalZIndexHelper(child, nextZIndex)
	}

	return nextZIndex
}

// GetEngine returns the underlying layout engine
func (e *PortalAwareLayoutEngine) GetEngine() *layout.Engine {
	return e.engine
}

// GetCollector returns the PortalCollector
func (e *PortalAwareLayoutEngine) GetCollector() *PortalCollector {
	return e.collector
}

// =============================================================================
// portalChildrenRootAdapter - A dummy root for laying out Portal children
// =============================================================================

// portalChildrenRootAdapter is a dummy root node that wraps Portal children
type portalChildrenRootAdapter struct {
	children []layout.Node
}

func (r *portalChildrenRootAdapter) ID() string {
	return "portal-root-adapter"
}

func (r *portalChildrenRootAdapter) Type() string {
	return "portal-root"
}

func (r *portalChildrenRootAdapter) Children() []layout.Node {
	return r.children
}

func (r *portalChildrenRootAdapter) GetPosition() (x, y int) {
	return 0, 0
}

func (r *portalChildrenRootAdapter) SetPosition(x, y int) {
}

func (r *portalChildrenRootAdapter) GetSize() (width, height int) {
	return 0, 0
}

func (r *portalChildrenRootAdapter) GetWidth() int {
	w, _ := r.GetSize()
	return w
}

func (r *portalChildrenRootAdapter) GetHeight() int {
	_, h := r.GetSize()
	return h
}

func (r *portalChildrenRootAdapter) SetSize(width, height int) {
}

func (r *portalChildrenRootAdapter) GetProp(key string) interface{} {
	return nil
}

// getPortalPriority returns the priority for a portal node
// Higher priority means closer to top (rendered last, z-index wise)
// Default priority is 0
func (e *PortalAwareLayoutEngine) getPortalPriority(fiber *reconciler.Fiber) int {
	if fiber == nil || fiber.Props == nil {
		return 0
	}

	if priority, ok := fiber.Props["priority"].(int); ok {
		return priority
	}

	// Default: 0
	return 0
}

// GetOverlayManager returns the overlay manager
func (e *PortalAwareLayoutEngine) GetOverlayManager() *layout.OverlayManager {
	return e.overlayManager
}
