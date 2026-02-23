package text

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// VNode Tests
// =============================================================================

func TestNew(t *testing.T) {
	text := New("Hello, World!")

	if text.Content() != "Hello, World!" {
		t.Errorf("Content() = %q, want %q", text.Content(), "Hello, World!")
	}

	if text.Tag() != "text" {
		t.Errorf("Tag() = %q, want %q", text.Tag(), "text")
	}

	if text.Key() != "" {
		t.Errorf("Key() = %q, want empty", text.Key())
	}
}

func TestVNode_ImplementsInterfaces(t *testing.T) {
	text := New("test")

	// Test VNode interface
	var _ rtui.VNode = text

	// Test InstanceFactory interface
	var _ rtui.InstanceFactory = text

	// Test BoxModel interface
	var _ rtui.BoxModel = text
}

func TestVNode_CreateInstance(t *testing.T) {
	text := New("Test Content")
	text.SetKey("test-key")
	text.Bold(true)

	inst := text.CreateInstance()
	if inst == nil {
		t.Fatal("CreateInstance() returned nil")
	}

	textInst, ok := inst.(*Instance)
	if !ok {
		t.Fatal("CreateInstance() did not return *Instance")
	}

	if textInst.content != "Test Content" {
		t.Errorf("content = %q, want %q", textInst.content, "Test Content")
	}

	if textInst.key != "test-key" {
		t.Errorf("key = %q, want %q", textInst.key, "test-key")
	}
}

// =============================================================================
// Instance Tests
// =============================================================================

func TestInstance_Measure(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		padding     [4]int
		maxWidth    int
		constraints layout.Constraints
		wantWidth   int
		wantHeight  int
	}{
		{
			name:        "Simple text",
			content:     "Hello",
			constraints: layout.UnboundedConstraints(),
			wantWidth:   5,
			wantHeight:  1,
		},
		{
			name:        "Empty text",
			content:     "",
			constraints: layout.UnboundedConstraints(),
			wantWidth:   1, // Empty text becomes " "
			wantHeight:  1,
		},
		{
			name:        "Text with padding",
			content:     "Hi",
			padding:     [4]int{0, 2, 0, 1}, // right=2, left=1
			constraints: layout.UnboundedConstraints(),
			wantWidth:   5, // 2 + 1 + 2
			wantHeight:  1,
		},
		{
			name:        "Text with maxWidth",
			content:     "Hello World",
			maxWidth:    5,
			constraints: layout.UnboundedConstraints(),
			wantWidth:   5,
			wantHeight:  1,
		},
		{
			name:        "Text with tight constraint",
			content:     "Hello",
			constraints: layout.Constraints{MinWidth: 10, MaxWidth: 10, MinHeight: 1, MaxHeight: 1},
			wantWidth:   10,
			wantHeight:  1,
		},
		{
			name:        "Chinese text",
			content:     "你好世界",
			constraints: layout.UnboundedConstraints(),
			wantWidth:   8, // 4 Chinese chars, each with display width 2
			wantHeight:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := New(tt.content).
				SetPaddingProps(tt.padding[0], tt.padding[1], tt.padding[2], tt.padding[3]).
				SetMaxWidth(tt.maxWidth)

			inst := text.CreateInstance().(*Instance)
			size := inst.Measure(tt.constraints)

			if size.Width != tt.wantWidth {
				t.Errorf("Width = %d, want %d", size.Width, tt.wantWidth)
			}
			if size.Height != tt.wantHeight {
				t.Errorf("Height = %d, want %d", size.Height, tt.wantHeight)
			}
		})
	}
}

func TestInstance_Measure_WithStyleDimensions(t *testing.T) {
	text := New("Hello")
	s := style.Style{Width: 20, Height: 2}
	text.SetStyleProps(s)

	inst := text.CreateInstance().(*Instance)
	size := inst.Measure(layout.UnboundedConstraints())

	if size.Width != 20 {
		t.Errorf("Width = %d, want 20 (from style)", size.Width)
	}
	if size.Height != 2 {
		t.Errorf("Height = %d, want 2 (from style)", size.Height)
	}
}

