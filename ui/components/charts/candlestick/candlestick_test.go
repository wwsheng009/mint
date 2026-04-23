package candlestick

import (
	"strings"
	"testing"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/charts/internal/palette"
)

func TestVNodeBuilder(t *testing.T) {
	upStyle := style.NewStyle().Foreground(style.Yellow).Bold(true)
	downStyle := style.NewStyle().Foreground(style.Magenta).Underline(true)
	flatStyle := style.NewStyle().Foreground(style.Cyan).Italic(true)
	wickStyle := style.NewStyle().Foreground(style.BrightBlack).Bold(true)
	volumeStyle := style.NewStyle().Foreground(style.BrightWhite).Reverse(true)

	vnode := NewBuilder([]Candle{
		{Label: "M", Open: 100, High: 110, Low: 98, Close: 108, Volume: 1200},
		{Label: "T", Open: 108, High: 112, Low: 101, Close: 103, Volume: 900},
	}).
		Title("Daily").
		Width(7).
		Height(5).
		ShowAxis(false).
		ShowGrid(true).
		ShowLegend(true).
		ShowVolume(true).
		VolumeHeight(2).
		UpStyle(upStyle).
		DownStyle(downStyle).
		FlatStyle(flatStyle).
		WickStyle(wickStyle).
		VolumeStyle(volumeStyle).
		Key("candle-1").
		Build()

	candle, ok := vnode.(*VNode)
	if !ok {
		t.Fatal("Build() should return *VNode")
	}
	if candle.Key() != "candle-1" {
		t.Fatalf("Key() = %q, want candle-1", candle.Key())
	}
	if candle.Title() != "Daily" {
		t.Fatalf("Title() = %q, want Daily", candle.Title())
	}
	if candle.Width() != 7 {
		t.Fatalf("Width() = %d, want 7", candle.Width())
	}
	if candle.Height() != 5 {
		t.Fatalf("Height() = %d, want 5", candle.Height())
	}
	if len(candle.Candles()) != 2 {
		t.Fatalf("Candles() len = %d, want 2", len(candle.Candles()))
	}
	if candle.ShowAxis() {
		t.Fatal("ShowAxis() = true, want false")
	}
	if !candle.ShowGrid() {
		t.Fatal("ShowGrid() = false, want true")
	}
	if !candle.ShowLegend() {
		t.Fatal("ShowLegend() = false, want true")
	}
	if !candle.ShowVolume() {
		t.Fatal("ShowVolume() = false, want true")
	}
	if candle.VolumeHeight() != 2 {
		t.Fatalf("VolumeHeight() = %d, want 2", candle.VolumeHeight())
	}
	if candle.UpStyle() != upStyle {
		t.Fatalf("UpStyle() = %+v, want %+v", candle.UpStyle(), upStyle)
	}
	if candle.DownStyle() != downStyle {
		t.Fatalf("DownStyle() = %+v, want %+v", candle.DownStyle(), downStyle)
	}
	if candle.FlatStyle() != flatStyle {
		t.Fatalf("FlatStyle() = %+v, want %+v", candle.FlatStyle(), flatStyle)
	}
	if candle.WickStyle() != wickStyle {
		t.Fatalf("WickStyle() = %+v, want %+v", candle.WickStyle(), wickStyle)
	}
	if candle.VolumeStyle() != volumeStyle {
		t.Fatalf("VolumeStyle() = %+v, want %+v", candle.VolumeStyle(), volumeStyle)
	}
}

func TestInstanceMeasureAndPaint(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle: "Daily",
		propCandles: []Candle{
			{Label: "M", Open: 100, High: 110, Low: 98, Close: 108},
			{Label: "T", Open: 108, High: 112, Low: 101, Close: 103},
			{Label: "W", Open: 103, High: 106, Low: 102, Close: 103},
		},
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
	if got := strings.TrimRight(bufferRowText(buf, 0), " "); got != "Daily" {
		t.Fatalf("title row = %q, want Daily", got)
	}
	if got := strings.TrimRight(bufferRowText(buf, 5), " "); got != "─────" {
		t.Fatalf("axis row = %q, want %q", got, "─────")
	}
	if got := strings.TrimRight(bufferRowText(buf, 6), " "); got != "M T W" {
		t.Fatalf("label row = %q, want %q", got, "M T W")
	}

	hasCandleGlyph := false
	for y := 1; y <= 4; y++ {
		row := bufferRowText(buf, y)
		if strings.ContainsRune(row, wickRune) ||
			strings.ContainsRune(row, upBodyRune) ||
			strings.ContainsRune(row, downBodyRune) ||
			strings.ContainsRune(row, flatBodyRune) {
			hasCandleGlyph = true
			break
		}
	}
	if !hasCandleGlyph {
		t.Fatal("plot rows do not contain candlestick glyphs")
	}
}

