package button

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/control"
)

type buttonSliceIntent struct {
	Path []int
}

func (buttonSliceIntent) IntentType() string { return "button.slice" }

// =============================================================================
// VNode Tests
// =============================================================================

func TestVNode_New(t *testing.T) {
	b := New("Click Me")
	if b.Variant() != VariantDefault {
		t.Errorf("Variant() = %v, want Default", b.Variant())
	}
	if b.Size() != SizeMedium {
		t.Errorf("Size() = %v, want Medium", b.Size())
	}
	if b.FocusStyle() != FocusStyleReverse {
		t.Errorf("FocusStyle() = %v, want Reverse", b.FocusStyle())
	}
}

func TestVNode_FluentAPI(t *testing.T) {
	pressIntent := intent.OpenModal("test")
	b := New("Test")
	b.SetKey("test-btn")
	b.SetVariant(VariantPrimary)
	b.SetSize(SizeLarge)
	b.SetFocusStyle(FocusStyleUnderline)
	b.SetDisabled(true)
	b.SetIntent(pressIntent)

	if b.Key() != "test-btn" {
		t.Errorf("Key() = %q, want %q", b.Key(), "test-btn")
	}
	if b.Variant() != VariantPrimary {
		t.Errorf("Variant() = %v, want Primary", b.Variant())
	}
	if b.Size() != SizeLarge {
		t.Errorf("Size() = %v, want Large", b.Size())
	}
	if b.FocusStyle() != FocusStyleUnderline {
		t.Errorf("FocusStyle() = %v, want Underline", b.FocusStyle())
	}
	if !b.Disabled() {
		t.Error("Disabled() = false, want true")
	}
	if b.PressIntent() == nil {
		t.Error("PressIntent() = nil, want intent")
	}
}
func TestVNode_CreateInstance(t *testing.T) {
	b := New("Test")
	b.SetVariant(VariantPrimary)
	b.SetDisabled(true)

	inst := b.CreateInstance()
	if inst == nil {
		t.Fatal("CreateInstance() returned nil")
	}

	// Verify instance implements interfaces
	if _, ok := inst.(rtui.PaintableInstance); !ok {
		t.Error("Instance should implement PaintableInstance")
	}
	if _, ok := inst.(rtui.FocusableInstance); !ok {
		t.Error("Instance should implement FocusableInstance")
	}
}

// =============================================================================
// Instance Tests
// =============================================================================

func TestInstance_New(t *testing.T) {
	props := rtui.Props{
		"key":      "test",
		"label":    "Click Me",
		"variant":  VariantPrimary,
		"size":     SizeLarge,
		"disabled": true,
	}

	inst := NewInstance(props)

	if inst.Key() != "test" {
		t.Errorf("Key() = %q, want %q", inst.Key(), "test")
	}
	if inst.state.Disabled != true {
		t.Error("Disabled state should be true")
	}
}

func TestInstance_SetFocus(t *testing.T) {
	inst := NewInstance(rtui.Props{"label": "Test"})

	inst.SetFocus(true)
	if !inst.HasFocus() {
		t.Error("HasFocus() = false after SetFocus(true)")
	}
	if !inst.state.Focused {
		t.Error("state.Focused should be true")
	}
	if !inst.IsDirty() {
		t.Error("SetFocus should mark instance dirty")
	}

	inst.dirty = false
	inst.SetFocus(false)
	if inst.HasFocus() {
		t.Error("HasFocus() = true after SetFocus(false)")
	}
}

func TestInstance_Disabled(t *testing.T) {
	inst := NewInstance(rtui.Props{"label": "Test"})

	if inst.IsDisabled() {
		t.Error("IsDisabled() = true, want false")
	}

	inst.state.Disabled = true
	if !inst.IsDisabled() {
		t.Error("IsDisabled() = false after setting disabled")
	}
}

func TestInstance_HandleAction(t *testing.T) {
	inst := NewInstance(rtui.Props{"label": "Test"})

	// Focus action
	if !inst.HandleAction(action.NewActionWithPayload(action.ActionFocus, nil)) {
		t.Error("HandleAction(Focus) should return true")
	}
	if !inst.HasFocus() {
		t.Error("HandleAction(Focus) should set focus state")
	}

	// Blur action
	inst.dirty = false
	if !inst.HandleAction(action.NewActionWithPayload(action.ActionBlur, nil)) {
		t.Error("HandleAction(Blur) should return true")
	}
	if inst.HasFocus() {
		t.Error("HandleAction(Blur) should clear focus state")
	}
}

