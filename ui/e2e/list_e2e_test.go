package e2e

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	runtimeintent "github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/runtime/store"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
	listcomp "github.com/wwsheng009/mint/ui/components/list"
)

type listFixtureState struct {
	SelectedIndex         int
	SelectedFieldChanges  int
	CheckedIndices        []int
	SelectionFieldChanges int
	StateChangeCount      int
	LastField             string
	LastValue             string
}

type listFixtureMeta struct {
	ComponentID           string
	SelectedField         string
	SelectionField        string
	FieldChangeIntentType string
	StateChangeIntentType string
}

func newListFixture() (ui.ComponentFunc, func(), func(), *store.Store[listFixtureState], listFixtureMeta) {
	fixtureStore := store.NewStore(listFixtureState{
		SelectedIndex: 0,
	})
	meta := listFixtureMeta{
		ComponentID:           "fixture.list.main",
		SelectedField:         "fixture.list.selected",
		SelectionField:        "fixture.list.checked",
		FieldChangeIntentType: runtimeintent.FieldChangeIntent{}.IntentType(),
		StateChangeIntentType: listcomp.StateChangeIntent{}.IntentType(),
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
					nextIndex, err := strconv.Atoi(i.Value)
					if err != nil {
						return runtimeintent.IntentResult{}
					}
					fixtureStore.Update(func(s listFixtureState) listFixtureState {
						s.SelectedIndex = nextIndex
						s.SelectedFieldChanges++
						s.LastField = i.Field
						s.LastValue = i.Value
						return s
					})
					return runtimeintent.HandledResult()
				case meta.SelectionField:
					fixtureStore.Update(func(s listFixtureState) listFixtureState {
						s.CheckedIndices = parseListCheckedIndices(i.Value)
						s.SelectionFieldChanges++
						s.LastField = i.Field
						s.LastValue = i.Value
						return s
					})
					return runtimeintent.HandledResult()
				default:
					return runtimeintent.IntentResult{}
				}
			}),
			runtimeintent.RegisterTypedRuntime(rt, func(_ *runtimeintent.ActionContext, i listcomp.StateChangeIntent) runtimeintent.IntentResult {
				if i.ComponentID != meta.ComponentID {
					return runtimeintent.IntentResult{}
				}
				fixtureStore.Update(func(s listFixtureState) listFixtureState {
					s.StateChangeCount++
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
		selectedIndex := ui.UseStoreSelector(fixtureStore, func(s listFixtureState) int { return s.SelectedIndex })
		selectedFieldChanges := ui.UseStoreSelector(fixtureStore, func(s listFixtureState) int { return s.SelectedFieldChanges })
		checkedIndices := ui.UseStoreSelector(fixtureStore, func(s listFixtureState) []int { return s.CheckedIndices })
		selectionFieldChanges := ui.UseStoreSelector(fixtureStore, func(s listFixtureState) int { return s.SelectionFieldChanges })
		stateChangeCount := ui.UseStoreSelector(fixtureStore, func(s listFixtureState) int { return s.StateChangeCount })
		lastField := ui.UseStoreSelector(fixtureStore, func(s listFixtureState) string { return s.LastField })
		lastValue := ui.UseStoreSelector(fixtureStore, func(s listFixtureState) string { return s.LastValue })

		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("List E2E Fixture").Build(),
				ui.NewListBuilder().
					SetID("list-main").
					ComponentID(meta.ComponentID).
					Header("Tasks").
					Rows([]string{"Build", "Test", "Deploy", "Ship"}).
					ViewportHeight(3).
					ShowBorder(true).
					MultiSelect().
					SelectedIndex(selectedIndex).
					CheckedIndices(checkedIndices...).
					ForField(runtimeintent.BindField(meta.SelectedField)).
					SelectionForField(runtimeintent.BindField(meta.SelectionField)).
					Build(),
				ui.NewButtonBuilder("List tail action").SetID("list-tail-action").Build(),
				ui.NewTextBuilder(fmt.Sprintf("SelectedIndex: %d", selectedIndex)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("SelectedFieldChanges: %d", selectedFieldChanges)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("CheckedIndices: %s", formatListCheckedIndices(checkedIndices))).Build(),
				ui.NewTextBuilder(fmt.Sprintf("SelectionFieldChanges: %d", selectionFieldChanges)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("StateChangeCount: %d", stateChangeCount)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("LastField: %s", lastField)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("LastValue: %s", formatListFieldValue(lastValue))).Build(),
			})
	}

	return appFn, initFn, cleanupFn, fixtureStore, meta
}

