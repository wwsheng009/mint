# ⚠️ [已过时] 为什么 framework/app.go 使用 Paintable 而不是 Layer 系统？

**⚠️ 此文档已过时，包含不准确的架构分析！**

**请参阅**: `docs/layer/FIBER_FIRST_LAYER_SYSTEM.md` - 正确的 Fiber-First Layer 系统说明

---

## 🚨 架构澄清

### 错误假设

本文档原始基于以下错误假设：
- ❌ Framework Paintable (旧系统) 和 Runtime Layer (新系统) 是**两套并存的系统**
- ❌ Framework/App 还没有完整集成 Layer 系统
- ❌ 需要手动在 Framework 层添加 layerManager

### 正确理解

**Mint TUI 只有**一套**渲染系统！**

架构路径是单一的：
```
framework/App.Run()
    ↓
DeclarativeNode.Paint() (Paintable 接口)
    ↓
PipelineRenderer.Render()
    ↓
    ├─→ hasLayerNodes(vnode) 检测
    │       ├─ Fiber 树检查 (优先)
    │       └─ VNode 树检查 (回退)
    │
    └─→ RenderingPipeline
            ├─→ RenderLayers() (有 Layer 时)
            └─→ Render() (无 Layer 时)
```

- **Framework Paintable 接口不是独立的渲染系统**，只是渲染流程的入口点
- **Layer 系统已经完整集成**在 `internal/render` 包中
- **PipelineRenderer 自动检测 Layer 标记**并调用 `RenderLayers()`
- **Fiber 存储持久化的 Layer 状态**，跨帧保持

**详细信息请参阅**: `docs/layer/FIBER_FIRST_LAYER_SYSTEM.md`

---

## 原文档内容（保留作为历史参考）

以下内容基于错误的架构假设，仅作为历史参考。**请勿依赖此文档进行当前架构的理解。**

---

**架构解析：两套系统的历史原因与融合方案**

---

## 📊 问题核心 (历史分析)

您发现了 Mint TUI 有**两套并存的渲染系统**（当时认为）：

1. **Framework Paintable 系统** - framework/app.go 使用
2. **Runtime Layer 系统** - runtime/layer/ 定义

**当时认为这是架构演进的历史遗留问题。**

---

## 🏛️ 两套系统对比 (历史分析)

### 系统1: Framework Paintable（旧系统）

**位置**: `framework/component/paintable.go`

```go
type Paintable interface {
    Node
    Paint(ctx PaintContext, buf *paint.Buffer)
}
```

**当时的使用路径**:
```
Application (用户代码)
    ↓
framework/App.Run()
    ↓
framework/App.render()
    ↓
检查: root.(component.Paintable)
    ↓
调用: root.Paint(ctx, buffer)
    ↓
直接绘制到 buffer
```

**当时的特点**:
- ✅ 简单直接
- ✅ Framework 层控制
- ❌ 不支持 Layer/ZIndex
- ❌ 没有覆盖层概念

### 系统2: Runtime Layer（新系统）

**位置**: `runtime/layer/manager.go` (已废弃)

```go
type Manager struct {
    collector *Collector
    layouts   LayerLayouts
}

func (m *Manager) CollectAndLayout(vnode, constraints, engine) {
    // 收集 LayerBase, LayerOverlay, LayerModal, LayerTooltip, LayerInspector
    // 分别布局每个层
    // 返回 LayerLayouts
}
```

**当时的使用路径**:
```
Application (带 Layer 标记的 VNode)
    ↓
LayerManager.CollectAndLayout()
    ↓
返回 LayerLayouts (per-layer layout)
    ↓
PaintEngine.PaintLayers(layouts, buffer)
    ↓
按 z-index 顺序渲染所有层
```

**当时的特点**:
- ✅ 支持 Layer/ZIndex
- ✅ 支持覆盖层
- ✅ 支持多层级合成
- ❌ 需要显式集成

---

## 📜 历史演进 (当时认为)

### 阶段1: 初始架构（V1）

```
Application
    ↓
Component.Render() → string
    ↓
直接输出到终端
```

**问题**:
- 没有布局系统
- 没有覆盖层
- 简单但不灵活

### 阶段2: Framework Paintable（V2）

```
Application
    ↓
framework/App.Run()
    ↓
检查 Paintable 接口
    ↓
root.Paint(ctx, buffer)
    ↓
buffer → terminal
```

**改进**:
- ✅ 添加 Buffer 抽象
- ✅ 添加 Paintable 接口
- ✅ 组件绘制能力
- ⚠️ 但没有 Layer 系统

### 阶段3: Runtime Layer 系统（V3）

```
Application (带 Layer 标记)
    ↓
LayerManager.CollectAndLayout()
    ↓
PaintEngine.PaintLayers()
    ↓
多层级渲染 (z-index 0-4)
```

**新增**:
- ✅ 完整的 Layer 系统
- ✅ ZIndex 支持
- ✅ Modal, Tooltip, Overlay 支持
- ⚠️ 但 framework/app 还没集成

### 当前状态（当时认为 V3.5）- 两套并存

```
路径1: Framework Paintable (旧)
framework/App → Paintable.Paint() → buffer

路径2: Runtime Layer (新)
runtime/layer → LayerManager → PaintEngine.PaintLayers() → buffer

⚠️ 两条路径并存，没有统一！
```

