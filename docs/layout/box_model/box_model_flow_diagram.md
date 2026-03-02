# Box Model 约束与测量流程图

## 1. 当前实现的问题流程

### 1.1 约束传播（错误）

```
Parent (100x100)
    │
    ├─ Engine.layoutNodeWithDepth()
    │   │
    │   ├─ ❌ 获取 Padding? NO!
    │   ├─ ❌ 获取 Border? YES
    │   └─ ❌ 获取 Margin? YES (仅用于布局)
    │
    ├─ Engine.Measure()
    │   │
    │   └─ Node.Measure(constraints)
    │       │
    │       └─ constraints = 100x100  // ❌ 没有扣除 padding/border
    │
    └─ 子节点收到: 100x100
        │
        └─ 返回: 内容尺寸 (例如 80x80)
            │
            └─ 父容器期待: 100x100
                │
                └─ ❌ 不一致！期望 100x100，但只有 80x80
```

### 1.2 布局位置计算（错误）

```
Container (0, 0, 100, 100)
    │
    ├─ FlexLayout.LayoutChildren(100, 100)
    │   │
    │   ├─ ❌ width/height = 100x100 (包含 padding?)
    │   │   不确定：可能包含也可能不包含
    │   │
    │   ├─ 子节点测量
    │   │   └─ childConstraints: 100x100 - padding(20) = 80x80
    │   │   └─ child.Measure() 返回: 60x60
    │   │
    │   ├─ 子节点位置
    │   │   └─ child.X = 0  // ❌ 没有加 padding
    │   │   └─ child.Y = 0  // ❌ 没有加 padding
    │   │
    │   └─ 返回 boxes
    │       └─ box.X = 0, box.Y = 0  // ❌ 相对于外部边界
    │
    └─ 返回到 Engine
        │
        ├─ ❌ 没有应用 padding 偏移
        └─ child.X = 0, child.Y = 0  // ❌ 错误的位置
```

## 2. 理想的正确流程

### 2.1 约束传播（正确）

```
Parent Container
    │ Dimensions: 100x100
    │ Padding: [10, 20, 10, 20] // top, right, bottom, left
    │ Border: Single (1 char)
    │
    ├─ Engine.layoutNodeWithDepth(x=0, y=0, constraints=100x100)
    │   │
    │   ├─ Step 1: 获取 BoxModel
    │   │   └─ Padding.Horizontal = 20 + 20 = 40
    │   │   └─ Padding.Vertical = 10 + 10 = 20
    │   │   └─ Border.Horizontal = 2
    │   │   └─ Border.Vertical = 2
    │   │   └─ 总水平占用 = 40 + 2 = 42
    │   │   └─ 总垂直占用 = 20 + 2 = 22
    │   │
    │   ├─ Step 2: 测量容器
    │   │   │
    │   │   └─ FlexLayout.Measure(constraints=100x100)
    │   │       │
    │   │       ├─ 计算内部空间
    │   │       │   └─ innerWidth = 100 - 40 = 60  // 扣除 padding
    │   │       │   └─ innerHeight = 100 - 20 = 80
    │   │       │
    │   │       ├─ 调用 child.Measure(innerConstraints)
    │   │       │   └─ child 收到: 60x80
    │   │       │
    │   │       └─ 返回: 总尺寸(包含 padding)
    │   │           └─ contentWidth = 60
    │   │           └─ totalWidth = 60 + 40 = 100
    │   │           └─ contentHeight = 80
    │   │           └─ totalHeight = 80 + 20 = 100
    │   │
    │   └─ Step 3: 尺寸
    │       └─ 100x100 ✅ 正确！
    │
    ├─ Step 4: 布局子节点
    │   │
    │   └─ FlexLayout.LayoutChildren(innerWidth=60, innerHeight=80)
    │       │
    │       ├─ 子节点 A: width=30
    │       ├─ 子节点 B: width=30
    │       ├─ Gap: 0
    │       │
    │       ├─ 计算位置（相对于 padding 内部）
    │       │   ├─ childA.X = 0, childA.Y = 0
    │       │   └─ childB.X = 30, childB.Y = 0
    │       │
    │       └─ 返回 boxes (内部坐标)
    │           └─ {X:0, Y:0, W:30, H:80}
    │           └─ {X:30, Y:0, W:30, H:80}
    │
    ├─ Step 5: 应用偏移
    │   │
    │   ├─ ContentOffsetX = padding_left(20) + border_offset(1) = 21
    │   ├─ ContentOffsetY = padding_top(10) + border_offset(1) = 11
    │   │
    │   └─ 转换到全局坐标
    │       ├─ childA.X = 0 + 21 = 21 ✅
    │       ├─ childA.Y = 0 + 11 = 11 ✅
    │       ├─ childB.X = 30 + 21 = 51 ✅
    │       └─ childB.Y = 0 + 11 = 11 ✅
    │
    └─ 最终布局 ✅
```

### 2.2 布局位置计算（正确）

