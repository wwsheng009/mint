package timer

import (
	"testing"
	"time"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

func TestVNodeDefaults(t *testing.T) {
	v := New()
	if v == nil {
		t.Fatal("New returned nil")
	}
	if v.Tag() != "timer" {
		t.Fatalf("Tag = %q, want timer", v.Tag())
	}
	if v.Mode() != ModeElapsed {
		t.Fatalf("Mode = %v, want elapsed", v.Mode())
	}
	if !v.Live() {
		t.Fatal("Timer should be live by default")
	}
	if v.ProgressWidth() != 12 {
		t.Fatalf("ProgressWidth = %d, want 12", v.ProgressWidth())
	}
}

func TestBuilderCountdown(t *testing.T) {
	startedAt := time.Date(2026, 5, 25, 8, 0, 0, 0, time.UTC)
	v := NewBuilder().
		Key("refresh").
		Label("Refresh").
		Countdown(90 * time.Second).
		StartedAt(startedAt).
		Now(startedAt.Add(30 * time.Second)).
		Static().
		ShowProgress(true).
		ProgressWidth(10).
		WarningBelow(5 * time.Second).
		BuildTyped()

	if v.Key() != "refresh" {
		t.Fatalf("Key = %q, want refresh", v.Key())
	}
	if v.Label() != "Refresh" {
		t.Fatalf("Label = %q, want Refresh", v.Label())
	}
	if v.Mode() != ModeCountdown {
		t.Fatalf("Mode = %v, want countdown", v.Mode())
	}
	if v.Duration() != 90*time.Second {
		t.Fatalf("Duration = %s, want 1m30s", v.Duration())
	}
	if v.Live() {
		t.Fatal("Static should disable live ticking")
	}
	if !v.ShowProgress() || v.ProgressWidth() != 10 {
		t.Fatalf("progress config = show:%v width:%d", v.ShowProgress(), v.ProgressWidth())
	}
}

func TestInstancePaintCountdownWithProgress(t *testing.T) {
	startedAt := time.Date(2026, 5, 25, 8, 0, 0, 0, time.UTC)
	inst := NewInstance(rtui.Props{
		propMode:          ModeCountdown,
		propLabel:         "Refresh",
		propDuration:      time.Minute,
		propStartedAt:     startedAt,
		propNow:           startedAt.Add(15 * time.Second),
		propShowProgress:  true,
		propProgressWidth: 12,
		propLive:          false,
	})

	cmds := inst.Paint(0, 0)
	if len(cmds) != 1 {
		t.Fatalf("Paint command count = %d, want 1", len(cmds))
	}
	if cmds[0].Text != "Refresh: 00:45 [##--------]" {
		t.Fatalf("Text = %q", cmds[0].Text)
	}
	if got := cmds[0].Style.FG; got != theme.Primary() {
		t.Fatalf("FG = %q, want primary", got)
	}
}

func TestInstancePaintElapsed(t *testing.T) {
	startedAt := time.Date(2026, 5, 25, 8, 0, 0, 0, time.UTC)
	inst := NewInstance(rtui.Props{
		propMode:      ModeElapsed,
		propLabel:     "Uptime",
		propStartedAt: startedAt,
		propNow:       startedAt.Add(90 * time.Second),
		propLive:      false,
	})

	cmds := inst.Paint(0, 0)
	if cmds[0].Text != "Uptime: 01:30" {
		t.Fatalf("Text = %q, want Uptime: 01:30", cmds[0].Text)
	}
}

func TestInstancePaintExpiredAndWarningStyles(t *testing.T) {
	startedAt := time.Date(2026, 5, 25, 8, 0, 0, 0, time.UTC)
	expired := NewInstance(rtui.Props{
		propMode:        ModeCountdown,
		propLabel:       "Retry",
		propDuration:    time.Second,
		propStartedAt:   startedAt,
		propNow:         startedAt.Add(2 * time.Second),
		propExpiredText: "ready",
		propLive:        false,
	})
	if expired.Paint(0, 0)[0].Text != "Retry: ready" {
		t.Fatalf("expired text = %q", expired.Paint(0, 0)[0].Text)
	}
	if got := expired.Paint(0, 0)[0].Style.FG; got != theme.Error() {
		t.Fatalf("expired FG = %q, want error", got)
	}

	warning := NewInstance(rtui.Props{
		propMode:         ModeCountdown,
		propDuration:     20 * time.Second,
		propStartedAt:    startedAt,
		propNow:          startedAt.Add(12 * time.Second),
		propWarningBelow: 10 * time.Second,
		propLive:         false,
	})
	if got := warning.Paint(0, 0)[0].Style.FG; got != theme.Warning() {
		t.Fatalf("warning FG = %q, want warning", got)
	}
}

func TestInstanceTickUpdatesDisplayedTime(t *testing.T) {
	startedAt := time.Date(2026, 5, 25, 8, 0, 0, 0, time.UTC)
	inst := NewInstance(rtui.Props{
		propMode:      ModeElapsed,
		propStartedAt: startedAt,
		propNow:       startedAt,
		propLive:      true,
	})
	inst.dirty = false

	if changed := inst.Tick(startedAt.Add(500 * time.Millisecond)); changed {
		t.Fatal("sub-second tick should not update timer")
	}
	if changed := inst.Tick(startedAt.Add(time.Second)); !changed {
		t.Fatal("one-second tick should update timer")
	}
	if inst.Paint(0, 0)[0].Text != "00:01" {
		t.Fatalf("Text after tick = %q, want 00:01", inst.Paint(0, 0)[0].Text)
	}
	if !inst.IsDirty() {
		t.Fatal("Tick should mark dirty when text changes")
	}
}

func TestInstanceDoesNotTickWhenStaticOrExpired(t *testing.T) {
	startedAt := time.Date(2026, 5, 25, 8, 0, 0, 0, time.UTC)
	static := NewInstance(rtui.Props{
		propLive: false,
	})
	if static.WantsTick() {
		t.Fatal("static timer should not tick")
	}

	expired := NewInstance(rtui.Props{
		propMode:      ModeCountdown,
		propDuration:  time.Second,
		propStartedAt: startedAt,
		propNow:       startedAt.Add(2 * time.Second),
		propLive:      true,
	})
	if expired.WantsTick() {
		t.Fatal("expired countdown should stop ticking")
	}
}

func TestInstanceMeasureAndWidthFit(t *testing.T) {
	startedAt := time.Date(2026, 5, 25, 8, 0, 0, 0, time.UTC)
	inst := NewInstance(rtui.Props{
		propLabel:     "Long refresh timer",
		propStartedAt: startedAt,
		propNow:       startedAt.Add(time.Minute),
		propWidth:     12,
		propLive:      false,
	})

	size := inst.Measure(layout.UnboundedConstraints())
	if size.Width != 12 || size.Height != 1 {
		t.Fatalf("Measure = %+v, want 12x1", size)
	}
	text := inst.Paint(0, 0)[0].Text
	if paint.StringWidth(text) != 12 {
		t.Fatalf("paint width = %d, want 12: %q", paint.StringWidth(text), text)
	}
	if text != "Long refr..." {
		t.Fatalf("fitted text = %q, want Long refr...", text)
	}
}

func TestExplicitStylesOverrideSemanticDefaults(t *testing.T) {
	startedAt := time.Date(2026, 5, 25, 8, 0, 0, 0, time.UTC)
	inst := NewInstance(rtui.Props{
		propMode:         ModeCountdown,
		propDuration:     20 * time.Second,
		propStartedAt:    startedAt,
		propNow:          startedAt.Add(12 * time.Second),
		propWarningBelow: 10 * time.Second,
		propWarningStyle: style.Style{}.Foreground(style.Magenta).Bold(true),
		propLive:         false,
	})

	cmd := inst.Paint(0, 0)[0]
	if cmd.Style.FG != style.Magenta || !cmd.Style.IsBold() {
		t.Fatalf("style = %+v, want magenta bold", cmd.Style)
	}
}
