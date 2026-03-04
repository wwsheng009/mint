package scheduler

import (
	"sync/atomic"
	"testing"
	"time"
)

// =============================================================================
// Basic Scheduler Tests
// =============================================================================

func TestScheduler_NewScheduler(t *testing.T) {
	s := NewScheduler()
	if s == nil {
		t.Fatal("expected scheduler to be created")
	}
	defer s.Shutdown()

	if s.HasPendingWork() {
		t.Error("expected no pending work on new scheduler")
	}
}

func TestScheduler_ScheduleFunc(t *testing.T) {
	s := NewScheduler()
	defer s.Shutdown()

	var executed int32
	task := s.ScheduleFunc(SyncLane, func(shouldYield ShouldYieldFunc) bool {
		atomic.AddInt32(&executed, 1)
		return true
	})

	if task == nil {
		t.Fatal("expected task to be returned")
	}
	if task.ID == 0 {
		t.Error("expected task to have non-zero ID")
	}
	if task.Lane != SyncLane {
		t.Errorf("expected lane %v, got %v", SyncLane, task.Lane)
	}
	if task.Canceled {
		t.Error("expected task not to be canceled")
	}

	if !s.HasPendingWork() {
		t.Error("expected pending work after schedule")
	}
	if s.GetQueueLength(SyncLane) != 1 {
		t.Errorf("expected 1 task in SyncLane, got %d", s.GetQueueLength(SyncLane))
	}
}

func TestScheduler_Flush(t *testing.T) {
	s := NewScheduler()
	defer s.Shutdown()

	var executed int32
	s.ScheduleFunc(SyncLane, func(shouldYield ShouldYieldFunc) bool {
		atomic.AddInt32(&executed, 1)
		return true
	})
	s.ScheduleFunc(InputLane, func(shouldYield ShouldYieldFunc) bool {
		atomic.AddInt32(&executed, 1)
		return true
	})
	s.ScheduleFunc(DefaultLane, func(shouldYield ShouldYieldFunc) bool {
		atomic.AddInt32(&executed, 1)
		return true
	})

	s.Flush()

	if atomic.LoadInt32(&executed) != 3 {
		t.Errorf("expected 3 tasks executed, got %d", executed)
	}
	if s.HasPendingWork() {
		t.Error("expected no pending work after flush")
	}
	if s.GetTotalQueueLength() != 0 {
		t.Errorf("expected 0 total tasks, got %d", s.GetTotalQueueLength())
	}
}

func TestScheduler_PriorityOrder(t *testing.T) {
	s := NewScheduler()
	defer s.Shutdown()

	var order []string

	// Schedule in reverse priority order
	s.ScheduleFunc(IdleLane, func(shouldYield ShouldYieldFunc) bool {
		order = append(order, "idle")
		return true
	})
	s.ScheduleFunc(DefaultLane, func(shouldYield ShouldYieldFunc) bool {
		order = append(order, "default")
		return true
	})
	s.ScheduleFunc(InputLane, func(shouldYield ShouldYieldFunc) bool {
		order = append(order, "input")
		return true
	})
	s.ScheduleFunc(SyncLane, func(shouldYield ShouldYieldFunc) bool {
		order = append(order, "sync")
		return true
	})

	s.Flush()

	// Verify priority order: Sync > Input > Default > Idle
	expected := []string{"sync", "input", "default", "idle"}
	if len(order) != len(expected) {
		t.Fatalf("expected %d executions, got %d", len(expected), len(order))
	}
	for i, v := range expected {
		if order[i] != v {
			t.Errorf("position %d: expected %s, got %s", i, v, order[i])
		}
	}
}

func TestScheduler_Cancel(t *testing.T) {
	s := NewScheduler()
	defer s.Shutdown()

	var executed int32
	task := s.ScheduleFunc(SyncLane, func(shouldYield ShouldYieldFunc) bool {
		atomic.AddInt32(&executed, 1)
		return true
	})

	task.Cancel()
	s.Flush()

	if atomic.LoadInt32(&executed) != 0 {
		t.Error("expected canceled task not to execute")
	}
}

