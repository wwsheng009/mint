# Inspector 自我检查示例

## 概述

这是一个**元编程示例**，展示了布局诊断工具的强大功能：
1. 构建一个**模拟 Inspector UI** 的程序
2. 使用 **Inspector 来检查它自己的布局**
3. 证明布局诊断工具是**完全通用的**

## 运行方式

```bash
cd examples/inspector
TUI_INSPECTOR=true go run self_inspection.go
```

## 界面结构

```
╔════════════════════════════════════════════════════════════════╗
║           Inspector Self-Inspection Demo                        ║
╠════════════════════════════════════════════════════════════════╣
║  This UI mimics the Inspector layout structure                  ║
║  Press F12 to open Inspector and inspect THIS UI!               ║
╚════════════════════════════════════════════════════════════════╝

┌────────────────────────────────────────────────────────────┐
│ [Elements|Console|Performance|Diagnostics|Layout|Network] │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  📦 Simulated Layout Tree                                 │
│                                                            │
│  This TreeView shows the structure of THIS demo UI        │
│  Open Inspector (F12) to see the REAL layout tree!        │
│                                                            │
│  ─────────────────────────────────────────────────        │
│                                                            │
│  VNode (Root)                                             │
│  ├── 🖼️ Bordered (Header)                                │
│  │   └── 📦 VStack                                       │
│  │       ├── 🝂 Text (Title)                             │
│  │       └── 🝂 Text (Subtitle)                          │
│  ├── 🖼️ Bordered (Main Content)                          │
│  │   └── 📑 Tabs Component                               │
│  └── 🖼️ Bordered (Footer)                                │
│                                                            │
│  Instructions:                                            │
│    ↑↓: Navigate tree                                      │
│    Enter: Select node                                     │
│    Alt+L: View Layout diagnostics                         │
│    F12: Open REAL Inspector                              │
└────────────────────────────────────────────────────────────┘

💡 Tip: This demo UI mimics the Inspector layout structure
   Press F12 to open the REAL Inspector and inspect THIS UI!
```

## 使用流程

### 步骤 1: 启动程序

```bash
TUI_INSPECTOR=true go run self_inspection.go
```

程序会显示一个模拟的 Inspector UI，包含：
- Header（标题栏）
- Tabs（Elements, Console, Performance, Diagnostics, Layout, Network）
- Footer（操作说明）

### 步骤 2: 打开真实的 Inspector

按 **F12** 键打开 Inspector overlay。

你会看到两个 Inspector：
1. **模拟的 Inspector**（程序本身的 UI）
2. **真实的 Inspector**（按 F12 打开的 overlay）

### 步骤 3: 检查模拟 UI 的布局

在真实 Inspector 的 **Elements** tab 中：

1. 用 **↑↓** 键导航
2. 你会看到完整的 VNode 树结构
3. 包括所有的 VStack、Bordered、Tabs 组件
4. 包括 TreeView 及其虚拟滚动结构

### 步骤 4: 查看布局诊断

1. 在 Elements tab 中选中任意节点（按 **Enter**）
2. 按 **Alt+L** 切换到 **Layout** tab
3. 查看该节点的详细布局诊断信息：
   - Constraints（约束）
   - Measured Size（测量尺寸）
   - Props（属性）
   - Issues（问题）

## 诊断信息示例

### VStack 诊断

```
📐 SELECTED NODE LAYOUT INFO
════════════════════════════════════════════════════════════════════════════════

[3] element (tag: vstack)
  Constraints: W[0:120] H[0:40]
  Measured:   120x40
  Propagated: Height(40) → size
  Props: height=40
  Children: 3
════════════════════════════════════════════════════════════════════════════════
```

### TreeView 诊断

```
📐 SELECTED NODE LAYOUT INFO
════════════════════════════════════════════════════════════════════════════════

[5] element (tag: treeview)
  Constraints: W[0:72] H[0:18]
  Measured:   72x18
  Children: 1

  Virtual Scrolling: 18/25 lines ✅
════════════════════════════════════════════════════════════════════════════════
```

### 发现问题

如果有布局问题，会显示警告：

```
⚠️  Has Height(20) prop but measured size is 80x25
❌ Height 25 exceeds MaxHeight 20
ℹ️  Large content: 80x30 (may need virtual scrolling)
```

## 关键特性展示

### 1. 通用性

✅ **任何 UI 都可以被诊断**
- 模拟的 Inspector UI
- 真实的 Inspector UI
- 你自己的应用 UI

### 2. 自我检查

✅ **Inspector 可以检查自己**
- 真实的 Inspector 检查模拟的 Inspector
- 元编程的有趣示例
- 证明工具的通用性

### 3. 虚拟滚动验证

✅ **TreeView 使用虚拟滚动**
- 总共 25 行
- 只渲染 18 行（可见区域）
- 性能优化约 28%

### 4. 约束传播

