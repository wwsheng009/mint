package animation

import (
	"testing"
	"time"
)

func TestManagerStartStoresFrameInterval(t *testing.T) {
	mgr := NewManager()
	mgr.Start(20)
	defer mgr.Stop()

	if got := mgr.getFrameTime(); got != 50*time.Millisecond {
		t.Fatalf("getFrameTime() = %v, want %v", got, 50*time.Millisecond)
	}
	if !mgr.IsRunning() {
		t.Fatal("manager should be running after Start")
	}
}

func TestManagerUpdateWithDeltaAdvancesAndCleansCompleted(t *testing.T) {
	mgr := NewManager()
	anim := NewAnimation("linear", AnimationCustom, 100*time.Millisecond).
		WithFromTo(0.0, 10.0).
		WithEasing(Linear)

	mgr.Add(anim)
	if !mgr.StartAnimation(anim.ID) {
		t.Fatal("StartAnimation should succeed")
	}

	mgr.UpdateWithDelta(50 * time.Millisecond)
	if got, ok := anim.Current.(float64); !ok || got != 5 {
		t.Fatalf("Current = %#v, want 5", anim.Current)
	}
	if anim.State != AnimationStateRunning {
		t.Fatalf("State = %q, want %q", anim.State, AnimationStateRunning)
	}
	if mgr.Count() != 1 {
		t.Fatalf("Count() = %d, want 1 before completion", mgr.Count())
	}

	mgr.UpdateWithDelta(50 * time.Millisecond)
	if anim.State != AnimationStateCompleted {
		t.Fatalf("State = %q, want %q", anim.State, AnimationStateCompleted)
	}
	if mgr.Count() != 0 {
		t.Fatalf("Count() = %d, want 0 after completion cleanup", mgr.Count())
	}
}

func TestManagerTickUsesTimeAnchors(t *testing.T) {
	mgr := NewManager()
	anim := NewAnimation("tick", AnimationCustom, 100*time.Millisecond).
		WithFromTo(0.0, 100.0).
		WithEasing(Linear)

	mgr.Add(anim)
	if !mgr.StartAnimation(anim.ID) {
		t.Fatal("StartAnimation should succeed")
	}

	start := time.Unix(0, 0)
	mgr.Tick(start)
	if anim.Elapsed != 0 {
		t.Fatalf("Elapsed after priming tick = %v, want 0", anim.Elapsed)
	}

	mgr.Tick(start.Add(25 * time.Millisecond))
	if got, ok := anim.Current.(float64); !ok || got != 25 {
		t.Fatalf("Current after 25ms = %#v, want 25", anim.Current)
	}
}

func TestManagerRepeatAlternateCompletesAndCleans(t *testing.T) {
	mgr := NewManager()
	anim := NewAnimation("alternate", AnimationCustom, 40*time.Millisecond).
		WithFromTo(0, 8).
		WithEasing(Linear).
		WithRepeat(2).
		WithAlternate(true)

	mgr.Add(anim)
	if !mgr.StartAnimation(anim.ID) {
		t.Fatal("StartAnimation should succeed")
	}

	mgr.UpdateWithDelta(40 * time.Millisecond)
	if anim.State != AnimationStateRunning {
		t.Fatalf("State after first cycle = %q, want %q", anim.State, AnimationStateRunning)
	}
	if anim.From != 0 || anim.To != 8 {
		t.Fatalf("config range = (%v -> %v), want original (0 -> 8)", anim.From, anim.To)
	}

	mgr.UpdateWithDelta(40 * time.Millisecond)
	if anim.State != AnimationStateCompleted {
		t.Fatalf("State after final cycle = %q, want %q", anim.State, AnimationStateCompleted)
	}
	if anim.Current != 0 {
		t.Fatalf("Current after alternate completion = %#v, want 0", anim.Current)
	}
	if mgr.Count() != 0 {
		t.Fatalf("Count() = %d, want 0 after repeated animation cleanup", mgr.Count())
	}
}

