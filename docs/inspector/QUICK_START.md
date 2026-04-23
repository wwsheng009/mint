# 布局诊断工具 - 快速使用指南

## 30 秒快速开始

```bash
# 1. 运行 Demo
cd examples/ui_demos/demo2_runtime_internals/inspector_overlay
TUI_INSPECTOR=true go run main.go

# 2. 按 F12 打开 Inspector

# 3. 按 5 切换到 Layout tab

# 4. 查看布局诊断信息！
```

---

## 详细步骤

### 第 1 步：打开 Inspector

按 **F12** 键打开 Inspector overlay。

你会看到：
```
┌─────────────────────────────────────┐
│ ╔═ INSPECTOR ═╗                   │
│ F12:关闭 | Alt+H/J/K/L:移动       │
│ [Elements] Console(2) ... Layout(5) │ ← 数字键切换 tab
└─────────────────────────────────────┘
```

### 第 2 步：切换到 Elements tab

按 **1** 键切换到 Elements tab。

### 第 3 步：选择一个节点

1. 使用 **↑↓** 键导航节点树
2. 按 **Enter** 选中一个节点

### 第 4 步：查看布局诊断

1. 按 **5** 键切换到 **Layout** tab
2. 查看选中节点的详细布局信息

---

## Tab 快捷键

| 快捷键 | Tab | 说明 |
|--------|-----|------|
| **1** | Elements | 元素树 |
| **2** | Console | 控制台 |
| **3** | Performance | 性能 |
| **4** | Diagnostics | 诊断 |
| **5** | Layout | **布局诊断** ← 新增！ |
| **6** | Network | 网络 |
| **Tab** | - | 下一个 tab |
| **Shift+Tab** | - | 上一个 tab |

---

## 导航快捷键

| 快捷键 | 功能 |
|--------|------|
| F12 | 打开/关闭 Inspector |
| ↑↓ | 在 Elements tab 中导航节点 |
| Enter | 选中节点 |
| Escape | 取消选择 |
| Alt+H | 向左移动窗口 |
| Alt+L | 向右移动窗口 |
| Alt+J | 向下移动窗口 |
| Alt+K | 向上移动窗口 |

**注意**: Alt+L 用于移动窗口，不是切换 tab！

---

## 诊断信息解读

### 基本输出

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

### 字段说明

| 字段 | 说明 | 示例 |
|------|------|------|
| **Constraints** | 传递给此节点的约束 | W[0:76] H[0:20] = 宽0-76，高0-20 |
| **Measured** | 实际测量的尺寸 | 76x20 = 宽76，高20 |
| **Propagated** | 约束如何从 props 转换 | Height(20) → constraints |
| **Props** | 节点的属性设置 | height=20 |
| **Children** | 子节点数量 | 5 个子节点 |

---

## 常见问题诊断

### ❌ 问题 1：尺寸超出约束

```
❌ Height 30 exceeds MaxHeight 20
⚠️  Has Height(20) prop but measured size is 80x30
```

**原因**：组件测量为 30 高，但约束最大为 20。

**可能原因**：
1. VStack 没有正确限制子组件高度
2. 子组件没有实现虚拟滚动
3. 约束没有正确传播

**解决方法**：
1. 检查父容器是否设置了 `.Height(n)` prop
2. 如果是 TreeView，确认虚拟滚动已启用
3. 检查约束传播链

### ✅ 正常情况：虚拟滚动

```
TreeView:
  Total Lines: 34
  Rendered: 18
  Virtual Scrolling: 18/34 lines ✅
```

**说明**：虚拟滚动正常工作，只渲染 18 行（可见部分），而不是全部 34 行。

**性能提升**：约 47%

### ⚠️ 警告：虚拟滚动未启用

```
TreeView:
  Total Lines: 34
  Rendered: 34
  Virtual Scrolling: 34/34 lines ⚠️ (all lines rendered)
```

**问题**：渲染了所有行，虚拟滚动未生效。

