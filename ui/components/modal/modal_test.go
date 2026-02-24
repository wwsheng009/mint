package modal

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	newtext "github.com/wwsheng009/mint/ui/components/text"
)

// =============================================================================
// VNode Tests
// =============================================================================

func TestVNode_New(t *testing.T) {
	vnode := New()
	if vnode == nil {
		t.Fatal("New() returned nil")
	}
	if vnode.Tag() != "modal" {
		t.Errorf("Expected tag 'modal', got '%s'", vnode.Tag())
	}
}

func TestVNode_ImplementsInterfaces(t *testing.T) {
	vnode := New()

	// Test VNode interface
	var _ rtui.VNode = vnode

	// Test InstanceFactory interface
	var _ rtui.InstanceFactory = vnode
}

func TestVNode_FluentAPI(t *testing.T) {
	textNode := newtext.New("Content")

	vnode := New().
		SetTitle("Test Modal").
		SetContent(textNode).
		SetFooter(newtext.New("Footer")).
		SetWidth(40).
		SetHeight(15).
		SetCloseable(true).
		SetCentered(true).
		SetBorderStyle("rounded")

	if vnode.title != "Test Modal" {
		t.Errorf("Expected title 'Test Modal', got '%s'", vnode.title)
	}
	if vnode.width != 40 {
		t.Errorf("Expected width 40, got %d", vnode.width)
	}
	if vnode.height != 15 {
		t.Errorf("Expected height 15, got %d", vnode.height)
	}
	if !vnode.closeable {
		t.Error("Expected closeable to be true")
	}
	if !vnode.centered {
		t.Error("Expected centered to be true")
	}
	if vnode.borderStyle != "rounded" {
		t.Errorf("Expected border style 'rounded', got '%s'", vnode.borderStyle)
	}
	if vnode.content != textNode {
		t.Error("Content mismatch")
	}
}

func TestVNode_OpenClose(t *testing.T) {
	vnode := New()

	if vnode.isOpen {
		t.Error("Expected modal to be closed initially")
	}

	vnode.SetOpen(true)
	if !vnode.isOpen {
		t.Error("Expected modal to be open after Open(true)")
	}

	vnode.Close()
	if vnode.isOpen {
		t.Error("Expected modal to be closed after Close()")
	}

	vnode.Open()
	if !vnode.isOpen {
		t.Error("Expected modal to be open after Open()")
	}

	vnode.Toggle()
	if vnode.isOpen {
		t.Error("Expected modal to be closed after Toggle()")
	}

	vnode.Toggle()
	if !vnode.isOpen {
		t.Error("Expected modal to be open after Toggle()")
	}
}

func TestVNode_CreateInstance(t *testing.T) {
	vnode := New().
		SetTitle("Test").
		SetWidth(30).
		SetHeight(10).
		SetContent(newtext.New("Content"))

	inst := vnode.CreateInstance()
	if inst == nil {
		t.Fatal("CreateInstance() returned nil")
	}
	if _, ok := inst.(*Instance); !ok {
		t.Error("CreateInstance() did not return *Instance")
	}
}

func TestVNode_Props(t *testing.T) {
	closeIntent := intent.CloseModal("test")

	vnode := New().
		SetTitle("Test Modal").
		SetOpen(true).
		SetWidth(40).
		SetHeight(15).
		SetIntent(closeIntent)

	props := vnode.Props()
	if props["title"] != "Test Modal" {
		t.Errorf("Expected prop 'title' to be 'Test Modal', got '%v'", props["title"])
	}
	if props["isOpen"] != true {
		t.Error("Expected prop 'isOpen' to be true")
	}
	if props["width"] != 40 {
		t.Errorf("Expected prop 'width' to be 40, got %v", props["width"])
	}
	if props["closeIntent"] != closeIntent {
		t.Error("Expected prop 'closeIntent' to match")
	}
}

func TestVNode_Layer(t *testing.T) {
	vnode := New()
	if vnode.GetLayer() != rtui.LayerOverlay {
		t.Errorf("Expected LayerOverlay, got %v", vnode.GetLayer())
	}
}

func TestVNode_BorderStyles(t *testing.T) {
	tests := []struct {
		style      string
		expected   string
	}{
		{"single", "single"},
		{"double", "double"},
		{"rounded", "rounded"},
		{"dashed", "dashed"},
	}

	for _, tt := range tests {
		t.Run(tt.style, func(t *testing.T) {
			vnode := New().SetBorderStyle(tt.style)
			if vnode.BorderStyle() != tt.expected {
				t.Errorf("Expected border style '%s', got '%s'", tt.expected, vnode.BorderStyle())
			}
		})
	}
}

func TestVNode_Children(t *testing.T) {
	content := newtext.New("Content")
	footer := newtext.New("Footer")

	vnode := New().
		SetContent(content).
		SetFooter(footer)

	children := vnode.Children()
	if children == nil {
		t.Fatal("Children should not be nil")
	}
	if len(children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(children))
	}
	if children[0] != content {
		t.Error("First child should be content")
	}
	if children[1] != footer {
		t.Error("Second child should be footer")
	}
}

