package progress

import "testing"

func TestStatusForState(t *testing.T) {
	tests := []struct {
		state string
		want  Status
	}{
		{state: "healthy", want: StatusSuccess},
		{state: "completed", want: StatusSuccess},
		{state: "running", want: StatusActive},
		{state: "syncing", want: StatusActive},
		{state: "pending_restart", want: StatusWarning},
		{state: "rate limited", want: StatusWarning},
		{state: "failed", want: StatusException},
		{state: "out_of_sync", want: StatusException},
		{state: "custom", want: StatusNormal},
	}

	for _, tt := range tests {
		if got := StatusForState(tt.state); got != tt.want {
			t.Fatalf("StatusForState(%q) = %v, want %v", tt.state, got, tt.want)
		}
	}
}

func TestProgressPresets(t *testing.T) {
	if got := ForState("Reload", 0, 0, "running"); !got.IsIndeterminate() || got.Status() != StatusActive {
		t.Fatalf("ForState running = indeterminate:%v status:%v, want active indeterminate", got.IsIndeterminate(), got.Status())
	}
	if got := ForState("Sync", 50, 100, "pending_restart"); got.Status() != StatusWarning {
		t.Fatalf("ForState pending_restart status = %v, want warning", got.Status())
	}
	if got := ForStateWithValue("Sync", 50, 100, "pending_restart", "items"); got.Status() != StatusWarning || !got.ShowValue() || got.Unit() != "items" {
		t.Fatalf("ForStateWithValue = status:%v showValue:%v unit:%q", got.Status(), got.ShowValue(), got.Unit())
	}
	if got := OperationalValue("runtime.sections", "Sections", 3, 4, "effective", "sections", 36); got.Key() != "runtime.sections" || got.Width() != 36 || got.Status() != StatusSuccess || !got.ShowValue() || got.Unit() != "sections" {
		t.Fatalf("OperationalValue = key:%q width:%d status:%v showValue:%v unit:%q", got.Key(), got.Width(), got.Status(), got.ShowValue(), got.Unit())
	}
	if got := Usage("CPU", 79, 100); got.Status() != StatusNormal {
		t.Fatalf("Usage 79%% status = %v, want normal", got.Status())
	}
	if got := Usage("CPU", 80, 100); got.Status() != StatusWarning {
		t.Fatalf("Usage 80%% status = %v, want warning", got.Status())
	}
	if got := Usage("CPU", 95, 100); got.Status() != StatusException {
		t.Fatalf("Usage 95%% status = %v, want exception", got.Status())
	}
	if got := UsageWithThresholds("Queue", 7, 10, 60, 90); got.Status() != StatusWarning {
		t.Fatalf("UsageWithThresholds status = %v, want warning", got.Status())
	}
	if got := UsageWithValue("Queue", 7, 10, "jobs"); got.Status() != StatusNormal || !got.ShowValue() || got.Unit() != "jobs" {
		t.Fatalf("UsageWithValue = status:%v showValue:%v unit:%q", got.Status(), got.ShowValue(), got.Unit())
	}
	if got := UsageWithValueThresholds("Queue", 9, 10, "jobs", 60, 90); got.Status() != StatusException || !got.ShowValue() || got.Unit() != "jobs" {
		t.Fatalf("UsageWithValueThresholds = status:%v showValue:%v unit:%q", got.Status(), got.ShowValue(), got.Unit())
	}
	if got := Busy("Loading"); !got.IsIndeterminate() || got.Status() != StatusActive {
		t.Fatalf("Busy = indeterminate:%v status:%v, want active indeterminate", got.IsIndeterminate(), got.Status())
	}
	if got := BusyWithKey("manager.feedback", "loading", 28); got.Key() != "manager.feedback" || got.Width() != 28 || !got.IsIndeterminate() || got.Status() != StatusActive {
		t.Fatalf("BusyWithKey = key:%q width:%d indeterminate:%v status:%v", got.Key(), got.Width(), got.IsIndeterminate(), got.Status())
	}
	if got := Complete("Done"); got.Value() != 100 || got.Status() != StatusSuccess {
		t.Fatalf("Complete = value:%d status:%v, want 100 success", got.Value(), got.Status())
	}
	if got := Failed("Failed"); got.Value() != 100 || got.Status() != StatusException {
		t.Fatalf("Failed = value:%d status:%v, want 100 exception", got.Value(), got.Status())
	}
}
