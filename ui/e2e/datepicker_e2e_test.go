package e2e

import (
	"fmt"
	"testing"
	"time"

	fwtheme "github.com/wwsheng009/mint/framework/theme"
	runtimeaction "github.com/wwsheng009/mint/runtime/action"
	runtimeintent "github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/runtime/store"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
)

type datePickerFixtureState struct {
	Selected string
}

type datePickerFixtureMeta struct {
	DateField             string
	FieldChangeIntentType string
}

func newDatePickerStaticApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("DatePicker Static E2E Fixture").Build(),
				ui.NewDatePickerBuilder().
					SetID("fixture-datepicker").
					ComponentID("fixture.datepicker").
					Value("2026-03-18").
					Width(18).
					Build(),
			})
	}
}

func newDatePickerInteractiveFixture() (ui.ComponentFunc, func(), func(), *store.Store[datePickerFixtureState], datePickerFixtureMeta) {
	fixtureStore := store.NewStore(datePickerFixtureState{Selected: "2026-04-05"})
	meta := datePickerFixtureMeta{
		DateField:             "fixture.datepicker.value",
		FieldChangeIntentType: runtimeintent.FieldChangeIntent{}.IntentType(),
	}
	unregisters := make([]func(), 0, 1)

	initFn := func() {
		rt := rtui.GetGlobalIntentRuntime()
		if rt == nil {
			return
		}
		unregisters = append(unregisters, runtimeintent.RegisterTypedRuntime(rt, func(_ *runtimeintent.ActionContext, i runtimeintent.FieldChangeIntent) runtimeintent.IntentResult {
			if i.Field != meta.DateField {
				return runtimeintent.IntentResult{}
			}
			fixtureStore.Update(func(s datePickerFixtureState) datePickerFixtureState {
				s.Selected = i.Value
				return s
			})
			return runtimeintent.HandledResult()
		}))
	}

	cleanupFn := func() {
		for index := len(unregisters) - 1; index >= 0; index-- {
			if unregisters[index] != nil {
				unregisters[index]()
			}
		}
	}

	appFn := func() ui.VNode {
		selected := ui.UseStoreSelector(fixtureStore, func(s datePickerFixtureState) string { return s.Selected })
		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewDatePickerBuilder().
					SetID("interactive-datepicker").
					ComponentID("fixture.datepicker").
					Value(selected).
					ForField(runtimeintent.BindField(meta.DateField)).
					Width(18).
					Build(),
				ui.NewTextBuilder("SelectedDate: " + selected).Build(),
			})
	}

	return appFn, initFn, cleanupFn, fixtureStore, meta
}

func newDatePickerInputFixture() (ui.ComponentFunc, func(), func(), *store.Store[datePickerFixtureState], datePickerFixtureMeta) {
	fixtureStore := store.NewStore(datePickerFixtureState{})
	meta := datePickerFixtureMeta{
		DateField:             "fixture.datepicker.input.value",
		FieldChangeIntentType: runtimeintent.FieldChangeIntent{}.IntentType(),
	}
	unregisters := make([]func(), 0, 1)

	initFn := func() {
		rt := rtui.GetGlobalIntentRuntime()
		if rt == nil {
			return
		}
		unregisters = append(unregisters, runtimeintent.RegisterTypedRuntime(rt, func(_ *runtimeintent.ActionContext, i runtimeintent.FieldChangeIntent) runtimeintent.IntentResult {
			if i.Field != meta.DateField {
				return runtimeintent.IntentResult{}
			}
			fixtureStore.Update(func(s datePickerFixtureState) datePickerFixtureState {
				s.Selected = i.Value
				return s
			})
			return runtimeintent.HandledResult()
		}))
	}

	cleanupFn := func() {
		for index := len(unregisters) - 1; index >= 0; index-- {
			if unregisters[index] != nil {
				unregisters[index]()
			}
		}
	}

	appFn := func() ui.VNode {
		selected := ui.UseStoreSelector(fixtureStore, func(s datePickerFixtureState) string { return s.Selected })
		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewDatePickerBuilder().
					SetID("input-datepicker").
					ComponentID("fixture.datepicker.input").
					Placeholder("YYYY-MM-DD").
					ForField(runtimeintent.BindField(meta.DateField)).
					Width(18).
					Build(),
				ui.NewButtonBuilder("Blur Target").SetID("date-blur-target").Build(),
				ui.NewTextBuilder("SelectedDate: " + selected).Build(),
			})
	}

	return appFn, initFn, cleanupFn, fixtureStore, meta
}

