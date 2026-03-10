# framework/app AI + MCP 集成设计

**文档状态**: Draft  
**适用范围**: `framework.App` / `ui.Run()` / `ui.RunApp()` 交互式应用  
**参考文档**: `framework/docs/AI_INTEGRATION.md`  
**代码基线**: 基于 2026-03-09 仓库现状审计

## 1. 背景

`framework/docs/AI_INTEGRATION.md` 已经定义了 AI 作为一级操作者的目标形态：

1. AI 不依赖截图/OCR。
2. AI 通过结构化状态理解 UI。
3. AI 通过语义化指令驱动 UI，而不是直接篡改内部渲染对象。

当前仓库已经具备大量支撑能力：

1. `framework.App` 已经有完整交互生命周期、事件泵、Action 路由、Focus、HitMap、Fiber-first 渲染链路。
2. `internal/render.DeclarativeNode` 已经能暴露 `VNode`、`Fiber`、`LayoutBox`、`PaintableBox`、`HitMap`、`ComponentInstance`。
3. `runtime/ai`、`runtime/state` 已经存在控制器和快照模型雏形。

但这些能力还没有被组织成一个可供外部 AI Agent 使用的统一宿主，更没有通过 MCP 暴露给外部工具链。

本设计的目标，是在不破坏当前交互式渲染主路径的前提下，为 `framework/app` 应用提供：

1. 启动后自动可用的本地 MCP Server。
2. 面向 AI 的结构化 Inspect 能力。
3. 面向 AI 的安全写操作能力。
4. 与当前 `Fiber -> LayoutBox -> PaintableBox -> HitMap -> ActionBridge` 实现对齐的最小侵入改造。

## 2. 目标与非目标

### 2.1 目标

1. 在交互式应用启动后自动启用 MCP 服务。
2. 通过 MCP 获取以下元信息：
   - 渲染后的 `VNode` 树
   - `Fiber` 树
   - `LayoutBox` 树
   - `PaintableBox` 树
   - `HitMap`
   - `Focus` / `Selection` / `ComponentInstance` 能力摘要
3. 支持通过指令操作应用：
   - 点击、输入、导航、聚焦
   - 获取组件值/状态
   - 设置组件值
   - 选中某个选项/条目
   - 读取表单数据
4. 保持 AI 侧“语义优先、内部结构次之”的使用原则。
5. 兼容 `framework.App` 直接使用方式与 `ui.Run()/ui.RunApp()` 高层入口。

### 2.2 非目标

1. 不让 AI 直接修改 `VNode`、`Fiber`、`LayoutBox`、`PaintableBox` 本身。
2. 不把 Inspector UI 直接变成 MCP Server；两者共享底层模型，但不共享 UI 逻辑。
3. 第一阶段不解决跨进程远程协作、安全审计、权限系统的全部问题。
4. 第一阶段不保证所有组件都支持“直接数据写入”，只保证一组核心组件和统一 Action 路径。

## 3. 当前实施状态审计

### 3.1 已有能力

### A. `framework.App` 已具备 AI 宿主所需的核心运行时骨架

当前 `framework.App` 已经拥有：

1. 生命周期：`SetRoot()`、`Init()`、`Run()`、`Close()`。
2. 事件系统：`Pump`、`InputProcessor`、`InputTracker`、`InteractionContext`。
3. 行为分发：`action.Router`、`actionbridge.Bridge`、`ScopeDispatcher`。
4. 焦点系统：`rtui.FiberFocusManager`。
5. 渲染后命中：`HitMap`，并在每帧渲染后同步回 `Pump`。
6. 调试入口：Inspector、调试记录器、直接事件注入能力。

这意味着 AI/MCP 不需要自建输入系统、命中系统、焦点系统和组件调用链。

### B. `DeclarativeNode` 已能导出关键元数据

当前 `internal/render.DeclarativeNode` 已有以下读侧能力：

1. `GetRenderedRoot()`：读取渲染后的 `VNode` 根。
2. `GetFiberRoot()`：读取 `Fiber` 根。
3. `GetLayoutRoot()` / `GetLayoutBoxes()`：读取布局树。
4. `GetPaintableRoot()` / `GetPaintableBoxes()`：读取绘制树。
5. `GetHitMap()`：读取渲染后命中表。
6. `GetComponentInstances()`：读取 `NodeID -> ComponentInstance` 映射。
7. `GetPaintDirtyRects()`：读取脏矩形提示。
8. `GetPortalRoots()`：读取 portal 布局根。

这些接口已经覆盖了用户提出的绝大多数“元信息”读取需求。

### C. 跨层 ID 基本已经打通

当前链路中已经存在稳定映射：

1. `Fiber.NodeID` 是运行时稳定标识。
2. `Fiber.ActionTargetID = fmt.Sprintf("%d", Fiber.NodeID)`。
3. `LayoutBox.ID` 使用字符串化的 `NodeID`。
4. `PaintableBox.NodeID` 也保留 `NodeID`。
5. `HitMap` 中已经能反查 `TargetFiber`。
6. `VNode.Key` 在 reconciler 路径下已大量承载 `/root/...` 路径信息，可作为调试路径。

这意味着我们可以用 `NodeID + Path + ComponentID` 作为跨层索引主键。

### D. `runtime/ai` 和 `runtime/state` 已有雏形

当前仓库已存在：

