package drawer

import (
	"strings"
	"testing"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	"github.com/wwsheng009/mint/runtime/paint"
	runtimeplatform "github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	newtext "github.com/wwsheng009/mint/ui/components/text"
)

type testCloseIntent struct{}

func (testCloseIntent) IntentType() string { return "drawer.test.close" }

// =============================================================================
// VNode Tests
// =============================================================================

func TestVNode_Defaults(t *testing.T) {
	v := New()

	if v.Tag() != "drawer" {
		t.Errorf("Tag() = %q, want %q", v.Tag(), "drawer")
	}
	if v.IsOpen() {
		t.Error("IsOpen() should be false by default")
	}
	if v.Placement() != PlacementRight {
		t.Errorf("Placement() = %v, want PlacementRight", v.Placement())
	}
	if v.Width() != 30 {
		t.Errorf("Width() = %d, want 30", v.Width())
	}
	if v.Height() != 15 {
		t.Errorf("Height() = %d, want 15", v.Height())
	}
	if !v.Closeable() {
		t.Error("Closeable() should be true by default")
	}
	if !v.CloseOnEsc() {
		t.Error("CloseOnEsc() should be true by default")
	}
	if !v.CloseOnBackdrop() {
		t.Error("CloseOnBackdrop() should be true by default")
	}
	if v.BorderStyle() != "single" {
		t.Errorf("BorderStyle() = %q, want %q", v.BorderStyle(), "single")
	}
	if !v.Shadow() {
		t.Error("Shadow() should be true by default")
	}
}

func TestVNode_FluentSetters(t *testing.T) {
	v := New().
		SetTitle("Settings").
		SetOpen(true).
		SetWidth(40).
		SetHeight(20).
		SetPadding(1).
		SetPlacement(PlacementLeft).
		SetBorderStyle("double").
		SetCloseable(false).
		SetCloseOnEsc(false).
		SetCloseOnBackdrop(false)

	if v.Title() != "Settings" {
		t.Errorf("Title() = %q, want %q", v.Title(), "Settings")
	}
	if !v.IsOpen() {
		t.Error("IsOpen() should be true")
	}
	if v.Width() != 40 {
		t.Errorf("Width() = %d, want 40", v.Width())
	}
	if v.Height() != 20 {
		t.Errorf("Height() = %d, want 20", v.Height())
	}
	if v.Padding() != 1 {
		t.Errorf("Padding() = %d, want 1", v.Padding())
	}
	if v.Placement() != PlacementLeft {
		t.Errorf("Placement() = %v, want PlacementLeft", v.Placement())
	}
	if v.BorderStyle() != "double" {
		t.Errorf("BorderStyle() = %q, want %q", v.BorderStyle(), "double")
	}
	if v.Closeable() {
		t.Error("Closeable() should be false")
	}
	if v.CloseOnEsc() {
		t.Error("CloseOnEsc() should be false")
	}
	if v.CloseOnBackdrop() {
		t.Error("CloseOnBackdrop() should be false")
	}
}

func TestVNode_PlacementConvenience(t *testing.T) {
	tests := []struct {
		name      string
		fn        func(*VNode) *VNode
		placement Placement
	}{
		{"Right", func(v *VNode) *VNode { return v.Right() }, PlacementRight},
		{"Left", func(v *VNode) *VNode { return v.Left() }, PlacementLeft},
		{"Top", func(v *VNode) *VNode { return v.Top() }, PlacementTop},
		{"Bottom", func(v *VNode) *VNode { return v.Bottom() }, PlacementBottom},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := tt.fn(New())
			if v.Placement() != tt.placement {
				t.Errorf("Placement() = %v, want %v", v.Placement(), tt.placement)
			}
		})
	}
}

func TestVNode_OpenToggle(t *testing.T) {
	v := New()
	if v.IsOpen() {
		t.Error("should be closed initially")
	}
	v.Open()
	if !v.IsOpen() {
		t.Error("should be open after Open()")
	}
	v.Close()
	if v.IsOpen() {
		t.Error("should be closed after Close()")
	}
	v.Toggle()
	if !v.IsOpen() {
		t.Error("should be open after Toggle()")
	}
	v.Toggle()
	if v.IsOpen() {
		t.Error("should be closed after second Toggle()")
	}
}