func TestInstance_HandleAction_Disabled(t *testing.T) {
	inst := NewInstance(rtui.Props{"label": "Test", "disabled": true})

	if inst.HandleAction(action.NewActionWithPayload(action.ActionFocus, nil)) {
		t.Error("HandleAction(Focus) on disabled should return false")
	}
}

func TestInstance_Paint(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"label":   "Test",
		"variant": VariantDefault,
		"size":    SizeMedium,
	})

	cmds := inst.Paint(0, 0)
	if len(cmds) != 1 {
		t.Fatalf("Paint() returned %d commands, want 1", len(cmds))
	}

	// Check the button text contains the label
	cmd := cmds[0]
	if cmd.X != 0 || cmd.Y != 0 {
		t.Errorf("Paint() position = (%d, %d), want (0, 0)", cmd.X, cmd.Y)
	}
}

func TestInstance_Paint_Focused(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"label":      "Test",
		"focusStyle": FocusStyleReverse,
	})

	inst.SetFocus(true)
	cmds := inst.Paint(0, 0)

	if len(cmds) != 1 {
		t.Fatalf("Paint() returned %d commands, want 1", len(cmds))
	}

	// Focused button should have * prefix
	cmd := cmds[0]
	if len(cmd.Text) == 0 || cmd.Text[0] != '*' {
		t.Errorf("Focused button text should start with '*', got %q", cmd.Text)
	}
}

func TestInstance_OnMountUnmount(t *testing.T) {
	inst := NewInstance(rtui.Props{"label": "Test"})

	// Should not panic
	inst.OnMount()
	inst.OnUnmount()
}

func TestInstance_SetProps(t *testing.T) {
	inst := NewInstance(rtui.Props{"label": "Original"})

	changed := inst.SetProps(rtui.Props{"label": "Updated"})
	if !changed {
		t.Error("SetProps should return true when props change")
	}
	if inst.label != "Updated" {
		t.Errorf("label = %q, want %q", inst.label, "Updated")
	}

	// No change
	changed = inst.SetProps(rtui.Props{"label": "Updated"})
	if changed {
		t.Error("SetProps should return false when props don't change")
	}
}

func TestInstance_SetPropsAllowsNonComparableIntent(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"label":       "Open",
		"pressIntent": buttonSliceIntent{Path: []int{1, 2}},
	})

	changed := inst.SetProps(rtui.Props{
		"label":       "Open",
		"pressIntent": buttonSliceIntent{Path: []int{1, 2}},
	})
	if changed {
		t.Fatal("SetProps should treat equal non-comparable intents as unchanged")
	}

	changed = inst.SetProps(rtui.Props{
		"label":       "Open",
		"pressIntent": buttonSliceIntent{Path: []int{1, 3}},
	})
	if !changed {
		t.Fatal("SetProps should detect non-comparable intent changes")
	}
}

// =============================================================================
// Builder Tests
// =============================================================================

func TestBuilder_FluentAPI(t *testing.T) {
	pressIntent := intent.OpenModal("settings")
	vnode := NewBuilder("Save").
		Key("save-btn").
		Primary().
		Large().
		FocusStyle(FocusStyleUnderline).
		Disabled(true).
		OnPress(pressIntent).
		PaddingAll(1).
		Build()

	b, ok := vnode.(*VNode)
	if !ok {
		t.Fatal("Build() should return *VNode")
	}

	if b.Key() != "save-btn" {
		t.Errorf("Key() = %q, want %q", b.Key(), "save-btn")
	}
	if b.Variant() != VariantPrimary {
		t.Errorf("Variant() = %v, want Primary", b.Variant())
	}
	if b.Size() != SizeLarge {
		t.Errorf("Size() = %v, want Large", b.Size())
	}
	if !b.Disabled() {
		t.Error("Disabled() = false, want true")
	}
}

func TestBuilder_ConvenienceFunctions(t *testing.T) {
	// Test B() function
	b1 := B("Test").Build()
	if b1 == nil {
		t.Error("B().Build() returned nil")
	}

	// Test Button() function
	b2 := Button("Test")
	if b2 == nil {
		t.Error("Button() returned nil")
	}

	// Test BuildInstance()
	inst := B("Test").BuildInstance()
	if inst == nil {
		t.Error("BuildInstance() returned nil")
	}
}

// =============================================================================
// Behavior Integration Tests
// =============================================================================

