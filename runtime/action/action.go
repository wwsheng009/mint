// Package action provides unified action dispatching and routing.
//
// Unified Action System:
// - Single Action struct supporting both string Target and uint64 TargetID
// - Comprehensive ActionType constants (merged from both packages)
// - Three-phase routing (Capture → Target → Bubble)
// - Middleware system for cross-cutting concerns
// - Input processing with KeyMap
// - Composite actions for concurrent/sequential execution
// - Scope-based dispatchers for action isolation
//
// This package is the result of merging framework/action and runtime/action.
// All imports from framework/action should be changed to runtime/action.
package action

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/wwsheng009/mint/runtime/event"
)

var actionIDCounter uint64

// Action represents a semantic user action (not raw input events).
// All state changes should be traceable to an Action for debugging and replay.
type Action struct {
	// Type is the semantic type of the action
	Type ActionType

	// Payload carries action-specific data
	Payload interface{}

	// Source identifies where the action originated (e.g., "keyboard", "mouse")
	Source string

	// Target is the semantic component ID (e.g., "button.submit", "form.login")
	// This is the primary way to identify the target component.
	Target string

	// TargetID is the internal NodeID (uint64) for fast lookup
	// Automatically computed from Target when set.
	TargetID uint64

	// ===== Enhanced features (from framework/action) =====

	// ID is a unique identifier for this action (useful for tracking/debugging)
	ID uint64

	// Timestamp when the action was created
	Timestamp time.Time

	// Meta holds extensibility metadata
	Meta map[string]interface{}

	// stopped indicates whether action propagation should be stopped
	stopped bool
}

// ActionType represents a semantic action type.
// Naming convention: verb + noun (e.g., "navigate_next", "input_char")
type ActionType string

// ============================================================================
// Navigation Actions
// ============================================================================

const (
	ActionNavigateFirst   ActionType = "navigate_first"   // Jump to beginning
	ActionNavigateLast    ActionType = "navigate_last"    // Jump to end
	ActionNavigateNext    ActionType = "navigate_next"    // Next item
	ActionNavigatePrev    ActionType = "navigate_prev"    // Previous item
	ActionNavigateUp      ActionType = "navigate_up"      // Move up
	ActionNavigateDown    ActionType = "navigate_down"    // Move down
	ActionNavigateLeft    ActionType = "navigate_left"    // Move left
	ActionNavigateRight   ActionType = "navigate_right"   // Move right
	ActionNavigatePageUp   ActionType = "navigate_page_up"   // Page up
	ActionNavigatePageDown ActionType = "navigate_page_down" // Page down
	ActionNavigateHome     ActionType = "navigate_home"     // Jump to home
	ActionNavigateEnd      ActionType = "navigate_end"      // Jump to end
)

// ============================================================================
// Editing Actions
// ============================================================================

const (
	ActionInputChar        ActionType = "input_char"         // Input single character
	ActionInputText        ActionType = "input_text"         // Input text string
	ActionDeleteChar       ActionType = "delete_char"        // Delete character
	ActionDeleteWord       ActionType = "delete_word"        // Delete word
	ActionDeleteLine       ActionType = "delete_line"        // Delete line
	ActionBackspace        ActionType = "backspace"          // Backspace key
	ActionCursorHome       ActionType = "cursor_home"        // Cursor to line start
	ActionCursorEnd        ActionType = "cursor_end"         // Cursor to line end
	ActionCursorUp         ActionType = "cursor_up"          // Move cursor up
	ActionCursorDown       ActionType = "cursor_down"        // Move cursor down
	ActionCursorLeft       ActionType = "cursor_left"        // Move cursor left
	ActionCursorRight      ActionType = "cursor_right"       // Move cursor right
	ActionCursorWordLeft   ActionType = "cursor_word_left"   // Move cursor left by word
	ActionCursorWordRight  ActionType = "cursor_word_right"  // Move cursor right by word
	ActionSelectAll        ActionType = "select_all"         // Select all
	ActionSelectWord       ActionType = "select_word"        // Select word
	ActionSelectLine       ActionType = "select_line"        // Select line
	ActionEnter            ActionType = "enter"              // Enter key
)

// ============================================================================
// Form Actions
// ============================================================================

