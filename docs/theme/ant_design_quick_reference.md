# Ant Design 快速参考指南 - Mint TUI

> 快速查找 Ant Design 组件在 Mint TUI 中的实现方法

## 🎨 颜色映射速查表

| Ant Design | Mint TUI | RGB (Nord) | 使用场景 |
|-----------|---------|-----------|---------|
| `colorPrimary` | `theme.Primary()` | 136,192,208 | 主按钮、链接、激活状态 |
| `colorSuccess` | `theme.Success()` | 163,190,140 | 成功提示、确认按钮 |
| `colorWarning` | `theme.Warning()` | 235,203,139 | 警告提示、注意 |
| `colorError` | `theme.Error()` | 191,97,106 | 错误提示、危险按钮 |
| `colorText` | `theme.Text()` | 236,239,244 | 主要文字、标题 |
| `colorTextSecondary` | `theme.Muted()` | 97,110,136 | 次要文字、说明 |
| `colorTextTertiary` | `theme.Placeholder()` | 97,110,136 | 占位符、帮助文本 |
| `colorBorder` | `theme.Border()` | 76,86,106 | 边框、分割线 |
| `colorBgContainer` | `theme.Surface()` | 59,66,82 | 容器背景、输入框 |
| `colorBgLayout` | `theme.BG()` | 46,52,64 | 页面背景 |

---

## 🧩 组件实现速查

### Button 按钮

```go
// Ant Design: <Button type="primary">Primary</Button>
app.ButtonBuilder("Primary").
    Variant(app.ButtonVariantPrimary).  // BG=PRIMARY, FG=BG
    OnClick(func() { /* action */ }).
    Build()

// Ant Design: <Button danger>Delete</Button>
app.ButtonBuilder("Delete").
    Variant(app.ButtonVariantDanger).  // BG=ERROR, FG=BG
    OnClick(func() { /* action */ }).
    Build()

// Ant Design: <Button disabled>Disabled</Button>
app.ButtonBuilder("Disabled").
    Disabled(true).  // FG=DISABLED_FG, BG=DISABLED_BG
    OnClick(func() { /* action */ }).
    Build()
```

### Input 输入框

```go
// Ant Design: <Input placeholder="Enter text" />
app.InputBuilder().
    Placeholder("Enter text").
    Width(24).
    OnChange(func(s string) { /* handle */ }).
    Build()

// Ant Design: <Input.Password />
app.InputBuilder().
    InputType(form.InputTypePassword).
    Placeholder("Password").
    Build()

// Focus 状态自动应用 FOCUS 颜色
```

### Form.Item 表单项

```go
// Ant Design 结构：
// <Form.Item label="Username" help="Enter username">
//   <Input />
// </Form.Item>

// Mint TUI 实现：
ui.VStack(
    // Label
    app.NewTextBuilder("Username:").
        Style(style.Style{}.Foreground(theme.Text()).Bold(true)).
        Build(),
    // Control
    app.InputBuilder().Placeholder("Enter username").Build(),
    // Help text
    app.NewTextBuilder("Enter username").
        Style(style.Style{}.Foreground(theme.Muted())).Build(),
)
```

### Modal 模态框

```go
// Ant Design: <Modal title="Confirm" onOk={ok} onCancel={cancel}>
ui.Modal(
    ui.Bordered().
        Style(string(theme.Warning())).  // Warning 边框
        Child(content).
        Build(),
).
OnClose(cancel).
CloseOnESC(true).
Build()
```

---

## 📐 尺寸转换

| Ant Design | Mint TUI | 说明 |
|-----------|---------|------|
| 8px | 1 cell | 基础间距单位 |
| 16px | 2 cells | 小间距（Gap） |
| 24px | 3 cells | 中等间距 |
| 32px | 4 cells | 大间距（Section） |
| height: 32px | height: 1 | 控件高度 |

---

## 🎯 表单布局最佳实践

### 标准布局（推荐）

```
LabelRightAlign: [Control__________]
                 Help text here
```

```go
// Label 宽度统一
labelWidth := 10

ui.VStack(
    // Label 行
    ui.HStack(
        Text("Username:").Width(labelWidth).Bold(),  // 右对齐 Label
        Text(" "),  // 间距
        Input().Width(30),  // Control
    ),
    // Help 行
    ui.HStack(
        Text("").Width(labelWidth + 1),  // 缩进到 Control 位置
        Text("Help text").Muted(),  // Help
    ),
)
```

### 窄屏布局

```
Label
[Control__________]
Help text
```

```go
ui.VStack(
    Text("Username:").Bold(),
    Input().Width(30),
    Text("Help text").Muted(),
)
```

---

## ⌨️ 键盘交互

### Tab 导航

```go
// 自动支持 Tab/Shift+Tab 导行
// 组件需实现 FocusableVNode 接口
```

### 快捷键

