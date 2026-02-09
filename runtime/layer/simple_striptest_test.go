package layer

import (
	"fmt"
	"os"
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestStripLayersSimple 简单测试 StripLayers
func TestStripLayersSimple(t *testing.T) {
	// 创建 appContent
	appContent := rtui.NewElement("app")

	// 创建 inspectorOverlay 并设置 layer
	inspectorOverlay := rtui.NewElement("inspector")
	inspectorOverlay.SetLayer(rtui.LayerInspector)

	fmt.Fprintf(os.Stderr, "inspectorOverlay.GetLayer() before Fragment: %d\n", inspectorOverlay.GetLayer())

	// 使用 Fragment 包裹
	root := rtui.Fragment(appContent, inspectorOverlay)

	fmt.Fprintf(os.Stderr, "root.Children() count: %d\n", len(root.Children()))

	// 调用 StripLayers
	collector := NewCollector()
	collector.Collect(root)  // 先 Collect
	baseTree := collector.StripLayers(root)  // 再 StripLayers

	fmt.Fprintf(os.Stderr, "baseTree type: %T\n", baseTree)
	fmt.Fprintf(os.Stderr, "baseTree.Children() count: %d\n", len(baseTree.Children()))

	// 验证：baseTree 应该只有 1 个 child (appContent)
	if len(baseTree.Children()) != 1 {
		t.Errorf("❌ Expected 1 child after stripping, got %d", len(baseTree.Children()))
	} else {
		fmt.Fprintf(os.Stderr, "✅ baseTree has 1 child (correct)\n")
	}

	// 验证：唯一的 child 是 appContent (不是 inspectorOverlay)
	firstChild := baseTree.Children()[0]
	fmt.Fprintf(os.Stderr, "firstChild type: %T\n", firstChild)

	// 检查 firstChild 的 layer
	fmt.Fprintf(os.Stderr, "firstChild.GetLayer(): %d\n", firstChild.GetLayer())

	if firstChild.GetLayer() != rtui.LayerBase {
		t.Errorf("❌ Expected LayerBase, got %d", firstChild.GetLayer())
	} else {
		fmt.Fprintf(os.Stderr, "✅ firstChild is LayerBase (correct)\n")
	}
}
