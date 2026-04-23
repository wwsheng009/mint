package e2e

import (
	"testing"

	"github.com/wwsheng009/mint/ui"
	panelcomp "github.com/wwsheng009/mint/ui/components/panel"
)

func newPanelStaticApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Panel E2E Fixture").Build(),
				panelcomp.NewBuilder().
					SetID("profile-panel").
					Title("Profile").
					Header(ui.NewTextBuilder("Header").Build()).
					Content(ui.NewTextBuilder("Body").Build()).
					Footer(ui.NewTextBuilder("Footer").Build()).
					Width(28).
					Rounded().
					Build(),
				panelcomp.NewBuilder().
					SetID("plain-panel").
					Content(ui.NewTextBuilder("Plain Body").Build()).
					NoBorder().
					Build(),
			})
	}
}

func TestE2EPanelBorderedTitleContentAndFooterRender(t *testing.T) {
	app, err := Run(newPanelStaticApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("Profile")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Header")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Body")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Footer")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("╭")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("╯")); err != nil {
		t.Fatal(err)
	}
}

func TestE2EPanelNoBorderStillRendersContent(t *testing.T) {
	app, err := Run(newPanelStaticApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("Plain Body")); err != nil {
		t.Fatal(err)
	}
}
