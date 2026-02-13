# Standalone UI Inspector

**独立 UI Inspector** - 类似浏览器开发者工具的独立调试界面

---

## 📖 概念说明

### 什么是 Standalone Inspector？

**Standalone Inspector** 是一个**独立的覆盖层（overlay）界面**，类似于 Chrome DevTools 或 Firefox Developer Tools。

**与集成版本的关键区别**：

| 特性 | Integrated Inspector | Standalone Inspector |
|------|---------------------|---------------------|
| **架构** | 集成到应用的 VNode 树中 | 独立的覆盖层 |
| **实现方式** | 修改应用 UI 结构 | 渲染为独立层 |
| **显示方式** | 作为侧边栏嵌入 | 叠加在应用之上 |
| **应用代码** | 需要修改 | **无需修改** |
| **隔离性** | 与应用耦合 | 与应用解耦 |
| **类似工具** | React DevTools (inline) | Chrome DevTools (window) |

---

## 🎯 设计理念

### 1. 完全独立

```go
// 应用代码保持不变
func MyApp() ui.VNode {
    return ui.VStack(
        Header(),
        Content(),
        Footer(),
    )
}

// Inspector 在外部附加
func main() {
    inspector := NewStandaloneInspector()
    inspector.Enable()

    root := MyApp()  // 原始应用
    inspector.AttachToApp(root)  // 附加分析

    if inspector.IsVisible() {
        overlay := inspector.RenderOverlay()
        return renderWithOverlay(root, overlay)  // 叠加显示
    }

    return root  // 正常显示
}
```

### 2. 非侵入式

- ✅ **不修改** 应用 VNode 树
- ✅ **不改变** 应用布局
- ✅ **不影响** 应用渲染逻辑
- ✅ 可以随时启用/禁用

### 3. 覆盖层渲染

```
┌─────────────────────────────────────────────┐
│  Application Layer (Original UI)            │
│  ┌──────────────────────────────────────┐  │
│  │  Header                               │  │
│  │  Content                              │  │
│  │  Footer                               │  │
│  └──────────────────────────────────────┘  │
├─────────────────────────────────────────────┤
│  Inspector Overlay (Independent)            │
│  ┌──────────────────────────────────────┐  │
│  │  [Elements] [Console] [Perf]         │  │
│  │  ──────────────────────────────────  │  │
│  │  📦 Layout Tree                      │  │
│  │  Nodes: 152 | Depth: 8               │  │
│  └──────────────────────────────────────┘  │
└─────────────────────────────────────────────┘
```

---

## 🚀 使用方法

### 运行独立 Inspector 演示

```bash
cd examples/ui_demos/demo2_runtime_internals/inspector_standalone

# 编译
go build -o demo2_standalone.exe main.go

# 运行（Inspector 关闭）
./demo2_standalone.exe

# 运行（Inspector 默认打开）
TUI_INSPECTOR=true ./demo2_standalone.exe
```

### 操作说明

1. **启动程序**
   ```bash
   ./demo2_standalone.exe
   ```

2. **点击 [I] Inspector 按钮**
   - Inspector 覆盖层出现在右侧
   - 显示 5 个标签页

3. **切换标签页**
   - Elements: 查看布局树
   - Console: 控制台消息
   - Performance: 性能指标
   - Diagnostics: 布局诊断
   - Network: 网络活动

4. **再次点击 [I] Inspector 按钮**
   - Inspector 隐藏
   - 应用恢复正常显示

---

## 📊 界面布局

### Inspector 关闭时

```
┌─ Runtime Scheduling Pipeline Visualization ─┐
│                                              │
│  [Event]    [setState]    [Scheduler]      │
│      ↓            ↓             ↓            │
└──────────────────────────────────────────────┘

┌─ Statistics ───────────────────────────────┐
│ Events:     0    Renders:    0    Buffers: 0│
└──────────────────────────────────────────────┘

┌─ Controls ─────────────────────────────────┐
│ [1] Event [2]setState [3]Scheduler         │
│ [4] Render [5]Reconcile [6] Layout          │
│ [7] Paint [0] Idle [I] Inspector          │
└──────────────────────────────────────────────┘
```

### Inspector 打开时

```
┌─ Runtime Pipeline ───┬─ UI INSPECTOR ──────┐
│                      │                      │
│  [Event] [setState]  │ [Elements] [Console] │
│       ↓      ↓       │ [Performance][Diag]  │
│                      │ ──────────────────── │
│ Statistics:          │                      │
│  Events: 0           │ 📦 Layout Tree       │
│  Renders: 0          │ Nodes: 152|Depth: 8  │
│                      │                      │
│ Controls:            │ Selected: ButtonVNode│
│  [1] [2] [3]         │ Path: /root/0/2      │
│  [4] [5] [6]         │                      │
│  [I] Inspector       │ Instructions:         │
│                      │  ↑↓: Navigate        │
│ Explanation:         │  Enter: Inspect      │
│  System idle...      │                      │
└──────────────────────┴──────────────────────┘
```

---

## 🔧 实现细节

### StandaloneInspector 结构

```go
type StandaloneInspector struct {
    mu sync.RWMutex

    // 状态
    enabled    bool
    visible    bool
    activeTab  InspectorTab

    // 数据源（独立实例）
    treeView    *TreeView
    perf        *PerformanceAnalyzer
    diagnostics *LayoutDiagnostics

    // VNode 追踪
    appRoot       ui.VNode
    selectedVNode ui.VNode

    // 覆盖层配置
    overlayWidth  int
    overlayHeight int
}
```

### 核心方法

