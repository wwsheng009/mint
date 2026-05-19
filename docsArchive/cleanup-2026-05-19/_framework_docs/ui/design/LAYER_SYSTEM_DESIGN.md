# Layer 系统设计文档

**版本**: v1.0
**日期**: 2026-01-31
**来源**: idea/idea4.3_modal.md
**状态**: 🟡 中优先级

---

## 一、概述

### 1.1 设计目标

实现一个**视觉层级管理系统**，用于处理需要脱离正常布局流的高级 UI 组件：

- Modal（模态框）
- Tooltip（提示框）
- Dropdown（下拉菜单）
- Context Menu（上下文菜单）
- Toast（通知）
- Overlay（覆盖层）

### 1.2 核心问题

**问题**: 在声明式 UI 中，如何让一个组件"浮"在其他所有内容之上？

**传统方案** (不推荐):
```go
// ❌ 在组件树中放置 Modal
VStack(
    Text("Content"),
    Modal(),  // 会参与布局，影响父容器
)
```

**正确方案**:
```go
// ✅ Modal 独立于主布局
ui.Layer(ui.LayerModal, ModalContent())
ui.VStack(
    Text("Content"),  // 不受 Modal 影响
)
```

---

## 二、架构设计

### 2.1 Layer 定义

```go
// layer/layer.go

package layer

// Layer 视觉层级
type Layer int

const (
    // LayerBase 基础层（默认内容）
    LayerBase Layer = iota

    // LayerOverlay 覆盖层（下拉菜单、Popover）
    LayerOverlay

    // LayerModal 模态框层
    LayerModal

    // LayerTooltip 提示框层
    LayerTooltip

    // LayerNotification 通知层
    LayerNotification
)

// String 返回层级名称
func (l Layer) String() string {
    switch l {
    case LayerBase:
        return "base"
    case LayerOverlay:
        return "overlay"
    case LayerModal:
        return "modal"
    case LayerTooltip:
        return "tooltip"
    case LayerNotification:
        return "notification"
    default:
        return "unknown"
    }
}

// ZIndex 返回 z-index（用于调试）
func (l Layer) ZIndex() int {
    return int(l)
}

// IsValid 检查层级是否有效
func (l Layer) IsValid() bool {
    return l >= LayerBase && l <= LayerNotification
}
```

### 2.2 Layer 树结构

```go
// layer/tree.go

package layer

import "github.com/wwsheng009/mint/ui"

// LayerNode 层级节点
type LayerNode struct {
    Layer   Layer
    ID      string
    Content ui.VNode
    Visible bool
    FocusID string
}

// LayerTree 层级树
type LayerTree struct {
    nodes map[Layer][]*LayerNode
    focus *FocusManager
}

// NewLayerTree 创建层级树
func NewLayerTree() *LayerTree {
    return &LayerTree{
        nodes: make(map[Layer][]*LayerNode),
        focus: NewFocusManager(),
    }
}

// Add 添加节点到指定层级
func (t *LayerTree) Add(layer Layer, id string, content ui.VNode) {
    if _, ok := t.nodes[layer]; !ok {
        t.nodes[layer] = make([]*LayerNode, 0)
    }

    node := &LayerNode{
        Layer:   layer,
        ID:      id,
        Content: content,
        Visible: true,
    }

    t.nodes[layer] = append(t.nodes[layer], node)
}

// Remove 从层级中移除节点
func (t *LayerTree) Remove(layer Layer, id string) bool {
    nodes := t.nodes[layer]
    for i, node := range nodes {
        if node.ID == id {
            // 复制切片，避免修改原数组
            t.nodes[layer] = append(nodes[:i], nodes[i+1:]...)
            return true
        }
    }
    return false
}

// Get 获取指定层级的所有节点
func (t *LayerTree) Get(layer Layer) []*LayerNode {
    if nodes, ok := t.nodes[layer]; ok {
        return nodes
    }
    return nil
}

// GetAll 按层级顺序获取所有节点
func (t *LayerTree) GetAll() []*LayerNode {
    result := make([]*LayerNode, 0)

    // 按层级顺序收集
    for l := LayerBase; l <= LayerNotification; l++ {
        result = append(result, t.nodes[l]...)
    }

    return result
}
```

