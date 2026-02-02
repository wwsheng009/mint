# Fiber Reconciler 迁移与激活计划

> **目标**: 完成 Fiber 功能的迁移，激活 Fiber reconciler，将 Fiber 从 ui/ 中解耦

---

## 现状分析

### 当前架构

```
┌─────────────────────────────────────────────────────────────┐
│                    当前状态（存在问题）                        │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  runtime/ui/          → Fiber 结构体定义 (包名: ui)          │
│  ui/fiber.go          → 重导出 runtime/ui.Fiber              │
│  ui/reconciler.go     → ❌ 旧 reconciler (遗留代码)           │
│  internal/reconciler/ → ⚠️ 依赖 ui/ 包 (违反设计)            │
│  internal/render/     → ⚠️ 未使用 reconciler                │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### 关键发现

| 类型 | 定义位置 | 说明 |
|------|---------|------|
| `Fiber` | `runtime/ui/fiber.go` | ✅ 正确位置 |
| `Lane`, `EffectFlag` | `runtime/ui/fiber.go` | ✅ 正确位置 |
| `InstanceManager` | `internal/state/instance_manager.go` | ✅ 正确位置 |
| `ComponentInstance` | `internal/state/instance.go` | ✅ 正确位置 |
| `ComponentContext` | `runtime/ui/hooks.go` | ✅ 正确位置 |
| `ui.InstanceManager` | `ui/instance_manager.go` | ⚠️ 类型别名到 internal/state |
| `ui.TextVNode`, `ui.ButtonVNode` | `ui/compat.go` | ❌ 存根类型 |

### 问题根源

`internal/reconciler/` 依赖 `ui/` 包的原因：
```go
// internal/reconciler/reconciler.go
import "github.com/wwsheng009/mint/ui"

// 使用的 ui 类型：
instanceMgr         *ui.InstanceManager         // ← 应该用 internal/state
interactionStateMgr *ui.InteractionStateManager // ← 应该用 internal/state
keyValidator        *ui.KeyValidator            // ← 应该用 internal/state
rootComponent       ui.ComponentFunc            // ← 应该用 runtime/ui
ctx                 *ui.ComponentContext        // ← 应该用 runtime/ui
rootComponentVNode  = ui.NewComponent(...)      // ← 应该用 runtime/ui
```

---

## 实施计划

### Phase 1: 修改 internal/reconciler 依赖

**目标**: 让 `internal/reconciler/` 只依赖 `runtime/ui`，不依赖 `ui/`

#### 1.1 修改 `internal/reconciler/reconciler.go`

```go
// 修改前：
import "github.com/wwsheng009/mint/ui"

// 修改后：
import (
    state "github.com/wwsheng009/mint/internal/state"
    rtui "github.com/wwsheng009/mint/runtime/ui"
)

// 类型替换：
instanceMgr         *state.InstanceManager         // 原 ui.InstanceManager
interactionStateMgr *state.InteractionStateManager // 原 ui.InteractionStateManager
keyValidator        *state.KeyValidator            // 原 ui.KeyValidator
rootComponent       rtui.ComponentFunc             // 原 ui.ComponentFunc
ctx                 *rtui.ComponentContext         // 原 ui.ComponentContext
```

#### 1.2 修改 `internal/reconciler/begin_work.go`

```go
// 修改前：
import "github.com/wwsheng009/mint/ui"

// 修改后：
import (
    state "github.com/wwsheng009/mint/internal/state"
    rtui "github.com/wwsheng009/mint/runtime/ui"
)

