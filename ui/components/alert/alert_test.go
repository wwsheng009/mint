package alert

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
	v := New()
	if v == nil {
		t.Fatal("New returned nil")
	}
	if v.Tag() != "alert" {
		t.Errorf("Tag = %q, want \"alert\"", v.Tag())
	}
	if v.AlertType() != AlertInfo {
		t.Errorf("Default AlertType = %d, want AlertInfo", v.AlertType())
	}
	if v.Closable() {
		t.Error("Default Closable should be false")
	}
}

func TestVNode_Setters(t *testing.T) {
	v := New().
		SetTitle("Warning").
		SetMessage("Disk full").
		SetAlertType(AlertWarning).
		SetClosable(true).
		SetCloseIntent("close")

	if v.Title() != "Warning" {
		t.Errorf("Title = %q, want \"Warning\"", v.Title())
	}
	if v.Message() != "Disk full" {
		t.Errorf("Message = %q, want \"Disk full\"", v.Message())
	}
	if v.AlertType() != AlertWarning {
		t.Errorf("AlertType = %d, want AlertWarning", v.AlertType())
	}
	if !v.Closable() {
		t.Error("Closable should be true")
	}
	if v.CloseIntent() != "close" {
		t.Errorf("CloseIntent = %v, want \"close\"", v.CloseIntent())
	}
}

func TestVNode_TypeHelpers(t *testing.T) {
	tests := []struct {
		name      string
		setFn     func(*VNode) *VNode
		wantType  AlertType
	}{
		{"Info", func(v *VNode) *VNode { return v.Info() }, AlertInfo},
		{"Success", func(v *VNode) *VNode { return v.Success() }, AlertSuccess},
		{"Warning", func(v *VNode) *VNode { return v.Warning() }, AlertWarning},
		{"Error", func(v *VNode) *VNode { return v.Error() }, AlertError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := tt.setFn(New())
			if v.AlertType() != tt.wantType {
				t.Errorf("AlertType = %d, want %d", v.AlertType(), tt.wantType)
			}
		})
	}
}

func TestVNode_Props(t *testing.T) {
	v := New().
		SetTitle("T").
		SetMessage("M").
		SetClosable(true)

	props := v.Props()
	if props[propTitle] != "T" {
		t.Errorf("Props title = %v, want \"T\"", props[propTitle])
	}
	if props[propMessage] != "M" {
		t.Errorf("Props message = %v, want \"M\"", props[propMessage])
	}
	if props[propClosable] != true {
		t.Errorf("Props closable = %v, want true", props[propClosable])
	}
}

func TestVNode_CreateInstance(t *testing.T) {
	v := New().SetTitle("Test").SetMessage("Hello").Success()
	inst := v.CreateInstance()
	if inst == nil {
		t.Fatal("CreateInstance returned nil")
	}
	ai, ok := inst.(*Instance)
	if !ok {
		t.Fatal("Instance is not *Instance")
	}
	if ai.title != "Test" {
		t.Errorf("title = %q, want \"Test\"", ai.title)
	}
	if ai.alertType != AlertSuccess {
		t.Errorf("alertType = %d, want AlertSuccess", ai.alertType)
	}
}

func TestVNode_ImplementsVNode(t *testing.T) {
	var _ rtui.VNode = New()
}

// =============================================================================
// Instance Tests
// =============================================================================

func TestInstance_New(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle:   "Alert Title",
		propMessage: "Something happened",
		propAlertType: AlertError,
		propClosable: true,
	})
	if inst.title != "Alert Title" {
		t.Errorf("title = %q, want \"Alert Title\"", inst.title)
	}
	if inst.message != "Something happened" {
		t.Errorf("message = %q, want \"Something happened\"", inst.message)
	}
	if inst.alertType != AlertError {
		t.Errorf("alertType = %d, want AlertError", inst.alertType)
	}
	if !inst.closable {
		t.Error("closable should be true")
	}
	if !inst.IsDirty() {
		t.Error("new instance should be dirty")
	}
}

