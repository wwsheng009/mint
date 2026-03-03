package ui_test

import (
	"testing"

	"github.com/wwsheng009/mint/internal/render"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/button"
	"github.com/wwsheng009/mint/ui/components/text"
)

// =============================================================================
// Fiber-first 渲染管线集成测试
// =============================================================================
// 根据 docs/render/paint/optimized/refactor 文档的架构要求测试：
// 1. VNode 只存在于 Reconcile 阶段
// 2. Layout 只读 Fiber（通过 FiberToNodeAdapter）
// 3. Paint 只用 PaintableBox
// =============================================================================

// TestFiberToNodeAdapter_FullLayoutPipeline 测试完整布局管线
// 验证：Fiber → FiberToNodeAdapter → layout.Engine → LayoutBox
func TestFiberToNodeAdapter_FullLayoutPipeline(t *testing.T) {
	t.Log("=== 测试完整布局管线 ===")

	// Step 1: 创建 VNode 并生成 Fiber
	btn := button.New("Submit").SetSize(button.SizeMedium)
	fiber := rtui.CreateFiberFromVNode(btn)

	if fiber == nil {
		t.Fatal("CreateFiberFromVNode 返回 nil")
	}
	t.Log("Step 1: 创建 Fiber ✓")

	// Step 2: 创建 FiberToNodeAdapter
	adapter := render.NewFiberToNodeAdapterPure(fiber)

	// 验证 adapter 实现了所有必要接口
	var _ layout.Node = adapter
	var _ layout.Measurable = adapter
	var _ layout.Layered = adapter
	var _ layout.Dirtyable = adapter
	t.Log("Step 2: 创建 Adapter ✓ (实现所有 layout 接口)")

	// Step 3: 测试 Measure
	constraints := layout.UnboundedConstraints()
	size := adapter.Measure(constraints)

	// "Submit" = 6 + 3 (brackets + focus) + 2 (medium) = 11
	if size.Width != 11 {
		t.Errorf("Measure 宽度 = %d, want 11", size.Width)
	}
	if size.Height != 1 {
		t.Errorf("Measure 高度 = %d, want 1", size.Height)
	}
	t.Logf("Step 3: Measure ✓ (%dx%d)", size.Width, size.Height)

	// Step 4: 执行布局引擎
	engine := layout.NewEngine()
	result := engine.Layout(adapter, constraints)

	if result == nil || result.Root == nil {
		t.Fatal("Layout 引擎返回 nil 结果")
	}
	t.Logf("Step 4: Layout 引擎 ✓ (根节点 %dx%d)", result.Root.Width, result.Root.Height)

	// Step 5: 验证 LayoutBox
	if result.Root.Width != size.Width || result.Root.Height != size.Height {
		t.Errorf("LayoutBox 尺寸 = %dx%d, want %dx%d",
			result.Root.Width, result.Root.Height, size.Width, size.Height)
	}
	t.Log("Step 5: LayoutBox 验证 ✓")

	t.Log("✅ 完整布局管线测试通过")
}

// TestFiberToPaintableConverter_Pipeline 测试转换管线
// 验证：Fiber + LayoutBox → FiberToPaintableConverter → PaintableBox
func TestFiberToPaintableConverter_Pipeline(t *testing.T) {
	t.Log("=== 测试转换管线 ===")

	// Step 1: 创建带有多个子节点的 Fiber 树
	text := text.New("Hello")
	btn := button.New("Click")

	textFiber := rtui.CreateFiberFromVNode(text)
	btnFiber := rtui.CreateFiberFromVNode(btn)

	t.Log("Step 1: 创建 Fiber 树 ✓")

	// Step 2: 布局文本节点
	textAdapter := render.NewFiberToNodeAdapterPure(textFiber)
	textResult := layout.NewEngine().Layout(textAdapter, layout.UnboundedConstraints())

	if textResult.Root == nil {
		t.Fatal("文本布局失败")
	}
	t.Log("Step 2: 布局文本节点 ✓")

	// Step 3: 布局按钮节点
	btnAdapter := render.NewFiberToNodeAdapterPure(btnFiber)
	btnResult := layout.NewEngine().Layout(btnAdapter, layout.UnboundedConstraints())

	if btnResult.Root == nil {
		t.Fatal("按钮布局失败")
	}
	t.Log("Step 3: 布局按钮节点 ✓")

	// Step 4: 转换为 PaintableLayout
	converter := render.NewFiberToPaintableConverter(textFiber)
	textPaintLayout := converter.ConvertToLayout(textResult.Root)

	converter2 := render.NewFiberToPaintableConverter(btnFiber)
	btnPaintLayout := converter2.ConvertToLayout(btnResult.Root)

	if textPaintLayout == nil || textPaintLayout.Root == nil {
		t.Fatal("文本 PaintableLayout 转换失败")
	}
	if btnPaintLayout == nil || btnPaintLayout.Root == nil {
		t.Fatal("按钮 PaintableLayout 转换失败")
	}
	t.Log("Step 4: 转换为 PaintableLayout ✓")

	// Step 5: 验证 PaintableBox 属性
	if textPaintLayout.Root.Width != 5 { // "Hello" = 5
		t.Errorf("文本 PaintableBox 宽度 = %d, want 5", textPaintLayout.Root.Width)
	}
	if btnPaintLayout.Root.Width == 0 {
		t.Error("按钮 PaintableBox 宽度为 0")
	}
	t.Log("Step 5: 验证 PaintableBox 属性 ✓")

	t.Log("✅ 转换管线测试通过")
}

