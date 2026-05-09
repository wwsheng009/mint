# 组件开发指南

本文档面向当前 Mint 源码，说明如何开发或维护 `ui/components` 下的 Fiber-first 组件。

## 当前组件模型

Mint 组件当前分为几层：

```text
Builder / shortcut API
  -> VNode declaration
  -> Fiber
  -> ComponentInstance
  -> layout
  -> paint
  -> Action / Intent interaction
```

核心原则：

- VNode 是声明输入，不保存可变运行时状态。
- Fiber 保存结构、identity、props/style/layer/layout inputs 和 `Instance`。
- 组件运行时状态放在 `ComponentInstance`。
- 绘制能力通过 `PaintableInstance` 或 PaintRegistry 提供。
- 交互优先走 `runtime/action` 和 `runtime/intent`。
- 动态列表必须提供稳定 key。

## 推荐文件结构

典型组件目录：

```text
ui/components/<name>/
  README.md
  builder.go
  vnode.go
  instance.go
  intent.go        # 有语义 intent 时
  *_test.go
```

不是所有组件都需要全部文件。基础设施组件、manager 型组件或图表内部包可以有自己的结构。

## 核心接口

关键源码：

- `../../runtime/ui/vnode.go`
- `../../runtime/ui/instance.go`
- `../../runtime/ui/fiber.go`
- `../../runtime/ui/fiber_util.go`
- `../../runtime/action`
- `../../runtime/intent`

常用 instance capability：

- `ComponentInstance`: 生命周期、props、key、dirty 状态。
- `PaintableInstance`: 自绘制组件。
- `FocusableInstance`: 可聚焦组件。
- `ActionHandlerInstance`: 处理 `*action.Action`。
- `RuntimeChildrenProvider`: 动态 overlay / popup / runtime children。
- `TickableInstance`: 需要周期 tick 的组件。
- `InstanceFactory`: VNode 创建持久实例。

## VNode

VNode 负责描述组件下一帧需要什么，并创建实例。

示意：

```go
type VNode struct {
    *rtui.ElementVNode
    props rtui.Props
}

func (v *VNode) CreateInstance() rtui.ComponentInstance {
    return NewInstance(v.props)
}
```

不要把 hover、pressed、focused、scroll offset 等可变运行时状态只存在 VNode 中。

## ComponentInstance

Instance 是组件运行期实体。它应保存：

- props 的当前副本。
- focus / hover / pressed / open / selected 等运行期状态。
- bounds / viewport / scroll 等运行时计算结果。
- action / intent 处理逻辑。

示意：

```go
type Instance struct {
    props rtui.Props
    key   string
    dirty bool
}

func (i *Instance) SetProps(props rtui.Props) bool {
    i.props = props
    i.MarkDirty()
    return true
}

func (i *Instance) MarkDirty() { i.dirty = true }
func (i *Instance) IsDirty() bool { return i.dirty }
```

具体接口签名以 `runtime/ui/instance.go` 为准。

## Paint

自绘制组件实现 `PaintableInstance`。Paint 不应读取 VNode。

```go
func (i *Instance) Paint(ctx paint.PaintContext, buf *paint.Buffer) {
    // Use instance props/state and current bounds.
}
```

当前 paint path 会通过 `FiberPaintableNode` 调用 `Fiber.Instance` 的绘制能力；若 instance 不可绘制，则可能使用按 tag 注册的 stateless paint fallback。

## Focus

可聚焦组件应通过 instance 能力接入 FocusManager：

- `SetFocus(bool)`
- `HasFocus() bool`
- disabled / focusable 判定相关方法

焦点状态写入 instance，不写入 VNode。Tab/Shift+Tab 和鼠标点击焦点都应通过 Fiber focus manager / app action path 进入。

## Action 和 Intent

组件内部低层交互走 `runtime/action`；面向业务语义的操作优先声明 `runtime/intent.Intent`。

Button 风格：

```go
type SaveIntent struct{}

func (SaveIntent) IntentType() string { return "Save" }

node := button.NewBuilder("Save").
    Primary().
    OnPress(SaveIntent{}).
    Build()
```

