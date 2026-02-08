# UI Inspector Demo

这个目录包含了集成 UI Inspector 的 demo2 版本，展示了完整的调试功能。

## 目录结构

```
inspector_demo/
├── main.go              # 完整的 demo2 + inspector 集成
├── simple/              # 简化的功能演示
│   └── main.go          # 独立的简单演示程序
├── Makefile             # 构建和运行脚本
├── run_with_inspector.sh # Shell 运行脚本
└── README_INSPECTOR.md  # 本文档
```

## 快速开始

### 方式 1: 使用 Makefile（推荐）

```bash
cd inspector_demo

# 构建并运行
make run

# 或运行简化演示
make run-simple

# 使用环境变量启用
make run-enabled
```

### 方式 2: 手动构建和运行

```bash
cd inspector_demo

# 构建完整版
go build -o demo2_inspector main.go
./demo2_inspector

# 构建简化版
cd simple
go build -o simple_demo main.go
./simple_demo
```

### 方式 3: 使用 Shell 脚本

```bash
cd inspector_demo
chmod +x run_with_inspector.sh
./run_with_inspector.sh
```

## 功能说明

### 完整版 (main.go)

集成了完整的 demo2 运行时内部可视化加上 UI Inspector：

1. **运行时管道可视化**
   - Event → setState → Scheduler → Render → Reconcile → Layout → Paint
   - 每个阶段的触发按钮
   - 实时统计信息（Events, Renders, Buffers）

2. **UI Inspector 面板**
   - 性能监控（FPS、内存）
   - 布局诊断（问题、警告）
   - 树视图统计
   - 选中元素信息

3. **交互控制**
   - `[I] Toggle Inspector` 按钮开关 inspector
   - 各个阶段的触发按钮

### 简化版 (simple/main.go)

专注于展示 UI Inspector 核心功能：

1. **性能分析**
   - 实时 FPS 显示
   - 内存使用监控
   - 渲染时间追踪

2. **布局诊断**
   - 自动问题检测
   - 严重程度分级
   - 建议修复方案

3. **树视图**
   - VNode 树统计
   - 深度和节点数

4. **报告导出**
   - 退出时自动生成 Markdown 报告

## Makefile 目标

```bash
make              # 构建（默认）
make build        # 构建完整版
make simple       # 构建简化版
make run          # 构建并运行完整版
make run-simple   # 构建并运行简化版
make run-enabled  # 运行并启用 TUI_INSPECTOR
make clean        # 清理构建产物
make help         # 显示帮助
```

## 环境变量

### TUI_INSPECTOR

启用 UI Inspector：

```bash
export TUI_INSPECTOR=true
go run main.go
```

或在运行时设置：

```bash
TUI_INSPECTOR=true ./demo2_inspector
```

## 使用示例

### 示例 1: 查看性能指标

运行程序后，观察 Inspector 面板中的性能部分：

```
┌─ Performance ─
FPS: 60.0 | Mem: 2.5 MB
```

### 示例 2: 检测布局问题

点击各个阶段按钮触发渲染，观察诊断输出：

```
┌─ Diagnostics ──
Problems: 1 ERR, 2 WARN
```

### 示例 3: 查看树结构

查看树视图统计：

```
┌─ Layout Tree ────
Nodes: 152 | Depth: 8
```

### 示例 4: 导出报告

程序退出时会自动生成 Markdown 格式的完整报告。

## 故障排除

### 编译错误

如果遇到编译错误，确保在正确的目录：

```bash
cd examples/ui_demos/demo2_runtime_internals/inspector_demo
```

### Inspector 不显示

1. 点击 `[I] Toggle Inspector` 按钮
2. 或设置环境变量 `TUI_INSPECTOR=true`

### 性能数据为空

需要等待几帧让监控收集数据。

## 集成到你的应用

### 步骤 1: 导入包

```go
import "github.com/wwsheng009/mint/internal/inspector"
```

### 步骤 2: 创建实例

```go
var (
    insp        = inspector.NewInspector()
    perf        = inspector.NewPerformanceAnalyzer()
    diagnostics = inspector.NewLayoutDiagnostics()
    treeView    = inspector.NewTreeView()
)
```

### 步骤 3: 在渲染中使用

```go
func MyComponent() ui.VNode {
    // 性能追踪
    perf.StartFrame()
    defer perf.EndFrame()

    // 构建内容
    content := ui.VStack(...)

    // 诊断分析
    diagnostics.Analyze(content)

    // 更新树
    treeView.SetRoot(content)

    return content
}
```

### 步骤 4: 显示 Inspector UI

```go
if showInspector {
    inspectorPanel := buildInspectorPanel()
    return ui.HStack(content, ui.Text("|"), inspectorPanel)
}
return content
```

## 相关文档

- [UI Inspector README](../../../../../../internal/inspector/README.md)
- [Phase 6: Layout Tree View](../../../../../../internal/inspector/PHASE6_REPORT.md)
- [Phase 7: Advanced Features](../../../../../../internal/inspector/PHASE7_REPORT.md)
- [Design Document](../../../../../../docs/plan/ui_inspector_design.md)

## 高级用法

### 自定义报告格式

```go
exporter := inspector.NewReportExporter(insp)

// Markdown
md := exporter.Export(inspector.FormatMarkdown)

// JSON
json := exporter.Export(inspector.FormatJSON)

// Text
text := exporter.Export(inspector.FormatText)

// Quick Summary
summary := exporter.QuickSummary()
```

### 属性编辑

```go
editor := inspector.NewPropertyEditor()

// 编辑 flex
editor.EditFlex(button, 2)

// 编辑约束
editor.EditConstraints(button, newConstraints)

// 查看历史
fmt.Println(editor.FormatHistory())
```

### 树搜索

```go
// 按类型搜索
buttons := treeView.FindNodesByType("Button")

// 按标签搜索
results := treeView.FindNodesByLabel("Submit")

// 按路径查找
node := treeView.FindNodeByPath("root.header.button")
```

## 反馈和问题

如有问题或建议，请查看：
- GitHub Issues
- 项目文档
- 或提交 Pull Request

---

**享受使用 UI Inspector 调试你的 TUI 应用！**
