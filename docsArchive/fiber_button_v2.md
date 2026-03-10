# Fiber-first Button 组件设计规范

> 本文档定义了 Fiber-first 架构下 Button 组件的设计要求，包括 VNode、Instance、Measure、Paint 等核心接口。

---

# 一、架构概览

## 1.1 核心分层

```
VNode (描述层)
    ↓ CreateInstance()
Fiber (结构层)
    ↓ 持有 Instance
Instance (运行层)
    ↓ 组合 Behaviors
Behavior (能力层)
    ↓ 发射 Intent
Action System (调度层)
```

## 1.2 设计原则

| 原则 | 说明 |
|------|------|
| VNode 纯声明 | 不包含状态、闭包、渲染逻辑 |
| Instance 纯运行 | 持有状态、处理事件、执行渲染 |
| Fiber 只做结构 | 树结构、调度标记、NodeID |
| 无跨层访问 | Instance 不访问 Fiber.Parent/Sibling |

---

# 二、目录结构

```
ui/components/button/
    vnode.go      # 描述层：ButtonVNode、Props、Builder
    instance.go   # 运行层：ButtonInstance、Measure、Paint、Behaviors
    builder.go    # 构建器：流畅 API
    types.go      # 类型定义：Variant、Size、FocusStyle
```

---

# 三、VNode 设计（描述层）

## 3.1 职责

- ✅ 定义组件 Props
- ✅ 实现 `InstanceFactory` 接口
- ✅ 实现 `FocusableVNode` 接口
- ✅ 实现 `rtui.VNode` 接口
- ❌ 不包含状态
- ❌ 不包含闭包（如 `OnClick(func())`)
- ❌ 不包含 Paint 逻辑

## 3.2 VNode 结构

```go
// vnode.go
package button

type VNode struct {
    *rtui.ElementVNode

    // === Identification ===
    key string

    // === Visual Props ===
    label      string
    variant    Variant
    size       Size
    focusStyle FocusStyle
    style      style.Style

    // === Layout Props ===
    padding   [4]int // top, right, bottom, left
    textAlign rtui.Align

    // === Intent Props (no closures!) ===
    pressIntent intent.Intent

    // === State Props ===
    disabled bool

    // === Focus state (transient, for rendering) ===
    hasFocus bool

    // === Box Model ===
    rtui.BoxModelMixin
}
```

## 3.3 必须实现的接口

```go
var (
    _ rtui.VNode           = (*VNode)(nil)
    _ rtui.InstanceFactory = (*VNode)(nil)
    _ rtui.FocusableVNode  = (*VNode)(nil)
    _ rtui.BoxModel        = (*VNode)(nil)
)
```

## 3.4 InstanceFactory 实现

```go
func (b *VNode) CreateInstance() rtui.ComponentInstance {
    props := rtui.Props{
        "key":         b.key,
        "label":       b.label,
        "variant":     b.variant,
        "size":        b.size,
        "focusStyle":  b.focusStyle,
        "style":       b.style,
        "pressIntent": b.pressIntent,
        "disabled":    b.disabled,
        "padding":     b.Padding(),
        "textAlign":   b.TextAlign(),
    }
    return NewInstance(props)
}
```

---

# 四、Instance 设计（运行层）⭐核心

## 4.1 职责

- ✅ 持有运行期状态
- ✅ 实现 `Measure` 方法（两段布局）
- ✅ 实现 `Paint` 方法
- ✅ 组合 Behaviors
- ✅ 处理 Action

## 4.2 Instance 结构

```go
// instance.go
package button

type Instance struct {
    // === Identification ===
    key string

    // === Props (from VNode, may change each render) ===
    label       string
    variant     Variant
    size        Size
    focusStyle  FocusStyle
    buttonStyle style.Style
    pressIntent intent.Intent
    padding     [4]int
    textAlign   rtui.Align

    // === Runtime State (managed by instance) ===
    state   control.InteractionState
    bounds  [4]int // x, y, w, h
    dirty   bool

    // === Intent Emitter ===
    intentEmitter func(intent.Intent)

    // === Behaviors ===
    behaviors *control.BehaviorList
}
```

## 4.3 必须实现的接口

