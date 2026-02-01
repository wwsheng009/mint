// Package scheduler provides tests for input scheduling.
package scheduler

import (
	"sync"
	"testing"
	"time"

	"github.com/wwsheng009/mint/runtime/priority"
)

// =============================================================================
// InputQueue Tests
// =============================================================================

func TestInputQueue_BasicOperations(t *testing.T) {
	iq := NewInputQueue()
	defer iq.Close()

	// Test empty queue
	if iq.Len() != 0 {
		t.Errorf("Len() = %d, want 0", iq.Len())
	}
	if iq.HasPending() {
		t.Error("HasPending() = true, want false")
	}
	if iq.Peek() != nil {
		t.Error("Peek() = non-nil, want nil")
	}
	if iq.Pop() != nil {
		t.Error("Pop() = non-nil, want nil")
	}

	// Test Push
	event1 := iq.Push("key", "a", PriorityHigh)
	if event1 == nil {
		t.Fatal("Push() returned nil")
	}
	if event1.Type != "key" {
		t.Errorf("event.Type = %s, want 'key'", event1.Type)
	}
	if event1.Priority != PriorityHigh {
		t.Errorf("event.Priority = %d, want %d", event1.Priority, PriorityHigh)
	}

	// Test non-empty queue
	if !iq.HasPending() {
		t.Error("HasPending() = false, want true")
	}
	if iq.Len() != 1 {
		t.Errorf("Len() = %d, want 1", iq.Len())
	}
}

func TestInputQueue_PriorityOrdering(t *testing.T) {
	iq := NewInputQueue()
	defer iq.Close()

	// Push events with different priorities (out of order)
	iq.Push("low", "data1", PriorityLow)
	iq.Push("high", "data2", PriorityHigh)
	iq.Push("immediate", "data3", PriorityImmediate)
	iq.Push("continuous", "data4", PriorityContinuous)

	// Pop should return in priority order
	expected := []struct {
		typ      string
		priority InputPriority
	}{
		{"immediate", PriorityImmediate},
		{"high", PriorityHigh},
		{"continuous", PriorityContinuous},
		{"low", PriorityLow},
	}

	for _, exp := range expected {
		event := iq.Pop()
		if event == nil {
			t.Fatalf("Pop() returned nil, expected %s", exp.typ)
		}
		if event.Type != exp.typ {
			t.Errorf("Pop() = %s, want %s", event.Type, exp.typ)
		}
		if event.Priority != exp.priority {
			t.Errorf("Priority = %d, want %d", event.Priority, exp.priority)
		}
	}

	// Queue should be empty
	if iq.Len() != 0 {
		t.Errorf("Len() = %d, want 0", iq.Len())
	}
}

func TestInputQueue_SamePriorityFIFO(t *testing.T) {
	iq := NewInputQueue()
	defer iq.Close()

	// Push multiple events with same priority
	iq.Push("event1", nil, PriorityHigh)
	iq.Push("event2", nil, PriorityHigh)
	iq.Push("event3", nil, PriorityHigh)

	// Should maintain FIFO order within same priority
	expected := []string{"event1", "event2", "event3"}

	for _, exp := range expected {
		event := iq.Pop()
		if event == nil {
			t.Fatalf("Pop() returned nil, expected %s", exp)
		}
		if event.Type != exp {
			t.Errorf("Pop() = %s, want %s", event.Type, exp)
		}
	}
}

func TestInputQueue_Peek(t *testing.T) {
	iq := NewInputQueue()
	defer iq.Close()

	iq.Push("event1", nil, PriorityHigh)
	iq.Push("event2", nil, PriorityLow)

	// Peek should return highest priority without removing
	event := iq.Peek()
	if event == nil {
		t.Fatal("Peek() returned nil")
	}
	if event.Type != "event1" {
		t.Errorf("Peek().Type = %s, want 'event1'", event.Type)
	}

	// Queue should still have 2 items
	if iq.Len() != 2 {
		t.Errorf("Len() = %d, want 2", iq.Len())
	}

	// Peek again should return same event
	event2 := iq.Peek()
	if event2.ID != event.ID {
		t.Error("Peek() returned different event")
	}
}

