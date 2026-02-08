package inspector

import (
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Inspector provides UI inspection capabilities
type Inspector struct {
	enabled       bool
	selectedVNode rtui.VNode
	hoveredVNode  rtui.VNode
	mouseX        int
	mouseY        int
}

// NewInspector creates a new Inspector instance
func NewInspector() *Inspector {
	return &Inspector{
		enabled: false,
	}
}

// Enable enables the inspector
func (i *Inspector) Enable() {
	i.enabled = true
}

// Disable disables the inspector
func (i *Inspector) Disable() {
	i.enabled = false
	i.selectedVNode = nil
	i.hoveredVNode = nil
}

// IsEnabled returns whether the inspector is enabled
func (i *Inspector) IsEnabled() bool {
	return i.enabled
}

// SetSelectedVNode sets the currently selected VNode
func (i *Inspector) SetSelectedVNode(vnode rtui.VNode) {
	i.selectedVNode = vnode
}

// GetSelectedVNode returns the currently selected VNode
func (i *Inspector) GetSelectedVNode() rtui.VNode {
	return i.selectedVNode
}

// GetHoveredVNode returns the currently hovered VNode
func (i *Inspector) GetHoveredVNode() rtui.VNode {
	return i.hoveredVNode
}

// GetMousePosition returns the current mouse position
func (i *Inspector) GetMousePosition() (int, int) {
	return i.mouseX, i.mouseY
}

// HandleMouseEvent processes a mouse event at position (x, y)
// Returns true if the inspector handled the event
func (i *Inspector) HandleMouseEvent(x, y int) bool {
	if !i.enabled {
		return false
	}

	i.mouseX = x
	i.mouseY = y

	// TODO: Call FindVNodeAt when layout is available
	// For now, this is a placeholder for Phase 2
	// In the full implementation, this would:
	// 1. Get the ComputedLayout from the rendering pipeline
	// 2. Call FindVNodeAt(layout, x, y)
	// 3. Update i.hoveredVNode
	// 4. Extract and display element info

	return false
}

// HandleKeyEvent processes a keyboard event
// Returns true if the inspector handled the event
func (i *Inspector) HandleKeyEvent(event KeyEvent) bool {
	if !i.enabled {
		return false
	}

	// TODO: Implement keyboard shortcuts
	// F12 or Ctrl+I - Toggle inspector
	// Tab - Switch between elements
	// Enter - View details
	// Esc - Close inspector

	return false
}

// KeyEvent represents a keyboard event
type KeyEvent struct {
	Key      string
	Ctrl     bool
	Alt      bool
	Shift    bool
}

// FindVNodeAt finds the VNode at the given position (x, y)
// This is a simplified version that will be enhanced with layout integration
func (i *Inspector) FindVNodeAt(root rtui.VNode, x, y int) rtui.VNode {
	return findVNodeAtRecursive(root, x, y, 0)
}

// findVNodeAtRecursive recursively searches for a VNode at position (x, y)
func findVNodeAtRecursive(vnode rtui.VNode, x, y int, depth int) rtui.VNode {
	if vnode == nil {
		return nil
	}

	// Check if this VNode contains the point
	if vnodeContains(vnode, x, y) {
		// This node contains the point, check its children
		children := vnode.Children()
		for _, child := range children {
			result := findVNodeAtRecursive(child, x, y, depth+1)
			if result != nil {
				return result
			}
		}
		// No child contains the point, return this node
		return vnode
	}

	return nil
}

// vnodeContains checks if a VNode contains the point (x, y)
func vnodeContains(vnode rtui.VNode, x, y int) bool {
	// Try to get bounds
	if boundsAware, ok := vnode.(interface{ GetBounds() [4]int }); ok {
		bounds := boundsAware.GetBounds()
		// bounds = [x, y, width, height]
		vx, vy, vw, vh := bounds[0], bounds[1], bounds[2], bounds[3]

		// Check if point is within bounds
		return x >= vx && x < vx+vw && y >= vy && y < vy+vh
	}

	// Fallback: assume VNode doesn't contain the point
	return false
}

// GetSelectedInfo returns ElementInfo for the selected VNode
func (i *Inspector) GetSelectedInfo() ElementInfo {
	if i.selectedVNode == nil {
		return ElementInfo{}
	}
	return ExtractElementInfo(i.selectedVNode)
}

// GetHoveredInfo returns ElementInfo for the hovered VNode
func (i *Inspector) GetHoveredInfo() ElementInfo {
	if i.hoveredVNode == nil {
		return ElementInfo{}
	}
	return ExtractElementInfo(i.hoveredVNode)
}
