package e2e

import (
	"testing"

	fwtheme "github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/ui"
	resultcomp "github.com/wwsheng009/mint/ui/components/result"
)

func newResultStaticApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Result E2E Fixture").Build(),
				resultcomp.NewBuilder().
					SetID("result-404").
					Status(resultcomp.Status404).
					Subtitle("The requested item was not found.").
					Build(),
				resultcomp.NewBuilder().
					SetID("result-success").
					Status(resultcomp.StatusSuccess).
					Title("Saved successfully").
					Subtitle("Your configuration has been applied.").
					Extra(ui.NewButtonBuilder("Back to list").SetID("result-back-btn").Build()).
					Bordered(true).
					Width(40).
					Build(),
			})
	}
}

func TestE2EResultPresetAndCustomContentRender(t *testing.T) {
	app, err := Run(newResultStaticApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("404")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("404 Not Found")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("The requested item was not found.")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Saved successfully")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Your configuration has been applied.")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByID("result-back-btn")); err != nil {
		t.Fatal(err)
	}
}

func TestE2EResultStatusStylesAndExtraRegionRender(t *testing.T) {
	app, err := Run(newResultStaticApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertStyle(ByText("404"), StyleExpect{
		HasFG:   true,
		FG:      fwtheme.Primary(),
		HasBold: true,
		Bold:    true,
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("Saved successfully"), StyleExpect{
		HasBold: true,
		Bold:    true,
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("Your configuration has been applied."), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Muted(),
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Back to list")); err != nil {
		t.Fatal(err)
	}
}
