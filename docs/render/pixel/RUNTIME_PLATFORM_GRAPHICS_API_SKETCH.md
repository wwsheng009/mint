# `runtime/platform/graphics` API 草案与伪代码清单

## 1. 文档目的

前面的文档已经明确了：

- 为什么 Group A 必须先落地
- `GraphicsMode / GraphicsCapabilities` 需要哪些字段
- 为什么第一阶段要用可选接口而不是强改 `Terminal / Screen`

但如果要真正开工，仍然缺一层更贴近代码的东西：

> `runtime/platform/graphics.go` 这一组文件，第一版到底应该长什么样。

本文件的目标就是把 Group A 再向前推进一层，给出：

- 建议文件拆分
- 推荐接口签名
- 环境变量覆盖方式
- capability 探测伪代码
- 第一批测试清单

## 2. 当前平台层的现实边界

对照当前实现：

- [terminal.go](/E:/projects/yao/wwsheng009/mint/runtime/platform/terminal.go)
- [screen.go](/E:/projects/yao/wwsheng009/mint/runtime/platform/screen.go)

当前平台层有两个很明确的约束：

1. `Terminal` / `Screen` 都是文本终端抽象
2. 默认实现没有图像协议、像素尺寸、图像对象生命周期管理

因此第一阶段最稳的方式不是重写现有接口，而是：

- 继续保留现有 `Terminal / Screen`
- 在 `runtime/platform` 下新增一组并列的 graphics 能力文件

## 3. 建议文件拆分

第一阶段建议最少拆成 4 个文件。

### 3.1 `runtime/platform/graphics.go`

职责：

- 定义核心类型与接口

建议放：

- `GraphicsMode`
- `GraphicsCapabilities`
- `DrawImageRequest`
- `GraphicsCapabilityProvider`
- `GraphicsPresenter`
- `GraphicsPresenterProvider`

### 3.2 `runtime/platform/graphics_env.go`

职责：

- 解析与 graphics 相关的环境变量覆盖

建议放：

- `MINT_GRAPHICS`
- `MINT_CELL_PIXELS`
- `MINT_GRAPHICS_STRICT`

### 3.3 `runtime/platform/graphics_probe.go`

职责：

- 负责 capability 探测
- 组合环境覆盖与保守启发式

建议放：

- `ProbeGraphicsCapabilities()`
- `probeGraphicsFromEnv()`
- `probeGraphicsHeuristics()`

### 3.4 `runtime/platform/graphics_kitty.go`

职责：

- 第一阶段实验性 Kitty presenter

建议放：

- `KittyGraphicsPresenter`
- `Present / Replace / Delete / Clear`

注意：

- Phase 1 不要求 `Sixel` 落地
- 先不要创建 `graphics_sixel.go`

## 4. 第一版类型草案

### 4.1 `GraphicsMode`

```go
type GraphicsMode int

const (
    GraphicsModeNone GraphicsMode = iota
    GraphicsModeKitty
    GraphicsModeSixel
    GraphicsModeInlineImage
)
```

说明：

- 第一阶段只要求实际支持 `None` 和 `Kitty`
- 其他模式先保留枚举位，避免后续反复改类型

### 4.2 `GraphicsCapabilities`

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

说明：

- `Reliable` 非常关键，它决定 `Auto` 是否可以安全进入 image mode
- `CellPixelWidth / CellPixelHeight` 未知时允许为 `0`
- `Notes` 必须保留，方便 prototype diagnostics

### 4.3 `DrawImageRequest`

第一阶段建议最小结构如下：

```go
type DrawImageRequest struct {
    ID              string
    PixelWidth      int
    PixelHeight     int
    CellX           int
    CellY           int
    CellWidth       int
    CellHeight      int
    RGBA            []byte
    AltText         string
    ReplaceIfExists bool
}
```

字段解释：

- `ID`
  - 图像对象 ID，Phase 1 允许由上层显式传入
- `PixelWidth / PixelHeight`
  - raster surface 尺寸
- `CellX / CellY / CellWidth / CellHeight`
  - 图像在 cell 网格中的逻辑位置
