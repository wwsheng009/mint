# Collapse

`collapse/` 提供折叠面板组件，支持：

- 多面板展开 / 收起
- `accordion` 手风琴模式
- 受控 `ActiveKeys(...)` 与非受控 `InitialActiveKeys(...)`
- 面板级 `disabled`
- 内置 `CollapseChangeIntent` 与可选 `OnChange(...)` / field 绑定

## 示例

```go
ui.NewCollapseBuilder().
    ComponentID("settings.collapse").
    Item(
        collapse.Section("General", ui.Text("General settings")).
            WithKey("general"),
    ).
    Item(
        collapse.Section("Advanced", ui.Text("Advanced settings")).
            WithKey("advanced"),
    ).
    InitialActiveKeys("general").
    Build()
```

## Intent

- header 点击会发出 `CollapseToggleIntent`
- 状态变化后组件会发出 `CollapseChangeIntent`
- 若配置了 `ActiveKeysForField(...)`，会同步发出逗号分隔的 `FieldChangeIntent`
