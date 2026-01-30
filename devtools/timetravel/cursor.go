// Package timetravel provides time travel cursor for DevTools.
//
// This file implements the time travel cursor for navigating
// through frame snapshots.
package timetravel

import (
	"sync"

	"github.com/wwsheng009/mint/devtools"
)

// TimeTravelCursor provides navigation through frame snapshots.
type TimeTravelCursor struct {
	mu        sync.RWMutex
	mgr       *SnapshotManager
	current   *FrameSnapshot
	bookmarks map[string]*FrameSnapshot

	// Event listeners
	onMove func(cursor *TimeTravelCursor)
}

// NewTimeTravelCursor creates a new time travel cursor.
func NewTimeTravelCursor(mgr *SnapshotManager) *TimeTravelCursor {
	return &TimeTravelCursor{
		mgr:       mgr,
		bookmarks: make(map[string]*FrameSnapshot),
	}
}

// SetManager sets the snapshot manager.
func (tc *TimeTravelCursor) SetManager(mgr *SnapshotManager) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.mgr = mgr
}

// GetCurrent returns the current snapshot.
func (tc *TimeTravelCursor) GetCurrent() *FrameSnapshot {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.current
}

// GetManager returns the snapshot manager.
func (tc *TimeTravelCursor) GetManager() *SnapshotManager {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.mgr
}

// MoveToFrame moves the cursor to a specific frame.
func (tc *TimeTravelCursor) MoveToFrame(frameID devtools.FrameID) bool {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	snapshot := tc.mgr.GetSnapshot(frameID)
	if snapshot == nil {
		return false
	}

	tc.current = snapshot
	tc.notifyMove()
	return true
}

// MoveToIndex moves the cursor to a specific index.
func (tc *TimeTravelCursor) MoveToIndex(index int) bool {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	snapshot := tc.mgr.GetSnapshotByIndex(index)
	if snapshot == nil {
		return false
	}

	tc.current = snapshot
	tc.notifyMove()
	return true
}

// MoveToFirst moves the cursor to the first frame.
func (tc *TimeTravelCursor) MoveToFirst() bool {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if tc.mgr.Count() == 0 {
		return false
	}

	snapshot := tc.mgr.GetSnapshotByIndex(0)
	if snapshot == nil {
		return false
	}

	tc.current = snapshot
	tc.notifyMove()
	return true
}

// MoveToLast moves the cursor to the last frame.
func (tc *TimeTravelCursor) MoveToLast() bool {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	count := tc.mgr.Count()
	if count == 0 {
		return false
	}

	snapshot := tc.mgr.GetSnapshotByIndex(count - 1)
	if snapshot == nil {
		return false
	}

	tc.current = snapshot
	tc.notifyMove()
	return true
}

// MoveNext moves the cursor to the next frame.
func (tc *TimeTravelCursor) MoveNext() bool {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if tc.current == nil {
		// Start from first
		snapshot := tc.mgr.GetSnapshotByIndex(0)
		if snapshot == nil {
			return false
		}
		tc.current = snapshot
		tc.notifyMove()
		return true
	}

	// Find next frame
	frameIDs := tc.mgr.GetFrameIDs()
	for i, id := range frameIDs {
		if id == tc.current.FrameID && i+1 < len(frameIDs) {
			snapshot := tc.mgr.GetSnapshot(frameIDs[i+1])
			if snapshot != nil {
				tc.current = snapshot
				tc.notifyMove()
				return true
			}
		}
	}

	return false
}

// MovePrev moves the cursor to the previous frame.
func (tc *TimeTravelCursor) MovePrev() bool {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if tc.current == nil {
		// Start from last
		count := tc.mgr.Count()
		if count == 0 {
			return false
		}
		snapshot := tc.mgr.GetSnapshotByIndex(count - 1)
		if snapshot == nil {
			return false
		}
		tc.current = snapshot
		tc.notifyMove()
		return true
	}

	// Find previous frame
	frameIDs := tc.mgr.GetFrameIDs()
	for i, id := range frameIDs {
		if id == tc.current.FrameID && i > 0 {
			snapshot := tc.mgr.GetSnapshot(frameIDs[i-1])
			if snapshot != nil {
				tc.current = snapshot
				tc.notifyMove()
				return true
			}
		}
	}

	return false
}