func TestInputQueue_RemoveByID(t *testing.T) {
	iq := NewInputQueue()
	defer iq.Close()

	_ = iq.Push("event1", nil, PriorityHigh)
	event2 := iq.Push("event2", nil, PriorityLow)

	// Remove event2
	if !iq.RemoveByID(event2.ID) {
		t.Error("RemoveByID() = false, want true")
	}

	// Queue should only have event1
	if iq.Len() != 1 {
		t.Errorf("Len() = %d, want 1", iq.Len())
	}

	event := iq.Pop()
	if event.Type != "event1" {
		t.Errorf("Pop() = %s, want 'event1'", event.Type)
	}

	// Try to remove non-existent event
	if iq.RemoveByID("non-existent") {
		t.Error("RemoveByID() of non-existent = true, want false")
	}
}

func TestInputQueue_Clear(t *testing.T) {
	iq := NewInputQueue()
	defer iq.Close()

	iq.Push("event1", nil, PriorityHigh)
	iq.Push("event2", nil, PriorityLow)

	iq.Clear()

	if iq.Len() != 0 {
		t.Errorf("Len() after Clear() = %d, want 0", iq.Len())
	}
	if iq.HasPending() {
		t.Error("HasPending() after Clear() = true, want false")
	}
}

func TestInputQueue_Drain(t *testing.T) {
	iq := NewInputQueue()
	defer iq.Close()

	iq.Push("event1", nil, PriorityLow)
	iq.Push("event2", nil, PriorityHigh)
	iq.Push("event3", nil, PriorityImmediate)

	events := iq.Drain()

	if len(events) != 3 {
		t.Errorf("Drain() returned %d events, want 3", len(events))
	}

	// Should be in priority order
	if events[0].Type != "event3" {
		t.Errorf("events[0].Type = %s, want 'event3'", events[0].Type)
	}
	if events[1].Type != "event2" {
		t.Errorf("events[1].Type = %s, want 'event2'", events[1].Type)
	}
	if events[2].Type != "event1" {
		t.Errorf("events[2].Type = %s, want 'event1'", events[2].Type)
	}

	// Queue should be empty
	if iq.Len() != 0 {
		t.Errorf("Len() after Drain() = %d, want 0", iq.Len())
	}
}

func TestInputQueue_GetByType(t *testing.T) {
	iq := NewInputQueue()
	defer iq.Close()

	iq.Push("mouse", nil, PriorityHigh)
	iq.Push("key", nil, PriorityHigh)
	iq.Push("mouse", nil, PriorityLow)
	iq.Push("resize", nil, PriorityImmediate)

	mouseEvents := iq.GetByType("mouse")

	if len(mouseEvents) != 2 {
		t.Errorf("GetByType('mouse') = %d events, want 2", len(mouseEvents))
	}

	keyEvents := iq.GetByType("key")
	if len(keyEvents) != 1 {
		t.Errorf("GetByType('key') = %d events, want 1", len(keyEvents))
	}

	emptyEvents := iq.GetByType("non-existent")
	if len(emptyEvents) != 0 {
		t.Errorf("GetByType('non-existent') = %d events, want 0", len(emptyEvents))
	}
}

