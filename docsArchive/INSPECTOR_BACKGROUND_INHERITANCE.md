# Inspector 背景继承解决方案

## 问题

当 Inspector 容器设置了蓝色背景后，内部子控件（Button、Text 等）仍然使用默认的黑色背景，导致**视觉不一致**：

```
┌──────────────────────────────────────┐
│     ╔═ INSPECTOR ═╗                 │  ← 蓝色容器背景
│     ┌──────────────┐                │
│     │ [Button]     │ ← 黑色背景      │  ← 不一致！
│     └──────────────┘                │
└──────────────────────────────────────┘
```

## 解决方案

实现**子控件自动继承父容器背景色**机制。当子控件没有明确设置背景色时，自动使用父容器的背景色。

### 实现原理

#### 1. 记录父容器背景

当容器节点有背景色时，在 `paintElement()` 中记录父容器背景：

```go
// internal/render/paint_engine.go:paintElement()

if nodeStyle.BG != "" && nodeStyle.BG != style.NoColor {
    e.paintContainerBackground(box, buffer, nodeStyle)

    // IMPORTANT: Store parent background for child inheritance
    if e.parentBackground == nil {
        e.parentBackground = make(map[*compute.ComputedBox]style.Color)
    }
    for _, childBox := range box.Children {
        e.parentBackground[childBox] = nodeStyle.BG
    }
}
```

#### 2. 应用背景继承

在 `paintNode()` 中检查并应用继承：

```go
// internal/render/paint_engine.go:paintNode()

// Check if we have a parent background to inherit
var parentBG style.Color
if e.parentBackground != nil {
    if inheritedBG, ok := e.parentBackground[box]; ok && inheritedBG != "" {
        parentBG = inheritedBG
    }
}

// For Paintable components (Button, etc.)
if paintable {
    commands := paintable.Paint(box.Box.X, box.Box.Y)
    for _, cmd := range commands {
        styleToApply := cmd.Style
        // If command has no background and parent has one, inherit it
        if parentBG != "" && (styleToApply.BG == "" || styleToApply.BG == style.NoColor) {
            styleToApply.BG = parentBG
        }
        buffer.SetString(cmd.X, cmd.Y, cmd.Text, styleToApply)
    }
    return nil
}

// For non-Paintable nodes (Text, VStack, etc.)
if parentBG != "" {
    nodeStyle := box.VNode.Style()
    if nodeStyle.BG == "" || nodeStyle.BG == style.NoColor {
        inheritedStyle := nodeStyle
        inheritedStyle.BG = parentBG
        box.VNode.SetStyle(inheritedStyle)
    }
}

// Clean up to avoid affecting siblings/cousins
delete(e.parentBackground, box)
```

#### 3. 每帧清理

在每帧渲染开始时清理继承映射：

```go
// internal/render/paint_engine.go:Paint()

func (e *PaintEngine) Paint(layout *compute.ComputedLayout, buffer *paint.Buffer) error {
    // Clear parent background map at the start of each frame
    e.parentBackground = make(map[*compute.ComputedBox]style.Color)

    // ... rest of rendering
}
```

## 技术细节

### 继承规则

1. **自动继承**：子控件如果没有设置 `Style.Background`，自动继承父容器背景
2. **显式优先**：如果子控件明确设置了背景色，使用子控件的背景（不继承）
3. **递归应用**：背景继承会递归应用到所有后代节点
4. **作用域限制**：背景继承只影响直接子节点和后代节点，不影响兄弟节点

### 适用组件

#### Paintable 组件（使用 DrawCmd）

- **Button** - 按钮组件
- **Input** - 输入框组件
- 任何实现 `Paint(int, int) []paint.DrawCmd` 的组件

继承方式：在应用每个 DrawCmd 时，如果 Style 没有背景，添加父容器背景。

#### 非 Paintable 组件（使用 VNode）

- **Text** - 文本组件
- **VStack/HStack** - 布局容器
- **Box** - 盒子容器

继承方式：直接修改 VNode 的 Style，添加父容器背景。

## 效果对比

### 修复前（背景不一致）

```
┌──────────────────────────────────────┐
│     ╔═ INSPECTOR ═╗                 │  ← 蓝色容器背景
│     F12:关闭 | 1-5:标签页            │
│     ┌──────────────────┐             │
│     │ [Button]         │ ← 黑色背景  │  ❌ 不一致
│     │ Some Text        │ ← 透明背景  │  ❌ 不一致
│     └──────────────────┘             │
└──────────────────────────────────────┘
```

### 修复后（视觉一致）

