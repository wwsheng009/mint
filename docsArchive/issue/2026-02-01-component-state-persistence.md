# 组件状态持久化架构改进

## 问题描述

### 当前架构问题

```
当前 mint 架构:
- 每次渲染创建新的 VNode
- 状态 (isHovered) 存储在 VNode 中
- resetInteractiveElements() 清空旧的 VNode 引用
- collectInteractiveElements() 收集新的 VNode
- 状态在重新渲染时丢失
```

### React 的解决方案

```jsx
// React 组件
function Button({ onClick }) {
  const [isHovered, setIsHovered] = useState(false);
  return <button onMouseEnter={() => setIsHovered(true)} />;
}

// React 内部模型
ComponentInstance {
  state: { isHovered: false },  // ← 状态持久化在实例
  props: { onClick },
  fiber: {...}  // 连接到虚拟 DOM
}

// 渲染流程:
// 1. 创建新的 VNode (虚拟 DOM)
// 2. 复用已有的 ComponentInstance
// 3. 实例的 state 保持不变
// 4. 新 VNode 与旧 VNode diff，更新到 DOM
```

## 目标架构

### 分离关注点

```
┌─────────────────────────────────────────────────────────────┐
│                    Declarative UI                          │
│                                                             │
│  ComponentFunc (纯函数)                                    │
│    ↓                                                       │
│  返回 VNode (只是 UI 描述，无状态)                           │
│                                                             │
├─────────────────────────────────────────────────────────────┤
│                    Runtime                                │
│                                                             │
│  ComponentInstance (持久化实例)                             │
│    ├─ state: { isHovered, value, ... }                     │
│    ├─ props: { label, onClick, ... }                        │
│    ├─ lifecycle: { onMount, onUpdate, onUnmount }          │
│    └─ key: string (用于识别和匹配)                          │
│                                                             │
│  实例池管理:                                                 │
│    ┌─────────┐    ┌─────────┐    ┌─────────┐             │
│    │Instance1│    │Instance2│    │Instance3│             │
│    │key="btn1"│    │key="btn2"│    │key="btn3"│             │
│    └─────────┘    └─────────┘    └─────────┘             │
│                                                             │
│  匹配算法:                                                   │
│    新 VNode.key -> 查找实例 -> 复用或创建                │
│    新 VNode.props != 旧 Instance.props -> 更新 props         │
│    实例.state -> 传递给 ComponentFunc 重新渲染               │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## 核心接口设计

### ComponentInstance 接口

```go
// ComponentInstance 组件实例接口
type ComponentInstance interface {
	// 获取实例的唯一标识
	Key() string

	// 设置属性
	SetProps(props map[string]interface{})
	GetProps() map[string]interface{}

	// 获取状态
	GetState() map[string]interface{}

	// 更新状态（触发重新渲染）
	SetState(state map[string]interface{})

	// 渲染组件，返回新的 VNode
	Render() VNode

	// 生命周期钩子
	OnMount()
	// OnUpdate(newProps, oldProps) bool // 返回 false 取消更新
	OnUnmount()
}

// InstanceManager 实例管理器
type InstanceManager struct {
	instances map[string]ComponentInstance
	keys      map[string]string // VNode key -> instance key
}

// 查找或创建实例
func (m *InstanceManager) GetOrCreate(key string, creator func() ComponentInstance) ComponentInstance
```

### VNode 扩展

```go
// VNode 接口扩展
type VNode interface {
	// 现有方法...

	// 新增：获取组件标识（用于实例匹配）
	GetKey() string
	GetComponentFunc() ComponentFunc
	GetProps() map[string]interface{}
}

// ComponentVNode 组件 VNode
type ComponentVNode struct {
	*ElementVNode
	key         string
	componentFunc ComponentFunc
	props       map[string]interface{}
}

func Component(key string, fn ComponentFunc, props map[string]interface{}) VNode {
	return &ComponentVNode{
		ElementVNode: NewElement("component"),
		key:         key,
		componentFunc: fn,
		props:       props,
	}
}
```

### Hooks 状态管理

```go
// useState 实现改进
func useState(initialValue interface{}) (getter, setter) {
	ctx := currentContext()

	stateKey := fmt.Sprintf("state_%d", ctx.nextStateIndex)
	ctx.nextStateIndex++

	return func() interface{} {
		if val, ok := ctx.state[stateKey]; ok {
			return val
		}
		return initialValue
	}, func(newValue interface{}) {
		oldValue := ctx.state[stateKey]
		ctx.state[stateKey] = newValue
		ctx.scheduleUpdate()

		// 触发 afterUpdate
		if ctx.afterUpdate != nil {
			ctx.afterUpdate(oldValue, newValue)
		}
	}
}

