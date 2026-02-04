// Package render provides declarative node implementation for bridging VNode and framework Component.
package render

import (
	"fmt"
	"os"
	"sync"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/framework/component"
	frameworkevent "github.com/wwsheng009/mint/framework/event"
	"github.com/wwsheng009/mint/internal/reconciler"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// DeclarativeNode - Bridges VNode with framework Component
// =============================================================================
// DeclarativeNode allows a VNode tree to be used as a framework Component.
// This enables mixing declarative UI (VNode) with imperative Components.

// DeclarativeNode wraps a VNode function for use as a framework Component
type DeclarativeNode struct {
	mu        sync.RWMutex
	root      rtui.VNode            // The root VNode of this tree
	renderFn  rtui.ComponentFunc     // Function that renders the VNode
	instance  *rtui.ComponentContext // Component instance for hooks
	focusMgr  *rtui.VNodeFocusManager // Focus manager for keyboard navigation

	// Framework integration
	fwApp     *framework.App         // Framework app (for triggering re-renders in non-Fiber mode)
	reconciler rtui.Reconciler       // Fiber reconciler (if enabled) - use interface to avoid import cycle
	renderer   rtui.VNodeRenderer    // VNode renderer (implements VNodeRenderer interface)
	useFiber   bool                // Whether Fiber mode is enabled
}

// NewDeclarativeNode creates a new declarative node from a VNode
func NewDeclarativeNode(vnode rtui.VNode) *DeclarativeNode {
	return &DeclarativeNode{
		root: vnode,
	}
}

// NewDeclarativeNodeFromFunc creates a new declarative node from a render function
func NewDeclarativeNodeFromFunc(fn rtui.ComponentFunc) *DeclarativeNode {
	node := &DeclarativeNode{
		renderFn: fn,
		instance: rtui.NewComponentContextForRoot(),
		focusMgr: rtui.NewVNodeFocusManager(),
		fwApp:    nil, // Will be set by SetFrameworkApp
		useFiber: false, // Default to non-Fiber mode
	}
	// Create the non-Fiber renderer
	node.renderer = NewNonFiberRenderer(node)
	return node
}

// SetFrameworkApp sets the framework app reference (called from ui/test.go in RunTest)
func (n *DeclarativeNode) SetFrameworkApp(app *framework.App) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.fwApp = app
}

// NewDeclarativeNodeFromFuncWithFiber creates a new declarative node with Fiber reconciler enabled
// This function is called from ui.Run when MINT_USE_FIBER is set
func NewDeclarativeNodeFromFuncWithFiber(fn rtui.ComponentFunc, fwApp *framework.App) *DeclarativeNode {
	// Import the reconciler package here to avoid import cycles in ui package
	// This is safe because internal/render can import internal/reconciler
	r := newFiberReconciler(fwApp, fn)
	focusMgr := rtui.NewVNodeFocusManager()

	// Set the focus manager on the reconciler so it can apply focus state before rendering
	if adapter, ok := r.(*fiberReconcilerAdapter); ok {
		adapter.SetFocusManager(focusMgr)
	}

	// Create the Fiber renderer with the render callback
	renderer := NewFiberRenderer(renderVNodeToBuffer)

	return &DeclarativeNode{
		renderFn:  fn,
		instance:  rtui.NewComponentContextForRoot(),
		focusMgr:  focusMgr,
		fwApp:     fwApp,
		reconciler: r,
		renderer:   renderer,
		useFiber:  true,
	}
}

// SetReconciler sets the Fiber reconciler for this node
// This is called by ui.Run when Fiber mode is enabled
func (n *DeclarativeNode) SetReconciler(r rtui.Reconciler) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.reconciler = r
	n.useFiber = r != nil
}

// =============================================================================
// framework.Node interface implementation
// =============================================================================

// ID returns the node ID
func (n *DeclarativeNode) ID() string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	if n.root != nil {
		if key := n.root.Key(); key != "" {
			return "declarative:" + key
		}
	}
	return "declarative:node"
}

// Type returns the node type
func (n *DeclarativeNode) Type() string {
	return "DeclarativeNode"
}

// Children returns child VNodes
func (n *DeclarativeNode) Children() []component.Node {
	// DeclarativeNode doesn't expose children as framework Nodes
	// They are managed internally through the VNode tree
	return nil
}

// =============================================================================
// framework.Measurable interface implementation
// =============================================================================

