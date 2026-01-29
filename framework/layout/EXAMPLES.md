# TUI 框架使用指南

本指南展示如何在 TUI 应用中使用 Scheduler 和 Flex 布局组件。

## 目录

- [快速开始](#快速开始)
- [Flex 布局使用](#flex-布局使用)
- [Scheduler 使用](#scheduler-使用)
- [完整示例](#完整示例)

## 快速开始

### 基本应用结构

```go
package main

import (
    "github.com/wwsheng009/mint/framework/layout"
    "github.com/wwsheng009/mint/runtime"
    "github.com/wwsheng009/mint/runtime/scheduler"
)

type App struct {
    rt      *runtime.Runtime
    sched   *scheduler.Scheduler
    root    *layout.Flex
}

func NewApp() *App {
    return &App{
        rt:    runtime.New(),
        sched: scheduler.New(),
        root:  buildRootLayout(),
    }
}

func (a *App) Run() error {
    // 主循环
    for {
        // 处理输入
        // ...

        // 渲染一帧
        a.renderFrame()
    }
}

func main() {
    app := NewApp()
    app.Run()
}
```

## Flex 布局使用

### 1. 创建简单布局

```go
func buildHeader() *layout.Flex {
    return layout.NewRow().
        Gap(2).
        Padding(1).
        Background("blue").
        AddChild(
            NewText("MyApp"),
            NewSpacer(),           // 占据剩余空间
            NewText("v1.0.0"),
        )
}

func buildMainLayout() *layout.Flex {
    return layout.NewColumn().
        AddChild(
            buildHeader(),
            buildContent(),
            buildFooter(),
        )
}
```

### 2. 表单布局

```go
type LoginForm struct {
    *layout.Flex
    username *Input
    password *Input
    loginBtn *Button
}

func NewLoginForm() *LoginForm {
    form := &LoginForm{
        username: NewInput(),
        password: NewInput(),
        loginBtn: NewButton("登录"),
    }

    form.Flex = layout.NewColumn().
        Gap(1).
        Padding(2).
        AddChild(
            NewLabel("用户名:"),
            form.username,
            NewLabel("密码:"),
            form.password,
            layout.NewRow().
                Gap(2).
                MainAlign(layout.MainEnd).
                AddChild(
                    NewButton("取消"),
                    form.loginBtn,
                ),
        )

    return form
}
```

### 3. 响应式布局

```go
func buildResponsiveLayout() *layout.Box {
    return layout.NewBox().
        WithBorder(true).
        WithChild(
            layout.NewRow().
                AddChild(
                    // 侧边栏 - 固定 20 列宽
                    layout.NewColumn().
                        Gap(1).
                        AddChild(
                            NewMenuItem("文件"),
                            NewMenuItem("编辑"),
                            NewMenuItem("视图"),
                        ),
                    // 内容区 - 弹性占据剩余空间
                    layout.NewColumn().
                        Gap(1).
                        Padding(1).
                        AddChild(
                            NewToolbar(),
                            NewEditor(),
                        ).
                        FlexGrow(0, 1), // 设置为可放大
                ),
        )
}
```

### 4. 弹性配置详解

```go
// 场景：三列布局，中间列自适应
func buildThreeColumn() *layout.Flex {
    return layout.NewRow().
        Gap(1).
        AddChild(
            // 左侧导航 - 固定 15 列
            NewNav(),

            // 中间内容 - 弹性
            NewContent().
                FlexGrow(0, 1), // 占据剩余空间

            // 右侧面板 - 固定 25 列
            NewPanel(),
        )
}

// 场景：按比例分配空间
func buildProportional() *layout.Flex {
    return layout.NewRow().
        Gap(1).
        AddChild(
            NewPanel1(), // 1 份
            NewPanel2(), // 2 份
            NewPanel3(), // 3 份
        ).
        Flex(0, layout.FlexConfig{Grow: 1}).
        Flex(1, layout.FlexConfig{Grow: 2}).
        Flex(2, layout.FlexConfig{Grow: 3})
}
```

### 5. 嵌套布局

```go
func buildComplexLayout() *layout.Flex {
    return layout.NewColumn().
        Gap(1).
        AddChild(
            // 顶部栏
            layout.NewRow().
                Padding(1).
                Background("gray").
                AddChild(
                    NewLogo(),
                    NewTitle("Dashboard"),
                    NewUserInfo(),
                ),

            // 主内容区 - 两列
            layout.NewRow().
                Gap(1).
                AddChild(
                    // 左侧
                    layout.NewColumn().
                        Gap(1).
                        AddChild(
                            NewStatsCard(),
                            NewChartCard(),
                        ).
                        FlexGrow(0, 1),
                    // 右侧
                    layout.NewColumn().
                        Gap(1).
                        AddChild(
                            NewActivityLog(),
                            NewQuickActions(),
                        ).
                        FlexGrow(0, 1),
                ),
        )
}
```

## Scheduler 使用

### 1. 基本集成

```go
type App struct {
    sched *scheduler.Scheduler
}

func (a *App) markDirty(component Component, level priority.DirtyLevel) {
    a.sched.MarkDirty(component.ID(), component, level)
}

func (a *App) renderFrame() {
    renderer := &FrameworkRenderer{buf: a.rt.Screen()}
    result := a.sched.ProcessNext(renderer, scheduler.DefaultProcessOptions())

    // 将渲染结果输出到屏幕
    a.rt.Flush()

    // 如果有剩余脏节点，继续处理
    if result.Remaining > 0 {
        // 可以在下一帧继续处理
    }
}
```

### 2. 渲染器实现

```go
type FrameworkRenderer struct {
    buf *paint.Buffer
}

func (r *FrameworkRenderer) Layout(node interface{}) {
    if n, ok := node.(component.Positionable); ok {
        // 布局逻辑
        n.PerformLayout()
    }
}

func (r *FrameworkRenderer) Paint(node interface{}) {
    if n, ok := node.(component.Paintable); ok {
        ctx := component.NewPaintContext(
            r.buf,
            n.GetX(), n.GetY(),
            n.GetWidth(), n.GetHeight(),
        )
        n.Paint(ctx, r.buf)
    }
}
```

### 3. 批量更新优化

```go
type App struct {
    sched *scheduler.Scheduler
}

// 批量更新场景
func (a *App) LoadData(items []Item) {
    a.sched.BeginBatch()
    defer a.sched.EndBatch(true)

    for _, item := range items {
        component := a.createComponent(item)
        a.sched.MarkDirty(component.ID(), component, priority.DirtyNormal)
    }
}

// 分页加载场景
func (a *App) LoadPage(page int) {
    items := fetchPage(page)

    a.sched.BeginBatch()
    for _, item := range items {
        comp := a.createComponent(item)
        a.root.AddChild(comp)
        a.sched.MarkDirty(comp.ID(), comp, priority.DirtyNormal)
    }
    a.sched.EndBatch(true)
}
```

### 4. 优先级策略

```go
func (a *App) updateComponent(c Component, changeType ChangeType) {
    var level priority.DirtyLevel

    switch changeType {
    case ChangeCritical:  // 关键交互（点击、输入）
        level = priority.DirtyHigh
    case ChangeContent:   // 内容变化
        level = priority.DirtyNormal
    case ChangeStyle:     // 样式变化
        level = priority.DirtyLow
    }

    a.sched.MarkDirty(c.ID(), c, level)
}

// 示例
func (a *App) OnButtonClick(btn *Button) {
    btn.SetState("pressed")
    a.updateComponent(btn, ChangeCritical) // 高优先级
}

func (a *App) OnProgressUpdate(pbar *ProgressBar, val int) {
    pbar.SetValue(val)
    a.updateComponent(pbar, ChangeNormal) // 普通优先级
}
```

## 完整示例

### 带状态管理的应用

```go
package main

import (
    "github.com/wwsheng009/mint/framework/layout"
    "github.com/wwsheng009/mint/framework/component"
    "github.com/wwsheng009/mint/runtime"
    "github.com/wwsheng009/mint/runtime/scheduler"
    "github.com/wwsheng009/mint/runtime/priority"
    "github.com/wwsheng009/mint/runtime/paint"
)

type DashboardApp struct {
    rt         *runtime.Runtime
    sched      *scheduler.Scheduler
    root       *layout.Flex
    sidebar    *layout.Flex
    content    *layout.Flex
    statusText *Text
}

func NewDashboardApp() *DashboardApp {
    app := &DashboardApp{
        rt:    runtime.New(),
        sched: scheduler.New(),
    }

    app.buildUI()
    return app
}

func (a *DashboardApp) buildUI() {
    // 构建侧边栏
    a.sidebar = layout.NewColumn().
        Gap(0).
        Padding(0).
        Background("gray").
        AddChild(
            NewMenuItem("📁 Dashboard", true),
            NewMenuItem("📊 Analytics"),
            NewMenuItem("📝 Reports"),
            NewMenuItem("⚙️  Settings"),
        )

    // 构建内容区
    a.content = layout.NewColumn().
        Gap(1).
        Padding(1).
        AddChild(
            NewText("Welcome to Dashboard"),
            NewStatsRow(),
            NewChart(),
        )

    // 构建根布局
    a.root = layout.NewRow().
        AddChild(
            // 侧边栏 - 固定宽度
            a.sidebar,
            // 内容区 - 弹性
            a.content.FlexGrow(0, 1),
        )

    // 状态栏
    a.statusText = NewText("Ready")
}

func (a *DashboardApp) Run() error {
    if err := a.rt.Init(); err != nil {
        return err
    }
    defer a.rt.Close()

    // 初始渲染
    a.markAllDirty()
    a.renderFrame()

    // 主循环
    for {
        event, err := a.rt.NextEvent()
        if err != nil {
            return err
        }

        if a.handleEvent(event) {
            a.renderFrame()
        }

        if a.shouldQuit {
            return nil
        }
    }
}

func (a *DashboardApp) renderFrame() {
    renderer := &appRenderer{
        rt:   a.rt,
        buf:  a.rt.Screen(),
        root: a.root,
    }

    result := a.sched.ProcessNext(renderer, scheduler.DefaultProcessOptions())
    a.rt.Flush()

    // 更新状态
    a.statusText.SetText(fmt.Sprintf("Rendered: %d, Remaining: %d",
        result.Processed, result.Remaining))
}

func (a *DashboardApp) markAllDirty() {
    a.markComponentDirty(a.root, priority.DirtyHigh)
}

func (a *DashboardApp) markComponentDirty(comp component.Node, level priority.DirtyLevel) {
    a.sched.MarkDirty(comp.ID(), comp, level)

    // 标记子组件
    if container, ok := comp.(component.Container); ok {
        for i := 0; i < container.ChildCount(); i++ {
            child := container.GetChild(i)
            a.markComponentDirty(child, level)
        }
    }
}

// 渲染器实现
type appRenderer struct {
    rt   *runtime.Runtime
    buf  *paint.Buffer
    root component.Node
}

func (r *appRenderer) Layout(node interface{}) {
    // 布局逻辑由 Flex 组件内部处理
}

func (r *appRenderer) Paint(node interface{}) {
    if n, ok := node.(component.Paintable); ok {
        x, y := n.GetPosition()
        w, h := n.GetSize()

        ctx := component.NewPaintContext(r.buf, x, y, w, h)
        n.Paint(ctx, r.buf)
    }
}

func main() {
    app := NewDashboardApp()
    if err := app.Run(); err != nil {
        panic(err)
    }
}
```

### 带实时更新的应用

```go
type MonitoringApp struct {
    sched    *scheduler.Scheduler
    logPanel *VirtualList
    metrics  *MetricsPanel
}

func (a *MonitoringApp) OnLogEntry(entry LogEntry) {
    // 使用批处理处理大量日志
    a.sched.BeginBatch()

    a.logPanel.AddItem(entry)
    a.sched.MarkDirty(a.logPanel.ID(), a.logPanel, priority.DirtyNormal)

    // 限制批次大小
    if a.sched.GetBatchSize() >= 100 {
        a.sched.FlushBatch()
        a.renderFrame()
    }

    a.sched.EndBatch(true)
}

func (a *MonitoringApp) OnMetricsUpdate(metrics Metrics) {
    // 高优先级更新指标
    a.metrics.Update(metrics)
    a.sched.MarkDirty(a.metrics.ID(), a.metrics, priority.DirtyHigh)
    a.renderFrame()
}

func (a *MonitoringApp) renderFrame() {
    renderer := &MyRenderer{}
    a.sched.ProcessNext(renderer, scheduler.ProcessOptions{
        TimeBudget: 5 * time.Millisecond, // 限制渲染时间
    })
}
```

## 最佳实践总结

### Flex 布局

1. **使用便捷构造器**: `NewRow()` / `NewColumn()` 覆盖 90% 场景
2. **合理嵌套**: 复杂布局拆分为多个简单容器
3. **弹性分配**: 用 `FlexGrow` 处理响应式空间分配
4. **间距管理**: 用 `Gap()` 而不是手动添加 Space 组件
5. **与 Box 组合**: Box 处理边框/背景，Flex 处理布局

### Scheduler

1. **批处理更新**: 一次更新多个组件时使用 `BeginBatch/EndBatch`
2. **优先级策略**: 交互用 High，内容用 Normal，装饰用 Low
3. **时间预算**: 根据目标帧率调整，一般 2-5ms per priority
4. **监控剩余**: `Remaining > 0` 时考虑下一帧继续处理
5. **与事件循环配合**: 每次事件处理后调用 `ProcessNext`
