# Testing - 测试工具模块

> Mock、Fixture、断言、测试辅助工具

## 功能概述

Testing 模块提供 DevTools 的测试工具，包括 Mock 对象、测试 Fixture 和断言函数，用于单元测试和集成测试。

## 核心组件

### 1. Mock (`mock.go`)

```go
// MockDevTools 模拟 DevTools
type MockDevTools struct {
    mu            sync.RWMutex
    enabled       bool
    currentFrame  FrameID
    events        []Event
    mutations     []Mutation
    layouts       []Layout
}

// 创建 Mock
func NewMock() *MockDevTools

// 启用
func (m *MockDevTools) Enable()

// 记录事件
func (m *MockDevTools) RecordEvent(eventType, nodeID, phase string, data map[string]interface{}) FrameID

// 获取记录的事件
func (m *MockDevTools) GetEvents() []Event

// 重置
func (m *MockDevTools) Reset()
```

### 2. Fixture (`fixture.go`)

```go
// Fixture 测试 Fixture
type Fixture struct {
    devtools *MockDevTools
}

// 创建 Fixture
func NewFixture() *Fixture

// 创建测试场景
func (f *Fixture) CreateSimpleScene() *Fixture

// 创建复杂场景
func (f *Fixture) CreateComplexScene() *Fixture

// 清理
func (f *Fixture) Teardown()
```

### 3. Assertion (`assertion.go`)

```go
// AssertHelper 断言助手
type AssertHelper struct {
    t *testing.T
}

// 创建断言助手
func NewAssert(t *testing.T) *AssertHelper

// 断言事件数量
func (a *AssertHelper) EventCount(count int)

// 断言存在事件
func (a *AssertHelper) HasEvent(eventType string) bool

// 断言因果关系
func (a *AssertHelper) HasCausalLink(from, to string) bool

// 断言快照差异
func (a *AssertHelper) DiffChangeCount(from, to FrameID, count int)
```

## 使用方法

### Mock 基础使用

```go
import "github.com/wwsheng009/mint/devtools/testing"

// 创建 Mock
mock := testing.NewMock()
mock.Enable()

// 记录事件
mock.RecordEvent("keypress", "button-1", "bubble", nil)
mock.RecordEvent("mutation", "button-1", "target", nil)

// 获取事件
events := mock.GetEvents()
fmt.Printf("Recorded %d events\n", len(events))

// 重置
mock.Reset()
```

### Fixture 使用

```go
// 创建 Fixture
fixture := testing.NewFixture()
defer fixture.Teardown()

// 创建简单测试场景
fixture.CreateSimpleScene()
// - 创建 3 个组件
// - 记录 5 个事件
// - 产生 2 个布局变更

// 获取 DevTools
dt := fixture.GetDevTools()

// 获取测试组件
components := fixture.GetComponents()
```

### 断言使用

```go
func TestMyComponent(t *testing.T) {
    fixture := testing.NewFixture()
    defer fixture.Teardown()

    fixture.CreateSimpleScene()

    // 使用断言
    assert := testing.NewAssert(t)

    // 断言事件数量
    assert.EventCount(5)

    // 断言存在特定事件
    if assert.HasEvent("keypress") {
        t.Log("Keypress event found")
    }

    // 断言因果关系
    if assert.HasCausalLink("keypress", "mutation") {
        t.Log("Causal link verified")
    }
}
```

## 测试场景

### 简单场景

```go
func (f *Fixture) CreateSimpleScene() *Fixture {
    // 创建按钮组件
    f.AddComponent("button-1", "Button", map[string]interface{}{
        "label": "Click me",
    })

    // 创建输入框组件
    f.AddComponent("input-1", "Input", map[string]interface{}{
        "value": "",
    })

    // 记录事件
    f.RecordEvent("focus", "input-1", "target", nil)
    f.RecordEvent("input", "input-1", "target", map[string]interface{}{
        "value": "hello",
    })

    return f
}
```

### 复杂场景

```go
func (f *Fixture) CreateComplexScene() *Fixture {
    // 创建多个组件
    for i := 0; i < 10; i++ {
        f.AddComponent(
            fmt.Sprintf("comp-%d", i),
            "Container",
            map[string]interface{}{
                "index": i,
            },
        )
    }

    // 记录多个事件
    for i := 0; i < 50; i++ {
        f.RecordEvent("keypress", fmt.Sprintf("comp-%d", i%10), "bubble", nil)
    }

    // 创建布局变更
    f.RecordLayout("comp-0", LayoutDelta{...})

    return f
}
```

## 辅助函数

### 组件创建

```go
// 创建按钮组件
func (f *Fixture) AddButton(id, label string) {
    f.AddComponent(id, "Button", map[string]interface{}{
        "label": label,
    })
}

// 创建输入框组件
func (f *Fixture) AddInput(id, value string) {
    f.AddComponent(id, "Input", map[string]interface{}{
        "value": value,
    })
}

// 创建容器组件
func (f *Fixture) AddContainer(id string, children []string) {
    f.AddComponent(id, "Container", map[string]interface{}{
        "children": children,
    })
}
```

### 事件记录

```go
// 记录键盘事件
func (f *Fixture) RecordKeypress(key string, nodeID string) {
    f.RecordEvent("keypress", nodeID, "bubble", map[string]interface{}{
        "key": key,
    })
}

// 记录鼠标事件
func (f *Fixture) RecordMouse(x, y int, button string) {
    f.RecordEvent("mouse", "", "target", map[string]interface{}{
        "x": x, "y": y, "button": button,
    })
}

// 记录焦点事件
func (f *Fixture) RecordFocus(nodeID string) {
    f.RecordEvent("focus", nodeID, "target", nil)
}
```

## 相关模块

| 模块 | 关系 |
|------|------|
| **所有模块** | Testing 模块为所有模块提供测试支持 |
| `devtools` | Mock 实现其接口 |
| `snapshot` | Fixture 创建测试快照 |
| `causal` | 断言验证因果链 |

## API 参考

### TestContext

```go
type TestContext struct {
    Name       string
    SetupFunc  func(*Fixture)
    TeardownFunc func(*Fixture)
    AssertFunc  func(*AssertHelper)
}
```

### TestData

```go
type TestData struct {
    Components []ComponentData
    Events     []EventData
    Layouts    []LayoutData
}
```

## 文件列表

- `mock.go` - Mock 对象
- `fixture.go` - 测试 Fixture
- `assertion.go` - 断言函数
