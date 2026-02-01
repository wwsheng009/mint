// Package scheduler provides mouse event handling with throttling.
package scheduler

import (
	"sync"
	"time"
)

// MouseEvent represents a mouse event.
type MouseEvent struct {
	// X is the column position (0-indexed).
	X int
	// Y is the row position (0-indexed).
	Y int
	// Button is the mouse button (1=left, 2=middle, 3=right, 4=scroll up, 5=scroll down).
	Button int
	// Type is the event type (press, release, motion, click).
	Type string
	// Mod indicates if a modifier key was pressed (shift, ctrl, alt).
	Mod bool
	// Timestamp when the event occurred.
	Timestamp time.Time
}

// MouseEventHandler handles mouse events.
type MouseEventHandler interface {
	// HandleMouse is called for each mouse event.
	HandleMouse(event *MouseEvent)
}

// MouseHandlerFunc is a function adapter for MouseEventHandler.
type MouseHandlerFunc func(event *MouseEvent)

// HandleMouse implements MouseEventHandler.
func (f MouseHandlerFunc) HandleMouse(event *MouseEvent) {
	f(event)
}

// ThrottleConfig configures mouse event throttling.
type ThrottleConfig struct {
	// MotionInterval is the minimum interval between motion events.
	MotionInterval time.Duration
	// ClickInterval is the minimum interval between click events.
	ClickInterval time.Duration
	// MaxDistance is the maximum distance to consider events as duplicate.
	MaxDistance int
}

// DefaultThrottleConfig returns the default throttle configuration.
func DefaultThrottleConfig() ThrottleConfig {
	return ThrottleConfig{
		MotionInterval: 16 * time.Millisecond, // ~60fps
		ClickInterval:  100 * time.Millisecond,
		MaxDistance:   2,
	}
}

// MouseMoveHandler handles mouse move events with throttling.
type MouseMoveHandler struct {
	mu       sync.RWMutex
	config   ThrottleConfig
	handler  MouseEventHandler
	lastTime time.Time
	lastX    int
	lastY    int
	pending  *MouseEvent
	timer    *time.Timer
	stopped  bool
}

// NewMouseMoveHandler creates a new mouse move handler.
func NewMouseMoveHandler(handler MouseEventHandler, config ThrottleConfig) *MouseMoveHandler {
	if config.MotionInterval == 0 {
		config = DefaultThrottleConfig()
	}

	return &MouseMoveHandler{
		config:  config,
		handler: handler,
		stopped: false,
	}
}

// Handle processes a mouse move event with throttling.
func (h *MouseMoveHandler) Handle(x, y int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.stopped {
		return
	}

	now := time.Now()
	event := &MouseEvent{
		X:         x,
		Y:         y,
		Type:      "motion",
		Timestamp: now,
	}

	// Check if this is a duplicate or too frequent
	elapsed := now.Sub(h.lastTime)

	if h.pending != nil {
		// Update pending event
		h.pending.X = x
		h.pending.Y = y
		h.pending.Timestamp = now
	} else if elapsed >= h.config.MotionInterval {
		// Send immediately if enough time has passed
		h.handler.HandleMouse(event)
		h.lastTime = now
		h.lastX = x
		h.lastY = y
	} else {
		// Queue for later
		h.pending = event
		h.scheduleFlush(elapsed)
	}
}

// scheduleFlush schedules a flush of the pending event.
func (h *MouseMoveHandler) scheduleFlush(delay time.Duration) {
	if h.timer != nil {
		h.timer.Stop()
	}

	h.timer = time.AfterFunc(h.config.MotionInterval-delay, func() {
		h.flushPending()
	})
}

// flushPending sends the pending event if any.
func (h *MouseMoveHandler) flushPending() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.pending != nil && !h.stopped {
		h.handler.HandleMouse(h.pending)
		h.lastTime = h.pending.Timestamp
		h.lastX = h.pending.X
		h.lastY = h.pending.Y
		h.pending = nil
	}
}

