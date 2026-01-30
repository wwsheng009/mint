// Package replay provides deterministic checking for DevTools.
//
// This file implements determinism verification for replay.
package replay

import (
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/wwsheng009/mint/devtools"
)

// DeterminismChecker verifies that replays are deterministic.
type DeterminismChecker struct {
	mu                sync.RWMutex
	originalSession   *RecordingSession
	replaySession     *RecordingSession
	checkpoints       map[devtools.FrameID]*Checkpoint
	enabled           bool

	// State tracking
	stateHashes       map[devtools.FrameID]string
	eventHashes       map[uint64]string
}

// Checkpoint represents a state checkpoint.
type Checkpoint struct {
	FrameID     devtools.FrameID
	Timestamp   time.Time
	StateHash   string
	EventHashes []EventHashEntry
	Metadata    map[string]interface{}
}

// EventHashEntry represents a hash of an event.
type EventHashEntry struct {
	EventID uint64
	Hash    string
}

// NewDeterminismChecker creates a new determinism checker.
func NewDeterminismChecker(original *RecordingSession) *DeterminismChecker {
	return &DeterminismChecker{
		originalSession: original,
		checkpoints:     make(map[devtools.FrameID]*Checkpoint),
		stateHashes:     make(map[devtools.FrameID]string),
		eventHashes:     make(map[uint64]string),
	}
}

// Enable enables the checker.
func (dc *DeterminismChecker) Enable() {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	dc.enabled = true
}

// Disable disables the checker.
func (dc *DeterminismChecker) Disable() {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	dc.enabled = false
}

// IsEnabled returns whether the checker is enabled.
func (dc *DeterminismChecker) IsEnabled() bool {
	dc.mu.RLock()
	defer dc.mu.RUnlock()
	return dc.enabled
}

// RecordCheckpoint records a checkpoint for a frame.
func (dc *DeterminismChecker) RecordCheckpoint(frameID devtools.FrameID, stateData []byte) {
	if !dc.IsEnabled() {
		return
	}

	dc.mu.Lock()
	defer dc.mu.Unlock()

	hash := hashBytes(stateData)
	dc.stateHashes[frameID] = hash

	checkpoint := &Checkpoint{
		FrameID:     frameID,
		Timestamp:   time.Now(),
		StateHash:   hash,
		EventHashes: make([]EventHashEntry, 0),
	}

	// Collect event hashes for this frame
	for _, event := range dc.originalSession.Events {
		if event.FrameID == frameID {
			eventHash := hashEventData(event)
			dc.eventHashes[event.CausalID] = eventHash
			checkpoint.EventHashes = append(checkpoint.EventHashes, EventHashEntry{
				EventID: event.CausalID,
				Hash:    eventHash,
			})
		}
	}

	dc.checkpoints[frameID] = checkpoint
}

// VerifyCheckpoint verifies a checkpoint against the original.
func (dc *DeterminismChecker) VerifyCheckpoint(frameID devtools.FrameID, stateData []byte) *VerificationResult {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	result := &VerificationResult{
		FrameID:  frameID,
		Verified: time.Now(),
	}

	// Get original checkpoint
	original, exists := dc.checkpoints[frameID]
	if !exists {
		result.Success = false
		result.Error = "no original checkpoint found"
		return result
	}

	// Compare state hash
	currentHash := hashBytes(stateData)
	result.StateMatch = currentHash == original.StateHash

	if !result.StateMatch {
		result.Error = fmt.Sprintf("state hash mismatch: expected %s, got %s",
			original.StateHash, currentHash)
	}

	result.Success = result.StateMatch
	return result
}

