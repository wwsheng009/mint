package slider

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

func TestVNode_Defaults(t *testing.T) {
	v := New()

	if v.Tag() != "slider" {
		t.Errorf("Tag = %q, want %q", v.Tag(), "slider")
	}

	props := v.Props()
	if props[propMax] != 100 {
		t.Errorf("default max = %v, want 100", props[propMax])
	}
	if props[propStep] != 1 {
		t.Errorf("default step = %v, want 1", props[propStep])
	}
	if props[propWidth] != 20 {
		t.Errorf("default width = %v, want 20", props[propWidth])
	}
	if props[propShowValue] != true {
		t.Errorf("default showValue = %v, want true", props[propShowValue])
	}
	if props[propValue] != 0 {
		t.Errorf("default value = %v, want 0", props[propValue])
	}
}

func TestVNode_Builder(t *testing.T) {
	v := New().
		SetLabel("Volume").
		SetValue(50).
		SetMin(10).
		SetMax(200).
		SetStep(5).
		SetWidth(30).
		SetDisabled(true).
		SetShowValue(false)

	props := v.Props()
	if props[propLabel] != "Volume" {
		t.Errorf("label = %v, want Volume", props[propLabel])
	}
	if props[propValue] != 50 {
		t.Errorf("value = %v, want 50", props[propValue])
	}
	if props[propMin] != 10 {
		t.Errorf("min = %v, want 10", props[propMin])
	}
	if props[propMax] != 200 {
		t.Errorf("max = %v, want 200", props[propMax])
	}
	if props[propStep] != 5 {
		t.Errorf("step = %v, want 5", props[propStep])
	}
	if props[propWidth] != 30 {
		t.Errorf("width = %v, want 30", props[propWidth])
	}
	if props[propDisabled] != true {
		t.Errorf("disabled = %v, want true", props[propDisabled])
	}
	if props[propShowValue] != false {
		t.Errorf("showValue = %v, want false", props[propShowValue])
	}
}

func TestVNode_SetProps(t *testing.T) {
	v := New()
	v.SetProps(rtui.Props{
		propValue: 42,
		propMin:   10,
		propMax:   80,
	})

	if v.Props()[propValue] != 42 {
		t.Errorf("value after SetProps = %v, want 42", v.Props()[propValue])
	}
	if v.Props()[propMin] != 10 {
		t.Errorf("min after SetProps = %v, want 10", v.Props()[propMin])
	}
}

func TestVNode_CreateInstance(t *testing.T) {
	v := New().SetValue(30)
	ci := v.CreateInstance()
	inst, ok := ci.(*Instance)
	if !ok {
		t.Fatal("CreateInstance should return *Instance")
	}
	if inst.GetValue() != 30 {
		t.Errorf("instance value = %d, want 30", inst.GetValue())
	}
}

func TestInstance_Clamping(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		min, max int
		want     int
	}{
		{"below min", -10, 0, 100, 0},
		{"above max", 200, 0, 100, 100},
		{"at min", 0, 0, 100, 0},
		{"at max", 100, 0, 100, 100},
		{"in range", 50, 0, 100, 50},
		{"custom range below", 5, 10, 90, 10},
		{"custom range above", 95, 10, 90, 90},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := NewInstance(rtui.Props{
				propValue: tt.value,
				propMin:   tt.min,
				propMax:   tt.max,
			})
			if inst.GetValue() != tt.want {
				t.Errorf("value = %d, want %d", inst.GetValue(), tt.want)
			}
		})
	}
}

func TestInstance_SetProps_ChangeDetection(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propValue: 50,
		propMin:   0,
		propMax:   100,
	})
	inst.ClearDirty()

	// same props → no change
	changed := inst.SetProps(rtui.Props{
		propValue: 50,
		propMin:   0,
		propMax:   100,
	})
	if changed {
		t.Error("SetProps with same values should return false")
	}

	// different value → change
	changed = inst.SetProps(rtui.Props{
		propValue: 60,
	})
	if !changed {
		t.Error("SetProps with different value should return true")
	}
}

