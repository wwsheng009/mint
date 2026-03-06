# Intent Bubble 实施方案（Phase 3）

> **目标**: 实现事件冒泡机制，解耦父子组件
> **预计工作量**: 3天
> **优先级**: 🟡 P1
> **依赖**: Phase 1 (Instance Tree), Phase 2 (Context System)

---

## 📋 执行摘要

Intent Bubble 类似 **DOM 事件冒泡**，用于组件间事件通信，解耦父子组件：

- ❌ 父子组件强耦合（闭包、回调）
- ❌ 难以实现事件拦截
- ✅ Intent 是纯数据结构，易于序列化和撤销/重做

Phase 3 的目标是实现完整的 Intent Bubble 机制，让组件通过 `Emit` 发射 Intent，父组件通过 `HandleIntent` 拦截。

---

## 🏗️ 一、核心设计

### 1.1 Intent Bubble 概念

```
逻辑树（Instance Tree）
├─ Form (HandleIntent: FormSubmitIntent)
│  ├─ Button (Emit: FormSubmitIntent)
│  │  ↑
│  │  └─ 冒泡
│  │
│  └─ Field (Emit: FieldChangeIntent)
│     ↑
│     └─ 冒泡 → Form (HandleIntent: FieldChangeIntent)
```

### 1.2 核心原则

1. **单向通信**: 子向父冒泡，不反向
2. **可拦截**: 父组件可以捕获并阻止冒泡
3. **纯数据**: Intent 是可序列化的数据结构
4. **声明式**: Intent 显式表达意图，而非底层事件

---

## 🔧 二、接口定义

### 2.1 Intent Core

文件：`runtime/intent/intent.go`

```go
package intent

// Intent 意图接口
type Intent interface {
	// Type 返回意图类型
	Type() string

	// Key 返回意图的唯一键（用于调试）
	Key() string
}

// IntentEmitter Intent 发射器接口
type IntentEmitter interface {
	// Emit 发射一个 Intent，向上冒泡
	Emit(i Intent)
}

// IntentHandler Intent 处理器接口
type IntentHandler interface {
	// HandleIntent 处理 Intent
	// 返回 true 表示已处理，停止冒泡
	// 返回 false 表示未处理，继续冒泡
	HandleIntent(i Intent) bool
}

// IntentFilter Intent 过滤器（可选）
type IntentFilter interface {
	// ShouldBubble 判断 Intent 是否应该冒泡
	ShouldBubble(i Intent) bool
}

// DefaultKey 默认的 Key 实现
func DefaultKey(i Intent) string {
	return i.Type()
}
```

### 2.2 Bubble 机制

文件：`runtime/intent/bubble.go`

```go
package intent

import (
	"fmt"
	"github.com/wwsheng009/mint/runtime/instance"
)

// Emit 向上冒泡 Intent
// 从 inst 开始，逐级向上查找处理器
func Emit(inst instance.ComponentInstance, i Intent) {
	if inst == nil {
		return
	}

	// 向上查找处理器
	current := inst
	depth := 0
	maxDepth := 100 // 防止无限循环

	for current != nil && depth < maxDepth {
		// 检查是否实现了 IntentHandler
		if handler, ok := current.(IntentHandler); ok {
			handled := handler.HandleIntent(i)
			if handled {
				// 已处理，停止冒泡
				return
			}
		}

		// 移动到父节点
		current = current.Parent()
		depth++
	}

	// 如果冒泡到根仍未处理，可以传递给全局 Store
	// 这是一个可选的扩展
	emitUnhandled(i)
}

// emitUnhandled 处理未处理的 Intent
// 可以传递给 Store 或记录日志
func emitUnhandled(i Intent) {
	// 日志记录
	fmt.Printf("[Intent] Unhandled intent: %s (key: %s)\n", i.Type(), i.Key())

	// 可选：传递到全局 Store
	// if store := GetGlobalStore(); store != nil {
	//     store.Dispatch(i)
	// }
}

// EmitWithSender 带发送者信息的 Emit（调试用）
func EmitWithSender(sender instance.ComponentInstance, i Intent, senderName string) {
	if sender == nil {
		Emit(sender, i)
		return
	}

	fmt.Printf("[Intent] Emitting: %s from %s\n", i.Type(), senderName)
	Emit(sender, i)
}
```

### 2.3 Intent 策略（可选）

