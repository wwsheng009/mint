# UI Inspector - 方案3: 框架级覆盖层实现计划

**Plan 3: Framework-Level Overlay Support for UI Inspector**

---

## 📋 方案概述

**核心思想**：扩展现有的 Layer 系统，添加 Inspector 作为最高级别的覆盖层，与应用同时渲染。

**用户体验**：
- 按 `F12` 切换：显示/隐藏 Inspector 覆盖层
- 应用界面：始终保持完整显示和交互
- Inspector 界面：作为覆盖层叠加在应用上方
- 支持同时查看应用和 Inspector

---

## 🎯 设计目标

### 必须满足的需求

1. ✅ **真正的覆盖层**：Inspector 不影响应用布局
2. ✅ **同时显示**：应用和 Inspector 可见
3. ✅ **独立交互**：两者都可正常交互
4. ✅ **框架集成**：利用现有 Layer 系统

### 架构原则

1. **最小侵入**：尽量复用现有架构
2. **向后兼容**：不影响现有功能
3. **可扩展性**：支持未来的覆盖层需求
4. **高性能**：高效的 buffer 合成

---

## 🏗️ 架构设计

### 现有架构分析

根据调研结果，Mint TUI 框架已有完善的 Layer 系统：

```
当前 Layer 层级：
┌─────────────────────────────────────┐
│ LayerTooltip (z-index: 3)           │  最高
├─────────────────────────────────────┤
│ LayerModal (z-index: 2)             │
├─────────────────────────────────────┤
│ LayerOverlay (z-index: 1)           │  下拉菜单等
├─────────────────────────────────────┤
│ LayerBase (z-index: 0)              │  基础内容
└─────────────────────────────────────┘
```

### 扩展后的架构

```
扩展 Layer 层级：
┌─────────────────────────────────────┐
│ LayerInspector (z-index: 4)         │  新增 - Inspector
├─────────────────────────────────────┤
│ LayerTooltip (z-index: 3)           │
├─────────────────────────────────────┤
│ LayerModal (z-index: 2)             │
├─────────────────────────────────────┤
│ LayerOverlay (z-index: 1)           │
├─────────────────────────────────────┤
│ LayerBase (z-index: 0)              │
└─────────────────────────────────────┘
```

### 渲染管道

```
VNode Tree (应用 + Inspector)
    ↓
LayerManager.CollectAndLayout()
    ↓
    ├─ Base Layer → ComputedLayout
    ├─ Overlay Layer → ComputedLayout
    ├─ Modal Layer → ComputedLayout
    ├─ Tooltip Layer → ComputedLayout
    └─ Inspector Layer → ComputedLayout  ← 新增
    ↓
PaintEngine.PaintLayers()
    ↓
    ├─ Base Buffer
    ├─ Overlay Buffer
    ├─ Modal Buffer
    ├─ Tooltip Buffer
    └─ Inspector Buffer                  ← 新增
    ↓
Compositor.Compose()
    ↓
Final Buffer (按 z-index 合成)
    ↓
Renderer.Render()
    ↓
Terminal Output
```

---

## 🔧 实现细节

### 1. 扩展 Layer 系统

**文件**: `runtime/ui/vnode.go`

```go
const (
    LayerBase      VNodeLayer = 0  // Normal UI content
    LayerOverlay   VNodeLayer = 1  // Dropdowns, popovers
    LayerModal     VNodeLayer = 2  // Modal dialogs
    LayerTooltip   VNodeLayer = 3  // Tooltips
    LayerInspector VNodeLayer = 4  // Inspector overlay (NEW)
)
```

**文件**: `runtime/layer/manager.go`

