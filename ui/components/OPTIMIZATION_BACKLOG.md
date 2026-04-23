# Mint UI Components Optimization Backlog

> 范围：仅针对已存在组件的增强与收口，不包含全新组件开发
>
> 基线文档：[ROADMAP.md](./ROADMAP.md)
>
> 更新日期：2026-03-28
>
> 状态说明：`ROADMAP.md` 中的 Phase 1-4 与本文件最初记录的组件功能 epic 已全部完成；本文现改为记录“后路线图阶段”的真实剩余优化项。

---

## 当前判断

`ui/components/` 当前已经不再处于“大批量补功能”的阶段，而是进入“稳定性、复用收口、文档与测试密度提升”为主的阶段。

- 原 backlog 中的 9 个功能型主题已经完成：
  - `Form`
  - `Input`
  - `Progress`
  - `Modal`
  - `Tooltip`
  - `List`
  - `Tabs`
  - `Table`
  - `TreeView`
- `go test ./ui/components/...` 当前可通过
- dedicated `ui/e2e` 覆盖已经补到大多数公共组件，剩余工作重点不再是“有没有组件”，而是“语义是否稳定、文档是否同步、基础设施是否继续收口”

---

## 当前活跃 Backlog

### 1. Form 稳定性与状态语义回归

#### 目标

继续把 `Form` 从“功能齐备”推进到“状态语义长期稳定”，重点盯住字段元状态、实例树解析路径和 compat API 的边界。

#### 主要入口

- `ui/components/form/instance.go`
- `ui/components/form/item.go`
- `ui/components/form/form_context.go`
- `ui/components/form/compat_registry.go`

#### 待办

- 补更系统的 `blur/change/validate/submit/reset` 组合回归矩阵
- 持续验证 `SetProps(values)`、`Reset()`、重新挂载后的字段元状态是否一致
- 持续验证 owner-bound 与 ownerless 两类路径下，`GetFormContext(...)` 和 compat registry 的边界不回退
- 为显式兼容 API 增加更清晰的文档注释和使用示例，避免后续误用成普通运行时入口

#### 验收标准

- 字段级 `touched/dirty/submitted` 在典型交互和重渲染路径下保持一致
- compat API 仍可用，但不会重新侵入普通运行时路径
- FormItem 的错误展示时机在 blur、submit 和 reset 后不回归

#### 建议测试

- `FormItem` untouched -> blur -> submit -> reset 状态回归
- `SetProps(values)` 后 touched/dirty/submitted 清理回归
- 同一表单树内解析与跨树 compat 访问的边界回归

---

### 2. Overlay 定位能力继续收口

#### 目标

把已经下沉到 `overlayposition` 的定位能力继续做实，使浮层组件共享同一套 viewport / candidate / clamp 逻辑，而不是各自复制一套近似实现。

#### 主要入口

- `ui/components/internal/overlayposition/`
- `ui/components/tooltip/`
- `ui/components/popover/`
- `ui/components/popconfirm/`
- `ui/components/statusbar/`
- `ui/components/menu/`

#### 当前状态

