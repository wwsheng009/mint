// Package snapshot provides state snapshot and restoration for DevTools.
package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/wwsheng009/mint/devtools"
)

// =============================================================================
// Snapshot Manager
// =============================================================================

// Manager manages snapshots for DevTools.
type Manager struct {
	mu            sync.RWMutex
	snapshots     map[devtools.FrameID]*Snapshot
	snapshotsByID map[SnapshotID]*Snapshot
	ordered       []devtools.FrameID  // Ordered frame IDs for iteration
	pool          *SnapshotPool
	maxSnapshots  int
	autoCapture        bool
	autoInterval       time.Duration
	lastAutoFrame      devtools.FrameID
	hasAutoCapture     bool // true after first NoteAutoCapture call
	persistDir         string  // Directory for persistent storage
	enabled            bool
}

// NewManager creates a new snapshot manager.
func NewManager(maxSnapshots int) *Manager {
	return &Manager{
		snapshots:     make(map[devtools.FrameID]*Snapshot),
		snapshotsByID: make(map[SnapshotID]*Snapshot),
		ordered:       make([]devtools.FrameID, 0, maxSnapshots),
		pool:          NewSnapshotPool(maxSnapshots),
		maxSnapshots:  maxSnapshots,
		autoCapture:   false,
		autoInterval:  0,
		enabled:       true,
	}
}

// Enable enables snapshot management.
func (m *Manager) Enable() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = true
}

// Disable disables snapshot management.
func (m *Manager) Disable() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = false
}

// IsEnabled returns true if snapshot management is enabled.
func (m *Manager) IsEnabled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.enabled
}

// =============================================================================
// Snapshot Capture
// =============================================================================

// Capture captures a snapshot at the current frame.
func (m *Manager) Capture(frameID devtools.FrameID, builder *Builder) (*Snapshot, error) {
	if !m.IsEnabled() {
		return nil, fmt.Errorf("snapshot manager is disabled")
	}

	snap := builder.Build()

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if snapshot already exists for this frame
	if _, exists := m.snapshots[frameID]; exists {
		return nil, fmt.Errorf("snapshot already exists for frame %d", frameID)
	}

	// Add to maps
	m.snapshots[frameID] = snap
	m.snapshotsByID[snap.ID] = snap
	m.ordered = append(m.ordered, frameID)

	// Enforce max snapshots limit
	if len(m.ordered) > m.maxSnapshots {
		// Remove oldest
	 oldestFrame := m.ordered[0]
		m.removeSnapshot(oldestFrame)
		m.ordered = m.ordered[1:]
	}

	return snap, nil
}

// CaptureQuick captures a snapshot with minimal data.
func (m *Manager) CaptureQuick(frameID devtools.FrameID) (*Snapshot, error) {
	snapID := SnapshotID(fmt.Sprintf("snap-%d", frameID))
	builder := NewBuilder(snapID, frameID)
	return m.Capture(frameID, builder)
}

// removeSnapshot removes a snapshot without locking (internal use).
func (m *Manager) removeSnapshot(frameID devtools.FrameID) {
	if snap, exists := m.snapshots[frameID]; exists {
		delete(m.snapshots, frameID)
		delete(m.snapshotsByID, snap.ID)
		m.pool.Release(snap)
	}
}

// =============================================================================
// Snapshot Retrieval
// =============================================================================

// Get retrieves a snapshot by frame ID.
func (m *Manager) Get(frameID devtools.FrameID) (*Snapshot, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	snap, exists := m.snapshots[frameID]
	return snap, exists
}

// GetByID retrieves a snapshot by snapshot ID.
func (m *Manager) GetByID(id SnapshotID) (*Snapshot, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	snap, exists := m.snapshotsByID[id]
	return snap, exists
}

// GetAll returns all snapshots in order.
func (m *Manager) GetAll() []*Snapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Snapshot, 0, len(m.ordered))
	for _, frameID := range m.ordered {
		if snap, exists := m.snapshots[frameID]; exists {
			result = append(result, snap)
		}
	}
	return result
}

