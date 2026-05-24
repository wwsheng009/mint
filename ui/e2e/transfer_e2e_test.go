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
	listcomp "github.com/wwsheng009/mint/ui/components/list"
	transfercomp "github.com/wwsheng009/mint/ui/components/transfer"
)

type transferFixtureState struct {
	TargetKeys   []string
	FieldChanges int
	LastField    string
	LastValue    string
}

type transferFixtureMeta struct {
	ComponentID           string
	SourceListID          string
	TargetListID          string
	SourceSearchID        string
	TargetField           string
	FieldChangeIntentType string
}

func newTransferFixture() (ui.ComponentFunc, func(), func(), *store.Store[transferFixtureState], transferFixtureMeta) {
	meta := transferFixtureMeta{
		ComponentID:           "fixture.transfer.members",
		SourceListID:          "fixture.transfer.members-source",
		TargetListID:          "fixture.transfer.members-target",
		SourceSearchID:        "fixture.transfer.members-source-search",
		TargetField:           "fixture.transfer.targetKeys",
		FieldChangeIntentType: runtimeintent.FieldChangeIntent{}.IntentType(),
	}
	fixtureStore := store.NewStore(transferFixtureState{
		TargetKeys: []string{"delta"},
	})

	unregisters := make([]func(), 0, 2)
	initFn := func() {
		rt := rtui.GetGlobalIntentRuntime()
		if rt == nil {
			return
		}
		unregisters = append(unregisters,
			runtimeintent.RegisterTypedRuntime(rt, func(_ *runtimeintent.ActionContext, i runtimeintent.FieldChangeIntent) runtimeintent.IntentResult {
				if i.Field != meta.TargetField {
					return runtimeintent.IntentResult{}
				}
				fixtureStore.Update(func(s transferFixtureState) transferFixtureState {
					s.TargetKeys = parseTransferKeys(i.Value)
					s.FieldChanges++
					s.LastField = i.Field
					s.LastValue = i.Value
					return s
				})
				return runtimeintent.HandledResult()
			}),
			runtimeintent.RegisterTypedRuntime(rt, func(_ *runtimeintent.ActionContext, i listcomp.StateChangeIntent) runtimeintent.IntentResult {
				if i.ComponentID != meta.SourceListID && i.ComponentID != meta.TargetListID {
					return runtimeintent.IntentResult{}
				}
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
		targetKeys := ui.UseStoreSelector(fixtureStore, func(s transferFixtureState) []string { return s.TargetKeys })
		fieldChanges := ui.UseStoreSelector(fixtureStore, func(s transferFixtureState) int { return s.FieldChanges })
		lastField := ui.UseStoreSelector(fixtureStore, func(s transferFixtureState) string { return s.LastField })
		lastValue := ui.UseStoreSelector(fixtureStore, func(s transferFixtureState) string { return s.LastValue })

		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Transfer E2E Fixture").Build(),
				transfercomp.NewBuilder().
					SetID("members-transfer").
					ComponentID(meta.ComponentID).
					Titles("Backlog", "Done").
					Operations("Send", "Return").
					ListWidth(22).
					ListHeight(4).
					Width(56).
					Items([]transfercomp.Item{
						transfercomp.NewItem("alpha", "Alpha"),
						transfercomp.NewItem("beta", "Beta"),
						{Key: "gamma", Title: "Gamma", Disabled: true},
						transfercomp.NewItem("delta", "Delta"),
					}).
					TargetKeys(targetKeys).
					ForField(runtimeintent.BindField(meta.TargetField)).
					Build(),
				ui.NewButtonBuilder("Transfer tail action").SetID("transfer-tail-action").Build(),
				ui.NewTextBuilder(fmt.Sprintf("TargetKeys: %s", formatTransferKeys(targetKeys))).Build(),
				ui.NewTextBuilder(fmt.Sprintf("FieldChanges: %d", fieldChanges)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("LastField: %s", formatTransferValue(lastField))).Build(),
				ui.NewTextBuilder(fmt.Sprintf("LastValue: %s", formatTransferValue(lastValue))).Build(),
			})
	}

	return appFn, initFn, cleanupFn, fixtureStore, meta
}

func newTransferSearchFixture() (ui.ComponentFunc, func(), func(), *store.Store[transferFixtureState], transferFixtureMeta) {
	meta := transferFixtureMeta{
		ComponentID:           "fixture.transfer.search",
		SourceListID:          "fixture.transfer.search-source",
		TargetListID:          "fixture.transfer.search-target",
		SourceSearchID:        "fixture.transfer.search-source-search",
		TargetField:           "fixture.transfer.search.targetKeys",
		FieldChangeIntentType: runtimeintent.FieldChangeIntent{}.IntentType(),
	}
	fixtureStore := store.NewStore(transferFixtureState{
		TargetKeys: []string{"delta"},
	})

	unregisters := make([]func(), 0, 1)
	initFn := func() {
		rt := rtui.GetGlobalIntentRuntime()
		if rt == nil {
			return
		}
		unregisters = append(unregisters, runtimeintent.RegisterTypedRuntime(rt, func(_ *runtimeintent.ActionContext, i runtimeintent.FieldChangeIntent) runtimeintent.IntentResult {
			if i.Field != meta.TargetField {
				return runtimeintent.IntentResult{}
			}
			fixtureStore.Update(func(s transferFixtureState) transferFixtureState {
				s.TargetKeys = parseTransferKeys(i.Value)
				s.FieldChanges++
				s.LastField = i.Field
				s.LastValue = i.Value
				return s
			})
			return runtimeintent.HandledResult()
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
		targetKeys := ui.UseStoreSelector(fixtureStore, func(s transferFixtureState) []string { return s.TargetKeys })
		fieldChanges := ui.UseStoreSelector(fixtureStore, func(s transferFixtureState) int { return s.FieldChanges })
		lastValue := ui.UseStoreSelector(fixtureStore, func(s transferFixtureState) string { return s.LastValue })

		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Transfer Search E2E Fixture").Build(),
				transfercomp.NewBuilder().
					SetID("search-transfer").
					ComponentID(meta.ComponentID).
					Titles("Backlog", "Done").
					Operations("Send", "Return").
					Searchable(true).
					BulkOperations(true).
					BulkOperationLabels("All Send", "All Return").
					SearchPlaceholders("Find backlog", "Find done").
					ListWidth(24).
					ListHeight(4).
					Width(62).
					Items([]transfercomp.Item{
						transfercomp.NewItem("alpha", "Alpha"),
						transfercomp.NewItem("beta", "Beta"),
						transfercomp.NewItem("gamma", "Gamma"),
						transfercomp.NewItem("delta", "Delta"),
					}).
					TargetKeys(targetKeys).
					ForField(runtimeintent.BindField(meta.TargetField)).
					Build(),
				ui.NewTextBuilder(fmt.Sprintf("TargetKeys: %s", formatTransferKeys(targetKeys))).Build(),
				ui.NewTextBuilder(fmt.Sprintf("FieldChanges: %d", fieldChanges)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("LastValue: %s", formatTransferValue(lastValue))).Build(),
			})
	}

	return appFn, initFn, cleanupFn, fixtureStore, meta
}

func TestE2ETransferSearchFiltersAndMovesVisibleItem(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newTransferSearchFixture()
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(100, 24), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("Backlog (3)")); err != nil {
		t.Fatal(err)
	}
	if err := app.Driver().Click(ByID(meta.SourceSearchID)); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitFocus(ByID(meta.SourceSearchID), 500*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	app.ClearIntentLogs()
	app.ClearRawInputs()
	for _, key := range []rune{'b', 'e', 't'} {
		if err := app.Driver().Key(key); err != nil {
			t.Fatal(err)
		}
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("Backlog (1/3)")); err != nil {
			return err
		}
		return app.AssertVisible(ByText("Beta"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "key:b"}); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Click(ByText("Beta")); err != nil {
		t.Fatal(err)
	}
	if err := app.Driver().Click(ByText("Send")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("TargetKeys: beta,delta")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("FieldChanges: 1")); err != nil {
			return err
		}
		return app.AssertVisible(ByText("LastValue: beta,delta"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}

	state := fixtureStore.Get()
	if formatTransferKeys(state.TargetKeys) != "beta,delta" || state.FieldChanges != 1 || state.LastValue != "beta,delta" {
		t.Fatalf("unexpected transfer search fixture state: %+v", state)
	}
}

func TestE2ETransferBulkMovesVisibleFilteredItems(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newTransferSearchFixture()
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(100, 24), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("All Send")); err != nil {
		t.Fatal(err)
	}
	if err := app.Driver().Click(ByID(meta.SourceSearchID)); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitFocus(ByID(meta.SourceSearchID), 500*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	app.ClearIntentLogs()
	app.ClearRawInputs()
	for _, key := range []rune{'a', 'l', 'p'} {
		if err := app.Driver().Key(key); err != nil {
			t.Fatal(err)
		}
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("Backlog (1/3)")); err != nil {
			return err
		}
		return app.AssertVisible(ByText("Alpha"))
	}); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Click(ByText("All Send")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("TargetKeys: alpha,delta")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("FieldChanges: 1")); err != nil {
			return err
		}
		return app.AssertVisible(ByText("LastValue: alpha,delta"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}

	state := fixtureStore.Get()
	if formatTransferKeys(state.TargetKeys) != "alpha,delta" || state.FieldChanges != 1 || state.LastValue != "alpha,delta" {
		t.Fatalf("unexpected transfer bulk fixture state: %+v", state)
	}
}

func TestE2ETransferKeyboardSelectionAndMoveToTarget(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newTransferFixture()
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(96, 22), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertFocus(ByComponentID(meta.SourceListID)); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Backlog (3)")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Done (1)")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("TargetKeys: delta")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Special(platform.KeyEnter); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "key:Enter"}); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Click(ByText("Send")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("Backlog (2)")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("Done (2)")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("TargetKeys: alpha,delta")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("FieldChanges: 1")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("LastField: fixture.transfer.targetKeys")); err != nil {
			return err
		}
		return app.AssertVisible(ByText("LastValue: alpha,delta"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "mouse:left:release"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}

	state := fixtureStore.Get()
	if formatTransferKeys(state.TargetKeys) != "alpha,delta" || state.FieldChanges != 1 || state.LastField != meta.TargetField || state.LastValue != "alpha,delta" {
		t.Fatalf("unexpected transfer fixture state after move to target: %+v", state)
	}
}

