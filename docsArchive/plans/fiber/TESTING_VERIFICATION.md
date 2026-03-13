# Fiber 统一架构重构 - 测试验证方法

**版本**: 1.0
**日期**: 2026-02-14
**负责人**: [待分配]
**状态**: 进行中

---

## 📋 目录

1. [测试策略概述](#测试策略概述)
2. [测试框架和工具](#测试框架和工具)
3. [测试分类](#测试分类)
4. [各阶段详细测试方案](#各阶段详细测试方案)
5. [自动化测试编写指南](#自动化测试编写指南)
6. [手动测试流程](#手动测试流程)
7. [性能基准和阈值](#性能基准和阈值)
8. [回归测试方案](#回归测试方案)
9. [测试数据准备](#测试数据准备)
10. [缺陷管理流程](#缺陷管理流程)

---

## 测试策略概述

### 测试金字塔

```
        ┌──────────┐
        │  E2E     │  ~10% (用户场景测试)
        │  Tests   │
       ─┴──────────┴─
      ┌────────────┐
      │ Integration│  ~20% (集成测试)
      │   Tests    │
     ─┴────────────┴───
    ┌─────────────────┐
    │   Unit Tests    │  ~70% (单元测试)
    │                 │
   ─┴─────────────────┴──
```

### 测试覆盖目标

| 指标 | 目标值 | 测量方法 |
|------|--------|---------|
| 整体代码覆盖率 | >75% | `go test -cover` |
| 核心包覆盖率 | >85% | `go test -cover ./runtime/...` |
| ui 包覆盖率 | >80% | `go test -cover ./ui/...` |
| render 包覆盖率 | >80% | `go test -cover ./runtime/render/...` |
| reconciler 包覆盖率 | >85% | `go test -cover ./internal/reconciler/...` |
| event 包覆盖率 | >80% | `go test -cover ./runtime/event/...` |

### 测试层级定义

| 层级 | 定义 | 示例 | 自动化 |
|------|------|------|--------|
| **单元测试** | 测试单个函数/方法 | `TestCreateFiberLayerCopy` | ✅ 是 |
| **集成测试** | 测试多个协作组件 | `TestLayoutFiberComplete` | ✅ 是 |
| **端到端测试** | 从用户角度测试完整流程 | `TestModalUserFlow` | ✅ 是 |
| **性能测试** | 测试性能指标 | `BenchmarkLayout` | ✅ 是 |
| **视觉测试** | 验证 UI 显示效果 | 手动/截图对比 | 部分 |

---

## 测试框架和工具

### 单元测试

#### Go 标准测试框架

```go
// 基础测试
func TestCreateFiber(t *testing.T) {
    // 测试逻辑
}

// 基准测试
func BenchmarkCreateFiber(b *testing.B) {
    // 基准测试逻辑
}

// 子测试
func TestFiber(t *testing.T) {
    t.Run("Layer copy", func(t *testing.T) {
        // 子测试 1
    })
    t.Run("Layer default value", func(t *testing.T) {
        // 子测试 2
    })
}
```

#### 表驱动测试

```go
func TestCreateFiberLayerCopy(t *testing.T) {
    tests := []struct {
        name     string
        vnode    *VNode
        expected Layer
    }{
        {
            name:     "base layer",
            vnode:    createVNodeWithLayer(LayerBase),
            expected: LayerBase,
        },
        {
            name:     "overlay layer",
            vnode:    createVNodeWithLayer(LayerOverlay),
            expected: LayerOverlay,
        },
        {
            name:     "modal layer",
            vnode:    createVNodeWithLayer(LayerModal),
            expected: LayerModal,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            fiber := CreateFiber(tt.vnode)
            if fiber.Layer != tt.expected {
                t.Errorf("Expected Layer %v, got %v", tt.expected, fiber.Layer)
            }
        })
    }
}
```

#### 断言库

使用 `github.com/stretchr/testify/assert`：

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestCreateFiber(t *testing.T) {
    vnode := createTestVNode()
    fiber := CreateFiber(vnode)

    assert.NotNil(t, fiber)
    assert.Equal(t, vnode.GetLayer(), fiber.Layer)
    assert.Nil(t, fiber.ComputedBox)
    assert.NotZero(t, fiber.NodeID)
}
```

### 集成测试

#### TestRunner (ui/test.go)

```go
func TestModalIntegration(t *testing.T) {
    app, err := ui.TestRun(ModalComponent, ui.TestWithSize(80, 24))
    if err != nil {
        t.Fatal(err)
    }
    defer app.Close()

    // Force render
    app.ForceRender()

    // Inject events
    app.InjectKey('o')  // Open modal
    app.InjectSpecialKey(platform.KeyEnter)

    // Assertions
    if err := app.AssertRender("Modal Title"); err != nil {
        t.Error(err)
    }

    // Verify component state
    ctx := app.GetContext()
    // ... verify state
}
```

#### RunTest (Framework App 测试)

```go
func TestButtonEventIntegration(t *testing.T) {
    ta, err := ui.RunTest(ButtonComponent, ui.WithWidth(40), ui.WithHeight(10))
    if err != nil {
        t.Fatal(err)
    }
    defer ta.Close()

    ta.ForceRender()
    ta.InjectMouse(10, 10, platform.MouseButtonLeft, platform.MouseActionPress)

    if err := ta.AssertRender("Clicked"); err != nil {
        t.Error(err)
    }
}
```

### Mock 和测试辅助

#### Sandbox Mock

```go
func TestEventInjection(t *testing.T) {
    sb := mock.New(80, 24)
    if err := sb.Initialize(nil); err != nil {
        t.Fatal(err)
    }
    defer sb.Close()

    // Inject events
    sb.InjectKey('a')
    sb.InjectSpecialKey(platform.KeyTab)
    sb.ProcessEvents()

    // Verify
    // ...
}
```

### 性能测试

#### Benchmark 基础

```go
func BenchmarkCreateFiber(b *testing.B) {
    vnode := createLargeVNode(1000) // 1000 个节点

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = CreateFiber(vnode)
    }
}

func BenchmarkLayoutFiber(b *testing.B) {
    fiber := createLargeFiber(1000)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = layoutFiber(fiber)
    }
}
```

#### pprof 性能分析

```bash
# CPU 分析
go test -cpuprofile=cpu.prof -bench=. ./...
go tool pprof cpu.prof

# 内存分析
go test -memprofile=mem.prof -bench=. ./...
go tool pprof mem.prof

# 内存分配分析
go test -allocspace=alloc.prof -bench=. ./...
go tool pprof alloc.prof
```

### 竞态检测

```bash
# 运行竞态检测器
go test -race ./...

# 短测试模式（跳过慢速测试）
go test -race -short ./...
```

### 覆盖率分析

```bash
# 生成覆盖率报告
go test -coverprofile=coverage.out ./...

# 查看覆盖率
go tool cover -func=coverage.out

# 生成 HTML 报告
go tool cover -html=coverage.out -o coverage.html
```

### 视觉测试工具

#### 缓存对比

```go
import (
    "testing"
    "github.com/wwsheng009/mint/internal/render"
    "github.com/stretchr/testify/require"
)

func TestVisualRendering(t *testing.T) {
    app := setupTestApp()
    defer app.Close()

    app.ForceRender()

    // 获取渲染输出
    buffer := app.GetBuffer()
    actual := buffer.String()

    // 对比预期
    expected := loadGoldenFile("testdata/rendering.golden")
    require.Equal(t, expected, actual)
}
```

---

## 测试分类

### 按测试目标分类

| 类别 | 目标 | 工具 | 自动化 |
|------|------|------|--------|
| **单元测试** | 验证单个函数/方法正确性 | go test | ✅ |
| **集成测试** | 验证组件协作正确性 | TestRunner, RunTest | ✅ |
| **E2E 测试** | 验证用户场景 | TestRunner | ✅ |
| **性能测试** | 验证性能基线 | Benchmark, pprof | ✅ |
| **压力测试** | 验证稳定性 | 长期运行测试 | ✅ |
| **视觉测试** | 验证 UI 显示效果 | 截图对比 | 部分 |
| **安全测试** | 验证无恶意代码 | Code Review | ❌ |

### 按测试阶段分类

| 阶段 | 测试类型 | 比例 |
|------|---------|------|
| **开发阶段** | 单元测试 | 100% |
| **集成阶段** | 单元 + 集成测试 | 80% + 20% |
| **系统测试** | 集成 + E2E + 性能 | 60% + 20% + 20% |
| **验收测试** | E2E + 视觉 + 手动 | 70% + 20% + 10% |

### 按测试方法分类

| 方法 | 优点 | 缺点 | 适用场景 |
|------|------|------|---------|
| **黑盒测试** | 从用户角度测试 | 难定位问题 | E2E 测试 |
| **白盒测试** | 覆盖所有代码路径 | 耗时 | 单元测试 |
| **灰盒测试** | 平衡 | 需要了解部分内部 | 集成测试 |

---

## 各阶段详细测试方案

### Phase 1: 基础设施测试

#### 1.1 Fiber 结构更新测试

**单元测试文件**: `runtime/ui/fiber_test.go`

```go
// TestFiberLayerFieldExists 验证 Layer 字段存在
func TestFiberLayerFieldExists(t *testing.T) {
    fiber := &Fiber{}
    fiber.Layer = LayerOverlay

    assert.Equal(t, LayerOverlay, fiber.Layer)
}

// TestFiberComputedBoxFieldExists 验证 ComputedBox 字段存在
func TestFiberComputedBoxFieldExists(t *testing.T) {
    fiber := &Fiber{}
    box := &compute.ComputedBox{}
    fiber.ComputedBox = box

    assert.Equal(t, box, fiber.ComputedBox)
}
```

#### 1.2 CreateFiber 测试

```go
// TestCreateFiberLayerCopy 验证 Layer 从 VNode 拷贝
func TestCreateFiberLayerCopy(t *testing.T) {
    tests := []struct {
        name    string
        layer   Layer
    }{
        {"base layer", LayerBase},
        {"overlay layer", LayerOverlay},
        {"modal layer", LayerModal},
        {"tooltip layer", LayerTooltip},
        {"inspector layer", LayerInspector},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            vnode := createTestVNodeWithLayer(tt.layer)
            fiber := CreateFiber(vnode)

            assert.Equal(t, tt.layer, fiber.Layer, "Layer should be copied from VNode")
        })
    }
}

// TestCreateFiberLayerDefaultValue 验证无效 Layer 设为默认值
func TestCreateFiberLayerDefaultValue(t *testing.T) {
    // 创建一个 Layer 无效的 VNode (假设 Layer(99) 无效)
    vnode := createTestVNodeWithLayer(Layer(99))
    fiber := CreateFiber(vnode)

    assert.Equal(t, LayerBase, fiber.Layer, "Invalid Layer should default to LayerBase")
}

// TestCreateFiberComputedBoxNil 验证 ComputedBox 初始化为 nil
func TestCreateFiberComputedBoxNil(t *testing.T) {
    vnode := createTestVNode()
    fiber := CreateFiber(vnode)

    assert.Nil(t, fiber.ComputedBox, "ComputedBox should be nil initially")
}

// BenchmarkCreateFiber 性能基准测试
func BenchmarkCreateFiber(b *testing.B) {
    sizes := []int{10, 100, 1000}
    for _, size := range sizes {
        b.Run(fmt.Sprintf("Size_%d", size), func(b *testing.B) {
            vnode := createLargeVNode(size)
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                _ = CreateFiber(vnode)
            }
        })
    }
}
```

#### 1.3 CloneFiber 测试

```go
// TestCloneFiberLayerPreserved 验证 Layer 被保留
func TestCloneFiberLayerPreserved(t *testing.T) {
    current := &Fiber{Layer: LayerModal}
    clone := CloneFiber(current)

    assert.Equal(t, LayerModal, clone.Layer, "Layer should be preserved")
}

// TestCloneFiberComputedBoxNil 验证 ComputedBox 重置为 nil
func TestCloneFiberComputedBoxNil(t *testing.T) {
    current := &Fiber{
        ComputedBox: &compute.ComputedBox{},
    }
    clone := CloneFiber(current)

    assert.Nil(t, clone.ComputedBox, "ComputedBox should be reset to nil")
}

// TestCloneFiberNodeIDPreserved 验证 NodeID 保持不变
func TestCloneFiberNodeIDPreserved(t *testing.T) {
    current := &Fiber{NodeID: generateNodeID()}
    clone := CloneFiber(current)

    assert.Equal(t, current.NodeID, clone.NodeID, "NodeID should be preserved")
}

// BenchmarkCloneFiber 性能基准测试
func BenchmarkCloneFiber(b *testing.B) {
    sizes := []int{10, 100, 1000}
    for _, size := range sizes {
        b.Run(fmt.Sprintf("Size_%d", size), func(b *testing.B) {
            current := createLargeFiber(size)
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                _ = CloneFiber(current)
            }
        })
    }
}
```

#### 1.4 Reconciler 测试

```go
// TestCompleteWorkLayerCopied 验证 CompleteWork 拷贝 Layer
func TestCompleteWorkLayerCopied(t *testing.T) {
    ctx := NewFiberContextForTest()
    defer CloseFiberContext(ctx)

    vnode := createTestVNodeWithLayer(LayerOverlay)
    workInProgress := CreateFiber(vnode)

    CompleteWork(ctx, workInProgress)

    assert.Equal(t, LayerOverlay, workInProgress.Layer, "Layer should be copied in CompleteWork")
}

// TestReconcileLayerPreserved 验证 Reconcile 后 Layer 保持不变
func TestReconcileLayerPreserved(t *testing.T) {
    tests := []struct {
        name      string
        oldLayer  Layer
        newLayer  Layer
        finalLayer Layer
    }{
        {"layer unchanged", LayerBase, LayerBase, LayerBase},
        {"layer changed", LayerBase, LayerOverlay, LayerOverlay},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            oldVNode := createTestVNodeWithLayer(tt.oldLayer)
            newVNode := createTestVNodeWithLayer(tt.newLayer)

            ctx := NewFiberContextForTest()
            defer CloseFiberContext(ctx)

            oldFiber := CreateFiber(oldVNode)
            newFiber, _ := reconcileFiber(ctx, oldFiber, newVNode)

            assert.Equal(t, tt.finalLayer, newFiber.Layer, "Final Layer should be correct")
        })
    }
}

// TestShouldUpdateUsesDiffKey 验证 shouldUpdate 使用 DiffKey
func TestShouldUpdateUsesDiffKey(t *testing.T) {
    tests := []struct {
        name        string
        oldDiffKey  string
        newDiffKey  string
        shouldUpdate bool
    }{
        {"same diff key", "item1", "item1", false},
        {"different diff key", "item1", "item2", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            old := &Fiber{DiffKey: tt.oldDiffKey}
            new := createTestVNode()
            newVNodeDiffKey = tt.newDiffKey // 需要设置 VNode 的 DiffKey

            result := shouldUpdate(old, new)
            assert.Equal(t, tt.shouldUpdate, result)
        })
    }
}
```

#### 1.5 ComputedBox 更新测试

```go
// TestComputedBoxNodeIDField 验证 NodeID 字段
func TestComputedBoxNodeIDField(t *testing.T) {
    box := &compute.ComputedBox{}
    box.NodeID = 12345

    assert.Equal(t, uint64(12345), box.NodeID)
}

// TestComputedBoxLayerField 验证 Layer 字段
func TestComputedBoxLayerField(t *testing.T) {
    box := &compute.ComputedBox{}
    box.Layer = LayerModal

    assert.Equal(t, LayerModal, box.Layer)
}
```

#### 1.6 BuildHitMapFromFiber 测试

```go
// TestBuildHitMapFromFiberEmpty 验证空 Fiber
func TestBuildHitMapFromFiberEmpty(t *testing.T) {
    fiber := &Fiber{}
    hitMap := BuildHitMapFromFiber(fiber)

    assert.NotNil(t, hitMap)
    assert.Equal(t, 0, hitMap.Len())
}

// TestBuildHitMapFromFiberSingle 验证单节点
func TestBuildHitMapFromFiberSingle(t *testing.T) {
    fiber := createTestFiber(1, LayerBase)
    fiber.ComputedBox = &compute.ComputedBox{
        NodeID:   fiber.NodeID,
        Layer:    fiber.Layer,
        BoxModel: createTestBoxModel(),
    }

    hitMap := BuildHitMapFromFiber(fiber)

    assert.Equal(t, 1, hitMap.Len())
    // 验证 NodeID 正确
}

// TestBuildHitMapFromFiberMultipleLevels 验证多层级 Fiber
func TestBuildHitMapFromFiberMultipleLevels(t *testing.T) {
    // 创建包含 Base、Overlay、Modal 的 Fiber 树
    fiber := createMultiLayerFiber()

    hitMap := BuildHitMapFromFiber(fiber)

    // 验证所有层级都被包含
    assert.Greater(t, hitMap.Len(), 0)
}

// TestBuildHitMapFromFiberZOrder 验证 HITMAP 按 Layer 排序 (Z-order)
func TestBuildHitMapFromFiberZOrder(t *testing.T) {
    fiber := createMultiLayerFiber()

    hitMap := BuildHitMapFromFiber(fiber)

    // 验证 HitMap 中元素按 Layer 排序
    // Base (0) < Overlay (1) < Modal (2) < Tooltip (3) < Inspector (4)
    entries := hitMap.Entries()
    var lastLayer Layer = -1
    for _, entry := range entries {
        assert.GreaterOrEqual(t, int(entry.Layer), int(lastLayer), "Entry.Layers should be in ascending order")
        lastLayer = entry.Layer
    }
}

// BenchmarkBuildHitMapFromFiber 性能测试
func BenchmarkBuildHitMapFromFiber(b *testing.B) {
    sizes := []int{100, 1000, 5000}
    for _, size := range sizes {
        b.Run(fmt.Sprintf("Size_%d", size), func(b *testing.B) {
            fiber := createLargeFiberWithComputedBox(size)
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                _ = BuildHitMapFromFiber(fiber)
            }
        })
    }
}
```

---

### Phase 2: Layout 重构测试

#### 2.1 layoutFiber 测试

```go
// TestLayoutFiberEmpty 验证空 Fiber
func TestLayoutFiberEmpty(t *testing.T) {
    fiber := &Fiber{}
    result := layoutFiber(fiber)

    assert.NotNil(t, result)
    assert.Nil(t, result.ComputedBox)
}

