# Debug Key Inspector - 快速入门指南

## 📦 文件说明

```
examples/debug_keys/
├── inspector.go            # 核心工具库
├── README.md              # 详细使用文档
├── example_integration.go  # 集成示例
└── quick_start.go         # 本文件 - 快速入门
```

## 🚀 快速开始

### 步骤 1: 运行独立工具（查看示例）

```bash
cd examples/debug_keys
go run inspector.go
```

### 步骤 2: 在你的应用中集成

在你的 `main.go` 中：

```go
package main

import (
    "github.com/wwsheng009/mint/examples/debug_keys"
    rtui "github.com/wwsheng009/mint/runtime/ui"
    "github.com/wwsheng009/mint/ui"
)

func main() {
    app := ui.NewApp()

    // 你的 root 组件
    root := func() rtui.VNode {
        return rtui.NewElement("vstack").
            SetChildren(
                rtui.NewElement("text").
                    SetProp("content", "Hello, Mint!"),
            )
    }

    // 渲染
    app.Render(root)

    // 🔑 添加 KEY 检查
    inspector := debug_keys.NewDebugKeyInspector()
    inspector.InspectVNodes(app.Root())
}
```

### 步骤 3: 查看 KEY 信息

运行你的程序，你会看到：

```
══════════════════════════════════════════════════════════════════════════════
🔑 VNODE TREE - KEY INFORMATION
══════════════════════════════════════════════════════════════════════════════
│─ fragment
  │─ vstack
    │─ text ⚠️无Key

✅ Total VNodes: 2
```

## 🎯 常用场景

### 场景 1: 调试 Modal 中的点击事件

```go
// 打开 Modal 后检查
modalOpen = true
app.MarkDirty()

// 检查 KEY 信息
inspector := debug_keys.NewDebugKeyInspector()
inspector.InspectVNodes(app.Root())

// 预期输出应该显示：
// │─ button Key:"/root/modal[0]/button[0]/key[btn-id]"
```

### 场景 2: 检查 Layer 节点

```go
inspector := debug_keys.NewDebugKeyInspector()
inspector.ShowLayers = true  // 显示 [MODAL], [OVERLAY] 等标记
inspector.InspectVNodes(app.Root())

// 预期输出：
// │─ div [MODAL] Key:"/root/modal[0]"
```

### 场景 3: 检查 Fiber 树

```go
// 需要访问 Fiber 树（在 reconciler 中）
reconciler := app.GetReconciler()
rootFiber := reconciler.Root Fiber()

inspector := debug_keys.NewDebugKeyInspector()
inspector.InspectFibers(rootFiber)

// 预期输出包含 Path 信息：
// │─ button Key:"..." Path:"/root/modal[0]/button[0]/key[btn-id]"
```

## 📊 输出解读

### 1. 颜色含义

- 🟡 **黄色**: Key 值
- 🔵 **蓝色**: Path 值
- 🟣 **紫色**: PathSegment 值
- 🔴 **红色**: Key ≠ Path（不一致）
- ⚠️ **灰色**: 无 Key 或无 Path（警告）

### 2. Layer 标记

- `[MODAL]`: Modal 层节点
- `[OVERLAY]`: Overlay 层节点
- `[TOOLTIP]`: Tooltip 层节点
- `[INSPECTOR]`: Inspector 层节点

### 3. 一致性检查

```go
inspector.CompareTrees(vnode, fiber)
```

输出：
```
✅ All keys are consistent!  // 正常
❌ Found X inconsistencies   // 有问题
```

## 🔧 配置选项

```go
inspector := debug_keys.NewDebugKeyInspector()

// 控制输出内容
inspector.MaxDepth = 20        // 最大遍历深度（默认 20）
inspector.ShowKeys = true      // 显示 Key（默认 true）
inspector.ShowPaths = true     // 显示 Path（默认 true）
inspector.ShowLayers = true    // 显示 Layer 标记（默认 true）
```

## 🎓 最佳实践

### 1. 开发模式集成

```go
// 在 app.go 中
type App struct {
    // ...
    debugMode bool
}

func (app *App) Render(root rtui.VNode) {
    // 正常渲染流程...

    // 开发模式下检查 KEY
    if app.debugMode {
        inspector := debug_keys.NewDebugKeyInspector()
        inspector.InspectVNodes(app.root)
    }
}
```

### 2. 事件处理后检查

```go
button.OnClick(func() {
    // 处理点击事件...

    // 检查 KEY 是否正确
    inspector := debug_keys.NewDebugKeyInspector()
    inspector.InspectVNodes(app.Root())
})
```

### 3. 条件检查

只在特定条件下启用：

```go
if os.Getenv("DEBUG_KEYS") == "true" {
    inspector := debug_keys.NewDebugKeyInspector()
    inspector.InspectVNodes(app.Root())
}
```

## 🐛 问题排查

### 问题 1: Modal 中的按钮无法点击

**检查步骤**:

1. 打开 Modal
2. 运行 Inspector
3. 查看按钮的 Key 是否为空或不正确

**可能原因**:
- VNode Key 与 Fiber Path 不一致
- `cloneExistingFiber` 条件未触发
- Layer 节点路径生成错误

**解决方案**:
参考 `docs/render/RENDERLAYERS_VNODE_KEY_FIX.md`

### 问题 2: 看到大量 ⚠️无Key

**原因**: 静态 UI 节点可能没有 Key

**是否正常**:
- ✅ 正常：静态文本、容器等不需要用户 Key
- ❌ 不正常：应该有 Key 的组件（如按钮）没有 Key

### 问题 3: Key ≠ Path 警告

**检查**:
```go
inspector.CompareTrees(vnode, fiber)
```

**原因**: VNode.SetKey() 未被调用或设置错误

**修复**: 确保 `cloneExistingFiber` 和 `createChildFiberWithIndex` 正确同步 Key

## 📚 更多资源

- [详细文档](README.md)
- [Multi-Layer Key Fix](../../docs/render/RENDERLAYERS_VNODE_KEY_FIX.md)
- [Layer System](../../docs/layout/LAYER_SYSTEM_ARCHITECTURE.md)

## 💡 提示

1. **只在开发时使用**: 生产环境应关闭 Inspector
2. **限制 MaxDepth**: 树太深时可以减少输出
3. **结合日志**: 配合 `TUI_DEBUG_HITMAP` 和 `TUI_DEBUG_RENDER` 使用
4. **保存输出**: 可以重定向到文件保存供分析

```bash
go run your_app.go 2>&1 | tee debug_output.txt
```

---

**需要帮助？** 查看文档或提交 Issue！
