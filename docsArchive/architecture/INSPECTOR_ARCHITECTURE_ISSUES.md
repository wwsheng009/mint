# Inspector 架构三大核心问题分析

## 问题 1: Inspector 过度干涉渲染层

### 现状问题

```go
// Inspector 必须显式调用 SetLayer()
func (si *StandaloneInspector) RenderOverlay() rtui.VNode {
    content := si.buildOverlayContent()

    // ❌ Inspector 需要知道 Layer 系统
    content.SetProps(ui.Props{"x": si.floatX, "y": si.floatY})
    content.SetLayer(rtui.LayerInspector)  // 耦合了 Layer 概念

    return content
}
```

**问题点**:
- Inspector 组件不应该知道自己属于哪个 Layer
- 应用的布局应该决定组件的位置和层级
- 违反了单一职责原则

### 解决方案：使用 Layer 包装器

```go
// 在应用层处理 Layer，而不是 Inspector 内部
func RuntimeDemoWithInspectorOverlay() ui.VNode {
    inspectorVisible := globalInspector.IsVisible()

    if inspectorVisible {
        // Inspector 只返回纯内容
        inspectorContent := globalInspector.RenderContent()

        // 应用层决定如何使用 Inspector
        return ui.Fragment(
            appContent,
            ui.InspectorLayer(  // 新增：Layer 包装器
                ui.Position(40, 5,  // 新增：位置包装器
                    inspectorContent,
                ),
            ),
        )
    }

    return appContent
}
```

---

## 问题 2: 渲染引擎的多层处理机制

### 2.1 当前实现分析

```go
// paint_engine.go:329-376
func (e *PaintEngine) PaintLayers(layouts LayerLayouts, buffer *paint.Buffer) error {
    // 定义渲染顺序（z-order）
    renderOrder := []rtui.Layer{
        rtui.LayerBase,      // 0: 基础层
        rtui.LayerOverlay,   // 1: 覆盖层
        rtui.LayerModal,     // 2: 模态框层
        rtui.LayerTooltip,   // 3: 提示层
        rtui.LayerInspector, // 4: Inspector 层
    }

    // 按 z-order 从低到高渲染
    for _, layer := range renderOrder {
        layout, ok := layouts[layer]
        if !ok || layout.Root == nil {
            continue
        }

        // 每个 layer 直接绘制到同一个 buffer
        e.Paint(layout, buffer)
    }

    return nil
}
```

### 2.2 为什么 Inspector 位置是 (40, 5) 而不是 (0, 0)?

```
屏幕坐标系统:
(0,0) ──────────────────────────────────── (119,0)
  │                                            │
  │  主应用内容 (0-39)    Inspector (40-119)   │
  │  ←───────────→        ←────────────────→   │
  │                                            │
(0,39)───────────────────────────────────────(119,39)
```

**Inspector 位置 = (40, 5) 的原因**:
1. **X = 40**: Inspector 宽度 80，屏幕宽度 120，所以 X = 120 - 80 = 40
2. **Y = 5**: 顶部留 5 行空白，避免遮挡应用标题

**为什么不是 (0, 0)**:
- (0, 0) 位置应该留给主应用内容
- Inspector 是覆盖层，应该"浮动"在应用之上
- 如果 Inspector 在 (0, 0)，会完全遮挡主应用

### 2.3 多层渲染的覆盖机制

当前实现：**直接覆盖绘制**

```
Buffer 初始状态: 全空

Step 1: Paint(LayerBase)
    buffer[0..39][0..39] = 主应用内容

Step 2: Paint(LayerInspector) at (40, 5)
    buffer[40..119][5..29] = Inspector 内容
    ← 直接覆盖，没有透明度合成

最终 Buffer:
    [0..39][0..39]    = 主应用 ✅ 可见
    [40..119][5..29]  = Inspector ✅ 可见
    [40..119][0..4]   = 空 ⚠️ 未使用
```

**问题**:
- ❌ 没有透明度处理
- ❌ 没有边界裁剪
- ❌ Inspector 超出屏幕部分会被绘制（虽然有边界检查，但不完善）

---

## 问题 3: 架构优化 - Root 节点多节点渲染

### 3.1 当前架构的限制

```go
// 只能返回一个 VNode
func RuntimeDemoWithInspectorOverlay() ui.VNode {
    if inspectorVisible {
        // ❌ 必须用 Fragment 或 VStack 包装多个节点
        return ui.Fragment(
            appContent,
            inspectorOverlay,
        )
    }
    return appContent
}
```

