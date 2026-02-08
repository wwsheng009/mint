# UI Inspector - 方案2: 模式切换实现计划

**Plan 2: Screen/Mode Switching for UI Inspector**

---

## 📋 方案概述

**核心思想**：在应用模式和 Inspector 模式之间切换，同一时间只显示一个界面。

**用户体验**：
- 按 `F12` 切换：应用界面 ⇄ Inspector 界面
- 应用界面：完全不受影响，保持原有布局
- Inspector 界面：完整显示调试工具
- 不支持并排显示

---

## 🎯 设计目标

### 必须满足的需求

1. ✅ **零干扰**：Inspector 不影响应用布局
2. ✅ **状态保持**：切换模式时保持应用状态
3. ✅ **简单可靠**：实现简单，不依赖复杂特性
4. ✅ **性能良好**：切换快速，无明显延迟

### 不支持的功能

- ❌ 并排显示（应用 + Inspector）
- ❌ 覆盖层显示（Inspector 在应用上方）
- ❌ 同时查看应用和 Inspector

---

## 🏗️ 架构设计

### 系统架构

```
┌─────────────────────────────────────────────┐
│  Framework App (framework/app.go)           │
│                                             │
│  currentMode: "app" | "inspector"           │
│  appStateCache: ApplicationState            │
│  inspector: StandaloneInspector             │
└─────────────────────────────────────────────┘
           ↓                    ↓
    ┌──────────┐        ┌──────────────┐
    │ App Mode │        │Inspector Mode│
    │          │        │              │
    │ demo2    │        │ Elements     │
    │ pipeline │        │ Performance  │
    │ controls │        │ Diagnostics  │
    └──────────┘        └──────────────┘
```

### 数据流

```
User Input (F12)
    ↓
Toggle Inspector Mode
    ↓
Save Current State
    ↓
Switch Mode
    ↓
Render New Mode
    ↓
Restore Previous State (if switching back)
```

---

## 🔧 实现细节

### 1. 扩展 Framework App

**文件**: `framework/app.go`

```go
type App struct {
    // ... 现有字段

    // Inspector 模式支持
    currentMode      AppMode          // 当前模式
    inspector        *inspector.StandaloneInspector
    appStateCache    *ApplicationState // 应用状态缓存
}

type AppMode int

const (
    ModeApp AppMode = iota
    ModeInspector
)

// ApplicationState 保存应用完整状态
type ApplicationState struct {
    VNode           ui.VNode
    State           map[string]interface{}
    Timestamp       time.Time
}
```

### 2. 模式切换逻辑

```go
// ToggleInspectorMode 切换 Inspector 模式
func (a *App) ToggleInspectorMode() {
    switch a.currentMode {
    case ModeApp:
        // App → Inspector
        a.captureAppState()
        a.currentMode = ModeInspector

    case ModeInspector:
        // Inspector → App
        a.currentMode = ModeApp
        // 应用状态已在 captureAppState 中保存
    }

    a.dirty = true // 触发重新渲染
}

// captureAppState 捕获当前应用状态
func (a *App) captureAppState() {
    a.appStateCache = &ApplicationState{
        VNode:     a.root, // 当前 VNode 树
        State:     extractState(a.root),
        Timestamp: time.Now(),
    }
}

// extractState 从 VNode 树提取状态
func extractState(root ui.VNode) map[string]interface{} {
    // 遍历 VNode 树，收集所有组件状态
    states := make(map[string]interface{})

    var walk func(ui.VNode)
    walk = func(node ui.VNode) {
        if node == nil {
            return
        }

        // 提取状态（如果有）
        if instance, ok := node.(*ui.ComponentVNode); ok {
            if inst := instance.Instance; inst != nil {
                states[inst.ID()] = inst.GetState()
            }
        }

        // 递归子节点
        for _, child := range node.Children() {
            walk(child)
        }
    }

    walk(root)
    return states
}
```

### 3. 修改渲染逻辑

```go
// render 应用模式感知的渲染
func (a *App) render() {
    switch a.currentMode {
    case ModeApp:
        a.renderAppMode()
    case ModeInspector:
        a.renderInspectorMode()
    }
}

// renderAppMode 渲染应用界面
func (a *App) renderAppMode() {
    // 正常的应用渲染流程
    root := a.rootFn()
    a.root = root

    // Layout
    layout := a.engine.Layout(root, a.constraints)

    // Paint
    buffer := a.paintEngine.Paint(layout)

    // Render
    a.renderer.Render(buffer)
}

// renderInspectorMode 渲染 Inspector 界面
func (a *App) renderInspectorMode() {
    // 附加到应用（仅分析，不渲染）
    if a.appStateCache != nil {
        a.inspector.AttachToApp(a.appStateCache.VNode)
    } else {
        // 如果没有缓存，使用当前 root
        a.inspector.AttachToApp(a.root)
    }

    // 渲染 Inspector 界面
    inspectorVNode := a.inspector.RenderFullInspector()

    // Layout
    layout := a.engine.Layout(inspectorVNode, a.constraints)

    // Paint
    buffer := a.paintEngine.Paint(layout)

    // Render
    a.renderer.Render(buffer)
}
```

