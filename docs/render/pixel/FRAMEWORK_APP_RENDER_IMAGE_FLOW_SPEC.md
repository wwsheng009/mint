# `framework.App.render()` 图像帧接入流程规格

## 1. 文档目的

前面的文档已经把：

- 图形能力层
- Scene 结构
- `linechart` image renderer
- prototype 验证闭环

分别拆清了。

现在还缺一份真正关键的桥梁文档：

> 当一帧里既有文本层又有 image layer 时，`framework.App.render()` 第一阶段到底要按什么顺序做事。

本文件的目标就是把这条时序写清楚，避免未来一边写代码一边重新争论渲染顺序。

## 2. 当前 `App.render()` 的现实路径

对照：

- [framework/app.go](/E:/projects/yao/wwsheng009/mint/framework/app.go)

当前主流程可以概括成：

1. 准备 `back buffer`
2. `Paint()` 文本层
3. 处理 selection / interaction / debug
4. 生成 `dirtyHints`
5. 同步走 `renderer.Render()`，或异步走 `asyncRenderer.SubmitFrame()`
6. render 后更新 `HitMap`
7. 通知 AI / tickable / dirty 状态

这说明：

- `App.render()` 已经是帧边界控制器
- 图像层如果要进主线，必须服从它的帧边界

## 3. 第一阶段必须固定的接入策略

### 3.1 文本层仍然先完成

第一阶段的核心原则：

- 文本层永远先于图像层完成

原因：

- 当前布局、selection、debug、HitMap 都围绕文本层工作
- 图像层只是附加视觉层，不应反向影响这些既有流程

### 3.2 含 image layer 的实验帧先绕过 `AsyncRenderer`

这条在前面已经多次出现，但这里需要作为接入流程的硬规则再写一次。

原因：

- 当前 `AsyncRenderer` 只适配文本 `Buffer`
- 如果在第一阶段强行让它接图像层，会显著扩大改造面

因此 Phase 1 的策略应是：

- 无图像层：保持现有同步/异步文本路径
- 有图像层：强制同步文本输出 + 图像提交

### 3.3 `HitMap` 仍按文本帧更新

第一阶段 image layer 不承担复杂交互，因此：

- `HitMap` 继续只依赖文本/布局树结果

这意味着：

- 图像提交应在文本层完成之后
- 但不需要把 image region 立即并入 `HitMap`

## 4. 推荐的新增辅助方法

为了避免把 `render()` 继续写成更大的单体函数，建议拆出几个私有 helper。

### 4.1 `buildTextFrame()`

职责：

- 获取 back buffer
- reset
- 调用 `Paint()`
- 处理 selection / interaction / debug

返回：

- `*paint.Buffer`
- `component.PaintContext`
- `dirtyHints []paint.Rect`

### 4.2 `buildExperimentalScene()`

职责：

- 检查根节点是否实现 `ExperimentalSceneProvider`
- 如果实现，则构造 `SceneFrame`
- 如果未实现，则返回只含文本层的场景或 `nil`

### 4.3 `renderTextFrameSync()`

职责：

- 走同步 `renderer.Render()`
- 负责首帧 `ForceFullRender()`
- 负责 `fmt.Print(output)`

### 4.4 `renderTextFrameAsync()`

职责：

- 仅用于纯文本帧
- 调用 `asyncRenderer.SubmitFrame(...)`

### 4.5 `presentImageLayers()`

职责：

- 从 `SceneFrame.Images` 转换成 `DrawImageRequest`
- 调用 `GraphicsPresenter`
- 处理 replace / delete / clear 语义
- 记录 diagnostics

## 5. 推荐时序

### 5.1 纯文本帧

伪代码建议：

```go
func (a *App) render() {
    buf, ctx, dirtyHints := a.buildTextFrame()

    if a.asyncRenderer != nil {
        a.renderTextFrameAsync(buf, dirtyHints)
    } else {
        a.renderTextFrameSync(buf, dirtyHints)
    }

    a.updateHitMapAfterRender()
    a.finishRenderFrame()
}
```

### 5.2 含图像层实验帧

