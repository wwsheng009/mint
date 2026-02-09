package layer

import (
	"fmt"
	"os"
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestStripLayersDebug 调试 StripLayers 的执行过程
func TestStripLayersDebug(t *testing.T) {
	os.Setenv("TUI_LAYER_DEBUG", "true")

	// 创建 appContent (使用 VStack，模拟真实的应用)
	appContent := rtui.VStack(
		rtui.NewElement("header"),
		rtui.NewElement("content"),
		rtui.NewElement("footer"),
	)

	// 创建 inspectorOverlay 并设置 layer
	inspectorOverlay := rtui.NewElement("inspector")
	inspectorOverlay.SetLayer(rtui.LayerInspector)

	fmt.Fprintf(os.Stderr, "\n=== Before Fragment ===\n")
	fmt.Fprintf(os.Stderr, "appContent type: %T\n", appContent)
	fmt.Fprintf(os.Stderr, "inspectorOverlay.GetLayer(): %d\n", inspectorOverlay.GetLayer())

	// 使用 Fragment 包裹
	root := rtui.Fragment(appContent, inspectorOverlay)

	fmt.Fprintf(os.Stderr, "\n=== After Fragment ===\n")
	fmt.Fprintf(os.Stderr, "root type: %T\n", root)
	fmt.Fprintf(os.Stderr, "root.GetLayer(): %d\n", root.GetLayer())
	fmt.Fprintf(os.Stderr, "root.Children() count: %d\n", len(root.Children()))

	// 调用 StripLayers
	collector := NewCollector()
	collector.Collect(root)
	baseTree := collector.StripLayers(root)

	fmt.Fprintf(os.Stderr, "\n=== After StripLayers ===\n")
	fmt.Fprintf(os.Stderr, "baseTree type: %T\n", baseTree)

	if baseTree == nil {
		t.Fatal("❌ baseTree is nil!")
	}

	fmt.Fprintf(os.Stderr, "baseTree.GetLayer(): %d\n", baseTree.GetLayer())
	fmt.Fprintf(os.Stderr, "baseTree.Children() count: %d\n", len(baseTree.Children()))

	if len(baseTree.Children()) > 0 {
		for i, child := range baseTree.Children() {
			fmt.Fprintf(os.Stderr, "  child %d: type=%T, layer=%d\n", i, child, child.GetLayer())
		}
	}
}
