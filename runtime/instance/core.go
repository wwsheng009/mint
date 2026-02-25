package instance

import (
	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/runtime/cmd"
	"github.com/wwsheng009/mint/runtime/layout"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
)

// =============================================================================
// Instance Layer - 持久化组件实例层
// =============================================================================
// 根据 fix1.md 设计文档：
// - VNode 是"设计图"（每帧重建）
// - Instance 是"活的组件"（持久存在）
// - Layout 是"位置信息"
//
// 核心职责分离：
// | 层              | 应该干什么         | 现在谁在干         |
// | -------------- | ------------- | ------------- |
// | VNode          | 描述 UI 结构（纯数据） | ✔ 但现在还背了事件和状态 |
// | Instance       | 组件"活体"，持久存在   | ❌ 没有（现在要加）    |
// | LayoutNode     | 几何布局          | ✔             |
// | HitMap         | 几何命中测试        | ✔             |
// | Updater        | 组件行为处理        | ❌ 被塞在 VNode   |

// InstanceState 实例状态机
type InstanceState int

const (
	StateCreated InstanceState = iota
	StateMounted
	StateUpdated
	StateUnmounted
)

// Instance 组件实例 - 持久存在的"活体"
//
// 根据 fix1.md：
// > VNode 会死很多次，Instance 只在"真的被移除"时才死。
type Instance struct {
	// === Identity ===
	ID   string   // VNode key 标识（向后兼容）
	NodeID uint64 // 唯一运行时标识 ⭐⭐⭐
	Type string // 组件类型

	// === Props & State ===
	Props map[string]interface{} // 当前帧描述
	State interface{}            // 持久状态（不再在 VNode 上）

	// === Event Handlers ===
	// 事件处理器不再在 VNode 上，而是在 Instance 上
	Handlers Handlers

	// === Tree Structure ===
	Parent   *Instance
	Children []*Instance

	// === Layout Reference ===
	Layout *layout.Node // 布局实体引用

	// === Lifecycle ===
	State2 InstanceState // 状态机

	// === Dirty Flags ===
	DirtySelf   bool // 自身内容变了（文字/样式）
	DirtyLayout bool // 尺寸可能变化

	// === Internal ===
	_used bool // Reconcile 时的临时标记
}

// Handlers 事件处理器集合
type Handlers struct {
	OnClick      func()
	OnMouseEnter func()
	OnMouseLeave func()
	OnKeyPress   func(key string)
	OnUpdate     func(msg runtimemsg.Msg) cmd.Cmd // 修复：使用 cmd.Cmd 而不是 interface{}
}

// NewInstance 创建新实例
func NewInstance(id, typ string, props map[string]interface{}) *Instance {
	return &Instance{
		ID:         id,
		NodeID:     0, // Will be set when registered via InstanceRegistry
		Type:       typ,
		Props:      props,
		State2:     StateCreated,
		DirtySelf:  true,
		DirtyLayout: true,
	}
}

// Mount 挂载实例（生命周期）
func (inst *Instance) Mount() {
	inst.State2 = StateMounted
	inst.DirtySelf = true
	inst.DirtyLayout = true
}

// Update 更新实例（生命周期）
func (inst *Instance) Update(props map[string]interface{}) {
	inst.Props = props
	inst.State2 = StateUpdated
	inst.DirtySelf = true
}

// Unmount 卸载实例（生命周期）
func (inst *Instance) Unmount() {
	inst.State2 = StateUnmounted
	// 清理资源
	inst.Handlers = Handlers{}
	inst.Children = nil
}

// Handle 处理消息（事件入口）
func (inst *Instance) Handle(msg runtimemsg.Msg) cmd.Cmd {
	log.UILogger.Debug("[Instance.Handle] Called for Instance ID=%s Type=%s, OnUpdate=%v", inst.ID, inst.Type, inst.Handlers.OnUpdate != nil)
	// 根据消息类型调用对应的处理器
	if inst.Handlers.OnUpdate != nil {
		log.UILogger.Debug("[Instance.Handle] ✅ Calling OnUpdate for Instance ID=%s", inst.ID)
		return inst.Handlers.OnUpdate(msg)
	}
	log.UILogger.Debug("[Instance.Handle] ❌ No OnUpdate handler for Instance ID=%s", inst.ID)
	return nil
}

// MarkLayoutDirty 标记布局脏（向上传播）
//
// 根据 fix1.md：
// > 布局是父决定子的位置，所以子变了，父必须知道。
func (inst *Instance) MarkLayoutDirty() {
	for p := inst; p != nil; p = p.Parent {
		if p.DirtyLayout {
			break // 已经脏过，停止传播
		}
		p.DirtyLayout = true
	}
}

// AddChild 添加子实例
func (inst *Instance) AddChild(child *Instance) {
	child.Parent = inst
	inst.Children = append(inst.Children, child)
}

// SetHandler 设置事件处理器
func (inst *Instance) SetHandler(name string, handler interface{}) {
	switch name {
	case "onClick":
		if h, ok := handler.(func()); ok {
			inst.Handlers.OnClick = h
		}
	case "onMouseEnter":
		if h, ok := handler.(func()); ok {
			inst.Handlers.OnMouseEnter = h
		}
	case "onMouseLeave":
		if h, ok := handler.(func()); ok {
			inst.Handlers.OnMouseLeave = h
		}
	case "onKeyPress":
		if h, ok := handler.(func(string)); ok {
			inst.Handlers.OnKeyPress = h
		}
	case "onUpdate":
		if h, ok := handler.(func(runtimemsg.Msg) cmd.Cmd); ok {
			inst.Handlers.OnUpdate = h
		}
	}
}

// Dirty 标记是否需要重新布局/绘制
func (inst *Instance) Dirty() bool {
	return inst.DirtySelf || inst.DirtyLayout
}

// ClearDirty 清除脏标记
func (inst *Instance) ClearDirty() {
	inst.DirtySelf = false
	inst.DirtyLayout = false
}
