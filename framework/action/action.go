// Package action provides action dispatching and routing.
//
// Action System Architecture:
// - Framework-level (framework/action) ← Legacy framework usage
// - Runtime-level (runtime/action) ← New Fiber-first components
//
// ⚠️ DEPRECATED: Action types and Action struct should eventually be migrated to use runtime/action.
// This file maintains ActionType constants for backward compatibility.
// New components (ui/components/*) should use runtime/action.ActionType directly.
//
// For now, both coexist:
// - framework/action.ActionType = framework use (legacy components)
// - runtime/action.ActionType = framework use (new ui/components)
package action

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"
)

var actionIDCounter uint64

// Action 语义化操作
// Action 是比原始 Event 更高层次的抽象，表示用户意图而非低级输入
type Action struct {
	Type     ActionType
	Payload  interface{} // 操作附带的数据
	Source   string      // 触发源（如 "keyboard", "mouse", "system"）
	TargetID uint64      // 目标节点 ID（对于鼠标事件，现在使用 uint64）

	// ===== Phase 0 增强 =====
	ID        uint64                 // 唯一标识符（用于追踪和调试）
	Timestamp time.Time              // 创建时间
	stopped   bool                   // 内部：是否停止传播
	Meta      map[string]interface{} // 扩展元数据
}

// ActionType Action 类型（语义化操作）
type ActionType string

