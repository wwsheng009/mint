# 约束传播问题分析

## 问题现状

TreeView 实现了 `Measurable` 接口，但在运行时没有收到 bounded height constraints。

从日志看到：
```
[TreeView.Measure] No bounded height, rendering all 34 lines
```

## 约束传播链

### 预期的传播链
```
Bordered(Height=25)
  ↓ measureBordered: innerConstraints.MaxHeight = 23
    VStack(content)
    ↓ measureLayoutChildren
      Tabs(Height=21 prop)
      ↓ Tabs.Measure() 返回 Height=21
        children: [tabBar, content(VStack)]
        ↓ Tabs.Measure() 测量 content
          maxContentHeight = 21 - 1 = 20 ✓
          content.Measure({MaxHeight=20})
        ↓ 但是！Tabs.SetChildren() 设置的 children 没有获得约束
        ↓ 布局引擎会再次单独测量这些 children
          VStack (from buildElementsTabContent)
          ↓ measureLayoutChildren
            children: [header, selectedInfo, treePreview(TreeView), instructions]
            treePreview.Measure({MaxHeight=Infinity}) ✗
```

### 问题所在

1. **Tabs.Measure()** 测量了 content，计算了自己的高度
2. 但是布局引擎在 **Layout.Position()** 阶段会再次定位 Tabs 的 children
3. 在定位时，children **不会自动获得** Tabs.Measure() 中使用的约束
4. VStack (content) 没有将 Tabs 中的 bounded height 传递给它的子节点

### 根本原因

**VStack (LayoutNode) 实现了 Measurable 吗？**

让我检查：

```go
// runtime/compute/engine.go:223-233
if measurable, ok := vnode.(interface {
    Measure(runtime.BoxConstraints) runtime.Size
}); ok {
    size := measurable.Measure(constraints)
    return size
}
```

LayoutNode (VStack 的类型) **没有实现** `Measure()` 方法，所以它会 fallback 到 `measureLayoutChildren()`。

**但是 `measureLayoutChildren()` 不会检查 LayoutNode 自己的 props**（比如从父组件继承的 height constraint）。

## 解决方案

### 方案 1：让 LayoutNode 实现 Measure()（推荐）

让 LayoutNode (VStack/HStack) 实现 Measure() 接口，这样它们可以：
1. 检查自己的 props 中是否有 height constraint
2. 如果有，将这个约束传递给子节点

### 方案 2：让 TreeView 使用 VBox 包装

TreeView 本身是 ElementVNode，而不是 LayoutNode。即使它实现了 Measure()，如果它被包装在 VStack 中，VStack 也不会传递自己的约束。

### 方案 3：直接使用 TreeView，不包装在 VStack 中

```go
// 当前：包装在 VStack 中
treePreview := ui.VStackBuilder(si.treeViewComponent).
    Flex(1).
    Build()

// 修改：直接使用 TreeView
treePreview := si.treeViewComponent
```

这样 TreeView 会直接接收父容器（Tabs 的 content VStack）的约束。

## 当前状态

✅ TreeView 实现了 Measurable 接口
✅ TreeView.Measure() 会检查 constraints.HasBoundedHeight()
✅ TreeView 会调用 SetViewportHeight() 启用虚拟滚动

❌ 但是约束没有正确传播到 TreeView
❌ TreeView.Measure() 被调用，但 constraints.MaxHeight = Infinity

## 下一步

**最简单的解决方案**：让 TreeView 直接作为子节点，不要包装在 VStack 中：

```go
// 在 buildElementsTabContent() 中
return rtui.VStack(
    header,
    selectedInfo,
    si.treeViewComponent,  // 直接使用，不包装
    instructions,
)
```

然后在父 VStack 上设置约束：
```go
// 在 buildOverlayContent() 中
content := ui.VStackBuilder(
    titleBar,
    tabsComponent,
    ui.Text("─"),
).Build()
```

并在 Tabs 的内容（即 buildElementsTabContent 返回的 VStack）上设置高度约束。

或者，让 buildElementsTabContent 返回的 VStack 使用 bounded height：
```go
func (si *StandaloneInspector) buildElementsTabContent() rtui.VNode {
    // ...
    result := rtui.VStack(
        header,
        selectedInfo,
        si.treeViewComponent,
        instructions,
    )

    // Apply height constraint
    return ui.VStackBuilder(result).
        Height(20).  // Or calculate dynamically
        Build()
}
```

## 关键洞察

**约束传播不是自动的！**

每个组件都必须：
1. 实现 Measurable 接口来接收约束
2. 或者使用 LayoutBuilder 并设置 Height/Width props
3. 或者被包装在有明确约束的容器中

TreeView 实现了 #1，但它被包装在 VStack 中（#2），而 VStack 本身没有获得约束。

**约束的来源必须明确**：
- Bordered.Height() → 传递给 child
- LayoutNode.Height() → 应该传递给 children
- Flex(1) → 需要父容器有 bounded height
