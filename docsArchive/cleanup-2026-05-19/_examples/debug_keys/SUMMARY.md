# 🔑 Debug Key Inspector - 完整文档

## 📦 包含的文件

```
examples/debug_keys/
├── inspector.go              # 核心工具库（命令行版本）
├── README.md                 # 详细使用文档
├── QUICKSTART.md             # 快速入门指南
├── example_integration.go    # 简单集成示例
├── ui_inspector.go          # ★ UI Inspector（交互式版本）
└── SUMMARY.md               # 本文件 - 完整总结
```

## 🎯 两个版本

### 1. 命令行版本 (`inspector.go`)

**适用场景**：
- 在开发过程中快速检查 KEY 信息
- 自动化测试中的验证
- CI/CD 流程中的检查

**特点**：
- 简单易用，一行代码调用
- 彩色输出，易于阅读
- 完整的统计和一致性检查
- 支持配置选项

**快速使用**：

```bash
# 查看示例
cd examples/debug_keys
go run inspector.go

# 在你的代码中
inspector := debug_keys.NewDebugKeyInspector()
inspector.InspectVNodes(root)
```

### 2. UI 版本 (`ui_inspector.go`) ★推荐

**适用场景**：
- 交互式调试
- 实时观察 KEY 变化
- 直观的层叠视图

**特点**：
- 在 UI 中实时显示 KEY 信息
- 可以切换显示模式（Keys、Paths、Layers）
- 树形结构展示，易于理解
- 点击按钮即可触发更新

**运行方法**：

```bash
cd examples/debug_keys
go run ui_inspector.go
```

**界面说明**：
- **切换 Inspector**: 打开/关闭 Inspector overlay
- **显示 Keys**: 显示每个节点的 Key 值
- **显示 Paths**: 显示每个节点的 Path 值
- **显示 Layer 标记**: 显示 [MODAL], [OVERLAY] 等标记
- **打开 Modal/Overlay**: 创建不同的 layer，观察 KEY 变化

## 🚀 快速开始

### 最快的方式（30 秒开始）

```bash
cd examples/debug_keys
go run ui_inspector.go
```

点击 "切换 Inspector"，你会看到：
```
🔑 KEY INSPECTOR
显示 Keys: true
显示 Paths: false
显示 Layers: true
─────────────────────────────
├─ vstack
│ ├─ text
│ ├─ button 🔑OPEN_MODAL
│ ├─ button 🔑SHOW_OVERLAY
│ ├─ button 🔑TOGGLE_INSPECTOR
├─ div [MODAL] 🔑/root/modal[0]
│ ├─ vstack
│   ├─ text
│   ├─ button 🔑MODAL_BTN
```

## 💡 使用场景

### 场景 1: 调试 Modal 按钮点击失败

```go
// 运行 UI Inspector
go run ui_inspector.go

// 点击 "打开 Modal"
// 点击 "切换 Inspector"
// 观察 Modal 中按钮的 KEY 是否正确
```

**预期结果**：
- Modal 根节点：`[MODAL] 🔑/root/modal[0]`
- Modal 中按钮：`🔑/root/modal[0]/vstack[0]/button[0]`

### 场景 2: 验证 Layer 路径生成

```bash
# 启动应用
go run ui_inspector.go

# 打开不同的 layers
# 1. 打开 Modal → 观察路径是否为 /root/modal[0]
# 2. 打开 Overlay → 观察路径是否为 /root/overlay[0]
# 3. 同时打开 → 观察两者是否有正确的独立路径
```

### 场景 3: 检查 VNode 和 Fiber 的一致性

```go
// 使用命令行 Inspector
inspector := debug_keys.NewDebugKeyInspector()

// 比较 VNode 和 Fiber 树
inspector.CompareTrees(vnode, fiber)

// 输出：
// ✅ All keys are consistent!  // 正常
// ❌ Found 3 inconsistencies   // 有问题
```

## 📊 输出解读

### 彩色含义

| 颜色 | 含义 |
|------|------|
| 🟡 黄色 | Key 值 |
| 🔵 蓝色 | Path 值 |
| 🟣 紫色 | PathSegment 值 |
| 🔴 红色 | Key ≠ Path（不一致） |
| ⚠️ 灰色 | 无 Key 或无 Path（警告） |

### Layer 标记

| 标记 | 含义 |
|------|------|
| `[MODAL]` | Modal 层节点 |
| `[OVERLAY]` | Overlay 层节点 |
| `[TOOLTIP]` | Tooltip 层节点 |
| `[INSPECTOR]` | Inspector 层节点 |
| 无标记 | Base 层（默认） |

