# 方案 A：边框作为容器属性 - 实施计划

## 概述

将边框从独立组件改为容器的视觉属性，所有容器天然支持边框，无需额外包裹。

## 核心理念

边框是容器的视觉装饰属性，而非功能组件。通过 Fiber 的属性系统传递边框配置。

## 目标

1. 所有容器组件（Stack/Grid/Wrap/Absolute）原生支持边框
2. Modal 成为纯容器组件，移除复杂的 Instance
3. 简化组件层次，减少组合复杂度
4. 保持向后兼容性

---

## 实施步骤

### Phase 1: Fiber 层扩展边框属性

#### 1.1 扩展 Fiber 结构
```go
// runtime/ui/fiber.go
type Fiber struct {
    // ... 现有字段 ...

    // ✨ 边框属性（新增）
    BorderStyle  layout.BorderStyle  // 边框样式
    BorderLabel  string              // 边框标题
}
```

#### 1.2 扩展 ElementVNode 辅助方法
```go
// runtime/ui/element_vnode.go
type ElementVNode struct {
    // ... 现有字段 ...

    // ✨ 边框属性（新增）
    borderStyle  layout.BorderStyle
    borderLabel  string
}

// 边框设置方法
func (e *ElementVNode) SetBorderStyle(style layout.BorderStyle) *ElementVNode {
    e.borderStyle = style
    return e
}

func (e *ElementVNode) SetBorderLabel(label string) *ElementVNode {
    e.borderLabel = label
    return e
}
```

---

### Phase 2: 完善完成工作流程（completeWork）

#### 2.1 更新边框属性同步逻辑
```go
// internal/render/reconciler.go - completeWork 阶段

// 在 LayoutPropertiesComplete 函数中添加边框属性同步
func completeWork(fiber *Fiber) {
    // ... 现有逻辑 ...

    // ✨ 同步边框属性
    syncBorderProperties(fiber)

    // ... 其他属性同步 ...
}

// 新增：边框属性同步函数
func syncBorderProperties(fiber *Fiber) {
    // 从 Props 中读取边框属性
    if styleProp, ok := fiber.Props["borderStyle"].(layout.BorderStyle); ok {
        fiber.BorderStyle = styleProp
    }
    if labelProp, ok := fiber.Props["label"].(string); ok {
        fiber.BorderLabel = labelProp
    }
    if titleProp, ok := fiber.Props["title"].(string) && fiber.Tag == "modal" {
        fiber.BorderLabel = titleProp  // Modal 的 title 映射到边框标签
    }
}
```

---

### Phase 3: Stack 容器支持边框

#### 3.1 Stack VNode 添加边框属性
```go
// ui/components/stack/vnode.go
type VNode struct {
    *rtui.ElementVNode

    // ... 现有字段 ...

    // ✨ 边框属性
    borderStyle  layout.BorderStyle
    borderLabel  string
}
```

#### 3.2 添加边框构建方法
```go
// 设置边框
func (s *VNode) Border(style layout.BorderStyle, label string) *VNode {
    s.borderStyle = style
    s.borderLabel = label
    return s
}

// 带样式边框
func (s *VNode) Bordered(style layout.BorderStyle) *VNode {
    return s.Border(style, "")
}

// 单线边框
func (s *VNode) SingleBorder(label string) *VNode {
    return s.Border(layout.BorderSingle, label)
}

// 双线边框
func (s *VNode) DoubleBorder(label string) *VNode {
    return s.Border(layout.BorderDouble, label)
}
```

#### 3.3 更新 Props 方法
```go
func (s *VNode) Props() rtui.Props {
    props := rtui.Props{
        "key":           s.key,
        "direction":     s.direction,
        "align":         s.align,
        "crossAlign":    s.crossAlign,
        "gap":           s.gap,
        "padding":       s.padding,
        "width":         s.width,
        "height":        s.height,
        "flex":          s.flex,
        "style":         s.style,
        "borderStyle":   s.borderStyle,   // ✨ 新增
        "label":         s.borderLabel,   // ✨ 新增
    }
    return props
}
```

---

### Phase 4: Grid 容器支持边框

