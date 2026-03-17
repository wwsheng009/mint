# `ui/e2e` API 参考

本文档总结当前 `ui/e2e` 已落地的最小交互式 E2E 测试 API。

---

## 1. 入口

### `Run`

```go
app, err := e2e.Run(MyApp, ui.WithSize(80, 24))
defer app.Close()
```

用途：

- 启动完整 `framework.App`
- 自动接好 `IntentRuntime`
- 自动启用 `Msg/Action` 测试探针
- 自动等待初始稳定态

### `RunWithSandbox`

```go
app, err := e2e.RunWithSandbox(MyApp, ui.WithSize(80, 24))
```

用途：

- 与 `Run` 类似
- 但底层事件源使用 `MockSandbox`

---

## 2. App 核心方法

### 生命周期

- `Close() error`

### 基础访问

- `Driver() *Driver`
- `FrameworkApp() *framework.App`
- `IntentRuntime() *intent.Runtime`
- `RenderString() string`
- `ForceRender()`
- `HitMap() *event.HitMap`
- `RootFiber() *rtui.Fiber`

### 等待类

- `AwaitIdle() error`
- `AwaitIdleFor(timeout time.Duration) error`
- `AwaitIntent(intentType string, timeout time.Duration) error`
- `AwaitMessage(name string, timeout time.Duration) error`
- `AwaitAction(actionType action.ActionType, timeout time.Duration) error`
- `AwaitTrace(match TraceMatch, timeout time.Duration) error`
- `AwaitFocus(locator Locator, timeout time.Duration) error`
- `Eventually(timeout, interval time.Duration, fn func(*App) error) error`

### 快照与导出

- `FocusSnapshot() (FocusSnapshot, bool)`
- `DiagnosticsSnapshot() DiagnosticsSnapshot`
- `SaveDiagnostics(dir string) error`
- `SaveDiagnosticsTemp(prefix string) (string, error)`
- `SaveDiagnosticsOnFailure(t *testing.T, prefix string) (string, error)`

---

## 3. Driver

`Driver` 表达高层用户动作，而不是裸事件。

### 键盘

- `Key(rune) error`
- `KeyWithMod(rune, platform.KeyModifier) error`
- `Special(platform.SpecialKey) error`
- `SpecialWithMod(platform.SpecialKey, platform.KeyModifier) error`
- `Type(text string) error`

### 鼠标

- `Click(locator Locator) error`
- `ClickAt(x, y int) error`

默认行为：

- 注入事件后自动 `AwaitIdle()`
- 自动记录 `RawInput` trace
- 自动记录焦点转移

---

## 4. Locator

`Locator` 用于定位可交互或可断言目标。

### 当前支持

- `At(x, y int)`
- `ByText(text string)`
- `ByID(id string)`
- `ByKey(key string)`
- `ByTag(tag string)`
- `ByTargetID(targetID string)`
- `ByComponentID(componentID string)`
- `Focused()`

### 当前解析来源

- Fiber tree
- HitMap
- 当前焦点
- 渲染文本

---

## 5. Focus / Bounds / Visibility

### 焦点

- `AssertFocus(locator Locator) error`
- `FocusTransitions() []FocusTransition`
- `ClearFocusTransitions()`
- `AssertFocusTransition(from, to Locator) error`

### 定位与 bounds

- `ResolveFiber(locator Locator) (*rtui.Fiber, error)`
- `ResolvePoint(locator Locator) (Point, error)`
- `BoundsOf(locator Locator) (layout.Rect, error)`
- `AssertBounds(locator Locator, expect layout.Rect) error`

### 可见性 / 命中

- `AssertVisible(locator Locator) error`
- `AssertHit(point Point, locator Locator) error`
- `AssertTargetID(locator Locator, targetID string) error`

---

## 6. 样式断言

### 单 cell 样式

- `CellAt(x, y int) (CellSnapshot, error)`
- `AssertCellStyleAt(x, y int, expect StyleExpect) error`

### 基于 locator 的样式

- `AssertStyle(locator Locator, expect StyleExpect) error`

### `StyleExpect`

当前支持子集匹配：

- `FG`
- `BG`
- `Bold`
- `Italic`
- `Underline`
- `Reverse`

每项都有 `HasXxx` 开关，未设置的字段不参与比较。

---

## 7. Trace / Probe

### 原始输入

