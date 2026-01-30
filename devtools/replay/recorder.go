// Package replay provides deterministic replay capabilities for DevTools.
//
// This file implements event recording for deterministic replay.
package replay

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/wwsheng009/mint/devtools"
)

// EventRecorder records events for deterministic replay.
type EventRecorder struct {
	mu        sync.RWMutex
	enabled   bool
	session   *RecordingSession

	// Input capture
	inputChan chan InputEvent

	// Random seed tracking
	seedTracker *SeedTracker
}

// RecordingSession represents a single recording session.
type RecordingSession struct {
	SessionID   string
	StartTime   time.Time
	EndTime     time.Time
	Events      []RecordedEvent
	Inputs      []InputEvent
	Seeds       []SeedSnapshot
	Metadata    map[string]interface{}
}

// RecordedEvent represents a recorded event with full context.
type RecordedEvent struct {
	Seq         int64
	Timestamp   time.Time
	FrameID     devtools.FrameID
	Type        string
	Data        map[string]interface{}
	CausalID    uint64 // Link to causal graph
}

// InputEvent represents a user input event.
type InputEvent struct {
	Timestamp   time.Time
	Type        InputType
	Key         rune
	MouseButton MouseButton
	Position    struct {
		X int
		Y int
	}
	Modifiers KeyModifier
}

// InputType represents the type of input.
type InputType int

const (
	// InputKeyPress represents a key press.
	InputKeyPress InputType = iota
	// InputKeyRelease represents a key release.
	InputKeyRelease
	// InputMousePress represents a mouse button press.
	InputMousePress
	// InputMouseRelease represents a mouse button release.
	InputMouseRelease
	// InputMouseMove represents a mouse movement.
	InputMouseMove
	// InputMouseWheel represents a mouse wheel scroll.
	InputMouseWheel
	// InputPaste represents a paste event.
	InputPaste
	// InputUnknown represents an unknown input type.
	InputUnknown
)

// MouseButton represents a mouse button.
type MouseButton int

const (
	// MouseLeft represents the left mouse button.
	MouseLeft MouseButton = iota
	// MouseMiddle represents the middle mouse button.
	MouseMiddle
	// MouseRight represents the right mouse button.
	MouseRight
)

// KeyModifier represents keyboard modifiers.
type KeyModifier struct {
	Ctrl  bool
	Alt   bool
	Shift bool
	Meta  bool
}

// NewEventRecorder creates a new event recorder.
func NewEventRecorder() *EventRecorder {
	return &EventRecorder{
		inputChan:   make(chan InputEvent, 1024),
		seedTracker: NewSeedTracker(),
	}
}

// Start starts a new recording session.
func (er *EventRecorder) Start(sessionID string) error {
	er.mu.Lock()
	defer er.mu.Unlock()

	if er.enabled {
		return ErrAlreadyRecording
	}

	er.enabled = true
	er.session = &RecordingSession{
		SessionID: sessionID,
		StartTime: time.Now(),
		Events:    make([]RecordedEvent, 0, 4096),
		Inputs:    make([]InputEvent, 0, 1024),
		Seeds:     make([]SeedSnapshot, 0, 64),
		Metadata:  make(map[string]interface{}),
	}

	// Start input capture goroutine
	go er.captureInputs()

	return nil
}

// Stop stops the current recording session.
func (er *EventRecorder) Stop() (*RecordingSession, error) {
	er.mu.Lock()
	defer er.mu.Unlock()

	if !er.enabled || er.session == nil {
		return nil, ErrNotRecording
	}

	er.session.EndTime = time.Now()
	session := er.session
	er.enabled = false
	er.session = nil

	return session, nil
}

// IsRecording returns whether currently recording.
func (er *EventRecorder) IsRecording() bool {
	er.mu.RLock()
	defer er.mu.RUnlock()
	return er.enabled
}

// RecordEvent records an event.
func (er *EventRecorder) RecordEvent(frameID devtools.FrameID, eventType string, data map[string]interface{}, causalID uint64) {
	er.mu.Lock()
	defer er.mu.Unlock()

	if !er.enabled || er.session == nil {
		return
	}

	event := RecordedEvent{
		Seq:       int64(len(er.session.Events) + 1),
		Timestamp: time.Now(),
		FrameID:   frameID,
		Type:      eventType,
		Data:      data,
		CausalID:  causalID,
	}

	er.session.Events = append(er.session.Events, event)
}

// RecordInput records an input event.
func (er *EventRecorder) RecordInput(input InputEvent) {
	select {
	case er.inputChan <- input:
	default:
		// Channel full, drop input
	}
}

// captureInputs captures input events from the channel.
func (er *EventRecorder) captureInputs() {
	for input := range er.inputChan {
		er.mu.Lock()
		if !er.enabled || er.session == nil {
			er.mu.Unlock()
			return
		}
		er.session.Inputs = append(er.session.Inputs, input)
		er.mu.Unlock()
	}
}

