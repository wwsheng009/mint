package progress

import (
	"strings"
	"testing"
	"time"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// VNode Tests
// =============================================================================

func TestVNode_New(t *testing.T) {
	p := New()
	if p == nil {
		t.Fatal("New returned nil")
	}
	if p.Tag() != "progress" {
		t.Errorf("Tag = %q, want %q", p.Tag(), "progress")
	}
	if p.Max() != 100 {
		t.Errorf("Default max = %d, want 100", p.Max())
	}
	if p.Width() != 30 {
		t.Errorf("Default width = %d, want 30", p.Width())
	}
	if p.ProgressType() != TypeLine {
		t.Errorf("Default type = %v, want %v", p.ProgressType(), TypeLine)
	}
	if p.Status() != StatusNormal {
		t.Errorf("Default status = %v, want %v", p.Status(), StatusNormal)
	}
	if p.GlyphStyle() != GlyphStyleUnicode {
		t.Errorf("Default glyph style = %v, want %v", p.GlyphStyle(), GlyphStyleUnicode)
	}
}

func TestVNode_Builder(t *testing.T) {
	p := NewBuilder().
		Value(50).
		Max(200).
		Label("Loading").
		Width(40).
		Circle().
		Success().
		ASCIIGlyphs().
		ShowPercent(false).
		ShowValue(true).
		Unit("items").
		Key("progress1").
		Build()

	vnode := p.(*VNode)
	if vnode.Value() != 50 {
		t.Errorf("Value = %d, want 50", vnode.Value())
	}
	if vnode.Max() != 200 {
		t.Errorf("Max = %d, want 200", vnode.Max())
	}
	if vnode.Label() != "Loading" {
		t.Errorf("Label = %q, want %q", vnode.Label(), "Loading")
	}
	if vnode.Width() != 40 {
		t.Errorf("Width = %d, want 40", vnode.Width())
	}
	if vnode.ShowPercent() {
		t.Error("ShowPercent should be false")
	}
	if !vnode.ShowValue() {
		t.Error("ShowValue should be true")
	}
	if vnode.Unit() != "items" {
		t.Errorf("Unit = %q, want items", vnode.Unit())
	}
	if vnode.ProgressType() != TypeCircle {
		t.Errorf("Type = %v, want %v", vnode.ProgressType(), TypeCircle)
	}
	if vnode.Status() != StatusSuccess {
		t.Errorf("Status = %v, want %v", vnode.Status(), StatusSuccess)
	}
	if vnode.GlyphStyle() != GlyphStyleASCII {
		t.Errorf("GlyphStyle = %v, want %v", vnode.GlyphStyle(), GlyphStyleASCII)
	}
}

func TestVNode_Builder_WarningAndState(t *testing.T) {
	warning := NewBuilder().Warning().BuildTyped()
	if warning.Status() != StatusWarning {
		t.Fatalf("Warning status = %v, want %v", warning.Status(), StatusWarning)
	}

	state := NewBuilder().State("pending_restart").BuildTyped()
	if state.Status() != StatusWarning {
		t.Fatalf("State status = %v, want %v", state.Status(), StatusWarning)
	}
}

func TestVNode_Builder_Block(t *testing.T) {
	p := NewBuilder().
		Value(60).
		Block().
		Build()

	vnode := p.(*VNode)
	if vnode.ProgressType() != TypeBlock {
		t.Fatalf("Type = %v, want %v", vnode.ProgressType(), TypeBlock)
	}
}

func TestVNode_Builder_Indeterminate(t *testing.T) {
	p := NewBuilder().
		Label("Reloading").
		Indeterminate().
		Build()

	vnode := p.(*VNode)
	if !vnode.IsIndeterminate() {
		t.Fatal("IsIndeterminate() = false, want true")
	}
	if vnode.Status() != StatusActive {
		t.Fatalf("Status() = %v, want %v", vnode.Status(), StatusActive)
	}
	if vnode.Label() != "Reloading" {
		t.Fatalf("Label() = %q, want Reloading", vnode.Label())
	}

	vnode.Determinate()
	if vnode.IsIndeterminate() {
		t.Fatal("IsIndeterminate() after Determinate() = true, want false")
	}
}

func TestVNode_Percent(t *testing.T) {
	tests := []struct {
		value int
		max   int
		want  int
	}{
		{0, 100, 0},
		{50, 100, 50},
		{100, 100, 100},
		{25, 50, 50},
		{1, 3, 33},
	}

	for _, tt := range tests {
		p := New().SetValue(tt.value).SetMax(tt.max)
		if p.Percent() != tt.want {
			t.Errorf("Percent(%d, %d) = %d, want %d", tt.value, tt.max, p.Percent(), tt.want)
		}
	}
}

func TestVNode_CreateInstance(t *testing.T) {
	p := New().SetValue(75).SetLabel("Progress").SetType(TypeDashboard).SetStatus(StatusActive)
	inst := p.CreateInstance()

	if inst == nil {
		t.Fatal("CreateInstance returned nil")
	}

	ci, ok := inst.(*Instance)
	if !ok {
		t.Fatal("Instance is not *Instance")
	}
	if ci.GetValue() != 75 {
		t.Errorf("Value = %d, want 75", ci.GetValue())
	}
	if ci.progressType != TypeDashboard {
		t.Errorf("Type = %v, want %v", ci.progressType, TypeDashboard)
	}
	if ci.status != StatusActive {
		t.Errorf("Status = %v, want %v", ci.status, StatusActive)
	}
}

func TestInstance_IndeterminateLineAnimatesWithoutPercent(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propIndeterminate: true,
		propLabel:         "Reloading",
		propWidth:         12,
	})

	if !inst.WantsTick() {
		t.Fatal("WantsTick() should be true for indeterminate progress")
	}

	cmds := inst.Paint(0, 0)
	if len(cmds) < 2 {
		t.Fatalf("Paint() returned %d commands, want at least 2", len(cmds))
	}
	if got, want := cmds[0].Text, "[●●●·······]"; got != want {
		t.Fatalf("line row = %q, want %q", got, want)
	}
	if got, want := cmds[1].Text, "Reloading: ..."; got != want {
		t.Fatalf("label = %q, want %q", got, want)
	}

	if !inst.Tick(time.Unix(0, 0)) {
		t.Fatal("first Tick should advance indeterminate progress")
	}
	next := inst.Paint(0, 0)[0].Text
	if next == cmds[0].Text {
		t.Fatalf("indeterminate row should change after tick: %q", next)
	}
}

