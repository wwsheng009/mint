# Group B/C Scene 与 `App.render()` 接入规格

## 1. 文档目的

本文件把 `PHASE1_TASK_BREAKDOWN.md` 里的 Group B 和 Group C 合并细化，解决两个最关键的问题：

1. 在当前 `paint.Buffer + Renderer + AsyncRenderer` 主链中，图像层应该怎么承载
2. `framework.App.render()` 第一阶段应该如何接入，而不把现有文本主链打乱

这份规格刻意比总架构更落地，也比组件实现更克制。

它的目标不是定义未来最终完美形态，而是定义：

> 第一阶段最小可行、可验证、可回退的 Scene 与 render-loop 接入方式

## 2. 当前代码主链复核

在设计 Group B/C 前，必须基于当前真实代码，而不是抽象想象。

### 2.1 当前 `App.render()` 的真实职责

见：

- `framework/app.go`

当前 `App.render()` 会做这些事：

1. 获取 `renderer.GetBackBuffer()`
2. `buf.Reset(terminalWidth, terminalHeight)`
3. 调用 `paintable.Paint(ctx, buf)`
4. 应用 app-selection 高亮
5. 更新交互实例注册表
6. 走同步 `renderer.Render()` 或异步 `asyncRenderer.SubmitFrame(...)`
7. render 后更新 `HitMap`
8. 通知 AI / tickable / dirty 状态

也就是说：

- `App.render()` 不是简单的“把 buffer 打印出来”
- 它已经承担了完整帧边界控制

### 2.2 当前既有两个文本输出路径

见：

- `runtime/paint/renderer.go`
- `runtime/paint/async_renderer.go`

当前存在：

- 同步 `Renderer`
- 异步 `AsyncRenderer`

其中：

- `Renderer` 负责文本双缓冲、line diff、dirty rect
- `AsyncRenderer` 负责文本帧暂存、区域 diff、节流输出

这说明：

- Group C 不能只考虑同步路径
- 也不能在第一阶段强行把图像塞进 `AsyncRenderer`，否则复杂度会立刻膨胀

### 2.3 当前 `HitMap` 仍是文本布局派生物

见：

- `framework/app.go`
- `framework/event/pump.go`

当前 `HitMap` 是渲染完成后，由布局/渲染树提供给 `Pump` 的。

第一阶段如果 image layer 只是视觉增强、不承载复杂交互，那么：

- `HitMap` 可以继续保持文本主导

这是一条重要的简化路径。

## 3. 第一阶段必须固定的架构决策

### 3.1 决策一：`paint.Buffer` 继续作为文本层唯一载体

第一阶段不改：

- `paint.Buffer` 的 cell 语义
- `Renderer` 的文本职责

图片图层不能塞进 `Buffer.Cells`。

### 3.2 决策二：新增 Scene / Frame 包装层，而不是替换 `Paint()` 契约

当前 `component.Paintable.Paint(ctx, buf)` 已经广泛存在。

第一阶段不建议直接推翻它。

更稳的方式是：

- `Paint()` 继续负责文本层
- Scene 作为额外包装层承载图像层

### 3.3 决策三：Phase 1 的 image frame 一律走同步提交

这是本规格最重要的一个决定。

原因很直接：

- 当前 `AsyncRenderer` 只懂文本 `Buffer`
- 它的 pending region、stage buffer、region diff 都是文本语义
- 如果第一阶段就让它承担图像层，会把 Group C 直接推成高风险重构

因此建议：

- 纯文本帧：继续走现有同步/异步路径
- 含图像层的实验帧：Phase 1 一律走同步提交

这样能显著降低第一阶段复杂度。

### 3.4 决策四：Phase 1 的 image layer 默认视为“非交互视觉层”

也就是说：

- 图像层暂不参与复杂 hit testing
- 当前 `HitMap` 继续来源于文本布局树
- 如果某个 image chart 需要交互，Phase 1 先不支持

这是为了把第一阶段控制在“渲染可行性验证”范围内。

## 4. Scene / Frame 的最小数据模型

### 4.1 `SceneFrame`

建议新增一个最小包装结构，例如：

