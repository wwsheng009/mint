package ui_test

import (
	"testing"

	"github.com/wwsheng009/mint/components/basic"
	"github.com/wwsheng009/mint/internal/render"
	"github.com/wwsheng009/mint/runtime/layout"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	newbutton "github.com/wwsheng009/mint/ui/components/button"
)

// =============================================================================
// Fiber + New Button 完整集成测试
// =============================================================================

// TestFiberNewButton_Integration 测试 Fiber 与新 button 的完整集成
func TestFiberNewButton_Integration(t *testing.T) {
	t.Log("=== Fiber + New Button 集成测试 ===")

	// 1. 创建新 button VNode
	btn := newbutton.New("Submit").
		SetVariant(newbutton.VariantPrimary).
		SetSize(newbutton.SizeMedium)

	t.Log("Step 1: 创建 VNode ✓")

	// 2. 验证 VNode 实现 InstanceFactory 接口
	var _ rtui.InstanceFactory = btn
	t.Log("Step 2: VNode 实现 InstanceFactory ✓")

	// 3. 从 VNode 创建 Fiber
	fiber := rtui.CreateFiberFromVNode(btn)
	if fiber == nil {
		t.Fatal("CreateFiberFromVNode 返回 nil")
	}
	t.Log("Step 3: 创建 Fiber ✓")

	// 4. 验证 Fiber 捕获了正确的信息
	if fiber.Tag != "button" {
		t.Errorf("Fiber.Tag = %q, want %q", fiber.Tag, "button")
	}
	if fiber.Type != rtui.VNodeElement {
		t.Errorf("Fiber.Type = %d, want %d (VNodeElement)", fiber.Type, rtui.VNodeElement)
	}
	t.Log("Step 4: Fiber 捕获正确信息 ✓")

	// 5. 验证 Fiber 创建了 Instance
	if fiber.Instance == nil {
		t.Fatal("Fiber.Instance 为 nil，期望 ButtonInstance")
	}
	t.Log("Step 5: Fiber 创建了 Instance ✓")

	// 6. 验证 Instance 类型正确
	_, ok := fiber.Instance.(*newbutton.Instance)
	if !ok {
		t.Fatalf("Fiber.Instance 类型 = %T, 期望 *newbutton.Instance", fiber.Instance)
	}
	t.Log("Step 6: Instance 类型正确 ✓")

	// 7. 验证 Instance 实现了 Measure 方法
	measurable, ok := fiber.Instance.(interface {
		Measure(layout.Constraints) layout.Size
	})
	if !ok {
		t.Fatal("Instance 未实现 Measure(layout.Constraints) layout.Size")
	}
	t.Log("Step 7: Instance 实现 Measure ✓")

	// 8. 调用 Measure 验证计算
	constraints := layout.UnboundedConstraints()
	size := measurable.Measure(constraints)
	// "Submit" = 6 + 3 (brackets + focus) + 2 (medium) = 11
	if size.Width != 11 || size.Height != 1 {
		t.Errorf("Measure() = %dx%d, want 11x1", size.Width, size.Height)
	}
	t.Logf("Step 8: Measure 计算正确 (11x1) ✓")

	// 9. 创建 FiberToNodeAdapter
	adapter := render.NewFiberToNodeAdapterPure(fiber)
	t.Log("Step 9: 创建 FiberToNodeAdapter ✓")

	// 10. 通过 Adapter 调用 Measure
	adapterSize := adapter.Measure(constraints)
	if adapterSize.Width != size.Width || adapterSize.Height != size.Height {
		t.Errorf("Adapter.Measure() = %dx%d, want %dx%d",
			adapterSize.Width, adapterSize.Height, size.Width, size.Height)
	}
	t.Logf("Step 10: Adapter.Measure 正确调用 Instance.Measure ✓")

	t.Logf("\n✅ 完整集成测试通过！")
	t.Logf("   VNode → Fiber → Instance → Measure 全链路工作正常")
}

