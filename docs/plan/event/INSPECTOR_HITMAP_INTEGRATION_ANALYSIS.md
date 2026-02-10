# Inspector HitMap 集成架构分析与实施方案

> **版本**: v1.0
> **日期**: 2026-02-10
> **状态**: 实施阶段
> **作者**: Claude (Sonnet 4.5)

---

## 📋 执行摘要

### 问题陈述

当前 Mint TUI 框架中，Inspector overlay 的 TreeView 组件无法正确响应鼠标点击事件。具体表现为：
- 点击位置与选中的节点不匹配
- 某些点击完全无响应
- 必须手动解析坐标，违反了"组件不再手写命中"的设计原则

### 根本原因

通过深入分析，发现**三个核心架构问题**：

1. **接口不兼容问题**：VNode、layout.Node、component.Node 三个接口系统的 Children() 方法返回不同类型
2. **Inspector 被 HitMap 排除**：Inspector 通过 Fragment 注入，Fragment 不在 layout.Node 层级中
3. **循环依赖风险**：尝试用适配器桥接 VNode 和 layout.Node 时产生循环导入

### 推荐解决方案

**统一 HitMap 构建系统**：从布局引擎设置的 ElementVNode.bounds 字段直接提取边界信息，避免类型转换和适配器。

**核心优势**：
- ✅ 无需适配器，避免循环导入
- ✅ 利用布局引擎已有的 bounds 信息
- ✅ 支持所有 VNode 类型（包括 Fragment/Inspector）
- ✅ 符合长期架构设计原则

---

## 🔍 当前架构深度分析

