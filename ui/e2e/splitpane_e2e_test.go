package e2e

import (
	"testing"

	"github.com/wwsheng009/mint/ui"
	splitpanecomp "github.com/wwsheng009/mint/ui/components/splitpane"
)

func newSplitPaneStaticApp() ui.ComponentFunc {
	return func() ui.VNode {
		horizontal := splitpanecomp.NewBuilder().
			Key("ops-horizontal").
			Width(56).
			Height(4).
			PrimarySize(16).
			Gap(1).
			Panes(
				ui.NewVStack().
					SetGap(0).
					SetChildrenList([]ui.VNode{
						ui.NewTextBuilder("Groups").Build(),
						ui.NewTextBuilder("default").Build(),
					}),
				ui.NewVStack().
					SetGap(0).
					SetChildrenList([]ui.VNode{
						ui.NewTextBuilder("Provider Details").Build(),
						ui.NewTextBuilder("healthy").Build(),
					}),
			).
			Build()

		vertical := splitpanecomp.NewBuilder().
			Vertical().
			Key("ops-vertical").
			Width(32).
			PrimarySize(1).
			Gap(0).
			Panes(
				ui.NewTextBuilder("Runtime Summary").Build(),
				ui.NewTextBuilder("Reload required: no").Build(),
			).
			Build()

		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("SplitPane E2E Fixture").Build(),
				horizontal,
				vertical,
			})
	}
}

func TestE2ESplitPaneHorizontalPanesAndSeparatorRender(t *testing.T) {
	app, err := Run(newSplitPaneStaticApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("Groups")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Provider Details")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("│")); err != nil {
		t.Fatal(err)
	}

	groupsPoint, err := app.ResolvePoint(ByText("Groups"))
	if err != nil {
		t.Fatal(err)
	}
	detailsPoint, err := app.ResolvePoint(ByText("Provider Details"))
	if err != nil {
		t.Fatal(err)
	}
	if groupsPoint.X >= detailsPoint.X {
		t.Fatalf("expected primary pane left of secondary pane, got primaryX=%d secondaryX=%d", groupsPoint.X, detailsPoint.X)
	}
}

func TestE2ESplitPaneVerticalPanesAndSeparatorRender(t *testing.T) {
	app, err := Run(newSplitPaneStaticApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("Runtime Summary")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Reload required: no")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("────────────────────────────────")); err != nil {
		t.Fatal(err)
	}

	summaryPoint, err := app.ResolvePoint(ByText("Runtime Summary"))
	if err != nil {
		t.Fatal(err)
	}
	reloadPoint, err := app.ResolvePoint(ByText("Reload required: no"))
	if err != nil {
		t.Fatal(err)
	}
	if summaryPoint.Y >= reloadPoint.Y {
		t.Fatalf("expected primary pane above secondary pane, got primaryY=%d secondaryY=%d", summaryPoint.Y, reloadPoint.Y)
	}
}
