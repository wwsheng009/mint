# ScrollView 组件 - 内容溢出解决方案

**ScrollView Component - Content Overflow Solution**

---

## 🎯 问题描述

用户反馈：
> "现在存在一个问题，如果内容高度超过了设置的固定的高度，如何处理，现在是直接溢出了，是否要设计一个可滚动的组件"

在 Inspector 中，当树视图内容超过固定高度（25行）时，内容会直接溢出，用户无法查看完整内容。

---

## ✅ 解决方案：ScrollView 组件

创建了一个通用的 **ScrollView 可滚动容器组件**，支持虚拟滚动。

### 设计思路

1. **通用性** - ScrollView 是一个独立的布局组件，可以在任何需要滚动功能的地方使用
2. **虚拟滚动** - 只渲染可见区域的内容，提高性能
3. **状态管理** - 滚动位置由外部维护（如 Inspector 的 treeScrollOffset）
4. **简单集成** - 通过 Builder 模式提供流畅的 API

### 组件位置

**文件**: `components/layout/scroll_view.go`

---

## 📦 组件特性

### ScrollViewBuilder API

```go
scrollView := layout.NewScrollView(content).
    Width(80).              // 视口宽度
    Height(20).             // 视口高度（可见行数）
    ScrollOffset(5).        // 当前滚动位置
    ShowBorder(true).       // 可选：显示边框
    Style(...).             // 可选：设置样式
    Build()
```

### 参数说明

| 参数 | 类型 | 说明 |
|------|------|------|
| `content` | `ui.VNode` | 要显示的内容（任何 VNode） |
| `Width` | `int` | 视口宽度（字符数） |
| `Height` | `int` | 视口高度（行数） |
| `ScrollOffset` | `int` | 当前进度条位置（行偏移） |
| `ShowBorder` | `bool` | 是否显示边框（默认 false） |
| `Style` | `style.Style` | 文本样式 |

---

## 🔄 工作原理

### 1. 内容提取

```go
// 从 VNode 中提取文本内容
contentText := b.extractTextContent(b.content)
lines := strings.Split(contentText, "\n")
```

### 2. 虚拟渲染

```go
// 计算可见范围
startLine := scrollOffset
endLine := startLine + viewportHeight

// 只渲染可见行
visibleLines := lines[startLine:endLine]
visibleText := strings.Join(visibleLines, "\n")
```

### 3. 滚动指示器

```go
if totalLines > viewportHeight {
    indicator := " ▼"  // 在顶部
    if scrollOffset >= maxOffset {
        indicator = " ▲"  // 在底部
    } else if scrollOffset > 0 {
        indicator = " ↕"  // 在中间
    }
    // 显示滚动指示器
}
```

---

## 🎨 使用示例

### 示例 1：Inspector 树视图

```go
// Inspector 中使用 ScrollView
allLines, totalLines := si.treeView.GetTreeLines()
si.treeTotalLines = totalLines

treeViewHeight := si.overlayHeight - 14

// 创建树文本内容
treeText := strings.Join(allLines, "\n")
treeContentNode := ui.Text(treeText)

// 使用 ScrollView 包装
treePreview := layout.NewScrollView(treeContentNode).
    Width(si.overlayWidth - 4).
    Height(treeViewHeight).
    ScrollOffset(si.treeScrollOffset).  // 由 Inspector 维护
    Build()
```

### 示例 2：带边框的日志查看器

```go
// 创建带边框的滚动日志查看器
logContent := ui.Text(allLogs)

logViewer := layout.NewScrollView(logContent).
    Width(100).
    Height(30).
    ScrollOffset(scrollPos).
    ShowBorder(true).  // 显示边框
    Build()
```

### 示例 3：状态管理

```go
// 在组件外部维护滚动状态
type MyComponent struct {
    scrollOffset int
}

func (c *MyComponent) Render() ui.VNode {
    content := ui.Text(c.getLongContent())

    return layout.NewScrollView(content).
        Width(80).
        Height(20).
        ScrollOffset(c.scrollOffset).  // 使用状态中的滚动位置
        Build()
}

// 键盘事件处理
func (c *MyComponent) HandleKeyDown(key string) {
    switch key {
    case "pgdn":
        c.scrollOffset += 20  // 向下滚动一页
    case "pgup":
        c.scrollOffset -= 20  // 向上滚动一页
    }
}
```

---

## 🎮 Inspector 集成

### 1. 状态字段

```go
type StandaloneInspector struct {
    // ...
    treeScrollOffset int  // 滚动位置
    treeTotalLines   int  // 总行数
}
```

### 2. 键盘事件处理

