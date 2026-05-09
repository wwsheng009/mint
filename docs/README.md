# Mint 文档总览

本目录保存 Mint TUI Framework 的现状文档、架构说明、组件说明、调试与测试指南，以及若干历史设计记录。本文档按当前源码结构更新，优先作为进入 `docs/` 的导航页使用。

## 当前项目事实

- 模块名: `github.com/wwsheng009/mint`
- Go 版本: `go 1.24.0`，`toolchain go1.24.2`
- 应用入口: `ui.Run` / `ui.RunApp`，兼容层入口为 `app.Run`
- 主要源码目录: `ui/`、`runtime/`、`framework/`、`internal/`、`sandbox/`、`devtools/`、`examples/`
- 组件目录: `ui/components/`，当前有 60 个顶层组件目录，另有图表子组件
- 示例目录: `examples/`，当前有 66 个顶层示例目录
- E2E 测试: `ui/e2e/`，当前有 58 个 `*_e2e_test.go`
- 默认渲染路径: Fiber-first；`ui.Run` 会创建 Fiber reconciler、默认 Portal roots、Intent runtime 和 framework app
- 状态管理: 推荐使用 `runtime/store`、`runtime/reducer`、`runtime/intent`，局部组件状态仍可使用 Hooks

## 推荐阅读路径

第一次阅读当前代码库时，建议按下面顺序进入：

1. [根 README](../README.md): 项目能力、快速开始、推荐示例。
2. [架构总览](architecture/README.md): 源码分层、运行路径、渲染与事件主干。
3. [组件索引](components/README.md): 组件库清单、源码 README、示例和 E2E 入口。
4. [Store/Reducer 文档](ui/store/README.md): 应用状态、Intent、Reducer、RunApp 相关文档。
5. [调试指南](debug/README.md): 日志、渲染、事件、图形与测试相关环境变量。
6. [Sandbox 快速开始](sandbox/QUICK_START_GUIDE.md): 测试运行、事件注入、回放和快照。
7. [E2E 测试文档](testing/e2e/README.md): 交互式测试 Driver、Locator、Trace 和断言能力。

## 目录结构

### `api/`

公开 API 与规范类文档。

- [component.md](api/component.md): 组件 API 与组件规范。
- [hooks.md](api/hooks.md): Hooks API。
- [memory-safety.md](api/memory-safety.md): goroutine、订阅、定时器等内存安全规范。
- [border.md](api/border.md): 边框 API。

### `architecture/`

当前架构入口文档。

- [README.md](architecture/README.md): 源码分层、运行流程、关键包职责。
- [design/FIBER_ARCHITECTURE.md](architecture/design/FIBER_ARCHITECTURE.md): Fiber 相关设计记录。

### `components/`

面向组件使用者的索引和专题文档。

- [README.md](components/README.md): 当前组件库清单与源码入口。
- [SCROLL_VIEW_COMPONENT.md](components/SCROLL_VIEW_COMPONENT.md): ScrollView。
- [VIRTUAL_LIST_COMPONENT.md](components/VIRTUAL_LIST_COMPONENT.md): VirtualList。
- [TABS_COMPONENT.md](components/TABS_COMPONENT.md): Tabs。
- [TREEVIEW_NAVIGATION.md](components/TREEVIEW_NAVIGATION.md): TreeView 导航。
- [DATEPICKER_COMPONENT.md](components/DATEPICKER_COMPONENT.md): DatePicker。
- [TIMEPICKER_COMPONENT.md](components/TIMEPICKER_COMPONENT.md): TimePicker。
- [grid/](components/grid/): Grid 设计、验收、调试文档。
- [control/](components/control/): Pressed 状态相关修复文档。

组件源码 README 主要在 `../ui/components/<component>/README.md`，这是更接近实现的第一手文档。

### `debug/` 和 `debugging/`

调试入口与环境变量说明。

- [debug/README.md](debug/README.md): 当前调试体系总览。
- [debug/environment_variables.md](debug/environment_variables.md): 以源码为准的环境变量清单。
- [debug/quick_start.md](debug/quick_start.md): 旧版快速调试说明，部分内容偏历史。
- [debugging/DEBUG_ENVIRONMENT_VARIABLES.md](debugging/DEBUG_ENVIRONMENT_VARIABLES.md): 兼容入口，指向当前环境变量文档。

### `event/`

事件与 Pressed 状态专题。

- [PRESSED_STATE_COMPLETE_SOLUTION.md](event/PRESSED_STATE_COMPLETE_SOLUTION.md)
- [long_term_event_architecture.md](event/long_term_event_architecture.md)