func TestInstance_IndeterminateCircleAndDashboardRenderActiveSegments(t *testing.T) {
	for _, typ := range []Type{TypeCircle, TypeDashboard} {
		inst := NewInstance(rtui.Props{
			propIndeterminate: true,
			propType:          typ,
			propShowPercent:   false,
		})
		rows := inst.visualRows()
		joined := strings.Join(rows, "\n")
		if !strings.Contains(joined, "▓") {
			t.Fatalf("type %v indeterminate rows = %q, want active segment", typ, joined)
		}
		if !strings.Contains(joined, "·") {
			t.Fatalf("type %v indeterminate rows = %q, want track segment", typ, joined)
		}
	}
}

// =============================================================================
// Instance Tests
// =============================================================================

func TestInstance_New(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propValue:  50,
		propMax:    200,
		propLabel:  "Downloading",
		propType:   TypeCircle,
		propStatus: StatusSuccess,
	})

	if inst.GetValue() != 50 {
		t.Errorf("Value = %d, want 50", inst.GetValue())
	}
	if inst.GetMax() != 200 {
		t.Errorf("Max = %d, want 200", inst.GetMax())
	}
	if inst.label != "Downloading" {
		t.Errorf("Label = %q, want %q", inst.label, "Downloading")
	}
	if inst.progressType != TypeCircle {
		t.Errorf("Type = %v, want %v", inst.progressType, TypeCircle)
	}
	if inst.status != StatusSuccess {
		t.Errorf("Status = %v, want %v", inst.status, StatusSuccess)
	}
}

func TestInstance_Measure(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propWidth: 20,
	})

	size := inst.Measure(layout.UnboundedConstraints())
	if size.Width != 20 {
		t.Errorf("Width = %d, want 20", size.Width)
	}
	if size.Height != 2 {
		t.Errorf("Height = %d, want 2", size.Height)
	}
}

