# Button.Paint 未被调用的根本原因分析

**日期**: 2025-01-07
**状态**: 进行中

---

## 已确认的事实

### ✅ SetBounds 被调用
```
[calculatePositions] BUTTON: SetBounds type assertion SUCCESS
calling SetBounds(x=1, y=12, w=18, h=0)
calling SetBounds(x=19, y=12, w=18, h=0)
```

**说明**：
- `calculatePositions()` 确实执行到了按钮
- 类型断言成功：`box.VNode.(interface{ SetBounds(...) })`
- 宽度计算正确：w=18

### ❌ Button.Paint() 未被调用

**说明**：
- 添加到 `Button.Paint()` 开头的日志完全没有输出
- 说明 `Button.Paint()` 根本没有被调用

---

## 问题定位

### 渲染流程

```
RenderingPipeline.Render()
├─ Phase 1: layoutEngine.Layout(vnode, constraints)
│  ├─ buildComputedBox() - 构建 ComputedBox 树
│  └─ calculatePositions() - 设置位置和 SetBounds ✅
└─ Phase 2: paintEngine.Paint(layout, buffer)
   ├─ paintNode(layout.Root, buffer)
   │  ├─ 类型断言：box.VNode.(interface{ Paint(...) })
   │  ├─ 如果成功 → paintable.Paint() ✅
   │  └─ 如果失败 → 继续处理子节点
```

### 问题出在哪里？

**PaintEngine.paintNode()** (internal/render/paint_engine.go:57-114)

```go
func (e *PaintEngine) paintNode(box *compute.ComputedBox, buffer *paint.Buffer) error {
    // ...
    // FIRST: Check if vnode implements Paintable interface
    if paintable, ok := box.VNode.(interface{ Paint(int, int) []paint.DrawCmd }); ok {
        // Component has custom paint logic - use it
        commands := paintable.Paint(box.Box.X, box.Box.Y)
        // ...
        return nil
    }
    // ...
}
```

**关键问题**：类型断言可能失败！

---

## 可能的原因

### 原因1：ComputedBox.VNode 不是 ButtonVNode

**推测**：`box.VNode` 可能是 `*ui.ElementVNode` 而不是 `*button.ButtonVNode`

**验证方法**：
```go
fmt.Fprintf(os.Stderr, "[Paint] VNode type=%T\n", box.VNode)
```

### 原因2：ButtonVNode 被包装了

**推测**：Wrap 组件在 Build 时可能把按钮包装成其他类型

**检查点**：
- `wrap.go:303-320` - Wrap.FillWidth() 是否修改了 VNode 类型？
- `layout.go` - LayoutNode.Build() 是否保留了原始 VNode？

### 原因3：PaintEngine 没有遍历到按钮

**推测**：PaintEngine.Paint() 可能在某个父节点就停止了

**检查点**：
- Bordered 组件的绘制逻辑
- HStack/VStack 的绘制逻辑

---

## 下一步调查

### 调查1：检查 VNode 类型

在 `PaintEngine.paintNode()` 开头添加：
```go
if e.debug {
    fmt.Fprintf(os.Stderr, "[Paint] node type=%T, tag=%s\n",
        box.VNode, getTag(box.VNode))
}
```

### 调查2：检查是否遍历到按钮

在 `paintNode()` 中添加子节点遍历日志：
```go
if e.debug {
    fmt.Fprintf(os.Stderr, "[Paint] children=%d\n", len(box.Children))
}
```

### 调查3：检查类型断言

```go
paintable, ok := box.VNode.(interface{ Paint(int, int) []paint.DrawCmd })
if e.debug {
    fmt.Fprintf(os.Stderr, "[Paint] Paintable=%v\n", ok)
}
```

---

## 临时解决方案

如果类型断言确实失败，可以：

1. **方案A**：修改 `Button.Paint()` 接收参数
   - 不依赖 VNode 类型断言
   - 通过其他机制传递布局宽度

2. **方案B**：在 LayoutNode 缓存布局宽度
   - 之前提出的方案1
   - 通过 props 传递 `_layoutWidth`

3. **方案C**：修复类型断言问题
   - 确保按钮 VNode 不被包装
   - 确保 ComputedBox 存储原始 ButtonVNode

---

**优先级**：先确认根本原因（调查1-3），再选择解决方案