---

## 三、Layer Manager

### 3.1 核心接口

```go
// layer/manager.go

package layer

import (
    "github.com/wwsheng009/mint/ui"
    "github.com/wwsheng009/mint/runtime/event"
)

// Manager 层级管理器
type Manager struct {
    tree     *LayerTree
    active   Layer  // 当前活跃层级
    modalStack []Layer // 模态框栈（支持嵌套）
}

// NewManager 创建层级管理器
func NewManager() *Manager {
    return &Manager{
        tree:     NewLayerTree(),
        active:   LayerBase,
        modalStack: make([]Layer, 0),
    }
}

// Show 显示指定层级的内容
func (m *Manager) Show(layer Layer, id string, content ui.VNode) {
    m.tree.Add(layer, id, content)

    if layer > m.active {
        m.active = layer
    }
}

// Hide 隐藏指定层级的内容
func (m *Manager) Hide(layer Layer, id string) {
    m.tree.Remove(layer, id)
    m.recalculateActive()
}

// Update 更新指定层级的内容
func (m *Manager) Update(layer Layer, id string, content ui.VNode) {
    m.Hide(layer, id)
    m.Show(layer, id, content)
}

// PushModal 推入模态框（支持嵌套）
func (m *Manager) PushModal(id string, content ui.VNode) {
    m.modalStack = append(m.modalStack, LayerModal)
    m.Show(LayerModal, id, content)
}

// PopModal 弹出模态框
func (m *Manager) PopModal() bool {
    if len(m.modalStack) == 0 {
        return false
    }

    // 移除当前模态框
    m.modalStack = m.modalStack[:len(m.modalStack)-1]
    m.recalculateActive()

    return true
}

// IsModalActive 是否有活跃的模态框
func (m *Manager) IsModalActive() bool {
    return len(m.modalStack) > 0
}

// recalculateActive 重新计算活跃层级
func (m *Manager) recalculateActive() {
    m.active = LayerBase

    for l := LayerModal; l >= LayerBase; l-- {
        if len(m.tree.Get(l)) > 0 {
            m.active = l
            break
        }
    }
}

// ActiveLayer 返回当前活跃层级
func (m *Manager) ActiveLayer() Layer {
    return m.active
}

// ShouldBlockInput 检查是否应该阻止输入到底层
func (m *Manager) ShouldBlockInput() bool {
    return m.active >= LayerModal
}
```

### 3.2 事件处理

```go
// layer/event.go

package layer

import "github.com/wwsheng009/mint/runtime/event"

// HandleEvent 处理事件
func (m *Manager) HandleEvent(e event.Event) bool {
    // 从最顶层开始处理事件
    for l := LayerNotification; l >= LayerBase; l-- {
        nodes := m.tree.Get(l)

        for _, node := range nodes {
            if !node.Visible {
                continue
            }

            // 尝试处理事件
            if m.dispatchEvent(node, e) {
                // 事件被处理，停止传播
                if e.IsPropagationStopped() {
                    return true
                }
            }
        }

        // Modal 层阻止事件向下传播
        if l == LayerModal && len(m.tree.Get(l)) > 0 {
            // 特殊处理：ESC 键关闭 Modal
            if e.Type() == event.EventKeyPress {
                ke := e.(*event.KeyEvent)
                if ke.Key == '\x1b' { // ESC
                    m.PopModal()
                    return true
                }
            }
            return true // 阻止事件继续向下
        }
    }

    return false
}

// dispatchEvent 分发事件到节点
func (m *Manager) dispatchEvent(node *LayerNode, e event.Event) bool {
    // 在节点上执行命中测试
    // ...

    return false
}
```

---

## 四、渲染集成

### 4.1 渲染流程

