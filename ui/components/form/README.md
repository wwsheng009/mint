# Form

表单容器组件，负责字段值、校验、提交/重置和字段级元状态的统一管理。

## 已支持

- `FormItem` 包装层
- `vertical` / `horizontal` / `inline` 布局
- 初始 `values`
- `submit` / `reset` intent
- 字段级 `touched` / `dirty` / `submitted`
- `FormItem.Help(...)` 字段帮助文本
- `FormItem.Required(true)` 必填字段标记
- `HasSubmitted()` / `GetSubmitCount()`
- `Field` / `Form` 绑定与实例树内 `GetFormContext(...)`

## 示例

```go
loginForm := ui.NewForm("loginForm").
    Label("Login").
    Layout(ui.FormHorizontal).
    AddChildren(
        ui.NewFormItem(
            "username",
            ui.NewInputBuilder().
                Placeholder("Username").
                Build(),
        ).Label("Username").Required(true).Help("Use your admin account name.").Build(),
        ui.NewFormItem(
            "password",
            ui.NewInputBuilder().
                Password().
                Placeholder("Password").
                Build(),
        ).Label("Password").Build(),
    )
```

如果只需要快速把若干字段包进一个默认表单，也可以直接用 `ui.Form(...)`。

## 输入项快捷构造

常见业务表单可以通过 `ui/components/forminputitem` 提供的声明式组合能力快速生成 `FormItem + Input + FieldBinding`；根包 `ui.FormInputItem`、`ui.FormPasswordItem`、`ui.FormSearchItem` 是面向 SDK 使用者的薄转发入口，避免在每个页面重复拼装 label、输入框和字段绑定。

```go
loginForm := ui.NewForm("loginForm").
    Label("Login").
    Layout(ui.FormHorizontal).
    AddChildren(
        ui.FormInputItem(
            "baseURL",
            "Gateway URL",
            state.BaseURL,
            ui.FormInputPlaceholder("http://127.0.0.1:8080"),
            ui.FormInputWidth(72),
            ui.FormInputForForm("loginForm"),
            ui.FormInputValidators(ui.Required()),
        ),
        ui.FormPasswordItem(
            "token",
            "Admin Token",
            state.Token,
            ui.FormInputWidth(72),
            ui.FormInputForForm("loginForm"),
        ),
    )
```

可用选项包括：

- `FormInputPlaceholder(text)`
- `FormInputWidth(width)`
- `FormInputType(inputType)`
- `FormInputDisabled(disabled)`
- `FormInputReadOnly(readOnly)`
- `FormInputMaxLen(maxLen)`
- `FormInputForForm(formID)`
- `FormInputLayout(layout)`
- `FormInputValidators(validators...)`
- `FormInputHelp(text)`
- `FormInputRequired()`

如需扩展内置选项，可以实现 `FormInputItemOption` 并修改公开的 `FormInputItemConfig`：

```go
custom := func(cfg *ui.FormInputItemConfig) {
    cfg.Placeholder = "name@example.com"
    cfg.Validators = []ui.Validator{ui.Email()}
}
```
