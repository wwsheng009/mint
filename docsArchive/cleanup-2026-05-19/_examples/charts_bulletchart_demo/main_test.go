package main

import (
	"testing"
	"time"

	"github.com/wwsheng009/mint/ui"
)

func TestBulletChartDemoRendersDirectionSignals(t *testing.T) {
	testApp, err := ui.RunTest(BulletChartDemo,
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
		"Charts BulletChart Demo",
		"Throughput: 82/100 target 75",
		"Latency Ceiling: 173/250 target 200",
		"Error Rate: 0/100 target 5",
	} {
		if err := testApp.AssertRender(text); err != nil {
			t.Fatalf("expected %q in render: %v", text, err)
		}
	}
}
