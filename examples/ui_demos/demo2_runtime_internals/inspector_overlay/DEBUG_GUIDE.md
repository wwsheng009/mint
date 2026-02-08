# Demo2 Inspector 调试指南

**使用 docs/debug 中的工具调试 Inspector 问题**

---

## 🚨 常见问题

### 问题 1: 点击 [I] Inspector 按钮没有反应

**原因**: Inspector 未启用（`Enable()` 未被调用）

**解决方案**:
```bash
# 方法 1: 查看调试输出
TUI_DEBUG=true ./demo2_inspector.exe

# 方法 2: 使用 Inspector 详细模式
TUI_INSPECTOR_VERBOSE=true ./demo2_inspector.exe

# 方法 3: 启用完整调试
TUI_DEBUG_UI=true ./demo2_inspector.exe
```

**预期输出**:
```
[DEMO2] Inspector enabled: true
[DEMO2] Inspector visible: false
[DEMO2] [I] button clicked, Inspector enabled=true, visible=false
[Inspector] Toggled: visible
[DEMO2] After toggle, Inspector visible=true
```

---

### 问题 2: F12 快捷键不工作

**原因**:
1. Framework App 未调用 `SetupInspectorShortcut()`
2. 终端不支持 F12 键

**解决方案**:
```bash
# 启用调试查看快捷键注册
TUI_DEBUG_UI=true ./demo2_inspector.exe 2>&1 | grep -i "shortcut"

# 应该看到：
# [APP] Inspector shortcuts registered: F12, Ctrl+D
```

**备用方案**: 使用 `Ctrl+D` 快捷键（更容易输入）

---

### 问题 3: Inspector 显示但不完整

**原因**: Layer 系统未正确检测到 Layer 标记

**解决方案**:
```bash
# 启用 Layer 调试
TUI_LAYER_DEBUG=true ./demo2_inspector.exe

# 启用渲染调试
TUI_DEBUG_RENDERING=true ./demo2_inspector.exe

# 组合调试
TUI_LAYER_DEBUG=true TUI_DEBUG_RENDERING=true ./demo2_inspector.exe
```

**预期输出**:
```
[PipelineRenderer] Using RenderLayers for multi-layer rendering
[RenderingPipeline] RenderLayers started
[RenderingPipeline] Layer layouts complete, rendering 2 layers
```

---

## 🔧 调试工具使用

### 1. 基础调试（TUI_DEBUG）

启用基础框架调试输出：

```bash
TUI_DEBUG=true ./demo2_inspector.exe
```

**输出内容**:
- `[APP]` - Framework App 状态
- `[DEMO2]` - Demo2 自定义调试信息
- `[Inspector]` - Inspector 状态（如果启用 verbose）

### 2. UI 调试（TUI_DEBUG_UI）

启用 UI 层的详细调试：

```bash
TUI_DEBUG_UI=true ./demo2_inspector.exe
```

**输出内容**:
- `[DEMO2]` - Demo2 状态信息
- `[APP]` - Framework 初始化信息
- 快捷键注册信息

### 3. Inspector 详细模式（TUI_INSPECTOR_VERBOSE）

启用 Inspector 内部调试：

```bash
TUI_INSPECTOR_VERBOSE=true ./demo2_inspector.exe
```

**输出内容**:
- `[Inspector] Enabled (F12 to toggle)`
- `[Inspector] Toggled: visible/hidden`
- `[Inspector] Rendering overlay`
- `[Inspector] Frame stats`

### 4. Layer 系统调试（TUI_LAYER_DEBUG）

查看 Layer 系统工作情况：

```bash
TUI_LAYER_DEBUG=true ./demo2_inspector.exe
```

**输出内容**:
- `[PipelineRenderer] hasLayers=true/false`
- `[RenderingPipeline] RenderLayers started`
- `[LayerManager] Collected N layers`
- `[PaintEngine] Painting layer N: LayerInspector`

### 5. 渲染调试（TUI_DEBUG_RENDERING）

查看渲染流程：

```bash
TUI_DEBUG_RENDERING=true ./demo2_inspector.exe
```

**输出内容**:
- `[DeclarativeNode.Paint]` - Paint 调用
- `[PipelineRenderer]` - PipelineRenderer 状态
- `[RenderingPipeline]` - RenderingPipeline 状态

### 6. 布局调试（TUI_UI_DEBUG_LAYOUT）

查看布局计算（使用 docs/debug 中的工具）：

```bash
TUI_UI_DEBUG_LAYOUT=true ./demo2_inspector.exe 2>&1 | head -100
```