### 4. 扩展 StandaloneInspector

**文件**: `internal/inspector/standalone_inspector.go`

```go
// RenderFullInspector 渲染完整的 Inspector 界面
// 用于模式切换方案（占据整个屏幕）
func (si *StandaloneInspector) RenderFullInspector() ui.VNode {
    si.mu.RLock()
    defer si.mu.RUnlock()

    // Inspector 标题
    header := ui.VStack(
        app.NewTextBuilder("╔═ UI INSPECTOR ═╗").
            Style(style.FgBold(style.Yellow)).
            Build(),
        app.NewTextBuilder("Press F12 to return to app").
            Style(style.Foreground(theme.Muted())).
            Build(),
        ui.Text(""),
    )

    // 构建所有标签页
    tabs := si.buildAllTabs()

    return ui.VStack(
        header,
        tabs,
    )
}

// buildAllTabs 构建所有标签页内容
func (si *StandaloneInspector) buildAllTabs() ui.VNode {
    return ui.VStack(
        si.buildElementsTabFull(),
        ui.Text("─").
            Style(style.Foreground(theme.Muted())).
            Build(),
        si.buildPerformanceTab(),
        ui.Text("─").
            Style(style.Foreground(theme.Muted())).
            Build(),
        si.buildDiagnosticsTab(),
    )
}

// buildElementsTabFull 构建完整的元素树标签页
func (si *StandaloneInspector) buildElementsTabFull() ui.VNode {
    stats := si.treeView.GetTreeStats()

    return ui.VStack(
        app.NewTextBuilder("📦 Layout Tree").
            Style(style.FgBold(style.Green)).
            Build(),
        app.NewTextBuilder(fmt.Sprintf("Total: %d nodes | Depth: %d | Leaves: %d",
            stats.TotalNodes, stats.MaxDepth, stats.LeafNodes)).
            Style(style.Foreground(style.Cyan)).
            Build(),
        ui.Text(""),

        // 完整的树显示（不是摘要）
        app.NewTextBuilder(si.treeView.FormatTree()).
            Style(style.Foreground(style.White)).
            Build(),
    )
}
```

---

## 📝 实现步骤

### Phase 1: 基础架构（1-2天）

**任务**：
1. [ ] 扩展 `framework/app.go` 添加模式支持
   - 添加 `AppMode` 类型
   - 添加 `currentMode` 字段
   - 添加 `inspector` 和 `appStateCache` 字段

2. [ ] 实现模式切换逻辑
   - `ToggleInspectorMode()` 方法
   - `captureAppState()` 方法
   - `extractState()` 辅助函数

3. [ ] 修改 `render()` 方法
   - 添加模式感知的渲染分支

**验收标准**：
- 编译通过
- 模式切换不崩溃
- 基础架构完整

### Phase 2: Inspector 渲染（1天）

**任务**：
1. [ ] 扩展 `StandaloneInspector`
   - 实现 `RenderFullInspector()` 方法
   - 实现 `buildAllTabs()` 方法
   - 实现完整版的标签页内容

2. [ ] 实现 `renderInspectorMode()`
   - 在 `framework/app.go` 中
   - 完整的 Inspector 界面渲染

**验收标准**：
- Inspector 界面正常显示
- 所有标签页可见
- 布局合理

### Phase 3: 状态保持（1天）

**任务**：
1. [ ] 完善状态捕获
   - 提取所有组件状态
   - 保存 VNode 树快照

2. [ ] 实现状态恢复
   - 从缓存恢复状态
   - 重新渲染应用

3. [ ] 测试状态持久性
   - 切换前后状态一致
   - 计数器、输入框等正常

**验收标准**：
- 切换模式后应用状态保持
- 按钮计数器不丢失
- 输入框内容保留

### Phase 4: 事件处理（1天）

**任务**：
1. [ ] 注册 F12 热键
   - 在 `framework/app.go` 中
   - 调用 `ToggleInspectorMode()`

2. [ ] Inspector 内事件
   - 标签页切换
   - 树节点选择
   - 属性编辑

3. [ ] 应用内事件
   - 确保应用事件正常工作
   - 切换回来后焦点正确

**验收标准**：
- F12 切换流畅
- Inspector 内交互正常
- 应用内交互不受影响

### Phase 5: 集成测试（1天）

**任务**：
1. [ ] 创建测试 demo
   - 基于 demo2
   - 添加状态丰富的内容
   - 测试所有交互

