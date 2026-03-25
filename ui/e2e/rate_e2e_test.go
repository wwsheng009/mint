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

type rateFixtureState struct {
	Satisfaction        int
	SatisfactionChanges int
	Quality             int
	QualityChanges      int
	LastField           string
	LastValue           string
}

type rateFixtureMeta struct {
	SatisfactionField     string
	QualityField          string
	FieldChangeIntentType string
}

func newRateFixture() (ui.ComponentFunc, func(), func(), *store.Store[rateFixtureState], rateFixtureMeta) {
	fixtureStore := store.NewStore(rateFixtureState{
		Satisfaction: 4,
		Quality:      3,
	})
	meta := rateFixtureMeta{
		SatisfactionField:     "fixture.rate.satisfaction",
		QualityField:          "fixture.rate.quality",
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
			case meta.SatisfactionField:
				nextValue, err := parseFixtureInt(i.Value)
				if err != nil {
					return runtimeintent.ErrorResult(err)
				}
				fixtureStore.Update(func(s rateFixtureState) rateFixtureState {
					s.Satisfaction = nextValue
					s.SatisfactionChanges++
					s.LastField = i.Field
					s.LastValue = i.Value
					return s
				})
				return runtimeintent.HandledResult()
			case meta.QualityField:
				nextValue, err := parseFixtureInt(i.Value)
				if err != nil {
					return runtimeintent.ErrorResult(err)
				}
				fixtureStore.Update(func(s rateFixtureState) rateFixtureState {
					s.Quality = nextValue
					s.QualityChanges++
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
		satisfaction := ui.UseStoreSelector(fixtureStore, func(s rateFixtureState) int { return s.Satisfaction })
		satisfactionChanges := ui.UseStoreSelector(fixtureStore, func(s rateFixtureState) int { return s.SatisfactionChanges })
		quality := ui.UseStoreSelector(fixtureStore, func(s rateFixtureState) int { return s.Quality })
		qualityChanges := ui.UseStoreSelector(fixtureStore, func(s rateFixtureState) int { return s.QualityChanges })
		lastField := ui.UseStoreSelector(fixtureStore, func(s rateFixtureState) string { return s.LastField })
		lastValue := ui.UseStoreSelector(fixtureStore, func(s rateFixtureState) string { return s.LastValue })

		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Rate E2E Fixture").Build(),
				ui.NewRateBuilder().
					SetID("rate-satisfaction").
					Label("Satisfaction").
					Count(5).
					Value(satisfaction).
					AllowClear(true).
					ShowValue(true).
					ForField(runtimeintent.BindField(meta.SatisfactionField)).
					Build(),
				ui.NewRateBuilder().
					SetID("rate-quality").
					Label("Quality").
					Count(5).
					Value(quality).
					AllowClear(true).
					ShowValue(true).
					Disabled(true).
					ForField(runtimeintent.BindField(meta.QualityField)).
					Build(),
				ui.NewButtonBuilder("Rate blur target").
					SetID("rate-blur-target").
					Build(),
				ui.NewTextBuilder(fmt.Sprintf("SatisfactionValue: %d", satisfaction)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("SatisfactionChanges: %d", satisfactionChanges)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("QualityValue: %d", quality)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("QualityChanges: %d", qualityChanges)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("LastField: %s", lastField)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("LastValue: %s", lastValue)).Build(),
			})
	}

	return appFn, initFn, cleanupFn, fixtureStore, meta
}

func TestE2ERateKeyboardClickAndClearFlow(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newRateFixture()
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(96, 18), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertFocus(ByID("rate-satisfaction")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("SatisfactionValue: 4")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Special(platform.KeyRight); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("SatisfactionValue: 5"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "key:Right"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("SatisfactionChanges: 1")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("LastField: fixture.rate.satisfaction")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("LastValue: 5")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Special(platform.KeyEnter); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("SatisfactionValue: 0"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "key:Enter"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("SatisfactionChanges: 2")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("LastValue: 0")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Click(ByID("rate-satisfaction")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("SatisfactionValue: 1"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "mouse:left:release"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("SatisfactionChanges: 3")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("LastValue: 1")); err != nil {
		t.Fatal(err)
	}

	state := fixtureStore.Get()
	if state.Satisfaction != 1 || state.SatisfactionChanges != 3 || state.LastField != meta.SatisfactionField || state.LastValue != "1" {
		t.Fatalf("unexpected rate fixture state after interaction flow: %+v", state)
	}
}

func TestE2ERateDisabledIgnoresClick(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newRateFixture()
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(96, 18), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.Driver().Click(ByID("rate-blur-target")); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitFocus(ByID("rate-blur-target"), 500*time.Millisecond); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Click(ByID("rate-quality")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "mouse:left:release"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitFocus(ByID("rate-blur-target"), 500*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("QualityValue: 3")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("QualityChanges: 0")); err != nil {
		t.Fatal(err)
	}
	for _, logEntry := range app.IntentLogs() {
		if logEntry.Type == meta.FieldChangeIntentType {
			if fieldIntent, ok := logEntry.Intent.(runtimeintent.FieldChangeIntent); ok && fieldIntent.Field == meta.QualityField {
				t.Fatalf("disabled rate should not emit quality field changes, got %+v", logEntry)
			}
		}
	}

	state := fixtureStore.Get()
	if state.Quality != 3 || state.QualityChanges != 0 || state.LastField != "" || state.LastValue != "" {
		t.Fatalf("unexpected disabled rate fixture state: %+v", state)
	}
}
