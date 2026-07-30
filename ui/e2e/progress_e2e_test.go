package e2e

import (
	"fmt"
	"testing"
	"time"

	fwtheme "github.com/wwsheng009/mint/framework/theme"
	runtimeintent "github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
	progresscomp "github.com/wwsheng009/mint/ui/components/progress"
)

type progressCycleAccentIntent struct{}

func (progressCycleAccentIntent) IntentType() string { return "e2e.progress.cycle_accent" }

type progressToggleCustomIntent struct{}

func (progressToggleCustomIntent) IntentType() string { return "e2e.progress.toggle_custom" }

type progressFixtureState struct {
	AccentIndex    int
	UseCustomColor bool
}

var progressFixtureAccents = []style.Color{
	style.Cyan,
	style.Yellow,
	style.Magenta,
}

func newProgressStaticApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Progress E2E Fixture").Build(),
				progresscomp.NewBuilder().
					SetID("progress-success-line").
					Value(60).
					Max(100).
					Width(12).
					Label("Uploading").
					Success().
					Build(),
				progresscomp.NewBuilder().
					SetID("progress-warning-line").
					Value(85).
					Max(100).
					Width(12).
					Label("Quota").
					Warning().
					Build(),
				progresscomp.NewBuilder().
					SetID("progress-value-line").
					Value(7).
					Max(10).
					Width(12).
					Label("Queue").
					ShowValue(true).
					Unit("jobs").
					Warning().
					Build(),
				progresscomp.NewBuilder().
					SetID("progress-block").
					Value(50).
					Max(100).
					Width(12).
					Label("Packed").
					Block().
					Build(),
				progresscomp.NewBuilder().
					SetID("progress-active-line").
					Value(50).
					Max(100).
					Width(12).
					Label("Syncing").
					ShowPercent(false).
					Active().
					Build(),
				progresscomp.NewBuilder().
					SetID("progress-indeterminate").
					Width(12).
					Label("Resolving").
					Indeterminate().
					Build(),
				progresscomp.NewBuilder().
					SetID("progress-circle").
					Value(100).
					Max(100).
					Width(5).
					Circle().
					Build(),
				progresscomp.NewBuilder().
					SetID("progress-dashboard").
					Value(100).
					Max(100).
					Width(7).
					Label("CPU").
					ShowPercent(false).
					Dashboard().
					Build(),
			})
	}
}

func newProgressInteractiveFixture() (ui.ComponentFunc, func(), func(), *store.Store[progressFixtureState]) {
	fixtureStore := store.NewStore(progressFixtureState{})
	unregisters := make([]func(), 0, 2)

	initFn := func() {
		rt := rtui.GetGlobalIntentRuntime()
		if rt == nil {
			return
		}
		unregisters = append(unregisters,
			runtimeintent.RegisterTypedRuntime(rt, func(_ *runtimeintent.ActionContext, _ progressCycleAccentIntent) runtimeintent.IntentResult {
				fixtureStore.Update(func(s progressFixtureState) progressFixtureState {
					s.AccentIndex = (s.AccentIndex + 1) % len(progressFixtureAccents)
					s.UseCustomColor = true
					return s
				})
				return runtimeintent.HandledResult()
			}),
			runtimeintent.RegisterTypedRuntime(rt, func(_ *runtimeintent.ActionContext, _ progressToggleCustomIntent) runtimeintent.IntentResult {
				fixtureStore.Update(func(s progressFixtureState) progressFixtureState {
					s.UseCustomColor = !s.UseCustomColor
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
		state := ui.UseStoreSelector(fixtureStore, func(s progressFixtureState) progressFixtureState { return s })
		progressStyle := style.Style{}
		if state.UseCustomColor {
			progressStyle = style.Style{}.Foreground(progressFixtureAccents[state.AccentIndex]).Bold(true)
		}

		mode := "semantic"
		if state.UseCustomColor {
			mode = "custom"
		}

		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Progress Interactive E2E Fixture").Build(),
				progresscomp.NewBuilder().
					SetID("progress-preview-line").
					Value(60).
					Max(100).
					Width(12).
					Label("Preview").
					Status(progresscomp.StatusNormal).
					Style(progressStyle).
					Build(),
				progresscomp.NewBuilder().
					SetID("progress-preview-block").
					Value(60).
					Max(100).
					Width(12).
					Label("Preview Block").
					Block().
					Status(progresscomp.StatusNormal).
					Style(progressStyle).
					Build(),
				ui.NewButtonBuilder("Cycle Color").
					SetID("progress-cycle-color").
					OnPress(progressCycleAccentIntent{}).
					Build(),
				ui.NewButtonBuilder("Toggle Custom Color").
					SetID("progress-toggle-custom").
					OnPress(progressToggleCustomIntent{}).
					Build(),
				ui.NewTextBuilder(fmt.Sprintf("Mode: %s", mode)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("AccentIndex: %d", state.AccentIndex)).Build(),
			})
	}

	return appFn, initFn, cleanupFn, fixtureStore
}

func TestE2EProgressLineBlockCircleAndDashboardRender(t *testing.T) {
	app, err := Run(newProgressStaticApp(), ui.WithSize(96, 28))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("[━━━━━━····]")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Uploading: 60%")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Quota: 85%")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Queue: 7/10 jobs (70%)")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("[█████░░░░░]")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Packed: 50%")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Syncing")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByID("progress-indeterminate")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Resolving: ...")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText(" ███ ")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("█   █")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("100%")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText(" █████ ")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("█     █")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("CPU")); err != nil {
		t.Fatal(err)
	}
}

func TestE2EProgressStatusStylesRender(t *testing.T) {
	app, err := Run(newProgressStaticApp(), ui.WithSize(96, 28))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertStyle(ByText("[━━━━━━····]"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Success(),
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("Quota: 85%"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Warning(),
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("Syncing"), StyleExpect{
		HasFG:   true,
		FG:      fwtheme.Focus(),
		HasBold: true,
		Bold:    true,
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("Resolving: ..."), StyleExpect{
		HasFG:   true,
		FG:      fwtheme.Focus(),
		HasBold: true,
		Bold:    true,
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("100%"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Primary(),
	}); err != nil {
		t.Fatal(err)
	}
}

func TestE2EProgressColorCycleAndToggleFlow(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore := newProgressInteractiveFixture()
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(96, 24), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("Mode: semantic")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("Preview: 60%"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Primary(),
	}); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Click(ByID("progress-cycle-color")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("Mode: custom")); err != nil {
			return err
		}
		if err := app.AssertVisible(ByText("AccentIndex: 1")); err != nil {
			return err
		}
		return app.AssertStyle(ByText("Preview: 60%"), StyleExpect{
			HasFG:   true,
			FG:      style.Yellow,
			HasBold: true,
			Bold:    true,
		})
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(progressCycleAccentIntent{}.IntentType()); err != nil {
		t.Fatal(err)
	}
	if state := fixtureStore.Get(); state.AccentIndex != 1 || !state.UseCustomColor {
		t.Fatalf("unexpected fixture state after cycle color: %+v", state)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Click(ByID("progress-toggle-custom")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("Mode: semantic")); err != nil {
			return err
		}
		return app.AssertStyle(ByText("Preview: 60%"), StyleExpect{
			HasFG:   true,
			FG:      fwtheme.Primary(),
			HasBold: true,
			Bold:    false,
		})
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(progressToggleCustomIntent{}.IntentType()); err != nil {
		t.Fatal(err)
	}
	if state := fixtureStore.Get(); state.AccentIndex != 1 || state.UseCustomColor {
		t.Fatalf("unexpected fixture state after toggle custom off: %+v", state)
	}
}
