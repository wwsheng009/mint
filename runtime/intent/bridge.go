package intent

import (
	"github.com/wwsheng009/mint/runtime/action"
)

// =============================================================================
// Action Bridge - Integration with existing Action System
// =============================================================================

// ActionBridge converts between Intent and Action systems.
// This allows gradual migration from Action to Intent.
type ActionBridge struct {
	dispatcher *Dispatcher
	registry   *Registry
}

// NewActionBridge creates a new action bridge.
func NewActionBridge(dispatcher *Dispatcher, registry *Registry) *ActionBridge {
	return &ActionBridge{
		dispatcher: dispatcher,
		registry:   registry,
	}
}

// IntentFromAction converts an Action to an Intent.
// This allows existing Action handlers to process Intents.
func IntentFromAction(a *action.Action) Intent {
	if a == nil {
		return nil
	}

	// Check if payload contains an intent
	if intent, ok := a.Payload.(Intent); ok {
		return intent
	}

	// Convert action type to generic intent
	return &actionIntent{
		actionType: string(a.Type),
		payload:    a.Payload,
		target:     a.Target,
	}
}

// ActionFromIntent converts an Intent to an Action.
// This allows Intent emission through existing Action channels.
func ActionFromIntent(intent Intent) *action.Action {
	return &action.Action{
		Type:    action.ActionType(intent.IntentType()),
		Payload: intent,
	}
}

// actionIntent is an adapter that wraps an Action as an Intent.
type actionIntent struct {
	actionType string
	payload    interface{}
	target     string
}

func (a *actionIntent) IntentType() string {
	return a.actionType
}

// RegisterIntentHandler registers an Intent handler that can handle Actions.
// This provides backward compatibility with existing Action-based code.
func (b *ActionBridge) RegisterIntentHandler(intentType string, handler Handler) func() {
	return b.registry.Register(intentType, handler)
}

// DispatchFromAction dispatches an Intent derived from an Action.
func (b *ActionBridge) DispatchFromAction(a *action.Action) IntentResult {
	intent := IntentFromAction(a)
	if intent == nil {
		return ErrorResult(nil)
	}
	return b.dispatcher.Dispatch(intent)
}

// =============================================================================
// Typed Action-Intent Conversion
// =============================================================================

// TypedActionIntent is a typed wrapper for converting Actions to Intents.
type TypedActionIntent[T Intent] struct {
	intent T
}

// FromAction extracts a typed intent from an action payload.
func FromAction[T Intent](a *action.Action) (T, bool) {
	if a == nil {
		var zero T
		return zero, false
	}
	if intent, ok := a.Payload.(T); ok {
		return intent, true
	}
	var zero T
	return zero, false
}

// ToAction converts a typed intent to an action.
func ToAction[T Intent](intent T) *action.Action {
	return ActionFromIntent(intent)
}

// =============================================================================
// Common Action-Intent Mappings
// =============================================================================

// ClickActionToIntent converts ActionClick to ClickIntent.
func ClickActionToIntent(a *action.Action, targetID string) ClickIntent {
	return ClickIntent{TargetID: targetID}
}

// EnterActionToIntent converts ActionEnter to PressIntent.
func EnterActionToIntent(a *action.Action, targetID string) PressIntent {
	return PressIntent{TargetID: targetID}
}

// FocusActionToIntent converts ActionFocus to FocusIntent.
func FocusActionToIntent(a *action.Action, targetID string) FocusIntent {
	return FocusIntent{TargetID: targetID}
}
