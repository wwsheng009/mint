package main

import (
	"strings"
	"testing"
	"time"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/ui/components/modal"
)

func TestStickyModalAcknowledgeAndBackdropBehavior(t *testing.T) {
	appStore.Set(AppState{})

	testApp, err := ui.RunTest(
		App,
		ui.WithSize(70, 40),
		ui.WithPluginSetup(func(app *framework.App) {
			app.AddMiddleware(modal.NewModalMiddleware())
		}),
		ui.WithInit(func() {
			appReducer.RegisterToGlobal(appStore)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	waitForRender(testApp)

	stickyX, stickyY := findTextPosition(t, testApp, "Sticky")
	clickAt(t, testApp, stickyX+1, stickyY)
	waitForRender(testApp)

	rendered := testApp.GetRenderString()
	if !strings.Contains(rendered, "Sticky Modal") {
		t.Fatalf("sticky modal should be visible after click, got:\n%s", rendered)
	}

	outsideX, outsideY := findOutsidePoint(t, testApp)
	clickAt(t, testApp, outsideX, outsideY)
	waitForRender(testApp)

	rendered = testApp.GetRenderString()
	if !strings.Contains(rendered, "Sticky Modal") {
		t.Fatalf("sticky modal should remain open after backdrop click, got:\n%s", rendered)
	}

	ackX, ackY := findTextPosition(t, testApp, "Acknowledge")
	clickAt(t, testApp, ackX+1, ackY)
	waitForRender(testApp)

	rendered = testApp.GetRenderString()
	if strings.Contains(rendered, "Sticky Modal") {
		t.Fatalf("sticky modal should close after acknowledge, got:\n%s", rendered)
	}

	basicX, basicY := findTextPosition(t, testApp, "Basic")
	clickAt(t, testApp, basicX+1, basicY)
	waitForRender(testApp)

	rendered = testApp.GetRenderString()
	if !strings.Contains(rendered, "Basic Modal") {
		t.Fatalf("basic modal should open after sticky closes, got:\n%s", rendered)
	}
}

func clickAt(t *testing.T, testApp *ui.TestableApp, x, y int) {
	t.Helper()
	if err := testApp.InjectMouse(x, y, platform.MouseLeft, platform.MousePress); err != nil {
		t.Fatalf("mouse press failed at (%d,%d): %v", x, y, err)
	}
	if err := testApp.InjectMouse(x, y, platform.MouseLeft, platform.MouseRelease); err != nil {
		t.Fatalf("mouse release failed at (%d,%d): %v", x, y, err)
	}
}

func waitForRender(testApp *ui.TestableApp) {
	time.Sleep(80 * time.Millisecond)
	testApp.ForceRender()
	time.Sleep(20 * time.Millisecond)
}

func findTextPosition(t *testing.T, testApp *ui.TestableApp, needle string) (int, int) {
	t.Helper()
	lines := strings.Split(testApp.GetRenderString(), "\n")
	for y, line := range lines {
		if idx := strings.Index(line, needle); idx >= 0 {
			return idx, y
		}
	}
	t.Fatalf("text %q not found in render:\n%s", needle, testApp.GetRenderString())
	return 0, 0
}

func findOutsidePoint(t *testing.T, testApp *ui.TestableApp) (int, int) {
	t.Helper()
	lines := strings.Split(testApp.GetRenderString(), "\n")
	for y, line := range lines {
		if strings.Contains(line, "Sticky Modal") {
			return 0, y
		}
	}
	t.Fatalf("sticky modal header not found in render:\n%s", testApp.GetRenderString())
	return 0, 0
}
