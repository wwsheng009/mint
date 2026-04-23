# Slider

滑块输入组件，适合音量、阈值和数值区间内的连续调节场景。

## 已支持

- `label`
- `value / min / max / step`
- `width`
- `disabled`
- `showValue`
- `ChangeIntent`
- `Field` / `Form` 绑定

## 示例

```go
ui.NewSliderBuilder().
    Label("Volume").
    Min(0).
    Max(100).
    Step(5).
    Value(40).
    Width(24).
    ShowValue(true).
    Build()
```

快捷函数：

```go
ui.Slider().
    Label("Volume").
    Value(40).
    Build()
```
