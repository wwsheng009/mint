# Inspector HitMap + Msg/Cmd 系统重构方案

> **版本**: v2.0 (修正版)
> **日期**: 2026-02-10
> **关键修正**: 使用 Update(Msg)/Cmd 系统，而非老的 HandleEvent(Event)

---

## 🎯 核心目标

**问题**：Inspector overlay 无法正确响应鼠标点击

**根本原因**：
1. Inspector 被 Fragment 包装，不在 layout.Node 层级，HitMap 不包含它
2. 手动坐标解析违反设计原则
3. **关键**：还在使用老的 `HandleEvent(Event)`，而非新的 `Update(Msg)/Cmd` 系统

**解决方案**：
1. 从 VNode 树构建 HitMap（包含 Inspector）
2. Router 使用 Msg/Cmd 系统分发事件
3. 组件实现 `Update(Msg) cmd.Cmd` 接口

---

## 📋 架构现状分析

### ✅ 已实现的基础设施

根据 Phase 1-6 完成报告，以下功能已实现：

#### Phase 1: HitMap 系统 ✅
- `runtime/event/hitmap.go` - HitMap 数据结构
- `BuildHitMap(layout.Node)` - 从 layout.Node 构建
- `framework/event/pump.go` - Pump 填充 TargetID/LocalX/LocalY

#### Phase 4: Msg/Cmd 系统 ✅
- `framework/msg/msg.go` - Msg 核心接口
- `framework/msg/mouse_msg.go` - MouseMsg（含 TargetID, LocalX, LocalY）
- `framework/msg/key_msg.go` - KeyMsg
- `framework/msg/adapter.go` - Event → Msg 适配器
  - `MouseEventToMsg(MouseEvent) MouseMsg` ✅
- `framework/component/updater.go` - Updater 接口
  - `Update(message msg.Msg) cmd.Cmd` ✅

### 🔴 当前问题

1. **HitMap 构建问题**：
   - App 只能从 `layout.Node` 构建 HitMap
   - Inspector 是 `VNode`（Fragment），不在 layout.Node 层级
   - 导致 Inspector 不在 HitMap 中

2. **接口使用问题**（关键）：
   - Pump 输出 `Event`（MouseEvent）
   - Router 分发 `Event`
   - 组件实现 `HandleEvent(Event) bool`（老接口）
   - **应该使用**：`Update(Msg) cmd.Cmd`（新接口）

---

## 🎯 正确的事件流程

### 当前（错误）流程

```
用户点击 → Pump.HitTest → MouseEvent
  → Router.Dispatch(Event)
  → 组件.HandleEvent(Event)  ❌ 老接口
  → 手动解析坐标 ❌
```

### 目标（正确）流程

```
用户点击 → Pump.HitTest → MouseEvent{TargetID, LocalX, LocalY}
  → Adapter: MouseEventToMsg(Event) → MouseMsg
  → Router.Dispatch(MouseMsg) 根据 TargetID
  → 组件.Update(MouseMsg) → Cmd  ✅ 新接口
  → 使用 MouseMsg.LocalX/LocalY ✅
```

---

## 📝 实施方案

### Phase 1: 从 VNode 树构建 HitMap（2-3天）

#### 目标
让 HitMap 包含 Inspector overlay 的所有节点

#### 步骤

**1.1 创建 VNode HitMap 构建函数**

```go
// runtime/event/hitmap_vnode.go (新建)
package runtimeevent

import (
    "github.com/wwsheng009/mint/runtime/ui"
)

// BuildHitMapFromVNode 从 VNode 树构建 HitMap
func BuildHitMapFromVNode(root ui.VNode, offsetX, offsetY int) *HitMap {
    hitMap := NewHitMap()
    buildVNodeRecursive(root, offsetX, offsetY, hitMap)
    return hitMap
}

// buildVNodeRecursive 递归构建 HitMap
func buildVNodeRecursive(node ui.VNode, offsetX, offsetY int, hitMap *HitMap) {
    // 检查节点是否实现了 GetBounds() 接口
    type BoundsGetter interface {
        GetBounds() [4]int
    }

    if bg, ok := node.(BoundsGetter); ok {
        bounds := bg.GetBounds()
        x, y, w, h := bounds[0], bounds[1], bounds[2], bounds[3]

        // 使用绝对坐标（bounds 已经是屏幕坐标）
        nodeID := getNodeID(node)

        entry := &HitMapEntry{
            NodeID: nodeID,
            Bounds: Rect{X: x, Y: y, Width: w, Height: h},
            LocalXY: func(globalX, globalY int) (int, int) {
                return globalX - x, globalY - y
            },
        }
        hitMap.Add(entry)
    }

    // 递归子节点
    for _, child := range node.Children() {
        buildVNodeRecursive(child, offsetX, offsetY, hitMap)
    }
}

// getNodeID 获取或生成节点 ID
func getNodeID(node ui.VNode) string {
    if key := node.Key(); key != "" {
        return key
    }
    if tag := node.Tag(); tag != "" {
        return tag
    }
    return fmt.Sprintf("node-%p", node)
}
```

