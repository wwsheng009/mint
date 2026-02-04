# Demo 4: Complex Layout System

## 概述

复杂布局 = **Flex + Grid + Absolute + Scroll 容器**

验证声明式 UI 能否处理复杂界面。

## 核心原则

布局引擎必须支持"约束传播"：

```go
Constraints {
    MinW, MaxW int
    MinH, MaxH int
}
```

父节点把约束传下去，子节点在约束内计算尺寸。

## 布局类型

### Flex (线性分配)

| 模式   | 说明         |
| ---- | ---------- |
| Row  | 水平分配主轴空间   |
| Column | 垂直分配主轴空间   |

```
[Fixed] [Flex=1] [Flex=2]
```

固定尺寸组件先分配，剩余空间按 FlexGrow 比例分配。

### Grid (二维分区)

适用于 Dashboard 类布局：

```
+--------+--------+
| CPU    | Memory |
+--------+--------+
| Logs            |
+-----------------+
```

### Absolute (绝对定位)

用于浮层/内部定位：
- 右上角 badge
- 输入框内提示
- Tooltip

### Scroll (滚动容器)

Layout 给固定高度，内部 Content 可超出：
```go
ui.ScrollY(
    ui.Column(...1000 rows...),
).Flex(1)
```

## 复杂布局示例 (IDE)

```
+---------------------------------------------------+
| Header                                            |
+-------------------+-------------------------------+
| Sidebar           | Content                      |
|                   |   +-----------------------+  |
|                   |   | Tabs                  |  |
|                   |   +-----------------------+  |
|                   |   | Editor (scroll)       |  |
|                   |   |                       |  |
|                   |   +-----------------------+  |
|                   |   | Console               |  |
+-------------------+-------------------------------+
| Footer                                            |
+---------------------------------------------------+
```

用代码表示：

```go
ui.Column(
    Header().Height(3),
    ui.Row(
        Sidebar().Width(24),
        ui.Column(
            Tabs().Height(3),
            EditorArea().Flex(1),
            ConsolePanel().Height(8),
        ).Flex(1),
    ).Flex(1),
    Footer().Height(1),
)
```

## 运行

```bash
cd examples/ui_demos/demo4_complex_layout
go run main.go
```

## 标签页

- **Flex**: Flex 布局演示
- **Grid**: Grid 布局演示
- **Absolute**: 绝对定位演示
- **Scroll**: 滚动容器演示
- **Complex**: IDE 级复杂布局

## 验证目标

布局系统成熟度标志：

| 能力         | 说明        |
| ---------- | --------- |
| 嵌套 Flex    | 多级分区      |
| Grid 跨行跨列  | Dashboard |
| 固定 + 自适应混合 | 真 UI      |
| 绝对定位       | 覆盖元素      |
| Scroll 区域  | 内容超出      |
| 最小/最大尺寸    | 防止挤爆      |

## 基于文档

`framework/docs/ui/demo/demo4_layout.md`
