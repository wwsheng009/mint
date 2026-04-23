# Mint 交互式 E2E 测试套件设计

## 1. 背景

当前仓库已经有 `ui/test.go`，它提供了几项非常有价值的能力：

- 通过 `RunTest` / `RunTestWithSandbox` 启动完整 `framework.App`
- 通过 `InjectKey` / `InjectMouse` / `InjectString` 注入输入事件
- 通过 `GetRenderString()`、`AssertRender()` 做文本级断言
- 通过 `GetFocusedIndex()` / `GetFocusedType()` 获取有限的焦点信息
- 通过 `GetFrameworkApp()` / `GetDeclarativeRoot()` 暴露底层对象用于高级测试

这些能力对于 smoke test、单组件交互验证、简单文本断言已经足够。

但是对于**强交互型程序**，当前测试工具仍有明显不足：

1. **消息链路不可观测**
   - 只能“注入输入”与“看最终文本”
   - 无法结构化断言 `RawInput → Msg → Action → Intent → State → Render`

2. **没有确定性的 idle/flush 语义**
   - 测试里通常要靠 `sleep` 或“多渲染几次”碰运气
   - 对异步消息、动画、延迟 intent、调度刷新不友好

3. **焦点信息太弱**
   - 只有 `focusedIndex` / `focusedType`
   - 缺少“当前聚焦的是哪个组件 / Fiber / locator / bounds / targetID”的断言面

4. **样式不可断言**
   - `GetRenderString()` 会丢失颜色、反色、bold、underline、border style 等信息
   - 无法验证 hover、selected、focus ring、错误态、禁用态等视觉语义

5. **缺少 selector / locator 抽象**
   - 只能按文本找结果，无法稳定定位业务组件
   - 对重复文案、动态列表、overlay、portal、虚拟化场景脆弱

6. **对交互性组件缺少 E2E 级验证机制**
   - 如 Input、Textarea、Select、Menu、Tabs、TreeView、Table、Modal、Drawer、Tooltip、Toast
   - 特别是：消息处理、焦点切换、命中、可视状态与最终意图是否一致

因此需要设计一套**面向交互型 TUI 的强 E2E 测试机制**。

---

## 2. 设计目标

### 2.1 核心目标

新的 E2E 测试套件应支持：

1. **完整交互回路测试**
   - 键盘
   - 鼠标
   - 输入法/字符流
   - hover / press / release / drag / wheel

2. **消息处理验证**
   - 验证事件是否被正确转换为 `Msg`
   - 验证 `Msg` 是否触发正确 `Action`
   - 验证 `Action` 是否进入正确 target / focused fiber / router
   - 验证 `Intent` 是否按预期发射、冒泡、进入 runtime

3. **焦点切换验证**
   - 当前焦点组件是谁
   - 焦点为何转移
   - 焦点是否符合 Tab / Shift+Tab / Arrow / Escape 等规则
   - overlay / modal / popup 的焦点边界是否正确

4. **样式与视觉语义验证**
   - 单元格样式
   - 当前选中行样式
   - focus / hover / active / disabled
   - border / separator / ellipsis / scroll indicator
   - layer / overlay / portal 显示效果

5. **确定性的异步测试**
   - 能等待“渲染稳定”
   - 能等待“Intent 已发出”
   - 能等待“焦点切换完成”
   - 能等待“指定条件满足”

6. **可扩展的 selector / locator 体系**
   - 按 componentID
   - 按业务 ID
   - 按 key
   - 按 targetID
   - 按文本
   - 按 tag/type/layer
   - 按 HitMap / bounds / focus state

### 2.2 非目标

这套设计暂时不追求：

- 替代现有所有单元测试
- 成为浏览器式截图对比系统
- 在第一阶段直接覆盖所有组件
- 立即把所有内部运行时都开放成公共 API

---

## 3. 设计原则

### 3.1 分层而非一把梭

不要直接做一个“大而全”的测试大对象。  
应拆成：

- **Driver**：注入交互
- **Probe**：采集运行时观测数据
- **Selector**：定位组件
- **Assert**：声明式断言
- **Await**：等待系统到稳定态
- **Trace**：记录事件/动作/意图/焦点/渲染时间线

