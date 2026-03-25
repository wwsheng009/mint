# Divider

分隔线组件，适合区块分组、列表分节和垂直内容间隔。

## 已支持

- horizontal / vertical
- `solid` / `dashed` / `dotted` / `double`
- 中心 label
- thickness
- `fillWidth`

## 示例

```go
ui.NewDividerBuilder().
    Label("Section").
    Double().
    Build()
```

快捷函数：

```go
ui.Divider()
ui.HDivider()
ui.VDivider()
```