// Stop stops the handler and flushes any pending event.
func (h *MouseMoveHandler) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.stopped = true

	if h.timer != nil {
		h.timer.Stop()
		h.timer = nil
	}

	// Flush pending event
	if h.pending != nil {
		h.handler.HandleMouse(h.pending)
		h.pending = nil
	}
}

// Reset clears the handler state.
func (h *MouseMoveHandler) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.timer != nil {
		h.timer.Stop()
		h.timer = nil
	}

	h.lastTime = time.Time{}
	h.lastX = 0
	h.lastY = 0
	h.pending = nil
	h.stopped = false
}

// =============================================================================
// Mouse Click Handler (Debouncing)
// =============================================================================

// MouseClickHandler handles mouse click events with debouncing.
type MouseClickHandler struct {
	mu          sync.RWMutex
	config      ThrottleConfig
	handler     MouseEventHandler
	lastTime    time.Time
	lastButton  int
	pending     *MouseEvent
	timer       *time.Timer
	stopped     bool
	clickCount  int // For double-click detection
}

// NewMouseClickHandler creates a new mouse click handler.
func NewMouseClickHandler(handler MouseEventHandler, config ThrottleConfig) *MouseClickHandler {
	if config.ClickInterval == 0 {
		config = DefaultThrottleConfig()
	}

	return &MouseClickHandler{
		config:  config,
		handler: handler,
		stopped: false,
	}
}

// Handle processes a mouse click event with debouncing.
func (h *MouseClickHandler) Handle(x, y, button int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.stopped {
		return
	}

	now := time.Now()
	event := &MouseEvent{
		X:         x,
		Y:         y,
		Button:    button,
		Type:      "click",
		Timestamp: now,
	}

	elapsed := now.Sub(h.lastTime)

	// Check for double-click
	if h.lastButton == button && elapsed < h.config.ClickInterval {
		h.clickCount++
		if h.clickCount == 2 {
			event.Type = "double-click"
			h.clickCount = 0
		}
	} else {
		h.clickCount = 1
	}

	// Debounce: cancel previous timer and schedule new one
	if h.timer != nil {
		h.timer.Stop()
	}

	h.pending = event
	h.timer = time.AfterFunc(h.config.ClickInterval, func() {
		h.flushPending()
	})

	h.lastTime = now
	h.lastButton = button
}

// flushPending sends the pending click event.
func (h *MouseClickHandler) flushPending() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.pending != nil && !h.stopped {
		h.handler.HandleMouse(h.pending)
		h.pending = nil
	}
}

// FlushImmediately sends any pending event immediately.
func (h *MouseClickHandler) FlushImmediately() {
	h.mu.Lock()

	if h.timer != nil {
		h.timer.Stop()
		h.timer = nil
	}

	pending := h.pending
	h.pending = nil

	h.mu.Unlock()

	if pending != nil {
		h.handler.HandleMouse(pending)
	}
}

// Stop stops the handler without flushing.
func (h *MouseClickHandler) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.stopped = true

	if h.timer != nil {
		h.timer.Stop()
		h.timer = nil
	}

	h.pending = nil
}

// =============================================================================
// Combined Mouse Handler
// =============================================================================

// MouseHandler handles all mouse events with appropriate throttling/debouncing.
type MouseHandler struct {
	mu            sync.RWMutex
	moveHandler   *MouseMoveHandler
	clickHandler  *MouseClickHandler
	pressHandler  MouseEventHandler
	releaseHandler MouseEventHandler
	wheelHandler  MouseEventHandler
}

// NewMouseHandler creates a new combined mouse handler.
func NewMouseHandler(move, click, press, release, wheel MouseEventHandler, config ThrottleConfig) *MouseHandler {
	return &MouseHandler{
		moveHandler:   NewMouseMoveHandler(move, config),
		clickHandler:  NewMouseClickHandler(click, config),
		pressHandler:  press,
		releaseHandler: release,
		wheelHandler:  wheel,
	}
}