func TestInstance_Measure_TitleAndMessage(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle:   "Title",
		propMessage: "Message",
	})
	size := inst.Measure(layout.Constraints{MaxWidth: 60, MaxHeight: 20})
	// title + message = 2 rows
	if size.Height != 2 {
		t.Errorf("Height = %d, want 2", size.Height)
	}
	if size.Width != 60 {
		t.Errorf("Width = %d, want 60", size.Width)
	}
}

func TestInstance_Measure_MessageOnly(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propMessage: "Just a message",
	})
	size := inst.Measure(layout.Constraints{MaxWidth: 50, MaxHeight: 10})
	if size.Height != 1 {
		t.Errorf("Height = %d, want 1", size.Height)
	}
}

func TestInstance_Measure_Closable(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle:    "T",
		propMessage:  "M",
		propClosable: true,
	})
	size := inst.Measure(layout.Constraints{MaxWidth: 40, MaxHeight: 10})
	// title + message + close hint = 3 rows
	if size.Height != 3 {
		t.Errorf("Height = %d, want 3", size.Height)
	}
}

func TestInstance_Paint_Info(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propMessage: "Info message",
	})
	inst.SetBounds(0, 0, 40, 2)
	cmds := inst.Paint(0, 0)
	if len(cmds) != 1 {
		t.Fatalf("Paint returned %d commands, want 1", len(cmds))
	}
	if !strings.Contains(cmds[0].Text, "Info message") {
		t.Errorf("Paint text %q should contain \"Info message\"", cmds[0].Text)
	}
}

func TestInstance_Paint_WithTitle(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle:   "Error!",
		propMessage: "Something went wrong",
		propAlertType: AlertError,
	})
	inst.SetBounds(0, 0, 40, 3)
	cmds := inst.Paint(0, 0)
	// title + message = 2 commands
	if len(cmds) != 2 {
		t.Fatalf("Paint returned %d commands, want 2", len(cmds))
	}
	if !strings.Contains(cmds[0].Text, "Error!") {
		t.Errorf("First cmd %q should contain \"Error!\"", cmds[0].Text)
	}
	if !strings.Contains(cmds[1].Text, "Something went wrong") {
		t.Errorf("Second cmd %q should contain \"Something went wrong\"", cmds[1].Text)
	}
}

func TestInstance_Paint_Closable(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propMessage:  "Check this",
		propClosable: true,
	})
	inst.SetBounds(0, 0, 40, 3)
	cmds := inst.Paint(0, 0)
	// message + close hint = 2 commands
	if len(cmds) != 2 {
		t.Fatalf("Paint returned %d commands, want 2", len(cmds))
	}
	if !strings.Contains(cmds[1].Text, "close") {
		t.Errorf("Close hint %q should contain \"close\"", cmds[1].Text)
	}
}

func TestInstance_SetProps(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propMessage: "original",
	})
	inst.MarkClean()

	inst.SetProps(rtui.Props{
		propMessage:   "updated",
		propAlertType: AlertSuccess,
	})

	if inst.message != "updated" {
		t.Errorf("message = %q, want \"updated\"", inst.message)
	}
	if inst.alertType != AlertSuccess {
		t.Errorf("alertType = %d, want AlertSuccess", inst.alertType)
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
	v := NewBuilder("Something happened").Build()
	if v.message != "Something happened" {
		t.Errorf("message = %q, want \"Something happened\"", v.message)
	}
}

func TestBuilder_Fluent(t *testing.T) {
	v := NewBuilder("msg").
		Key("a1").
		Title("My Alert").
		Warning().
		Closable(true).
		CloseIntent("dismiss").
		Build()

	if v.key != "a1" {
		t.Errorf("key = %q, want \"a1\"", v.key)
	}
	if v.title != "My Alert" {
		t.Errorf("title = %q, want \"My Alert\"", v.title)
	}
	if v.alertType != AlertWarning {
		t.Errorf("alertType = %d, want AlertWarning", v.alertType)
	}
	if !v.closable {
		t.Error("closable should be true")
	}
	if v.closeIntent != "dismiss" {
		t.Errorf("closeIntent = %v, want \"dismiss\"", v.closeIntent)
	}
}
