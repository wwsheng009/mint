# Mint UI 迁移指南

**版本**: v1.0
**日期**: 2026-01-31
**适用版本**: v0.x → v1.0

---

## 目录

1. [概述](#一概述)
2. [迁移前准备](#二迁移前准备)
3. [组件迁移](#三组件迁移)
4. [状态管理迁移](#四状态管理迁移)
5. [事件处理迁移](#五事件处理迁移)
6. [样式迁移](#六样式迁移)
7. [布局迁移](#七布局迁移)
8. [常见问题](#八常见问题)
9. [检查清单](#九检查清单)

---

## 一、概述

### 1.1 迁移目标

将现有 Mint UI 应用从**命令式架构**迁移到**声明式架构**。

### 1.2 主要变化

| 方面 | 命令式 (旧) | 声明式 (新) |
|------|------------|------------|
| 组件创建 | `NewXXX()` | `ui.XXX()` |
| 属性设置 | `SetXXX()` | 链式调用 |
| 状态管理 | `StateHolder` | `useState` |
| 事件处理 | `SetOnClick()` | `.OnClick()` |
| 布局 | `container.Add()` | `ui.HStack/VStack` |

### 1.3 兼容性策略

- **共存期**：新旧 API 可在过渡期内共存
- **标记废弃**：旧 API 标记为 `_legacy`
- **适配器**：提供适配器保持兼容
- **逐步迁移**：建议按组件逐步迁移

---

## 二、迁移前准备

### 2.1 评估迁移范围

```bash
# 检查使用的组件
grep -r "display\." .       # 显示组件
grep -r "input\." .        # 输入组件
grep -r "interactive\." .  # 交互组件
grep -r "layout\." .       # 布局组件

# 统计代码量
find . -name "*.go" | xargs wc -l
```

### 2.2 备份现有代码

```bash
# 创建备份分支
git checkout -b backup-before-migration

# 或者打标签
git tag pre-migration
```

### 2.3 更新依赖

```bash
# 更新到最新版本
go get -u github.com/wwsheng009/mint@latest

go mod tidy
```

---

## 三、组件迁移

### 3.1 Text 组件

#### 旧 API

```go
import "github.com/wwsheng009/mint/framework/display"

text := display.NewText("Hello, World!")
text.SetFgColor(color.Red)
text.SetBold(true)
text.SetAlign(display.AlignCenter)
container.Add(text)
```

#### 新 API

```go
import "github.com/wwsheng009/mint/ui"

text := ui.Text("Hello, World!").
    FgColor(color.Red).
    Bold(true).
    Align(ui.AlignCenter)

// 或在布局中直接使用
ui.VStack(
    ui.Text("Hello, World!").
        FgColor(color.Red).
        Bold(true).
        Align(ui.AlignCenter),
)
```

#### 迁移步骤

1. 将 `display.NewText()` 替换为 `ui.Text()`
2. 将 `SetXXX()` 调用改为链式调用
3. 删除 `container.Add(text)`，直接在布局中使用

---

### 3.2 Button 组件

#### 旧 API

```go
import "github.com/wwsheng009/mint/framework/interactive"

button := interactive.NewButton("Click Me")
button.SetOnClick(func() {
    fmt.Println("Clicked!")
})
container.Add(button)
```

#### 新 API

```go
import "github.com/wwsheng009/mint/ui"

button := ui.Button("Click Me").
    OnClick(func() {
        fmt.Println("Clicked!")
    })

// 或直接在布局中
ui.VStack(
    ui.Button("Click Me").OnClick(func() {
        fmt.Println("Clicked!")
    }),
)
```

#### 迁移步骤

1. 将 `interactive.NewButton()` 替换为 `ui.Button()`
2. 将 `SetOnClick()` 改为 `.OnClick()`
3. 删除 `container.Add(button)`

---

### 3.3 Input 组件

#### 旧 API

```go
import "github.com/wwsheng009/mint/framework/input"

textInput := input.NewTextInput()
textInput.SetPlaceholder("Enter text...")
textInput.SetOnChange(func(value string) {
    fmt.Println("Input:", value)
})
container.Add(textInput)
```

#### 新 API

```go
import "github.com/wwsheng009/mint/ui"

input := ui.Input("Enter text...").
    OnChange(func(value string) {
        fmt.Println("Input:", value)
    })

// 或受控组件
func SearchBox() ui.VNode {
    query, setQuery := ui.UseState("")

    return ui.Input("Search...").
        Value(query).
        OnChange(func(value string) {
            setQuery(value)
        })
}
```

#### 迁移步骤

1. 将 `input.NewTextInput()` 替换为 `ui.Input()`
2. 将 `SetPlaceholder()` 改为构造函数参数
3. 将 `SetOnChange()` 改为 `.OnChange()`
4. 考虑使用受控组件模式

---

### 3.4 布局组件

#### 旧 API (Box)

```go
import "github.com/wwsheng009/mint/framework/layout"

box := layout.NewBox()
box.SetBorder(true)
box.SetPadding(4)
box.SetBackground(color.Blue)

box.Add(display.NewText("Content"))
container.Add(box)
```

#### 新 API

```go
import "github.com/wwsheng009/mint/ui"

ui.Box().
    Border(true).
    Padding(4).
    Background(color.Blue).
    Child(
        ui.Text("Content"),
    )
```

#### 旧 API (Flex)

```go
import "github.com/wwsheng009/mint/framework/layout"

flex := layout.NewFlex(layout.DirectionRow)
flex.SetGap(2)
flex.Add(display.NewText("A"))
flex.Add(display.NewText("B"))
flex.Add(display.NewText("C"))
container.Add(flex)
```

#### 新 API

```go
import "github.com/wwsheng009/mint/ui"

ui.HStack(
    ui.Text("A"),
    ui.Text("B"),
    ui.Text("C"),
).Gap(2)
```

---

### 3.5 列表组件

#### 旧 API

```go
import "github.com/wwsheng009/mint/framework/display"

list := display.NewList()
list.SetItems([]string{"A", "B", "C"})
list.SetOnSelect(func(index int) {
    fmt.Println("Selected:", index)
})
container.Add(list)
```

#### 新 API

```go
import "github.com/wwsheng009/mint/ui"

ui.List([]string{"A", "B", "C"}).
    OnSelect(func(index int) {
        fmt.Println("Selected:", index)
    })

// 或使用 For 动态生成
func ItemList(items []string) ui.VNode {
    return ui.VStack(
        ui.For(items, func(i int, item string) ui.VNode {
            return ui.Text(item).
                OnClick(func() {
                    fmt.Println("Selected:", item)
                }).
                Key(item)
        }),
    )
}
```

---

## 四、状态管理迁移

### 4.1 StateHolder → useState

#### 旧 API

```go
import "github.com/wwsheng009/mint/framework/component"

type Counter struct {
    *component.BaseComponent
    *component.StateHolder
}

func NewCounter() *Counter {
    c := &Counter{
        BaseComponent: component.NewBaseComponent("counter"),
        StateHolder:   component.NewStateHolder(),
    }
    c.SetState("count", 0)
    return c
}

func (c *Counter) Render() VNode {
    count := c.GetState("count").(int)

    return ui.VStack(
        ui.Text(fmt.Sprintf("Count: %d", count)),
        ui.Button("Increment").OnClick(func() {
            c.SetState("count", count+1)
            c.MarkDirty()
        }),
    )
}
```

#### 新 API (函数组件)

```go
import "github.com/wwsheng009/mint/ui"

func Counter() ui.VNode {
    count, setCount := ui.UseStateInt(0)

    return ui.VStack(
        ui.Text(fmt.Sprintf("Count: %d", count)),
        ui.Button("Increment").OnClick(func() {
            setCount(count + 1)
        }),
    )
}
```

#### 迁移步骤

1. 将结构体组件改为函数组件
2. 将 `StateHolder` 替换为 `useState`
3. 删除 `GetState()`/`SetState()` 调用
4. 删除 `MarkDirty()` 调用（自动处理）

---

### 4.2 ReactiveStore → useContext

#### 旧 API

```go
import "github.com/wwsheng009/mint/framework/binding"

store := binding.NewReactiveStore()
store.Set("user", User{Name: "Alice"})

func UserProfile() VNode {
    user := store.Get("user").(User)

    return ui.Text("Hello, " + user.Name)
}

store.Subscribe("user", func(key string, value interface{}) {
    // 更新界面
})
```

#### 新 API

```go
import "github.com/wwsheng009/mint/ui"

// 创建上下文
var UserContext = ui.CreateContext(nil)

func App() ui.VNode {
    user, setUser := ui.UseState(User{Name: "Alice"})

    return ui.Provider(UserContext, user,
        UserProfile(),
    )
}

func UserProfile() ui.VNode {
    user := ui.UseContext(UserContext).(User)

    return ui.Text("Hello, " + user.Name)
}
```

---

## 五、事件处理迁移

### 5.1 点击事件

#### 旧 API

```go
button := interactive.NewButton("Click")
button.SetOnClick(func() {
    fmt.Println("Clicked")
})
```

#### 新 API

```go
ui.Button("Click").OnClick(func() {
    fmt.Println("Clicked")
})
```

---

### 5.2 捕获阶段事件

#### 新 API (新增)

```go
ui.Button("Click").
    OnClickCapture(func(e ui.Event) {
        fmt.Println("Capture phase")
        e.StopPropagation() // 阻止继续传播
    }).
    OnClick(func(e ui.Event) {
        fmt.Println("Bubble phase")
    })
```

---

### 5.3 键盘事件

#### 旧 API

```go
component.SetOnKeyPress(func(key rune, mod input.KeyMod) {
    if key == 'q' && mod == input.ModCtrl {
        // 退出
    }
})
```

#### 新 API

```go
ui.Text("Press Ctrl+Q to quit").
    OnKeyPress(func(e ui.KeyEvent) {
        if e.Key == 'q' && e.Mod == ui.ModCtrl {
            // 退出
        }
    })
```

---

## 六、样式迁移

### 6.1 颜色设置

#### 旧 API

```go
text.SetFgColor(color.RGB(255, 0, 0))
text.SetBgColor(color.ANSI(9)) // ANSI 红色
```

#### 新 API

```go
ui.Text("Hello").
    FgColor(color.RGB(255, 0, 0))

// 或使用预定义颜色
ui.Text("Hello").
    FgColor(color.Red)
```

---

### 6.2 样式组合

#### 旧 API

```go
text.SetBold(true)
text.SetItalic(true)
text.SetUnderline(true)
```

#### 新 API

```go
ui.Text("Hello").
    Bold(true).
    Italic(true).
    Underline(true)

// 或使用 Style
myStyle := ui.Style().
    Bold().
    Italic().
    Underline().
    Build()

ui.Text("Hello").Style(myStyle)
```

---

### 6.3 主题样式

#### 旧 API

```go
text.SetStyleID("myText")
text.SetState("focus")
// 主题系统自动应用样式
```

#### 新 API

```go
// 使用上下文传递主题
func ThemedText() ui.VNode {
    theme := ui.UseContext(ThemeContext).(theme.Theme)

    return ui.Text("Hello").
        FgColor(theme.TextColor).
        BgColor(theme.BackgroundColor)
}
```

---

## 七、布局迁移

### 7.1 容器布局

#### 旧 API

```go
container := layout.NewBox()
container.SetBorder(true)
container.SetPadding(4)

// 添加子组件
container.Add(text1)
container.Add(text2)
container.Add(text3)

parent.Add(container)
```

#### 新 API

```go
ui.Box().
    Border(true).
    Padding(4).
    Children(
        ui.Text("1"),
        ui.Text("2"),
        ui.Text("3"),
    )

// 或直接嵌套
ui.Box().Border(true).Padding(4).
    Child(
        ui.VStack(
            ui.Text("1"),
            ui.Text("2"),
            ui.Text("3"),
        ),
    )
```

---

### 7.2 Flex 布局

#### 旧 API

```go
flex := layout.NewFlex(layout.DirectionRow)
flex.SetGap(2)
flex.SetAlign(layout.AlignCenter)

flex.AddItem(item1, layout.Flex(1))
flex.AddItem(item2, layout.Flex(2))
flex.AddItem(item3, layout.Flex(1))
```

#### 新 API

```go
ui.HStack(
    ui.Text("A").Flex(1),
    ui.Text("B").Flex(2),
    ui.Text("C").Flex(1),
).Gap(2).Align(ui.AlignCenter)
```

---

### 7.3 绝对定位

#### 旧 API

```go
component.SetPosition(10, 5)
component.SetSize(20, 10)
```

#### 新 API

```go
ui.Absolute().
    X(10).Y(5).
    Width(20).Height(10).
    Child(
        ui.Text("Positioned"),
    )
```

---

## 八、常见问题

### 8.1 如何处理生命周期方法？

#### 旧 API

```go
type MyComponent struct {
    *component.BaseComponent
}

func (c *MyComponent) OnMount(ctx Context) {
    // 初始化
}

func (c *MyComponent) OnUnmount(ctx Context) {
    // 清理
}
```

#### 新 API

```go
func MyComponent() ui.VNode {
    ui.UseEffect(func() {
        // 初始化 (OnMount)

        return func() {
            // 清理 (OnUnmount)
        }
    }, nil)

    return ui.Text("Hello")
}
```

---

### 8.2 如何访问组件实例？

#### 新 API (useRef)

```go
func InputWithFocus() ui.VNode {
    inputRef := ui.UseRef(nil)

    ui.UseEffect(func() {
        if input := inputRef.Get().(*InputElement); input != nil {
            input.Focus()
        }
    }, nil)

    return ui.Input("Focus me...").Ref(inputRef)
}
```

---

### 8.3 如何处理表单提交？

#### 旧 API

```go
form := form.NewForm()
form.SetOnSubmit(func(data form.Data) {
    // 处理提交
})
```

#### 新 API

```go
func LoginForm() ui.VNode {
    username, setUsername := ui.UseState("")
    password, setPassword := ui.UseState("")

    return ui.VStack(
        ui.Input("Username").
            Value(username).
            OnChange(setUsername),

        ui.Input("Password").
            Password(true).
            Value(password).
            OnChange(setPassword),

        ui.Button("Login").OnClick(func() {
            // 直接访问状态
            login(username, password)
        }),
    )
}
```

---

### 8.4 如何处理列表 Key？

#### 问题：使用索引作为 Key

```go
// ❌ 错误
ui.For(items, func(i int, item Item) ui.VNode {
    return ui.Text(item.Name).Key(strconv.Itoa(i))
})
```

#### 解决方案：使用稳定的 ID

```go
// ✅ 正确
ui.For(items, func(i int, item Item) ui.VNode {
    return ui.Text(item.Name).Key(item.ID)
})
```

---

### 8.5 如何避免闭包陷阱？

#### 问题：闭包捕获旧值

```go
// ❌ 错误
func Counter() ui.VNode {
    count, setCount := ui.UseStateInt(0)

    return ui.Button("Click").OnClick(func() {
        fmt.Println(count) // 总是打印 0
    })
}
```

#### 解决方案：使用函数更新

```go
// ✅ 正确
func Counter() ui.VNode {
    count, setCount := ui.UseStateInt(0)

    return ui.Button("Click").OnClick(func() {
        setCount(count + 1) // 使用当前值
    })
}

// 或使用 updater 函数
func Counter() ui.VNode {
    count, setCount := ui.UseStateInt(0)

    return ui.Button("Click").OnClick(func() {
        setCount(func(prev int) int {
            return prev + 1
        })
    })
}
```

---

## 九、检查清单

### 9.1 迁移前

- [ ] 备份现有代码
- [ ] 评估迁移范围
- [ ] 更新依赖版本
- [ ] 阅读新 API 文档

### 9.2 迁移中

- [ ] 更新 import 路径
- [ ] 替换组件创建方式
- [ ] 更新属性设置为链式调用
- [ ] 迁移状态管理到 Hooks
- [ ] 更新事件处理
- [ ] 更新布局代码

### 9.3 迁移后

- [ ] 编译通过
- [ ] 单元测试通过
- [ ] 集成测试通过
- [ ] 功能验证
- [ ] 性能验证

### 9.4 代码审查

- [ ] 移除未使用的导入
- [ ] 移除废弃的 API 调用
- [ ] 统一代码风格
- [ ] 添加必要注释
- [ ] 更新文档

---

## 十、快速参考

### 组件 API 对照表

| 旧 API | 新 API |
|--------|--------|
| `display.NewText()` | `ui.Text()` |
| `interactive.NewButton()` | `ui.Button()` |
| `input.NewTextInput()` | `ui.Input()` |
| `display.NewList()` | `ui.List()` |
| `display.NewTable()` | `ui.Table()` |
| `layout.NewBox()` | `ui.Box()` |
| `layout.NewFlex(dir)` | `ui.HStack()` / `ui.VStack()` |

### 方法对照表

| 旧 API | 新 API |
|--------|--------|
| `SetX()` | `.X(value)` |
| `SetOnClick(fn)` | `.OnClick(fn)` |
| `SetOnChange(fn)` | `.OnChange(fn)` |
| `Add(child)` | `.Child(child)` 或 `Children(...)` |
| `container.Add(c)` | 直接在布局中使用 |

### Hooks 对照表

| 旧 API | 新 API |
|--------|--------|
| `StateHolder` | `useState` |
| `ReactiveStore` | `useContext` + `useState` |
| `OnMount/OnUnmount` | `useEffect` |
| (无对应) | `useMemo` |
| (无对应) | `useCallback` |
| (无对应) | `useRef` |

---

**文档结束**

**版本历史**:
- v1.0 (2026-01-31): 初始版本
