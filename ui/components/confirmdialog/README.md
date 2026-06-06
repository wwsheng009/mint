# ConfirmDialog

`confirmdialog` 是面向运维和管理台危险操作的 Fiber-first 确认弹窗组合组件。它通过 centered `Modal` 承载确认界面，并复用 `Descriptions`、`Input` 和 `Button`，提供统一的目标摘要、影响说明、审计 reason、确认/取消 intent。

## 能力

- `Open(true/false)` 控制弹窗是否渲染。
- 打开时运行时根节点是 fixed centered modal-layer surface；不要用普通 `vstack` / `panel` 只设置 `LayerModal` 替代，否则只能保证覆盖层级，不能保证相对 viewport 居中。
- `Title(...)` / `Message(...)` / `Warning(...)` 展示操作上下文。
- `Target(...)` / `SensitiveTarget(...)` / `APITarget(...)` / `ImpactTarget(...)` 展示目标摘要、敏感目标、将调用的 API route 和操作影响。
- `ReasonField(...)` 通过 `intent.BindField(...)` 接入 Store/Reducer。
- `ReasonRequired(true)` 在 reason 为空时禁用确认按钮。
- `AuditReason(field, value)` 快速启用必填审计 reason。
- `ConfirmPhrase(expected, field, value)` / `RequirePhrase(expected, field, value)` 增加二次确认文本，输入值必须和 expected 完全一致才启用确认按钮。
- 确认按钮禁用时会在 footer 前自动显示低强调原因，例如缺少 reason、确认短语不匹配或外部 `DisabledReason(...)`。
- 启用二次确认文本时，弹窗会保护最小高度，避免确认输入框和 footer 挤压。
- `Danger(confirmText)` 快速配置危险操作确认按钮。
- `NewDangerOperation(...)` 生成带目标摘要、warning、必填 reason 和 typed intent 的危险操作确认模板。
- `OnConfirm(...)` / `OnCancel(...)` 派发 typed Intent。
- `ConfirmVariant(...)`、`ConfirmText(...)`、`CancelText(...)` 定制操作语义。
- Fiber-first：`VNode + Instance + RuntimeChildrenProvider`。

## 示例

```go
confirmdialog.NewBuilder().
    Key("disable-key.confirm").
    Title("Disable Provider Key").
    Message("Disable the selected provider key and record an audit reason.").
    Warning("Traffic may fail over to other available keys.").
    Open(state.ShowConfirm).
    Target(confirmdialog.APITarget("POST", "/admin/loadbalancer/providers/{provider}/keys/{key}/disable")).
    Target(confirmdialog.ImpactTarget("high", "Traffic may fail over to other available keys.")).
    Target(confirmdialog.Target("provider", "Provider", providerName)).
    Target(confirmdialog.SensitiveTarget("key", "Key", keyID)).
    ReasonField("actionReason").
    ReasonValue(state.ActionReason).
    ReasonRequired(true).
    ConfirmPhrase("DISABLE", "confirmPhrase", state.ConfirmPhrase).
    ConfirmText("Disable").
    ConfirmVariant(button.VariantDanger).
    OnConfirm(DisableKeyIntent{}).
    OnCancel(CloseConfirmIntent{}).
    Build()
```

运维危险操作可以使用预设减少重复配置：

```go
confirmdialog.NewDangerOperation(
    "disable-key.confirm",
    "Disable Provider Key",
    "Disable the selected provider key and record an audit reason.",
    "Traffic may fail over to another available key.",
    "Disable",
    "actionReason",
    state.ActionReason,
    DisableKeyIntent{},
    CloseConfirmIntent{},
    confirmdialog.APITarget("POST", "/admin/loadbalancer/providers/{provider}/keys/{key}/disable"),
    confirmdialog.ImpactTarget("high", "Traffic may fail over to another available key."),
    confirmdialog.Target("provider", "Provider", providerName),
    confirmdialog.SensitiveTarget("key", "Key", keyID),
).Build()
```

需要更高风险确认时，可以叠加二次确认短语：

```go
confirmdialog.NewDangerOperation(
    "clear-circuit-breakers.confirm",
    "Clear Circuit Breakers",
    "Clear circuit breaker state for the selected scope.",
    "This may immediately send traffic back to previously unhealthy targets.",
    "Clear",
    "actionReason",
    state.ActionReason,
    ClearCircuitBreakersIntent{},
    CloseConfirmIntent{},
    confirmdialog.Target("scope", "Scope", scopeLabel),
).
    RequirePhrase("CLEAR", "confirmPhrase", state.ConfirmPhrase).
    Build()
```

## 运维建议

- 禁用、清理熔断、运行态重置、reload、删除等操作应通过 `ConfirmDialog` 或等价确认流程。
- 弹窗内必须展示清晰目标，例如 group/provider/key。
- 影响运行态或配置的操作建议展示将调用的 HTTP method/path，便于操作者在确认前核对具体 Admin API。
- 影响运行态、流量路由或 provider/key 可用性的操作建议使用 `ImpactTarget(level, text)` 给出结构化影响摘要，避免只把影响藏在 warning 文本里。
- 当 reason、确认短语和 footer 同时占用高度且目标摘要过多时，`ConfirmDialog` 会约束 target summary 数量；provider key 类操作会优先保留 provider、key、key state、route state 等关键核对字段，避免小视口下只看到动作目标却看不到派生路由阻断。
- 需要审计的操作必须绑定 reason 字段，且 reason 为空时禁用确认。
- 删除、清理全部、批量禁用、清理熔断等高风险动作建议叠加 `RequirePhrase(...)`，避免误触。
- 当 reason 或确认短语未满足时，`ConfirmDialog` 会自动提示确认按钮禁用原因；业务层只需要提供额外的外部阻断原因，例如目标快照失效或权限不足。
- 不要在目标摘要中渲染完整 secret、token、authorization header 或 provider key。

## 测试

- 单测：`go test ./ui/components/confirmdialog`
- E2E：`go test ./ui/e2e -run "^TestE2EConfirmDialog"`
