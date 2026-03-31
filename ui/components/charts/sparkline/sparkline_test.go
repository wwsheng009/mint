package sparkline

import (
	"strings"
	"testing"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/charts/internal/palette"
)

func TestVNodeBuilder(t *testing.T) {
	vnode := NewBuilder([]float64{1, 2, 3}).
		Title("Trend").
		Width(5).
		InlineLabel("live").
		HighlightMinMax(true).
		AutoHeight().
		ASCII().
		Key("spark-1").
		Build()

	spark, ok := vnode.(*VNode)
	if !ok {
		t.Fatal("Build() should return *VNode")
	}
	if spark.Key() != "spark-1" {
		t.Fatalf("Key() = %q, want spark-1", spark.Key())
	}
	if spark.Title() != "Trend" {
		t.Fatalf("Title() = %q, want Trend", spark.Title())
	}
	if spark.Width() != 5 {
		t.Fatalf("Width() = %d, want 5", spark.Width())
	}
	if spark.Height() != 0 {
		t.Fatalf("Height() = %d, want 0", spark.Height())
	}
	if spark.InlineLabel() != "live" {
		t.Fatalf("InlineLabel() = %q, want live", spark.InlineLabel())
	}
	if !spark.HighlightMinMax() {
		t.Fatal("HighlightMinMax() = false, want true")
	}
	if !spark.AutoHeightEnabled() {
		t.Fatal("AutoHeightEnabled() = false, want true")
	}
	if spark.Mode() != RenderModeASCII {
		t.Fatalf("Mode() = %v, want RenderModeASCII", spark.Mode())
	}
}

func TestInstanceMeasureAndPaint(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propData:  []float64{1, 2, 3, 4},
		propTitle: "Trend",
		propWidth: 4,
	})

	size := inst.Measure(layout.UnboundedConstraints())
	if size.Width < 5 {
		t.Fatalf("Measure().Width = %d, want >= 5", size.Width)
	}
	if size.Height != 2 {
		t.Fatalf("Measure().Height = %d, want 2", size.Height)
	}

	cmds := inst.Paint(0, 0)
	if len(cmds) < 2 {
		t.Fatalf("Paint() commands = %d, want >= 2", len(cmds))
	}
	if cmds[0].Text != "Trend" {
		t.Fatalf("title text = %q, want Trend", cmds[0].Text)
	}
	if got := paint.StringWidth(strings.TrimRight(bufferRowText(drawCmdsToBuffer(size.Width, size.Height, cmds), 1), " ")); got != 4 {
		t.Fatalf("sparkline width = %d, want 4", got)
	}
}

func TestInstanceFixedHeightInlineLabelAndHighlight(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propData:            []float64{1, 4, 2, 6},
		propTitle:           "Trend",
		propWidth:           4,
		propHeight:          3,
		propInlineLabel:     "live",
		propHighlightMinMax: true,
	})

	size := inst.Measure(layout.UnboundedConstraints())
	if size.Height != 4 {
		t.Fatalf("Measure().Height = %d, want 4", size.Height)
	}

	buf := drawCmdsToBuffer(size.Width, size.Height, inst.Paint(0, 0))
	if got := strings.TrimRight(bufferRowText(buf, 3), " "); !strings.Contains(got, "live") {
		t.Fatalf("bottom row = %q, want inline label live", got)
	}

	foundMin := false
	foundMax := false
	for y := 1; y < size.Height; y++ {
		for x := 0; x < 4; x++ {
			cell := buf.Cells[y][x]
			if cell.Cluster == "" || cell.Cluster == " " || cell.IsContinuation {
				continue
			}
			if cell.Style.FG == palette.DownColor() {
				foundMin = true
			}
			if cell.Style.FG == palette.UpColor() {
				foundMax = true
			}
		}
	}
	if !foundMin || !foundMax {
		t.Fatalf("highlight styles missing: min=%v max=%v", foundMin, foundMax)
	}
}

func TestInstanceAdaptiveHeightFromConstraints(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propData:       []float64{1, 2, 3, 4},
		propTitle:      "Trend",
		propWidth:      4,
		propAutoHeight: true,
	})

	size := inst.Measure(layout.NewConstraints(0, layout.MaxInt, 0, 4))
	if size.Height != 4 {
		t.Fatalf("Measure().Height = %d, want 4", size.Height)
	}

	buf := drawCmdsToBuffer(size.Width, size.Height, inst.Paint(0, 0))
	plotRowsWithGlyphs := 0
	for y := 1; y < size.Height; y++ {
		row := bufferRowText(buf, y)
		if strings.ContainsAny(row, "▁▂▃▄▅▆▇█") {
			plotRowsWithGlyphs++
		}
	}
	if plotRowsWithGlyphs == 0 {
		t.Fatal("adaptive sparkline did not render multi-row plot glyphs")
	}
}

