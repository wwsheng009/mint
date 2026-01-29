# Framework App vs Runtime Engine 功能对比

**文档版本**: 1.0
**创建日期**: 2026-01-29
**状态**: 框架层 `framework/app.go` 是完整的应用框架，`runtime/engine/` 是独立的渲染引擎

---

## 概述

| 特性 | `framework/app.go` | `runtime/engine/engine.go` |
|------|-------------------|---------------------------|
| **定位** | 高级应用框架 | 底层渲染引擎 |
| **目标用户** | TUI 应用开发者 | 需要精细控制的场景 |
| **依赖关系** | 独立运行 | 可被 app.go 集成（可选） |
| **稳定性** | 生产就绪 | 实验性 |

---

## 功能对比矩阵

### 1. 事件处理

| 功能 | framework/app.go | runtime/engine/engine.go |
|------|-------------------|---------------------------|
| 输入读取 | ✅ `event.Pump` | ✅ `platform.InputReader` |
| 事件转换 | ✅ framework/event 类型 | ✅ runtime/event 类型 |
| 事件分发 | ✅ Router + KeyMap | ✅ DispatchEvent (三阶段) |
| 事件过滤 | ✅ `SetEventFilter` | ❌ 无 |
| 事件队列 | ✅ Pump 内部管理 | ✅ 显式事件队列 |

**差异**: `app.go` 使用 Framework 层事件系统，`engine.go` 使用 Runtime 层事件系统（支持三阶段传播）。

### 2. 渲染系统

| 功能 | framework/app.go | runtime/engine/engine.go |
|------|-------------------|---------------------------|
| 渲染器 | ✅ `paint.Renderer` | ✅ `paint.Renderer` |
| 双缓冲 | ✅ 是 | ✅ 是 |
| Diff 优化 | ✅ 是 | ✅ 是 |
| Run Merging | ✅ 是 | ✅ 是 |
| 光标优化 | ✅ 是 | ✅ 是 |
| 输出模式 | ✅ direct/diff/debug | ❌ 仅 diff |

**差异**: `app.go` 支持多种输出模式，`engine.go` 只支持 diff 模式。

### 3. 帧调度

| 功能 | framework/app.go | runtime/engine/engine.go |
|------|-------------------|---------------------------|
| 帧率控制 | ✅ `render.Throttler` | ✅ `frameInterval` |
| 目标 FPS | ✅ `SetFPS()` | ✅ `SetFrameInterval()` |
| 实际 FPS | ✅ `ActualFPS()` | ❌ 无 |
| 自适应帧率 | ✅ `EnableAdaptiveFPS()` | ❌ 无 |
| 渲染统计 | ✅ `GetRenderStats()` | ❌ 无 |

**差异**: `app.go` 提供更完整的帧率控制 API。

### 4. 组件系统

| 功能 | framework/app.go | runtime/engine/engine.go |
|------|-------------------|---------------------------|
| 组件接口 | ✅ `component.Node` | ✅ `engine.Renderable` |
| 组件树 | ✅ 完整的 Container 系统 | ❌ 只有 root |
| 组件上下文 | ✅ `ComponentContext` | ❌ 无 |
| 组件生命周期 | ✅ Mount/Unmount | ❌ 无 |

**差异**: `app.go` 提供完整的组件管理系统。

### 5. 布局系统

| 功能 | framework/app.go | runtime/engine/engine.go |
|------|-------------------|---------------------------|
| Layout 接口 | ✅ `component.Layout` | ✅ `engine.Layoutable` |
| 内置布局 | ✅ Flex, Box, Grid 等 | ❌ 无 |
| 布局计算 | ✅ 组件树自动布局 | ❌ 需手动调用 |

**差异**: `app.go` 有丰富的布局系统。

### 6. 焦点管理

| 功能 | framework/app.go | runtime/engine/engine.go |
|------|-------------------|---------------------------|
| 焦点接口 | ✅ `Focusable` | ✅ `focus.Manager` 集成 |
| 焦点导航 | ⚠️ 内嵌在组件中 | ✅ 几何导航 |
| 焦点陷阱 | ❌ 无 | ✅ `FocusTrap` |

**差异**: `engine.go` 有更完善的焦点管理系统。

### 7. 事件传播

| 功能 | framework/app.go | runtime/engine/engine.go |
|------|-------------------|---------------------------|
| 命中测试 | ❌ 无 | ✅ `event.HitTest` |
| 三阶段传播 | ❌ 无 | ✅ Capture → Target → Bubble |
| 事件委托 | ❌ 无 | ✅ `EventDelegator` |

**差异**: `engine.go` 支持更高级的事件传播模型。

### 8. 空闲优化

| 功能 | framework/app.go | runtime/engine/engine.go |
|------|-------------------|---------------------------|
| 空闲检测 | ❌ 无（持续 60fps） | ✅ 3秒超时进入空闲 |
| 资源节省 | ❌ 无 | ✅ 空闲时停止 ticker |

**差异**: `engine.go` 有更好的资源优化。

### 9. 调试功能

| 功能 | framework/app.go | runtime/engine/engine.go |
|------|-------------------|---------------------------|
| 调试模式 | ✅ `SetDebugMode()` | ❌ 无 |
| 调试录制 | ✅ `debug.Recorder` | ❌ 无 |
| 日志输出 | ✅ 文件 + 控制台 | ❌ 无 |

