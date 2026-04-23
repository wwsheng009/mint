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

type switchFixtureState struct {
	AirplaneEnabled bool
	AirplaneChanges int
	PowerEnabled    bool
	PowerChanges    int
	LastField       string
}

type switchFixtureMeta struct {
	AirplaneField         string
	PowerField            string
	FieldChangeIntentType string
}

func newSwitchFixture() (ui.ComponentFunc, func(), func(), *store.Store[switchFixtureState], switchFixtureMeta) {
	fixtureStore := store.NewStore(switchFixtureState{})
	meta := switchFixtureMeta{
		AirplaneField:         "fixture.switch.airplane",
		PowerField:            "fixture.switch.power",
		FieldChangeIntentType: runtimeintent.FieldChangeIntent{}.IntentType(),
	}

	unregisters := make([]func(), 0, 1)
	initFn := func() {
		rt := rtui.GetGlobalIntentRuntime()
		if rt == nil {
			return
		}
		unregisters = append(unregisters, runtimeintent.RegisterTypedRuntime(rt, func(_ *runtimeintent.ActionContext, i runtimeintent.FieldChangeIntent) runtimeintent.IntentResult {
			switch i.Field {
			case meta.AirplaneField:
				fixtureStore.Update(func(s switchFixtureState) switchFixtureState {
					s.AirplaneEnabled = i.Value == "true"
					s.AirplaneChanges++
					s.LastField = i.Field
					return s
				})
				return runtimeintent.HandledResult()
			case meta.PowerField:
				fixtureStore.Update(func(s switchFixtureState) switchFixtureState {
					s.PowerEnabled = i.Value == "true"
					s.PowerChanges++
					s.LastField = i.Field
					return s
				})
				return runtimeintent.HandledResult()
			default:
				return runtimeintent.IntentResult{}
			}
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
		airplaneEnabled := ui.UseStoreSelector(fixtureStore, func(s switchFixtureState) bool { return s.AirplaneEnabled })
		airplaneChanges := ui.UseStoreSelector(fixtureStore, func(s switchFixtureState) int { return s.AirplaneChanges })
		powerEnabled := ui.UseStoreSelector(fixtureStore, func(s switchFixtureState) bool { return s.PowerEnabled })
		powerChanges := ui.UseStoreSelector(fixtureStore, func(s switchFixtureState) int { return s.PowerChanges })
		lastField := ui.UseStoreSelector(fixtureStore, func(s switchFixtureState) string { return s.LastField })

		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Switch E2E Fixture").Build(),
				ui.NewSwitchBuilder().
					SetID("switch-airplane").
					Label("Airplane mode").
					Checked(airplaneEnabled).
					Labels("ON", "OFF").
					ForField(runtimeintent.BindField(meta.AirplaneField)).
					Build(),
				ui.NewSwitchBuilder().
					SetID("switch-power").
					Label("Power saver").
					Checked(powerEnabled).
					Labels("YES", "NO").
					Disabled(true).
					ForField(runtimeintent.BindField(meta.PowerField)).
					Build(),
				ui.NewTextBuilder(fmt.Sprintf("AirplaneEnabled: %t", airplaneEnabled)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("AirplaneChanges: %d", airplaneChanges)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("PowerEnabled: %t", powerEnabled)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("PowerChanges: %d", powerChanges)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("LastField: %s", lastField)).Build(),
			})
	}

	return appFn, initFn, cleanupFn, fixtureStore, meta
}

func TestE2ESwitchKeyboardAndClickToggleFlow(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newSwitchFixture()
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(84, 18), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertFocus(ByID("switch-airplane")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("AirplaneEnabled: false")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Special(platform.KeyEnter); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("AirplaneEnabled: true"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "key:Enter"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("AirplaneChanges: 1")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("LastField: fixture.switch.airplane")); err != nil {
		t.Fatal(err)
	}

	point, err := app.ResolvePoint(ByText("Airplane mode"))
	if err != nil {
		t.Fatal(err)
	}
	if err := app.AssertHit(point, ByID("switch-airplane")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Click(ByID("switch-airplane")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("AirplaneEnabled: false"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "mouse:left:release"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("AirplaneChanges: 2")); err != nil {
		t.Fatal(err)
	}

	state := fixtureStore.Get()
	if state.AirplaneEnabled || state.AirplaneChanges != 2 || state.LastField != meta.AirplaneField {
		t.Fatalf("unexpected switch fixture state after toggle flow: %+v", state)
	}
}

func TestE2ESwitchDisabledIgnoresClick(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newSwitchFixture()
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(84, 18), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	point, err := app.ResolvePoint(ByText("Power saver"))
	if err != nil {
		t.Fatal(err)
	}
	if err := app.AssertHit(point, ByID("switch-power")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Click(ByID("switch-power")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "mouse:left:release"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("PowerEnabled: false")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("PowerChanges: 0")); err != nil {
		t.Fatal(err)
	}

	for _, logEntry := range app.IntentLogs() {
		if logEntry.Type == meta.FieldChangeIntentType {
			t.Fatalf("disabled switch should not emit field change intents, got %+v", logEntry)
		}
	}

	state := fixtureStore.Get()
	if state.PowerEnabled || state.PowerChanges != 0 || state.LastField != "" {
		t.Fatalf("unexpected disabled switch state after click: %+v", state)
	}
}
