// Package timetravel provides state replay capabilities for DevTools.
//
// This file implements state replay from snapshots for time travel debugging.
package timetravel

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/wwsheng009/mint/devtools"
)

// ReplayEngine manages replaying state from snapshots.
type ReplayEngine struct {
	mu        sync.RWMutex
	mgr       *SnapshotManager
	cursor    *TimeTravelCursor

	// Replay state
	isReplaying    bool
	replaySpeed    float64 // 1.0 = normal speed
	replayCallback func(frameID devtools.FrameID, snapshot *FrameSnapshot)
}

// NewReplayEngine creates a new replay engine.
func NewReplayEngine(mgr *SnapshotManager, cursor *TimeTravelCursor) *ReplayEngine {
	re := &ReplayEngine{
		mgr:        mgr,
		cursor:     cursor,
		replaySpeed: 1.0,
	}
	return re
}

// SetReplayCallback sets the callback for replay events.
func (re *ReplayEngine) SetReplayCallback(fn func(frameID devtools.FrameID, snapshot *FrameSnapshot)) {
	re.mu.Lock()
	defer re.mu.Unlock()
	re.replayCallback = fn
}

// SetReplaySpeed sets the replay speed multiplier.
func (re *ReplayEngine) SetReplaySpeed(speed float64) {
	re.mu.Lock()
	defer re.mu.Unlock()
	re.replaySpeed = speed
}

// GetReplaySpeed returns the current replay speed.
func (re *ReplayEngine) GetReplaySpeed() float64 {
	re.mu.RLock()
	defer re.mu.RUnlock()
	return re.replaySpeed
}

// IsReplaying returns whether replay is in progress.
func (re *ReplayEngine) IsReplaying() bool {
	re.mu.RLock()
	defer re.mu.RUnlock()
	return re.isReplaying
}

// ReplayFrom replays from the specified frame.
func (re *ReplayEngine) ReplayFrom(frameID devtools.FrameID) error {
	re.mu.Lock()
	defer re.mu.Unlock()

	snapshot := re.mgr.GetSnapshot(frameID)
	if snapshot == nil {
		return fmt.Errorf("frame %d not found", frameID)
	}

	re.isReplaying = true

	// Move cursor to start frame
	re.cursor.MoveToFrame(frameID)

	// Call callback for start frame
	re.notifyFrame(frameID, snapshot)

	// Replay forward through all subsequent frames
	frameIDs := re.mgr.GetFrameIDs()
	startIndex := -1
	for i, id := range frameIDs {
		if id == frameID {
			startIndex = i
			break
		}
	}

	if startIndex == -1 {
		re.isReplaying = false
		return fmt.Errorf("frame %d not found in index", frameID)
	}

	for i := startIndex + 1; i < len(frameIDs); i++ {
		if !re.isReplaying {
			break
		}

		nextSnapshot := re.mgr.GetSnapshot(frameIDs[i])
		if nextSnapshot != nil {
			re.cursor.MoveToFrame(frameIDs[i])
			re.notifyFrame(frameIDs[i], nextSnapshot)
		}
	}

	re.isReplaying = false
	return nil
}

// ReplayRange replays a range of frames.
func (re *ReplayEngine) ReplayRange(startID, endID devtools.FrameID) error {
	re.mu.Lock()
	defer re.mu.Unlock()

	frameIDs := re.mgr.GetFrameIDs()

	// Find range
	startIndex := -1
	endIndex := -1
	for i, id := range frameIDs {
		if id == startID {
			startIndex = i
		}
		if id == endID {
			endIndex = i
		}
	}

	if startIndex == -1 || endIndex == -1 {
		return fmt.Errorf("frame range not found")
	}

	if startIndex > endIndex {
		startIndex, endIndex = endIndex, startIndex
	}

	re.isReplaying = true

	for i := startIndex; i <= endIndex; i++ {
		if !re.isReplaying {
			break
		}

		snapshot := re.mgr.GetSnapshot(frameIDs[i])
		if snapshot != nil {
			re.cursor.MoveToFrame(frameIDs[i])
			re.notifyFrame(frameIDs[i], snapshot)
		}
	}

	re.isReplaying = false
	return nil
}

