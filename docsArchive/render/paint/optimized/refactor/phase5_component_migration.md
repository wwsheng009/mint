# Phase 5: 组件迁移

## 概述
**时间**: 5-7 天
**优先级**: P1（重要）
**依赖**: Phase 1-4 完成

## 目标
将所有组件迁移到 Fiber-first 架构，实现 PaintableInstance 接口。

---

## 架构约束

> **核心原则**：
> 1. VNode 不持有闭包，使用 ActionTargetID 进行事件路由
> 2. Instance 通过 ActionHandlerInstance 接口处理事件
> 3. 所有事件通过 ActionBridge 路由

### 组件通信模型

根据 Fiber-first 架构设计，组件间通信应遵循以下原则：

```
VNode (声明)
    ↓ 只包含 ActionTargetID
Reconcile
    ↓
Fiber (存储 ActionTargetID)
    ↓
Event → ActionBridge → Dispatcher → Instance.HandleAction()
```

### 禁止的模式

```go
// ❌ 错误：VNode 持有闭包
type VNode struct {
    onClick func()  // 禁止
}

// ❌ 错误：VNode 持有 Intent
type VNode struct {
    actionIntent intent.Intent  // 禁止
}

// ❌ 错误：Instance 直接发射 Intent
type Instance struct {
    intentEmitter func(intent.Intent)  // 禁止
}
```

### 正确的模式

```go
// ✅ 正确：VNode 只声明 ActionTargetID
type VNode struct {
    actionTargetID string  // 用于路由到组件
}

// ✅ 正确：Instance 实现 ActionHandlerInstance 接口
type Instance struct {
    // 无闭包，通过接口处理事件
}

func (inst *Instance) HandleAction(actionType string, payload interface{}) bool {
    // 处理事件
}
```

---

## 迁移优先级

### P0: 基础组件（必须）
| 组件 | 原位置 | 目标位置 | 状态 |
|------|--------|----------|------|
| Text | components/text | ui/components/text | 待迁移 |
| VStack | components/layout | ui/components/stack | 待迁移 |
| HStack | components/layout | ui/components/stack | 待迁移 |
| Spacer | components/layout | ui/components/spacer | 待迁移 |

### P1: 交互组件（重要）
| 组件 | 原位置 | 目标位置 | 状态 |
|------|--------|----------|------|
| Input | components/input | ui/components/input | 待迁移 |
| Checkbox | components/checkbox | ui/components/checkbox | 待迁移 |
| Select | components/select | ui/components/select | 待迁移 |

### P2: 复杂组件（可选）
| 组件 | 原位置 | 目标位置 | 状态 |
|------|--------|----------|------|
| Table | components/table | ui/components/table | 待迁移 |
| Modal | components/modal | ui/components/modal | 待迁移 |
| List | components/list | ui/components/list | 待迁移 |

---

## 迁移模板

### 目录结构
```
ui/components/<component>/
├── vnode.go        # VNode 描述（纯声明）
├── instance.go     # Instance 运行期实体（状态 + 渲染）
├── builder.go      # Builder 流式 API（可选）
├── *_test.go       # 单元测试
└── README.md       # 组件文档
```

### VNode 模板

**文件**: `vnode.go`

```go
package <component>

import (
    rtui "github.com/wwsheng009/mint/runtime/ui"
    "github.com/wwsheng009/mint/runtime/style"
)

// VNode is the component description.
// It contains ONLY declarative information - no state, no closures, no paint logic.
type VNode struct {
    *rtui.ElementVNode

    // === Identification ===
    key string

    // === Visual Props ===
    // ... component-specific props

    // === Event Handling (使用 ActionTargetID，而非闭包) ===
    actionTargetID string

    // === State Props ===
    disabled bool

    // === Box Model ===
    rtui.BoxModelMixin
}

// Ensure VNode implements required interfaces
var (
    _ rtui.VNode           = (*VNode)(nil)
    _ rtui.InstanceFactory = (*VNode)(nil)
    // Add more interfaces as needed
)

// New creates a new VNode
func New(...) *VNode {
    return &VNode{
        ElementVNode: rtui.NewElement("<component>"),
        // Initialize props
    }
}

// CreateInstance creates Instance from VNode description
func (v *VNode) CreateInstance() rtui.ComponentInstance {
    props := rtui.Props{
        // Collect all props (不包括闭包)
        "actionTargetID": v.actionTargetID,
    }
    return NewInstance(props)
}

// SetActionTargetID sets the action target ID for event routing
func (v *VNode) SetActionTargetID(id string) *VNode {
    v.actionTargetID = id
    return v
}

// Builder methods...
```

