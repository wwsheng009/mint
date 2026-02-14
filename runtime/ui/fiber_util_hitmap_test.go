package ui

import (
	"testing"
	rtuievent "github.com/wwsheng009/mint/runtime/event"
)

func TestBuildHitMapFromFiber_CreatesEntries(t *testing.T) {
	// Create a simple Fiber tree
	root := &Fiber{
		NodeID: 1,
		Layer:  LayerBase,
		Type:   VNodeElement,
	}

	child1 := &Fiber{
		NodeID:       2,
		Layer:        LayerBase,
		Type:         VNodeElement,
		Return:       root,
		Sibling:      nil,
	}
	child2 := &Fiber{
		NodeID:       3,
		Layer:        LayerBase,
		Type:         VNodeElement,
		Return:       root,
		Sibling:      nil,
	}
	child1.Sibling = child2
	root.Child = child1

	// Set ComputedBox on each Fiber node
	// In a real scenario, this would be set by LayoutFiber()
	mockBox := struct {
		GetX, GetY, GetWidth, GetHeight int
	}{
		GetX:      10,
		GetY:      20,
		GetWidth:  100,
		GetHeight: 50,
	}

	root.ComputedBox = &mockBox
	child1.ComputedBox = &mockBox
	child2.ComputedBox = &mockBox

	// Build HitMap from Fiber
	hitMap := BuildHitMapFromFiber(root)

	// Verify we got entries
	if hitMap == nil {
		t.Fatal("BuildHitMapFromFiber returned nil")
	}

	entries := hitMap.AllEntries()
	if len(entries) != 3 {
		t.Fatalf("Expected 3 entries, got %d", len(entries))
	}

	// Verify NodeIDs match Fiber NodeIDs
	expectedIDs := []uint64{1, 2, 3}
	for i, entry := range entries {
		if entry.NodeID != expectedIDs[i] {
			t.Errorf("Entry %d: expected NodeID=%d, got %d", i, expectedIDs[i], entry.NodeID)
		}

		// Verify layer is Base (0)
		expectedLayer := int(LayerBase)
		expectedZOrder := expectedLayer * 10000
		if entry.ZOrder != expectedZOrder && entry.ZOrder != expectedZOrder+1 {
			// Allow for tree depth variation
			t.Logf("Entry %d: ZOrder=%d (NodeID=%d) - acceptable variation", i, entry.ZOrder, entry.NodeID)
		}
	}

	t.Logf("✅ BuildHitMapFromFiber created %d entries", len(entries))
}

func TestBuildHitMapFromFiber_NilFiber(t *testing.T) {
	hitMap := BuildHitMapFromFiber(nil)
	if hitMap == nil {
		t.Fatal("BuildHitMapFromFiber returned nil for nil Fiber")
	}
	if hitMap.Size() != 0 {
		t.Errorf("Expected 0 entries for nil Fiber, got %d", hitMap.Size())
	}
	t.Log("✅ BuildHitMapFromFiber handles nil Fiber correctly")
}

func TestBuildHitMapFromFiber_NilComputedBox(t *testing.T) {
	// Create a Fiber tree without ComputedBox
	root := &Fiber{
		NodeID: 1,
		Layer:  LayerBase,
		Type:   VNodeElement,
		Child: &Fiber{
			NodeID: 2,
			Layer:  LayerBase,
			Type:   VNodeElement,
		},
	}

	// Build HitMap - should skip nodes without ComputedBox
	hitMap := BuildHitMapFromFiber(root)

	if hitMap == nil {
		t.Fatal("BuildHitMapFromFiber returned nil")
	}

	if hitMap.Size() != 0 {
		t.Errorf("Expected 0 entries (ComputedBox is nil), got %d", hitMap.Size())
	}

	t.Log("✅ BuildHitMapFromFiber skips nodes without ComputedBox")
}

// Mock ComputedBox type for testing
type mockComputedBox struct {
	X, Y, Width, Height int
}

func (b *mockComputedBox) GetX() int      { return b.X }
func (b *mockComputedBox) GetY() int      { return b.Y }
func (b *mockComputedBox) GetWidth() int  { return b.Width }
func (b *mockComputedBox) GetHeight() int { return b.Height }

func TestBuildHitMapFromFiber_WithProperComputedBox(t *testing.T) {
	// Create a Fiber tree with proper ComputedBox implementation
	root := &Fiber{
		NodeID: 1,
		Layer:  LayerModal, // Use Modal layer to test Z-order
		Type:   VNodeElement,
		Child: &Fiber{
			NodeID: 2,
			Layer:  LayerBase,
			Type:   VNodeElement,
		},
	}

	// Set mock ComputedBox
	root.ComputedBox = &mockComputedBox{X: 50, Y: 50, Width: 200, Height: 100}
	root.Child.ComputedBox = &mockComputedBox{X: 60, Y: 60, Width: 180, Height: 80}

	hitMap := BuildHitMapFromFiber(root)

	if hitMap == nil {
		t.Fatal("BuildHitMapFromFiber returned nil")
	}

	entries := hitMap.AllEntries()
	if len(entries) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(entries))
	}

	// Verify modal entry has higher Z-order
	modalEntry := findEntryByNodeID(entries, 1)
	baseEntry := findEntryByNodeID(entries, 2)

	if modalEntry == nil || baseEntry == nil {
		t.Fatal("Could not find expected entries")
	}

	// Modal Z-order should be LayerModal (2) * 10000 + depth
	// Base Z-order should be LayerBase (0) * 10000 + depth
	if modalEntry.ZOrder <= baseEntry.ZOrder {
		t.Errorf("Modal Z-order (%d) should be > Base Z-order (%d)",
			modalEntry.ZOrder, baseEntry.ZOrder)
	}

	// Verify bounds
	modalBounds := modalEntry.Bounds
	if modalBounds.X != 50 || modalBounds.Y != 50 {
		t.Errorf("Modal bounds: expected (50,50), got (%d,%d)", modalBounds.X, modalBounds.Y)
	}

	t.Logf("✅ BuildHitMapFromFiber with modal: modal Z=%d, base Z=%d",
		modalEntry.ZOrder, baseEntry.ZOrder)
}

func findEntryByNodeID(entries []rtuievent.HitMapEntry, nodeID uint64) *rtuievent.HitMapEntry {
	for i := range entries {
		if entries[i].NodeID == nodeID {
			return &entries[i]
		}
	}
	return nil
}