```go
// layer/render.go

package layer

import (
    "github.com/wwsheng009/mint/framework/render"
    "github.com/wwsheng009/mint/runtime/paint"
)

// Render 渲染所有层级
func (m *Manager) Render(buffer *paint.Buffer) error {
    // 按层级顺序渲染
    for l := LayerBase; l <= LayerNotification; l++ {
        nodes := m.tree.Get(l)

        for _, node := range nodes {
            if !node.Visible {
                continue
            }

            // 渲染节点内容
            if err := m.renderNode(node, buffer); err != nil {
                return err
            }
        }
    }

    return nil
}

// renderNode 渲染单个节点
func (m *Manager) renderNode(node *LayerNode, buffer *paint.Buffer) error {
    // 将 VNode 转换为 DrawCmd
    cmds := render.VNodeToDrawCmds(node.Content)

    // 执行绘制命令
    for _, cmd := range cmds {
        render.ExecuteDrawCmd(cmd, buffer)
    }

    return nil
}
```

### 4.2 布局处理

```go
// layer/layout.go

package layer

import (
    "github.com/wwsheng009/mint/runtime/layout"
)

// Layout 布局所有层级
func (m *Manager) Layout(width, height int) error {
    // Base 层使用全屏布局
    m.layoutLayer(LayerBase, 0, 0, width, height)

    // Modal 层独立布局（居中）
    if nodes := m.tree.Get(LayerModal); len(nodes) > 0 {
        m.layoutModal(nodes, width, height)
    }

    // Tooltip 紧跟目标位置
    m.layoutTooltips(width, height)

    return nil
}

// layoutModal 布局模态框
func (m *Manager) layoutModal(nodes []*LayerNode, width, height int) {
    for _, node := range nodes {
        // 计算模态框尺寸
        contentSize := m.measureContent(node.Content)

        // 居中显示
        modalWidth := min(contentSize.Width + 4, width-4)
        modalHeight := min(contentSize.Height + 4, height-4)

        x := (width - modalWidth) / 2
        y := (height - modalHeight) / 2

        // 设置布局位置
        m.setNodeLayout(node, x, y, modalWidth, modalHeight)
    }
}

// layoutTooltips 布局提示框
func (m *Manager) layoutTooltips(width, height int) {
    nodes := m.tree.Get(LayerTooltip)

    for _, node := range nodes {
        // 根据目标位置布局
        // ...
    }
}
```

---

## 五、UI 集成

### 5.1 声明式 API

```go
// ui/layer.go

package ui

import "github.com/wwsheng009/mint/framework/layer"

// 全局层级管理器
var globalLayerManager = layer.NewManager()

// Layer 在指定层级显示内容
func Layer(l layer.Layer, id string, content VNode) VNode {
    globalLayerManager.Show(l, id, content)
    return content
}

// Modal 显示模态框
func Modal(id string, content VNode) VNode {
    globalLayerManager.PushModal(id, content)
    return content
}

// CloseModal 关闭模态框
func CloseModal() {
    globalLayerManager.PopModal()
}

// Tooltip 显示提示框
func Tooltip(id string, content VNode) VNode {
    globalLayerManager.Show(layer.LayerTooltip, id, content)
    return content
}

// Toast 显示通知
func Toast(id string, content VNode) VNode {
    globalLayerManager.Show(layer.LayerNotification, id, content)
    return content
}
```

### 5.2 使用示例

```go
// 示例 1: 模态框
func App() VNode {
    return ui.VStack(
        ui.Text("Main Content"),
        ui.Button("Open Modal").OnClick(func() {
            ui.Modal("my-modal", ModalContent())
        }),
    )
}

func ModalContent() VNode {
    return ui.Box().Border(true).Padding(2).Child(
        ui.VStack(
            ui.Text("Modal Title").Bold(true),
            ui.Separator(),
            ui.Text("Modal content goes here"),
            ui.HStack(
                ui.Button("Cancel").OnClick(func() {
                    ui.CloseModal()
                }),
                ui.Button("OK"),
            ),
        ),
    )
}

// 示例 2: Tooltip
func WithTooltip() VNode {
    return ui.Text("Hover me").OnMouseEnter(func() {
        ui.Tooltip("tip-1", ui.Text("This is a tooltip"))
    })
}

// 示例 3: Toast
func ShowToast(message string) {
    ui.Toast("toast-"+message, ui.Text(message))
}
```

---

## 六、焦点管理

### 6.1 Focus Trap

