# Mint 布局系统调试工具开发方案

## 目录
- [1. 调试工具概述](#1-调试工具概述)
- [2. 约束追踪器](#2-约束追踪器)
- [3. 布局可视化工具](#3-布局可视化工具)
- [4. 维度检查工具](#4-维度检查工具)
- [5. 性能分析工具](#5-性能分析工具)
- [6. 交互式调试器](#6-交互式调试器)
- [7. 工具集成方案](#7-工具集成方案)

---

## 1. 调试工具概述

### 1.1 设计目标

1. **可视化约束传播**：清晰展示约束如何从父组件传递到子组件
2. **维度转换追踪**：追踪外部/内部维度的转换过程
3. **实时布局树显示**：可视化布局树的结构和状态
4. **性能瓶颈识别**：找出 Measure/Paint 的性能问题
5. **用户友好界面**：提供直观的可视化和交互体验

### 1.2 工具架构

```
┌─────────────────────────────────────────────────────────┐
│                  Debug Tools Layer                      │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  ┌───────────┐  ┌───────────┐  ┌───────────┐           │
│  │ 约束追踪器 │  │ 布局可视化 │  │ 维度检查器 │           │
│  │  Tracer   │  │ Visualizer │  │  Checker  │           │
│  └─────┬─────┘  └─────┬─────┘  └─────┬─────┘           │
│        │              │              │                  │
│        └──────────────┼──────────────┘                  │
│                       │                                 │
│                       ▼                                 │
│              ┌──────────────┐                          │
│              │  Debug Core  │                          │
│              │  (Collector) │                          │
│              └──────┬───────┘                          │
│                     │                                  │
│        ┌────────────┼────────────┐                     │
│        │            │            │                     │
│        ▼            ▼            ▼                     │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐                │
│  │ Console │  │  HTML   │  │  TUI    │                │
│  │ Output  │  │ Report  │  │  UI     │                │
│  └─────────┘  └─────────┘  └─────────┘                │
│                                                         │
└─────────────────────────────────────────────────────────┘
                       │
                       ▼
              ┌──────────────┐
              │ Layout System│
              └──────────────┘
```

### 1.3 用户场景

**场景 1：为什么 Panel 内容溢出了？**

```go
// 用户代码
panel.New().
    SetWidth(20).
    SetHeight(3).
    Content(text.New("Long text").Wrap(true))

// 调试
SetEnv("LAYOUT_DEBUG", "trace")
// 输出：
//  [tracer] root → hstack: {0..50} ✗ {0..100}
//  [tracer] hstack → border(panel): {0..50} ✗ {0..100}
//  [tracer] border → vstack: {18..18} ✗ {0..98}  ← 内部宽度 18
//  [tracer] vstack → text: {18..18} ✗ {0..98}
//  [tracer] text Measure: 4 lines (width=18)
//  [tracer] text Paint: bounds {0,0,20,3} ✗  only 3 rows
//  ⚠️  Text content (4 lines) exceeds paint bounds (3 rows)
```

**场景 2：Auto-height Panel 为什么高度不对？**

```go
// 用户代码
HStack(width=50).
    Children(panel.New().Width(20), panel.New().Width(20))

// 调试
SetEnv("LAYOUT_DEBUG", "visualize")
// 输出可视化树：
//  HStack (50w, 5h)
//  ├─ Panel1 (20w, 3h) ✓
//  └─ Panel2 (20w, 2h)✓
//     └─ Border (18w, 0h) [auto-height]
//        └─ VStack (18w, ?)
//           └─ Text (18w, 1h) ✓
```

---

## 2. 约束追踪器

### 2.1 核心设计

```go
// 文件: ui/layout/constraints/tracer.go

package constraints

import (
    "fmt"
    "os"
    "sync"
    "time"

    "github.com/wwsheng009/mint/runtime/layout"
)

// ============================================================================
// Tracer - 约束传播追踪器
// ============================================================================

type Tracer struct {
    mu       sync.Mutex
    enabled  bool
    entries  []Entry
    tree     *TreeNode
    filters  FilterConfig
    output   OutputConfig
}

type Entry struct {
    Seq        int           // 序列号
    Timestamp  time.Time     // 时间戳
    From       string        // 来源组件
    To         string        // 目标组件
    Path       string        // 完整路径（如 /hstack/panel0/border/vstack/text）
    Input     layout.Constraints // 输入约束
    Output    layout.Constraints // 输出约束
    Dimension layout.Size    // 测量结果
    Reason    string         // 约束修改原因
    Delta     DeltaChanges   // 约束变化量
}

type DeltaChanges struct {
    WidthChanged  bool
    HeightChanged bool
    WidthDelta    int
    HeightDelta   int
    MaxWidthDelta int
    MaxHeightDelta int
}

type TreeNode struct {
    ID         string
    Tag        string
    Constraints layout.Constraints
    Dimension   layout.Size
    Children   []*TreeNode
    Parent     *TreeNode
}

type FilterConfig struct {
    IncludeTags     []string  // 包含的组件标签
    ExcludeTags     []string  // 排除的组件标签
    MinDepth        int       // 最小深度
    MaxDepth        int       // 最大深度
    ShowOnlyChanged bool      // 只显示约束变化的步骤
}

type OutputConfig struct {
    Format       OutputFormat // 输出格式
    ShowTimestamp bool       // 显示时间戳
    ShowDelta     bool       // 显示变化量
    ShowPath      bool       // 显示完整路径
    Compact       bool       // 紧凑模式
}

type OutputFormat int

const (
    FormatConsole OutputFormat = iota
    FormatJSON
    FormatHTML
    FormatMarkdown
)

// ============================================================================
// 全局单例
// ============================================================================

var globalTracer = &Tracer{
    entries: make([]Entry, 0),
    tree:    &TreeNode{ID: "root"},
}

func Enable() {
    globalTracer.mu.Lock()
    defer globalTracer.mu.Unlock()
    globalTracer.enabled = true
    globalTracer.entries = nil
    globalTracer.tree = &TreeNode{ID: "root"}
}

func Disable() {
    globalTracer.mu.Lock()
    defer globalTracer.mu.Unlock()
    globalTracer.enabled = false
}

func IsEnabled() bool {
    globalTracer.mu.Lock()
    defer globalTracer.mu.Unlock()
    return globalTracer.enabled
}

// ============================================================================
// 约束追踪 API
// ============================================================================

// Trace 测量阶段的约束传递
func Trace(from, to, path string, input, output layout.Constraints, resultSize layout.Size, reason string) {
    if !IsEnabled() {
        return
    }

    globalTracer.mu.Lock()
    defer globalTracer.mu.Unlock()

    seq := len(globalTracer.entries)
    delta := computeDelta(input, output)

    entry := Entry{
        Seq:        seq,
        Timestamp:  time.Now(),
        From:       from,
        To:         to,
        Path:       path,
        Input:      input,
        Output:     output,
        Dimension:  resultSize,
        Reason:     reason,
        Delta:      delta,
    }

    // 应用过滤器
    if globalTracer.shouldInclude(entry) {
        globalTracer.entries = append(globalTracer.entries, entry)

        // 更新树
        globalTracer.updateTree(path, output, resultSize)
    }
}

func computeDelta(input, output layout.Constraints) DeltaChanges {
    delta := DeltaChanges{
        WidthChanged:    input.MaxWidth != output.MaxWidth ||
                         input.MinWidth != output.MinWidth,
        HeightChanged:   input.MaxHeight != output.MaxHeight ||
                         input.MinHeight != output.MinHeight,
    }

    if delta.WidthChanged {
        delta.MaxWidthDelta = output.MaxWidth - input.MaxWidth
    }

    if delta.HeightChanged {
        delta.MaxHeightDelta = output.MaxHeight - input.MaxHeight
    }

    return delta
}

func (t *Tracer) shouldInclude(entry Entry) bool {
    // 标签过滤
    if len(t.filters.IncludeTags) > 0 {
        included := false
        for _, tag := range t.filters.IncludeTags {
            if entry.To == tag || entry.From == tag {
                included = true
                break
            }
        }
        if !included {
            return false
        }
    }

    for _, tag := range t.filters.ExcludeTags {
        if entry.To == tag || entry.From == tag {
            return false
        }
    }

    // 变化过滤
    if t.filters.ShowOnlyChanged && !entry.Delta.WidthChanged && !entry.Delta.HeightChanged {
        return false
    }

    return true
}

func (t *Tracer) updateTree(path string, constraints layout.Constraints, size layout.Size) {
    // 解析路径
    parts := splitPath(path)

    // 查找或创建节点
    current := t.tree
    nodePath := "root"

    for i, part := range parts {
        nodePath += "/" + part

        child := t.findChild(current, part)
        if child == nil {
            child = &TreeNode{
                ID:   nodePath,
                Tag:  part,
                Parent: current,
            }
            current.Children = append(current.Children, child)
        }

        // 最后一个节点更新约束和尺寸
        if i == len(parts)-1 {
            child.Constraints = constraints
            child.Dimension = size
        }

        current = child
    }
}

func (t *Tracer) findChild(node *TreeNode, tag string) *TreeNode {
    for _, child := range node.Children {
        if child.Tag == tag {
            return child
        }
    }
    return nil
}

// ============================================================================
// 输出方法
// ============================================================================

func Dump() string {
    globalTracer.mu.Lock()
    defer globalTracer.mu.Unlock()

    switch globalTracer.output.Format {
    case FormatJSON:
        return globalTracer.dumpJSON()
    case FormatHTML:
        return globalTracer.dumpHTML()
    case FormatMarkdown:
        return globalTracer.dumpMarkdown()
    default:
        return globalTracer.dumpConsole()
    }
}

func (t *Tracer) dumpConsole() string {
    var buf strings.Builder
    buf.WriteString("╔══════════════════════════════════════════════════════════════════╗\n")
    buf.WriteString("║                    Constraint Propagation Trace               ║\n")
    buf.WriteString("╚══════════════════════════════════════════════════════════════════╝\n\n")

    for _, entry := range t.entries {
        buf.WriteString(t.formatEntry(entry))
    }

    // 输出树形图
    buf.WriteString("\n╔══════════════════════════════════════════════════════════════════╗\n")
    buf.WriteString("║                          Layout Tree                          ║\n")
    buf.WriteString("╚══════════════════════════════════════════════════════════════════╝\n\n")
    buf.WriteString(t.dumpTree(t.tree, 0))

    return buf.String()
}

func (t *Tracer) formatEntry(entry Entry) string {
    var buf strings.Builder

    // 序号和时间戳
    if t.output.ShowTimestamp {
        buf.WriteString(fmt.Sprintf("Step %d [%s]\n", entry.Seq,
            entry.Timestamp.Format("15:04:05.000")))
    } else {
        buf.WriteString(fmt.Sprintf("Step %d\n", entry.Seq))
    }

    // 路径
    if t.output.ShowPath {
        buf.WriteString(fmt.Sprintf("  Path: %s\n", entry.Path))
    }

    // 约束传播
    buf.WriteString(fmt.Sprintf("  %s → %s\n", entry.From, entry.To))

    // 输入约束
    buf.WriteString(fmt.Sprintf("  Input:    %s\n", formatConstraints(entry.Input)))

    // 箭头（带变化指示）
    arrow := "↓"
    if entry.Delta.WidthChanged || entry.Delta.HeightChanged {
        arrow += "✗"  // 约束被修改
    }
    buf.WriteString(fmt.Sprintf("         %s\n", arrow))

    // 输出约束
    outStr := formatConstraints(entry.Output)
    if entry.Delta.WidthChanged {
        outStr += fmt.Sprintf(" (Δ width=%d)", entry.Delta.MaxWidthDelta)
    }
    if entry.Delta.HeightChanged {
        outStr += fmt.Sprintf(" (Δ height=%d)", entry.Delta.MaxHeightDelta)
    }
    buf.WriteString(fmt.Sprintf("  Output:   %s\n", outStr))

    // 尺寸结果
    if entry.Dimension.Width > 0 || entry.Dimension.Height > 0 {
        buf.WriteString(fmt.Sprintf("  Dimension: %s\n", formatSize(entry.Dimension)))
    }

    // 原因
    if entry.Reason != "" {
        buf.WriteString(fmt.Sprintf("  Reason:   %s\n", entry.Reason))
    }

    // 检测潜在问题
    if entry.Dimension.Height > entry.Input.MaxHeight && entry.Input.MaxHeight > 0 {
        buf.WriteString(fmt.Sprintf("  ⚠️  Measure height (%d) exceeds MaxHeight (%d)\n",
            entry.Dimension.Height, entry.Input.MaxHeight))
    }

    buf.WriteString("\n")
    return buf.String()
}

func (t *Tracer) dumpTree(node *TreeNode, depth int) string {
    var buf strings.Builder

    indent := strings.Repeat("│  ", depth)

    if depth == 0 {
        buf.WriteString("root\n")
    } else {
        isLast := true
        if node.Parent != nil {
            for i, sibling := range node.Parent.Children {
                if sibling == node {
                    isLast = i == len(node.Parent.Children)-1
                    break
                }
            }
        }

        if isLast {
            buf.WriteString(fmt.Sprintf("%s└─ %s\n", indent, node.Tag))
        } else {
            buf.WriteString(fmt.Sprintf("%s├─ %s\n", indent, node.Tag))
        }
    }

    if depth > 0 {
        buf.WriteString(fmt.Sprintf("%s%s  Constraints: %s\n", indent, " ", formatConstraints(node.Constraints)))
        if node.Dimension.Width > 0 || node.Dimension.Height > 0 {
            buf.WriteString(fmt.Sprintf("%s%s  Dimension:  %s\n", indent, " ", formatSize(node.Dimension)))
        }
    }

    for _, child := range node.Children {
        buf.WriteString(t.dumpTree(child, depth+1))
    }

    return buf.String()
}

func formatConstraints(c layout.Constraints) string {
    return fmt.Sprintf("{%d..%d} ✗ {%d..%d}",
        c.MinWidth, c.MaxWidth, c.MinHeight, c.MaxHeight)
}

func formatSize(s layout.Size) string {
    return fmt.Sprintf("%dw x %dh", s.Width, s.Height)
}

func splitPath(path string) []string {
    if path == "" || path == "/" {
        return []string{}
    }
    return strings.Split(strings.Trim(path, "/"), "/")
}

// ============================================================================
// 配置方法
// ============================================================================

func SetFormat(format OutputFormat) {
    globalTracer.mu.Lock()
    defer globalTracer.mu.Unlock()
    globalTracer.output.Format = format
}

func SetFilter(config FilterConfig) {
    globalTracer.mu.Lock()
    defer globalTracer.mu.Unlock()
    globalTracer.filters = config
}

func ShowTimestamps(show bool) {
    globalTracer.mu.Lock()
    defer globalTracer.mu.Unlock()
    globalTracer.output.ShowTimestamp = show
}

func ShowDeltas(show bool) {
    globalTracer.mu.Lock()
    defer globalTracer.mu.Unlock()
    globalTracer.output.ShowDelta = show
}

func CompactMode(compact bool) {
    globalTracer.mu.Lock()
    defer globalTracer.mu.Unlock()
    globalTracer.output.Compact = compact
}

// ============================================================================
// JSON 输出
// ============================================================================

func (t *Tracer) dumpJSON() string {
    type JSONEntry struct {
        Seq       int                    `json:"seq"`
        Timestamp string                 `json:"timestamp"`
        From      string                 `json:"from"`
        To        string                 `json:"to"`
        Path      string                 `json:"path"`
        Input     layout.Constraints     `json:"input"`
        Output    layout.Constraints     `json:"output"`
        Dimension layout.Size            `json:"dimension"`
        Reason    string                 `json:"reason,omitempty"`
        Delta     DeltaChanges           `json:"delta,omitempty"`
    }

    entries := make([]JSONEntry, len(t.entries))
    for i, e := range t.entries {
        entries[i] = JSONEntry{
            Seq:       e.Seq,
            Timestamp: e.Timestamp.Format(time.RFC3339Nano),
            From:      e.From,
            To:        e.To,
            Path:      e.Path,
            Input:     e.Input,
            Output:    e.Output,
            Dimension: e.Dimension,
            Reason:    e.Reason,
            Delta:     e.Delta,
        }
    }

    data, _ := json.MarshalIndent(struct {
        Entries []JSONEntry `json:"entries"`
        Tree    *TreeNode   `json:"tree,omitempty"`
    }{
        Entries: entries,
        Tree:    t.tree,
    }, "", "  ")

    return string(data)
}

// ============================================================================
// 清除
// ============================================================================

func Clear() {
    globalTracer.mu.Lock()
    defer globalTracer.mu.Unlock()
    globalTracer.entries = nil
    globalTracer.tree = &TreeNode{ID: "root"}
}

func GetEntries() []Entry {
    globalTracer.mu.Lock()
    defer globalTracer.mu.Unlock()
    return append([]Entry{}, globalTracer.entries...)
}
```

### 2.2 组件集成

```go
// 使用示例：在组件中集成约束追踪

// ui/components/border/instance.go

import "github.com/wwsheng009/mint/ui/layout/constraints"

func (i *Instance) Measure(constraints layout.Constraints) Size {
    path := i.GetPath()  // 需要实现获取路径的方法
    tag := i.Tag()
    childTag := i.child.Tag()

    // 计算子元素约束
    childConstraints := i.computeChildConstraints(constraints)

    // 追踪约束修改
    constraints.Trace(
        tag, childTag, path,
        constraints, childConstraints,
        layout.Size{},  // 暂时为空，测量后更新
        fmt.Sprintf("Applied border padding (%d)", i.borderPadding),
    )

    // 测量子元素
    childSize := i.child.Measure(childConstraints)

    // 计算外部尺寸
    outerSize := i.computeOuterSize(childSize)

    // 更新追踪结果
    if constraints.IsEnabled() {
        entries := constraints.GetEntries()
        if len(entries) > 0 {
            lastEntry := &entries[len(entries)-1]
            lastEntry.Dimension = outerSize
        }
    }

    return outerSize
}

// Path 管理
type Instance struct {
    // ...
    path string
}

func (i *Instance) SetPath(path string) *Instance {
    i.path = path
    return i
}

func (i *Instance) GetPath() string {
    return i.path
}
```

---

## 3. 布局可视化工具

### 3.1 核心设计

```go
// 文件: ui/layout/visualizer/tree.go

package visualizer

import (
    "fmt"
    "strings"

    "github.com/wwsheng009/mint/runtime/layout"
    "github.com/wwsheng009/mint/runtime/paint"
    rtui "github.com/wwsheng009/mint/runtime/ui"
)

// ============================================================================
// Visualizer - 布局树可视化
// ============================================================================

type Visualizer struct {
    root      rtui.VNode
    states    map[string]*NodeState
    bounds    map[string]paint.Rect
    showConstraints bool
    showDimensions  bool
}

type NodeState struct {
    ID          string
    Tag         string
    Path        string
    Constraints layout.Constraints
    Dimension   layout.Size
    Bounds      paint.Rect
    Parent      string
    Children    []string
    Status      Status
}

type Status int

const (
    StatusOk Status = iota
    StatusWarning
    StatusError
)

func NewVisualizer(root rtui.VNode) *Visualizer {
    return &Visualizer{
        root:        root,
        states:      make(map[string]*NodeState),
        bounds:      make(map[string]paint.Rect),
        showConstraints: true,
        showDimensions:  true,
    }
}

// ============================================================================
// 构建可视化状态
// ============================================================================

func (v *Visualizer) Build(constraints layout.Constraints) error {
    return v.visitNode(v.root, "", "root", constraints)
}

func (v *Visualizer) visitNode(node rtui.VNode, parentPath, nodeID string, constraints layout.Constraints) error {
    path := parentPath + "/" + node.Tag()

    // 创建节点状态
    state := &NodeState{
        ID:          nodeID,
        Tag:         node.Tag(),
        Path:        path,
        Constraints: constraints,
        Parent:      parentPath,
        Status:      StatusOk,
    }

    // 获取或创建实例
    if factory, ok := node.(rtui.InstanceFactory); ok {
        inst := factory.CreateInstance()
        size := inst.Measure(constraints)
        state.Dimension = size

        // 检查问题
        if size.Height > constraints.MaxHeight && constraints.MaxHeight > 0 {
            state.Status = StatusWarning
        }
    }

    v.states[nodeID] = state

    // 递归处理子节点
    children := node.Children()
    state.Children = make([]string, len(children))

    childConstraints := computeChildConstraints(node, constraints)

    for i, child := range children {
        childID := fmt.Sprintf("%s/%d", nodeID, i)
        state.Children[i] = childID

        err := v.visitNode(child, path, childID, childConstraints)
        if err != nil {
            return err
        }
    }

    return nil
}

// ============================================================================
// 输出方法
// ============================================================================

func (v *Visualizer) PrintTree() string {
    var buf strings.Builder
    buf.WriteString(v.renderNode("root", 0))
    return buf.String()
}

func (v *Visualizer) renderNode(nodeID string, depth int) string {
    state := v.states[nodeID]
    if state == nil {
        return ""
    }

    var buf strings.Builder
    indent := strings.Repeat("│  ", depth)

    // 节点头部
    icon := "✓"
    if state.Status == StatusWarning {
        icon = "⚠️"
    } else if state.Status == StatusError {
        icon = "✗"
    }

    buf.WriteString(fmt.Sprintf("%s%s %s", indent, icon, state.Tag))

    // 显示路径
    if depth > 0 {
        buf.WriteString(fmt.Sprintf(" [%s]", state.ID))
    }

    buf.WriteString("\n")

    // 显示约束
    if v.showConstraints && state.Constraints != (layout.Constraints{}) {
        buf.WriteString(fmt.Sprintf("%s│  Constraints: %s\n",
            indent, formatConstraints(state.Constraints)))
    }

    // 显示尺寸
    if v.showDimensions && state.Dimension.Width > 0 || state.Dimension.Height > 0 {
        buf.WriteString(fmt.Sprintf("%s│  Dimension:  %s\n",
            indent, formatSize(state.Dimension)))
    }

    // 显示警告
    if state.Status != StatusOk {
        buf.WriteString(fmt.Sprintf("%s│  ⚠️  %s\n",
            indent, state.getStatusMessage()))
    }

    // 递归渲染子节点
    for i, childID := range state.Children {
        isLast := i == len(state.Children)-1
        childIndent := indent
        if isLast {
            childIndent = childIndent[:len(childIndent)-2] + "└─ "
        } else {
            childIndent = childIndent[:len(childIndent)-2] + "├─ "
        }
        buf.WriteString(v.renderNodeCustom(childID, childIndent))
    }

    return buf.String()
}

func (v *Visualizer) renderNodeCustom(nodeID, prefix string) string {
    state := v.states[nodeID]
    if state == nil {
        return ""
    }

    var buf strings.Builder

    // 节点头部
    icon := "✓"
    if state.Status == StatusWarning {
        icon = "⚠️ "
    } else if state.Status == StatusError {
        icon = "✗ "
    }

    buf.WriteString(fmt.Sprintf("%s%s %s\n", prefix, icon, state.Tag))

    // 显示约束
    if v.showConstraints && state.Constraints != (layout.Constraints{}) {
        buf.WriteString(fmt.Sprintf("%s│  Constraints: %s\n",
            prefix, formatConstraints(state.Constraints)))
    }

    // 显示尺寸
    if v.showDimensions && state.Dimension.Width > 0 || state.Dimension.Height > 0 {
        buf.WriteString(fmt.Sprintf("%s│  Dimension:  %s\n",
            prefix, formatSize(state.Dimension)))
    }

    // 显示警告
    if state.Status != StatusOk {
        buf.WriteString(fmt.Sprintf("%s│  ⚠️  %s\n",
            prefix, state.getStatusMessage()))
    }

    // 递归渲染子节点
    for i, childID := range state.Children {
        isLast := i == len(state.Children)-1
        childPrefix := prefix
        if isLast {
            childPrefix += "    "
        } else {
            childPrefix += "│   "
        }
        buf.WriteString(v.renderNodeCustom(childID, childPrefix+"└─ "))
    }

    return buf.String()
}

func (s *NodeState) getStatusMessage() string {
    if s.Status == StatusWarning {
        return fmt.Sprintf("Dimension %s exceeds constraints %s",
            formatSize(s.Dimension), formatConstraints(s.Constraints))
    }
    if s.Status == StatusError {
        return "Error in layout computation"
    }
    return ""
}

func formatConstraints(c layout.Constraints) string {
    return fmt.Sprintf("{%d..%d} ✗ {%d..%d}",
        c.MinWidth, c.MaxWidth, c.MinHeight, c.MaxHeight)
}

func formatSize(s layout.Size) string {
    return fmt.Sprintf("%dw x %dh", s.Width, s.Height)
}

func computeChildConstraints(node rtui.VNode, constraints layout.Constraints) layout.Constraints {
    // 这是一个简化版本，实际实现需要根据组件类型调整
    return constraints
}

// ============================================================================
// HTML 可视化
// ============================================================================

func (v *Visualizer) PrintHTML() string {
    var buf strings.Builder

    buf.WriteString(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Layout Tree Visualization</title>
    <style>
        body { font-family: 'Courier New', monospace; padding: 20px; background: #1a1a1a; color: #e0e0e0; }
        .tree-node { margin-left: 20px; border-left: 1px solid #444; padding-left: 10px; }
        .node-header { font-weight: bold; }
        .status-ok { color: #4CAF50; }
        .status-warning { color: #FF9800; }
        .status-error { color: #F44336; }
        .constraints { color: #9E9E9E; font-size: 0.9em; }
        .dimension { color: #2196F3; font-size: 0.9em; }
        .warning { background: #FF5722; color: white; padding: 2px 6px; border-radius: 3px; }
    </style>
</head>
<body>
    <h1>Layout Tree Visualization</h1>
`)
    buf.WriteString(v.renderHTMLNode("root"))
    buf.WriteString(`
</body>
</html>`)
    return buf.String()
}

func (v *Visualizer) renderHTMLNode(nodeID string) string {
    state := v.states[nodeID]
    if state == nil {
        return ""
    }

    statusClass := "status-ok"
    if state.Status == StatusWarning {
        statusClass = "status-warning"
    } else if state.Status == StatusError {
        statusClass = "status-error"
    }

    var buf strings.Builder

    // 节点头部
    buf.WriteString(fmt.Sprintf(`<div class="tree-node">
        <div class="node-header">
            <span class="%s">%s</span> %s
        </div>`, statusClass, getIcon(state.Status), state.Tag))

    // 约束
    if v.showConstraints && state.Constraints != (layout.Constraints{}) {
        buf.WriteString(fmt.Sprintf(`<div class="constraints">Constraints: %s</div>`,
            formatConstraintsHTML(state.Constraints)))
    }

    // 尺寸
    if v.showDimensions && state.Dimension.Width > 0 || state.Dimension.Height > 0 {
        buf.WriteString(fmt.Sprintf(`<div class="dimension">Dimension: %s</div>`,
            formatSizeHTML(state.Dimension)))
    }

    // 警告
    if state.Status != StatusOk {
        buf.WriteString(fmt.Sprintf(`<div class="warning">%s</div>`,
            state.getStatusMessage()))
    }

    // 子节点
    for _, childID := range state.Children {
        buf.WriteString(v.renderHTMLNode(childID))
    }

    buf.WriteString(`</div>`)

    return buf.String()
}

func formatConstraintsHTML(c layout.Constraints) string {
    return fmt.Sprintf("{%d..%d} × {%d..%d}",
        c.MinWidth, c.MaxWidth, c.MinHeight, c.MaxHeight)
}

func formatSizeHTML(s layout.Size) string {
    return fmt.Sprintf("%dw × %dh", s.Width, s.Height)
}

func getIcon(status Status) string {
    switch status {
    case StatusOk:
        return "✓"
    case StatusWarning:
        return "⚠️"
    case StatusError:
        return "✗"
    default:
        return "?"
    }
}
```

---

## 4. 维度检查工具

### 4.1 核心设计

```go
// 文件: ui/layout/debug/checker.go

package debug

import (
    "fmt"
    "strings"

    "github.com/wwsheng009/mint/runtime/layout"
    "github.com/wwsheng009/mint/runtime/paint"
    rtui "github.com/wwsheng009/mint/runtime/ui"
)

// ============================================================================
// DimensionChecker - 维度检查器
// ============================================================================

type DimensionChecker struct {
    issues []Issue
}

type Issue struct {
    Severity  Severity
    Component string
    Path      string
    Message   string
    Suggestion string
}

type Severity int

const (
    SeverityInfo Severity = iota
    SeverityWarning
    SeverityError
)

func NewDimensionChecker() *DimensionChecker {
    return &DimensionChecker{
        issues: make([]Issue, 0),
    }
}

// ============================================================================
// 检查方法
// ============================================================================

func (dc *DimensionChecker) CheckNode(node rtui.VNode, path string, constraints layout.Constraints) {
    // 检查约束有效性
    dc.checkConstraints(node, path, constraints)

    // 检查尺寸
    factory, ok := node.(rtui.InstanceFactory)
    if ok {
        size := factory.CreateInstance().Measure(constraints)
        dc.checkDimension(node, path, constraints, size)
    }

    // 递归检查子节点
    children := node.Children()
    childConstraints := computeChildConstraints(node, constraints)

    for i, child := range children {
        childPath := fmt.Sprintf("%s/%s", path, child.Tag())
        dc.CheckNode(child, childPath, childConstraints)
    }
}

func (dc *DimensionChecker) checkConstraints(node rtui.VNode, path string, constraints layout.Constraints) {
    // 检查 Min <= Max
    if constraints.MinWidth > constraints.MaxWidth && constraints.MaxWidth > 0 {
        dc.issues = append(dc.issues, Issue{
            Severity:  SeverityError,
            Component: node.Tag(),
            Path:      path,
            Message:   fmt.Sprintf("MinWidth (%d) > MaxWidth (%d)",
                constraints.MinWidth, constraints.MaxWidth),
            Suggestion: "Ensure MinWidth <= MaxWidth",
        })
    }

    if constraints.MinHeight > constraints.MaxHeight && constraints.MaxHeight > 0 {
        dc.issues = append(dc.issues, Issue{
            Severity:  SeverityError,
            Component: node.Tag(),
            Path:      path,
            Message:   fmt.Sprintf("MinHeight (%d) > MaxHeight (%d)",
                constraints.MinHeight, constraints.MaxHeight),
            Suggestion: "Ensure MinHeight <= MaxHeight",
        })
    }

    // 检查负值
    if constraints.MinWidth < 0 {
        dc.issues = append(dc.issues, Issue{
            Severity:  SeverityWarning,
            Component: node.Tag(),
            Path:      path,
            Message:   fmt.Sprintf("Negative MinWidth: %d", constraints.MinWidth),
            Suggestion: "Use non-negative values",
        })
    }

    if constraints.MinHeight < 0 {
        dc.issues = append(dc.issues, Issue{
            Severity:  SeverityWarning,
            Component: node.Tag(),
            Path:      path,
            Message:   fmt.Sprintf("Negative MinHeight: %d", constraints.MinHeight),
            Suggestion: "Use non-negative values",
        })
    }
}

func (dc *DimensionChecker) checkDimension(node rtui.VNode, path string, constraints layout.Constraints, size layout.Size) {
    // 检查尺寸在约束范围内
    if size.Width < constraints.MinWidth {
        dc.issues = append(dc.issues, Issue{
            Severity:  SeverityWarning,
            Component: node.Tag(),
            Path:      path,
            Message:   fmt.Sprintf("Width (%d) < MinWidth (%d)",
                size.Width, constraints.MinWidth),
            Suggestion: "Consider adjusting MinWidth or content",
        })
    }

    if size.Width > constraints.MaxWidth && constraints.MaxWidth > 0 {
        dc.issues = append(dc.issues, Issue{
            Severity:  SeverityError,
            Component: node.Tag(),
            Path:      path,
            Message:   fmt.Sprintf("Width (%d) > MaxWidth (%d)",
                size.Width, constraints.MaxWidth),
            Suggestion: "Content exceeds available width",
        })
    }

    if size.Height < constraints.MinHeight {
        dc.issues = append(dc.issues, Issue{
            Severity:  SeverityWarning,
            Component: node.Tag(),
            Path:      path,
            Message:   fmt.Sprintf("Height (%d) < MinHeight (%d)",
                size.Height, constraints.MinHeight),
            Suggestion: "Consider adjusting MinHeight",
        })
    }

    if size.Height > constraints.MaxHeight && constraints.MaxHeight > 0 {
        dc.issues = append(dc.issues, Issue{
            Severity:  SeverityWarning,
            Component: node.Tag(),
            Path:      path,
            Message:   fmt.Sprintf("Height (%d) > MaxHeight (%d), content will be cropped",
                size.Height, constraints.MaxHeight),
            Suggestion: "Increase height or enable content cropping",
        })
    }

    // 检查零尺寸
    if size.Width == 0 && constraints.MinWidth > 0 {
        dc.issues = append(dc.issues, Issue{
            Severity:  SeverityInfo,
            Component: node.Tag(),
            Path:      path,
            Message:   "Width measured as 0",
            Suggestion: "Check if there are layout constraints preventing sizing",
        })
    }

    if size.Height == 0 && constraints.MinHeight > 0 {
        dc.issues = append(dc.issues, Issue{
            Severity:  SeverityInfo,
            Component: node.Tag(),
            Path:      path,
            Message:   "Height measured as 0",
            Suggestion: "Check if there are layout constraints preventing sizing",
        })
    }
}

// ============================================================================
// 输出方法
// ============================================================================

func (dc *DimensionChecker) PrintReport() string {
    var buf strings.Builder

    if len(dc.issues) == 0 {
        buf.WriteString("✓ No dimension issues found\n")
        return buf.String()
    }

    buf.WriteString(fmt.Sprintf("Found %d issue(s):\n\n", len(dc.issues)))

    errorCount := 0
    warningCount := 0
    infoCount := 0

    for _, issue := range dc.issues {
        buf.WriteString(dc.formatIssue(issue))
        buf.WriteString("\n")

        switch issue.Severity {
        case SeverityError:
            errorCount++
        case SeverityWarning:
            warningCount++
        case SeverityInfo:
            infoCount++
        }
    }

    // 摘要
    buf.WriteString("────────────────────────────────────────\n")
    buf.WriteString(fmt.Sprintf("Summary: %d error(s), %d warning(s), %d info\n",
        errorCount, warningCount, infoCount))

    return buf.String()
}

func (dc *DimensionChecker) formatIssue(issue Issue) string {
    icon := "ℹ️ "
    prefix := "  "

    switch issue.Severity {
    case SeverityWarning:
        icon = "⚠️ "
    case SeverityError:
        icon = "✗ "
    }

    var buf strings.Builder
    buf.WriteString(fmt.Sprintf("%s%s[%s] %s\n", prefix, icon, issue.Component, issue.Path))
    buf.WriteString(fmt.Sprintf("%s│  %s\n", prefix, issue.Message))
    if issue.Suggestion != "" {
        buf.WriteString(fmt.Sprintf("%s│  💡 %s\n", prefix, issue.Suggestion))
    }

    return buf.String()
}

func (dc *DimensionChecker) GetIssues() []Issue {
    return dc.issues
}

func (dc *DimensionChecker) GetErrors() []Issue {
    var errors []Issue
    for _, issue := range dc.issues {
        if issue.Severity == SeverityError {
            errors = append(errors, issue)
        }
    }
    return errors
}

func (dc *DimensionChecker) GetWarnings() []Issue {
    var warnings []Issue
    for _, issue := range dc.issues {
        if issue.Severity == SeverityWarning {
            warnings = append(warnings, issue)
        }
    }
    return warnings
}

func (dc *DimensionChecker) HasErrors() bool {
    for _, issue := range dc.issues {
        if issue.Severity == SeverityError {
            return true
        }
    }
    return false
}

// ============================================================================
// Panel 专用检查
// ============================================================================

func (dc *DimensionChecker) CheckPanel(panel rtui.VNode) {
    // 检查 Panel 的外部/内部维度一致性
    // 确保外部维度 = 内部维度 + 边框 padding
}

func computeChildConstraints(node rtui.VNode, constraints layout.Constraints) layout.Constraints {
    return constraints
}
```

---

## 5. 性能分析工具

### 5.1 核心设计

```go
// 文件: ui/layout/performance/analyzer.go

package performance

import (
    "fmt"
    "sort"
    "time"

    "github.com/wwsheng009/mint/runtime/layout"
    "github.com/wwsheng009/mint/runtime/paint"
    rtui "github.com/wwsheng009/mint/runtime/ui"
)

// ============================================================================
// PerformanceAnalyzer - 性能分析器
// ============================================================================

type PerformanceAnalyzer struct {
    measurements map[string]*ComponentStats
    enabled      bool
}

type ComponentStats struct {
    Tag            string
    MeasureCount   int
    MeasureTotal   time.Duration
    MeasureMin     time.Duration
    MeasureMax     time.Duration
    PaintCount     int
    PaintTotal     time.Duration
    PaintMin       time.Duration
    PaintMax       time.Duration
    CacheHitCount  int
    CacheMissCount int
}

func NewPerformanceAnalyzer() *PerformanceAnalyzer {
    return &PerformanceAnalyzer{
        measurements: make(map[string]*ComponentStats),
    }
}

func (pa *PerformanceAnalyzer) Enable() {
    pa.enabled = true
}

func (pa *PerformanceAnalyzer) Disable() {
    pa.enabled = false
}

// ============================================================================
// 性能追踪 API
// ============================================================================

func (pa *PerformanceAnalyzer) StartMeasure(tag string) func() {
    if !pa.enabled {
        return func() {}
    }

    start := time.Now()

    return func() {
        duration := time.Since(start)
        pa.recordMeasure(tag, duration)
    }
}

func (pa *PerformanceAnalyzer) StartPaint(tag string) func() {
    if !pa.enabled {
        return func() {}
    }

    start := time.Now()

    return func() {
        duration := time.Since(start)
        pa.recordPaint(tag, duration)
    }
}

func (pa *PerformanceAnalyzer) RecordCacheHit(tag string) {
    if !pa.enabled {
        return
    }

    stats := pa.getOrCreateStats(tag)
    stats.CacheHitCount++
}

func (pa *PerformanceAnalyzer) RecordCacheMiss(tag string) {
    if !pa.enabled {
        return
    }

    stats := pa.getOrCreateStats(tag)
    stats.CacheMissCount++
}

// ============================================================================
// 记录方法
// ============================================================================

func (pa *PerformanceAnalyzer) recordMeasure(tag string, duration time.Duration) {
    stats := pa.getOrCreateStats(tag)
    stats.MeasureCount++
    stats.MeasureTotal += duration

    if stats.MeasureMin == 0 || duration < stats.MeasureMin {
        stats.MeasureMin = duration
    }
    if duration > stats.MeasureMax {
        stats.MeasureMax = duration
    }
}

func (pa *PerformanceAnalyzer) recordPaint(tag string, duration time.Duration) {
    stats := pa.getOrCreateStats(tag)
    stats.PaintCount++
    stats.PaintTotal += duration

    if stats.PaintMin == 0 || duration < stats.PaintMin {
        stats.PaintMin = duration
    }
    if duration > stats.PaintMax {
        stats.PaintMax = duration
    }
}

func (pa *PerformanceAnalyzer) getOrCreateStats(tag string) *ComponentStats {
    stats, ok := pa.measurements[tag]
    if !ok {
        stats = &ComponentStats{Tag: tag}
        pa.measurements[tag] = stats
    }
    return stats
}

// ============================================================================
// 输出方法
// ============================================================================

func (pa *PerformanceAnalyzer) PrintReport() string {
    var buf strings.Builder

    buf.WriteString("╔══════════════════════════════════════════════════════════════════╗\n")
    buf.WriteString("║                    Performance Report                          ║\n")
    buf.WriteString("╚══════════════════════════════════════════════════════════════════╝\n\n")

    // 按总时间排序
    var sortedStats []*ComponentStats
    for _, stats := range pa.measurements {
        sortedStats = append(sortedStats, stats)
    }

    sort.Slice(sortedStats, func(i, j int) bool {
        return sortedStats[i].MeasureTotal > sortedStats[j].MeasureTotal
    })

    // 打印表头
    buf.WriteString("┌─ Component ──┬─ Measure Count ──┬─ Total ──┬─ Avg ──┬─ Min ──┬─ Max ──┐\n")

    // 打印每个组件的统计
    for _, stats := range sortedStats {
        if stats.MeasureCount == 0 {
            continue
        }

        avg := stats.MeasureTotal / time.Duration(stats.MeasureCount)

        buf.WriteString(fmt.Sprintf("│ %-12s │ %14d │ %7s │ %5s │ %5s │ %5s │\n",
            stats.Tag,
            stats.MeasureCount,
            formatDuration(stats.MeasureTotal),
            formatDuration(avg),
            formatDuration(stats.MeasureMin),
            formatDuration(stats.MeasureMax)))
    }

    buf.WriteString("└──────────────┴────────────────┴─────────┴───────┴───────┴───────┘\n\n")

    // Paint 统计
    buf.WriteString("┌─ Component ──┬─ Paint Count ───┬─ Total ──┬─ Avg ──┬─ Min ──┬─ Max ──┐\n")

    for _, stats := range sortedStats {
        if stats.PaintCount == 0 {
            continue
        }

        avg := stats.PaintTotal / time.Duration(stats.PaintCount)

        buf.WriteString(fmt.Sprintf("│ %-12s │ %13d │ %7s │ %5s │ %5s │ %5s │\n",
            stats.Tag,
            stats.PaintCount,
            formatDuration(stats.PaintTotal),
            formatDuration(avg),
            formatDuration(stats.PaintMin),
            formatDuration(stats.PaintMax)))
    }

    buf.WriteString("└──────────────┴────────────────────────┴─────────┴───────┴───────┴───────┘\n\n")

    // 缓存统计
    buf.WriteString("┌─ Component ──┬─ Cache Hits ────┬─ Cache Misses ──┬─ Hit Rate ────┐\n")

    for _, stats := range sortedStats {
        total := stats.CacheHitCount + stats.CacheMissCount
        if total == 0 {
            continue
        }

        hitRate := float64(stats.CacheHitCount) / float64(total) * 100

        buf.WriteString(fmt.Sprintf("│ %-12s │ %13d │ %13d │ %10.1f%% │\n",
            stats.Tag,
            stats.CacheHitCount,
            stats.CacheMissCount,
            hitRate))
    }

    buf.WriteString("└──────────────┴────────────────┴──────────────────────────────────────┘\n")

    return buf.String()
}

func formatDuration(d time.Duration) string {
    if d < time.Microsecond {
        return fmt.Sprintf("%dns", d.Nanoseconds())
    }
    if d < time.Millisecond {
        return fmt.Sprintf("%.1fµs", float64(d.Microseconds()))
    }
    return fmt.Sprintf("%.1fms", float64(d.Milliseconds()))
}

func (pa *PerformanceAnalyzer) Reset() {
    pa.measurements = make(map[string]*ComponentStats)
}

func (pa *PerformanceAnalyzer) GetStats(tag string) *ComponentStats {
    return pa.measurements[tag]
}
```

---

## 6. 交互式调试器

### 6.1 TUI 调试器设计

```go
// 文件: tools/debugger/main.go

package main

import (
    "fmt"
    "os"

    "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    "github.com/wwsheng009/mint/ui/layout"
    "github.com/wwsheng009/mint/ui/components/panel"
    newstack "github.com/wwsheng009/mint/ui/components/stack"
    newtext "github.com/wwsheng009/mint/ui/components/text"
)

// ============================================================================
// TUI 调试器
// ============================================================================

type model struct {
    treeView      string
    constraintsView string
    selectedNode  string
    dimensions    map[string]string
    issues        []string
    helpVisible   bool
}

type tickMsg time.Time

func initialModel() model {
    // 构建测试布局
    root := createTestLayout()
    visualizer := layout.NewVisualizer(root)
    visualizer.Build(layout.Constraints{
        MaxWidth:  100,
        MaxHeight: 100,
    })

    // 检查问题
    checker := layout.NewDimensionChecker()
    checker.CheckNode(root, "", layout.Constraints{
        MaxWidth:  100,
        MaxHeight: 100,
    })

    issues := make([]string, 0)
    if checker.HasErrors() {
        for _, issue := range checker.GetErrors() {
            issues = append(issues, issue.Message)
        }
    }

    return model{
        treeView:       visualizer.PrintTree(),
        constraintsView: layout.constraints.Dump(),
        dimensions:     make(map[string]string),
        issues:         issues,
        helpVisible:    false,
    }
}

func createTestLayout() rtui.VNode {
    panel1 := panel.NewBuilder().
        Width(20).
        Height(5).
        Title("Panel 1").
        Content(newtext.New("Fixed content")).
        Build()

    panel2 := panel.NewBuilder().
        Width(20).
        Title("Panel 2").
        Content(newtext.New("Auto content that wraps to multiple lines").Wrap(true)).
        Build()

    return newstack.New(newstack.Row).
        Width(100).
        Gap(2).
        Children(panel1, panel2)
}

func (m model) Init() tea.Cmd {
    return tea.Tick(time.Second, func(t time.Time) tea.Msg {
        return tickMsg(t)
    })
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "q":
            return m, tea.Quit

        case "h":
            m.helpVisible = !m.helpVisible

        case "up", "k":
            // 向上选择节点
            m.selectPreviousNode()

        case "down", "j":
            // 向下选择节点
            m.selectNextNode()

        case "d":
            // 显示维度详情
            m.dimensions = m.getNodeDimensions(m.selectedNode)
        }
    }

    return m, nil
}

func (m model) View() string {
    var buf strings.Builder

    // 标题
    titleStyle := lipgloss.NewStyle().
        Foreground(lipgloss.Color("#FAFAFA")).
        Background(lipgloss.Color("#7D56F4")).
        Padding(0, 2)

    buf.WriteString(titleStyle.Render("Mint Layout Debugger"))
    buf.WriteString("\n\n")

    // 使用列布局
    mainView := lipgloss.JoinHorizontal(
        lipgloss.Top,
        m.leftPanel(),
        m.rightPanel(),
    )

    buf.WriteString(mainView)

    // 帮助面板
    if m.helpVisible {
        buf.WriteString("\n\n")
        buf.WriteString(m.helpPanel())
    }

    return buf.String()
}

func (m model) leftPanel() string {
    treeStyle := lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("#6262ED")).
        Padding(1)

    treeTitle := lipgloss.NewStyle().
        Foreground(lipgloss.Color("#A6ACCD")).
        Bold(true).
        Render("Layout Tree")

    content := lipgloss.JoinVertical(
        lipgloss.Left,
        treeTitle,
        "",
        m.treeView,
    )

    return treeStyle.Render(content)
}

func (m model) rightPanel() string {
    style := lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("#6262ED")).
        Padding(1)

    var content []string

    // 选中节点信息
    if m.selectedNode != "" {
        content = append(content, m.nodeInfoPanel())
        content = append(content, "")
    }

    // 约束追踪
    constraintsTitle := lipgloss.NewStyle().
        Foreground(lipgloss.Color("#A6ACCD")).
        Bold(true).
        Render("Constraint Propagation")

    content = append(content, constraintsTitle)
    content = append(content, "")
    content = append(content, m.constraintsView)

    // 问题列表
    if len(m.issues) > 0 {
        content = append(content, "")
        issuesTitle := lipgloss.NewStyle().
            Foreground(lipgloss.Color("#FF6B6B")).
            Bold(true).
            Render("Issues")

        content = append(content, issuesTitle)
        content = append(content, "")
        for _, issue := range m.issues {
            content = append(content, "• "+issue)
        }
    }

    return style.Render(lipgloss.JoinVertical(lipgloss.Left, content...))
}

func (m model) nodeInfoPanel() string {
    title := lipgloss.NewStyle().
        Foreground(lipgloss.Color("#4ADE80")).
        Bold(true).
        Render(fmt.Sprintf("Selected: %s", m.selectedNode))

    return lipgloss.NewStyle().
        Foreground(lipgloss.Color("#E0E0E0")).
        Render(title)
}

func (m model) helpPanel() string {
    helpStyle := lipgloss.NewStyle().
        Border(lipgloss.DoubleBorder()).
        BorderForeground(lipgloss.Color("#5C5C70")).
        Padding(1)

    content := fmt.Sprintf(
        "Keyboard Shortcuts:\n"+
            "  q / Ctrl+C  Quit\n"+
            "  ↑ / k        Select previous node\n"+
            "  ↓ / j        Select next node\n"+
            "  d            Show dimension details\n"+
            "  h            Toggle this help panel",
    )

    return helpStyle.Render(content)
}

// ============================================================================
// Main
// ============================================================================

func main() {
    // 运行可视化布局Demo
    RunInteractiveDebugger()
}

func RunInteractiveDebugger() {
    p := tea.NewProgram(initialModel())
    if _, err := p.Run(); err != nil {
        fmt.Printf("Alas, there's been an error: %v", err)
        os.Exit(1)
    }
}
```

---

## 7. 工具集成方案

### 7.1 环境变量配置

```go
// 文件: ui/layout/debug/config.go

package debug

import (
    "os"
    "strconv"
    "strings"
)

const (
    EnvDebugMode       = "LAYOUT_DEBUG"
    EnvTraceMode       = "LAYOUT_TRACE"
    EnvVisualizeMode   = "LAYOUT_VISUALIZE"
    EnvCheckMode       = "LAYOUT_CHECK"
    EnvProfileMode     = "LAYOUT_PROFILE"
)

// DebugMode 调试模式类型
type DebugMode int

const (
    DebugModeOff DebugMode = iota
    DebugModeTrace
    DebugModeVisualize
    DebugModeCheck
    DebugModeProfile
)

type Config struct {
    Mode        DebugMode
    TraceOnly   []string
    Console     bool
    File        bool
    FilePath    string
    Format      OutputFormat
}

func LoadConfig() *Config {
    config := &Config{
        Mode:        detectMode(),
        TraceOnly:   getTraceOnly(),
        Console:     getBoolEnv("LAYOUT_DEBUG_CONSOLE", true),
        File:        getBoolEnv("LAYOUT_DEBUG_FILE", false),
        FilePath:    getStringEnv("LAYOUT_DEBUG_FILE_PATH", "layout_debug.log"),
        Format:      parseFormat(getStringEnv("LAYOUT_DEBUG_FORMAT", "console")),
    }

    return config
}

func detectMode() DebugMode {
    debugMode := os.Getenv(EnvDebugMode)
    traceMode := os.Getenv(EnvTraceMode)
    visualizeMode := os.Getenv(EnvVisualizeMode)
    checkMode := os.Getenv(EnvCheckMode)
    profileMode := os.Getenv(EnvProfileMode)

    if profileMode != "" {
        return DebugModeProfile
    }
    if checkMode != "" {
        return DebugModeCheck
    }
    if visualizeMode != "" {
        return DebugModeVisualize
    }
    if traceMode != "" {
        return DebugModeTrace
    }
    if debugMode != "" {
        return DebugModeTrace  // 默认是 trace
    }

    return DebugModeOff
}

func getTraceOnly() []string {
    value := os.Getenv("LAYOUT_TRACE_ONLY")
    if value == "" {
        return nil
    }
    return strings.Split(value, ",")
}

func getBoolEnv(key string, defaultValue bool) bool {
    value := os.Getenv(key)
    if value == "" {
        return defaultValue
    }
    return strings.ToLower(value) == "true" || value == "1"
}

func getStringEnv(key, defaultValue string) string {
    value := os.Getenv(key)
    if value == "" {
        return defaultValue
    }
    return value
}

func parseFormat(value string) OutputFormat {
    switch strings.ToLower(value) {
    case "json":
        return FormatJSON
    case "html":
        return FormatHTML
    case "md", "markdown":
        return FormatMarkdown
    default:
        return FormatConsole
    }
}
```

### 7.2 命令行工具

```go
// 文件: cmd/layout-debug/main.go

package main

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/spf13/cobra"
    "github.com/wwsheng009/mint/ui/layout"
    "github.com/wwsheng009/mint/ui/layout/constraints"
    "github.com/wwsheng009/mint/ui/layout/visualizer"
    "github.com/wwsheng009/mint/ui/layout/debug"
)

