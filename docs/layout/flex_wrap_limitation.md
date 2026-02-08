# 布局系统限制：flex-wrap 自动换行

## 当前状态

Mint TUI 的布局系统**不支持** CSS `flex-wrap: wrap` 功能。

## 问题场景

当子元素总宽度超过父容器宽度时：

### CSS Flexbox (支持换行)
```css
.container {
  display: flex;
  flex-wrap: wrap;  /* ✅ 自动换行 */
  gap: 8px;
}
```

### Mint TUI (不支持换行)
```go
ui.HStackBuilder(
    item1, item2, item3, // ... many items
).
    Gap(1).
    Build()
// ❌ 如果总宽度超过容器，会溢出或被截断
```

## demo2 当前实现

所有按钮在一行显示：

```go
allButtons := ui.HStackBuilder(
    "[1] Event", "[2] setState", "[3] Scheduler",
    "[4] Render", "[5] Reconcile", "[6] Layout",
    "[7] Paint", "[0] Idle", // 8 个按钮
).
    Gap(1).
    Align(ui.AlignStart).
    Build()
```

**预期宽度计算：**
- 每个按钮约 15-20 字符
- 7 个 gap × 1 字符 = 7 字符
- 边框：2 字符
- 总计：约 130-170 字符

**问题：** 在 100 字符宽度的终端中，按钮会超出边界。

## 解决方案

### 方案 1: 手动分行（推荐）

最简单、最可靠的方法：

```go
row1 := ui.HStackBuilder(buttons[0:4]).Gap(1).Align(ui.AlignStart).Build()
row2 := ui.HStackBuilder(buttons[4:8]).Gap(1).Align(ui.AlignStart).Build()

ui.VStack(row1, row2).Gap(0).Build()
```

**优点：**
- ✅ 完全控制每行的内容
- ✅ 可预测的布局
- ✅ 适配不同屏幕宽度

**缺点：**
- ❌ 需要手动计算分行
- ❌ 固定分行，不响应宽度变化

---

### 方案 2: 使用 Wrap 组件（已实现） ✅

**好消息！** Mint TUI 现在提供了 `Wrap` 组件，支持自动换行。

详细文档：[Wrap Component](./wrap_component.md)

**使用示例：**

```go
import (
    "github.com/wwsheng009/mint/app"
    ui "github.com/wwsheng009/mint/ui"
)

buttons := []ui.VNode{
    app.ButtonBuilder("[1] Event").Build(),
    app.ButtonBuilder("[2] State").Build(),
    app.ButtonBuilder("[3] Scheduler").Build(),
    app.ButtonBuilder("[4] Render").Build(),
    app.ButtonBuilder("[5] Reconcile").Build(),
    app.ButtonBuilder("[6] Layout").Build(),
    app.ButtonBuilder("[7] Paint").Build(),
    app.ButtonBuilder("[0] Idle").Build(),
}

wrapped := app.WrapBuilder(buttons...).
    Gap(1).
    RowGap(0).
    ScreenWidth(98).  // 100 - border (2)
    Align(ui.AlignStart).
    Build()
```

**优点：**
- ✅ 自动适配屏幕宽度
- ✅ 响应式布局
- ✅ 类似 CSS flex-wrap
- ✅ 基于现有的 HStack/VStack，性能好
- ✅ 支持多种对齐方式
- ✅ 智能宽度估算（Button, Text, Input）

**使用场景：**
- 按钮网格/控制面板
- 标签云
- 响应式导航
- 任何需要自动换行的布局

---

### 方案 3: 使用固定宽度容器

限制每个按钮的宽度，确保能放下：

```go
ui.HStackBuilder(
    btn1.Width(12).Build(),
    btn2.Width(12).Build(),
    btn3.Width(12).Build(),
    // ...
).
    Gap(1).
    Build()
```

**计算：**
- 8 个按钮 × 12 字符 = 96 字符
- 7 个 gap × 1 字符 = 7 字符
- 边框 = 2 字符
- 总计 = 105 字符（略微超出 100 字符）