// Measure calculates the ideal size for this declarative node
func (n *DeclarativeNode) Measure(maxWidth, maxHeight int) (width, height int) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	// Default size if not measured
	width, height = maxWidth, 1

	// Try to get size from VNode if it has layout info
	if n.root != nil {
		if props := n.root.Props(); props != nil {
			if w := props.GetInt("width"); w > 0 {
				width = minInt(w, maxWidth)
			}
			if h := props.GetInt("height"); h > 0 {
				height = minInt(h, maxHeight)
			}
		}
	}

	return width, height
}

// =============================================================================
// framework.Paintable interface implementation
// =============================================================================

// Paint renders the VNode tree to the buffer
func (n *DeclarativeNode) Paint(ctx component.PaintContext, buf *paint.Buffer) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Debug logging
	if os.Getenv("TUI_DEBUG_UI") == "true" {
		fmt.Fprintf(os.Stderr, "DeclarativeNode.Paint: ctx.X=%d, ctx.Y=%d, buf=%dx%d, useFiber=%v\n",
			ctx.Bounds.X, ctx.Bounds.Y, buf.Width, buf.Height, n.useFiber)
	}

	// If Fiber reconciler is enabled, use it for rendering
	if n.useFiber && n.reconciler != nil {
		// The reconciler handles the entire render cycle
		// The reconciler will call the render function with proper context management
		n.reconciler.Render(ctx, buf, func() rtui.VNode {
			if n.renderFn != nil {
				return n.renderFn()
			}
			return n.root
		})

		// Get the rendered VNode tree from the reconciler for focus management
		// This avoids calling renderFn() twice which would create different VNode objects
		n.root = n.reconciler.GetRenderedRoot()
		return
	}

	// Non-Fiber mode: traditional rendering
	// Initialize component context if needed
	if n.instance == nil && n.renderFn != nil {
		n.instance = rtui.NewComponentContextForRoot()
		if os.Getenv("TUI_DEBUG_UI") == "true" {
			fmt.Fprintf(os.Stderr, "DeclarativeNode.Paint: created new ComponentContext\n")
		}
	}

	// Ensure root VNode is rendered
	if n.renderFn != nil {
		// Reset and set component context for hooks
		n.instance.ResetContext()
		rtui.SetCurrentContext(n.instance)

		if os.Getenv("TUI_DEBUG_UI") == "true" {
			fmt.Fprintf(os.Stderr, "DeclarativeNode.Paint: calling renderFn()...\n")
		}

		// Call render function to get root VNode
		n.root = n.renderFn()

		if os.Getenv("TUI_DEBUG_UI") == "true" {
			fmt.Fprintf(os.Stderr, "DeclarativeNode.Paint: renderFn() returned, root type=%d\n", n.root.Type())
		}

		// Expand all ComponentVNodes recursively (non-Fiber mode only)
		// This ensures nested components are rendered with proper hook context
		n.root = n.expandComponents(n.root)

		// Clear current context
		rtui.SetCurrentContext(nil)
	}

	// Collect focusable nodes and update focus manager (after expansion)
	if n.focusMgr != nil && n.root != nil {
		focusable := rtui.CollectFocusable(n.root)
		// In non-Fiber mode, use UpdateFocusableList instead of SetFocusable
		// because VNode instances are recreated on each render, so ID matching
		// (which includes pointer addresses) would fail. We preserve the focus
		// index directly instead.
		n.focusMgr.UpdateFocusableList(focusable)
		// Clamp focus index to valid range in case the number of focusable nodes changed
		currentIndex := n.focusMgr.CurrentIndex()
		if currentIndex >= len(focusable) {
			currentIndex = len(focusable) - 1
		}
		if currentIndex < 0 && len(focusable) > 0 {
			currentIndex = 0
		}
		if currentIndex >= 0 {
			n.focusMgr.SetFocusByIndex(currentIndex)
		}
		if os.Getenv("TUI_DEBUG_UI") == "true" {
			fmt.Fprintf(os.Stderr, "DeclarativeNode.Paint: collected %d focusable nodes\n", len(focusable))
		}

		// Apply focus state to VNodes (non-Fiber mode)
		n.applyFocus(focusable)
	}

	if n.root == nil {
		if os.Getenv("TUI_DEBUG_UI") == "true" {
			fmt.Fprintf(os.Stderr, "DeclarativeNode.Paint: root is nil, returning\n")
		}
		return
	}

	// Walk the VNode tree and paint each node
	n.PaintVNode(n.root, ctx.Bounds.X, ctx.Bounds.Y, buf)

	if os.Getenv("TUI_DEBUG_UI") == "true" {
		fmt.Fprintf(os.Stderr, "DeclarativeNode.Paint: painting complete\n")
	}
}

