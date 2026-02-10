package action

import (
	"fmt"
)

// ActionTarget 是组件处理 Action 的接口
//
// 实现 ActionTarget 的组件可以接收和处理语义化的 Action，
// 而不需要直接处理原始的键盘/鼠标事件。
//
// 使用示例：
//   type Button struct {
//       id string
//       onClick func()
//   }
//
//   func (b *Button) HandleAction(action *Action) bool {
//       if action.Type == ActionClick && action.TargetID == b.id {
//           b.onClick()
//           return true
//       }
//       return false
//   }
//
//   func (b *Button) GetSupportedActions() []ActionType {
//       return []ActionType{ActionClick}
//   }
type ActionTarget interface {
	// HandleAction 处理一个 Action
	//
	// 返回值:
	//   true  - Action 已被处理，停止传播
	//   false - Action 未被处理，继续传播
	//
	// 实现注意事项:
	//   - 检查 Action.Type 是否支持
	//   - 检查 Action.TargetID 是否匹配（如果有）
	//   - 提取 Action.Payload 中的数据
	//   - 执行相应的操作
	//   - 返回 true 表示已处理
	HandleAction(action *Action) bool

	// GetSupportedActions 返回此组件支持的 Action 类型列表
	//
	// 返回的列表用于：
	//   - 文档生成
	//   - 调试信息
	//   - UI 提示（如显示可用的快捷键）
	//
	// 如果返回 nil 或空切片，表示组件不声明支持的 Action
	GetSupportedActions() []ActionType

	// CanHandleAction 检查组件是否能处理特定的 Action（可选）
	//
	// 这是 HandleAction 的"预检查"版本，不会修改组件状态
	// 用于 UI 高亮、鼠标光标样式等场景
	//
	// 如果组件不需要此功能，可以返回 false
	CanHandleAction(action *Action) bool
}

// ============================================================================
// ActionTarget 辅助接口
// ============================================================================

// FocusableActionTarget 是支持焦点的 ActionTarget
//
// 实现此接口的组件可以接收键盘焦点并处理键盘导航
type FocusableActionTarget interface {
	ActionTarget

	// Focus 设置焦点到该组件
	// 返回 true 表示焦点已设置，false 表示无法获得焦点
	Focus() bool

	// Blur 移除该组件的焦点
	Blur()

	// IsFocused 检查组件是否当前有焦点
	IsFocused() bool

	// IsFocusable 检查组件是否可以获得焦点
	IsFocusable() bool
}

// ScrollableActionTarget 是支持滚动的 ActionTarget
//
// 实现此接口的组件可以处理滚动 Action
type ScrollableActionTarget interface {
	ActionTarget

	// CanScroll 检查是否可以在指定方向滚动
	// delta > 0 表示向上滚动，delta < 0 表示向下滚动
	// 返回 true 表示可以滚动，false 表示已到边界
	CanScroll(delta int) bool

	// Scroll 执行滚动
	// delta > 0 表示向上滚动，delta < 0 表示向下滚动
	// 返回 true 表示滚动成功，false 表示无法滚动
	Scroll(delta int) bool

	// GetScrollPosition 获取当前滚动位置
	// 返回 (当前位置, 总范围, 可见范围)
	GetScrollPosition() (current, total, visible int)
}

// EditableActionTarget 是支持文本编辑的 ActionTarget
//
// 实现此接口的组件可以处理文本输入和编辑 Action
type EditableActionTarget interface {
	ActionTarget

	// InsertText 在光标位置插入文本
	InsertText(text string) bool

	// DeleteText 删除文本
	// direction: -1 表示向后删除（Backspace），1 表示向前删除（Delete）
	// 返回 true 表示删除成功，false 表示无法删除
	DeleteText(direction int) bool

	// ReplaceText 替换所有文本
	ReplaceText(text string) bool

	// GetText 获取当前文本内容
	GetText() string

	// GetCursorPosition 获取光标位置
	GetCursorPosition() int

	// SetCursorPosition 设置光标位置
	SetCursorPosition(pos int) bool
}

