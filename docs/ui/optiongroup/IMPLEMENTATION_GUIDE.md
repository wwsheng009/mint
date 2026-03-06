# OptionGroup 实施指南

> **更新日期**: 2026-03-06
> **版本**: 1.0
> **状态**: 🚧 实施准备就绪

---

## 🎯 实施目标

修复 OptionGroup 组件"子选项无法选中"的问题，采用闭包包装（Closure Wrapping）方案，确保每个选项能够正确响应鼠标点击和键盘操作。

---

## 📦 准备工作

### 环境验证

在开始实施前，确保以下环境正常：

```bash
# 1. 验证 Go 版本
go version  # 应该是 Go 1.21+

# 2. 验证项目路径
cd E:\projects\yao\wwsheng009\mint
pwd        # 确认在项目根目录

# 3. 运行现有测试
go test ./ui/components/optiongroup/...
# 应该显示 PASS（虽然功能有问题，但测试应该通过）
```

### 创建备份分支（可选）

```bash
git checkout -b optiongroup-fix-closure

# 如果需要保留当前状态
git add .
git commit -m "WIP: OptionGroup closure fix preparation"
```

---

## 🔧 Phase 1: 实施闭包方案

### 目标
使用闭包包装解决父子回调传递问题。

### 文件修改清单

| 文件 | 修改内容 | 删除行数 | 新增行数 |
|------|---------|---------|---------|
| `vnode.go` | 修改 `Children()`，使用闭包包装 | 0 | ~5 |
| `vnode.go` | 修改 `CreateInstance()`，使用闭包包装 | 0 | ~3 |
| **总计** | | **0** | **~8** |

### 详细实施步骤

#### 步骤 1.1: 修改 `OptionGroup.VNode.Children()`

文件：`ui/components/optiongroup/vnode.go`
位置：第 135-158 行

**原代码**：
```go
// Children returns child nodes - the options as individual OptionVNodes.
func (o *VNode) Children() []rtui.VNode {
	if o.options == nil {
		return nil
	}
	children := make([]rtui.VNode, len(o.options))
	for i, opt := range o.options {
		child := NewOptionVNodeDeferred(opt.Value, opt.Label, i, o.mode)
		// Apply parent disabled state to child
		if o.disabled {
			child.SetDisabled(true)
		}
		// Apply the selectFunc if it's been set by the parent Instance
		if o.optionSelectFunc != nil {
			child.SetSelectFunc(o.optionSelectFunc)
		}
		// Pass parentCallback as a prop for later use
		child.SetProps(rtui.Props{"parentCallback": o.optionSelectFunc})
		children[i] = child
	}
	return children
}
```

**修改后**：
```go
// Children returns child nodes - the options as individual OptionVNodes.
func (o *VNode) Children() []rtui.VNode {
	if o.options == nil {
		return nil
	}
	children := make([]rtui.VNode, len(o.options))
	for i, opt := range o.options {
		child := NewOptionVNodeDeferred(opt.Value, opt.Label, i, o.mode)
		// Apply parent disabled state to child
		if o.disabled {
			child.SetDisabled(true)
		}
		
		// ⭐ 使用闭包包装，延迟查找 o.optionSelectFunc
		// 闭包捕获 VNode 的引用，在执行时查找最新的回调
		child.SetSelectFunc(func(value string) {
			if o.optionSelectFunc != nil {
				o.optionSelectFunc(value)
			}
		})
		
		// 通过 Props 传递闭包（子实例从这里获取）
		child.SetProps(rtui.Props{
			"value":      opt.Value,
			"label":      opt.Label,
			"disabled":   o.disabled,
			"mode":       o.mode,
			"idx":        i,
			"selectFunc": func(value string) {
				if o.optionSelectFunc != nil {
					o.optionSelectFunc(value)
				}
			},
		})
		
		children[i] = child
	}
	return children
}
```

**关键变化**：
1. ✅ 移除了 `if o.optionSelectFunc != nil` 检查（闭包内部检查）
2. ✅ 增加了闭包包装 `func(value string) { ... }`
3. ✅ 通过 Props 传递闭包（确保子实例能够获取）

---

#### 步骤 1.2: 确认 `OptionGroup.VNode.CreateInstance()`

