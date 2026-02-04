# Demo 5: IDE Interface Demo

## 概述

这是一个 **IDE 级场景级 Demo**，用来压测架构、验证设计是否真的工业级。

## 覆盖能力

| 系统           | 是否覆盖 |
| ------------ | ---- |
| 声明式组件        | ✔    |
| Grid 布局      | ✔    |
| Flex 布局      | ✔    |
| 虚拟列表         | ✔    |
| Scroll       | ✔    |
| 输入系统         | ✔    |
| 光标管理         | ✔    |
| Layer        | ✔    |
| Modal        | ✔    |
| Diff 优化      | ✔    |
| 样式系统         | ✔    |
| 增量布局         | ✔    |
| 增量 Reconcile | ✔    |

## 界面结构

```
+-------------------------------------------------------+
| File  Edit  View  Run                                |
+-----------+-------------------------------------------+
| Explorer  | main.go                                    |
| > src     |-------------------------------------------|
|   > ui    | func main() {                             |
|   > core  |     ui.Run(App)                           |
|           | }                                         |
|           | (scroll)                                  |
+-----------+-----------------------------+-------------+
| Console (logs streaming)                | Problems    |
+-----------------------------------------+-------------+
```

## 模块结构

```
app/
 ├── main.go
 ├── state/
 │     ├── app_state.go
 │     ├── editor_state.go
 │     ├── filetree_state.go
 │     └── console_state.go
 ├── ui/
 │     ├── layout.go
 │     ├── header.go
 │     ├── sidebar.go
 │     ├── editor.go
 │     ├── console.go
 │     ├── statusbar.go
 │     ├── tabs.go
 │     └── modals/
 │           ├── command_palette.go
 │           └── find_modal.go
```

## 核心组件

### FileTree (侧边栏)
- 虚拟列表
- 折叠树
- Focus 行

### Editor (编辑器)
- 输入系统
- 光标定位
- 滚动同步
- 仅可见行渲染

### Tabs (标签页)
- 多文件标签
- 激活状态高亮
- 关闭按钮

### Console (控制台)
- 实时日志流
- 自动滚动
- 颜色编码

### Command Palette (命令面板)
- Modal Layer
- Focus Trap
- ESC 关闭

## 运行

```bash
cd examples/ui_demos/demo5_ide
go run main.go
```

## 验证目标

如果这个 Demo 能够：
- resize 不抖
- 输入不卡
- 日志滚动流畅
- 打开 Modal 不穿透

那么引擎已经达成：
> **工业级 UI Runtime**

## 基于文档

`framework/docs/ui/demo/demo5_ide.md`