### 路径格式

```
/root/{layer}[{index}]/{type}[{index}]...
```

示例：
- `/root/base[0]` - Base 根节点
- `/root/modal[0]` - Modal 根节点
- `/root/modal[0]/vstack[0]/button[0]` - Modal 中的按钮
- `/root/modal[0]/button[0]/key[btn-id]` - 带 user key 的按钮

## 🔧 配置选项

### 命令行版本

```go
inspector := debug_keys.NewDebugKeyInspector()

inspector.MaxDepth = 20        // 最大遍历深度
inspector.ShowKeys = true      // 显示 Key 信息
inspector.ShowPaths = true     // 显示 Path 信息
inspector.ShowLayers = true    // 显示 Layer 标记
```

### UI 版本

在界面中点击 checkbox 实时切换：
- `[x] 显示 Keys`
- `[x] 显示 Paths`
- `[x] 显示 Layer 标记`

## 🎓 最佳实践

### 1. 开发时使用 UI Inspector

```bash
# 实时观察 KEY 变化
go run ui_inspector.go
```

### 2. 测试时使用命令行 Inspector

```go
if testing.Short() {
    inspector := debug_keys.NewDebugKeyInspector()
    inspector.InspectVNodes(root)
    inspector.PrintStatistics(root, nil)
}
```

### 3. CI/CD 中验证一致性

```go
// 确保所有 KEY 都正确同步
inspector := debug_keys.NewDebugKeyInspector()
mismatches := inspector.CompareTrees(vnode, fiber)
if mismatches > 0 {
    t.Fatalf("Found %d key mismatches", mismatches)
}
```

## 🐛 问题排查

### 问题 1: Modal 中的按钮无法点击

**症状**: 点击无反应

**检查步骤**：
1. 打开 UI Inspector
2. 找到按钮节点
3. 确认 KEY 不为空且格式正确

**预期**: `🔑/root/modal[0]/button[0]`

**修复**: 确保 `cloneExistingFiber` 和 `createChildFiberWithIndex` 正确工作

### 问题 2: 看到大量 ⚠️无Key

**是否正常**: 取决于节点类型

- ✅ 正常：text 元素、容器等
- ❌ 不正常：应该有 Key 的组件（如按钮）

### 问题 3: Layer 节点路径错误

**症状**: Modal 的路径是 `/root/base[0]/modal[0]` 而不是 `/root/modal[0]`

**检查**: `createChildFiberWithIndex` 中的 `isLayerNode` 检查

**修复**: 添加 `isLayerNode := vnode.GetLayer() != rtui.LayerBase`

## 📚 相关文档

- [Layer System Architecture](../../docs/layout/LAYER_SYSTEM_ARCHITECTURE.md)
- [Multi-Layer Key Fix](../../docs/render/RENDERLAYERS_VNODE_KEY_FIX.md)
- [Event Routing](../../docs/issue/event_refactor/BUTTON_EVENT_ROUTING_FIX_SUMMARY.md)

## 🎁 额外功能

### 1. 统计信息

```go
inspector.PrintStatistics(vnode, fiber)

// 输出：
// VNode Tree:
//   Total nodes: 7
//   With key: 3 (42.9%)
//   Layer nodes: 1
```

### 2. 一致性检查

```go
inspector.CompareTrees(vnode, fiber)

// 输出：
// VNode unique keys: 3
// Fiber unique keys: 7
// ✅ All keys are consistent!
```

### 3. 树形结构显示

```
├─ fragment
│ ├─ vstack
│ │ ├─ text ⚠️无Key
│ │ └─ button 🔑btn-open
```

## 🔐 安全提示

1. **生产环境禁用**: 只在开发时使用
2. **限制输出深度**: 避免过大的输出
3. **缓存结果**: 频繁调用可能影响性能
4. **保护敏感信息**: KEY 可能包含敏感信息

## 🚀 性能优化

```go
// 性能优化提示
inspector.MaxDepth = 10  // 限制深度
inspector.ShowKeys = false  // 禁用不需要的输出
```

## 📞 获取帮助

- 📖 查看 [README.md](README.md) 了解详细用法
- 🚀 查看 [QUICKSTART.md](QUICKSTART.md) 快速上手
- 🐛 提交 Issue 报告问题

---

**推荐使用流程**：
1. 先阅读这个 SUMMARY.md 了解概况
2. 参考 QUICKSTART.md 快速开始
3. 需要细节时查看 README.md
4. 遇到问题时查看问题排查章节
