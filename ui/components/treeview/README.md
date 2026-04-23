# TreeView

树形视图组件，适合文件树、导航树和分层选择场景。

## 已支持

- 展开 / 折叠
- lazy load
- 搜索与搜索结果分页
- 单选 / 多选
- 同父级 subtree 拖拽排序
- checkbox selection
- `Field` / selection binding

## 状态语义

- `ExpandLevel(...)` 适合声明默认展开层级；需要外部受控展开时使用 `ExpandedKeys(...)` 或 `ExpandedPaths(...)`。
- 搜索既支持组件内部过滤 `SearchQuery(...)`，也支持外部受控 `SearchQueryControlled(...)` + `SearchMatchesControlled(...)` / `SearchMatchPathsControlled(...)`。
- checkbox 选择可通过 `SelectionMode(...)`、`CheckedKeys(...)`、`CheckedPaths(...)`、`OnSelectionChange(...)`、`SelectionForField(...)` 接到业务状态。
- 懒加载可用 `OnLazyLoad(...)` 或 `OnLazyLoadChildren(...)`；同父级拖拽排序需要显式打开 `Reorderable(true)` 并处理 `OnReorder(...)`。

## 示例

```go
ui.TreeView().
    ComponentID("project.tree").
    Nodes([]ui.TreeNode{
        {NodeID: 1, Content: "src", NodeType: "folder"},
        {NodeID: 2, Content: "main.go", NodeType: "file", Indent: 2},
    }).
    ExpandLevel(1).
    ShowIcons(true).
    Build()
```

## 测试入口

- 单测：`go test ./ui/components/treeview`
- 重点覆盖：`treeview_test.go` 中的搜索分页、受控展开 / 选择、lazy load、search stats 与 drag reorder
- E2E：`go test ./ui/e2e -run TestE2ETreeView`