// TestFiberNewButton_MultipleButtons 测试多个 button 的 Fiber 树
func TestFiberNewButton_MultipleButtons(t *testing.T) {
	t.Log("=== 多 Button Fiber 树测试 ===")

	// 创建多个独立的 button
	okBtn := newbutton.New("OK").SetSize(newbutton.SizeSmall)
	cancelBtn := newbutton.New("Cancel").SetSize(newbutton.SizeMedium)
	submitBtn := newbutton.New("Submit").SetVariant(newbutton.VariantPrimary)

	// 分别创建 Fiber
	okFiber := rtui.CreateFiberFromVNode(okBtn)
	cancelFiber := rtui.CreateFiberFromVNode(cancelBtn)
	submitFiber := rtui.CreateFiberFromVNode(submitBtn)

	// 验证所有 button 都有 Instance
	if okFiber.Instance == nil {
		t.Error("OK Button Instance 为 nil")
	}
	if cancelFiber.Instance == nil {
		t.Error("Cancel Button Instance 为 nil")
	}
	if submitFiber.Instance == nil {
		t.Error("Submit Button Instance 为 nil")
	}

	// 测量每个 button
	okInst := okFiber.Instance.(*newbutton.Instance)
	cancelInst := cancelFiber.Instance.(*newbutton.Instance)
	submitInst := submitFiber.Instance.(*newbutton.Instance)

	constraints := layout.UnboundedConstraints()

	okSize := okInst.Measure(constraints)
	cancelSize := cancelInst.Measure(constraints)
	submitSize := submitInst.Measure(constraints)

	// 验证尺寸
	// OK: 2 + 3 = 5 (small)
	// Cancel: 6 + 3 + 2 = 11 (medium)
	// Submit: 6 + 3 + 2 = 11 (medium)
	if okSize.Width != 5 {
		t.Errorf("OK Button Width = %d, want 5", okSize.Width)
	}
	if cancelSize.Width != 11 {
		t.Errorf("Cancel Button Width = %d, want 11", cancelSize.Width)
	}
	if submitSize.Width != 11 {
		t.Errorf("Submit Button Width = %d, want 11", submitSize.Width)
	}

	t.Logf("OK Button: %dx%d", okSize.Width, okSize.Height)
	t.Logf("Cancel Button: %dx%d", cancelSize.Width, cancelSize.Height)
	t.Logf("Submit Button: %dx%d", submitSize.Width, submitSize.Height)

	t.Log("✅ 多 Button Fiber 树测试通过！")
}

// TestFiberNewButton_WithTextNodes 测试 button 与文本节点混合
func TestFiberNewButton_WithTextNodes(t *testing.T) {
	t.Log("=== Button + 文本节点混合测试 ===")

	// 创建文本节点
	title := basic.NewText("Dialog Title")
	description := basic.NewText("Please confirm your action")

	// 创建 button
	confirmBtn := newbutton.New("Confirm").SetVariant(newbutton.VariantPrimary)
	cancelBtn := newbutton.New("Cancel")

	// 创建 Fiber
	titleFiber := rtui.CreateFiberFromVNode(title)
	_ = rtui.CreateFiberFromVNode(description) // descFiber
	confirmFiber := rtui.CreateFiberFromVNode(confirmBtn)
	_ = rtui.CreateFiberFromVNode(cancelBtn) // cancelFiber

	// 验证文本节点
	if titleFiber.Type != rtui.VNodeText {
		t.Errorf("Title Fiber.Type = %d, want %d (VNodeText)", titleFiber.Type, rtui.VNodeText)
	}
	if titleFiber.MemoizedState != "Dialog Title" {
		t.Errorf("Title content = %v, want %q", titleFiber.MemoizedState, "Dialog Title")
	}

	// 验证 button 节点
	if confirmFiber.Type != rtui.VNodeElement {
		t.Errorf("Confirm Button Fiber.Type = %d, want %d (VNodeElement)",
			confirmFiber.Type, rtui.VNodeElement)
	}
	if confirmFiber.Instance == nil {
		t.Error("Confirm Button Instance 为 nil")
	}

	// 通过 Adapter 测量所有节点
	constraints := layout.UnboundedConstraints()

	titleAdapter := render.NewFiberToNodeAdapterPure(titleFiber)
	confirmAdapter := render.NewFiberToNodeAdapterPure(confirmFiber)

	titleSize := titleAdapter.Measure(constraints)
	confirmSize := confirmAdapter.Measure(constraints)

	// 验证测量结果
	if titleSize.Width != len("Dialog Title") {
		t.Errorf("Title Width = %d, want %d", titleSize.Width, len("Dialog Title"))
	}
	if confirmSize.Width == 0 {
		t.Error("Confirm Button Width = 0, expected non-zero")
	}

	t.Logf("Title: %dx%d", titleSize.Width, titleSize.Height)
	t.Logf("Confirm Button: %dx%d", confirmSize.Width, confirmSize.Height)

	t.Log("✅ Button + 文本节点混合测试通过！")
}

