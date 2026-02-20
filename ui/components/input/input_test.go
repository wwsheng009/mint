package input

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/layout"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// VNode Tests
// =============================================================================

func TestVNode_New(t *testing.T) {
	input := New()
	if input == nil {
		t.Fatal("New returned nil")
	}
	if input.Tag() != "input" {
		t.Errorf("Tag = %q, want %q", input.Tag(), "input")
	}
}

func TestVNode_Defaults(t *testing.T) {
	input := New()
	if input.Value() != "" {
		t.Error("Default value should be empty")
	}
	if input.Placeholder() != "" {
		t.Error("Default placeholder should be empty")
	}
	if input.InputType() != TypeText {
		t.Error("Default type should be TypeText")
	}
	if input.Disabled() {
		t.Error("Default disabled should be false")
	}
	if input.ReadOnly() {
		t.Error("Default readOnly should be false")
	}
}

func TestVNode_Builder(t *testing.T) {
	input := NewBuilder().
		Placeholder("Enter name").
		Value("John").
		MaxLen(20).
		Key("name").
		Build()

	vnode := input.(*VNode)
	if vnode.Placeholder() != "Enter name" {
		t.Errorf("Placeholder = %q, want %q", vnode.Placeholder(), "Enter name")
	}
	if vnode.Value() != "John" {
		t.Errorf("Value = %q, want %q", vnode.Value(), "John")
	}
	if vnode.MaxLen() != 20 {
		t.Errorf("MaxLen = %d, want 20", vnode.MaxLen())
	}
	if vnode.Key() != "name" {
		t.Errorf("Key = %q, want %q", vnode.Key(), "name")
	}
}

func TestVNode_Password(t *testing.T) {
	input := New().SetPassword()
	if input.InputType() != TypePassword {
		t.Error("Type should be TypePassword")
	}
}

func TestVNode_CreateInstance(t *testing.T) {
	input := New().SetValue("test").SetPlaceholder("placeholder")
	inst := input.CreateInstance()

	if inst == nil {
		t.Fatal("CreateInstance returned nil")
	}

	ci, ok := inst.(*Instance)
	if !ok {
		t.Fatal("Instance is not *Instance")
	}
	if ci.GetValue() != "test" {
		t.Errorf("Value = %q, want %q", ci.GetValue(), "test")
	}
}

// =============================================================================
// Instance Tests
// =============================================================================

func TestInstance_New(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"placeholder": "Enter text",
		"value":       "hello",
	})

	if inst.placeholder != "Enter text" {
		t.Errorf("Placeholder = %q, want %q", inst.placeholder, "Enter text")
	}
	if inst.GetValue() != "hello" {
		t.Errorf("Value = %q, want %q", inst.GetValue(), "hello")
	}
}

func TestInstance_Measure(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		width     int
		wantMin   int
	}{
		{"Empty value", "", 0, 12}, // 10 min content + 2 bracket padding
		{"Short value", "hi", 0, 12},
		{"Long value", "hello world", 0, 13}, // 11 content + 2 padding
		{"With explicit width", "hi", 20, 22}, // 20 width + 2 padding
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := NewInstance(rtui.Props{
				"value": tt.value,
				"width": tt.width,
			})

			size := inst.Measure(layout.UnboundedConstraints())

			if size.Width < tt.wantMin {
				t.Errorf("Width = %d, want >= %d", size.Width, tt.wantMin)
			}
			// Height is 3 (1 content + 2 border lines) for default BorderSingle
			if size.Height != 3 {
				t.Errorf("Height = %d, want 3", size.Height)
			}
		})
	}
}

func TestInstance_InsertText(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"value": "hello",
	})

	// Cursor at end
	inst.SetCursorPos(5)
	inst.InsertText(" world")

	if inst.GetValue() != "hello world" {
		t.Errorf("Value = %q, want %q", inst.GetValue(), "hello world")
	}
	if inst.CursorPos() != 11 {
		t.Errorf("CursorPos = %d, want 11", inst.CursorPos())
	}
}

func TestInstance_InsertText_AtBeginning(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"value": "world",
	})
	inst.SetCursorPos(0)
	inst.InsertText("hello ")
	if inst.GetValue() != "hello world" {
		t.Errorf("Value = %q, want %q", inst.GetValue(), "hello world")
	}
	if inst.CursorPos() != 6 {
		t.Errorf("CursorPos = %d, want 6", inst.CursorPos())
	}
}

func TestInstance_InsertText_MaxLen(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"value":  "abc",
		"maxLen": 5,
	})

	inst.SetCursorPos(3)

	// Should succeed
	if !inst.InsertText("de") {
		t.Error("InsertText should succeed")
	}
	if inst.GetValue() != "abcde" {
		t.Errorf("Value = %q, want %q", inst.GetValue(), "abcde")
	}

	// Should fail (exceeds max)
	inst.SetCursorPos(5)
	if inst.InsertText("f") {
		t.Error("InsertText should fail (max length)")
	}
	if inst.GetValue() != "abcde" {
		t.Errorf("Value should not change")
	}
}

func TestInstance_DeleteText(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"value": "hello",
	})

	// Delete from end (backspace)
	inst.SetCursorPos(5)
	if !inst.DeleteText(-1) {
		t.Error("DeleteText should succeed")
	}
	if inst.GetValue() != "hell" {
		t.Errorf("Value = %q, want %q", inst.GetValue(), "hell")
	}
	if inst.CursorPos() != 4 {
		t.Errorf("CursorPos = %d, want 4", inst.CursorPos())
	}

	// Delete at cursor (delete key)
	inst.SetCursorPos(2)
	if !inst.DeleteText(1) {
		t.Error("DeleteText should succeed")
	}
	if inst.GetValue() != "hel" {
		t.Errorf("Value = %q, want %q", inst.GetValue(), "hel")
	}
}