func TestInstance_Measure_CircleAndDashboard(t *testing.T) {
	circle := NewInstance(rtui.Props{
		propType:  TypeCircle,
		propWidth: 2,
	})
	circleSize := circle.Measure(layout.UnboundedConstraints())
	if circleSize.Width != circleVisualWidth {
		t.Errorf("circle width = %d, want %d", circleSize.Width, circleVisualWidth)
	}
	if circleSize.Height != circleVisualHeight+1 {
		t.Errorf("circle height = %d, want %d", circleSize.Height, circleVisualHeight+1)
	}

	dashboard := NewInstance(rtui.Props{
		propType:        TypeDashboard,
		propWidth:       4,
		propLabel:       "CPU",
		propShowPercent: false,
	})
	dashboardSize := dashboard.Measure(layout.UnboundedConstraints())
	if dashboardSize.Width != dashboardVisualWidth {
		t.Errorf("dashboard width = %d, want %d", dashboardSize.Width, dashboardVisualWidth)
	}
	if dashboardSize.Height != dashboardVisualHeight+1 {
		t.Errorf("dashboard height = %d, want %d", dashboardSize.Height, dashboardVisualHeight+1)
	}
}

func TestInstance_SetValue(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propMax: 100,
	})

	inst.SetValue(50)
	if inst.GetValue() != 50 {
		t.Errorf("Value = %d, want 50", inst.GetValue())
	}

	inst.SetValue(-10)
	if inst.GetValue() != 0 {
		t.Errorf("Value = %d, want 0", inst.GetValue())
	}

	inst.SetValue(200)
	if inst.GetValue() != 100 {
		t.Errorf("Value = %d, want 100", inst.GetValue())
	}
}

func TestInstance_SetValue_AnimatesDisplayedPercent(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propValue:       20,
		propMax:         100,
		propWidth:       22,
		propShowPercent: false,
	})

	inst.SetValue(80)
	if inst.GetValue() != 80 {
		t.Fatalf("GetValue() = %d, want 80", inst.GetValue())
	}
	if got := inst.Percent(); got != 20 {
		t.Fatalf("Percent() before tick = %d, want 20", got)
	}
	if !inst.WantsTick() {
		t.Fatal("WantsTick() should be true while value tween is active")
	}

	before := inst.Paint(0, 0)[0].Text
	if !inst.Tick(time.Unix(0, 0)) {
		t.Fatal("first Tick should advance value tween")
	}

	midPercent := inst.Percent()
	if midPercent <= 20 || midPercent >= 80 {
		t.Fatalf("mid Percent() = %d, want between 20 and 80", midPercent)
	}

	midBar := inst.Paint(0, 0)[0].Text
	if midBar == before {
		t.Fatalf("bar should change during tween: before=%q after=%q", before, midBar)
	}

	inst.Tick(time.Unix(0, int64(valueTweenDuration)))
	if got := inst.Percent(); got != 80 {
		t.Fatalf("Percent() after completion = %d, want 80", got)
	}
	if inst.WantsTick() {
		t.Fatal("WantsTick() should be false after tween completion")
	}
}

func TestInstance_Percent(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propValue: 25,
		propMax:   50,
	})

	if inst.Percent() != 50 {
		t.Errorf("Percent = %d, want 50", inst.Percent())
	}
}

func TestInstance_SetProps_AnimatesPercentAcrossMaxChange(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propValue:       50,
		propMax:         100,
		propWidth:       22,
		propShowPercent: false,
	})

	changed := inst.SetProps(rtui.Props{
		propValue: 50,
		propMax:   200,
	})
	if !changed {
		t.Fatal("SetProps() should report changed when max changes")
	}
	if got := inst.Percent(); got != 50 {
		t.Fatalf("Percent() before tick = %d, want 50", got)
	}
	if !inst.WantsTick() {
		t.Fatal("WantsTick() should be true while percent tween is active")
	}

	if !inst.Tick(time.Unix(0, 0)) {
		t.Fatal("first Tick should advance percent tween after SetProps")
	}
	midPercent := inst.Percent()
	if midPercent <= 25 || midPercent >= 50 {
		t.Fatalf("mid Percent() = %d, want between 25 and 50", midPercent)
	}

	inst.Tick(time.Unix(0, int64(valueTweenDuration)))
	if got := inst.Percent(); got != 25 {
		t.Fatalf("Percent() after completion = %d, want 25", got)
	}
}

func TestInstance_Paint_Line(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propValue: 50,
		propMax:   100,
		propWidth: 12,
	})

	cmds := inst.Paint(0, 0)
	if len(cmds) != 2 {
		t.Fatalf("Paint returned %d commands, want 2", len(cmds))
	}
	if cmds[0].Text != "[━━━━━·····]" {
		t.Errorf("bar = %q, want %q", cmds[0].Text, "[━━━━━·····]")
	}
	if cmds[1].Text != "50%" {
		t.Errorf("label = %q, want %q", cmds[1].Text, "50%")
	}
}