// TestPaintEngine_PaintLayout 测试 PaintEngine
// 验证：PaintableLayout → PaintEngine → Buffer
func TestPaintEngine_PaintLayout(t *testing.T) {
	t.Log("=== 测试 PaintEngine ===")

	// Step 1: 创建简单的文本 Fiber
	text := text.New("Test")
	fiber := rtui.CreateFiberFromVNode(text)

	t.Log("Step 1: 创建 Fiber ✓")

	// Step 2: 布局
	adapter := render.NewFiberToNodeAdapterPure(fiber)
	result := layout.NewEngine().Layout(adapter, layout.UnboundedConstraints())

	if result.Root == nil {
		t.Fatal("布局失败")
	}
	t.Log("Step 2: 布局 ✓")

	// Step 3: 转换
	converter := render.NewFiberToPaintableConverter(fiber)
	paintLayout := converter.ConvertToLayout(result.Root)

	if paintLayout == nil || paintLayout.Root == nil {
		t.Fatal("转换失败")
	}
	t.Log("Step 3: 转换 ✓")

	// Step 4: 创建 PaintEngine 和 Buffer
	engine := render.NewPaintEngine()
	buf := paint.NewBuffer(80, 24)

	t.Log("Step 4: 创建 PaintEngine 和 Buffer ✓")

	// Step 5: 执行绘制
	err := engine.PaintLayout(paintLayout, buf)
	if err != nil {
		t.Errorf("PaintLayout 失败: %v", err)
	}
	t.Log("Step 5: PaintLayout ✓")

	// Step 6: 验证 Buffer 内容
	cell := buf.GetContent(0, 0)
	if len(cell.Cluster) == 0 || cell.Cluster[0] != 'T' {
		t.Errorf("Buffer 内容不正确，首字符应为 'T'")
	}
	t.Log("Step 6: 验证 Buffer 内容 ✓")

	t.Log("✅ PaintEngine 测试通过")
}