```go
var (
    _ rtui.ComponentInstance     = (*Instance)(nil)
    _ rtui.PaintableInstance     = (*Instance)(nil)
    _ rtui.FocusableInstance     = (*Instance)(nil)
    _ rtui.ActionHandlerInstance = (*Instance)(nil)
    _ control.Instance           = (*Instance)(nil)
    // ⭐ 两段布局：必须实现 Measure
    _ interface {
        Measure(layout.Constraints) layout.Size
    } = (*Instance)(nil)
)
```

---

# 五、两段测量机制（Two-Pass Layout）⭐核心

## 5.1 概述

Fiber-first 架构采用**两段布局**：

| 阶段 | 方法 | 职责 |
|------|------|------|
| Pass 1 | `Measure(constraints)` | 计算理想尺寸，不设置位置 |
| Pass 2 | `SetPosition(x, y)` / `SetBounds()` | 设置最终位置和尺寸 |

## 5.2 Measure 方法实现

```go
// Measure 实现 layout.Measurable 接口
// 计算按钮在给定约束下的理想尺寸
// 这是两段布局的第一阶段：测量自然尺寸，不设置位置
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
    if inst == nil {
        return layout.Size{}
    }

    // 1. 计算内容宽度：label + brackets + focus indicator
    label := inst.label
    if label == "" {
        label = " " // 空按钮仍有最小宽度
    }

    // Width: label长度 + 2(括号[]) + 1(焦点指示器)
    contentWidth := utf8.RuneCountInString(label) + 3

    // Height: 单行按钮高度始终为 1
    contentHeight := 1

    // 2. 应用 Size 修饰符
    switch inst.size {
    case SizeSmall:
        // Small: 无额外 padding
    case SizeMedium:
        // Medium: 两侧各 +1
        contentWidth += 2
    case SizeLarge:
        // Large: 两侧各 +2
        contentWidth += 4
    }

    // 3. 应用用户指定的 padding
    horizontalPadding := inst.padding[1] + inst.padding[3] // right + left
    verticalPadding := inst.padding[0] + inst.padding[2]   // top + bottom

    width := contentWidth + horizontalPadding
    height := contentHeight + verticalPadding

    // 4. 应用约束
    width = constraints.ConstrainWidth(width)
    height = constraints.ConstrainHeight(height)

    // 5. 应用显式 style 尺寸（覆盖计算值）
    if inst.buttonStyle.Width > 0 {
        width = constraints.ConstrainWidth(inst.buttonStyle.Width)
    }
    if inst.buttonStyle.Height > 0 {
        height = constraints.ConstrainHeight(inst.buttonStyle.Height)
    }

    return layout.Size{Width: width, Height: height}
}
```

## 5.3 GetNaturalSize 辅助方法

```go
// GetNaturalSize 返回按钮的自然（无约束）尺寸
// 用于当按钮被拉伸时的对齐计算
func (inst *Instance) GetNaturalSize() (width, height int) {
    label := inst.label
    if label == "" {
        label = " "
    }

    width = utf8.RuneCountInString(label) + 3 // label + brackets + focus
    height = 1

    switch inst.size {
    case SizeSmall:
        // no extra padding
    case SizeMedium:
        width += 2
    case SizeLarge:
        width += 4
    }

    return width, height
}
```

## 5.4 尺寸计算规则

| Size | 基础宽度 | 公式 |
|------|----------|------|
| Small | `len(label) + 3` | 最紧凑 |
| Medium | `len(label) + 3 + 2` | 默认 |
| Large | `len(label) + 3 + 4` | 更宽松 |

**示例**：`label = "Click Me"` (8 chars)

| Size | 计算 | 宽度 |
|------|------|------|
| Small | 8 + 3 | 11 |
| Medium | 8 + 3 + 2 | 13 |
| Large | 8 + 3 + 4 | 15 |

---

# 六、FiberToNodeAdapter 集成

## 6.1 适配器职责

`FiberToNodeAdapter` 将 Fiber 树适配为 `layout.Node` 接口，供布局引擎使用。

## 6.2 Measure 调用链

```
layout.Engine
    ↓ 调用 Node.Measure()
FiberToNodeAdapter.Measure()
    ↓ 检查 fiber.Instance
Instance.Measure()
    ↓ 返回计算尺寸
layout.Size
```

## 6.3 适配器 Measure 实现

