# Sandbox 调试技巧指南

本文档介绍如何使用 Mint 的 Sandbox 测试功能来调试 TUI 应用。

---

## 目录

1. [快速开始](#快速开始)
2. [RunTest API](#runtest-api)
3. [事件注入](#事件注入)
4. [调试技巧](#调试技巧)
5. [常见场景](#常见场景)
6. [最佳实践](#最佳实践)

---

## 快速开始

### 基本测试模板

```go
func TestMyComponent(t *testing.T) {
    t.Setenv("MINT_USE_FIBER", "true")
    t.Setenv("TUI_DEBUG_UI", "true")

    // 创建测试应用
    testApp, err := ui.RunTest(MyComponent,
        ui.WithWidth(40),
        ui.WithHeight(12),
    )
    if err != nil {
        t.Fatalf("RunTest failed: %v", err)
    }
    defer testApp.Close()

    // 等待初始化
    time.Sleep(100 * time.Millisecond)

    // 注入事件
    testApp.InjectSpecialKey(platform.KeyEnter)

    // 等待更新
    time.Sleep(50 * time.Millisecond)

    // 验证结果
    rendered := testApp.GetRenderString()
    if !strings.Contains(rendered, "Expected Text") {
        t.Error("Expected text not found")
    }
}
```

---

## RunTest API

### 创建测试应用

```go
testApp, err := ui.RunTest(ComponentFunc,
    ui.WithWidth(80),    // 窗口宽度
    ui.WithHeight(24),   // 窗口高度
    ui.WithTitle("Test"), // 窗口标题
)
```

### TestableApp 接口

```go
type TestableApp struct {
    fwApp *framework.App  // 框架应用
    root  *declarativeRoot // 根组件
    opts  *Options         // 配置
}
```

### 常用方法

| 方法 | 用途 |
|------|------|
| `Close()` | 关闭测试应用 |
| `InjectKey(rune)` | 注入字符键 |
| `InjectSpecialKey(SpecialKey)` | 注入特殊键 |
| `GetRenderString()` | 获取渲染结果字符串 |
| `GetBuffer()` | 获取渲染缓冲区 |
| `GetFrameworkApp()` | 获取框架应用 |

---

## 事件注入

### 键盘事件

```go
// 字符键
testApp.InjectKey('a')
testApp.InjectKey('1')
testApp.InjectKey('q')  // 退出

// 特殊键
testApp.InjectSpecialKey(platform.KeyEnter)
testApp.InjectSpecialKey(platform.KeyTab)
testApp.InjectSpecialKey(platform.KeyEscape)
testApp.InjectSpecialKey(platform.KeyBackspace)
testApp.InjectSpecialKey(platform.KeyUp)
testApp.InjectSpecialKey(platform.KeyDown)
testApp.InjectSpecialKey(platform.KeyLeft)
testApp.InjectSpecialKey(platform.KeyRight)

// 带修饰符
// 需要通过 RawInput 直接注入
```

### RawInput 注入（高级）

```go
// 创建原始输入
raw := platform.RawInput{
    Type:    platform.InputKeyPress,
    Key:     'a',
    Special: platform.KeyUnknown,
    Modifiers: platform.ModCtrl | platform.ModShift,
}

// 通过 framework App 注入
testApp.GetFrameworkApp().InjectEvent(raw)
```

---

## 调试技巧

### 技巧 1: 环境变量控制日志

```go
// 启用调试输出
t.Setenv("TUI_DEBUG_UI", "true")

// 启用 Pump 调试
t.Setenv("TUI_DEBUG_PUMP", "true")

// 启用 Fiber 模式
t.Setenv("MINT_USE_FIBER", "true")
```

### 技巧 2: 在组件中添加日志

```go
func MyCounter() ui.VNode {
    count, setCount, _ := ui.UseStateInt(0)

    // 调试：打印当前值
    if os.Getenv("TUI_DEBUG_UI") == "true" {
        fmt.Fprintf(os.Stderr, "[MyCounter] Render: count=%d\n", count)
    }

    ui.On(ui.SimpleIncrementIntent{}, func(ctx *intent.ActionContext) {
        if os.Getenv("TUI_DEBUG_UI") == "true" {
            fmt.Fprintf(os.Stderr, "[MyCounter] intent: incrementing\n")
        }
        setCount(func(old int) int { return old + 1 })
    })

    return ui.VStack(
        ui.NewTextBuilder(fmt.Sprintf("Count: %d", count)).Build(),
        ui.NewButtonBuilder("+").
            OnPress(ui.SimpleIncrementIntent{}).
            Build(),
    )
}
```

### 技巧 3: 追踪 VNode 生命周期

```go
func MyComponent() ui.VNode {
    value, _, _ := ui.UseStateInt(0)

    // 创建带追踪的 VNode
    textNode := ui.NewTextBuilder(fmt.Sprintf("Value: %d", value)).Build()

    if os.Getenv("TUI_DEBUG_UI") == "true" {
        fmt.Fprintf(os.Stderr, "[VNode] Created TextVNode:\n")
        fmt.Fprintf(os.Stderr, "  ptr=%p\n", textNode)
        fmt.Fprintf(os.Stderr, "  content=%s\n", textNode.(*ui.TextVNode).Content())
    }

    return textNode
}
```

### 技巧 4: 在关键路径添加日志

#### hooks.go - useState
```go
func useState(initial interface{}) (interface{}, func(interface{})) {
    ctx := getCurrentContext()
    hookIndex := ctx.HookIndex
    hook := ctx.getOrCreateHook(HookState)

    if !hook.Initialized {
        hook.Value = initial
        hook.Initialized = true

        if os.Getenv("TUI_DEBUG_UI") == "true" {
            fmt.Fprintf(os.Stderr, "[useState] INIT: hookIndex=%d, value=%v\n",
                hookIndex, initial)
        }
    } else {
        if os.Getenv("TUI_DEBUG_UI") == "true" {
            fmt.Fprintf(os.Stderr, "[useState] REUSE: hookIndex=%d, value=%v\n",
                hookIndex, hook.Value)
        }
    }

    currentValue := hook.Value
    // ... setter 逻辑
    return currentValue, setState
}
```

#### reconciler.go - 追踪 VNode 更新
```go
func cloneExistingFiber(returnFiber *Fiber, current *Fiber, vnode VNode) *Fiber {
    if os.Getenv("TUI_DEBUG_UI") == "true" {
        fmt.Fprintf(os.Stderr, "[cloneExistingFiber]\n")
        fmt.Fprintf(os.Stderr, "  old VNode ptr=%p, type=%T\n", current.VNode, current.VNode)
        fmt.Fprintf(os.Stderr, "  new VNode ptr=%p, type=%T\n", vnode, vnode)
    }

    fiber := CloneFiber(current)
    fiber.VNode = vnode  // 关键：更新 VNode

    return fiber
}
```

### 技巧 5: 强制立即渲染

```go
// 等待异步更新
testApp.InjectSpecialKey(platform.KeyEnter)
time.Sleep(50 * time.Millisecond)

// 强制立即渲染以获取最新状态
testApp.GetFrameworkApp().ForceRenderNow()

// 获取渲染结果
rendered := testApp.GetRenderString()
```

### 技巧 6: 检查内部状态

```go
// 获取 framework.App
fwApp := testApp.GetFrameworkApp()

// 检查状态
if os.Getenv("TUI_DEBUG_UI") == "true" {
    fmt.Fprintf(os.Stderr, "[State] App state=%d\n", fwApp.GetState())
    fmt.Fprintf(os.Stderr, "[State] IsDirty=%v\n", fwApp.IsDirty())
}
```

### 技巧 7: 捕获 panic

```go
func TestPanicRecovery(t *testing.T) {
    defer func() {
        if r := recover(); r != nil {
            t.Logf("Caught panic: %v", r)
            // 打印堆栈
            debug.PrintStack()
        }
    }()

    testApp, _ := ui.RunTest(MyComponent)
    defer testApp.Close()

    // 可能触发 panic 的操作
    testApp.InjectSpecialKey(platform.KeyEnter)
}
```

---

## 常见场景

### 场景 1: 测试按钮点击

```go
func TestButtonClick(t *testing.T) {
    testApp, _ := ui.RunTest(MyComponent)
    defer testApp.Close()

    time.Sleep(100 * time.Millisecond)

    // 1. 切换焦点到按钮
    testApp.InjectSpecialKey(platform.KeyTab)
    time.Sleep(50 * time.Millisecond)

    // 2. 触发点击
    testApp.InjectSpecialKey(platform.KeyEnter)
    time.Sleep(100 * time.Millisecond)

    // 3. 验证
    rendered := testApp.GetRenderString()
    if strings.Contains(rendered, "Clicked!") {
        t.Log("✅ Button click works")
    }
}
```

### 场景 2: 测试文本输入

```go
func TestTextInput(t *testing.T) {
    testApp, _ := ui.RunTest(FormComponent)
    defer testApp.Close()

    // 切换到输入框
    testApp.InjectSpecialKey(platform.KeyTab)
    time.Sleep(50 * time.Millisecond)

    // 输入文本
    testApp.InjectKey('H')
    testApp.InjectKey('e')
    testApp.InjectKey('l')
    testApp.InjectKey('l')
    testApp.InjectKey('o')
    time.Sleep(50 * time.Millisecond)

    // 验证
    rendered := testApp.GetRenderString()
    if strings.Contains(rendered, "Hello") {
        t.Log("✅ Text input works")
    }
}
```

### 场景 3: 测试状态更新

```go
func TestStateUpdate(t *testing.T) {
    testApp, _ := ui.RunTest(Counter)
    defer testApp.Close()

    // 获取初始状态
    initial := testApp.GetRenderString()
    t.Logf("Initial:\n%s", initial)

    // 触发状态更新
    testApp.InjectSpecialKey(platform.KeyTab)  // 切换到按钮
    testApp.InjectSpecialKey(platform.KeyEnter) // 点击
    time.Sleep(100 * time.Millisecond)

    // 获取更新后状态
    updated := testApp.GetRenderString()
    t.Logf("Updated:\n%s", updated)

    // 验证状态变化
    if !strings.Contains(initial, "Count: 1") &&
        strings.Contains(updated, "Count: 1") {
        t.Log("✅ State update works")
    }
}
```

### 场景 4: 测试焦点切换

```go
func TestFocusNavigation(t *testing.T) {
    testApp, _ := ui.RunTest(FormWithMultipleInputs)
    defer testApp.Close()

    time.Sleep(100 * time.Millisecond)

    // 记录初始焦点状态
    initial := testApp.GetRenderString()
    hasFocus1 := strings.Contains(initial, "[focused]")

    // 按 Tab 切换焦点
    testApp.InjectSpecialKey(platform.KeyTab)
    time.Sleep(50 * time.Millisecond)

    // 检查焦点移动
    afterTab := testApp.GetRenderString()
    hasFocus2 := strings.Contains(afterTab, "[focused]")

    if hasFocus1 != hasFocus2 {
        t.Log("✅ Focus changed after Tab")
    }
}
```

### 场景 5: 追踪事件流

```go
func TestEventFlow(t *testing.T) {
    t.Setenv("TUI_DEBUG_UI", "true")

    // 设置自定义日志
    ui.SetDebugLogger(func(msg string) {
        if strings.Contains(msg, "HandleEvent") ||
           strings.Contains(msg, "onClick") ||
           strings.Contains(msg, "setState") {
            t.Logf("[EVENT] %s", msg)
        }
    })

    testApp, _ := ui.RunTest(MyComponent)
    defer testApp.Close()

    // 触发事件
    testApp.InjectSpecialKey(platform.KeyEnter)

    // 检查日志
    // ...
}
```

---

## 最佳实践

### 1. 测试结构

```go
func TestXxx(t *testing.T) {
    // 1. 设置环境变量
    t.Setenv("MINT_USE_FIBER", "true")
    t.Setenv("TUI_DEBUG_UI", "false")  // 生产测试关闭调试

    // 2. 创建测试应用
    testApp, err := ui.RunTest(Component, opts...)
    if err != nil {
        t.Fatalf("RunTest: %v", err)
    }
    defer testApp.Close()  // 确保清理

    // 3. 等待初始化
    time.Sleep(100 * time.Millisecond)

    // 4. 执行操作
    // ...

    // 5. 验证结果
    // ...
}
```

### 2. 等待策略

```go
// ❌ 不好：固定等待
time.Sleep(1 * time.Second)

// ✅ 好：足够等待 + 验证
for i := 0; i < 10; i++ {
    rendered := testApp.GetRenderString()
    if strings.Contains(rendered, "Expected") {
        break  // 找到了，提前退出
    }
    time.Sleep(10 * time.Millisecond)
}
```

### 3. 日志策略

```go
// 开发阶段：启用详细日志
t.Setenv("TUI_DEBUG_UI", "true")

// CI/生产测试：关闭日志
t.Setenv("TUI_DEBUG_UI", "false")

// 使用 t.Log 而不是 fmt.Println
t.Log("This appears in test output")
t.Logf("Value: %d", value)
t.Error("This is a test error")
```

### 4. 隔离测试

```go
// 每个测试独立创建应用
func TestA(t *testing.T) {
    testApp, _ := ui.RunTest(Component)
    defer testApp.Close()
    // ...
}

func TestB(t *testing.T) {
    testApp, _ := ui.RunTest(Component)  // 新实例
    defer testApp.Close()
    // ...
}
```

### 5. 调试日志分层

```go
// Level 1: 组件层 (最常用)
if os.Getenv("TUI_DEBUG_UI") == "true" {
    fmt.Fprintf(os.Stderr, "[Component] value=%d\n", value)
}

// Level 2: Hooks 层
if os.Getenv("TUI_DEBUG_HOOKS") == "true" {
    fmt.Fprintf(os.Stderr, "[Hook] setState: %d -> %d\n", old, new)
}

// Level 3: Reconciler 层
if os.Getenv("TUI_DEBUG_RECONCILER") == "true" {
    fmt.Fprintf(os.Stderr, "[Reconciler] clone fiber\n")
}
```

---

## 调试命令速查

```bash
# 运行单个测试
go test -v -run TestMyComponent ./examples/myapp/

# 运行所有测试
go test -v ./examples/myapp/

# 启用调试输出
TUI_DEBUG_UI=true go test -v ./examples/myapp/

# 运行基准测试
go test -bench=. ./examples/myapp/

# 覆盖率
go test -cover ./examples/myapp/

# 详细输出
go test -v -count=1 ./examples/myapp/
```

---

## 故障排查

### 问题：测试超时

**原因**: 使用了 `ui.Run()` 而非 `ui.RunTest()`

**解决**:
```go
// ❌ 错误 - 需要手动交互
err := ui.Run(Component)

// ✅ 正确 - 自动化测试
testApp, err := ui.RunTest(Component)
defer testApp.Close()
```

### 问题：状态没有更新

**调试步骤**:
1. 确认 onClick 被调用
2. 确认 setState 被调用
3. 检查 hook.Value 是否更新
4. 检查组件是否重新渲染
5. 检查 VNode 是否被正确创建

```go
// 添加追踪日志
if os.Getenv("TUI_DEBUG_UI") == "true" {
    fmt.Fprintf(os.Stderr, "[1] onClick called\n")
    fmt.Fprintf(os.Stderr, "[2] setState called\n")
    fmt.Fprintf(os.Stderr, "[3] hook.Value=%d\n", hook.Value)
    fmt.Fprintf(os.Stderr, "[4] rendering with count=%d\n", count)
    fmt.Fprintf(os.Stderr, "[5] VNode content=%s\n", vnode.Content())
}
```

### 问题：事件没有触发

**检查清单**:
- [ ] 焦点在正确的元素上
- [ ] 按钮没有禁用
- [ ] 事件处理器正确注册
- [ ] 等待时间足够

```go
// 检查焦点状态
testApp.InjectSpecialKey(platform.KeyTab)
time.Sleep(100 * time.Millisecond)  // 等待焦点切换完成

// 检查按钮状态
rendered := testApp.GetRenderString()
if strings.Contains(rendered, "[disabled]") {
    t.Log("Button is disabled")
}
```
