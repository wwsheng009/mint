# Fiber Counter 按钮重复问题详细分析报告

## 问题概述

在 `examples/fiber_counter` 示例中，当 HStack 包含 `[Button, Text(" "), Button]` 时，第二个按钮会重复显示。添加第二个空格文本后问题消失。

**受影响的代码**:
```go
ui.HStack(
    app.ButtonBuilder(" - ").OnPress(intent.Decrement("count", 1)).Build(),
    ui.Text(" "),  // 单个空格
    app.ButtonBuilder(" + ").OnPress(intent.Increment("count", 1)).Build(),
)
```

**临时解决方案**:
```go
ui.HStack(
    app.ButtonBuilder(" - ").OnPress(intent.Decrement("count", 1)).Build(),
    ui.Text(" "),  // 第一个空格
    ui.Text(" "),  // 额外空格 - 临时解决方案
    app.ButtonBuilder(" + ").OnPress(intent.Increment("count", 1)).Build(),
)
```

## 调查过程

### 1. Fiber Tree 结构测试

通过 `TestFullReconcile_HStackButtons` 测试，确认 fiber tree 结构正确：

```
[hstack, NodeID=1]
  ├─ [button, NodeID=5, DiffKey=_idx_0]  // "-" 按钮
  ├─ [text, NodeID=6, DiffKey=_idx_1]    // 空格文本
  └─ [button, NodeID=7, DiffKey=_idx_2]  // "+" 按钮
```

**结论**: Fiber tree 中没有重复的 NodeID，reconcile 层面没有问题。

### 2. Reconciliation 逻辑验证

检查了以下关键函数：

1. **shouldUpdate**: 修复后使用 `newSiblingIndex` 而不是 `current.SiblingIndex`
2. **CloneFiber**: 修复后清空 `Child` 和 `Sibling` 指针
3. **findMatchingChild**: 基于正确的 index 进行匹配

**结论**: Reconciliation 逻辑正确。

### 3. ComputedBox 结构测试

通过 `TestFullReconcile_HStackButtons` 验证了 computed box 中的子节点也是正确的。

**结论**: ComputedBox 层面没有问题。

### 4. HitMap 测试

创建了 `TestBuildHitMap_HStackButtons` 测试，验证了 HitMap 中没有重复的 NodeID。

**结论**: HitMap 层面没有问题。

## 问题特征分析

### 关键线索

1. **奇数 vs 偶数子节点**:
   - 3个子节点 `[Button, Text, Button]` → 按钮重复
   - 4个子节点 `[Button, Text, Text, Button]` → 正常
   
2. **相同类型节点间隔**:
   - 问题场景: Button(0) → Text(1) → Button(2)
   - Button(2) 可能与 Button(0) 产生了某种关联

3. **HStack 特定**:
   - 问题只在 HStack 中出现
   - VStack 没有类似问题

### 可能的根本原因

基于调查，问题可能出现在以下几个方面：

#### 假设 1: Gap 计算问题

HStack 中使用 gap 来分隔子元素。当有奇数个子节点时，最后一个按钮可能被错误定位。

**代码位置**: `runtime/compute/fiber_only_layout.go`

```go
// 布局计算代码
for _, child := range box.Children {
    // 计算子节点位置
    e.calculateFiberOnlyPositions(child, childX, childY)
    childX += child.Box.Width + gap  // 问题可能在这里
}
```

#### 假设 2: Hit Testing 边界计算

按钮的边界可能有重叠，导致同一个按钮在不同位置被击中。

**代码位置**: `runtime/ui/fiber_util.go` - `BuildHitMapFromFiber`

如果有两个按钮的边界重叠，点击测试可能同时命中两个。

#### 假设 3: 渲染缓存问题

渲染引擎可能缓存了之前的按钮状态，导致在重新渲染时出现了"残影"。

#### 假设 4: Z-order 问题

在 HitMap 中，z-order 计算依赖于 `Layer * 10000 + treeDepth`。如果 treeDepth 计算有误，可能导致按钮被错误排序。

```go
// BuildHitMapFromFiber 中的 z-order 计算
depth := 0
for p := fiber.Return; p != nil; p = p.Return {
    depth++
}
zOrder := int(layer)*10000 + depth
```

### 为什么添加第二个空格能解决问题？

1. **改变索引分布**:
   - 4个子节点: Button(0) → Text(1) → Text(2) → Button(3)
   - Button(3) 与 Button(0) 的差异更明显

2. **改变布局计算**:
   - 额外的 gap 改变了 childX 的累积值
   - 按钮的最终位置可能避免了重叠

3. **改变 HitMap 条目数量**:
   - 多一个条目改变了 HitMap 的排序
   - 可能避免了某种边界情况

## 未覆盖的测试区域

由于无法直接运行和观察 UI，以下区域需要进一步调查：

1. **实际渲染输出**:
   - 需要查看实际屏幕上的按钮位置
   - 验证是否有边界重叠

2. **Hit Testing 行为**:
   - 点击两个按钮会触发什么？
   - 是否有一个按钮实际上不可点击？

