package tag

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
	v := New("hello")
	if v == nil {
		t.Fatal("New returned nil")
	}
	if v.Tag() != "tag" {
		t.Errorf("Tag = %q, want \"tag\"", v.Tag())
	}
	if v.Color() != ColorDefault {
		t.Errorf("Default Color = %d, want ColorDefault", v.Color())
	}
	if v.Closable() {
		t.Error("Default Closable should be false")
	}
	if v.Text() != "hello" {
		t.Errorf("Text = %q, want \"hello\"", v.Text())
	}
}

func TestVNode_Setters(t *testing.T) {
	v := New("tag").
		SetText("updated").
		SetColor(ColorSuccess).
		SetClosable(true).
		SetIcon("★").
		SetCloseIntent("close")

	if v.Text() != "updated" {
		t.Errorf("Text = %q, want \"updated\"", v.Text())
	}
	if v.Color() != ColorSuccess {
		t.Errorf("Color = %d, want ColorSuccess", v.Color())
	}
	if !v.Closable() {
		t.Error("Closable should be true")
	}
	if v.Icon() != "★" {
		t.Errorf("Icon = %q, want \"★\"", v.Icon())
	}
	if v.CloseIntent() != "close" {
		t.Errorf("CloseIntent = %v, want \"close\"", v.CloseIntent())
	}
}

func TestVNode_ColorHelpers(t *testing.T) {
	tests := []struct {
		name      string
		setFn     func(*VNode) *VNode
		wantColor TagColor
	}{
		{"Default", func(v *VNode) *VNode { return v.Default() }, ColorDefault},
		{"Primary", func(v *VNode) *VNode { return v.Primary() }, ColorPrimary},
		{"Success", func(v *VNode) *VNode { return v.Success() }, ColorSuccess},
		{"Warning", func(v *VNode) *VNode { return v.Warning() }, ColorWarning},
		{"Error", func(v *VNode) *VNode { return v.Error() }, ColorError},
		{"Processing", func(v *VNode) *VNode { return v.Processing() }, ColorProcessing},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := tt.setFn(New("test"))
			if v.Color() != tt.wantColor {
				t.Errorf("Color = %d, want %d", v.Color(), tt.wantColor)
			}
		})
	}
}

func TestVNode_Props(t *testing.T) {
	v := New("hello").
		SetColor(ColorPrimary).
		SetClosable(true).
		SetIcon("★")

	props := v.Props()
	if props[propText] != "hello" {
		t.Errorf("Props text = %v, want \"hello\"", props[propText])
	}
	if props[propColor] != ColorPrimary {
		t.Errorf("Props color = %v, want ColorPrimary", props[propColor])
	}
	if props[propClosable] != true {
		t.Errorf("Props closable = %v, want true", props[propClosable])
	}
	if props[propIcon] != "★" {
		t.Errorf("Props icon = %v, want \"★\"", props[propIcon])
	}
}

func TestVNode_SetProps(t *testing.T) {
	v := New("original")
	v.SetProps(rtui.Props{
		propText:     "updated",
		propColor:    ColorError,
		propClosable: true,
	})

	if v.Text() != "updated" {
		t.Errorf("Text = %q, want \"updated\"", v.Text())
	}
	if v.Color() != ColorError {
		t.Errorf("Color = %d, want ColorError", v.Color())
	}
	if !v.Closable() {
		t.Error("Closable should be true")
	}
}

func TestVNode_CreateInstance(t *testing.T) {
	v := New("test").Primary().SetClosable(true)
	inst := v.CreateInstance()
	if inst == nil {
		t.Fatal("CreateInstance returned nil")
	}
	ti, ok := inst.(*Instance)
	if !ok {
		t.Fatal("Instance is not *Instance")
	}
	if ti.text != "test" {
		t.Errorf("text = %q, want \"test\"", ti.text)
	}
	if ti.color != ColorPrimary {
		t.Errorf("color = %d, want ColorPrimary", ti.color)
	}
	if !ti.closable {
		t.Error("closable should be true")
	}
}

func TestVNode_ImplementsVNode(t *testing.T) {
	var _ rtui.VNode = New("test")
}

// =============================================================================
// Instance Tests
// =============================================================================

func TestInstance_New(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propText:     "hello",
		propColor:    ColorPrimary,
		propClosable: true,
		propIcon:     "★",
	})
	if inst.text != "hello" {
		t.Errorf("text = %q, want \"hello\"", inst.text)
	}
	if inst.color != ColorPrimary {
		t.Errorf("color = %d, want ColorPrimary", inst.color)
	}
	if !inst.closable {
		t.Error("closable should be true")
	}
	if inst.icon != "★" {
		t.Errorf("icon = %q, want \"★\"", inst.icon)
	}
	if !inst.IsDirty() {
		t.Error("new instance should be dirty")
	}
}

