package rate

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

func TestVNode_Defaults(t *testing.T) {
	v := New()

	if v.Tag() != "rate" {
		t.Errorf("Tag = %q, want %q", v.Tag(), "rate")
	}

	props := v.Props()
	if props[propCount] != defaultCount {
		t.Errorf("default count = %v, want %d", props[propCount], defaultCount)
	}
	if props[propAllowClear] != true {
		t.Errorf("default allowClear = %v, want true", props[propAllowClear])
	}
	if props[propCharacter] != defaultCharacter {
		t.Errorf("default character = %v, want %q", props[propCharacter], defaultCharacter)
	}
	if props[propValue] != 0 {
		t.Errorf("default value = %v, want 0", props[propValue])
	}
}

func TestVNode_Builder(t *testing.T) {
	v := New().
		SetLabel("Quality").
		SetValue(3).
		SetCount(7).
		SetAllowClear(false).
		SetCharacter("*").
		SetDisabled(true).
		SetShowValue(true)

	props := v.Props()
	if props[propLabel] != "Quality" {
		t.Errorf("label = %v, want Quality", props[propLabel])
	}
	if props[propValue] != 3 {
		t.Errorf("value = %v, want 3", props[propValue])
	}
	if props[propCount] != 7 {
		t.Errorf("count = %v, want 7", props[propCount])
	}
	if props[propAllowClear] != false {
		t.Errorf("allowClear = %v, want false", props[propAllowClear])
	}
	if props[propCharacter] != "*" {
		t.Errorf("character = %v, want *", props[propCharacter])
	}
	if props[propDisabled] != true {
		t.Errorf("disabled = %v, want true", props[propDisabled])
	}
	if props[propShowValue] != true {
		t.Errorf("showValue = %v, want true", props[propShowValue])
	}
}

func TestVNode_SetProps(t *testing.T) {
	v := New()
	v.SetProps(rtui.Props{
		propValue: 4,
		propCount: 8,
	})

	if v.Props()[propValue] != 4 {
		t.Errorf("value after SetProps = %v, want 4", v.Props()[propValue])
	}
	if v.Props()[propCount] != 8 {
		t.Errorf("count after SetProps = %v, want 8", v.Props()[propCount])
	}
}

func TestVNode_CreateInstance(t *testing.T) {
	v := New().SetValue(4).SetCount(5)
	ci := v.CreateInstance()
	inst, ok := ci.(*Instance)
	if !ok {
		t.Fatal("CreateInstance should return *Instance")
	}
	if inst.GetValue() != 4 {
		t.Errorf("instance value = %d, want 4", inst.GetValue())
	}
}

func TestInstance_Clamping(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propValue: 10,
		propCount: 5,
	})
	if inst.GetValue() != 5 {
		t.Errorf("value should clamp to count: got %d, want 5", inst.GetValue())
	}

	inst = NewInstance(rtui.Props{
		propValue: -1,
		propCount: 5,
	})
	if inst.GetValue() != 0 {
		t.Errorf("value should clamp to 0: got %d, want 0", inst.GetValue())
	}

	inst = NewInstance(rtui.Props{
		propCount: 0,
	})
	if inst.GetCount() != 1 {
		t.Errorf("count should clamp to 1: got %d, want 1", inst.GetCount())
	}
}

func TestInstance_HandleAction_LeftRight(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propValue: 2,
		propCount: 5,
	})

	handled := inst.HandleAction(&action.Action{Type: action.ActionCursorRight})
	if !handled {
		t.Error("CursorRight should be handled")
	}
	if inst.GetValue() != 3 {
		t.Errorf("after right: value = %d, want 3", inst.GetValue())
	}

	handled = inst.HandleAction(&action.Action{Type: action.ActionCursorLeft})
	if !handled {
		t.Error("CursorLeft should be handled")
	}
	if inst.GetValue() != 2 {
		t.Errorf("after left: value = %d, want 2", inst.GetValue())
	}
}

func TestInstance_HandleAction_HomeEnd(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propValue: 2,
		propCount: 5,
	})

	inst.HandleAction(&action.Action{Type: action.ActionCursorHome})
	if inst.GetValue() != 0 {
		t.Errorf("after Home: value = %d, want 0", inst.GetValue())
	}

	inst.HandleAction(&action.Action{Type: action.ActionCursorEnd})
	if inst.GetValue() != 5 {
		t.Errorf("after End: value = %d, want 5", inst.GetValue())
	}
}

