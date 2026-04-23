# store_mixed_demo

运行：

```bash
go run ./examples/store_mixed_demo/
```

## 光标说明

本示例中的 `Input` 已切换为插入态细光标（竖线），配置方式如下：

```go
ui.NewInputBuilder().
    Placeholder("用户名").
    InsertCursor().
    Build()
```

也可以进一步配置闪烁：

```go
ui.NewInputBuilder().
    InsertCursor().
    CursorBlinkInterval(350 * time.Millisecond).
    Build()
```
