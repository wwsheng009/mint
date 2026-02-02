package ui

import (
	"testing"
	"time"

	"github.com/wwsheng009/mint/runtime/priority"
)

// TestNewUIScheduler tests scheduler creation
func TestNewUIScheduler(t *testing.T) {
	s := NewUIScheduler()

	if s == nil {
		t.Fatal("NewUIScheduler() returned nil")
	}

	if s.scheduler == nil {
		t.Error("scheduler field is nil")
	}
}

// TestNewUISchedulerWithBudget tests scheduler with custom budget
func TestNewUISchedulerWithBudget(t *testing.T) {
	budget := 5 * time.Millisecond
	s := NewUISchedulerWithBudget(budget)

	if s == nil {
		t.Fatal("NewUISchedulerWithBudget() returned nil")
	}
}

// TestScheduleUpdate tests scheduling fiber updates
func TestScheduleUpdate(t *testing.T) {
	s := NewUIScheduler()
	fiber := CreateFiber(Text("Test"))

	s.ScheduleUpdate(fiber, LaneSyncLane)

	count := s.TotalDirtyCount()
	if count != 1 {
		t.Errorf("DirtyCount = %d, want 1", count)
	}
}

// TestScheduleFiberTree tests scheduling entire tree
func TestScheduleFiberTree(t *testing.T) {
	s := NewUIScheduler()

	tree := VStack(
		Text("A"),
		Text("B"),
		Text("C"),
	)

	fiber := CreateFiberFromVNode(tree)
	// Mark fibers with effects so they get scheduled
	fiber.Flags = EffectUpdate

	s.ScheduleFiberTree(fiber, LaneDefaultLane)

	count := s.TotalDirtyCount()
	// Should have at least the root fiber
	if count < 1 {
		t.Errorf("DirtyCount = %d, want at least 1", count)
	}
}

// TestBatching tests batch mode
func TestBatching(t *testing.T) {
	s := NewUIScheduler()

	if s.IsBatching() {
		t.Error("Should not be batching initially")
	}

	s.BeginBatch()
	if !s.IsBatching() {
		t.Error("Should be batching after BeginBatch()")
	}

	fiber := CreateFiber(Text("Test"))
	s.ScheduleUpdate(fiber, LaneSyncLane)

	// During batching, updates may not be immediately visible
	s.EndBatch(true)

	// After flush, should be visible
	count := s.TotalDirtyCount()
	if count == 0 {
		t.Error("Should have dirty nodes after flush")
	}
}

// TestClear tests clearing dirty nodes
func TestClear(t *testing.T) {
	s := NewUIScheduler()
	fiber := CreateFiber(Text("Test"))

	s.ScheduleUpdate(fiber, LaneSyncLane)
	if s.TotalDirtyCount() == 0 {
		t.Error("Should have dirty nodes")
	}

	s.Clear()
	if s.TotalDirtyCount() != 0 {
		t.Error("Should have no dirty nodes after Clear()")
	}
}

