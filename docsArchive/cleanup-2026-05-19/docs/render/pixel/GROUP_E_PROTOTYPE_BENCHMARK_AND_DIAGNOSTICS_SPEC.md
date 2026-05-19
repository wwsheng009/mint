# Group E Prototype、Benchmark 与 Diagnostics 规格

## 1. 文档目的

Group E 的职责不是“再补一份说明文档”，而是定义：

> 第一阶段 image prototype 应该如何被演示、测量、诊断与止损。

如果没有这一层，Phase 1 很容易退化成：

- 觉得图更好看
- 觉得可能更慢
- 但没有客观证据

## 2. Group E 的核心目标

第一阶段 Group E 必须输出三类东西：

1. 一个可以运行的 prototype 页面
2. 一组最小 benchmark 指标
3. 一组可保存的 diagnostics 产物

缺一不可。

## 3. Prototype 页面要求

### 3.1 单独示例目录

建议目录：

- `examples/charts_linechart_image_prototype/`

不要直接塞进现有 gallery。

### 3.2 Prototype 页面最低展示内容

建议一页内至少包含：

- 文本 `linechart`
- image `linechart`
- 当前图形能力摘要
- 当前 backend 选择结果

也就是说，prototype 页面本身就应该是“对照实验”。

### 3.3 推荐页面布局

建议最小布局：

1. 标题区
2. 左侧文本图
3. 右侧 image 图
4. 底部 diagnostics / benchmark 摘要

这样便于：

- 肉眼对比
- 截图
- 后续写报告

## 4. Benchmark 的最小指标

### 4.1 必须采集的时间指标

- 首帧渲染时间
- 单次数据更新后渲染时间
- resize 后首帧时间

### 4.2 必须采集的输出指标

- 文本输出字节数
- 图像提交次数
- 图像 payload 大小

### 4.3 必须采集的缓存指标

如果第一阶段已有缓存能力，至少记录：

- image cache hit
- image cache miss

如果第一阶段尚未落缓存，也要在 benchmark 结果里明确写出“当前未启用缓存”。

## 5. Diagnostics 产物要求

### 5.1 建议保存的产物

至少建议保存：

- capability dump
- benchmark 结果
- image layer 元数据
- 失败时的错误信息

### 5.2 如果能保存 bitmap，更好

第一阶段建议允许 prototype 把离屏 raster 结果导出为：

- PNG
- 或原始像素 dump

这样能帮助排查：

- 是 raster 出错
- 还是终端协议提交出错

### 5.3 建议的输出目录

第一阶段可用实验性目录，例如：

- `tmp/pixel_prototype/`
- `artifacts/pixel_prototype/`

不建议一开始就塞进现有 e2e testdata。

## 6. 推荐的 benchmark 运行方式

### 6.1 先做 demo 内嵌 benchmark

第一阶段最稳的方式不是先写复杂 benchmark 框架，而是：

- prototype 页面运行时输出最小计时信息

这样更容易先拿到真实数据。

### 6.2 后续再考虑标准 benchmark 命令

等原型稳定后，再拆成：

- `go test -bench`
- 或独立 benchmark 命令

不要把这一步前置。

## 7. 与 CI / e2e 的关系

### 7.1 第一阶段不要求完整 image CI golden

原因：

- 当前 CI 更擅长文本 snapshot
- image golden 会引入额外环境依赖

### 7.2 第一阶段建议的自动化边界

建议自动化只覆盖：

- text fallback 路径
- capability 逻辑
- prototype 是否能启动
- diagnostics 是否能生成最小元数据

### 7.3 视觉收益先允许人工对照

第一阶段视觉对照可接受：

- 本地运行
- 截图
- 文档记录

这比过早搭完整 image CI 更现实。

## 8. Go / No-Go 判断矩阵

### 8.1 Go 条件

满足以下条件才建议继续往主线推进：

- image plot 视觉收益明显
- 文本路径不受破坏
- prototype 能稳定重绘与 resize
- capability 与 fallback 行为稳定

### 8.2 No-Go 条件

出现以下情况应暂停扩大范围：

- 单图收益不明显
- 输出/重绘成本明显失控
- alternate screen 清理不稳定
- diagnostics 无法定位问题

## 9. Prototype 文档要求

每个 prototype 页面应至少附带：

- 运行前提
- 目标终端
- 启动命令
- 已知限制
- 当前未做事项

这样后续评审不会把 PoC 误读成生产方案。

## 10. 与现有方案文档的关系

Group E 不是替代：

- `PERFORMANCE_IMPACT_AND_OPTIMIZATION_PLAN.md`

而是把那份文档中的性能观点，压缩成第一阶段真正要执行的验证闭环。

## 11. 一句话结论

Group E 的价值不在“多一个 demo”，而在：

**让 image prototype 不再停留在主观感受层，而是拥有可运行、可对照、可测量、可止损的验证闭环。**

## 12. 下一步实现参考

如果要从 Group E 规格继续推进到具体仓库落位，建议紧接着参考：

- [EXAMPLES_LINECHART_IMAGE_PROTOTYPE_LAYOUT_SPEC.md](./EXAMPLES_LINECHART_IMAGE_PROTOTYPE_LAYOUT_SPEC.md)
- [PROTOTYPE_BENCHMARK_AND_ARTIFACT_LAYOUT_SPEC.md](./PROTOTYPE_BENCHMARK_AND_ARTIFACT_LAYOUT_SPEC.md)

它们分别把 Group E 再向前推进成：

- prototype 示例目录与页面布局规范
- benchmark / diagnostics 产物的目录、命名与入库边界规范
