package form

import (
	"sync"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/validation"
)

// =============================================================================
// Instance - Form Runtime Entity
// =============================================================================

// Instance is the runtime entity for Form components.
// It acts as a container for form fields and manages validation state.
//
// Architecture (Phase 6: Intent Bubble Migration):
// - Implements TreeContainer to manage child field instances
// - Implements IntentHandler to process form-related intents:
//   * FormFieldChangeIntent: Track field changes and dirty state
//   * FormFieldBlurIntent: Trigger field validation
//   * FormValidateIntent: Validate form/field
//   * FormSubmitIntent: Validate and submit form data
//   * FormResetIntent: Reset form to initial state
type Instance struct {
	// === Identification ===
	key string

	// === Props (from VNode, may change each render) ===
	label       string
	formStyle   style.Style
	onSubmit    intent.Intent // Intent to emit on successful submit
	onReset     intent.Intent // Intent to emit on reset
	validateAll bool          // Whether to validate all fields on submit (default: true)

	// === Runtime State (managed by instance) ===
	mu            sync.RWMutex
	values        map[string]interface{} // Current field values
	initialValues map[string]interface{}  // Initial field values (for reset)
	errors        map[string]string        // Field validation errors
	isValid       bool                     // Overall form validity
	isSubmitting  bool                     // Submission in progress
	dirty         bool                     // Form has unsubmitted changes

	// === Instance Tree (Phase 1) ===
	childInstances []rtui.ComponentInstance

	// === Layout ===
	bounds [4]int // x, y, w, h
}

var (
	_ rtui.ComponentInstance = (*Instance)(nil)
	_ rtui.PaintableInstance = (*Instance)(nil)
	_ intent.IntentHandler   = (*Instance)(nil)
	_ interface {
		Measure(layout.Constraints) layout.Size
	} = (*Instance)(nil)
)

// =============================================================================

// NewInstance creates a new Form Instance.
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		key:           getStringProp(props, "key", "form"),
		label:         getStringProp(props, "label", ""),
		formStyle:     getStyleProp(props),
		validateAll:   getBoolProp(props, "validateAll", true),
		values:        make(map[string]interface{}),
		initialValues: make(map[string]interface{}),
		errors:        make(map[string]string),
		isValid:       true,
	}

	if v, ok := props["onSubmit"].(intent.Intent); ok {
		inst.onSubmit = v
	}
	if v, ok := props["onReset"].(intent.Intent); ok {
		inst.onReset = v
	}

	// Initialize with field values from props
	if fieldValues, ok := props["values"].(map[string]interface{}); ok {
		for k, v := range fieldValues {
			inst.values[k] = v
			inst.initialValues[k] = v
		}
	}

	return inst
}

// =============================================================================
// ComponentInstance Interface
// =============================================================================

func (inst *Instance) Key() string { return inst.key }

func (inst *Instance) SetKey(key string) { inst.key = key }

func (inst *Instance) Init(props rtui.Props) {
	inst.SetProps(props)
}

func (inst *Instance) Destroy() {
	// Clean up child instances
	_ = inst.ClearChildren()
}

func (inst *Instance) OnMount() {
	// Phase 2: Context System
	// Register form instance so children can access it via FormContext
	formID := inst.Key()
	if formID != "" {
		RegisterForm(formID, inst)
	}
}

func (inst *Instance) OnUnmount() {
	// Unregister form instance to clean up
	formID := inst.Key()
	if formID != "" {
		UnregisterForm(formID)
	}
}

func (inst *Instance) SetProps(props rtui.Props) bool {
	inst.mu.Lock()
	defer inst.mu.Unlock()

	changed := false

	if v, ok := props["label"].(string); ok {
		if inst.label != v {
			inst.label = v
			changed = true
		}
	}
	// Check if style prop exists and changed
	if s, hasStyle := props["style"].(style.Style); hasStyle {
		if s != inst.formStyle {
			inst.formStyle = s
			changed = true
		}
	}
	if v, ok := props["onSubmit"].(intent.Intent); ok {
		inst.onSubmit = v
	}
	if v, ok := props["onReset"].(intent.Intent); ok {
		inst.onReset = v
	}
	if v, ok := props["validateAll"].(bool); ok {
		inst.validateAll = v
	}

	if changed {
		inst.dirty = true
	}

	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	inst.mu.RLock()
	defer inst.mu.RUnlock()

	return rtui.Props{
		"key":         inst.key,
		"label":       inst.label,
		"validateAll": inst.validateAll,
	}
}

func (inst *Instance) MarkDirty() { inst.dirty = true }
func (inst *Instance) IsDirty() bool  { return inst.dirty }
func (inst *Instance) GetContext() *rtui.ComponentContext { return nil }

// =============================================================================
// Instance Tree Methods (Phase 1)
// =============================================================================