// =============================================================================
// Builder Tests
// =============================================================================

func TestBuilder_FluentAPI(t *testing.T) {
	vnode := NewBuilder().
		Title("Modal Title").
		Content(newtext.New("Content")).
		Footer(newtext.New("Footer")).
		Width(50).
		Height(20).
		Centered(true).
		Closeable(true).
		Rounded().
		BuildVNode()

	if vnode.title != "Modal Title" {
		t.Errorf("Expected title 'Modal Title', got '%s'", vnode.title)
	}
	if vnode.width != 50 {
		t.Errorf("Expected width 50, got %d", vnode.width)
	}
	if vnode.height != 20 {
		t.Errorf("Expected height 20, got %d", vnode.height)
	}
	if !vnode.centered {
		t.Error("Expected centered to be true")
	}
	if !vnode.closeable {
		t.Error("Expected closeable to be true")
	}
}

func TestBuilder_OpenClose(t *testing.T) {
	vnode := NewBuilder().
		Opened().
		BuildVNode()

	if !vnode.isOpen {
		t.Error("Expected modal to be open after Opened()")
	}

	vnode = NewBuilder().
		Closed().
		BuildVNode()

	if vnode.isOpen {
		t.Error("Expected modal to be closed after Closed()")
	}
}

func TestBuilder_Size(t *testing.T) {
	vnode := NewBuilder().
		Size(30, 15).
		BuildVNode()

	if vnode.width != 30 {
		t.Errorf("Expected width 30, got %d", vnode.width)
	}
	if vnode.height != 15 {
		t.Errorf("Expected height 15, got %d", vnode.height)
	}
}

func TestBuilder_BorderStyles(t *testing.T) {
	tests := []struct {
		name     string
		method   func(*Builder) *Builder
		expected string
	}{
		{"Single", func(b *Builder) *Builder { return b.Single() }, "single"},
		{"Double", func(b *Builder) *Builder { return b.Double() }, "double"},
		{"Rounded", func(b *Builder) *Builder { return b.Rounded() }, "rounded"},
		{"Dashed", func(b *Builder) *Builder { return b.Dashed() }, "dashed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vnode := tt.method(NewBuilder()).BuildVNode()
			if vnode.BorderStyle() != tt.expected {
				t.Errorf("Expected border style '%s', got '%s'", tt.expected, vnode.BorderStyle())
			}
		})
	}
}

func TestBuilder_ConvenienceFunctions(t *testing.T) {
	content := newtext.New("Content")

	// Test Of
	vnode := Of(content)
	if vnode.(*VNode).content != content {
		t.Error("Of() should set content")
	}

	// Test OfSize
	vnode = OfSize(content, 40, 15)
	vn := vnode.(*VNode)
	if vn.width != 40 || vn.height != 15 {
		t.Error("OfSize() should set dimensions")
	}

	// Test Titled
	vnode = Titled("Title", content)
	vn = vnode.(*VNode)
	if vn.title != "Title" {
		t.Error("Titled() should set title")
	}

	// Test Alert
	vnode = Alert("Alert", "Alert message")
	if vnode.(*VNode).title != "Alert" {
		t.Error("Alert() should set title")
	}

	// Test Confirm
	vnode = Confirm("Confirm", "Confirm message")
	if vnode.(*VNode).title != "Confirm" {
		t.Error("Confirm() should set title")
	}
}

// =============================================================================
// Instance Tests
// =============================================================================

func TestInstance_NewInstance(t *testing.T) {
	props := rtui.Props{
		"title":       "Test Modal",
		"isOpen":      true,
		"width":       40,
		"height":      15,
		"borderStyle": "rounded",
	}

	inst := NewInstance(props)
	if inst == nil {
		t.Fatal("NewInstance() returned nil")
	}
	if inst.title != "Test Modal" {
		t.Errorf("Expected title 'Test Modal', got '%s'", inst.title)
	}
	if !inst.isOpen {
		t.Error("Expected isOpen to be true")
	}
	if inst.width != 40 {
		t.Errorf("Expected width 40, got %d", inst.width)
	}
	if inst.borderStyle != "rounded" {
		t.Errorf("Expected border style 'rounded', got '%s'", inst.borderStyle)
	}
}

func TestInstance_Measure(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"isOpen": true,
		"width":  40,
		"height": 15,
	})

	size := inst.Measure(layout.Constraints{
		MinWidth:  0,
		MaxWidth:  100,
		MinHeight: 0,
		MaxHeight: 100,
	})

	if size.Width != 40 {
		t.Errorf("Expected width 40, got %d", size.Width)
	}
	if size.Height != 15 {
		t.Errorf("Expected height 15, got %d", size.Height)
	}
}

func TestInstance_MeasureClosed(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"isOpen": false,
		"width":  40,
		"height": 15,
	})

	size := inst.Measure(layout.Constraints{})

	if size.Width != 0 || size.Height != 0 {
		t.Error("Closed modal should have size 0,0")
	}
}