3. **边界框计算**:
   - 检查 ComputedBox 的实际大小和位置
   - 验证 gap 是否被正确应用

## 建议的调查步骤

### 步骤 1: 添加详细的布局日志

在 `runtime/compute/fiber_only_layout.go` 的 HStack 布局函数中添加日志：

```go
func (e *Engine) layoutFiberOnlyHStack(box *ComputedBox, x, y int) {
    // ... 现有代码 ...

    e.debugLog.Printf("[HStack] Layout children at (%d, %d):", x, y)
    for i, child := range box.Children {
        e.debugLog.Printf("[HStack]   [%d] Tag=%s, NodeID=%d, x=%d, w=%d",
            i, child.Tag, child.NodeID, child.Box.X, child.Box.Width)
    }
}
```

### 步骤 2: 添加 HitMap 调试

在 `BuildHitMapFromFiber` 中添加日志，输出每个按钮的边界：

```go
// 对于 HStack 的子按钮，输出详细的边界信息
if fiber.Tag == "button" {
    log.Printf("[HitMap] Button NodeID=%d, Bounds=(%d,%d,%d,%d), ZOrder=%d",
        fiber.NodeID, x, y, width, height, zOrder)
}
```

### 步骤 3: 创建可视化调试工具

运行 fiber_counter 时输出 ASCII 艺术图，显示按钮和文本的相对位置：

```
+------+  +------+  +------+
|  -   |  |      |  |  +   |
+------+  | 1sp  |  +------+
          +------+
          
如果 1sp 处有重复：
+------+  +------+  +------+
|  -   |  |  +   |  |  +   |  <-- 重复的 "+"
+------+  +------+  +------+
          ↑ 第二个位置
```

### 步骤 4: 检查 sibling 链表的完整性

在渲染前验证 sibling 链表没有循环引用或重复：

```go
func validateSiblingChain(fiber *Fiber) error {
    seen := make(map[uint64]bool)
    for child := fiber.Child; child != nil; child = child.Sibling {
        if seen[child.NodeID] {
            return fmt.Errorf("duplicate NodeID %d in sibling chain", child.NodeID)
        }
        seen[child.NodeID] = true
    }
    return nil
}
```

## 代码修复方向

基于假设，可能的修复方案：

### 修复方案 1: 强制唯一的 Gap 计算

修改 HStack 的 gap 计算，确保奇数个子节点也能正确分布：

```go
// 在 layoutFiberOnlyHStack 中
totalChildWidth := 0
for i, child := range box.Children {
    totalChildWidth += child.Box.Width
    // 只在非最后一个子节点后加 gap
    if i < len(box.Children)-1 {
        totalChildWidth += gap
    }
}
```

### 修复方案 2: 添加边界验证

在 ComputedBox 创建后验证边界不重叠：

```go
func (cb *ComputedBox) ValidateBounds() error {
    // 检查所有子节点的边界
    // 如果有重叠，记录警告或调整位置
    return nil
}
```

### 修复方案 3: 明确的 Key 设置

为每个按钮添加明确的 key，确保 reconciliation 正确：

```go
ui.HStack(
    app.ButtonBuilder(" - ").
        Key("dec-button").
        OnPress(intent.Decrement("count", 1)).
        Build(),
    ui.Text(" "),
    app.ButtonBuilder(" + ").
        Key("inc-button").
        OnPress(intent.Increment("count", 1)).
        Build(),
)
```

## 总结

1. **问题**: HStack 包含 3 个子节点时，最后一个按钮重复显示
   - 4+ 个子节点时正常
   - Fiber tree 和 reconciliation 正确

2. **根本原因**: 推测在布局、Hit Testing 或渲染层面
   - 可能涉及 gap 计算、边界重叠或缓存

3. **临时解决方案**: 添加第二个空格
   - 改变了索引分布和布局计算
   - 不是一个根本性的修复

4. **下一步**: 需要实际运行 UI 并观察
   - 添加详细的日志
   - 创建可视化调试工具
   - 验证边界框和 HitMap

## 相关文件

- `examples/fiber_counter/main.go` - 受影响的示例
- `runtime/compute/fiber_only_layout.go` - 布局计算
- `runtime/ui/fiber_util.go` - HitMap 计算
- `internal/reconciler/diff.go` - Reconciliation 逻辑

## 测试代码

以下测试用例可以帮助验证修复：

```go
// TestHStackButtonLayout 验证 3 个子节点的 HStack 布局
func TestHStackButtonLayout(t *testing.T) {
    // 创建 [Button, Text, Button]
    // 验证边界不重叠
    // 验证 HitMap 中有 3 个正确的条目
}

// TestHStackButtonHitTesting 验证点击测试
func TestHStackButtonHitTesting(t *testing.T) {
    // 模拟点击每个按钮
    // 验证只有一个按钮被激活
}
```

---

**报告日期**: 2026-02-27
**报告人**: Qwen Code
**状态**: 待进一步调查