### 1. 三套并行的节点系统

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Mint TUI 节点系统                            │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐     │
│  │   component.Node│  │   layout.Node   │  │     VNode       │     │
│  │   (运行时组件)   │  │   (布局系统)     │  │   (声明式UI)     │     │
│  └────────┬────────┘  └────────┬────────┘  └────────┬────────┘     │
│           │                    │                    │               │
│           v                    v                    v               │
│  Children() []Component  Children() []Node  Children() []VNode      │
│  HandleEvent(Event)    GetPosition()      Type() VNodeType         │
│                        SetPosition()      Props() Props             │
│                        GetBounds()        Style() Style             │
│                                            bounds [4]int  ⚠️       │
│           └────────────────────┬────────────────────┘              │
│                                │                                    │
│                                v                                    │
│                      ┌─────────────────┐                            │
│                      │  类型不兼容！     │                            │
│                      │  无法直接转换！   │                            │
│                      └─────────────────┘                            │
└─────────────────────────────────────────────────────────────────────┘
```

**问题本质**：
- `component.Node`：面向对象的组件系统，带事件处理
- `layout.Node`：布局引擎接口，操作几何信息
- `VNode`：声明式 UI 接口，描述渲染树
- **三者无法相互转换**，因为 Children() 返回类型不同

### 2. HitMap 构建的当前实现

**文件**：`framework/app.go:1015-1069`

```go
// Phase 1: 构建 HitMap（在每次渲染后）
if a.root != nil {
    // 方法1：尝试从 layout.Node 构建
    if layoutRoot, ok := a.root.(layout.Node); ok {
        a.hitMap = runtimeevent.BuildHitMap(layoutRoot)
        // ...
    } else if vnodeRoot, ok := a.root.(rtui.VNode); ok {
        // 方法2：从 VNode 构建（支持 Inspector overlay）
        // 通过 VNodeAdapter 将 VNode 转换为 layout.Node
        layoutAdapter := rtui.AsLayoutNode(vnodeRoot)
        a.hitMap = runtimeevent.BuildHitMap(layoutAdapter)
        // ...
    }
}
```

**问题分析**：

1. **App.root 的实际类型**：
   - App.root 是 `Paintable` 接口（即 component.Node）
   - Inspector 通过 Hook 注入：`Fragment(appNode, inspectorContent)`
   - Fragment 是 VNode，不是 layout.Node

2. **类型断言失败**：
   ```go
   // 这行通常失败，因为 root 不是 layout.Node
   if layoutRoot, ok := a.root.(layout.Node); ok { ... }
   ```

3. **VNodeAdapter 的循环导入**：
   ```go
   // runtime/ui/vnode_adapter.go
   package ui

   import (
       "github.com/wwsheng009/mint/runtime/layout"  // ❌ 循环导入！
   )

   type VNodeAdapter struct {
       VNode VNode
   }

   func (a *VNodeAdapter) Children() []layout.Node {
       // 尝试转换 VNode.Children() []VNode → []layout.Node
       // 但导致：framework → ui → framework
   }
   ```

### 3. Inspector 的注入机制

**文件**：`internal/inspector/hook.go:21-73`

```go
func CreateInspectorHook(inspector *StandaloneInspector) render.VNodeHook {
    return func(vnode rtui.VNode) rtui.VNode {
        if !inspector.IsVisible() {
            return vnode
        }

        inspectorContent := inspector.RenderContent()
        inspectorContent.SetLayer(rtui.LayerInspector)

        // ⚠️ 关键：使用 Fragment 包装
        return rtui.Fragment(vnode, inspectorContent)
    }
}
```

**问题**：
- Fragment 是 VNode 容器，不是 layout.Node
- Fragment 的子节点（包括 Inspector）不在布局树的层级结构中
- HitMap.BuildHitMap(layout.Node) 无法遍历 Fragment 的子节点

### 4. 布局引擎已设置 bounds

**文件**：`runtime/ui/element.go:6-15`

```go
type ElementVNode struct {
    vnodeType VNodeType
    tag       string
    key       string
    props     Props
    children  []VNode
    style     style.Style
    // Layout bounds (set by layout engine for hit testing)
    bounds [4]int // [x, y, width, height]  ✅ 布局引擎已设置！
}
```

**关键发现**：
- ✅ 布局引擎已经计算并设置了所有 VNode 的 bounds
- ✅ ElementVNode.bounds 字段包含正确的位置和大小信息
- ✅ 无需重新估算或转换类型
- ❌ 但 HitMap 构建函数无法访问 VNode 的 bounds

### 5. 手动坐标解析违反设计原则

**文件**：`internal/inspector/standalone_inspector.go:1715-1757`

```go
// ❌ 错误做法：手动计算 localX/localY
if ev.X >= minX && ev.X < maxX && ev.Y >= minY && ev.Y < maxY {
    // Convert to overlay coordinates
    localX := ev.X - minX  // ❌ 手动坐标计算
    localY := ev.Y - minY

    localEv := &frameworkevent.MouseEvent{
        BaseEvent: frameworkevent.NewBaseEvent(eventType),
        X:         localX,
        Y:         localY,
        LocalX:    localX,
        LocalY:    localY,
        Button:    ev.Button,
    }
    // ...
}
```

**违反的设计原则**（来自 `docs/event/long_term_event_architecture.md:24`）：
> "MouseMsg 在 Pump 阶段即带 TargetID/localXY，组件不再手写命中。"

---

## 🎯 正确的解决方案

### 核心思路

**直接从 VNode 树构建 HitMap**，利用布局引擎已设置的 bounds 字段，完全绕过 layout.Node 接口。

### 架构设计

```
┌─────────────────────────────────────────────────────────────────────┐
│                      统一 HitMap 构建流程                            │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  1. App.render() 调用布局引擎                                         │
│     └─> 布局引擎设置所有 VNode 的 bounds 字段                        │
│                                                                      │
│  2. App.render() 从 VNode 树构建 HitMap                              │
│     └─> 遍历 VNode.Children()                                       │
│     └─> 对每个 VNode，提取 bounds 字段                               │
│     └─> 创建 HitMapEntry{NodeID, Bounds, LocalXY}                    │
│                                                                      │
│  3. Pump 处理鼠标事件                                                 │
│     └─> hitMap.HitTest(x, y) → TargetID, LocalX, LocalY             │
│     └─> 填充 ev.TargetID, ev.LocalX, ev.LocalY                       │
│                                                                      │
│  4. Router 分发事件                                                   │
│     └─> 根据 TargetID 直接分发到目标组件                             │
│     └─> 组件使用 ev.LocalX/LocalY，无需手动计算                       │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### 代码结构

