package ui_test

import (
	"testing"

	"github.com/wwsheng009/mint/runtime"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestBorderedFillWidth tests that Bordered containers properly stretch
// when fillWidth is enabled
func TestBorderedFillWidth(t *testing.T) {
	tests := []struct {
		name           string
		fillWidth      bool
		constraintWidth int
		expectWidth    int
	}{
		{
			name:           "Bordered with fillWidth=true in bounded container",
			fillWidth:      true,
			constraintWidth: 100,
			expectWidth:    100, // Should fill available width
		},
		{
			name:           "Bordered with fillWidth=false in bounded container",
			fillWidth:      false,
			constraintWidth: 100,
			expectWidth:    0, // Should use content's natural width
		},
		{
			name:           "Bordered with fillWidth=true in unbounded container",
			fillWidth:      true,
			constraintWidth: 0, // Unbounded
			expectWidth:    0, // Should use content's natural width
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a bordered container with child
			borderedBuilder := rtui.Bordered()
			if tt.fillWidth {
				borderedBuilder.FillWidth()
			}
			borderedBuilder.Child(rtui.Element("text").Prop("content", "Test").Build())

			// Build to get VNode
			bordered := borderedBuilder.Build()

			// Type assert to BorderedNode to access Measure method
			borderedNode, ok := bordered.(*rtui.BorderedNode)
			if !ok {
				t.Fatalf("Expected BorderedNode, got %T", bordered)
			}

			// Measure with constraints
			constraints := runtime.BoxConstraints{
				MinWidth: 0,
				MaxWidth: tt.constraintWidth,
			}
			if tt.constraintWidth == 0 {
				constraints.MaxWidth = runtime.Infinity
			}

			size := borderedNode.Measure(constraints)

			// Verify width
			expectedWidth := tt.expectWidth
			if expectedWidth == 0 {
				// For natural size, width should be content width + border (2)
				// Content "Test" is 4 chars, plus 2 for border = 6
				if size.Width != 6 {
					t.Errorf("Expected natural width to be content+border (6), got %d", size.Width)
				}
			} else {
				// For fillWidth=true in bounded container, width should match constraint
				if size.Width != expectedWidth {
					t.Errorf("Expected width %d, got %d", expectedWidth, size.Width)
				}
			}

			t.Logf("Test '%s': fillWidth=%v, constraintWidth=%d, gotWidth=%d",
				tt.name, tt.fillWidth, tt.constraintWidth, size.Width)
		})
	}
}

// TestBorderedFillWidthVsFlex compares fillWidth with flex behavior
func TestBorderedFillWidthVsFlex(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func() rtui.VNode
		expectFlex  bool
		expectFill  bool
	}{
		{
			name: "Bordered with FillWidth()",
			setupFunc: func() rtui.VNode {
				return rtui.Bordered().
					Child(rtui.Element("text").Prop("content", "Test").Build()).
					FillWidth().
					Build()
			},
			expectFlex:  false,
			expectFill:  true,
		},
		{
			name: "Bordered with Flex(1)",
			setupFunc: func() rtui.VNode {
				return rtui.Bordered().
					Child(rtui.Element("text").Prop("content", "Test").Build()).
					Flex(1).
					Build()
			},
			expectFlex:  true,
			expectFill:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vnode := tt.setupFunc()

			// Check GetLayoutInfo
			info := rtui.GetLayoutInfo(vnode)

			expectedFlex := 0
			if tt.expectFlex {
				expectedFlex = 1
			}
			if info.Flex != expectedFlex {
				t.Errorf("Expected Flex=%v, got %v", expectedFlex, info.Flex)
			}
			if info.FillWidth != tt.expectFill {
				t.Errorf("Expected FillWidth=%v, got %v", tt.expectFill, info.FillWidth)
			}

			t.Logf("Test '%s': Flex=%v, FillWidth=%v",
				tt.name, info.Flex, info.FillWidth)
		})
	}
}

// TestBorderedFillWidthInVStack tests fillWidth behavior in VStack layout
func TestBorderedFillWidthInVStack(t *testing.T) {
	// Create a Bordered with fillWidth=true
	bordered := rtui.Bordered().
		Child(rtui.Element("text").Prop("content", "Content").Build()).
		FillWidth().
		Build()

	vstack := rtui.VStack(bordered)

	// Type assert to get Measure method
	layoutNode, ok := vstack.(*rtui.LayoutNode)
	if !ok {
		t.Fatalf("Expected LayoutNode, got %T", vstack)
	}

	// Measure with bounded width
	constraints := runtime.BoxConstraints{
		MinWidth: 0,
		MaxWidth: 80,
	}

	size := layoutNode.Measure(constraints)

	// VStack should expand to fill width (80), Bordered should too (minus border)
	// Bordered content width = 80 - 2 (border) = 78, total = 80
	if size.Width != 80 {
		t.Errorf("Expected VStack with Bordered(fillWidth) child to be 80 wide, got %d", size.Width)
	}

	t.Logf("VStack size: %dx%d (Bordered with fillWidth should fill parent width)", size.Width, size.Height)
}

// TestBorderedNoFillWidthInVStack tests Bordered without fillWidth in VStack
func TestBorderedNoFillWidthInVStack(t *testing.T) {
	// Create a Bordered WITHOUT fillWidth
	bordered := rtui.Bordered().
		Child(rtui.Element("text").Prop("content", "Content").Build()).
		Build()

	vstack := rtui.VStack(bordered)

	// Type assert to get Measure method
	layoutNode, ok := vstack.(*rtui.LayoutNode)
	if !ok {
		t.Fatalf("Expected LayoutNode, got %T", vstack)
	}

	// Measure with bounded width
	constraints := runtime.BoxConstraints{
		MinWidth: 0,
		MaxWidth: 80,
	}

	size := layoutNode.Measure(constraints)

	// Bordered should only be as wide as its content
	// Content "Content" = 7 chars + 2 border = 9
	if size.Width > 20 { // Allow some padding but not full width
		t.Errorf("Expected Bordered without fillWidth to use natural width (~9), got %d", size.Width)
	}

	t.Logf("VStack size: %dx%d (Bordered without fillWidth should use natural width)", size.Width, size.Height)
}
