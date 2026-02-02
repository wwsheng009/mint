package ui

// FocusableVNode is an extension interface for VNodes that can receive focus.
// Components that implement this interface can participate in keyboard navigation
// and receive focused keyboard events (like Enter/Space).
type FocusableVNode interface {
	VNode

	// SetFocus sets the focus state of this component.
	// When hasFocus is true, the component should visually indicate focus
	// and prepare to receive keyboard input.
	SetFocus(hasFocus bool)

	// IsFocusable returns whether this component can currently receive focus.
	// Components may return false when disabled or hidden.
	IsFocusable() bool

	// GetFocusID returns a unique identifier for focus persistence.
	// This allows the focus manager to restore focus after re-renders.
	// If the component has a Key set, it should be used as the focus ID.
	GetFocusID() string

	// Label returns a text label for this focusable element.
	// Used for testing and debugging to identify elements.
	// Default implementation returns empty string.
	Label() string
}
