package form

import (
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
	"reflect"
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
//   - FormFieldChangeIntent: Track field changes and dirty state
//   - FormFieldBlurIntent: Trigger field validation
//   - FormValidateIntent: Validate form/field
//   - FormSubmitIntent: Validate and submit form data
//   - FormResetIntent: Reset form to initial state
type Instance struct {
	// === Identification ===
	key string

	// === Props (from VNode, may change each render) ===
	label       string
	layout      FormLayout
	formStyle   style.Style
	onSubmit    intent.Intent // Intent to emit on successful submit
	onReset     intent.Intent // Intent to emit on reset
	validateAll bool          // Whether to validate all fields on submit (default: true)

	// === Runtime State (managed by instance) ===
	mu               sync.RWMutex
	values           map[string]interface{} // Current field values
	initialValues    map[string]interface{} // Initial field values (for reset)
	errors           map[string]string      // Field validation errors
	touchedFields    map[string]bool        // Fields that have been blurred/visited
	dirtyFields      map[string]bool        // Fields that differ from initial values
	validators       map[string][]validation.Validator
	validatorSources map[string]map[string][]validation.Validator
	isValid          bool // Overall form validity
	isSubmitting     bool // Submission in progress
	dirty            bool // Form has unsubmitted changes
	showAllErrors    bool // Show validation errors for all fields after validate-all/submit

	// === Instance Tree (Phase 1) ===
	childInstances   []rtui.ComponentInstance
	subscribers      map[int]func(string)
	nextSubscriberID int

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
		key:              proputil.GetString(props, "key", "form"),
		label:            proputil.GetString(props, "label", ""),
		layout:           getLayoutProp(props, LayoutVertical),
		formStyle:        proputil.GetStyle(props, "style", style.Style{}),
		validateAll:      proputil.GetBool(props, "validateAll", true),
		values:           make(map[string]interface{}),
		initialValues:    make(map[string]interface{}),
		errors:           make(map[string]string),
		touchedFields:    make(map[string]bool),
		dirtyFields:      make(map[string]bool),
		validators:       make(map[string][]validation.Validator),
		validatorSources: make(map[string]map[string][]validation.Validator),
		subscribers:      make(map[int]func(string)),
		isValid:          true,
	}

	if v, ok := props[propOnSubmit].(intent.Intent); ok {
		inst.onSubmit = v
	}
	if v, ok := props[propOnReset].(intent.Intent); ok {
		inst.onReset = v
	}

	// Initialize with field values from props
	if fieldValues, ok := props[propValues].(map[string]interface{}); ok {
		inst.values = cloneValuesMap(fieldValues)
		inst.initialValues = cloneValuesMap(fieldValues)
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
	valuesChanged := false

	if v, ok := props[propLabel].(string); ok {
		if inst.label != v {
			inst.label = v
			changed = true
		}
	}
	if layout := getLayoutProp(props, inst.layout); inst.layout != layout {
		inst.layout = layout
		changed = true
	}
	// Check if style prop exists and changed
	if s, hasStyle := props[propStyle].(style.Style); hasStyle {
		if s != inst.formStyle {
			inst.formStyle = s
			changed = true
		}
	}
	if v, ok := props[propOnSubmit].(intent.Intent); ok {
		inst.onSubmit = v
	}
	if v, ok := props[propOnReset].(intent.Intent); ok {
		inst.onReset = v
	}
	if v, ok := props[propValidateAll].(bool); ok {
		inst.validateAll = v
	}
	if v, ok := props[propValues].(map[string]interface{}); ok {
		nextValues := cloneValuesMap(v)
		if !reflect.DeepEqual(inst.values, nextValues) || !reflect.DeepEqual(inst.initialValues, nextValues) {
			inst.values = cloneValuesMap(v)
			inst.initialValues = nextValues
			inst.errors = make(map[string]string)
			inst.touchedFields = make(map[string]bool)
			inst.dirtyFields = make(map[string]bool)
			inst.isValid = true
			inst.dirty = false
			inst.showAllErrors = false
			valuesChanged = true
			changed = true
		}
	}

	if changed {
		inst.dirty = inst.dirty || !valuesChanged
	}

	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	inst.mu.RLock()
	defer inst.mu.RUnlock()

	return rtui.Props{
		propKey:         inst.key,
		propLabel:       inst.label,
		propLayout:      inst.layout,
		propValidateAll: inst.validateAll,
	}
}

func (inst *Instance) MarkDirty()                         { inst.dirty = true }
func (inst *Instance) IsDirty() bool                      { return inst.dirty }
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
// Sets the parent reference on the child to enable Intent Bubble.
func (inst *Instance) AddChild(child rtui.ComponentInstance) {
	if child == nil {
		return
	}
	inst.mu.Lock()
	defer inst.mu.Unlock()

	// Check if child already exists (pointer comparison)
	for _, existing := range inst.childInstances {
		if existing == child {
			return // Already added
		}
	}

	// Add to child instances list
	inst.childInstances = append(inst.childInstances, child)

	// Set parent reference for Intent Bubble (Phase 2 fix: P0-2 in INTENT_BUBBLE_AUDIT_REPORT.md)
	// Use SetParent method for cross-package access (requires BaseComponentInstance)
	if childWithSetParent, ok := child.(interface{ SetParent(rtui.ComponentInstance) }); ok {
		childWithSetParent.SetParent(inst)
	}
}

// RemoveChild implements TreeContainer interface.
// Clears the parent reference on the removed child.
func (inst *Instance) RemoveChild(child rtui.ComponentInstance) {
	if child == nil {
		return
	}
	inst.mu.Lock()
	defer inst.mu.Unlock()

	for i, existing := range inst.childInstances {
		if existing == child {
			inst.childInstances = append(inst.childInstances[:i], inst.childInstances[i+1:]...)

			// Clear parent reference to prevent memory leak (Phase 2 fix)
			if childWithSetParent, ok := child.(interface{ SetParent(rtui.ComponentInstance) }); ok {
				childWithSetParent.SetParent(nil)
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
func (inst *Instance) handleFieldChange(changeIntent FormFieldChangeIntent) {
	inst.mu.Lock()
	inst.values[changeIntent.Field] = changeIntent.Value
	inst.syncFieldDirtyLocked(changeIntent.Field)

	// Clear error for this field
	delete(inst.errors, changeIntent.Field)
	inst.updateValidityLocked()
	inst.mu.Unlock()

	inst.notifySubscribers(changeIntent.Field)
}

// handleFieldBlur processes a field blur event (trigger validation).
func (inst *Instance) handleFieldBlur(blurIntent FormFieldBlurIntent) {
	inst.mu.Lock()
	inst.values[blurIntent.Field] = blurIntent.Value
	inst.touchedFields[blurIntent.Field] = true
	inst.syncFieldDirtyLocked(blurIntent.Field)

	// Validate this field
	inst.validateFieldLocked(blurIntent.Field)
	inst.updateValidityLocked()
	inst.mu.Unlock()

	inst.notifySubscribers(blurIntent.Field)
}

// handleValidate processes a validation request.
func (inst *Instance) handleValidate(validateIntent FormValidateIntent) {
	inst.mu.Lock()
	changedField := validateIntent.Field
	if validateIntent.Field == "" {
		// Validate entire form
		inst.showAllErrors = true
		inst.validateFormLocked()
		changedField = ""
	} else {
		// Validate specific field
		inst.validateFieldLocked(validateIntent.Field)
		inst.updateValidityLocked()
	}
	inst.mu.Unlock()

	inst.notifySubscribers(changedField)
}

// handleSubmit processes a form submission.
func (inst *Instance) handleSubmit(submitIntent FormSubmitIntent) {
	inst.mu.Lock()
	notifyField := ""

	if inst.isSubmitting {
		inst.mu.Unlock()
		return
	}

	if submitIntent.Data != nil {
		for field, value := range submitIntent.Data {
			inst.values[field] = value
			inst.syncFieldDirtyLocked(field)
			delete(inst.errors, field)
		}
		inst.updateValidityLocked()
	}
	inst.showAllErrors = true

	// Validate all fields if required
	if inst.validateAll {
		inst.validateFormLocked()

		// If form is invalid, don't submit
		if !inst.isValid {
			inst.mu.Unlock()
			inst.notifySubscribers(notifyField)
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

	inst.notifySubscribers(notifyField)
}

// handleReset processes a form reset.
func (inst *Instance) handleReset() {
	inst.mu.Lock()

	// Reset values to initial values only
	// Create a new values map with only initial values
	newValues := make(map[string]interface{}, len(inst.initialValues))
	for k, v := range inst.initialValues {
		newValues[k] = v
	}
	inst.values = newValues

	// Clear errors
	inst.errors = make(map[string]string)
	inst.touchedFields = make(map[string]bool)
	inst.dirtyFields = make(map[string]bool)
	inst.showAllErrors = false

	// Reset validity
	inst.isValid = true

	// Reset dirty state (form is clean after reset)
	inst.dirty = false

	onReset := inst.onReset
	inst.mu.Unlock()

	// Emit reset intent (if configured)
	if onReset != nil {
		intent.Emit(inst, onReset)
	}

	inst.notifySubscribers("")
}

// validateFormLocked validates all fields.
func (inst *Instance) validateFormLocked() {
	// Clear current errors
	inst.errors = make(map[string]string)

	// Validate each field
	for field := range inst.collectFieldsLocked() {
		inst.validateFieldLocked(field)
	}

	// Update overall validity
	inst.updateValidityLocked()
}

func (inst *Instance) collectFieldsLocked() map[string]struct{} {
	fields := make(map[string]struct{}, len(inst.values)+len(inst.validators)+len(inst.validatorSources))
	for field := range inst.values {
		fields[field] = struct{}{}
	}
	for field := range inst.validators {
		fields[field] = struct{}{}
	}
	for field := range inst.validatorSources {
		fields[field] = struct{}{}
	}
	return fields
}

func (inst *Instance) combinedValidatorsLocked(field string) []validation.Validator {
	var combined []validation.Validator
	if validators := inst.validators[field]; len(validators) > 0 {
		combined = append(combined, validators...)
	}
	if sources := inst.validatorSources[field]; len(sources) > 0 {
		for _, validators := range sources {
			combined = append(combined, validators...)
		}
	}
	return combined
}

// validateFieldLocked validates a single field.
func (inst *Instance) validateFieldLocked(field string) {
	validators := inst.combinedValidatorsLocked(field)
	if len(validators) == 0 {
		delete(inst.errors, field)
		return
	}

	value, exists := inst.values[field]
	if !exists {
		value = nil
	}

	for _, validator := range validators {
		if validator == nil {
			continue
		}
		if err := validator.Validate(value); err != nil {
			message := err.Error()
			if message == "" {
				message = validator.Message()
			}
			inst.errors[field] = message
			return
		}
	}

	delete(inst.errors, field)
}

// updateValidityLocked updates the form validity flags.
func (inst *Instance) updateValidityLocked() {
	inst.isValid = len(inst.errors) == 0
}

func (inst *Instance) notifySubscribers(field string) {
	inst.mu.RLock()
	callbacks := make([]func(string), 0, len(inst.subscribers))
	for _, subscriber := range inst.subscribers {
		callbacks = append(callbacks, subscriber)
	}
	inst.mu.RUnlock()

	for _, callback := range callbacks {
		if callback != nil {
			callback(field)
		}
	}
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
	inst.values[field] = value
	inst.syncFieldDirtyLocked(field)
	// Clear error for this field
	delete(inst.errors, field)
	inst.updateValidityLocked()
	inst.mu.Unlock()

	inst.notifySubscribers(field)
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
	for k, v := range values {
		inst.values[k] = v
		inst.syncFieldDirtyLocked(k)
	}
	inst.updateValidityLocked()
	inst.mu.Unlock()

	inst.notifySubscribers("")
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

// IsFieldTouched returns whether a field has been visited/blurred.
func (inst *Instance) IsFieldTouched(field string) bool {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.touchedFields[field]
}

// IsFieldDirty returns whether a field differs from its initial value.
func (inst *Instance) IsFieldDirty(field string) bool {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.dirtyFields[field]
}

// ShouldShowError returns whether a field error should be rendered.
func (inst *Instance) ShouldShowError(field string) bool {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.shouldShowErrorLocked(field)
}

// Layout returns the form's default item layout.
func (inst *Instance) Layout() FormLayout {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.layout
}

// Subscribe registers a listener for form state changes.
func (inst *Instance) Subscribe(fn func(string)) func() {
	if fn == nil {
		return func() {}
	}

	inst.mu.Lock()
	id := inst.nextSubscriberID
	inst.nextSubscriberID++
	inst.subscribers[id] = fn
	inst.mu.Unlock()

	return func() {
		inst.mu.Lock()
		delete(inst.subscribers, id)
		inst.mu.Unlock()
	}
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

// =============================================================================
// AddValidator adds a validator for a field.
func (inst *Instance) AddValidator(field string, validator validation.Validator) {
	if field == "" || validator == nil {
		return
	}

	inst.mu.Lock()
	inst.validators[field] = append(inst.validators[field], validator)
	inst.validateFieldLocked(field)
	inst.updateValidityLocked()
	inst.mu.Unlock()

	inst.notifySubscribers(field)
}

// RemoveValidator removes a validator from a field.
func (inst *Instance) RemoveValidator(field string, validator validation.Validator) {
	if field == "" || validator == nil {
		return
	}

	inst.mu.Lock()
	validators := inst.validators[field]
	filtered := validators[:0]
	removed := false
	for _, existing := range validators {
		if !removed && reflect.DeepEqual(existing, validator) {
			removed = true
			continue
		}
		filtered = append(filtered, existing)
	}
	if len(filtered) == 0 {
		delete(inst.validators, field)
	} else {
		inst.validators[field] = append([]validation.Validator(nil), filtered...)
	}
	inst.validateFieldLocked(field)
	inst.updateValidityLocked()
	inst.mu.Unlock()

	inst.notifySubscribers(field)
}

func (inst *Instance) setValidatorSource(field string, source string, validators []validation.Validator) {
	if field == "" || source == "" {
		return
	}

	inst.mu.Lock()
	if inst.validatorSources[field] == nil {
		inst.validatorSources[field] = make(map[string][]validation.Validator)
	}

	current := inst.validatorSources[field][source]
	if reflect.DeepEqual(current, validators) {
		inst.mu.Unlock()
		return
	}

	if len(validators) == 0 {
		delete(inst.validatorSources[field], source)
		if len(inst.validatorSources[field]) == 0 {
			delete(inst.validatorSources, field)
		}
	} else {
		inst.validatorSources[field][source] = append([]validation.Validator(nil), validators...)
	}

	inst.validateFieldLocked(field)
	inst.updateValidityLocked()
	inst.mu.Unlock()

	inst.notifySubscribers(field)
}

func (inst *Instance) clearValidatorSource(field string, source string) {
	inst.setValidatorSource(field, source, nil)
}

func (inst *Instance) syncFieldDirtyLocked(field string) {
	if field == "" {
		return
	}

	currentValue, hasCurrent := inst.values[field]
	initialValue, hasInitial := inst.initialValues[field]
	isDirty := hasCurrent != hasInitial || !reflect.DeepEqual(currentValue, initialValue)
	if isDirty {
		inst.dirtyFields[field] = true
	} else {
		delete(inst.dirtyFields, field)
	}
	inst.dirty = len(inst.dirtyFields) > 0
}

func (inst *Instance) shouldShowErrorLocked(field string) bool {
	if field == "" {
		return false
	}
	return inst.showAllErrors || inst.touchedFields[field]
}

func cloneValuesMap(values map[string]interface{}) map[string]interface{} {
	cloned := make(map[string]interface{}, len(values))
	for k, v := range values {
		cloned[k] = v
	}
	return cloned
}

func getLayoutProp(props rtui.Props, fallback FormLayout) FormLayout {
	if value, ok := props[propLayout].(FormLayout); ok {
		return normalizeLayout(value)
	}
	if value, ok := props[propLayout].(string); ok {
		return normalizeLayout(FormLayout(value))
	}
	return normalizeLayout(fallback)
}
