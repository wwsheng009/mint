package e2e

import (
	"fmt"
	"testing"
	"time"

	"github.com/wwsheng009/mint/ui"
	absolutecomp "github.com/wwsheng009/mint/ui/components/absolute"
)

type tooltipFixtureMeta struct {
	HoverText    string
	FallbackText string
	DelayedText  string
	RightText    string
	CornerText   string
}

func newTooltipFixture() (ui.ComponentFunc, tooltipFixtureMeta) {
	meta := tooltipFixtureMeta{
		HoverText:    "Deploy current release",
		FallbackText: "Top edge tooltip fallback",
		DelayedText:  "Delayed deployment details",
		RightText:    "Right edge tooltip fallback",
		CornerText:   "Corner right-top tooltip fallback",
	}

	appFn := func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Tooltip E2E Fixture").Build(),
				ui.NewTooltipBuilder(
					ui.NewButtonBuilder("Top edge anchor").SetID("tooltip-top-anchor").Build(),
					meta.FallbackText,
				).
					Top().
					Delay(0).
					Build(),
				ui.NewVStack().
					SetWidth(72).
					SetHeight(1).
					SetChildrenList([]ui.VNode{
						absolutecomp.NewBuilder(
							ui.NewTooltipBuilder(
								ui.NewButtonBuilder("Corner anchor").SetID("tooltip-corner-anchor").Build(),
								meta.CornerText,
							).
								RightTop().
								Delay(0).
								Build(),
						).
							Right(absolutecomp.AbsolutePos(0)).
							Top(absolutecomp.AbsolutePos(0)).
							Width(18).
							Height(1).
							Build(),
					}),
				ui.NewTextBuilder("Workspace: alpha").Build(),
				ui.NewTextBuilder("Branch: tooltip").Build(),
				ui.NewTextBuilder("Runtime: ready").Build(),
				ui.NewTooltipBuilder(
					ui.NewButtonBuilder("Deploy").SetID("tooltip-hover-anchor").Build(),
					meta.HoverText,
				).
					Auto().
					Delay(0).
					Build(),
				ui.NewTooltipBuilder(
					ui.NewButtonBuilder("Delayed Deploy").SetID("tooltip-delayed-anchor").Build(),
					meta.DelayedText,
				).
					Auto().
					Delay(120 * time.Millisecond).
					Build(),
				ui.NewVStack().
					SetWidth(72).
					SetHeight(1).
					SetChildrenList([]ui.VNode{
						absolutecomp.NewBuilder(
							ui.NewTooltipBuilder(
								ui.NewButtonBuilder("Right edge anchor").SetID("tooltip-right-anchor").Build(),
								meta.RightText,
							).
								Right().
								Delay(0).
								Build(),
						).
							Right(absolutecomp.AbsolutePos(0)).
							Top(absolutecomp.AbsolutePos(0)).
							Width(22).
							Height(1).
							Build(),
					}),
				ui.NewTextBuilder("Status: idle").Build(),
				ui.NewButtonBuilder("Neutral hover target").SetID("tooltip-neutral-target").Build(),
			})
	}

	return appFn, meta
}