// VerifyFull verifies the entire replay session.
func (dc *DeterminismChecker) VerifyFull(replaySession *RecordingSession) *FullVerificationReport {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	report := &FullVerificationReport{
		StartTime: time.Now(),
	}

	// Verify event count
	report.OriginalEventCount = len(dc.originalSession.Events)
	report.ReplayEventCount = len(replaySession.Events)
	report.EventCountMatch = report.OriginalEventCount == report.ReplayEventCount

	// Verify input count
	report.OriginalInputCount = len(dc.originalSession.Inputs)
	report.ReplayInputCount = len(replaySession.Inputs)
	report.InputCountMatch = report.OriginalInputCount == report.ReplayInputCount

	// Verify each checkpoint
	for frameID, checkpoint := range dc.checkpoints {
		frameReport := &FrameVerificationReport{
			FrameID:    devtools.FrameID(frameID),
			StateMatch: true, // Assume match until proven otherwise
		}

		// Verify state hash
		if stateHash, exists := dc.stateHashes[devtools.FrameID(frameID)]; exists {
			frameReport.StateMatch = stateHash == checkpoint.StateHash
			if !frameReport.StateMatch {
				report.MismatchCount++
			} else {
				report.MatchCount++
			}
		}

		// Verify event hashes
		for _, eventEntry := range checkpoint.EventHashes {
			if eventHash, exists := dc.eventHashes[eventEntry.EventID]; exists {
				if eventHash != eventEntry.Hash {
					frameReport.EventMismatches++
				}
			}
		}

		report.Frames = append(report.Frames, frameReport)
	}

	report.EndTime = time.Now()
	report.Duration = report.EndTime.Sub(report.StartTime)
	report.Success = report.MismatchCount == 0 &&
		report.EventCountMatch &&
		report.InputCountMatch

	return report
}

// GetCheckpoint returns a checkpoint by frame ID.
func (dc *DeterminismChecker) GetCheckpoint(frameID devtools.FrameID) *Checkpoint {
	dc.mu.RLock()
	defer dc.mu.RUnlock()
	return dc.checkpoints[frameID]
}

// GetAllCheckpoints returns all checkpoints.
func (dc *DeterminismChecker) GetAllCheckpoints() []*Checkpoint {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	checkpoints := make([]*Checkpoint, 0, len(dc.checkpoints))
	for _, cp := range dc.checkpoints {
		checkpoints = append(checkpoints, cp)
	}
	return checkpoints
}

// GetStateHash returns the state hash for a frame.
func (dc *DeterminismChecker) GetStateHash(frameID devtools.FrameID) (string, bool) {
	dc.mu.RLock()
	defer dc.mu.RUnlock()
	hash, exists := dc.stateHashes[frameID]
	return hash, exists
}

// CompareSessions compares two sessions for determinism.
func (dc *DeterminismChecker) CompareSessions(session1, session2 *RecordingSession) *ComparisonReport {
	report := &ComparisonReport{
		StartTime: time.Now(),
	}

	// Compare metadata
	report.Session1ID = session1.SessionID
	report.Session2ID = session2.SessionID

	// Compare event counts
	report.EventCountDelta = len(session1.Events) - len(session2.Events)

	// Compare input counts
	report.InputCountDelta = len(session1.Inputs) - len(session2.Inputs)

	// Compare durations
	report.DurationDelta = session1.GetDuration() - session2.GetDuration()

	// Find first divergence point
	minLen := len(session1.Events)
	if len(session2.Events) < minLen {
		minLen = len(session2.Events)
	}

	for i := 0; i < minLen; i++ {
		if !eventsEqual(session1.Events[i], session2.Events[i]) {
			report.FirstDivergence = i
			break
		}
	}

	report.EndTime = time.Now()
	report.Duration = report.EndTime.Sub(report.StartTime)

	return report
}

// hashBytes creates a hash of byte data.
func hashBytes(data []byte) string {
	// Simple hash function for demonstration
	// In production, use a proper hash like SHA-256
	if len(data) == 0 {
		return "empty"
	}
	// Use hex encoding of first 32 bytes as a simple hash
	maxLen := 32
	if len(data) < maxLen {
		maxLen = len(data)
	}
	return hex.EncodeToString(data[:maxLen])
}

// hashEventData creates a hash of event data.
func hashEventData(event RecordedEvent) string {
	// Create a hash based on key event properties
	data := fmt.Sprintf("%d:%s:%d:%v", event.Seq, event.Type, event.FrameID, event.Data)
	return hashBytes([]byte(data))
}

// VerificationResult represents a single verification result.
type VerificationResult struct {
	FrameID    devtools.FrameID
	Verified   time.Time
	Success    bool
	StateMatch bool
	Error      string
}