```go
// runtime/event/hitmap_vnode.go (新文件)
package runtimeevent

// BuildHitMapFromVNode 从 VNode 树构建 HitMap
func BuildHitMapFromVNode(root VNode, offsetX, offsetY int) *HitMap {
    hitMap := NewHitMap()
    buildHitMapRecursive(root, offsetX, offsetY, hitMap)
    return hitMap
}

// buildHitMapRecursive 递归构建 HitMap
func buildHitMapRecursive(node VNode, offsetX, offsetY int, hitMap *HitMap) {
    // 1. 检查节点是否有 bounds（ElementVNode 有，ComponentVNode 可能没有）
    if boundsGetter, ok := node.(interface{ GetBounds() [4]int }); ok {
        bounds := boundsGetter.GetBounds()
        x, y, w, h := bounds[0], bounds[1], bounds[2], bounds[3]

        // 2. 计算绝对坐标
        absX := offsetX + x
        absY := offsetY + y

        // 3. 获取或生成节点 ID
        nodeID := getNodeID(node)

        // 4. 添加到 HitMap
        entry := &HitMapEntry{
            NodeID: nodeID,
            Bounds: Rect{X: absX, Y: absY, Width: w, Height: h},
            LocalXY: func(globalX, globalY int) (int, int) {
                return globalX - absX, globalY - absY
            },
        }
        hitMap.Add(entry)
    }

    // 5. 递归处理子节点（累积偏移）
    for _, child := range node.Children() {
        childOffsetX := offsetX
        childOffsetY := offsetY

        // 如果子节点有 bounds，累加偏移
        if boundsGetter, ok := child.(interface{ GetBounds() [4]int }); ok {
            bounds := boundsGetter.GetBounds()
            childOffsetX += bounds[0]
            childOffsetY += bounds[1]
        }

        buildHitMapRecursive(child, childOffsetX, childOffsetY, hitMap)
    }
}

// getNodeID 获取或生成节点 ID
func getNodeID(node VNode) string {
    // 优先使用 key
    if key := node.Key(); key != "" {
        return key
    }

    // 否则使用 tag
    if tag := node.Tag(); tag != "" {
        return tag
    }

    // 否则生成唯一 ID
    return fmt.Sprintf("node-%p", node)
}
```

### 集成到 App.render()

```go
// framework/app.go
func (a *App) render() {
    // ... 现有渲染逻辑 ...

    // ============================================================================
    // Phase 1: 构建 HitMap（从 VNode 树）
    // ============================================================================
    if a.root != nil {
        // 获取 VNode 树
        var vnodeTree rtui.VNode
        if painter, ok := a.root.(interface{ Render() rtui.VNode }); ok {
            vnodeTree = painter.Render()
        }

        if vnodeTree != nil {
            // 从 VNode 树构建 HitMap（包括 Fragment/Inspector overlay）
            a.hitMap = runtimeevent.BuildHitMapFromVNode(vnodeTree, 0, 0)

            // Phase 1-6: 将 HitMap 传递给 Pump
            if a.pump != nil {
                a.pump.SetHitMap(a.hitMap)
            }

            // DEBUG: 输出 HitMap 统计信息
            if os.Getenv("TUI_DEBUG_HITMAP") == "true" {
                fmt.Fprintf(os.Stderr, "[APP] HitMap built from VNode: %d entries\n", a.hitMap.Size())
            }
        }
    }

    // ...
}
```

---

## 📝 详细实施方案

### Phase 0: 准备工作（1天）

#### 任务 0.1: 分析当前代码结构

**目标**：确认所有相关文件和依赖关系。

**文件清单**：
- `runtime/event/hitmap.go` - HitMap 数据结构
- `runtime/event/hittest.go` - 当前命中测试实现
- `framework/app.go` - App 主循环和 HitMap 集成
- `framework/event/pump.go` - 事件泵
- `runtime/ui/vnode.go` - VNode 接口定义
- `runtime/ui/element.go` - ElementVNode 实现
- `runtime/ui/fragment.go` - Fragment 实现
- `internal/inspector/hook.go` - Inspector 注入 Hook
- `internal/inspector/standalone_inspector.go` - Inspector 事件处理

