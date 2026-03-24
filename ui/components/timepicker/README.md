# TimePicker

TUI 时间选择组件，支持直接输入 `HH:mm`，也支持弹出时间面板进行键盘/鼠标选择。

## 用法

```go
ui.NewTimePickerBuilder().
    SetID("ship-time").
    ComponentID("form.ship_time").
    Placeholder("HH:mm").
    ForField(intent.ForField(intent.StateKey[string]("filters.ship_time"))).
    Build()
```

## 已支持

- 手动输入 `HH:mm`
- blur 时自动把 `9:5` 规范化为 `09:05`
- 弹出时间面板
- 键盘导航：方向键 / `Home` / `End` / `Enter` / `Esc`
- 鼠标选择小时与分钟
- `FieldChangeIntent` / Form 绑定