// SelectableActionTarget 是支持选择的 ActionTarget
//
// 实现此接口的组件可以处理选择相关 Action
type SelectableActionTarget interface {
	ActionTarget

	// Select 选择当前项
	Select() bool

	// IsSelected 检查当前项是否被选中
	IsSelected() bool

	// ToggleSelection 切换选择状态
	ToggleSelection() bool

	// GetSelectedCount 获取选中项的数量
	GetSelectedCount() int
}

// ExpandableActionTarget 是支持展开/折叠的 ActionTarget
//
// 实现此接口的组件可以处理展开和折叠 Action
type ExpandableActionTarget interface {
	ActionTarget

	// Expand 展开组件
	Expand() bool

	// Collapse 折叠组件
	Collapse() bool

	// IsExpanded 检查组件是否展开
	IsExpanded() bool

	// Toggle 切换展开/折叠状态
	Toggle() bool
}

// DraggableActionTarget 是支持拖拽的 ActionTarget
//
// 实现此接口的组件可以处理拖拽相关 Action
type DraggableActionTarget interface {
	ActionTarget

	// StartDrag 开始拖拽
	StartDrag(action *Action) bool

	// Drag 拖拽移动
	Drag(action *Action) bool

	// EndDrag 结束拖拽
	EndDrag(action *Action) bool

	// IsDragging 检查是否正在拖拽
	IsDragging() bool
}

// ============================================================================
// ActionTarget 辅助函数和工具
// ============================================================================

// HandleActionWithFallback 使用 ActionTarget 处理 Action，支持回退逻辑
//
// 如果 target 无法处理 Action，则调用 fallback 函数
//
// 返回值:
//   true  - Action 已被处理
//   false - Action 未被处理
func HandleActionWithFallback(target ActionTarget, action *Action, fallback func(*Action) bool) bool {
	// 先尝试使用 target 处理
	if target != nil && target.HandleAction(action) {
		return true
	}

	// 回退到 fallback 函数
	if fallback != nil {
		return fallback(action)
	}

	return false
}

// CanHandleActionOrFallback 检查是否能处理 Action，支持回退逻辑
func CanHandleActionOrFallback(target ActionTarget, action *Action, fallback func(*Action) bool) bool {
	if target != nil && target.CanHandleAction(action) {
		return true
	}

	if fallback != nil {
		return fallback(action)
	}

	return false
}

// GetActionTargets 从组件树中提取所有 ActionTarget
//
// nodes: 组件节点列表
// 返回: 所有实现 ActionTarget 的组件
func GetActionTargets(nodes []interface{}) []ActionTarget {
	targets := make([]ActionTarget, 0)

	for _, node := range nodes {
		if target, ok := node.(ActionTarget); ok {
			targets = append(targets, target)
		}
	}

	return targets
}

// FilterActionTargets 过滤出支持特定 Action 的目标
//
// targets: ActionTarget 列表
// actionType: 要过滤的 Action 类型
// 返回: 支持指定 Action 的目标列表
func FilterActionTargets(targets []ActionTarget, actionType ActionType) []ActionTarget {
	filtered := make([]ActionTarget, 0)

	for _, target := range targets {
		actions := target.GetSupportedActions()
		for _, supported := range actions {
			if supported == actionType {
				filtered = append(filtered, target)
				break
			}
		}
	}

	return filtered
}

// FindActionTarget 查找支持特定 Action 的第一个目标
//
// targets: ActionTarget 列表
// actionType: 要查找的 Action 类型
// 返回: 第一个支持指定 Action 的目标，如果没有则返回 nil
func FindActionTarget(targets []ActionTarget, actionType ActionType) ActionTarget {
	for _, target := range targets {
		if target.CanHandleAction(NewAction(actionType)) {
			return target
		}
	}
	return nil
}