**依赖关系**：
```
framework/app.go
  └─> runtime/event (构建 HitMap)
       └─> runtime/ui (访问 VNode 接口)
            └─> framework (component.Node)

⚠️ 循环依赖风险：
framework → runtime/event → runtime/ui → framework
```

#### 任务 0.2: 确认布局引擎 bounds 设置

**目标**：验证布局引擎正确设置了 VNode.bounds 字段。

**验证步骤**：
1. 在 `runtime/layout/layout.go` 中搜索 `bounds` 字段赋值
2. 在 `runtime/ui/element.go` 中确认 `GetBounds()` 方法存在
3. 添加调试日志输出 bounds 信息

**验收标准**：
- [ ] 确认 ElementVNode 有 `GetBounds() [4]int` 方法
- [ ] 确认布局引擎在计算后设置 bounds
- [ ] 确认 Fragment 子节点的 bounds 被正确设置

### Phase 1: 实现 VNode HitMap 构建（2-3天）

#### 任务 1.1: 创建 VNode HitMap 构建函数

**文件**：`runtime/event/hitmap_vnode.go` (新建)

**代码**：见上文"代码结构"部分。

**关键点**：
1. 直接从 VNode 接口提取 bounds
2. 递归遍历 Children() []VNode
3. 累积父节点的偏移量
4. 处理 Fragment 的多个子节点

**验收标准**：
- [ ] 函数能正确处理 ElementVNode
- [ ] 函数能正确处理 Fragment
- [ ] 函数能正确处理 ComponentVNode
- [ ] 单元测试覆盖所有 VNode 类型

#### 任务 1.2: 扩展 HitMap 支持 VNode ID

**文件**：`runtime/event/hitmap.go`

**修改**：
```go
// HitMapEntry 添加可选的 VNode 引用
type HitMapEntry struct {
    NodeID  string
    Bounds  Rect
    LocalXY func(int, int) (int, int)
    VNode   ui.VNode  // 可选：用于后续优化
}
```

**验收标准**：
- [ ] HitMapEntry 可以存储 VNode 引用（可选）
- [ ] 不影响现有 layout.Node 的 HitMap 构建

#### 任务 1.3: 集成到 App.render()

**文件**：`framework/app.go:1015-1069`

**修改**：
```go
// 替换现有的 VNodeAdapter 逻辑
if vnodeTree := a.getVNodeTree(); vnodeTree != nil {
    a.hitMap = runtimeevent.BuildHitMapFromVNode(vnodeTree, 0, 0)

    if a.pump != nil {
        a.pump.SetHitMap(a.hitMap)
    }

    if os.Getenv("TUI_DEBUG_HITMAP") == "true" {
        fmt.Fprintf(os.Stderr, "[APP] HitMap built from VNode: %d entries\n", a.hitMap.Size())
    }
}

// 辅助方法
func (a *App) getVNodeTree() ui.VNode {
    if painter, ok := a.root.(interface{ Render() ui.VNode }); ok {
        return painter.Render()
    }
    return nil
}
```

**验收标准**：
- [ ] App.render() 成功构建 HitMap
- [ ] HitMap 包含 Inspector overlay 的节点
- [ ] 调试日志显示正确的节点数量

#### 任务 1.4: 单元测试

**文件**：`runtime/event/hitmap_vnode_test.go` (新建)

**测试用例**：
```go
func TestBuildHitMapFromVNode_ElementVNode(t *testing.T) {
    // 创建带有 bounds 的 ElementVNode
    // 构建 HitMap
    // 验证 entry 存在且 bounds 正确
}

func TestBuildHitMapFromVNode_Fragment(t *testing.T) {
    // 创建 Fragment(appNode, inspectorNode)
    // 构建 HitMap
    // 验证两个子节点都在 HitMap 中
}

func TestBuildHitMapFromVNode_Nested(t *testing.T) {
    // 创建深层嵌套的 VNode 树
    // 验证偏移量计算正确
}
```

