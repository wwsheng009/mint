# Inspector 渲染失败深度分析

## 用户的问题

1. **启用 TUI_INSPECTOR=true 后，为什么原来的界面无法显示？**
2. **Inspector 的处理流程究竟是什么样的？**

## 完整的数据流分析

### 阶段 1: 用户设置环境变量

```go
// main.go:46-50
if os.Getenv("TUI_INSPECTOR") == "true" {
    globalInspector.ToggleVisibility()  // 设置 inspector.visible = true
    fmt.Println("UI Inspector auto-enabled")
}
```

**状态**: `globalInspector.visible = true`

### 阶段 2: 构建 VNode 树

```go
// main.go:125-157
func RuntimeDemoWithInspectorOverlay() ui.VNode {
    inspectorVisible := globalInspector.IsVisible()  // true

    if inspectorVisible {
        // 问题代码！
        inspectorOverlay := globalInspector.RenderOverlay()
        return ui.VStack(
            appContent,      // LayerBase
            inspectorOverlay, // LayerInspector
        )
    }
    return appContent
}
```

**关键问题**: 使用 `ui.VStack()` 包裹两个不同 layer 的节点！

### 阶段 3: VNode 树结构

#### Without Inspector (正常工作)

```
RuntimeDemo()
  └─ VStack (appContent)
      ├─ HeaderPanel
      ├─ PipelineVisualization
      ├─ StatisticsPanel
      ├─ ControlPanel
      └─ ExplanationPanel
```

**所有节点都是 LayerBase** → 使用标准渲染流程 ✅

#### With Inspector (问题)

```
RuntimeDemo()
  └─ VStack (root)  ← 新增的容器
      ├─ appContent (LayerBase)     ← VStack 的第一个子节点
      └─ inspectorOverlay (LayerInspector) ← VStack 的第二个子节点
```

**问题**: VStack 包含了不同 layer 的节点！

### 阶段 4: PipelineRenderer 检测 Layer

```go
// pipeline_renderer.go:105-124
func (r *PipelineRenderer) hasLayerNodes(vnode rtui.VNode) bool {
    // 检查当前节点
    layer := vnode.GetLayer()
    if layer != rtui.LayerBase && layer.IsValid() {
        return true  // ✅ 检测到 LayerInspector
    }

    // 递归检查子节点
    for _, child := range vnode.Children() {
        if r.hasLayerNodes(child) {
            return true
        }
    }
    return false
}
```

**检测结果**: `hasLayers = true` ✅

### 阶段 5: RenderLayers 流程

```go
// rendering_pipeline.go:145-186
func (p *RenderingPipeline) RenderLayers(...) {
    layerMgr := layer.NewManager()

    // 步骤 1: 收集所有 layer 节点
    layerMgr.CollectAndLayout(vnode, constraints, p.layoutEngine)

    // 步骤 2: 获取所有 layer 的布局
    layouts := layerMgr.GetLayouts()

    // 步骤 3: 绘制所有 layer
    p.paintEngine.PaintLayers(layouts, buffer)
}
```

### 阶段 6: Collector.Collect - 关键步骤！

```go
// runtime/layer/collector.go:127-159
func (c *Collector) Collect(vnode rtui.VNode) {
    c.walk(vnode)
}

func (c *Collector) walk(vnode rtui.VNode) {
    // 检查此节点是否有非 base layer
    if layer := vnode.GetLayer(); layer != rtui.LayerBase && layer.IsValid() {
        // 创建 layer 节点
        node := &LayerNode{
            Layer:   layer,
            Content: vnode,
            Visible: c.isVisible(vnode),
        }
        c.layers.Add(layer, node)

        // 关键：不遍历 layer 节点的子节点
        return  // ← 直接返回！
    }

    // 递归遍历子节点
    for _, child := range vnode.Children() {
        c.walk(child)
    }
}
```

**遍历过程**:

```
walk(VStack root)
  ├─ child 0: appContent (LayerBase) → 递归
  │   ├─ HeaderPanel (LayerBase) → 递归
  │   ├─ PipelineVisualization (LayerBase) → 递归
  │   └─ ... (所有 appContent 的子节点)
  │
  └─ child 1: inspectorOverlay (LayerInspector) → 提取到 layers[LayerInspector]
      └─ return (不遍历 inspectorOverlay 的子节点)
```

**结果**:
- `layers[rtui.LayerBase]` = appContent 的所有子节点（分散的）
- `layers[rtui.LayerInspector]` = inspectorOverlay

### 阶段 7: StripLayers - 问题所在！

```go
// runtime/layer/collector.go:214-228
func (c *Collector) StripLayers(vnode rtui.VNode) rtui.VNode {
    if vnode == nil {
        return nil
    }

    // 如果此节点本身就是 layer 节点，返回 nil
    if vnode.GetLayer() != rtui.LayerBase {
        return nil
    }

    // 克隆节点并过滤掉 layer 子节点
    cloned := c.cloneWithoutLayers(vnode)
    return cloned
}
```

**应用**:

