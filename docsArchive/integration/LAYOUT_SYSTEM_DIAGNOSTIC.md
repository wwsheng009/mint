# 布局系统诊断报告

## 症状

- Inspector overlay 不显示
- 或者 TreeView 内容溢出
- 或者布局系统崩溃（无限递归、内存不足）

## 根本原因分析

### 1. 双重测量系统问题

**问题**：Compute engine 和 LayoutNode.Measure() 是两套独立的实现

```go
// compute/engine.go:223-241
// PRIORITY 1: Use Measurable interface
if measurable, ok := vnode.(interface {
    Measure(runtime.BoxConstraints) runtime.Size
}); ok {
    return measurable.Measure(constraints)  // ← 会调用 LayoutNode.Measure()
}

// PRIORITY 2: Check for known layout types
if tagger, ok := vnode.(interface{ Tag() string }); ok {
    switch tagger.Tag() {
    case "hstack", "vstack":
        return e.measureLayoutChildren(vnode, constraints)  // ← 但这里会优先！
    }
}
```

**实际情况**：
- LayoutNode 确实实现了 Measure() 接口 (runtime/ui/layout.go:306)
- 但 compute engine 检查到 "vstack" tag 后，直接调用 measureLayoutChildren()
- **LayoutNode.Measure() 永远不会被调用！**

**影响**：
- LayoutNode.Measure() 中的约束传播逻辑不会执行
- 所有依赖都落在 measureLayoutChildren() 上

### 2. measureLayoutChildren 的约束传播缺陷

**原始代码**（已修复）：
```go
// VStack 测量子节点
childConstraints := runtime.BoxConstraints{
    MinWidth:  childMinWidth,
    MaxWidth:  innerMaxWidth,
    MinHeight: 0,
    MaxHeight: runtime.Infinity,  // ❌ 问题：没有传递高度约束！
}
```

**修复后的代码**：
```go
// 添加了 innerMaxHeight 定义
innerMaxHeight := runtime.Infinity
if constraints.HasBoundedHeight() {
    innerMaxHeight = max(0, constraints.MaxHeight-paddingHeight)
}

// 使用 innerMaxHeight 而不是 Infinity
childConstraints := runtime.BoxConstraints{
    MaxHeight: innerMaxHeight,  // ✓ 已修复
}
```

### 3. LayoutNode props 检查缺失

**问题**：measureLayoutChildren() 最初没有检查 LayoutNode 的 width/height props

**修复**：
```go
// 添加了对 props 的检查
if props != nil {
    if h, ok := props["height"].(int); ok && h > 0 {
        explicitHeight = h
        hasHeightProp = true
    }
}

// 使用 props 约束
if hasHeightProp {
    constraints = runtime.NewBoxConstraints(
        constraints.MinWidth,
        constraints.MaxWidth,
        explicitHeight,
        explicitHeight,
    )
}
```

### 4. Tabs 组件的约束传播问题

**问题**：Tabs 组件有两重角色：
1. 作为 Measurable：Tabs.Measure() 使用 height prop
2. 作为容器：children 会被单独测量

**导致的冲突**：
- Tabs.Measure() 测量 content 时使用 bounded constraints
- 但 layout engine 测量 Tabs 的 children 时，使用不同的约束
- content 不会继承 Tabs 的 height constraint

**失败的修复尝试**：
```go
// ❌ 导致无限递归
func (t *TabsVNode) updateActiveContent() {
    // ...
    contentNode = ui.VStackBuilder(contentNode).Height(contentHeight).Build()
    t.SetChildren([]ui.VNode{tabBarNode, contentNode})
    // ↑ SetChildren 可能触发某些回调，导致无限循环
}
```

### 5. TreeView 组件的角色混乱

**问题**：TreeView 既是组件，又是容器

```go
// TreeView 作为子节点时的结构
TreeView
  └─ children: [VStack]  ← SetChildren() 设置的
        └─ [Text, Text, ...]  ← 实际的树内容

// 当 TreeView 被添加到 VStack 时：
VStack
  ├─ header
  ├─ selectedInfo
  ├─ treeView  // ← TreeView 本身
  │   └─ children: [VStack]  ← TreeView 的子节点
  │         └─ [Text, Text, ...]
  └─ instructions
```

**问题**：
- TreeView.Measure() 会被调用，但它返回的大小不是它的实际渲染大小
- TreeView 的实际内容在它的 children 中（一个 VStack）
- 但它的 children 大小和 Measure() 返回的大小可能不一致

