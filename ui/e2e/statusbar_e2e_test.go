package e2e

import (
	"fmt"
	"strings"
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

type statusbarOverlayCornerFixtureSpec struct {
	Title         string
	HelpKey       string
	HelpText      string
	Placement     statusbarcomp.TooltipPlacement
	AnchorOnRight bool
	StickToBottom bool
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

func newStatusbarOverlayCornerFixture(spec statusbarOverlayCornerFixtureSpec) ui.ComponentFunc {
	return func() ui.VNode {
		children := make([]ui.VNode, 0, 32)
		children = append(children, ui.NewTextBuilder(spec.Title).Build())

		if spec.StickToBottom {
			for i := 0; i < 20; i++ {
				children = append(children, ui.NewTextBuilder(fmt.Sprintf("Spacer %02d", i+1)).Build())
			}
		}

		builder := statusbarcomp.NewBuilder().
			DefaultTheme().
			HelpFallback("Ready").
			HelpDisplayMode(statusbarcomp.HelpDisplayOverlay).
			TooltipPlacement(spec.Placement).
			TooltipMaxWidth(24)

		if spec.AnchorOnRight {
			builder = builder.
				Left(statusbarcomp.Text("Mode: stable").WithWidth(16)).
				Center(statusbarcomp.Text("Overlay: corner").WithWidth(24).WithAlign(ui.AlignCenter)).
				Right(statusbarcomp.Text("Help").WithKey(spec.HelpKey).WithHelp(spec.HelpText).WithWidth(16).WithAlign(ui.AlignEnd))
		} else {
			builder = builder.
				Left(statusbarcomp.Text("Help").WithKey(spec.HelpKey).WithHelp(spec.HelpText).WithWidth(16)).
				Center(statusbarcomp.Text("Overlay: corner").WithWidth(24).WithAlign(ui.AlignCenter)).
				Right(statusbarcomp.Text("Mode: stable").WithWidth(16).WithAlign(ui.AlignEnd))
		}

		children = append(children, builder.BuildWithHelp())

		if !spec.StickToBottom {
			children = append(children,
				ui.NewTextBuilder("Workspace: alpha").Build(),
				ui.NewTextBuilder("Queue: healthy").Build(),
				ui.NewTextBuilder("Runtime: idle").Build(),
			)
		}

		return ui.NewVStack().
			SetGap(0).
			SetChildrenList(children)
	}
}

func newStatusbarOperationalPresetFixture() ui.ComponentFunc {
	return func() ui.VNode {
		now := time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Statusbar Operational Preset Fixture").Build(),
				ui.StatusBarWithTheme(
					ui.StatusBarThemeDefault(),
					ui.StatusBarSections(
						ui.StatusBarEndpoint("local"),
						ui.StatusBarProfile("dev"),
						ui.StatusBarUser("admin"),
						ui.StatusBarRole("ops"),
						ui.StatusBarPage("jobs"),
					),
					ui.StatusBarSections(ui.StatusBarStateBadge("healthy")),
					nil,
				),
				ui.StatusBarWithTheme(
					ui.StatusBarThemeDefault(),
					ui.StatusBarSections(
						ui.StatusBarScope("provider"),
						ui.StatusBarTarget("openai/key-1"),
						ui.StatusBarSelection("job-1"),
						ui.StatusBarFilter("failed"),
					),
					nil,
					nil,
				),
				ui.StatusBarWithTheme(
					ui.StatusBarThemeDefault(),
					ui.StatusBarSections(
						ui.StatusBarCount("keys", 12),
						ui.StatusBarLatency(250*time.Millisecond),
						ui.StatusBarUptime(3*time.Hour),
						ui.StatusBarHotkey("r", "reload"),
						ui.StatusBarSeparator(),
						ui.StatusBarLastSync(now.Add(-2*time.Minute), now),
						ui.StatusBarAutoRefresh(30*time.Second),
					),
					nil,
					nil,
				),
			})
	}
}

