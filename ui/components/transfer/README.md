# Transfer

穿梭框组件，适合成员分配、权限选择和双列表迁移场景。

## 已支持

- 双列表迁移
- 源列表 / 目标列表内置搜索过滤
- 源列表 / 目标列表分页
- 当前过滤结果的批量迁移按钮
- `targetKeys` 受控 / 非受控模式
- 禁用项过滤
- 带描述/禁用态的运维 item 预设
- 自定义标题与操作文案
- `Field` / `Form` 绑定
- 组件级 `ChangeIntent`
- 运维分配预设：默认开启搜索、分页、当前可见项批量移动，适合权限、scope、provider/key 和通知对象选择

## 示例

```go
ui.NewTransferBuilder().
    ComponentID("members-transfer").
    Titles("Available", "Chosen").
    Operations(">>", "<<").
    BulkOperations(true).
    BulkOperationLabels("All >>", "<< All").
    Searchable(true).
    PageSize(20).
    SearchPlaceholders("Find available", "Find chosen").
    Items([]ui.TransferItem{
        ui.NewTransferItem("a", "Alpha"),
        ui.NewTransferItem("b", "Beta"),
        ui.NewTransferItem("c", "Gamma"),
    }).
    InitialTargetKeys([]string{"b"}).
    Build()
```

运维分配预设：

```go
transfer.OperationalAssignment("provider-scope", []transfer.Item{
    transfer.NewItemWithDescription("group-default", "default", "Primary traffic group"),
    transfer.DisabledItem("group-archive", "archive", "Read-only archived group"),
}, []string{"group-default"}).
    Titles("Available groups", "Selected groups").
    Build()
```

通过 `ui.TransferOperationalAssignment(...)` 可以直接从顶层 SDK 使用同一预设。默认行为：

- 源/目标标题：`Available` / `Selected`
- 操作文案：`Add` / `Remove`
- 批量文案：`Add visible` / `Remove visible`
- 搜索占位：`Search available` / `Search selected`
- 每页数量：`20`
- 列宽/高度：`28` / `8`

绑定到字段时，当前目标列表 key 会通过 `FieldChangeIntent` 一起发出。

搜索是组件内部状态，适合大多数本地过滤场景；需要外部状态控制时可使用 `SearchValues(source, target)`，需要非受控初始值时可使用 `InitialSearchValues(source, target)`。过滤匹配 `key`、`title` 和 `description`，并按空格拆分为多个必须命中的关键词。

`PageSize(n)` 可为两侧列表启用本地分页，搜索后页码会回到第一页。启用分页后，标题显示当前页范围，例如 `Backlog (1-20/86)`，列表下方显示 Prev / Page x/y / Next。

批量按钮通过 `BulkOperations(true)` 开启，默认文案为 `All >` / `< All`，可通过 `BulkOperationLabels(toTarget, toSource)` 覆盖。批量迁移只作用于当前可见项，并跳过禁用项；未启用分页时“当前可见项”等于当前过滤结果，启用分页后等于当前页过滤结果。

## 验证

- Unit：`go test ./ui/components/transfer -count=1 -p 1`
- SDK shortcut：`go test ./ui -run Transfer -count=1 -p 1`
- E2E：`go test ./ui/e2e -run "^TestE2ETransfer" -count=1 -p 1`
