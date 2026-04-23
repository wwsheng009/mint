# Inspector Overlay 显示问题 - 最终状态报告

## 当前状态

### ✅ 已修复的问题

1. **StripLayers 逻辑错误** - 修复了对 LayerBase 子节点递归的错误
2. **节点类型转换错误** - 修复了将 LayoutNode/BorderedNode 转换为 ElementVNode 导致属性丢失的问题
3. **布局高度错误** - 修复后高度从 1073741823 恢复到正常的 19

### ✅ 测试结果

**所有测试通过**:
```bash
cd runtime/layer
go test -v -run TestStripLayers
```

```
✅ StripLayers correctly preserved base content
✅ baseTree has 1 children
✅ appContent has 3 children
✅ StripLayers correctly removed multiple layers
✅ StripLayers handles edge cases correctly
```

### ⚠️ 遗留问题

**启用 Inspector 时界面不显示**

#### 不使用 Inspector (正常工作)
```bash
TUI_INSPECTOR=false ./demo_fixed.exe
```
**结果**: ✅ 界面正常显示
- Runtime Scheduling Pipeline Visualization 标题
- Event 流程可视化
- Statistics 面板
- Control 按钮面板
- System idle 提示

#### 使用 Inspector (不显示)
```bash
TUI_INSPECTOR=true ./demo_fixed.exe
```
**结果**: ❌ 只显示初始化消息，界面不渲染
- 输出: "UI Inspector auto-enabled"
- 输出: "Starting Mint TUI Demo"
- 然后没有界面内容

### 调试信息

启用 `TUI_LAYER_DEBUG=true`:
```
[PipelineRenderer] hasLayers=true
[CollectAndLayout] baseTree has 1 children (after stripping)
[CollectAndLayout]   child 0: layer=0 type=Element
[positionInspector] original=(0,0) target=(80,5)
[positionInspector] after shift: inspector=(80,5) size=80x5
```

✅ Layer 系统工作正常
✅ Inspector 被正确剥离
✅ Inspector 被正确定位

但是 `PaintLayers` 没有被调用或没有输出到终端。

## 根本原因

**DeclarativeNode.Paint()** 中，当有 layer 时使用 `PipelineRenderer.RenderLayers()`，但是可能：
1. `PaintLayers()` 调用了，但 buffer 没有输出到终端
2. 或者 `RenderLayers()` 返回了错误，fallback 到 `Render()` 但也失败了
3. 或者 buffer 内容为空

### 建议的调试步骤

#### 步骤 1: 检查 RenderLayers 是否被调用

在 `internal/render/rendering_pipeline.go:145` 的 `RenderLayers()` 中添加日志：

```go
func (p *RenderingPipeline) RenderLayers(...) error {
    fmt.Fprintf(os.Stderr, "[RenderLayers] Called\n")

    // ... existing code ...

    // Paint all layers
    if err := p.paintEngine.PaintLayers(layouts, buffer); err != nil {
        fmt.Fprintf(os.Stderr, "[RenderLayers] PaintLayers ERROR: %v\n", err)
        return err
    }

    fmt.Fprintf(os.Stderr, "[RenderLayers] PaintLayers SUCCESS\n")
    return nil
}
```

#### 步骤 2: 检查 layouts 的内容

```go
// 在 RenderLayers() 中
layouts := layerMgr.GetLayouts()
fmt.Fprintf(os.Stderr, "[RenderLayers] Total layers: %d\n", len(layouts))

for layer, layout := range layouts {
    if layout.Root != nil {
        fmt.Fprintf(os.Stderr, "[RenderLayers] Layer %d: root=(%d,%d) size=%dx%d\n",
            layer, layout.Root.Box.X, layout.Root.Box.Y,
            layout.Root.Box.Width, layout.Root.Box.Height)
    }
}
```

#### 步骤 3: 检查 buffer 内容

```go
// 在 PaintLayers() 之后
fmt.Fprintf(os.Stderr, "[RenderLayers] Buffer size: %dx%d\n",
    buffer.Width, buffer.Height)

// 检查 buffer 是否为空
hasContent := false
for y := 0; y < buffer.Height; y++ {
    for x := 0; x < buffer.Width; x++ {
        cell := buffer.Cells[y][x]
        if cell.Rune != 0 || cell.Style != "" {
            hasContent = true
            break
        }
    }
    if hasContent {
        break
    }
}
fmt.Fprintf(os.Stderr, "[RenderLayers] Buffer has content: %v\n", hasContent)
```

## 临时解决方案

### 方案 A: 不使用 Layer 系统（已验证工作）

暂时不使用 Inspector overlay 的 layer 功能：

**main.go**:
```go
func RuntimeDemoWithInspectorOverlay() ui.VNode {
    currentPhase, setCurrentPhase := ui.UseStateString("idle")
    // ... state hooks ...

    appContent := buildDemoContent(...)

    // 附加 inspector 用于分析，但不显示 overlay
    globalInspector.AttachToApp(appContent)

    // 不返回 overlay，直接返回 appContent
    return appContent  // 暂时移除 overlay
}
```

这样 Inspector 还可以分析 VNode tree，但不会作为 overlay 显示。

### 方案 B: 使用内联 Inspector（不使用 layer）

将 Inspector 直接嵌入到主内容中，而不是作为 overlay：

```go
func RuntimeDemoWithInspectorOverlay() ui.VNode {
    appContent := buildDemoContent(...)

    inspectorVisible := globalInspector.IsVisible()

    if inspectorVisible {
        inspectorContent := globalInspector.RenderOverlay()
        // 将 Inspector 作为主内容的一部分，而不是 overlay
        return ui.VStack(
            appContent,
            ui.Text(""),  // 分隔线
            inspectorContent,  // 直接嵌入，不使用 layer
        )
    }

    return appContent
}
```

### 方案 C: 使用独立的 Inspector 窗口

创建一个分割窗口，左边是主应用，右边是 Inspector：

```go
// 使用 HStack 创建分割视图
return ui.HStack(
    appContent,
    ui.VStack(
        ui.Text("Inspector"),
        globalInspector.RenderOverlay(),
    ).Width(40),  // 固定宽度
)
```

## 推荐方案

**建议先使用方案 A**（不显示 Inspector overlay），这样：
1. ✅ 主应用可以正常显示和使用
2. ✅ Inspector 的分析功能仍然可用（通过快捷键查看）
3. ✅ 不影响正在进行的其他开发工作
4. ✅ 可以有时间继续调试 layer 系统的问题

然后作为独立任务调查 Layer 渲染问题。
