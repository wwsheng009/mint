# 基于路径的组件Key自动生成系统

> **目标**: 实现类似React的组件路径系统，自动为每个组件生成唯一Key
> **格式**: `/root/{layer}[{index}]/{type}[{index}]/.../{userKey}`
> **核心原则**: 系统路径（位置） + 用户Key（语义） = 完整唯一Key

---

## 1. 路径格式设计

### 1.1 路径字符串格式

```
格式: /root/{layer}[{index}]/{type}[{index}]/.../{type}[{index}]

示例:
/root/base[0]/panel[0]/vstack[0]/button[0]
/root/modal[1]/dialog[0]/form[0]/input[2]
/root/inspector[2]/tree[0]/node[15]
```

### 1.2 路径组成元素

| 组成部分 | 说明 | 示例 |
|---------|------|------|
| **root** | 固定根节点 | `root` |
| **layer** | 渲染层级 | `base`, `modal`, `overlay`, `inspector` |
| **type** | 组件类型 | `panel`, `list`, `button`, `input` |
| **index** | 同类型组件索引 | `[0]`, `[1]`, `[2]` |

### 1.3 完整路径示例

```
用户代码:
├─ VStack()
│  ├─ Panel().Key("my-panel")      // 系统路径: /root/base[0]/panel[0]
│  │  ├─ List()                    // 用户Key追加: /root/base[0]/panel[0]/my-panel
│  │  │  ├─ Button("OK").Key("btn-ok")        // 系统: /root/base[0]/panel[0]/my-panel/list[0]/button[0]
│  │  │  │                                       // 用户: /root/base[0]/panel[0]/my-panel/list[0]/btn-ok
│  │  │  ├─ Button("Cancel").Key("btn-cancel")  // 系统: /root/base[0]/panel[0]/my-panel/list[0]/button[1]
│  │  │  │                                       // 用户: /root/base[0]/panel[0]/my-panel/list[0]/btn-cancel
│  │  │  └─ Button("Apply")                     // 系统: /root/base[0]/panel[0]/my-panel/list[0]/button[2]
│  │  │                                         // 用户未设置: /root/base[0]/panel[0]/my-panel/list[0]/button[2]
│  └─ Panel()                     // 系统: /root/base[0]/panel[1]
│     └─ Input().Key("username")  // 系统: /root/base[0]/panel[1]/input[0]
│                                 // 用户: /root/base[0]/panel[1]/username
│
└─ Modal()                        // LayerModal
   └─ Button("Close").Key("btn-close")  // 系统: /root/modal[0]/button[0]
                                       // 用户: /root/modal[0]/btn-close
```

**关键规则**:
- ✅ 系统路径: 始终基于组件在树中的真实位置生成
- ✅ 用户Key: 只作为叶子节点追加到系统路径末尾
- ❌ 不允许: 用户Key覆盖或替换系统路径的任何部分

---

## 2. 实现架构

### 2.1 Fiber扩展

```go
// internal/reconciler/fiber_path.go

type Fiber struct {
    // ... 现有字段 ...

    // ✨ 新增: 路径相关字段
    Path         string       // 完整路径: /root/base[0]/panel[0]/button[1]
    PathSegment  string       // 当前路径段: button[1]
    Layer        ui.Layer     // 所属层级
    SiblingIndex int          // 在兄弟节点中的索引
    TypeIndex    map[string]int // 按类型统计的索引 {button: 2, panel: 1}
}
```

### 2.2 路径生成器

