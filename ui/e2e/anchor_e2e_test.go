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
	anchorcomp "github.com/wwsheng009/mint/ui/components/anchor"
)

type anchorFixtureState struct {
	ActiveKey    string
	ChangeCount  int
	LastField    string
	LastValue    string
}

type anchorFixtureMeta struct {
	ComponentID           string
	ListComponentID       string
	ActiveField           string
	FieldChangeIntentType string
}

func newAnchorFixture() (ui.ComponentFunc, func(), func(), *store.Store[anchorFixtureState], anchorFixtureMeta) {
	fixtureStore := store.NewStore(anchorFixtureState{
		ActiveKey: "intro",
	})
	meta := anchorFixtureMeta{
		ComponentID:           "fixture.anchor.main",
		ListComponentID:       "fixture.anchor.main-list",
		ActiveField:           "fixture.anchor.active",
		FieldChangeIntentType: runtimeintent.FieldChangeIntent{}.IntentType(),
	}

	unregisters := make([]func(), 0, 1)
	initFn := func() {
		rt := rtui.GetGlobalIntentRuntime()
		if rt == nil {
			return
		}
		unregisters = append(unregisters, runtimeintent.RegisterTypedRuntime(rt, func(_ *runtimeintent.ActionContext, i runtimeintent.FieldChangeIntent) runtimeintent.IntentResult {
			if i.Field != meta.ActiveField {
				return runtimeintent.IntentResult{}
			}
			fixtureStore.Update(func(s anchorFixtureState) anchorFixtureState {
				s.ActiveKey = i.Value
				s.ChangeCount++
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
		activeKey := ui.UseStoreSelector(fixtureStore, func(s anchorFixtureState) string { return s.ActiveKey })
		changeCount := ui.UseStoreSelector(fixtureStore, func(s anchorFixtureState) int { return s.ChangeCount })
		lastField := ui.UseStoreSelector(fixtureStore, func(s anchorFixtureState) string { return s.LastField })
		lastValue := ui.UseStoreSelector(fixtureStore, func(s anchorFixtureState) string { return s.LastValue })

		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Anchor E2E Fixture").Build(),
				ui.NewAnchorBuilder().
					SetID("anchor-main").
					ComponentID(meta.ComponentID).
					Title("Contents").
					ViewportHeight(4).
					ShowBorder(true).
					ActiveKey(activeKey).
					Items([]anchorcomp.Item{
						anchorcomp.NewItem("intro", "Introduction"),
						anchorcomp.NewItem("guide", "Guide Section"),
						{Key: "archive", Title: "Archive Section", Disabled: true},
						anchorcomp.NewItem("reference", "Reference Docs"),
					}).
					ForField(runtimeintent.BindField(meta.ActiveField)).
					Build(),
				ui.NewButtonBuilder("Anchor tail action").SetID("anchor-tail-action").Build(),
				ui.NewTextBuilder(fmt.Sprintf("ActiveKey: %s", activeKey)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("ChangeCount: %d", changeCount)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("LastField: %s", lastField)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("LastValue: %s", lastValue)).Build(),
			})
	}

	return appFn, initFn, cleanupFn, fixtureStore, meta
}

func TestE2EAnchorKeyboardNavigationFlow(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newAnchorFixture()
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(92, 18), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByComponentID(meta.ListComponentID)); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertFocus(ByComponentID(meta.ListComponentID)); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("ActiveKey: intro")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Special(platform.KeyDown); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("ActiveKey: guide")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("ChangeCount: 1")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("LastField: fixture.anchor.active")); err != nil {
			return err
		}
		return app.AssertVisible(ByText("LastValue: guide"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "key:Down"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}

	state := fixtureStore.Get()
	if state.ActiveKey != "guide" || state.ChangeCount != 1 || state.LastField != meta.ActiveField || state.LastValue != "guide" {
		t.Fatalf("unexpected anchor fixture state after keyboard flow: %+v", state)
	}
}

func TestE2EAnchorClickActivatesEnabledItem(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newAnchorFixture()
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(92, 18), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	point, err := app.ResolvePoint(ByText("Reference Docs"))
	if err != nil {
		t.Fatal(err)
	}
	if err := app.AssertHit(point, ByComponentID(meta.ListComponentID)); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().ClickAt(point.X, point.Y); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("ActiveKey: reference")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("ChangeCount: 1")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("LastField: fixture.anchor.active")); err != nil {
			return err
		}
		return app.AssertVisible(ByText("LastValue: reference"))
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
	if state.ActiveKey != "reference" || state.ChangeCount != 1 || state.LastField != meta.ActiveField || state.LastValue != "reference" {
		t.Fatalf("unexpected anchor fixture state after click flow: %+v", state)
	}
}

func TestE2EAnchorDisabledIgnoreAndTabOut(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newAnchorFixture()
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(92, 18), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	point, err := app.ResolvePoint(ByText("Archive Section"))
	if err != nil {
		t.Fatal(err)
	}
	if err := app.AssertHit(point, ByComponentID(meta.ListComponentID)); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Click(ByText("Archive Section")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "mouse:left:release"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("ActiveKey: intro")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("ChangeCount: 0")); err != nil {
		t.Fatal(err)
	}

	for _, logEntry := range app.IntentLogs() {
		if logEntry.Type == meta.FieldChangeIntentType {
			if fieldIntent, ok := logEntry.Intent.(runtimeintent.FieldChangeIntent); ok && fieldIntent.Field == meta.ActiveField {
				t.Fatalf("disabled anchor item should not emit active field changes, got %+v", logEntry)
			}
		}
	}

	app.ClearRawInputs()
	if err := app.Driver().Special(platform.KeyTab); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertFocus(ByID("anchor-tail-action")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "key:Tab"}); err != nil {
		t.Fatal(err)
	}

	state := fixtureStore.Get()
	if state.ActiveKey != "intro" || state.ChangeCount != 0 || state.LastField != "" || state.LastValue != "" {
		t.Fatalf("unexpected anchor fixture state after disabled flow: %+v", state)
	}
}
