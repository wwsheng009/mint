package clock

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

func TestVNode_Builder(t *testing.T) {
	fixed := time.Date(2026, 3, 29, 3, 0, 0, 0, time.UTC)
	vnode := NewBuilder().
		Radius(6).
		StaticTime(fixed).
		HideSeconds().
		HideDigital().
		BuildTyped()

	if vnode.Tag() != "clock" {
		t.Fatalf("Tag() = %q, want %q", vnode.Tag(), "clock")
	}
	if vnode.Shape() != DialShapeCircle {
		t.Fatalf("Shape() = %v, want circle", vnode.Shape())
	}
	if vnode.Radius() != 6 {
		t.Fatalf("Radius() = %d, want 6", vnode.Radius())
	}
	if vnode.RadiusY() != 6 {
		t.Fatalf("RadiusY() = %d, want 6", vnode.RadiusY())
	}
	if vnode.Live() {
		t.Fatal("Live() should be false after StaticTime()")
	}
	if vnode.ShowSecondHand() {
		t.Fatal("ShowSecondHand() should be false after HideSeconds()")
	}
	if vnode.ShowDigital() {
		t.Fatal("ShowDigital() should be false after HideDigital()")
	}
	if vnode.Preset() != PresetNone {
		t.Fatalf("Preset() = %v, want none", vnode.Preset())
	}
	if vnode.HandRenderStyle() != HandRenderStyleASCII {
		t.Fatalf("HandRenderStyle() = %v, want ASCII", vnode.HandRenderStyle())
	}
	if !vnode.DialStyle().IsEmpty() || !vnode.TickStyle().IsEmpty() || !vnode.CenterStyle().IsEmpty() || !vnode.DigitalStyle().IsEmpty() {
		t.Fatal("default face styles should be empty")
	}
	if !vnode.HourHandStyle().IsEmpty() || !vnode.MinuteHandStyle().IsEmpty() || !vnode.SecondHandStyle().IsEmpty() {
		t.Fatal("default hand styles should be empty")
	}
	if !vnode.TimeValue().Equal(fixed) {
		t.Fatalf("TimeValue() = %v, want %v", vnode.TimeValue(), fixed)
	}
}

func TestVNode_UnicodeHandsBuilder(t *testing.T) {
	vnode := NewBuilder().
		UnicodeHands().
		BuildTyped()

	if vnode.HandRenderStyle() != HandRenderStyleUnicode {
		t.Fatalf("HandRenderStyle() = %v, want Unicode", vnode.HandRenderStyle())
	}
}

func TestVNode_EllipseBuilder(t *testing.T) {
	vnode := NewBuilder().
		Ellipse().
		Radii(7, 4).
		BuildTyped()

	if vnode.Shape() != DialShapeEllipse {
		t.Fatalf("Shape() = %v, want ellipse", vnode.Shape())
	}
	if vnode.RadiusX() != 7 {
		t.Fatalf("RadiusX() = %d, want 7", vnode.RadiusX())
	}
	if vnode.RadiusY() != 4 {
		t.Fatalf("RadiusY() = %d, want 4", vnode.RadiusY())
	}
}

func TestVNode_CellAspectBuilder(t *testing.T) {
	vnode := NewBuilder().
		CellAspectX(1.5).
		BuildTyped()

	if vnode.CellAspectX() != 1.5 {
		t.Fatalf("CellAspectX() = %v, want 1.5", vnode.CellAspectX())
	}
}

func TestVNode_InvalidCellAspectFallsBackToDefault(t *testing.T) {
	vnode := NewBuilder().
		CellAspectX(0).
		BuildTyped()

	if vnode.CellAspectX() != DefaultCellAspectX {
		t.Fatalf("CellAspectX() = %v, want default %v", vnode.CellAspectX(), DefaultCellAspectX)
	}
}

func TestVNode_PresetBuilder(t *testing.T) {
	vnode := NewBuilder().
		Preset(PresetNeon).
		BuildTyped()

	if vnode.Preset() != PresetNeon {
		t.Fatalf("Preset() = %v, want neon", vnode.Preset())
	}
}

