# Phase 1: Fiber 结构优化

## 概述
**时间**: 1-2 天
**优先级**: P0（必须）
**依赖**: 无

## 目标
清理 Fiber 结构，删除冗余字段，为后续优化打下基础。

---

## 当前问题

### 1. 重复字段
```go
// runtime/ui/fiber.go
type Fiber struct {
    // ...
    Instance        ComponentInstance  // ✅ 新字段（Fiber-first）
    ComponentInstance ComponentInstance // ❌ 旧字段（重复）
    // ...
}
```

### 2. ComputedBox 类型不明确
```go
// 当前
ComputedBox interface{} // ❌ 类型不明确

// 目标
// 方案 A: 直接使用 runtime/layout.LayoutBox
// 方案 B: 直接使用 paint.PaintableBox
```

### 3. Deprecated 字段未删除
```go
// runtime/ui/fiber.go
// === Focusable Support (DEPRECATED - Use Instance.FocusableInstance instead) ===
FocusableVNode FocusableVNode // ❌ 应该删除

// === Focusable Metadata (Fiber-first) ===
// DEPRECATED: Use Instance.(FocusableInstance) instead
FocusableMeta *FocusableMeta // ❌ 应该删除
```

---

## 实施步骤

### Step 1.1: 删除重复的 ComponentInstance 字段

**文件**: `runtime/ui/fiber.go`

**操作**:
1. 将所有使用 `ComponentInstance` 字段的代码改为使用 `Instance`
2. 删除 `ComponentInstance` 字段定义

**代码变更**:

```go
// Before
type Fiber struct {
    // ...
    Instance        ComponentInstance
    ComponentInstance ComponentInstance // 删除这个
}

// After
type Fiber struct {
    // ...
    Instance ComponentInstance
}
```

**搜索命令**:
```bash
grep -r "f\.ComponentInstance" --include="*.go"
```

**测试命令**:
```bash
go test ./runtime/ui/... -v
```

---

### Step 1.2: 明确 ComputedBox 类型

**文件**: `runtime/ui/fiber.go`

**分析**:
- ComputedBox 用于存储布局结果
- 当前是 interface{}，需要明确类型

**重要架构原则**:

> **Fiber 只存储布局输入（Style），不存储布局输出**
> 
> 布局结果由 `layout.LayoutBox` 表示，由 `runtime/layout.Engine` 输出
> 布局结果不存储在 Fiber 中，而是在 Layout 阶段生成，传递给 Paint 阶段

**推荐方案**：删除 Fiber.ComputedBox 字段

```go
// Before
type Fiber struct {
    ComputedBox interface{} // ❌ 删除此字段
}

// After
type Fiber struct {
    // 布局输入保留在 Fiber
    Style Style // ✅ 保留
    
    // 布局结果不存储在 Fiber，由 LayoutEngine 输出
    // ComputedBox 已删除
}
```

**理由**：
1. **关注点分离**：Fiber 负责存储输入数据，LayoutEngine 负责输出布局结果
2. **纯函数布局**：布局是纯计算，输入 Fiber，输出 LayoutBox
3. **避免状态混乱**：布局结果不应该回写到 Fiber

**数据流**：

```
Fiber (Style, Instance) → FiberToNodeAdapter → layout.Engine.Layout() → layout.LayoutBox
```

**影响文件**:
- `runtime/ui/fiber.go`
- `runtime/compute/adapter_fiber.go` - 不再从 Fiber.ComputedBox 读取
- `internal/render/layout_switcher.go` - 使用 layout.LayoutBox
- `internal/render/fiber_adapter.go` - 适配器直接读取 Fiber.Style

---

### Step 1.3: 删除 Deprecated 字段

**文件**: `runtime/ui/fiber.go`

**要删除的字段**:
```go
// 删除以下字段
FocusableVNode FocusableVNode
FocusableMeta  *FocusableMeta
```

**前提条件**:
- 确认所有 Focusable 操作已迁移到 Instance
- 确认 GetFocusableInstance() 方法可用

**代码变更**:

```go
// 删除这些字段
type Fiber struct {
    // ...
    
    // === Focusable Support (DEPRECATED - Use Instance.FocusableInstance instead) ===
    // FocusableVNode FocusableVNode // 删除
    
    // === Focusable Metadata (Fiber-first) ===
    // DEPRECATED: Use Instance.(FocusableInstance) instead
    // FocusableMeta *FocusableMeta // 删除
}
```

**搜索命令**:
```bash
grep -r "FocusableVNode\|FocusableMeta" --include="*.go"
```

---

### Step 1.4: 更新 GetInstance 方法

**文件**: `runtime/ui/fiber.go`

**代码变更**:

```go
// Before
func (f *Fiber) GetInstance() ComponentInstance {
    // Priority: Instance field (new) > ComponentInstance field (legacy)
    if f.Instance != nil {
        return f.Instance
    }
    return f.ComponentInstance
}

// After
func (f *Fiber) GetInstance() ComponentInstance {
    return f.Instance
}
```

---

### Step 1.5: 添加辅助方法

**文件**: `runtime/ui/fiber.go`

**新增方法**:

```go
// HasInstance returns true if fiber has an instance.
func (f *Fiber) HasInstance() bool {
    return f.Instance != nil
}

// HasStyle returns true if fiber has style defined.
func (f *Fiber) HasStyle() bool {
    return f.Style != nil
}

// GetStyle returns the style for layout.
// This is the layout input, not the layout result.
func (f *Fiber) GetStyle() Style {
    return f.Style
}
```

> **注意**：不再需要 `GetLayoutBox()` 和 `SetLayoutBox()` 方法，
> 因为布局结果不存储在 Fiber 中。

---

## 测试计划

### 单元测试
```bash
# 测试 Fiber 结构
go test ./runtime/ui -run TestFiber -v

# 测试 Instance 获取
go test ./runtime/ui -run TestGetInstance -v

# 测试 LayoutBox
go test ./runtime/ui -run TestLayoutBox -v
```

### 集成测试
```bash
# 运行所有 runtime 测试
go test ./runtime/... -v

# 运行渲染测试
go test ./internal/render/... -v
```

### 示例验证
```bash
# 运行示例应用
cd examples/fiber_counter
go run main.go

# 验证功能正常
# 1. 界面显示正常
# 2. 交互正常
# 3. 无错误日志
```

---

## 验收标准

### 代码标准
- [x] 删除 ComponentInstance 重复字段
- [x] 删除 ComputedBox 字段（布局结果不存储在 Fiber）
- [x] 删除 FocusableVNode 字段
- [x] 删除 FocusableMeta 字段
- [x] 更新 GetInstance 方法
- [x] 添加辅助方法（HasInstance, HasStyle, GetStyle）

### 测试标准
- [ ] 所有单元测试通过
- [ ] 所有集成测试通过
- [ ] 示例应用正常运行
- [ ] 无编译警告

### 文档标准
- [ ] 更新 fiber.go 注释
- [ ] 更新相关文档

---

## 回滚计划

### Git Tag
```bash
# 开始前打 tag
git tag -a refactor-phase1-start -m "Start Phase 1 refactoring"

# 完成后打 tag
git tag -a refactor-phase1-done -m "Complete Phase 1 refactoring"
```

### 回滚命令
```bash
git checkout refactor-phase1-start
```

---

## 完成检查清单

### 代码变更
- [ ] 修改 runtime/ui/fiber.go
- [ ] 搜索并替换所有 f.ComponentInstance
- [ ] 删除 ComputedBox 字段（不存储布局结果）
- [ ] 删除 deprecated 字段
- [ ] 添加辅助方法

### 测试验证
- [ ] go test ./runtime/ui/... -v
- [ ] go test ./runtime/... -v
- [ ] go test ./internal/render/... -v
- [ ] 运行示例应用验证

### 文档更新
- [ ] 更新 fiber.go 注释
- [ ] 更新 CHANGELOG
- [ ] 更新迁移指南

---

**下一步**: [Phase 2: Layout 引擎优化](./phase2_layout_engine.md)
