# Inspector 布局诊断工具 - 快捷键说明

## ⚠️ 重要更新

之前文档中提到的 `Alt+L` 快捷键**不正确**，因为 Alt+L 已被占用（用于向右移动窗口）。

## ✅ 正确的快捷键

### Tab 切换

| 快捷键 | Tab | 说明 |
|--------|-----|------|
| **1** | Elements | 元素树（显示 VNode 结构） |
| **2** | Console | 控制台 |
| **3** | Performance | 性能监控 |
| **4** | Diagnostics | 诊断信息 |
| **5** | Layout | **布局诊断** ← 按这个键！ |
| **6** | Network | 网络 |
| **Tab** | - | 循环切换到下一个 tab |
| **Shift+Tab** | - | 循环切换到上一个 tab |

### 窗口移动

| 快捷键 | 功能 |
|--------|------|
| **Alt+H** | 向左移动窗口 |
| **Alt+L** | 向右移动窗口 ⚠️ 不是切换 tab！ |
| **Alt+K** | 向上移动窗口 |
| **Alt+J** | 向下移动窗口 |

### 节点导航（在 Elements tab 中）

| 快捷键 | 功能 |
|--------|------|
| **↑↓** | 上下导航节点 |
| **Enter** | 选中节点 |
| **E** | 展开/折叠节点 |
| **PgUp/PgDn** | 快速滚动 |
| **Home/End** | 跳到顶部/底部 |

### 其他快捷键

| 快捷键 | 功能 |
|--------|------|
| **F12** | 打开/关闭 Inspector |
| **Ctrl+D** | 按键调试模式 |
| **Escape** | 取消选择或关闭 |

---

## 使用流程

### 查看布局诊断的正确步骤

1. **F12** - 打开 Inspector
2. **1** - 切换到 Elements tab
3. **↑↓** - 导航到你想要的节点
4. **Enter** - 选中该节点（重要！）
5. **5** - 切换到 Layout tab
6. 查看布局诊断信息

---

## 示例

### 场景：诊断 TreeView 的布局

```bash
# 1. 运行程序
cd examples/ui_demos/demo2_runtime_internals/inspector_overlay
TUI_INSPECTOR=true go run main.go

# 2. 等待程序启动

# 3. 按键操作：
F12    → 打开 Inspector
1      → 切换到 Elements tab
↑↓     → 找到 TreeView 节点
Enter  → 选中 TreeView
5      → 切换到 Layout tab
查看布局信息
```

---

## 常见错误

### ❌ 错误：按 Alt+L 试图切换到 Layout tab

**问题**：Alt+L 是用于移动窗口的，不是切换 tab。

**正确做法**：按 **5** 键切换到 Layout tab。

### ❌ 错误：没有选中节点就按 5

**问题**：Layout tab 显示 "No VNode to analyze"

**正确做法**：
1. 先按 **1** 切换到 Elements tab
2. 用 **↑↓** 找到节点
3. 按 **Enter** 选中
4. 再按 **5** 查看 Layout 信息

---

## Tab 顺序

当前的 tab 顺序是：

```
[Elements] Console(2) Performance(3) Diagnostics(4) [Layout(5)] Network(6)
    ↑1           ↑2             ↑3                ↑4           ↑5       ↑6
```

Layout tab 在第 5 个位置，所以快捷键是 **5**。

---

## 总结

**记住**：
- ✅ Layout tab 的快捷键是 **5**，不是 Alt+L
- ✅ Alt+L 用于向右移动窗口
- ✅ 使用数字键 1-6 快速切换 tabs

现在可以正确使用布局诊断工具了！🎉