const (
	ActionSubmit    ActionType = "submit"    // Submit form
	ActionCancel    ActionType = "cancel"    // Cancel operation
	ActionValidate  ActionType = "validate"  // Validate input
	ActionReset     ActionType = "reset"     // Reset form
	ActionClear     ActionType = "clear"     // Clear input
)

// ============================================================================
// Selection Actions
// ============================================================================

const (
	ActionSelectItem      ActionType = "select_item"       // Select item
	ActionDeselectItem    ActionType = "deselect_item"     // Deselect item
	ActionToggleSelect    ActionType = "toggle_select"    // Toggle selection
	ActionSelectRange     ActionType = "select_range"      // Select range
	ActionSelect          ActionType = "select"            // Select current
	ActionToggle          ActionType = "toggle"            // Toggle state
	ActionExpand          ActionType = "expand"            // Expand
	ActionCollapse        ActionType = "collapse"          // Collapse
)

// ============================================================================
// Mouse Actions
// ============================================================================

const (
	ActionClick        ActionType = "click"         // Mouse click
	ActionDoubleClick   ActionType = "double_click"  // Double click
	ActionTripleClick   ActionType = "triple_click"  // Triple click
	ActionMousePress    ActionType = "mouse_press"   // Mouse press down
	ActionMouseRelease  ActionType = "mouse_release" // Mouse release up
	ActionMouseMotion   ActionType = "mouse_motion"  // Mouse move
	ActionMouseDrag     ActionType = "mouse_drag"    // Mouse drag
	ActionMouseWheel    ActionType = "mouse_wheel"   // Mouse wheel
	ActionMouseWheelUp  ActionType = "mouse_wheel_up"   // Wheel up
	ActionMouseWheelDown ActionType = "mouse_wheel_down" // Wheel down
	ActionRightClick    ActionType = "right_click"   // Right click
	ActionRightRelease  ActionType = "right_release" // Right mouse release
	ActionMiddleClick   ActionType = "middle_click"  // Middle click
	ActionMiddlePress   ActionType = "middle_press"  // Middle press
	ActionMiddleRelease ActionType = "middle_release"// Middle mouse release
	ActionHover         ActionType = "hover"         // Mouse hover
	ActionDragStart     ActionType = "drag_start"    // Start dragging
	ActionDragMove      ActionType = "drag_move"     // Dragging moved
	ActionDragEnd       ActionType = "drag_end"      // End dragging
)

// ============================================================================
// View Actions
// ============================================================================

const (
	ActionScroll      ActionType = "scroll"       // Scroll action
	ActionScrollUp    ActionType = "scroll_up"    // Scroll up
	ActionScrollDown  ActionType = "scroll_down"  // Scroll down
	ActionScrollLeft  ActionType = "scroll_left"  // Scroll left
	ActionScrollRight ActionType = "scroll_right" // Scroll right
	ActionZoomIn      ActionType = "zoom_in"      // Zoom in
	ActionZoomOut     ActionType = "zoom_out"     // Zoom out
	ActionZoomReset   ActionType = "zoom_reset"   // Reset zoom
	ActionResize      ActionType = "resize"       // Window resize (Payload: Size{W,H})
)

// ============================================================================
// Window Actions
// ============================================================================

const (
	ActionQuit       ActionType = "quit"       // Quit application
	ActionClose      ActionType = "close"      // Close window
	ActionMaximize   ActionType = "maximize"   // Maximize
	ActionMinimize   ActionType = "minimize"   // Minimize
	ActionFullscreen ActionType = "fullscreen" // Fullscreen
)

// ============================================================================
// System Actions
// ============================================================================

const (
	ActionCopy     ActionType = "copy"     // Copy
	ActionCut      ActionType = "cut"      // Cut
	ActionPaste    ActionType = "paste"    // Paste
	ActionUndo     ActionType = "undo"     // Undo
	ActionRedo     ActionType = "redo"     // Redo
	ActionSearch   ActionType = "search"   // Search
	ActionHelp     ActionType = "help"     // Help
	ActionRefresh  ActionType = "refresh"  // Refresh
	ActionInspect  ActionType = "inspect"  // Toggle inspector
	ActionFocus    ActionType = "focus"    // Set focus
	ActionBlur     ActionType = "blur"     // Lose focus
)

