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

type timePickerFixtureState struct {
	Selected string
}

type timePickerFixtureMeta struct {
	TimeField             string
	FieldChangeIntentType string
}

func newTimePickerStaticApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("TimePicker Static E2E Fixture").Build(),
				ui.NewTimePickerBuilder().
					SetID("fixture-timepicker").
					ComponentID("fixture.timepicker").
					Value("09:30").
					Width(10).
					Build(),
			})
	}
}

func newTimePickerInteractiveFixture() (ui.ComponentFunc, func(), func(), *store.Store[timePickerFixtureState], timePickerFixtureMeta) {
	fixtureStore := store.NewStore(timePickerFixtureState{Selected: "09:30"})
	meta := timePickerFixtureMeta{
		TimeField:             "fixture.timepicker.value",
		FieldChangeIntentType: runtimeintent.FieldChangeIntent{}.IntentType(),
	}
	unregisters := make([]func(), 0, 1)

	initFn := func() {
		rt := rtui.GetGlobalIntentRuntime()
		if rt == nil {
			return
		}
		unregisters = append(unregisters, runtimeintent.RegisterTypedRuntime(rt, func(_ *runtimeintent.ActionContext, i runtimeintent.FieldChangeIntent) runtimeintent.IntentResult {
			if i.Field != meta.TimeField {
				return runtimeintent.IntentResult{}
			}
			fixtureStore.Update(func(s timePickerFixtureState) timePickerFixtureState {
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
		selected := ui.UseStoreSelector(fixtureStore, func(s timePickerFixtureState) string { return s.Selected })
		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTimePickerBuilder().
					SetID("interactive-timepicker").
					ComponentID("fixture.timepicker").
					Value(selected).
					ForField(runtimeintent.BindField(meta.TimeField)).
					Width(10).
					Build(),
				ui.NewTextBuilder("SelectedTime: " + selected).Build(),
			})
	}

	return appFn, initFn, cleanupFn, fixtureStore, meta
}

func newTimePickerInputFixture() (ui.ComponentFunc, func(), func(), *store.Store[timePickerFixtureState], timePickerFixtureMeta) {
	fixtureStore := store.NewStore(timePickerFixtureState{})
	meta := timePickerFixtureMeta{
		TimeField:             "fixture.timepicker.input.value",
		FieldChangeIntentType: runtimeintent.FieldChangeIntent{}.IntentType(),
	}
	unregisters := make([]func(), 0, 1)

	initFn := func() {
		rt := rtui.GetGlobalIntentRuntime()
		if rt == nil {
			return
		}
		unregisters = append(unregisters, runtimeintent.RegisterTypedRuntime(rt, func(_ *runtimeintent.ActionContext, i runtimeintent.FieldChangeIntent) runtimeintent.IntentResult {
			if i.Field != meta.TimeField {
				return runtimeintent.IntentResult{}
			}
			fixtureStore.Update(func(s timePickerFixtureState) timePickerFixtureState {
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
		selected := ui.UseStoreSelector(fixtureStore, func(s timePickerFixtureState) string { return s.Selected })
		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTimePickerBuilder().
					SetID("input-timepicker").
					ComponentID("fixture.timepicker.input").
					Placeholder("HH:mm").
					ForField(runtimeintent.BindField(meta.TimeField)).
					Width(10).
					Build(),
				ui.NewButtonBuilder("Blur Target").SetID("blur-target").Build(),
				ui.NewTextBuilder("SelectedTime: " + selected).Build(),
			})
	}

	return appFn, initFn, cleanupFn, fixtureStore, meta
}

func TestE2ETimePickerStaticRenderAndPopupStyles(t *testing.T) {
	app, err := Run(newTimePickerStaticApp(), ui.WithSize(90, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByID("fixture-timepicker")); err != nil {
		t.Fatal(err)
	}

	if err := app.Driver().Click(ByID("fixture-timepicker")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByID("fixture-timepicker-popup")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Time")); err != nil {
		t.Fatal(err)
	}

	popupBounds, err := app.BoundsOf(ByID("fixture-timepicker-popup"))
	if err != nil {
		t.Fatal(err)
	}
	triggerBounds, err := app.BoundsOf(ByID("fixture-timepicker"))
	if err != nil {
		t.Fatal(err)
	}
	if popupBounds.X != triggerBounds.X {
		t.Fatalf("popup x = %d, want trigger x %d", popupBounds.X, triggerBounds.X)
	}
	if popupBounds.Y != triggerBounds.Y+triggerBounds.Height {
		t.Fatalf("popup y = %d, want trigger bottom %d", popupBounds.Y, triggerBounds.Y+triggerBounds.Height)
	}
	hourX, hourY := timePickerPoint(popupBounds.X, popupBounds.Y, popupBounds.Width, true, 2)
	if err := app.AssertCellStyleAt(hourX, hourY, StyleExpect{
		HasFG:   true,
		FG:      fwtheme.Text(),
		HasBG:   true,
		BG:      fwtheme.Select(),
		HasBold: true,
		Bold:    true,
	}); err != nil {
		t.Fatal(err)
	}

	minuteX, minuteY := timePickerPoint(popupBounds.X, popupBounds.Y, popupBounds.Width, false, 2)
	if err := app.AssertCellStyleAt(minuteX, minuteY, StyleExpect{
		HasFG:        true,
		FG:           fwtheme.Text(),
		HasUnderline: true,
		Underline:    true,
	}); err != nil {
		t.Fatal(err)
	}
}

func TestE2ETimePickerPopupSelectionFlow(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newTimePickerInteractiveFixture()
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(90, 24), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("SelectedTime: 09:30")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	if err := app.Driver().Click(ByID("interactive-timepicker")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByID("interactive-timepicker-popup")); err != nil {
		t.Fatal(err)
	}

	popupBounds, err := app.BoundsOf(ByID("interactive-timepicker-popup"))
	if err != nil {
		t.Fatal(err)
	}
	hourX, hourY := timePickerPoint(popupBounds.X, popupBounds.Y, popupBounds.Width, true, 4)
	if err := app.AssertHit(Point{X: hourX, Y: hourY}, ByID("interactive-timepicker-popup")); err != nil {
		t.Fatal(err)
	}
	if err := app.Driver().ClickAt(hourX, hourY); err != nil {
		t.Fatal(err)
	}

	minuteX, minuteY := timePickerPoint(popupBounds.X, popupBounds.Y, popupBounds.Width, false, 3)
	if err := app.AssertHit(Point{X: minuteX, Y: minuteY}, ByID("interactive-timepicker-popup")); err != nil {
		t.Fatal(err)
	}
	if err := app.Driver().ClickAt(minuteX, minuteY); err != nil {
		t.Fatal(err)
	}

	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("SelectedTime: 11:31"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}
	state := fixtureStore.Get()
	if state.Selected != "11:31" {
		t.Fatalf("selected = %q, want 11:31", state.Selected)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByID("interactive-timepicker-popup")); err == nil {
			return fmt.Errorf("timepicker popup still visible")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestE2ETimePickerKeyboardNavigationCommitFlow(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newTimePickerInteractiveFixture()
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(90, 24), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertFocus(ByID("interactive-timepicker")); err != nil {
		t.Fatal(err)
	}

	app.ClearRawInputs()
	app.ClearIntentLogs()

	if err := app.Driver().Special(platform.KeyEnter); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByID("interactive-timepicker-popup"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.Driver().Special(platform.KeyDown); err != nil {
		t.Fatal(err)
	}
	if err := app.Driver().Special(platform.KeyRight); err != nil {
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
		runtimeaction.ActionNavigateDown,
		runtimeaction.ActionNavigateRight,
		runtimeaction.ActionNavigateDown,
		runtimeaction.ActionEnter,
	); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertActionHandled(runtimeaction.ActionNavigateRight, "keyboard_target"); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("SelectedTime: 10:31"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}
	state := fixtureStore.Get()
	if state.Selected != "10:31" {
		t.Fatalf("selected = %q, want 10:31", state.Selected)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByID("interactive-timepicker-popup")); err == nil {
			return fmt.Errorf("timepicker popup still visible")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestE2ETimePickerDirectInputBlurNormalization(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newTimePickerInputFixture()
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(90, 24), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertFocus(ByID("input-timepicker")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("SelectedTime: ")); err != nil {
		t.Fatal(err)
	}

	app.ClearRawInputs()
	app.ClearIntentLogs()

	if err := injectKeyMsg(app, '9'); err != nil {
		t.Fatal(err)
	}
	if err := injectKeyMsg(app, ':'); err != nil {
		t.Fatal(err)
	}
	if err := injectKeyMsg(app, '5'); err != nil {
		t.Fatal(err)
	}
	if state := fixtureStore.Get(); state.Selected != "" {
		t.Fatalf("selected before blur = %q, want empty", state.Selected)
	}
	if err := app.Driver().Special(platform.KeyTab); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertFocus(ByID("blur-target")); err != nil {
		t.Fatal(err)
	}

	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("SelectedTime: 09:05"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}
	state := fixtureStore.Get()
	if state.Selected != "09:05" {
		t.Fatalf("selected after blur = %q, want 09:05", state.Selected)
	}
}

func timePickerPoint(popupX, popupY, popupWidth int, hourColumn bool, row int) (int, int) {
	contentWidth := popupWidth - 2
	start := 1 + maxInt(0, (contentWidth-5)/2)
	colX := start
	if !hourColumn {
		colX = start + 3
	}
	return popupX + colX, popupY + 3 + row
}
