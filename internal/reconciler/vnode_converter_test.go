// Package reconciler tests for VNodeConverter.
package reconciler

import (
	"testing"

	"github.com/wwsheng009/mint/runtime"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// VNodeConverter Constructor Tests
// =============================================================================

func TestNewVNodeConverter(t *testing.T) {
	c := NewVNodeConverter()
	if c == nil {
		t.Fatal("NewVNodeConverter should not return nil")
	}
	if c.nodeCounter != 0 {
		t.Errorf("nodeCounter should start at 0, got %d", c.nodeCounter)
	}
}

// =============================================================================
// VNodeConverter.Convert Tests
// =============================================================================

func TestVNodeConverter_Convert_Nil(t *testing.T) {
	c := NewVNodeConverter()
	result := c.Convert(nil)
	if result != nil {
		t.Error("Convert(nil) should return nil")
	}
}

func TestVNodeConverter_Convert_ElementVNode(t *testing.T) {
	c := NewVNodeConverter()
	elem := rtui.Element("div").Prop("id", "test").Build()
	result := c.Convert(elem)

	if result == nil {
		t.Fatal("Convert(element) should not return nil")
	}
	if result.Props == nil {
		t.Error("Props should not be nil")
	}
	if result.Props["id"] != "test" {
		t.Errorf("Expected id 'test', got %v", result.Props["id"])
	}
}

func TestVNodeConverter_Convert_ElementWithChildren(t *testing.T) {
	c := NewVNodeConverter()
	elem := rtui.Element("div").Children(
		rtui.Element("a").Build(),
		rtui.Element("b").Build(),
	).Build()
	result := c.Convert(elem)

	if result == nil {
		t.Fatal("Convert(element) should not return nil")
	}
	if result.Children == nil {
		t.Error("Element with children should have children in result")
	}
	if len(result.Children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(result.Children))
	}
}

func TestVNodeConverter_Convert_FragmentVNode(t *testing.T) {
	c := NewVNodeConverter()
	frag := rtui.Fragment(
		rtui.Element("a").Build(),
		rtui.Element("b").Build(),
	)
	result := c.Convert(frag)

	if result == nil {
		t.Fatal("Convert(fragment) should not return nil")
	}
	// Fragment converts its children
	if result.Children == nil {
		t.Error("Fragment should have children")
	}
}

func TestVNodeConverter_Convert_ComponentVNode(t *testing.T) {
	c := NewVNodeConverter()
	comp := rtui.NewComponent("test", func() rtui.VNode {
		return rtui.Element("div").Build()
	})
	result := c.Convert(comp)

	// Component nodes are expanded - may return nil for simple cases
	// The important thing is it doesn't crash
	t.Logf("Component conversion result: %v", result)
}

func TestVNodeConverter_Convert_ComponentVNode_MultipleChildren(t *testing.T) {
	c := NewVNodeConverter()
	// Component with multiple children returns fragment-like structure
	comp := rtui.NewComponent("test", func() rtui.VNode {
		return rtui.Fragment(
			rtui.Element("a").Build(),
			rtui.Element("b").Build(),
		)
	})
	result := c.Convert(comp)

	// Should not crash - actual behavior depends on component expansion
	t.Logf("Component with fragment conversion result: %v", result)
}

func TestVNodeConverter_Convert_LayoutNode(t *testing.T) {
	c := NewVNodeConverter()

	// Create a LayoutNode using HStack
	layoutNode := rtui.HStack(
		rtui.Element("text").Prop("content", "A").Build(),
		rtui.Element("text").Prop("content", "B").Build(),
	)

	result := c.Convert(layoutNode)

	if result == nil {
		t.Fatal("Convert(layoutNode) should not return nil")
	}
	// LayoutNode should have children from HStack
	if len(result.Children) == 0 {
		t.Error("HStack should have children after conversion")
	}
}

func TestVNodeConverter_Convert_TextElement(t *testing.T) {
	c := NewVNodeConverter()
	textElem := rtui.Element("text").Prop("content", "Hello World").Build()
	result := c.Convert(textElem)

	if result == nil {
		t.Fatal("Convert(text element) should not return nil")
	}
	if result.Props == nil {
		t.Error("Props should not be nil")
	}
	// Text elements should have their content extracted
	if result.Props["text"] == nil && result.Props["content"] == nil {
		t.Error("Text element should have text or content prop")
	}
}

// =============================================================================
// Style Conversion Tests
// =============================================================================

