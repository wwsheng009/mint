# Inspector Layout 诊断功能

## 概述

Inspector 现在包含一个**通用的布局诊断功能**，可以帮助开发者分析和调试任何 UI 组件的布局问题。

## 功能特性

### 1. 通用性

✅ **可用于任何 VNode**
- 不局限于特定组件
- 可以分析应用中的任何节点
- 支持选中节点或整个应用根节点

✅ **全面的诊断信息**
- 约束传播链分析
- 尺寸测量结果
- Props 检查
- 约束违反检测
- 性能问题识别

### 2. 使用方式

#### 通过 F12 Inspector 访问

1. **启动应用**
   ```bash
   cd examples/ui_demos/demo2_runtime_internals/inspector_overlay
   TUI_INSPECTOR=true go run main.go
   ```

2. **打开 Inspector**
   - 按 `F12` 或 `Ctrl+D` 切换 Inspector

3. **切换到 Layout tab**
   - 使用数字键 `5` 切换到 Layout 诊断 tab
   - 或者在 tab bar 中点击 "Layout"

4. **查看诊断信息**
   - 在 Elements tab 中选中一个节点（按 Enter）
   - 切换到 Layout tab 查看该节点的布局信息
   - 如果没有选中节点，会显示整个应用根节点的信息

### 3. 诊断信息说明

#### 显示的信息

```
📐 SELECTED NODE LAYOUT INFO
════════════════════════════════════════════════════════════════════════════════

[0] element (tag: vstack)
  Constraints: W[0:76] H[0:20]
  Measured:   76x20
  Propagated: Height(20) → size
  Props: height=20, flex=1
  Children: 5
════════════════════════════════════════════════════════════════════════════════
```

#### 字段说明

| 字段 | 说明 |
|------|------|
| `[0]` | 节点深度 |
| `element` | 节点类型 |
| `(tag: vstack)` | 节点标签（如果有） |
| `Constraints` | 传递给此节点的约束 (MinWidth:MaxWidth MinHeight:MaxHeight) |
| `Measured` | 实际测量的尺寸 |
| `Propagated` | 约束如何从 props 传播到尺寸 |
| `Props` | 节点的属性设置 |
| `Children` | 子节点数量 |
| `Issues` | 检测到的问题列表 |

#### 问题和警告

诊断工具会检测并显示以下问题：

- ❌ **尺寸超出约束**: Height/Width 超过 MaxHeight/MaxWidth
- ⚠️ **Props 未生效**: 设置了 Height/Width prop 但尺寸不匹配
- ⚠️ **缺少约束**: VStack/HStack 有多个子元素但没有边界约束
- ℹ️ **内容过大**: 组件尺寸过大，建议使用虚拟滚动
- ℹ️ **虚拟滚动状态**: TreeView 显示实际渲染行数 vs 总行数

## 实现细节

### 架构

```
┌─────────────────────────────────────────┐
│         F12 Inspector Overlay           │
├─────────────────────────────────────────┤
│ [Elements|Console|Perf|Diag|Layout|Net] │
├─────────────────────────────────────────┤
│                                         │
│  📐 Layout Diagnostics                  │
│  ─────────────────────                  │
│                                         │
│  Selected: Element                      │
│  ────────────────                       │
│                                         │
│  [0] element (tag: vstack)              │
│    Constraints: W[0:76] H[0:20]         │
│    Measured:   76x20                    │
│    ✅ Height(20) → size                 │
│                                         │
│  Virtual Scrolling: 18/34 lines ✅      │
│                                         │
└─────────────────────────────────────────┘
```

### 核心组件

1. **LayoutDiagnostic** (`internal/inspector/layout_diagnostic.go`)
   - 通用诊断引擎
   - 可以分析任何 VNode
   - 提供详细的约束传播信息

2. **TabLayout** (`internal/inspector/standalone_inspector.go`)
   - Inspector 中的 Layout tab
   - 显示选中节点的布局信息
   - 语法高亮显示问题

3. **诊断工具** (`tools/layout_diagnostic.go`)
   - 独立的命令行工具
   - 用于批量测试和验证
   - 不会影响应用运行

## 使用场景

### 场景 1: 调试组件尺寸问题

**问题**: 组件没有按预期显示尺寸

**步骤**:
1. 在 Elements tab 找到问题组件
2. 按 Enter 选中组件
3. 切换到 Layout tab
4. 查看 Constraints 和 Measured 字段
5. 检查是否有约束违反的警告