// PaintVNode recursively paints a VNode and its children.
//
// Rendering strategy:
// 1. If node implements Paintable: use Paint(x, y) → []DrawCmd → write to buffer
// 2. Otherwise: handle by node type (Text/Element/Fragment)
// 3. Always process children for container nodes
func (n *DeclarativeNode) PaintVNode(vnode rtui.VNode, x, y int, buf *paint.Buffer) {
	if vnode == nil {
		return
	}

	// Debug logging
	if os.Getenv("TUI_DEBUG_UI") == "true" {
		fmt.Fprintf(os.Stderr, "[PaintVNode] vnode type=%d (%s), x=%d, y=%d, actual type=%T\n",
			vnode.Type(), vnode.Type(), x, y, vnode)
	}

	// Set component bounds for mouse hit testing
	// Check if vnode implements SetBounds method
	if boundsAware, ok := vnode.(interface{ SetBounds(x, y, width, height int) }); ok {
		// Get the component's measured size
		width := 0
		height := 0
		if measurable, ok := vnode.(interface {
			Measure(constraints runtime.BoxConstraints) runtime.Size
		}); ok {
			size := measurable.Measure(runtime.BoxConstraints{})
			width = size.Width
			height = size.Height
		}
		// Debug logging
		if os.Getenv("TUI_DEBUG_UI") == "true" {
			fmt.Fprintf(os.Stderr, "[PaintVNode] Setting bounds: x=%d, y=%d, w=%d, h=%d\n", x, y, width, height)
		}
		// Set bounds for hit testing
		boundsAware.SetBounds(x, y, width, height)
	} else {
		if os.Getenv("TUI_DEBUG_UI") == "true" {
			fmt.Fprintf(os.Stderr, "[PaintVNode] vnode does not implement SetBounds\n")
		}
	}

	// Check if vnode implements Paintable interface (custom rendering)
	if paintable, ok := vnode.(interface{ Paint(int, int) []paint.DrawCmd }); ok {
		// Component has custom paint logic - use it
		commands := paintable.Paint(x, y)
		for _, cmd := range commands {
			buf.SetString(cmd.X, cmd.Y, cmd.Text, cmd.Style)
		}
		// Paintable components handle their own rendering, including children
		return
	}

	// Handle built-in VNode types
	switch vnode.Type() {
	case rtui.VNodeText:
		n.paintText(vnode, x, y, buf)

	case rtui.VNodeElement:
		n.paintElement(vnode, x, y, buf)

	case rtui.VNodeComponent:
		// Component nodes should be expanded before painting
		// In non-Fiber mode, expandComponents() handles this
		// In Fiber mode, components are already expanded in the Fiber tree

	case rtui.VNodeFragment:
		// Fragment - just paint children, no self-rendering
		n.paintChildren(vnode, x, y, buf)
	}

	// For non-Paintable elements, paint children after self-rendering
	if vnode.Type() == rtui.VNodeElement {
		children := vnode.Children()
		if len(children) > 0 {
			// Use shared layout detection utility
			layoutInfo := rtui.GetLayoutInfo(vnode)
			isHStack := layoutInfo.IsHorizontal
			gap := layoutInfo.Gap

			if isHStack {
				// Horizontal layout: paint children on the same line with x offset
				childX := x
				for _, child := range children {
					n.PaintVNode(child, childX, y, buf)
					// Move X by the width of this child + gap
					childX += n.MeasureVNodeWidth(child) + gap
				}
			} else {
				// Vertical layout: paint children on different lines
				childY := y
				for _, child := range children {
					n.PaintVNode(child, x, childY, buf)
					childY += n.MeasureVNodeHeight(child)
				}
			}
		}
	}
}

// paintText paints a text VNode
func (n *DeclarativeNode) paintText(vnode rtui.VNode, x, y int, buf *paint.Buffer) {
	text := rtui.GetTextContent(vnode)
	if text != "" {
		buf.SetString(x, y, text, vnode.Style())
	}
}

// paintElement paints an element VNode
func (n *DeclarativeNode) paintElement(vnode rtui.VNode, x, y int, buf *paint.Buffer) {
	// Check if element has text content (for text elements created with ui.Text)
	if content := rtui.GetTextContent(vnode); content != "" {
		buf.SetString(x, y, content, vnode.Style())
		return // Don't paint children for text elements
	}
	// For non-text elements, children will be painted after the switch
}

// paintChildren paints children of a VNode
func (n *DeclarativeNode) paintChildren(vnode rtui.VNode, x, y int, buf *paint.Buffer) {
	children := vnode.Children()
	if len(children) == 0 {
		return
	}

	childY := y
	for _, child := range children {
		n.PaintVNode(child, x, childY, buf)
		childY += n.MeasureVNodeHeight(child)
	}
}

