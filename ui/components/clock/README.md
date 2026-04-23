# Clock

圆盘时钟组件，适合显示当前时间、固定时间快照和简单的仪表式时钟界面。

## 已支持

- ASCII 圆盘表盘
- 时针 / 分针 / 秒针
- 实时时钟模式
- 固定时间模式
- 平滑秒针
- ASCII / Unicode 指针样式
- 圆形 / 椭圆形表盘
- `Shape(...) / Circle() / Ellipse()`
- `Radius(...) / RadiusX(...) / RadiusY(...) / Radii(...)`
- `CellAspectX(...)`，可调终端字符横向补偿
- 内置视觉预设：Default / Classic / Neon / Minimal / Alert
- 可导出的 `Theme` 对象，可从 preset 克隆后局部覆写
- `ThemePreset(...)` / `Theme.WithPreset(...)` 这种更顺手的 preset 入口
- `Theme.Merge(...)`，可组合基础主题层和强调主题层
- 外圈 / 刻度 / 中心点 / 数字独立样式
- 时针 / 分针 / 秒针独立样式
- `radius`
- `showDigital`
- 自定义 `style`

## 示例

```go
ui.NewClockBuilder().
    Circle().
    Radius(6).
    CellAspectX(2.0).
    Realtime().
    Preset(clock.PresetNeon).
    ShowSecondHand(true).
    SmoothSecond(true).
    UnicodeHands().
    DialStyle(style.Style{}.Foreground(style.BrightBlack)).
    TickStyle(style.Style{}.Foreground(style.BrightWhite).Bold(true)).
    CenterStyle(style.Style{}.Foreground(style.Yellow).Bold(true)).
    DigitalStyle(style.Style{}.Foreground(style.Cyan).Bold(true)).
    HourHandStyle(style.Style{}.Foreground(style.BrightYellow).Bold(true)).
    MinuteHandStyle(style.Style{}.Foreground(style.BrightCyan).Bold(true)).
    SecondHandStyle(style.Style{}.Foreground(style.BrightRed).Bold(true)).
    Build()
```

椭圆表盘：

```go
ui.NewClockBuilder().
    Ellipse().
    Radii(7, 4).
    CellAspectX(2.25).
    StaticTime(time.Date(2026, 3, 29, 9, 15, 30, 0, time.UTC)).
    ShowSecondHand(true).
    Build()
```

固定时间：

```go
ui.NewClockBuilder().
    Radius(5).
    StaticTime(time.Date(2026, 3, 29, 9, 15, 30, 0, time.UTC)).
    Preset(clock.PresetClassic).
    ASCIIHands().
    Build()
```

从 preset 克隆并局部改色：

```go
accented := clock.ThemePreset(clock.PresetClassic).
    WithDigitalStyle(style.Style{}.Foreground(style.BrightMagenta).Bold(true)).
    WithMinuteHandStyle(style.Style{}.Foreground(style.BrightMagenta).Bold(true))

ui.NewClockBuilder().
    Radius(5).
    StaticTime(time.Date(2026, 3, 29, 9, 15, 30, 0, time.UTC)).
    Theme(accented).
    Build()
```

从空主题直接挂 preset：

```go
accented := clock.Theme{}.
    WithDigitalStyle(style.Style{}.Foreground(style.BrightMagenta).Bold(true)).
    WithPreset(clock.PresetClassic)
```

组合主题层：

```go
base := clock.ThemeForPreset(clock.PresetClassic)
accent := clock.Theme{}.
    WithDigitalStyle(style.Style{}.Foreground(style.BrightMagenta).Bold(true)).
    WithSecondHandStyle(style.Style{}.Foreground(style.BrightGreen))

ui.NewClockBuilder().
    Theme(base.Merge(accent)).
    Build()
```

覆盖规则：

- 先应用默认主题
- 再应用 `Preset(...)`
- 最后应用显式 `DialStyle(...) / HourHandStyle(...) / ...`
- `Theme(...)` 是一组显式样式的批量设置，因此也会覆盖 `Preset(...)`

几何规则：

- `Radius(...)` 会同时设置横向和纵向半径
- `Ellipse()` 下可用 `RadiusX(...) / RadiusY(...) / Radii(...)` 分别控制宽高
- `Circle()` 下宽高会保持一致，纵向半径会跟随主半径
- `CellAspectX(...)` 默认是 `2.0`
- 渲染时会按 `CellAspectX(...)` 补偿终端字符单元的宽高比，所以“视觉上的正圆”在字符网格里通常会比高度更宽
- 不同终端字体下如果圆盘看起来偏瘦或偏胖，可以直接调 `CellAspectX(...)`

## 测试

- 单元测试：`go test ./ui/components/clock`
- E2E：`go test ./ui/e2e -run TestE2EClock`

## Demo

- 运行交互示例：`go run ./examples/clock_demo/`
