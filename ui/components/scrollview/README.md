# ScrollView

文本型滚动容器组件，适合长文本、日志输出和固定视口内容区。

`ScrollView` 会抽取 child 文本并自行绘制，不保留 child tree 参与普通布局、hitmap、焦点和事件派发。包含按钮、输入、表格、tabs 等交互控件的复杂页面应使用 `PageViewport`。

## 已支持

- child 内容容器
- width / height 视口
- `scrollOffset`
- border
- scroll indicator

## 示例

```go
ui.NewScrollViewBuilder().
    Child(ui.Text("Line 1\nLine 2\nLine 3")).
    Width(30).
    Height(6).
    ShowBorder(true).
    Build()
```

快捷函数：

```go
ui.ScrollView(ui.Text("Long content"))
```
