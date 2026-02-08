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
		// Still check for toggle shortcuts even when disabled
		if event.Key == "f12" || (event.Key == "i" && event.Ctrl) {
			i.Enable()
			return true
		}
		return false
	}

	// Handle keyboard shortcuts when enabled
	switch {
	case event.Key == "f12" || (event.Key == "i" && event.Ctrl):
		// Toggle inspector
		i.Disable()
		return true

	case event.Key == "escape":
		// Close inspector or clear selection
		if i.selectedVNode != nil {
			i.selectedVNode = nil
			return true
		}
		i.Disable()
		return true

	case event.Key == "tab":
		// Navigate to next element
		i.NavigateToNextElement()
		return true

	case event.Key == "enter":
		// View details (handled by caller via GetSelectedInfo())
		// For now, just return true to indicate we handled it
		return true
	}

	return false
}

// KeyEvent represents a keyboard event
type KeyEvent struct {
	Key   string
	Ctrl  bool
	Alt   bool
	Shift bool
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

// NavigateToNextElement navigates to the next selectable element in the tree
// This method is public for testing purposes and can be used for programmatic navigation
func (i *Inspector) NavigateToNextElement() {
	// This is a simplified implementation
	// TODO: In Phase 6, we'll implement proper tree traversal with visible elements only
	// For now, this is a placeholder that demonstrates the concept

	// If no element is selected, we need a root to start from
	// This will be provided when integrated with the rendering pipeline
	if i.selectedVNode == nil {
		// Can't navigate without a starting point
		return
	}

	// Try to find next element by traversing the tree
	nextElement := i.FindNextSelectable(i.selectedVNode)
	if nextElement != nil {
		i.selectedVNode = nextElement
	}
}

// FindNextSelectable finds the next selectable element after the given node
// This is a simplified implementation that will be enhanced in Phase 6
func (i *Inspector) FindNextSelectable(startNode rtui.VNode) rtui.VNode {
	// Get all selectable elements in the tree
	allElements := i.CollectAllElements(startNode)

	if len(allElements) == 0 {
		return nil
	}

	// Find current element index
	currentIndex := -1
	for idx, elem := range allElements {
		if elem == startNode {
			currentIndex = idx
			break
		}
	}

	// Move to next element (wrap around)
	if currentIndex >= 0 {
		nextIndex := (currentIndex + 1) % len(allElements)
		return allElements[nextIndex]
	}

	// If current element not found, return first element
	return allElements[0]
}

// CollectAllElements collects all selectable VNodes from a starting node
// This performs a breadth-first traversal to find interactive elements
// This method is public for testing and can be used for custom navigation logic
func (i *Inspector) CollectAllElements(root rtui.VNode) []rtui.VNode {
	var elements []rtui.VNode

	if root == nil {
		return elements
	}

	// BFS traversal
	queue := []rtui.VNode{root}
	visited := make(map[rtui.VNode]bool)

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current] {
			continue
		}
		visited[current] = true

			// Check if this element is selectable (interactive)
		if i.IsSelectable(current) {
			elements = append(elements, current)
		}

		// Add children to queue
		children := current.Children()
		queue = append(queue, children...)
	}

	return elements
}

// IsSelectable checks if a VNode is selectable (interactive)
// Selectable elements are those that users can interact with
// This method is public for testing and can be used for custom navigation logic
func (i *Inspector) IsSelectable(vnode rtui.VNode) bool {
	if vnode == nil {
		return false
	}

	// Check tag for interactive components
	if tagger, ok := vnode.(interface{ Tag() string }); ok {
		tag := tagger.Tag()
		switch tag {
		case "button", "input", "checkbox", "select", "textarea":
			return true
		}
	}

	// Check if it has click handlers or other interactions
	if props := vnode.Props(); props != nil {
		if _, hasOnClick := props["onClick"]; hasOnClick {
			return true
		}
		if _, hasOnChange := props["onChange"]; hasOnChange {
			return true
		}
	}

	return false
}