// ReplayTo replays from current position to the specified frame.
func (re *ReplayEngine) ReplayTo(frameID devtools.FrameID) error {
	current := re.cursor.GetCurrent()
	if current == nil {
		return fmt.Errorf("no current frame")
	}

	return re.ReplayRange(current.FrameID, frameID)
}

// Stop stops the current replay.
func (re *ReplayEngine) Stop() {
	re.mu.Lock()
	defer re.mu.Unlock()
	re.isReplaying = false
}

// notifyFrame notifies the replay callback for a frame.
func (re *ReplayEngine) notifyFrame(frameID devtools.FrameID, snapshot *FrameSnapshot) {
	if re.replayCallback != nil {
		re.replayCallback(frameID, snapshot)
	}
}

// StepForward moves one frame forward in replay.
func (re *ReplayEngine) StepForward() bool {
	return re.cursor.MoveNext()
}

// StepBackward moves one frame backward in replay.
func (re *ReplayEngine) StepBackward() bool {
	return re.cursor.MovePrev()
}

// StateApplier applies snapshot state to components.
type StateApplier interface {
	ApplyComponentState(componentID uint32, state *ComponentState) error
	ApplyLayoutState(layout *LayoutSnapshot) error
	ApplyRepaintState(repaint *RepaintSnapshot) error
}

// ReplaySession manages a single replay session.
type ReplaySession struct {
	mu       sync.RWMutex
	engine   *ReplayEngine
	applier  StateApplier
	startID  devtools.FrameID
	endID    devtools.FrameID
	currentID devtools.FrameID
}

// NewReplaySession creates a new replay session.
func NewReplaySession(engine *ReplayEngine, applier StateApplier) *ReplaySession {
	return &ReplaySession{
		engine:  engine,
		applier: applier,
	}
}

// Start starts the replay session.
func (rs *ReplaySession) Start(frameID devtools.FrameID) error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	rs.startID = frameID
	rs.currentID = frameID

	snapshot := rs.engine.mgr.GetSnapshot(frameID)
	if snapshot == nil {
		return fmt.Errorf("frame %d not found", frameID)
	}

	return rs.applySnapshot(snapshot)
}

// Next advances to the next frame.
func (rs *ReplaySession) Next() error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	frameIDs := rs.engine.mgr.GetFrameIDs()
	currentIndex := -1

	for i, id := range frameIDs {
		if id == rs.currentID {
			currentIndex = i
			break
		}
	}

	if currentIndex == -1 || currentIndex+1 >= len(frameIDs) {
		return fmt.Errorf("no next frame")
	}

	nextID := frameIDs[currentIndex+1]
	rs.currentID = nextID

	snapshot := rs.engine.mgr.GetSnapshot(nextID)
	if snapshot == nil {
		return fmt.Errorf("frame %d not found", nextID)
	}

	return rs.applySnapshot(snapshot)
}

// Prev goes back to the previous frame.
func (rs *ReplaySession) Prev() error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	frameIDs := rs.engine.mgr.GetFrameIDs()
	currentIndex := -1

	for i, id := range frameIDs {
		if id == rs.currentID {
			currentIndex = i
			break
		}
	}

	if currentIndex <= 0 {
		return fmt.Errorf("no previous frame")
	}

	prevID := frameIDs[currentIndex-1]
	rs.currentID = prevID

	snapshot := rs.engine.mgr.GetSnapshot(prevID)
	if snapshot == nil {
		return fmt.Errorf("frame %d not found", prevID)
	}

	return rs.applySnapshot(snapshot)
}

// JumpTo jumps to a specific frame.
func (rs *ReplaySession) JumpTo(frameID devtools.FrameID) error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	snapshot := rs.engine.mgr.GetSnapshot(frameID)
	if snapshot == nil {
		return fmt.Errorf("frame %d not found", frameID)
	}

	rs.currentID = frameID
	return rs.applySnapshot(snapshot)
}

