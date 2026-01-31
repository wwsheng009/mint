package ui

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/style"
)

// TestVNodeType tests VNode type identification
func TestVNodeType(t *testing.T) {
	tests := []struct {
		name     string
		node     VNode
		expected VNodeType
	}{
		{"Element", NewElement("div"), VNodeElement},
		{"Text", NewText("hello"), VNodeText},
		{"Component", NewComponent("Test", func() VNode { return NewText("comp") }), VNodeComponent},
		{"Fragment", NewFragment(), VNodeFragment},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.node.Type(); got != tt.expected {
				t.Errorf("Type() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestVNodeProps tests Props operations
func TestVNodeProps(t *testing.T) {
	node := NewElement("div")

	// Test SetProps
	props := Props{
		"class": "container",
		"id":    "main",
	}
	node.SetProps(props)

	// Test Get
	if got := node.Props().GetString("class"); got != "container" {
		t.Errorf("GetString(class) = %v, want %v", got, "container")
	}

	// Test GetBool
	node.SetProps(Props{"visible": true})
	if got := node.Props().GetBool("visible"); !got {
		t.Errorf("GetBool(visible) = %v, want true", got)
	}

	// Test GetInt
	node.SetProps(Props{"count": 42})
	if got := node.Props().GetInt("count"); got != 42 {
		t.Errorf("GetInt(count) = %v, want 42", got)
	}

	// Test Get with default (using nil check)
	if got := node.Props().Get("missing"); got != nil {
		t.Errorf("Get(missing) = %v, want nil", got)
	}
}

// TestVNodeKey tests Key operations
func TestVNodeKey(t *testing.T) {
	node := NewElement("div")

	// Initial key should be empty
	if got := node.Key(); got != "" {
		t.Errorf("Key() = %v, want empty", got)
	}

	// SetKey
	node.SetKey("test-key")
	if got := node.Key(); got != "test-key" {
		t.Errorf("Key() = %v, want test-key", got)
	}
}

// TestVNodeChildren tests Children operations
func TestVNodeChildren(t *testing.T) {
	parent := NewElement("div")

	// Initial children should be empty
	if got := len(parent.Children()); got != 0 {
		t.Errorf("len(Children()) = %v, want 0", got)
	}

	// Add children
	child1 := NewText("child1")
	child2 := NewText("child2")
	parent.SetChildren([]VNode{child1, child2})

	children := parent.Children()
	if got := len(children); got != 2 {
		t.Errorf("len(Children()) = %v, want 2", got)
	}
}

// TestVNodeStyle tests Style operations
func TestVNodeStyle(t *testing.T) {
	node := NewElement("div")

	// Initial style should be empty
	initialStyle := node.Style()
	if initialStyle.FG != "" || initialStyle.BG != "" {
		t.Errorf("Initial style should be empty, got %v", initialStyle)
	}

	// SetStyle
	newStyle := style.Style{FG: "red", BG: "blue"}
	node.SetStyle(newStyle)

	gotStyle := node.Style()
	if gotStyle.FG != "red" || gotStyle.BG != "blue" {
		t.Errorf("Style() = %v, want {FG:red, BG:blue}", gotStyle)
	}
}

// TestTextVNodeContent tests TextVNode content
func TestTextVNodeContent(t *testing.T) {
	content := "Hello, World!"
	text := NewText(content)

	if got := text.Content(); got != content {
		t.Errorf("Content() = %v, want %v", got, content)
	}
}

// TestTextBuilder tests Text builder pattern
func TestTextBuilder(t *testing.T) {
	text := NewTextBuilder("Hello").
		FgColor("red").
		Bold(true).
		Build()

	// Check it's a TextVNode
	if _, ok := text.(*TextVNode); !ok {
		t.Error("Build() should return *TextVNode")
	}

	// Check style
	s := text.Style()
	if s.FG != "red" {
		t.Errorf("FG = %v, want red", s.FG)
	}
	if !s.IsBold() {
		t.Error("Bold should be true")
	}
}

// TestButtonVNode tests ButtonVNode
func TestButtonVNode(t *testing.T) {
	label := "Click Me"
	btn := NewButton(label)

	if got := btn.Label(); got != label {
		t.Errorf("Label() = %v, want %v", got, label)
	}

	// Test OnClick
	clicked := false
	btn.SetOnClick(func() {
		clicked = true
	})

	if btn.OnClick() == nil {
		t.Error("OnClick() should not be nil")
	}

	// Trigger click
	btn.OnClick()()
	if !clicked {
		t.Error("OnClick handler should have been called")
	}
}

// TestButtonBuilder tests Button builder pattern
func TestButtonBuilder(t *testing.T) {
	handler := func() {
		// Handler for testing
	}

	btn := ButtonBuilder("Click").
		OnClick(handler).
		Disabled(true).
		Build()

	// Check it's a ButtonVNode
	if _, ok := btn.(*ButtonVNode); !ok {
		t.Error("Build() should return *ButtonVNode")
	}

	// Check properties
	b := btn.(*ButtonVNode)
	if !b.Disabled() {
		t.Error("Disabled() should be true")
	}

	if b.OnClick() == nil {
		t.Error("OnClick() should not be nil")
	}
}

// TestComponentVNode tests ComponentVNode
func TestComponentVNode(t *testing.T) {
	renderFn := func() VNode {
		return NewText("Component Content")
	}
	comp := NewComponent("TestComponent", renderFn)

	if got := comp.Render(); got == nil {
		t.Error("Render() should not return nil")
	}

	// Check rendered content
	text, ok := comp.Render().(*TextVNode)
	if !ok {
		t.Error("Render() should return *TextVNode")
	}
	if text.Content() != "Component Content" {
		t.Errorf("Content() = %v, want 'Component Content'", text.Content())
	}
}

// TestFragmentVNode tests FragmentVNode
func TestFragmentVNode(t *testing.T) {
	child1 := NewText("child1")
	child2 := NewText("child2")

	frag := NewFragment(child1, child2)

	children := frag.Children()
	if got := len(children); got != 2 {
		t.Errorf("len(Children()) = %v, want 2", got)
	}
}

// TestPropsMerge tests Props merging
func TestPropsMerge(t *testing.T) {
	p1 := Props{"a": 1, "b": 2}
	p2 := Props{"b": 3, "c": 4}

	merged := p1.Merge(p2)

	// p1 and p2 should not be modified
	if p1["b"] != 2 {
		t.Errorf("p1[b] = %v, want 2 (not modified)", p1["b"])
	}

	// merged should have all keys
	if merged["a"] != 1 {
		t.Errorf("merged[a] = %v, want 1", merged["a"])
	}
	if merged["b"] != 3 {
		t.Errorf("merged[b] = %v, want 3 (p2 overwrites p1)", merged["b"])
	}
	if merged["c"] != 4 {
		t.Errorf("merged[c] = %v, want 4", merged["c"])
	}
}

// TestPropsClone tests Props cloning
func TestPropsClone(t *testing.T) {
	original := Props{"a": 1, "b": 2}
	cloned := original.Clone()

	// Modify clone
	cloned["a"] = 999

	// Original should not be modified
	if original["a"] != 1 {
		t.Errorf("original[a] = %v, want 1 (not modified by clone)", original["a"])
	}

	// Clone should have new value
	if cloned["a"] != 999 {
		t.Errorf("cloned[a] = %v, want 999", cloned["a"])
	}
}

// TestPropsGetVariants tests Props Get* methods
func TestPropsGetVariants(t *testing.T) {
	p := Props{
		"str":    "hello",
		"num":    42,
		"flag":   true,
		"missing": nil,
	}

	// GetString
	if got := p.GetString("str"); got != "hello" {
		t.Errorf("GetString(str) = %v, want hello", got)
	}
	if got := p.GetString("missing"); got != "" {
		t.Errorf("GetString(missing) = %v, want empty", got)
	}

	// GetInt
	if got := p.GetInt("num"); got != 42 {
		t.Errorf("GetInt(num) = %v, want 42", got)
	}
	if got := p.GetInt("missing"); got != 0 {
		t.Errorf("GetInt(missing) = %v, want 0", got)
	}

	// GetBool
	if got := p.GetBool("flag"); got != true {
		t.Errorf("GetBool(flag) = %v, want true", got)
	}
	if got := p.GetBool("missing"); got != false {
		t.Errorf("GetBool(missing) = %v, want false", got)
	}
}

// BenchmarkVNodeCreate benchmarks VNode creation
func BenchmarkVNodeCreate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewElement("div")
	}
}

// BenchmarkTextCreate benchmarks TextVNode creation
func BenchmarkTextCreate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewText("Hello")
	}
}

// BenchmarkTextBuilder benchmarks Text builder
func BenchmarkTextBuilder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewTextBuilder("Hello").FgColor("red").Build()
	}
}
