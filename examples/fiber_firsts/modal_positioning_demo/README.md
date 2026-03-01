# Modal Positioning Demo

演示如何控制 Modal 的显示位置。

---

## 示例程序

本目录包含两个示例程序：

1. **交互式演示** (`main.go`) - 点击按钮切换显示不同位置的 Modal
2. **直接渲染测试** (`direct_render.go`) - 同时渲染所有 Modal，用于快速检查位置

---

## 功能演示

这个示例展示了 6 种不同的 Modal 定位方式：

1. **居中显示（Center）** - 默认且最常用的定位方式
2. **左对齐（Left Aligned）** - 使用无左侧 Spacer 的相对定位
3. **右对齐（Right Aligned）** - 使用左侧 Spacer 推到右侧
4. **顶部对齐（Top Aligned）** - 使用无顶部 Spacer 的相对定位
5. **底部对齐（Bottom Aligned）** - 使用顶部 Spacer 推到底部
6. **自定义位置（Custom 30%, 20%）** - 使用精确的 Spacer 比例控制位置

---

## 运行方式

### 交互式演示 (`main.go`)

#### 方式 1: 使用 Go 直接运行

```bash
cd examples/fiber_firsts/modal_positioning_demo
go run main.go
```

#### 方式 2: 编译后运行

```bash
# 从项目根目录编译
go build -o bin/modal_positioning_demo.exe examples/fiber_firsts/modal_positioning_demo/main.go

# 运行
bin\modal_positioning_demo.exe
```

### 直接渲染测试 (`direct_render.go`)

由于直接渲染测试会同时显示 6 个 Modal，这个示例更适合用于：

1. **检查渲染位置是否正确**
2. **验证布局系统的工作方式**
3. **快速迭代调试**

运行方式：

```bash
cd examples/fiber_firsts/modal_positioning_demo
go run direct_render.go
```

或编译后运行：

```bash
# 从项目根目录编译
go build -o bin/direct_render_test.exe examples/fiber_firsts/modal_positioning_demo/direct_render.go

# 运行
bin\direct_render_test.exe
```

---

## 使用方法

1. 启动程序后，会看到一个主菜单
2. 点击任意按钮显示对应位置的 Modal
3. 状态栏会显示当前 Modal 的位置方式和说明
4. 按 ESC 键或点击 Modal 外部区域关闭 Modal
5. 可以尝试不同的位置方式看看效果

---

## 代码示例

### 居中显示（默认）

```go
modal.NewBuilder().
    Title("Centered Modal").
    Content(app.Text("This is centered")).
    Width(50).
    Height(12).
    Center().  // 居中显示
    Build()
```

### 左对齐

```go
ui.HStack(
    modal.NewBuilder().
        Title("Left Aligned").
        Centered(false).  // 关闭居中
        Build(),
    ui.Spacer().Build(),  // Spacer 推到左侧
)
```

### 右对齐

```go
ui.HStack(
    ui.Spacer().Build(),  // Spacer 推到右侧
    modal.NewBuilder().
        Title("Right Aligned").
        Centered(false).  // 关闭居中
        Build(),
)
```

### 自定义位置 (30% 左, 20% 上)

```go
ui.VStack(
    ui.Spacer().Flex(2).Build(),  // Top: 20%
    ui.HStack(
        ui.Spacer().Flex(3).Build(),  // Left: 30%
        modal.NewBuilder().
            Title("Custom Position").
            Centered(false).
            Build(),
        ui.Spacer().Flex(7).Build(),  // Right: 70%
    ),
    ui.Spacer().Flex(8).Build(),  // Bottom: 80%
)
```

---

## 定位原理

### Spacer 比例计算

使用 Spacer 的 `Flex` 属性可以自定义百分比位置：

```
水平位置:
  left%  = leftSpacer / (leftSpacer + rightSpacer)
  right% = rightSpacer / (leftSpacer + rightSpacer)

垂直位置:
  top%    = topSpacer / (topSpacer + bottomSpacer)
  bottom% = bottomSpacer / (topSpacer + bottomSpacer)
```

### 示例：位置在 (30%, 20%)

- Vertical: Top spacer = 2, Bottom spacer = 8
  - `2 / (2 + 8) = 2 / 10 = 20%` (顶部)

- Horizontal: Left spacer = 3, Right spacer = 7
  - `3 / (3 + 7) = 3 / 10 = 30%` (左侧)

---

## 技术细节

### Intent 驱动

这个示例使用 Intent 来管理 Modal 的开关状态：

```go
type ShowPositioningIntent struct {
    Position string  // "center", "left", "right", etc.
}

type ClosePositioningIntent struct{}
```

### 状态管理

使用 `ui.UseStateString()` 管理 Modal 的打开状态：

```go
position, setPosition := ui.UseStateString("")
```

---

## 相关文档

更多关于 Modal 位置控制的详细信息，请参考：

- [Modal 位置控制指南](../../../../ui/components/modal/POSITIONING.md)
- [Modal 组件文档](../../../../ui/components/modal/README.md)
- [插件架构文档](../../../../ui/components/modal/PLUGIN_ARCHITECTURE.md)

---

## 最佳实践

### ✅ DO

1. **默认使用居中** - 这是用户体验最好的方式
2. **使用 Spacer 控制位置** - Spacer 是控制相对定位的最佳方式
3. **保持比例一致** - 使用 Flex 比例而非固定像素
4. **考虑响应式** - 布局会自动适应不同的终端尺寸

### ❌ DON'T

1. **不要硬编码坐标** - 避免在不同终端尺寸上出现问题
2. **不要让 Modal 过于靠近边缘** - 至少保留 2 个字符的边距
3. **不要忽视键盘操作** - ESC 键是关闭 Modal 的标准方式

---

## 键盘快捷键

| 按键 | 功能 |
|------|------|
| ESC | 关闭当前打开的 Modal |

---

## 故障排除

### Modal 没有显示

- 检查是否设置了 `Open(true)`
- 确认 `WithPluginSetup` 中注册了 Modal 中间件

### 位置不正确

- 检查 `Centered(false)` 是否正确设置
- 验证 Spacer 的 `Flex` 比例是否正确

### 无法用 ESC 关闭

- 确认 `app.AddMiddleware(modal.NewModalMiddleware())` 已调用
- 检查 Modal 是否正确注册到全局注册表