// FullVerificationReport represents a full verification report.
type FullVerificationReport struct {
	StartTime            time.Time
	EndTime              time.Time
	Duration             time.Duration
	Success              bool
	OriginalEventCount   int
	ReplayEventCount     int
	EventCountMatch      bool
	OriginalInputCount   int
	ReplayInputCount     int
	InputCountMatch      bool
	MatchCount           int
	MismatchCount        int
	Frames               []*FrameVerificationReport
}

// FrameVerificationReport represents verification for a single frame.
type FrameVerificationReport struct {
	FrameID           devtools.FrameID
	StateMatch        bool
	EventMismatches   int
}

// ComparisonReport represents a comparison of two sessions.
type ComparisonReport struct {
	StartTime         time.Time
	EndTime           time.Time
	Duration          time.Duration
	Session1ID        string
	Session2ID        string
	EventCountDelta   int
	InputCountDelta   int
	DurationDelta     time.Duration
	FirstDivergence   int
}

// DeterminismReport provides a comprehensive determinism report.
type DeterminismReport struct {
	SessionID      string
	StartTime      time.Time
	EndTime        time.Time
	TotalFrames    int
	DeterministicFrames int
	NonDeterministicFrames int
	DeterminismRate float64
	Issues         []DeterminismIssue
}

// DeterminismIssue represents a determinism issue.
type DeterminismIssue struct {
	FrameID      devtools.FrameID
	Type         IssueType
	Description  string
	Severity     Severity
}

// IssueType represents the type of issue.
type IssueType int

const (
	// IssueStateMismatch indicates state mismatch.
	IssueStateMismatch IssueType = iota
	// IssueEventMismatch indicates event mismatch.
	IssueEventMismatch
	// IssueSeedMismatch indicates seed mismatch.
	IssueSeedMismatch
	// IssueInputMismatch indicates input mismatch.
	IssueInputMismatch
)

// Severity represents the severity of an issue.
type Severity int

const (
	// SeverityLow indicates low severity.
	SeverityLow Severity = iota
	// SeverityMedium indicates medium severity.
	SeverityMedium
	// SeverityHigh indicates high severity.
	SeverityHigh
)

// GenerateReport generates a determinism report.
func (dc *DeterminismChecker) GenerateReport() *DeterminismReport {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	report := &DeterminismReport{
		SessionID:   dc.originalSession.SessionID,
		StartTime:   dc.originalSession.StartTime,
		EndTime:     dc.originalSession.EndTime,
		TotalFrames: len(dc.checkpoints),
		Issues:      make([]DeterminismIssue, 0),
	}

	for frameID, checkpoint := range dc.checkpoints {
		if stateHash, exists := dc.stateHashes[frameID]; exists {
			if stateHash == checkpoint.StateHash {
				report.DeterministicFrames++
			} else {
				report.NonDeterministicFrames++
				report.Issues = append(report.Issues, DeterminismIssue{
					FrameID:     devtools.FrameID(frameID),
					Type:        IssueStateMismatch,
					Description: fmt.Sprintf("state hash changed at frame %d", frameID),
					Severity:    SeverityHigh,
				})
			}
		}
	}

	if report.TotalFrames > 0 {
		report.DeterminismRate = float64(report.DeterministicFrames) / float64(report.TotalFrames) * 100
	}

	return report
}

// Reset resets the checker.
func (dc *DeterminismChecker) Reset() {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	dc.checkpoints = make(map[devtools.FrameID]*Checkpoint)
	dc.stateHashes = make(map[devtools.FrameID]string)
	dc.eventHashes = make(map[uint64]string)
}

// Stats returns statistics about the checker.
func (dc *DeterminismChecker) Stats() *CheckerStats {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	return &CheckerStats{
		Enabled:         dc.enabled,
		CheckpointCount: len(dc.checkpoints),
		StateHashCount:  len(dc.stateHashes),
		EventHashCount:  len(dc.eventHashes),
	}
}

// CheckerStats represents checker statistics.
type CheckerStats struct {
	Enabled         bool
	CheckpointCount int
	StateHashCount  int
	EventHashCount  int
}
