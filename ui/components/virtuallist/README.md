# VirtualList

虚拟列表组件，适合长列表、日志流和只渲染可见区的场景。

## 已支持

- `items`
- `itemCount`
- `itemHeight`
- `visibleCount`
- `width / height / size`
- `scrollOffset`
- `selectedIndex`
- `allowScroll`
- `listStyle` / `selectedStyle`

## 示例

```go
ui.VirtualList().
    Items([]string{"Alpha", "Beta", "Gamma", "Delta"}).
    VisibleCount(3).
    ItemHeight(1).
    Size(28, 6).
    SelectedIndex(1).
    Build()
```

快捷函数：

```go
ui.VirtualListOfSize([]string{"Alpha", "Beta", "Gamma"}, 28, 6)
```
