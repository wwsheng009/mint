# 事件系统重构实施方案

> **版本**: v1.0
> **日期**: 2025-02-10
> **状态**: 待实施

---

## Phase 0: 前置修复（立即，1-2 天）✅

### 目标
修复 Inspector 立即可用的快速修复

### 已完成项

```go
// ✅ runtime/ui/element.go
type ElementVNode struct {
    bounds [4]int  // 新增字段
}

func (e *ElementVNode) SetBounds(x, y, width, height int) {
    e.bounds = [4]int{x, y, width, height}
}

// ✅ internal/inspector/inspector.go
func vnodeContains(vnode rtui.VNode, x, y int) bool {
    // 支持两种签名
    if boundsAware, ok := vnode.(interface{ GetBounds() [4]int }); ok {
        bounds := boundsAware.GetBounds()
        vx, vy, vw, vh := bounds[0], bounds[1], bounds[2], bounds[3]
        return x >= vx && x < vx+vw && y >= vy && y < vy+vh
    }

    if boundsAware, ok := vnode.(interface{ GetBounds() (int, int, int, int) }); ok {
        vx, vy, vw, vh := boundsAware.GetBounds()
        return x >= vx && x < vx+vw && y >= vy && y < vy+vh
    }

    return false
}

// ✅ internal/inspector/standalone_inspector.go
func (si *StandaloneInspector) handleOverlayMouse(localX, localY int, eventType frameworkevent.EventType, btn frameworkevent.MouseButton) bool {
    // 修复边界计算逻辑，支持 rows 1-3
    if localY < 1 || localY > 3 {
        return false
    }
    // ... 正确的 tab 切换逻辑
}
```

### 验收标准
- [x] ElementVNode 有 bounds 字段
- [x] vnodeContains 支持两种签名
- [x] handleOverlayMouse 修复边界计算
- [ ] Inspector tab 点击工作正常
- [ ] Hover 检测显示正确组件

---

## Phase 1: HitMap 系统（Week 1-2）

### 目标
实现统一的命中测试，消除各组件手写 bounds 的需求

### 1.1 定义 HitMap 结构

**文件**: `runtime/event/hitmap.go`

```go
package event

import (
    "sort"
    "github.com/wwsheng009/mint/runtime/layout"
)

// Rect 矩形区域
type Rect struct {
    X, Y, Width, Height int
}

// Contains 检查点是否在矩形内
func (r Rect) Contains(x, y int) bool {
    return x >= r.X && x < r.X+r.Width &&
           y >= r.Y && y < r.Y+r.Height
}

// HitMapEntry 命中条目
type HitMapEntry struct {
    NodeID    string
    Node      layout.Node
    Bounds    Rect
    LocalXY   func(screenX, screenY int) (localX, localY int)
    ZOrder   int
}

// HitMap 命中映射表
type HitMap struct {
    entries   []HitMapEntry
    root      layout.Node
    buildTime int64
}

// BuildHitMap 从布局树构建
func BuildHitMap(root layout.Node) *HitMap {
    hm := &HitMap{
        root: root,
        entries: make([]HitMapEntry, 0),
    }
    hm.walkAndBuild(root, 0, 0, 0)
    hm.sortByZOrder()
    return hm
}

// walkAndBuild 递归构建 HitMap
func (hm *HitMap) walkAndBuild(node layout.Node, x, y int, zOrder int) {
    if node == nil {
        return
    }

    // 获取节点位置和大小
    nodeX, nodeY := node.GetPosition()
    width, height := node.GetSize()

    // 创建条目
    entry := HitMapEntry{
        NodeID:  node.ID(),
        Node:    node,
        Bounds:  Rect{X: nodeX, Y: nodeY, Width: width, Height: height},
        LocalXY: func(screenX, screenY int) (int, int) {
            return screenX - nodeX, screenY - nodeY
        },
        ZOrder: zOrder,
    }
    hm.entries = append(hm.entries, entry)

    // 递归处理子节点
    children := node.Children()
    for _, child := range children {
        hm.walkAndBuild(child, x, y, zOrder+1)
    }
}

// sortByZOrder 按 Z-order 排序（上层优先）
func (hm *HitMap) sortByZOrder() {
    sort.Slice(hm.entries, func(i, j int) bool {
        return hm.entries[i].ZOrder < hm.entries[j].ZOrder
    })
}

// HitTest 执行命中测试
func (hm *HitMap) HitTest(x, y int) *HitMapEntry {
    // 从上到下查找（Z-index 降序）
    for i := len(hm.entries) - 1; i >= 0; i-- {
        entry := &hm.entries[i]
        if entry.Bounds.Contains(x, y) {
            return entry
        }
    }
    return nil
}

// FindByID 按节点 ID 查找（用于测试）
func (hm *HitMap) FindByID(id string) *HitMapEntry {
    for i := range hm.entries {
        if hm.entries[i].NodeID == id {
            return &hm.entries[i]
        }
    }
    return nil
}

// Dump 调试输出
func (hm *HitMap) Dump() string {
    var buf strings.Builder
    buf.WriteString("=== HitMap ===\n")
    for _, entry := range hm.entries {
        buf.WriteString(fmt.Sprintf("[%s] %v (Z:%d)\n",
            entry.NodeID, entry.Bounds, entry.ZOrder))
    }
    return buf.String()
}
```

