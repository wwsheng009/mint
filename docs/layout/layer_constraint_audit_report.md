# 布局系统与Layer系统约束功能审查报告

**审查日期**: 2025-02-07
**审查范围**: Layer系统与布局引擎之间的约束传递和交互
**问题级别**: 🔴 严重 - Modal宽度约束失效

---

## 一、执行摘要

经过全面审查，发现了一个**关键性Bug**导致Modal组件的`Width(40)`属性不生效。Modal显示为全屏宽度(80)而非指定的40字符宽度。

**根本原因**: `BorderedNode.Measure()`方法只检查`Style.Width`，而`BorderedBuilder.Width()`将值存储在`Props["width"]`中，两者位置不一致。

---

## 二、系统架构概览

### 2.1 Layer系统约束流程

```
LayerManager.CollectAndLayout()
    │
    ├─► 收集Layer节点
    │   └─► Collector.Collect(vnode)
    │
    ├─► 布局Base层
    │   └─► engine.Layout(baseTree, constraints)
    │
    └─► 布局每个Layer (Modal/Overlay/Tooltip)
        └─► layoutLayer(node, layer, constraints, engine)
            │
            ├─► 为Layer创建约束
            │   └─► layerConstraints = BoxConstraints{
            │       MinWidth: 0,
            │       MaxWidth: constraints.MaxWidth,  // 全屏宽度!
            │       ...
            │   }
            │
            ├─► engine.Layout(content, layerConstraints)
            │
            └─► centerModal(root, constraints)  // 后处理居中
```

### 2.2 Engine双路径布局

```
buildComputedBox(vnode, constraints)
    │
    ├─► TryMeasureLayout(vnode, constraints)
    │   └─► 实现 LayoutMeasurer 接口?
    │       ├─► YES: MeasureLayout() [单次遍历路径]
    │       └─► NO: fallback到两遍路径
    │
    └─► fallback (两遍路径)
        ├─► measureVNode(vnode, constraints)
        └─► getChildConstraints() 计算子节点约束
```

---

## 三、发现的问题

### 🔴 问题1: BorderedNode Width属性处理不一致

**位置**: `runtime/ui/layout.go:1030-1139`

**问题描述**:

```go
// BorderedBuilder.Width() 将宽度存储在 Props 中
func (b *BorderedBuilder) Width(n int) *BorderedBuilder {
    b.node.SetProp("width", n)  // ❌ 存储在 Props
    return b
}

// BorderedNode.Measure() 只检查 Style
func (bn *BorderedNode) Measure(constraints runtime.BoxConstraints) runtime.Size {
    // ...
    // ❌ 只检查 elemStyle.Width，不检查 Props["width"]
    elemStyle := bn.Style()
    if elemStyle.Width > 0 {
        totalWidth = elemStyle.Width
    }
    // ...
}
```

**影响**:
- `ui.Bordered().Width(40)` 的设置被完全忽略
- Modal在Layer系统中使用全屏约束时，无法限制宽度

**修复方案**:

```go
func (bn *BorderedNode) Measure(constraints runtime.BoxConstraints) runtime.Size {
    // ... 现有代码 ...

    // 🔧 修复：同时检查 Props["width"] 和 Style.Width
    explicitWidth := 0
    if props := bn.Props(); props != nil {
        if w, ok := props["width"].(int); ok && w > 0 {
            explicitWidth = w
        }
    }
    if elemStyle.Width > 0 {
        explicitWidth = elemStyle.Width
    }
    if explicitWidth > 0 {
        totalWidth = explicitWidth
    }

    return runtime.Size{Width: totalWidth, Height: totalHeight}
}
```

---

### 🟡 问题2: LayerManager的Modal约束过于宽松

**位置**: `runtime/layer/manager.go:114-121`

**问题描述**:

```go
case rtui.LayerModal:
    // Modals使用全屏约束并将被居中
    layerConstraints = runtime.BoxConstraints{
        MinWidth:  0,                    // ❌ 无最小宽度
        MaxWidth:  constraints.MaxWidth, // ❌ 允许扩展到全屏
        MinHeight: 0,
        MaxHeight: constraints.MaxHeight,
    }
```

**影响**:
- Modal可以扩展到全屏宽度（如果内容需要）
- 依赖Modal自身的`Width()`属性限制大小
- 但由于问题1，`Width()`属性不生效

**设计考虑**:
- 当前设计允许"自适应大小"的Modal（如根据内容动态调整）
- 与"固定大小"的Modal（通过`Width()`指定）需要兼容

**可选方案**:

**方案A**: 保持当前宽松约束，修复问题1即可
- 优点：支持动态大小Modal
- 缺点：完全依赖Content节点正确实现尺寸限制

**方案B**: 检查Content是否有明确的width约束
```go
case rtui.LayerModal:
    layerConstraints = runtime.BoxConstraints{
        MinWidth:  0,
        MaxWidth:  constraints.MaxWidth,
        MinHeight: 0,
        MaxHeight: constraints.MaxHeight,
    }
    // 🔧 如果Content有明确width，使用tight约束
    if props := node.Content.Props(); props != nil {
        if w, ok := props["width"].(int); ok && w > 0 {
            layerConstraints.MaxWidth = w
        }
    }
```

---

### 🟢 问题3: Engine.measureBordered与BorderedNode.Measure逻辑重复

