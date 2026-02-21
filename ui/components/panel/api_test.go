package panel

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	newtext "github.com/wwsheng009/mint/ui/components/text"
)

// ============================================================================
// Phase 2 Tests - API Enhancements
// ============================================================================

// =============================================================================
// VNode API Enhancement Tests
// =============================================================================

func TestVNode_SetOuterWidth_BackwardCompatibility(t *testing.T) {
	// SetOuterWidth should be an alias for SetWidth
	v := New()

	// Test with SetOuterWidth
	v.SetOuterWidth(20)
	if v.width != 20 {
		t.Errorf("Expected width 20, got %d", v.width)
	}

	// Test with SetWidth (old API)
	v.SetWidth(30)
	if v.width != 30 {
		t.Errorf("Expected width 30, got %d", v.width)
	}
}

func TestVNode_SetInnerWidth(t *testing.T) {
	tests := []struct {
		name           string
		borderStyle    layout.BorderStyle
		innerWidth     int
		expectedOuter  int
	}{
		{
			name:           "Single border",
			borderStyle:    layout.BorderSingle,
			innerWidth:     20,
			expectedOuter:  22, // 20 + 2
		},
		{
			name:           "Rounded border",
			borderStyle:    layout.BorderRounded,
			innerWidth:     20,
			expectedOuter:  22, // 20 + 2
		},
		{
			name:           "Double border",
			borderStyle:    layout.BorderDouble,
			innerWidth:     20,
			expectedOuter:  22, // 20 + 2
		},
		{
			name:           "No border",
			borderStyle:    layout.BorderNone,
			innerWidth:     20,
			expectedOuter:  20, // 20 + 0
		},
		{
			name:           "Zero inner width",
			borderStyle:    layout.BorderSingle,
			innerWidth:     0,
			expectedOuter:  2, // 0 + 2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New().SetBorderStyle(tt.borderStyle)
			v.SetInnerWidth(tt.innerWidth)

			if v.width != tt.expectedOuter {
				t.Errorf("Expected outer width %d, got %d", tt.expectedOuter, v.width)
			}

			// Composed should be invalidated
			if v.composed != nil {
				t.Error("Composed should be nil after dimension change")
			}
		})
	}
}

func TestVNode_SetOuterHeight_BackwardCompatibility(t *testing.T) {
	// SetOuterHeight should be an alias for SetHeight
	v := New()

	v.SetOuterHeight(10)
	if v.height != 10 {
		t.Errorf("Expected height 10, got %d", v.height)
	}
}

func TestVNode_SetInnerHeight(t *testing.T) {
	tests := []struct {
		name            string
		borderStyle     layout.BorderStyle
		innerHeight     int
		expectedOuter   int
	}{
		{
			name:           "Single border",
			borderStyle:    layout.BorderSingle,
			innerHeight:    5,
			expectedOuter:  7, // 5 + 2
		},
		{
			name:           "Rounded border",
			borderStyle:    layout.BorderRounded,
			innerHeight:    5,
			expectedOuter:  7, // 5 + 2
		},
		{
			name:           "No border",
			borderStyle:    layout.BorderNone,
			innerHeight:    5,
			expectedOuter:  5, // 5 + 0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New().SetBorderStyle(tt.borderStyle)
			v.SetInnerHeight(tt.innerHeight)

			if v.height != tt.expectedOuter {
				t.Errorf("Expected outer height %d, got %d", tt.expectedOuter, v.height)
			}
		})
	}
}

func TestVNode_SetContentWidth_Alias(t *testing.T) {
	v := New()
	v.SetContentWidth(20)

	borderPadding := 2 // default BorderSingle
	expected := 20 + borderPadding
	if v.width != expected {
		t.Errorf("Expected width %d, got %d", expected, v.width)
	}
}

func TestVNode_SetContentSize_Alias(t *testing.T) {
	v := New()
	v.SetContentSize(4)

	borderPadding := 2 // default BorderSingle
	expected := 4 + borderPadding
	if v.height != expected {
		t.Errorf("Expected height %d, got %d", expected, v.height)
	}
}