// TestLayoutFiberSingleElement 验证单元素
func TestLayoutFiberSingleElement(t *testing.T) {
    fiber := createTestFiber(1, LayerBase)
    result := layoutFiber(fiber)

    assert.NotNil(t, result)
    assert.NotNil(t, result.ComputedBox, "ComputedBox should be created")
    assert.Equal(t, result.NodeID, result.ComputedBox.NodeID, "NodeID should match")
    assert.Equal(t, result.Layer, result.ComputedBox.Layer, "Layer should match")
}

// TestLayoutFiberVStack 验证 VStack 布局
func TestLayoutFiberVStack(t *testing.T) {
    stack := createVStackFiber(3, LayerBase) // 3 元素
    result := layoutFiber(stack)

    assert.NotNil(t, result)
    assert.NotNil(t, result.ComputedBox)

    // 验证子元素的 ComputedBox
    child := result.FirstChild
    count := 0
    for child != nil {
        assert.NotNil(t, child.ComputedBox, "Child ComputedBox should be created")
        assert.Equal(t, child.NodeID, child.ComputedBox.NodeID)
        child = child.Sibling
        count++
    }
    assert.Equal(t, 3, count, "Should have 3 children")
}

// TestLayoutFiberHStack 验证 HStack 布局
func TestLayoutFiberHStack(t *testing.T) {
    stack := createHStackFiber(3, LayerBase) // 3 元素
    result := layoutFiber(stack)

    assert.NotNil(t, result)
    assert.NotNil(t, result.ComputedBox)

    // 验证子元素水平排列
    // 可以通过比较 ComputedBox.X/Y 坐标验证
}