当前源码主干在 `../framework/event`、`../runtime/input`、`../runtime/msg`、`../runtime/action`、`../runtime/interaction`。

### `features/`

已落地或专题化的用户功能文档。

- [mouse-text-selection.md](features/mouse-text-selection.md): 鼠标文本选择。
- [focus/](features/focus/): Tab 焦点、鼠标点击焦点等专题。

### `fiber/`

Fiber-first 相关设计和快速参考。

- [fiber_first/FIBER_PAINT_ARCHITECTURE.md](fiber/fiber_first/FIBER_PAINT_ARCHITECTURE.md)
- [fiber_first/consolidated/README.md](fiber/fiber_first/consolidated/README.md)
- [fiber_first/consolidated/FIBER_FIRST_ARCHITECTURE.md](fiber/fiber_first/consolidated/FIBER_FIRST_ARCHITECTURE.md)
- [fiber_first/consolidated/FIBER_FIRST_QUICK_REFERENCE.md](fiber/fiber_first/consolidated/FIBER_FIRST_QUICK_REFERENCE.md)

### `guide/`

开发指南与迁移指南。

- [component-development-guide.md](guide/component-development-guide.md): 组件开发指南。
- [migration-guide.md](guide/migration-guide.md): 迁移指南。
- [key-handling/](guide/key-handling/): 按键处理、调试和大小写保留说明。

### `howto/`

具体操作型指南。

- [migrate-to-targetbounds.md](howto/migrate-to-targetbounds.md)

### `inspector/`

Inspector 使用和实现说明。

- [README.md](inspector/README.md)
- [QUICK_START.md](inspector/QUICK_START.md)
- [KEYBOARD_SHORTCUTS.md](inspector/KEYBOARD_SHORTCUTS.md)
- [implementation/](inspector/implementation/)
- [features/](inspector/features/)

当前源码主干在 `../internal/inspector` 和 `../framework/inspector_integration.go`。

### `layer/`

Layer、Portal、Overlay 与 Fiber-first Layer 渲染文档。

- [LAYER_SYSTEM_ARCHITECTURE.md](layer/LAYER_SYSTEM_ARCHITECTURE.md)
- [FIBER_FIRST_LAYER_SYSTEM.md](layer/FIBER_FIRST_LAYER_SYSTEM.md)
- [PORTAL_IMPLEMENTATION.md](layer/PORTAL_IMPLEMENTATION.md)
- [IMPLEMENTATION_GUIDE.md](layer/IMPLEMENTATION_GUIDE.md)

当前源码相关入口包括 `../runtime/ui/portal.go`、`../runtime/types/layer.go`、`../internal/render/portal_layout_adapter.go`。

### `layout/`

布局系统、盒模型、边框、Modal、Portal、可视化工具和重构记录。

- [README.md](layout/README.md): 布局文档索引。
- [core_concepts/](layout/core_concepts/): Flex、Stretch、Wrap、Layer 等核心概念。
- [box_model/](layout/box_model/): 盒模型当前状态和流程图。
- [border/](layout/border/): 边框处理流程。
- [modal/](layout/modal/): Modal 居中和定位。
- [portal/](layout/portal/): Portal 设计。
- [visualizer_usage_guide.md](layout/visualizer_usage_guide.md): 布局可视化工具。

当前源码相关入口包括 `../runtime/layout`、`../ui/layout`、`../ui/layout/dsl`、`../ui/layout/visualizer`。

### `log/`

日志环境变量和 Logger 规范。

- [LOGGER_ENV_VAR_STANDARD.md](log/LOGGER_ENV_VAR_STANDARD.md)

当前源码主干在 `../internal/log`。

### `platform/`

平台输入、Ctrl 键、信号处理和 Darwin/Linux 差异说明。

- [platform.md](platform/platform.md)
- [CTRL_C_EXIT_MECHANISM.md](platform/CTRL_C_EXIT_MECHANISM.md)
- [CTRL_KEY_HANDLING.md](platform/CTRL_KEY_HANDLING.md)
- [DARWIN_INPUT.md](platform/DARWIN_INPUT.md)
- [LINUX_INPUT.md](platform/LINUX_INPUT.md)
- [SIGNAL_HANDLING.md](platform/SIGNAL_HANDLING.md)

当前源码主干在 `../runtime/platform`。

### `render/`

渲染系统入口。

- [README.md](render/README.md): 当前渲染文档索引。
- [diff/diff.md](render/diff/diff.md): diff 渲染。
- [hook/README.md](render/hook/README.md): render hook。
- [paint/optimized/](render/paint/optimized/): Fiber-first paint pipeline 与迁移说明。
- [pixel/README.md](render/pixel/README.md): 图形/像素渲染设计和 PoC 计划。

