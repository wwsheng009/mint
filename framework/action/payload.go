// Package action provides semantic Action types and Payload structures.
// Payload must be immutable, value types for concurrent safety.
package action

// =============================================================================
// Payload Structures (Immutable, Value Types)
// =============================================================================
// Design principles (from fiber_action_payload.md):
// - Payload must be immutable
// - Must not reference mutable shared state
// - Recommended types: int, string, struct (value)
// - Forbidden: pointers
// =============================================================================

// ClickPayload carries mouse click event data.
type ClickPayload struct {
	X      int // Global X coordinate
	Y      int // Global Y coordinate
	LocalX int // Local X coordinate relative to target
	LocalY int // Local Y coordinate relative to target
	Button int // Mouse button (0=left, 1=middle, 2=right)
}

// InputPayload carries text input data.
type InputPayload struct {
	Value string // The input value
}

// ChangePayload carries value change data (generic).
type ChangePayload struct {
	Value interface{} // The changed value (string, bool, int, etc.)
}

// KeyPayload carries keyboard event data.
type KeyPayload struct {
	Rune    rune // The character (if printable)
	Special int  // Special key code
	Shift   bool
	Ctrl    bool
	Alt     bool
}

// FocusPayload carries focus event data.
type FocusPayload struct {
	FromID string // Previous focus target ID
	ToID   string // New focus target ID
}

// SubmitPayload carries form submit data.
type SubmitPayload struct {
	FormID string
	Values map[string]interface{}
}

// NavigatePayload carries navigation event data.
type NavigatePayload struct {
	Direction string // "next", "prev", "up", "down", "left", "right"
	Step      int    // Number of steps (for page navigation)
}

// =============================================================================
// Payload Helper Functions
// =============================================================================

// NewClickPayload creates a ClickPayload.
func NewClickPayload(x, y, localX, localY, button int) ClickPayload {
	return ClickPayload{
		X:      x,
		Y:      y,
		LocalX: localX,
		LocalY: localY,
		Button: button,
	}
}

// NewInputPayload creates an InputPayload.
func NewInputPayload(value string) InputPayload {
	return InputPayload{Value: value}
}

// NewChangePayload creates a ChangePayload.
func NewChangePayload(value interface{}) ChangePayload {
	return ChangePayload{Value: value}
}

// NewKeyPayload creates a KeyPayload.
func NewKeyPayload(rune rune, special int, shift, ctrl, alt bool) KeyPayload {
	return KeyPayload{
		Rune:    rune,
		Special: special,
		Shift:   shift,
		Ctrl:    ctrl,
		Alt:     alt,
	}
}