### 3.2 尽量复用现有运行时暴露面

已有可复用能力包括：

- `framework.App.InjectEvent(...)`
- `framework.App.GetRenderer()`
- `render.DeclarativeNode.GetFocusManager()`
- `render.DeclarativeNode.GetHitMap()`
- `render.DeclarativeNode.GetFocusedIndex()`
- `render.DeclarativeNode.GetFocusedType()`
- 全局 `IntentRuntime`

新的 E2E 设计应优先站在这些 seam 上，而不是另起炉灶。

### 3.3 “文本断言”升级为“结构化断言”

文本断言仍保留，但它应该只是最底层的一种断言方式。  
E2E 套件更应该支持：

- 样式矩阵断言
- 焦点断言
- 消息链路断言
- 命中框断言
- 组件状态断言

### 3.4 保持确定性

交互 E2E 测试必须尽量减少：

- `sleep`
- 不可控动画
- 竞争条件
- 非确定时序

所以必须设计：

- `AwaitIdle()`
- `AwaitIntent(...)`
- `AwaitFocus(...)`
- `AwaitRenderStable(...)`
- 可选的**测试时钟 / 手动 tick**

---

## 4. 当前 `ui/test.go` 的能力与缺口

### 4.1 已有能力

`ui/test.go` 目前已经具备：

- 完整 App 启动
- 事件注入
- 渲染缓冲提取
- 文本断言
- 焦点索引/类型断言
- 底层 app / declarative root 暴露

这部分是新 E2E 套件的天然基础。

### 4.2 缺口分类

#### A. Probe 不足

缺少：

- Intent 记录器
- Action 记录器
- Msg 记录器
- Focus 迁移日志
- Cell style 快照
- HitMap 查询与 selector 对接

#### B. Await 不足

缺少：

- 等待 app 空闲
- 等待渲染稳定
- 等待 intent 发出
- 等待消息 drain 完成

#### C. Selector 不足

缺少：

- `ByComponentID("...")`
- `ByID("...")`
- `ByKey("...")`
- `ByText("...")`
- `ByTargetID("...")`
- `ByTag("input")`
- `Focused()`
- `At(x,y)`

#### D. Assert 不足

缺少：

- `AssertFocus(...)`
- `AssertStyle(...)`
- `AssertIntent(...)`
- `AssertAction(...)`
- `AssertBounds(...)`
- `AssertHit(...)`
- `AssertCell(...)`
- `AssertLayerVisible(...)`

---

## 5. 总体架构

建议新增一套分层架构：

```text
E2E Suite
├── Harness         # 启动、关闭、flush、截图、trace
├── Driver          # key/mouse/type/click/drag/scroll/tab 等交互
├── Selector        # 按 componentID/id/key/text/tag/focus/hitmap 定位
├── Probe
│   ├── RenderProbe # buffer/cell/style/hitmap/layer
│   ├── FocusProbe  # focus manager / focus transitions
│   ├── MsgProbe    # raw input / msg / action / intent / dispatch trace
│   └── StateProbe  # 可选的组件级 runtime snapshot
├── Await
│   ├── AwaitIdle
│   ├── AwaitRenderStable
│   ├── AwaitFocus
│   ├── AwaitIntent
│   └── Eventually
└── Assert
    ├── Text assertions
    ├── Style assertions
    ├── Focus assertions
    ├── Intent assertions
    ├── Action assertions
    └── Layout / bounds assertions
```

---

## 6. 核心对象设计

## 6.1 `E2EApp`

建议新增一个比 `TestableApp` 更高阶的对象，例如：

```go
type E2EApp struct {
    app          *framework.App
    root         *render.DeclarativeNode
    driver       *Driver
    selectors    *SelectorRegistry
    renderProbe  *RenderProbe
    focusProbe   *FocusProbe
    msgProbe     *MessageProbe
    await        *Awaiter
    trace        *TraceLog
    clock        TestClock
}
```

职责：

