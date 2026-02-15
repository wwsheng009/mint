# Component Test Fixtures

通用 UI 组件测试配置库，用于在多个测试文件中复用组件定义。

## 概述

本包将 `demo1_full_featured` 中的 UI 配置抽象为可复用的测试组件，避免循环依赖问题，支持灵活配置和扩展。

## 快速开始

```go
import "github.com/wwsheng009/mint/examples/component_fixtures"

// 使用预定义 fixture
fixture := component_fixtures.GetFixture("demo1_full_app")
vnode := fixture.Build()
```

## 核心 API

### 1. 预定义 Fixtures

通过 `StandardFixtures()` 获取所有预定义组件：

```go
fixtures := component_fixtures.StandardFixtures()
for _, f := range fixtures {
    vnode := f.Build()
    fmt.Printf("%s: %s\n", f.Name, f.Description)
}
```

| Fixture 名称 | 描述 | 节点数 |
|-------------|------|--------|
| `demo1_full_app` | 完整 Demo1 应用 | 43 |
| `demo1_header` | 标题栏组件 | 6 |
| `demo1_main_body` | 主体布局 | 15 |
| `demo1_modal` | 确认弹窗 | 16 |
| `simple_vstack` | 简单垂直布局 | 4 |
| `simple_hstack` | 简单水平布局 | 4 |
| `nested_layout` | 嵌套布局 | 7 |
| `bordered_content` | 带边框内容 | 2 |
| `flex_layout` | Flex 布局 | 4 |
| `keyed_items` | 带键节点 | 4 |

### 2. Demo1 配置构建器

使用配置选项灵活构建组件：

```go
vnode := component_fixtures.BuildDemo1App(
    component_fixtures.WithCount(42),
    component_fixtures.WithInput("test input"),
    component_fixtures.WithItems([]string{"A", "B", "C", "D"}),
    component_fixtures.WithTheme("nord"),
    component_fixtures.WithSize(120, 40),
)
```

**配置选项：**

| 函数 | 说明 |
|-----|------|
| `WithCount(n int)` | 设置点击计数器值 |
| `WithInput(s string)` | 设置输入框内容 |
| `WithItems(items []string)` | 设置日志列表项 |
| `WithTheme(name string)` | 设置主题名称 |
| `WithSize(w, h int)` | 设置尺寸 |

### 3. 独立组件构建

```go
// 标题栏
header := component_fixtures.BuildDemo1Header(10) // count = 10

// 主体布局
body := component_fixtures.BuildDemo1MainBody(5, "hello", []string{"Item 1", "Item 2"})

// 确认弹窗
modal := component_fixtures.BuildDemo1ConfirmModal(func() {
    fmt.Println("Modal closed")
})

// 调试面板
debug := component_fixtures.BuildDemo1DebugPanel()
```

### 4. 测试辅助函数

```go
// 构建简单树
tree := component_fixtures.BuildVNodeTree(depth, breadth)

// 构建带键树（用于协调测试）
keyedTree := component_fixtures.BuildKeyedVNodeTree(depth, breadth, "root")

// 构建混合键树
mixedTree := component_fixtures.BuildMixedKeyedTree(3, 2) // 3 keyed + 2 non-keyed

// 统计节点数
count := component_fixtures.CountNodes(vnode)
```

## 使用示例

### 基础测试

```go
func TestMyComponent(t *testing.T) {
    // 获取 fixture
    fixture := component_fixtures.GetFixture("demo1_header")
    if fixture == nil {
        t.Fatal("fixture not found")
    }

    // 构建 VNode
    vnode := fixture.Build()

    // 转换为 Fiber
    fiber := reconciler.CreateFiberFromVNode(vnode)

    // 验证
    if component_fixtures.CountNodes(vnode) != countFibers(fiber) {
        t.Error("node count mismatch")
    }
}
```

### 自定义配置测试

```go
func TestWithCustomConfig(t *testing.T) {
    tests := []struct {
        name string
        opts []component_fixtures.Demo1ConfigOption
    }{
        {"empty", nil},
        {"with_count", []component_fixtures.Demo1ConfigOption{
            component_fixtures.WithCount(100),
        }},
        {"full", []component_fixtures.Demo1ConfigOption{
            component_fixtures.WithCount(50),
            component_fixtures.WithInput("test"),
            component_fixtures.WithItems([]string{"A", "B"}),
        }},
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            vnode := component_fixtures.BuildDemo1App(tc.opts...)
            if vnode == nil {
                t.Fatal("build failed")
            }
        })
    }
}
```

### NodeID 稳定性测试

```go
func TestNodeIDStability(t *testing.T) {
    // 构建初始树
    vnode := component_fixtures.BuildKeyedVNodeTree(2, 3, "root")
    fiber := reconciler.CreateFiberFromVNode(vnode)

    // 记录初始 NodeID
    initialIDs := collectNodeIDs(fiber)

    // 重新构建相同树
    vnode2 := component_fixtures.BuildKeyedVNodeTree(2, 3, "root")
    engine := compute.NewEngine()
    engine.Layout(vnode2, fiber, constraints)

    // 验证 NodeID 不变
    afterIDs := collectNodeIDs(fiber)
    // ... 比较 initialIDs 和 afterIDs
}
```

## 目录结构

```
examples/component_fixtures/
├── fixtures.go       # 组件定义和辅助函数
├── fixtures_test.go  # 使用示例和测试
└── README.md         # 本文档
```

## 设计说明

### 为什么放在 examples 目录？

为了避免循环依赖：
- `runtime/compute` 包被 `app` 包间接依赖
- 组件需要使用 `app`、`framework/theme` 等包
- 将 fixtures 放在独立包中避免循环导入

### ComponentFixture 结构

```go
type ComponentFixture struct {
    Name        string           // 唯一标识
    Description string           // 描述说明
    Build       func() ui.VNode  // 构建函数
}
```

### 扩展 Fixtures

```go
// 添加自定义 fixture
customFixture := component_fixtures.ComponentModel{
    Name:        "my_custom",
    Description: "自定义组件",
    Build: func() ui.VNode {
        return ui.VStackBuilder(
            ui.Text("Custom"),
        ).Build()
    },
}
```