---

## 🔧 当时的解决方案

### 方案A: 在 App 中集成 Layer 系统（推荐，但实际不需要）

**修改 framework/app.go**:

```go
type App struct {
    // ... 现有字段

    // Layer 支持 (NEW)
    layerManager *layer.Manager  // ❌ 实际不需要！
    useLayers    bool             // ❌ 实际不需要！
}

func (a *App) render() {
    if a.useLayers {
        // 新路径: 使用 Layer 系统
        a.renderWithLayers()
    } else {
        // 旧路径: 使用 Paintable (当前)
        a.renderWithPaintable()
    }
}
```

**实际上**: Layer 系统已经在 `internal/render` 包中完整实现，Framework 通过 DeclarativeNode 桥接。

### 方案B: 让 Paintable 组件支持 Layer

**修改 component.Paintable**:

```go
type Paintable interface {
    Node

    // 现有方法
    Paint(ctx PaintContext, buf *paint.Buffer)

    // 新方法 (optional)
    GetLayer() Layer  // 返回组件所在的层
}
```

**实际上**: VNode 接口已经有 `GetLayer()` 方法，不需要修改 Paintable。

---

## 💡 当时的解决方案（临时方案）

### 我们当时做了"部分集成":

```go
// 在 demo2 中
func RuntimeDemoWithInspectorOverlay() ui.VNode {
    // 1. 构建应用内容
    appContent := buildDemoContent(...)

    // 2. 获取 Inspector 覆盖层
    if showInspector {
        inspectorOverlay := globalInspector.RenderOverlay()
        // inspectorOverlay.SetLayer(LayerInspector) 已设置

        // 3. 通过 VStack 组合
        return ui.VStack(appContent, inspectorOverlay)
    }

    return appContent
}
```

**这是"伪 Layer"**:
- ✅ Inspector 有 Layer 标记
- ✅ PaintEngine 知道如何渲染
- ❌ 但通过 VStack 组合，不算真正的 Layer 系统

**实际上**: 这是正确的 Fiber-first 架构，VStack 组合被 Fiber 树管理，Layer 属性存储在 Fiber 节点中。

---

## 🎯 实际的解决方案 (现在理解)

### Fiber-First 架构下的 Layer 系统

1. **Layer 存储在 Fiber 节点**
   ```go
   type Fiber struct {
       // ...
       Layer Layer  // ← 持久化的 Layer 状态
   }
   ```

2. **初始化时从 VNode 获取**
   ```go
   return &Fiber{
       // ...
       Layer: vnode.GetLayer(),  // 初始化
   }
   ```

3. **自动检测和渲染**
   ```go
   // PipelineRenderer
   hasLayers := r.hasLayerNodes(vnode)
   if hasLayers {
       r.pipeline.RenderLayers(vnode, fiber, constraints, buf)
   } else {
       r.pipeline.Render(vnode, fiber, constraints, buf)
   }
   ```

4. **多层级渲染**
   - `layout.Engine.Layout()` - 统一布局
   - `applyLayerTransforms()` - Modal 居中等
   - `buildPaintablePlanes()` - 按层级分组
   - `PaintEngine.PaintPaintablePlanes()` - 绘制

---

## ✅ 更新建议

### 对于开发者

1. **阅读新文档**
   - `docs/layer/FIBER_FIRST_LAYER_SYSTEM.md` - 正确的架构说明

2. **理解 Fiber-first 架构**
   - Layer 存储在 Fiber，不在独立的管理器中
   - 通过 DeclarativeNode 桥接，不是两套系统

3. **使用组件 Layer API**
   - `VNode.GetLayer()` / `VNode.SetLayer()`
   - Builder 便捷方法

### 对于文档维护者

1. **归档此文档**
   - 移动到 `docs/deprecated/` 目录
   - 添加 "DEPRECATED" 前缀

2. **更新交叉引用**
   - 指向 `FIBER_FIRST_LAYER_SYSTEM.md`
   - 更新相关文档的链接

---

## 🎓 经验教训

### 1. 架构理解的重要性

- ❌ 错误: 认为有"两套并存系统"
- ✅ 正确: Framework 通过 DeclarativeNode 桥接到 Runtime，Layer 系统已集成

### 2. 深入代码分析

- 需要仔细阅读 `internal/render/` 包的实现
- 不应仅凭接口推测架构
- Fiber-first 架构改变了状态管理方式

### 3. 文档时效性

- 架构演进可能使文档过时
- 需要定期验证文档是否与代码一致
- 标记过时文档比删除更有价值（保留历史）

---

## 📚 正确的文档

### 推荐文档

- ✅ `docs/layer/FIBER_FIRST_LAYER_SYSTEM.md` - Fiber-First Layer 系统完整说明
- ✅ `internal/render/` - 源码实现是最好的文档

### 过时文档

- ❌ `docs/layer/TWO_RENDERING_SYSTEMS_EXPLAINED.md` - 本文档（已过时）

---

**文档版本**: 1.0 (DEPRECATED)
**创建日期**: 2025-02-08
**归档日期**: 2026-02-23
**状态**: ❌ **已过时 - 基于 Fiber-first 架构更新**
**替代文档**: ✅ **`docs/layer/FIBER_FIRST_LAYER_SYSTEM.md`**
