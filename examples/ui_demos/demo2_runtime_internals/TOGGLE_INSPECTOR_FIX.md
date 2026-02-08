# Toggle Inspector 按钮修复说明

## 问题描述

点击 `[I] Toggle Inspector` 按钮后，Inspector 面板没有显示或隐藏。

## 根本原因

问题在于 React 状态管理：

1. **全局变量更新了**：`inspectorEnabled` 全局变量被修改
2. **UI 状态未更新**：没有调用 state setter，所以 UI 不会重新渲染
3. **结果**：界面保持不变

## 技术细节

### 问题代码

```go
func RuntimeDemoWithInspector() ui.VNode {
    // ❌ 忽略了 setter
    showInspector, _ := ui.UseStateBool(inspectorEnabled)

    // ...
    if showInspector {
        // 显示 Inspector
    }
}

// 按钮点击
app.ButtonBuilder("[I] Toggle Inspector").
    OnClick(func() {
        inspectorEnabled = !inspectorEnabled  // ✅ 全局变量更新
        globalInspector.Enable()               // ✅ Inspector 启用
        // ❌ 但 UI 没有重新渲染！
    })
```

### 修复后的代码

```go
func RuntimeDemoWithInspector() ui.VNode {
    // ✅ 保存 setter
    showInspector, setShowInspector := ui.UseStateBool(inspectorEnabled)

    // 传递 setter
    ControlPanel(..., setShowInspector)

    // ...
    if showInspector {
        // 显示 Inspector
    }
}

// ControlPanel 接收参数
func ControlPanel(
    // ... 其他参数
    setShowInspector func(bool),  // ✅ 接收 setter
) ui.VNode {
    // ...

    app.ButtonBuilder("[I] Toggle Inspector").
        OnClick(func() {
            newState := !inspectorEnabled
            inspectorEnabled = newState

            if newState {
                globalInspector.Enable()
            } else {
                globalInspector.Disable()
            }

            setShowInspector(newState)  // ✅ 触发 UI 重新渲染
        })
}
```

## 关键点

### React 状态更新规则

在 Mint TUI（使用 React 模式）中：

1. **必须调用 state setter**
   ```go
   // ✅ 正确
   show, setShow := ui.UseStateBool(false)
   setShow(true)  // 触发重新渲染

   // ❌ 错误
   show, _ := ui.UseStateBool(false)
   show = true  // 不会触发重新渲染
   ```

2. **全局变量 vs State**
   - 全局变量：保存数据
   - State：触发 UI 更新
   - 两者需要配合使用

3. **闭包捕获**
   ```go
   // ✅ 正确
   func() {
       setShowInspector(true)  // 使用闭包中的 setter
   }

   // ❌ 错误
   func() {
       inspectorEnabled = true  // 只修改全局变量
   }
   ```

## 修复效果

### 修复前

```
点击 [I] Toggle Inspector
    ↓
全局变量 inspectorEnabled 更新
    ↓
globalInspector.Enable() 被调用
    ↓
❌ UI 没有重新渲染
    ↓
界面保持不变
```

### 修复后

```
点击 [I] Toggle Inspector
    ↓
全局变量 inspectorEnabled 更新
    ↓
globalInspector.Enable() 被调用
    ↓
✅ setShowInspector(true) 被调用
    ↓
✅ UI 重新渲染
    ↓
✅ Inspector 面板显示
```

## 使用方法

### 运行修复后的版本

```bash
cd examples/ui_demos/demo2_runtime_internals/inspector_demo
./demo2_inspector
```

### 操作步骤

1. **启动程序**
   ```bash
   ./demo2_inspector
   ```

2. **点击 [I] Toggle Inspector 按钮**
   - 按 `Tab` 切换到按钮
   - 按 `Enter` 点击
   - 或直接点击（支持鼠标）

3. **观察效果**
   - ✅ Inspector 面板显示在右侧
   - ✅ 包含性能、诊断、树信息
   - ✅ 再次点击隐藏

4. **测试其他功能**
   - 点击其他按钮触发管道阶段
   - 观察 Inspector 数据实时更新
   - 点击 [I] 再次隐藏 Inspector

## 预期输出

### Inspector 关闭时

```
┌─ Runtime Scheduling Pipeline ─┐
│                               │
│  [Event] [setState]...       │
│                               │
└───────────────────────────────┘
```

### Inspector 打开时

```
┌─ Main Content ────────┬─ Inspector ────┐
│ Pipeline               │                 │
│ Statistics             │ Performance:     │
│ Controls               │   FPS: 60.0      │
│ Explanation           │   Mem: 2.5 MB    │
│                       │                 │
└───────────────────────┴─────────────────┘
```

## 验证修复

### 测试步骤

1. 运行程序
2. 观察 Inspector 初始状态（应该隐藏）
3. 点击 `[I] Toggle Inspector`
4. ✅ Inspector 面板应该显示
5. 再次点击 `[I] Toggle Inspector`
6. ✅ Inspector 面板应该隐藏
7. 多次切换验证稳定性

### 调试信息

如果仍有问题，检查：

```bash
# 添加调试输出
echo "TUI_INSPECTOR=true" | ./demo2_inspector

# 或在代码中添加日志
fmt.Printf("Inspector enabled: %v\n", inspectorEnabled)
fmt.Printf("Show inspector: %v\n", showInspector)
```

## 相关提交

- Commit: `ac57f65b` - fix(inspector): fix Toggle Inspector button not updating UI

## 经验教训

### React 模式最佳实践

1. **始终保存 state setter**
   ```go
   state, setState := ui.UseStateBool(false)
   // 不要写成: state, _ := ui.UseStateBool(false)
   ```

2. **更新状态时调用 setter**
   ```go
   // 更新全局变量
   globalVar = newValue
   // 更新 UI
   setState(newValue)
   ```

3. **传递 setter 给子组件**
   ```go
   // 父组件
   show, setShow := ui.UseStateBool(false)
   ChildComponent(setShow)

   // 子组件
   func ChildComponent(setShow func(bool)) {
       // 使用 setShow 触发更新
   }
   ```

### 状态管理流程

```
用户操作
    ↓
事件处理器
    ↓
更新全局变量 + 调用 setState
    ↓
React 重新渲染
    ↓
UI 更新
```

---

**修复日期**: 2025-02-08
**状态**: ✅ 已修复并测试
**Commit**: `ac57f65b`
