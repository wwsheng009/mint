package scheduler

import (
	"testing"
	"time"

	"github.com/wwsheng009/mint/runtime/priority"
)

// mockNode is a simple mock node for testing.
type mockNode struct {
	id string
}

func (m *mockNode) ID() string { return m.id }

// mockRenderer implements Renderer for testing.
type mockRenderer struct {
	layoutCalled map[string]bool
	paintCalled  map[string]bool
}

func newMockRenderer() *mockRenderer {
	return &mockRenderer{
		layoutCalled: make(map[string]bool),
		paintCalled:  make(map[string]bool),
	}
}

func (m *mockRenderer) Layout(node interface{}) {
	if n, ok := node.(*mockNode); ok {
		m.layoutCalled[n.id] = true
	}
}

func (m *mockRenderer) Paint(node interface{}) {
	if n, ok := node.(*mockNode); ok {
		m.paintCalled[n.id] = true
	}
}

func TestScheduler_MarkDirty(t *testing.T) {
	s := New()
	node := &mockNode{id: "test1"}

	s.MarkDirty("test1", node, priority.DirtyNormal)

	if !s.IsDirty("test1") {
		t.Error("expected node to be marked dirty")
	}

	if count := s.TotalDirtyCount(); count != 1 {
		t.Errorf("expected 1 dirty node, got %d", count)
	}
}

func TestScheduler_ClearDirty(t *testing.T) {
	s := New()
	node := &mockNode{id: "test1"}

	s.MarkDirty("test1", node, priority.DirtyNormal)
	s.ClearDirty("test1")

	if s.IsDirty("test1") {
		t.Error("expected node to be cleared")
	}

	if count := s.TotalDirtyCount(); count != 0 {
		t.Errorf("expected 0 dirty nodes, got %d", count)
	}
}

func TestScheduler_DirtyCount(t *testing.T) {
	s := New()

	s.MarkDirty("high", &mockNode{id: "high"}, priority.DirtyHigh)
	s.MarkDirty("normal1", &mockNode{id: "normal1"}, priority.DirtyNormal)
	s.MarkDirty("normal2", &mockNode{id: "normal2"}, priority.DirtyNormal)
	s.MarkDirty("low", &mockNode{id: "low"}, priority.DirtyLow)

	counts := s.DirtyCount()

	if counts[priority.DirtyHigh] != 1 {
		t.Errorf("expected 1 high priority, got %d", counts[priority.DirtyHigh])
	}
	if counts[priority.DirtyNormal] != 2 {
		t.Errorf("expected 2 normal priority, got %d", counts[priority.DirtyNormal])
	}
	if counts[priority.DirtyLow] != 1 {
		t.Errorf("expected 1 low priority, got %d", counts[priority.DirtyLow])
	}
}

func TestScheduler_BeginEndBatch(t *testing.T) {
	s := New()

	s.BeginBatch()
	if !s.IsBatching() {
		t.Error("expected batching to be active")
	}

	s.MarkDirty("test1", &mockNode{id: "test1"}, priority.DirtyNormal)
	s.MarkDirty("test2", &mockNode{id: "test2"}, priority.DirtyNormal)

	// While batching, dirty count should be 0 (nodes are in batch)
	if count := s.TotalDirtyCount(); count != 0 {
		t.Errorf("expected 0 dirty nodes while batching, got %d", count)
	}
	if size := s.GetBatchSize(); size != 2 {
		t.Errorf("expected batch size 2, got %d", size)
	}

	s.EndBatch(true)

	if s.IsBatching() {
		t.Error("expected batching to be inactive")
	}
	if count := s.TotalDirtyCount(); count != 2 {
		t.Errorf("expected 2 dirty nodes after flush, got %d", count)
	}
}

func TestScheduler_ProcessNext(t *testing.T) {
	s := New()
	renderer := newMockRenderer()

	s.MarkDirty("high", &mockNode{id: "high"}, priority.DirtyHigh)
	s.MarkDirty("normal1", &mockNode{id: "normal1"}, priority.DirtyNormal)
	s.MarkDirty("normal2", &mockNode{id: "normal2"}, priority.DirtyNormal)
	s.MarkDirty("low", &mockNode{id: "low"}, priority.DirtyLow)

	result := s.ProcessNext(renderer, DefaultProcessOptions())

	if result.Processed != 4 {
		t.Errorf("expected 4 processed nodes, got %d", result.Processed)
	}
	if result.Remaining != 0 {
		t.Errorf("expected 0 remaining, got %d", result.Remaining)
	}
	if !renderer.layoutCalled["high"] || !renderer.paintCalled["high"] {
		t.Error("expected high priority node to be processed")
	}
}

