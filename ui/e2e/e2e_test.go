package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	runtimeaction "github.com/wwsheng009/mint/runtime/action"
	runtimeintent "github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
)

type firstPressIntent struct{}

func (firstPressIntent) IntentType() string { return "e2e.first_press" }

type secondPressIntent struct{}

func (secondPressIntent) IntentType() string { return "e2e.second_press" }

func e2eFixture() ui.VNode {
	return ui.NewVStack().
		SetGap(0).
		SetChildrenList([]ui.VNode{
			ui.NewButtonBuilder("First").SetID("first-btn").OnPress(firstPressIntent{}).Build(),
			ui.NewButtonBuilder("Second").SetID("second-btn").OnPress(secondPressIntent{}).Build(),
			ui.NewPaginationBuilder().SetID("pager").ComponentID("fixture.pagination").Total(120).PageSize(10).CurrentPage(2).Build(),
			ui.NewTextBuilder("Styled").Style(style.Style{}.Foreground(style.Red).Background(style.Blue).Bold(true)).Build(),
		})
}

func e2eFixtureInit() {
	rt := rtui.GetGlobalIntentRuntime()
	if rt == nil {
		return
	}
	runtimeintent.RegisterTypedRuntime(rt, func(_ *runtimeintent.ActionContext, _ firstPressIntent) runtimeintent.IntentResult {
		return runtimeintent.HandledResult()
	})
	runtimeintent.RegisterTypedRuntime(rt, func(_ *runtimeintent.ActionContext, _ secondPressIntent) runtimeintent.IntentResult {
		return runtimeintent.HandledResult()
	})
}