func main() {
    var rootCmd = &cobra.Command{
        Use:   "layout-debug",
        Short: "Mint layout debugging tools",
    }

    // trace 命令
    traceCmd := &cobra.Command{
        Use:   "trace [demo-name]",
        Short: "Trace constraint propagation",
        Args:  cobra.ExactArgs(1),
        Run:   runTrace,
    }

    // visualize 命令
    visualizeCmd := &cobra.Command{
        Use:   "visualize [demo-name]",
        Short: "Visualize layout tree",
        Args:  cobra.ExactArgs(1),
        Run:   runVisualize,
    }

    // check 命令
    checkCmd := &cobra.Command{
        Use:   "check [demo-name]",
        Short: "Check for dimension issues",
        Args:  cobra.ExactArgs(1),
        Run:   runCheck,
    }

    // profile 命令
    profileCmd := &cobra.Command{
        Use:   "profile [demo-name]",
        Short: "Profile layout performance",
        Args:  cobra.ExactArgs(1),
        Run:   runProfile,
    }

    rootCmd.AddCommand(traceCmd, visualizeCmd, checkCmd, profileCmd)

    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}

func runTrace(cmd *cobra.Command, args []string) {
    demoName := args[0]

    // 启用追踪
    os.Setenv("LAYOUT_TRACE", "true")

    // 加载并运行 demo
    root := loadDemo(demoName)

    // 渲染并追踪
    constraints.Enable()
    root.Measure(layout.Constraints{
        MaxWidth:  100,
        MaxHeight: 100,
    })

    // 输出追踪结果
    fmt.Println(constraints.Dump())

    // 保存到文件
    if shouldSaveToFile() {
        outputFile := filepath.Join("debug", fmt.Sprintf("%s_trace.txt", demoName))
        saveToFile(outputFile, constraints.Dump())
    }
}

