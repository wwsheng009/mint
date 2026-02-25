// Package action provides semantic action types and payload structures.
//
// Payload Design Principles:
// - Payload must be immutable (value types)
// - Must not reference mutable shared state
// - Recommended: int, string, struct (value)
// - Forbidden: pointers to mutable data
package action

// =============================================================================
// Immutable Payload Structures
// =============================================================================

// ClickPayload carries mouse click event data
type ClickPayload struct {
	X      int // Global X coordinate
	Y      int // Global Y coordinate
	LocalX int // Local X coordinate relative to target
	LocalY int // Local Y coordinate relative to target
	Button int // Mouse button (0=left, 1=middle, 2=right)
}

// InputPayload carries text input data
type InputPayload struct {
	Value string // The input value
}

// ChangePayload carries value change data (generic)
type ChangePayload struct {
	Value interface{} // The changed value (string, bool, int, etc.)
}

// KeyPayload carries keyboard event data
type KeyPayload struct {
	Rune    rune   // The character (if printable)
	Special int    // Special key code
	Shift   bool   // Shift modifier
	Ctrl    bool   // Ctrl modifier
	Alt     bool   // Alt modifier
}

// FocusPayload carries focus event data
type FocusPayload struct {
	FromID string // Previous focus target ID
	ToID   string // New focus target ID
}

// SubmitPayload carries form submit data
type SubmitPayload struct {
	FormID string               // Form identifier
	Values map[string]interface{} // Form field values
}

// NavigatePayload carries navigation event data
type NavigatePayload struct {
	Direction string // "next", "prev", "up", "down", "left", "right", "first", "last"
	Step      int    // Number of steps (for page navigation)
}

// ResizePayload carries window resize data
type ResizePayload struct {
	W int // New width
	H int // New height
}

// SelectPayload carries selection event data
type SelectPayload struct {
	Index  int      // Selected index
	ID     string   // Selected item ID
	Value  string   // Selected item value
	Values []string // Multiple selected values (for multi-select)
}

// =============================================================================
// Payload Helper Functions
// =============================================================================

// NewClickPayload creates a ClickPayload
func NewClickPayload(x, y, localX, localY, button int) ClickPayload {
	return ClickPayload{
		X:      x,
		Y:      y,
		LocalX: localX,
		LocalY: localY,
		Button: button,
	}
}

// NewInputPayload creates an InputPayload
func NewInputPayload(value string) InputPayload {
	return InputPayload{Value: value}
}

// NewChangePayload creates a ChangePayload
func NewChangePayload(value interface{}) ChangePayload {
	return ChangePayload{Value: value}
}

// NewKeyPayload creates a KeyPayload
func NewKeyPayload(rune rune, special int, shift, ctrl, alt bool) KeyPayload {
	return KeyPayload{
		Rune:    rune,
		Special: special,
		Shift:   shift,
		Ctrl:    ctrl,
		Alt:     alt,
	}
}

// NewFocusPayload creates a FocusPayload
func NewFocusPayload(fromID, toID string) FocusPayload {
	return FocusPayload{
		FromID: fromID,
		ToID:   toID,
	}
}

// NewSubmitPayload creates a SubmitPayload
func NewSubmitPayload(formID string, values map[string]interface{}) SubmitPayload {
	return SubmitPayload{
		FormID: formID,
		Values: values,
	}
}

// NewNavigatePayload creates a NavigatePayload
func NewNavigatePayload(direction string, step int) NavigatePayload {
	return NavigatePayload{
		Direction: direction,
		Step:      step,
	}
}

// NewResizePayload creates a ResizePayload
func NewResizePayload(w, h int) ResizePayload {
	return ResizePayload{
		W: w,
		H: h,
	}
}

// NewSelectPayload creates a SelectPayload
func NewSelectPayload(index int, id, value string, values []string) SelectPayload {
	return SelectPayload{
		Index:  index,
		ID:     id,
		Value:  value,
		Values: values,
	}
}

// =============================================================================
// Action Payload Conversion Methods
// =============================================================================

// AsClickPayload returns payload as ClickPayload
func (a *Action) AsClickPayload() (*ClickPayload, bool) {
	if p, ok := a.Payload.(ClickPayload); ok {
		return &p, true
	}
	return nil, false
}

// AsInputPayload returns payload as InputPayload
func (a *Action) AsInputPayload() (*InputPayload, bool) {
	if p, ok := a.Payload.(InputPayload); ok {
		return &p, true
	}
	return nil, false
}

// AsChangePayload returns payload as ChangePayload
func (a *Action) AsChangePayload() (*ChangePayload, bool) {
	if p, ok := a.Payload.(ChangePayload); ok {
		return &p, true
	}
	return nil, false
}

// AsKeyPayload returns payload as KeyPayload
func (a *Action) AsKeyPayload() (*KeyPayload, bool) {
	if p, ok := a.Payload.(KeyPayload); ok {
		return &p, true
	}
	return nil, false
}

// AsFocusPayload returns payload as FocusPayload
func (a *Action) AsFocusPayload() (*FocusPayload, bool) {
	if p, ok := a.Payload.(FocusPayload); ok {
		return &p, true
	}
	return nil, false
}

// AsSubmitPayload returns payload as SubmitPayload
func (a *Action) AsSubmitPayload() (*SubmitPayload, bool) {
	if p, ok := a.Payload.(SubmitPayload); ok {
		return &p, true
	}
	return nil, false
}

// AsNavigatePayload returns payload as NavigatePayload
func (a *Action) AsNavigatePayload() (*NavigatePayload, bool) {
	if p, ok := a.Payload.(NavigatePayload); ok {
		return &p, true
	}
	return nil, false
}

// AsResizePayload returns payload as ResizePayload
func (a *Action) AsResizePayload() (*ResizePayload, bool) {
	if p, ok := a.Payload.(ResizePayload); ok {
		return &p, true
	}
	return nil, false
}

// AsSelectPayload returns payload as SelectPayload
func (a *Action) AsSelectPayload() (*SelectPayload, bool) {
	if p, ok := a.Payload.(SelectPayload); ok {
		return &p, true
	}
	return nil, false
}