伪代码建议：

```go
func (a *App) render() {
    buf, ctx, dirtyHints := a.buildTextFrame()
    scene := a.buildExperimentalScene(ctx, buf)

    if scene == nil || !scene.HasImages() {
        // 走现有纯文本逻辑
        ...
        return
    }

    // Phase 1: 图像帧强制同步提交
    a.renderTextFrameSync(scene.Text, dirtyHints)
    a.presentImageLayers(scene)

    a.updateHitMapAfterRender()
    a.finishRenderFrame()
}
```

### 5.3 为什么是“先文本，再图像”

原因有四个：

- 首帧清屏和隐藏光标目前已经在文本路径里处理
- 文本层是图像层的布局基准
- 图像层暂不参与 selection / HitMap
- 图像提交失败时更容易无损回退到文本层

## 6. 首帧行为约束

### 6.1 首帧文本行为不变

第一阶段应继续沿用当前首帧逻辑：

- 清屏
- 隐藏光标
- 文本层全量渲染

### 6.2 首帧图像行为

建议：

- 首帧文本输出成功后，再提交图像层
- 如果图像层失败，当前帧至少还能看到文本 fallback

不要做：

- 图像先提交、文本后绘制

## 7. 清理与退出行为

### 7.1 `App.Close()` 必须触发图像清理

当前 `App.Close()` 已经负责：

- 停 `Pump`
- 停 `AsyncRenderer`
- 显示光标
- 清屏 / 退出

Phase 1 建议新增：

- 如果当前存在 `GraphicsPresenter`
- 在 `ShowCursor()` / `clearScreen()` 之前先 `Clear()`

### 7.2 `Resize()` 的建议行为

第一阶段建议：

- resize 时文本层仍按当前逻辑走
- 图像层统一视为失效
- 下一帧重新生成 image layer

不要在第一阶段做：

- 图像层局部 resize 复用

## 8. 失败处理策略

### 8.1 图像 presenter 不可用

建议行为：

- 记录 diagnostics
- 当前帧继续保留文本层
- 后续帧自动回退到文本模式

### 8.2 图像提交中途失败

建议行为：

- 不回滚已成功输出的文本层
- 标记 image mode 降级
- 下帧直接走纯文本

### 8.3 `ExperimentalSceneProvider` 构建失败

建议行为：

- 当作无图像层
- 继续正常走纯文本路径

## 9. 与 `AsyncRenderer` 的关系

### 9.1 第一阶段绝不修改 `AsyncRenderer` 协议

也就是说，Phase 1 不应：

- 给 `AsyncRenderer.SubmitFrame()` 增 image 参数
- 在 `PartialFrameBuffer` 里塞 image layer
- 在 `renderPending()` 里追加协议输出

### 9.2 Phase 2 才讨论的事情

等 Phase 1 跑通后，才考虑：

- image-aware async pipeline
- mixed frame throttling
- image dirty region coalescing

## 10. 推荐 diagnostics 输出点

第一阶段建议在以下位置记录 diagnostics：

### 10.1 `buildExperimentalScene()` 后

记录：

- 是否生成了 image layer
- image count
- graphics mode

### 10.2 `presentImageLayers()` 后

记录：

- 成功提交数
- 失败提交数
- presenter 返回的对象 ID

### 10.3 回退时

记录：

- 为什么从 image 回退到 text
- 是 capability 不可靠、scene 构建失败，还是 presenter 提交失败

## 11. 第一阶段验收标准

### 11.1 顺序正确性

- 文本层先完成
- 图像层后提交
- `HitMap` 仍稳定更新

### 11.2 回退正确性

- 图像失败不影响文本层
- 后续帧可自动回到 text path

### 11.3 对现有主链影响受控

- 纯文本页面仍然可走当前 `AsyncRenderer`
- 只有实验性含图像帧才走同步旁路

## 12. 一句话结论

`framework.App.render()` 第一阶段最稳的接入方式是：

**保持现有文本帧主链不动，只在文本帧完成后、命中测试更新前，给实验性场景增加一个同步的 image layer 提交通道。**
