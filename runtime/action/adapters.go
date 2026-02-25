// Package action provides backward compatibility adapters for transitioning
// from the old framework/action and runtime/action to the unified system.
//
// This package should be used during the migration transition period only.
// After migration is complete, imports from framework/action should be removed.
package action

import (
	"fmt"

	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	runtimeplatform "github.com/wwsheng009/mint/runtime/platform"
)

// ========================================================================
// NOTE: To avoid circular import, framework-specific adapters (UpdaterAdapter,
// EventHandlerAdapter, ActionHandlerInstanceAdapter) should be defined in
// framework package, not here. This file only contains runtime-level adapters.
// ========================================================================

// ============================================================================
// Functional Adapters
// ============================================================================

// FuncAdapter wraps a function to implement Target interface
type FuncAdapter struct {
	fn ActionHandlerFunc
	id string
}

// NewFuncAdapter creates a function adapter
func NewFuncAdapter(id string, fn ActionHandlerFunc) *FuncAdapter {
	return &FuncAdapter{
		fn: fn,
		id: id,
	}
}

// ID returns the adapter ID
func (f *FuncAdapter) ID() string {
	return f.id
}

// HandleAction calls the wrapped function
func (f *FuncAdapter) HandleAction(a *Action) bool {
	return f.fn(a)
}

// ============================================================================
// Type Conversion Utilities
// ============================================================================

// ActionToMsg converts Action to legacy Msg format
// This is used by UpdaterAdapter for backward compatibility
func ActionToMsg(act *Action) runtimemsg.Msg {
	emptyMod := runtimemsg.Modifiers{}

	switch act.Type {
	case ActionInputText:
		if s, ok := act.GetPayloadString(); ok && len(s) > 0 {
			return runtimemsg.NewKeyMsg([]rune(s)[0], runtimeplatform.KeyUnknown, emptyMod)
		}
	case ActionInputChar:
		if r, ok := act.GetPayloadRune(); ok {
			return runtimemsg.NewKeyMsg(r, runtimeplatform.KeyUnknown, emptyMod)
		}
	case ActionNavigateUp:
		return runtimemsg.NewKeyMsg(0, runtimeplatform.KeyUp, emptyMod)
	case ActionNavigateDown:
		return runtimemsg.NewKeyMsg(0, runtimeplatform.KeyDown, emptyMod)
	case ActionNavigateLeft:
		return runtimemsg.NewKeyMsg(0, runtimeplatform.KeyLeft, emptyMod)
	case ActionNavigateRight:
		return runtimemsg.NewKeyMsg(0, runtimeplatform.KeyRight, emptyMod)
	case ActionNavigatePageUp:
		return runtimemsg.NewKeyMsg(0, runtimeplatform.KeyPageUp, emptyMod)
	case ActionNavigatePageDown:
		return runtimemsg.NewKeyMsg(0, runtimeplatform.KeyPageDown, emptyMod)
	case ActionNavigateHome:
		return runtimemsg.NewKeyMsg(0, runtimeplatform.KeyHome, emptyMod)
	case ActionNavigateEnd:
		return runtimemsg.NewKeyMsg(0, runtimeplatform.KeyEnd, emptyMod)
	case ActionEnter:
		return runtimemsg.NewKeyMsg(0, runtimeplatform.KeyEnter, emptyMod)
	case ActionBackspace:
		return runtimemsg.NewKeyMsg(0, runtimeplatform.KeyBackspace, emptyMod)
	case ActionDeleteChar:
		return runtimemsg.NewKeyMsg(0, runtimeplatform.KeyDelete, emptyMod)
	case ActionClick:
		if x, y, ok := act.GetPayloadPoint(); ok {
			return runtimemsg.NewMouseMsg(x, y, runtimemsg.MouseLeft, runtimemsg.MouseActionPress)
		}
	case ActionRightClick:
		if x, y, ok := act.GetPayloadPoint(); ok {
			return runtimemsg.NewMouseMsg(x, y, runtimemsg.MouseRight, runtimemsg.MouseActionPress)
		}
	case ActionMiddleClick:
		if x, y, ok := act.GetPayloadPoint(); ok {
			return runtimemsg.NewMouseMsg(x, y, runtimemsg.MouseMiddle, runtimemsg.MouseActionPress)
		}
	case ActionScroll:
		if p, ok := act.Payload.(runtimemsg.MouseMsg); ok {
			return &p
		}
	}
	return nil
}

// ============================================================================
// Migration Helpers
// ============================================================================

// MigrateTarget converts a target from string to uint64 or vice versa
// Used during migration when changing target ID formats
func MigrateTarget(oldID interface{}) (string, uint64) {
	switch v := oldID.(type) {
	case string:
		return v, StringToTargetID(v)
	case uint64:
		return fmt.Sprintf("%d", v), v
	default:
		return "", 0
	}
}

// StringFromUint64ID converts uint64 to string with optional prefix
// Use "action_" prefix if the ID looks like an auto-generated one
func StringFromUint64ID(id uint64) string {
	return fmt.Sprintf("%d", id)
}

// BatchAction executes multiple actions as a batch (dispatcher-friendly)
// This is different from CompositeAction (which uses context)
type BatchAction struct {
	actions []*Action
}

// NewBatchAction creates a batch action
func NewBatchAction(actions ...*Action) *BatchAction {
	return &BatchAction{
		actions: actions,
	}
}

// AddAction adds an action to the batch
func (b *BatchAction) AddAction(action *Action) {
	b.actions = append(b.actions, action)
}

// Execute executes all actions in sequence through dispatcher
// Note: This uses the simple Action result, not CompositeAction's ActionResult
func (b *BatchAction) Execute(dispatcher *Dispatcher) []SimpleActionResult {
	results := make([]SimpleActionResult, 0, len(b.actions))
	for _, a := range b.actions {
		handled := dispatcher.Dispatch(a)
		results = append(results, SimpleActionResult{
			Action:  a,
			Handled: handled,
		})
	}
	return results
}

// SimpleActionResult represents simple execution result
type SimpleActionResult struct {
	Action  *Action
	Handled bool
}

// DelayedAction delays action execution
type DelayedAction struct {
	action   *Action
	delay    int64 // milliseconds
	cancel   chan struct{}
}

// NewDelayedAction creates a delayed action
func NewDelayedAction(action *Action, delayMs int64) *DelayedAction {
	return &DelayedAction{
		action: action,
		delay:  delayMs,
		cancel: make(chan struct{}),
	}
}

// Cancel cancels the delayed action
func (d *DelayedAction) Cancel() {
	close(d.cancel)
}

// ThrottledAction throttles action execution
type ThrottledAction struct {
	action       *Action
	minInterval  int64 // milliseconds
	lastExecuted int64
}

// NewThrottledAction creates a throttled action
func NewThrottledAction(action *Action, minIntervalMs int64) *ThrottledAction {
	return &ThrottledAction{
		action:      action,
		minInterval: minIntervalMs,
	}
}

// IsNowAllowed checks if action can be executed now
func (t *ThrottledAction) IsNowAllowed(now int64) bool {
	return now-t.lastExecuted >= t.minInterval
}

// UpdateLastExecution updates the last execution timestamp
func (t *ThrottledAction) UpdateLastExecution(now int64) {
	t.lastExecuted = now
}