// MeasureVNodeWidth measures the width of a VNode for horizontal layout
func (n *DeclarativeNode) MeasureVNodeWidth(vnode rtui.VNode) int {
	if vnode == nil {
		return 0
	}

	// Check the VNode type first
	switch vnode.Type() {
	case rtui.VNodeText:
		return len(rtui.GetTextContent(vnode))

	case rtui.VNodeElement:
		// Check if it's a button
		if btn, ok := vnode.(interface{ Label() string }); ok {
			return len(btn.Label()) + 2 // [label] with brackets
		}
		// For elements, try text content
		if content := rtui.GetTextContent(vnode); content != "" {
			return len(content)
		}
		// Check for label in props
		if props := vnode.Props(); props != nil {
			if label, ok := props["label"].(string); ok {
				return len(label) + 2 // [label]
			}
		}
		return 0
	}

	// For containers, calculate width as sum of children's widths
	layoutInfo := rtui.GetLayoutInfo(vnode)
	if layoutInfo.IsHorizontal {
		// HStack: sum of children's widths
		width := 0
		for _, child := range vnode.Children() {
			width += n.MeasureVNodeWidth(child)
		}
		return width
	}
	// VStack: return 0 (no width contribution)
	return 0
}

// MeasureVNodeHeight measures the height of a VNode for vertical layout
func (n *DeclarativeNode) MeasureVNodeHeight(vnode rtui.VNode) int {
	if vnode == nil {
		return 0
	}

	// Check for explicit height in props
	if props := vnode.Props(); props != nil {
		if h := props.GetInt("height"); h > 0 {
			return h
		}
	}

	// Check the VNode type
	switch vnode.Type() {
	case rtui.VNodeText, rtui.VNodeElement:
		// Text elements are single line
		if rtui.GetTextContent(vnode) != "" {
			return 1
		}
		// For elements without text content, check if they're containers
		layoutInfo := rtui.GetLayoutInfo(vnode)
		if layoutInfo.IsHorizontal {
			// HStack: height is max of children's heights
			maxHeight := 0
			for _, child := range vnode.Children() {
				childHeight := n.MeasureVNodeHeight(child)
				if childHeight > maxHeight {
					maxHeight = childHeight
				}
			}
			return maxHeight
		} else {
			// VStack: height is sum of children's heights
			totalHeight := 0
			for _, child := range vnode.Children() {
				totalHeight += n.MeasureVNodeHeight(child)
			}
			return totalHeight
		}
	case rtui.VNodeFragment:
		// Fragment: height is sum of children's heights
		totalHeight := 0
		for _, child := range vnode.Children() {
			totalHeight += n.MeasureVNodeHeight(child)
		}
		return totalHeight
	}

	// Default height for leaf nodes
	return 1
}

// applyFocus applies focus state to VNodes based on the current focus index.
// This is called in non-Fiber mode to ensure the focused element is visually highlighted.
func (n *DeclarativeNode) applyFocus(focusable []rtui.FocusableVNode) {
	if n.focusMgr == nil {
		return
	}

	focusedIndex := n.focusMgr.CurrentIndex()
	if focusedIndex < 0 || focusedIndex >= len(focusable) {
		// No valid focus, clear all focus
		for _, elem := range focusable {
			elem.SetFocus(false)
		}
		return
	}

	// Set focus by index
	for i, elem := range focusable {
		if i == focusedIndex {
			if os.Getenv("TUI_DEBUG_FOCUS") == "true" {
				fmt.Fprintf(os.Stderr, "[applyFocus] setting focus=true on index %d (%s)\n", i, elem.GetFocusID())
			}
			elem.SetFocus(true)
		} else {
			elem.SetFocus(false)
		}
	}
}