2. [ ] 端到端测试
   - 启动应用
   - 交互产生状态
   - 切换 Inspector
   - 查看 Inspector 数据
   - 切回应用
   - 验证状态完整

3. [ ] 性能测试
   - 切换延迟 < 100ms
   - 内存使用合理
   - 无明显卡顿

**验收标准**：
- 所有测试通过
- 性能达标
- 用户体验良好

---

## 📂 文件清单

### 需要修改的文件

1. **`framework/app.go`**
   - 添加 `AppMode` 类型和相关字段
   - 实现 `ToggleInspectorMode()`
   - 修改 `render()` 方法

2. **`internal/inspector/standalone_inspector.go`**
   - 添加 `RenderFullInspector()` 方法
   - 添加 `buildAllTabs()` 方法
   - 扩展标签页内容

### 需要创建的文件

1. **`examples/ui_demos/demo2_runtime_internals/inspector_mode_switch/main.go`**
   - 演示程序

2. **`examples/ui_demos/demo2_runtime_internals/inspector_mode_switch/README.md`**
   - 使用说明

---

## ⚙️ 配置选项

### 环境变量

```bash
# 启用 Inspector 模式切换
export TUI_INSPECTOR_MODE=switch

# 启用详细日志
export TUI_INSPECTOR_VERBOSE=true
```

### 代码配置

```go
// 在应用初始化时
app := ui.Run(myApp,
    ui.WithInspectorModeSwitch(), // 启用模式切换
)
```

---

## 📊 优势与劣势

### ✅ 优势

1. **实现简单**
   - 修改点少，逻辑清晰
   - 不需要复杂的 buffer 合成
   - 易于调试和维护

2. **完全隔离**
   - 应用和 Inspector 完全独立
   - 零干扰，不影响应用布局
   - 状态管理清晰

3. **性能良好**
   - 只渲染一个界面
   - 无额外的 buffer 合成开销
   - 切换快速

4. **可靠性高**
   - 逻辑简单，bug 少
   - 易于测试
   - 边界情况明确

### ❌ 劣势

1. **用户体验**
   - 不能同时查看应用和 Inspector
   - 需要频繁切换
   - 可能影响调试效率

2. **状态限制**
   - 应用在 Inspector 模式下不可见
   - 无法实时看到应用变化
   - 可能错过动态行为

3. **功能受限**
   - 不支持并排显示
   - 不支持覆盖层显示
   - 未来扩展性有限

---

## 🎯 适用场景

### 适合使用此方案的场景

1. **深度调试**
   - 需要专注查看 Inspector 数据
   - 不需要实时查看应用

2. **状态分析**
   - 分析布局树
   - 查看性能指标
   - 诊断问题

3. **简单场景**
   - 应用状态变化不频繁
   - 不需要实时监控

### 不适合的场景

1. **实时调试**
   - 需要同时看应用和 Inspector
   - 实时查看应用变化

2. **交互调试**
   - 需要在应用上操作并立即看到 Inspector 反馈

---

## 📖 参考资料

### 相关文件

1. **Framework Core**
   - `framework/app.go` - 应用主逻辑
   - `internal/render/rendering_pipeline.go` - 渲染管道
   - `internal/render/paint_engine.go` - 绘制引擎

2. **Inspector**
   - `internal/inspector/standalone_inspector.go` - 独立 Inspector
   - `internal/inspector/tree_view.go` - 树视图
   - `internal/inspector/performance.go` - 性能分析

3. **State Management**
   - `runtime/ui/component.go` - 组件状态
   - `runtime/ui/instance.go` - 组件实例

### 相关文档

- `INSPECTOR_ARCHITECTURE_COMPARISON.md` - 架构对比
- `STANDALONE_INSPECTOR.md` - 独立 Inspector 说明
- `docs/plan/ui_inspector_design.md` - UI Inspector 设计文档

---

## 📅 时间估算

| Phase | 任务 | 预计时间 |
|-------|------|---------|
| Phase 1 | 基础架构 | 1-2 天 |
| Phase 2 | Inspector 渲染 | 1 天 |
| Phase 3 | 状态保持 | 1 天 |
| Phase 4 | 事件处理 | 1 天 |
| Phase 5 | 集成测试 | 1 天 |
| **总计** | | **5-6 天** |

---

## ✅ 验收标准

### 功能验收

- [ ] F12 切换应用 ⇄ Inspector
- [ ] 应用状态在切换后保持
- [ ] Inspector 显示完整信息
- [ ] 所有交互正常工作
- [ ] 性能满足要求（切换 < 100ms）

### 质量验收

- [ ] 代码审查通过
- [ ] 单元测试覆盖率 > 80%
- [ ] 集成测试全部通过
- [ ] 无已知 bug
- [ ] 文档完整

---

**创建日期**: 2025-02-08
**状态**: 📋 计划中
**预计完成**: 5-6 天
**优先级**: 高