// ============================================================================
// Focus Actions
// ============================================================================

const (
	ActionFocusGained ActionType = "focus_gained" // Gained focus
	ActionFocusLost   ActionType = "focus_lost"   // Lost focus
	ActionFocusNext   ActionType = "focus_next"   // Move focus next
	ActionFocusPrev   ActionType = "focus_prev"   // Move focus prev
)

// ============================================================================
// UI Interaction Actions
// ============================================================================

const (
	// Mouse hover actions
	ActionMouseEnter ActionType = "mouse_enter" // Mouse entered component
	ActionMouseLeave ActionType = "mouse_leave" // Mouse left component

	// Press actions
	ActionPressStart ActionType = "press_start" // Press started (mouse down)
	ActionPressEnd   ActionType = "press_end"   // Press ended (mouse up)
	ActionPress      ActionType = "press"       // Generic press action

	// Release actions
	ActionRelease ActionType = "release" // Release action

	// Enable/Disable actions
	ActionEnable  ActionType = "enable"  // Enable component
	ActionDisable ActionType = "disable" // Disable component

	// Activation actions
	ActionActivate   ActionType = "activate"   // Activate component
	ActionDeactivate ActionType = "deactivate" // Deactivate component
)

// ============================================================================
// Data Actions
// ============================================================================

const (
	ActionDataLoad   ActionType = "data_load"   // Data loaded
	ActionDataUpdate ActionType = "data_update" // Data updated
	ActionDataError  ActionType = "data_error"  // Data error
)

// ============================================================================
// Composite Actions (init/mount/unmount)
// ============================================================================

const (
	ActionInit    ActionType = "init"    // Component initialization
	ActionMount   ActionType = "mount"   // Component mounted
	ActionUnmount ActionType = "unmount" // Component unmounted
)

// ============================================================================
// AI Actions (V3 from runtime/action)
// ============================================================================

const (
	ActionAIInspect  ActionType = "ai_inspect"  // AI inspect UI state
	ActionAIFind     ActionType = "ai_find"     // AI find component
	ActionAIQuery    ActionType = "ai_query"    // AI query state
	ActionAIDispatch ActionType = "ai_dispatch" // AI dispatch action
	ActionAIWait     ActionType = "ai_wait"     // AI wait for state
	ActionAIWatch    ActionType = "ai_watch"    // AI monitor state
)

// ============================================================================
// Action Classification Methods
// ============================================================================

// IsNavigation checks if this is a navigation action
func (a *Action) IsNavigation() bool {
	switch a.Type {
	case ActionNavigateFirst, ActionNavigateLast, ActionNavigateNext,
		ActionNavigatePrev, ActionNavigateUp, ActionNavigateDown,
		ActionNavigateLeft, ActionNavigateRight, ActionNavigatePageUp,
		ActionNavigatePageDown, ActionNavigateHome, ActionNavigateEnd,
		ActionFocusNext, ActionFocusPrev:
		return true
	}
	return false
}

// IsEditing checks if this is an editing action
func (a *Action) IsEditing() bool {
	switch a.Type {
	case ActionInputChar, ActionInputText, ActionDeleteChar,
		ActionDeleteWord, ActionDeleteLine, ActionBackspace,
		ActionCursorHome, ActionCursorEnd, ActionCursorUp,
		ActionCursorDown, ActionCursorLeft, ActionCursorRight,
		ActionCursorWordLeft, ActionCursorWordRight,
		ActionSelectAll, ActionSelectWord, ActionSelectLine, ActionEnter:
		return true
	}
	return false
}

// IsForm checks if this is a form action
func (a *Action) IsForm() bool {
	switch a.Type {
	case ActionSubmit, ActionCancel, ActionValidate, ActionReset, ActionClear:
		return true
	}
	return false
}

// IsSelection checks if this is a selection action
func (a *Action) IsSelection() bool {
	switch a.Type {
	case ActionSelectItem, ActionDeselectItem, ActionToggleSelect,
		ActionSelectRange, ActionSelect, ActionToggle:
		return true
	}
	return false
}