```go
// LayerManager 管理所有层
type LayerManager struct {
    // ... 现有字段

    // Inspector 支持
    inspectorEnabled bool
    inspectorVNode   ui.VNode
}

// EnableInspector 启用 Inspector 覆盖层
func (lm *LayerManager) EnableInspector(inspector ui.VNode) {
    lm.inspectorEnabled = true
    lm.inspectorVNode = inspector
}

// DisableInspector 禁用 Inspector 覆盖层
func (lm *LayerManager) DisableInspector() {
    lm.inspectorEnabled = false
    lm.inspectorVNode = nil
}

// CollectAndLayout 收集并布局所有层
func (lm *LayerManager) CollectAndLayout(
    root ui.VNode,
    engine *compute.Engine,
    constraints BoxConstraints,
) *LayoutResult {
    // ... 现有逻辑

    // 新增：收集 Inspector 层
    if lm.inspectorEnabled && lm.inspectorVNode != nil {
        inspectorLayout := engine.Layout(lm.inspectorVNode, constraints)
        result.InspectorLayer = inspectorLayout
    }

    return result
}
```

### 2. 扩展 Paint Engine

**文件**: `internal/render/paint_engine.go`

```go
// PaintEngine 绘制引擎
type PaintEngine struct {
    // ... 现有字段
}

// PaintLayers 绘制所有层
func (pe *PaintEngine) PaintLayers(result *LayoutResult) *paint.Buffer {
    // 1. 绘制 Base 层
    baseBuffer := pe.Paint(result.BaseLayer)

    // 2. 绘制其他层
    overlays := []*paint.Buffer{}

    if result.OverlayLayer != nil {
        overlays = append(overlays, pe.Paint(result.OverlayLayer))
    }
    if result.ModalLayer != nil {
        overlays = append(overlays, pe.Paint(result.ModalLayer))
    }
    if result.TooltipLayer != nil {
        overlays = append(overlays, pe.Paint(result.TooltipLayer))
    }

    // 3. 新增：绘制 Inspector 层
    if result.InspectorLayer != nil {
        overlays = append(overlays, pe.Paint(result.InspectorLayer))
    }

    // 4. 合成所有层
    return pe.ComposeLayers(baseBuffer, overlays)
}

// ComposeLayers 合成多个 buffer
func (pe *PaintEngine) ComposeLayers(
    base *paint.Buffer,
    overlays []*paint.Buffer,
) *paint.Buffer {
    // 使用现有的 Compositor
    return pe.compositor.Compose(base, overlays)
}
```

### 3. 扩展 Compositor

**文件**: `runtime/paint/compositor.go`

```go
// Compositor 合成多个层
type Compositor struct {
    // ... 现有字段
}

// Compose 合成 base 和 overlays
func (c *Compositor) Compose(
    base *paint.Buffer,
    overlays []*paint.Buffer,
) *paint.Buffer {
    // 创建结果 buffer
    result := base.Clone()

    // 按 z-index 顺序合成
    for _, overlay := range overlays {
        c.composeOverlay(result, overlay)
    }

    return result
}

// composeOverlay 合成单个覆盖层
func (c *Compositor) composeOverlay(
    base *paint.Buffer,
    overlay *paint.Buffer,
) {
    width, height := base.Size()

    for y := 0; y < height; y++ {
        for x := 0; x < width; x++ {
            // 获取 overlay 的 cell
            cell, ok := overlay.GetCell(x, y)
            if !ok || cell.Empty {
                continue // 跳过空 cell
            }

            // 检查 z-index
            if cell.ZIndex > 0 {
                // 覆盖 base 的 cell
                base.SetContent(x, y, cell.Rune, cell.Style, cell.ZIndex)
            }
        }
    }
}
```

### 4. 扩展 Framework App

**文件**: `framework/app.go`

```go
type App struct {
    // ... 现有字段

    // Inspector 支持
    inspector        *inspector.StandaloneInspector
    inspectorEnabled bool
}

// EnableInspector 启用 Inspector
func (a *App) EnableInspector() {
    a.inspector = inspector.NewStandaloneInspector()
    a.inspector.Enable()
    a.inspectorEnabled = true

    // 注册到 LayerManager
    a.layerManager.EnableInspector(a.inspector.RenderOverlay())
}

// DisableInspector 禁用 Inspector
func (a *App) DisableInspector() {
    a.inspectorEnabled = false
    a.layerManager.DisableInspector()
}

// ToggleInspector 切换 Inspector
func (a *App) ToggleInspector() {
    if a.inspectorEnabled {
        a.DisableInspector()
    } else {
        a.EnableInspector()
    }
    a.dirty = true
}

// render 渲染（修改版）
func (a *App) render() {
    // 1. 构建 VNode 树
    root := a.rootFn()
    a.root = root

    // 2. 附加 Inspector（仅分析）
    if a.inspector != nil {
        a.inspector.AttachToApp(root)
        // 更新 Inspector VNode 到 LayerManager
        if a.inspectorEnabled {
            a.layerManager.inspectorVNode = a.inspector.RenderOverlay()
        }
    }

    // 3. Layout（包含所有层）
    layoutResult := a.layerManager.CollectAndLayout(
        root,
        a.engine,
        a.constraints,
    )

    // 4. Paint（包含所有层）
    buffer := a.paintEngine.PaintLayers(layoutResult)

    // 5. Render
    a.renderer.Render(buffer)
}
```