```
StripLayers(VStack root)
  └─ VStack 本身是 LayerBase → 不是 layer 节点
      └─ 调用 cloneWithoutLayers(VStack)
```

### 阶段 8: cloneWithoutLayers - 真正的 bug！

```go
// runtime/layer/collector.go:231-289
func (c *Collector) cloneWithoutLayers(vnode rtui.VNode) rtui.VNode {
    // 获取非 layer 子节点
    var nonLayerChildren []rtui.VNode
    changed := false

    for _, child := range vnode.Children() {
        if child.GetLayer() != rtui.LayerBase {
            // 这是 layer 节点，跳过
            changed = true
            continue
        }
        // 保留普通节点
        filteredChild := c.cloneWithoutLayers(child)
        nonLayerChildren = append(nonLayerChildren, filteredChild)
    }

    // 如果没有子节点变化，返回原节点
    if !changed {
        return vnode
    }

    // 创建新的 VNode
    switch n := vnode.(type) {
    case *rtui.LayoutNode, *rtui.BorderedNode, *rtui.ComponentVNode:
        // 问题代码！
        vnode.SetChildren(nonLayerChildren)  // 直接修改原节点！
        return vnode
    // ...
    }
}
```

**实际执行**:

```
VStack root (LayoutNode)
  ├─ child 0: appContent (LayerBase) → 保留
  └─ child 1: inspectorOverlay (LayerInspector) → 跳过

结果:
nonLayerChildren = [appContent]
changed = true

进入 switch case *rtui.LayoutNode:
    vnode.SetChildren(nonLayerChildren)  // ← 修改了 VStack！
    return vnode  // 返回被修改的 VStack
```

**baseTree 结果**:

```
VStack root (被修改)
  └─ appContent  (唯一子节点)
```

### 阶段 9: LayoutEngine.Layout - 问题显现！

```
LayoutEngine.Layout(baseTree, constraints)
  └─ Layout(VStack with 1 child, constraints: 120x40)
      └─ VStack 布局计算:
          ├─ VStack 自己的 Box: (0, 0, 120, ?)
          └─ appContent 的 Box: (0, 0, 120, ?)
```

**问题**: VStack 尝试布局，但是：
1. VStack 只有一个子节点 `appContent`
2. `appContent` 本身也是一个 VStack
3. 嵌套的 VStack 导致布局高度计算错误

### 阶段 10: 为什么 Without Inspector 工作正常？

```
RuntimeDemo()
  └─ appContent (直接返回，没有外层 VStack)
      └─ VStack(...)
          ├─ HeaderPanel
          ├─ PipelineVisualization
          └─ ...
```

**LayoutEngine.Layout(appContent, constraints)**:
- 直接布局 appContent（VStack）
- VStack 有 5 个子节点
- 正确计算高度

## 根本原因

**问题**: 使用 `ui.VStack(appContent, inspectorOverlay)` 创建了一个**额外的容器**，导致：

1. ❌ **appContent 被嵌套在另一个 VStack 中**
2. ❌ **StripLayers 返回的是被修改的 VStack，而不是原始的 appContent**
3. ❌ **LayoutEngine 布局的是嵌套的 VStack，导致高度计算错误**

## 解决方案

### 方案 1: 不使用 VStack（推荐）

```go
func RuntimeDemoWithInspectorOverlay() ui.VNode {
    inspectorVisible := globalInspector.IsVisible()

    // 构建 appContent
    appContent := buildDemoContent(...)

    if inspectorVisible {
        // 使用 Fragment 而不是 VStack
        inspectorOverlay := globalInspector.RenderOverlay()
        return ui.Fragment(appContent, inspectorOverlay)
    }

    return appContent
}
```

**为什么有效**:
- Fragment 不会创建额外的布局节点
- StripLayers 会正确返回 appContent

### 方案 2: 直接返回多个节点（如果支持）

```go
func RuntimeDemoWithInspectorOverlay() ui.VNode {
    inspectorVisible := globalInspector.IsVisible()

    if inspectorVisible {
        // 创建一个临时根节点来持有两个子节点
        // 但这个根节点应该是透明的，不参与布局
        inspectorOverlay := globalInspector.RenderOverlay()

        // 使用特殊的 "layer-root" 容器
        return NewLayerRoot(
            appContent,
            inspectorOverlay,
        )
    }

    return appContent
}
```

### 方案 3: 修改 StripLayers 逻辑

```go
// 在 cloneWithoutLayers 中，检测到只有一个 LayerBase 子节点时
// 直接返回该子节点，而不是创建新的容器

if len(nonLayerChildren) == 1 && nonLayerChildren[0] == vnode.Children()[0] {
    // 只有一个子节点且它是 LayerBase，直接返回
    return nonLayerChildren[0]
}
```

## 推荐实施

**使用方案 1**：将 `ui.VStack()` 改为 `ui.Fragment()` 或直接返回多个节点。

这样可以确保：
1. ✅ StripLayers 返回原始的 appContent
2. ✅ LayoutEngine 布局的是 appContent，而不是嵌套的 VStack
3. ✅ Inspector 作为独立的 layer 被正确处理
