# 方案C实现分析：多VStack树支持

## 1. 当前系统架构

### 1.1 渲染流程

```
FrameworkApp.render()
  ↓
DeclarativeNode.Paint(ctx, buf)
  ↓
PipelineRenderer.Render(vnode, x, y, buf)
  ├─ hasLayerNodes(vnode) → 检查VNode树中的layer属性
  ├─ 如果有layer节点 → RenderLayers()
  │   ├─ LayerManager.CollectAndLayout(vnode, constraints, engine)
  │   │   ├─ Collect() - 遍历VNode树，收集带layer属性的节点
  │   │   ├─ StripLayers() - 从主树剥离layer节点 ✅
  │   │   └─ 返回: baseTree (不含layer节点)
  │   ├─ LayoutEngine.Layout(baseTree) → layout[LayerBase]
  │   ├─ LayoutEngine.Layout(layerNode) → layout[LayerInspector]
  │   │   └─ ... 为每个layer单独计算布局 ✅
  │   └─ GetLayouts() → map[Layer]ComputedLayout
  │   └─ PaintEngine.PaintLayers(layouts, buffer)
  │       ├─ LayerBase → 渲染到buffer
  │       ├─ LayerOverlay → 渲染到buffer
  │       ├─ LayerModal → 渲染到buffer
  │       └─ LayerInspector → 渲染到buffer (最上层) ✅
  └─ 如果没有layer → Render() - 标准单层渲染
```

### 1.2 关键发现

**PaintEngine.PaintLayers() (internal/render/paint_engine.go:329-376)**

```go
// 渲染顺序：从低到高z-index
renderOrder := []rtui.Layer{
    rtui.LayerBase,      // 主内容
    rtui.LayerOverlay,    // 普通overlay
    rtui.LayerModal,      // 模态对话框
    rtui.LayerTooltip,    // 提示
    rtui.LayerInspector,  // Inspector (最上层) ✅
}

for _, l := range renderOrder {
    layout, ok := layouts[l]  // 获取该layer的layout
    if !ok || layout.Root == nil {
        continue
    }

    // 每个layer有独立的布局和位置
    layout.Root.Box.X, layout.Root.Box.Y
    layout.Root.Box.Width, layout.Root.Box.Height

    // Paint到同一个buffer (合成)
    e.Paint(layout, buffer)
}
```

**关键点**：
- ✅ 每个layer有独立的ComputedLayout
- ✅ 每个layer有自己的位置和尺寸
- ✅ 所有layer被合成到同一个buffer
- ✅ LayerInspector是最高层

## 2. 当前问题诊断

### 2.1 Inspector Overlay不显示的原因

**症状**：从日志看到
```
[DEMO] ✅ Inspector overlay rendered: *ui.BorderedNode
[DEMO] Overlay has 1 children
```

但屏幕上看不到。

**原因**：虽然layer系统会剥离Inspector overlay，但它的布局信息不完整。

### 2.2 问题根源

当前的代码（main.go）：
```go
return ui.VStack(
    appContent,        // LayerBase
    inspectorOverlay,  // 有LayerInspector属性，但仍在VStack中
)
```

**问题**：
1. VStack不知道如何处理layer属性
2. VStack只是垂直排列子元素
3. LayerManager会剥离inspectorOverlay
4. **但剥离后的overlay没有明确的Rect信息**

从代码看（runtime/layer/manager.go:77-98）：
```go
// Layout each collected layer
for layer, nodes := range m.collector.GetLayers() {
    for _, node := range nodes {
        layerLayout, err := m.layoutLayer(node, layer, constraints, engine)
        // layoutLayer需要计算位置和尺寸
        m.layouts[layer] = layerLayout
    }
}
```

**layoutLayer需要**：
- 明确的width/height约束
- 明确的position (x, y)
- 但inspectorOverlay在VStack中时，这些信息依赖于VStack的布局

## 3. 方案C实现评估

### 3.1 当前系统是否支持多VStack树？

**答案：不支持，也不需要！**

**原因**：
1. App只有一个root字段（framework/app.go:38）
2. 渲染流程是：`root → Paint() → PipelineRenderer`
3. **LayerManager设计**：从单个tree中提取layer节点，不是管理多个独立tree

**LayerManager的工作方式**：
- 输入：单个包含layer节点的VNode树
- 处理：剥离layer节点，分别layout
- 输出：多个独立的ComputedLayout
- 合成：PaintEngine将所有layer合成到一个buffer

### 3.2 方案C的正确理解

**方案C ≠ 多个独立的VStack树**

**方案C的正确含义**：
- Inspector overlay完全独立于主VNode tree
- 不在主树的VStack中
- 通过layer系统直接管理
- 需要有独立的尺寸和位置信息

## 4. 实现方案

### 方案A：增强Inspector Overlay的布局信息（推荐）