func TestInstance_HandleAction_ToggleCycle(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propValue:      4,
		propCount:      5,
		propAllowClear: true,
	})

	inst.HandleAction(&action.Action{Type: action.ActionToggle})
	if inst.GetValue() != 5 {
		t.Errorf("toggle should advance to max: got %d, want 5", inst.GetValue())
	}

	inst.HandleAction(&action.Action{Type: action.ActionToggle})
	if inst.GetValue() != 0 {
		t.Errorf("toggle at max with allowClear should clear: got %d, want 0", inst.GetValue())
	}
}

func TestInstance_HandleAction_SelectItem(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propCount: 5,
	})

	handled := inst.HandleAction(action.NewActionWithPayload(action.ActionSelectItem, 4))
	if !handled {
		t.Fatal("ActionSelectItem with payload should be handled")
	}
	if inst.GetValue() != 4 {
		t.Errorf("value = %d, want 4", inst.GetValue())
	}
}

func TestInstance_HandleAction_Disabled(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propValue:    3,
		propDisabled: true,
	})

	handled := inst.HandleAction(&action.Action{Type: action.ActionCursorRight})
	if handled {
		t.Error("disabled rate should not handle actions")
	}
	if inst.GetValue() != 3 {
		t.Errorf("disabled rate value changed: %d, want 3", inst.GetValue())
	}
}

func TestInstance_Paint_BasicStars(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propValue: 3,
		propCount: 5,
	})

	cmds := inst.Paint(0, 0)
	if len(cmds) != 1 {
		t.Fatalf("expected 1 draw command, got %d", len(cmds))
	}
	if cmds[0].Text != "★★★☆☆" {
		t.Errorf("stars = %q, want %q", cmds[0].Text, "★★★☆☆")
	}
}

func TestInstance_Paint_WithLabelAndValue(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propLabel:     "Quality",
		propValue:     2,
		propCount:     5,
		propShowValue: true,
	})

	cmds := inst.Paint(0, 0)
	if len(cmds) != 3 {
		t.Fatalf("expected 3 draw commands, got %d", len(cmds))
	}
	if cmds[0].Text != "Quality: " {
		t.Errorf("label text = %q, want %q", cmds[0].Text, "Quality: ")
	}
	if cmds[1].Text != "★★☆☆☆" {
		t.Errorf("stars = %q, want %q", cmds[1].Text, "★★☆☆☆")
	}
	if cmds[2].Text != "2/5" {
		t.Errorf("value text = %q, want %q", cmds[2].Text, "2/5")
	}
}

func TestInstance_Measure(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propLabel:     "Rate",
		propCount:     5,
		propShowValue: true,
	})

	size := inst.Measure(layout.Constraints{
		MaxWidth:  100,
		MaxHeight: 10,
	})

	if size.Height != 1 {
		t.Errorf("height = %d, want 1", size.Height)
	}
	if size.Width < 12 {
		t.Errorf("width = %d, want >= 12", size.Width)
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
		propValue:        2,
		propCount:        5,
		propChangeIntent: intent.FieldChangeIntent{Field: "rating"},
	})
	inst.SetIntentEmitter(func(i intent.Intent) {
		emitted = i
	})

	inst.HandleAction(&action.Action{Type: action.ActionCursorRight})

	fci, ok := emitted.(intent.FieldChangeIntent)
	if !ok {
		t.Fatalf("expected FieldChangeIntent, got %T", emitted)
	}
	if fci.Field != "rating" {
		t.Errorf("field = %q, want %q", fci.Field, "rating")
	}
	if fci.Value != "3" {
		t.Errorf("value = %q, want %q", fci.Value, "3")
	}
}

func TestFluentBuilder(t *testing.T) {
	node := NewBuilder().
		Label("Satisfaction").
		Value(4).
		Count(7).
		AllowClear(false).
		Character("*").
		ShowValue(true).
		Build()

	props := node.Props()
	if props[propLabel] != "Satisfaction" {
		t.Errorf("label = %v, want Satisfaction", props[propLabel])
	}
	if props[propValue] != 4 {
		t.Errorf("value = %v, want 4", props[propValue])
	}
	if props[propCount] != 7 {
		t.Errorf("count = %v, want 7", props[propCount])
	}
	if props[propCharacter] != "*" {
		t.Errorf("character = %v, want *", props[propCharacter])
	}
}

func TestFluentBuilder_BuildInstance(t *testing.T) {
	inst := NewBuilder().
		Value(3).
		Count(5).
		BuildInstance()

	if inst.GetValue() != 3 {
		t.Errorf("instance value = %d, want 3", inst.GetValue())
	}
}

func TestRate_Aliases(t *testing.T) {
	r1 := Rate()
	if r1.Tag() != "rate" {
		t.Error("Rate() should return a rate VNode")
	}

	r2 := NewRate()
	if r2.Tag() != "rate" {
		t.Error("NewRate() should return a rate VNode")
	}
}
