# Layout

整体页面骨架布局组件，适合应用壳、后台管理页和左右侧栏结构。

## 已支持

- `Header` / `LeftSider` / `Content` / `RightSider` / `Footer` 槽位
- `Gap` 与 `BodyGap`
- 根容器 `Width` / `Height`
- 根样式透传
- body 与 content 默认 `flex` 填充

## 示例

```go
ui.NewLayoutBuilder().
    Header(ui.Text("Header")).
    LeftSider(ui.Text("Navigation")).
    Content(ui.Text("Main Content")).
    RightSider(ui.Text("Inspector")).
    Footer(ui.Text("Footer")).
    Gap(1).
    BodyGap(2).
    Width(80).
    Height(24).
    Build()
```

如果只需要主内容区，也可以直接用快捷函数：

```go
ui.Layout(ui.Text("Main Content"))
```
