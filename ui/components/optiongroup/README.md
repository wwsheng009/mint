# OptionGroup

选项组组件，适合单选和多选的轻量离散选择场景。

## 已支持

- `single` / `multiple` 选择模式
- `vertical` / `horizontal` 排列
- `selected` / `selecteds`
- `label`
- `spacing`
- `disabled`
- `SelectIntent`
- `Field` 绑定

## 示例

```go
ui.OptionGroup([]ui.OptionGroupOption{
    ui.NewOptionGroupOption("beijing", "Beijing"),
    ui.NewOptionGroupOption("shanghai", "Shanghai"),
    ui.NewOptionGroupOption("shenzhen", "Shenzhen"),
}).
    Multiple().
    Horizontal().
    Spacing(2).
    Build()
```

根包也提供了类型和常量别名：

```go
ui.OptionGroupSingle
ui.OptionGroupMultiple
ui.OptionGroupVertical
ui.OptionGroupHorizontal
```
