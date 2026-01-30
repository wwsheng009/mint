// Package replay provides random seed tracking for deterministic replay.
//
// This file implements seed tracking for reproducible randomness.
package replay

import (
	"sync"
	"time"

	"github.com/wwsheng009/mint/devtools"
)

// SeedTracker tracks random seeds for deterministic replay.
type SeedTracker struct {
	mu    sync.RWMutex
	seeds map[devtools.FrameID]*SeedSnapshot
}

// SeedSnapshot represents a snapshot of random state.
type SeedSnapshot struct {
	FrameID   devtools.FrameID
	Timestamp time.Time
	Source    string // "math/rand", "crypto/rand", etc.
	Value     int64
	State     []byte // Full state if needed
}

// NewSeedTracker creates a new seed tracker.
func NewSeedTracker() *SeedTracker {
	return &SeedTracker{
		seeds: make(map[devtools.FrameID]*SeedSnapshot),
	}
}

// Capture captures the current random seed for a frame.
func (st *SeedTracker) Capture(frameID devtools.FrameID, source string, seed int64) {
	st.mu.Lock()
	defer st.mu.Unlock()

	st.seeds[frameID] = &SeedSnapshot{
		FrameID:   frameID,
		Timestamp: time.Now(),
		Source:    source,
		Value:     seed,
	}
}

// CaptureWithState captures seed with full state.
func (st *SeedTracker) CaptureWithState(frameID devtools.FrameID, source string, seed int64, state []byte) {
	st.mu.Lock()
	defer st.mu.Unlock()

	st.seeds[frameID] = &SeedSnapshot{
		FrameID:   frameID,
		Timestamp: time.Now(),
		Source:    source,
		Value:     seed,
		State:     state,
	}
}

// GetSeedForFrame returns the seed for a specific frame.
func (st *SeedTracker) GetSeedForFrame(frameID devtools.FrameID) *SeedSnapshot {
	st.mu.RLock()
	defer st.mu.RUnlock()

	return st.seeds[frameID]
}

// ApplySeed applies a seed to the random source.
func (st *SeedTracker) ApplySeed(source string, seed int64) {
	// This would interface with the actual random source
	// For now, it's a placeholder for the seed application logic
	_ = source
	_ = seed
}

// GetAllSeeds returns all captured seeds.
func (st *SeedTracker) GetAllSeeds() []*SeedSnapshot {
	st.mu.RLock()
	defer st.mu.RUnlock()

	seeds := make([]*SeedSnapshot, 0, len(st.seeds))
	for _, seed := range st.seeds {
		seeds = append(seeds, seed)
	}
	return seeds
}

// GetSeedsForSource returns seeds for a specific source.
func (st *SeedTracker) GetSeedsForSource(source string) []*SeedSnapshot {
	st.mu.RLock()
	defer st.mu.RUnlock()

	var seeds []*SeedSnapshot
	for _, seed := range st.seeds {
		if seed.Source == source {
			seeds = append(seeds, seed)
		}
	}
	return seeds
}

// Clear clears all captured seeds.
func (st *SeedTracker) Clear() {
	st.mu.Lock()
	defer st.mu.Unlock()

	st.seeds = make(map[devtools.FrameID]*SeedSnapshot)
}

// SaveToSession saves seeds to a recording session.
func (st *SeedTracker) SaveToSession(session *RecordingSession) {
	st.mu.RLock()
	defer st.mu.RUnlock()

	session.Seeds = make([]SeedSnapshot, 0, len(st.seeds))
	for _, seed := range st.seeds {
		session.Seeds = append(session.Seeds, *seed)
	}
}

// LoadFromSession loads seeds from a recording session.
func (st *SeedTracker) LoadFromSession(session *RecordingSession) {
	st.mu.Lock()
	defer st.mu.Unlock()

	st.seeds = make(map[devtools.FrameID]*SeedSnapshot)
	for i := range session.Seeds {
		seed := session.Seeds[i]
		st.seeds[seed.FrameID] = &seed
	}
}

// SeedHistory maintains a history of seed changes.
type SeedHistory struct {
	mu      sync.RWMutex
	history []SeedChange
}

// SeedChange represents a change in seed value.
type SeedChange struct {
	Timestamp time.Time
	FrameID   devtools.FrameID
	Source    string
	OldValue  int64
	NewValue  int64
}

// NewSeedHistory creates a new seed history.
func NewSeedHistory() *SeedHistory {
	return &SeedHistory{
		history: make([]SeedChange, 0, 1024),
	}
}

// Record records a seed change.
func (sh *SeedHistory) Record(frameID devtools.FrameID, source string, oldValue, newValue int64) {
	sh.mu.Lock()
	defer sh.mu.Unlock()

	sh.history = append(sh.history, SeedChange{
		Timestamp: time.Now(),
		FrameID:   frameID,
		Source:    source,
		OldValue:  oldValue,
		NewValue:  newValue,
	})
}

// GetHistory returns the full history.
func (sh *SeedHistory) GetHistory() []SeedChange {
	sh.mu.RLock()
	defer sh.mu.RUnlock()

	history := make([]SeedChange, len(sh.history))
	copy(history, sh.history)
	return history
}

// GetHistoryForFrame returns history for a specific frame.
func (sh *SeedHistory) GetHistoryForFrame(frameID devtools.FrameID) []SeedChange {
	sh.mu.RLock()
	defer sh.mu.RUnlock()

	var changes []SeedChange
	for _, change := range sh.history {
		if change.FrameID == frameID {
			changes = append(changes, change)
		}
	}
	return changes
}

// GetHistoryForSource returns history for a specific source.
func (sh *SeedHistory) GetHistoryForSource(source string) []SeedChange {
	sh.mu.RLock()
	defer sh.mu.RUnlock()

	var changes []SeedChange
	for _, change := range sh.history {
		if change.Source == source {
			changes = append(changes, change)
		}
	}
	return changes
}

// Clear clears the history.
func (sh *SeedHistory) Clear() {
	sh.mu.Lock()
	defer sh.mu.Unlock()

	sh.history = make([]SeedChange, 0, 1024)
}

// DeterministicSeed provides a deterministic seed generator.
type DeterministicSeed struct {
	current int64
}

// NewDeterministicSeed creates a new deterministic seed generator.
func NewDeterministicSeed(initial int64) *DeterministicSeed {
	return &DeterministicSeed{
		current: initial,
	}
}

// Next returns the next seed value.
func (ds *DeterministicSeed) Next() int64 {
	ds.current++
	// Use a simple LCG for reproducibility
	ds.current = (ds.current * 1103515245 + 12345) & 0x7fffffff
	return ds.current
}

// Reset resets the seed generator.
func (ds *DeterministicSeed) Reset() {
	ds.current = 0
}

// SetCurrent sets the current seed value.
func (ds *DeterministicSeed) SetCurrent(value int64) {
	ds.current = value
}

// GetCurrent returns the current seed value.
func (ds *DeterministicSeed) GetCurrent() int64 {
	return ds.current
}

// SeedSource constants.
const (
	SeedSourceMathRand   = "math/rand"
	SeedSourceCryptoRand = "crypto/rand"
	SeedSourceCustom     = "custom"
)
