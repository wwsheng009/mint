package scatterplot

import (
	"strings"
	"testing"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/charts/internal/palette"
)

func TestVNodeBuilder(t *testing.T) {
	vnode := NewBuilder([]Point{{X: 1, Y: 2}, {X: 3, Y: 5}}).
		Title("Scatter").
		Series(
			Series{Name: "API", Points: []Point{{X: 1, Y: 2}, {X: 3, Y: 5}}},
			Series{Name: "Worker", Points: []Point{{X: 2, Y: 1}, {X: 4, Y: 4}}, Glyph: '◆'},
		).
		Width(9).
		Height(4).
		Domain(NewDomain(0, 10, 0, 12)).
		Viewport(NewViewport(2, 8, 1, 9)).
		XReferenceLines(2.5, 7.5).
		YReferenceLineLabeled(6, "Floor").
		XReferenceBands(NewLabeledReferenceBand(2, 4, "Focus")).
		YReferenceBands(NewLabeledReferenceBand(5, 8, "Risk")).
		ShowAxis(false).
		ShowGrid(true).
		ShowLegend(true).
		Key("scatter-1").
		Build()

	scatter, ok := vnode.(*VNode)
	if !ok {
		t.Fatal("Build() should return *VNode")
	}
	if scatter.Key() != "scatter-1" {
		t.Fatalf("Key() = %q, want scatter-1", scatter.Key())
	}
	if scatter.Title() != "Scatter" {
		t.Fatalf("Title() = %q, want Scatter", scatter.Title())
	}
	if scatter.Width() != 9 {
		t.Fatalf("Width() = %d, want 9", scatter.Width())
	}
	if scatter.Height() != 4 {
		t.Fatalf("Height() = %d, want 4", scatter.Height())
	}
	if got := scatter.Domain(); got != NewDomain(0, 10, 0, 12) {
		t.Fatalf("Domain() = %+v, want %+v", got, NewDomain(0, 10, 0, 12))
	}
	if got := scatter.Viewport(); got != NewViewport(2, 8, 1, 9) {
		t.Fatalf("Viewport() = %+v, want %+v", got, NewViewport(2, 8, 1, 9))
	}
	if got := scatter.XReferenceLines(); !float64SlicesEqual(got, []float64{2.5, 7.5}) {
		t.Fatalf("XReferenceLines() = %+v, want [2.5 7.5]", got)
	}
	if got := scatter.YReferenceLines(); !float64SlicesEqual(got, []float64{6}) {
		t.Fatalf("YReferenceLines() = %+v, want [6]", got)
	}
	if got := scatter.YReferenceLineDefs(); !referenceLineSlicesEqual(got, []ReferenceLine{{Value: 6, Label: "Floor"}}) {
		t.Fatalf("YReferenceLineDefs() = %+v, want [{6 Floor}]", got)
	}
	if got := scatter.XReferenceBands(); !referenceBandSlicesEqual(got, []ReferenceBand{{Min: 2, Max: 4, Label: "Focus"}}) {
		t.Fatalf("XReferenceBands() = %+v, want [{2 4 Focus}]", got)
	}
	if got := scatter.YReferenceBands(); !referenceBandSlicesEqual(got, []ReferenceBand{{Min: 5, Max: 8, Label: "Risk"}}) {
		t.Fatalf("YReferenceBands() = %+v, want [{5 8 Risk}]", got)
	}
	if len(scatter.Series()) != 2 {
		t.Fatalf("Series() len = %d, want 2", len(scatter.Series()))
	}
	if scatter.ShowAxis() {
		t.Fatal("ShowAxis() = true, want false")
	}
	if !scatter.ShowGrid() {
		t.Fatal("ShowGrid() = false, want true")
	}
	if !scatter.ShowLegend() {
		t.Fatal("ShowLegend() = false, want true")
	}
}

