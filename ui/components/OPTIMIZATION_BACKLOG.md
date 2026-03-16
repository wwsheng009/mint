# Mint UI Components Optimization Backlog

> 范围：仅针对已存在组件的增强与收口，不包含全新组件开发
>
> 基线文档：[ROADMAP.md](./ROADMAP.md)
>
> 生成日期：2026-03-17

---

## 目标

这份文档把 `ui/components/ROADMAP.md` 中仍待增强的现有组件，整理成可执行的优化 backlog。

文档侧重 4 件事：

- 明确每个组件当前还缺什么
- 给出建议的实施顺序
- 标出主要入口文件，降低启动成本
- 给出验收标准和测试方向，避免“功能做了但语义没收口”

---

## 当前建议顺序

### 第一轮

先做收益高、影响广、实现风险相对可控的增强项。

1. `Form`
2. `Input`
3. `Progress`
4. `Modal`
5. `Tooltip`

### 第二轮

在第一轮收口后，处理中等规模的结构增强。

### 第三轮

最后处理状态面和布局面最重的组件。

7. `TreeView`

---

## 第一轮 Backlog

### 1. Form

#### 当前缺口

- 字段级 `touched/dirty` 已完成（2026-03-17）
- `FormItem` 已按 `touched` 或提交后校验状态控制错误展示
- `GetFormContext` 仍依赖 registry 兼容层
- 下一步是继续收敛跨树兼容路径，减少新逻辑对 registry 的依赖

#### 主要入口

- `ui/components/form/instance.go`
- `ui/components/form/item.go`
- `ui/components/form/form_context.go`
- `ui/components/form/intent.go`
- `ui/components/form/vnode.go`

#### 建议任务

1. 收敛 registry 兼容层
   - 保留 `GetFormContext` 兼容 API
   - 但让 `FormItem` 与普通运行时路径优先走祖先实例解析
   - 减少新逻辑继续依赖 registry

2. 继续补字段状态辅助 API
   - 如有需要补 `GetTouchedFields` / `GetDirtyFields`
   - 评估是否需要更明确的字段级提交状态

#### 风险点

- `blur`、`change`、`validate`、`submit`、`reset` 的状态语义仍需持续保持一致
- registry 兼容层缩减时，要避免跨树访问场景回归

#### 验收标准

- 已有字段级 `touched/dirty` 行为保持稳定
- `Reset` 与 `SetProps(values)` 后字段元状态保持一致
- 兼容 API 仍可用，但核心运行时进一步减少对 registry 的依赖

#### 建议测试

- 字段元状态回归
- `SetProps(values)` 后字段级状态归零
- `FormItem` 在 untouched 时不显示错误，在 blur 或提交后显示错误

---

### 2. Input

#### 已完成（2026-03-17）

- `TypeNumber` 现在具备完整 `InputNumber` 运行时语义，非法字符不会进入最终值
- 新增 `AllowNegative(bool)` 和 `AllowDecimal(bool)`，用于控制负号和小数点模式
- blur 会归一化前导零、`.5` / `-.5`、`1.`、`-0`、悬空 `-` / `-.` 等临界输入
- blur 归一化如果改写了值，会先补发一次 change，再发 blur，确保 Form 与非 Form 路径最终值一致

#### 后续

- 第二阶段再考虑 `Min`、`Max`、`Step`

---

### 3. Progress

#### 已完成（2026-03-17）

- 新增 `status`：`normal`、`success`、`exception`、`active`
- 新增 `type`：`line`、`circle`、`dashboard`
- `circle` 和 `dashboard` 已有稳定的 ASCII/分段式 TUI 表达
- 三种模式的 `Measure` 与 `label/showPercent` 布局已统一收口

#### 后续

- 如果要继续增强，可单独追加 `active` 的逐帧动画而不影响当前 API

---

### 4. Modal

#### 已完成（2026-03-17）

- 已补 `Confirm(...)`、`Info(...)`、`Success(...)`、`Error(...)`、`Warning(...)` 静态 helper，并保留 `Alert(...)`
- helper 默认会生成可用 modal：标题、内容、默认 footer、默认关闭策略都已收口
- helper 返回 `*Builder`，可以继续叠加普通 builder 配置
- helper footer 按钮通过 modal 包内统一 close intent handler 复用 `requestClose()`，不会绕开现有 ESC / backdrop 关闭语义
- 已补 `ui` shortcut：`ModalInfo/Success/Warning/Error`

#### 后续

- 如果需要，可继续补更细的 helper 选项模板，例如 icon/prefix、按钮文案国际化、按钮布局预设

---

### 5. Tooltip

#### 已完成（2026-03-17）

- `delay` 维持现有能力
- `placement` 已扩到完整 12 方位：`top/topLeft/topRight`、`bottom/bottomLeft/bottomRight`、`left/leftTop/leftBottom`、`right/rightTop/rightBottom`
- `auto` 与显式 placement 都已接入统一的候选回退策略
- 当所有候选位置都无法完整放入视口时，会自动 clamp 到可见区域

#### 后续