1. `runtime/ai.Controller`
2. `runtime/ai.RuntimeController`
3. `runtime/state.Snapshot`
4. `runtime/state.Tracker`

这说明语义层协议已经有基础模型，不需要完全重写。

### E. 组件实例已经存在一批可复用的读写能力

当前很多组件实例已经暴露了明确方法，例如：

1. `input` / `textarea`：`GetValue()`、`SetValue()`。
2. `form`：`GetValue(field)`、`SetValue(field, value)`、`GetValues()`。
3. `button` / `checkbox` / `input` / `select` / `textarea` / `optiongroup`：大量实现了 `GetProp()` / `SetProp()`。
4. 交互组件普遍实现 `HandleAction()`、`SetFocus()`、`HasFocus()`、`IsDisabled()`。
5. `optiongroup` 已有 `SelectOption()` / `DeselectOption()`。

这为“AI 直接设值/选中/查询值”提供了现实基础。

### 3.2 当前缺口

### A. `framework.App` 和 `runtime/ai.RuntimeController` 还没有接上

目前两边的宿主模型并不一致：

1. `runtime/ai.RuntimeController` 依赖 `action.Dispatcher`。
2. `framework.App` 当前主要走的是 `action.Router + actionbridge.Bridge`。
3. `runtime/ai.RuntimeController` 使用 `focus.Manager`。
4. `framework.App` 实际使用的是 `rtui.FiberFocusManager`。
5. `framework.App` 没有挂接 `state.Tracker`。

结论：当前 `runtime/ai.RuntimeController` 不能直接复用到 `framework.App`。

### B. 没有 App 级别的结构化快照构建器

虽然原始树可读，但还缺：

1. `framework.App` 每帧可复用的 `state.Snapshot`。
2. 统一的 `ComponentID / NodeID / Path` 索引。
3. 针对 MCP 的 JSON-safe 序列化。
4. `Find` / `Query` / `WaitUntil` 的 App 级实现。

### C. 缺少面向外部协程的安全调用入口

MCP Server 通常运行在单独 goroutine 中。当前 `framework.App` 主循环不是为并发外部调用设计的：

1. `render()`、`processMsg()`、`hitMap` 更新、`root` 读取都主要依赖主线程时序。
2. 直接从 MCP 请求 goroutine 遍历 `Fiber`、`LayoutBox`、`PaintableBox` 存在竞态风险。
3. 直接从外部 goroutine 调用组件实例写接口也不安全。

结论：必须增加“主循环串行执行”的控制队列或调用桥。

### D. 直接写操作能力目前是碎片化的

当前并不存在统一的“组件可读写能力协议”，而是组件各自暴露方法：

1. `SetValue(string)`
2. `SetProp(key, value)`
3. `SelectOption(string)`
4. `HandleAction(*action.Action)`

结论：需要一个能力注册表或适配层，屏蔽组件差异。

### E. `Fiber.Props` 不能作为 AI 读模型的权威来源

当前实现已经明确证明：

1. `Fiber.Props` 是 Fiber 创建时的快照。
2. 它可能不是最新值。
3. 对 AI 来说，如果直接依赖 `Fiber.Props`，会导致 Inspect 和 selector 匹配错误。

结论：元信息构建必须优先读取 `Instance.GetProps()`，其次才是 Fiber 快照。

### F. 交互式 App 嵌入式 MCP 不能默认走 stdio

这是本设计最关键的协议约束之一：

1. 交互式 TUI 已经占用标准输入读取按键/鼠标。
2. 交互式 TUI 也已经占用标准输出绘制终端。
3. 如果把嵌入式 MCP 也绑定到 stdio，会与现有交互路径直接冲突。

结论：嵌入式交互式 App 的 MCP 默认传输必须是本地网络/本地管道，而不是 stdio。

### G. `runtime/ai` 是早期原型目录，不应继续作为主实现目录

当前 `runtime/ai` 的问题不是“少功能”，而是“宿主假设已经过期”：

1. 它依赖的是 `runtime/core.Runtime + action.Dispatcher + focus.Manager + state.Tracker` 这套宿主模型。
2. 当前真实交互式应用主路径已经是 `framework.App + action.Router + actionbridge.Bridge + rtui.FiberFocusManager + DeclarativeNode`。
3. `runtime/ai.RuntimeController` 与当前 `framework.App` 的 Action/Focus/Render 组织方式并不对齐。
4. 继续在 `runtime/ai` 上叠加 App 级能力，只会把最新实现反向塞进旧抽象，后续更难维护。

结论：

1. `runtime/ai` 不应继续承担 App 级 AI/MCP 集成主实现。
2. 它应该被收缩为过渡兼容层，或仅保留少量纯模型/说明文件。
3. 新方案应围绕 `framework.App`、`internal/render.DeclarativeNode` 与 `runtime/ui` 当前结构重建。

## 4. 设计原则

1. **AI 是一级操作者**  
   AI 优先走语义化 Action 和安全能力接口，不走坐标模拟和 OCR。

2. **读写分离**  
   `VNode/Fiber/LayoutBox/PaintableBox` 只读；写操作通过 Action 或 Capability Adapter。

3. **主循环串行化**  
   所有直接读取运行时树、直接操作组件实例的行为，都必须在 App 主循环上下文中执行。

