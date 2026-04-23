# Group A 图形能力层技术规格

## 1. 文档目的

`PHASE1_TASK_BREAKDOWN.md` 已经把第一阶段拆成多个任务组，但 Group A 仍然需要一份真正可编码的规格文档。

本文件只回答一件事：

> 在当前系统里，图形能力探测层应该长什么样，才能支撑后续 image prototype，又不破坏现有文本主链。

这里讨论的是：

- 图形能力探测
- cell/pixel 尺寸元数据
- 图像输出能力的最小抽象入口

本文件明确不讨论：

- Scene 结构
- `App.render()` 如何提交图像层
- `linechart` 的具体 image renderer

这些内容留给 Group B / Group C。

## 2. 当前系统约束

在设计 Group A 之前，必须先承认当前平台层的现实边界。

### 2.1 `runtime/platform.Terminal` 目前只有文本终端能力

见：

- `runtime/platform/terminal.go`

当前 `Terminal` 只暴露：

- `Init / Close`
- `EnterAlternateScreen / ExitAlternateScreen`
- `EnableRawMode / DisableRawMode`
- `ShowCursor / HideCursor / MoveCursor`
- `Write / WriteString`
- `GetSize`

这意味着：

- 现有 `Terminal` 并不知道图形协议
- 现有 `Terminal` 也没有像素尺寸接口

### 2.2 `runtime/platform.Screen` 也是文本输出抽象

见：

- `runtime/platform/screen.go`

当前 `Screen` 只暴露：

- `Write`
- `Flush`
- `Clear`
- `AltScreen`
- `Size`

它适合作为文本帧输出抽象，但还不是图像输出抽象。

### 2.3 `framework.App` 当前并不直接依赖 `Terminal` / `Screen` 完成输出

见：

- `framework/app.go`
- `runtime/paint/renderer.go`
- `runtime/paint/async_renderer.go`

当前主线是：

1. 组件树把内容画进 `paint.Buffer`
2. `paint.Renderer` 生成 ANSI 文本输出
3. `App.render()` 直接 `fmt.Print(...)`
4. 如启用 `AsyncRenderer`，则由异步渲染器统一调度文本输出

这个事实非常重要：

- Group A 不能假设图像能力一定会挂在 `App` 的现有输出路径上
- Group A 必须先提供“能力与后端”抽象，再由 Group C 决定怎么接入主渲染链

### 2.4 当前只有颜色能力探测，没有图形能力探测

见：

- `framework/theme/output.go`

当前已有的是：

- `ColorModeTrueColor`
- `ColorMode256`
- `ColorMode16`
- `ColorModeNone`

这说明：

- “终端能力探测”这个概念在仓库里是存在的
- 但图形协议能力需要独立一套模型，不能硬塞进 `ColorMode`

## 3. Group A 的设计目标

Group A 的目标必须足够克制。

### 3.1 必须做到

- 能明确表达“当前终端不支持图像协议”
- 能表达“当前终端支持哪一种图像协议”
- 能表达 cell 对应像素尺寸是否可靠
- 能为后续图像输出层提供最小 capability 元数据
- 不破坏现有 `Terminal` / `Screen` / `InputReader` 代码路径

### 3.2 明确不做

- 不在 Group A 里接入 `App.render()`
- 不在 Group A 里实现完整图像提交流程
- 不在 Group A 里引入复杂多协议适配矩阵
- 不在 Group A 里决定 charts 组件 API

## 4. 设计原则

### 4.1 能力探测与协议输出解耦

不要把“探测到什么能力”和“如何发送图片协议”耦合在一起。

Group A 应分成两层：

- 能力探测结果
- 可选的图像输出能力提供者

这样即使后续协议输出实现发生变化，能力模型仍然稳定。

### 4.2 以可选接口扩展，而不是修改现有核心接口

第一阶段不建议直接修改：

- `platform.Terminal`
- `platform.Screen`

为强制实现图形方法。

原因：

- 这会立即破坏现有默认实现
- 也会把未实现图像能力的平台全部拖进改造范围

更稳的方式是：

- 保留现有接口
- 新增可选能力接口

### 4.3 Group A 只保证“安全识别”，不保证“全协议完备”

第一阶段能力探测的目标不是做终端兼容大全，而是：

- 在明确支持的场景下稳定识别
- 在不确定场景下保守回退到 `None`

如果探测结果不可靠，宁可退到文本路径，也不要误判进入 image mode。

### 4.4 能力结果应可缓存、可覆盖、可诊断