// useHoverState 专门处理 hover 状态
func useHoverState() (isHovered bool, setIsHovered func(bool)) {
	ctx := currentContext()

	hoverKey := fmt.Sprintf("hover_%d", ctx.nextStateIndex)
	ctx.nextStateIndex++

	// 使用 ref 来保持状态跨渲染
	state := ctx.getOrSetRef(hoverKey, false)

	return func() bool {
		return state.(bool)
	}, func(v bool) {
		ctx.setRef(hoverKey, v)
	}
}
```

## 实施步骤

### Phase 1: 核心接口定义（1-2天）

- [ ] 定义 `ComponentInstance` 接口
- [ ] 定义 `InstanceManager` 结构
- [ ] 扩展 `VNode` 接口添加 `GetKey()`
- [ ] 创建 `ComponentVNode` 类型

### Phase 2: 实例管理器（2-3天）

- [ ] 实现 `InstanceManager.GetOrCreate()`
- [ ] 实现 key 匹配算法
- [ ] 实例生命周期管理
- [ ] 单元测试

### Phase 3: Hooks 状态迁移（2-3天）

- [ ] 修改 `useState` 使用实例状态
- [ ] 添加 `useRef()` 用于跨渲染状态保持
- [ ] 添加 `useHoverState()` 专用 hook
- [ ] 迁移现有 hover 状态到 hooks

### Phase 4: 渲染流程集成（2-3天）

- [ ] 修改 `declarativeRoot.Paint()` 集成实例管理器
- [ ] 实现 diff 算法：VNode -> 实例匹配
- [ ] 实现 props 更新检测
- [ ] 集成测试

### Phase 5: 组件生命周期（2-3天）

- [ ] 实现 `OnMount` / `OnUnmount` 钩子
- [ ] 实现 `OnUpdate` 钩子
- [ ] 清理逻辑

### Phase 6: 文档和示例（1-2天）

- [ ] 编写组件开发指南
- - ] 更新示例代码
- [ ] 迁移指南

### Phase 7: 测试和验证（2-3天）

- [ ] 端到端测试
- [ ] 性能测试
- [ ] 回归测试
- [ ] Bug 修复

## 兼容性策略

### 渐进式迁移

```
阶段 1: 临时方案（当前完成）
  - 渲染后恢复 hover 状态
  - 使用 bounds 匹配
  - 问题：位置改变时仍会丢失

阶段 2: 新架构（本计划）
  - 引入 ComponentInstance
  - 状态持久化到实例
  - 保持向后兼容

阶段 3: 完全迁移（未来）
  - 所有状态使用 hooks
  - VNode 纯粹用作渲染描述
  - 移除 VNode 中的状态字段
```

### 向后兼容

```go
// 保留旧的直接在 VNode 中设置状态的方式
type ButtonVNode struct {
	*ElementVNode
	label       string
	onClick     func()
	// ...

	// Deprecated: 使用 useHoverState() 代替
	isHovered   bool // 保留以兼容，标记为 deprecated
}

// 新的推荐方式
func MyComponent() ui.VNode {
	isHovered, setIsHovered := ui.UseHoverState()

	return ui.Button("Click me").OnMouseEnter(func() {
		setIsHovered(true)
	})
}
```

## 性能考虑

### 实例池管理

```go
type InstanceManager struct {
	instances map[string]ComponentInstance
	maxInstances int              // 防止内存泄漏
}

// 限制活跃实例数量
func (m *InstanceManager) GetOrCreate(key string, creator func() ComponentInstance) ComponentInstance {
	// 如果超过限制，清理最久未使用的实例
	if len(m.instances) >= m.maxInstances {
		m.cleanupOldest()
	}
	// ...
}
```

### Diff 优化

```go
// Props 比较优化
func propsEqual(new, old map[string]interface{}) bool {
	if len(new) != len(old) {
		return false
	}
	for k, v := range new {
		if old[k] != v {
			return false
		}
	}
	return true
}

