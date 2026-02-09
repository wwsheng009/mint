package layer

import (
	"fmt"
	"os"
	"testing"

	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/compute"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestModalVsInspector 对比 modal 和 inspector 的渲染
func TestModalVsInspector(t *testing.T) {
	os.Setenv("TUI_LAYER_DEBUG", "true")

	// 创建 appContent (简单版本)
	appContent := rtui.Bordered().
		Label("Main").
		Child(rtui.NewElement("App Content")).
		Build()

	// 测试 1: 使用 Modal
	fmt.Fprintf(os.Stderr, "\n=== Test 1: With Modal ===\n")
	modalOverlay := rtui.Bordered().
		Label("MODAL").
		Child(rtui.NewElement("Modal Content")).
		Build()
	modalOverlay.SetLayer(rtui.LayerModal)

	rootWithModal := rtui.Fragment(appContent, modalOverlay)

	manager := NewManager()
	engine := compute.NewEngine()
	constraints := runtime.BoxConstraints{
		MinWidth:  0, MaxWidth: 120,
		MinHeight: 0, MaxHeight: 40,
	}

	if err := manager.CollectAndLayout(rootWithModal, constraints, engine); err != nil {
		t.Fatalf("CollectAndLayout with modal failed: %v", err)
	}

	layouts := manager.GetLayouts()
	fmt.Fprintf(os.Stderr, "Total layers: %d\n", len(layouts))

	for layer, layout := range layouts {
		if layout.Root != nil {
			fmt.Fprintf(os.Stderr, "Layer %d: root=(%d,%d) size=%dx%d\n",
				layer, layout.Root.Box.X, layout.Root.Box.Y,
				layout.Root.Box.Width, layout.Root.Box.Height)
		}
	}

	// 测试 2: 使用 Inspector
	fmt.Fprintf(os.Stderr, "\n=== Test 2: With Inspector ===\n")
	inspectorOverlay := rtui.Bordered().
		Label("INSPECTOR").
		Child(rtui.NewElement("Inspector Content")).
		Build()
	// IMPORTANT: SetProps BEFORE SetLayer, otherwise SetProps wipes out the _layer property
	inspectorOverlay.SetProps(rtui.Props{
		"x": 40,
		"y": 5,
	})
	inspectorOverlay.SetLayer(rtui.LayerInspector)

	rootWithInspector := rtui.Fragment(appContent, inspectorOverlay)

	manager2 := NewManager()
	if err := manager2.CollectAndLayout(rootWithInspector, constraints, engine); err != nil {
		t.Fatalf("CollectAndLayout with inspector failed: %v", err)
	}

	layouts2 := manager2.GetLayouts()
	fmt.Fprintf(os.Stderr, "Total layers: %d\n", len(layouts2))

	for layer, layout := range layouts2 {
		if layout.Root != nil {
			fmt.Fprintf(os.Stderr, "Layer %d: root=(%d,%d) size=%dx%d\n",
				layer, layout.Root.Box.X, layout.Root.Box.Y,
				layout.Root.Box.Width, layout.Root.Box.Height)
		}
	}

	// 比较
	fmt.Fprintf(os.Stderr, "\n=== Comparison ===\n")
	hasModalInTest1, _ := layouts[rtui.LayerModal]
	hasInspectorInTest1, _ := layouts[rtui.LayerInspector]
	hasModalInTest2, _ := layouts2[rtui.LayerModal]
	hasInspectorInTest2, _ := layouts2[rtui.LayerInspector]

	fmt.Fprintf(os.Stderr, "Test 1 (with Modal): hasModal=%v, hasInspector=%v\n", hasModalInTest1, hasInspectorInTest1)
	fmt.Fprintf(os.Stderr, "Test 2 (with Inspector): hasModal=%v, hasInspector=%v\n", hasModalInTest2, hasInspectorInTest2)
}
