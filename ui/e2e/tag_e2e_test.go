package e2e

import (
	"testing"

	fwtheme "github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
	tagcomp "github.com/wwsheng009/mint/ui/components/tag"
)

func newTagStaticApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Tag E2E Fixture").Build(),
				tagcomp.NewBuilder("Beta").
					Primary().
					Icon("*").
					Closable(true).
					Build(),
				tagcomp.NewBuilder("Stable").
					Success().
					Build(),
			})
	}
}

func TestE2ETagIconClosableSuffixAndTextRender(t *testing.T) {
	app, err := Run(newTagStaticApp(), ui.WithSize(96, 20))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("* Beta ×")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Stable")); err != nil {
		t.Fatal(err)
	}
}

func TestE2ETagVariantStylesRender(t *testing.T) {
	app, err := Run(newTagStaticApp(), ui.WithSize(96, 20))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertStyle(ByText("* Beta ×"), StyleExpect{
		HasBG:   true,
		BG:      style.Color("blue"),
		HasFG:   true,
		FG:      fwtheme.BG(),
		HasBold: true,
		Bold:    true,
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("Stable"), StyleExpect{
		HasBG:   true,
		BG:      style.Color("green"),
		HasFG:   true,
		FG:      fwtheme.BG(),
		HasBold: true,
		Bold:    true,
	}); err != nil {
		t.Fatal(err)
	}
}
