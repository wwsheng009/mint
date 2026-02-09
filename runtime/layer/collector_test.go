package layer

import (
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestStripLayersPreservesBaseContent verifies that StripLayers correctly removes
// layer nodes while preserving base content
func TestStripLayersPreservesBaseContent(t *testing.T) {
	// Create a tree similar to main.go:
	// VStack(
	//     appContent (VStack with multiple children),
	//     inspectorOverlay (LayerInspector),
	// )

	appContent := rtui.VStack(
		rtui.NewElement("text"),
		rtui.NewElement("text"),
		rtui.NewElement("text"),
	)

	inspectorOverlay := rtui.VStack(
		rtui.NewElement("text"),
	)
	inspectorOverlay.SetLayer(rtui.LayerInspector)

	root := rtui.VStack(
		appContent,
		inspectorOverlay,
	)

	collector := NewCollector()

	// Strip layers
	baseTree := collector.StripLayers(root)

	if baseTree == nil {
		t.Fatal("baseTree is nil after stripping")
	}

	// Verify baseTree still has appContent
	children := baseTree.Children()
	if len(children) != 1 {
		t.Errorf("Expected 1 child after stripping, got %d", len(children))
	}

	// Verify the remaining child is appContent (not inspectorOverlay)
	if children[0].GetLayer() != rtui.LayerBase {
		t.Errorf("Expected remaining child to be LayerBase, got %v", children[0].GetLayer())
	}

	// Verify appContent's children are preserved
	appChildren := children[0].Children()
	if len(appChildren) != 3 {
		t.Errorf("Expected appContent to have 3 children, got %d", len(appChildren))
	}

	t.Logf("✅ StripLayers correctly preserved base content")
	t.Logf("✅ baseTree has %d children", len(children))
	t.Logf("✅ appContent has %d children", len(appChildren))
}

// TestStripLayersMultipleLayers verifies stripping multiple layer types
func TestStripLayersMultipleLayers(t *testing.T) {
	// Create tree with multiple layers
	baseContent := rtui.NewElement("text")
	modalOverlay := rtui.NewElement("text")
	modalOverlay.SetLayer(rtui.LayerModal)

	inspectorOverlay := rtui.NewElement("text")
	inspectorOverlay.SetLayer(rtui.LayerInspector)

	root := rtui.VStack(
		baseContent,
		modalOverlay,
		inspectorOverlay,
	)

	collector := NewCollector()
	baseTree := collector.StripLayers(root)

	if baseTree == nil {
		t.Fatal("baseTree is nil after stripping")
	}

	children := baseTree.Children()
	if len(children) != 1 {
		t.Errorf("Expected 1 child after stripping, got %d", len(children))
	}

	t.Logf("✅ StripLayers correctly removed multiple layers")
}

// TestStripLayersEmptyTree verifies edge case of empty tree
func TestStripLayersEmptyTree(t *testing.T) {
	collector := NewCollector()

	// Test nil
	baseTree := collector.StripLayers(nil)
	if baseTree != nil {
		t.Error("Expected nil for nil input")
	}

	// Test empty VStack
	emptyVStack := rtui.VStack()
	baseTree = collector.StripLayers(emptyVStack)
	if baseTree == nil {
		t.Error("Expected non-nil for empty VStack")
	}

	t.Logf("✅ StripLayers handles edge cases correctly")
}