- 封装启动与关闭
- 封装 probe 安装
- 管理 trace 和 await
- 提供统一断言入口

### 6.1.1 建议 API

```go
app, err := e2e.Run(MyApp,
    e2e.WithViewport(80, 24),
    e2e.WithDeterministicClock(),
    e2e.WithIntentTrace(),
    e2e.WithActionTrace(),
    e2e.WithStyleSnapshot(),
)
defer app.Close()
```

---

## 6.2 `Driver`

负责注入交互动作，抽象出“用户行为”，而不是裸事件。

### 建议 API

```go
app.Driver().Key('a')
app.Driver().Special(platform.KeyTab)
app.Driver().Combo(platform.KeyTab, platform.ModShift)
app.Driver().Type("hello")
app.Driver().Click(e2e.ByComponentID("submit"))
app.Driver().ClickAt(12, 8)
app.Driver().DoubleClick(...)
app.Driver().Hover(...)
app.Driver().Drag(from, to)
app.Driver().WheelDown(...)
```

### 设计要点

1. 底层仍可复用 `InjectEvent`
2. 对上层暴露行为语义：
   - `TabToNextFocus()`
   - `ActivateFocused()`
   - `OpenOverlay()`
3. `Driver` 每个操作后可选自动：
   - `AwaitIdle()`
   - `RecordTraceStep(...)`

---

## 6.3 `Selector`

交互型 E2E 测试最重要的能力之一，是**稳定定位组件**。

### 建议 selector 类型

```go
e2e.ByComponentID("tree.search")
e2e.ByID("username-input")
e2e.ByKey("email")
e2e.ByText("Submit")
e2e.ByTag("input")
e2e.ByTargetID("action-123")
e2e.Focused()
e2e.At(10, 5)
```

### Selector 解析来源

1. `HitMap`
2. Fiber / DeclarativeNode 暴露的焦点信息
3. 渲染缓冲中的文本与样式
4. 组件 props 中的 `componentID` / `id` / `key`

### 设计建议

优先级建议：

1. `componentID`
2. `id`
3. `key`
4. `targetID`
5. `text`
6. 坐标

因为文本最脆弱，坐标最不稳定。

---

## 6.4 `RenderProbe`

当前 `GetRenderString()` 会丢掉绝大多数视觉语义。  
E2E 套件需要结构化渲染快照：

```go
type CellSnapshot struct {
    X, Y        int
    Cluster     string
    FG          style.Color
    BG          style.Color
    Bold        bool
    Italic      bool
    Underline   bool
    Reverse     bool
    Width       int
    Continuation bool
}

type FrameSnapshot struct {
    Width, Height int
    Cells         [][]CellSnapshot
    HitMap        *event.HitMap
    Focused       LocatorRef
    Timestamp     time.Time
}
```

### 它要支持的断言

- 某个 cell 的文本
- 某个 cell 的样式
- 某个 locator 的可见文本
- 某个 locator 是否处于 selected/focus/disabled 样式
- overlay / popup 是否真的在前景层渲染

### 样式断言示例

```go
app.AssertCellStyle(
    e2e.At(12, 8),
    e2e.StyleExpect{
        FG: style.Black,
        BG: style.BrightCyan,
        Bold: true,
    },
)
```

---

## 6.5 `FocusProbe`

焦点测试不能只靠索引。

建议暴露：

```go
type FocusSnapshot struct {
    CurrentFiberID   uint64
    CurrentComponentID string
    CurrentID        string
    CurrentKey       string
    CurrentTag       string
    CurrentBounds    image.Rect
    FocusIndex       int
    FocusType        int
}
```

### 目标

- 验证当前焦点是谁
- 验证焦点切换顺序
- 验证 modal / drawer / popup 打开后焦点是否被 trap
- 验证关闭 overlay 后焦点是否回归触发源

### 断言示例

```go
app.AssertFocus(e2e.ByComponentID("search-input"))
app.AssertFocusOrder(
    e2e.ByID("username"),
    e2e.ByID("password"),
    e2e.ByText("Submit"),
)
```

---

## 6.6 `MessageProbe`

