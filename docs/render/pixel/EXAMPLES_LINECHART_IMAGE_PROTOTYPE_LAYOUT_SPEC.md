# `examples/charts_linechart_image_prototype` 目录与页面布局规格

> 状态更新（2026-04-01）  
> 本文档描述的是最初为 chart 像素渲染原型设计的页面与目录方案。当前仓库已经收口到新的策略：`linechart` 组件层不再接入像素绘制，chart 继续走文本渲染；终端图形能力仅保留给专用图片控件。  
> 因此本文件现在更适合作为历史设计记录，而不是当前 chart 产品行为的实现准则。

## 1. 文档目的

前面的文档已经把：

- Group A 的能力层
- Group B/C 的 Scene 与 `App.render()` 接入
- Group D 的 `linechart` image renderer
- Group E 的 prototype / benchmark / diagnostics 目标

分别拆开了。

但如果真正进入实现，还需要把 prototype 本身的仓库落位方式定清楚：

> `examples/charts_linechart_image_prototype/` 目录到底应该怎么组织，页面该展示什么，和现有 demo/e2e 体系怎么衔接。

本文件就是这份落位规格。

## 2. 为什么需要单独 prototype 目录

当前仓库里已经有：

- [examples/charts_linechart_demo](/E:/projects/yao/wwsheng009/mint/examples/charts_linechart_demo)
- [ui/e2e/charts_linechart_demo_e2e_test.go](/E:/projects/yao/wwsheng009/mint/ui/e2e/charts_linechart_demo_e2e_test.go)

这条线已经很好地承载了：

- 纯文本 `linechart`
- 轴标签模式
- 文本 snapshot 回归

因此 image prototype 不应该直接混进现有 demo。原因有三点：

1. 现有 `charts_linechart_demo` 仍应代表稳定的文本主线
2. image prototype 在第一阶段是实验性能力，不适合污染现有展示入口
3. 它需要额外 diagnostics、benchmark 和能力信息区，这些都不属于普通 demo

## 3. 推荐目录结构

建议严格对齐当前 `examples/charts_linechart_demo` 的仓库风格，只做必要扩展。

推荐目录：

```text
examples/charts_linechart_image_prototype/
├── README.md
├── main.go
├── main_test.go
└── demo/
    ├── demo.go
    ├── model.go
    └── benchmark_notes.go
```

### 3.1 `main.go`

职责：

- 保持为薄入口
- 只负责 `ui.Run(...)`
- 设置固定窗口尺寸
- 调用 `demo.Build()`

这样和现有：

- [examples/charts_linechart_demo/main.go](/E:/projects/yao/wwsheng009/mint/examples/charts_linechart_demo/main.go)

保持一致。

### 3.2 `demo/demo.go`

职责：

- 承载 prototype 页面的实际视图构建
- 供 `main.go` 和后续 e2e / smoke test 复用

这和当前 `charts_linechart_demo` 里 `demo/Build()` 的组织方式一致，应该延续。

### 3.3 `demo/model.go`

职责：

- 存放 prototype 使用的固定数据集
- 能同时支撑：
  - 文本 baseline
  - image baseline
  - 小尺寸和中尺寸场景

建议不要把数据散落在 `demo.go` 里。

### 3.4 `demo/benchmark_notes.go`

职责：

- 定义 prototype 页面底部要显示的 benchmark 字段标签
- 或提供 diagnostics 字段摘要格式

第一阶段可以非常轻，不一定要复杂逻辑，但最好不要全部写死在页面模板中。

### 3.5 `main_test.go`

职责：

- 最小 smoke test
- 确保 `Build()` 可运行
- 确保 prototype 入口不会在最基础层面编译失败

第一阶段它不承担完整视觉回归职责。

## 4. 页面应展示的最小内容

第一阶段 prototype 页面不应追求“像正式 demo 一样完整”，而应追求“能做实验对照”。

建议页面固定包含 5 块区域。

### 4.1 标题区

至少显示：

- `Charts LineChart Image Prototype`
- 当前实验目标简述

例如：

- `Compare text and image rendering for the same line chart`

### 4.2 能力摘要区

至少显示：

- graphics mode
- capability reliability
- cell pixel size
- 当前 backend 选择结果

这块内容非常关键，因为 prototype 页本质上是实验页面，不是普通组件展示页。

