# 渲染管线简化完成

## 概述

本文档记录了已完成的重构工作，包括：
1. **Scheduler 接口解耦** - 移除 `framework.App` 依赖
2. **Fiber-first 默认启用** - 简化渲染模式
3. **Portal-aware layout 默认启用** - 简化 Portal 处理
4. **移除 Legacy Paint** - 清理旧的渲染代码

---

## 1. Scheduler 接口解耦

### 目标

解除 `internal/reconciler` 包对 `framework.App` 的强耦合。

### 变更

#### 引入 Scheduler 接口

文件：`internal/reconciler/scheduler.go`

```go
// Scheduler 是最小化的调度器接口
// 只包含 MarkDirty 方法，用于通知框架需要重新渲染
type Scheduler interface {
    // MarkDirty 标记应用为脏状态，请求重新渲染
    MarkDirty()
}

// framework.App 实现了此接口
```

#### 更新 DeclarativeNode

文件：`internal/render/declarative_node.go`

**之前：**
```go
// 构造函数接收 app 参数
func NewDeclarativeNodeFromFuncWithFiber(
    app framework.Scheduler,
    fn ComponentFunc,
) *DeclarativeNode {
    // ...
}

// 内部存储完整的 framework.App
type DeclarativeNode struct {
    reconciler *Reconciler
    app        *framework.App  // 强耦合
}
```

**之后：**
```go
// 构造函数不再接收 app 参数
func NewDeclarativeNodeFromFuncWithFiber(fn ComponentFunc) *DeclarativeNode {
    node := &DeclarativeNode{
        renderMode:     RenderModeFiberFirst,
        fiberFirst:     true,
    }
    reconciler := NewReconciler(node)
    node.reconciler = reconciler
    node.fiberRoot = reconciler.CreateFiberRoot()
    return node
}

// 通过 SetApp 方法设置 scheduler（可选）
func (n *DeclarativeNode) SetApp(app interface{}) {
    if scheduler, ok := app.(Scheduler); ok {
        n.reconciler.SetScheduler(scheduler)
    }
}

// 内部只存储 Scheduler 接口
type Reconciler struct {
    scheduler Scheduler  // 接口解耦
}
```

#### 使用模式变更

**之前：**
```go
// 一体化构造
node := render.NewDeclarativeNodeFromFuncWithFiber(app, buildUI)
app.SetRoot(node)
```

**之后：**
```go
// 两步构造
node := render.NewDeclarativeNodeFromFuncWithFiber(buildUI)
node.SetApp(app)  // 可选，仅交互式应用需要
app.SetRoot(node)
```

### 影响范围

| 功能 | 需要调用 SetApp() |
|------|------------------|
| 静态渲染 | ❌ 不需要 |
| 布局计算 | ❌ 不需要 |
| `ui.Run()` | ✅ 需要 |
| `setState()` | ✅ 需要 |
| Intents | ✅ 需要 |

### 好处

1. **解耦** - `internal/render` 不再依赖 `framework` 包
2. **灵活性** - 可以使用任何实现 `Scheduler` 接口的对象
3. **可选性** - 静态场景不需要完整的 App 上下文

---

## 2. Fiber-first 默认启用

### 背景

之前有三种渲染模式：
- `RenderModeLegacy` - 旧的纯 Legacy Paint
- `RenderModeFiberFirst` - Fiber优先，Legacy fallback
- `RenderModeBoth` - 双模式对比（用于调试）

现在简化为：
- `RenderModeLegacy` - **DEPRECATED**，仅用于向后兼容
- `RenderModeFiberFirst` - **默认**，Fiber-first 无 fallback

### 变更

#### 简化初始化

文件：`internal/render/declarative_node.go`

**之前：**
```go
func initFiberFirstPipeline() {
    // 检查环境变量
    useFiber := os.Getenv("MINT_FIBER_FIRST") != "false"

    if useFiber {
        fiberFirstEnabled = true
        renderMode = RenderModeFiberFirst
    }
}
```

**之后：**
```go
// 始终启用
func initFiberFirstPipeline() {
    fiberFirstEnabled = true
    renderMode = RenderModeFiberFirst
    // 移除环境变量检查
}
```

#### 简化 Fiber 模式检测

**之前：**
```go
// 使用 useFiber 标志
if useFiber {
    // Fiber-first 路径
}
```

**之后：**
```go
// 使用 reconciler != nil 检测
if reconciler != nil {
    // Fiber-first 路径
}
```

#### 移除 ComparePaint

删除了 `comparePaint()` 函数和 `RenderModeBoth` 模式。

### 好处

1. **简单** - 减少配置选项
2. **明确** - 不再有双模式混淆
3. **性能** - 不再有 fallback 开销

---

## 3. Portal-aware Layout 默认启用

### 背景

Portal 组件需要特殊的布局处理：
- **主树阶段**：忽略 Portal，收集所有 Portal 根节点
- **Overlay 阶段**：独立处理 Portal，使用 Root 坐标系

之前需要显式启用，现在默认启用。

### 变更

#### 添加初始化函数

文件：`internal/render/declarative_node.go`

