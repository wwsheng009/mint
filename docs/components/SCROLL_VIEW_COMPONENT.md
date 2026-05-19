# ScrollView 组件

`ScrollView` 是 Mint 当前的滚动容器组件，源码位于 `ui/components/scrollview/`。它适合展示长文本、日志、树形文本快照、代码片段等固定视口内容。

当前实现是 Fiber-first 组件：

```text
scrollview.VNode
  -> scrollview.Instance
  -> Measure(...)
  -> Paint(...)
```

组件声明只保存 props；滚动状态、内容行缓存、边界和 paint 行为由运行期 `Instance` 管理。

## 当前源码位置

| 内容 | 路径 |
|---|---|
| Builder API | `ui/components/scrollview/builder.go` |
| VNode props | `ui/components/scrollview/vnode.go` |
| Runtime instance / paint / action handling | `ui/components/scrollview/instance.go` |
| 根包快捷入口 | `ui/shortcuts.go` |
| E2E 覆盖 | `ui/e2e/scrollview_e2e_test.go` |
| SDK 示例 | `examples/mvp_components_demo`、`examples/ui_demos/` |

不要再引用历史路径 `components/layout/scroll_view.go`。

## API 入口

推荐从根包 `ui` 使用：

```go
ui.NewScrollViewBuilder().
    Child(ui.Text("Line 1\nLine 2\nLine 3")).
    Width(40).
    Height(8).
    ScrollOffset(0).
    ShowBorder(true).
    Build()
```

常用快捷函数：

```go
ui.ScrollView(ui.Text("Long content"))
ui.ScrollSize(ui.Text("Long content"), 40, 8)
ui.ScrollBordered(ui.Text("Long content"), 40, 8)
```

组件包也提供同等入口：

```go
import "github.com/wwsheng009/mint/ui/components/scrollview"

scrollview.NewBuilder().
    Child(ui.Text(content)).
    Width(60).
    Height(12).
    ShowBorder(true).
    Build()
```

## Builder 方法

| 方法 | 说明 |
|---|---|
| `Child(rtui.VNode)` | 设置要滚动展示的内容 |
| `Width(int)` | 设置内容视口宽度，`0` 表示自动宽度 |
| `Height(int)` | 设置内容视口高度，`0` 表示显示全部内容 |
| `ScrollOffset(int)` | 设置非受控模式下的初始滚动行偏移 |
| `InitialScrollOffset(int)` | `ScrollOffset` 的语义化别名 |
| `ScrollOffsetControlled(int)` | 设置受控滚动偏移，每次 props 更新都会覆盖实例状态 |
| `ShowBorder(bool)` / `Border()` / `NoBorder()` | 控制边框 |
| `ShowIndicator(bool)` | 控制边框右侧滚动指示符 |
| `Style(style.Style)` | 设置内容和边框使用的基础样式 |
| `Key(string)` | 设置 diff key |
| `SetID(string)` | 设置业务 ID，用于定位和 Portal anchoring |

`Build()` 返回 `rtui.VNode`，`BuildVNode()` 返回具体的 `*scrollview.VNode`。

## 内容模型

`ScrollView` 当前以文本行为核心模型。`Instance.extractContent()` 会从 child 中提取文本内容：

- 读取 props 中的 `content`。
- 识别 `ui/components/text.VNode`。
- 识别实现 `Content() string` 的节点。
- 对 children 递归提取，并用换行拼接。

因此它最适合长文本和按行展示的内容。需要复杂行组件、选中项、高亮项或大数据列表时，优先使用 `VirtualList` 或 `List.BuildVirtualList()`。

## 滚动行为

`ScrollView` 支持两种滚动使用方式：

1. 非受控初始偏移：使用 `ScrollOffset()` / `InitialScrollOffset()` 设置首次偏移，之后实例可以通过 action 自己改变滚动位置。
2. 受控偏移：使用 `ScrollOffsetControlled()`，父组件每次 render 传入新的偏移，实例会同步到该值。

运行期实例支持：

```go
inst.ScrollBy(delta)
inst.ScrollTo(offset)
inst.ScrollTop()
inst.ScrollBottom()
inst.PageUp()
inst.PageDown()
inst.CanScrollUp()
inst.CanScrollDown()
inst.GetScrollOffset()
inst.GetTotalLines()
inst.GetViewportSize()
inst.IsScrollable()
```

这些方法在组件实例层存在。普通应用代码通常通过 action、状态和重新 render 来驱动，而不是直接持有实例。

## Action 支持

`scrollview.Instance` 实现了 action target 能力，支持：

| Action | 行为 |
|---|---|
| `ActionScroll` | 根据滚轮 delta 滚动 |
| `ActionNavigateUp` / `ActionNavigateDown` | 上下滚动一行 |
| `ActionNavigatePageUp` / `ActionNavigatePageDown` | 翻页滚动 |
| `ActionNavigateHome` / `ActionNavigateEnd` | 到顶部 / 底部 |

鼠标滚轮会经由 `framework/event.Pump -> runtime/msg.MouseMsg -> runtime/action.InputProcessor -> framework.App.processMsg` 进入 action 路由。命中目标由 HitMap / `TargetFiber` 提供。

## 渲染细节

- 无边框时，组件只绘制视口内可见行。
- 有边框时，实际绘制尺寸为 `width + 2` 和 `height + 2`。
- 内容超过宽度会被截断。
- `ShowIndicator(true)` 且内容高度超过视口时，右边框底部会显示滚动位置提示。
- 高度为 `0` 时按全部内容高度测量，通常不会产生垂直滚动。

## 示例：日志视图

```go
func LogPanel(lines []string, offset int) ui.VNode {
    return ui.NewScrollViewBuilder().
        Child(ui.Text(strings.Join(lines, "\n"))).
        Width(80).
        Height(18).
        ScrollOffsetControlled(offset).
        ShowBorder(true).
        ShowIndicator(true).
        Build()
}
```

## 示例：代码片段

```go
func CodePreview(source string) ui.VNode {
    return ui.ScrollBordered(
        ui.Text(source),
        72,
        16,
    )
}
```

## ScrollView 与 VirtualList

| 场景 | 推荐组件 |
|---|---|
| 长文本、日志、代码、树形文本快照 | `ScrollView` |
| 大量字符串项、需要 selected index、每项样式函数 | `VirtualList` |
| 富行模型、选择/搜索/checkbox 等列表交互 | `List` 或 `List.BuildVirtualList()` |

## 测试与验证

相关测试：

```bash
go test ./ui/components/scrollview -count=1
go test ./ui/e2e -run ScrollView -count=1
```

当前文档描述的是 `ui/components/scrollview` 的现行实现；如果历史设计文档提到 `components/layout/scroll_view.go`，应视为旧路径。

旧的 `examples/fiber_firsts/scrollview_demo` 是低层 Fiber 渲染探针，已归档到 `docsArchive/cleanup-2026-05-19/_examples/fiber_firsts/scrollview_demo/`。当前 SDK 学习路径以组件文档、E2E 覆盖和更高层示例为准。
