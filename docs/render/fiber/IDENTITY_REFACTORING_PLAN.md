# Mint UI Runtime Identity 重构方案

**创建日期**: 2026-02-13
**目标**: 彻底解决 VNode Key 同步问题，引入 Runtime NodeID 系统实现稳定的运行时标识
**优先级**: 高（架构级重构）

---

## 一、背景与现状

### 1.1 当前架构问题

现有的 TUI 系统依赖 Fiber 驱动 VNode.Key 的最终值，导致以下问题：

```
VNode.Key ← 被 Fiber.Path 覆盖
Fiber.Path ← 用于 Instance key
HitMap.NodeID ← 用 VNode.Key
```

这个设计存在三个核心问题：

**问题 1: VNode 不再是纯声明树**
- 原本 VNode.Key 应该只用于 sibling diff
- 现在变成 UI identity、HitMap identity、Instance identity
- 已经把 VNode 当成"运行时实体"在用

**问题 2: Fiber 成为全局 identity 源**
- cloneExistingFiber 必须同步 key
- createChildFiber 必须正确生成 path
- StripLayers 不能破坏 key
- 任何一处漏掉，系统就崩

**问题 3: Layer 剥离和 Fiber 树不一致**
- Fiber 树管 identity
- StripLayers 生成新 VNode
- Layout 再用这个 VNode
- 依赖的是 key 的"值相等"，而不是结构引用相等

### 1.2 具体症状

当前已发现的问题：

- Modal 中的按钮点击事件无法触发
- HitMap 中无法找到对应的 Component Instance
- 调试显示 Instance key 与 HitMap NodeID 不匹配
- `createChildFiberWithIndex` 的 `isRootChild` 条件过于严格，导致非根节点不能获得 layer-based path
- `cloneExistingFiber` 的 Path 同步条件 `strings.HasPrefix(current.Path, "/root/")` 过于严格

### 1.3 已完成的临时修复

已在 `RENDERLAYERS_VNODE_KEY_FIX.md` 中完成了两个阶段：

- **阶段 1**: 修复 `cloneExistingFiber` 条件，移除 `/root/` 前缀限制
- **阶段 2**: 修复 `isRootChild` 逻辑，为 layer 节点添加 layer-based path

这些修复解决了当前的症状，但**架构隐患依然存在**。

---

## 二、目标架构

### 2.1 最终目标

建立清晰的 Identity 分层模型：

```
VNode (声明层)
  └─ DiffKey (string)     ← 仅用于 sibling diff

Fiber (协调层)
  ├─ NodeID (uint64)      ← 唯一运行时 ID ⭐⭐⭐
  ├─ DiffKey (string)
  └─ Path (string, debug only)

ComputedBox (布局层)
  └─ NodeID (uint64)

HitMap (命中测试层)
  └─ NodeID (uint64)

InstanceRegistry (实例层)
  └─ map[NodeID]Instance
```

**核心原则**：

> 运行时身份只认 NodeID，不认 key，不认 path。

### 2.2 数据流向

```
VNode (声明)
  ↓ diff
Fiber (生成 RuntimeNodeID)
  ↓ layout
ComputedBox (保存 RuntimeNodeID)
  ↓ hittest
HitMap (用 RuntimeNodeID)
  ↓ lookup
InstanceMap[RuntimeNodeID]
```

VNode 不参与 runtime identity。

### 2.3 优势

| 优势 | 说明 |
|------|------|
| Layer 剥离完全不影响 ID | StripLayers 再怎么 clone VNode，都不会影响 Fiber.NodeID |
| Path 变化不会影响 ID | reorder、insert、animate、portal、overlay 都不会破坏 ID |
| 不需要 cloneExistingFiber 同步 key | 整个 key 同步问题消失 |
| 不需要 isRootChild 检查 | Path 只用于调试，不用于 identity |
| 支持高级特性 | Portal、Multi Root、Suspense、动画系统、离屏缓存等 |

---

## 三、NodeID 设计

### 3.1 类型选择

```go
type NodeID uint64
```

选择理由：
- 比 string 快
- 比 path 稳定
- 比 hash 可控
- 足够大

### 3.2 分配策略

纯 runtime allocator（推荐）：

