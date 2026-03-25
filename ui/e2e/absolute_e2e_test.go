package e2e

import (
	"testing"

	"github.com/wwsheng009/mint/ui"
	absolutecomp "github.com/wwsheng009/mint/ui/components/absolute"
)

func newAbsoluteStaticApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewVStack().
					SetWidth(20).
					SetHeight(4).
					SetChildrenList([]ui.VNode{
						absolutecomp.NewBuilder(ui.NewTextBuilder("TL").Build()).
							Left(absolutecomp.AbsolutePos(0)).
							Top(absolutecomp.AbsolutePos(0)).
							Width(2).
							Height(1).
							Build(),
					}),
				ui.NewVStack().
					SetWidth(20).
					SetHeight(4).
					SetChildrenList([]ui.VNode{
						absolutecomp.NewBuilder(ui.NewTextBuilder("BR").Build()).
							Right(absolutecomp.AbsolutePos(0)).
							Bottom(absolutecomp.AbsolutePos(0)).
							Width(2).
							Height(1).
							Build(),
					}),
				ui.NewVStack().
					SetWidth(20).
					SetHeight(5).
					SetChildrenList([]ui.VNode{
						absolutecomp.NewBuilder(ui.NewTextBuilder("CC").Build()).
							Left(absolutecomp.RelativePos(50)).
							Top(absolutecomp.RelativePos(50)).
							Anchor(absolutecomp.AnchorCenter).
							Width(2).
							Height(1).
							Build(),
						absolutecomp.NewBuilder(ui.NewTextBuilder("AT").Build()).
							Left(absolutecomp.AbsolutePos(3)).
							Top(absolutecomp.AbsolutePos(1)).
							Width(2).
							Height(1).
							Build(),
					}),
			})
	}
}

func newAbsoluteAtApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetWidth(20).
			SetHeight(5).
			SetChildrenList([]ui.VNode{
				absolutecomp.NewBuilder(ui.NewTextBuilder("AT").Build()).
					Left(absolutecomp.AbsolutePos(3)).
					Top(absolutecomp.AbsolutePos(1)).
					Width(2).
					Height(1).
					Build(),
			})
	}
}

func TestE2EAbsolutePlacesChildrenInExpectedRegions(t *testing.T) {
	app, err := Run(newAbsoluteStaticApp(), ui.WithSize(40, 20))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	tlPoint, err := app.ResolvePoint(ByText("TL"))
	if err != nil {
		t.Fatal(err)
	}
	brPoint, err := app.ResolvePoint(ByText("BR"))
	if err != nil {
		t.Fatal(err)
	}
	ccPoint, err := app.ResolvePoint(ByText("CC"))
	if err != nil {
		t.Fatal(err)
	}

	if tlPoint.X != 0 || tlPoint.Y != 0 {
		t.Fatalf("expected TL at (0,0), got (%d,%d)", tlPoint.X, tlPoint.Y)
	}
	if brPoint.X <= tlPoint.X || brPoint.Y <= tlPoint.Y {
		t.Fatalf("expected BR below/right of TL, got TL=(%d,%d) BR=(%d,%d)", tlPoint.X, tlPoint.Y, brPoint.X, brPoint.Y)
	}
	if ccPoint.X <= tlPoint.X || ccPoint.X >= brPoint.X {
		t.Fatalf("expected CC horizontally between TL and BR, got TLX=%d CCX=%d BRX=%d", tlPoint.X, ccPoint.X, brPoint.X)
	}
	if ccPoint.Y <= tlPoint.Y || ccPoint.Y >= brPoint.Y+5 {
		t.Fatalf("unexpected CC vertical position, got TLY=%d CCY=%d BRY=%d", tlPoint.Y, ccPoint.Y, brPoint.Y)
	}
}

func TestE2EAbsoluteAtPlacesChildAtExactOffset(t *testing.T) {
	app, err := Run(newAbsoluteAtApp(), ui.WithSize(20, 5))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	atPoint, err := app.ResolvePoint(ByText("AT"))
	if err != nil {
		t.Fatal(err)
	}

	if atPoint.X != 3 || atPoint.Y != 1 {
		t.Fatalf("expected AT at (3,1), got (%d,%d)", atPoint.X, atPoint.Y)
	}
}