**示例**:
```
❌ Height 30 exceeds MaxHeight 20
⚠️  Has Height(20) prop but measured size is 80x30
```

### 场景 2: 检查虚拟滚动

**问题**: TreeView 渲染太多行，性能差

**步骤**:
1. 在 Elements tab 选中 TreeView
2. 切换到 Layout tab
3. 查找 Virtual Scrolling 信息

**示例**:
```
Virtual Scrolling: 18/34 lines ✅
```

如果显示:
```
Virtual Scrolling: 34/34 lines ⚠️ (all lines rendered)
```
说明虚拟滚动未生效，需要检查约束传播。

### 场景 3: 分析约束传播

**问题**: 不确定约束是否正确传递

**步骤**:
1. 在 Elements tab 选中父容器
2. 切换到 Layout tab 查看 Props 和 Constraints
3. 逐层向下检查子节点
4. 确认每层都正确传递了约束

**示例**:
```
Parent (VStack):
  Props: height=20
  Constraints: W[0:80] H[0:25]
  Measured: 80x20
  ✅ Height(20) → constraints

Child (TreeView):
  Constraints: W[0:80] H[0:20]
  Measured: 80x18
  Virtual Scrolling: 18/34 lines ✅
```

## 高级用法

### 独立诊断工具

对于自动化测试或批量分析，可以使用独立工具:

```bash
go run ./tools/layout_diagnostic.go
```

输出 7 个测试的完整诊断报告，包括:
- VStack 约束传播测试
- TreeView 虚拟滚动测试
- Tabs 约束遵守测试
- UpdateLines() 状态保持测试
- Inspector 模拟场景测试

### 集成到自定义 Inspector

如果你的应用有自定义 Inspector，可以集成布局诊断功能:

```go
import "github.com/wwsheng009/mint/internal/inspector"

// 创建诊断实例
diagnostic := inspector.NewLayoutDiagnostic()

// 分析节点
result := diagnostic.AnalyzeSelectedNode(yourVNode, constraints)

// 格式化输出
formatted := diagnostic.FormatSingleResult(result)
fmt.Println(formatted)
```

## 性能考虑

- 诊断操作**不会影响应用性能**
- 只在 Inspector 打开时执行
- 分析深度限制为 20 层（防止无限递归）
- 结果会被缓存，不会重复分析

## 扩展性

### 添加新的诊断规则

在 `LayoutDiagnostic.checkForIssues()` 中添加自定义规则:

```go
// 检查特定模式
if yourCondition {
    result.Issues = append(result.Issues,
        "⚠️  Your custom warning message")
}
```

### 自定义显示格式

修改 `LayoutDiagnostic.FormatSingleResult()` 来自定义输出格式。

## 最佳实践

1. **先检查约束传播**
   - 确认父容器正确设置了约束
   - 检查 Props 是否正确应用

2. **关注虚拟滚动**
   - TreeView 应该只渲染可见行
   - 如果渲染所有行，性能会下降

3. **使用 Elements tab 定位问题**
   - 先找到问题节点
   - 再用 Layout tab 分析原因

4. **对比工作正常的组件**
   - 分析正常工作的组件
   - 对比约束设置
   - 找出差异

## 常见问题

### Q: Layout tab 显示 "No VNode to analyze"

**A**: 你需要先在 Elements tab 中选中一个节点（按 Enter），或者确保应用有 root 节点。

### Q: 诊断信息显示不完整

**A**: 诊断信息会限制行数以适应 overlay 高度。完整的诊断信息可以通过独立工具 `tools/layout_diagnostic.go` 查看。

### Q: 为什么有些组件没有 Measure() 方法?

**A**: 简单的叶子组件（如 TextVNode）不需要实现 Measure()，它们的尺寸由父容器决定。这是正常的。

### Q: 如何理解 Constraints?

**A**:
- `W[min:max]` = 宽度约束范围
- `H[min:max]` = 高度约束范围
- `Infinity` 表示无边界
- 例如 `W[0:80] H[0:20]` 表示最大宽80，最大高20

## 总结

Layout 诊断功能是一个**强大且通用的工具**，可以帮助你:
- ✅ 快速定位布局问题
- ✅ 理解约束传播机制
- ✅ 优化虚拟滚动性能
- ✅ 验证组件尺寸设置

它不仅限于 Inspector 本身，还可以用于诊断**任何 TUI 应用**的布局问题。