```go
type IDAllocator struct {
    next uint64
}

func (a *IDAllocator) Next() NodeID {
    a.next++
    return NodeID(a.next)
}
```

在 Fiber mount 时分配：
```go
fiber.NodeID = idAllocator.Next()
```

在 Fiber clone 时保留：
```go
fiber.NodeID = current.NodeID
```

**不要**：
- ❌ hash
- ❌ 用 path
- ❌ 结构生成
- ❌ 字符串拼接

---

## 四、需要修改的模块

按系统层次列出所有需要修改的文件和接口。

### 4.1 Fiber 层

**文件**：
- `internal/reconciler/diff.go`
- `internal/reconciler/fiber.go`
- `internal/reconciler/fiber_test.go`

**修改内容**：

```go
type Fiber struct {
    // 新增：唯一运行时 ID
    NodeID  NodeID

    // 现有字段
    DiffKey string
    Path    string   // 仅用于调试
    Key     string

    // ... 其他字段
}
```

**在 createChildFiber 时**：

```go
fiber.NodeID = idAllocator.Next()
```

**在 cloneExistingFiber 时**：

```go
fiber.NodeID = current.NodeID  // 保留
```

**删除**：

```go
// ❌ 移除这行
vnode.SetKey(fiber.Path)
```

### 4.2 Instance Registry

**文件**：
- `framework/app.go`
- `internal/runtime/instance_registry.go`
- `tests/instance_test.go`

**修改内容**：

```go
type InstanceRegistry struct {
    // 新增：NodeID 索引
    instancesByID map[NodeID]*Instance

    // 旧版：key 索引（迁移期间保留）
    instancesByKey map[string]*Instance
}
```

**Instance 结构**：

```go
type Instance struct {
    NodeID NodeID

    // ... 其他字段
}
```

**注册接口**：

```go
// 新接口
func (r *InstanceRegistry) RegisterInstanceByID(nodeID NodeID, instance *Instance)

// 旧接口（迁移期间保留）
func (r *InstanceRegistry) RegisterInstanceByKey(key string, instance *Instance)
```

**查找接口**：

```go
// 新接口
func (r *InstanceRegistry) GetInstanceByID(nodeID NodeID) *Instance

// 旧接口（迁移期间保留）
func (r *InstanceRegistry) GetInstanceByKey(key string) *Instance
```

### 4.3 Layout Engine

**文件**：
- `internal/layout/computed_box.go`
- `runtime/layout_engine.go`

**修改 ComputedBox**：

```go
type ComputedBox struct {
    Rect   Rect
    VNode  rtui.VNode

    // 新增：关联的 Fiber NodeID
    NodeID NodeID
}
```

**布局时的逻辑**：

```go
func (e *LayoutEngine) Layout(vnode rtui.VNode, constraints Constraints) *ComputedBox {
    box := &ComputedBox{
        VNode:  vnode,
        NodeID: currentFiber.NodeID,  // ← 从 Fiber 获取
    }

    // 计算布局...
    return box
}
```

**删除**：

```go
// ❌ 不要再读取 VNode.Key 作为 ID
box.NodeID = box.VNode.Key()
```

### 4.4 HitMap

**文件**：
- `internal/reconciler/reconciler.go`
- `internal/render/hitmap.go`
- `internal/render/hitmap_test.go`

**修改 HitEntry**：

```go
type HitEntry struct {
    Rect   Rect
    NodeID NodeID

    // 旧字段（迁移期间保留）
    VNodeKey string
}
```

**HitTest 接口**：

```go
// 新接口
func (h *HitMap) HitTest(x, y int) NodeID

// 旧接口（迁移期间保留）
func (h *HitMap) HitTestByKey(x, y int) string
```

**构建 HitMap**：

```go
func buildHitMapFromComputedBox(box *ComputedBox, hitMap *HitMap) {
    // 新逻辑：使用 NodeID
    hitMap.Add(box.Rect, box.NodeID)

    // 旧逻辑（迁移期间保留）
    if box.VNode != nil && box.VNode.Key() != "" {
        hitMap.AddByKey(box.Rect, box.VNode.Key())
    }
}
```

### 4.5 Event Routing

**文件**：
- `framework/app.go`
- `internal/runtime/event_dispatcher.go`

