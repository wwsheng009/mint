# Style API 重构提案

## 当前问题

### 1. 冗长的空结构体初始化

**现状**：
```go
style.Style{}.Foreground(color).Background(bg).Bold(true)
style.Style{}.Foreground(color)
style.Style{}.Reverse(true)
```

**问题**：
- 每次都要写 `style.Style{}{}` 前缀
- 代码重复，可读性差
- IDE 自动补全效率低

### 2. 缺少常用样式预设

**现状**：
```go
// 每次都要重复组合
style.Style{}.Foreground(color).Bold(true)
style.Style{}.Reverse(true)
style.Style{}.Foreground(color).Background(bg).Bold(true).Underline(true)
```

**问题**：
- 常用样式组合没有预设
- 开发者需要重复构建相同的样式
- 容易出现不一致

### 3. 空样式不直观

**现状**：
```go
if currentStyle != (style.Style{}) {
    currentStyle = style.Style{}
}
```

**问题**：
- `style.Style{}{}` 不直观
- 没有明确的"空样式"语义
- 零值比较不够优雅

## 重构方案

### 方案 1：添加构造函数（推荐）

#### 1.1 基础构造函数

```go
// 在 runtime/style/style.go 中添加

// New 创建新样式（别名，更短）
func New() Style {
    return Style{}
}

// Foreground 创建只有前景色的样式
func Foreground(c Color) Style {
    return Style{FG: c}
}

// Background 创建只有背景色的样式
func Background(c Color) Style {
    return Style{BG: c}
}

// Bold 创建粗体样式
func Bold() Style {
    return Style{isBold: true}
}

// Italic 创建斜体样式
func Italic() Style {
    return Style{isItalic: true}
}

// Underline 创建下划线样式
func Underline() Style {
    return Style{isUnderline: true}
}

// Reverse 创建反白样式
func Reverse() Style {
    return Style{isReverse: true}
}
```

**使用示例**：
```go
// 之前
style.Style{}.Foreground(color)
style.Style{}.Bold(true)
style.Style{}.Reverse(true)

// 之后
style.Foreground(color)
style.Bold()
style.Reverse()
```

#### 1.2 组合构造函数

```go
// FgBold 创建前景色+粗体样式
func FgBold(c Color) Style {
    return Style{FG: c, isBold: true}
}

// FgBg 创建前景色+背景色样式
func FgBg(fg, bg Color) Style {
    return Style{FG: fg, BG: bg}
}

// FgBgBold 创建前景色+背景色+粗体样式
func FgBgBold(fg, bg Color) Style {
    return Style{FG: fg, BG: bg, isBold: true}
}

// FgUnderline 创建前景色+下划线样式
func FgUnderline(c Color) Style {
    return Style{FG: c, isUnderline: true}
}
```

**使用示例**：
```go
// 之前
style.Style{}.Foreground(theme.Text()).Bold(true)
style.Style{}.Foreground(theme.Text()).Background(theme.Primary()).Bold(true)

// 之后
style.FgBold(theme.Text())
style.FgBgBold(theme.Text(), theme.Primary())
```

### 方案 2：添加常量预设

```go
// 在 runtime/style/presets.go 中添加（新文件）

package style

// 常用样式预设
var (
    // None 空样式（语义化的零值）
    None = Style{}

    // Reverse 反白样式（常用于高亮）
    ReverseStyle = Style{isReverse: true}

    // Bold 粗体样式
    BoldStyle = Style{isBold: true}

    // Underline 下划线样式
    UnderlineStyle = Style{isUnderline: true}
)

// WithForeground 返回带指定前景色的样式
func WithForeground(c Color) Style {
    return Style{FG: c}
}

// WithBackground 返回带指定背景色的样式
func WithBackground(c Color) Style {
    return Style{BG: c}
}

// WithFgBg 返回带前景色和背景色的样式
func WithFgBg(fg, bg Color) Style {
    return Style{FG: fg, BG: bg}
}
```

**使用示例**：
```go
// 之前
if currentStyle != (style.Style{}) {
    currentStyle = style.Style{}
}

// 之后
if currentStyle != style.None {
    currentStyle = style.None
}

// 之前
cursorStyle := style.Style{}.Reverse(true)

// 之后
cursorStyle := style.ReverseStyle
// 或者保持链式调用的灵活性
cursorStyle := style.Reverse()
```