func TestVNode_SetWrappedTextContent(t *testing.T) {
	content := "This is a long text that should wrap"
	width := 20

	v := New().SetWrappedTextContent(content, width)

	// Width should include border padding
	borderPadding := 2 // default BorderSingle
	if v.width != width+borderPadding {
		t.Errorf("Expected width %d, got %d", width+borderPadding, v.width)
	}

	// Content should be a Text node with Wrap=true
	if v.content == nil {
		t.Fatal("Content should not be nil")
	}

	textNode, ok := v.content.(*newtext.VNode)
	if !ok {
		t.Fatal("Content should be a Text VNode")
	}

	// Check if it's wrapped (we can't directly check Wrap, but we set it)
	if textNode == nil {
		t.Error("Text node should not be nil")
	}
}

func TestVNode_SetTextContent(t *testing.T) {
	content := "Auto-sized content"
	v := New().SetTextContent(content)

	if v.content == nil {
		t.Fatal("Content should not be nil")
	}

	// Should be a Text node
	textNode, ok := v.content.(*newtext.VNode)
	if !ok {
		t.Fatal("Content should be a Text VNode")
	}

	if textNode == nil {
		t.Error("Text node should not be nil")
	}
}

func TestVNode_SetPlainContent(t *testing.T) {
	content := "Plain content"
	v := New().SetPlainContent(content)

	if v.content == nil {
		t.Fatal("Content should not be nil")
	}

	textNode, ok := v.content.(*newtext.VNode)
	if !ok {
		t.Fatal("Content should be a Text VNode")
	}

	if textNode == nil {
		t.Error("Text node should not be nil")
	}
}

func TestVNode_GetOuterDimensions(t *testing.T) {
	v := New().SetWidth(30).SetHeight(10)

	width, height := v.GetOuterDimensions()
	if width != 30 {
		t.Errorf("Expected width 30, got %d", width)
	}
	if height != 10 {
		t.Errorf("Expected height 10, got %d", height)
	}
}

func TestVNode_GetInnerDimensions(t *testing.T) {
	tests := []struct {
		name           string
		borderStyle    layout.BorderStyle
		outerWidth     int
		outerHeight    int
		expectedWidth  int
		expectedHeight int
	}{
		{
			name:           "Single border",
			borderStyle:    layout.BorderSingle,
			outerWidth:     22,
			outerHeight:    7,
			expectedWidth:  20, // 22 - 2
			expectedHeight: 5,  // 7 - 2
		},
		{
			name:           "Rounded border",
			borderStyle:    layout.BorderRounded,
			outerWidth:     22,
			outerHeight:    7,
			expectedWidth:  20,
			expectedHeight: 5,
		},
		{
			name:           "No border",
			borderStyle:    layout.BorderNone,
			outerWidth:     20,
			outerHeight:    5,
			expectedWidth:  20, // 20 - 0
			expectedHeight: 5,  // 5 - 0
		},
		{
			name:           "Smaller than padding",
			borderStyle:    layout.BorderSingle,
			outerWidth:     1,
			outerHeight:    1,
			expectedWidth:  0, // max(0, 1 - 2)
			expectedHeight: 0, // max(0, 1 - 2)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New().
				SetBorderStyle(tt.borderStyle).
				SetWidth(tt.outerWidth).
				SetHeight(tt.outerHeight)

			width, height := v.GetInnerDimensions()
			if width != tt.expectedWidth {
				t.Errorf("Expected inner width %d, got %d", tt.expectedWidth, width)
			}
			if height != tt.expectedHeight {
				t.Errorf("Expected inner height %d, got %d", tt.expectedHeight, height)
			}
		})
	}
}

func TestVNode_GetContentWidth_Height(t *testing.T) {
	v := New().SetWidth(22).SetHeight(7) // Single border by default

	width := v.GetContentWidth()
	height := v.GetContentHeight()

	if width != 20 {
		t.Errorf("Expected content width 20, got %d", width)
	}
	if height != 5 {
		t.Errorf("Expected content height 5, got %d", height)
	}
}

