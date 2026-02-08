# Demo2: Runtime Internals + Inspector Overlay

**带有 UI Inspector 覆盖层的运行时内部演示**

---

## ✨ 功能特性

- ✅ **Runtime Pipeline 可视化** - 展示完整的渲染流程
- ✅ **UI Inspector 集成** - 按 F12 或点击 [I] 按钮打开
- ✅ **蓝色背景标题栏** ⭐ NEW - 视觉上易于区分
- ✅ **键盘移动面板** ⭐ NEW - Alt+H/J/K/L 或方向键
- ✅ **Layer 系统支持** - Inspector 作为覆盖层（z-index: 4）
- ✅ **完整交互** - 应用和 Inspector 都保持可交互
- ✅ **快捷键丰富** - F12/Ctrl+D 切换，1-5 切换标签页

---

## 🚀 快速开始

```bash
# 编译
go build -o demo2_inspector.exe main.go

# 运行
./demo2_inspector.exe

# 按 F12 打开 Inspector
# 使用 Alt+H/J/K/L 移动面板
# 使用 1-5 切换标签页
```

---

## 🎮 Inspector 操作指南

### 打开/关闭 Inspector

| 快捷键 | 功能 |
|--------|------|
| **F12** 或 **Ctrl+D** | 切换 Inspector 显示 |
| **[I] 按钮** | 点击按钮切换 |

### 移动面板 ⭐ NEW

| 快捷键 | 功能 | 移动距离 |
|--------|------|---------|
| **Alt+H** 或 **Alt+←** | 向左移动 | -2 X |
| **Alt+L** 或 **Alt+→** | 向右移动 | +2 X |
| **Alt+K** 或 **Alt+↑** | 向上移动 | -1 Y |
| **Alt+J** 或 **Alt+↓** | 向下移动 | +1 Y |

**提示**: 水平移动幅度更大（2格），垂直移动较小（1格）

### 切换标签页

| 快捷键 | 标签页 |
|--------|--------|
| **1** | Elements（元素树） |
| **2** | Console（控制台） |
| **3** | Performance（性能） |
| **4** | Diagnostics（诊断） |
| **5** | Network（网络） |

---

## 🎨 视觉效果

### Inspector 标题栏（蓝色背景）

```
        ╔═ INSPECTOR ═╗           ← 黄色粗体，蓝色背景
        F12:关闭 | 1-5:标签页       ← 白色文字，蓝色背景
        Alt+H/J/K/L:移动面板       ← 亮白色文字，蓝色背景
```

### Inspector 面板

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ Runtime Scheduling Pipeline Visualization                                │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  Events:     0    Renders:     0    Buffers:     0                         │
│                                                                              │
│  ╔═ INSPECTOR ═╗                   ← 蓝色背景，易于识别                    │
│  F12:关闭 | 1-5:标签页                                                       │
│  Alt+H/J/K/L:移动面板                                                      │
│  ┌──────────────────────┐                                                │
│  │ Elements Tree         │                                                │
│  │ - AppRoot             │                                                │
│  │   - VStack            │                                                │
│  │     - HeaderPanel     │                                                │
│  └──────────────────────┘                                                │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 🐛 调试

### 查看面板移动

```bash
# 启用 Inspector 详细输出
TUI_INSPECTOR_VERBOSE=true ./demo2_inspector.exe

# 移动面板时会看到：
# [Inspector] Moved by (-2, 0) to (78, 5)
# [APP] Inspector moved to (78, 5)
```

### 完整诊断

```bash
# 使用测试脚本（Linux/Mac）
./test.sh
# 选择 5 - 完整诊断

# 或手动运行（Windows）
test.bat
# 选择 5 - 完整诊断

# 或手动命令
TUI_DEBUG=true TUI_DEBUG_UI=true TUI_INSPECTOR_VERBOSE=true \
TUI_LAYER_DEBUG=true TUI_DEBUG_RENDERING=true \
./demo2_inspector.exe 2>&1 | tee demo2_debug.log
```

---

## ⚠️ 故障排除

### Inspector 无法显示？

**诊断**:
```bash
TUI_DEBUG=true TUI_INSPECTOR_VERBOSE=true ./demo2_inspector.exe
```

**检查**: `[DEMO2] Inspector enabled: true`

### 面板无法移动？

**检查**:
1. Inspector 必须先打开（按 F12）
2. 确认使用 Alt+组合键（不是单独按 H/J/K/L）
3. 查看调试输出：
   ```bash
   TUI_INSPECTOR_VERBOSE=true ./demo2_inspector.exe
   ```
   应该看到：`[Inspector] Moved by (...) to (...)`

### 快捷键不响应？

**可能原因**:
1. 终端不支持 Alt+组合键
2. 其他应用占用了快捷键

**解决方案**:
- 尝试使用不同的组合键
- 检查终端设置

---

## 📚 文档

- `INSPECTOR_ENHANCEMENTS.md` - ⭐ Inspector 优化详细说明
- `DEBUG_GUIDE.md` - 完整调试指南
- `LAYER_SYSTEM_ARCHITECTURE.md` - Layer 系统架构
- `DEMO2_INSPECTOR_FIX.md` - 问题修复报告

---

## ✅ 验证成功的标准

1. **启动成功**
   ```bash
   UI Inspector ready - Press [I] button or F12/Ctrl+D to toggle
   [APP] Inspector shortcuts registered: F12, Ctrl+D (toggle)
   [APP] Panel movement: Alt+H/J/K/L or Alt+Arrow keys
   [APP] Tab switching: 1-5
   ```

2. **打开 Inspector**（按 F12）
   - Inspector 面板显示
   - 蓝色标题栏清晰可见
   - 不影响应用布局

3. **移动面板**（Alt+H/J/K/L）
   - 面板移动流畅
   - 位置立即生效
   - 不会移出屏幕

4. **切换标签页**（按 1-5）
   - 标签页快速切换
   - 内容正确更新

---

**版本**: 2.0
**更新**: 2025-02-08
**新增**: 背景色区分 + 键盘移动功能
**状态**: ✅ 已完成并测试