func TestThemeForPreset(t *testing.T) {
	theme := ThemeForPreset(PresetClassic)

	if theme.DialStyle.FG != style.BrightBlack {
		t.Fatalf("ThemeForPreset(PresetClassic).DialStyle.FG = %q, want bright-black", theme.DialStyle.FG)
	}
	if theme.TickStyle.FG != style.BrightWhite {
		t.Fatalf("ThemeForPreset(PresetClassic).TickStyle.FG = %q, want bright-white", theme.TickStyle.FG)
	}
	if theme.HourHandStyle.FG != style.BrightYellow {
		t.Fatalf("ThemeForPreset(PresetClassic).HourHandStyle.FG = %q, want bright-yellow", theme.HourHandStyle.FG)
	}
}

func TestThemePreset(t *testing.T) {
	if got := ThemePreset(PresetNeon); got != ThemeForPreset(PresetNeon) {
		t.Fatal("ThemePreset(PresetNeon) should match ThemeForPreset(PresetNeon)")
	}
}

func TestTheme_Merge(t *testing.T) {
	base := ThemeForPreset(PresetClassic)
	overlay := Theme{}.
		WithBaseStyle(style.Style{}.Background(style.Black)).
		WithDigitalStyle(style.Style{}.Foreground(style.Magenta)).
		WithSecondHandStyle(style.Style{}.Foreground(style.Green))

	merged := base.Merge(overlay)

	if merged.BaseStyle.FG != style.BrightWhite {
		t.Fatalf("merged.BaseStyle.FG = %q, want bright-white", merged.BaseStyle.FG)
	}
	if merged.BaseStyle.BG != style.Black {
		t.Fatalf("merged.BaseStyle.BG = %q, want black", merged.BaseStyle.BG)
	}
	if merged.DigitalStyle.FG != style.Magenta {
		t.Fatalf("merged.DigitalStyle.FG = %q, want magenta", merged.DigitalStyle.FG)
	}
	if merged.SecondHandStyle.FG != style.Green {
		t.Fatalf("merged.SecondHandStyle.FG = %q, want green", merged.SecondHandStyle.FG)
	}
	if merged.HourHandStyle.FG != style.BrightYellow {
		t.Fatalf("merged.HourHandStyle.FG = %q, want bright-yellow", merged.HourHandStyle.FG)
	}
}

func TestTheme_WithPreset(t *testing.T) {
	theme := Theme{}.
		WithDigitalStyle(style.Style{}.Foreground(style.Magenta)).
		WithSecondHandStyle(style.Style{}.Foreground(style.Green)).
		WithPreset(PresetClassic)

	if theme.DialStyle.FG != style.BrightBlack {
		t.Fatalf("theme.DialStyle.FG = %q, want bright-black", theme.DialStyle.FG)
	}
	if theme.DigitalStyle.FG != style.Magenta {
		t.Fatalf("theme.DigitalStyle.FG = %q, want magenta", theme.DigitalStyle.FG)
	}
	if theme.SecondHandStyle.FG != style.Green {
		t.Fatalf("theme.SecondHandStyle.FG = %q, want green", theme.SecondHandStyle.FG)
	}
}

func TestVNode_ThemeBuilder(t *testing.T) {
	theme := ThemeForPreset(PresetNeon).
		WithDigitalStyle(style.Style{}.Foreground(style.Magenta)).
		WithSecondHandStyle(style.Style{}.Foreground(style.Green))

	vnode := NewBuilder().
		Theme(theme).
		BuildTyped()

	if vnode.Preset() != PresetNone {
		t.Fatalf("Preset() = %v, want none when only Theme(...) is applied", vnode.Preset())
	}
	if vnode.Theme().BaseStyle.FG != style.BrightCyan {
		t.Fatalf("Theme().BaseStyle.FG = %q, want bright-cyan", vnode.Theme().BaseStyle.FG)
	}
	if vnode.DigitalStyle().FG != style.Magenta {
		t.Fatalf("DigitalStyle().FG = %q, want magenta", vnode.DigitalStyle().FG)
	}
	if vnode.SecondHandStyle().FG != style.Green {
		t.Fatalf("SecondHandStyle().FG = %q, want green", vnode.SecondHandStyle().FG)
	}
}

