# ScrollView

滚动容器组件，适合长文本、日志输出和固定视口内容区。

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
