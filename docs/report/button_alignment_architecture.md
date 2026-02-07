# Button Text Alignment - Architecture Decision

## Problem Statement

如何在 TUI 中实现按钮文本对齐，同时保持"布局与渲染分离"的架构原则？

## Technical Constraint

在 TUI 中，没有"padding"概念，只能通过输出字符来填充空间：

```
Web:  <div style="padding: 2px; text-align: center">Button</div>
TUI:  输出 "  [Button]  "  (必须手动添加空格)
```

## Architecture Options

### Option 1: Component Handles Padding (Current Implementation)

```go
// Button.Paint()
if layoutWidth > naturalWidth {
    buttonText += strings.Repeat(" ", padding)  // 组件负责填充
}
```

**Pros:**
- 简单直接
- 每个组件可以自定义填充行为

**Cons:**
- ❌ 违反"布局与渲染分离"原则
- ❌ 每个组件都需要重复实现相同逻辑
- ❌ 难以统一管理对齐策略

### Option 2: Layout Engine Handles Position (Previous Implementation)

```go
// layoutHStack()
alignedX = childX + (allocatedWidth - naturalWidth) / 2
calculatePositions(child, alignedX, y)  // 布局引擎调整位置

// Button.Paint()
return buttonText  // 只返回自然宽度文本
```

**Pros:**
- ✅ 符合"布局与渲染分离"原则
- ✅ 布局引擎统一管理对齐
- ✅ 组件不需要关心布局细节

**Cons:**
- ⚠️ 文本在容器内居中，但按钮本身宽度不统一
- ⚠️ 视觉上可能不够整齐

### Option 3: Hybrid Approach (Recommended)

```go
// 1. 布局引擎计算对齐位置（已有）
layoutHStack() → alignedX = x + (allocated - natural) / 2

// 2. 布局引擎通过 bounds 传递对齐信息
SetBounds(x, y, width, height, {
    textAlign: AlignCenter,  // 布局引擎决定的对齐方式
})

// 3. 组件读取对齐信息并应用（可选）
Button.Paint() {
    if bounds.textAlign == AlignCenter {
        // 添加居中填充
    }
}
```

**Pros:**
- ✅ 布局引擎控制对齐策略（通过容器 Align 属性）
- ✅ 组件负责具体实现（添加空格）
- ✅ 清晰的职责划分
- ✅ 支持 CSS flex-like 行为

**Cons:**
- 需要扩展 bounds 数据结构

---

## Recommended Implementation

### Step 1: Extend Bounds with Alignment Info

```go
type Bounds struct {
    X, Y, Width, Height int
    TextAlign           ui.Align  // 新增：文本对齐
}

SetBounds(x, y, w, h int, options ...BoundsOption)
```

### Step 2: Layout Engine Sets Alignment

```go
// layoutHStack()
if child.NaturalWidth < child.Box.Width {
    // 传递容器的对齐方式给子元素
    textAlign := mainAlign  // 容器的 Align 属性
    child.SetBounds(alignedX, y, child.Box.Width, child.Box.Height,
        WithTextAlign(textAlign))
}
```

### Step 3: Component Reads Alignment

```go
// Button.Paint()
textAlign := b.bounds.TextAlign  // 从 bounds 读取
if layoutWidth > naturalWidth {
    switch textAlign {
    case AlignCenter:
        // 居中填充
    case AlignEnd:
        // 右对齐填充
    case AlignStart:
        // 左对齐填充
    }
}
```

---

## CSS Flex Mapping

| CSS Flex              | Mint TUI Implementation            |
|-----------------------|-----------------------------------|
| `justify-content`     | HStack.Align()                    |
| `align-items`         | HStack.CrossAlign()               |
| `flex`                | Flex property on child            |
| Text alignment        | Container.Align → Child.TextAlign|

### Example

```css
/* CSS */
.container {
    display: flex;
    justify-content: center;  /* 容器居中对齐 */
}

.button {
    flex: 1;  /* 拉伸填满 */
    text-align: center;  /* 文本居中 */
}
```

```go
// Mint TUI
HStack(
    Button("Button1").Flex(1),     // flex: 1
    Button("Button2").Flex(1),
).
    Align(ui.AlignCenter).  // justify-content: center
    Build()

// Layout engine automatically:
// 1. Distributes space to flex children
// 2. Aligns them to center
// 3. Passes AlignCenter to buttons for text alignment
```

---

## Conclusion

**架构原则优先级：**

1. ✅ **布局引擎控制位置** - calculatePositions() 计算对齐位置
2. ✅ **布局引擎传递对齐信息** - 通过 SetBounds 传递 textAlign
3. ✅ **组件负责渲染** - Button.Paint() 根据对齐信息添加空格

这样既保持了架构的清晰性，又解决了 TUI 的技术限制。

**下一步行动：**
1. 扩展 bounds 结构支持对齐信息
2. 布局引擎通过 SetBounds 传递对齐方式
3. 组件读取并应用对齐信息

---

**Implementation Priority:** High
**Complexity:** Medium
**Impact:** Major architecture improvement
