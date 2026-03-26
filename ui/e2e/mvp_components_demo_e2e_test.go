package e2e

import (
	"fmt"
	"testing"

	runtimeintent "github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/ui"
)

func newMVPComponentsDemoDateTimeFixture() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("MVP Components Demo Fixture").Build(),
				ui.HStack(
					ui.NewTextBuilder("Country:").Build(),
					ui.Text("  "),
					ui.NewButtonBuilder("Dummy").SetID("dummy-country").Build(),
				),
				ui.HStack(
					ui.NewTextBuilder("Ship Date:").Build(),
					ui.Text(" "),
					ui.NewDatePickerBuilder().
						SetID("profile.shipDate").
						ComponentID("profile.shipDate").
						Value("2026-04-05").
						ForField(runtimeintent.BindField("shipDate")).
						Width(18).
						Build(),
				),
				ui.HStack(
					ui.NewTextBuilder("Ship Time:").Build(),
					ui.Text(" "),
					ui.NewTimePickerBuilder().
						SetID("profile.shipTime").
						ComponentID("profile.shipTime").
						Value("09:30").
						ForField(runtimeintent.BindField("shipTime")).
						Width(10).
						Build(),
				),
				ui.HStack(
					ui.NewTextBuilder("Bio:").Build(),
					ui.Text("      "),
					ui.NewTextareaBuilder().
						SetID("profile.bio").
						Rows(4).
						Cols(38).
						Placeholder("Tell us about yourself...").
						Build(),
				),
			})
	}
}

func TestE2EMVPComponentsDemoDatePickerPopupAnchorsBelowTrigger(t *testing.T) {
	app, err := Run(newMVPComponentsDemoDateTimeFixture(), ui.WithSize(90, 35))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.Driver().Click(ByID("profile.shipDate")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByID("profile.shipDate-popup")); err != nil {
		t.Fatal(err)
	}

	assertPopupAnchoredBelowTrigger(t, app, "profile.shipDate", "profile.shipDate-popup")
}

func TestE2EMVPComponentsDemoTimePickerPopupAnchorsBelowTrigger(t *testing.T) {
	app, err := Run(newMVPComponentsDemoDateTimeFixture(), ui.WithSize(90, 35))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.Driver().Click(ByID("profile.shipTime")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByID("profile.shipTime-popup")); err != nil {
		t.Fatal(err)
	}

	assertPopupAnchoredBelowTrigger(t, app, "profile.shipTime", "profile.shipTime-popup")
}

func assertPopupAnchoredBelowTrigger(t *testing.T, app *App, triggerID, popupID string) {
	t.Helper()

	triggerBounds, err := app.BoundsOf(ByID(triggerID))
	if err != nil {
		t.Fatal(err)
	}
	popupBounds, err := app.BoundsOf(ByID(popupID))
	if err != nil {
		t.Fatal(err)
	}

	if triggerBounds.X == 0 && triggerBounds.Y == 0 {
		t.Fatalf("trigger %q stayed at origin; test cannot validate anchoring\n%s", triggerID, popupDiagnostics(app))
	}

	if popupBounds.X != triggerBounds.X {
		t.Fatalf("popup %q X = %d, want %d\n%s", popupID, popupBounds.X, triggerBounds.X, popupDiagnostics(app))
	}

	expectedY := triggerBounds.Y + triggerBounds.Height
	if popupBounds.Y != expectedY {
		t.Fatalf("popup %q Y = %d, want %d\n%s", popupID, popupBounds.Y, expectedY, popupDiagnostics(app))
	}
}

func popupDiagnostics(app *App) string {
	if app == nil || app.testApp == nil {
		return ""
	}
	root := app.testApp.GetDeclarativeRoot()
	if root == nil {
		return ""
	}
	return fmt.Sprintf("layout:\n%s\nportal:\n%s\npaintable:\n%s", root.GetLayoutTreeString(), root.GetPortalTreeString(), root.GetPaintableTreeString())
}