func TestVNode_GetBorderPadding(t *testing.T) {
	tests := []struct {
		name           string
		borderStyle    layout.BorderStyle
		expectedW      int
		expectedH      int
	}{
		{
			name:        "Single border",
			borderStyle: layout.BorderSingle,
			expectedW:   2,
			expectedH:   2,
		},
		{
			name:        "Rounded border",
			borderStyle: layout.BorderRounded,
			expectedW:   2,
			expectedH:   2,
		},
		{
			name:        "Double border",
			borderStyle: layout.BorderDouble,
			expectedW:   2,
			expectedH:   2,
		},
		{
			name:        "No border",
			borderStyle: layout.BorderNone,
			expectedW:   0,
			expectedH:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New().SetBorderStyle(tt.borderStyle)
			w, h := v.GetBorderPadding()
			if w != tt.expectedW {
				t.Errorf("Expected width padding %d, got %d", tt.expectedW, w)
			}
			if h != tt.expectedH {
				t.Errorf("Expected height padding %d, got %d", tt.expectedH, h)
			}
		})
	}
}

func TestVNode_SetOuterSize(t *testing.T) {
	v := New().SetOuterSize(30, 10)

	if v.width != 30 {
		t.Errorf("Expected width 30, got %d", v.width)
	}
	if v.height != 10 {
		t.Errorf("Expected height 10, got %d", v.height)
	}
}

func TestVNode_SetInnerSize(t *testing.T) {
	v := New().SetInnerSize(20, 5)

	// Should include border padding
	borderPadding := 2 // default BorderSingle
	if v.width != 20+borderPadding {
		t.Errorf("Expected width %d, got %d", 20+borderPadding, v.width)
	}
	if v.height != 5+borderPadding {
		t.Errorf("Expected height %d, got %d", 5+borderPadding, v.height)
	}
}

func TestVNode_FixedMethods(t *testing.T) {
	// Test FixedSize
	v1 := New().FixedSize(30, 10)
	if v1.width != 30 || v1.height != 10 {
		t.Errorf("FixedSize failed: got %dx%d", v1.width, v1.height)
	}

	// Test AutoHeight
	v2 := New().AutoHeight()
	if v2.height != 0 {
		t.Errorf("AutoHeight failed: got height %d", v2.height)
	}

	// Test AutoWidth
	v3 := New().AutoWidth()
	if v3.width != 0 {
		t.Errorf("AutoWidth failed: got width %d", v3.width)
	}

	// Test AutoSize
	v4 := New().AutoSize()
	if v4.width != 0 || v4.height != 0 {
		t.Errorf("AutoSize failed: got %dx%d", v4.width, v4.height)
	}

	// Test FixedWidthAutoHeight
	v5 := New().FixedWidthAutoHeight(30)
	if v5.width != 30 || v5.height != 0 {
		t.Errorf("FixedWidthAutoHeight failed: got %dx%d", v5.width, v5.height)
	}

	// Test FixedHeightAutoWidth
	v6 := New().FixedHeightAutoWidth(10)
	if v6.width != 0 || v6.height != 10 {
		t.Errorf("FixedHeightAutoWidth failed: got %dx%d", v6.width, v6.height)
	}
}

func TestVNode_WithMethods(t *testing.T) {
	v := New().
		WithTitle("Test Title").
		WithContent(newtext.New("Content")).
		WithOuterDimensions(30, 10).
		WithBorderStyleAndColor(layout.BorderDouble, style.Color("red"))

	if v.title != "Test Title" {
		t.Errorf("Expected title 'Test Title', got '%s'", v.title)
	}
	if v.width != 30 || v.height != 10 {
		t.Errorf("Expected dimensions 30x10, got %dx%d", v.width, v.height)
	}
	if v.borderStyle != layout.BorderDouble {
		t.Errorf("Expected BorderDouble, got %v", v.borderStyle)
	}
	if v.borderColor != "red" {
		t.Errorf("Expected color 'red', got '%s'", v.borderColor)
	}
}