### 5. Inspector 覆盖层渲染

**文件**: `internal/inspector/standalone_inspector.go`

```go
// RenderOverlay 渲染 Inspector 覆盖层
// 用于框架级覆盖层方案（作为 z-index 最高层）
func (si *StandaloneInspector) RenderOverlay() ui.VNode {
    si.mu.RLock()
    defer si.mu.RUnlock()

    if !si.visible {
        return nil
    }

    // 设置为 Inspector 层
    inspectorVNode := si.buildInspectorOverlay()

    // 设置 layer 属性
    if props := inspectorVNode.Props(); props != nil {
        props["layer"] = LayerInspector
    }

    return inspectorVNode
}

// buildInspectorOverlay 构建覆盖层 UI
func (si *StandaloneInspector) buildInspectorOverlay() ui.VNode {
    // 使用半透明背景（如果终端支持）
    overlay := ui.Bordered().
        Style(string(theme.Info())).
        Child(si.buildOverlayContent()).
        Width(40).   // 固定宽度
        Height(30).  // 固定高度
        Build()

    // 设置位置（右上角）
    // 注意：需要 Layout 支持 absolute positioning
    return overlay
}

// buildOverlayContent 构建覆盖层内容
func (si *StandaloneInspector) buildOverlayContent() ui.VNode {
    return ui.VStack(
        si.buildTabBar(),
        ui.Text("─"),
        si.buildActiveTabContent(),
    )
}
```

---

## 📝 实现步骤

### Phase 1: Layer 系统扩展（2-3天）

**任务**：
1. [ ] 添加 `LayerInspector` 常量
   - 文件：`runtime/ui/vnode.go`

2. [ ] 扩展 `LayerManager`
   - 添加 Inspector 支持方法
   - 修改 `CollectAndLayout()` 处理 Inspector 层
   - 文件：`runtime/layer/manager.go`

3. [ ] 修改 `LayoutResult`
   - 添加 `InspectorLayer` 字段
   - 文件：相关结构定义

**验收标准**：
- Layer 系统支持 Inspector 层
- 编译通过
- 单元测试通过

### Phase 2: Paint Engine 扩展（2天）

**任务**：
1. [ ] 扩展 `PaintEngine`
   - 修改 `PaintLayers()` 处理 Inspector 层
   - 文件：`internal/render/paint_engine.go`

2. [ ] 增强 `Compositor`
   - 改进 `Compose()` 方法
   - 优化 z-index 合成逻辑
   - 文件：`runtime/paint/compositor.go`

3. [ ] 性能优化
   - 只重绘 dirty 层
   - 局部更新优化

**验收标准**：
- 所有层正确合成
- 性能无明显退化
- 内存使用合理

### Phase 3: Inspector 覆盖层（2天）

**任务**：
1. [ ] 实现 `RenderOverlay()`
   - 返回带 Layer 标记的 VNode
   - 文件：`internal/inspector/standalone_inspector.go`

2. [ ] 设计覆盖层 UI
   - 紧凑型布局
   - 位置固定（如右上角）
   - 可拖动（可选）

3. [ ] 事件处理
   - 覆盖层内事件
   - 传递到应用的事件
   - 焦点管理

**验收标准**：
- Inspector 正确显示为覆盖层
- 不影响应用布局
- 交互正常

### Phase 4: Framework 集成（2天）

**任务**：
1. [ ] 扩展 `App` 结构
   - 添加 Inspector 相关字段
   - 文件：`framework/app.go`