```go
// 启用 Inspector
func (si *StandaloneInspector) Enable()

// 切换可见性（按钮点击）
func (si *StandaloneInspector) ToggleVisibility()

// 附加到应用（仅分析，不修改）
func (si *StandaloneInspector) AttachToApp(root ui.VNode)

// 渲染覆盖层
func (si *StandaloneInspector) RenderOverlay() ui.VNode
```

### 渲染流程

```go
func RuntimeDemoStandalone() ui.VNode {
    // 1. 构建应用内容
    demoContent := buildDemoContent(...)

    // 2. 附加 Inspector（仅分析）
    globalInspector.AttachToApp(demoContent)

    // 3. 检查是否显示覆盖层
    if globalInspector.IsVisible() {
        overlay := globalInspector.RenderOverlay()

        // 4. 返回应用 + 覆盖层
        return ui.HStack(demoContent, ui.Text("│"), overlay)
    }

    // 5. 否则只返回应用
    return demoContent
}
```

---

## 📈 优势对比

### Standalone Inspector 优势

1. **零侵入**
   - 不修改应用代码
   - 不改变应用结构
   - 可以随时启用/禁用

2. **清晰的边界**
   - Inspector 逻辑独立
   - 不会意外影响应用
   - 易于维护和调试

3. **真正的覆盖层**
   - 类似浏览器 DevTools
   - 符合开发者直觉
   - 可以独立更新

4. **易于集成**
   ```go
   // 只需 3 行代码
   inspector := NewStandaloneInspector()
   inspector.Enable()
   // 在渲染函数中检查 IsVisible() 并渲染覆盖层
   ```

### Integrated Inspector 局限

1. **需要修改应用**
   - 必须传递 state setter
   - 需要在 VNode 树中插入 Inspector 组件
   - 增加应用复杂度

2. **耦合度高**
   - Inspector 与应用在同一个 VNode 树
   - 可能影响应用布局
   - 难以完全隔离

3. **不符合直觉**
   - 开发者期望独立的 DevTools 窗口
   - 而不是嵌入到应用中

---

## 🎓 使用场景

### 场景 1: 日常开发

```go
func main() {
    inspector := NewStandaloneInspector()

    // 开发时启用
    if os.Getenv("DEBUG") == "true" {
        inspector.Enable()
    }

    ui.Run(MyApp, ui.WithTitle("My App"))
}
```

### 场景 2: 生产环境

```go
func main() {
    inspector := NewStandaloneInspector()

    // 生产环境默认禁用
    // 需要时通过环境变量启用
    if os.Getenv("TUI_INSPECTOR") == "true" {
        inspector.Enable()
    }

    ui.Run(MyApp)
}
```

### 场景 3: 调试特定问题

```go
func main() {
    inspector := NewStandaloneInspector()

    // 只在特定条件下启用
    if os.Getenv("DEBUG_LAYOUT") == "true" {
        inspector.Enable()
        inspector.SetActiveTab(TabDiagnostics)
    }

    ui.Run(MyApp)
}
```

---

## 🔍 标签页说明

### Elements (元素树)

显示应用的 VNode 树结构：
- 节点总数
- 最大深度
- 叶节点数
- 选中的元素信息

### Console (控制台)

显示日志消息（TODO）

### Performance (性能)

实时性能指标：
- FPS
- 内存使用
- 渲染时间
- GC 统计

### Diagnostics (诊断)

布局问题检测：
- 约束冲突
- 尺寸问题
- 内容溢出
- 修复建议

### Network (网络)

网络活动监控（TODO for TUI）

---

## 📚 相关文件

### 核心实现

- `internal/inspector/standalone_inspector.go` - 独立 Inspector 实现
- `internal/inspector/tree_view.go` - 树视图
- `internal/inspector/performance.go` - 性能分析
- `internal/inspector/layout_diagnostics.go` - 布局诊断

### 演示程序

- `examples/ui_demos/demo2_runtime_internals/inspector_standalone/main.go` - 演示程序

### 文档

- `STANDALONE_INSPECTOR.md` - 本文档
- `INSPECTOR_INTEGRATION_SUMMARY.md` - 集成版本总结
- `QUICKSTART_INSPECTOR.md` - 快速开始指南

---

## ⚙️ 配置选项

### 环境变量

```bash
# 启用 Inspector
export TUI_INSPECTOR=true

# 启用详细日志
export TUI_DEBUG_INSPECTOR=true
```

### 代码配置

```go
inspector := NewStandaloneInspector()

// 设置覆盖层尺寸
inspector.SetOverlaySize(80, 25)

// 设置位置（未来支持）
inspector.SetPosition(PositionRight)
// inspector.SetPosition(PositionFloating)
```

---

## 🚧 未来改进

1. **真正的覆盖层渲染**
   - 当前使用 HStack 并排显示
   - 未来可以实现真正的 z-index 覆盖

2. **快捷键支持**
   - F12 切换 Inspector
   - Ctrl+Shift+I 打开
   - 数字键 1-5 切换标签

3. **持久化**
   - 记住上次打开的标签
   - 保存窗口位置和大小

4. **更多标签页**
   - Sources: 源码查看
   - Network: 网络请求
   - Profiler: 性能分析

---

## 📖 总结

**Standalone Inspector** 是一个**真正的独立调试工具**：

✅ **不修改应用代码**
✅ **独立的覆盖层界面**
✅ **类似浏览器 DevTools**
✅ **随时启用/禁用**
✅ **零侵入集成**

与 Integrated Inspector 相比，Standalone Inspector 更符合开发者对调试工具的期望，是一个更好的长期解决方案。

---

**创建日期**: 2025-02-08
**状态**: ✅ 实现完成
**演示位置**: `examples/ui_demos/demo2_runtime_internals/inspector_standalone/`