### Instance 模板

**文件**: `instance.go`

```go
package <component>

import (
    "github.com/wwsheng009/mint/runtime/paint"
    "github.com/wwsheng009/mint/runtime/style"
    rtui "github.com/wwsheng009/mint/runtime/ui"
    "github.com/wwsheng009/mint/ui/components/control"
)

// Instance is the runtime entity
type Instance struct {
    // === Props (from VNode) ===
    // ... props
    actionTargetID string

    // === Runtime State ===
    state  control.InteractionState
    bounds [4]int // x, y, w, h
    dirty  bool

    // === Behaviors ===
    behaviors *control.BehaviorList
}

// Ensure Instance implements required interfaces
var (
    _ rtui.ComponentInstance     = (*Instance)(nil)
    _ rtui.PaintableInstance     = (*Instance)(nil)
    _ rtui.ActionHandlerInstance = (*Instance)(nil)  // ✅ 实现事件处理接口
    // Add more interfaces as needed
)

// NewInstance creates Instance from props
func NewInstance(props rtui.Props) *Instance {
    inst := &Instance{
        actionTargetID: getStringProp(props, "actionTargetID", ""),
        dirty: true,
    }

    inst.initBehaviors()
    return inst
}

// initBehaviors initializes behavior composition
func (inst *Instance) initBehaviors() {
    inst.behaviors = control.NewBehaviorList(
        // Add behaviors
    )
}

// ========== ComponentInstance Interface ==========

func (inst *Instance) Key() string { return inst.key }
func (inst *Instance) SetKey(key string) { inst.key = key }
func (inst *Instance) Init(props rtui.Props) { /* ... */ }
func (inst *Instance) Destroy() { /* ... */ }
func (inst *Instance) OnMount() { /* ... */ }
func (inst *Instance) OnUnmount() { /* ... */ }
func (inst *Instance) SetProps(props rtui.Props) bool { /* ... */ }
func (inst *Instance) GetProps() rtui.Props { return inst.props }
func (inst *Instance) MarkDirty() { inst.dirty = true }
func (inst *Instance) IsDirty() bool { return inst.dirty }
func (inst *Instance) GetContext() *rtui.ComponentContext { return inst.context }

// ========== PaintableInstance Interface ==========

// Paint renders the component
func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
    // 1. Resolve style based on state
    style := inst.resolveStyle()

    // 2. Build display content
    content := inst.buildContent()

    // 3. Return draw commands
    return []paint.DrawCmd{{
        X:     x,
        Y:     y,
        Text:  content,
        Style: style,
    }}
}

// ========== ActionHandlerInstance Interface ==========

// CanHandleAction checks if this instance can handle the action type
func (inst *Instance) CanHandleAction(actionType string) bool {
    // 实现事件类型检查
    return actionType == "click" || actionType == "press"
}

// HandleAction handles the action (替代闭包回调)
func (inst *Instance) HandleAction(actionType string, payload interface{}) bool {
    if inst.state.Disabled {
        return false
    }

    switch actionType {
    case "click", "press":
        // 执行点击逻辑
        // 注意：这里不能直接调用闭包，而是触发状态变化
        inst.state.Pressed = true
        inst.dirty = true
        return true
    }

    return false
}

// resolveStyle resolves visual style based on state
func (inst *Instance) resolveStyle() style.Style {
    s := inst.style

    if inst.state.Disabled {
        s = s.Foreground(theme.DisabledFG())
    } else if inst.state.Focused {
        s = s.Foreground(theme.Focus()).Bold(true)
    } else if inst.state.Hovered {
        s = s.Underline(true)
    }

    return s
}

// buildContent builds the display content
func (inst *Instance) buildContent() string {
    // Component-specific logic
    return ""
}
```

---

## 详细迁移步骤

### Step 5.1: 迁移 Text 组件

**原位置**: `components/text/`
**目标位置**: `ui/components/text/`

#### 5.1.1 创建目录
```bash
mkdir -p ui/components/text
```

#### 5.1.2 创建 vnode.go

