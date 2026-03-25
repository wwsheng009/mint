# Spin

加载指示组件，适合异步请求、局部 loading 和占位中的等待态。

## 已支持

- `small` / `default` / `large`
- `tip`
- `delay`
- `spinning` 开关
- TickFrame 动画

## 示例

```go
ui.NewSpinBuilder().
    Large().
    Tip("Loading data").
    Build()
```

也可以直接用快捷函数：

```go
ui.Spin("Loading")
```
