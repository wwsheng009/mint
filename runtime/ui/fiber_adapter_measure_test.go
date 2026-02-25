package ui_test

import (
	"testing"

	"github.com/wwsheng009/mint/components/basic"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/layout"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	newbutton "github.com/wwsheng009/mint/ui/components/button"
	newtext "github.com/wwsheng009/mint/ui/components/text"
)

// TestFiberToNodeAdapter_TextNode tests that Fiber correctly captures text node info
func TestFiberToNodeAdapter_TextNode(t *testing.T) {
	// Create a text VNode
	textContent := "Hello, World!"
	textVNode := basic.NewText(textContent)

	// Verify VNode properties
	if textVNode.Type() != rtui.VNodeText {
		t.Errorf("TextVNode.Type() = %d, want %d (VNodeText)", textVNode.Type(), rtui.VNodeText)
	}
	if textVNode.Content() != textContent {
		t.Errorf("TextVNode.Content() = %q, want %q", textVNode.Content(), textContent)
	}

	// Test Measure on VNode
	size := textVNode.Measure(runtime.BoxConstraints{})
	expectedWidth := len([]rune(textContent))
	if size.Width != expectedWidth || size.Height != 1 {
		t.Errorf("TextVNode.Measure() = %dx%d, want %dx1", size.Width, size.Height, expectedWidth)
	}

	// Create Fiber from VNode
	fiber := rtui.CreateFiberFromVNode(textVNode)

	// Verify Fiber captured Type correctly
	if fiber.Type != rtui.VNodeText {
		t.Errorf("Fiber.Type = %d, want %d (VNodeText)", fiber.Type, rtui.VNodeText)
	}

	// Verify Fiber captured text content in MemoizedState
	if fiber.MemoizedState == nil {
		t.Error("Fiber.MemoizedState is nil, expected text content")
	} else if content, ok := fiber.MemoizedState.(string); !ok {
		t.Errorf("Fiber.MemoizedState type = %T, expected string", fiber.MemoizedState)
	} else if content != textContent {
		t.Errorf("Fiber.MemoizedState = %q, want %q", content, textContent)
	}
}

// TestFiberToNodeAdapter_NewButton tests the new ui/components/button with Measure
func TestFiberToNodeAdapter_NewButton(t *testing.T) {
	// Create a new button VNode
	buttonVNode := newbutton.New("Click Me")

	// Verify VNode properties
	if buttonVNode.Type() != rtui.VNodeElement {
		t.Errorf("ButtonVNode.Type() = %d, want %d (VNodeElement)", buttonVNode.Type(), rtui.VNodeElement)
	}
	if buttonVNode.Tag() != "button" {
		t.Errorf("ButtonVNode.Tag() = %q, want %q", buttonVNode.Tag(), "button")
	}

	// Create Fiber from VNode
	fiber := rtui.CreateFiberFromVNode(buttonVNode)

	// Verify Fiber properties
	if fiber.Type != rtui.VNodeElement {
		t.Errorf("Fiber.Type = %d, want %d (VNodeElement)", fiber.Type, rtui.VNodeElement)
	}
	if fiber.Tag != "button" {
		t.Errorf("Fiber.Tag = %q, want %q", fiber.Tag, "button")
	}

	// Verify Instance was created
	if fiber.Instance == nil {
		t.Fatal("Fiber.Instance is nil, expected ButtonInstance")
	}

	// Test Instance Measure method directly
	instance := fiber.Instance.(*newbutton.Instance)
	constraints := layout.Constraints{MinWidth: 0, MaxWidth: 100, MinHeight: 0, MaxHeight: 10}
	size := instance.Measure(constraints)

	// "Click Me" = 8 chars + 3 (brackets + focus) + 2 (medium padding) = 13
	if size.Width == 0 || size.Height == 0 {
		t.Errorf("Instance.Measure() = %dx%d, expected non-zero", size.Width, size.Height)
	}
	t.Logf("Button Instance.Measure() = %dx%d", size.Width, size.Height)
}

