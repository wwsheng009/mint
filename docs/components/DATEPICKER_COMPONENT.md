# DatePicker Component

`DatePicker` is a TUI date input with two entry modes:

- Direct text input in `YYYY-MM-DD`
- Popup calendar selection with keyboard or mouse

## Builder Example

```go
ui.NewDatePickerBuilder().
    SetID("ship-date").
    ComponentID("form.ship_date").
    Value("2026-04-05").
    ForField(intent.BindField("schedule.date")).
    Width(18).
    Build()
```

## Supported Interaction

- Type a date directly
- `Enter` opens the popup and confirms the highlighted day
- Arrow keys move by day or week
- `PageUp` / `PageDown` move by month
- `Home` / `End` jump to the start or end of the week
- Mouse click selects a day or changes month from the popup header
- Blur commits valid input and restores the last committed value for invalid input

## Binding

When used with `ForField(...)`, the component emits `FieldChangeIntent` with the final date string.

## Tests

- Component tests: [datepicker_test.go](/E:/projects/yao/wwsheng009/mint/ui/components/datepicker/datepicker_test.go)
- E2E tests: [datepicker_e2e_test.go](/E:/projects/yao/wwsheng009/mint/ui/e2e/datepicker_e2e_test.go)
