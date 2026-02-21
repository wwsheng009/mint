package panel

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	newtext "github.com/wwsheng009/mint/ui/components/text"
)

// =============================================================================
// Builder Enhancement Tests
// =============================================================================

func TestBuilder_OuterWidthHeight(t *testing.T) {
	b := NewBuilder()
	b.OuterWidth(30).OuterHeight(10)

	if b.vnode.width != 30 {
		t.Errorf("Expected width 30, got %d", b.vnode.width)
	}
	if b.vnode.height != 10 {
		t.Errorf("Expected height 10, got %d", b.vnode.height)
	}
}

func TestBuilder_OuterSize(t *testing.T) {
	b := NewBuilder()
	b.OuterSize(30, 10)

	if b.vnode.width != 30 {
		t.Errorf("Expected width 30, got %d", b.vnode.width)
	}
	if b.vnode.height != 10 {
		t.Errorf("Expected height 10, got %d", b.vnode.height)
	}
}

func TestBuilder_InnerWidth(t *testing.T) {
	tests := []struct {
		name          string
		borderStyle   layout.BorderStyle
		innerWidth    int
		expectedOuter int
	}{
		{"Single border", layout.BorderSingle, 20, 22},
		{"Rounded border", layout.BorderRounded, 20, 22},
		{"Double border", layout.BorderDouble, 20, 22},
		{"No border", layout.BorderNone, 20, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBuilder()
			b.BorderStyle(tt.borderStyle).InnerWidth(tt.innerWidth)

			if b.vnode.width != tt.expectedOuter {
				t.Errorf("Expected outer width %d, got %d", tt.expectedOuter, b.vnode.width)
			}
		})
	}
}

func TestBuilder_ContentWidth_Alias(t *testing.T) {
	b := NewBuilder()
	b.ContentWidth(20)

	borderPadding := 2
	expected := 20 + borderPadding
	if b.vnode.width != expected {
		t.Errorf("Expected width %d, got %d", expected, b.vnode.width)
	}
}

func TestBuilder_InnerHeight(t *testing.T) {
	tests := []struct {
		name           string
		borderStyle    layout.BorderStyle
		innerHeight    int
		expectedOuter  int
	}{
		{"Single border", layout.BorderSingle, 5, 7},
		{"Rounded border", layout.BorderRounded, 5, 7},
		{"No border", layout.BorderNone, 5, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBuilder()
			b.BorderStyle(tt.borderStyle).InnerHeight(tt.innerHeight)

			if b.vnode.height != tt.expectedOuter {
				t.Errorf("Expected outer height %d, got %d", tt.expectedOuter, b.vnode.height)
			}
		})
	}
}

func TestBuilder_ContentHeight_Alias(t *testing.T) {
	b := NewBuilder()
	b.ContentHeight(4)

	borderPadding := 2
	expected := 4 + borderPadding
	if b.vnode.height != expected {
		t.Errorf("Expected height %d, got %d", expected, b.vnode.height)
	}
}

func TestBuilder_ContentSize(t *testing.T) {
	b := NewBuilder()
	b.ContentSize(20, 5)

	borderPadding := 2
	if b.vnode.width != 20+borderPadding {
		t.Errorf("Expected width %d, got %d", 20+borderPadding, b.vnode.width)
	}
	if b.vnode.height != 5+borderPadding {
		t.Errorf("Expected height %d, got %d", 5+borderPadding, b.vnode.height)
	}
}

func TestBuilder_AutoMethods(t *testing.T) {
	// Test AutoWidth
	b1 := NewBuilder().AutoWidth()
	if b1.vnode.width != 0 {
		t.Errorf("Expected width 0, got %d", b1.vnode.width)
	}

	// Test AutoHeight
	b2 := NewBuilder().AutoHeight()
	if b2.vnode.height != 0 {
		t.Errorf("Expected height 0, got %d", b2.vnode.height)
	}

	// Test AutoSize
	b3 := NewBuilder().AutoSize()
	if b3.vnode.width != 0 || b3.vnode.height != 0 {
		t.Errorf("Expected 0x0, got %dx%d", b3.vnode.width, b3.vnode.height)
	}
}

func TestBuilder_Fixed(t *testing.T) {
	b := NewBuilder()
	b.Fixed(30, 10)

	if b.vnode.width != 30 {
		t.Errorf("Expected width 30, got %d", b.vnode.width)
	}
	if b.vnode.height != 10 {
		t.Errorf("Expected height 10, got %d", b.vnode.height)
	}
}

func TestBuilder_FixedInner(t *testing.T) {
	b := NewBuilder()
	b.FixedInner(20, 5)

	borderPadding := 2
	if b.vnode.width != 20+borderPadding {
		t.Errorf("Expected width %d, got %d", 20+borderPadding, b.vnode.width)
	}
	if b.vnode.height != 5+borderPadding {
		t.Errorf("Expected height %d, got %d", 5+borderPadding, b.vnode.height)
	}
}

func TestBuilder_FixedWidthAutoHeight(t *testing.T) {
	b := NewBuilder().FixedWidthAutoHeight(30)

	if b.vnode.width != 30 {
		t.Errorf("Expected width 30, got %d", b.vnode.width)
	}
	if b.vnode.height != 0 {
		t.Errorf("Expected height 0, got %d", b.vnode.height)
	}
}