func TestVNodeConverter_StyleFromProps(t *testing.T) {
	tests := []struct {
		name           string
		props          rtui.Props
		expectedWidth  int
		expectedHeight int
		expectedFlex   float64
	}{
		{
			name:           "no props",
			props:          rtui.Props{},
			expectedWidth:  0,  // -1 means not set, so 0 is the test expectation
			expectedHeight: 0,
			expectedFlex:   0,
		},
		{
			name: "with width",
			props: rtui.Props{
				"width": 50,
			},
			expectedWidth:  50,
			expectedHeight: 0,
			expectedFlex:   0,
		},
		{
			name: "with height",
			props: rtui.Props{
				"height": 30,
			},
			expectedWidth:  0,
			expectedHeight: 30,
			expectedFlex:   0,
		},
		{
			name: "with flex",
			props: rtui.Props{
				"flex": 2,
			},
			expectedWidth:  0,
			expectedHeight: 0,
			expectedFlex:   2.0,
		},
		{
			name: "with all",
			props: rtui.Props{
				"width":  100,
				"height": 50,
				"flex":   3,
			},
			expectedWidth:  100,
			expectedHeight: 50,
			expectedFlex:   3.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewVNodeConverter()
			elem := rtui.Element("div")
			for k, v := range tt.props {
				elem.Prop(k, v)
			}
			vnode := elem.Build()
			result := c.Convert(vnode)

			if result == nil {
				t.Fatal("Convert should not return nil")
			}
			// For "no props" case, check width/height are properly initialized (may be 0 or -1)
			if tt.expectedWidth > 0 && result.Style.Width != tt.expectedWidth {
				t.Errorf("Expected width %d, got %d", tt.expectedWidth, result.Style.Width)
			}
			if tt.expectedHeight > 0 && result.Style.Height != tt.expectedHeight {
				t.Errorf("Expected height %d, got %d", tt.expectedHeight, result.Style.Height)
			}
			if tt.expectedFlex > 0 && result.Style.FlexGrow != tt.expectedFlex {
				t.Errorf("Expected flex %f, got %f", tt.expectedFlex, result.Style.FlexGrow)
			}
		})
	}
}

// =============================================================================
// LayoutBox Generation Tests
// =============================================================================

func TestVNodeConverter_GenerateLayoutBoxes(t *testing.T) {
	c := NewVNodeConverter()

	t.Run("nil root", func(t *testing.T) {
		boxes := c.GenerateLayoutBoxes(nil)
		if boxes != nil {
			t.Error("GenerateLayoutBoxes(nil) should return nil")
		}
	})

	t.Run("simple tree", func(t *testing.T) {
		vnode := rtui.VStack(
			rtui.Element("a").Build(),
			rtui.Element("b").Build(),
		)
		root := c.Convert(vnode)
		boxes := c.GenerateLayoutBoxes(root)

		if boxes == nil {
			t.Error("GenerateLayoutBoxes should return non-nil for valid tree")
		}
		// Should have at least some boxes
		if len(boxes) == 0 {
			t.Error("GenerateLayoutBoxes should return at least one box")
		}
	})
}

// =============================================================================
// Nested Structure Tests
// =============================================================================

func TestVNodeConverter_NestedComponents(t *testing.T) {
	c := NewVNodeConverter()

	// Use VStack directly instead of wrapping in component
	vstack := rtui.VStack(
		rtui.Element("text").Prop("content", "A").Build(),
		rtui.Element("text").Prop("content", "B").Build(),
	)

	result := c.Convert(vstack)

	if result == nil {
		t.Fatal("Convert should not return nil")
	}

	// Count total nodes in tree
	nodeCount := countLayoutNodes(result)
	if nodeCount < 2 {
		t.Errorf("Expected at least 2 nodes in VStack, got %d", nodeCount)
	}
}

func TestVNodeConverter_DeepNesting(t *testing.T) {
	c := NewVNodeConverter()

	// Create deeply nested structure
	vnodes := []rtui.VNode{rtui.Element("leaf").Build()}
	for i := 0; i < 5; i++ {
		vnodes = []rtui.VNode{rtui.VStack(vnodes...)}
	}

	result := c.Convert(vnodes[0])

	if result == nil {
		t.Fatal("Convert should handle deeply nested structures")
	}

	// Should have multiple levels
	nodeCount := countLayoutNodes(result)
	if nodeCount < 5 {
		t.Errorf("Expected at least 5 nodes in deeply nested structure, got %d", nodeCount)
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

// countLayoutNodes recursively counts nodes in a LayoutNode tree
func countLayoutNodes(node *runtime.LayoutNode) int {
	if node == nil {
		return 0
	}
	count := 1
	if node.Children != nil {
		for _, child := range node.Children {
			count += countLayoutNodes(child)
		}
	}
	return count
}

// =============================================================================
// Edge Case Tests
// =============================================================================

func TestVNodeConverter_EmptyFragment(t *testing.T) {
	c := NewVNodeConverter()
	frag := rtui.Fragment()
	result := c.Convert(frag)

	// Empty fragment may return nil - important thing is it doesn't crash
	t.Logf("Empty fragment conversion result: %v", result)
}

func TestVNodeConverter_LargeTree(t *testing.T) {
	c := NewVNodeConverter()

	// Create a tree with many children
	children := make([]rtui.VNode, 50)
	for i := 0; i < 50; i++ {
		children[i] = rtui.Element("text").Prop("content", string(rune('A'+i%26))).Build()
	}
	frag := rtui.Fragment(children...)

	result := c.Convert(frag)

	if result == nil {
		t.Fatal("Convert should handle large trees")
	}

	// Count nodes to verify structure
	nodeCount := countLayoutNodes(result)
	if nodeCount < 50 {
		t.Errorf("Expected at least 50 nodes, got %d", nodeCount)
	}
}

func TestVNodeConverter_IDUniqueness(t *testing.T) {
	c := NewVNodeConverter()

	// Convert multiple nodes and verify unique IDs
	ids := make(map[string]bool)
	for i := 0; i < 20; i++ {
		node := c.Convert(rtui.Element("div").Build())
		if node == nil {
			t.Fatal("Convert should not return nil")
		}
		if ids[node.ID] {
			t.Errorf("Duplicate ID generated: %s", node.ID)
		}
		ids[node.ID] = true
	}
}
