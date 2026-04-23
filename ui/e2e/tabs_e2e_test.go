package e2e

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	runtimeintent "github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/runtime/store"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
	tabscomp "github.com/wwsheng009/mint/ui/components/tabs"
)

type tabsFixtureState struct {
	ObservedCurrent  int
	FieldChangeCount int
	TabChangeCount   int
	LastTabID        string
	LastTabLabel     string
}

type tabsFixtureMeta struct {
	ID                    string
	ComponentID           string
	CurrentField          string
	FieldChangeIntentType string
	TabChangeIntentType   string
}

func newTabsFixture(id, componentID string, items []tabscomp.TabItem) (ui.ComponentFunc, func(), func(), *store.Store[tabsFixtureState], tabsFixtureMeta) {
	fixtureStore := store.NewStore(tabsFixtureState{})
	meta := tabsFixtureMeta{
		ID:                    id,
		ComponentID:           componentID,
		CurrentField:          componentID + ".current",
		FieldChangeIntentType: runtimeintent.FieldChangeIntent{}.IntentType(),
		TabChangeIntentType:   tabscomp.TabChangeIntent{}.IntentType(),
	}

	unregisters := make([]func(), 0, 2)
	initFn := func() {
		rt := rtui.GetGlobalIntentRuntime()
		if rt == nil {
			return
		}
		unregisters = append(unregisters,
			runtimeintent.RegisterTypedRuntime(rt, func(_ *runtimeintent.ActionContext, i runtimeintent.FieldChangeIntent) runtimeintent.IntentResult {
				if i.Field != meta.CurrentField {
					return runtimeintent.IntentResult{}
				}
				nextIndex, err := strconv.Atoi(i.Value)
				if err != nil {
					return runtimeintent.ErrorResult(err)
				}
				if nextIndex < 0 || nextIndex >= len(items) {
					return runtimeintent.ErrorResult(fmt.Errorf("tabs field index %d out of range", nextIndex))
				}
				fixtureStore.Update(func(s tabsFixtureState) tabsFixtureState {
					s.ObservedCurrent = nextIndex
					s.FieldChangeCount++
					return s
				})
				return runtimeintent.HandledResult()
			}),
			runtimeintent.RegisterTypedRuntime(rt, func(_ *runtimeintent.ActionContext, i tabscomp.TabChangeIntent) runtimeintent.IntentResult {
				if i.ComponentID != meta.ComponentID {
					return runtimeintent.IntentResult{}
				}
				fixtureStore.Update(func(s tabsFixtureState) tabsFixtureState {
					s.TabChangeCount++
					s.LastTabID = i.TabID
					s.LastTabLabel = i.TabLabel
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
		observedCurrent := ui.UseStoreSelector(fixtureStore, func(s tabsFixtureState) int { return s.ObservedCurrent })
		fieldChangeCount := ui.UseStoreSelector(fixtureStore, func(s tabsFixtureState) int { return s.FieldChangeCount })
		tabChangeCount := ui.UseStoreSelector(fixtureStore, func(s tabsFixtureState) int { return s.TabChangeCount })
		lastTabID := ui.UseStoreSelector(fixtureStore, func(s tabsFixtureState) string { return s.LastTabID })
		lastTabLabel := ui.UseStoreSelector(fixtureStore, func(s tabsFixtureState) string { return s.LastTabLabel })

		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Tabs E2E Fixture").Build(),
				ui.NewTabsBuilder().
					SetID(meta.ID).
					ComponentID(meta.ComponentID).
					Tabs(items).
					ActiveTab(observedCurrent).
					FieldIntent(runtimeintent.BindField(meta.CurrentField)).
					Width(72).
					Build(),
				ui.NewTextBuilder(fmt.Sprintf("ObservedCurrent: %d", observedCurrent)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("FieldChangeCount: %d", fieldChangeCount)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("TabChangeCount: %d", tabChangeCount)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("LastTabID: %s", lastTabID)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("LastTabLabel: %s", lastTabLabel)).Build(),
			})
	}

	return appFn, initFn, cleanupFn, fixtureStore, meta
}

func TestE2ETabsKeyboardNavigationSkipsDisabledAndHotkeys(t *testing.T) {
	items := []tabscomp.TabItem{
		tabscomp.Item("overview", "Overview").WithHotkey('o'),
		tabscomp.Item("logs", "Logs").WithHotkey('l').WithDisabled(true),
		tabscomp.Item("settings", "Settings").WithHotkey('s'),
	}
	appFn, initFn, cleanupFn, fixtureStore, meta := newTabsFixture("workspace-tabs", "fixture.tabs.workspace", items)
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(84, 16), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertFocus(ByComponentID(meta.ComponentID)); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("ObservedCurrent: 0")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Special(platform.KeyRight); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("ObservedCurrent: 2"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "key:Right"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.TabChangeIntentType); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("FieldChangeCount: 1")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("TabChangeCount: 1")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("LastTabID: settings")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("LastTabLabel: Settings")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Key('o'); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("ObservedCurrent: 0"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "key:o"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.TabChangeIntentType); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("FieldChangeCount: 2")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("TabChangeCount: 2")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("LastTabID: overview")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("LastTabLabel: Overview")); err != nil {
		t.Fatal(err)
	}

	state := fixtureStore.Get()
	if state.ObservedCurrent != 0 || state.FieldChangeCount != 2 || state.TabChangeCount != 2 || state.LastTabID != "overview" || state.LastTabLabel != "Overview" {
		t.Fatalf("unexpected tabs fixture state after keyboard flow: %+v", state)
	}
}

func TestE2ETabsClickIgnoresDisabledAndActivatesEnabledTab(t *testing.T) {
	items := []tabscomp.TabItem{
		tabscomp.Item("overview", "Overview"),
		tabscomp.Item("logs", "Logs").WithDisabled(true),
		tabscomp.Item("settings", "Settings"),
	}
	appFn, initFn, cleanupFn, fixtureStore, meta := newTabsFixture("workspace-tabs", "fixture.tabs.workspace", items)
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(84, 16), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	logsPoint, err := app.ResolvePoint(ByText("Logs"))
	if err != nil {
		t.Fatal(err)
	}
	if err := app.AssertHit(logsPoint, ByID(meta.ID)); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Click(ByText("Logs")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "mouse:left:release"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("ObservedCurrent: 0")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("FieldChangeCount: 0")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("TabChangeCount: 0")); err != nil {
		t.Fatal(err)
	}
	for _, logEntry := range app.IntentLogs() {
		if logEntry.Type == meta.FieldChangeIntentType || logEntry.Type == meta.TabChangeIntentType {
			t.Fatalf("disabled tab click should not emit change intents, got %+v", logEntry)
		}
	}

	settingsPoint, err := app.ResolvePoint(ByText("Settings"))
	if err != nil {
		t.Fatal(err)
	}
	if err := app.AssertHit(settingsPoint, ByID(meta.ID)); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Click(ByText("Settings")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("ObservedCurrent: 2"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "mouse:left:release"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.TabChangeIntentType); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("FieldChangeCount: 1")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("TabChangeCount: 1")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("LastTabID: settings")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("LastTabLabel: Settings")); err != nil {
		t.Fatal(err)
	}

	state := fixtureStore.Get()
	if state.ObservedCurrent != 2 || state.FieldChangeCount != 1 || state.TabChangeCount != 1 || state.LastTabID != "settings" || state.LastTabLabel != "Settings" {
		t.Fatalf("unexpected tabs fixture state after click flow: %+v", state)
	}
}
