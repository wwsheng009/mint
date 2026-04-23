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

type inputFixtureState struct {
	NameValue     string
	NameChanges   int
	AmountValue   string
	AmountChanges int
	LockedValue   string
	LockedChanges int
	LastField     string
	LastValue     string
}

type inputFixtureMeta struct {
	NameField             string
	AmountField           string
	LockedField           string
	FieldChangeIntentType string
}

func newInputFixture() (ui.ComponentFunc, func(), func(), *store.Store[inputFixtureState], inputFixtureMeta) {
	fixtureStore := store.NewStore(inputFixtureState{
		LockedValue: "locked",
	})
	meta := inputFixtureMeta{
		NameField:             "fixture.input.name",
		AmountField:           "fixture.input.amount",
		LockedField:           "fixture.input.locked",
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
			case meta.NameField:
				fixtureStore.Update(func(s inputFixtureState) inputFixtureState {
					s.NameValue = i.Value
					s.NameChanges++
					s.LastField = i.Field
					s.LastValue = i.Value
					return s
				})
				return runtimeintent.HandledResult()
			case meta.AmountField:
				fixtureStore.Update(func(s inputFixtureState) inputFixtureState {
					s.AmountValue = i.Value
					s.AmountChanges++
					s.LastField = i.Field
					s.LastValue = i.Value
					return s
				})
				return runtimeintent.HandledResult()
			case meta.LockedField:
				fixtureStore.Update(func(s inputFixtureState) inputFixtureState {
					s.LockedValue = i.Value
					s.LockedChanges++
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
		nameValue := ui.UseStoreSelector(fixtureStore, func(s inputFixtureState) string { return s.NameValue })
		nameChanges := ui.UseStoreSelector(fixtureStore, func(s inputFixtureState) int { return s.NameChanges })
		amountValue := ui.UseStoreSelector(fixtureStore, func(s inputFixtureState) string { return s.AmountValue })
		amountChanges := ui.UseStoreSelector(fixtureStore, func(s inputFixtureState) int { return s.AmountChanges })
		lockedValue := ui.UseStoreSelector(fixtureStore, func(s inputFixtureState) string { return s.LockedValue })
		lockedChanges := ui.UseStoreSelector(fixtureStore, func(s inputFixtureState) int { return s.LockedChanges })
		lastField := ui.UseStoreSelector(fixtureStore, func(s inputFixtureState) string { return s.LastField })
		lastValue := ui.UseStoreSelector(fixtureStore, func(s inputFixtureState) string { return s.LastValue })

		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Input E2E Fixture").Build(),
				ui.NewInputBuilder().
					SetID("input-name").
					Placeholder("Your name").
					Value(nameValue).
					Width(16).
					ForField(runtimeintent.BindField(meta.NameField)).
					Build(),
				ui.NewInputBuilder().
					SetID("input-amount").
					Placeholder("Amount").
					Type(ui.InputNumber).
					AllowDecimal(false).
					Min(0).
					Max(10).
					Step(5).
					Value(amountValue).
					Width(10).
					ForField(runtimeintent.BindField(meta.AmountField)).
					Build(),
				ui.NewInputBuilder().
					SetID("input-locked").
					Value(lockedValue).
					ReadOnly(true).
					Width(12).
					ForField(runtimeintent.BindField(meta.LockedField)).
					Build(),
				ui.NewButtonBuilder("Blur target").
					SetID("input-blur-target").
					Build(),
				ui.NewTextBuilder(fmt.Sprintf("NameValue: %s", formatInputValue(nameValue))).Build(),
				ui.NewTextBuilder(fmt.Sprintf("NameChanges: %d", nameChanges)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("AmountValue: %s", formatInputValue(amountValue))).Build(),
				ui.NewTextBuilder(fmt.Sprintf("AmountChanges: %d", amountChanges)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("LockedValue: %s", formatInputValue(lockedValue))).Build(),
				ui.NewTextBuilder(fmt.Sprintf("LockedChanges: %d", lockedChanges)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("LastField: %s", lastField)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("LastValue: %s", formatInputValue(lastValue))).Build(),
			})
	}

	return appFn, initFn, cleanupFn, fixtureStore, meta
}

