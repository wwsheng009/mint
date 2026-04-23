package e2e

import (
	"testing"
	"time"

	fwtheme "github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
	clockcomp "github.com/wwsheng009/mint/ui/components/clock"
)

func newClockStaticApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Clock E2E Fixture").Build(),
				clockcomp.NewBuilder().
					Radius(4).
					StaticTime(time.Date(2026, 3, 29, 3, 0, 0, 0, time.UTC)).
					HideSeconds().
					SetID("clock-static").
					Build(),
				clockcomp.NewBuilder().
					Radius(4).
					StaticTime(time.Date(2026, 3, 29, 9, 15, 30, 0, time.UTC)).
					SetID("clock-styled").
					Style(style.Style{}.Foreground(style.Yellow).Bold(true)).
					Build(),
			})
	}
}

func TestE2EClockStaticRender(t *testing.T) {
	app, err := Run(newClockStaticApp(), ui.WithSize(96, 28))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("Clock E2E Fixture")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("03:00")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("09:15:30")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("@---O")); err != nil {
		t.Fatal(err)
	}
}

func TestE2EClockStyleRender(t *testing.T) {
	app, err := Run(newClockStaticApp(), ui.WithSize(96, 28))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertStyle(ByText("09:15:30"), StyleExpect{
		HasFG:   true,
		FG:      style.Yellow,
		HasBold: true,
		Bold:    true,
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("03:00"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Primary(),
	}); err != nil {
		t.Fatal(err)
	}
}