- `RGBA`
  - 第一阶段直接用裸像素载荷即可
- `AltText`
  - 调试、日志、fallback 说明
- `ReplaceIfExists`
  - 允许 presenter 在同 ID 下直接替换

### 4.4 能力读取接口

```go
type GraphicsCapabilityProvider interface {
    GraphicsCapabilities() GraphicsCapabilities
}
```

### 4.5 图像提交接口

```go
type GraphicsPresenter interface {
    Capabilities() GraphicsCapabilities
    Present(req DrawImageRequest) (string, error)
    Replace(id string, req DrawImageRequest) error
    Delete(id string) error
    Clear() error
}
```

### 4.6 Presenter 提供者接口

```go
type GraphicsPresenterProvider interface {
    GraphicsPresenter() GraphicsPresenter
}
```

## 5. 为什么不用修改现有 `Terminal` / `Screen`

这是第一阶段必须明确写死的决策。

### 5.1 如果直接改 `Terminal`

例如加：

```go
GraphicsCapabilities() GraphicsCapabilities
DrawImage(req DrawImageRequest) error
```

问题是：

- 当前所有 `Terminal` 实现都会被迫改动
- 默认文本终端实现也要补空实现
- 第一阶段范围会被不必要地放大

### 5.2 如果改 `Screen`

问题类似：

- `Screen` 当前是纯文本帧输出抽象
- 把图像对象生命周期塞进去，会让 `Screen` 的语义变重

### 5.3 更稳的方式

第一阶段推荐：

- `Terminal` / `Screen` 保持原样
- 由新的实验性对象单独实现 `GraphicsCapabilityProvider`
- 如有需要，再额外实现 `GraphicsPresenterProvider`

## 6. 建议的环境变量约定

第一阶段为了可控性，建议至少支持 3 个环境变量。

### 6.1 `MINT_GRAPHICS`

建议支持值：

- `auto`
- `off`
- `kitty`

行为建议：

- `off`
  - 强制 `GraphicsModeNone`
- `kitty`
  - 强制 `GraphicsModeKitty`
- `auto`
  - 走启发式探测

### 6.2 `MINT_CELL_PIXELS`

建议格式：

- `8x16`
- `10x20`

用途：

- 覆盖 cell 对应像素尺寸
- 便于 prototype 和 benchmark 在固定环境下运行

### 6.3 `MINT_GRAPHICS_STRICT`

建议支持：

- `0`
- `1`

建议行为：

- `1`
  - 如果能力不可靠，则强制回 `GraphicsModeNone`
- `0`
  - 允许实验性能力继续保留，但 `Reliable=false`

## 7. capability 探测伪代码

### 7.1 总入口

```go
func ProbeGraphicsCapabilities() GraphicsCapabilities {
    if caps, ok := probeGraphicsFromEnv(); ok {
        return caps
    }

    caps := probeGraphicsHeuristics()
    if strictGraphicsModeEnabled() && !caps.Reliable {
        return GraphicsCapabilities{
            Mode:        GraphicsModeNone,
            Reliable:    true,
            ProbeSource: "strict-fallback",
            Notes:       []string{"graphics probe not reliable, fallback to none"},
        }
    }
    return caps
}
```

### 7.2 环境变量覆盖

```go
func probeGraphicsFromEnv() (GraphicsCapabilities, bool) {
    mode := strings.TrimSpace(strings.ToLower(os.Getenv("MINT_GRAPHICS")))
    if mode == "" {
        return GraphicsCapabilities{}, false
    }

    switch mode {
    case "off":
        return GraphicsCapabilities{
            Mode:        GraphicsModeNone,
            Reliable:    true,
            ProbeSource: "env-override",
        }, true
    case "kitty":
        w, h, notes := parseCellPixelsFromEnv()
        return GraphicsCapabilities{
            Mode:              GraphicsModeKitty,
            Reliable:          w > 0 && h > 0,
            CellPixelWidth:    w,
            CellPixelHeight:   h,
            SupportsPlacement: true,
            SupportsReplace:   true,
            SupportsDelete:    true,
            ProbeSource:       "env-override",
            Notes:             notes,
        }, true
    case "auto":
        return GraphicsCapabilities{}, false
    default:
        return GraphicsCapabilities{
            Mode:        GraphicsModeNone,
            Reliable:    true,
            ProbeSource: "env-invalid-fallback",
            Notes:       []string{"invalid MINT_GRAPHICS value, fallback to none"},
        }, true
    }
}
```

