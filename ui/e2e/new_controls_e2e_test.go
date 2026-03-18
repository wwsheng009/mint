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
	popconfirmcomp "github.com/wwsheng009/mint/ui/components/popconfirm"
)

type newControlsFixtureState struct {
	CollapseField     string
	PopoverVisible    bool
	PopconfirmVisible bool
	PopconfirmOpen    string
	PopconfirmOK      int
	PopconfirmCancel  int
}

type newControlsFixtureMeta struct {
	CollapseField         string
	PopoverField          string
	PopoverComponentID    string
	PopconfirmField       string
	PopconfirmComponentID string
	FieldChangeIntentType string
	PopconfirmConfirmType string
	PopconfirmCancelType  string
}

type togglePopoverFixtureIntent struct{}
type openPopconfirmFixtureIntent struct{}

func (togglePopoverFixtureIntent) IntentType() string  { return "e2e.toggle_popover_fixture" }
func (openPopconfirmFixtureIntent) IntentType() string { return "e2e.open_popconfirm_fixture" }

func newStaticControlsApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("New Controls Static E2E Fixture").Build(),
				ui.NewDescriptionsBuilder().
					SetID("fixture-descriptions").
					Title("Build Info").
					Column(2).
					LabelStyle(style.NewStyle().Foreground(style.Cyan).Bold(true)).
					Item(ui.NewDescriptionsItem("Version", ui.Text("v1.2.3"))).
					Item(ui.NewDescriptionsItem("Commit", ui.Text("9e356fc0")).WithSpan(2)).
					Build(),
				ui.NewStatisticBuilder().
					SetID("fixture-statistic").
					Title("Revenue").
					Prefix("$").
					Value(12345.67).
					Precision(2).
					Up().
					ValueStyle(style.NewStyle().Foreground(style.Green).Bold(true)).
					Build(),
				ui.NewTimelineBuilder().
					SetID("fixture-timeline").
					Item(ui.NewTimelineItem("Deploy queued").WithLabel("09:30").WithStatus(ui.TimelineStatusWarning)).
					Item(ui.NewTimelineItem("Smoke tests").WithLabel("09:45").WithDescription("Awaiting approval")).
					Pending("Waiting for release manager").
					Build(),
				ui.NewResultBuilder().
					SetID("fixture-result").
					Status(ui.ResultStatus404).
					Icon("⛔").
					Subtitle("Requested resource missing.").
					Build(),
			})
	}
}