// GetRange returns snapshots in the frame range [from, to].
func (m *Manager) GetRange(from, to devtools.FrameID) []*Snapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Snapshot, 0)
	for _, frameID := range m.ordered {
		if frameID >= from && frameID <= to {
			if snap, exists := m.snapshots[frameID]; exists {
				result = append(result, snap)
			}
		}
	}
	return result
}

// GetLatest returns the most recent snapshot.
func (m *Manager) GetLatest() (*Snapshot, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.ordered) == 0 {
		return nil, false
	}

	frameID := m.ordered[len(m.ordered)-1]
	snap, exists := m.snapshots[frameID]
	return snap, exists
}

// GetOldest returns the oldest snapshot.
func (m *Manager) GetOldest() (*Snapshot, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.ordered) == 0 {
		return nil, false
	}

	frameID := m.ordered[0]
	snap, exists := m.snapshots[frameID]
	return snap, exists
}

// =============================================================================
// Snapshot Management
// =============================================================================

// Delete removes a snapshot.
func (m *Manager) Delete(frameID devtools.FrameID) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.snapshots[frameID]; !exists {
		return false
	}

	// Remove from ordered slice
	newOrdered := make([]devtools.FrameID, 0, len(m.ordered)-1)
	for _, fid := range m.ordered {
		if fid != frameID {
			newOrdered = append(newOrdered, fid)
		}
	}
	m.ordered = newOrdered

	m.removeSnapshot(frameID)
	return true
}

// DeleteByID removes a snapshot by snapshot ID.
func (m *Manager) DeleteByID(id SnapshotID) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	snap, exists := m.snapshotsByID[id]
	if !exists {
		return false
	}

	return m.Delete(snap.FrameID)
}

// Clear removes all snapshots.
func (m *Manager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for frameID := range m.snapshots {
		m.removeSnapshot(frameID)
	}
	m.ordered = m.ordered[:0]
}

// Resize changes the maximum number of snapshots.
func (m *Manager) Resize(newMax int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.maxSnapshots = newMax

	// Trim excess snapshots
	for len(m.ordered) > newMax {
		oldestFrame := m.ordered[0]
		m.removeSnapshot(oldestFrame)
		m.ordered = m.ordered[1:]
	}
}

// =============================================================================
// Auto Capture
// =============================================================================

// SetAutoCapture enables or disables automatic snapshot capture.
func (m *Manager) SetAutoCapture(enabled bool, interval time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.autoCapture = enabled
	m.autoInterval = interval
}

// ShouldAutoCapture returns true if a snapshot should be auto-captured.
func (m *Manager) ShouldAutoCapture(frameID devtools.FrameID) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.autoCapture || m.autoInterval == 0 {
		return false
	}

	// Capture first frame
	if !m.hasAutoCapture {
		return true
	}

	// Check interval
	_, exists := m.snapshots[m.lastAutoFrame]
	if !exists {
		return true
	}

	elapsed := frameID - m.lastAutoFrame
	// For time-based interval, we'd need real timestamps
	// For now, use frame count as proxy
	intervalFrames := devtools.FrameID(m.autoInterval.Seconds() * 60) // Assume 60fps
	return elapsed >= intervalFrames || !exists
}

// NoteAutoCapture notes that a snapshot was captured (for auto-capture tracking).
func (m *Manager) NoteAutoCapture(frameID devtools.FrameID) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastAutoFrame = frameID
	m.hasAutoCapture = true
}

// =============================================================================
// Statistics
// =============================================================================

// Stats provides snapshot statistics.
type Stats struct {
	TotalSnapshots  int           `json:"total_snapshots"`
	MaxSnapshots    int           `json:"max_snapshots"`
	OldestFrame     devtools.FrameID `json:"oldest_frame,omitempty"`
	NewestFrame     devtools.FrameID `json:"newest_frame,omitempty"`
	PoolCreated     int           `json:"pool_created"`
	PoolReused      int           `json:"pool_reused"`
	AutoCapture     bool          `json:"auto_capture"`
	AutoInterval    time.Duration `json:"auto_interval_ms"`
}