**关键点**：
- ✅ 直接访问 VNode 接口，无需适配器
- ✅ 利用布局引擎已设置的 bounds 字段
- ✅ 支持 Fragment（多子节点遍历）
- ✅ 无循环导入风险（runtime/event 不导入 runtime/ui）

**1.2 集成到 App.render()**

```go
// framework/app.go
func (a *App) render() {
    // ... 现有渲染逻辑 ...

    // ============================================================================
    // Phase 1: 构建 HitMap（从 VNode 树，包括 Inspector）
    // ============================================================================
    if a.root != nil {
        // 获取 VNode 树（包括 Fragment）
        var vnodeTree ui.VNode
        if painter, ok := a.root.(interface{ Render() ui.VNode }); ok {
            vnodeTree = painter.Render()
        }

        if vnodeTree != nil {
            // 从 VNode 树构建 HitMap（包括 Inspector overlay）
            a.hitMap = runtimeevent.BuildHitMapFromVNode(vnodeTree, 0, 0)

            // 传递给 Pump
            if a.pump != nil {
                a.pump.SetHitMap(a.hitMap)
            }

            // DEBUG
            if os.Getenv("TUI_DEBUG_HITMAP") == "true" {
                fmt.Fprintf(os.Stderr, "[APP] HitMap from VNode: %d entries\n", a.hitMap.Size())
            }
        }
    }
}
```

**1.3 单元测试**

```go
// runtime/event/hitmap_vnode_test.go
func TestBuildHitMapFromVNode_WithFragment(t *testing.T) {
    // 创建 Fragment(appNode, inspectorNode)
    appNode := ui.Text("App Content")
    inspectorNode := ui.Text("Inspector")

    fragment := ui.Fragment(appNode, inspectorNode)

    // 构建 HitMap
    hitMap := BuildHitMapFromVNode(fragment, 0, 0)

    // 验证两个节点都在 HitMap 中
    assert.Equal(t, 2, hitMap.Size())
}
```

### Phase 2: Router 使用 Msg/Cmd 系统（1-2天）

#### 目标
让 Router 分发 Msg 而不是 Event

#### 步骤

**2.1 检查 Router 当前实现**

确认 Router 的分发逻辑：
```bash
grep -r "Router.Dispatch" framework/
grep -r "HandleEvent" framework/
```

**2.2 修改 Router 支持 Msg 分发**

```go
// framework/event/router.go
func (r *Router) Dispatch(ev Event) bool {
    // 步骤1: Event → Msg 转换
    var msg msg.Msg
    switch e := ev.(type) {
    case *MouseEvent:
        msg = msg.MouseEventToMsg(e)
    case *KeyEvent:
        msg = msg.KeyEventToMsg(e)
    default:
        // 不支持的事件类型，回退到老的处理方式
        return r.dispatchEventLegacy(ev)
    }

    if msg == nil {
        return false
    }

    // 步骤2: 根据 TargetID 直接分发
    if mouseMsg, ok := msg.(*msg.MouseMsg); ok && mouseMsg.TargetID != "" {
        return r.dispatchToTarget(mouseMsg.TargetID, msg)
    }

    // 步骤3: 无 TargetID，回退到层次遍历
    return r.dispatchToHierarchy(msg)
}

// dispatchToTarget 根据 TargetID 直接分发 Msg
func (r *Router) dispatchToTarget(targetID string, msg msg.Msg) bool {
    // 从 HitMap 查找组件实例
    component := r.findComponentByID(targetID)
    if component == nil {
        return false
    }

    // 优先使用 Update(Msg)
    if updater, ok := component.(component.Updater); ok {
        cmd := updater.Update(msg)
        if cmd != nil {
            r.executeCmd(cmd)
        }
        return true
    }

    // 回退到 HandleEvent（兼容过渡期）
    if handler, ok := component.(event.Component); ok {
        // Msg → Event 反向转换（如果需要）
        return handler.HandleEvent(msgToEvent(msg))
    }

    return false
}
```

**关键点**：
- ✅ Event → Msg 转换在 Router 中进行
- ✅ 根据 TargetID 直接分发，无需层次遍历
- ✅ 优先使用 `Update(Msg)`，回退到 `HandleEvent(Event)`（过渡期）

