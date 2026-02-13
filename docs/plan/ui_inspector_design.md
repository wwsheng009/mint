# UI Inspector - 开发设计说明书

**文档版本**: 1.0
**创建日期**: 2025-01-07
**状态**: 设计阶段

---

## 1. 概述

### 1.1 目标

创建一个强大的**UI元素检查器**（UI Inspector），帮助开发者在运行时调试和分析Mint TUI应用的布局问题。

### 1.2 核心价值

- 📊 **可视化布局信息**：实时显示元素的位置、尺寸、约束
- 🔍 **调试flex布局**：显示flex属性、自然宽度vs布局宽度
- 🎯 **快速定位问题**：悬停或选中元素立即显示详细信息
- 📝 **开发者友好**：清晰的格式化输出，易于理解

### 1.3 使用场景

```go
// 开发者只需设置环境变量
TUI_INSPECTOR=true go run myapp.go

// 或在代码中启用
os.Setenv("TUI_DEBUG_INSPECTOR", "true")
```

---

## 2. 功能规格

### 2.1 核心功能

#### 功能1: 元素信息面板

**显示位置**: 侧边栏或浮动窗口

**显示内容**:
```
┌─ UI Inspector ─────────────────────────────────┐
│ Element: ButtonVNode                           │
│ ├── Type                                      │
│ │   VNode Type: VNodeComponent                │
│ │   Tag: button                               │
│ │   Class: .btn-primary                       │
│ ├── Position                                  │
│ │   X: 1                                      │
│ │   Y: 12                                     │
│ ├── Size                                      │
│ │   Width: 18                                 │
│ │   Height: 1                                 │
│ ├── Layout                                    │
│ │   Natural Width: 14                         │
│ │   Layout Width: 18 ✅                       │
│ │   Padding: +4                               │
│ │   Flex: 1                                   │
│ ├── Bounds                                    │
│ │   [x: 1, y: 12, w: 18, h: 1]               │
│ ├── Constraints                               │
│ │   MinWidth: 18                              │
│ │   MaxWidth: 18                              │
│ │   MinHeight: 0                              │
│ │   MaxHeight: 1                              │
│ └── Properties                                │
│     Label: "[1] Event"                        │
│     HasFocus: true                            │
│     Disabled: false                           │
└────────────────────────────────────────────────┘
```

#### 功能2: 交互模式

**模式A: 鼠标悬停（默认）**
- 鼠标移动到元素上时，自动显示该元素的信息
- 实时更新，跟随鼠标位置

**模式B: 键盘导航**
- 按 `F12` 或 `Ctrl+I` 切换检查器开关
- 按 `Tab` 在元素间切换
- 按 `Enter` 查看选中元素的详细信息
- 按 `Esc` 关闭检查器

**模式C: 点击选中**
- 点击元素后锁定显示
- 再次点击取消锁定

#### 功能3: 布局可视化

**边框高亮**:
- 选中元素时显示边框
- 子元素显示不同颜色的边框
- Flex子元素用特殊颜色标记

**尺寸标注**:
- 在元素上方显示宽度
- 在元素左侧显示高度
- 箭头指示约束关系

#### 功能4: 布局树视图

```
┌─ Layout Tree ─────────────────────────────────┐
│ 📦 VStack (root)                              │
│   📦 BorderedNode                             │
│     📦 HStack (Header)                        │
│       📦 Text "Runtime..."                    │
│   📦 BorderedNode                             │
│     📦 HStack (Pipeline)                      │
│       📦 Text "[Event]"                       │
│       📦 Text "[setState]"                    │
│   📦 BorderedNode                             │
│     📦 HStack ⭐ FLEX WRAP                    │
│       🔵 Button "[1] Event]" 18x1             │
│       🔵 Button "[2]setState]" 18x1           │
│       🔵 Button "[3]Scheduler]" 18x1          │
└────────────────────────────────────────────────┘
```

**图标说明**:
- 📦 容器元素
- 🔵 Button/可交互元素
- 📝 Text元素
- ⭐ Flex容器

---

## 3. 技术架构

### 3.1 组件结构

```
internal/inspector/
├── inspector.go          # 主检查器
├── element_info.go       # 元素信息提取
├── overlay.go            # 视觉覆盖层
├── sidebar.go            # 侧边栏面板
├── keybinding.go         # 键盘交互
└── tree_view.go          # 布局树视图
```

### 3.2 核心接口

#### Inspector 接口

```go
type Inspector struct {
    enabled       bool
    selectedVNode ui.VNode
    hoveredVNode  ui.VNode
    overlay       *Overlay
    sidebar       *Sidebar
    treeView      *TreeView
}

// 启动检查器
func NewInspector() *Inspector

// 处理鼠标事件
func (i *Inspector) HandleMouseEvent(x, y int) bool

// 处理键盘事件
func (i *Inspector) HandleKeyEvent(event KeyEvent) bool

// 渲染检查器UI
func (i *Inspector) Paint(buf *paint.Buffer) error

// 查找指定位置的VNode
func (i *Inspector) FindVNodeAt(x, y int) ui.VNode
```

