# 渲染路径对比分析：为什么 TUI_INSPECTOR=false 能显示，true 不能显示

## 两种渲染路径

### 路径 A: TUI_INSPECTOR=false (工作正常 ✅)

```go
// main.go
RuntimeDemoWithInspectorOverlay()
  └─ return appContent  // 直接返回 appContent，没有 Fragment

// pipeline_renderer.go
PipelineRenderer.Render(vnode, ...)
  └─ hasLayerNodes() = false  // 没有 layer 节点
      └─ RenderingPipeline.Render(vnode, constraints, buffer)
          └─ LayoutEngine.Layout(appContent, constraints)
          └─ PaintEngine.Paint(layout, buffer)
              └─ paintNode(layout.Root, buffer)
                  └─ 绘制所有内容到 buffer
```

**结果**: ✅ 界面正常显示

---

### 路径 B: TUI_INSPECTOR=true (不显示 ❌)

```go
// main.go
RuntimeDemoWithInspectorOverlay()
  └─ return Fragment(appContent, inspectorOverlay)

// pipeline_renderer.go
PipelineRenderer.Render(vnode, ...)
  └─ hasLayerNodes() = true  // 有 LayerInspector 节点
      └─ RenderingPipeline.RenderLayers(vnode, constraints, buffer)
          └─ LayerManager.CollectAndLayout(vnode, constraints, engine)
              ├─ StripLayers() → baseTree
              ├─ Layout(baseTree) → baseLayout
              └─ Layout(inspectorOverlay) → inspectorLayout
          └─ PaintEngine.PaintLayers(layouts, buffer)
              ├─ Paint(baseLayout, buffer)
              └─ Paint(inspectorLayout, buffer)
```

**结果**: ❌ 界面不显示

---

## 关键差异

### 差异 1: VNode 树结构

**Without Inspector**:
```
appContent (VStack with 5 children)
```

**With Inspector**:
```
Fragment(appContent, inspectorOverlay)
  ↓ StripLayers
appContent (VStack with 5 children)
```

理论上应该相同，但是 Fragment 可能导致某些差异。

### 差异 2: 渲染方法

**Without Inspector**: `RenderingPipeline.Render()`
```go
func (p *RenderingPipeline) Render(...) error {
    layout, err := p.layoutEngine.Layout(vnode, constraints)
    return p.paintEngine.Paint(layout, buffer)
}
```

**With Inspector**: `RenderingPipeline.RenderLayers()`
```go
func (p *RenderingPipeline) RenderLayers(...) error {
    layerMgr.CollectAndLayout(vnode, constraints, p.layoutEngine)
    layouts := layerMgr.GetLayouts()
    return p.paintEngine.PaintLayers(layouts, buffer)
}
```

**关键**: `Paint()` vs `PaintLayers()`

---

## 问题定位

让我检查 `PaintEngine.PaintLayers()` 的实现，看看它是否真的将内容绘制到 buffer：

```go
// paint_engine.go:329-376
func (e *PaintEngine) PaintLayers(layouts LayerLayouts, buffer *paint.Buffer) error {
    renderOrder := []rtui.Layer{
        rtui.LayerBase,
        rtui.LayerOverlay,
        rtui.LayerModal,
        rtui.LayerTooltip,
        rtui.LayerInspector,
    }

    for _, layer := range renderOrder {
        layout, ok := layouts[layer]
        if !ok || layout.Root == nil {
            continue
        }

        // ❓ 这里调用了 Paint
        e.Paint(layout, buffer)

        // Modal 特殊处理
        if layer == rtui.LayerModal {
            e.paintModalBackdrop(layout.Root, buffer)
        }
    }

    return nil
}
```

**关键**: `PaintLayers` 内部调用 `Paint()`，所以理论上应该能正常绘制。

---

## 可能的问题

### 问题 1: Buffer 清空

检查是否有 buffer 清空逻辑：

```go
// PaintEngine.Paint() 开始时
func (e *PaintEngine) Paint(layout, buffer) {
    // ❓ 是否有清空 buffer 的逻辑？
    // buffer.Clear()  // ← 如果有这个，会导致问题
}
```

如果 `Paint()` 在开始时清空 buffer，那么：
1. Paint(baseLayout) 绘制到 buffer
2. Paint(inspectorLayout) 开始时清空 buffer ❌
3. Inspector 被绘制，但 base 被清空 ❌
4. 结果：只有 Inspector，没有 base

### 问题 2: 布局传递

检查 `Paint()` 的调用方式：

