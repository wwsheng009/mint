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

func TestMVPComponentsDemo_SelectMouseClickCommitsOption(t *testing.T) {
	initStore()
	appReducer.RegisterToGlobal(appStore)
	runtimeApp = nil

	testApp, err := ui.RunTest(App,
		ui.WithWidth(70),
		ui.WithHeight(35),
		ui.WithInteractionMode(ui.InteractionModeInteractive),
		ui.WithPluginSetup(func(app *framework.App) {
			runtimeApp = app
			selectcomp.Install(app)
			applyRuntimeInteractionMode(appStore.Get().InteractionMode)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	trigger := waitForMVPHit(t, testApp, func(fiber *rtui.Fiber) bool {
		return fiber.Tag == "select" && fiber.ID == "profile.country"
	})
	if err := testApp.InjectMouse(trigger.Bounds.X+1, trigger.Bounds.Y, platform.MouseLeft, platform.MousePress); err != nil {
		t.Fatal(err)
	}
	if err := testApp.InjectMouse(trigger.Bounds.X+1, trigger.Bounds.Y, platform.MouseLeft, platform.MouseRelease); err != nil {
		t.Fatal(err)
	}

	popup := waitForMVPHit(t, testApp, func(fiber *rtui.Fiber) bool {
		return fiber.Tag == "select-popup"
	})
	if err := testApp.InjectMouse(popup.Bounds.X+2, popup.Bounds.Y+2, platform.MouseLeft, platform.MousePress); err != nil {
		t.Fatal(err)
	}
	if err := testApp.InjectMouse(popup.Bounds.X+2, popup.Bounds.Y+2, platform.MouseLeft, platform.MouseRelease); err != nil {
		t.Fatal(err)
	}

	waitForMVPStoreCountry(t, "1")
}

func waitForMVPStoreCountry(t *testing.T, want string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if state := appStore.Get(); state.Country == want {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("store country = %q, want %q", appStore.Get().Country, want)
}

func waitForMVPHit(t *testing.T, testApp *ui.TestableApp, match func(*rtui.Fiber) bool) event.HitMapEntry {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		waitForMVPCycle(testApp)
		if entry, ok := findMVPHit(testApp.GetFrameworkApp().GetHitMap(), match); ok {
			return entry
		}
	}
	t.Fatal("expected matching hitmap entry")
	return event.HitMapEntry{}
}

func waitForMVPCycle(testApp *ui.TestableApp) {
	time.Sleep(40 * time.Millisecond)
	testApp.ForceRender()
	time.Sleep(20 * time.Millisecond)
}

func findMVPHit(hitMap *event.HitMap, match func(*rtui.Fiber) bool) (event.HitMapEntry, bool) {
	if hitMap == nil {
		return event.HitMapEntry{}, false
	}
	for _, entry := range hitMap.AllEntries() {
		fiber, ok := entry.TargetFiber.(*rtui.Fiber)
		if !ok || fiber == nil {
			continue
		}
		if match(fiber) {
			return entry, true
		}
	}
	return event.HitMapEntry{}, false
}
