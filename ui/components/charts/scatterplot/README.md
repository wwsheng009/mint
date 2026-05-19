# ScatterPlot

`scatterplot` 是 charts 组件族里的二维点图组件。

它适合表达两类问题：

- 两个连续指标之间的相关性
- 多系列点云在同一坐标系里的聚集、分离和阈值关系

当前第一版范围：

- 单系列和多系列点渲染
- 基础 `title / legend / grid / axis`
- 基于线性比例尺的 `x / y` 域映射
- 支持自定义 `domain`
- 支持显式 `viewport`
- 支持可命名的 `x / y` 参考线
- 支持可命名的 `x / y` 参考区间带
- 支持密集点位冲突时的自动 collision 计数标记
- 系列级颜色与可选 `Glyph`

## 最小示例

```go
scatterplot.NewBuilder(nil).
    Title("Service Correlation").
    Series(
        scatterplot.Series{
            Name: "API",
            Points: []scatterplot.Point{
                {X: 2, Y: 3},
                {X: 4, Y: 6},
                {X: 8, Y: 9},
            },
        },
        scatterplot.Series{
            Name: "Worker",
            Glyph: '◆',
            Points: []scatterplot.Point{
                {X: 3, Y: 4},
                {X: 6, Y: 5},
                {X: 7, Y: 8},
            },
        },
    ).
    Domain(scatterplot.NewDomain(0, 10, 0, 12)).
    XReferenceLineLabeled(5, "Target").
    YReferenceBandLabeled(6, 9, "Risk").
    Width(13).
    Height(6).
    ShowLegend(true).
    ShowGrid(true).
    ShowAxis(true).
    Build()
```

当前推荐示例见 `examples/charts_gallery_demo`；历史独立示例已归档到 `docsArchive/cleanup-2026-05-19/_examples/charts_scatterplot_demo/`。

## 推荐用法

- 需要稳定坐标语义时，优先显式设置 `Domain(...)`，不要完全依赖数据最值。
- 需要局部放大时，使用 `Viewport(...)` 裁剪可见窗口；`Domain` 和 `Viewport` 不要混成一个概念。
- 需要表达阈值时，单个阈值用 `XReferenceLineLabeled(...)` / `YReferenceLineLabeled(...)`，区间阈值用 `XReferenceBandLabeled(...)` / `YReferenceBandLabeled(...)`。
- 系列较多时，优先给次级系列设置不同 `Glyph`，不要只依赖颜色区分。
- 如果预期点位会明显重叠，保留 collision 标记比强行画所有点更可读。

## 设计说明

### Domain 与 Viewport

- `Domain` 定义完整坐标系范围。
- `Viewport` 定义当前实际可见窗口。
- 终端宽高有限时，这两个层级分开能明显降低“数据一变，点位全体跳动”的问题。

### Reference Layers

- 参考线适合表达单一阈值，例如目标值、告警线、容量上限。
- 参考区间带适合表达安全区、风险区、关注区。
- 命名 reference line / band 会自动进入 legend，这一点比把阈值写进标题或 footer 稳定得多。

### Dense Data

- 当多个点映射到同一字符单元格时，组件会输出 collision 标记而不是简单覆盖。
- 这条策略的优先级高于“尽可能保留原始 glyph”，因为终端里的第一约束是可辨识性。

当前不包含：

- 交互选择
- tooltip
- trend line / regression
- 交互式缩放和平移
