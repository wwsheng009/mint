# Chart Raster

计划中的字符栅格化包。

负责：

- 点到 cell 的映射
- 线段离散化
- 柱形填充
- marker 绘制

它和 `canvas/` 的关系是：

- `raster/` 负责“怎么把几何对象离散化”
- `canvas/` 负责“怎么把离散结果写成终端字符”

