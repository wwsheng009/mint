package ui

import (
	"testing"
	"time"
)

func TestFeedbackBuilderFactories(t *testing.T) {
	if vnode := NewAlertBuilder("Heads up").Warning().Build(); vnode == nil {
		t.Fatal("NewAlertBuilder().Build() returned nil")
	}
	if vnode := NewNotificationBuilder("Saved").Success().Build(); vnode == nil {
		t.Fatal("NewNotificationBuilder().Build() returned nil")
	}
	if vnode := NewTagBuilder("Beta").Primary().Build(); vnode == nil {
		t.Fatal("NewTagBuilder().Build() returned nil")
	}
	if vnode := NewSpinBuilder().Tip("Loading").Build(); vnode == nil {
		t.Fatal("NewSpinBuilder().Build() returned nil")
	}
	if vnode := NewClockBuilder().Radius(4).Build(); vnode == nil {
		t.Fatal("NewClockBuilder().Build() returned nil")
	}
	if vnode := NewTimerBuilder().Label("Refresh").Countdown(time.Minute).Build(); vnode == nil {
		t.Fatal("NewTimerBuilder().Build() returned nil")
	}
	if vnode := NewImageBuilder().Alt("Preview").SourceRGBA([]byte{255, 255, 255, 255}, 1, 1).Build(); vnode == nil {
		t.Fatal("NewImageBuilder().Build() returned nil")
	}
}

func TestFeedbackShortcuts(t *testing.T) {
	if vnode := Alert("Heads up"); vnode.Tag() != "alert" {
		t.Fatalf("Alert().Tag() = %q, want alert", vnode.Tag())
	}
	if vnode := Notification("Saved"); vnode.Tag() != "notification" {
		t.Fatalf("Notification().Tag() = %q, want notification", vnode.Tag())
	}
	if vnode := Tag("Beta"); vnode.Tag() != "tag" {
		t.Fatalf("Tag().Tag() = %q, want tag", vnode.Tag())
	}
	if vnode := Spin("Loading"); vnode.Tag() != "spin" {
		t.Fatalf("Spin().Tag() = %q, want spin", vnode.Tag())
	}
	if vnode := Clock(4); vnode.Tag() != "clock" {
		t.Fatalf("Clock().Tag() = %q, want clock", vnode.Tag())
	}
	if vnode := CountdownTimer("Refresh", time.Minute); vnode.Tag() != "timer" {
		t.Fatalf("CountdownTimer().Tag() = %q, want timer", vnode.Tag())
	}
	if vnode := AutoRefreshTimer("Refresh", time.Minute); vnode.Tag() != "timer" {
		t.Fatalf("AutoRefreshTimer().Tag() = %q, want timer", vnode.Tag())
	}
	if vnode := RetryAfterTimer("Retry", time.Now().Add(time.Minute)); vnode.Tag() != "timer" {
		t.Fatalf("RetryAfterTimer().Tag() = %q, want timer", vnode.Tag())
	}
	if vnode := CountdownUntilWithKey("jobs.next", "Next run", time.Now().Add(time.Minute), time.Now(), 34); vnode.Tag() != "timer" {
		t.Fatalf("CountdownUntilWithKey().Tag() = %q, want timer", vnode.Tag())
	} else if vnode.Key() != "jobs.next" || vnode.Props()["width"] != 34 || vnode.Props()["showProgress"] != true {
		t.Fatalf("CountdownUntilWithKey props = key:%q props:%+v", vnode.Key(), vnode.Props())
	}
	if vnode := OperationElapsedTimer("Reload", time.Now().Add(-time.Minute)); vnode.Tag() != "timer" {
		t.Fatalf("OperationElapsedTimer().Tag() = %q, want timer", vnode.Tag())
	}
	if vnode := OperationElapsedTimerWithKey("reload.elapsed", "Reload", time.Now().Add(-time.Minute), time.Now(), 36); vnode.Tag() != "timer" {
		t.Fatalf("OperationElapsedTimerWithKey().Tag() = %q, want timer", vnode.Tag())
	} else if vnode.Key() != "reload.elapsed" || vnode.Props()["width"] != 36 {
		t.Fatalf("OperationElapsedTimerWithKey props = key:%q props:%+v", vnode.Key(), vnode.Props())
	}
	if vnode := ProgressIndeterminate("Reloading"); vnode.Tag() != "progress" {
		t.Fatalf("ProgressIndeterminate().Tag() = %q, want progress", vnode.Tag())
	}
	if ProgressStatusForState("pending_restart") != ProgressStatusWarning {
		t.Fatalf("ProgressStatusForState(pending_restart) = %v, want warning", ProgressStatusForState("pending_restart"))
	}
	if vnode := ProgressForState("Sync", 50, 100, "pending_restart"); vnode.Tag() != "progress" {
		t.Fatalf("ProgressForState().Tag() = %q, want progress", vnode.Tag())
	}
	if vnode := ProgressForStateWithValue("Sync", 50, 100, "pending_restart", "items"); vnode.Tag() != "progress" {
		t.Fatalf("ProgressForStateWithValue().Tag() = %q, want progress", vnode.Tag())
	}
	if vnode := ProgressOperationalValue("runtime.sections", "Sections", 3, 4, "effective", "sections", 36); vnode.Tag() != "progress" {
		t.Fatalf("ProgressOperationalValue().Tag() = %q, want progress", vnode.Tag())
	} else if vnode.Key() != "runtime.sections" || vnode.Props()["width"] != 36 || vnode.Props()["showValue"] != true {
		t.Fatalf("ProgressOperationalValue props = key:%q props:%+v", vnode.Key(), vnode.Props())
	}
	if vnode := ProgressUsage("CPU", 82, 100); vnode.Tag() != "progress" {
		t.Fatalf("ProgressUsage().Tag() = %q, want progress", vnode.Tag())
	}
	if vnode := ProgressUsageWithValue("Queue", 7, 10, "jobs"); vnode.Tag() != "progress" {
		t.Fatalf("ProgressUsageWithValue().Tag() = %q, want progress", vnode.Tag())
	}
	if vnode := ProgressUsageWithValueThresholds("Queue", 9, 10, "jobs", 60, 90); vnode.Tag() != "progress" {
		t.Fatalf("ProgressUsageWithValueThresholds().Tag() = %q, want progress", vnode.Tag())
	}
	if vnode := ProgressBusy("Reloading"); vnode.Tag() != "progress" {
		t.Fatalf("ProgressBusy().Tag() = %q, want progress", vnode.Tag())
	}
	if vnode := ProgressBusyWithKey("manager.feedback", "loading", 28); vnode.Tag() != "progress" {
		t.Fatalf("ProgressBusyWithKey().Tag() = %q, want progress", vnode.Tag())
	} else if vnode.Key() != "manager.feedback" || vnode.Props()["width"] != 28 || vnode.Props()["indeterminate"] != true {
		t.Fatalf("ProgressBusyWithKey props = key:%q props:%+v", vnode.Key(), vnode.Props())
	}
	if vnode := ProgressComplete("Done"); vnode.Tag() != "progress" {
		t.Fatalf("ProgressComplete().Tag() = %q, want progress", vnode.Tag())
	}
	if vnode := ProgressFailed("Failed"); vnode.Tag() != "progress" {
		t.Fatalf("ProgressFailed().Tag() = %q, want progress", vnode.Tag())
	}
}