4. **跨层稳定标识**  
   `NodeID` 作为跨 `Fiber/Layout/Paintable/HitMap` 主键；`Path` 作为调试路径；`ComponentID` 作为业务定位键。

5. **语义快照优先，调试树按需获取**  
   高频场景使用 `state.Snapshot`；大体量树信息按需抓取，避免每帧巨量序列化。

6. **现有能力优先复用**  
   优先复用 `DeclarativeNode`、`ActionBridge`、`HitMap`、`Instance`、`Inspector` 中已实现的能力，不重复造轮子。

7. **对外默认本地、默认最小暴露**  
   默认仅监听本机地址，并支持只读/读写模式切换。

8. **依赖方向必须单向流动**  
   `runtime/*` 不反向依赖 `framework` 或 `internal/*`；AI 协议层不直接接触宿主实现指针；宿主层只做装配。

9. **原型目录先收缩再复用**  
   `runtime/ai` 先做职责拆分和兼容层处理，再决定是否保留少量纯类型；不能继续作为主实现目录扩展。

## 5. 总体架构

```text
MCP Client
    |
    v
[ MCP Transport ]
    |
    v
[ internal/ai/mcp.Server ]
    |
    v
[ internal/ai/service.Service ]
    |                         \
    | read                     \ write
    v                           v
[ Snapshot Cache ]         [ App Invoke Queue ]
    |                           |
    v                           v
[ framework.App AI facade ] -> [ framework.App main loop ]
                                  |
                                  v
                 [ ActionBridge / Focus / DeclarativeNode ]
                                  |
                                  v
               [ VNode / Fiber / LayoutBox / PaintableBox / HitMap ]
```

核心思想：

1. `framework.App` 是唯一真实宿主。
2. `framework` 只暴露宿主侧 API，真实实现放在 `internal/ai/*`。
3. 读操作分成两类：
   - 轻量读：走缓存快照。
   - 重量级树读取：通过主循环调用按需构建。
4. 写操作统一通过主循环调用执行，确保线程安全。

## 6. 模块设计

### 6.1 目录结构与包边界

不建议把新实现继续堆在 `runtime/ai`，也不建议新建一个需要反向依赖 `framework.App` 的 `framework/ai` 子包。

推荐目录结构：

```text
framework/
  ai_config.go            # 公开配置类型：AIConfig / MCPConfig
  ai_enable.go            # App.EnableAI / 生命周期挂接
  ai_invoke.go            # App.Invoke / 主循环串行调用桥
  ai_api.go               # App 对外暴露的 AI facade / helper

ui/
  ai_options.go           # WithAI / WithMCP / ui.Run 集成

internal/ai/
  service/
    service.go            # AI 子系统总装配，生命周期管理
    host.go               # Host 接口定义，只暴露 App 所需窄接口
  snapshot/
    builder.go            # 轻量语义快照构建
    tree_builder.go       # VNode/Fiber/Layout/Paintable 树快照
    cache.go              # 最新帧缓存与 watcher
  selector/
    engine.go             # selector 解析与匹配
  capability/
    registry.go           # 组件能力识别
    input.go              # input/textarea adapter
    form.go               # form adapter
    select.go             # select/optiongroup adapter
  serialize/
    sanitize.go           # JSON-safe 序列化与裁剪
  mcp/
    server.go             # MCP server 生命周期
    tools.go              # tools 实现
    resources.go          # resources 实现
    transport_http.go     # 本地 HTTP 传输
    auth.go               # token / read-only

internal/inspectmeta/
  path.go                 # /root/... path 提取
  vnode.go                # VNode 元信息提取
  fiber.go                # Fiber 元信息提取
  tree.go                 # 通用树快照 helper

runtime/ai/
  doc.go                  # 标记为 early prototype / deprecated
  compatibility.go        # 可选：保留向后兼容别名或过渡说明
```

其中：

1. `framework` 只放宿主公开 API 和装配代码。
2. 真实实现放在 `internal/ai/*`，避免新的公共包反向引用 `framework`。
3. `internal/inspector` 中纯元信息提取逻辑应下沉到 `internal/inspectmeta`，供 Inspector 和 AI 共同复用。
4. `runtime/ai` 只保留兼容层或迁移说明，不再承载主实现。

推荐依赖方向：

```text
ui
  -> framework
      -> internal/ai/service
          -> internal/ai/{snapshot,selector,capability,mcp,serialize}
          -> internal/inspectmeta
          -> internal/render
          -> runtime/*

internal/render
  -> framework/component
  -> runtime/*

runtime/*
  -> runtime/* only
```

硬性约束：

1. `runtime/*` 禁止导入 `framework` 或 `internal/*`。
2. `internal/ai/*` 禁止导入 `framework` 包本身。
3. `internal/render` 禁止导入 `internal/ai/mcp`，避免渲染层与协议层耦合。
4. `framework` 不应依赖一个再反向依赖 `framework` 的子包，例如 `framework -> framework/ai -> framework` 这种结构必须避免。

旧 `runtime/ai` 文件建议迁移方式：

1. `runtime/ai/controller.go`
   - 不再作为主控制器接口目录继续扩展。
   - 若其中的语义接口仍有价值，迁移为 `framework` 宿主公开 facade 的方法集合，或拆成 `internal/ai/service` 内部接口。

2. `runtime/ai/runtime_controller.go`
   - 直接视为过期实现。
   - 不做“原样迁移”。
   - App 级控制器应围绕 `framework.App` 重新实现。

