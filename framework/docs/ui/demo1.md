好嘞，这里给你一份 **“全功能展示型 Demo”**，专门用来验证你这套 TUI 引擎的架构完整性。这个程序会同时覆盖：

✅ 声明式组件
✅ 状态系统
✅ 布局系统
✅ Modal
✅ Layer
✅ Input
✅ Focus
✅ Scroll
✅ 虚拟列表
✅ 动画
✅ 事件处理

相当于你这个 UI Runtime 的「集成验收用例」。

下面是 **示例 Demo App（Go 风格伪代码，接近真实实现）**：

---

## 🚀 App 入口

```go
func main() {
    ui.Run(App)
}
```

---

## 🧠 根组件

```go
func App() ui.Node {
    count, setCount := ui.UseState(0)
    showModal, setShowModal := ui.UseState(false)
    input, setInput := ui.UseState("")
    items := make([]string, 10000)

    for i := range items {
        items[i] = fmt.Sprintf("Log line #%d", i)
    }

    return ui.Screen(
        ui.Column(
            Header(count, setShowModal),
            ui.Row(
                Sidebar(setCount),
                ContentArea(input, setInput, items),
            ),
        ),

        // Layer: Modal
        ui.If(showModal, func() ui.Node {
            return ConfirmModal(func() {
                setShowModal(false)
            })
        }),
    )
}
```

---

## 🧩 Header（状态 + 布局）

```go
func Header(count int, openModal func(bool)) ui.Node {
    return ui.Box().
        Height(3).
        Bg(ui.ColorBlue).
        Padding(1).
        Child(
            ui.Row(
                ui.Text("🚀 TUI Engine Demo").Bold(true),
                ui.Spacer(),
                ui.Button("Open Modal", func() {
                    openModal(true)
                }),
                ui.Text(fmt.Sprintf("Clicks: %d", count)),
            ),
        )
}
```

---

## 📚 Sidebar（事件）

```go
func Sidebar(setCount func(int)) ui.Node {
    return ui.Box().
        Width(20).
        Border(ui.BorderRounded).
        Padding(1).
        Child(
            ui.Column(
                ui.Text("Menu").Bold(true),
                ui.Button("Add Count", func() {
                    setCount(func(v int) int { return v + 1 })
                }),
                ui.Button("Quit", ui.Quit),
            ),
        )
}
```

---

## 🧠 内容区（Input + Scroll + VirtualList）

```go
func ContentArea(value string, setValue func(string), items []string) ui.Node {
    return ui.Column(
        ui.Box().
            Height(3).
            Border(ui.BorderRounded).
            Padding(1).
            Child(
                ui.Input().
                    Value(value).
                    Placeholder("Type something...").
                    OnChange(setValue),
            ),

        ui.Box().
            Flex(1).
            Border(ui.BorderRounded).
            Child(
                ui.VirtualList(
                    len(items),
                    1,
                    func(i int) ui.Node {
                        return ui.Text(items[i])
                    },
                ),
            ),
    )
}
```

---

## 🪟 Modal（Layer + 动画 + Focus Trap）

```go
func ConfirmModal(onClose func()) ui.Node {
    return ui.Modal(func() ui.Node {
        return ui.Box().
            Width(40).
            Height(8).
            Border(ui.BorderDouble).
            Padding(1).
            Animate(ui.FadeIn | ui.ScaleIn).
            Child(
                ui.Column(
                    ui.Text("Are you sure?").Bold(true),
                    ui.Spacer(),
                    ui.Row(
                        ui.Spacer(),
                        ui.Button("Cancel", onClose),
                        ui.Button("OK", onClose),
                    ),
                ),
            )
    }).CloseOnEsc(true)
}
```

---

## 🎯 这个 Demo 实际验证了什么？

| 功能      | 在哪里体现               |
| ------- | ------------------- |
| 声明式 UI  | 所有组件写法              |
| 状态 Hook | UseState            |
| 局部刷新    | count / input       |
| 布局系统    | Column / Row / Flex |
| Layer   | Modal               |
| 动画      | Animate()           |
| Focus   | Input + Modal       |
| 事件系统    | Button              |
| Scroll  | VirtualList 容器      |
| 虚拟化     | 10000 行日志           |
| Diff    | 状态更新时只刷新局部          |

---

## 🧠 运行后界面结构大概：

```
+--------------------------------------------------+
| 🚀 TUI Engine Demo        [Open Modal] Clicks: 2 |
+-----------+--------------------------------------+
| Menu      | [ Input box....................... ] |
| Add Count |--------------------------------------|
| Quit      | Log line #0                          |
|           | Log line #1                          |
|           | ...                                  |
|           | (scroll)                             |
+-----------+--------------------------------------+
```

弹窗出现时：

```
      ┌──────────────────────────┐
      │      Are you sure?       │
      │                          │
      │        [Cancel] [OK]     │
      └──────────────────────────┘
```

---

## 🔥 这是一个“引擎完整度验证 Demo”

如果这个 Demo：

* 不闪屏
* 滚动流畅
* 输入不卡
* Modal 正确拦截事件
* 光标始终准确

那么你的架构就是 **生产级 TUI Runtime 设计**。

---

如果你愿意，下一步我可以给你：

> 🧠 **内部 Runtime 调度流程图（从 setState → Buffer 输出全过程）**

那是引擎核心运作的“总装图”。