func waitForRenderedText(t *testing.T, app *App, text string) {
	t.Helper()
	if err := app.AwaitIdleFor(200 * time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(1200*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if !strings.Contains(app.RenderString(), text) {
			return fmt.Errorf("render does not contain %q", text)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
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
	if overlayPoint.Y <= anchorBounds.Y {
		t.Fatalf("overlay help row = %d, want below anchor row %d for auto bottom-bias placement", overlayPoint.Y, anchorBounds.Y)
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

func TestE2EStatusbarOperationalPresetsRender(t *testing.T) {
	app, err := Run(newStatusbarOperationalPresetFixture(), ui.WithSize(120, 6))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{
		"Statusbar Operational Preset Fixture",
		"endpoint: local",
		"profile: dev",
		"user: admin",
		"role: ops",
		"page: jobs",
		"healthy",
		"scope: provider",
		"target: openai/key-1",
		"selection: job-1",
		"filter: failed",
		"keys: 12",
		"latency: 250ms",
		"uptime: 3h",
		"r reload",
		"last sync: 2m ago",
		"refresh: 30s",
	} {
		waitForRenderedText(t, app, text)
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

func TestE2EStatusbarOverlayTopRightCornerFallsBelowWithinRightFamily(t *testing.T) {
	app, err := Run(
		newStatusbarOverlayCornerFixture(statusbarOverlayCornerFixtureSpec{
			Title:         "Statusbar Top Right Corner Fixture",
			HelpKey:       "statusbar-top-right-corner-help",
			HelpText:      "TR corner help",
			Placement:     statusbarcomp.TooltipPlacementTop,
			AnchorOnRight: true,
		}),
		ui.WithSize(72, 24),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	anchorBounds, err := app.BoundsOf(ByKey("statusbar-top-right-corner-help"))
	if err != nil {
		t.Fatal(err)
	}

	if err := app.Driver().Move(ByKey("statusbar-top-right-corner-help")); err != nil {
		t.Fatal(err)
	}
	waitForRenderedText(t, app, "TR corner help")

	overlayPoint, err := app.ResolvePoint(ByText("TR corner help"))
	if err != nil {
		t.Fatal(err)
	}
	assertOverlayBelowAndShiftedLeft(t, "statusbar top-right corner help", overlayPoint.X, overlayPoint.Y, anchorBounds.X, anchorBounds.Y, anchorBounds.Width, 72)
}

func TestE2EStatusbarOverlayTopLeftCornerFallsBelowWithinLeftFamily(t *testing.T) {
	app, err := Run(
		newStatusbarOverlayCornerFixture(statusbarOverlayCornerFixtureSpec{
			Title:     "Statusbar Top Left Corner Fixture",
			HelpKey:   "statusbar-top-left-corner-help",
			HelpText:  "TL corner help",
			Placement: statusbarcomp.TooltipPlacementTop,
		}),
		ui.WithSize(72, 24),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	anchorBounds, err := app.BoundsOf(ByKey("statusbar-top-left-corner-help"))
	if err != nil {
		t.Fatal(err)
	}

	if err := app.Driver().Move(ByKey("statusbar-top-left-corner-help")); err != nil {
		t.Fatal(err)
	}
	waitForRenderedText(t, app, "TL corner help")

	overlayPoint, err := app.ResolvePoint(ByText("TL corner help"))
	if err != nil {
		t.Fatal(err)
	}
	assertOverlayBelowAndShiftedRight(t, "statusbar top-left corner help", overlayPoint.X, overlayPoint.Y, anchorBounds.X, anchorBounds.Y, 72)
}

func TestE2EStatusbarOverlayBottomRightCornerFallsAboveWithinRightFamily(t *testing.T) {
	app, err := Run(
		newStatusbarOverlayCornerFixture(statusbarOverlayCornerFixtureSpec{
			Title:         "Statusbar Bottom Right Corner Fixture",
			HelpKey:       "statusbar-bottom-right-corner-help",
			HelpText:      "BR corner help",
			Placement:     statusbarcomp.TooltipPlacementBottom,
			AnchorOnRight: true,
			StickToBottom: true,
		}),
		ui.WithSize(72, 24),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	anchorBounds, err := app.BoundsOf(ByKey("statusbar-bottom-right-corner-help"))
	if err != nil {
		t.Fatal(err)
	}

	if err := app.Driver().Move(ByKey("statusbar-bottom-right-corner-help")); err != nil {
		t.Fatal(err)
	}
	waitForRenderedText(t, app, "BR corner help")

	overlayPoint, err := app.ResolvePoint(ByText("BR corner help"))
	if err != nil {
		t.Fatal(err)
	}
	assertOverlayAboveAndShiftedLeft(t, "statusbar bottom-right corner help", overlayPoint.X, overlayPoint.Y, anchorBounds.X, anchorBounds.Y, anchorBounds.Width, 72)
}

func TestE2EStatusbarOverlayBottomLeftCornerFallsAboveWithinLeftFamily(t *testing.T) {
	app, err := Run(
		newStatusbarOverlayCornerFixture(statusbarOverlayCornerFixtureSpec{
			Title:         "Statusbar Bottom Left Corner Fixture",
			HelpKey:       "statusbar-bottom-left-corner-help",
			HelpText:      "BL corner help",
			Placement:     statusbarcomp.TooltipPlacementBottom,
			StickToBottom: true,
		}),
		ui.WithSize(72, 24),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	anchorBounds, err := app.BoundsOf(ByKey("statusbar-bottom-left-corner-help"))
	if err != nil {
		t.Fatal(err)
	}

	if err := app.Driver().Move(ByKey("statusbar-bottom-left-corner-help")); err != nil {
		t.Fatal(err)
	}
	waitForRenderedText(t, app, "BL corner help")

	overlayPoint, err := app.ResolvePoint(ByText("BL corner help"))
	if err != nil {
		t.Fatal(err)
	}
	assertOverlayAboveAndShiftedRight(t, "statusbar bottom-left corner help", overlayPoint.X, overlayPoint.Y, anchorBounds.X, anchorBounds.Y, 72)
}

func newStatusbarDualAxisClampFixture(placement statusbarcomp.TooltipPlacement, helpKey, helpText string) ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				statusbarcomp.NewBuilder().
					DefaultTheme().
					HelpFallback("Ready").
					HelpDisplayMode(statusbarcomp.HelpDisplayOverlay).
					TooltipPlacement(placement).
					TooltipMaxWidth(16).
					Left(statusbarcomp.Text("Help").WithKey(helpKey).WithHelp(helpText).WithWidth(6)).
					Center(statusbarcomp.Text("Run").WithWidth(4).WithAlign(ui.AlignCenter)).
					Right(statusbarcomp.Text("OK").WithWidth(2).WithAlign(ui.AlignEnd)).
					BuildWithHelp(),
			})
	}
}

func TestE2EStatusbarTopDualAxisClampKeepsTopArrow(t *testing.T) {
	app, err := RunWithSandbox(
		newStatusbarDualAxisClampFixture(statusbarcomp.TooltipPlacementTop, "statusbar-dualaxis-top-help", "ClampHelp"),
		ui.WithSize(14, 4),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	app.FrameworkApp().SetConfigSize(14, 4)
	app.FrameworkApp().Resize(14, 4)
	app.ForceRender()
	if err := app.AwaitIdleFor(500 * time.Millisecond); err != nil {
		t.Fatal(err)
	}

	if err := app.Driver().Move(ByKey("statusbar-dualaxis-top-help")); err != nil {
		t.Fatal(err)
	}
	waitForRenderedText(t, app, "ClampHelp")
	if err := app.AssertVisible(ByText("▼")); err != nil {
		t.Fatalf("top-family dual-axis clamp should keep bottom arrow: %v", err)
	}
	if err := app.AssertVisible(ByText("▲")); err == nil {
		t.Fatal("top-family dual-axis clamp should not render a top-border arrow")
	}
}

func TestE2EStatusbarBottomDualAxisClampKeepsBottomArrow(t *testing.T) {
	app, err := RunWithSandbox(
		newStatusbarDualAxisClampFixture(statusbarcomp.TooltipPlacementBottom, "statusbar-dualaxis-bottom-help", "ClampHelp"),
		ui.WithSize(14, 4),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	app.FrameworkApp().SetConfigSize(14, 4)
	app.FrameworkApp().Resize(14, 4)
	app.ForceRender()
	if err := app.AwaitIdleFor(500 * time.Millisecond); err != nil {
		t.Fatal(err)
	}

	if err := app.Driver().Move(ByKey("statusbar-dualaxis-bottom-help")); err != nil {
		t.Fatal(err)
	}
	waitForRenderedText(t, app, "ClampHelp")
	if err := app.AssertVisible(ByText("▲")); err != nil {
		t.Fatalf("bottom-family dual-axis clamp should keep top arrow: %v", err)
	}
	if err := app.AssertVisible(ByText("▼")); err == nil {
		t.Fatal("bottom-family dual-axis clamp should not render a bottom-border arrow")
	}
}