func TestVNode_HandStylesBuilder(t *testing.T) {
	vnode := NewBuilder().
		DialStyle(style.Style{}.Foreground(style.BrightBlack)).
		TickStyle(style.Style{}.Foreground(style.White)).
		CenterStyle(style.Style{}.Foreground(style.Yellow)).
		DigitalStyle(style.Style{}.Foreground(style.Cyan)).
		HourHandStyle(style.Style{}.Foreground(style.Red)).
		MinuteHandStyle(style.Style{}.Foreground(style.Green)).
		SecondHandStyle(style.Style{}.Foreground(style.Blue)).
		BuildTyped()

	if vnode.DialStyle().FG != style.BrightBlack {
		t.Fatalf("DialStyle().FG = %q, want bright-black", vnode.DialStyle().FG)
	}
	if vnode.TickStyle().FG != style.White {
		t.Fatalf("TickStyle().FG = %q, want white", vnode.TickStyle().FG)
	}
	if vnode.CenterStyle().FG != style.Yellow {
		t.Fatalf("CenterStyle().FG = %q, want yellow", vnode.CenterStyle().FG)
	}
	if vnode.DigitalStyle().FG != style.Cyan {
		t.Fatalf("DigitalStyle().FG = %q, want cyan", vnode.DigitalStyle().FG)
	}
	if vnode.HourHandStyle().FG != style.Red {
		t.Fatalf("HourHandStyle().FG = %q, want red", vnode.HourHandStyle().FG)
	}
	if vnode.MinuteHandStyle().FG != style.Green {
		t.Fatalf("MinuteHandStyle().FG = %q, want green", vnode.MinuteHandStyle().FG)
	}
	if vnode.SecondHandStyle().FG != style.Blue {
		t.Fatalf("SecondHandStyle().FG = %q, want blue", vnode.SecondHandStyle().FG)
	}
}

func TestInstance_Paint_StaticTime(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propRadius:         4,
		propLive:           false,
		propTime:           time.Date(2026, 3, 29, 3, 0, 0, 0, time.UTC),
		propShowSecondHand: false,
		propShowDigital:    true,
	})

	rows := inst.DebugRows()
	if len(rows) != 10 {
		t.Fatalf("Paint returned %d rows, want 10", len(rows))
	}
	if len(rows[4]) != 17 {
		t.Fatalf("center row width = %d, want 17", len(rows[4]))
	}
	if !strings.Contains(rows[4], "@---O") {
		t.Fatalf("center row = %q, want @---O hand segment", rows[4])
	}
	centerX := inst.centerX()
	if len(rows[1]) <= centerX || rows[1][centerX] != '+' {
		t.Fatalf("top minute hand row = %q, want + at center column %d", rows[1], centerX)
	}
	if rows[9] != "03:00" {
		t.Fatalf("digital text = %q, want %q", rows[9], "03:00")
	}
	if strings.Contains(strings.Join(rows[:9], "\n"), ".") {
		t.Fatal("static clock with hidden seconds should not render second hand tip")
	}
}

func TestInstance_Measure(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propRadius:         4,
		propLive:           false,
		propTime:           time.Date(2026, 3, 29, 9, 15, 30, 0, time.UTC),
		propShowSecondHand: true,
		propShowDigital:    true,
	})

	size := inst.Measure(layout.UnboundedConstraints())
	if size.Width != 17 {
		t.Fatalf("Width = %d, want 17", size.Width)
	}
	if size.Height != 10 {
		t.Fatalf("Height = %d, want 10", size.Height)
	}
}

func TestInstance_Measure_CustomCellAspectX(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propRadius:         4,
		propCellAspectX:    1.0,
		propLive:           false,
		propTime:           time.Date(2026, 3, 29, 9, 15, 30, 0, time.UTC),
		propShowSecondHand: true,
		propShowDigital:    true,
	})

	size := inst.Measure(layout.UnboundedConstraints())
	if size.Width != 9 {
		t.Fatalf("Width = %d, want 9 with cellAspectX=1.0", size.Width)
	}
	if size.Height != 10 {
		t.Fatalf("Height = %d, want 10", size.Height)
	}
}

func TestInstance_Measure_InvalidCellAspectFallsBackToDefault(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propRadius:         4,
		propCellAspectX:    0.0,
		propLive:           false,
		propTime:           time.Date(2026, 3, 29, 9, 15, 30, 0, time.UTC),
		propShowSecondHand: true,
		propShowDigital:    true,
	})

	size := inst.Measure(layout.UnboundedConstraints())
	if size.Width != 17 {
		t.Fatalf("Width = %d, want 17 after default aspect fallback", size.Width)
	}
}

