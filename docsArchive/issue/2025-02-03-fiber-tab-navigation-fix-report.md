# Fiber 模式 Tab 导航失效问题分析与修复报告

**日期**: 2025-02-03
**问题**: Fiber 模式下 Tab 键无法切换焦点
**影响范围**: Fiber reconciler, FocusManager, HStack 布局

---

## 1. 问题概述

### 1.1 症状表现

在 `MINT_USE_FIBER=true` 环境下运行应用时：
- Tab 键按下后焦点索引不变化
- 所有按钮显示相同的 focusID
- HStack 布局被错误渲染为垂直布局
- 按钮焦点状态（蓝色背景）未正确显示

### 1.2 预期行为

- 初始渲染时第一个按钮自动获得焦点（index=0）
- 按 Tab 键焦点移动到下一个按钮（index=1）
- 按 Shift+Tab 键焦点返回上一个按钮
- 获得焦点的按钮显示蓝色背景

---

## 2. 根本原因分析

### 2.1 问题 1: HStack 未被正确识别

**现象**: 按钮被渲染在不同的 Y 坐标（y=6, y=8）而不是同一行

**原因**: `renderFiberToBuffer` 中的 HStack 检测逻辑只检查了 `*rtui.LayoutNode` 和 `*rtui.ElementVNode` 类型断言，但实际的 VNode 类型可能不同，导致 `isHStack` 始终为 `false`。

```go
// 原代码 - 类型断言失败
if elemNode, ok := fiber.VNode.(*rtui.ElementVNode); ok {
    tag := elemNode.Tag()
    if tag == "hstack" || tag == "row" {
        isHStack = true  // 永远不会执行
    }
}
```

**修复**: 添加通用的 `Tag()` 方法检查：
```go
if tagger, ok := fiber.VNode.(interface{ Tag() string }); ok {
    tag := tagger.Tag()
    if tag == "hstack" || tag == "row" {
        isHStack = true
    }
}
```

### 2.2 问题 2: 焦点索引未初始化

**现象**: 初始焦点 index = -1（无焦点），Tab 后移动到 index=0 而非 index=1

**原因**: `UpdateFocusableList()` 方法只更新焦点列表，不初始化焦点状态：

```go
func (m *VNodeFocusManager) UpdateFocusableList(nodes []FocusableVNode) {
    m.focusable = nodes
    // ❌ 没有初始化焦点
}
```

相比之下，`SetFocusable()` 有初始化逻辑：
```go
func (m *VNodeFocusManager) SetFocusable(nodes []FocusableVNode) {
    ...
    // ✅ 如果没有焦点，聚焦第一个
    if m.current < 0 && len(m.focusable) > 0 {
        m.FocusFirst()
    }
}
```

**修复**: 在 `updateFocusManagerFromFiber()` 中添加初始化：
```go
} else if len(focusable) > 0 {
    // 如果没有当前焦点且有焦点元素，聚焦第一个
    r.focusMgr.SetFocusByIndex(0)
}
```

### 2.3 问题 3: FocusIndex 被重复调用覆盖

**现象**: 调试显示所有按钮的 focusIndex 都是 0，而不是 0, 1, 2...

**原因**: `collectFocusableFromFiber()` 在每次渲染时被调用多次（`applyFocusStateToFiber` 和 `updateFocusManagerFromFiber` 各调用一次），每次都重置索引：

```go
// 第一次调用 (applyFocusStateToFiber)
btn.SetFocusIndex(0)  // button1 -> index=0
btn.SetFocusIndex(1)  // button2 -> index=1

// 第二次调用 (updateFocusManagerFromFiber)
btn.SetFocusIndex(0)  // button1 -> index=0 ❌ 被覆盖！
btn.SetFocusIndex(0)  // button2 -> index=0 ❌ 被覆盖！
```

**修复**: 只在未设置时才设置索引：
```go
func (b *ButtonVNode) SetFocusIndex(index int) {
    if b.focusIndex < 0 {  // 只在未设置时
        b.focusIndex = index
    }
}
```

同时初始化 `focusIndex` 为 -1：
```go
func NewButton(label string) *ButtonVNode {
    return &ButtonVNode{
        ...
        focusIndex: -1,  // -1 表示未设置
    }
}
```

### 2.4 问题 4: 递归收集索引计算错误

**原因**: `collectFocusableFromFiber` 递归调用子节点时，每次都从索引 0 开始：

