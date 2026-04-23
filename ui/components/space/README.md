# Space

统一子元素间距的轻量布局组件，适合按钮组、工具栏和表单内联区块。

## 已支持

- `horizontal` / `vertical`
- `small` / `middle` / `large`
- `wrap`
- `split`
- cross-axis `align`
- `width`

## 示例

```go
ui.NewSpaceBuilder().
    Large().
    Split("|").
    Children(
        ui.Text("Build"),
        ui.Text("Test"),
        ui.Text("Deploy"),
    ).
    Build()
```

需要自动换行时可以配合 `Wrap(true)` 和 `Width(...)` 使用，也可以直接用 `ui.Space(...)` 快捷函数。