// GetStats returns current statistics.
func (m *Manager) GetStats() Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := Stats{
		TotalSnapshots: len(m.ordered),
		MaxSnapshots:   m.maxSnapshots,
		AutoCapture:    m.autoCapture,
		AutoInterval:   m.autoInterval,
	}

	if len(m.ordered) > 0 {
		stats.OldestFrame = m.ordered[0]
		stats.NewestFrame = m.ordered[len(m.ordered)-1]
	}

	poolCreated, poolReused := m.pool.Stats()
	stats.PoolCreated = poolCreated
	stats.PoolReused = poolReused

	return stats
}

// =============================================================================
// Persistence
// =============================================================================

// SetPersistDir sets the directory for persistent snapshot storage.
func (m *Manager) SetPersistDir(dir string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create directory if it doesn't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create persist directory: %w", err)
	}

	m.persistDir = dir
	return nil
}

// Save saves a snapshot to persistent storage.
func (m *Manager) Save(frameID devtools.FrameID) error {
	if m.persistDir == "" {
		return fmt.Errorf("no persist directory set")
	}

	snap, exists := m.Get(frameID)
	if !exists {
		return fmt.Errorf("snapshot not found for frame %d", frameID)
	}

	data, err := snap.Serialize()
	if err != nil {
		return fmt.Errorf("failed to serialize snapshot: %w", err)
	}

	filename := filepath.Join(m.persistDir, fmt.Sprintf("snapshot_%d.json", frameID))
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write snapshot file: %w", err)
	}

	return nil
}

// Load loads a snapshot from persistent storage.
func (m *Manager) Load(frameID devtools.FrameID) error {
	if m.persistDir == "" {
		return fmt.Errorf("no persist directory set")
	}

	filename := filepath.Join(m.persistDir, fmt.Sprintf("snapshot_%d.json", frameID))
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read snapshot file: %w", err)
	}

	snap, err := DeserializeSnapshot(data)
	if err != nil {
		return fmt.Errorf("failed to deserialize snapshot: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.snapshots[frameID] = snap
	m.snapshotsByID[snap.ID] = snap

	// Add to ordered list if not present
	found := false
	for _, fid := range m.ordered {
		if fid == frameID {
			found = true
			break
		}
	}
	if !found {
		m.ordered = append(m.ordered, frameID)
	}

	return nil
}

// SaveAll saves all snapshots to persistent storage.
func (m *Manager) SaveAll() error {
	if m.persistDir == "" {
		return fmt.Errorf("no persist directory set")
	}

	m.mu.RLock()
	ordered := make([]devtools.FrameID, len(m.ordered))
	copy(ordered, m.ordered)
	m.mu.RUnlock()

	for _, frameID := range ordered {
		if err := m.Save(frameID); err != nil {
			return fmt.Errorf("failed to save snapshot %d: %w", frameID, err)
		}
	}

	return nil
}

// LoadAll loads all snapshots from persistent storage.
func (m *Manager) LoadAll() error {
	if m.persistDir == "" {
		return fmt.Errorf("no persist directory set")
	}

	files, err := os.ReadDir(m.persistDir)
	if err != nil {
		return fmt.Errorf("failed to read persist directory: %w", err)
	}

	for _, file := range files {
		var frameID devtools.FrameID
		if _, err := fmt.Sscanf(file.Name(), "snapshot_%d.json", &frameID); err == nil {
			if err := m.Load(frameID); err != nil {
				return fmt.Errorf("failed to load snapshot %d: %w", frameID, err)
			}
		}
	}

	return nil
}

// =============================================================================
// Export/Import
// =============================================================================

// ExportJSON exports all snapshots as JSON.
func (m *Manager) ExportJSON() ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	snapshots := m.GetAll()
	return json.MarshalIndent(snapshots, "", "  ")
}

// ImportJSON imports snapshots from JSON.
func (m *Manager) ImportJSON(data []byte) error {
	var snapshots []*Snapshot
	if err := json.Unmarshal(data, &snapshots); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, snap := range snapshots {
		m.snapshots[snap.FrameID] = snap
		m.snapshotsByID[snap.ID] = snap

		// Add to ordered list
		found := false
		for _, fid := range m.ordered {
			if fid == snap.FrameID {
				found = true
				break
			}
		}
		if !found {
			m.ordered = append(m.ordered, snap.FrameID)
		}
	}

	return nil
}
