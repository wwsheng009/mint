package intent

import (
	"fmt"
)

// =============================================================================
// Built-in Handlers
// =============================================================================

// SetupBuiltinHandlers registers handlers for all built-in intent types.
// This should be called during application initialization.
//
// All builtin handlers are marked as OVERRIDABLE, allowing applications using
// Store + Reducer pattern to customize the behavior.
//
// Example:
//
//	rt := intent.NewRuntime()
//	intent.SetupBuiltinHandlers(rt)
func SetupBuiltinHandlers(rt *Runtime) {
	// State handlers (overridable for Store-based apps)
	RegisterTypedRuntimeWithOpts(rt, handleSetState, WithOverridable(true))
	RegisterTypedRuntimeWithOpts(rt, handleToggle, WithOverridable(true))
	RegisterTypedRuntimeWithOpts(rt, handleIncrement, WithOverridable(true))

	// Navigation handlers (overridable)
	RegisterTypedRuntimeWithOpts(rt, handleNavigate, WithOverridable(true))

	// Focus handlers (overridable)
	RegisterTypedRuntimeWithOpts(rt, handleFocus, WithOverridable(true))
	RegisterTypedRuntimeWithOpts(rt, handleBlur, WithOverridable(true))

	// Modal handlers (overridable - critical for Store-based modal state)
	RegisterTypedRuntimeWithOpts(rt, handleOpenModal, WithOverridable(true))
	RegisterTypedRuntimeWithOpts(rt, handleCloseModal, WithOverridable(true))

	// Tooltip handlers (overridable)
	RegisterTypedRuntimeWithOpts(rt, handleShowTooltip, WithOverridable(true))
	RegisterTypedRuntimeWithOpts(rt, handleHideTooltip, WithOverridable(true))

	// Form handlers (overridable)
	RegisterTypedRuntimeWithOpts(rt, handleSubmitForm, WithOverridable(true))
	RegisterTypedRuntimeWithOpts(rt, handleValidateForm, WithOverridable(true))

	// ✨ MVP: Field Change handler - marked as overridable
	// Users can override this with their own Store-based handler
	RegisterTypedWithOpts(rt.Registry, handleFieldChange, WithOverridable(true))

	// Data handlers (overridable)
	RegisterTypedRuntimeWithOpts(rt, handleLoadData, WithOverridable(true))
	RegisterTypedRuntimeWithOpts(rt, handleRefresh, WithOverridable(true))
}

// =============================================================================
// State Handlers
// =============================================================================

func handleSetState(ctx *ActionContext, i SetStateIntent) IntentResult {
	ctx.SetState(i.Key, i.Value)
	ctx.ScheduleUpdate()
	return HandledResult()
}

func handleToggle(ctx *ActionContext, i ToggleIntent) IntentResult {
	// Get current value
	current, ok := ctx.GetState(i.Key)
	if !ok {
		ctx.SetState(i.Key, true)
	} else {
		if b, ok := current.(bool); ok {
			ctx.SetState(i.Key, !b)
		} else {
			ctx.SetState(i.Key, true)
		}
	}
	ctx.ScheduleUpdate()
	return HandledResult()
}

func handleIncrement(ctx *ActionContext, i IncrementIntent) IntentResult {
	// Get current value
	current, ok := ctx.GetState(i.Key)
	if !ok {
		ctx.SetState(i.Key, i.Delta)
	} else {
		switch v := current.(type) {
		case int:
			ctx.SetState(i.Key, v+i.Delta)
		case int64:
			ctx.SetState(i.Key, v+int64(i.Delta))
		case float64:
			ctx.SetState(i.Key, v+float64(i.Delta))
		default:
			ctx.SetState(i.Key, i.Delta)
		}
	}
	ctx.ScheduleUpdate()
	return HandledResult()
}

// =============================================================================
// Navigation Handlers
// =============================================================================

func handleNavigate(ctx *ActionContext, i NavigateIntent) IntentResult {
	// Store navigation state
	ctx.SetState("__navigation_path", i.Path)
	if i.Params != nil {
		ctx.SetState("__navigation_params", i.Params)
	}
	ctx.ScheduleUpdate()
	return HandledResult()
}

