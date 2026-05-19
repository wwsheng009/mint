# Prototype Benchmark 与 Diagnostics 产物布局规格

## 1. 文档目的

`GROUP_E_PROTOTYPE_BENCHMARK_AND_DIAGNOSTICS_SPEC.md` 已经说明：

- prototype 要有 benchmark
- 要有 diagnostics
- 需要 go / no-go 判断闭环

但还缺一件更具体的事：

> benchmark 结果、capability dump、scene 摘要、图像导出文件，第一阶段应该放在哪里、叫什么、哪些进仓库、哪些绝不进仓库。

本文件就是这份产物布局规格。

## 2. 当前仓库已有的产物模式

对照当前仓库：

- `ui/e2e/testdata/*.render.txt`
- `examples/*`

可以看出当前稳定模式是：

- 文本 e2e snapshot 进入 `ui/e2e/testdata`
- 业务示例与 demo 进入 `examples/`

这意味着：

- image prototype 的实验性产物不能直接混进当前文本 snapshot 目录
- 否则会把稳定回归资产和实验性诊断资产混在一起

## 3. 第一阶段产物分类

建议把 prototype 产物明确分成 3 类。

### 3.1 进仓库的稳定文档型产物

包括：

- prototype `README.md`
- benchmark 字段说明
- diagnostics 字段说明

这些是稳定资产，可以进仓库。

### 3.2 运行期生成、默认不进仓库的实验产物

包括：

- capability dump
- benchmark JSON
- scene summary
- bitmap 导出
- presenter 输出日志

这些应该默认落到工作目录的临时 artifact 路径，不进入 git。

### 3.3 未来可能进入仓库的少量 golden

例如后续如果确定了：

- 某些 capability dump baseline
- 某些 image metadata golden

再单独评估是否入库。Phase 1 不做这个决定。

## 4. 推荐产物目录

### 4.1 第一阶段推荐根目录

建议使用：

- `artifacts/pixel/`

作为本地运行时产物根目录。

原因：

- 语义清晰
- 不与 `ui/e2e/testdata` 冲突
- 便于统一忽略和清理

### 4.2 推荐子目录

建议按 prototype 名称和时间戳拆分：

```text
artifacts/pixel/
└── linechart-prototype/
    └── 20260401-153000/
        ├── capability.json
        ├── benchmark.json
        ├── scene-summary.txt
        ├── image-layer-0.png
        ├── image-layer-0.meta.json
        └── presenter.log
```

这样做的好处是：

- 一次运行一组目录
- 不会互相覆盖
- 便于失败时整体打包

## 5. 每类产物的建议格式

### 5.1 `capability.json`

建议字段至少包括：

- `mode`
- `reliable`
- `cellPixelWidth`
- `cellPixelHeight`
- `probeSource`
- `notes`

### 5.2 `benchmark.json`

建议字段至少包括：

- `firstFrameMs`
- `updateFrameMs`
- `resizeFrameMs`
- `textOutputBytes`
- `imagePayloadBytes`
- `cacheHit`
- `cacheMiss`

### 5.3 `scene-summary.txt`

建议文本化输出：

- 文本层尺寸
- image layer 数量
- 每个 image layer 的 `cell rect`
- 每个 image layer 的 `pixel size`
- 是否启用 fallback

### 5.4 `image-layer-*.png`

如果第一阶段允许导出位图，建议每层单独导出。

原因：

- 更容易分辨是 raster 问题还是合成问题

### 5.5 `image-layer-*.meta.json`

建议记录：

- `id`
- `hash`
- `cellRect`
- `pixelWidth`
- `pixelHeight`
- `altText`

### 5.6 `presenter.log`

建议记录：

- `Present`
- `Replace`
- `Delete`
- `Clear`

这些调用的最小时间线。

## 6. 什么不能进入 `ui/e2e/testdata`

第一阶段明确不要放进：

- bitmap 文件
- capability dump
- benchmark JSON
- presenter 原始日志

原因：

- `ui/e2e/testdata` 当前是稳定文本快照资产
- 一旦混入实验性二进制或运行期诊断文件，会让目录语义失真

## 7. 建议的环境变量

为了让 prototype 更容易落地，建议支持：

- `MINT_PIXEL_ARTIFACT_DIR`

如果未设置，则默认：

- `artifacts/pixel/linechart-prototype/<timestamp>/`

如果设置，则优先使用显式目录。

## 8. 输出时机建议

### 8.1 成功运行时

建议至少输出：

- `capability.json`
- `benchmark.json`
- `scene-summary.txt`

### 8.2 失败时

建议额外输出：

- 已生成的 image layer bitmap
- presenter.log
- 错误摘要

这样失败现场更完整。

## 9. 与 prototype 页面内摘要的关系

页面中只显示简要信息，例如：

- `mode: kitty`
- `first frame: 12ms`
- `payload: 32KB`

完整信息进入 artifact 目录。

这条边界必须清楚，否则 prototype 页面会过载。

## 10. 是否需要 `.gitignore`

如果后续进入代码实现阶段，建议补：

- `artifacts/pixel/`

到合适的忽略规则中。

但在当前文档阶段，只需要明确它属于：

- 运行期生成目录
- 默认不入库

## 11. Phase 1 的最小要求

第一阶段不要求所有产物都齐全。

建议最低要求：

- `capability.json`
- `benchmark.json`
- `scene-summary.txt`

如果能额外导出 `image-layer-0.png`，更好，但不是必须前置条件。

## 12. 一句话结论

第一阶段 prototype 的 benchmark 与 diagnostics 产物应当：

**稳定描述、统一落位、默认不入库，并与现有文本 e2e snapshot 目录严格分离。**