**输出内容**:
- 完整的组件树
- 每个 VNode 的位置和尺寸
- Flex 布局计算详情

### 7. 组合调试（完整诊断）

同时启用多个调试模式：

```bash
# 启用所有调试
TUI_DEBUG=true TUI_DEBUG_UI=true TUI_INSPECTOR_VERBOSE=true \
TUI_LAYER_DEBUG=true TUI_DEBUG_RENDERING=true \
./demo2_inspector.exe 2>&1 | tee demo2_debug.log
```

---

## 📊 调试流程

### 步骤 1: 验证 Inspector 初始化

```bash
TUI_DEBUG=true TUI_DEBUG_UI=true ./demo2_inspector.exe 2>&1 | grep -i inspector
```

**预期输出**:
```
UI Inspector ready - Press [I] button or F12/Ctrl+D to toggle
[DEMO2] Inspector enabled: true
[DEMO2] Inspector visible: false
[APP] Inspector shortcuts registered: F12, Ctrl+D
```

**检查点**:
- ✅ `Inspector enabled: true` - Inspector 已启用
- ✅ `shortcuts registered` - 快捷键已注册

### 步骤 2: 验证按钮点击

```bash
TUI_DEBUG=true TUI_INSPECTOR_VERBOSE=true ./demo2_inspector.exe
# 点击 [I] Inspector 按钮
```

**预期输出**:
```
[DEMO2] [I] button clicked, Inspector enabled=true, visible=false
[Inspector] Toggled: visible
[DEMO2] After toggle, Inspector visible=true
```

**检查点**:
- ✅ `[I] button clicked` - 按钮点击被捕获
- ✅ `Toggled: visible` - Inspector 状态切换成功

### 步骤 3: 验证 Layer 渲染

```bash
TUI_LAYER_DEBUG=true TUI_DEBUG_RENDERING=true ./demo2_inspector.exe
# 点击 [I] Inspector 按钮
```

**预期输出**:
```
[PipelineRenderer] Using RenderLayers for multi-layer rendering
[RenderingPipeline] RenderLayers started
[LayerManager] Collecting layers...
[LayerManager] Found LayerInspector
[PaintEngine] Painting layer 4: LayerInspector
```

**检查点**:
- ✅ `Using RenderLayers` - 使用 Layer 渲染
- ✅ `Found LayerInspector` - 找到 Inspector 层

### 步骤 4: 验证快捷键

```bash
TUI_DEBUG=true TUI_INSPECTOR_VERBOSE=true ./demo2_inspector.exe
# 按 F12 或 Ctrl+D
```

**预期输出**:
```
[APP] KeyPress event received
[DEMO2] Inspector toggled: now visible=true
[Inspector] Toggled: visible
```

**检查点**:
- ✅ `KeyPress event received` - 键盘事件被捕获
- ✅ `Inspector toggled` - Inspector 状态切换

---

## 🐛 故障排除

### 问题: Inspector enabled=false

**症状**:
```
[DEMO2] Inspector enabled: false
```

**原因**: `Enable()` 未被调用

**解决**: 已在最新代码中修复 - 第 37 行无条件调用 `globalInspector.Enable()`

### 问题: Toggled but not visible

**症状**:
```
[DEMO2] After toggle, Inspector visible=true
```
但界面上看不到 Inspector

**可能原因**:
1. Layer 系统未检测到 Layer 标记
2. Inspector VNode 未添加到树中

**诊断**:
```bash
TUI_LAYER_DEBUG=true TUI_DEBUG_RENDERING=true ./demo2_inspector.exe
```

**检查**:
- `hasLayerNodes()` 是否返回 true
- `RenderLayers()` 是否被调用
- Inspector Overlay 是否在 VNode 树中

### 问题: 快捷键无反应

**症状**: 按 F12 或 Ctrl+D 无反应

**诊断**:
```bash
TUI_DEBUG=true ./demo2_inspector.exe 2>&1 | grep -i "key\|shortcut"
```

**检查**:
- `[APP] Inspector shortcuts registered` - 快捷键已注册
- `[APP] KeyPress event received` - 键盘事件被捕获

**可能原因**:
1. 终端不支持 F12 - 使用 Ctrl+D 代替
2. 快捷键未注册 - 检查 `SetupInspectorShortcut()` 是否被调用

### 问题: Inspector 显示但无数据

**症状**: Inspector 窗口显示但内容为空

**诊断**:
```bash
TUI_INSPECTOR_VERBOSE=true ./demo2_inspector.exe
```

