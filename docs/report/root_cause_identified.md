# Button.Paint 未被调用的根本原因

**问题分析完成** ✅

---

## 根本原因

**PaintEngine 根本没有被调用！**

### 证据链

1. **SetBounds 被调用** ✅
   ```
   [calculatePositions] BUTTON: SetBounds type assertion SUCCESS
   calling SetBounds(x=1, y=12, w=18, h=0)
   ```
   说明：`calculatePositions()` 确实执行到了按钮

2. **Button.Paint() 没有输出** ❌
   ```
   添加了日志：
   fmt.Fprintf(os.Stderr, "[Button.Paint] label=%q, ...\n", ...)

   但输出中完全没有出现！
   ```
   说明：`Button.Paint()` 没有被调用

3. **SetBounds 调用来自哪里？**
   ```go
   // runtime/compute/engine.go:856-864
   if boundsAware, ok := box.VNode.(interface{ SetBounds(...) }); ok {
       fmt.Fprintf(os.Stderr, "[calculatePositions] BUTTON: SetBounds type assertion SUCCESS\n")
       boundsAware.SetBounds(x, y, box.Box.Width, box.Box.Height)
   }
   ```
   说明：类型断言成功，ButtonVNode 确实在 ComputedBox 中

---

## 推导出的结论

### 问题1：Demo2 没有使用新的两阶段渲染！

**证据**：
- Button.Paint() 没被调用
- 说明 PaintEngine.Paint() 没被执行
- 说明 RenderingPipeline.Render() 没被使用
- **说明 Demo2 使用的是旧的渲染路径！**

### 问题2：旧的渲染路径在哪里？

**检查**：
```go
// ui/app.go:109-150
fwApp := framework.NewApp()
declarativeRoot = render.NewDeclarativeNodeFromFunc(app)
fwApp.SetRoot(declarativeRoot)
return fwApp.Run()
```

**问题**：framework.App 的 Run() 使用的是什么渲染器？

---

## 解决方案

### 方案A：确保使用新的渲染管线

**步骤1**：检查 framework.App 使用什么渲染器

**步骤2**：如果没有使用 PipelineRenderer，修改配置

**步骤3**：验证 PaintEngine.Paint() 被调用

### 方案B：在旧渲染路径中也调用 SetBounds

如果 Demo2 使用旧渲染，那么：
1. 找到旧渲染的布局代码
2. 添加 SetBounds 调用

---

## 下一步

**立即检查**：
```bash
cd examples/ui_demos/demo2_runtime_internals
MINT_USE_LEGACY_RENDERER=false ./demo2.exe
```

或者检查 framework.App 的默认渲染器配置。