func TestInstance_Measure_Ellipse(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propShape:          DialShapeEllipse,
		propRadiusX:        6,
		propRadiusY:        4,
		propLive:           false,
		propTime:           time.Date(2026, 3, 29, 9, 15, 30, 0, time.UTC),
		propShowSecondHand: true,
		propShowDigital:    true,
	})

	size := inst.Measure(layout.UnboundedConstraints())
	if size.Width != 25 {
		t.Fatalf("Width = %d, want 25", size.Width)
	}
	if size.Height != 10 {
		t.Fatalf("Height = %d, want 10", size.Height)
	}
}

func TestInstance_CircleIgnoresVerticalRadiusMismatch(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propShape:       DialShapeCircle,
		propRadiusX:     6,
		propRadiusY:     4,
		propLive:        false,
		propShowDigital: false,
	})

	size := inst.Measure(layout.UnboundedConstraints())
	if size.Width != 25 {
		t.Fatalf("Width = %d, want 25", size.Width)
	}
	if size.Height != 13 {
		t.Fatalf("Height = %d, want 13", size.Height)
	}
}

func TestInstance_Measure_LargerCellAspectX(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propRadius:         4,
		propCellAspectX:    3.0,
		propLive:           false,
		propTime:           time.Date(2026, 3, 29, 9, 15, 30, 0, time.UTC),
		propShowSecondHand: true,
		propShowDigital:    true,
	})

	size := inst.Measure(layout.UnboundedConstraints())
	if size.Width != 25 {
		t.Fatalf("Width = %d, want 25 with cellAspectX=3.0", size.Width)
	}
}

func TestInstance_Paint_StaticTime_UnicodeHands(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propRadius:         4,
		propLive:           false,
		propTime:           time.Date(2026, 3, 29, 3, 0, 0, 0, time.UTC),
		propShowSecondHand: false,
		propShowDigital:    true,
		propHandStyle:      HandRenderStyleUnicode,
	})

	rows := inst.DebugRows()
	if !strings.Contains(rows[4], "@───■") {
		t.Fatalf("center row = %q, want @───■ hand segment", rows[4])
	}
	topRow := []rune(rows[1])
	if len(topRow) <= inst.centerX() || topRow[inst.centerX()] != '●' {
		t.Fatalf("top minute hand row = %q, want ● at center column", rows[1])
	}
}

func TestInstance_Paint_StaticTime_Ellipse(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propShape:          DialShapeEllipse,
		propRadiusX:        6,
		propRadiusY:        4,
		propLive:           false,
		propTime:           time.Date(2026, 3, 29, 3, 0, 0, 0, time.UTC),
		propShowSecondHand: false,
		propShowDigital:    true,
	})

	rows := inst.DebugRows()
	if len(rows) != 10 {
		t.Fatalf("Paint returned %d rows, want 10", len(rows))
	}
	if len(rows[4]) != 25 {
		t.Fatalf("ellipse center row width = %d, want 25", len(rows[4]))
	}
	if !strings.Contains(rows[4], "@-----O") {
		t.Fatalf("ellipse center row = %q, want @-----O hand segment", rows[4])
	}
	if rows[9] != "03:00" {
		t.Fatalf("digital text = %q, want %q", rows[9], "03:00")
	}
}

func TestInstance_Paint_HandStyles(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propRadius:          4,
		propLive:            false,
		propTime:            time.Date(2026, 3, 29, 9, 15, 30, 0, time.UTC),
		propShowSecondHand:  true,
		propShowDigital:     true,
		propHourHandStyle:   style.Style{}.Foreground(style.Red),
		propMinuteHandStyle: style.Style{}.Foreground(style.Green),
		propSecondHandStyle: style.Style{}.Foreground(style.Blue),
	})

	buf := renderClockToBuffer(inst)

	hourX, hourY := handTipFor(inst, 9.258333333333333/12, maxInt(1, int(math.Round(float64(inst.renderRadiusX())*0.5))), maxInt(1, int(math.Round(float64(inst.radiusY)*0.5))))
	minuteX, minuteY := handTipFor(inst, 15.5/60, maxInt(1, int(math.Round(float64(inst.renderRadiusX())*0.75))), maxInt(1, int(math.Round(float64(inst.radiusY)*0.75))))
	secondX, secondY := handTipFor(inst, 30.0/60, maxInt(1, inst.renderRadiusX()-1), maxInt(1, inst.radiusY-1))

	if got := buf.GetContent(hourX, hourY).Style.FG; got != style.Red {
		t.Fatalf("hour hand FG = %q, want red", got)
	}
	if got := buf.GetContent(minuteX, minuteY).Style.FG; got != style.Green {
		t.Fatalf("minute hand FG = %q, want green", got)
	}
	if got := buf.GetContent(secondX, secondY).Style.FG; got != style.Blue {
		t.Fatalf("second hand FG = %q, want blue", got)
	}
}