func TestScheduler_GetPendingLanes(t *testing.T) {
	s := NewScheduler()
	defer s.Shutdown()

	s.ScheduleFunc(SyncLane, func(shouldYield ShouldYieldFunc) bool { return true })
	s.ScheduleFunc(InputLane, func(shouldYield ShouldYieldFunc) bool { return true })
	s.ScheduleFunc(IdleLane, func(shouldYield ShouldYieldFunc) bool { return true })

	lanes := s.GetPendingLanes()

	if lanes&SyncLane == 0 {
		t.Error("expected SyncLane to be pending")
	}
	if lanes&InputLane == 0 {
		t.Error("expected InputLane to be pending")
	}
	if lanes&IdleLane == 0 {
		t.Error("expected IdleLane to be pending")
	}
}

// =============================================================================
// Callback Tests
// =============================================================================

func TestScheduler_Callbacks(t *testing.T) {
	var startCount, completeCount, yieldCount int32

	s := NewScheduler(
		WithOnWorkStart(func(task *ScheduledTask) {
			atomic.AddInt32(&startCount, 1)
		}),
		WithOnWorkComplete(func(task *ScheduledTask) {
			atomic.AddInt32(&completeCount, 1)
		}),
		WithOnWorkYield(func(task *ScheduledTask) {
			atomic.AddInt32(&yieldCount, 1)
		}),
	)
	defer s.Shutdown()

	s.ScheduleFunc(SyncLane, func(shouldYield ShouldYieldFunc) bool {
		return true // completed
	})

	s.Flush()

	if atomic.LoadInt32(&startCount) != 1 {
		t.Errorf("expected 1 start callback, got %d", startCount)
	}
	if atomic.LoadInt32(&completeCount) != 1 {
		t.Errorf("expected 1 complete callback, got %d", completeCount)
	}
	if atomic.LoadInt32(&yieldCount) != 0 {
		t.Errorf("expected 0 yield callbacks, got %d", yieldCount)
	}
}

func TestScheduler_YieldCallback(t *testing.T) {
	var yieldCount int32

	s := NewScheduler(
		WithOnWorkYield(func(task *ScheduledTask) {
			atomic.AddInt32(&yieldCount, 1)
		}),
	)
	defer s.Shutdown()

	// Task that yields on first call, completes on second
	callCount := 0
	s.ScheduleFunc(DefaultLane, func(shouldYield ShouldYieldFunc) bool {
		callCount++
		if callCount == 1 {
			return false // yield
		}
		return true // complete
	})

	s.Flush()

	if atomic.LoadInt32(&yieldCount) == 0 {
		t.Error("expected at least one yield callback")
	}
}

// =============================================================================
// BatchScheduler Tests
// =============================================================================

func TestBatchScheduler_Basic(t *testing.T) {
	s := NewScheduler()
	defer s.Shutdown()

	batch := NewBatchScheduler(s, SyncLane)
	if batch.Count() != 0 {
		t.Error("expected empty batch")
	}

	var executed int32
	batch.AddFunc(func(shouldYield ShouldYieldFunc) bool {
		atomic.AddInt32(&executed, 1)
		return true
	})
	batch.AddFunc(func(shouldYield ShouldYieldFunc) bool {
		atomic.AddInt32(&executed, 1)
		return true
	})

	if batch.Count() != 2 {
		t.Errorf("expected 2 tasks in batch, got %d", batch.Count())
	}

	batch.Flush()

	if atomic.LoadInt32(&executed) != 2 {
		t.Errorf("expected 2 tasks executed, got %d", executed)
	}
}

func TestBatchScheduler_Cancel(t *testing.T) {
	s := NewScheduler()
	defer s.Shutdown()

	batch := NewBatchScheduler(s, SyncLane)

	var executed int32
	batch.AddFunc(func(shouldYield ShouldYieldFunc) bool {
		atomic.AddInt32(&executed, 1)
		return true
	})
	batch.AddFunc(func(shouldYield ShouldYieldFunc) bool {
		atomic.AddInt32(&executed, 1)
		return true
	})

	batch.Cancel()
	s.Flush()

	if atomic.LoadInt32(&executed) != 0 {
		t.Error("expected no tasks to execute after cancel")
	}
	if batch.Count() != 0 {
		t.Error("expected empty batch after cancel")
	}
}

// =============================================================================
// Lane Tests
// =============================================================================

func TestLane_Priority(t *testing.T) {
	tests := []struct {
		name     string
		lane     Lane
		expected int
	}{
		{"SyncLane", SyncLane, 0},
		{"InputLane", InputLane, 1},
		{"DefaultLane", DefaultLane, 3},
		{"TransitionLane", TransitionLane, 4},
		{"IdleLane", IdleLane, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.lane.Priority() != tt.expected {
				t.Errorf("%s: expected priority %d, got %d", tt.name, tt.expected, tt.lane.Priority())
			}
		})
	}
}

