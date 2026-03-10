# 单次遍历布局重构实施计划

## 一、背景与问题

### 1.1 当前问题

当前布局系统存在**双重测量**问题：

```
buildComputedBox(vnode, parent, constraints)
│
├─► measureVNode(vnode, constraints)
│   └─► 调用 LayoutNode.Measure() [第一次测量子节点]
│
└─► 对每个子节点:
    ├─► getChildConstraints() [使用不同的约束逻辑!]
    └─► buildComputedBox(child, ...) [第二次测量子节点]
```

**导致的问题**：
1. 约束逻辑在两个地方不一致（`layout.go` 的 `Measure()` 和 `engine.go` 的 `getChildConstraints()`）
2. 需要硬编码特殊处理（如 VStack 中的 HStack 子节点）
3. 难以维护和扩展

### 1.2 重构目标

1. **消除双重测量**：每个节点每轮布局只测量一次
2. **统一约束逻辑**：约束传递逻辑集中在一个地方
3. **清晰的职责分离**：引擎负责树构建，节点负责测量逻辑
4. **保持向后兼容**：渐进式迁移，不破坏现有代码

---

## 二、架构设计

### 2.1 新接口定义

```go
// runtime/compute/measurer.go

// LayoutMeasurer 由想要自定义布局逻辑的节点实现
type LayoutMeasurer interface {
    VNode

    // MeasureLayout 测量此节点并返回布局信息
    // 引擎使用此信息构建 ComputedBox 树
    //
    // 关键: 此方法应该:
    // 1. 测量子节点（使用 engine.MeasureChild）
    // 2. 计算自身尺寸
    // 3. 返回每个子节点使用的约束（供 buildComputedBox 重用）
    MeasureLayout(
        engine *Engine,
        constraints BoxConstraints,
        parent *ComputedBox,
    ) LayoutMeasurement
}

// LayoutMeasurement 包含布局节点的测量结果
type LayoutMeasurement struct {
    // Size 是此节点的最终尺寸
    Size Size

    // ChildConstraints 是传递给每个子节点的约束
    // 必须与 vnode.Children() 一一对应
    ChildConstraints []BoxConstraints

    // ChildSizes 是子节点的测量结果（可选，用于缓存）
    ChildSizes []Size
}
```

### 2.2 新的 Pipeline 流程

```
buildComputedBox(vnode, parent, constraints)
│
├─► 检查 vnode 是否实现 LayoutMeasurer
│   │
│   ├─► 是: 调用 vnode.MeasureLayout(e, constraints, parent)
│   │        ├── 测量所有子节点（在 MeasureLayout 内部）
│   │        ├── 计算自身尺寸
│   │        └── 返回 LayoutMeasurement
│   │             ├── Size
│   │             ├── ChildConstraints[] (已测量的约束)
│   │             └── ChildSizes[] (可选)
│   │
│   └─► 否: 使用旧的两遍方法
│        ├── measureVNode() 测量自身
│        └── getChildConstraints() 计算子节点约束
│
└─► 构建子节点的 ComputedBox
    └─► 使用 MeasureLayout 返回的 ChildConstraints
```

---

## 三、实施步骤

### 阶段一：基础设施（第 1-2 天）

#### 任务 1.1：创建新接口和类型

**文件**：`runtime/compute/measurer.go`（新建）

```go
package compute

import "github.com/wwsheng009/mint/runtime"

// LayoutMeasurer 接口定义
type LayoutMeasurer interface {
    VNode
    MeasureLayout(engine *Engine, constraints runtime.BoxConstraints, parent *ComputedBox) LayoutMeasurement
}

// LayoutMeasurement 结构定义
type LayoutMeasurement struct {
    Size            runtime.Size
    ChildConstraints []runtime.BoxConstraints
    ChildSizes      []runtime.Size
}

// NewLayoutMeasurement 创建测量结果
func NewLayoutMeasurement(size runtime.Size, childConstraints []runtime.BoxConstraints) LayoutMeasurement {
    return LayoutMeasurement{
        Size:            size,
        ChildConstraints: childConstraints,
        ChildSizes:      make([]runtime.Size, len(childConstraints)),
    }
}
```

**验收标准**：
- [ ] 文件创建成功
- [ ] 代码编译通过
- [ ] 添加基础测试

#### 任务 1.2：添加 MeasureChild 辅助方法

**文件**：`runtime/compute/engine.go`

```go
// MeasureChild 测量单个子节点，供 LayoutMeasurer 实现使用
func (e *Engine) MeasureChild(child VNode, constraints runtime.BoxConstraints) runtime.Size {
    return e.measureVNode(child, constraints)
}
```

**验收标准**：
- [ ] 方法添加成功
- [ ] 测试验证功能正确

---

### 阶段二：LayoutNode 实现（第 3-5 天）

#### 任务 2.1：实现 LayoutNode.MeasureLayout

**文件**：`runtime/ui/layout.go`