文件：`ui/components/optiongroup/vnode.go`
位置：第 265-280 行

**验证代码**（无需修改，但需要确认存在）：
```go
// CreateInstance creates a new OptionGroup Instance from this VNode description.
func (o *VNode) CreateInstance() rtui.ComponentInstance {
	props := rtui.Props{
		"key":          o.key,
		"label":        o.label,
		"style":        o.style,
		"selectIntent": o.selectIntent,
		"disabled":     o.disabled,
		"mode":         o.mode,
		"options":      o.options,
		"selected":     o.selected,
		"selecteds":    o.selecteds,
		"orientation":  o.orientation,
		"spacing":      o.spacing,
	}
	inst := NewInstance(props)

	// Set the callback so child OptionVNodes can select options
	o.optionSelectFunc = inst.SelectOption

	return inst
}
```

**确认要点**：
- ✅ `o.optionSelectFunc = inst.SelectOption` 在 `NewInstance()` 之后执行
- ✅ 这确保了 `CreateInstance()` 在 `Children()` 之前设置回调

---

### 验证闭包方案

#### 测试 1: 单元测试

```bash
cd E:\projects\yao\wwsheng009\mint
go test ./ui/components/optiongroup/... -v
```

**预期结果**：
```
=== RUN   TestNewBuilder
--- PASS: TestNewBuilder (0.00s)
=== RUN   TestVNodeTag
--- PASS: TestVNodeTag (0.00s)
=== RUN   TestVNodeKey
--- PASS: TestVNodeKey (0.00s)
...
PASS
ok      github.com/wwsheng009/mint/ui/components/optiongroup    0.123s
```

#### 测试 2: 编译示例程序

```bash
# 测试 multiselect_demo
cd examples/multiselect_demo
go build -o multiselect_demo.exe

# 测试 typed_intent_demo
cd ../typed_intent_demo
go build -o typed_intent_demo.exe
```

**预期结果**：
- 无编译错误
- 生成可执行文件

---

## 🧹 Phase 2: 清理冗余代码

### 目标
移除方案B（全局注册表）预留的未使用代码，简化代码结构。

### 文件修改清单

| 文件 | 修改内容 | 删除行数 | 新增行数 |
|------|---------|---------|---------|
| `option.go` | 移除全局注册表代码 | ~60 | 0 |
| `instance.go` | 移除 `vnode` 和 `childInstances` 字段 | ~10 | 0 |
| **总计** | | **~70** | **0** |

---

### 步骤 2.1: 清理 `option.go`

文件：`ui/components/optiongroup/option.go`
位置：第 1-60 行

**删除内容**：
```go
// =============================================================================
// Parent Instance Registry
// =============================================================================

// parentRegistry stores mapping from parent key to parent Instance.
var parentRegistry = struct {
	sync.RWMutex
	registry map[string]*Instance
}{
	registry: make(map[string]*Instance),
}

// registerParent registers an OptionGroup Instance by its key.
func registerParent(key string, inst *Instance) {
	if key == "" || inst == nil {
		return
	}
	parentRegistry.Lock()
	parentRegistry.registry[key] = inst
	parentRegistry.Unlock()
}

// unregisterParent unregisters an OptionGroup Instance.
func unregisterParent(key string) {
	if key == "" {
		return
	}
	parentRegistry.Lock()
	delete(parentRegistry.registry, key)
	parentRegistry.Unlock()
}

// lookupParent looks up an OptionGroup Instance by parent key.
func lookupParent(key string) *Instance {
	if key == "" {
		return nil
	}
	parentRegistry.RLock()
	defer parentRegistry.RUnlock()
	return parentRegistry.registry[key]
}
```

**同时移除导入**：
```go
// 移除
- "sync"
```

**修改位置**：删除这些行（代码编辑器会提示删除范围）

---

### 步骤 2.2: 清理 `OptionInstance` 字段

文件：`ui/components/optiongroup/option.go`
位置：第 256-260 行

