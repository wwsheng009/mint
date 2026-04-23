package e2e

import (
	"testing"

	fwtheme "github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/ui"
	descriptionscomp "github.com/wwsheng009/mint/ui/components/descriptions"
)

func newDescriptionsStaticApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Descriptions E2E Fixture").Build(),
				descriptionscomp.NewBuilder().
					SetID("descriptions-horizontal").
					Title("Build Info").
					Extra(ui.NewTextBuilder("readonly").Build()).
					Column(2).
					Bordered(true).
					Item(descriptionscomp.Field("Version", "v1.2.3")).
					Item(descriptionscomp.Field("Commit", "308cc4b5")).
					Build(),
				descriptionscomp.NewBuilder().
					SetID("descriptions-vertical").
					Vertical().
					Colon(false).
					Item(descriptionscomp.Field("Status", "Ready")).
					Item(descriptionscomp.Field("Region", "ap-south")).
					Build(),
			})
	}
}

func TestE2EDescriptionsHeaderBorderAndHorizontalFieldsRender(t *testing.T) {
	app, err := Run(newDescriptionsStaticApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("Build Info")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("readonly")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Version:")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("v1.2.3")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Commit:")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("308cc4b5")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("┌")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("┘")); err != nil {
		t.Fatal(err)
	}
}

func TestE2EDescriptionsVerticalLayoutOmitsColonAndUsesLabelStyle(t *testing.T) {
	app, err := Run(newDescriptionsStaticApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("Status")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Ready")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Region")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("ap-south")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Status:")); err == nil {
		t.Fatal("vertical descriptions should not render label colon")
	}
	if err := app.AssertStyle(ByText("Status"), StyleExpect{
		HasFG:   true,
		FG:      fwtheme.Muted(),
		HasBold: true,
		Bold:    true,
	}); err != nil {
		t.Fatal(err)
	}
}
