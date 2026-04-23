# Cascader

TUI 级联选择组件，适合省市区、分类树和多级菜单式选择。

## 已支持

- 层级 option 数据
- 列式展开
- 键盘与鼠标导航
- 叶子节点提交
- `changeOnSelect`
- `value` 受控 / `defaultValue` 非受控模式
- `Field` / `Form` 绑定
- 组件级 `ChangeIntent`

## 示例

```go
ui.NewCascaderBuilder().
    SetID("location").
    Placeholder("Select location").
    Options([]ui.CascaderOption{
        ui.NewCascaderOption("zj", "Zhejiang",
            ui.NewCascaderOption("hz", "Hangzhou"),
            ui.NewCascaderOption("nb", "Ningbo"),
        ),
        ui.NewCascaderOption("js", "Jiangsu",
            ui.NewCascaderOption("nj", "Nanjing"),
        ),
    }).
    DefaultValue("zj", "hz").
    Build()
```

如果需要在中间层级就提交，可以打开 `ChangeOnSelect(true)`。
