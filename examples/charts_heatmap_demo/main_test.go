package main

import (
	"testing"
	"time"

	"github.com/wwsheng009/mint/ui"
)

func TestHeatmapDemoRendersCoreSignals(t *testing.T) {
	testApp, err := ui.RunTest(HeatmapDemo,
		ui.WithWidth(72),
		ui.WithHeight(20),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	time.Sleep(100 * time.Millisecond)
	testApp.ForceRender()
	time.Sleep(100 * time.Millisecond)

	for _, text := range []string{
		"Charts Heatmap Demo",
		"Regional Hotspots",
		"L ░▒▓█ H",
		"T W T",
		"Sou~",
		"Eur~",
		"Asi~",
	} {
		if err := testApp.AssertRender(text); err != nil {
			t.Fatalf("expected %q in render: %v", text, err)
		}
	}
}
