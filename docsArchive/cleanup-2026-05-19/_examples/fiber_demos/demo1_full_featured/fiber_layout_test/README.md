# Fiber Layout Test

测试 `BuildComputedBoxFiberOnly` 函数，验证 Fiber-first 布局计算。

## 概述

本测试演示了纯 Fiber 树布局计算流程：
1. 使用 `component_fixtures` 构建 VNode 树
2. 转换为 Fiber 树
3. 调用 `BuildComputedBoxFiberOnly` 进行布局
4. 验证布局结果

## 运行

```bash
go run ./examples/fiber_demos/demo1_full_featured/fiber_layout_test/
```

## 测试内容

### Test 1: 基本布局测试
- 使用默认约束 (80x24)
- 打印完整的 ComputedBox 树结构

### Test 2: 不同约束测试
测试多种终端尺寸：
- Small (40x12)
- Medium (80x24)
- Large (120x40)
- Wide (200x10)
- Tall (60x50)

### Test 3: 独立组件测试
测试各个组件的布局：
- Header
- MainBody
- Modal
- SimpleVStack
- SimpleHStack
- NestedLayout
- FlexLayout

### Test 4: 节点计数验证
比较 VNode 数量与 ComputedBox 数量，确保一一对应。

## 输出示例

```
=== Fiber Layout Test Demo ===
Testing BuildComputedBoxFiberOnly from fiber_only_layout.go

=== Test 1: Default Constraints (80x24) ===
[OK] Layout completed successfully
     Root Box: X=0, Y=0, W=0, H=0
     Children: 3
     HitMap Size: 17

=== Test 4: Node Count Verification ===
Fixture                   VNode  ComputedBox      Match
-------------------------------------------------------
demo1_full_app               43           43         OK
demo1_header                  6            6         OK
demo1_main_body              15           15         OK
...
```

## 相关文档

- [`runtime/compute/fiber_only_layout.go`](../../../../runtime/compute/fiber_only_layout.go) - Fiber-only 布局实现
- [`docs/plan/fiber/fiber_first.md`](../../../../docs/plan/fiber/fiber_first.md) - Fiber-first 架构设计
- [`examples/component_fixtures`](../../../component_fixtures/) - 测试组件配置