const (
	// ============================================================================
	// 导航类 Action - 用于在列表、树、表格等组件中移动焦点
	// ============================================================================
	ActionNavigateNext     ActionType = "navigate_next"     // 下一个项
	ActionNavigatePrev     ActionType = "navigate_prev"     // 上一个项
	ActionNavigateUp       ActionType = "navigate_up"       // 向上
	ActionNavigateDown     ActionType = "navigate_down"     // 向下
	ActionNavigateLeft     ActionType = "navigate_left"     // 向左
	ActionNavigateRight    ActionType = "navigate_right"    // 向右
	ActionNavigatePageUp   ActionType = "navigate_page_up"  // 向上翻页
	ActionNavigatePageDown ActionType = "navigate_page_down" // 向下翻页
	ActionNavigateHome     ActionType = "navigate_home"     // 跳到开头
	ActionNavigateEnd      ActionType = "navigate_end"      // 跳到结尾

	// ============================================================================
	// 选择类 Action - 用于选择、切换状态
	// ============================================================================
	ActionSelect   ActionType = "select"   // 选择当前项
	ActionToggle   ActionType = "toggle"   // 切换状态（展开/折叠、选中/未选中）
	ActionExpand   ActionType = "expand"   // 展开（树节点、下拉菜单等）
	ActionCollapse ActionType = "collapse" // 折叠

	// ============================================================================
	// 编辑类 Action - 用于文本输入和编辑
	// ============================================================================
	ActionInputText  ActionType = "input_text"  // 输入文本（Payload 为 string）
	ActionDeleteChar ActionType = "delete_char" // 删除字符
	ActionDeleteWord ActionType = "delete_word" // 删除单词
	ActionDeleteLine ActionType = "delete_line" // 删除行
	ActionBackspace  ActionType = "backspace"   // 退格
	ActionEnter      ActionType = "enter"       // 回车（确认输入）

	// ============================================================================
	// 表单类 Action - 用于表单提交和取消
	// ============================================================================
	ActionSubmit   ActionType = "submit"   // 提交表单
	ActionCancel   ActionType = "cancel"   // 取消操作
	ActionValidate ActionType = "validate" // 验证输入
	ActionReset    ActionType = "reset"    // 重置表单

	// ============================================================================
	// 系统类 Action - 系统级操作
	// ============================================================================
	ActionQuit    ActionType = "quit"    // 退出应用
	ActionFocus   ActionType = "focus"   // 获得焦点（Payload 为节点 ID）
	ActionBlur    ActionType = "blur"    // 失去焦点
	ActionInspect ActionType = "inspect" // 切换 Inspector（调试工具）
	ActionRefresh ActionType = "refresh" // 刷新内容

	// ============================================================================
	// 鼠标特定 Action - 鼠标交互
	// ============================================================================
	ActionClick       ActionType = "click"        // 左键点击（Payload 为 {X, Y int}）
	ActionDoubleClick ActionType = "double_click" // 双击
	ActionRightClick  ActionType = "right_click"  // 右键点击
	ActionMiddleClick ActionType = "middle_click" // 中键点击
	ActionScroll      ActionType = "scroll"       // 滚轮滚动（Payload 为 delta int）
	ActionHover       ActionType = "hover"        // 鼠标悬停（Payload 为 {X, Y int}）
	ActionDragStart   ActionType = "drag_start"   // 开始拖拽
	ActionDragMove    ActionType = "drag_move"    // 拖拽移动
	ActionDragEnd     ActionType = "drag_end"     // 结束拖拽

	// ============================================================================
	// 剪贴板 Action - 剪贴板操作
	// ============================================================================
	ActionCopy  ActionType = "copy"  // 复制
	ActionCut   ActionType = "cut"   // 剪切
	ActionPaste ActionType = "paste" // 粘贴

	// ============================================================================
	// 搜索类 Action - 搜索和查找
	// ============================================================================
	ActionSearch      ActionType = "search"       // 开始搜索
	ActionSearchNext  ActionType = "search_next"  // 下一个匹配项
	ActionSearchPrev  ActionType = "search_prev"  // 上一个匹配项
	ActionReplace     ActionType = "replace"      // 替换
	ActionReplaceAll  ActionType = "replace_all"  // 替换所有

	// ============================================================================
	// 视图类 Action - 视图操作
	// ============================================================================
	ActionZoomIn     ActionType = "zoom_in"     // 放大
	ActionZoomOut    ActionType = "zoom_out"    // 缩小
	ActionZoomReset  ActionType = "zoom_reset"  // 重置缩放
	ActionSplitView  ActionType = "split_view"  // 分割视图
	ActionCloseView  ActionType = "close_view"  // 关闭视图
	ActionMaximize   ActionType = "maximize"   // 最大化
	ActionMinimize   ActionType = "minimize"   // 最小化
	ActionFullscreen ActionType = "fullscreen" // 全屏

	// ============================================================================
	// Phase 0 新增：系统级 Action
	// ============================================================================
	ActionInit    ActionType = "init"    // 组件初始化
	ActionMount   ActionType = "mount"   // 组件挂载
	ActionUnmount ActionType = "unmount" // 组件卸载
	ActionResize  ActionType = "resize"  // 窗口调整（Payload 为 Size{W, H}）

	// ============================================================================
	// Phase 0 新增：焦点 Action
	// ============================================================================
	ActionFocusGained ActionType = "focus_gained" // 获得焦点
	ActionFocusLost   ActionType = "focus_lost"   // 失去焦点
	ActionFocusNext   ActionType = "focus_next"   // 焦点后移
	ActionFocusPrev   ActionType = "focus_prev"   // 焦点前移

	// ============================================================================
	// Phase 0 新增：数据 Action
	// ============================================================================
	ActionDataLoad   ActionType = "data_load"   // 数据加载
	ActionDataUpdate ActionType = "data_update" // 数据更新
	ActionDataError  ActionType = "data_error"  // 数据错误

	// ============================================================================
	// Phase 0 新增：撤销/重做
	// ============================================================================
	ActionUndo ActionType = "undo" // 撤销
	ActionRedo ActionType = "redo" // 重做

	// ============================================================================
	// 自定义 Action - 扩展点
	// ============================================================================
	ActionCustom ActionType = "custom" // 自定义操作（Type 为格式如 "custom:xxx"）
)

// ============================================================================
// Action 分类方法 - 用于快速判断 Action 类型
// ============================================================================

// IsNavigation 是否为导航 Action
func (a *Action) IsNavigation() bool {
	switch a.Type {
	case ActionNavigateNext, ActionNavigatePrev, ActionNavigateUp,
		ActionNavigateDown, ActionNavigateLeft, ActionNavigateRight,
		ActionNavigatePageUp, ActionNavigatePageDown,
		ActionNavigateHome, ActionNavigateEnd:
		return true
	}
	return false
}

// IsEditing 是否为编辑 Action
func (a *Action) IsEditing() bool {
	switch a.Type {
	case ActionInputText, ActionDeleteChar, ActionDeleteWord,
		ActionDeleteLine, ActionBackspace, ActionEnter:
		return true
	}
	return false
}

