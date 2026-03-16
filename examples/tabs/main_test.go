package main

import (
	"testing"
	"time"

	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/ui"
)

func resetTabsStore() {
	tabsStore.Set(AppState{ActiveTab: TabHome})
}

func TestTabsDemoArrowNavigationUpdatesContent(t *testing.T) {
	resetTabsStore()

	testApp, err := ui.RunTest(MainComponent,
		ui.WithWidth(56),
		ui.WithHeight(20),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	time.Sleep(100 * time.Millisecond)

	if err := testApp.AssertRender("Welcome to the Home tab!"); err != nil {
		t.Fatalf("expected initial Home content: %v", err)
	}

	if err := testApp.InjectSpecialKey(platform.KeyRight); err != nil {
		t.Fatalf("failed to inject right arrow: %v", err)
	}
	time.Sleep(100 * time.Millisecond)
	testApp.ForceRender()
	time.Sleep(100 * time.Millisecond)

	if err := testApp.AssertRender("User Profile"); err != nil {
		t.Fatalf("expected Profile content after right arrow: %v", err)
	}
}

func TestTabsDemoCtrlTabUpdatesContent(t *testing.T) {
	resetTabsStore()

	testApp, err := ui.RunTest(MainComponent,
		ui.WithWidth(56),
		ui.WithHeight(20),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	time.Sleep(100 * time.Millisecond)

	if err := testApp.InjectSpecialKeyWithMod(platform.KeyTab, platform.ModCtrl); err != nil {
		t.Fatalf("failed to inject Ctrl+Tab: %v", err)
	}
	time.Sleep(100 * time.Millisecond)
	testApp.ForceRender()
	time.Sleep(100 * time.Millisecond)

	if err := testApp.AssertRender("User Profile"); err != nil {
		t.Fatalf("expected Profile content after Ctrl+Tab: %v", err)
	}
}
