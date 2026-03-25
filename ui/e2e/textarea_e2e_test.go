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

type textareaFixtureState struct {
	NotesValue    string
	NotesChanges  int
	LockedValue   string
	LockedChanges int
	LastField     string
	LastValue     string
}

type textareaFixtureMeta struct {
	NotesField            string
	LockedField           string
	FieldChangeIntentType string
}

func newTextareaFixture() (ui.ComponentFunc, func(), func(), *store.Store[textareaFixtureState], textareaFixtureMeta) {
	fixtureStore := store.NewStore(textareaFixtureState{
		LockedValue: "locked",
	})
	meta := textareaFixtureMeta{
		NotesField:            "fixture.textarea.notes",
		LockedField:           "fixture.textarea.locked",
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
			case meta.NotesField:
				fixtureStore.Update(func(s textareaFixtureState) textareaFixtureState {
					s.NotesValue = i.Value
					s.NotesChanges++
					s.LastField = i.Field
					s.LastValue = i.Value
					return s
				})
				return runtimeintent.HandledResult()
			case meta.LockedField:
				fixtureStore.Update(func(s textareaFixtureState) textareaFixtureState {
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
		notesValue := ui.UseStoreSelector(fixtureStore, func(s textareaFixtureState) string { return s.NotesValue })
		notesChanges := ui.UseStoreSelector(fixtureStore, func(s textareaFixtureState) int { return s.NotesChanges })
		lockedValue := ui.UseStoreSelector(fixtureStore, func(s textareaFixtureState) string { return s.LockedValue })
		lockedChanges := ui.UseStoreSelector(fixtureStore, func(s textareaFixtureState) int { return s.LockedChanges })
		lastField := ui.UseStoreSelector(fixtureStore, func(s textareaFixtureState) string { return s.LastField })
		lastValue := ui.UseStoreSelector(fixtureStore, func(s textareaFixtureState) string { return s.LastValue })

		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Textarea E2E Fixture").Build(),
				ui.NewTextareaBuilder().
					SetID("textarea-notes").
					Placeholder("Write notes").
					Value(notesValue).
					Rows(3).
					Cols(12).
					ForField(runtimeintent.BindField(meta.NotesField)).
					Build(),
				ui.NewTextareaBuilder().
					SetID("textarea-locked").
					Value(lockedValue).
					Rows(2).
					Cols(12).
					Disabled(true).
					ForField(runtimeintent.BindField(meta.LockedField)).
					Build(),
				ui.NewButtonBuilder("Textarea blur target").
					SetID("textarea-blur-target").
					Build(),
				ui.NewTextBuilder(fmt.Sprintf("NotesValue: %s", formatTextareaValue(notesValue))).Build(),
				ui.NewTextBuilder(fmt.Sprintf("NotesChanges: %d", notesChanges)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("LockedValue: %s", formatTextareaValue(lockedValue))).Build(),
				ui.NewTextBuilder(fmt.Sprintf("LockedChanges: %d", lockedChanges)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("LastField: %s", lastField)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("LastValue: %s", formatTextareaValue(lastValue))).Build(),
			})
	}

	return appFn, initFn, cleanupFn, fixtureStore, meta
}

func TestE2ETextareaMultilineEditingAndVerticalMovement(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newTextareaFixture()
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(96, 22), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertFocus(ByID("textarea-notes")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("NotesValue: <empty>")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	for _, key := range []rune{'a', 'b'} {
		if err := app.Driver().Key(key); err != nil {
			t.Fatal(err)
		}
	}
	if err := app.Driver().Special(platform.KeyEnter); err != nil {
		t.Fatal(err)
	}
	for _, key := range []rune{'c', 'd'} {
		if err := app.Driver().Key(key); err != nil {
			t.Fatal(err)
		}
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("NotesValue: ab\\ncd"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "key:a"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "key:Enter"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("NotesChanges: 5")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("LastField: fixture.textarea.notes")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("LastValue: ab\\ncd")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Special(platform.KeyUp); err != nil {
		t.Fatal(err)
	}
	if err := app.Driver().Key('X'); err != nil {
		t.Fatal(err)
	}
	if err := app.Driver().Special(platform.KeyDown); err != nil {
		t.Fatal(err)
	}
	if err := app.Driver().Key('Y'); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("NotesValue: abX\\ncdY"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "key:Up"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "key:Down"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("NotesChanges: 7")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("LastValue: abX\\ncdY")); err != nil {
		t.Fatal(err)
	}

	state := fixtureStore.Get()
	if state.NotesValue != "abX\ncdY" || state.NotesChanges != 7 || state.LastField != meta.NotesField || state.LastValue != "abX\ncdY" {
		t.Fatalf("unexpected textarea fixture state after multiline flow: %+v", state)
	}
}

func TestE2ETextareaDisabledIgnoresClickAndTyping(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newTextareaFixture()
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(96, 22), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.Driver().Click(ByID("textarea-blur-target")); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitFocus(ByID("textarea-blur-target"), 500*time.Millisecond); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Click(ByID("textarea-locked")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "mouse:left:release"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitFocus(ByID("textarea-blur-target"), 500*time.Millisecond); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Key('z'); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "key:z"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("LockedValue: locked")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("LockedChanges: 0")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("NotesValue: <empty>")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("NotesChanges: 0")); err != nil {
		t.Fatal(err)
	}
	for _, logEntry := range app.IntentLogs() {
		if logEntry.Type == meta.FieldChangeIntentType {
			if fieldIntent, ok := logEntry.Intent.(runtimeintent.FieldChangeIntent); ok && (fieldIntent.Field == meta.LockedField || fieldIntent.Field == meta.NotesField) {
				t.Fatalf("disabled textarea flow should not emit textarea field changes, got %+v", logEntry)
			}
		}
	}

	state := fixtureStore.Get()
	if state.LockedValue != "locked" || state.LockedChanges != 0 || state.NotesValue != "" || state.NotesChanges != 0 || state.LastField != "" || state.LastValue != "" {
		t.Fatalf("unexpected textarea fixture state after disabled flow: %+v", state)
	}
}

func formatTextareaValue(value string) string {
	if value == "" {
		return "<empty>"
	}
	return strings.ReplaceAll(value, "\n", "\\n")
}
