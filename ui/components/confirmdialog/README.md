# ConfirmDialog

`confirmdialog` 是面向运维和管理台危险操作的 Fiber-first 确认弹窗组合组件。它在 modal layer 上合成带边框的确认界面，并复用 `Descriptions`、`Input` 和 `Button`，提供统一的目标摘要、影响说明、审计 reason、确认/取消 intent。

## 能力

- `Open(true/false)` 控制弹窗是否渲染。
- `Title(...)` / `Message(...)` / `Warning(...)` 展示操作上下文。
- `Target(...)` / `SensitiveTarget(...)` 展示目标摘要，敏感目标默认脱敏。
- `ReasonField(...)` 通过 `intent.BindField(...)` 接入 Store/Reducer。
- `ReasonRequired(true)` 在 reason 为空时禁用确认按钮。
- `AuditReason(field, value)` 快速启用必填审计 reason。
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
    Target(confirmdialog.Target("provider", "Provider", providerName)).
    Target(confirmdialog.SensitiveTarget("key", "Key", keyID)).
    ReasonField("actionReason").
    ReasonValue(state.ActionReason).
    ReasonRequired(true).
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
    confirmdialog.Target("provider", "Provider", providerName),
    confirmdialog.SensitiveTarget("key", "Key", keyID),
).Build()
```

## 运维建议

- 禁用、清理熔断、运行态重置、reload、删除等操作应通过 `ConfirmDialog` 或等价确认流程。
- 弹窗内必须展示清晰目标，例如 group/provider/key。
- 需要审计的操作必须绑定 reason 字段，且 reason 为空时禁用确认。
- 不要在目标摘要中渲染完整 secret、token、authorization header 或 provider key。

## 测试

- 单测：`go test ./ui/components/confirmdialog`
- E2E：`go test ./ui/e2e -run "^TestE2EConfirmDialog"`