// GetCurrentID returns the current frame ID.
func (rs *ReplaySession) GetCurrentID() devtools.FrameID {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.currentID
}

// GetStartID returns the start frame ID.
func (rs *ReplaySession) GetStartID() devtools.FrameID {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.startID
}

// applySnapshot applies a snapshot to the applier.
func (rs *ReplaySession) applySnapshot(snapshot *FrameSnapshot) error {
	if rs.applier == nil {
		return nil
	}

	// Apply component states
	for _, state := range snapshot.ComponentStates {
		if err := rs.applier.ApplyComponentState(state.ComponentID, state); err != nil {
			return fmt.Errorf("failed to apply component state: %w", err)
		}
	}

	// Apply layout state
	if snapshot.LayoutState != nil {
		if err := rs.applier.ApplyLayoutState(snapshot.LayoutState); err != nil {
			return fmt.Errorf("failed to apply layout state: %w", err)
		}
	}

	// Apply repaint state
	if snapshot.RepaintState != nil {
		if err := rs.applier.ApplyRepaintState(snapshot.RepaintState); err != nil {
			return fmt.Errorf("failed to apply repaint state: %w", err)
		}
	}

	return nil
}

// ExportSnapshot exports a snapshot to JSON.
func (rs *ReplaySession) ExportSnapshot(frameID devtools.FrameID) ([]byte, error) {
	snapshot := rs.engine.mgr.GetSnapshot(frameID)
	if snapshot == nil {
		return nil, fmt.Errorf("frame %d not found", frameID)
	}

	return json.MarshalIndent(snapshot, "", "  ")
}

// ImportSnapshot imports a snapshot from JSON.
func (rs *ReplaySession) ImportSnapshot(data []byte) (*FrameSnapshot, error) {
	var snapshot FrameSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, fmt.Errorf("failed to unmarshal snapshot: %w", err)
	}
	return &snapshot, nil
}

// ExportRange exports a range of frames as JSON.
func (rs *ReplaySession) ExportRange(startID, endID devtools.FrameID) ([]byte, error) {
	frameIDs := rs.engine.mgr.GetFrameIDs()

	startIndex := -1
	endIndex := -1
	for i, id := range frameIDs {
		if id == startID {
			startIndex = i
		}
		if id == endID {
			endIndex = i
		}
	}

	if startIndex == -1 || endIndex == -1 {
		return nil, fmt.Errorf("frame range not found")
	}

	if startIndex > endIndex {
		startIndex, endIndex = endIndex, startIndex
	}

	var snapshots []*FrameSnapshot
	for i := startIndex; i <= endIndex; i++ {
		snapshot := rs.engine.mgr.GetSnapshot(frameIDs[i])
		if snapshot != nil {
			snapshots = append(snapshots, snapshot)
		}
	}

	return json.MarshalIndent(snapshots, "", "  ")
}

// ReplayStats provides statistics about a replay session.
type ReplayStats struct {
	TotalFrames    int
	CurrentFrame   int
	StartFrameID   devtools.FrameID
	EndFrameID     devtools.FrameID
	ReplaySpeed    float64
}

// GetStats returns replay statistics.
func (rs *ReplaySession) GetStats() *ReplayStats {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	stats := &ReplayStats{
		TotalFrames:  rs.engine.mgr.Count(),
		ReplaySpeed:  rs.engine.GetReplaySpeed(),
		StartFrameID: rs.startID,
	}

	frameIDs := rs.engine.mgr.GetFrameIDs()
	for i, id := range frameIDs {
		if id == rs.currentID {
			stats.CurrentFrame = i
		}
		if id == rs.startID {
			stats.StartFrameID = id
		}
	}

	if len(frameIDs) > 0 {
		stats.EndFrameID = frameIDs[len(frameIDs)-1]
	}

	return stats
}