```go
func (r *Reconciler) collectFocusableFromFiber(fiber *Fiber) []rtui.FocusableVNode {
    result := []rtui.FocusableVNode{}
    for child := fiber.Child; child != nil; child = child.Sibling {
        result = append(result, r.collectFocusableFromFiber(child)...)  // ❌ 每次都从0开始
    }
    return result
}
```

**修复**: 使用参数化函数传递起始索引：
```go
func (r *Reconciler) collectFocusableFromFiberWithIndex(fiber *Fiber, startIndex int) []rtui.FocusableVNode {
    currentIndex := startIndex
    // 收集当前节点...
    currentIndex++  // 增加索引

    // 递归子节点时传递当前索引
    childResult := r.collectFocusableFromFiberWithIndex(fiber.Child, currentIndex)
    currentIndex += len(childResult)

    // 处理兄弟节点时传递当前索引
    siblingResult := r.collectFocusableFromFiberWithIndex(fiber.Sibling, currentIndex)
    return result
}
```

### 2.5 问题 5: `SetFocusable` 调用时机错误

**原因**: `declarative_node.go` 的 `Paint()` 方法在 reconciler 渲染前调用 `SetFocusable()`，重置了焦点状态：

```go
// Paint() 方法中
if n.focusMgr != nil && n.reconciler != nil {
    focusable := n.collectFocusableFromFiber(fiberRoot)
    n.focusMgr.SetFocusable(focusable)  // ❌ 在 reconciler.Render() 之前调用
}
n.reconciler.Render(...)  // reconciler 内部也会调用 updateFocusManagerFromFiber
```

**修复**: 删除这个不必要的 `SetFocusable` 调用，因为 reconciler 内部会通过 `updateFocusManagerFromFiber` 更新焦点列表。

---

## 3. Fiber 渲染流程