// expandComponents recursively expands all ComponentVNodes in the tree
// This is used in non-Fiber mode to ensure components are rendered with proper hook context
// The hook context must be set before calling this function
func (n *DeclarativeNode) expandComponents(vnode rtui.VNode) rtui.VNode {
	if vnode == nil {
		return nil
	}

	// If this is a ComponentVNode, expand it by calling its render function
	if vnode.Type() == rtui.VNodeComponent {
		if componentVNode, ok := vnode.(*rtui.ComponentVNode); ok {
			rendered := componentVNode.Render()
			if rendered != nil {
				// Recursively expand the rendered VNode
				return n.expandComponents(rendered)
			}
		}
		return nil
	}

	// For non-component nodes, we need to expand their children
	// Create a new VNode with expanded children
	switch v := vnode.(type) {
	case *rtui.ElementVNode:
		// Expand children of ElementVNode
		children := vnode.Children()
		if len(children) == 0 {
			return vnode
		}
		expandedChildren := make([]rtui.VNode, 0, len(children))
		for _, child := range children {
			expanded := n.expandComponents(child)
			if expanded != nil {
				expandedChildren = append(expandedChildren, expanded)
			}
		}
		// Create a clone with expanded children
		cloned := rtui.NewElement(v.Tag())
		cloned.SetProps(vnode.Props())
		cloned.SetKey(vnode.Key())
		cloned.SetStyle(vnode.Style())
		cloned.SetChildren(expandedChildren)
		return cloned

	case *rtui.LayoutNode:
		// LayoutNode embeds ElementVNode, so similar handling
		children := vnode.Children()
		if len(children) == 0 {
			return vnode
		}
		expandedChildren := make([]rtui.VNode, 0, len(children))
		for _, child := range children {
			expanded := n.expandComponents(child)
			if expanded != nil {
				expandedChildren = append(expandedChildren, expanded)
			}
		}
		// For LayoutNode, we can't directly clone it because private fields
		// Instead, modify the existing node's children in place
		vnode.SetChildren(expandedChildren)
		return vnode

	case *rtui.FragmentVNode:
		// Fragment - expand children
		children := vnode.Children()
		if len(children) == 0 {
			return vnode
		}
		expandedChildren := make([]rtui.VNode, 0, len(children))
		for _, child := range children {
			expanded := n.expandComponents(child)
			if expanded != nil {
				expandedChildren = append(expandedChildren, expanded)
			}
		}
		return rtui.Fragment(expandedChildren...)

	default:
		// For TextVNode and other leaf nodes, return as-is
		return vnode
	}
}

// =============================================================================
// framework.Mountable interface implementation
// =============================================================================

// Mount mounts this node to a parent container
func (n *DeclarativeNode) Mount(parent component.Container) {
	n.mu.Lock()
	defer n.mu.Unlock()
	// DeclarativeNode doesn't need explicit mount handling
	// The VNode tree is self-contained
}

// Unmount unmounts this node from its parent
func (n *DeclarativeNode) Unmount() {
	n.mu.Lock()
	defer n.mu.Unlock()
	// Cleanup any resources
	n.root = nil
}

// IsMounted returns whether this node is mounted
func (n *DeclarativeNode) IsMounted() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.root != nil
}

// =============================================================================
// VNode access
// =============================================================================

// Root returns the root VNode
func (n *DeclarativeNode) Root() rtui.VNode {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.root
}

// UpdateRoot updates the root VNode
func (n *DeclarativeNode) UpdateRoot(vnode rtui.VNode) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.root = vnode
}

// =============================================================================
// framework.Event Component interface implementation
// =============================================================================

// HandleEvent processes events by distributing them to child components
func (n *DeclarativeNode) HandleEvent(ev frameworkevent.Event) bool {
	if os.Getenv("TUI_DEBUG_UI") == "true" {
		fmt.Fprintf(os.Stderr, "DeclarativeNode.HandleEvent: event type=%d\n", ev.Type())
	}

	n.mu.RLock()
	root := n.root
	focusMgr := n.focusMgr
	useFiber := n.useFiber
	reconciler := n.reconciler
	n.mu.RUnlock()

	if root == nil {
		return false
	}

	// 1. Let focus manager handle navigation (Tab, Shift+Tab)
	if focusMgr != nil {
		handled, shouldRender := focusMgr.HandleEvent(ev)
		if handled {
			if os.Getenv("TUI_DEBUG_UI") == "true" {
				fmt.Fprintf(os.Stderr, "DeclarativeNode.HandleEvent: focus manager handled event, shouldRender=%v\n", shouldRender)
			}
			// Request a re-render when focus changes
			// In Fiber mode, use the reconciler; in non-Fiber mode, mark as dirty
			if shouldRender {
				if useFiber && reconciler != nil {
					// Fiber mode: schedule reconciler update
					if r, ok := reconciler.(*fiberReconcilerAdapter); ok {
						r.r.ScheduleUpdate(rtui.LaneSyncLane)
					}
				} else {
					// Non-Fiber mode: mark framework app as dirty to trigger re-render
					// This ensures the updated focus state is painted
					if fwApp := n.getFrameworkApp(); fwApp != nil {
						fwApp.MarkDirty()
					}
				}
			}
			return true
		}

		// 2. Try to dispatch to the focused element first
		if focusMgr.DispatchToFocused(ev) {
			if os.Getenv("TUI_DEBUG_UI") == "true" {
				fmt.Fprintf(os.Stderr, "DeclarativeNode.HandleEvent: focused element handled event\n")
			}
			return true
		}
	}

	// 3. Fall back to global event distribution
	return n.distributeEventToVNode(root, ev)
}

