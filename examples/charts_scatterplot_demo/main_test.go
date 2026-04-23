package main

import (
	"testing"
	"time"

	"github.com/wwsheng009/mint/ui"
)

func TestScatterPlotDemoRendersCoreSignals(t *testing.T) {
	testApp, err := ui.RunTest(ScatterPlotDemo,
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
		"Charts ScatterPlot Demo",
		"Service Correlation",
		"● API",
		"◆ Worker",
		"│ x: Target",
		"░ y: Risk",
		"x:0..10 y:0..12",
	} {
		if err := testApp.AssertRender(text); err != nil {
			t.Fatalf("expected %q in render: %v", text, err)
		}
	}
}