```go
// RenderLayers() 中
e.Paint(layout, buffer)  // ← 直接传递 layout

// Render() 中
e.Paint(layout, buffer)  // ← 直接传递 layout
```

两者应该相同，但可能有细微差别。

### 问题 3: Fragment 处理

检查 `Paint()` 对 Fragment 的处理：

```go
// paint_engine.go:161-163
case rtui.VNodeFragment:
    // Fragment - just paint children, no self-rendering
    return e.paintChildren(box, buffer)
```

如果 baseLayout.Root 是 Fragment，它会直接 paintChildren，可能导致某些问题。

---

## 验证方法

### 测试 1: 添加 buffer 内容检查

在 `PaintLayers()` 完成后检查 buffer：

```go
func (e *PaintEngine) PaintLayers(layouts LayerLayouts, buffer *paint.Buffer) error {
    // ... existing code ...

    // 检查 buffer 内容
    if os.Getenv("TUI_DEBUG_RENDER") == "true" {
        contentCount := 0
        for y := 0; y < buffer.Height; y++ {
            for x := 0; x < buffer.Width; x++ {
                if buffer.Cells[y][x].Cluster != "" {
                    contentCount++
                }
            }
        }
        fmt.Fprintf(os.Stderr, "[PaintLayers] Buffer content after rendering: %d cells\n", contentCount)
    }

    return nil
}
```

### 测试 2: 对比 Render() 和 RenderLayers()

创建一个测试，使用相同的 VNode，分别调用 Render() 和 RenderLayers()，比较结果。

### 测试 3: 禁用 Inspector layer

临时禁用 Inspector layer 的绘制：

```go
for _, layer := range renderOrder {
    if layer == rtui.LayerInspector {
        continue  // 跳过 Inspector
    }
    e.Paint(layout, buffer)
}
```

如果跳过 Inspector 后界面能显示，说明是 Inspector layer 的问题。

---

## 最可能的原因

根据调试输出：

```
[PaintEngine.Paint] START: layout.Root=*ui.LayoutNode, box=(0,0,80x19)
... (很多 "Paintable: NO")
[PaintEngine.Paint] END: err=<nil>

[PaintEngine.Paint] START: layout.Root=*ui.BorderedNode, box=(40,5,80x5)
... (很多 "Paintable: NO")
```

**推测**: `PaintLayers()` 完成后，buffer 的内容没有被正确输出到终端。

### 原因 1: Framework.App 不支持多层渲染

framework.App 可能期望的是单层渲染，不支持 PaintLayers() 的输出。

### 原因 2: Buffer 复制问题

RenderLayers() 可能创建了一个新的 buffer，而不是使用传入的 buffer。

### 原因 3: 终端刷新时机

框架可能在 PaintLayers() 完成前刷新了终端。

---

## 建议的排查步骤

### 步骤 1: 添加 buffer 内容日志

在 `PaintLayers()` 完成后立即检查 buffer 内容。

### 步骤 2: 检查 Framework.App 的终端输出

查看 `framework/app.go` 中的终端刷新逻辑。

### 步骤 3: 对比单层和多层渲染

创建最小化的测试，对比单层和多层渲染的 buffer 内容。

---

## 临时解决方案

如果问题无法快速定位，可以考虑：

### 方案 A: 禁用多层渲染

```go
// pipeline_renderer.go
func (r *PipelineRenderer) Render(vnode rtui.VNode, ...) error {
    hasLayers := r.hasLayerNodes(vnode)

    if hasLayers {
        // 临时禁用多层渲染，使用单层渲染
        // 这样 Inspector 会作为普通组件渲染
        return r.pipeline.Render(vnode, constraints, buffer)
    }

    return r.pipeline.Render(vnode, constraints, buffer)
}
```

### 方案 B: 将 Inspector 作为普通组件

```go
// main.go
if inspectorVisible {
    // 不使用 layer 系统，直接嵌入到 UI 中
    appWithInspector := ui.VStack(
        appContent,
        globalInspector.RenderOverlay(),  // 没有 SetLayer
    )
    return appWithInspector
}
```

这样虽然会失去独立 layer 的优势，但至少可以正常显示。

---

## 结论

**问题的核心**：TUI_INSPECTOR=false 使用单层渲染（工作正常），TUI_INSPECTOR=true 使用多层渲染（不显示）。

**最可能的原因**：buffer 内容没有被正确输出到终端，或者多层渲染的 buffer 处理有 bug。

**下一步**：添加 buffer 内容检查日志，确认 PaintLayers() 完成后 buffer 是否有内容。