// IsSelection 是否为选择类 Action
func (a *Action) IsSelection() bool {
	switch a.Type {
	case ActionSelect, ActionToggle, ActionExpand, ActionCollapse:
		return true
	}
	return false
}

// IsForm 是否为表单类 Action
func (a *Action) IsForm() bool {
	switch a.Type {
	case ActionSubmit, ActionCancel, ActionValidate, ActionReset:
		return true
	}
	return false
}

// IsSystem 是否为系统类 Action
func (a *Action) IsSystem() bool {
	switch a.Type {
	case ActionQuit, ActionFocus, ActionBlur, ActionInspect, ActionRefresh:
		return true
	}
	return false
}

// IsMouse 是否为鼠标 Action
func (a *Action) IsMouse() bool {
	switch a.Type {
	case ActionClick, ActionDoubleClick, ActionRightClick, ActionMiddleClick,
		ActionScroll, ActionHover, ActionDragStart, ActionDragMove, ActionDragEnd:
		return true
	}
	return false
}

// IsClipboard 是否为剪贴板 Action
func (a *Action) IsClipboard() bool {
	switch a.Type {
	case ActionCopy, ActionCut, ActionPaste:
		return true
	}
	return false
}

// IsSearch 是否为搜索类 Action
func (a *Action) IsSearch() bool {
	switch a.Type {
	case ActionSearch, ActionSearchNext, ActionSearchPrev,
		ActionReplace, ActionReplaceAll:
		return true
	}
	return false
}

// IsView 是否为视图类 Action
func (a *Action) IsView() bool {
	switch a.Type {
	case ActionZoomIn, ActionZoomOut, ActionZoomReset,
		ActionSplitView, ActionCloseView, ActionMaximize,
		ActionMinimize, ActionFullscreen:
		return true
	}
	return false
}

// IsData 是否为数据类 Action (Phase 0 新增)
func (a *Action) IsData() bool {
	switch a.Type {
	case ActionDataLoad, ActionDataUpdate, ActionDataError:
		return true
	}
	return false
}

// IsUndoRedo 是否为撤销/重做 Action (Phase 0 新增)
func (a *Action) IsUndoRedo() bool {
	return a.Type == ActionUndo || a.Type == ActionRedo
}

// RequiresTarget 是否需要目标节点
func (a *Action) RequiresTarget() bool {
	// 鼠标 Action 通常需要目标
	return a.IsMouse() && a.TargetID != 0
}

// StopPropagation 停止 Action 传播 (Phase 0 新增)
func (a *Action) StopPropagation() {
	a.stopped = true
}

// IsStopped 检查是否停止传播 (Phase 0 新增)
func (a *Action) IsStopped() bool {
	return a.stopped
}

// IsNavigationAction 检查是否是导航 Action
// 导航 Action 由全局处理器处理，不需要 TargetID
func (a *Action) IsNavigationAction() bool {
	switch a.Type {
	case ActionNavigateNext, ActionNavigatePrev,
		ActionNavigateUp, ActionNavigateDown,
		ActionNavigateLeft, ActionNavigateRight,
		ActionNavigatePageUp, ActionNavigatePageDown,
		ActionNavigateHome, ActionNavigateEnd:
		return true
	}
	return false
}

// Reset 停止传播标志 (Phase 0 新增)
func (a *Action) resetStopped() {
	a.stopped = false
}

// GetPayloadString 获取字符串类型的 Payload
func (a *Action) GetPayloadString() (string, bool) {
	if s, ok := a.Payload.(string); ok {
		return s, true
	}
	return "", false
}

// GetPayloadInt 获取整数类型的 Payload
func (a *Action) GetPayloadInt() (int, bool) {
	if i, ok := a.Payload.(int); ok {
		return i, true
	}
	return 0, false
}

// GetPayloadPoint 获取点类型的 Payload（用于鼠标坐标）
func (a *Action) GetPayloadPoint() (x, y int, ok bool) {
	if p, ok := a.Payload.(struct{ X, Y int }); ok {
		return p.X, p.Y, true
	}
	if p, ok := a.Payload.(map[string]int); ok {
		// 检查是否包含 x 和 y
		x, hasX := p["x"]
		y, hasY := p["y"]
		if hasX && hasY {
			return x, y, true
		}
	}
	return 0, 0, false
}

