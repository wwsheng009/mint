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

type radioFixtureState struct {
	NewsletterEnabled bool
	NewsletterChanges int
	AlertsEnabled     bool
	AlertsChanges     int
	SelectedFrequency string
	FrequencyChanges  int
	LastField         string
	LastValue         string
}

type radioFixtureMeta struct {
	NewsletterField       string
	AlertsField           string
	FrequencyField        string
	FieldChangeIntentType string
}

func newRadioFixture() (ui.ComponentFunc, func(), func(), *store.Store[radioFixtureState], radioFixtureMeta) {
	fixtureStore := store.NewStore(radioFixtureState{
		SelectedFrequency: "weekly",
	})
	meta := radioFixtureMeta{
		NewsletterField:       "fixture.radio.newsletter",
		AlertsField:           "fixture.radio.alerts",
		FrequencyField:        "fixture.radio.frequency",
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
			case meta.NewsletterField:
				fixtureStore.Update(func(s radioFixtureState) radioFixtureState {
					s.NewsletterEnabled = i.Value == "true"
					s.NewsletterChanges++
					s.LastField = i.Field
					s.LastValue = i.Value
					return s
				})
				return runtimeintent.HandledResult()
			case meta.AlertsField:
				fixtureStore.Update(func(s radioFixtureState) radioFixtureState {
					s.AlertsEnabled = i.Value == "true"
					s.AlertsChanges++
					s.LastField = i.Field
					s.LastValue = i.Value
					return s
				})
				return runtimeintent.HandledResult()
			case meta.FrequencyField:
				fixtureStore.Update(func(s radioFixtureState) radioFixtureState {
					s.SelectedFrequency = i.Value
					s.FrequencyChanges++
					s.LastField = i.Field
					s.LastValue = i.Value
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
		newsletterEnabled := ui.UseStoreSelector(fixtureStore, func(s radioFixtureState) bool { return s.NewsletterEnabled })
		newsletterChanges := ui.UseStoreSelector(fixtureStore, func(s radioFixtureState) int { return s.NewsletterChanges })
		alertsEnabled := ui.UseStoreSelector(fixtureStore, func(s radioFixtureState) bool { return s.AlertsEnabled })
		alertsChanges := ui.UseStoreSelector(fixtureStore, func(s radioFixtureState) int { return s.AlertsChanges })
		selectedFrequency := ui.UseStoreSelector(fixtureStore, func(s radioFixtureState) string { return s.SelectedFrequency })
		frequencyChanges := ui.UseStoreSelector(fixtureStore, func(s radioFixtureState) int { return s.FrequencyChanges })
		lastField := ui.UseStoreSelector(fixtureStore, func(s radioFixtureState) string { return s.LastField })
		lastValue := ui.UseStoreSelector(fixtureStore, func(s radioFixtureState) string { return s.LastValue })

		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Radio E2E Fixture").Build(),
				ui.NewRadioBuilder().
					SetID("radio-newsletter").
					Label("Newsletter opt-in").
					Checked(newsletterEnabled).
					ForField(runtimeintent.BindField(meta.NewsletterField)).
					Build(),
				ui.NewRadioBuilder().
					SetID("radio-alerts").
					Label("Product alerts").
					Checked(alertsEnabled).
					Disabled(true).
					ForField(runtimeintent.BindField(meta.AlertsField)).
					Build(),
				ui.NewRadioGroupBuilder([]ui.RadioOption{
					ui.NewRadioOption("daily", "Daily"),
					ui.NewRadioOption("weekly", "Weekly"),
					ui.NewRadioOption("monthly", "Monthly"),
				}).
					SetID("radio-frequency").
					Label("Digest frequency").
					Selected(selectedFrequency).
					ForField(runtimeintent.BindField(meta.FrequencyField)).
					Vertical().
					Build(),
				ui.NewTextBuilder(fmt.Sprintf("NewsletterEnabled: %t", newsletterEnabled)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("NewsletterChanges: %d", newsletterChanges)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("AlertsEnabled: %t", alertsEnabled)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("AlertsChanges: %d", alertsChanges)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("SelectedFrequency: %s", selectedFrequency)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("FrequencyChanges: %d", frequencyChanges)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("LastField: %s", lastField)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("LastValue: %s", lastValue)).Build(),
			})
	}

	return appFn, initFn, cleanupFn, fixtureStore, meta
}