```go
package text

import (
    rtui "github.com/wwsheng009/mint/runtime/ui"
    "github.com/wwsheng009/mint/runtime/style"
)

type VNode struct {
    *rtui.ElementVNode
    key     string
    content string
    style   style.Style
    rtui.BoxModelMixin
}

var (
    _ rtui.VNode           = (*VNode)(nil)
    _ rtui.InstanceFactory = (*VNode)(nil)
)

func New(content string) *VNode {
    return &VNode{
        ElementVNode: rtui.NewElement("text"),
        content:      content,
    }
}

func (v *VNode) CreateInstance() rtui.ComponentInstance {
    return NewInstance(rtui.Props{
        "content": v.content,
        "style":   v.style,
    })
}

// Setters...
func (v *VNode) SetContent(content string) *VNode {
    v.content = content
    return v
}

func (v *VNode) SetStyle(s style.Style) *VNode {
    v.style = s
    return v
}
```

#### 5.1.3 创建 instance.go

```go
package text

import (
    "github.com/wwsheng009/mint/runtime/paint"
    "github.com/wwsheng009/mint/runtime/style"
    rtui "github.com/wwsheng009/mint/runtime/ui"
)

type Instance struct {
    key     string
    content string
    style   style.Style
    dirty   bool
}

var (
    _ rtui.ComponentInstance = (*Instance)(nil)
    _ rtui.PaintableInstance = (*Instance)(nil)
)

func NewInstance(props rtui.Props) *Instance {
    return &Instance{
        content: getStringProp(props, "content", ""),
        style:   getStyleProp(props),
        dirty:   true,
    }
}

func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
    if inst.content == "" {
        return nil
    }
    
    return []paint.DrawCmd{{
        X:     x,
        Y:     y,
        Text:  inst.content,
        Style: inst.style,
    }}
}

// Implement other interface methods...
```

#### 5.1.4 创建测试

```go
func TestTextVNode(t *testing.T) {
    vnode := text.New("Hello")
    assert.Equal(t, "Hello", vnode.Content())
}

func TestTextInstance(t *testing.T) {
    inst := text.NewInstance(rtui.Props{
        "content": "Hello",
    })
    
    cmds := inst.Paint(0, 0)
    assert.Len(t, cmds, 1)
    assert.Equal(t, "Hello", cmds[0].Text)
}
```

---

### Step 5.2: 迁移 VStack/HStack 组件

**原位置**: `components/layout/stack.go`
**目标位置**: `ui/components/stack/`

#### 5.2.1 创建目录
```bash
mkdir -p ui/components/stack
```

#### 5.2.2 创建 vnode.go

```go
package stack

import (
    rtui "github.com/wwsheng009/mint/runtime/ui"
    "github.com/wwsheng009/mint/runtime/style"
)

type Direction int

const (
    DirectionVertical Direction = iota
    DirectionHorizontal
)

type VNode struct {
    *rtui.ElementVNode
    key       string
    direction Direction
    children  []rtui.VNode
    gap       int
    align     rtui.Align
    style     style.Style
    rtui.BoxModelMixin
}

var (
    _ rtui.VNode           = (*VNode)(nil)
    _ rtui.InstanceFactory = (*VNode)(nil)
)

func VStack(children ...rtui.VNode) *VNode {
    return &VNode{
        ElementVNode: rtui.NewElement("vstack"),
        direction:    DirectionVertical,
        children:     children,
    }
}

func HStack(children ...rtui.VNode) *VNode {
    return &VNode{
        ElementVNode: rtui.NewElement("hstack"),
        direction:    DirectionHorizontal,
        children:     children,
    }
}

func (v *VNode) CreateInstance() rtui.ComponentInstance {
    return NewInstance(rtui.Props{
        "direction": v.direction,
        "children":  v.children,
        "gap":       v.gap,
        "align":     v.align,
        "style":     v.style,
    })
}

// Setters...
func (v *VNode) Gap(gap int) *VNode {
    v.gap = gap
    return v
}

func (v *VNode) Align(align rtui.Align) *VNode {
    v.align = align
    return v
}
```

#### 5.2.3 创建 instance.go

```go
package stack

import (
    "github.com/wwsheng009/mint/runtime/paint"
    rtui "github.com/wwsheng009/mint/runtime/ui"
)

type Instance struct {
    direction Direction
    children  []rtui.VNode
    gap       int
    align     rtui.Align
    dirty     bool
}

var (
    _ rtui.ComponentInstance = (*Instance)(nil)
    _ rtui.PaintableInstance = (*Instance)(nil)
)

func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
    // Stack 组件本身不产生绘制命令
    // 它的子元素由各自的 Instance.Paint() 绘制
    return nil
}

// Stack 是布局容器，子元素在布局阶段处理
```

---

### Step 5.3: 迁移 Input 组件

**原位置**: `components/input/`
**目标位置**: `ui/components/input/`

#### 5.3.1 创建 vnode.go

