# Interactive TreeView Demo

运行：

```bash
go run ./examples/fiber_firsts/treeview_demo
```

这个示例展示了：

- `SearchQueryControlled` + `SearchMatchPathsControlled(...)` + `SearchPending(...)` 的异步受控搜索
- 当前 match 驱动的搜索结果分页（`SearchPageSize`）
- `SelectedIndexControlled` + `ExpandedPaths(...)` 受控导航与展开状态
- `SelectionForField` + `CheckedPaths(...)` 受控多选
- 自定义全局 demo intent 驱动“上一/下一匹配”“展开全部/收起全部”
- `OnLazyLoad` + `LazyLoadSuccess/FailureIntent` 同步/异步懒加载与错误重试
- `RowStyleFn`、`MatchStyle`、`ShowSearchStats`、`Compact`、`ShowLineNums` 等配置

交互提示：

- `Tab` 切换搜索框和 TreeView 焦点
- 输入搜索词后会先进入异步 `pending`，结果返回后再更新高亮和当前结果页
- TreeView 聚焦后可用 `↑ ↓ ← →`、`Home`、`End`、`PageUp`、`PageDown`
- `Enter` / `Space` 切换 checkbox 选择
- 父节点 checkbox 会联动整个子树
- `r` 或聚焦 TreeView 后按 `F5` 重试当前 lazy/error 节点
- `F2` / `F3` 上一条 / 下一条匹配，搜索结果页会随当前 match 自动翻页
- `F4` / `F6` 全展开 / 全折叠
- `F7` 切换异步 lazy load 成功或失败模式
- `F8` / `F9` / `F10` 切换 `none` / `single` / `multi`
- `F11` 清空勾选
- `F12` 重置示例
- `Ctrl+L` 清空搜索
