# Context System 实施方案（Phase 2）

> **目标**: 实现依赖注入机制，解决 Props Drilling
> **预计工作量**: 3天
> **优先级**: 🔴 P0
> **依赖**: Phase 1 (Instance Tree)

---

## 📋 执行摘要

Context 类似 **React Context**，用于跨层级依赖注入，解决：

- ❌ Props Drilling（深层嵌套需手动传递 Props）
- ❌ 组合组件需手动回调
- ❌ 全局状态（Theme、Router）难以访问

Phase 2 的目标是实现完整的 Context System，让组件可以通过 `UseContext` 访问任意层级的父组件提供的数据。

---

## 🏗️ 一、核心设计

### 1.1 Context 概念

```
逻辑树（Instance Tree）
├─ App (提供 ThemeContext)
│  ├─ Page
│  │  ├─ Toolbar
│  │  │  └─ Button (消费 ThemeContext)
│  │  └─ Form (提供 FormContext)
│  │     └─ Field (消费 FormContext 和 ThemeContext)
│  └─ Footer
```

### 1.2 核心原则

1. **类型安全**: 使用 Go 泛型确保类型安全
2. **不可变性**: Context 值一旦设置不可修改
3. **作用域**: Context 只在 Fiber 树中有效
4. **嵌套**: 支持多层 Context 覆盖

---

## 🔧 二、接口定义

### 2.1 Context Core

文件：`runtime/context/context.go`

```go
package context

import (
	"sync"
)

// ContextKey Context 键类型（类型安全）
typedef ContextKey string

// ContextValue Context 值接口
type ContextValue interface {
	// 可以添加额外的方法
}

// FiberContext Fiber 级别的 Context 存储
type FiberContext struct {
	mu     sync.RWMutex
	values map[ContextKey]any
	parent *FiberContext // 支持嵌套
}

// NewContext 创建新的 Context
// 如果 parent 为 nil，创建根 Context
func NewContext(parent *FiberContext) *FiberContext {
	return &FiberContext{
		values: make(map[ContextKey]any),
		parent: parent,
	}
}

// Provide 提供一个 Context 值
// 在组件 VNode 或 VNode.Children() 中调用
func (c *FiberContext) Provide(key ContextKey, value any) {
	if c == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.values[key] = value
}

// UseContext 从 Context 中获取值
// 在组件 Instance 中调用
func (c *FiberContext) UseContext(key ContextKey) any {
	if c == nil {
		return nil
	}

	// 当前层级查找
	c.mu.RLock()
	v, ok := c.values[key]
	c.mu.RUnlock()

	if ok {
		return v
	}

	// 向上查找
	if c.parent != nil {
		return c.parent.UseContext(key)
	}

	return nil
}

// UseContextValue 类型安全的 Context 访问（Go 1.18+ 泛型）
func UseContextValue[T any](c *FiberContext, key ContextKey) (T, bool) {
	v := c.UseContext(key)
	if v == nil {
		var zero T
		return zero, false
	}

	t, ok := v.(T)
	if !ok {
		var zero T
		return zero, false
	}

	return t, true
}

// HasContext 检查是否有指定的 Context
func (c *FiberContext) HasContext(key ContextKey) bool {
	if c == nil {
		return false
	}

	c.mu.RLock()
	_, ok := c.values[key]
	c.mu.RUnlock()

	if ok {
		return true
	}

	if c.parent != nil {
		return c.parent.HasContext(key)
	}

	return false
}

// AllKeys 获取当前层级的所有 Key（调试用）
func (c *FiberContext) AllKeys() []ContextKey {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]ContextKey, 0, len(c.values))
	for k := range c.values {
		keys = append(keys, k)
	}

	return keys
}

// Clear 清空当前层级的所有值（通常在 fiber 卸载时调用）
func (c *FiberContext) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 保留 parent，只清空当前层级
	c.values = make(map[ContextKey]any)
}
```

### 2.2 Context Provider 组件

文件：`runtime/ui/context_provider.go`

