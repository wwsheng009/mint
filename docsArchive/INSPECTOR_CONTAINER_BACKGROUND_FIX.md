# Inspector 容器背景色增强

## 问题描述

用户要求给 **整个 Inspector 弹出框容器**增加背景色，而不是给单个控件增加背景。

之前的实现尝试给单独的 Text 控件设置背景，但这不是正确的解决方案。

## 根本原因

VNode 渲染系统（`PaintEngine.paintElement()`）之前**没有支持容器级别的背景渲染**。它只处理文本内容，而忽略了 VNode 的 `Style.Background` 属性。

## 解决方案

### 1. 增强 PaintEngine 支持容器背景渲染

**文件**: `internal/render/paint_engine.go`

**修改**: 在 `paintElement()` 方法中添加背景渲染逻辑：

```go
// paintElement paints an element node
func (e *PaintEngine) paintElement(box *compute.ComputedBox, buffer *paint.Buffer) {
    // ... 现有的文本处理逻辑 ...

    // ENHANCEMENT: Paint container background if set
    // This allows elements like Inspector panels to have solid backgrounds
    nodeStyle := box.VNode.Style()
    if nodeStyle.BG != "" && nodeStyle.BG != style.NoColor {
        e.paintContainerBackground(box, buffer, nodeStyle)
    }

    // For non-text elements, children will be painted after the switch
}
```

**新增方法**: `paintContainerBackground()`

```go
// paintContainerBackground fills the entire container area with background color
// This is used to create solid backgrounds for panels like Inspector
func (e *PaintEngine) paintContainerBackground(box *compute.ComputedBox, buffer *paint.Buffer, bgStyle style.Style) {
    // Create background style (only BG, no foreground)
    backgroundStyle := style.Style{}.Background(bgStyle.BG)

    // Fill entire container area with background color
    for y := 0; y < box.Box.Height; y++ {
        for x := 0; x < box.Box.Width; x++ {
            // Get current cell
            cell := buffer.GetContent(box.Box.X+x, box.Box.Y+y)

            // If cell is empty or has no background, set background
            if cell.Cluster == "" || cell.Cluster == " " {
                buffer.SetCell(box.Box.X+x, box.Box.Y+y, ' ', backgroundStyle)
            } else {
                // Cell has content - preserve content but add background
                // Merge existing style with background
                mergedStyle := cell.Style
                mergedStyle.BG = bgStyle.BG
                buffer.SetCell(box.Box.X+x, box.Box.Y+y, []rune(cell.Cluster)[0], mergedStyle)
            }
        }
    }

    if e.debug || os.Getenv("TUI_PAINT_DEBUG") == "true" {
        fmt.Fprintf(os.Stderr, "[Paint.paintContainerBackground] Filled %dx%d area at (%d,%d) with BG=%s\n",
            box.Box.Width, box.Box.Height, box.Box.X, box.Box.Y, bgStyle.BG)
    }
}
```

### 2. 修改 Inspector 使用容器背景

**文件**: `internal/inspector/standalone_inspector.go`

**修改**: 在 `buildOverlayContent()` 方法中给整个容器设置背景色：

```go
// Content VStack with background color applied
content := rtui.VStackBuilder(
    header,
    ui.Text("─"), // Separator
    activeTabContent,
).
    Width(si.overlayWidth).
    Height(si.overlayHeight).
    Build()

// Apply background color to the entire content container
// This will be rendered by the enhanced PaintEngine.paintContainerBackground()
content.SetStyle(style.NewStyle().Background(style.Blue))

// Wrap in bordered box
panel := rtui.Bordered().
    Style(string(theme.Border())).
    Child(content).
    Build()
```

**移除的内容**:
- 移除了单独的 Text 背景行（`contentBackgroundLine1` 等）
- 移除了 `backgroundFiller` 占位符
- 简化了样式链式调用（改为单行调用以避免语法错误）

## 技术细节

### 容器背景渲染算法

