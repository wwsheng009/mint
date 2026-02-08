# UI Inspector Demo - Quick Start Guide

快速指南：在 demo2 上使用 UI Inspector 调试工具

## 🚀 一键运行

### Windows (PowerShell)

```powershell
cd examples\ui_demos\demo2_runtime_internals\inspector_demo

# 方式 1: 运行简化演示
.\simple\simple_demo

# 方式 2: 运行完整演示
.\demo2_inspector
```

### Linux/macOS

```bash
cd examples/ui_demos/demo2_runtime_internals/inspector_demo

# 方式 1: 运行简化演示
./simple/simple_demo

# 方式 2: 运行完整演示
./demo2_inspector

# 方式 3: 使用 Makefile
make run-simple   # 简化演示
make run          # 完整演示
```

## 📺 演示截图说明

### 简化演示界面

```
╔══════════════════════════════════════════════════╗
║     UI Inspector Feature Demonstration              ║
╚══════════════════════════════════════════════════╝

┌─ Performance ──────────────────────────────────┐
│ FPS: 60.0                                        │
│ Memory: 2.5 MB                                   │
└─────────────────────────────────────────────────┘

┌─ Layout Diagnostics ───────────────────────────┐
│ ✓ No layout problems detected                   │
└─────────────────────────────────────────────────┘

┌─ Layout Tree ────────────────────────────────────┐
│ Total Nodes: 45 | Max Depth: 4 | Leaf Nodes: 12  │
└─────────────────────────────────────────────────┘

┌─ Demo Content ───────────────────────────────────┐
│                                                 │
│  [Hello, UI Inspector!]                         │
│  [This is a demonstration.]                      │
│  [Button 1] [Button 2]                          │
│                                                 │
└─────────────────────────────────────────────────┘

Instructions:
  • Observe real-time performance metrics above
  • Layout diagnostics run automatically each frame
  • Tree view shows the VNode structure
  • Press Ctrl+C to exit and see the final report
```

### 完整演示界面

```
┌─ Runtime Scheduling Pipeline Visualization ─┐
│                                              │
│  [Event]    [setState]    [Scheduler] ...  │
│      ↓            ↓             ↓           │
└──────────────────────────────────────────────┘

┌─ Statistics ───────────────────────────────┐
│ Events:     0    Renders:    0    Buffers: 0│
└──────────────────────────────────────────────┘

┌─ Controls ─────────────────────────────────┐
│ [1] Event [2]setState [3]Scheduler         │
│ [4] Render [5]Reconcile [6] Layout          │
│ [7] Paint [0] Idle [I] Toggle Inspector   │
└──────────────────────────────────────────────┘

┌─ UI INSPECTOR ────────────────┬─────────────┐
│ ┌─ Performance ─             │             │
│ │ FPS: 60.0 | Mem: 2.5 MB    │             │
│ └────────────────────────────┘             │
│ ┌─ Diagnostics ──            │ Explanation │
│ │ Problems: 0 ERR, 1 WARN    │             │
│ └────────────────────────────┘             │
│ ┌─ Layout Tree ────           │             │
│ │ Nodes: 152 | Depth: 8      │             │
│ └────────────────────────────┘             │
│ Instructions:                │             │
│   F12: Toggle                │             │
│   Tab: Next element          │             │
└───────────────────────────────┴─────────────┘
```

## 🎯 核心功能演示

### 1. 性能监控

**观察目标**:
- FPS 帧率
- 内存使用
- 渲染时间

**操作**:
1. 运行程序
2. 观察右上角的 "Performance" 面板
3. 点击各个按钮触发渲染
4. 看到 FPS 和内存实时更新

**预期输出**:
```
FPS: 60.0
Memory: 2.5 MB
```

### 2. 布局诊断

**观察目标**:
- 布局问题检测
- 错误和警告数量

**操作**:
1. 运行程序
2. 观察 "Diagnostics" 面板
3. 如果有布局问题，会显示具体数量

**预期输出**:
```
✓ No layout problems detected
```
或
```
Problems: 1 ERR, 2 WARN
```

### 3. 树视图

