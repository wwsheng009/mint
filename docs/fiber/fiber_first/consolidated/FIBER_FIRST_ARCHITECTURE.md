# Fiber-First 统一架构文档

> 核心原则：VNode 管声明，Fiber 管结构，ComponentInstance 管运行时状态和行为。

## 当前状态

Fiber-first 是 `ui.Run` 的默认声明式渲染路径。当前已经实现 Fiber tree、实例持久化、Fiber-first layout/paint、Portal-aware layout 和 Fiber focus；但主 reconciler 仍是同步工作循环，尚未在默认路径中实现可中断渲染、时间切片或 lane-based preemption。

## 架构概览

```text
┌────────────────────────────────────────────┐
│ VNode declaration                           │
│ - user-facing declarative description       │
│ - recreated by ComponentFunc on render      │
└───────────────────┬────────────────────────┘
                    │ reconcile
                    v
┌────────────────────────────────────────────┐
│ Fiber tree                                  │
│ - persistent structure                      │
│ - DiffKey / NodeID / layer / props / style  │
│ - points to persistent ComponentInstance    │
└───────────────────┬────────────────────────┘
                    │ layout adapter
                    v
┌────────────────────────────────────────────┐
│ runtime/layout                              │
│ - FiberToNodeAdapter                        │
│ - optional PortalAwareLayoutEngine          │
│ - produces LayoutBox tree                   │
└───────────────────┬────────────────────────┘
                    │ convert
                    v
┌────────────────────────────────────────────┐
│ Paint tree                                  │
│ - paint.PaintableBox                        │
│ - paint.PaintablePlanes                     │
│ - FiberPaintableNode delegates to Instance  │
└───────────────────┬────────────────────────┘
                    │ paint
                    v
┌────────────────────────────────────────────┐
│ paint.Buffer / SceneFrame / terminal output │
└────────────────────────────────────────────┘
```

## Core Roles

| Role | Source | Responsibility |
|---|---|---|
| VNode | `runtime/ui`, `ui` | Declarative input for the next UI tree. |
| Fiber | `runtime/ui/fiber.go` | Persistent structure, identity, props/style/layer/layout inputs, lanes metadata and `Instance`. |
| ComponentInstance | `runtime/ui/instance.go` | Persistent component state, lifecycle, optional paint/focus/action capabilities. |
| LayoutBox | `runtime/layout` | Computed layout tree. |
| PaintableBox | `runtime/paint` | Paint-stage node containing bounds, layer, zIndex and `PaintableNode`. |
| FiberPaintableNode | `internal/render/converter.go` | Paintable wrapper that delegates to `Fiber.Instance` when possible. |

## Reconciler Flow

The default reconciler flow is synchronous:

```text
Reconciler.Render()
  -> prepareFreshStack()
  -> workLoopSync()
  -> performUnitOfWork()
       -> BeginWork()
       -> child subtree
       -> CompleteWork()
       -> collectChildEffects()
       -> sibling subtree
  -> CommitRoot()
```

`CommitRoot()` currently performs runtime bookkeeping:

1. Delete cleanup.
2. Portal linking.
3. Apply focus state to Fiber instances.
4. Call `renderer.SetFiber(root)` when available.
5. Update FocusManager from Fiber.
6. Finish render context.
7. Run hook effects.
8. Clean unused instances.

Layout and paint happen after reconciliation in `DeclarativeNode.fiberFirstPaint()`.

## Diff And Identity

Children reconciliation uses `DiffKey`:

- User key from `vnode.Key()` wins.
- Otherwise the fallback key is `_idx_<siblingIndex>`.
- Dynamic lists are expected to provide stable user keys.
- `Path` / `PathSegment` are debug and tooling metadata, not diff identity.

`CloneFiber()` preserves `NodeID`, `DiffKey`, layout inputs and `ActionTargetID`, and reuses `Instance`.

## Layout

Fiber-first layout uses `internal/render.FiberToNodeAdapter` to expose Fiber as `runtime/layout.Node`.

Important details:

- Layout reads Fiber props/style/layout fields.
- Position and size are computed in `LayoutBox`; `SetPosition` / `SetSize` primarily synchronize instance bounds.
- Border and box model information can come from instance capabilities or Fiber props.
- Portal-aware layout uses two phases: main tree first, portal children second.

## Paint

Paint flow:

```text
LayoutBox tree
  -> FiberToPaintableConverter
  -> paint.PaintableBox tree
  -> PaintablePlanes
  -> PaintEngine.PaintPaintablePlanes()
```

Paint priority:

1. `Fiber.Instance` implements `PaintableInstance`.
2. PaintRegistry fallback by `Fiber.Tag`.
3. No output.

## Focus

Focus is Fiber-first:

- FocusManager collects focusable Fiber nodes.
- Focus trap can be scoped by active layer.
- Focus state is applied to `ComponentInstance` through instance focus APIs.
- `FocusableVNode` and `FocusableMeta` should be treated as compatibility or historical concepts.

## Event Status

Current implementation is mixed:

- Main app loop processes `runtime/msg.Msg` through Action/InputProcessor/ActionBridge paths.
- Mouse messages can carry `TargetFiber`, `TargetID`, `TargetBounds` and local coordinates from HitMap.
- `DeclarativeNode.HandleEvent()` still contains a legacy VNode distribution fallback.

Target direction:

- Event routing should use HitMap/TargetFiber and ActionBridge.
- Runtime event dispatch should avoid VNode traversal.

## Scheduling Status

Lane values and update queue metadata exist. The default `internal/reconciler.Reconciler` still uses `workLoopSync()`.

Not currently active in the default reconciler:

- interruptible render
- time slicing
- lane preemption
- resumable work loop

`runtime/ui/fiber_scheduler.go` is separate from the main reconciler path.

## Migration Rules

For component authors:

- Do not store mutable runtime state on VNode.
- Use `InstanceFactory` to create a persistent `ComponentInstance`.
- Implement `PaintableInstance` for custom painting.
- Implement focus capabilities when the component participates in keyboard focus.
- Use stable keys for dynamic lists.
- Prefer Intent/Action routing for semantic interactions.

## Success Criteria

Practical success criterion:

> After reconciliation, the default layout and paint path runs from Fiber and ComponentInstance without traversing VNode. VNode remains the declaration format for producing the next Fiber tree.

## Related Files

- `runtime/ui/fiber.go`
- `runtime/ui/fiber_util.go`
- `runtime/ui/instance.go`
- `internal/reconciler/reconciler.go`
- `internal/reconciler/diff.go`
- `internal/render/declarative_node.go`
- `internal/render/fiber_adapter.go`
- `internal/render/converter.go`
- `internal/render/paint_engine.go`
- `internal/render/portal_layout_adapter.go`
