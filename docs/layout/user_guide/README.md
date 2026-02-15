# 布局系统文档索引

**Layout System Documentation Index**

欢迎来到 Mint TUI 布局系统文档中心！

---

## 📚 快速导航

### 入门文档

| 文档 | 描述 | 适合读者 |
|------|------|---------|
| **[背景快速参考](./background_quick_reference.md)** | 容器背景的快速上手指南 | 所有人 |
| **[Wrap 速查表](./wrap_cheatsheet.md)** | Wrap 组件的快速参考 | 所有开发者 |
| **[Flex 对比](./flex_comparison.md)** | Mint TUI vs CSS Flexbox | 有 Web 开发经验者 |

### 核心概念

| 文档 | 描述 | 重要性 |
|------|------|--------|
| **[容器背景渲染](./container_background_rendering.md)** | 容器背景、遮挡、继承系统 | ⭐⭐⭐ |
| **[Flex 布局](./flex_layout.md)** | Mint TUI 布局系统详解 | ⭐⭐⭐ |
| **[Layer 系统](./layer_system_guide.md)** | 渲染层级和覆盖层 | ⭐⭐ |
| **[Wrap 组件](./wrap_component.md)** | 自动换行布局组件 | ⭐⭐⭐ |

### 高级主题

| 文档 | 描述 | 难度 |
|------|------|------|
| **[Stretch 布局](./stretch_layout.md)** | FillWidth/FillHeight 机制 | ⭐⭐⭐ |
| **[Layer 架构](./LAYER_LAYOUT_ARCHITECTURE_REVIEW.md)** | Layer 系统架构分析 | ⭐⭐⭐⭐ |
| **[Constraints](./getChildconstraints_architecture_analysis.md)** | 约束传递机制 | ⭐⭐⭐⭐ |

### 常见问题

| 文档 | 描述 |
|------|------|
| **[Flex Wrap 限制](./flex_wrap_limitation.md)** | 为什么不支持 flex-wrap，如何解决 |
| **[Layer 约束审计](./layer_constraint_audit_report.md)** | Layer 系统约束问题分析 |

### 重构历史

| 文档 | 描述 |
|------|------|
| **[布局重构](./layout_refactor.md)** | 布局系统重构历史 |
| **[渲染重构](./LAYOUT_RENDERING_REFACTOR.md)** | 渲染管线重构 |
| **[单次重构](./single_pass_refactor_summary.md)** | 单次布局重构总结 |

---

## 🚀 按场景查找

### 我想实现...

#### 自动换行布局

→ 查看 **[Wrap 组件](./wrap_component.md)**

```go
wrapped := app.WrapBuilder(buttons...).
    Gap(1).
    ScreenWidth(80).
    Build()
```

#### 容器有背景色

→ 查看 **[容器背景渲染](./container_background_rendering.md)**

```go
container.SetStyle(style.NewStyle().Background(style.Blue))
```

#### 元素居中对齐

→ 查看 **[Flex 布局](./flex_layout.md)**

```go
ui.HStackBuilder(...).
    Align(ui.AlignCenter).
    Build()
```

#### 横向填充剩余空间

→ 查看 **[Stretch 布局](./stretch_layout.md)**

```go
ui.Text("").FillWidth().Build()
```

#### 浮动面板/覆盖层

→ 查看 **[Layer 系统](./layer_system_guide.md)**

```go
vnode.SetLayer(rtui.LayerOverlay)
```

#### Inspector 调试面板

