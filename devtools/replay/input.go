// Package replay provides input capture for deterministic replay.
//
// This file implements input event capture for replay.
package replay

import (
	"sync"
	"time"
)

// InputCapture captures user input for replay.
type InputCapture struct {
	mu          sync.RWMutex
	enabled     bool
	filter      InputFilter
	inputChan   chan InputEvent
	buffer      []InputEvent
	bufferSize  int

	// Statistics
	capturedCount int64
	droppedCount  int64
}

// InputFilter determines which inputs to capture.
type InputFilter interface {
	ShouldCapture(input InputEvent) bool
}

// DefaultInputFilter captures all inputs.
type DefaultInputFilter struct{}

// ShouldCapture returns true for all inputs.
func (f *DefaultInputFilter) ShouldCapture(input InputEvent) bool {
	return true
}

// NewInputCapture creates a new input capture.
func NewInputCapture(bufferSize int) *InputCapture {
	return &InputCapture{
		inputChan:  make(chan InputEvent, bufferSize*2),
		buffer:     make([]InputEvent, 0, bufferSize),
		bufferSize: bufferSize,
		filter:     &DefaultInputFilter{},
	}
}

// Enable enables input capture.
func (ic *InputCapture) Enable() {
	ic.mu.Lock()
	defer ic.mu.Unlock()
	ic.enabled = true
}

// Disable disables input capture.
func (ic *InputCapture) Disable() {
	ic.mu.Lock()
	defer ic.mu.Unlock()
	ic.enabled = false
}

// IsEnabled returns whether capture is enabled.
func (ic *InputCapture) IsEnabled() bool {
	ic.mu.RLock()
	defer ic.mu.RUnlock()
	return ic.enabled
}

// SetFilter sets the input filter.
func (ic *InputCapture) SetFilter(filter InputFilter) {
	ic.mu.Lock()
	defer ic.mu.Unlock()
	ic.filter = filter
}

// Capture captures an input event.
func (ic *InputCapture) Capture(input InputEvent) bool {
	ic.mu.RLock()
	enabled := ic.enabled
	filter := ic.filter
	ic.mu.RUnlock()

	if !enabled || (filter != nil && !filter.ShouldCapture(input)) {
		return false
	}

	select {
	case ic.inputChan <- input:
		return true
	default:
		// Channel full, count as dropped
		ic.mu.Lock()
		ic.droppedCount++
		ic.mu.Unlock()
		return false
	}
}

// CaptureKeyPress captures a key press.
func (ic *InputCapture) CaptureKeyPress(key rune, modifiers KeyModifier) bool {
	return ic.Capture(InputEvent{
		Timestamp: time.Now(),
		Type:      InputKeyPress,
		Key:       key,
		Modifiers: modifiers,
	})
}

// CaptureKeyRelease captures a key release.
func (ic *InputCapture) CaptureKeyRelease(key rune, modifiers KeyModifier) bool {
	return ic.Capture(InputEvent{
		Timestamp: time.Now(),
		Type:      InputKeyRelease,
		Key:       key,
		Modifiers: modifiers,
	})
}

// CaptureMousePress captures a mouse button press.
func (ic *InputCapture) CaptureMousePress(button MouseButton, x, y int) bool {
	return ic.Capture(InputEvent{
		Timestamp: time.Now(),
		Type:      InputMousePress,
		MouseButton: button,
		Position: struct {
			X int
			Y int
		}{X: x, Y: y},
	})
}

// CaptureMouseRelease captures a mouse button release.
func (ic *InputCapture) CaptureMouseRelease(button MouseButton, x, y int) bool {
	return ic.Capture(InputEvent{
		Timestamp: time.Now(),
		Type:      InputMouseRelease,
		MouseButton: button,
		Position: struct {
			X int
			Y int
		}{X: x, Y: y},
	})
}

