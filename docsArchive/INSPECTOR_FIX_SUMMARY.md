# Inspector 点击问题修复方案 - 快速摘要

> **版本**: v1.0
> **日期**: 2026-02-10
> **详细文档**: [INSPECTOR_HITMAP_INTEGRATION_ANALYSIS.md](./INSPECTOR_HITMAP_INTEGRATION_ANALYSIS.md)

---

## 🎯 问题核心

**现象**：Inspector 的 TreeView 无法正确响应鼠标点击，点击位置与选中节点不匹配。

**根本原因**：
1. **接口不兼容**：VNode、layout.Node、component.Node 三套系统的 Children() 返回类型不同
2. **Inspector 被排除**：Inspector 通过 Fragment 注入，不在 layout.Node 层级，导致 HitMap 不包含它
3. **手动坐标解析**：违反了"组件不再手写命中"的设计原则

---

## 💡 解决方案

**核心思路**：直接从 VNode 树构建 HitMap，利用布局引擎已设置的 bounds 字段。

```
┌─────────────────────────────────────────────────────┐
│                 正确的事件流程                        │
├─────────────────────────────────────────────────────┤
│                                                      │
│  用户点击 (x, y)                                      │
│      ↓                                               │
│  Pump.HitTest(x, y)                                  │
│      ↓                                               │
│  HitMap 包含 Inspector overlay！ ✅                  │
│      ↓                                               │
│  返回: TargetID = "inspector-treeview-123"           │
│       LocalX = 5, LocalY = 8                         │
│      ↓                                               │
│  Router.Dispatch(ev)                                 │
│      ↓                                               │
│  TreeView.HandleEvent(ev)                            │
│      ↓                                               │
│  TreeView 使用 ev.LocalY=8，选中第8行 ✅              │
│                                                      │
└─────────────────────────────────────────────────────┘
```

---

## 📝 实施步骤 (5-6天)

### Phase 1: 实现 VNode HitMap 构建 (2-3天)

1. **创建 `runtime/event/hitmap_vnode.go`**
   - `BuildHitMapFromVNode()` - 从 VNode 树构建 HitMap
   - 直接提取 ElementVNode.bounds 字段
   - 递归遍历 Children()

2. **集成到 App.render()**
   - 替换现有的 VNodeAdapter 逻辑
   - 从 VNode 树直接构建 HitMap

3. **编写单元测试**
   - 测试 ElementVNode、Fragment、ComponentVNode
   - 测试深层嵌套

### Phase 2: 清理手动事件处理 (1-2天)

1. **移除 Inspector 的手动坐标解析** (`standalone_inspector.go:1715-1757`)
2. **移除 Panel 的事件转发** (`panel.go:27-45`)
3. **简化 Tabs 的事件处理** (`tabs.go:198-223`)

### Phase 3: 验证和测试 (1天)

1. 集成测试：打开 Inspector，点击 TreeView，验证选中正确
2. 性能测试：HitMap 构建 < 1ms，HitTest < 0.1ms
3. 调试工具：`TUI_DEBUG_HITMAP=true` 可视化

---

## ✅ 预期收益

| 指标 | 当前 | 目标 | 提升 |
|------|------|------|------|
| Inspector 点击成功率 | ~30% | 100% | +233% |
| 点击位置准确性 | 经常错误 | 完全准确 | ∞ |
| 手动坐标解析代码 | ~100 行 | 0 行 | -100% |

---

## 🔧 关键代码示例

### VNode HitMap 构建

```go
// runtime/event/hitmap_vnode.go
func BuildHitMapFromVNode(root VNode, offsetX, offsetY int) *HitMap {
    hitMap := NewHitMap()
    buildHitMapRecursive(root, offsetX, offsetY, hitMap)
    return hitMap
}

func buildHitMapRecursive(node VNode, offsetX, offsetY int, hitMap *HitMap) {
    // 提取 bounds（布局引擎已设置）
    if boundsGetter, ok := node.(interface{ GetBounds() [4]int }); ok {
        bounds := boundsGetter.GetBounds()
        x, y, w, h := bounds[0], bounds[1], bounds[2], bounds[3]

        // 添加到 HitMap
        entry := &HitMapEntry{
            NodeID: getNodeID(node),
            Bounds: Rect{X: x, Y: y, Width: w, Height: h},
            LocalXY: func(globalX, globalY int) (int, int) {
                return globalX - x, globalY - y
            },
        }
        hitMap.Add(entry)
    }

    // 递归子节点
    for _, child := range node.Children() {
        buildHitMapRecursive(child, offsetX, offsetY, hitMap)
    }
}
```

### Inspector 简化后

```go
// ✅ 正确做法：直接使用 Pump 填充的坐标
func (si *StandaloneInspector) HandleEvent(ev event.Event) bool {
    me, ok := ev.(*frameworkevent.MouseEvent)
    if !ok {
        return false
    }

    // Pump 已经设置了 TargetID 和 LocalX/LocalY
    // Inspector 直接使用这些信息，无需手动计算！
    fmt.Printf("Click on %s at (%d, %d)\n", me.TargetID, me.LocalX, me.LocalY)

    // 分发到目标组件
    // ...
}
```

---

## ⚠️ 关键风险

1. **ComponentVNode 可能没有 bounds** → 检查接口实现，跳过无 bounds 节点
2. **偏移量计算错误** → 使用绝对坐标，不累积偏移
3. **循环导入** → 定义最小接口在 runtime/event，避免导入 runtime/ui

---

## 📚 相关文档

- **详细分析**：[INSPECTOR_HITMAP_INTEGRATION_ANALYSIS.md](./INSPECTOR_HITMAP_INTEGRATION_ANALYSIS.md)
- **长期架构**：[long_term_event_architecture.md](../../event/long_term_event_architecture.md)
- **Msg 设计**：[MSG_UNIFICATION.md](./MSG_UNIFICATION.md)

---

**状态**: 待审核实施
**工期**: 5-6 天
**优先级**: 高 (解决当前功能缺陷)