- `popover` / `popconfirm` 已把 vertical auto placement、candidate fallback 与 viewport clamp 主路径进一步下沉到 `overlayposition`
- `tooltip` 的 top/bottom 与 left/right 显式 placement candidate 展开已复用 `overlayposition` 共享 vertical / horizontal helper
- `statusbar` 已复用共享候选位置 helper，但 auto placement 偏好仍保留本地策略，以兼容既有 hover 布局语义
- `statusbar` 的显式 `TooltipPlacementTop` / `TooltipPlacementBottom` 现在会在保留 edges-first 候选顺序的前提下，处理顶角 / 底角场景下对侧 vertical family 的回退；overlay help 运行时也已接入真实 viewport 注入，避免 tooltip-layer 下因零尺寸 bounds 丢失顶角 / 底角 fallback；对应双轴受限几何、极窄 viewport 下保持原 family 的 left-edge clamp，以及“同侧/对侧候选都放不下时的双轴 clamp”都已补到组件单测，arrow-capable 的 dual-axis clamp 也已有 dedicated e2e
- `runtime/ui` 已暴露异步 render request 入口，`tooltip` 的 delayed show 现在会在 timer 回调后主动请求重绘
- `tooltip` 单测与 e2e 已补 `delay` 生命周期、child-hover 路径、延迟显示，以及右边界 `right` fallback 回归
- `tooltip` 的 `right-top` / `left-bottom` 角落 placement 现在会优先回退到镜像 horizontal family，再退回上下方向；对应 corner clamp、`right-top` e2e，以及“horizontal family 都放不下时回退到 `top` / `bottom`”的窄视口单测都已补；`internal/overlayposition` 的共享 horizontal helper / resolver 现在也有独立单测覆盖 `RightTop` / `LeftBottom` 以及对称的 `RightBottom` / `LeftTop` 回退链；`bottom-left` / `bottom-right` 的底角回退单测、显式 `top-right` / `bottom-right` 的 left-edge clamp 保持原 vertical family，以及“没有任何 vertical candidate 能完整放入时”的双轴 clamp 现在也都已由组件单测覆盖
- `menu` 根 popup 已接通 viewport-aware anchored placement fallback / clamp，定位尺寸改为真实外框（`surface + shadow`）；`context menu` 已接通基于真实外框尺寸的 viewport clamp；submenu 级联面板也已支持右侧放不下时自动翻转到左侧、纵向 viewport clamp、沿父级方向连续展开，并在极窄 viewport 下按最终 clamp 侧继续推导后续方向，同时覆盖整棵 cascade 的 hit bounds；这套 cascade 规则已开始下沉到 `internal/overlayposition`，共享 helper 也已补 preferred-left、top-edge、double-axis clamp，以及“left-edge clamp 后下一层镜像回右侧”与“right-edge clamp 后下一层镜像回左侧”的方向传递矩阵；`menu` e2e 也已覆盖 bottom-right upward clamp、窄底角 left-edge clamp + upward clamp，以及左右两侧的镜像回退链路，而且两条镜像链路都已进一步补了靠近底边时伴随 upward clamp 的组合场景，并新增了更极端的底边 zig-zag 级联链路
- `popover` / `popconfirm` / `tooltip` 在极窄 viewport 下的显式 `top-left` / `top-right` / `bottom-left` / `bottom-right` clamp 语义，目前都已由组件单测对称覆盖，确保在 left-edge clamp 时仍保持原 vertical family；`internal/overlayposition` 的共享 vertical helper / resolver 也已补对称的 `TopLeft` / `TopRight` / `BottomLeft` / `BottomRight` candidate 顺序、顶角/底角 family fallback、窄视口 clamp，以及“没有任何 vertical candidate 能完整放入时的双轴 clamp”单测；其中 `popover` 已补 arrow-based dedicated e2e，用于验证 dual-axis clamp 后仍保留 `top` / `bottom` family，`tooltip` / `popconfirm` 继续以组件单测为主，避免把终端宽度噪声引入主套件
- 跨组件一致性回归已覆盖 `tooltip` / `popover` / `popconfirm` 的 `auto` 与显式 `top` 在顶边场景下的 fallback 行为；同时也已覆盖 `tooltip` / `popover` / `popconfirm` / `statusbar` 在左右边界且仍有可用垂直空间时，显式 `top` / `bottom` 会保持各自 family 并分别向左/向右选择横向候选；此外显式 `top-left` / `top-right` 在顶角场景下回退到下方同侧 family、`bottom-left` / `bottom-right` 在底角场景下回退到上方同侧 family 的 corner 语义，现在也已补到 `tooltip` / `popover` / `popconfirm` / `statusbar` 的 e2e，其中 `statusbar` 同时保留了 dedicated corner e2e，`go test ./ui/e2e/...` 当前通过

#### 待办

- 如果 runtime 后续暴露更完整的 portal / anchor 布局信息，继续把这些信息下沉到共享 helper；当前 `statusbar` overlay help 已先接入真实 viewport 注入
- 继续收口 `menu` 的 submenu 级联路径，重点转为是否把这套共享 cascade helper 推广到更多 overlay / menu 变体，以及继续补更多组合角落回归，尤其是更多多层 cascade 的顶部/底部/左右同时受限矩阵
- 继续扩 `Tooltip` 的 placement / fallback / clamp 边界回归，重点把更多角落组合和 clamp 极限场景补到跨组件 e2e，而不只停留在 `tooltip` 单测
- 评估是否把 `statusbar` 的 auto placement 偏好也纳入共享策略，前提是不破坏既有 hover 布局语义
- 继续补跨组件一致性回归：`tooltip`、`popover`、`popconfirm`、`statusbar` 在更多边界条件下行为一致，重点转向更极端的多轴 clamp / 窄视口组合，而不再停留在基础 corner family

#### 验收标准

- 新旧 overlay 组件在边界场景下不再出现各自不同的 fallback 行为
- 显式 placement、auto placement、outside-click、ESC 收口语义保持一致
- `delay`、viewport clamp 和候选位置回退在 e2e 与单测层都有稳定覆盖

#### 建议测试

- `overlayposition` 共享 vertical / horizontal / cascade helper 的单测，覆盖 auto bias、candidate 顺序、family fallback、direction 推导与 clamp
- 顶部、底部、左右边缘 anchor 的 placement/fallback 回归
- auto placement 与显式 placement 的 clamp 回归
- hover/click/manual 三类触发方式的关闭路径回归

---

### 3. 文档同步与 README 收口

#### 目标

