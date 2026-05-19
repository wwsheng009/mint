# Debug Key Inspector

调试工具，用于显示渲染后所有层次的 KEY 信息。可以帮助你：
- 查看 VNode 树的所有 Key 分布
- 查看 Fiber 树的 Path 和 Key 分布
- 验证 Layer 节点（Modal, Overlay, Tooltip, Inspector）的正确性
- 检查 VNode 和 Fiber 的 Key 一致性
- 确认 HitMap NodeID 与 Instance Key 是否匹配

## 快速开始

### 1. 导入工具包

```go
import "github.com/wwsheng009/mint/examples/debug_keys"
```

### 2. 创建 Inspector 实例

```go
inspector := debug_keys.NewDebugKeyInspector()

// 可选配置
inspector.MaxDepth = 20        // 最大遍历深度
inspector.ShowKeys = true      // 显示 Key 信息
inspector.ShowPaths = true     // 显示 Path 信息
inspector.ShowLayers = true    // 显示 Layer 信息
```

### 3. 在渲染完成后调用

```go
// 获取渲染后的 VNode 和 Fiber
rootVNode := app.Root()
rootFiber := app.GetCurrentFiber()  // 或通过 reconciler 获取

// 检查 VNode 树
inspector.InspectVNodes(rootVNode)

// 检查 Fiber 树
inspector.InspectFibers(rootFiber)

// 比较 VNode 和 Fiber 的一致性（重要！）
inspector.CompareTrees(rootVNode, rootFiber)

// 显示统计信息
inspector.PrintStatistics(rootVNode, rootFiber)
```

## 完整示例

```go
package main

import (
    rtui "github.com/wwsheng009/mint/runtime/ui"
    "github.com/wwsheng009/mint/ui"
    "github.com/wwsheng009/mint/examples/debug_keys"
    "github.com/wwsheng009/mint/internal/reconciler"
)

func main() {
    app := ui.NewApp()

    // 创建一个包含 Modal 的 UI
    var modalOpen bool
    root := func() rtui.VNode {
        baseContent := rtui.NewElement("vstack").
            SetChildren(
                rtui.NewElement("text").
                    SetProp("content", "点击按钮打开 Modal"),
                rtui.NewElement("button").
                    SetProp("label", "打开 Modal").
                    OnClick(func() {
                        modalOpen = true
                        app.MarkDirty()
                    }),
            )

        var modalNode rtui.VNode
        if modalOpen {
            modalContent := rtui.NewElement("vstack").
                SetChildren(
                    rtui.NewElement("text").SetProp("content", "这是 Modal"),
                    rtui.NewElement("button").
                        SetProp("label", "Modal 中的按钮").
                        SetKey("modal-btn").
                        OnClick(func() {
                            fmt.Println("Button clicked!")
                        }),
                )
            modalNode = ui.Modal(modalContent)
        }

        children := []rtui.VNode{baseContent}
        if modalNode != nil {
            children = append(children, modalNode)
        }

        return rtui.NewElement("fragment").
            SetChildren(children...)
    }

    // 执行渲染
    app.Render(root)

    // 🔑 使用 Inspector 检查 KEY 信息
    inspector := debug_keys.NewDebugKeyInspector()

    // 获取渲染后的树
    vnodes := app.RootVNode()
    fibers := app.GetCurrentFiber()  // 需要实现这个方法

    // 显示 KEY 信息
    inspector.InspectVNodes(vnodes)
    inspector.InspectFibers(fibers)

    // 检查一致性
    inspector.CompareTrees(vnodes, fibers)

    // 显示统计
    inspector.PrintStatistics(vnodes, fibers)
}
```

## 输出示例

