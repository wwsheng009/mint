package e2e

import (
	"testing"

	linechartdemo "github.com/wwsheng009/mint/examples/charts_linechart_demo/demo"
	"github.com/wwsheng009/mint/ui"
)

func newChartsLineChartDemoApp() ui.ComponentFunc {
	return linechartdemo.Build
}

func TestE2EChartsLineChartDemoRender(t *testing.T) {
	app, err := Run(newChartsLineChartDemoApp(), ui.WithSize(72, 22))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{
		"Charts LineChart Demo",
		"Readability Baseline",
		"Line Axis Auto",
		"Line Axis Dense",
		"Line Axis Sparse",
		"4 5 6 7 8 9",
		"4   6     9",
	} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestE2EChartsLineChartDemoSnapshot(t *testing.T) {
	app, err := Run(newChartsLineChartDemoApp(), ui.WithSize(72, 22))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	defer func() {
		_, _ = app.SaveDiagnosticsOnFailure(t, "mint-e2e-charts-linechart-demo-")
	}()

	assertRenderSnapshot(t, app, "charts_linechart_demo_72x22.render.txt")
}