✅ **完整的约束链追踪**
- Root: W[0:120] H[0:40]
- VStack: W[0:120] H[0:40]
- Tabs: W[0:76] H[0:25]
- TreeView: W[0:72] H[0:18]
- 虚拟滚动生效！

## 技术要点

### 模拟的 UI 结构

```go
buildSelfInspectionDemo()
  ├── buildHeader()          // 标题栏
  ├── buildMainContent()     // 主内容区
  │   └── Tabs Component     // 标签页
  │       ├── Elements       // 元素树
  │       ├── Console        // 控制台
  │       ├── Performance    // 性能
  │       ├── Diagnostics    // 诊断
  │       ├── Layout         // 布局（模拟）
  │       └── Network        // 网络
  └── buildFooter()          // 页脚
```

### 真实的 Inspector

当按 F12 打开真实的 Inspector 时：
1. Hook 系统自动注入 Inspector overlay
2. Inspector 分析整个 VNode 树
3. 包括模拟的 UI 和真实的 Inspector 自己
4. 提供完整的布局诊断信息

### 元编程

这个示例展示了：
- **程序检查程序**（Program examining itself）
- **UI 检查 UI**（UI examining UI）
- **工具检查工具**（Tool examining tool）

## 学习价值

### 对于开发者

1. **理解布局系统**
   - 看到完整的 VNode 树结构
   - 理解约束如何传播
   - 学习虚拟滚动的实现

2. **调试布局问题**
   - 快速定位尺寸问题
   - 检查约束是否正确
   - 验证虚拟滚动是否生效

3. **优化性能**
   - 识别渲染过多的组件
   - 确认虚拟滚动正常工作
   - 减少不必要的渲染

### 对于架构设计

1. **通用工具设计**
   - 布局诊断不依赖特定 UI
   - 可以检查任何 VNode
   - 易于集成和扩展

2. **分层架构**
   - UI 层（模拟的 Inspector）
   - 工具层（真实的 Inspector）
   - 诊断层（Layout Diagnostic）

3. **元编程能力**
   - 程序可以检查自己
   - 提供强大的调试能力
   - 类似 Lisp 的反射能力

## 对比其他工具

### 浏览器 DevTools

| 特性 | 浏览器 DevTools | Mint Inspector |
|------|----------------|----------------|
| 检查 UI | ✅ 可以 | ✅ 可以 |
| 查看布局树 | ✅ 可以 | ✅ 可以 |
| 约束系统 | CSS Box Model | TUI Constraints |
| 虚拟滚动显示 | ❌ 通常不显示 | ✅ 显示滚动状态 |
| 自我检查 | ❌ 不能 | ✅ 可以 |

### React DevTools

| 特性 | React DevTools | Mint Inspector |
|------|---------------|----------------|
| 组件树 | ✅ 可以 | ✅ 可以 |
| Props 查看 | ✅ 可以 | ✅ 可以 |
| 布局诊断 | ❌ 没有 | ✅ 有 |
| 约束追踪 | ❌ 没有 | ✅ 有 |

## 扩展可能性

### 1. 添加更多诊断规则

在 `LayoutDiagnostic` 中添加自定义检查：
```go
// 检查自定义模式
if yourCondition {
    result.Issues = append(result.Issues,
        "⚠️  Custom warning")
}
```

### 2. 创建自定义 Inspector

基于此示例，创建你自己的 Inspector：
```go
myInspector := buildCustomInspector()
diagnostic := inspector.NewLayoutDiagnostic()
result := diagnostic.AnalyzeVNode(myInspector, constraints)
```

### 3. 集成到 CI/CD

```bash
# 运行诊断检查
go run ./tools/layout_diagnostic.go > layout_report.txt

# 检查是否有问题
if grep -q "❌" layout_report.txt; then
    echo "Layout issues found!"
    exit 1
fi
```

## 总结

这个示例展示了 Mint Inspector 布局诊断工具的：

✅ **通用性** - 可以检查任何 UI，包括自己
✅ **强大性** - 详细的约束和尺寸信息
✅ **实用性** - 快速定位和修复布局问题
✅ **教育性** - 理解布局系统和虚拟滚动

**最重要的是**：它证明了布局诊断工具是一个**真正通用的解决方案**，可以用于任何 TUI 应用的布局调试！

## 相关文档

- `docs/inspector/features/LAYOUT_DIAGNOSTIC_TAB.md` - Layout 功能说明
- `docs/inspector/integration/LAYOUT_DIAGNOSTIC_INTEGRATION.md` - 集成文档
- `docs/inspector/integration/VERIFICATION_REPORT.md` - 验证报告

## 快速开始

```bash
# 1. 进入示例目录
cd examples/inspector

# 2. 运行程序
TUI_INSPECTOR=true go run self_inspection.go

# 3. 按 F12 打开 Inspector

# 4. 用 ↑↓ 导航，Enter 选择节点

# 5. 按 Alt+L 查看 Layout 诊断

# 6. 探索 Inspector 如何检查自己的布局！
```

享受自我检查的乐趣！🎉