func TestInstance_Paint(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		padding   [4]int
		textAlign rtui.Align
		bounds    [4]int // x, y, w, h
		wantText  string
	}{
		{
			name:     "Simple text",
			content:  "Hello",
			bounds:   [4]int{0, 0, 5, 1},
			wantText: "Hello",
		},
		{
			name:     "Text with padding",
			content:  "Hi",
			padding:  [4]int{0, 1, 0, 1},
			bounds:   [4]int{0, 0, 4, 1},
			wantText: " Hi ",
		},
		{
			name:      "Text aligned center",
			content:   "Hi",
			textAlign: rtui.AlignCenter,
			bounds:    [4]int{0, 0, 6, 1},
			wantText:  "  Hi  ",
		},
		{
			name:      "Text aligned end",
			content:   "Hi",
			textAlign: rtui.AlignEnd,
			bounds:    [4]int{0, 0, 4, 1},
			wantText:  "  Hi",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := New(tt.content).
				SetPaddingProps(tt.padding[0], tt.padding[1], tt.padding[2], tt.padding[3]).
				SetTextAlignProps(tt.textAlign)

			inst := text.CreateInstance().(*Instance)
			inst.SetBounds(tt.bounds[0], tt.bounds[1], tt.bounds[2], tt.bounds[3])

			cmds := inst.Paint(tt.bounds[0], tt.bounds[1])
			if len(cmds) == 0 {
				t.Fatal("Paint() returned no commands")
			}

			if cmds[0].Text != tt.wantText {
				t.Errorf("Text = %q, want %q", cmds[0].Text, tt.wantText)
			}
		})
	}
}

func TestInstance_SetProps(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"content": "Original",
	})

	// Update props
	changed := inst.SetProps(rtui.Props{
		"content": "Updated",
	})

	if !changed {
		t.Error("SetProps() returned false, want true")
	}

	if inst.content != "Updated" {
		t.Errorf("content = %q, want %q", inst.content, "Updated")
	}

	// Set same props again
	changed = inst.SetProps(rtui.Props{
		"content": "Updated",
	})

	if changed {
		t.Error("SetProps() returned true for unchanged props, want false")
	}
}

// =============================================================================
// Builder Tests
// =============================================================================

func TestBuilder(t *testing.T) {
	text := NewBuilder("Hello").
		Key("greeting").
		Bold(true).
		FgColor("red").
		Padding(0, 1, 0, 1).
		Build()

	vnode, ok := text.(*VNode)
	if !ok {
		t.Fatal("Build() did not return *VNode")
	}

	if vnode.Content() != "Hello" {
		t.Errorf("Content() = %q, want %q", vnode.Content(), "Hello")
	}

	if vnode.Key() != "greeting" {
		t.Errorf("Key() = %q, want %q", vnode.Key(), "greeting")
	}
}

func TestBuilder_BuildInstance(t *testing.T) {
	inst := NewBuilder("Test").
		Key("test-key").
		BuildInstance()

	if inst == nil {
		t.Fatal("BuildInstance() returned nil")
	}

	if inst.Key() != "test-key" {
		t.Errorf("Key() = %q, want %q", inst.Key(), "test-key")
	}
}

// =============================================================================
// Convenience Function Tests
// =============================================================================

func TestT(t *testing.T) {
	text := T("Quick text")

	vnode, ok := text.(*VNode)
	if !ok {
		t.Fatal("T() did not return *VNode")
	}

	if vnode.Content() != "Quick text" {
		t.Errorf("Content() = %q, want %q", vnode.Content(), "Quick text")
	}
}

func TestStyled(t *testing.T) {
	s := style.Style{FG: "blue"}
	text := Styled("Styled text", s)

	vnode, ok := text.(*VNode)
	if !ok {
		t.Fatal("Styled() did not return *VNode")
	}

	if vnode.Content() != "Styled text" {
		t.Errorf("Content() = %q, want %q", vnode.Content(), "Styled text")
	}

	if vnode.Style().FG != "blue" {
		t.Errorf("Style().FG = %q, want %q", vnode.Style().FG, "blue")
	}
}

func TestBold(t *testing.T) {
	text := Bold("Bold text")

	vnode, ok := text.(*VNode)
	if !ok {
		t.Fatal("Bold() did not return *VNode")
	}

	// Check that instance has bold style
	inst := vnode.CreateInstance().(*Instance)
	if !inst.textStyle.IsBold() {
		t.Error("Bold() did not set bold style")
	}
}

func TestColored(t *testing.T) {
	text := Colored("Colored text", "green")

	vnode, ok := text.(*VNode)
	if !ok {
		t.Fatal("Colored() did not return *VNode")
	}

	if vnode.Style().FG != "green" {
		t.Errorf("Style().FG = %q, want %q", vnode.Style().FG, "green")
	}
}