func TestVNode_BorderConvenience(t *testing.T) {
	tests := []struct {
		name  string
		fn    func(*VNode) *VNode
		style string
	}{
		{"Single", func(v *VNode) *VNode { return v.Single() }, "single"},
		{"Double", func(v *VNode) *VNode { return v.Double() }, "double"},
		{"Rounded", func(v *VNode) *VNode { return v.Rounded() }, "rounded"},
		{"Dashed", func(v *VNode) *VNode { return v.Dashed() }, "dashed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := tt.fn(New())
			if v.BorderStyle() != tt.style {
				t.Errorf("BorderStyle() = %q, want %q", v.BorderStyle(), tt.style)
			}
		})
	}
}

func TestVNode_Props_RoundTrip(t *testing.T) {
	original := New().
		SetTitle("Title").
		SetOpen(true).
		SetPlacement(PlacementBottom).
		SetWidth(50).
		SetHeight(10).
		SetPadding(2).
		SetBorderStyle("rounded").
		SetShadow(false)

	props := original.Props()

	restored := New()
	restored.SetProps(props)

	if restored.Title() != original.Title() {
		t.Errorf("Title mismatch: got %q, want %q", restored.Title(), original.Title())
	}
	if restored.IsOpen() != original.IsOpen() {
		t.Errorf("IsOpen mismatch: got %v, want %v", restored.IsOpen(), original.IsOpen())
	}
	if restored.Placement() != original.Placement() {
		t.Errorf("Placement mismatch: got %v, want %v", restored.Placement(), original.Placement())
	}
	if restored.Width() != original.Width() {
		t.Errorf("Width mismatch: got %d, want %d", restored.Width(), original.Width())
	}
	if restored.Height() != original.Height() {
		t.Errorf("Height mismatch: got %d, want %d", restored.Height(), original.Height())
	}
	if restored.BorderStyle() != original.BorderStyle() {
		t.Errorf("BorderStyle mismatch: got %q, want %q", restored.BorderStyle(), original.BorderStyle())
	}
	if restored.Shadow() != original.Shadow() {
		t.Errorf("Shadow mismatch: got %v, want %v", restored.Shadow(), original.Shadow())
	}
}

func TestVNode_Children(t *testing.T) {
	content := newtext.New("body content")
	footer := newtext.New("footer content")

	v := New().SetContent(content).SetFooter(footer)

	children := v.Children()
	if len(children) != 2 {
		t.Fatalf("Children() len = %d, want 2", len(children))
	}
}

func TestVNode_Children_ContentOnly(t *testing.T) {
	content := newtext.New("body")
	v := New().SetContent(content)

	children := v.Children()
	if len(children) != 1 {
		t.Fatalf("Children() len = %d, want 1", len(children))
	}
}

func TestVNode_Type(t *testing.T) {
	v := New()
	if v.Type() != rtui.VNodeElement {
		t.Errorf("Type() = %v, want VNodeElement", v.Type())
	}
}

func TestVNode_Layer(t *testing.T) {
	v := New()
	if v.GetLayer() != rtui.LayerModal {
		t.Errorf("GetLayer() = %v, want LayerModal", v.GetLayer())
	}
}

func TestVNode_InstanceFactory(t *testing.T) {
	v := New().SetTitle("Test").SetOpen(true).SetWidth(35)
	inst := v.CreateInstance()
	if inst == nil {
		t.Fatal("CreateInstance() returned nil")
	}
}

// =============================================================================
// Builder Tests
// =============================================================================

func TestBuilder_Basic(t *testing.T) {
	vnode := NewBuilder().
		Title("My Drawer").
		Open(true).
		Width(30).
		Right().
		Build()

	if vnode == nil {
		t.Fatal("Build() returned nil")
	}
	if vnode.Tag() != "drawer" {
		t.Errorf("Tag() = %q, want %q", vnode.Tag(), "drawer")
	}
}

func TestBuilder_BuildVNode(t *testing.T) {
	v := NewBuilder().
		Title("Settings").
		Width(40).
		Left().
		Rounded().
		BuildVNode()

	if v.Title() != "Settings" {
		t.Errorf("Title = %q, want %q", v.Title(), "Settings")
	}
	if v.Width() != 40 {
		t.Errorf("Width = %d, want 40", v.Width())
	}
	if v.Placement() != PlacementLeft {
		t.Errorf("Placement = %v, want PlacementLeft", v.Placement())
	}
	if v.BorderStyle() != "rounded" {
		t.Errorf("BorderStyle = %q, want %q", v.BorderStyle(), "rounded")
	}
}

