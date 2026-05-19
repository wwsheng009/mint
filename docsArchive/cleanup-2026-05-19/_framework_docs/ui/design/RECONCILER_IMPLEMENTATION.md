# Reconciler 实施方案

## 📋 文档概述

**文档类型**: 实施方案
**目标**: 完整实现 Mint UI 的 Fiber Reconciler
**预计工期**: 4 周
**版本**: v1.0
**最后更新**: 2026-02-01

---

## 🎯 实施目标

### 当前问题

```go
// 当前：declarativeRoot.Paint() 直接渲染，无协调过程
func (d *declarativeRoot) Paint(ctx, buffer) {
    vnode := d.appFn()                    // 1. 获取 VNode
    d.renderVNode(vnode, ctx.X, ctx.Y, buffer) // 2. 直接渲染
    // 没有 reconcile，没有优先级，没有时间切片
}
```

### 目标架构

```go
// 目标：通过 Reconciler 协调渲染
func (d *declarativeRoot) Paint(ctx, buffer) {
    d.reconciler.Render(ctx, buffer)
}

// Reconciler 内部：
// 1. ScheduleUpdate(stateUpdate, lane)
// 2. WorkLoop(deadline) - 可中断
// 3. BeginWork/CompleteWork - 协调
// 4. CommitRoot() - 提交变更
```

---

## 📂 文件创建清单

### Phase 1: 基础协调器 (Week 1)

#### 1.1 `ui/reconciler.go` - 核心协调器

```go
package ui

import (
    "time"
    "github.com/wwsheng009/mint/framework"
    "github.com/wwsheng009/mint/runtime/paint"
)

// Reconciler 管理 Fiber 树和协调过程
type Reconciler struct {
    // === Fiber 树 ===
    root           *Fiber  // 当前树的根
    workInProgress *Fiber  // 工作中的树根

    // === 状态 ===
    lanes     Lane   // 待处理的 lanes
    isWorking bool   // 是否正在工作

    // === 调度 ===
    deadline    time.Time       // 当前帧的截止时间
    timeBudget  time.Duration   // 每帧的时间预算
    callback    func()          // 完成回调

    // === 集成 ===
    app        *framework.App  // 框架应用
    instanceMgr *InstanceManager // 实例管理器

    // === Buffer ===
    buffer *paint.Buffer       // 渲染目标
}

// ReconcilerConfig 配置协调器
type ReconcilerConfig struct {
    TimeBudget      time.Duration  // 时间切片预算
    EnableProfiling bool          // 启用性能分析
}

// NewReconciler 创建新的协调器
func NewReconciler(app *framework.App, config ReconcilerConfig) *Reconciler {
    return &Reconciler{
        app:           app,
        timeBudget:    config.TimeBudget,
        instanceMgr:   NewInstanceManager(),
    }
}

// === 调度更新 ===

// ScheduleUpdate 调度一次状态更新
func (r *Reconciler) ScheduleUpdate(lane Lane) {
    r.lanes = mergeLanes(r.lanes, lane)
    r.requestWork()
}

// requestWork 请求执行工作循环
func (r *Reconciler) requestWork() {
    if r.app != nil {
        r.app.MarkDirty()
    }
}

// === 工作循环 ===

// Render 执行渲染（公开 API）
func (r *Reconciler) Render(ctx component.PaintContext, buffer *paint.Buffer) {
    r.buffer = buffer

    // Phase 1: 简单同步渲染
    r.workLoopSync()
}

// workLoopSync 同步工作循环（Phase 1）
func (r *Reconciler) workLoopSync() {
    // 创建 VNode 树
    // ... 获取 vnode ...

    // 创建 Fiber 树
    r.root = CreateFiberFromVNode(vnode)

    // 处理所有工作
    r.workLoopSyncInternal()

    // 提交变更
    r.CommitRoot()
}

// workLoopSyncInternal 内部同步工作循环
func (r *Reconciler) workLoopSyncInternal() {
    workInProgress := r.createWorkInProgress(r.root)

    for workInProgress != nil {
        // BeginWork
        workInProgress = BeginWork(nil, workInProgress)

        // CompleteWork
        workInProgress = CompleteWork(nil, workInProgress)

        // 移动到下一个工作单元
        workInProgress = r.getNextWorkUnit(workInProgress)
    }

    r.workInProgress = nil
}

// === Commit ===

// CommitRoot 提交所有变更
func (r *Reconciler) CommitRoot() {
    if r.root == nil {
        return
    }

    // 渲染 Fiber 树到 buffer
    r.renderFiberToBuffer(r.root, r.buffer)
}

// renderFiberToBuffer 渲染 Fiber 树到 buffer
func (r *Reconciler) renderFiberToBuffer(fiber *Fiber, buffer *paint.Buffer) {
    if fiber == nil {
        return
    }

    // 根据 Fiber 类型渲染
    // ... 渲染逻辑 ...
}

// === 辅助方法 ===

// createWorkInProgress 创建工作副本
func (r *Reconciler) createWorkInProgress(root *Fiber) *Fiber {
    if root == nil {
        return nil
    }
    // Clone fiber for work
    return CloneFiber(root)
}

// getNextWorkUnit 获取下一个工作单元
func (r *Reconciler) getNextWorkUnit(current *Fiber) *Fiber {
    if current == nil {
        return nil
    }

    // 深度优先遍历：子节点 -> 兄弟节点
    if current.Child != nil {
        return current.Child
    }
    return current.Sibling
}

// hasMoreWork 检查是否还有工作
func (r *Reconciler) hasMoreWork() bool {
    return r.workInProgress != nil || r.lanes != LaneNoLane
}
```

