# 事件系统重构计划

## 目标

根据 `fix1.md` 设计文档，利用现有 Fiber 系统，彻底重构事件系统。

## 架构现状

**已有**：
- ✅ Fiber 系统 (`runtime/ui/fiber.go`)
- ✅ ComponentInstance 接口 (`runtime/ui/instance.go`)
- ✅ VNode → Fiber reconcile 逻辑
- ✅ 持久化状态管理

**问题**：
- ❌ HitMap 只存储 NodeID，不存储 Fiber/Instance
- ❌ Component Registry 每帧重建
- ❌ 事件流程：HitMap → ID → Registry → VNode (临时)

## 重构目标架构

```
事件流程（新）：
HitMap → LayoutNode → Fiber/Instance → Handler
        (含 Fiber/  (持久化实例)
        Instance    (存储状态+事件)
        引用)
```

## 重构步骤

### Phase 1: 修改 HitMap 存储结构

**文件**: `runtime/event/hitmap.go`

**修改**：
```go
type HitMapEntry struct {
    NodeID   string
    Node     layout.Node
    Bounds   layout.Rect
    LocalXY  func(screenX, screenY int) (localX, localY int)
    ZOrder   int

    // ✨ 新增：直接引用 Fiber/Instance
    Fiber    *Fiber  // 或 Instance
}
```

### Phase 2: 构建 HitMap 时填充 Fiber 引用

**文件**: `runtime/event/hitmap.go`

**修改 `BuildHitMap`**：
```go
func BuildHitMapFromFiberTree(root *Fiber) *HitMap {
    // 遍历 Fiber 树，构建 HitMap
    // 每个 entry 保存对应的 Fiber 引用
}
```

### Phase 3: 修改事件分发逻辑

**文件**: `framework/event/pump.go`

**当前**：
```go
entry := hitMap.HitTest(x, y)
mouseMsg.TargetID = entry.NodeID  // 只传 ID
```

**改为**：
```go
entry := hitMap.HitTest(x, y)
mouseMsg.Target = entry.Fiber  // 直接传 Fiber/Instance
```

**文件**: `framework/app.go`

**当前**：
```go
component := a.componentReg.Lookup(mouseMsg.TargetID)
component.Update(mouseMsg)
```

**改为**：
```go
if mouseMsg.Target != nil {
    mouseMsg.Target.Handle(mouseMsg)  // 直接调用
}
```

### Phase 4: 删除 Component Registry

**文件**: `framework/app.go`

**删除**：
- `componentReg *component.Registry`
- `buildComponentRegistry()` 方法
- 所有 Registry 相关代码

### Phase 5: 修改 MouseMsg 结构

**文件**: `runtime/msg/mouse.go`

**当前**：
```go
type MouseMsg struct {
    TargetID string
    ...
}
```

**改为**：
```go
type MouseMsg struct {
    Target interface{}  // Fiber 或 Instance
    // 或更具体的类型：
    TargetInstance ComponentInstance
    TargetFiber    *Fiber
    ...
}
```

### Phase 6: 清理旧代码

**删除**：
1. `framework/component/registry.go` - 不再需要
2. `framework/app.go` 中的 Registry 相关逻辑
3. 每帧重建 Registry 的代码

**保留**：
- Fiber 系统（核心）
- Instance 系统（核心）
- HitMap（修改后）

## 预期效果

**性能**：
- ❌ 每帧重建 Registry
- ✅ HitMap 每帧重建（轻量）
- ✅ 事件分发 O(1)

**架构清晰度**：
- ❌ VNode (临时) + Registry (每帧重建)
- ✅ Fiber/Instance (持久) + HitMap (缓存)

**代码复杂度**：
- 减少约 200 行 Registry 相关代码
- 事件分发路径缩短 50%

## 风险评估

**低风险**：
- Fiber 系统已经成熟
- Instance 机制已实现
- 只是改变引用方式

**需要注意**：
- HitMap 构建时机（在 Fiber reconcile 之后）
- Fiber 生命周期管理
- 内存泄漏检查（Fiber 引用）

## 实施顺序

1. ✅ 分析现有 Fiber 系统
2. ⬜ 修改 HitMapEntry 添加 Fiber 引用
3. ⬜ 修改 HitMap 构建逻辑
4. ⬜ 修改 MouseMsg 结构
5. ⬜ 修改 Pump 事件分发
6. ⬜ 修改 App.handleMsg
7. ⬜ 删除 Registry 代码
8. ⬜ 测试所有按钮
9. ⬜ 性能对比

## 验证标准

**功能验证**：
- ✅ 按钮点击正常
- ✅ Modal 按钮正常
- ✅ Inspector 正常
- ✅ 列表渲染正常

**性能验证**：
- ✅ 无 Registry 重建开销
- ✅ 事件分发耗时 < 1μs
- ✅ 内存无泄漏

**架构验证**：
- ✅ VNode 不存储事件处理器
- ✅ Instance 持久化状态
- ✅ HitMap 直接引用 Instance
