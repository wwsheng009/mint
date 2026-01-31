// Package snapshot provides state snapshot and restoration for DevTools.
package snapshot

import (
	"fmt"
	"testing"
	"time"

	"github.com/wwsheng009/mint/devtools"
)

// TestSnapshotPool tests the snapshot pool.
func TestSnapshotPool(t *testing.T) {
	pool := NewSnapshotPool(5)

	// Acquire and release
	snap1 := pool.Acquire()
	if snap1 == nil {
		t.Fatal("expected non-nil snapshot")
	}

	pool.Release(snap1)

	// Should reuse
	snap2 := pool.Acquire()
	if snap2 != snap1 {
		t.Error("expected to reuse snapshot")
	}

	// Check stats
	created, reused := pool.Stats()
	if created != 1 {
		t.Errorf("expected 1 created, got %d", created)
	}
	if reused != 1 {
		t.Errorf("expected 1 reused, got %d", reused)
	}
}

// TestSnapshotBuilder tests the snapshot builder.
func TestSnapshotBuilder(t *testing.T) {
	builder := NewBuilder("test-1", devtools.FrameID(42))

	builder.SetWindowSize(80, 24)
	builder.SetTheme("dark")
	builder.AddTag("manual")
	builder.SetLabel("test", "value")

	snap := builder.Build()

	if snap.ID != "test-1" {
		t.Errorf("expected ID test-1, got %s", snap.ID)
	}

	if snap.FrameID != 42 {
		t.Errorf("expected frame ID 42, got %d", snap.FrameID)
	}

	if snap.Global.WindowSize.Width != 80 {
		t.Errorf("expected width 80, got %d", snap.Global.WindowSize.Width)
	}

	if len(snap.Metadata.Tags) != 1 {
		t.Errorf("expected 1 tag, got %d", len(snap.Metadata.Tags))
	}
}

// TestManager tests the snapshot manager.
func TestManager(t *testing.T) {
	mgr := NewManager(3)

	// Test capture
	builder1 := NewBuilder("snap-1", devtools.FrameID(1))
	snap1, err := mgr.Capture(1, builder1)
	if err != nil {
		t.Fatalf("failed to capture: %v", err)
	}

	builder2 := NewBuilder("snap-2", devtools.FrameID(2))
	_, _ = mgr.Capture(2, builder2)

	// Test get
	retrieved, ok := mgr.Get(1)
	if !ok {
		t.Fatal("failed to get snapshot")
	}

	if retrieved.ID != snap1.ID {
		t.Errorf("expected snap1, got %s", retrieved.ID)
	}

	// Test get all
	all := mgr.GetAll()
	if len(all) != 2 {
		t.Errorf("expected 2 snapshots, got %d", len(all))
	}

	// Test get range
	r := mgr.GetRange(1, 2)
	if len(r) != 2 {
		t.Errorf("expected 2 in range, got %d", len(r))
	}

	// Test max snapshots limit
	builder3 := NewBuilder("snap-3", devtools.FrameID(3))
	mgr.Capture(3, builder3)

	builder4 := NewBuilder("snap-4", devtools.FrameID(4))
	mgr.Capture(4, builder4)

	// Should have removed snap-1
	_, ok = mgr.Get(1)
	if ok {
		t.Error("expected snap-1 to be removed due to max limit")
	}

	// Test delete
	mgr.Delete(2)
	_, ok = mgr.Get(2)
	if ok {
		t.Error("expected snap-2 to be deleted")
	}

	// Test clear
	mgr.Clear()
	if len(mgr.GetAll()) != 0 {
		t.Error("expected no snapshots after clear")
	}
}

// TestManagerAutoCapture tests auto capture functionality.
func TestManagerAutoCapture(t *testing.T) {
	mgr := NewManager(100)

	// Enable auto capture with large interval (1 second = ~60 frames)
	mgr.SetAutoCapture(true, 1*time.Second)

	if !mgr.ShouldAutoCapture(0) {
		t.Error("expected first frame to be captured")
	}

	// Actually capture frame 0 and note it
	builder := NewBuilder("snap-0", devtools.FrameID(0))
	mgr.Capture(0, builder)
	mgr.NoteAutoCapture(0)

	// Should not capture frame 1 (elapsed=1, interval=60)
	if mgr.ShouldAutoCapture(1) {
		t.Error("expected not to capture frame 1 (interval 1s)")
	}
}

