// Package render provides VNode rendering implementations.
package render

import (
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Common Utilities
// =============================================================================

// getBuffer safely extracts a paint.Buffer from an interface{}
// Returns nil if the buffer is invalid or wrong type
func getBuffer(buffer interface{}) *paint.Buffer {
	if buf, ok := buffer.(*paint.Buffer); ok && buf != nil {
		return buf
	}
	return nil
}

// measureExplicitDimensions checks for explicit width/height in props
// Returns (width, height) with 0 meaning not explicitly set
func measureExplicitDimensions(vnode rtui.VNode) (width, height int) {
	if vnode == nil {
		return 0, 0
	}
	props := vnode.Props()
	if props == nil {
		return 0, 0
	}
	return props.GetInt("width"), props.GetInt("height")
}

// =============================================================================
// NonFiberRenderer
// =============================================================================

// NonFiberRenderer implements VNodeRenderer for traditional (non-Fiber) rendering.
// It walks the VNode tree directly and renders each node to the buffer.
type NonFiberRenderer struct {
	// owner is the DeclarativeNode that owns this renderer
	// This allows access to methods like measureVNodeWidth
	owner *DeclarativeNode
}

// NewNonFiberRenderer creates a new NonFiberRenderer.
func NewNonFiberRenderer(owner *DeclarativeNode) *NonFiberRenderer {
	return &NonFiberRenderer{owner: owner}
}

// Render renders a VNode tree to a buffer at the specified position.
// This implements the VNodeRenderer interface.
func (r *NonFiberRenderer) Render(vnode rtui.VNode, x, y int, buffer interface{}) {
	buf := getBuffer(buffer)
	if buf == nil {
		return
	}
	r.owner.PaintVNode(vnode, x, y, buf)
}

// Measure returns the width and height of a VNode.
// This implements the VNodeRenderer interface.
func (r *NonFiberRenderer) Measure(vnode rtui.VNode) (width, height int) {
	if vnode == nil {
		return 0, 0
	}

	// Check if VNode implements Measurable interface (Phase 4 improvement)
	type measurable interface {
		Measure(constraints runtime.BoxConstraints) runtime.Size
	}
	if m, ok := vnode.(measurable); ok {
		size := m.Measure(runtime.BoxConstraints{
			MinWidth:  0,
			MaxWidth:  1000,
			MinHeight: 0,
			MaxHeight: 1000,
		})
		return size.Width, size.Height
	}

	// Fallback: Get the width using the existing measurement logic
	width = r.owner.MeasureVNodeWidth(vnode)

	// Height is typically 1 for leaf nodes, calculated for containers
	height = r.measureVNodeHeight(vnode)

	return width, height
}

// measureVNodeHeight measures the height of a VNode.
func (r *NonFiberRenderer) measureVNodeHeight(vnode rtui.VNode) int {
	if vnode == nil {
		return 0
	}

	// Check for explicit height in props
	if props := vnode.Props(); props != nil {
		if h := props.GetInt("height"); h > 0 {
			return h
		}
	}

	// For containers, count children
	switch vnode.Type() {
	case rtui.VNodeText, rtui.VNodeElement:
		// Leaf nodes are typically 1 line
		return 1

	case rtui.VNodeFragment:
		// Fragments span their children
		children := vnode.Children()
		height := 0
		for _, child := range children {
			height += r.measureVNodeHeight(child)
		}
		return height

	default:
		return 1
	}
}

// =============================================================================
// FiberRenderer - Implements VNodeRenderer for Fiber mode
// =============================================================================

// FiberRenderer implements VNodeRenderer for Fiber-based rendering.
// It renders through the Fiber reconciler's render callback.
type FiberRenderer struct {
	// renderCallback is the function that actually renders VNodes to the buffer
	renderCallback func(rtui.VNode, int, int, *paint.Buffer)
}

// NewFiberRenderer creates a new FiberRenderer with the given render callback.
func NewFiberRenderer(renderCallback func(rtui.VNode, int, int, *paint.Buffer)) *FiberRenderer {
	return &FiberRenderer{
		renderCallback: renderCallback,
	}
}

// Render renders a VNode to a buffer at the specified position.
// This implements the VNodeRenderer interface.
func (r *FiberRenderer) Render(vnode rtui.VNode, x, y int, buffer interface{}) {
	buf := getBuffer(buffer)
	if buf == nil {
		return
	}
	if r.renderCallback != nil {
		r.renderCallback(vnode, x, y, buf)
	}
}

// Measure returns the width and height of a VNode.
// This implements the VNodeRenderer interface.
func (r *FiberRenderer) Measure(vnode rtui.VNode) (width, height int) {
	if vnode == nil {
		return 0, 0
	}

	// Check for explicit dimensions first
	w, h := measureExplicitDimensions(vnode)
	if w > 0 {
		width = w
	}
	if h > 0 {
		height = h
	}

	// Check if VNode implements Measurable interface (Phase 4 improvement)
	type measurable interface {
		Measure(constraints runtime.BoxConstraints) runtime.Size
	}
	if m, ok := vnode.(measurable); ok {
		// Use the component's Measure implementation
		// Provide loose constraints for measurement
		size := m.Measure(runtime.BoxConstraints{
			MinWidth:  0,
			MaxWidth: 1000, // Large default max width
			MinHeight: 0,
			MaxHeight: 1000,
		})
		return size.Width, size.Height
	}

	// Fallback: For text content, measure the text length
	if width == 0 {
		if text := rtui.GetTextContent(vnode); text != "" {
			width = len(text)
			height = 1
		}
	}

	// For buttons, measure label + brackets
	if width == 0 {
		if btn, ok := vnode.(interface{ Label() string }); ok {
			width = len(btn.Label()) + 2
			height = 1
		}
	}

	// Default dimensions
	if width == 0 {
		width = 1
	}
	if height == 0 {
		height = 1
	}

	return width, height
}
