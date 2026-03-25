# Progress

进度展示组件，适合任务进度、资源占用和阶段性完成度展示。

## 已支持

- `line` / `circle` / `dashboard`
- `normal` / `success` / `exception` / `active`
- `active` 状态逐帧动画
- 自定义 `label`
- 自定义 `width`
- `showPercent`

## 示例

```go
ui.NewProgressBuilder().
    Value(60).
    Max(100).
    Label("Uploading").
    Active().
    Width(24).
    Build()
```

快捷函数：

```go
ui.Progress(60, 100)
```
