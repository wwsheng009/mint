# Inspector 背景遮挡修复

## 问题

背景色生效了，但是**被 Inspector 遮挡的应用内容还是可见**（"透视"效果）。

### 根本原因

之前的 `paintContainerBackground()` 实现有问题：

```go
// 旧代码 - 有问题
cell := buffer.GetContent(box.Box.X+x, box.Box.Y+y)
if cell.Cluster == "" || cell.Cluster == " " {
    buffer.SetCell(box.Box.X+x, box.Box.Y+y, ' ', backgroundStyle)
} else {
    // 保留现有内容，只添加背景
    mergedStyle := cell.Style
    mergedStyle.BG = bgStyle.BG
    buffer.SetCell(box.Box.X+x, box.Box.Y+y, []rune(cell.Cluster)[0], mergedStyle)
}
```

**问题**：
1. Inspector 作为覆盖层（LayerInspector），在底层应用内容之后渲染
2. 当读取 `buffer.GetContent()` 时，读取的是**底层已渲染的应用内容**
3. 代码会"保留"这些内容并只添加背景，导致透视效果
4. 这不是真正的"遮挡"

## 解决方案

**无条件填充整个区域**，不管当前位置是否有内容：

```go
// 新代码 - 正确
func (e *PaintEngine) paintContainerBackground(box *compute.ComputedBox, buffer *paint.Buffer, bgStyle style.Style) {
    backgroundStyle := style.Style{}.Background(bgStyle.BG)

    // CRITICAL: Unconditionally fill entire container area
    // This occludes any content rendered underneath
    for y := 0; y < box.Box.Height; y++ {
        for x := 0; x < box.Box.Width; x++ {
            // Use space character ' ' to clear any existing content
            buffer.SetCell(box.Box.X+x, box.Box.Y+y, ' ', backgroundStyle)
        }
    }
}
```

### 关键改进

1. **无条件填充**：不再检查 cell 是否有内容，直接用空格和背景色覆盖
2. **清除内容**：使用空格字符 `' '` 替换任何现有内容
3. **遮挡底层**：确保底层的所有内容都被遮挡

### 渲染顺序

正确的渲染顺序：

```
1. paintContainerBackground()
   └─ 填充整个区域，遮挡底层内容

2. paintChildren()
   └─ 绘制子元素（边框、文字等）
   └─ 子元素绘制在背景之上，正常显示
```

## 效果对比

### 修复前（透视效果）

```
┌──────────────────────────────────────┐
│ Runtime Scheduling Pipeline         │  ← 应用内容可见
│ [Event] [setState] [Scheduler] ...  │  ← 透过 Inspector 可见
│                                      │
│     ╔═ INSPECTOR ═╗                 │  ← Inspector 半透明
│     F12:关闭 | 1-5:标签页            │     能看到背后的内容
└──────────────────────────────────────┘
```

### 修复后（完全遮挡）

```
┌──────────────────────────────────────┐
│ Runtime Scheduling Pipeline         │
│ [Event] [setState] [Scheduler] ...  │
│                                      │
│     ╔═ INSPECTOR ═╗                 │  ← Inspector 完全不透明
│     F12:关闭 | 1-5:标签页            │     蓝色背景遮挡背后的内容
│     ┌──────────────────┐             │
│     │ Tree View        │             │  ← 内容在背景之上正常显示
│     └──────────────────┘             │
└──────────────────────────────────────┘
```

## 技术细节

### TUI 渲染的"遮挡"原理

在终端 UI 中，没有真正的 alpha 通道或透明度。遮挡效果通过**绘制顺序**实现：

```
Layer Base (应用内容)
  ↓ 渲染到 buffer
Layer Inspector (Inspector 覆盖层)
  ↓ paintContainerBackground() - 用背景色覆盖 base 层的内容
  ↓ paintChildren() - 在背景之上绘制 Inspector 的内容
```

### 为什么之前不工作

```go
// 旧代码逻辑：
1. 读取 buffer 当前内容（可能是 base 层的应用内容）
2. 如果有内容，保留内容，只添加背景
3. 结果：base 层的内容 + Inspector 背景 = 透视效果

// 新代码逻辑：
1. 无条件用空格 + 背景色覆盖
2. 完全清除 base 层的内容
3. 结果：纯净的背景 = 完全遮挡
```

## 调试

### 启用调试输出

```bash
# 查看背景遮挡过程
TUI_PAINT_DEBUG=true ./demo2_inspector_fixed.exe

# 输出示例：
# [Paint.paintContainerBackground] Occluded 38x20 area at (80, 5) with BG=blue
```

注意日志消息从 "Filled" 改为 "**Occluded**"，强调遮挡功能。

## 验收标准

- [x] Inspector 背景完全遮挡底层应用内容
- [x] Inspector 的边框和文字正常显示
- [x] 移动 Inspector 时，遮挡区域跟随移动
- [x] 关闭 Inspector 时，底层内容正确恢复
- [x] 所有代码编译通过

## 相关文件

### 修改的文件

1. **`internal/render/paint_engine.go`**
   - 修改 `paintContainerBackground()` 方法
   - 改为无条件填充，实现真正的遮挡

2. **`INSPECTOR_OCCLUSION_FIX.md`** - 本文档

## 适用场景

这个修复适用于任何需要**背景遮挡**的场景：

### 1. Modal 对话框

```go
modal := rtui.VStackBuilder(
    ui.Text("重要消息"),
    ui.Text("确定要删除吗？"),
    ui.HStack(
        ui.Text("[确定]"),
        ui.Text("[取消]"),
    ),
).
    Width(40).
    Height(10).
    Build()

// 设置背景遮挡底层内容
modal.SetStyle(style.NewStyle().Background(style.Black))
```

### 2. Dropdown 菜单

```go
dropdown := rtui.VStackBuilder(
    ui.Text("选项 1"),
    ui.Text("选项 2"),
    ui.Text("选项 3"),
).
    Width(20).
    Build()

// 菜单背景遮挡下面的内容
dropdown.SetStyle(style.NewStyle().Background(style.White).Foreground(style.Black))
```

### 3. Tooltip 提示

```go
tooltip := rtui.VStackBuilder(
    ui.Text("这是一个提示"),
).
    Width(15).
    Build()

// 提示背景遮挡下面的内容
tooltip.SetStyle(style.NewStyle().Background(style.Yellow).Foreground(style.Black))
```

## 未来优化

### 1. 圆角裁剪

当前是矩形裁剪，未来可以实现圆角：

```go
// 伪代码
func paintRoundedBackground(box, buffer, bgStyle) {
    for y := 0; y < box.Box.Height; y++ {
        for x := 0; x < box.Box.Width; x++ {
            if isCorner(x, y) && !isInCircle(x, y) {
                continue // 跳过圆角外的区域
            }
            buffer.SetCell(box.Box.X+x, box.Box.Y+y, ' ', backgroundStyle)
        }
    }
}
```

### 2. 边框内边距处理

当前边框内的内容区域需要特殊处理：

```go
// 对于有边框的容器，只填充内容区域
if hasBorder {
    fillArea(box.Box.X+1, box.Box.Y+1, box.Box.Width-2, box.Box.Height-2)
}
```

### 3. 部分透明（模拟）

可以通过颜色选择模拟半透明：

```go
// 使用较暗的背景色模拟半透明
dimmedBg := blendColor(bgStyle.BG, underlyingContent.FG, 0.3)
```

---

**实施日期**: 2025-02-08
**状态**: ✅ 完成并测试
**版本**: 1.0
**修复问题**: Inspector 背景透视效果
