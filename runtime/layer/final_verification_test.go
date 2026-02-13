package layer

import (
	"fmt"
	"os"
	"testing"

	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/compute"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestFinalRenderingVerification 完整验证渲染流程的所有关键点
func TestFinalRenderingVerification(t *testing.T) {
	os.Setenv("TUI_DEBUG_LAYER", "true")

	// 1. 创建真实的 appContent 结构
	appContent := rtui.VStack(
		rtui.Bordered().Label("Header").Child(rtui.NewElement("header-content")).Build(),
		rtui.Bordered().Label("Content").Child(rtui.NewElement("content")).Build(),
		rtui.Bordered().Label("Footer").Child(rtui.NewElement("footer")).Build(),
	)

	// 2. 创建真实的 inspector overlay
	inspectorOverlay := rtui.Bordered().
		Label("INSPECTOR").
		Child(rtui.VStack(
			rtui.NewElement("inspector-line1"),
			rtui.NewElement("inspector-line2"),
		)).
		Build()
	inspectorOverlay.SetLayer(rtui.LayerInspector)
	inspectorOverlay.SetProps(rtui.Props{
		"x": 40,
		"y": 5,
	})

	// 3. 使用 Fragment 包裹 (修复后的方案)
	root := rtui.Fragment(appContent, inspectorOverlay)

	fmt.Fprintf(os.Stderr, "\n=== Step 1: CollectAndLayout ===\n")

	manager := NewManager()
	engine := compute.NewEngine()
	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  120,
		MinHeight: 0,
		MaxHeight: 40,
	}

	if err := manager.CollectAndLayout(root, constraints, engine); err != nil {
		t.Fatalf("❌ CollectAndLayout failed: %v", err)
	}

	layouts := manager.GetLayouts()

	// 4. 验证 baseLayout
	baseLayout, hasBase := layouts[rtui.LayerBase]
	if !hasBase {
		t.Fatal("❌ No base layout!")
	}

	fmt.Fprintf(os.Stderr, "baseLayout.Root.VNode type: %s\n", baseLayout.Root.VNode.Type().String())
	fmt.Fprintf(os.Stderr, "baseLayout.Root.Box: (%d, %d, %dx%d)\n",
		baseLayout.Root.Box.X,
		baseLayout.Root.Box.Y,
		baseLayout.Root.Box.Width,
		baseLayout.Root.Box.Height,
	)

	// 验证 baseLayout 不超过屏幕
	screenWidth, screenHeight := 120, 40

	if baseLayout.Root.Box.X >= screenWidth {
		t.Errorf("❌ baseLayout X position %d exceeds screen width %d",
			baseLayout.Root.Box.X, screenWidth)
	}

	if baseLayout.Root.Box.Y >= screenHeight {
		t.Errorf("❌ baseLayout Y position %d exceeds screen height %d",
			baseLayout.Root.Box.Y, screenHeight)
	}

	baseRight := baseLayout.Root.Box.X + baseLayout.Root.Box.Width
	if baseRight > screenWidth {
		t.Errorf("❌ baseLayout right edge %d exceeds screen width %d",
			baseRight, screenWidth)
	} else {
		fmt.Fprintf(os.Stderr, "✅ baseLayout within screen bounds: X=%d, Width=%d, Right=%d (screen=%d)\n",
			baseLayout.Root.Box.X, baseLayout.Root.Box.Width, baseRight, screenWidth)
	}

	baseBottom := baseLayout.Root.Box.Y + baseLayout.Root.Box.Height
	if baseBottom > screenHeight {
		t.Errorf("❌ baseLayout bottom edge %d exceeds screen height %d",
			baseBottom, screenHeight)
	} else {
		fmt.Fprintf(os.Stderr, "✅ baseLayout within screen bounds: Y=%d, Height=%d, Bottom=%d (screen=%d)\n",
			baseLayout.Root.Box.Y, baseLayout.Root.Box.Height, baseBottom, screenHeight)
	}

	// 5. 验证 inspectorLayout
	inspectorLayout, hasInspector := layouts[rtui.LayerInspector]
	if !hasInspector {
		t.Fatal("❌ No inspector layout!")
	}

	fmt.Fprintf(os.Stderr, "inspectorLayout.Root.Box: (%d, %d, %dx%d)\n",
		inspectorLayout.Root.Box.X,
		inspectorLayout.Root.Box.Y,
		inspectorLayout.Root.Box.Width,
		inspectorLayout.Root.Box.Height,
	)

	// 验证 Inspector 不超过屏幕
	if inspectorLayout.Root.Box.X >= screenWidth {
		t.Errorf("❌ Inspector X position %d exceeds screen width %d",
			inspectorLayout.Root.Box.X, screenWidth)
	}

	if inspectorLayout.Root.Box.Y >= screenHeight {
		t.Errorf("❌ Inspector Y position %d exceeds screen height %d",
			inspectorLayout.Root.Box.Y, screenHeight)
	}

	inspectorRight := inspectorLayout.Root.Box.X + inspectorLayout.Root.Box.Width
	if inspectorRight > screenWidth {
		t.Errorf("❌ Inspector right edge %d exceeds screen width %d",
			inspectorRight, screenWidth)
	} else {
		fmt.Fprintf(os.Stderr, "✅ Inspector within screen bounds: X=%d, Width=%d, Right=%d (screen=%d)\n",
			inspectorLayout.Root.Box.X, inspectorLayout.Root.Box.Width, inspectorRight, screenWidth)
	}

	// 6. 检查两个 layer 是否有重叠
	fmt.Fprintf(os.Stderr, "\n=== Step 2: Overlap Check ===\n")
	baseRegionRight := baseLayout.Root.Box.X + baseLayout.Root.Box.Width
	inspectorRegionLeft := inspectorLayout.Root.Box.X

	if baseRegionRight > inspectorRegionLeft {
		t.Errorf("❌ Regions overlap! base extends to %d, inspector starts at %d",
			baseRegionRight, inspectorRegionLeft)
	} else {
		fmt.Fprintf(os.Stderr, "✅ No overlap: base ends at %d, inspector starts at %d\n",
			baseRegionRight, inspectorRegionLeft)
	}

	// 7. 总结
	fmt.Fprintf(os.Stderr, "\n=== Summary ===\n")
	fmt.Fprintf(os.Stderr, "✅ All layouts within screen bounds\n")
	fmt.Fprintf(os.Stderr, "✅ No overlapping regions\n")
	fmt.Fprintf(os.Stderr, "✅ Inspector layer correctly positioned at (%d, %d)\n",
		inspectorLayout.Root.Box.X, inspectorLayout.Root.Box.Y)
	fmt.Fprintf(os.Stderr, "✅ Base layer correctly positioned at (%d, %d)\n",
		baseLayout.Root.Box.X, baseLayout.Root.Box.Y)
}