func TestInstance_BehaviorComposition(t *testing.T) {
	pressIntent := intent.Click("test-btn")
	inst := NewInstance(rtui.Props{
		"label":       "Test",
		"pressIntent": pressIntent,
	})

	// Should have all behaviors
	if inst.behaviors == nil {
		t.Fatal("behaviors is nil")
	}
	if len(inst.behaviors.List()) != 4 {
		t.Errorf("Should have 4 behaviors, got %d", len(inst.behaviors.List()))
	}

	// Test focusable behavior
	inst.HandleAction(action.NewActionWithPayload(action.ActionFocus, nil))
	if !inst.HasFocus() {
		t.Error("Focus action should set focus state")
	}

	// Test pressable behavior
	inst.HandleAction(action.NewActionWithPayload(action.ActionPress, nil))
	if !inst.state.Pressed {
		t.Error("Press action should set pressed state")
	}
}

func TestInstance_StateTransitions(t *testing.T) {
	inst := NewInstance(rtui.Props{"label": "Test"})

	// Initial state should be idle
	if !inst.state.IsIdle() {
		t.Error("Initial state should be idle")
	}

	// Focus
	inst.HandleAction(action.NewActionWithPayload(action.ActionFocus, nil))
	if inst.state.Focused != true {
		t.Error("State should have Focused=true")
	}

	// Hover
	inst.HandleAction(action.NewActionWithPayload(action.ActionMouseEnter, nil))
	if inst.state.Hovered != true {
		t.Error("State should have Hovered=true")
	}

	// Press
	inst.HandleAction(action.NewActionWithPayload(action.ActionPress, nil))
	if inst.state.Pressed != true {
		t.Error("State should have Pressed=true")
	}
}

// =============================================================================
// control.Instance Interface Tests
// =============================================================================

func TestInstance_ControlInterface(t *testing.T) {
	inst := NewInstance(rtui.Props{"label": "Test"})

	// Test control.Instance interface methods
	var _ control.Instance = inst

	// GetState/SetState
	state := inst.GetState()
	if state == nil {
		t.Fatal("GetState() returned nil")
	}

	newState := control.InteractionState{Focused: true}
	inst.SetState(newState)
	if !inst.state.Focused {
		t.Error("SetState should update state")
	}

	// GetBounds/SetBounds
	inst.SetBounds(10, 20, 100, 30)
	x, y, w, h := inst.GetBounds()
	if x != 10 || y != 20 || w != 100 || h != 30 {
		t.Errorf("GetBounds() = (%d, %d, %d, %d), want (10, 20, 100, 30)", x, y, w, h)
	}

	// GetProp/SetProp
	inst.SetProp("disabled", true)
	v, ok := inst.GetProp("disabled")
	if !ok || v != true {
		t.Errorf("GetProp(disabled) = (%v, %v), want (true, true)", v, ok)
	}
}

// =============================================================================
// Chinese Character Width Tests
// =============================================================================

func TestInstance_Measure_ChineseWidth(t *testing.T) {
	tests := []struct {
		name      string
		label     string
		wantWidth int
	}{
		{"ASCII label", "OK", 7},         // 2 + 3 + 2 (SizeMedium padding)
		{"Single Chinese char", "确定", 9}, // 4 + 3 + 2
		{"Mixed content", "保存Save", 13},  // 8 + 3 + 2
		{"Chinese only", "提交更改", 13},     // 8 + 3 + 2
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := NewInstance(rtui.Props{"label": tt.label})
			size := inst.Measure(layout.UnboundedConstraints())

			// Calculate expected width: display width of label + 3 (brackets + focus indicator) + 2 (SizeMedium padding)
			labelDisplayWidth := paint.StringWidth(tt.label)
			expectedWidth := labelDisplayWidth + 5 // 3 for brackets/focus + 2 for SizeMedium

			t.Logf("Label: %q, Display width: %d, Expected total: %d, Got: %d",
				tt.label, labelDisplayWidth, expectedWidth, size.Width)

			if size.Width != tt.wantWidth {
				t.Errorf("Width = %d, want %d (label: %q)", size.Width, tt.wantWidth, tt.label)
			}
		})
	}
}

func TestInstance_Paint_ChineseWidth(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"label": "确定",
		"size":  SizeMedium,
	})

	cmds := inst.Paint(0, 0)
	if len(cmds) == 0 {
		t.Fatal("Paint() returned no commands")
	}

	t.Logf("Paint output: %q", cmds[0].Text)

	// Verify the button text contains the Chinese label
	if !containsChinese(cmds[0].Text, "确定") {
		t.Errorf("Button text should contain '确定', got %q", cmds[0].Text)
	}
}

func containsChinese(s, substr string) bool {
	// Simple check if the string contains the expected Chinese characters
	for _, r := range substr {
		found := false
		for _, sr := range s {
			if sr == r {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