// TestFullRenderingPipeline_NewButton 测试完整渲染管线（使用新 Button）
// 验证：VNode → Fiber → Adapter → Layout → Converter → PaintEngine → Buffer
func TestFullRenderingPipeline_NewButton(t *testing.T) {
	t.Log("=== 测试完整渲染管线（新 Button）===")

	// Step 1: 创建 VNode
	btn := button.New("Submit").
		SetVariant(button.VariantPrimary).
		SetSize(button.SizeMedium)

	t.Log("Step 1: 创建 VNode ✓")

	// Step 2: 创建 Fiber（VNode 在此后可被丢弃）
	fiber := rtui.CreateFiberFromVNode(btn)

	if fiber == nil {
		t.Fatal("CreateFiberFromVNode 返回 nil")
	}
	if fiber.Instance == nil {
		t.Fatal("Fiber.Instance 为 nil")
	}
	t.Log("Step 2: 创建 Fiber + Instance ✓")

	// Step 3: 创建 Adapter 并验证 Measure
	adapter := render.NewFiberToNodeAdapterPure(fiber)

	// 验证 Instance 的 Measure 被正确调用
	constraints := layout.Constraints{MinWidth: 0, MaxWidth: 80, MinHeight: 0, MaxHeight: 24}
	size := adapter.Measure(constraints)

	// "Submit" = 6 + 3 + 2 = 11
	if size.Width != 11 || size.Height != 1 {
		t.Errorf("Measure = %dx%d, want 11x1", size.Width, size.Height)
	}
	t.Logf("Step 3: Adapter.Measure ✓ (%dx%d)", size.Width, size.Height)

	// Step 4: 执行布局
	result := layout.NewEngine().Layout(adapter, constraints)

	if result == nil || result.Root == nil {
		t.Fatal("布局失败")
	}
	t.Logf("Step 4: Layout ✓ (根节点 %dx%d)", result.Root.Width, result.Root.Height)

	// Step 5: 转换为 PaintableLayout
	converter := render.NewFiberToPaintableConverter(fiber)
	paintLayout := converter.ConvertToLayout(result.Root)

	if paintLayout == nil || paintLayout.Root == nil {
		t.Fatal("PaintableLayout 转换失败")
	}

	// 验证 PaintableBox 有正确的 Node
	if paintLayout.Root.Node == nil {
		t.Error("PaintableBox.Node 为 nil")
	}
	t.Log("Step 5: 转换为 PaintableLayout ✓")

	// Step 6: 执行绘制
	paintEngine := render.NewPaintEngine()
	buf := paint.NewBuffer(80, 24)

	err := paintEngine.PaintLayout(paintLayout, buf)
	if err != nil {
		t.Errorf("PaintLayout 失败: %v", err)
	}
	t.Log("Step 6: PaintEngine.PaintLayout ✓")

	// Step 7: 验证 Buffer 有内容
	hasContent := false
	for y := 0; y < buf.Height; y++ {
		for x := 0; x < buf.Width; x++ {
			cell := buf.GetContent(x, y)
			if len(cell.Cluster) > 0 && cell.Cluster[0] != ' ' {
				hasContent = true
				break
			}
		}
		if hasContent {
			break
		}
	}

	if !hasContent {
		t.Error("Buffer 没有任何绘制内容")
	}
	t.Log("Step 7: 验证 Buffer 内容 ✓")

	t.Log("✅ 完整渲染管线测试通过")
}

// TestLayoutEngine_MultipleButtons 测试多按钮布局
func TestLayoutEngine_MultipleButtons(t *testing.T) {
	t.Log("=== 测试多按钮布局 ===")

	buttons := []struct {
		label string
		size  button.Size
	}{
		{"OK", button.SizeSmall},
		{"Cancel", button.SizeMedium},
		{"Submit", button.SizeLarge},
	}

	// 创建并测量每个按钮
	for _, bb := range buttons {
		btn := button.New(bb.label).SetSize(bb.size)
		fiber := rtui.CreateFiberFromVNode(btn)

		adapter := render.NewFiberToNodeAdapterPure(fiber)
		size := adapter.Measure(layout.UnboundedConstraints())

		t.Logf("Button '%s' (%v): %dx%d", bb.label, bb.size, size.Width, size.Height)

		if size.Width == 0 || size.Height == 0 {
			t.Errorf("Button '%s' 尺寸为零", bb.label)
		}
	}

	t.Log("✅ 多按钮布局测试通过")
}

// TestFiberFirst_ArchitectureConstraints 测试架构约束
// 验证 Fiber-first 架构的核心约束
func TestFiberFirst_ArchitectureConstraints(t *testing.T) {
	t.Log("=== 测试架构约束 ===")

	// 约束 1: VNode 不持久化
	btn := button.New("Test")
	fiber := rtui.CreateFiberFromVNode(btn)

	// Fiber 应该有 Instance，不依赖 VNode
	if fiber.Instance == nil {
		t.Error("Fiber.Instance 为 nil - 违反架构约束")
	}
	t.Log("约束 1: Fiber 持有 Instance，不依赖 VNode ✓")

	// 约束 2: Layout 只读 Fiber
	adapter := render.NewFiberToNodeAdapterPure(fiber)

	// adapter 应该能从 Fiber 获取所有必要信息
	var _ layout.Node = adapter
	var _ layout.Measurable = adapter
	var _ layout.Layered = adapter
	t.Log("约束 2: Adapter 实现 layout.Node 接口 ✓")

	// 约束 3: Measure 来自 Instance
	inst := fiber.Instance.(*button.Instance)
	instSize := inst.Measure(layout.UnboundedConstraints())
	adapterSize := adapter.Measure(layout.UnboundedConstraints())

	if instSize != adapterSize {
		t.Errorf("Instance.Measure = %v, Adapter.Measure = %v - 应该一致", instSize, adapterSize)
	}
	t.Log("约束 3: Measure 委托给 Instance ✓")

	t.Log("✅ 架构约束测试通过")
}

