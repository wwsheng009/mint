# Mint

Mint 是一个现代化、声明式、可测试的 Go TUI 框架。项目当前已经不再是只有基础控件的原型仓库，而是具备完整组件库、Store/Reducer/Intent 状态管理、Fiber-first 渲染与布局路径、Layer/Portal/Focus 运行时、图表组件、Sandbox/E2E 测试设施，以及 DevTools、Inspector、AI MCP 集成能力的终端 UI 开发栈。

## 当前实施状态

截至 `2026-04-23`，Mint 已从“框架原型期”进入“能力收口与持续增强期”。核心运行时、主流组件、示例体系和测试基础设施已经可用，当前工作重点不再是补齐最基础的组件，而是继续做复杂组件增强、文档同步和测试密度提升。

| 指标 | 当前情况 |
|------|----------|
| Go 版本 | `go 1.24.0`，`toolchain go1.24.2` |
| 组件/支撑目录 | `60` 个（排除 `ui/components/docs` 与 `ui/components/internal`） |
| 顶层示例目录 | `66` 个 |
| E2E 套件 | `58` 个 `ui/e2e/*_e2e_test.go` |
| 组件路线图 | `ui/components/ROADMAP.md` 中 Phase 1-4 已完成收口 |
| 最近全量验证 | `go test ./... -count=1` 于 `2026-04-23` 通过 |

当前剩余工作主要集中在：

- 已有组件的能力补齐和语义收口
- 复杂组件 README、迁移指南与示例说明同步
- `internal/` 支撑模块与高复杂交互场景的测试覆盖继续加密

## 项目功能

### 1. 声明式 UI 与应用模型

- 使用 `ui.Run(...)` 启动应用，根视图返回 `ui.VNode`
- 组件以 Builder 风格为主，适合链式声明和静态组合
- 提供 `VStack`、`HStack`、`Flex`、`Grid`、`Row/Col`、`Space`、`Layout` 等布局原语
- 支持主题、样式、边框、宽高、对齐、滚动和可聚焦节点组合

### 2. 状态管理

- 组件内局部状态仍可使用 Hooks，例如 `UseEffect`、`UseRef` 等
- 应用级状态推荐使用 `store.NewStore(...) + reducer.NewBuilder(...) + intent.Intent`
- 通过 `ui.UseStoreSelector(...)` 从 Store 中订阅精确状态切片
- 表单与字段交互可结合 `intent.BindField(...)`、`FieldChangeIntent` 等机制统一处理
- 当前交互类组件示例以 `.OnPress(...)` 触发 intent 为主

### 3. 运行时与交互基础设施

- 已具备 Fiber-first 渲染路径和可持续演进的运行时结构
- 已实现 Focus 管理、Tab/方向键切换、鼠标交互、滚动容器和可测试输入注入
- 已实现 Layer、Portal、Overlay、Popup、Modal、Drawer 等浮层能力
- 菜单、Tooltip、Popover、Popconfirm 等组件已接入 viewport-aware placement、fallback 和 clamp 逻辑

### 4. 组件库

Mint 已具备完整的终端组件库能力，覆盖常见应用场景：

- 表单输入：`Input`、`Textarea`、`Select`、`Checkbox`、`Radio`、`Switch`、`Slider`、`Rate`、`DatePicker`、`TimePicker`、`Cascader`、`Transfer`、`Form`
- 数据展示：`Table`、`List`、`VirtualList`、`TreeView`、`Descriptions`、`Statistic`、`Timeline`、`Badge`、`Tag`
- 反馈与状态：`Alert`、`Spin`、`Notification`、`Toast`、`Result`、`Skeleton`、`Progress`
- 导航与容器：`Tabs`、`Menu`、`Breadcrumb`、`Pagination`、`Steps`、`Anchor`、`Panel`
- 浮层与弹出：`Modal`、`Drawer`、`Tooltip`、`Popover`、`Popconfirm`
- 布局与基础：`Text`、`Divider`、`ScrollView`、`Absolute`、`Wrap`、`Grid`、`Layout`、`Row/Col`

完整清单与状态请查看 [ui/components/ROADMAP.md](./ui/components/ROADMAP.md)。

### 5. 图表与数据可视化

图表能力已经进入可用阶段，并配有示例和 E2E 验证：

- `sparkline`
- `bulletchart`
- `barchart`
- `linechart`
- `heatmap`
- `scatterplot`
- `candlestick`

图表目录说明见 [ui/components/charts/README.md](./ui/components/charts/README.md)，综合示例见 [examples/charts_gallery_demo](./examples/charts_gallery_demo)。

### 6. 测试、调试与可观测性

- 提供 `sandbox`、record/replay、snapshot、mock/real sandbox 等测试支撑
- 已有大批 `ui/e2e` 用例覆盖组件回归、布局、浮层和图表渲染
- 提供 DevTools、Timeline、Time Travel、Replay、Observation、Remote 调试能力
- 提供 Inspector、布局诊断、节点检查和运行时分析辅助

### 7. AI / MCP 集成

