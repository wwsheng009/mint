# Panel

容器面板组件，适合卡片式布局、设置分组和带标题的内容块。

## 已支持

- 标题 / header / content / footer
- width / height / flex
- padding
- 边框样式与颜色
- border label

## 示例

```go
ui.NewPanelBuilder().
    Title("Profile").
    Content(ui.Text("Body")).
    Footer(ui.Text("Footer")).
    Width(40).
    Rounded().
    Build()
```

快捷函数：

```go
ui.Panel(ui.Text("Body"))
ui.PanelTitled("Profile", ui.Text("Body"))
```
