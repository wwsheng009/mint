// Layer 系统架构优化方案 - 示例代码
//
// 这个文件展示了如何优化当前的 Layer 系统，
// 解决用户提出的四个核心问题

package main

import (
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// 问题 1: Inspector 过多干涉渲染层
// =============================================================================

// ❌ 当前方案 (问题)
func CurrentApproach_InspectorInterferes() ui.VNode {
	inspectorOverlay := globalInspector.RenderOverlay()

	// Inspector 需要知道 Layer 系统
	inspectorOverlay.SetLayer(rtui.LayerInspector)  // ← 耦合！

	return ui.Fragment(appContent, inspectorOverlay)
}

// ✅ 方案 A: 外部标记 Layer
func SolutionA_ExternalLayerMarking() ui.VNode {
	// Inspector 不需要知道 Layer 系统
	inspectorContent := globalInspector.RenderOverlay()

	// 由应用层决定 Inspector 是什么 layer
	MarkAsLayer(inspectorContent, rtui.LayerInspector)

	return ui.Fragment(appContent, inspectorContent)
}

// Layer 标记辅助函数
func MarkAsLayer(vnode rtui.VNode, layer rtui.Layer) {
	vnode.SetLayer(layer)
}

// ✅ 方案 B: 使用 Layer 容器组件
func SolutionB_LayerContainer() ui.VNode {
	return ui.VStack(
		appContent,
		NewInspectorLayer(
			globalInspector.RenderOverlay(),  // 不需要 SetLayer
		),
		NewModalLayer(
			someModal,
		),
	)
}

// =============================================================================
// 问题 2: 渲染引擎多层渲染机制
// =============================================================================

// PaintEngine.PaintLayers() 的详细工作流程

/*
多 Layer 渲染流程：

1. 收集和布局阶段 (LayerManager.CollectAndLayout)

   输入: VNode 树
         └─ Fragment(appContent, inspectorOverlay)

   步骤:
   ├─ Collector.Collect(vnode)
   │   ├─ 遍历 VNode 树
   │   ├─ 发现 appContent (LayerBase) → 记录到 layers[LayerBase]
   │   └─ 发现 inspectorOverlay (LayerInspector) → 记录到 layers[LayerInspector]
   │
   ├─ StripLayers(vnode)
   │   └─ 返回 appContent (移除 inspectorOverlay)
   │
   ├─ Layout(baseTree, constraints)
   │   └─ 计算 appContent 的布局
   │       └─ baseLayout.Root.Box = (0, 0, 120, 40)
   │
   └─ 对每个 layer 节点:
       ├─ Layout(inspectorOverlay, constraints)
       │   └─ 计算 inspectorOverlay 的布局
       │       └─ 初始位置 (0, 0, 80, 25)
       │
       └─ positionInspector(node, layout)
           ├─ 读取 Props: {"x": 40, "y": 5}
           ├─ 计算偏移: offsetX = 40 - 0 = 40
           │             offsetY = 5 - 0 = 5
           └─ 应用偏移: 最终位置 (40, 5, 80, 25)

   输出: layouts map[Layer]ComputedLayout
         ├─ layers[LayerBase] = baseLayout
         └─ layers[LayerInspector] = inspectorLayout

2. 绘制阶段 (PaintEngine.PaintLayers)

   输入: layouts map

   步骤:
   ├─ 清空 buffer (或复用之前的 buffer)
   │
   └─ 按 z-order 顺序绘制:
       │
       ├─ Layer 0: Base
       │   └─ Paint(baseLayout, buffer)
       │       └─ 从 (0, 0) 开始绘制 appContent
       │           └─ buffer[0...39][0...119] 被填充
       │
       ├─ Layer 1: Overlay (如果有)
       │   └─ 绘制到指定位置
       │
       ├─ Layer 2: Modal (如果有)
       │   ├─ 绘制 modal 内容
       │   └─ 绘制半透明背景遮罩
       │
       ├─ Layer 3: Tooltip (如果有)
       │   └─ 绘制到鼠标位置附近
       │
       └─ Layer 4: Inspector
           └─ Paint(inspectorLayout, buffer)
               └─ 从 (40, 5) 开始绘制
                   └─ buffer[5...29][40...119] 被填充

   输出: buffer (包含所有 layer 的内容)

3. 输出阶段

   buffer → Terminal 显示
*/

// =============================================================================
// 问题 3: 架构优化 - 多个节点同时渲染
// =============================================================================

// ✅ 当前已经支持！使用 Fragment

func CurrentSolution_MultipleNodes() ui.VNode {
	// Fragment 允许多个子节点，不创建额外的布局容器
	return ui.Fragment(
		appContent,              // Layer 0
		inspectorOverlay,        // Layer 4
		modalOverlay,            // Layer 2
		tooltip,                 // Layer 3
	)
}

// Fragment 的工作原理:

/*
Fragment 是一个虚拟容器，具有以下特性：

1. 不创建实际的布局节点
   - 没有对应的 LayoutNode
   - 不占用布局空间
   - 不参与布局计算

2. 仅仅作为子节点的载体
   - Children() 返回所有子节点
   - 遍历时直接访问子节点

3. 在 StripLayers 中的处理
   - Fragment 不会阻止 StripLayers 递归
   - StripLayers 会遍历 Fragment 的所有子节点
   - 正确提取和分离不同 layer 的节点

示例：

VNode 树:
Fragment
  ├─ appContent (LayerBase)
  ├─ inspectorOverlay (LayerInspector)
  └─ modalOverlay (LayerModal)

StripLayers 处理:
├─ walk(Fragment)
│   ├─ walk(appContent)
│   │   └─ 递归所有子节点 (都是 LayerBase)
│   ├─ walk(inspectorOverlay)
│   │   └─ 发现 LayerInspector → 提取到 layers[LayerInspector]
│   └─ walk(modalOverlay)
│       └─ 发现 LayerModal → 提取到 layers[LayerModal]
│
└─ baseTree = Fragment(appContent) (modal 和 inspector 被移除)

结果:
- baseLayout: appContent 在 (0, 0)
- inspectorLayout: inspectorOverlay 在 (40, 5)
- modalLayout: modalOverlay 在居中位置
*/

// =============================================================================
// 问题 4: Inspector 位置和覆盖显示
// =============================================================================

// Inspector 位置配置

const (
	// Inspector 默认配置
	InspectorWidth  = 80
	InspectorHeight = 25
	InspectorMargin = 5
)

// 计算 Inspector 位置的函数

func CalculateInspectorPosition(screenWidth, screenHeight int) (x, y int) {
	// 策略 1: 右上角 (当前实现)
	x = screenWidth - InspectorWidth
	if x < 0 {
		x = 0  // 防止负数
	}
	y = InspectorMargin

	// 策略 2: 居中右侧
	// x = (screenWidth + InspectorWidth) / 2 - InspectorWidth
	// if x < 0 { x = 0 }
	// y = (screenHeight - InspectorHeight) / 2

	// 策略 3: 左上角
	// x = InspectorMargin
	// y = InspectorMargin

	return x, y
}

// 使用示例

func InspectorWithCalculatedPosition() ui.VNode {
	// 获取屏幕尺寸
	screenWidth, screenHeight := 120, 40

	// 计算位置
	x, y := CalculateInspectorPosition(screenWidth, screenHeight)

	// 创建 Inspector overlay
	inspectorOverlay := globalInspector.RenderOverlay()

	// 设置位置
	inspectorOverlay.SetProps(ui.Props{
		"x": x,  // 计算得到的位置，而不是硬编码
		"y": y,
	})

	return inspectorOverlay
}

// =============================================================================
// 覆盖显示原理
// =============================================================================

/*
覆盖显示的工作原理：

1. 绝对定位系统

   每个 layer 有自己的坐标原点和布局：

   LayerBase (appContent):
     └─ 布局约束: (0, 0, 120, 40)
         └─ 布局结果: Box(0, 0, 120, 40)
             └─ 占据整个屏幕

   LayerInspector (inspectorOverlay):
     └─ 布局约束: (0, 0, 120, 40)  ← 注意：初始位置也是 (0, 0)
         └─ 布局结果: Box(0, 0, 80, 25)  ← Inspector 的自然大小
             └─ positionInspector 调整: Box(40, 5, 80, 25)
                 └─ 最终位置: 从 (40, 5) 开始

2. 位置调整过程

   初始布局 (LayoutEngine):
   inspectorLayout.Root.Box = (0, 0, 80, 25)

   读取位置属性 (positionInspector):
   props["x"] = 40
   props["y"] = 5

   计算偏移:
   offsetX = 40 - 0 = 40
   offsetY = 5 - 0 = 5

   应用偏移 (shiftPositions):
   将 inspectorLayout 的所有节点向右移动 40，向下移动 5
   inspectorLayout.Root.Box = (40, 5, 80, 25)

3. 绘制过程 (PaintEngine.PaintLayers)

   Buffer 初始状态: 空 (120x40)

   步骤 1: 绘制 LayerBase
   ┌───────────────────────────────────────────┐
   │ Paint(baseLayout, buffer)                 │
   │   └─ 从 (0, 0) 开始绘制                   │
   │       └─ buffer[0..39][0..119] 被填充    │
   │                                           │
   │ Buffer 状态:                              │
   │ ┌─────────────────────────────────────┐  │
   │ │ App Content (Runtime Pipeline...)  │  │
   │ │ Statistics: Events: 0, Renders: 0  │  │
   │ │ Control Panel: [1] [2] [3] ...    │  │
   │ └─────────────────────────────────────┘  │
   └───────────────────────────────────────────┘

   步骤 2: 绘制 LayerInspector
   ┌───────────────────────────────────────────┐
   │ Paint(inspectorLayout, buffer)            │
   │   └─ 从 (40, 5) 开始绘制                  │
   │       └─ buffer[5..29][40..119] 被填充   │
   │                                           │
   │ Buffer 状态 (覆盖后):                     │
   │ ┌──────────────────┐ ┌────────────────┐  │
   │ │ App Content      │ │ INSPECTOR      │  │
   │ │ Runtime Pipeline │ │ Elements Tree  │  │
   │ │ Statistics       │ │ ┌──────────┐  │  │
   │ │ Control Panel    │ │ │ VStack   │  │  │
   │ │                  │ │ │ Bordered │  │  │
   │ │                  │ │ └──────────┘  │  │
   │ └──────────────────┘ └────────────────┘  │
   └───────────────────────────────────────────┘

   关键点:
   - buffer[0..4][*] 保持不变 (Inspector 上方)
   - buffer[5..29][0..39] 保持不变 (Inspector 左侧)
   - buffer[5..29][40..119] 被覆盖 (Inspector 区域)
   - buffer[30..39][*] 保持不变 (Inspector 下方)

4. 为什么不会互相干扰？

   因为每个 layer 绘制在不同的区域：

   LayerBase:    (0, 0) 到 (119, 39) - 但实际内容只在 (0, 0) 到 (39, 39)
   LayerInspector: (40, 5) 到 (119, 29)

   重叠区域: (40, 5) 到 (39, 29) → 空集，没有重叠！

   所以两个 layer 的内容不会互相覆盖。

5. 如果 Inspector 在 (0, 0) 会怎样？

   如果 Inspector 位置是 (0, 0):

   Inspector 会完全覆盖 appContent
   用户看不到应用界面

   这就是为什么 Inspector 位置是 (40, 5) 而不是 (0, 0)
*/

// =============================================================================
// 推荐的架构优化方案
// =============================================================================

// 方案 1: Layer 容器组件

type LayerContainer struct {
	layerType rtui.Layer
	children  []ui.VNode
}

func NewInspectorLayer(children ...ui.VNode) *LayerContainer {
	return &LayerContainer{
		layerType: rtui.LayerInspector,
		children:  children,
	}
}

func NewModalLayer(children ...ui.VNode) *LayerContainer {
	return &LayerContainer{
		layerType: rtui.LayerModal,
		children:  children,
	}
}

func (lc *LayerContainer) GetLayer() rtui.Layer {
	return lc.layerType
}

func (lc *LayerContainer) Children() []ui.VNode {
	return lc.children
}

// 使用示例
func RecommendedArchitecture() ui.VNode {
	return ui.VStack(
		appContent,
		NewInspectorLayer(
			globalInspector.RenderOverlay(),
		),
	)
}

// 方案 2: 位置枚举

type InspectorPosition int

const (
	PositionTopLeft InspectorPosition = iota
	PositionTopRight
	PositionBottomLeft
	PositionBottomRight
	PositionCenter
	PositionCustom
)

func (p InspectorPosition) Calculate(screenWidth, screenHeight, compWidth, compHeight int) (x, y int) {
	switch p {
	case PositionTopLeft:
		return 5, 5
	case PositionTopRight:
		return screenWidth - compWidth - 5, 5
	case PositionBottomLeft:
		return 5, screenHeight - compHeight - 5
	case PositionBottomRight:
		return screenWidth - compWidth - 5, screenHeight - compHeight - 5
	case PositionCenter:
		return (screenWidth - compWidth) / 2, (screenHeight - compHeight) / 2
	default:
		return 0, 0
	}
}

// 使用示例
func InspectorWithPositionEnum() ui.VNode {
	inspectorOverlay := globalInspector.RenderOverlay()

	// 使用枚举而不是魔法数字
	pos := PositionTopRight
	x, y := pos.Calculate(120, 40, 80, 25)

	inspectorOverlay.SetProps(ui.Props{
		"x": x,
		"y": y,
	})

	return inspectorOverlay
}

// =============================================================================
// 总结
// =============================================================================

/*
当前 Layer 系统的优势:
✅ 支持多个 layer 同时渲染
✅ 独立的布局和定位
✅ 正确的 z-order 渲染
✅ Inspector 和应用内容同时可见

需要改进的地方:
❌ 组件需要显式调用 SetLayer() (耦合)
❌ 硬编码的位置值 (40, 5)
❌ 缺少类型安全的 Layer 容器

推荐的改进:
1. 引入 LayerContainer 组件
2. 使用位置枚举代替魔法数字
3. 实现自动 Layer 推断 (可选)
4. 添加更多集成测试

当前可用的最佳实践:
✅ 使用 Fragment 包裹多个 layer 节点
✅ Inspector 位置通过 CalculateInspectorPosition() 计算
✅ Inspector 和应用内容通过绝对定位实现覆盖显示
*/
