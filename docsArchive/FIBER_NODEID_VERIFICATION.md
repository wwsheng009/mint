# Fiber NodeID 稳定性验证方案

## 问题诊断

### 当前实现的问题

`App.fiberRoot` 字段虽然被添加，但**没有被真正使用**：

```go
// framework/app.go - 错误的实现
func (a *App) render() {
    if provider, ok := a.root.(interface{ GetFiberRoot() *rtui.Fiber }); ok {
        currentFiberRoot := provider.GetFiberRoot()

        // ❌ 获取了 fiberRoot，但没有传递给 Reconciler
        // ❌ 保存到 a.fiberRoot，但没有真正复用

        if a.fiberRoot == nil && currentFiberRoot != nil {
            a.fiberRoot = currentFiberRoot
        } else if a.fiberRoot != nil && currentFiberRoot != nil {
            a.fiberRoot = currentFiberRoot  // 只是指针赋值
        }
    }

    // 调用 Paint()，Reconciler 仍然创建新 Fiber 树
    paintable.Paint(ctx, buf)
}
```

### 实际渲染流程

```
App.render()
  ↓
DeclarativeNode.Paint()                    // declarative_node.go:237
  ↓
renderWithFiberContext()                 // declarative_node.go:336
  ↓
n.reconciler.Render()                     // reconciler.go:104
  ↓
  prepareFreshStack()          ← 创建新 VNode 树
  workLoopSync()                ← 创建新 Fiber 树！
  CommitRoot()                 ← 提交新 Fiber 树
```

**关键发现**：`Reconciler.Render()` 每次都创建新的 Fiber 树，不使用 `App.fiberRoot`。

## 正确的验证方式

### 方式 1：运行时日志验证（推荐）

设置环境变量启用详细日志：

```bash
export TUI_DEBUG_UI=true
export TUI_DEBUG_HITMAP=true
go run ./examples/hello
```

观察日志中的 NodeID：

```log
[App.render] Render 1
[Engine] [buildComputedBoxWithSize] Set NodeID=1 from Fiber (type=VNodeElement)
[Engine] [buildComputedBoxWithSize] Set NodeID=2 from Fiber (type=VNodeElement)
[Engine] [buildComputedBoxWithSize] Set NodeID=3 from Fiber (type=VNodeElement)

[App.render] Render 2  ← 第二次渲染
[Engine] [buildComputedBoxWithSize] Set NodeID=4 from Fiber (type=VNodeElement)  ← ⚠️ 不同！
[Engine] [buildComputedBoxWithSize] Set NodeID=5 from Fiber (type=VNodeElement)
[Engine] [buildComputedBoxWithSize] Set NodeID=6 from Fiber (type=VNodeElement)
```

**如果 NodeID 保持不变**（1,2,3 → 1,2,3），说明 Fiber 树被正确复用。

### 方式 2：集成测试验证

创建集成测试，模拟真实渲染流程：

```go
// framework/app_fiber_integration_test.go
func TestFiberNodeIDStability(t *testing.T) {
    // 记录每次渲染的 NodeID
    var renderNodeIDs [][]uint64

    for i := 0; i < 3; i++ {
        // 模拟触发渲染（实际通过 event 或 tick）
        app.TriggerRender()

        // 收集当前树中所有 Fiber NodeID
        nodeIDs := collectFiberNodeIDs(app.fiberRoot)
        renderNodeIDs = append(renderNodeIDs, nodeIDs)

        t.Logf("Render %d: NodeIDs = %v", i+1, nodeIDs)
    }

    // 验证：NodeID 应该保持稳定
    if reflect.DeepEqual(renderNodeIDs[0], renderNodeIDs[1]) &&
       reflect.DeepEqual(renderNodeIDs[1], renderNodeIDs[2]) {
        t.Log("✅ NodeID 保持稳定")
    } else {
        t.Error("❌ NodeID 不稳定")
    }
}

func collectFiberNodeIDs(fiber *rtui.Fiber) []uint64 {
    var ids []uint64
    walkFiber(fiber, func(f *rtui.Fiber) {
        ids = append(ids, f.NodeID)
    })
    return ids
}
```

### 方式 3：Reconciler 内部验证（最准确）

直接检查 `Reconciler.root` 字段：

```go
// internal/reconciler/reconciler.go
type Reconciler struct {
    root           *Fiber // ← 这是实际的 Fiber 树根
    workInProgress *Fiber
}

// 在 reconciler_test.go 中
func TestNodeIDStability(t *testing.T) {
    reconciler := NewReconciler(app, rootComponent, config)

    // 第一次渲染
    reconciler.Render(ctx, buf, renderFunc)
    root1 := reconciler.root

    // 记录 NodeID
    nodeIDs1 := collectFiberNodeIDs(root1)

    // 第二次渲染（模拟）
    reconciler.Render(ctx, buf, renderFunc)
    root2 := reconciler.root

    // 记录 NodeID
    nodeIDs2 := collectFiberNodeIDs(root2)

    // 验证
    if reflect.DeepEqual(nodeIDs1, nodeIDs2) {
        t.Log("✅ NodeID 保持稳定")
    } else {
        t.Logf("❌ NodeID 变化：\n第一次: %v\n第二次: %v", nodeIDs1, nodeIDs2)
    }
}
```

## 关键要点

### 1. App.fiberRoot 字段的作用

**当前状态**：
- ✅ 字段存在
- ✅ 可以赋值
- ❌ 没有被 Reconciler 使用

