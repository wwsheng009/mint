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
	popconfirmcomp "github.com/wwsheng009/mint/ui/components/popconfirm"
)

type popconfirmBackgroundIntent struct{}

func (popconfirmBackgroundIntent) IntentType() string { return "e2e.popconfirm.background" }

type popconfirmFixtureState struct {
	OpenValue        string
	OpenChanges      int
	ConfirmCount     int
	CancelCount      int
	BackgroundClicks int
}

type popconfirmFixtureMeta struct {
	ComponentID           string
	OpenField             string
	FieldChangeIntentType string
	ConfirmIntentType     string
	CancelIntentType      string
	BackgroundIntentType  string
}

func newPopconfirmFixture() (ui.ComponentFunc, func(), func(), *store.Store[popconfirmFixtureState], popconfirmFixtureMeta) {
	fixtureStore := store.NewStore(popconfirmFixtureState{
		OpenValue: "false",
	})
	backgroundIntent := popconfirmBackgroundIntent{}
	meta := popconfirmFixtureMeta{
		ComponentID:           "fixture.popconfirm.delete",
		OpenField:             "fixture.popconfirm.open",
		FieldChangeIntentType: runtimeintent.FieldChangeIntent{}.IntentType(),
		ConfirmIntentType:     popconfirmcomp.PopconfirmConfirmIntent{}.IntentType(),
		CancelIntentType:      popconfirmcomp.PopconfirmCancelIntent{}.IntentType(),
		BackgroundIntentType:  backgroundIntent.IntentType(),
	}

	unregisters := make([]func(), 0, 4)
	initFn := func() {
		rt := rtui.GetGlobalIntentRuntime()
		if rt == nil {
			return
		}
		unregisters = append(unregisters,
			runtimeintent.RegisterTypedRuntime(rt, func(_ *runtimeintent.ActionContext, i runtimeintent.FieldChangeIntent) runtimeintent.IntentResult {
				if i.Field != meta.OpenField {
					return runtimeintent.IntentResult{}
				}
				fixtureStore.Update(func(s popconfirmFixtureState) popconfirmFixtureState {
					s.OpenValue = i.Value
					s.OpenChanges++
					return s
				})
				return runtimeintent.HandledResult()
			}),
			runtimeintent.RegisterTypedRuntime(rt, func(_ *runtimeintent.ActionContext, i popconfirmcomp.PopconfirmConfirmIntent) runtimeintent.IntentResult {
				if i.ComponentID != meta.ComponentID {
					return runtimeintent.IntentResult{}
				}
				fixtureStore.Update(func(s popconfirmFixtureState) popconfirmFixtureState {
					s.ConfirmCount++
					return s
				})
				return runtimeintent.HandledResult()
			}),
			runtimeintent.RegisterTypedRuntime(rt, func(_ *runtimeintent.ActionContext, i popconfirmcomp.PopconfirmCancelIntent) runtimeintent.IntentResult {
				if i.ComponentID != meta.ComponentID {
					return runtimeintent.IntentResult{}
				}
				fixtureStore.Update(func(s popconfirmFixtureState) popconfirmFixtureState {
					s.CancelCount++
					return s
				})
				return runtimeintent.HandledResult()
			}),
			runtimeintent.RegisterTypedRuntime(rt, func(_ *runtimeintent.ActionContext, _ popconfirmBackgroundIntent) runtimeintent.IntentResult {
				fixtureStore.Update(func(s popconfirmFixtureState) popconfirmFixtureState {
					s.BackgroundClicks++
					return s
				})
				return runtimeintent.HandledResult()
			}),
		)
	}

	cleanupFn := func() {
		for i := len(unregisters) - 1; i >= 0; i-- {
			if unregisters[i] != nil {
				unregisters[i]()
			}
		}
	}

	appFn := func() ui.VNode {
		openValue := ui.UseStoreSelector(fixtureStore, func(s popconfirmFixtureState) string { return s.OpenValue })
		openChanges := ui.UseStoreSelector(fixtureStore, func(s popconfirmFixtureState) int { return s.OpenChanges })
		confirmCount := ui.UseStoreSelector(fixtureStore, func(s popconfirmFixtureState) int { return s.ConfirmCount })
		cancelCount := ui.UseStoreSelector(fixtureStore, func(s popconfirmFixtureState) int { return s.CancelCount })
		backgroundClicks := ui.UseStoreSelector(fixtureStore, func(s popconfirmFixtureState) int { return s.BackgroundClicks })

		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Popconfirm E2E Fixture").Build(),
				ui.NewPopconfirmBuilder(
					ui.NewButtonBuilder("Delete record").Build(),
				).
					SetID("fixture-popconfirm").
					ComponentID(meta.ComponentID).
					Title("Delete record?").
					Description("This action cannot be undone.").
					OkText("Delete now").
					CancelText("Keep item").
					Placement(ui.PopconfirmPlacementBottomLeft).
					OpenForField(runtimeintent.BindField(meta.OpenField)).
					Build(),
				ui.NewTextBuilder(fmt.Sprintf("OpenValue: %s", openValue)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("OpenChanges: %d", openChanges)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("ConfirmCount: %d", confirmCount)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("CancelCount: %d", cancelCount)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("BackgroundClicks: %d", backgroundClicks)).Build(),
				ui.NewTextBuilder(" ").Build(),
				ui.NewTextBuilder(" ").Build(),
				ui.NewButtonBuilder("Popconfirm background action").
					SetID("popconfirm-background-btn").
					OnPress(backgroundIntent).
					Build(),
			})
	}

	return appFn, initFn, cleanupFn, fixtureStore, meta
}

