package timer

import (
	"testing"
	"time"
)

func TestTimerOperationalPresets(t *testing.T) {
	refresh := AutoRefresh("", 30*time.Second)
	if refresh.Label() != "Refresh" {
		t.Fatalf("AutoRefresh label = %q, want Refresh", refresh.Label())
	}
	if refresh.Mode() != ModeCountdown || refresh.Duration() != 30*time.Second {
		t.Fatalf("AutoRefresh mode/duration = %v/%s, want countdown/30s", refresh.Mode(), refresh.Duration())
	}
	if !refresh.ShowProgress() || refresh.ProgressWidth() != defaultTimerProgressWidth {
		t.Fatalf("AutoRefresh progress = show:%v width:%d", refresh.ShowProgress(), refresh.ProgressWidth())
	}
	if refresh.WarningBelow() != 5*time.Second || refresh.ExpiredText() != "now" {
		t.Fatalf("AutoRefresh warning/expired = %s/%q, want 5s/now", refresh.WarningBelow(), refresh.ExpiredText())
	}

	deadline := time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)
	retry := RetryAfter("Retry provider", deadline)
	if retry.Label() != "Retry provider" || retry.Deadline() != deadline {
		t.Fatalf("RetryAfter label/deadline = %q/%s", retry.Label(), retry.Deadline())
	}
	if retry.ExpiredText() != "ready" || !retry.ShowProgress() {
		t.Fatalf("RetryAfter expired/progress = %q/%v, want ready/true", retry.ExpiredText(), retry.ShowProgress())
	}

	startedAt := time.Date(2026, 5, 25, 9, 0, 0, 0, time.UTC)
	elapsed := OperationElapsed("\n", startedAt)
	if elapsed.Label() != "Elapsed" || elapsed.Mode() != ModeElapsed || elapsed.StartedAt() != startedAt {
		t.Fatalf("OperationElapsed = label:%q mode:%v started:%s", elapsed.Label(), elapsed.Mode(), elapsed.StartedAt())
	}
}

func TestTimerWarningWindow(t *testing.T) {
	tests := []struct {
		interval time.Duration
		want     time.Duration
	}{
		{interval: 0, want: 0},
		{interval: 9 * time.Second, want: 3 * time.Second},
		{interval: 30 * time.Second, want: 5 * time.Second},
		{interval: 2 * time.Minute, want: 10 * time.Second},
	}

	for _, tt := range tests {
		if got := timerWarningWindow(tt.interval); got != tt.want {
			t.Fatalf("timerWarningWindow(%s) = %s, want %s", tt.interval, got, tt.want)
		}
	}
}