### 1.2 扩展 MouseEvent

**文件**: `framework/event/mouse_event.go`

```go
package event

// MouseEvent 鼠标事件（增强版）
type MouseEvent struct {
    BaseEvent

    // 屏幕坐标
    X, Y int

    // 命中信息（由 Pump 填充）
    TargetID string     // 目标节点 ID
    LocalX   int        // 相对于目标的坐标
    LocalY   int

    // 鼠标操作
    Button   MouseButton
    Action   MouseAction
    Delta    int  // 滚轮增量（+1/-1）

    // 修饰键
    Modifiers Modifier
}

type MouseAction int

const (
    MousePress MouseAction = iota
    MouseRelease
    MouseMove
    MouseWheel
)

// ToAction 转换为 Action（Phase 2 实现）
func (e *MouseEvent) ToAction() *action.Action {
    return nil  // 在 Phase 2 实现
}
```

### 1.3 App 集成 HitMap

**文件**: `framework/app.go`

```go
// App 添加字段
type App struct {
    // ... 现有字段
    hitMap *event.HitMap
}

// renderAndBuildHitMap 渲染并构建 HitMap
func (a *App) renderAndBuildHitMap() {
    // 1. 执行渲染（包括布局计算）
    a.render()

    // 2. 从布局树构建 HitMap
    if layoutRoot, ok := a.root.(layout.Node); ok {
        a.hitMap = event.BuildHitMap(layoutRoot)
    }
}

// GetHitMap 获取 HitMap
func (a *App) GetHitMap() *event.HitMap {
    return a.hitMap
}
```

### 1.4 Pump 填充命中信息

**文件**: `runtime/pump.go`

```go
func (p *Pump) handleMouseInput(x, y int, btn MouseButton, action MouseAction) *MouseEvent {
    ev := &MouseEvent{
        X: x, Y: y,
        Button: btn,
        Action: action,
        BaseEvent: BaseEvent{
            timestamp: time.Now(),
            eventType: EventMousePress,
        },
    }

    // 填充命中信息
    if p.app != nil {
        if hitMap := p.app.GetHitMap(); hitMap != nil {
            if entry := hitMap.HitTest(x, y); entry != nil {
                ev.TargetID = entry.NodeID
                ev.LocalX, ev.LocalY = entry.LocalXY(x, y)
            }
        }
    }

    return ev
}
```

### 验收标准
- [ ] HitMap 正确构建所有节点
- [ ] `HitTest(x,y)` 返回正确目标
- [ ] MouseEvent 包含 TargetID/LocalX/LocalY
- [ ] Inspector tab 点击工作正常
- [ ] Hover 检测显示正确组件
- [ ] 单元测试：`TestHitMap_Build`, `TestHitMap_HitTest`, `TestHitMap_FindByID`

---

## Phase 2: Action 系统（Week 3-4）

### 目标
实现语义化的 Action 层，Event → Action 转换

### 2.1 定义 Action 类型

**文件**: `framework/action/action.go`