**修改事件路由逻辑**：

```go
func (a *App) handleEvent(x, y int) {
    // 新逻辑：使用 NodeID
    nodeID := a.hitMap.HitTest(x, y)
    instance := a.instanceRegistry.GetInstanceByID(nodeID)

    // Fallback：旧逻辑（迁移期间保留）
    if instance == nil {
        key := a.hitMap.HitTestByKey(x, y)
        instanceKey := "vnode:" + key
        instance = a.instanceRegistry.GetInstanceByKey(instanceKey)
    }

    if instance != nil {
        instance.DispatchEvent(event)
    }
}
```

### 4.6 Layer System

**文件**：
- `runtime/layer/collector.go`
- `runtime/layer/manager.go`

**修改 LayerNode 结构**：

```go
type LayerNode struct {
    Layer   rtui.Layer
    ID      string
    Content rtui.VNode
    Visible bool
    FocusID string

    // 新增：关联的 Fiber NodeID
    NodeID NodeID
}
```

**walk 方法**：

```go
func (c *Collector) walk(vnode rtui.VNode, fiber *reconciler.Fiber) {
    if layer := vnode.GetLayer(); layer != rtui.LayerBase && layer.IsValid() {
        node := &LayerNode{
            Layer:   layer,
            ID:      c.getNodeID(vnode),
            Content: vnode,
            Visible: c.isVisible(vnode),
            FocusID: "",
            // 新增：保存 Fiber NodeID
            NodeID:  fiber.NodeID,
        }
        c.layers.Add(layer, node)
        return
    }
    // ...
}
```

### 4.7 Inspector / DevTools

**文件**：
- `internal/render/inspector.go` (如果有)

**修改策略**：
- 保留 Path 字段，但只用于显示和调试
- 不要用于查找或 identity
- 使用 NodeID 进行组件查找

```go
// 显示时
fmt.Printf("Component: %s (NodeID: %d, Path: %s)", instance.Type, instance.NodeID, fiber.Path)

// 查找时
func (i *Inspector) FindComponent(nodeID NodeID) *Instance {
    return i.instanceRegistry.GetInstanceByID(nodeID)
}
```

---

## 五、无痛渐进迁移路径

不能一次性全改，否则会炸。我们做 3 阶段迁移。

### 5.1 阶段 1：引入 NodeID（不删除旧逻辑）⭐ 最重要

**目标**：引入新的 NodeID 系统，与旧系统并行运行

**步骤**：

1. **新增 IDAllocator**

   文件：`internal/reconciler/id_allocator.go`

   ```go
   package reconciler

   type NodeID uint64

   type IDAllocator struct {
       next uint64
   }

   func (a *IDAllocator) Next() NodeID {
       a.next++
       return NodeID(a.next)
   }
   ```
   ```bash
   go test ./internal/reconciler -run TestIDAllocator
   ```
2. **Fiber 增加 NodeID**

   文件：`internal/reconciler/fiber.go`

   ```go
   type Fiber struct {
       NodeID  NodeID   // 新增
       DiffKey string
       Path    string
       Key     string
       // ... 其他字段
   }
   ```

   文件：`internal/reconciler/diff.go`

   ```go
   // 在 createChildFiber 时
   fiber.NodeID = idAllocator.Next()

   // 在 cloneExistingFiber 时
   fiber.NodeID = current.NodeID
   ```

3. **ComputedBox 增加 NodeID**

   文件：`internal/layout/computed_box.go`

   ```go
   type ComputedBox struct {
       Rect   Rect
       VNode  rtui.VNode
       NodeID NodeID  // 新增
   }
   ```

   文件：`runtime/layout_engine.go`

   ```go
   func (e *LayoutEngine) Layout(vnode rtui.VNode, fiber *reconciler.Fiber, constraints Constraints) *ComputedBox {
       box := &ComputedBox{
           VNode:  vnode,
           NodeID: fiber.NodeID,  // ← 从 Fiber 获取
       }
       // ... 计算布局
       return box
   }
   ```