func TestInstanceMeasureAndPaint(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle:    "Scatter",
		propPoints:   []Point{{X: 1, Y: 2}, {X: 3, Y: 5}, {X: 5, Y: 4}},
		propWidth:    5,
		propHeight:   4,
		propShowAxis: true,
	})

	size := inst.Measure(layout.UnboundedConstraints())
	if size.Width < 5 {
		t.Fatalf("Measure().Width = %d, want >= 5", size.Width)
	}
	if size.Height != 7 {
		t.Fatalf("Measure().Height = %d, want 7", size.Height)
	}

	buf := drawCmdsToBuffer(size.Width, size.Height, inst.Paint(0, 0))
	if got := strings.TrimRight(bufferRowText(buf, 0), " "); got != "Scatter" {
		t.Fatalf("title row = %q, want Scatter", got)
	}
	if got := strings.TrimRight(bufferRowText(buf, 5), " "); got != "─────" {
		t.Fatalf("axis row = %q, want %q", got, "─────")
	}
	if got := strings.TrimRight(bufferRowText(buf, 6), " "); got != "x:1..5 y:2..5" {
		t.Fatalf("summary row = %q, want %q", got, "x:1..5 y:2..5")
	}

	hasPoint := false
	for y := 1; y <= 4; y++ {
		if strings.ContainsRune(bufferRowText(buf, y), defaultPointRune) {
			hasPoint = true
			break
		}
	}
	if !hasPoint {
		t.Fatal("plot rows do not contain point glyph")
	}
}

func TestInstanceLegendAndGrid(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle:      "Scatter",
		propSeriesName: "API",
		propPoints:     []Point{{X: 1, Y: 2}, {X: 3, Y: 5}, {X: 5, Y: 4}},
		propWidth:      7,
		propHeight:     4,
		propShowGrid:   true,
		propShowLegend: true,
		propShowAxis:   true,
	})

	size := inst.Measure(layout.UnboundedConstraints())
	buf := drawCmdsToBuffer(size.Width, size.Height, inst.Paint(0, 0))
	if got := strings.TrimRight(bufferRowText(buf, 1), " "); got != "● API" {
		t.Fatalf("legend row = %q, want %q", got, "● API")
	}

	hasGrid := false
	for y := 2; y < 2+inst.plotHeight(); y++ {
		row := bufferRowText(buf, y)
		if strings.ContainsRune(row, '┈') || strings.ContainsRune(row, '┆') {
			hasGrid = true
			break
		}
	}
	if !hasGrid {
		t.Fatal("plot rows do not contain grid glyph")
	}
}

func TestInstanceMultiSeriesPaint(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propSeries: []Series{
			{Name: "API", Points: []Point{{X: 1, Y: 2}, {X: 3, Y: 5}}},
			{Name: "Worker", Points: []Point{{X: 2, Y: 1}, {X: 4, Y: 4}}, Glyph: '◆'},
		},
		propWidth:      7,
		propHeight:     4,
		propShowLegend: true,
		propShowAxis:   true,
	})

	size := inst.Measure(layout.UnboundedConstraints())
	buf := drawCmdsToBuffer(size.Width, size.Height, inst.Paint(0, 0))

	if got := strings.TrimRight(bufferRowText(buf, 0), " "); got != "● API" {
		t.Fatalf("legend row 1 = %q, want %q", got, "● API")
	}
	if got := strings.TrimRight(bufferRowText(buf, 1), " "); got != "◆ Worker" {
		t.Fatalf("legend row 2 = %q, want %q", got, "◆ Worker")
	}

	hasSeries0 := false
	hasSeries1 := false
	hasGlyph1 := false
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
			if cell.Cluster == "◆" {
				hasGlyph1 = true
			}
		}
	}
	if !hasSeries0 || !hasSeries1 {
		t.Fatalf("multi-series colors missing: series0=%v series1=%v", hasSeries0, hasSeries1)
	}
	if !hasGlyph1 {
		t.Fatal("plot rows do not contain custom glyph ◆")
	}
}

func TestInstanceCustomDomainSummaryAndPlacement(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle:    "Scatter",
		propPoints:   []Point{{X: 2, Y: 3}},
		propWidth:    5,
		propHeight:   4,
		propDomain:   NewDomain(0, 10, 0, 12),
		propShowAxis: true,
	})

	size := inst.Measure(layout.UnboundedConstraints())
	buf := drawCmdsToBuffer(size.Width, size.Height, inst.Paint(0, 0))
	if got := strings.TrimRight(bufferRowText(buf, 6), " "); got != "x:0..10 y:0..12" {
		t.Fatalf("summary row = %q, want %q", got, "x:0..10 y:0..12")
	}
	if got := buf.Cells[3][1].Cluster; got != string(defaultPointRune) {
		t.Fatalf("point cluster at custom-domain position = %q, want %q", got, string(defaultPointRune))
	}
	if got := buf.Cells[3][2].Cluster; got == string(defaultPointRune) {
		t.Fatalf("point should not stay at auto-domain center position")
	}
}