```go
// MeasureLayout 实现 LayoutMeasurer 接口
func (l *LayoutNode) MeasureLayout(
    engine *compute.Engine,
    constraints runtime.BoxConstraints,
    parent *compute.ComputedBox,
) compute.LayoutMeasurement {
    if l.direction == DirectionRow {
        return l.measureHStackLayout(engine, constraints)
    }
    return l.measureVStackLayout(engine, constraints)
}

func (l *LayoutNode) measureHStackLayout(
    engine *compute.Engine,
    constraints runtime.BoxConstraints,
) compute.LayoutMeasurement {
    // 1. 获取布局属性
    layoutInfo := GetLayoutInfo(l)
    gap := layoutInfo.Gap
    padding := layoutInfo.Padding
    // ... paddingWidth, paddingHeight 计算

    // 2. 计算内部约束
    innerMaxHeight := runtime.Infinity
    if constraints.HasBoundedHeight() {
        innerMaxHeight = max(0, constraints.MaxHeight-paddingHeight)
    }

    // 3. 第一遍：识别 flex 子节点，测量非 flex 子节点
    children := l.Children()
    childConstraints := make([]runtime.BoxConstraints, len(children))
    childSizes := make([]runtime.Size, len(children))

    var flexChildren []flexChildInfo
    var fixedWidth int
    var maxHeight int
    flexTotalFactor := 0

    for i, child := range children {
        childInfo := GetLayoutInfo(child)
        if childInfo.Flex > 0 {
            flexChildren = append(flexChildren, flexChildInfo{
                child:  child,
                index:  i,
                factor: childInfo.Flex,
            })
            flexTotalFactor += childInfo.Flex
        } else {
            // 测量非 flex 子节点
            childConstraints[i] = runtime.BoxConstraints{
                MinWidth:  0,
                MaxWidth:  runtime.Infinity,
                MinHeight: 0,
                MaxHeight: innerMaxHeight,
            }
            childSizes[i] = engine.MeasureChild(child, childConstraints[i])
            fixedWidth += childSizes[i].Width
            if childSizes[i].Height > maxHeight {
                maxHeight = childSizes[i].Height
            }
        }
    }

    // 4. Flex 分配
    if len(flexChildren) > 0 && constraints.HasBoundedWidth() {
        availableWidth := constraints.MaxWidth - paddingWidth - (len(children)-1)*gap
        remainingSpace := availableWidth - fixedWidth

        for _, fc := range flexChildren {
            flexWidth := (remainingSpace * fc.factor) / flexTotalFactor
            if flexWidth < 0 {
                flexWidth = 0
            }
            childConstraints[fc.index] = runtime.BoxConstraints{
                MinWidth:  flexWidth,
                MaxWidth:  flexWidth,
                MinHeight: 0,
                MaxHeight: innerMaxHeight,
            }
            childSizes[fc.index] = engine.MeasureChild(fc.child, childConstraints[fc.index])
        }
    }

    // 5. 计算总尺寸
    totalWidth := fixedWidth + paddingWidth
    maxHeight += paddingHeight

    // 应用约束
    if maxHeight < constraints.MinHeight {
        maxHeight = constraints.MinHeight
    }

    return compute.LayoutMeasurement{
        Size:            runtime.Size{Width: totalWidth, Height: maxHeight},
        ChildConstraints: childConstraints,
        ChildSizes:      childSizes,
    }
}
```

**验收标准**：
- [ ] `measureHStackLayout` 实现完成
- [ ] `measureVStackLayout` 实现完成
- [ ] 交叉轴约束处理正确（VStack 中的 HStack 收到 tight 约束）

#### 任务 2.2：VStack 中的 HStack 特殊处理

在 `measureVStackLayout` 中：

```go
// 对于 VStack 中的 HStack 子节点，创建 tight 约束以支持对齐
childMinWidth := 0
isHS := isHStack(child)
if innerMaxWidth != runtime.Infinity && isHS {
    // HStack 在 VStack 中填充宽度以支持主轴对齐
    childMinWidth = innerMaxWidth
}

childConstraints[i] = runtime.BoxConstraints{
    MinWidth:  childMinWidth,
    MaxWidth:  innerMaxWidth,
    MinHeight: 0,
    MaxHeight: runtime.Infinity,
}
```

**验收标准**：
- [ ] HStack 子节点收到正确的 tight 约束
- [ ] Modal 内容居中测试通过

---

### 阶段三：Engine 集成（第 6-7 天）

#### 任务 3.1：更新 buildComputedBox

**文件**：`runtime/compute/engine.go`

```go
func (e *Engine) buildComputedBox(vnode VNode, parent *ComputedBox, constraints runtime.BoxConstraints) *ComputedBox {
    // ... 现有的初始化代码 ...

    // 尝试使用新的单遍测量
    if measurer, ok := vnode.(LayoutMeasurer); ok {
        measurement := measurer.MeasureLayout(e, constraints, parent)

        // 设置自身尺寸
        box.Box.Width = measurement.Size.Width
        box.Box.Height = measurement.Size.Height

        // 使用预计算的约束构建子节点
        for i, child := range vnode.Children() {
            childConstraints := measurement.ChildConstraints[i]
            childBox := e.buildComputedBox(child, box, childConstraints)
            if childBox != nil {
                box.Children = append(box.Children, childBox)
            }
        }
    } else {
        // 回退到旧的两遍方法
        size := e.measureVNode(vnode, constraints)
        box.Box.Width = size.Width
        box.Box.Height = size.Height

        for _, child := range vnode.Children() {
            childConstraints := e.getChildConstraints(vnode, child, constraints, size)
            childBox := e.buildComputedBox(child, box, childConstraints)
            if childBox != nil {
                box.Children = append(box.Children, childBox)
            }
        }
    }

    // ... 现有的缓存代码 ...
    return box
}
```