// distributeEventToVNode recursively distributes an event to VNode tree
func (n *DeclarativeNode) distributeEventToVNode(vnode rtui.VNode, ev frameworkevent.Event) bool {
	if os.Getenv("TUI_DEBUG_UI") == "true" {
		fmt.Fprintf(os.Stderr, "distributeEventToVNode: called with vnode type=%d, actual type=%T\n", vnode.Type(), vnode)
	}

	if vnode == nil {
		if os.Getenv("TUI_DEBUG_UI") == "true" {
			fmt.Fprintf(os.Stderr, "distributeEventToVNode: vnode is nil\n")
		}
		return false
	}

	// Check if this VNode implements the Component interface
	if component, ok := vnode.(frameworkevent.Component); ok {
		if os.Getenv("TUI_DEBUG_UI") == "true" {
			fmt.Fprintf(os.Stderr, "distributeEventToVNode: VNode type=%d implements frameworkevent.Component, calling HandleEvent\n", vnode.Type())
		}
		if component.HandleEvent(ev) {
			// Event was handled by this component - stop propagation
			// This prevents keyboard events from triggering multiple buttons
			return true
		}
	}

	// Try to distribute to children
	children := vnode.Children()
	if len(children) > 0 {
		if os.Getenv("TUI_DEBUG_UI") == "true" {
			fmt.Fprintf(os.Stderr, "distributeEventToVNode: VNode type=%d has %d children, distributing...\n", vnode.Type(), len(children))
		}
		for _, child := range children {
			if n.distributeEventToVNode(child, ev) {
				// Event was handled by a child - stop propagation
				return true
			}
		}
	}

	return false
}

// =============================================================================
// Test Helper Methods
// =============================================================================

// GetFocusManager returns the focus manager for this declarative node
func (n *DeclarativeNode) GetFocusManager() *rtui.VNodeFocusManager {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.focusMgr
}

// GetRenderer returns the VNode renderer for this declarative node
func (n *DeclarativeNode) GetRenderer() rtui.VNodeRenderer {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.renderer
}

// GetFocusedIndex returns the index of the currently focused element
func (n *DeclarativeNode) GetFocusedIndex() int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	if n.focusMgr == nil {
		return -1
	}
	return n.focusMgr.CurrentIndex()
}

// GetFocusedType returns the type of the currently focused element
func (n *DeclarativeNode) GetFocusedType() int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	if n.focusMgr == nil {
		return 0
	}
	current := n.focusMgr.GetCurrent()
	if current == nil {
		return 0
	}
	// Return VNodeType as int
	return int(current.Type())
}

// getFrameworkApp returns the framework app (for triggering re-renders)
func (n *DeclarativeNode) getFrameworkApp() *framework.App {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.fwApp
}

// GetButtons returns all button VNodes in the tree
func (n *DeclarativeNode) GetButtons() []rtui.FocusableVNode {
	n.mu.RLock()
	defer n.mu.RUnlock()

	// In Fiber mode, collect buttons directly from the Fiber tree
	// This preserves the original button VNode objects with their layout bounds
	if n.useFiber && n.reconciler != nil {
		// Access the reconciler directly to get the Fiber root
		// The reconciler is a fiberReconcilerAdapter which wraps the actual reconciler
		if adapter, ok := n.reconciler.(*fiberReconcilerAdapter); ok {
			if fiberRoot := adapter.r.GetFiberRoot(); fiberRoot != nil {
				return n.collectButtonsFromFiber(fiberRoot)
			}
		}
	}

	// Non-Fiber mode: use the traditional collection method
	return collectByType(n.root, func(vnode rtui.VNode) bool {
		// Check if this is a button by checking if it has a Tag() method that returns "button"
		if tag, hasTag := vnode.(interface{ Tag() string }); hasTag && tag.Tag() == "button" {
			if focusable, ok := vnode.(rtui.FocusableVNode); ok {
				return focusable.IsFocusable()
			}
		}
		return false
	})
}