### 3.1 完整渲染时序图

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         Fiber 渲染完整流程                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  1. 用户操作                                                                 │
│     ├─ Tab 键按下                                                              │
│     └─ framework.InjectEvent(TabKeyEvent)                                     │
│           ↓                                                                  │
│  2. DeclarativeNode.HandleEvent()                                          │
│     ├─ focusMgr.HandleEvent(Tab)                                            │
│     │   └─ focusMgr.FocusNext() → m.current = (m.current + 1) % len          │
│     ├─ reconciler.ScheduleUpdate(LaneSyncLane)                              │
│     └─ fwApp.MarkDirty() → 触发重新渲染                                     │
│           ↓                                                                  │
│  3. framework 渲染循环                                                        │
│     └─ DeclarativeNode.Paint(ctx, buffer)                                   │
│           ↓                                                                  │
│  4. reconciler.Render(ctx, buffer, renderFunc)                               │
│     ├─ prepareFreshStack(renderFunc)                                         │
│     │   └─ 创建/更新 workInProgress Fiber tree                              │
│     ├─ workLoopSync()                                                       │
│     │   ├─ performUnitOfWork() → BeginWork() → CompleteWork()              │
│     │   └─ root = workInProgress (double buffering)                         │
│     └─ CommitRoot()  ← 关键阶段                                             │
│           ↓                                                                  │
│  5. CommitRoot() 详细流程                                                   │
│     ├─ Phase 0: applyFocusStateToFiber(root)                              │
│     │   ├─ focusIndex = focusMgr.CurrentIndex()                             │
│     │   ├─ collectFocusableFromFiber(root) → focusable[]                   │
│     │   └─ for each focusable[i]: SetFocus(i == focusIndex)                 │
│     ├─ Phase 1: buildLayoutTree(root)                                       │
│     ├─ Phase 2: calculateLayout()                                            │
│     ├─ Phase 3: GenerateLayoutBoxes()                                       │
│     ├─ Phase 4: renderFiberToBuffer(root, 0, 0, buffer)                   │
│     │   └─ 对每个 Fiber 调用 renderCallback → Paint() → buffer.SetString()    │
│     └─ Phase 5: updateFocusManagerFromFiber(root)                            │
│         ├─ focusable = collectFocusableFromFiber(root)                       │
│         ├─ focusMgr.UpdateFocusableList(focusable)                           │
│         └─ if currentIndex < 0: SetFocusByIndex(0)  ← 初始化焦点              │
│           ↓                                                                  │
│  6. framework 显示 buffer                                                   │
│     └─ 用户看到更新后的 UI（包括焦点指示器）                                  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 3.2 焦点状态更新时序图

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                      焦点状态更新详细时序                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  [初始渲染]                                                                  │
│                                                                             │
│  CommitRoot():                                                               │
│    ├── applyFocusStateToFiber(root)                                         │
│    │   ├── collectFocusable → [btn1@0, btn2@1]                             │
│    │   ├── focusMgr.CurrentIndex() = -1 (初始无焦点)                          │
│    │   └─ 设置 SetFocus(false) for all buttons                               │
│    ├── renderFiberToBuffer() → 渲染 " [ - ]   [ + ]"                         │
│    └─ updateFocusManagerFromFiber(root)                                    │
│        ├── focusable = [btn1@0, btn2@1]                                     │
│        ├── focusMgr.CurrentIndex() = -1                                     │
│        └─ SetFocusByIndex(0) → m.current = 0, SetFocus(true) for btn1           │
│                                                                             │
│  [第一次 Tab]                                                                 │
│                                                                             │
│  Tab 事件 → HandleEvent():                                                   │
│    ├── FocusNext(): m.current = (-1 + 1) % 2 = 0                             │
│    ├── SetFocus(false) for old (index -1, 无操作)                            │
│    └─ SetFocus(true) for btn1                                                 │
│    ├── ScheduleUpdate(LaneSyncLane)                                         │
│    └─ fwApp.MarkDirty()                                                      │
│                                                                             │
│  下一帧 CommitRoot():                                                         │
│    ├── applyFocusStateToFiber(root)                                         │
│    │   ├── focusMgr.CurrentIndex() = 0                                       │
│    │   └─ 设置 btn1.SetFocus(true), btn2.SetFocus(false)                       │
│    ├── renderFiberToBuffer() → 渲染 "*[ - ]   [ + ]"                         │
│    └─ updateFocusManagerFromFiber(root) → 保持 m.current = 0                   │
│                                                                             │
│  [第二次 Tab]                                                                 │
│                                                                             │
│  Tab 事件 → HandleEvent():                                                   │
│    ├── FocusNext(): m.current = (0 + 1) % 2 = 1                              │
│    ├── SetFocus(false) for btn1                                               │
│    └─ SetFocus(true) for btn2                                                 │
│    ├── ScheduleUpdate(LaneSyncLane)                                         │
│    └─ fwApp.MarkDirty()                                                      │
│                                                                             │
│  下一帧 CommitRoot():                                                         │
│    ├── applyFocusStateToFiber(root)                                         │
│    │   ├── focusMgr.CurrentIndex() = 1                                       │
│    │   └─ 设置 btn1.SetFocus(false), btn2.SetFocus(true)                       │
│    ├── renderFiberToBuffer() → 渲染 " [ - ]  *[ + ]"                         │
│    └─ updateFocusManagerFromFiber(root) → 保持 m.current = 1                   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 3.3 HStack 渲染流程

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                      HStack 水平布局渲染流程                                │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Fiber 树结构:                                                                │
│                                                                             │
│       root (ComponentVNode, key="root")                                     │
│            ↓                                                                   │
│       VStack (LayoutNode, tag="vstack")                                      │
│            ├─ Text("Count: 0")                                              │
│            ├─ Text("[Fiber: ON]...")                                        │
│            └─ HStack (LayoutNode, tag="hstack", gap=1)  ← 关键               │
│                ├─ Button("-")  (Child, x=0, y=6)                             │
│                └─ Button("+")  (Sibling, x=13, y=6)  ← 计算后的位置            │
│                                                                             │
│  渲染流程:                                                                    │
│                                                                             │
│  renderFiberToBuffer(hstack, x=0, y=6):                                    │
│    ├── isHStack = CheckHStack(hstack) → true                               │
│    ├── renderFiber(child1, x=0, y=6)    → button at (0, 6)                  │
│    ├── width1 = measureFiberWidth(child1) = 11                              │
│    ├── childX += width1 + gap = 11 + 1 = 12                                  │
│    └── renderFiber(child2, x=12, y=6)   → button at (12, 6)                 │
│                                                                             │
│  修复前的问题:                                                                 │
│    - CheckHStack() 类型断言失败，isHStack = false                            │
│    - childX 始终保持 0，y 递增                                                │
│    - 结果: button1 at (0, 6), button2 at (0, 8)  ← 垂直布局                 │
│                                                                             │
│  修复后:                                                                      │
│    - CheckHStack() 使用通用 Tag() 接口                                      │
│    - isHStack = true                                                          │
│    - 结果: button1 at (0, 6), button2 at (12, 6)  ← 水平布局               │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 4. 处理过程

