package main

import (
	"strings"
	"testing"
	"time"

	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/ui"
)

const (
	focusedOpenModalButton = ">[ [Open Modal] ]"
	focusedAddCountButton  = ">[ Add Count ]"
	focusedQuitButton      = ">[ Quit ]"
)

func newDemoTestApp(t *testing.T) *ui.TestableApp {
	t.Helper()

	prevState := appStore.Get()
	appStore.Set(AppState{Count: 0, ShowModal: false, Input: ""})
	t.Cleanup(func() { appStore.Set(prevState) })

	testApp, err := ui.RunTest(App,
		ui.WithWidth(80),
		ui.WithHeight(24),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = testApp.Close() })

	waitForDemoIdle(t, testApp)
	return testApp
}

func waitForDemoRender(t *testing.T, testApp *ui.TestableApp, timeout time.Duration, condition func(string) bool) string {
	t.Helper()

	deadline := time.Now().Add(timeout)
	var rendered string
	for {
		testApp.ForceRender()
		rendered = testApp.GetRenderString()
		if condition == nil || condition(rendered) {
			return rendered
		}
		if !time.Now().Before(deadline) {
			return rendered
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func waitForDemoIdle(t *testing.T, testApp *ui.TestableApp) string {
	t.Helper()

	rendered := waitForDemoRender(t, testApp, 2*time.Second, func(rendered string) bool {
		return strings.Contains(rendered, "TUI Engine Demo") && testApp.GetFocusedIndex() >= 0
	})
	if !strings.Contains(rendered, "TUI Engine Demo") {
		t.Fatalf("demo app did not render initial content; last render:\n%s", rendered)
	}
	return rendered
}

func settleDemoRender(t *testing.T, testApp *ui.TestableApp) string {
	t.Helper()

	time.Sleep(30 * time.Millisecond)
	testApp.ForceRender()
	return testApp.GetRenderString()
}

func injectDemoSpecialKey(t *testing.T, testApp *ui.TestableApp, key platform.SpecialKey) string {
	t.Helper()
	if err := testApp.InjectSpecialKey(key); err != nil {
		t.Fatalf("failed to inject %v: %v", key, err)
	}
	return settleDemoRender(t, testApp)
}

func focusButton(t *testing.T, testApp *ui.TestableApp, marker string) {
	t.Helper()

	for attempt := 0; attempt < 8; attempt++ {
		testApp.ForceRender()
		rendered := testApp.GetRenderString()
		if strings.Contains(rendered, marker) {
			return
		}

		if err := testApp.InjectSpecialKey(platform.KeyTab); err != nil {
			t.Fatalf("failed to inject Tab while focusing %q: %v", marker, err)
		}
		settleDemoRender(t, testApp)
	}

	testApp.ForceRender()
	t.Fatalf("failed to focus %q; last render:\n%s", marker, testApp.GetRenderString())
}

func pressEnter(t *testing.T, testApp *ui.TestableApp) {
	t.Helper()

	before := testApp.GetRenderString()
	if err := testApp.InjectSpecialKey(platform.KeyEnter); err != nil {
		t.Fatalf("failed to inject Enter: %v", err)
	}
	waitForDemoRender(t, testApp, 120*time.Millisecond, func(rendered string) bool {
		return rendered != before
	})
}

func openModalFromFocusedButton(t *testing.T, testApp *ui.TestableApp) {
	t.Helper()
	focusButton(t, testApp, focusedOpenModalButton)
	pressEnter(t, testApp)
}

func openModalViaStore(t *testing.T, testApp *ui.TestableApp) {
	t.Helper()

	current := appStore.Get()
	appStore.Set(AppState{
		Count:     current.Count,
		ShowModal: true,
		Input:     current.Input,
	})

	waitForDemoRender(t, testApp, 150*time.Millisecond, func(rendered string) bool {
		return strings.Contains(rendered, "*** Are you sure? ***")
	})
}
