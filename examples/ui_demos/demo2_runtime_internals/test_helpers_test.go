package main

import (
	"strings"
	"testing"
	"time"

	ui "github.com/wwsheng009/mint/ui"
)

func waitForDemo2Render(t *testing.T, testApp *ui.TestableApp, timeout time.Duration, condition func(string) bool) string {
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

func waitForDemo2Idle(t *testing.T, testApp *ui.TestableApp) string {
	t.Helper()
	return waitForDemo2Render(t, testApp, 300*time.Millisecond, func(rendered string) bool {
		return strings.TrimSpace(rendered) != ""
	})
}

func settleDemo2Render(t *testing.T, testApp *ui.TestableApp) string {
	t.Helper()

	time.Sleep(20 * time.Millisecond)
	testApp.ForceRender()
	return testApp.GetRenderString()
}

func logDemo2Snapshot(t *testing.T, title, rendered string, maxLines int) {
	t.Helper()
	t.Logf("%s\n%s\n=== End ===", title, truncateLines(rendered, maxLines))
}