```go
package context

import (
	"github.com/wwsheng009/mint/runtime/context"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Provider Context 提供者组件
// Props:
//   - "key": context.ContextKey
//   - "value": any
//   - "child": rtui.VNode
type Provider struct {
	rtui.ElementVNode

	key   context.ContextKey
	value any
}

func NewProvider(key context.ContextKey, value any, child rtui.VNode) *Provider {
	return &Provider{
		ElementVNode: rtui.NewElement("provider"),
		key:         key,
		value:       value,
	}
}

func (p *Provider) Key() string {
	return "provider:" + string(p.key)
}

func (p *Provider) Type() string {
	return "provider"
}

func (p *Provider) Tag() string {
	return "provider"
}

func (p *Provider) Props() rtui.Props {
	return rtui.Props{
		"contextKey": p.key,
		"contextValue": p.value,
	}
}

func (p *Provider) Children() []rtui.VNode {
	return []rtui.VNode{p.value.(rtui.VNode)}
}
```

---

## 🔗 三、Fiber 集成

### 3.1 修改 Fiber 结构

文件：`runtime/ui/fiber.go`

```go
package ui

import (
	fcontext "github.com/wwsheng009/mint/runtime/context"
)

type Fiber struct {
	// ... 原有字段

	// Context 栈
	Context *fcontext.FiberContext
}
```

### 3.2 修改 CreateFiber

文件：`runtime/ui/fiber_util.go`

```go
// CreateFiber 创建 Fiber 节点
func CreateFiber(vnode VNode, parent *Fiber) *Fiber {
	// ... 原有逻辑

	// 继承或创建 Context
	var fcontext *fcontext.FiberContext
	if parent != nil {
		// 继承父 Context
		fcontext = fcontext.NewContext(parent.Context)
	} else {
		// 创建根 Context
		fcontext = fcontext.NewContext(nil)
	}

	f := &Fiber{
		// ... 原有字段
		Context: fcontext,
	}

	return f
}
```

### 3.3 处理 Provider 组件

文件：`internal/reconciler/reconciler.go`

```go
func reconcile(returnFiber, currentFiber, workInProgress *Fiber) *Fiber {
	// ... 原有逻辑

	// 如果是 Provider 组件，注入 Context
	if vnode.Type() == "provider" {
		reconcileProvider(workInProgress, vnode)
	}

	// ... 原有逻辑
}

func reconcileProvider(f *Fiber, vnode VNode) {
	// 从 Props 获取 key 和 value
	key := vnode.Props()["contextKey"].(context.ContextKey)
	value := vnode.Props()["contextValue"]

	// 提供到 Context
	f.Context.Provide(key, value)
}
```

---

## 🧩 四、使用示例

### 4.1 OptionGroup 注入 Context

文件：`ui/components/optiongroup/instance.go`

```go
package optiongroup

import (
	fcontext "github.com/wwsheng009/mint/runtime/context"
	// ...
)

// ===== 定义 Context Key =====
const OptionGroupContext fcontext.ContextKey = "optionGroup"

// ===== OptionGroup 提供 Context =====

func (inst *Instance) OnMount() {
	inst.BaseInstance.OnMount()

	// 获取当前 Fiber
	f := inst.GetFiber()

	// 提供 OptionGroup Instance 到 Context
	if f != nil && f.Context != nil {
		f.Context.Provide(OptionGroupContext, inst)
	}
}

func (inst *Instance) OnUnmount() {
	inst.BaseInstance.OnUnmount()

	// 清理 Context（可选，Fiber 会自动处理）
}
```

### 4.2 Option 消费 Context

文件：`ui/components/optiongroup/option.go`

