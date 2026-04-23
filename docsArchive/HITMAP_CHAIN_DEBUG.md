# HitMap 处理链路调试

> **Date**: 2026-02-10
> **Status**: 🔍 调试中

---

## 问题

Modal 上的按钮点击位置错误，HitMap 记录的位置与实际显示位置不一致。

---

## 完整处理链路

### 1. 渲染流程

```
app.render()
  ↓
DeclarativeNode.Paint(ctx, buf)
  ↓
PipelineRendererAdapter.GetPipeline().Render(vnode, 0, 0, buf)
  ↓
PipelineRenderer.Render(vnode, x, y, buf)
  ↓ (检测到 layers)
RenderingPipeline.RenderLayers(vnode, constraints, buf)
```

### 2. Layer Layout 流程

```
RenderingPipeline.RenderLayers()
  ↓
layerMgr := layer.NewManager()
  ↓
layerMgr.CollectAndLayout(vnode, constraints, engine)
  ↓
Manager.CollectAndLayout()
  ├─ collector.Collect(vnode) - 收集 layer 节点
  ├─ m.layoutLayer() for each layer
  │   ↓
  │   engine.Layout(node.Content, layerConstraints)
  │     ├─ buildComputedBox() - 计算 layout
  │     ├─ calculatePositions() - 设置位置 (初始在 0,0)
  │     └─ buildHitMapFromComputedBoxes(root) - 构建 HitMap (❌ 此时位置还是 0,0)
  │   ↓
  │   centerModal(layout.Root, constraints) - 修改位置到中心 ✅
  │   ↓
  │   buildHitMapFromComputedBox(layout.Root) - 重新构建 HitMap (✅ 使用中心位置)
  ↓
layerMgr.GetMergedHitMap() - 合并所有 layers 的 HitMap
  ↓
RenderingPipeline.lastHitMap = mergedHitMap
```

### 3. HitMap 获取流程

```
app.render()
  ↓
paintable.Paint(ctx, buf)
  ↓
RenderingPipeline.lastHitMap (已保存)
  ↓
declNode.GetHitMap()
  ↓
pipeline.GetHitMap()
  ↓
返回 lastHitMap
  ↓
a.pump.SetHitMap(hitMap)
```

---

## 关键点

### ✅ 正确的流程

1. **Layout**: `engine.Layout()` 计算位置
2. **Transform**: `centerModal()` 修改位置
3. **Rebuild**: `buildHitMapFromComputedBox()` 使用修改后的位置

### ❌ 之前的问题

如果 HitMap 在步骤 1 构建，步骤 2 修改位置后没有重新构建，就会导致位置不一致。

---

## 调试方法

### 启用调试输出

```bash
TUI_DEBUG_HITMAP=true go run main.go 2>&1 | grep -E "HitMap|Entry|Bounds|pos="
```

### 期望看到的输出

```
[layoutLayer] Layer=2, Root pos=(50,10) size=20x5, HitMap entries=5
[buildHitMapFromComputedBox] Entry: ID=button-cancel, Bounds=(55,12,8x1)
[buildHitMapFromComputedBox] Entry: ID=button-ok, Bounds=(65,12,6x1)
[DeclarativeNode.GetHitMap] Returning HitMap with 15 entries
[APP] ✅ Got HitMap from RenderingPipeline: 15 entries (includes layer transforms)
```

### 检查点

1. **Modal Root 位置**: 应该在屏幕中心，不是 (0,0)
2. **Button Bounds**: 应该在 Modal Root 位置 + 偏移
3. **HitMap 大小**: 应该 > 0，包含所有按钮

---

## 当前实现

### `Manager.layoutLayer()` (✅ 正确)

```go
// 1. Layout
layout, err := engine.Layout(node.Content, layerConstraints)

// 2. Center
if layer == rtui.LayerModal && layout.Root != nil {
    m.centerModal(layout.Root, constraints)
}

// 3. Rebuild HitMap AFTER centering
if layout.Root != nil {
    layout.HitMap = m.buildHitMapFromComputedBox(layout.Root)
}
```

### `buildHitMapFromComputedBox()` (✅ 正确)

```go
func (m *Manager) buildHitMapFromComputedBox(root *compute.ComputedBox) *event.HitMap {
    var walk func(box *compute.ComputedBox, zOrder int)
    walk = func(box *compute.ComputedBox, zOrder int) {
        // ✅ 使用 box.Box.X/Y (已包含 centering 偏移)
        entry := event.HitMapEntryInternal{
            Bounds: layout.Rect{
                X:      box.Box.X,  // ✅ Final position AFTER centering
                Y:      box.Box.Y,
                Width:  box.Box.Width,
                Height: box.Box.Height,
            },
        }
        entries = append(entries, entry)
    }

    walk(root, 0)
    return event.BuildHitMapFromEntries(entries)
}
```

---

## 测试步骤

### 1. 编译

```bash
cd E:\projects\yao\wwsheng009\mint
go build ./runtime/layer ./internal/render ./framework
```

### 2. 运行 demo1

```bash
cd examples/ui_demos/demo1_full_featured
TUI_DEBUG_HITMAP=true go run main.go 2>&1 | grep -E "HitMap|Entry|Bounds|pos="
```

### 3. 测试点击

1. 点击 "[Open Modal]" 按钮
2. 查看 Modal 是否在中心位置
3. 查看 HitMap 条目的位置是否在中心
4. 点击 Modal 中的 "[ Cancel ]" 按钮
5. 检查是否响应

---

## 可能的问题

### 1. HitMap 仍然是旧位置

**原因**: `buildHitMapFromComputedBox()` 没有被调用

**检查**: 查看日志中是否有 `[buildHitMapFromComputedBox]` 输出

### 2. HitMap 被覆盖

**原因**: `app.render()` 中的回退逻辑重新构建了 HitMap

**检查**: 查看日志中是否有 `⚠️ HitMap built from layout.Node` 输出

### 3. Modal 没有被 centering

**原因**: `centerModal()` 没有被调用或没有修改位置

**检查**: 查看日志中 Modal Root 的 pos 是否为 (0,0)

---

## 下一步

1. 运行调试输出，查看实际的 HitMap 条目位置
2. 检查 `app.render()` 是否使用了正确的 HitMap
3. 验证 Pump 是否收到了正确的 HitMap