// Parent implements TreeComponent interface (intent bubble).
// Returns nil as Form is typically a root or top-level container.
func (inst *Instance) Parent() interface{} {
	return nil
}

// Children returns child field instances.
func (inst *Instance) Children() []rtui.ComponentInstance {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.childInstances
}

// AddChild implements TreeContainer interface.
func (inst *Instance) AddChild(child rtui.ComponentInstance) {
	if child == nil {
		return
	}
	inst.mu.Lock()
	defer inst.mu.Unlock()

	// Check if child already exists
	for _, existing := range inst.childInstances {
		if existing == child {
			return
		}
	}
	inst.childInstances = append(inst.childInstances, child)
}

// RemoveChild implements TreeContainer interface.
func (inst *Instance) RemoveChild(child rtui.ComponentInstance) {
	if child == nil {
		return
	}
	inst.mu.Lock()
	defer inst.mu.Unlock()

	for i, existing := range inst.childInstances {
		if existing == child {
			inst.childInstances = append(inst.childInstances[:i], inst.childInstances[i+1:]...)
			// Clear parent reference on child to prevent memory leak
			if childWithParent, ok := child.(interface{ SetParent(interface{}) }); ok {
				childWithParent.SetParent(nil)
			}
			break
		}
	}
}

// ClearChildren removes all child instances.
func (inst *Instance) ClearChildren() error {
	inst.mu.Lock()
	defer inst.mu.Unlock()

	// Clear parent references on all children to prevent memory leak
	for _, child := range inst.childInstances {
		if childWithParent, ok := child.(interface{ SetParent(interface{}) }); ok {
			childWithParent.SetParent(nil)
		}
	}
	inst.childInstances = inst.childInstances[:0]
	return nil
}

// =============================================================================
// IntentHandler Interface (Phase 6)
// =============================================================================

// HandleIntent implements the intent.IntentHandler interface.
// This method is called when intents bubble up the instance tree.
func (inst *Instance) HandleIntent(i intent.Intent) bool {
	// Filter intents for this form
	if !inst.shouldHandleIntent(i) {
		return false
	}

	switch v := i.(type) {
	case FormFieldChangeIntent:
		inst.handleFieldChange(v)
		return true

	case FormFieldBlurIntent:
		inst.handleFieldBlur(v)
		return true

	case FormValidateIntent:
		inst.handleValidate(v)
		return true

	case FormSubmitIntent:
		inst.handleSubmit(v)
		return true

	case FormResetIntent:
		inst.handleReset()
		return true
	}

	return false
}

// shouldHandleIntent checks if an intent is for this form.
func (inst *Instance) shouldHandleIntent(i intent.Intent) bool {
	// Extract form ID from intent if possible
	var formID string
	switch v := i.(type) {
	case FormFieldChangeIntent:
		formID = v.FormID
	case FormFieldBlurIntent:
		formID = v.FormID
	case FormValidateIntent:
		formID = v.FormID
	case FormSubmitIntent:
		formID = v.FormID
	case FormResetIntent:
		formID = v.FormID
	default:
		return false
	}

	return formID == inst.key
}

// handleFieldChange processes a field value change.
func (inst *Instance) handleFieldChange(intent FormFieldChangeIntent) {
	inst.mu.Lock()
	defer inst.mu.Unlock()

	// Update value
	inst.values[intent.Field] = intent.Value

	// Mark form as dirty if field changed
	if intent.IsDirty {
		inst.dirty = true
	}

	// Clear error for this field
	delete(inst.errors, intent.Field)
}

// handleFieldBlur processes a field blur event (trigger validation).
func (inst *Instance) handleFieldBlur(intent FormFieldBlurIntent) {
	inst.mu.Lock()
	defer inst.mu.Unlock()

	// Update value
	inst.values[intent.Field] = intent.Value

	// Validate this field
	inst.validateField(intent.Field)

	// Update overall validity
	inst.updateValidity()
}

// handleValidate processes a validation request.
func (inst *Instance) handleValidate(intent FormValidateIntent) {
	if intent.Field == "" {
		// Validate entire form
		inst.validateForm()
	} else {
		// Validate specific field
		inst.validateField(intent.Field)
	}
}

// handleSubmit processes a form submission.
func (inst *Instance) handleSubmit(submitIntent FormSubmitIntent) {
	inst.mu.Lock()

	if inst.isSubmitting {
		inst.mu.Unlock()
		return
	}

	// Validate all fields if required
	if inst.validateAll {
		inst.validateForm()

		// If form is invalid, don't submit
		if !inst.isValid {
			inst.mu.Unlock()
			return
		}
	}

	inst.isSubmitting = true
	inst.mu.Unlock()

	// Emit submit intent (if configured)
	if inst.onSubmit != nil {
		intent.Emit(inst, inst.onSubmit)
	}

	inst.mu.Lock()
	inst.isSubmitting = false
	inst.mu.Unlock()
}

