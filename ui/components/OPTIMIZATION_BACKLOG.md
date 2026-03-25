# Mint UI Components Optimization Backlog

> 范围：仅针对已存在组件的增强与收口，不包含全新组件开发
>
> 基线文档：[ROADMAP.md](./ROADMAP.md)
>
> 更新日期：2026-03-25
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

#### 待办

- 如果 runtime 后续暴露更完整的真实 viewport / portal 布局信息，继续把这部分信息下沉到共享 helper
- 继续排查是否还有浮层组件未完全复用 `overlayposition`
- 为 `Tooltip` 的 `delay`、`placement`、fallback、clamp 增加更系统的回归覆盖
- 补跨组件一致性回归：`tooltip`、`popover`、`popconfirm`、`statusbar` 在相同边界条件下行为一致

#### 验收标准

- 新旧 overlay 组件在边界场景下不再出现各自不同的 fallback 行为
- 显式 placement、auto placement、outside-click、ESC 收口语义保持一致
- `delay`、viewport clamp 和候选位置回退在 e2e 与单测层都有稳定覆盖

#### 建议测试

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

- 更新 `COMPONENT_MIGRATION_GUIDE.md` 中仍残留的旧目录、旧示例和过时状态
- 优先补齐复杂组件 README：`select`、`menu`、`table`、`treeview`、`tooltip`、`popover`、`popconfirm`、`drawer`
- 对“已实现但容易误解”的 controlled/uncontrolled 语义补更明确的说明
- 对新增 dedicated e2e 覆盖批次补文档入口，降低后续排查成本

#### 验收标准

- 路线图、backlog、迁移指南、README 之间不再互相矛盾
- 复杂组件 README 能解释关键状态语义、安装方式和常见测试入口

---

### 4. 测试空白收敛到支撑模块

#### 目标

公共组件的测试覆盖已经大幅补齐，下一阶段要把“无测试空白”继续压缩到真正低风险的内部工具。

#### 主要入口

- `ui/components/internal/proputil/`
- `ui/components/internal/selection/`
- 各复杂组件的跨组件集成测试

#### 待办

- 评估 `internal/proputil`、`internal/selection` 是否值得补最低限度单测
- 为高耦合交互路径补跨组件回归，而不是继续只堆单组件 happy-path e2e
- 优先补“最容易因重渲染或 action 路径变化而退化”的场景：受控/非受控切换、虚拟滚动、overlay 关闭链路、表单状态同步

#### 验收标准

- 测试空白主要收敛到纯薄封装或低复杂度 helper
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
