package e2e

import (
	"testing"
	"time"

	fwtheme "github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/ui"
	timercomp "github.com/wwsheng009/mint/ui/components/timer"
)

func newTimerStaticApp() ui.ComponentFunc {
	return func() ui.VNode {
		startedAt := time.Date(2026, 5, 25, 8, 0, 0, 0, time.UTC)
		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Timer E2E Fixture").Build(),
				timercomp.NewBuilder().
					SetID("timer-countdown").
					Label("Refresh").
					Countdown(time.Minute).
					StartedAt(startedAt).
					Now(startedAt.Add(15 * time.Second)).
					Static().
					ShowProgress(true).
					ProgressWidth(12).
					Build(),
				timercomp.NewBuilder().
					SetID("timer-elapsed").
					Label("Uptime").
					Elapsed().
					StartedAt(startedAt).
					Now(startedAt.Add(90 * time.Second)).
					Static().
					Build(),
				ui.NewTimerBuilder().
					SetID("timer-expired").
					Label("Retry").
					Until(startedAt.Add(10 * time.Second)).
					StartedAt(startedAt).
					Now(startedAt.Add(12 * time.Second)).
					Static().
					ExpiredText("ready").
					Build(),
			})
	}
}

func TestE2ETimerStaticRender(t *testing.T) {
	app, err := Run(newTimerStaticApp(), ui.WithSize(80, 16))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{
		"Timer E2E Fixture",
		"Refresh: 00:45 [██░░░░░░░░]",
		"Uptime: 01:30",
		"Retry: ready",
	} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestE2ETimerSemanticStyles(t *testing.T) {
	app, err := Run(newTimerStaticApp(), ui.WithSize(80, 16))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertStyle(ByText("Refresh: 00:45 [██░░░░░░░░]"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Primary(),
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("Retry: ready"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Error(),
	}); err != nil {
		t.Fatal(err)
	}
}
