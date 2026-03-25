# Empty

空状态组件，适合列表为空、过滤无结果和占位页场景。

## 已支持

- 自定义 description
- 自定义 ASCII image
- 样式覆盖

## 示例

```go
ui.NewEmptyBuilder().
    Description("No records found").
    Image("[ ]").
    Build()
```

快捷函数：

```go
ui.Empty("No records found")
```