// handleReset processes a form reset.
func (inst *Instance) handleReset() {
	inst.mu.Lock()
	defer inst.mu.Unlock()

	// Reset values to initial values only
	// Create a new values map with only initial values
	newValues := make(map[string]interface{}, len(inst.initialValues))
	for k, v := range inst.initialValues {
		newValues[k] = v
	}
	inst.values = newValues

	// Clear errors
	inst.errors = make(map[string]string)

	// Reset validity
	inst.isValid = true

	// Reset dirty state (form is clean after reset)
	inst.dirty = false

	// Emit reset intent (if configured)
	if inst.onReset != nil {
		intent.Emit(inst, inst.onReset)
	}
}

// validateField validates a single field.
func (inst *Instance) validateField(field string) {
	_, exists := inst.values[field]
	if !exists {
		return
	}

	// TODO: Implement field validators
	// For now, mark field as valid
	delete(inst.errors, field)
}

// validateForm validates all fields.
func (inst *Instance) validateForm() {
	// Clear current errors
	inst.errors = make(map[string]string)

	// Validate each field
	for field := range inst.values {
		inst.validateField(field)
	}

	// Update overall validity
	inst.updateValidity()
}

// updateValidity updates isDirty and isValid flags.
func (inst *Instance) updateValidity() {
	inst.isValid = len(inst.errors) == 0
}

// =============================================================================
// Public API Methods
// =============================================================================

// GetValue returns the value of a field.
func (inst *Instance) GetValue(field string) (interface{}, bool) {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	value, exists := inst.values[field]
	return value, exists
}

// SetValue sets the value of a field.
func (inst *Instance) SetValue(field string, value interface{}) {
	inst.mu.Lock()
	defer inst.mu.Unlock()
	inst.values[field] = value
	inst.dirty = true
	// Clear error for this field
	delete(inst.errors, field)
}

// GetValues returns all field values.
func (inst *Instance) GetValues() map[string]interface{} {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	values := make(map[string]interface{}, len(inst.values))
	for k, v := range inst.values {
		values[k] = v
	}
	return values
}

// SetValues sets multiple field values.
func (inst *Instance) SetValues(values map[string]interface{}) {
	inst.mu.Lock()
	defer inst.mu.Unlock()
	for k, v := range values {
		inst.values[k] = v
	}
	inst.dirty = true
}

// GetError returns the validation error for a field.
func (inst *Instance) GetError(field string) (string, bool) {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	error, exists := inst.errors[field]
	return error, exists
}

// GetErrors returns all validation errors.
func (inst *Instance) GetErrors() map[string]string {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	errors := make(map[string]string, len(inst.errors))
	for k, v := range inst.errors {
		errors[k] = v
	}
	return errors
}

// IsValid returns whether the form is valid (no validation errors).
func (inst *Instance) IsValid() bool {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.isValid
}

// IsSubmitting returns whether the form is currently being submitted.
func (inst *Instance) IsSubmitting() bool {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.isSubmitting
}

// =============================================================================
// PaintableInstance Interface (Minimal Implementation)
// =============================================================================

// Paint renders the form. Forms typically don't paint directly,
// they rely on child components for rendering.
func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	// Forms are containers, rendering is handled by children
	// Optionally render label if set
	if inst.label != "" {
		return []paint.DrawCmd{
			{
				X:     x,
				Y:     y,
				Text:  inst.label,
				Style: inst.formStyle,
			},
		}
	}
	return nil
}

// =============================================================================
// Measurable Interface
// =============================================================================

// Measure implements layout.Measurable interface.
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	// Forms delegate measurement to children
	// Return a reasonable default size
	return layout.Size{
		Width:  constraints.MaxWidth,
		Height: constraints.MaxHeight,
	}
}

// =============================================================================
// Helper Props Extraction
// =============================================================================

func getStringProp(props rtui.Props, key string, defaultValue string) string {
	if v, ok := props[key].(string); ok {
		return v
	}
	return defaultValue
}

func getBoolProp(props rtui.Props, key string, defaultValue bool) bool {
	if v, ok := props[key].(bool); ok {
		return v
	}
	return defaultValue
}

func getStyleProp(props rtui.Props) style.Style {
	if v, ok := props["style"].(style.Style); ok {
		return v
	}
	return style.Style{}
}

// =============================================================================
// Validation Integration (Placeholder)
// =============================================================================

// AddValidator adds a validator for a field.
// This is a placeholder for future validation integration.
func (inst *Instance) AddValidator(field string, validator validation.Validator) {
	// TODO: Implement field-level validators
}

// RemoveValidator removes a validator from a field.
// This is a placeholder for future validation integration.
func (inst *Instance) RemoveValidator(field string, validator validation.Validator) {
	// TODO: Implement field-level validators
}
