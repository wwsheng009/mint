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

## 示例

```go
ui.TreeView().
    Nodes([]ui.TreeNode{
        {NodeID: 1, Content: "src", NodeType: "folder"},
        {NodeID: 2, Content: "main.go", NodeType: "file", Indent: 2},
    }).
    ExpandLevel(1).
    ShowIcons(true).
    Build()
```