func TestInstance_Paint_ASCIIGlyphs(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propValue:      50,
		propMax:        100,
		propWidth:      12,
		propGlyphStyle: GlyphStyleASCII,
	})

	cmds := inst.Paint(0, 0)
	if len(cmds) != 2 {
		t.Fatalf("Paint returned %d commands, want 2", len(cmds))
	}
	if cmds[0].Text != "[=====-----]" {
		t.Fatalf("ASCII bar = %q, want %q", cmds[0].Text, "[=====-----]")
	}
}

func TestInstance_Paint_Block(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propType:  TypeBlock,
		propValue: 50,
		propMax:   100,
		propWidth: 12,
	})

	cmds := inst.Paint(0, 0)
	if len(cmds) != 2 {
		t.Fatalf("Paint returned %d commands, want 2", len(cmds))
	}
	if cmds[0].Text != "[█████░░░░░]" {
		t.Errorf("bar = %q, want %q", cmds[0].Text, "[█████░░░░░]")
	}
	if cmds[1].Text != "50%" {
		t.Errorf("label = %q, want %q", cmds[1].Text, "50%")
	}
}

func TestInstance_Paint_WithLabel(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propValue:       75,
		propMax:         100,
		propWidth:       12,
		propLabel:       "Loading",
		propShowPercent: true,
	})

	cmds := inst.Paint(0, 0)
	if len(cmds) != 2 {
		t.Fatalf("Paint returned %d commands, want 2", len(cmds))
	}
	if cmds[1].Text != "Loading: 75%" {
		t.Errorf("Label = %q, want %q", cmds[1].Text, "Loading: 75%")
	}
}

func TestInstance_Paint_WithValueLabel(t *testing.T) {
	tests := []struct {
		name string
		unit string
		want string
	}{
		{name: "word unit", unit: "items", want: "Queue: 42/100 items (42%)"},
		{name: "compact unit", unit: "ms", want: "Latency: 42ms/100ms (42%)"},
		{name: "percent unit", unit: "%", want: "CPU: 42%/100% (42%)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			label := "Queue"
			if tt.unit == "ms" {
				label = "Latency"
			}
			if tt.unit == "%" {
				label = "CPU"
			}
			inst := NewInstance(rtui.Props{
				propValue:       42,
				propMax:         100,
				propWidth:       12,
				propLabel:       label,
				propShowPercent: true,
				propShowValue:   true,
				propUnit:        tt.unit,
			})

			cmds := inst.Paint(0, 0)
			if len(cmds) != 2 {
				t.Fatalf("Paint returned %d commands, want 2", len(cmds))
			}
			if cmds[1].Text != tt.want {
				t.Fatalf("Label = %q, want %q", cmds[1].Text, tt.want)
			}
		})
	}
}

func TestInstance_Paint_ValueLabelWithoutPercent(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propValue:       3,
		propMax:         10,
		propLabel:       "Retries",
		propShowPercent: false,
		propShowValue:   true,
	})

	cmds := inst.Paint(0, 0)
	if len(cmds) != 2 {
		t.Fatalf("Paint returned %d commands, want 2", len(cmds))
	}
	if cmds[1].Text != "Retries: 3/10" {
		t.Fatalf("Label = %q, want value label", cmds[1].Text)
	}
}

func TestInstance_Paint_NoPercent(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propValue:       50,
		propMax:         100,
		propWidth:       12,
		propLabel:       "Processing",
		propShowPercent: false,
	})

	cmds := inst.Paint(0, 0)
	if len(cmds) != 2 {
		t.Fatalf("Paint returned %d commands, want 2", len(cmds))
	}
	if cmds[1].Text != "Processing" {
		t.Errorf("Label = %q, want %q", cmds[1].Text, "Processing")
	}
}

func TestInstance_Paint_Circle(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propType:  TypeCircle,
		propValue: 100,
		propMax:   100,
		propWidth: 5,
	})

	cmds := inst.Paint(0, 0)
	if len(cmds) != 4 {
		t.Fatalf("Paint returned %d commands, want 4", len(cmds))
	}

	rows := drawCmdTexts(cmds)
	if rows[0] != " ███ " {
		t.Errorf("row 0 = %q, want %q", rows[0], " ███ ")
	}
	if rows[1] != "█   █" {
		t.Errorf("row 1 = %q, want %q", rows[1], "█   █")
	}
	if rows[2] != " ███ " {
		t.Errorf("row 2 = %q, want %q", rows[2], " ███ ")
	}
	if rows[3] != "100%" {
		t.Errorf("row 3 = %q, want %q", rows[3], "100%")
	}
}