// TestLayoutFiberNested 验证嵌套布局
func TestLayoutFiberNested(t *testing.T) {
    // VStack > VStack > Element
    fiber := createNestedFiber(3, 3) // 3 层，每层 3 元素
    result := layoutFiber(fiber)

    assert.NotNil(t, result)

    // 递归验证所有节点都有 ComputedBox
    verifyAllComputedBoxes(t, result)
}

// verifyAllComputedBoxes 辅助函数：递归验证所有节点
func verifyAllComputedBoxes(t *testing.T, fiber *Fiber) {
    assert.NotNil(t, fiber.ComputedBox, "Fiber should have ComputedBox")

    child := fiber.FirstChild
    for child != nil {
        verifyAllComputedBoxes(t, child)
        child = child.Sibling
    }
}

// BenchmarkLayoutFiber 性能基准测试
func BenchmarkLayoutFiber(b *testing.B) {
    configurations := []struct {
        name    string
        depth   int
        breadth int
    }{
        {"ShallowWide", 3, 100},   // 浅宽树
        {"DeepNarrow", 100, 3},    // 深窄树
        {"Balanced", 10, 10},      // 平衡树
        {"Simple", 1, 1000},       // 单层 1000 节点
    }

    for _, config := range configurations {
        b.Run(config.name, func(b *testing.B) {
            fiber := createNestedFiber(config.depth, config.breadth)
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                _ = layoutFiber(fiber)
            }
        })
    }
}
```

#### 2.2 measureFiber 测试

```go
// TestMeasureFiberNodeID 验证 NodeID 填充
func TestMeasureFiberNodeID(t *testing.T) {
    fiber := createTestFiber(1, LayerBase)
    fiber.NodeID = 12345

    measureFiber(fiber)

    assert.NotNil(t, fiber.ComputedBox)
    assert.Equal(t, uint64(12345), fiber.ComputedBox.NodeID)
}

// TestMeasureFiberLayer 验证 Layer 填充
func TestMeasureFiberLayer(t *testing.T) {
    fiber := createTestFiber(1, LayerOverlay)

    measureFiber(fiber)

    assert.NotNil(t, fiber.ComputedBox)
    assert.Equal(t, LayerOverlay, fiber.ComputedBox.Layer)
}

// TestMeasureFiberRecursive 验证递归测量
func TestMeasureFiberRecursive(t *testing.T) {
    fiber := createNestedFiber(5, 5)

    measureFiber(fiber)

    count := 0
    verifyAllComputedBoxesHelper(fiber, &count)
    assert.Equal(t, 15625, count, "All nodes should have ComputedBox") // 5^5
}

// verifyAllComputedBoxesHelper 辅助函数
func verifyAllComputedBoxesHelper(fiber *Fiber, count *int) {
    if fiber.ComputedBox != nil {
        *count++
    }

    child := fiber.FirstChild
    for child != nil {
        verifyAllComputedBoxesHelper(child, count)
        child = child.Sibling
    }
}

// TestMeasureFiberBoxModel 验证 BoxModel 计算
func TestMeasureFiberBoxModel(t *testing.T) {
    fiber := createTestFiberWithBoxModel(10, 5, 20, 15) // padding: [10,5,10,5], margin: [20,15,20,15]

    measureFiber(fiber)

    assert.NotNil(t, fiber.ComputedBox.BoxModel)
    // 验证具体的 BoxModel 值
}
```

#### 2.3 Modal/Overlay Layout 验证

```go
// Integration test using TestRunner
func TestModalLayoutIntegration(t *testing.T) {
    app, err := ui.TestRun(ModalComponent, ui.TestWithSize(80, 24))
    if err != nil {
        t.Fatal(err)
    }
    defer app.Close()

    app.ForceRender()

    // 打开 Modal
    app.InjectMouse(40, 12, platform.MouseButtonLeft, platform.MouseActionPress)

    if err := app.AssertRender("Modal Content"); err != nil {
        t.Error(err)
    }

    // 验证 Modal 的 Layer
    fibers := app.GetFibers()
    modalFiber := findFiberByKey(fibers, "modal")
    assert.NotNil(t, modalFiber)
    assert.Equal(t, LayerModal, modalFiber.Layer)
}

// TestOverlayLayoutIntegration Overlay 布局集成测试
func TestOverlayLayoutIntegration(t *testing.T) {
    app, err := ui.TestRun(OverlayComponent, ui.TestWithSize(80, 24))
    if err != nil {
        t.Fatal(err)
    }
    defer app.Close()

    app.ForceRender()

    // 触发 Overlay
    app.InjectKey('o')

    if err := app.AssertRender("Overlay Content"); err != nil {
        t.Error(err)
    }

    // 验证 Overlay 的 Layer
    fibers := app.GetFibers()
    overlayFiber := findFiberByKey(fibers, "overlay")
    assert.NotNil(t, overlayFiber)
    assert.Equal(t, LayerOverlay, overlayFiber.Layer)
}
```

#### 2.4 所有 Layout 调用更新验证

```go
// TestAllLayoutCallsUpdated 验证所有旧 API 调用已移除
func TestAllLayoutCallsUpdated(t *testing.T) {
    // 通过代码分析工具或 grep 验证
    // 在实际中，这应该是一个独立的 CI/CD 检查

    packages := []string{
        "./framework/...",
        "./devtools/...",
        "./examples/...",
    }

    for _, pkg := range packages {
        cmd := exec.Command("go", "grep", "-r", `Layout.*\*VNode`, pkg)
        output, err := cmd.CombinedOutput()
        if err == nil {
            t.Errorf("Found deprecated VNode Layout calls in %s:\n%s", pkg, output)
        }
    }
}
```

---

### Phase 3: RenderPlane 引入测试

#### 3.1 RenderPlanes 类型测试

```go
// TestRenderPlanesBuildFromFiber 单节点 Base
func TestRenderPlanesBuildFromFiberBase(t *testing.T) {
    fiber := createTestFiber(1, LayerBase)
    planes := BuildRenderPlanes(fiber)

    assert.NotNil(t, planes)
    assert.Equal(t, 1, len(planes.Base))
    assert.Equal(t, 0, len(planes.Overlay))
    assert.Equal(t, 0, len(planes.Modal))
    assert.Equal(t, 0, len(planes.Tooltip))
    assert.Equal(t, 0, len(planes.Inspector))
}

