package e2e

import (
	"testing"

	fwtheme "github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/ui"
	timelinecomp "github.com/wwsheng009/mint/ui/components/timeline"
)

func newTimelineStaticApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Timeline E2E Fixture").Build(),
				timelinecomp.NewBuilder().
					SetID("timeline-main").
					Item(
						timelinecomp.Event("Build completed").
							WithLabel("09:30").
							WithDescription("CI pipeline finished").
							WithStatus(timelinecomp.StatusSuccess),
					).
					Item(
						timelinecomp.Event("Deploy started").
							WithLabel("09:45").
							WithStatus(timelinecomp.StatusWarning),
					).
					Pending("Waiting for smoke tests").
					Build(),
				timelinecomp.NewBuilder().
					SetID("timeline-reverse").
					Reverse(true).
					Item(timelinecomp.Event("Reverse oldest").WithLabel("L1")).
					Item(timelinecomp.Event("Reverse newest").WithLabel("L2")).
					Build(),
			})
	}
}

func TestE2ETimelineItemsDescriptionsPendingAndMarkersRender(t *testing.T) {
	app, err := Run(newTimelineStaticApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("09:30")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Build completed")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("CI pipeline finished")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("09:45")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Deploy started")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Pending")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Waiting for smoke tests")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("▲")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("○")); err != nil {
		t.Fatal(err)
	}
}

func TestE2ETimelineStylesAndReverseOrderRender(t *testing.T) {
	app, err := Run(newTimelineStaticApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertStyle(ByText("09:30"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Muted(),
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("▲"), StyleExpect{
		HasFG:   true,
		FG:      fwtheme.Warning(),
		HasBold: true,
		Bold:    true,
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("○"), StyleExpect{
		HasFG:   true,
		FG:      fwtheme.Muted(),
		HasBold: true,
		Bold:    true,
	}); err != nil {
		t.Fatal(err)
	}

	newestPoint, err := app.ResolvePoint(ByText("Reverse newest"))
	if err != nil {
		t.Fatal(err)
	}
	oldestPoint, err := app.ResolvePoint(ByText("Reverse oldest"))
	if err != nil {
		t.Fatal(err)
	}
	if newestPoint.Y >= oldestPoint.Y {
		t.Fatalf("expected reverse newest above reverse oldest, got newestY=%d oldestY=%d", newestPoint.Y, oldestPoint.Y)
	}
}