#### 1.2 `ui/begin_work.go` - BeginWork 阶段

```go
package ui

// BeginWork 处理 Fiber 的开始阶段
// 这是 reconciliation 发生的地方
func BeginWork(current, workInProgress *Fiber) *Fiber {
    if workInProgress == nil {
        return nil
    }

    // 根据 Fiber 类型分发处理
    switch workInProgress.Type {
    case VNodeComponent:
        return beginWorkComponent(current, workInProgress)
    case VNodeText:
        return beginWorkText(current, workInProgress)
    case VNodeElement:
        return beginWorkElement(current, workInProgress)
    case VNodeFragment:
        return beginWorkFragment(current, workInProgress)
    case VNodeLayout:
        return beginWorkLayout(current, workInProgress)
    default:
        return workInProgress
    }
}

// beginWorkComponent 处理组件 Fiber
func beginWorkComponent(current, workInProgress *Fiber) *Fiber {
    componentVNode := workInProgress.VNode.(*ComponentVNode)

    // 处理更新队列
    processUpdateQueue(workInProgress)

    // 调用组件函数获取子节点
    var children []VNode
    if componentVNode.fn != nil {
        vnode := componentVNode.fn()
        children = []VNode{vnode}
    } else if componentVNode.fnWithProps != nil {
        vnode := componentVNode.fnWithProps(workInProgress.Props)
        children = []VNode{vnode}
    }

    // Reconcile 子节点
    workInProgress.Child = reconcileChildren(
        workInProgress,
        current != nil ? current.Child : nil,
        children,
        workInProgress.Lanes,
    )

    return workInProgress
}

// beginWorkText 处理文本 Fiber
func beginWorkText(current, workInProgress *Fiber) *Fiber {
    // 文本节点没有子节点，直接返回
    return workInProgress
}

// beginWorkElement 处理元素 Fiber
func beginWorkElement(current, workInProgress *Fiber) *Fiber {
    elementVNode := workInProgress.VNode.(*ElementVNode)
    children := elementVNode.Children()

    // Reconcile 子节点
    workInProgress.Child = reconcileChildren(
        workInProgress,
        current != nil ? current.Child : nil,
        children,
        workInProgress.Lanes,
    )

    return workInProgress
}

// beginWorkFragment 处理片段 Fiber
func beginWorkFragment(current, workInProgress *Fiber) *Fiber {
    fragmentVNode := workInProgress.VNode.(*FragmentVNode)
    children := fragmentVNode.Children()

    // Reconcile 子节点
    workInProgress.Child = reconcileChildren(
        workInProgress,
        current != nil ? current.Child : nil,
        children,
        workInProgress.Lanes,
    )

    return workInProgress
}

// beginWorkLayout 处理布局 Fiber
func beginWorkLayout(current, workInProgress *Fiber) *Fiber {
    layoutVNode := workInProgress.VNode.(*LayoutNode)
    children := layoutVNode.Children()

    // Reconcile 子节点
    workInProgress.Child = reconcileChildren(
        workInProgress,
        current != nil ? current.Child : nil,
        children,
        workInProgress.Lanes,
    )

    return workInProgress
}

// processUpdateQueue 处理更新队列
func processUpdateQueue(workInProgress *Fiber) {
    if workInProgress.UpdateQueue == nil {
        return
    }

    // 处理所有更新
    for update := workInProgress.UpdateQueue.First; update != nil; update = update.Next {
        // 应用更新
        if fn, ok := update.Payload.(func(interface{}) interface{}); ok {
            result := fn(workInProgress.MemoizedState)
            workInProgress.MemoizedState = result
        } else {
            workInProgress.MemoizedState = update.Payload
        }
    }

    // 清空队列
    workInProgress.UpdateQueue = nil
}
```