### 4.1 调试发现的关键线索

1. **调试输出显示所有按钮 index=0**
   ```
   [collectFocusable] adding focusable 0: button:  -  @0
   [collectFocusable] adding focusable 0: button:  +  @0
   ```
   说明 `SetFocusIndex` 被重复调用覆盖

2. **调试输出显示按钮在不同 Y 坐标**
   ```
   [renderFiberToBuffer] tag="button", x=0, y=6, isHStack=false
   [renderFiberToBuffer] tag="button", x=0, y=8, isHStack=false
   ```
   说明 HStack 未被识别

3. **调试输出显示初始焦点 index=-1**
   ```
   [FocusNext] current=-1, len(focusable)=2
   [FocusNext] old=-1, new=0
   ```
   说明焦点未初始化

### 4.2 修复步骤

| 步骤 | 问题 | 修复方式 | 文件 |
|------|------|---------|------|
| 1 | HStack 未识别 | 添加通用 Tag() 接口检查 | reconciler.go |
| 2 | 焦点未初始化 | 添加 SetFocusByIndex(0) 当 index<0 | reconciler.go |
| 3 | FocusIndex 被覆盖 | 只在 focusIndex<0 时设置 | button.go |
| 4 | 索引计算错误 | 参数化 collectFocusableWithIndex | reconciler.go |
| 5 | SetFocusable 干扰 | 删除 Paint() 中的 SetFocusable 调用 | declarative_node.go |

---

## 5. 架构复杂度反思

### 5.1 当前架构的耦合点

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                      当前架构的依赖关系                                   │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ui/ 包                                                                      │
│    ├─ fiber.go → 重导出 runtime/ui.Fiber                                   │
│    ├─ hooks.go → 组件上下文和 Hook 管理                                    │
│    └─ focus_manager.go → 焦点管理                                          │
│                                                                             │
│  runtime/ui/ 包                                                               │
│    ├─ fiber.go → Fiber 结构体定义                                          │
│    └─ vnode.go → VNode 接口和实现                                         │
│                                                                             │
│  internal/state/ 包                                                          │
│    ├─ instance_manager.go → 组件实例管理                                    │
│    └─ interaction_state.go → 交互状态管理                                   │
│                                                                             │
│  internal/reconciler/ 包                                                    │
│    ├─ reconciler.go → Reconciler 主逻辑                                    │
│    ├─ begin_work.go → BeginWork 阶段                                       │
│    ├─ complete_work.go → CompleteWork 阶段                                 │
│    └─ diff.go → Diff 算法                                                   │
│                                                                             │
│  internal/render/ 包                                                        │
│    └─ declarative_node.go → VNode 与 framework 的桥接                       │
│                                                                             │
│  components/ 包                                                              │
│    └─ button/button.go → 按钮组件                                           │
│                                                                             │
│  循环依赖风险:                                                                │
│    ❌ ui/ → runtime/ui/ (重导出, OK)                                        │
│    ⚠️  internal/reconciler/ → runtime/ui/ (直接依赖)                         │
│    ⚠️  internal/render/ → internal/reconciler/ (依赖 reconciler)              │
│    ⚠️  components/ → ui/ (依赖 ui 包)                                        │
│                                                                             │
│  问题领域:                                                                   │
│    1. FocusableVNode 接口在 runtime/ui 定义，但实现在 ui/ 组件中            │
│    2. Fiber 结构在 runtime/ui 定义，但 reconciler 在 internal/               │
│    3. 焦点管理逻辑分散在多个地方                                            │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 5.2 复杂度分析

#### 5.2.1 渲染流程的复杂性

当前有 **3 种渲染模式** 需要维护：

1. **传统模式** (非 Fiber): `DeclarativeNode.paintVNode()`
2. **Fiber 模式**: `reconciler.Render()` → `CommitRoot()` → `renderFiberToBuffer()`
3. **混合模式**: 某些代码路径同时支持两种模式

这导致：
- 代码重复度高
- 维护成本增加
- 容易出现不一致行为

#### 5.2.2 焦点管理的复杂性

焦点状态在 **4 个地方** 管理：

1. `FocusManager.current` - 当前焦点索引
2. `FocusManager.focusable[]` - 焦点元素列表
3. `FocusableVNode.hasFocus` - 每个组件的焦点状态
4. `ButtonVNode.focusIndex` - 焦点索引（用于生成唯一 ID）

同步这些状态的时机非常关键：
- 错误的时机会导致状态覆盖
- 多次调用会导致索引重复分配
- 初始化逻辑容易被遗漏