func TestE2ETransferDisabledIgnoreAndMoveBackToSource(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newTransferFixture()
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(96, 22), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	disabledPoint, err := app.ResolvePoint(ByText("Gamma"))
	if err != nil {
		t.Fatal(err)
	}
	if err := app.AssertHit(disabledPoint, ByComponentID(meta.SourceListID)); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().ClickAt(disabledPoint.X, disabledPoint.Y); err != nil {
		t.Fatal(err)
	}
	if err := app.Driver().Click(ByText("Send")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(300*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("Backlog (3)")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("Done (1)")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("TargetKeys: delta")); err != nil {
			return err
		}
		return app.AssertVisible(ByText("FieldChanges: 0"))
	}); err != nil {
		t.Fatal(err)
	}

	deltaPoint, err := app.ResolvePoint(ByText("Delta"))
	if err != nil {
		t.Fatal(err)
	}
	if err := app.AssertHit(deltaPoint, ByComponentID(meta.TargetListID)); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().ClickAt(deltaPoint.X, deltaPoint.Y); err != nil {
		t.Fatal(err)
	}
	if err := app.Driver().Click(ByText("Return")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("Backlog (4)")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("Done (0)")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("TargetKeys: <empty>")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("FieldChanges: 1")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("LastField: fixture.transfer.targetKeys")); err != nil {
			return err
		}
		return app.AssertVisible(ByText("LastValue: <empty>"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}

	app.ClearRawInputs()
	if err := app.Driver().Special(platform.KeyTab); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertFocus(ByID("transfer-tail-action")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "key:Tab"}); err != nil {
		t.Fatal(err)
	}

	state := fixtureStore.Get()
	if len(state.TargetKeys) != 0 || state.FieldChanges != 1 || state.LastField != meta.TargetField || state.LastValue != "" {
		t.Fatalf("unexpected transfer fixture state after move to source: %+v", state)
	}
}

func parseTransferKeys(value string) []string {
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
	if len(keys) == 0 {
		return nil
	}
	return keys
}

func formatTransferKeys(keys []string) string {
	if len(keys) == 0 {
		return "<empty>"
	}
	return strings.Join(keys, ",")
}

func formatTransferValue(value string) string {
	if strings.TrimSpace(value) == "" {
		return "<empty>"
	}
	return value
}
