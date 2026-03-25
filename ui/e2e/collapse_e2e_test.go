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
	collapsecomp "github.com/wwsheng009/mint/ui/components/collapse"
)

type collapseFixtureState struct {
	ObservedActiveKeys []string
	FieldChangeCount   int
	LastField          string
	LastValue          string
}

type collapseFixtureMeta struct {
	ID                    string
	ComponentID           string
	ActiveField           string
	FieldChangeIntentType string
}

func newCollapseFixture() (ui.ComponentFunc, func(), func(), *store.Store[collapseFixtureState], collapseFixtureMeta) {
	fixtureStore := store.NewStore(collapseFixtureState{
		ObservedActiveKeys: []string{"overview"},
	})
	meta := collapseFixtureMeta{
		ID:                    "fixture-collapse",
		ComponentID:           "fixture.collapse",
		ActiveField:           "fixture.collapse.active",
		FieldChangeIntentType: runtimeintent.FieldChangeIntent{}.IntentType(),
	}

	unregisters := make([]func(), 0, 1)
	initFn := func() {
		rt := rtui.GetGlobalIntentRuntime()
		if rt == nil {
			return
		}
		unregisters = append(unregisters,
			runtimeintent.RegisterTypedRuntime(rt, func(_ *runtimeintent.ActionContext, i runtimeintent.FieldChangeIntent) runtimeintent.IntentResult {
				if i.Field != meta.ActiveField {
					return runtimeintent.IntentResult{}
				}
				fixtureStore.Update(func(s collapseFixtureState) collapseFixtureState {
					s.ObservedActiveKeys = parseCollapseActiveKeys(i.Value)
					s.FieldChangeCount++
					s.LastField = i.Field
					s.LastValue = i.Value
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
		observedActiveKeys := ui.UseStoreSelector(fixtureStore, func(s collapseFixtureState) []string { return s.ObservedActiveKeys })
		fieldChangeCount := ui.UseStoreSelector(fixtureStore, func(s collapseFixtureState) int { return s.FieldChangeCount })
		lastField := ui.UseStoreSelector(fixtureStore, func(s collapseFixtureState) string { return s.LastField })
		lastValue := ui.UseStoreSelector(fixtureStore, func(s collapseFixtureState) string { return s.LastValue })

		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Collapse E2E Fixture").Build(),
				ui.NewCollapseBuilder().
					SetID(meta.ID).
					ComponentID(meta.ComponentID).
					Item(collapsecomp.Section("Overview Panel", ui.NewTextBuilder("Overview body").Build()).WithKey("overview")).
					Item(collapsecomp.Section("Advanced Panel", ui.NewTextBuilder("Advanced body").Build()).WithKey("advanced")).
					Item(collapsecomp.Section("Locked Panel", ui.NewTextBuilder("Locked body").Build()).WithKey("locked").WithDisabled(true)).
					AccordionMode().
					InitialActiveKeys("overview").
					Width(36).
					ActiveKeysForField(runtimeintent.BindField(meta.ActiveField)).
					Build(),
				ui.NewButtonBuilder("Collapse tail action").SetID("collapse-tail-action").Build(),
				ui.NewTextBuilder(fmt.Sprintf("ObservedActiveKeys: %s", formatCollapseKeys(observedActiveKeys))).Build(),
				ui.NewTextBuilder(fmt.Sprintf("FieldChangeCount: %d", fieldChangeCount)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("LastField: %s", lastField)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("LastValue: %s", formatCollapseValue(lastValue))).Build(),
			})
	}

	return appFn, initFn, cleanupFn, fixtureStore, meta
}

func TestE2ECollapseClickAccordionAndDisabledIgnore(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newCollapseFixture()
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(92, 24), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("Overview body")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("ObservedActiveKeys: overview")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Click(ByText("Advanced Panel")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("Advanced body")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("ObservedActiveKeys: advanced")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("FieldChangeCount: 1")); err != nil {
			return err
		}
		return app.AssertVisible(ByText("LastValue: advanced"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("Overview body")); err == nil {
			return fmt.Errorf("overview body still visible after accordion switch")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "mouse:left:release"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Click(ByText("Locked Panel")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("ObservedActiveKeys: advanced")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("FieldChangeCount: 1")); err != nil {
		t.Fatal(err)
	}
	for _, logEntry := range app.IntentLogs() {
		if logEntry.Type == meta.FieldChangeIntentType {
			t.Fatalf("disabled collapse panel should not emit change intents, got %+v", logEntry)
		}
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Click(ByText("Advanced Panel")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("ObservedActiveKeys: <empty>")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("FieldChangeCount: 2")); err != nil {
			return err
		}
		return app.AssertVisible(ByText("LastValue: <empty>"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("Advanced body")); err == nil {
			return fmt.Errorf("advanced body still visible after collapsing active panel")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	state := fixtureStore.Get()
	if got := formatCollapseKeys(state.ObservedActiveKeys); got != "<empty>" || state.FieldChangeCount != 2 || state.LastValue != "" {
		t.Fatalf("unexpected collapse state after click flow: %+v", state)
	}
}

func TestE2ECollapseKeyboardEnterAndTabSkipDisabled(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newCollapseFixture()
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(92, 24), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Special(platform.KeyEnter); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("ObservedActiveKeys: <empty>")); err != nil {
			return err
		}
		return app.AssertVisible(ByText("FieldChangeCount: 1"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("Overview body")); err == nil {
			return fmt.Errorf("overview body still visible after keyboard collapse")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "key:Enter"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Special(platform.KeyTab); err != nil {
		t.Fatal(err)
	}
	if err := app.Driver().Special(platform.KeyEnter); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("Advanced body")); err != nil {
			return err
		}
		return app.AssertVisible(ByText("ObservedActiveKeys: advanced"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "key:Tab"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "key:Enter"}); err != nil {
		t.Fatal(err)
	}

	app.ClearRawInputs()
	if err := app.Driver().Special(platform.KeyTab); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertFocus(ByID("collapse-tail-action")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "key:Tab"}); err != nil {
		t.Fatal(err)
	}

	state := fixtureStore.Get()
	if got := formatCollapseKeys(state.ObservedActiveKeys); got != "advanced" || state.FieldChangeCount != 2 || state.LastValue != "advanced" {
		t.Fatalf("unexpected collapse state after keyboard flow: %+v", state)
	}
}

func parseCollapseActiveKeys(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	keys := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		keys = append(keys, part)
	}
	return keys
}

func formatCollapseKeys(keys []string) string {
	if len(keys) == 0 {
		return "<empty>"
	}
	return strings.Join(keys, ",")
}

func formatCollapseValue(value string) string {
	if strings.TrimSpace(value) == "" {
		return "<empty>"
	}
	return value
}
