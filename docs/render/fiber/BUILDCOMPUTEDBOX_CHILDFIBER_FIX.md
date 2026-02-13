# buildComputedBoxWithSize ChildFiber 实现详细步骤

## 问题分析

当前 `buildComputedBoxWithSize` 函数中，子节点的 NodeID 没有被正确赋值。

### 根本原因

1. `buildComputedBoxWithSize` 函数接收 `fiber` 参数（父节点的 Fiber）
2. 创建 `box` 时，`box.NodeID` 设置为 `fiber.NodeID`
3. 但在循环构建子节点时，调用 `e.buildComputedBox(child, nil, ...)`
4. `nil` 导致子节点的 `box.NodeID` 保持为 0

### 当前代码结构

```go
// runtime/compute/engine.go:130
func (e *Engine) buildComputedBoxWithSize(vnode VNode, fiber *reconciler.Fiber, parent *ComputedBox, constraints runtime.BoxConstraints, preMeasuredSize *runtime.Size) *ComputedBox {
    // ... 创建 box ...
    box := &ComputedBox{
        VNode:        vnode,
        Parent:       parent,
        Box:          runtime.Box{X: 0, Y: 0, Width: 0, Height: 0},
        NaturalWidth:  0,
        NodeID:       0, // Will be set from Fiber if provided
    }

    // 设置父节点的 NodeID
    if fiber != nil {
        box.NodeID = fiber.NodeID
    }

    // ... 测量 NaturalWidth ...

    // 构建子节点时问题：
    for i, child := range vnode.Children() {
        childConstraints := measurement.ChildConstraints[i]
        // ❌ 这里传了 nil！子节点的 NodeID 会保持为 0
        childBox := e.buildComputedBox(child, nil, box, childConstraints)
        if childBox != nil {
            box.Children = append(box.Children, childBox)
        }
    }
}
```

## 解决方案：Option 2 - ChildFiber 字段

### 第一步：在 ComputedBox 结构中添加 ChildFiber 字段

**文件**: `runtime/compute/types.go`

**位置**: 在 `ComputedBox` 结构体中，`NodeID` 字段之后添加

```go
type ComputedBox struct {
    // ... 现有字段 ...
    NodeID uint64

    // === 新增 ===
    // ChildFiber stores the Fiber node for this box (used for NodeID propagation to children)
    // See: docs/render/fiber/FIBER_ID.md - Option 2 implementation
    ChildFiber *rtui.Fiber
}
```

### 第二步：在 buildComputedBoxWithSize 中设置 ChildFiber

**文件**: `runtime/compute/engine.go`

**位置**: 在 `buildComputedBoxWithSize` 函数中，循环构建子节点之前

```go
func (e *Engine) buildComputedBoxWithSize(vnode VNode, fiber *reconciler.Fiber, parent *ComputedBox, constraints runtime.BoxConstraints, preMeasuredSize *runtime.Size) *ComputedBox {
    if vnode == nil {
        return nil
    }

    box := &ComputedBox{
        VNode:        vnode,
        Parent:       parent,
        Box:          runtime.Box{X: 0, Y: 0, Width: 0, Height: 0},
        NaturalWidth:  0,
        NodeID:       0, // Will be set from Fiber if provided
        ChildFiber:   nil, // ✅ 新增：Placeholder for child's Fiber (will be set before building children)
    }

    // Phase 3: Set NodeID from Fiber for stable identity
    // ✅ 修改：从父节点继承 NodeID（如果父节点存在）
    if parent != nil && parent.NodeID != 0 {
        box.NodeID = parent.NodeID
    } else if fiber != nil {
        box.NodeID = fiber.NodeID
    }

    // ... 测量 NaturalWidth 等逻辑 ...

    // 构建子节点
    // ...
    vnodeChildren := vnode.Children()
    box.Children = make([]*ComputedBox, 0, len(vnodeChildren))

    // ✅ 修改：在循环中设置 ChildFiber
    for i, child := range vnodeChildren() {
        childConstraints := measurement.ChildConstraints[i]

        // ✅ 新增：设置 box.ChildFiber = childFiber（在调用 buildComputedBox 之前）
        var childFiber *rtui.Fiber
        if fiber != nil && fiber.Child != nil && fiber.Child.VNode == child {
            childFiber = fiber.Child
        }

        // ✅ 修改：传递 childFiber 而不是 nil
        childBox := e.buildComputedBox(child, childFiber, box, childConstraints)
        if childBox != nil {
            box.Children = append(box.Children, childBox)
        }
    }
}
```

### 第三步：处理 FALLBACK 分支

**文件**: `runtime/compute/engine.go`

**位置**: 在 `buildComputedBoxWithSize` 函数的 FALLBACK 部分（使用 legacy two-pass approach）

```go
    // FALLBACK: Use legacy two-pass approach
    size := e.measureVNode(vnode, constraints)
    box.Box.Width = size.Width
    box.Box.Height = size.Height

    // REFRESH children: Measure() might have updated children list
    vnodeChildren = vnode.Children()

    // Build children layout boxes
    box.Children = make([]*ComputedBox, 0, len(vnodeChildren))

    // ✅ 修改：在循环中设置 ChildFiber
    for _, child := range vnodeChildren {
        childConstraints := e.getChildConstraints(vnode, child, constraints, size)

        // ✅ 新增：设置 box.ChildFiber = childFiber
        var childFiber *rtui.Fiber
        if fiber != nil && fiber.Child != nil && fiber.Child.VNode == child {
            childFiber = fiber.Child
        }

        // ✅ 修改：传递 childFiber 而不是 nil
        childBox := e.buildComputedBox(child, childFiber, box, childConstraints)
        if childBox != nil {
            box.Children = append(box.Children, childBox)
        }
    }
```