func TestManagerDelayOverflowAdvancesSameTick(t *testing.T) {
	mgr := NewManager()
	anim := NewAnimation("delay-overflow", AnimationCustom, 100*time.Millisecond).
		WithFromTo(0.0, 10.0).
		WithDelay(10 * time.Millisecond).
		WithEasing(Linear)

	mgr.Add(anim)
	if !mgr.StartAnimation(anim.ID) {
		t.Fatal("StartAnimation should succeed")
	}

	mgr.UpdateWithDelta(16 * time.Millisecond)
	if got, ok := anim.Current.(float64); !ok || got != 0.6 {
		t.Fatalf("Current after 16ms with 10ms delay = %#v, want 0.6", anim.Current)
	}
}

func TestManagerRestartRestoresOriginalConfig(t *testing.T) {
	mgr := NewManager()
	anim := NewAnimation("restart", AnimationCustom, 100*time.Millisecond).
		WithFromTo(0.0, 10.0).
		WithDelay(20 * time.Millisecond).
		WithRepeat(2).
		WithAlternate(true).
		WithEasing(Linear)

	mgr.Add(anim)
	if !mgr.StartAnimation(anim.ID) {
		t.Fatal("StartAnimation should succeed")
	}

	mgr.UpdateWithDelta(20 * time.Millisecond)
	mgr.UpdateWithDelta(100 * time.Millisecond)
	if anim.State != AnimationStateRunning {
		t.Fatalf("State after first play = %q, want %q", anim.State, AnimationStateRunning)
	}
	if anim.Current != 10.0 {
		t.Fatalf("Current after first play = %#v, want 10", anim.Current)
	}

	if !mgr.StopAnimation(anim.ID) {
		t.Fatal("StopAnimation should succeed")
	}
	if anim.Delay != 20*time.Millisecond {
		t.Fatalf("Delay after stop = %v, want %v", anim.Delay, 20*time.Millisecond)
	}
	if anim.Repeat != 2 {
		t.Fatalf("Repeat after stop = %d, want 2", anim.Repeat)
	}
	if anim.From != 0.0 || anim.To != 10.0 {
		t.Fatalf("range after stop = (%v -> %v), want original (0 -> 10)", anim.From, anim.To)
	}

	if !mgr.StartAnimation(anim.ID) {
		t.Fatal("StartAnimation should restart successfully")
	}
	mgr.UpdateWithDelta(20 * time.Millisecond)
	if anim.Current != 0.0 {
		t.Fatalf("Current after restart delay = %#v, want 0", anim.Current)
	}
	mgr.UpdateWithDelta(50 * time.Millisecond)
	if got, ok := anim.Current.(float64); !ok || got != 5 {
		t.Fatalf("Current after restart progress = %#v, want 5", anim.Current)
	}
}

func TestManagerPauseStopAndCancelAnimation(t *testing.T) {
	mgr := NewManager()
	anim := NewAnimation("control", AnimationCustom, 100*time.Millisecond).
		WithFromTo(1.0, 2.0)
	mgr.Add(anim)

	if !mgr.StartAnimation(anim.ID) {
		t.Fatal("StartAnimation should succeed")
	}
	if !mgr.PauseAnimation(anim.ID) {
		t.Fatal("PauseAnimation should succeed")
	}
	if anim.State != AnimationStatePaused {
		t.Fatalf("State after pause = %q, want %q", anim.State, AnimationStatePaused)
	}
	if !mgr.StopAnimation(anim.ID) {
		t.Fatal("StopAnimation should succeed")
	}
	if anim.State != AnimationStateIdle {
		t.Fatalf("State after stop = %q, want %q", anim.State, AnimationStateIdle)
	}
	if anim.Elapsed != 0 {
		t.Fatalf("Elapsed after stop = %v, want 0", anim.Elapsed)
	}
	if anim.Current != anim.From {
		t.Fatalf("Current after stop = %#v, want %#v", anim.Current, anim.From)
	}
	if !mgr.CancelAnimation(anim.ID) {
		t.Fatal("CancelAnimation should succeed")
	}
	if mgr.Count() != 0 {
		t.Fatalf("Count() after cancel = %d, want 0", mgr.Count())
	}
}
