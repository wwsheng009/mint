# Ant Design 在 Mint TUI 中的实施指南

> 基于 `docs/theme/design/ant_design.md` 的实践指南

## 📋 目录

1. [核心设计映射](#核心设计映射)
2. [组件实施示例](#组件实施示例)
3. [表单布局最佳实践](#表单布局最佳实践)
4. [键盘交互实现](#键盘交互实现)
5. [完整 Demo 示例](#完整-demo-示例)

---

## 核心设计映射

### 🎨 Design Token 映射

Ant Design 的 Design Token 完美对应 Mint TUI 的语义色系统：

| Ant Design Token | Mint TUI 语义色 | RGB (Nord) | 用途 |
|-----------------|----------------|-----------|------|
| colorPrimary | `theme.Primary()` | 136, 192, 208 | 主按钮、链接 |
| colorSuccess | `theme.Success()` | 163, 190, 140 | 成功状态 |
| colorWarning | `theme.Warning()` | 235, 203, 139 | 警告提示 |
| colorError | `theme.Error()` | 191, 97, 106 | 错误状态 |
| colorText | `theme.Text()` | 236, 239, 244 | 主文字 |
| colorTextSecondary | `theme.Muted()` | 97, 110, 136 | 次级文字 |
| colorTextTertiary | `theme.Placeholder()` | 97, 110, 136 | 占位符 |
| borderColor | `theme.Border()` | 76, 86, 106 | 边框 |
| colorBgContainer | `theme.Surface()` | 59, 66, 82 | 容器背景 |
| colorBgLayout | `theme.BG()` | 46, 52, 64 | 页面背景 |

### 📏 尺寸系统映射

Ant Design 使用 `8px grid`，Mint TUI 使用 `1 cell`：

| Ant Design | Mint TUI | 说明 |
|-----------|---------|------|
| 8px | 1 cell | 基础间距单位 |
| 16px (paddingSM) | 2 cells | 小间距 |
| 24px (paddingMD) | 3 cells | 中等间距 |
| 32px (paddingLG) | 4 cells | 大间距 |

---

## 组件实施示例

### 1. Button 组件 - Ant Design 风格

Ant Design 的 Button 类型完全对应 Mint TUI：

```go
package main

import (
    "github.com/wwsheng009/mint/app"
    "github.com/wwsheng009/mint/ui"
)

// Ant Design 风格的按钮组
func AntStyleButtons() ui.VNode {
    return ui.HStack(
        // Primary Button - 对应 Ant 的 type="primary"
        app.ButtonBuilder("Primary").
            Variant(app.ButtonVariantPrimary).  // BG=PRIMARY, FG=BG
            OnClick(func() {
                // 主要操作
            }).
            Build(),

        // Default Button - 对应 Ant 的默认按钮
        app.ButtonBuilder("Default").
            Variant(app.ButtonVariantDefault).  // BG=SURFACE, FG=TEXT
            OnClick(func() {
                // 次要操作
            }).
            Build(),

        // Danger Button - 对应 Ant 的 danger 属性
        app.ButtonBuilder("Delete").
            Variant(app.ButtonVariantDanger).   // BG=ERROR, FG=BG
            OnClick(func() {
                // 危险操作
            }).
            Build(),

        // Success Button - 扩展类型
        app.ButtonBuilder("Confirm").
            Variant(app.ButtonVariantSuccess).  // BG=SUCCESS, FG=BG
            OnClick(func() {
                // 确认操作
            }).
            Build(),
    ).Gap(2).Build()  // 16px 间距 = 2 cells
}
```

**Ant Design 对应关系**：

| Ant Design | Mint TUI |
|-----------|---------|
| `<Button type="primary">` | `Variant(ButtonVariantPrimary)` |
| `<Button danger>` | `Variant(ButtonVariantDanger)` |
| `<Button disabled>` | `Disabled(true)` |

---

### 2. Form.Item - Ant Design 的表单结构

Ant Design 的 Form.Item 结构在 TUI 中的实现：

```go
// Ant Design 风格的表单项
func AntStyleFormItem() ui.VNode {
    username, setUsername := ui.UseStateString("")
    email, setEmail := ui.UseStateString("")

    return ui.VStack(
        // Form Item 1
        ui.VStackBuilder(
            // Label - 使用 TEXT 颜色
            app.NewTextBuilder("Username:").
                Style(style.Style{}.Foreground(theme.Text()).Bold(true)).
                Build(),

            // Control - Input with SURFACE 背景
            app.InputBuilder().
                Value(username).
                Placeholder("Enter username").
                Width(24).  // Ant 的 controlHeightLG
                OnChange(setUsername).
                Build(),

            // Help text - 使用 MUTED 颜色
            app.NewTextBuilder("Only letters and numbers").
                Style(style.Style{}.Foreground(theme.Muted())).
                Build(),
        ).Gap(1).Build(),  // 内部间距 1 cell

        // Form Item 2
        ui.VStackBuilder(
            app.NewTextBuilder("Email:").
                Style(style.Style{}.Foreground(theme.Text()).Bold(true)).
                Build(),

            app.InputBuilder().
                Value(email).
                Placeholder("Enter email").
                Width(24).
                OnChange(setEmail).
                Build(),
        ).Gap(1).Build(),
    ).Gap(2).Build()  // 表单项之间 2 cells 间距
}
```

**Ant Design 结构对应**：

```jsx
// Ant Design (React)
<Form.Item
  label="Username"
  help="Only letters and numbers"
>
  <Input placeholder="Enter username" />
</Form.Item>
```

```go
// Mint TUI (Go)
ui.VStack(
    Text("Username:").Bold(),           // Label
    Input().Placeholder("..."),          // Control
    Text("Only letters...").Muted(),     // Help
)
```

---

### 3. Table 组件 - Ant Design 数据表格

```go
// Ant Design 风格的表格
func AntStyleTable() ui.VNode {
    // 表头 - 使用 SURFACE 背景 + TEXT 加粗
    header := ui.HStackBuilder(
        app.NewTextBuilder("Name").
            Style(style.Style{}.
                Foreground(theme.Text()).
                Background(theme.Surface()).
                Bold(true)).
            Width(20).Build(),
        app.NewTextBuilder("Age").
            Style(style.Style{}.
                Foreground(theme.Text()).
                Background(theme.Surface()).
                Bold(true)).
            Width(10).Build(),
        app.NewTextBuilder("Address").
            Style(style.Style{}.
                Foreground(theme.Text()).
                Background(theme.Surface()).
                Bold(true)).
            Width(30).Build(),
    ).Gap(1).Build()

    // 数据行 - 使用 BG 背景
    row1 := ui.HStackBuilder(
        app.NewTextBuilder("John").
            Style(style.Style{}.
                Foreground(theme.Text()).
                Background(theme.BG())).
            Width(20).Build(),
        app.NewTextBuilder("32").
            Style(style.Style{}.
                Foreground(theme.Text()).
                Background(theme.BG())).
            Width(10).Build(),
        app.NewTextBuilder("New York").
            Style(style.Style{}.
                Foreground(theme.Text()).
                Background(theme.BG())).
            Width(30).Build(),
    ).Gap(1).Build()

    // Hover 行 - 使用 SURFACE 背景
    hoveredRow := ui.HStackBuilder(
        app.NewTextBuilder("Jane").
            Style(style.Style{}.
                Foreground(theme.Text()).
                Background(theme.Surface())).  // Hover 效果
            Width(20).Build(),
        // ... 其他列
    ).Gap(1).Build()

    return ui.VStack(
        header,
        ui.NewText("").Style(style.Style{}.Foreground(theme.Border())).Build(),  // 分割线
        row1,
        hoveredRow,
    ).Gap(0).Build()
}
```

**Ant Design 对应**：

| Ant Design | Mint TUI |
|-----------|---------|
| Table header BG | `theme.Surface()` |
| Row hover BG | `theme.Surface()` |
| Row selected BG | `theme.Select()` |
| Border color | `theme.Border()` |

---

### 4. Modal 组件 - Ant Design 弹窗

```go
// Ant Design 风格的确认对话框
func AntStyleConfirmModal(onConfirm, onCancel func()) ui.VNode {
    return ui.Modal(
        ui.Bordered().
            Style(string(theme.Warning())).  // Warning 边框
            Width(40).
            Child(
                ui.VStackBuilder(
                    ui.Text(""),
                    // 标题 - 居中
                    ui.HStackBuilder(
                        app.NewTextBuilder("*** Confirm Delete ***").
                            Style(style.Style{}.
                                Foreground(theme.Warning()).
                                Bold(true)).
                            Build(),
                    ).Align(ui.AlignCenter).Build(),
                    ui.Text(""),
                    // 内容
                    ui.HStackBuilder(
                        app.NewTextBuilder("Are you sure?").
                            Style(style.Style{}.Foreground(theme.Text())).
                            Build(),
                    ).Align(ui.AlignCenter).Build(),
                    ui.Text(""),
                    // Footer 按钮 - 居中
                    ui.HStackBuilder(
                        // Secondary 按钮 - 取消
                        app.ButtonBuilder("[ Cancel ]").
                            Variant(app.ButtonVariantSecondary).
                            OnClick(onCancel).
                            Build(),
                        ui.Text(" "),
                        // Primary 按钮 - 确认
                        app.ButtonBuilder("[ OK ]").
                            Variant(app.ButtonVariantPrimary).
                            OnClick(onConfirm).
                            Build(),
                    ).Align(ui.AlignCenter).Build(),
                    ui.Text(""),
                    // 提示文字
                    ui.HStackBuilder(
                        app.NewTextBuilder("Press ESC to close").
                            Style(style.Style{}.Foreground(theme.Placeholder())).
                            Build(),
                    ).Align(ui.AlignCenter).Build(),
                ).Build(),
            ).
            Build(),
    ).
    OnClose(onCancel).
    CloseOnESC(true).
    Build()
}
```

**Ant Design 对应**：

```jsx
// Ant Design
<Modal
  title="Confirm Delete"
  onOk={handleConfirm}
  onCancel={handleCancel}
>
  <p>Are you sure?</p>
</Modal>
```

---

## 表单布局最佳实践

### 标准表单布局（推荐）

Ant Design 的右对齐 Label 在 TUI 中最清晰：

```go
func StandardFormLayout() ui.VNode {
    // 计算最长的 Label 长度
    labelWidth := 12  // "Confirm Password: " 的长度

    return ui.VStack(
        // Form Item 1
        ui.HStackBuilder(
            // Label - 右对齐，固定宽度
            app.NewTextBuilder("Username:").
                Style(style.Style{}.Foreground(theme.Text()).Bold(true)).
                Width(labelWidth).
                Build(),
            // 间距
            ui.Text(" "),
            // Control
            app.InputBuilder().
                Placeholder("Enter username").
                Width(30).
                Build(),
        ).Build(),

        // Form Item 2
        ui.HStackBuilder(
            app.NewTextBuilder("Password:").
                Style(style.Style{}.Foreground(theme.Text()).Bold(true)).
                Width(labelWidth).
                Build(),
            ui.Text(" "),
            app.InputBuilder().
                InputType(form.InputTypePassword).
                Placeholder("Enter password").
                Width(30).
                Build(),
        ).Build(),

        // Form Item 3 - 带 Help text
        ui.VStackBuilder(
            ui.HStackBuilder(
                app.NewTextBuilder("Email:").
                    Style(style.Style{}.Foreground(theme.Text()).Bold(true)).
                    Width(labelWidth).
                    Build(),
                ui.Text(" "),
                app.InputBuilder().
                    Placeholder("Enter email").
                    Width(30).
                    Build(),
            ).Build(),
            // Help text - 与 Control 左对齐
            ui.HStackBuilder(
                ui.Text("").Width(labelWidth + 1).Build(),  // 缩进到 Control 位置
                app.NewTextBuilder("We'll never share your email").
                    Style(style.Style{}.Foreground(theme.Muted())).
                    Build(),
            ).Build(),
        ).Gap(1).Build(),
    ).Gap(2).Build()  // Form Item 之间 2 cells 间距
}
```

**效果**：

```
Username: [__________________]
Password: [__________________]
Email:    [__________________]
          We'll never share your email
```

### 错误状态显示

```go
func FormItemWithError() ui.VNode {
    email, setEmail := ui.UseStateString("")
    hasError, setHasError := ui.UseStateBool(false)

    var helpText ui.VNode
    if hasError {
        // Error 状态 - 红色 ERROR 颜色
        helpText = app.NewTextBuilder("Invalid email format").
            Style(style.Style{}.Foreground(theme.Error()))
    } else {
        // 正常 Help - MUTED 颜色
        helpText = app.NewTextBuilder("example@domain.com").
            Style(style.Style{}.Foreground(theme.Muted()))
    }

    return ui.VStackBuilder(
        ui.HStackBuilder(
            app.NewTextBuilder("Email:").
                Style(style.Style{}.Foreground(theme.Text()).Bold(true)).
                Width(10).Build(),
            ui.Text(" "),
            app.InputBuilder().
                Value(email).
                Width(30).
                OnChange(func(s string) {
                    setEmail(s)
                    // 简单验证
                    setHasError(!strings.Contains(s, "@"))
                }).
                Build(),
        ).Build(),
        // Help / Error 文本
        ui.HStackBuilder(
            ui.Text("").Width(11).Build(),  // 缩进
            helpText.Build(),
        ).Build(),
    ).Gap(1).Build()
}
```

---

## 键盘交互实现

### 焦点管理

Ant Design 的焦点系统在 Mint TUI 中的实现：

```go
func KeyboardNavigationDemo() ui.VNode {
    focusIndex, setFocusIndex := ui.UseStateInt(0)

    // 模拟多个可聚焦元素
    items := []struct{
        label string
        action func()
    }{
        {"Button 1", func() {}},
        {"Button 2", func() {}},
        {"Button 3", func() {}},
    }

    return ui.VStack(
        ui.Text("Use Tab to navigate, Enter to select"),
        ui.Text(""),
        // 渲染按钮列表
        ui.VStackFromSlice(
            items,
            func(i int, item struct{label string; action func()}) ui.VNode {
                isFocused := focusIndex == i
                return app.ButtonBuilder(item.label).
                    FocusStyle(app.FocusStyleBracket).
                    OnClick(item.action).
                    Build()
            },
        ),
    )
}
```

### 快捷键绑定

```go
// 全局快捷键 - 对应 Ant Design 的全局操作
func setupGlobalHotkeys(app *framework.App) {
    // Ctrl+C - 退出应用
    app.BindKey("Ctrl+c", func(ev *event.KeyEvent) bool {
        app.Quit()
        return true
    })

    // / - 聚焦搜索框
    app.BindKey("/", func(ev *event.KeyEvent) bool {
        // 聚焦到搜索输入框
        return true
    })

    // ? - 打开帮助
    app.BindKey("?", func(ev *event.KeyEvent) bool {
        // 打开帮助 Modal
        return true
    })

    // Esc - 关闭当前 Modal
    app.BindKey("Esc", func(ev *event.KeyEvent) bool {
        // 关闭最顶层的 Modal
        return true
    })
}
```

### Modal 焦点锁

```go
func ModalWithFocusTrap() ui.VNode {
    showModal, setShowModal := ui.UseStateBool(false)

    if showModal {
        return ui.Modal(
            ConfirmModal(
                func() { setShowModal(false) },
            ),
        ).
        // 焦点锁 - 只在 Modal 内循环
        OnClose(func() { setShowModal(false) }).
        CloseOnESC(true).
        Build()
    }

    return app.ButtonBuilder("Open Modal").
        OnClick(func() { setShowModal(true) }).
        Build()
}
```

---

## 完整 Demo 示例

### Ant Design 风格的用户注册表单

```go
package main

import (
    "github.com/wwsheng009/mint/app"
    "github.com/wwsheng009/mint/framework/theme"
    "github.com/wwsheng009/mint/runtime/style"
    "github.com/wwsheng009/mint/ui"
)

// AntDesignStyleForm - 完整的 Ant Design 风格表单
func AntDesignStyleForm() ui.VNode {
    username, setUsername := ui.UseStateString("")
    email, setEmail := ui.UseStateString("")
    password, setPassword := ui.UseStateString("")
    agree, setAgree := ui.UseStateBool(false)

    return ui.Bordered().
        Style(string(theme.Border())).
        Child(
            ui.VStackBuilder(
                // 表单标题 - 使用 PRIMARY 颜色
                ui.HStackBuilder(
                    app.NewTextBuilder("User Registration").
                        Style(style.Style{}.
                            Foreground(theme.Primary()).
                            Bold(true)).
                        Build(),
                ).Align(ui.AlignCenter).Build(),

                ui.Text(""),

                // 分组 1: Basic Info
                ui.HStackBuilder(
                    app.NewTextBuilder("[ Basic Info ]").
                        Style(style.Style{}.
                            Foreground(theme.Primary())).
                        Build(),
                ).Align(ui.AlignCenter).Build(),

                ui.Text(""),

                // Username Field
                formItem("Username:", "Enter username", 30, username, setUsername),

                // Email Field
                formItem("Email:", "Enter email", 30, email, setEmail),

                ui.Text(""),

                // 分组 2: Security
                ui.HStackBuilder(
                    app.NewTextBuilder("[ Security ]").
                        Style(style.Style{}.
                            Foreground(theme.Primary())).
                        Build(),
                ).Align(ui.AlignCenter).Build(),

                ui.Text(""),

                // Password Field
                formItemPassword("Password:", "Enter password", 30, password, setPassword),

                ui.Text(""),

                // Checkbox - 对应 Ant Design 的 Checkbox
                ui.HStackBuilder(
                    app.NewTextBuilder("").
                        Width(11).Build(),  // Label 占位
                    app.CheckboxBuilder().
                        Label("I agree to the terms and conditions").
                        Checked(agree).
                        OnChange(setAgree).
                        Build(),
                ).Build(),

                ui.Text(""),

                // Footer 按钮 - Primary + Secondary
                ui.HStackBuilder(
                    app.ButtonBuilder("[ Cancel ]").
                        Variant(app.ButtonVariantSecondary).
                        OnClick(func() {
                            // Reset form
                        }).
                        Build(),
                    ui.Text(" "),
                    app.ButtonBuilder("[ Register ]").
                        Variant(app.ButtonVariantPrimary).
                        Disabled(!agree).  // Ant Design: disabled state
                        OnClick(func() {
                            // Submit form
                        }).
                        Build(),
                ).Align(ui.AlignCenter).Build(),

                ui.Text(""),
            ).Gap(2).Build(),  // 统一间距
        ).Build()
}

// 辅助函数 - Form Item
func formItem(label, placeholder string, width int, value string, onChange func(string)) ui.VNode {
    return ui.VStackBuilder(
        ui.HStackBuilder(
            app.NewTextBuilder(label).
                Style(style.Style{}.Foreground(theme.Text()).Bold(true)).
                Width(10).Build(),  // 固定 Label 宽度
            ui.Text(" "),
            app.InputBuilder().
                Value(value).
                Placeholder(placeholder).
                Width(width).
                OnChange(onChange).
                Build(),
        ).Build(),
        ui.HStackBuilder(
            ui.Text("").Width(11).Build(),
            app.NewTextBuilder("").
                Style(style.Style{}.Foreground(theme.Muted())).
                Build(),
        ).Build(),
    ).Gap(1).Build()
}

// 辅助函数 - Password Form Item
func formItemPassword(label, placeholder string, width int, value string, onChange func(string)) ui.VNode {
    return ui.VStackBuilder(
        ui.HStackBuilder(
            app.NewTextBuilder(label).
                Style(style.Style{}.Foreground(theme.Text()).Bold(true)).
                Width(10).Build(),
            ui.Text(" "),
            app.InputBuilder().
                Value(value).
                InputType(form.InputTypePassword).
                Placeholder(placeholder).
                Width(width).
                OnChange(onChange).
                Build(),
        ).Build(),
    ).Gap(1).Build()
}
```

---

## Ant Design 组件实现检查清单

### ✅ 已实现

- [x] Button (Primary, Secondary, Danger, Success, Default, Disabled)
- [x] Input (Normal, Focus, Placeholder, Disabled, Error)
- [x] Checkbox (Normal, Checked, Focus, Disabled)
- [x] Select (Normal, Focus, Disabled)
- [x] Textarea (Normal, Focus, Placeholder, Disabled)
- [x] Modal (Overlay, FocusTrap, CloseOnESC)

### 🚧 待实现

- [ ] Radio (单选框)
- [ ] Switch (开关)
- [ ] Slider (滑块)
- [ ] DatePicker (日期选择)
- [ ] Upload (文件上传)
- [ ] Table (排序、选择)
- [ ] Tabs (标签页)
- [ ] Breadcrumb (面包屑)
- [ ] Progress (进度条)
- [ ] Spin (加载动画)

---

## 总结

### 关键要点

1. **语义色映射**: Ant Design 的 colorPrimary → theme.Primary()
2. **尺寸映射**: Ant Design 的 8px → 1 cell
3. **结构保留**: Form.Item 的 Label + Control + Help 结构
4. **交互保留**: Focus system, Tab navigation, Modal focus trap
5. **状态保留**: hover, focus, disabled, error 状态

### 实施建议

1. **优先级**: 先完成基础表单组件（Input, Button, Checkbox）
2. **布局**: 使用 Ant Design 的表单布局规则（右对齐 Label）
3. **键盘**: 实现完整的焦点管理和快捷键
4. **测试**: 参考 Ant Design 官方示例进行对比测试

### 参考资源

- Ant Design 官方文档: https://ant.design/
- 本项目设计规范: `docs/theme/design/ant_design.md`
- 主题系统指南: `docs/theme/theme_system_guide.md`
- 组件配色检查: `docs/theme/check_1.md`
