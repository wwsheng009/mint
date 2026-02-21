package panel

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	newtext "github.com/wwsheng009/mint/ui/components/text"
)

// =============================================================================
// Summary Tests for Phase 2 API Enhancements
// =============================================================================

func TestPhase2_VNodeInnerDimensions(t *testing.T) {
	// Test SetInnerWidth
	v1 := New().SetBorderStyle(layout.BorderSingle).SetInnerWidth(20)
	if v1.width != 22 { // 20 + 2 border padding
		t.Errorf("Expected outer width 22, got %d", v1.width)
	}

	// Test SetInnerHeight
	v2 := New().SetBorderStyle(layout.BorderSingle).SetInnerHeight(5)
	if v2.height != 7 { // 5 + 2 border padding
		t.Errorf("Expected outer height 7, got %d", v2.height)
	}

	// Test GetInnerDimensions
	v3 := New().SetBorderStyle(layout.BorderSingle).SetOuterSize(22, 7)
	innerW, innerH := v3.GetInnerDimensions()
	if innerW != 20 || innerH != 5 {
		t.Errorf("Expected inner dimensions 20x5, got %dx%d", innerW, innerH)
	}
}

func TestPhase2_VNodeAutoMethods(t *testing.T) {
	// Test AutoSize
	v := New().AutoSize()
	if v.width != 0 || v.height != 0 {
		t.Errorf("Expected auto size 0x0, got %dx%d", v.width, v.height)
	}

	// Test FixedWidthAutoHeight
	v2 := New().FixedWidthAutoHeight(30)
	if v2.width != 30 || v2.height != 0 {
		t.Errorf("Expected 30x0, got %dx%d", v2.width, v2.height)
	}
}

func TestPhase2_VNodeTextContent(t *testing.T) {
	// Test SetWrappedTextContent
	v := New().SetWrappedTextContent("Hello world", 20)
	if v.width != 22 { // 20 + 2 border padding
		t.Errorf("Expected width 22, got %d", v.width)
	}
	if v.content == nil {
		t.Error("Content should not be nil")
	}
}

func TestPhase2_VNodePresets(t *testing.T) {
	// Test InfoPanel
	info := InfoPanel("Info", "Message")
	if info.title != "Info" {
		t.Errorf("Expected title 'Info', got '%s'", info.title)
	}
	if info.borderStyle != layout.BorderSingle {
		t.Error("Expected BorderSingle")
	}
	if info.borderColor != "blue" {
		t.Errorf("Expected color 'blue', got '%s'", info.borderColor)
	}

	// Test ErrorPanel
	err := ErrorPanel("Error", "Error message")
	if err.borderStyle != layout.BorderDouble {
		t.Error("Expected BorderDouble")
	}
	if err.borderColor != "red" {
		t.Errorf("Expected color 'red', got '%s'", err.borderColor)
	}

	// Test TextPanel
	tp := TextPanel("Title", "Content", 20)
	if tp.title != "Title" {
		t.Errorf("Expected title 'Title', got '%s'", tp.title)
	}
	if tp.width != 22 { // 20 + 2 border padding
		t.Errorf("Expected width 22, got %d", tp.width)
	}
}

func TestPhase2_BuilderInnerDimensions(t *testing.T) {
	// Test InnerWidth
	v1 := NewBuilder().BorderStyle(layout.BorderSingle).InnerWidth(20).BuildVNode()
	if v1.width != 22 {
		t.Errorf("Expected outer width 22, got %d", v1.width)
	}

	// Test InnerHeight
	v2 := NewBuilder().BorderStyle(layout.BorderSingle).InnerHeight(5).BuildVNode()
	if v2.height != 7 {
		t.Errorf("Expected outer height 7, got %d", v2.height)
	}

	// Test ContentSize
	v3 := NewBuilder().ContentSize(20, 5).BuildVNode()
	if v3.width != 22 || v3.height != 7 {
		t.Errorf("Expected 22x7, got %dx%d", v3.width, v3.height)
	}
}

func TestPhase2_BuilderAutoMethods(t *testing.T) {
	// Test AutoSize
	v := NewBuilder().AutoSize().BuildVNode()
	if v.width != 0 || v.height != 0 {
		t.Errorf("Expected 0x0, got %dx%d", v.width, v.height)
	}

	// Test Fixed
	v2 := NewBuilder().Fixed(30, 10).BuildVNode()
	if v2.width != 30 || v2.height != 10 {
		t.Errorf("Expected 30x10, got %dx%d", v2.width, v2.height)
	}

	// Test FixedInner
	v3 := NewBuilder().FixedInner(20, 5).BuildVNode()
	if v3.width != 22 || v3.height != 7 {
		t.Errorf("Expected 22x7, got %dx%d", v3.width, v3.height)
	}
}

func TestPhase2_BuilderTextContent(t *testing.T) {
	// Test WithWrappedText
	v := NewBuilder().WithWrappedText("Hello world", 20).BuildVNode()
	if v.width != 22 {
		t.Errorf("Expected width 22, got %d", v.width)
	}
	if v.content == nil {
		t.Error("Content should not be nil")
	}

	// Test TextPanel method
	v2 := NewBuilder().TextPanel("Title", "Content", 20).BuildVNode()
	if v2.title != "Title" {
		t.Errorf("Expected title 'Title', got '%s'", v2.title)
	}
	if v2.width != 22 {
		t.Errorf("Expected width 22, got %d", v2.width)
	}
}

