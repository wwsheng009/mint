# Drawer

侧边抽屉组件，适合设置面板、详情预览和边缘滑出式操作区。

## 已支持

- left / right / top / bottom placement
- 受控 open 状态
- header / content / footer
- ESC / backdrop 关闭
- close intent
- border / shadow 样式

## 状态语义

- 当前 API 以受控 `Open(true/false)` 为主；父级负责决定抽屉是否显示。
- `CloseOnEsc(...)` 和 `CloseOnBackdrop(...)` 分别控制键盘 ESC 与遮罩点击关闭。
- `OnClose(...)` 负责把关闭动作抛给外层；不需要额外安装中间件。
- `Placement(...)`、`Width(...)`、`Height(...)` 共同决定抽屉从哪一侧进入以及可见区域大小。

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

## 测试入口

- 单测：`go test ./ui/components/drawer`
- 重点覆盖：`drawer_test.go` 中的 placement、open/close、ESC / backdrop 关闭链路和测量行为
- 当前尚无 dedicated `ui/e2e` 抽屉用例
