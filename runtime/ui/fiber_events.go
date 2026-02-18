package ui

import (
	"reflect"
)

// =============================================================================
// Fiber Event System (Fiber-first Architecture)
// =============================================================================
// In Fiber-first architecture:
// - VNode only declares "what I want" (intent)
// - Fiber holds "what I am now" (runtime state)
// - Events and focus are runtime state, so they belong to Fiber
//
// Design Principles:
// 1. Components implement Getter interfaces (GetOnClick, GetOnChange, etc.)
// 2. completeWork extracts handlers via interface detection
// 3. Clean separation: users call SetOnClick(), internal extraction is automatic
// =============================================================================

// =============================================================================
// EventHandlerProvider Interfaces (Clean API)
// =============================================================================
// Components implement these interfaces to provide event handlers.
// The reconciler detects these interfaces during completeWork and extracts handlers.

// ClickHandlerProvider for components with click event
type ClickHandlerProvider interface {
	GetOnClick() func()
}

// ChangeHandlerProvider for components with change event
type ChangeHandlerProvider interface {
	GetOnChange() interface{} // func(string), func(bool), etc.
}

// FocusHandlerProvider for components with focus events
type FocusHandlerProvider interface {
	GetOnFocus() func()
	GetOnBlur() func()
}

// MouseHandlerProvider for components with mouse events
type MouseHandlerProvider interface {
	GetOnMouseEnter() func()
	GetOnMouseLeave() func()
	GetOnMousePress() func()
	GetOnMouseRelease() func()
}

// KeyboardHandlerProvider for components with keyboard events
type KeyboardHandlerProvider interface {
	GetOnKeyDown() func(string)
	GetOnKeyUp() func(string)
}

// ActionHandlerProvider for Action system integration
type ActionHandlerProvider interface {
	GetOnAction() func(actionType string, payload interface{}) bool
}

// SubmitHandlerProvider for components with submit event
type SubmitHandlerProvider interface {
	GetOnSubmit() func()
}

// =============================================================================
// EventMap - Holds event handlers for a Fiber node
// =============================================================================

// EventMap holds event handlers for a Fiber node.
// This is populated during completeWork from VNode via interface detection.
type EventMap struct {
	// === Mouse events ===
	OnClick        func()
	OnMouseEnter   func()
	OnMouseLeave   func()
	OnMousePress   func()
	OnMouseRelease func()

	// === Keyboard events ===
	OnKeyDown  func(key string)
	OnKeyUp    func(key string)
	OnKeyPress func(key string)

	// === Focus events ===
	OnFocus func()
	OnBlur  func()

	// === Action System Integration ===
	OnAction func(actionType string, payload interface{}) bool

	// === Generic handlers (for extensibility) ===
	handlers map[string]interface{}
}

// FocusableMeta holds focusable metadata and runtime state for a Fiber node.
// This is the Fiber-first approach: all runtime state is in Fiber, not VNode.
type FocusableMeta struct {
	// Configuration (set during Fiber creation)
	TabIndex int
	Disabled bool
	FocusID  string

	// Runtime State (managed by FocusManager/HitTestManager)
	HasFocus  bool // Set by FocusManager during keyboard navigation
	IsHovered bool // Set by HitTestManager during mouse interaction
}

// IsFocusable returns true if the Fiber can receive focus
func (f *FocusableMeta) IsFocusable() bool {
	return f != nil && !f.Disabled && f.TabIndex >= 0
}

// SetFocus sets the focus state
func (f *FocusableMeta) SetFocus(focused bool) {
	if f != nil {
		f.HasFocus = focused
	}
}

// SetHover sets the hover state
func (f *FocusableMeta) SetHover(hovered bool) {
	if f != nil {
		f.IsHovered = hovered
	}
}

// NewEventMap creates a new EventMap
func NewEventMap() *EventMap {
	return &EventMap{
		handlers: make(map[string]interface{}),
	}
}

// HasHandlers returns true if the EventMap has any event handlers
func (e *EventMap) HasHandlers() bool {
	if e == nil {
		return false
	}
	return e.OnClick != nil ||
		e.OnMouseEnter != nil ||
		e.OnMouseLeave != nil ||
		e.OnMousePress != nil ||
		e.OnMouseRelease != nil ||
		e.OnKeyDown != nil ||
		e.OnKeyUp != nil ||
		e.OnKeyPress != nil ||
		e.OnFocus != nil ||
		e.OnBlur != nil ||
		e.OnAction != nil ||
		len(e.handlers) > 0
}