#### 1.3 `ui/complete_work.go` - CompleteWork 阶段

```go
package ui

// CompleteWork 处理 Fiber 的完成阶段
// 这里完成工作并收集 effects
func CompleteWork(current, workInProgress *Fiber) *Fiber {
    if workInProgress == nil {
        return nil
    }

    // 根据 Fiber 类型分发处理
    switch workInProgress.Type {
    case VNodeComponent:
        return completeWorkComponent(current, workInProgress)
    case VNodeText:
        return completeWorkText(current, workInProgress)
    case VNodeElement:
        return completeWorkElement(current, workInProgress)
    case VNodeLayout:
        return completeWorkLayout(current, workInProgress)
    default:
        return workInProgress
    }
}

// completeWorkComponent 完成组件工作
func completeWorkComponent(current, workInProgress *Fiber) *Fiber {
    // 标记 effect flags
    workInProgress.Flags |= EffectUpdate

    return workInProgress
}

// completeWorkText 完成文本工作
func completeWorkText(current, workInProgress *Fiber) *Fiber {
    return workInProgress
}

// completeWorkElement 完成元素工作
func completeWorkElement(current, workInProgress *Fiber) *Fiber {
    return workInProgress
}

// completeWorkLayout 完成布局工作
func completeWorkLayout(current, workInProgress *Fiber) *Fiber {
    return workInProgress
}
```

### Phase 2: Commit 阶段 (Week 2)

#### 2.1 `ui/commit.go` - Commit 阶段

```go
package ui

import "github.com/wwsheng009/mint/runtime/paint"

// CommitRoot 提交所有变更到 buffer
func (r *Reconciler) CommitRoot() {
    if r.root == nil {
        return
    }

    // 三阶段 commit
    r.commitBeforeMutationEffects(r.root)
    r.commitMutationEffects(r.root, r.buffer)
    r.commitLayoutEffects(r.root)
}

// commitBeforeMutationEffects 执行变更前 effects
func (r *Reconciler) commitBeforeMutationEffects(root *Fiber) {
    // 遍历 Fiber 树，执行 getSnapshotBeforeUpdate
    walkFiber(root, func(fiber *Fiber) bool {
        // 处理 snapshot effects
        return true
    })
}

// commitMutationEffects 执行变更 effects
func (r *Reconciler) commitMutationEffects(root *Fiber, buffer *paint.Buffer) {
    // 遍历 Fiber 树，渲染到 buffer
    r.renderFiberToBuffer(root, buffer)
}

// commitLayoutEffects 执行布局 effects
func (r *Reconciler) commitLayoutEffects(root *Fiber) {
    // 执行 useEffect 回调
    // 处理 effect cleanup
}

// renderFiberToBuffer 渲染 Fiber 树到 buffer
func (r *Reconciler) renderFiberToBuffer(fiber *Fiber, buffer *paint.Buffer) {
    if fiber == nil {
        return
    }

    // 深度优先渲染
    r.renderFiber(fiber, 0, 0, buffer)

    // 渲染兄弟节点
    for sibling := fiber.Sibling; sibling != nil; sibling = sibling.Sibling {
        r.renderFiberToBuffer(sibling, buffer)
    }
}

// renderFiber 渲染单个 Fiber
func (r *Reconciler) renderFiber(fiber *Fiber, x, y int, buffer *paint.Buffer) {
    // 根据 Fiber.VNode 类型渲染
    switch vnode := fiber.VNode.(type) {
    case *TextVNode:
        r.renderText(vnode, x, y, buffer)
    case *ButtonVNode:
        r.renderButton(vnode, x, y, buffer)
    case *InputVNode:
        r.renderInput(vnode, x, y, buffer)
    // ... 其他类型 ...
    }
}

// walkFiber 遍历 Fiber 树
func walkFiber(root *Fiber, callback func(*Fiber) bool) {
    if root == nil {
        return
    }

    if !callback(root) {
        return
    }

    walkFiber(root.Child, callback)
    walkFiber(root.Sibling, callback)
}
```