func TestInstance_Paint_Dashboard(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propType:        TypeDashboard,
		propValue:       100,
		propMax:         100,
		propWidth:       7,
		propLabel:       "CPU",
		propShowPercent: false,
	})

	cmds := inst.Paint(0, 0)
	if len(cmds) != 3 {
		t.Fatalf("Paint returned %d commands, want 3", len(cmds))
	}

	rows := drawCmdTexts(cmds)
	if rows[0] != " █████ " {
		t.Errorf("row 0 = %q, want %q", rows[0], " █████ ")
	}
	if rows[1] != "█     █" {
		t.Errorf("row 1 = %q, want %q", rows[1], "█     █")
	}
	if rows[2] != "CPU" {
		t.Errorf("row 2 = %q, want %q", rows[2], "CPU")
	}
}

func TestInstance_Paint_Circle_PartialSegment(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propType:        TypeCircle,
		propValue:       80,
		propMax:         100,
		propWidth:       5,
		propShowPercent: true,
	})

	rows := drawCmdTexts(inst.Paint(0, 0))
	if rows[0] != " ███ " {
		t.Fatalf("row 0 = %q, want %q", rows[0], " ███ ")
	}
	if rows[1] != "·   █" {
		t.Fatalf("row 1 = %q, want %q", rows[1], "·   █")
	}
	if rows[2] != " ▒██ " {
		t.Fatalf("row 2 = %q, want %q", rows[2], " ▒██ ")
	}
	if rows[3] != "80%" {
		t.Fatalf("row 3 = %q, want %q", rows[3], "80%")
	}
}

func TestInstance_Paint_Dashboard_PartialSegment(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propType:        TypeDashboard,
		propValue:       80,
		propMax:         100,
		propWidth:       7,
		propShowPercent: true,
	})

	rows := drawCmdTexts(inst.Paint(0, 0))
	if rows[0] != " ████▒ " {
		t.Fatalf("row 0 = %q, want %q", rows[0], " ████▒ ")
	}
	if rows[1] != "█     ·" {
		t.Fatalf("row 1 = %q, want %q", rows[1], "█     ·")
	}
	if rows[2] != "80%" {
		t.Fatalf("row 2 = %q, want %q", rows[2], "80%")
	}
}

func TestInstance_Paint_StatusStyles(t *testing.T) {
	tests := []struct {
		name      string
		props     rtui.Props
		status    Status
		wantFG    string
		wantBold  bool
		wantBlink bool
		wantBar   string
	}{
		{name: "success", status: StatusSuccess, wantFG: string(theme.Success()), wantBar: "[━━━━━·····]"},
		{name: "exception", status: StatusException, wantFG: string(theme.Error()), wantBar: "[━━━━━·····]"},
		{name: "warning", status: StatusWarning, wantFG: string(theme.Warning()), wantBar: "[━━━━━·····]"},
		{name: "active", status: StatusActive, wantFG: string(theme.Focus()), wantBold: true, wantBlink: true, wantBar: "[●━━━━·····]"},
		{name: "block active", props: rtui.Props{propType: TypeBlock}, status: StatusActive, wantFG: string(theme.Focus()), wantBold: true, wantBlink: true, wantBar: "[▓████░░░░░]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			props := rtui.Props{
				propValue:       50,
				propMax:         100,
				propWidth:       12,
				propStatus:      tt.status,
				propShowPercent: false,
			}
			for key, value := range tt.props {
				props[key] = value
			}

			inst := NewInstance(props)

			cmds := inst.Paint(0, 0)
			if len(cmds) != 1 {
				t.Fatalf("Paint returned %d commands, want 1", len(cmds))
			}
			if cmds[0].Text != tt.wantBar {
				t.Fatalf("bar = %q, want %q", cmds[0].Text, tt.wantBar)
			}
			if string(cmds[0].Style.FG) != tt.wantFG {
				t.Fatalf("fg = %q, want %q", cmds[0].Style.FG, tt.wantFG)
			}
			if cmds[0].Style.IsBold() != tt.wantBold {
				t.Fatalf("bold = %v, want %v", cmds[0].Style.IsBold(), tt.wantBold)
			}
			if cmds[0].Style.IsBlink() != tt.wantBlink {
				t.Fatalf("blink = %v, want %v", cmds[0].Style.IsBlink(), tt.wantBlink)
			}
		})
	}
}