// TestFiberToNodeAdapter_NewButtonWithVariant tests button variants
func TestFiberToNodeAdapter_NewButtonWithVariant(t *testing.T) {
	tests := []struct {
		name    string
		variant newbutton.Variant
		size    newbutton.Size
	}{
		{"Default/Small", newbutton.VariantDefault, newbutton.SizeSmall},
		{"Primary/Medium", newbutton.VariantPrimary, newbutton.SizeMedium},
		{"Danger/Large", newbutton.VariantDanger, newbutton.SizeLarge},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			btn := newbutton.New("Test").
				SetVariant(tt.variant).
				SetSize(tt.size)

			fiber := rtui.CreateFiberFromVNode(btn)

			if fiber.Instance == nil {
				t.Fatal("Fiber.Instance is nil")
			}

			instance := fiber.Instance.(*newbutton.Instance)
			constraints := layout.UnboundedConstraints()
			size := instance.Measure(constraints)

			t.Logf("Button variant=%v, size=%v, measured=%dx%d",
				tt.variant, tt.size, size.Width, size.Height)

			if size.Width == 0 || size.Height == 0 {
				t.Errorf("Button should have non-zero size")
			}
		})
	}
}

// TestFiberToNodeAdapter_ButtonSizeCalculation tests precise size calculations
func TestFiberToNodeAdapter_ButtonSizeCalculation(t *testing.T) {
	tests := []struct {
		name       string
		label      string
		size       newbutton.Size
		wantWidth  int
		wantHeight int
	}{
		// Width = len(label) + 3 (brackets + focus) + size modifier
		// Small: +0, Medium: +2, Large: +4
		{"Small OK", "OK", newbutton.SizeSmall, 5, 1},            // 2 + 3 = 5
		{"Medium OK", "OK", newbutton.SizeMedium, 7, 1},          // 2 + 3 + 2 = 7
		{"Large OK", "OK", newbutton.SizeLarge, 9, 1},            // 2 + 3 + 4 = 9
		{"Small Submit", "Submit", newbutton.SizeSmall, 9, 1},    // 6 + 3 = 9
		{"Medium Submit", "Submit", newbutton.SizeMedium, 11, 1}, // 6 + 3 + 2 = 11
		{"Large Submit", "Submit", newbutton.SizeLarge, 13, 1},   // 6 + 3 + 4 = 13
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			btn := newbutton.New(tt.label).SetSize(tt.size)
			fiber := rtui.CreateFiberFromVNode(btn)

			instance := fiber.Instance.(*newbutton.Instance)
			size := instance.Measure(layout.UnboundedConstraints())

			if size.Width != tt.wantWidth {
				t.Errorf("Width = %d, want %d", size.Width, tt.wantWidth)
			}
			if size.Height != tt.wantHeight {
				t.Errorf("Height = %d, want %d", size.Height, tt.wantHeight)
			}
		})
	}
}

// TestFiberToNodeAdapter_ButtonWithConstraints tests constraint application
func TestFiberToNodeAdapter_ButtonWithConstraints(t *testing.T) {
	btn := newbutton.New("Click Me") // natural width = 13
	fiber := rtui.CreateFiberFromVNode(btn)
	instance := fiber.Instance.(*newbutton.Instance)

	tests := []struct {
		name       string
		constraint layout.Constraints
		wantWidth  int
	}{
		{"Unbounded", layout.UnboundedConstraints(), 13},
		{"Tight 20", layout.TightConstraints(20, 1), 20},
		{"Max 10", layout.Constraints{MinWidth: 0, MaxWidth: 10, MinHeight: 0, MaxHeight: 5}, 10}, // constrained
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size := instance.Measure(tt.constraint)
			if size.Width != tt.wantWidth {
				t.Errorf("Width = %d, want %d", size.Width, tt.wantWidth)
			}
		})
	}
}