- 如果后面接入真实视口/portal 布局信息，可把当前 `viewport` 回退逻辑进一步下沉到统一 overlay 定位层
- `delay` 回归验证，避免已有能力退化

---

## 第二轮 Backlog

### 6. List

#### 当前缺口

- `List.Item` / item 模型已完成（2026-03-17）
- `VirtualList` 渲染桥接已完成（2026-03-17）
- 选择、搜索、高亮与 `VirtualList` 的进一步同步已完成（2026-03-17）

#### 主要入口

- `ui/components/list/vnode.go`
- `ui/components/list/builder.go`
- `ui/components/list/instance.go`
- `ui/components/virtuallist/`

#### 建议任务

1. 引入 item 模型
   - 支持 `title`
   - 支持 `description`
   - 支持 `prefix/suffix`
   - 保留字符串列表兼容层

2. ~~补 `List.Item` builder 或辅助结构~~
   - ✅ 已完成（2026-03-17）

3. 为大数据量场景增加与 `VirtualList` 的桥接层
   - ~~先做渲染桥接~~
   - ✅ 已完成（2026-03-17）
   - ~~再做选择、搜索、高亮同步~~
   - ✅ 已完成（2026-03-17）

#### 风险点

- 当前 `[]string` API 很简单，升级 item 模型时要保证兼容
- `VirtualList` 集成可能会影响选择和搜索语义

#### 验收标准

- 旧 API 继续可用
- 新 item 模型可表达 richer row
- 大列表能切到虚拟渲染路径
- 当前已具备正式 bridge API 与双向 state sync，List 增强项已全部完成

#### 建议测试

- 旧 `[]string` 列表回归
- item 模型绘制
- 搜索与选择在虚拟渲染模式下保持一致

#### 本轮完成

- 新增 `RowItem` 结构与 `Item(...)` helper，支持 `title`、`description`、`prefix`、`suffix`
- `Builder` / `VNode` / `Instance` 已支持 `Items(...)` / `AddItem(...)`，并统一降级成现有 `rows []string` 语义
- 旧 `rows []string` API 保持不变，selection/search/paint 继续基于扁平化 row text 工作
- 新增 `ToVirtualList()` / `BuildVirtualList()` bridge，能把当前 List 的 rows、search filter、scrollOffset、selectedIndex 快照成 `VirtualList`
- 新增 `ToVirtualBridge()` / `BuildVirtualBridge()` / `SyncToList(...)`，把 `VirtualList` 的 source index、selection、scrollOffset 回写到 `List`
- `VirtualList` 新增 `ItemStyleFn`，bridge 现在会把 `rowStyleFn` / `matchStyle` 一并透传到虚拟渲染路径

---

### 7. Tabs

#### 当前缺口

- `card` 样式已完成（2026-03-17）
- `closable` 已完成（2026-03-17）
- 拖拽排序已完成（2026-03-17）

#### 主要入口

- `ui/components/tabs/vnode.go`
- `ui/components/tabs/builder.go`
- `ui/components/tabs/instance.go`
- `ui/components/tabs/intent.go`

#### 建议任务

1. ~~先加 `card` 样式~~
   - ~~这是纯视觉增强，最容易独立交付~~
   - ✅ 已完成（2026-03-17）

2. ~~再加 `closable`~~
   - ~~TabItem 增加 `closable`~~
   - ~~关闭后 active index 自动修正~~
   - ~~支持关闭事件 intent~~
   - ✅ 已完成（2026-03-17）

3. ~~最后做拖拽排序~~
   - ~~如果当前交互层不支持鼠标拖拽闭环，可先保留设计，不急着落地~~
   - ✅ 已完成（2026-03-17）

#### 风险点

- 关闭当前标签页后的 active tab 选择策略必须一致
- 受控模式与非受控模式下关闭逻辑要统一
- 拖拽排序要避免和点击选中、close hitbox 语义互相污染

#### 验收标准

- `card` 样式不影响现有 intent
- closable tab 可关闭且 active 状态正确
- disabled tab 与 closable tab 共存不异常
- 拖拽排序后 tab order、active tab 与 close hitbox 语义仍正确
- Tabs 增强项已全部完成

#### 建议测试

- 关闭当前 tab
- 关闭最后一个 tab
- 关闭非激活 tab
- controlled 模式下 change/close 行为
- press/move/release 拖拽排序
- ActionClick/ActionHover/ActionMouseRelease 路径下的拖拽排序

#### 本轮完成

- 新增 `TabVariant`，支持 `line` / `card`
- `TabItem` 新增 `Closable` 语义，Builder/VNode 支持 close intent 配置
- 关闭命中已从 tab 选择区拆分出独立 close hitbox，`card` 与普通样式都可关闭
- 关闭后 active tab 会按“优先右侧，否则左侧”的策略自动修正；`SetProps` 在 tabs 变化时也会尽量按 tab ID 保持语义一致
- 新增 `TabCloseIntent`，支持通过 `componentID` 冒泡/全局处理关闭事件
- `Builder` / `VNode` / `Instance` 已接入 `card` 样式渲染，且不影响现有 intent、active tab、click bounds 语义
- 新增 `Reorderable` / `OnReorder` / `TabReorderIntent`，Tabs 支持按 press/move/release 拖拽排序
- reorderable 模式下点击与拖拽已拆分：按下进入 drag candidate，未拖动则在 release 时选中，拖动则保持原 active tab 并完成重排