#### 4.1 Grid VNode 添加边框属性
```go
// ui/components/grid/vnode.go
type VNode struct {
    *rtui.ElementVNode

    // ... 现有字段 ...

    // ✨ 边框属性
    borderStyle  layout.BorderStyle
    borderLabel  string
}
```

#### 4.2 添加边框构建方法
```go
func (g *VNode) Border(style layout.BorderStyle, label string) *VNode {
    g.borderStyle = style
    g.borderLabel = label
    return g
}

func (g *VNode) Bordered(style layout.BorderStyle) *VNode {
    return g.Border(style, "")
}
```

---

### Phase 5: Wrap 容器支持边框

#### 5.1 Wrap VNode 添加边框属性
```go
// ui/components/wrap/vnode.go
type VNode struct {
    *rtui.ElementVNode

    // ... 现有字段 ...

    // ✨ 边框属性
    borderStyle  layout.BorderStyle
    borderLabel  string
}
```

#### 5.2 添加边框构建方法
```go
func (w *VNode) Border(style layout.BorderStyle, label string) *VNode {
    w.borderStyle = style
    w.borderLabel = label
    return w
}
```

---

### Phase 6: Absolute 容器支持边框

#### 6.1 Absolute VNode 添加边框属性
```go
// ui/components/absolute/vnode.go
type VNode struct {
    *rtui.ElementVNode

    // ... 现有字段 ...

    // ✨ 边框属性
    borderStyle  layout.BorderStyle
    borderLabel  string
}
```

#### 6.2 添加边框构建方法
```go
func (a *VNode) Border(style layout.BorderStyle, label string) *VNode {
    a.borderStyle = style
    a.borderLabel = label
    return a
}
```

---

### Phase 7: FiberToNodeAdapter 支持边框

#### 7.1 更新 GetBorder 方法
```go
// internal/render/fiber_adapter.go
func (a *FiberToNodeAdapter) GetBorder() layout.Border {
    if a.fiber == nil {
        return layout.Border{Style: layout.BorderNone}
    }

    // ✨ 优先使用 Fiber 的边框属性
    if a.fiber.BorderStyle.HasBorder() {
        return layout.Border{
            Style: a.fiber.BorderStyle,
            Label: a.fiber.BorderLabel,
        }
    }

    // 向后兼容：支持旧的 Props 方式
    switch a.fiber.Tag {
    case "bordered":
        style := layout.BorderNone
        if s, ok := a.fiber.Props["style"].(layout.BorderStyle); ok {
            style = s
        }
        label := ""
        if l, ok := a.fiber.Props["label"].(string); ok {
            label = l
        }
        return layout.Border{Style: style, Label: label}

    case "modal":
        style := layout.BorderDouble  // Modal 默认双线边框
        if s, ok := a.fiber.Props["borderStyle"].(layout.BorderStyle); ok {
            style = s
        }
        label := ""
        if l, ok := a.fiber.Props["title"].(string); ok {
            label = l
        }
        return layout.Border{Style: style, Label: label}
    }

    return layout.Border{Style: layout.BorderNone}
}
```

---

### Phase 8: Modal 重构为纯容器

#### 8.1 Modal VNode 简化
```go
// ui/components/modal/vnode.go
type VNode struct {
    *rtui.ElementVNode

    // ✨ 边框属性
    borderStyle  layout.BorderStyle
    title        string

    // 移除：不再需要独立的 Instance
    // Modal 可以作为带边框的 VStack 使用
}

func New(title string) *VNode {
    return &VNode{
        ElementVNode: rtui.NewElement("modal"),
        borderStyle:  layout.BorderDouble,  // Modal 默认双线边框
        title:        title,
    }
}

// 实现 InstanceFactory - 返回 nil 使用默认渲染
func (m *VNode) CreateInstance() rtui.ComponentInstance {
    return nil  // 使用 Fiber-to-Layout 布局
}

// 添加边框配置方法
func (m *VNode) Border(style layout.BorderStyle) *VNode {
    m.borderStyle = style
    return m
}
```

#### 8.2 Modal VNode Props 更新
```go
func (m *VNode) Props() rtui.Props {
    return rtui.Props{
        "key":         m.key,
        "title":       m.title,
        "borderStyle": m.borderStyle,  // ✨ 边框属性
        "children":    m.children,
        "style":       m.style,
    }
}
```