func TestInstanceLegendAndGrid(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle: "Daily",
		propCandles: []Candle{
			{Label: "M", Open: 100, High: 110, Low: 98, Close: 108},
			{Label: "T", Open: 108, High: 112, Low: 101, Close: 103},
			{Label: "W", Open: 103, High: 106, Low: 102, Close: 103},
		},
		propWidth:      5,
		propHeight:     4,
		propShowGrid:   true,
		propShowLegend: true,
		propShowAxis:   false,
	})

	size := inst.Measure(layout.UnboundedConstraints())
	buf := drawCmdsToBuffer(size.Width, size.Height, inst.Paint(0, 0))
	if got := strings.TrimRight(bufferRowText(buf, 1), " "); got != "▓ Up" {
		t.Fatalf("legend row 1 = %q, want %q", got, "▓ Up")
	}
	if got := strings.TrimRight(bufferRowText(buf, 2), " "); got != "█ Down" {
		t.Fatalf("legend row 2 = %q, want %q", got, "█ Down")
	}
	if got := strings.TrimRight(bufferRowText(buf, 3), " "); got != "■ Flat" {
		t.Fatalf("legend row 3 = %q, want %q", got, "■ Flat")
	}

	hasGrid := false
	for y := 4; y < 4+inst.plotHeight(); y++ {
		if strings.ContainsRune(bufferRowText(buf, y), '┈') {
			hasGrid = true
			break
		}
	}
	if !hasGrid {
		t.Fatal("plot rows do not contain grid glyph")
	}
}

func TestInstanceTrendStyles(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propCandles: []Candle{
			{Label: "M", Open: 100, High: 110, Low: 98, Close: 108},
			{Label: "T", Open: 108, High: 112, Low: 101, Close: 103},
			{Label: "W", Open: 103, High: 106, Low: 102, Close: 103},
		},
		propWidth:    5,
		propHeight:   4,
		propShowAxis: false,
	})

	size := inst.Measure(layout.UnboundedConstraints())
	buf := drawCmdsToBuffer(size.Width, size.Height, inst.Paint(0, 0))

	hasUp := false
	hasDown := false
	hasFlat := false
	for y := 0; y < buf.Height; y++ {
		for x := 0; x < buf.Width; x++ {
			cell := buf.Cells[y][x]
			if cell.IsContinuation || cell.Cluster == "" || cell.Cluster == " " {
				continue
			}
			switch cell.Style.FG {
			case palette.UpColor():
				hasUp = true
			case palette.DownColor():
				hasDown = true
			case palette.FlatColor():
				hasFlat = true
			}
		}
	}
	if !hasUp || !hasDown || !hasFlat {
		t.Fatalf("trend colors missing: up=%v down=%v flat=%v", hasUp, hasDown, hasFlat)
	}
}

func TestInstanceVolumePanel(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle: "Daily Volume",
		propCandles: []Candle{
			{Label: "M", Open: 100, High: 110, Low: 98, Close: 108, Volume: 1200},
			{Label: "T", Open: 108, High: 112, Low: 101, Close: 103, Volume: 900},
			{Label: "W", Open: 103, High: 106, Low: 102, Close: 103, Volume: 600},
		},
		propWidth:        5,
		propHeight:       4,
		propShowAxis:     true,
		propShowLegend:   true,
		propShowVolume:   true,
		propVolumeHeight: 2,
	})

	size := inst.Measure(layout.UnboundedConstraints())
	if size.Height != 13 {
		t.Fatalf("Measure().Height = %d, want 13", size.Height)
	}

	buf := drawCmdsToBuffer(size.Width, size.Height, inst.Paint(0, 0))
	if got := strings.TrimRight(bufferRowText(buf, 4), " "); got != "▆ Volume" {
		t.Fatalf("legend volume row = %q, want %q", got, "▆ Volume")
	}
	if got := strings.TrimRight(bufferRowText(buf, 11), " "); got != "─────" {
		t.Fatalf("axis row = %q, want %q", got, "─────")
	}
	if got := strings.TrimRight(bufferRowText(buf, 12), " "); got != "M T W" {
		t.Fatalf("label row = %q, want %q", got, "M T W")
	}

	hasVolumeGlyph := false
	for y := 9; y <= 10; y++ {
		if strings.ContainsRune(bufferRowText(buf, y), volumeRune) {
			hasVolumeGlyph = true
			break
		}
	}
	if !hasVolumeGlyph {
		t.Fatal("volume rows do not contain volume glyphs")
	}
}