// 类型替换：
// ui.VNode → rtui.VNode
// ui.ComponentVNode → rtui.ComponentVNode
// ui.ComponentFunc → rtui.ComponentFunc
// ui.ComponentContext → rtui.ComponentContext
// ui.ComponentInstance → state.ComponentInstance
// ui.InstanceManager → state.InstanceManager
```

#### 1.3 修改 `internal/reconciler/complete_work.go`

同样的替换模式。

#### 1.4 修改 `internal/reconciler/diff.go`

同样的替换模式。

#### 1.5 修改 `internal/reconciler/fiber.go`

删除重复的 Fiber 类型定义，使用 `rtui.Fiber`。

#### 1.6 修改 `internal/reconciler/vnode_converter.go`

确保只使用 `rtui` 类型。

---

### Phase 2: 集成 Reconciler 到 DeclarativeNode

**目标**: 让 `DeclarativeNode.Paint()` 使用 Fiber reconciler

#### 2.1 修改 `internal/render/declarative_node.go`

```go
import (
    rtreconciler "github.com/wwsheng009/mint/internal/reconciler"
    rtui "github.com/wwsheng009/mint/runtime/ui"
)

type DeclarativeNode struct {
    mu         sync.RWMutex
    root       rtui.VNode
    renderFn   rtui.ComponentFunc
    instance   *rtui.ComponentContext
    focusMgr   *rtui.VNodeFocusManager

    // 新增字段
    reconciler *rtreconciler.Reconciler  // Fiber reconciler
    useFiber   bool                       // 是否启用 Fiber
}

// NewDeclarativeNodeFromFunc 添加 fiber 选项
func NewDeclarativeNodeFromFunc(fn rtui.ComponentFunc, enableFiber bool) *DeclarativeNode {
    node := &DeclarativeNode{
        renderFn: fn,
        instance: rtui.NewComponentContextForRoot(),
        focusMgr: rtui.NewVNodeFocusManager(),
        useFiber: enableFiber,
    }

    if enableFiber {
        // 创建 reconciler（稍后设置 app）
        node.reconciler = rtreconciler.NewReconciler(nil, fn,
            rtreconciler.ReconcilerConfig{
                EnableFiber: true,
            })
    }

    return node
}

// SetApp 设置 framework.App（用于 reconciler 调度）
func (n *DeclarativeNode) SetApp(app *framework.App) {
    if n.reconciler != nil {
        n.reconciler.SetApp(app)
    }
}

// Paint 修改为使用 reconciler
func (n *DeclarativeNode) Paint(ctx component.PaintContext, buf *paint.Buffer) {
    // ... 现有的初始化代码 ...

    if n.useFiber && n.reconciler != nil {
        // 使用 Fiber reconciler 渲染
        n.reconciler.Render(ctx, buf, n.renderFn)
    } else {
        // 原有的直接渲染逻辑
        if n.renderFn != nil {
            n.instance.ResetContext()
            rtui.SetCurrentContext(n.instance)
            n.root = n.renderFn()
            rtui.SetCurrentContext(nil)
        }
        if n.root != nil {
            n.paintVNode(n.root, ctx.X, ctx.Y, buf)
        }
    }
}
```

---

### Phase 3: 激活 Fiber 模式

**目标**: 通过环境变量控制 Fiber 模式

#### 3.1 修改 `ui/app.go`

```go
// Run 添加 Fiber 支持
func Run(app ComponentFunc, opts ...Option) error {
    // 检查是否启用 Fiber
    enableFiber := os.Getenv("MINT_USE_FIBER") == "true"

    fwApp := framework.NewApp()
    declarativeRoot := render.NewDeclarativeNodeFromFunc(app, enableFiber)
    declarativeRoot.SetApp(fwApp) // 设置 app 供 reconciler 使用

    fwApp.SetRoot(declarativeRoot)

    // ... 现有的配置代码 ...
}
```

#### 3.2 添加状态更新调度

在 `runtime/ui/hooks.go` 中的 `setState` 函数中添加调度：

```go
func (s *StateHook[T]) setState(newState T) {
    // ... 现有代码 ...

    // 如果启用了 Fiber，通知 reconciler
    if reconciler := getCurrentReconciler(); reconciler != nil {
        reconciler.ScheduleUpdate(rtui.LaneSyncLane)
    }
}
```

---

### Phase 4: 清理遗留代码

**目标**: 删除 `ui/` 中的 reconciler 相关代码

#### 4.1 删除以下文件

```bash
ui/reconciler.go
ui/begin_work.go
ui/complete_work.go
ui/diff.go
ui/scheduler.go
```

#### 4.2 保留的文件

```bash
# Fiber 重导出层（保留）
ui/fiber.go
ui/fiber_util.go  # 如果存在