文件：`runtime/intent/strategy.go`

```go
package intent

import (
	"github.com/wwsheng009/mint/runtime/instance"
)

// BubbleStrategy 冒泡策略
type BubbleStrategy interface {
	// ShouldEmit 判断是否应该发射
	ShouldEmit(inst instance.ComponentInstance, i Intent) bool

	// ShouldBubble 判断是否应该继续冒泡
	ShouldBubble(current instance.ComponentInstance, i Intent) bool
}

// DefaultStrategy 默认策略
type DefaultStrategy struct{}

func (s *DefaultStrategy) ShouldEmit(inst instance.ComponentInstance, i Intent) bool {
	// 默认总是发射
	return true
}

func (s *DefaultStrategy) ShouldBubble(current instance.ComponentInstance, i Intent) bool {
	// 默认总是冒泡
	return true
}

// StopAtRootStrategy 到根节点停止
type StopAtRootStrategy struct{}

func (s *StopAtRootStrategy) ShouldBubble(current instance.ComponentInstance, i Intent) bool {
	// 到根节点停止
	return current.Parent() != nil
}

// EmitWithStrategy 使用特定策略发射
func EmitWithStrategy(inst instance.ComponentInstance, i Intent, strategy BubbleStrategy) {
	if inst == nil || strategy == nil {
		Emit(inst, i)
		return
	}

	if !strategy.ShouldEmit(inst, i) {
		return
	}

	current := inst
	depth := 0
	maxDepth := 100

	for current != nil && depth < maxDepth {
		if handler, ok := current.(IntentHandler); ok {
			handled := handler.HandleIntent(i)
			if handled {
				return
			}
		}

		if !strategy.ShouldBubble(current, i) {
			return
		}

		current = current.Parent()
		depth++
	}

	emitUnhandled(i)
}
```

---

## 🧩 三、OptionGroup Intent 示例

### 3.1 定义 Intent

文件：`ui/components/optiongroup/intent.go`

```go
package optiongroup

import "github.com/wwsheng009/mint/runtime/intent"

// OptionSelectIntent 选项选中Intent
type OptionSelectIntent struct {
	// GroupKey 所属的 OptionGroup 的 key
	GroupKey string

	// Value 选中的值
	Value string

	// IsSelected 是否为选中状态（false 表示取消选中）
	IsSelected bool
}

func (OptionSelectIntent) Type() string {
	return "OptionGroup:Select"
}

func (i OptionSelectIntent) Key() string {
	return fmt.Sprintf("optiongroup:%s:%s", i.GroupKey, i.Value)
}

// OptionFocusIntent 选项选中Intent
type OptionFocusIntent struct {
	GroupKey string
	Value    string
}

func (OptionFocusIntent) Type() string {
	return "OptionGroup:Focus"
}

func (i OptionFocusIntent) Key() string {
	return fmt.Sprintf("optiongroup:%s:%s:focus", i.GroupKey, i.Value)
}
```

### 3.2 Option 发射 Intent

文件：`ui/components/optiongroup/option.go`

```go
package optiongroup

import (
	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	// ...
)

// ===== 实现 IntentEmitter =====

type OptionInstance struct {
	instance.BaseInstance

	// ... 原有字段
	groupKey string // 用于 Intent
}

// ===== 发射 Intent =====

func (inst *OptionInstance) HandleAction(act *action.Action) bool {
	switch act.Type {
	case action.ActionClick, action.ActionEnter:
		// 发射 OptionSelectIntent
		intent.Emit(inst, OptionSelectIntent{
			GroupKey:   inst.groupKey,
			Value:      inst.value,
			IsSelected: !inst.state.Selected,
		})
		return true

	case action.ActionFocus:
		// 发射 OptionFocusIntent（可选）
		intent.Emit(inst, OptionFocusIntent{
			GroupKey: inst.groupKey,
			Value:    inst.value,
		})
		return true
	}
	return false
}

// ===== 或者使用 WithSender（调试） =====

func (inst *OptionInstance) HandleActionDebug(act *action.Action) bool {
	if act.Type == action.ActionClick {
		intent.EmitWithSender(
			inst,
			OptionSelectIntent{
				GroupKey:   inst.groupKey,
				Value:      inst.value,
				IsSelected: true,
			},
			fmt.Sprintf("Option(%s)", inst.value),
		)
		return true
	}
	return false
}
```

