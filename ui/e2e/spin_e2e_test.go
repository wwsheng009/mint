package e2e

import (
	"testing"

	fwtheme "github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
	spincomp "github.com/wwsheng009/mint/ui/components/spin"
)

func newSpinStaticApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Spin E2E Fixture").Build(),
				spincomp.NewBuilder().
					Large().
					Tip("Loading data").
					Build(),
				spincomp.NewBuilder().
					Small().
					Tip("Queued").
					Build(),
				spincomp.NewBuilder().
					Tip("Hidden spinner").
					Spinning(false).
					Build(),
			})
	}
}

func TestE2ESpinVisibleTipsAndHiddenStoppedSpinnerRender(t *testing.T) {
	app, err := Run(newSpinStaticApp(), ui.WithSize(96, 20))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("Loading data")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Queued")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Hidden spinner")); err == nil {
		t.Fatal("stopped spinner tip should not be rendered")
	}
}

func TestE2ESpinStyleRender(t *testing.T) {
	app, err := Run(newSpinStaticApp(), ui.WithSize(96, 20))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertStyle(ByText("Loading data"), StyleExpect{
		HasBG:   true,
		BG:      style.Color("cyan"),
		HasFG:   true,
		FG:      fwtheme.BG(),
		HasBold: true,
		Bold:    true,
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("Queued"), StyleExpect{
		HasBG:   true,
		BG:      style.Color("cyan"),
		HasFG:   true,
		FG:      fwtheme.BG(),
		HasBold: true,
		Bold:    true,
	}); err != nil {
		t.Fatal(err)
	}
}
