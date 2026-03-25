# Drawer

侧边抽屉组件，适合设置面板、详情预览和边缘滑出式操作区。

## 已支持

- left / right / top / bottom placement
- 受控 open 状态
- header / content / footer
- ESC / backdrop 关闭
- close intent
- border / shadow 样式

## 示例

```go
ui.NewDrawerBuilder().
    Title("Settings").
    Content(ui.Text("Drawer body")).
    Placement(ui.DrawerRight).
    Width(32).
    Opened().
    Build()
```

快捷函数：

```go
ui.Drawer(ui.Text("Drawer body"))
ui.DrawerTitled("Settings", ui.Text("Drawer body"))
```