### 方案 3：Fluent Builder 模式增强

保持现有链式调用，但添加 `Merge` 和 `Apply` 方法：

```go
// Merge 合并另一个样式到当前样式
func (s Style) Merge(other Style) Style {
    if other.FG != "" && other.FG != NoColor {
        s.FG = other.FG
    }
    if other.BG != "" && other.BG != NoColor {
        s.BG = other.BG
    }
    if other.isBold {
        s.isBold = true
    }
    if other.isItalic {
        s.isItalic = true
    }
    if other.isUnderline {
        s.isUnderline = true
    }
    if other.isReverse {
        s.isReverse = true
    }
    return s
}

// Clone 克隆当前样式
func (s Style) Clone() Style {
    return s
}

// IsEmpty 检查是否为空样式
func (s Style) IsEmpty() bool {
    return s.FG == "" && s.BG == "" &&
        !s.isBold && !s.isItalic && !s.isUnderline &&
        !s.isStrikethrough && !s.isReverse && !s.isBlink
}
```

**使用示例**：
```go
// 合并样式
baseStyle := style.Foreground(theme.Text())
customStyle := baseStyle.Merge(style.BoldStyle)

// 检查空样式
if !currentStyle.IsEmpty() {
    // ...
}
```

## 推荐实现策略

### 阶段 1：添加便捷构造函数（低风险）

**新增文件**：`runtime/style/constructors.go`

```go
package style

// =============================================================================
// 便捷构造函数
// =============================================================================

// New 创建新样式（更简短的别名）
func New() Style {
    return Style{}
}

// Foreground 创建只有前景色的样式
func Foreground(c Color) Style {
    return Style{FG: c}
}

// Background 创建只有背景色的样式
func Background(c Color) Style {
    return Style{BG: c}
}

// Bold 创建粗体样式
func Bold() Style {
    return Style{isBold: true}
}

// Italic 创建斜体样式
func Italic() Style {
    return Style{isItalic: true}
}

// Underline 创建下划线样式
func Underline() Style {
    return Style{isUnderline: true}
}

// Reverse 创建反白样式
func Reverse() Style {
    return Style{isReverse: true}
}

// =============================================================================
// 组合构造函数
// =============================================================================

// FgBold 创建前景色+粗体样式
func FgBold(c Color) Style {
    return Style{FG: c, isBold: true}
}

// FgBg 创建前景色+背景色样式
func FgBg(fg, bg Color) Style {
    return Style{FG: fg, BG: bg}
}

// FgBgBold 创建前景色+背景色+粗体样式
func FgBgBold(fg, bg Color) Style {
    return Style{FG: fg, BG: bg, isBold: true}
}

// FgUnderline 创建前景色+下划线样式
func FgUnderline(c Color) Style {
    return Style{FG: c, isUnderline: true}
}

// FgBgUnderline 创建前景色+背景色+下划线样式
func FgBgUnderline(fg, bg Color) Style {
    return Style{FG: fg, BG: bg, isUnderline: true}
}

// =============================================================================
// 常量预设
// =============================================================================

// None 空样式（语义化的零值）
const None = Style{}

// ReverseStyle 反白样式预设
var ReverseStyle = Reverse()

// BoldStyle 粗体样式预设
var BoldStyle = Bold()

// UnderlineStyle 下划线样式预设
var UnderlineStyle = Underline()
```

### 阶段 2：添加辅助方法（低风险）

**在 `runtime/style/style.go` 中添加**：

```go
// IsEmpty 检查是否为空样式
func (s Style) IsEmpty() bool {
    return s.FG == "" && s.BG == "" &&
        !s.isBold && !s.isItalic && !s.isUnderline &&
        !s.isStrikethrough && !s.isReverse && !s.isBlink
}

// Merge 合并另一个样式到当前样式
func (s Style) Merge(other Style) Style {
    result := s
    if other.FG != "" && other.FG != NoColor {
        result.FG = other.FG
    }
    if other.BG != "" && other.BG != NoColor {
        result.BG = other.BG
    }
    if other.isBold {
        result.isBold = true
    }
    if other.isItalic {
        result.isItalic = true
    }
    if other.isUnderline {
        result.isUnderline = true
    }
    if other.isStrikethrough {
        result.isStrikethrough = true
    }
    if other.isReverse {
        result.isReverse = true
    }
    if other.isBlink {
        result.isBlink = true
    }
    return result
}
```

