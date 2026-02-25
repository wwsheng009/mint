package textarea

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/layout"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// VNode Tests
// =============================================================================

func TestVNode_New(t *testing.T) {
	ta := New()
	if ta == nil {
		t.Fatal("New returned nil")
	}
	if ta.Tag() != "textarea" {
		t.Errorf("Tag = %q, want %q", ta.Tag(), "textarea")
	}
	if ta.Rows() != 3 {
		t.Errorf("Default rows = %d, want 3", ta.Rows())
	}
	if ta.Cols() != 40 {
		t.Errorf("Default cols = %d, want 40", ta.Cols())
	}
}

func TestVNode_Builder(t *testing.T) {
	ta := NewBuilder().
		Placeholder("Enter text").
		Value("initial").
		Rows(5).
		Cols(30).
		Key("mytextarea").
		Build()

	vnode := ta.(*VNode)
	if vnode.Placeholder() != "Enter text" {
		t.Errorf("Placeholder = %q, want %q", vnode.Placeholder(), "Enter text")
	}
	if vnode.Value() != "initial" {
		t.Errorf("Value = %q, want %q", vnode.Value(), "initial")
	}
	if vnode.Rows() != 5 {
		t.Errorf("Rows = %d, want 5", vnode.Rows())
	}
	if vnode.Cols() != 30 {
		t.Errorf("Cols = %d, want 30", vnode.Cols())
	}
}

func TestVNode_CreateInstance(t *testing.T) {
	ta := New().SetValue("test value").SetRows(4)
	inst := ta.CreateInstance()

	if inst == nil {
		t.Fatal("CreateInstance returned nil")
	}

	ci, ok := inst.(*Instance)
	if !ok {
		t.Fatal("Instance is not *Instance")
	}
	if ci.GetValue() != "test value" {
		t.Errorf("Value = %q, want %q", ci.GetValue(), "test value")
	}
}

// =============================================================================
// Instance Tests
// =============================================================================

func TestInstance_New(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"value": "hello\nworld",
		"rows":  5,
		"cols":  20,
	})

	if inst.GetValue() != "hello\nworld" {
		t.Errorf("Value = %q, want %q", inst.GetValue(), "hello\nworld")
	}
	if inst.rows != 5 {
		t.Errorf("Rows = %d, want 5", inst.rows)
	}
}

func TestInstance_Measure(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"rows": 3,
		"cols": 20,
	})

	size := inst.Measure(layout.UnboundedConstraints())

	// Width = cols + 4, Height = rows + 2
	if size.Width != 24 {
		t.Errorf("Width = %d, want 24", size.Width)
	}
	if size.Height != 5 {
		t.Errorf("Height = %d, want 5", size.Height)
	}
}

func TestInstance_InsertText(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"value": "hello",
	})

	inst.InsertText(" world")
	if inst.GetValue() != "hello world" {
		t.Errorf("Value = %q, want %q", inst.GetValue(), "hello world")
	}

	// Insert newline
	inst.InsertText("\nnew line")
	if inst.GetValue() != "hello world\nnew line" {
		t.Errorf("Value = %q, want %q", inst.GetValue(), "hello world\nnew line")
	}
}

func TestInstance_InsertText_MaxLen(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"value":  "abc",
		"maxLen": 5,
	})

	if !inst.InsertText("de") {
		t.Error("InsertText should succeed")
	}
	if inst.GetValue() != "abcde" {
		t.Errorf("Value = %q, want %q", inst.GetValue(), "abcde")
	}

	// Should fail (exceeds max)
	if inst.InsertText("f") {
		t.Error("InsertText should fail (max length)")
	}
}

func TestInstance_SetValue(t *testing.T) {
	inst := NewInstance(rtui.Props{})

	inst.SetValue("new value")
	if inst.GetValue() != "new value" {
		t.Errorf("Value = %q, want %q", inst.GetValue(), "new value")
	}

	// Setting same value should not mark dirty
	inst.ClearDirty()
	inst.SetValue("new value")
	if inst.IsDirty() {
		t.Error("Should not be dirty when setting same value")
	}
}

func TestInstance_Disabled(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"disabled": true,
	})

	if !inst.IsDisabled() {
		t.Error("Should be disabled")
	}

	if inst.HandleAction(action.NewAction(action.ActionInputText)) {
		t.Error("Disabled instance should not handle input action")
	}

	// Insert should fail
	if inst.InsertText("x") {
		t.Error("InsertText should fail when disabled")
	}
}

func TestInstance_Focus(t *testing.T) {
	inst := NewInstance(rtui.Props{})

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
	inst := NewInstance(rtui.Props{
		"value": "line1\nline2",
		"rows":  3,
		"cols":  10,
	})

	cmds := inst.Paint(0, 0)

	// Should have: top border + 2 content lines + 1 empty line + bottom border = 5
	if len(cmds) != 5 {
		t.Errorf("Paint returned %d commands, want 5", len(cmds))
	}

	// Check top border
	if cmds[0].Text[0:1] != "+" {
		t.Errorf("Top border should start with '+', got %q", cmds[0].Text[0:1])
	}

	// Check bottom border
	if cmds[4].Text[0:1] != "+" {
		t.Errorf("Bottom border should start with '+', got %q", cmds[4].Text[0:1])
	}
}
