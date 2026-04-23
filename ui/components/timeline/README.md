# Timeline

`timeline/` 提供纵向时间轴组件，适合事件流、发布记录和操作历史展示。

支持：

- `label` / `content` / `description`
- `status` 与自定义 `dot`
- `pending` 尾项
- `reverse` 反转顺序

## 示例

```go
ui.NewTimelineBuilder().
    Item(
        timeline.Event("Build completed").
            WithLabel("09:30").
            WithDescription("CI pipeline finished").
            WithStatus(timeline.StatusSuccess),
    ).
    Item(
        timeline.Event("Deploy started").
            WithLabel("09:45"),
    ).
    Pending("Waiting for smoke tests").
    Build()
```