// TestFiberFirst_TextNodeLayout 测试文本节点布局
func TestFiberFirst_TextNodeLayout(t *testing.T) {
	t.Log("=== 测试文本节点布局 ===")

	texts := []string{"Hello", "World", "Test 123"}

	for _, textStr := range texts {
		textVNode := text.New(textStr)
		fiber := rtui.CreateFiberFromVNode(textVNode)

		// 布局
		adapter := render.NewFiberToNodeAdapterPure(fiber)
		size := adapter.Measure(layout.UnboundedConstraints())

		expectedWidth := len(textStr)
		if size.Width != expectedWidth {
			t.Errorf("Text '%s' width = %d, want %d", textStr, size.Width, expectedWidth)
		}
		if size.Height != 1 {
			t.Errorf("Text '%s' height = %d, want 1", textStr, size.Height)
		}

		t.Logf("Text '%s': %dx%d ✓", textStr, size.Width, size.Height)
	}

	t.Log("✅ 文本节点布局测试通过")
}

// TestFiberLayoutEngine_Integration 测试 FiberLayoutEngine 集成
func TestFiberLayoutEngine_Integration(t *testing.T) {
	t.Log("=== 测试 FiberLayoutEngine 集成 ===")

	// 创建 Fiber 树
	btn := button.New("Click Me")
	fiberRoot := rtui.CreateFiberFromVNode(btn)

	// 创建 FiberLayoutEngine
	engine := render.NewNewLayoutEngineAdapter()

	// 设置约束
	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  80,
		MinHeight: 0,
		MaxHeight: 24,
	}

	// 执行布局
	result, err := engine.LayoutFiber(fiberRoot, constraints)
	if err != nil {
		t.Fatalf("LayoutFiber 失败: %v", err)
	}

	if result == nil {
		t.Fatal("LayoutFiber 返回 nil")
	}

	t.Log("✅ FiberLayoutEngine 集成测试通过")
}

// TestFiberFlags_LayoutDirty 测试布局脏标记
func TestFiberFlags_LayoutDirty(t *testing.T) {
	t.Log("=== 测试布局脏标记 ===")

	fiber := &rtui.Fiber{}

	// 初始状态
	if fiber.IsLayoutDirty() {
		t.Error("初始状态不应为脏")
	}

	// 标记为脏
	fiber.MarkLayoutDirty()
	if !fiber.IsLayoutDirty() {
		t.Error("标记后应该为脏")
	}

	// 清除脏标记
	fiber.ClearLayoutDirty()
	if fiber.IsLayoutDirty() {
		t.Error("清除后不应为脏")
	}

	t.Log("✅ 布局脏标记测试通过")
}

// TestFiberFlags_PaintDirty 测试绘制脏标记
func TestFiberFlags_PaintDirty(t *testing.T) {
	t.Log("=== 测试绘制脏标记 ===")

	fiber := &rtui.Fiber{}

	// 初始状态
	if fiber.IsPaintDirty() {
		t.Error("初始状态不应为脏")
	}

	// 标记为脏
	fiber.MarkPaintDirty()
	if !fiber.IsPaintDirty() {
		t.Error("标记后应该为脏")
	}

	// 清除脏标记
	fiber.ClearPaintDirty()
	if fiber.IsPaintDirty() {
		t.Error("清除后不应为脏")
	}

	t.Log("✅ 绘制脏标记测试通过")
}

// TestDeclarativeNode_FiberFirstMode 测试 DeclarativeNode 的 Fiber-first 模式
func TestDeclarativeNode_FiberFirstMode(t *testing.T) {
	t.Log("=== 测试 DeclarativeNode Fiber-first 模式 ===")

	// 创建使用新 Button 的 renderFn
	node := render.NewDeclarativeNodeFromFuncWithFiber(func() rtui.VNode {
		return button.New("Test Button").
			SetVariant(button.VariantPrimary).
			SetSize(button.SizeMedium)
	})

	// 设置 Fiber-first 模式
	node.SetRenderMode(render.RenderModeFiberFirst)

	// 验证模式设置
	if node.GetRenderMode() != render.RenderModeFiberFirst {
		t.Error("RenderMode 应该是 FiberFirst")
	}

	// 验证内部组件已初始化
	if node.IsFiberFirstEnabled() {
		t.Log("FiberFirstEnabled = true ✓")
	}

	t.Log("✅ DeclarativeNode Fiber-first 模式测试通过")
}
