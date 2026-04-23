# Notification

通知弹窗组件，适合保存成功、异步任务完成和全局状态提示。

## 已支持

- `info` / `success` / `warning` / `error`
- 标题与正文
- `closable`
- `duration`
- `placement`
- `closeIntent`

## 示例

```go
ui.NewNotificationBuilder("Settings saved").
    Title("Success").
    Success().
    Placement(ui.NotificationTopRight).
    Build()
```

如果只需要一条默认通知，也可以直接用：

```go
ui.Notification("Settings saved")
```