3. `runtime/ai/operations.go`
   - 暂不作为第一阶段必须保留。
   - 若后续仍需保留“操作序列”概念，应在新的 AppController 稳定后，重建为宿主侧 helper，而不是直接搬迁旧实现。

4. `runtime/ai/README.md`
   - 保留为迁移说明。
   - 明确标注该目录是 early prototype，与当前交互式宿主链路不一致。

### 6.2 `framework.App` 宿主公开面

`framework` 包本身应只暴露宿主侧 API，不承载复杂实现细节。

建议保留在 `framework` 包内的内容：

1. `AIConfig`
2. `MCPConfig`
3. `App.EnableAI()`
4. `App.Invoke()`
5. 必要的对外只读 helper，如 `App.AIStatus()`、`App.AIEndpoint()`

原因：

1. `framework.App` 是当前交互式宿主。
2. 若把这些公开能力放到 `framework/ai` 子包，极易形成 `framework` 与 `framework/ai` 之间的反向依赖。

### 6.3 `internal/ai/service`

职责：

1. 挂载到 `framework.App`。
2. 管理 MCP Server 生命周期。
3. 维护最新轻量快照缓存。
4. 协调 selector / capability / snapshot / mcp 等子模块。
5. 在 `App.Init()` / `render()` / `App.Close()` 中接收生命周期回调。

建议结构：

```go
package service

type Host interface {
    Invoke(ctx context.Context, fn func() (any, error)) (any, error)
    MarkDirty()
    IsRunning() bool
}

type Service struct {
    host       Host
    cfg        Config
    controller Controller
    mcp        *mcp.Server
    latest     atomic.Pointer[FrameSnapshot]
}
```

说明：

1. `service` 包不应该导入 `framework` 包。
2. `framework.App` 通过实现 `service.Host` 所需的窄接口接入。
3. 这样可以避免 `framework -> internal/ai/service -> framework` 循环。

### 6.4 `AppController`

职责：

1. 实现当前项目结构下的 App 级 AI 控制面。
2. 提供增强的树级检查接口。
3. 提供面向组件能力的写操作。

建议分层：

```go
type AppController struct {
    builder   *SnapshotBuilder
    executor  *Executor
    selectors *SelectorEngine
}
```

建议实现：

1. 保留 `runtime/ai.Controller` 中有价值的语义接口思想：
   - `Inspect()`
   - `Find()`
   - `Query()`
   - `WaitUntil()`
   - `Dispatch()`
   - `Click()`
   - `Input()`
   - `Navigate()`
2. 但不继续直接复用 `runtime/ai.RuntimeController` 实现。
3. 增加 App 专属扩展接口：
   - `InspectFrame(opts)`
   - `GetTree(kind, query)`
   - `GetNode(locator)`
   - `GetValue(locator)`
   - `SetValue(locator, value)`
   - `SetProp(locator, key, value)`
   - `Select(locator, value)`
   - `GetFormData(locator)`
   - `SetFormField(locator, field, value)`

### 6.5 `SnapshotBuilder`

职责：

1. 从 `framework.App` 与 `DeclarativeNode` 提取语义快照。
2. 构建轻量索引。
3. 对外输出 JSON-safe 结构。

建议拆成两条路径：

#### 轻量快照路径

每次 `render()` 后生成并缓存：

1. `state.Snapshot`
2. `ComponentIndex`
3. `NodeIndex`
4. `FocusPath`
5. `RenderSeq`

#### 重量级树路径

按需通过 `App.Invoke()` 构建：

1. `VNodeTreeSnapshot`
2. `FiberTreeSnapshot`
3. `LayoutTreeSnapshot`
4. `PaintableTreeSnapshot`
5. `HitMapSnapshot`

### 6.6 `Executor`

职责：

1. 执行所有 AI 写操作。
2. 统一走主循环调用。
3. 优先使用语义 Action，其次走能力适配器。

执行优先级：

1. 显式安全能力适配器。
2. `ActionBridge` / `ActionRouter` 语义分发。
3. 必要时的测试级原始事件注入回退。

说明：

1. 第一阶段不建议默认走 `InjectEvent()`，因为它目前语义是测试接口。
2. 第一阶段主要依赖 Action 与实例能力适配。

### 6.7 `CapabilityRegistry`

职责：

1. 识别组件实例支持的 AI 能力。
2. 统一封装不同组件的读写接口。
3. 在 `Inspect` 结果中附带可用能力列表。

建议能力集合：

```go
type Capability string

const (
    CapabilityAction     Capability = "action"
    CapabilityFocusable  Capability = "focusable"
    CapabilityValueRead  Capability = "value.read"
    CapabilityValueWrite Capability = "value.write"
    CapabilityPropRead   Capability = "prop.read"
    CapabilityPropWrite  Capability = "prop.write"
    CapabilitySelect     Capability = "select"
    CapabilityFormRead   Capability = "form.read"
    CapabilityFormWrite  Capability = "form.write"
)
```

建议初始适配覆盖：

1. `input`
2. `textarea`
3. `checkbox`
4. `button`
5. `form`
6. `select`
7. `optiongroup`

第二阶段再覆盖：

1. `list`
2. `table`
3. `treeview`
4. `virtuallist`

### 6.8 `internal/ai/mcp.Server`

职责：

