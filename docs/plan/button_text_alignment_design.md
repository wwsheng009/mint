# 按钮文本居中对齐 - 架构设计方案

**日期**: 2025-01-07
**状态**: 设计阶段

---

## 问题分析

### 当前问题

**症状**: 按钮被拉伸后，文本左对齐，没有居中

```
当前：>[ [1] Event ]     (14字符文本 + 5个右边空格，左对齐)❌
期望：  >[ [1] Event ]   (2个左边空格 + 14字符文本 + 3个右边空格，居中)✅
```

### 根本原因

**架构问题**: 对齐逻辑混在渲染代码中，违反了"布局与渲染分离"原则

当前错误实现：
```go
// components/button/button.go:650
buttonText += strings.Repeat(" ", padding)  // ❌ 渲染时手动填充
```

---

## 正确的架构

### 职责分离

```
┌─────────────────────────────────────────────────────────┐
│  Layout Engine (布局引擎)                               │
│                                                          │
│  职────────────────────────────────────────────────┐   │
│  │  LayoutNode.MeasureLayout()                     │   │
│  │    - 计算flex分配                              │   │
│  │    - 返回 LayoutMeasurement                  │   │   │
│  └────────────────────────────────────────────────┘   │
│                      ↓                                 │
│  ┌────────────────────────────────────────────────┐   │
│  │  Engine.calculatePositions()                   │   │
│  │    - 读取容器的 align 属性                    │   │
│  │    - 计算每个子元素的 (x, y) 位置             │   │
│  │    - 根据 align 调整子元素位置                │   │
│  │      * AlignStart:  x = left                   │   │
│  │      * AlignCenter: x = left + (alloc-natural)/2│   │
│  │      * AlignEnd:    x = left + (alloc-natural)  │   │
│  └────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────┘
                      ↓
┌─────────────────────────────────────────────────────────┐
│  Paint Engine (渲染引擎)                                 │
│                                                          │
│  ┌────────────────────────────────────────────────┐   │
│  │  Button.Paint()                                 │   │
│  │    - 返回自然宽度文本（不含填充空格）            │   │
│  │    - 在 (x, y) 位置绘制                           │   │
│  └────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────┘
```

---

## 实现方案

### 阶段1: 在ComputedBox中存储自然宽度

**问题**: calculatePositions() 需要知道子元素的自然宽度才能计算居中位置

**方案**: 在LayoutEngine.Layout()阶段，同时存储自然宽度和布局宽度

```go
// runtime/compute/computed_box.go
type ComputedBox struct {
    Box           Box
    VNode         ui.VNode
    Children      []*ComputedBox

    // ⭐ 新增：存储自然宽度（用于对齐计算）
    NaturalWidth  int
}
```

### 阶段2: 修改calculatePositions()支持对齐

**位置**: `runtime/compute/engine.go:layoutHStack()`

**实现**:

```go
func (e *Engine) layoutHStack(box *ComputedBox, x, y int) {
    layoutInfo := rtui.GetLayoutInfo(box.VNode)
    mainAlign := layoutInfo.Align

    for i, child := range box.Children {
        childBoxWidth := child.Box.Width        // 布局宽度（分配的）
        childNaturalWidth := child.NaturalWidth  // 自然宽度（文本）

        // 计算childX（考虑对齐）
        childX := x  // 默认左对齐

        if childNaturalWidth < childBoxWidth {
            // 子元素没有填满分配空间，应用对齐
            switch mainAlign {
            case rtui.AlignCenter:
                // 居中：左边填充 = (分配宽度 - 自然宽度) / 2
                padding := (childBoxWidth - childNaturalWidth) / 2
                childX = x + padding

            case rtui.AlignEnd:
                // 右对齐：左边填充 = 分配宽度 - 自然宽度
                padding := childBoxWidth - childNaturalWidth
                childX = x + padding

            case rtui.AlignStart:
                // 左对齐（默认）：不需要调整
                childX = x
            }
        }

        // 递归布局子元素
        e.calculatePositions(child, childX, y)

        // 下一个元素的起始位置
        x += child.Box.Width + layoutInfo.Gap
    }
}
```