func TestInstance_HandleAction_LeftRight(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propValue: 50,
		propStep:  5,
		propMin:   0,
		propMax:   100,
	})

	// right → increase
	handled := inst.HandleAction(&action.Action{Type: action.ActionCursorRight})
	if !handled {
		t.Error("CursorRight should be handled")
	}
	if inst.GetValue() != 55 {
		t.Errorf("after right: value = %d, want 55", inst.GetValue())
	}

	// left → decrease
	handled = inst.HandleAction(&action.Action{Type: action.ActionCursorLeft})
	if !handled {
		t.Error("CursorLeft should be handled")
	}
	if inst.GetValue() != 50 {
		t.Errorf("after left: value = %d, want 50", inst.GetValue())
	}
}

func TestInstance_HandleAction_HomeEnd(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propValue: 50,
		propMin:   10,
		propMax:   90,
	})

	inst.HandleAction(&action.Action{Type: action.ActionCursorHome})
	if inst.GetValue() != 10 {
		t.Errorf("after Home: value = %d, want 10", inst.GetValue())
	}

	inst.HandleAction(&action.Action{Type: action.ActionCursorEnd})
	if inst.GetValue() != 90 {
		t.Errorf("after End: value = %d, want 90", inst.GetValue())
	}
}

func TestInstance_HandleAction_NavigateAliases(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propValue: 50,
		propStep:  10,
		propMin:   0,
		propMax:   100,
	})

	if !inst.HandleAction(&action.Action{Type: action.ActionNavigateRight}) {
		t.Fatal("NavigateRight should be handled")
	}
	if inst.GetValue() != 60 {
		t.Fatalf("after NavigateRight: value = %d, want 60", inst.GetValue())
	}

	if !inst.HandleAction(&action.Action{Type: action.ActionNavigateHome}) {
		t.Fatal("NavigateHome should be handled")
	}
	if inst.GetValue() != 0 {
		t.Fatalf("after NavigateHome: value = %d, want 0", inst.GetValue())
	}

	if !inst.HandleAction(&action.Action{Type: action.ActionNavigateEnd}) {
		t.Fatal("NavigateEnd should be handled")
	}
	if inst.GetValue() != 100 {
		t.Fatalf("after NavigateEnd: value = %d, want 100", inst.GetValue())
	}
}

func TestInstance_HandleAction_ClampOnStep(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propValue: 98,
		propStep:  5,
		propMax:   100,
	})

	inst.HandleAction(&action.Action{Type: action.ActionCursorRight})
	if inst.GetValue() != 100 {
		t.Errorf("stepping past max: value = %d, want 100", inst.GetValue())
	}

	inst.HandleAction(&action.Action{Type: action.ActionCursorRight})
	if inst.GetValue() != 100 {
		t.Errorf("at max, right again: value = %d, want 100", inst.GetValue())
	}
}

func TestInstance_HandleAction_Disabled(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propValue:    50,
		propDisabled: true,
	})

	handled := inst.HandleAction(&action.Action{Type: action.ActionCursorRight})
	if handled {
		t.Error("disabled slider should not handle actions")
	}
	if inst.GetValue() != 50 {
		t.Errorf("disabled slider value changed: %d, want 50", inst.GetValue())
	}
}

func TestInstance_Paint_BasicTrack(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propValue:     50,
		propMin:       0,
		propMax:       100,
		propWidth:     10,
		propShowValue: true,
	})

	cmds := inst.Paint(0, 0)
	if len(cmds) < 1 {
		t.Fatal("expected at least 1 draw command")
	}

	track := cmds[0].Text
	// 50% of width 10 = 5 filled + 5 unfilled
	expected := "█████░░░░░"
	if track != expected {
		t.Errorf("track = %q, want %q", track, expected)
	}

	// Should have a value label
	if len(cmds) < 2 {
		t.Fatal("expected value label draw command")
	}
	if cmds[1].Text != "50" {
		t.Errorf("value text = %q, want %q", cmds[1].Text, "50")
	}
}

