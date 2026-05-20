package demo

import (
	"github.com/wwsheng009/mint/ui"
	linechartcomp "github.com/wwsheng009/mint/ui/components/charts/linechart"
)

// Build returns the linechart demo view shared by the example and e2e tests.
func Build() ui.VNode {
	return ui.NewVStack().
		SetGap(0).
		SetChildrenList([]ui.VNode{
			ui.NewTextBuilder("Charts LineChart Demo").Build(),
			readabilityChart(),
			ui.HStackBuilder(
				ui.Flex(axisModeChart("linechart-demo-auto", "Line Axis Auto", linechartcomp.AxisLabelModeAuto), 1),
				ui.Flex(axisModeChart("linechart-demo-dense", "Line Axis Dense", linechartcomp.AxisLabelModeDense), 1),
				ui.Flex(axisModeChart("linechart-demo-sparse", "Line Axis Sparse", linechartcomp.AxisLabelModeSparse), 1),
			).Gap(2).Stretch().Build(),
		})
}

func readabilityChart() ui.VNode {
	return linechartcomp.NewBuilder([]float64{2, 4, 3, 6, 5, 8, 7, 9, 8, 10}).
		SetID("linechart-demo-readability").
		Title("Readability Baseline").
		Labels([]string{"03/20", "03/21", "03/22", "03/23", "03/24", "03/25", "03/26", "03/27", "03/28", "03/29"}).
		AutoAxisLabels().
		Width(31).
		Height(6).
		ShowLegend(false).
		ShowGrid(true).
		ShowAxis(true).
		ShowPoints(false).
		Build()
}

func axisModeChart(id, title string, mode linechartcomp.AxisLabelMode) ui.VNode {
	labels := []string{"03/24", "03/25", "03/26", "03/27", "03/28", "03/29"}
	data := []float64{1, 9, 2, 8, 3, 7}

	return linechartcomp.NewBuilder(data).
		SetID(id).
		Title(title).
		Labels(labels).
		AxisLabelMode(mode).
		Width(11).
		Height(4).
		ShowLegend(false).
		ShowGrid(true).
		ShowAxis(true).
		ShowPoints(false).
		Build()
}
