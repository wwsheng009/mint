# Select

选择器组件，适合单选、多选、tags 输入和分组选项场景。

## 已支持

- 单选 / 多选 / `tags`
- `OptGroup`
- `filterOption`
- `placeholder`
- `disabled`
- overlay popup
- `Field` / `Form` 绑定

## 示例

```go
ui.NewSelectBuilder().
    Options([]ui.SelectOption{
        ui.NewSelectOption("beijing", "Beijing"),
        ui.NewSelectOption("shanghai", "Shanghai"),
    }).
    Placeholder("Select city").
    FilterOption(true).
    Build()
```

需要标签模式时可直接打开 `TagsMode(true)`。
