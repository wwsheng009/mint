package e2e

import (
	"testing"

	"github.com/wwsheng009/mint/ui"
	spacecomp "github.com/wwsheng009/mint/ui/components/space"
)

func newSpaceStaticApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Space E2E Fixture").Build(),
				spacecomp.NewBuilder().
					SetID("space-horizontal").
					Large().
					Split("|").
					Children(
						ui.NewTextBuilder("Build").Build(),
						ui.NewTextBuilder("Test").Build(),
						ui.NewTextBuilder("Deploy").Build(),
					).
					Build(),
				spacecomp.NewBuilder().
					SetID("space-vertical").
					Vertical().
					Split("/").
					Children(
						ui.NewTextBuilder("North").Build(),
						ui.NewTextBuilder("South").Build(),
					).
					Build(),
			})
	}
}

func TestE2ESpaceHorizontalSplitRender(t *testing.T) {
	app, err := Run(newSpaceStaticApp(), ui.WithSize(96, 20))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("Build")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Test")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Deploy")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("|")); err != nil {
		t.Fatal(err)
	}

	buildPoint, err := app.ResolvePoint(ByText("Build"))
	if err != nil {
		t.Fatal(err)
	}
	testPoint, err := app.ResolvePoint(ByText("Test"))
	if err != nil {
		t.Fatal(err)
	}
	deployPoint, err := app.ResolvePoint(ByText("Deploy"))
	if err != nil {
		t.Fatal(err)
	}
	if buildPoint.Y != testPoint.Y || testPoint.Y != deployPoint.Y {
		t.Fatalf("expected horizontal space children on one row, got buildY=%d testY=%d deployY=%d", buildPoint.Y, testPoint.Y, deployPoint.Y)
	}
}

func TestE2ESpaceVerticalSplitRender(t *testing.T) {
	app, err := Run(newSpaceStaticApp(), ui.WithSize(96, 20))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("North")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("South")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("/")); err != nil {
		t.Fatal(err)
	}

	northPoint, err := app.ResolvePoint(ByText("North"))
	if err != nil {
		t.Fatal(err)
	}
	southPoint, err := app.ResolvePoint(ByText("South"))
	if err != nil {
		t.Fatal(err)
	}
	if southPoint.Y <= northPoint.Y {
		t.Fatalf("expected South below North, got northY=%d southY=%d", northPoint.Y, southPoint.Y)
	}
}
