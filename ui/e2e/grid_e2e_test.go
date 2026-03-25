package e2e

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
	gridcomp "github.com/wwsheng009/mint/ui/components/grid"
)

func newGridStaticApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Grid E2E Fixture").Build(),
				gridcomp.New().
					SetColumns(gridcomp.Fixed(4), gridcomp.Fixed(4)).
					SetRows(gridcomp.Fixed(1), gridcomp.Fixed(1)).
					SetChildrenAuto([]ui.VNode{
						ui.NewTextBuilder("A1").Build(),
						ui.NewTextBuilder("A2").Build(),
						ui.NewTextBuilder("B1").Build(),
						ui.NewTextBuilder("B2").Build(),
					}).
					SingleCellBorders().
					SetCellBorderColor("cyan"),
				gridcomp.New().
					SetColumns(gridcomp.Fixed(4), gridcomp.Fixed(4)).
					SetRows(gridcomp.Fixed(1), gridcomp.Fixed(1)).
					SetChildrenAuto([]ui.VNode{
						ui.NewTextBuilder("R1").Build(),
						ui.NewTextBuilder("R2").Build(),
						ui.NewTextBuilder("R3").Build(),
						ui.NewTextBuilder("R4").Build(),
					}).
					RoundedCellBorders(),
			})
	}
}

func TestE2EGridAutoPositionAndCellBordersRender(t *testing.T) {
	app, err := Run(newGridStaticApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("A1")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("A2")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("B1")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("B2")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("┌")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("┬")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("┘")); err != nil {
		t.Fatal(err)
	}

	a1Point, err := app.ResolvePoint(ByText("A1"))
	if err != nil {
		t.Fatal(err)
	}
	a2Point, err := app.ResolvePoint(ByText("A2"))
	if err != nil {
		t.Fatal(err)
	}
	b1Point, err := app.ResolvePoint(ByText("B1"))
	if err != nil {
		t.Fatal(err)
	}

	if a1Point.Y != a2Point.Y {
		t.Fatalf("expected A1 and A2 on same row, got y=%d and y=%d", a1Point.Y, a2Point.Y)
	}
	if b1Point.Y <= a1Point.Y {
		t.Fatalf("expected B1 below A1, got B1 y=%d A1 y=%d", b1Point.Y, a1Point.Y)
	}
}

func TestE2EGridRoundedBordersAndBorderColorRender(t *testing.T) {
	app, err := Run(newGridStaticApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("╭")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("╯")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("R1")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("R4")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("┌"), StyleExpect{
		HasFG: true,
		FG:    style.Color("cyan"),
	}); err != nil {
		t.Fatal(err)
	}
}