```go
// layer/focus.go

package layer

import "github.com/wwsheng009/mint/runtime/focus"

// FocusTrap 焦点陷阱（用于 Modal）
type FocusTrap struct {
    layer  Layer
    id     string
    nodes  []string // 可聚焦的节点 ID
    index int       // 当前聚焦索引
}

// NewFocusTrap 创建焦点陷阱
func NewFocusTrap(layer Layer, id string) *FocusTrap {
    return &FocusTrap{
        layer: layer,
        id:    id,
        nodes: make([]string, 0),
    }
}

// AddNode 添加可聚焦节点
func (t *FocusTrap) AddNode(id string) {
    t.nodes = append(t.nodes, id)
}

// Activate 激活焦点陷阱
func (t *FocusTrap) Activate() {
    if len(t.nodes) > 0 {
        // 聚焦第一个节点
        focus.SetFocus(t.nodes[0])
        t.index = 0
    }
}

// Next 聚焦下一个节点
func (t *FocusTrap) Next() bool {
    if t.index >= len(t.nodes)-1 {
        return false
    }

    t.index++
    focus.SetFocus(t.nodes[t.index])
    return true
}

// Previous 聚焦上一个节点
func (t *FocusTrap) Previous() bool {
    if t.index <= 0 {
        return false
    }

    t.index--
    focus.SetFocus(t.nodes[t.index])
    return true
}
```

### 6.2 TAB 键处理

```go
// layer/tab.go

package layer

// HandleTab 处理 TAB 键
func (m *Manager) HandleTab(forward bool) bool {
    // 检查当前活跃层级
    active := m.ActiveLayer()

    if active < LayerModal {
        // 没有 Modal 层活跃，正常处理
        return false
    }

    // Modal 层活跃，在 Modal 内循环
    nodes := m.tree.Get(active)
    for _, node := range nodes {
        if trap, ok := node.Metadata["focusTrap"].(*FocusTrap); ok {
            if forward {
                trap.Next()
            } else {
                trap.Previous()
            }
            return true
        }
    }

    return false
}
```

---

## 七、实施计划

### 阶段 1: 基础实现

- [ ] 实现 Layer 类型定义
- [ ] 实现 LayerTree
- [ ] 实现 Manager 基础功能

### 阶段 2: 渲染集成

- [ ] 实现层级渲染
- [ ] 实现独立布局
- [ ] 集成到主渲染循环

### 阶段 3: 焦点管理

- [ ] 实现 FocusTrap
- [ ] 实现 TAB 键处理
- [ ] 实现 ESC 关闭

### 阶段 4: UI 集成

- [ ] 实现 ui.Layer()
- [ ] 实现 ui.Modal()
- [ ] 实现其他便捷函数

---

## 八、测试策略

```go
// layer/manager_test.go

func TestLayerShowHide(t *testing.T) {
    m := NewManager()

    content := ui.Text("Test")
    m.Show(LayerModal, "test-modal", content)

    nodes := m.tree.Get(LayerModal)
    assert.Equal(t, 1, len(nodes))
    assert.Equal(t, "test-modal", nodes[0].ID)

    m.Hide(LayerModal, "test-modal")

    nodes = m.tree.Get(LayerModal)
    assert.Equal(t, 0, len(nodes))
}

func TestModalStack(t *testing.T) {
    m := NewManager()

    m.PushModal("modal1", ui.Text("Modal 1"))
    assert.True(t, m.IsModalActive())

    m.PushModal("modal2", ui.Text("Modal 2"))
    assert.Equal(t, 2, len(m.modalStack))

    m.PopModal()
    assert.True(t, m.IsModalActive())

    m.PopModal()
    assert.False(t, m.IsModalActive())
}

func TestEventBlocking(t *testing.T) {
    m := NewManager()

    // 没有 Modal 时，不阻止
    assert.False(t, m.ShouldBlockInput())

    m.PushModal("modal", ui.Text("Modal"))
    assert.True(t, m.ShouldBlockInput())

    m.PopModal()
    assert.False(t, m.ShouldBlockInput())
}
```

---

**文档版本**: v1.0
**最后更新**: 2026-01-31