对于复杂组件，可在 instance 中实现 `HandleAction(*action.Action) bool`，或提供组件自己的 `HandleIntent(...)` 机制。

## Builder

Builder 是组件对外 API 的主要形态：

```go
node := table.NewBuilder().
    Columns([]table.TableColumn{
        {Title: "ID", Width: 6},
        {Title: "Name", Width: 20},
    }).
    AddRow("1", "Alice").
    ShowBorder(true).
    Build()
```

根 `ui` 包通过 `ui/shortcuts*.go` 暴露大量 `NewXBuilder` 和快捷函数。新增组件时，如果希望根包可直接使用，需要同步增加 shortcut 和测试。

## Hooks

Hooks 适合应用视图或函数组件里的局部状态。当前 `UseStateInt` 返回三个值：

```go
count, setCount, getCount := ui.UseStateInt(0)
```

示例：

```go
func Counter() ui.VNode {
    count, setCount, _ := ui.UseStateInt(0)

    return ui.VStack(
        ui.Text(fmt.Sprintf("Count: %d", count)),
        ui.NewButtonBuilder("+1").
            OnPress(IncrementIntent{}).
            Build(),
    )
}
```

跨组件业务状态更推荐 Store/Reducer/Intent，而不是把大量业务状态散落在 Hooks 中。

## Dynamic Lists And Keys

动态列表必须提供稳定 key。当前 reconciler 会对动态列表 key 策略做检查，避免 `_idx_N` fallback 在插入、删除、排序时造成状态错配。

推荐：

```go
for _, item := range items {
    children = append(children,
        itemView(item).SetKey(item.ID),
    )
}
```

不要用数组索引作为动态业务列表的长期 key。

## Runtime Children

Overlay、popup、dropdown、portal 类组件可通过 runtime children 机制提供运行期子节点。例如 Select、Popover、DatePicker、TimePicker 等组件会根据 open state 暴露 popup children。

开发这类组件时应关注：

- Portal root。
- Layer。
- Anchor / placement。
- viewport clamp / fallback。
- HitMap target。
- click outside。
- focus trap。

## Testing

组件包单元测试：

```bash
go test ./ui/components/<name> -count=1
```

组件库测试：

```bash
go test ./ui/components/... -count=1
```

E2E：

```bash
go test ./ui/e2e/... -count=1
go test ./ui/e2e -run "<ComponentName>" -count=1
```

资源受限环境下，先按组件包运行，再跑聚合测试。

## Development Checklist

- [ ] VNode 不保存可变运行时状态。
- [ ] Instance 实现必要 lifecycle 和 props 更新。
- [ ] Paint 使用 instance state，不依赖 VNode。
- [ ] 可聚焦组件接入 focus instance 能力。
- [ ] 鼠标交互考虑 HitMap / TargetBounds / TargetFiber。
- [ ] 业务交互优先用 Intent。
- [ ] 动态列表使用稳定 key。
- [ ] package README 包含当前 builder API 示例。
- [ ] 增加包内测试。
- [ ] 需要用户交互时增加或更新 `ui/e2e` 测试。
- [ ] 如暴露到根 `ui` 包，同步 `ui/shortcuts*.go` 和 shortcut 测试。

## 参考组件

- `../../ui/components/button`: 基础 action/intent/focus/paint 模式。
- `../../ui/components/input`: 文本输入和 focus。
- `../../ui/components/select`: popup、runtime children、overlay。
- `../../ui/components/tabs`: navigation、intent、field binding。
- `../../ui/components/table`: 数据展示、选择、分页、过滤。
- `../../ui/components/charts/linechart`: chart builder 和 text/image backend。

## 相关文档

- [../components/README.md](../components/README.md)
- [../architecture/README.md](../architecture/README.md)
- [../fiber/fiber_first/consolidated/README.md](../fiber/fiber_first/consolidated/README.md)
- [../ui/store/README.md](../ui/store/README.md)
- [../testing/e2e/README.md](../testing/e2e/README.md)