**验收标准**：
- [ ] 所有测试通过
- [ ] 覆盖率 > 80%

### Phase 2: 清理手动事件处理（1-2天）

#### 任务 2.1: 移除 Inspector 的手动坐标解析

**文件**：`internal/inspector/standalone_inspector.go:1715-1757`

**修改**：
```go
// ❌ 删除这段代码：
/*
if ev.X >= minX && ev.X < maxX && ev.Y >= minY && ev.Y < maxY {
    localX := ev.X - minX
    localY := ev.Y - minY
    // ...
}
*/

// ✅ Inspector 现在直接从 Pump 接收正确的事件
func (si *StandaloneInspector) HandleEvent(ev event.Event) bool {
    if me, ok := ev.(*frameworkevent.MouseEvent); ok {
        // Pump 已经设置了 TargetID 和 LocalX/LocalY
        // Inspector 直接使用这些信息
        return si.handleMouseEvent(me)
    }
    return false
}

func (si *StandaloneInspector) handleMouseEvent(me *frameworkevent.MouseEvent) bool {
    // 使用 me.TargetID 识别目标组件
    // 使用 me.LocalX/LocalY 作为本地坐标
    // 无需手动计算！
}
```

**验收标准**：
- [ ] Inspector 的 HandleEvent 不再手动计算坐标
- [ ] TreeView 正确响应点击
- [ ] 点击位置与选中节点匹配

#### 任务 2.2: 移除 Panel 的事件转发

**文件**：`components/container/panel.go:27-45`

**修改**：
```go
// ❌ 删除 HandleEvent 方法
/*
func (p *Panel) HandleEvent(ev event.Event) bool {
    // ...
}
*/

// ✅ Panel 现在不需要手动转发事件
// Router 根据 TargetID 直接分发到内容组件
```

**验收标准**：
- [ ] Panel 不再实现 HandleEvent
- [ ] Panel 内的组件（如 Tabs）正常接收事件

#### 任务 2.3: 简化 Tabs 的事件处理

**文件**：`components/navigation/tabs.go:198-223`

**修改**：
```go
func (t *TabsVNode) HandleEvent(ev frameworkevent.Event) bool {
    me, ok := ev.(*frameworkevent.MouseEvent)
    if !ok || ev.Type() != frameworkevent.EventMousePress {
        return false
    }

    // 使用 ev.TargetID 识别点击的是 tab 栏还是内容区
    if strings.HasPrefix(me.TargetID, "tab-bar-") {
        return t.handleTabBarClick(me.LocalX, me.LocalY)
    }

    // 使用 ev.TargetID 直接分发到激活的 tab 内容
    // 无需手动转发！
    return false
}
```

**验收标准**：
- [ ] Tabs 不再手动转发到内容组件
- [ ] Tab 栏点击正常工作
- [ ] Tab 内容组件正常接收事件

### Phase 3: 验证和测试（1天）

#### 任务 3.1: 集成测试

**测试场景**：
1. 打开/关闭 Inspector (F12)
2. 点击 Inspector 的 TreeView
3. 验证选中正确的节点
4. 点击 Inspector 的其他组件（Tabs, Input, Button）
5. 验证所有组件正常响应

**验收标准**：
- [ ] Inspector TreeView 点击正确
- [ ] Inspector 所有组件响应正常
- [ ] 主应用组件不受影响

#### 任务 3.2: 性能测试

**测试指标**：
- HitMap 构建时间
- HitTest 命中时间
- 内存占用

**验收标准**：
- [ ] HitMap 构建时间 < 1ms (1000 节点)
- [ ] HitTest 命中时间 < 0.1ms
- [ ] 内存增长 < 10%

#### 任务 3.3: 调试工具

**功能**：
- `TUI_DEBUG_HITMAP=true` 可视化 HitMap
- 显示每个节点的 bounds
- 显示命中测试结果

**验收标准**：
- [ ] 调试输出清晰易读
- [ ] 帮助快速定位问题

---

## 📊 预期收益

### 功能性收益