func TestInstance_Paint_FaceStyles(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propRadius:       4,
		propLive:         false,
		propTime:         time.Date(2026, 3, 29, 3, 0, 0, 0, time.UTC),
		propShowDigital:  true,
		propDialStyle:    style.Style{}.Foreground(style.BrightBlack),
		propTickStyle:    style.Style{}.Foreground(style.White),
		propCenterStyle:  style.Style{}.Foreground(style.Yellow),
		propDigitalStyle: style.Style{}.Foreground(style.Cyan),
	})

	buf := renderClockToBuffer(inst)

	dialX, ok := findRuneOnRow(buf, 0, 'o')
	if !ok {
		t.Fatal("did not find dial glyph on top row")
	}
	if got := buf.GetContent(dialX, 0).Style.FG; got != style.BrightBlack {
		t.Fatalf("dial FG = %q, want bright-black", got)
	}
	tickX, tickY := inst.tickCellFor(11)
	if got := buf.GetContent(tickX, tickY).Style.FG; got != style.White {
		t.Fatalf("tick FG = %q, want white", got)
	}
	if got := buf.GetContent(inst.centerX(), inst.centerY()).Style.FG; got != style.Yellow {
		t.Fatalf("center FG = %q, want yellow", got)
	}
	if got := buf.GetContent(0, inst.heightCells()).Style.FG; got != style.Cyan {
		t.Fatalf("digital FG = %q, want cyan", got)
	}
}

func TestInstance_Paint_ClassicPresetStyles(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propRadius:         4,
		propLive:           false,
		propTime:           time.Date(2026, 3, 29, 9, 15, 30, 0, time.UTC),
		propShowSecondHand: true,
		propShowDigital:    true,
		propPreset:         PresetClassic,
	})

	buf := renderClockToBuffer(inst)

	dialX, ok := findRuneOnRow(buf, 0, 'o')
	if !ok {
		t.Fatal("did not find classic dial glyph on top row")
	}
	if got := buf.GetContent(dialX, 0).Style.FG; got != style.BrightBlack {
		t.Fatalf("classic dial FG = %q, want bright-black", got)
	}
	tickX, tickY := inst.tickCellFor(11)
	if got := buf.GetContent(tickX, tickY).Style.FG; got != style.BrightWhite {
		t.Fatalf("classic tick FG = %q, want bright-white", got)
	}
	if got := buf.GetContent(inst.centerX(), inst.centerY()).Style.FG; got != style.Yellow {
		t.Fatalf("classic center FG = %q, want yellow", got)
	}
	if got := buf.GetContent(0, inst.heightCells()).Style.FG; got != style.Cyan {
		t.Fatalf("classic digital FG = %q, want cyan", got)
	}
	hourX, hourY := handTipFor(inst, 9.258333333333333/12, maxInt(1, int(math.Round(float64(inst.renderRadiusX())*0.5))), maxInt(1, int(math.Round(float64(inst.radiusY)*0.5))))
	minuteX, minuteY := handTipFor(inst, 15.5/60, maxInt(1, int(math.Round(float64(inst.renderRadiusX())*0.75))), maxInt(1, int(math.Round(float64(inst.radiusY)*0.75))))
	secondX, secondY := handTipFor(inst, 30.0/60, maxInt(1, inst.renderRadiusX()-1), maxInt(1, inst.radiusY-1))
	if got := buf.GetContent(hourX, hourY).Style.FG; got != style.BrightYellow {
		t.Fatalf("classic hour FG = %q, want bright-yellow", got)
	}
	if got := buf.GetContent(minuteX, minuteY).Style.FG; got != style.BrightCyan {
		t.Fatalf("classic minute FG = %q, want bright-cyan", got)
	}
	if got := buf.GetContent(secondX, secondY).Style.FG; got != style.BrightRed {
		t.Fatalf("classic second FG = %q, want bright-red", got)
	}
}