```
┌──────────────────────────────────────┐
│     ╔═ INSPECTOR ═╗                 │  ← 蓝色容器背景
│     F12:关闭 | 1-5:标签页            │  ← 继承蓝色背景
│     ┌──────────────────┐             │
│     │ [Button]         │ ← 蓝色背景  │  ✅ 一致
│     │ Some Text        │ ← 蓝色背景  │  ✅ 一致
│     └──────────────────┘             │
└──────────────────────────────────────┘
```

## 调试支持

### 启用继承调试

```bash
# 查看背景继承过程
TUI_PAINT_DEBUG=true ./demo2_inspector_inherit.exe

# 输出示例：
# [Paint.paintNode]   🎨 Inherited parent BG=blue
# [Paint.paintNode]   🎨 Paintable inherited parent BG=blue
```

### 调试标志

- `🎨 Inherited parent BG` - 非 Paintable 组件继承了背景
- `🎨 Paintable inherited parent BG` - Paintable 组件继承了背景

## 使用场景

### 1. Inspector 面板

```go
// Inspector 容器设置蓝色背景
content.SetStyle(style.NewStyle().Background(style.Blue))

// 内部的所有控件自动继承蓝色背景
// Button、Text 等都显示为蓝色背景
```

### 2. Modal 对话框

```go
modal := rtui.VStackBuilder(
    ui.Text("确认删除？"),
    ui.HStack(
        ui.Text("[确定]"),
        ui.Text("[取消]"),
    ),
).
    Width(40).
    Height(10).
    Build()

// 设置深色背景，内部控件自动继承
modal.SetStyle(style.NewStyle().Background(style.Black))
```

### 3. Dropdown 菜单

```go
dropdown := rtui.VStackBuilder(
    ui.Text("选项 1"),
    ui.Text("选项 2"),
    ui.Text("选项 3"),
).
    Width(20).
    Build()

// 设置白色背景，所有选项自动继承
dropdown.SetStyle(style.NewStyle().Background(style.White).Foreground(style.Black))
```

## 限制和注意事项

### 1. 显式设置优先

如果子控件明确设置了背景色，不会继承：

```go
// 子控件有自己的背景色
button := app.ButtonBuilder("[按钮]").
    Style(style.NewStyle().Background(style.Red)).  // 显式设置
    Build()

// 即使父容器是蓝色，按钮仍然是红色
```

### 2. 性能考虑

- **内存开销**：每帧创建新的 `parentBackground` map
- **计算开销**：每个节点需要检查是否需要继承
- **优化**：map 操作是 O(1)，影响很小

### 3. 继承深度

背景继承是**单层**的：
- 父容器 → 子节点
- 祖父容器 → 孙节点（通过中间节点传递）

不需要递归查找祖先节点，因为渲染时父节点已经处理了继承。

### 4. 样式合并

对于 Paintable 组件，背景继承通过**样式合并**实现：

```go
// 合并策略
styleToApply = cmd.Style           // 子控件原始样式
if parentBG != "" && cmd.Style.BG == "" {
    styleToApply.BG = parentBG     // 添加父容器背景
}
```

## 相关文件

### 修改的文件

1. **`internal/render/paint_engine.go`**
   - 添加 `parentBackground` 字段到 PaintEngine 结构
   - 修改 `paintElement()` 记录父容器背景
   - 修改 `paintNode()` 应用背景继承
   - 修改 `Paint()` 清理每帧状态

2. **`INSPECTOR_BACKGROUND_INHERITANCE.md`** - 本文档

## 未来扩展

### 1. 可配置继承

允许禁用背景继承：

```go
// 在容器上设置禁用继承标志
container.SetProp("disableBackgroundInherit", true)
```

### 2. 选择性继承

只继承特定属性的背景：

```go
// 只继承边框内的背景，不继承外边距区域
container.SetProp("inheritMode", "content-only")
```

### 3. 多级继承

支持从多层祖先继承背景（如果父容器也没有背景）：

```go
// 递归查找祖先节点的背景
func findAncestorBackground(box) style.Color {
    for ancestor := box.Parent; ancestor != nil; ancestor = ancestor.Parent {
        if bg := ancestor.VNode.Style().BG; bg != "" {
            return bg
        }
    }
    return ""
}
```

## 验收标准

- [x] Inspector 子控件自动继承蓝色背景
- [x] 视觉效果一致（没有黑色块）
- [x] 显式设置背景的子控件不受影响
- [x] 移动 Inspector 时，继承效果正常
- [x] 所有代码编译通过
- [x] 调试输出清晰

---

**实施日期**: 2025-02-08
**状态**: ✅ 完成并测试
**版本**: 1.0
**解决问题**: Inspector 子控件背景不一致
