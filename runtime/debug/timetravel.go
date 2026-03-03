// Package debug provides debugging tools for Mint UI applications.
//
// The TimeTravelDebugger enables state history navigation for debugging:
//   - Record all state changes
//   - Navigate back and forth through history
//   - Jump to specific points in time
//   - Export/import state snapshots
package debug

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/wwsheng009/mint/runtime/intent"
)

// =============================================================================
// TimeTravelDebugger
// =============================================================================

// Snapshot represents a single state snapshot in time.
type Snapshot struct {
	// Index is the position in history
	Index int
	// Timestamp when the snapshot was taken
	Timestamp time.Time
	// Intent that triggered this state change (optional)
	Intent intent.Intent
	// State is the serialized state
	State json.RawMessage
	// Label is an optional description
	Label string
}

// TimeTravelDebugger enables state history navigation.
// It records all state changes and allows jumping between states.
//
// Usage:
//
//	dbg := debug.NewTimeTravelDebugger[AppState]()
//	store.Subscribe(dbg.Record)
//
//	// Navigate history
//	dbg.Undo()     // Go back one step
//	dbg.Redo()     // Go forward one step
//	dbg.JumpTo(5)  // Jump to specific index
type TimeTravelDebugger[T any] struct {
	mu sync.RWMutex

	// history stores all snapshots
	history []Snapshot

	// currentIndex is the current position in history
	currentIndex int

	// maxHistory limits the number of snapshots stored
	maxHistory int

	// applyState is called to restore a state
	applyState func(T)

	// serialize converts state to JSON
	serialize func(T) (json.RawMessage, error)

	// deserialize converts JSON to state
	deserialize func(json.RawMessage) (T, error)

	// onRecord is called when a new snapshot is recorded
	onRecord func(Snapshot)

	// onJump is called when jumping to a different state
	onJump func(Snapshot)
}

// TimeTravelOption configures the TimeTravelDebugger.
type TimeTravelOption[T any] func(*TimeTravelDebugger[T])

// WithMaxHistory sets the maximum number of snapshots to store.
func WithMaxHistory[T any](max int) TimeTravelOption[T] {
	return func(d *TimeTravelDebugger[T]) {
		d.maxHistory = max
	}
}

// WithApplyState sets the callback to apply a state.
func WithApplyState[T any](fn func(T)) TimeTravelOption[T] {
	return func(d *TimeTravelDebugger[T]) {
		d.applyState = fn
	}
}

// WithSerialization sets custom serialization functions.
func WithSerialization[T any](
	serialize func(T) (json.RawMessage, error),
	deserialize func(json.RawMessage) (T, error),
) TimeTravelOption[T] {
	return func(d *TimeTravelDebugger[T]) {
		d.serialize = serialize
		d.deserialize = deserialize
	}
}

// WithOnRecord sets the callback for when a snapshot is recorded.
func WithOnRecord[T any](fn func(Snapshot)) TimeTravelOption[T] {
	return func(d *TimeTravelDebugger[T]) {
		d.onRecord = fn
	}
}

// WithOnJump sets the callback for when jumping to a state.
func WithOnJump[T any](fn func(Snapshot)) TimeTravelOption[T] {
	return func(d *TimeTravelDebugger[T]) {
		d.onJump = fn
	}
}

// NewTimeTravelDebugger creates a new time travel debugger.
func NewTimeTravelDebugger[T any](opts ...TimeTravelOption[T]) *TimeTravelDebugger[T] {
	d := &TimeTravelDebugger[T]{
		history:    make([]Snapshot, 0),
		maxHistory: 100, // Default: keep last 100 states
		serialize:  defaultSerialize[T],
		deserialize: defaultDeserialize[T],
	}

	for _, opt := range opts {
		opt(d)
	}

	return d
}

// Record captures the current state.
// This is typically called from a Store subscription.
func (d *TimeTravelDebugger[T]) Record(state T) {
	d.RecordWithIntent(state, nil, "")
}

// RecordWithIntent captures the current state with intent context.
func (d *TimeTravelDebugger[T]) RecordWithIntent(state T, i intent.Intent, label string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Serialize state
	stateJSON, err := d.serialize(state)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[TimeTravel] Failed to serialize state: %v\n", err)
		return
	}

	// Create snapshot
	snapshot := Snapshot{
		Index:     len(d.history),
		Timestamp: time.Now(),
		Intent:    i,
		State:     stateJSON,
		Label:     label,
	}

	// If we're not at the end of history, truncate future
	if d.currentIndex < len(d.history)-1 {
		d.history = d.history[:d.currentIndex+1]
	}

	// Add to history
	d.history = append(d.history, snapshot)
	d.currentIndex = len(d.history) - 1

	// Enforce max history limit
	if len(d.history) > d.maxHistory {
		trim := len(d.history) - d.maxHistory
		d.history = d.history[trim:]
		d.currentIndex = len(d.history) - 1
	}

	// Notify
	if d.onRecord != nil {
		d.onRecord(snapshot)
	}
}

