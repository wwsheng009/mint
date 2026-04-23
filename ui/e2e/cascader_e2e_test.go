package e2e

import (
	"fmt"
	"testing"
	"time"

	runtimeaction "github.com/wwsheng009/mint/runtime/action"
	runtimeintent "github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/runtime/store"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
	cascadercomp "github.com/wwsheng009/mint/ui/components/cascader"
)

type cascaderFixtureState struct {
	CityValue          string
	RegionValue        string
	CityFieldChanges   int
	RegionFieldChanges int
	LastField          string
	LastValue          string
}

type cascaderFixtureMeta struct {
	CityField             string
	RegionField           string
	FieldChangeIntentType string
}

func newCascaderFixture() (ui.ComponentFunc, func(), func(), *store.Store[cascaderFixtureState], cascaderFixtureMeta) {
	fixtureStore := store.NewStore(cascaderFixtureState{})
	meta := cascaderFixtureMeta{
		CityField:             "fixture.cascader.city",
		RegionField:           "fixture.cascader.region",
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
				case meta.CityField:
					fixtureStore.Update(func(s cascaderFixtureState) cascaderFixtureState {
						s.CityValue = i.Value
						s.CityFieldChanges++
						s.LastField = i.Field
						s.LastValue = i.Value
						return s
					})
					return runtimeintent.HandledResult()
				case meta.RegionField:
					fixtureStore.Update(func(s cascaderFixtureState) cascaderFixtureState {
						s.RegionValue = i.Value
						s.RegionFieldChanges++
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
		cityValue := ui.UseStoreSelector(fixtureStore, func(s cascaderFixtureState) string { return s.CityValue })
		regionValue := ui.UseStoreSelector(fixtureStore, func(s cascaderFixtureState) string { return s.RegionValue })
		cityFieldChanges := ui.UseStoreSelector(fixtureStore, func(s cascaderFixtureState) int { return s.CityFieldChanges })
		regionFieldChanges := ui.UseStoreSelector(fixtureStore, func(s cascaderFixtureState) int { return s.RegionFieldChanges })
		lastField := ui.UseStoreSelector(fixtureStore, func(s cascaderFixtureState) string { return s.LastField })
		lastValue := ui.UseStoreSelector(fixtureStore, func(s cascaderFixtureState) string { return s.LastValue })

		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Cascader E2E Fixture").Build(),
				cascadercomp.NewBuilder().
					SetID("city-cascader").
					ComponentID("fixture.cascader.city").
					Placeholder("Select city").
					Width(22).
					Options(cascaderFixtureOptions()).
					ForField(runtimeintent.BindField(meta.CityField)).
					Build(),
				cascadercomp.NewBuilder().
					SetID("region-cascader").
					ComponentID("fixture.cascader.region").
					Placeholder("Select region").
					Width(22).
					ChangeOnSelect(true).
					Options(cascaderFixtureOptions()).
					ForField(runtimeintent.BindField(meta.RegionField)).
					Build(),
				cascadercomp.NewBuilder().
					SetID("disabled-cascader").
					ComponentID("fixture.cascader.disabled").
					Placeholder("Disabled region").
					Width(22).
					Disabled(true).
					Options(cascaderFixtureOptions()).
					Build(),
				ui.NewButtonBuilder("Cascader tail action").SetID("cascader-tail-action").Build(),
				ui.NewTextBuilder(fmt.Sprintf("CityValue: %s", formatCascaderValue(cityValue))).Build(),
				ui.NewTextBuilder(fmt.Sprintf("RegionValue: %s", formatCascaderValue(regionValue))).Build(),
				ui.NewTextBuilder(fmt.Sprintf("CityFieldChanges: %d", cityFieldChanges)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("RegionFieldChanges: %d", regionFieldChanges)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("LastField: %s", formatCascaderValue(lastField))).Build(),
				ui.NewTextBuilder(fmt.Sprintf("LastValue: %s", formatCascaderValue(lastValue))).Build(),
			})
	}

	return appFn, initFn, cleanupFn, fixtureStore, meta
}

func TestE2ECascaderKeyboardLeafCommitSkipsDisabledOption(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newCascaderFixture()
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(96, 22), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertFocus(ByID("city-cascader")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("CityValue: <empty>")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Special(platform.KeyEnter); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("Zhejiang")); err != nil {
			return err
		}
		return app.AssertVisible(ByText("Jiangsu"))
	}); err != nil {
		t.Fatal(err)
	}

	if err := app.Driver().Special(platform.KeyEnter); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("Hangzhou")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("Shaoxing")); err != nil {
			return err
		}
		return app.AssertVisible(ByText("Ningbo"))
	}); err != nil {
		t.Fatal(err)
	}

	if err := app.Driver().Special(platform.KeyDown); err != nil {
		t.Fatal(err)
	}
	if err := app.Driver().Special(platform.KeyEnter); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertActionSequence(
		runtimeaction.ActionEnter,
		runtimeaction.ActionEnter,
		runtimeaction.ActionNavigateDown,
		runtimeaction.ActionEnter,
	); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("CityValue: zj/nb")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("CityFieldChanges: 1")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("LastField: fixture.cascader.city")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("LastValue: zj/nb")); err != nil {
			return err
		}
		return app.AssertVisible(ByText("Zhejiang / Ningbo"))
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
	if state.CityValue != "zj/nb" || state.CityFieldChanges != 1 || state.LastField != meta.CityField || state.LastValue != "zj/nb" {
		t.Fatalf("unexpected cascader fixture state after city commit: %+v", state)
	}
}

func TestE2ECascaderChangeOnSelectAndTabSkipsDisabled(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newCascaderFixture()
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(96, 22), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertFocus(ByID("city-cascader")); err != nil {
		t.Fatal(err)
	}

	app.ClearRawInputs()
	if err := app.Driver().Special(platform.KeyTab); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertFocus(ByID("region-cascader")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertActionSequence(runtimeaction.ActionNavigateNext); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Special(platform.KeyEnter); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("Zhejiang")); err != nil {
			return err
		}
		return app.AssertVisible(ByText("Jiangsu"))
	}); err != nil {
		t.Fatal(err)
	}

	if err := app.Driver().Special(platform.KeyEnter); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertActionSequence(runtimeaction.ActionEnter, runtimeaction.ActionEnter); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("RegionValue: zj")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("RegionFieldChanges: 1")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("LastField: fixture.cascader.region")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("LastValue: zj")); err != nil {
			return err
		}
		return app.AssertVisible(ByText("Zhejiang"))
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
	if err := app.AssertFocus(ByID("cascader-tail-action")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertActionSequence(runtimeaction.ActionNavigateNext); err != nil {
		t.Fatal(err)
	}

	state := fixtureStore.Get()
	if state.RegionValue != "zj" || state.RegionFieldChanges != 1 || state.CityFieldChanges != 0 || state.LastField != meta.RegionField || state.LastValue != "zj" {
		t.Fatalf("unexpected cascader fixture state after region commit: %+v", state)
	}
}

func cascaderFixtureOptions() []cascadercomp.Option {
	return []cascadercomp.Option{
		cascadercomp.Node("zj", "Zhejiang",
			cascadercomp.Leaf("hz", "Hangzhou"),
			cascadercomp.Option{Value: "sx", Label: "Shaoxing", Disabled: true},
			cascadercomp.Leaf("nb", "Ningbo"),
		),
		cascadercomp.Node("js", "Jiangsu",
			cascadercomp.Leaf("nj", "Nanjing"),
		),
	}
}

func formatCascaderValue(value string) string {
	if value == "" {
		return "<empty>"
	}
	return value
}