func TestInstance_Measure_Simple(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propText: "hello",
	})
	size := inst.Measure(layout.Constraints{MaxWidth: 80, MaxHeight: 10})
	// " hello " = 7 chars
	if size.Width != 7 {
		t.Errorf("Width = %d, want 7", size.Width)
	}
	if size.Height != 1 {
		t.Errorf("Height = %d, want 1", size.Height)
	}
}

func TestInstance_Measure_WithIcon(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propText: "hello",
		propIcon: "★",
	})
	size := inst.Measure(layout.Constraints{MaxWidth: 80, MaxHeight: 10})
	// " ★ hello " = 1 + 1 + 1 + 5 + 1 = 9 chars... wait
	// icon="★" (1 char) + " " (1 char) + text="hello" (5 chars) + padding=2 = 9
	if size.Width != 9 {
		t.Errorf("Width = %d, want 9", size.Width)
	}
	if size.Height != 1 {
		t.Errorf("Height = %d, want 1", size.Height)
	}
}

func TestInstance_Measure_Closable(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propText:     "hello",
		propClosable: true,
	})
	size := inst.Measure(layout.Constraints{MaxWidth: 80, MaxHeight: 10})
	// text(5) + closable(2) + padding(2) = 9
	if size.Width != 9 {
		t.Errorf("Width = %d, want 9", size.Width)
	}
}

func TestInstance_Measure_MaxWidth(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propText: "a very long tag text",
	})
	size := inst.Measure(layout.Constraints{MaxWidth: 10, MaxHeight: 10})
	if size.Width != 10 {
		t.Errorf("Width = %d, want 10 (clamped)", size.Width)
	}
}

func TestInstance_Paint_Default(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propText: "hello",
	})
	inst.SetBounds(0, 0, 10, 1)
	cmds := inst.Paint(0, 0)
	if len(cmds) != 1 {
		t.Fatalf("Paint returned %d commands, want 1", len(cmds))
	}
	if !strings.Contains(cmds[0].Text, "hello") {
		t.Errorf("Paint text %q should contain \"hello\"", cmds[0].Text)
	}
}

func TestInstance_Paint_WithIcon(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propText: "tag",
		propIcon: "★",
	})
	inst.SetBounds(0, 0, 10, 1)
	cmds := inst.Paint(0, 0)
	if len(cmds) != 1 {
		t.Fatalf("Paint returned %d commands, want 1", len(cmds))
	}
	if !strings.Contains(cmds[0].Text, "★") {
		t.Errorf("Paint text %q should contain \"★\"", cmds[0].Text)
	}
	if !strings.Contains(cmds[0].Text, "tag") {
		t.Errorf("Paint text %q should contain \"tag\"", cmds[0].Text)
	}
}

func TestInstance_Paint_Closable(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propText:     "tag",
		propClosable: true,
	})
	inst.SetBounds(0, 0, 10, 1)
	cmds := inst.Paint(0, 0)
	if len(cmds) != 1 {
		t.Fatalf("Paint returned %d commands, want 1", len(cmds))
	}
	if !strings.Contains(cmds[0].Text, "×") {
		t.Errorf("Paint text %q should contain \"×\"", cmds[0].Text)
	}
}

func TestInstance_SetProps(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propText: "original",
	})
	inst.MarkClean()

	inst.SetProps(rtui.Props{
		propText:  "updated",
		propColor: ColorSuccess,
	})

	if inst.text != "updated" {
		t.Errorf("text = %q, want \"updated\"", inst.text)
	}
	if inst.color != ColorSuccess {
		t.Errorf("color = %d, want ColorSuccess", inst.color)
	}
	if !inst.IsDirty() {
		t.Error("SetProps should mark instance dirty")
	}
}

func TestInstance_ImplementsComponentInstance(t *testing.T) {
	var _ rtui.ComponentInstance = NewInstance(rtui.Props{})
}

// =============================================================================
// Builder Tests
// =============================================================================

func TestBuilder_Basic(t *testing.T) {
	v := NewBuilder("hello").Build()
	if v.text != "hello" {
		t.Errorf("text = %q, want \"hello\"", v.text)
	}
}

func TestBuilder_Fluent(t *testing.T) {
	v := NewBuilder("tag").
		Key("t1").
		Text("updated").
		Primary().
		Closable(true).
		Icon("★").
		CloseIntent("dismiss").
		Build()

	if v.key != "t1" {
		t.Errorf("key = %q, want \"t1\"", v.key)
	}
	if v.text != "updated" {
		t.Errorf("text = %q, want \"updated\"", v.text)
	}
	if v.color != ColorPrimary {
		t.Errorf("color = %d, want ColorPrimary", v.color)
	}
	if !v.closable {
		t.Error("closable should be true")
	}
	if v.icon != "★" {
		t.Errorf("icon = %q, want \"★\"", v.icon)
	}
	if v.closeIntent != "dismiss" {
		t.Errorf("closeIntent = %v, want \"dismiss\"", v.closeIntent)
	}
}
