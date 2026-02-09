# Inspector TreeView 溢出问题分析与解决方案

## 🔍 问题现象

从截图观察，Inspector的Layout Tree内容**溢出了边框**：

```
╔═ INSPECTOR ═╗
│📦 Layout Tree │
│*   ┌─Layout Tree │
│     └──📦LayoutNode  ← 内容继续向下，超出边框底部
│   │├── 🖼️BorderedNode
│   │  └── 📦LayoutNode
│...更多内容...
╚════════════════════  ← 边框在这里结束，但内容还在继续
```

## 📊 根本原因分析

### 1. TreeView渲染所有行

**文件**: `components/display/treeview.go:760`

```go
func (t *TreeView) regenerateDisplay() {
    var lineNodes []ui.VNode
    for i, line := range t.lines {  // ← 遍历所有100行！
        // 为每一行创建VNode
        lineNodes = append(lineNodes, ...)
    }
    result := ui.VStack(lineNodes...)  // ← 渲染所有行
    t.currentRender = result
}
```

**问题**：即使TreeView有`SetViewportHeight(20)`，它仍然渲染所有100行！

### 2. VStack渲染所有子节点

**文件**: `internal/inspector/standalone_inspector.go:627-632`

```go
return rtui.VStack(
    header,         // 3行
    selectedInfo,   // 4行
    treeWithStatus,  // 100行！
    instructions,    // 6行
)
```

VStack不裁剪内容，会渲染所有子节点。

### 3. Bordered的Height()不裁剪内容

**文件**: `internal/inspector/standalone_inspector.go:329-335`

```go
panel := rtui.Bordered().
    Style("double").
    Label("INSPECTOR").
    Child(content).
    Width(si.overlayWidth).
    Height(si.overlayHeight).  // ← 只是布局提示
    Build()
```

`Height()`只是在布局计算时使用，**不会实际裁剪渲染内容**。

## 🎯 解决方案

### 方案1：修改TreeView只渲染可见行（正确）

**优点**：
- ✅ 保留TreeView的所有交互性
- ✅ 性能最好（只渲染可见内容）
- ✅ 真正的虚拟滚动

**缺点**：
- ❌ 需要大量修改TreeView组件
- ❌ 复杂度高

**实现**：

修改`components/display/treeview.go:760`：

```go
func (t *TreeView) regenerateDisplay() {
    var lineNodes []ui.VNode

    // 计算可见范围
    startLine := t.scrollOffset
    endLine := startLine + t.viewportHeight
    if endLine > len(t.lines) {
        endLine = len(t.lines)
    }

    // 只渲染可见行
    for i := startLine; i < endLine; i++ {
        line := t.lines[i]
        // ... 创建VNode
        lineNodes = append(lineNodes, ...)
    }

    result := ui.VStack(lineNodes...)
    t.currentRender = result
}
```

### 方案2：使用ScrollView包装（临时）

**优点**：
- ✅ 快速实现
- ✅ 可以约束内容高度

**缺点**：
- ❌ 会丢失TreeView的交互性（焦点、选择、展开/折叠）
- ❌ ScrollView会把VNode转换成文本

**实现**：

修改`internal/inspector/standalone_inspector.go:597-603`：

```go
// Wrap TreeView in ScrollView to constrain height
scrollView := layout.NewScrollView(treePreview).
    Height(treeViewHeight).
    Width(si.overlayWidth - 4).
    ScrollOffset(si.treeScrollOffset).
    Build()

// Combine status line with scrolled tree
var treeWithStatus ui.VNode
if len(statusLines) > 0 {
    treeWithStatus = rtui.VStackBuilder(append(statusLines, scrollView)...).Build()
} else {
    treeWithStatus = scrollView
}
```

### 方案3：TreeView组件提供GetVisibleLines()方法

**优点**：
- ✅ 保持TreeView的交互性
- ✅ 只需添加新方法

**实现**：

在TreeView组件中添加：

```go
// GetVisibleLines returns only the visible lines based on viewport
func (t *TreeView) GetVisibleLines() []ui.VNode {
    var lineNodes []ui.VNode

    startLine := t.scrollOffset
    endLine := startLine + t.viewportHeight
    if endLine > len(t.lines) {
        endLine = len(t.lines)
    }

    for i := startLine; i < endLine; i++ {
        line := t.lines[i]
        // ... 创建VNode（与regenerateDisplay相同）
        lineNodes = append(lineNodes, ...)
    }

    return lineNodes
}
```

然后在Inspector中使用：

```go
// Get only visible lines from TreeView
visibleLines := si.treeViewComponent.GetVisibleLines()
treePreview := ui.VStack(visibleLines...)
```

## 📝 推荐方案

**短期**：使用方案2（ScrollView），快速修复溢出问题

**长期**：实施方案1（修改TreeView），实现真正的虚拟滚动

## 🎨 为什么Bordered.Height()不裁剪内容？

在TUI中，`Height()`有两个作用：

1. **布局阶段**：告诉布局引擎这个组件需要多高
2. **渲染阶段**：某些组件会根据Height()裁剪内容

但是，**大多数组件不裁剪内容**！它们会：
- 渲染所有子节点
- 然后由上层（如终端/viewport）决定是否显示

**需要内容裁剪的组件**必须自己实现：
- ScrollView - 提取可见内容
- VirtualList - 只渲染可见项
- TextView (可能) - 只渲染可见文本

普通容器（VStack, HStack, Bordered）**不裁剪内容**。

## 📚 参考资料

- `components/layout/scroll_view.go` - ScrollView实现
- `components/display/treeview.go` - TreeView实现
- `internal/inspector/standalone_inspector.go` - Inspector实现

## 🔗 相关问题

- React中的虚拟滚动：`react-window`, `react-virtualized`
- TUI中的视口裁剪：需要手动实现