**验收标准**：
- [ ] 代码编译通过
- [ ] LayoutNode 使用新路径
- [ ] 其他节点回退到旧路径

---

### 阶段四：清理与优化（第 8-10 天）

#### 任务 4.1：移除旧的 Measure 方法

在确认新实现工作正常后：

**文件**：`runtime/ui/layout.go`

```go
// 移除或标记为废弃:
// func (l *LayoutNode) Measure(constraints BoxConstraints) Size { ... }
```

**验收标准**：
- [ ] 所有测试通过
- [ ] 旧方法标记为废弃
- [ ] 文档更新

#### 任务 4.2：简化 getChildConstraints

移除 `getChildConstraints()` 中的特殊处理：

```go
// 移除 VStack case 中的 HStack 特殊处理
// 因为现在约束逻辑在 MeasureLayout 中统一处理
```

**验收标准**：
- [ ] 代码简化
- [ ] 测试验证

#### 任务 4.3：移除 Measurable 接口

最终目标：

```go
// Measurable 接口可以废弃，因为所有测量逻辑现在通过 LayoutMeasurer 处理
```

**验收标准**：
- [ ] 所有节点迁移到 LayoutMeasurer
- [ ] Measurable 接口标记为废弃

---

## 四、测试计划

### 4.1 单元测试

| 测试 | 描述 | 文件 |
|-----|------|------|
| TestLayoutMeasurement | 验证 LayoutMeasurement 结构创建 | runtime/compute/measurer_test.go |
| TestMeasureChild | 验证 MeasureChild 方法 | runtime/compute/engine_test.go |
| TestMeasureHStackLayout | 验证 HStack 布局测量 | runtime/ui/layout_test.go |
| TestMeasureVStackLayout | 验证 VStack 布局测量 | runtime/ui/layout_test.go |
| TestCrossAxisConstraints | 验证交叉轴约束（HStack in VStack） | runtime/ui/layout_test.go |

### 4.2 集成测试

| 测试 | 描述 | 文件 |
|-----|------|------|
| TestModalCentering | Modal 内容居中 | examples/sandbox/demo/modal_center_test.go |
| TestFlexDistribution | Flex 空间分配 | runtime/compute/integration_test.go |
| TestNestedLayouts | 嵌套布局 | runtime/compute/integration_test.go |

### 4.3 回归测试

运行完整的测试套件：

```bash
# 单元测试
go test ./runtime/...

# 集成测试
go test ./examples/sandbox/demo/...

# 性能测试
go test -bench=. ./runtime/...
```

---

## 五、风险评估与缓解

### 风险矩阵

| 风险 | 可能性 | 影响 | 缓解措施 |
|-----|-------|------|---------|
| 引入新的 bug | 中 | 高 | 完善的测试覆盖，渐进式迁移 |
| 性能回归 | 低 | 中 | 基准测试，性能监控 |
| 破坏向后兼容 | 低 | 高 | 保留回退路径，版本迁移 |
| 迁移复杂度高 | 中 | 中 | 详细文档，分阶段实施 |

### 回滚计划

如果重构出现问题：

1. **阶段一回滚**：移除新接口，继续使用旧方法
2. **阶段二回滚**：在 buildComputedBox 中优先使用旧路径
3. **完全回滚**：恢复到重构前的代码

---

## 六、里程碑

| 里程碑 | 日期 | 交付物 |
|--------|------|--------|
| M1: 基础设施完成 | Day 2 | measurer.go, MeasureChild 方法 |
| M2: LayoutNode 实现 | Day 5 | MeasureLayout 实现 |
| M3: Engine 集成 | Day 7 | buildComputedBox 更新 |
| M4: 测试验证 | Day 9 | 所有测试通过 |
| M5: 清理完成 | Day 10 | 移除旧代码 |

---

## 七、成功标准

1. **功能正确性**：所有现有测试通过
2. **架构改进**：消除双重测量，约束逻辑统一
3. **性能**：无性能回归（允许 5% 波动）
4. **代码质量**：代码行数减少，复杂度降低
5. **可维护性**：新的特殊处理更容易添加

---

## 八、后续优化

重构完成后，可以考虑：

1. **缓存优化**：缓存 LayoutMeasurement 结果
2. **增量布局**：只重新测量变化的子树
3. **并行测量**：独立的子树可以并行测量
4. **约束推断**：自动推断最优约束

---

*文档创建日期: 2024-02-06*
*预计完成日期: 2024-02-16*
*负责人: Claude*
