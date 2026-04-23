package animation

import (
	"testing"
	"time"
)

func TestSequenceTracksChildAnimationsSequentially(t *testing.T) {
	first := NewAnimation("first", AnimationCustom, 100*time.Millisecond).
		WithFromTo(0.0, 10.0).
		WithEasing(Linear)
	second := NewAnimation("second", AnimationCustom, 200*time.Millisecond).
		WithFromTo(10.0, 30.0).
		WithEasing(Linear)

	seq := Sequence("seq", first, second)
	mgr := NewManager()
	mgr.Add(seq)
	if !mgr.StartAnimation(seq.ID) {
		t.Fatal("StartAnimation should succeed")
	}

	mgr.UpdateWithDelta(50 * time.Millisecond)
	if got, ok := seq.Current.(float64); !ok || got != 5 {
		t.Fatalf("Current at 50ms = %#v, want 5", seq.Current)
	}

	mgr.UpdateWithDelta(50 * time.Millisecond)
	if got, ok := seq.Current.(float64); !ok || got != 10 {
		t.Fatalf("Current at 100ms = %#v, want 10", seq.Current)
	}

	mgr.UpdateWithDelta(100 * time.Millisecond)
	if got, ok := seq.Current.(float64); !ok || got != 20 {
		t.Fatalf("Current at 200ms = %#v, want 20", seq.Current)
	}

	mgr.UpdateWithDelta(100 * time.Millisecond)
	if got, ok := seq.Current.(float64); !ok || got != 30 {
		t.Fatalf("Current at 300ms = %#v, want 30", seq.Current)
	}
	if seq.State != AnimationStateCompleted {
		t.Fatalf("State after sequence completion = %q, want %q", seq.State, AnimationStateCompleted)
	}
}

func TestSequenceDurationIncludesDelayAndRepeats(t *testing.T) {
	first := NewAnimation("first", AnimationCustom, 10*time.Millisecond).
		WithFromTo(0, 1).
		WithDelay(5 * time.Millisecond)
	second := NewAnimation("second", AnimationCustom, 20*time.Millisecond).
		WithFromTo(1, 2).
		WithRepeat(2).
		WithRepeatDelay(3 * time.Millisecond)

	seq := Sequence("seq", first, second)
	want := 5*time.Millisecond + 10*time.Millisecond + 20*time.Millisecond*2 + 3*time.Millisecond
	if seq.Duration != want {
		t.Fatalf("Sequence duration = %v, want %v", seq.Duration, want)
	}
	if seq.From != 0 {
		t.Fatalf("Sequence From = %#v, want 0", seq.From)
	}
	if seq.To != 2 {
		t.Fatalf("Sequence To = %#v, want 2", seq.To)
	}
}

func TestSequencePanicsOnInfiniteChildAnimation(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Sequence should panic when given an infinite child animation")
		}
	}()

	Sequence("seq",
		NewAnimation("pulse", AnimationCustom, 50*time.Millisecond).
			WithFromTo(0.0, 1.0).
			WithRepeat(-1),
	)
}