// CaptureMouseMove captures a mouse movement.
func (ic *InputCapture) CaptureMouseMove(x, y int) bool {
	return ic.Capture(InputEvent{
		Timestamp: time.Now(),
		Type:      InputMouseMove,
		Position: struct {
			X int
			Y int
		}{X: x, Y: y},
	})
}

// CaptureMouseWheel captures a mouse wheel scroll.
func (ic *InputCapture) CaptureMouseWheel(delta int) bool {
	return ic.Capture(InputEvent{
		Timestamp: time.Now(),
		Type:      InputMouseWheel,
	})
}

// Start starts the capture goroutine.
func (ic *InputCapture) Start() {
	go ic.captureLoop()
}

// Stop stops the capture.
func (ic *InputCapture) Stop() {
	ic.Disable()
	close(ic.inputChan)
}

// captureLoop is the capture loop.
func (ic *InputCapture) captureLoop() {
	for input := range ic.inputChan {
		ic.mu.Lock()
		if len(ic.buffer) >= ic.bufferSize {
			// Remove oldest
			ic.buffer = ic.buffer[1:]
		}
		ic.buffer = append(ic.buffer, input)
		ic.capturedCount++
		ic.mu.Unlock()
	}
}

// GetCapturedInputs returns all captured inputs.
func (ic *InputCapture) GetCapturedInputs() []InputEvent {
	ic.mu.RLock()
	defer ic.mu.RUnlock()

	inputs := make([]InputEvent, len(ic.buffer))
	copy(inputs, ic.buffer)
	return inputs
}

// GetCapturedInputsSince returns inputs since a given time.
func (ic *InputCapture) GetCapturedInputsSince(since time.Time) []InputEvent {
	ic.mu.RLock()
	defer ic.mu.RUnlock()

	var inputs []InputEvent
	for _, input := range ic.buffer {
		if input.Timestamp.After(since) || input.Timestamp.Equal(since) {
			inputs = append(inputs, input)
		}
	}
	return inputs
}

// Clear clears the captured inputs.
func (ic *InputCapture) Clear() {
	ic.mu.Lock()
	defer ic.mu.Unlock()
	ic.buffer = make([]InputEvent, 0, ic.bufferSize)
}

// GetStats returns capture statistics.
func (ic *InputCapture) GetStats() *InputCaptureStats {
	ic.mu.RLock()
	defer ic.mu.RUnlock()

	return &InputCaptureStats{
		CapturedCount: ic.capturedCount,
		DroppedCount:  ic.droppedCount,
		BufferSize:    ic.bufferSize,
		BufferUsed:    len(ic.buffer),
		Enabled:       ic.enabled,
	}
}

// InputCaptureStats represents input capture statistics.
type InputCaptureStats struct {
	CapturedCount int64
	DroppedCount  int64
	BufferSize    int
	BufferUsed    int
	Enabled       bool
}

// InputSequence represents a sequence of inputs.
type InputSequence struct {
	Inputs []InputEvent
}

// NewInputSequence creates a new input sequence.
func NewInputSequence() *InputSequence {
	return &InputSequence{
		Inputs: make([]InputEvent, 0),
	}
}

// Add adds an input to the sequence.
func (is *InputSequence) Add(input InputEvent) {
	is.Inputs = append(is.Inputs, input)
}

// AddKeyPress adds a key press to the sequence.
func (is *InputSequence) AddKeyPress(key rune) *InputSequence {
	is.Inputs = append(is.Inputs, InputEvent{
		Type: InputKeyPress,
		Key:  key,
	})
	return is
}

// AddKeyRelease adds a key release to the sequence.
func (is *InputSequence) AddKeyRelease(key rune) *InputSequence {
	is.Inputs = append(is.Inputs, InputEvent{
		Type: InputKeyRelease,
		Key:  key,
	})
	return is
}

