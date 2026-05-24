# FormDialog

`formdialog` is a Fiber-first framed form dialog for bounded operational tasks:
create/edit settings, collect an audit reason, run a scoped action, or submit a
short configuration form without leaving the current table or page context.

## Capabilities

- Framed form wrapper built from `Form + Button`.
- Controlled open state through `Open(...)` / `Opened()` / `Closed()`.
- Form identity through `FormID(...)`, with normal field binding support.
- `vertical` / `horizontal` / `inline` form layout.
- Initial `Values(...)` and `Value(...)`.
- Submit, cancel, and close intents.
- Submit button variants, disabled state, and disabled reason text.
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

## Operational Guidance

- Use `FormDialog` for short bounded forms where the operator should preserve the
  current list/table context.
- Use `Drawer` for longer create/edit workflows that need multiple sections or
  extended inspection.
- Use `ConfirmDialog` when the task only needs target summary, warning text, and
  an audit reason.
- Keep dangerous operations gated by a clear target, reason, disabled reason,
  and final submit intent.

## Tests

- Unit: `go test ./ui/components/formdialog`
- SDK shortcut: `go test ./ui -run FormDialog`
- E2E: `go test ./ui/e2e -run "^TestE2EFormDialog"`
