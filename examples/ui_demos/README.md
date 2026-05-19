# Mint TUI UI Demos

本目录包含基于历史设计稿创建并持续维护的示例程序。原始设计稿已归档到 `../../docsArchive/cleanup-2026-05-19/_framework_docs/ui/demo/`。

## 目录结构

```
ui_demos/
├── demo1_full_featured/    # 全功能展示 Demo
├── demo2_runtime_internals/ # Runtime 调度可视化
├── demo3_styling/          # 样式系统 (TUI CSS)
├── demo4_complex_layout/   # 复杂布局系统
├── demo5_ide/              # IDE 界面 Demo
└── README.md               # 本文件
```

## 各 Demo 说明

### Demo 1: Full-Featured Demo

**目录**: `demo1_full_featured/`

验证 Mint TUI 引擎的架构完整性，覆盖：
- 声明式组件
- 状态系统 (Hooks)
- 布局系统 (Flex, VStack, HStack)
- Modal (Layer)
- Input with Focus
- Scroll 容器
- VirtualList
- 事件处理

**运行**:
```bash
cd demo1_full_featured
go run main.go
```

**来源设计稿**: `../../docsArchive/cleanup-2026-05-19/_framework_docs/ui/demo/demo1.md`

---

### Demo 2: Runtime Internals Visualization

**目录**: `demo2_runtime_internals/`

可视化从 `setState` 到终端 Buffer 输出的完整调度链路：

```
Event → setState → Scheduler → Render → Reconcile → Layout → Paint → Buffer Diff → Terminal
```

**运行**:
```bash
cd demo2_runtime_internals
go run main.go
```

**来源设计稿**: `../../docsArchive/cleanup-2026-05-19/_framework_docs/ui/demo/demo2_inside.md`

---

### Demo 3: Styling System (TUI CSS)

**目录**: `demo3_styling/`

样式系统 = **CSS Box Model + 终端颜色/属性系统 + 继承规则**

特性：
- Box Model (Padding, Margin, Border)
- Color (Foreground, Background)
- Text Attributes (Bold, Italic, Underline)
- Style Inheritance
- Theme System

**运行**:
```bash
cd demo3_styling
go run main.go
```

**来源设计稿**: `../../docsArchive/cleanup-2026-05-19/_framework_docs/ui/demo/demo3_with_style.md`

---

### Demo 4: Complex Layout System

**目录**: `demo4_complex_layout/`

复杂布局 = **Flex + Grid + Absolute + Scroll 容器**

验证声明式 UI 能否处理复杂界面。

**运行**:
```bash
cd demo4_complex_layout
go run main.go
```

**来源设计稿**: `../../docsArchive/cleanup-2026-05-19/_framework_docs/ui/demo/demo4_layout.md`

---

### Demo 5: IDE Interface Demo

**目录**: `demo5_ide/`

IDE 级场景 Demo，用来压测架构、验证设计是否真的工业级。

界面结构：
- Menu Bar
- File Explorer (侧边栏)
- Editor (编辑器)
- Tabs (标签页)
- Terminal/Console (控制台)
- Status Bar
- Command Palette (命令面板)

**运行**:
```bash
cd demo5_ide
go run main.go
```

**来源设计稿**: `../../docsArchive/cleanup-2026-05-19/_framework_docs/ui/demo/demo5_ide.md`

---

## 运行所有 Demos

```bash
# Demo 1
cd examples/ui_demos/demo1_full_featured && go run main.go

# Demo 2
cd examples/ui_demos/demo2_runtime_internals && go run main.go

# Demo 3
cd examples/ui_demos/demo3_styling && go run main.go

# Demo 4
cd examples/ui_demos/demo4_complex_layout && go run main.go

# Demo 5
cd examples/ui_demos/demo5_ide && go run main.go
```

## 引擎能力覆盖

| Demo | 声明式组件 | 状态系统 | 布局系统 | Layer | Focus | Scroll | 虚拟化 | 样式系统 |
|------|----------|---------|---------|-------|-------|--------|-------|---------|
| 1    | ✔        | ✔      | ✔       | ✔     | ✔     | ✔      | ✔     | ✔       |
| 2    | ✔        | ✔      | -       | -     | -     | -      | -     | ✔       |
| 3    | -        | -      | -       | -     | -     | -      | -     | ✔       |
| 4    | ✔        | -      | ✔       | -     | -     | ✔      | -     | ✔       |
| 5    | ✔        | ✔      | ✔       | ✔     | ✔     | ✔      | ✔     | ✔       |

## 总结

这些 Demo 验证了 Mint TUI 框架已经达到：

> **Terminal UI Engine Runtime** 级别

而不是简单的 TUI Framework。引擎具备：
- React 架构思想
- Flutter 渲染模型
- VSCode 编辑器核心能力
- 游戏引擎级调度
- GPU 批处理思想的终端渲染优化