**移除字段**：
```go
type OptionInstance struct {
	// === Identification ===
	key  string
	idx  int
	value string
	label string

	// === State (synced from parent) ===
	state control.InteractionState

	// === Mode ===
	mode SelectMode

	// === Parent Reference ===
	- selectFunc SelectOptionFunc
	- parentKey  string // ← 删除

	// === Style/Base ===
	optionStyle style.Style
	bounds      [4]int // x, y, w, h
	dirty       bool

	// === Intent Emitter ===
	intentEmitter func(intent.Intent)

	// === Behaviors ===
	behaviors *control.BehaviorList
}
```

**注意**：保留 `selectFunc` 字段，因为 NewOptionInstance 会设置它。

---

### 步骤 2.3: 清理 `OptionGroup.Instance` 字段

文件：`ui/components/optiongroup/instance.go`
位置：第 40-45 行

**移除字段**：
```go
type Instance struct {
	// === Identification ===
	key string

	// === Props (from VNode, may change each render) ===
	label        string
	optionStyle  style.Style
	selectIntent intent.Intent

	// === Runtime State (managed by instance) ===
	state     control.InteractionState
	mode      SelectMode
	options   []Option
	selected  string
	selecteds []string
	bounds    [4]int
	dirty     bool

	// === Intent Emitter ===
	intentEmitter func(intent.Intent)

	- // === VNode Reference (for updating child callbacks) ===
	- vnode *VNode  // ← 删除

	- // === Child Instances (for direct callback setup) ===
	- childInstances []*OptionInstance  // ← 删除

	// === Behaviors ===
	behaviors *control.BehaviorList
}
```

---

### 步骤 2.4: 清理 `updateOptionCallbacks` 方法

文件：`ui/components/optiongroup/instance.go`
位置：第 138-161 行

**删除方法**：
```go
// updateOptionCallbacks updates the selectFunc on all child option instances.
// This is called during OnMount to ensure children have access to the parent's SelectOption method.
func (inst *Instance) updateOptionCallbacks() {
	// Access children through the associated VNode if available
	// In a production system, this would be done via a parent-child reference
	// For now, we'll defer this to another approach

	// Note: The actual implementation would need access to child instances.
	// Since the Fiber system doesn't provide direct parent-to-child instance references,
	// we'll use a different mechanism: update the VNode so future children get the callback.

	// The VNode.optionSelectFunc is already set in CreateInstance,
	// so future renders will have the callback.
}
```

**同时更新 `OnMount`**：
```go
// OnMount implements ComponentInstance.
// After refactoring: Update child option callbacks to ensure they can select values.
func (inst *Instance) OnMount() {
	inst.behaviors.OnMount(inst)

	// ⭐ 移除 updateOptionCallbacks 调用
	// inst.updateOptionCallbacks()  // ← 删除这行
}
```

---

### 步骤 2.5: 同步移除未使用的导入

文件：`ui/components/optiongroup/instance.go`

**导入区域**保持不变（没有新增导入需要清理）

---

### 验证清理结果

#### 测试 1: 编译检查

```bash
cd E:\projects\yao\wwsheng009\mint
go build ./ui/components/optiongroup/...
```

**预期结果**：
- 无编译错误
- 无未使用的变量/导入警告

#### 测试 2: 单元测试

```bash
go test ./ui/components/optiongroup/... -v
```

**预期结果**与之前相同（全部通过）

---

## ✅ Phase 3: 功能验证

### 目标
验证修复后的 OptionGroup 能够正常工作。

### 手动测试步骤

#### 测试 1: 运行 multiselect_demo

```bash
cd E:\projects\yao\wwsheng009\mint\examples\multiselect_demo
go run main.go
```

**测试步骤**：
1. 启动应用
2. 尝试点击 "Method 1: OptionGroup" 下的选项
   - [ ] 点击 "Fire" → 应该选中该选项
   - [ ] 点击 "Ice" → 应该选中，同时 Fire 仍然选中（多选）
   - [ ] 再次点击 "Fire" → 应该取消选中
3. 按 Tab 键
   - [ ] 应该能够在选项间按顺序移动焦点
   - [ ] 焦点应该显示在选中项或下一个选项上
4. 按 Enter/Space 键
   - [ ] 应该能够选中/取消选中当前选项
5. 点击 "Submit" 按钮
   - [ ] 如果有选中项，应该显示结果
   - [ ] 如果没有选中项，应该显示错误消息

