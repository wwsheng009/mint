# FormDialog

`formdialog` is a Fiber-first modal form dialog for bounded operational tasks:
create/edit settings, collect an audit reason, run a scoped action, or submit a
short configuration form without leaving the current table or page context.

## Capabilities

- Modal form wrapper built from `Modal + Form + Button`, rendered on `LayerModal`.
- Controlled open state through `Open(...)` / `Opened()` / `Closed()`.
- Form identity through `FormID(...)`, with normal field binding support.
- `Target(...)` / `SensitiveTarget(...)` show a compact operation target summary before fields.
- `vertical` / `horizontal` / `inline` form layout.
- Initial `Values(...)` and `Value(...)`.
- Submit, cancel, and close intents.
- Submit button variants, disabled state, and disabled reason text.
- `AuditReason(field, value)` appends a required audit reason field using the current form id.
- `NewDangerReasonAction(...)` builds an opened danger form dialog for short operational actions.
- Close policy fields are part of the declarative contract; app reducers still
  control whether the dialog is shown.
- Children stay declarative VNodes; no callback state is stored in the component.
- Component lives under `ui/components/formdialog/` and follows Fiber-first
  `VNode + Instance + RuntimeChildrenProvider` shape.

## Example

```go
ui.NewFormDialogBuilder().
    Key("reload-runtime").
    Title("Reload Runtime").
    Description("Reload runtime configuration for the selected gateway instance.").
    FormID("reload-runtime-form").
    Open(true).
    Width(76).
    Height(22).
    Target(ui.FormDialogTarget("instance", "Instance", "gateway-a")).
    Target(ui.FormDialogSensitiveTarget("token", "Admin token", "agw_example_token")).
    Children(
        ui.FormInputItem(
            "reason",
            "Reason",
            state.ActionReason,
            ui.FormInputForForm("reload-runtime-form"),
            ui.FormInputWidth(52),
            ui.FormInputValidators(ui.Required()),
        ),
    ).
    SubmitText("Reload").
    SubmitVariant(ui.ButtonVariantDanger).
    OnSubmit(ReloadRuntimeIntent{}).
    OnCancel(CloseDialogIntent{}).
    Build()
```

For a short high-risk operation form with only an audit reason, use the preset:

```go
formdialog.NewDangerReasonAction(
    "reload-runtime-dialog",
    "Reload Runtime",
    "Reload runtime configuration for the selected gateway instance.",
    "reload-runtime-form",
    "actionReason",
    state.ActionReason,
    "Reload",
    ReloadRuntimeIntent{},
    CloseDialogIntent{},
).Build()
```

## Operational Guidance

- Use `FormDialog` for short bounded forms where the operator should preserve the
  current list/table context.
- Use `Drawer` for longer create/edit workflows that need multiple sections or
  extended inspection.
- Use `ConfirmDialog` when the task only needs target summary, warning text, and
  an audit reason.
- Keep dangerous operations gated by a clear target, reason, disabled reason,
  and final submit intent.
- Put selected group/provider/key/job/alert metadata in `Target(...)`; use
  `SensitiveTarget(...)` for anything that may contain a token or secret.

## Tests

- Unit: `go test ./ui/components/formdialog`
- SDK shortcut: `go test ./ui -run FormDialog`
- E2E: `go test ./ui/e2e -run "^TestE2EFormDialog"`
