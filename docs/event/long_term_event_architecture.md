# 长期事件架构优化方案（Mint TUI）

## 背景与痛点
- **命中分散**：鼠标命中依赖各组件手写 `SetBounds/ContainsPoint`，遗漏即丢事件。Inspector、Tabs 曾因未写或接口不一致导致点击/hover 失效。
- **阶段缺失**：事件只有“谁先处理谁截断”，没有捕获/目标/冒泡语义，拦截与合成（如全局快捷键 vs 局部控件）难以推理。
- **输入转译不统一**：RawInput → Event → 组件的分发路径与 Inspector 特例混杂，Wheel 缺 delta，Move 高频无节流。
- **测试脆弱**：测试用例需写绝对坐标，布局变化即失效；缺少按节点路径/ID 注入的能力。
- **异步散乱**：Tick/IO/goroutine 分布在组件里，缺少集中 Cmd 管理，难测难回收。

## 借鉴 Bubble Tea 的原则
- **消息单一化**：一切输入/异步产出皆 Msg。
- **纯函数 Update**：状态演进在 `Update(Msg)`，输出 Cmd 副作用。
- **渲染分离**：渲染只读 Model（VNode 树）。
- **可组合 Cmd**：异步、定时、IO 均以 Cmd 表达，便于测试与回收。

## 目标状态（适配 Mint）
1) **统一 Msg 层**  
   - 定义 `Msg`：`KeyMsg{Rune,Special,Mod}`、`MouseMsg{X,Y,Button,Action,Delta,TargetID,LocalX,LocalY}`、`ResizeMsg{W,H}`、`TickMsg`、`CustomMsg any`。  
   - Pump 直接输出 Msg（保留兼容桥接到旧 Event）。

2) **集中 HitTest 与路径**  
   - 布局阶段生成 `HitMap`：节点 ID → 绝对 bounds，保留 Z 序。  
   - 提供 `HitTest(x,y)` 返回 `TargetID`、`localXY`、路径（root→target）。  
   - MouseMsg 在 Pump 阶段即带 TargetID/localXY，组件不再手写命中。

3) **阶段化分发**  
   - 捕获 → 目标 → 冒泡，允许 `StopPropagation/PreventDefault`。  
   - 支持监听优先级（如全局快捷键、Inspector overlay）。

4) **组件 Update API**  
   - `Update(Msg) (Cmd, bool)`；bool=consumed。  
   - DeclarativeNode 按路径（有 targetID 时定向）或按层次分发；旧 `HandleEvent` 由适配器桥接。

5) **Cmd 体系**  
   - 标准 Cmd：`After(d)`, `Tick(d)`, `Batch`, `IO`（占位）。  
   - App 主循环统一执行 Cmd → Msg，组件不再自起 goroutine。

6) **测试与工具**  
   - `TestableApp.InjectMsg`，`InjectMouseByID(nodeID, action)` 利用 HitMap 自动定位坐标。  
   - Debug: 渲染 HitMap/命中路径（TUI_DEBUG_UI），快速诊断丢事件。

## 分阶段落地计划

### Phase 1：命中与 Wheel 增强（低风险，立即收益）
- Render 后生成 `HitMap`（由 Renderer/compute 导出最新布局）。  
- Pump 填充 MouseMsg 的 `TargetID/localXY`；向旧 MouseEvent 也补充这些字段。  
- 扩展 MouseWheel：增加 `Delta`（+1/-1）；TreeView/ScrollView 支持 delta 滚动。  
- 加入 Move/Wheel 节流（scheduler/mouse.go）：默认 120Hz→60Hz，或按像素阈值。

### Phase 2：Msg 适配层（兼容过渡）
- 在 App loop：Event → Msg；若组件未实现 Update，桥接到旧 HandleEvent。  
- DeclarativeNode 优先用 target 路径定向 Update；无 target 时走层次递归。  
- Inspector overlay 改为普通节点消费 MouseMsg/KeyMsg，不再手写坐标解析。

### Phase 3：Cmd/异步统一
- 提供 Cmd 工具包；Pump 接收 Cmd 产出的 Msg。  
- 重构 Tick/定时逻辑（文本光标闪烁、性能采样等）为 Cmd；删除分散 goroutine。

### Phase 4：阶段化 & 优先级
- 引入捕获/冒泡模型：路径由 HitMap 提供。  
- 注册监听可带优先级（如全局快捷键>overlay>应用控件）。  
- 为 Debug 提供事件轨迹日志。

### Phase 5：测试与可视化
- TestableApp 支持按节点 ID 注入；提供 `AssertHit(nodeID)`、`DumpHitMap()`。  
- TUI_DEBUG_UI：可视化命中框、最近处理路径。

## 当前合理性评估
- **合理**：单线程 loop + Pump 渠道模式简单可靠；Hook/Inspector 已可工作。  
- **不足**：命中分散、阶段缺失、滚轮方向缺失、测试坐标脆弱、异步分散。  
- **优先级排序**：HitMap & WheelDelta > Msg 适配 > Cmd 统一 > 阶段化。

## 预期收益
- 点击/hover 稳定性显著提升（命中不再依赖每个控件手写 bounds）。  
- 事件行为可预测（捕获/冒泡 + 优先级）。  
- 测试可维护性提升（按节点 ID 注入）。  
- Inspector/Overlay 去特例化；滚轮与高频鼠标性能改善。  
- 统一的 Cmd/Msg 便于追踪和回收异步，降低内存与并发风险。

## 风险与缓解
- **改动面大**：用适配层保留旧 HandleEvent，逐步迁移。  
- **性能**：HitMap 构建需高效（可复用布局树，O(n)）；Move/Wheel 节流减负。  
- **API 变更**：为 Update 引入但不强制；文档、示例、测试同步更新。

## 近期待办（第一波）
1. Render 阶段导出 `HitMap`；MouseEvent/MouseMsg 写入 Target/local。  
2. MouseWheel 增加 Delta；TreeView/ScrollView 接入 delta。  
3. Move/Wheel 简单节流（60Hz）。  
4. 添加注入辅助：`InjectMouseByID(nodeID, action)`（TestableApp）。

完成后再推进 Msg/Update/Cmd 适配层。以上步骤不破坏现有 API，风险可控。 