func TestInstance_Paint_WithLabel(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propLabel:     "Vol",
		propValue:     25,
		propMin:       0,
		propMax:       100,
		propWidth:     4,
		propShowValue: false,
	})

	cmds := inst.Paint(0, 0)
	if len(cmds) < 2 {
		t.Fatalf("expected at least 2 draw commands, got %d", len(cmds))
	}

	// first cmd is label
	if cmds[0].Text != "Vol: " {
		t.Errorf("label text = %q, want %q", cmds[0].Text, "Vol: ")
	}
	// second cmd is track
	if cmds[1].Text != "█░░░" {
		t.Errorf("track = %q, want %q", cmds[1].Text, "█░░░")
	}
}

func TestInstance_Paint_NoValue(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propValue:     50,
		propWidth:     10,
		propShowValue: false,
	})

	cmds := inst.Paint(0, 0)
	if len(cmds) != 1 {
		t.Errorf("expected 1 draw command (no value label), got %d", len(cmds))
	}
}

func TestInstance_Measure(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propWidth:     20,
		propShowValue: true,
		propMax:       100,
	})

	size := inst.Measure(layout.Constraints{
		MaxWidth:  100,
		MaxHeight: 10,
	})

	if size.Height != 1 {
		t.Errorf("height = %d, want 1", size.Height)
	}
	if size.Width < 20 {
		t.Errorf("width = %d, want >= 20", size.Width)
	}
}

func TestInstance_Focus(t *testing.T) {
	inst := NewInstance(rtui.Props{})

	if inst.HasFocus() {
		t.Error("should not have focus initially")
	}

	inst.SetFocus(true)
	if !inst.HasFocus() {
		t.Error("should have focus after SetFocus(true)")
	}

	inst.SetFocus(false)
	if inst.HasFocus() {
		t.Error("should not have focus after SetFocus(false)")
	}
}

func TestInstance_IntentEmission(t *testing.T) {
	var emitted intent.Intent

	inst := NewInstance(rtui.Props{
		propValue:        50,
		propStep:         10,
		propChangeIntent: intent.FieldChangeIntent{Field: "volume"},
	})
	inst.SetIntentEmitter(func(i intent.Intent) {
		emitted = i
	})

	inst.HandleAction(&action.Action{Type: action.ActionCursorRight})

	fci, ok := emitted.(intent.FieldChangeIntent)
	if !ok {
		t.Fatalf("expected FieldChangeIntent, got %T", emitted)
	}
	if fci.Field != "volume" {
		t.Errorf("field = %q, want %q", fci.Field, "volume")
	}
	if fci.Value != "60" {
		t.Errorf("value = %q, want %q", fci.Value, "60")
	}
}

func TestFluentBuilder(t *testing.T) {
	node := NewBuilder().
		Label("Brightness").
		Value(75).
		Min(0).
		Max(200).
		Step(10).
		Width(30).
		ShowValue(true).
		Build()

	props := node.Props()
	if props[propLabel] != "Brightness" {
		t.Errorf("label = %v, want Brightness", props[propLabel])
	}
	if props[propValue] != 75 {
		t.Errorf("value = %v, want 75", props[propValue])
	}
	if props[propMax] != 200 {
		t.Errorf("max = %v, want 200", props[propMax])
	}
}

func TestFluentBuilder_BuildInstance(t *testing.T) {
	inst := NewBuilder().
		Value(30).
		Min(0).
		Max(50).
		BuildInstance()

	if inst.GetValue() != 30 {
		t.Errorf("instance value = %d, want 30", inst.GetValue())
	}
}

func TestSlider_Aliases(t *testing.T) {
	s1 := Slider()
	if s1.Tag() != "slider" {
		t.Error("Slider() should return a slider VNode")
	}

	s2 := NewSlider()
	if s2.Tag() != "slider" {
		t.Error("NewSlider() should return a slider VNode")
	}
}