func TestCompactLabelRowFoldsDenseDateLabels(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propCandles: []Candle{
			{Label: "03/24", Open: 100, High: 110, Low: 98, Close: 108},
			{Label: "03/25", Open: 108, High: 112, Low: 101, Close: 103},
			{Label: "03/26", Open: 103, High: 106, Low: 102, Close: 103},
			{Label: "03/27", Open: 103, High: 109, Low: 98, Close: 108},
			{Label: "03/28", Open: 108, High: 114, Low: 104, Close: 111},
			{Label: "03/29", Open: 111, High: 116, Low: 109, Close: 115},
		},
		propWidth:    11,
		propHeight:   4,
		propShowAxis: true,
	})

	row := inst.compactLabelRow(inst.visibleLabels(), inst.plotPositions(), '•')
	if row != "4 5 6 7 8 9" {
		t.Fatalf("compactLabelRow() = %q, want %q", row, "4 5 6 7 8 9")
	}
}

func TestInstanceCustomElementStyles(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propCandles: []Candle{
			{Label: "M", Open: 100, High: 110, Low: 95, Close: 107, Volume: 1800},
			{Label: "T", Open: 107, High: 112, Low: 101, Close: 103, Volume: 1200},
			{Label: "W", Open: 103, High: 116, Low: 100, Close: 103, Volume: 900},
		},
		propWidth:        5,
		propHeight:       5,
		propShowAxis:     false,
		propShowVolume:   true,
		propVolumeHeight: 2,
		propUpStyle:      style.NewStyle().Foreground(style.Yellow).Background(style.Blue).Bold(true),
		propDownStyle:    style.NewStyle().Foreground(style.Magenta).Underline(true),
		propFlatStyle:    style.NewStyle().Foreground(style.Cyan).Italic(true),
		propWickStyle:    style.NewStyle().Foreground(style.BrightBlack).Bold(true),
		propVolumeStyle:  style.NewStyle().Foreground(style.BrightWhite).Reverse(true),
	})

	size := inst.Measure(layout.UnboundedConstraints())
	buf := drawCmdsToBuffer(size.Width, size.Height, inst.Paint(0, 0))

	foundUpBody := false
	foundDownBody := false
	foundFlatBody := false
	foundWick := false
	foundVolume := false
	for y := 0; y < buf.Height; y++ {
		for x := 0; x < buf.Width; x++ {
			cell := buf.Cells[y][x]
			if cell.IsContinuation || cell.Cluster == "" || cell.Cluster == " " {
				continue
			}
			switch cell.Cluster {
			case string(upBodyRune):
				if cell.Style.FG == style.Yellow && cell.Style.BG == style.Blue && cell.Style.IsBold() {
					foundUpBody = true
				}
			case string(downBodyRune):
				if cell.Style.FG == style.Magenta && cell.Style.IsUnderline() {
					foundDownBody = true
				}
			case string(flatBodyRune):
				if cell.Style.FG == style.Cyan && cell.Style.IsItalic() {
					foundFlatBody = true
				}
			case string(wickRune):
				if cell.Style.FG == style.BrightBlack && cell.Style.IsBold() {
					foundWick = true
				}
			case string(volumeRune):
				if cell.Style.FG == style.BrightWhite && cell.Style.IsReverse() {
					foundVolume = true
				}
			}
		}
	}

	if !foundUpBody || !foundDownBody || !foundFlatBody || !foundWick || !foundVolume {
		t.Fatalf("custom styles missing: up=%v down=%v flat=%v wick=%v volume=%v", foundUpBody, foundDownBody, foundFlatBody, foundWick, foundVolume)
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
