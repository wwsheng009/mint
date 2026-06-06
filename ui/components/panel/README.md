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
ui.TablePanelWithScope("Requests", tableNode, "sort=current page latency desc", "requests unavailable", 126)
ui.ContentPanel("Actions", actionsNode, "actions unavailable", 126)
ui.StackPanel("History", []ui.VNode{tableNode, paginationNode}, "history unavailable", 62)
ui.StackPanelWithScope("Distribution", []ui.VNode{chartNode}, "source=analytics", "distribution unavailable", 126)
ui.OperationsPanelWithScope("Runtime Operations", progressNodes, "runtime=available", "runtime unavailable", 126)
```

`TablePanelWithScope` 适合单张表需要就近说明数据范围、筛选或本地排序范围的场景，在表格下追加低强调 `Scope: ...` 行；无表格内容时仍显示标准空态，不单独渲染 scope。
`ContentPanel` 适合一个主控件或一个已经组合好的工作区节点；`StackPanel` 适合按顺序堆叠多个节点，例如表格、分页条和选择工具栏。
`StackPanelWithScope` 适合表格、图表和辅助上下文同处一个面板的场景，在已有节点后追加低强调 `Scope: ...` 行；无节点时仍显示标准空态。
`OperationsPanelWithScope` 适合运维进度/健康面板，在已有节点后追加低强调 `Scope: ...` 行，用于说明数据来源、筛选范围或诊断接口可用性；当没有可展示节点时仍显示标准空态，不单独渲染 scope。
