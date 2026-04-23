package e2e

import (
	"testing"

	fwtheme "github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/ui"
	cursorcomp "github.com/wwsheng009/mint/ui/components/cursor"
)

func newCursorStaticApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Cursor E2E Fixture").Build(),
				cursorcomp.NewBuilder().
					Key("bar-cursor").
					Bar().
					Steady().
					Build(),
				cursorcomp.NewBuilder().
					Key("block-cursor").
					Block().
					Glyph("X").
					Steady().
					Theme(cursorcomp.ThemeAccent).
					Build(),
				cursorcomp.NewBuilder().
					Key("hidden-cursor").
					Bar().
					Visible(false).
					Build(),
			})
	}
}

func TestE2ECursorVisibleAndHiddenRender(t *testing.T) {
	app, err := Run(newCursorStaticApp(), ui.WithSize(96, 20))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("|")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("X")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Cursor E2E Fixture")); err != nil {
		t.Fatal(err)
	}
}

func TestE2ECursorStylesRender(t *testing.T) {
	app, err := Run(newCursorStaticApp(), ui.WithSize(96, 20))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertStyle(ByText("|"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Caret(),
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("X"), StyleExpect{
		HasFG:      true,
		FG:         fwtheme.Primary(),
		HasReverse: true,
		Reverse:    true,
	}); err != nil {
		t.Fatal(err)
	}
}
