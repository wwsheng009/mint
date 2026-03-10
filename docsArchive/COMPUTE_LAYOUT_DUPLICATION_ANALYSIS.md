# runtime/compute 与 runtime/layout 重复代码分析

## 概述

`runtime/compute` 和 `runtime/layout` 两个包存在大量功能重复。这是因为架构演进：
- `runtime/compute` - 旧版布局引擎（VNode 依赖）
- `runtime/layout` - 新版布局引擎（Fiber-first，无 VNode 依赖）

---

## 一、重复功能对比

| 功能 | runtime/compute | runtime/layout | 状态 |
|------|-----------------|----------------|------|
| **布局盒子** | `ComputedBox` (~100行) | `LayoutBox` (~40行) | ⚠️ 重复 |
| **布局结果** | `ComputedLayout` | `LayoutResult` | ⚠️ 重复 |
| **布局引擎** | `Engine` (~2000行) | `Engine` (~300行) | ⚠️ 重复 |
| **缓存** | `LayoutCache` | `Cache` | ⚠️ 重复 |
| **脏标记** | `DirtyTracker` | `DirtyTracker` | ⚠️ 重复 |
| **约束类型** | `runtime.BoxConstraints` | `Constraints` | ⚠️ 重复 |
| **Flex 布局** | engine.go 内实现 | `flex.go` | ⚠️ 重复 |
| **边框处理** | `measureBordered()` | `border.go` | ⚠️ 重复 |
| **表格布局** | `measureTable()` | `table.go` | ⚠️ 重复 |
| **HitMap** | 内置 | `hitmap.go` | ⚠️ 重复 |
| **验证器** | `BoundsValidator` | `Validator` | ⚠️ 重复 |
| **Grid 布局** | ❌ 无 | `grid.go` | ✅ 仅 layout |
| **Wrap 布局** | ❌ 无 | `wrap.go` | ✅ 仅 layout |
| **绝对定位** | ❌ 无 | `position.go` | ✅ 仅 layout |
| **Layer 支持** | ❌ 无 | `layer.go` | ✅ 仅 layout |

---

## 二、类型对比

### 2.1 布局盒子

**runtime/compute ComputedBox:**
```go
type ComputedBox struct {
    VNode VNode           // ❌ VNode 依赖（deprecated）
    runtime.Box           // 位置和尺寸
    Children []*ComputedBox
    Parent *ComputedBox
    LayoutDirty bool
    RenderedText string
    NaturalWidth int
    NodeID uint64         // ✅ Fiber 身份
    DiffKey string        // ✅ Fiber 脏跟踪
    Layer rtui.Layer      // ✅ 图层支持
    ChildFiber *rtui.Fiber
}
```

**runtime/layout LayoutBox:**
```go
type LayoutBox struct {
    ID string
    X, Y int
    Width, Height int
    Baseline int
    Layer Layer
    ZIndex int
    Border Border
    Children []*LayoutBox
}
```

**分析：**
- `ComputedBox` 包含 VNode 引用（遗留）
- `LayoutBox` 是纯布局数据，无 UI 框架依赖
- `ComputedBox` 的 NodeID/DiffKey/Layer 字段是迁移期间添加的

### 2.2 布局引擎

**runtime/compute Engine (~2000行):**
- 直接操作 VNode 树
- 包含 Flex、HStack、VStack、表格等布局算法
- 使用 `runtime.BoxConstraints`
- 与 `reconciler.Fiber` 耦合

**runtime/layout Engine (~300行):**
- 操作抽象的 `Node` 接口
- 通过接口扩展支持不同布局（FlexStyleProvider, GridStyleProvider 等）
- 使用自己的 `Constraints` 类型
- 无外部依赖，纯布局引擎

---

## 三、布局算法重复

### 3.1 Flex 布局

**runtime/compute (engine.go):**
- 在 `measureLayoutChildren()` 中实现
- ~200 行代码
- 与 VNode 耦合

**runtime/layout (flex.go):**
- `FlexLayout` 类型
- ~500 行代码
- 完整的 Flexbox 实现（包括 shrink 计算）
- 支持 `FlexStyleProvider` 接口

**结论：** `runtime/layout/flex.go` 更完整，应该保留

### 3.2 表格布局

**runtime/compute:**
- `measureTable()` 方法
- 简单实现

**runtime/layout:**
- `table.go` 独立文件
- 更完整的实现

**结论：** `runtime/layout/table.go` 更完整

### 3.3 边框处理

**runtime/compute:**
- `measureBordered()` 方法
- 在 engine.go 中

**runtime/layout:**
- `border.go` 独立文件
- `Border` 类型，包含完整边框信息

**结论：** `runtime/layout/border.go` 更模块化

---

## 四、缓存系统对比

### runtime/compute LayoutCache:
```go
type LayoutCache struct {
    cache map[LayoutCacheKey]LayoutCacheEntry
}
type LayoutCacheKey struct {
    VNodeType   string
    VNodeKey    string
    Constraints runtime.BoxConstraints
    PropsHash   uint64
    ContentHash uint64
}
```

### runtime/layout Cache:
```go
type Cache struct {
    entries map[string]*CachedLayout
    maxSize int
}
// 使用 SHA256 哈希作为键
```

