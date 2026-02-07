# Style API 迁移示例

本文档展示如何在实际代码中使用新的 Style API。

## 基础示例

### 示例 1：文本样式

**旧代码** (`examples/ui_demos/demo1_full_featured/main.go:85`):
```go
app.NewTextBuilder("TUI Engine Demo").
    Style(style.Style{}.
        Foreground(theme.Text()).
        Background(theme.Primary()).
        Bold(true)).
    Build()
```

**新代码**:
```go
app.NewTextBuilder("TUI Engine Demo").
    Style(style.FgBgBold(theme.Text(), theme.Primary())).
    Build()
```

**改进**：
- 代码行数减少 1 行
- 字符数减少 24 个（从 75 减少到 51）
- 更易读：`FgBgBold` 清晰表达"前景色+背景色+粗体"

---

### 示例 2：光标样式

**旧代码** (`framework/cursor/cursor.go:82`):
```go
style:          style.Style{}.Reverse(true),
```

**新代码**（选项 A - 使用构造函数）:
```go
style:          style.Reverse(),
```

**新代码**（选项 B - 使用预设）:
```go
style:          style.ReverseStyle,
```

**推荐**：选项 B（预设）用于常用样式，选项 A（构造函数）用于自定义样式

---

### 示例 3：边框样式

**旧代码** (`examples/debug_test/main.go:72`):
```go
borderStyle := style.Style{}.Foreground(borderColor)
```

**新代码**:
```go
borderStyle := style.Foreground(borderColor)
```

**改进**：
- 字符数减少 11 个
- 更简洁，无需 `.Style{}{}` 前缀

---

### 示例 4：列表项样式（Ant Design）

**旧代码** (`examples/ant_design_demo/main.go:99-109`):
```go
if item.Hovered {
    itemStyle = style.Style{}.
        Foreground(theme.Primary()).
        Background(theme.Primary().WithOpacity("0.1"))
} else if item.Selected {
    itemStyle = style.Style{}.
        Foreground(theme.Primary()).
        Background(theme.Primary().WithOpacity("0.2"))
} else if item.Disabled {
    itemStyle = style.Style{}.
        Foreground(theme.Muted())
}
```

**新代码**:
```go
if item.Hovered {
    itemStyle = style.FgBg(
        theme.Primary(),
        theme.Primary().WithOpacity("0.1"),
    )
} else if item.Selected {
    itemStyle = style.FgBg(
        theme.Primary(),
        theme.Primary().WithOpacity("0.2"),
    )
} else if item.Disabled {
    itemStyle = style.Foreground(theme.Muted())
}
```

**改进**：
- 从 9 行减少到 7 行
- 减少了重复的 `style.Style{}.` 前缀
- 代码更清晰，逻辑更突出

---

### 示例 5：检查空样式

**旧代码** (`framework/app.go:856`):
```go
if currentStyle != (style.Style{}) {
    if currentStyle != (style.Style{}) {
        output.WriteString("\x1b[0m")
    }
    if cell.Style != (style.Style{}) {
        output.WriteString(cell.Style.ToANSI())
    }
    currentStyle = cell.Style
}
```

**新代码**:
```go
if !currentStyle.IsEmpty() {
    if !currentStyle.IsEmpty() {
        output.WriteString("\x1b[0m")
    }
    if !cell.Style.IsEmpty() {
        output.WriteString(cell.Style.ToANSI())
    }
    currentStyle = cell.Style
}
```

**改进**：
- 更语义化：`IsEmpty()` 比 `!= (style.Style{})` 更清晰
- 更易读：意图明确，无需理解零值比较
- 减少括号，代码更简洁

---

### 示例 6：合并样式

**旧代码**（使用多个链式调用）:
```go
baseStyle := style.Style{}.Foreground(theme.Text())
modifiedStyle := baseStyle.Bold(true).Background(theme.Primary())
```

**新代码**（选项 A - 使用 Merge）:
```go
baseStyle := style.Foreground(theme.Text())
modifiedStyle := baseStyle.Merge(style.FgBgBold(theme.Primary(), theme.Primary()))
```

**新代码**（选项 B - 直接构造）:
```go
modifiedStyle := style.FgBgBold(theme.Text(), theme.Primary())
```

**推荐**：选项 B（直接构造）更简单，除非需要动态合并多个样式

---

### 示例 7：表格标题样式

**旧代码** (`components/data/table.go:33`):
```go
headerStyle:  style.Style{}.Bold(true),
```

**新代码**（选项 A - 使用构造函数）:
```go
headerStyle:  style.Bold(),
```

**新代码**（选项 B - 使用预设）:
```go
headerStyle:  style.BoldStyle,
```

**推荐**：选项 B（预设）用于常用样式

---

### 示例 8：调试面板样式

**旧代码** (`examples/devtools_demo/interactive_panel/main.go:194-229`):
```go
helpStyle := style.Style{}.Foreground(style.Yellow)
eventStyle := style.Style{}.Foreground(style.Magenta)
titleStyle := style.Style{}.Foreground(style.Cyan).Bold(true)
labelStyle := style.Style{}.Foreground(style.Green)
valueStyle := style.Style{}.Foreground(style.White)
cellStyle := style.Style{}.Foreground(style.Blue)
// ...
cellStyle = style.Style{}.Foreground(style.BrightBlack)
```

