# Fiber Architecture

**状态**: 部分实现，默认渲染路径已 Fiber-first；可中断渲染和时间切片尚未接入主 reconciler。

## Overview

Mint uses Fiber as the persistent structure behind declarative UI. The public declaration format is still VNode, but the default `ui.Run` path reconciles VNode into Fiber, then performs layout and paint from Fiber and component instances.

Current implemented path:

```text
ComponentFunc / VNode
  -> internal/reconciler.Reconciler
  -> Fiber tree
  -> ComponentInstance persistence
  -> internal/render Fiber layout adapter
  -> runtime/layout
  -> paint.PaintableBox / PaintablePlanes
  -> PaintEngine
```

## Implemented

- Fiber tree with parent/child/sibling/alternate links.
- NodeID, DiffKey and path metadata.
- Component instance creation and reuse.
- BeginWork / CompleteWork traversal.
- Synchronous render work loop.
- Commit bookkeeping for deletions, portals, focus, renderer fiber sync and hook effects.
- Fiber-first layout adapter.
- Portal-aware layout support.
- Fiber-to-paintable conversion.
- PaintEngine rendering of paintable planes.
- FiberFocusManager integration.

## Not Yet Implemented In The Default Reconciler

- Interruptible render.
- Time slicing.
- Lane-based preemption.
- Resumable work loop.

Lane values exist as metadata, but `internal/reconciler.Reconciler` currently runs through `workLoopSync()`.

## Fiber Node

The actual Fiber definition lives in `runtime/ui/fiber.go`. Important fields include:

- Tree links: `Return`, `Child`, `Sibling`, `Alternate`.
- Declaration data: `Props`, `MemoizedProps`, `MemoizedState`, `Style`, `Tag`.
- Identity: `Key`, `DiffKey`, `Path`, `PathSegment`, `NodeID`, `ActionTargetID`.
- Scheduling metadata: `Flags`, `SubtreeFlags`, `Lanes`, `ChildLanes`.
- Layout inputs: width/height/flex/position/margin/padding/border-related fields.
- Layer and portal metadata.
- Runtime entity: `Instance ComponentInstance`.

## Render Phase

The current main render phase is synchronous:

```text
Reconciler.Render()
  -> prepareFreshStack()
  -> workLoopSync()
  -> performUnitOfWork()
       -> BeginWork()
       -> process child subtree
       -> CompleteWork()
       -> collectChildEffects()
       -> process sibling subtree
```

`prepareFreshStack()` wraps the root render function as a root component and initializes the work-in-progress tree.

## Commit Phase

Current `CommitRoot()` performs runtime bookkeeping:

1. `commitDeletions(root)`
2. `linkPortalsToRoots(root)`
3. `applyFocusStateToFiber(root)`
4. `renderer.SetFiber(root)` when supported
5. `updateFocusManagerFromFiber(root)`
6. `ctx.FinishRender()`
7. `ctx.RunEffects()`
8. `cleanupUnusedInstances()`

It does not directly perform terminal painting. Layout and paint are executed by `DeclarativeNode.fiberFirstPaint()` after reconciliation.

## Diff Identity

Current child reconciliation uses `DiffKey`:

- User key from `vnode.Key()`.
- Fallback key `_idx_<siblingIndex>`.
- `Fiber.Path` is for debug/tooling and does not drive reconciliation.
- Dynamic lists should provide stable keys.

## ComponentInstance

`Fiber.Instance` points to a `runtime/ui.ComponentInstance`. It is not `paint.PaintableBox`.

Important instance capabilities:

- `ComponentInstance`: lifecycle, props, key and dirty flags.
- `PaintableInstance`: optional paint capability.
- `FocusableInstance`: optional focus capability.
- Additional action/tick/box-model/scene capabilities depending on component.

Paint stage creates `paint.PaintableBox` nodes from layout results and Fiber references. `PaintableBox.Node` usually wraps Fiber through `FiberPaintableNode`, which delegates to `Fiber.Instance` if it implements `PaintableInstance`.

## Scheduling Note

`runtime/ui/fiber_scheduler.go` contains a separate Fiber scheduler abstraction. The default reconciler in `internal/reconciler` does not currently use it for preemption or time slicing.

## Time Travel

Time travel debugging belongs to the Store/Runtime debugging layer, not to the Fiber reconciler itself. See `runtime/debug` and Store/Reducer docs for state history usage.

## Related Documentation

- [Fiber-first consolidated architecture](../../fiber/fiber_first/consolidated/FIBER_FIRST_ARCHITECTURE.md)
- [Fiber-first quick reference](../../fiber/fiber_first/consolidated/FIBER_FIRST_QUICK_REFERENCE.md)
- [Store/Reducer guide](../../ui/store/guides/README.md)
- [Render docs](../../render/README.md)
