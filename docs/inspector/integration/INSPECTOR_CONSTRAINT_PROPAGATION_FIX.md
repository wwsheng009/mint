# Inspector 约束传播问题修复

## 问题诊断

用户报告 `examples/ui_demos/demo2_runtime_internals/inspector_overlay/main.go` 的约束没有生效。

经过深入调查，我发现了以下问题：

### 根本原因

在 `internal/inspector/standalone_inspector.go:494-506` 的 `buildElementsTabContent()` 方法中：

```go
} else {
    // Update existing component with new lines
    updated := display.NewTreeView().  // ← 问题：每次都创建新的 TreeView 实例！
        FromLines(allLines).
        ExpandLevel(1).
        ShowIcons(true).
        Compact(false).
        Build().(*display.TreeView)

    // Preserve navigation state
    updated.SetFocusIndex(si.treeViewComponent.GetFocusIndex())
    updated.SetScrollOffset(si.treeScrollOffset)
    si.treeViewComponent = updated  // ← 替换了整个实例
}
```

**问题**：每次 `buildElementsTabContent()` 被调用时，都会创建一个 **新的 TreeView 实例**。新实例的 `viewportHeight` 字段初始化为 0，丢失了之前由布局引擎设置的高度约束。

**症状**：
- 第一次渲染：`TreeView.Measure()` 被调用，设置 `viewportHeight=7`（正确）
- 第二次渲染：创建了新 TreeView 实例，`viewportHeight=0`（错误）
- 结果：虚拟滚动失效，TreeView 渲染所有 34 行而不是只渲染 7 行

## 解决方案

### 1. 添加 `UpdateLines()` 方法

**文件**: `components/display/treeview.go:661-685`

```go
// UpdateLines updates the tree lines without creating a new TreeView instance
// This preserves the viewportHeight that was set by the layout engine
func (t *TreeView) UpdateLines(lines []string) {
	if t.builder == nil {
		return
	}

	// Parse new lines
	t.builder.sourceLines = lines
	t.builder.parseLines(lines)

	// Update lines from the builder's node
	t.lines = t.builder.node.lines
	t.totalLines = len(t.lines)

	// Re-render with the new lines (preserving viewportHeight)
	t.regenerateDisplay()
}
```

**关键**：这个方法更新行数据 **而不重新创建实例**，因此保留了 `viewportHeight` 状态。

### 2. 修改 Inspector 以使用 `UpdateLines()`

**文件**: `internal/inspector/standalone_inspector.go:493-497`

```go
} else {
    // Update existing TreeView with new lines WITHOUT creating a new instance
    // This preserves the viewportHeight that was set by the layout engine
    si.treeViewComponent.UpdateLines(allLines)
}
```

**修改前**：每次调用 `NewTreeView().Build()` 创建新实例
**修改后**：调用 `UpdateLines()` 更新现有实例

## 测试验证

### 1. UpdateLines 保留 viewportHeight 测试

**文件**: `components/display/treeview_update_test.go`

```go
func TestTreeViewUpdateLinesPreservesViewportHeight(t *testing.T) {
    // 创建 TreeView
    treeView := NewTreeView().FromLines(lines).Build()

    // 第一次测量：设置 viewportHeight=5
    constraints1 := runtime.BoxConstraints{MaxHeight: 5}
    measurable.Measure(constraints1)
    assert.Equal(t, 5, treeView.viewportHeight)  // ✅

    // 更新行数据
    treeView.UpdateLines(newLines)
    assert.Equal(t, 5, treeView.viewportHeight)  // ✅ 仍然为 5！

    // 第二次测量：更新 viewportHeight=8
    constraints2 := runtime.BoxConstraints{MaxHeight: 8}
    measurable.Measure(constraints2)
    assert.Equal(t, 8, treeView.viewportHeight)  // ✅

    // 再次更新行数据
    treeView.UpdateLines(lines)
    assert.Equal(t, 8, treeView.viewportHeight)  // ✅ 仍然为 8！
}
```

**结果**：✅ PASS - `UpdateLines()` 正确保留 `viewportHeight`

### 2. Inspector 约束传播测试

**文件**: `internal/inspector/constraint_propagation_test.go:19-110`

```go
func TestInspectorElementsTabVStackConstraints(t *testing.T) {
    insp := NewStandaloneInspector()
    insp.overlayWidth = 80
    insp.overlayHeight = 25

    elementsContent := insp.buildElementsTabContent()

    // 检查 height prop
    props := elementsContent.Props()
    heightProp, hasHeightProp := props["height"].(int)
    assert.True(t, hasHeightProp)
    assert.Equal(t, 20, heightProp)  // ✅

    // 测量 VStack
    constraints := runtime.BoxConstraints{MaxHeight: 20}
    measurable.Measure(constraints)

    // VStack 应该尊重约束
    assert.Equal(t, 20, size.Height)  // ✅
}
```

**结果**：✅ PASS - VStack 正确设置了 Height(20) prop 并尊重约束

### 3. TreeView 约束测试

**文件**: `components/display/treeview_test.go:11-49`

```go
func TestTreeViewWithBoundedHeight(t *testing.T) {
    treeView := NewTreeView().FromLines(lines).Build()

    constraints := runtime.BoxConstraints{
        MaxHeight: 5,  // 只显示 5 行
    }

    size := treeView.Measure(constraints)

    // TreeView 应该返回约束的高度
    assert.Equal(t, 5, size.Height)  // ✅
}
```

**结果**：✅ PASS - TreeView 正确尊重 MaxHeight 约束

## 修复总结

### 问题
- Inspector 中的 TreeView 每次更新时都创建新实例
- 新实例的 `viewportHeight` 初始化为 0
- 导致虚拟滚动失效，TreeView 渲染所有行

### 解决方案
1. ✅ 添加 `TreeView.UpdateLines()` 方法
2. ✅ 修改 Inspector 使用 `UpdateLines()` 而不是创建新实例
3. ✅ 添加单元测试验证 `UpdateLines()` 保留状态
4. ✅ 添加约束传播测试验证整个流程

### 影响
- ✅ TreeView 现在保持 `viewportHeight` 状态
- ✅ 虚拟滚动正常工作（只渲染可见行）
- ✅ Inspector 性能提升（不需要重新渲染 34 行，只渲染 7 行）
- ✅ 约束正确传播：VStack → TreeView

### 测试结果
- ✅ `TestTreeViewUpdateLinesPreservesViewportHeight` - PASS
- ✅ `TestInspectorElementsTabVStackConstraints` - PASS
- ✅ `TestTreeViewWithBoundedHeight` - PASS
- ✅ 所有现有测试仍然 PASS

## 相关文件

1. `components/display/treeview.go:661-685` - 添加 `UpdateLines()` 方法
2. `internal/inspector/standalone_inspector.go:493-497` - 使用 `UpdateLines()`
3. `components/display/treeview_update_test.go` - 添加 `UpdateLines()` 测试
4. `internal/inspector/constraint_propagation_test.go` - 添加约束传播测试

## 未来改进

1. **TreeViewBuilder 的 `lines` 字段**：当前 `TreeViewBuilder` 没有 `lines` 字段，需要通过 `node.lines` 访问。可以考虑重构 TreeViewBuilder 使其更加清晰。

2. **状态管理**：考虑将 TreeView 的所有状态（viewportHeight, scrollOffset, focusIndex 等）封装到一个 State 对象中，便于管理。

3. **更多测试**：可以添加更多集成测试，模拟完整的 Inspector 渲染流程，确保所有约束都正确传播。
