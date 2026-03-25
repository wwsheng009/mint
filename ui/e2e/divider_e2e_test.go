package e2e

import (
	"testing"

	fwtheme "github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
	dividercomp "github.com/wwsheng009/mint/ui/components/divider"
)

func newDividerStaticApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Divider E2E Fixture").Build(),
				dividercomp.NewBuilder().
					SetID("divider-solid-section").
					Double().
					FillWidth(false).
					Label("Section").
					Style(style.NewStyle().Foreground(fwtheme.Warning())).
					Build(),
				dividercomp.NewBuilder().
					SetID("divider-dashed").
					Dashed().
					FillWidth(false).
					Build(),
				dividercomp.NewBuilder().
					SetID("divider-dotted").
					Dotted().
					FillWidth(false).
					Build(),
			})
	}
}

func TestE2EDividerSolidDashedAndDottedRender(t *testing.T) {
	app, err := Run(newDividerStaticApp(), ui.WithSize(96, 20))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("Section")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("═")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("- -")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("· ")); err != nil {
		t.Fatal(err)
	}
}

func TestE2EDividerCustomStyleAppliesToLabel(t *testing.T) {
	app, err := Run(newDividerStaticApp(), ui.WithSize(96, 20))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertStyle(ByText("Section"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Warning(),
	}); err != nil {
		t.Fatal(err)
	}
}
