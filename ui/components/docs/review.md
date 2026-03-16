# Mint UI Components Review

> 更新时间：2026-03-16
> 范围：`ui/components/`（排除 `docs/`、`internal/`）

---

## 总体判断

`ui/components/` 已经不是“骨架期”，而是“已大面积落地、少量边角未收口”的状态。

- 当前共有 39 个组件/支撑目录
- 其中 35 个严格遵循 `builder.go + vnode.go + instance.go` 规范
- `toast`、`statusbar`、`control`、`validation` 属于特例或基础设施模块
- `go test ./ui/components/...` 当前可通过；无测试文件的包已收敛到 `internal/` 支撑模块

---

## 成熟度概览

### 第一梯队

- `select/`：overlay popup、多选、tags、`filterOption`、portal、outside-click 收口已经齐备
- `menu/`：menubar、dropdown、context menu、popup、submenu、typeahead、shortcut 安装和关闭中间件完整
- `treeview/`：懒加载、搜索、多选、受控展开、可见节点缓存已经成型
- `table/`：分页、搜索、列过滤、选择和滚动条能力可用
- `form/`：`FormItem`、字段意图、validator 联动已打通
- `drawer/`：组件、运行时和测试均已存在

### 第二梯队

- `button/`、`input/`、`checkbox/`、`modal/`、`tabs/`、`list/`、`notification/`、`toast/`、`tooltip/`、`grid/`、`panel/`、`statusbar/` 已具备稳定可用的 API 与测试基础
- `tooltip/` 的实际能力已经超过旧路线图，当前支持 top/bottom/left/right、delay 和 layer
- `statusbar/` 采用“builder + section/help 内部组件”的组合式路线，不是传统单组件目录

### 相对薄弱

- `progress/`：仍以线性进度条为主，尚未扩展到 Circle / Dashboard
- `empty/`：功能与测试已补齐，但能力面仍较窄
- `control/`、`validation/`：更接近基础设施包，不适合用单一 UI 组件标准衡量

---

## 当前主要问题

### 1. 文档漂移高于代码缺口

- `ROADMAP.md` 之前低估了目录数，也把 `Drawer` 误记为未完成
- `COMPONENT_MIGRATION_GUIDE.md` 仍残留 `stack` / `border` 目录引用与过时示例状态
- `grid/README.md` 曾将追踪与调试能力误记为“未实现”

### 2. 表单仍保留兼容层 registry

- `FormItem` 的运行时联动现在已优先沿实例树解析祖先 `Form`
- `GetFormContext()` 和跨树兼容路径仍保留 registry 回退
- 剩余工作主要是字段级 `touched` / `dirty` 状态，以及进一步收敛兼容层

### 3. 文档卫生需要持续整理

- 旧版评审文档曾混入对话式 AI 草稿，现已清理
- 大体量设计文档仍有历史信息与当前实现交织的情况，需要按“现状 / 历史”分层

---

## 正向信号

- `list/` 与 `table/` 已开始共享选择模型，说明内部抽象在收敛
- 覆盖层能力在 `select/`、`menu/`、`modal/`、`drawer/` 之间已经形成相对一致的运行时思路
- `form/` 的 `FormItem + validator source` 机制已经从“能跑”进入“可维护”阶段

---

## 建议的后续顺序

1. 继续清理组件级文档漂移，优先 `ROADMAP.md`、迁移指南、复杂组件 README。
2. 在 `form/` 上补字段级状态，再决定是否完全移除兼容 registry。
3. 继续把测试空白从公共组件收敛到内部支撑模块，优先补 `internal/` 下仍无测试的工具包。
