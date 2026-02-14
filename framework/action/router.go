package action

import (
	"sort"

	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/event"
)

// ============================================================================
// Phase 0 新增：中间件和全局处理器
// ============================================================================

// ActionMiddleware 中间件接口
type ActionMiddleware interface {
	// Name 中间件名称
	Name() string

	// Before 在 Action 分发前调用
	// 返回 nil 表示拦截该 Action
	// 返回修改后的 Action 继续传播
	Before(action *Action) *Action

	// After 在 Action 分发后调用
	After(action *Action, result *RouterResult)
}

// GlobalActionHandler 全局处理器接口 (Phase 0 新增)
type GlobalActionHandler interface {
	// HandleGlobalAction 处理无目标的 Action
	HandleGlobalAction(action *Action) bool
	Priority() int
}

// MiddlewareChain 中间件链 (Phase 0 新增)
type MiddlewareChain struct {
	middlewares []ActionMiddleware
}

// NewMiddlewareChain 创建中间件链
func NewMiddlewareChain(middlewares ...ActionMiddleware) *MiddlewareChain {
	return &MiddlewareChain{middlewares: middlewares}
}

// Before 依次调用所有中间件的 Before 方法
func (c *MiddlewareChain) Before(action *Action) *Action {
	for _, mw := range c.middlewares {
		if action == nil {
			return nil
		}
		action = mw.Before(action)
	}
	return action
}

// After 逆序调用所有中间件的 After 方法
func (c *MiddlewareChain) After(action *Action, result *RouterResult) {
	// 逆序调用
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		c.middlewares[i].After(action, result)
	}
}

// Add 添加中间件
func (c *MiddlewareChain) Add(middleware ActionMiddleware) {
	c.middlewares = append(c.middlewares, middleware)
}

// ============================================================================

// Router 实现 Action 的三阶段分发系统
//
// Router 负责将语义化的 Action 分发给正确的组件：
//   1. Capture Phase: 从根向下到目标，按优先级调用捕获处理器
//   2. Target Phase: 在目标组件调用 HandleAction()
//   3. Bubble Phase: 从目标向上到根，调用冒泡处理器
//
// 使用示例：
//   router := NewRouter(rootNode)
//   router.AddCaptureHandler(inspector, PriorityHigh)
//   router.SetMiddleware(NewMiddlewareChain(loggingMW))
//   result := router.Dispatch(action)
type Router struct {
	// Root 是组件树的根节点
	Root *runtime.LayoutNode

	// CaptureHandlers 是捕获阶段处理器列表（按优先级排序）
	CaptureHandlers []*CaptureHandlerEntry

	// BubbleHandlers 是冒泡阶段处理器列表
	BubbleHandlers []*BubbleHandlerEntry

	// TargetHandlers 是特定目标的处理器映射（按 targetID 索引）
	TargetHandlers map[uint64]*TargetHandlerEntry

	// ===== Phase 0 新增 =====
	// GlobalHandlers 全局处理器（无目标 Action）
	GlobalHandlers []GlobalActionHandler

	// Middleware 中间件链
	Middleware *MiddlewareChain
}

// CaptureHandlerEntry 表示一个捕获阶段的处理器
type CaptureHandlerEntry struct {
	Handler  CaptureActionHandler // 处理器
	Priority int                  // 优先级（数值越大优先级越高）
	ID       string               // 处理器唯一标识
}

// BubbleHandlerEntry 表示一个冒泡阶段的处理器
type BubbleHandlerEntry struct {
	Handler  BubbleActionHandler // 处理器
	ID       string              // 处理器唯一标识
}

// TargetHandlerEntry 表示一个目标处理器
type TargetHandlerEntry struct {
	Handler ActionTarget // 目标组件
	TargetID uint64      // 目标 ID（现在使用 uint64）
}

// CaptureActionHandler 是捕获阶段的处理器接口
// 处理器可以在 Action 到达目标之前拦截它
type CaptureActionHandler interface {
	// HandleCapture 处理捕获阶段的 Action
	// 返回 true 表示停止传播，返回 false 表示继续传播
	HandleCapture(act *Action, target *runtime.LayoutNode) bool

	// Priority 返回处理器的优先级
	// 数值越大优先级越高，优先级高的处理器先执行
	Priority() int
}

