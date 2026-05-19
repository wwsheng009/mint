# Inspector TreeView Auto-Sizing

本文记录 Inspector TreeView 自动尺寸与滚动的当前状态。早期实现说明曾把 TreeView 包在 ScrollView 中，让 ScrollView 提供固定 viewport；当前源码已经以 `ui/components/treeview` 自身的测量、绘制和 action 状态为主。

## 当前结论

- Inspector overlay 本身仍是固定尺寸面板。
- Elements tab 使用 `VStackBuilder` 组合 header、selected info、TreeView、instructions。
- 父容器通过 `.Width(...)`、`.Height(...)` 和 `.Flex(1)` 向子树传递布局约束。
- TreeView 当前由 `componenttreeview.NewBuilder().FromLines(...).Build()` 创建。
- TreeView instance 自己维护滚动偏移、选中项、可见范围和滚动条。

## 当前源码关系

```text
internal/inspector/standalone_inspector.go
  -> componenttreeview.NewBuilder().FromLines(si.treeLines)
  -> ui/components/treeview.VNode
  -> ui/components/treeview.Instance
  -> Measure(layout.Constraints)
  -> Paint(x, y)
```

`ui/components/scrollview` 仍存在，但不是当前 Inspector Elements tab 的主 wrapper。

## 当前代码形态

```go
treePreview := componenttreeview.NewBuilder().
    FromLines(si.treeLines).
    ExpandLevel(-1).
    ShowIcons(true).
    Compact(false).
    Build()

return ui.VStackBuilder(
    header,
    selectedInfo,
    treePreview,
    instructions,
).
    Width(si.overlayWidth - 4).
    Height(si.overlayHeight - 4).
    Flex(1).
    Build()
```

## TreeView 滚动来源

TreeView 组件 props / state 中包含：

| 字段 | 说明 |
|---|---|
| `scrollOffset` | 当前滚动偏移 |
| `scrollOffsetControlled` | 是否由外部控制滚动偏移 |
| `selectedIndex` | 当前选中行 |
| `viewportHeight` | 可见高度 |
| `showScrollbar` | 是否绘制滚动条 |
| `allowScroll` | 是否允许滚动 action |

运行期 `Instance` 会根据这些值和布局 bounds 绘制可见行，并处理导航、选择、展开折叠等 action。

## 与 ScrollView 的边界

| 场景 | 当前推荐 |
|---|---|
| Inspector Elements 树 | `ui/components/treeview` |
| 通用长文本、日志、代码片段 | `ui/components/scrollview` |
| 大量字符串项列表 | `ui/components/virtuallist` |
| 富列表行、选择模式、搜索和 checkbox | `ui/components/list` |

不要再把 TreeView 的主滚动行为描述为 `components/layout/scroll_view.go`。如果需要通用 ScrollView，请使用 `ui.NewScrollViewBuilder()` 或 `ui.ScrollBordered(...)`。

## 验证

```bash
go test ./internal/inspector -run "TreeView|Inspector|Flex|Scroll" -count=1
go test ./ui/components/treeview -count=1
go test ./ui/components/scrollview -count=1
```

这些测试分别覆盖 Inspector 集成、TreeView 自身行为，以及通用 ScrollView 组件。

## 维护约束

- 文档中的 Inspector TreeView 示例应引用 `componenttreeview.NewBuilder()` 或 `ui/components/treeview`。
- 不要再引用 `components/layout/scroll_view.go`。
- 不要把历史 line number 当作当前源码位置；应以当前文件搜索结果为准。
