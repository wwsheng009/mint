package main

import (
	"strings"
	"testing"

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

	testApp.ForceRender()
	render := testApp.GetRenderString()
	for _, want := range []string{
		"Ship Date:",
		"Ship Time:",
		"2026-04-05",
		"09:30",
	} {
		if !strings.Contains(render, want) {
			t.Fatalf("render missing %q\n%s", want, render)
		}
	}
}
