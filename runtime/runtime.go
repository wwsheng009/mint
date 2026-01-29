package runtime

import (
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
)

// Component is the base interface for all UI components.
//
// This is the minimal interface that all components must implement.
// The View() method returns the component's visual representation.
type Component interface {
	// View returns the component's visual representation as a string.
	// This is called during the Render phase.
	View() string
}

// Event is a placeholder for future event system
// v1: simplified, will be expanded in Phase 3
type Event struct {
	X, Y int
	Type string
	Data interface{}
}

// FocusableComponent is an interface for components that can receive focus.
// This is the minimal interface required for focus management.
type FocusableComponent interface {
	SetFocus(focus bool)
	IsFocusable() bool
}

// Rect represents a rectangle region.
type Rect struct {
	X      int
	Y      int
	Width  int
	Height int
}

// Runtime is the main interface for the Yao TUI Runtime.
//
// It provides a clean API for:
//   - Layout: Calculate geometry (measure + layout phases)
//   - Render: Generate frames from layout results
//   - Dispatch: Handle events (Phase 3)
//   - Focus: Manage keyboard navigation (Phase 3)
type Runtime interface {
	// Layout performs a complete layout pass on the root node.
	//
	// This includes:
	//   1. Measure phase: Calculate intrinsic sizes bottom-up
	//   2. Layout phase: Assign positions top-down
	//
	// The constraints (c) are the available space from the screen/window.
	//
	// Returns a LayoutResult containing all positioned nodes.
	Layout(root *LayoutNode, c BoxConstraints) LayoutResult

	// Render generates a Frame from a LayoutResult.
	//
	// This is the Render phase, which:
	//   - Creates a CellBuffer (virtual canvas)
	//   - Renders all nodes in Z-Index order
	//   - Returns a Frame that can be output to the terminal
	//
	// The resulting Frame.String() can be used to update the terminal.
	Render(result LayoutResult) Frame

	// Dispatch handles an input event (keyboard, mouse, etc.).
	//
	// v1: placeholder, will be implemented in Phase 3
	// For now, events are handled by existing Bubble Tea system.
	Dispatch(ev Event)

	// FocusNext moves focus to the next focusable component.
	//
	// v1: placeholder, will be implemented in Phase 3
	// For now, focus is handled by existing focus manager.
	FocusNext()
}

// Frame represents a rendered frame (virtual canvas).
//
// It contains the complete rendered output that can be sent to the terminal.
type Frame struct {
	Buffer *CellBuffer
	Width  int
	Height int
	Dirty  bool // True if this frame has changes from previous
}

// String returns the frame as a string for terminal output.
// This is the primary way to send a frame to Bubble Tea's View() method.
func (f Frame) String() string {
	if f.Buffer == nil {
		return ""
	}
	return f.Buffer.String()
}

// =============================================================================
// Type Aliases to paint.Buffer and paint.Cell
// =============================================================================

// CellBuffer is an alias to paint.Buffer.
// All buffer functionality is now provided by the paint package.
type CellBuffer = paint.Buffer

// Cell is an alias to paint.Cell.
// All cell functionality is now provided by the paint package.
type Cell = paint.Cell

// =============================================================================
// Backward Compatibility Functions
// These functions provide compatibility with existing code that uses
// runtime.CellBuffer. Methods are now directly on paint.Buffer.
// =============================================================================

// SetContentRuntime sets a cell at the given position using individual style parameters.
// This is a package-level function used by the render package adapter to avoid circular imports.
// It delegates to the paint.Buffer SetContent method with style construction.
func SetContentRuntime(b *CellBuffer, x, y, z int, char rune, bold, underline, italic bool, nodeID string) {
	w, h := b.Width, b.Height
	if x < 0 || x >= w || y < 0 || y >= h {
		return
	}

	// Check Z-Index
	if z < b.Cells[y][x].ZIndex {
		return
	}

	s := style.Style{}
	if bold {
		s = s.Bold(true)
	}
	if underline {
		s = s.Underline(true)
	}
	if italic {
		s = s.Italic(true)
	}

	b.Cells[y][x] = Cell{
		Cluster: string(char),
		Style:   s,
		ZIndex:  z,
		NodeID:  nodeID,
	}
}

// =============================================================================
// Legacy Types (kept for compatibility, will be deprecated)
// =============================================================================

// CellStyle is an alias to style.Style for backward compatibility.
// Use style.Style directly in new code.
type CellStyle = style.Style