4. **HitMap 同时存两种 ID**

   文件：`internal/render/hitmap.go`

   ```go
   type HitEntry struct {
       Rect    Rect
       NodeID  NodeID      // 新增
       VNodeKey string     // 保留旧字段
   }

   func (h *HitMap) Add(rect Rect, nodeID NodeID, vNodeKey string) {
       entry := HitEntry{
           Rect:     rect,
           NodeID:   nodeID,
           VNodeKey: vNodeKey,
       }
       h.entries = append(h.entries, entry)
   }

   func (h *HitMap) HitTest(x, y int) (NodeID, string) {
       // 新接口：同时返回 NodeID 和 VNodeKey
       for _, entry := range h.entries {
           if entry.Rect.Contains(x, y) {
               return entry.NodeID, entry.VNodeKey
           }
       }
       return 0, ""
   }
   ```

   文件：`internal/reconciler/reconciler.go`

   ```go
   func buildHitMapFromComputedBox(box *ComputedBox, hitMap *HitMap) {
       hitMap.Add(box.Rect, box.NodeID, box.VNode.Key())
   }
   ```

5. **InstanceRegistry 同时支持两种索引**

   文件：`internal/runtime/instance_registry.go`

   ```go
   type InstanceRegistry struct {
       instancesByID  map[NodeID]*Instance   // 新增
       instancesByKey map[string]*Instance  // 保留
   }

   func (r *InstanceRegistry) RegisterInstance(instance *Instance, key string, nodeID NodeID) {
       r.instancesByID[nodeID] = instance
       if key != "" {
           r.instancesByKey[key] = instance
       }
   }

   func (r *InstanceRegistry) GetInstance(nodeID NodeID, key string) *Instance {
       // 优先用 NodeID
       if instance, ok := r.instancesByID[nodeID]; ok {
           return instance
       }
       // Fallback 到 key
       return r.instancesByKey[key]
   }
   ```

6. **Event Routing 优先用 NodeID**

   文件：`framework/app.go`

   ```go
   nodeID, key := a.hitMap.HitTest(x, y)
   instance := a.instanceRegistry.GetInstance(nodeID, "vnode:"+key)
   ```

**验证**：
```bash
# 运行所有测试
go test ./...

# 验证 Modal、Overlay、Tooltip 等场景
go test ./tests/modal_test.go -v
go test ./tests/overlay_test.go -v
```

**预期结果**：
- ✅ 系统仍然能正常运行
- ✅ 所有现有功能不受影响
- ✅ 新的 NodeID 系统开始工作

### 5.2 阶段 2：双轨验证

**目标**：验证 NodeID 与旧 key 系统的一致性

**步骤**：

1. **添加 Assertion**

   文件：`internal/runtime/instance_registry.go`

   ```go
   func (r *InstanceRegistry) RegisterInstance(instance *Instance, key string, nodeID NodeID) {
       r.instancesByID[nodeID] = instance
       if key != "" {
           r.instancesByKey[key] = instance
       }

       // Debug: 验证一致性
       if key != "" {
           if existingKeyInstance, ok := r.instancesByKey[key]; ok && existingKeyInstance != instance {
               log.Warnf("Key %s maps to multiple instances", key)
           }
       }
   }
   ```

   文件：`framework/app.go`

   ```go
   nodeID, key := a.hitMap.HitTest(x, y)
   instanceByNodeID := a.instanceRegistry.GetInstanceByID(nodeID)
   instanceByKey := a.instanceRegistry.GetInstanceByKey("vnode:" + key)

   // Debug: 验证一致性
       if instanceByNodeID != instanceByKey {
           log.Warnf("Identity mismatch: NodeID=%d vs Key=%s", nodeID, "vnode:"+key)
       }
   }
   ```

2. **添加调试日志**

   文件：`internal/reconciler/reconciler.go`

   ```go
   func buildHitMapFromComputedBox(box *ComputedBox, hitMap *HitMap) {
       log.Debugf("[HitMap] NodeID=%d, VNode.Key=%q", box.NodeID, box.VNode.Key())
       hitMap.Add(box.Rect, box.NodeID, box.VNode.Key())
   }
   ```

3. **运行完整测试套件**

   ```bash
   # 启用调试日志
   export TUI_DEBUG_HITMAP=true

   # 运行所有测试
   go test ./... -v

   # 特别测试这些场景
   go test ./tests/modal_test.go -run TestModalButtonClick -v
   go test ./tests/overlay_test.go -v
   go test ./tests/reorder_test.go -v
   go test ./tests/dynamic_test.go -v
   ```