// IsMouse checks if this is a mouse action
func (a *Action) IsMouse() bool {
	switch a.Type {
	case ActionClick, ActionDoubleClick, ActionTripleClick,
		ActionMousePress, ActionMouseRelease, ActionMouseMotion,
		ActionMouseDrag, ActionMouseWheel, ActionMouseWheelUp,
		ActionMouseWheelDown, ActionRightClick, ActionMiddleClick,
		ActionMiddlePress, ActionHover, ActionDragStart, ActionDragMove,
		ActionDragEnd:
		return true
	}
	return false
}

// IsSystem checks if this is a system action
func (a *Action) IsSystem() bool {
	switch a.Type {
	case ActionQuit, ActionClose, ActionMaximize, ActionMinimize,
		ActionFullscreen, ActionCopy, ActionCut, ActionPaste,
		ActionUndo, ActionRedo, ActionSearch, ActionHelp,
		ActionRefresh, ActionInspect, ActionFocus, ActionBlur:
		return true
	}
	return false
}

// IsView checks if this is a view action
func (a *Action) IsView() bool {
	switch a.Type {
	case ActionScroll, ActionScrollUp, ActionScrollDown, ActionScrollLeft,
		ActionScrollRight, ActionZoomIn, ActionZoomOut, ActionZoomReset,
		ActionResize:
		return true
	}
	return false
}

// IsData checks if this is a data action
func (a *Action) IsData() bool {
	switch a.Type {
	case ActionDataLoad, ActionDataUpdate, ActionDataError:
		return true
	}
	return false
}

// IsUndoRedo checks if this is an undo/redo action
func (a *Action) IsUndoRedo() bool {
	return a.Type == ActionUndo || a.Type == ActionRedo
}

// IsAIAction checks if this is an AI action
func (a *Action) IsAIAction() bool {
	switch a.Type {
	case ActionAIInspect, ActionAIFind, ActionAIQuery,
		ActionAIDispatch, ActionAIWait, ActionAIWatch:
		return true
	}
	return false
}

// RequiresTarget checks if this action needs a target
func (a *Action) RequiresTarget() bool {
	return a.IsMouse() && (a.Target != "" || a.TargetID != 0)
}

// StopPropagation stops the action from propagating further
func (a *Action) StopPropagation() {
	a.stopped = true
}

// IsStopped checks if propagation has been stopped
func (a *Action) IsStopped() bool {
	return a.stopped
}

// NewAction creates a new action with auto-generated ID and timestamp
func NewAction(actionType ActionType) *Action {
	id := atomic.AddUint64(&actionIDCounter, 1)
	return &Action{
		Type:      actionType,
		ID:        id,
		Timestamp: time.Now(),
		Meta:      make(map[string]interface{}),
	}
}

// NewActionWithPayload creates an action with payload
func NewActionWithPayload(actionType ActionType, payload interface{}) *Action {
	id := atomic.AddUint64(&actionIDCounter, 1)
	return &Action{
		Type:      actionType,
		Payload:   payload,
		ID:        id,
		Timestamp: time.Now(),
		Meta:      make(map[string]interface{}),
	}
}

// NewActionFromKey creates an action from a key event
func NewActionFromKey(actionType ActionType, source string) *Action {
	id := atomic.AddUint64(&actionIDCounter, 1)
	return &Action{
		Type:      actionType,
		Source:    source,
		ID:        id,
		Timestamp: time.Now(),
		Meta:      make(map[string]interface{}),
	}
}

// NewActionFromMouse creates a mouse action
func NewActionFromMouse(actionType ActionType, localX, localY int) *Action {
	id := atomic.AddUint64(&actionIDCounter, 1)
	return &Action{
		Type:      actionType,
		Source:    "mouse",
		Payload:   struct{ X, Y int }{X: localX, Y: localY},
		ID:        id,
		Timestamp: time.Now(),
		Meta:      make(map[string]interface{}),
	}
}

// WithTarget sets the semantic target ID and automatically computes TargetID
func (a *Action) WithTarget(target string) *Action {
	a.Target = target
	a.TargetID = event.StringToNodeID(target)
	return a
}