// TestLaneToPriority tests lane to priority conversion
func TestLaneToPriority(t *testing.T) {
	tests := []struct {
		name     string
		lane     Lane
		expected priority.DirtyLevel
	}{
		{"SyncLane", LaneSyncLane, priority.DirtyHigh},
		{"InputContinuous", LaneInputContinuousLane, priority.DirtyHigh},
		{"Default", LaneDefaultLane, priority.DirtyNormal},
		{"Idle", LaneIdleLane, priority.DirtyLow},
		{"NoLane", LaneNoLane, priority.DirtyLow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := laneToPriority(tt.lane)
			if result != tt.expected {
				t.Errorf("laneToPriority(%v) = %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}

// TestGetFiberID tests fiber ID generation
func TestGetFiberID(t *testing.T) {
	fiber := &Fiber{
		Tag:  "test",
		Type: VNodeText,
		Key:  "my-key",
	}

	id := getFiberID(fiber)
	if id == "" {
		t.Error("getFiberID() should not return empty")
	}

	// With key, should use key
	if fiber.Key != "" {
		expected := "fiber:my-key"
		if id != expected {
			t.Errorf("getFiberID() = %v, want %v", id, expected)
		}
	}
}

// TestFiberRendererAdapter tests the renderer adapter
func TestFiberRendererAdapter(t *testing.T) {
	layoutCalled := false
	paintCalled := false

	renderer := &DefaultFiberRenderer{
		onLayout: func(fiber *Fiber) {
			layoutCalled = true
		},
		onPaint: func(fiber *Fiber) {
			paintCalled = true
		},
	}

	adapter := &fiberRendererAdapter{renderer: renderer}
	fiber := CreateFiber(Text("Test"))

	adapter.Layout(fiber)
	adapter.Paint(fiber) // Also call Paint

	if !layoutCalled {
		t.Error("Layout should be called")
	}
	if !paintCalled {
		t.Error("Paint should be called")
	}
}

// TestTimeSlice tests time slicing
func TestTimeSlice(t *testing.T) {
	budget := 10 * time.Millisecond
	ts := NewTimeSlice(budget)

	if !ts.ShouldContinue() {
		t.Error("ShouldContinue should return true initially")
	}

	// Test that ShouldContinue returns false after budget is exceeded
	ts.deadline = time.Now().Add(-time.Millisecond) // Set deadline in the past

	if ts.ShouldContinue() {
		t.Error("ShouldContinue should return false when budget exceeded")
	}

	// Test that Elapsed reports correctly
	if ts.Elapsed() < 0 {
		t.Error("Elapsed should be positive when deadline is in the past")
	}
}

// TestThrottlerAdapter tests throttler adapter
func TestThrottlerAdapter(t *testing.T) {
	throttler := NewThrottlerAdapter(60)

	// First render should always be allowed
	if !throttler.ShouldRender() {
		t.Error("First render should be allowed")
	}

	// Immediate second render should be throttled
	if throttler.ShouldRender() {
		t.Error("Immediate second render should be throttled")
	}

	// Wait and try again
	time.Sleep(20 * time.Millisecond)
	if !throttler.ShouldRender() {
		t.Error("Render should be allowed after waiting")
	}
}

// TestRenderController tests render controller
func TestRenderController(t *testing.T) {
	controller := NewRenderController()

	if controller == nil {
		t.Fatal("NewRenderController() returned nil")
	}

	// Test strategy
	controller.SetStrategy(StrategyAlways)
	// With no throttling, should render
	if !controller.throttler.ShouldRender() {
		t.Error("Should render with StrategyAlways")
	}

	// Test FPS
	controller.SetTargetFPS(30)
	if controller.throttler.FPS() != 30 {
		t.Errorf("FPS = %d, want 30", controller.throttler.FPS())
	}
}

// TestWorkLoop tests work loop
func TestWorkLoop(t *testing.T) {
	callCount := 0

	renderer := &DefaultFiberRenderer{
		onLayout: func(fiber *Fiber) {
			callCount++
		},
		onPaint: func(fiber *Fiber) {
			callCount++
		},
	}

	loop := NewWorkLoop(renderer)
	if loop.IsRunning() {
		t.Error("Should not be running initially")
	}

	loop.Start()
	if !loop.IsRunning() {
		t.Error("Should be running after Start()")
	}

	fiber := CreateFiber(Text("Test"))
	loop.SetRoot(fiber)
	loop.Invalidate(LaneSyncLane)

	result := loop.ProcessFrame()
	if result.Processed < 0 {
		t.Errorf("Processed = %d, want >= 0", result.Processed)
	}

	loop.Stop()
	if loop.IsRunning() {
		t.Error("Should not be running after Stop()")
	}
}

// BenchmarkScheduleUpdate benchmarks update scheduling
func BenchmarkScheduleUpdate(b *testing.B) {
	s := NewUIScheduler()
	fiber := CreateFiber(Text("Test"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.BeginBatch()
		s.ScheduleUpdate(fiber, LaneSyncLane)
		s.EndBatch(true)
		s.Clear()
	}
}

// BenchmarkFiberTreeScheduling benchmarks tree scheduling
func BenchmarkFiberTreeScheduling(b *testing.B) {
	s := NewUIScheduler()

	var children []VNode
	for i := 0; i < 50; i++ {
		children = append(children, Text("Item"))
	}
	tree := VStack(children...)
	fiber := CreateFiberFromVNode(tree)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.ScheduleFiberTree(fiber, LaneDefaultLane)
		s.Clear()
	}
}