4. **收集不一致性报告**

   记录所有 mismatch 警告，分析原因：
   如果不一致，检查：
   - Fiber.NodeID 是否正确分配
   - ComputedBox.NodeID 是否正确传递
   - HitMap 是否同时存储了两个 ID
   - InstanceRegistry 是否正确注册

**预期结果**：
- ✅ 没有 identity mismatch 警告
- ✅ NodeID 和 key 始终指向同一个 Instance
- ✅ 所有测试通过

### 5.3 阶段 3：删除旧 Identity

**目标**：完全切换到 NodeID 系统

**步骤**：

1. **删除 VNode.Key 同步逻辑**

   文件：`internal/reconciler/diff.go`

   ```go
   // ❌ 删除这些行
   if current.Path != "" {
       vnode.SetKey(current.Path)
   } else if current.Key != "" {
       vnode.SetKey(current.Key)
   }
   ```

2. **删除 isRootChild 逻辑**

   文件：`internal/reconciler/diff.go`

   ```go
   // ❌ 删除或保留仅用于调试
   // isRootChild := returnFiber != nil && returnFiber.Key == "root" && returnFiber.Path == "/root"
   // if isRootChild {
   //     typePath = pathGenerator.generateRootPath(vnode)
   // }
   ```

   Path 只用于调试，不再参与 identity。

3. **删除旧的 key 索引**

   文件：`internal/runtime/instance_registry.go`

   ```go
   type InstanceRegistry struct {
       instancesByID map[NodeID]*Instance
       // ❌ 删除
       // instancesByKey map[string]*Instance
   }

   // ❌ 删除旧接口
   // func (r *InstanceRegistry) RegisterInstanceByKey(...)
   // func (r *InstanceRegistry) GetInstanceByKey(...)
   ```

4. **简化 HitMap**

   文件：`internal/render/hitmap.go`

   ```go
   type HitEntry struct {
       Rect   Rect
       NodeID NodeID
       // ❌ 删除
       // VNodeKey string
   }

   func (h *HitMap) Add(rect Rect, nodeID NodeID) {
       entry := HitEntry{
           Rect:   rect,
           NodeID: nodeID,
       }
       h.entries = append(h.entries, entry)
   }

   func (h *HitMap) HitTest(x, y int) NodeID {
       for _, entry := range h.entries {
           if entry.Rect.Contains(x, y) {
               return entry.NodeID
           }
       }
       return 0
   }
   ```

5. **简化 Event Routing**

   文件：`framework/app.go`

   ```go
   func (a *App) handleEvent(x, y int) {
       nodeID := a.hitMap.HitTest(x, y)
       instance := a.instanceRegistry.GetInstanceByID(nodeID)

       if instance != nil {
           instance.DispatchEvent(event)
       }
   }
   ```

6. **更新所有测试**

   将所有依赖 key 的测试更新为使用 NodeID。

7. **清理无用的代码**

   删除所有 `vnode.SetKey(fiber.Path)` 调用
   删除所有 `"vnode:" + key` 拼接
   删除所有从 `box.VNode.Key()` 读取 ID 的逻辑

8. **保留 Path 仅用于调试**

   文件：`internal/reconciler/diff.go`

   ```go
   // Path 仅保留用于调试和日志
   fiber.Path = typePath
   ```

**验证**：
```bash
# 运行完整测试套件
go test ./... -v

# 运行所有集成测试
go test ./tests/... -v

# 性能测试
go test ./bench/... -bench=.
```

**预期结果**：
- ✅ 所有测试通过
- ✅ 系统完全基于 NodeID 运行
- ✅ 不再有任何 key 同步逻辑
- ✅ Path 仅用于调试和日志

---

## 六、迁移影响面评估