// WithTargetID sets the internal TargetID directly
func (a *Action) WithTargetID(targetID uint64) *Action {
	a.TargetID = targetID
	// Note: We don't set Target string here to avoid reverse conversion
	// Use WithTarget() for semantic IDs
	return a
}

// WithPayload sets the action payload
func (a *Action) WithPayload(payload interface{}) *Action {
	a.Payload = payload
	return a
}

// WithSource sets the action source
func (a *Action) WithSource(source string) *Action {
	a.Source = source
	return a
}

// Clone creates a shallow copy of the action
func (a *Action) Clone() *Action {
	metaCopy := make(map[string]interface{})
	for k, v := range a.Meta {
		metaCopy[k] = v
	}
	return &Action{
		Type:      a.Type,
		Payload:   a.Payload,
		Source:    a.Source,
		Target:    a.Target,
		TargetID:  a.TargetID,
		ID:        atomic.AddUint64(&actionIDCounter, 1),
		Timestamp: a.Timestamp,
		Meta:      metaCopy,
	}
}

// String returns a string representation of the action
func (a *Action) String() string {
	var sb strings.Builder
	sb.WriteString(string(a.Type))

	if a.Target != "" {
		sb.WriteString("{")
		sb.WriteString(a.Target)
		sb.WriteString("}")
	} else if a.TargetID != 0 {
		fmt.Fprintf(&sb, "[%d]", a.TargetID)
	}

	if a.Payload != nil {
		fmt.Fprintf(&sb, "(%v)", a.Payload)
	}

	if a.Source != "" {
		fmt.Fprintf(&sb, "[%s]", a.Source)
	}

	return sb.String()
}

// ============================================================================
// Payload Helper Methods
// ============================================================================

// GetPayloadString returns the payload as a string
func (a *Action) GetPayloadString() (string, bool) {
	if s, ok := a.Payload.(string); ok {
		return s, true
	}
	return "", false
}

// GetPayloadInt returns the payload as an int
func (a *Action) GetPayloadInt() (int, bool) {
	if i, ok := a.Payload.(int); ok {
		return i, true
	}
	return 0, false
}

// GetPayloadPoint returns coordinates from payload
func (a *Action) GetPayloadPoint() (x, y int, ok bool) {
	if p, ok := a.Payload.(struct{ X, Y int }); ok {
		return p.X, p.Y, true
	}
	if p, ok := a.Payload.(map[string]int); ok {
		x, hasX := p["x"]
		y, hasY := p["y"]
		if hasX && hasY {
			return x, y, true
		}
	}
	return 0, 0, false
}

// GetPayloadSize returns size from payload
func (a *Action) GetPayloadSize() (w, h int, ok bool) {
	if s, ok := a.Payload.(struct{ W, H int }); ok {
		return s.W, s.H, true
	}
	if s, ok := a.Payload.(map[string]int); ok {
		w, hasW := s["w"]
		h, hasH := s["h"]
		if hasW && hasH {
			return w, h, true
		}
	}
	return 0, 0, false
}

// GetPayloadRune returns payload as rune
func (a *Action) GetPayloadRune() (rune, bool) {
	if r, ok := a.Payload.(rune); ok {
		return r, true
	}
	return 0, false
}

// ============================================================================
// Meta Helper Methods
// ============================================================================

// GetMeta retrieves a metadata value
func (a *Action) GetMeta(key string) (interface{}, bool) {
	if a.Meta == nil {
		return nil, false
	}
	val, ok := a.Meta[key]
	return val, ok
}

// SetMeta sets a metadata value
func (a *Action) SetMeta(key string, value interface{}) {
	if a.Meta == nil {
		a.Meta = make(map[string]interface{})
	}
	a.Meta[key] = value
}

// DeleteMeta removes a metadata value
func (a *Action) DeleteMeta(key string) {
	if a.Meta != nil {
		delete(a.Meta, key)
	}
}

// ============================================================================
// TargetID Conversion Helpers
// ============================================================================

// StringToTargetID converts a string ID to uint64 NodeID using FNV-1a hash
func StringToTargetID(source string) uint64 {
	return event.StringToNodeID(source)
}

// TargetIDToString (placeholder for potential reverse mapping)
// Currently not used; direct string-to-target is preferred
