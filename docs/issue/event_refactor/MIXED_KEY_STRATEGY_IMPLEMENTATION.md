# 混合Key策略实现计划

> **目标**: 实现静态UI自动生成路径Key + 动态列表强制要求用户Key的混合策略
> **状态**: 🟡 实施中
> **创建日期**: 2026-02-12

---

## 1. 设计总结

### 1.1 核心策略

```
┌─────────────────────────────────────────────────────────────┐
│                    Key生成策略                              │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  优先级1: 用户显式Key                                         │
│    → 直接使用用户的Key                                       │
│    → 示例: "btn-submit", "username"                         │
│                                                              │
│  优先级2: 动态列表检测                                        │
│    → 如果父节点是列表类型 → panic强制要求Key                  │
│    → 示例: List, GridView, VirtualList                      │
│                                                              │
│  优先级3: 静态UI组件                                          │
│    → 自动生成基于路径的Key                                   │
│    → 示例: /root/base[0]/vstack[0]/panel[0]                 │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### 1.2 关键原则

✅ **用户Key优先** - 兼容现有代码，不破坏已有功能
✅ **动态列表安全** - 强制要求Key，避免状态丢失
✅ **静态UI便利** - 自动生成Key，用户无感知
✅ **清晰的错误** - panic时提供详细的修复建议

---

## 2. 实现步骤

### Phase 1: 基础设施 (必需)

#### 1.1 创建路径生成器

**文件**: `internal/reconciler/path_generator.go`

```go
package reconciler

import (
    "fmt"
    rtui "github.com/wwsheng009/mint/runtime/ui"
)

// PathGenerator 生成组件的路径Key
type PathGenerator struct {
    // 缓存路径段，避免重复计算
    segmentCache map[string]int
}

// NewPathGenerator 创建路径生成器
func NewPathGenerator() *PathGenerator {
    return &PathGenerator{
        segmentCache: make(map[string]int),
    }
}

// GeneratePath 生成组件的Key
// 返回: 完整的路径字符串，如 /root/base[0]/vstack[0]/panel[0]
func (pg *PathGenerator) GeneratePath(
    parent *Fiber,
    vnode rtui.VNode,
    siblingIndex int,
) string {
    // 1. 检查是否是根节点
    if parent == nil {
        return pg.generateRootPath(vnode)
    }

    // 2. 获取组件类型标识
    typeID := pg.getTypeIdentifier(vnode)

    // 3. 获取该类型的索引
    index := pg.getTypeIndex(parent, typeID, siblingIndex)

    // 4. 生成路径段
    segment := fmt.Sprintf("%s[%d]", typeID, index)

    // 5. 拼接完整路径
    return parent.Path + "/" + segment
}

// generateRootPath 生成根节点路径
func (pg *PathGenerator) generateRootPath(vnode rtui.VNode) string {
    layer := vnode.GetLayer()
    layerName := getLayerName(layer)
    return fmt.Sprintf("/root/%s[0]", layerName)
}

// getTypeIdentifier 获取组件的类型标识
func (pg *PathGenerator) getTypeIdentifier(vnode rtui.VNode) string {
    switch v := vnode.(type) {
    case *rtui.ComponentVNode:
        return v.Name()
    case *rtui.ElementVNode:
        return v.Tag()
    case *rtui.TextVNode:
        return "text"
    case *rtui.FragmentVNode:
        return "fragment"
    default:
        return "unknown"
    }
}

// getTypeIndex 获取组件类型在父节点中的索引
func (pg *PathGenerator) getTypeIndex(
    parent *Fiber,
    typeID string,
    siblingIndex int,
) int {
    if parent == nil {
        return 0
    }

    // 统计在当前索引之前有多少同类型的兄弟节点
    count := 0
    child := parent.Child

    for i := 0; i < siblingIndex && child != nil; i++ {
        childTypeID := pg.getTypeIdentifier(child.VNode)
        if childTypeID == typeID {
            count++
        }
        child = child.Sibling
    }

    return count
}