func TestE2ETooltipHoverShowsAndHides(t *testing.T) {
	appFn, meta := newTooltipFixture()

	app, err := Run(appFn, ui.WithSize(72, 12))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.Driver().Move(ByID("tooltip-neutral-target")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText(meta.HoverText)); err == nil {
		t.Fatalf("tooltip %q should be hidden before hover", meta.HoverText)
	}

	if err := app.Driver().Move(ByID("tooltip-hover-anchor")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText(meta.HoverText))
	}); err != nil {
		t.Fatal(err)
	}

	if err := app.Driver().Move(ByID("tooltip-neutral-target")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText(meta.HoverText)); err == nil {
			return fmt.Errorf("tooltip %q still visible after hover exit", meta.HoverText)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestE2ETooltipTopPlacementFallsBackBelowViewportEdge(t *testing.T) {
	appFn, meta := newTooltipFixture()

	app, err := Run(appFn, ui.WithSize(72, 12))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	anchorBounds, err := app.BoundsOf(ByID("tooltip-top-anchor"))
	if err != nil {
		t.Fatal(err)
	}

	if err := app.Driver().Move(ByID("tooltip-top-anchor")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText(meta.FallbackText))
	}); err != nil {
		t.Fatal(err)
	}

	tooltipPoint, err := app.ResolvePoint(ByText(meta.FallbackText))
	if err != nil {
		t.Fatal(err)
	}
	if tooltipPoint.Y <= anchorBounds.Y {
		t.Fatalf("tooltip row = %d, want below anchor row %d after top-edge fallback", tooltipPoint.Y, anchorBounds.Y)
	}
	if tooltipPoint.X < 0 || tooltipPoint.X >= 72 {
		t.Fatalf("tooltip point x = %d, want within viewport", tooltipPoint.X)
	}
}

func TestE2ETooltipDelayWaitsBeforeShowing(t *testing.T) {
	appFn, meta := newTooltipFixture()

	app, err := Run(appFn, ui.WithSize(72, 14))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.Driver().Move(ByID("tooltip-delayed-anchor")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText(meta.DelayedText)); err == nil {
		t.Fatalf("tooltip %q should stay hidden immediately after hover", meta.DelayedText)
	}

	time.Sleep(30 * time.Millisecond)
	if err := app.AssertVisible(ByText(meta.DelayedText)); err == nil {
		t.Fatalf("tooltip %q should still be hidden before delay elapses", meta.DelayedText)
	}

	if err := app.Eventually(700*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText(meta.DelayedText))
	}); err != nil {
		t.Fatal(err)
	}
}

func TestE2ETooltipRightPlacementFallsBackLeftOfViewportEdge(t *testing.T) {
	appFn, meta := newTooltipFixture()

	app, err := Run(appFn, ui.WithSize(72, 14))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	anchorBounds, err := app.BoundsOf(ByID("tooltip-right-anchor"))
	if err != nil {
		t.Fatal(err)
	}

	if err := app.Driver().Move(ByID("tooltip-right-anchor")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText(meta.RightText))
	}); err != nil {
		t.Fatal(err)
	}

	tooltipPoint, err := app.ResolvePoint(ByText(meta.RightText))
	if err != nil {
		t.Fatal(err)
	}
	if tooltipPoint.X >= anchorBounds.X {
		t.Fatalf("tooltip column = %d, want left of anchor column %d after right-edge fallback", tooltipPoint.X, anchorBounds.X)
	}
	if tooltipPoint.X < 0 || tooltipPoint.X >= 72 {
		t.Fatalf("tooltip point x = %d, want within viewport", tooltipPoint.X)
	}
}

func TestE2ETooltipRightTopPlacementFallsBackToMirroredLeftTopNearCorner(t *testing.T) {
	appFn, meta := newTooltipFixture()

	app, err := Run(appFn, ui.WithSize(72, 14))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	anchorBounds, err := app.BoundsOf(ByID("tooltip-corner-anchor"))
	if err != nil {
		t.Fatal(err)
	}

	if err := app.Driver().Move(ByID("tooltip-corner-anchor")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText(meta.CornerText))
	}); err != nil {
		t.Fatal(err)
	}

	tooltipPoint, err := app.ResolvePoint(ByText(meta.CornerText))
	if err != nil {
		t.Fatal(err)
	}
	if tooltipPoint.X >= anchorBounds.X {
		t.Fatalf("tooltip column = %d, want left of corner anchor column %d after mirrored fallback", tooltipPoint.X, anchorBounds.X)
	}
	if tooltipPoint.Y != anchorBounds.Y {
		t.Fatalf("tooltip row = %d, want same row as corner anchor row %d after mirrored top fallback", tooltipPoint.Y, anchorBounds.Y)
	}
}
