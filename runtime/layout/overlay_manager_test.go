package layout

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewOverlayManager(t *testing.T) {
	om := NewOverlayManager()
	assert.NotNil(t, om)
	assert.Equal(t, 0, om.Size())
}

func TestOverlayManager_PushAndTop(t *testing.T) {
	om := NewOverlayManager()

	box1 := &LayoutBox{ID: "box1"}
	box2 := &LayoutBox{ID: "box2"}

	// Push first entry
	om.Push("portal1", box1, "root1", 10)

	assert.Equal(t, 1, om.Size())
	top := om.Top()
	assert.NotNil(t, top)
	assert.Equal(t, "portal1", top.ID)
	assert.Equal(t, box1, top.Box)
	assert.Equal(t, "root1", top.PortalRootID)
	assert.Equal(t, 10, top.Priority)

	// Push second entry with higher priority
	om.Push("portal2", box2, "root2", 20)

	assert.Equal(t, 2, om.Size())
	top = om.Top()
	assert.NotNil(t, top)
	assert.Equal(t, "portal2", top.ID)  // Higher priority should be on top
}

func TestOverlayManager_Pop(t *testing.T) {
	om := NewOverlayManager()

	box1 := &LayoutBox{ID: "box1"}
	box2 := &LayoutBox{ID: "box2"}

	om.Push("portal1", box1, "root1", 10)
	om.Push("portal2", box2, "root2", 20)

	// Pop top entry (portal2)
	popped := om.Pop()
	assert.NotNil(t, popped)
	assert.Equal(t, "portal2", popped.ID)
	assert.Equal(t, 1, om.Size())

	// Top should now be portal1
	top := om.Top()
	assert.NotNil(t, top)
	assert.Equal(t, "portal1", top.ID)

	// Pop remaining
	popped = om.Pop()
	assert.NotNil(t, popped)
	assert.Equal(t, "portal1", popped.ID)
	assert.Equal(t, 0, om.Size())

	// Pop on empty stack
	popped = om.Pop()
	assert.Nil(t, popped)
}

func TestOverlayManager_GetAll(t *testing.T) {
	om := NewOverlayManager()

	box1 := &LayoutBox{ID: "box1"}
	box2 := &LayoutBox{ID: "box2"}
	box3 := &LayoutBox{ID: "box3"}

	om.Push("portal1", box1, "root1", 10)
	om.Push("portal2", box2, "root2", 30)
	om.Push("portal3", box3, "root3", 20)

	// GetAll should return all active entries
	entries := om.GetAll()
	assert.Equal(t, 3, len(entries))

	// Should be sorted by priority (ascending: 10, 20, 30)
	assert.Equal(t, "portal1", entries[0].ID)  // priority 10
	assert.Equal(t, "portal3", entries[1].ID)  // priority 20
	assert.Equal(t, "portal2", entries[2].ID)  // priority 30
}

func TestOverlayManager_GetByID(t *testing.T) {
	om := NewOverlayManager()

	box1 := &LayoutBox{ID: "box1"}
	om.Push("portal1", box1, "root1", 10)

	// Get existing
	entry := om.GetByID("portal1")
	assert.NotNil(t, entry)
	assert.Equal(t, "portal1", entry.ID)
	assert.Equal(t, box1, entry.Box)

	// Get non-existent
	entry = om.GetByID("nonexistent")
	assert.Nil(t, entry)
}

func TestOverlayManager_Remove(t *testing.T) {
	om := NewOverlayManager()

	box1 := &LayoutBox{ID: "box1"}
	box2 := &LayoutBox{ID: "box2"}

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

	// Remove second
	om.Remove("portal2")
	assert.Equal(t, 0, om.Size())

	top = om.Top()
	assert.Nil(t, top)
}

func TestOverlayManager_Clear(t *testing.T) {
	om := NewOverlayManager()

	om.Push("portal1", &LayoutBox{ID: "box1"}, "root1", 10)
	om.Push("portal2", &LayoutBox{ID: "box2"}, "root2", 20)
	om.Push("portal3", &LayoutBox{ID: "box3"}, "root3", 30)

	assert.Equal(t, 3, om.Size())

	om.Clear()

	assert.Equal(t, 0, om.Size())
	assert.Nil(t, om.Top())

	// GetByID should return nil
	assert.Nil(t, om.GetByID("portal1"))
	assert.Nil(t, om.GetByID("portal2"))
	assert.Nil(t, om.GetByID("portal3"))
}

func TestOverlayManager_PushSameID(t *testing.T) {
	om := NewOverlayManager()

	box1 := &LayoutBox{ID: "box1"}
	box2 := &LayoutBox{ID: "box2"}

	// Push with same ID twice
	om.Push("portal1", box1, "root1", 10)
	om.Push("portal1", box2, "root2", 20)

	// Should only have one entry (new one replaced old)
	assert.Equal(t, 1, om.Size())

	top := om.Top()
	assert.NotNil(t, top)
	assert.Equal(t, "portal1", top.ID)  // ID should be the same as pushed
	assert.Equal(t, "root2", top.PortalRootID)  // PortalRootID should be the second one
	assert.Equal(t, box2, top.Box)
	assert.Equal(t, 20, top.Priority)
}

func TestOverlayManager_PriorityOrder(t *testing.T) {
	om := NewOverlayManager()

	// Push in random priority order
	om.Push("p3", &LayoutBox{ID: "box3"}, "root3", 30)
	om.Push("p1", &LayoutBox{ID: "box1"}, "root1", 10)
	om.Push("p2", &LayoutBox{ID: "box2"}, "root2", 20)
	om.Push("p4", &LayoutBox{ID: "box4"}, "root4", 40)

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

func TestOverlayManager_ConcurrentAccess(t *testing.T) {
	om := NewOverlayManager()
	done := make(chan bool)

	// Concurrent pushes
	for i := 0; i < 10; i++ {
		go func(n int) {
			om.Push("p"+string(rune(n)), &LayoutBox{ID: "box"}, "root", n)
			done <- true
		}(i)
	}

	// Concurrent reads
	results := make(chan int)
	for i := 0; i < 5; i++ {
		go func() {
			size := om.Size()
			results <- size
		}()
	}

	// Wait for all pushes
	for i := 0; i < 10; i++ {
		<-done
	}

	// Wait for all reads
	for i := 0; i < 5; i++ {
		<-results
	}

	// Should have all entries without deadlock
	assert.Equal(t, 10, om.Size())
}
