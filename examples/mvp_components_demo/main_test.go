package main

import (
	"strings"
	"testing"
	"time"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/ui"
	selectcomp "github.com/wwsheng009/mint/ui/components/select"
)

func TestMVPComponentsDemo_RendersDateAndTimePickers(t *testing.T) {
	initStore()
	appReducer.RegisterToGlobal(appStore)
	runtimeApp = nil

	testApp, err := ui.RunTest(App,
		ui.WithWidth(80),
		ui.WithHeight(40),
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

	wants := []string{
		"Ship Date:",
		"Ship Time:",
		"2026-04-05",
		"09:30",
	}

	deadline := time.Now().Add(2 * time.Second)
	var render string
	for time.Now().Before(deadline) {
		testApp.ForceRender()
		render = testApp.GetRenderString()

		missing := ""
		for _, want := range wants {
			if !strings.Contains(render, want) {
				missing = want
				break
			}
		}
		if missing == "" {
			return
		}

		time.Sleep(20 * time.Millisecond)
	}

	for _, want := range wants {
		if !strings.Contains(render, want) {
			t.Fatalf("render missing %q\n%s", want, render)
		}
	}
}