# 兼容层（暂时保留，后续清理）
ui/compat.go
```

---

### Phase 5: 验证与测试

#### 5.1 运行测试

```bash
# 基础测试
go test ./...

# Fiber 模式测试
MINT_USE_FIBER=true go run examples/counter/main.go
MINT_USE_FIBER=true go run examples/fiber_counter/main.go
```

#### 5.2 验证清单

- [ ] `internal/reconciler/` 不再导入 `"github.com/wwsheng009/mint/ui"`
- [ ] `MINT_USE_FIBER=true` 时使用 Fiber reconciler
- [ ] `MINT_USE_FIBER=false` 时使用传统渲染
- [ ] 所有示例正常运行
- [ ] 状态更新正确触发重新渲染

---

## 关键文件清单

### 需要修改的文件

| 文件 | 修改类型 |
|------|---------|
| `internal/reconciler/reconciler.go` | 修改导入和类型 |
| `internal/reconciler/begin_work.go` | 修改导入和类型 |
| `internal/reconciler/complete_work.go` | 修改导入和类型 |
| `internal/reconciler/diff.go` | 修改导入和类型 |
| `internal/reconciler/fiber.go` | 删除重复类型定义 |
| `internal/render/declarative_node.go` | 集成 reconciler |
| `ui/app.go` | 添加 Fiber 选项 |
| `runtime/ui/hooks.go` | 添加调度支持（可选） |

### 需要删除的文件

| 文件 | 原因 |
|------|------|
| `ui/reconciler.go` | 遗留代码，功能已迁移 |
| `ui/begin_work.go` | 遗留代码，功能已迁移 |
| `ui/complete_work.go` | 遗留代码，功能已迁移 |
| `ui/diff.go` | 遗留代码，功能已迁移 |
| `ui/scheduler.go` | 遗留代码，未使用 |

---

## 类型映射表

| ui 类型 | runtime/ui 或 internal/state 类型 |
|---------|----------------------------------|
| `ui.VNode` | `rtui.VNode` |
| `ui.ComponentFunc` | `rtui.ComponentFunc` |
| `ui.ComponentContext` | `rtui.ComponentContext` |
| `ui.ComponentVNode` | `rtui.ComponentVNode` |
| `ui.ElementVNode` | `rtui.ElementVNode` |
| `ui.LayoutNode` | `rtui.LayoutNode` |
| `ui.Fiber` | `rtui.Fiber` |
| `ui.Lane` | `rtui.Lane` |
| `ui.EffectFlag` | `rtui.EffectFlag` |
| `ui.InstanceManager` | `state.InstanceManager` |
| `ui.InteractionStateManager` | `state.InteractionStateManager` |
| `ui.KeyValidator` | `state.KeyValidator` |
| `ui.ComponentInstance` | `state.ComponentInstance` |
| `ui.NewComponent()` | `rtui.NewComponent()` |

---

## 验证步骤

1. **编译检查**:
   ```bash
   go build ./...
   ```

2. **传统模式运行**:
   ```bash
   go run examples/counter/main.go
   ```

3. **Fiber 模式运行**:
   ```bash
   MINT_USE_FIBER=true go run examples/counter/main.go
   ```

4. **依赖检查**:
   ```bash
   grep -r "github.com/wwsheng009/mint/ui\"" internal/reconciler/
   # 应该返回空
   ```

---

## 预期结果

```
┌─────────────────────────────────────────────────────────────┐
│                    目标架构（正确）                           │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  runtime/ui/          → Fiber 结构体定义                     │
│  internal/state/      → 状态管理                             │
│  internal/reconciler/ → Reconciler (只依赖 runtime/ui)      │
│  internal/render/     → DeclarativeNode (使用 reconciler)   │
│  ui/fiber.go          → 重导出层                             │
│                                                              │
│  依赖关系:                                                   │
│  internal/ → runtime/ui  ✅                                  │
│  ui/ → runtime/ui          ✅ (重导出)                       │
│  components/ → runtime/ui  ✅                                │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```