func TestInstance_WantsTick(t *testing.T) {
	tests := []struct {
		name string
		inst *Instance
		want bool
	}{
		{
			name: "active in progress",
			inst: NewInstance(rtui.Props{
				propValue:  50,
				propMax:    100,
				propStatus: StatusActive,
			}),
			want: true,
		},
		{
			name: "active complete",
			inst: NewInstance(rtui.Props{
				propValue:  100,
				propMax:    100,
				propStatus: StatusActive,
			}),
			want: false,
		},
		{
			name: "normal",
			inst: NewInstance(rtui.Props{
				propValue:  50,
				propMax:    100,
				propStatus: StatusNormal,
			}),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.inst.WantsTick(); got != tt.want {
				t.Fatalf("WantsTick() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInstance_Tick_AnimatesLineActiveProgress(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propValue:       50,
		propMax:         100,
		propWidth:       12,
		propStatus:      StatusActive,
		propShowPercent: false,
	})

	before := inst.Paint(0, 0)
	if got := before[0].Text; got != "[●━━━━·····]" {
		t.Fatalf("initial bar = %q, want %q", got, "[●━━━━·····]")
	}

	if changed := inst.Tick(time.Unix(0, 0)); !changed {
		t.Fatal("first Tick should advance animation")
	}
	after := inst.Paint(0, 0)
	if got := after[0].Text; got != "[━●━━━·····]" {
		t.Fatalf("bar after tick = %q, want %q", got, "[━●━━━·····]")
	}

	if changed := inst.Tick(time.Unix(0, int64(activeTickInterval/2))); changed {
		t.Fatal("tick before interval should not advance animation")
	}
	still := inst.Paint(0, 0)
	if got := still[0].Text; got != "[━●━━━·····]" {
		t.Fatalf("bar after short tick = %q, want %q", got, "[━●━━━·····]")
	}
}

func TestInstance_Tick_AnimatesBlockActiveProgress(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propType:        TypeBlock,
		propValue:       50,
		propMax:         100,
		propWidth:       12,
		propStatus:      StatusActive,
		propShowPercent: false,
	})

	before := inst.Paint(0, 0)
	if got := before[0].Text; got != "[▓████░░░░░]" {
		t.Fatalf("initial bar = %q, want %q", got, "[▓████░░░░░]")
	}

	if changed := inst.Tick(time.Unix(0, 0)); !changed {
		t.Fatal("first Tick should advance animation")
	}
	after := inst.Paint(0, 0)
	if got := after[0].Text; got != "[█▓███░░░░░]" {
		t.Fatalf("bar after tick = %q, want %q", got, "[█▓███░░░░░]")
	}
}

func TestInstance_Tick_AnimatesCircleActiveProgress(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propType:        TypeCircle,
		propValue:       50,
		propMax:         100,
		propStatus:      StatusActive,
		propShowPercent: false,
	})

	before := drawCmdTexts(inst.Paint(0, 0))
	if !inst.Tick(time.Unix(0, 0)) {
		t.Fatal("first Tick should advance circle animation")
	}
	after := drawCmdTexts(inst.Paint(0, 0))

	if before[0] == after[0] && before[1] == after[1] && before[2] == after[2] {
		t.Fatal("circle active animation should change painted rows")
	}
	if !strings.HasPrefix(after[0], " █▓█ ") {
		t.Fatalf("circle active row 0 after tick = %q, want prefix %q", after[0], " █▓█ ")
	}
}

func TestInstance_Tick_AnimatesDashboardActiveProgress(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propType:        TypeDashboard,
		propValue:       70,
		propMax:         100,
		propStatus:      StatusActive,
		propShowPercent: false,
	})

	before := drawCmdTexts(inst.Paint(0, 0))
	if !inst.Tick(time.Unix(0, 0)) {
		t.Fatal("first Tick should advance dashboard animation")
	}
	after := drawCmdTexts(inst.Paint(0, 0))

	if before[0] == after[0] && before[1] == after[1] {
		t.Fatal("dashboard active animation should change painted rows")
	}
	if !strings.HasPrefix(after[0], " ▓██▓· ") {
		t.Fatalf("dashboard active row 0 after tick = %q, want prefix %q", after[0], " ▓██▓· ")
	}
}

func drawCmdTexts(cmds []paint.DrawCmd) []string {
	rows := make([]string, len(cmds))
	for i, cmd := range cmds {
		rows[i] = cmd.Text
	}
	return rows
}