func TestBuilder_FixedHeightAutoWidth(t *testing.T) {
	b := NewBuilder().FixedHeightAutoWidth(10)

	if b.vnode.width != 0 {
		t.Errorf("Expected width 0, got %d", b.vnode.width)
	}
	if b.vnode.height != 10 {
		t.Errorf("Expected height 10, got %d", b.vnode.height)
	}
}

func TestBuilder_WithTextContent(t *testing.T) {
	content := "Auto-sized content"
	b := NewBuilder().WithTitle("Test").WithTextContent(content)

	if b.vnode.content == nil {
		t.Fatal("Content should not be nil")
	}

	textNode, ok := b.vnode.content.(*newtext.VNode)
	if !ok {
		t.Fatal("Content should be a Text VNode")
	}

	if textNode == nil {
		t.Error("Text node should not be nil")
	}
}

func TestBuilder_WithWrappedText(t *testing.T) {
	content := "Long text that should wrap"
	width := 20

	b := NewBuilder().WithWrappedText(content, width)

	// Width should include border padding
	borderPadding := 2
	if b.vnode.width != width+borderPadding {
		t.Errorf("Expected width %d, got %d", width+borderPadding, b.vnode.width)
	}

	if b.vnode.content == nil {
		t.Fatal("Content should not be nil")
	}

	textNode, ok := b.vnode.content.(*newtext.VNode)
	if !ok {
		t.Fatal("Content should be a Text VNode")
	}

	if textNode == nil {
		t.Error("Text node should not be nil")
	}
}

func TestBuilder_TextPanel(t *testing.T) {
	title := "My Title"
	content := "Text content goes here"
	width := 20

	b := NewBuilder().TextPanel(title, content, width)

	if b.vnode.title != title {
		t.Errorf("Expected title '%s', got '%s'", title, b.vnode.title)
	}

	// Width should include border padding
	borderPadding := 2
	if b.vnode.width != width+borderPadding {
		t.Errorf("Expected width %d, got %d", width+borderPadding, b.vnode.width)
	}

	if b.vnode.content == nil {
		t.Fatal("Content should not be nil")
	}
}

func TestBuilder_WithBorder(t *testing.T) {
	b := NewBuilder().
		WithBorder(layout.BorderDouble, style.Color("red"))

	if b.vnode.borderStyle != layout.BorderDouble {
		t.Errorf("Expected BorderDouble, got %v", b.vnode.borderStyle)
	}
	if b.vnode.borderColor != "red" {
		t.Errorf("Expected color 'red', got '%s'", b.vnode.borderColor)
	}
}

func TestBuilder_WithContentOnly(t *testing.T) {
	content := newtext.New("Content")
	b := NewBuilder().WithContentOnly(content)

	if b.vnode.content != content {
		t.Error("Content mismatch")
	}

	// Should enable auto sizing
	if b.vnode.width != 0 || b.vnode.height != 0 {
		t.Errorf("Expected auto size 0x0, got %dx%d", b.vnode.width, b.vnode.height)
	}
}

func TestBuilder_MaybeTitle(t *testing.T) {
	// With non-empty title
	b1 := NewBuilder().MaybeTitle("Test")
	if b1.vnode.title != "Test" {
		t.Error("Expected title 'Test'")
	}

	// With empty title
	b2 := NewBuilder().MaybeTitle("")
	if b2.vnode.title != "" {
		t.Error("Expected empty title")
	}
}

func TestBuilder_IfTitle(t *testing.T) {
	predicateTrue := func() bool { return true }
	predicateFalse := func() bool { return false }

	// If predicate is true
	b1 := NewBuilder().IfTitle("Test", predicateTrue)
	if b1.vnode.title != "Test" {
		t.Error("Expected title 'Test' with true predicate")
	}

	// If predicate is false
	b2 := NewBuilder().IfTitle("Test", predicateFalse)
	if b2.vnode.title != "" {
		t.Error("Expected empty title with false predicate")
	}
}

func TestBuilder_BuildVNode(t *testing.T) {
	b := NewBuilder().
		Title("Test").
		Width(30).
		Height(10)

	vnode := b.BuildVNode()
	if vnode == nil {
		t.Fatal("BuildVNode() returned nil")
	}
	if vnode.title != "Test" {
		t.Errorf("Expected title 'Test', got '%s'", vnode.title)
	}
	if vnode.width != 30 || vnode.height != 10 {
		t.Errorf("Expected size 30x10, got %dx%d", vnode.width, vnode.height)
	}
}

func TestBuilder_BackwardCompatibility_AllOldMethods(t *testing.T) {
	content := newtext.New("Content")
	header := newtext.New("Header")
	footer := newtext.New("Footer")

	b := NewBuilder().
		Key("test-key").
		Title("Test Title").
		Header(header).
		Content(content).
		Footer(footer).
		Width(40).
		Height(15).
		Flex(1).
		Padding(2).
		Rounded().
		BorderColor(style.Color("blue")).
		BorderLabel("Label")

	if b.vnode.width != 40 || b.vnode.height != 15 {
		t.Errorf("Expected size 40x15, got %dx%d", b.vnode.width, b.vnode.height)
	}
	if b.vnode.flex != 1 {
		t.Errorf("Expected flex 1, got %d", b.vnode.flex)
	}
	if b.vnode.padding != 2 {
		t.Errorf("Expected padding 2, got %d", b.vnode.padding)
	}
	if b.vnode.borderStyle != layout.BorderRounded {
		t.Error("Expected rounded border")
	}
}