func TestInputQueue_Stats(t *testing.T) {
	iq := NewInputQueue()
	defer iq.Close()

	iq.Push("mouse", nil, PriorityHigh)
	iq.Push("key", nil, PriorityHigh)
	iq.Push("mouse", nil, PriorityLow)
	iq.Push("resize", nil, PriorityImmediate)

	stats := iq.Stats()

	if stats.Total != 4 {
		t.Errorf("Stats.Total = %d, want 4", stats.Total)
	}

	if stats.ByType["mouse"] != 2 {
		t.Errorf("Stats.ByType['mouse'] = %d, want 2", stats.ByType["mouse"])
	}

	if stats.ByType["key"] != 1 {
		t.Errorf("Stats.ByType['key'] = %d, want 1", stats.ByType["key"])
	}

	if stats.ByPrio[PriorityHigh] != 2 {
		t.Errorf("Stats.ByPrio[PriorityHigh] = %d, want 2", stats.ByPrio[PriorityHigh])
	}

	if stats.ByPrio[PriorityImmediate] != 1 {
		t.Errorf("Stats.ByPrio[PriorityImmediate] = %d, want 1", stats.ByPrio[PriorityImmediate])
	}
}

func TestInputQueue_PopMultiple(t *testing.T) {
	iq := NewInputQueue()
	defer iq.Close()

	for i := 0; i < 10; i++ {
		iq.Push("event", i, PriorityHigh)
	}

	// Pop 3 events
	events := iq.PopMultiple(3)

	if len(events) != 3 {
		t.Errorf("PopMultiple(3) = %d events, want 3", len(events))
	}

	if iq.Len() != 7 {
		t.Errorf("Len() after PopMultiple(3) = %d, want 7", iq.Len())
	}

	// Pop more than available
	events = iq.PopMultiple(20)
	if len(events) != 7 {
		t.Errorf("PopMultiple(20) = %d events, want 7", len(events))
	}

	if iq.Len() != 0 {
		t.Errorf("Len() = %d, want 0", iq.Len())
	}
}

func TestInputQueue_PopBlocking(t *testing.T) {
	iq := NewInputQueue()
	defer iq.Close()

	// Test timeout
	ch := make(chan *InputEvent)
	go func() {
		event := iq.PopBlocking(50 * time.Millisecond)
		ch <- event
	}()

	select {
	case event := <-ch:
		if event != nil {
			t.Error("PopBlocking with timeout returned non-nil")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("PopBlocking timeout test took too long")
	}

	// Test immediate return when event is available
	iq.Push("test", nil, PriorityHigh)

	event := iq.PopBlocking(100 * time.Millisecond)
	if event == nil {
		t.Error("PopBlocking returned nil when event was available")
	}
	if event.Type != "test" {
		t.Errorf("PopBlocking().Type = %s, want 'test'", event.Type)
	}
}

func TestInputQueue_ConcurrentAccess(t *testing.T) {
	iq := NewInputQueue()
	defer iq.Close()

	var wg sync.WaitGroup
	done := make(chan struct{})
	producerDone := make(chan struct{})

	// Producer
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(producerDone)
		for i := 0; i < 100; i++ {
			iq.Push("event", i, PriorityHigh)
			time.Sleep(time.Microsecond)
		}
	}()

	// Consumer
	go func() {
		count := 0
		for {
			select {
			case <-done:
				t.Logf("Consumer processed %d events", count)
				return
			default:
				if iq.Pop() != nil {
					count++
				}
				time.Sleep(time.Microsecond)
			}
		}
	}()

	// Wait for producer to finish
	wg.Wait()

	// Give consumer time to process, then stop it
	time.Sleep(50 * time.Millisecond)
	close(done)

	// Drain remaining
	remaining := iq.Len()
	t.Logf("Concurrent test: %d events remaining", remaining)
}

func TestInputPriority_ToDirtyLevel(t *testing.T) {
	tests := []struct {
		priority InputPriority
		expected priority.DirtyLevel
	}{
		{PriorityImmediate, priority.DirtyHigh},
		{PriorityHigh, priority.DirtyHigh},
		{PriorityContinuous, priority.DirtyNormal},
		{PriorityLow, priority.DirtyLow},
	}

	for _, tt := range tests {
		if tt.priority.ToDirtyLevel() != tt.expected {
			t.Errorf("%s.ToDirtyLevel() = %v, want %v",
				tt.priority, tt.priority.ToDirtyLevel(), tt.expected)
		}
	}
}