func TestE2ERadioStandaloneSelectOnlyOnceAndDisabledIgnore(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newRadioFixture()
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(88, 20), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertFocus(ByID("radio-newsletter")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("NewsletterEnabled: false")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("AlertsEnabled: false")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Special(platform.KeyEnter); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("NewsletterEnabled: true"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "key:Enter"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("(*) Newsletter opt-in")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("NewsletterChanges: 1")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("LastField: fixture.radio.newsletter")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("LastValue: true")); err != nil {
		t.Fatal(err)
	}

	newsletterPoint, err := app.ResolvePoint(ByText("Newsletter opt-in"))
	if err != nil {
		t.Fatal(err)
	}
	if err := app.AssertHit(newsletterPoint, ByID("radio-newsletter")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Click(ByID("radio-newsletter")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "mouse:left:release"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("NewsletterEnabled: true")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("NewsletterChanges: 1")); err != nil {
		t.Fatal(err)
	}
	for _, logEntry := range app.IntentLogs() {
		if logEntry.Type == meta.FieldChangeIntentType {
			if fieldIntent, ok := logEntry.Intent.(runtimeintent.FieldChangeIntent); ok && fieldIntent.Field == meta.NewsletterField {
				t.Fatalf("checked standalone radio should not emit duplicate field changes, got %+v", logEntry)
			}
		}
	}

	alertsPoint, err := app.ResolvePoint(ByText("Product alerts"))
	if err != nil {
		t.Fatal(err)
	}
	if err := app.AssertHit(alertsPoint, ByID("radio-alerts")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Click(ByID("radio-alerts")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "mouse:left:release"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("AlertsEnabled: false")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("AlertsChanges: 0")); err != nil {
		t.Fatal(err)
	}
	for _, logEntry := range app.IntentLogs() {
		if logEntry.Type == meta.FieldChangeIntentType {
			if fieldIntent, ok := logEntry.Intent.(runtimeintent.FieldChangeIntent); ok && fieldIntent.Field == meta.AlertsField {
				t.Fatalf("disabled radio should not emit alerts field changes, got %+v", logEntry)
			}
		}
	}

	state := fixtureStore.Get()
	if !state.NewsletterEnabled || state.NewsletterChanges != 1 || state.AlertsEnabled || state.AlertsChanges != 0 || state.LastField != meta.NewsletterField || state.LastValue != "true" {
		t.Fatalf("unexpected radio fixture state after standalone flow: %+v", state)
	}
}

func TestE2ERadioGroupClickFlowUpdatesSelectedFieldValue(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newRadioFixture()
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(88, 20), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("SelectedFrequency: weekly")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("(*) Weekly")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("( ) Monthly")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Click(ByText("Monthly")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("SelectedFrequency: monthly"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "mouse:left:release"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("( ) Weekly")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("(*) Monthly")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("FrequencyChanges: 1")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("LastField: fixture.radio.frequency")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("LastValue: monthly")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Click(ByText("Daily")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("SelectedFrequency: daily"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "mouse:left:release"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("(*) Daily")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("( ) Monthly")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("FrequencyChanges: 2")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("LastValue: daily")); err != nil {
		t.Fatal(err)
	}

	state := fixtureStore.Get()
	if state.SelectedFrequency != "daily" || state.FrequencyChanges != 2 || state.LastField != meta.FrequencyField || state.LastValue != "daily" {
		t.Fatalf("unexpected radio fixture state after group flow: %+v", state)
	}
}