func runVisualize(cmd *cobra.Command, args []string) {
    demoName := args[0]

    root := loadDemo(demoName)

    viz := visualizer.NewVisualizer(root)
    viz.Build(layout.Constraints{
        MaxWidth:  100,
        MaxHeight: 100,
    })

    output := viz.PrintHTML()

    // 保存 HTML
    outputFile := filepath.Join("debug", fmt.Sprintf("%s_visualize.html", demoName))
    saveToFile(outputFile, output)

    fmt.Printf("Visualization saved to %s\n", outputFile)
}

func runCheck(cmd *cobra.Command, args []string) {
    demoName := args[0]

    root := loadDemo(demoName)

    checker := debug.NewDimensionChecker()
    checker.CheckNode(root, "", layout.Constraints{
        MaxWidth:  100,
        MaxHeight: 100,
    })

    report := checker.PrintReport()
    fmt.Println(report)

    if checker.HasErrors() {
        os.Exit(1)
    }
}

func runProfile(cmd *cobra.Command, args []string) {
    demoName := args[0]

    root := loadDemo(demoName)

    analyzer := performance.NewPerformanceAnalyzer()
    analyzer.Enable()

    // 多次运行测量
    for i := 0; i < 100; i++ {
        done := analyzer.StartMeasure("root")

        root.Measure(layout.Constraints{
            MaxWidth:  100,
            MaxHeight: 100,
        })

        done()
    }

    report := analyzer.PrintReport()
    fmt.Println(report)
}

// 辅助函数
func loadDemo(name string) rtui.VNode {
    // 加载 demo
    switch name {
    case "panel":
        return demos.PanelDemo()
    case "hstack":
        return demos.HStackDemo()
    // ...
    default:
        panic(fmt.Sprintf("Unknown demo: %s", name))
    }
}

func shouldSaveToFile() bool {
    return os.Getenv("LAYOUT_DEBUG_FILE") == "true"
}

func saveToFile(path, content string) {
    dir := filepath.Dir(path)
    os.MkdirAll(dir, 0755)
    os.WriteFile(path, []byte(content), 0644)
}
```

---

**文档版本**: 1.0
**最后更新**: 2026-02-21
**作者**: Qwen Code
