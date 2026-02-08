package inspector

import (
	"testing"

	"github.com/wwsheng009/mint/app"
)

// TestNewSidebar tests creating a new sidebar
func TestNewSidebar(t *testing.T) {
	sidebar := NewSidebar()

	if sidebar == nil {
		t.Fatal("Expected non-nil Sidebar")
	}

	if !sidebar.IsEnabled() {
		t.Error("New sidebar should be enabled by default")
	}

	if sidebar.width != 40 {
		t.Errorf("Expected default width 40, got %d", sidebar.width)
	}

	if sidebar.height != 20 {
		t.Errorf("Expected default height 20, got %d", sidebar.height)
	}
}

// TestSidebarEnableDisable tests enabling and disabling sidebar
func TestSidebarEnableDisable(t *testing.T) {
	sidebar := NewSidebar()

	// Test disable
	sidebar.Disable()
	if sidebar.IsEnabled() {
		t.Error("Sidebar should be disabled after Disable()")
	}

	// Test enable
	sidebar.Enable()
	if !sidebar.IsEnabled() {
		t.Error("Sidebar should be enabled after Enable()")
	}
}

// TestSetWidth tests setting sidebar width
func TestSetWidth(t *testing.T) {
	sidebar := NewSidebar()

	sidebar.SetWidth(60)
	if sidebar.width != 60 {
		t.Errorf("Expected width 60, got %d", sidebar.width)
	}
}

// TestSetHeight tests setting sidebar height
func TestSetHeight(t *testing.T) {
	sidebar := NewSidebar()

	sidebar.SetHeight(30)
	if sidebar.height != 30 {
		t.Errorf("Expected height 30, got %d", sidebar.height)
	}
}

// TestToggleSection tests toggling section collapse state
func TestToggleSection(t *testing.T) {
	sidebar := NewSidebar()

	// Initially not collapsed
	if sidebar.collapsed["type"] {
		t.Error("Section should not be collapsed initially")
	}

	// Toggle to collapsed
	sidebar.ToggleSection("type")
	if !sidebar.collapsed["type"] {
		t.Error("Section should be collapsed after toggle")
	}

	// Toggle back to expanded
	sidebar.ToggleSection("type")
	if sidebar.collapsed["type"] {
		t.Error("Section should not be collapsed after second toggle")
	}
}

// TestSetShowPaths tests show paths control
func TestSetShowPaths(t *testing.T) {
	sidebar := NewSidebar()

	sidebar.SetShowPaths(false)
	if sidebar.showPaths {
		t.Error("showPaths should be false")
	}

	sidebar.SetShowPaths(true)
	if !sidebar.showPaths {
		t.Error("showPaths should be true")
	}
}

// TestSetShowProps tests show properties control
func TestSetShowProps(t *testing.T) {
	sidebar := NewSidebar()

	sidebar.SetShowProps(false)
	if sidebar.showProps {
		t.Error("showProps should be false")
	}

	sidebar.SetShowProps(true)
	if !sidebar.showProps {
		t.Error("showProps should be true")
	}
}

// TestFormatSidebar tests sidebar formatting
func TestFormatSidebar(t *testing.T) {
	sidebar := NewSidebar()

	button := app.ButtonBuilder("Test Button").Build()
	info := ExtractElementInfo(button)

	// Set bounds for complete info
	if boundsAware, ok := button.(interface{ SetBounds(int, int, int, int) }); ok {
		boundsAware.SetBounds(10, 5, 18, 1)
		info.Bounds = [4]int{10, 5, 18, 1}
		info.Position.X = 10
		info.Position.Y = 5
		info.Size.Width = 18
		info.Size.Height = 1
	}

	output := sidebar.FormatSidebar(info)

	if output == "" {
		t.Error("Expected non-empty sidebar output")
	}

	// Check for key sections
	requiredSections := []string{
		"UI Inspector",
		"Element:",
		"Type",
		"Position",
		"Size",
		"Layout",
		"Bounds",
		"Constraints",
	}

	for _, section := range requiredSections {
		if !contains(output, section) {
			t.Errorf("Output should contain section: %s", section)
		}
	}
}

// TestFormatSidebar_Collapsed tests sidebar with collapsed sections
func TestFormatSidebar_Collapsed(t *testing.T) {
	sidebar := NewSidebar()

	button := app.ButtonBuilder("Test").Build()
	info := ExtractElementInfo(button)

	// Collapse the type section
	sidebar.ToggleSection("type")

	output := sidebar.FormatSidebar(info)

	if !contains(output, "(collapsed)") {
		t.Error("Output should show collapsed section")
	}

	if !contains(output, "+ Type (collapsed)") {
		t.Error("Type section should be marked as collapsed")
	}
}

// TestFormatCompact tests compact formatting
func TestFormatCompact(t *testing.T) {
	sidebar := NewSidebar()

	button := app.ButtonBuilder("Test Button").Build()
	info := ExtractElementInfo(button)

	// Set position and size
	info.Position.X = 10
	info.Position.Y = 5
	info.Size.Width = 18
	info.Size.Height = 1

	output := sidebar.FormatCompact(info)

	if output == "" {
		t.Error("Expected non-empty compact output")
	}

	// Check for key information
	requiredParts := []string{
		"ButtonVNode",
		"Test Button",
		"@(10,5)",
		"18x1",
	}

	for _, part := range requiredParts {
		if !contains(output, part) {
			t.Errorf("Output should contain: %s", part)
		}
	}
}