func TestE2EFocusAndIntentTrace(t *testing.T) {
	app, err := Run(e2eFixture, ui.WithSize(40, 10), ui.WithInit(e2eFixtureInit))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertFocus(ByID("first-btn")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	if err := app.Driver().Special(platform.KeyEnter); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitIntent("e2e.first_press", 500*time.Millisecond); err != nil {
		t.Fatal(err)
	}

	if err := app.Driver().Special(platform.KeyTab); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertFocus(ByID("second-btn")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertFocusTransition(ByID("first-btn"), ByID("second-btn")); err != nil {
		t.Fatal(err)
	}
}

func TestE2EClickByTextAndAssertStyle(t *testing.T) {
	app, err := Run(e2eFixture, ui.WithSize(40, 10), ui.WithInit(e2eFixtureInit))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	app.ClearIntentLogs()
	if err := app.Driver().Click(ByText("Second")); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitIntent("e2e.second_press", 500*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentSequence("e2e.second_press"); err != nil {
		t.Fatal(err)
	}

	if err := app.AssertStyle(ByText("Styled"), StyleExpect{
		HasFG:   true,
		FG:      style.Red,
		HasBG:   true,
		BG:      style.Blue,
		HasBold: true,
		Bold:    true,
	}); err != nil {
		t.Fatal(err)
	}

	point, err := app.ResolvePoint(ByText("Second"))
	if err != nil {
		t.Fatal(err)
	}
	if err := app.AssertHit(point, ByID("second-btn")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByComponentID("fixture.pagination")); err != nil {
		t.Fatal(err)
	}
}

func TestE2EResolveFiberByComponentIDAndBounds(t *testing.T) {
	app, err := Run(e2eFixture, ui.WithSize(40, 10), ui.WithInit(e2eFixtureInit))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	fiber, err := app.ResolveFiber(ByComponentID("fixture.pagination"))
	if err != nil {
		t.Fatal(err)
	}
	if fiber.ID != "pager" {
		t.Fatalf("resolved fiber ID = %q, want pager", fiber.ID)
	}

	bounds, err := app.BoundsOf(ByComponentID("fixture.pagination"))
	if err != nil {
		t.Fatal(err)
	}
	if bounds.Width <= 0 || bounds.Height <= 0 {
		t.Fatalf("bounds = %v, want positive size", bounds)
	}
	if err := app.AssertBounds(ByComponentID("fixture.pagination"), bounds); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByTag("pagination")); err != nil {
		t.Fatal(err)
	}
}

func TestE2EAssertFocusByComponentIDAndBoundsPoint(t *testing.T) {
	app, err := Run(e2eFixture, ui.WithSize(40, 10), ui.WithInit(e2eFixtureInit))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.Driver().Special(platform.KeyTab); err != nil {
		t.Fatal(err)
	}
	if err := app.Driver().Special(platform.KeyTab); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertFocus(ByComponentID("fixture.pagination")); err != nil {
		t.Fatal(err)
	}

	bounds, err := app.BoundsOf(ByComponentID("fixture.pagination"))
	if err != nil {
		t.Fatal(err)
	}
	center := layout.Rect{
		X:      bounds.X + bounds.Width/2,
		Y:      bounds.Y + bounds.Height/2,
		Width:  1,
		Height: 1,
	}
	if err := app.AssertBounds(At(center.X, center.Y), center); err != nil {
		t.Fatal(err)
	}
}

func TestE2ERawInputTraceAndIntentSequence(t *testing.T) {
	app, err := Run(e2eFixture, ui.WithSize(40, 10), ui.WithInit(e2eFixtureInit))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	app.ClearRawInputs()
	app.ClearIntentLogs()
	app.ClearFocusTransitions()

	if err := app.Driver().Special(platform.KeyEnter); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitMessage("key", 500*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitAction(runtimeaction.ActionEnter, 500*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitTrace(TraceMatch{Kind: TraceMsg, Name: "key"}, 500*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := app.Driver().Special(platform.KeyTab); err != nil {
		t.Fatal(err)
	}
	if err := app.Driver().Special(platform.KeyEnter); err != nil {
		t.Fatal(err)
	}

	raws := app.RawInputs()
	if len(raws) != 3 {
		t.Fatalf("raw input trace len = %d, want 3", len(raws))
	}
	if raws[0].Name() != "key:Enter" || raws[1].Name() != "key:Tab" || raws[2].Name() != "key:Enter" {
		t.Fatalf("raw trace names = [%s %s %s]", raws[0].Name(), raws[1].Name(), raws[2].Name())
	}
	if err := app.AssertMessageSequence("key", "key", "key"); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertActionSequence(runtimeaction.ActionEnter, runtimeaction.ActionNavigateNext, runtimeaction.ActionEnter); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertLastMessage("key"); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertLastAction(runtimeaction.ActionEnter); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertActionHandled(runtimeaction.ActionEnter, "keyboard_target"); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertActionHandled(runtimeaction.ActionNavigateNext, "navigation"); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentSequence("e2e.first_press", "e2e.second_press"); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertLastIntent("e2e.second_press"); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled("e2e.first_press"); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled("e2e.second_press"); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "key:Enter"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceIntentDispatch, Name: "e2e.second_press"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceSequence(
		TraceMatch{Kind: TraceRawInput, Name: "key:Enter"},
		TraceMatch{Kind: TraceMsg, Name: "key"},
		TraceMatch{Kind: TraceAction, Name: string(runtimeaction.ActionEnter)},
		TraceMatch{Kind: TraceIntentDispatch, Name: "e2e.first_press"},
	); err != nil {
		t.Fatal(err)
	}

	trace := app.TraceEvents()
	if len(trace) < 5 {
		t.Fatalf("trace len = %d, want >= 5", len(trace))
	}
}

func TestE2EAwaitFocusAndEventually(t *testing.T) {
	app, err := Run(e2eFixture, ui.WithSize(40, 10), ui.WithInit(e2eFixtureInit))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.Driver().Special(platform.KeyTab); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitFocus(ByID("second-btn"), 500*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertFocus(ByID("second-btn"))
	}); err != nil {
		t.Fatal(err)
	}
}

func TestE2EDiagnosticsSnapshotAndSave(t *testing.T) {
	app, err := Run(e2eFixture, ui.WithSize(40, 10), ui.WithInit(e2eFixtureInit))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.Driver().Special(platform.KeyEnter); err != nil {
		t.Fatal(err)
	}
	if err := app.Driver().Special(platform.KeyTab); err != nil {
		t.Fatal(err)
	}
	if err := app.Driver().Special(platform.KeyEnter); err != nil {
		t.Fatal(err)
	}

	snapshot := app.DiagnosticsSnapshot()
	if !strings.Contains(snapshot.Render, "Styled") {
		t.Fatalf("render snapshot missing expected text: %q", snapshot.Render)
	}
	if snapshot.Focus == nil || snapshot.Focus.ID != "second-btn" {
		t.Fatalf("focus snapshot = %+v, want second-btn", snapshot.Focus)
	}
	if len(snapshot.RawInputs) == 0 || len(snapshot.Messages) == 0 || len(snapshot.Actions) == 0 || len(snapshot.Intents) == 0 {
		t.Fatalf("diagnostics snapshot missing traces: %+v", snapshot)
	}

	dir := t.TempDir()
	if err := app.SaveDiagnostics(dir); err != nil {
		t.Fatal(err)
	}

	renderBytes, err := os.ReadFile(filepath.Join(dir, "render.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(renderBytes), "Styled") {
		t.Fatalf("render.txt missing expected text: %q", string(renderBytes))
	}

	diagBytes, err := os.ReadFile(filepath.Join(dir, "diagnostics.json"))
	if err != nil {
		t.Fatal(err)
	}
	var decoded DiagnosticsSnapshot
	if err := json.Unmarshal(diagBytes, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Focus == nil || decoded.Focus.ID != "second-btn" {
		t.Fatalf("decoded focus snapshot = %+v, want second-btn", decoded.Focus)
	}

	traceBytes, err := os.ReadFile(filepath.Join(dir, "trace.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(traceBytes), "\"kind\"") {
		t.Fatalf("trace.json missing expected content: %q", string(traceBytes))
	}

	tempDir, err := app.SaveDiagnosticsTemp("mint-e2e-test-")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(tempDir, "diagnostics.json")); err != nil {
		t.Fatalf("expected diagnostics.json in temp dir: %v", err)
	}
}

func TestE2ESaveDiagnosticsOnFailure(t *testing.T) {
	app, err := Run(e2eFixture, ui.WithSize(40, 10), ui.WithInit(e2eFixtureInit))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if dir, err := app.SaveDiagnosticsOnFailure(t, "mint-e2e-pass-"); err != nil {
		t.Fatal(err)
	} else if dir != "" {
		t.Fatalf("expected no diagnostics for passing test, got %q", dir)
	}
}

func TestE2EByTargetIDAndByTag(t *testing.T) {
	app, err := Run(e2eFixture, ui.WithSize(40, 10), ui.WithInit(e2eFixtureInit))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	secondFiber, err := app.ResolveFiber(ByID("second-btn"))
	if err != nil {
		t.Fatal(err)
	}
	point, err := app.ResolvePoint(ByID("second-btn"))
	if err != nil {
		t.Fatal(err)
	}
	if err := app.AssertHit(point, ByTargetID(secondFiber.ActionTargetID)); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTargetID(ByID("second-btn"), secondFiber.ActionTargetID); err != nil {
		t.Fatal(err)
	}

	pagerFiber, err := app.ResolveFiber(ByTag("pagination"))
	if err != nil {
		t.Fatal(err)
	}
	if pagerFiber.Tag != "pagination" {
		t.Fatalf("resolved fiber tag = %q, want pagination", pagerFiber.Tag)
	}
}
