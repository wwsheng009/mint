# Demo2 使用指南

**Demo2**: Runtime Internals Visualization（运行时内部可视化）

这个 demo 展示了 Mint TUI 的完整渲染管道，从事件到终端输出的全过程。

---

## 🚀 快速开始

### 方式 1: 运行原始 demo2（推荐新手）

```bash
cd examples/ui_demos/demo2_runtime_internals

# 直接运行已编译的二进制
./demo2.exe                    # Windows
./demo2_runtime_internals.exe  # 也可以用这个

# 或者从源代码运行
go run main.go
```

### 方式 2: 运行带 Inspector 的版本（推荐调试）

```bash
cd examples/ui_demos/demo2_runtime_internals/inspector_demo

# Windows
.\demo2_inspector.exe

# Linux/macOS
./demo2_inspector
```

### 方式 3: 运行简化 Inspector 演示

```bash
cd examples/ui_demos/demo2_runtime_internals/inspector_demo/simple

# Windows
.\simple_demo.exe

# Linux/macOS
./simple_demo
```

---

## 📖 Demo2 功能说明

### 可视化的渲染管道

Demo2 展示了完整的渲染管道：

```
Event → setState → Scheduler → Render → Reconcile → Layout → Paint → Buffer → Terminal
```

### 界面布局

```
┌─ Runtime Scheduling Pipeline Visualization ─┐
│                                              │
│  [Event]    [setState]    [Scheduler]      │
│      ↓            ↓             ↓            │
└──────────────────────────────────────────────┘

┌─ Statistics ───────────────────────────────┐
│ Events:     0    Renders:    0    Buffers: 0│
└──────────────────────────────────────────────┘

┌─ Controls ─────────────────────────────────┐
│ [1] Event [2]setState [3]Scheduler         │
│ [4] Render [5]Reconcile [6] Layout          │
│ [7] Paint [0] Idle                         │
└──────────────────────────────────────────────┘

┌─ Phase Explanation ───────────────────────┐
│ System idle, waiting for events...         │
└──────────────────────────────────────────────┘
```

---

## 🎮 操作指南

### 基本控制

1. **启动程序**
   ```bash
   ./demo2.exe
   ```

2. **触发不同阶段**
   - 按 `Tab` 在按钮间切换
   - 按 `Enter` 点击按钮
   - 或直接点击按钮（如果支持鼠标）

3. **观察变化**
   - 点击 `[1] Event` → Event 计数增加，显示 "Event" 阶段
   - 点击 `[2] setState` → 显示 "setState" 阶段
   - 点击 `[3] Scheduler` → Render 计数增加
   - 依此类推...

### 按钮说明

| 按钮 | 功能 | 说明 |
|------|------|------|
| `[1] Event` | 触发事件 | 模拟接收到用户输入事件 |
| `[2] setState` | 状态更新 | 标记组件需要重新渲染 |
| `[3] Scheduler` | 调度器 | 批量处理脏组件 |
| `[4] Render` | 渲染 | 调用组件函数生成 VNode |
| `[5] Reconcile` | 协调 | 对比新旧 VNode 树 |
| `[6] Layout` | 布局 | 计算位置和尺寸 |
| `[7] Paint` | 绘制 | 渲染到缓冲区 |
| `[0] Idle` | 空闲 | 重置到空闲状态 |

---

## 🔧 高级功能

### 1. 带 Inspector 的版本

**运行**:
```bash
cd inspector_demo
./demo2_inspector
```

**特点**:
- ✅ 右侧显示 Inspector 面板
- ✅ 实时性能监控（FPS、内存）
- ✅ 布局诊断（问题检测）
- ✅ 树视图统计
- ✅ `[I] Toggle Inspector` 按钮开关

**界面**:
```
┌─ Main Content ────────┬─ Inspector ────┐
│ Pipeline               │                 │
│ Statistics             │ Performance:     │
│ Controls               │   FPS: 60.0      │
│ Explanation           │   Mem: 2.5 MB    │
│                       │                 │
└───────────────────────┴─────────────────┘
```

### 2. 调试模式

启用布局调试：
```bash
# 设置环境变量
export TUI_UI_DEBUG_LAYOUT=true
# 或
export TUI_LAYOUT_DEBUG=true

# 运行
go run main.go
```

### 3. 主题设置

Demo2 使用 Nord 主题，可以修改：
```go
// 在 main.go 中
theme.SetTheme("nord")    // 当前
// theme.SetTheme("gruvbox")
// theme.SetTheme("dracula")
```

---

## 📊 输出示例

### 正常运行

```
┌─ Runtime Scheduling Pipeline Visualization ─┐
│                                              │
│  [Event]    [setState]    [Scheduler]      │
│      ↓            ↓             ↓            │
│                                              │
│  [Render]   [Reconcile]    [Layout]         │
│      ↓            ↓             ↓            │
│                                              │
│  [Paint]                                     │
│      ↓                                        │
└──────────────────────────────────────────────┘

┌─ Statistics ───────────────────────────────┐
│ Events:    15   Renders:   12   Buffers:  10 │
└──────────────────────────────────────────────┘

┌─ Controls ─────────────────────────────────┐
│ [1] Event [2]setState [3]Scheduler [4] Render│
│ [5]Reconcile [6] Layout [7] Paint [0] Idle   │
└──────────────────────────────────────────────┘

┌─ Phase Explanation ───────────────────────┐
│ Render: Component functions called to      │
│ generate VNode trees from state.            │
└──────────────────────────────────────────────┘
```

