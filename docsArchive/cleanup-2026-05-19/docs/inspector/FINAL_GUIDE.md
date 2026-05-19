# 布局诊断工具 - 最终使用指南

## ✅ 修复完成

所有问题已修复，测试通过：
- ✅ Layout tab 已添加
- ✅ 数字键 5 可以切换到 Layout tab
- ✅ Tab 显示包含数字标记
- ✅ 布局诊断功能正常工作

---

## 🚀 快速开始

### 1. 运行程序

```bash
cd examples/ui_demos/demo2_runtime_internals/inspector_overlay
TUI_INSPECTOR=true go run main.go
```

### 2. 打开 Inspector

按 **F12** 键打开 Inspector overlay。

你会看到：
```
┌──────────────────────────────────────────────┐
│ ╔═ INSPECTOR ═╗                           │
│ F12:关闭 | Alt+H/J/K/L:移动               │
│ [Elements(1)] Console(2) ... [Layout(5)]   │ ← 带数字标记
└──────────────────────────────────────────────┘
```

### 3. 切换到 Layout tab

按数字键 **5** 切换到 Layout tab。

---

## 📖 Tab 快捷键

| 快捷键 | Tab | 说明 |
|--------|-----|------|
| **1** | Elements | 元素树 - 显示 VNode 结构 |
| **2** | Console | 控制台 |
| **3** | Performance | 性能监控 |
| **4** | Diagnostics | 诊断信息 |
| **5** | **Layout** | **布局诊断** ← 按 5！ |
| **6** | Network | 网络 |

---

## 🔍 使用布局诊断

### 步骤 1：选择节点

1. 按 **1** 切换到 Elements tab
2. 使用 **↑↓** 键导航节点树
3. 按 **Enter** 选中一个节点

### 步骤 2：查看布局信息

1. 按 **5** 切换到 Layout tab
2. 查看选中节点的详细布局诊断

### 诊断信息示例

```
📐 SELECTED NODE LAYOUT INFO
════════════════════════════════════════════════════════════════════════════════

[3] element (tag: vstack)
  Constraints: W[0:76] H[0:20]
  Measured:   76x20
  Propagated: Height(20) → constraints
  Props: height=20
  Children: 5
════════════════════════════════════════════════════════════════════════════════
```

---

## ⚠️ 重要注意事项

### Alt+L 用于移动窗口

- **Alt+H** = 向左移动窗口
- **Alt+L** = 向右移动窗口 ⚠️ 不是切换 tab！
- **Alt+K** = 向上移动窗口
- **Alt+J** = 向下移动窗口

要切换到 Layout tab，请按 **数字键 5**，不是 Alt+L！

---

## 📊 验证测试

运行测试验证功能：

```bash
cd E:/projects/yao/wwsheng009/mint
go test -v ./internal/inspector -run TestLayoutTab
```

预期输出：
```
✅ TestLayoutTabShortcut - PASS
  ✅ Key_5_should_switch_to_Layout_tab - PASS
✅ TestLayoutTabInTabItems - PASS
✅ TestLayoutTabBuildContent - PASS
```

---

## 🎯 完整工作流程

### 诊断 TreeView 的布局问题

```
1. F12           → 打开 Inspector
2. 1             → 切换到 Elements tab
3. ↑↓            → 导航找到 TreeView
4. Enter         → 选中 TreeView
5. 5             → 切换到 Layout tab
6. 查看信息       → 确认虚拟滚动状态
```

**虚拟滚动正常**：
```
TreeView:
  Total Lines: 34
  Rendered: 18
  Virtual Scrolling: 18/34 lines ✅
```

**虚拟滚动未生效**：
```
TreeView:
  Total Lines: 34
  Rendered: 34
  Virtual Scrolling: 34/34 lines ⚠️ (all lines rendered)
```

---

## 🔧 故障排查

### 问题 1：按 5 没有反应

**可能原因**：Inspector 没有打开

**解决方法**：
1. 先按 F12 打开 Inspector
2. 再按 5 切换到 Layout tab

### 问题 2：Layout tab 显示 "No VNode to analyze"

**可能原因**：没有选中任何节点

**解决方法**：
1. 按 1 切换到 Elements tab
2. 用 ↑↓ 导航
3. 按 Enter 选中节点
4. 再按 5 回到 Layout tab

### 问题 3：Tab bar 没有显示数字

**可能原因**：使用的是旧版本代码

**解决方法**：
```bash
cd E:/projects/yao/wwsheng009/mint
git pull
go build ./internal/inspector
```

---

## 📝 总结

### 核心要点

1. ✅ **快捷键是数字键 5**，不是 Alt+L
2. ✅ **必须先选中节点**（在 Elements tab 中按 Enter）
3. ✅ **Layout tab 显示选中节点的布局信息**

### 对比

| 功能 | Elements tab | Layout tab |
|------|--------------|------------|
| 快捷键 | 1 | 5 |
| 显示内容 | VNode 树结构 | 选中节点的布局诊断 |
| 用途 | 导航和选择节点 | 查看约束、尺寸、问题 |
| 类比 | DOM 树 | Computed 样式 |

### 现在可以使用

布局诊断工具已经完全修复并验证可用！

**立即开始**：
```bash
cd examples/ui_demos/demo2_runtime_internals/inspector_overlay
TUI_INSPECTOR=true go run main.go
# F12 → 1 → ↑↓ → Enter → 5 → 享受布局诊断！
```

---

**文档位置**：
- 快速开始：`docs/inspector/QUICK_START.md`
- 快捷键说明：`docs/inspector/KEYBOARD_SHORTCUTS.md`
- 修复总结：`/docsArchive/INSPECTOR_FIX_SUMMARY.md`
- 本指南：`docs/inspector/FINAL_GUIDE.md`
