package ui

import "testing"

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
}
