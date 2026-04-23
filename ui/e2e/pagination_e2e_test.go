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
)

type paginationFixtureState struct {
	CurrentPage          int
	FieldChangeCount     int
	DisabledPage         int
	DisabledFieldChanges int
	LastField            string
	LastValue            string
}

type paginationFixtureMeta struct {
	PageField             string
	DisabledPageField     string
	FieldChangeIntentType string
}

func newPaginationFixture() (ui.ComponentFunc, func(), func(), *store.Store[paginationFixtureState], paginationFixtureMeta) {
	fixtureStore := store.NewStore(paginationFixtureState{
		CurrentPage:  2,
		DisabledPage: 1,
	})
	meta := paginationFixtureMeta{
		PageField:             "fixture.pagination.current",
		DisabledPageField:     "fixture.pagination.disabled.current",
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
				switch i.Field {
				case meta.PageField:
					nextPage, err := strconv.Atoi(i.Value)
					if err != nil {
						return runtimeintent.IntentResult{}
					}
					fixtureStore.Update(func(s paginationFixtureState) paginationFixtureState {
						s.CurrentPage = nextPage
						s.FieldChangeCount++
						s.LastField = i.Field
						s.LastValue = i.Value
						return s
					})
					return runtimeintent.HandledResult()
				case meta.DisabledPageField:
					nextPage, err := strconv.Atoi(i.Value)
					if err != nil {
						return runtimeintent.IntentResult{}
					}
					fixtureStore.Update(func(s paginationFixtureState) paginationFixtureState {
						s.DisabledPage = nextPage
						s.DisabledFieldChanges++
						s.LastField = i.Field
						s.LastValue = i.Value
						return s
					})
					return runtimeintent.HandledResult()
				default:
					return runtimeintent.IntentResult{}
				}
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
		currentPage := ui.UseStoreSelector(fixtureStore, func(s paginationFixtureState) int { return s.CurrentPage })
		fieldChangeCount := ui.UseStoreSelector(fixtureStore, func(s paginationFixtureState) int { return s.FieldChangeCount })
		disabledPage := ui.UseStoreSelector(fixtureStore, func(s paginationFixtureState) int { return s.DisabledPage })
		disabledFieldChanges := ui.UseStoreSelector(fixtureStore, func(s paginationFixtureState) int { return s.DisabledFieldChanges })
		lastField := ui.UseStoreSelector(fixtureStore, func(s paginationFixtureState) string { return s.LastField })
		lastValue := ui.UseStoreSelector(fixtureStore, func(s paginationFixtureState) string { return s.LastValue })

		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Pagination E2E Fixture").Build(),
				ui.NewPaginationBuilder().
					SetID("pagination-main").
					ComponentID("fixture.pagination.main").
					Total(120).
					PageSize(10).
					CurrentPage(currentPage).
					MaxButtons(5).
					PageForField(runtimeintent.BindField(meta.PageField)).
					Build(),
				ui.NewPaginationBuilder().
					SetID("pagination-disabled").
					ComponentID("fixture.pagination.disabled").
					Total(30).
					PageSize(10).
					CurrentPage(disabledPage).
					Disabled(true).
					ShowTotal(false).
					PageForField(runtimeintent.BindField(meta.DisabledPageField)).
					Build(),
				ui.NewButtonBuilder("Pagination tail action").SetID("pagination-tail-action").Build(),
				ui.NewTextBuilder(fmt.Sprintf("CurrentPage: %d", currentPage)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("FieldChangeCount: %d", fieldChangeCount)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("DisabledPage: %d", disabledPage)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("DisabledFieldChanges: %d", disabledFieldChanges)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("LastField: %s", lastField)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("LastValue: %s", lastValue)).Build(),
			})
	}

	return appFn, initFn, cleanupFn, fixtureStore, meta
}

func TestE2EPaginationKeyboardNavigationAndFieldSync(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newPaginationFixture()
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(96, 18), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertFocus(ByID("pagination-main")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("CurrentPage: 2")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("[3]")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Special(platform.KeyRight); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("CurrentPage: 3")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("FieldChangeCount: 1")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("LastField: fixture.pagination.current")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("LastValue: 3")); err != nil {
			return err
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "key:Right"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Special(platform.KeyLeft); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("CurrentPage: 2")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("FieldChangeCount: 2")); err != nil {
			return err
		}
		return app.AssertVisible(ByText("LastValue: 2"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "key:Left"}); err != nil {
		t.Fatal(err)
	}

	state := fixtureStore.Get()
	if state.CurrentPage != 2 || state.FieldChangeCount != 2 || state.LastValue != "2" {
		t.Fatalf("unexpected pagination fixture state after keyboard flow: %+v", state)
	}
}

func TestE2EPaginationClickDisabledIgnoreAndTabSkip(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newPaginationFixture()
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(96, 18), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	prevPoint, err := app.ResolvePoint(ByText("Prev"))
	if err != nil {
		t.Fatal(err)
	}
	if err := app.AssertHit(prevPoint, ByID("pagination-main")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Click(ByText("Prev")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("CurrentPage: 1")); err != nil {
			return err
		}
		return app.AssertVisible(ByText("FieldChangeCount: 1"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "mouse:left:release"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}

	app.ClearRawInputs()
	if err := app.Driver().Special(platform.KeyTab); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertFocus(ByID("pagination-tail-action")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "key:Tab"}); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Click(ByID("pagination-disabled")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "mouse:left:release"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("DisabledPage: 1")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("DisabledFieldChanges: 0")); err != nil {
		t.Fatal(err)
	}
	for _, logEntry := range app.IntentLogs() {
		if logEntry.Type == meta.FieldChangeIntentType {
			if fieldIntent, ok := logEntry.Intent.(runtimeintent.FieldChangeIntent); ok && fieldIntent.Field == meta.DisabledPageField {
				t.Fatalf("disabled pagination should not emit disabled field changes, got %+v", logEntry)
			}
		}
	}

	state := fixtureStore.Get()
	if state.CurrentPage != 1 || state.FieldChangeCount != 1 || state.DisabledFieldChanges != 0 {
		t.Fatalf("unexpected pagination fixture state after click and disabled flow: %+v", state)
	}
}