```go
// fiber_adapter.go
func (a *FiberToNodeAdapter) Measure(constraints layout.Constraints) layout.Size {
    if a.fiber == nil {
        return layout.Size{}
    }

    // 0. 特殊处理文本节点
    if a.fiber.Type == rtui.VNodeText {
        textContent := a.GetTextContent()
        if textContent != "" {
            return layout.Size{
                Width:  constraints.ConstrainWidth(len(textContent)),
                Height: constraints.ConstrainHeight(1),
            }
        }
    }

    // 1. 尝试从 Instance 获取尺寸（⭐ 优先级最高）
    if a.fiber.Instance != nil {
        if measurable, ok := a.fiber.Instance.(interface {
            Measure(layout.Constraints) layout.Size
        }); ok {
            return measurable.Measure(constraints)
        }
    }

    // 2. 从 Style 获取固定尺寸
    if a.fiber.Style.Width > 0 || a.fiber.Style.Height > 0 {
        return layout.Size{
            Width:  constraints.ConstrainWidth(a.fiber.Style.Width),
            Height: constraints.ConstrainHeight(a.fiber.Style.Height),
        }
    }

    // 3. 默认值
    return layout.Size{}
}
```

---

# 七、Behavior 层设计

## 7.1 能力组合模式

Button 不直接实现所有交互逻辑，而是组合标准 Behaviors：

```go
func (inst *Instance) initBehaviors() {
    pressable := control.NewPressableBehavior(inst.pressIntent)

    inst.behaviors = control.NewBehaviorList(
        &control.FocusableBehavior{},
        pressable,
        &control.HoverableBehavior{},
        &control.DisableableBehavior{},
    )
}
```

## 7.2 标准 Behaviors

| Behavior | 职责 |
|----------|------|
| FocusableBehavior | 焦点进入/离开 |
| PressableBehavior | 按压发射 Intent |
| HoverableBehavior | 鼠标悬停 |
| DisableableBehavior | 禁用状态 |

## 7.3 好处

- ✅ 行为复用（Checkbox、MenuItem 等共享）
- ✅ 单一职责
- ✅ 可插拔
- ✅ 可测试

---

# 八、Paint 设计

## 8.1 Paint 方法

```go
func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
    // 1. 构建 label 文本
    labelText := inst.buildLabelText()

    // 2. 解析样式
    buttonStyle := inst.resolveStyle()

    // 3. 构建 button 文本（含对齐）
    buttonText := inst.buildButtonText(labelText, buttonStyle)

    return []paint.DrawCmd{{
        X:     x,
        Y:     y,
        Text:  buttonText,
        Style: buttonStyle,
    }}
}
```

## 8.2 样式解析

```go
func (inst *Instance) resolveStyle() style.Style {
    s := inst.buttonStyle

    // 1. Variant 样式
    if s.FG == "" && s.BG == "" {
        switch inst.variant {
        case VariantPrimary:
            s = s.Foreground(theme.BG()).Background(theme.Primary()).Bold(true)
        case VariantDanger:
            s = s.Foreground(theme.BG()).Background(theme.Error()).Bold(true)
        // ...
        }
    }

    // 2. State 样式（优先级：Disabled > Focused > Hovered）
    if inst.state.Disabled {
        s = s.Foreground(theme.DisabledFG()).Background(theme.DisabledBG())
    } else if inst.state.Focused {
        switch inst.focusStyle {
        case FocusStyleReverse:
            s = s.Foreground(theme.Foreground()).Background(theme.Focus()).Bold(true)
        case FocusStyleUnderline:
            s = s.Foreground(theme.FocusBright()).Underline(true).Bold(true)
        // ...
        }
    }

    return s
}
```

---

# 九、旧 vs 新 Button 对比

| 功能 | 旧 `components/button` | 新 `ui/components/button` |
|------|------------------------|---------------------------|
| Measure 位置 | VNode 上 | Instance 上 ⭐ |
| 约束类型 | `runtime.BoxConstraints` | `layout.Constraints` |
| 尺寸类型 | `runtime.Size` | `layout.Size` |
| 两段测量 | ✅ | ✅ |
| InstanceFactory | ✅ | ✅ |
| Behavior 层 | 部分 | 完整 |
| Intent 模式 | 部分 | 完整（无闭包） |

---

# 十、完整实现示例

## 10.1 文件：vnode.go