func TestE2EListKeyboardSelectionAndToggleFlow(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newListFixture()
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(96, 20), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertFocus(ByComponentID(meta.ComponentID)); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("SelectedIndex: 0")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("CheckedIndices: <empty>")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Special(platform.KeyDown); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("SelectedIndex: 1")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("SelectedFieldChanges: 1")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("StateChangeCount: 1")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("LastField: fixture.list.selected")); err != nil {
			return err
		}
		return app.AssertVisible(ByText("LastValue: 1"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "key:Down"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.StateChangeIntentType); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Special(platform.KeyEnter); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("CheckedIndices: 1")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("SelectionFieldChanges: 1")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("StateChangeCount: 2")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("LastField: fixture.list.checked")); err != nil {
			return err
		}
		return app.AssertVisible(ByText("LastValue: 1"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "key:Enter"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.StateChangeIntentType); err != nil {
		t.Fatal(err)
	}

	state := fixtureStore.Get()
	if state.SelectedIndex != 1 || state.SelectedFieldChanges != 1 || formatListCheckedIndices(state.CheckedIndices) != "1" || state.SelectionFieldChanges != 1 || state.StateChangeCount != 2 {
		t.Fatalf("unexpected list fixture state after keyboard flow: %+v", state)
	}
}

func TestE2EListClickToggleAndTabOut(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newListFixture()
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(96, 20), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	point, err := app.ResolvePoint(ByText("Deploy"))
	if err != nil {
		t.Fatal(err)
	}
	if err := app.AssertHit(point, ByComponentID(meta.ComponentID)); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Click(ByText("Deploy")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("SelectedIndex: 2")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("SelectedFieldChanges: 1")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("CheckedIndices: 2")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("SelectionFieldChanges: 1")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("LastField: fixture.list.checked")); err != nil {
			return err
		}
		return app.AssertVisible(ByText("LastValue: 2"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "mouse:left:release"}); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Click(ByText("Deploy")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("SelectedIndex: 2")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("CheckedIndices: <empty>")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("SelectionFieldChanges: 2")); err != nil {
			return err
		}
		return app.AssertVisible(ByText("LastValue: <empty>"))
	}); err != nil {
		t.Fatal(err)
	}

	app.ClearRawInputs()
	if err := app.Driver().Special(platform.KeyTab); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertFocus(ByID("list-tail-action")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "key:Tab"}); err != nil {
		t.Fatal(err)
	}

	state := fixtureStore.Get()
	if state.SelectedIndex != 2 || state.SelectedFieldChanges != 1 || len(state.CheckedIndices) != 0 || state.SelectionFieldChanges != 2 {
		t.Fatalf("unexpected list fixture state after click and tab flow: %+v", state)
	}
}

func parseListCheckedIndices(value string) []int {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	indices := make([]int, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		index, err := strconv.Atoi(part)
		if err != nil {
			continue
		}
		indices = append(indices, index)
	}
	return indices
}

func formatListCheckedIndices(indices []int) string {
	if len(indices) == 0 {
		return "<empty>"
	}
	parts := make([]string, len(indices))
	for i, index := range indices {
		parts[i] = strconv.Itoa(index)
	}
	return strings.Join(parts, ",")
}

func formatListFieldValue(value string) string {
	if strings.TrimSpace(value) == "" {
		return "<empty>"
	}
	return value
}