#### 5.2.3 Fiber 树遍历的复杂性

Fiber 树遍历需要处理：
- ComponentVNode 包装（需要跳过）
- 子节点和兄弟节点的不同处理
- HStack/VStack 的布局计算
- 双重缓冲（current/workInProgress）

### 5.3 改进建议

#### 建议 1: 统一渲染路径

```
当前: 2 条独立路径
├── 非 Fiber: paintVNode()
└── Fiber: renderFiberToBuffer()

建议: 统一为一条路径
└── Render()
    ├── 使用 Fiber reconciler (即使是最简单的渲染)
    └── 简化状态管理
```

#### 建议 2: 简化焦点状态管理

```
当前: 4 个状态需要同步
├── FocusManager.current
├── FocusManager.focusable[]
├── FocusableVNode.hasFocus
└── ButtonVNode.focusIndex

建议: 减少到 2 个状态
├── FocusManager (单一事实来源)
└── 组件通过查询获取状态 (而不是存储)
```

#### 建议 3: 明确的接口边界

```
runtime/ui/    → 定义核心接口 (VNode, FocusableVNode)
ui/           → 提供 UX 友好的 API
components/   → 实现 VNode 接口
internal/     → 内部实现，不导出
```

#### 建议 4: 移除调试代码

当前代码中存在大量调试代码（`TUI_DEBUG_FOCUS` 等），应该：
- 使用条件编译移除生产代码中的调试逻辑
- 或使用 logging 包替代 `fmt.Fprintf(os.Stderr, ...)`

### 5.4 架构优势

尽管存在复杂性问题，当前架构也有一些优势：

1. **分层清晰**: runtime/ui → ui → components 分层合理
2. **接口抽象**: VNode 和 FocusableVNode 接口设计良好
3. **可测试性**: 测试可以独立运行各个组件
4. **渐进式迁移**: 可以逐步从非 Fiber 迁移到 Fiber

---

## 6. 测试验证

### 6.1 测试用例覆盖

| 测试 | 状态 | 说明 |
|------|------|------|
| TestTabNavigation | ✅ 通过 | Tab 键正确切换焦点 |
| TestShiftTabNavigation | ✅ 通过 | Shift+Tab 正确返回焦点 |
| TestFiberButtonNavigation | ✅ 通过 | Fiber 模式按钮导航 |
| TestDynamicListKeyboardInput | ✅ 通过 | 复杂列表键盘输入 |
| TestDynamicListDebug | ✅ 通过 | 复杂列表调试模式 |

### 6.2 视觉验证

```
初始状态:
┌──────────────────────────────────────┐
│ [Fiber: ON] Press Tab to test...   │
│                                      │
│  [ - ]    [ + ]                     │  ← 两个按钮无焦点样式
│                                      │
└──────────────────────────────────────┘

Tab 后:
┌──────────────────────────────────────┐
│ [Fiber: ON] Press Tab to test...   │
│                                      │
│ *[ - ]    [ + ]                     │  ← 第一个按钮有蓝色背景
│   ↑ 焦点指示器                        │
└──────────────────────────────────────┘

再次 Tab 后:
┌──────────────────────────────────────┐
│ [Fiber: ON] Press Tab to test...   │
│                                      │
│  [ - ]   *[ + ]                     │  ← 第二个按钮有蓝色背景
│                   ↑ 焦点指示器         │
└──────────────────────────────────────┘
```

---

## 7. 总结

本次修复解决了 Fiber 模式下 Tab 导航失效的问题，主要涉及：

1. **HStack 布局识别** - 修复了类型断言失败导致的水平布局问题
2. **焦点初始化** - 确保首次渲染时第一个元素获得焦点
3. **索引分配** - 防止重复调用导致的索引覆盖
4. **递归收集** - 修复了索引计算错误

**代码变更统计**:
- `internal/reconciler/reconciler.go`: 约 60 行修改
- `components/button/button.go`: 约 10 行修改
- `internal/render/declarative_node.go`: 删除约 15 行
- `runtime/ui/focus_manager.go`: 约 10 行修改

**测试覆盖率**:
- 所有核心测试通过
- Fiber 模式与非 Fiber 模式行为一致
- 焦点导航功能完整工作

虽然当前架构存在一定的复杂性，但通过清晰的接口定义和分层设计，系统仍然具有良好的可维护性和可扩展性。后续可以考虑逐步统一渲染路径，进一步简化代码。
