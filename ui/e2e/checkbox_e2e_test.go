package e2e

import (
	"fmt"
	"strings"
	"testing"
	"time"

	runtimeintent "github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/runtime/store"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
)

type checkboxFixtureState struct {
	TermsAccepted     bool
	TermsChanges      int
	MarketingEnabled  bool
	MarketingChanges  int
	SelectedInterests []string
	InterestsChanges  int
	LastField         string
	LastValue         string
}

type checkboxFixtureMeta struct {
	TermsField            string
	MarketingField        string
	InterestsField        string
	FieldChangeIntentType string
}

func newCheckboxFixture() (ui.ComponentFunc, func(), func(), *store.Store[checkboxFixtureState], checkboxFixtureMeta) {
	fixtureStore := store.NewStore(checkboxFixtureState{
		MarketingEnabled: true,
	})
	meta := checkboxFixtureMeta{
		TermsField:            "fixture.checkbox.terms",
		MarketingField:        "fixture.checkbox.marketing",
		InterestsField:        "fixture.checkbox.interests",
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
			case meta.TermsField:
				fixtureStore.Update(func(s checkboxFixtureState) checkboxFixtureState {
					s.TermsAccepted = i.Value == "true"
					s.TermsChanges++
					s.LastField = i.Field
					s.LastValue = i.Value
					return s
				})
				return runtimeintent.HandledResult()
			case meta.MarketingField:
				fixtureStore.Update(func(s checkboxFixtureState) checkboxFixtureState {
					s.MarketingEnabled = i.Value == "true"
					s.MarketingChanges++
					s.LastField = i.Field
					s.LastValue = i.Value
					return s
				})
				return runtimeintent.HandledResult()
			case meta.InterestsField:
				fixtureStore.Update(func(s checkboxFixtureState) checkboxFixtureState {
					s.SelectedInterests = parseCheckboxValues(i.Value)
					s.InterestsChanges++
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
		termsAccepted := ui.UseStoreSelector(fixtureStore, func(s checkboxFixtureState) bool { return s.TermsAccepted })
		termsChanges := ui.UseStoreSelector(fixtureStore, func(s checkboxFixtureState) int { return s.TermsChanges })
		marketingEnabled := ui.UseStoreSelector(fixtureStore, func(s checkboxFixtureState) bool { return s.MarketingEnabled })
		marketingChanges := ui.UseStoreSelector(fixtureStore, func(s checkboxFixtureState) int { return s.MarketingChanges })
		selectedInterests := ui.UseStoreSelector(fixtureStore, func(s checkboxFixtureState) []string { return s.SelectedInterests })
		interestsChanges := ui.UseStoreSelector(fixtureStore, func(s checkboxFixtureState) int { return s.InterestsChanges })
		lastField := ui.UseStoreSelector(fixtureStore, func(s checkboxFixtureState) string { return s.LastField })
		lastValue := ui.UseStoreSelector(fixtureStore, func(s checkboxFixtureState) string { return s.LastValue })

		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Checkbox E2E Fixture").Build(),
				ui.NewCheckboxBuilder().
					SetID("checkbox-terms").
					Label("Accept terms").
					Checked(termsAccepted).
					ForField(runtimeintent.BindField(meta.TermsField)).
					Build(),
				ui.NewCheckboxBuilder().
					SetID("checkbox-marketing").
					Label("Marketing emails").
					Checked(marketingEnabled).
					Disabled(true).
					ForField(runtimeintent.BindField(meta.MarketingField)).
					Build(),
				ui.NewCheckboxGroupBuilder([]ui.CheckboxOption{
					ui.NewCheckboxOption("dev", "Development"),
					ui.NewCheckboxOption("design", "Design"),
					ui.NewCheckboxOption("docs", "Documentation"),
				}).
					SetID("checkbox-interests").
					Label("Interests").
					Selecteds(selectedInterests).
					ForField(runtimeintent.BindField(meta.InterestsField)).
					Vertical().
					Build(),
				ui.NewTextBuilder(fmt.Sprintf("TermsAccepted: %t", termsAccepted)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("TermsChanges: %d", termsChanges)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("MarketingEnabled: %t", marketingEnabled)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("MarketingChanges: %d", marketingChanges)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("SelectedInterests: %s", formatCheckboxValues(selectedInterests))).Build(),
				ui.NewTextBuilder(fmt.Sprintf("InterestsChanges: %d", interestsChanges)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("LastField: %s", lastField)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("LastValue: %s", lastValue)).Build(),
			})
	}

	return appFn, initFn, cleanupFn, fixtureStore, meta
}

