package e2e

import (
	"testing"

	"github.com/wwsheng009/mint/ui"
	wrapcomp "github.com/wwsheng009/mint/ui/components/wrap"
)

func newWrapStaticApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Wrap E2E Fixture").Build(),
				wrapcomp.NewBuilder().
					SetID("wrap-main").
					Width(8).
					Gap(1).
					SingleBorder().
					Children(
						ui.NewTextBuilder("One").Build(),
						ui.NewTextBuilder("Two").Build(),
						ui.NewTextBuilder("Tri").Build(),
					).
					Build(),
			})
	}
}

func TestE2EWrapBorderLabelAndWrappedChildrenRender(t *testing.T) {
	app, err := Run(newWrapStaticApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("One")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Two")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Tri")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("┌")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("┘")); err != nil {
		t.Fatal(err)
	}
}

func TestE2EWrapChildrenFlowOntoMultipleRows(t *testing.T) {
	app, err := Run(newWrapStaticApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	onePoint, err := app.ResolvePoint(ByText("One"))
	if err != nil {
		t.Fatal(err)
	}
	twoPoint, err := app.ResolvePoint(ByText("Two"))
	if err != nil {
		t.Fatal(err)
	}
	triPoint, err := app.ResolvePoint(ByText("Tri"))
	if err != nil {
		t.Fatal(err)
	}

	if onePoint.Y != twoPoint.Y {
		t.Fatalf("expected One and Two on same row, got y=%d and y=%d", onePoint.Y, twoPoint.Y)
	}
	if triPoint.Y <= twoPoint.Y {
		t.Fatalf("expected Tri on a later row, got triY=%d twoY=%d", triPoint.Y, twoPoint.Y)
	}
}