```go
type SceneFrame struct {
    Text   *Buffer
    Images []ImageLayer
}
```

这里的关键点是：

- `Text` 仍然是现有 `paint.Buffer`
- `Images` 是增量图层

第一阶段不需要更重的 compositor 树。

### 4.2 `ImageLayer`

建议最小结构如下：

```go
type ImageLayer struct {
    ID          string
    CellRect    Rect
    PixelWidth  int
    PixelHeight int
    RGBA        []byte
    Hash        uint64
    ZIndex      int
    AltText     string
}
```

字段说明：

- `ID`
  - 图像对象标识；用于 replace/delete
- `CellRect`
  - 图像在现有布局中的 cell 区域
- `PixelWidth / PixelHeight`
  - 最终 raster 结果尺寸
- `RGBA`
  - 第一阶段可接受的直接像素载荷
- `Hash`
  - 用于缓存命中判断
- `ZIndex`
  - 第一阶段可保留，但不要求做复杂层次竞争
- `AltText`
  - diagnostics / fallback / 调试说明

### 4.3 `SceneDiagnostics`

建议第一阶段就预留最小诊断结构：

```go
type SceneDiagnostics struct {
    GraphicsMode string
    ImageCount   int
    Notes        []string
}
```

第一阶段不要求它进入所有主链，但 prototype 和 benchmark 很需要它。

## 5. Scene 的构造方式

这是第一阶段最容易失控的点，必须先定边界。

### 5.1 不改 `component.Paintable` 主接口

第一阶段不建议把 `Paint(ctx, buf)` 改成：

- `PaintScene(...)`
- `PaintFrame(...)`
- 返回复合结构

因为这样会冲击整个组件生态。

### 5.2 使用可选扩展接口

建议新增一个实验性可选接口，例如：

```go
type ExperimentalSceneProvider interface {
    BuildExperimentalScene(ctx component.PaintContext) *paint.SceneFrame
}
```

Phase 1 的特点是：

- 只有少数 prototype 页面或根节点实现它
- 普通页面完全不需要感知它

### 5.3 Scene 与文本层的关系

推荐关系是：

- 文本层仍然由 `Paint()` 直接写入 `Buffer`
- `BuildExperimentalScene()` 只补充 `Images`
- `SceneFrame.Text` 可直接引用当前 back buffer

这样可以避免“双份文本渲染”。

## 6. `App.render()` 的最小接入方案

### 6.1 保持现有文本帧构建顺序不变

也就是说：

1. `GetBackBuffer()`
2. `Reset()`
3. `Paint()`
4. selection / interaction / debug

这些顺序 Phase 1 不改。

### 6.2 在文本帧构建完成后，可选收集 Scene

建议新增逻辑：

- 如果根节点实现 `ExperimentalSceneProvider`
- 则在文本层完成后收集 `SceneFrame`

注意：

- 不是先建 Scene 再画文本
- 而是文本层先完成，再补图像层

### 6.3 Phase 1 图像帧的推荐输出顺序

推荐顺序如下：

1. 先渲染文本层
2. 再提交图像层
3. 再更新 `HitMap`

原因：

- 当前 `HitMap` 主要依赖文本布局树
- Phase 1 图像层不承担复杂交互
- 先稳定文本，再叠加图像，更容易诊断

### 6.4 为什么不先更新 `HitMap`

因为第一阶段图像层没有引入自己的 hit-test 模型。

如果未来 image layer 要参与交互，才需要重新设计：

- image region -> chart domain

但那不是 Group C 的目标。

## 7. 同步路径与异步路径的处理策略

### 7.1 Phase 1 对异步渲染的处理原则

**如果帧里包含 image layer，Phase 1 推荐直接绕过 `AsyncRenderer`。**

原因：

- `AsyncRenderer` 当前只认识文本 `Buffer`
- 它的 `stage`, `pendingRects`, `RegionDiff`, `copyBufferRect` 都是文本设计
- 如果为了 PoC 修改它，会显著扩大写入范围

### 7.2 具体建议

建议逻辑是：

- `SceneFrame.Images == 0`
  - 保持现有同步/异步路径