### Phase 3: 子节点协调 (Week 3-4)

#### 3.1 `ui/reconcile.go` - 子节点协调算法

```go
package ui

// reconcileChildren 协调子节点
func reconcileChildren(
    returnFiber *Fiber,
    currentFirstChild *Fiber,
    newChildren []VNode,
    lanes Lane,
) *Fiber {

    // Phase 1: 简单协调（按索引匹配）
    return reconcileChildrenArray(returnFiber, currentFirstChild, newChildren, lanes)
}

// reconcileChildrenArray 简单数组协调
func reconcileChildrenArray(
    returnFiber *Fiber,
    currentFirstChild *Fiber,
    newChildren []VNode,
    lanes Lane,
) *Fiber {

    var resultingFirstChild *Fiber
    var previousFiber *Fiber

    newChildrenLen := len(newChildren)

    for i := 0; i < newChildrenLen; i++ {
        var currentChild *Fiber

        // 找到对应的当前 child
        if currentFirstChild != nil && i == 0 {
            currentChild = currentFirstChild
        } else if previousFiber != nil && previousFiber.Sibling != nil {
            currentChild = previousFiber.Sibling
        }

        // 创建或更新 Fiber
        newFiber := updateSlot(
            returnFiber,
            currentChild,
            newChildren[i],
            lanes,
        )

        if i == 0 {
            resultingFirstChild = newFiber
        } else {
            previousFiber.Sibling = newFiber
        }

        previousFiber = newFiber
    }

    // 标记删除的节点
    markRemainingChildren(currentFirstChild, previousFiber)

    return resultingFirstChild
}

// updateSlot 更新单个槽位
func updateSlot(
    returnFiber *Fiber,
    oldFiber *Fiber,
    newVNode VNode,
    lanes Lane,
) *Fiber {

    // 情况 1: 新增
    if oldFiber == nil {
        return createFiberFromVNode(newVNode, returnFiber)
    }

    // 情况 2: 删除
    if newVNode == nil {
        return nil
    }

    // 情况 3: 类型不同 - 替换
    if oldFiber.Type != newVNode.Type() {
        return createFiberFromVNode(newVNode, returnFiber)
    }

    // 情况 4: Key 不同 - 替换
    if oldFiber.Key != newVNode.Key() {
        return createFiberFromVNode(newVNode, returnFiber)
    }

    // 情况 5: 更新
    return updateFiber(oldFiber, newVNode, lanes)
}

// createFiberFromVNode 从 VNode 创建 Fiber
func createFiberFromVNode(vnode VNode, returnFiber *Fiber) *Fiber {
    fiber := CreateFiber(vnode)
    fiber.Return = returnFiber
    return fiber
}

// updateFiber 更新现有 Fiber
func updateFiber(oldFiber *Fiber, newVNode VNode, lanes Lane) *Fiber {
    // Clone 并更新
    newFiber := CloneFiber(oldFiber)
    newFiber.VNode = newVNode
    newFiber.PendingProps = newVNode.Props()
    newFiber.Lanes = lanes
    return newFiber
}

// markRemainingChildren 标记剩余子节点为删除
func markRemainingChildren(currentFirstChild *Fiber, lastProcessed *Fiber) {
    // 找到第一个未处理的子节点
    var nextChild *Fiber
    if lastProcessed == nil {
        nextChild = currentFirstChild
    } else {
        nextChild = lastProcessed.Sibling
    }

    // 标记所有后续节点为删除
    for child := nextChild; child != nil; child = child.Sibling {
        child.Flags |= EffectDeletion
    }
}
```