**新代码**:
```go
helpStyle := style.Foreground(style.Yellow)
eventStyle := style.Foreground(style.Magenta)
titleStyle := style.FgBold(style.Cyan)
labelStyle := style.Foreground(style.Green)
valueStyle := style.Foreground(style.White)
cellStyle := style.Foreground(style.Blue)
// ...
cellStyle = style.Foreground(style.BrightBlack)
```

**改进**：
- 字符数减少约 80 个
- 代码更一致，减少了重复的 `.Style{}{}` 前缀
- `titleStyle` 使用 `FgBold` 更简洁

---

## 常见模式

### 模式 1：前景色 + 粗体（最常用）

**旧代码**:
```go
style.Style{}.Foreground(color).Bold(true)
```

**新代码**:
```go
style.FgBold(color)
```

**减少字符数**：18 个

---

### 模式 2：前景色 + 背景色

**旧代码**:
```go
style.Style{}.Foreground(fg).Background(bg)
```

**新代码**:
```go
style.FgBg(fg, bg)
```

**减少字符数**：19 个

---

### 模式 3：反白样式

**旧代码**:
```go
style.Style{}.Reverse(true)
```

**新代码**:
```go
style.Reverse()
// 或
style.ReverseStyle
```

**减少字符数**：10 个

---

### 模式 4：单个属性

**旧代码**:
```go
style.Style{}.Foreground(c)
style.Style{}.Background(c)
style.Style{}.Bold(true)
style.Style{}.Italic(true)
style.Style{}.Underline(true)
```

**新代码**:
```go
style.Foreground(c)
style.Background(c)
style.Bold()
style.Italic()
style.Underline()
```

**每个减少字符数**：8-11 个

---

## 完整迁移案例

### 案例：Debug Test 示例

**旧代码** (`examples/debug_test/main.go:72-73, 220-229`):
```go
borderStyle := style.Style{}.Foreground(borderColor)
textStyle := style.Style{}.Foreground(style.Cyan)
// ...
emptyStyle := style.Style{}
titleStyle := style.Style{}.Foreground(style.Cyan).Bold(true)
infoStyle := style.Style{}.Foreground(style.White)
```

**新代码**:
```go
borderStyle := style.Foreground(borderColor)
textStyle := style.Foreground(style.Cyan)
// ...
emptyStyle := style.None
titleStyle := style.FgBold(style.Cyan)
infoStyle := style.Foreground(style.White)
```

**统计**：
- 总行数不变
- 总字符数减少约 60 个
- 可读性提升

---

## 性能对比

根据基准测试结果 (`runtime/style/constructors_test.go`):

| 场景 | 旧 API | 新 API | 提升 |
|------|--------|--------|------|
| FgBgBold(Red, Blue) | 15.35 ns/op | 0.88 ns/op | **17.5x** |
| Bold() | 0.91 ns/op | 1.29 ns/op | 0.7x (略慢，但可忽略) |

**结论**：新 API 在复杂样式构建上性能显著提升，在简单样式上性能相当。

---

## 迁移策略

### 策略 1：渐进式迁移（推荐）

1. **新代码**：直接使用新 API
   ```go
   // 新组件使用新 API
   titleStyle := style.FgBold(theme.Text())
   ```

2. **旧代码**：保持不变，除非需要修改
   ```go
   // 旧代码继续使用旧 API
   oldStyle := style.Style{}.Foreground(c).Bold(true)
   ```

3. **修改时迁移**：修改旧代码时顺便迁移
   ```go
   // 修改附近代码时，将旧 API 替换为新 API
   oldStyle := style.Style{}.Foreground(c).Bold(true) // 删除
   newStyle := style.FgBold(c)                        // 添加
   ```

### 策略 2：批量迁移（可选）

如果项目规模较小，可以使用文本替换批量迁移：

| 查找 | 替换为 |
|------|--------|
| `style.Style{}.Reverse(true)` | `style.Reverse()` |
| `style.Style{}.Bold(true)` | `style.Bold()` |
| `style.Style{}.Italic(true)` | `style.Italic()` |
| `style.Style{}.Underline(true)` | `style.Underline()` |
| `style.Style{}.Foreground(\([^)]+\)).Bold(true)` | `style.FgBold($1)` |
| `style.Style{}.Foreground(\([^)]+\)).Background(\([^)]+\))` | `style.FgBg($1, $2)` |

**警告**：批量替换前务必备份，并仔细检查结果！

---

## 总结

### 新 API 的优势

1. **更简洁**：减少 20-40% 的代码字符数
2. **更易读**：意图更清晰，减少了重复的 `.Style{}{}` 前缀
3. **更快速**：复杂样式构建性能提升 17.5x
4. **向后兼容**：旧 API 继续工作，可渐进式迁移

### 迁移建议

- ✅ 新代码使用新 API
- ✅ 常用模式使用预设（`ReverseStyle`, `BoldStyle`）
- ✅ 修改旧代码时顺便迁移
- ⚠️ 不强制迁移旧代码（除非需要修改）
- ⚠️ 复杂样式组合考虑使用 `Merge()`

### 下一步

1. 更新组件代码（`components/`）使用新 API
2. 更新示例代码（`examples/`）使用新 API
3. 更新文档使用新 API
4. 考虑添加代码分析规则，推荐使用新 API
