# Absolute

绝对定位容器，适合角标、浮层锚点和局部自由排布场景。

## 已支持

- `left / top / right / bottom`
- `anchor`
- `zIndex`
- `width / height / size`
- 样式和边框配置
- 便捷定位函数

## 示例

```go
ui.NewAbsoluteBuilder(ui.Text("Badge")).
    Right(ui.AbsolutePos(0)).
    Top(ui.AbsolutePos(0)).
    Anchor(ui.AnchorTopRight).
    RoundedBorder("NEW").
    Build()
```

快捷函数：

```go
ui.At(ui.Text("Center"), 10, 4)
ui.TopLeft(ui.Text("TL"))
ui.BottomRight(ui.Text("BR"))
ui.CenterAbs(ui.Text("Centered"))
```
