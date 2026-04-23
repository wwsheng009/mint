package radio

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/optiongroup"
)

func TestVNode_New(t *testing.T) {
	r := New("Choice A")
	if r == nil {
		t.Fatal("New returned nil")
	}
	if r.Label() != "Choice A" {
		t.Errorf("Label = %q, want %q", r.Label(), "Choice A")
	}
	if r.Tag() != "radio" {
		t.Errorf("Tag = %q, want %q", r.Tag(), "radio")
	}
}

func TestVNode_Builder(t *testing.T) {
	r := NewBuilder().
		Label("Newsletter").
		Checked(true).
		Disabled(false).
		Key("newsletter").
		Build()

	vnode := r.(*VNode)
	if vnode.Label() != "Newsletter" {
		t.Errorf("Label = %q, want %q", vnode.Label(), "Newsletter")
	}
	if !vnode.Checked() {
		t.Error("Checked should be true")
	}
	if vnode.Key() != "newsletter" {
		t.Errorf("Key = %q, want %q", vnode.Key(), "newsletter")
	}
}

func TestInstance_SelectOnlyChecksOnce(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propLabel: "Choice A",
	})

	var emitted []intent.Intent
	inst.SetIntentEmitter(func(i intent.Intent) {
		emitted = append(emitted, i)
	})
	inst.selectIntent = intent.BindField("choice")

	if !inst.Select() {
		t.Fatal("Select should return true")
	}
	if !inst.IsChecked() {
		t.Fatal("radio should be checked after Select")
	}

	inst.Select()
	if len(emitted) != 1 {
		t.Fatalf("emitted intents = %d, want 1", len(emitted))
	}

	fieldChange, ok := emitted[0].(intent.FieldChangeIntent)
	if !ok {
		t.Fatalf("emitted intent = %T, want FieldChangeIntent", emitted[0])
	}
	if fieldChange.Value != "true" {
		t.Errorf("FieldChangeIntent.Value = %q, want %q", fieldChange.Value, "true")
	}
}

func TestInstance_HandleAction(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propLabel: "Choice A",
	})

	handled := inst.HandleAction(action.NewAction(action.ActionToggle))
	if !handled {
		t.Fatal("toggle action should be handled")
	}
	if !inst.IsChecked() {
		t.Fatal("radio should be checked after toggle")
	}
}

func TestInstance_MeasureAndPaint(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propLabel:   "Choice A",
		propChecked: true,
	})

	size := inst.Measure(layout.UnboundedConstraints())
	if size.Width != 12 {
		t.Errorf("Width = %d, want 12", size.Width)
	}
	if size.Height != 1 {
		t.Errorf("Height = %d, want 1", size.Height)
	}

	cmds := inst.Paint(0, 0)
	if len(cmds) != 1 {
		t.Fatalf("Paint returned %d commands, want 1", len(cmds))
	}
	if cmds[0].Text != "(*) Choice A" {
		t.Errorf("Text = %q, want %q", cmds[0].Text, "(*) Choice A")
	}
}

func TestGroupBuilder_WrapsOptionGroupSingleMode(t *testing.T) {
	group := NewGroupBuilder([]Option{
		{Value: "a", Label: "A"},
		{Value: "b", Label: "B"},
	}).
		Label("Pick one").
		Selected("b").
		Horizontal().
		Spacing(2).
		BuildTyped()

	if group.Tag() != "radiogroup" {
		t.Errorf("Tag = %q, want %q", group.Tag(), "radiogroup")
	}
	if group.Mode() != optiongroup.ModeSingle {
		t.Errorf("Mode = %v, want %v", group.Mode(), optiongroup.ModeSingle)
	}
	if group.Selected() != "b" {
		t.Errorf("Selected = %q, want %q", group.Selected(), "b")
	}
	if group.Orientation() != OrientationHorizontal {
		t.Errorf("Orientation = %v, want %v", group.Orientation(), OrientationHorizontal)
	}
	if len(group.Options()) != 2 {
		t.Errorf("Options len = %d, want 2", len(group.Options()))
	}
}

func TestGroupInstance_SelectOption(t *testing.T) {
	inst := NewGroupBuilder([]Option{
		{Value: "a", Label: "A"},
		{Value: "b", Label: "B"},
	}).BuildInstance()

	inst.SelectOption("b")
	if inst.GetSelected() != "b" {
		t.Errorf("Selected = %q, want %q", inst.GetSelected(), "b")
	}
}