```go
package input

import (
    rtui "github.com/wwsheng009/mint/runtime/ui"
    "github.com/wwsheng009/mint/runtime/style"
    "github.com/wwsheng009/mint/runtime/intent"
)

type InputType int

const (
    InputTypeText InputType = iota
    InputTypePassword
    InputTypeNumber
)

type VNode struct {
    *rtui.ElementVNode
    key        string
    inputType  InputType
    value      string
    placeholder string
    maxLength  int
    style      style.Style
    disabled   bool
    changeIntent intent.Intent
    rtui.BoxModelMixin
}

var (
    _ rtui.VNode           = (*VNode)(nil)
    _ rtui.InstanceFactory = (*VNode)(nil)
    _ rtui.FocusableVNode  = (*VNode)(nil)
)

func New() *VNode {
    return &VNode{
        ElementVNode: rtui.NewElement("input"),
        inputType:    InputTypeText,
    }
}

func (v *VNode) CreateInstance() rtui.ComponentInstance {
    return NewInstance(rtui.Props{
        "inputType":   v.inputType,
        "value":       v.value,
        "placeholder": v.placeholder,
        "maxLength":   v.maxLength,
        "style":       v.style,
        "disabled":    v.disabled,
        "changeIntent": v.changeIntent,
    })
}
```

#### 5.3.2 创建 instance.go

```go
package input

import (
    "github.com/wwsheng009/mint/runtime/paint"
    "github.com/wwsheng009/mint/runtime/style"
    rtui "github.com/wwsheng009/mint/runtime/ui"
    "github.com/wwsheng009/mint/ui/components/control"
)

type Instance struct {
    // Props
    inputType   InputType
    value       string
    placeholder string
    maxLength   int
    style       style.Style
    changeIntent intent.Intent
    
    // State
    state  control.InteractionState
    cursor int
    dirty  bool
    
    // Behaviors
    behaviors *control.BehaviorList
}

var (
    _ rtui.ComponentInstance     = (*Instance)(nil)
    _ rtui.PaintableInstance     = (*Instance)(nil)
    _ rtui.FocusableInstance     = (*Instance)(nil)
    _ rtui.ActionHandlerInstance = (*Instance)(nil)
)

func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
    // 1. Resolve style
    style := inst.resolveStyle()
    
    // 2. Build display text
    displayText := inst.buildDisplayText()
    
    // 3. Add cursor if focused
    if inst.state.Focused {
        displayText = inst.addCursor(displayText)
    }
    
    return []paint.DrawCmd{{
        X:     x,
        Y:     y,
        Text:  displayText,
        Style: style,
    }}
}

func (inst *Instance) buildDisplayText() string {
    if inst.value == "" {
        return inst.placeholder
    }
    
    if inst.inputType == InputTypePassword {
        return strings.Repeat("*", len(inst.value))
    }
    
    return inst.value
}

func (inst *Instance) HandleAction(actionType string, payload interface{}) bool {
    switch actionType {
    case "KeyPress":
        return inst.handleKeyPress(payload)
    case "Backspace":
        return inst.handleBackspace()
    }
    return false
}
```

---

## 测试计划

### 单元测试
```bash
# 测试每个组件
go test ./ui/components/text -v
go test ./ui/components/stack -v
go test ./ui/components/input -v
```

### 集成测试
```bash
# 测试组件在应用中的表现
go test ./examples -v
```

### 可视化验证
```bash
# 运行示例应用
cd examples/fiber_counter
go run main.go
```

---

## 验收标准

### 每个组件
- [ ] VNode 只包含声明性属性
- [ ] VNode 不持有闭包（使用 Intent）
- [ ] Instance 实现 PaintableInstance
- [ ] Instance 使用 Behavior 组合
- [ ] 单元测试覆盖所有状态
- [ ] 文档完整

### 整体
- [ ] 所有 P0 组件迁移完成
- [ ] 所有 P1 组件迁移完成
- [ ] 示例应用正常运行
- [ ] 性能无退化

---

## 完成检查清单

### Text 组件
- [ ] vnode.go
- [ ] instance.go
- [ ] builder.go
- [ ] text_test.go
- [ ] README.md

### Stack 组件
- [ ] vnode.go
- [ ] instance.go
- [ ] stack_test.go
- [ ] README.md

### Input 组件
- [ ] vnode.go
- [ ] instance.go
- [ ] input_test.go
- [ ] README.md

### 其他组件
- [ ] Checkbox
- [ ] Select
- [ ] Table
- [ ] Modal

---

**完成**: 所有组件迁移完成后，项目将完全采用 Fiber-first 架构！