1. 把 `AppController` 暴露为 MCP 工具和资源。
2. 管理会话与访问控制。
3. 隔离协议层与运行时层。

说明：

1. `mcp` 应只依赖 `service` 暴露的控制接口和 DTO。
2. `mcp` 不应直接读取 `framework.App`、`DeclarativeNode` 或 `ComponentInstance`。
3. 这样协议层可以独立演进，不把运行时指针泄漏到协议边界。

## 7. App 侧改造点

### 7.1 `framework.App` 新增字段

建议新增：

```go
type App struct {
    ...
    aiService *aiservice.Service
    invokeQ   chan invokeRequest
    renderSeq uint64
}
```

其中：

1. `aiService` 挂载 AI/MCP 子系统。
2. `invokeQ` 让外部 goroutine 把任务投递到主循环。
3. `renderSeq` 用于等待“下一帧已提交”。

### 7.2 `framework.App` 新增方法

```go
func (a *App) EnableAI(cfg AIConfig) error
func (a *App) AIController() *AIController
func (a *App) AIStatus() AIStatus
func (a *App) Invoke(ctx context.Context, fn func() (any, error)) (any, error)
```

说明：

1. `EnableAI` 只负责配置和挂接，不立即启动服务。
2. 真正启动发生在 `Init()` 之后、主循环开始之前。
3. `Invoke()` 是整个方案的线程安全基础设施，后续 Inspector/DevTools 也可复用。

### 7.3 `App.Init()` 改造

在 `pump.Start()`、`asyncRenderer.Start()` 完成后启动 AI 服务：

```go
if a.aiService != nil {
    if err := a.aiService.Start(); err != nil {
        return err
    }
}
```

原因：

1. 此时根节点已经设置完成。
2. 事件泵已经可用。
3. App 处于 Running 前夕，适合启动本地服务。

### 7.4 `App.render()` 改造

在已有 `HitMap` 更新之后、`dirty=false` 之前通知 AI 服务：

```go
if a.aiService != nil {
    a.renderSeq++
    a.aiService.OnAfterRender(ai.RenderContext{
        RenderSeq: a.renderSeq,
    })
}
```

`OnAfterRender()` 内部完成：

1. 轻量语义快照构建。
2. 索引更新。
3. watcher 通知。

### 7.5 `App.Close()` 改造

在关闭事件泵和恢复终端前停止 AI 服务：

```go
if a.aiService != nil {
    _ = a.aiService.Stop()
}
```

## 8. 高层入口集成

### 8.1 `ui.Run()` / `ui.RunApp()` 建议新增 Option

建议新增：

```go
type AIConfig = framework.AIConfig
type MCPConfig = framework.MCPConfig

func WithAI(cfg AIConfig) Option
func WithMCP(cfg MCPConfig) Option
```

实现方式：

1. 本质上仍通过 `ui.WithPluginSetup()` 调用 `fwApp.EnableAI(cfg)`。
2. 不破坏当前 `ui.Run()` 入口风格。
3. 允许纯 `framework.App` 和高层 `ui` API 共享同一实现。

### 8.2 自动启用策略

“自动启用”建议解释为：

1. 只要应用显式开启 `WithAI/WithMCP` 或 `EnableAI()`，服务就会在 `Run()` 启动后自动拉起。
2. 不建议所有交互式 App 默认无条件启动 MCP。

原因：

1. 安全风险。
2. 端口占用与部署复杂度。
3. 很多终端程序并不需要远程 AI 能力。

## 9. 数据模型设计

### 9.1 轻量快照 `FrameSnapshot`

```go
type FrameSnapshot struct {
    RenderSeq  uint64
    Timestamp  time.Time
    FocusPath  []string
    Snapshot   *state.Snapshot
    Indexes    SnapshotIndexes
    Warnings   []string
}
```

其中 `SnapshotIndexes` 建议包含：

```go
type SnapshotIndexes struct {
    ByComponentID map[string]NodeLocator
    ByNodeID      map[uint64]NodeLocator
    ByPath        map[string]NodeLocator
    ByActionID    map[string]NodeLocator
}
```

`NodeLocator`：

```go
type NodeLocator struct {
    ComponentID    string
    NodeID         uint64
    Path           string
    ActionTargetID string
    Tag            string
    Layer          string
}
```

### 9.2 扩展组件快照

在 `state.ComponentState` 的基础上，建议在 `Metadata` 中附加扩展信息：

```go
type ComponentMetadata struct {
    NodeID         uint64
    Path           string
    Key            string
    DiffKey        string
    Tag            string
    Layer          string
    Capabilities   []string
    InstanceType   string
    HasVNode       bool
    HasFiber       bool
    HasLayoutBox   bool
    HasPaintable   bool
}
```

设计原则：

1. `state.Snapshot` 保持语义视角。
2. 调试元信息通过 `Metadata` 或扩展结构承载。
3. 避免把整个 `Fiber`/`LayoutBox` 直接塞进 `Snapshot`。

### 9.3 树快照

四类树建议统一成同一风格：

```go
type TreeKind string

const (
    TreeVNode     TreeKind = "vnode"
    TreeFiber     TreeKind = "fiber"
    TreeLayout    TreeKind = "layout"
    TreePaintable TreeKind = "paintable"
)

type TreeNodeSnapshot struct {
    Kind       TreeKind
    NodeID     uint64
    Path       string
    Type       string
    Tag        string
    Key        string
    DiffKey    string
    ComponentID string
    Bounds     Rect
    Layer      string
    Props      map[string]any
    State      map[string]any
    Meta       map[string]any
    Children   []*TreeNodeSnapshot
}
```