### 4.3 文本对照图

必须显示一张文本 `linechart`：

- 同一份数据
- 同一逻辑尺寸
- 尽量同一 plot area

这是所有视觉收益判断的基线。

### 4.4 图像对照图

必须显示一张 image `linechart`：

- 和文本图使用同一份数据
- 和文本图尺寸可比
- 最好是并排展示

### 4.5 底部诊断摘要区

至少显示：

- 首帧渲染时间
- 最近一次更新时间
- image payload 大小
- 当前是否启用缓存

注意：

- 这里显示的是摘要，不要求在页面上展示完整 diagnostics
- 完整 diagnostics 应导出到单独 artifact

## 5. 推荐页面布局

### 5.1 第一阶段推荐布局

推荐单屏布局如下：

1. 顶部：标题 + 简述
2. 第二行：capability / backend 摘要
3. 中部左右：`Text` vs `Image`
4. 底部：benchmark / diagnostics 摘要

### 5.2 为什么不推荐复杂多图布局

第一阶段 prototype 的任务不是做 gallery，而是做实验对照。

如果页面里同时塞入：

- 多系列
- 多尺寸
- 多 backend
- 多终端能力状态

会把原型目标稀释掉。

## 6. 推荐固定尺寸

第一阶段建议至少固定两档尺寸。

### 6.1 主 prototype 页面尺寸

建议：

- `80x24`
- 或 `96x28`

理由：

- 足够放下标题、capability、文本图、图像图和摘要区
- 仍然保持终端页面的实际约束

### 6.2 小尺寸专项模式

后续如果需要，可以在同一 prototype 里支持切换：

- `small plot mode`

但不建议第一阶段默认就堆第二个尺寸面板。

## 7. 与现有 e2e 的关系

### 7.1 第一阶段不建议把 image 视觉结果纳入现有文本 snapshot

原因：

- `ui/e2e/testdata` 当前基本都是文本 `.render.txt`
- image prototype 的视觉收益不能靠纯文本快照表达

### 7.2 第一阶段建议的最小自动化

prototype 相关自动化建议只做：

- demo 是否可构建
- capability 摘要文本是否可见
- text fallback 路径是否可运行

### 7.3 为什么不立即做 image e2e

因为 Phase 1 的目标是先证明：

- capability 层成立
- `App.render()` 能接图像层
- image plot 真有收益

这还不是稳定的生产级回归阶段。

## 8. 页面中的诊断内容应如何表达

### 8.1 页面内只放摘要

建议只显示：

- `mode: kitty`
- `reliable: yes/no`
- `cell: 8x16`
- `backend: image`
- `first frame: 12ms`
- `update: 4ms`

### 8.2 详细信息导出到 artifact

包括：

- 完整 capability dump
- scene summary
- 图像层元数据
- benchmark JSON

这些不应该全塞进页面。

## 9. README 应包含什么

`examples/charts_linechart_image_prototype/README.md` 第一阶段至少要包含：

- 目标
- 当前支持的终端范围
- 启动命令
- 推荐环境变量
- 当前已知限制
- diagnostics 输出说明

### 9.1 建议明确写出的限制

例如：

- 第一阶段只验证 Kitty
- 只验证单图
- 文本与图像混合帧先走同步提交
- 不支持复杂交互

## 10. 推荐测试策略

### 10.1 `main_test.go`

第一阶段建议只做：

- `Build()` smoke test
- 关键标题文本可见性检查

### 10.2 后续再考虑单独 e2e

如果 Phase 1 验证通过，Phase 2 再评估是否需要：

- `ui/e2e/charts_linechart_image_prototype_e2e_test.go`

但这不应成为第一阶段前置条件。

## 11. 与现有 `charts_linechart_demo` 的关系

建议明确分工：

### 11.1 `charts_linechart_demo`

继续负责：

- 稳定文本能力展示
- 轴标签策略展示
- 文本 snapshot 回归

### 11.2 `charts_linechart_image_prototype`

负责：

- 图像路径实验
- 文本/图像对照
- capability / diagnostics / benchmark 验证

这两者不要混成一个目录。

## 12. 一句话结论

`examples/charts_linechart_image_prototype` 的正确定位不是“又一个普通 demo”，而是：

**一个与现有文本 demo 并行存在、专门服务 image prototype 验证闭环的实验页面。**
