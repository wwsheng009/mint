package e2e

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
	textcomp "github.com/wwsheng009/mint/ui/components/text"
)

func newTextStaticApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Text E2E Fixture").Build(),
				ui.NewTextBuilder("Styled sample").
					FgColor("green").
					BgColor("blue").
					Bold(true).
					Underline(true).
					Padding(0, 1, 0, 1).
					Build(),
				ui.NewTextBuilder("VeryLongText").
					MaxWidth(5).
					Build(),
				textcomp.New("alpha beta gamma").
					SetWrap(true).
					SetMaxWidth(7),
			})
	}
}

func TestE2ETextStylesPaddingAndTruncationRender(t *testing.T) {
	app, err := Run(newTextStaticApp(), ui.WithSize(96, 20))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("Styled sample")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("Styled sample"), StyleExpect{
		HasFG:        true,
		FG:           style.Color("green"),
		HasBG:        true,
		BG:           style.Color("blue"),
		HasBold:      true,
		Bold:         true,
		HasUnderline: true,
		Underline:    true,
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("VeryL")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("VeryLongText")); err == nil {
		t.Fatal("expected truncated text to hide full content")
	}
}

func TestE2ETextWrapRendersAcrossMultipleLines(t *testing.T) {
	app, err := Run(newTextStaticApp(), ui.WithSize(96, 20))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	alphaPoint, err := app.ResolvePoint(ByText("alpha"))
	if err != nil {
		t.Fatal(err)
	}
	betaPoint, err := app.ResolvePoint(ByText("beta"))
	if err != nil {
		t.Fatal(err)
	}
	gammaPoint, err := app.ResolvePoint(ByText("gamma"))
	if err != nil {
		t.Fatal(err)
	}

	if alphaPoint.Y >= betaPoint.Y {
		t.Fatalf("expected alpha above beta, got alphaY=%d betaY=%d", alphaPoint.Y, betaPoint.Y)
	}
	if betaPoint.Y >= gammaPoint.Y {
		t.Fatalf("expected beta above gamma, got betaY=%d gammaY=%d", betaPoint.Y, gammaPoint.Y)
	}
}