// SetHandler sets a generic event handler by name
func (e *EventMap) SetHandler(name string, handler interface{}) {
	if e.handlers == nil {
		e.handlers = make(map[string]interface{})
	}
	e.handlers[name] = handler
}

// GetHandler gets a generic event handler by name
func (e *EventMap) GetHandler(name string) interface{} {
	if e.handlers == nil {
		return nil
	}
	return e.handlers[name]
}

// DispatchClick triggers the OnClick handler
func (e *EventMap) DispatchClick() bool {
	if e == nil || e.OnClick == nil {
		return false
	}
	e.OnClick()
	return true
}

// DispatchKeyDown triggers the OnKeyDown handler
func (e *EventMap) DispatchKeyDown(key string) bool {
	if e == nil || e.OnKeyDown == nil {
		return false
	}
	e.OnKeyDown(key)
	return true
}

// DispatchFocus triggers the OnFocus handler
func (e *EventMap) DispatchFocus() {
	if e != nil && e.OnFocus != nil {
		e.OnFocus()
	}
}

// DispatchBlur triggers the OnBlur handler
func (e *EventMap) DispatchBlur() {
	if e != nil && e.OnBlur != nil {
		e.OnBlur()
	}
}

// DispatchAction dispatches an Action to the OnAction handler
func (e *EventMap) DispatchAction(actionType string, payload interface{}) bool {
	if e == nil || e.OnAction == nil {
		return false
	}
	return e.OnAction(actionType, payload)
}

// DispatchHandler dispatches an event to a named handler
func (e *EventMap) DispatchHandler(name string, arg interface{}) bool {
	if e == nil {
		return false
	}

	handler := e.GetHandler(name)
	if handler == nil {
		return false
	}

	switch h := handler.(type) {
	case func():
		h()
		return true
	case func(interface{}):
		h(arg)
		return true
	case func(string):
		if s, ok := arg.(string); ok {
			h(s)
			return true
		}
	case func(bool):
		if b, ok := arg.(bool); ok {
			h(b)
			return true
		}
	case func(int):
		if i, ok := arg.(int); ok {
			h(i)
			return true
		}
	default:
		v := reflect.ValueOf(handler)
		if v.Kind() == reflect.Func {
			t := v.Type()
			if t.NumIn() == 0 {
				v.Call(nil)
				return true
			} else if t.NumIn() == 1 {
				argV := reflect.ValueOf(arg)
				if argV.Type().ConvertibleTo(t.In(0)) {
					v.Call([]reflect.Value{argV.Convert(t.In(0))})
					return true
				}
			}
		}
	}
	return false
}

// =============================================================================
// Event Name Constants
// =============================================================================

const (
	EventClick        = "onClick"
	EventMouseEnter   = "onMouseEnter"
	EventMouseLeave   = "onMouseLeave"
	EventMousePress   = "onMousePress"
	EventMouseRelease = "onMouseRelease"
	EventDoubleClick  = "onDoubleClick"
	EventKeyDown      = "onKeyDown"
	EventKeyUp        = "onKeyUp"
	EventKeyPress     = "onKeyPress"
	EventFocus        = "onFocus"
	EventBlur         = "onBlur"
	EventChange       = "onChange"
	EventSubmit       = "onSubmit"
	EventScroll       = "onScroll"
)

// =============================================================================
// Action Type Constants (for OnAction handler)
// =============================================================================

const (
	ActionTypeNavigateNext     = "navigate_next"
	ActionTypeNavigatePrev     = "navigate_prev"
	ActionTypeNavigateUp       = "navigate_up"
	ActionTypeNavigateDown     = "navigate_down"
	ActionTypeNavigateLeft     = "navigate_left"
	ActionTypeNavigateRight    = "navigate_right"
	ActionTypeNavigatePageUp   = "navigate_page_up"
	ActionTypeNavigatePageDown = "navigate_page_down"
	ActionTypeSubmit           = "submit"
	ActionTypeCancel           = "cancel"
	ActionTypeEnter            = "enter"
	ActionTypeClick            = "click"
	ActionTypeDoubleClick      = "double_click"
	ActionTypeFocus            = "focus"
	ActionTypeBlur             = "blur"
	ActionTypeQuit             = "quit"
)

// ActionTypeToEventName maps Action types to event names
var ActionTypeToEventName = map[string]string{
	ActionTypeClick:       EventClick,
	ActionTypeEnter:       EventClick,
	ActionTypeDoubleClick: EventDoubleClick,
	ActionTypeFocus:       EventFocus,
	ActionTypeBlur:        EventBlur,
	ActionTypeSubmit:      EventSubmit,
}
