package e2e

import (
	"fmt"
	"testing"
	"time"

	runtimeintent "github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/runtime/store"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
)

type buttonPrimaryPressIntent struct{}

func (buttonPrimaryPressIntent) IntentType() string { return "e2e.button.primary_press" }

type buttonSecondaryPressIntent struct{}

func (buttonSecondaryPressIntent) IntentType() string { return "e2e.button.secondary_press" }

type buttonDisabledPressIntent struct{}

func (buttonDisabledPressIntent) IntentType() string { return "e2e.button.disabled_press" }

type buttonFixtureState struct {
	PrimaryPresses   int
	SecondaryPresses int
	LastIntent       string
}

type buttonFixtureMeta struct {
	PrimaryIntentType   string
	SecondaryIntentType string
	DisabledIntentType  string
}

func newButtonFixture() (ui.ComponentFunc, func(), func(), *store.Store[buttonFixtureState], buttonFixtureMeta) {
	fixtureStore := store.NewStore(buttonFixtureState{})
	meta := buttonFixtureMeta{
		PrimaryIntentType:   buttonPrimaryPressIntent{}.IntentType(),
		SecondaryIntentType: buttonSecondaryPressIntent{}.IntentType(),
		DisabledIntentType:  buttonDisabledPressIntent{}.IntentType(),
	}

	unregisters := make([]func(), 0, 2)
	initFn := func() {
		rt := rtui.GetGlobalIntentRuntime()
		if rt == nil {
			return
		}
		unregisters = append(unregisters,
			runtimeintent.RegisterTypedRuntime(rt, func(_ *runtimeintent.ActionContext, _ buttonPrimaryPressIntent) runtimeintent.IntentResult {
				fixtureStore.Update(func(s buttonFixtureState) buttonFixtureState {
					s.PrimaryPresses++
					s.LastIntent = meta.PrimaryIntentType
					return s
				})
				return runtimeintent.HandledResult()
			}),
			runtimeintent.RegisterTypedRuntime(rt, func(_ *runtimeintent.ActionContext, _ buttonSecondaryPressIntent) runtimeintent.IntentResult {
				fixtureStore.Update(func(s buttonFixtureState) buttonFixtureState {
					s.SecondaryPresses++
					s.LastIntent = meta.SecondaryIntentType
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
		primaryPresses := ui.UseStoreSelector(fixtureStore, func(s buttonFixtureState) int { return s.PrimaryPresses })
		secondaryPresses := ui.UseStoreSelector(fixtureStore, func(s buttonFixtureState) int { return s.SecondaryPresses })
		lastIntent := ui.UseStoreSelector(fixtureStore, func(s buttonFixtureState) string { return s.LastIntent })

		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Button E2E Fixture").Build(),
				ui.NewButtonBuilder("Save changes").
					SetID("button-primary").
					OnPress(buttonPrimaryPressIntent{}).
					Build(),
				ui.NewButtonBuilder("Disabled action").
					SetID("button-disabled").
					Disabled(true).
					OnPress(buttonDisabledPressIntent{}).
					Build(),
				ui.NewButtonBuilder("Open preview").
					SetID("button-secondary").
					OnPress(buttonSecondaryPressIntent{}).
					Build(),
				ui.NewButtonBuilder("Button tail action").
					SetID("button-tail").
					Build(),
				ui.NewTextBuilder(fmt.Sprintf("PrimaryPresses: %d", primaryPresses)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("SecondaryPresses: %d", secondaryPresses)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("LastIntent: %s", lastIntent)).Build(),
			})
	}

	return appFn, initFn, cleanupFn, fixtureStore, meta
}

func TestE2EButtonKeyboardAndClickFlow(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newButtonFixture()
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(84, 16), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertFocus(ByID("button-primary")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("PrimaryPresses: 0")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("SecondaryPresses: 0")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Special(platform.KeyEnter); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("PrimaryPresses: 1")); err != nil {
			return err
		}
		return app.AssertVisible(ByText("LastIntent: e2e.button.primary_press"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "key:Enter"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.PrimaryIntentType); err != nil {
		t.Fatal(err)
	}

	point, err := app.ResolvePoint(ByText("Open preview"))
	if err != nil {
		t.Fatal(err)
	}
	if err := app.AssertHit(point, ByID("button-secondary")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Click(ByID("button-secondary")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("SecondaryPresses: 1")); err != nil {
			return err
		}
		return app.AssertVisible(ByText("LastIntent: e2e.button.secondary_press"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "mouse:left:release"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.SecondaryIntentType); err != nil {
		t.Fatal(err)
	}

	state := fixtureStore.Get()
	if state.PrimaryPresses != 1 || state.SecondaryPresses != 1 || state.LastIntent != meta.SecondaryIntentType {
		t.Fatalf("unexpected button fixture state after keyboard and click flow: %+v", state)
	}
}

func TestE2EButtonDisabledIgnoreAndTabSkip(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newButtonFixture()
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(84, 16), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	point, err := app.ResolvePoint(ByText("Disabled action"))
	if err != nil {
		t.Fatal(err)
	}
	if err := app.AssertHit(point, ByID("button-disabled")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Click(ByID("button-disabled")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "mouse:left:release"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("PrimaryPresses: 0")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("SecondaryPresses: 0")); err != nil {
		t.Fatal(err)
	}
	for _, logEntry := range app.IntentLogs() {
		if logEntry.Type == meta.PrimaryIntentType || logEntry.Type == meta.SecondaryIntentType || logEntry.Type == meta.DisabledIntentType {
			t.Fatalf("disabled button should not emit press intents, got %+v", logEntry)
		}
	}

	app.ClearRawInputs()
	if err := app.Driver().Special(platform.KeyTab); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertFocus(ByID("button-secondary")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "key:Tab"}); err != nil {
		t.Fatal(err)
	}

	app.ClearRawInputs()
	if err := app.Driver().Special(platform.KeyTab); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertFocus(ByID("button-tail")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "key:Tab"}); err != nil {
		t.Fatal(err)
	}

	state := fixtureStore.Get()
	if state.PrimaryPresses != 0 || state.SecondaryPresses != 0 || state.LastIntent != "" {
		t.Fatalf("unexpected button fixture state after disabled and tab flow: %+v", state)
	}
}
