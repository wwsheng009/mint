// Portal System Test (Phase 3)
// Tests for OverlayManager and Portal functionality
package portal_demo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wwsheng009/mint/runtime/layout"
)

// TestOverlayManager_Basic tests the overlay manager stack operations
func TestOverlayManager_Basic(t *testing.T) {
	om := layout.NewOverlayManager()

	// Test empty state
	assert.Equal(t, 0, om.Size())
	assert.Nil(t, om.Top())

	// Push a portal entry
	box1 := &layout.LayoutBox{ID: "entry1"}
	om.Push("portal1", box1, "root1", 10)

	assert.Equal(t, 1, om.Size())
	top := om.Top()
	assert.NotNil(t, top)
	assert.Equal(t, "portal1", top.ID)
	assert.Equal(t, 10, top.Priority)

	// Push another entry with higher priority
	box2 := &layout.LayoutBox{ID: "entry2"}
	om.Push("portal2", box2, "root2", 20)

	assert.Equal(t, 2, om.Size())
	top = om.Top()
	assert.NotNil(t, top)
	assert.Equal(t, "portal2", top.ID)

	// Pop the top entry
	popped := om.Pop()
	assert.NotNil(t, popped)
	assert.Equal(t, "portal2", popped.ID)
	assert.Equal(t, 1, om.Size())

	// Pop remaining
	popped = om.Pop()
	assert.NotNil(t, popped)
	assert.Equal(t, "portal1", popped.ID)
	assert.Equal(t, 0, om.Size())
	assert.Nil(t, om.Top())
}

// TestOverlayManager_PriorityOrdering tests that entries are ordered by priority
func TestOverlayManager_PriorityOrdering(t *testing.T) {
	om := layout.NewOverlayManager()

	// Push entries in random priority order
	om.Push("p3", &layout.LayoutBox{ID: "box3"}, "root3", 30)
	om.Push("p1", &layout.LayoutBox{ID: "box1"}, "root1", 10)
	om.Push("p2", &layout.LayoutBox{ID: "box2"}, "root2", 20)
	om.Push("p4", &layout.LayoutBox{ID: "box4"}, "root4", 40)

	// GetAll should return in ascending priority order
	entries := om.GetAll()
	assert.Equal(t, 4, len(entries))
	assert.Equal(t, "p1", entries[0].ID)  // priority 10
	assert.Equal(t, "p2", entries[1].ID)  // priority 20
	assert.Equal(t, "p3", entries[2].ID)  // priority 30
	assert.Equal(t, "p4", entries[3].ID)  // priority 40

	// Top should be highest priority
	top := om.Top()
	assert.Equal(t, "p4", top.ID)
}

// TestOverlayManager_GetByID tests retrieving entries by ID
func TestOverlayManager_GetByID(t *testing.T) {
	om := layout.NewOverlayManager()

	box1 := &layout.LayoutBox{ID: "box1"}
	box2 := &layout.LayoutBox{ID: "box2"}

	om.Push("portal1", box1, "root1", 10)
	om.Push("portal2", box2, "root2", 20)

	// Get existing
	entry := om.GetByID("portal1")
	assert.NotNil(t, entry)
	assert.Equal(t, "portal1", entry.ID)
	assert.Equal(t, box1, entry.Box)

	// Get non-existent
	entry = om.GetByID("nonexistent")
	assert.Nil(t, entry)
}

// TestOverlayManager_Remove tests removing entries
func TestOverlayManager_Remove(t *testing.T) {
	om := layout.NewOverlayManager()

	box1 := &layout.LayoutBox{ID: "box1"}
	box2 := &layout.LayoutBox{ID: "box2"}

	om.Push("portal1", box1, "root1", 10)
	om.Push("portal2", box2, "root2", 20)

	assert.Equal(t, 2, om.Size())

	// Remove first
	om.Remove("portal1")
	assert.Equal(t, 1, om.Size())

	// Check that portal1 is removed
	entry := om.GetByID("portal1")
	assert.Nil(t, entry)

	// Top should still be portal2
	top := om.Top()
	assert.Equal(t, "portal2", top.ID)
}

// TestOverlayManager_Clear tests clearing all entries
func TestOverlayManager_Clear(t *testing.T) {
	om := layout.NewOverlayManager()

	om.Push("portal1", &layout.LayoutBox{ID: "box1"}, "root1", 10)
	om.Push("portal2", &layout.LayoutBox{ID: "box2"}, "root2", 20)
	om.Push("portal3", &layout.LayoutBox{ID: "box3"}, "root3", 30)

	assert.Equal(t, 3, om.Size())

	om.Clear()

	assert.Equal(t, 0, om.Size())
	assert.Nil(t, om.Top())

	// GetByID should return nil for all
	assert.Nil(t, om.GetByID("portal1"))
	assert.Nil(t, om.GetByID("portal2"))
	assert.Nil(t, om.GetByID("portal3"))
}

// TestOverlayManager_PushSameID tests pushing with same ID
func TestOverlayManager_PushSameID(t *testing.T) {
	om := layout.NewOverlayManager()

	box1 := &layout.LayoutBox{ID: "box1"}
	box2 := &layout.LayoutBox{ID: "box2"}

	// Push with same ID twice
	om.Push("portal1", box1, "root1", 10)
	om.Push("portal1", box2, "root2", 20)

	// Should only have one entry (new one replaced old)
	assert.Equal(t, 1, om.Size())

	top := om.Top()
	assert.NotNil(t, top)
	assert.Equal(t, "portal1", top.ID)
	assert.Equal(t, "root2", top.PortalRootID)
	assert.Equal(t, box2, top.Box)
	assert.Equal(t, 20, top.Priority)
}

// TestEngine_OverlayManager tests that Engine has OverlayManager
func TestEngine_OverlayManager(t *testing.T) {
	engine := layout.NewEngine()

	om := engine.GetOverlayManager()
	assert.NotNil(t, om)
	assert.Equal(t, 0, om.Size())

	// Verify it's the same instance
	om.Push("test", &layout.LayoutBox{ID: "test"}, "root", 10)
	assert.Equal(t, 1, om.Size())

	om2 := engine.GetOverlayManager()
	assert.Equal(t, 1, om2.Size())  // Same instance
}

// TestOverlayEntry_Fields tests OverlayEntry structure
func TestOverlayEntry_Fields(t *testing.T) {
	om := layout.NewOverlayManager()

	box := &layout.LayoutBox{ID: "test-box", Width: 40, Height: 10}
	om.Push("test-portal", box, "test-root", 50)

	entry := om.GetByID("test-portal")
	assert.NotNil(t, entry)

	assert.Equal(t, "test-portal", entry.ID)
	assert.Equal(t, "test-root", entry.PortalRootID)
	assert.Equal(t, 50, entry.Priority)
	assert.True(t, entry.Active)
	assert.Equal(t, box, entry.Box)
}
