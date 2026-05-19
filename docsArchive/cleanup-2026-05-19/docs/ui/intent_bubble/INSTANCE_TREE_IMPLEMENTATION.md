# Instance Tree 实施方案（Phase 1）

> **目标**: 建立组件逻辑树，解决父子关系问题
> **预计工作量**: 2天
> **优先级**: 🔴 P0
> **依赖**: 无

---

## 📋 执行摘要

Mint 当前只有 **Fiber 调度树**，但缺少 **Instance 逻辑树**。这导致：

- ❌ Instance 之间没有父子关系
- ❌ 子组件无法访问父组件
- ❌ 必须依赖 hack（全局注册表、闭包）

Phase 1 的目标是建立 Instance Tree，为后续的 Context 和 Intent Bubble 打下基础。

---

## 🏗️ 一、核心设计

### 1.1 Instance Tree 概念

```
Fiber Tree (调度树)
├─ Fiber (OptionGroup)
│  ├─ Fiber (Option 1)
│  ├─ Fiber (Option 2)
│  └─ Fiber (Option 3)
     ↓ Fiber.Instance
Instance Tree (逻辑树)
├─ OptionGroupInstance
│  ├─ OptionInstance 1
│  ├─ OptionInstance 2
│  └─ OptionInstance 3
```

### 1.2 核心接口

```go
// runtime/ui/instance.go
package instance

// ComponentInstance 组件实例接口
type ComponentInstance interface {
    // ===== 组件树关系 =====
    Parent() ComponentInstance
    Children() []ComponentInstance

    // ===== 生命周期 =====
    OnMount()
    OnUpdate()
    OnUnmount()

    // ===== 渲染 =====
    Layout(ctx LayoutContext)
    Paint(ctx PaintContext) []paint.DrawCmd

    // ===== 意图处理（Phase 3） =====
    HandleIntent(intent.Intent) bool
}

// AddChilder 添加子组件能力
type AddChilder interface {
    AddChild(ComponentInstance)
}
```

```go
// runtime/ui/instance.go
package instance

import "github.com/wwsheng009/mint/runtime/ui"

// BaseInstance 基础实例实现
type BaseInstance struct {
    // 组件树关系
    parent   ComponentInstance
    children []ComponentInstance

    // 生命周期状态
    mounted  bool
    dirty    bool

    // Fiber 引用
    fiber *fiber.Fiber
}

// Parent 获取父组件
func (b *BaseInstance) Parent() ComponentInstance {
    return b.parent
}

// Children 获取子组件列表
func (b *BaseInstance) Children() []ComponentInstance {
    return b.children
}

// AddChild 添加子组件（由 Fiber 调用）
func (b *BaseInstance) AddChild(c ComponentInstance) {
    b.children = append(b.children, c)

    // 如果子实例也是 BaseInstance，设置父引用
    if child, ok := c.(*BaseInstance); ok {
        child.parent = b
    }
}

// RemoveChild 移除子组件
func (b *BaseInstance) RemoveChild(c ComponentInstance) {
    for i, child := range b.children {
        if child == c {
            b.children = append(b.children[:i], b.children[i+1:]...)

            if childInst, ok := child.(*BaseInstance); ok {
                childInst.parent = nil
            }
            break
        }
    }
}

// SetFiber 设置 Fiber 引用
func (b *BaseInstance) SetFiber(f *fiber.Fiber) {
    b.fiber = f
}

// GetFiber 获取 Fiber 引用
func (b *BaseInstance) GetFiber() *fiber.Fiber {
    return b.fiber
}

// ===== 默认生命周期实现（可覆盖） =====

func (b *BaseInstance) OnMount() {
    b.mounted = true
    b.dirty = true
}

func (b *BaseInstance) OnUpdate() {
    b.dirty = true
}

func (b *BaseInstance) OnUnmount() {
    b.mounted = false
}
```

---

## 🔧 二、修改 Fiber 创建逻辑

### 2.1 修改 `CreateFiber`

文件：`runtime/ui/fiber_util.go`

```go
// CreateFiber 创建 Fiber 节点
func CreateFiber(vnode VNode, parent *Fiber) *Fiber {
	// ... 原有逻辑

	f := &Fiber{
		VNode:      vnode,
		Tag:        vnode.Tag(),
		Type:       vnode.Type(),
		Key:        vnode.Key(),
		Parent:     parent,
		EffectTag:  EffectTagPlacement,
		Props:      vnode.Props(),
		Alternate:  nil,
		Instance:   nil, // 稍后设置
	}

	// ... 原有逻辑

	return f
}
```

### 2.2 添加 `mountInstance`

文件：`runtime/ui/fiber_util.go`（新增）

