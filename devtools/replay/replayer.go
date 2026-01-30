// Package replay provides event replay capabilities for DevTools.
//
// This file implements event replaying for deterministic replay.
package replay

import (
	"fmt"
	"sync"
	"time"

	"github.com/wwsheng009/mint/devtools"
)

// EventReplayer replays events for deterministic testing.
type EventReplayer struct {
	mu           sync.RWMutex
	session      *RecordingSession
	running      bool
	paused       bool
	currentFrame int

	// Callbacks
	onEvent      func(RecordedEvent)
	onInput      func(InputEvent)
	onFrameStart func(devtools.FrameID)
	onFrameEnd   func(devtools.FrameID)

	// Timing control
	replaySpeed  float64
	realTimeMode bool

	// Seed tracker
	seedTracker  *SeedTracker
}

// NewEventReplayer creates a new event replayer.
func NewEventReplayer(session *RecordingSession) *EventReplayer {
	return &EventReplayer{
		session:      session,
		replaySpeed:  1.0,
		realTimeMode: false,
		seedTracker:  NewSeedTracker(),
	}
}

// Load loads a recording session for replay.
func (er *EventReplayer) Load(session *RecordingSession) {
	er.mu.Lock()
	defer er.mu.Unlock()

	er.session = session
	er.currentFrame = 0
}

// Start starts replaying the session.
func (er *EventReplayer) Start() error {
	er.mu.Lock()
	defer er.mu.Unlock()

	if er.session == nil {
		return ErrNoSessionLoaded
	}

	if er.running {
		return ErrAlreadyReplaying
	}

	er.running = true
	er.paused = false
	er.currentFrame = 0

	// Initialize seed tracker from session
	er.seedTracker.LoadFromSession(er.session)

	go er.replayLoop()

	return nil
}

// Stop stops the replay.
func (er *EventReplayer) Stop() {
	er.mu.Lock()
	defer er.mu.Unlock()

	er.running = false
	er.paused = false
}

// Pause pauses the replay.
func (er *EventReplayer) Pause() {
	er.mu.Lock()
	defer er.mu.Unlock()

	if er.running {
		er.paused = true
	}
}

// Resume resumes the replay.
func (er *EventReplayer) Resume() {
	er.mu.Lock()
	defer er.mu.Unlock()

	if er.running && er.paused {
		er.paused = false
	}
}

// IsRunning returns whether replay is running.
func (er *EventReplayer) IsRunning() bool {
	er.mu.RLock()
	defer er.mu.RUnlock()
	return er.running
}

// IsPaused returns whether replay is paused.
func (er *EventReplayer) IsPaused() bool {
	er.mu.RLock()
	defer er.mu.RUnlock()
	return er.paused
}

// SetReplaySpeed sets the replay speed multiplier.
func (er *EventReplayer) SetReplaySpeed(speed float64) {
	er.mu.Lock()
	defer er.mu.Unlock()
	er.replaySpeed = speed
}

// GetReplaySpeed returns the replay speed.
func (er *EventReplayer) GetReplaySpeed() float64 {
	er.mu.RLock()
	defer er.mu.RUnlock()
	return er.replaySpeed
}

// SetRealTimeMode enables or disables real-time replay.
func (er *EventReplayer) SetRealTimeMode(enabled bool) {
	er.mu.Lock()
	defer er.mu.Unlock()
	er.realTimeMode = enabled
}

// GetProgress returns the current replay progress.
func (er *EventReplayer) GetProgress() *ReplayProgress {
	er.mu.RLock()
	defer er.mu.RUnlock()

	progress := &ReplayProgress{
		CurrentEvent: er.currentFrame,
		TotalEvents:  len(er.session.Events),
		Running:      er.running,
		Paused:       er.paused,
		Speed:        er.replaySpeed,
	}

	if len(er.session.Events) > 0 {
		progress.PercentComplete = float64(er.currentFrame) / float64(len(er.session.Events)) * 100
	}

	return progress
}

