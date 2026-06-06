package e2e

import (
	"fmt"
	"testing"
	"time"

	runtimeaction "github.com/wwsheng009/mint/runtime/action"
	runtimeintent "github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/runtime/store"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
)

type pageViewportPressIntent struct {
	Name string
}

func (i pageViewportPressIntent) IntentType() string { return "e2e.pageviewport." + i.Name }

type pageViewportFixtureState struct {
	Presses int
	Last    string
}

func newPageViewportFixture(offset *int) (ui.ComponentFunc, func(), func(), *store.Store[pageViewportFixtureState]) {
	fixtureStore := store.NewStore(pageViewportFixtureState{})
	unregisters := make([]func(), 0, 2)
	initFn := func() {
		rt := rtui.GetGlobalIntentRuntime()
		if rt == nil {
			return
		}
		for _, name := range []string{"first", "tail"} {
			name := name
			unregisters = append(unregisters, rt.Register("e2e.pageviewport."+name, runtimeintent.HandlerFunc(func(_ *runtimeintent.ActionContext, _ runtimeintent.Intent) runtimeintent.IntentResult {
				fixtureStore.Update(func(s pageViewportFixtureState) pageViewportFixtureState {
					s.Presses++
					s.Last = name
					return s
				})
				return runtimeintent.HandledResult()
			})))
		}
	}
	cleanupFn := func() {
		for i := len(unregisters) - 1; i >= 0; i-- {
			if unregisters[i] != nil {
				unregisters[i]()
			}
		}
	}
	appFn := func() ui.VNode {
		state := ui.UseStoreSelector(fixtureStore, func(s pageViewportFixtureState) pageViewportFixtureState { return s })
		content := ui.NewVStack().SetGap(0).SetChildrenList([]ui.VNode{
			ui.Text("Viewport row 0"),
			ui.NewButtonBuilder("First action").SetID("pageviewport-first-action").OnPress(pageViewportPressIntent{Name: "first"}).Build(),
			ui.Text("Viewport row 2"),
			ui.Text("Viewport row 3"),
			ui.NewButtonBuilder("Tail action").SetID("pageviewport-tail-action").OnPress(pageViewportPressIntent{Name: "tail"}).Build(),
			ui.Text("Viewport row 5"),
		})
		viewport := ui.NewPageViewportBuilder().Child(content).Size(34, 4)
		if offset != nil {
			viewport.ScrollOffset(*offset)
		}
		return ui.PageStack(
			ui.Text("PageViewport E2E Fixture"),
			viewport.Build(),
			ui.Text(fmt.Sprintf("Presses: %d Last: %s", state.Presses, state.Last)),
		)
	}
	return appFn, initFn, cleanupFn, fixtureStore
}

func TestE2EPageViewportClipsPaintAndHitMap(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore := newPageViewportFixture(nil)
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(80, 16), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{"Viewport row 0", "First action", "Viewport row 3"} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatalf("expected %q to be visible: %v\n%s", text, err, app.RenderString())
		}
	}
	for _, locator := range []Locator{ByText("Tail action"), ByID("pageviewport-tail-action")} {
		if err := app.AssertVisible(locator); err == nil {
			t.Fatalf("tail action should be clipped out of viewport\n%s", app.RenderString())
		}
	}
	indicator, err := app.CellAt(33, 1)
	if err != nil {
		t.Fatal(err)
	}
	if indicator.Cluster != "#" {
		t.Fatalf("scroll indicator top cell = %q, want #\n%s", indicator.Cluster, app.RenderString())
	}
	bottomIndicator, err := app.CellAt(33, 4)
	if err != nil {
		t.Fatal(err)
	}
	if bottomIndicator.Cluster != "v" {
		t.Fatalf("scroll indicator bottom cell = %q, want v\n%s", bottomIndicator.Cluster, app.RenderString())
	}

	point, err := app.ResolvePoint(ByID("pageviewport-first-action"))
	if err != nil {
		t.Fatal(err)
	}
	if err := app.AssertHit(point, ByID("pageviewport-first-action")); err != nil {
		t.Fatalf("visible child button should stay hit-testable: %v\n%s", err, app.RenderString())
	}
	if _, err := app.ResolvePoint(ByID("pageviewport-tail-action")); err == nil {
		t.Fatalf("clipped child button should not be in hitmap\n%s", app.RenderString())
	}

	app.ClearIntentLogs()
	if err := app.Driver().Click(ByID("pageviewport-first-action")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("Presses: 1 Last: first"))
	}); err != nil {
		t.Fatal(err)
	}
	if got := fixtureStore.Get(); got.Presses != 1 || got.Last != "first" {
		t.Fatalf("unexpected fixture state after visible click: %+v", got)
	}
}

func TestE2EPageViewportControlledOffsetRevealsScrolledChildren(t *testing.T) {
	offset := 4
	appFn, initFn, cleanupFn, fixtureStore := newPageViewportFixture(&offset)
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(80, 16), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{"Tail action", "Viewport row 5"} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatalf("expected %q after controlled scroll: %v\n%s", text, err, app.RenderString())
		}
	}
	if err := app.AssertVisible(ByText("First action")); err == nil {
		t.Fatalf("first action should be above the scrolled viewport\n%s", app.RenderString())
	}
	if _, err := app.ResolvePoint(ByID("pageviewport-first-action")); err == nil {
		t.Fatalf("scrolled-out first action should not be in hitmap\n%s", app.RenderString())
	}

	point, err := app.ResolvePoint(ByID("pageviewport-tail-action"))
	if err != nil {
		t.Fatal(err)
	}
	if err := app.AssertHit(point, ByID("pageviewport-tail-action")); err != nil {
		t.Fatalf("scrolled child button should be hit-testable: %v\n%s", err, app.RenderString())
	}

	app.ClearIntentLogs()
	if err := app.Driver().Click(ByID("pageviewport-tail-action")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("Presses: 1 Last: tail"))
	}); err != nil {
		t.Fatal(err)
	}
	if got := fixtureStore.Get(); got.Presses != 1 || got.Last != "tail" {
		t.Fatalf("unexpected fixture state after scrolled click: %+v", got)
	}
}

func TestE2EPageViewportWheelScrollsUncontrolledViewport(t *testing.T) {
	appFn, initFn, cleanupFn, _ := newPageViewportFixture(nil)
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(80, 16), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	scrollPoint, err := app.ResolvePoint(ByText("Viewport row 2"))
	if err != nil {
		t.Fatal(err)
	}
	app.ClearRawInputs()
	if err := wheelAtPoint(app, scrollPoint.X, scrollPoint.Y, platform.MouseWheelDown); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertActionSequence(runtimeaction.ActionScroll); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertActionHandled(runtimeaction.ActionScroll, "mouse_target"); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("Tail action")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("Viewport row 0")); err == nil {
			return fmt.Errorf("row 0 should be scrolled out of viewport")
		}
		indicator, err := app.CellAt(33, 1)
		if err != nil {
			return err
		}
		if indicator.Cluster != "^" {
			return fmt.Errorf("top scroll indicator = %q, want ^", indicator.Cluster)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