```go
func (si *StandaloneInspector) HandleKeyEvent(key string, alt, ctrl bool) bool {
    // 只在 Elements 标签页处理滚动
    if si.activeTab == TabElements {
        treeViewHeight := si.overlayHeight - 14

        switch key {
        case "pgup":
            si.treeScrollOffset -= treeViewHeight
            // 边界检查...
            return true
        case "pgdn":
            si.treeScrollOffset += treeViewHeight
            // 边界检查...
            return true
        case "home":
            si.treeScrollOffset = 0
            return true
        case "end":
            si.treeScrollOffset = maxOffset
            return true
        }
    }
}
```

### 3. 更新说明

```go
instructions := rtui.VStack(
    app.NewTextBuilder("Instructions:").Build(),
    app.NewTextBuilder("  PgUp/PgDn: Scroll tree").Build(),
    app.NewTextBuilder("  Home/End: Top/Bottom").Build(),
    // ...
)
```

---

## 🚀 性能优势

### 虚拟滚动

```
传统方式（全部渲染）：
- 1000 行内容
- 创建 1000 个 VNode
- 全部参与 Diff 和 Paint
- CPU 和内存消耗大

ScrollView（虚拟滚动）：
- 1000 行内容
- 只渲染可见的 20 行
- 只有 20 个 VNode 参与
- 性能提升 50 倍
```

### 内存优化

```go
// 传统方式
for i := 0; i < 1000; i++ {
    children = append(children, renderLine(i))
}
// 内存: 1000 个 VNode 对象

// ScrollView
visibleLines := allLines[start:end]  // 只有 20 行
// 内存: 20 个 VNode 对象
```

---

## 📊 与其他方案的对比

| 方案 | 优点 | 缺点 | 适用场景 |
|------|------|------|---------|
| **ScrollView** | ✅ 通用组件<br>✅ 虚拟滚动<br>✅ 简单 API<br>✅ 高性能 | ⚠️ 文本内容限制 | 长文本、日志、树视图 |
| **VirtualList** | ✅ 通用组件<br>✅ 虚拟滚动<br>✅ 支持复杂项 | ⚠️ 需要 item renderer | 列表、表格 |
| **手动分页** | ✅ 完全控制 | ❌ 代码复杂<br>❌ 不可复用 | 特殊需求 |

---

## 🔧 扩展方向

### 未来增强

1. **水平滚动支持**
   ```go
   .ScrollOffsetX(x)
   .ScrollOffsetY(y)
   ```

2. **动态内容加载**
   ```go
   .OnScroll(func(offset int) {
       // 懒加载更多内容
   })
   ```

3. **滚动条样式**
   ```go
   .ScrollbarStyle(...)
   ```

4. **内容缓存**
   ```go
   .EnableCache(true)
   ```

---

## 📁 相关文件

| 文件 | 说明 |
|------|------|
| `components/layout/scroll_view.go` | ScrollView 组件实现 |
| `components/layout/virtual_scroll.go` | VirtualList 组件（更复杂） |
| `internal/inspector/standalone_inspector.go` | Inspector 使用 ScrollView |
| `app/app.go` | ScrollView 导出 |

---

## 🎓 使用最佳实践

### ✅ 推荐

```go
// 1. 在外部维护滚动状态
type MyComponent struct {
    scrollOffset int
}

// 2. 使用 Builder 模式
scrollView := layout.NewScrollView(content).
    Width(80).
    Height(20).
    ScrollOffset(c.scrollOffset).
    Build()

// 3. 处理边界情况
if c.scrollOffset < 0 {
    c.scrollOffset = 0
}
maxOffset := totalLines - viewportHeight
if c.scrollOffset > maxOffset {
    c.scrollOffset = maxOffset
}
```

### ❌ 避免

```go
// ❌ 不要在 ScrollView 内部维护状态
// （TUI 是声明式的，每次渲染都重新创建）

// ❌ 不要忘记处理边界
// （会导致越界错误）

// ❌ 不要设置过小的 viewport
// （用户体验差）
```

---

## 🎯 总结

### 问题解决

✅ **内容溢出** - 通过虚拟滚动只渲染可见内容
✅ **性能问题** - 大幅减少 VNode 数量
✅ **通用性** - 可在任何需要滚动的地方使用
✅ **简单性** - 清晰的 Builder API

### 架构价值

1. **关注点分离** - 滚动逻辑与业务逻辑分离
2. **可复用性** - 一个组件服务整个项目
3. **可维护性** - 集中管理滚动行为
4. **可测试性** - 独立的组件易于测试

---

**版本**: 1.0
**状态**: ✅ 已实现并集成
**日期**: 2025-02-08
**作者**: Claude + 用户协作