#### ElementInfo 结构

```go
type ElementInfo struct {
    VNode       ui.VNode
    Type        string
    Tag         string
    Position    Position
    Size        Size
    Layout      LayoutInfo
    Bounds      [4]int
    Constraints BoxConstraints
    Properties  map[string]interface{}
}

type LayoutInfo struct {
    NaturalWidth  int
    LayoutWidth   int
    Padding       int
    Flex          int
    IsFlexChild   bool
}

// 从VNode提取信息
func ExtractElementInfo(vnode ui.VNode) ElementInfo
```

### 3.3 VNode查找算法

```go
// 使用ComputedLayout进行高效查找
func (i *Inspector) FindVNodeAt(x, y int) ui.VNode {
    // 1. 从根节点开始
    // 2. 递归检查子节点
    // 3. 找到包含(x,y)的最深层节点
    // 4. 返回该节点
}
```

### 3.4 视觉覆盖层

```go
type Overlay struct {
    selectedBorder []rune   // 选中元素的边框样式
    childBorder    []rune   // 子元素边框样式
    flexBorder     []rune   // Flex子元素边框样式
    showDimensions bool     // 显示尺寸标注
}

// 在buffer上绘制覆盖层
func (o *Overlay) Paint(buf *paint.Buffer, vnode ui.VNode) {
    // 1. 绘制选中边框
    // 2. 绘制尺寸标注
    // 3. 绘制箭头和约束关系
}
```

---

## 4. 实现计划

### Phase 1: 基础信息提取 (1-2天)

**目标**: 提取VNode的布局信息

**任务**:
- [ ] 实现 `ElementInfo` 结构体
- [ ] 实现 `ExtractElementInfo()` 函数
- [ ] 支持基本信息：类型、标签、位置、尺寸
- [ ] 支持布局信息：自然宽度、布局宽度、flex属性
- [ ] 支持bounds和constraints

**验收标准**:
```go
info := ExtractElementInfo(button)
// info.Type == "ButtonVNode"
// info.NaturalWidth == 14
// info.LayoutWidth == 18
// info.Flex == 1
```

### Phase 2: 鼠标交互 (1天)

**目标**: 鼠标悬停显示元素信息

**任务**:
- [ ] 实现鼠标位置追踪
- [ ] 实现 `FindVNodeAt(x, y)` 算法
- [ ] 实现悬停高亮显示
- [ ] 实现简单的信息面板

**验收标准**:
- 鼠标移动到按钮上时显示按钮信息
- 信息面板实时更新

### Phase 3: 键盘导航 (1天)

**目标**: 键盘快捷键控制

**任务**:
- [ ] 实现快捷键注册系统
- [ ] F12/Ctrl+I 切换检查器
- [ ] Tab 在元素间切换
- [ ] Enter 查看详情
- [ ] Esc 关闭检查器

**验收标准**:
- 按F12打开/关闭检查器
- 按Tab在不同元素间切换

### Phase 4: 视觉增强 (1-2天)

**目标**: 美化显示效果

**任务**:
- [ ] 实现边框高亮
- [ ] 实现尺寸标注
- [ ] 实现不同颜色区分不同类型元素
- [ ] 实现动画效果（可选）

**验收标准**:
- 选中元素有明显的边框
- 尺寸标注清晰可见
- 不同类型元素颜色不同

### Phase 5: 侧边栏面板 (1天)

**目标**: 详细信息面板

**任务**:
- [ ] 实现侧边栏布局
- [ ] 格式化显示所有信息
- [ ] 支持折叠/展开
- [ ] 支持复制信息

**验收标准**:
- 信息面板结构清晰
- 所有信息正确显示

### Phase 6: 布局树视图 (1天)

**目标**: 可视化VNode树

**任务**:
- [ ] 实现树遍历
- [ ] 实现树状显示
- [ ] 支持展开/折叠节点
- [ ] 支持搜索节点

**验收标准**:
- 树结构正确显示
- 可以展开/折叠节点

### Phase 7: 高级功能 (可选)

**任务**:
- [ ] 性能分析（渲染时间、内存使用）
- [ ] 布局问题检测（约束冲突、溢出）
- [ ] 实时编辑属性（修改flex值立即看到效果）
- [ ] 导出布局报告

---

## 5. 数据流

### 5.1 启用流程

```
1. 检测环境变量 TUI_INSPECTOR=true
2. 创建 Inspector 实例
3. 注册到 framework.App
4. 注入到渲染管线
```

### 5.2 交互流程

