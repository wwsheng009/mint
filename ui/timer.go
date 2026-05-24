package ui

import (
	"time"

	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/timer"
)

type TimerBuilder = timer.Builder
type TimerVNode = timer.VNode
type TimerMode = timer.Mode

const (
	TimerModeElapsed   = timer.ModeElapsed
	TimerModeCountdown = timer.ModeCountdown
)

// NewTimerBuilder creates a Timer builder.
func NewTimerBuilder() *timer.Builder {
	return timer.NewBuilder()
}

// ElapsedTimer creates a live elapsed timer.
func ElapsedTimer(label string, startedAt time.Time) rtui.VNode {
	return timer.NewBuilder().
		Label(label).
		Elapsed().
		StartedAt(startedAt).
		Build()
}

// CountdownTimer creates a live countdown timer.
func CountdownTimer(label string, duration time.Duration) rtui.VNode {
	return timer.NewBuilder().
		Label(label).
		Countdown(duration).
		Build()
}

// CountdownUntil creates a live countdown timer ending at a deadline.
func CountdownUntil(label string, deadline time.Time) rtui.VNode {
	return timer.NewBuilder().
		Label(label).
		Until(deadline).
		Build()
}
