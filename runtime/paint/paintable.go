package paint

import (
	"github.com/wwsheng009/mint/runtime/style"
)

// =============================================================================
// Paintable - Component Self-Rendering Interface
// =============================================================================
// This interface allows components to implement their own rendering logic
// by generating draw commands that the renderer will execute.
//
// This follows the "GPU-like" rendering model where components generate
// commands rather than directly manipulating pixels.
//
// Architecture:
//   components/ (implement Paintable) → generates DrawCmd → renderer executes
//
// This avoids circular dependencies:
//   - ui/ and components/ import runtime/paint (for Paintable interface)
//   - runtime/ stores components as interface{} (no import needed)
//   - reconciler/ calls Paint() via type assertion
// =============================================================================

// Paintable is implemented by components that can generate their own draw commands.
//
// Components implementing this interface can control their own rendering
// by returning a list of DrawCmd that will be executed by the renderer.
//
// Example:
//   func (t *TextVNode) Paint(x, y int) []DrawCmd {
//       return []DrawCmd{{
//           X:     x,
//           Y:     y,
//           Text:  t.content,
//           Style: t.style,
//       }}
//   }
type Paintable interface {
	// Paint generates draw commands for rendering this component.
	//
	// Parameters:
	//   x, y - The position where this component should be rendered
	//         (calculated by the layout engine)
	//
	// Returns:
	//   []DrawCmd - List of draw commands to execute
	//
	// The renderer will execute these commands in order to draw the component.
	Paint(x, y int) []DrawCmd
}

// =============================================================================
// Extended Paintable with Advanced Features
// =============================================================================

// MeasurablePaintable combines measuring and painting capabilities.
//
// Components that can both calculate their preferred size and generate
// draw commands should implement this interface.
//
// This is the recommended interface for custom components as it enables
// proper layout calculation before rendering.
type MeasurablePaintable interface {
	// Measurable is embedded for size calculation
	// (defined in runtime/measurable.go, avoiding circular import)
	// Measure(constraints BoxConstraints) Size

	// Paintable for rendering
	Paintable
}

// =============================================================================
// DrawCmd Types (for future extensibility)
// =============================================================================

// CmdType represents the type of a draw command.
type CmdType int

const (
	// CmdText draws text at a position
	CmdText CmdType = iota
	// CmdFill fills a rectangular area
	CmdFill
	// CmdBox draws a box/border
	CmdBox
	// CmdCustom represents a custom drawing operation
	CmdCustom
)

// Type returns the type of this draw command.
func (c DrawCmd) Type() CmdType {
	return CmdText
}

// =============================================================================
// Helper Functions for Component Implementation
// =============================================================================

// NewTextCmd creates a text draw command.
func NewTextCmd(x, y int, text string, st style.Style) DrawCmd {
	return DrawCmd{
		X:     x,
		Y:     y,
		Text:  text,
		Style: st,
	}
}

// NewFillCmd creates a fill draw command.
func NewFillCmd(x, y, w, h int, char rune, st style.Style) DrawCmd {
	return DrawCmd{
		X:     x,
		Y:     y,
		Text:  string(char),
		Style: st,
	}
}

// NewBoxCmd creates a box draw command.
func NewBoxCmd(x, y, w, h int, st style.Style) DrawCmd {
	return DrawCmd{
		X:     x,
		Y:     y,
		Text:  "", // Box drawing handled by renderer
		Style: st,
	}
}
