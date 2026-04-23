# Steps 组件现状

> 更新时间：2026-03-17
> 状态：基础能力已可用，文档与回归已补齐

---

## 快速概览

| 维度 | 状态 | 备注 |
|------|------|------|
| 代码实现 | ✅ 可用 | `builder.go`、`vnode.go`、`instance.go`、`intent.go` 已齐 |
| 交互能力 | ✅ 可用 | 支持键盘导航、鼠标点击、受控/非受控 current、StepChangeIntent |
| 展示能力 | ✅ 可用 | 支持 horizontal / vertical、description、icon、status、progressDot、percent |
| 测试覆盖 | ✅ 基本覆盖 | `steps_test.go` 已覆盖核心 API；`ui/e2e/steps_e2e_test.go` 已补真实交互回归 |
| 用户文档 | ✅ 已补入口 | 当前 README 提供状态说明与使用入口 |

---

## 当前已验证能力

- 基础 Builder / VNode / Instance 结构完整
- `horizontal` / `vertical` 两种方向渲染可用
- `current` / `initialCurrent` 的受控与非受控模式可用
- `description` / `icon` / `status` / `progressDot` / `percent` 已接通
- 键盘导航支持 `left/right/up/down/home/end`
- 鼠标点击可切换当前 step
- 会发出 `StepChangeIntent`
- `CurrentForField(...)` 支持发出 `FieldChangeIntent`
- `ui/e2e` 已覆盖键盘导航与 vertical click 的真实交互链路

---

## 主要文件

| 文件 | 说明 |
|------|------|
| `builder.go` | Fluent API |
| `vnode.go` | 不可变描述与 item model |
| `instance.go` | Measure / Paint / HandleAction / HandleIntent |
| `intent.go` | `StepChangeIntent` 定义 |
| `steps_test.go` | 单元测试 |

---

## 快速示例

```go
ui.NewStepsBuilder().
    SetID("checkout-steps").
    ComponentID("checkout.steps").
    Items([]ui.StepsItem{
        ui.NewStepsItem("Cart"),
        ui.NewStepsItem("Address").WithDescription("fill shipping info"),
        ui.NewStepsItem("Pay"),
    }).
    Current(1).
    Vertical().
    Build()
```

---

## 相关测试

- `ui/components/steps/steps_test.go`
- `ui/e2e/steps_e2e_test.go`