func TestPhase2_BuilderConvenience(t *testing.T) {
	// Test WithBorder
	v := NewBuilder().WithBorder(layout.BorderDouble, style.Color("red")).BuildVNode()
	if v.borderStyle != layout.BorderDouble {
		t.Error("Expected BorderDouble")
	}
	if v.borderColor != "red" {
		t.Errorf("Expected color 'red', got '%s'", v.borderColor)
	}

	// Test WithContentOnly
	content := newtext.New("Content")
	v2 := NewBuilder().WithContentOnly(content).BuildVNode()
	if v2.content != content {
		t.Error("Content mismatch")
	}
}

func TestPhase2_UtilityFunctions(t *testing.T) {
	// Test CalculateOuterWidth
	outerW := CalculateOuterWidth(20, layout.BorderSingle)
	if outerW != 22 {
		t.Errorf("Expected outer width 22, got %d", outerW)
	}

	// Test CalculateInnerWidth
	innerW := CalculateInnerWidth(22, layout.BorderSingle)
	if innerW != 20 {
		t.Errorf("Expected inner width 20, got %d", innerW)
	}

	// Test CalculateOuterHeight
	outerH := CalculateOuterHeight(5, layout.BorderDouble)
	if outerH != 7 {
		t.Errorf("Expected outer height 7, got %d", outerH)
	}

	// Test CalculateInnerHeight
	innerH := CalculateInnerHeight(7, layout.BorderDouble)
	if innerH != 5 {
		t.Errorf("Expected inner height 5, got %d", innerH)
	}
}

func TestPhase2_GlobalFactoryFunctions(t *testing.T) {
	// Test Colored
	b1 := Colored(style.Color("red")).BuildVNode()
	if b1.borderColor != "red" {
		t.Errorf("Expected color 'red', got '%s'", b1.borderColor)
	}

	// Test Styled
	b2 := Styled(layout.BorderDouble, style.Color("blue")).BuildVNode()
	if b2.borderStyle != layout.BorderDouble || b2.borderColor != "blue" {
		t.Error("Expected BorderDouble with blue color")
	}

	// Test FixedSize
	b3 := FixedSize(30, 10).BuildVNode()
	if b3.width != 30 || b3.height != 10 {
		t.Errorf("Expected 30x10, got %dx%d", b3.width, b3.height)
	}

	// Test FixedContentSize
	b4 := FixedContentSize(20, 5).BuildVNode()
	if b4.width != 22 || b4.height != 7 {
		t.Errorf("Expected 22x7, got %dx%d", b4.width, b4.height)
	}

	// Test Auto
	b5 := Auto().BuildVNode()
	if b5.width != 0 || b5.height != 0 {
		t.Errorf("Expected 0x0, got %dx%d", b5.width, b5.height)
	}
}

func TestPhase2_ConvenienceFunctions(t *testing.T) {
	content := newtext.New("Content")

	// Test AutoContent
	p1 := AutoContent(content)
	if p1 == nil {
		t.Fatal("AutoContent returned nil")
	}

	// Test TitledAuto
	p2 := TitledAuto("Title", content)
	if p2 == nil {
		t.Fatal("TitledAuto returned nil")
	}

	// Test Text
	p3 := Text("Title", "Content")
	if p3 == nil {
		t.Fatal("Text returned nil")
	}

	// Test Info
	p4 := Info("Info", "Message")
	if p4 == nil {
		t.Fatal("Info returned nil")
	}

	// Test Error
	p5 := Error("Error", "Message")
	if p5 == nil {
		t.Fatal("Error returned nil")
	}

	// Test Box
	p6 := Box(content, 20, 10)
	if p6 == nil {
		t.Fatal("Box returned nil")
	}

	// Test BoxInner
	p7 := BoxInner(content, 18, 8)
	if p7 == nil {
		t.Fatal("BoxInner returned nil")
	}
}

func TestPhase2_BackwardCompatibility(t *testing.T) {
	content := newtext.New("Content")

	// Test all old API still works
	v := New().
		SetTitle("Test").
		SetContent(content).
		SetWidth(30).
		SetHeight(10).
		SetBorderStyle(layout.BorderRounded).
		SetBorderColor(style.Color("blue"))

	if v.title != "Test" {
		t.Errorf("Expected title 'Test', got '%s'", v.title)
	}
	if v.width != 30 || v.height != 10 {
		t.Errorf("Expected 30x10, got %dx%d", v.width, v.height)
	}

	// Test old builder methods still work
	v2 := NewBuilder().
		Title("Test").
		Content(content).
		Width(30).
		Height(10).
		BorderStyle(layout.BorderRounded).
		BorderColor(style.Color("blue")).
		BuildVNode()

	if v2.title != "Test" {
		t.Errorf("Expected title 'Test', got '%s'", v2.title)
	}
	if v2.width != 30 || v2.height != 10 {
		t.Errorf("Expected 30x10, got %dx%d", v2.width, v2.height)
	}
}