func TestInstance_Disabled(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"value":    "test",
		"disabled": true,
	})

	if !inst.IsDisabled() {
		t.Error("Should be disabled")
	}

	// Disabled instance should not handle actions
	if inst.CanHandleAction("input") {
		t.Error("Disabled instance should not handle input action")
	}

	// Insert should fail
	inst.state.Disabled = false // Temporarily enable
	inst.InsertText("x")
	inst.state.Disabled = true
	if inst.InsertText("y") {
		t.Error("InsertText should fail when disabled")
	}
}

func TestInstance_ReadOnly(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"value":   "test",
		"readOnly": true,
	})

	// Active = readOnly in our implementation
	if !inst.state.Active {
		t.Error("Should be read-only (Active)")
	}

	// ReadOnly instance should not handle editing actions
	if inst.CanHandleAction("input") {
		t.Error("ReadOnly instance should not handle input action")
	}
}

func TestInstance_HandleAction(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"value": "test",
	})

	// Input action
	handled := inst.HandleAction("input", "x")
	if !handled {
		t.Error("Input action should be handled")
	}
	if inst.GetValue() != "testx" {
		t.Errorf("Value = %q, want %q", inst.GetValue(), "testx")
	}

	// Backspace action
	handled = inst.HandleAction("backspace", nil)
	if !handled {
		t.Error("Backspace action should be handled")
	}
	if inst.GetValue() != "test" {
		t.Errorf("Value = %q, want %q", inst.GetValue(), "test")
	}

	// Unknown action
	handled = inst.HandleAction("unknown", nil)
	if handled {
		t.Error("Unknown action should not be handled")
	}
}

func TestInstance_Focus(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"value": "test",
	})

	if inst.HasFocus() {
		t.Error("Should not have focus initially")
	}

	inst.SetFocus(true)
	if !inst.HasFocus() {
		t.Error("Should have focus after SetFocus(true)")
	}

	inst.SetFocus(false)
	if inst.HasFocus() {
		t.Error("Should not have focus after SetFocus(false)")
	}
}

func TestInstance_Paint(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		placeholder string
		inputType Type
		want     string
	}{
		{"Empty with placeholder", "", "Enter text", TypeText, "Enter text"},
		{"With value", "hello", "Enter text", TypeText, "hello     "},
		{"Password type", "secret", "Password", TypePassword, "******    "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := NewInstance(rtui.Props{
				"value":       tt.value,
				"placeholder": tt.placeholder,
				"inputType":   tt.inputType,
			})

			cmds := inst.Paint(0, 0)
			// Paint returns 5 commands: 4 border commands + 1 text command (for BorderSingle default)
			if len(cmds) != 5 {
				t.Fatalf("Paint returned %d commands, want 5", len(cmds))
			}

			// Text is in the last command (index 4)
			if cmds[4].Text != tt.want {
				t.Errorf("Text = %q, want %q", cmds[4].Text, tt.want)
			}
		})
	}
}

func TestInstance_Paint_BorderNone(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"value":       "test",
		"borderStyle": layout.BorderNone,
	})

	cmds := inst.Paint(0, 0)
	// BorderNone uses brackets: [text] - returns 3 commands
	if len(cmds) != 3 {
		t.Fatalf("Paint returned %d commands, want 3", len(cmds))
	}

	// First command should be "["
	if cmds[0].Text != "[" {
		t.Errorf("First cmd = %q, want [", cmds[0].Text)
	}
	// Second command should be the text
	if cmds[1].Text != "test      " {
		t.Errorf("Text = %q, want %q", cmds[1].Text, "test      ")
	}
	// Third command should be "]"
	if cmds[2].Text != "]" {
		t.Errorf("Third cmd = %q, want ]", cmds[2].Text)
	}
}

func TestInstance_Paint_WithWidth(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"value": "hi",
		"width": 10,
	})

	cmds := inst.Paint(0, 0)
	// Paint returns 5 commands: 4 border commands + 1 text command
	if len(cmds) != 5 {
		t.Fatalf("Paint returned %d commands, want 5", len(cmds))
	}

	// Text is in the last command, padded to content width (10)
	expected := "hi        "
	if cmds[4].Text != expected {
		t.Errorf("Text = %q, want %q", cmds[4].Text, expected)
	}
}

// =============================================================================
// FocusableVNode Tests
// =============================================================================

func TestVNode_IsFocusable(t *testing.T) {
	// Enabled input is focusable
	input := New()
	if !input.IsFocusable() {
		t.Error("Enabled input should be focusable")
	}

	// Disabled input is not focusable
	input = New().SetDisabled(true)
	if input.IsFocusable() {
		t.Error("Disabled input should not be focusable")
	}

	// ReadOnly input is not focusable
	input = New().SetReadOnly(true)
	if input.IsFocusable() {
		t.Error("ReadOnly input should not be focusable")
	}
}

func TestVNode_GetFocusID(t *testing.T) {
	// With key
	input := New()
	input.SetKey("mykey")
	if input.GetFocusID() != "input:mykey" {
		t.Errorf("FocusID = %q, want %q", input.GetFocusID(), "input:mykey")
	}

	// Without key, uses placeholder
	input = New().SetPlaceholder("Name")
	if input.GetFocusID() != "input:Name" {
		t.Errorf("FocusID = %q, want %q", input.GetFocusID(), "input:Name")
	}
}
