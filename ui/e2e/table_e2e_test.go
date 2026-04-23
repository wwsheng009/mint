package e2e

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	runtimeaction "github.com/wwsheng009/mint/runtime/action"
	runtimeintent "github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/runtime/store"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
	tablecomp "github.com/wwsheng009/mint/ui/components/table"
)

type tableFixtureState struct {
	SelectedRow          int
	CurrentPage          int
	SelectedFieldChanges int
	PageFieldChanges     int
	StateChangeCount     int
	SortColumn           int
	SortDescending       bool
	LastField            string
	LastValue            string
}

type tableFixtureMeta struct {
	ComponentID           string
	SelectedField         string
	PageField             string
	FieldChangeIntentType string
	StateChangeIntentType string
}

func newTableFixture() (ui.ComponentFunc, func(), func(), *store.Store[tableFixtureState], tableFixtureMeta) {
	fixtureStore := store.NewStore(tableFixtureState{
		SelectedRow: -1,
		CurrentPage: 0,
		SortColumn:  -1,
	})
	meta := tableFixtureMeta{
		ComponentID:           "fixture.table.orders",
		SelectedField:         "fixture.table.selected",
		PageField:             "fixture.table.page",
		FieldChangeIntentType: runtimeintent.FieldChangeIntent{}.IntentType(),
		StateChangeIntentType: tablecomp.StateChangeIntent{}.IntentType(),
	}

	unregisters := make([]func(), 0, 2)
	initFn := func() {
		rt := rtui.GetGlobalIntentRuntime()
		if rt == nil {
			return
		}
		unregisters = append(unregisters,
			runtimeintent.RegisterTypedRuntime(rt, func(_ *runtimeintent.ActionContext, i runtimeintent.FieldChangeIntent) runtimeintent.IntentResult {
				switch i.Field {
				case meta.SelectedField:
					nextSelected, err := strconv.Atoi(i.Value)
					if err != nil {
						return runtimeintent.IntentResult{}
					}
					fixtureStore.Update(func(s tableFixtureState) tableFixtureState {
						s.SelectedRow = nextSelected
						s.SelectedFieldChanges++
						s.LastField = i.Field
						s.LastValue = i.Value
						return s
					})
					return runtimeintent.HandledResult()
				case meta.PageField:
					nextPage, err := strconv.Atoi(i.Value)
					if err != nil {
						return runtimeintent.IntentResult{}
					}
					fixtureStore.Update(func(s tableFixtureState) tableFixtureState {
						s.CurrentPage = nextPage
						s.PageFieldChanges++
						s.LastField = i.Field
						s.LastValue = i.Value
						return s
					})
					return runtimeintent.HandledResult()
				default:
					return runtimeintent.IntentResult{}
				}
			}),
			runtimeintent.RegisterTypedRuntime(rt, func(_ *runtimeintent.ActionContext, i tablecomp.StateChangeIntent) runtimeintent.IntentResult {
				if i.ComponentID != meta.ComponentID {
					return runtimeintent.IntentResult{}
				}
				fixtureStore.Update(func(s tableFixtureState) tableFixtureState {
					s.StateChangeCount++
					s.SortColumn = i.SortColumn
					s.SortDescending = i.SortDescending
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
		selectedRow := ui.UseStoreSelector(fixtureStore, func(s tableFixtureState) int { return s.SelectedRow })
		currentPage := ui.UseStoreSelector(fixtureStore, func(s tableFixtureState) int { return s.CurrentPage })
		selectedFieldChanges := ui.UseStoreSelector(fixtureStore, func(s tableFixtureState) int { return s.SelectedFieldChanges })
		pageFieldChanges := ui.UseStoreSelector(fixtureStore, func(s tableFixtureState) int { return s.PageFieldChanges })
		stateChangeCount := ui.UseStoreSelector(fixtureStore, func(s tableFixtureState) int { return s.StateChangeCount })
		sortColumn := ui.UseStoreSelector(fixtureStore, func(s tableFixtureState) int { return s.SortColumn })
		sortDescending := ui.UseStoreSelector(fixtureStore, func(s tableFixtureState) bool { return s.SortDescending })
		lastField := ui.UseStoreSelector(fixtureStore, func(s tableFixtureState) string { return s.LastField })
		lastValue := ui.UseStoreSelector(fixtureStore, func(s tableFixtureState) string { return s.LastValue })

		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Table E2E Fixture").Build(),
				tablecomp.NewBuilder().
					SetID("orders-table").
					ComponentID(meta.ComponentID).
					Columns([]tablecomp.TableColumn{
						{Title: "ID", Width: 4},
						{Title: "Name", Width: 10, Sortable: true},
					}).
					Rows([][]string{
						{"3", "Carol"},
						{"2", "Bob"},
						{"1", "Alice"},
						{"4", "Dave"},
					}).
					PageSize(2).
					ShowBorder(true).
					ShowFooter(true).
					SelectedIndex(selectedRow).
					CurrentPage(currentPage).
					ForField(runtimeintent.BindField(meta.SelectedField)).
					PageForField(runtimeintent.BindField(meta.PageField)).
					Build(),
				ui.NewButtonBuilder("Table tail action").SetID("table-tail-action").Build(),
				ui.NewTextBuilder(fmt.Sprintf("SelectedRow: %d", selectedRow)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("CurrentPage: %d", currentPage)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("SelectedFieldChanges: %d", selectedFieldChanges)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("PageFieldChanges: %d", pageFieldChanges)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("StateChangeCount: %d", stateChangeCount)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("SortColumn: %d", sortColumn)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("SortDescending: %t", sortDescending)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("LastField: %s", formatTableValue(lastField))).Build(),
				ui.NewTextBuilder(fmt.Sprintf("LastValue: %s", formatTableValue(lastValue))).Build(),
			})
	}

	return appFn, initFn, cleanupFn, fixtureStore, meta
}

func TestE2ETableSortableHeaderClickFlow(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newTableFixture()
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(96, 22), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertFocus(ByID("orders-table")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Carol")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Bob")); err != nil {
		t.Fatal(err)
	}

	namePoint, err := app.ResolvePoint(ByText("Name"))
	if err != nil {
		t.Fatal(err)
	}
	if err := app.AssertHit(namePoint, ByID("orders-table")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().ClickAt(namePoint.X, namePoint.Y); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertActionSequence(runtimeaction.ActionClick); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("Alice")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("Bob")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("SortColumn: 1")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("SortDescending: false")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("StateChangeCount: 1")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("Carol")); err == nil {
			return fmt.Errorf("Carol should not remain on first page after ascending sort")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.StateChangeIntentType); err != nil {
		t.Fatal(err)
	}

	namePoint, err = app.ResolvePoint(ByText("Name"))
	if err != nil {
		t.Fatal(err)
	}
	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().ClickAt(namePoint.X, namePoint.Y); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertActionSequence(runtimeaction.ActionClick); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("Dave")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("Carol")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("SortColumn: 1")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("SortDescending: true")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("StateChangeCount: 2")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("Alice")); err == nil {
			return fmt.Errorf("Alice should not remain on first page after descending sort")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.StateChangeIntentType); err != nil {
		t.Fatal(err)
	}

	state := fixtureStore.Get()
	if state.SortColumn != 1 || !state.SortDescending || state.StateChangeCount != 2 {
		t.Fatalf("unexpected table fixture state after sort flow: %+v", state)
	}
}

func TestE2ETableKeyboardPageNavigationAndFieldSync(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newTableFixture()
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(96, 22), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertFocus(ByID("orders-table")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("SelectedRow: -1")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("CurrentPage: 0")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Special(platform.KeyDown); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertActionSequence(runtimeaction.ActionNavigateDown); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("SelectedRow: 0")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("SelectedFieldChanges: 1")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("PageFieldChanges: 1")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("LastField: fixture.table.page")); err != nil {
			return err
		}
		return app.AssertVisible(ByText("LastValue: 0"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Special(platform.KeyPageDown); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertActionSequence(runtimeaction.ActionNavigatePageDown); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("SelectedRow: 2")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("CurrentPage: 1")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("SelectedFieldChanges: 2")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("PageFieldChanges: 2")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("LastField: fixture.table.page")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("LastValue: 1")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("Alice")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("Dave")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("Bob")); err == nil {
			return fmt.Errorf("Bob should not remain visible after page down")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.StateChangeIntentType); err != nil {
		t.Fatal(err)
	}

	app.ClearRawInputs()
	if err := app.Driver().Special(platform.KeyTab); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertFocus(ByID("table-tail-action")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertActionSequence(runtimeaction.ActionNavigateNext); err != nil {
		t.Fatal(err)
	}

	state := fixtureStore.Get()
	if state.SelectedRow != 2 || state.CurrentPage != 1 || state.SelectedFieldChanges != 2 || state.PageFieldChanges != 2 {
		t.Fatalf("unexpected table fixture state after keyboard flow: %+v", state)
	}
}

func formatTableValue(value string) string {
	if strings.TrimSpace(value) == "" {
		return "<empty>"
	}
	return value
}