// TestFiberNewButton_MeasureAllButtons 测量所有 button 变体
func TestFiberNewButton_MeasureAllButtons(t *testing.T) {
	t.Log("=== 测量所有 Button 变体 ===")

	tests := []struct {
		label      string
		variant    newbutton.Variant
		size       newbutton.Size
		wantWidth  int
		wantHeight int
	}{
		// Small buttons: width = len(label) + 3 (brackets + focus)
		{"OK", newbutton.VariantDefault, newbutton.SizeSmall, 5, 1},   // 2 + 3 = 5
		{"Cancel", newbutton.VariantDefault, newbutton.SizeSmall, 9, 1}, // 6 + 3 = 9
		// Medium buttons: width = len(label) + 3 + 2 (medium padding)
		{"OK", newbutton.VariantPrimary, newbutton.SizeMedium, 7, 1},   // 2 + 3 + 2 = 7
		{"Cancel", newbutton.VariantPrimary, newbutton.SizeMedium, 11, 1}, // 6 + 3 + 2 = 11
		{"Submit", newbutton.VariantPrimary, newbutton.SizeMedium, 11, 1},  // 6 + 3 + 2 = 11
		// Large buttons: width = len(label) + 3 + 4 (large padding)
		{"OK", newbutton.VariantDanger, newbutton.SizeLarge, 9, 1},    // 2 + 3 + 4 = 9
		{"Cancel", newbutton.VariantDanger, newbutton.SizeLarge, 13, 1}, // 6 + 3 + 4 = 13
		{"Submit", newbutton.VariantDanger, newbutton.SizeLarge, 13, 1},  // 6 + 3 + 4 = 13
	}

	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			btn := newbutton.New(tt.label).
				SetVariant(tt.variant).
				SetSize(tt.size)

			fiber := rtui.CreateFiberFromVNode(btn)

			// 直接测量 Instance
			inst := fiber.Instance.(*newbutton.Instance)
			size := inst.Measure(layout.UnboundedConstraints())

			if size.Width != tt.wantWidth {
				t.Errorf("Width = %d, want %d", size.Width, tt.wantWidth)
			}
			if size.Height != tt.wantHeight {
				t.Errorf("Height = %d, want %d", size.Height, tt.wantHeight)
			}

			// 通过 Adapter 测量
			adapter := render.NewFiberToNodeAdapterPure(fiber)
			adapterSize := adapter.Measure(layout.UnboundedConstraints())

			if adapterSize.Width != size.Width || adapterSize.Height != size.Height {
				t.Errorf("Adapter mismatch: got %dx%d, want %dx%d",
					adapterSize.Width, adapterSize.Height, size.Width, size.Height)
			}
		})
	}
}

// TestFiberNewButton_ConstraintsPropagation 测试约束传播
func TestFiberNewButton_ConstraintsPropagation(t *testing.T) {
	t.Log("=== 约束传播测试 ===")

	btn := newbutton.New("Click Me") // 自然宽度 13
	fiber := rtui.CreateFiberFromVNode(btn)
	inst := fiber.Instance.(*newbutton.Instance)

	tests := []struct {
		name      string
		constraint layout.Constraints
		wantWidth int
	}{
		{"无约束", layout.UnboundedConstraints(), 13},
		{"紧约束 20x1", layout.TightConstraints(20, 1), 20},
		{"最大宽度 15", layout.Constraints{MaxWidth: 15, MaxHeight: 10}, 13},
		{"最大宽度 10", layout.Constraints{MaxWidth: 10, MaxHeight: 10}, 10},
		{"最小宽度 20", layout.Constraints{MinWidth: 20, MaxWidth: 50, MaxHeight: 10}, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size := inst.Measure(tt.constraint)
			if size.Width != tt.wantWidth {
				t.Errorf("Width = %d, want %d (constraint: %+v)", size.Width, tt.wantWidth, tt.constraint)
			}
		})
	}
}

// TestFiberNewButton_LayoutEngineWithButtons 测试布局引擎与 button 配合
func TestFiberNewButton_LayoutEngineWithButtons(t *testing.T) {
	t.Log("=== 布局引擎 + Button 测试 ===")

	// 创建多个 button 形成简单布局
	btn1 := newbutton.New("File")
	btn2 := newbutton.New("Edit")
	btn3 := newbutton.New("View")

	fiber1 := rtui.CreateFiberFromVNode(btn1)
	fiber2 := rtui.CreateFiberFromVNode(btn2)
	fiber3 := rtui.CreateFiberFromVNode(btn3)

	// 分别创建 adapter 并测量
	adapter1 := render.NewFiberToNodeAdapterPure(fiber1)
	adapter2 := render.NewFiberToNodeAdapterPure(fiber2)
	adapter3 := render.NewFiberToNodeAdapterPure(fiber3)

	constraints := layout.UnboundedConstraints()

	size1 := adapter1.Measure(constraints)
	size2 := adapter2.Measure(constraints)
	size3 := adapter3.Measure(constraints)

	t.Logf("File: %dx%d", size1.Width, size1.Height)
	t.Logf("Edit: %dx%d", size2.Width, size2.Height)
	t.Logf("View: %dx%d", size3.Width, size3.Height)

	// 验证所有 button 都有有效尺寸
	if size1.Width == 0 || size1.Height == 0 {
		t.Error("File button has zero size")
	}
	if size2.Width == 0 || size2.Height == 0 {
		t.Error("Edit button has zero size")
	}
	if size3.Width == 0 || size3.Height == 0 {
		t.Error("View button has zero size")
	}

	t.Log("✅ 布局引擎 + Button 测试通过！")
}