// TestRenderPlanesBuildFromFiberMultiLayer 多层级
func TestRenderPlanesBuildFromFiberMultiLayer(t *testing.T) {
    fiber := createMultiLayerFiber() // 包含所有层级
    planes := BuildRenderPlanes(fiber)

    assert.NotNil(t, planes)
    assert.Greater(t, len(planes.Base), 0)
    assert.Greater(t, len(planes.Overlay), 0)
    assert.Greater(t, len(planes.Modal), 0)
    assert.Greater(t, len(planes.Tooltip), 0)
    assert.Greater(t, len(planes.Inspector), 0)
}

// TestRenderPlanesBuildFromFiberEmpty 空 Fiber
func TestRenderPlanesBuildFromFiberEmpty(t *testing.T) {
    fiber := &Fiber{}
    planes := BuildRenderPlanes(fiber)

    assert.NotNil(t, planes)
    assert.Equal(t, 0, len(planes.Base))
    assert.Equal(t, 0, len(planes.Overlay))
    assert.Equal(t, 0, len(planes.Modal))
    assert.Equal(t, 0, len(planes.Tooltip))
    assert.Equal(t, 0, len(planes.Inspector))
}

// TestRenderPlanesBuildFromFiberDefaultLayer 未设置 Layer 的节点归到 Base
func TestRenderPlanesBuildFromFiberDefaultLayer(t *testing.T) {
    fiber := createTestFiber(1, Layer(99)) // 无效 Layer
    fiber.ComputedBox = &compute.ComputedBox{Layer: fiber.Layer}
    planes := BuildRenderPlanes(fiber)

    assert.NotNil(t, planes)
    // 应该归到 Base（或其他默认处理）
    assert.GreaterOrEqual(t, len(planes.Base), 1)
}

// BenchmarkBuildRenderPlanes 性能测试
func BenchmarkBuildRenderPlanes(b *testing.B) {
    sizes := []int{100, 1000, 5000}
    for _, size := range sizes {
        b.Run(fmt.Sprintf("Size_%d", size), func(b *testing.B) {
            fiber := createLargeFiberWithComputedBox(size)
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                _ = BuildRenderPlanes(fiber)
            }
        })
    }
}
```

#### 3.2 LayerManager 测试

```go
// TestLayerManagerBuildRenderPlanes 验证新 API
func TestLayerManagerBuildRenderPlanes(t *testing.T) {
    fiber := createMultiLayerFiber()
    planes := BuildRenderPlanes(fiber) // 通过 LayerManager

    assert.NotNil(t, planes)
    assert.Greater(t, len(planes.Base), 0)
}

// TestStripLayersDeprecated 验证旧 API 已废弃
func TestStripLayersDeprecated(t *testing.T) {
    // 这个测试验证旧函数存在但有 Deprecated 标记
    // 可以用反射检查 godoc 注释

    // 旧函数仍然可以使用（向后兼容）
    vnode := createTestVNode()
    _ = StripLayers(vnode) // 不应该报错

    // TODO: 验证有 Deprecated 注释（需要 godoc 工具）
}
```

---

### Phase 4: 废弃 StripLayers 测试

#### 4.1 功能一致性测试

```go
// TestRenderPlanesEqualsStrippedLayers 验证新旧 API 结果一致
func TestRenderPlanesEqualsStrippedLayers(t *testing.T) {
    // 创建一个复杂的 VNode/Fiber 树
    vnode := createComplexVNode()
    fiber := CreateFiber(vnode)
    layoutFiber(fiber)

    // 使用新 API
    planes := BuildRenderPlanes(fiber)

    // 使用旧 API（暂时保留）
    strippedVNodes := StripLayers(vnode)
    for _, stripped := range strippedVNodes {
        // Layout stripped VNodes
        // 获取旧的 ComputedBoxes
    }

    // 验证结果一致
    // 比较所有 ComputedBox 的位置、大小
}

// TestOldAPIBackwardCompatibility 验证旧 API 向后兼容
func TestOldAPIBackwardCompatibility(t *testing.T) {
    tests := []struct {
        name  string
        vnode *VNode
    }{
        {"simple", createTestVNode()},
        {"modal", createModalVNode()},
        {"overlay", createOverlayVNode()},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // 旧 API 应该仍然可以工作
            result := StripLayers(tt.vnode)
            assert.NotNil(t, result)
        })
    }
}
```

---

### Phase 5: Render 更新测试

#### 5.1 FiberRenderer.Render() 测试

```go
// TestFiberRendererRenderOrder 验证渲染顺序
func TestFiberRendererRenderOrder(t *testing.T) {
    fiber := createMultiLayerFiber()
    planes := BuildRenderPlanes(fiber)

    renderer := NewFiberRenderer()
    output := renderer.Render(planes)

    // 验证输出中 Base 在前，Inspector 在后
    // 可以通过检查输出中的特定文本顺序验证
    assert.True(t, strings.HasPrefix(output, "Base"))
    assert.True(t, strings.Contains(output, "Overlay"))
    assert.True(t, strings.Contains(output, "Modal"))
    assert.True(t, strings.Contains(output, "Inspector"))
}

// TestFiberRendererRenderComputedBox 验证 renderComputedBox
func TestFiberRendererRenderComputedBox(t *testing.T) {
    box := &compute.ComputedBox{
        NodeID:   123,
        Layer:    LayerBase,
        BoxModel: createTestBoxModel(),
        Content:  "Test Content",
        Style:    createTestStyle(),
    }

    renderer := NewFiberRenderer()
    output := renderer.renderComputedBox(box)

    assert.Contains(t, output, "Test Content")
}

// BenchmarkFiberRendererRender 性能测试
func BenchmarkFiberRendererRender(b *testing.B) {
    sizes := []int{10, 100, 1000}
    for _, size := range sizes {
        b.Run(fmt.Sprintf("Size_%d", size), func(b *testing.B) {
            fiber := createLargeFiberWithComputedBox(size)
            planes := BuildRenderPlanes(fiber)
            renderer := NewFiberRenderer()

            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                _ = renderer.Render(planes)
            }
        })
    }
}
```

#### 5.2 视觉测试

```go
// TestVisualRenderingBaseLayer Base 层视觉测试
func TestVisualRenderingBaseLayer(t *testing.T) {
    app := setupTestApp(BaseOnlyComponent)
    defer app.Close()

    app.ForceRender()

    buffer := app.GetBuffer()
    actual := buffer.String()

    expected := loadGoldenFile("testdata/base_layer.golden")
    assert.Equal(t, expected, actual)
}

// TestVisualRenderingModalOverlay Modal/Overlay 视觉测试
func TestVisualRenderingModalOverlay(t *testing.T) {
    app := setupTestApp(ModalComponent)
    defer app.Close()

    app.ForceRender()
    // 打开 Modal
    app.InjectMouse(40, 12, platform.MouseButtonLeft, platform.MouseActionPress)

    buffer := app.GetBuffer()
    actual := buffer.String()

    expected := loadGoldenFile("testdata/modal_open.golden")
    assert.Equal(t, expected, actual)
}

// 如果golden文件不存在，自动生成（仅开发时使用）
func loadGoldenFile(filename string) string {
    data, err := os.ReadFile(filename)
    if err != nil {
        if os.IsNotExist(err) {
            // 开发时可以自动生成 golden 文件
            // os.WriteFile(filename, []byte(expected), 0644)
        }
        return ""
    }
    return string(data)
}
```

---

### Phase 6: HitMap 更新测试

#### 6.1 HitTest Z-ordering 测试

```go
// TestHitTestZOrder 验证 Z-order 优先返回上层元素
func TestHitTestZOrder(t *testing.T) {
    fiber := createOverlappingFibers() // 创建重叠的 Base 和 Modal
    hitMap := BuildHitMapFromFiber(fiber)

    // 测试点在 overlap 区域
    x, y := 10, 10
    entry := hitMap.HitTest(x, y)

    // 应该返回 Modal（上层）
    if entry.Layer != LayerModal {
        t.Errorf("Expected Modal, got Layer %v", entry.Layer)
    }
}