// replayLoop is the main replay loop.
func (er *EventReplayer) replayLoop() {
	eventIndex := 0

	for er.running {
		er.mu.RLock()
		paused := er.paused
		speed := er.replaySpeed
		realTime := er.realTimeMode
		er.mu.RUnlock()

		if paused {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		if eventIndex >= len(er.session.Events) {
			// End of session
			er.running = false
			break
		}

		event := er.session.Events[eventIndex]

		// Handle timing in real-time mode
		if realTime && eventIndex > 0 {
			originalDelay := event.Timestamp.Sub(er.session.Events[eventIndex-1].Timestamp)
			adjustedDelay := time.Duration(float64(originalDelay) / speed)
			time.Sleep(adjustedDelay)
		}

		// Notify frame start
		if er.onFrameStart != nil {
			er.onFrameStart(event.FrameID)
		}

		// Emit inputs for this frame
		er.emitInputsForFrame(event.FrameID)

		// Emit the event
		if er.onEvent != nil {
			er.onEvent(event)
		}

		// Apply seed for this frame
		if er.seedTracker != nil {
			if seed := er.seedTracker.GetSeedForFrame(event.FrameID); seed != nil {
				er.seedTracker.ApplySeed(seed.Source, seed.Value)
			}
		}

		// Notify frame end
		if er.onFrameEnd != nil {
			er.onFrameEnd(event.FrameID)
		}

		eventIndex++
		er.currentFrame = eventIndex
	}
}

// emitInputsForFrame emits all inputs for a specific frame.
func (er *EventReplayer) emitInputsForFrame(frameID devtools.FrameID) {
	if er.onInput == nil {
		return
	}

	inputs := er.session.GetInputsForFrame(frameID)
	for _, input := range inputs {
		er.onInput(input)
	}
}

// JumpToFrame jumps to a specific frame.
func (er *EventReplayer) JumpToFrame(frameID devtools.FrameID) error {
	er.mu.Lock()
	defer er.mu.Unlock()

	// Find the frame in the session
	targetIndex := -1
	for i, event := range er.session.Events {
		if event.FrameID == frameID {
			targetIndex = i
			break
		}
	}

	if targetIndex == -1 {
		return fmt.Errorf("frame %d not found in session", frameID)
	}

	er.currentFrame = targetIndex

	// Emit all events up to the target frame
	if er.onEvent != nil {
		for i := 0; i <= targetIndex; i++ {
			er.onEvent(er.session.Events[i])
		}
	}

	return nil
}

// StepFrame advances one frame.
func (er *EventReplayer) StepFrame() error {
	er.mu.Lock()
	defer er.mu.Unlock()

	if er.session == nil {
		return ErrNoSessionLoaded
	}

	if er.currentFrame >= len(er.session.Events) {
		return ErrEndOfSession
	}

	event := er.session.Events[er.currentFrame]

	if er.onFrameStart != nil {
		er.onFrameStart(event.FrameID)
	}

	er.emitInputsForFrame(event.FrameID)

	if er.onEvent != nil {
		er.onEvent(event)
	}

	if er.seedTracker != nil {
		if seed := er.seedTracker.GetSeedForFrame(event.FrameID); seed != nil {
			er.seedTracker.ApplySeed(seed.Source, seed.Value)
		}
	}

	if er.onFrameEnd != nil {
		er.onFrameEnd(event.FrameID)
	}

	er.currentFrame++

	return nil
}

// SetEventCallback sets the callback for events.
func (er *EventReplayer) SetEventCallback(fn func(RecordedEvent)) {
	er.mu.Lock()
	defer er.mu.Unlock()
	er.onEvent = fn
}

// SetInputCallback sets the callback for inputs.
func (er *EventReplayer) SetInputCallback(fn func(InputEvent)) {
	er.mu.Lock()
	defer er.mu.Unlock()
	er.onInput = fn
}

// SetFrameStartCallback sets the callback for frame start.
func (er *EventReplayer) SetFrameStartCallback(fn func(devtools.FrameID)) {
	er.mu.Lock()
	defer er.mu.Unlock()
	er.onFrameStart = fn
}

// SetFrameEndCallback sets the callback for frame end.
func (er *EventReplayer) SetFrameEndCallback(fn func(devtools.FrameID)) {
	er.mu.Lock()
	defer er.mu.Unlock()
	er.onFrameEnd = fn
}

// GetSession returns the current session.
func (er *EventReplayer) GetSession() *RecordingSession {
	er.mu.RLock()
	defer er.mu.RUnlock()
	return er.session
}

// ReplayProgress represents the replay progress.
type ReplayProgress struct {
	CurrentEvent    int
	TotalEvents     int
	PercentComplete float64
	Running         bool
	Paused          bool
	Speed           float64
}

// ReplayResult represents the result of a replay.
type ReplayResult struct {
	Success        bool
	EventsReplayed int
	InputsReplayed int
	Duration       time.Duration
	Errors         []ReplayError
}

// ReplayError represents an error during replay.
type ReplayError struct {
	Event    RecordedEvent
	Message  string
}

// Verify verifies that the replay matches the original recording.
func (er *EventReplayer) Verify(comparison *RecordingSession) *VerificationReport {
	report := &VerificationReport{
		StartTime: time.Now(),
	}

	if er.session == nil || comparison == nil {
		report.Success = false
		report.Errors = append(report.Errors, "nil session")
		return report
	}

	// Verify event count
	if len(er.session.Events) != len(comparison.Events) {
		report.MismatchCount++
		report.Errors = append(report.Errors,
			fmt.Sprintf("event count mismatch: %d vs %d",
				len(er.session.Events), len(comparison.Events)))
	}

	// Verify input count
	if len(er.session.Inputs) != len(comparison.Inputs) {
		report.MismatchCount++
		report.Errors = append(report.Errors,
			fmt.Sprintf("input count mismatch: %d vs %d",
				len(er.session.Inputs), len(comparison.Inputs)))
	}

	// Verify each event
	for i := 0; i < len(er.session.Events) && i < len(comparison.Events); i++ {
		if !eventsEqual(er.session.Events[i], comparison.Events[i]) {
			report.MismatchCount++
			report.Errors = append(report.Errors,
				fmt.Sprintf("event %d mismatch", i))
		} else {
			report.MatchCount++
		}
	}

	report.EndTime = time.Now()
	report.Duration = report.EndTime.Sub(report.StartTime)
	report.Success = report.MismatchCount == 0

	return report
}

// eventsEqual compares two recorded events.
func eventsEqual(a, b RecordedEvent) bool {
	return a.Seq == b.Seq &&
		a.FrameID == b.FrameID &&
		a.Type == b.Type &&
		a.CausalID == b.CausalID
}

// VerificationReport represents a verification report.
type VerificationReport struct {
	StartTime    time.Time
	EndTime      time.Time
	Duration     time.Duration
	Success      bool
	MatchCount   int
	MismatchCount int
	Errors       []string
}

// GetMatchRate returns the match rate as a percentage.
func (vr *VerificationReport) GetMatchRate() float64 {
	total := vr.MatchCount + vr.MismatchCount
	if total == 0 {
		return 100.0
	}
	return float64(vr.MatchCount) / float64(total) * 100
}

// Errors
var (
	ErrNoSessionLoaded    = &ReplayerError{Msg: "no session loaded"}
	ErrAlreadyReplaying   = &ReplayerError{Msg: "already replaying"}
	ErrEndOfSession       = &ReplayerError{Msg: "end of session reached"}
)

// ReplayerError represents a replayer error.
type ReplayerError struct {
	Msg string
}

func (e *ReplayerError) Error() string {
	return e.Msg
}

// ReplaySession represents an active replay session.
type ReplaySession struct {
	ID            string
	Replayer      *EventReplayer
	StartTime     time.Time
	EndTime       time.Time
	CurrentFrame  devtools.FrameID
	TotalFrames   int
	Breakpoints   map[devtools.FrameID]bool
}

// NewReplaySession creates a new replay session.
func NewReplaySession(session *RecordingSession) *ReplaySession {
	return &ReplaySession{
		ID:          fmt.Sprintf("replay_%d", time.Now().Unix()),
		Replayer:    NewEventReplayer(session),
		StartTime:   time.Now(),
		Breakpoints: make(map[devtools.FrameID]bool),
		TotalFrames: len(session.Events),
	}
}

// AddBreakpoint adds a breakpoint at a frame.
func (rs *ReplaySession) AddBreakpoint(frameID devtools.FrameID) {
	rs.Breakpoints[frameID] = true
}

// RemoveBreakpoint removes a breakpoint.
func (rs *ReplaySession) RemoveBreakpoint(frameID devtools.FrameID) {
	delete(rs.Breakpoints, frameID)
}

// HasBreakpoint checks if there's a breakpoint at a frame.
func (rs *ReplaySession) HasBreakpoint(frameID devtools.FrameID) bool {
	return rs.Breakpoints[frameID]
}

// ClearBreakpoints clears all breakpoints.
func (rs *ReplaySession) ClearBreakpoints() {
	rs.Breakpoints = make(map[devtools.FrameID]bool)
}

// GetBreakpoints returns all breakpoints.
func (rs *ReplaySession) GetBreakpoints() []devtools.FrameID {
	frames := make([]devtools.FrameID, 0, len(rs.Breakpoints))
	for frame := range rs.Breakpoints {
		frames = append(frames, frame)
	}
	return frames
}