| 模块 | 阶段 1 风险 | 阶段 1 复杂度 | 阶段 2 风险 | 阶段 2 复杂度 | 阶段 3 风险 | 阶段 3 复杂度 |
|------|-------------|---------------|-------------|---------------|-------------|---------------|
| Fiber | 低 | 低 | 低 | 低 | 低 | 低 |
| Layout | 低 | 低 | 低 | 低 | 低 | 低 |
| HitMap | 中 | 中 | 低 | 低 | 中 | 中 |
| Event | 中 | 中 | 低 | 低 | 中 | 中 |
| Layer | 中 | 中 | 低 | 低 | 中 | 中 |
| Layer | 中 | 中 | 低 | 低 | 中 | 中 |
| DevTools | 低 | 低 | 低 | 低 | 低 | 低 |

**总体风险**：可控

**总体复杂度**：中等

**关键路径**：
1. Fiber → Layout → HitMap → Event
2. 确保 NodeID 在整个链路中正确传递

---

## 七、测试策略

### 7.1 单元测试

**文件**：`internal/reconciler/id_allocator_test.go`

```go
func TestIDAllocator(t *testing.T) {
    allocator := &IDAllocator{}

    id1 := allocator.Next()
    id2 := allocator.Next()

    assert.Equal(t, NodeID(1), id1)
    assert.Equal(t, NodeID(2), id2)
    assert.True(t, id1 < id2)
}
```

**文件**：`internal/reconciler/fiber_test.go`

```go
func TestFiberNodeID(t *testing.T) {
    allocator := &IDAllocator{}

    // 创建 Fiber 时分配 NodeID
    fiber := createChildFiberWithIndex(nil, nil, vnode, 0, allocator)
    assert.NotEqual(t, NodeID(0), fiber.NodeID)

    // 克隆时保留 NodeID
    clone := cloneExistingFiber(nil, fiber, newNode, 0)
    assert.Equal(t, fiber.NodeID, clone.NodeID)
}
```

### 7.2 集成测试

**文件**：`tests/modal_node_id_test.go`

```go
func TestModalButtonNodeID(t *testing.T) {
    app := ui.NewTestApp()

    var buttonNodeID NodeID
    var modalNodeID NodeID

    root := func() rtui.VNode {
        children := []rtui.VNode{
            rtui.NewElement("button").
                SetKey("open-modal-btn").
                SetLabel("Open Modal").
                OnClick(func() {
                    // ...
                }),
        }

        // 添加 modal
        children = append(children,
            rtui.NewElement("modal").
                SetKey("test-modal").
                SetLayer(rtui.LayerModal).
                SetChildren(
                    rtui.NewElement("text").SetText("Modal Content"),
                    rtui.NewElement("button").
                        SetKey("modal-btn").
                        SetLabel("Click Me").
                        OnClick(func() {
                            // ...
                        }),
                ),
        )

        return rtui.NewElement("vstack").SetChildren(children...)
    }

    ui.Run(root)

    // 验证 NodeID 正确分配
    assert.NotEqual(t, NodeID(0), modalNodeID)
    assert.NotEqual(t, NodeID(0), buttonNodeID)
    assert.NotEqual(t, modalNodeID, buttonNodeID)

    // 验证 HitMap 使用 NodeID
    instance := app.instanceRegistry.GetInstanceByID(buttonNodeID)
    assert.NotNil(t, instance)
}
```

**文件**：`tests/event_routing_node_id_test.go`

```go
func TestEventRoutingWithNodeID(t *testing.T) {
    app := ui.NewTestApp()

    var clicked bool

    root := rtui.NewElement("button").
        SetLabel("Click Me").
        OnClick(func() {
            clicked = true
        })

    ui.Run(root)

    // 触发点击事件
    simulateClick(10, 10)

    // 验证事件被触发
    assert.True(t, clicked)

    // 验证使用 NodeID 查找实例
    logContains(t, "NodeID hit test successful")
}
```

### 7.3 性能测试

**文件**：`bench/node_id_bench_test.go`

```go
func BenchmarkNodeIDLookup(b *testing.B) {
    registry := &InstanceRegistry{
        instancesByID: make(map[NodeID]*Instance),
        // instancesByKey: make(map[string]*Instance),  // 性能对比
    }

    // 注册 10000 个实例
    for i := 0; i < 10000; i++ {
        instance := &Instance{NodeID: NodeID(i)}
        registry.RegisterInstanceByID(NodeID(i), instance)
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        registry.GetInstanceByID(NodeID(i % 10000))
    }
}

func BenchmarkKeyLookup(b *testing.B) {
    registry := &InstanceRegistry{
        instancesByKey: make(map[string]*Instance),
    }

    // 注册 10000 个实例
    for i := 0; i < 10000; i++ {
        instance := &Instance{}
        key := fmt.Sprintf("vnode:%d", i)
        registry.RegisterInstanceByKey(key, instance)
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        key := fmt.Sprintf("vnode:%d", i%10000)
        registry.GetInstanceByKey(key)
    }
}
```

