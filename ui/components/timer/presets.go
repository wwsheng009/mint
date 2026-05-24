package timer

import (
	"strings"
	"time"
)

const (
	defaultTimerProgressWidth = 12
)

// AutoRefresh creates a countdown timer for periodic refresh loops.
func AutoRefresh(label string, interval time.Duration) *VNode {
	label = timerPresetLabel(label, "Refresh")
	interval = normalizeDuration(interval)
	return NewBuilder().
		Label(label).
		Countdown(interval).
		ShowProgress(true).
		ProgressWidth(defaultTimerProgressWidth).
		WarningBelow(timerWarningWindow(interval)).
		ExpiredText("now").
		BuildTyped()
}

// RetryAfter creates a countdown timer for retry-after or cooldown windows.
func RetryAfter(label string, deadline time.Time) *VNode {
	label = timerPresetLabel(label, "Retry")
	return NewBuilder().
		Label(label).
		Until(deadline).
		ShowProgress(true).
		ProgressWidth(defaultTimerProgressWidth).
		WarningBelow(10 * time.Second).
		ExpiredText("ready").
		BuildTyped()
}

// OperationElapsed creates an elapsed timer for a running operation.
func OperationElapsed(label string, startedAt time.Time) *VNode {
	label = timerPresetLabel(label, "Elapsed")
	return NewBuilder().
		Label(label).
		Elapsed().
		StartedAt(startedAt).
		BuildTyped()
}

func timerPresetLabel(label, fallback string) string {
	label = strings.TrimSpace(strings.NewReplacer(
		"\r\n", " ",
		"\n", " ",
		"\r", " ",
		"\t", " ",
	).Replace(label))
	if label == "" {
		return fallback
	}
	return label
}

func timerWarningWindow(interval time.Duration) time.Duration {
	interval = normalizeDuration(interval)
	if interval <= 0 {
		return 0
	}
	if interval <= 15*time.Second {
		return interval / 3
	}
	if interval <= time.Minute {
		return 5 * time.Second
	}
	return 10 * time.Second
}
