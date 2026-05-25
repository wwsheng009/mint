# FormInputItem

`forminputitem/` 是 `FormItem + Input + FieldBinding` 的声明式组合组件，用于减少业务表单中重复的字段拼装代码。

## 示例

```go
node := ui.FormInputItem(
    "baseURL",
    "Gateway URL",
    state.BaseURL,
    ui.FormInputPlaceholder("http://127.0.0.1:8080"),
    ui.FormInputWidth(72),
    ui.FormInputForForm("loginForm"),
    ui.FormInputRequired(),
    ui.FormInputHelp("Use the Admin API base URL."),
    ui.FormInputValidators(ui.Required()),
)
```

## 常用选项

- `ui.FormInputPlaceholder(text)`
- `ui.FormInputWidth(width)`
- `ui.FormInputForForm(formID)`
- `ui.FormInputLayout(layout)`
- `ui.FormInputValidators(validators...)`
- `ui.FormInputHelp(text)`
- `ui.FormInputRequired()`

## 可用变体

- `ui.FormInputItem(...)`
- `ui.FormPasswordItem(...)`
- `ui.FormSearchItem(...)`

## Fiber-first 约束

- 只生成 FormItem 与 Input VNode。
- 字段变更通过 `intent.BindField(...)` 和可选 `intent.BindForm(...)` 声明。
- 校验器作为配置传入，不在组件包内处理业务提交或网络请求。

## 测试

```powershell
go test ./ui/components/forminputitem
```
