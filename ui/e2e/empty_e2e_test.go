package e2e

import (
	"testing"

	fwtheme "github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/ui"
	emptycomp "github.com/wwsheng009/mint/ui/components/empty"
)

func newEmptyStaticApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Empty E2E Fixture").Build(),
				emptycomp.NewBuilder().Build(),
				emptycomp.NewBuilder().
					Description("No records found").
					Image("[ ]").
					Build(),
			})
	}
}

func TestE2EEmptyDefaultAndCustomRender(t *testing.T) {
	app, err := Run(newEmptyStaticApp(), ui.WithSize(96, 20))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("( ∅ )")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("No Data")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("[ ]")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("No records found")); err != nil {
		t.Fatal(err)
	}
}

func TestE2EEmptyDefaultUsesMutedStyle(t *testing.T) {
	app, err := Run(newEmptyStaticApp(), ui.WithSize(96, 20))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertStyle(ByText("( ∅ )"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Muted(),
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("No Data"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Muted(),
	}); err != nil {
		t.Fatal(err)
	}
}
