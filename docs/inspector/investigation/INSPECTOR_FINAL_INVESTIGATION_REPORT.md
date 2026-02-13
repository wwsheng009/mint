# Inspector 渲染问题 - 最终调查结果

## 确认的现象

### 测试 1: TUI_INSPECTOR=false
```bash
./demo_fixed2.exe
```
**结果**: ✅ 界面正常显示，所有组件可见

### 测试 2: TUI_INSPECTOR=true
```bash
TUI_INSPECTOR=true ./demo_fixed2.exe
```
**结果**: ❌ 只显示初始化消息，界面内容不显示

---

## 已排除的问题

### 1. ❌ Fragment 嵌套布局问题
- **假设**: Fragment 创建了额外的布局容器
- **验证**: ✅ Fragment 正确解包
- **测试**: `TestStripLayersFragmentUnwrap` 通过
- **结论**: Fragment 不是问题

### 2. ❌ StripLayers 逻辑问题
- **假设**: StripLayers 没有正确移除 Inspector
- **验证**: ✅ StripLayers 正确解包 Fragment，返回 appContent
- **测试**: `TestStripLayersDebug` 通过
- **结论**: StripLayers 不是问题

### 3. ❌ SetLayer 不生效
- **假设**: SetLayer() 没有正确设置 layer 属性
- **验证**: ✅ SetLayer 工作正常
- **测试**: `TestBorderedNodeSetLayer` 通过
- **结论**: SetLayer 不是问题

### 4. ❌ Collector 不识别 Inspector
- **假设**: Collector.Collect() 没有收集到 Inspector
- **验证**: ✅ Collector 正确识别 LayerInspector
- **测试**: `TestCollectorWithInspector` 通过
- **结论**: Collector 不是问题

### 5. ❌ 布局计算错误
- **假设**: baseLayout 或 inspectorLayout 的位置/大小不正确
- **验证**: ✅ 两个 layout 都正确计算
- **调试输出**:
  ```
  baseLayout.Root.Box: (0, 0, 80x19)  ✅
  inspectorLayout.Root.Box: (40, 5, 80x5)  ✅
  ```
- **结论**: 布局计算不是问题

---

## 真正的问题

### 对比两种渲染路径

#### 路径 A: TUI_INSPECTOR=false
```
RuntimeDemoWithInspectorOverlay()
  └─ return appContent  (VStack)
      └─ PipelineRenderer.Render()
          └─ RenderingPipeline.Render()
              └─ PaintEngine.Paint(layout, buffer)
                  └─ buffer 内容输出到终端 ✅
```

#### 路径 B: TUI_INSPECTOR=true
```
RuntimeDemoWithInspectorOverlay()
  └─ return Fragment(appContent, inspectorOverlay)
      └─ PipelineRenderer.Render()
          └─ hasLayerNodes() = true
          └─ RenderingPipeline.RenderLayers()
              └─ LayerManager.CollectAndLayout()
              └─ PaintEngine.PaintLayers(layouts, buffer)
                  ├─ Paint(baseLayout, buffer)
                  └─ Paint(inspectorLayout, buffer)
                  └─ buffer 内容没有输出到终端 ❌
```

### 关键差异

| 方面 | Render() | RenderLayers() |
|------|----------|---------------|
| **VNode** | 单个 appContent | Fragment(appContent, inspectorOverlay) |
| **Layout** | 单层 layout | 多层 layout (base + inspector) |
| **Paint** | 调用一次 Paint() | 调用两次 Paint() |
| **结果** | ✅ 显示 | ❌ 不显示 |

### 可能的原因

#### 原因 1: Buffer 复制/传递问题

**PaintLayers()** 可能使用了错误的 buffer，或者创建了新的 buffer。

**验证**: 检查 `PaintEngine.PaintLayers()` 的 buffer 参数

#### 原因 2: Framework.App 不支持多层渲染

**framework.App** 可能不支持多层渲染的结果，需要特殊处理。

**验证**: 检查 `framework/app.go` 中的终端输出逻辑

#### 原因 3: Buffer 清空逻辑

**Paint()** 可能在开始时清空 buffer，导致第二次调用时第一次绘制的内容丢失。

**验证**: 检查 `PaintEngine.Paint()` 的开头是否有清空逻辑

---

## 需要进一步调查的点

### 1. 检查 PaintLayers 后的 buffer 内容

在 `RenderingPipeline.RenderLayers()` 末尾添加：

```go
func (p *RenderingPipeline) RenderLayers(...) error {
    // ... existing code ...

    // Paint all layers
    if err := p.paintEngine.PaintLayers(layouts, buffer); err != nil {
        return err
    }

    // 验证 buffer 内容
    if os.Getenv("TUI_DEBUG_RENDER") == "true" {
        contentCount := 0
        for y := 0; y < buffer.Height; y++ {
            for x := 0; x < buffer.Width; x++ {
                if buffer.Cells[y][x].Cluster != "" {
                    contentCount++
                }
            }
        }
        fmt.Fprintf(os.Stderr, "[RenderLayers] Buffer content: %d cells\n", contentCount)
    }

    return nil
}
```

### 2. 检查 Framework.App 的终端输出

查看 `framework/app.go` 中的终端刷新逻辑，确认是否需要特殊处理多层渲染。

### 3. 检查 Paint() 是否清空 buffer

查看 `PaintEngine.Paint()` 的实现，确认是否在开始时清空 buffer。

---

## 临时解决方案

### 方案 A: 禁用多层渲染（快速修复）

```go
// pipeline_renderer.go
func (r *PipelineRenderer) Render(vnode rtui.VNode, ...) error {
    hasLayers := r.hasLayerNodes(vnode)

    if hasLayers {
        // 临时禁用多层渲染
        // 直接使用 Render()，让 Inspector 作为普通组件渲染
        return r.pipeline.Render(vnode, constraints, buffer)
    }

    return r.pipeline.Render(vnode, constraints, buffer)
}
```

**优点**: 可以快速让界面显示
**缺点**: Inspector 会作为普通组件嵌入到 UI 中，不是独立的 overlay

### 方案 B: 将 Inspector 作为普通组件

```go
// main.go
if inspectorVisible {
    // 不使用 Fragment 和 layer 系统
    // 直接嵌入到 VStack 中
    appWithInspector := ui.VStack(
        appContent,
        globalInspector.RenderContent(),  // 注意：不调用 RenderOverlay()
    )
    return appWithInspector
}
```

**优点**: 简单直接
**缺点**: Inspector 不是独立的 overlay，会占据布局空间

---

## 总结

### 确认的事实

1. ✅ 所有渲染逻辑工作正常（layout、paint、collect）
2. ✅ Layer 系统正确识别和分离 base 和 inspector
3. ✅ 单层渲染（TUI_INSPECTOR=false）工作正常
4. ❌ 多层渲染（TUI_INSPECTOR=true）不显示

### 问题定位

**问题不在渲染逻辑本身，而在 buffer 输出到终端的环节。**

RenderLayers() 正确地将内容绘制到 buffer，但是 buffer 的内容没有被正确输出到终端。

### 建议的下一步

1. 添加 buffer 内容检查日志，确认 PaintLayers() 完成后 buffer 是否有内容
2. 检查 framework/App 的终端输出逻辑
3. 如果无法快速定位，使用临时方案禁用多层渲染
