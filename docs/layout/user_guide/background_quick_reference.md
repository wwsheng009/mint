# 容器背景快速参考

**Container Background Quick Reference**

> 当前 API 提示：布局/背景行为仍有参考价值，但示例中如出现 `app.ButtonBuilder` 或其他 `app.*` 组件 builder，需要改写为当前 `ui.NewXBuilder` API。`app` 包现在主要作为 `app.Run` / `app.Quit` 兼容入口。

快速指南：如何在 Mint TUI 中使用容器背景。

---

## 基础用法

### 1. 设置容器背景

```go
import (
    rtui "github.com/wwsheng009/mint/runtime/ui"
    "github.com/wwsheng009/mint/runtime/style"
)

// 创建容器
container := rtui.VStack(
    ui.Text("内容 1"),
    ui.Text("内容 2"),
)

// 设置背景色
container.SetStyle(style.NewStyle().Background(style.Blue))
```

### 2. 带边框的容器

```go
content := rtui.VStack(
    ui.Text("标题"),
    ui.Text("内容"),
)

// 设置背景
content.SetStyle(style.NewStyle().Background(style.Blue))

// 包裹边框（重要：大小设置在外层）
panel := rtui.Bordered().
    Style(string(theme.Border())).
    Child(content).
    Width(40).    // ← 在外层设置大小
    Height(15).
    Build()
```

### 3. 子控件覆盖继承

```go
// 容器背景
container.SetStyle(style.NewStyle().Background(style.Blue))

// 子控件设置自己的背景（不继承）
button := app.ButtonBuilder("[按钮]").
    Style(style.NewStyle().Background(style.Red)).
    Build()
```

---

## 常用场景

### Modal 对话框

```go
func ShowModal() rtui.VNode {
    content := rtui.VStack(
        ui.Text("确认删除？"),
        ui.HStack(
            ui.Text("[确定]"),
            ui.Text("[取消]"),
        ),
    )

    // 深色背景 + 白色文字
    content.SetStyle(style.NewStyle().
        Background(style.Black).
        Foreground(style.White),
    )

    return rtui.Bordered().
        Style("double").
        Child(content).
        Width(40).
        Height(10).
        Build()
}
```

### Dropdown 菜单

```go
func ShowDropdown() rtui.VNode {
    menu := rtui.VStack(
        ui.Text("选项 1"),
        ui.Text("选项 2"),
        ui.Text("选项 3"),
    )

    // 白色背景 + 黑色文字
    menu.SetStyle(style.NewStyle().
        Background(style.White).
        Foreground(style.Black),
    )

    return menu
}
```

### Tooltip 提示

```go
func ShowTooltip(text string) rtui.VNode {
    tooltip := rtui.VStack(ui.Text(text))

    // 黄色背景 + 黑色文字
    tooltip.SetStyle(style.NewStyle().
        Background(style.Yellow).
        Foreground(style.Black),
    )

    return tooltip
}
```

### Inspector 面板

```go
func InspectorPanel() rtui.VNode {
    header := rtui.VStack(
        ui.Text("╔═ INSPECTOR ═╗"),
        ui.Text("F12:关闭"),
    )

    content := rtui.VStack(
        header,
        ui.Text("─"),
        buildTree(),
    )

    // 蓝色背景
    content.SetStyle(style.NewStyle().Background(style.Blue))

    return rtui.Bordered().
        Style(string(theme.Border())).
        Child(content).
        Width(38).
        Height(20).
        Build()
}
```

---

## 颜色方案

### 深色主题

```go
// 黑色背景 + 白色文字
style.NewStyle().
    Background(style.Black).
    Foreground(style.White)

// 深蓝背景 + 浅色文字
style.NewStyle().
    Background(style.Blue).
    Foreground(style.White)
```

### 浅色主题

```go
// 白色背景 + 黑色文字
style.NewStyle().
    Background(style.White).
    Foreground(style.Black)

// 浅灰背景 + 深色文字
style.NewStyle().
    Background(style.BrightWhite).
    Foreground(style.Black)
```