func TestBuilder_BuildInstance(t *testing.T) {
	inst := NewBuilder().
		Title("Hello").
		Open(true).
		Width(30).
		BuildInstance()

	if inst == nil {
		t.Fatal("BuildInstance() returned nil")
	}
	if !inst.isOpen {
		t.Error("instance should be open")
	}
}

func TestBuilder_Convenience(t *testing.T) {
	tests := []struct {
		name   string
		vnode  rtui.VNode
		expect func(t *testing.T, v *VNode)
	}{
		{
			name:  "Opened",
			vnode: NewBuilder().Opened().Build(),
			expect: func(t *testing.T, v *VNode) {
				if !v.IsOpen() {
					t.Error("Opened() should set isOpen=true")
				}
			},
		},
		{
			name:  "Closed",
			vnode: NewBuilder().Opened().Closed().Build(),
			expect: func(t *testing.T, v *VNode) {
				if v.IsOpen() {
					t.Error("Closed() should set isOpen=false")
				}
			},
		},
		{
			name:  "NoShadow",
			vnode: NewBuilder().NoShadow().Build(),
			expect: func(t *testing.T, v *VNode) {
				if v.Shadow() {
					t.Error("NoShadow() should set showShadow=false")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, ok := tt.vnode.(*VNode)
			if !ok {
				t.Fatal("Build() should return *VNode")
			}
			tt.expect(t, v)
		})
	}
}

func TestBuilder_Style(t *testing.T) {
	s := style.NewStyle().Foreground("#ff0000").Background("#0000ff")
	v := NewBuilder().Style(s).BuildVNode()
	if v.Style().FG != "#ff0000" {
		t.Errorf("FG = %q, want %q", v.Style().FG, "#ff0000")
	}
}

func TestBuilder_OnClose(t *testing.T) {
	closeIntent := testCloseIntent{}
	v := NewBuilder().OnClose(closeIntent).BuildVNode()
	_ = v
	// Intent stored in vnode — verify it passes through to the instance
	inst := NewBuilder().OnClose(closeIntent).Open(true).BuildInstance()
	if inst.closeIntent == nil {
		t.Error("closeIntent should be set on instance")
	}
	if _, ok := inst.closeIntent.(testCloseIntent); !ok {
		t.Errorf("closeIntent type = %T, want testCloseIntent", inst.closeIntent)
	}
}

// =============================================================================
// Convenience Function Tests
// =============================================================================

func TestOf(t *testing.T) {
	content := newtext.New("hello")
	vnode := Of(content)
	if vnode == nil {
		t.Fatal("Of() returned nil")
	}
	if vnode.Tag() != "drawer" {
		t.Errorf("Tag() = %q, want %q", vnode.Tag(), "drawer")
	}
}

func TestTitled(t *testing.T) {
	vnode := Titled("My Title", newtext.New("content"))
	v := vnode.(*VNode)
	if v.Title() != "My Title" {
		t.Errorf("Title = %q, want %q", v.Title(), "My Title")
	}
}

func TestFromPlacement(t *testing.T) {
	tests := []struct {
		name      string
		fn        func() rtui.VNode
		placement Placement
	}{
		{"FromRight", func() rtui.VNode { return FromRight("title", newtext.New("c")) }, PlacementRight},
		{"FromLeft", func() rtui.VNode { return FromLeft("title", newtext.New("c")) }, PlacementLeft},
		{"FromTop", func() rtui.VNode { return FromTop("title", newtext.New("c")) }, PlacementTop},
		{"FromBottom", func() rtui.VNode { return FromBottom("title", newtext.New("c")) }, PlacementBottom},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vnode := tt.fn()
			v := vnode.(*VNode)
			if v.Placement() != tt.placement {
				t.Errorf("Placement = %v, want %v", v.Placement(), tt.placement)
			}
		})
	}
}

// =============================================================================
// Instance Tests
// =============================================================================

