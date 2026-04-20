# Charts LineChart Image Prototype

这个 example 现在保留为一个“收口后的状态页”，不是可用的 chart 像素渲染示例。

当前策略：

- `linechart` 组件继续走稳定的文本渲染路径
- chart 侧的 `ImagePlotBackend()` API 仅为兼容保留，运行时会规范化回文本 backend
- 终端图形能力仍然保留在基础 graphics / image-layer 机制里，但它们只应服务于专用图片控件，不再作为 charts 的默认接入方式

保留这个 example 的原因：

- 让现有 prototype / e2e / diagnostics 入口继续存在，不破坏既有开发工具链
- 明确展示“chart 像素后端已暂停”的当前产品策略
- 为后续独立图片控件保留一个可复用的图形诊断页骨架

运行：

```powershell
go run ./examples/charts_linechart_image_prototype
```

你会看到：

- 左右两侧都显示文本 `linechart`
- 页面 banner 明确提示 `Chart image backend paused`
- `Diagnostics` 仍会显示当前终端 graphics probe 结果，但这些结果不会再驱动 chart 切到像素绘制

如果后续要恢复 chart 像素渲染，不应直接依赖这个 example 的当前行为，而应先重新评估终端协议稳定性，再决定是否重新开放 chart 层接入。