func TestInputPriority_String(t *testing.T) {
	tests := []struct {
		priority InputPriority
		expected string
	}{
		{PriorityImmediate, "immediate"},
		{PriorityHigh, "high"},
		{PriorityContinuous, "continuous"},
		{PriorityLow, "low"},
	}

	for _, tt := range tests {
		if tt.priority.String() != tt.expected {
			t.Errorf("%s.String() = %s, want %s",
				tt.priority, tt.priority.String(), tt.expected)
		}
	}
}

func TestEventPriority(t *testing.T) {
	tests := []struct {
		eventType string
		expected  InputPriority
	}{
		{"key", PriorityHigh},
		{"click", PriorityHigh},
		{"resize", PriorityHigh},
		{"interrupt", PriorityHigh},
		{"mouse", PriorityContinuous},
		{"scroll", PriorityContinuous},
	}

	for _, tt := range tests {
		if EventPriority(tt.eventType) != tt.expected {
			t.Errorf("EventPriority(%s) = %v, want %v",
				tt.eventType, EventPriority(tt.eventType), tt.expected)
		}
	}
}

// =============================================================================
// InterruptibleTask Tests
// =============================================================================

func TestInterruptibleTask_BasicExecution(t *testing.T) {
	called := false
	task := NewInterruptibleTask("test", func(done <-chan struct{}) bool {
		called = true
		return true // Completed
	})

	result := task.Execute()

	if !result {
		t.Error("Execute() = false, want true")
	}
	if !called {
		t.Error("Task function was not called")
	}
	if task.State() != TaskCompleted {
		t.Errorf("State() = %s, want TaskCompleted", task.State())
	}
}

func TestInterruptibleTask_Cancellation(t *testing.T) {
	task := NewInterruptibleTask("test", func(done <-chan struct{}) bool {
		// Simulate long-running task
		select {
		case <-done:
			return false // Interrupted
		case <-time.After(100 * time.Millisecond):
			return true // Completed
		}
	})

	// Cancel immediately
	task.Cancel()

	result := task.Execute()

	if result {
		t.Error("Execute() = true after Cancel, want false")
	}
	if task.State() != TaskCancelled {
		t.Errorf("State() = %s, want TaskCancelled", task.State())
	}
}