// TestFiberTree_Structure tests building a complete Fiber tree
func TestFiberTree_Structure(t *testing.T) {
	// Create a simple structure with new button
	title := basic.NewText("Title")
	okBtn := newbutton.New("OK")
	cancelBtn := newbutton.New("Cancel")

	// Create Fiber trees
	titleFiber := rtui.CreateFiberFromVNode(title)
	okFiber := rtui.CreateFiberFromVNode(okBtn)
	cancelFiber := rtui.CreateFiberFromVNode(cancelBtn)

	// Verify title is text node
	if titleFiber.Type != rtui.VNodeText {
		t.Errorf("Title Fiber.Type = %d, want %d (VNodeText)", titleFiber.Type, rtui.VNodeText)
	}
	if titleFiber.MemoizedState != "Title" {
		t.Errorf("Title Fiber.MemoizedState = %v, want %q", titleFiber.MemoizedState, "Title")
	}

	// Verify buttons have instances
	if okFiber.Instance == nil {
		t.Error("OK Button should have Instance")
	}
	if cancelFiber.Instance == nil {
		t.Error("Cancel Button should have Instance")
	}

	// Verify button tags
	if okFiber.Tag != "button" {
		t.Errorf("OK Button Fiber.Tag = %q, want %q", okFiber.Tag, "button")
	}
	if cancelFiber.Tag != "button" {
		t.Errorf("Cancel Button Fiber.Tag = %q, want %q", cancelFiber.Tag, "button")
	}
}

// TestFiber_InstanceProperties tests that Fiber captures all button properties
func TestFiber_InstanceProperties(t *testing.T) {
	btn := newbutton.New("Test Button").
		SetVariant(newbutton.VariantPrimary).
		SetSize(newbutton.SizeLarge).
		SetDisabled(true)

	fiber := rtui.CreateFiberFromVNode(btn)

	if fiber.Instance == nil {
		t.Fatal("Fiber.Instance is nil")
	}

	inst := fiber.Instance.(*newbutton.Instance)

	// Verify disabled state was transferred
	if !inst.IsDisabled() {
		t.Error("Disabled = false, want true")
	}

	// Verify we can get props back
	props := inst.GetProps()
	if props["label"] != "Test Button" {
		t.Errorf("Label = %q, want %q", props["label"], "Test Button")
	}
	if props["variant"] != newbutton.VariantPrimary {
		t.Errorf("Variant = %v, want Primary", props["variant"])
	}
	if props["size"] != newbutton.SizeLarge {
		t.Errorf("Size = %v, want Large", props["size"])
	}
}

// =============================================================================
// New Text Component Tests (ui/components/text)
// =============================================================================

// TestFiberToNodeAdapter_NewText tests the new ui/components/text with Measure
func TestFiberToNodeAdapter_NewText(t *testing.T) {
	// Create a new text VNode
	textVNode := newtext.New("Hello, World!")

	// Verify VNode properties
	if textVNode.Type() != rtui.VNodeElement {
		t.Errorf("TextVNode.Type() = %d, want %d (VNodeElement)", textVNode.Type(), rtui.VNodeElement)
	}
	if textVNode.Tag() != "text" {
		t.Errorf("TextVNode.Tag() = %q, want %q", textVNode.Tag(), "text")
	}
	if textVNode.Content() != "Hello, World!" {
		t.Errorf("TextVNode.Content() = %q, want %q", textVNode.Content(), "Hello, World!")
	}

	// Create Fiber from VNode
	fiber := rtui.CreateFiberFromVNode(textVNode)

	// Verify Fiber properties
	if fiber.Type != rtui.VNodeElement {
		t.Errorf("Fiber.Type = %d, want %d (VNodeElement)", fiber.Type, rtui.VNodeElement)
	}
	if fiber.Tag != "text" {
		t.Errorf("Fiber.Tag = %q, want %q", fiber.Tag, "text")
	}

	// Verify Instance was created
	if fiber.Instance == nil {
		t.Fatal("Fiber.Instance is nil, expected TextInstance")
	}

	// Test Instance Measure method directly
	instance := fiber.Instance.(*newtext.Instance)
	constraints := layout.Constraints{MinWidth: 0, MaxWidth: 100, MinHeight: 0, MaxHeight: 10}
	size := instance.Measure(constraints)

	// "Hello, World!" = 13 chars
	if size.Width != 13 {
		t.Errorf("Instance.Measure() width = %d, want 13", size.Width)
	}
	if size.Height != 1 {
		t.Errorf("Instance.Measure() height = %d, want 1", size.Height)
	}
}

