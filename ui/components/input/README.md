# Input

单行输入组件，适合文本、密码、数字和搜索场景。

## 已支持

- `text` / `password` / `number` / `email`
- `placeholder`
- `maxLen`
- `readOnly` / `disabled`
- `prefix` / `suffix`
- `addonBefore` / `addonAfter`
- `Search()` 变体
- number 模式下负号 / 小数控制
- number 模式下 `min` / `max`
- number 模式下上下方向键 `step`
- `Field` / `Form` 绑定

## 示例

```go
ui.NewInputBuilder().
    Placeholder("Search users").
    Search().
    Prefix("@").
    AddonAfter("/profile").
    Build()
```

```go
ui.NewInputBuilder().
    Type(ui.InputNumber).
    Min(0).
    Max(100).
    Step(5).
    Value("20").
    Build()
```

快捷函数：

```go
ui.Input("Username")
ui.SearchInput("Search")
```