```go
package action

// Action 语义化操作
type Action struct {
    Type     ActionType
    Payload  interface{}
    Source   string  // 触发源
    TargetID string  // 目标节点 ID
}

// ActionType Action 类型
type ActionType string

const (
    // 导航类
    ActionNavigateNext      ActionType = "navigate_next"
    ActionNavigatePrev      ActionType = "navigate_prev"
    ActionNavigateUp        ActionType = "navigate_up"
    ActionNavigateDown      ActionType = "navigate_down"
    ActionNavigatePageUp    ActionType = "navigate_page_up"
    ActionNavigatePageDown  ActionType = "navigate_page_down"
    ActionNavigateHome      ActionType = "navigate_home"
    ActionNavigateEnd       ActionType = "navigate_end"

    // 选择类
    ActionSelect   ActionType = "select"
    ActionToggle   ActionType = "toggle"
    ActionExpand   ActionType = "expand"
    ActionCollapse ActionType = "collapse"

    // 编辑类
    ActionInputText   ActionType = "input_text"
    ActionDeleteChar  ActionType = "delete_char"
    ActionDeleteWord  ActionType = "delete_word"
    ActionDeleteLine  ActionType = "delete_line"

    // 表单类
    ActionSubmit   ActionType = "submit"
    ActionCancel   ActionType = "cancel"
    ActionValidate ActionType = "validate"

    // 系统类
    ActionQuit    ActionType = "quit"
    ActionFocus   ActionType = "focus"
    ActionBlur    ActionType = "blur"
    ActionInspect ActionType = "inspect"  // Inspector 切换

    // 鼠标特定
    ActionClick       ActionType = "click"
    ActionDoubleClick ActionType = "double_click"
    ActionRightClick  ActionType = "right_click"
    ActionScroll      ActionType = "scroll"
    ActionHover       ActionType = "hover"
)

// IsNavigation 是否为导航 Action
func (a *Action) IsNavigation() bool {
    switch a.Type {
    case ActionNavigateNext, ActionNavigatePrev, ActionNavigateUp,
         ActionNavigateDown, ActionNavigatePageUp, ActionNavigatePageDown,
         ActionNavigateHome, ActionNavigateEnd:
        return true
    }
    return false
}

// IsEditing 是否为编辑 Action
func (a *Action) IsEditing() bool {
    switch a.Type {
    case ActionInputText, ActionDeleteChar, ActionDeleteWord, ActionDeleteLine:
        return true
    }
    return false
}
```

### 2.2 Event → Action 转换器

**文件**: `framework/action/processor.go`

```go
package action

// InputProcessor 将 Event 转换为 Action
type InputProcessor struct {
    keyMap *KeyMap
}

// Process 转换事件
func (p *InputProcessor) Process(ev event.Event) *Action {
    switch e := ev.(type) {
    case *event.KeyEvent:
        return p.processKeyEvent(e)
    case *event.MouseEvent:
        return p.processMouseEvent(e)
    }
    return nil
}

// processKeyEvent 处理键盘事件
func (p *InputProcessor) processKeyEvent(ev *KeyEvent) *Action {
    // 1. 优先使用 KeyMap 语义化
    if act := p.keyMap.Lookup(ev); act != nil {
        act.TargetID = ""  // 键盘事件通常无特定目标
        return act
    }

    // 2. 默认转换规则
    switch {
    case ev.Special == KeyTab:
        return &Action{Type: ActionNavigateNext}
    case ev.Special == KeyEnter:
        return &Action{Type: ActionSelect}
    case ev.Key == 'q' && ev.Modifiers.Has(ModCtrl):
        return &Action{Type: ActionQuit}
    case ev.Key >= 32 && ev.Key <= 126:
        // 可打印字符
        return &Action{
            Type:    ActionInputText,
            Payload: string(ev.Key),
        }
    }
    return nil
}

// processMouseEvent 处理鼠标事件
func (p *InputProcessor) processMouseEvent(ev *MouseEvent) *Action {
    switch ev.Action {
    case MousePress:
        if ev.Button == MouseLeft {
            return &Action{
                Type:     ActionClick,
                TargetID: ev.TargetID,
                Payload: struct{ X, Y int }{ev.LocalX, ev.LocalY},
            }
        } else if ev.Button == MouseRight {
            return &Action{
                Type:     ActionRightClick,
                TargetID: ev.TargetID,
            }
        }
    case MouseWheel:
        return &Action{
            Type:     ActionScroll,
            TargetID: ev.TargetID,
            Payload:  ev.Delta,
        }
    case MouseMove:
        return &Action{
            Type:     ActionHover,
            TargetID: ev.TargetID,
            Payload: struct{ X, Y int }{ev.LocalX, ev.LocalY},
        }
    }
    return nil
}
```