// TestDiffer tests the snapshot differer.
func TestDiffer(t *testing.T) {
	differ := NewDiffer()

	// Create two snapshots with differences
	snap1 := &Snapshot{
		ID:      "snap-1",
		FrameID: 1,
		States: make(map[devtools.NodeID]*ComponentState),
	}

	snap2 := &Snapshot{
		ID:      "snap-2",
		FrameID: 2,
		States: make(map[devtools.NodeID]*ComponentState),
	}

	// Add component to both
	nodeID := devtools.NodeID("button-1")
	snap1.States[nodeID] = &ComponentState{
		NodeID:   nodeID,
		Type:     "Button",
		Visible:  true,
		Bounds:   Rect{X: 0, Y: 0, Width: 10, Height: 1},
	}

	snap2.States[nodeID] = &ComponentState{
		NodeID:   nodeID,
		Type:     "Button",
		Visible:  false, // Changed
		Bounds:   Rect{X: 0, Y: 0, Width: 10, Height: 1},
	}

	// Add new component to snap2
	newNodeID := devtools.NodeID("button-2")
	snap2.States[newNodeID] = &ComponentState{
		NodeID:   newNodeID,
		Type:     "Button",
		Visible:  true,
		Bounds:   Rect{X: 0, Y: 2, Width: 10, Height: 1},
	}

	// Compare
	diff := differ.Compare(snap1, snap2)

	if !diff.HasChanges() {
		t.Error("expected changes")
	}

	// Should have 1 addition, 1 modification
	if diff.Summary.ComponentsAdded != 1 {
		t.Errorf("expected 1 added, got %d", diff.Summary.ComponentsAdded)
	}

	if diff.Summary.ComponentsModified != 1 {
		t.Errorf("expected 1 modified, got %d", diff.Summary.ComponentsModified)
	}

	// Format changes
	lines := diff.FormatChanges()
	if len(lines) != 2 {
		t.Errorf("expected 2 change lines, got %d", len(lines))
	}
}

// TestTimeTravelRange tests time travel range functionality.
func TestTimeTravelRange(t *testing.T) {
	// Create a series of snapshots
	snapshots := make([]*Snapshot, 5)
	for i := 0; i < 5; i++ {
		snapshots[i] = &Snapshot{
			ID:      SnapshotID(fmt.Sprintf("snap-%d", i)),
			FrameID: devtools.FrameID(i),
			States: make(map[devtools.NodeID]*ComponentState),
		}

		// Add a component that changes each frame
		nodeID := devtools.NodeID("counter")
		snapshots[i].States[nodeID] = &ComponentState{
			NodeID: nodeID,
			Type:   "Text",
			State:  map[string]interface{}{"value": i},
		}
	}

	// Create time travel range
	ttr := NewTimeTravelRange(snapshots)
	ttr.Compute()

	diffs := ttr.GetAllDiffs()
	if len(diffs) != 4 {
		t.Errorf("expected 4 diffs, got %d", len(diffs))
	}

	// Find changes for the counter node
	history := ttr.GetChangeHistory("counter")
	if len(history) != 4 {
		t.Errorf("expected 4 changes for counter, got %d", len(history))
	}
}

// TestSnapshotSerialization tests snapshot serialization.
func TestSnapshotSerialization(t *testing.T) {
	snap := &Snapshot{
		ID:      "test-serialize",
		FrameID: 123,
		Timestamp: time.Now().UTC(),
		States: map[devtools.NodeID]*ComponentState{
			"node-1": {
				NodeID:  "node-1",
				Type:    "Button",
				Visible: true,
				Bounds:  Rect{X: 0, Y: 0, Width: 10, Height: 1},
			},
		},
		Global: GlobalState{
			WindowSize: WindowSize{Width: 80, Height: 24},
		},
	}

	// Serialize
	data, err := snap.Serialize()
	if err != nil {
		t.Fatalf("failed to serialize: %v", err)
	}

	// Deserialize
	restored, err := DeserializeSnapshot(data)
	if err != nil {
		t.Fatalf("failed to deserialize: %v", err)
	}

	if restored.ID != snap.ID {
		t.Errorf("expected ID %s, got %s", snap.ID, restored.ID)
	}

	if restored.FrameID != snap.FrameID {
		t.Errorf("expected frame ID %d, got %d", snap.FrameID, restored.FrameID)
	}

	if len(restored.States) != len(snap.States) {
		t.Errorf("expected %d states, got %d", len(snap.States), len(restored.States))
	}
}

// TestManagerPersistence tests snapshot persistence.
func TestManagerPersistence(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	mgr := NewManager(10)
	mgr.SetPersistDir(tmpDir)

	// Capture a snapshot
	builder := NewBuilder("persist-1", devtools.FrameID(1))
	builder.SetWindowSize(80, 24)
	snap, _ := mgr.Capture(1, builder)

	// Save
	err := mgr.Save(1)
	if err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	// Clear and reload
	mgr.Clear()
	err = mgr.Load(1)
	if err != nil {
		t.Fatalf("failed to load: %v", err)
	}

	// Verify
	retrieved, ok := mgr.Get(1)
	if !ok {
		t.Fatal("failed to get loaded snapshot")
	}

	if retrieved.ID != snap.ID {
		t.Errorf("expected ID %s, got %s", snap.ID, retrieved.ID)
	}
}

// TestChangeTypeString tests ChangeType string conversion.
func TestChangeTypeString(t *testing.T) {
	tests := []struct {
		ct       ChangeType
		expected string
	}{
		{ChangeAdded, "added"},
		{ChangeRemoved, "removed"},
		{ChangeModified, "modified"},
		{ChangeMoved, "moved"},
		{ChangeType(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.ct.String() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.ct.String())
			}
		})
	}
}
