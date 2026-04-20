package demo

import (
	"fmt"

	runtimeplatform "github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/ui"
	linechartcomp "github.com/wwsheng009/mint/ui/components/charts/linechart"
)

var prototypeData = []float64{2, 5, 3, 7, 6, 9, 8, 11, 10, 12}

var prototypeLabels = []string{"03/20", "03/21", "03/22", "03/23", "03/24", "03/25", "03/26", "03/27", "03/28", "03/29"}

const (
	prototypeChartWidth  = 31
	prototypeChartHeight = 6
)

// Build returns the image-prototype view shared by the example and e2e tests.
func Build() ui.VNode {
	status := PrototypeStatus()

	return ui.NewVStack().
		SetGap(0).
		SetChildrenList([]ui.VNode{
			ui.NewTextBuilder("Charts LineChart Image Prototype").Build(),
			ui.NewTextBuilder("Chart pixel backend is currently paused; terminal graphics remain reserved for dedicated image controls").Build(),
			ui.NewTextBuilder(status.Banner).Build(),
			ui.HStackBuilder(
				ui.Flex(prototypeChart("linechart-image-prototype-text", "Text Backend", linechartcomp.RenderBackendText), 1),
				ui.Flex(prototypeChart("linechart-image-prototype-image", status.ImageTitle, linechartcomp.RenderBackendImagePlot), 1),
			).Gap(4).Stretch().Build(),
			buildDiagnosticsPanel(),
		})
}

// PrototypeDiagnostics returns stable diagnostics lines for the prototype UI and tests.
func PrototypeDiagnostics() []string {
	status := PrototypeStatus()
	return []string{
		"Graphics: " + status.Capabilities.Summary(),
		"Display: " + status.Display,
		"Backends: text vs requested-image-plot(paused)",
		"Scene: " + prototypeSceneSummary(),
	}
}

type PrototypeRuntimeStatus struct {
	Capabilities runtimeplatform.GraphicsCapabilities
	Banner       string
	Display      string
	ImageTitle   string
}

// PrototypeStatus reports whether the current terminal can actually display the
// generated image layer or will fall back to plain text output.
func PrototypeStatus() PrototypeRuntimeStatus {
	caps := runtimeplatform.ProbeGraphicsCapabilities()
	return PrototypeRuntimeStatus{
		Capabilities: caps,
		Banner:       "Chart image backend paused: charts currently render as text only. Terminal graphics support, when available, is reserved for dedicated image controls.",
		Display:      "charts render text only; chart pixel backend paused",
		ImageTitle:   "Requested Image Plot Backend (paused to text)",
	}
}

func buildDiagnosticsPanel() ui.VNode {
	children := []ui.VNode{
		ui.NewTextBuilder("Diagnostics").Build(),
	}
	for _, line := range PrototypeDiagnostics() {
		children = append(children, ui.NewTextBuilder(line).Build())
	}

	return ui.NewVStack().
		SetGap(0).
		SetChildrenList(children)
}

func prototypeChart(id, title string, backend linechartcomp.RenderBackend) ui.VNode {
	builder := linechartcomp.NewBuilder(prototypeData).
		SetID(id).
		Title(title).
		Labels(prototypeLabels).
		AutoAxisLabels().
		Width(prototypeChartWidth).
		Height(prototypeChartHeight).
		ShowLegend(false).
		ShowGrid(true).
		ShowAxis(true).
		ShowPoints(false)

	switch backend {
	case linechartcomp.RenderBackendImagePlot:
		builder.ImagePlotBackend()
	default:
		builder.TextBackend()
	}

	return builder.Build()
}

func prototypeSceneSummary() string {
	inst := prototypeChartInstance(linechartcomp.RenderBackendImagePlot)
	inst.SetBounds(0, 0, prototypeChartWidth, prototypeChartHeight+2)

	layers := inst.SceneLayers()
	actualBackend, _ := inst.GetProps()["renderBackend"].(linechartcomp.RenderBackend)
	if len(layers) == 0 {
		if !linechartcomp.SupportsImagePlotBackend() {
			return fmt.Sprintf("images=0 backend=%s requested=image-plot-disabled", actualBackend.String())
		}
		return fmt.Sprintf("images=0 backend=%s", actualBackend.String())
	}

	layer := layers[0]
	return fmt.Sprintf(
		"images=%d backend=%s cells=%dx%d pixels=%dx%d",
		len(layers),
		actualBackend.String(),
		layer.Bounds.Width,
		layer.Bounds.Height,
		layer.PixelWidth,
		layer.PixelHeight,
	)
}

func prototypeChartInstance(backend linechartcomp.RenderBackend) *linechartcomp.Instance {
	return prototypeChartVNode(backend).CreateInstance().(*linechartcomp.Instance)
}

func prototypeChartVNode(backend linechartcomp.RenderBackend) *linechartcomp.VNode {
	builder := linechartcomp.NewBuilder(prototypeData).
		Title("Prototype").
		Labels(prototypeLabels).
		AutoAxisLabels().
		Width(prototypeChartWidth).
		Height(prototypeChartHeight).
		ShowLegend(false).
		ShowGrid(true).
		ShowAxis(true).
		ShowPoints(false)

	switch backend {
	case linechartcomp.RenderBackendImagePlot:
		builder.ImagePlotBackend()
	default:
		builder.TextBackend()
	}

	return builder.BuildTyped()
}