```
Container (0, 0, 100, 100)
    │ Padding: [10, 20, 10, 20]
    │ Border: Single
    │
    ├─ layoutNodeWithDepth(x=0, y=0)
    │   │
    │   ├─ GetBoxModel()
    │   │   ├─ HorizontalPadding = 20 + 20 + 2 = 42
    │   │   └─ VerticalPadding = 10 + 10 + 2 = 22
    │   │   └─ ContentOffsetX = 20 + 1 = 21
    │   │   └─ ContentOffsetY = 10 + 1 = 11
    │   │
    │   ├─ Measure(constraints=100x100)
    │   │   └─ 返回 Size{100, 100}
    │   │
    │   ├─ 内部空间
    │   │   ├─ innerWidth = 100 - 42 = 58
    │   │   └─ innerHeight = 100 - 22 = 78
    │   │
    │   ├─ FlexLayout.LayoutChildren(58, 78)
    │   │   │
    │   │   ├─ 子节点 A
    │   │   │   └─ innerConstraints: 58x78
    │   │   │   └─ 返回 Size{20, 78}
    │   │   │
    │   │   ├─ 子节点 B
    │   │   │   └─ innerConstraints: 58x78
    │   │   │   └─ 返回 Size{20, 78}
    │   │   │
    │   │   └─ 返回内部布局
    │   │       ├─ boxA.X = 0, boxA.Y = 0
    │   │       └─ boxB.X = 20, boxB.Y = 0
    │   │
    │   └─ 递归布局子节点
    │       ├─ childA
    │       │   ├─ x = containerX + boxA.X + ContentOffsetX = 0 + 0 + 21 = 21
    │       │   ├─ y = containerY + boxA.Y + ContentOffsetY = 0 + 0 + 11 = 11
    │       │   └─ constraints: 20x78
    │       │
    │       └─ childB
    │           ├─ x = 0 + 20 + 21 = 41
    │           ├─ y = 0 + 0 + 11 = 11
    │           └─ constraints: 20x78
    │
    └─ ✅ 正确的布局
        ┌─────────────────────────────────┐
        │ ┌─────────────────────────────┐ │
        │ │ │ A │ ── │ B │ │            │ │
        │ │ └─────────────────────┘      │ │
        │ │ Padding: Top=10              │ │
        └─────────────────────────────────┘
          Left=20  Right=20
```

## 3. 数据流对比

### 3.1 当前实现（有问题）

```
┌────────────────────────────────────────────────────┐
│              父容器约束: 100x100                     │
└────────────────────────────────────────────────────┘
    │
    ▼
┌────────────────────────────────────────────────────┐
│  FlexLayout.Measure(constraints=100x100)           │
│  ❌ 没有处理 Padding                                │
│      │                                              │
│      ├─ childConstraints = 100x100  // ❌ 没扣减    │
│      │                                              │
│      └─ 返回 Size{...}  // ❌ 是否包含 padding 不确定 │
└────────────────────────────────────────────────────┘
    │
    ▼
┌────────────────────────────────────────────────────┐
│  子节点 Measure(constraints=100x100)               │
│      ↓                                              │
│  返回 Size{80, 80}                                  │
└────────────────────────────────────────────────────┘
    │
    ▼
┌────────────────────────────────────────────────────┐
│  FlexLayout 累加子节点                              │
│  ❌ 不确定是否应该加 padding                         │
└────────────────────────────────────────────────────┘
```

### 3.2 修复后（正确）

```
┌────────────────────────────────────────────────────┐
│              父容器约束: 100x100                     │
│              Padding: [10, 20, 10, 20]              │
└────────────────────────────────────────────────────┘
    │
    ▼
┌────────────────────────────────────────────────────┐
│  Engine.Measure(node, constraints=100x100)         │
│      │                                              │
│      ├─ GetBoxModel()                              │
│      │   └─ HorizontalPadding = 40                  │
│      │   └─ VerticalPadding = 20                    │
│      │                                              │
│      ├─ innerConstraints = (60, 80)  // ✅ 扣除了  │
│      │                                              │
│      ├─ node.Measure(innerConstraints)             │
│      │      ↓                                       │
│      │   返回 contentSize{60, 80}                   │
│      │                                              │
│      └─ 返回 totalSize{100, 100}  // ✅ 加回了     │
│          (60 + 40, 80 + 20)                        │
└────────────────────────────────────────────────────┘
    │
    ▼
┌────────────────────────────────────────────────────┐
│  Engine.layoutNodeWithDepth(...)                  │
│      │                                              │
│      ├─ ContentOffsetX = 20 + 1 = 21               │
│      ├─ ContentOffsetY = 10 + 1 = 11               │
│      │                                              │
│      ├─ innerWidth = 100 - 42 = 58                 │
│      ├─ innerHeight = 100 - 22 = 78                │
│      │                                              │
│      ├─ FlexLayout.LayoutChildren(58, 78)          │
│      │   └─ 返回内部布局 (相对于 padding)           │
│      │       {X:0, Y:0, W:20, H:78}                │
│      │       {X:20, Y:0, W:20, H:78}               │
│      │                                              │
│      └─ 应用偏移                                   │
│          ├─ childA.X = 0 + 21 = 21                 │
│          ├─ childA.Y = 0 + 11 = 11                 │
│          ├─ childB.X = 20 + 21 = 41                 │
│          └─ childB.Y = 0 + 11 = 11                 │
└────────────────────────────────────────────────────┘
```