### 阶段3: 移除Button.Paint()中的填充逻辑

**位置**: `components/button/button.go:637-656`

**修改**:

```go
// ❌ 删除这部分代码
if b.bounds[2] > 0 {
    layoutWidth := b.bounds[2]
    if layoutWidth > buttonWidth {
        padding := layoutWidth - buttonWidth
        buttonText += strings.Repeat(" ", padding)  // ❌ 删除
        buttonWidth = layoutWidth
    }
}
```

**原因**: 对齐由布局引擎处理，渲染时不应再添加空格

---

## 数据流

### 完整流程

```
1. 用户调用:
   WrapBuilder(
       Button("Test"),
       Button("Hello"),
   ).Align(ui.AlignCenter).FillWidth()

2. Wrap.Build():
   → 创建HStack
   → 设置 hstack.align = AlignCenter
   → 给每个button设置 flex=1

3. LayoutEngine.Layout():
   → HStack.MeasureLayout()
     * 计算flex分配：每个button分配19字符
     * 返回 LayoutMeasurement

   → Engine.calculatePositions()
     * 读取 hstack.align = AlignCenter
     * 遍历每个button子元素
     * 计算button.NaturalWidth = 14
     * 计算：childX = 1 + (19-14)/2 = 1 + 2.5 = 3
     * 调用 calculatePositions(button, 3, y)

4. Button.Paint():
   * 绘制位置 x=3 (不是1！)
   * 只返回自然宽度文本：`>[ [1] Event ]`
   * 文本在(3, y)位置绘制，自动居中✅

5. 视觉结果:
   |  >[ [1] Event ]   |  (2个左边空格 + 14文本 + 3个右边空格)
   ↑ 3个空格缩进
```

---

## 技术挑战

### 挑战1: 获取自然宽度

**问题**: calculatePositions() 时如何获取子元素的自然宽度？

**方案A**: 存储在ComputedBox中（推荐）
```go
type ComputedBox struct {
    ...
    NaturalWidth int  // 自然宽度
}
```

**方案B**: 动态测量
```go
// 在calculatePositions()中测量
if measurable, ok := child.VNode.(interface{
    Measure(runtime.BoxConstraints) runtime.Size
}); ok {
    size := measurable.Measure(runtime.BoxConstraints{MaxWidth: Infinity})
    naturalWidth := size.Width
}
```

### 挑战2: VStack中的对齐

VStack的crossAxis同样需要对齐支持：
```go
case rtui.AlignCenter:
    childY = y + (box.Box.Height - child.Box.Height) / 2
```

这部分已经实现（第969-977行）✅

### 挑战3: 不同align模式的支持

- **AlignStart**: 左对齐（默认）
- **AlignCenter**: 居中对齐
- **AlignEnd**: 右对齐
- **AlignSpaceBetween**: 两端对齐（不适用于单个元素）
- **AlignSpaceAround**: 环绕对齐（不适用于单个元素）

---

## 实现优先级

### P0: 基础架构（必须）

1. ✅ 移除Button.Paint()中的填充逻辑
2. ⏳ 在ComputedBox中添加NaturalWidth字段
3. ⏳ 修改layoutHStack()支持居中对齐

### P1: 改进对齐支持

4. ⏳ 支持AlignEnd（右对齐）
5. ⏳ 支持VStack的crossAxis居中（已实现✅）

### P2: 高级功能

6. ⏳ 支持自定义对齐（per-child alignment）
7. ⏳ 支持Text元素的对齐
8. ⏳ 支持Input元素的对齐

---

## 示例代码

### 修改1: 添加NaturalWidth到ComputedBox