---

## 🔧 修改现有文件

### `ui/app.go` 集成

```go
// declarativeRoot 包装声明式组件
type declarativeRoot struct {
    component.Node
    appFn               ComponentFunc
    ctx                 *ComponentContext
    app                 *framework.App
    instanceManager     *InstanceManager

    // === 新增：Reconciler ===
    reconciler          *Reconciler     // Fiber 协调器
    useFiber            bool            // 是否使用 Fiber
    activeComponentKeys []string
    // ... 其他字段 ...
}

// newDeclarativeRoot 创建新的声明式根组件
func newDeclarativeRoot(fn ComponentFunc, app *framework.App) component.Node {
    dr := &declarativeRoot{
        appFn:               fn,
        ctx:                 newComponentContext("App"),
        app:                 app,
        instanceManager:     NewInstanceManager(),
        activeComponentKeys: make([]string, 0),
        useFiber:            os.Getenv("MINT_USE_FIBER") == "true", // 环境变量控制
        // ... 其他字段初始化 ...
    }

    // 创建 Reconciler
    if dr.useFiber {
        dr.reconciler = NewReconciler(app, ReconcilerConfig{
            TimeBudget: 5 * time.Millisecond,
        })
    }

    // ... 其他初始化 ...
    return dr
}

// Paint 实现 component.Paintable
func (d *declarativeRoot) Paint(ctx component.PaintContext, buffer *paint.Buffer) {
    // === 使用 Fiber ===
    if d.useFiber && d.reconciler != nil {
        d.paintWithFiber(ctx, buffer)
        return
    }

    // === 传统渲染 ===
    d.paintLegacy(ctx, buffer)
}

// paintWithFiber 使用 Fiber 渲染
func (d *declarativeRoot) paintWithFiber(ctx component.PaintContext, buffer *paint.Buffer) {
    // 调用 Reconciler 渲染
    d.reconciler.Render(ctx, buffer)
}

// paintLegacy 传统渲染方式（保留作为后备）
func (d *declarativeRoot) paintLegacy(ctx component.PaintContext, buffer *paint.Buffer) {
    // 当前实现保持不变
    // ... 现有代码 ...
}
```

---

## 📊 实施进度

| 阶段 | 文件 | 状态 | 预计时间 |
|------|------|------|---------|
| Phase 1 | `reconciler.go` | ⏳ 待实施 | Day 1-2 |
| Phase 1 | `begin_work.go` | ⏳ 待实施 | Day 3-4 |
| Phase 1 | `complete_work.go` | ⏳ 待实施 | Day 5 |
| Phase 1 | `app.go` 集成 | ⏳ 待实施 | Day 5 |
| Phase 2 | `commit.go` | ⏳ 待实施 | Week 2 |
| Phase 2 | `effects.go` | ⏳ 待实施 | Week 2 |
| Phase 3 | `reconcile.go` | ⏳ 待实施 | Week 3 |

---

## ✅ 验收标准

### Phase 1 验收

- [ ] Reconciler 能够协调 VNode 树
- [ ] BeginWork 正确处理不同 VNode 类型
- [ ] CompleteWork 正确完成工作
- [ ] 能够渲染简单的组件树

### Phase 2 验收

- [ ] CommitRoot 正确输出到 buffer
- [ ] Effects 正确执行和清理
- [ ] 状态更新触发重新渲染

### Phase 3 验收

- [ ] Key 协调正常工作
- [ ] 组件列表增删正确处理
- [ ] 无内存泄漏

---

## 📚 参考资料

- React Fiber Architecture
- React Reconciliation Algorithm
- Mint UI Fiber 架构设计

---

**文档版本**: v1.0
**最后更新**: 2026-02-01