---

## 第三轮 Backlog

### 8. Table

#### 当前缺口

- 固定列已完成（2026-03-17）
- 树形数据已完成（2026-03-17）
- `expandable` 已完成（2026-03-17）
- 列宽百分比 / 自适应已完成（2026-03-17）

#### 主要入口

- `ui/components/table/vnode.go`
- `ui/components/table/builder.go`
- `ui/components/table/instance.go`
- `ui/components/table/intent.go`
- `ui/components/table/selection.go`

#### 建议拆分

1. ~~Epic A: expandable rows~~
   - ~~行展开 intent~~
   - ~~行展开状态~~
   - ~~子内容渲染~~
   - ✅ 已完成（2026-03-17）

2. ~~Epic B: tree data~~
   - ~~行层级结构~~
   - ~~展开折叠~~
   - ~~与排序、过滤、分页的交互规则~~
   - ✅ 已完成（2026-03-17）

3. ~~Epic C: fixed columns~~
   - ~~左右固定列~~
   - ~~横向滚动同步~~
   - ~~视口裁剪与边框绘制~~
   - ✅ 已完成（2026-03-17）

4. ~~Epic D: column width strategy~~
   - ~~百分比宽度~~
   - ~~自适应宽度~~
   - ~~最小宽度 / 最大宽度~~
   - ✅ 已完成（2026-03-17）

#### 风险点

- 分页、排序、过滤、选择、滚动已经存在，任何新能力都容易耦合
- 固定列会直接碰布局与绘制边界
- 树形数据和 expandable 不能做成两套独立状态机

#### 验收标准

- 每个 epic 都能独立上线，不要求一次全做完
- 排序、过滤、分页、选择的现有行为不回归

#### 本轮完成

- `TableColumn` 新增 `WidthPercent`、`MinWidth`、`MaxWidth`
- 列宽计算已支持固定宽度、百分比宽度、内容自适应宽度混用
- 在受限宽度下会先按百分比计算，再把剩余空间分配给 auto 列；超出时按 `MinWidth` 下限收缩
- 新增 expandable rows：按 source row 管理展开状态，支持声明式 expanded content、点击/Enter 切换、附加内容行渲染
- 展开状态已接入现有分页/排序/过滤视图，并可通过 `expandIntent` / `expandIntentField` 发出变更
- 新增 fixed columns：支持左右固定列、中心区域水平视口裁剪，以及 `ActionScrollLeft/Right` / `ActionNavigateLeft/Right` 下的横向滚动
- 新增 tree data：通过 `treeParents` 映射描述层级结构，复用 `expandedIndices` 管理树节点展开状态，并在首列进行层级缩进显示

#### 建议测试

- 展开行与分页组合
- 树形数据与选择组合
- 固定列与横向滚动组合
- 百分比列宽与窄屏裁剪

---

### 9. TreeView

#### 当前缺口

- 缺拖拽排序
- 缺异步搜索高亮分页

#### 主要入口

- `ui/components/treeview/vnode.go`
- `ui/components/treeview/builder.go`
- `ui/components/treeview/instance.go`
- `ui/components/treeview/instance_helpers.go`
- `ui/components/treeview/instance_keyboard.go`
- `ui/components/treeview/instance_paint.go`
- `ui/components/treeview/intent.go`

#### 建议拆分

1. Epic A: 异步搜索高亮分页
   - 搜索结果分页
   - 当前匹配项定位
   - 大树场景下高亮与滚动联动

2. Epic B: 拖拽排序
   - 节点拖拽 intent
   - 拖拽预览
   - 重排后的 path / key / selection / expanded state 修正

#### 风险点

- `TreeView` 当前是最大组件之一，状态面已经很宽
- lazy-load、selection、expanded、search 互相之间容易出现状态不同步
- 拖拽一旦引入 path 变更，很多缓存都要重算

#### 验收标准

- 异步搜索不会破坏现有搜索和滚动逻辑
- 拖拽后 expanded / selected / focused 状态仍正确
- 懒加载节点的拖拽边界清晰

#### 建议测试

- 搜索分页切换
- 搜索结果高亮与选中联动
- 拖拽后 path 更新
- 拖拽后 expanded keys 与 selected keys 回归

---

## 建议交付方式

### Sprint A

- `Form`
- `Input`

### Sprint B

- `Progress`
- `Modal`
- `Tooltip`

### Sprint C

- `List`
- `Tabs`

### Sprint D+

- `Table`
- `TreeView`

---

## 文档修正建议

在执行 backlog 之前，建议顺手修正两处文档漂移：

1. `Tooltip` 的 `delay` 已经实现，不应再作为待完成功能
2. `ROADMAP.md` 应增加到本文件的链接，避免后续 backlog 与 roadmap 再次分叉

---

## 附注

当前 `go test ./ui/components/...` 可通过，因此本文档描述的事项属于增强型工作，不是阻塞型修复。
