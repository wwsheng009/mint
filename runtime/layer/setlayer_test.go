package layer

import (
	"fmt"
	"os"
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestSetLayerBasic 验证 SetLayer 是否正确工作
func TestSetLayerBasic(t *testing.T) {
	// 创建一个简单的节点
	node := rtui.NewElement("div")

	// 检查初始 layer
	fmt.Fprintf(os.Stderr, "Initial layer: %d\n", node.GetLayer())
	if node.GetLayer() != rtui.LayerBase {
		t.Errorf("❌ Expected LayerBase, got %d", node.GetLayer())
	}

	// 设置为 Inspector
	nodeWithLayer := node.SetLayer(rtui.LayerInspector)

	// 检查新的 layer
	fmt.Fprintf(os.Stderr, "After SetLayer: %d\n", nodeWithLayer.GetLayer())
	if nodeWithLayer.GetLayer() != rtui.LayerInspector {
		t.Errorf("❌ Expected LayerInspector, got %d", nodeWithLayer.GetLayer())
	} else {
		fmt.Fprintf(os.Stderr, "✅ SetLayer works correctly\n")
	}
}

// TestCollectorWithInspector 验证 Collector 是否能识别 Inspector
func TestCollectorWithInspector(t *testing.T) {
	// 创建 appContent
	appContent := rtui.NewElement("app")

	// 创建 inspectorOverlay 并设置 layer
	inspectorOverlay := rtui.NewElement("inspector")
	inspectorOverlay.SetLayer(rtui.LayerInspector)

	fmt.Fprintf(os.Stderr, "inspectorOverlay.GetLayer() = %d\n", inspectorOverlay.GetLayer())

	// 使用 Fragment 包裹
	root := rtui.Fragment(appContent, inspectorOverlay)

	// 收集 layer 节点
	collector := NewCollector()
	collector.Collect(root)

	// 检查是否收集到 Inspector
	hasInspector := collector.HasInspector()
	fmt.Fprintf(os.Stderr, "Has Inspector: %v\n", hasInspector)

	if !hasInspector {
		t.Fatal("❌ Collector did not find Inspector!")
	}

	// 获取 Inspector 节点
	inspectorNodes := collector.GetInspectorNodes()
	fmt.Fprintf(os.Stderr, "Inspector nodes count: %d\n", len(inspectorNodes))

	if len(inspectorNodes) == 0 {
		t.Fatal("❌ No inspector nodes found!")
	}

	fmt.Fprintf(os.Stderr, "✅ Collector correctly found Inspector\n")
}
