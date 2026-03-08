# Cursor 组件

`cursor` 提供统一的光标模型，支持：

- 独立光标组件（单独渲染一个光标）
- `input` / `textarea` 内嵌光标（推荐）
- 可配置形状（块状/下划线/细竖线）、闪烁频率、主题色

包路径：

```go
import "github.com/wwsheng009/mint/ui/components/cursor"
```

## 快速使用

### 1) 在 `Input`/`Textarea` 中使用细插入光标（推荐）

```go
ui.NewInputBuilder().
    Placeholder("请输入用户名").
    InsertCursor(). // 细竖线插入态光标
    Build()
```

```go
ui.NewTextareaBuilder().
    Value("hello").
    CursorShape(cursor.ShapeBar).
    CursorBlinkInterval(350 * time.Millisecond).
    Build()
```

### 2) 使用独立 Cursor 组件

```go
cursor.NewBuilder().
    Key("caret").
    Bar().
    FastBlink().
    Theme(cursor.ThemeCaret).
    Build()
```

## Config

`cursor.Config` 字段：

- `Blink bool`：是否闪烁
- `BlinkInterval time.Duration`：闪烁间隔
- `Shape cursor.Shape`：光标形状
- `Theme cursor.ThemeRole`：主题色角色
- `Style style.Style`：额外样式覆盖
- `Glyph string`：自定义字形（主要用于 `ShapeBar`）

默认配置（`DefaultConfig()`）：

- `Blink=true`
- `BlinkInterval=500ms`
- `Shape=ShapeBlock`
- `Theme=ThemeCaret`

## 常量

### 形状（`Shape`）

- `ShapeBlock`：块状
- `ShapeUnderline`：下划线
- `ShapeBar`：细竖线（插入态）

### 主题角色（`ThemeRole`）

- `ThemeCaret`
- `ThemeFocus`
- `ThemeText`
- `ThemeMuted`
- `ThemeAccent`

### 闪烁预设

- `FastBlinkInterval = 250ms`
- `NormalBlinkInterval = 500ms`
- `SlowBlinkInterval = 800ms`

## Builder API（独立 Cursor）

- `Key(string)`
- `Style(style.Style)`
- `Config(cursor.Config)`
- `Shape(cursor.Shape)` / `Block()` / `Underline()` / `Bar()`
- `Theme(cursor.ThemeRole)`
- `Glyph(string)`
- `Visible(bool)`
- `Blink(bool)` / `Steady()`
- `BlinkInterval(time.Duration)` / `FastBlink()` / `NormalBlink()` / `SlowBlink()`
- `Build()` / `BuildTyped()`

## Input/Textarea 的光标 API

`input.Builder` / `textarea.Builder`：

- `CursorConfig(cursor.Config)`
- `CursorShape(cursor.Shape)`
- `InsertCursor()` / `BlockCursor()` / `UnderlineCursor()`
- `CursorBlink(bool)`
- `CursorBlinkInterval(time.Duration)`

对应 VNode 也有同名 `Set...` 方法。

## 行为说明

- 光标只在“已聚焦且可编辑”时显示（禁用/只读状态不会显示）。
- 光标闪烁由框架 tick 驱动（默认主循环 tick 间隔约 16ms）。
- 仅设置 `Shape/Theme` 这类部分配置时，仍会保留默认闪烁行为。

## 从块状迁移到插入态（推荐）

原来（块状）：

```go
ui.NewInputBuilder().
    Placeholder("Username").
    Build()
```

迁移后（细插入光标）：

```go
ui.NewInputBuilder().
    Placeholder("Username").
    InsertCursor().
    Build()
```

进一步自定义：

```go
ui.NewInputBuilder().
    Placeholder("Username").
    InsertCursor().
    CursorBlinkInterval(350 * time.Millisecond).
    Build()
```