func TestOperationalFeedbackShortcut(t *testing.T) {
	busy := OperationalFeedback("manager.feedback", true, "captcha", "failed", "saved", 28)
	if busy.Tag() != "progress" {
		t.Fatalf("busy tag = %q, want progress", busy.Tag())
	}
	if busy.Key() != "manager.feedback" {
		t.Fatalf("busy key = %q, want manager.feedback", busy.Key())
	}
	if got := busy.Props()["label"]; got != "captcha" {
		t.Fatalf("busy label = %v, want captcha", got)
	}
	if got := busy.Props()["width"]; got != 28 {
		t.Fatalf("busy width = %v, want 28", got)
	}

	defaultBusy := OperationalFeedback("manager.feedback", true, "", "", "", 0)
	if got := defaultBusy.Props()["label"]; got != "loading" {
		t.Fatalf("default busy label = %v, want loading", got)
	}
	if got := defaultBusy.Props()["width"]; got != 28 {
		t.Fatalf("default busy width = %v, want 28", got)
	}

	errNode := OperationalFeedback("manager.feedback", false, "", "failed", "saved", 28)
	if errNode.Tag() != "alert" {
		t.Fatalf("error tag = %q, want alert", errNode.Tag())
	}
	if got := errNode.Props()["message"]; got != "failed" {
		t.Fatalf("error message = %v, want failed", got)
	}

	notice := OperationalFeedback("manager.feedback", false, "", "", "saved", 28)
	if notice.Tag() != "text" {
		t.Fatalf("notice tag = %q, want text", notice.Tag())
	}
	if got := notice.Props()["content"]; got != "saved" {
		t.Fatalf("notice content = %v, want saved", got)
	}

	empty := OperationalFeedback("manager.feedback", false, "", "", "", 28)
	if empty.Tag() != "text" {
		t.Fatalf("empty tag = %q, want text", empty.Tag())
	}
	if got := empty.Props()["content"]; got != "" {
		t.Fatalf("empty content = %v, want empty", got)
	}
}