```go
package optiongroup

import (
	fcontext "github.com/wwsheng009/mint/runtime/context"
	// ...
)

type OptionInstance struct {
	instance.BaseInstance

	// ... 原有字段

	// 缓存的父实例（可选）
	group *Instance
}

// ===== 方式1: OnMount 时缓存 =====

func (inst *OptionInstance) OnMount() {
	inst.BaseInstance.OnMount()

	f := inst.GetFiber()
	if f != nil && f.Context != nil {
		// 类型安全的 Context 访问
		group, ok := fcontext.UseContextValue[*Instance](
			f.Context,
			OptionGroupContext,
		)

		if ok {
			inst.group = group
		}
	}
}

// ===== 方式2: 动态查找（推荐） =====

func (inst *OptionInstance) GetParentGroup() (*Instance, bool) {
	f := inst.GetFiber()
	if f == nil || f.Context == nil {
		return nil, false
	}

	return fcontext.UseContextValue[*Instance](
		f.Context,
		OptionGroupContext,
	)
}

// ===== 使用 Context =====

func (inst *OptionInstance) HandleAction(act *action.Action) bool {
	switch act.Type {
	case action.ActionClick, action.ActionEnter:
		// 方式1: 使用缓存的引用
		if inst.group != nil {
			inst.group.SelectOption(inst.value)
			return true
		}

		// 方式2: 动态查找
		group, ok := inst.GetParentGroup()
		if ok {
			group.SelectOption(inst.value)
			return true
		}
	}
	return false
}
```

### 4.3 Theme Context 示例

```go
// 定义 Theme 类型
type Theme struct {
	PrimaryColor   string
	SecondaryColor string
	BackgroundColor string
	TextColor      string
}

// 定义 Context Key
const ThemeContext fcontext.ContextKey = "theme"

// ===== 提供者 =====

func NewThemeProvider(theme Theme, child rtui.VNode) *Provider {
	return NewProvider(ThemeContext, theme, child)
}

// ===== 消费者 =====

func NewButton(text string) VNode {
	return NewComponent("button", &Button{Text: text})
}

type Button struct {
	ElementVNode
	Text string
}

func (b *Button) CreateInstance() ComponentInstance {
	return &ButtonInstance{
		BaseInstance: instance.BaseInstance{},
		Text:         b.Text,
	}
}

type ButtonInstance struct {
	instance.BaseInstance
	Text string
}

func (b *ButtonInstance) Layout(ctx LayoutContext) Rect {
	// 从 Context 获取 Theme
	f := b.GetFiber()

	if theme, ok := fcontext.UseContextValue[Theme](f.Context, ThemeContext); ok {
		// 使用 theme.PrimaryColor
		// ...
	}

	// ... Layout 逻辑
	return ctx.Bounds
}
```

### 4.4 Form Context 示例

```go
// 定义 Form 类型
type FormInstance struct {
	instance.BaseInstance
	fields map[string]*FieldInstance
	// ...
}

const FormContext fcontext.ContextKey = "form"

// ===== Form 提供者 =====

func (inst *FormInstance) OnMount() {
	inst.BaseInstance.OnMount()

	f := inst.GetFiber()
	if f != nil && f.Context != nil {
		f.Context.Provide(FormContext, inst)
	}
}

// ===== Field 消费者 =====

type FieldInstance struct {
	instance.BaseInstance
	name string
}

func (inst *FieldInstance) HandleAction(act *action.Action) bool {
	// 从 Context 获取 Form
	f := inst.GetFiber()

	if form, ok := fcontext.UseContextValue[*FormInstance](f.Context, FormContext); ok {
		// 注册到 form 或者获取 form 的状态
		// ...
	}

	return false
}
```

---

## 🧪 五、单元测试

### 5.1 Context Core 测试

文件：`runtime/context/context_test.go`