// HandleMotion handles a mouse motion event.
func (h *MouseHandler) HandleMotion(x, y int) {
	h.mu.RLock()
	handler := h.moveHandler
	h.mu.RUnlock()

	if handler != nil {
		handler.Handle(x, y)
	}
}

// HandleClick handles a mouse click event.
func (h *MouseHandler) HandleClick(x, y, button int) {
	h.mu.RLock()
	handler := h.clickHandler
	h.mu.RUnlock()

	if handler != nil {
		handler.Handle(x, y, button)
	}
}

// HandlePress handles a mouse press event.
func (h *MouseHandler) HandlePress(x, y, button int) {
	h.mu.RLock()
	handler := h.pressHandler
	h.mu.RUnlock()

	if handler != nil {
		handler.HandleMouse(&MouseEvent{
			X:         x,
			Y:         y,
			Button:    button,
			Type:      "press",
			Timestamp: time.Now(),
		})
	}
}

// HandleRelease handles a mouse release event.
func (h *MouseHandler) HandleRelease(x, y, button int) {
	h.mu.RLock()
	handler := h.releaseHandler
	h.mu.RUnlock()

	if handler != nil {
		handler.HandleMouse(&MouseEvent{
			X:         x,
			Y:         y,
			Button:    button,
			Type:      "release",
			Timestamp: time.Now(),
		})
	}
}

// HandleWheel handles a mouse wheel event.
func (h *MouseHandler) HandleWheel(x, y int, delta int) {
	h.mu.RLock()
	handler := h.wheelHandler
	h.mu.RUnlock()

	if handler != nil {
		handler.HandleMouse(&MouseEvent{
			X:         x,
			Y:         y,
			Button:    delta,
			Type:      "wheel",
			Timestamp: time.Now(),
		})
	}
}

// Flush flushes all pending events.
func (h *MouseHandler) Flush() {
	h.mu.RLock()
	moveHandler := h.moveHandler
	clickHandler := h.clickHandler
	h.mu.RUnlock()

	if clickHandler != nil {
		clickHandler.FlushImmediately()
	}
	if moveHandler != nil {
		moveHandler.flushPending()
	}
}

// Stop stops all handlers.
func (h *MouseHandler) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.moveHandler != nil {
		h.moveHandler.Stop()
	}
	if h.clickHandler != nil {
		h.clickHandler.Stop()
	}
}

// Reset resets all handlers.
func (h *MouseHandler) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.moveHandler != nil {
		h.moveHandler.Reset()
	}
}

// =============================================================================
// Mouse Tracking
// =============================================================================

// MouseTracker tracks mouse position and button state.
type MouseTracker struct {
	mu       sync.RWMutex
	x        int
	y        int
	buttons  map[int]bool
	inWindow bool
}

// NewMouseTracker creates a new mouse tracker.
func NewMouseTracker() *MouseTracker {
	return &MouseTracker{
		buttons: make(map[int]bool),
	}
}

// UpdatePosition updates the mouse position.
func (t *MouseTracker) UpdatePosition(x, y int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.x = x
	t.y = y
}

// UpdateButton updates a button's pressed state.
func (t *MouseTracker) UpdateButton(button int, pressed bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if pressed {
		t.buttons[button] = true
	} else {
		delete(t.buttons, button)
	}
}

// Position returns the current mouse position.
func (t *MouseTracker) Position() (x, y int) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.x, t.y
}

// IsPressed returns true if a button is currently pressed.
func (t *MouseTracker) IsPressed(button int) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.buttons[button]
}

// PressedButtons returns all currently pressed buttons.
func (t *MouseTracker) PressedButtons() []int {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make([]int, 0, len(t.buttons))
	for button := range t.buttons {
		result = append(result, button)
	}

	return result
}

// SetInWindow sets whether the mouse is in the window.
func (t *MouseTracker) SetInWindow(in bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.inWindow = in
}

// IsInWindow returns true if the mouse is in the window.
func (t *MouseTracker) IsInWindow() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.inWindow
}
