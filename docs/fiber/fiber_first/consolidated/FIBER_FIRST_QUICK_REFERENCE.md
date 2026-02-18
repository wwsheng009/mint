# Fiber-First 快速参考指南

## 核心概念速查

### 三层架构

| 层级 | 角色 | 生命周期 | 示例 |
|------|------|----------|------|
| VNode | 描述 | 短期（render后丢弃） | `ButtonVNode{label: "Click"}` |
| Fiber | 结构 | 长期（运行期持久） | `Fiber{Type: ButtonType, Key: "btn1"}` |
| paint.PaintableBox | 行为+状态 | 长期（运行期持久） | `ButtonInstance{hasFocus: true}` |

### 关键原则

```
❌ Fiber 不持有 VNode 引用
❌ Fiber 不存储业务回调函数
❌ Layout 不能访问 VNode
❌ Event 不能访问 VNode

✅ Fiber 持有 paint.PaintableBox 引用
✅ Layout 只读 Fiber
✅ Event 通过 ActionBridge 路由
✅ VNode 在 commit 后丢弃
```

---

## 常用代码模式

### 创建 Fiber

```go
func CreateFiber(vnode VNode) *Fiber {
    // ... extract props from vnode

    // Create Instance from VNode
    var instance ComponentInstance
    if factory, ok := vnode.(InstanceFactory); ok {
        instance = factory.CreateInstance()
    }

    return &Fiber{
        Props:    props,
        Style:    style,
        Instance: instance,  // paint.PaintableBox persists
        // NO VNode reference!
    }
}
```

### 克隆 Fiber

```go
func CloneFiber(fiber *Fiber) *Fiber {
    return &Fiber{
        // ... copy other fields
        Instance: fiber.Instance,  // REUSE, never clone
    }
}
```

### 事件分发

```go
func (b *ActionBridge) DispatchFromFiber(
    start *Fiber,
    actionType action.ActionType,
    payload interface{},
) bool {
    for f := start; f != nil; f = f.Return {
        if f.ActionTargetID == "" {
            continue
        }
        
        a := action.NewAction(actionType).
            WithTarget(f.ActionTargetID).
            WithPayload(payload)
        
        if handled := b.dispatcher.Dispatch(a); handled {
            return true
        }
    }
    return false
}
```

### 绘制

```go
func (inst *ButtonInstance) Paint(x, y int) []paint.DrawCmd {
    // Use instance state only
    // No VNode dependency!
}
```

---

## 迁移清单

### 从 VNode → Fiber

- [ ] Type
- [ ] Key
- [ ] Props → MemoizedProps
- [ ] Style
- [ ] Text → MemoizedState
- [ ] ActionTargetID

### 从 Fiber 删除

- [ ] VNode 引用
- [ ] LayoutBox（移到 paint.PaintableBox）
- [ ] 业务回调函数

---

## 判断标准

### 成功标准

问自己：
> 如果我删除 VNode struct，Layout + Render 是否还能运行？

如果答案是 **YES**，你就完成了 Fiber-first。

### 常见错误

1. **Fiber 持有 VNode**
   ```go
   fiber.VNode = vnode  // ❌ 错误
   ```

2. **Layout 访问 VNode**
   ```go
   vnode.Props()  // ❌ 错误
   ```

3. **Event 直接调用回调**
   ```go
   fiber.OnClick()  // ❌ 错误
   ```

4. **Instance 在 Clone 时被复制**
   ```go
   wip.Instance = clone(old.Instance)  // ❌ 错误
   ```

---

## 性能优化要点

### 不要过早优化

```
架构清晰 > 性能极限
```

### 优化顺序

1. **结构纯化**（现在）
   - Layout 彻底 Fiber-first
   - diffChildren 完全 Fiber-only
   - 删除 VNode 依赖

2. **调度纯化**（之后）
   - Lane 优先级
   - 批处理
   - 时间切片

3. **并发增强**（最后）
   - Lazy clone
   - subtree bailout
   - 内存优化

---

## 关键文件位置

| 功能 | 文件 |
|------|------|
| Fiber 定义 | `runtime/ui/fiber.go` |
| Instance 接口 | `runtime/ui/instance.go` |
| ActionBridge | `runtime/bridge/actionbridge/bridge.go` |
| Fiber 工具 | `runtime/ui/fiber_util.go` |
| Button 示例 | `components/button/` |

---

## 调试技巧

### 检查 VNode 泄漏

```bash
# 全局搜索 fiber.VNode
grep -r "fiber\.VNode" --include="*.go"

# 全局搜索 vnode 运行期访问
grep -r "vnode\." --include="*.go" | grep -v "//.*vnode"
```

### 检查 Instance 复用

```go
// 添加日志
func CloneFiber(fiber *Fiber) *Fiber {
    if fiber.Instance != nil {
        log.Printf("Reusing instance: %T", fiber.Instance)
    }
    return &Fiber{
        Instance: fiber.Instance,  // REUSE
    }
}
```

---

## 一句话总结

> **VNode 是"描述"，Instance 是"行为"，Fiber 是"调度结构"，三者必须彻底解耦。**
