# Modal 位置控制指南

本文档说明如何控制 Modal 的显示位置。

**🔗 交互式示例**: [Modal Positioning Demo](../../../../examples/fiber_firsts/modal_positioning_demo/README.md)

---

## 目录

- [当前支持的定位方式](#当前支持的定位方式)
- [居中显示](#居中显示)
- [相对定位（默认）](#相对定位默认)
- [使用容器布局控制位置](#使用容器布局控制位置)
- [高级位置控制（绝对定位）](#高级位置控制绝对定位)
- [完整示例](#完整示例)

---

## 当前支持的定位方式

目前 Modal 组件通过 `centered` 属性来控制是否居中显示：

| 定位方式 | 说明 | Builder 方法 |
|---------|------|-------------|
| **居中显示** | Modal 在父容器的中心显示（默认） | `.Center()` 或 `.Centered(true)` |
| **非居中** | 使用父容器的布局系统定位 | `.Centered(false)` |

---

## 居中显示

默认情况下，Modal 是居中显示的：

```go
import (
    "github.com/wwsheng009/mint/ui"
    "github.com/wwsheng009/mint/ui/components/modal"
    "github.com/wwsheng009/mint/app"
)

func App() ui.VNode {
    return ui.VStack(
        app.Button("Open Modal", func() {
            // 打开 Modal
        }),
        // 居中显示的 Modal
        modal.NewBuilder().
            Title("Centered Modal").
            Content(app.Text("This modal is centered by default")).
            Width(50).
            Height(15).
            Center().  // 明确设置为居中
            Build(),
    )
}
```

### 居中实现原理

Modal 的 `centered` 属性为 `true` 时，布局引擎会计算父容器的中心位置，并将 Modal 放置在中心。

---

## 相对定位（默认）

当设置 `.Centered(false)` 时，Modal 会按照父容器的布局系统进行定位。

### 使用 Flex 容器定位

```go
func App() ui.VNode {
    isOpen, setIsOpen := ui.UseStateBool(false)

    return ui.VStack(
        ui.HStack(
            app.Text("Left aligned modal"),
            app.Button("Open", func() {
                setIsOpen(true)
            }),
        ),

        // Modal 会按照 Flex 布局定位（非居中）
        modal.NewBuilder().
            Title("Relative Position").
            Content(app.Text("This modal follows the container's layout")).
            Width(50).
            Height(15).
            Centered(false).  // 关闭居中，使用相对定位
            Open(isOpen).
            Build(),
    )
}
```

### 与绝对定位容器的区别

| 定位方式 | Modal 位置 | 适用场景 |
|---------|-----------|----------|
| `Centered(true)` | 父容器中心 | 大多数弹窗场景 |
| `Centered(false)` | 按父容器布局流排列 | 需要与其他元素配合布局时 |

---

## 使用容器布局控制位置

通过将 Modal 放入不同的容器，可以实现各种定位效果。

### 1. 左对齐 Modal

```go
func App() ui.VNode {
    return ui.HStack(
        ui.Text("  "),  // 左边距
        modal.NewBuilder().
            Title("Left Aligned").
            Content(app.Text("Left aligned modal")).
            Centered(false).
            Build(),
    )
}
```

### 2. 右对齐 Modal

```go
func App() ui.VNode {
    return ui.HStack(
        ui.Text("  "),  // 占位符
        modal.NewBuilder().
            Title("Right Aligned").
            Content(app.Text("Right aligned modal")).
            Centered(false).
            Build(),
        ui.Text("  "),  // 右边距
    )
}
```

### 3. 使用 Spacer 控制位置

```go
func App() ui.VNode {
    return ui.HStack(
        // 左边 20% 区域
        ui.Spacer().
            SetFlex(1).
            Build(),

        // Modal 放在中间
        modal.NewBuilder().
            Title("Custom Position").
            Content(app.Text("Positioned with spacers")).
            Centered(false).
            Build(),

        // 右边 20% 区域
        ui.Spacer().
            SetFlex(1).
            Build(),
    )
}
```

### 4. 垂直位置控制

```go
func App() ui.VNode {
    return ui.VStack(
        // 顶部区域
        ui.Spacer().
            SetFlex(1).
            Build(),

        // Modal 在垂直方向的 1/3 处
        modal.NewBuilder().
            Title("Top Positioned").
            Content(app.Text("Modal at 1/3 height")).
            Centered(false).
            Build(),

        // 底部区域（2 倍大小，使 Modal 在 1/3 处）
        ui.Spacer().
            SetFlex(2).
            Build(),
    )
}
```

---

## 高级位置控制（绝对定位）

如果需要精确控制 Modal 的位置（如相对于窗口的绝对坐标），可以使用绝对定位。

### 当前限制

当前 Modal 组件的 VNode 尚未暴露 `Position` 属性用于绝对定位。但这可以通过未来的扩展实现。

### 推荐方案（相对定位 + Spacer组合）

当前使用相对定位 + Spacer 可以实现大部分定位需求：

```go
func App() ui.VNode {
    return ui.VStack(
        // 顶部 15% 区域
        ui.Spacer().Flex(3).Build(),

        ui.HStack(
            // 左侧 10% 区域
            ui.Spacer().Flex(2).Build(),

            // Modal 在 (10%, 15%) 位置
            modal.NewBuilder().
                Title("Custom Position").
                Content(app.Text("Position: (10%, 15%)")).
                Width(60).
                Height(15).
                Centered(false).
                Build(),

            // 右侧 10% 区域
            ui.Spacer().Flex(2).Build(),
        ),

        // 底部 70% 区域
        ui.Spacer().Flex(14).Build(),
    )
}
```

### 计算比例

要实现特定的百分比位置：

```
top%    = topSpacer / totalSpacer
left%   = leftSpacer / totalSpacer
right%  = rightSpacer / totalSpacer
bottom% = bottomSpacer / totalSpacer
```

例如，要实现 `(20%, 30%)` 的位置：

- Top spacer: 2 份（占 20%）
- Bottom spacer: 8 份（占 80%）
- Left spacer: 3 份（占 30%）
- Right spacer: 7 份（占 70%）

```go
func App() ui.VNode {
    return ui.VStack(
        // Top: 20%
        ui.Spacer().Flex(2).Build(),

        ui.HStack(
            // Left: 30%
            ui.Spacer().Flex(3).Build(),

            // Modal 在 (20%, 30%)
            modal.NewBuilder().
                Title("20%, 30% Position").
                Content(app.Text("Positioned at 20% top, 30% left")).
                Centered(false).
                Build(),

            // Right: 70%
            ui.Spacer().Flex(7).Build(),
        ),

        // Bottom: 80%
        ui.Spacer().Flex(8).Build(),
    )
}
```

---

## 完整示例

### 示例 1：居中显示（最常用）

```go
modal.NewBuilder().
    Title("Centered Modal").
    Content(app.Text("This is the most common usage")).
    Width(50).
    Height(15).
    Center().  // 居中
    Build()
```

### 示例 2：左下角

```go
ui.VStack(
    // 占据大部分空间，将 Modal 推到底部
    ui.Spacer().SetFlex(9).Build(),

    ui.HStack(
        modal.NewBuilder().
            Title("Bottom Left").
            Content(app.Text("Bottom left corner")).
            Centered(false).
            Build(),

        // 占据右侧空间
        ui.Spacer().SetFlex(1).Build(),
    ),
)
```

### 示例 3：右上角

```go
ui.VStack(
    // 占据左侧空间，将 Modal 推到顶部
    ui.Spacer().SetFlex(1).Build(),

    ui.HStack(
        // 占据左侧空间，将 Modal 推到右侧
        ui.Spacer().SetFlex(1).Build(),

        modal.NewBuilder().
            Title("Top Right").
            Content(app.Text("Top right corner")).
            Centered(false).
            Build(),
    ),

    // 占据底部空间
    ui.Spacer().SetFlex(9).Build(),
)
```

### 示例 4：四角显示多个 Modal

```go
ui.VStack(
    // Top spacer
    ui.Spacer().SetFlex(5).Build(),

    ui.HStack(
        // Left spacer
        ui.Spacer().SetFlex(5).Build(),

        // Modal 1: 顶部居中
        modal.NewBuilder().
            Title("Top Center").
            Content(app.Text("Top center modal")).
            Centered(false).
            Build(),

        // Right spacer
        ui.Spacer().SetFlex(5).Build(),
    ),

    // Middle spacer (vertical)
    ui.Spacer().SetFlex(10).Build(),

    // Bottom modals row (if needed)
    ui.Spacer().SetFlex(5).Build(),
)
```

---

## 最佳实践

### ✅ DO

1. **默认使用居中** - 居中的 Modal 对用户体验最好
2. **使用 Spacer 控制位置** - Spacer 是控制相对定位的最佳方式
3. **考虑响应式布局** - 使用 Flex 比例而非固定像素
4. **保持一致性** - 应用内的 Modal 位置应保持一致

### ❌ DON'T

1. **不要使用硬编码坐标** - 避免在不同屏幕尺寸上出现问题
2. **不要让 Modal 过于靠近边缘** - 至少保留 2 个字符的边距
3. **不要堆叠多个 Modal** - 如果需要多个，使用垂直/水平布局容器
4. **不要忽视 Z-Index** - 层级高的组件会覆盖层级低的

---

## 常见问题

### Q: 如何让 Modal 固定在屏幕某个位置？
A: 当前使用 Spacer 组合可以实现。未来可能通过 `Position` 属性实现绝对定位。

### Q: 可以同时打开多个显示在不同位置的 Modal 吗？
A: 可以，为每个 Modal 设置不同的容器布局（使用 Spacer）。

### Q: 如何让 Modal 跟随某个按钮？
A: 将 Modal 和按钮放在同一个布局容器中，位置会自然对齐。

### Q: Centered 和 Spacer 可以同时使用吗？
A: 可以，但 Centered 优先级更高。建议二选一。

---

## 总结

| 定位方式 | 推荐场景 | 实现方式 |
|---------|---------|---------|
| **居中** | 大多数弹窗 | `.Center()` |
| **左对齐** | 紧贴左侧 | `.Centered(false)` + 左侧无 Spacer |
| **右对齐** | 紧贴右侧 | `.Centered(false)` + 右侧有 Spacer |
| **自定义位置** | 特定比例位置 | `.Centered(false)` + 四周 Spacer 组合 |

当前 Modal 组件通过 `centered` 属性和容器布局的组合，可以满足大多数定位需求。
