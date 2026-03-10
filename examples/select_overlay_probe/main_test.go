package main

import (
	"testing"
	"time"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/runtime/event"
	"github.com/wwsheng009/mint/runtime/platform"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
	selectcomp "github.com/wwsheng009/mint/ui/components/select"
)

func TestSelectOverlayProbe_KeyboardCommitViaPump(t *testing.T) {
	resetProbeState()

	testApp, err := ui.RunTest(SelectOverlayProbe,
		ui.WithWidth(50),
		ui.WithHeight(14),
		ui.WithPluginSetup(func(app *framework.App) {
			selectcomp.Install(app)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	waitForText(t, testApp, "Selected: United States")

	if err := testApp.InjectSpecialKey(platform.KeyEnter); err != nil {
		t.Fatal(err)
	}
	waitForHitTag(t, testApp, "select-popup")

	if err := testApp.InjectSpecialKey(platform.KeyDown); err != nil {
		t.Fatal(err)
	}
	waitForPopupHighlight(t, testApp, 1)

	if err := testApp.InjectSpecialKey(platform.KeyEnter); err != nil {
		t.Fatal(err)
	}
	waitForText(t, testApp, "Selected: China")
}

func TestSelectOverlayProbe_MouseClickCommitViaPump(t *testing.T) {
	resetProbeState()

	testApp, err := ui.RunTest(SelectOverlayProbe,
		ui.WithWidth(50),
		ui.WithHeight(14),
		ui.WithPluginSetup(func(app *framework.App) {
			selectcomp.Install(app)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	waitForText(t, testApp, "Selected: United States")

	trigger := waitForHitTag(t, testApp, "select")
	if err := testApp.InjectMouse(trigger.Bounds.X+1, trigger.Bounds.Y, platform.MouseLeft, platform.MousePress); err != nil {
		t.Fatal(err)
	}
	if err := testApp.InjectMouse(trigger.Bounds.X+1, trigger.Bounds.Y, platform.MouseLeft, platform.MouseRelease); err != nil {
		t.Fatal(err)
	}

	popup := waitForHitTag(t, testApp, "select-popup")
	if err := testApp.InjectMouse(popup.Bounds.X+2, popup.Bounds.Y+2, platform.MouseLeft, platform.MousePress); err != nil {
		t.Fatal(err)
	}
	if err := testApp.InjectMouse(popup.Bounds.X+2, popup.Bounds.Y+2, platform.MouseLeft, platform.MouseRelease); err != nil {
		t.Fatal(err)
	}

	waitForText(t, testApp, "Selected: China")
	waitForPopupClosed(t, testApp)
}

func waitForText(t *testing.T, testApp *ui.TestableApp, want string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		waitForPumpCycle(testApp)
		if err := testApp.AssertRender(want); err == nil {
			return
		}
	}
	t.Fatalf("render never contained %q\nrendered:\n%s", want, testApp.GetRenderString())
}

func waitForHitTag(t *testing.T, testApp *ui.TestableApp, tag string) event.HitMapEntry {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		waitForPumpCycle(testApp)
		if entry, ok := findHitEntryByTag(testApp.GetFrameworkApp().GetHitMap(), tag); ok {
			return entry
		}
	}
	t.Fatalf("hitmap never contained tag %q", tag)
	return event.HitMapEntry{}
}

func waitForPopupHighlight(t *testing.T, testApp *ui.TestableApp, want int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		waitForPumpCycle(testApp)
		if highlighted, ok := currentPopupHighlightedIndex(testApp); ok && highlighted == want {
			return
		}
	}
	t.Fatalf("popup highlightedIndex never reached %d", want)
}

func waitForPopupClosed(t *testing.T, testApp *ui.TestableApp) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		waitForPumpCycle(testApp)
		if _, ok := findHitEntryByTag(testApp.GetFrameworkApp().GetHitMap(), "select-popup"); !ok {
			return
		}
	}
	t.Fatal("select-popup remained in hitmap after commit")
}

func waitForPumpCycle(testApp *ui.TestableApp) {
	time.Sleep(40 * time.Millisecond)
	testApp.ForceRender()
	time.Sleep(20 * time.Millisecond)
}

func findHitEntryByTag(hitMap *event.HitMap, tag string) (event.HitMapEntry, bool) {
	if hitMap == nil {
		return event.HitMapEntry{}, false
	}
	for _, entry := range hitMap.AllEntries() {
		fiber, ok := entry.TargetFiber.(*rtui.Fiber)
		if !ok || fiber == nil {
			continue
		}
		if fiber.Tag == tag {
			return entry, true
		}
	}
	return event.HitMapEntry{}, false
}

func currentPopupHighlightedIndex(testApp *ui.TestableApp) (int, bool) {
	hitMap := testApp.GetFrameworkApp().GetHitMap()
	entry, ok := findHitEntryByTag(hitMap, "select-popup")
	if !ok {
		return 0, false
	}
	fiber, ok := entry.TargetFiber.(*rtui.Fiber)
	if !ok || fiber == nil || fiber.Instance == nil {
		return 0, false
	}
	props := fiber.Instance.GetProps()
	value, ok := props["highlightedIndex"].(int)
	if !ok {
		return 0, false
	}
	return value, true
}
