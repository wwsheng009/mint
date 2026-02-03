// Package reconciler implements Fiber reconciliation for incremental UI updates.
//
// This package provides the declarative UI reconciliation system for mint.
//
// # Overview
//
// The reconciler manages a Fiber tree representation of the UI and performs
// incremental updates by comparing the current tree with a new tree and applying
// only the necessary changes.
//
// 协调器负责:
//   - 管理 VNode 树的创建和更新
//   - 执行 Diff 算法对比新旧树
//   - 管理 Fiber 节点的工作循环
//   - 调度组件的挂载、更新和卸载
//
// # Main Components
//
//   - Reconciler: Orchestrates the reconciliation process
//   - Fiber: Tree nodes representing UI elements
//   - TreeWalker: Utility for traversing Fiber trees
//   - Lane: Priority system for update scheduling
//
// # Lifecycle
//
// 1. ScheduleUpdate(lane) - Mark that an update is needed
// 2. Render() - Execute reconciliation and commit changes
// 3. CommitRoot() - Apply changes to the render buffer
//
// # Phase System
//
// Updates are scheduled with priority lanes:
//   - LaneSyncLane - Highest priority, synchronous updates
//   - LaneInputLane - User input handling
//   - LaneDefaultLane - Normal updates
//
// # Double Buffering
//
// The reconciler uses double buffering between current and workInProgress
// Fiber trees to ensure consistent state during reconciliation.
//
// # Usage
//
//	reconciler := NewReconciler(app, rootComponent, ReconcilerConfig{
//	    EnableFiber: true,
//	    TimeBudget: 5 * time.Millisecond,
//	})
//	reconciler.Render(ctx, buffer, renderFunc)
//
// 核心类型:
//   - Reconciler: 协调器主结构
//   - Fiber: 工作单元节点
//   - WorkLoop: 工作循环
//
// ⚠️ 此包是内部实现，不对外暴露。
package reconciler