**原理**：让inspectorOverlay在被VStack管理时，就记录完整的布局信息供layer系统使用。

**实现**：

#### 步骤1：给Inspector overlay添加明确的尺寸和位置

修改 `internal/inspector/standalone_inspector.go`：

```go
// RenderOverlay() 返回的overlay已有明确尺寸
func (si *StandaloneInspector) RenderOverlay() rtui.VNode {
    // ... 现有代码 ...

    content := si.buildOverlayContent()
    content.SetLayer(rtui.LayerInspector)

    // 关键：添加明确的布局信息
    content.SetProps(ui.Props{
        "absolute-position": ui.Position{
            X: 0,
            Y: 0,
        },
        "explicit-size": ui.Size{
            Width:  si.overlayWidth,
            Height: si.overlayHeight,
        },
        "layer": int(rtui.LayerInspector),
    })

    return content
}
```

#### 步骤2：确保LayerManager正确处理overlay

修改 `runtime/layer/manager.go`：

```go
// layoutNode中添加对显式size/position的处理
func (m *Manager) layoutLayer(
    node rtui.VNode,
    layer rtui.Layer,
    constraints runtime.BoxConstraints,
    engine *compute.Engine,
) (*compute.ComputedLayout, error) {

    // 检查是否有显式的尺寸信息
    if props := node.Props(); props != nil {
        if size, ok := props["explicit-size"]; ok {
            // 使用显式尺寸
            // ...
        }
        if pos, ok := props["absolute-position"]; ok {
            // 使用绝对定位
            // ...
        }
    }

    // 否则使用约束计算
    return engine.Layout(node, constraints)
}
```

### 方案B：在FrameworkApp中添加独立的overlay管理

**原理**：在App层面管理inspector overlay，不放在主VNode tree中。

**实现**：

#### 修改FrameworkApp

```go
// framework/app.go
type App struct {
    root              component.Node
    overlayTree       component.Node  // 新增：独立的overlay树
    hasOverlay        bool
    // ...
}

func (a *App) render() {
    buf := a.renderer.GetBackBuffer()
    buf.Reset(a.terminalWidth, a.terminalHeight)

    // 渲染主树
    if paintable, ok := a.root.(component.Paintable); ok {
        paintable.Paint(ctx, buf)
    }

    // 如果有overlay，单独渲染并合成
    if a.hasOverlay && a.overlayTree != nil {
        if paintable, ok := a.overlayTree.(component.Paintable); ok {
            // 使用绝对定位渲染overlay
            overlayCtx := component.PaintContext{
                AvailableWidth:  a.terminalWidth,
                AvailableHeight: a.terminalHeight,
                X: 0,
                Y: 0,
            }
            paintable.Paint(overlayCtx, buf)
        }
    }
}
```

#### 修改main.go

```go
func RuntimeDemoWithInspectorOverlay() ui.VNode {
    // 不再把inspector放在VStack中
    appContent := buildDemoContent(...)
    globalInspector.AttachToApp(appContent)

    // 通过FrameworkApp设置overlay
    if globalInspector.IsVisible() {
        inspectorOverlay := globalInspector.RenderOverlay()

        // 获取FrameworkApp引用（通过context或其他方式）
        // fwApp.SetOverlay(inspectorOverlay)
    }

    return appContent  // 只返回appContent，不含overlay
}
```

### 方案C：完全重构layer系统（不推荐）

**评估**：
- 复杂度极高
- 影响整个渲染架构
- 风险大
- 收益小

## 5. 推荐方案：方案A（增强Inspector布局信息）

### 5.1 实现步骤

#### 步骤1：在Inspector中添加显式尺寸和位置

```go
// internal/inspector/standalone_inspector.go
func (si *StandaloneInspector) RenderOverlay() rtui.VNode {
    si.mu.RLock()
    defer si.mu.RUnlock()

    if !si.visible {
        return nil
    }

    content := si.buildOverlayContent()

    // 设置layer属性
    content.SetLayer(rtui.LayerInspector)

    // 关键：添加明确的布局属性
    content.SetProps(ui.Props{
        "explicit-size": ui.Size{
            Width:  si.overlayWidth,
            Height: si.overlayHeight,
        },
        "absolute-position": ui.Position{
            X: 0,
            Y: 0,
        },
        "z-index": int(rtui.LayerInspector),
    })

    return content
}
```

#### 步骤2：确保LayoutNode处理显式尺寸

检查`runtime/compute/engine.go`或`runtime/layout/`是否有处理explicit-size的逻辑。

#### 步骤3：验证LayerManager

确保`runtime/layer/manager.go`中的`layoutLayer`能正确处理有显式尺寸的节点。

### 5.2 优势

1. ✅ **最小改动** - 只需修改Inspector相关代码
2. ✅ **兼容现有系统** - LayerManager和PaintEngine不需要大改
3. ✅ **符合设计意图** - Inspector是独立的overlay，有独立的布局
4. ✅ **易于测试** - 可以单独验证Inspector的布局

