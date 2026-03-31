package main

import (
	"testing"
	"time"

	"github.com/wwsheng009/mint/ui"
)

func TestLineChartDemoRendersAxisModeSignals(t *testing.T) {
	testApp, err := ui.RunTest(LineChartDemo,
		ui.WithWidth(72),
		ui.WithHeight(22),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	time.Sleep(100 * time.Millisecond)
	testApp.ForceRender()
	time.Sleep(100 * time.Millisecond)

	for _, text := range []string{
		"Charts LineChart Demo",
		"Readability Baseline",
		"Line Axis Auto",
		"Line Axis Dense",
		"Line Axis Sparse",
		"4 5 6 7 8 9",
		"4   6     9",
	} {
		if err := testApp.AssertRender(text); err != nil {
			t.Fatalf("expected %q in render: %v", text, err)
		}
	}
}