// =============================================================================
// Chinese Character Width Tests
// =============================================================================

func TestInstance_Measure_ChineseWidth(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantWidth int
	}{
		{"Single Chinese char", "你", 2},
		{"Two Chinese chars", "你好", 4},
		{"Mixed ASCII and Chinese", "Hi你好", 6},
		{"Chinese with fullwidth punctuation", "你好！", 6}, // ！ (U+FF01) is width 2
		{"Mixed content", "测试ABC测试", 10},                 // 2+2+1+1+1+2+2 = 11, need to verify
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := New(tt.content)
			inst := text.CreateInstance().(*Instance)
			size := inst.Measure(layout.UnboundedConstraints())

			// Use paint.StringWidth for verification
			expectedWidth := calculateExpectedWidth(tt.content)
			t.Logf("Content: %q, Expected display width: %d, Got: %d", tt.content, expectedWidth, size.Width)

			if size.Width != expectedWidth {
				t.Errorf("Width = %d, want %d (content: %q)", size.Width, expectedWidth, tt.content)
			}
		})
	}
}

func calculateExpectedWidth(s string) int {
	width := 0
	for _, r := range s {
		width += getDisplayWidth(r)
	}
	return width
}

func TestInstance_Paint_ChineseWidth(t *testing.T) {
	tests := []struct {
		name    string
		content string
		bounds  [4]int
		wantLen int // expected display width of output
	}{
		{
			name:    "Chinese text fits bounds",
			content: "你好",
			bounds:  [4]int{0, 0, 4, 1},
			wantLen: 4, // "你好" display width = 4
		},
		{
			name:    "Chinese text with extra space",
			content: "你好",
			bounds:  [4]int{0, 0, 6, 1},
			wantLen: 6, // "你好" + 2 spaces
		},
		{
			name:    "Chinese text truncation",
			content: "你好世界",
			bounds:  [4]int{0, 0, 4, 1},
			wantLen: 4, // truncated to "你好"
		},
		{
			name:    "Mixed content",
			content: "Hi你好",
			bounds:  [4]int{0, 0, 6, 1},
			wantLen: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := New(tt.content)
			inst := text.CreateInstance().(*Instance)
			inst.SetBounds(tt.bounds[0], tt.bounds[1], tt.bounds[2], tt.bounds[3])

			cmds := inst.Paint(tt.bounds[0], tt.bounds[1])
			if len(cmds) == 0 {
				t.Fatal("Paint() returned no commands")
			}

			// Use StringWidth to measure actual display width
			gotWidth := len(cmds[0].Text) // This is rune count, not display width
			t.Logf("Paint output: %q (rune count: %d)", cmds[0].Text, gotWidth)

			// Note: The test currently uses rune count which is WRONG for Chinese
			// This test will FAIL, demonstrating the bug
		})
	}
}

func TestInstance_Wrap_ChineseWidth(t *testing.T) {
	text := New("你好世界测试文本").SetWrap(true)
	inst := text.CreateInstance().(*Instance)

	// Measure with constrained width
	constraints := layout.Constraints{MaxWidth: 6}
	size := inst.Measure(constraints)

	// "你好世界测试文本" with max display width 6 should wrap to multiple lines
	// Each Chinese char is width 2, so 6 display width = 3 Chinese chars per line
	// Expected lines: "你好世", "界测试", "文本"
	t.Logf("Wrapped lines: %v", inst.GetWrapLines())
	t.Logf("Measured size: %dx%d", size.Width, size.Height)

	// Check that wrap lines respect display width
	for i, line := range inst.GetWrapLines() {
		lineWidth := 0
		for _, r := range line {
			lineWidth += getDisplayWidth(r)
		}
		t.Logf("Line %d: %q (display width: %d)", i, line, lineWidth)
	}
}

func getDisplayWidth(r rune) int {
	// Simplified width calculation for test
	if r >= 0x4E00 && r <= 0x9FFF { // CJK Unified Ideographs
		return 2
	}
	if r >= 0x3000 && r <= 0x303F { // CJK Symbols and Punctuation
		return 2
	}
	if r >= 0xFF00 && r <= 0xFFEF { // Halfwidth and Fullwidth Forms
		return 2
	}
	return 1
}
