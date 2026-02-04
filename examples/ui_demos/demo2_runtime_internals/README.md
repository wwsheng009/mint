# Demo 2: Runtime Internals Visualization

## 概述

这个 Demo 可视化了从 `setState` 到最终写入终端 Buffer 的完整 Runtime 调度链路。

## 调度流水线

```
Event → setState → Scheduler → Render → Reconcile → Layout → Paint → Buffer Diff → Terminal
```

## 各阶段说明

| 阶段        | 说明                    |
| --------- | --------------------- |
| Event     | 键盘/鼠标/Resize 事件捕获   |
| setState  | 组件状态变化，标记 dirty       |
| Scheduler | 批处理脏组件，调度渲染           |
| Render    | 调用组件函数生成 VNode 树     |
| Reconcile | Diff 算法比较新旧 VNode    |
| Layout    | 约束传播，计算位置和尺寸          |
| Paint     | 渲染到 back buffer       |
| Buffer Diff| 对比前后 buffer，最小化输出 |

## 运行

```bash
cd examples/ui_demos/demo2_runtime_internals
go run main.go
```

## 交互

按数字键 1-7 触发各阶段：
- `[1]` Event - 模拟事件输入
- `[2]` setState - 状态变化
- `[3]` Scheduler - 调度器
- `[4]` Render - 渲染
- `[5]` Reconcile - 协调
- `[6]` Layout - 布局
- `[7]` Paint - 绘制
- `[0]` Idle - 返回空闲状态

## 验证目标

理解这套系统本质上是：
> **终端版浏览器渲染引擎** (Terminal版 React + Skia + GPU Pipeline)

| 浏览器         | 你的 TUI      |
| ----------- | ----------- |
| DOM         | VNode       |
| Render Tree | RNode       |
| CSS Layout  | Flex/Layout |
| Compositor  | Layer 系统    |
| Canvas      | Buffer      |
| GPU Diff    | Buffer Diff |

## 基于文档

`framework/docs/ui/demo/demo2_inside.md`