// =============================================================================
// Focus Handlers
// =============================================================================

func handleFocus(ctx *ActionContext, i FocusIntent) IntentResult {
	ctx.SetState("__focus_target", i.TargetID)
	ctx.ScheduleUpdate()
	return HandledResult()
}

func handleBlur(ctx *ActionContext, i BlurIntent) IntentResult {
	ctx.SetState("__focus_target", "")
	ctx.ScheduleUpdate()
	return HandledResult()
}

// =============================================================================
// Modal Handlers
// =============================================================================

func handleOpenModal(ctx *ActionContext, i OpenModalIntent) IntentResult {
	ctx.SetState("__modal_id", i.ModalID)
	if i.Data != nil {
		ctx.SetState("__modal_data", i.Data)
	}
	ctx.SetState("__modal_visible", true)
	ctx.ScheduleUpdate()
	return HandledResult()
}

func handleCloseModal(ctx *ActionContext, i CloseModalIntent) IntentResult {
	ctx.SetState("__modal_visible", false)
	if i.Result != nil {
		ctx.SetState("__modal_result", i.Result)
	}
	ctx.ScheduleUpdate()
	return HandledResult()
}

// =============================================================================
// Tooltip Handlers
// =============================================================================

func handleShowTooltip(ctx *ActionContext, i ShowTooltipIntent) IntentResult {
	ctx.SetState("__tooltip_content", i.Content)
	ctx.SetState("__tooltip_target", i.TargetID)
	ctx.SetState("__tooltip_visible", true)
	ctx.ScheduleUpdate()
	return HandledResult()
}

func handleHideTooltip(ctx *ActionContext, i HideTooltipIntent) IntentResult {
	ctx.SetState("__tooltip_visible", false)
	ctx.ScheduleUpdate()
	return HandledResult()
}

// =============================================================================
// Form Handlers
// =============================================================================

func handleSubmitForm(ctx *ActionContext, i SubmitFormIntent) IntentResult {
	// Store form submission state
	ctx.SetState(fmt.Sprintf("__form_%s_submitting", i.FormID), true)
	ctx.SetState(fmt.Sprintf("__form_%s_data", i.FormID), i.Data)
	ctx.ScheduleUpdate()
	return HandledResult()
}

func handleValidateForm(ctx *ActionContext, i ValidateFormIntent) IntentResult {
	// Trigger validation state
	ctx.SetState(fmt.Sprintf("__form_%s_validating", i.FormID), true)
	ctx.ScheduleUpdate()
	return HandledResult()
}

// =============================================================================
// Field Change Handler (MVP)
// =============================================================================

// handleFieldChange handles FieldChangeIntent by updating the state with the runtime value.
// This is the MVP handler that makes State the single source of truth.
//
// Data Flow:
//   Instance (buffer) → FieldChangeIntent → State (authority) → VNode → Instance
func handleFieldChange(ctx *ActionContext, i FieldChangeIntent) IntentResult {
	// ✨ MVP: State is the single source of truth
	// The value from Instance (user input) becomes the new state value
	ctx.SetState(i.Field, i.Value)
	ctx.ScheduleUpdate()
	return HandledResult()
}

// =============================================================================
// Data Handlers (Transition)
// =============================================================================

func handleLoadData(ctx *ActionContext, i LoadDataIntent) IntentResult {
	// Mark as loading
	ctx.SetState(fmt.Sprintf("__data_%s_loading", i.Key), true)
	ctx.ScheduleUpdate()

	// Note: Actual data loading should be done by custom handlers
	// This is just a placeholder that sets the loading state
	return HandledResult()
}

func handleRefresh(ctx *ActionContext, i RefreshIntent) IntentResult {
	// Mark keys as refreshing
	for _, key := range i.Keys {
		ctx.SetState(fmt.Sprintf("__data_%s_refreshing", key), true)
	}
	ctx.ScheduleUpdate()
	return HandledResult()
}