```go
// 全局快捷键
app.BindKey("Ctrl+c", func(ev *event.KeyEvent) bool {
    app.Quit()
    return true
})

app.BindKey("/", func(ev *event.KeyEvent) bool {
    // 聚焦搜索框
    return true
})

app.BindKey("Esc", func(ev *event.KeyEvent) bool {
    // 关闭 Modal
    return true
})
```

### Modal 焦点锁

```go
ui.Modal(content).
    CloseOnESC(true).  // Esc 关闭
    Build()
// 焦点自动锁定在 Modal 内
```

---

## 🎨 状态颜色

### Normal 状态

```go
style.Style{}.
    Foreground(theme.Text()).    // 文字
    Background(theme.Surface())  // 背景
```

### Focus 状态

```go
style.Style{}.
    Foreground(theme.Focus()).  // 焦点色
    Bold(true)                   // 加粗
```

### Error 状态

```go
style.Style{}.
    Foreground(theme.Error()).  // 错误色
    Underline(true)             // 下划线
```

### Disabled 状态

```go
style.Style{}.
    Foreground(theme.DisabledFG()).  // 禁用文字
    Background(theme.DisabledBG())   // 禁用背景
```

---

## 🚀 快速模板

### 表单页面模板

```go
func FormPage() ui.VNode {
    data, setData := ui.UseStatePtr(&FormData{})

    return ui.VStack(
        Header(),
        ui.Text(""),
        ui.Bordered().
            Style(string(theme.Border())).
            Child(
                ui.VStack(
                    FormItem("Name:", "Enter name", data.Name, setData),
                    FormItem("Email:", "Enter email", data.Email, setData),
                ).Gap(1),
            ).
            Build(),
        ActionButtons(),
    ).Gap(2).Build()
}
```

### Modal 模板

```go
func ConfirmModal(onOk, onCancel func()) ui.VNode {
    return ui.Modal(
        ui.Bordered().
            Style(string(theme.Warning())).
            Width(40).
            Child(
                ui.VStack(
                    ui.Text(""),
                    Title("Are you sure?").Center().Primary(),
                    ui.Text(""),
                    Buttons(onOk, onCancel).Center(),
                ).Gap(1),
            ).
            Build().
    ).
    OnClose(onCancel).
    CloseOnESC(true).
    Build()
}
```

### Table 模板

```go
func Table() ui.VNode {
    return ui.VStack(
        // Header
        TableRow(
            TableCell("Name", 20).Surface().Bold(),
            TableCell("Age", 10).Surface().Bold(),
            TableCell("Address", 30).Surface().Bold(),
        ),
        Divider(),
        // Row 1
        TableRow(
            TableCell("John", 20),
            TableCell("32", 10),
            TableCell("NYC", 30),
        ),
        // Row 2 (Hover)
        TableRow(
            TableCell("Jane", 20).Surface(),  // Hover effect
            TableCell("28", 10).Surface(),
            TableCell("LA", 30).Surface(),
        ),
    )
}
```

---

## 📊 组件实现状态

### ✅ 已完全实现

- [x] Button (所有 variants)
- [x] Input (包括 Password)
- [x] Checkbox (所有状态)
- [x] Select (下拉选择)
- [x] Textarea (多行输入)
- [x] Modal (焦点锁 + Esc 关闭)
- [x] Form.Item (Label + Control + Help)

### 🚧 部分实现

- [~] Table (基础结构)
- [~] Tabs (基础切换)

### ❌ 待实现

- [ ] Radio (单选框)
- [ ] Switch (开关)
- [ ] Slider (滑块)
- [ ] DatePicker (日期选择)
- [ ] Upload (文件上传)
- [ ] Tree (树形结构)
- [ ] Breadcrumb (面包屑)
- [ ] Progress (进度条)

---

## 💡 设计原则

### 1. 颜色语义化

✅ 正确：
```go
Foreground(theme.Primary())  // 语义化
```

❌ 错误：
```go
Foreground("#3cb371")  // 硬编码颜色
```

### 2. 状态优先级

```
Disabled > Error > Warning > Success > Focus > Hover > Normal
```

### 3. 间距统一

```
使用 Gap(1), Gap(2), Gap(4)  // 8px, 16px, 32px
而非 Gap(3), Gap(7)  // 奇数间距
```

### 4. 一致性

所有相同状态的组件使用相同的颜色：
- 所有 Focus 状态都用 `theme.Focus()`
- 所有 Error 状态都用 `theme.Error()`
- 所有 Disabled 状态都用 `theme.DisabledFG/BG`

---

## 🔗 相关文档

- [Ant Design 官方文档](https://ant.design/)
- [Ant Design 设计理念](../design/ant_design.md)
- [主题系统指南](../theme_system_guide.md)
- [组件配色检查](../component_color_compliance.md)
- [实施完整指南](../ant_design_implementation_guide.md)

---

**最后更新**: 2026-02-07
**主题版本**: Nord (默认)