func TestE2EInputTextTypingBackspaceAndReadOnlyIgnore(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newInputFixture()
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(96, 20), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertFocus(ByID("input-name")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("NameValue: <empty>")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	for _, key := range []rune{'a', 'b', 'c', 'd'} {
		if err := app.Driver().Key(key); err != nil {
			t.Fatal(err)
		}
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("NameValue: abcd"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "key:a"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("NameChanges: 4")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("LastField: fixture.input.name")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("LastValue: abcd")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Special(platform.KeyBackspace); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("NameValue: abc"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "key:Backspace"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("NameChanges: 5")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("LastValue: abc")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Click(ByID("input-locked")); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitFocus(ByID("input-locked"), 500*time.Millisecond); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Type("x"); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "key:x"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("LockedValue: locked")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("LockedChanges: 0")); err != nil {
		t.Fatal(err)
	}
	for _, logEntry := range app.IntentLogs() {
		if logEntry.Type == meta.FieldChangeIntentType {
			if fieldIntent, ok := logEntry.Intent.(runtimeintent.FieldChangeIntent); ok && fieldIntent.Field == meta.LockedField {
				t.Fatalf("readOnly input should not emit locked field changes, got %+v", logEntry)
			}
		}
	}

	state := fixtureStore.Get()
	if state.NameValue != "abc" || state.NameChanges != 5 || state.LockedValue != "locked" || state.LockedChanges != 0 || state.LastField != meta.NameField || state.LastValue != "abc" {
		t.Fatalf("unexpected input fixture state after text flow: %+v", state)
	}
}

func TestE2EInputNumberStepAndBlurNormalization(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newInputFixture()
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(96, 20), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.Driver().Click(ByID("input-amount")); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitFocus(ByID("input-amount"), 500*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("AmountValue: <empty>")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	for _, key := range []rune{'0', '5'} {
		if err := app.Driver().Key(key); err != nil {
			t.Fatal(err)
		}
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("AmountValue: 05"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "key:0"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("AmountChanges: 2")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("LastField: fixture.input.amount")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("LastValue: 05")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Click(ByID("input-blur-target")); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitFocus(ByID("input-blur-target"), 500*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("AmountValue: 5"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("AmountChanges: 3")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("LastValue: 5")); err != nil {
		t.Fatal(err)
	}

	if err := app.Driver().Click(ByID("input-amount")); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitFocus(ByID("input-amount"), 500*time.Millisecond); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Special(platform.KeyUp); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("AmountValue: 10"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "key:Up"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("AmountChanges: 4")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("LastValue: 10")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Special(platform.KeyUp); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "key:Up"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("AmountValue: 10")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("AmountChanges: 4")); err != nil {
		t.Fatal(err)
	}
	for _, logEntry := range app.IntentLogs() {
		if logEntry.Type == meta.FieldChangeIntentType {
			if fieldIntent, ok := logEntry.Intent.(runtimeintent.FieldChangeIntent); ok && fieldIntent.Field == meta.AmountField {
				t.Fatalf("number input at max should not emit extra amount field changes, got %+v", logEntry)
			}
		}
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Special(platform.KeyDown); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("AmountValue: 5"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "key:Down"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("AmountChanges: 5")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("LastValue: 5")); err != nil {
		t.Fatal(err)
	}

	state := fixtureStore.Get()
	if state.AmountValue != "5" || state.AmountChanges != 5 || state.LastField != meta.AmountField || state.LastValue != "5" {
		t.Fatalf("unexpected input fixture state after number flow: %+v", state)
	}
}

func formatInputValue(value string) string {
	if value == "" {
		return "<empty>"
	}
	return value
}
