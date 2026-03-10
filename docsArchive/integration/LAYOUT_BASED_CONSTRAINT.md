# Layout-Based Component Sizing (No Manual Heights)

## 最佳实践

### ✅ 推荐：使用 Flex 布局

**让布局系统自动分配空间，无需手动计算高度：**

```go
// ✅ GOOD - 使用 Flex(1)
treePreview := ui.VStackBuilder(treeRender).
    Width(si.overlayWidth - 4).
    Flex(1).  // 布局系统自动分配可用空间
    Build()

// ✅ 父容器通过 Tabs 传递约束
// Tabs(Height=21) → content (MaxHeight=20) → VStack → treePreview(Flex=1)
```

### ❌ 避免：手动计算高度

**不要手动计算和设置固定高度：**

```go
// ❌ BAD - 手动计算高度
fixedHeaderHeight := 3 + 1 + 2 + 4  // title + tabs + stats + selectedInfo
instructionsHeight := 6
treeViewHeight := si.overlayHeight - fixedHeaderHeight - instructionsHeight

treePreview := ui.VStackBuilder(treeRender).
    Width(si.overlayWidth - 4).
    Height(treeViewHeight).  // 手动计算 - 容易出错！
    Build()

// ❌ 还要手动设置虚拟滚动
si.treeViewComponent.SetViewportHeight(treeViewHeight)
```

## 为什么 Flex(1) 更好？

### 1. 自动适应容器大小

```go
Flex(1)  // 自动填充可用空间
```

- ✅ 容器变大 → treePreview 变大
- ✅ 容器变小 → treePreview 变小
- ✅ 无需重新计算
- ✅ 适应不同屏幕尺寸

### 2. 简化代码

**手动计算**（需要维护）：
```go
// 需要知道：
// - titleBar: 3 lines
// - tabBar: 1 line
// - separator: 1 line
// - selectedInfo: 4 lines
// - instructions: 6 lines
// Total: 15 lines
treeViewHeight := overlayHeight - 15  // 魔法数字！
```

**Flex 布局**（自动维护）：
```go
// 无需知道其他组件的高度
treePreview.Flex(1)  // 自动填充剩余空间
```

### 3. 减少错误

**手动计算容易出错**：
- 标题高度改变 → 忘记更新计算
- 添加新组件 → 忘记调整
- 换行数统计 → 计算错误

**Flex 自动适应**：
- 标题高度改变 → 自动调整
- 添加新组件 → 自动调整
- 无需手动维护

## 约束传播链

```
┌─────────────────────────────────────────────────────────────┐
│ Bordered(Height=25)                                         │
│   ↓ measureBordered: innerConstraints.MaxHeight = 23        │
│ ┌───────────────────────────────────────────────────────────┐│
│ │ VStack(content) - bounded height (23)                    ││
│ │   ↓ measureLayoutChildren.VStack                          ││
│ │ ┌───────────────────────────────────────────────────────┐││
│ │ │ Tabs(Height=21 prop)                                   │││
│ │ │   ↓ Tabs.Measure() checks props!                      │││
│ │ │   ↓ contentConstraints.MaxHeight = 21 - 1 = 20       │││
│ │ │ ┌────────────────────────────────────────────────────┐│││
│ │ │ │ VStack (from buildElementsTabContent)              ││││
│ │ │ │   ↓ bounded height (20) ✓                          ││││
│ │ │ │   ↓ Flex distribution triggered!                  ││││
│ │ │ │   ├─ header (natural height: ~3)                   ││││
│ │ │ │   ├─ selectedInfo (natural height: ~4)             ││││
│ │ │ │   ├─ treePreview (Flex=1) → 20 - 3 - 4 - 6 = 7 ✓ ││││
│ │ │ │   └─ instructions (natural height: ~6)             ││││
│ │ │ └────────────────────────────────────────────────────┘│││
│ │ └───────────────────────────────────────────────────────┘││
│ └───────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

**关键点**：
1. ✅ 每个层级正确传递约束
2. ✅ Flex(1) 在有 bounded height 时才工作
3. ✅ 布局引擎自动计算剩余空间
4. ✅ 无需手动计算任何高度

## 虚拟滚动 vs. 简单渲染

### 小规模树（< 100 节点）

```go
// ✅ 推荐：渲染所有节点，让布局系统裁剪
si.treeViewComponent.SetViewportHeight(0)  // 0 = 渲染全部