func TestInstance_SetProps_OpenClose(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propIsOpen: false,
		propWidth:  30,
		propHeight: 20,
	})

	if inst.isOpen {
		t.Error("should start closed")
	}
	if inst.registered {
		t.Error("should not be registered when closed")
	}

	// Open it
	changed := inst.SetProps(rtui.Props{propIsOpen: true})
	if !changed {
		t.Error("SetProps should return true when state changes")
	}
	if !inst.isOpen {
		t.Error("should be open now")
	}
	if !inst.registered {
		t.Error("should be registered when open")
	}

	// Close it
	changed = inst.SetProps(rtui.Props{propIsOpen: false})
	if !changed {
		t.Error("SetProps should return true when state changes")
	}
	if inst.isOpen {
		t.Error("should be closed now")
	}
	if inst.registered {
		t.Error("should be unregistered when closed")
	}
}

func TestInstance_SetProps_NoChange(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle:  "Hello",
		propIsOpen: false,
	})

	// Set same props — no visible state change
	changed := inst.SetProps(rtui.Props{
		propIsOpen: false,
	})
	if changed {
		t.Error("SetProps should return false when nothing significant changes")
	}
}

func TestInstance_GetProps(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle:     "Drawer Title",
		propIsOpen:    true,
		propWidth:     35,
		propPlacement: int(PlacementLeft),
	})

	props := inst.GetProps()
	if props[propTitle] != "Drawer Title" {
		t.Errorf("GetProps title = %v, want %q", props[propTitle], "Drawer Title")
	}
	if props[propIsOpen] != true {
		t.Errorf("GetProps isOpen = %v, want true", props[propIsOpen])
	}
	if props[propWidth] != 35 {
		t.Errorf("GetProps width = %v, want 35", props[propWidth])
	}
}

func TestInstance_Cleanup(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propIsOpen: true,
		propWidth:  30,
	})
	inst.registered = true
	globalRegistry.register(inst)

	inst.cleanup()

	if inst.isOpen {
		t.Error("isOpen should be false after cleanup")
	}
	if inst.registered {
		t.Error("registered should be false after cleanup")
	}
}

// =============================================================================
// Instance Paint Tests
// =============================================================================

func TestInstance_Paint_Closed(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propIsOpen: false,
		propWidth:  30,
		propHeight: 20,
	})

	cmds := inst.Paint(0, 0)
	if len(cmds) != 0 {
		t.Errorf("Paint() on closed drawer should return nil, got %d cmds", len(cmds))
	}
}

func TestInstance_Paint_Open_HasBorder(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propIsOpen:  true,
		propWidth:   20,
		propHeight:  10,
	})

	cmds := inst.Paint(0, 0)
	if len(cmds) == 0 {
		t.Fatal("Paint() on open drawer should return commands")
	}
}

func TestInstance_Paint_Title(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propIsOpen: true,
		propTitle:  "Settings",
		propWidth:  20,
		propHeight: 10,
	})

	cmds := inst.Paint(0, 0)
	found := false
	for _, cmd := range cmds {
		if strings.Contains(cmd.Text, "Settings") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Paint() should include title text")
	}
}

func TestInstance_Paint_Sets_Bounds(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propIsOpen: true,
		propWidth:  25,
		propHeight: 12,
	})

	inst.Paint(5, 3)

	x, y, w, h := inst.GetBounds()
	if x != 5 || y != 3 {
		t.Errorf("bounds position = (%d,%d), want (5,3)", x, y)
	}
	if w != 25 || h != 12 {
		t.Errorf("bounds size = (%d,%d), want (25,12)", w, h)
	}
}

func TestInstance_Paint_NoShadow(t *testing.T) {
	withShadow := NewInstance(rtui.Props{
		propIsOpen:     true,
		propWidth:      20,
		propHeight:     8,
		propShowShadow: true,
	})
	withoutShadow := NewInstance(rtui.Props{
		propIsOpen:     true,
		propWidth:      20,
		propHeight:     8,
		propShowShadow: false,
	})

	cmdsWithShadow := withShadow.Paint(0, 0)
	cmdsWithoutShadow := withoutShadow.Paint(0, 0)

	// Shadow adds extra commands
	if len(cmdsWithShadow) <= len(cmdsWithoutShadow) {
		t.Error("shadow should add more draw commands")
	}
}

// =============================================================================
// Instance Action Handling Tests
// =============================================================================

