# Skeleton

骨架屏占位组件，适合在异步数据加载时给出稳定布局预览。

## 用法

```go
ui.NewSkeletonBuilder().
    Avatar(true).
    AvatarShape(ui.SkeletonShapeRound).
    TitleWidth(18).
    ParagraphRows(3).
    ParagraphWidths(28, 28, 16).
    Content(
        ui.NewTextBuilder("Loaded profile").Build(),
    ).
    Build()
```

## 已支持

- `loading` 占位 / 实际内容切换
- 头像占位：`square` / `round`
- 标题与段落行数、宽度配置
- `active` 强化样式
- 根容器样式与占位样式覆写
