# Demo启动问题已修复

## 问题
用户反馈："现在示例程序无法启动"

## 根本原因
在ScrollView的auto-height模式下，直接返回TextVNode导致类型不匹配。

## 修复方案
修改`components/layout/scroll_view.go`，在auto-height模式下也返回LayoutNode：

```go
// Auto-height mode
if b.height <= 0 {
    textNode := ui.Text(visibleText)
    textNode.SetProps(ui.Props{"flex": 1, ...})

    // ✅ Wrap in VStackBuilder to maintain LayoutNode type
    result := ui.VStackBuilder(textNode).Width(b.width).Build()

    return result
}
```

## 验证结果

程序现在**正常启动**：

```bash
$ cd examples/ui_demos/demo2_runtime_internals/inspector_overlay
$ go run main.go

UI Inspector ready - Press [I] button or F12/Ctrl+D to toggle
Starting Mint TUI Demo - Press F12 or Ctrl+D to toggle Inspector

[UI界面正常显示]
┌──────────────────────────────────────────────────────────────────────────────┐
│  Runtime Scheduling Pipeline Visualization                                   │
└──────────────────────────────────────────────────────────────────────────────┘
...
System idle, waiting for events...
```

## 如何使用

1. **启动程序**：
   ```bash
   go run main.go
   ```

2. **切换Inspector**：
   - 按 **F12** 键
   - 或按 **Ctrl+D**

3. **查看树结构**：
   - Inspector会显示在右侧
   - Elements标签页显示布局树
   - 使用Flex布局自动扩展

## Flex布局验证

Inspector现在使用Flex布局：
- Header: 固定高度
- SelectedInfo: 固定高度
- **TreeWithStatus: flex: 1** (自动填充剩余空间) ✅
- Instructions: 固定高度

运行测试验证：
```bash
cd internal/inspector
go test -v -run TestInspectorFlexLayout
```

输出：
```
✅ treeWithStatus has flex=1 (should be 1)
✅ Flex layout correctly configured
```

## 文件修改

- `components/layout/scroll_view.go` - auto-height模式返回LayoutNode
- `internal/inspector/standalone_inspector.go` - 使用rtui.Flex包装