func TestE2EPopconfirmAnchorOpenCancelAndConfirmFlows(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newPopconfirmFixture()
	defer cleanupFn()

	app, err := Run(appFn,
		ui.WithSize(96, 24),
		ui.WithInit(initFn),
		ui.WithPluginSetup(func(app *framework.App) {
			popconfirmcomp.Install(app)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("OpenValue: false")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	if err := app.Driver().Click(ByText("Delete record")); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitIntent(meta.FieldChangeIntentType, 500*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("Delete record?")); err != nil {
			return err
		}
		if fixtureStore.Get().OpenValue != "true" {
			return fmt.Errorf("open value = %q, want true", fixtureStore.Get().OpenValue)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	if err := app.Driver().Click(ByText("Keep item")); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitIntent(meta.CancelIntentType, 500*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("Delete record?")); err == nil {
			return fmt.Errorf("popconfirm still visible after cancel")
		}
		state := fixtureStore.Get()
		if state.OpenValue != "false" {
			return fmt.Errorf("open value = %q, want false", state.OpenValue)
		}
		if state.CancelCount != 1 {
			return fmt.Errorf("cancel count = %d, want 1", state.CancelCount)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	if err := app.Driver().Click(ByText("Delete record")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("Delete record?"))
	}); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	if err := app.Driver().Click(ByText("Delete now")); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitIntent(meta.ConfirmIntentType, 500*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("Delete record?")); err == nil {
			return fmt.Errorf("popconfirm still visible after confirm")
		}
		state := fixtureStore.Get()
		if state.OpenValue != "false" {
			return fmt.Errorf("open value = %q, want false", state.OpenValue)
		}
		if state.ConfirmCount != 1 {
			return fmt.Errorf("confirm count = %d, want 1", state.ConfirmCount)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	state := fixtureStore.Get()
	if state.CancelCount != 1 || state.ConfirmCount != 1 || state.OpenValue != "false" {
		t.Fatalf("unexpected popconfirm state after confirm/cancel flow: %+v", state)
	}
}

func TestE2EPopconfirmOutsideBackgroundClickClosesAndContinuesDispatch(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newPopconfirmFixture()
	defer cleanupFn()

	app, err := Run(appFn,
		ui.WithSize(96, 24),
		ui.WithInit(initFn),
		ui.WithPluginSetup(func(app *framework.App) {
			popconfirmcomp.Install(app)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.Driver().Click(ByText("Delete record")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("Delete record?"))
	}); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	if err := app.Driver().Click(ByID("popconfirm-background-btn")); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitIntent(meta.BackgroundIntentType, 500*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("Delete record?")); err == nil {
			return fmt.Errorf("popconfirm still visible after outside background click")
		}
		state := fixtureStore.Get()
		if state.OpenValue != "false" {
			return fmt.Errorf("open value = %q, want false", state.OpenValue)
		}
		if state.BackgroundClicks != 1 {
			return fmt.Errorf("background clicks = %d, want 1", state.BackgroundClicks)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	state := fixtureStore.Get()
	if state.BackgroundClicks != 1 || state.OpenValue != "false" {
		t.Fatalf("unexpected popconfirm state after outside close: %+v", state)
	}
}