func TestInstanceReferenceLinesRender(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle:    "Scatter",
		propPoints:   []Point{{X: 2, Y: 2}},
		propWidth:    5,
		propHeight:   4,
		propDomain:   NewDomain(0, 10, 0, 10),
		propXRefs:    []float64{5},
		propYRefs:    []float64{5},
		propShowAxis: false,
	})

	size := inst.Measure(layout.UnboundedConstraints())
	buf := drawCmdsToBuffer(size.Width, size.Height, inst.Paint(0, 0))

	hasVertical := false
	hasHorizontal := false
	for y := 1; y <= 4; y++ {
		row := bufferRowText(buf, y)
		if strings.ContainsRune(row, '│') || strings.ContainsRune(row, '┼') {
			hasVertical = true
		}
		if strings.ContainsRune(row, '─') || strings.ContainsRune(row, '┼') {
			hasHorizontal = true
		}
	}
	if !hasVertical || !hasHorizontal {
		t.Fatalf("reference lines missing: vertical=%v horizontal=%v", hasVertical, hasHorizontal)
	}
}

func TestInstanceReferenceLineLegend(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle:      "Scatter",
		propSeriesName: "API",
		propPoints:     []Point{{X: 5, Y: 5}},
		propWidth:      7,
		propHeight:     5,
		propXRefDefs:   []ReferenceLine{{Value: 5, Label: "Target"}},
		propYRefDefs:   []ReferenceLine{{Value: 6, Label: "Floor"}},
		propShowLegend: true,
		propShowAxis:   false,
	})

	size := inst.Measure(layout.UnboundedConstraints())
	buf := drawCmdsToBuffer(size.Width, size.Height, inst.Paint(0, 0))
	if got := strings.TrimRight(bufferRowText(buf, 1), " "); got != "● API" {
		t.Fatalf("legend row 1 = %q, want %q", got, "● API")
	}
	if got := strings.TrimRight(bufferRowText(buf, 2), " "); got != "│ x: Target" {
		t.Fatalf("legend row 2 = %q, want %q", got, "│ x: Target")
	}
	if got := strings.TrimRight(bufferRowText(buf, 3), " "); got != "─ y: Floor" {
		t.Fatalf("legend row 3 = %q, want %q", got, "─ y: Floor")
	}
}

func TestInstanceReferenceBandsRender(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle:  "Scatter",
		propPoints: []Point{{X: 5, Y: 5}},
		propWidth:  7,
		propHeight: 5,
		propDomain: NewDomain(0, 10, 0, 10),
		propXBands: []ReferenceBand{{Min: 2, Max: 4}},
		propYBands: []ReferenceBand{{Min: 6, Max: 8}},
	})

	size := inst.Measure(layout.UnboundedConstraints())
	buf := drawCmdsToBuffer(size.Width, size.Height, inst.Paint(0, 0))

	hasLight := false
	hasOverlap := false
	for y := 1; y <= 5; y++ {
		row := bufferRowText(buf, y)
		if strings.ContainsRune(row, bandRuneLight) {
			hasLight = true
		}
		if strings.ContainsRune(row, bandRuneMedium) || strings.ContainsRune(row, bandRuneHeavy) {
			hasOverlap = true
		}
	}
	if !hasLight {
		t.Fatal("plot rows do not contain light reference band glyph")
	}
	if !hasOverlap {
		t.Fatal("plot rows do not contain overlapped reference band glyph")
	}
}

