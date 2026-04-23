# Radio

单选组件，提供独立 `Radio` 和 `RadioGroup` 两种用法，适合互斥选择场景。

## 已支持

- 独立 radio
- `RadioGroup`
- horizontal / vertical 布局
- 受控 `checked` / `selected`
- `Field` / `Form` 绑定
- 组件级选择 intent

## 示例

```go
ui.NewRadioGroupBuilder([]ui.RadioOption{
    ui.NewRadioOption("a", "Option A"),
    ui.NewRadioOption("b", "Option B"),
}).
    Label("Pick one").
    Selected("b").
    Vertical().
    Build()
```

独立单项也可以直接构建：

```go
ui.NewRadioBuilder().
    Label("Newsletter").
    Checked(true).
    Build()
```
