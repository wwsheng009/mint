# `runtime/paint/scene` API 草案与最小数据模型

## 1. 文档目的

`GROUP_BC_SCENE_AND_APP_RENDER_INTEGRATION_SPEC.md` 已经说明：

- 第一阶段不能把图像塞进 `paint.Buffer`
- 需要一个最小 Scene 包装层
- 这个包装层要能和当前 `App.render()`、`Renderer`、`AsyncRenderer` 共存

但如果要实际编码，仍然缺一个更贴近实现的回答：

> `runtime/paint/scene.go` 第一版应该定义哪些类型、哪些字段、哪些接口。

本文件就是这份 API 草案。

## 2. 当前 `paint` 层的现实边界

当前 `runtime/paint` 里最核心的几个现实约束是：

- [buffer.go](/E:/projects/yao/wwsheng009/mint/runtime/paint/buffer.go) 仍然是纯 cell 文本缓冲
- [renderer.go](/E:/projects/yao/wwsheng009/mint/runtime/paint/renderer.go) 的输出仍然是 ANSI 文本
- [async_renderer.go](/E:/projects/yao/wwsheng009/mint/runtime/paint/async_renderer.go) 的 `stage` 和 `RegionDiff` 也都是文本语义

所以第一阶段 Scene 的设计必须满足：

1. 不污染 `Buffer`
2. 不破坏 `Renderer`
3. 允许 `App.render()` 在文本层之外拿到额外图像层

## 3. 第一阶段 Scene 的职责边界

### 3.1 Scene 负责什么

第一阶段 Scene 只负责：

- 把“这一帧的文本层”和“这一帧的图像层”组织到一起
- 为 `App.render()` 提供统一的数据载体
- 为 diagnostics 和 prototype 提供最小元数据

### 3.2 Scene 不负责什么

第一阶段 Scene 不负责：

- 替代 `Renderer`
- 实现图像协议输出
- 统一文本和图像的复杂 z-order 合成
- 全局异步调度
- 交互命中测试

## 4. 推荐文件拆分

第一阶段建议先只新增一个文件：

- `runtime/paint/scene.go`

如果后续扩展需要，也可以再拆：

- `scene_diag.go`
- `image_layer.go`

但第一阶段不必过早拆多文件。

## 5. 建议的最小类型集

### 5.1 `SceneFrame`

这是第一阶段最核心的结构。

建议草案：

```go
type SceneFrame struct {
    Text        *Buffer
    Images      []ImageLayer
    Diagnostics *SceneDiagnostics
}
```

字段说明：

- `Text`
  - 当前帧文本层，仍然是现有 `paint.Buffer`
- `Images`
  - 当前帧附加图像层
- `Diagnostics`
  - 这帧的附加调试信息，可选

### 5.2 `ImageLayer`

建议第一阶段最小结构如下：

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

字段解释：

- `ID`
  - 图像对象 ID；允许上层显式提供，也允许后续由 presenter 分配
- `CellRect`
  - 图像逻辑覆盖的 cell 区域
- `PixelWidth / PixelHeight`
  - 位图真实尺寸
- `RGBA`
  - 第一阶段直接裸像素，不引入复杂编码类型
- `Hash`
  - 可用于缓存判断
- `ZIndex`
  - 第一阶段保留字段，但只做最简单排序
- `AltText`
  - diagnostics / fallback / 调试说明

### 5.3 `SceneDiagnostics`

建议最小结构：

```go
type SceneDiagnostics struct {
    GraphicsMode string
    ImageCount   int
    Notes        []string
}
```

这层的目标是：

- 帮助 prototype 输出可读诊断
- 帮助后续 image path 出错时定位

## 6. 为什么不直接把 `DrawImageRequest` 放进 `SceneFrame`

这是一个需要提前说明的设计点。

### 6.1 `SceneFrame` 存 `ImageLayer`，而不是 presenter 请求

推荐存：

- `ImageLayer`

不推荐直接存：

- `platform.DrawImageRequest`

原因：

