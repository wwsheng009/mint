# 架构复杂性分析：为什么需要 Component Registry？

## 问题背景

用户质疑：**为什么需要在 app 层面维护 Component Registry？这太复杂了！**

## 当前架构的复杂性

### 数据流

```
1. render() → 创建 VNode 树（每次重建）
   ↓
2. 从 VNode 树构建 HitMap (ID → Bounds)
   ↓
3. 从 VNode 树构建 Registry (ID → Updater)
   ↓
4. 用户点击 → HitTest 找到 TargetID
   ↓
5. 用 TargetID 查 Registry 找组件
   ↓
6. 调用组件.Update(msg)
```

### 核心矛盾

#### 1. HitMapEntry 已经有 Node 引用

```go
type HitMapEntry struct {
    NodeID string
    Node   layout.Node  // ← 已经有引用！
    Bounds layout.Rect
}
```

但是不能直接调用：
```go
entry.Node.Update(mouseMsg)  // ❌ 编译错误
```

原因：**layout.Node 接口没有事件处理方法**

```go
type Node interface {
    ID() string
    Type() string
    Children() []Node
    GetPosition/SetPosition
    GetSize/SetSize
    // ❌ 没有 HandleEvent 或 Update 方法
}
```

#### 2. VNode 每次渲染都重建

```go
// 每次 render() 都执行
func (a *App) Render() rtui.VNode {
    return app.ButtonBuilder("Click").
        OnClick(func() { /*...*/ }).
        Build()  // ← 创建新的 ButtonVNode 实例
}
```

**问题**：
- VNode 是临时描述对象（类似 React 的 Virtual DOM）
- onClick 存储在 VNode 上
- 用户点击时，VNode 已经被替换

#### 3. 时序问题

```
Frame 1:
  render() → VNode1{onClick: handler1}
  buildHitMap(VNode1) → HitMap["button"] = VNode1
  buildRegistry(VNode1) → Registry["button"] = VNode1

Frame 2 (用户还没点击):
  render() → VNode2{onClick: handler2}  // ← 新的 VNode
  buildHitMap(VNode2) → HitMap["button"] = VNode2
  buildRegistry(VNode2) → Registry["button"] = VNode2
  // VNode1 被丢弃！

Frame 3 (用户点击):
  HitTest → TargetID = "button"
  Registry.Lookup("button") → VNode2  // ← 找到当前帧的 VNode
  VNode2.Update(mouseMsg)
  VNode2.onClick()  // ← 调用当前帧的 handler
```

**关键洞察**：
- Registry 每帧都重建
- 总是返回**当前帧**的 VNode
- onClick 是当前帧的闭包，状态是最新的
- 所以即使 VNode 是临时的，也能工作！

## 为什么这么复杂？

### React 的解决方案

React 有两层架构：

```
JSX (临时描述) → Fiber (持久实例) → DOM
```

- Fiber 节点是持久的
- 事件处理存储在 Fiber 上
- Fiber 有稳定的状态

### 我们的架构

```
VNode (临时描述) → ??? → Terminal
```

- **没有持久的组件实例层**
- VNode 既是描述也是"实例"
- 每次都重建

### 为什么不用 Fiber？

1. **Fiber 很复杂**：需要 diff 算法、调度器、reconciliation
2. **我们的场景简单**：TUI 应用没有复杂的 DOM 操作
3. **性能够用**：终端渲染本身就很慢，不需要极致优化

## 根本问题

**VNode 职责过重**：
1. 作为渲染描述
2. 存储事件处理器
3. 维护组件状态

**导致**：
- 无法直接引用（每次都变）
- 需要 Registry 做间接映射
- 架构复杂

## 解决方案

### 方案 1：保留 Component Registry（当前）

**优点**：
- 不需要改动架构
- Registry 总是返回当前帧的 VNode
- onClick 闭包捕获最新状态

**缺点**：
- 每帧重建 Registry（开销）
- 需要维护两套映射（HitMap + Registry）
- 架构复杂

### 方案 2：引入持久实例层（类似 Fiber）

```go
type ComponentInstance struct {
    ID       string
    VNode    *VNode  // 当前帧的 VNode
    State    interface{}
    OnClick  func()
}

// 事件处理
instance := InstanceMap.Lookup(targetID)
instance.OnClick()  // 持久的处理器
```

**优点**：
- 实例持久，可直接引用
- 状态管理清晰
- 不需要每帧重建

**缺点**：
- 需要实现 diff 算法
- 架构更复杂
- 可能过度设计

### 方案 3：HitMap 直接存储 Updater

```go
type HitMapEntry struct {
    NodeID string
    Node   layout.Node
    Bounds layout.Rect
    Updater component.Updater  // ← 直接存储 Updater
}

// 事件处理
entry := hitMap.HitTest(x, y)
if entry != nil && entry.Updater != nil {
    entry.Updater.Update(mouseMsg)
}
```

**优点**：
- 不需要额外的 Registry
- 减少一层间接映射
- 架构更简洁

**缺点**：
- HitMap 变得更重
- 每次重建 HitMap 都要提取 Updater

## 结论

当前架构虽然复杂，但有其合理性：
1. **VNode 每帧重建是 React-like 的本质**
2. **Registry 桥接了 HitMap 和 Updater**
3. **onClick 闭包确保状态最新**

但确实可以简化：
- 考虑方案 3（HitMap 直接存储 Updater）
- 减少一层间接映射

## 下一步

1. 评估方案 3 的可行性
2. 测试性能差异
3. 如果更简洁且性能可接受，考虑重构
