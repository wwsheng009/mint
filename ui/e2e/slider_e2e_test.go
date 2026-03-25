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

type sliderFixtureState struct {
	Volume            int
	VolumeChanges     int
	Brightness        int
	BrightnessChanges int
	LastField         string
	LastValue         string
}

type sliderFixtureMeta struct {
	VolumeField           string
	BrightnessField       string
	FieldChangeIntentType string
}

func newSliderFixture() (ui.ComponentFunc, func(), func(), *store.Store[sliderFixtureState], sliderFixtureMeta) {
	fixtureStore := store.NewStore(sliderFixtureState{
		Volume:     40,
		Brightness: 70,
	})
	meta := sliderFixtureMeta{
		VolumeField:           "fixture.slider.volume",
		BrightnessField:       "fixture.slider.brightness",
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
			case meta.VolumeField:
				nextValue, err := parseFixtureInt(i.Value)
				if err != nil {
					return runtimeintent.ErrorResult(err)
				}
				fixtureStore.Update(func(s sliderFixtureState) sliderFixtureState {
					s.Volume = nextValue
					s.VolumeChanges++
					s.LastField = i.Field
					s.LastValue = i.Value
					return s
				})
				return runtimeintent.HandledResult()
			case meta.BrightnessField:
				nextValue, err := parseFixtureInt(i.Value)
				if err != nil {
					return runtimeintent.ErrorResult(err)
				}
				fixtureStore.Update(func(s sliderFixtureState) sliderFixtureState {
					s.Brightness = nextValue
					s.BrightnessChanges++
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
		volume := ui.UseStoreSelector(fixtureStore, func(s sliderFixtureState) int { return s.Volume })
		volumeChanges := ui.UseStoreSelector(fixtureStore, func(s sliderFixtureState) int { return s.VolumeChanges })
		brightness := ui.UseStoreSelector(fixtureStore, func(s sliderFixtureState) int { return s.Brightness })
		brightnessChanges := ui.UseStoreSelector(fixtureStore, func(s sliderFixtureState) int { return s.BrightnessChanges })
		lastField := ui.UseStoreSelector(fixtureStore, func(s sliderFixtureState) string { return s.LastField })
		lastValue := ui.UseStoreSelector(fixtureStore, func(s sliderFixtureState) string { return s.LastValue })

		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Slider E2E Fixture").Build(),
				ui.NewSliderBuilder().
					SetID("slider-volume").
					Label("Volume").
					Min(0).
					Max(100).
					Step(5).
					Width(12).
					Value(volume).
					ShowValue(true).
					ForField(runtimeintent.BindField(meta.VolumeField)).
					Build(),
				ui.NewSliderBuilder().
					SetID("slider-brightness").
					Label("Brightness").
					Min(0).
					Max(100).
					Step(10).
					Width(12).
					Value(brightness).
					ShowValue(true).
					Disabled(true).
					ForField(runtimeintent.BindField(meta.BrightnessField)).
					Build(),
				ui.NewButtonBuilder("Slider blur target").
					SetID("slider-blur-target").
					Build(),
				ui.NewTextBuilder(fmt.Sprintf("VolumeValue: %d", volume)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("VolumeChanges: %d", volumeChanges)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("BrightnessValue: %d", brightness)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("BrightnessChanges: %d", brightnessChanges)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("LastField: %s", lastField)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("LastValue: %s", lastValue)).Build(),
			})
	}

	return appFn, initFn, cleanupFn, fixtureStore, meta
}

func TestE2ESliderKeyboardAdjustmentsUseRuntimeArrowKeys(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newSliderFixture()
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(96, 18), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertFocus(ByID("slider-volume")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("VolumeValue: 40")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Special(platform.KeyRight); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("VolumeValue: 45"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "key:Right"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("VolumeChanges: 1")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("LastField: fixture.slider.volume")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("LastValue: 45")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	for i := 0; i < 11; i++ {
		if err := app.Driver().Special(platform.KeyRight); err != nil {
			t.Fatal(err)
		}
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("VolumeValue: 100"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "key:Right"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("VolumeChanges: 12")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("LastValue: 100")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Special(platform.KeyRight); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "key:Right"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("VolumeValue: 100")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("VolumeChanges: 12")); err != nil {
		t.Fatal(err)
	}
	for _, logEntry := range app.IntentLogs() {
		if logEntry.Type == meta.FieldChangeIntentType {
			if fieldIntent, ok := logEntry.Intent.(runtimeintent.FieldChangeIntent); ok && fieldIntent.Field == meta.VolumeField {
				t.Fatalf("slider at max should not emit extra volume field changes, got %+v", logEntry)
			}
		}
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Special(platform.KeyLeft); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("VolumeValue: 95"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "key:Left"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("VolumeChanges: 13")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("LastValue: 95")); err != nil {
		t.Fatal(err)
	}

	state := fixtureStore.Get()
	if state.Volume != 95 || state.VolumeChanges != 13 || state.LastField != meta.VolumeField || state.LastValue != "95" {
		t.Fatalf("unexpected slider fixture state after keyboard flow: %+v", state)
	}
}

func TestE2ESliderDisabledSkipsInteraction(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newSliderFixture()
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(96, 18), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.Driver().Special(platform.KeyTab); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitFocus(ByID("slider-blur-target"), 500*time.Millisecond); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Click(ByID("slider-brightness")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "mouse:left:release"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitFocus(ByID("slider-blur-target"), 500*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("BrightnessValue: 70")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("BrightnessChanges: 0")); err != nil {
		t.Fatal(err)
	}
	for _, logEntry := range app.IntentLogs() {
		if logEntry.Type == meta.FieldChangeIntentType {
			if fieldIntent, ok := logEntry.Intent.(runtimeintent.FieldChangeIntent); ok && fieldIntent.Field == meta.BrightnessField {
				t.Fatalf("disabled slider should not emit brightness field changes, got %+v", logEntry)
			}
		}
	}

	state := fixtureStore.Get()
	if state.Brightness != 70 || state.BrightnessChanges != 0 || state.LastField != "" || state.LastValue != "" {
		t.Fatalf("unexpected disabled slider fixture state: %+v", state)
	}
}