```go
package context

import (
	"sync"
	"testing"
)

func TestContext_ProvideAndUse(t *testing.T) {
	TestKey := ContextKey("test-key")
	TestValue := 123

	ctx := NewContext(nil)
	ctx.Provide(TestKey, TestValue)

	result := ctx.UseContext(TestKey)

	if result != TestValue {
		t.Errorf("expected %v, got %v", TestValue, result)
	}
}

func TestContext_NestedLookup(t *testing.T) {
	ParentKey := ContextKey("parent")
	ChildKey := ContextKey("child")

	parent := NewContext(nil)
	parent.Provide(ParentKey, "parent-value")

	child := NewContext(parent)
	child.Provide(ChildKey, "child-value")

	// 子可以访问自己的值
	if child.UseContext(ChildKey) != "child-value" {
		t.Error("child should have its own value")
	}

	// 子可以访问父的值
	if child.UseContext(ParentKey) != "parent-value" {
		t.Error("child should access parent value")
	}

	// 父不能访问子的值
	if parent.UseContext(ChildKey) != nil {
		t.Error("parent should not access child value")
	}
}

func TestContext_Override(t *testing.T) {
	TestKey := ContextKey("override-test")

	parent := NewContext(nil)
	parent.Provide(TestKey, "parent-value")

	child := NewContext(parent)
	child.Provide(TestKey, "child-value")

	// 子应该返回自己的值
	if child.UseContext(TestKey) != "child-value" {
		t.Error("child should override parent value")
	}

	// 父应该返回自己的值
	if parent.UseContext(TestKey) != "parent-value" {
		t.Error("parent should keep its value")
	}
}

func TestContext_HasContext(t *testing.T) {
	TestKey := ContextKey("has-test")

	ctx := NewContext(nil)

	if ctx.HasContext(TestKey) {
		t.Error("empty context should not have key")
	}

	ctx.Provide(TestKey, "value")

	if !ctx.HasContext(TestKey) {
		t.Error("context should have key after provide")
	}
}

func TestContext_ThreadSafety(t *testing.T) {
	TestKey := ContextKey("thread-test")
	ctx := NewContext(nil)

	var wg sync.WaitGroup
	concurrency := 100

	// 并发写入
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			ctx.Provide(TestKey, val)
		}(i)
	}

	wg.Wait()

	// 验证可以读取
	if ctx.UseContext(TestKey) == nil {
		t.Error("should be able to read after concurrent writes")
	}
}
```

### 5.2 泛型测试

文件：`runtime/context/generic_test.go`

```go
package context

import (
	"testing"
)

func TestUseContextValue_TypeSafety(t *testing.T) {
	StringKey := ContextKey("string")
	IntKey := ContextKey("int")

	ctx := NewContext(nil)
	ctx.Provide(StringKey, "hello")
	ctx.Provide(IntKey, 42)

	// 类型安全的访问
	str, ok := UseContextValue[string](ctx, StringKey)
	if !ok {
		t.Error("should retrieve string value")
	}
	if str != "hello" {
		t.Error("string value mismatch")
	}

	num, ok := UseContextValue[int](ctx, IntKey)
	if !ok {
		t.Error("should retrieve int value")
	}
	if num != 42 {
		t.Error("int value mismatch")
	}

	// 类型不匹配
	_, ok = UseContextValue[string](ctx, IntKey)
	if ok {
		t.Error("type mismatch should return false")
	}
}

type TestStruct struct {
	Name string
	Age  int
}

func TestUseContextValue_CustomType(t *testing.T) {
	TestKey := ContextKey("struct")

	ctx := NewContext(nil)
	value := TestStruct{
		Name: "test",
		Age:  30,
	}
	ctx.Provide(TestKey, value)

	result, ok := UseContextValue[TestStruct](ctx, TestKey)
	if !ok {
		t.Error("should retrieve custom type")
	}
	if result.Name != "test" || result.Age != 30 {
		t.Error("custom type value mismatch")
	}
}
```

### 5.3 Fiber 集成测试

文件：`runtime/ui/context_test.go`

```go
package ui

import (
	"testing"

	fcontext "github.com/wwsheng009/mint/runtime/context"
)

func TestFiberContext_Propagation(t *testing.T) {
	TestKey := fcontext.ContextKey("fiber-test")

	// 创建 Fiber 树
	childVNode := NewText("Child").Key("child")
	parentVNode := NewStack().
		Children(childVNode).
		Key("parent")

	root := CreateFiberFromVNode(parentVNode)

	// 在根节点提供 Context
	root.Context.Provide(TestKey, "root-value")

	// 验证子节点可以访问
	childFiber := findFiberByKey(root, "child")
	if childFiber == nil {
		t.Fatal("child fiber not found")
	}

	value := childFiber.Context.UseContext(TestKey)
	if value != "root-value" {
		t.Error("child should access root context value")
	}
}

func findFiberByKey(root *Fiber, key string) *Fiber {
	var result *Fiber
	var traverse func(*Fiber)
	traverse = func(f *Fiber) {
		if f.Key() == key {
			result = f
			return
		}
		if f.Child != nil {
			traverse(f.Child)
		}
		if f.Sibling != nil {
			traverse(f.Sibling)
		}
	}
	traverse(root)
	return result
}
```

---

## ✅ 六、验证检查清单