**位置**:
- `runtime/compute/engine.go:537-620` (`measureBordered`)
- `runtime/ui/layout.go:1030-1139` (`BorderedNode.Measure`)

**问题描述**:
两个函数都实现了Bordered节点的测量逻辑，但实现细节略有不同：
- `measureBordered`正确处理了`Props["width"]`
- `BorderedNode.Measure`没有处理

**影响**:
- 导致不同路径下行为不一致
- 维护两套相似代码增加出错风险

**长期方案**: 统一到单一实现，让`BorderedNode.Measure()`成为唯一的测量入口。

---

## 四、约束传递链路分析

### 4.1 正常约束流 (Base Layer)

```
VStack (主容器)
    │
    ├─► constraints: {Min: 0, Max: 80}  (80x24屏幕)
    │
    └─► child (HStack)
        │
        └─► constraints: {Min: 0, Max: 80}
            │
            └─► 返回尺寸: 80x1 (填充可用宽度)
```

### 4.2 Modal约束流 (LayerModal)

```
LayerManager.layoutLayer()
    │
    ├─► layerConstraints: {Min: 0, Max: 80}  (全屏约束)
    │
    └─► Bordered.Width(40)  ❌ 被忽略!
        │
        └─► Measure(layerConstraints)
            │
            ├─► 检查 Style.Width → 0 (没有设置)
            ├─► 检查 Props["width"] → ❌ 未检查!
            │
            └─► 返回尺寸: 80x? (填充到MaxWidth)
```

---

## 五、修复优先级

| 优先级 | 问题 | 影响 | 修复工作量 |
|--------|------|------|------------|
| P0 | BorderedNode.Width()不生效 | Modal全屏显示 | 小 |
| P1 | LayerManager Modal约束过于宽松 | Modal可能过大 | 中 |
| P2 | 重复测量逻辑 | 维护成本高 | 大 |

---

## 六、修复建议

### 立即修复 (P0)

修改 `runtime/ui/layout.go` 的 `BorderedNode.Measure()` 方法：

```go
// 第1129-1136行，修改为：
// Apply explicit style dimensions if set
elemStyle := bn.Style()
explicitWidth := 0
explicitHeight := 0

if elemStyle.Width > 0 {
    explicitWidth = elemStyle.Width
}
if elemStyle.Height > 0 {
    explicitHeight = elemStyle.Height
}

// 🔧 FIX: 同时检查 Props 中的 width/height
if props := bn.Props(); props != nil {
    if w, ok := props["width"].(int); ok && w > 0 && explicitWidth == 0 {
        explicitWidth = w
    }
    if h, ok := props["height"].(int); ok && h > 0 && explicitHeight == 0 {
        explicitHeight = h
    }
}

if explicitWidth > 0 {
    totalWidth = explicitWidth
}
if explicitHeight > 0 {
    totalHeight = explicitHeight
}
```

### 后续优化 (P1)

在 `LayerManager.layoutLayer()` 中添加对显式尺寸的尊重：

```go
case rtui.LayerModal:
    layerConstraints = runtime.BoxConstraints{
        MinWidth:  0,
        MaxWidth:  constraints.MaxWidth,
        MinHeight: 0,
        MaxHeight: constraints.MaxHeight,
    }
    // 尊重Modal Content的显式尺寸
    if content := node.Content; content != nil {
        if props := content.Props(); props != nil {
            if w, ok := props["width"].(int); ok && w > 0 {
                // 使用tight宽度约束
                layerConstraints.MinWidth = w
                layerConstraints.MaxWidth = w
            }
        }
    }
```

---

## 七、测试验证

修复后应验证以下场景：

```bash
# 1. 基础Modal宽度测试
go test -v -run TestModalCentering ./examples/sandbox/demo/

# 2. 使用TUI_LAYOUT_DEBUG查看约束传递
TUI_LAYOUT_DEBUG=true go test -v -run TestModalCentering

# 3. 运行完整Demo验证视觉效果
cd examples/ui_demos/demo1_full_featured && go run main.go
```

**预期结果**:
- Modal宽度应为40字符（左右各有20字符空白用于居中）
- Modal内的HStack/内容应正确居中

---

## 八、附录：代码位置索引

| 文件 | 行号 | 内容 |
|------|------|------|
| `runtime/layer/manager.go` | 103-157 | `layoutLayer()` - Layer约束创建 |
| `runtime/layer/manager.go` | 159-207 | `centerModal()` - Modal居中逻辑 |
| `runtime/compute/engine.go` | 537-620 | `measureBordered()` - Bordered测量 |
| `runtime/compute/engine.go` | 777-842 | `getChildConstraints()` - 子节点约束 |
| `runtime/ui/layout.go` | 848-858 | `BorderedBuilder.Width/Height()` - Builder方法 |
| `runtime/ui/layout.go` | 1030-1139 | `BorderedNode.Measure()` - 尺寸计算 |
| `runtime/ui/layout_measurement.go` | 40-208 | `measureHStackLayout()` - HStack单次遍历 |
| `runtime/ui/layout_measurement.go` | 214-382 | `measureVStackLayout()` - VStack单次遍历 |

---

*报告生成时间: 2025-02-07*
*审查者: Claude*
*状态: ✅ P0问题已修复*
*修复位置: runtime/ui/layout.go:1141-1157*
