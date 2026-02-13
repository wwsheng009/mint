package layer

import (
	"fmt"
	"os"
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestInspectorLayerTracking 追踪 Inspector layer 的问题
func TestInspectorLayerTracking(t *testing.T) {
	os.Setenv("TUI_DEBUG_LAYER", "true")

	// 创建 appContent
	appContent := rtui.NewElement("app")

	// 创建 Inspector
	inspector := rtui.Bordered().
		Label("INSPECTOR").
		Child(rtui.NewElement("content")).
		Build()

	fmt.Fprintf(os.Stderr, "\n=== Step 1: Create Inspector ===\n")
	fmt.Fprintf(os.Stderr, "inspector.GetLayer() before SetLayer: %d\n", inspector.GetLayer())

	// 设置 layer
	inspectorWithLayer := inspector.SetLayer(rtui.LayerInspector)

	fmt.Fprintf(os.Stderr, "inspectorWithLayer.GetLayer(): %d\n", inspectorWithLayer.GetLayer())
	fmt.Fprintf(os.Stderr, "inspectorWithLayer == inspector: %v\n", inspectorWithLayer == inspector)

	// 使用 Fragment 包裹
	root := rtui.Fragment(appContent, inspectorWithLayer)

	fmt.Fprintf(os.Stderr, "\n=== Step 2: In Fragment ===\n")
	fmt.Fprintf(os.Stderr, "root.Children() count: %d\n", len(root.Children()))

	if len(root.Children()) >= 2 {
		secondChild := root.Children()[1]
		fmt.Fprintf(os.Stderr, "secondChild type: %T\n", secondChild)
		fmt.Fprintf(os.Stderr, "secondChild.GetLayer(): %d\n", secondChild.GetLayer())

		if secondChild.GetLayer() != rtui.LayerInspector {
			t.Errorf("❌ Expected LayerInspector (4), got %d", secondChild.GetLayer())
		} else {
			fmt.Fprintf(os.Stderr, "✅ Fragment preserves Inspector layer\n")
		}
	}

	// Collect
	fmt.Fprintf(os.Stderr, "\n=== Step 3: Collect ===\n")
	collector := NewCollector()
	collector.Collect(root)

	inspectorNodes := collector.GetInspectorNodes()
	fmt.Fprintf(os.Stderr, "Collected %d inspector nodes\n", len(inspectorNodes))

	if len(inspectorNodes) == 0 {
		t.Error("❌ Inspector was not collected!")
	} else {
		fmt.Fprintf(os.Stderr, "✅ Inspector was collected\n")
	}
}