// TestFiberToNodeAdapter_NewTextWithPadding tests text with padding
func TestFiberToNodeAdapter_NewTextWithPadding(t *testing.T) {
	textVNode := newtext.New("Hi").
		SetPaddingProps(0, 2, 0, 1) // right=2, left=1

	fiber := rtui.CreateFiberFromVNode(textVNode)

	if fiber.Instance == nil {
		t.Fatal("Fiber.Instance is nil")
	}

	inst := fiber.Instance.(*newtext.Instance)
	size := inst.Measure(layout.UnboundedConstraints())

	// "Hi" = 2 + left padding 1 + right padding 2 = 5
	if size.Width != 5 {
		t.Errorf("Width = %d, want 5", size.Width)
	}
	if size.Height != 1 {
		t.Errorf("Height = %d, want 1", size.Height)
	}
}

// TestFiberToNodeAdapter_NewTextWithMaxWidth tests text with max width
func TestFiberToNodeAdapter_NewTextWithMaxWidth(t *testing.T) {
	textVNode := newtext.New("This is a very long text").SetMaxWidth(10)

	fiber := rtui.CreateFiberFromVNode(textVNode)

	if fiber.Instance == nil {
		t.Fatal("Fiber.Instance is nil")
	}

	inst := fiber.Instance.(*newtext.Instance)
	size := inst.Measure(layout.UnboundedConstraints())

	// Max width should limit the text
	if size.Width != 10 {
		t.Errorf("Width = %d, want 10 (max width)", size.Width)
	}
}

// TestFiberToNodeAdapter_NewTextChinese tests Chinese text
func TestFiberToNodeAdapter_NewTextChinese(t *testing.T) {
	textVNode := newtext.New("你好世界")

	fiber := rtui.CreateFiberFromVNode(textVNode)

	if fiber.Instance == nil {
		t.Fatal("Fiber.Instance is nil")
	}

	inst := fiber.Instance.(*newtext.Instance)
	size := inst.Measure(layout.UnboundedConstraints())

	// "你好世界" = 4 Chinese chars, each with display width 2 = 8
	if size.Width != 8 {
		t.Errorf("Width = %d, want 8 (Chinese chars have display width 2)", size.Width)
	}
}

// TestFiberToNodeAdapter_NewTextWithStyle tests text with style dimensions
func TestFiberToNodeAdapter_NewTextWithStyle(t *testing.T) {
	textVNode := newtext.New("Short").
		Bold(true).
		Foreground("red")

	fiber := rtui.CreateFiberFromVNode(textVNode)

	if fiber.Instance == nil {
		t.Fatal("Fiber.Instance is nil")
	}

	inst := fiber.Instance.(*newtext.Instance)

	// Verify style was transferred
	style := inst.GetStyle()
	if !style.IsBold() {
		t.Error("Bold = false, want true")
	}
	if style.FG != "red" {
		t.Errorf("FG = %q, want %q", style.FG, "red")
	}

	// Measure should work correctly
	size := inst.Measure(layout.UnboundedConstraints())
	if size.Width != 5 { // "Short" = 5 chars
		t.Errorf("Width = %d, want 5", size.Width)
	}
}