// getLayerName 获取Layer的名称
func getLayerName(layer rtui.Layer) string {
    switch layer {
    case rtui.LayerBase:
        return "base"
    case rtui.LayerOverlay:
        return "overlay"
    case rtui.LayerModal:
        return "modal"
    case rtui.LayerTooltip:
        return "tooltip"
    case rtui.LayerInspector:
        return "inspector"
    default:
        return "unknown"
    }
}
```

#### 1.2 创建列表检测器

**文件**: `internal/reconciler/list_detector.go`

```go
package reconciler

import (
    "fmt"
    "os"
    rtui "github.com/wwsheng009/mint/runtime/ui"
)

// 动态列表类型的标签
var dynamicListTags = map[string]bool{
    "List":       true,
    "GridView":   true,
    "VirtualList": true,
    "ForEach":    true,
}

// isDynamicList 检查父节点是否是动态列表
func isDynamicList(parent *Fiber) bool {
    if parent == nil {
        return false
    }

    // 检查1: 标签匹配
    if dynamicListTags[parent.Tag] {
        return true
    }

    // 检查2: VNode类型检查
    if vnode, ok := parent.VNode.(interface{ IsDynamicList() bool }); ok {
        return vnode.IsDynamicList()
    }

    // 检查3: Props标记
    if parent.VNode.Props() != nil {
        if isDynamic, ok := parent.VNode.Props()["_dynamicList"].(bool); ok {
            return isDynamic
        }
    }

    return false
}

// requireKeyPanic 当动态列表缺少Key时panic
func requireKeyPanic(parent *Fiber, vnode rtui.VNode, siblingIndex int) {
    panicMsg := fmt.Sprintf(
        "\n"+
        "╔═══════════════════════════════════════════════════════════════╗\n"+
        "║  Dynamic List Requires Key Error                             ║\n"+
        "╚═══════════════════════════════════════════════════════════════╝\n"+
        "\n"+
        "The component at index %d is missing a required key.\n"+
        "\n"+
        "Parent Component: %s\n"+
        "Child Component: %s\n"+
        "\n"+
        "Dynamic lists require each child to have a stable key for:\n"+
        "  • Preserving component state across renders\n"+
        "  • Correct event routing\n"+
        "  • Proper reconciliation\n"+
        "\n"+
        "How to Fix:\n"+
        "───────────\n"+
        "  ❌ Wrong:\n"+
        "     List().Children(\n"+
        "       Item(item1).Build(),\n"+
        "       Item(item2).Build(),\n"+
        "     )\n"+
        "\n"+
        "  ✅ Correct:\n"+
        "     List().Children(\n"+
        "       Item(item1).Key(item1.ID).Build(),\n"+
        "       Item(item2).Key(item2.ID).Build(),\n"+
        "     )\n"+
        "\n"+
        "Recommended: Use a unique identifier from your data:\n"+
        "  • item.ID\n"+
        "  • item.UUID\n"+
        "  • item.Slug\n"+
        "\n"+
        "See: https://react.dev/learn/rendering-lists#why-does-react-need-keys\n",
        siblingIndex,
        parent.Tag,
        getTypeIdentifier(vnode),
    )

    if os.Getenv("TUI_DEBUG_KEY") == "true" {
        // 调试模式：打印调用栈
        panic(panicMsg)
    } else {
        // 生产模式：简洁的错误信息
        panic(fmt.Sprintf(
            "Dynamic list '%s' requires key for child at index %d. "+
            "Use .Key(item.ID) to provide a stable identifier.",
            parent.Tag, siblingIndex,
        ))
    }
}
```

#### 1.3 扩展Fiber结构

**文件**: `internal/reconciler/fiber_extensions.go`

```go
package reconciler

