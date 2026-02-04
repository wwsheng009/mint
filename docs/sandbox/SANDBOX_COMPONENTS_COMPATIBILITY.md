# Sandbox 与新 UI 组件架构兼容性报告

> **日期**: 2026-02-04
> **版本**: v1.0
> **状态**: ✅ 完全兼容

## 概述

本报告验证了 Mint TUI 框架的 Sandbox 测试系统与新的 UI 组件架构的兼容性。

**测试结论**: ✅ 所有 Sandbox 功能与新的 `runtime/ui` VNode 系统完全兼容。

---

## 架构变更

### 之前的架构 (ui/ 包内)

```
ui/
├── vnode.go           # VNode 定义
├── element.go         # ElementVNode
├── component.go       # ComponentVNode
├── hooks.go           # Hooks 实现
└── ...                # 其他
```

### 当前架构 (runtime/ui/ + ui/ 重导出)

```
runtime/ui/            # 核心类型定义
├── vnode.go           # VNode 接口
├── element.go         # ElementVNode
├── component.go       # ComponentVNode
├── hooks.go           # Hooks 实现
└── ...

ui/                    # 重导出层
├── vnode.go           # 类型别名重导出
├── element.go         # 函数重导出
├── hooks.go           # Hooks 函数
└── compat.go          # 兼容层
```

---

## 兼容性验证结果

### ✅ 测试模式

| 测试模式 | API | 状态 | 说明 |
|---------|-----|------|------|
| Mock Sandbox | `ui.RunTest(app, opts...)` | ✅ | 新版 API，支持完整框架功能 |
| Mock Sandbox | `ui.RunTestWithSandbox(app, opts...)` | ✅ | 支持 Sandbox 高级功能 |
| Direct MockSandbox | `mock.New(width, height)` | ✅ | 直接使用 MockSandbox |
| TestHelper | `sb.Helper()` 链式 API | ✅ | 流畅的测试辅助器 |

### ✅ 组件 API

| 组件类型 | API | 状态 | 说明 |
|---------|-----|------|------|
| Text | `app.Text()`, `app.NewTextBuilder()` | ✅ | 支持样式构建器 |
| Button | `app.ButtonBuilder().OnClick()` | ✅ | Builder 模式 |
| Input | `app.InputBuilder()` | ✅ | 表单输入组件 |
| Layout | `ui.VStack()`, `ui.HStack()`, `ui.Box()` | ✅ | 布局容器 |

### ✅ Sandbox 功能

| 功能 | API | 状态 | 说明 |
|------|-----|------|------|
| 事件注入 | `InjectKey()`, `InjectSpecialKey()` | ✅ | 键盘/鼠标事件 |
| 事件录制 | `sandbox.NewEventRecorder()` | ✅ | 记录/回放 |
| 快照系统 | `Snapshot()`, `Restore()` | ✅ | 状态保存/恢复 |
| 队列统计 | `QueueStats()` | ✅ | 性能监控 |
| TestHelper | 链式 API | ✅ | 简化测试代码 |

### ✅ 框架集成

| 框架特性 | API | 状态 | 说明 |
|---------|-----|------|------|
| VNode 接口 | `ui.VNode` | ✅ | 统一的节点接口 |
| Hooks | `UseStateInt()`, `UseEffect()` | ✅ | 状态管理 |
| Fiber Reconciler | 自动集成 | ✅ | 虚拟 DOM diff |
| 事件分发 | 自动集成 | ✅ | 事件处理系统 |

---

## 测试覆盖

### 运行的测试

```
examples/sandbox/01_event_recording    ✅ PASS
examples/sandbox/02_snapshot             ✅ PASS
examples/sandbox/03_test_helper          ✅ PASS
examples/sandbox/04_queue_stats          ✅ PASS
examples/sandbox/05_injection_strategy   ✅ PASS
examples/sandbox/06_comprehensive         ✅ PASS
examples/sandbox/demo                    ✅ PASS
examples/sandbox/demo/compatibility       ✅ PASS (新增)
```