继续把“代码已经实现，但文档仍停留在旧阶段”的问题清掉，避免路线图、迁移指南、README 和实际能力再次分叉。

#### 主要入口

- `ui/components/ROADMAP.md`
- `ui/components/COMPONENT_MIGRATION_GUIDE.md`
- `ui/components/docs/review.md`
- 各复杂组件目录下 README

#### 待办

- 继续拆分 `COMPONENT_MIGRATION_GUIDE.md` 中“大段历史迁移模板”和“当前组件入口”，避免读者把迁移期说明误解为现状文档
- 继续整理 `ui/components/docs/` 下历史分析文档的“现状 / 历史”分层，避免旧结论继续被误读成当前实现
- 继续为其余复杂或支撑型组件 README 补“状态语义 / 安装方式 / 测试入口”，避免信息只停留在测试文件里

#### 验收标准

- 路线图、backlog、迁移指南、README 之间不再互相矛盾
- 复杂组件 README 能解释关键状态语义、安装方式和常见测试入口

---

### 4. 测试密度继续向高风险交互倾斜

#### 目标

公共组件的测试覆盖已经大幅补齐，支撑模块里最明显的“无测试空白”也已基本收口；下一阶段要把测试密度继续倾斜到最容易因重渲染、action 路径或跨组件联动而退化的高风险交互。

#### 主要入口

- `ui/components/internal/proputil/`
- `ui/components/internal/selection/`
- `ui/components/internal/scroll/`
- `ui/components/internal/overlayposition/`
- 各复杂组件的跨组件集成测试

#### 当前状态

- `internal/proputil` 已补最小单测，覆盖 `GetString`、`GetBool`、`GetInt`、`GetStyle`、`GetIntent`
- `internal/selection` 已补最小单测，锁定 `SelectionMode` 的稳定值与互异性
- `internal/scroll` 已有基础单测，覆盖 action delta、viewport normalize / visible range / page up-down 与 scrollbar 绘制
- `internal/overlayposition` 已有 shared helper 矩阵单测，覆盖 vertical / horizontal placement family 与 cascade direction / clamp
- 当前剩余测试空白的重点，已经从“支撑模块有没有测试”，转成“高耦合交互链路是否有 dedicated regression coverage”

#### 待办

- 为高耦合交互路径补跨组件回归，而不是继续只堆单组件 happy-path e2e
- 优先补“最容易因重渲染或 action 路径变化而退化”的场景：受控/非受控切换、虚拟滚动、overlay 关闭链路、表单状态同步
- 继续评估哪些跨组件链路值得从大而宽的全量 e2e，收口成更聚焦、更稳定的 regression fixture

#### 验收标准

- 支撑模块的测试空白主要收敛到纯薄封装或低复杂度 helper
- 高风险交互路径至少有一层 dedicated regression coverage

---

## 可选的低优先级增强

这些不属于当前阻塞项，但如果后续有明确产品需求，可以作为下一批增强候选。

- `popover` / `popconfirm`：
  - 更丰富的内容模板
  - 更完整的 left/right placement 家族
- `collapse` / `descriptions` / `statistic` / `timeline` / `result` / `empty`：
  - 更丰富的展示模板和布局变体
- `absolute` / `wrap` / `cursor` / `control` / `validation`：
  - 进一步明确其“公共组件”还是“基础设施模块”的定位

---

## 已完成批次（历史记录）

以下原 backlog 主题均已完成，不再作为活跃待办继续推进：

- `Form`
  - compat registry 已显式收口
  - `GetTouchedFields` / `GetDirtyFields` / `GetSubmittedFields` / `HasSubmitted` / `GetSubmitCount` 已补齐
- `Input`
  - `InputNumber` 语义、`AllowNegative`、`AllowDecimal`、`Min`、`Max`、`Step` 已完成
- `Progress`
  - `line` / `circle` / `dashboard` 与 `status` / `active` tick 动画已完成
- `Modal`
  - `Confirm/Info/Success/Error/Warning` 静态 helper 与 footer 变体已完成
- `Tooltip`
  - 12 方位 placement、fallback、clamp、delay、共享 overlay 定位已完成
- `List`
  - `RowItem` / `List.Item`、`VirtualList` bridge 与 state sync 已完成
- `Tabs`
  - `card`、`closable`、drag reorder 已完成
- `Table`
  - expandable rows、tree data、fixed columns、width strategy 已完成
- `TreeView`
  - 异步搜索结果分页、高亮与同父级 subtree 拖拽排序已完成

---

## 维护约定

后续如果某项已经完成：

- 直接从“当前活跃 Backlog”移除
- 必要时补到“已完成批次”
- 不再把已完成 epic 保留在“建议顺序”或“Sprint 计划”里

这样可以避免 backlog 再次退化成“历史完成记录和真实待办混写”的状态。