### 3.2 理想架构

**目标**: 一个 root 节点下可以有多个子节点同时渲染

```go
// 理想方案 1: 直接返回多个节点
func RuntimeDemoWithInspectorOverlay() []ui.VNode {
    var nodes []ui.VNode
    nodes = append(nodes, appContent)

    if globalInspector.IsVisible() {
        nodes = append(nodes, globalInspector.RenderContent())
    }

    return nodes  // ← 返回节点数组
}

// 理想方案 2: 使用 Scene
func RuntimeDemoWithInspectorOverlay() *Scene {
    scene := NewScene()
    scene.Add(appContent)

    if globalInspector.IsVisible() {
        scene.Add(globalInspector.RenderContent())
    }

    return scene
}
```

### 3.3 渲染引擎如何处理多个 root 节点

```go
// 改进的渲染流程
func (r *PipelineRenderer) Render(vnode rtui.VNode, ...) error {
    // Step 1: 提取所有 root 节点
    roots := r.extractRoots(vnode)

    // Step 2: 检测是否有 layer 节点
    if r.hasLayerNodes(roots...) {
        return r.renderLayers(roots, ...)
    }

    // Step 3: 标准渲染
    return r.renderStandard(roots, ...)
}

func (r *PipelineRenderer) extractRoots(vnode rtui.VNode) []ui.VNode {
    // 处理 Fragment
    if fragment, ok := vnode.(*FragmentVNode); ok {
        return fragment.Children()
    }

    // 处理数组（新 API）
    if array, ok := vnode.([]ui.VNode); ok {
        return array
    }

    // 处理单个节点
    return []ui.VNode{vnode}
}

func (r *PipelineRenderer) hasLayerNodes(roots ...ui.VNode) bool {
    for _, root := range roots {
        if root.GetLayer() != rtui.LayerBase {
            return true
        }
    }
    return false
}

func (r *PipelineRenderer) renderLayers(roots []ui.VNode, ...) error {
    // Step 1: 收集所有 layer 节点
    collector := layer.NewCollector()
    for _, root := range roots {
        collector.Collect(root)
    }

    // Step 2: 分别布局每个 layer
    layouts := make(map[rtui.Layer]*compute.ComputedLayout)
    for layer, nodes := range collector.GetLayers() {
        for _, node := range nodes {
            layout, err := r.pipeline.GetLayoutEngine().Layout(
                node.Content,
                constraints,
            )
            if err != nil {
                return err
            }

            // 合并同一 layer 的多个节点
            if existing, ok := layouts[layer]; ok {
                layouts[layer] = r.mergeLayouts(existing, layout)
            } else {
                layouts[layer] = layout
            }
        }
    }

    // Step 3: 按顺序绘制所有 layer
    return r.paintEngine.PaintLayers(layouts, buffer)
}
```

---

## 推荐的架构优化方案

### 方案 A: 最小改动（立即可实施）

**保持当前架构，改进文档和示例**:

```go
// main.go
func RuntimeDemoWithInspectorOverlay() ui.VNode {
    inspectorVisible := globalInspector.IsVisible()

    if inspectorVisible {
        // 使用 Fragment 组合多个节点
        return ui.Fragment(
            appContent,                          // LayerBase (默认)
            globalInspector.RenderOverlay(),    // LayerInspector (Inspector 内部设置)
        )
    }

    return appContent
}
```

**优点**: 无需修改核心代码
**缺点**: Inspector 仍然需要知道 Layer 系统

---

### 方案 B: Layer 包装器（推荐）

**引入 Layer 装饰器，解耦组件和层系统**:

```go
// 新增: Layer 包装函数
func LayerBase(content ui.VNode) ui.VNode {
    content.SetLayer(rtui.LayerBase)
    return content
}

func LayerInspector(content ui.VNode) ui.VNode {
    content.SetLayer(rtui.LayerInspector)
    return content
}

func LayerModal(content ui.VNode) ui.VNode {
    content.SetLayer(rtui.LayerModal)
    return content
}

// 使用
func RuntimeDemoWithInspectorOverlay() ui.VNode {
    inspectorVisible := globalInspector.IsVisible()

    if inspectorVisible {
        // Inspector 只返回纯内容
        inspectorContent := globalInspector.RenderContent()

        // 应用层决定 Layer
        return ui.Fragment(
            LayerBase(appContent),
            LayerInspector(
                Position(40, 5, inspectorContent),  // 也在应用层决定位置
            ),
        )
    }

    return LayerBase(appContent)
}

// Inspector 实现简化
func (si *StandaloneInspector) RenderContent() ui.VNode {
    // 不再需要 SetLayer()
    return si.buildOverlayContent()
}
```