### 新增兼容性测试

**文件**: `examples/sandbox/demo/sandbox_compatibility_test.go`

测试内容:
1. `TestSandboxWithUIComponents` - UI 组件兼容性
   - BasicUIComponents - 基本 UI 组件测试
   - DirectSandboxAPI - 直接 Sandbox API 测试
   - StyledText - 样式文本测试
   - LayoutComponents - 布局组件测试

2. `TestSandboxCompatibilitySummary` - 兼容性摘要

3. `TestSandboxFeatureMatrix` - 功能矩阵测试
   - EventInjection - 事件注入
   - Snapshot - 快照系统
   - QueueStats - 队列统计
   - EventRecording - 事件录制

---

## API 变更对照表

### 创建测试应用

| 之前 | 当前 | 说明 |
|------|------|------|
| `ui.TestRun(app, opts...)` | `ui.RunTest(app, opts...)` | 推荐使用新版 API |
| N/A | `ui.RunTestWithSandbox(app, opts...)` | 支持高级功能 |

### 组件创建

| 之前 | 当前 | 说明 |
|------|------|------|
| `ui.Text("content")` | `app.Text("content")` | 使用 app 包 |
| `ui.Button("label", onClick)` | `app.ButtonBuilder("label").OnClick(onClick).Build()` | Builder 模式 |
| `ui.VStack(children...)` | `ui.VStack(children...)` | 保持不变 |

### Sandbox 选项

| 之前 | 当前 | 说明 |
|------|------|------|
| `ui.TestWithSize(w, h)` | `ui.WithSize(w, h)` | 选项名称变更 |
| N/A | `ui.WithWidth(w)` | 新增单独设置宽度 |
| N/A | `ui.WithHeight(h)` | 新增单独设置高度 |

---

## 兼容层说明

### ui/compat.go

为保持向后兼容，`ui/compat.go` 提供了旧组件类型的访问器方法：

```go
// InputVNode 访问器
func (n *InputVNode) Value() string
func (n *InputVNode) Placeholder() string

// ButtonVNode 访问器
func (n *ButtonVNode) Label() string
func (n *ButtonVNode) OnClick() func()
func (n *ButtonVNode) Disabled() bool

// 其他组件类似...
```

这些访问器确保 `internal/reconciler` 和其他内部模块可以正常工作。

---

## 迁移指南

### 从旧 API 迁移到新 API

#### 1. 导入路径

```go
// 无需更改，继续使用 ui 包
import "github.com/wwsheng009/mint/ui"
```

#### 2. 组件创建

```go
// 旧代码 (仍然可用)
text := ui.Text("Hello")
button := ui.Button("Click", func() {})

// 新代码 (推荐)
text := app.Text("Hello")
button := app.ButtonBuilder("Click").OnClick(func() {}).Build()

// 样式化文本
text := app.NewTextBuilder("Hello").
    FgColor("green").
    Bold(true).
    Build()
```

#### 3. 测试代码

```go
// 旧 API (已弃用但仍可用)
testApp, _ := ui.TestRun(MyApp, ui.TestWithSize(80, 24))

// 新 API (推荐)
testApp, _ := ui.RunTest(MyApp, ui.WithSize(80, 24))

// 带 Sandbox 高级功能
testApp, _ := ui.RunTestWithSandbox(MyApp, ui.WithSize(80, 24))
```

---

## 已知问题

无。所有现有 Sandbox 功能与新的 UI 组件架构完全兼容。

---

## 总结

1. **完全兼容**: Sandbox 系统与新的 `runtime/ui` VNode 系统完全兼容
2. **API 稳定**: 现有测试代码无需修改即可继续使用
3. **新功能**: `ui.RunTest()` 和 `ui.RunTestWithSandbox()` 提供更好的集成
4. **向后兼容**: `ui/compat.go` 确保内部模块正常工作

---

**文档版本**: v1.0
**最后更新**: 2026-02-04
**作者**: Mint TUI 框架团队
