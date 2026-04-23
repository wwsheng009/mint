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

	for attempt := 0; attempt < 10; attempt++ {
		time.Sleep(30 * time.Millisecond)
		testApp.ForceRender()
		if testApp.GetFocusedIndex() >= 0 {
			return testApp
		}
	}

	testApp.ForceRender()
	return testApp
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
		time.Sleep(30 * time.Millisecond)
	}

	testApp.ForceRender()
	t.Fatalf("failed to focus %q; last render:\n%s", marker, testApp.GetRenderString())
}

func pressEnter(t *testing.T, testApp *ui.TestableApp) {
	t.Helper()

	if err := testApp.InjectSpecialKey(platform.KeyEnter); err != nil {
		t.Fatalf("failed to inject Enter: %v", err)
	}
	time.Sleep(200 * time.Millisecond)
	testApp.ForceRender()
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

	time.Sleep(150 * time.Millisecond)
	testApp.ForceRender()
}