// BubbleActionHandler 是冒泡阶段的处理器接口
// 处理器可以在 Action 从目标冒泡回来时处理它
type BubbleActionHandler interface {
	// HandleBubble 处理冒泡阶段的 Action
	// 返回 true 表示停止传播，返回 false 表示继续传播
	HandleBubble(act *Action, target *runtime.LayoutNode) bool
}

// RouterResult 表示 Action 分发的结果
type RouterResult struct {
	// Handled 表示是否有处理器处理了该 Action
	Handled bool

	// Stopped 表示 Action 是否被停止传播
	Stopped bool

	// Phase 表示 Action 在哪个阶段被处理
	Phase ActionPhase
}

// ActionPhase 表示 Action 分发的阶段
type ActionPhase int

const (
	// ActionPhaseNone 表示未开始
	ActionPhaseNone ActionPhase = iota

	// ActionPhaseCapture 捕获阶段（从根到目标）
	ActionPhaseCapture

	// ActionPhaseTarget 目标阶段（在目标组件）
	ActionPhaseTarget

	// ActionPhaseBubble 冒泡阶段（从目标到根）
	ActionPhaseBubble
)

// String 返回阶段的字符串表示
func (p ActionPhase) String() string {
	switch p {
	case ActionPhaseNone:
		return "None"
	case ActionPhaseCapture:
		return "Capture"
	case ActionPhaseTarget:
		return "Target"
	case ActionPhaseBubble:
		return "Bubble"
	default:
		return "Unknown"
	}
}

// NewRouter 创建新的 Router (Phase 0 增强: 初始化中间件链)
func NewRouter(root *runtime.LayoutNode) *Router {
	return &Router{
		Root:            root,
		CaptureHandlers: make([]*CaptureHandlerEntry, 0),
		BubbleHandlers:  make([]*BubbleHandlerEntry, 0),
		TargetHandlers:  make(map[uint64]*TargetHandlerEntry),
		GlobalHandlers:  make([]GlobalActionHandler, 0),
		Middleware:      NewMiddlewareChain(),
	}
}

// SetMiddleware 设置中间件链 (Phase 0 新增)
func (r *Router) SetMiddleware(chain *MiddlewareChain) {
	r.Middleware = chain
}

// AddMiddleware 添加中间件 (Phase 0 新增)
func (r *Router) AddMiddleware(middleware ActionMiddleware) {
	if r.Middleware == nil {
		r.Middleware = NewMiddlewareChain()
	}
	r.Middleware.Add(middleware)
}

// AddGlobalHandler 添加全局处理器 (Phase 0 新增)
func (r *Router) AddGlobalHandler(handler GlobalActionHandler) {
	r.GlobalHandlers = append(r.GlobalHandlers, handler)
	// 按优先级排序
	sort.Slice(r.GlobalHandlers, func(i, j int) bool {
		return r.GlobalHandlers[i].Priority() > r.GlobalHandlers[j].Priority()
	})
}

// AddCaptureHandler 添加捕获阶段处理器
func (r *Router) AddCaptureHandler(handler CaptureActionHandler, id string) {
	entry := &CaptureHandlerEntry{
		Handler:  handler,
		Priority: handler.Priority(),
		ID:       id,
	}

	r.CaptureHandlers = append(r.CaptureHandlers, entry)

	// 按优先级排序（从高到低）
	r.sortCaptureHandlers()
}

// AddBubbleHandler 添加冒泡阶段处理器
func (r *Router) AddBubbleHandler(handler BubbleActionHandler, id string) {
	entry := &BubbleHandlerEntry{
		Handler: handler,
		ID:      id,
	}

	r.BubbleHandlers = append(r.BubbleHandlers, entry)
}

// RegisterTarget 注册目标处理器
// 通常由组件树遍历时自动调用
func (r *Router) RegisterTarget(targetID uint64, handler ActionTarget) {
	r.TargetHandlers[targetID] = &TargetHandlerEntry{
		Handler:  handler,
		TargetID: targetID,
	}
}

// UnregisterTarget 注销目标处理器
func (r *Router) UnregisterTarget(targetID uint64) {
	delete(r.TargetHandlers, targetID)
}

// sortCaptureHandlers 按优先级排序捕获处理器（从高到低）
func (r *Router) sortCaptureHandlers() {
	sort.Slice(r.CaptureHandlers, func(i, j int) bool {
		return r.CaptureHandlers[i].Priority > r.CaptureHandlers[j].Priority
	})
}