### 6.1 编译检查

```bash
cd E:\projects\yao\wwsheng009\mint
go build ./runtime/context/...
go build ./ui/components/context/...
```

**预期结果**:
- ✅ 无编译错误
- ✅ 泛型正常工作（Go 1.18+）

### 6.2 单元测试

```bash
# 测试 Context Core
golangci-lint run ./runtime/context/...

# 运行测试
go test -v ./runtime/context/...

# 运行泛型测试
go test -v ./runtime/context/... -run TestUseContextValue
```

**预期结果**:
- ✅ 所有测试通过
- ✅ 并发安全测试通过
- ✅ 泛型类型安全

### 6.3 集成测试

```bash
# 测试 OptionGroup Context 集成
cd ui/components/optiongroup
go test -v .

# 运行完整示例
cd examples/multiselect_demo
go run main.go
```

**手动测试**:
- [ ] 应用正常启动
- [ ] OptionGroup 注入 Context 成功
- [ ] Option 消费 Context 成功
- [ ] 点击选项可以选中（通过 Context）

### 6.4 性能测试

```bash
# 基准测试
cd runtime/context
go test -bench=. -benchmem .
```

**预期结果**:
- `BenchmarkProvide`: ~10-50 ns/op
- `BenchmarkUseContext`: ~5-20 ns/op
- `BenchmarkUseContextValue`: ~5-20 ns/op
- 无内存分配（Context 已预分配）

---

## 🐛 七、常见问题

### 问题 1: Context Key 冲突

**症状**:
不同组件使用相同的 Context Key，导致值被覆盖。

**原因**:
Context Key 是字符串，可能意外冲突。

**解决方案**:
使用命名空间：

```go
// 定义带有命名空间的 Key
const OptionGroupContext fcontext.ContextKey = "optiongroup/group"
const OptionGroupItemContext fcontext.ContextKey = "optiongroup/item"

// 或者使用包级变量
var OptionGroupContext = fcontext.ContextKey("github.com/wwsheng009/mint/ui/components/optiongroup:group")
```

### 问题 2: Context 在 Fiber 重建时丢失

**症状**:
重新渲染后，Context 值丢失。

**原因**:
Fiber 重建时未正确继承 Context。

**解决方案**:
确保 `CreateFiber` 正确继承：

```go
func CreateFiber(vnode VNode, parent *Fiber) *Fiber {
	f := &Fiber{
		// ...
		Context: fcontext.NewContext(parent.Context), // 确保继承
	}
	return f
}
```

### 问题 3: 泛型类型断言失败

**症状**:
`UseContextValue[T]` 返回 `false`。

**原因**:
存储的类型与请求的类型不匹配。

**解决方案**：
```go
// 错误示例
ctx.Provide(ThemeContext("theme"), theme) // ❌ 错误：创建了新类型
ctx.Provide(ThemeContext, theme)          // ✅ 正确

// 确保类型完全匹配
// 如果使用指针，必须一致
ctx.Provide(ThemeContext, &theme)  // 存储指针
theme, ok := UseContextValue[*Theme](ctx, ThemeContext) // 获取指针 ✅
```

---

## 📚 八、参考资料

### 设计参考

1. **React Context API**
   - https://react.dev/reference/react/useContext
   - 消费者模式

2. **Go 泛型**
   - https://go.dev/doc/tutorial/generics
   - 类型安全

3. **Flutter InheritedWidget**
   - https://api.flutter.dev/flutter/widgets/InheritedWidget-class.html
   - 继承机制

### Mint 相关

- `runtime/context/context.go` - Context 实现
- `runtime/ui/fiber.go` - Fiber 集成
- `ui/components/optiongroup/` - OptionGroup 示例

---

## 🎯 九、下一步

### Phase 3: Intent Bubble

完成 Context System 后，可以实施：

- [ ] 定义 Intent 接口
- [ ] 实现 Intent Bubble 机制
- [ ] 定义 OptionGroup Intent
- [ ] 修改 Option 发射 Intent

Phase 3 将进一步解耦父子组件，实现完全声明式的事件处理。

---

**文档状态**: ✅ 设计完成，待实施

**预计开始日期**: Phase 1 完成后
