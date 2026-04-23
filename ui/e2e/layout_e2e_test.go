package e2e

import (
	"testing"

	"github.com/wwsheng009/mint/ui"
	layoutcomp "github.com/wwsheng009/mint/ui/components/layout"
)

func newLayoutStaticApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Layout E2E Fixture").Build(),
				layoutcomp.NewBuilder().
					SetID("layout-full").
					Header(ui.NewTextBuilder("Shell Header").Build()).
					LeftSider(ui.NewTextBuilder("Left Nav").Build()).
					Content(ui.NewTextBuilder("Main Content").Build()).
					RightSider(ui.NewTextBuilder("Right Tools").Build()).
					Footer(ui.NewTextBuilder("Shell Footer").Build()).
					Gap(1).
					BodyGap(2).
					Width(60).
					Height(12).
					Build(),
				layoutcomp.NewBuilder().
					SetID("layout-content-only").
					Content(ui.NewTextBuilder("Only Content").Build()).
					Build(),
			})
	}
}

func TestE2ELayoutShellRegionsRenderInExpectedOrder(t *testing.T) {
	app, err := Run(newLayoutStaticApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	headerPoint, err := app.ResolvePoint(ByText("Shell Header"))
	if err != nil {
		t.Fatal(err)
	}
	leftPoint, err := app.ResolvePoint(ByText("Left Nav"))
	if err != nil {
		t.Fatal(err)
	}
	contentPoint, err := app.ResolvePoint(ByText("Main Content"))
	if err != nil {
		t.Fatal(err)
	}
	rightPoint, err := app.ResolvePoint(ByText("Right Tools"))
	if err != nil {
		t.Fatal(err)
	}
	footerPoint, err := app.ResolvePoint(ByText("Shell Footer"))
	if err != nil {
		t.Fatal(err)
	}

	if headerPoint.Y >= leftPoint.Y {
		t.Fatalf("expected header above body, got headerY=%d bodyY=%d", headerPoint.Y, leftPoint.Y)
	}
	if leftPoint.Y != contentPoint.Y || contentPoint.Y != rightPoint.Y {
		t.Fatalf("expected left/content/right on same row, got leftY=%d contentY=%d rightY=%d", leftPoint.Y, contentPoint.Y, rightPoint.Y)
	}
	if !(leftPoint.X < contentPoint.X && contentPoint.X < rightPoint.X) {
		t.Fatalf("expected left/content/right in horizontal order, got leftX=%d contentX=%d rightX=%d", leftPoint.X, contentPoint.X, rightPoint.X)
	}
	if footerPoint.Y <= contentPoint.Y {
		t.Fatalf("expected footer below body, got footerY=%d bodyY=%d", footerPoint.Y, contentPoint.Y)
	}
}

func TestE2ELayoutContentOnlyRender(t *testing.T) {
	app, err := Run(newLayoutStaticApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("Only Content")); err != nil {
		t.Fatal(err)
	}
}
