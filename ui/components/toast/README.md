# Toast

轻量通知组件，适合保存成功、错误提醒和短时状态反馈。

## 已支持

- `title`
- `message`
- `info / success / warning / error`
- `duration`
- `closeIntent`
- `style`
- `padding`
- `base / overlay / modal / tooltip / inspector` layer

## 示例

```go
ui.NewToastBuilder("Saved successfully").
    Title("Profile").
    Success().
    Duration(2 * time.Second).
    OverlayLayer().
    Build()
```

快捷函数：

```go
ui.Toast("Saved successfully")
ui.ToastSuccess("Saved successfully")
ui.ToastError("Save failed")
```