**实际意义**：
- `App.fiberRoot` 只是**缓存** Reconciler 的 Fiber 树引用
- 真正的复用逻辑在 `Reconciler` 内部

### 2. Fiber 树复用机制

在 `Reconciler.workLoopSync()` 中 (reconciler.go:176-206)：

```go
func (r *Reconciler) workLoopSync() {
    // 设置全局 reconciler
    currentReconciler = r

    // 处理工作单元（创建/更新 Fiber 树）
    r.performUnitOfWork(r.workInProgress)

    // ✨ 关键：交换双缓冲
    // workInProgress 成为新的 root
    r.root = r.workInProgress  // ← 实际在这里更新
    r.workInProgress = nil
}
```

**双缓冲机制**：
- `r.root`：当前提交的 Fiber 树
- `r.workInProgress`：工作树（每帧重建）
- 交换操作使得 `r.root` 可以复用（通过 Alternate）

### 3. NodeID 分配机制

在 `CreateFiber()` 或 `cloneExistingFiber()` 时分配：

```go
func createFiber(...) *Fiber {
    // 分配全局唯一 NodeID
    nodeID := atomic.AddUint64(&fiberNodeIDCounter, 1)

    return &Fiber{
        NodeID: nodeID,  // ← 新节点
        ...
    }
}
```

**关键问题**：
- 如果总是调用 `createFiber()`，NodeID 会递增
- 如果调用 `cloneExistingFiber()`，NodeID 保持不变

### 4. reconcileChildren 的作用

在 `begin_work.go:175-180` 中：

```go
func beginWorkElement(...) *Fiber {
    workInProgress.Child = reconcileChildren(
        workInProgress,      // 父 Fiber
        currentChild,        // 之前的子 Fiber 树
        newChildren,         // 新的 VNode 子节点
        workInProgress.Lanes,
    )
}
```

**`reconcileChildren()` 负责**：
- 对比 `currentChild` 和 `newChildren`
- 决定是创建新 Fiber 还是复用现有 Fiber
- 决定权在 `beginWorkElement` 中，不在 `reconcileChildren`

## 当前代码问题总结

### 问题 1：App.fiberRoot 未使用 ❌

```go
// framework/app.go
type App struct {
    fiberRoot *rtui.Fiber  // ← 存在但未使用
}

// 实际上，Reconciler 管理自己的 Fiber 树
// App.fiberRoot 只是可选的缓存/观察者
```

### 问题 2：NodeID 不稳定的真正原因 ⚠️

**根本原因**：每次渲染调用 `CreateFiberFromVNode()` 创建新树

在 `renderWithFiberContext()` 中 (declarative_node.go:336-358)：

```go
func (n *DeclarativeNode) renderWithFiberContext() rtui.VNode {
    n.reconciler.Render(ctx, buf, func() rtui.VNode {
        vnode := n.renderFn()  // ← 每次创建新 VNode 树
        return vnode
    })

    // Reconciler 内部调用 CreateFiberFromVNode()
    // 每次创建新 Fiber 树 → NodeID 递增
}
```

### 正确的理解：Fiber 树复用是 Reconciler 内部机制

**无需在 App 层面实现复用**：

1. ✅ `Reconciler` 已经有双缓冲机制
2. ✅ `reconcileChildren()` 已经实现复用逻辑
3. ✅ `createOrReuseFiber()` 应该存在

**验证应该聚焦于 Reconciler 的内部逻辑**。

## 建议的修复方向

### 方向 1：检查 Reconciler 是否正确复用 Fiber

```go
// internal/reconciler/reconciler.go
func (r *Reconciler) Render(...) {
    // 检查：是否复用现有 Fiber 树
    if r.root != nil {
        t.Logf("Existing Fiber root: NodeID=%d", r.root.NodeID)
    }

    r.prepareFreshStack(renderFunc)
    r.workLoopSync()

    // 检查：Root NodeID 是否保持
    if r.root != nil {
        t.Logf("After render: Fiber root NodeID=%d", r.root.NodeID)
    }
}
```

### 方向 2：确保 createOrReuseFiber 逻辑存在

```go
// 应该有这样的逻辑：
func createOrReuseFiber(returnFiber *Fiber, vnode rtui.VNode) *Fiber {
    // 如果 returnFiber != nil，尝试复用
    if returnFiber != nil && shouldUpdate(returnFiber, vnode) {
        return cloneExistingFiber(returnFiber, vnode)
    }

    // 否则创建新 Fiber
    return createFiber(vnode)
}
```

### 方向 3：验证实例测试

```bash
# 运行现有的 Fiber 测试
go test -v -run TestReconcileChildren ./internal/reconciler/...
go test -v -run TestFiberSync ./internal/reconciler/...

# 观察是否通过
```

## 结论

1. ❌ **App.fiberRoot 字段对 NodeID 稳定性没有直接帮助**
   - 它只是缓存的引用
   - 真正的复用逻辑在 Reconciler 内部

2. ✅ **应该关注 Reconciler 的内部逻辑**
   - `reconcileChildren()` 是否正确实现？
   - `createOrReuseFiber()` 是否存在？
   - 双缓冲机制是否生效？

3. ✅ **正确的验证方式**
   - 集成测试（模拟多次渲染）
   - 日志输出（观察 NodeID 变化）
   - 直接检查 `Reconciler.root` 字段

需要我：
1. 检查 Reconciler 的 createOrReuseFiber 逻辑？
2. 运行现有的 reconcileChildren 测试？
3. 创建集成测试验证实际渲染行为？