**预期结果**：
- ✅ 所有交互正常
- ✅ 状态正确更新

#### 测试 2: 运行 typed_intent_demo

```bash
cd E:\projects\yao\wwsheng009\mint\examples\typed_intent_demo
go run main.go
```

**测试步骤**：
1. 启动应用
2. 在 "City Selection" 部分尝试单选
   - [ ] 点击 "Beijing" → 应该选中
   - [ ] 点击 "Shanghai" → 应该切换到 Shanghai
3. 在 "Interests" 部分尝试多选
   - [ ] 点击多个选项 → 应该同时选中
4. 按 Tab 键测试导航
5. 按 Enter/Space 键测试键盘操作

**预期结果**：
- ✅ 单选切换正常
- ✅ 多选累加/移除正常

---

### 自动化测试（可选）

如果需要添加集成测试，可以创建：

```go
// ui/components/optiongroup/integration_test.go
package optiongroup

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

func TestOptionGroup_SelectionIntegration(t *testing.T) {
	options := []Option{
		{Value: "opt1", Label: "Option 1"},
		{Value: "opt2", Label: "Option 2"},
		{Value: "opt3", Label: "Option 3"},
	}

	vnode := NewBuilder(options).
		Key("test-group").
		Mode(ModeMultiple).
		Build()

	// 创建 Instance
	inst := vnode.CreateInstance()
	inst.Init(vnode.Props())

	// 确保组件已挂载
	inst.OnMount()

	// 模拟点击第一个选项
	action := action.NewAction(action.ActionClick)
	childInst := findChildInstance(inst, 0)
	if childInst == nil {
		t.Fatal("无法找到子实例")
	}

	childInst.HandleAction(action)

	// 验证选中状态
	if !inst.isOptionSelected("opt1") {
		t.Errorf("期望 opt1 被选中，但实际未被选中")
	}
}

// findChildInstance 查找指定索引的子实例
// 注意：这需要 Fiber 树支持，可能需要在集成环境中实现
func findChildInstance(parent *Instance, idx int) rtui.ComponentInstance {
	// 实现略：需要遍历 Fiber 树获取子实例
	// 这里仅作为示例
	return nil
}
```

---

## 🐛 Phase 4: 问题排查

### 常见问题与解决方案

#### 问题 1: 编译错误 - "undefined: registerParent"

**症状**：
```
option.go:47: undefined: registerParent
```

**原因**：
清理代码时遗漏了某处调用。

**解决方案**：
```bash
# 搜索使用 registerParent 的地方
grep -rn "registerParent" ./ui/components/optiongroup/
```

检查是否还有调用 `registerParent`, `unregisterParent`, `lookupParent` 的地方，将其删除。

---

#### 问题 2: 点击选项无响应

**症状**：
闭包方案实施后，点击选项仍无响应。

**排查步骤**：

1. **检查 Props 传递**：
   在 `Children()` 中添加日志：
   ```go
   log.Printf("[DEBUG] Setting selectFunc for option %s, selectFunc = %v", opt.Value, o.optionSelectFunc)
   ```

2. **检查子实例获取**：
   在 `NewOptionInstance()` 中添加日志：
   ```go
   log.Printf("[DEBUG] NewOptionInstance selectFunc from props: %v", getSelectFuncProp(props))
   ```

3. **检查 HandleAction**：
   在 `OptionInstance.HandleAction()` 中添加日志：
   ```go
   log.Printf("[DEBUG] HandleAction: Type=%s, selectFunc=%v", act.Type, inst.selectFunc)
   ```

4. **检查闭包执行**：
   在 `SelectOption` 中添加日志：
   ```go
   func (inst *Instance) SelectOption(value string) {
       log.Printf("[DEBUG] SelectOption called: value=%s", value)
       // ... 原有逻辑
   }
   ```

**可能原因与解决方案**：

| 原因 | 解决方案 |
|------|---------|
| Props 未正确设置 | 检查 `Children()` 中的 Props 传递 |
| 闭包捕获了 nil 值 | 确保 `CreateInstance()` 在 `Children()` 之前执行 |
| 子实例未使用 Props | 检查 `NewOptionInstance()` 是否从 Props 读取 |
| 事件未正确分发 | 检查 `IntentEmitter` 是否正确设置 |