```
用户移动鼠标
    ↓
Inspector.HandleMouseEvent(x, y)
    ↓
FindVNodeAt(x, y)
    ↓
ExtractElementInfo(vnode)
    ↓
更新侧边栏显示
    ↓
绘制视觉覆盖层
```

### 5.3 渲染流程

```
正常渲染
    ↓
PaintEngine.Paint(layout, buf)
    ↓
Inspector.Paint(buf)  // 额外的覆盖层
    ↓
绘制到终端
```

---

## 6. 配置选项

### 环境变量

```bash
# 启用检查器
TUI_INSPECTOR=true

# 检查器模式
TUI_INSPECTOR_MODE=hover|click|keyboard

# 显示详细程度
TUI_DEBUG_INSPECTOR=1|2|3

# 边框样式
TUI_INSPECTOR_BORDER=single|double|dashed
```

### 配置文件

```yaml
# .tui-inspector.yaml
inspector:
  enabled: true
  mode: hover
  sidebar:
    position: right  # left|right
    width: 40
  overlay:
    showBorders: true
    showDimensions: true
    showArrows: false
  colors:
    selected: "red"
    hovered: "yellow"
    flexChild: "cyan"
    container: "green"
```

---

## 7. 性能考虑

### 7.1 性能优化策略

1. **缓存元素信息**: 只有在VNode树变化时重新提取
2. **延迟查找**: 使用空间索引加速 `FindVNodeAt`
3. **按需渲染**: 只在检查器启用时渲染覆盖层
4. **节流更新**: 鼠标移动时限制更新频率（如30fps）

### 7.2 性能指标

- 查找元素: < 1ms
- 提取信息: < 0.5ms
- 渲染覆盖层: < 5ms
- 总体开销: < 10% 渲染时间

---

## 8. 测试计划

### 8.1 单元测试

```go
func TestExtractElementInfo(t *testing.T) {
    button := Button("Test")
    info := ExtractElementInfo(button)

    assert.Equal(t, "ButtonVNode", info.Type)
    assert.Equal(t, "button", info.Tag)
}

func TestFindVNodeAt(t *testing.T) {
    vnode := buildTestVTree()
    inspector := NewInspector()

    found := inspector.FindVNodeAt(5, 10)
    assert.Equal(t, expectedNode, found)
}
```

### 8.2 集成测试

- 在demo2中启用检查器
- 验证所有元素信息正确显示
- 验证交互功能正常

### 8.3 手动测试清单

- [ ] 鼠标悬停显示信息
- [ ] Tab切换元素
- [ ] 侧边栏信息正确
- [ ] 边框高亮正确
- [ ] 布局树显示正确
- [ ] 性能可接受

---

## 9. 文档需求

### 9.1 用户文档

- **README**: 如何启用和使用检查器
- **快捷键**: 所有快捷键列表
- **示例**: 截图和演示

### 9.2 开发者文档

- **架构文档**: 组件结构和接口
- **API文档**: 所有公开接口
- **贡献指南**: 如何添加新功能

---

## 10. 风险和限制

### 10.1 技术风险

- **性能影响**: 覆盖层渲染可能影响性能
- **兼容性**: 某些VNode可能无法提取信息
- **终端限制**: 某些终端可能不支持覆盖层

### 10.2 限制

- 只能在启用检查器的应用中使用
- 需要VNode支持bounds和constraints
- 复杂布局可能显示困难

---

## 11. 时间估算

| Phase | 任务 | 估算时间 |
|-------|------|---------|
| 1 | 基础信息提取 | 1-2天 |
| 2 | 鼠标交互 | 1天 |
| 3 | 键盘导航 | 1天 |
| 4 | 视觉增强 | 1-2天 |
| 5 | 侧边栏面板 | 1天 |
| 6 | 布局树视图 | 1天 |
| 7 | 高级功能 | 可选 |
| **总计** | | **7-9天** |

---

## 12. 下一步行动

1. ✅ 创建设计文档（本文档）
2. ⏳ 创建 `internal/inspector` 目录
3. ⏳ 实现 Phase 1: 基础信息提取
4. ⏳ 编写单元测试
5. ⏳ 在demo2中集成测试
6. ⏳ 收集反馈并迭代

---

## 13. 附录

### 13.1 参考资料

- Chrome DevTools: 灵感来源
- React DevTools: 组件树视图
- Flutter DevTools: 性能分析

### 13.2 相关Issue

- FillWidth调试需求
- 布局问题追踪

### 13.3 术语表

- **VNode**: Virtual Node，虚拟节点
- **Bounds**: 边界，元素的布局位置和尺寸
- **Constraints**: 约束，布局约束条件
- **Flex**: Flex布局，弹性布局
- **Natural Width**: 自然宽度，内容本身的宽度
- **Layout Width**: 布局宽度，flex计算后的宽度

---

**文档维护**: 本文档应随实现进展持续更新
**最后更新**: 2025-01-07
