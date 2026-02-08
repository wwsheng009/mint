# Demo2 with UI Inspector Integration

这个版本的 demo2 集成了完整的 UI Inspector 调试工具，提供实时的性能监控、布局诊断、树视图和属性编辑功能。

## 功能特性

### 1. 实时性能监控
- **FPS 显示**: 实时帧率显示
- **内存使用**: 堆内存分配监控
- **渲染时间**: 平均每帧渲染时间

### 2. 布局问题诊断
- **约束冲突**: MinWidth > MaxWidth 等约束问题
- **尺寸不一致**: Bounds 与 Layout 尺寸不匹配
- **溢出检测**: 内容超出边界
- **严重程度分级**: Critical, Error, Warning, Info

### 3. 布局树视图
- **节点统计**: 总节点数、最大深度
- **树结构**: 完整的 VNode 树结构
- **路径追踪**: 点分隔的节点路径

### 4. 选中元素信息
- **元素类型**: 当前选中元素的类型
- **位置和尺寸**: 元素的 bounds 信息
- **标签显示**: 元素的 label 或 tag

## 运行方式

### 方式 1: 启用 Inspector

```bash
# 设置环境变量启用 inspector
export TUI_INSPECTOR=true

# 运行 demo2
cd examples/ui_demos/demo2_runtime_internals
go run main_with_inspector.go
```

### 方式 2: 使用集成版本（推荐）

```bash
# 直接运行，程序内置 [I] Toggle Inspector 按钮
cd examples/ui_demos/demo2_runtime_internals
go run main_with_inspector.go

# 在程序中点击 [I] Toggle Inspector 按钮来切换
```

### 方式 3: 快速测试

```bash
# 一键运行脚本
./run_with_inspector.sh
```

## 使用说明

### 基本操作

1. **启动程序**
   ```bash
   go run main_with_inspector.go
   ```

2. **查看 Inspector 面板**
   - Inspector 会在右侧显示（如果启用）
   - 包含性能、诊断、树视图、选中元素四个部分

3. **切换 Inspector**
   - 点击 `[I] Toggle Inspector` 按钮开启/关闭
   - 或设置环境变量 `TUI_INSPECTOR=true`

4. **测试布局**
   - 点击各个阶段按钮：`[1] Event`, `[2] setState`, 等
   - 观察 Inspector 面板中的实时数据

### Inspector 面板说明

```
╔═ UI INSPECTOR ═╗
┌─ Performance ─
FPS: 60.0 | Mem: 2.5 MB

┌─ Diagnostics ──
Problems: 0 ERR, 2 WARN

┌─ Layout Tree ────
Nodes: 152 | Depth: 8

┌─ Selected ──────
ButtonVNode (Submit)
Pos: (10, 5) | Size: 18x1

─────────────────
F12: Toggle
Tab: Next element
```

### 快捷键

- **F12**: 切换 Inspector 开关
- **Tab**: 导航到下一个可选元素
- **Enter**: 查看元素详情
- **Esc**: 清除选择或关闭 Inspector

## 高级功能

### 1. 生成完整报告

```go
// 在代码中添加
report := globalExporter.Export(inspector.FormatMarkdown)
fmt.Println(report)
```

### 2. 保存报告到文件

```go
// 导出 Markdown 报告
mdReport := globalExporter.Export(inspector.FormatMarkdown)
os.WriteFile("inspector_report.md", []byte(mdReport), 0644)

// 导出 JSON 报告
jsonReport := globalExporter.Export(inspector.FormatJSON)
os.WriteFile("inspector_report.json", []byte(jsonReport), 0644)
```

### 3. 实时属性编辑

```go
// 修改按钮的 flex 值
globalEditor.EditFlex(button, 2)

// 修改约束
newConstraints := runtime.BoxConstraints{
    MinWidth: 10,
    MaxWidth: 50,
}
globalEditor.EditConstraints(button, newConstraints)

// 查看编辑历史
history := globalEditor.FormatHistory()
fmt.Println(history)
```

### 4. 详细的树视图

```go
// 获取完整树结构
treeOutput := globalTreeView.FormatTree()
fmt.Println(treeOutput)

// 搜索节点
buttons := globalTreeView.FindNodesByType("Button")
fmt.Printf("Found %d buttons\n", len(buttons))

// 展开所有节点
globalTreeView.ExpandAll()
```