说明：

1. `VNode` 树主要用于“描述层”调试。
2. `Fiber` 树是运行时身份和行为层。
3. `LayoutBox` 树是几何层。
4. `PaintableBox` 树是绘制层。

### 9.4 JSON-safe 序列化

必须新增统一 Sanitizer，处理以下问题：

1. `Props` 里可能有函数、闭包、`intent.Intent`、复杂对象。
2. `State` 里可能有不可 JSON 化对象。
3. 需要限制深度、数组长度、map 项数。

建议策略：

1. 基础类型原样输出。
2. `fmt.Stringer` 输出字符串。
3. `func` 输出 `"<func>"`。
4. `intent.Intent` 输出 `"<intent:Type>"`。
5. 复杂结构递归裁剪。
6. 循环引用输出 `"<cycle>"`。

## 10. 选择器与定位策略

### 10.1 兼容现有选择器

保留 `runtime/ai` 现有语法：

1. `#component-id`
2. `.Type`
3. `[key=value]`
4. `*`

### 10.2 新增调试定位语法

建议扩展：

1. `@12345`  
   按 `NodeID` 精确定位

2. `path:/root/base[0]/vstack[0]/...`  
   按路径定位

3. `layer:modal`  
   限定层

4. `cap:value.write`  
   查找支持某项 AI 能力的组件

原因：

1. `component_id` 不一定存在。
2. 调试树跨层定位更适合 `NodeID + Path`。

### 10.3 跨层映射规则

建议统一采用：

1. `NodeID`  
   `Fiber/LayoutBox/PaintableBox/HitMap` 主键。

2. `Path`  
   `VNode` 与 `Fiber` 对齐主键。

3. `ComponentID`  
   业务定位主键，优先取 `fiber.ID`，其次取实例 key，再次取 `ActionTargetID`。

## 11. 读路径设计

### 11.1 轻量 Inspect

`Inspect()` 走缓存：

1. 不直接触碰运行中树结构。
2. 返回最近一次渲染后的 `FrameSnapshot.Snapshot`。
3. 满足高频 `Find/Query/WaitUntil`。

### 11.2 树级 Inspect

`GetTree()` 走 `App.Invoke()`：

1. 在主循环中读取 `DeclarativeNode`。
2. 使用现有 getter：
   - `GetRenderedRoot()`
   - `GetFiberRoot()`
   - `GetLayoutRoot()`
   - `GetPaintableRoot()`
   - `GetHitMap()`
3. 立即序列化为不可变快照后返回。

这样可以避免：

1. 跨 goroutine 直接持有运行时指针。
2. 返回半更新状态。

### 11.3 轻量快照构建来源

语义快照建议以 `Fiber + Instance + LayoutRoot` 为主：

1. 组件身份：来自 `Fiber`。
2. 最新 props：优先来自 `Instance.GetProps()`。
3. 动态状态：来自能力适配器。
4. `Rect`：来自 `LayoutRoot` / `HitMap` 映射。
5. `Visible`：由布局与层可见性推导。
6. `Disabled`：优先来自 `FocusableInstance.IsDisabled()` 或 `GetProp("disabled")`。

说明：

1. 不建议直接把 `VNode.Props()` 当权威值。
2. 更不建议把 `Fiber.Props` 当最新值。

## 12. 写路径设计

### 12.1 统一执行流程

```text
MCP Tool
  -> AppController method
  -> App.Invoke()
  -> Resolve locator
  -> CapabilityRegistry choose strategy
  -> Execute mutation
  -> app.MarkDirty()
  -> wait next render if needed
  -> return updated snapshot
```

### 12.2 写操作优先级

### 第一优先级：语义 Action

适用于：

1. `click`
2. `input_text`
3. `navigate`
4. `focus`
5. `select`

优势：

1. 与“AI 是一级操作者”原则一致。
2. 能复用现有 `ActionBridge` / `ActionRouter` / `HandleAction()`。

### 第二优先级：安全能力写入

适用于：

1. `set_value`
2. `set_prop`
3. `set_form_field`
4. `select_option`

使用前提：

1. 组件实例显式暴露稳定写接口。
2. 该接口行为明确，不依赖终端原始输入。

### 第三优先级：原始事件回退

仅建议保留为兼容模式，不建议第一阶段默认启用。

### 12.3 典型写接口映射

### 输入框

1. 优先 `ActionInputText`。
2. 若需要“直接设值”，使用 `SetValue(string)`。

### Checkbox

1. 优先 `ActionToggle` / `ActionClick`。
2. 或通过 `SetProp("checked", bool)`。

### Form

1. 读取：`GetValues()` / `GetValue(field)`。
2. 写入：`SetValue(field, value)` / `SetValues(...)`。
3. 提交：`ActionSubmit`。

### OptionGroup / Select

1. 优先 `select(value)` 语义接口。
2. 底层映射到 `SelectOption(value)` / `DeselectOption(value)` 或相关 Action。

### 12.4 `WaitUntil` 设计

实现方式：

1. 先检查当前缓存快照。
2. 不满足则订阅 `OnAfterRender()` 产生的快照流。
3. 超时返回。

