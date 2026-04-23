# 交互式 E2E 测试文档

本目录描述 Mint 面向交互型 TUI 应用的端到端测试框架设计。

## 当前状态

- Phase 1 最小实现已落地到 `ui/e2e/`
- Phase 2 的基础 selector / bounds 能力已部分落地
- Phase 2.5 的基础 trace / hit assertion 能力已部分落地
- Phase 2.6 的焦点等待 / 焦点转移断言已部分落地
- 已开始覆盖 modal/overlay 这类复杂交互场景
- 当前已包含 backdrop close / overlay hit / modal 内按钮命中等 fixture 回归
- 当前已包含 select overlay popup 的 filter / commit / outside click close 回归
- 当前已包含 treeview 的 search / lazy load / selection / drag reorder 联动回归
- 当前已包含 steps 的 keyboard navigation / vertical click 回归
- 当前已提供：
  - `Run` / `RunWithSandbox`
  - `Driver`
  - mouse `Press` / `Move` / `Release` / `Drag`
  - `AwaitIdle`
  - RawInput trace + Msg trace + Action trace + Intent dispatch log trace
  - 基础 locator（`At` / `ByText` / `ByID` / `ByKey` / `ByTag` / `ByTargetID` / `ByComponentID` / `Focused`）
  - 基础焦点断言
  - 基于 cell 的样式断言
  - `ResolveFiber` / `BoundsOf` / `AssertBounds`
  - `AssertHit` / `AssertVisible`
  - `AssertIntentSequence`
  - `AssertLastIntent` / `AssertIntentHandled`
  - `AssertMessageSequence` / `AssertActionSequence`
  - `AwaitTrace`
  - `AwaitMessage` / `AwaitAction`
  - `AssertLastMessage` / `AssertLastAction`
  - `AssertActionHandled`
  - `AssertTargetID`
  - `AssertTraceContains` / `AssertTraceSequence`
  - `DiagnosticsSnapshot` / `SaveDiagnostics` / `SaveDiagnosticsTemp` / `SaveDiagnosticsOnFailure`
  - `AwaitFocus` / `Eventually`
  - `FocusTransitions` / `AssertFocusTransition`

## 文档

- `API_REFERENCE.md`  
  当前 `ui/e2e` 已落地 API 的正式参考，适合直接查能力面、方法清单、限制项与推荐用法。

- `INTERACTIVE_E2E_SUITE_DESIGN.md`  
  交互式 E2E 测试套件的完整设计：目标、架构、消息/焦点/样式观测面、断言模型、分阶段落地计划。

## 适用范围

该设计主要面向以下测试场景：

- 键盘/鼠标驱动的复杂交互流程
- 焦点切换、可达性与导航路径验证
- Action / Msg / Intent 级消息处理验证
- 样式、高亮、边框、选中态等视觉语义验证
- Portal / Overlay / HitMap / 异步状态更新类组件验证

## 与现有工具关系

现有 `ui/test.go` 仍适合：

- 快速 smoke test
- 简单渲染断言
- 基础事件注入

新的 E2E 测试套件设计目标不是替代它，而是在其现有能力之上，补齐：

- 结构化消息观测
- 确定性 idle/flush
- 焦点与命中路径断言
- 样式级断言
- selector/locator
- 交互回放与追踪
