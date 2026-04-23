package animation

import (
	"testing"
	"time"
)

func TestTweenDriver_ProgressAndCompletion(t *testing.T) {
	driver := NewTweenDriver(TweenDriverConfig{
		From:      0,
		To:        10,
		Duration:  100 * time.Millisecond,
		AutoStart: true,
	})

	start := time.Unix(0, 0)
	if changed := driver.Tick(start); changed {
		t.Fatal("first tick should only prime the driver")
	}
	if changed := driver.Tick(start.Add(50 * time.Millisecond)); !changed {
		t.Fatal("expected tween to advance at 50ms")
	}
	if got := driver.Value(); got != 5 {
		t.Fatalf("Value() = %v, want 5", got)
	}
	if driver.Done() {
		t.Fatal("driver should not be done at 50ms")
	}
	if changed := driver.Tick(start.Add(100 * time.Millisecond)); !changed {
		t.Fatal("expected tween completion to report change")
	}
	if got := driver.Value(); got != 10 {
		t.Fatalf("Value() = %v, want 10", got)
	}
	if !driver.Done() {
		t.Fatal("driver should be done at 100ms")
	}
	if driver.WantsTick() {
		t.Fatal("completed tween should not want more ticks")
	}
}

func TestTweenDriver_Delay(t *testing.T) {
	driver := NewTweenDriver(TweenDriverConfig{
		From:      1,
		To:        3,
		Duration:  100 * time.Millisecond,
		Delay:     30 * time.Millisecond,
		AutoStart: true,
	})

	start := time.Unix(0, 0)
	driver.Tick(start)
	if driver.Started() {
		t.Fatal("driver should still be delayed after priming tick")
	}
	if changed := driver.Tick(start.Add(20 * time.Millisecond)); changed {
		t.Fatal("driver should not change before delay elapses")
	}
	if driver.Started() {
		t.Fatal("driver should still be delayed at 20ms")
	}
	if changed := driver.Tick(start.Add(40 * time.Millisecond)); !changed {
		t.Fatal("driver should report delayed start")
	}
	if !driver.Started() {
		t.Fatal("driver should have started after delay")
	}
}

func TestLoopDriver_ProgressCyclesAndDelay(t *testing.T) {
	driver := NewLoopDriver(LoopDriverConfig{
		Duration:  80 * time.Millisecond,
		Delay:     20 * time.Millisecond,
		Cycles:    2,
		AutoStart: true,
	})

	start := time.Unix(0, 0)
	driver.Tick(start)
	if driver.Started() {
		t.Fatal("loop should be delayed after priming tick")
	}
	if changed := driver.Tick(start.Add(10 * time.Millisecond)); changed {
		t.Fatal("loop should not advance before delay")
	}
	if changed := driver.Tick(start.Add(30 * time.Millisecond)); !changed {
		t.Fatal("loop should report delayed start")
	}
	if !driver.Started() {
		t.Fatal("loop should have started after delay")
	}
	if got := driver.StepIndex(4); got != 0 {
		t.Fatalf("StepIndex() = %d, want 0 at loop start", got)
	}
	if changed := driver.Tick(start.Add(70 * time.Millisecond)); !changed {
		t.Fatal("loop should advance within first cycle")
	}
	if got := driver.StepIndex(4); got != 2 {
		t.Fatalf("StepIndex() = %d, want 2", got)
	}
	if changed := driver.Tick(start.Add(110 * time.Millisecond)); !changed {
		t.Fatal("loop should wrap into second cycle")
	}
	if got := driver.Cycle(); got != 1 {
		t.Fatalf("Cycle() = %d, want 1", got)
	}
	if changed := driver.Tick(start.Add(190 * time.Millisecond)); !changed {
		t.Fatal("loop should report completion on final cycle")
	}
	if !driver.Done() {
		t.Fatal("loop should be done after configured cycles")
	}
	if driver.WantsTick() {
		t.Fatal("completed loop should not want more ticks")
	}
}