// TestHitTestBaseLayer 验证 Base 层可以点击
func TestHitTestBaseLayer(t *testing.T) {
    fiber := createBaseOnlyFiber()
    hitMap := BuildHitMapFromFiber(fiber)

    x, y := 10, 10
    entry := hitMap.HitTest(x, y)

    assert.Equal(t, LayerBase, entry.Layer)
}

// TestHitTestModalBlocksBase 验证 Modal 阻挡 Base 点击
func TestHitTestModalBlocksBase(t *testing.T) {
    fiber := createModalOverBaseFiber() // Modal 覆盖 Base
    hitMap := BuildHitMapFromFiber(fiber)

    // 点击 overlap 区域
    x, y := 10, 10
    entry := hitMap.HitTest(x, y)

    // 应该返回 Modal，而不是 Base
    assert.Equal(t, LayerModal, entry.Layer)
}

// BenchmarkHitMapHitTest 性能测试
func BenchmarkHitMapHitTest(b *testing.B) {
    sizes := []int{100, 1000, 5000}
    for _, size := range sizes {
        b.Run(fmt.Sprintf("Size_%d", size), func(b *testing.B) {
            fiber := createLargeFiberWithComputedBox(size)
            hitMap := BuildHitMapFromFiber(fiber)

            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                _ = hitMap.HitTest(100, 100)
            }
        })
    }
}
```

#### 6.2 事件冒泡测试

```go
// TestEventBubblingAcrossLayers 验证跨层事件不冒泡
func TestEventBubblingAcrossLayers(t *testing.T) {
    app, err := ui.TestRun(ModalComponent, ui.TestWithSize(80, 24))
    if err != nil {
        t.Fatal(err)
    }
    defer app.Close()

    app.ForceRender()

    // 打开 Modal
    app.InjectMouse(40, 12, platform.MouseButtonLeft, platform.MouseActionPress)

    // 点击 Modal 内部
    app.InjectMouse(40, 12, platform.MouseButtonLeft, platform.MouseActionPress)

    // 验证：只有 Modal 的点击事件被触发，Base 的不应该被触发
    // 需要组件内部有计数器或日志
}

// TestStopPropagationStopsBubbling 验证 StopPropagation 有效
func TestStopPropagationStopsBubbling(t *testing.T) {
    app, err := ui.TestRun(NestingComponent, ui.TestWithSize(80, 24))
    if err != nil {
        t.Fatal(err)
    }
    defer app.Close()

    app.ForceRender()

    // 点击内部元素（设置 StopPropagation）
    app.InjectMouse(30, 10, platform.MouseButtonLeft, platform.MouseActionPress)

    // 验证：只触发内部元素的事件，不冒泡
}
```

---

### Phase 7: 清理和优化测试

#### 7.1 性能验证

```go
// TestPerformanceBaseline 验证性能不低于基线
func TestPerformanceBaseline(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping benchmark in short mode")
    }

    sizes := []struct {
        name  string
        nodes int
    }{
        {"small", 100},
        {"medium", 1000},
        {"large", 5000},
    }

    baselines := map[string]time.Duration{
        "Layout_small":  1 * time.Millisecond,
        "Layout_medium": 10 * time.Millisecond,
        "Layout_large":  50 * time.Millisecond,
        "Render_small":  5 * time.Millisecond,
        "Render_medium": 20 * time.Millisecond,
        "Render_large":  100 * time.Millisecond,
    }

    for _, size := range sizes {
        t.Run(size.name, func(t *testing.T) {
            fiber := createLargeFiberWithComputedBox(size.nodes)

            // Benchmark Layout
            start := time.Now()
            for i := 0; i < 100; i++ {
                _ = layoutFiber(fiber)
            }
            layoutTime := time.Since(start) / 100

            key := fmt.Sprintf("Layout_%s", size.name)
            baseline := baselines[key]
            if layoutTime > baseline*2 { // 允许 2 倍偏差
                t.Errorf("%s took %v, baseline %v", key, layoutTime, baseline)
            }
        })
    }
}
```

#### 7.2 内存泄漏测试

```go
// TestMemoryLeak 运行长时间检查内存是否泄漏
func TestMemoryLeak(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping memory leak test in short mode")
    }

    var m1, m2 runtime.MemStats
    runtime.ReadMemStats(&m1)

    // 运行 1000 次
    for i := 0; i < 1000; i++ {
        fiber := createLargeFiberWithComputedBox(100)
        _ = layoutFiber(fiber)
        _ = BuildRenderPlanes(fiber)
        _ = BuildHitMapFromFiber(fiber)
    }

    // 强制 GC
    runtime.GC()
    runtime.ReadMemStats(&m2)

    // 内存增长应该很小（允许一些增长，但不能持续）
    memDiff := m2.Alloc - m1.Alloc
    maxAllowedDiff := uint64(10 * 1024 * 1024) // 10MB

    if memDiff > maxAllowedDiff {
        t.Errorf("Memory leak detected: %d bytes (max allowed: %d)", memDiff, maxAllowedDiff)
    }
}

// TestMemoryLeakLongRun 更长时间的测试
func TestMemoryLeakLongRun(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping long run test in short mode")
    }

    var lastAlloc uint64
    iterations := 10000

    for i := 0; i < iterations; i++ {
        fiber := createLargeFiberWithComputedBox(100)
        _ = layoutFiber(fiber)
        _ = BuildRenderPlanes(fiber)
        _ = BuildHitMapFromFiber(fiber)

        if i%1000 == 0 {
            runtime.GC()
            var m runtime.MemStats
            runtime.ReadMemStats(&m)

            if lastAlloc > 0 {
                growth := m.Alloc - lastAlloc
                if growth > 5*1024*1024 { // 每 1000 次增长不应超过 5MB
                    t.Errorf("Memory growing too fast at iteration %d: %d bytes", i, growth)
                }
            }
            lastAlloc = m.Alloc
        }
    }
}
```

---

### Phase 8: 综合测试

#### 8.1 端到端测试

```go
// TestE2EApplicationLifecycle 完整应用生命周期 E2E 测试
func TestE2EApplicationLifecycle(t *testing.T) {
    app, err := ui.TestRun(ComplexApplication, ui.TestWithSize(80, 24))
    if err != nil {
        t.Fatal(err)
    }
    defer app.Close()

    // 1. 初始渲染
    app.ForceRender()
    if err := app.AssertRender("Welcome"); err != nil {
        t.Error(err)
    }

    // 2. 用户交互
    app.InjectKey('n') // 导航到下一页
    app.ForceRender()
    if err := app.AssertRender("Page 2"); err != nil {
        t.Error(err)
    }

    // 3. 打开 Modal
    app.InjectMouse(40, 12, platform.MouseButtonLeft, platform.MouseActionPress)
    app.ForceRender()
    if err := app.AssertRender("Modal"); err != nil {
        t.Error(err)
    }

    // 4. 关闭 Modal
    app.InjectKey(platform.KeyEscape)
    app.ForceRender()
    if err := app.AssertNotRender("Modal"); err != nil {
        t.Error(err)
    }

    // 5. 验证状态保持
    if err := app.AssertRender("Page 2"); err != nil {
        t.Error(err)
    }
}

// TestE2EModalInteraction Modal 交互 E2E 测试
func TestE2EModalInteraction(t *testing.T) {
    app, err := ui.TestRun(ModalComponent, ui.TestWithSize(80, 24))
    if err != nil {
        t.Fatal(err)
    }
    defer app.Close()

    app.ForceRender()

    // 打开 Modal
    app.InjectMouse(40, 12, platform.MouseButtonLeft, platform.MouseActionPress)
    app.ForceRender()

    // 点击 Modal 内部的按钮
    app.InjectMouse(30, 8, platform.MouseButtonLeft, platform.MouseActionPress)

    // 验证按钮被点击
    if err := app.AssertRender("Button Clicked"); err != nil {
        t.Error(err)
    }

    // 点击 Modal 外部区域（应该不触发外部元素）
    app.InjectMouse(5, 5, platform.MouseButtonLeft, platform.MouseActionPress)

    // 验证 Modal 仍然打开
    if err := app.AssertRender("Modal"); err != nil {
        t.Error(err)
    }
}
```

#### 8.2 用户场景测试

**手动测试清单**

| 示例 | 测试场景 | 操作步骤 | 预期结果 | 状态 |
|------|---------|---------|---------|------|
| Counter | 增加减计数器 | 按 + 键 5 次，- 键 3 次 | 显示 2 | ⬜ |
| Timer | 定时器启动/停止 | 按 S 启动，等待 2 秒，按 S 停止 | 显示 2+ 秒 | ⬜ |
| Input | 文本输入 | 输入 "Hello World"，Backspace 5 次 | 显示 "Hello" | ⬜ |
| Modal | Modal 弹出/关闭 | 按 Enter 打开，Esc 关闭 | 正确显示/隐藏 | ⬜ |
| Overlay | Overlay 下拉 | 按 Tab 聚焦，Enter 展开 | 显示下拉列表 | ⬜ |
| Progress | 进度条动画 | 启动任务，观察进度 | 平滑动画 | ⬜ |
| Table | 表格滚动 | 按 Down/Up 键 | 正确滚动 | ⬜ |

**测试脚本示例**

```bash
#!/bin/bash
# run_manual_tests.sh