### 3.3 OptionGroup 处理 Intent

文件：`ui/components/optiongroup/instance.go`

```go
package optiongroup

import (
	"github.com/wwsheng009/mint/runtime/intent"
	// ...
)

// ===== 实现 IntentHandler =====

// 确保 *Instance 实现了 IntentHandler
var _ intent.IntentHandler = (*Instance)(nil)

func (inst *Instance) HandleIntent(i intent.Intent) bool {
	switch v := i.(type) {
	case OptionSelectIntent:
		// 只处理自己组的 Intent
		if v.GroupKey == inst.key {
			inst.SelectOption(v.Value)
			return true // 已处理
		}

		// 不是自己组的 Intent，继续冒泡
		return false

	case OptionFocusIntent:
		// 只处理自己组的 Focus Intent
		if v.GroupKey == inst.key {
			inst.FocusOption(v.Value)
			return true
		}
		return false
	}

	// 其他 Intent 不处理，继续冒泡
	return false
}

// ===== SelectOption 实现 =====

func (inst *Instance) SelectOption(value string) {
	switch inst.mode {
	case ModeSingle:
		// 单选：只选中一个
		inst.selected = value
		inst.selecteds = nil

	case ModeMultiple:
		// 多选：添加/移除
		index := indexString(inst.selecteds, value)
		if index >= 0 {
			// 移除
			inst.selecteds = append(inst.selecteds[:index], inst.selecteds[index+1:]...)
		} else {
			// 添加
			inst.selecteds = append(inst.selecteds, value)
		}
	}

	// 更新子实例状态
	inst.UpdateChildrenState()

	// 触发重绘
	inst.dirty = true

	// 可选：发射 Intent 到 Store
	if inst.intentEmitter != nil {
		inst.emitFieldChange()
	}
}

// ===== UpdateChildrenState 更新子实例状态 =====

func (inst *Instance) UpdateChildrenState() {
	for _, childInst := range inst.Children() {
		if opt, ok := childInst.(*OptionInstance); ok {
			opt.UpdateSelected(inst.selected, inst.selecteds)
		}
	}
}
```

### 3.4 OptionInstance 更新状态

文件：`ui/components/optiongroup/option.go`

```go
// UpdateSelected 从父实例更新选中状态
func (inst *OptionInstance) UpdateSelected(selected string, selecteds []string) {
	var isSelected bool

	switch inst.mode {
	case ModeSingle:
		isSelected = (selected == inst.value)

	case ModeMultiple:
		isSelected = containsString(selecteds, inst.value)
	}

	// 更新状态
	inst.state.Selected = isSelected
}
```

---

## 🌊 四、其他组件 Intent 示例

### 4.1 Form + Field 示例

```go
// ===== Intent 定义 =====

type FieldChangeIntent struct {
	FormKey string
	Name    string
	Value   string
}

func (FieldChangeIntent) Type() string {
	return "Form:FieldChange"
}

type FormSubmitIntent struct {
	FormKey string
	Data    map[string]string
}

func (FormSubmitIntent) Type() string {
	return "Form:Submit"
}

// ===== Field 发射 Intent =====

type FieldInstance struct {
	instance.BaseInstance
	formKey string
	name    string
	value   string
}

func (inst *FieldInstance) HandleAction(act *action.Action) bool {
	if act.Type == action.ActionInput {
		// 发射 FieldChangeIntent
		intent.Emit(inst, FieldChangeIntent{
			FormKey: inst.formKey,
			Name:    inst.name,
			Value:   inst.value,
		})
		return true
	}
	return false
}

// ===== Form 处理 Intent =====

type FormInstance struct {
	instance.BaseInstance
	key   string
	fields map[string]*FieldInstance
}

func (inst *FormInstance) HandleIntent(i intent.Intent) bool {
	switch v := i.(type) {
	case FieldChangeIntent:
		if v.FormKey == inst.key {
			inst.UpdateField(v.Name, v.Value)
			return true
		}
		return false

	case FormSubmitIntent:
		if v.FormKey == inst.key {
			inst.Submit(v.Data)
			return true
		}
		return false
	}
	return false
}
```

### 4.2 Menu 示例