// DispatchActionToTargets 将 Action 分发给目标列表
//
// 按照 ActionTarget 的顺序依次分发，直到某个目标处理了 Action
//
// 返回值:
//   true  - 至少有一个目标处理了 Action
//   false - 没有目标处理 Action
func DispatchActionToTargets(action *Action, targets ...ActionTarget) bool {
	for _, target := range targets {
		if target.HandleAction(action) {
			return true
		}
	}
	return false
}

// DispatchActionToTargetsWithFallback 将 Action 分发给目标列表，支持回退
//
// 依次尝试每个目标，如果所有目标都无法处理，则调用 fallback
//
// 返回值:
//   true  - Action 已被处理
//   false - Action 未被处理
func DispatchActionToTargetsWithFallback(action *Action, fallback func(*Action) bool, targets ...ActionTarget) bool {
	// 先尝试所有目标
	if DispatchActionToTargets(action, targets...) {
		return true
	}

	// 回退到 fallback 函数
	if fallback != nil {
		return fallback(action)
	}

	return false
}

// ============================================================================
// ActionTarget 基础实现
// ============================================================================

// BaseActionTarget 提供了 ActionTarget 的基础实现
//
// 嵌入此结构可以快速实现 ActionTarget 接口
type BaseActionTarget struct {
	// supportedActions 是此组件支持的 Action 类型列表
	supportedActions []ActionType
}

// NewBaseActionTarget 创建基础 ActionTarget
func NewBaseActionTarget(supportedActions ...ActionType) *BaseActionTarget {
	return &BaseActionTarget{
		supportedActions: supportedActions,
	}
}

// HandleAction 默认实现（总是返回 false）
func (b *BaseActionTarget) HandleAction(action *Action) bool {
	return false
}

// GetSupportedActions 返回支持的 Action 列表
func (b *BaseActionTarget) GetSupportedActions() []ActionType {
	return b.supportedActions
}

// CanHandleAction 检查是否支持指定的 Action
func (b *BaseActionTarget) CanHandleAction(action *Action) bool {
	for _, supported := range b.supportedActions {
		if supported == action.Type {
			return true
		}
	}
	return false
}

// AddSupportedActions 添加支持的 Action 类型
func (b *BaseActionTarget) AddSupportedActions(actions ...ActionType) {
	b.supportedActions = append(b.supportedActions, actions...)
}

// ============================================================================
// ActionTarget 组合器
// ============================================================================

// CompositeActionTarget 组合多个 ActionTarget 为一个
//
// 当 Action 到达时，依次尝试每个子目标，直到某个处理了它
type CompositeActionTarget struct {
	targets []ActionTarget
}

// NewCompositeActionTarget 创建组合 ActionTarget
func NewCompositeActionTarget(targets ...ActionTarget) *CompositeActionTarget {
	return &CompositeActionTarget{
		targets: targets,
	}
}

// AddTarget 添加子目标
func (c *CompositeActionTarget) AddTarget(target ActionTarget) {
	c.targets = append(c.targets, target)
}

// HandleAction 依次尝试每个子目标
func (c *CompositeActionTarget) HandleAction(action *Action) bool {
	for _, target := range c.targets {
		if target.HandleAction(action) {
			return true
		}
	}
	return false
}

// GetSupportedActions 返回所有子目标支持的 Action 并集
func (c *CompositeActionTarget) GetSupportedActions() []ActionType {
	actionsMap := make(map[ActionType]bool)

	for _, target := range c.targets {
		for _, action := range target.GetSupportedActions() {
			actionsMap[action] = true
		}
	}

	actions := make([]ActionType, 0, len(actionsMap))
	for action := range actionsMap {
		actions = append(actions, action)
	}

	return actions
}