cd examples/counter
echo "Testing Counter..."
go run . &
PID=$!
sleep 2
# 使用 tmux send-keys 或其他工具模拟按键
kill $PID

cd ../modal
echo "Testing Modal..."
go run . &
PID=$!
sleep 2
# ...
kill $PID
```

#### 8.3 压力测试

```go
// TestStressDeepHierarchy 深层级组件树压力测试
func TestStressDeepHierarchy(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping stress test in short mode")
    }

    // 创建 100 层深的组件树
    fiber := createDeepFiber(100)

    start := time.Now()
    _ = layoutFiber(fiber)
    _ = BuildRenderPlanes(fiber)
    _ = BuildHitMapFromFiber(fiber)
    duration := time.Since(start)

    t.Logf("Deep hierarchy (100 levels) took %v", duration)
    assert.Less(t, duration, 5*time.Second, "Should complete in reasonable time")
}

// TestStressWideBreadth 宽节点数压力测试
func TestStressWideBreadth(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping stress test in short mode")
    }

    // 创建 1000 个子节点
    fiber := createWideFiber(1000)

    start := time.Now()
    _ = layoutFiber(fiber)
    _ = BuildRenderPlanes(fiber)
    _ = BuildHitMapFromFiber(fiber)
    duration := time.Since(start)

    t.Logf("Wide breadth (1000 children) took %v", duration)
    assert.Less(t, duration, 5*time.Second, "Should complete in reasonable time")
}

// TestStressRapidRendering 快速连续渲染压力测试
func TestStressRapidRendering(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping stress test in short mode")
    }

    fiber := createLargeFiberWithComputedBox(500)

    start := time.Now()
    for i := 0; i < 1000; i++ {
        _ = layoutFiber(fiber)
        _ = BuildRenderPlanes(fiber)
        _ = BuildHitMapFromFiber(fiber)
    }
    duration := time.Since(start)

    t.Logf("1000 rapid renders took %v", duration)
    avgTime := duration / 1000
    assert.Less(t, avgTime, 10*time.Millisecond, "Average render should be fast")
}

// TestStressEventDispatching 大量事件注入压力测试
func TestStressEventDispatching(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping stress test in short mode")
    }

    app, err := ui.TestRun(SimpleComponent, ui.TestWithSize(80, 24))
    if err != nil {
        t.Fatal(err)
    }
    defer app.Close()

    app.ForceRender()

    start := time.Now()
    for i := 0; i < 10000; i++ {
        app.InjectMouse(10, 10, platform.MouseButtonLeft, platform.MouseActionPress)
    }
    app.ForceRender()
    duration := time.Since(start)

    t.Logf("10000 events took %v", duration)
    assert.Less(t, duration, 10*time.Second, "Should handle events efficiently")
}
```

---

## 自动化测试编写指南

### 单元测试编写规范

#### 命名规范

```go
// 函数命名：Test + 函数名 + 描述
func TestCreateFiberLayerCopy(t *testing.T) { ... }
func TestLayoutFiberEmpty(t *testing.T) { ... }

// 子测试命名使用 t.Run() 的 name 参数
t.Run("with base layer", func(t *testing.T) { ... })
t.Run("with overlay layer", func(t *testing.T) { ... })

