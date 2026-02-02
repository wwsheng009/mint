// Package render provides declarative node implementation for bridging VNode and framework Component.
package render

import (
	"fmt"
	"os"
	"sync"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/framework/component"
	frameworkevent "github.com/wwsheng009/mint/framework/event"
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

	// Fiber reconciler integration
	reconciler rtui.Reconciler // Fiber reconciler (if enabled) - use interface to avoid import cycle
	useFiber   bool            // Whether Fiber mode is enabled
}

// NewDeclarativeNode creates a new declarative node from a VNode
func NewDeclarativeNode(vnode rtui.VNode) *DeclarativeNode {
	return &DeclarativeNode{
		root: vnode,
	}
}

// NewDeclarativeNodeFromFunc creates a new declarative node from a render function
func NewDeclarativeNodeFromFunc(fn rtui.ComponentFunc) *DeclarativeNode {
	return &DeclarativeNode{
		renderFn: fn,
		instance: rtui.NewComponentContextForRoot(),
		focusMgr: rtui.NewVNodeFocusManager(),
		useFiber: false, // Default to non-Fiber mode
	}
}

// NewDeclarativeNodeFromFuncWithFiber creates a new declarative node with Fiber reconciler enabled
// This function is called from ui.Run when MINT_USE_FIBER is set
func NewDeclarativeNodeFromFuncWithFiber(fn rtui.ComponentFunc, fwApp *framework.App) *DeclarativeNode {
	// Import the reconciler package here to avoid import cycles in ui package
	// This is safe because internal/render can import internal/reconciler
	r := newFiberReconciler(fwApp, fn)
	return &DeclarativeNode{
		renderFn:  fn,
		instance:  rtui.NewComponentContextForRoot(),
		focusMgr:  rtui.NewVNodeFocusManager(),
		reconciler: r,
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
			ctx.X, ctx.Y, buf.Width, buf.Height, n.useFiber)
	}

	// If Fiber reconciler is enabled, use it for rendering
	if n.useFiber && n.reconciler != nil {
		// The reconciler handles the entire render cycle
		// We need to call it with the render function
		n.reconciler.Render(ctx, buf, func() rtui.VNode {
			if n.renderFn != nil {
				return n.renderFn()
			}
			return n.root
		})

		// After reconciler renders, update root for event handling
		if n.renderFn != nil {
			n.root = n.renderFn()
		}

		// Collect focusable nodes
		if n.focusMgr != nil && n.root != nil {
			focusable := rtui.CollectFocusable(n.root)
			n.focusMgr.SetFocusable(focusable)
		}
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

		// Collect focusable nodes and update focus manager
		if n.focusMgr != nil && n.root != nil {
			focusable := rtui.CollectFocusable(n.root)
			n.focusMgr.SetFocusable(focusable)
			if os.Getenv("TUI_DEBUG_UI") == "true" {
				fmt.Fprintf(os.Stderr, "DeclarativeNode.Paint: collected %d focusable nodes\n", len(focusable))
			}
		}

		// Clear current context
		rtui.SetCurrentContext(nil)
	}

	if n.root == nil {
		if os.Getenv("TUI_DEBUG_UI") == "true" {
			fmt.Fprintf(os.Stderr, "DeclarativeNode.Paint: root is nil, returning\n")
		}
		return
	}

	// Walk the VNode tree and paint each node
	n.paintVNode(n.root, ctx.X, ctx.Y, buf)

	if os.Getenv("TUI_DEBUG_UI") == "true" {
		fmt.Fprintf(os.Stderr, "DeclarativeNode.Paint: painting complete\n")
	}
}

// paintVNode recursively paints a VNode and its children
//
// Rendering strategy:
// 1. If node implements Paintable: use Paint(x, y) → []DrawCmd → write to buffer
// 2. Otherwise: handle by node type (Text/Element/Fragment)
// 3. Always process children for container nodes
func (n *DeclarativeNode) paintVNode(vnode rtui.VNode, x, y int, buf *paint.Buffer) {
	if vnode == nil {
		return
	}

	// Debug logging
	if os.Getenv("TUI_DEBUG_UI") == "true" {
		fmt.Fprintf(os.Stderr, "[paintVNode] vnode type=%d (%s), x=%d, y=%d, actual type=%T\n",
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
			fmt.Fprintf(os.Stderr, "[paintVNode] Setting bounds: x=%d, y=%d, w=%d, h=%d\n", x, y, width, height)
		}
		// Set bounds for hit testing
		boundsAware.SetBounds(x, y, width, height)
	} else {
		if os.Getenv("TUI_DEBUG_UI") == "true" {
			fmt.Fprintf(os.Stderr, "[paintVNode] vnode does not implement SetBounds\n")
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
		// Component nodes - should implement Paintable, but if not, skip
		// (they should handle their own rendering)

	case rtui.VNodeFragment:
		// Fragment - just paint children, no self-rendering
		n.paintChildren(vnode, x, y, buf)
	}

	// For non-Paintable elements, paint children after self-rendering
	if vnode.Type() == rtui.VNodeElement {
		children := vnode.Children()
		if len(children) > 0 {
			childY := y
			for _, child := range children {
				n.paintVNode(child, x, childY, buf)
				childY++
			}
		}
	}
}

// paintText paints a text VNode
func (n *DeclarativeNode) paintText(vnode rtui.VNode, x, y int, buf *paint.Buffer) {
	text := ""
	// Get text content from props
	if props := vnode.Props(); props != nil {
		text = props.GetString("content")
	}

	if text != "" {
		buf.SetString(x, y, text, vnode.Style())
	}
}

// paintElement paints an element VNode
func (n *DeclarativeNode) paintElement(vnode rtui.VNode, x, y int, buf *paint.Buffer) {
	// Check if element has text content (for text elements created with ui.Text)
	if props := vnode.Props(); props != nil {
		if content := props.GetString("content"); content != "" {
			buf.SetString(x, y, content, vnode.Style())
			return // Don't paint children for text elements
		}
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
		n.paintVNode(child, x, childY, buf)
		childY++
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
	n.mu.RUnlock()

	if root == nil {
		return false
	}

	// 1. Let focus manager handle navigation (Tab, Shift+Tab)
	if focusMgr != nil {
		handled, _ := focusMgr.HandleEvent(ev)
		if handled {
			if os.Getenv("TUI_DEBUG_UI") == "true" {
				fmt.Fprintf(os.Stderr, "DeclarativeNode.HandleEvent: focus manager handled event\n")
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

// GetButtons returns all button VNodes in the tree
func (n *DeclarativeNode) GetButtons() []rtui.FocusableVNode {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return collectByType(n.root, func(vnode rtui.VNode) bool {
		if focusable, ok := vnode.(rtui.FocusableVNode); ok {
			// Check if this is a button by its tag
			if vnode.Type() == rtui.VNodeElement {
				if tag, ok := vnode.(interface{ Tag() string }); ok && tag.Tag() == "button" {
					return focusable.IsFocusable()
				}
			}
		}
		return false
	})
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
// This is separated to avoid import cycles - ui package calls this via NewDeclarativeNodeFromFuncWithFiber.

// newFiberReconciler creates a new Fiber reconciler for the given app and render function
// TODO: Implement actual Fiber reconciler integration. Currently a stub to avoid import cycles.
// The actual reconciler lives in internal/reconciler but importing it would create a cycle:
//   internal/render → internal/reconciler → ui → internal/render
func newFiberReconciler(fwApp *framework.App, fn rtui.ComponentFunc) rtui.Reconciler {
	// Stub implementation - returns nil for now
	// In the future, this will create and configure the Fiber reconciler
	return nil
}