- `SceneFrame.Images > 0`
  - 强制使用同步文本 `renderer.Render()`
  - 然后调用 `GraphicsPresenter`

这样可以让第一阶段的时序简单且可控。

### 7.3 Phase 2 再考虑什么

等 Phase 1 验证通过后，Phase 2 才考虑：

- 异步图像帧暂存
- 文本 + 图像统一节流
- 图像 dirty region
- 图像缓存和替换策略

## 8. 生命周期与清理

### 8.1 必须定义图像对象生命周期

只要图像层进入 alternate screen，就必须明确：

- 首帧创建
- 更新时替换
- 页面关闭时删除
- App 退出时清理

### 8.2 清理责任放在哪

Phase 1 建议由 `GraphicsPresenter` 负责：

- `Replace`
- `Delete`
- `Clear`

`App.Close()` 和 `ExitAltScreen()` 只负责触发统一清理，不直接操作协议细节。

### 8.3 失败时的退回策略

如果图像提交失败：

- 记录 diagnostics
- 本帧继续保留文本层
- 后续帧自动降级到 text mode

不要因为图像后端失败而影响整个页面输出。

## 9. Prototype 场景下的最小实现建议

### 9.1 Prototype 页面优先

第一阶段不建议先把 image path 塞回所有 `linechart` 真实页面。

更稳的顺序是：

- 先在 `examples/charts_linechart_image_prototype/` 下做 prototype 页面
- prototype 根节点实现 `ExperimentalSceneProvider`
- `linechart` 的实验性 image renderer 先只服务这个 demo

### 9.2 为什么这比直接改通用组件更稳

因为这样可以：

- 控制写入范围
- 控制回滚成本
- 把“渲染链验证”和“公开组件 API 设计”拆开

如果 PoC 失败，不会把整个 charts 组件层一起拖下水。

## 10. 对现有测试体系的影响

### 10.1 文本测试继续保留

当前：

- 组件单测
- 文本 e2e
- 文本 snapshot

都继续有效，因为它们覆盖的是 text fallback 路径。

### 10.2 图像路径暂时不强求并入现有文本 snapshot 体系

第一阶段只要求：

- prototype demo
- capability dump
- image diagnostics
- 最小 benchmark

不要在 Group C 阶段就试图把所有图像路径做成完整 CI golden 体系。

## 11. 验收标准

### 11.1 架构验收

- `paint.Buffer` 主模型不被破坏
- `component.Paintable` 主接口不被推翻
- `App.render()` 对无图像页面行为不变

### 11.2 时序验收

- 图像帧不会打乱文本帧输出
- alternate screen 退出后无图像残留
- 图像失败时文本页面仍正常

### 11.3 Phase 1 复杂度控制验收

- `AsyncRenderer` 可保持原样工作于文本帧
- 图像帧不要求立刻进入 async pipeline
- prototype 级别可验证即可，不追求全局接入完美

## 12. 关键风险

### 12.1 风险一：试图一口气把图像接进 `AsyncRenderer`

这是 Group C 最大的复杂度陷阱。

建议 Phase 1 明确规避。

### 12.2 风险二：过早改动组件公共契约

如果一开始就修改 `Paintable` 或所有 chart builder，回滚成本会很高。

### 12.3 风险三：图像层承担交互责任

这会把 hit testing、viewport、selection 一起卷进来，超出 Phase 1 范围。

## 13. 一句话结论

Group B/C 的正确方向不是“让图像赶紧跑起来”，而是：

**在保留当前文本主链的前提下，用最小 Scene 包装层和同步图像提交路径，完成第一阶段的可行性闭环。**

## 14. 下一步实现参考

如果要从 Group B/C 规格直接进入编码，建议紧接着参考：

- [RUNTIME_PAINT_SCENE_API_SKETCH.md](./RUNTIME_PAINT_SCENE_API_SKETCH.md)
- [FRAMEWORK_APP_RENDER_IMAGE_FLOW_SPEC.md](./FRAMEWORK_APP_RENDER_IMAGE_FLOW_SPEC.md)

它们分别把 Group B/C 继续收敛成：

- `runtime/paint/scene.go` 的类型草案
- `framework.App.render()` 的时序、辅助方法与伪代码