后续 PoC 和 benchmark 会频繁依赖 capability 结果，因此它必须：

- 可缓存
- 可通过环境变量强制覆盖
- 可打印 diagnostics

## 5. 建议的数据模型

### 5.1 `GraphicsMode`

第一阶段建议保守定义：

```go
type GraphicsMode int

const (
    GraphicsModeNone GraphicsMode = iota
    GraphicsModeKitty
    GraphicsModeSixel
    GraphicsModeInlineImage
)
```

但 Phase 1 实际要求只需要保证：

- `GraphicsModeNone`
- `GraphicsModeKitty`

其余值可以先保留枚举位，不要求在第一阶段完全实现。

### 5.2 `GraphicsCapabilities`

建议最小结构如下：

```go
type GraphicsCapabilities struct {
    Mode               GraphicsMode
    Reliable           bool
    CellPixelWidth     int
    CellPixelHeight    int
    SupportsPlacement  bool
    SupportsReplace    bool
    SupportsDelete     bool
    SupportsCrop       bool
    SupportsZOrder     bool
    ProbeSource        string
    Notes              []string
}
```

字段说明：

- `Mode`
  - 当前识别出的图形协议模式
- `Reliable`
  - 当前探测结果是否可安全用于启用 image mode
- `CellPixelWidth / CellPixelHeight`
  - 单个字符格对应像素尺寸；未知时允许为 `0`
- `SupportsPlacement`
  - 是否支持指定图像放置位置
- `SupportsReplace`
  - 是否支持图像对象更新/替换
- `SupportsDelete`
  - 是否支持图像对象删除
- `SupportsCrop`
  - 是否支持区域裁剪
- `SupportsZOrder`
  - 是否有可用的图层/覆盖语义
- `ProbeSource`
  - 本次结果来自什么来源，例如 `env-override`、`heuristic-term`
- `Notes`
  - 用于 diagnostics 的附加说明

### 5.3 `CellPixelMetrics` 是否单独拆结构

第一阶段不必单独拆成结构体。

原因：

- 目前只需要最小信息
- Phase 1 更重要的是能力判定是否可靠，而不是做完整显示器度量模型

如果后续需要记录：

- logical vs physical pixel
- DPI
- terminal scaling

再单独拆出 `CellPixelMetrics` 会更合适。

## 6. 建议的接口分层

### 6.1 能力读取接口

建议新增一个可选接口：

```go
type GraphicsCapabilityProvider interface {
    GraphicsCapabilities() GraphicsCapabilities
}
```

用途：

- 给 `App`
- 给 prototype
- 给 benchmark
- 给 diagnostics

统一读取能力结果。

### 6.2 图像输出接口

第一阶段建议不要把图像输出直接塞进 `Terminal` 或 `Screen`。

建议单独定义：

```go
type GraphicsPresenter interface {
    Capabilities() GraphicsCapabilities
    Present(req DrawImageRequest) (string, error)
    Replace(id string, req DrawImageRequest) error
    Delete(id string) error
    Clear() error
}
```

其中：

- 返回值 `string` 可作为 image handle / object id
- 这层只负责“图像对象生命周期”
- 不负责决定 `App.render()` 的时序

### 6.3 提供者接口

为了不污染现有 `Terminal` / `Screen` 主接口，建议再加一层提供者：

```go
type GraphicsPresenterProvider interface {
    GraphicsPresenter() GraphicsPresenter
}
```

这样 Phase 1 可以做到：

- 默认平台实现什么都不实现
- 实验性平台后端按需实现

## 7. 建议的文件拆分

建议第一批按下面方式放置。

### 7.1 核心类型

- `runtime/platform/graphics.go`

建议放：

- `GraphicsMode`
- `GraphicsCapabilities`
- `GraphicsCapabilityProvider`
- `GraphicsPresenter`
- `GraphicsPresenterProvider`

### 7.2 探测逻辑

- `runtime/platform/graphics_probe.go`

建议放：

- `ProbeGraphicsCapabilities()`
- 协议识别逻辑
- 环境变量覆盖逻辑

### 7.3 环境变量解析

- `runtime/platform/graphics_env.go`

建议放：

- `MINT_GRAPHICS`
- `MINT_CELL_PIXELS`
- 其他 Phase 1 override 解析

这样做的好处是：

- 类型与探测实现解耦
- 后续协议扩展不需要一直改一个大文件

## 8. 建议的探测策略

### 8.1 第一优先级：显式环境变量覆盖

