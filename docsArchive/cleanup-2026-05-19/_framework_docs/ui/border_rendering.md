# 边框渲染设计文档

## 1. 问题背景

当前边框是通过手动打印边框字符实现的，这会占用布局空间。类似于 CSS 的 `border` 属性，我们希望边框能够自动渲染在内容区域之外，不占用内容布局空间。

### 1.1 当前问题

```
当前实现（边框占用空间）:
┌─────────┐
│ content │  <- 边框字符 │ 也占据布局位置
└─────────┘

期望实现（边框不占用空间）:
┌─────────┐
│ content │  <- 内容区域与边框分离
└─────────┘
```

### 1.2 设计目标

1. **边框不占用内容空间**: 内容区域的宽度/高度计算不应包含边框
2. **支持多种边框样式**: 单线、双线、圆角、虚线等
3. **支持边框标题**: 可在顶部边框中显示标题
4. **支持边框颜色**: 可自定义边框颜色

## 2. 当前架构分析

### 2.1 代码结构

```
runtime/ui/layout.go
├── BorderStyle (枚举)
├── BorderedNode (结构体)
│   ├── ElementVNode (嵌入)
│   ├── borderStyle
│   ├── borderColor
│   └── borderLabel
├── Bordered() (构造函数)
├── BorderedBuilder (构建器)
├── RenderBorder() (生成边框 VNode)
└── GetBorderChars() (获取边框字符)

internal/render/declarative_node.go
├── paintBordered() (渲染边框到 Buffer)
└── MeasureVNodeWidth/Height() (测量内容尺寸)
```

### 2.2 处理流程

```
用户代码
    │
    ▼
ui.Bordered().Child(content).Build()
    │
    ▼
VNode 树构建 (BorderedNode 包含子 VNode)
    │
    ▼
布局阶段 (测量内容尺寸，不包含边框)
    │
    ▼
paintBordered() (渲染边框 + 内容到 Buffer)
    │
    ▼
终端显示
```

### 2.3 当前问题

1. **复杂的类型断言**: `paintBordered` 中使用了多个 interface 类型断言
2. **双重渲染**: `RenderBorder()` 生成 VNode，但 `paintBordered` 直接绘制字符
3. **调试困难**: 边框渲染逻辑分散在多个文件中
4. **职责不清晰**: 边框是布局属性还是装饰属性？

## 3. 优化设计

### 3.1 核心概念

边框是**装饰属性**，不是布局属性。类似于：
- HTML 的 `border` 属性
- CSS 的 `outline` 属性

### 3.2 简化架构

```
┌─────────────────────────────────────────┐
│         BorderedNode (VNode)            │
├─────────────────────────────────────────┤
│ • children: []VNode                    │
│ • borderStyle: BorderStyle             │
│ • borderColor: string                  │
│ • borderLabel: string                  │
├─────────────────────────────────────────┤
│ 方法:                                   │
│ • GetBorderChars() -> (rune...)        │
│ • PaintBorder(x, y, w, h, Buffer)      │
└─────────────────────────────────────────┘
```

### 3.3 渲染流程

```
1. 布局阶段:
   - 测量子节点尺寸 (contentWidth, contentHeight)
   - BorderedNode 的尺寸 = 子节点尺寸 (边框不计入)

2. 渲染阶段:
   - 获取边框字符 (通过 GetBorderChars())
   - 在内容区域外绘制边框:
     • 顶部: (x, y) 到 (x+w+1, y)
     • 左侧: (x, y+1) 到 (x, y+h+1)
     • 右侧: (x+w+1, y+1) 到 (x+w+1, y+h+1)
     • 底部: (x, y+h+1) 到 (x+w+1, y+h+1)
```

### 3.4 边框字符定义

| 样式 | 左上 | 右上 | 左下 | 右下 | 横线 | 竖线 |
|------|------|------|------|------|------|------|
| Single | ┌ | ┐ | └ | ┘ | ─ | │ |
| Double | ╔ | ╗ | ╚ | ╝ | ═ | ║ |
| Rounded | ╭ | ╮ | ╰ | ╯ | ─ | │ |
| Dashed | + | + | + | + | - | \| |

## 4. 实现方案

### 4.1 方案 A: 直接绘制（推荐）

边框在渲染阶段直接绘制到 Buffer，不参与布局计算。

**优点**:
- 简单直接
- 边框不影响布局
- 易于调试

**缺点**:
- 边框可能与其他内容重叠（需要 z-index 概念）

### 4.2 方案 B: 虚拟节点扩展

边框作为特殊的 VNode 属性，在渲染时特殊处理。

**优点**:
- 与现有 VNode 体系集成
- 可以通过属性控制

**缺点**:
- 需要修改渲染器
- 复杂度较高

## 5. 待解决问题

1. **Z-Index**: 边框应该绘制在内容之上还是之下？
2. **边框宽度**: 当前只支持 1 字符宽度的边框
3. **内边距**: Padding 如何与边框配合？
4. **嵌套边框**: 多层嵌套时如何处理？

## 6. 实现检查清单

- [ ] BorderedNode 基本结构
- [ ] GetBorderChars() 方法
- [ ] PaintBorder() 方法
- [ ] 无标题边框渲染
- [ ] 带标题边框渲染
- [ ] 多种边框样式支持
- [ ] 边框颜色支持
- [ ] 单元测试
- [ ] 集成测试

## 7. 测试用例

```go
// 1. 简单边框
Bordered().Child(Text("Hello")).Build()
// 期望: ┌─────┐
//       │Hello│
//       └─────┘

// 2. 带标题边框
Bordered().Label("Title").Child(Text("Hello")).Build()
// 期望: ┌─ Title ──┐
//       │Hello     │
//       └──────────┘

// 3. 多行内容
Bordered().Child(VStack(
    Text("Line 1"),
    Text("Line 2"),
)).Build()
// 期望: ┌────────┐
//       │Line 1  │
//       │Line 2  │
//       └────────┘

// 4. 不同样式
Bordered().Style("double").Child(Text("Hello")).Build()
// 期望: ╔═══════╗
//       ║Hello ║
//       ╚═══════╝
```