// collectButtonsFromFiber traverses the Fiber tree to collect button VNodes
// This preserves the original button objects with their layout bounds
func (n *DeclarativeNode) collectButtonsFromFiber(fiber *reconciler.Fiber) []rtui.FocusableVNode {
	var result []rtui.FocusableVNode

	if fiber == nil || fiber.VNode == nil {
		return result
	}

	// Skip the root ComponentVNode wrapper
	if fiber.Key == "root" && fiber.VNode.Type() == rtui.VNodeComponent {
		return n.collectButtonsFromFiber(fiber.Child)
	}

	// Check if current VNode is a button
	if n.isButtonVNode(fiber.VNode) {
		if focusable, ok := fiber.VNode.(rtui.FocusableVNode); ok && focusable.IsFocusable() {
			result = append(result, focusable)
		}
	}

	// Recursively check children
	if child := fiber.Child; child != nil {
		result = append(result, n.collectButtonsFromFiber(child)...)
	}

	// Recursively check siblings
	if sibling := fiber.Sibling; sibling != nil {
		result = append(result, n.collectButtonsFromFiber(sibling)...)
	}

	return result
}

// isButtonVNode checks if a VNode is a button element
func (n *DeclarativeNode) isButtonVNode(vnode rtui.VNode) bool {
	if vnode == nil {
		return false
	}
	// Check if this is a button by checking if it has a Tag() method that returns "button"
	if tag, hasTag := vnode.(interface{ Tag() string }); hasTag && tag.Tag() == "button" {
		return true
	}
	return false
}

// collectFocusableFromFiber collects all focusable elements from the Fiber tree
// This includes buttons, inputs, checkboxes, textareas, and selects
func (n *DeclarativeNode) collectFocusableFromFiber(fiber *reconciler.Fiber) []rtui.FocusableVNode {
	var result []rtui.FocusableVNode

	if fiber == nil || fiber.VNode == nil {
		return result
	}

	// Skip the root ComponentVNode wrapper
	if fiber.Key == "root" && fiber.VNode.Type() == rtui.VNodeComponent {
		return n.collectFocusableFromFiber(fiber.Child)
	}

	// Check if current VNode is focusable
	if focusable, ok := fiber.VNode.(rtui.FocusableVNode); ok {
		if focusable.IsFocusable() {
			// Set focusIndex for buttons to ensure unique FocusID
			if btn, ok := fiber.VNode.(interface{ SetFocusIndex(int) }); ok {
				btn.SetFocusIndex(len(result))
			}
			result = append(result, focusable)
		}
	}

	// Recursively check children
	if child := fiber.Child; child != nil {
		result = append(result, n.collectFocusableFromFiber(child)...)
	}

	// Recursively check siblings
	if sibling := fiber.Sibling; sibling != nil {
		result = append(result, n.collectFocusableFromFiber(sibling)...)
	}

	return result
}

// GetInputs returns all input VNodes in the tree
func (n *DeclarativeNode) GetInputs() []rtui.FocusableVNode {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return collectByType(n.root, func(vnode rtui.VNode) bool {
		if focusable, ok := vnode.(rtui.FocusableVNode); ok {
			// Check if this is an input by its tag
			if vnode.Type() == rtui.VNodeElement {
				if tag, ok := vnode.(interface{ Tag() string }); ok && tag.Tag() == "input" {
					return focusable.IsFocusable()
				}
			}
		}
		return false
	})
}

// GetTextareas returns all textarea VNodes in the tree
func (n *DeclarativeNode) GetTextareas() []rtui.FocusableVNode {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return collectByType(n.root, func(vnode rtui.VNode) bool {
		if focusable, ok := vnode.(rtui.FocusableVNode); ok {
			if vnode.Type() == rtui.VNodeElement {
				if tag, ok := vnode.(interface{ Tag() string }); ok && tag.Tag() == "textarea" {
					return focusable.IsFocusable()
				}
			}
		}
		return false
	})
}

// GetCheckboxes returns all checkbox VNodes in the tree
func (n *DeclarativeNode) GetCheckboxes() []rtui.FocusableVNode {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return collectByType(n.root, func(vnode rtui.VNode) bool {
		if focusable, ok := vnode.(rtui.FocusableVNode); ok {
			if vnode.Type() == rtui.VNodeElement {
				if tag, ok := vnode.(interface{ Tag() string }); ok && tag.Tag() == "checkbox" {
					return focusable.IsFocusable()
				}
			}
		}
		return false
	})
}

// GetSelects returns all select VNodes in the tree
func (n *DeclarativeNode) GetSelects() []rtui.FocusableVNode {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return collectByType(n.root, func(vnode rtui.VNode) bool {
		if focusable, ok := vnode.(rtui.FocusableVNode); ok {
			if vnode.Type() == rtui.VNodeElement {
				if tag, ok := vnode.(interface{ Tag() string }); ok && tag.Tag() == "select" {
					return focusable.IsFocusable()
				}
			}
		}
		return false
	})
}

// =============================================================================
// Utility functions
// =============================================================================

