package e2e

import (
	"testing"

	fwtheme "github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
	alertcomp "github.com/wwsheng009/mint/ui/components/alert"
)

func newAlertStaticApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Alert E2E Fixture").Build(),
				alertcomp.NewBuilder("Disk usage is high").
					Title("Warning").
					Warning().
					Closable(true).
					Build(),
				alertcomp.NewBuilder("Configuration saved").
					Success().
					Build(),
			})
	}
}

func TestE2EAlertTitleMessageAndCloseHintRender(t *testing.T) {
	app, err := Run(newAlertStaticApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("Warning")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Disk usage is high")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("[press x to close]")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Configuration saved")); err != nil {
		t.Fatal(err)
	}
}

func TestE2EAlertSeverityStylesRender(t *testing.T) {
	app, err := Run(newAlertStaticApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertStyle(ByText("Warning"), StyleExpect{
		HasBG:   true,
		BG:      style.Color("yellow"),
		HasFG:   true,
		FG:      fwtheme.BG(),
		HasBold: true,
		Bold:    true,
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("Configuration saved"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Foreground(),
	}); err != nil {
		t.Fatal(err)
	}
}