func TestInstance_Paint_ExplicitStylesOverridePreset(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propRadius:         4,
		propLive:           false,
		propTime:           time.Date(2026, 3, 29, 9, 15, 30, 0, time.UTC),
		propShowSecondHand: true,
		propShowDigital:    true,
		propPreset:         PresetClassic,
		propHourHandStyle:  style.Style{}.Foreground(style.Green),
		propDigitalStyle:   style.Style{}.Foreground(style.Magenta),
	})

	buf := renderClockToBuffer(inst)

	hourX, hourY := handTipFor(inst, 9.258333333333333/12, maxInt(1, int(math.Round(float64(inst.renderRadiusX())*0.5))), maxInt(1, int(math.Round(float64(inst.radiusY)*0.5))))
	if got := buf.GetContent(hourX, hourY).Style.FG; got != style.Green {
		t.Fatalf("explicit hour FG = %q, want green", got)
	}
	if got := buf.GetContent(0, inst.heightCells()).Style.FG; got != style.Magenta {
		t.Fatalf("explicit digital FG = %q, want magenta", got)
	}
}

func TestInstance_Paint_ThemeObjectOverlaysPreset(t *testing.T) {
	vnode := NewBuilder().
		Radius(4).
		StaticTime(time.Date(2026, 3, 29, 9, 15, 30, 0, time.UTC)).
		ShowSecondHand(true).
		ShowDigital(true).
		Preset(PresetClassic).
		Theme(Theme{
			DigitalStyle:    style.Style{}.Foreground(style.Magenta),
			SecondHandStyle: style.Style{}.Foreground(style.Green),
		}).
		BuildTyped()

	inst := NewInstance(vnode.Props())
	buf := renderClockToBuffer(inst)

	dialX, ok := findRuneOnRow(buf, 0, 'o')
	if !ok {
		t.Fatal("did not find preset dial glyph on top row")
	}
	if got := buf.GetContent(dialX, 0).Style.FG; got != style.BrightBlack {
		t.Fatalf("preset dial FG = %q, want bright-black", got)
	}
	secondX, secondY := handTipFor(inst, 30.0/60, maxInt(1, inst.renderRadiusX()-1), maxInt(1, inst.radiusY-1))
	if got := buf.GetContent(secondX, secondY).Style.FG; got != style.Green {
		t.Fatalf("theme second-hand FG = %q, want green", got)
	}
	if got := buf.GetContent(0, inst.heightCells()).Style.FG; got != style.Magenta {
		t.Fatalf("theme digital FG = %q, want magenta", got)
	}
}

func TestInstance_Paint_MergedTheme(t *testing.T) {
	mergedTheme := ThemeForPreset(PresetClassic).Merge(
		Theme{}.
			WithDigitalStyle(style.Style{}.Foreground(style.Magenta)).
			WithSecondHandStyle(style.Style{}.Foreground(style.Green)).
			WithCenterStyle(style.Style{}.Foreground(style.White)),
	)

	vnode := NewBuilder().
		Radius(4).
		StaticTime(time.Date(2026, 3, 29, 9, 15, 30, 0, time.UTC)).
		ShowSecondHand(true).
		ShowDigital(true).
		Theme(mergedTheme).
		BuildTyped()

	inst := NewInstance(vnode.Props())
	buf := renderClockToBuffer(inst)

	dialX, ok := findRuneOnRow(buf, 0, 'o')
	if !ok {
		t.Fatal("did not find merged-theme dial glyph on top row")
	}
	if got := buf.GetContent(dialX, 0).Style.FG; got != style.BrightBlack {
		t.Fatalf("merged-theme dial FG = %q, want bright-black", got)
	}
	if got := buf.GetContent(inst.centerX(), inst.centerY()).Style.FG; got != style.White {
		t.Fatalf("merged-theme center FG = %q, want white", got)
	}
	secondX, secondY := handTipFor(inst, 30.0/60, maxInt(1, inst.renderRadiusX()-1), maxInt(1, inst.radiusY-1))
	if got := buf.GetContent(secondX, secondY).Style.FG; got != style.Green {
		t.Fatalf("merged-theme second-hand FG = %q, want green", got)
	}
	if got := buf.GetContent(0, inst.heightCells()).Style.FG; got != style.Magenta {
		t.Fatalf("merged-theme digital FG = %q, want magenta", got)
	}
}