```go
// internal/reconciler/path_generator.go

type PathGenerator struct {
    // 路径段生成缓存
    segmentCache map[string]int
}

// GeneratePath 生成组件的完整路径
func (pg *PathGenerator) GeneratePath(
    parent *Fiber,
    vnode ui.VNode,
    siblingIndex int,
) string {
    // 1. 检查用户是否显式设置了Key
    if userKey := vnode.Key(); userKey != "" {
        return pg.generatePathWithUserKey(parent, userKey)
    }

    // 2. 自动生成基于路径的Key
    return pg.generateAutoPath(parent, vnode, siblingIndex)
}

// generatePathWithUserKey 用户设置了Key，追加到系统路径末尾
// 规则: 用户Key不覆盖系统路径，而是作为最后一个路径段
func (pg *PathGenerator) generatePathWithUserKey(
    parent *Fiber,
    vnode ui.VNode,
    userKey string,
    siblingIndex int,
) string {
    // 1. 先生成系统路径（基于组件真实位置）
    systemPath := pg.generateSystemPath(parent, vnode, siblingIndex)

    // 2. 将用户Key追加到系统路径末尾
    // 这确保了路径反映真实的树结构，同时保留用户的语义标识
    return systemPath + "/" + userKey
}

// generateSystemPath 生成基于组件位置的系统路径
func (pg *PathGenerator) generateSystemPath(
    parent *Fiber,
    vnode ui.VNode,
    siblingIndex int,
) string {
    // 1. 获取组件类型标识
    typeID := pg.getTypeIdentifier(vnode)

    // 2. 获取Layer信息
    layer := getLayer(vnode)

    // 3. 统计父节点中该类型的索引
    index := pg.getTypeIndex(parent, typeID, siblingIndex)

    // 4. 生成路径段
    segment := fmt.Sprintf("%s[%d]", typeID, index)

    // 5. 拼接完整路径
    if parent == nil {
        // 根节点，包含Layer信息
        return fmt.Sprintf("/root/%s[0]", getLayerName(layer))
    }
    return parent.Path + "/" + segment
}

// generateAutoPath 用户未设置Key，使用系统路径
func (pg *PathGenerator) generateAutoPath(
    parent *Fiber,
    vnode ui.VNode,
    siblingIndex int,
) string {
    // 直接使用系统路径作为Key
    return pg.generateSystemPath(parent, vnode, siblingIndex)
}

// getTypeIdentifier 获取组件类型标识
func (pg *PathGenerator) getTypeIdentifier(vnode ui.VNode) string {
    switch v := vnode.(type) {
    case *ui.ComponentVNode:
        return v.Name()
    case *ui.ElementVNode:
        return v.Tag()
    case *TextVNode:
        return "text"
    case *FragmentVNode:
        return "fragment"
    default:
        return "unknown"
    }
}

// getTypeIndex 获取组件类型在父节点中的索引
// siblingIndex: 当前节点在兄弟节点中的位置
// 返回值: 该类型在父节点中的索引（从0开始）
func (pg *PathGenerator) getTypeIndex(parent *Fiber, typeID string, siblingIndex int) int {
    if parent == nil {
        return 0
    }

    // 遍历父节点的子节点，统计同类型组件的数量
    // 注意：这里应该基于已经创建的兄弟节点来统计
    typeCount := 0
    child := parent.Child

    for i := 0; i < siblingIndex && child != nil; i++ {
        childSegment := extractPathSegment(child.Path)
        if getTypeIDFromSegment(childSegment) == typeID {
            typeCount++
        }
        child = child.Sibling
    }

    return typeCount
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

### 2.3 修改Fiber创建流程

```go
// internal/reconciler/diff.go

var pathGenerator = &PathGenerator{
    segmentCache: make(map[string]int),
}

// createChildFiber 创建子Fiber（修改版）
func createChildFiber(
    returnFiber *Fiber,
    vnode ui.VNode,
    lanes Lane,
    siblingIndex int,
) *Fiber {
    fiber := CreateFiberFromVNode(vnode)
    fiber.Return = returnFiber
    fiber.Lanes = lanes
    fiber.Props = vnode.Props()

    // ✨ 生成路径
    userKey := vnode.Key()
    if userKey != "" {
        // 用户设置了Key：系统路径 + 用户Key
        fiber.Path = pathGenerator.GeneratePathWithUserKey(returnFiber, vnode, userKey, siblingIndex)
        fiber.Key = fiber.Path  // Key = 完整路径
    } else {
        // 用户未设置Key：使用系统路径
        fiber.Path = pathGenerator.GenerateAutoPath(returnFiber, vnode, siblingIndex)
        fiber.Key = fiber.Path  // Key = 系统路径
        vnode.SetKey(fiber.Path)  // 同步到VNode
    }

    fiber.PathSegment = extractPathSegment(fiber.Path)
    fiber.Layer = getLayer(vnode)
    fiber.SiblingIndex = siblingIndex

    return fiber
}