当前源码相关入口包括 `../internal/render`、`../runtime/paint`、`../runtime/render`、`../runtime/platform/*graphics*`。

### `sandbox/`

Mock、Real、Replay Sandbox 和测试辅助能力。

- [QUICK_START_GUIDE.md](sandbox/QUICK_START_GUIDE.md)
- [API_REFERENCE.md](sandbox/API_REFERENCE.md)
- [SANDBOX_ADVANCED_FEATURES.md](sandbox/SANDBOX_ADVANCED_FEATURES.md)
- [SANDBOX_DEBUG_GUIDE.md](sandbox/SANDBOX_DEBUG_GUIDE.md)
- [fiber_debug/](sandbox/fiber_debug/)

当前源码主干在 `../sandbox` 和 `../ui/test.go`。

### `state/`

运行时状态、Fiber state 和最佳实践。

- [README.md](state/README.md)
- [BEST_PRACTICES.md](state/BEST_PRACTICES.md)
- [FIBER_STATE_ARCHITECTURE.md](state/FIBER_STATE_ARCHITECTURE.md)

当前源码相关入口包括 `../runtime/state`、`../internal/state`。

### `testing/`

测试工具、E2E 和测试设计。

- [TESTING_TOOL.md](testing/TESTING_TOOL.md)
- [TESTABLE_INPUT_DESIGN.md](testing/TESTABLE_INPUT_DESIGN.md)
- [e2e/README.md](testing/e2e/README.md)
- [e2e/API_REFERENCE.md](testing/e2e/API_REFERENCE.md)

当前源码主干在 `../ui/e2e`、`../ui/test.go`、`../sandbox/testing`。

### `theme/`

主题系统设计、Ant Design 风格映射、spacing/background 和渲染流程。

- [theme_system_guide.md](theme/theme_system_guide.md)
- [theme_rendering_flow.md](theme/theme_rendering_flow.md)
- [spacing_and_background_design.md](theme/spacing_and_background_design.md)
- [design/](theme/design/)

当前源码主干在 `../framework/theme`。

### `ui/`

UI 运行时专题文档。

- [intent_bubble/](ui/intent_bubble/): Intent bubble、Context、Instance tree 等设计与实现。
- [store/README.md](ui/store/README.md): Store/Reducer/Intent 文档入口。
- [optiongroup/](ui/optiongroup/): OptionGroup 专题。

### `ai/`

AI / MCP 集成设计。

- [design/framework_app_ai_mcp_design.md](ai/design/framework_app_ai_mcp_design.md)

当前源码相关入口包括 `../internal/ai`、`../runtime/ai`、`../framework/ai_*.go`、`../examples/ai_mcp_demo`。

### `superpowers/`

日期命名的计划与规格文档。

- [plans/](superpowers/plans/)
- [specs/](superpowers/specs/)

## 源码入口速查

| 关注点 | 源码入口 |
|---|---|
| 应用启动 | `../ui/app.go`, `../app/app.go`, `../framework/app.go` |
| 声明式 VNode / Hooks | `../ui`, `../runtime/ui` |
| 组件库 | `../ui/components` |
| Fiber reconciler | `../internal/reconciler`, `../runtime/ui/fiber*.go` |
| 渲染 pipeline | `../internal/render`, `../runtime/paint`, `../runtime/render` |
| 布局 | `../runtime/layout`, `../ui/layout` |
| Intent / Store / Reducer | `../runtime/intent`, `../runtime/store`, `../runtime/reducer`, `../runtime/statemachine` |
| 事件与输入 | `../framework/event`, `../runtime/input`, `../runtime/msg`, `../runtime/action` |
| Focus / Selection | `../runtime/focus`, `../runtime/selection`, `../runtime/input` |
| Layer / Portal | `../runtime/ui/portal.go`, `../runtime/types/layer.go`, `../internal/render/portal_layout_adapter.go` |
| Sandbox / Test | `../sandbox`, `../ui/test.go`, `../ui/e2e` |
| DevTools / Inspector | `../devtools`, `../internal/inspector` |
| AI / MCP | `../internal/ai`, `../runtime/ai`, `../framework/ai_*.go` |

## 文档维护约定

- 面向当前使用者的文档优先放在 `guide/`、`api/`、`components/`、`debug/`、`testing/`、`sandbox/`。
- 架构和设计说明放在对应子系统目录，例如 `architecture/`、`render/`、`layout/`、`layer/`。
- 历史记录和已归档内容优先保留在 `../docsArchive`，从 `docs/` 链接时应显式标注“已归档”。
- 新增入口文档时同步更新本文件，避免出现已经不存在的目录或文件链接。
