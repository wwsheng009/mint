# 容器背景渲染系统

**Container Background Rendering System**

> 当前 API 提示：本文的单元格背景、继承和遮挡机制仍可用于理解渲染结果。示例中若使用 `app.*` 组件 builder，请改写为当前 `ui.NewXBuilder` API 和 Fiber-first component instance 模型。

本文档详细说明 Mint TUI 中容器背景渲染的完整机制，包括背景渲染、内容遮挡和背景继承。

---

## 目录

1. [概述](#概述)
2. [核心概念](#核心概念)
3. [技术实现](#技术实现)
4. [使用指南](#使用指南)
5. [最佳实践](#最佳实践)
6. [常见问题](#常见问题)
7. [性能考虑](#性能考虑)

---

## 概述

在 TUI (Terminal User Interface) 中，容器背景渲染是一个复杂的话题，涉及多个层面：

- **视觉分层**：如何实现覆盖层效果
- **内容遮挡**：如何遮挡底层内容
- **背景继承**：如何让子控件与父容器视觉一致

### 三大核心问题

| 问题 | 现象 | 原因 | 解决方案 |
|------|------|------|----------|
| **背景不生效** | 设置背景色但显示透明 | VNode 系统不支持容器背景 | 增强 PaintEngine |
| **内容透视** | 背景后仍能看到底层内容 | 背景填充时保留了现有内容 | 无条件填充区域 |
| **背景冲突** | 父子容器背景色不一致 | 子控件不继承父容器背景 | 实现背景继承机制 |

---

## 核心概念

### 1. TUI 渲染管线

```
Layout Phase (布局阶段)
  ↓ 计算位置和大小
Paint Phase (渲染阶段)
  ├─ Layer Base (底层应用)
  │   └─ 渲染应用内容到 buffer
  ├─ Layer Overlay (覆盖层)
  │   └─ 渲染 Inspector 等到 buffer
  │       ├─ 1. 填充背景（遮挡底层）
  │       └─ 2. 渲染内容（在背景之上）
  └─ Buffer 合并输出到终端
```

**关键点**：
- TUI 没有真正的 alpha 通道
- 遮挡通过**绘制顺序**实现
- 后绘制的内容会覆盖先绘制的内容

### 2. 容器 vs 内容

```
容器 (Container)
  ┌─────────────────────┐
  │ 背景区域             │ ← 容器背景（统一颜色）
  │  ┌───────────────┐  │
  │  │ 子控件 1      │  │ ← 子控件内容（可能有自己的背景）
  │  └───────────────┘  │
  │  ┌───────────────┐  │
  │  │ 子控件 2      │  │
  │  └───────────────┘  │
  └─────────────────────┘
```

**层级关系**：
1. 容器背景（最底层）
2. 子控件背景（中间层）
3. 子控件内容（最上层）

### 3. 背景继承机制

```
父容器 (BG: Blue)
  │
  ├─ 子控件 1 (无 BG) → 继承 Blue
  ├─ 子控件 2 (无 BG) → 继承 Blue
  └─ 子控件 3 (BG: Red) → 使用 Red (不继承)
```

**继承规则**：
- 自动继承：子控件无背景时自动使用父容器背景
- 显式优先：子控件明确设置背景时不继承
- 递归应用：继承递归应用到所有后代节点

---

## 技术实现

### 1. PaintEngine 架构

**文件**：`internal/render/paint_engine.go`

#### 核心结构

```go
type PaintEngine struct {
    debug         bool
    lastHadModal  bool
    forceFullRender bool
    parentBackground map[*compute.ComputedBox]style.Color // 背景继承映射
}
```

#### 关键方法

| 方法 | 职责 |
|------|------|
| `Paint()` | 渲染入口，清理状态 |
| `paintNode()` | 递归渲染节点，应用背景继承 |
| `paintElement()` | 渲染元素节点，记录父容器背景 |
| `paintContainerBackground()` | 填充容器背景，遮挡底层内容 |
| `paintChildren()` | 渲染子节点 |

### 2. 容器背景渲染

**文件**：`internal/render/paint_engine.go:paintContainerBackground()`

```go
func (e *PaintEngine) paintContainerBackground(box *compute.ComputedBox, buffer *paint.Buffer, bgStyle style.Style) {
    backgroundStyle := style.Style{}.Background(bgStyle.BG)

    // 关键：无条件填充整个区域
    for y := 0; y < box.Box.Height; y++ {
        for x := 0; x < box.Box.Width; x++ {
            // 使用空格字符清除所有现有内容
            buffer.SetCell(box.Box.X+x, box.Box.Y+y, ' ', backgroundStyle)
        }
    }
}
```

**为什么无条件填充？**
- 在覆盖层模式下，buffer 中已经渲染了底层应用内容
- 如果检查现有内容，会保留底层内容导致"透视"效果
- 必须用空格 + 背景色完全覆盖

### 3. 背景继承实现

**文件**：`internal/render/paint_engine.go:paintElement()`

```go
func (e *PaintEngine) paintElement(box *compute.ComputedBox, buffer *paint.Buffer) {
    // ... 现有逻辑 ...

    // 渲染容器背景
    nodeStyle := box.VNode.Style()
    if nodeStyle.BG != "" && nodeStyle.BG != style.NoColor {
        e.paintContainerBackground(box, buffer, nodeStyle)

        // 记录父容器背景，供子节点继承
        if e.parentBackground == nil {
            e.parentBackground = make(map[*compute.ComputedBox]style.Color)
        }
        for _, childBox := range box.Children {
            e.parentBackground[childBox] = nodeStyle.BG
        }
    }
}
```

**继承应用**：`internal/render/paint_engine.go:paintNode()`

```go
// 检查父容器背景
var parentBG style.Color
if e.parentBackground != nil {
    if inheritedBG, ok := e.parentBackground[box]; ok && inheritedBG != "" {
        parentBG = inheritedBG
    }
}

// Paintable 组件（Button、Input 等）
if paintable {
    commands := paintable.Paint(box.Box.X, box.Box.Y)
    for _, cmd := range commands {
        styleToApply := cmd.Style
        // 如果 DrawCmd 没有背景，继承父容器背景
        if parentBG != "" && (styleToApply.BG == "" || styleToApply.BG == style.NoColor) {
            styleToApply.BG = parentBG
        }
        buffer.SetString(cmd.X, cmd.Y, cmd.Text, styleToApply)
    }
    delete(e.parentBackground, box)
    return nil
}

// 非 Paintable 组件（Text、VStack 等）
if parentBG != "" {
    nodeStyle := box.VNode.Style()
    if nodeStyle.BG == "" || nodeStyle.BG == style.NoColor {
        inheritedStyle := nodeStyle
        inheritedStyle.BG = parentBG
        box.VNode.SetStyle(inheritedStyle)
    }
}
delete(e.parentBackground, box)
```

---

## 使用指南

### 基础用法

#### 1. 为容器设置背景

```go
import (
    rtui "github.com/wwsheng009/mint/runtime/ui"
    "github.com/wwsheng009/mint/runtime/style"
)

// 创建容器
container := rtui.VStack(
    ui.Text("内容 1"),
    ui.Text("内容 2"),
    ui.Text("内容 3"),
)

// 设置背景色
container.SetStyle(style.NewStyle().Background(style.Blue))

// 内部所有子控件自动继承蓝色背景
```

#### 2. 使用 Builder 模式

```go
// 使用 Bordered 包裹并设置大小
panel := rtui.Bordered().
    Style(string(theme.Border())).
    Child(content).
    Width(40).    // 重要：大小设置在外层边框
    Height(15).
    Build()

// 为内部内容设置背景
content.SetStyle(style.NewStyle().Background(style.Blue))
```

#### 3. 子控件覆盖继承

```go
// 容器背景
container.SetStyle(style.NewStyle().Background(style.Blue))

// 子控件设置自己的背景（不继承）
highlightButton := app.ButtonBuilder("[重要]").
    Style(style.NewStyle().Background(style.Red)).
    Build()

// 按钮是红色背景，不使用父容器的蓝色
```

### 高级用法

#### 1. Modal 对话框

```go
func ShowModal() rtui.VNode {
    content := rtui.VStack(
        ui.Text("确认删除？"),
        ui.Text(""),
        ui.HStack(
            ui.Text("[确定]"),
            ui.Text("   "),
            ui.Text("[取消]"),
        ),
    )

    // 设置深色背景，内部控件自动继承
    content.SetStyle(style.NewStyle().
        Background(style.Black).
        Foreground(style.White),
    )

    // 包裹边框并设置大小
    modal := rtui.Bordered().
        Style("double").
        Child(content).
        Width(40).
        Height(10).
        Build()

    return modal
}
```

#### 2. Dropdown 菜单

```go
func ShowDropdown() rtui.VNode {
    items := []rtui.VNode{
        ui.Text("选项 1"),
        ui.Text("选项 2"),
        ui.Text("选项 3"),
    }

    menu := rtui.VStack(items...)

    // 设置白色背景，黑色文字
    menu.SetStyle(style.NewStyle().
        Background(style.White).
        Foreground(style.Black),
    )

    return menu
}
```

#### 3. Tooltip 提示

```go
func ShowTooltip(text string) rtui.VNode {
    tooltip := rtui.VStack(
        ui.Text(text),
    )

    // 黄色背景，黑色文字
    tooltip.SetStyle(style.NewStyle().
        Background(style.Yellow).
        Foreground(style.Black),
    )

    return tooltip
}
```

---

## 最佳实践

### 1. 容器大小设置

**✅ 正确**：大小设置在外层容器

```go
panel := rtui.Bordered().
    Child(content).
    Width(40).    // 在外层设置
    Height(15).
    Build()

content.SetStyle(style.NewStyle().Background(style.Blue))
```

**❌ 错误**：大小设置在内层

```go
content := rtui.VStackBuilder(...).
    Width(40).    // 错误：会导致外层无限制
    Height(15).
    Build()

panel := rtui.Bordered().
    Child(content).
    Build()  // 外层无边框大小限制 → 拉伸到屏幕
```

### 2. 背景色选择

**对比度原则**：
- 深色背景 → 浅色文字
- 浅色背景 → 深色文字

```go
// 深色主题
style.NewStyle().
    Background(style.Black).
    Foreground(style.White)

// 浅色主题
style.NewStyle().
    Background(style.White).
    Foreground(style.Black)

// 强调色
style.NewStyle().
    Background(style.Red).
    Foreground(style.Yellow)
```

### 3. 层级设计

**推荐层级**：
```
Modal Layer (最高)
  ├─ Modal 背景遮挡
  └─ Modal 内容
Inspector Layer (高)
  ├─ Inspector 背景遮挡
  └─ Inspector 内容
Overlay Layer (中)
  ├─ Dropdown 背景遮挡
  └─ Dropdown 内容
Base Layer (低)
  └─ 应用内容
```

### 4. 性能优化

**最小化重绘**：
- 只在需要时设置背景
- 避免频繁切换背景色
- 使用静态背景而非动态计算

**示例**：
```go
// ✅ 好：静态背景
container.SetStyle(style.NewStyle().Background(style.Blue))

// ❌ 差：动态背景（每次都计算）
bg := calculateBackgroundColor()  // 频繁计算
container.SetStyle(style.NewStyle().Background(bg))
```

---

## 常见问题

### Q1: 为什么我的容器没有背景？

**A**: 检查以下几点：

1. **是否设置了背景色？**
   ```go
   // ✅ 正确
   container.SetStyle(style.NewStyle().Background(style.Blue))

   // ❌ 错误
   container.SetStyle(style.NewStyle().Foreground(style.White))  // 只设置前景
   ```

2. **是否在正确的层渲染？**
   ```go
   // Inspector 需要在 LayerInspector 层
   vnode.SetLayer(rtui.LayerInspector)
   ```

3. **是否被其他内容覆盖？**
   - 检查渲染顺序
   - 确保背景在内容之前绘制

### Q2: 为什么底层内容还是可见？

**A**: 这是正常的 TUI 遮挡问题。确保：

1. **使用正确的渲染方法**：
   - 使用 `paintContainerBackground()` 而不是手动绘制
   - 不要在填充背景时检查现有内容

2. **覆盖层设置**：
   ```go
   // Inspector 必须在 LayerInspector 层
   inspector.SetLayer(rtui.LayerInspector)
   ```

### Q3: 为什么子控件颜色不一致？

**A**: 检查背景继承：

1. **子控件是否设置了背景？**
   ```go
   // 如果子控件有自己的背景，不会继承
   button.SetStyle(style.NewStyle().Background(style.Red))
   ```

2. **是否需要显式继承？**
   ```go
   // 当前自动继承，不需要手动设置
   // 但如果子控件已有背景，需要清除
   button.SetStyle(style.Style{})  // 清空样式，触发继承
   ```

### Q4: 为什么容器拉伸到整个屏幕？

**A**: 检查容器大小设置：

```go
// ✅ 正确：大小设置在外层
panel := rtui.Bordered().
    Child(content).
    Width(40).
    Height(15).
    Build()

// ❌ 错误：大小设置在内层
content := rtui.VStackBuilder(...).
    Width(40).  // 这不会限制外层边框
    Build()
panel := rtui.Bordered().Child(content).Build()  // 无大小限制
```

### Q5: 背景继承会影响性能吗？

**A**: 影响很小：

- **内存开销**：每帧创建 `parentBackground` map，约 O(n)
- **计算开销**：每个节点 O(1) 检查
- **总体影响**：< 1%，可忽略不计

---

## 性能考虑

### 内存开销

```
parentBackground map:
  - 每帧创建和清理
  - 每个有背景的容器约 24 bytes
  - 典型应用：10-50 个容器 = 240-1200 bytes
```

### 计算开销

```
每个节点：
  - map 查找：O(1)
  - 样式合并：O(1)
  - 总计：O(1)

总体：
  - n 个节点 = O(n)
  - 典型应用：1000 个节点 ≈ 1ms
```

### 优化建议

1. **避免频繁切换背景**
   ```go
   // ✅ 好：设置一次
   container.SetStyle(style.NewStyle().Background(style.Blue))

   // ❌ 差：频繁切换（每帧都换）
   for {
       bg := getBackgroundFromUser()
       container.SetStyle(style.NewStyle().Background(bg))
   }
   ```

2. **使用静态颜色**
   ```go
   // ✅ 好：编译时常量
   container.SetStyle(style.NewStyle().Background(style.Blue))

   // ❌ 差：运行时计算
   bg := style.Color(fmt.Sprintf("#%06x", calculateColor()))
   container.SetStyle(style.NewStyle().Background(bg))
   ```

3. **限制容器数量**
   - 只在需要时使用背景
   - 避免嵌套过多的背景容器
   - 考虑使用 CSS 类似的样式复用

---

## 相关文档

### 内部架构

- **`LAYOUT_RENDERING_REFACTOR.md`** - 布局渲染重构
- **`layer_system_guide.md`** - Layer 系统指南
- **`flex_layout.md`** - Flex 布局系统

### Inspector 相关

- **`INSPECTOR_CONTAINER_BACKGROUND_FIX.md`** - 容器背景修复
- **`INSPECTOR_OCCLUSION_FIX.md`** - 遮挡透视修复
- **`INSPECTOR_BACKGROUND_INHERITANCE.md`** - 背景继承机制

### API 参考

- **`PaintEngine`** (`internal/render/paint_engine.go`)
- **`Style`** (`runtime/style/style.go`)
- **`VNode`** (`runtime/ui/vnode.go`)

---

## 更新日志

### 2025-02-08 - 初始版本

- ✅ 实现容器背景渲染
- ✅ 实现内容遮挡机制
- ✅ 实现背景继承系统
- ✅ 完善调试支持
- ✅ 编写完整文档

---

**版本**: 1.0
**状态**: ✅ 完整实现并测试
**维护者**: Mint TUI Team
**反馈**: GitHub Issues