// extractPathSegment 从完整路径中提取最后一段
func extractPathSegment(path string) string {
    parts := strings.Split(path, "/")
    return parts[len(parts)-1]
}
```

---

## 3. Layer处理

### 3.1 Layer顺序

```go
// runtime/ui/layer.go

const (
    LayerBase      Layer = iota  // 0 - 基础层
    LayerOverlay                 // 1 - 覆盖层
    LayerModal                   // 2 - 模态层
    LayerTooltip                 // 3 - 提示层
    LayerInspector               // 4 - Inspector层
)

func getLayerName(layer Layer) string {
    switch layer {
    case LayerBase:      return "base"
    case LayerOverlay:   return "overlay"
    case LayerModal:     return "modal"
    case LayerTooltip:   return "tooltip"
    case LayerInspector: return "inspector"
    default:             return "unknown"
    }
}
```

### 3.2 路径中包含Layer

```
示例1: 基础层按钮
/root/base[0]/panel[0]/button[0]

示例2: 模态层按钮
/root/modal[0]/dialog[0]/button[0]

示例3: Inspector层按钮
/root/inspector[0]/panel[0]/button[0]
```

---

## 4. 实例匹配流程

### 4.1 当前流程（问题）

```
问题:
1. 用户没设置Key → 不创建Instance
2. 事件路由失败 → 所有按钮触发第一个处理器
```

### 4.2 新流程（解决）

```
1. 用户没设置Key
   ↓
2. 自动生成路径Key: /root/base[0]/panel[0]/button[1]
   ↓
3. 创建Instance: key="vnode:/root/base[0]/panel[0]/button[1]"
   ↓
4. HitMap.NodeID = "btn-event" (用户设置的Key) 或 "/root/base[0]/panel[0]/button[1]" (自动生成的Key)
   ↓
5. 事件路由成功 ✅
```

---

## 5. Key生成规则

### 5.1 规则说明

| 场景 | 系统路径 | 用户Key | 最终Key（用于Instance） |
|------|---------|---------|----------------------|
| 用户设置Key | `/root/base[0]/panel[0]` | `my-panel` | `/root/base[0]/panel[0]/my-panel` |
| 用户未设置 | `/root/base[0]/panel[0]` | (空) | `/root/base[0]/panel[0]` |
| 嵌套+用户Key | `/root/base[0]/panel[0]/list[0]` | `btn-ok` | `/root/base[0]/panel[0]/list[0]/btn-ok` |

### 5.2 关键原则

✅ **系统路径优先**: 始终基于组件在树中的真实位置生成
✅ **用户Key追加**: 用户Key只作为最后一段路径
✅ **保留语义**: 用户Key提供业务语义（如"btn-ok"）
✅ **稳定唯一**: 路径天然唯一且稳定

### 5.3 伪代码

```go
func generateFiberKey(parent *Fiber, vnode VNode, siblingIndex int) string {
    // 1. 始终生成系统路径（反映组件真实位置）
    systemPath := generateSystemPath(parent, vnode, siblingIndex)

    // 2. 检查用户是否设置了Key
    userKey := vnode.Key()

    // 3. 决定最终Key
    if userKey != "" {
        // 用户Key: 追加到系统路径末尾
        return systemPath + "/" + userKey
    } else {
        // 无用户Key: 直接使用系统路径
        return systemPath
    }
}