```go
// ===== Intent 定义 =====

type MenuSelectIntent struct {
	MenuKey string
	ItemKey string
}

func (MenuSelectIntent) Type() string {
	return "Menu:Select"
}

type MenuCloseIntent struct {
	MenuKey string
}

func (MenuCloseIntent) Type() string {
	return "Menu:Close"
}

// ===== MenuItem 发射 Intent =====

func (inst *MenuItemInstance) HandleAction(act *action.Action) bool {
	if act.Type == action.ActionClick {
		intent.Emit(inst, MenuSelectIntent{
			MenuKey: inst.menuKey,
			ItemKey: inst.itemKey,
		})
		return true
	}
	return false
}

// ===== Menu 处理 Intent =====

func (inst *MenuInstance) HandleIntent(i intent.Intent) bool {
	switch v := i.(type) {
	case MenuSelectIntent:
		if v.MenuKey == inst.key {
			inst.SelectItem(v.ItemKey)
			return true
		}
		return false

	case MenuCloseIntent:
		if v.MenuKey == inst.key {
			inst.Close()
			return true
		}
		return false
	}
	return false
}
```

---

## 🧪 五、单元测试

### 5.1 Intent Core 测试

文件：`runtime/intent/intent_test.go`

```go
package intent

import (
	"testing"
)

// ===== Test Intent 定义 =====

type TestIntent struct {
	typ string
	key string
}

func (i TestIntent) Type() string {
	return i.typ
}

func (i TestIntent) Key() string {
	return i.key
}

// ===== Mock Handler =====

type MockHandler struct {
	handled bool
	value   string
}

func (h *MockHandler) HandleIntent(i Intent) bool {
	if i.Type() == "test" {
		h.handled = true
		h.value = i.Key()
		return true // 停止冒泡
	}
	return false // 继续冒泡
}

// ===== 测试用例 =====

func TestEmit_Simple(t *testing.T) {
	handler := &MockHandler{}
	tree := buildTestTree(handler)

	intent.Emit(tree.root, TestIntent{
		typ: "test",
		key: "test-value",
	})

	if !handler.handled {
		t.Error("handler should be called")
	}
	if handler.value != "test-value" {
		t.Errorf("expected 'test-value', got '%s'", handler.value)
	}
}

func TestEmit_Bubble(t *testing.T) {
	childHandler := &MockHandler{}
	parentHandler := &MockHandler{}

	tree := buildTestTreeWithHandlers(parentHandler, childHandler)

	intent.Emit(tree.child, TestIntent{
		typ: "test",
		key: "test-value",
	})

	// 子处理器应该先被调用
	if !childHandler.handled {
		t.Error("child handler should be called")
	}

	// 父处理器不应该被调用（子处理器已处理）
	if parentHandler.handled {
		t.Error("parent handler should not be called (child handled it)")
	}
}

func TestEmit_StopPropagation(t *testing.T) {
	handler1 := &MockHandler{}
	handler2 := &MockHandler{}
	handler3 := &MockHandler{}

	tree := buildThreeLevelTree(handler1, handler2, handler3)

	intent.Emit(tree.bottom, TestIntent{
		typ: "test",
		key: "test-value",
	})

	// 底层处理器应该被调用
	if !handler3.handled {
		t.Error("bottom handler should be called")
	}

	// 中层和顶层不应该被调用
	if handler2.handled || handler1.handled {
		t.Error("middle and top handlers should not be called")
	}
}

// ===== Helper Functions =====

type testTree struct {
	root   *mockInstance
	mid    *mockInstance
	bottom *mockInstance
	child  *mockInstance
}

type mockInstance struct {
	instance.BaseInstance
	handler IntentHandler
}

func (m *mockInstance) HandleIntent(i Intent) bool {
	if m.handler != nil {
		return m.handler.HandleIntent(i)
	}
	return false
}

func buildTestTree(handler IntentHandler) testTree {
	root := &mockInstance{
		BaseInstance: instance.BaseInstance{},
		handler:      handler,
	}
	child := &mockInstance{
		BaseInstance: instance.BaseInstance{},
	}

	root.AddChild(child)

	return testTree{
		root:  root,
		child: child,
	}
}

func buildTestTreeWithHandlers(parentHandler, childHandler IntentHandler) testTree {
	parent := &mockInstance{
		BaseInstance: instance.BaseInstance{},
		handler:      parentHandler,
	}
	child := &mockInstance{
		BaseInstance: instance.BaseInstance{},
		handler:      childHandler,
	}

	parent.AddChild(child)

	return testTree{
		root:  parent,
		child: child,
	}
}

func buildThreeLevelTree(handler1, handler2, handler3 IntentHandler) testTree {
	top := &mockInstance{
		BaseInstance: instance.BaseInstance{},
		handler:      handler1,
	}
	mid := &mockInstance{
		BaseInstance: instance.BaseInstance{},
		handler:      handler2,
	}
	bottom := &mockInstance{
		BaseInstance: instance.BaseInstance{},
		handler:      handler3,
	}

	top.AddChild(mid)
	mid.AddChild(bottom)

	return testTree{
		root:   top,
		mid:    mid,
		bottom: bottom,
	}
}
```