### 强调色

```go
// 红色背景（错误/警告）
style.NewStyle().
    Background(style.Red).
    Foreground(style.Yellow)

// 绿色背景（成功）
style.NewStyle().
    Background(style.Green).
    Foreground(style.White)

// 黄色背景（提示）
style.NewStyle().
    Background(style.Yellow).
    Foreground(style.Black)
```

---

## 常见陷阱

### ❌ 错误 1：大小设置在内层

```go
// 错误：会导致外层无边框大小限制
content := rtui.VStackBuilder(...).
    Width(40).  // 错误位置
    Build()
panel := rtui.Bordered().Child(content).Build()
```

### ✅ 正确 1：大小设置在外层

```go
// 正确：外层边框限制大小
content := rtui.VStack(...)
content.SetStyle(style.NewStyle().Background(style.Blue))
panel := rtui.Bordered().
    Child(content).
    Width(40).    // 正确位置
    Height(15).
    Build()
```

### ❌ 错误 2：忘记设置背景

```go
// 创建容器但没有设置背景
container := rtui.VStack(...)
// 忘记：container.SetStyle(...)
// 结果：透明背景，看不到效果
```

### ✅ 正确 2：明确设置背景

```go
container := rtui.VStack(...)
container.SetStyle(style.NewStyle().Background(style.Blue))
```

### ❌ 错误 3：前景背景冲突

```go
// 深色背景 + 深色文字 = 看不清
style.NewStyle().
    Background(style.Black).
    Foreground(style.Blue)  // 可读性差
```

### ✅ 正确 3：高对比度

```go
// 深色背景 + 浅色文字 = 易读
style.NewStyle().
    Background(style.Black).
    Foreground(style.White)  // 高对比度
```

---

## 调试技巧

### 启用背景渲染调试

```bash
# 查看背景渲染过程
TUI_PAINT_DEBUG=true ./your_app

# 输出示例：
# [Paint.paintContainerBackground] Occluded 40x15 area at (10, 5) with BG=blue
# [Paint.paintNode]   🎨 Inherited parent BG=blue
```

### 检查背景是否生效

```go
// 在代码中检查
vnode.SetStyle(style.NewStyle().Background(style.Blue))

// 验证
if vnode.Style().BG == style.Blue {
    fmt.Println("✅ 背景设置成功")
} else {
    fmt.Println("❌ 背景设置失败")
}
```

### 可视化背景区域

```go
// 临时方法：用显眼的颜色测试
container.SetStyle(style.NewStyle().
    Background(style.Red).  // 使用红色便于观察
    Foreground(style.Yellow),
)
```

---

## 性能提示

### ✅ 好的做法

```go
// 静态颜色（编译时确定）
container.SetStyle(style.NewStyle().Background(style.Blue))

// 设置一次，不频繁更改
container.SetStyle(style.NewStyle().Background(style.Blue))
// 之后不再修改
```

### ❌ 差的做法

```go
// 动态颜色（运行时计算）
bg := calculateBackgroundColor()
container.SetStyle(style.NewStyle().Background(bg))

// 频繁更改（每帧都换）
for {
    bg := getUserPreference()
    container.SetStyle(style.NewStyle().Background(bg))
}
```

---

## 快速检查清单

使用容器背景时，检查以下项目：

- [ ] 容器大小设置正确（外层边框）
- [ ] 背景色已设置（`SetStyle(Background(...))`）
- [ ] 前景背景对比度足够
- [ ] 子控件不需要自己的背景（会自动继承）
- [ ] 不在循环中频繁更改背景
- [ ] 使用合适的颜色方案
- [ ] 测试了在不同终端下的效果

---

## 相关文档

- **完整文档**: `container_background_rendering.md`
- **架构**: `LAYOUT_RENDERING_REFACTOR.md`
- **Layer 系统**: `layer_system_guide.md`
- **Flex 布局**: `flex_layout.md`

---

**版本**: 1.0
**更新**: 2025-02-08
**作者**: Mint TUI Team