// 示例
// Panel().Key("header") -> /root/base[0]/panel[0]/header
// Panel()               -> /root/base[0]/panel[0]
```

---

## 6. 关键设计原则

### 6.1 用户Key ≠ 完整Key

| 概念 | 说明 | 示例 |
|------|------|------|
| **系统路径** | 基于组件在树中的真实位置 | `/root/base[0]/panel[0]/list[0]` |
| **用户Key** | 用户提供的语义标识 | `btn-ok`, `username`, `login-form` |
| **完整Key** | 系统路径 + 用户Key | `/root/base[0]/panel[0]/list[0]/btn-ok` |

### 6.2 为什么用户Key不能覆盖系统路径？

❌ **错误设计**: 用户Key完全覆盖
```
用户设置: .Key("btn-ok")
最终Key: "btn-ok"

问题：
1. 两个不同位置的"btn-ok"会冲突
2. 失去组件在树中的位置信息
3. 调试时无法知道组件在哪里
4. 列表重排时无法正确匹配
```

✅ **正确设计**: 用户Key追加到系统路径
```
系统路径: /root/base[0]/panel[0]/list[0]
用户Key: "btn-ok"
最终Key: /root/base[0]/panel[0]/list[0]/btn-ok

优势：
1. 天然唯一（路径唯一）
2. 保留位置信息（便于调试）
3. 用户Key提供语义（便于理解）
4. 列表重排时路径更新（自动适配）
```

### 6.3 类比：文件系统路径

```
文件系统:
- 系统路径: /home/user/projects/app/src/main.go
- 文件名: main.go（用户可见）
- 完整路径: /home/user/projects/app/src/main.go（唯一标识）

Mint TUI:
- 系统路径: /root/base[0]/panel[0]/list[0]
- 用户Key: btn-ok（用户可见）
- 完整路径: /root/base[0]/panel[0]/list[0]/btn-ok（唯一标识）
```

## 7. 优势与挑战

### 6.1 优势

✅ **自动生成**: 用户不需要手动设置Key
✅ **唯一性保证**: 路径天然唯一
✅ **可调试**: 路径直接反映组件位置
✅ **Layer感知**: 路径包含层级信息
✅ **向后兼容**: 保留用户显式设置的Key

### 6.2 挑战

⚠️ **动态列表**: 列表项重排时路径可能变化
⚠️ **性能**: 路径生成有额外开销
⚠️ **长度**: 路径可能很长（可优化为哈希）

### 6.3 解决方案

| 问题 | 解决方案 |
|------|---------|
| 动态列表路径变化 | 对列表项强制要求用户设置Key |
| 性能开销 | 缓存路径段，延迟计算 |
| 路径过长 | 使用短类型名 + 可选哈希 |

---

## 8. 实现步骤

### Phase 1: 基础设施
1. 扩展Fiber结构，添加Path相关字段
2. 实现PathGenerator
3. 修改createChildFiber

### Phase 2: Layer集成
1. 在路径中包含Layer信息
2. 处理多层渲染
3. 更新HitMap匹配逻辑

### Phase 3: 调试工具
1. 添加路径可视化
2. Inspector显示组件路径
3. 路径搜索功能

### Phase 4: 优化
1. 路径缓存
2. 短路径生成
3. 性能测试

---

## 9. 测试用例

```go
func TestPathGeneration(t *testing.T) {
    // 测试1: 简单路径
    // Input: Panel > Button
    // Expected: /root/base[0]/panel[0]/button[0]

    // 测试2: 多个Button
    // Input: Panel > Button, Button, Button
    // Expected: /root/base[0]/panel[0]/button[0]
    //           /root/base[0]/panel[0]/button[1]
    //           /root/base[0]/panel[0]/button[2]

    // 测试3: 用户Key
    // Input: Panel(key="my-panel") > Button
    // Expected: /root/base[0]/my-panel/button[0]

    // 测试4: Modal层
    // Input: Modal > Button
    // Expected: /root/modal[0]/button[0]

    // 测试5: 嵌套结构
    // Input: VStack > Panel > List > Button
    // Expected: /root/base[0]/vstack[0]/panel[0]/list[0]/button[0]
}
```

---

## 10. 与React对比

| 特性 | React | Mint TUI (Proposed) |
|------|-------|-------------------|
| **Key来源** | 用户手动设置 | 自动生成 + 用户设置 |
| **路径系统** | 内部实现（不暴露） | 暴露给用户调试 |
| **默认行为** | index作为Key | 路径作为Key |
| **Layer支持** | ❌ 无 | ✅ 有 |
| **调试友好** | ⚠️ 一般 | ✅ 非常友好 |

---

## 11. 示例代码

### 10.1 使用前（需要手动设置Key）

```go
// ❌ 必须手动设置每个Key才能正确路由事件
func renderButtonList() ui.VNode {
    return ui.VStack(
        app.Button("Button 1").Key("btn-1").Build(),
        app.Button("Button 2").Key("btn-2").Build(),
        app.Button("Button 3").Key("btn-3").Build(),
    )
}