### Inspector 输出

```
╔═ UI INSPECTOR ═╗
┌─ Performance ─
FPS: 60.0 | Mem: 2.5 MB

┌─ Diagnostics ──
✓ No layout problems

┌─ Layout Tree ────
Nodes: 152 | Depth: 8

Instructions:
  F12: Toggle
  Tab: Next element
```

---

## 🛠️ 编译说明

### 从源代码编译

```bash
cd examples/ui_demos/demo2_runtime_internals

# 编译原始版本
go build -o demo2.exe main.go

# 编译 Inspector 版本
cd inspector_demo
go build -o demo2_inspector.exe main.go

# 编译简化版本
cd simple
go build -o simple_demo.exe main.go
```

### 使用 Make（Inspector 版本）

```bash
cd examples/ui_demos/demo2_runtime_internals/inspector_demo

# 查看帮助
make help

# 构建
make build

# 运行
make run
make run-simple
make run-enabled
```

---

## 📚 相关文档

### Demo2 相关

- `README.md` - Demo2 基本说明
- `CHANGELOG.md` - 更新日志
- `WRAP_MIGRATION.md` - Wrap 组件迁移说明
- `FILLWIDTH_FIX.md` - FillWidth 修复说明

### Inspector 相关

- `QUICKSTART_INSPECTOR.md` - 快速开始指南
- `INSPECTOR_README.md` - 详细使用说明
- `INSPECTOR_INTEGRATION_SUMMARY.md` - 集成总结
- `inspector_demo/README.md` - Inspector Demo 说明

---

## 🎯 推荐学习路径

### 初学者

1. **运行原始 demo2**
   ```bash
   ./demo2.exe
   ```

2. **尝试所有按钮**
   - 点击每个按钮观察效果
   - 理解渲染管道的流程

3. **阅读阶段说明**
   - 底部的解释面板会更新
   - 理解每个阶段的作用

### 中级用户

1. **使用 Inspector 版本**
   ```bash
   cd inspector_demo
   ./demo2_inspector
   ```

2. **观察性能指标**
   - FPS 变化
   - 内存使用

3. **查看布局问题**
   - 诊断面板
   - 树统计

### 高级用户

1. **启用调试模式**
   ```bash
   TUI_LAYOUT_DEBUG=true go run main.go
   ```

2. **修改代码**
   - 调整管道显示
   - 添加新阶段
   - 修改布局

3. **集成 Inspector**
   - 参考集成代码
   - 添加到你的应用

---

## 💡 使用技巧

### 技巧 1: 理解渲染流程

按顺序点击按钮，观察管道流动：
```
[1] Event → [2] setState → [3] Scheduler → [4] Render
→ [5] Reconcile → [6] Layout → [7] Paint → [0] Idle
```

### 技巧 2: 监控性能

使用 Inspector 版本观察：
- FPS 是否稳定在 60
- 内存是否增长
- 渲染时间是否正常

### 技巧 3: 调试布局

```bash
# 启用布局调试
TUI_LAYOUT_DEBUG=true ./demo2.exe

# 查看布局信息
# 会打印每个元素的布局详情
```

### 技巧 4: 自动化测试

```bash
# 运行测试
go test -v -run TestLayoutInfo

# 查看输出
cat demo_test.go
```

---

## ❓ 常见问题

### Q: 如何退出？

A: 按 `Ctrl+C` 或 `Ctrl+D`，或关闭终端窗口

### Q: 按钮没反应？

A:
- 确保 NumLock 关闭
- 尝试用 Tab 切换焦点
- 检查终端是否支持鼠标

### Q: 窗口太小？

A: 调整终端窗口大小到至少 100x35

### Q: 性能数据异常？

A:
- 正常运行几秒后数据才会准确
- Inspector 开启时会有轻微性能开销

### Q: 如何重新编译？

A:
```bash
go clean
go build -o demo2.exe main.go
```

---

## 🎓 下一步

### 1. 深入学习

- 阅读源代码 `main.go`
- 理解渲染管道实现
- 学习 UI Inspector 使用

### 2. 修改和实验

- 修改按钮布局
- 添加新阶段
- 改变主题颜色
- 集成到你的应用

### 3. 查看 Inspector 功能

```bash
cd inspector_demo
cat README.md           # 详细文档
cd simple
go run main.go          # 运行简化演示
```

---

## 📞 获取帮助

- 查看相关文档
- 运行 `make help`（Inspector 版本）
- 阅读 UI Inspector 文档

---

**享受探索 Mint TUI 的渲染管道！** 🚀

**提示**: 如果你是第一次使用，建议先运行原始 demo2 (`./demo2.exe`)，然后再尝试 Inspector 版本。
