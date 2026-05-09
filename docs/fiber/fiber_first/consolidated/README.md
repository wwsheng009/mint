# Fiber-First 文档整理说明

本目录保存 Fiber-first 相关的合并文档和快速参考。当前源码已经采用 Fiber-first 作为 `ui.Run` 的默认声明式渲染路径，但部分早期设计目标仍未完全落地，因此阅读这些文档时需要区分“当前实现”和“目标设计”。

## 当前源码事实

默认声明式路径是：

```text
VNode declaration
  -> internal/reconciler Fiber tree
  -> persistent runtime/ui.ComponentInstance
  -> runtime/layout via FiberToNodeAdapter
  -> LayoutBox
  -> FiberToPaintableConverter
  -> paint.PaintableBox / PaintablePlanes
  -> PaintEngine
  -> paint.Buffer / terminal output
```

关键术语：

- `VNode`: 用户声明层输入，每次 render 重新创建，用于 reconcile。
- `Fiber`: 持久结构节点，保存 tree links、props、style、layer、NodeID、DiffKey、layout inputs 和 `Instance`。
- `ComponentInstance`: 持久运行时实例，挂在 `Fiber.Instance`，保存组件状态和生命周期。
- `PaintableInstance`: `ComponentInstance` 可选能力，提供组件绘制逻辑。
- `paint.PaintableBox`: layout 后生成的 paint tree 节点，包含坐标、尺寸、layer、zIndex 和 `PaintableNode`，不是组件运行时实例。

## 重要更正

早期合并文档曾把 `Instance` 和 `paint.PaintableBox` 混用，这是不准确的。当前应使用：

```text
Fiber.Instance = runtime/ui.ComponentInstance
PaintableBox = transient paint-stage layout result node
```

也就是说，`ComponentInstance` 是长期存在的组件状态实体；`paint.PaintableBox` 是每次布局/绘制阶段生成或更新的绘制树节点。

## 已实现与未实现

已在主路径落地：

- VNode 到 Fiber 的 reconcile。
- Fiber alternate / clone / NodeID / DiffKey。
- 组件实例持久化与复用。
- Fiber-first layout adapter。
- Portal-aware two-phase layout。
- Fiber 到 paintable tree 转换。
- PaintEngine 绘制 `PaintablePlanes`。
- FiberFocusManager 将焦点写入实例。

仍属于目标或部分实现：

- 主 reconciler 的可中断渲染。
- 主 reconciler 的时间切片。
- lane-based preemption。
- 完全移除 VNode event fallback。

`runtime/ui/fiber_scheduler.go` 有独立调度器实现，但 `internal/reconciler.Reconciler` 当前默认主路径仍通过 `workLoopSync()` 同步处理整棵树。

## 本目录文档

- [FIBER_FIRST_ARCHITECTURE.md](FIBER_FIRST_ARCHITECTURE.md): 当前 Fiber-first 架构和目标边界。
- [FIBER_FIRST_QUICK_REFERENCE.md](FIBER_FIRST_QUICK_REFERENCE.md): 术语、规则、文件位置和调试参考。

## 维护原则

- 修改 Fiber 或 render pipeline 后，优先更新本目录入口和快速参考。
- 不要把 `ComponentInstance` 写成 `paint.PaintableBox`。
- 不要把 lane metadata 描述成默认路径已具备抢占/时间切片，除非 `internal/reconciler` 主路径完成接入。
- 若描述目标设计，显式标注“目标”或“计划中”。
