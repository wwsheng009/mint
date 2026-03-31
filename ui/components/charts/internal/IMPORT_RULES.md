# Charts Internal Import Rules

> 这份文档把 `ui/components/charts/internal/` 的依赖规则写成显式约束，避免后续演化时重新引入父子包耦合。

## 允许的依赖方向

- `ui` 顶层可以 import 具体图表组件
- 具体图表组件可以 import `charts/internal/*`
- `charts/internal/*` 可以 import `runtime/*`
- `charts/internal/*` 可以 import `framework/theme`
- `charts/internal/*` 在确有必要时可以 import `ui/components/internal/*`

## 禁止的依赖方向

- `charts/internal/*` 不得 import 任意具体图表组件
- 图表组件不得 import `github.com/wwsheng009/mint/ui`
- 图表组件之间不得互相 import 来复用局部 helper
- `charts/internal/*` 不应 import `charts/model` 之外的上层聚合包

## 自动化检查

当前已有根层测试 [`../import_rules_test.go`](../import_rules_test.go) 覆盖以下硬约束：

- `charts` 根目录不得出现可编译 Go 父包文件
- 具体图表组件不得 import `github.com/wwsheng009/mint/ui`
- `charts/internal/*` 不得 import 任意具体图表组件
- 具体图表组件之间不得互相 import
- `charts/internal/*` 只能依赖白名单层级：`runtime/*`、`framework/theme`、`ui/components/internal/*`、`charts/internal/*`、`charts/model`

这份文档里剩余的“尽量不要 / 不应”类约束，当前仍以代码评审和后续增量检查为主。

## 共享逻辑下沉规则

- 只被单个图表使用的 helper，留在该组件目录
- 两个及以上图表复用、且属于实现细节的能力，下沉到 `charts/internal/*`
- 两个及以上图表复用、且必须对外公开的数据契约，才考虑进入 `charts/model`

## 当前建议的职责边界

- `canvas/` 负责字符画布与 `paint.Buffer` 互转
- `raster/` 负责线、点、柱等图元离散化
- `scale/` 负责数值域到字符坐标的映射
- `axis/` 负责轴线、网格线、标签行生成
- `layout/` 负责 header / plot / footer 的测量与排布
- `palette/` 负责主题语义色映射与颜色降级
- `downsample/` 负责宽度受限时的数据压缩

## 触发重构的信号

出现以下情况时，应优先重构依赖而不是继续复制代码：

- 两个图表开始复制同一份 clipping 或 label-fitting 逻辑
- 某个 `internal` 包开始想 import 某个具体图表类型
- `ui/shortcuts_charts.go` 之外的上层包开始直接依赖图表内部实现
