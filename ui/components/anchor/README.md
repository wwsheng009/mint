# Anchor

锚点导航组件，适合文档目录、设置页分组导航和长表单章节索引。

## 已支持

- 层级 links / 子项缩进展示
- 当前项高亮
- 键盘与鼠标切换
- `activeKey` 受控 / 非受控模式
- `Field` / `Form` 绑定
- 组件级 `ChangeIntent`

## 示例

```go
ui.NewAnchorBuilder().
    Title("Contents").
    Width(24).
    ViewportHeight(8).
    InitialActiveKey("api").
    Items([]ui.AnchorItem{
        ui.NewAnchorItem("intro", "Introduction"),
        ui.NewAnchorItem("guide", "Guide",
            ui.NewAnchorItem("api", "API"),
        ),
    }).
    Build()
```

也可以直接用快捷函数：

```go
ui.AnchorNav([]ui.AnchorItem{
    ui.NewAnchorItem("intro", "Introduction"),
    ui.NewAnchorItem("api", "API"),
})
```
