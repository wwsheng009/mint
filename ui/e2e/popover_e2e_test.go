package e2e

import (
	"fmt"
	"testing"
	"time"

	"github.com/wwsheng009/mint/framework"
	runtimeintent "github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/store"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
	popovercomp "github.com/wwsheng009/mint/ui/components/popover"
)

type popoverBackgroundIntent struct{}

func (popoverBackgroundIntent) IntentType() string { return "e2e.popover.background" }

type popoverFixtureState struct {
	BackgroundClicks int
}

type popoverFixtureMeta struct {
	BackgroundIntentType string
}

func newPopoverFixture() (ui.ComponentFunc, func(), func(), *store.Store[popoverFixtureState], popoverFixtureMeta) {
	fixtureStore := store.NewStore(popoverFixtureState{})
	backgroundIntent := popoverBackgroundIntent{}
	meta := popoverFixtureMeta{
		BackgroundIntentType: backgroundIntent.IntentType(),
	}

	unregisters := make([]func(), 0, 1)
	initFn := func() {
		rt := rtui.GetGlobalIntentRuntime()
		if rt == nil {
			return
		}
		unregisters = append(unregisters, runtimeintent.RegisterTypedRuntime(rt, func(_ *runtimeintent.ActionContext, _ popoverBackgroundIntent) runtimeintent.IntentResult {
			fixtureStore.Update(func(s popoverFixtureState) popoverFixtureState {
				s.BackgroundClicks++
				return s
			})
			return runtimeintent.HandledResult()
		}))
	}

	cleanupFn := func() {
		for i := len(unregisters) - 1; i >= 0; i-- {
			if unregisters[i] != nil {
				unregisters[i]()
			}
		}
	}

	appFn := func() ui.VNode {
		backgroundClicks := ui.UseStoreSelector(fixtureStore, func(s popoverFixtureState) int { return s.BackgroundClicks })

		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Popover E2E Fixture").Build(),
				ui.NewPopoverBuilder(
					ui.NewButtonBuilder("Click popover anchor").
						SetID("popover-click-anchor").
						OnPress(popovercomp.ToggleWithID("fixture.popover.click")).
						Build(),
				).
					SetID("fixture-popover-click").
					ComponentID("fixture.popover.click").
					Title("Click Popover").
					Body("Click popover body").
					Placement(ui.PopoverPlacementBottomLeft).
					Trigger(ui.PopoverTriggerClick).
					Build(),
				ui.NewPopoverBuilder(
					ui.NewTextBuilder("Hover popover anchor").SetID("popover-hover-anchor").Build(),
				).
					SetID("fixture-popover-hover").
					ComponentID("fixture.popover.hover").
					Title("Hover Popover").
					Body("Hover popover body").
					Placement(ui.PopoverPlacementBottomLeft).
					Trigger(ui.PopoverTriggerHover).
					Build(),
				ui.NewTextBuilder(fmt.Sprintf("BackgroundClicks: %d", backgroundClicks)).Build(),
				ui.NewTextBuilder(" ").Build(),
				ui.NewTextBuilder(" ").Build(),
				ui.NewTextBuilder(" ").Build(),
				ui.NewButtonBuilder("Popover background action").
					SetID("popover-background-btn").
					OnPress(backgroundIntent).
					Build(),
				ui.NewButtonBuilder("Popover neutral target").
					SetID("popover-neutral-target").
					Build(),
			})
	}

	return appFn, initFn, cleanupFn, fixtureStore, meta
}

func TestE2EPopoverClickOpenAndOutsideCloseFlow(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newPopoverFixture()
	defer cleanupFn()

	app, err := Run(appFn,
		ui.WithSize(96, 20),
		ui.WithInit(initFn),
		ui.WithPluginSetup(func(app *framework.App) {
			popovercomp.Install(app)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("BackgroundClicks: 0")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	if err := app.Driver().Click(ByID("popover-click-anchor")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("Click popover body"))
	}); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	if err := app.Driver().Click(ByID("popover-background-btn")); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitIntent(meta.BackgroundIntentType, 500*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("Click popover body")); err == nil {
			return fmt.Errorf("click popover body still visible after outside background click")
		}
		return app.AssertVisible(ByText("BackgroundClicks: 1"))
	}); err != nil {
		t.Fatal(err)
	}

	state := fixtureStore.Get()
	if state.BackgroundClicks != 1 {
		t.Fatalf("backgroundClicks = %d, want 1", state.BackgroundClicks)
	}
}

func TestE2EPopoverHoverShowsAndHides(t *testing.T) {
	appFn, initFn, cleanupFn, _ /* fixtureStore */, _ := newPopoverFixture()
	defer cleanupFn()

	app, err := Run(appFn,
		ui.WithSize(96, 20),
		ui.WithInit(initFn),
		ui.WithPluginSetup(func(app *framework.App) {
			popovercomp.Install(app)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.Driver().Move(ByID("popover-neutral-target")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Hover popover body")); err == nil {
		t.Fatal("hover popover should be hidden before hover")
	}

	if err := app.Driver().Move(ByID("popover-hover-anchor")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("Hover popover body"))
	}); err != nil {
		t.Fatal(err)
	}

	if err := app.Driver().Move(ByID("popover-neutral-target")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("Hover popover body")); err == nil {
			return fmt.Errorf("hover popover body still visible after hover exit")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