// 问题：
// - 不设置Key → 无Instance → 事件路由失败
// - Key冲突 → 路由到错误的处理器
```

### 10.2 使用后（自动生成Key）

```go
// ✅ 自动生成路径Key，事件路由正常工作
func renderButtonList() ui.VNode {
    return ui.VStack(
        app.Button("Button 1").Build(),
        app.Button("Button 2").Build(),
        app.Button("Button 3").Build(),
    )
}

// 系统自动生成：
// - Button 1: /root/base[0]/vstack[0]/button[0]
// - Button 2: /root/base[0]/vstack[0]/button[1]
// - Button 3: /root/base[0]/vstack[0]/button[2]
```

### 10.3 推荐模式（系统路径 + 用户语义）

```go
// ✅ 系统路径保证唯一性，用户Key提供语义
func renderForm() ui.VNode {
    return ui.VStack(
        app.Panel().Key("login-form").Build(
            app.Input("Username").Key("username").Build(),
            app.Input("Password").Key("password").Build(),
            app.Button("Login").Key("btn-login").Build(),
        ),
    )
}

// 生成的完整路径：
// - Panel:     /root/base[0]/vstack[0]/panel[0]/login-form
// - Username:  /root/base[0]/vstack[0]/panel[0]/login-form/input[0]/username
// - Password:  /root/base[0]/vstack[0]/panel[0]/login-form/input[1]/password
// - Login:     /root/base[0]/vstack[0]/panel[0]/login-form/button[0]/btn-login
```

### 10.4 Layer示例

```go
// ✅ Layer信息自动包含在路径中
func renderApp() ui.VNode {
    return ui.LayerStack(
        ui.VStack(
            app.Button("Main Button").Build(),
            // -> /root/base[0]/vstack[0]/button[0]
        ),
        ui.Modal(
            app.Button("Close").Key("modal-close").Build(),
            // -> /root/modal[0]/button[0]/modal-close
        ),
        ui.Inspector(
            app.Button("Inspect").Build(),
            // -> /root/inspector[0]/button[0]
        ),
    )
}
```

---

**结论**: 这个设计提供了自动化的Key生成系统，系统路径保证组件的唯一性和位置信息，用户Key提供业务语义。两者结合既保证正确性又提供可读性。

---

## 附录：关键设计对比

### 旧设计（错误理解）

```
用户Key: "btn-ok"
最终Key: "btn-ok"  ❌ 覆盖系统路径

问题：
- 失去位置信息
- 无法区分同名的不同组件
- 调试困难
```

### 新设计（正确理解）

```
系统路径: /root/base[0]/panel[0]/list[0]
用户Key: "btn-ok"
最终Key: /root/base[0]/panel[0]/list[0]/btn-ok  ✅

优势：
- 保留完整的位置信息
- 用户Key提供语义
- 天然唯一性
- 便于调试
```