// collectByType collects VNodes that match a predicate function
func collectByType(root rtui.VNode, predicate func(rtui.VNode) bool) []rtui.FocusableVNode {
	var result []rtui.FocusableVNode

	if root == nil {
		return result
	}

	// Check current node
	if focusable, ok := root.(rtui.FocusableVNode); ok {
		if predicate(root) {
			result = append(result, focusable)
		}
	}

	// Recursively check children
	for _, child := range root.Children() {
		childResult := collectByType(child, predicate)
		result = append(result, childResult...)
	}

	return result
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// =============================================================================
// Fiber Reconciler Integration
// =============================================================================
// These functions create and configure the Fiber reconciler.

// fiberReconcilerAdapter adapts internal/reconciler.Reconciler to rtui.Reconciler interface
type fiberReconcilerAdapter struct {
	r *reconciler.Reconciler
}

// Render executes the rendering process (adapter method with interface{} parameters)
func (a *fiberReconcilerAdapter) Render(ctx interface{}, buffer interface{}, renderFunc func() rtui.VNode) {
	// Type assert to concrete types
	paintCtx, ok := ctx.(component.PaintContext)
	paintBuffer, ok := buffer.(*paint.Buffer)
	if !ok || paintBuffer == nil {
		return
	}

	// Call the actual reconciler's Render method
	a.r.Render(paintCtx, paintBuffer, renderFunc)
}

// SetApp sets the framework app (adapter method)
func (a *fiberReconcilerAdapter) SetApp(app interface{}) {
	if fwApp, ok := app.(*framework.App); ok {
		a.r.SetApp(fwApp)
	}
}

// SetFocusManager sets the focus manager (adapter method)
func (a *fiberReconcilerAdapter) SetFocusManager(mgr *rtui.VNodeFocusManager) {
	a.r.SetFocusManager(mgr)
}

// GetRenderedRoot returns the rendered VNode tree (adapter method)
func (a *fiberReconcilerAdapter) GetRenderedRoot() rtui.VNode {
	return a.r.GetRenderedRoot()
}

// newFiberReconciler creates a new Fiber reconciler for the given app and render function
func newFiberReconciler(fwApp *framework.App, fn rtui.ComponentFunc) rtui.Reconciler {
	// Create the actual reconciler from internal/reconciler
	r := reconciler.NewReconciler(fwApp, fn, reconciler.ReconcilerConfig{
		EnableFiber: true,
	})

	// Set up the render callback to actually render VNodes to the buffer
	// This callback is called by renderFiber for each non-component VNode
	r.SetRenderCallback(func(vnode rtui.VNode, x, y int, buffer *paint.Buffer) {
		renderVNodeToBuffer(vnode, x, y, buffer)
	})

	// Wrap in adapter to satisfy rtui.Reconciler interface
	return &fiberReconcilerAdapter{r: r}
}

// renderVNodeToBuffer renders a single VNode to the buffer
// This is used as the render callback for the Fiber reconciler
func renderVNodeToBuffer(vnode rtui.VNode, x, y int, buffer *paint.Buffer) {
	if vnode == nil {
		return
	}

	// Check if vnode implements Paintable interface (custom rendering)
	if paintable, ok := vnode.(interface{ Paint(int, int) []paint.DrawCmd }); ok {
		if os.Getenv("TUI_DEBUG_FOCUS") == "true" {
			// Get label for debug
			label := "?"
			if l, ok := vnode.(interface{ Label() string }); ok {
				label = l.Label()
			}
			fmt.Fprintf(os.Stderr, "[renderVNodeToBuffer] BEFORE Paint, label=%q, x=%d, y=%d\n", label, x, y)
		}
		commands := paintable.Paint(x, y)
		for _, cmd := range commands {
			if os.Getenv("TUI_DEBUG_FOCUS") == "true" {
				fmt.Fprintf(os.Stderr, "[renderVNodeToBuffer] AFTER Paint, x=%d, y=%d, text=%q\n", cmd.X, cmd.Y, cmd.Text)
			}
			buffer.SetString(cmd.X, cmd.Y, cmd.Text, cmd.Style)
		}
		return
	}

	// Handle built-in VNode types
	switch vnode.Type() {
	case rtui.VNodeText:
		if text := rtui.GetTextContent(vnode); text != "" {
			buffer.SetString(x, y, text, vnode.Style())
		}

	case rtui.VNodeElement:
		// Check if element has text content
		if content := rtui.GetTextContent(vnode); content != "" {
			buffer.SetString(x, y, content, vnode.Style())
			return
		}
		// Elements that don't have text content are containers (buttons, etc.)
		// They should implement Paintable interface for custom rendering
	}
}
