// Package action provides a compatibility layer for migrating to runtime/action.
// This package re-exports runtime/action types to help with gradual migration.
//
// DEPRECATED: New code should use "github.com/wwsheng009/mint/runtime/action" directly.
// This package will be removed after migration is complete.
package action

import (
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/action"
)

// ============================================================================
// Re-export core types from runtime/action
// ============================================================================

// Action re-exports runtime/action.Action
type Action = action.Action

// ActionType re-exports runtime/action.ActionType
type ActionType = action.ActionType

// Target re-exports runtime/action.Target
type Target = action.Target

// ActionPhase re-exports runtime/action.ActionPhase
type ActionPhase = action.ActionPhase

// Router re-exports runtime/action.Router
type Router = action.Router

// RouterResult re-exports runtime/action.RouterResult
type RouterResult = action.RouterResult

// Dispatcher re-exports runtime/action.Dispatcher
type Dispatcher = action.Dispatcher

// InputProcessor re-exports runtime/action.InputProcessor
type InputProcessor = action.InputProcessor

// KeyMap re-exports runtime/action.KeyMap
type KeyMap = action.KeyMap

// ScopeDispatcher re-exports runtime/action.ScopeDispatcher
type ScopeDispatcher = action.ScopeDispatcher

// Error re-exports runtime/action.Error
type Error = action.Error

// ErrorType re-exports runtime/action.ErrorType
type ErrorType = action.ErrorType

// ============================================================================
// Action constants re-exported from runtime/action
// ============================================================================

const (
	// Navigation
	ActionNavigateNext     = action.ActionNavigateNext
	ActionNavigatePrev     = action.ActionNavigatePrev
	ActionNavigateUp       = action.ActionNavigateUp
	ActionNavigateDown     = action.ActionNavigateDown
	ActionNavigateLeft     = action.ActionNavigateLeft
	ActionNavigateRight    = action.ActionNavigateRight
	ActionNavigatePageUp   = action.ActionNavigatePageUp
	ActionNavigatePageDown = action.ActionNavigatePageDown
	ActionNavigateHome     = action.ActionNavigateHome
	ActionNavigateEnd      = action.ActionNavigateEnd

	// Selection
	ActionSelect   = action.ActionSelect
	ActionToggle   = action.ActionToggle
	ActionExpand   = action.ActionExpand
	ActionCollapse = action.ActionCollapse

	// Editing
	ActionInputText  = action.ActionInputText
	ActionDeleteChar = action.ActionDeleteChar
	ActionDeleteWord = action.ActionDeleteWord
	ActionDeleteLine = action.ActionDeleteLine
	ActionBackspace  = action.ActionBackspace
	ActionEnter      = action.ActionEnter

	// Form
	ActionSubmit   = action.ActionSubmit
	ActionCancel   = action.ActionCancel
	ActionValidate = action.ActionValidate
	ActionReset    = action.ActionReset
	ActionClear     = action.ActionClear

	// System
	ActionQuit    = action.ActionQuit
	ActionFocus   = action.ActionFocus
	ActionBlur    = action.ActionBlur
	ActionInspect = action.ActionInspect
	ActionRefresh = action.ActionRefresh

	// Mouse
	ActionClick       = action.ActionClick
	ActionDoubleClick = action.ActionDoubleClick
	ActionRightClick  = action.ActionRightClick
	ActionMiddleClick = action.ActionMiddleClick
	ActionScroll      = action.ActionScroll
	ActionHover       = action.ActionHover
	ActionDragStart   = action.ActionDragStart
	ActionDragMove    = action.ActionDragMove
	ActionDragEnd     = action.ActionDragEnd

	// Clipboard
	ActionCopy  = action.ActionCopy
	ActionCut   = action.ActionCut
	ActionPaste = action.ActionPaste

	// Phase constants (from router)
	PhaseCapture = action.ActionPhaseCapture
	PhaseTarget  = action.ActionPhaseTarget
	PhaseBubble  = action.ActionPhaseBubble
	PhaseNone    = action.ActionPhaseNone
)