→ 查看 **[Layer 系统指南](./layer_system_guide.md#inspector)**

---

## 📖 阅读路径

### 初学者路径

1. **快速开始**: [背景快速参考](./background_quick_reference.md)
2. **基础布局**: [Flex 对比](./flex_comparison.md)
3. **实践示例**: [Wrap 速查表](./wrap_cheatsheet.md)

### 进阶路径

1. **深入理解**: [Flex 布局](./flex_layout.md)
2. **容器背景**: [容器背景渲染](./container_background_rendering.md)
3. **Layer 系统**: [Layer 指南](./layer_system_guide.md)
4. **高级布局**: [Stretch 布局](./stretch_layout.md)

### 专家路径

1. **架构分析**: [Layer 架构](./LAYER_LAYOUT_ARCHITECTURE_REVIEW.md)
2. **约束系统**: [Constraints](./getChildconstraints_architecture_analysis.md)
3. **重构历史**: [布局重构](./layout_refactor.md)
4. **渲染管线**: [渲染重构](./LAYOUT_RENDERING_REFACTOR.md)

---

## 🛠️ 常用代码片段

### 1. 带背景的容器

```go
content := rtui.VStack(
    ui.Text("标题"),
    ui.Text("内容"),
)
content.SetStyle(style.NewStyle().Background(style.Blue))

panel := rtui.Bordered().
    Child(content).
    Width(40).
    Height(15).
    Build()
```

### 2. 自动换行的按钮网格

```go
wrapped := app.WrapBuilder(buttons...).
    Gap(1).
    RowGap(0).
    ScreenWidth(80).
    Align(ui.AlignStart).
    Build()
```

### 3. 居中对齐的内容

```go
centered := ui.VStackBuilder(
    ui.Text("标题"),
    ui.Text("内容"),
).
    Align(ui.AlignCenter).  // 主轴居中
    AlignContent(ui.AlignCenter).  // 交叉轴居中
    Build()
```

### 4. 填充剩余空间

```go
ui.HStack(
    ui.Text("左侧"),
    ui.Text("").FillWidth(),  // 填充中间
    ui.Text("右侧"),
)
```

### 5. 固定大小的容器

```go
fixedBox := rtui.Bordered().
    Child(content).
    Width(40).
    Height(15).
    Build()
```

---

## 🔍 问题排查

### 容器拉伸到整个屏幕

**原因**: 大小设置在内层，外层无边框大小限制

**解决**: 将 `Width()`/`Height()` 移到外层 `Bordered()`

```go
// ❌ 错误
content := rtui.VStackBuilder(...).Width(40).Build()
panel := rtui.Bordered().Child(content).Build()

// ✅ 正确
content := rtui.VStack(...)
content.SetStyle(style.NewStyle().Background(style.Blue))
panel := rtui.Bordered().Child(content).Width(40).Build()
```

### 子控件背景不一致

**原因**: 子控件没有继承父容器背景

**解决**: 子控件会自动继承，无需手动设置

```go
// 父容器设置背景
container.SetStyle(style.NewStyle().Background(style.Blue))
// 子控件自动继承蓝色背景
```

### 内容超出容器边界

**原因**: 使用了 `HStack` 而不是 `Wrap`

**解决**: 使用 `Wrap` 组件实现自动换行

```go
// ❌ 单行，可能超出
ui.HStack(items...)

// ✅ 自动换行
app.WrapBuilder(items...).ScreenWidth(80).Build()
```

---

## 📊 文档统计

| 类别 | 数量 | 总字数 |
|------|------|--------|
| 入门文档 | 3 | ~30,000 |
| 核心概念 | 4 | ~50,000 |
| 高级主题 | 3 | ~40,000 |
| 常见问题 | 2 | ~15,000 |
| 重构历史 | 3 | ~45,000 |
| **总计** | **15** | **~180,000** |

---

## 🤝 贡献指南

### 如何改进文档？

1. **发现错误**: 提交 Issue 或 PR
2. **添加示例**: 在相关文档中添加代码示例
3. **完善说明**: 补充遗漏的概念或细节
4. **翻译文档**: 提供其他语言版本

### 文档规范

- 使用 Markdown 格式
- 代码示例可运行
- 添加适当的表情符号 📚
- 保持目录结构清晰
- 及时更新版本号

---

## 🔗 外部资源

### CSS Flexbox

- [MDN: Flexbox](https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_Flexible_Box_Layout)
- [CSS-Tricks: A Complete Guide to Flexbox](https://css-tricks.com/snippets/css/a-guide-to-flexbox/)

### TUI 相关

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - Elm 架构的 TUI 框架
- [tcell](https://github.com/gdamore/tcell) - TUI 终端抽象层
- [lipgloss](https://github.com/charmbracelet/lipgloss) - TUI 样式库

---

## 📝 更新日志

### 2025-02-08

- ✅ 添加容器背景渲染系统文档
- ✅ 添加背景快速参考
- ✅ 更新 flex_wrap_limitation.md
- ✅ 创建文档索引 README

### 更早版本

- 详见各个文档的更新历史

---

## 📧 联系方式

- **问题反馈**: GitHub Issues
- **功能请求**: GitHub Discussions
- **文档 PR**: 欢迎提交

---

**维护者**: Mint TUI Team
**最后更新**: 2025-02-08
**版本**: 1.0