func TestLane_PickHighestPriority(t *testing.T) {
	tests := []struct {
		name     string
		lanes    Lane
		expected Lane
	}{
		{"single SyncLane", SyncLane, SyncLane},
		{"single IdleLane", IdleLane, IdleLane},
		{"multiple - Sync highest", SyncLane | InputLane | IdleLane, SyncLane},
		{"multiple - Input highest", InputLane | DefaultLane | IdleLane, InputLane},
		{"multiple - Default highest", DefaultLane | IdleLane, DefaultLane},
		{"no lanes", NoLane, NoLane},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PickHighestPriorityLane(tt.lanes)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestLane_IsLowerPriorityThan(t *testing.T) {
	if SyncLane.IsLowerPriorityThan(InputLane) {
		t.Error("SyncLane should not be lower priority than InputLane")
	}
	if !IdleLane.IsLowerPriorityThan(SyncLane) {
		t.Error("IdleLane should be lower priority than SyncLane")
	}
	if DefaultLane.IsLowerPriorityThan(DefaultLane) {
		t.Error("same lane should not be lower priority")
	}
}

func TestLane_GetDeadline(t *testing.T) {
	tests := []struct {
		name         string
		lane         Lane
		minExpected  time.Duration
	}{
		{"SyncLane", SyncLane, 0},
		{"InputLane", InputLane, 0},
		{"DefaultLane", DefaultLane, time.Millisecond},
		{"IdleLane", IdleLane, time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deadline := GetDeadline(tt.lane)
			if deadline < tt.minExpected {
				t.Errorf("%s: expected at least %v, got %v", tt.name, tt.minExpected, deadline)
			}
		})
	}
}

// =============================================================================
// PerformWork Tests
// =============================================================================

func TestScheduler_PerformWork(t *testing.T) {
	s := NewScheduler()
	defer s.Shutdown()

	var executed int32
	s.ScheduleFunc(SyncLane, func(shouldYield ShouldYieldFunc) bool {
		atomic.AddInt32(&executed, 1)
		return true
	})

	s.PerformWork()

	if atomic.LoadInt32(&executed) != 1 {
		t.Errorf("expected 1 task executed, got %d", executed)
	}
}

func TestScheduler_IsPerformingWork(t *testing.T) {
	s := NewScheduler()
	defer s.Shutdown()

	if s.IsPerformingWork() {
		t.Error("expected not performing work initially")
	}

	// This is tricky to test without race conditions
	// The flag is set during performWorkUntilDeadline
	s.ScheduleFunc(SyncLane, func(shouldYield ShouldYieldFunc) bool {
		// During execution, IsPerformingWork should be true
		return true
	})

	s.Flush()

	if s.IsPerformingWork() {
		t.Error("expected not performing work after flush")
	}
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestScheduler_MultipleFlushes(t *testing.T) {
	s := NewScheduler()
	defer s.Shutdown()

	var count int32

	for i := 0; i < 3; i++ {
		s.ScheduleFunc(SyncLane, func(shouldYield ShouldYieldFunc) bool {
			atomic.AddInt32(&count, 1)
			return true
		})
		s.Flush()
	}

	if atomic.LoadInt32(&count) != 3 {
		t.Errorf("expected 3 total executions, got %d", count)
	}
}

func TestScheduler_Shutdown(t *testing.T) {
	s := NewScheduler()

	// Schedule a long-running task
	started := make(chan struct{})
	blocked := make(chan struct{})
	s.ScheduleFunc(SyncLane, func(shouldYield ShouldYieldFunc) bool {
		close(started)
		<-blocked
		return true
	})

	// Start work in goroutine
	go s.Flush()

	// Wait for task to start
	<-started

	// Shutdown should stop the scheduler
	s.Shutdown()

	// Allow task to complete
	close(blocked)

	// After shutdown, new tasks should not execute
	var executed int32
	s.ScheduleFunc(SyncLane, func(shouldYield ShouldYieldFunc) bool {
		atomic.AddInt32(&executed, 1)
		return true
	})
	s.Flush()

	if atomic.LoadInt32(&executed) != 0 {
		t.Error("expected no execution after shutdown")
	}
}