### 5. 布局问题详细分析

```go
// 分析整个树
problems := globalDiagnostics.Analyze(rootVNode)

// 按严重程度过滤
errors := globalDiagnostics.GetProblemsBySeverity(inspector.SeverityError)

// 输出详细报告
fmt.Println(globalDiagnostics.FormatProblems())
```

## 输出示例

### 性能监控输出
```
Performance Metrics:
  Frames: 1000
  FPS: 60.0
  Last Render: 16.67 ms
  Avg Render: 16.50 ms
  Memory:
    Heap Alloc: 2.5 MB
    Heap Sys: 3.2 MB
    GC Count: 12
```

### 布局诊断输出
```
Layout Diagnostics:
  Total: 5
  Critical: 0
  Errors: 1
  Warnings: 3
  Info: 1

Details:
  [1] ERR: Zero Size
    Location: root.button
    Message: Element has zero width but natural width is 14
    Suggestion: Check layout constraints and parent sizing
```

### 树视图输出
```
┌─ Layout Tree ─────────────────────────────────
└── 📦LayoutNode
│  ├── 📦LayoutNode
│  │  ├── 🔵ButtonVNode(Event)
│  │  ├── 🔵ButtonVNode(setState)
│  │  └── 🔵ButtonVNode(Scheduler)
│  ├── 📦LayoutNode
│  └── 📝ElementVNode(Runtime Pipeline...)
└─────────────────────────────────────────────┘
```

## 集成到现有应用

### 步骤 1: 导入 inspector

```go
import "github.com/wwsheng009/mint/internal/inspector"
```

### 步骤 2: 创建全局实例

```go
var globalInspector *inspector.Inspector
var globalPerf *inspector.PerformanceAnalyzer
var globalDiagnostics *inspector.LayoutDiagnostics
var globalTreeView *inspector.TreeView
```

### 步骤 3: 初始化（在 main 函数）

```go
func main() {
    globalInspector = inspector.NewInspector()
    globalPerf = inspector.NewPerformanceAnalyzer()
    globalDiagnostics = inspector.NewLayoutDiagnostics()
    globalTreeView = inspector.NewTreeView()

    if os.Getenv("TUI_INSPECTOR") == "true" {
        globalInspector.Enable()
        globalPerf.Enable()
    }

    // ... 运行应用
}
```

### 步骤 4: 在渲染循环中使用

```go
func MyComponent() ui.VNode {
    // 性能监控
    globalPerf.StartFrame()
    defer globalPerf.EndFrame()

    // 构建内容
    content := ui.VStack(...)

    // 布局诊断
    globalDiagnostics.Analyze(content)

    // 树视图更新
    globalTreeView.SetRoot(content)

    return content
}
```

### 步骤 5: 添加 Inspector UI

```go
if inspectorEnabled {
    inspectorPanel := buildInspectorPanel()
    return ui.HStack(content, ui.Text("│"), inspectorPanel)
}
return content
```

## 故障排除

### Inspector 不显示

1. 检查环境变量是否设置：
   ```bash
   echo $TUI_INSPECTOR  # 应该输出 "true"
   ```

2. 确保按钮被点击（`[I] Toggle Inspector`）

3. 查看控制台是否有错误信息

### 性能数据为空

- 需要等待几帧让性能监控收集数据
- 确保 `globalPerf.StartFrame()` 和 `EndFrame()` 被正确调用

### 树视图为空

- 确保 `globalTreeView.SetRoot()` 被调用
- 检查传入的 VNode 不为 nil

## 相关文档

- [UI Inspector README](../../../internal/inspector/README.md)
- [Phase 6 Report: Layout Tree View](../../../internal/inspector/PHASE6_REPORT.md)
- [Phase 7 Report: Advanced Features](../../../internal/inspector/PHASE7_REPORT.md)
- [UI Inspector Design Document](../../../docs/plan/ui_inspector_design.md)

## 下一步

1. 尝试不同的 Inspector 功能
2. 生成并查看完整报告
3. 实时编辑属性观察效果
4. 集成到你自己的应用中
