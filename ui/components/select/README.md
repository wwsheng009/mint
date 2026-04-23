# Select

选择器组件，适合单选、多选、`tags` 输入和分组选项场景。

## 已支持

- 单选 / 多选 / `tags`
- `OptGroup`
- `filterOption` 与过滤占位文案
- `placeholder`、`disabled`
- overlay popup、portal root、outside-click 收口
- `OnChange`、`ForField`、`ForForm`

## 状态语义

- `Selected(...)` 用于声明当前单选索引，`SelectedIndices(...)` 用于声明多选 / `tags` 当前值。
- `TagsMode(true)` 会切到 `SelectionTags` 语义；关闭后默认回到多选语义。
- 需要浮层渲染时使用 `OverlayPopup(true)`；如果同时开启 `CloseOnOutside(true)`，应用启动时应调用一次 `select.Install(app)`，让 outside-click 关闭链路接入中间件。
- `ForField(...)` 在单选模式下写回选中索引，在多选 / `tags` 模式下写回逗号拼接后的索引串；`ForForm(...)` 用于接到表单字段同步链路。

## 示例

```go
ui.NewSelectBuilder().
    SetID("city-select").
    Options([]ui.SelectOption{
        ui.NewSelectOption("beijing", "Beijing"),
        ui.NewSelectOption("shanghai", "Shanghai"),
    }).
    Placeholder("Select city").
    FilterOption(true).
    OverlayPopup(true).
    CloseOnOutside(true).
    Build()
```

需要标签模式时可直接打开 `TagsMode(true)`。

## 安装方式

如果使用 overlay popup 且希望 outside-click 自动关闭，应用启动时调用一次：

```go
selectcomp.Install(app)
```

## 测试入口

- 单测：`go test ./ui/components/select`
- 重点覆盖：`select_test.go` 中的选中态、多选字段同步、portal / anchor 定位与 popup 提交路径
- E2E：`go test ./ui/e2e -run TestE2ESelect`
