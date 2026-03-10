# 样式状态修复与 Component ID 系统优化

**日期**: 2026-01-31  
**状态**: ✅ 已完成  
**相关文档**: [session-ses_3ec9.md](../../session-ses_3ec9.md)

---

## 问题背景

根据 `session-ses_3ec9.md` 文档记录的 Bug 修复过程，Mint TUI 框架存在以下问题：

1. **样式属性无法关闭** - 从有到无的样式变化无法正确渲染
2. **焦点管理系统不稳定** - 依赖指针比较，组件重建后焦点匹配失败
3. **Diff 渲染算法缺陷** - 宽字符处理、内存泄漏等问题

---

## 修复内容

### 1. 样式属性关闭检测 (style_state.go)

**问题**: 原 `buildDiffCodes()` 只处理属性**开启**（从无到有），无法处理**关闭**（从有到无）

**解决方案**: 添加属性关闭检测逻辑

```go
// Check if we need to turn OFF any attributes - this requires a reset
needsReset := false
if from.IsBold() && !to.IsBold() {
    needsReset = true
}
if from.IsItalic() && !to.IsItalic() {
    needsReset = true
}
// ... 其他属性检测

if from.BG != "" && to.BG == "" {
    needsReset = true
}
if from.FG != "" && to.FG == "" {
    needsReset = true
}

// If we need to reset attributes, reset and apply full new style
if needsReset {
    return "\x1b[0m" + s.fullStyle(to)
}
```

**文件修改**:
- `runtime/paint/style_state.go` (第57-87行)

---

### 2. Component ID 系统优化 (UI层)

**问题**: 焦点管理依赖内存指针比较，组件销毁重建后指针变化导致焦点丢失

**解决方案**: 引入稳定的 `focusIndex` 系统

#### 2.1 为所有交互组件添加 focusIndex 字段

```go
// ButtonVNode
type ButtonVNode struct {
    *ElementVNode
    // ... 其他字段
    focusIndex int // Index for focus management, set during collection
}

// InputVNode
type InputVNode struct {
    *ElementVNode
    // ... 其他字段
    focusIndex int
}

// CheckboxVNode
type CheckboxVNode struct {
    *ElementVNode
    // ... 其他字段
    focusIndex int
}

// SelectVNode
type SelectVNode struct {
    *ElementVNode
    // ... 其他字段
    focusIndex int
}

// TextareaVNode
type TextareaVNode struct {
    *ElementVNode
    // ... 其他字段
    focusIndex int
}
```

**文件修改**:
- `ui/button.go`
- `ui/input.go`
- `ui/checkbox.go`
- `ui/select.go`

#### 2.2 重构焦点判断逻辑

**之前** (指针比较):
```go
localInputIndex := d.focusedIndex - len(d.buttons)
isFocused := d.focusedType == 1 && 
    localInputIndex >= 0 && 
    localInputIndex < len(d.inputs) && 
    d.inputs[localInputIndex] == node  // ⚠️ 指针比较
```

**之后** (索引比较):
```go
isFocused := d.focusedType == 1 && 
    node.focusIndex >= 0 && 
    node.focusIndex == d.focusedIndex  // ✅ 稳定索引
```

#### 2.3 修复递归收集 Bug

**问题**: `collectInteractiveElements` 在递归时会重置集合

**解决方案**: 分离重置和收集逻辑

```go
// 调用前重置
d.resetInteractiveElements()

// 然后收集 (不再内部重置)
d.collectInteractiveElements(vnode)
```

**新增方法**:
```go
func (d *declarativeRoot) resetInteractiveElements() {
    d.buttons = d.buttons[:0]
    d.inputs = d.inputs[:0]
    d.textareas = d.textareas[:0]
    d.checkboxes = d.checkboxes[:0]
    d.selects = d.selects[:0]
}
```

**文件修改**:
- `ui/app.go`
- `ui/component_test.go`

#### 2.4 清理冗余代码

移除不再使用的渲染索引字段:
```go
// 已删除字段:
renderBtnIndex      int
renderInputIndex    int
renderTextareaIndex int
renderCheckboxIndex int
renderSelectIndex   int
```

---

### 3. Diff 渲染算法修复 (已完成)

根据 `runtime/paint/BUGFIX_diff_rendering.md`:

| 问题 | 状态 | 修复内容 |
|------|------|---------|
| 无限循环 | ✅ | Width=0时确保x至少前进1 |
| swapBuffers逻辑错误 | ✅ | 交换后添加复制逻辑 |
| 内存泄漏 | ✅ | 复用visited数组 |
| IsCellChanged不一致 | ✅ | 统一使用IsCellChanged |

---

## 测试结果

### UI 包测试
```
=== RUN   TestCounterComponent
--- PASS: TestCounterComponent (0.00s)
=== RUN   TestButtonInteraction
--- PASS: TestButtonInteraction (0.00s)
=== RUN   TestButtonInteractionByLabel
--- PASS: TestButtonInteractionByLabel (0.00s)
=== RUN   TestMultiModalFlow
--- PASS: TestMultiModalFlow (0.00s)

PASS
ok      github.com/wwsheng009/mint/ui   0.149s
```

### Runtime/Paint 包测试
```
=== RUN   TestRenderer_DiffRendering
--- PASS: TestRenderer_DiffRendering (0.00s)
=== RUN   TestRenderer_WideCharStyleChange
--- PASS: TestRenderer_WideCharStyleChange (0.00s)
=== RUN   TestWideCharIntegration
--- PASS: TestWideCharIntegration (0.00s)

PASS
ok      github.com/wwsheng009/mint/runtime/paint  0.105s
```

### Framework/Component 包测试
```
PASS
ok      github.com/wwsheng009/mint/framework/component  0.108s
```

---

## 影响分析

### 正向影响
1. ✅ 样式属性切换更加可靠
2. ✅ 焦点管理不再依赖不稳定指针
3. ✅ 组件重建后焦点状态正确恢复
4. ✅ 代码更加清晰简洁

### 风险评估
- **低风险**: 纯内部实现改进，API 保持不变
- **测试覆盖**: 所有相关测试用例已通过
- **向后兼容**: 不影响现有应用代码

---

## 后续建议

1. **监控生产环境**: 观察 `TUI_DEBUG_WATCHDOG` 日志（如启用）
2. **扩展测试**: 添加更多边界条件测试
3. **文档更新**: 更新开发者文档说明焦点管理机制
4. **性能优化**: 考虑进一步优化渲染管线

---

## 相关文件变更

```
ui/
├── app.go              (重构焦点管理逻辑)
├── button.go           (添加 focusIndex 字段)
├── input.go            (添加 focusIndex 字段)
├── checkbox.go         (添加 focusIndex 字段)
├── select.go           (添加 focusIndex 字段)
└── component_test.go   (更新测试调用)

runtime/paint/
└── style_state.go      (添加属性关闭检测)
```