**解决方法**：
1. 检查 TreeView 是否收到高度约束
2. 确认父容器有 `.Height(n)` prop
3. 检查 `viewportHeight` 是否正确设置

---

## 实用技巧

### 技巧 1：对比父子和子节点

选择父节点查看约束，然后选择子节点查看接收到的约束：

```
Parent (VStack):
  Props: height=20
  Constraints: W[0:80] H[0:25]
  Measured: 80x20

Child (TreeView):
  Constraints: W[0:80] H[0:20]  ← 正确继承了父节点的高度
  Measured: 80x18
  Virtual Scrolling: 18/34 ✅   ← 虚拟滚动工作正常
```

### 技巧 2：追踪约束传播

逐层向上检查约束来源：

```
Root
  └─ Constraints: W[0:120] H[0:40]
      └─ VStack (Height=40)
          └─ Constraints: W[0:120] H[0:40]
              └─ Tabs (Height=25)
                  └─ Constraints: W[0:120] H[0:25]
                      └─ VStack (Height=20)
                          └─ Constraints: W[0:120] H[0:20]  ← 找到约束来源！
```

### 技巧 3：验证虚拟滚动

对于 TreeView：
1. 查看 "Virtual Scrolling" 行
2. 如果显示 `18/34 ✅` = 正常
3. 如果显示 `34/34 ⚠️` = 需要检查约束

---

## 示例场景

### 场景：TreeView 性能问题

**问题**：TreeView 有 100 行，滚动很慢。

**诊断步骤**：

1. **F12** 打开 Inspector
2. **1** 切换到 Elements tab
3. **↑↓** 导航找到 TreeView
4. **Enter** 选中 TreeView
5. **5** 切换到 Layout tab
6. 查看诊断输出

**可能的结果**：

❌ **坏情况**：
```
Constraints: W[0:80] H[Infinity]  ← 无高度约束！
Measured: 80x100
Virtual Scrolling: 100/100 ⚠️   ← 渲染所有行
```

✅ **好情况**：
```
Constraints: W[0:80] H[0:20]     ← 有高度约束
Measured: 80x18
Virtual Scrolling: 18/100 ✅     ← 只渲染可见行
```

**解决**：确保父容器设置了 `.Height(n)` prop。

---

## 常见问题

### Q: Layout tab 显示 "No VNode to analyze"

**A**: 需要先在 Elements tab 中选中一个节点（按 Enter）。

### Q: 为什么有些组件没有 Measure() 方法?

**A**: 简单的叶子组件（如 TextVNode）不需要实现 Measure()，它们的尺寸由父容器决定。这是正常的。

### Q: 如何理解 Constraints 格式?

**A**:
- `W[min:max]` = 宽度约束范围
- `H[min:max]` = 高度约束范围
- `Infinity` = 无边界
- 示例：`W[0:80] H[0:20]` = 最大宽80，最大高20

### Q: Alt+L 不起作用？

**A**: Alt+L 用于移动窗口，不是切换 tab。要切换到 Layout tab，请按 **5** 键。

---

## 更多资源

- **详细功能说明**: `docs/inspector/features/LAYOUT_DIAGNOSTIC_TAB.md`
- **集成文档**: `docsArchive/integration/LAYOUT_DIAGNOSTIC_INTEGRATION.md`
- **验证报告**: `docsArchive/integration/VERIFICATION_REPORT.md`
- **独立工具**: `tools/layout_diagnostic.go`

---

## 总结

布局诊断工具让你的 TUI 调试更简单：

1. ✅ **快速定位**：F12 → 5 即可查看
2. ✅ **详细诊断**：完整的约束和尺寸信息
3. ✅ **自动检测**：自动发现常见问题
4. ✅ **性能优化**：验证虚拟滚动是否生效

**开始使用**: 按 F12 → 1 (Elements) → Enter (选择) → 5 (Layout) → 享受强大的布局诊断能力！🚀
