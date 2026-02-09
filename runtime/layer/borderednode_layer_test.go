package layer

import (
	"fmt"
	"os"
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestBorderedNodeSetLayer 验证 BorderedNode.SetLayer 是否生效
func TestBorderedNodeSetLayer(t *testing.T) {
	// 创建一个 BorderedNode
	inspector := rtui.Bordered().
		Label("INSPECTOR").
		Child(rtui.NewElement("content")).
		Build()

	fmt.Fprintf(os.Stderr, "\n=== Before SetLayer ===\n")
	fmt.Fprintf(os.Stderr, "inspector type: %T\n", inspector)
	fmt.Fprintf(os.Stderr, "inspector.GetLayer(): %d\n", inspector.GetLayer())

	// 设置 layer
	inspectorWithLayer := inspector.SetLayer(rtui.LayerInspector)

	fmt.Fprintf(os.Stderr, "\n=== After SetLayer ===\n")
	fmt.Fprintf(os.Stderr, "inspectorWithLayer type: %T\n", inspectorWithLayer)
	fmt.Fprintf(os.Stderr, "inspectorWithLayer.GetLayer(): %d\n", inspectorWithLayer.GetLayer())

	if inspectorWithLayer.GetLayer() != rtui.LayerInspector {
		t.Errorf("❌ Expected LayerInspector (4), got %d", inspectorWithLayer.GetLayer())
	} else {
		fmt.Fprintf(os.Stderr, "✅ SetLayer works correctly\n")
	}

	// 验证是否是同一个实例
	if inspectorWithLayer != inspector {
		fmt.Fprintf(os.Stderr, "⚠️  SetLayer() 返回了新实例\n")
	} else {
		fmt.Fprintf(os.Stderr, "✅ SetLayer() 返回了同一个实例\n")
	}
}

// TestBorderedNodeInFragment 验证 Fragment 中的 BorderedNode SetLayer
func TestBorderedNodeInFragment(t *testing.T) {
	// 创建 appContent
	appContent := rtui.NewElement("app")

	// 创建 BorderedNode Inspector
	inspector := rtui.Bordered().
		Label("INSPECTOR").
		Child(rtui.NewElement("content")).
		Build()

	fmt.Fprintf(os.Stderr, "\n=== Before SetLayer ===\n")
	fmt.Fprintf(os.Stderr, "inspector.GetLayer(): %d\n", inspector.GetLayer())

	// 设置 layer
	inspector.SetLayer(rtui.LayerInspector)

	fmt.Fprintf(os.Stderr, "\n=== After SetLayer ===\n")
	fmt.Fprintf(os.Stderr, "inspector.GetLayer(): %d\n", inspector.GetLayer())

	if inspector.GetLayer() != rtui.LayerInspector {
		t.Errorf("❌ Expected LayerInspector (4), got %d", inspector.GetLayer())
	}

	// 使用 Fragment 包裹
	root := rtui.Fragment(appContent, inspector)

	fmt.Fprintf(os.Stderr, "\n=== In Fragment ===\n")
	fmt.Fprintf(os.Stderr, "root.Children() count: %d\n", len(root.Children()))

	// 检查 Fragment 的第二个 child 的 layer
	secondChild := root.Children()[1]
	fmt.Fprintf(os.Stderr, "secondChild type: %T\n", secondChild)
	fmt.Fprintf(os.Stderr, "secondChild.GetLayer(): %d\n", secondChild.GetLayer())

	if secondChild.GetLayer() != rtui.LayerInspector {
		t.Errorf("❌ Expected LayerInspector (4), got %d", secondChild.GetLayer())
	} else {
		fmt.Fprintf(os.Stderr, "✅ Inspector in Fragment has correct layer\n")
	}
}