func TestInstance_Paint_WithPresetTheme(t *testing.T) {
	vnode := NewBuilder().
		Radius(4).
		StaticTime(time.Date(2026, 3, 29, 9, 15, 30, 0, time.UTC)).
		ShowSecondHand(true).
		ShowDigital(true).
		Theme(Theme{}.
			WithDigitalStyle(style.Style{}.Foreground(style.Magenta)).
			WithSecondHandStyle(style.Style{}.Foreground(style.Green)).
			WithPreset(PresetClassic)).
		BuildTyped()

	inst := NewInstance(vnode.Props())
	buf := renderClockToBuffer(inst)

	dialX, ok := findRuneOnRow(buf, 0, 'o')
	if !ok {
		t.Fatal("did not find with-preset dial glyph on top row")
	}
	if got := buf.GetContent(dialX, 0).Style.FG; got != style.BrightBlack {
		t.Fatalf("with-preset dial FG = %q, want bright-black", got)
	}
	secondX, secondY := handTipFor(inst, 30.0/60, maxInt(1, inst.renderRadiusX()-1), maxInt(1, inst.radiusY-1))
	if got := buf.GetContent(secondX, secondY).Style.FG; got != style.Green {
		t.Fatalf("with-preset second-hand FG = %q, want green", got)
	}
	if got := buf.GetContent(0, inst.heightCells()).Style.FG; got != style.Magenta {
		t.Fatalf("with-preset digital FG = %q, want magenta", got)
	}
}

func TestInstance_WantsTick(t *testing.T) {
	live := NewInstance(rtui.Props{})
	if !live.WantsTick() {
		t.Fatal("live clock should want ticks")
	}

	static := NewInstance(rtui.Props{
		propLive: false,
		propTime: time.Date(2026, 3, 29, 12, 0, 0, 0, time.UTC),
	})
	if static.WantsTick() {
		t.Fatal("static clock should not want ticks")
	}
}

func TestInstance_Tick_LiveClockUpdatesDisplayTime(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propLive:           true,
		propShowSecondHand: false,
	})
	inst.displayTime = time.Date(2026, 3, 29, 1, 2, 0, 0, time.UTC)

	start := time.Date(2026, 3, 29, 1, 2, 58, 0, time.UTC)
	if changed := inst.Tick(start); changed {
		t.Fatal("first Tick should only prime the loop")
	}

	next := start.Add(2 * time.Second)
	if changed := inst.Tick(next); !changed {
		t.Fatal("second Tick should update live display time")
	}
	if got := inst.displayTime.Format("15:04"); got != "01:03" {
		t.Fatalf("displayTime minute = %q, want %q", got, "01:03")
	}
}

func drawCmdTexts(cmds []paint.DrawCmd) []string {
	rows := make([]string, len(cmds))
	for i, cmd := range cmds {
		rows[i] = cmd.Text
	}
	return rows
}

func renderClockToBuffer(inst *Instance) *paint.Buffer {
	size := inst.Measure(layout.UnboundedConstraints())
	buf := paint.NewBuffer(size.Width, size.Height)
	for _, cmd := range inst.Paint(0, 0) {
		buf.SetString(cmd.X, cmd.Y, cmd.Text, cmd.Style)
	}
	return buf
}

func handTipFor(inst *Instance, fraction float64, lengthX, lengthY int) (int, int) {
	angle := clockAngle(fraction)
	return inst.centerX() + int(math.Round(math.Cos(angle)*float64(lengthX))),
		inst.centerY() + int(math.Round(math.Sin(angle)*float64(lengthY)))
}

func findRuneOnRow(buf *paint.Buffer, y int, glyph rune) (int, bool) {
	for x := 0; x < buf.Width; x++ {
		cluster := buf.GetContent(x, y).Cluster
		if cluster != "" && []rune(cluster)[0] == glyph {
			return x, true
		}
	}
	return 0, false
}