```go
// mountInstance 挂载组件实例
func mountInstance(f *Fiber) {
	// 创建实例
	inst := f.VNode.CreateInstance()
	f.Instance = inst

	// 建立父子关系
	if f.Parent != nil && f.Parent.Instance != nil {
		parentInst := f.Parent.Instance

		// 检查父实例是否实现了 AddChilder
		if parent, ok := parentInst.(AddChilder); ok {
			parent.AddChild(inst)
		}

		// 如果实例是 BaseInstance，也设置 Fiber 引用
		if base, ok := inst.(*instance.BaseInstance); ok {
			base.SetFiber(f)
		}
	} else {
		// 根节点也设置 Fiber 引用
		if base, ok := inst.(*instance.BaseInstance); ok {
			base.SetFiber(f)
		}
	}

	// 触发 OnMount
	inst.OnMount()
}
```

### 2.3 修改 `reconcileChildren`

文件：`internal/reconciler/reconciler.go`

```go
func reconcileChildren(returnFiber, newChildVNode VNode) *Fiber {
	if newChildVNode == nil {
		return nil
	}

	// 创建子 Fiber
	childFiber := CreateFiber(newChildVNode, returnFiber)

	// 挂载子实例
	mountInstance(childFiber)

	return childFiber
}
```

### 2.4 添加 `unmountInstance`

文件：`runtime/ui/fiber_util.go`（新增）

```go
// unmountInstance 卸载组件实例
func unmountInstance(f *Fiber) {
	if f.Instance == nil {
		return
	}

	// 先递归卸载所有子组件
	child := f.Child
	for child != nil {
		unmountInstance(child)
		child = child.Sibling
	}

	// 从父实例的 children 列表中移除
	if f.Parent != nil && f.Parent.Instance != nil {
		parentInst := f.Parent.Instance
		if parentRemover, ok := parentInst.(interface{ RemoveChild(ComponentInstance) }); ok {
			parentRemover.RemoveChild(f.Instance)
		}
	}

	// 触发 OnUnmount
	f.Instance.OnUnmount()

	// 清理引用
	f.Instance = nil
}
```

---

## 🔄 三、更新现有组件

### 3.1 OptionGroup 组件

文件：`ui/components/optiongroup/instance.go`

```go
package optiongroup

import (
	"github.com/wwsheng009/mint/runtime/instance"
	// ... 其他导入
)

// Instance OptionGroup 组件实例
type Instance struct {
	// 组合 BaseInstance
	instance.BaseInstance

	// 原有字段保持不变
	key       string
	label     string
	options   []Option
	selected  string
	selecteds []string
	// ...
}

// NewInstance 创建实例
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		// ... 原有初始化逻辑
	}
	return inst
}

// SelectOption 选择选项（原方法保持不变）
func (inst *Instance) SelectOption(value string) {
	// ... 原有逻辑
}

// ===== 不再需要 hack 方法 =====
// 之前的 updateOptionCallbacks、registerParent 等都删除

// 现在可以通过 Children() 直接访问子实例
func (inst *Instance) UpdateChildrenState() {
	for _, childInst := range inst.Children() {
		if opt, ok := childInst.(*OptionInstance); ok {
			opt.UpdateSelected(inst.selected, inst.selecteds)
		}
	}
}
```

### 3.2 Option 子组件

文件：`ui/components/optiongroup/option.go`

```go
package optiongroup

import (
	"github.com/wwsheng009/mint/runtime/instance"
	// ... 其他导入
)

type OptionInstance struct {
	// 组合 BaseInstance
	instance.BaseInstance

	// 原有字段
	key  string
	idx  int
	value string
	label string
	// ...

	// ===== 不再需要 =====
	// selectFunc SelectOptionFunc  ← 删除
	// parentKey  string            ← 删除
}

// NewOptionInstance 创建实例
func NewOptionInstance(props rtui.Props) *OptionInstance {
	inst := &OptionInstance{
		// ... 原有初始化逻辑
	}
	return inst
}

// ===== 访问父组件 =====
func (inst *OptionInstance) GetParentGroup() (*Instance, bool) {
	if inst.Parent() == nil {
		return nil, false
	}

	group, ok := inst.Parent().(*Instance)
	return group, ok
}

// HandleAction 处理动作（修改后不再依赖 selectFunc）
func (inst *OptionInstance) HandleAction(act *action.Action) bool {
	switch act.Type {
	case action.ActionClick, action.ActionEnter, action.ActionToggle:
		// 通过 Parent() 获取父实例
		group, ok := inst.GetParentGroup()
		if ok {
			group.SelectOption(inst.value)
		}
		return true
	}
	return false
}
```

### 3.3 Panel 组件

文件：`ui/components/panel/instance.go`

