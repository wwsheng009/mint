# Popconfirm

`popconfirm/` 提供带确认/取消操作的气泡确认框，适合删除、重置和其他危险操作前的二次确认。

支持：

- `title` + `description`
- `click` / `hover` / `manual` 触发
- `PlacementTop` 默认定位，也支持 `PlacementAuto` + top / bottom 系列 6 个 placement；`auto` 会优先使用顶部空间，不足时回退到底部
- `auto` 与显式 placement 会共享同一套 fallback / clamp 逻辑，顶边和左右边界下不再保留单独的 portal 偏移分支
- `OK` / `Cancel` 文案自定义
- confirm / cancel 按钮 variant 与 footer layout
- `Install(app)` 后统一收口 ESC / outside-click 关闭
- 本地 open/toggle intents
- 全局 `PopconfirmConfirmIntent` / `PopconfirmCancelIntent`
- overlay 坐标换算已复用 `ui/components/internal/overlayposition`

## 状态语义

- `InitialOpen(...)` 用于非受控初始打开状态，`Open(...)` 用于显式受控打开状态。
- 建议显式设置 `ComponentID(...)`，这样本地 `ToggleWithID(...)` / `OpenWithID(...)` / `CloseWithID(...)` 与全局 confirm / cancel 事件才能稳定绑定到同一实例。
- `PopconfirmConfirmIntent` 与 `PopconfirmCancelIntent` 是全局意图，适合把真正的删除 / 重置动作放到业务层统一处理。

## 示例

```go
ui.NewPopconfirmBuilder(
    ui.NewButtonBuilder("Delete").Danger().Build(),
).
    ComponentID("delete.confirm").
    Title("Delete this record?").
    Description("This action cannot be undone.").
    OkVariant(button.VariantDanger).
    FooterLayout(popconfirm.FooterLayoutCenter).
    Build()
```

如果希望 ESC 和 outside-click 能关闭已打开的 Popconfirm，需要在应用启动时调用一次 `popconfirm.Install(app)`。

## 测试入口

- 单测：`go test ./ui/components/popconfirm`
- 重点覆盖：`popconfirm_test.go` 中的 placement、footer layout、confirm / cancel intent 与 middleware 关闭链路；显式 `top-left` / `top-right` 在顶角场景下回退到下方同侧 family、`bottom-left` / `bottom-right` 在底角场景下回退到上方同侧 family 的 corner 几何，E2E 也已补对应专用 fixture；极窄 viewport 下显式 `top-left` / `top-right` / `bottom-left` / `bottom-right` 仍保持各自 vertical family，目前组件单测除了 left-edge clamp，也覆盖了“没有任何 vertical candidate 能完整放入时”的双轴 clamp 与 resolved placement
- E2E：`go test ./ui/e2e -run TestE2EPopconfirm`