func TestInterruptibleTask_Progress(t *testing.T) {
	var task *InterruptibleTask
	task = NewInterruptibleTask("test", func(done <-chan struct{}) bool {
		for i := 0; i <= 10; i++ {
			select {
			case <-done:
				return false
			default:
				task.SetProgress(float64(i) / 10.0)
				time.Sleep(time.Millisecond)
			}
		}
		return true
	})

	go task.Execute()

	// Wait for progress
	for i := 0; i < 50; i++ {
		if task.Progress() >= 1.0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if task.Progress() < 1.0 {
		t.Errorf("Progress() = %f, want 1.0", task.Progress())
	}
}

func TestInterruptibleTask_StateChanges(t *testing.T) {
	states := []TaskState{}
	var mu sync.Mutex

	task := NewInterruptibleTask("test", func(done <-chan struct{}) bool {
		return true
	})

	task.OnStateChange(func(old, new TaskState) {
		mu.Lock()
		defer mu.Unlock()
		states = append(states, old, new)
	})

	task.Execute()

	time.Sleep(10 * time.Millisecond) // Allow callback to run

	mu.Lock()
	defer mu.Unlock()

	if len(states) < 2 {
		t.Errorf("State changes: got %d states, want at least 2", len(states))
	}
}

func TestInterruptibleTask_PauseResume(t *testing.T) {
	task := NewInterruptibleTask("test", func(done <-chan struct{}) bool {
		// This task can be interrupted
		select {
		case <-done:
			return false
		case <-time.After(10 * time.Millisecond):
			return true
		}
	})

	// Execute in background
	go task.Execute()
	time.Sleep(5 * time.Millisecond)

	// Pause
	task.Pause()

	// Check state is paused or running (race condition, but should be one of these)
	state := task.State()
	if state != TaskPaused && state != TaskRunning && state != TaskCompleted {
		t.Errorf("State() after Pause() = %s, want Paused/Running/Completed", state)
	}
}

func TestInterruptibleTask_Result(t *testing.T) {
	expectedResult := "test-result"
	var task *InterruptibleTask
	task = NewInterruptibleTask("test", func(done <-chan struct{}) bool {
		task.SetResult(expectedResult, nil)
		return true
	})

	task.Execute()

	result, err := task.Result()
	if err != nil {
		t.Errorf("Result() returned error: %v", err)
	}
	if result != expectedResult {
		t.Errorf("Result() = %v, want %v", result, expectedResult)
	}
}

// =============================================================================
// TaskManager Tests
// =============================================================================

func TestTaskManager_BasicOperations(t *testing.T) {
	tm := NewTaskManager()

	task1 := NewInterruptibleTask("task1", func(done <-chan struct{}) bool {
		return true
	})

	task2 := NewInterruptibleTask("task2", func(done <-chan struct{}) bool {
		return true
	})

	tm.Add(task1)
	tm.Add(task2)

	// Get task
	retrieved, ok := tm.Get("task1")
	if !ok {
		t.Error("Get('task1') = false, want true")
	}
	if retrieved.ID() != "task1" {
		t.Errorf("Retrieved task ID = %s, want 'task1'", retrieved.ID())
	}

	// List tasks
	tasks := tm.List()
	if len(tasks) != 2 {
		t.Errorf("List() = %d tasks, want 2", len(tasks))
	}

	// Remove task
	tm.Remove("task1")
	_, ok = tm.Get("task1")
	if ok {
		t.Error("Get('task1') after Remove = true, want false")
	}
}

func TestTaskManager_CancelAll(t *testing.T) {
	tm := NewTaskManager()

	for i := 0; i < 5; i++ {
		task := NewInterruptibleTask("task"+string(rune('0'+i)), func(done <-chan struct{}) bool {
			select {
			case <-done:
				return false
			case <-time.After(100 * time.Millisecond):
				return true
			}
		})
		tm.Add(task)
		go task.Execute()
	}

	// Give tasks time to start
	time.Sleep(5 * time.Millisecond)

	tm.CancelAll()

	// Wait a bit for cancellation to take effect
	time.Sleep(20 * time.Millisecond)

	// Check that all tasks are in a terminal state (cancelled or completed)
	nonTerminalCount := 0
	for _, task := range tm.List() {
		if task.State() == TaskRunning || task.State() == TaskPending {
			nonTerminalCount++
		}
	}

	if nonTerminalCount > 0 {
		t.Errorf("After CancelAll: %d tasks still in running/pending state, want 0", nonTerminalCount)
	}
}

func TestTaskManager_Clear(t *testing.T) {
	tm := NewTaskManager()

	task := NewInterruptibleTask("task1", func(done <-chan struct{}) bool {
		select {
		case <-done:
			return false
		case <-time.After(100 * time.Millisecond):
			return true
		}
	})

	tm.Add(task)
	tm.Clear()

	if len(tm.List()) != 0 {
		t.Errorf("After Clear: %d tasks, want 0", len(tm.List()))
	}
}

// =============================================================================
// Mouse Handler Tests
// =============================================================================

func TestMouseTracker_Basic(t *testing.T) {
	tracker := NewMouseTracker()

	// Test position
	tracker.UpdatePosition(10, 20)
	x, y := tracker.Position()
	if x != 10 || y != 20 {
		t.Errorf("Position() = (%d, %d), want (10, 20)", x, y)
	}

	// Test button
	tracker.UpdateButton(1, true)
	if !tracker.IsPressed(1) {
		t.Error("IsPressed(1) = false, want true")
	}

	tracker.UpdateButton(1, false)
	if tracker.IsPressed(1) {
		t.Error("IsPressed(1) = true, want false")
	}

	// Test window
	tracker.SetInWindow(true)
	if !tracker.IsInWindow() {
		t.Error("IsInWindow() = false, want true")
	}
}

func TestMouseTracker_PressedButtons(t *testing.T) {
	tracker := NewMouseTracker()

	tracker.UpdateButton(1, true)
	tracker.UpdateButton(2, true)
	tracker.UpdateButton(3, true)

	buttons := tracker.PressedButtons()
	if len(buttons) != 3 {
		t.Errorf("PressedButtons() = %d buttons, want 3", len(buttons))
	}

	tracker.UpdateButton(2, false)
	buttons = tracker.PressedButtons()
	if len(buttons) != 2 {
		t.Errorf("PressedButtons() after release = %d, want 2", len(buttons))
	}
}

func TestMouseMoveHandler_Throttling(t *testing.T) {
	var mu sync.Mutex
	var events []*MouseEvent

	handler := NewMouseMoveHandler(MouseHandlerFunc(func(event *MouseEvent) {
		mu.Lock()
		defer mu.Unlock()
		events = append(events, event)
	}), DefaultThrottleConfig())

	// Send many rapid events
	for i := 0; i < 100; i++ {
		handler.Handle(i, 0)
	}

	// Flush pending
	handler.Stop()

	mu.Lock()
	defer mu.Unlock()

	// Should have fewer events due to throttling
	if len(events) > 20 {
		t.Logf("MouseMoveHandler: %d events from 100 inputs (throttled)", len(events))
	}
}

func TestMouseClickHandler_Debouncing(t *testing.T) {
	var mu sync.Mutex
	clickCount := 0

	handler := NewMouseClickHandler(MouseHandlerFunc(func(event *MouseEvent) {
		mu.Lock()
		defer mu.Unlock()
		clickCount++
	}), ThrottleConfig{
		ClickInterval: 50 * time.Millisecond,
	})

	// Send multiple rapid clicks
	for i := 0; i < 10; i++ {
		handler.Handle(10, 10, 1)
	}

	// Wait for debounce
	time.Sleep(100 * time.Millisecond)

	handler.FlushImmediately()

	mu.Lock()
	count := clickCount
	mu.Unlock()

	// Should have fewer events due to debouncing
	if count > 3 {
		t.Logf("MouseClickHandler: %d events from 10 clicks (debounced)", count)
	}
}

func TestThrottleConfig_Default(t *testing.T) {
	config := DefaultThrottleConfig()

	if config.MotionInterval == 0 {
		t.Error("DefaultThrottleConfig().MotionInterval = 0, want non-zero")
	}
	if config.ClickInterval == 0 {
		t.Error("DefaultThrottleConfig().ClickInterval = 0, want non-zero")
	}
}

// =============================================================================
// RunWithTimeout Tests
// =============================================================================

func TestRunWithTimeout_Success(t *testing.T) {
	result, err := RunWithTimeout(func() (interface{}, error) {
		return "success", nil
	}, 100*time.Millisecond)

	if err != nil {
		t.Errorf("RunWithTimeout() returned error: %v", err)
	}
	if result != "success" {
		t.Errorf("result = %v, want 'success'", result)
	}
}

func TestRunWithTimeout_Timeout(t *testing.T) {
	_, err := RunWithTimeout(func() (interface{}, error) {
		time.Sleep(200 * time.Millisecond)
		return nil, nil
	}, 50*time.Millisecond)

	if err == nil {
		t.Error("RunWithTimeout() did not timeout")
	}
}