func TestE2ECheckboxStandaloneKeyboardClickAndDisabledIgnore(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newCheckboxFixture()
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(88, 20), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertFocus(ByID("checkbox-terms")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("TermsAccepted: false")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("MarketingEnabled: true")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Special(platform.KeyEnter); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("TermsAccepted: true"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "key:Enter"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("[X] Accept terms")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("TermsChanges: 1")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("LastField: fixture.checkbox.terms")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("LastValue: true")); err != nil {
		t.Fatal(err)
	}

	termsPoint, err := app.ResolvePoint(ByText("Accept terms"))
	if err != nil {
		t.Fatal(err)
	}
	if err := app.AssertHit(termsPoint, ByID("checkbox-terms")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Click(ByID("checkbox-terms")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("TermsAccepted: false"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "mouse:left:release"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("[ ] Accept terms")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("TermsChanges: 2")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("LastValue: false")); err != nil {
		t.Fatal(err)
	}

	marketingPoint, err := app.ResolvePoint(ByText("Marketing emails"))
	if err != nil {
		t.Fatal(err)
	}
	if err := app.AssertHit(marketingPoint, ByID("checkbox-marketing")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Click(ByID("checkbox-marketing")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "mouse:left:release"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("MarketingEnabled: true")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("MarketingChanges: 0")); err != nil {
		t.Fatal(err)
	}
	for _, logEntry := range app.IntentLogs() {
		if logEntry.Type == meta.FieldChangeIntentType {
			if fieldIntent, ok := logEntry.Intent.(runtimeintent.FieldChangeIntent); ok && fieldIntent.Field == meta.MarketingField {
				t.Fatalf("disabled checkbox should not emit marketing field changes, got %+v", logEntry)
			}
		}
	}

	state := fixtureStore.Get()
	if state.TermsAccepted || state.TermsChanges != 2 || state.MarketingChanges != 0 || !state.MarketingEnabled || state.LastField != meta.TermsField || state.LastValue != "false" {
		t.Fatalf("unexpected checkbox fixture state after standalone flow: %+v", state)
	}
}

func TestE2ECheckboxGroupClickFlowEmitsCommaSeparatedFieldValue(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newCheckboxFixture()
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(88, 20), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("SelectedInterests: <none>")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("[ ] Development")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("[ ] Design")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Click(ByText("Development")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("SelectedInterests: dev"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "mouse:left:release"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("[X] Development")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("InterestsChanges: 1")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("LastField: fixture.checkbox.interests")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("LastValue: dev")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Click(ByText("Design")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("SelectedInterests: dev,design"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "mouse:left:release"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("[X] Design")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("InterestsChanges: 2")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("LastValue: dev,design")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Click(ByText("Development")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("SelectedInterests: design"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "mouse:left:release"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("[ ] Development")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("[X] Design")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("InterestsChanges: 3")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("LastValue: design")); err != nil {
		t.Fatal(err)
	}

	state := fixtureStore.Get()
	if !sameCheckboxValues(state.SelectedInterests, []string{"design"}) || state.InterestsChanges != 3 || state.LastField != meta.InterestsField || state.LastValue != "design" {
		t.Fatalf("unexpected checkbox fixture state after group flow: %+v", state)
	}
}

func parseCheckboxValues(value string) []string {
	if value == "" {
		return nil
	}
	return strings.Split(value, ",")
}

func formatCheckboxValues(values []string) string {
	if len(values) == 0 {
		return "<none>"
	}
	return strings.Join(values, ",")
}

func sameCheckboxValues(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