### 2.3 KeyMap 上下文感知

**文件**: `framework/action/keymap.go`

```go
package action

// KeyMap 键盘映射（支持上下文）
type KeyMap struct {
    globalMappings map[KeySignature]*Action
    contextStack   []string
}

// KeySignature 按键签名
type KeySignature struct {
    Key      rune
    Special  SpecialKey
    Modifiers KeyModifier
    Context  string  // 上下文（如 "input", "tree"）
}

// Lookup 查找映射
func (km *KeyMap) Lookup(ev *KeyEvent) *Action {
    currentContext := ""
    if len(km.contextStack) > 0 {
        currentContext = km.contextStack[len(km.contextStack)-1]
    }

    sig := KeySignature{
        Key:      ev.Key,
        Special:  ev.Special,
        Modifiers: ev.Modifiers,
        Context:  currentContext,
    }

    return km.globalMappings[sig]
}

// PushContext 推入上下文
func (km *KeyMap) PushContext(ctx string) {
    km.contextStack = append(km.contextStack, ctx)
}

// PopContext 弹出上下文
func (km *KeyMap) PopContext() {
    if len(km.contextStack) > 0 {
        km.contextStack = km.contextStack[:len(km.contextStack)-1]
    }
}

// Bind 绑定快捷键
func (km *KeyMap) Bind(key rune, special SpecialKey, modifiers KeyModifier,
    context string, action *Action) {
    sig := KeySignature{
        Key:      key,
        Special:  special,
        Modifiers: modifiers,
        Context:  context,
    }
    km.globalMappings[sig] = action
}
```

### 2.4 ActionTarget 接口

**文件**: `framework/component/action_target.go`

```go
package component

// ActionTarget 响应语义化 Action 的组件
type ActionTarget interface {
    // HandleAction 处理 Action
    // 返回 true 表示已处理
    HandleAction(*action.Action) bool
}

// 示例：TreeView 实现
func (t *TreeView) HandleAction(act *action.Action) bool {
    switch act.Type {
    case action.ActionNavigateDown:
        t.MoveDown()
        return true
    case action.ActionNavigateUp:
        t.MoveUp()
        return true
    case action.ActionSelect:
        if lineIdx, ok := act.Payload.(int); ok {
            t.SelectLine(lineIdx)
        }
        return true
    case action.ActionExpand:
        t.ExpandCurrent()
        return true
    case action.ActionCollapse:
        t.CollapseCurrent()
        return true
    case action.ActionScroll:
        if delta, ok := act.Payload.(int); ok {
            t.ScrollBy(delta)
        }
        return true
    }
    return false
}
```

### 验收标准
- [ ] Action 类型定义完整
- [ ] KeyEvent 正确转换为 Navigate/Input/Quit
- [ ] MouseEvent 正确转换为 Click/Scroll/Hover
- [ ] KeyMap 支持上下文切换
- [ ] 组件实现 HandleAction 接口
- [ ] 单元测试覆盖所有转换规则

---

## Phase 3: Router 三阶段分发（Week 5）

### 目标
实现 V3 规范的完整 Router，Inspector 作为捕获监听器

### 3.1 Router 实现

**文件**: `runtime/event/router.go`