建议支持：

- `MINT_GRAPHICS=off`
- `MINT_GRAPHICS=kitty`
- `MINT_GRAPHICS=auto`

建议支持：

- `MINT_CELL_PIXELS=8x16`

原因：

- PoC 阶段最需要的是可控性
- 不能完全依赖运行环境自动探测

### 8.2 第二优先级：保守启发式探测

如果没有显式覆盖，再走自动探测。

Phase 1 建议只做保守判断，例如：

- 已知 Kitty 环境才返回 `GraphicsModeKitty`
- 其余全部先回 `GraphicsModeNone`

不要在第一阶段做激进猜测。

### 8.3 像素尺寸未知时的策略

如果协议存在，但 cell 像素尺寸未知，建议：

- `Mode` 可以保留实际协议值
- 但 `Reliable=false`
- 默认不自动进入 image mode

这样后续 `Auto` 策略能更稳地回退。

## 9. 和现有代码的接入方式

### 9.1 不修改 `InputReader`

`runtime/platform/input.go` 当前职责是：

- 产生 `RawInput`
- 控制鼠标捕获

它不应该承担图像能力探测。

### 9.2 不强制修改 `DefaultTerminal` / `DefaultScreen`

默认实现可以继续只实现文本能力。

如果需要实验性图像输出后端，可以新增新的具体类型，而不是强改默认实现。

### 9.3 `framework.App` 只依赖“可选能力”

Group C 真正接入 `App.render()` 时，应该只做：

- 判断某个对象是否实现 `GraphicsCapabilityProvider`
- 判断某个对象是否实现 `GraphicsPresenterProvider`

而不是要求 `App` 现在就直接依赖具体协议实现。

## 10. Phase 1 明确决策

为避免后续方案漂移，建议现在就固定这几个决策。

### 10.1 决策一：Phase 1 只要求 `Kitty + None`

理由：

- `Kitty` 作为第一条实验路径更容易得到稳定能力与放置语义
- 同时做 `Sixel` 会把 Group A 复杂度放大

### 10.2 决策二：探测错误时一律保守回退

理由：

- 误判进入 image mode 的代价远高于误判退回文本

### 10.3 决策三：能力结果在一次应用生命周期内默认缓存

理由：

- 运行期频繁重新探测意义不大
- 也会增加不必要的不确定性

如需重新探测，允许后续新增显式刷新接口，但不是 Phase 1 必要项。

## 11. 验收标准

### 11.1 功能验收

- 能明确返回 `GraphicsModeNone`
- 在实验性 Kitty 环境下能返回 `GraphicsModeKitty`
- 能识别显式 override
- 能在 diagnostics 中打印能力结果

### 11.2 稳定性验收

- 未实现图像能力的平台不需要改代码即可继续工作
- capability 探测失败不会影响文本主链
- 不会因 capability 探测引入额外终端状态污染

### 11.3 工程验收

- Group A 的新接口不要求现有所有平台实现立即跟进
- 不修改 charts 组件代码也能编译通过

## 12. 推荐测试

建议至少写以下测试。

### 12.1 单元测试

- 环境变量覆盖优先级
- `auto` 在未知环境下回退到 `None`
- `MINT_CELL_PIXELS` 解析成功/失败

### 12.2 诊断测试

- capability dump 输出字段完整

### 12.3 回退测试

- 无图形能力时不影响现有文本 app 启动

## 13. 非目标与风险

### 13.1 非目标

- 不在 Group A 定义完整图像协议编码格式
- 不在 Group A 设计最终 Scene 模型
- 不在 Group A 处理 hit testing

### 13.2 风险

- 如果 `GraphicsCapabilities` 字段定义过重，会把 Group A 拖成一个大重构
- 如果定义过轻，又会让 Group B / Group C 返工

因此本规格的核心原则是：

**只定义 Phase 1 真正会消费的能力字段。**

## 14. 一句话结论

Group A 的正确目标不是“把图像画出来”，而是：

**在不破坏当前文本平台抽象的前提下，建立一套可靠、可回退、可诊断的图形能力层。**

## 15. 下一步实现参考

如果要从文档直接进入编码，建议紧接着参考：

- [RUNTIME_PLATFORM_GRAPHICS_API_SKETCH.md](./RUNTIME_PLATFORM_GRAPHICS_API_SKETCH.md)

它进一步把 Group A 收敛成：

- 推荐文件拆分
- 第一版接口签名
- 环境变量设计
- 探测伪代码
- 测试清单