**问题：** 即使固定宽度，仍然可能超出。

---

### 方案 4: 滚动容器

如果按钮太多，使用滚动：

```go
// 类似 CSS overflow-x: auto
scrollable := ui.Scrollable().Content(allButtons).Build()
```

**问题：** 终端 UI 的滚动实现复杂，性能开销大。

---

## 推荐做法

### ✅ 推荐方案：使用 Wrap 组件

**最佳实践：**

```go
// 自动换行，适配不同终端宽度
wrapped := app.WrapBuilder(buttons...).
    Gap(1).
    RowGap(0).
    ScreenWidth(availableWidth).
    Align(ui.AlignStart).
    Build()
```

**优点：**
- 简单、可靠、性能好
- 自动适配终端宽度（80-120 字符）
- 易于维护和调试
- 响应式布局

详细使用指南：[Wrap Component Documentation](./wrap_component.md)

---

### 备选方案：手动分行（特殊情况）

如果需要精确控制每行内容，可以手动分行：

```go
// 根据终端宽度合理分组
row1 := ui.HStackBuilder(btn1, btn2, btn3, btn4).Gap(1).Build()
row2 := ui.HStackBuilder(btn5, btn6, btn7, btn8).Gap(1).Build()

ui.VStack(row1, row2).Gap(0).Build()
```

**适用场景：**
- 需要精确控制每行内容
- 功能性分组（如将相关按钮放在同一行）
- 固定布局需求

---

### 长期（架构改进）

**参考 CSS Flexbox，实现完整的 flex-wrap 支持：**

```go
ui.HStackBuilder(items...).
    FlexWrap(ui.WrapWrap). // 类似 flex-wrap: wrap
    AlignContent(ui.AlignStart).
    Build()
```

**需要的扩展：**
1. `FlexWrap` 属性
2. `AlignContent` 属性（控制多行对齐）
3. 布局引擎支持多行计算
4. 动态行高计算

---

## 当前 demo2 的建议

### 选项 1: 保持单行（当前）

```go
allButtons := ui.HStackBuilder(
    btn1, btn2, btn3, btn4, btn5, btn6, btn7, btn8,
).
    Gap(1).
    Align(ui.AlignStart).
    Build()
```

**适用场景：** 终端宽度 ≥ 120 字符

**问题：** 在 80-100 字符终端会溢出

---

### 选项 2: 手动分行（稳妥）

```go
row1 := ui.HStackBuilder(btn1, btn2, btn3, btn4).Gap(1).Build()
row2 := ui.HStackBuilder(btn5, btn6, btn7, btn8).Gap(1).Build()
content := ui.VStack(row1, row2).Gap(0).Build()
```

**适用场景：** 所有终端宽度

**优点：** 不会溢出，布局稳定

---

### 选项 3: 按功能分组

```go
row1 := ui.HStackBuilder(
    // Pipeline phases
    btn1, btn2, btn3, btn4,
).Gap(1).Build()

row2 := ui.HStackBuilder(
    // More phases + idle
    btn5, btn6, btn7, btn8,
).Gap(1).Build()
```

**优点：**
- 逻辑分组清晰
- 每行功能相关

---

## 技术限制总结

| 特性 | CSS Flexbox | Mint TUI | 说明 |
|------|-------------|----------|------|
| `flex-wrap: nowrap` | ✅ 默认 | ✅ 默认 | 单行，不换行 |
| `flex-wrap: wrap` | ✅ 支持 | ✅ **支持** | 使用 `Wrap` 组件 |
| `flex-wrap: wrap-reverse` | ✅ 支持 | ❌ 不支持 | - |
| `align-content` | ✅ 支持 | ⚠️ 部分 | 支持 per-row alignment |
| 响应式宽度 | ✅ 媒体查询 | ⚠️ 有限 | 终端宽度固定 |

**解决方案：**
- 自动换行：使用 `app.WrapBuilder()` - [查看文档](./wrap_component.md)
- 手动分行：使用 `HStack` + `VStack` 组合

---

## 总结

**当前最佳实践（2024+）：**