### 5.2 OptionGroup Intent 测试

文件：`ui/components/optiongroup/intent_test.go`

```go
package optiongroup

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/intent"
)

func TestOptionGroup_SelectIntent(t *testing.T) {
	options := []Option{
		{Value: "opt1", Label: "Option 1"},
		{Value: "opt2", Label: "Option 2"},
	}

	vnode := NewBuilder(options).
		Key("test-group").
		Build()

	groupInst := vnode.CreateInstance()
	insts := groupInst.Children()

	// 发射 Intent
	intent.Emit(insts[0].(*OptionInstance), OptionSelectIntent{
		GroupKey:   "test-group",
		Value:      "opt1",
		IsSelected: true,
	})

	// 验证选中状态
	if groupInst.selected != "opt1" {
		t.Errorf("expected 'opt1', got '%s'", groupInst.selected)
	}
}

func TestOptionGroup_MultiSelectIntent(t *testing.T) {
	options := []Option{
		{Value: "opt1", Label: "Option 1"},
		{Value: "opt2", Label: "Option 2"},
	}

	vnode := NewBuilder(options).
		Key("test-group").
		Mode(ModeMultiple).
		Build()

	groupInst := vnode.CreateInstance()

	// 选中选项 1
	intent.Emit(groupInst.Children()[0].(*OptionInstance), OptionSelectIntent{
		GroupKey:   "test-group",
		Value:      "opt1",
		IsSelected: true,
	})

	// 选中选项 2
	intent.Emit(groupInst.Children()[1].(*OptionInstance), OptionSelectIntent{
		GroupKey:   "test-group",
		Value:      "opt2",
		IsSelected: true,
	})

	// 验证多选状态
	if len(groupInst.selecteds) != 2 {
		t.Errorf("expected 2 selected, got %d", len(groupInst.selecteds))
	}
}

func TestOptionGroup_UnhandledIntent(t *testing.T) {
	options := []Option{
		{Value: "opt1", Label: "Option 1"},
	}

	vnode := NewBuilder(options).
		Key("test-group").
		Build()

	groupInst := vnode.CreateInstance()

	// 发射不同组的 Intent（应该被忽略）
	intent.Emit(groupInst.Children()[0].(*OptionInstance), OptionSelectIntent{
		GroupKey:   "other-group", // 不同的 key
		Value:      "opt1",
		IsSelected: true,
	})

	// 验证未选中
	if groupInst.selected != "" {
		t.Errorf("expected no selection, got '%s'", groupInst.selected)
	}
}
```

---

## ✅ 六、验证检查清单

### 6.1 编译检查

```bash
cd E:\projects\yao\wwsheng009\mint
go build ./runtime/intent/...
```

**预期结果**:
- ✅ 无编译错误
- ✅ 接口实现正确

### 6.2 单元测试

```bash
# 测试 Intent Core
go test -v ./runtime/intent/...

# 测试 OptionGroup Intent
go test -v ./ui/components/optiongroup/...
```

**预期结果**:
- ✅ 所有测试通过
- ✅ 冒泡机制正确
- ✅ 停止传播正确

### 6.3 集成测试

```bash
# 运行 OptionGroup 示例
cd examples/radiogroup_demo
go run main.go
```

**手动测试**:
- [ ] 应用正常启动
- [ ] 点击选项发射 Intent
- [ ] OptionGroup 接收并处理 Intent
- [ ] 单选切换正常
- [ ] 多选累加/移除正常

### 6.4 性能测试

```bash
# 基准测试
cd runtime/intent
go test -bench=. -benchmem .
```

**预期结果**:
- `BenchmarkEmit`: ~100-500 ns/op
- `BenchmarkHandleIntent`: ~10-50 ns/op
- 无内存泄漏

---

