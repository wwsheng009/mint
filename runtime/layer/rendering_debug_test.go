package layer

import (
	"fmt"
	"os"
	"testing"

	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/compute"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestRenderingDebug 完整追踪渲染流程，找出为什么没有显示
func TestRenderingDebug(t *testing.T) {
	os.Setenv("TUI_LAYER_DEBUG", "true")

	// 创建简化版的应用内容
	appContent := rtui.VStack(
		createBorderedText("Header", "This is the header"),
		createBorderedText("Content", "This is the main content"),
		createBorderedText("Footer", "This is the footer"),
	)

	// 创建 Inspector
	inspectorOverlay := rtui.Bordered().
		Label("INSPECTOR").
		Child(createBorderedText("Inspector", "Inspector content")).
		Build()
	inspectorOverlay.SetLayer(rtui.LayerInspector)
	inspectorOverlay.SetProps(rtui.Props{
		"x": 40,
		"y": 5,
	})

	// 使用 Fragment 包裹
	root := rtui.Fragment(appContent, inspectorOverlay)

	// 1. 收集和布局
	manager := NewManager()
	engine := compute.NewEngine()
	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  120,
		MinHeight: 0,
		MaxHeight: 40,
	}

	fmt.Fprintf(os.Stderr, "\n=== Step 1: CollectAndLayout ===\n")
	if err := manager.CollectAndLayout(root, constraints, engine); err != nil {
		t.Fatalf("CollectAndLayout failed: %v", err)
	}

	layouts := manager.GetLayouts()
	baseLayout, hasBase := layouts[rtui.LayerBase]
	if !hasBase {
		t.Fatal("❌ No base layout!")
	}

	inspectorLayout, hasInspector := layouts[rtui.LayerInspector]
	if !hasInspector {
		t.Fatal("❌ No inspector layout!")
	}

	fmt.Fprintf(os.Stderr, "baseLayout.Root type: %T\n", baseLayout.Root)
	fmt.Fprintf(os.Stderr, "baseLayout.Root.Box: (%d, %d, %dx%d)\n",
		baseLayout.Root.Box.X, baseLayout.Root.Box.Y,
		baseLayout.Root.Box.Width, baseLayout.Root.Box.Height)

	fmt.Fprintf(os.Stderr, "inspectorLayout.Root type: %T\n", inspectorLayout.Root)
	fmt.Fprintf(os.Stderr, "inspectorLayout.Root.Box: (%d, %d, %dx%d)\n",
		inspectorLayout.Root.Box.X, inspectorLayout.Root.Box.Y,
		inspectorLayout.Root.Box.Width, inspectorLayout.Root.Box.Height)

	// 验证 baseLayout 类型是 LayoutNode (不是 Fragment)
	if baseLayout.Root.VNode.Type() != rtui.VNodeElement {
		t.Logf("✅ baseLayout.Root.VNode type: %s", baseLayout.Root.VNode.Type().String())
	}
	if baseLayout.Root.VNode.Type() == rtui.VNodeFragment {
		t.Errorf("❌ baseLayout.Root.VNode is Fragment, should be LayoutNode")
	}

	// 验证高度合理
	if baseLayout.Root.Box.Height < 10 {
		t.Errorf("❌ baseLayout height too small: %d (expected >= 10)",
			baseLayout.Root.Box.Height)
	}

	fmt.Fprintf(os.Stderr, "\n=== Summary ===\n")
	fmt.Fprintf(os.Stderr, "✅ Base layer correctly laid out as LayoutNode\n")
	fmt.Fprintf(os.Stderr, "✅ Base layer has reasonable height: %d\n", baseLayout.Root.Box.Height)
	fmt.Fprintf(os.Stderr, "✅ Inspector positioned at (%d, %d)\n",
		inspectorLayout.Root.Box.X, inspectorLayout.Root.Box.Y)
}

// 辅助函数：创建带边框的文本
func createBorderedText(label, text string) rtui.VNode {
	textNode := rtui.NewElement("span")
	textNode.SetProps(rtui.Props{
		"text": text,
	})

	return rtui.Bordered().
		Label(label).
		Child(textNode).
		Build()
}