2. [ ] 实现切换逻辑
   - `EnableInspector()`
   - `DisableInspector()`
   - `ToggleInspector()`

3. [ ] 修改渲染流程
   - 集成 Inspector 到渲染管道
   - 更新 LayerManager

4. [ ] 热键支持
   - 注册 F12
   - 调用 ToggleInspector()

**验收标准**：
- F12 切换流畅
- Inspector 正常显示/隐藏
- 应用完全不受影响

### Phase 5: 高级功能（2-3天）

**任务**：
1. [ ] 位置调整
   - 右上角 / 右下角 / 左上角 / 左下角
   - 用户配置

2. [ ] 尺寸调整
   - 固定尺寸选项
   - 拖动调整（可选）

3. [ ] 透明度支持
   - 如果终端支持真彩色
   - 半透明背景

4. [ ] 多 Inspector 支持
   - 为未来预留
   - 多实例管理

**验收标准**：
- 高级功能正常工作
- 用户体验良好
- 配置系统完善

### Phase 6: 测试与优化（2天）

**任务**：
1. [ ] 单元测试
   - Layer 系统测试
   - Paint Engine 测试
   - Compositor 测试

2. [ ] 集成测试
   - 端到端测试
   - 性能测试
   - 压力测试

3. [ ] 优化
   - 性能分析
   - 内存优化
   - 渲染优化

**验收标准**：
- 所有测试通过
- 性能达标
- 无明显 bug

---

## 📂 文件清单

### 需要修改的文件

1. **Layer 系统**
   - `runtime/ui/vnode.go` - 添加 LayerInspector
   - `runtime/layer/manager.go` - 扩展 LayerManager
   - `runtime/layer/collector.go` - 可能需要调整

2. **渲染系统**
   - `internal/render/paint_engine.go` - 扩展 PaintLayers
   - `runtime/paint/compositor.go` - 增强 Compose
   - `internal/render/rendering_pipeline.go` - 可能需要调整

3. **Framework**
   - `framework/app.go` - 添加 Inspector 支持

4. **Inspector**
   - `internal/inspector/standalone_inspector.go` - 添加 RenderOverlay()

### 需要创建的文件

1. **测试文件**
   - `runtime/layer/manager_test.go` - LayerManager 测试
   - `internal/render/paint_engine_test.go` - PaintEngine 测试
   - `runtime/paint/compositor_test.go` - Compositor 测试

2. **示例**
   - `examples/ui_demos/demo2_runtime_internals/inspector_overlay/main.go`
   - `examples/ui_demos/demo2_runtime_internals/inspector_overlay/README.md`

---

## ⚙️ 配置选项

### 环境变量

```bash
# 启用 Inspector 覆盖层模式
export TUI_INSPECTOR_MODE=overlay

# 覆盖层位置
export TUI_INSPECTOR_POSITION=top-right  # top-left, top-right, bottom-left, bottom-right

# 覆盖层尺寸
export TUI_INSPECTOR_WIDTH=40
export TUI_INSPECTOR_HEIGHT=30
```

### 代码配置

```go
// 在应用初始化时
app := ui.Run(myApp,
    ui.WithInspectorOverlay(
        inspector.Position(inspector.PositionTopRight),
        inspector.Size(40, 30),
    ),
)
```

---

## 📊 优势与劣势

### ✅ 优势

1. **真正的覆盖层**
   - 不影响应用布局
   - 同时显示应用和 Inspector
   - 符合用户期望

2. **框架集成**
   - 利用现有 Layer 系统
   - 架构清晰
   - 易于维护

3. **可扩展性**
   - 支持未来的覆盖层需求
   - 可添加更多层类型
   - 灵活的配置

4. **用户体验**
   - 无需频繁切换
   - 实时查看应用状态
   - 流畅的交互

### ❌ 劣势

1. **复杂度高**
   - 修改多个核心模块
   - 需要深入理解框架
   - 调试难度大

2. **性能开销**
   - 多 buffer 合成
   - 可能影响渲染性能
   - 需要仔细优化

3. **实现周期长**
   - 需要较长时间
   - 测试工作量大
   - 风险较高

---

## 🎯 适用场景