### 5.3 风险评估

- **低风险** - Inspector是独立组件，改动不影响主应用
- **测试成本** - 需要验证layer系统正确处理显式尺寸
- **兼容性** - 需要确保其他使用layer的系统不受影响

## 6. 验证测试

### 测试1：验证Inspector overlay有完整布局信息

```go
func TestInspectorOverlayLayout(t *testing.T) {
    inspector := NewStandaloneInspector()
    inspector.Enable()
    inspector.ToggleVisibility()

    overlay := inspector.RenderOverlay()

    // 检查props
    props := overlay.Props()
    assert.NotNil(t, props["explicit-size"], "Inspector should have explicit size")
    assert.NotNil(t, props["absolute-position"], "Inspector should have absolute position")
}
```

### 测试2：验证LayerManager正确处理

```go
func TestLayerManagerExplicitSize(t *testing.T) {
    // 创建带显式尺寸的节点
    node := rtui.Text("Test")
    node.SetProps(ui.Props{
        "explicit-size": ui.Size{Width: 80, Height: 25},
    })

    layerMgr := layer.NewManager()
    layerMgr.CollectAndLayout(node, constraints, engine)

    layouts := layerMgr.GetLayouts()
    layout, ok := layouts[rtui.LayerInspector]

    assert.True(t, ok, "LayerInspector should have layout")
    assert.NotNil(t, layout.Root.Box, "Layout should have valid Box")
}
```

## 7. 实施计划

### 阶段1：添加显式布局属性
- [ ] 修改`RenderOverlay()`添加explicit-size和absolute-position
- [ ] 检查ui.Size和ui.Position是否存在
- [ ] 如果不存在，创建这些类型

### 阶段2：验证Layer系统处理
- [ ] 检查LayoutEngine是否处理explicit-size
- [ ] 如需要，修改layout逻辑
- [ ] 测试layout结果

### 阶段3：集成测试
- [ ] 运行demo验证Inspector显示
- [ ] 测试F12切换
- [ ] 验证Inspector不会被VStack影响

### 阶段4：文档和清理
- [ ] 更新DEBUG_GUIDE.md
- [ ] 添加架构说明
- [ ] 清理临时测试文件

## 8. 备选方案：如果方案A不可行

如果LayerManager不能处理显式尺寸，则实施方案B（FrameworkApp管理overlay）：

### 修改FrameworkApp结构

```go
type App struct {
    root         component.Node
    overlayRoot  component.Node  // 新增：独立的overlay树
    // ...
}

func (a *App) SetOverlay(overlay component.Node) {
    a.overlayRoot = overlay
    a.hasOverlay = (overlay != nil)
    a.MarkDirty()
}

func (a *App) render() {
    // 渲染主树
    if paintable, ok := a.root.(component.Paintable); ok {
        ctx := PaintContext{
            AvailableWidth:  a.terminalWidth,
            AvailableHeight: a.terminalHeight,
            X: 0, Y: 0,
        }
        paintable.Paint(ctx, buf)
    }

    // 渲染overlay（如果存在）
    if a.hasOverlay && a.overlayRoot != nil {
        if paintable, ok := a.overlayRoot.(component.Paintable); ok {
            // Overlay覆盖在主内容之上
            paintable.Paint(ctx, buf)
        }
    }
}
```

### 使用方式

```go
// main.go
func RuntimeDemoWithInspectorOverlay() ui.VNode {
    appContent := buildDemoContent(...)
    globalInspector.AttachToApp(appContent)

    // 不再把inspector放在VStack中
    // 而是通过FrameworkApp设置
    if globalInspector.IsVisible() {
        fwApp.SetOverlay(globalInspector.RenderOverlay())
    }

    return appContent
}
```

## 9. 总结

### 9.1 当前系统状态

✅ **Layer系统已存在且工作正常**
- LayerManager可以剥离layer节点
- 每个layer有独立的ComputedLayout
- PaintEngine可以合成多个layers

❌ **Inspector overlay缺少布局信息**
- 虽然设置了LayerInspector属性
- 但在VStack中，缺少明确的size/position
- 导致layer剥离后，layout信息不完整

### 9.2 推荐方案

**方案A**：给Inspector overlay添加显式布局属性
- 最小改动
- 符合layer系统设计
- Inspector保持独立

**方案B**（备选）：FrameworkApp管理overlay
- 需要修改FrameworkApp核心
- 改动较大
- 更彻底的隔离

**方案C**（不推荐）：完全重构layer系统
- 风险太大
- 收益不明确

### 9.3 下一步

您希望我：
1. **实施方案A** - 添加显式布局属性（推荐）
2. **实施方案B** - FrameworkApp管理overlay
3. 先**分析Size/Position类型**是否已存在

请告诉我您的选择，或者我先检查Size/Position类型是否存在？