import (
    "strings"
    rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Fiber扩展字段（添加到现有Fiber结构）
// 注意: 这些字段应该添加到 runtime/ui/fiber.go 的 Fiber 结构中

/*
type Fiber struct {
    // ... 现有字段 ...

    // ✨ 新增: 路径相关字段
    Path         string  // 完整路径: /root/base[0]/panel[0]
    PathSegment  string  // 当前路径段: panel[0]
    SiblingIndex int     // 在兄弟节点中的索引
}
*/

// extractPathSegment 从完整路径中提取最后一段
func extractPathSegment(path string) string {
    if path == "" {
        return ""
    }

    parts := strings.Split(path, "/")
    if len(parts) == 0 {
        return ""
    }

    return parts[len(parts)-1]
}

// getTypeIDFromSegment 从路径段中提取类型ID
// 例如: "button[2]" -> "button"
func getTypeIDFromSegment(segment string) string {
    idx := strings.Index(segment, "[")
    if idx == -1 {
        return segment
    }
    return segment[:idx]
}
```

### Phase 2: 核心逻辑 (必需)

#### 2.1 修改createChildFiber

**文件**: `internal/reconciler/diff.go`

```go
// createChildFiber 创建子Fiber（混合策略版本）
func createChildFiber(
    returnFiber *Fiber,
    vnode rtui.VNode,
    lanes Lane,
    siblingIndex int,
) *Fiber {
    fiber := CreateFiberFromVNode(vnode)
    fiber.Return = returnFiber
    fiber.Lanes = lanes
    fiber.Props = vnode.Props()
    fiber.SiblingIndex = siblingIndex

    // ✨ 混合Key策略
    userKey := vnode.Key()

    if userKey != "" {
        // 优先级1: 用户设置了Key → 直接使用
        fiber.Key = userKey
        fiber.Path = returnFiber.Path + "/" + userKey
    } else if isDynamicList(returnFiber) {
        // 优先级2: 动态列表 → 强制要求Key
        requireKeyPanic(returnFiber, vnode, siblingIndex)
    } else {
        // 优先级3: 静态UI → 自动生成路径Key
        fiber.Path = pathGenerator.GeneratePath(returnFiber, vnode, siblingIndex)
        fiber.Key = fiber.Path
        vnode.SetKey(fiber.Path)
    }

    fiber.PathSegment = extractPathSegment(fiber.Path)

    return fiber
}
```

#### 2.2 修改cloneExistingFiber

**文件**: `internal/reconciler/diff.go`

```go
// cloneExistingFiber 克隆现有Fiber（保持路径）
func cloneExistingFiber(returnFiber *Fiber, current *Fiber, vnode rtui.VNode) *Fiber {
    fiber := CloneFiber(current)
    fiber.Return = returnFiber
    fiber.VNode = vnode
    fiber.Props = vnode.Props()
    fiber.Lanes = LaneNoLane
    fiber.Flags = EffectNoEffect

    // ✨ 保持路径不变（复用Instance的关键）
    userKey := vnode.Key()
    if userKey != "" && userKey != current.Key {
        // 用户修改了Key，重新生成路径
        fiber.Key = userKey
        fiber.Path = returnFiber.Path + "/" + userKey
    } else {
        // 保持原来的路径和Key
        fiber.Path = current.Path
        fiber.Key = current.Key
    }
    fiber.PathSegment = current.PathSegment

    // Link to alternate
    fiber.Alternate = current
    if current.Alternate != nil {
        current.Alternate.Alternate = nil
    }

    return fiber
}
```

#### 2.3 初始化pathGenerator

**文件**: `internal/reconciler/reconciler.go`

```go
// Reconciler 结构扩展
type Reconciler struct {
    // ... 现有字段 ...

    // ✨ 新增: 路径生成器
    pathGenerator *PathGenerator
}

// NewReconciler 创建reconciler（修改版）
func NewReconciler(app *framework.App, rootComponent rtui.ComponentFunc, config ReconcilerConfig) *Reconciler {
    timeBudget := config.TimeBudget
    if timeBudget == 0 {
        timeBudget = 5 * time.Millisecond
    }

    return &Reconciler{
        app:                 app,
        rootComponent:       rootComponent,
        instanceMgr:         state.NewInstanceManager(),
        interactionStateMgr: state.NewInteractionStateManager(),
        keyValidator:        state.NewKeyValidator(),
        timeBudget:          timeBudget,
        ctx:                 rtui.NewComponentContextForRoot(),
        enableFiber:         config.EnableFiber,
        vnodeConverter:      NewVNodeConverter(),
        pathGenerator:       NewPathGenerator(), // ✨ 初始化
    }
}
```

### Phase 3: 调试和测试 (推荐)

#### 3.1 添加调试日志

**文件**: `internal/reconciler/diff.go`

```go
// createChildFiber 创建子Fiber（添加调试日志）
func createChildFiber(
    returnFiber *Fiber,
    vnode rtui.VNode,
    lanes Lane,
    siblingIndex int,
) *Fiber {
    // ... 现有代码 ...

    // ✨ 调试日志
    if os.Getenv("TUI_DEBUG_KEY") == "true" {
        log.UILogger.Debug(
            "[createChildFiber] parent=%s, siblingIdx=%d, userKey=%q, finalKey=%q, path=%q",
            returnFiber.Tag,
            siblingIndex,
            userKey,
            fiber.Key,
            fiber.Path,
        )
    }

    return fiber
}
```

#### 3.2 创建测试文件

**文件**: `internal/reconciler/mixed_key_strategy_test.go`

```go
package reconciler

import (
    "testing"
    rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestStaticUI_AutoKey 测试静态UI自动生成Key
func TestStaticUI_AutoKey(t *testing.T) {
    // 创建静态VNode（无Key）
    vnode := rtui.NewElement("vstack").
        AddChild(rtui.NewElement("panel")).
        AddChild(rtui.NewElement("panel")).
        AddChild(rtui.NewElement("button"))

    // 创建Fiber
    root := &Fiber{Path: "/root/base[0]"}
    fiber := createChildFiber(root, vnode, LaneSyncLane, 0)

    // 验证自动生成的Key
    expectedKey := "/root/base[0]/vstack[0]"
    if fiber.Key != expectedKey {
        t.Errorf("Expected key %q, got %q", expectedKey, fiber.Key)
    }
}

// TestDynamicList_RequireKey 测试动态列表强制要求Key
func TestDynamicList_RequireKey(t *testing.T) {
    // 创建列表容器
    listFiber := &Fiber{
        Tag:  "List",
        Path: "/root/base[0]/list[0]",
    }

    // 创建子VNode（无Key）
    itemVNode := rtui.NewElement("item")

    // 应该panic
    defer func() {
        if r := recover(); r == nil {
            t.Errorf("Expected panic for missing key in dynamic list")
        }
    }()

    createChildFiber(listFiber, itemVNode, LaneSyncLane, 0)
}

// TestUserKey_Priority 测试用户Key优先
func TestUserKey_Priority(t *testing.T) {
    // 创建带用户Key的VNode
    vnode := rtui.NewElement("button")
    vnode.SetKey("my-button")

    root := &Fiber{Path: "/root/base[0]/panel[0]"}
    fiber := createChildFiber(root, vnode, LaneSyncLane, 0)

    // 验证使用用户的Key
    expectedKey := "my-button"
    if fiber.Key != expectedKey {
        t.Errorf("Expected key %q, got %q", expectedKey, fiber.Key)
    }

    // 验证路径包含用户的Key
    expectedPath := "/root/base[0]/panel[0]/my-button"
    if fiber.Path != expectedPath {
        t.Errorf("Expected path %q, got %q", expectedPath, fiber.Path)
    }
}
```

---

## 3. 兼容性保证

### 3.1 现有代码不受影响

```go
// ✅ 现有代码（已有Key）完全兼容
ButtonBuilder("Submit").Key("btn-submit").Build()
// → Key: "btn-submit"（使用用户的Key）
```

### 3.2 新代码享受便利

```go
// ✅ 新代码（静态UI）不需要Key
VStack(
  Panel(),
  Panel(),
  Button(),
)
// → 自动生成: /root/base[0]/vstack[0]/panel[0]
//              /root/base[0]/vstack[0]/panel[1]
//              /root/base[0]/vstack[0]/button[0]
```

### 3.3 动态列表强制安全

```go
// ✅ 动态列表必须设置Key
List().Children(
  Item(item1).Key(item1.ID).Build(),
  Item(item2).Key(item2.ID).Build(),
)
```

---

## 4. 迁移指南

### 4.1 静态UI（无需修改）

```go
// ✅ 不需要修改
func renderStaticUI() ui.VNode {
  return VStack(
    Header(),
    Content(),
    Footer(),
  )
}
```

### 4.2 动态列表（必须修改）

```go
// ❌ 修改前（会panic）
func renderList(items []Item) ui.VNode {
  children := []ui.VNode{}
  for _, item := range items {
    children = append(children, Item(item).Build())
  }
  return List().Children(children...)
}

// ✅ 修改后（正确）
func renderList(items []Item) ui.VNode {
  children := []ui.VNode{}
  for _, item := range items {
    children = append(children,
      Item(item).Key(item.ID).Build(),  // 添加Key
    )
  }
  return List().Children(children...)
}
```

### 4.3 混合场景

```go
// ✅ 混合场景：部分使用用户Key
func renderMixed() ui.VNode {
  return VStack(
    Panel().Key("header-panel").Build(  // 用户Key
      Header(),
    ),
    Panel().Build(                         // 自动Key
      List().Children(
        Item(item1).Key(item1.ID).Build(),  // 列表项必须有Key
        Item(item2).Key(item2.ID).Build(),
      ),
    ),
  )
}
```

---

## 5. 测试计划

### 5.1 单元测试

- [ ] TestStaticUI_AutoKey
- [ ] TestDynamicList_RequireKey
- [ ] TestUserKey_Priority
- [ ] TestPathGeneration_Nesting
- [ ] TestPathGeneration_Layers
- [ ] TestListOperations_Insert
- [ ] TestListOperations_Delete
- [ ] TestListOperations_Reorder

### 5.2 集成测试

- [ ] 完整应用渲染测试
- [ ] 事件路由测试
- [ ] Instance复用测试
- [ ] 状态保持测试

### 5.3 性能测试

- [ ] 路径生成性能
- [ ] Key查找性能
- [ ] 内存占用测试

---

## 6. 文档更新

### 6.1 需要更新的文档

- [x] PATH_BASED_KEY_DESIGN.md（已创建）
- [x] PATH_BASED_KEY_FEASIBILITY_ANALYSIS.md（已创建）
- [x] REACT_LIST_HANDLING.md（已创建）
- [x] MIXED_KEY_STRATEGY_IMPLEMENTATION.md（本文档）
- [ ] README.md（添加Key使用指南）
- [ ] MIGRATION_GUIDE.md（迁移指南）
- [ ] API_REFERENCE.md（API文档）

---

## 7. 实施时间表

| 阶段 | 任务 | 预计时间 | 状态 |
|------|------|---------|------|
| Phase 1 | 基础设施（路径生成器、列表检测器） | 2小时 | 🟡 待开始 |
| Phase 2 | 核心逻辑（修改diff.go） | 2小时 | ⚪ 未开始 |
| Phase 3 | 测试和调试 | 2小时 | ⚪ 未开始 |
| Phase 4 | 文档和示例 | 1小时 | ⚪ 未开始 |
| **总计** | | **7小时** | |

---

## 8. 风险和缓解

### 8.1 风险

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|---------|
| 破坏现有代码 | 高 | 中 | 用户Key优先，完全兼容 |
| 性能下降 | 中 | 低 | 路径缓存，只在需要时生成 |
| 用户遗忘Key | 高 | 高 | 清晰的错误提示 + 示例 |
| 路径冲突 | 低 | 低 | 路径天然唯一 |

### 8.2 回滚计划

如果出现问题：
1. 禁用路径生成（通过环境变量）
2. 回退到纯用户Key模式
3. 提供迁移工具

---

## 9. 成功标准

✅ **功能正确**
- 静态UI自动生成Key
- 动态列表强制要求Key
- 用户Key优先

✅ **性能可接受**
- 路径生成开销 < 5%
- Key查找性能不变

✅ **开发体验**
- 清晰的错误提示
- 简单的迁移指南
- 完整的文档

---

**创建日期**: 2026-02-12
**最后更新**: 2026-02-12
**状态**: 🟡 准备实施