func newControlsFixture() (ui.ComponentFunc, func(), func(), *store.Store[newControlsFixtureState], newControlsFixtureMeta) {
	fixtureStore := store.NewStore(newControlsFixtureState{})
	meta := newControlsFixtureMeta{
		CollapseField:         "fixture.collapse.active",
		PopoverField:          "fixture.popover.open",
		PopoverComponentID:    "fixture.popover",
		PopconfirmField:       "fixture.popconfirm.open",
		PopconfirmComponentID: "fixture.popconfirm",
		FieldChangeIntentType: runtimeintent.FieldChangeIntent{}.IntentType(),
		PopconfirmConfirmType: popconfirmcomp.PopconfirmConfirmIntent{}.IntentType(),
		PopconfirmCancelType:  popconfirmcomp.PopconfirmCancelIntent{}.IntentType(),
	}

	unregisters := make([]func(), 0, 5)
	initFn := func() {
		rt := rtui.GetGlobalIntentRuntime()
		if rt == nil {
			return
		}
		unregisters = append(unregisters,
			runtimeintent.RegisterTypedRuntime(rt, func(_ *runtimeintent.ActionContext, i runtimeintent.FieldChangeIntent) runtimeintent.IntentResult {
				switch i.Field {
				case meta.CollapseField:
					fixtureStore.Update(func(s newControlsFixtureState) newControlsFixtureState {
						s.CollapseField = i.Value
						return s
					})
					return runtimeintent.HandledResult()
				case meta.PopconfirmField:
					fixtureStore.Update(func(s newControlsFixtureState) newControlsFixtureState {
						s.PopconfirmOpen = i.Value
						s.PopconfirmVisible = i.Value == "true"
						return s
					})
					return runtimeintent.HandledResult()
				default:
					return runtimeintent.IntentResult{}
				}
			}),
			runtimeintent.RegisterTypedRuntime(rt, func(_ *runtimeintent.ActionContext, _ togglePopoverFixtureIntent) runtimeintent.IntentResult {
				fixtureStore.Update(func(s newControlsFixtureState) newControlsFixtureState {
					s.PopoverVisible = !s.PopoverVisible
					return s
				})
				return runtimeintent.HandledResult()
			}),
			runtimeintent.RegisterTypedRuntime(rt, func(_ *runtimeintent.ActionContext, _ openPopconfirmFixtureIntent) runtimeintent.IntentResult {
				fixtureStore.Update(func(s newControlsFixtureState) newControlsFixtureState {
					s.PopconfirmVisible = !s.PopconfirmVisible
					s.PopconfirmOpen = fmt.Sprintf("%t", s.PopconfirmVisible)
					return s
				})
				return runtimeintent.HandledResult()
			}),
			runtimeintent.RegisterTypedRuntime(rt, func(_ *runtimeintent.ActionContext, i popconfirmcomp.PopconfirmConfirmIntent) runtimeintent.IntentResult {
				if i.ComponentID != meta.PopconfirmComponentID {
					return runtimeintent.IntentResult{}
				}
				fixtureStore.Update(func(s newControlsFixtureState) newControlsFixtureState {
					s.PopconfirmOK++
					return s
				})
				return runtimeintent.HandledResult()
			}),
			runtimeintent.RegisterTypedRuntime(rt, func(_ *runtimeintent.ActionContext, i popconfirmcomp.PopconfirmCancelIntent) runtimeintent.IntentResult {
				if i.ComponentID != meta.PopconfirmComponentID {
					return runtimeintent.IntentResult{}
				}
				fixtureStore.Update(func(s newControlsFixtureState) newControlsFixtureState {
					s.PopconfirmCancel++
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
		collapseField := ui.UseStoreSelector(fixtureStore, func(s newControlsFixtureState) string { return s.CollapseField })
		popoverVisible := ui.UseStoreSelector(fixtureStore, func(s newControlsFixtureState) bool { return s.PopoverVisible })
		popconfirmVisible := ui.UseStoreSelector(fixtureStore, func(s newControlsFixtureState) bool { return s.PopconfirmVisible })
		popconfirmOpen := ui.UseStoreSelector(fixtureStore, func(s newControlsFixtureState) string { return s.PopconfirmOpen })
		popconfirmOK := ui.UseStoreSelector(fixtureStore, func(s newControlsFixtureState) int { return s.PopconfirmOK })
		popconfirmCancel := ui.UseStoreSelector(fixtureStore, func(s newControlsFixtureState) int { return s.PopconfirmCancel })

		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("New Controls Interaction E2E Fixture").Build(),
				ui.NewCollapseBuilder().
					SetID("fixture-collapse").
					ComponentID("fixture.collapse").
					Item(ui.NewCollapseItem("Overview Panel", ui.Text("Overview body")).WithKey("overview")).
					Item(ui.NewCollapseItem("Advanced Panel", ui.Text("Advanced details")).WithKey("advanced")).
					InitialActiveKeys("overview").
					ActiveKeysForField(runtimeintent.BindField(meta.CollapseField)).
					Build(),
				ui.NewVStack().
					SetGap(1).
					SetChildrenList([]ui.VNode{
						ui.NewHStack().
							SetGap(2).
							SetChildrenList([]ui.VNode{
								ui.NewButtonBuilder("Toggle Popover").SetID("toggle-popover").OnPress(togglePopoverFixtureIntent{}).Build(),
								ui.NewPopoverBuilder(
									ui.NewTextBuilder("Popover anchor").SetID("popover-anchor").Build(),
								).
									SetID("fixture-popover").
									ComponentID(meta.PopoverComponentID).
									Title("Popover Title").
									Body("Popover body text").
									Trigger(ui.PopoverTriggerManual).
									Open(popoverVisible).
									Build(),
							}),
						ui.NewHStack().
							SetGap(2).
							SetChildrenList([]ui.VNode{
								ui.NewButtonBuilder("Open Confirm").SetID("open-popconfirm").OnPress(openPopconfirmFixtureIntent{}).Build(),
								ui.NewPopconfirmBuilder(
									ui.NewTextBuilder("Delete anchor").SetID("popconfirm-anchor").Build(),
								).
									SetID("fixture-popconfirm").
									ComponentID(meta.PopconfirmComponentID).
									Title("Delete record?").
									Description("Cannot undo.").
									CancelText("No").
									OkText("Yes").
									Trigger(ui.PopconfirmTriggerManual).
									Open(popconfirmVisible).
									Placement(ui.PopconfirmPlacementBottomLeft).
									OpenForField(runtimeintent.BindField(meta.PopconfirmField)).
									Build(),
							}),
					}),
				ui.NewTextBuilder(fmt.Sprintf("CollapseField: %s", collapseField)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("PopoverVisible: %t", popoverVisible)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("PopconfirmOpen: %s", popconfirmOpen)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("PopconfirmOK: %d", popconfirmOK)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("PopconfirmCancel: %d", popconfirmCancel)).Build(),
			})
	}

	return appFn, initFn, cleanupFn, fixtureStore, meta
}

func TestE2ENewControlsStaticRenderAndStyles(t *testing.T) {
	app, err := Run(newStaticControlsApp(), ui.WithSize(110, 28))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("Build Info")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("Version:"), StyleExpect{
		HasFG:   true,
		FG:      style.Cyan,
		HasBold: true,
		Bold:    true,
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("12,345.67")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("12,345.67"), StyleExpect{
		HasFG:   true,
		FG:      style.Green,
		HasBold: true,
		Bold:    true,
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("▲"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Warning(),
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Waiting for release manager")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("404 Not Found")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("⛔"), StyleExpect{
		HasFG:   true,
		FG:      fwtheme.Primary(),
		HasBold: true,
		Bold:    true,
	}); err != nil {
		t.Fatal(err)
	}
}

func TestE2ECollapseAndPopoverInteractionFlow(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newControlsFixture()
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(110, 40), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("Overview body")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Click(ByText("Advanced Panel")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("Advanced details"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("CollapseField: overview,advanced")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "mouse:left:release"}); err != nil {
		t.Fatal(err)
	}

	state := fixtureStore.Get()
	if state.CollapseField != "overview,advanced" {
		t.Fatalf("unexpected collapse state after click: %+v", state)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Click(ByID("toggle-popover")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("Popover body text"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("PopoverVisible: true")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	if err := app.Driver().Click(ByID("toggle-popover")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("Popover body text")); err == nil {
			return fmt.Errorf("popover body still visible")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("PopoverVisible: false")); err != nil {
		t.Fatal(err)
	}

	state = fixtureStore.Get()
	if state.PopoverVisible {
		t.Fatalf("unexpected popover state after close: %+v", state)
	}
}

func TestE2EPopconfirmOpenCloseFlow(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, _ := newControlsFixture()
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(110, 40), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.Driver().Click(ByID("open-popconfirm")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("Delete record?"))
	}); err != nil {
		t.Fatal(err)
	}
	state := fixtureStore.Get()
	if state.PopconfirmOpen != "true" || !state.PopconfirmVisible {
		t.Fatalf("expected popconfirm to be open after trigger: %+v", state)
	}

	if err := app.Driver().Click(ByID("open-popconfirm")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("Delete record?")); err == nil {
			return fmt.Errorf("popconfirm still visible after external close")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	state = fixtureStore.Get()
	if state.PopconfirmOpen != "false" || state.PopconfirmVisible {
		t.Fatalf("unexpected popconfirm state after close flow: %+v", state)
	}
}
