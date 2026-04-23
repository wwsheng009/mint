package bulletchart

import (
	"strings"
	"testing"

	fwtheme "github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

func TestVNodeBuilder(t *testing.T) {
	vnode := NewBuilder().
		Label("Latency").
		Value(72).
		Target(80).
		Max(100).
		Width(12).
		HigherIsBetter().
		QualitativeRanges(
			QualitativeRange{Limit: 40, Glyph: '░'},
			QualitativeRange{Limit: 70, Glyph: '▒'},
			QualitativeRange{Limit: 100, Glyph: '▓'},
		).
		TargetMarkerRune('╻').
		BelowValueLabel().
		Key("bullet-1").
		Build()

	bullet, ok := vnode.(*VNode)
	if !ok {
		t.Fatal("Build() should return *VNode")
	}
	if bullet.Key() != "bullet-1" {
		t.Fatalf("Key() = %q, want bullet-1", bullet.Key())
	}
	if bullet.Label() != "Latency" {
		t.Fatalf("Label() = %q, want Latency", bullet.Label())
	}
	if bullet.Target() != 80 {
		t.Fatalf("Target() = %d, want 80", bullet.Target())
	}
	if bullet.TargetMarkerRune() != '╻' {
		t.Fatalf("TargetMarkerRune() = %q, want ╻", string(bullet.TargetMarkerRune()))
	}
	if bullet.ValueLabelMode() != ValueLabelModeBelow {
		t.Fatalf("ValueLabelMode() = %v, want Below", bullet.ValueLabelMode())
	}
	if bullet.Direction() != DirectionHigherBetter {
		t.Fatalf("Direction() = %v, want HigherBetter", bullet.Direction())
	}
	if got := bullet.QualitativeRanges(); len(got) != 3 || got[1].Glyph != '▒' {
		t.Fatalf("QualitativeRanges() = %+v, want 3 normalized ranges", got)
	}
}

func TestInstanceMeasureAndPaint(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propLabel:          "Latency",
		propValue:          72,
		propTarget:         80,
		propMax:            100,
		propWidth:          12,
		propShowTarget:     true,
		propShowValueText:  true,
		propValueLabelMode: ValueLabelModeBelow,
		propQualitativeRanges: []QualitativeRange{
			{Limit: 40, Glyph: '░'},
			{Limit: 70, Glyph: '▒'},
			{Limit: 100, Glyph: '▓'},
		},
		propTargetMarkerRune: '╻',
	})

	size := inst.Measure(layout.UnboundedConstraints())
	if size.Width < 12 {
		t.Fatalf("Measure().Width = %d, want >= 12", size.Width)
	}
	if size.Height != 2 {
		t.Fatalf("Measure().Height = %d, want 2", size.Height)
	}

	cmds := inst.Paint(0, 0)
	chartRow := joinRow(cmds, 0)
	if !strings.Contains(chartRow, "╻") {
		t.Fatalf("chart row = %q, want target marker", chartRow)
	}
	if !strings.Contains(chartRow, "▓") {
		t.Fatalf("chart row = %q, want qualitative range glyph", chartRow)
	}
	if metaRow := joinRow(cmds, 1); metaRow != "Latency: 72/100 target 80" {
		t.Fatalf("label row = %q, want %q", metaRow, "Latency: 72/100 target 80")
	}
}

func TestInstanceInlineValueLabelAuto(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propValue:         60,
		propTarget:        75,
		propMax:           100,
		propWidth:         22,
		propShowTarget:    true,
		propShowValueText: true,
	})

	cmds := inst.Paint(0, 0)
	chartRow := joinRow(cmds, 0)
	if !strings.Contains(chartRow, "60/100 target 75") {
		t.Fatalf("inline chart row = %q, want inline value summary", chartRow)
	}
	if metaRow := joinRow(cmds, 1); metaRow != "" {
		t.Fatalf("meta row = %q, want empty for inline mode", metaRow)
	}
}

func TestInstanceAutoValueLabelFallsBackBelowForDenseSummary(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propLabel:         "Latency",
		propValue:         173,
		propTarget:        200,
		propMax:           250,
		propWidth:         22,
		propShowTarget:    true,
		propShowValueText: true,
	})

	cmds := inst.Paint(0, 0)
	if chartRow := joinRow(cmds, 0); strings.Contains(chartRow, "173/250 target 200") {
		t.Fatalf("chart row = %q, want auto mode to move value summary below", chartRow)
	}
	if metaRow := joinRow(cmds, 1); metaRow != "Latency: 173/250 target 200" {
		t.Fatalf("meta row = %q, want %q", metaRow, "Latency: 173/250 target 200")
	}
}

