package event

import (
	"os"

	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/runtime/event"
)

// EventHandler is the interface for event handlers.
type EventHandler interface {
	HandleEvent(Event) bool
}

// EventHandlerFunc is a function adapter for EventHandler.
type EventHandlerFunc func(Event) bool

func (f EventHandlerFunc) HandleEvent(ev Event) bool {
	return f(ev)
}

// Router routes events to appropriate handlers.
type Router struct {
	globalHandlers  map[EventType][]EventHandler
	captureHandlers []EventHandler
}

// NewRouter creates a new event router.
func NewRouter() *Router {
	return &Router{
		globalHandlers: make(map[EventType][]EventHandler),
	}
}

// Subscribe subscribes to events of the given type.
func (r *Router) Subscribe(eventType EventType, handler EventHandler) func() {
	r.globalHandlers[eventType] = append(r.globalHandlers[eventType], handler)
	return func() {
		r.Unsubscribe(eventType, handler)
	}
}

// Unsubscribe unsubscribes a handler from events.
func (r *Router) Unsubscribe(eventType EventType, handler EventHandler) {
	handlers := r.globalHandlers[eventType]
	for i, h := range handlers {
		if h == handler {
			r.globalHandlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}
}

// Route routes an event through the handler chain.
func (r *Router) Route(ev Event) {
	// Capture phase
	for _, handler := range r.captureHandlers {
		if handler.HandleEvent(ev) {
			return
		}
	}

	// Global handlers
	if handlers, ok := r.globalHandlers[ev.Type()]; ok {
		for _, handler := range handlers {
			if handler.HandleEvent(ev) {
				return
			}
		}
	}
}

// AddCaptureHandler adds a capture-phase handler.
func (r *Router) AddCaptureHandler(handler EventHandler) {
	r.captureHandlers = append(r.captureHandlers, handler)
}

// ============================================================================
// KeyMap for Keyboard Shortcuts
// ============================================================================

// KeyMap manages keyboard shortcut bindings.
type KeyMap struct {
	bindings map[string]EventHandler
}

// NewKeyMap creates a new key map.
func NewKeyMap() *KeyMap {
	return &KeyMap{
		bindings: make(map[string]EventHandler),
	}
}

// BindFunc binds a keyboard combo to a handler function.
func (k *KeyMap) BindFunc(combo string, handler func(*KeyEvent)) error {
	k.bindings[combo] = EventHandlerFunc(func(ev Event) bool {
		if keyEv, ok := ev.(*KeyEvent); ok {
			handler(keyEv)
			return true
		}
		return false
	})
	return nil
}

// Bind binds a keyboard combo to an event handler.
func (k *KeyMap) Bind(combo string, handler EventHandler) error {
	k.bindings[combo] = handler
	return nil
}

// Lookup looks up a handler for a keyboard event.
func (k *KeyMap) Lookup(ev *KeyEvent) (EventHandler, bool) {
	// Build combo string with modifiers
	combo := k.buildComboString(ev.Key, ev.Modifiers)

	// Debug output
	if os.Getenv("TUI_DEBUG_UI") == "true" || os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
		log.UILogger.Debug("[KeyMap] Lookup: Rune=%c Name=%q Modifiers=%d Combo=%q",
			ev.Key.Rune, ev.Key.Name, ev.Modifiers, combo)
	}

	// Try combo with modifiers first (e.g., "alt+k", "ctrl+d")
	if combo != "" {
		if handler, ok := k.bindings[combo]; ok {
			if os.Getenv("TUI_DEBUG_UI") == "true" || os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
				log.UILogger.Debug("[KeyMap] Found handler for combo '%s'", combo)
			}
			return handler, true
		}
	}

	// Try character key without modifiers
	if ev.Key.Rune > 0 {
		if handler, ok := k.bindings[string(ev.Key.Rune)]; ok {
			return handler, true
		}
	}

	// Try special key name without modifiers
	if ev.Key.Name != "" {
		if handler, ok := k.bindings[ev.Key.Name]; ok {
			return handler, true
		}
	}

	return nil, false
}

// buildComboString builds a combo string like "alt+k" or "ctrl+d" from key and modifiers
func (k *KeyMap) buildComboString(key Key, modifiers Modifier) string {
	if modifiers == 0 {
		return ""
	}

	var combo string

	// Add modifier prefixes
	if modifiers&ModAlt != 0 {
		combo += "alt+"
	}
	if modifiers&ModCtrl != 0 {
		combo += "ctrl+"
	}
	if modifiers&ModShift != 0 {
		combo += "shift+"
	}

	// Add key name or rune
	if key.Rune > 0 {
		combo += string(key.Rune)
	} else if key.Name != "" {
		combo += key.Name
	} else {
		return ""
	}

	return combo
}

// ============================================================================
// Event Type Wrappers
// ============================================================================

// MousePressEvent wraps runtime mouse press events.
type MousePressEvent struct {
	*BaseEvent
	X, Y   int
	Button MouseButton
}

// NewMousePressEvent creates a new mouse press event.
func NewMousePressEvent(x, y int, button MouseButton) *MousePressEvent {
	return &MousePressEvent{
		BaseEvent: NewBaseEvent(event.EventMousePress),
		X:         x,
		Y:         y,
		Button:    button,
	}
}

// ResizeEvent wraps runtime resize events.
type ResizeEvent struct {
	*BaseEvent
	OldWidth, OldHeight int
	NewWidth, NewHeight int
}

// NewResizeEvent creates a new resize event.
func NewResizeEvent(oldW, oldH, newW, newH int) *ResizeEvent {
	return &ResizeEvent{
		BaseEvent: NewBaseEvent(event.EventResize),
		OldWidth:  oldW,
		OldHeight: oldH,
		NewWidth:  newW,
		NewHeight: newH,
	}
}

// CloseEvent wraps close events (uses EventQuit from runtime).
type CloseEvent struct {
	*BaseEvent
}

// NewCloseEvent creates a new close event.
func NewCloseEvent() *CloseEvent {
	return &CloseEvent{
		BaseEvent: NewBaseEvent(event.EventQuit),
	}
}
