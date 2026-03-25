# Validation

表单校验工具包，给 `Form` 和字段绑定提供可组合的验证器。

## 已支持

- `Required`
- `MinLength` / `MaxLength` / `Length`
- `Min` / `Max` / `Range`
- `Pattern`
- `Email`
- `URL`
- `OneOf`
- `Custom`
- `NewAllValidator` / `NewAnyValidator`
- `Chain()` fluent API

## 示例

```go
validator := validation.NewChain().
    Required().
    MinLength(6).
    Pattern(`^[a-zA-Z0-9_]+$`).
    Build()

if err := validator.Validate("mint_user"); err != nil {
    // handle validation error
}
```

需要组合多条规则时，优先用 `Chain()` 或 `NewAllValidator(...)`；需要“满足其一”时，用 `NewAnyValidator(...)`。
