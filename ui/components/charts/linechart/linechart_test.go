package linechart

import (
	"strings"
	"testing"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/charts/internal/palette"
)

func TestVNodeBuilder(t *testing.T) {
	vnode := NewBuilder([]float64{1, 3, 2, 5}).
		Title("Trend").
		Labels([]string{"03/24", "03/25", "03/26", "03/27"}).
		SparseAxisLabels().
		Series(
			Series{Name: "CPU", Data: []float64{1, 3, 2, 5}},
			Series{Name: "MEM", Data: []float64{2, 4, 3, 5}},
		).
		Width(7).
		Height(4).
		ShowAxis(false).
		ShowGrid(true).
		ShowLegend(true).
		ShowPoints(false).
		Key("line-1").
		Build()

	line, ok := vnode.(*VNode)
	if !ok {
		t.Fatal("Build() should return *VNode")
	}
	if line.Key() != "line-1" {
		t.Fatalf("Key() = %q, want line-1", line.Key())
	}
	if line.Title() != "Trend" {
		t.Fatalf("Title() = %q, want Trend", line.Title())
	}
	if line.Width() != 7 {
		t.Fatalf("Width() = %d, want 7", line.Width())
	}
	if got := line.Labels(); len(got) != 4 || got[0] != "03/24" {
		t.Fatalf("Labels() = %+v, want first dense label preserved", got)
	}
	if line.AxisLabelMode() != AxisLabelModeSparse {
		t.Fatalf("AxisLabelMode() = %v, want %v", line.AxisLabelMode(), AxisLabelModeSparse)
	}
	if line.Height() != 4 {
		t.Fatalf("Height() = %d, want 4", line.Height())
	}
	if len(line.Series()) != 2 {
		t.Fatalf("Series() len = %d, want 2", len(line.Series()))
	}
	if line.ShowAxis() {
		t.Fatal("ShowAxis() = true, want false")
	}
	if !line.ShowGrid() {
		t.Fatal("ShowGrid() = false, want true")
	}
	if !line.ShowLegend() {
		t.Fatal("ShowLegend() = false, want true")
	}
	if line.ShowPoints() {
		t.Fatal("ShowPoints() = true, want false")
	}
}

func TestInstanceMeasureAndPaint(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle:      "Trend",
		propData:       []float64{1, 3, 2, 5, 4},
		propWidth:      5,
		propHeight:     4,
		propShowAxis:   true,
		propShowPoints: true,
	})

	size := inst.Measure(layout.UnboundedConstraints())
	if size.Width < 5 {
		t.Fatalf("Measure().Width = %d, want >= 5", size.Width)
	}
	if size.Height != 6 {
		t.Fatalf("Measure().Height = %d, want 6", size.Height)
	}

	buf := drawCmdsToBuffer(size.Width, size.Height, inst.Paint(0, 0))
	if got := strings.TrimRight(bufferRowText(buf, 0), " "); got != "Trend" {
		t.Fatalf("title row = %q, want Trend", got)
	}
	if got := strings.TrimRight(bufferRowText(buf, 5), " "); got != "─────" {
		t.Fatalf("axis row = %q, want %q", got, "─────")
	}

	hasPoint := false
	for y := 1; y <= 4; y++ {
		if strings.ContainsRune(bufferRowText(buf, y), pointGlyph) {
			hasPoint = true
			break
		}
	}
	if !hasPoint {
		t.Fatal("plot rows do not contain point glyph")
	}
}

func TestInstanceWidthSampling(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propData:   []float64{1, 2, 3, 4, 5},
		propWidth:  3,
		propHeight: 3,
	})

	if inst.sampleCount() != 2 {
		t.Fatalf("sampleCount() = %d, want 2", inst.sampleCount())
	}
	if inst.plotWidth() != 3 {
		t.Fatalf("plotWidth() = %d, want 3", inst.plotWidth())
	}
}

func TestResampleForContinuityPreservesTurningPoints(t *testing.T) {
	data := []float64{1, 9, 2, 8, 3, 7, 4, 6, 5}
	sampled := resampleForContinuity(data, 5)
	if len(sampled) != 5 {
		t.Fatalf("len(sampled) = %d, want 5", len(sampled))
	}
	if sampled[0] != 1 || sampled[len(sampled)-1] != 5 {
		t.Fatalf("sample endpoints = %+v, want first=1 last=5", sampled)
	}
	hasHighPeak := false
	for _, value := range sampled {
		if value >= 8 {
			hasHighPeak = true
			break
		}
	}
	if !hasHighPeak {
		t.Fatalf("sampled = %+v, want to preserve a high turning point", sampled)
	}
}

func TestInstanceNarrowWidthContinuityRender(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle:      "Trend",
		propData:       []float64{1, 9, 2, 8, 3, 7, 4, 6, 5},
		propWidth:      5,
		propHeight:     4,
		propShowAxis:   false,
		propShowGrid:   false,
		propShowLegend: false,
		propShowPoints: true,
	})

	size := inst.Measure(layout.UnboundedConstraints())
	buf := drawCmdsToBuffer(size.Width, size.Height, inst.Paint(0, 0))

	topRow := bufferRowText(buf, 1)
	hasTopPeak := strings.ContainsRune(topRow, pointGlyph)
	if !hasTopPeak {
		t.Fatalf("top plot row = %q, want preserved peak point", topRow)
	}
}
func TestInstanceLegendAndGrid(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle:      "Trend",
		propSeriesName: "CPU",
		propData:       []float64{1, 3, 2, 5, 4},
		propWidth:      5,
		propHeight:     4,
		propShowAxis:   true,
		propShowGrid:   true,
		propShowLegend: true,
		propShowPoints: true,
	})

	size := inst.Measure(layout.UnboundedConstraints())
	buf := drawCmdsToBuffer(size.Width, size.Height, inst.Paint(0, 0))
	if got := strings.TrimRight(bufferRowText(buf, 1), " "); got != "● CPU" {
		t.Fatalf("legend row = %q, want %q", got, "● CPU")
	}

	hasGrid := false
	for y := 2; y <= 5; y++ {
		if strings.ContainsRune(bufferRowText(buf, y), '┈') {
			hasGrid = true
			break
		}
	}
	if !hasGrid {
		t.Fatal("plot rows do not contain grid glyph")
	}
}

