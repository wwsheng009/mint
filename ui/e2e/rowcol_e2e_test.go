package e2e

import (
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
	rowcolcomp "github.com/wwsheng009/mint/ui/components/rowcol"
)

func newRowColStaticApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("RowCol E2E Fixture").Build(),
				rowcolcomp.NewRowBuilder().
					SetID("row-wrap").
					Gutter(2, 1).
					Width(48).
					Children(
						rowcolcomp.NewColBuilder().Span(12).Children(ui.NewTextBuilder("Alpha").Build()).Build(),
						rowcolcomp.NewColBuilder().Span(12).Children(ui.NewTextBuilder("Beta").Build()).Build(),
						rowcolcomp.NewColBuilder().Span(12).Children(ui.NewTextBuilder("Gamma").Build()).Build(),
					).
					Build(),
				rowcolcomp.NewRowBuilder().
					SetID("row-offset").
					Justify(rtui.AlignEnd).
					Width(48).
					Children(
						rowcolcomp.NewColBuilder().Span(6).Offset(6).Children(ui.NewTextBuilder("Main").Build()).Build(),
					).
					Build(),
				rowcolcomp.NewColBuilder().
					SetID("col-stack").
					Children(
						ui.NewTextBuilder("Top").Build(),
						ui.NewTextBuilder("Bottom").Build(),
					).
					Build(),
			})
	}
}

func TestE2ERowColRowWrapRender(t *testing.T) {
	app, err := Run(newRowColStaticApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	alphaPoint, err := app.ResolvePoint(ByText("Alpha"))
	if err != nil {
		t.Fatal(err)
	}
	betaPoint, err := app.ResolvePoint(ByText("Beta"))
	if err != nil {
		t.Fatal(err)
	}
	gammaPoint, err := app.ResolvePoint(ByText("Gamma"))
	if err != nil {
		t.Fatal(err)
	}

	if alphaPoint.Y != betaPoint.Y {
		t.Fatalf("expected Alpha and Beta on same row, got y=%d and y=%d", alphaPoint.Y, betaPoint.Y)
	}
	if gammaPoint.Y <= alphaPoint.Y {
		t.Fatalf("expected Gamma below first row, got gammaY=%d firstRowY=%d", gammaPoint.Y, alphaPoint.Y)
	}
	if betaPoint.X <= alphaPoint.X {
		t.Fatalf("expected Beta to the right of Alpha, got alphaX=%d betaX=%d", alphaPoint.X, betaPoint.X)
	}
}

func TestE2ERowColOffsetAndColStackRender(t *testing.T) {
	app, err := Run(newRowColStaticApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("Main")); err != nil {
		t.Fatal(err)
	}

	topPoint, err := app.ResolvePoint(ByText("Top"))
	if err != nil {
		t.Fatal(err)
	}
	bottomPoint, err := app.ResolvePoint(ByText("Bottom"))
	if err != nil {
		t.Fatal(err)
	}
	if bottomPoint.Y <= topPoint.Y {
		t.Fatalf("expected Bottom below Top, got topY=%d bottomY=%d", topPoint.Y, bottomPoint.Y)
	}
}