func TestInstance_Paint(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"isOpen":  true,
		"title":   "Test",
		"width":   20,
		"height":  8,
		"modalStyle": style.Style{FG: style.Color("blue")},
	})

	cmds := inst.Paint(0, 0)
	if cmds == nil {
		t.Error("Paint() should return commands")
	}
	if len(cmds) == 0 {
		t.Error("Paint() should return non-empty commands")
	}

	// Verify bounds are set
	if inst.bounds[2] != 20 {
		t.Errorf("Expected bounds width 20, got %d", inst.bounds[2])
	}
	if inst.bounds[3] != 8 {
		t.Errorf("Expected bounds height 8, got %d", inst.bounds[3])
	}
}

func TestInstance_PaintClosed(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"isOpen": false,
	})

	cmds := inst.Paint(0, 0)
	if cmds != nil {
		t.Error("Closed modal should not paint")
	}
}

func TestInstance_BorderChars(t *testing.T) {
	tests := []struct {
		style     string
		horizontal rune
		vertical   rune
		topleft    rune
	}{
		{"single", '─', '│', '┌'},
		{"double", '═', '║', '╔'},
		{"rounded", '─', '│', '╭'},
		{"dashed", '─', '│', '┌'},
	}

	for _, tt := range tests {
		t.Run(tt.style, func(t *testing.T) {
			inst := NewInstance(rtui.Props{
				"borderStyle": tt.style,
			})
			chars := inst.getBorderChars()
			if chars.horizontal != tt.horizontal {
				t.Errorf("Expected horizontal rune '%c', got '%c'", tt.horizontal, chars.horizontal)
			}
			if chars.vertical != tt.vertical {
				t.Errorf("Expected vertical rune '%c', got '%c'", tt.vertical, chars.vertical)
			}
			if chars.topLeft != tt.topleft {
				t.Errorf("Expected topLeft rune '%c', got '%c'", tt.topleft, chars.topLeft)
			}
		})
	}
}

func TestInstance_SetProps(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"isOpen": false,
		"title":  "Old Title",
	})

	changed := inst.SetProps(rtui.Props{
		"isOpen": true,
		"title":  "New Title",
	})

	if !changed {
		t.Error("SetProps should return true when props changed")
	}
	if !inst.isOpen {
		t.Error("isOpen should be updated")
	}
	if inst.title != "New Title" {
		t.Error("title should be updated")
	}
}

func TestInstance_HandleAction(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"isOpen":    true,
		"closeable": true,
	})

	// Test close action
	handled := inst.HandleAction("close", nil)
	if !handled {
		t.Error("HandleAction should return true for close action")
	}
	if inst.isOpen {
		t.Error("Modal should be closed after close action")
	}

	// Test closed modal doesn't handle actions
	inst.isOpen = true
	inst.HandleAction("close", nil)
	inst.isOpen = false
	handled = inst.HandleAction("close", nil)
	if handled {
		t.Error("Closed modal should not handle actions")
	}
}

func TestInstance_CanHandleAction(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"isOpen":    true,
		"closeable": true,
	})

	tests := []struct {
		action   string
		handlable bool
	}{
		{"close", true},
		{"click_outside", true},
		{"escape", true},
		{"other", false},
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			canHandle := inst.CanHandleAction(tt.action)
			if canHandle != tt.handlable {
				t.Errorf("Expected CanHandleAction('%s') to be %v, got %v", tt.action, tt.handlable, canHandle)
			}
		})
	}
}

func TestInstance_ContainsPoint(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"width":  40,
		"height": 15,
	})
	inst.bounds = [4]int{10, 20, 40, 15}

	tests := []struct {
		x, y     int
		contains bool
	}{
		{10, 20, true},    // top-left
		{49, 34, true},    // bottom-right (inside)
		{50, 20, false},   // right of bounds
		{10, 35, false},   // below bounds
		{5, 15, false},    // before bounds
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			contains := inst.containsPoint(tt.x, tt.y)
			if contains != tt.contains {
				t.Errorf("Expected containsPoint(%d,%d) to be %v, got %v", tt.x, tt.y, tt.contains, contains)
			}
		})
	}
}

func TestInstance_PropHelpers(t *testing.T) {
	props := rtui.Props{
		"title":       "Test Title",
		"isOpen":      true,
		"width":       40,
		"style":       style.Style{FG: style.Color("red")},
		"content":     newtext.New("Content"),
		"closeable":   true,
		"centered":    false,
	}

	if got := getStringProp(props, "title", ""); got != "Test Title" {
		t.Errorf("Expected 'Test Title', got '%s'", got)
	}
	if got := getBoolProp(props, "isOpen", false); !got {
		t.Error("Expected true")
	}
	if got := getIntProp(props, "width", 0); got != 40 {
		t.Errorf("Expected 40, got %d", got)
	}
	if got := getStringProp(props, "style", ""); got != "" {
		t.Error("Style should be handled by getStyleProp")
	}
	if got := getChildProp(props, "content"); got == nil {
		t.Error("Expected content child")
	}
}