预期结果：NodeID 查找比 key 查找快 10 倍以上。

### 7.4 回归测试

在迁移过程中，每个阶段完成后都运行完整的回归测试：

```bash
# 运行所有测试
go test ./... -v

# 检查覆盖率
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# 竞争检测
go test ./... -race
```

---

## 八、迁移完成后的最终结构

```
VNode (声明层)
  └─ DiffKey (string)     ← 仅用于 sibling diff
        ↓
Fiber (协调层)
  ├─ NodeID (uint64)      ← 唯一运行时 ID ⭐
  ├─ DiffKey (string)
  └─ Path (string, debug only)
        ↓
Layout & ComputedBox (布局层)
  └─ NodeID (uint64)
        ↓
HitMap (命中测试层)
  └─ NodeID (uint64)
        ↓
InstanceRegistry (实例层)
  └─ map[NodeID]Instance
        ↓
Event Dispatch (事件分发)
```

这是一套真正 UI Runtime 的 identity 模型。

---

## 九、风险与缓解措施

### 9.1 潜在风险

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| NodeID 分配冲突 | 中 | 低 | 使用 increment 分配器，保证全局唯一 |
| ComputedBox.NodeID 未正确传递 | 高 | 中 | 添加 assertion 和日志 |
| 测试覆盖不足 | 高 | 中 | 编写集成测试和性能测试 |
| 性能回归 | 低 | 低 | 基准测试对比 |
| 向后兼容性问题 | 中 | 低 | 三阶段迁移，旧系统逐步下线 |

### 9.2 回滚方案

如果在阶段 1 或阶段 2 发现严重问题：

1. 立即停止迁移
2. 保留旧 key 系统继续运行
3. 分析问题原因
4. 修复后重新开始阶段 1

如果在阶段 3 发现问题：

1. 回退到阶段 2（双轨运行）
2. 继续使用旧系统，同时修复新系统
3. 修复后再次尝试阶段 3

---

## 十、迁移后的收益

### 10.1 架构收益

| 收益 | 说明 |
|------|------|
| Layer 重排不再影响 identity | StripLayers 不再破坏 key |
| Path 改变不影响事件 | Path 只用于调试，不用于 identity |
| Reorder 不破坏 Instance | NodeID 始终稳定 |
| 不需要同步 key | 删除所有 vnode.SetKey(fiber.Path) |
| 不会再出现 key mismatch bug | NodeID 永远唯一 |
| 更清晰的分层 | 声明层、协调层、布局层、运行时层职责分明 |

### 10.2 性能收益

- NodeID (uint64) 查找比 string 快 10 倍以上
- 减少字符串拼接和内存分配
- 减少 hash 计算

### 10.3 可扩展性

迁移完成后，可以轻松实现：
- Portal
- Multi Root
- Suspense
- 动画系统
- 离屏缓存
- 虚拟列表
- DevTools Time Travel
- 热重载

---

## 十一、未来扩展能力

引入 NodeID 后，以下复杂特性将变得可行：

### 11.1 Portal

```go
func Portal(target NodeID, children rtui.VNode) rtui.VNode {
    // children 渲染到 target 指定的位置
    // NodeID 保持不变
}
```

### 11.2 Off-screen Cache

```go
type OffscreenCache struct {
    boxes map[NodeID]*ComputedBox
}
```

### 11.3 Virtual List

```go
type VirtualList struct {
    itemIDs []NodeID  // 稳定的 item ID
}
```

### 11.4 Time Travel Debugging

```go
type StateSnapshot struct {
    fiberTree map[NodeID]*Fiber
}
```

---

## 十二、执行计划

### 12.1 时间估算