```go
package button

import (
    "github.com/wwsheng009/mint/runtime/intent"
    "github.com/wwsheng009/mint/runtime/style"
    rtui "github.com/wwsheng009/mint/runtime/ui"
)

type VNode struct {
    *rtui.ElementVNode
    key         string
    label       string
    variant     Variant
    size        Size
    focusStyle  FocusStyle
    style       style.Style
    padding     [4]int
    textAlign   rtui.Align
    pressIntent intent.Intent
    disabled    bool
    hasFocus    bool
    rtui.BoxModelMixin
}

func New(label string) *VNode {
    return &VNode{
        ElementVNode: rtui.NewElement("button"),
        label:        label,
        variant:      VariantDefault,
        size:         SizeMedium,
        focusStyle:   FocusStyleReverse,
    }
}

func (b *VNode) CreateInstance() rtui.ComponentInstance {
    return NewInstance(rtui.Props{
        "key":         b.key,
        "label":       b.label,
        "variant":     b.variant,
        "size":        b.size,
        "focusStyle":  b.focusStyle,
        "style":       b.style,
        "pressIntent": b.pressIntent,
        "disabled":    b.disabled,
        "padding":     b.Padding(),
        "textAlign":   b.TextAlign(),
    })
}

// ... 其他方法
```

## 10.2 文件：instance.go

```go
package button

import (
    "strings"
    "unicode/utf8"

    "github.com/wwsheng009/mint/framework/theme"
    "github.com/wwsheng009/mint/runtime/intent"
    "github.com/wwsheng009/mint/runtime/layout"
    "github.com/wwsheng009/mint/runtime/paint"
    "github.com/wwsheng009/mint/runtime/style"
    rtui "github.com/wwsheng009/mint/runtime/ui"
    "github.com/wwsheng009/mint/ui/components/control"
)

type Instance struct {
    key          string
    label        string
    variant      Variant
    size         Size
    focusStyle   FocusStyle
    buttonStyle  style.Style
    pressIntent  intent.Intent
    padding      [4]int
    textAlign    rtui.Align
    state        control.InteractionState
    bounds       [4]int
    dirty        bool
    intentEmitter func(intent.Intent)
    behaviors    *control.BehaviorList
}

var (
    _ rtui.ComponentInstance     = (*Instance)(nil)
    _ rtui.PaintableInstance     = (*Instance)(nil)
    _ rtui.FocusableInstance     = (*Instance)(nil)
    _ rtui.ActionHandlerInstance = (*Instance)(nil)
    _ control.Instance           = (*Instance)(nil)
    _ interface {
        Measure(layout.Constraints) layout.Size
    } = (*Instance)(nil)
)

func NewInstance(props rtui.Props) *Instance {
    inst := &Instance{
        key:         getStringProp(props, "key", ""),
        label:       getStringProp(props, "label", ""),
        variant:     getVariantProp(props, VariantDefault),
        size:        getSizeProp(props, SizeMedium),
        focusStyle:  getFocusStyleProp(props, FocusStyleReverse),
        buttonStyle: getStyleProp(props),
        pressIntent: getIntentProp(props),
        padding:     getPaddingProp(props),
        textAlign:   getTextAlignProp(props, rtui.AlignStart),
        dirty:       true,
    }

    inst.state = control.InteractionState{
        Disabled: getBoolProp(props, "disabled", false),
    }

    inst.initBehaviors()
    return inst
}

func (inst *Instance) initBehaviors() {
    pressable := control.NewPressableBehavior(inst.pressIntent)
    inst.behaviors = control.NewBehaviorList(
        &control.FocusableBehavior{},
        pressable,
        &control.HoverableBehavior{},
        &control.DisableableBehavior{},
    )
}

// ⭐ Measure - 两段布局第一阶段
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
    if inst == nil {
        return layout.Size{}
    }

    label := inst.label
    if label == "" {
        label = " "
    }

    contentWidth := utf8.RuneCountInString(label) + 3
    contentHeight := 1

    switch inst.size {
    case SizeSmall:
    case SizeMedium:
        contentWidth += 2
    case SizeLarge:
        contentWidth += 4
    }

    horizontalPadding := inst.padding[1] + inst.padding[3]
    verticalPadding := inst.padding[0] + inst.padding[2]

    width := constraints.ConstrainWidth(contentWidth + horizontalPadding)
    height := constraints.ConstrainHeight(contentHeight + verticalPadding)

    if inst.buttonStyle.Width > 0 {
        width = constraints.ConstrainWidth(inst.buttonStyle.Width)
    }
    if inst.buttonStyle.Height > 0 {
        height = constraints.ConstrainHeight(inst.buttonStyle.Height)
    }

    return layout.Size{Width: width, Height: height}
}

// Paint - 渲染
func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
    labelText := inst.buildLabelText()
    buttonStyle := inst.resolveStyle()
    buttonText := inst.buildButtonText(labelText, buttonStyle)

    return []paint.DrawCmd{{
        X:     x,
        Y:     y,
        Text:  buttonText,
        Style: buttonStyle,
    }}
}

// ... 其他方法
```

