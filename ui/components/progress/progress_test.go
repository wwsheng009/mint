package progress

import (
	"strings"
	"testing"

	"github.com/wwsheng009/mint/runtime/layout"
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
}

func TestVNode_Builder(t *testing.T) {
	p := NewBuilder().
		Value(50).
		Max(200).
		Label("Loading").
		Width(40).
		ShowPercent(false).
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
}

func TestVNode_Percent(t *testing.T) {
	tests := []struct {
		value, max, want int
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
	p := New().SetValue(75).SetLabel("Progress")
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
}

// =============================================================================
// Instance Tests
// =============================================================================

func TestInstance_New(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"value": 50,
		"max":   200,
		"label": "Downloading",
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
}

func TestInstance_Measure(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"width": 20,
	})

	size := inst.Measure(layout.UnboundedConstraints())

	if size.Width != 20 {
		t.Errorf("Width = %d, want 20", size.Width)
	}
	// Height = 2 (bar + percentage label)
	if size.Height != 2 {
		t.Errorf("Height = %d, want 2", size.Height)
	}
}

func TestInstance_SetValue(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"max": 100,
	})

	inst.SetValue(50)
	if inst.GetValue() != 50 {
		t.Errorf("Value = %d, want 50", inst.GetValue())
	}

	// Value below 0 should be clamped to 0
	inst.SetValue(-10)
	if inst.GetValue() != 0 {
		t.Errorf("Value = %d, want 0", inst.GetValue())
	}

	// Value above max should be clamped to max
	inst.SetValue(200)
	if inst.GetValue() != 100 {
		t.Errorf("Value = %d, want 100", inst.GetValue())
	}
}

func TestInstance_Percent(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"value": 25,
		"max":   50,
	})

	if inst.Percent() != 50 {
		t.Errorf("Percent = %d, want 50", inst.Percent())
	}
}

func TestInstance_Paint(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"value": 50,
		"max":   100,
		"width": 12,
	})

	cmds := inst.Paint(0, 0)

	// Should have 2 commands: bar + percentage
	if len(cmds) != 2 {
		t.Fatalf("Paint returned %d commands, want 2", len(cmds))
	}

	// Check bar format
	bar := cmds[0].Text
	if !strings.HasPrefix(bar, "[") || !strings.HasSuffix(bar, "]") {
		t.Errorf("Bar should be enclosed in brackets, got %q", bar)
	}

	// Check percentage label
	label := cmds[1].Text
	if label != "50%" {
		t.Errorf("Label = %q, want %q", label, "50%")
	}
}

func TestInstance_Paint_WithLabel(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"value":       75,
		"max":         100,
		"width":       12,
		"label":       "Loading",
		"showPercent": true,
	})

	cmds := inst.Paint(0, 0)

	// Should have bar + label
	if len(cmds) != 2 {
		t.Fatalf("Paint returned %d commands, want 2", len(cmds))
	}

	// Check label format
	label := cmds[1].Text
	if label != "Loading: 75%" {
		t.Errorf("Label = %q, want %q", label, "Loading: 75%")
	}
}

func TestInstance_Paint_NoPercent(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"value":       50,
		"max":         100,
		"width":       12,
		"label":       "Processing",
		"showPercent": false,
	})

	cmds := inst.Paint(0, 0)

	// Should still have 2 commands (bar + label)
	if len(cmds) != 2 {
		t.Fatalf("Paint returned %d commands, want 2", len(cmds))
	}

	// Check label is just the label text
	label := cmds[1].Text
	if label != "Processing" {
		t.Errorf("Label = %q, want %q", label, "Processing")
	}
}