func TestScheduler_ProcessNextWithLimit(t *testing.T) {
	s := New()
	renderer := newMockRenderer()

	s.MarkDirty("n1", &mockNode{id: "n1"}, priority.DirtyNormal)
	s.MarkDirty("n2", &mockNode{id: "n2"}, priority.DirtyNormal)
	s.MarkDirty("n3", &mockNode{id: "n3"}, priority.DirtyNormal)

	opts := ProcessOptions{MaxNodes: 2}
	result := s.ProcessNext(renderer, opts)

	if result.Processed != 2 {
		t.Errorf("expected 2 processed nodes, got %d", result.Processed)
	}
	if result.Remaining != 1 {
		t.Errorf("expected 1 remaining, got %d", result.Remaining)
	}
}

func TestScheduler_ProcessNextWithTimeBudget(t *testing.T) {
	s := NewWithBudget(1 * time.Microsecond) // Very short time budget
	renderer := newMockRenderer()

	// Add many nodes
	for i := 0; i < 1000; i++ {
		id := "node" + string(rune('0'+i%10))
		s.MarkDirty(id, &mockNode{id: id}, priority.DirtyNormal)
	}

	opts := ProcessOptions{TimeBudget: 1 * time.Microsecond}
	result := s.ProcessNext(renderer, opts)

	if result.Processed == 0 {
		t.Error("expected at least one node to be processed")
	}
	// With a very short time budget and many nodes, some should remain
	if result.Remaining == 0 {
		t.Logf("Warning: all nodes processed, time budget may not have been limiting. Processed: %d", result.Processed)
	}
}

func TestScheduler_ShouldFlush(t *testing.T) {
	s := NewWithConfig(2*time.Millisecond, 10*time.Millisecond, 3)

	s.BeginBatch()

	// Batch size below limit
	s.MarkDirty("n1", &mockNode{id: "n1"}, priority.DirtyNormal)
	s.MarkDirty("n2", &mockNode{id: "n2"}, priority.DirtyNormal)
	if s.ShouldFlush() {
		t.Error("should not flush yet (batch not full)")
	}

	// Batch size at limit
	s.MarkDirty("n3", &mockNode{id: "n3"}, priority.DirtyNormal)
	if !s.ShouldFlush() {
		t.Error("should flush (batch full)")
	}
}

func TestScheduler_Clear(t *testing.T) {
	s := New()

	s.MarkDirty("test1", &mockNode{id: "test1"}, priority.DirtyNormal)
	s.MarkDirty("test2", &mockNode{id: "test2"}, priority.DirtyNormal)

	s.BeginBatch()
	s.MarkDirty("test3", &mockNode{id: "test3"}, priority.DirtyNormal)

	s.Clear()

	if s.IsBatching() {
		t.Error("expected batching to be inactive after clear")
	}
	if count := s.TotalDirtyCount(); count != 0 {
		t.Errorf("expected 0 dirty nodes after clear, got %d", count)
	}
	if size := s.GetBatchSize(); size != 0 {
		t.Errorf("expected batch size 0 after clear, got %d", size)
	}
}

func TestScheduler_MarkLayoutDirty(t *testing.T) {
	s := New()
	renderer := newMockRenderer()

	node := &mockNode{id: "test"}
	s.MarkLayoutDirty("test", node, priority.DirtyNormal)

	result := s.ProcessNext(renderer, DefaultProcessOptions())

	if result.Processed != 1 {
		t.Errorf("expected 1 processed node, got %d", result.Processed)
	}
	if !renderer.layoutCalled["test"] {
		t.Error("expected layout to be called")
	}
	if renderer.paintCalled["test"] {
		t.Error("expected paint not to be called")
	}
}

func TestScheduler_MarkPaintDirty(t *testing.T) {
	s := New()
	renderer := newMockRenderer()

	node := &mockNode{id: "test"}
	s.MarkPaintDirty("test", node, priority.DirtyNormal)

	result := s.ProcessNext(renderer, DefaultProcessOptions())

	if result.Processed != 1 {
		t.Errorf("expected 1 processed node, got %d", result.Processed)
	}
	if renderer.layoutCalled["test"] {
		t.Error("expected layout not to be called")
	}
	if !renderer.paintCalled["test"] {
		t.Error("expected paint to be called")
	}
}

func TestScheduler_PriorityOrder(t *testing.T) {
	s := New()
	renderer := newMockRenderer()

	// Add nodes in random order
	s.MarkDirty("low", &mockNode{id: "low"}, priority.DirtyLow)
	s.MarkDirty("high", &mockNode{id: "high"}, priority.DirtyHigh)
	s.MarkDirty("normal", &mockNode{id: "normal"}, priority.DirtyNormal)

	// Process only normal priority
	opts := ProcessOptions{
		PriorityLevels: []priority.DirtyLevel{priority.DirtyNormal},
	}
	result := s.ProcessNext(renderer, opts)

	if result.Processed != 1 {
		t.Errorf("expected 1 processed node, got %d", result.Processed)
	}
	if result.Remaining != 2 {
		t.Errorf("expected 2 remaining, got %d", result.Remaining)
	}
	if !renderer.layoutCalled["normal"] {
		t.Error("expected normal to be processed")
	}
	if renderer.layoutCalled["high"] || renderer.layoutCalled["low"] {
		t.Error("expected only normal priority to be processed")
	}
}