func TestInstanceDenseAxisLabelsFoldAndStaySparse(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle:         "Trend",
		propData:          []float64{1, 9, 2, 8, 3, 7},
		propLabels:        []string{"03/24", "03/25", "03/26", "03/27", "03/28", "03/29"},
		propAxisLabelMode: AxisLabelModeDense,
		propWidth:         11,
		propHeight:        4,
		propShowAxis:      true,
		propShowGrid:      true,
		propShowLegend:    false,
		propShowPoints:    true,
	})

	size := inst.Measure(layout.UnboundedConstraints())
	buf := drawCmdsToBuffer(size.Width, size.Height, inst.Paint(0, 0))

	axisRow := strings.TrimRight(bufferRowText(buf, size.Height-2), " ")
	if axisRow != "───────────" {
		t.Fatalf("axis row = %q, want %q", axisRow, "───────────")
	}

	labelRow := strings.TrimRight(bufferRowText(buf, size.Height-1), " ")
	if labelRow != "4 5 6 7 8 9" {
		t.Fatalf("label row = %q, want %q", labelRow, "4 5 6 7 8 9")
	}

	bottomPlotRow := strings.TrimRight(bufferRowText(buf, 4), " ")
	if strings.ContainsRune(bottomPlotRow, '┈') {
		t.Fatalf("bottom plot row = %q, want grid to avoid the row nearest dense labels", bottomPlotRow)
	}
}

func TestInstanceSparseAxisLabelsReduceLabelDensity(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle:         "Trend",
		propData:          []float64{1, 9, 2, 8, 3, 7},
		propLabels:        []string{"03/24", "03/25", "03/26", "03/27", "03/28", "03/29"},
		propAxisLabelMode: AxisLabelModeSparse,
		propWidth:         11,
		propHeight:        4,
		propShowAxis:      true,
		propShowGrid:      true,
		propShowLegend:    false,
		propShowPoints:    true,
	})

	size := inst.Measure(layout.UnboundedConstraints())
	buf := drawCmdsToBuffer(size.Width, size.Height, inst.Paint(0, 0))
	labelRow := strings.TrimRight(bufferRowText(buf, size.Height-1), " ")
	if labelRow != "4   6     9" {
		t.Fatalf("sparse label row = %q, want %q", labelRow, "4   6     9")
	}
}

func TestFoldAxisLabelPrefersTrailingDigit(t *testing.T) {
	if got := foldAxisLabel("03/29"); got != "9" {
		t.Fatalf("foldAxisLabel(date) = %q, want 9", got)
	}
	if got := foldAxisLabel("Mon"); got != "M" {
		t.Fatalf("foldAxisLabel(word) = %q, want M", got)
	}
}

func TestInstanceMultiSeriesPaint(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propSeries: []Series{
			{Name: "CPU", Data: []float64{1, 3, 2, 5, 4}},
			{Name: "MEM", Data: []float64{2, 2, 4, 3, 5}},
		},
		propWidth:      7,
		propHeight:     4,
		propShowAxis:   true,
		propShowLegend: true,
		propShowPoints: true,
	})

	size := inst.Measure(layout.UnboundedConstraints())
	buf := drawCmdsToBuffer(size.Width, size.Height, inst.Paint(0, 0))

	if got := strings.TrimRight(bufferRowText(buf, 0), " "); got != "● CPU" {
		t.Fatalf("legend row 1 = %q, want %q", got, "● CPU")
	}
	if got := strings.TrimRight(bufferRowText(buf, 1), " "); got != "● MEM" {
		t.Fatalf("legend row 2 = %q, want %q", got, "● MEM")
	}

	hasSeries0 := false
	hasSeries1 := false
	for y := 2; y < 2+inst.plotHeight(); y++ {
		for x := 0; x < buf.Width; x++ {
			cell := buf.Cells[y][x]
			if cell.IsContinuation || cell.Cluster == "" || cell.Cluster == " " {
				continue
			}
			if cell.Style.FG == palette.SeriesColor(0) {
				hasSeries0 = true
			}
			if cell.Style.FG == palette.SeriesColor(1) {
				hasSeries1 = true
			}
		}
	}
	if !hasSeries0 || !hasSeries1 {
		t.Fatalf("multi-series colors missing: series0=%v series1=%v", hasSeries0, hasSeries1)
	}
}

func drawCmdsToBuffer(width, height int, cmds []paint.DrawCmd) *paint.Buffer {
	buf := paint.NewBuffer(width, height)
	for _, cmd := range cmds {
		buf.SetString(cmd.X, cmd.Y, cmd.Text, cmd.Style)
	}
	return buf
}

func bufferRowText(buf *paint.Buffer, y int) string {
	if y < 0 || y >= buf.Height {
		return ""
	}
	var builder strings.Builder
	for x := 0; x < buf.Width; x++ {
		cell := buf.Cells[y][x]
		if cell.IsContinuation {
			continue
		}
		cluster := cell.Cluster
		if cluster == "" {
			cluster = " "
		}
		builder.WriteString(cluster)
	}
	return builder.String()
}
