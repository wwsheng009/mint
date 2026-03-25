# Breadcrumb

面包屑导航组件，适合路径展示、层级导航和窄宽度下的当前位置提示。

## 已支持

- 自定义 item 列表
- 当前项高亮
- 自定义分隔符
- 窄宽度下左侧折叠
- item / current / separator 独立样式

## 示例

```go
ui.NewBreadcrumbBuilder().
    Labels("Home", "Workspace").
    Item(ui.BreadcrumbItem{
        Label:   "Breadcrumb",
        Current: true,
    }).
    Separator(" > ").
    MaxWidth(24).
    Build()
```

也可以用快捷函数直接从 items 构建：

```go
ui.Breadcrumb([]ui.BreadcrumbItem{
    {Label: "Home"},
    {Label: "Docs"},
    {Label: "API", Current: true},
})
```