1. **遍历整个容器区域**: 双重循环遍历所有 cell (x, y)
2. **检查当前 cell 内容**:
   - 如果是空 cell 或只有空格: 直接设置背景
   - 如果有内容: 保留内容，但合并样式添加背景
3. **使用 buffer.SetCell()**: 为每个位置设置带背景的字符

### 样式合并策略

```go
// Cell has content - preserve content but add background
mergedStyle := cell.Style  // 保留现有样式
mergedStyle.BG = bgStyle.BG // 添加背景色
buffer.SetCell(x, y, content, mergedStyle)
```

## 效果对比

### 之前（无容器背景）

```
┌──────────────────────────────────────┐
│ Runtime Scheduling Pipeline         │  ← 应用标题（默认样式）
├──────────────────────────────────────┤
│ [╔═ INSPECTOR ═╗]                   │  ← Inspector（无背景，难以区分）
│ [F12: close | 1-5: tabs]             │
│                                      │
└──────────────────────────────────────┘
```

### 现在（蓝色背景容器）

```
┌──────────────────────────────────────┐
│ Runtime Scheduling Pipeline         │  ← 应用标题（默认样式）
├──────────────────────────────────────┤
│        ╔═ INSPECTOR ═╗              │  ← Inspector 蓝色标题
│        F12:关闭 | 1-5:标签页         │
│        Alt+H/J/K/L:移动面板          │
│        ┌──────────────────┐          │  ← 整个容器有蓝色背景
│        │ Tree View        │          │
│        └──────────────────┘          │
└──────────────────────────────────────┘
```

## 调试支持

### 启用背景渲染调试

```bash
# 查看容器背景渲染过程
TUI_PAINT_DEBUG=true ./demo2_inspector.exe

# 输出示例：
# [Paint.paintContainerBackground] Filled 38x20 area at (80, 5) with BG=blue
```

## 验收标准

- [x] Inspector 容器有明显的背景色
- [x] 背景覆盖整个容器区域（不仅是文本行）
- [x] 背景不影响子元素的渲染
- [x] 所有代码编译通过
- [x] 向后兼容（不影响没有设置背景的组件）

## 相关文件

### 修改的文件

1. **`internal/render/paint_engine.go`**
   - 修改 `paintElement()` 添加背景渲染调用
   - 新增 `paintContainerBackground()` 方法

2. **`internal/inspector/standalone_inspector.go`**
   - 修改 `buildOverlayContent()` 使用容器背景
   - 简化样式链式调用

### 新增的文件

3. **`INSPECTOR_CONTAINER_BACKGROUND_FIX.md`** - 本文档

## 未来扩展

这个增强不仅限于 Inspector，现在**任何 VNode 都可以通过设置 `Style.Background` 来获得容器背景色**：

```go
// 为任意容器设置背景
myContainer := rtui.VStackBuilder(
    ui.Text("Content"),
    ui.Text("More content"),
).
    Width(40).
    Height(10).
    Build()

myContainer.SetStyle(style.NewStyle().Background(style.Red))
```

## 限制和注意事项

1. **性能**: 背景渲染需要遍历整个容器区域，对于超大容器可能有性能影响
2. **Z-Index**: 背景在子元素**之前**渲染，子元素内容会覆盖背景
3. **透明度**: TUI 不支持真正的透明度，背景是实色的
4. **兼容性**: 需要使用增强后的 PaintEngine（已包含在 framework 中）

## 测试

### 单元测试

已创建测试文件验证功能：
- `internal/inspector/standalone_inspector_bg_test.go`

### 集成测试

运行 demo2 验证效果：
```bash
cd examples/ui_demos/demo2_runtime_internals/inspector_overlay
go build -o demo2_inspector.exe main.go
./demo2_inspector.exe
```

操作步骤：
1. 按 **F12** 打开 Inspector
2. 观察 Inspector 面板是否有蓝色背景
3. 使用 **Alt+H/J/K/L** 移动面板，背景应该跟随面板移动

---

**实施日期**: 2025-02-08
**状态**: ✅ 完成并测试
**版本**: 1.0
**兼容性**: 向后兼容，不影响现有组件