```go
package panel

import (
	"github.com/wwsheng009/mint/runtime/instance"
	// ...
)

type Instance struct {
	instance.BaseInstance
	// ... 原有字段
}

// 不需要额外代码，AddChild 会自动工作
```

---

## 🧪 四、单元测试

### 4.1 BaseInstance 测试

文件：`runtime/ui/instance_tree_test.go`

```go
package instance

import (
	"testing"
)

// MockComponentInstance 用于测试
type MockComponentInstance struct {
	BaseInstance
	name string
}

func TestBaseInstance_TreeRelationships(t *testing.T) {
	parent := &MockComponentInstance{name: "parent"}
	child1 := &MockComponentInstance{name: "child1"}
	child2 := &MockComponentInstance{name: "child2"}

	// 测试 AddChild
	parent.AddChild(child1)
	parent.AddChild(child2)

	// 验证父关系
	if child1.Parent() != parent {
		t.Error("child1.Parent() should be parent")
	}
	if child2.Parent() != parent {
		t.Error("child2.Parent() should be parent")
	}

	// 验证子列表
	children := parent.Children()
	if len(children) != 2 {
		t.Errorf("expected 2 children, got %d", len(children))
	}
}

func TestBaseInstance_RemoveChild(t *testing.T) {
	parent := &MockComponentInstance{name: "parent"}
	child1 := &MockComponentInstance{name: "child1"}
	child2 := &MockComponentInstance{name: "child2"}

	parent.AddChild(child1)
	parent.AddChild(child2)

	// 移除 child1
	parent.RemoveChild(child1)

	// 验证移除
	children := parent.Children()
	if len(children) != 1 {
		t.Errorf("expected 1 child after removal, got %d", len(children))
	}
	if children[0] != child2 {
		t.Error("remaining child should be child2")
	}

	// 验证父引用被清理
	if child1.Parent() != nil {
		t.Error("removed child parent should be nil")
	}
}
```

### 4.2 Fiber 集成测试

文件：`runtime/ui/instance_tree_test.go`

```go
package ui

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/instance"
)

func TestInstanceTree_Building(t *testing.T) {
	// 创建 VNode 树
	childVNode1 := NewText("Child 1").Key("child1")
	childVNode2 := NewText("Child 2").Key("child2")
	parentVNode := NewStack().
		Children(childVNode1, childVNode2).
		Key("parent")

	// 创建 Fiber 树
	root := CreateFiberFromVNode(parentVNode)

	// 验证 Instance 树
	parentInst := root.Instance
	if parentInst == nil {
		t.Fatal("root Instance should not be nil")
	}

	// 验证子实例
	childInsts := parentInst.Children()
	if len(childInsts) != 2 {
		t.Errorf("expected 2 child instances, got %d", len(childInsts))
	}

	// 验证父关系
	for _, child := range childInsts {
		if child.Parent() != parentInst {
			t.Error("child parent should be root instance")
		}
	}
}

func TestInstanceTree_Unmount(t *testing.T) {
	// ... 卸载测试
}
```

### 4.3 OptionGroup 集成测试

文件：`runtime/ui/instance_tree_test.go`

```go
package optiongroup

import (
	"testing"
)

func TestOptionGroup_InstanceTree(t *testing.T) {
	options := []Option{
		{Value: "opt1", Label: "Option 1"},
		{Value: "opt2", Label: "Option 2"},
	}

	vnode := NewBuilder(options).
		Key("test-group").
		Build()

	inst := vnode.CreateInstance()

	// 验证子实例
	children := inst.Children()
	if len(children) != 2 {
		t.Errorf("expected 2 child instances, got %d", len(children))
	}

	// 验证子实例类型
	for i, child := range children {
		if _, ok := child.(*OptionInstance); !ok {
			t.Errorf("child %d should be *OptionInstance", i)
		}

		// 验证父关系
		if child.Parent() != inst {
			t.Error("child parent should be group instance")
		}
	}
}

func TestOptionInstance_ParentLookup(t *testing.T) {
	options := []Option{
		{Value: "opt1", Label: "Option 1"},
	}

	vnode := NewBuilder(options).Build()
	groupInst := vnode.CreateInstance()

	childInst := groupInst.Children()[0].(*OptionInstance)

	// 测试 GetParentGroup
	group, ok := childInst.GetParentGroup()
	if !ok {
		t.Error("GetParentGroup should succeed")
	}
	if group != groupInst {
		t.Error("GetParentGroup should return correct instance")
	}
}
```

---

## ✅ 五、验证检查清单

### 5.1 编译检查

```bash
cd E:\projects\yao\wwsheng009\mint
go build ./runtime/...
go build ./ui/components/...
```

**预期结果**:
- ✅ 无编译错误
- ✅ 无未使用的导入警告

### 5.2 单元测试