1. ✅ 使用 `HStackBuilder` + `Gap()` 代替手动空格
2. ✅ 使用 `Align(ui.AlignStart)` 从左到右排列
3. ✅ 使用 `FillWidth()` 让容器横向拉伸
4. ✅ **使用 `Wrap` 组件实现自动换行** - [查看文档](./wrap_component.md)
5. ✅ 自动适配不同终端宽度（80-120 字符）

**使用建议：**

| 场景 | 推荐方案 | 示例 |
|------|---------|------|
| 自动换行 | `app.WrapBuilder()` | 按钮网格、标签云 |
| 精确控制 | 手动 `HStack` + `VStack` | 固定布局 |
| 单行布局 | `HStackBuilder()` | 导航栏 |

**相关文档：**
- [Wrap Component](./wrap_component.md) - 完整的使用指南和 API 参考
- [Flex Layout Comparison](./flex_layout.md) - Flexbox 对比
- [Layout Best Practices](../guide/layout.md) - 布局最佳实践

---

## 相关资源

- **CSS Flexbox:** https://developer.mozilla.org/en-US/docs/Web/CSS/flex-wrap
- **Mint TUI 布局系统:** `docs/layout/flex_comparison.md`
- **布局问题修复:** `docs/layout/issue/`

---

## 附录：容器背景渲染系统

**更新日期**: 2025-02-08

Mint TUI 现在支持完整的容器背景渲染系统，包括：

### 功能特性

1. **容器背景渲染** - 为整个容器设置统一的背景色
2. **内容遮挡** - 自动遮挡底层内容，避免"透视"效果
3. **背景继承** - 子控件自动继承父容器背景，确保视觉一致

### 快速示例

```go
import (
    rtui "github.com/wwsheng009/mint/runtime/ui"
    "github.com/wwsheng009/mint/runtime/style"
)

// 创建带背景的容器
content := rtui.VStack(
    ui.Text("内容 1"),
    ui.Text("内容 2"),
    ui.Text("内容 3"),
)

// 设置背景色（内部所有子控件自动继承）
content.SetStyle(style.NewStyle().Background(style.Blue))

// 包裹边框并设置大小
panel := rtui.Bordered().
    Style(string(theme.Border())).
    Child(content).
    Width(40).    // 重要：大小设置在外层
    Height(15).
    Build()
```

### 使用场景

- **Modal 对话框** - 深色背景遮挡底层应用
- **Inspector 面板** - 蓝色背景便于区分
- **Dropdown 菜单** - 白色背景突出选项
- **Tooltip 提示** - 黄色背景吸引注意

### 相关文档

- **完整文档**: [Container Background Rendering](./container_background_rendering.md)
- **快速参考**: [Background Quick Reference](./background_quick_reference.md)
- **实现细节**:
  - `INSPECTOR_CONTAINER_BACKGROUND_FIX.md` - 容器背景修复
  - `INSPECTOR_OCCLUSION_FIX.md` - 遮挡透视修复
  - `INSPECTOR_BACKGROUND_INHERITANCE.md` - 背景继承机制
  - `INSPECTOR_BACKGROUND_COMPLETE_SOLUTION.md` - 完整解决方案

### 调试支持

```bash
# 查看背景渲染过程
TUI_PAINT_DEBUG=true ./your_app

# 输出示例：
# [Paint.paintContainerBackground] Occluded 40x15 area at (10, 5) with BG=blue
# [Paint.paintNode]   🎨 Inherited parent BG=blue
```

### 最佳实践

| 场景 | 背景色 | 前景色 | 用途 |
|------|--------|--------|------|
| Modal 对话框 | `style.Black` | `style.White` | 深色主题 |
| Dropdown 菜单 | `style.White` | `style.Black` | 浅色主题 |
| Inspector 面板 | `style.Blue` | `style.White` | 强调区分 |
| 错误提示 | `style.Red` | `style.Yellow` | 警告强调 |
| 成功提示 | `style.Green` | `style.White` | 成功反馈 |

---

**回到**: [布局系统限制](#当前状态) | [目录](#目录)
