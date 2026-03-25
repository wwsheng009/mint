package e2e

import (
	"testing"

	fwtheme "github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/ui"
	badgecomp "github.com/wwsheng009/mint/ui/components/badge"
)

func newBadgeStaticApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Badge E2E Fixture").Build(),
				badgecomp.NewBuilder("Inbox").
					Count(120).
					OverflowCount(99).
					Error().
					Build(),
				badgecomp.NewBuilder("Live").
					Dot(true).
					Success().
					Build(),
				badgecomp.NewBuilder("Queue").
					Count(0).
					ShowZero(true).
					Processing().
					Build(),
				badgecomp.NewBuilder("").
					Text("NEW").
					Primary().
					Build(),
			})
	}
}

func TestE2EBadgeCountDotZeroAndCustomTextRender(t *testing.T) {
	app, err := Run(newBadgeStaticApp(), ui.WithSize(96, 20))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("Inbox [99+]")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Live ●")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Queue [0]")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("[NEW]")); err != nil {
		t.Fatal(err)
	}
}

func TestE2EBadgeStatusStylesMatchVariants(t *testing.T) {
	app, err := Run(newBadgeStaticApp(), ui.WithSize(96, 20))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertStyle(ByText("[99+]"), StyleExpect{
		HasBG:   true,
		BG:      fwtheme.Error(),
		HasFG:   true,
		FG:      fwtheme.BG(),
		HasBold: true,
		Bold:    true,
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("●"), StyleExpect{
		HasFG:   true,
		FG:      fwtheme.Success(),
		HasBold: true,
		Bold:    true,
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("[0]"), StyleExpect{
		HasBG: true,
		BG:    fwtheme.Primary(),
		HasFG: true,
		FG:    fwtheme.BG(),
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("[NEW]"), StyleExpect{
		HasBG: true,
		BG:    fwtheme.Primary(),
		HasFG: true,
		FG:    fwtheme.BG(),
	}); err != nil {
		t.Fatal(err)
	}
}