// MoveBySteps moves the cursor by a relative number of steps.
func (tc *TimeTravelCursor) MoveBySteps(steps int) bool {
	if steps == 0 {
		return true
	}

	if steps > 0 {
		for i := 0; i < steps; i++ {
			if !tc.MoveNext() {
				return false
			}
		}
	} else {
		for i := 0; i < -steps; i++ {
			if !tc.MovePrev() {
				return false
			}
		}
	}

	return true
}

// JumpToEvent jumps to the frame containing a specific event.
func (tc *TimeTravelCursor) JumpToEvent(eventID devtools.EventID) bool {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	snapshots := tc.mgr.GetAllSnapshots()
	for _, snapshot := range snapshots {
		if snapshot.CausalGraph != nil {
			if event := snapshot.CausalGraph.GetEvent(eventID); event != nil {
				tc.current = snapshot
				tc.notifyMove()
				return true
			}
		}
	}

	return false
}

// JumpToMutation jumps to the frame containing a specific mutation.
func (tc *TimeTravelCursor) JumpToMutation(mutationID devtools.MutationID) bool {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	snapshots := tc.mgr.GetAllSnapshots()
	for _, snapshot := range snapshots {
		if snapshot.CausalGraph != nil {
			if mutation := snapshot.CausalGraph.GetMutation(mutationID); mutation != nil {
				tc.current = snapshot
				tc.notifyMove()
				return true
			}
		}
	}

	return false
}

// JumpToComponentChange jumps to the next frame where a component changed.
func (tc *TimeTravelCursor) JumpToComponentChange(componentID uint32) bool {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	snapshots := tc.mgr.GetAllSnapshots()
	startIndex := 0

	// Find current position
	if tc.current != nil {
		for i, snapshot := range snapshots {
			if snapshot.FrameID == tc.current.FrameID {
				startIndex = i + 1
				break
			}
		}
	}

	// Search forward
	for i := startIndex; i < len(snapshots); i++ {
		if _, exists := snapshots[i].ComponentStates[componentID]; exists {
			tc.current = snapshots[i]
			tc.notifyMove()
			return true
		}
	}

	return false
}

// GetIndex returns the current cursor index.
func (tc *TimeTravelCursor) GetIndex() int {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	if tc.current == nil {
		return -1
	}

	frameIDs := tc.mgr.GetFrameIDs()
	for i, id := range frameIDs {
		if id == tc.current.FrameID {
			return i
		}
	}

	return -1
}

// GetTotalFrames returns the total number of frames.
func (tc *TimeTravelCursor) GetTotalFrames() int {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.mgr.Count()
}

// CanMoveForward returns whether there's a next frame.
func (tc *TimeTravelCursor) CanMoveForward() bool {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	if tc.current == nil {
		return tc.mgr.Count() > 0
	}

	frameIDs := tc.mgr.GetFrameIDs()
	for i, id := range frameIDs {
		if id == tc.current.FrameID {
			return i+1 < len(frameIDs)
		}
	}

	return false
}

// CanMoveBackward returns whether there's a previous frame.
func (tc *TimeTravelCursor) CanMoveBackward() bool {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	if tc.current == nil {
		return tc.mgr.Count() > 0
	}

	frameIDs := tc.mgr.GetFrameIDs()
	for i, id := range frameIDs {
		if id == tc.current.FrameID {
			return i > 0
		}
	}

	return false
}

// AtStart returns whether the cursor is at the first frame.
func (tc *TimeTravelCursor) AtStart() bool {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	if tc.current == nil {
		return true
	}

	frameIDs := tc.mgr.GetFrameIDs()
	return len(frameIDs) > 0 && frameIDs[0] == tc.current.FrameID
}

// AtEnd returns whether the cursor is at the last frame.
func (tc *TimeTravelCursor) AtEnd() bool {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	if tc.current == nil {
		return true
	}

	frameIDs := tc.mgr.GetFrameIDs()
	return len(frameIDs) > 0 && frameIDs[len(frameIDs)-1] == tc.current.FrameID
}