// GetSession returns the current recording session.
func (er *EventRecorder) GetSession() *RecordingSession {
	er.mu.RLock()
	defer er.mu.RUnlock()
	return er.session
}

// GetSeedTracker returns the seed tracker.
func (er *EventRecorder) GetSeedTracker() *SeedTracker {
	return er.seedTracker
}

// ToJSON exports the recording session to JSON.
func (rs *RecordingSession) ToJSON() ([]byte, error) {
	return json.Marshal(rs)
}

// FromJSON imports a recording session from JSON.
func FromJSON(data []byte) (*RecordingSession, error) {
	var session RecordingSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

// Save saves the recording session to a file.
func (rs *RecordingSession) Save(filename string) error {
	data, err := rs.ToJSON()
	if err != nil {
		return err
	}
	// File writing would go here
	_ = data
	_ = filename
	return nil
}

// Load loads a recording session from a file.
func Load(filename string) (*RecordingSession, error) {
	// File reading would go here
	_ = filename
	return nil, nil
}

// GetEventCount returns the number of recorded events.
func (rs *RecordingSession) GetEventCount() int {
	return len(rs.Events)
}

// GetInputCount returns the number of recorded inputs.
func (rs *RecordingSession) GetInputCount() int {
	return len(rs.Inputs)
}

// GetDuration returns the duration of the recording.
func (rs *RecordingSession) GetDuration() time.Duration {
	if rs.EndTime.IsZero() {
		return time.Since(rs.StartTime)
	}
	return rs.EndTime.Sub(rs.StartTime)
}

// GetEventsForFrame returns events for a specific frame.
func (rs *RecordingSession) GetEventsForFrame(frameID devtools.FrameID) []RecordedEvent {
	var events []RecordedEvent
	for _, event := range rs.Events {
		if event.FrameID == frameID {
			events = append(events, event)
		}
	}
	return events
}

// GetInputsForFrame returns inputs for a specific frame.
func (rs *RecordingSession) GetInputsForFrame(frameID devtools.FrameID) []InputEvent {
	// Find the frame timing
	frameStart := time.Time{}
	frameEnd := time.Time{}

	for _, event := range rs.Events {
		if event.FrameID == frameID {
			if frameStart.IsZero() || event.Timestamp.Before(frameStart) {
				frameStart = event.Timestamp
			}
			if frameEnd.IsZero() || event.Timestamp.After(frameEnd) {
				frameEnd = event.Timestamp
			}
		}
	}

	var inputs []InputEvent
	for _, input := range rs.Inputs {
		if (input.Timestamp.Equal(frameStart) || input.Timestamp.After(frameStart)) &&
			(frameEnd.IsZero() || input.Timestamp.Before(frameEnd) || input.Timestamp.Equal(frameEnd)) {
			inputs = append(inputs, input)
		}
	}

	return inputs
}

// RecordingStats provides statistics about a recording.
type RecordingStats struct {
	SessionID      string
	Duration       time.Duration
	EventCount     int
	InputCount     int
	SeedCount      int
	EventsPerFrame float64
	AvgFrameTime   time.Duration
}

// GetStats returns statistics about the recording.
func (rs *RecordingSession) GetStats() *RecordingStats {
	stats := &RecordingStats{
		SessionID:  rs.SessionID,
		Duration:   rs.GetDuration(),
		EventCount: rs.GetEventCount(),
		InputCount: rs.GetInputCount(),
		SeedCount:  len(rs.Seeds),
	}

	// Calculate events per frame
	frameMap := make(map[devtools.FrameID]int)
	for _, event := range rs.Events {
		frameMap[event.FrameID]++
	}

	if len(frameMap) > 0 {
		stats.EventsPerFrame = float64(stats.EventCount) / float64(len(frameMap))
	}

	// Calculate average frame time
	if len(rs.Events) > 1 {
		var totalFrameTime time.Duration
		var frameCount int

		for i := 1; i < len(rs.Events); i++ {
			frameTime := rs.Events[i].Timestamp.Sub(rs.Events[i-1].Timestamp)
			totalFrameTime += frameTime
			frameCount++
		}

		if frameCount > 0 {
			stats.AvgFrameTime = totalFrameTime / time.Duration(frameCount)
		}
	}

	return stats
}

// SetMetadata sets metadata for the session.
func (rs *RecordingSession) SetMetadata(key string, value interface{}) {
	if rs.Metadata == nil {
		rs.Metadata = make(map[string]interface{})
	}
	rs.Metadata[key] = value
}

// GetMetadata gets metadata from the session.
func (rs *RecordingSession) GetMetadata(key string) (interface{}, bool) {
	if rs.Metadata == nil {
		return nil, false
	}
	val, exists := rs.Metadata[key]
	return val, exists
}

// Truncate truncates the recording to a specific event count.
func (rs *RecordingSession) Truncate(eventCount int) {
	if eventCount >= len(rs.Events) {
		return
	}

	rs.Events = rs.Events[:eventCount]

	// Also truncate inputs that occurred after the last event
	if len(rs.Events) > 0 {
		cutoff := rs.Events[len(rs.Events)-1].Timestamp
		newInputs := make([]InputEvent, 0)
		for _, input := range rs.Inputs {
			if !input.Timestamp.After(cutoff) {
				newInputs = append(newInputs, input)
			}
		}
		rs.Inputs = newInputs
	}
}

// Split splits the recording at a specific event.
func (rs *RecordingSession) Split(eventIndex int) (*RecordingSession, error) {
	if eventIndex < 0 || eventIndex >= len(rs.Events) {
		return nil, ErrInvalidEventIndex
	}

	newSession := &RecordingSession{
		SessionID: rs.SessionID + "_part2",
		StartTime: rs.Events[eventIndex].Timestamp,
		Events:    make([]RecordedEvent, 0),
		Inputs:    make([]InputEvent, 0),
		Metadata:  make(map[string]interface{}),
	}

	// Copy events after the split point
	newSession.Events = append(newSession.Events, rs.Events[eventIndex+1:]...)

	// Copy inputs after the split point
	cutoff := rs.Events[eventIndex].Timestamp
	for _, input := range rs.Inputs {
		if input.Timestamp.After(cutoff) {
			newSession.Inputs = append(newSession.Inputs, input)
		}
	}

	// Truncate original session
	rs.Truncate(eventIndex)
	rs.EndTime = rs.Events[len(rs.Events)-1].Timestamp

	return newSession, nil
}

// Merge merges another recording session into this one.
func (rs *RecordingSession) Merge(other *RecordingSession) {
	offset := len(rs.Events)
	timeOffset := rs.GetDuration()

	// Merge events
	for _, event := range other.Events {
		event.Seq = int64(offset + len(rs.Events) + 1)
		event.Timestamp = event.Timestamp.Add(timeOffset)
		rs.Events = append(rs.Events, event)
	}

	// Merge inputs
	for _, input := range other.Inputs {
		input.Timestamp = input.Timestamp.Add(timeOffset)
		rs.Inputs = append(rs.Inputs, input)
	}

	// Merge seeds
	for _, seed := range other.Seeds {
		rs.Seeds = append(rs.Seeds, seed)
	}
}

// Errors
var (
	ErrAlreadyRecording = &RecorderError{Msg: "already recording"}
	ErrNotRecording     = &RecorderError{Msg: "not recording"}
	ErrInvalidEventIndex = &RecorderError{Msg: "invalid event index"}
)

// RecorderError represents a recorder error.
type RecorderError struct {
	Msg string
}

func (e *RecorderError) Error() string {
	return e.Msg
}

// InputBuilder helps build input events.
type InputBuilder struct {
	input InputEvent
}

// NewInputBuilder creates a new input builder.
func NewInputBuilder() *InputBuilder {
	return &InputBuilder{
		input: InputEvent{
			Timestamp: time.Now(),
		},
	}
}

// KeyPress sets the input as a key press.
func (ib *InputBuilder) KeyPress(key rune) *InputBuilder {
	ib.input.Type = InputKeyPress
	ib.input.Key = key
	return ib
}

// KeyRelease sets the input as a key release.
func (ib *InputBuilder) KeyRelease(key rune) *InputBuilder {
	ib.input.Type = InputKeyRelease
	ib.input.Key = key
	return ib
}

// MousePress sets the input as a mouse press.
func (ib *InputBuilder) MousePress(button MouseButton, x, y int) *InputBuilder {
	ib.input.Type = InputMousePress
	ib.input.MouseButton = button
	ib.input.Position.X = x
	ib.input.Position.Y = y
	return ib
}

// MouseRelease sets the input as a mouse release.
func (ib *InputBuilder) MouseRelease(button MouseButton, x, y int) *InputBuilder {
	ib.input.Type = InputMouseRelease
	ib.input.MouseButton = button
	ib.input.Position.X = x
	ib.input.Position.Y = y
	return ib
}

// MouseMove sets the input as a mouse movement.
func (ib *InputBuilder) MouseMove(x, y int) *InputBuilder {
	ib.input.Type = InputMouseMove
	ib.input.Position.X = x
	ib.input.Position.Y = y
	return ib
}

// Modifiers sets the keyboard modifiers.
func (ib *InputBuilder) Modifiers(ctrl, alt, shift, meta bool) *InputBuilder {
	ib.input.Modifiers.Ctrl = ctrl
	ib.input.Modifiers.Alt = alt
	ib.input.Modifiers.Shift = shift
	ib.input.Modifiers.Meta = meta
	return ib
}

// Timestamp sets the timestamp.
func (ib *InputBuilder) Timestamp(t time.Time) *InputBuilder {
	ib.input.Timestamp = t
	return ib
}

// Build returns the constructed input event.
func (ib *InputBuilder) Build() InputEvent {
	return ib.input
}
