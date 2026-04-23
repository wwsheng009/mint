# Charts Internal Rendering Rules

> 这份文档记录图表内部渲染层的共同规则，重点覆盖 backlog 里仍然容易漂移的降级、裁剪和布局冲突约束。

## Render Mode 降级规则

默认原则不是“尽量高级”，而是“在当前尺寸下尽量稳定可读”。

- `sparkline` 在单行趋势场景优先使用最稳定的字符序列；只有当具体组件显式区分模式时，才切 `braille / block / ascii`
- `linechart` 优先保留折线连续性，其次才追求字符细腻度
- `barchart`、`bulletchart`、`heatmap` 优先保证量级差异和标签可读性
- 当终端宽高不足时，应优先退化信息密度而不是硬保留复杂 glyph

## Clipping 规则

- 所有 plot 绘制都必须限制在 plot area 内
- 轴线、网格线、reference line、reference band 都必须经过同一套坐标裁剪
- label 和 value text 一旦超出 plot area，应移动到 footer 或压缩，而不是写出边界
- clipping 应发生在共享层或统一 helper 中，不应让每个组件自己定义一套越界行为

## 标签与网格冲突

- 优先保留数据 glyph 和标签，其次才是网格线
- 轴标签空间不足时，先做折叠、缩写、采样
- 不要为了保留完整标签而把 plot area 压缩到失去可读性
- 网格线与标签冲突时，允许局部断开网格，不要覆盖标签文本

## 宽字符与测量

- 所有标签宽度统一使用 `paint.StringWidth()`
- 不允许用 `len()` 估算终端显示宽度
- 多行 header / footer 的最终宽度应以实际渲染宽度而不是字符数量计算

## Snapshot 稳定性

- 会显著改变字符布局的行为，应优先补独立 fixture 和 snapshot
- 不要只依赖 gallery snapshot 来验证单个图表新特性
- 对 e2e 可见性断言，优先选择稳定的文本信号，不依赖偶然空白布局