### 阶段 3：迁移现有代码（可选）

**不需要立即迁移**，新旧 API 可以共存：
- 旧代码继续使用 `style.Style{}{}`
- 新代码使用便捷构造函数
- 逐步迁移，降低风险

## 迁移示例

### 示例 1：Text 组件

**之前**：
```go
app.NewTextBuilder("Title").
    Style(style.Style{}.Foreground(theme.Text()).Bold(true)).
    Build()
```

**之后**：
```go
app.NewTextBuilder("Title").
    Style(style.FgBold(theme.Text())).
    Build()
```

### 示例 2：Cursor 样式

**之前**：
```go
cursorStyle := style.Style{}.Reverse(true)
```

**之后（方案 A - 使用预设）**：
```go
cursorStyle := style.ReverseStyle
```

**之后（方案 B - 使用构造函数）**：
```go
cursorStyle := style.Reverse()
```

### 示例 3：空样式检查

**之前**：
```go
if currentStyle != (style.Style{}) {
    // ...
}
```

**之后**：
```go
if !currentStyle.IsEmpty() {
    // ...
}
```

## API 对比表

| 需求 | 旧 API | 新 API | 改进 |
|------|--------|--------|------|
| 空样式 | `style.Style{}` | `style.None` 或 `style.New()` | 更语义化 |
| 前景色 | `style.Style{}.Foreground(c)` | `style.Foreground(c)` | 少 8 字符 |
| 粗体 | `style.Style{}.Bold(true)` | `style.Bold()` | 少 10 字符 |
| 反白 | `style.Style{}.Reverse(true)` | `style.Reverse()` | 少 10 字符 |
| 前景+粗体 | `style.Style{}.Foreground(c).Bold(true)` | `style.FgBold(c)` | 少 18 字符 |
| 前景+背景+粗体 | `style.Style{}.Foreground(fg).Background(bg).Bold(true)` | `style.FgBgBold(fg, bg)` | 少 30 字符 |

## 向后兼容性

✅ **完全向后兼容**：
- 所有现有 `style.Style{}{}` 代码继续工作
- 新 API 是纯添加，不修改现有方法
- 可以逐步迁移，无需一次性改动

## 实现优先级

### P0（高优先级）
- ✅ 基础构造函数：`Foreground()`, `Background()`, `Bold()`, `Reverse()`
- ✅ 组合构造函数：`FgBold()`, `FgBg()`, `FgBgBold()`
- ✅ 常量预设：`None`, `ReverseStyle`, `BoldStyle`

### P1（中优先级）
- ⏳ 辅助方法：`IsEmpty()`, `Merge()`
- ⏳ 更多组合：`FgUnderline()`, `FgBgUnderline()`

### P2（低优先级）
- ⏸️ 文档更新
- ⏸️ 示例代码迁移
- ⏸️ 性能优化（如需要）

## 代码示例：完整的重构前后对比

### 场景：Ant Design Demo 中的列表项样式

**重构前**（`examples/ant_design_demo/main.go:99-109`）：
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

**重构后**：
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
- 代码更简洁（从 9 行减少到 7 行）
- 更易读（减少了重复的 `style.Style{}.`）
- 更少的函数调用（减少了 4 次方法调用）

## 总结

### 收益
1. **可读性提升**：代码更简洁，意图更清晰
2. **开发效率提升**：减少重复代码，IDE 补全更友好
3. **维护性提升**：常用样式集中管理，易于修改
4. **向后兼容**：不破坏现有代码

### 风险
- ⚠️ 新增 API 可能增加学习曲线（但旧 API 仍可用）
- ⚠️ 需要维护两套 API（但新 API 可以逐步替代旧 API）

### 建议
✅ **推荐实施**：
- 收益远大于风险
- 向后兼容，可以渐进式迁移
- 实施成本低，主要是添加新函数
