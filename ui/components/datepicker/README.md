# DatePicker

TUI 日期选择组件，支持直接输入 `YYYY-MM-DD`，也支持弹出日历进行键盘/鼠标选择。

## 用法

```go
ui.NewDatePickerBuilder().
    SetID("ship-date").
    ComponentID("form.ship_date").
    Placeholder("YYYY-MM-DD").
    ForField(intent.ForField(intent.StateKey[string]("filters.ship_date"))).
    Build()
```

## 已支持

- 手动输入 `YYYY-MM-DD`
- 弹出月视图日历
- 键盘导航：方向键 / `Home` / `End` / `PageUp` / `PageDown`
- 鼠标选择日期与切换月份
- `FieldChangeIntent` / Form 绑定
