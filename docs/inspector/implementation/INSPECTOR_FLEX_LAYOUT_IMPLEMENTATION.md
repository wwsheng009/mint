# Inspector Flex布局实现

## 问题

用户反馈：**"为什么不用布局系统进行布局，比如flex等工具"**

之前的实现使用手动计算高度：
```go
treeViewHeight := si.overlayHeight - 14  // 手动计算
scrollContainer := layout.NewScrollView(treePreview).
    Height(treeViewHeight).  // 固定高度
    Build()
```

## 解决方案

使用Mint TUI的`ui.Flex()`进行自动布局：

### 1. 使用ui.Flex包装内容

```go
// standalone_inspector.go buildElementsTabContent()
return rtui.VStack(
    header,                           // 固定高度
    selectedInfo,                     // 固定高度
    rtui.Flex(treeWithStatus, 1),     // flex: 1 - 自动填充剩余空间 ✅
    instructions,                     // 固定高度
)
```

**关键变化**：
- `rtui.Flex(treeWithStatus, 1)` 给treeWithStatus设置`flex: 1`
- VStack会自动计算剩余空间并分配给flex=1的子元素
- 不需要手动计算`overlayHeight - 14`

### 2. ScrollView支持auto-height模式

```go
// components/layout/scroll_view.go
func (b *ScrollViewBuilder) Build() ui.VNode {
    // Auto-height mode: if height is 0 or not set
    if b.height <= 0 {
        // 渲染全部内容（不裁剪）
        visibleText := contentText

        textNode.SetProps(ui.Props{
            "flex": 1,               // 允许flex扩展
            "scroll-content": ...,   // 存储原始内容用于滚动
            "total-lines": ...,      // 总行数
        })

        return textNode
    }

    // Fixed-height mode: 裁剪内容到viewport高度
    // ...
}
```

**两种模式**：
- **Auto-height mode** (height=0): 渲染全部内容，依赖父容器flex约束
- **Fixed-height mode** (height>0): 手动裁剪内容到指定高度

### 3. Inspector使用auto-height模式

```go
// standalone_inspector.go
scrollContainer := layout.NewScrollView(treePreview).
    // Height() NOT set - 使用auto-height模式
    Width(si.overlayWidth - 4).
    ScrollOffset(si.treeScrollOffset).
    Build()
```

**好处**：
- ScrollView渲染全部内容
- 父容器的`ui.Flex()`约束实际高度
- 高度自动适应overlay大小变化

## Flex布局工作原理

### Flexbox类比

```
CSS Flexbox                    Mint TUI Flex
────────────────────────────────────────────────
.container {                   rtui.VStack(
    display: flex;                header,
    flex-direction: column;       selectedInfo,
}                                rtui.Flex(tree, 1),  ← flex-grow: 1
                                 instructions,
.item {                     }
    flex-grow: 1;    →    )
}
```

### 布局流程

```
1. VStack计算所有子元素的自然高度
   ├─ header:          3行 (固定)
   ├─ selectedInfo:    4行 (固定)
   ├─ treeWithStatus:  100行 (内容高度，但flex: 1)
   └─ instructions:    6行 (固定)

2. VStack容器高度 = 25行 (overlayHeight - 边框)

3. 计算剩余空间
   固定高度总和 = 3 + 4 + 6 = 13行
   剩余空间 = 25 - 13 = 12行

4. 分配给flex: 1的元素
   treeWithStatus高度 = min(内容高度100, 剩余空间12) = 12行

5. ScrollView在12行高度内显示内容
   (支持滚动查看全部100行)
```

## 测试验证

### TestInspectorFlexLayout输出

```
✅ Elements tab has 4 children
✅ Child #2 (treeWithStatus): *ui.LayoutNode
✅ treeWithStatus has flex=1 (should be 1)
✅ Flex layout correctly configured

✅ Flex Layout Summary:
  - Header: fixed height
  - SelectedInfo: fixed height
  - TreeWithStatus: flex: 1 (grows to fill space) ✅
  - Instructions: fixed height
```

## 对比

### 之前：手动计算
```go
treeViewHeight := overlayHeight - 14  // 硬编码
scrollContainer.Height(treeViewHeight)
```
❌ 不灵活
❌ 需要手动维护计算公式
❌ overlayHeight变化时需要更新

### 现在：Flex布局
```go
rtui.Flex(treeWithStatus, 1)  // 自动扩展
scrollContainer  // height不设置
```
✅ 自动适应
✅ 布局系统自动计算
✅ overlayHeight变化时自动调整

## 文件修改

### 1. `internal/inspector/standalone_inspector.go`
```diff
- treeViewHeight := si.overlayHeight - 14
- scrollContainer := layout.NewScrollView(treePreview).
-     Height(treeViewHeight).
-     Build()
- return rtui.VStack(header, selectedInfo, treeWithStatus, instructions)

+ scrollContainer := layout.NewScrollView(treePreview).
+     Build()
+ return rtui.VStack(
+     header,
+     selectedInfo,
+     rtui.Flex(treeWithStatus, 1),  // flex: 1
+     instructions,
+ )
```

### 2. `components/layout/scroll_view.go`
```diff
func (b *ScrollViewBuilder) Build() ui.VNode {
+   // Auto-height mode
+   if b.height <= 0 {
+       // Render all content, let parent flex constrain
+       textNode.SetProps(ui.Props{"flex": 1, ...})
+       return textNode
+   }
+
    // Fixed-height mode (existing code)
    ...
}
```

## Flex API参考

### ui.Flex()

```go
// 让VNode在父容器中flex扩展
func Flex(vnode VNode, flexFactors ...int) VNode

// 用法
ui.Flex(content, 1)  // flex-grow: 1
ui.Flex(content)     // 默认 flex-grow: 1
```

### Props

```go
ui.Props{
    "flex": 1,              // flex-grow系数
    "scroll-content": ...,  // ScrollView: 原始内容
    "total-lines": ...,     // ScrollView: 总行数
}
```

## 总结

使用Flex布局的优势：

1. **自动计算空间** - 不需要手动`overlayHeight - 14`
2. **响应式布局** - overlay尺寸变化时自动调整
3. **声明式API** - 清楚表达布局意图（flex: 1）
4. **标准做法** - 符合现代UI框架模式（React Flexbox）

这正是用户期望的："用布局系统进行布局，比如flex等工具" ✅
