# Inspector Flex 布局实现

本文记录 Inspector 当前源码中的 Elements 面板布局方式。早期实现文档曾描述 `layout.NewScrollView(...)` 包装 TreeView 的方案；当前 `internal/inspector/standalone_inspector.go` 已改为直接使用 `ui/components/treeview`，由父级 `VStackBuilder` 提供宽高约束。

## 当前实现位置

| 功能 | 当前源码 |
|---|---|
| Inspector overlay | `internal/inspector/standalone_inspector.go` |
| Inspector 内部树数据 | `internal/inspector/tree_view.go` |
| UI TreeView 组件 | `ui/components/treeview/` |
| TreeView builder | `ui/components/treeview/builder.go` |
| TreeView instance / measure / paint | `ui/components/treeview/instance.go` |

## 当前布局模型

Elements tab 当前结构简化如下：

```text
VStackBuilder(
  header,
  selectedInfo,
  treePreview,
  instructions,
)
  .Width(overlayWidth - 4)
  .Height(overlayHeight - 4)
  .Flex(1)
```

其中 `treePreview` 由 `ui/components/treeview` 构造：

```go
treePreview := componenttreeview.NewBuilder().
    FromLines(si.treeLines).
    ExpandLevel(-1).
    ShowIcons(true).
    Compact(false).
    Build()
```

当前代码没有在该位置创建 `ScrollView` wrapper，也不再使用历史 API `layout.NewScrollView`。

## TreeView 的职责

`ui/components/treeview.Instance` 是当前树视图的运行期实体，负责：

- 保存 `scrollOffset`、`selectedIndex`、展开状态和 checked state。
- 根据 `viewportHeight` / layout bounds 计算可见范围。
- 实现 `rtui.FocusableInstance`，参与 Tab 焦点系统。
- 实现 `rtui.ActionHandlerInstance`，处理导航、选择、展开折叠等 action。
- 实现 `Measure(layout.Constraints)` 和 `Paint(x, y)`。

这意味着当前 TreeView 本身具备滚动与交互能力，不需要再通过旧 `components/layout/scroll_view.go` 提供主滚动容器。

## 为什么不再写旧 ScrollView 路径

历史文档中的：

```go
scrollContainer := layout.NewScrollView(treePreview)
```

对应旧目录设计。当前组件目录是：

```text
ui/components/scrollview/
```

但 Inspector Elements tab 的当前源码并没有使用它包裹 TreeView。ScrollView 仍可用于长文本、日志、代码片段等文本内容场景；TreeView 这类有选择、展开、搜索、滚动状态的组件应优先使用自己的 instance 行为。

## 调试与测试

建议优先运行当前 Inspector / TreeView 相关测试：

```bash
go test ./internal/inspector -run "TreeView|Inspector|Flex|Scroll" -count=1
go test ./ui/components/treeview -count=1
```

如果排查布局约束，重点看：

- `standalone_inspector.go` 中 overlay 宽高传入。
- `VStackBuilder(...).Width(...).Height(...).Flex(1)`。
- `ui/components/treeview.Instance.Measure` 是否收到期望 constraints。
- `TreeView` 的 `viewportHeight`、`scrollOffset` 和 `selectedIndex` 是否同步。

## 维护约束

- 不要再把 `components/layout/scroll_view.go` 写作当前实现路径。
- 不要把 Inspector 当前 TreeView 主路径描述为 ScrollView wrapper。
- 若文档讨论通用 ScrollView，请引用 `ui/components/scrollview/` 和 `ui.NewScrollViewBuilder()`。
