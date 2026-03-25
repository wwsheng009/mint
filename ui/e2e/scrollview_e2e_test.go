package e2e

import (
	"fmt"
	"testing"
	"time"

	runtimeaction "github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/ui"
	scrollviewcomp "github.com/wwsheng009/mint/ui/components/scrollview"
)

func newScrollViewStaticApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("ScrollView E2E Fixture").Build(),
				scrollviewcomp.NewBuilder().
					SetID("scroll-main").
					Child(ui.NewTextBuilder("Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6").Build()).
					Width(12).
					Height(3).
					ShowBorder(true).
					ShowIndicator(true).
					Build(),
				scrollviewcomp.NewBuilder().
					SetID("scroll-plain").
					Child(ui.NewTextBuilder("Alpha\nBeta\nGamma\nDelta").Build()).
					Width(10).
					Height(2).
					ShowBorder(false).
					Build(),
			})
	}
}

func wheelAt(app *App, locator Locator, action platform.MouseAction) error {
	point, err := app.ResolvePoint(locator)
	if err != nil {
		return err
	}
	return wheelAtPoint(app, point.X, point.Y, action)
}

func wheelAtPoint(app *App, x, y int, action platform.MouseAction) error {
	raw := app.Driver().mouseRawInput(x, y, platform.MouseNone, action)
	return app.Driver().mouseSequence(false, 0, raw)
}

func TestE2EScrollViewWheelScrollAndIndicatorRender(t *testing.T) {
	app, err := Run(newScrollViewStaticApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("Line 1")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Line 3")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("↓")); err != nil {
		t.Fatal(err)
	}
	scrollPoint, err := app.ResolvePoint(ByText("Line 2"))
	if err != nil {
		t.Fatal(err)
	}

	app.ClearRawInputs()
	if err := wheelAtPoint(app, scrollPoint.X, scrollPoint.Y, platform.MouseWheelDown); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "mouse:none:wheel_down"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertActionSequence(runtimeaction.ActionScroll); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertActionHandled(runtimeaction.ActionScroll, "mouse_target"); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("Line 2")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("Line 4")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("↕")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("Line 1")); err == nil {
			return fmt.Errorf("Line 1 should be scrolled out of viewport")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	app.ClearRawInputs()
	for i := 0; i < 3; i++ {
		if err := wheelAtPoint(app, scrollPoint.X, scrollPoint.Y, platform.MouseWheelDown); err != nil {
			t.Fatal(err)
		}
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("Line 4")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("Line 6")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("↑")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("Line 3")); err == nil {
			return fmt.Errorf("Line 3 should be above bottom viewport")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestE2EScrollViewBorderlessWheelScrollRender(t *testing.T) {
	app, err := Run(newScrollViewStaticApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("Alpha")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Beta")); err != nil {
		t.Fatal(err)
	}
	scrollPoint, err := app.ResolvePoint(ByText("Beta"))
	if err != nil {
		t.Fatal(err)
	}

	app.ClearRawInputs()
	if err := wheelAtPoint(app, scrollPoint.X, scrollPoint.Y, platform.MouseWheelDown); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "mouse:none:wheel_down"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertActionSequence(runtimeaction.ActionScroll); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertActionHandled(runtimeaction.ActionScroll, "mouse_target"); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("Beta")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("Gamma")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("Alpha")); err == nil {
			return fmt.Errorf("Alpha should be scrolled out of borderless viewport")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	app.ClearRawInputs()
	if err := wheelAtPoint(app, scrollPoint.X, scrollPoint.Y, platform.MouseWheelUp); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "mouse:none:wheel_up"}); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("Alpha")); err != nil {
			return err
		}
		return app.AssertVisible(ByText("Beta"))
	}); err != nil {
		t.Fatal(err)
	}
}