// RemoveCaptureHandler 移除捕获阶段处理器
func (r *Router) RemoveCaptureHandler(id string) {
	for i, entry := range r.CaptureHandlers {
		if entry.ID == id {
			// 删除元素
			r.CaptureHandlers = append(r.CaptureHandlers[:i], r.CaptureHandlers[i+1:]...)
			return
		}
	}
}

// RemoveBubbleHandler 移除冒泡阶段处理器
func (r *Router) RemoveBubbleHandler(id string) {
	for i, entry := range r.BubbleHandlers {
		if entry.ID == id {
			r.BubbleHandlers = append(r.BubbleHandlers[:i], r.BubbleHandlers[i+1:]...)
			return
		}
	}
}

// GetRoot 获取根节点
func (r *Router) GetRoot() *runtime.LayoutNode {
	return r.Root
}

// SetRoot 设置根节点
func (r *Router) SetRoot(root *runtime.LayoutNode) {
	r.Root = root
}

// GetCaptureHandlers 获取所有捕获处理器
func (r *Router) GetCaptureHandlers() []*CaptureHandlerEntry {
	return r.CaptureHandlers
}

// GetBubbleHandlers 获取所有冒泡处理器
func (r *Router) GetBubbleHandlers() []*BubbleHandlerEntry {
	return r.BubbleHandlers
}

// GetTargetHandlers 获取所有目标处理器
func (r *Router) GetTargetHandlers() map[uint64]*TargetHandlerEntry {
	return r.TargetHandlers
}

// Dispatch 分发 Action 到正确的处理器
// 实现完整的三阶段传播：Middleware → Capture → Target → Bubble
// (Phase 0 增强: 支持中间件和全局处理器)
func (r *Router) Dispatch(act *Action) *RouterResult {
	result := &RouterResult{
		Handled: false,
		Stopped: false,
		Phase:   ActionPhaseNone,
	}

	// 如果没有 Action，直接返回
	if act == nil {
		return result
	}

	// Phase 0: 应用中间件（Before）
	if r.Middleware != nil {
		act = r.Middleware.Before(act)
		if act == nil {
			// 被中间件拦截
			result.Handled = true
			result.Stopped = true
			return result
		}
	}

	// 检查 Action 是否已停止传播
	if act.IsStopped() {
		result.Handled = true
		result.Stopped = true
		r.callMiddlewareAfter(act, result)
		return result
	}

	// Phase 0: 全局处理器（无目标 Action）
	if act.TargetID == 0 && len(r.GlobalHandlers) > 0 {
		for _, handler := range r.GlobalHandlers {
			if handler.HandleGlobalAction(act) {
				result.Handled = true
				r.callMiddlewareAfter(act, result)
				return result
			}
		}
	}

	// 如果有 TargetID，先找到目标节点
	var targetNode *runtime.LayoutNode
	if act.TargetID != 0 {
		targetNode = r.findNodeByID(act.TargetID)
		if targetNode == nil {
			// 目标不存在，仍然执行 Capture 和 Bubble
			targetNode = r.Root
		}
	} else {
		// 没有 TargetID，使用根节点作为目标
		targetNode = r.Root
	}

	// Phase 1: Capture（从根到目标）
	if r.capturePhase(act, targetNode, result) {
		r.callMiddlewareAfter(act, result)
		return result
	}

	// Phase 2: Target（在目标组件）
	if r.targetPhase(act, targetNode, result) {
		r.callMiddlewareAfter(act, result)
		return result
	}

	// Phase 3: Bubble（从目标到根）
	if r.bubblePhase(act, targetNode, result) {
		r.callMiddlewareAfter(act, result)
		return result
	}

	// 调用中间件 After
	r.callMiddlewareAfter(act, result)
	return result
}

// callMiddlewareAfter 调用中间件的 After 方法 (Phase 0 新增)
func (r *Router) callMiddlewareAfter(act *Action, result *RouterResult) {
	if r.Middleware != nil {
		r.Middleware.After(act, result)
	}
}

// capturePhase 执行捕获阶段
// 从根节点向下到目标节点，按优先级调用捕获处理器
func (r *Router) capturePhase(act *Action, target *runtime.LayoutNode, result *RouterResult) bool {
	result.Phase = ActionPhaseCapture

	// 按优先级调用所有捕获处理器
	for _, entry := range r.CaptureHandlers {
		if entry.Handler == nil {
			continue
		}

		// 调用捕获处理器
		stopped := entry.Handler.HandleCapture(act, target)
		if stopped {
			result.Handled = true
			result.Stopped = true
			return true // 停止传播
		}
	}

	return false // 继续传播
}