// 如果 props 没变，跳过组件重新渲染
func (m *InstanceManager) UpdateIfNeeded(instance ComponentInstance, newProps map[string]interface{}) bool {
	if propsEqual(newProps, instance.GetProps()) {
		return false // 无需更新
	}
	instance.SetProps(newProps)
	return true // 需要重新渲染
}
```

## 测试计划

### 单元测试

```go
func TestInstanceManager_GetOrCreate(t *testing.T) {
	m := NewInstanceManager()

	// 第一次创建
	instance1 := m.GetOrCreate("btn1", func() ComponentInstance {
		return &ButtonInstance{key: "btn1"}
	})

	// 第二次获取，应该返回同一实例
	instance2 := m.GetOrCreate("btn1", func() ComponentInstance {
		return &ButtonInstance{key: "btn1-unique"}
	})

	assert.Equal(t, instance1, instance2)
}

func TestState_Persistence(t *testing.T) {
	// 创建组件并设置状态
	instance := &ButtonInstance{}
	instance.SetState("isHovered", true)

	// 模拟重新渲染
	newProps := map[string]interface{}{}
	instance.SetProps(newProps)

	// 状态应该保持
	assert.True(t, instance.GetState()["isHovered"].(bool))
}
```

### 集成测试

```go
func TestHoverState_Persistence(t *testing.T) {
	// 模拟多次渲染
	root := newDeclarativeRoot(app, 50, 20)

	// 第一次渲染
	vnode1 := root.appFn()
	root.renderVNode(vnode1, 0, 0, buffer)

	// 模拟鼠标进入
	mouseEvent := &MouseEvent{Type: EventMouseEnter, X: 5, Y: 10}
	root.handleMouseEvent(mouseEvent)

	// 检查 hover 状态
	assert.True(t, root.buttons[0].IsHovered())

	// 第二次渲染（模拟状态保持）
	vnode2 := root.appFn()
	root.renderVNode(vnode2, 0, 0, buffer)

	// hover 状态应该保持
	assert.True(t, root.buttons[0].IsHovered())
}
```

## 风险评估

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| 架构变更大 | 现有代码可能需要适配 | 提供兼容层，分阶段迁移 |
| 性能开销 | 实例管理增加复杂度 | 性能基准测试，优化热点 |
| 内存占用 | 持久化实例增加内存 | 实例池限制，LRU 清理 |
| 调试难度 | 状态流转更复杂 | 完善日志，调试工具 |

## 时间估算

| 阶段 | 工作量 | 时间 |
|------|--------|------|
| Phase 1: 核心接口 | 中 | 1-2 天 |
| Phase 2: 实例管理器 | 中-高 | 2-3 天 |
| Phase 3: Hooks 迁移 | 高 | 2-3 天 |
| Phase 4: 渲染集成 | 高 | 2-3 天 |
| Phase 5: 生命周期 | 中 | 2-3 天 |
| Phase 6: 文档示例 | 低-中 | 1-2 天 |
| Phase 7: 测试验证 | 中 | 2-3 天 |
| **总计** | | **14-21 天** |

## 相关文件

### 需要修改的文件

| 文件 | 修改内容 |
|------|---------|
| `ui/hooks.go` | 添加 `useRef`, `useHoverState` |
| `ui/app.go` | 集成 `InstanceManager` |
| `ui/vnode.go` | 扩展 `VNode` 接口 |
| `ui/component.go` | 新增 `ComponentInstance` 接口 |
| `ui/instance.go` | 新增实例管理器 |
| `ui/button.go` | 修改为使用 hooks |

### 需要新增的文件

| 文件 | 说明 |
|------|------|
| `ui/instance.go` | 实例管理器 |
| `ui/instance_manager.go` | 实例管理器实现 |
| `ui/button_instance.go` | Button 组件实例 |
| `ui/input_instance.go` | Input 组件实例 |
| `ui/checkbox_instance.go` | Checkbox 组件实例 |
| 等等... | 其他组件实例 |

## 参考资料

- React Hooks: https://react.dev/reference/react
- React Reconciliation: https://react.dev/learn/understanding-react-ui-render-functions
- Fiber Architecture: https://github.com/acdlite/react/blob/main/README.md