func TestInstance_HandleAction_Closed(t *testing.T) {
	inst := NewInstance(rtui.Props{propIsOpen: false})
	act := &action.Action{Type: action.ActionCancel}
	if inst.HandleAction(act) {
		t.Error("HandleAction should return false when drawer is closed")
	}
}

func TestInstance_HandleAction_Cancel_Closes(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propIsOpen:     true,
		propCloseable:  true,
		propCloseOnEsc: true,
	})
	act := &action.Action{Type: action.ActionCancel}
	result := inst.HandleAction(act)
	if !result {
		t.Error("HandleAction should return true for Cancel when open")
	}
	if inst.isOpen {
		t.Error("drawer should be closed after Cancel action")
	}
}

func TestInstance_HandleAction_Cancel_NotCloseable(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propIsOpen:    true,
		propCloseable: false,
	})
	act := &action.Action{Type: action.ActionCancel}
	inst.HandleAction(act)
	if !inst.isOpen {
		t.Error("drawer should remain open when closeable=false")
	}
}

func TestInstance_HandleKeyMessage_Escape(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propIsOpen:     true,
		propCloseable:  true,
		propCloseOnEsc: true,
	})
	keyMsg := &runtimemsg.KeyMsg{Special: runtimeplatform.KeyEscape}
	result := inst.HandleKeyMessage(keyMsg)
	if !result {
		t.Error("HandleKeyMessage should return true for ESC when open")
	}
	if inst.isOpen {
		t.Error("drawer should be closed after ESC")
	}
}

func TestInstance_HandleKeyMessage_Escape_Disabled(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propIsOpen:     true,
		propCloseable:  true,
		propCloseOnEsc: false,
	})
	keyMsg := &runtimemsg.KeyMsg{Special: runtimeplatform.KeyEscape}
	inst.HandleKeyMessage(keyMsg)
	if !inst.isOpen {
		t.Error("drawer should remain open when closeOnEsc=false")
	}
}

func TestInstance_HandleKeyMessage_Closed(t *testing.T) {
	inst := NewInstance(rtui.Props{propIsOpen: false})
	keyMsg := &runtimemsg.KeyMsg{Special: runtimeplatform.KeyEscape}
	if inst.HandleKeyMessage(keyMsg) {
		t.Error("HandleKeyMessage should return false when drawer is closed")
	}
}

func TestInstance_HandleMouseMessage_OutsideClick(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propIsOpen:          true,
		propWidth:           20,
		propHeight:          10,
		propCloseable:       true,
		propCloseOnBackdrop: true,
	})

	// Set bounds so we know what's "inside"
	inst.bounds = [4]int{50, 5, 20, 10}

	// Click outside
	mouseMsg := &runtimemsg.MouseMsg{
		Action: runtimemsg.MouseActionPress,
		X:      0,
		Y:      0,
	}
	result := inst.HandleMouseMessage(mouseMsg)
	if !result {
		t.Error("HandleMouseMessage should return true for outside click")
	}
	if inst.isOpen {
		t.Error("drawer should close on outside click")
	}
}

func TestInstance_HandleMouseMessage_InsideClick(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propIsOpen:          true,
		propWidth:           20,
		propHeight:          10,
		propCloseable:       true,
		propCloseOnBackdrop: true,
	})

	inst.bounds = [4]int{50, 5, 20, 10}

	// Click inside
	mouseMsg := &runtimemsg.MouseMsg{
		Action: runtimemsg.MouseActionPress,
		X:      55,
		Y:      8,
	}
	inst.HandleMouseMessage(mouseMsg)
	if !inst.isOpen {
		t.Error("drawer should remain open on inside click")
	}
}

func TestInstance_HandleMouseMessage_OutsideClick_Disabled(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propIsOpen:          true,
		propWidth:           20,
		propHeight:          10,
		propCloseable:       true,
		propCloseOnBackdrop: false,
	})
	inst.bounds = [4]int{50, 5, 20, 10}

	mouseMsg := &runtimemsg.MouseMsg{
		Action: runtimemsg.MouseActionPress,
		X:      0,
		Y:      0,
	}
	inst.HandleMouseMessage(mouseMsg)
	if !inst.isOpen {
		t.Error("drawer should remain open when closeOnBackdrop=false")
	}
}

// =============================================================================
// CloseIntent Tests
// =============================================================================

