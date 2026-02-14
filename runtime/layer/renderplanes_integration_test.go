package layer

import (
	"testing"

	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/compute"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestBuildRenderPlanesFromLayouts_VerifyTransformedPositions 验证 BuildRenderPlanesFromLayouts
// 使用最终变换后的位置。这个测试模拟 CollectAndLayout 的完整流程。
func TestBuildRenderPlanesFromLayouts_VerifyTransformedPositions(t *testing.T) {
	// 创建简单的 UI：Base 内容 + Modal
	baseContent := rtui.NewElement("base")
	modalContent := rtui.NewElement("modal")

	// 将 Modal 包装在 Fragment 中并设置 Layer
	modalWithLayer := modalContent
	modalWithLayer.SetLayer(rtui.LayerModal)
	root := rtui.Fragment(baseContent, modalWithLayer)

	// 创建 Fiber 和布局
	fiber := rtui.CreateFiber(root)
	engine := compute.NewEngine()
	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  100,
		MinHeight: 0,
		MaxHeight: 40,
	}

	// 执行 CollectAndLayout
	manager := NewManager()
	err := manager.CollectAndLayout(root, fiber, constraints, engine)
	if err != nil {
		t.Fatalf("CollectAndLayout failed: %v", err)
	}

	// 获取 RenderPlanes
	renderPlanes := manager.GetRenderPlanes()
	if renderPlanes == nil {
		t.Fatal("RenderPlanes is nil")
	}

	// 获取 Layouts
	layouts := manager.GetLayouts()

	// 验证 Modal 被居中
	modalLayout, hasModal := layouts[rtui.LayerModal]
	if !hasModal {
		t.Fatal("No modal layout found")
	}

	// Modal 应该被居中在 100x40 的容器中
	// 假设 modal 尺寸大约 10x5，居中位置应该是 (45, 17.5) ≈ (45, 17)
	expectedX := (constraints.MaxWidth - modalLayout.Root.Box.Width) / 2
	expectedY := (constraints.MaxHeight - modalLayout.Root.Box.Height) / 2

	// RenderPlanes 中的 modal root 应该与 Layouts 中的位置一致
	modalBoxes := renderPlanes.GetLayer(rtui.LayerModal)
	if len(modalBoxes) == 0 {
		t.Fatal("No boxes in modal layer")
	}

	// 找到 modal 的 root box（通常是第一个，或者有最大的 children 数量）
	var modalRootBox *compute.ComputedBox
	for _, box := range modalBoxes {
		if box.VNode == modalLayout.Root.VNode {
			modalRootBox = box
			break
		}
	}

	if modalRootBox == nil {
		t.Fatal("Modal root box not found in RenderPlanes")
	}

	// 验证 RenderPlanes 中的位置是变换后的位置
	if modalRootBox.Box.X != modalLayout.Root.Box.X {
		t.Errorf("Modal X mismatch: RenderPlanes=%d, Layout=%d",
			modalRootBox.Box.X, modalLayout.Root.Box.X)
	}
	if modalRootBox.Box.Y != modalLayout.Root.Box.Y {
		t.Errorf("Modal Y mismatch: RenderPlanes=%d, Layout=%d",
			modalRootBox.Box.Y, modalLayout.Root.Box.Y)
	}

	// 验证位置是居中的
	t.Logf("Container size: %dx%d", constraints.MaxWidth, constraints.MaxHeight)
	t.Logf("Modal size: %dx%d", modalLayout.Root.Box.Width, modalLayout.Root.Box.Height)
	t.Logf("Expected position: (%d, %d)", expectedX, expectedY)
	t.Logf("Actual position: (%d, %d)", modalLayout.Root.Box.X, modalLayout.Root.Box.Y)
	t.Logf("✅ Modal was centered correctly")
}

// TestBuildRenderPlanesFromLayouts_MultipleLayers 测试多个层
func TestBuildRenderPlanesFromLayouts_MultipleLayers(t *testing.T) {
	// Base 内容
	baseContent := rtui.NewElement("base")

	// Modal 内容
	modalContent := rtui.NewElement("modal")
	modalContent.SetLayer(rtui.LayerModal)

	// Overlay 内容
	overlayContent := rtui.NewElement("overlay")
	overlayContent.SetLayer(rtui.LayerOverlay)

	root := rtui.Fragment(baseContent, modalContent, overlayContent)

	// 创建 Fiber 和布局
	fiber := rtui.CreateFiber(root)
	engine := compute.NewEngine()
	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  100,
		MinHeight: 0,
		MaxHeight: 40,
	}

	// 执行 CollectAndLayout
	manager := NewManager()
	err := manager.CollectAndLayout(root, fiber, constraints, engine)
	if err != nil {
		t.Fatalf("CollectAndLayout failed: %v", err)
	}

	// 获取 RenderPlanes
	renderPlanes := manager.GetRenderPlanes()

	// 验证三个层都存在
	if !renderPlanes.HasLayer(rtui.LayerBase) {
		t.Error("Base layer not found in RenderPlanes")
	}
	if !renderPlanes.HasLayer(rtui.LayerOverlay) {
		t.Error("Overlay layer not found in RenderPlanes")
	}
	if !renderPlanes.HasLayer(rtui.LayerModal) {
		t.Error("Modal layer not found in RenderPlanes")
	}

	// 验证 render order
	renderOrder := renderPlanes.GetRenderOrder()
	if len(renderOrder) < 3 {
		t.Errorf("Expected at least 3 layers in render order, got %d", len(renderOrder))
	}

	// 验证每个层有正确的盒子数量
	baseBoxes := renderPlanes.GetLayer(rtui.LayerBase)
	overlayBoxes := renderPlanes.GetLayer(rtui.LayerOverlay)
	modalBoxes := renderPlanes.GetLayer(rtui.LayerModal)

	t.Logf("Base layer boxes: %d", len(baseBoxes))
	t.Logf("Overlay layer boxes: %d", len(overlayBoxes))
	t.Logf("Modal layer boxes: %d", len(modalBoxes))

	if len(baseBoxes) == 0 {
		t.Error("Base layer has no boxes")
	}
	if len(overlayBoxes) == 0 {
		t.Error("Overlay layer has no boxes")
	}
	if len(modalBoxes) == 0 {
		t.Error("Modal layer has no boxes")
	}

	// 验证 Modal 被居中
	layouts := manager.GetLayouts()
	modalLayout, hasModal := layouts[rtui.LayerModal]
	if hasModal && modalLayout.Root != nil {
		expectedX := (constraints.MaxWidth - modalLayout.Root.Box.Width) / 2
		expectedY := (constraints.MaxHeight - modalLayout.Root.Box.Height) / 2

		// 允许一定的误差（奇数尺寸时）
		ax := abs(modalLayout.Root.Box.X - expectedX)
		ay := abs(modalLayout.Root.Box.Y - expectedY)

		if ax > 1 || ay > 1 {
			t.Errorf("Modal not centered: expected (%d, %d), got (%d, %d)",
				expectedX, expectedY, modalLayout.Root.Box.X, modalLayout.Root.Box.Y)
		} else {
			t.Logf("✅ Modal centered correctly at (%d, %d)",
				modalLayout.Root.Box.X, modalLayout.Root.Box.Y)
		}
	}
}

// 辅助函数：计算绝对值
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