这样不需要轮询树结构，也不会在非主线程反复读运行时树。

## 13. MCP 接口设计

### 13.1 传输层

根据当前交互式 TUI 嵌入场景，建议：

### 默认方案：本地 Streamable HTTP

理由：

1. 不与 TUI 的 stdin/stdout 冲突。
2. 跨平台最直接。
3. 便于本地 IDE/Agent 连接。

建议默认监听：

1. `127.0.0.1`
2. 随机端口或配置端口
3. 启动时打印一次受控连接信息到日志，而不是标准输出主画面

### 预留方案：Windows Named Pipe / Unix Domain Socket

用于未来进一步收紧本地访问边界，但第一阶段不是必须。

### 不建议作为嵌入式默认方案：stdio

原因前面已说明：会和交互式终端 I/O 冲突。

### 13.2 MCP 暴露形态

建议同时提供：

1. `tools`
2. `resources`

其中：

1. `tools` 用于主动操作和参数化查询。
2. `resources` 用于读取当前快照、树和节点详情。

### 13.3 建议工具清单

### 读工具

1. `mint.inspect`
   - 返回轻量语义快照

2. `mint.find`
   - 按 selector 返回组件摘要

3. `mint.query`
   - 查询组件状态/属性

4. `mint.get_tree`
   - 获取 `vnode/fiber/layout/paintable` 树

5. `mint.get_node`
   - 获取单节点跨层详情

6. `mint.get_value`
   - 读取组件值

7. `mint.get_form_data`
   - 读取表单数据

8. `mint.wait_until`
   - 等待条件满足

### 写工具

1. `mint.dispatch`
2. `mint.click`
3. `mint.input`
4. `mint.focus`
5. `mint.navigate`
6. `mint.set_value`
7. `mint.set_prop`
8. `mint.select`
9. `mint.set_form_field`

### 13.4 建议资源 URI

1. `mint://frame/current`
2. `mint://tree/vnode`
3. `mint://tree/fiber`
4. `mint://tree/layout`
5. `mint://tree/paintable`
6. `mint://node/{node_id}`
7. `mint://component/{component_id}`
8. `mint://hitmap/current`

### 13.5 错误模型

建议统一错误码：

1. `app_not_running`
2. `component_not_found`
3. `node_not_found`
4. `unsupported_capability`
5. `read_only`
6. `timeout`
7. `serialization_error`
8. `invalid_selector`
9. `invalid_argument`

## 14. 快照构建细节

### 14.1 轻量快照构建顺序

建议顺序：

1. 获取 `FiberRoot`
2. 获取 `LayoutRoot`
3. 构建 `NodeID -> Rect`
4. 遍历 `Fiber`
5. 对每个 Fiber：
   - 生成 `ComponentID`
   - 抽取 `Type/Tag/Path/Key`
   - 从 `Instance.GetProps()` 提取 props
   - 从能力适配器提取 state
   - 从 `NodeID -> Rect` 填充 `Rect`
   - 推导 `Visible/Disabled`
   - 记录可用能力
6. 填充 `FocusPath`
7. 写入 `FrameSnapshot`

### 14.2 VNode 树快照构建

建议复用现有 Inspector 的能力：

1. 路径提取逻辑已经在 `internal/inspector/element_info.go` 中验证过。
2. `VNode.Key` 已经常常承载 `/root/...` 路径。
3. `ExtractElementInfo()` 可复用为第一阶段的 VNode 元信息构建器。

建议：

1. 不直接依赖 Inspector UI 代码。
2. 把纯提取逻辑下沉到 `internal/inspectmeta` 这类共享叶子包。

### 14.3 Layout / Paintable 树快照构建

已有 getter 足够：

1. `GetLayoutRoot()`
2. `GetPaintableRoot()`

需要补的只有：

1. 统一序列化结构。
2. 通过 `NodeID` 与语义快照对齐。

## 15. 安全与配置

### 15.1 配置结构建议

```go
type Config struct {
    Enabled      bool
    ReadOnly     bool
    AutoStart    bool
    Capture      CaptureConfig
    MCP          MCPConfig
}

type MCPConfig struct {
    Enabled      bool
    Transport    string // "http", "pipe"
    Host         string
    Port         int
    AuthToken    string
    ExposeTrees  bool
    ExposeWrite  bool
}
```

### 15.2 默认安全策略

建议默认值：

1. 仅监听 `127.0.0.1`
2. 自动生成随机 token
3. 支持 `ReadOnly=true`
4. 大树导出可关闭
5. 写操作可单独关闭

### 15.3 日志与审计

建议记录：

1. 客户端连接
2. 工具调用
3. 写操作目标
4. 执行动作类型
5. 超时和失败原因

不要记录：

1. 整个 UI 大快照原文
2. 大体量 props/state 原文
3. 敏感输入值的明文

## 16. 性能设计

### 16.1 默认只缓存轻量快照

原因：

1. `Fiber/Layout/Paintable` 全量树每帧序列化开销较大。
2. 大多数 AI 查询只需要组件状态与定位信息。

### 16.2 树级数据按需抓取

适用于：

1. 诊断布局问题
2. 调试渲染差异
3. 分析某个节点的跨层映射

### 16.3 大结果裁剪

建议支持：

1. `max_depth`
2. `max_children`
3. `include_props`
4. `include_state`
5. `root_locator`