```
══════════════════════════════════════════════════════════════════════════════
🔑 VNODE TREE - KEY INFORMATION
══════════════════════════════════════════════════════════════════════════════
│─ fragment
  │─ vstack
    │─ text ⚠️无Key
    │─ button ⚠️无Key
  │─ div [MODAL] Key:"/root/modal[0]"
    │─ vstack
      │─ text ⚠️无Key
      │─ button Key:"/root/modal[0]/button[0]/key[modal-btn]"

✅ Total VNodes: 7

══════════════════════════════════════════════════════════════════════════════
🌳 FIBER TREE - PATH & KEY INFORMATION
══════════════════════════════════════════════════════════════════════════════
│─ fragment Key:"/root/base[0]" Path:"/root/base[0]"
  │─ vstack Key:"/root/base[0]/vstack[0]" Path:"/root/base[0]/vstack[0]"
    │─ text Key:"/root/base[0]/vstack[0]/text[0]" Path:"/root/base[0]/vstack[0]/text[0]"
    │─ button Key:"/root/base[0]/vstack[0]/button[0]" Path:"/root/base[0]/vstack[0]/button[0]"
  │─ div [MODAL] Key:"/root/modal[0]" Path:"/root/modal[0]"
    │─ vstack Key:"/root/modal[0]/vstack[0]" Path:"/root/modal[0]/vstack[0]"
      │─ text Key:"/root/modal[0]/vstack[0]/text[0]" Path:"/root/modal[0]/vstack[0]/text[0]"
      │─ button Key:"/root/modal[0]/vstack[0]/button[0]/key[modal-btn]" Path:"/root/modal[0]/vstack[0]/button[0]/key[modal-btn]"

✅ Total Fibers: 7

══════════════════════════════════════════════════════════════════════════════
🔍 VNode vs Fiber - KEY CONSISTENCY CHECK
══════════════════════════════════════════════════════════════════════════════

📊 Key Distribution:
  VNode unique keys: 3
  Fiber unique keys: 7

⚠️  Inconsistencies:
  ✅ All keys are consistent!

══════════════════════════════════════════════════════════════════════════════
📈 STATISTICS
══════════════════════════════════════════════════════════════════════════════

📊 VNode Tree:
  Total nodes: 7
  With key: 3 (42.9%)
  Without key: 4 (57.1%)
  Layer nodes: 1

🌳 Fiber Tree:
  Total nodes: 7
  With key: 7 (100.0%)
  Without key: 0 (0.0%)
  With path: 7 (100.0%)
  Layer nodes: 1
```

## 输出说明

- 🟡 **黄色文字**: Key 值
- 🔵 **蓝色文字**: Path 值
- 🟣 **紫色文字**: PathSegment 值
- 🔴 **红色警告**: Key ≠ Path (不一致)
- ⚠️ **灰色警告**: 无 Key 或无 Path
- `[MODAL]`, `[OVERLAY]` 等: Layer 标记

## 常见问题排查

### 问题 1: 模态框中的按钮无法点击

**症状**: HitMap 找不到 Instance

**检查步骤**:

1. 运行 Inspector
2. 查找按钮的 Key 是否同步

```
VNode: Key:"/root/modal[0]/button[0]/key[modal-btn]"
Fiber: Path:"/root/modal[0]/button[0]/key[modal-btn]"
```

3. 如果不一致，检查 `cloneExistingFiber` 和 `createChildFiberWithIndex` 的逻辑

### 问题 2: Layer 节点没有正确的路径

**症状**: Modal 的路径不是 `/root/modal[0]`，而是基于父节点

**检查步骤**:

1. 查找 Modal 节点的路径
2. 确认是否是 `/root/modal[0]` 格式
3. 如果不是，检查 `createChildFiberWithIndex` 中的 `isLayerNode` 检查

### 问题 3: Instance Key 与 HitMap NodeID 不匹配

**症状**: Instance key 存在但 HitMap 找不到

**检查步骤**:

1. 使用 `CompareTrees` 检查一致性
2. 查看 VNode 和 Fiber 的 Key 是否同步
3. 如果 Fiber 的 Path 不为空但 VNode 的 Key 为空，修复 `cloneExistingFiber`

## API 参考

### DebugKeyInspector

#### 方法

- `InspectVNodes(vnode rtui.VNode)` - 显示 VNode 树
- `InspectFibers(fiber *reconciler.Fiber)` - 显示 Fiber 树
- `CompareTrees(vnode rtui.VNode, fiber *reconciler.Fiber)` - 比较一致性
- `PrintStatistics(vnode rtui.VNode, fiber *reconciler.Fiber)` - 显示统计

#### 配置

- `MaxDepth int` - 最大遍历深度
- `ShowKeys bool` - 显示 Key 信息
- `ShowPaths bool` - 显示 Path 信息
- `ShowLayers bool` - 显示 Layer 信息

## 运行示例

```bash
cd examples/debug_keys
go run inspector.go
```

## 集成到你的应用

你可以在开发模式下启用 Inspector：

```go
// 在 app.go 或 main.go 中
func initDebugMode() {
    if os.Getenv("DEBUG_KEYS") == "true" {
        // 注册调试命令
        // 每次渲染后自动调用 Inspector
    }
}

// 或者在需要的地方手动调用
if debugMode {
    inspector := debug_keys.NewDebugKeyInspector()
    inspector.InspectVNodes(root)
    inspector.InspectFibers(fiber)
}
```

## 相关文档

- [RenderLayers VNode Key Fix](../../docs/render/RENDERLAYERS_VNODE_KEY_FIX.md)
- [Layer System Architecture](../../docs/layout/LAYER_SYSTEM_ARCHITECTURE.md)
- [Event Routing](../../docs/issue/event_refactor/BUTTON_EVENT_ROUTING_FIX_SUMMARY.md)