### 7.3 启发式探测

第一阶段建议极度保守：

```go
func probeGraphicsHeuristics() GraphicsCapabilities {
    term := os.Getenv("TERM")
    program := os.Getenv("TERM_PROGRAM")

    if looksLikeKitty(term, program) {
        return GraphicsCapabilities{
            Mode:              GraphicsModeKitty,
            Reliable:          false,
            SupportsPlacement: true,
            SupportsReplace:   true,
            SupportsDelete:    true,
            ProbeSource:       "heuristic-kitty",
            Notes:             []string{"kitty detected heuristically; cell pixel size unknown"},
        }
    }

    return GraphicsCapabilities{
        Mode:        GraphicsModeNone,
        Reliable:    true,
        ProbeSource: "heuristic-none",
    }
}
```

关键点：

- 只在“明显像 Kitty”时才给 `Kitty`
- 像素尺寸未知时不标记为可靠

## 8. Presenter 的最小行为约束

### 8.1 `Present`

职责：

- 创建或提交一个图像对象
- 返回最终对象 ID

第一阶段不要求：

- 复杂复用策略
- 分块上传

### 8.2 `Replace`

职责：

- 用新内容替换同一对象 ID

如果底层协议不支持真正 replace，也允许：

- `Delete + Present`

但这一点应该封装在 presenter 内部。

### 8.3 `Delete`

职责：

- 删除单个图像对象

### 8.4 `Clear`

职责：

- 清掉 presenter 管理范围内的所有图像对象
- 用于 alternate screen 退出、prototype 页面关闭

## 9. 第一版 Kitty presenter 的边界

第一阶段 Kitty presenter 只建议做这些事：

- 接收 RGBA bitmap
- 接收 cell placement 信息
- 输出协议串
- 管理对象 ID

明确不做：

- 分块图像缓存系统
- 多图层高级 z-order 合成
- 图像区域差分
- 复杂错误恢复

## 10. 推荐测试拆分

### 10.1 `graphics_env_test.go`

覆盖：

- `MINT_GRAPHICS=off`
- `MINT_GRAPHICS=kitty`
- `MINT_GRAPHICS=auto`
- `MINT_CELL_PIXELS` 成功/失败解析

### 10.2 `graphics_probe_test.go`

覆盖：

- 无环境变量时默认回 `None`
- 启发式 Kitty 识别
- strict 模式下不可靠能力回退

### 10.3 `graphics_caps_test.go`

覆盖：

- `GraphicsCapabilities` 的序列化/字符串摘要
- diagnostics 输出稳定性

### 10.4 `graphics_kitty_test.go`

第一阶段只需要最小覆盖：

- presenter 返回对象 ID
- `Replace / Delete / Clear` 的对象生命周期语义

不要求：

- 完整协议 golden

## 11. 推荐的实现顺序

建议顺序如下：

1. `graphics.go`
2. `graphics_env.go`
3. `graphics_probe.go`
4. `graphics_env_test.go / graphics_probe_test.go`
5. `graphics_kitty.go`
6. presenter 最小测试

不要一开始就写 Kitty presenter。

原因：

- 先把类型和探测稳定下来
- 后续 Scene / App.render 接入才不会返工

## 12. 与后续 Group B/C 的边界

Group A 完成后，Group B/C 才能安全依赖：

- `GraphicsCapabilities`
- `GraphicsPresenter`
- `DrawImageRequest`

但 Group A 不应该提前决定：

- `SceneFrame` 长什么样
- `App.render()` 在同步还是异步链上提交图像
- image layer 如何参与 `HitMap`

## 13. 一句话结论

这份 API 草案的核心价值是：

**把 Group A 从“知道要有一个能力层”推进到“知道第一批文件、接口、环境变量和探测逻辑具体该怎么写”。**