// CanHandleAction 检查是否有子目标能处理 Action
func (c *CompositeActionTarget) CanHandleAction(action *Action) bool {
	for _, target := range c.targets {
		if target.CanHandleAction(action) {
			return true
		}
	}
	return false
}

// String 返回组合目标的字符串表示（用于调试）
func (c *CompositeActionTarget) String() string {
	return fmt.Sprintf("CompositeActionTarget{%d targets}", len(c.targets))
}

// ============================================================================
// ActionTargetAdapter 将函数适配为 ActionTarget
// ============================================================================

// ActionHandlerFunc 是处理 Action 的函数类型
type ActionHandlerFunc func(action *Action) bool

// ActionTargetAdapter 将函数适配为 ActionTarget 接口
//
// 使用示例：
//   handler := func(action *Action) bool {
//       if action.Type == ActionClick {
//           fmt.Println("Clicked!")
//           return true
//       }
//       return false
//   }
//   target := NewActionTargetAdapter(
//       []ActionType{ActionClick},
//       handler,
//   )
type ActionTargetAdapter struct {
	*BaseActionTarget
	handler ActionHandlerFunc
}

// NewActionTargetAdapter 创建 ActionTarget 适配器
func NewActionTargetAdapter(supportedActions []ActionType, handler ActionHandlerFunc) *ActionTargetAdapter {
	return &ActionTargetAdapter{
		BaseActionTarget: NewBaseActionTarget(supportedActions...),
		handler:          handler,
	}
}

// HandleAction 调用处理器函数
func (a *ActionTargetAdapter) HandleAction(action *Action) bool {
	if a.handler != nil {
		return a.handler(action)
	}
	return false
}

// CanHandleAction 检查处理器是否能处理 Action
func (a *ActionTargetAdapter) CanHandleAction(action *Action) bool {
	// 检查是否在支持的列表中
	if !a.BaseActionTarget.CanHandleAction(action) {
		return false
	}

	// 如果有处理器，尝试调用（调用 CanHandleAction 会检查实现）
	if a.handler != nil {
		// 创建一个测试 Action 来检查处理器是否想处理它
		// 这不是完美的方案，但在大多数情况下有效
		return true
	}

	return false
}

// ============================================================================
// 调试和诊断
// ============================================================================

// ActionTargetInfo 提供关于 ActionTarget 的调试信息
type ActionTargetInfo struct {
	Target           ActionTarget
	SupportedActions []ActionType
	CanHandleClick   bool
	CanHandleNavigate bool
	CanHandleEdit    bool
	CanHandleScroll  bool
}

// GetActionTargetInfo 获取 ActionTarget 的调试信息
func GetActionTargetInfo(target ActionTarget) *ActionTargetInfo {
	info := &ActionTargetInfo{
		Target:           target,
		SupportedActions: target.GetSupportedActions(),
	}

	// 测试各种 Action 类型
	info.CanHandleClick = target.CanHandleAction(NewAction(ActionClick))
	info.CanHandleNavigate = target.CanHandleAction(NewAction(ActionNavigateDown))
	info.CanHandleEdit = target.CanHandleAction(NewAction(ActionInputText))
	info.CanHandleScroll = target.CanHandleAction(NewAction(ActionScroll))

	return info
}

// String 返回 ActionTargetInfo 的字符串表示
func (i *ActionTargetInfo) String() string {
	return fmt.Sprintf(
		"ActionTargetInfo{Supported=%d, Click=%v, Navigate=%v, Edit=%v, Scroll=%v}",
		len(i.SupportedActions),
		i.CanHandleClick,
		i.CanHandleNavigate,
		i.CanHandleEdit,
		i.CanHandleScroll,
	)
}

// DumpActionTargets 导出 ActionTarget 列表的调试信息
func DumpActionTargets(targets []ActionTarget) string {
	var result string

	for i, target := range targets {
		info := GetActionTargetInfo(target)
		result += fmt.Sprintf("[%d] %s\n", i, info.String())
	}

	return result
}
