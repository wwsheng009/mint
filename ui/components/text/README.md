# Text

基础文本组件，适合普通文案、标题、标签和轻量样式化输出。

## 已支持

- 内容设置
- 前景 / 背景色
- `bold` / `italic` / `underline`
- padding
- 文本对齐
- `maxWidth`

## 示例

```go
ui.NewTextBuilder("Hello, Mint").
    Bold(true).
    FgColor("green").
    Padding(0, 1, 0, 1).
    Build()
```

快捷函数：

```go
ui.Text("Hello")
```