func TestInstanceReferenceBandLegend(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle:      "Scatter",
		propSeriesName: "API",
		propPoints:     []Point{{X: 5, Y: 5}},
		propWidth:      7,
		propHeight:     5,
		propXBands:     []ReferenceBand{{Min: 2, Max: 4, Label: "Focus"}},
		propYBands:     []ReferenceBand{{Min: 6, Max: 8, Label: "Risk"}},
		propShowLegend: true,
		propShowAxis:   false,
	})

	size := inst.Measure(layout.UnboundedConstraints())
	buf := drawCmdsToBuffer(size.Width, size.Height, inst.Paint(0, 0))
	if got := strings.TrimRight(bufferRowText(buf, 1), " "); got != "● API" {
		t.Fatalf("legend row 1 = %q, want %q", got, "● API")
	}
	if got := strings.TrimRight(bufferRowText(buf, 2), " "); got != "░ x: Focus" {
		t.Fatalf("legend row 2 = %q, want %q", got, "░ x: Focus")
	}
	if got := strings.TrimRight(bufferRowText(buf, 3), " "); got != "░ y: Risk" {
		t.Fatalf("legend row 3 = %q, want %q", got, "░ y: Risk")
	}
}

func TestInstanceViewportSummaryAndClipping(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle: "Scatter",
		propSeries: []Series{
			{
				Name:   "API",
				Points: []Point{{X: 1, Y: 1}, {X: 4, Y: 6}, {X: 8, Y: 9}},
			},
		},
		propWidth:    5,
		propHeight:   4,
		propDomain:   NewDomain(0, 10, 0, 10),
		propViewport: NewViewport(3, 7, 2, 8),
		propShowAxis: true,
	})

	size := inst.Measure(layout.UnboundedConstraints())
	buf := drawCmdsToBuffer(size.Width, size.Height, inst.Paint(0, 0))
	if got := strings.TrimRight(bufferRowText(buf, 6), " "); got != "x:3..7 y:2..8" {
		t.Fatalf("summary row = %q, want %q", got, "x:3..7 y:2..8")
	}

	pointCount := 0
	for y := 1; y <= 4; y++ {
		for x := 0; x < buf.Width; x++ {
			if buf.Cells[y][x].Cluster == string(defaultPointRune) {
				pointCount++
			}
		}
	}
	if pointCount != 1 {
		t.Fatalf("visible point count = %d, want 1", pointCount)
	}
	if got := buf.Cells[2][1].Cluster; got != string(defaultPointRune) {
		t.Fatalf("viewport point cluster = %q, want %q", got, string(defaultPointRune))
	}
}

func TestInstanceCollisionMarkerRender(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle: "Scatter",
		propSeries: []Series{
			{
				Name:   "API",
				Points: []Point{{X: 2, Y: 2}, {X: 2, Y: 2}},
			},
			{
				Name:   "Worker",
				Points: []Point{{X: 2, Y: 2}},
				Glyph:  '◆',
			},
		},
		propWidth:    5,
		propHeight:   4,
		propDomain:   NewDomain(0, 10, 0, 10),
		propShowAxis: false,
	})

	size := inst.Measure(layout.UnboundedConstraints())
	buf := drawCmdsToBuffer(size.Width, size.Height, inst.Paint(0, 0))

	foundTwo := false
	foundThree := false
	for y := 1; y <= 4; y++ {
		for x := 0; x < buf.Width; x++ {
			cell := buf.Cells[y][x]
			if cell.Cluster != "2" && cell.Cluster != "3" {
				continue
			}
			if cell.Cluster == "2" {
				foundTwo = true
			}
			if cell.Cluster == "3" {
				foundThree = true
			}
			if cell.Style.FG != palette.CollisionColor() {
				t.Fatalf("collision style fg = %q, want %q", cell.Style.FG, palette.CollisionColor())
			}
		}
	}
	if !foundThree {
		t.Fatal("plot rows do not contain 3-count collision marker")
	}
	if foundTwo {
		t.Fatal("single hotspot fixture should not contain 2-count collision marker")
	}
}

func TestCollisionRuneForCount(t *testing.T) {
	tests := []struct {
		count int
		want  rune
	}{
		{count: 1, want: defaultPointRune},
		{count: 2, want: '2'},
		{count: 3, want: '3'},
		{count: 9, want: '9'},
		{count: 10, want: maxCollisionRune},
	}

	for _, tt := range tests {
		if got := collisionRuneForCount(tt.count); got != tt.want {
			t.Fatalf("collisionRuneForCount(%d) = %q, want %q", tt.count, string(got), string(tt.want))
		}
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
