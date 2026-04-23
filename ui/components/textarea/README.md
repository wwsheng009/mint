# Textarea

多行文本输入组件，适合备注、描述和大段文本编辑。

## 已支持

- `rows` / `cols`
- `maxLen`
- 滚动偏移与 scrollbar
- 光标样式配置
- `Field` / `Form` 绑定
- `OnSubmit`

## 示例

```go
ui.NewTextareaBuilder().
    Placeholder("Enter details").
    Rows(5).
    Cols(40).
    ShowScrollbar(true).
    Build()
```

快捷函数：

```go
ui.Textarea("Description")
```