// Benchmark 命名：Benchmark + 函数名
func BenchmarkCreateFiber(b *testing.B) { ... }
```

#### 表驱动测试模板

```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name      string
        input     InputType
        want      OutputType
        wantErr   bool
        errMsg    string
    }{
        {
            name:  "normal case",
            input: normalInput,
            want:  expectedOutput,
        },
        {
            name:    "error case",
            input:   errorInput,
            wantErr: true,
            errMsg:  "expected error message",
        },
        // 更多测试用例...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FunctionName(tt.input)

            if tt.wantErr {
                assert.Error(t, err)
                if tt.errMsg != "" {
                    assert.Contains(t, err.Error(), tt.errMsg)
                }
                return
            }

            assert.NoError(t, err)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

#### 辅助函数模式

```go
// 创建测试数据的辅助函数
func createTestVNode() *VNode {
    return &VNode{
        Type:     VNodeElement,
        DiffKey:  "test-vnode",
        Children: nil,
    }
}

func createTestVNodeWithLayer(layer Layer) *VNode {
    vnode := createTestVNode()
    vnode.SetLayer(layer)
    return vnode
}

func createTestFiber(nodes int, layer Layer) *Fiber {
    root := &Fiber{
        NodeID: generateNodeID(),
        Layer:  layer,
    }

    current := root
    for i := 1; i < nodes; i++ {
        child := &Fiber{
            NodeID: generateNodeID(),
            Layer:  layer,
        }
        current.FirstChild = child
        current = child
    }

    return root
}

func createLargeFiberWithComputedBox(size int) *Fiber {
    fiber := createTestFiber(size, LayerBase)

    // 为所有节点添加 ComputedBox
    fillComputedBoxes(fiber)

    return fiber
}

func fillComputedBoxes(fiber *Fiber) {
    fiber.ComputedBox = &compute.ComputedBox{
        NodeID:   fiber.NodeID,
        Layer:    fiber.Layer,
        BoxModel: createTestBoxModel(),
        Content:  "Test",
    }

    child := fiber.FirstChild
    for child != nil {
        fillComputedBoxes(child)
        child = child.Sibling
    }
}
```

### 集成测试编写规范

#### TestRunner 模板

```go
func TestComponentIntegration(t *testing.T) {
    // 1. 设置
    app, err := ui.TestRun(MyComponent,
        ui.TestWithSize(80, 24),
        ui.WithTimeout(10*time.Second),
    )
    if err != nil {
        t.Fatal(err)
    }
    defer app.Close()

    // 2. 强制渲染
    app.ForceRender()

    // 3. 初始断言
    if err := app.AssertRender("Expected Text"); err != nil {
        t.Error(err)
    }

    // 4. 注入事件
    app.InjectKey('a')
    app.InjectSpecialKey(platform.KeyEnter)
    app.InjectMouse(10, 10, platform.MouseButtonLeft, platform.MouseActionPress)

    // 5. 断言结果
    app.ForceRender()
    if err := app.AssertRender("Changed Text"); err != nil {
        t.Error(err)
    }

    // 6. 获取内部状态
    ctx := app.GetContext()
    // ... 验证状态

    // 7. 保存缓冲区（用于调试）
    if t.Failed() {
        app.SaveBufferToFile("failure_buffer.txt")
        t.Logf("Buffer saved to failure_buffer.txt")
    }
}
```

### Mock 使用规范

```go
// 使用 Sandbox mock 测试事件处理
func TestEventHandlingWithMock(t *testing.T) {
    sb := mock.New(80, 24)
    if err := sb.Initialize(nil); err != nil {
        t.Fatal(err)
    }
    defer sb.Close()

    // Mock 时钟
    mockClock := &mocks.Clock{}
    ctx := WithClock(context.Background(), mockClock)

    // 运行应用
    app := NewAppWithMockSandbox(sb)

    // 注入事件
    sb.InjectKey('a')
    sb.InjectSpecialKey(platform.KeyEnter)
    sb.ProcessEvents()

    // 验证
    assert.Equal(t, 1, app.GetKeyPressCount())
}
```

---

## 手动测试流程

### 测试准备

```bash
# 1. 准备测试环境
export TEST_ENV=manual
export MINT_USE_FIBER=true

# 2. 编译所有示例
cd examples
for dir in */; do
    echo "Building $dir..."
    cd "$dir"
    go build .
    cd ..
done

# 3. 准备测试记录表格
# 使用 Google Sheets 或本地文件记录测试结果
```

### Modal 手动测试流程

```markdown
## Modal 手动测试清单

### 准备工作
- [ ] 进入 `examples/modal` 目录
- [ ] 运行 `go run .`
- [ ] 等待应用启动

### 测试场景 1: Modal 打开
步骤:
1. 按 `Enter` 键
2. 观察：Modal 弹框应该出现

预期:
- [ ] Modal 弹框出现
- [ ] Modal 覆盖在其他内容之上（Z-order 正确）
- [ ] Modal 下方内容不可见
- [ ] Modal 标题和内容显示正确

### 测试场景 2: Modal 关闭
步骤:
1. 按 `Esc` 键
2. 观察：Modal 应该消失

预期:
- [ ] Modal 消失
- [ ] 下方内容恢复显示
- [ ] Modal 之前的状态保持

### 测试场景 3: Modal 内部按钮点击
步骤:
1. 打开 Modal
2. 鼠标点击 Modal 内部的 "OK" 按钮
3. 观察：按钮点击效果

预期:
- [ ] 按钮有视觉反馈（高亮或按下状态）
- [ ] 按钮功能正常执行
- [ ] Modal 保持打开状态（除非按钮是关闭按钮）

### 测试场景 4: Modal 外部点击
步骤:
1. 打开 Modal
2. 鼠标点击 Modal 外部区域
3. 观察：是否触发外部元素

预期:
- [ ] 外部元素不触发（Modal 阻挡）
- [ ] Modal 保持打开状态

### 测试场景 5: 多层Modal（如果有）
步骤:
1. 打开 Modal 1
2. 在 Modal 1 中打开 Modal 2
3. 点击 Modal 2 内部
4. 观察：哪个 Modal 响应

预期:
- [ ] Modal 2 响应点击
- [ ] Modal 1 不响应
- [ ] Modal 2 在 Modal 1 之上（Z-order）

### 记录结果
- 总测试项: 20
- 通过: [数量]
- 失败: [数量]
- 失败详情: [描述]
```

### 键盘导航手动测试流程

```markdown
## 键盘导航测试清单

### 准备工作
- [ ] 运行一个包含多个焦点的组件（Button, Input等）
- [ ] 确认焦点指示器可见

### 测试场景 1: Tab 前进
步骤:
1. 按 `Tab` 键 5 次
2. 观察：焦点移动路径

预期:
- [ ] 焦点按正确顺序移动
- [ ] 聚焦到每个可聚焦元素
- [ ] 循环到开头

### 测试场景 2: Shift+Tab 后退
步骤:
1. 按 `Shift+Tab` 5 次
2. 观察：焦点移动路径

预期:
- [ ] 焦点反向移动
- [ ] 正确返回上一个元素

### 测试场景 3: 箭头键导航
步骤:
1. 使用上下左右箭头键
2. 观察：空间导航

预期:
- [ ] 箭头键按空间位置导航
- [ ] 跳转到最近的元素

### 测试场景 4: Enter/Space 激活
步骤:
1. 聚焦到 Button
2. 按 `Enter` 或 `Space`
3. 观察：按钮是否激活

预期:
- [ ] 按钮被点击
- [ ] 触发相应操作

### 记录结果
- Tab 顺序正确: [是/否]
- Shift+Tab 正确: [是/否]
- 箭头键导航: [是/否]
- 激活键工作: [是/否]
```

---

## 性能基准和阈值

### 性能基线测量

```go
// Benchmark helper
func measureBaseline() {
    sizes := []int{10, 50, 100, 500, 1000}
    for _, size := range sizes {
        fmt.Printf("Size: %d\n", size)

        fiber := createLargeFiberWithComputedBox(size)

        // Layout
        fmt.Printf("  Layout: ")
        res1 := testing.Benchmark(func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                _ = layoutFiber(fiber)
            }
        })
        fmt.Printf("%v\n", res1)

        // RenderPlanes
        fmt.Printf("  RenderPlanes: ")
        res2 := testing.Benchmark(func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                _ = BuildRenderPlanes(fiber)
            }
        })
        fmt.Printf("%v\n", res2)

        // HitMap
        fmt.Printf("  HitMap: ")
        res3 := testing.Benchmark(func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                _ = BuildHitMapFromFiber(fiber)
            }
        })
        fmt.Printf("%v\n", res3)
    }
}
```

### 性能阈值

| 指标 | 节点数 | 阈值 | 测量命令 |
|------|--------|------|---------|
| Layout | 100 | < 1ms | `go test -bench=BenchmarkLayoutFiber -benchtime=1x` |
| Layout | 1000 | < 10ms | `go test -bench=BenchmarkLayoutFiber -benchtime=1x` |
| Layout | 5000 | < 50ms | `go test -bench=BenchmarkLayoutFiber -benchtime=1x` |
| RenderPlanes | 100 | < 1ms | `go test -bench=BenchmarkBuildRenderPlanes -benchtime=1x` |
| RenderPlanes | 1000 | < 5ms | `go test -bench=BenchmarkBuildRenderPlanes -benchtime=1x` |
| HitMap 构建 | 100 | < 1ms | `go test -bench=BenchmarkBuildHitMapFromFiber -benchtime=1x` |
| HitMap 构建 | 1000 | < 5ms | `go test -bench=BenchmarkBuildHitMapFromFiber -benchtime=1x` |
| HitTest | 任意 | < 1ms | `go test -bench=BenchmarkHitMapHitTest -benchtime=1x` |
| 帧渲染 | 1000节点 | < 16ms (60fps) | 手动测试 + FPS 计数 |
| 内存占用 | 空闲 | < 10MB | `pprof` |
| 内存占用 | 1000节点 | < 50MB | `pprof` |

### 性能退化检测

```go
// TestPerformanceRegression 检测性能退化
func TestPerformanceRegression(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping regression test in short mode")
    }

    // 这些值应该从文件或 CI 中读取
    baselines := map[string]float64{
        "Layout_1000":    10.0,  // ms
        "Render_1000":    20.0,  // ms
        "HitMap_1000":    5.0,   // ms
        "RenderPlanes_1000": 5.0, // ms
    }

    fiber := createLargeFiberWithComputedBox(1000)

    // Layout
    start := time.Now()
    for i := 0; i < 100; i++ {
        _ = layoutFiber(fiber)
    }
    layoutTime := float64(time.Since(start)) / 100.0 / float64(time.Millisecond)

    t.Logf("Layout (1000 nodes): %.2f ms", layoutTime)
    if layoutTime > baselines["Layout_1000"]*1.2 { // 允许 20% 偏差
        t.Errorf("Layout degraded: %.2f > %.2f", layoutTime, baselines["Layout_1000"]*1.2)
    }

    // ... 其他指标
}
```

---

## 回归测试方案

### 回归测试策略

```go
// TestRegressionFiberNodeIdentity 验证 Fiber NodeID 不参与 diff
func TestRegressionFiberNodeIdentity(t *testing.T) {
    // 这是旧版本的 bug：NodeID 被用于 diff 匹配
    // 重构后应该使用 DiffKey 而非 NodeID

    oldVNode := createTestVNode()
    oldFiber := CreateFiber(oldVNode)
    oldFiber.NodeID = 123

    // 创建一个 DiffKey 相同但 NodeID 不同的新 VNode
    newVNode := createTestVNode()
    newFiber, shouldUpdate := shouldUpdate(oldFiber, newVNode)

    // 应该不更新（DiffKey 相同）
    assert.False(t, shouldUpdate, "shouldUpdate should use DiffKey, not NodeID")
}
```

### 对比测试套件

```bash
#!/bin/bash
# run_comparison_tests.sh

# 运行重构后版本的测试
echo "Running new version tests..."
go test ./... -cover > new_results.txt
cp coverage.out new_coverage.out

# 切换到旧版本（假设有 branch）
git checkout old_branch

# 运行旧版本测试
echo "Running old version tests..."
go test ./... -cover > old_results.txt
cp coverage.out old_coverage.out

# 对比结果
echo "Comparing results..."
diff old_results.txt new_results.txt

# 对比覆盖率
echo "Comparing coverage..."
go tool cover -func=old_coverage.out > old_func.txt
go tool cover -func=new_coverage.out > new_func.txt
diff old_func.txt new_func.txt

# 切换回新版本
git checkout new_branch
```

### 视觉回归测试

```go
// TestVisualRegression 视觉回归测试
func TestVisualRegression(t *testing.T) {
    app, err := ui.TestRun(StandardComponent, ui.TestWithSize(80, 24))
    if err != nil {
        t.Fatal(err)
    }
    defer app.Close()

    app.ForceRender()
    buffer := app.GetBuffer()

    // 保存当前输出
    actual := buffer.String()

    // 加载旧版本的 Golden 文件
    goldenPath := "testdata/regression/golden_standard_component.txt"
    golden, err := os.ReadFile(goldenPath)
    if err != nil {
        if os.IsNotExist(err) {
            // 首次运行，创建 Golden 文件
            os.WriteFile(goldenPath, []byte(actual), 0644)
            t.Skipf("Golden file created at %s", goldenPath)
        } else {
            t.Fatal(err)
        }
    }

    // 对比
    if string(golden) != actual {
        // 输出差异
        diff := difflib.UnifiedDiff{
            A:       difflib.SplitLines(string(golden)),
            B:       difflib.SplitLines(actual),
            Context: 3,
        }
        diffText, _ := difflib.GetUnifiedDiffString(diff)
        t.Errorf("Visual regression detected:\n%s", diffText)

        // 保存失败输出用于对比
        os.WriteFile("testdata/regression/failed_output.txt", []byte(actual), 0644)
    }
}
```

---

## 测试数据准备

### 测试数据集

```go
// testdata/vnodes.go - 测试 VNode 数据
var (
    // 简单单节点
    SimpleVNode = &VNode{
        Type:    VNodeElement,
        DiffKey: "simple",
        Layer:   LayerBase,
    }

    // 组件树
    ComponentTree = &VNode{
        Type:    VNodeElement,
        DiffKey: "root",
        Layer:   LayerBase,
        Children: []*VNode{
            {
                Type:    VNodeElement,
                DiffKey: "child1",
                Layer:   LayerBase,
            },
            {
                Type:    VNodeElement,
                DiffKey: "child2",
                Layer:   LayerOverlay,
            },
        },
    }

    // Modal 组件
    ModalVNode = &VNode{
        Type:    VNodeElement,
        DiffKey: "modal",
        Layer:   LayerModal,
        Children: []*VNode{
            {
                Type:    VNodeElement,
                DiffKey: "modal-content",
                Layer:   LayerModal,
            },
        },
    }
)

// testdata/fibers.go - 测试 Fiber 数据
func DeepFiber(depth int) *Fiber {
    root := &Fiber{
        NodeID:  generateNodeID(),
        Layer:   LayerBase,
        DiffKey: "root",
    }

    current := root
    for i := 1; i < depth; i++ {
        child := &Fiber{
            NodeID:  generateNodeID(),
            Layer:   LayerBase,
            DiffKey: fmt.Sprintf("level%d", i),
        }
        current.FirstChild = child
        current = child
    }

    return root
}

func WideFiberbreadth int) *Fiber {
    root := &Fiber{
        NodeID:  generateNodeID(),
        Layer:   LayerBase,
        DiffKey: "root",
    }

    var lastSibling *Fiber
    for i := 0; i < breadth; i++ {
        child := &Fiber{
            NodeID:  generateNodeID(),
            Layer:   LayerBase,
            DiffKey: fmt.Sprintf("child%d", i),
        }

        if lastSibling == nil {
            root.FirstChild = child
        } else {
            lastSibling.Sibling = child
        }
        lastSibling = child
    }

    return root
}
```

### Golden 文件

```
testdata/
├── golden/
│   ├── base_layer.txt          # Base 层渲染输出
│   ├── modal_open.txt          # Modal 打开状态
│   ├── overlay_expanded.txt    # Overlay 展开状态
│   └── table_scrolled.txt      # Table 滚动状态
├── regression/
│   ├── golden_standard_component.txt  # 标准组件 Golden
│   └── golden_multi_layer.txt         # 多层级 Golden
└── scenarios/
    ├── user_login_flow.txt
    └── data_entry_flow.txt
```

---

## 缺陷管理流程

### 缺陷分类

| 严重度 | 定义 | 响应时间 | 示例 |
|--------|------|---------|------|
| **Critical** | 阻塞性问题，无法继续 | 立即 | 编译失败、核心功能失效 |
| **High** | 主要功能受损 | 4小时 | Modal 不显示、事件不响应 |
| **Medium** | 次要功能受损 | 1天 | 某些样式错误、性能退化 |
| **Low** | 文字错误、小问题 | 3天 | 拼写错误、注释不完整 |

### 缺陷报告模板

```markdown
## 缺陷报告: [标题]

### 基本信息
- **标题**: [简短描述]
- **严重度**: Critical/High/Medium/Low
- **优先级**: P0/P1/P2/P3
- **Phase**: Phase 1-8
- **发现时间**: [日期时间]
- **发现人**: [姓名]

### 描述
[详细描述问题]

### 复现步骤
1. [操作步骤 1]
2. [操作步骤 2]
3. [观察结果]

### 预期行为
[应该发生什么]

### 实际行为
[实际发生什么]

### 环境信息
- **操作系统**: [OS]
- **Go 版本**: [版本]
- **Git 分支**: [分支]
- **Commit**: [Hash]

### 附件
- [ ] 截图
- [ ] 日志
- [ ] 测试代码

### 根因分析
[问题根本原因]

### 修复建议
[建议的修复方案]

### 状态
- [ ] 待修复
- [ ] 修复中
- [ ] 已修复
- [ ] 待验证
- [ ] 已关闭
```

### 缺陷跟踪工具

推荐使用 GitHub Issues：

```yaml
# Issue 标签
labels:
  - bug: 缺陷
  - phase1: Phase 1 相关
  - phase2: Phase 2 相关
  - phase3: Phase 3 相关
  - performance: 性能问题
  - regression: 回归问题
  - blocking: 阻塞性问题

# Issue 模板
bug_template: |
  **Bug Description**
  [描述]

  **Steps to Reproduce**
  1.
  2.
  3.

  **Expected Behavior**

  **Actual Behavior**

  **Environment**
  - OS:
  - Go Version:
  - Branch:
```

---

## 附录

### A. CI/CD 集成

```yaml
# .github/workflows/test.yml

name: Test

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

jobs:
  test:
    runs-on: ubuntu-latest

    strategy:
      matrix:
        go-version: ['1.24', '1.25']

    steps:
    - uses: actions/checkout@v3

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ matrix.go-version }}

    - name: Install dependencies
      run: go mod download

    - name: Run tests
      run: go test -race -coverprofile=coverage.out ./...

    - name: Run benchmarks
      run: go test -bench=. -benchmem ./...

    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        files: ./coverage.out

    - name: Check lint
      uses: golangci/golangci-lint-action@v3
      with:
        version: latest
```

### B. 常见测试反模式

```go
// ❌ 反模式 1: 过于复杂的 setup
func TestComplex(t *testing.T) {
    // 太多 setup 代码
    app := createComplexApp()
    config := createComplexConfig()
    db := setupTestDB()
    // ... 100 行 setup

    // 只有 3 行 actual test
    result := doSomething()
    assert.Equal(t, expected, result)
}

// ✅ 正确: 简洁的测试，把 setup 放到辅助函数
func TestSimple(t *testing.T) {
    app := setupTestApp(t) // 辅助函数

    result := doSomething(app)
    assert.Equal(t, expected, result)
}

// ❌ 反模式 2: 测试实现细节
func TestInternalFunction(t *testing.T) {
    // 直接测试内部函数
    internalFunction()
}

// ✅ 正确: 测试公共行为
func TestPublicBehavior(t *testing.T) {
    result := PublicFunction()
    assert.Equal(t, expected, result)
}

// ❌ 反模式 3: 魔法数字
func TestMagic(t *testing.T) {
    assert.Equal(t, 42, result) // 42 是什么？
}

// ✅ 正确: 使用命名常量
const ExpectedAnswer = 42

func TestNamed(t *testing.T) {
    assert.Equal(t, ExpectedAnswer, result)
}

// ❌ 反模式 4: 测试太多东西
func TestTooManyThings(t *testing.T) {
    // 测试了太多场景
    assert.Equal(t, a, result.A)
    assert.Equal(t, b, result.B)
    assert.Equal(t, c, result.C)
    // ... 10+ 断言
}

// ✅ 正确: 拆分成多个小测试
func TestPropertyA(t *testing.T) {
    assert.Equal(t, expectedA, result.A)
}

func TestPropertyB(t *testing.T) {
    assert.Equal(t, expectedB, result.B)
}
```

### C. 测试工具命令速查

```bash
# 基础测试
go test                          # 运行当前包测试
go test ./...                    # 运行所有包测试
go test -v ./...                 # 详细输出
go test -run TestSpecific ./...  # 运行特定测试

# 覆盖率
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
go tool cover -html=coverage.out

# 基准测试
go test -bench=. ./...
go test -bench=BenchmarkFunc -benchmem ./...
go test -bench=. -benchtime=10s ./...

# 竞态检测
go test -race ./...
go test -race -short ./...

# pprof
go test -cpuprofile=cpu.prof -bench=. ./...
go test -memprofile=mem.prof -bench=. ./...
go tool pprof cpu.prof
go tool pprof mem.prof

# Lint
golangci-lint run ./...
golangci-lint run --timeout=5m ./...
```

---

**文档版本**: 1.0
**最后更新**: 2026-02-14
**维护者**: [待分配]