// AddMouseClick adds a mouse click to the sequence.
func (is *InputSequence) AddMouseClick(button MouseButton, x, y int) *InputSequence {
	// Press
	is.Inputs = append(is.Inputs, InputEvent{
		Type:         InputMousePress,
		MouseButton:  button,
		Position:     struct{ X, Y int }{X: x, Y: y},
	})
	// Release
	is.Inputs = append(is.Inputs, InputEvent{
		Type:         InputMouseRelease,
		MouseButton:  button,
		Position:     struct{ X, Y int }{X: x, Y: y},
	})
	return is
}

// AddText adds text input to the sequence.
func (is *InputSequence) AddText(text string) *InputSequence {
	for _, r := range text {
		is.Inputs = append(is.Inputs, InputEvent{
			Type: InputKeyPress,
			Key:  r,
		})
	}
	return is
}

// Delay adds a delay to the sequence.
func (is *InputSequence) Delay(duration time.Duration) *InputSequence {
	// The last input's timestamp is used as reference
	if len(is.Inputs) > 0 {
		lastTime := is.Inputs[len(is.Inputs)-1].Timestamp
		newTime := lastTime.Add(duration)
		// Add a placeholder input for the delay
		is.Inputs = append(is.Inputs, InputEvent{
			Timestamp: newTime,
			Type:      InputUnknown,
		})
	}
	return is
}

// GetInputs returns the input sequence.
func (is *InputSequence) GetInputs() []InputEvent {
	return is.Inputs
}

// Clear clears the sequence.
func (is *InputSequence) Clear() {
	is.Inputs = make([]InputEvent, 0)
}

// Length returns the number of inputs in the sequence.
func (is *InputSequence) Length() int {
	return len(is.Inputs)
}

// Macro represents a reusable input macro.
type Macro struct {
	Name   string
	Inputs []InputEvent
}

// NewMacro creates a new macro.
func NewMacro(name string) *Macro {
	return &Macro{
		Name:   name,
		Inputs: make([]InputEvent, 0),
	}
}

// Record records inputs into the macro.
func (m *Macro) Record(inputs []InputEvent) {
	m.Inputs = append(m.Inputs, inputs...)
}

// Play plays the macro through an input capture.
func (m *Macro) Play(capture *InputCapture) error {
	for _, input := range m.Inputs {
		if !capture.Capture(input) {
			// Input was dropped or filtered
		}
	}
	return nil
}

// GetInputs returns the macro's inputs.
func (m *Macro) GetInputs() []InputEvent {
	return m.Inputs
}

// Clear clears the macro.
func (m *Macro) Clear() {
	m.Inputs = make([]InputEvent, 0)
}

// MacroRegistry manages a collection of macros.
type MacroRegistry struct {
	macros map[string]*Macro
	mu     sync.RWMutex
}

// NewMacroRegistry creates a new macro registry.
func NewMacroRegistry() *MacroRegistry {
	return &MacroRegistry{
		macros: make(map[string]*Macro),
	}
}

// Register registers a macro.
func (mr *MacroRegistry) Register(macro *Macro) {
	mr.mu.Lock()
	defer mr.mu.Unlock()
	mr.macros[macro.Name] = macro
}

// Unregister unregisters a macro.
func (mr *MacroRegistry) Unregister(name string) {
	mr.mu.Lock()
	defer mr.mu.Unlock()
	delete(mr.macros, name)
}

// Get gets a macro by name.
func (mr *MacroRegistry) Get(name string) (*Macro, bool) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()
	macro, exists := mr.macros[name]
	return macro, exists
}

// List returns all macro names.
func (mr *MacroRegistry) List() []string {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	names := make([]string, 0, len(mr.macros))
	for name := range mr.macros {
		names = append(names, name)
	}
	return names
}

// Clear clears all macros.
func (mr *MacroRegistry) Clear() {
	mr.mu.Lock()
	defer mr.mu.Unlock()
	mr.macros = make(map[string]*Macro)
}