```go
// runtime/compute/computed_box.go
type ComputedBox struct {
    Box           Box
    VNode         ui.VNode
    Children      []*ComputedBox
    NaturalWidth  int  // ⭐ 新增
}
```

### 修改2: 在LayoutEngine中设置NaturalWidth

```go
// runtime/compute/engine.go:buildComputedBox()
func (e *Engine) buildComputedBox(vnode ui.VNode) *ComputedBox {
    box := &ComputedBox{
        VNode: vnode,
    }

    // 测量自然宽度
    if measurable, ok := vnode.(interface{
        Measure(runtime.BoxConstraints) runtime.Size
    }); ok {
        size := measurable.Measure(runtime.BoxConstraints{MaxWidth: Infinity})
        box.NaturalWidth = size.Width
    }

    // ... 其他逻辑
    return box
}
```

### 修改3: layoutHStack支持对齐

```go
// runtime/compute/engine.go:layoutHStack()
func (e *Engine) layoutHStack(box *ComputedBox, x, y int) {
    layoutInfo := rtui.GetLayoutInfo(box.VNode)
    mainAlign := layoutInfo.Align

    for i, child := range box.Children {
        childX := x

        // ⭐ 应用对齐
        if child.NaturalWidth < child.Box.Width {
            switch mainAlign {
            case rtui.AlignCenter:
                padding := (child.Box.Width - child.NaturalWidth) / 2
                childX = x + padding
            case rtui.AlignEnd:
                padding := child.Box.Width - child.NaturalWidth
                childX = x + padding
            }
        }

        e.calculatePositions(child, childX, childY)
        x += child.Box.Width + layoutInfo.Gap
    }
}
```

### 修改4: 移除Button的填充逻辑

```go
// components/button/button.go:Paint()
func (b *ButtonVNode) Paint(x, y int) []paint.DrawCmd {
    // ... 构建buttonText ...

    // ❌ 删除这部分
    // if b.bounds[2] > 0 {
    //     padding := ...
    //     buttonText += strings.Repeat(" ", padding)
    // }

    // ✅ 只返回自然宽度文本
    return []paint.DrawCmd{
        paint.NewTextCmd(x, y, buttonText, buttonStyle),
    }
}
```

---

## 验证方法

### 测试1: 左对齐（默认）

```go
WrapBuilder(
    Button("Test"),
).Align(ui.AlignStart)  // 或不设置
```

预期：
```
|[Test]            |
```

### 测试2: 居中对齐

```go
WrapBuilder(
    Button("Test"),
).Align(ui.AlignCenter)
```

预期：
```
|     [Test]      |
```

### 测试3: 右对齐

```go
WrapBuilder(
    Button("Test"),
).Align(ui.AlignEnd)
```

预期：
```
|            [Test]|
```

---

## 时间估算

| 任务 | 估算时间 |
|------|---------|
| 移除Button填充逻辑 | 5分钟 |
| 添加NaturalWidth字段 | 15分钟 |
| 修改buildComputedBox | 15分钟 |
| 修改layoutHStack支持对齐 | 30分钟 |
| 测试和验证 | 30分钟 |
| **总计** | **1.5小时** |

---

## 风险评估

### 低风险

- 修改只影响布局引擎，不影响其他组件
- 向后兼容：AlignStart（默认）行为不变

### 中风险

- 需要测试所有align模式
- 可能影响现有UI的对齐效果

### 缓解措施

- 添加详细日志
- 保留旧的Button填充逻辑作为fallback
- 分阶段发布，先支持AlignCenter

---

## 下一步行动

1. ⏳ 创建功能分支
2. ⏳ 实现NaturalWidth存储
3. ⏳ 实现layoutHStack对齐支持
4. ⏳ 移除Button填充逻辑
5. ⏳ 测试验证
6. ⏳ 更新文档

---

**结论**: 这是一个架构级别的改进，将使Mint TUI的对齐系统更加规范和强大。建议实施。