**观察目标**:
- VNode 节点总数
- 最大深度
- 叶节点数

**操作**:
1. 运行程序
2. 观察 "Layout Tree" 面板
3. 数据实时更新

**预期输出**:
```
Total Nodes: 152 | Max Depth: 8 | Leaf Nodes: 45
```

### 4. 报告导出

**观察目标**:
- 完整的 Markdown 报告
- JSON 格式数据
- 性能统计

**操作**:
1. 运行简化演示
2. 按 `Ctrl+C` 退出
3. 查看终端输出的报告

**预期输出**:
```markdown
# UI Inspector Report

**Generated**: 2025-02-08 15:30:45

## Performance
- **Total Frames**: 1234
- **FPS**: 60.0
- **Memory**: 2.5 MB
...
```

## 🔧 高级功能

### 查看完整树结构

虽然简化演示只显示统计，但你可以在代码中添加：

```go
// 获取完整树结构
treeOutput := treeView.FormatTree()
fmt.Println(treeOutput)
```

输出：
```
┌─ Layout Tree ─────────────────────────────────
└── 📦LayoutNode
│  ├── 🔵ButtonVNode(Button 1)
│  ├── 🔵ButtonVNode(Button 2)
│  └── 📝ElementVNode(Hello)
└─────────────────────────────────────────────┘
```

### 详细问题分析

```go
// 获取所有问题
problems := diagnostics.GetProblems()

// 只看错误
errors := diagnostics.GetProblemsBySeverity(inspector.SeverityError)

// 格式化输出
fmt.Println(diagnostics.FormatProblems())
```

### 编辑属性

```go
editor := inspector.NewPropertyEditor()

// 修改按钮 flex 值
err := editor.EditFlex(button, 2)
if err != nil {
    log.Printf("Edit failed: %v", err)
}

// 查看编辑历史
fmt.Println(editor.FormatHistory())
```

## 📊 性能基准

在标准硬件上的预期性能：

| 指标 | 预期值 | 说明 |
|------|--------|------|
| FPS | 60 ± 5 | 正常渲染速度 |
| 内存 | < 5 MB | 程序占用 |
| 渲染时间 | 16-20 ms | 每帧耗时 |
| 节点数 | ~150 | demo2 的树大小 |

## 🐛 常见问题

### Q: Inspector 不显示

A: 点击 `[I] Toggle Inspector` 按钮或设置环境变量：
```bash
export TUI_INSPECTOR=true
./demo2_inspector
```

### Q: 性能数据为空

A: 需要等待几帧让监控系统收集数据。通常 1-2 秒后就会显示。

### Q: 编译错误

A: 确保在正确目录并更新依赖：
```bash
cd examples/ui_demos/demo2_runtime_internals/inspector_demo
go mod tidy
go build
```

### Q: 窗口太小

A: 增加终端窗口大小：
```bash
# 推荐 120x40 或更大
go run simple/main.go
```

## 📚 下一步

1. **阅读文档**
   - [UI Inspector README](../../../../internal/inspector/README.md)
   - [Phase 6 报告](../../../../internal/inspector/PHASE6_REPORT.md)
   - [Phase 7 报告](../../../../internal/inspector/PHASE7_REPORT.md)

2. **集成到你的应用**
   - 参考本演示的代码
   - 复制初始化代码
   - 添加 Inspector UI 面板

3. **探索功能**
   - 尝试不同的 Inspector 功能
   - 生成和查看报告
   - 实时编辑属性观察效果

## 🎓 学习路径

### 初级（开始使用）

1. ✅ 运行简化演示
2. ✅ 观察性能面板
3. ✅ 查看诊断信息
4. ✅ 理解树统计

### 中级（日常调试）

1. ✅ 集成到你的应用
2. ✅ 使用 Toggle 按钮
3. ✅ 分析布局问题
4. ✅ 导出报告

### 高级（深度分析）

1. ✅ 实时属性编辑
2. ✅ 树搜索和过滤
3. ✅ 自定义报告格式
4. ✅ 性能优化分析

---

**享受使用 UI Inspector 调试你的 TUI 应用！** 🚀