**检查**:
- `AttachToApp()` 是否被调用
- `StartFrame()` / `EndFrame()` 是否被调用
- VNode 树是否正确传递

---

## 📚 调试环境变量速查

| 环境变量 | 作用 | 输出示例 |
|---------|------|---------|
| `TUI_DEBUG=true` | 基础调试 | `[APP]`, `[DEMO2]` |
| `TUI_DEBUG_UI=true` | UI 调试 | `[DEMO2]`, 快捷键信息 |
| `TUI_INSPECTOR_VERBOSE=true` | Inspector 详细 | `[Inspector]` 状态 |
| `TUI_LAYER_DEBUG=true` | Layer 系统 | Layer 收集和渲染 |
| `TUI_DEBUG_RENDERING=true` | 渲染流程 | Pipeline 状态 |
| `TUI_UI_DEBUG_LAYOUT=true` | 布局调试 | 完整组件树 |
| `TUI_UI_DEBUG_ALL=true` | 所有调试 | 所有以上内容 |

---

## 🎯 快速诊断命令

### 一键诊断脚本

```bash
#!/bin/bash
# demo2_diagnose.sh

echo "=== Demo2 Inspector 诊断 ==="
echo ""
echo "1. 基础状态："
TUI_DEBUG=true TUI_DEBUG_UI=true ./demo2_inspector.exe 2>&1 | grep -E "Inspector|shortcut" | head -5

echo ""
echo "2. Inspector 详细模式（请点击 [I] 按钮）："
TUI_INSPECTOR_VERBOSE=true ./demo2_inspector.exe 2>&1 | grep -E "\[Inspector\]|\[DEMO2\].*Inspector"

echo ""
echo "3. Layer 渲染（请点击 [I] 按钮）："
TUI_LAYER_DEBUG=true ./demo2_inspector.exe 2>&1 | grep -E "Layer|Render"
```

### Windows PowerShell 版本

```powershell
# demo2_diagnose.ps1

Write-Host "=== Demo2 Inspector 诊断 ===" -ForegroundColor Cyan
Write-Host ""

Write-Host "1. 基础状态：" -ForegroundColor Yellow
$env:TUI_DEBUG = "true"
$env:TUI_DEBUG_UI = "true"
.\demo2_inspector.exe 2>&1 | Select-String "Inspector|shortcut" | Select-Object -First 5

Write-Host ""
Write-Host "2. Inspector 详细模式（请点击 [I] 按钮）：" -ForegroundColor Yellow
$env:TUI_INSPECTOR_VERBOSE = "true"
.\demo2_inspector.exe 2>&1 | Select-String "\[Inspector\]|\[DEMO2\].*Inspector"

Write-Host ""
Write-Host "3. Layer 渲染（请点击 [I] 按钮）：" -ForegroundColor Yellow
$env:TUI_LAYER_DEBUG = "true"
.\demo2_inspector.exe 2>&1 | Select-String "Layer|Render"
```

---

## ✅ 成功运行的标准

运行 demo2 时，应该看到：

1. **启动信息**:
```
UI Inspector ready - Press [I] button or F12/Ctrl+D to toggle
Starting Mint TUI Demo - Press F12 or Ctrl+D to toggle Inspector
```

2. **调试信息**（如果启用 TUI_DEBUG）:
```
[DEMO2] Inspector enabled: true
[DEMO2] Inspector visible: false
[APP] Inspector shortcuts registered: F12, Ctrl+D
```

3. **点击 [I] 按钮后**（如果启用 TUI_INSPECTOR_VERBOSE）:
```
[DEMO2] [I] button clicked, Inspector enabled=true, visible=false
[Inspector] Toggled: visible
[DEMO2] After toggle, Inspector visible=true
```

4. **Layer 渲染信息**（如果启用 TUI_LAYER_DEBUG）:
```
[PipelineRenderer] Using RenderLayers for multi-layer rendering
[RenderingPipeline] RenderLayers started
[PaintEngine] Painting layer 4: LayerInspector
```

---

## 📖 相关文档

- `docs/debug/environment_variables.md` - 完整环境变量参考
- `docs/debug/quick_start.md` - 快速开始
- `LAYER_SYSTEM_ARCHITECTURE.md` - Layer 系统架构说明
- `INSPECTOR_OVERLAY_IMPLEMENTATION_SUMMARY.md` - Inspector 实施总结

---

**创建日期**: 2025-02-08
**状态**: ✅ 已验证
**版本**: 1.0