func TestResolvedRenderMode(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propData: []float64{1, 2, 3, 4, 5, 6},
	})

	inst.renderMode = RenderModeAuto
	inst.width = 8
	if got := inst.resolvedRenderMode(1); got != RenderModeBraille {
		t.Fatalf("resolvedRenderMode(auto, single row wide) = %v, want RenderModeBraille", got)
	}

	inst.width = 4
	if got := inst.resolvedRenderMode(1); got != RenderModeASCII {
		t.Fatalf("resolvedRenderMode(auto, single row narrow) = %v, want RenderModeASCII", got)
	}

	inst.width = 8
	if got := inst.resolvedRenderMode(3); got != RenderModeBlock {
		t.Fatalf("resolvedRenderMode(auto, multi row) = %v, want RenderModeBlock", got)
	}

	inst.renderMode = RenderModeBraille
	if got := inst.resolvedRenderMode(3); got != RenderModeBlock {
		t.Fatalf("resolvedRenderMode(braille, multi row) = %v, want RenderModeBlock", got)
	}

	inst.renderMode = RenderModeASCII
	if got := inst.resolvedRenderMode(3); got != RenderModeASCII {
		t.Fatalf("resolvedRenderMode(ascii, multi row) = %v, want RenderModeASCII", got)
	}
}

func TestRenderModesProduceDifferentGlyphSets(t *testing.T) {
	data := []float64{1, 2, 3, 4, 5, 6, 7, 8}

	braille := renderSparklineRow(t, NewInstance(rtui.Props{
		propData:       data,
		propWidth:      8,
		propRenderMode: RenderModeBraille,
	}))
	block := renderSparklineRow(t, NewInstance(rtui.Props{
		propData:       data,
		propWidth:      8,
		propRenderMode: RenderModeBlock,
	}))
	ascii := renderSparklineRow(t, NewInstance(rtui.Props{
		propData:       data,
		propWidth:      8,
		propRenderMode: RenderModeASCII,
	}))
	autoNarrow := renderSparklineRow(t, NewInstance(rtui.Props{
		propData:       data,
		propWidth:      4,
		propRenderMode: RenderModeAuto,
	}))

	if !strings.ContainsAny(braille, string(brailleGlyphs)) {
		t.Fatalf("braille row = %q, want braille glyphs", braille)
	}
	if !strings.ContainsAny(block, string(blockGlyphs)) {
		t.Fatalf("block row = %q, want block glyphs", block)
	}
	if !strings.ContainsAny(ascii, string(asciiGlyphs)) {
		t.Fatalf("ascii row = %q, want ascii glyphs", ascii)
	}
	if !strings.ContainsAny(autoNarrow, string(asciiGlyphs)) {
		t.Fatalf("auto narrow row = %q, want ascii fallback glyphs", autoNarrow)
	}
	if braille == block || braille == ascii || block == ascii {
		t.Fatalf("render modes should differ: braille=%q block=%q ascii=%q", braille, block, ascii)
	}
}

func TestExplicitASCIIMultiRowPreservesEnhancements(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propData:            []float64{1, 4, 2, 6},
		propTitle:           "Trend",
		propWidth:           4,
		propHeight:          3,
		propInlineLabel:     "live",
		propHighlightMinMax: true,
		propRenderMode:      RenderModeASCII,
	})

	size := inst.Measure(layout.UnboundedConstraints())
	buf := drawCmdsToBuffer(size.Width, size.Height, inst.Paint(0, 0))

	foundASCII := false
	foundMin := false
	foundMax := false
	for y := 1; y < size.Height; y++ {
		row := bufferRowText(buf, y)
		if strings.ContainsAny(row, string(asciiGlyphs)) {
			foundASCII = true
		}
		for x := 0; x < 4; x++ {
			cell := buf.Cells[y][x]
			if cell.Cluster == "" || cell.Cluster == " " || cell.IsContinuation {
				continue
			}
			if cell.Style.FG == palette.DownColor() {
				foundMin = true
			}
			if cell.Style.FG == palette.UpColor() {
				foundMax = true
			}
		}
	}
	if !foundASCII {
		t.Fatal("explicit ascii multi-row sparkline did not render ascii glyphs")
	}
	if !strings.Contains(strings.TrimRight(bufferRowText(buf, size.Height-1), " "), "live") {
		t.Fatal("inline label missing from ascii multi-row sparkline")
	}
	if !foundMin || !foundMax {
		t.Fatalf("highlight styles missing in ascii multi-row mode: min=%v max=%v", foundMin, foundMax)
	}
}

func renderSparklineRow(t *testing.T, inst *Instance) string {
	t.Helper()
	size := inst.Measure(layout.UnboundedConstraints())
	buf := drawCmdsToBuffer(size.Width, size.Height, inst.Paint(0, 0))
	rowIndex := size.Height - 1
	if inst.title != "" {
		rowIndex = 1
	}
	return strings.TrimRight(bufferRowText(buf, rowIndex), " ")
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
