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
	if vnode := OperationElapsedTimer("Reload", time.Now().Add(-time.Minute)); vnode.Tag() != "timer" {
		t.Fatalf("OperationElapsedTimer().Tag() = %q, want timer", vnode.Tag())
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
	if vnode := ProgressComplete("Done"); vnode.Tag() != "progress" {
		t.Fatalf("ProgressComplete().Tag() = %q, want progress", vnode.Tag())
	}
	if vnode := ProgressFailed("Failed"); vnode.Tag() != "progress" {
		t.Fatalf("ProgressFailed().Tag() = %q, want progress", vnode.Tag())
	}
}
