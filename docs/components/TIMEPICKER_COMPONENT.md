# TimePicker Component

`TimePicker` is a TUI time input for `HH:mm` values with direct typing and popup selection.

## Builder Example

```go
ui.NewTimePickerBuilder().
    SetID("ship-time").
    ComponentID("form.ship_time").
    Value("09:30").
    ForField(intent.BindField("schedule.time")).
    Width(10).
    Build()
```

## Supported Interaction

- Type time directly in `HH:mm`
- Blur normalizes partial valid input like `9:5` to `09:05`
- `Enter` opens the popup and confirms the highlighted time
- `Left` / `Right` switch hour and minute segment
- `Up` / `Down` change the active segment
- `Home` / `End` jump to segment min or max
- `PageUp` / `PageDown` adjust the hour
- Mouse click picks hour and minute from the popup list

## Binding

When used with `ForField(...)`, the component emits `FieldChangeIntent` with the normalized time string.

## Tests

- Component tests: [timepicker_test.go](/E:/projects/yao/wwsheng009/mint/ui/components/timepicker/timepicker_test.go)
- E2E tests: [timepicker_e2e_test.go](/E:/projects/yao/wwsheng009/mint/ui/e2e/timepicker_e2e_test.go)
