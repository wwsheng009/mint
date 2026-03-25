# Switch

开关组件，适合布尔状态切换和配置项启停。

## 已支持

- `checked` 受控状态
- 自定义开 / 关文案
- 键盘与鼠标切换
- `Field` / `Form` 绑定
- 组件级 toggle intent

## 示例

```go
ui.NewSwitchBuilder().
    Label("Wi-Fi").
    Checked(true).
    Labels("ON", "OFF").
    Build()
```

快捷创建也可直接用：

```go
ui.Switch("Bluetooth", true)
```