// targetPhase 执行目标阶段
// 在目标组件调用 HandleAction()
func (r *Router) targetPhase(act *Action, target *runtime.LayoutNode, result *RouterResult) bool {
	result.Phase = ActionPhaseTarget

	// 如果没有 TargetID，跳过目标阶段
	if act.TargetID == 0 {
		return false
	}

	// 查找目标处理器
	handlerEntry, exists := r.TargetHandlers[act.TargetID]
	if !exists || handlerEntry.Handler == nil {
		return false
	}

	// 检查目标是否能处理该 Action
	if !handlerEntry.Handler.CanHandleAction(act) {
		return false
	}

	// 调用目标的 HandleAction
	handled := handlerEntry.Handler.HandleAction(act)
	if handled {
		result.Handled = true
		return true // 处理完成，不再冒泡
	}

	return false // 继续冒泡
}

// bubblePhase 执行冒泡阶段
// 从目标节点向上到根节点，调用冒泡处理器
func (r *Router) bubblePhase(act *Action, target *runtime.LayoutNode, result *RouterResult) bool {
	result.Phase = ActionPhaseBubble

	// 先调用全局冒泡处理器
	for _, entry := range r.BubbleHandlers {
		if entry.Handler == nil {
			continue
		}

		stopped := entry.Handler.HandleBubble(act, target)
		if stopped {
			result.Handled = true
			result.Stopped = true
			return true // 停止传播
		}
	}

	// 沿着父链向上冒泡
	current := target
	for current != nil {
		// 检查当前节点是否是 ActionTarget
		if current.Component != nil && current.Component.Instance != nil {
			if targetHandler, ok := current.Component.Instance.(ActionTarget); ok {
				// 检查是否能处理该 Action
				if targetHandler.CanHandleAction(act) {
					handled := targetHandler.HandleAction(act)
					if handled {
						result.Handled = true
						return true
					}
				}
			}
		}

		// 移动到父节点
		current = current.Parent
	}

	return false
}

// findNodeByID 根据 ID 查找节点
func (r *Router) findNodeByID(id uint64) *runtime.LayoutNode {
	if r.Root == nil {
		return nil
	}

	return r.findNodeRecursive(r.Root, id)
}

// findNodeRecursive 递归查找节点
func (r *Router) findNodeRecursive(node *runtime.LayoutNode, id uint64) *runtime.LayoutNode {
	if node == nil {
		return nil
	}

	// 将节点 ID 字符串转换为 uint64 进行比较
	nodeID := r.stringToNodeID(node.ID)
	if nodeID == id {
		return node
	}

	// 递归检查子节点
	for _, child := range node.Children {
		if found := r.findNodeRecursive(child, id); found != nil {
			return found
		}
	}

	return nil
}

// BuildTargetRegistry 遍历组件树，注册所有 ActionTarget
// 这应该在每次渲染后调用，以保持注册表更新
func (r *Router) BuildTargetRegistry() {
	if r.Root == nil {
		return
	}

	// 清空现有注册表
	r.TargetHandlers = make(map[uint64]*TargetHandlerEntry)

	// 递归注册所有节点
	r.registerNodeRecursive(r.Root)
}

// registerNodeRecursive 递归注册节点
func (r *Router) registerNodeRecursive(node *runtime.LayoutNode) {
	if node == nil {
		return
	}

	// 如果节点实现了 ActionTarget，注册它
	if node.Component != nil && node.Component.Instance != nil {
		if target, ok := node.Component.Instance.(ActionTarget); ok {
			if node.ID != "" {
				nodeID := r.stringToNodeID(node.ID)
				r.RegisterTarget(nodeID, target)
			}
		}
	}

	// 递归注册子节点
	for _, child := range node.Children {
		r.registerNodeRecursive(child)
	}
}

// stringToNodeID 将字符串 ID 转换为 uint64 NodeID
// 使用与 runtime/event 包相同的 FNV-1a 哈希算法
func (r *Router) stringToNodeID(id string) uint64 {
	if id == "" {
		return 0
	}
	// 使用 runtime/event 包的转换函数以确保一致性
	return event.StringToNodeID(id)
}
