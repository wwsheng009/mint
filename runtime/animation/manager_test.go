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
	if anim.From != 8 || anim.To != 0 {
		t.Fatalf("alternate swap = (%v -> %v), want (8 -> 0)", anim.From, anim.To)
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
