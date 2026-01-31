好，正式从“宏观体系”落地到：

> 🛠 **SDK 核心 API 设计草案 v1**

目标不是“全”，而是：

> **让开发者能写 80% UI，而你还能保留底层演进空间**

设计原则：

* 声明式
* 极简
* 不暴露引擎概念
* 与 Web 认知对齐（降低学习成本）

---

# 🧱 一、应用入口

必须做到 3 行启动。

```go
func main() {
    ui.Run(App)
}

func App() {
    ui.View(func() {
        ui.Text("Hello TUI")
    })
}
```

---

# 🧩 二、组件声明模型

### 所有 UI 都在 `ui.View` 作用域中

```go
ui.View(func() {
    ui.Row(
        ui.Text("CPU"),
        ui.Progress(0.7),
    )
})
```

内部自动构建 VNode 树。

---

# 🧠 三、Hooks 系统

### `useState`

```go
count, setCount := ui.UseState(0)

ui.Button("Add", func() {
    setCount(count + 1)
})
```

---

### `useEffect`

```go
ui.UseEffect(func() {
    ticker := time.NewTicker(time.Second)
    go func() {
        for range ticker.C {
            setCount(c => c+1)
        }
    }()
    return func() { ticker.Stop() }
}, [])
```

---

### `useMemo`

```go
value := ui.UseMemo(func() int {
    return heavyCompute(data)
}, []any{data})
```

---

# 📐 四、布局 API（核心）

采用 Flexbox 思维（但更简单）。

---

## 行 / 列

```go
ui.Row(
    ui.Text("Left"),
    ui.Text("Right"),
)

ui.Column(
    ui.Text("Top"),
    ui.Text("Bottom"),
)
```

---

## Flex 控制

```go
ui.Row(
    ui.Box().Grow(1).Child(ui.Text("Main")),
    ui.Box().Width(20).Child(ui.Text("Side")),
)
```

---

## Box 模型

```go
ui.Box().
    Padding(1).
    Margin(1).
    Border(ui.BorderRounded).
    Background(ui.ColorBlue).
    Child(ui.Text("Content"))
```

---

# 🎨 五、样式系统

轻量化，不搞 CSS 字符串。

```go
ui.Text("Error").
    Color(ui.Red).
    Bold().
    Bg(ui.DarkGray)
```

---

# 🎞 六、动画 API（高层封装）

```go
progress := ui.UseSpring(0.0)

ui.Button("Start", func(){
    progress.Set(1.0)
})

ui.Progress(progress.Value())
```

底层走动画调度，不进 Diff。

---

# 🖱 七、事件系统

```go
ui.Box().
    OnClick(func(){ ... }).
    OnKey("enter", func(){ ... })
```

支持：

* Click
* Hover
* Key
* Focus

---

# 🧭 八、列表（必须内置虚拟）

```go
ui.List(items, func(item Item) {
    ui.Row(
        ui.Text(item.Name),
        ui.Text(item.Status),
    )
})
```

开发者不需要知道虚拟滚动存在。

---

# 🪟 九、弹层系统

```go
ui.Modal(func(){
    ui.Text("Confirm?")
    ui.Button("OK", close)
})
```

自动进入顶层 Layer。

---

# 🔧 十、调试 API（开发期）

```go
ui.DebugLayout(true)
ui.DebugRepaint(true)
ui.ShowFPS(true)
```

---

# 🧠 十一、App 级 API

```go
ui.Theme(myTheme)
ui.SetTitle("Server Monitor")
ui.Quit()
```

---

# 📦 十二、组件写法（供生态使用）

```go
func Card(title string, body ui.Node) {
    ui.Box().
        Border(ui.BorderRounded).
        Padding(1).
        Child(
            ui.Column(
                ui.Text(title).Bold(),
                body,
            ),
        )
}
```

---

# 🚀 开发者实际体验应是：

```go
func App() {
    count, setCount := ui.UseState(0)

    ui.View(func() {
        ui.Column(
            ui.Text("Counter"),
            ui.Text(fmt.Sprint(count)),
            ui.Button("Add", func() { setCount(count+1) }),
        )
    })
}
```

没有 Fiber
没有 Diff
没有 Buffer

但底层全在运作。

---

# 🎯 这一版 API 覆盖了：

| 能力     | 是否支持 |
| ------ | ---- |
| 声明式 UI | ✅    |
| 状态系统   | ✅    |
| 副作用    | ✅    |
| 布局     | ✅    |
| 样式     | ✅    |
| 动画     | ✅    |
| 列表     | ✅    |
| 事件     | ✅    |
| 弹层     | ✅    |

已经能支撑真实应用。

---

下一步如果继续，就进入：

> 🧩 **底层如何把这些 API 映射到 RNode / Fiber / Buffer（SDK → Engine 桥接层）**

这是“SDK 表面优雅，底层仍高性能”的关键实现层。