```bash
# 测试 BaseInstance
go test ./runtime/ui/...

# 测试 Fiber 集成
go test ./runtime/ui/...

# 测试 OptionGroup
go test ./ui/components/optiongroup/...
```

**预期结果**:
- ✅ 所有测试通过
- ✅ 测试覆盖率 > 90%

### 5.3 集成测试

```bash
# 运行 OptionGroup 示例
cd examples/multiselect_demo
go run main.go
```

**手动测试**:
- [ ] 应用正常启动
- [ ] 点击选项可以选中（通过 Parent() 调用）
- [ ] Tab 键导航正常
- [ ] 键盘 Enter/Space 提交正常

### 5.4 性能测试

```bash
# 基准测试
go test -bench=. -benchmem ./runtime/ui/...
go test -bench=. -benchmem ./ui/components/optiongroup/...
```

**预期结果**:
- `BenchmarkAddChild`: ~10-50 ns/op
- `BenchmarkParentLookup`: ~1-10 ns/op
- 无内存泄漏

---

## 📝 六、清理工作

### 6.1 删除冗余代码

从 `ui/components/optiongroup/` 删除：

1. **option.go**:
   - 第 1-58 行: 全局注册表代码 (`registerParent`, `lookupParent` 等)
   - 删除 `parentKey` 字段
   - 删除导入 `"sync"`

2. **instance.go**:
   - 删除 `vnode *VNode` 字段
   - 删除 `childInstances []*OptionInstance` 字段
   - 删除 `updateOptionCallbacks()` 方法

3. **vnode.go**:
   - 删除 `parentCallback` 相关代码

### 6.2 清理测试

从 `ui/components/optiongroup/optiongroup_test.go` 中删除相关测试。

---

## 🐛 七、常见问题

### 问题 1: 循环引用导致内存泄漏

**症状**:
长时间运行后内存占用持续增长。

**原因**:
Parent 和 Children 形成循环引用，GC 无法回收。

**解决方案**:
Go 的 GC 可以正确处理双向引用。但为了安全：

```go
func (b *BaseInstance) RemoveChild(c ComponentInstance) {
	// 先清空子引用
	if child, ok := c.(*BaseInstance); ok {
		child.parent = nil
	}

	// 再从数组中移除
	for i, child := range b.children {
		if child == c {
			b.children = append(b.children[:i], b.children[i+1:]...)
			break
		}
	}
}
```

### 问题 2: 子组件在 AddChild 时未初始化

**症状**:
`AddChild` 调用时出现 panic。

**原因**:
子实例还未完全初始化。

**解决方案**:
确保 `mountInstance` 的调用顺序：

```go
func mountInstance(f *Fiber) {
	// 1. 创建实例
	inst := f.VNode.CreateInstance()

	// 2. 设置到 Fiber
	f.Instance = inst

	// 3. 建立父子关系（现在 inst 已经完全初始化）
	if f.Parent != nil {
		// ...
	}

	// 4. 触发生命周期
	inst.OnMount()
}
```

### 问题 3: 并发安全问题

**症状**:
多线程环境下出现数据竞争。

**原因**:
`AddChild` 和 `RemoveChild` 可能被并发调用。

**解决方案**:
虽然 Mint 当前是单线程调度，但为了未来安全：

```go
import "sync"

type BaseInstance struct {
	parent   ComponentInstance
	children []ComponentInstance

	mu sync.RWMutex
}

func (b *BaseInstance) AddChild(c ComponentInstance) {
	b.mu.Lock()
	defer b.mu.Unlock()
	// ...
}

func (b *BaseInstance) Children() []ComponentInstance {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.children
}
```

---

## 📚 八、参考资料

### 设计参考

1. **React Fiber Architecture**
   - https://react.dev/learn/understanding-reacts-rendering
   - Fiber tree → Component Instance 映射

2. **Flutter RenderObject**
   - https://docs.flutter.dev/resources/architectural-overview
   - Widget → Element → RenderObject

3. **SwiftUI View Hierarchy**
   - https://developer.apple.com/documentation/swiftui/view-hierarchy
   - 树形结构和生命周期管理

### Mint 相关

- `runtime/ui/fiber.go` - Fiber 结构
- `runtime/ui/vnode.go` - VNode 接口
- `ui/components/optiongroup/` - OptionGroup 实现

---

## 🎯 九、下一步

### Phase 2: Context System

完成 Instance Tree 后，可以实施：

- [ ] 定义 `ContextKey` 和 `FiberContext`
- [ ] 集成到 Fiber
- [ ] 实现上下文提供者和消费者

Phase 2 将解决 Props Drilling 问题，让跨层级通信更加优雅。

---

**文档状态**: ✅ 设计完成，待实施

**预计开始日期**: 2026-03-XX