**优点**:
- ✅ 组件不需要知道 Layer 系统
- ✅ 应用层完全控制布局
- ✅ 更符合关注点分离原则

**缺点**:
- 需要新增 Layer 包装函数
- 需要修改 Inspector API

---

### 方案 C: Scene Graph（长期目标）

**引入 Scene 概念，支持动态节点管理**:

```go
// 新增: Scene 类型
type Scene struct {
    nodes []SceneNode
}

type SceneNode struct {
    VNode   ui.VNode
    Layer   rtui.Layer
    Position Position
    ZIndex  int
}

func (s *Scene) Add(vnode ui.VNode, layer rtui.Layer, pos Position) {
    s.nodes = append(s.nodes, SceneNode{
        VNode:   vnode,
        Layer:   layer,
        Position: pos,
    })
}

// 使用
func RuntimeDemoWithInspectorOverlay() *Scene {
    scene := &Scene{}

    // 添加主应用
    scene.Add(appContent, rtui.LayerBase, Position{X: 0, Y: 0})

    // 条件性添加 Inspector
    if globalInspector.IsVisible() {
        inspectorContent := globalInspector.RenderContent()
        scene.Add(inspectorContent, rtui.LayerInspector, Position{X: 40, Y: 5})
    }

    return scene
}

// 渲染引擎处理 Scene
func (r *PipelineRenderer) Render(vnode ui.VNode, ...) error {
    if scene, ok := vnode.(*Scene); ok {
        return r.renderScene(scene, ...)
    }
    return r.renderStandard(vnode, ...)
}

func (r *PipelineRenderer) renderScene(scene *Scene, ...) error {
    // 按 Z-order 排序
    sort.Slice(scene.nodes, func(i, j int) bool {
        return scene.nodes[i].Layer.ZIndex() < scene.nodes[j].Layer.ZIndex()
    })

    // 渲染每个节点
    for _, node := range scene.nodes {
        r.renderNode(node.VNode, node.Position, ...)
    }

    return nil
}
```

**优点**:
- ✅ 完全解耦组件和层系统
- ✅ 支持动态添加/移除节点
- ✅ 灵活的位置和层级控制
- ✅ 易于扩展（动画、过渡等）

**缺点**:
- 需要大量重构
- 需要引入新的 Scene 概念
- 向后兼容性问题

---

## 总结与建议

### 三个问题的核心症结

| 问题 | 根本原因 | 影响 |
|------|---------|------|
| **1. Inspector 过度干涉** | 组件内部直接调用 SetLayer() | 耦合度高，违反单一职责 |
| **2. 多层渲染不清晰** | 缺少明确的合成机制 | 位置计算、覆盖逻辑不透明 |
| **3. Root 节点限制** | 渲染引擎假设单一 root | 需要用 Fragment workaround |

### 实施建议

**短期（本周）**:
1. ✅ 保持当前架构
2. ✅ 改进文档，说明 Inspector 和 Layer 的关系
3. ✅ 添加示例代码展示最佳实践

**中期（本月）**:
1. 🔄 引入 Layer 包装器（方案 B）
2. 🔄 修改 Inspector API，移除内部 SetLayer() 调用
3. 🔄 改进位置计算，支持屏幕自适应

**长期（下季度）**:
1. 🎯 设计并实现 Scene Graph（方案 C）
2. 🎯 支持动态节点管理
3. 🎯 添加动画和过渡效果

### 当前最佳实践

```go
// 推荐的使用方式
func RuntimeDemoWithInspectorOverlay() ui.VNode {
    inspectorVisible := globalInspector.IsVisible()

    if inspectorVisible {
        // 使用 Fragment 避免额外布局容器
        return ui.Fragment(
            appContent,                       // 主应用
            globalInspector.RenderOverlay(),  // Inspector overlay
        )
    }

    return appContent
}
```

**关键点**:
1. 使用 `ui.Fragment()` 而不是 `ui.VStack()` 避免嵌套布局
2. Inspector 内部调用 `SetLayer()` 是当前架构的必要实现
3. 渲染引擎会自动处理多层合成
