package e2e

import (
	"testing"
	"time"

	runtimeaction "github.com/wwsheng009/mint/runtime/action"
	runtimeintent "github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/runtime/store"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
)

func TestE2EDataTableOperationalStates(t *testing.T) {
	appFn := func() ui.VNode {
		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("DataTable E2E Fixture").Build(),
				ui.DataTable(
					[]ui.TableColumn{
						{Title: "Provider", Width: 16},
						{Title: "Status", Width: 12},
					},
					[][]string{
						{"openai", "healthy"},
						{"azure", "degraded"},
					},
					ui.DataTableServerPagination(2, 25, 76),
					ui.DataTableOperationalStyle(),
				),
				ui.DataTable(
					[]ui.TableColumn{{Title: "Job", Width: 20}},
					[][]string{{"sync"}},
					ui.DataTableLoading(true),
					ui.DataTableLoadingText("Loading jobs..."),
					ui.DataTableOperationalStyle(),
				),
				ui.DataTable(
					[]ui.TableColumn{{Title: "Alert", Width: 34}},
					[][]string{{"latency"}},
					ui.DataTableErrorText("alerts API unavailable"),
					ui.DataTableOperationalStyle(),
				),
			})
	}

	app, err := Run(appFn, ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{
		"DataTable E2E Fixture",
		"Page 2/4 · Total 76 · Size 25",
		"Loading jobs...",
		"Loading",
		"alerts API unavailable",
		"Error · alerts API unavailable",
	} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatalf("expected %q to be visible: %v", text, err)
		}
	}
}

type dataTableKeyFixtureState struct {
	SelectedKey  string
	ActivatedKey string
}

func TestE2EDataTableStableKeySelectionAndActivation(t *testing.T) {
	const (
		selectedField  = "fixture.datatable.selected_key"
		activatedField = "fixture.datatable.activated_key"
	)

	fixtureStore := store.NewStore(dataTableKeyFixtureState{})
	var unregister func()
	initFn := func() {
		rt := rtui.GetGlobalIntentRuntime()
		if rt == nil {
			return
		}
		unregister = runtimeintent.RegisterTypedRuntime(rt, func(_ *runtimeintent.ActionContext, i runtimeintent.FieldChangeIntent) runtimeintent.IntentResult {
			switch i.Field {
			case selectedField:
				fixtureStore.Update(func(s dataTableKeyFixtureState) dataTableKeyFixtureState {
					s.SelectedKey = i.Value
					return s
				})
				return runtimeintent.HandledResult()
			case activatedField:
				fixtureStore.Update(func(s dataTableKeyFixtureState) dataTableKeyFixtureState {
					s.ActivatedKey = i.Value
					return s
				})
				return runtimeintent.HandledResult()
			default:
				return runtimeintent.IntentResult{}
			}
		})
	}
	defer func() {
		if unregister != nil {
			unregister()
		}
	}()

	appFn := func() ui.VNode {
		selectedKey := ui.UseStoreSelector(fixtureStore, func(s dataTableKeyFixtureState) string { return s.SelectedKey })
		activatedKey := ui.UseStoreSelector(fixtureStore, func(s dataTableKeyFixtureState) string { return s.ActivatedKey })

		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("DataTable Stable Key Fixture").Build(),
				ui.DataTable(
					[]ui.TableColumn{
						{Title: "Provider", Width: 18},
						{Title: "Status", Width: 10},
					},
					[][]string{
						{"openai", "healthy"},
						{"azure", "degraded"},
					},
					ui.DataTableComponentID("fixture.datatable.providers"),
					ui.DataTableRowKeys([]string{"provider.openai", "provider.azure"}),
					ui.DataTableSelectedKey(selectedKey),
					ui.DataTableSelectedKeyField(selectedField),
					ui.DataTableActivateKeyField(activatedField),
					ui.DataTableOperationalStyle(),
				),
				ui.NewTextBuilder("SelectedKey: " + formatDataTableKey(selectedKey)).Build(),
				ui.NewTextBuilder("ActivatedKey: " + formatDataTableKey(activatedKey)).Build(),
			})
	}

	app, err := Run(appFn, ui.WithSize(96, 18), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("SelectedKey: <empty>")); err != nil {
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
		return app.AssertVisible(ByText("SelectedKey: provider.openai"))
	}); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Special(platform.KeyEnter); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertActionSequence(runtimeaction.ActionEnter); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("ActivatedKey: provider.openai"))
	}); err != nil {
		t.Fatal(err)
	}
}

func formatDataTableKey(value string) string {
	if value == "" {
		return "<empty>"
	}
	return value
}