- `RawInputs() []RawInputEvent`
- `ClearRawInputs()`

说明：

- `ClearRawInputs()` 会同时清空当前 `RawInput` / `Message` / `Action` / `Trace` / `FocusTransitions` 本地缓冲
- `Intent` dispatch log 需要单独调用 `ClearIntentLogs()`

### Msg / Action

- `MessageEvents() []MessageEvent`
- `ActionEvents() []ActionEvent`
- `LastMessage() (MessageEvent, bool)`
- `LastAction() (ActionEvent, bool)`

### Intent

- `IntentLogs() []intent.DispatchLog`
- `ClearIntentLogs()`
- `LastIntentLog() (intent.DispatchLog, bool)`

### 统一 trace

- `TraceEvents() []TraceEvent`

### Trace kind

- `TraceRawInput`
- `TraceMsg`
- `TraceAction`
- `TraceIntentDispatch`
- `TraceFocusTransition`

### 通用 trace matcher

```go
TraceMatch{
    Kind: e2e.TraceAction,
    Name: "enter",
}
```

---

## 8. 消息 / 动作 / 意图断言

### Intent

- `AssertIntentSequence(intentTypes ...string) error`
- `AssertLastIntent(intentType string) error`
- `AssertIntentHandled(intentType string) error`

### Message

- `AssertMessageSequence(names ...string) error`
- `AssertLastMessage(name string) error`

### Action

- `AssertActionSequence(types ...action.ActionType) error`
- `AssertLastAction(actionType action.ActionType) error`
- `AssertActionHandled(actionType action.ActionType, stage string) error`

### 通用 Trace

- `AssertTraceContains(match TraceMatch) error`
- `AssertTraceSequence(matches ...TraceMatch) error`

---

## 9. Diagnostics

### `DiagnosticsSnapshot`

当前包含：

- `Render`
- `Focus`
- `RawInputs`
- `Messages`
- `Actions`
- `Intents`
- `FocusTransitions`
- `Trace`
- `HitMap`

### 导出文件

`SaveDiagnostics(dir)` 当前会生成：

- `render.txt`
- `diagnostics.json`
- `trace.json`

---

## 10. 当前已验证的复杂场景

通过 `ui/e2e` 包内回归，当前至少已覆盖：

- 基础按钮交互
- Fiber selector / bounds
- RawInput / Msg / Action / Intent / Focus trace
- Diagnostics 导出
- Modal / overlay
  - 打开
  - backdrop close
  - overlay 内按钮命中
  - background click 不泄漏
- Select overlay popup
  - 打开 / 关闭
  - filter 输入
  - popup 内命中与样式断言
  - 提交后状态更新
  - outside click close 且不泄漏到 background action

---

## 11. 当前限制

当前 `ui/e2e` 仍是 Phase 1 / 2 演进中的最小可用层，暂未覆盖：

- 自动失败时强制导出 diagnostics（目前提供 helper，需要调用方自行通过 `t.Cleanup` 接入）
- 通用 `AssertLastTrace(...)`
- `Msg` / `Action` payload 级 typed matcher
- 更丰富的 overlay 矩阵（例如 menu / select tags create-tag / tooltip delay）
- 手动测试时钟 / tick 控制
- 无 sleep 的更强 idle 语义

---

## 12. 推荐用法

### 最小交互脚本

```go
app, err := e2e.Run(MyApp, ui.WithSize(80, 24))
defer app.Close()

app.Driver().Special(platform.KeyTab)
app.Driver().Special(platform.KeyEnter)

if err := app.AssertFocus(e2e.ByID("submit-btn")); err != nil {
    t.Fatal(err)
}
```

### 链路断言

```go
if err := app.AssertTraceSequence(
    e2e.TraceMatch{Kind: e2e.TraceRawInput, Name: "key:Enter"},
    e2e.TraceMatch{Kind: e2e.TraceMsg, Name: "key"},
    e2e.TraceMatch{Kind: e2e.TraceAction, Name: "enter"},
    e2e.TraceMatch{Kind: e2e.TraceIntentDispatch, Name: "form.Submit"},
); err != nil {
    t.Fatal(err)
}
```

### 失败诊断

```go
t.Cleanup(func() {
    _, _ = app.SaveDiagnosticsOnFailure(t, "mint-e2e-")
})

if err := app.AssertFocus(e2e.ByID("expected")); err != nil {
    t.Fatal(err)
}
```
