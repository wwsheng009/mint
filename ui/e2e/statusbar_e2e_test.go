package e2e

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/wwsheng009/mint/runtime/intent"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
	statusbarcomp "github.com/wwsheng009/mint/ui/components/statusbar"
)

type statusbarActivateIntent struct{ Token string }
type statusbarNeutralIntent struct{}

func (i statusbarActivateIntent) IntentType() string {
	return "e2e.statusbar.activate." + i.Token
}

func (statusbarNeutralIntent) IntentType() string { return "e2e.statusbar.neutral" }

type statusbarFixtureMeta struct {
	ActivateIntentType string
	HelpText           string
}

var statusbarFixtureSeq int64

func newStatusbarFixture(displayMode statusbarcomp.HelpDisplayMode) (ui.ComponentFunc, func(), func(), statusbarFixtureMeta) {
	token := fmt.Sprintf("%d", atomic.AddInt64(&statusbarFixtureSeq, 1))
	activateIntent := statusbarActivateIntent{Token: token}
	meta := statusbarFixtureMeta{
		ActivateIntentType: activateIntent.IntentType(),
		HelpText:           "Inspect active details",
	}

	unregisters := make([]func(), 0, 1)
	initFn := func() {
		rt := rtui.GetGlobalIntentRuntime()
		if rt == nil {
			return
		}
		unregisters = append(unregisters,
			rt.Register(meta.ActivateIntentType, intent.HandlerFunc(func(_ *intent.ActionContext, _ intent.Intent) intent.IntentResult {
				return intent.HandledResult()
			})),
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
		children := make([]ui.VNode, 0, 10)
		children = append(children,
			ui.NewTextBuilder("Statusbar E2E Fixture").Build(),
			ui.NewButtonBuilder("Neutral Action").SetID("statusbar-top-action").OnPress(statusbarNeutralIntent{}).Build(),
			ui.NewTextBuilder("Workspace: alpha").Build(),
			ui.NewTextBuilder("Queue: healthy").Build(),
			ui.NewTextBuilder("Agent: mint").Build(),
			ui.NewTextBuilder("Branch: overlay").Build(),
			ui.NewTextBuilder("Viewport: compact").Build(),
			ui.NewTextBuilder("Runtime: idle").Build(),
			ui.NewTextBuilder("Focus: anchored").Build(),
			statusbarcomp.NewBuilder().
				DefaultTheme().
				HelpFallback("Ready").
				HelpDisplayMode(displayMode).
				TooltipPlacement(statusbarcomp.TooltipPlacementAuto).
				Left(statusbarcomp.Text("Help").WithKey("statusbar-help").WithHelp(meta.HelpText)).
				Left(statusbarcomp.ActionText("Run", activateIntent).WithKey("statusbar-activate")).
				Center(statusbarcomp.Text("Status: green")).
				Right(statusbarcomp.Text("F1 Help")).
				BuildWithHelp(),
		)

		return ui.NewVStack().
			SetGap(0).
			SetChildrenList(children)
	}

	return appFn, initFn, cleanupFn, meta
}

func TestE2EStatusbarOverlayHelpTracksHoverPlacementAndHide(t *testing.T) {
	appFn, initFn, cleanupFn, meta := newStatusbarFixture(statusbarcomp.HelpDisplayOverlay)
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(72, 12), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.Driver().Move(ByID("statusbar-top-action")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText(meta.HelpText)); err == nil {
		t.Fatalf("overlay help %q should be hidden before hover", meta.HelpText)
	}

	anchorBounds, err := app.BoundsOf(ByKey("statusbar-help"))
	if err != nil {
		t.Fatal(err)
	}

	app.ClearRawInputs()
	if err := app.Driver().Move(ByKey("statusbar-help")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText(meta.HelpText))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "mouse:none:move"}); err != nil {
		t.Fatal(err)
	}

	overlayPoint, err := app.ResolvePoint(ByText(meta.HelpText))
	if err != nil {
		t.Fatal(err)
	}
	if overlayPoint.Y >= anchorBounds.Y {
		t.Fatalf("overlay help row = %d, want above anchor row %d", overlayPoint.Y, anchorBounds.Y)
	}

	if err := app.Driver().Move(ByID("statusbar-top-action")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText(meta.HelpText)); err == nil {
			return fmt.Errorf("overlay help %q still visible after hover exit", meta.HelpText)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestE2EStatusbarInlineHelpFallbackAndActivation(t *testing.T) {
	appFn, initFn, cleanupFn, meta := newStatusbarFixture(statusbarcomp.HelpDisplayInline)
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(72, 12), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.Driver().Move(ByID("statusbar-top-action")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("Ready"))
	}); err != nil {
		t.Fatal(err)
	}

	if err := app.Driver().Move(ByKey("statusbar-help")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText(meta.HelpText))
	}); err != nil {
		t.Fatal(err)
	}

	if err := app.Driver().Move(ByID("statusbar-top-action")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("Ready"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText(meta.HelpText)); err == nil {
		t.Fatalf("inline help %q should revert to fallback after hover exit", meta.HelpText)
	}

	app.ClearIntentLogs()
	if err := app.Driver().Click(ByKey("statusbar-activate")); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitIntent(meta.ActivateIntentType, 500*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.ActivateIntentType); err != nil {
		t.Fatal(err)
	}

	logs := app.IntentLogs()
	handled := false
	for _, log := range logs {
		if log.Type == meta.ActivateIntentType && log.Handled {
			handled = true
			break
		}
	}
	if !handled {
		t.Fatalf("handled activation intent %q not found in dispatch logs: %+v", meta.ActivateIntentType, logs)
	}
}