---

# 十一、测试验证

## 11.1 单元测试

```go
func TestButtonInstance_Measure(t *testing.T) {
    tests := []struct {
        name       string
        label      string
        size       Size
        wantWidth  int
    }{
        {"Small", "OK", SizeSmall, 5},      // 2 + 3 = 5
        {"Medium", "OK", SizeMedium, 7},    // 2 + 3 + 2 = 7
        {"Large", "OK", SizeLarge, 9},      // 2 + 3 + 4 = 9
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            btn := New(tt.label).SetSize(tt.size)
            fiber := rtui.CreateFiberFromVNode(btn)

            inst := fiber.Instance.(*Instance)
            size := inst.Measure(layout.UnboundedConstraints())

            if size.Width != tt.wantWidth {
                t.Errorf("Width = %d, want %d", size.Width, tt.wantWidth)
            }
            if size.Height != 1 {
                t.Errorf("Height = %d, want 1", size.Height)
            }
        })
    }
}
```

## 11.2 Fiber 集成测试

```go
func TestFiberToNodeAdapter_ButtonMeasure(t *testing.T) {
    btn := New("Click Me")
    fiber := rtui.CreateFiberFromVNode(btn)

    adapter := NewFiberToNodeAdapterPure(fiber)
    constraints := layout.Constraints{MaxWidth: 100, MaxHeight: 10}

    size := adapter.Measure(constraints)

    // "Click Me" = 8 + 3 + 2 (Medium) = 13
    if size.Width != 13 {
        t.Errorf("Width = %d, want 13", size.Width)
    }
    if size.Height != 1 {
        t.Errorf("Height = %d, want 1", size.Height)
    }
}
```

---

# 十二、设计检查清单

## ✅ VNode 检查

- [ ] 只包含声明性 Props
- [ ] 实现 `InstanceFactory.CreateInstance()`
- [ ] 实现 `FocusableVNode` 接口
- [ ] 无闭包（OnClick 等）
- [ ] 无状态

## ✅ Instance 检查

- [ ] 实现 `Measure(layout.Constraints) layout.Size` ⭐
- [ ] 实现 `Paint(x, y int) []paint.DrawCmd`
- [ ] 实现 `FocusableInstance` 接口
- [ ] 组合标准 Behaviors
- [ ] 无 Fiber 结构访问

## ✅ 两段布局检查

- [ ] Measure 只计算尺寸，不设置位置
- [ ] Measure 应用约束 `ConstrainWidth/Height`
- [ ] Measure 支持 style 覆盖
- [ ] 支持 Size 修饰符
- [ ] 支持 padding

---

# 十三、常见设计错误

| ❌ 错误 | ✅ 正确 |
|--------|--------|
| Button 直接持 `OnClick(func())` | Button 发射 `PressIntent` |
| Instance 访问 `fiber.Parent` | Instance 只访问自己状态 |
| Measure 设置位置 | Measure 只返回尺寸 |
| Paint 依赖 VNode | Paint 只用 Instance 数据 |
| 每个组件写自己的状态机 | 使用统一 `InteractionState` |

---

# 十四、总结

Button 是 Fiber-first Runtime 的基础组件，其设计直接影响：

1. **事件系统** - Intent vs Closure
2. **生命周期** - Instance 持久化
3. **布局系统** - 两段测量
4. **可扩展性** - Behavior 组合

正确的 Button 设计是：

```
VNode (声明)
    ↓
Fiber (结构)
    ↓
Instance (运行)
    ├── Measure() → Size
    ├── Paint() → DrawCmd
    └── Behaviors → Intent
```

---

# 附录：相关文档

- [Fiber-first 渲染管线](./FIBER_FIRST_RENDER_PIPELINE.md)
- [两段布局引擎](../../layout/TWO_PASS_LAYOUT.md)
- [Action Runtime](./fiber_action.md)
- [统一交互状态机](./INTERACTION_STATE_MACHINE.md)