func TestInstance_RequestClose_EmitsIntent(t *testing.T) {
	var emitted intent.Intent
	closeIntent := testCloseIntent{}

	inst := NewInstance(rtui.Props{
		propIsOpen:      true,
		propCloseable:   true,
		propCloseOnEsc:  true,
		propCloseIntent: closeIntent,
	})
	inst.intentEmitter = func(i intent.Intent) {
		emitted = i
	}

	keyMsg := &runtimemsg.KeyMsg{Special: runtimeplatform.KeyEscape}
	inst.HandleKeyMessage(keyMsg)

	if emitted == nil {
		t.Error("closeIntent should be emitted on ESC close")
	}
}

// =============================================================================
// Measure Tests
// =============================================================================

func TestInstance_Measure_Closed(t *testing.T) {
	inst := NewInstance(rtui.Props{propIsOpen: false, propWidth: 30, propHeight: 20})
	size := inst.Measure(layout.UnboundedConstraints())
	if size.Width != 0 || size.Height != 0 {
		t.Errorf("Measure closed drawer = (%d,%d), want (0,0)", size.Width, size.Height)
	}
}

func TestInstance_Measure_Open(t *testing.T) {
	inst := NewInstance(rtui.Props{propIsOpen: true, propWidth: 30, propHeight: 20})
	size := inst.Measure(layout.UnboundedConstraints())
	if size.Width != 30 || size.Height != 20 {
		t.Errorf("Measure open drawer = (%d,%d), want (30,20)", size.Width, size.Height)
	}
}

// =============================================================================
// GetBoxModel Tests
// =============================================================================

func TestInstance_GetBoxModel_Closed(t *testing.T) {
	inst := NewInstance(rtui.Props{propIsOpen: false})
	bm := inst.GetBoxModel()
	if bm.Border.Style != layout.BorderNone {
		t.Errorf("closed drawer should have BorderNone, got %v", bm.Border.Style)
	}
}

func TestInstance_GetBoxModel_Open_WithTitle(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propIsOpen: true,
		propTitle:  "Settings",
		propPadding: 1,
	})
	bm := inst.GetBoxModel()
	// padding(1) + title rows(2) = 3
	if bm.Padding.Top != 3 {
		t.Errorf("top padding = %d, want 3", bm.Padding.Top)
	}
	if bm.Padding.Left != 1 || bm.Padding.Right != 1 || bm.Padding.Bottom != 1 {
		t.Errorf("unexpected padding: %+v", bm.Padding)
	}
}

func TestInstance_GetBoxModel_Open_NoTitle(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propIsOpen:  true,
		propPadding: 1,
	})
	bm := inst.GetBoxModel()
	if bm.Padding.Top != 1 {
		t.Errorf("top padding = %d, want 1 (no title)", bm.Padding.Top)
	}
}

// =============================================================================
// Registry Tests
// =============================================================================

func TestRegistry_RegisterUnregister(t *testing.T) {
	reg := &drawerRegistry{drawers: make(map[*Instance]bool)}

	inst1 := NewInstance(rtui.Props{propIsOpen: true})
	inst2 := NewInstance(rtui.Props{propIsOpen: true})

	reg.register(inst1)
	reg.register(inst2)

	if len(reg.order) != 2 {
		t.Errorf("registry order len = %d, want 2", len(reg.order))
	}

	reg.unregister(inst1)
	if len(reg.order) != 1 {
		t.Errorf("registry order len = %d, want 1 after unregister", len(reg.order))
	}
	if reg.order[0] != inst2 {
		t.Error("remaining registry entry should be inst2")
	}
}

func TestRegistry_RegisterNil(t *testing.T) {
	reg := &drawerRegistry{drawers: make(map[*Instance]bool)}
	reg.register(nil) // should not panic
	reg.unregister(nil)
}

// =============================================================================
// Helpers
// =============================================================================

func findCommandAt(cmds []paint.DrawCmd, x, y int) (paint.DrawCmd, bool) {
	for _, cmd := range cmds {
		if cmd.X == x && cmd.Y == y {
			return cmd, true
		}
	}
	return paint.DrawCmd{}, false
}

func findCommandContaining(cmds []paint.DrawCmd, y int, text string) (paint.DrawCmd, bool) {
	for _, cmd := range cmds {
		if cmd.Y == y && strings.Contains(cmd.Text, text) {
			return cmd, true
		}
	}
	return paint.DrawCmd{}, false
}