## 4. 关键概念图示

### 4.1 尺寸构成

```
总宽度 = 内容宽度 + 左边距 + 右边距 + 左边框 + 右边框

示例：Padding=[10, 20], Border=Single

       ┌─┬─────────────────┬─┐  总宽度
       │ │                 │ │
       │ │    内容区域      │ │  内容宽度
       │ │                 │ │  │
左padding │                 │ │ 右padding
  10      │                 │ │    20
       │ │                 │ │
       └─┴─────────────────┴─┘
       边框              边框
        1                1

总宽度 = 内容 + 10 + 20 + 1 + 1 = 内容 + 32
```

### 4.2 坐标系统

```
全局坐标系统:
(0, 0) ──────────────────────────────► X
   │
   │
   ▼ Y

容器边界: (x, y, width, height)
    │ Padding: [pt, pr, pb, pl]
    │ Border: Single
    │
    ┌──────────────────────────────────┐  ← y (容器顶部)
    │ ┌────────────────────────────┐ │
    │ │ 边框占用 1                 │ │
    │ ├────────────────────────────┤ │  ← y + 1 (边框底)
    │ │ │ Padding Top (pt)       │ │ │
    │ │ ├────────────────────────┤ │ │  ← y + 1 + pt (内容区)
    │ │ │ │                    │ │ │
    │ │ │ │   内容区域          │ │ │
    │ │ │ │                    │ │ │
    │ │ │ └────────────────────┘ │ │ │
    │ │ └────────────────────────┘ │ │  ← y + 1 + pt + contentH
    │ │ Padding Bottom (pb)          │ │
    │ └────────────────────────────┘ │
    └──────────────────────────────────┘  ← y + height (容器底)
      ↑                                 ↑
      x                             x + width
      |
      x + 1 + pl (内容区左)

内容区左上角: (x + 1 + pl, y + 1 + pt)
内容区右下角: (x + width - 1 - pr, y + height - 1 - pb)
```

## 5. 约束传播的层次结构

```
Root Constraints: 800x600
    │
    ├─ Engine.Measure(root, 800x600)
    │   │
    │   ├─ [扣除 root box model]
    │   │   └─ innerConstraints: 700x500
    │   │
    │   ├─ root.Measure(700x500)
    │   │   │
    │   │   ├─ [扣除 root.Padding]
    │   │   │   └─ childConstraints: 600x400
    │   │   │
    │   │   ├─ child1.Measure(600x400)
    │   │   │   └─ 返回: 内容尺寸
    │   │   │
    │   │   ├─ child2.Measure(600x400)
    │   │   │   └─ 返回: 内容尺寸
    │   │   │
    │   │   └─ 返回: 总尺寸(600+PaddingH, 400+PaddingV)
    │   │       = 750x450 ✅
    │   │
    │   └─ 返回: 总尺寸(750+root.PaddingH, 450+root.PaddingV)
    │       = 800x600 ✅
    │
    └─ ✅ 约束正确传播和返回
```

## 6. 布局坐标计算的层次结构

```
Root Box: (0, 0, 800, 600)
    │ Padding: [50, 50, 50, 50]
    │ Border: Single
    │
    ├─ ContentOffset: (51, 51)  // 50(padding) + 1(border/2)
    │
    ├─ InnerSpace: (800-102, 600-102) = (698, 498)
    │
    ├─ FlexLayout.LayoutChildren(698, 498)
    │   │
    │   ├─ Child A (内部坐标)
    │   │   └─ {X:0, Y:0, W:300, H:498}
    │   │
    │   └─ Child B (内部坐标)
    │       └─ {X:300, Y:0, W:398, H:498}
    │
    ├─ 转换为全局坐标
    │   │
    │   ├─ Child A
    │   │   ├─ X = 0 + 51 = 51
    │   │   └─ Y = 0 + 51 = 51
    │   │
    │   └─ Child B
    │       ├─ X = 300 + 51 = 351
    │       └─ Y = 0 + 51 = 51
    │
    └─ 递归布局子节点...

Child A: (51, 51, 300, 498)
    │ Padding: [20, 20, 20, 20]
    │ Border: None
    │
    ├─ ContentOffset: (20, 20)
    │
    ├─ InnerSpace: (300-40, 498-40) = (260, 458)
    │
    ├─ 子布局...
    │
    └─ ✅ 坐标正确累积
```

## 总结

当前实现的主要问题：
1. ❌ Padding 在约束传播中未正确扣除
2. ❌ Padding 在尺寸返回时未正确加回
3. ❌ Padding 偏移在布局中未正确应用
4. ❌ Margin 语义不一致

修复后的正确流程：
1. ✅ 测量前扣除 Padding/Border
2. ✅ 测量后加回 Padding/Border
3. ✅ 布局时应用 Padding/Border 偏移
4. ✅ 统一的 BoxModel 接口
