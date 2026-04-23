package e2e

import (
	"testing"

	linechartprototype "github.com/wwsheng009/mint/examples/charts_linechart_image_prototype/demo"
	"github.com/wwsheng009/mint/ui"
)

func newChartsLineChartImagePrototypeApp() ui.ComponentFunc {
	return linechartprototype.Build
}

func TestE2EChartsLineChartImagePrototypeRender(t *testing.T) {
	app, err := Run(newChartsLineChartImagePrototypeApp(), ui.WithSize(104, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{
		"Charts LineChart Image Prototype",
		"Text Backend",
		"Chart image backend paused",
		"Requested Image Plot Backend (paused to text)",
		"Diagnostics",
		"Graphics:",
		"Display:",
		"Backends: text vs requested-image-plot(paused)",
		"Scene: images=0 backend=text requested=image-plot-disabled",
	} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatal(err)
		}
	}
}