**分析：**
- 两者实现类似功能
- compute 版本使用更详细的键结构
- layout 版本使用 SHA256 哈希

---

## 五、当前调用情况

### runtime/compute 的调用者：

| 调用位置 | 用途 |
|----------|------|
| `rendering_pipeline.go` | `compute.Engine.Layout()` |
| `layer/manager.go` | `compute.Engine.Layout()` |
| `declarative_node.go` | 通过 RenderingPipeline 间接使用 |

### runtime/layout 的调用者：

| 调用位置 | 用途 |
|----------|------|
| `layout_switcher.go` | `NewLayoutEngineAdapter` 使用 `layout.Engine` |
| `fiber_adapter.go` | `FiberToNodeAdapterPure` 转换 Fiber 到 layout.Node |
| Fiber-first 路径 | `fiberFirstPaint()` 通过 `NewLayoutEngineAdapter` |

---

## 六、迁移建议

### 阶段 1：标记 deprecated（已完成）
- `runtime/compute` 中的类型和方法已标记 deprecated

### 阶段 2：统一布局类型
1. 将 `ComputedBox` 的特殊字段（NodeID, DiffKey, Layer）迁移到 `LayoutBox`
2. 创建适配器将 `LayoutBox` 转换为 Paint 阶段需要的格式

### 阶段 3：删除 runtime/compute 中的布局算法
1. 删除 `engine.go` 中的 `measureLayoutChildren()`, `measureBordered()`, `measureTable()`
2. 保留适配器代码（`adapter_*.go`）用于 VNode 兼容

### 阶段 4：统一缓存系统
1. 使用 `runtime/layout/cache.go` 作为唯一缓存
2. 在 compute 包中创建包装器

### 阶段 5：完全移除 runtime/compute
1. 所有布局通过 `runtime/layout` 完成
2. `rendering_pipeline.go` 使用 `layout.Engine`
3. 删除 `runtime/compute` 包

---

## 七、保留与删除建议

### ✅ 保留（runtime/layout）
- `types.go` - 核心类型定义
- `flex.go` - Flexbox 布局
- `grid.go` - 网格布局
- `wrap.go` - 换行布局
- `table.go` - 表格布局
- `position.go` - 绝对定位
- `border.go` - 边框处理
- `cache.go` - 布局缓存
- `dirty.go` - 脏标记
- `layer.go` - 图层支持
- `hitmap.go` - 命中映射
- `validator.go` - 验证器

### ⚠️ 待迁移（runtime/compute）
- `types.go` 中的 `ComputedBox` → 需要保留（Paint 阶段依赖）
- `engine.go` 中的布局算法 → 迁移到 layout 包
- `adapter_*.go` → 保留用于 VNode 兼容

### ❌ 可删除（runtime/compute）
- `LayoutCache` → 被 `layout.Cache` 替代
- `DirtyTracker` → 被 `layout.DirtyTracker` 替代
- `FlexDistributionInfo` → 被 `layout.FlexCache` 替代
- `measureLayoutChildren()` → 被 `layout.FlexLayout` 替代
- `measureBordered()` → 被 `layout.Border` 替代
- `measureTable()` → 被 `layout.Table` 替代

---

## 八、代码量统计

| 包 | 文件 | 行数 | 保留? |
|----|------|------|-------|
| runtime/compute | engine.go | ~2237 | ⚠️ 需重构 |
| runtime/compute | types.go | ~200 | ⚠️ 部分保留 |
| runtime/compute | cache.go | ~180 | ❌ 可删除 |
| runtime/compute | dirty_tracker.go | ~120 | ❌ 可删除 |
| runtime/compute | adapter_*.go | ~300 | ✅ 保留 |
| **compute 总计** | | **~3037** | |
| | | | |
| runtime/layout | types.go | ~500 | ✅ 保留 |
| runtime/layout | flex.go | ~550 | ✅ 保留 |
| runtime/layout | grid.go | ~400 | ✅ 保留 |
| runtime/layout | table.go | ~200 | ✅ 保留 |
| runtime/layout | wrap.go | ~150 | ✅ 保留 |
| runtime/layout | position.go | ~200 | ✅ 保留 |
| runtime/layout | border.go | ~100 | ✅ 保留 |
| runtime/layout | cache.go | ~200 | ✅ 保留 |
| runtime/layout | dirty.go | ~80 | ✅ 保留 |
| runtime/layout | layer.go | ~100 | ✅ 保留 |
| runtime/layout | hitmap.go | ~150 | ✅ 保留 |
| runtime/layout | validator.go | ~100 | ✅ 保留 |
| **layout 总计** | | **~2730** | |

---

## 九、下一步行动

1. [ ] **评估 ComputedBox 依赖** - PaintEngine 是否可以直接使用 LayoutBox？
2. [ ] **创建统一适配器** - 将 LayoutBox 转换为 Paint 阶段需要的格式
3. [ ] **重构 RenderingPipeline** - 使用 layout.Engine 替代 compute.Engine
4. [ ] **删除重复代码** - 移除 compute 中的布局算法
5. [ ] **更新测试** - 确保迁移后测试通过
