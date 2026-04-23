package main

import (
	"testing"
	"time"

	"github.com/wwsheng009/mint/ui"
)

func TestChartsGalleryDemoRendersCoreSections(t *testing.T) {
	testApp, err := ui.RunTest(ChartsGalleryDemo,
		ui.WithWidth(80),
		ui.WithHeight(24),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	time.Sleep(100 * time.Millisecond)
	testApp.ForceRender()
	time.Sleep(100 * time.Millisecond)

	for _, text := range []string{
		"Mint Charts Gallery",
		"KPI Pulse",
		"SLO Bullet Charts",
		"Traffic + Tape",
		"Throughput + Hotspots",
		"Tape",
		"Scatter",
		"Hotspots",
		"range 1..6 avg 3.5",
	} {
		if err := testApp.AssertRender(text); err != nil {
			t.Fatalf("expected %q in render: %v", text, err)
		}
	}
}

func TestChartsGalleryDemoRendersChartSignals(t *testing.T) {
	testApp, err := ui.RunTest(ChartsGalleryDemo,
		ui.WithWidth(80),
		ui.WithHeight(24),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	time.Sleep(100 * time.Millisecond)
	testApp.ForceRender()
	time.Sleep(100 * time.Millisecond)

	for _, text := range []string{
		"● API",
		"● Worker",
		"Latency",
		"Availability",
		"Ingress",
		"Egress",
		"NA",
		"EU",
		"DB",
		"M T W",
	} {
		if err := testApp.AssertRender(text); err != nil {
			t.Fatalf("expected %q in render: %v", text, err)
		}
	}
}