#### 8.3 移除或简化 Modal Instance
```go
// ui/components/modal/instance.go
// 如果 Modal 不需要特殊状态管理，可以：
// 1. 完全移除 instance.go 和 vnode.go 的 InstanceFactory 实现
// 2. 或保留为空实现（仅用于向后兼容）

type Instance struct {
    // 移除所有 Measure 和 Paint 逻辑
    // 这些由 Fiber-first 渲染管道处理
}
```

---

### Phase 9: Border 组件兼容层

#### 9.1 Border 保留为兼容装饰器
```go
// ui/components/border/vnode.go
// Border 组件保持不变，作为向后兼容的选项
// 创建时它内部会设置边框属性

type VNode struct {
    *rtui.ElementVNode

    style  layout.BorderStyle
    label  string
    // 边框作为一个属性容器
}

func New(style layout.BorderStyle) *VNode {
    return &VNode{
        ElementVNode: rtui.NewElement("bordered"),
        style:        style,
    }
}

func (b *VNode) SetLabel(label string) *VNode {
    b.label = label
    return b
}
```

#### 9.2 Border Instance 简化
```go
// ui/components/border/instance.go
// 简化 Instance，只处理渲染，不参与布局测量
// 边框的布局位置由 Fiber 处理
```

---

### Phase 10: 测试和验证

#### 10.1 单元测试
- [ ] Fiber 边框属性读写
- [ ] FiberToNodeAdapter.GetBorder() 各种场景
- [ ] 各容器 VNode 的边框配置

#### 10.2 集成测试
- [ ] Modal 渲染测试
- [ ] Stack/Grid/Wrap/Absolute 边框测试
- [ ] 边框偏移正确性测试

#### 10.3 更新示例代码
- [ ] 更新所有使用边框的示例
- [ ] 添加新的 API 使用文档

---

## API 变更示例

### 旧 API（需要组合）
```go
// 需要 Border + Stack 两层
border.New(layout.BorderDouble).SetLabel("标题").SetChildren(
    vstack.New().Padding(1, 1, 1, 1).SetChildren(
        text.New("内容"),
    ),
)
```

### 新 API（原生支持）
```go
// 直接在 Stack 上设置边框
vstack.New().
    Border(layout.BorderDouble, "标题").
    Padding(1, 1, 1, 1).
    SetChildren(
        text.New("内容"),
    )
```

---

## 向后兼容性

| 旧方式 | 新方式 | 兼容性 |
|--------|--------|--------|
| `border.New(...)` | `*.Border(...)` | ✅ 旧方式继续支持 |
| Border 组件 | 容器边框属性 | ✅ 旧组件保留 |
| Modal Instance | Modal 容器 | ⚠️ 需迁移（但 API 尽量一致） |

---

## 预期收益

1. **简化组件层次**：减少一层 Border 组件
2. **提升开发体验**：链式调用更直观
3. **降低学习成本**：一个 API 搞定所有容器
4. **更符合 Fiber-first**：通过 Fiber 属性传递
5. **易于扩展**：未来可添加背景、阴影等

---

## 风险评估

| 风险 | 可能性 | 影响 | 缓解措施 |
|------|--------|------|----------|
| 破坏现有组件 | 低 | 中 | 保留 Border 组件作为兼容层 |
| 性能下降 | 低 | 低 | 边框属性只是内存字段，无额外开销 |
| 布局错误 | 中 | 高 | 充分测试边框偏移逻辑 |
| 文档滞后 | 高 | 中 | 同步更新文档和示例 |

---

## 时间估算

| Phase | 任务 | 预计时间 |
|-------|------|----------|
| 1 | Fiber 层扩展 | 0.5 天 |
| 2 | completeWork 更新 | 0.5 天 |
| 3 | Stack 支持 | 0.5 天 |
| 4 | Grid 支持 | 0.5 天 |
| 5 | Wrap 支持 | 0.5 天 |
| 6 | Absolute 支持 | 0.5 天 |
| 7 | FiberToNodeAdapter | 1 天 |
| 8 | Modal 重构 | 1 天 |
| 9 | Border 兼容层 | 0.5 天 |
| 10 | 测试验证 | 1 天 |
| **合计** | | **7 天** |
