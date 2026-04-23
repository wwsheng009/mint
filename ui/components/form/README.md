# Form

表单容器组件，负责字段值、校验、提交/重置和字段级元状态的统一管理。

## 已支持

- `FormItem` 包装层
- `vertical` / `horizontal` / `inline` 布局
- 初始 `values`
- `submit` / `reset` intent
- 字段级 `touched` / `dirty` / `submitted`
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
        ).Label("Username").Build(),
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