这是当前系统最欠缺、但对交互组件最关键的部分。

建议记录至少 5 层信息：

1. `RawInput`
2. `Msg`
3. `Action`
4. `Intent`
5. Render / Focus side effect

### 建议结构

```go
type TraceEventKind string

const (
    TraceRawInput TraceEventKind = "raw_input"
    TraceMsg      TraceEventKind = "msg"
    TraceAction   TraceEventKind = "action"
    TraceIntent   TraceEventKind = "intent"
    TraceFocus    TraceEventKind = "focus"
    TraceRender   TraceEventKind = "render"
)

type TraceEvent struct {
    Step       int
    Kind       TraceEventKind
    Name       string
    Payload    any
    Timestamp  time.Time
}
```

### 它应该支持的验证

- 某次点击到底是否生成了 `ActionClick`
- 某个 key 是否先变成 `KeyMsg` 再变成 `ActionNavigateNext`
- 组件是否发出了预期 `Intent`
- 焦点切换是否发生在预期动作之后
- 某次输入是否导致了 render

### 典型用例

#### 消息处理验证

```go
app.Driver().Special(platform.KeyTab)
app.AwaitIdle()
app.AssertAction("navigate_next")
app.AssertFocus(e2e.ByID("password"))
```

#### Intent 验证

```go
app.Driver().Click(e2e.ByText("Next"))
app.AwaitIntent("pagination.PageChangeIntent")
app.AssertLastIntent(func(i pagination.PageChangeIntent) {
    require.Equal(t, 0, i.FromPage)
    require.Equal(t, 1, i.ToPage)
})
```

---

## 7. Await / Flush 机制

交互式 E2E 套件的成败，很大程度取决于“等待系统稳定”的机制。

## 7.1 为什么现有测试容易脆

因为很多交互不是同步一跳完成的：

- Input 可能触发多次 `FieldChangeIntent`
- Select / Menu / Modal 可能经历 overlay mount
- TreeView / Table / Pagination 可能先变 action，再变 intent，再 render
- 异步 lazy load / async search 需要等待外部 intent 回流

### 所以必须有统一的等待模型

建议提供：

```go
app.AwaitIdle()
app.AwaitRenderStable()
app.AwaitFocus(locator)
app.AwaitIntent("treeview.SearchResultsIntent")
app.Eventually(func(a *E2EApp) bool { ... })
```

## 7.2 `AwaitIdle()` 的定义建议

默认可以用“可操作定义”而不是“绝对定义”：

当满足以下条件时判定为 idle：

1. 最近一次输入已处理完成
2. 当前没有新的 `Msg` / `Action` / `Intent` 被追加
3. 渲染缓冲连续两次快照一致
4. 若启用了手动时钟，则当前没有待推进任务

### 实现建议

第一阶段可以保守一些：

- 事件注入后循环：
  - 强制 render
  - 读取 trace 长度与 buffer hash
  - 连续 N 次不变即视为稳定

不需要一开始就强依赖复杂调度器内省。

---

## 8. 焦点验证设计

交互组件最常见的 bug 之一，就是**逻辑状态变化了，但焦点错了**。

典型场景：

- Tab 顺序错误
- Shift+Tab 回退错误
- overlay 打开后焦点没进去
- overlay 关闭后焦点没回源
- TreeView / List / Table 当前选中与焦点不一致

### 建议断言模型

#### 8.1 当前焦点断言

```go
app.AssertFocus(e2e.ByComponentID("search-input"))
```

#### 8.2 焦点路径断言

```go
app.Driver().Special(platform.KeyTab)
app.AwaitIdle()
app.AssertFocusTransition(
    e2e.ByID("username"),
    e2e.ByID("password"),
)
```

#### 8.3 Focus Trap 断言

```go
app.Driver().Click(e2e.ByText("Open Modal"))
app.AwaitIdle()
app.AssertFocusWithin(e2e.ByComponentID("settings.modal"))
```

---

## 9. 样式断言设计

文本相同并不代表 UI 语义正确。

例如：

- 焦点态是反色还是边框高亮
- disabled 是否真的变灰
- error 是否真的红色
- selected 是否真的高亮
- current page 是否真的使用 selected style