// TestFormatCompact_WithFlex tests compact formatting with flex info
func TestFormatCompact_WithFlex(t *testing.T) {
	sidebar := NewSidebar()

	button := app.ButtonBuilder("Test").Build()
	info := ExtractElementInfo(button)

	// Add flex info
	info.Layout.Flex = 1
	info.Layout.NaturalWidth = 14
	info.Layout.LayoutWidth = 18

	output := sidebar.FormatCompact(info)

	if !contains(output, "flex:1") {
		t.Error("Output should contain flex information")
	}

	if !contains(output, "nat:14->18") {
		t.Error("Output should contain natural width expansion")
	}
}

// TestFormatTable tests table formatting for multiple elements
func TestFormatTable(t *testing.T) {
	sidebar := NewSidebar()

	button1 := app.ButtonBuilder("Button1").Build()
	button2 := app.ButtonBuilder("Button2").Build()

	info1 := ExtractElementInfo(button1)
	info2 := ExtractElementInfo(button2)

	elements := []ElementInfo{info1, info2}

	output := sidebar.FormatTable(elements)

	if output == "" {
		t.Error("Expected non-empty table output")
	}

	// Check for table structure
	if !contains(output, "┌─") || !contains(output, "└") {
		t.Error("Output should have table borders")
	}

	if !contains(output, "ButtonVNode") {
		t.Error("Output should contain element types")
	}
}

// TestFormatTable_Empty tests table formatting with no elements
func TestFormatTable_Empty(t *testing.T) {
	sidebar := NewSidebar()

	output := sidebar.FormatTable([]ElementInfo{})

	if output != "No elements" {
		t.Errorf("Expected 'No elements', got '%s'", output)
	}
}

// TestGetCopyableText tests copyable text generation
func TestGetCopyableText(t *testing.T) {
	sidebar := NewSidebar()

	button := app.ButtonBuilder("Test Button").Build()
	info := ExtractElementInfo(button)

	// Set various info
	info.Position.X = 10
	info.Position.Y = 5
	info.Size.Width = 18
	info.Size.Height = 1
	info.Bounds = [4]int{10, 5, 18, 1}
	info.Constraints.MinWidth = 18
	info.Constraints.MaxWidth = 18
	info.Constraints.MinHeight = 0
	info.Constraints.MaxHeight = 1

	output := sidebar.GetCopyableText(info)

	if output == "" {
		t.Error("Expected non-empty copyable text")
	}

	// Check for sections
	requiredSections := []string{
		"=== UI Inspector Element Info ===",
		"Type:",
		"Position:",
		"Size:",
		"Bounds:",
		"Constraints:",
	}

	for _, section := range requiredSections {
		if !contains(output, section) {
			t.Errorf("Output should contain section: %s", section)
		}
	}
}

// TestGetCopyableText_WithProperties tests copyable text with properties
func TestGetCopyableText_WithProperties(t *testing.T) {
	sidebar := NewSidebar()

	button := app.ButtonBuilder("Test").Build()
	info := ExtractElementInfo(button)

	// Add properties
	info.Properties = map[string]interface{}{
		"Label":     "Test",
		"HasFocus":  true,
		"Disabled":  false,
		"onClick":   "handler",
	}

	output := sidebar.GetCopyableText(info)

	if !contains(output, "Properties:") {
		t.Error("Output should contain properties section")
	}

	if !contains(output, "Label: Test") {
		t.Error("Output should contain Label property")
	}

	if !contains(output, "HasFocus: true") {
		t.Error("Output should contain HasFocus property")
	}
}

// TestBuildVNode tests building VNode from sidebar
func TestBuildVNode(t *testing.T) {
	sidebar := NewSidebar()

	button := app.ButtonBuilder("Test").Build()
	info := ExtractElementInfo(button)

	vnode := sidebar.BuildVNode(info)

	if vnode == nil {
		t.Fatal("Expected non-nil VNode")
	}

	// VNode should not be empty text
	// (We can't easily test the actual content without rendering)
}

// TestBuildVNode_Disabled tests building VNode when sidebar is disabled
func TestBuildVNode_Disabled(t *testing.T) {
	sidebar := NewSidebar()
	sidebar.Disable()

	button := app.ButtonBuilder("Test").Build()
	info := ExtractElementInfo(button)

	vnode := sidebar.BuildVNode(info)

	if vnode == nil {
		t.Fatal("Expected non-nil VNode")
	}

	// Should be empty text when disabled
	// (We can't easily test this without checking the text content)
}

// TestBuildVNode_EmptyInfo tests building VNode with empty info
func TestBuildVNode_EmptyInfo(t *testing.T) {
	sidebar := NewSidebar()

	info := ElementInfo{}

	vnode := sidebar.BuildVNode(info)

	if vnode == nil {
		t.Fatal("Expected non-nil VNode")
	}
}

// TestBuildCompactVNode tests building compact VNode
func TestBuildCompactVNode(t *testing.T) {
	sidebar := NewSidebar()

	button := app.ButtonBuilder("Test Button").Build()
	info := ExtractElementInfo(button)

	info.Position.X = 10
	info.Position.Y = 5

	vnode := sidebar.BuildCompactVNode(info)

	if vnode == nil {
		t.Fatal("Expected non-nil VNode")
	}
}

// TestFormatTruncate tests string truncation
func TestFormatTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{"No truncation", "hello", 10, "hello"},
		{"Exact length", "hello", 5, "hello"},
		{"Needs truncation", "hello world", 8, "hello..."},
		{"Empty string", "", 10, ""},
		{"Very short", "hi", 10, "hi"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTruncate(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestMax tests max helper function
func TestMax(t *testing.T) {
	tests := []struct {
		a, b, expected int
	}{
		{5, 3, 5},
		{3, 5, 5},
		{5, 5, 5},
		{0, 0, 0},
		{-1, 1, 1},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := max(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("max(%d, %d) = %d, expected %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}
