# Transfer

穿梭框组件，适合成员分配、权限选择和双列表迁移场景。

## 已支持

- 双列表迁移
- `targetKeys` 受控 / 非受控模式
- 禁用项过滤
- 自定义标题与操作文案
- `Field` / `Form` 绑定
- 组件级 `ChangeIntent`

## 示例

```go
ui.NewTransferBuilder().
    ComponentID("members-transfer").
    Titles("Available", "Chosen").
    Operations(">>", "<<").
    Items([]ui.TransferItem{
        ui.NewTransferItem("a", "Alpha"),
        ui.NewTransferItem("b", "Beta"),
        ui.NewTransferItem("c", "Gamma"),
    }).
    InitialTargetKeys([]string{"b"}).
    Build()
```

绑定到字段时，当前目标列表 key 会通过 `FieldChangeIntent` 一起发出。