- 仓库内已具备嵌入式 MCP Server 能力
- 支持 `http` / `pipe` 等传输方式和可选鉴权
- 可在示例中演示 UI 与 AI/MCP 工具暴露的集成方式

相关示例见 [examples/ai_mcp_demo](./examples/ai_mcp_demo)。

## 推荐使用方式

当前更推荐把 Mint 当作“声明式视图 + 类型化状态流 + 可测试运行时”来使用，而不是只把它当作一组零散组件。

- 组件内短生命周期状态：使用 Hooks，适合局部副作用、定时器、引用和临时 UI 状态
- 应用级业务状态：使用 `Store + Reducer + Intent`，保持单一状态源和可预测状态变更
- 交互分发：按钮、菜单、快捷操作等通过 `.OnPress(...)` 发送 intent
- 复杂表单：通过 Field 绑定、验证、受控/非受控模式组合实现
- 复杂弹层：优先使用已接入 Layer/Portal 的组件能力，不建议自行拼接底层 overlay 逻辑

如果你是第一次接触当前架构，建议先从以下示例进入：

- [examples/counter](./examples/counter)
- [examples/store_reducer_demo](./examples/store_reducer_demo)
- [examples/timer](./examples/timer)
- [examples/menu_demo](./examples/menu_demo)
- [examples/ui_demos/demo1_full_featured](./examples/ui_demos/demo1_full_featured)

## 快速开始

### 环境要求

- Go `1.24+`

### 安装

```bash
git clone git@github.com:wwsheng009/mint.git
cd mint
go mod download
```

### 运行示例

```bash
go run ./examples/counter
go run ./examples/store_reducer_demo
go run ./examples/timer
go run ./examples/menu_demo
go run ./examples/charts_gallery_demo
go run ./examples/ai_mcp_demo
```

### 最小示例

下面这个示例展示了当前推荐的 `Store + Reducer + Intent + OnPress` 用法：

```go
package main

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/ui"
)

type AppState struct {
	Count int
}

type IncrementIntent struct{}

func (IncrementIntent) IntentType() string { return "CounterIncrement" }
func (IncrementIntent) StayPressed() bool  { return true }

var appStore = store.NewStore(AppState{})

func init() {
	reducer.NewBuilder[AppState]().
		On(IncrementIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Count++
			return s
		}).
		BuildAndRegister(intent.DefaultRegistry(), appStore)
}

func App() ui.VNode {
	count := ui.UseStoreSelector(appStore, func(s AppState) int { return s.Count })

	return ui.VStack(
		ui.NewTextBuilder("Mint Counter").Bold(true).FgColor("cyan").Build(),
		ui.NewTextBuilder(fmt.Sprintf("Count: %d", count)).FgColor("green").Build(),
		ui.NewButtonBuilder(" +1 ").
			Variant(ui.ButtonVariantPrimary).
			OnPress(IncrementIntent{}).
			Build(),
	)
}

func main() {
	if err := ui.Run(App,
		ui.WithWidth(40),
		ui.WithHeight(10),
		ui.WithTitle("Mint Counter"),
	); err != nil {
		panic(err)
	}
}
```

## 示例与文档入口

### 示例入口

- 基础状态与组件：[`examples/counter`](./examples/counter)、[`examples/timer`](./examples/timer)
- Store/Reducer 架构：[`examples/store_reducer_demo`](./examples/store_reducer_demo)、[`examples/store_mixed_demo`](./examples/store_mixed_demo)
- 菜单与浮层：[`examples/menu_demo`](./examples/menu_demo)
- 完整 UI 演示：[`examples/ui_demos`](./examples/ui_demos)、[`examples/fiber_demos`](./examples/fiber_demos)
- 图表：[`examples/charts_gallery_demo`](./examples/charts_gallery_demo)
- DevTools：[`examples/devtools_demo`](./examples/devtools_demo)
- Sandbox：[`examples/sandbox`](./examples/sandbox)
- AI / MCP：[`examples/ai_mcp_demo`](./examples/ai_mcp_demo)

### 文档入口

- 文档总览：[docs/README.md](./docs/README.md)
- 组件路线图：[ui/components/ROADMAP.md](./ui/components/ROADMAP.md)
- Store/Reducer 指南：[docs/ui/store/guides/README.md](./docs/ui/store/guides/README.md)
- Sandbox 文档：[docs/sandbox/QUICK_START_GUIDE.md](./docs/sandbox/QUICK_START_GUIDE.md)
- Inspector 文档：[docs/inspector/README.md](./docs/inspector/README.md)

## 测试与质量

当前仓库已经具备单元测试、E2E、Sandbox、Snapshot 和示例验证的组合能力。

```bash
go test ./... -count=1
```

你也可以按模块运行：

```bash
go test ./ui/components/... -count=1
go test ./ui/e2e/... -count=1
go test ./sandbox/... -count=1
```

## 后续重点

项目当前不是“缺少基础框架”，而是继续推进以下收口工作：

- 对较复杂组件继续补齐交互细节、受控模式和文档说明
- 对历史设计文档、迁移指南、示例代码做现状同步
- 对 `internal/` 支撑层、特殊边界条件和高复杂交互场景继续补测试

## 许可证

MIT License
