# Transfer

穿梭框组件，适合成员分配、权限选择和双列表迁移场景。

## 已支持

- 双列表迁移
- 源列表 / 目标列表内置搜索过滤
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
    Searchable(true).
    SearchPlaceholders("Find available", "Find chosen").
    Items([]ui.TransferItem{
        ui.NewTransferItem("a", "Alpha"),
        ui.NewTransferItem("b", "Beta"),
        ui.NewTransferItem("c", "Gamma"),
    }).
    InitialTargetKeys([]string{"b"}).
    Build()
```

绑定到字段时，当前目标列表 key 会通过 `FieldChangeIntent` 一起发出。

搜索是组件内部状态，适合大多数本地过滤场景；需要外部状态控制时可使用 `SearchValues(source, target)`，需要非受控初始值时可使用 `InitialSearchValues(source, target)`。过滤匹配 `key`、`title` 和 `description`，并按空格拆分为多个必须命中的关键词。

## 验证

- Unit：`go test ./ui/components/transfer -count=1 -p 1`
- SDK shortcut：`go test ./ui -run Transfer -count=1 -p 1`
- E2E：`go test ./ui/e2e -run "^TestE2ETransfer" -count=1 -p 1`