### Phase 3: 组件实现 Update(Msg)（1-2天）

#### 目标
让 Inspector 和其他组件使用 Update(Msg) 接口

#### 步骤

**3.1 Inspector 实现 Update(Msg)**

```go
// internal/inspector/standalone_inspector.go

// 删除老的手动坐标解析代码 ❌
/*
func (si *StandaloneInspector) HandleEvent(ev event.Event) bool {
    // 手动计算 localX, localY ❌
    if ev.X >= minX && ev.X < maxX { ... }
}
*/

// ✅ 实现 Update(Msg)
func (si *StandaloneInspector) Update(message msg.Msg) cmd.Cmd {
    switch m := message.(type) {
    case *msg.MouseMsg:
        return si.handleMouseMsg(m)
    case *msg.KeyMsg:
        return si.handleKeyMsg(m)
    }
    return cmd.None()
}

func (si *StandaloneInspector) handleMouseMsg(mouseMsg *msg.MouseMsg) cmd.Cmd {
    // ✅ Pump 已经设置了 TargetID 和 LocalX/LocalY
    // Inspector 直接使用这些信息

    fmt.Printf("[Inspector] Mouse %s on %s at (%d, %d)\n",
        mouseMsg.Action,
        mouseMsg.TargetID,
        mouseMsg.LocalX,
        mouseMsg.LocalY)

    // 根据 TargetID 分发到内部组件
    if strings.HasPrefix(mouseMsg.TargetID, "inspector-tree") {
        // 分发到 TreeView
        return si.dispatchToTreeView(mouseMsg)
    }

    return cmd.None()
}

func (si *StandaloneInspector) dispatchToTreeView(mouseMsg *msg.MouseMsg) cmd.Cmd {
    // TreeView 也应该实现 Update(Msg)
    if treeView, ok := si.treeView.(component.Updater); ok {
        return treeView.Update(mouseMsg)
    }

    // 兼容过渡期：TreeView 还在用 HandleEvent
    if treeView, ok := si.treeView.(event.Component); ok {
        mouseEvent := msgToMouseEvent(mouseMsg)
        handled := treeView.HandleEvent(mouseEvent)
        if handled {
            return cmd.None()
        }
    }

    return cmd.None()
}
```

**3.2 TreeView 实现 Update(Msg)**

```go
// components/tree/tree.go

// ✅ 新接口
func (t *TreeView) Update(message msg.Msg) cmd.Cmd {
    switch m := message.(type) {
    case *msg.MouseMsg:
        if m.IsClick() {
            return t.handleClick(m)
        }
        if m.IsScroll() {
            return t.handleScroll(m)
        }
    case *msg.KeyMsg:
        return t.handleKey(m)
    }
    return cmd.None()
}

func (t *TreeView) handleClick(mouseMsg *msg.MouseMsg) cmd.Cmd {
    // ✅ 直接使用 LocalX/LocalY，无需手动计算
    row := mouseMsg.LocalY

    if row >= 0 && row < len(t.nodes) {
        t.selectedNode = t.nodes[row]
        t.markDirty() // 标记需要重新渲染
        fmt.Printf("[TreeView] Selected row %d: %s\n", row, t.selectedNode.Text)
    }

    return cmd.None() // 无副作用
}

func (t *TreeView) handleScroll(mouseMsg *msg.MouseMsg) cmd.Cmd {
    // 使用 mouseMsg.Delta 滚动
    t.scrollOffset += mouseMsg.Delta
    t.markDirty()
    return cmd.None()
}

// 保留老接口用于兼容（过渡期）
func (t *TreeView) HandleEvent(ev event.Event) bool {
    // 转换到 Update(Msg)
    msg := eventToMsg(ev)
    if msg != nil {
        cmd := t.Update(msg)
        return cmd != nil
    }
    return false
}
```

**3.3 Panel 和 Tabs 清理**

