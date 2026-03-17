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

type stepsFixtureState struct {
	ObservedCurrent int
	ChangeCount     int
	LastTitle       string
}

type stepsFixtureMeta struct {
	ID                    string
	ComponentID           string
	CurrentField          string
	FieldChangeIntentType string
}

func newStepsFixture(id, componentID string, items []ui.StepsItem, direction ui.StepsDirection, progressDot bool, percent int) (ui.ComponentFunc, func(), func(), *store.Store[stepsFixtureState], stepsFixtureMeta) {
	fixtureStore := store.NewStore(stepsFixtureState{})
	meta := stepsFixtureMeta{
		ID:                    id,
		ComponentID:           componentID,
		CurrentField:          componentID + ".current",
		FieldChangeIntentType: runtimeintent.FieldChangeIntent{}.IntentType(),
	}

	unregisters := make([]func(), 0, 1)
	initFn := func() {
		rt := rtui.GetGlobalIntentRuntime()
		if rt == nil {
			return
		}
		unregisters = append(unregisters, runtimeintent.RegisterTypedRuntime(rt, func(_ *runtimeintent.ActionContext, i runtimeintent.FieldChangeIntent) runtimeintent.IntentResult {
			if i.Field != meta.CurrentField {
				return runtimeintent.IntentResult{}
			}
			nextIndex, err := strconv.Atoi(i.Value)
			if err != nil {
				return runtimeintent.ErrorResult(err)
			}
			if nextIndex < 0 || nextIndex >= len(items) {
				return runtimeintent.ErrorResult(fmt.Errorf("steps field index %d out of range", nextIndex))
			}
			fixtureStore.Update(func(s stepsFixtureState) stepsFixtureState {
				s.ObservedCurrent = nextIndex
				s.ChangeCount++
				s.LastTitle = items[nextIndex].Title
				return s
			})
			return runtimeintent.HandledResult()
		}))
	}

	cleanupFn := func() {
		rt := rtui.GetGlobalIntentRuntime()
		for i := len(unregisters) - 1; i >= 0; i-- {
			if unregisters[i] != nil {
				unregisters[i]()
			}
		}
		if rt != nil {
			runtimeintent.SetupBuiltinHandlers(rt)
		}
	}

	appFn := func() ui.VNode {
		observedCurrent := ui.UseStoreSelector(fixtureStore, func(s stepsFixtureState) int { return s.ObservedCurrent })
		changeCount := ui.UseStoreSelector(fixtureStore, func(s stepsFixtureState) int { return s.ChangeCount })
		lastTitle := ui.UseStoreSelector(fixtureStore, func(s stepsFixtureState) string { return s.LastTitle })

		builder := ui.NewStepsBuilder().
			SetID(meta.ID).
			ComponentID(meta.ComponentID).
			Items(items).
			Current(observedCurrent).
			Direction(direction).
			ProgressDot(progressDot).
			Percent(percent).
			CurrentForField(runtimeintent.BindField(meta.CurrentField))

		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Steps E2E Fixture").Build(),
				builder.Build(),
				ui.NewTextBuilder(fmt.Sprintf("Observed: %d", observedCurrent)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("ChangeCount: %d", changeCount)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("LastTitle: %s", lastTitle)).Build(),
			})
	}

	return appFn, initFn, cleanupFn, fixtureStore, meta
}

func TestE2EStepsKeyboardNavigationFlow(t *testing.T) {
	items := []ui.StepsItem{
		ui.NewStepsItem("Cart"),
		ui.NewStepsItem("Address"),
		ui.NewStepsItem("Pay"),
	}
	appFn, initFn, cleanupFn, fixtureStore, meta := newStepsFixture("checkout-steps", "fixture.steps.horizontal", items, ui.StepsHorizontal, true, 42)
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(80, 14), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertFocus(ByComponentID(meta.ComponentID)); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Cart")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("Cart"), StyleExpect{
		HasBold: true,
		Bold:    true,
	}); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Special(platform.KeyRight); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("Observed: 1"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "key:Right"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentSequence(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("ChangeCount: 1")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("LastTitle: Address")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertFocus(ByComponentID(meta.ComponentID)); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("Address"), StyleExpect{
		HasBold: true,
		Bold:    true,
	}); err != nil {
		t.Fatal(err)
	}

	state := fixtureStore.Get()
	if state.ObservedCurrent != 1 || state.ChangeCount != 1 || state.LastTitle != "Address" {
		t.Fatalf("unexpected steps fixture state after keyboard navigation: %+v", state)
	}
}

func TestE2EStepsVerticalClickFlow(t *testing.T) {
	items := []ui.StepsItem{
		ui.NewStepsItem("Draft").WithDescription("draft docs"),
		ui.NewStepsItem("Review").WithDescription("review pending"),
		ui.NewStepsItem("Publish").WithDescription("ship now"),
	}
	appFn, initFn, cleanupFn, fixtureStore, meta := newStepsFixture("release-steps", "fixture.steps.vertical", items, ui.StepsVertical, false, -1)
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(80, 18), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertFocus(ByComponentID(meta.ComponentID)); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("review pending")); err != nil {
		t.Fatal(err)
	}

	point, err := app.ResolvePoint(ByText("Publish"))
	if err != nil {
		t.Fatal(err)
	}
	if err := app.AssertHit(point, ByID(meta.ID)); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Click(ByText("Publish")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("Observed: 2"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "mouse:left:release"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("ChangeCount: 1")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("LastTitle: Publish")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("Publish"), StyleExpect{
		HasBold: true,
		Bold:    true,
	}); err != nil {
		t.Fatal(err)
	}

	state := fixtureStore.Get()
	if state.ObservedCurrent != 2 || state.ChangeCount != 1 || state.LastTitle != "Publish" {
		t.Fatalf("unexpected steps fixture state after click: %+v", state)
	}
}