## 🐛 七、常见问题

### 问题 1: Intent 导致无限循环

**症状**:
应用卡死，CPU 占用 100%

**原因**:
Intent 处理器中发射了新的 Intent。

**解决方案**:
```go
// 错误示例
func (inst *Instance) HandleIntent(i Intent) bool {
	if i.Type() == "Select" {
		// 在处理过程中发射新 Intent → 无限循环
		intent.Emit(inst, SomeOtherIntent{})
		return true
	}
}

// 正确示例：异步处理
func (inst *Instance) HandleIntent(i Intent) bool {
	if i.Type() == "Select" {
		// 延迟到下一个事件循环
		go func() {
			intent.Emit(inst, SomeOtherIntent{})
		}()
		return true
	}
}
```

### 问题 2: Intent 未被处理

**症状**:
发射的 Intent 没有任何反应。

**原因**:
冒泡到根仍未被处理。

**解决方案**:
```go
// 确保实现了 IntentHandler
var _ IntentHandler = (*Instance)(nil)

func (inst *Instance) HandleIntent(i Intent) bool {
	// 必须返回正确的布尔值
	if i.Type() == "MyType" {
		return true // 已处理
	}
	return false // 继续冒泡
}

// 添加日志调试
func (inst *Instance) HandleIntent(i Intent) bool {
	fmt.Printf("[DEBUG] %s.HandleIntent: %v\n", inst.key, i.Type())
	// ...
}
```

### 问题 3: Intent 中的数据被修改

**症状**:
Intent 数据在传递过程中被意外修改。

**原因**:
Intent 结构中有指针字段，被多个组件共享。

**解决方案**:
```go
// 错误示例：使用指针
type MyIntent struct {
	Data *map[string]string // ❌ 易变
}

// 正确示例：使用值
type MyIntent struct {
	Data map[string]string // ✅ 副本
}

// 或者使用不可变类型
type MyIntent struct {
	Data []Field // ⭐ 字段不可变
}
```

---

## 📚 八、参考资料

### 设计参考

1. **DOM Event System**
   - https://developer.mozilla.org/en-US/docs/Web/API/Event/bubbles
   - 事件冒泡和捕获

2. **React Synthetic Events**
   - https://react.dev/reference/react-dom/common-events
   - 合成事件系统

3. **BLoC Pattern**
   - https://bloclibrary.dev/
   - Intent/Event 模式

### Mint 相关

- `runtime/intent/intent.go` - Intent 接口
- `runtime/intent/bubble.go` - 冒泡机制
- `ui/components/optiongroup/intent.go` - OptionGroup Intent
- [Phase 1: Instance Tree](./INSTANCE_TREE_IMPLEMENTATION.md)
- [Phase 2: Context System](./CONTEXT_IMPLEMENTATION.md)

---

## 🎯 九、总结

### Mint Runtime 2.0 三大通信能力

完成 Phase 1-3 后，Mint 将具备完整的组件通信能力：

| 能力 | Phase | 用途 |
|------|-------|------|
| **Instance Tree** | Phase 1 | 父子关系、生命周期 |
| **Context System** | Phase 2 | 依赖注入、跨层级访问 |
| **Intent Bubble** | Phase 3 | 事件冒泡、解耦通信 |

### 组合使用示例

```go
// Form 提供 Context + 处理 Intent
type FormInstance struct {
	instance.BaseInstance
}

func (inst *FormInstance) OnMount() {
	// 方式1: 提供到 Context
	fContext.Provide(FormContext, inst)
}

func (inst *FormInstance) HandleIntent(i Intent) bool {
	// 方式2: 处理 Intent
	if _, ok := i.(FieldChangeIntent); ok {
		// ...
		return true
	}
	return false
}

// Field 消费 Context + 发射 Intent
type FieldInstance struct {
	instance.BaseInstance
}

func (inst *FieldInstance) OnMount() {
	// 方式1: 从 Context 获取
	form := UseContext[*FormInstance](FormContext)
	// ...
}

func (inst *FieldInstance) HandleAction(act Action) bool {
	// 方式2: 发射 Intent
	intent.Emit(inst, FieldChangeIntent{...})
	return true
}
```

---

**文档状态**: ✅ 设计完成，待实施

**完成状态**: 🎉 Phase 1-3 架构设计全部完成

**下一步**: 开始实施 Phase 1（Instance Tree）