| 阶段 | 任务 | 工作量（人天） |
|------|------|----------------|
| 阶段 1 | 引入 NodeID 系统 | 2-3 天 |
| 阶段 1 | 编写单元测试 | 1-2 天 |
| 阶段 2 | 双轨验证 | 1-2 天 |
| 阶段 3 | 删除旧 Identity | 2-3 天 |
| 阶段 3 | 更新测试 | 1-2 天 |
| 测试和调试 | 2-3 天 |

**总计**: 约 9-15 天

### 12.2 里程碑

- **里程碑 1**: NodeID 系统引入完成，所有测试通过
- **里程碑 2**: 双轨验证无 warning
- **里程碑 3**: 完全切换到 NodeID，性能提升

---

## 十三、附录

### 13.1 相关文档

- [Layer System Architecture](../layout/LAYER_SYSTEM_ARCHITECTURE.md)
- [Event Routing Fix Summary](../issue/event_refactor/BUTTON_EVENT_ROUTING_FIX_SUMMARY.md)
- [Fiber Reconciliation Guide](../plan/FIBER_RECONCILER_MIGRATION.md)
- [Mixed Key Strategy Implementation](../issue/event_refactor/MIXED_KEY_STRATEGY_IMPLEMENTATION.md)
- [RENDERLAYERS_VNODE_KEY_FIX.md](./RENDERLAYERS_VNODE_KEY_FIX.md)
- [FIBER_FIX.md](./FIBER_FIX.md)
- [FIBER_ID.md](./FIBER_ID.md)

### 13.2 调试命令

```bash
# 启用调试日志
export TUI_DEBUG_HITMAP=true
export TUI_DEBUG_RENDER=true
export TUI_DEBUG_UI=true

# 运行测试
go test -v ./tests/modal_button_click_test.go

# 检查性能
go test -bench=. -benchmem ./bench/

# 检查覆盖率
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 13.3 验证清单

#### 阶段 1 完成时：

- [ ] Fiber 增加 NodeID 字段
- [ ] IDAllocator 实现并测试通过
- [ ] ComputedBox 增加 NodeID 字段
- [ ] HitMap 同时存 NodeID 和 VNodeKey
- [ ] InstanceRegistry 同时支持两种索引
- [ ] Event Routing 优先用 NodeID，有 fallback
- [ ] 所有现有测试通过
- [ ] Modal、Overlay、Tooltip 场景测试通过

#### 阶段 2 完成时：

- [ ] 添加 identity mismatch assertion
- [ ] 添加调试日志
- [ ] 运行完整测试套件
- [ ] 没有不一致性警告
- [ ] 收集并分析所有日志

#### 阶段 3 完成时：

- [ ] 删除 VNode.Key 同步逻辑
- [ ] 删除 isRootChild identity 逻辑
- [ ] 删除旧的 key 索引
- [ ] 简化 HitMap 为只用 NodeID
- [ ] 简化 Event Routing
- [ ] 更新所有测试
- [ ] 清理无用代码
- [ ] Path 仅用于调试
- [ ] 所有测试通过
- [ ] 性能测试通过
- [ ] 无 regression

---

## 十四、结论

### 14.1 当前情况

- ✅ 已完成临时修复（阶段 1 + 阶段 2 of RENDERLAYERS_VNODE_KEY_FIX.md）
- ✅ 解决了当前的症状（Modal 点击不工作）
- ⚠️ 但架构隐患依然存在

### 14.2 最优路径

> 现在就抽离 identity，实施 NodeID 系统

### 14.3 分阶段执行

1. **立即开始阶段 1**: 引入 NodeID 系统（不删除旧逻辑）
2. **验证后进入阶段 2**: 双轨验证（确保一致性）
3. **验证后进入阶段 3**: 删除旧 Identity（完全切换）

### 14.4 最终价值

这套 TUI 已经是一个"准框架"级别的 UI Runtime。

通过这次重构：
- ✅ 建立清晰的分层模型
- ✅ 实现稳定的运行时身份
- ✅ 支持高级 UI 特性
- ✅ 提升性能和可维护性

---

**文档完成**: 2026-02-13
**状态**: ✅ 方案完整，等待评审
**优先级**: 高（架构级重构）
**预计工期**: 9-15 天
**风险等级**: 中等（已缓解）