### 适合使用此方案的场景

1. **实时调试**
   - 需要同时查看应用和 Inspector
   - 实时监控应用状态

2. **交互调试**
   - 在应用上操作并立即看到反馈
   - 动态查看布局变化

3. **长期方案**
   - 作为框架的永久功能
   - 支持多种覆盖层需求

### 不适合的场景

1. **快速原型**
   - 需要快速实现
   - 时间紧迫

2. **简单场景**
   - 不需要覆盖层功能
   - 方案2 已足够

---

## 📖 参考资料

### 相关文件

1. **Layer 系统**
   - `runtime/ui/vnode.go` - Layer 类型定义
   - `runtime/layer/manager.go` - LayerManager
   - `runtime/layer/collector.go` - Layer 收集器

2. **渲染系统**
   - `internal/render/rendering_pipeline.go` - 渲染管道
   - `internal/render/paint_engine.go` - 绘制引擎
   - `runtime/paint/compositor.go` - Buffer 合成器
   - `runtime/paint/buffer.go` - Buffer 结构

3. **Framework**
   - `framework/app.go` - 应用主逻辑
   - `framework/screen/manager.go` - 屏幕管理

### 相关文档

- `docs/plan/inspector_mode_switching.md` - 方案2 计划
- `INSPECTOR_ARCHITECTURE_COMPARISON.md` - 架构对比
- `docs/plan/ui_inspector_design.md` - UI Inspector 设计

---

## 📅 时间估算

| Phase | 任务 | 预计时间 |
|-------|------|---------|
| Phase 1 | Layer 系统扩展 | 2-3 天 |
| Phase 2 | Paint Engine 扩展 | 2 天 |
| Phase 3 | Inspector 覆盖层 | 2 天 |
| Phase 4 | Framework 集成 | 2 天 |
| Phase 5 | 高级功能 | 2-3 天 |
| Phase 6 | 测试与优化 | 2 天 |
| **总计** | | **12-15 天** |

---

## ✅ 验收标准

### 功能验收

- [ ] Inspector 作为覆盖层显示
- [ ] 应用完全不受影响
- [ ] F12 切换流畅
- [ ] 所有交互正常
- [ ] 性能满足要求（< 50ms 渲染时间）

### 质量验收

- [ ] 代码审查通过
- [ ] 单元测试覆盖率 > 80%
- [ ] 集成测试全部通过
- [ ] 无已知 bug
- [ ] 文档完整

### 性能验收

- [ ] 渲染时间增加 < 20%
- [ ] 内存使用增加 < 10MB
- [ ] 切换延迟 < 50ms
- [ ] 无明显卡顿

---

## 🚀 风险与缓解

### 主要风险

1. **性能风险**
   - 风险：多 buffer 合成可能影响性能
   - 缓解：仔细优化，使用 dirty region

2. **复杂度风险**
   - 风险：修改核心模块可能引入 bug
   - 缓解：充分的测试，逐步实施

3. **兼容性风险**
   - 风险：可能破坏现有功能
   - 缓解：向后兼容测试

### 降低风险的措施

1. **渐进式实施**
   - 分阶段实施
   - 每阶段充分测试
   - 保持主分支稳定

2. **特性开关**
   - 使用环境变量控制
   - 可以快速禁用
   - 不影响默认行为

3. **完善的测试**
   - 单元测试
   - 集成测试
   - 性能测试

---

## 📊 方案对比总结

| 维度 | 方案2 (模式切换) | 方案3 (覆盖层) |
|------|-----------------|---------------|
| **实现难度** | 简单 | 复杂 |
| **开发时间** | 5-6 天 | 12-15 天 |
| **用户体验** | 需要切换 | 同时显示 |
| **性能影响** | 几乎无 | 有一定开销 |
| **可扩展性** | 有限 | 优秀 |
| **维护成本** | 低 | 中等 |
| **适用场景** | 深度调试 | 实时调试 |
| **推荐程度** | 快速实施 | 长期方案 |

---

**创建日期**: 2025-02-08
**状态**: 📋 计划中
**预计完成**: 12-15 天
**优先级**: 中（取决于需求）
**依赖**: 现有 Layer 系统稳定性
