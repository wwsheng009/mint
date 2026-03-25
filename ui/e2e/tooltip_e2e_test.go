package e2e

import (
	"fmt"
	"testing"
	"time"

	"github.com/wwsheng009/mint/ui"
)

type tooltipFixtureMeta struct {
	HoverText    string
	FallbackText string
}

func newTooltipFixture() (ui.ComponentFunc, tooltipFixtureMeta) {
	meta := tooltipFixtureMeta{
		HoverText:    "Deploy current release",
		FallbackText: "Top edge tooltip fallback",
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