```go
package event

// Router 事件路由器（V3）
type Router struct {
    captureHandlers []CaptureHandler
    focus           *focus.Manager
    root            Node
    processor       *action.InputProcessor
}

// NewRouter 创建 Router
func NewRouter(root Node) *Router {
    return &Router{
        captureHandlers: make([]CaptureHandler, 0),
        root:            root,
        processor:       action.NewInputProcessor(),
    }
}

// AddCaptureHandler 添加捕获处理器
func (r *Router) AddCaptureHandler(handler CaptureHandler) {
    r.captureHandlers = append(r.captureHandlers, handler)
    // 按优先级排序
    sort.Slice(r.captureHandlers, func(i, j int) bool {
        return r.captureHandlers[i].Priority() > r.captureHandlers[j].Priority()
    })
}

// Route 路由事件（三阶段）
func (r *Router) Route(ev Event) bool {
    // 1. Capture Phase
    if r.capturePhase(ev) {
        return true
    }

    // 2. 转换为 Action
    act := r.processor.Process(ev)
    if act == nil {
        return false
    }

    // 3. Target Phase
    if r.targetPhase(act) {
        return true
    }

    // 4. Bubble Phase
    return r.bubblePhase(act)
}

// capturePhase 捕获阶段
func (r *Router) capturePhase(ev Event) bool {
    ev.SetPhase(PhaseCapture)

    for _, handler := range r.captureHandlers {
        if handler.HandleCapture(ev) {
            return true
        }
        if ev.IsPropagationStopped() {
            return true
        }
    }
    return false
}

// targetPhase 目标阶段
func (r *Router) targetPhase(act *action.Action) bool {
    target := r.findNodeByID(act.TargetID)
    if target == nil {
        return false
    }

    if handler, ok := target.(component.ActionTarget); ok {
        return handler.HandleAction(act)
    }

    return false
}

// bubblePhase 冒泡阶段
func (r *Router) bubblePhase(act *action.Action) bool {
    target := r.findNodeByID(act.TargetID)
    if target == nil {
        return false
    }

    parents := r.getParentChain(target)

    for i := len(parents) - 1; i >= 0; i-- {
        if handler, ok := parents[i].(component.ActionTarget); ok {
            if handler.HandleAction(act) {
                return true
            }
        }
    }

    return false
}
```

### 3.2 Inspector 作为捕获监听器

**文件**: `internal/inspector/standalone_inspector.go`

```go
// 实现 CaptureHandler 接口
func (si *StandaloneInspector) HandleCapture(ev event.Event) bool {
    if !si.visible || !si.enabled {
        return false
    }

    // 检查鼠标是否在 overlay 上
    if mouseEv, ok := ev.(*event.MouseEvent); ok {
        if si.isMouseOverOverlay(mouseEv.X, mouseEv.Y) {
            return si.forwardToOverlay(mouseEv.LocalX, mouseEv.LocalY)
        }
    }

    // F12 快捷键
    if keyEv, ok := ev.(*event.KeyEvent); ok {
        if keyEv.Special == KeyF12 {
            si.ToggleVisibility()
            return true
        }
    }

    return false
}

func (si *StandaloneInspector) Priority() int {
    return PriorityInspector  // 最高优先级
}
```

### 验收标准
- [ ] Router 实现三阶段分发
- [ ] Inspector 正确捕获 overlay 事件
- [ ] 全局快捷键在 Capture Phase 处理
- [ ] 事件可停止传播
- [ ] 单元测试覆盖三阶段流程

---

## Phase 4-6: 简述

详见独立文档：
- Phase 4: [Msg/Cmd 系统](./MSG_UNIFICATION.md)
- Phase 5: [测试工具](./TASKS.md)
- Phase 6: [性能优化](./TASKS.md)

---

## 总体验收标准

### 功能完整性
- [ ] Inspector tab 点击工作
- [ ] Hover 检测显示正确组件
- [ ] 所有事件流经三阶段
- [ ] Action 语义化完整
- [ ] Msg/Cmd 可组合

### 性能指标
- [ ] HitMap 构建 < 10ms (1000 节点)
- [ ] MouseMove 节流到 60Hz
- [ ] 事件延迟 < 16ms

### 测试覆盖率
- [ ] 单元测试 > 80%
- [ ] 集成测试覆盖主要流程
- [ ] 测试可按 ID 注入事件

### 文档完整性
- [ ] API 文档更新
- [ ] 迁移指南完整
- [ ] 示例代码同步

---

**下一步**: 查看 [Msg 统一设计](./MSG_UNIFICATION.md) 和 [任务列表](./TASKS.md)