**差异**: `app.go` 有完整的调试支持。

### 10. 错误处理

| 功能 | framework/app.go | runtime/engine.go |
|------|-------------------|---------------------------|
| Panic 恢复 | ✅ `core.Recovery` | ❌ 无 |
| 错误日志 | ✅ 文件记录 | ❌ 无 |

**差异**: `app.go` 有生产级的错误处理。

### 11. 主题系统

| 功能 | framework/app.go | runtime/engine.go |
|------|-------------------|---------------------------|
| 主题管理 | ✅ `theme.Manager` | ❌ 无 |
| 主题切换 | ✅ `SetTheme()` | ❌ 无 |

**差异**: `app.go` 支持主题系统。

---

## API 兼容性

### framework/app.go 独有 API

以下 API 只在 `framework/app.go` 中存在：

```go
// 主题
app.InitTheme("dark")
app.SetTheme("light")
app.GetTheme()

// 调试
app.SetDebugMode(true)
app.EnableRecovery()

// 帧率
app.SetFPS(30)
app.ActualFPS()
app.EnableAdaptiveFPS(true)
app.GetRenderStats()

// 事件过滤
app.SetEventFilter(func(ev event.Event) bool {
    return ev.Type() != event.EventQuit
})
```

### runtime/engine/engine.go 独有 API

以下 API 只在 `runtime/engine/engine.go` 中存在：

```go
// 空闲控制
engine.SetIdleTimeout(5 * time.Second)
engine.IsIdle()

// 布局框（用于命中测试）
engine.SetLayoutBoxes(boxes)

// 焦点管理
engine.SetFocusManager(focusManager)

// 直接事件投递
engine.PostEvent(ev)
engine.PostRawInput(raw)
```

---

## 代码结构对比

### framework/app.go 结构

```
App
├── 组件管理
│   ├── SetRoot(component.Node)
│   ├── GetRoot() component.Node
│   └── injectComponentContext()
├── 事件系统
│   ├── Router + KeyMap
│   ├── Pump (输入读取)
│   └── eventFilter
├── 渲染系统
│   ├── renderer (paint.Renderer)
│   ├── render() 方法
│   └── outputBufferDirect/Diff()
├── 生命周期
│   ├── Init()
│   ├── Run()
│   └── Close()
├── 高级功能
│   ├── theme.Manager
│   ├── debug.Recorder
│   ├── core.Recovery
│   └── render.Throttler
└── 配置
    ├── SetFPS()
    ├── SetDebugMode()
    └── SetTheme()
```

### runtime/engine/engine.go 结构

```
Engine
├── 核心渲染
│   ├── renderer (paint.Renderer)
│   ├── frame() 方法
│   └── RequestRepaint()
├── 事件处理
│   ├── eventQueue
│   ├── inputEvents
│   ├── convertInputLoop()
│   └── handleEvent()
├── 输入读取
│   ├── InputReader
│   └── RawInput 转换
├── 组件
│   ├── root (Renderable)
│   ├── Layoutable (可选)
│   └── Updatable (可选)
├── 焦点管理
│   ├── focusManager
│   └── SetFocusManager()
├── 命中测试
│   ├── layoutBoxes
│   └── SetLayoutBoxes()
└── 运行控制
    ├── Run()
    ├── Stop()
    └── IsRunning/IsIdle()
```

---

## 使用场景

### 使用 framework/app.go（推荐）

**适用场景**：
- 标准 TUI 应用开发
- 需要完整的组件系统
- 需要主题支持
- 需要调试工具
- 需要生产级错误处理

**示例**：
```go
app := framework.NewApp()
app.SetRoot(myComponent)
app.InitTheme("dark")
app.SetDebugMode(true)

app.Run()
```

### 使用 runtime/engine/engine.go

**适用场景**：
- 需要精细控制渲染循环
- 需要三阶段事件传播
- 需要焦点管理系统
- 自定义渲染管线

**示例**：
```go
engine := engine.New(80, 24, myRenderable)
engine.SetFocusManager(focus.NewManager(rootNode))
engine.SetLayoutBoxes(layoutBoxes)

engine.Run()
```

---

## 集成方案（未实施）

### 方案 A：App 使用 Engine（已放弃）

**问题**：
- 破坏了 app.go 的现有功能
- 删除了关键的输入处理代码
- API 不兼容

### 方案 B：Engine 作为独立模块

**推荐**：
- `runtime/engine/` 作为可选的渲染引擎
- 可以独立使用
- 未来可以由 app.go 内部集成（但不破坏现有 API）
- 用户可以选择使用哪个层次

---

## 总结

| 方面 | framework/app.go | runtime/engine/engine.go |
|------|-------------------|---------------------------|
| **完整性** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |
| **易用性** | ⭐⭐⭐⭐⭐ | ⭐⭐ |
| **功能丰富度** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |
| **性能优化** | ⭐⭐⭐ | ⭐⭐⭐⭐ |
| **生产就绪** | ⭐⭐⭐⭐⭐ | ⭐⭐ |
| **扩展性** | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ |

**结论**: `framework/app.go` 是完整的应用框架，应该作为主要使用方式。`runtime/engine/engine.go` 是实验性的底层引擎，适合高级用户和特殊场景。