func TestE2EDatePickerStaticRenderAndPopupStyles(t *testing.T) {
	app, err := Run(newDatePickerStaticApp(), ui.WithSize(90, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByID("fixture-datepicker")); err != nil {
		t.Fatal(err)
	}

	if err := app.Driver().Click(ByID("fixture-datepicker")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByID("fixture-datepicker-popup")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("2026-03")); err != nil {
		t.Fatal(err)
	}

	popupBounds, err := app.BoundsOf(ByID("fixture-datepicker-popup"))
	if err != nil {
		t.Fatal(err)
	}
	cellX, cellY := dateCellPoint(time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, time.March, 18, 0, 0, 0, 0, time.UTC), popupBounds.X, popupBounds.Y)
	if err := app.AssertCellStyleAt(cellX, cellY, StyleExpect{
		HasFG: true,
		FG:    fwtheme.BG(),
		HasBG: true,
		BG:    fwtheme.Primary(),
	}); err != nil {
		t.Fatal(err)
	}
}

func TestE2EDatePickerPopupSelectionFlow(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newDatePickerInteractiveFixture()
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(90, 24), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("SelectedDate: 2026-04-05")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	if err := app.Driver().Click(ByID("interactive-datepicker")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("2026-04")); err != nil {
		t.Fatal(err)
	}

	popupBounds, err := app.BoundsOf(ByID("interactive-datepicker-popup"))
	if err != nil {
		t.Fatal(err)
	}
	cellX, cellY := dateCellPoint(time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, time.April, 6, 0, 0, 0, 0, time.UTC), popupBounds.X, popupBounds.Y)
	if err := app.AssertHit(Point{X: cellX, Y: cellY}, ByID("interactive-datepicker-popup")); err != nil {
		t.Fatal(err)
	}
	if err := app.Driver().ClickAt(cellX, cellY); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("SelectedDate: 2026-04-06"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}
	state := fixtureStore.Get()
	if state.Selected != "2026-04-06" {
		t.Fatalf("selected = %q, want 2026-04-06", state.Selected)
	}

	if err := app.Driver().Click(ByID("interactive-datepicker")); err != nil {
		t.Fatal(err)
	}
	popupBounds, err = app.BoundsOf(ByID("interactive-datepicker-popup"))
	if err != nil {
		t.Fatal(err)
	}
	if err := app.Driver().ClickAt(popupBounds.X+popupBounds.Width-2, popupBounds.Y+1); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("2026-05")); err != nil {
		t.Fatal(err)
	}
}

func TestE2EDatePickerKeyboardNavigationCommitFlow(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newDatePickerInteractiveFixture()
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(90, 24), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertFocus(ByID("interactive-datepicker")); err != nil {
		t.Fatal(err)
	}

	app.ClearRawInputs()
	app.ClearIntentLogs()

	if err := app.Driver().Special(platform.KeyEnter); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByID("interactive-datepicker-popup"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.Driver().Special(platform.KeyPageDown); err != nil {
		t.Fatal(err)
	}
	if err := app.Driver().Special(platform.KeyEnter); err != nil {
		t.Fatal(err)
	}

	if err := app.AssertActionSequence(
		runtimeaction.ActionEnter,
		runtimeaction.ActionNavigatePageDown,
		runtimeaction.ActionEnter,
	); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertActionHandled(runtimeaction.ActionNavigatePageDown, "keyboard_target"); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("SelectedDate: 2026-05-05"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}
	state := fixtureStore.Get()
	if state.Selected != "2026-05-05" {
		t.Fatalf("selected = %q, want 2026-05-05", state.Selected)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByID("interactive-datepicker-popup")); err == nil {
			return fmt.Errorf("datepicker popup still visible")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestE2EDatePickerDirectInputCommitFlow(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newDatePickerInputFixture()
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(90, 24), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertFocus(ByID("input-datepicker")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	if err := injectTextMsg(app, "2026-04-06"); err != nil {
		t.Fatal(err)
	}
	if err := app.Driver().Special(platform.KeyTab); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertFocus(ByID("date-blur-target")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("SelectedDate: 2026-04-06"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}
	state := fixtureStore.Get()
	if state.Selected != "2026-04-06" {
		t.Fatalf("selected = %q, want 2026-04-06", state.Selected)
	}
}

func dateCellPoint(month, target time.Time, popupX, popupY int) (int, int) {
	start := monthStartForTest(month)
	firstWeekday := weekdayIndexMondayForTest(start.Weekday())
	gridStart := start.AddDate(0, 0, -firstWeekday)
	diff := int(target.Sub(gridStart).Hours() / 24)
	row := diff / 7
	col := diff % 7
	return popupX + 1 + col*3 + 1, popupY + 3 + row
}

func monthStartForTest(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.UTC)
}

func weekdayIndexMondayForTest(weekday time.Weekday) int {
	if weekday == time.Sunday {
		return 6
	}
	return int(weekday - time.Monday)
}
