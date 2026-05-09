# Fiber-First 快速参考指南

## 核心概念

| 层级 | 角色 | 生命周期 | 示例 |
|---|---|---|---|
| VNode | 声明输入 | 每次 render 创建，用于 reconcile | `ButtonVNode{label: "Save"}` |
| Fiber | 结构节点 | 跨 render 复用 | `Fiber{DiffKey: "save", NodeID: 42}` |
| ComponentInstance | 行为和状态 | 跨 render 复用，挂在 `Fiber.Instance` | `ButtonInstance{focused: true}` |
| PaintableBox | paint tree 节点 | 布局/绘制阶段产物 | `PaintableBox{X,Y,Width,Height,Layer}` |

## 当前主路径

```text
ComponentFunc
  -> VNode tree
  -> Fiber reconciler
  -> Fiber tree + ComponentInstance
  -> FiberToNodeAdapter
  -> runtime/layout.Engine
  -> optional PortalAwareLayoutEngine
  -> FiberToPaintableConverter
  -> paint.PaintablePlanes
  -> PaintEngine
  -> paint.Buffer
```

## 关键原则

```text
VNode 是声明输入。
Fiber 保存结构、身份、props/style/layer/layout inputs 和运行时 Instance。
ComponentInstance 保存组件状态和生命周期。
PaintableBox 是 paint 阶段节点，不是组件实例。
```

默认 Fiber-first layout/paint 不应依赖 VNode 树；但 VNode 仍是每次 render 的声明输入，legacy render/event fallback 也仍存在。

## Scheduling 状态

Lane 已在 Fiber 类型中定义并作为 metadata 传播，但主 reconciler 当前通过 `workLoopSync()` 同步处理整棵树。默认路径尚未实现：

- lane-based preemption
- time slicing
- interruptible/resumable work loop

如需描述这些能力，请标注为目标设计或实验能力。

## DiffKey 规则

当前 diff 使用 `DiffKey`：

- 用户 key 优先: `vnode.Key()`
- 无用户 key 时使用 fallback: `_idx_<siblingIndex>`
- `Fiber.Key` 当前可视为 `DiffKey` 的兼容字段
- `Fiber.Path` / `PathSegment` 主要用于 debug、inspector、hitmap 追踪，不参与 diff
- 动态列表无稳定用户 key 时会触发检查，避免状态错配

## 创建 Fiber

```go
func CreateFiber(vnode VNode) *Fiber {
    // Extract props, style, layer, layout inputs and IDs from VNode.
    var instance ComponentInstance
    if factory, ok := vnode.(InstanceFactory); ok {
        instance = factory.CreateInstance()
    }

    return &Fiber{
        Props:    props,
        Style:    style,
        Instance: instance,
    }
}
```

`CloneFiber()` 会复用 `Instance`，不会克隆组件实例。

## Paint 优先级

`FiberPaintableNode.Paint()` 当前按以下顺序绘制：

1. 如果 `Fiber.Instance` 实现 `PaintableInstance`，调用实例绘制。
2. 否则按 `Fiber.Tag` 查询 PaintRegistry 的 stateless paint fallback。
3. 没有实现则不输出。

## Focus 模型

当前焦点模型是 Fiber-first：

- `FiberFocusManager` 收集可聚焦 Fiber。
- 可聚焦能力来自实例接口或兼容 focus API。
- 焦点状态写入 `ComponentInstance`，不是写入 VNode。
- `FocusableVNode` / `FocusableMeta` 属于兼容或历史概念，不是主模型。

## Event 状态

当前状态：

- `framework.App.processMsg` 是主消息/Action 路径。
- 鼠标消息通过 HitMap 可带 `TargetID`、`TargetFiber`、`TargetBounds`、`LocalX/Y`。
- Focus 相关事件由 Fiber focus manager 处理。
- `DeclarativeNode.HandleEvent()` 仍有 VNode tree fallback，会调用 legacy `frameworkevent.Component.HandleEvent()`。

目标方向：

- 使用 HitMap/TargetFiber/ActionBridge 完成事件路由。
- 避免运行时事件系统遍历 VNode 树。

## 关键文件位置

| 功能 | 文件 |
|---|---|
| Fiber 定义 | `runtime/ui/fiber.go` |
| Fiber 创建/克隆 | `runtime/ui/fiber_util.go` |
| ComponentInstance 接口 | `runtime/ui/instance.go` |
| Reconciler 主流程 | `internal/reconciler/reconciler.go` |
| BeginWork / CompleteWork | `internal/reconciler/begin_work.go`, `internal/reconciler/complete_work.go` |
| DiffKey / children reconcile | `internal/reconciler/diff.go` |
| DeclarativeNode Fiber path | `internal/render/declarative_node.go` |
| Fiber layout adapter | `internal/render/fiber_adapter.go` |
| Fiber -> Paintable conversion | `internal/render/converter.go` |
| PaintEngine | `internal/render/paint_engine.go` |
| Portal-aware layout | `internal/render/portal_layout_adapter.go` |
| ActionBridge | `runtime/bridge/actionbridge/bridge.go` |
| Button example | `ui/components/button/` |

## 调试技巧

查找不该持久引用 VNode 的代码：

```bash
rg "fiber\\.VNode|VNode.*Fiber|vnode\\." --glob "*.go"
```

查看 Fiber/render 相关日志：

```bash
TUI_LOG_OUTPUT=console TUI_DEBUG_FIBER=true TUI_DEBUG_RENDER=true go run ./examples/counter
```

查看 HitMap 和鼠标目标：

```bash
TUI_LOG_OUTPUT=console TUI_DEBUG_HITMAP=true TUI_DEBUG_PUMP=true go run ./examples/modal
```

## 成功标准

务实标准：

> Reconciliation 完成后，默认 Fiber-first layout/paint 路径不需要保留或遍历 VNode tree；VNode 仍作为下一次 render 的声明输入存在。

这比“删除 VNode struct 后还能运行”更符合当前源码现实。