// ============================================================================
// Re-export constructor functions from runtime/action
// ============================================================================

// NewAction re-exports runtime/action.NewAction
func NewAction(actionType ActionType) *Action {
	return action.NewAction(actionType)
}

// NewActionWithPayload re-exports runtime/action.NewActionWithPayload
func NewActionWithPayload(actionType ActionType, payload interface{}) *Action {
	return action.NewActionWithPayload(actionType, payload)
}

// NewActionFromKey re-exports runtime/action.NewActionFromKey
func NewActionFromKey(actionType ActionType, source string) *Action {
	return action.NewActionFromKey(actionType, source)
}

// NewActionFromMouse re-exports runtime/action.NewActionFromMouse
func NewActionFromMouse(actionType ActionType, localX, localY int) *Action {
	return action.NewActionFromMouse(actionType, localX, localY)
}

// NewInputProcessor re-exports runtime/action.NewInputProcessor
func NewInputProcessor() *InputProcessor {
	return action.NewInputProcessor()
}

// NewKeyMap re-exports runtime/action.NewKeyMap
func NewKeyMap() *KeyMap {
	return action.NewKeyMap()
}

// NewScopeDispatcher re-exports runtime/action.NewScopeDispatcher
func NewScopeDispatcher(parent *action.ScopeDispatcher) *ScopeDispatcher {
	return action.NewScopeDispatcher(parent)
}

// NewScopeDispatcherWithName re-exports runtime/action.NewScopeDispatcherWithName
func NewScopeDispatcherWithName(parent *action.ScopeDispatcher, name string) *ScopeDispatcher {
	return action.NewScopeDispatcherWithName(parent, name)
}

// NewRouter re-exports runtime/action.NewRouter
func NewRouter(root *runtime.LayoutNode) *Router {
	return action.NewRouter(root)
}

// NewDispatcher re-exports runtime/action.NewDispatcher
func NewDispatcher() *Dispatcher {
	return action.NewDispatcher()
}

// SetCurrentScopeDispatcher re-exports runtime/action.SetCurrentScopeDispatcher
func SetCurrentScopeDispatcher(d *ScopeDispatcher) {
	action.SetCurrentScopeDispatcher(d)
}

// GetCurrentScopeDispatcher re-exports runtime/action.GetCurrentScopeDispatcher
func GetCurrentScopeDispatcher() *ScopeDispatcher {
	return action.GetCurrentScopeDispatcher()
}

// ============================================================================
// Legacy compatibility - ActionTarget for framework components
// ============================================================================

// ActionTarget is the legacy interface for framework components.
// This interface is retained for backward compatibility with existing framework components.
//
// DEPRECATED: New components should implement runtime/ui.ActionHandlerInstance instead
// or directly register with the ActionRouter using runtime/action.Target.
type ActionTarget interface {
	HandleAction(action *Action) bool
	GetSupportedActions() []ActionType
}

// BaseActionTarget provides a base implementation for ActionTarget
type BaseActionTarget struct {
	supportedActions []ActionType
}

// NewBaseActionTarget creates a base action target
func NewBaseActionTarget(supportedActions ...ActionType) *BaseActionTarget {
	return &BaseActionTarget{
		supportedActions: supportedActions,
	}
}

// HandleAction returns false (placeholder)
func (b *BaseActionTarget) HandleAction(action *Action) bool {
	return false
}

// GetSupportedActions returns supported actions
func (b *BaseActionTarget) GetSupportedActions() []ActionType {
	return b.supportedActions
}

// CanHandleAction checks if action is supported
func (b *BaseActionTarget) CanHandleAction(a *Action) bool {
	for _, supported := range b.supportedActions {
		if supported == a.Type {
			return true
		}
	}
	return false
}

// IsNavigationAction checks if the action is a navigation action (legacy helper)
func IsNavigationAction(a *Action) bool {
	return a.IsNavigation()
}