### 16.4 Watcher 背压

若后续支持订阅：

1. 单客户端固定缓冲区。
2. 慢客户端只保留最新帧。
3. 不阻塞主渲染循环。

## 17. 测试策略

### 17.1 单元测试

新增以下测试组：

1. `SnapshotBuilder`
   - `Fiber -> Snapshot`
   - `Layout -> Rect`
   - `Instance props/state` 提取
   - JSON-safe 序列化

2. `CapabilityRegistry`
   - input/textarea/form/select/optiongroup/button/checkbox

3. `App.Invoke`
   - 主循环串行化
   - 超时
   - 退出时取消

4. `MCP tools`
   - tool 参数校验
   - 读操作
   - 写操作
   - 只读模式

### 17.2 集成测试

建议基于现有 `framework.App` / `ui.TestRun` / demo：

1. 启动应用后自动拉起 MCP。
2. `mint.inspect` 能看到 input/button/select。
3. `mint.input` 能更新 Input/Textarea。
4. `mint.select` 能更新 OptionGroup/Select。
5. `mint.get_tree` 能返回 fiber/layout/paintable。
6. 写操作后 `wait_until` 能等到下一帧快照。

### 17.3 回归测试重点

1. 不影响现有终端交互延迟。
2. 不影响现有 Inspector。
3. 不破坏 Action 路由。
4. 不引入数据竞争。

## 18. 分阶段实施计划

### Phase 0: `runtime/ai` 收缩与目录重构

目标：

1. 明确 `runtime/ai` 为 early prototype，不再承载主实现。
2. 拆出可复用的纯叶子能力，迁移到 `internal/inspectmeta` 或宿主侧公开类型。
3. 建立 `framework -> internal/ai/* -> runtime/*` 的单向依赖结构。
4. 清理会导致未来循环引用的目录方案。

交付物：

1. `runtime/ai/doc.go` 迁移说明或废弃说明。
2. `framework/ai_config.go` / `framework/ai_enable.go` / `framework/ai_invoke.go`。
3. `internal/ai/service/*` 基础骨架。
4. `internal/inspectmeta/*` 基础骨架。

### Phase 1: App 宿主与轻量快照

目标：

1. `framework.App.EnableAI()`
2. `App.Invoke()`
3. `AppController.Inspect/Find/Query/WaitUntil`
4. 每帧轻量 `state.Snapshot`

交付物：

1. `internal/ai/service/service.go`
2. `internal/ai/snapshot/builder.go`
3. `framework/ai_api.go`

### Phase 2: 树级元信息导出

目标：

1. `GetTree/GetNode`
2. `VNode/Fiber/Layout/Paintable/HitMap` 统一树快照
3. 统一 `NodeID/Path/ComponentID` 索引

### Phase 3: 组件写能力

目标：

1. `CapabilityRegistry`
2. `set_value`
3. `set_prop`
4. `select`
5. `set_form_field`

### Phase 4: MCP Server

目标：

1. 本地传输
2. tools/resources 暴露
3. 认证与只读模式

### Phase 5: 高层 UI API 集成

目标：

1. `ui.WithAI`
2. `ui.WithMCP`
3. demo 与文档

## 19. 风险与应对

### 风险 1：并发读写运行时树导致竞态

应对：

1. 所有重量级读取与全部写操作都走 `App.Invoke()`。
2. 缓存只保存不可变快照，不保存原始树指针。

### 风险 2：组件能力不统一，写操作覆盖率不足

应对：

1. 第一阶段先覆盖核心组件。
2. 对未覆盖组件统一返回 `unsupported_capability`。
3. 优先保留 Action 路径作为通用保底方案。

### 风险 3：元信息序列化膨胀

应对：

1. 轻量快照默认缓存。
2. 大树按需抓取。
3. 所有树接口支持裁剪参数。

### 风险 4：MCP 传输与终端 I/O 冲突

应对：

1. 嵌入式交互式 App 默认使用本地 HTTP。
2. 不在主终端 stdout 输出协议流。

### 风险 5：`Fiber.Props` 过期导致 Inspect 不一致

应对：

1. 快照构建优先读取实例 props。
2. 只把 `Fiber.Props` 当兜底或调试字段。

## 20. 最终建议

基于当前仓库状态，最务实的实现路线不是“从 `runtime/ai` 直接接一层 MCP”，而是：

1. 先完成 **`runtime/ai` 收缩与目录重构**，不要继续在原型目录上叠实现。
2. 再在 `framework.App` 上建立 **AI Service + 主循环调用桥 + 轻量快照**。
3. 再把 `DeclarativeNode` 已有的 `VNode/Fiber/Layout/Paintable/HitMap/Instance` 读能力组织成统一树快照。
4. 再通过 **CapabilityRegistry + ActionBridge** 打通写路径。
5. 最后以 **本地 Streamable HTTP MCP Server** 方式对外暴露。

这样做的原因很直接：

1. 当前仓库最成熟、最稳定的是真实运行中的 `framework.App` 和 Fiber-first 渲染链路。
2. 用户要求的所有核心能力，其实大部分“底层原料”已经存在。
3. 真正缺的是“统一宿主、线程安全调用、目录边界、快照模型、协议导出”，不是渲染或交互基础设施本身。

这个方向改动最小、复用最多、能最快得到可工作的第一版。
