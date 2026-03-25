# Checkbox

勾选组件，提供独立 `Checkbox` 和 `CheckboxGroup` 两种用法。

## 已支持

- 独立 checkbox
- `CheckboxGroup`
- `indeterminate`
- horizontal / vertical group 布局
- `Field` / `Form` 绑定
- 组件级 toggle intent

## 示例

```go
ui.NewCheckboxBuilder().
    Label("Accept Terms").
    Checked(true).
    Build()
```

分组场景：

```go
ui.NewCheckboxGroupBuilder([]ui.CheckboxOption{
    ui.NewCheckboxOption("a", "Option A"),
    ui.NewCheckboxOption("b", "Option B"),
}).
    Selecteds([]string{"a"}).
    Build()
```