### 9.1 单 cell 样式断言

```go
app.AssertCellStyle(e2e.At(10, 4), e2e.StyleExpect{
    FG: style.Black,
    BG: style.BrightCyan,
    Bold: true,
})
```

### 9.2 Locator 样式断言

```go
app.AssertStyle(e2e.ByComponentID("orders.pagination"), func(snapshot e2e.NodeSnapshot) {
    require.True(t, snapshot.ContainsStyledText("[3]"))
})
```

### 9.3 样式快照

建议支持将一帧导出成：

- 文本版 snapshot
- cell+style JSON snapshot
- trace+frame 组合 snapshot

这比纯字符串 golden file 更稳定、更有解释性。

---

## 10. 命中与坐标验证

当前很多交互组件都依赖 HitMap / bounds：

- button
- input
- menu
- select popup
- modal / drawer
- treeview
- table
- pagination

所以 E2E 套件需要支持：

```go
app.AssertHit(e2e.At(12, 8), e2e.ByComponentID("submit"))
app.AssertBounds(e2e.ByID("search-input"), expectedRect)
```

### 目标

- 验证点击位置是否命中正确组件
- 验证 overlay / popup 的坐标是否正确
- 验证 portal 后命中框是否更新到最终布局

---

## 11. 建议的测试 DSL

推荐风格应偏“场景脚本”：

```go
app := e2e.Run(t, MyApp,
    e2e.WithViewport(80, 24),
    e2e.WithIntentTrace(),
    e2e.WithStyleSnapshots(),
)
defer app.Close()

app.Step("tab 到搜索框", func(s *e2e.Session) {
    s.Special(platform.KeyTab)
    s.AwaitIdle()
    s.AssertFocus(e2e.ByID("search-input"))
})

app.Step("输入查询并验证结果", func(s *e2e.Session) {
    s.Type("mint")
    s.AwaitIntent("treeview.SearchResultsIntent")
    s.AssertRenderContains("mint")
    s.AssertStyle(e2e.ByComponentID("tree.results"), ...)
})
```

这样能把测试从“裸事件 + sleep + 手工字符串查找”升级成“可读的行为脚本”。

---

## 12. 建议新增的能力分层

## Phase 1 — 在现有 `ui/test.go` 上方加观测层

目标：快速提升可测性，不大动现有 runtime。

新增：

- `E2EApp`
- `AwaitIdle()`
- `IntentTrace`
- `ActionTrace`
- `GetCellSnapshot()`
- `AssertFocus(locator)`
- `ByComponentID` / `ByID` / `ByText`

### 价值

- 成本最低
- 能快速覆盖最痛的交互问题
- 对现有测试体系最友好

## Phase 2 — 增加结构化 Render / Focus / HitMap Probe

新增：

- 完整 `FrameSnapshot`
- `FocusSnapshot`
- `BoundsSnapshot`
- `HitMap` 定位器桥接

### 能覆盖

- 焦点切换
- 样式断言
- portal / overlay 坐标问题

## Phase 3 — 引入确定性时钟与消息 drain

新增：

- 测试时钟
- 手动 tick
- async/transition 队列 drain
- 更可靠的 `AwaitIdle`

### 能覆盖

- 动画
- toast duration
- tooltip delay
- async search / lazy load / debounce

## Phase 4 — 录制回放与失败诊断

新增：

- trace 录制
- 测试失败自动导出
  - render snapshot
  - style snapshot
  - focus snapshot
  - intent trace
  - action trace

### 目标

测试失败后不只告诉你“没找到文本”，而是能告诉你：

- 焦点停在谁身上
- 最后一个 action 是什么
- 最后一个 intent 是什么
- 缓冲区和样式是什么
- 命中图和 bounds 是什么

---

## 13. 建议的目录与代码组织

建议未来实现时采用如下结构：