```go
func initPortalLayoutSupport() {
    // 默认启用，可通过环境变量禁用
    val := os.Getenv("MINT_PORTAL_LAYOUT")
    usePortalLayout := val != "" && val != "0" && strings.ToLower(val) != "false"

    if usePortalLayout {
        // 默认就是启用的
        // 这里只保留禁用的逻辑
    }
}
```

### 环境变量控制

| 环境变量 | 值 | 效果 |
|----------|-----|------|
| (未设置) | - | **启用** Portal layout |
| `MINT_PORTAL_LAYOUT=1` | 1 | 启用 Portal layout |
| `MINT_PORTAL_LAYOUT=0` | 0 | 禁用 Portal layout |
| `MINT_PORTAL_LAYOUT=false` | false | 禁用 Portal layout |

### 好处

1. **默认正确** - 大多数场景不需要配置
2. **可回退** - 通过环境变量仍可禁用
3. **向前兼容** - `SetUsePortalLayout()` 方法仍可用

---

## 4. 移除 Legacy Paint Fallback

### 背景

之前在 Fiber-first 模式下，如果遇到无法处理的情况会 fallback到 legacyPaint。这导致了：
- 代码维护成本高
- 路径切换复杂
- 调试困难

### 变更

#### 移除 fallback 调用

文件：`internal/render/declarative_node.go`

**之前的 fiberFirstPaint：**
```go
func (n *DeclarativeNode) fiberFirstPaint(buffer *paint.Buffer) error {
    // ... 尝试 Fiber-first ...
    if err != nil {
        // 回退到 legacyPaint
        return n.legacyPaint(buffer)
    }
}
```

**之后：**
```go
func (n *DeclarativeNode) fiberFirstPaint(buffer *paint.Buffer) error {
    // ... Fiber-first 路径 ...
    if err != nil {
        // 记录错误，不 fallback
        log.PaintLogger.Error("Fiber-first paint failed: %v", err)
        return err
    }
}
```

#### Paint 函数变更

```go
func (n *DeclarativeNode) Paint(buffer *paint.Buffer) error {
    // 之前：如果 renderMode == Legacy 则调用 legacyPaint
    // 之后：始终调用 fiberFirstPaint
    return n.fiberFirstPaint(buffer)
}
```

#### 移除所有 8 处 fallback

以下位置都已移除 `legacyPaint()` fallback：
1. `fiberFirstPaint()` 入口
2. Layout 主阶段错误处理
3. Portal 收集错误处理
4. Overlay 阶段错误处理
5. Fiber 树遍历错误处理
6. 绘制阶段错误处理
7. 边界检查错误处理
8. 最后的兜底处理

### 好处

1. **单一路径** - 只有一条渲染路径，更易理解
2. **更健壮** - 问题会立即暴露，而不是被隐藏
3. **更易调试** - 错误信息明确

---

## 5. API 迁移

### DeclarativeNode API

| 旧 API | 新 API | 说明 |
|--------|--------|------|
| `NewDeclarativeNodeFromFuncWithFiber(app, fn)` | `NewDeclarativeNodeFromFuncWithFiber(fn)` + `SetApp(app)` | 两步构造 |

### RenderMode

| 旧值 | 新值 | 说明 |
|------|------|------|
| `RenderModeLegacy` | `RenderModeLegacy` | 保留，标记为 DEPRECATED |
| `RenderModeFiberFirst` | `RenderModeFiberFirst` | 默认 |
| `RenderModeBoth` | **移除** | 删除 |

### Fiber 模式检测

| 旧方式 | 新方式 |
|--------|--------|
| `useFiber` 标志变量 | `reconciler != nil` |

---

## 6. 文件变更列表

### 修改的文件

| 文件 | 变更类型 |
|------|----------|
| `internal/reconciler/scheduler.go` | 新增 Scheduler 接口 |
| `internal/reconciler/reconciler.go` | 使用 Scheduler 接口 |
| `internal/render/fiber_reconciler_adapter.go` | 更新为 Scheduler 接口 |
| `internal/render/declarative_node.go` | 主要重构 |
| `examples/fiber_firsts/portal_demo/debug/main.go` | 注释更新 |
| `examples/fiber_demos/layout_comparison/main.go` | 添加 Intent Runtime 初始化 |

### 50+ 示例/测试更新

所有使用 `NewDeclarativeNodeFromFuncWithFiber` 的文件都已更新为两步模式。

### 删除的文件

| 文件 | 原因 |
|------|------|
| `examples/complex_test/main_debug_test.go` | 已过时 |
| `examples/complex_test/main_focus_debug_test.go` | 已过时 |

---

## 7. 总结

### 完成的目标

1. ✅ 解耦 `internal/render` 和 `framework.App`
2. ✅ 默认启用 Fiber-first 模式
3. ✅ 默认启用 Portal-aware layout
4. ✅ 移除所有 Legacy Paint fallback
5. ✅ 移除 `RenderModeBoth` 和 `comparePaint`
6. ✅ 简化 API 和使用模式

### 技术债务

以下标记为 DEPRECATED，未来版本中移除：

1. `RenderModeLegacy` - 旧的渲染模式
2. `legacyPaint()` 函数
3. 相关联的测试代码

### 下一步

1. 更新文档
2. 更新示例代码
3. 考虑在 v2.0 中完全移除 Legacy Paint