| 指标 | 当前状态 | 目标状态 | 提升 |
|------|---------|---------|------|
| Inspector 鼠标事件成功率 | ~30% | 100% | +233% |
| 点击位置准确性 | 经常错误 | 完全准确 | ∞ |
| 组件需手写 bounds | 每个组件 | 0 个 | -100% |
| 手动坐标解析代码 | ~100 行 | 0 行 | -100% |

### 架构性收益

- ✅ **统一接口**：所有组件通过 HitMap 接收事件，无特殊处理
- ✅ **符合设计原则**："组件不再手写命中"
- ✅ **可扩展性**：新增 overlay 组件无需修改事件系统
- ✅ **可测试性**：可按节点 ID 注入事件进行测试

### 性能收益

- ⚡ **减少遍历**：Router 根据 TargetID 直接分发，无需层次遍历
- ⚡ **缓存友好**：HitMap 结构紧凑，命中测试高效
- ⚡ **内存可控**：每帧重建 HitMap，无泄漏风险

---

## 🚨 风险与缓解

### 风险 1: ComponentVNode 无 bounds 字段

**风险描述**：
- ComponentVNode 是包装器，可能没有 bounds 字段
- 布局引擎可能只设置 ElementVNode 的 bounds

**缓解措施**：
1. 在 `buildHitMapRecursive` 中检查节点是否实现 `GetBounds()`
2. 如果没有 bounds，跳过该节点，继续递归子节点
3. 确保组件内的 VNode 元素有正确的 bounds

**代码**：
```go
func buildHitMapRecursive(node VNode, offsetX, offsetY int, hitMap *HitMap) {
    // 只有有 bounds 的节点才添加到 HitMap
    if boundsGetter, ok := node.(interface{ GetBounds() [4]int }); ok {
        bounds := boundsGetter.GetBounds()
        // 添加到 HitMap...
    }

    // 无论有没有 bounds，都递归子节点
    for _, child := range node.Children() {
        buildHitMapRecursive(child, offsetX, offsetY, hitMap)
    }
}
```

### 风险 2: 偏移量计算错误

**风险描述**：
- 多层嵌套时偏移量可能累积错误
- Flex、Stack 等布局可能有特殊偏移规则

**缓解措施**：
1. 使用布局引擎设置的绝对坐标（bounds[0], bounds[1]）
2. 不手动累积偏移，直接使用 bounds 的绝对坐标
3. 单元测试覆盖深层嵌套场景

**代码**：
```go
// ✅ 使用绝对坐标（bounds 已是屏幕坐标）
absX := bounds[0]
absY := bounds[1]

// ❌ 不要累积偏移
// absX := offsetX + bounds[0]  // 错误！
```

### 风险 3: Fragment 子节点坐标问题

**风险描述**：
- Fragment 的子节点可能有不同的坐标原点
- Inspector overlay 的坐标系统可能与主应用不同

**缓解措施**：
1. 确保 Inspector 也是一个 VNode，有正确的 bounds
2. 在 Hook 中为 Inspector 设置正确的位置和大小
3. 测试覆盖 Fragment 场景

**代码**：
```go
// internal/inspector/hook.go
func CreateInspectorHook(inspector *StandaloneInspector) render.VNodeHook {
    return func(vnode rtui.VNode) rtui.VNode {
        if !inspector.IsVisible() {
            return vnode
        }

        inspectorContent := inspector.RenderContent()
        inspectorContent.SetLayer(rtui.LayerInspector)

        // 确保 Inspector 有正确的 bounds
        // （布局引擎会设置）

        return rtui.Fragment(vnode, inspectorContent)
    }
}
```

### 风险 4: 循环导入

**风险描述**：
- `runtime/event` 导入 `runtime/ui`
- `runtime/ui` 导入 `framework`
- `framework` 导入 `runtime/event`

**缓解措施**：
1. **关键**：`runtime/event` 不导入 `runtime/ui`
2. 使用接口解耦：定义 `BoundsNode` 接口在 `runtime/event`
3. `runtime/ui` 实现 `BoundsNode` 接口