// RecordFunc returns a function that can be used as a Store subscriber.
func (d *TimeTravelDebugger[T]) RecordFunc() func(T) {
	return d.Record
}

// Undo goes back one step in history.
func (d *TimeTravelDebugger[T]) Undo() bool {
	return d.JumpTo(d.currentIndex - 1)
}

// Redo goes forward one step in history.
func (d *TimeTravelDebugger[T]) Redo() bool {
	return d.JumpTo(d.currentIndex + 1)
}

// JumpTo jumps to a specific index in history.
func (d *TimeTravelDebugger[T]) JumpTo(index int) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	if index < 0 || index >= len(d.history) {
		return false
	}

	if d.applyState == nil {
		return false
	}

	snapshot := d.history[index]

	// Deserialize state
	state, err := d.deserialize(snapshot.State)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[TimeTravel] Failed to deserialize state: %v\n", err)
		return false
	}

	// Apply state
	d.applyState(state)
	d.currentIndex = index

	// Notify
	if d.onJump != nil {
		d.onJump(snapshot)
	}

	return true
}

// GetCurrentState returns the current state.
func (d *TimeTravelDebugger[T]) GetCurrentState() (T, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.currentIndex < 0 || d.currentIndex >= len(d.history) {
		var zero T
		return zero, fmt.Errorf("no state in history")
	}

	return d.deserialize(d.history[d.currentIndex].State)
}

// GetHistory returns all snapshots.
func (d *TimeTravelDebugger[T]) GetHistory() []Snapshot {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make([]Snapshot, len(d.history))
	copy(result, d.history)
	return result
}

// GetCurrentIndex returns the current position in history.
func (d *TimeTravelDebugger[T]) GetCurrentIndex() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.currentIndex
}

// CanUndo returns true if there's a previous state.
func (d *TimeTravelDebugger[T]) CanUndo() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.currentIndex > 0
}

// CanRedo returns true if there's a next state.
func (d *TimeTravelDebugger[T]) CanRedo() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.currentIndex < len(d.history)-1
}

// Clear clears all history.
func (d *TimeTravelDebugger[T]) Clear() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.history = make([]Snapshot, 0)
	d.currentIndex = -1
}

// Export exports history as JSON.
func (d *TimeTravelDebugger[T]) Export() ([]byte, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return json.MarshalIndent(struct {
		History      []Snapshot `json:"history"`
		CurrentIndex int        `json:"currentIndex"`
	}{
		History:      d.history,
		CurrentIndex: d.currentIndex,
	}, "", "  ")
}

// Import imports history from JSON.
func (d *TimeTravelDebugger[T]) Import(data []byte) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	var imported struct {
		History      []Snapshot `json:"history"`
		CurrentIndex int        `json:"currentIndex"`
	}

	if err := json.Unmarshal(data, &imported); err != nil {
		return err
	}

	d.history = imported.History
	d.currentIndex = imported.CurrentIndex

	return nil
}

// =============================================================================
// Default Serialization
// =============================================================================

func defaultSerialize[T any](state T) (json.RawMessage, error) {
	data, err := json.Marshal(state)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

func defaultDeserialize[T any](data json.RawMessage) (T, error) {
	var state T
	if err := json.Unmarshal(data, &state); err != nil {
		return state, err
	}
	return state, nil
}

// =============================================================================
// Snapshot Methods
// =============================================================================

// GetIntentType returns the intent type name.
func (s Snapshot) GetIntentType() string {
	if s.Intent == nil {
		return "initial"
	}
	return s.Intent.IntentType()
}

// FormatTime returns a formatted timestamp.
func (s Snapshot) FormatTime() string {
	return s.Timestamp.Format("15:04:05.000")
}

// String returns a human-readable representation.
func (s Snapshot) String() string {
	return fmt.Sprintf("[%d] %s - %s (%s)",
		s.Index,
		s.FormatTime(),
		s.GetIntentType(),
		s.Label,
	)
}

// =============================================================================
// Debug Panel Helper
// =============================================================================

// DebugPanelState represents state for a debug panel UI.
type DebugPanelState struct {
	// History is the list of snapshots
	History []Snapshot `json:"history"`
	// CurrentIndex is the current position
	CurrentIndex int `json:"currentIndex"`
	// CanUndo indicates if undo is possible
	CanUndo bool `json:"canUndo"`
	// CanRedo indicates if redo is possible
	CanRedo bool `json:"canRedo"`
}

// GetDebugPanelState returns state for building a debug panel.
func (d *TimeTravelDebugger[T]) GetDebugPanelState() DebugPanelState {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return DebugPanelState{
		History:      d.GetHistory(),
		CurrentIndex: d.currentIndex,
		CanUndo:      d.currentIndex > 0,
		CanRedo:      d.currentIndex < len(d.history)-1,
	}
}
