package switchcomp

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

func TestVNode_New(t *testing.T) {
	sw := New("Airplane mode")
	if sw == nil {
		t.Fatal("New returned nil")
	}
	if sw.Label() != "Airplane mode" {
		t.Errorf("Label = %q, want %q", sw.Label(), "Airplane mode")
	}
	if sw.Tag() != "switch" {
		t.Errorf("Tag = %q, want %q", sw.Tag(), "switch")
	}
	if sw.CheckedLabel() != defaultCheckedLabel {
		t.Errorf("CheckedLabel = %q, want %q", sw.CheckedLabel(), defaultCheckedLabel)
	}
	if sw.UncheckedLabel() != defaultUncheckedLabel {
		t.Errorf("UncheckedLabel = %q, want %q", sw.UncheckedLabel(), defaultUncheckedLabel)
	}
}

func TestVNode_Builder(t *testing.T) {
	sw := NewBuilder().
		Label("Wi-Fi").
		Checked(true).
		Labels("YES", "NO").
		Key("wifi").
		Build()

	vnode := sw.(*VNode)
	if vnode.Label() != "Wi-Fi" {
		t.Errorf("Label = %q, want %q", vnode.Label(), "Wi-Fi")
	}
	if !vnode.Checked() {
		t.Error("Checked should be true")
	}
	if vnode.CheckedLabel() != "YES" {
		t.Errorf("CheckedLabel = %q, want %q", vnode.CheckedLabel(), "YES")
	}
	if vnode.UncheckedLabel() != "NO" {
		t.Errorf("UncheckedLabel = %q, want %q", vnode.UncheckedLabel(), "NO")
	}
	if vnode.Key() != "wifi" {
		t.Errorf("Key = %q, want %q", vnode.Key(), "wifi")
	}
}

func TestVNode_CreateInstance(t *testing.T) {
	sw := New("Bluetooth").SetChecked(true).SetLabels("YES", "NO")
	inst := sw.CreateInstance()

	if inst == nil {
		t.Fatal("CreateInstance returned nil")
	}

	si, ok := inst.(*Instance)
	if !ok {
		t.Fatal("Instance is not *Instance")
	}
	if si.Label() != "Bluetooth" {
		t.Errorf("Label = %q, want %q", si.Label(), "Bluetooth")
	}
	if !si.IsChecked() {
		t.Error("Checked should be true")
	}
	if si.CheckedLabel() != "YES" {
		t.Errorf("CheckedLabel = %q, want %q", si.CheckedLabel(), "YES")
	}
}

func TestInstance_ToggleEmitsFieldChangeIntent(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propLabel:        "Airplane mode",
		propToggleIntent: intent.BindField("airplane"),
	})

	var emitted []intent.Intent
	inst.SetIntentEmitter(func(i intent.Intent) {
		emitted = append(emitted, i)
	})

	if !inst.Toggle() {
		t.Fatal("Toggle should return true")
	}
	if !inst.IsChecked() {
		t.Fatal("switch should be checked after first toggle")
	}

	if inst.Toggle() {
		t.Fatal("Toggle should return false after second toggle")
	}
	if inst.IsChecked() {
		t.Fatal("switch should be unchecked after second toggle")
	}

	if len(emitted) != 2 {
		t.Fatalf("emitted intents = %d, want 2", len(emitted))
	}

	first, ok := emitted[0].(intent.FieldChangeIntent)
	if !ok {
		t.Fatalf("first emitted intent = %T, want FieldChangeIntent", emitted[0])
	}
	if first.Field != "airplane" || first.Value != "true" {
		t.Errorf("first intent = %+v, want Field=airplane Value=true", first)
	}

	second, ok := emitted[1].(intent.FieldChangeIntent)
	if !ok {
		t.Fatalf("second emitted intent = %T, want FieldChangeIntent", emitted[1])
	}
	if second.Value != "false" {
		t.Errorf("second intent Value = %q, want %q", second.Value, "false")
	}
}

func TestInstance_HandleAction(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propLabel: "Wi-Fi",
	})

	handled := inst.HandleAction(action.NewAction(action.ActionToggle))
	if !handled {
		t.Fatal("toggle action should be handled")
	}
	if !inst.IsChecked() {
		t.Fatal("switch should be checked after toggle")
	}

	handled = inst.HandleAction(action.NewAction(action.ActionClick))
	if !handled {
		t.Fatal("click action should be handled")
	}
	if inst.IsChecked() {
		t.Fatal("switch should be unchecked after click")
	}
}

func TestInstance_Disabled(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propDisabled: true,
		propLabel:    "Power",
	})

	if !inst.IsDisabled() {
		t.Fatal("switch should be disabled")
	}
	if inst.HandleAction(action.NewAction(action.ActionToggle)) {
		t.Fatal("disabled switch should not handle toggle")
	}
}

func TestInstance_SetChecked(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propLabel: "Power",
	})

	inst.SetChecked(true)
	if !inst.IsChecked() {
		t.Fatal("switch should be checked")
	}

	inst.ClearDirty()
	inst.SetChecked(true)
	if inst.IsDirty() {
		t.Fatal("setting the same checked value should not mark dirty")
	}
}

func TestInstance_MeasureAndPaint(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propLabel:          "Bluetooth",
		propChecked:        true,
		propCheckedLabel:   "YES",
		propUncheckedLabel: "NO",
	})

	size := inst.Measure(layout.UnboundedConstraints())
	if size.Width != 17 {
		t.Errorf("Width = %d, want %d", size.Width, 17)
	}
	if size.Height != 1 {
		t.Errorf("Height = %d, want %d", size.Height, 1)
	}

	cmds := inst.Paint(0, 0)
	if len(cmds) != 2 {
		t.Fatalf("Paint returned %d commands, want 2", len(cmds))
	}
	if cmds[0].Text != "[YES  ]" {
		t.Errorf("track = %q, want %q", cmds[0].Text, "[YES  ]")
	}
	if cmds[1].Text != "Bluetooth" {
		t.Errorf("label = %q, want %q", cmds[1].Text, "Bluetooth")
	}
	if cmds[1].X != 8 {
		t.Errorf("label X = %d, want %d", cmds[1].X, 8)
	}
}

func TestInstance_PaintUncheckedWithoutLabel(t *testing.T) {
	inst := NewInstance(rtui.Props{})

	cmds := inst.Paint(0, 0)
	if len(cmds) != 1 {
		t.Fatalf("Paint returned %d commands, want 1", len(cmds))
	}
	if cmds[0].Text != "[  OFF]" {
		t.Errorf("track = %q, want %q", cmds[0].Text, "[  OFF]")
	}
}

func TestInstance_Focus(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propLabel: "Sync",
	})

	if inst.HasFocus() {
		t.Fatal("switch should not have focus initially")
	}

	inst.SetFocus(true)
	if !inst.HasFocus() {
		t.Fatal("switch should have focus after SetFocus(true)")
	}

	inst.SetFocus(false)
	if inst.HasFocus() {
		t.Fatal("switch should not have focus after SetFocus(false)")
	}
}