## 当前状态总结

### ✅ 已修复

1. **measureLayoutChildren 使用 innerMaxHeight**
   - VStack 现在会传递父容器的 bounded height 给子节点
   - 文件：runtime/compute/engine.go:463-466, 503

2. **LayoutNode props 检查**
   - measureLayoutChildren 现在会检查并使用 width/height props
   - 文件：runtime/compute/engine.go:320-357

3. **TreeView 实现了 Measure() 接口**
   - TreeView 可以响应布局约束
   - 文件：components/display/treeview.go:872-905

### ❌ 仍然存在的问题

1. **Tabs 组件的约束传播**
   - content 不会继承 Tabs 的 height constraint
   - 需要在 Tabs 层面解决

2. **TreeView 的角色混乱**
   - 既是组件又是容器，children 大小和 Measure() 大小不一致
   - 需要明确 TreeView 应该如何参与布局

3. **双重测量系统的冲突**
   - LayoutNode.Measure() 存在但不会被调用
   - measureLayoutChildren() 是实际使用的实现
   - 需要统一到一套系统

### 🔍 需要进一步调查的问题

1. **为什么 Inspector overlay 不显示？**
   - 可能是渲染层问题
   - 可能是布局计算失败
   - 需要检查 Inspector 的 layer 设置

2. **约束传播的完整链路**
   - Bordered → VStack → Tabs → VStack → TreeView
   - 每一层是否正确传递约束？

3. **flex 布局的计算**
   - Flex(1) 的子节点是否正确获得分配的空间？
   - flex distribution 逻辑是否正确？

## 修复建议

### 短期修复（快速解决 Inspector 问题）

1. **移除 measureLayoutChildren 中的 tag 检查**
   ```go
   // 不应该检查 tag，应该让 LayoutNode.Measure() 处理
   // 删除这段代码：
   // case "hstack", "vstack":
   //     return e.measureLayoutChildren(vnode, constraints)
   ```

2. **或者删除 LayoutNode.Measure()**
   - 既然不会调用，就删除这个混淆的实现
   - 统一使用 measureLayoutChildren()

### 长期修复（架构级别）

1. **统一测量系统**
   - 选择一个测量路径：
     - Option A: 全部使用 Measure() 接口（推荐）
     - Option B: 全部使用 compute engine 的逻辑

2. **明确组件角色**
   - TreeView 应该是纯组件，不应该有 children
   - 或者 TreeView 应该明确是一个容器组件

3. **约束传播规范**
   - 定义明确的规则：props 如何影响约束
   - 父容器如何传递约束给子节点
   - Measurable 组件如何影响布局

## 下一步行动

1. ✅ 回退失败的 Tabs 修改
2. ❓ 确认 Inspector overlay 为什么不显示
3. ❓ 测试当前的 measureLayoutChildren 修复是否有效
4. ❓ 决定长期架构方向

## 测试用例

需要添加以下测试来验证修复：

```go
// Test constraint propagation through VStack
func TestVStackPropagatesHeightConstraint(t *testing.T) {
    outer := ui.VStackBuilder(
        ui.Text("Line 1"),
        ui.Text("Line 2"),
        ui.Text("Line 3"),
    ).Height(10).Build()  // Bounded height

    // 子节点应该收到 MaxHeight=10 的约束
}

// Test TreeView with bounded constraints
func TestTreeViewWithBoundedHeight(t *testing.T) {
    treeView := display.NewTreeView().
        FromLines([]string{"A", "B", "C"}).
        Build()

    // 模拟布局引擎的测量
    constraints := runtime.BoxConstraints{
        MaxWidth: 80,
        MaxHeight: 5,  // Bounded!
    }

    size := treeView.Measure(constraints)
    // 应该返回 height=5，并且只渲染 5 行
}
```

## 相关文件

- runtime/compute/engine.go - 布局引擎核心
- runtime/ui/layout.go - LayoutNode 定义
- components/display/treeview.go - TreeView 组件
- components/navigation/tabs.go - Tabs 组件
- internal/inspector/standalone_inspector.go - Inspector 实现

## 时间线

- [ ] 2025-02-09: 发现双重测量系统问题
- [ ] 2025-02-09: 修复 measureLayoutChildren 约束传播
- [ ] 2025-02-09: 尝试修复 Tabs 约束传播（失败，无限递归）
- [ ] 2025-02-09: 回退 Tabs 修改
- [ ] 待定: 确定为什么 Inspector 不显示
- [ ] 待定: 完成所有修复并验证
