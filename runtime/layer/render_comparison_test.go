package layer

import (
	"fmt"
	"os"
	"testing"

	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/compute"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestSingleLayerRendering 测试单层渲染（TUI_INSPECTOR=false 的路径）
func TestSingleLayerRendering(t *testing.T) {
	// 创建应用内容
	appContent := rtui.VStack(
		rtui.Bordered().Label("H1").Child(rtui.NewElement("Header 1")).Build(),
		rtui.Bordered().Label("H2").Child(rtui.NewElement("Header 2")).Build(),
		rtui.Bordered().Label("H3").Child(rtui.NewElement("Header 3")).Build(),
	)

	engine := compute.NewEngine()
	constraints := runtime.BoxConstraints{
		MinWidth:  0, MaxWidth: 120,
		MinHeight: 0, MaxHeight: 40,
	}

	// 单层布局
	layout, err := engine.Layout(appContent, constraints)
	if err != nil {
		t.Fatalf("Layout failed: %v", err)
	}

	fmt.Fprintf(os.Stderr, "\n=== Single Layer Layout ===\n")
	fmt.Fprintf(os.Stderr, "layout.Root.Box: (%d, %d, %dx%d)\n",
		layout.Root.Box.X, layout.Root.Box.Y,
		layout.Root.Box.Width, layout.Root.Box.Height)

	// 创建 buffer 并绘制
	buffer := paint.NewBuffer(120, 40)

	// 绘制（简化版，模拟 PaintEngine.Paint）
	paintText(buffer, 0, 0, "Header 1")
	paintText(buffer, 0, 1, "Header 2")
	paintText(buffer, 0, 2, "Header 3")

	// 检查 buffer 内容
	contentCount := countBufferContent(buffer)
	fmt.Fprintf(os.Stderr, "Buffer content: %d cells\n", contentCount)

	if contentCount == 0 {
		t.Error("❌ Buffer is empty after single layer rendering!")
	} else {
		fmt.Fprintf(os.Stderr, "✅ Single layer rendering: %d cells in buffer\n", contentCount)
	}
}

// TestMultiLayerRendering 测试多层渲染（TUI_INSPECTOR=true 的路径）
func TestMultiLayerRendering(t *testing.T) {
	// 创建应用内容
	appContent := rtui.VStack(
		rtui.Bordered().Label("H1").Child(rtui.NewElement("Header 1")).Build(),
		rtui.Bordered().Label("H2").Child(rtui.NewElement("Header 2")).Build(),
		rtui.Bordered().Label("H3").Child(rtui.NewElement("Header 3")).Build(),
	)

	// 创建 Inspector
	inspectorOverlay := rtui.Bordered().
		Label("INSPECTOR").
		Child(rtui.NewElement("Inspector Content")).
		Build()
	inspectorOverlay.SetLayer(rtui.LayerInspector)
	inspectorOverlay.SetProps(rtui.Props{"x": 40, "y": 5})

	// 使用 Fragment 包裹
	root := rtui.Fragment(appContent, inspectorOverlay)

	// 多层布局
	manager := NewManager()
	engine := compute.NewEngine()
	constraints := runtime.BoxConstraints{
		MinWidth:  0, MaxWidth: 120,
		MinHeight: 0, MaxHeight: 40,
	}

	if err := manager.CollectAndLayout(root, constraints, engine); err != nil {
		t.Fatalf("CollectAndLayout failed: %v", err)
	}

	layouts := manager.GetLayouts()

	fmt.Fprintf(os.Stderr, "\n=== Multi Layer Layout ===\n")
	fmt.Fprintf(os.Stderr, "Total layers: %d\n", len(layouts))

	baseLayout, hasBase := layouts[rtui.LayerBase]
	if !hasBase {
		t.Fatal("❌ No base layout!")
	}

	fmt.Fprintf(os.Stderr, "baseLayout.Root.Box: (%d, %d, %dx%d)\n",
		baseLayout.Root.Box.X, baseLayout.Root.Box.Y,
		baseLayout.Root.Box.Width, baseLayout.Root.Box.Height)

	inspectorLayout, hasInspector := layouts[rtui.LayerInspector]
	if !hasInspector {
		t.Fatal("❌ No inspector layout!")
	}

	fmt.Fprintf(os.Stderr, "inspectorLayout.Root.Box: (%d, %d, %dx%d)\n",
		inspectorLayout.Root.Box.X, inspectorLayout.Root.Box.Y,
		inspectorLayout.Root.Box.Width, inspectorLayout.Root.Box.Height)

	// 创建 buffer 并绘制
	buffer := paint.NewBuffer(120, 40)

	// 模拟 PaintLayers: 先绘制 base
	paintText(buffer, 0, 0, "Header 1")
	paintText(buffer, 0, 1, "Header 2")
	paintText(buffer, 0, 2, "Header 3")

	fmt.Fprintf(os.Stderr, "After base layer: %d cells\n", countBufferContent(buffer))

	// 再绘制 inspector（模拟覆盖）
	paintText(buffer, 40, 5, "Inspector Content")

	// 检查 buffer 内容
	contentCount := countBufferContent(buffer)
	fmt.Fprintf(os.Stderr, "After inspector layer: %d cells\n", contentCount)

	if contentCount == 0 {
		t.Error("❌ Buffer is empty after multi layer rendering!")
	} else {
		fmt.Fprintf(os.Stderr, "✅ Multi layer rendering: %d cells in buffer\n", contentCount)
	}
}

// 辅助函数
func paintText(buffer *paint.Buffer, x, y int, text string) {
	if x >= 0 && x < buffer.Width && y >= 0 && y < buffer.Height {
		for i, r := range text {
			if x+i < buffer.Width {
				buffer.Cells[y][x+i].Cluster = string(r)
			}
		}
	}
}

func countBufferContent(buffer *paint.Buffer) int {
	count := 0
	for y := 0; y < buffer.Height; y++ {
		for x := 0; x < buffer.Width; x++ {
			if buffer.Cells[y][x].Cluster != "" {
				count++
			}
		}
	}
	return count
}