treePreview := ui.VStackBuilder(treeRender).
    Flex(1).
    Build()
```

**优点**：
- 简单直接
- 无需手动管理 viewport
- 布局系统自动裁剪溢出
- 性能影响可忽略（~32 节点）

### 大规模树（> 1000 节点）

```go
// ⚠️ 如果需要虚拟滚动，使用约束感知组件
// 参考：treeview_constraint_example.go

constrainedTreeView := NewConstrainedTreeView(treeView)
treePreview := ui.VStackBuilder(constrainedTreeView.GetRender()...).
    Flex(1).
    Build()
```

**何时使用**：
- 树节点数 > 1000
- 渲染性能问题
- 需要严格控制内存

## Inspector 的选择

**当前实现**：方案 1（无虚拟滚动）

**原因**：
1. Inspector 的树很小（~32 节点）
2. 渲染所有节点性能可忽略
3. 代码简单，易维护
4. 完全通过 Flex 布局约束

**验证**：
```bash
cd examples/ui_demos/demo2_runtime_internals/inspector_overlay
TUI_INSPECTOR=true go run main.go
```

✅ 正常显示
✅ 无性能问题
✅ 自动适应容器大小

## 迁移指南

### 从手动计算到 Flex 布局

**之前**：
```go
// 手动计算
treeViewHeight := overlayHeight - 15
si.treeViewComponent.SetViewportHeight(treeViewHeight)

treePreview := ui.VStackBuilder(treeRender).
    Height(treeViewHeight).  // 固定高度
    Build()
```

**之后**：
```go
// 不设置 viewportHeight（渲染全部）
// Flex 布局自动分配空间
treePreview := ui.VStackBuilder(treeRender).
    Flex(1).  // 弹性高度
    Build()
```

### 需要修改的地方

1. ✅ 移除手动高度计算
2. ✅ 移除 `SetViewportHeight()` 调用
3. ✅ 使用 `Flex(1)` 替代 `Height(n)`
4. ✅ 确保父容器有 bounded height

### 不需要修改的地方

- ✅ TreeView 组件本身
- ✅ 导航逻辑（上下键、PageUp/Down）
- ✅ 展开/折叠功能
- ✅ 选择逻辑

## 常见问题

### Q: Flex(1) 不生效？

**A**: 检查父容器是否有 bounded height：
```go
// ❌ 父容器无 bounded height
parent := ui.VStack(child)  // MaxHeight = Infinity

// ✅ 父容器有 bounded height
parent := ui.VStackBuilder(child).
    Height(25).  // 或者从 Tabs/Border 传递
    Build()
```

### Q: 内容溢出？

**A**: 检查约束传播链：
1. Border → 传递 bounded height ✓
2. VStack → 接收 bounded height ✓
3. Tabs → 使用 height prop ✓
4. VStack (content) → 接收 bounded height ✓
5. treePreview.Flex(1) → 工作 ✓

### Q: 如何调试？

```bash
# 启用布局调试
TUI_LAYOUT_DEBUG=true go run main.go

# 启用布局警告
TUI_LAYOUT_WARNINGS=true go run main.go

# Inspector 详细日志
TUI_DEBUG_INSPECTOR=true go run main.go
```

## 性能对比

### 手动计算 + 虚拟滚动
```
优点：适合大规模树（>1000 节点）
缺点：
  - 需要手动维护高度计算
  - 代码复杂
  - 容易出错
```

### Flex 布局 + 简单渲染
```
优点：
  - 自动适应容器
  - 代码简单
  - 无需维护
  - 适合小规模树（<100 节点）

缺点：
  - 大规模树（>1000 节点）可能有性能问题
```

## 推荐方案

| 场景 | 节点数 | 推荐方案 |
|------|--------|----------|
| Inspector | ~32 | Flex(1) + 无虚拟滚动 ✅ |
| 中等树 | <100 | Flex(1) + 无虚拟滚动 ✅ |
| 大型树 | 100-1000 | Flex(1) + 条件虚拟滚动 |
| 超大树 | >1000 | Flex(1) + 约束感知虚拟滚动 |

对于 **Inspector**，当前方案是最优选择。