---

#### 问题 3: Tab 键卡住或焦点不移动

**症状**：
按 Tab 键时焦点不移动或陷入循环。

**可能原因**：
1. 焦点管理器未正确注册子实例
2. 子实例未实现 `FocusableInstance` 接口

**解决方案**：
```go
// 检查子实例是否实现了 FocusableInstance
var _ rtui.FocusableInstance = (*OptionInstance)(nil)

// 检查 OnMount 中是否注册了焦点
func (inst *OptionInstance) OnMount() {
    // 确保调用了 behaviors.OnMount
    inst.behaviors.OnMount(inst)
}
```

---

#### 问题 4: 内存泄漏

**症状**：
长时间运行后内存占用持续增长。

**可能原因**：
1. 闭包捕获了不需要的大型对象
2. 回调函数持有对父实例的强引用

**解决方案**：
```go
// 使用内存分析工具
go test -memprofile=mem.prof ./ui/components/optiongroup/...
go tool pprof mem.prof

// 检查闭包是否持有不必要的引用
// 闭包应该只捕获必要的值
func (o *VNode) Children() []rtui.VNode {
    for i, opt := range o.options {
        // ✅ 好：只捕获必要的引用
        closure := func(value string) {
            if o.optionSelectFunc != nil {
                o.optionSelectFunc(value)
            }
        }
        
        // ❌ 不好：捕获了整个 VNode
        badClosure := func(value string) {
            o.doSomethingComplex()  // 可能持有其他引用
        }
    }
}
```

---

## 📊 Phase 5: 性能分析

### 基准测试

创建性能基准测试：

```go
// ui/components/optiongroup/benchmark_test.go
package optiongroup

import (
	"testing"
)

func BenchmarkOptionGroup_Creation(b *testing.B) {
	options := make([]Option, 100)
	for i := 0; i < 100; i++ {
		options[i] = Option{
			Value: fmt.Sprintf("opt%d", i),
			Label: fmt.Sprintf("Option %d", i),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vnode := NewBuilder(options).Build()
		_ = vnode.CreateInstance()
	}
}

func BenchmarkOptionGroup_Paint(b *testing.B) {
	options := make([]Option, 10)
	for i := 0; i < 10; i++ {
		options[i] = Option{
			Value: fmt.Sprintf("opt%d", i),
			Label: fmt.Sprintf("Option %d", i),
		}
	}

	vnode := NewBuilder(options).Build()
	inst := vnode.CreateInstance()
	inst.Init(vnode.Props())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = inst.Paint(0, 0)
	}
}
```

**运行基准测试**：
```bash
go test -bench=. -benchmem ./ui/components/optiongroup/
```

**预期结果**：
- `BenchmarkOptionGroup_Creation`: ~100-500 ns/op
- `BenchmarkOptionGroup_Paint`: ~1-10 μs/op
- 无内存分配（Paint 应该复用）

---

## 📝 Phase 6: 文档更新

### 更新组件文档

更新组件头部的注释：

```go
// Package optiongroup provides a Store + Reducer compatible option group component.
//
// Architecture (Fiber-first):
//
//   - OptionGroup is a composite component that manages multiple options
//   - Each option is an independent Fiber node (OptionVNode + OptionInstance)
//   - Parent-Child communication uses closure wrapping to defer callback binding
//
//   VNode Tree:
//       OptionGroup.VNode
//           ├── OptionVNode (opt1)  → OptionInstance
//           ├── OptionVNode (opt2)  → OptionInstance
//           └── OptionVNode (opt3)  → OptionInstance
//
//   Callback Flow:
//       1. OptionGroup.VNode.CreateInstance() sets o.optionSelectFunc
//       2. OptionGroup.VNode.Children() creates child VNodes with closures
//       3. OptionInstance.HandleAction() executes the closure
//       4. Closure calls o.optionSelectFunc(value)
//       5. o.optionSelectFunc wraps inst.SelectOption(value)
//
// Usage:
//
//	options := []optiongroup.Option{
//	    {Value: "fire", Label: "Fire 🔥"},
//	    {Value: "ice", Label: "Ice ❄️"},
//	}
//
//	group := optiongroup.NewBuilder(options).
//	    Key("weapon-selector").
//	    Label("Weapons:").
//	    Mode(optiongroup.ModeMultiple).
//	    ForField(intent.BindField("weapons")).
//	    Vertical().
//	    Build()
//
// Reference:
//   - ARCHITECTURE_ANALYSIS_REPORT.md - Deep analysis and solution comparison
//   - CURRENT_STATUS.md - Current implementation status
//   - IMPLEMENTATION_GUIDE.md - This file, detailed implementation guide
package optiongroup
```

