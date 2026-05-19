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

`../../../docsArchive/cleanup-2026-05-19/_framework_docs/ui/demo/demo1.md`