func TestVNode_Presets(t *testing.T) {
	tests := []struct {
		name         string
		presetFunc   func() *VNode
		checkFunc    func(*testing.T, *VNode)
	}{
		{
			name: "InfoPanel",
			presetFunc: func() *VNode {
				return InfoPanel("Info", "Information message")
			},
			checkFunc: func(t *testing.T, v *VNode) {
				if v.title != "Info" {
					t.Errorf("Expected title 'Info', got '%s'", v.title)
				}
				if v.borderStyle != layout.BorderSingle {
					t.Errorf("Expected BorderSingle, got %v", v.borderStyle)
				}
				if v.borderColor != "blue" {
					t.Errorf("Expected color 'blue', got '%s'", v.borderColor)
				}
				if v.content == nil {
					t.Error("Content should not be nil")
				}
			},
		},
		{
			name: "WarningPanel",
			presetFunc: func() *VNode {
				return WarningPanel("Warning", "Warning message")
			},
			checkFunc: func(t *testing.T, v *VNode) {
				if v.borderColor != "yellow" {
					t.Errorf("Expected color 'yellow', got '%s'", v.borderColor)
				}
			},
		},
		{
			name: "ErrorPanel",
			presetFunc: func() *VNode {
				return ErrorPanel("Error", "Error message")
			},
			checkFunc: func(t *testing.T, v *VNode) {
				if v.borderStyle != layout.BorderDouble {
					t.Errorf("Expected BorderDouble, got %v", v.borderStyle)
				}
				if v.borderColor != "red" {
					t.Errorf("Expected color 'red', got '%s'", v.borderColor)
				}
			},
		},
		{
			name: "SuccessPanel",
			presetFunc: func() *VNode {
				return SuccessPanel("Success", "Success message")
			},
			checkFunc: func(t *testing.T, v *VNode) {
				if v.borderColor != "green" {
					t.Errorf("Expected color 'green', got '%s'", v.borderColor)
				}
			},
		},
		{
			name: "TextPanel",
			presetFunc: func() *VNode {
				return TextPanel("Text Panel", "Text content", 20)
			},
			checkFunc: func(t *testing.T, v *VNode) {
				if v.title != "Text Panel" {
					t.Errorf("Expected title 'Text Panel', got '%s'", v.title)
				}
				// Width should include border padding
				expectedWidth := 20 + 2
				if v.width != expectedWidth {
					t.Errorf("Expected width %d, got %d", expectedWidth, v.width)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := tt.presetFunc()
			tt.checkFunc(t, v)
		})
	}
}

// =============================================================================
// Utility Function Tests
// =============================================================================

func TestCalculateOuterWidth(t *testing.T) {
	tests := []struct {
		name       string
		innerWidth int
		style      layout.BorderStyle
		expected   int
	}{
		{"Single border 20", 20, layout.BorderSingle, 22},
		{"Rounded border 20", 20, layout.BorderRounded, 22},
		{"Double border 20", 20, layout.BorderDouble, 22},
		{"No border 20", 20, layout.BorderNone, 20},
		{"Zero width", 0, layout.BorderSingle, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateOuterWidth(tt.innerWidth, tt.style)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestCalculateOuterHeight(t *testing.T) {
	tests := []struct {
		name        string
		innerHeight int
		style       layout.BorderStyle
		expected    int
	}{
		{"Single border 5", 5, layout.BorderSingle, 7},
		{"Rounded border 5", 5, layout.BorderRounded, 7},
		{"Double border 5", 5, layout.BorderDouble, 7},
		{"No border 5", 5, layout.BorderNone, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateOuterHeight(tt.innerHeight, tt.style)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestCalculateInnerWidth(t *testing.T) {
	tests := []struct {
		name        string
		outerWidth  int
		style       layout.BorderStyle
		expected    int
	}{
		{"Single border 22", 22, layout.BorderSingle, 20},
		{"Rounded border 22", 22, layout.BorderRounded, 20},
		{"No border 20", 20, layout.BorderNone, 20},
		{"Smaller than padding", 1, layout.BorderSingle, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateInnerWidth(tt.outerWidth, tt.style)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestCalculateInnerHeight(t *testing.T) {
	tests := []struct {
		name         string
		outerHeight  int
		style        layout.BorderStyle
		expected     int
	}{
		{"Single border 7", 7, layout.BorderSingle, 5},
		{"Rounded border 7", 7, layout.BorderRounded, 5},
		{"No border 5", 5, layout.BorderNone, 5},
		{"Smaller than padding", 1, layout.BorderSingle, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateInnerHeight(tt.outerHeight, tt.style)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}