---

## 🎉 Phase 7: 提交与发布

### 提交清单

在提交代码前，确保：

- [ ] 所有测试通过（`go test ./ui/components/optiongroup/...`）
- [ ] 示例程序正常运行并功能正确
- [ ] 代码已格式化（`go fmt ./ui/components/optiongroup/...`）
- [ ] 无编译错误或警告
- [ ] 文档已更新

### Git 提交

```bash
# 添加修改的文件
git add ui/components/optiongroup/vnode.go
git add ui/components/optiongroup/option.go
git add ui/components/optiongroup/instance.go

# 创建提交
git commit -m "fix(optiongroup): 修复闭包包装，解决子选项无法选中问题

- 使用闭包包装延迟查找父回调
- 清理全局注册表等冗余代码
- 更新文档说明新架构

问题：
- 子选项的 selectFunc 在创建时为 nil
- 因为 Fiber 创建时序导致回调无法传递

解决方案：
- 在 Children() 中使用闭包捕获 VNode 引用
- 闭包延迟查找最新的 optionSelectFunc
- 确保 CreateInstance() 在 Children() 之前设置回调

测试：
- 所有单元测试通过
- multiselect_demo 和 typed_intent_demo 正常运行

参考：
- docs/ui/optiongroup/ARCHITECTURE_ANALYSIS_REPORT.md
- docs/ui/optiongroup/CURRENT_STATUS.md
- docs/ui/optiongroup/IMPLEMENTATION_GUIDE.md"

# 推送到远程（如果需要）
git push origin optiongroup-fix-closure
```

---

## 📞 获取帮助

如果遇到问题：

1. 查看相关文档：
   - [架构分析报告](./ARCHITECTURE_ANALYSIS_REPORT.md)
   - [当前状态](./CURRENT_STATUS.md)

2. 检查示例代码：
   - `examples/multiselect_demo/main.go`
   - `examples/typed_intent_demo/main.go`

3. 运行调试：
   ```bash
   # 启用详细日志
   LOG_LEVEL=debug go run ./examples/multiselect_demo/
   ```

4. 搜索相关 issue 或讨论

---

## 🔗 相关资源

### 设计文档
- [Store + Reducer 架构](../../examples/typesafe_form_demo_runapp/README.md)
- [Fiber-first 架构](../../runtime/ui/fiber.go)

### API 文档
- [OptionGroup Builder API](../builder.go)
- [VNode 接口](../../runtime/ui/vnode.go)

### 示例程序
- [Multi-Select Demo](../../examples/multiselect_demo/)
- [Type-Safe Form Demo](../../examples/typesafe_form_demo_runapp/)

---

## 📅 时间线

| 阶段 | 任务 | 预计时间 | 状态 |
|------|------|---------|------|
| ✅ Phase 0 | 环境准备 | 15分钟 | 已完成 |
| 🚧 Phase 1 | 实施闭包方案 | 2小时 | 待开始 |
| 🚧 Phase 2 | 清理冗余代码 | 1小时 | 待开始 |
| 🚧 Phase 3 | 功能验证 | 2小时 | 待开始 |
| ⏸️ Phase 4 | 问题排查 | 1-4小时 | 待定 |
| ⏸️ Phase 5 | 性能分析 | 1小时 | 待定 |
| ⏸️ Phase 6 | 文档更新 | 1小时 | 待定 |
| ⏸️ Phase 7 | 提交发布 | 30分钟 | 待定 |

**总预计时间**：约 8-13 小时

---

**祝实施顺利！** 🚀