```go
// components/container/panel.go

// ❌ 删除手动事件转发
/*
func (p *Panel) HandleEvent(ev event.Event) bool {
    if contentComponent, ok := p.content.(event.Component); ok {
        return contentComponent.HandleEvent(ev)
    }
}
*/

// ✅ Panel 不需要 HandleEvent，Router 根据 TargetID 直接分发到内容

// components/navigation/tabs.go

func (t *TabsVNode) Update(message msg.Msg) cmd.Cmd {
    mouseMsg, ok := message.(*msg.MouseMsg)
    if !ok || !mouseMsg.IsClick() {
        return nil
    }

    // 使用 TargetID 识别点击的是 tab 栏还是内容区
    if strings.HasPrefix(mouseMsg.TargetID, "tab-bar-") {
        return t.handleTabBarClick(mouseMsg)
    }

    // Tab 内容会直接接收到事件，无需转发
    return nil
}

func (t *TabsVNode) handleTabBarClick(mouseMsg *msg.MouseMsg) cmd.Cmd {
    // 使用 mouseMsg.LocalX 识别点击了哪个 tab
    clickedIndex := t.findTabIndexByPosition(mouseMsg.LocalX)
    if clickedIndex >= 0 {
        t.activeTab = clickedIndex
        t.markDirty()
    }
    return cmd.None()
}
```

### Phase 4: 验证和测试（1天）

#### 集成测试场景

```bash
cd examples/ui_demos/demo2_runtime_internals/inspector_overlay
go run main.go

# 测试清单
# 1. 按 F12 打开 Inspector
# 2. 点击 Inspector 的 TreeView
#    - 验证选中正确的节点
#    - 验证控制台输出正确的 LocalY
# 3. 按 ↑↓ 键导航 TreeView
# 4. 点击其他 Inspector 组件（Tabs, Input, Button）
# 5. 关闭 Inspector（F12）
# 6. 验证主应用组件正常工作
```

#### 调试工具

```bash
# 启用 HitMap 调试
export TUI_DEBUG_HITMAP=true

# 查看 HitMap 内容
# 应该看到 inspector-treeview, inspector-tabs 等节点
```

---

## ✅ 验收标准

### 功能验收
- [ ] Inspector TreeView 点击正确选中节点
- [ ] 控制台输出 `Selected row X` 中的 X 与点击位置一致
- [ ] Inspector 所有组件（TreeView, Tabs, Input, Button）正常响应
- [ ] 主应用组件不受影响

### 架构验收
- [ ] HitMap 包含 Inspector overlay 节点
- [ ] Router 分发 Msg（MouseMsg/KeyMsg）
- [ ] 组件实现 `Update(Msg) cmd.Cmd` 接口
- [ ] 无手动坐标解析代码
- [ ] 无事件转发代码（Panel/Tabs）

### 代码质量
- [ ] 单元测试覆盖率 > 80%
- [ ] 删除所有手动坐标计算代码
- [ ] 删除所有手动事件转发代码

---

## 📊 预期收益

| 指标 | 当前 | 目标 | 提升 |
|------|------|------|------|
| Inspector 点击成功率 | ~30% | 100% | +233% |
| 点击位置准确性 | 经常错误 | 完全准确 | ∞ |
| 手动坐标代码 | ~100 行 | 0 行 | -100% |
| 事件转发代码 | ~50 行 | 0 行 | -100% |
| 使用老接口 HandleEvent | 所有组件 | 0 个 | -100% |
| 使用新接口 Update(Msg) | 0 个 | 所有组件 | +∞ |

---

## 🔑 关键差异对比

### 老（错误）做法 ❌

```go
func (inspector *Inspector) HandleEvent(ev event.Event) bool {
    me := ev.(*MouseEvent)

    // ❌ 手动检查 bounds
    if me.X >= minX && me.X < maxX && me.Y >= minY && me.Y < maxY {
        // ❌ 手动计算 localX/localY
        localX := me.X - minX
        localY := me.Y - minY

        // ❌ 手动转发到子组件
        if child != nil {
            return child.HandleEvent(me)
        }
    }
    return false
}
```

### 新（正确）做法 ✅

```go
func (inspector *Inspector) Update(message msg.Msg) cmd.Cmd {
    mouseMsg, ok := message.(*msg.MouseMsg)
    if !ok {
        return cmd.None()
    }

    // ✅ Pump 已经设置了 TargetID
    fmt.Printf("Click on %s\n", mouseMsg.TargetID)

    // ✅ Pump 已经计算了 LocalX/LocalY
    fmt.Printf("Local: (%d, %d)\n", mouseMsg.LocalX, mouseMsg.LocalY)

    // ✅ Router 根据 TargetID 直接分发到目标组件
    // 无需手动转发！

    return cmd.None()
}
```

---

## 📚 相关文档

- **Msg/Cmd 设计**: [MSG_UNIFICATION.md](./MSG_UNIFICATION.md)
- **Phase 4 完成报告**: [PHASE_4_COMPLETION.md](./PHASE_4_COMPLETION.md)
- **长期架构**: [long_term_event_architecture.md](../../event/long_term_event_architecture.md)

---

**状态**: 待审核实施
**工期**: 5-7 天
**优先级**: 高（修正架构偏差，解决功能缺陷）