// GetPayloadSize 获取 Size 类型的 Payload (Phase 0 新增)
func (a *Action) GetPayloadSize() (w, h int, ok bool) {
	if s, ok := a.Payload.(struct{ W, H int }); ok {
		return s.W, s.H, true
	}
	if s, ok := a.Payload.(map[string]int); ok {
		w, hasW := s["w"]
		h, hasH := s["h"]
		if hasW && hasH {
			return w, h, true
		}
	}
	return 0, 0, false
}

// GetMeta 获取元数据 (Phase 0 新增)
func (a *Action) GetMeta(key string) (interface{}, bool) {
	if a.Meta == nil {
		return nil, false
	}
	val, ok := a.Meta[key]
	return val, ok
}

// SetMeta 设置元数据 (Phase 0 新增)
func (a *Action) SetMeta(key string, value interface{}) {
	if a.Meta == nil {
		a.Meta = make(map[string]interface{})
	}
	a.Meta[key] = value
}

// String 返回 Action 的字符串表示
func (a *Action) String() string {
	var sb strings.Builder
	sb.WriteString(string(a.Type))

	if a.TargetID != 0 {
		fmt.Fprintf(&sb, "@%d", a.TargetID)
	}

	if a.Payload != nil {
		fmt.Fprintf(&sb, "(%v)", a.Payload)
	}

	if a.Source != "" {
		fmt.Fprintf(&sb, " [%s]", a.Source)
	}

	return sb.String()
}

// NewAction 创建一个新的 Action (Phase 0 增强: 自动分配 ID 和时间戳)
func NewAction(actionType ActionType) *Action {
	return &Action{
		Type:     actionType,
		Payload:  nil,
		Source:   "",
		TargetID: 0,
		ID:       atomic.AddUint64(&actionIDCounter, 1),
		Timestamp: time.Now(),
		Meta:     make(map[string]interface{}),
	}
}

// NewActionWithPayload 创建带 Payload 的 Action
func NewActionWithPayload(actionType ActionType, payload interface{}) *Action {
	return &Action{
		Type:      actionType,
		Payload:   payload,
		Source:    "",
		TargetID:  0,
		ID:        atomic.AddUint64(&actionIDCounter, 1),
		Timestamp: time.Now(),
		Meta:      make(map[string]interface{}),
	}
}

// NewActionFromMouse 创建鼠标 Action (Phase 0 增强: 自动分配 ID 和时间戳)
func NewActionFromMouse(actionType ActionType, targetID uint64, localX, localY int) *Action {
	return &Action{
		Type:      actionType,
		Payload:   struct{ X, Y int }{X: localX, Y: localY},
		Source:    "mouse",
		TargetID:  targetID,
		ID:        atomic.AddUint64(&actionIDCounter, 1),
		Timestamp: time.Now(),
		Meta:      make(map[string]interface{}),
	}
}

// NewActionFromKey 创建键盘 Action (Phase 0 增强: 自动分配 ID 和时间戳)
func NewActionFromKey(actionType ActionType, source string) *Action {
	return &Action{
		Type:      actionType,
		Payload:   nil,
		Source:    source,
		TargetID:  0,
		ID:        atomic.AddUint64(&actionIDCounter, 1),
		Timestamp: time.Now(),
		Meta:      make(map[string]interface{}),
	}
}

// Clone 克隆 Action
func (a *Action) Clone() *Action {
	return &Action{
		Type:     a.Type,
		Payload:  a.Payload,
		Source:   a.Source,
		TargetID: a.TargetID,
	}
}

// WithTarget 设置目标节点 ID，返回新的 Action
func (a *Action) WithTarget(targetID uint64) *Action {
	cloned := a.Clone()
	cloned.TargetID = targetID
	return cloned
}

// WithPayload 设置 Payload，返回新的 Action
func (a *Action) WithPayload(payload interface{}) *Action {
	cloned := a.Clone()
	cloned.Payload = payload
	return cloned
}

// WithSource 设置源，返回新的 Action
func (a *Action) WithSource(source string) *Action {
	cloned := a.Clone()
	cloned.Source = source
	return cloned
}
