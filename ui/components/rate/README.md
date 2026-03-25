# Rate

评分组件，适合星级反馈、满意度打分和轻量评价场景。

## 已支持

- `label`
- `value`
- `count`
- `allowClear`
- `character`
- `disabled`
- `showValue`
- `ChangeIntent`
- `Field` / `Form` 绑定

## 示例

```go
ui.NewRateBuilder().
    Label("Satisfaction").
    Count(5).
    Value(4).
    AllowClear(true).
    ShowValue(true).
    Build()
```

快捷函数：

```go
ui.Rate().
    Value(4).
    ShowValue(true).
    Build()
```