```text
ui/e2e/
├── app.go              # E2EApp / Session
├── driver.go           # 键鼠输入 DSL
├── await.go            # AwaitIdle / Eventually / AwaitIntent
├── selector.go         # locator / selector
├── assert_text.go
├── assert_focus.go
├── assert_style.go
├── assert_intent.go
├── probe_render.go
├── probe_focus.go
├── probe_message.go
├── trace.go
└── fixtures.go
```

与现有 `ui/test.go` 的关系：

- `ui/test.go` 保留为轻量入口
- `ui/e2e/` 作为强交互测试层
- 两者可以共享 App 启动与事件注入能力

---

## 14. 与现有系统的接缝建议

### 14.1 `framework.App`

继续复用：

- `InjectEvent`
- `GetRenderer`
- `GetFocusManager`

建议新增（未来）：

- `DrainEvents()`
- `IsIdle()`
- `GetLastProcessedMsg()`
- `GetActionLog()`

### 14.2 `render.DeclarativeNode`

继续复用：

- `GetFocusManager()`
- `GetHitMap()`
- `GetFocusedIndex()`
- `GetFocusedType()`
- `GetRenderer()`

建议新增（未来）：

- `GetRootFiber()`
- `SnapshotFocus()`
- `SnapshotHitTargets()`

### 14.3 `Intent Runtime`

建议通过包装全局 `IntentRuntime` 来提供：

- intent trace
- typed intent assertions
- last intent snapshot

### 14.4 `paint.Buffer`

建议从 buffer 提供：

- 结构化 cell 样式读取
- 行/列切片
- locator 到 cell 范围映射

---

## 15. 推荐的最小首批覆盖组件

第一批建议优先覆盖：

1. `Input`
2. `Textarea`
3. `Button`
4. `Select`
5. `Menu`
6. `Tabs`
7. `TreeView`
8. `Table`
9. `Modal`
10. `Pagination`

原因：

- 它们最依赖焦点
- 最依赖消息处理
- 最依赖样式语义
- 最容易出现“看起来能跑，实际上交互不稳定”的问题

---

## 16. 推荐断言能力清单

### 文本类

- `AssertRenderContains(text)`
- `AssertRenderNotContains(text)`

### 焦点类

- `AssertFocus(locator)`
- `AssertNoFocus(locator)`
- `AssertFocusWithin(locator)`

### 样式类

- `AssertCellStyle(...)`
- `AssertStyle(locator, ...)`

### 交互类

- `AssertIntent(intentType)`
- `AssertLastIntent(...)`
- `AssertAction(actionType)`
- `AssertMessage(msgType)`

### 布局/命中类

- `AssertBounds(locator, rect)`
- `AssertHit(point, locator)`
- `AssertVisible(locator)`

### 稳定性类

- `AwaitIdle()`
- `AwaitRenderStable()`
- `Eventually(...)`

---

## 17. 失败输出建议

测试失败时，建议自动输出：

1. 当前 render 文本
2. 当前焦点快照
3. 最近 20 条 trace event
4. 命中图摘要
5. 可选的样式快照

示例：

```text
E2E ASSERTION FAILED: expected focus on componentID=search-input

Current focus:
- componentID: submit-button
- tag: button
- bounds: (12,8)-(24,9)

Recent trace:
1. raw_input: Tab
2. msg: KeyMsg(Tab)
3. action: navigate_next
4. focus: submit-button

Render:
...
```

这样的错误信息才足够支持排查复杂交互问题。

---

## 18. 结论

当前 `ui/test.go` 已经是一个不错的基础层，但它更像：

- “完整 App 驱动的测试启动器”
- 而不是“交互型 E2E 测试框架”

要真正覆盖复杂交互组件，需要在其之上补齐：

- **结构化消息观测**
- **焦点探针**
- **样式快照**
- **selector/locator**
- **确定性 await/idle**
- **trace 与失败诊断**

推荐路线不是推倒重来，而是：

1. 站在 `ui/test.go`、`framework.App`、`DeclarativeNode`、`IntentRuntime` 现有 seam 上扩展
2. 先做最小的 Probe + Await + Assert
3. 再逐步升级成真正的交互式 E2E 套件

这会是对 Mint 交互性组件测试能力提升最大的基础设施之一。
