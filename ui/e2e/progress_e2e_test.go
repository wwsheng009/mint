package e2e

import (
	"testing"

	fwtheme "github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/ui"
	progresscomp "github.com/wwsheng009/mint/ui/components/progress"
)

func newProgressStaticApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Progress E2E Fixture").Build(),
				progresscomp.NewBuilder().
					SetID("progress-success-line").
					Value(60).
					Max(100).
					Width(12).
					Label("Uploading").
					Success().
					Build(),
				progresscomp.NewBuilder().
					SetID("progress-active-line").
					Value(50).
					Max(100).
					Width(12).
					Label("Syncing").
					ShowPercent(false).
					Active().
					Build(),
				progresscomp.NewBuilder().
					SetID("progress-circle").
					Value(100).
					Max(100).
					Width(5).
					Circle().
					Build(),
				progresscomp.NewBuilder().
					SetID("progress-dashboard").
					Value(100).
					Max(100).
					Width(7).
					Label("CPU").
					ShowPercent(false).
					Dashboard().
					Build(),
			})
	}
}

func TestE2EProgressLineCircleAndDashboardRender(t *testing.T) {
	app, err := Run(newProgressStaticApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("[======----]")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Uploading: 60%")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Syncing")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText(" ### ")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("#   #")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("100%")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText(" ##### ")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("#     #")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("CPU")); err != nil {
		t.Fatal(err)
	}
}

func TestE2EProgressStatusStylesRender(t *testing.T) {
	app, err := Run(newProgressStaticApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertStyle(ByText("[======----]"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Success(),
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("Syncing"), StyleExpect{
		HasFG:   true,
		FG:      fwtheme.Focus(),
		HasBold: true,
		Bold:    true,
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("100%"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Primary(),
	}); err != nil {
		t.Fatal(err)
	}
}
