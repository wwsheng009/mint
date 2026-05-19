# VirtualList 虚拟列表组件

`VirtualList` 是 Mint 当前的虚拟滚动列表组件，源码位于 `ui/components/virtuallist/`。它面向大量字符串项，只绘制当前可见窗口，适合日志列表、文件列表、搜索结果、长数据预览等场景。

当前实现是 Fiber-first 组件：

```text
virtuallist.VNode
  -> virtuallist.Instance
  -> Measure(...)
  -> Paint(...)
  -> ActionHandlerInstance
```

组件声明保存列表 props；运行期实例保存滚动偏移、选中项、bounds 和 dirty 状态。

## 当前源码位置

| 内容 | 路径 |
|---|---|
| Builder API | `ui/components/virtuallist/builder.go` |
| VNode props | `ui/components/virtuallist/vnode.go` |
| Runtime instance / paint / action handling | `ui/components/virtuallist/instance.go` |
| 根包快捷入口 | `ui/shortcuts.go` |
| 单元测试 | `ui/components/virtuallist/virtuallist_test.go` |
| 示例 | `examples/virtuallist` |

不要再使用历史 API `app.NewVirtualList(...)`、`layout.NewVirtualList(...)`，也不要引用旧路径 `components/layout/virtual_scroll.go`。

## API 入口

推荐从根包 `ui` 使用：

```go
items := []string{"Alpha", "Beta", "Gamma", "Delta"}

ui.NewVirtualListBuilder().
    Items(items).
    VisibleCount(3).
    ItemHeight(1).
    Size(32, 5).
    SelectedIndex(1).
    Build()
```

根包快捷函数：

```go
ui.VirtualList().
    Items(items).
    Width(40).
    Height(10).
    Build()

ui.VirtualListOfSize(items, 40, 10)
```

组件包入口：

```go
import "github.com/wwsheng009/mint/ui/components/virtuallist"

virtuallist.NewBuilder().
    Items(items).
    Width(40).
    Height(10).
    Build()
```

## Builder 方法

| 方法 | 说明 |
|---|---|
| `Items([]string)` | 设置列表项。当前 VirtualList 的主数据模型是字符串切片 |
| `ItemCount(int)` | 设置总项数；小于 `len(items)` 或小于等于 0 时会被规范化为 `len(items)` |
| `ItemHeight(int)` | 设置单项高度，当前 paint 路径按单行项绘制 |
| `VisibleCount(int)` | 设置可见项数量，决定虚拟窗口范围 |
| `Width(int)` / `Height(int)` / `Size(w, h)` | 设置组件测量和绘制尺寸 |
| `ScrollOffset(int)` | 设置当前可见窗口起始项 |
| `SelectedIndex(int)` | 设置选中项索引 |
| `AllowScroll(bool)` | 启用或禁用 action 滚动 / 导航 |
| `ListStyle(style.Style)` | 设置普通项样式 |
| `SelectedStyle(style.Style)` | 设置选中项样式 |
| `ItemStyleFn(func(int, string) style.Style)` | 按项动态样式 |
| `FgColor(...)` / `BgColor(...)` | 便捷设置普通项颜色 |
| `SelectedFgColor(...)` / `SelectedBgColor(...)` | 便捷设置选中项颜色 |
| `AddItem(string)` | 追加单个字符串项 |
| `Key(string)` | 设置 diff key |
| `SetID(string)` | 设置业务 ID，用于定位和 Portal anchoring |

`Build()` 返回 `rtui.VNode`，`BuildVNode()` 返回具体 `*virtuallist.VNode`，`BuildInstance()` 可在测试或低层场景直接构造实例。

## 渲染模型

当前 `VirtualList` 是叶子组件，不接受 child VNode：

```go
func (v *VNode) Children() []rtui.VNode {
    return []rtui.VNode{}
}
```

绘制时：

- `scrollOffset` 决定可见窗口起点。
- `visibleCount` 决定窗口终点。
- `width` 和 `height` 决定外框尺寸。
- 内容会绘制在边框内。
- 文本过长会按显示宽度截断并追加 `..`。
- `selectedIndex` 命中的项会合并 `selectedStyle`。
- `ItemStyleFn` 返回的样式会先合并到普通项样式，再叠加选中样式。

## 运行期行为

`virtuallist.Instance` 实现了 `ActionHandlerInstance`。当 `AllowScroll(true)` 时支持：

| Action | 行为 |
|---|---|
| `ActionNavigateUp` / `ActionNavigateDown` | 移动选中项 |
| `ActionNavigateHome` / `ActionNavigateEnd` | 滚动到顶部 / 底部 |
| `ActionNavigatePageUp` / `ActionNavigatePageDown` | 按 `visibleCount` 翻页 |
| `ActionSelect` | 在未选中时选中第一项；已有选中项时确认处理 |

注意：滚轮消息会先映射为 `ActionScroll`。当前 `VirtualList` 的 `HandleAction` 没有处理 `ActionScroll`，因此典型交互仍建议在父组件中把滚轮或按钮操作转成状态更新，再通过 `ScrollOffset(...)` 重新声明列表。

可用于测试或内部断言的方法包括：

```go
inst.GetVisibleRange()
inst.IsItemAtEnd()
inst.GetItem(index)
inst.GetOffset()
inst.VisibleCount()
inst.ListHeight()
inst.ListWidth()
inst.SelectedIndex()
```

## 示例：受状态驱动的列表

```go
func VirtualLog(items []string, offset, selected int) ui.VNode {
    return ui.NewVirtualListBuilder().
        Items(items).
        ItemCount(len(items)).
        Width(72).
        Height(12).
        VisibleCount(10).
        ScrollOffset(offset).
        SelectedIndex(selected).
        ListStyle(style.Style{}.Foreground(style.BrightWhite)).
        SelectedStyle(style.Style{}.Foreground(style.Black).Background(style.Yellow).Bold(true)).
        Build()
}
```

父组件可以通过 button intent、store selector、快捷键或自定义 action 更新 `offset` / `selected`，再重新 render。

## 示例：快速创建固定尺寸列表

```go
func SmallList() ui.VNode {
    return ui.VirtualListOfSize(
        []string{"Alpha", "Beta", "Gamma"},
        28,
        6,
    )
}
```

## 与 List 的关系

`ui/components/list` 提供更丰富的行模型、选择模式、checkbox、搜索和高亮等能力。需要把 List 映射为虚拟列表时，可以使用：

```go
ui.List().
    Rows(rows).
    SelectedIndex(selected).
    BuildVirtualList()
```

底层会通过 `ui/components/list/virtual_bridge.go` 生成 `virtuallist.VNode`，并保留源行索引映射。

## 性能边界

`VirtualList` 的性能收益来自只绘制 `visibleCount` 个可见项，而不是整组 items。实际内存仍取决于你传入的 `[]string`。如果数据源本身非常大，建议在父层按窗口加载或缓存字符串，再把当前数据传入组件。

当前组件没有通用的 `renderItem func(int) VNode` API；如果需要复杂行渲染，应使用 `List`、组合组件，或为特定场景新增专用组件。

## 测试与验证

相关测试：

```bash
go test ./ui/components/virtuallist -count=1
go test ./ui -run VirtualList -count=1
go test ./examples/virtuallist -count=1
```

当前文档描述的是 `ui/components/virtuallist` 的现行实现。历史文档中基于 `app.NewVirtualList(itemCount, renderItem)` 的示例不再代表当前公共 API。