**代码**：
```go
// runtime/event/hitmap_vnode.go
package runtimeevent

// 定义最小接口，避免导入 runtime/ui
type BoundsNode interface {
    Children() []BoundsNode  // 注意：返回自己的接口类型
    GetBounds() [4]int
    Key() string
    Tag() string
}

// BuildHitMapFromVNode 使用 BoundsNode 接口
func BuildHitMapFromVNode(root BoundsNode, offsetX, offsetY int) *HitMap {
    // ...
}

// runtime/ui/vnode.go
package ui

// VNode 实现 BoundsNode 接口
// （已经实现了所有方法）
```

---

## 📅 时间表

| 阶段 | 任务 | 工期 | 负责人 | 状态 |
|------|------|------|--------|------|
| Phase 0 | 准备工作 | 1天 | - | 待开始 |
| Phase 1.1 | 创建 VNode HitMap 构建 | 1天 | - | 待开始 |
| Phase 1.2 | 扩展 HitMap | 0.5天 | - | 待开始 |
| Phase 1.3 | 集成到 App | 0.5天 | - | 待开始 |
| Phase 1.4 | 单元测试 | 1天 | - | 待开始 |
| Phase 2.1 | 清理 Inspector | 0.5天 | - | 待开始 |
| Phase 2.2 | 清理 Panel | 0.5天 | - | 待开始 |
| Phase 2.3 | 清理 Tabs | 0.5天 | - | 待开始 |
| Phase 3.1 | 集成测试 | 0.5天 | - | 待开始 |
| Phase 3.2 | 性能测试 | 0.5天 | - | 待开始 |
| Phase 3.3 | 调试工具 | 0.5天 | - | 待开始 |
| **总计** | | **5-6天** | | |

---

## ✅ 验收标准

### 功能验收

- [ ] Inspector TreeView 点击正确选中节点
- [ ] Inspector 所有组件（TreeView, Tabs, Input, Button）响应正常
- [ ] 主应用组件事件处理不受影响
- [ ] 点击位置与反馈完全匹配

### 代码质量验收

- [ ] 无手动坐标解析代码
- [ ] 无事件转发代码
- [ ] 单元测试覆盖率 > 80%
- [ ] 代码审查通过

### 性能验收

- [ ] HitMap 构建时间 < 1ms (1000 节点)
- [ ] HitTest 命中时间 < 0.1ms
- [ ] 内存增长 < 10%
- [ ] FPS 保持稳定（60 FPS）

### 文档验收

- [ ] API 文档更新
- [ ] 架构文档更新
- [ ] 示例代码更新

---

## 🔗 相关文档

### 设计文档
- [long_term_event_architecture.md](../../event/long_term_event_architecture.md) - 长期架构规划
- [MSG_UNIFICATION.md](./MSG_UNIFICATION.md) - Msg 统一设计
- [README.md](./README.md) - 事件系统重构索引

### 实现文档
- [IMPLEMENTATION_PLAN.md](./IMPLEMENTATION_PLAN.md) - 原实施方案
- [TASKS.md](./TASKS.md) - 任务列表

### 代码文件
- [runtime/event/hitmap.go](../../runtime/event/hitmap.go) - HitMap 实现
- [framework/app.go](../../framework/app.go) - App 主循环
- [framework/event/pump.go](../../framework/event/pump.go) - 事件泵
- [internal/inspector/hook.go](../../internal/inspector/hook.go) - Inspector Hook
- [internal/inspector/standalone_inspector.go](../../internal/inspector/standalone_inspector.go) - Inspector 实现

---

## 📞 后续支持

### 实施期间支持

如遇到问题，请检查：
1. 调试日志：`TUI_DEBUG_HITMAP=true`
2. HitMap 可视化工具
3. 单元测试输出

### 实施后优化

完成本方案后，可考虑：
1. **Phase 4**：增量更新 HitMap（仅更新变化节点）
2. **Phase 5**：空间索引优化（四叉树加速 HitTest）
3. **Phase 6**：完整 Msg/Cmd 系统（见 MSG_UNIFICATION.md）

---

**文档版本**: v1.0
**最后更新**: 2026-02-10
**状态**: 待审核
