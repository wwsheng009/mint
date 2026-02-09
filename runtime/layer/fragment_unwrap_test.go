package layer

import (
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestStripLayersFragmentUnwrap 验证 StripLayers 对 Fragment 的处理
func TestStripLayersFragmentUnwrap(t *testing.T) {
	// 创建 appContent (VStack)
	appContent := rtui.VStack(
		rtui.NewElement("text"),
		rtui.NewElement("text"),
		rtui.NewElement("text"),
	)

	// 创建 inspectorOverlay
	inspectorOverlay := rtui.NewElement("inspector")
	inspectorOverlay.SetLayer(rtui.LayerInspector)

	// 使用 Fragment 包裹 (修复后的方案)
	root := rtui.Fragment(appContent, inspectorOverlay)

	// StripLayers
	collector := NewCollector()
	collector.Collect(root)
	baseTree := collector.StripLayers(root)

	t.Logf("baseTree type: %T", baseTree)

	// 验证：Fragment 被解包，直接返回 appContent
	if baseTree == nil {
		t.Fatal("❌ baseTree is nil")
	}

	// 检查是否是 VStack (appContent)
	if _, ok := baseTree.(*rtui.LayoutNode); ok {
		t.Logf("✅ Fragment unwrapped: baseTree is LayoutNode")
	} else {
		t.Errorf("❌ Fragment not unwrapped: baseTree is %T", baseTree)
	}

	// 验证子节点
	children := baseTree.Children()
	t.Logf("baseTree has %d children", len(children))
	if len(children) != 3 {
		t.Errorf("❌ Expected 3 children, got %d", len(children))
	}
}
