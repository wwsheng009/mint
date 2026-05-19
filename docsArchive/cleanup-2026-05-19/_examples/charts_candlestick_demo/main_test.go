package main

import (
	"testing"
	"time"

	"github.com/wwsheng009/mint/ui"
)

func TestCandlestickDemoRendersCoreSignals(t *testing.T) {
	testApp, err := ui.RunTest(CandlestickDemo,
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
		"Charts Candlestick Demo",
		"Daily Tape",
		"Up",
		"Down",
		"Flat",
		"Volume",
		"M  T W T",
	} {
		if err := testApp.AssertRender(text); err != nil {
			t.Fatalf("expected %q in render: %v", text, err)
		}
	}
}