// Bookmark saves the current position as a named bookmark.
func (tc *TimeTravelCursor) Bookmark(name string) bool {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if tc.current == nil {
		return false
	}

	tc.bookmarks[name] = tc.current
	return true
}

// GotoBookmark moves the cursor to a saved bookmark.
func (tc *TimeTravelCursor) GotoBookmark(name string) bool {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	snapshot, exists := tc.bookmarks[name]
	if !exists {
		return false
	}

	tc.current = snapshot
	tc.notifyMove()
	return true
}

// RemoveBookmark removes a saved bookmark.
func (tc *TimeTravelCursor) RemoveBookmark(name string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	delete(tc.bookmarks, name)
}

// GetBookmarks returns all bookmark names.
func (tc *TimeTravelCursor) GetBookmarks() []string {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	bookmarks := make([]string, 0, len(tc.bookmarks))
	for name := range tc.bookmarks {
		bookmarks = append(bookmarks, name)
	}
	return bookmarks
}

// SetOnMove sets a callback function called when the cursor moves.
func (tc *TimeTravelCursor) SetOnMove(fn func(cursor *TimeTravelCursor)) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.onMove = fn
}

// notifyMove calls the move callback if set.
func (tc *TimeTravelCursor) notifyMove() {
	if tc.onMove != nil {
		// Call outside the lock to avoid deadlock
		fn := tc.onMove
		tc.mu.Unlock()
		fn(tc)
		tc.mu.Lock()
	}
}

// GetDiffToNext returns the diff from current to next frame.
func (tc *TimeTravelCursor) GetDiffToNext() *SnapshotDiff {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	if tc.current == nil {
		return nil
	}

	currentIndex := tc.GetIndex()
	if currentIndex < 0 || currentIndex >= tc.mgr.Count()-1 {
		return nil
	}

	nextSnapshot := tc.mgr.GetSnapshotByIndex(currentIndex + 1)
	if nextSnapshot == nil {
		return nil
	}

	return tc.current.Diff(nextSnapshot)
}

// GetDiffToPrev returns the diff from current to previous frame.
func (tc *TimeTravelCursor) GetDiffToPrev() *SnapshotDiff {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	if tc.current == nil {
		return nil
	}

	currentIndex := tc.GetIndex()
	if currentIndex <= 0 {
		return nil
	}

	prevSnapshot := tc.mgr.GetSnapshotByIndex(currentIndex - 1)
	if prevSnapshot == nil {
		return nil
	}

	return prevSnapshot.Diff(tc.current)
}

// SearchByTime searches for frames within a time range.
func (tc *TimeTravelCursor) SearchByTime(start, end int64) []devtools.FrameID {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	snapshots := tc.mgr.GetAllSnapshots()
	var results []devtools.FrameID

	for _, snapshot := range snapshots {
		timestamp := snapshot.Timestamp.Unix()
		if timestamp >= start && timestamp <= end {
			results = append(results, snapshot.FrameID)
		}
	}

	return results
}

// SearchByComponent searches for frames where a component appears.
func (tc *TimeTravelCursor) SearchByComponent(componentID uint32) []devtools.FrameID {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	snapshots := tc.mgr.GetAllSnapshots()
	var results []devtools.FrameID

	for _, snapshot := range snapshots {
		if _, exists := snapshot.ComponentStates[componentID]; exists {
			results = append(results, snapshot.FrameID)
		}
	}

	return results
}

// CursorInfo provides information about the cursor state.
type CursorInfo struct {
	FrameID      devtools.FrameID
	Index        int
	TotalFrames  int
	CanMoveForward bool
	CanMoveBackward bool
	AtStart      bool
	AtEnd        bool
	Bookmarks    []string
}

// GetInfo returns information about the cursor state.
func (tc *TimeTravelCursor) GetInfo() *CursorInfo {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	info := &CursorInfo{
		TotalFrames:  tc.mgr.Count(),
		CanMoveForward: tc.CanMoveForward(),
		CanMoveBackward: tc.CanMoveBackward(),
		AtStart:      tc.AtStart(),
		AtEnd:        tc.AtEnd(),
	}

	if tc.current != nil {
		info.FrameID = tc.current.FrameID
		info.Index = tc.GetIndex()
	}

	info.Bookmarks = tc.GetBookmarks()

	return info
}