- `SceneFrame` 应属于 `paint` 层，不应反向依赖 `platform`
- `ImageLayer` 更接近渲染语义
- `DrawImageRequest` 是 presenter 层的提交语义

### 6.2 转换责任放在哪

推荐在 Group C 接入层做：

- `ImageLayer -> DrawImageRequest`

这样可以保持：

- `paint` 不反向依赖 `platform`
- `platform` 不感知 `Buffer`

## 7. 建议的辅助接口

### 7.1 `ExperimentalSceneProvider`

这个接口不一定要定义在 `paint` 包里，但它必须依赖 `SceneFrame`。

建议草案：

```go
type ExperimentalSceneProvider interface {
    BuildExperimentalScene(ctx component.PaintContext, text *paint.Buffer) *paint.SceneFrame
}
```

这里我建议把 `text *paint.Buffer` 显式传进去，而不是只传 `ctx`。

理由：

- 第一阶段 image layer 往往需要复用文本布局阶段已经算出来的尺寸和边界
- 有些 prototype 根节点可能希望直接在已有文本帧上附加图像层

如果不想让接口依赖 `paint.Buffer`，也可以退回更保守版本：

```go
BuildExperimentalScene(ctx component.PaintContext) *paint.SceneFrame
```

但就第一阶段实现便利性而言，传入文本层引用更实用。

## 8. 建议的最小辅助方法

第一阶段 Scene 至少建议附带这几个 helper。

### 8.1 `HasImages()`

```go
func (f *SceneFrame) HasImages() bool
```

用途：

- `App.render()` 快速判断是否需要走图像提交路径

### 8.2 `SortImageLayers()`

```go
func (f *SceneFrame) SortImageLayers()
```

用途：

- 第一阶段只按 `ZIndex` 做稳定排序

### 8.3 `CloneShallow()`

```go
func (f *SceneFrame) CloneShallow() *SceneFrame
```

用途：

- 便于 diagnostics / benchmark 记录
- 避免 prototype 里直接共享可变切片导致混乱

### 8.4 `Summary()`

```go
func (f *SceneFrame) Summary() string
```

用途：

- 调试输出
- prototype 页面或日志摘要

## 9. 内存与复制策略

### 9.1 第一阶段避免深拷贝整张位图

`RGBA []byte` 可能很大，因此第一阶段不要默认在 Scene helper 里深拷贝。

建议策略：

- `SceneFrame` 默认持有 image layer 的引用
- 如需持久化 diagnostics，再显式导出

### 9.2 Scene 不负责像素缓存

第一阶段 Scene 只是帧容器，不是缓存层。

以下职责不要提前塞进 Scene：

- bitmap 缓存
- tile 缓存
- payload 缓存

这会让 Group B 复杂度失控。

## 10. Phase 1 的时序定位

Scene 的正确时序位置应是：

1. 文本 `Paint()`
2. 构造 `SceneFrame`
3. 文本层输出
4. 图像层提交

也就是说，Scene 是帧组织层，不是底层绘制器。

## 11. 推荐测试

### 11.1 `scene_test.go`

建议覆盖：

- `HasImages()`
- `SortImageLayers()`
- `Summary()`

### 11.2 `SceneFrame` 结构完整性测试

建议验证：

- `Text=nil` 的非法场景
- `Images=nil` 的空图像帧场景
- `Diagnostics=nil` 的可选场景

### 11.3 顺序测试

如果实现了排序，建议验证：

- `ZIndex` 升序/稳定顺序

## 12. 与 Group C 的边界

Scene 只回答：

- 这一帧有哪些层

它不回答：

- 如何 `fmt.Print`
- 如何走 `AsyncRenderer`
- 如何提交到 `GraphicsPresenter`

这些都应留给 `App.render()` 集成层。

## 13. 一句话结论

`runtime/paint/scene` 第一阶段最正确的定位是：

**用极小的数据结构把“文本帧 + 图像层”组织起来，而不是试图在 `paint` 层提前完成整套混合渲染器。**