## NodeID 传播链

修复后，NodeID 将正确传播：

```
Parent Fiber (NodeID: 123)
  └─ box.NodeID = 123 (从父节点继承)
       └─ box.ChildFiber = childFiber (设置为子节点的 Fiber)
            └─ childBox.NodeID = childFiber.NodeID (在 buildComputedBox 中设置)
                 └─ 子节点的 ComputedBox 现在有正确的 NodeID
```

## 完整的 NodeID 传播链

```
Fiber.NodeID → ComputedBox.NodeID → box.ChildFiber → buildComputedBox() → childBox.NodeID
```

这确保了所有层级的节点都能获得正确的 NodeID！

## 操作步骤总结

### 步骤 1: 修改 runtime/compute/types.go

1. 打开文件 `runtime/compute/types.go`
2. 找到 `ComputedBox` 结构体定义（大约第31行）
3. 在 `NodeID uint64` 之后添加以下代码：

```go
// ChildFiber stores the Fiber node for this box (used for NodeID propagation to children)
// See: docs/render/fiber/FIBER_ID.md - Option 2 implementation
ChildFiber *rtui.Fiber
```

### 步骤 2: 修改 runtime/compute/engine.go - buildComputedBoxWithSize 函数

1. 打开文件 `runtime/compute/engine.go`
2. 找到 `buildComputedBoxWithSize` 函数定义（大约第130行）
3. 在创建 `box` 结构体时，添加 `ChildFiber: nil,`

```go
box := &ComputedBox{
    VNode:        vnode,
    Parent:       parent,
    Box:          runtime.Box{X: 0, Y: 0, Width: 0, Height: 0},
    NaturalWidth:  0,
    NodeID:       0, // Will be set from Fiber if provided
    ChildFiber:   nil, // 新增这一行
}
```

4. 修改 NodeID 设置逻辑，从父节点继承：

```go
// 修改前
if fiber != nil {
    box.NodeID = fiber.NodeID
}

// 修改后
if parent != nil && parent.NodeID != 0 {
    box.NodeID = parent.NodeID
} else if fiber != nil {
    box.NodeID = fiber.NodeID
}
```

5. 找到第一个构建子节点的循环（大约第209行）

**单次测量分支（Single-pass）:**
```go
// 修改前
for i, child := range vnode.Children() {
    childConstraints := measurement.ChildConstraints[i]
    childBox := e.buildComputedBox(child, nil, box, childConstraints)
    if childBox != nil {
        box.Children = append(box.Children, childBox)
    }
}

// 修改后
for i, child := range vnode.Children() {
    childConstraints := measurement.ChildConstraints[i]

    // 新增：在循环中设置 ChildFiber
    var childFiber *rtui.Fiber
    if fiber != nil && fiber.Child != nil && fiber.Child.VNode == child {
        childFiber = fiber.Child
    }

    // 修改：传递 childFiber 而不是 nil
    childBox := e.buildComputedBox(child, childFiber, box, childConstraints)
    if childBox != nil {
        box.Children = append(box.Children, childBox)
    }
}
```

**FALLBACK 分支（Legacy two-pass）:**
```go
// 修改前
for _, child := range vnodeChildren {
    childConstraints := e.getChildConstraints(vnode, child, constraints, size)
    childBox := e.buildComputedBox(child, nil, box, childConstraints)
    if childBox != nil {
        box.Children = append(box.Children, childBox)
    }
}

// 修改后
for _, child := range vnodeChildren {
    childConstraints := e.getChildConstraints(vnode, child, constraints, size)

    // 新增：在循环中设置 ChildFiber
    var childFiber *rtui.Fiber
    if fiber != nil && fiber.Child != nil && fiber.Child.VNode == child {
        childFiber = fiber.Child
    }

    // 修改：传递 childFiber 而不是 nil
    childBox := e.buildComputedBox(child, childFiber, box, childConstraints)
    if childBox != nil {
        box.Children = append(box.Children, childBox)
    }
}
```

### 步骤 3: 验证编译

```bash
cd E:\projects\yao\wwsheng009\mint
go build ./...
```

### 步骤 4: 提交代码

```bash
git add runtime/compute/types.go runtime/compute/engine.go
git commit -m "feat(compute): Add ChildFiber field and implement NodeID propagation to children

- Add ChildFiber *rtui.Fiber field to ComputedBox struct
- Set box.ChildFiber before building child boxes in buildComputedBoxWithSize
- Inherit NodeID from parent when parent exists
- Pass childFiber to buildComputedBox instead of nil

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

## 关键点总结

1. **ComputedBox.ChildFiber** - 存储子节点的 Fiber 引用
2. **设置时机** - 在调用 `buildComputedBox` 之前设置 `box.ChildFiber`
3. **Fiber 获取** - 通过 `fiber.Child.VNode == child` 检查找到对应的子 Fiber
4. **NodeID 继承** - 如果父节点有 NodeID，子节点继承父节点的 NodeID
5. **两处修改** - 单次测量分支和 FALLBACK 分支都需要修改
