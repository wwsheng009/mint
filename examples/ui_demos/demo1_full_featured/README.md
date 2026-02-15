# Demo 1: Full-Featured Demo App

## 概述

这是一个**全功能展示型 Demo**，专门用来验证 Mint TUI 引擎的架构完整性。

## 覆盖的功能

| 功能      | 验证点                |
| ------- | ------------------- |
| 声明式组件   | 所有组件写法              |
| 状态 Hook | UseState, UseStateInt |
| 局部刷新    | count / input       |
| 布局系统    | VStack / HStack    |
| Layer   | Modal               |
| Focus   | Input + Modal       |
| 事件系统    | Button              |
| Scroll  | 虚拟列表容器              |
| 虚拟化     | 100 行日志             |
| Diff    | 状态更新时只刷新局部          |

## 运行

```bash
cd examples/ui_demos/demo1_full_featured
go run main.go
```

## 预期界面

```
+--------------------------------------------------+
| TUI Engine Demo                            [Open Modal] Clicks: 2 |
+-----------+--------------------------------------+
| Menu      | [ Input box....................... ] |
| Add Count |--------------------------------------|
| Subtract  | Log line #0000                          |
| Quit      | Log line #0001                          |
|           | ...                                  |
+-----------+--------------------------------------+
```

弹窗出现时：

```
      +--------------------------------------+
      |                                      |
      |    *** Are you sure? ***             |
      |                                      |
      |        [Cancel] [OK]                 |
      |                                      |
      +--------------------------------------+
```

## 验证目标

如果这个 Demo 能够：
- 不闪屏
- 滚动流畅
- 输入不卡
- Modal 正确拦截事件
- 光标始终准确

那么架构就是 **生产级 TUI Runtime 设计**。

## 基于文档

`framework/docs/ui/demo/demo1.md`

---

## 组件测试配置

本 Demo 的 UI 配置已被提取为通用测试组件，可供多个测试文件复用。

### 位置

```
examples/component_fixtures/
├── fixtures.go       # 组件定义
├── fixtures_test.go  # 测试示例
└── README.md         # 使用文档
```

### 快速使用

```go
import "github.com/wwsheng009/mint/examples/component_fixtures"

// 使用预定义 fixture
fixture := component_fixtures.GetFixture("demo1_full_app")
vnode := fixture.Build()

// 自定义配置
vnode := component_fixtures.BuildDemo1App(
    component_fixtures.WithCount(42),
    component_fixtures.WithInput("test"),
    component_fixtures.WithItems([]string{"A", "B", "C"}),
)
```

### 可用 Fixtures

| 名称 | 描述 | 节点数 |
|-----|------|--------|
| `demo1_full_app` | 完整应用 | 43 |
| `demo1_header` | 标题栏 | 6 |
| `demo1_main_body` | 主体布局 | 15 |
| `demo1_modal` | 确认弹窗 | 16 |
| `simple_vstack` | 简单垂直布局 | 4 |
| `simple_hstack` | 简单水平布局 | 4 |
| `nested_layout` | 嵌套布局 | 7 |
| `bordered_content` | 带边框内容 | 2 |
| `flex_layout` | Flex 布局 | 4 |
| `keyed_items` | 带键节点 | 4 |

### 测试辅助函数

```go
// 构建测试树
tree := component_fixtures.BuildVNodeTree(depth, breadth)
keyedTree := component_fixtures.BuildKeyedVNodeTree(depth, breadth, "root")
mixedTree := component_fixtures.BuildMixedKeyedTree(3, 2)

// 统计节点
count := component_fixtures.CountNodes(vnode)
```

### 详细文档

参见 [`examples/component_fixtures/README.md`](../../component_fixtures/README.md)
