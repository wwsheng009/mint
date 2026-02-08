package inspector

import (
	"testing"

	"github.com/wwsheng009/mint/app"
)

// TestExtractElementInfo_Button tests extracting info from a Button
func TestExtractElementInfo_Button(t *testing.T) {
	button := app.ButtonBuilder("[1] Event").Build()

	info := ExtractElementInfo(button)

	// Check basic info
	if info.Type == "" {
		t.Error("Type should not be empty")
	}
	if info.Tag != "button" {
		t.Errorf("Expected tag 'button', got '%s'", info.Tag)
	}
	if info.Label != "[1] Event" {
		t.Errorf("Expected label '[1] Event', got '%s'", info.Label)
	}
}

// TestExtractElementInfo_Text tests extracting info from Text
func TestExtractElementInfo_Text(t *testing.T) {
	text := app.Text("Hello World")

	info := ExtractElementInfo(text)

	// Check basic info
	if info.Type == "" {
		t.Error("Type should not be empty")
	}
	if info.Label != "Hello World" {
		t.Errorf("Expected label 'Hello World', got '%s'", info.Label)
	}

	// Check natural width
	if info.Layout.NaturalWidth != 11 { // "Hello World" = 11 chars
		t.Errorf("Expected natural width 11, got %d", info.Layout.NaturalWidth)
	}
}

// TestExtractElementInfo_WithBounds tests extracting info with bounds
func TestExtractElementInfo_WithBounds(t *testing.T) {
	t.Skip("SetBounds integration needs actual layout engine")

	button := app.ButtonBuilder("Test").Build()

	// Simulate SetBounds being called
	if boundsAware, ok := button.(interface{ SetBounds(int, int, int, int) }); ok {
		boundsAware.SetBounds(5, 10, 20, 1)
	}

	info := ExtractElementInfo(button)

	// Check bounds
	if info.Bounds[0] != 5 || info.Bounds[1] != 10 ||
	   info.Bounds[2] != 20 || info.Bounds[3] != 1 {
		t.Errorf("Expected bounds [5 10 20 1], got %v", info.Bounds)
	}

	// Check position
	if info.Position.X != 5 || info.Position.Y != 10 {
		t.Errorf("Expected position (5, 10), got (%d, %d)",
			info.Position.X, info.Position.Y)
	}

	// Check size
	if info.Size.Width != 20 || info.Size.Height != 1 {
		t.Errorf("Expected size 20x1, got %dx%d",
			info.Size.Width, info.Size.Height)
	}

	// Check layout width
	if info.Layout.LayoutWidth != 20 {
		t.Errorf("Expected layout width 20, got %d", info.Layout.LayoutWidth)
	}
}

// TestExtractElementInfo_Flex tests flex property extraction
func TestExtractElementInfo_Flex(t *testing.T) {
	t.Skip("SetProp integration needs proper props handling")

	// Create a button with flex prop through ElementVNode
	button := app.ButtonBuilder("Test").Build()

	// Get the element and set prop
	if elem, ok := button.(interface{ SetProp(string, interface{}) }); ok {
		elem.SetProp("flex", 1)
	}

	info := ExtractElementInfo(button)

	// Check flex
	if info.Layout.Flex != 1 {
		t.Errorf("Expected flex 1, got %d", info.Layout.Flex)
	}

	if !info.Layout.IsFlexChild {
		t.Error("Expected IsFlexChild to be true")
	}
}

// TestExtractElementInfo_NilVNode tests handling of nil VNode
func TestExtractElementInfo_NilVNode(t *testing.T) {
	info := ExtractElementInfo(nil)

	if info.Type != "nil" {
		t.Errorf("Expected type 'nil', got '%s'", info.Type)
	}

	if info.VNode != nil {
		t.Error("Expected VNode to be nil")
	}
}

// TestFormatElementInfo tests formatting of ElementInfo
func TestFormatElementInfo(t *testing.T) {
	button := app.ButtonBuilder("Test Button").Build()

	info := ExtractElementInfo(button)
	formatted := FormatElementInfo(info)

	if formatted == "" {
		t.Error("Formatted output should not be empty")
	}

	// Check for expected sections (without Bounds since we haven't set it)
	expectedSections := []string{
		"Element:",
		"Tag:",
		"Position:",
		"Size:",
		"Layout:",
	}

	for _, section := range expectedSections {
		if !contains(formatted, section) {
			t.Errorf("Formatted output should contain '%s'", section)
		}
	}
}

// TestExtractElementInfo_NaturalWidthCalculation tests natural width calculation
func TestExtractElementInfo_NaturalWidthCalculation(t *testing.T) {
	tests := []struct {
		name           string
		label          string
		expectedWidth  int
	}{
		{"Short label", "OK", 6},     // "OK" (2) + brackets (2) + focus space (2)
		{"Medium label", "Cancel", 10}, // "Cancel" (6) + brackets (2) + focus space (2)
		{"Long label", "Submit Form", 15}, // "Submit Form" (11) + brackets (2) + focus space (2)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			button := app.ButtonBuilder(tt.label).Build()
			info := ExtractElementInfo(button)

			if info.Layout.NaturalWidth != tt.expectedWidth {
				t.Errorf("Expected natural width %d, got %d",
					tt.expectedWidth, info.Layout.NaturalWidth)
			}
		})
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