func TestInstanceTargetMarkerStyle(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propValue:             60,
		propTarget:            75,
		propMax:               100,
		propWidth:             20,
		propShowTarget:        true,
		propShowValueText:     false,
		propTargetMarkerRune:  '┆',
		propTargetMarkerStyle: style.NewStyle().Foreground(style.Red).Bold(true),
	})

	cmds := inst.Paint(0, 0)
	foundMarker := false
	for _, cmd := range cmds {
		if !strings.Contains(cmd.Text, "┆") {
			continue
		}
		foundMarker = true
		if cmd.Style.FG != style.Red || !cmd.Style.IsBold() {
			t.Fatalf("target marker style = %+v, want FG=red Bold=true", cmd.Style)
		}
	}
	if !foundMarker {
		t.Fatal("custom target marker not rendered")
	}
}

func TestDefaultQualitativeRangeSemanticStyles(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propValue:         0,
		propMax:           90,
		propWidth:         11,
		propShowTarget:    false,
		propShowValueText: false,
	})

	cells := inst.chartCells()
	if got := cells[0].s.FG; got != fwtheme.Muted() {
		t.Fatalf("low qualitative band FG = %q, want %q", got, fwtheme.Muted())
	}
	if got := cells[4].s.FG; got != fwtheme.Secondary() {
		t.Fatalf("mid qualitative band FG = %q, want %q", got, fwtheme.Secondary())
	}
	if got := cells[8].s.FG; got != fwtheme.Text() {
		t.Fatalf("high qualitative band FG = %q, want %q", got, fwtheme.Text())
	}
}

func TestDefaultTargetMarkerStyleIsEmphasized(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propValue:         60,
		propTarget:        75,
		propMax:           100,
		propWidth:         20,
		propShowTarget:    true,
		propShowValueText: false,
	})

	cells := inst.chartCells()
	target := cells[inst.targetIndex(len(cells))].s
	if target.FG != fwtheme.Warning() {
		t.Fatalf("target marker FG = %q, want %q", target.FG, fwtheme.Warning())
	}
	if !target.IsBold() {
		t.Fatalf("target marker style = %+v, want Bold=true", target)
	}
}

func TestDirectionalSemanticStyles(t *testing.T) {
	higher := NewInstance(rtui.Props{
		propValue:         0,
		propMax:           90,
		propWidth:         11,
		propShowTarget:    false,
		propShowValueText: false,
		propDirection:     DirectionHigherBetter,
	})
	higherCells := higher.chartCells()
	if got := higherCells[0].s.FG; got != fwtheme.Error() {
		t.Fatalf("higher-is-better low band FG = %q, want %q", got, fwtheme.Error())
	}
	if got := higherCells[4].s.FG; got != fwtheme.Warning() {
		t.Fatalf("higher-is-better mid band FG = %q, want %q", got, fwtheme.Warning())
	}
	if got := higherCells[8].s.FG; got != fwtheme.Success() {
		t.Fatalf("higher-is-better high band FG = %q, want %q", got, fwtheme.Success())
	}
	if got := higher.resolveTargetMarkerStyle().FG; got != fwtheme.Success() {
		t.Fatalf("higher-is-better target FG = %q, want %q", got, fwtheme.Success())
	}

	lower := NewInstance(rtui.Props{
		propValue:         0,
		propMax:           90,
		propWidth:         11,
		propShowTarget:    false,
		propShowValueText: false,
		propDirection:     DirectionLowerBetter,
	})
	lowerCells := lower.chartCells()
	if got := lowerCells[0].s.FG; got != fwtheme.Success() {
		t.Fatalf("lower-is-better low band FG = %q, want %q", got, fwtheme.Success())
	}
	if got := lowerCells[4].s.FG; got != fwtheme.Warning() {
		t.Fatalf("lower-is-better mid band FG = %q, want %q", got, fwtheme.Warning())
	}
	if got := lowerCells[8].s.FG; got != fwtheme.Error() {
		t.Fatalf("lower-is-better high band FG = %q, want %q", got, fwtheme.Error())
	}
	if got := lower.resolveTargetMarkerStyle().FG; got != fwtheme.Error() {
		t.Fatalf("lower-is-better target FG = %q, want %q", got, fwtheme.Error())
	}
}

func joinRow(cmds []paint.DrawCmd, y int) string {
	row := ""
	for _, cmd := range cmds {
		if cmd.Y != y {
			continue
		}
		if cmd.X > len(row) {
			row += strings.Repeat(" ", cmd.X-len(row))
		}
		row += cmd.Text
	}
	return strings.TrimRight(row, " ")
}
