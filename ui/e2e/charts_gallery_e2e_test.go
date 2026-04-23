package e2e

import (
	"testing"

	gallerydemo "github.com/wwsheng009/mint/examples/charts_gallery_demo/gallery"
	fwtheme "github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
)

func newChartsGalleryApp() ui.ComponentFunc {
	return gallerydemo.Build
}

func TestE2EChartsGalleryRender(t *testing.T) {
	app, err := Run(newChartsGalleryApp(), ui.WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{
		"Mint Charts Gallery",
		"KPI Pulse",
		"Traffic + Tape",
		"SLO Bullet Charts",
		"Throughput + Hotspots",
		"Tape",
		"Scatter",
		"Hotspots",
		"range 1..6 avg 3.5",
		"18.4k",
		"99.94%",
		"● API",
		"● Worker",
		"Latency: 173/250 target 200",
		"Availability: 996/1000 target 999",
		"█ Ingress",
		"█ Egress",
		"NA",
		"EU",
		"DB",
		"M T W",
	} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestE2EChartsGalleryPaletteStyles(t *testing.T) {
	app, err := Run(newChartsGalleryApp(), ui.WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, tc := range []struct {
		text string
		fg   style.Color
	}{
		{text: "● API", fg: fwtheme.Primary()},
		{text: "● Worker", fg: fwtheme.Accent()},
		{text: "█ Ingress", fg: fwtheme.Primary()},
		{text: "█ Egress", fg: fwtheme.Accent()},
	} {
		if err := app.AssertStyle(ByText(tc.text), StyleExpect{
			HasFG: true,
			FG:    tc.fg,
		}); err != nil {
			t.Fatal(err)
		}
	}
}

func TestE2EChartsGalleryRenderSnapshot(t *testing.T) {
	app, err := Run(newChartsGalleryApp(), ui.WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	defer func() {
		_, _ = app.SaveDiagnosticsOnFailure(t, "mint-e2e-charts-gallery-")
	}()

	assertRenderSnapshot(t, app, "charts_gallery_80x24.render.txt")
}
