package devtools

import (
	"sync/atomic"
	"testing"
)

func TestNodeID(t *testing.T) {
	id := NodeID("test_node")
	if id != "test_node" {
		t.Errorf("NodeID = %v, want %v", id, "test_node")
	}
}

func TestMutationID(t *testing.T) {
	id := MutationID(123)
	if id != 123 {
		t.Errorf("MutationID = %v, want %v", id, 123)
	}
}

func TestFrameID(t *testing.T) {
	id := FrameID(456)
	if id != 456 {
		t.Errorf("FrameID = %v, want %v", id, 456)
	}
}

func TestChangeMask(t *testing.T) {
	mask := ChangeRect | ChangeZ
	if mask&ChangeRect == 0 {
		t.Error("ChangeRect not set")
	}
	if mask&ChangeZ == 0 {
		t.Error("ChangeZ not set")
	}
	if mask&ChangeVisibility != 0 {
		t.Error("ChangeVisibility should not be set")
	}
}

func TestRect(t *testing.T) {
	r := Rect{X: 10, Y: 20, Width: 100, Height: 200}
	if r.X != 10 || r.Y != 20 || r.Width != 100 || r.Height != 200 {
		t.Errorf("Rect = %+v, want {X:10, Y:20, Width:100, Height:200}", r)
	}
}

func TestNodeDelta(t *testing.T) {
	delta := NodeDelta{
		ID:   NodeID("test"),
		Mask: ChangeRect,
		Rect: &Rect{X: 1, Y: 2, Width: 3, Height: 4},
	}

	if delta.ID != "test" {
		t.Errorf("ID = %v, want %v", delta.ID, "test")
	}
	if delta.Rect == nil {
		t.Error("Rect is nil")
	}
	if delta.Mask != ChangeRect {
		t.Errorf("Mask = %v, want %v", delta.Mask, ChangeRect)
	}
}

func TestLayoutDelta(t *testing.T) {
	delta := LayoutDelta{
		FrameID: FrameID(10),
		Added:   []NodeID{"node1", "node2"},
		Removed: []NodeID{"node3"},
		Changed: []NodeDelta{},
	}

	if delta.FrameID != 10 {
		t.Errorf("FrameID = %v, want %v", delta.FrameID, 10)
	}
	if len(delta.Added) != 2 {
		t.Errorf("Added length = %v, want %v", len(delta.Added), 2)
	}
	if len(delta.Removed) != 1 {
		t.Errorf("Removed length = %v, want %v", len(delta.Removed), 1)
	}
}

func TestEventEntry(t *testing.T) {
	entry := EventEntry{
		Type:   "click",
		Target: NodeID("button1"),
		Phase:  "bubble",
		Data:   map[string]interface{}{"x": 100, "y": 200},
	}

	if entry.Type != "click" {
		t.Errorf("Type = %v, want %v", entry.Type, "click")
	}
	if entry.Target != "button1" {
		t.Errorf("Target = %v, want %v", entry.Target, "button1")
	}
	if entry.Data["x"] != 100 {
		t.Errorf("Data[x] = %v, want %v", entry.Data["x"], 100)
	}
}

func TestConfig(t *testing.T) {
	config := DefaultConfig()

	if config.BufferSize != 4096 {
		t.Errorf("BufferSize = %v, want %v", config.BufferSize, 4096)
	}
	if !config.EnableOverlay {
		t.Error("EnableOverlay should be true")
	}
	if !config.EnableMutationTap {
		t.Error("EnableMutationTap should be true")
	}
}

func TestDevToolsEnableDisable(t *testing.T) {
	dt := New()

	if dt.IsEnabled() {
		t.Error("Should be disabled initially")
	}

	dt.Enable()

	if !dt.IsEnabled() {
		t.Error("Should be enabled after Enable()")
	}

	dt.Disable()

	if dt.IsEnabled() {
		t.Error("Should be disabled after Disable()")
	}
}

func TestDevToolsRecordEvent(t *testing.T) {
	dt := New()
	dt.Enable()

	dt.RecordEvent("click", "button1", "bubble", nil)

	// Just ensure it doesn't panic
	dt.EndFrame()
}

func TestDebugOverlay(t *testing.T) {
	overlay := NewDebugOverlay(80, 24)

	overlay.Highlight("node1", Rect{X: 10, Y: 10, Width: 20, Height: 20})

	if !overlay.IsShown("node1") {
		t.Error("node1 should be shown")
	}

	overlay.Clear()

	if overlay.IsShown("node1") {
		t.Error("node1 should not be shown after Clear")
	}
}

func TestEventBus(t *testing.T) {
	bus := NewEventBus(16)

	if bus.IsEnabled() {
		t.Error("Should be disabled initially")
	}

	bus.Enable()

	if !bus.IsEnabled() {
		t.Error("Should be enabled after Enable()")
	}

	// Emit some events
	for i := 0; i < 10; i++ {
		bus.Emit(DebugEvent{
			Type: EventLayout,
			Frame: i,
		})
	}

	if bus.WritePosition() != 10 {
		t.Errorf("WritePosition = %v, want %v", bus.WritePosition(), 10)
	}

	bus.Close()
}

func TestMutationTap(t *testing.T) {
	// Reset to clean state
	ResetMutationTap()

	if IsMutationTapEnabled() {
		t.Error("Should be disabled initially")
	}

	EnableMutationTap()

	if !IsMutationTapEnabled() {
		t.Error("Should be enabled after EnableMutationTap()")
	}

	// Record some mutations
	for i := uint32(0); i < 10; i++ {
		RecordMutation(i, uint16(i), uint8(i%4), uint64(i), uint64(i+1))
	}

	// Poll mutations
	var pos uint32 = 0
	mutations := PollMutations(&pos)

	if len(mutations) != 10 {
		t.Errorf("Got %d mutations, want 10", len(mutations))
	}

	DisableMutationTap()
}

func TestAtomicOperations(t *testing.T) {
	var enabled uint32

	if atomic.LoadUint32(&enabled) != 0 {
		t.Error("Should be 0 initially")
	}

	atomic.StoreUint32(&enabled, 1)

	if atomic.LoadUint32(&enabled) != 1 {
		t.Error("Should be 1 after Store")
	}

	if !atomic.CompareAndSwapUint32(&enabled, 1, 0) {
		t.Error("CompareAndSwap should succeed")
	}

	if atomic.LoadUint32(&enabled) != 0 {
		t.Error("Should be 0 after CompareAndSwap")
	}
}

func TestNextPowerOfTwo(t *testing.T) {
	bus := NewEventBus(100) // Should round up to 128

	// The internal buffer size should be a power of 2
	if bus.WritePosition() != 0 {
		t.Error("WritePosition should be 0 initially")
	}
}
