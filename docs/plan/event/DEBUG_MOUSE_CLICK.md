# 调试鼠标点击不工作问题

## 问题
- Tab/Modal 按钮键盘可以工作（Enter、Tab）
- 但鼠标点击不工作

## 诊断步骤

### 1. 运行调试模式查看组件注册

```bash
cd examples/ui_demos/demo1_full_featured
TUI_DEBUG_UI=true go run main.go 2>&1 | grep -i "component\|hitmap"
```

期望输出：
```
[APP] root type: *render.DeclarativeNode
[APP] HitMap built from VNode: X entries
[APP] Component registry built: Y components
[APP] Registered component: <id>
...
```

### 2. 点击按钮时查看事件路由

按 Tab 键移动焦点到按钮，然后用鼠标点击。应该看到：
```
[APP] Direct routing: MouseMsg → <component-id>
[APP] Component returned Cmd: ...
```

如果看到 `Component not found: <id>`，说明 ID 不匹配。

### 3. 检查 HitMap 内容

在 `framework/app.go` 的 `render()` 方法后添加调试输出：

```go
if os.Getenv("TUI_DEBUG_HITMAP") == "true" {
    for i, entry := range a.hitMap.AllEntries() {
        fmt.Fprintf(os.Stderr, "[HITMAP] Entry %d: ID=%s, Bounds=%v\n",
            i, entry.NodeID, entry.Bounds)
    }
}
```

### 4. 检查 ComponentRegistry 内容

在 `framework/app.go` 的 `buildComponentRegistry()` 方法后添加调试输出：

```go
if os.Getenv("TUI_DEBUG_UI") == "true" {
    a.componentReg.Each(func(id string, updater component.Updater) {
        fmt.Fprintf(os.Stderr, "[REGISTRY] Component: %s, Type: %T\n", id, updater)
    })
}
```

## 可能的原因

### 原因 1: 组件没有在 HitMap 中
- 症状：点击按钮时没有任何调试输出
- 解决：检查组件的 bounds 是否正确（width > 0 && height > 0）

### 原因 2: ID 不匹配
- 症状：看到 "Component not found"
- 解决：确保 HitMap 使用的 NodeID 与 ComponentRegistry 使用的 ID 一致

### 原因 3: 组件未实现 Updater 接口
- 症状：组件没有出现在 registry 构建输出中
- 解决：确保组件实现了 `component.Updater` 接口

### 原因 4: LocalX/LocalY 不正确
- 症状：事件被路由到组件，但组件没有响应
- 解决：检查 HitMap 的 LocalXY 函数是否正确

## 快速验证

在 demo1 启动后，点击一个按钮，然后立即按 Ctrl+C 退出，查看输出中是否有：
- `[APP] Direct routing: MouseMsg → ...`
- 如果没有，说明 HitTest 没找到组件
- 如果有但后面是 "Component not found"，说明 ID 不匹配
