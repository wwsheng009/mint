package form

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/validation"
)

// =============================================================================
// Intent Bubble Tests (Phase 6)
// =============================================================================

// TestFormInstance_HandleIntent_FieldChange tests FormFieldChangeIntent handling.
func TestFormInstance_HandleIntent_FieldChange(t *testing.T) {
	// Arrange
	props := rtui.Props{
		"key":   "testForm",
		"label": "Test Form",
	}
	inst := NewInstance(props)

	// Act: Emit field change intent
	changeIntent := FormFieldChangeIntent{
		FormID:  "testForm",
		Field:   "username",
		Value:   "johndoe",
		IsDirty: true,
	}
	inst.HandleIntent(changeIntent)

	// Assert
	value, exists := inst.GetValue("username")
	if !exists {
		t.Fatal("Expected username value to exist")
	}
	if value != "johndoe" {
		t.Errorf("Expected username='johndoe', got '%v'", value)
	}

	// Check dirty state
	if !inst.IsDirty() {
		t.Error("Expected form to be dirty after field change")
	}

	// Assert: No error for changed field
	_, hasError := inst.GetError("username")
	if hasError {
		t.Error("Expected no error for field after change")
	}
}

// TestFormInstance_HandleIntent_FieldBlur tests FormFieldBlurIntent handling.
func TestFormInstance_HandleIntent_FieldBlur(t *testing.T) {
	// Arrange
	props := rtui.Props{
		"key": "testForm",
	}
	inst := NewInstance(props)

	// Act: Emit field blur intent
	blurIntent := FormFieldBlurIntent{
		FormID: "testForm",
		Field:  "email",
		Value:  "invalid-email",
	}
	inst.HandleIntent(blurIntent)

	// Assert: Value should be updated
	value, exists := inst.GetValue("email")
	if !exists {
		t.Fatal("Expected email value to exist")
	}
	if value != "invalid-email" {
		t.Errorf("Expected email='invalid-email', got '%v'", value)
	}

	// Assert: Form should be valid (placeholder validation)
	if !inst.IsValid() {
		t.Error("Expected form to be valid")
	}

	if !inst.IsFieldTouched("email") {
		t.Error("Expected email field to be marked touched after blur")
	}
	if !inst.IsFieldDirty("email") {
		t.Error("Expected email field to be marked dirty after blur")
	}
}

func TestFormInstance_FieldStateTracking(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"key":    "profileForm",
		"values": map[string]interface{}{"email": "initial@example.com"},
	})

	if inst.IsFieldTouched("email") {
		t.Fatal("expected field to start untouched")
	}
	if inst.IsFieldDirty("email") {
		t.Fatal("expected field to start clean")
	}

	inst.HandleIntent(FormFieldChangeIntent{
		FormID:  "profileForm",
		Field:   "email",
		Value:   "edited@example.com",
		IsDirty: true,
	})

	if inst.IsFieldTouched("email") {
		t.Fatal("field change should not mark field touched")
	}
	if !inst.IsFieldDirty("email") {
		t.Fatal("field change should mark field dirty when value differs from initial")
	}

	inst.HandleIntent(FormFieldBlurIntent{
		FormID: "profileForm",
		Field:  "email",
		Value:  "edited@example.com",
	})

	if !inst.IsFieldTouched("email") {
		t.Fatal("field blur should mark field touched")
	}

	inst.HandleIntent(FormFieldChangeIntent{
		FormID:  "profileForm",
		Field:   "email",
		Value:   "initial@example.com",
		IsDirty: true,
	})

	if inst.IsFieldDirty("email") {
		t.Fatal("field should return to clean state when value matches initial value again")
	}
}

func TestFormInstance_FieldStateCollections(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"key": "profileForm",
		"values": map[string]interface{}{
			"email": "initial@example.com",
			"name":  "Mint",
			"role":  "user",
		},
	})

	inst.HandleIntent(FormFieldBlurIntent{
		FormID: "profileForm",
		Field:  "name",
		Value:  "Mint",
	})
	inst.HandleIntent(FormFieldChangeIntent{
		FormID:  "profileForm",
		Field:   "role",
		Value:   "admin",
		IsDirty: true,
	})
	inst.HandleIntent(FormFieldBlurIntent{
		FormID: "profileForm",
		Field:  "email",
		Value:  "changed@example.com",
	})

	touched := inst.GetTouchedFields()
	if len(touched) != 2 || touched[0] != "email" || touched[1] != "name" {
		t.Fatalf("touched fields = %v, want [email name]", touched)
	}

	dirty := inst.GetDirtyFields()
	if len(dirty) != 2 || dirty[0] != "email" || dirty[1] != "role" {
		t.Fatalf("dirty fields = %v, want [email role]", dirty)
	}
}

func TestFormInstance_SubmittedFieldStateTracking(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"key": "profileForm",
		"values": map[string]interface{}{
			"email": "",
			"name":  "Mint",
		},
	})
	inst.AddValidator("email", validation.Required())

	inst.HandleIntent(FormValidateIntent{FormID: "profileForm"})
	if !inst.IsFieldSubmitted("email") || !inst.IsFieldSubmitted("name") {
		t.Fatal("expected validate-all to mark current fields as submitted")
	}
	if inst.HasSubmitted() || inst.GetSubmitCount() != 0 {
		t.Fatal("expected validate-all to not count as form submit")
	}
	submitted := inst.GetSubmittedFields()
	if len(submitted) != 2 || submitted[0] != "email" || submitted[1] != "name" {
		t.Fatalf("submitted fields = %v, want [email name]", submitted)
	}

	inst.SetProps(rtui.Props{
		"values": map[string]interface{}{
			"email": "server@example.com",
		},
	})
	if got := inst.GetSubmittedFields(); len(got) != 0 {
		t.Fatalf("expected prop-driven sync to clear submitted fields, got %v", got)
	}

	inst.HandleIntent(FormSubmitIntent{
		FormID: "profileForm",
		Data: map[string]interface{}{
			"email": "server@example.com",
			"role":  "admin",
		},
	})
	if !inst.HasSubmitted() || inst.GetSubmitCount() != 1 {
		t.Fatalf("submit status = (%v,%d), want (true,1)", inst.HasSubmitted(), inst.GetSubmitCount())
	}
	submitted = inst.GetSubmittedFields()
	if len(submitted) != 2 || submitted[0] != "email" || submitted[1] != "role" {
		t.Fatalf("submitted fields after submit = %v, want [email role]", submitted)
	}

	inst.HandleIntent(Reset("profileForm"))
	if inst.HasSubmitted() || inst.GetSubmitCount() != 0 {
		t.Fatalf("expected reset to clear submit status, got (%v,%d)", inst.HasSubmitted(), inst.GetSubmitCount())
	}
	if got := inst.GetSubmittedFields(); len(got) != 0 {
		t.Fatalf("expected reset to clear submitted fields, got %v", got)
	}
}

func TestFormInstance_SubmitCountTracksInvalidAttempts(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"key": "profileForm",
		"values": map[string]interface{}{
			"email": "",
		},
	})
	inst.AddValidator("email", validation.Required())

	inst.HandleIntent(FormSubmitIntent{
		FormID: "profileForm",
		Data: map[string]interface{}{
			"email": "",
		},
	})
	if !inst.HasSubmitted() || inst.GetSubmitCount() != 1 {
		t.Fatalf("invalid submit status = (%v,%d), want (true,1)", inst.HasSubmitted(), inst.GetSubmitCount())
	}

	inst.HandleIntent(FormSubmitIntent{
		FormID: "profileForm",
		Data: map[string]interface{}{
			"email": "ok@example.com",
		},
	})
	if inst.GetSubmitCount() != 2 {
		t.Fatalf("submitCount = %d, want 2", inst.GetSubmitCount())
	}
}

// TestFormInstance_HandleIntent_Validate tests FormValidateIntent handling.
func TestFormInstance_HandleIntent_Validate(t *testing.T) {
	// Arrange
	props := rtui.Props{
		"key": "testForm",
	}
	inst := NewInstance(props)

	// Set initial values
	inst.SetValue("name", "Test User")
	inst.SetValue("email", "test@example.com")

	// Act: Validate entire form
	validateIntent := FormValidateIntent{
		FormID: "testForm",
	}
	inst.HandleIntent(validateIntent)

	// Assert: Form should be valid
	if !inst.IsValid() {
		t.Error("Expected form to be valid")
	}

	// Assert: No errors
	errors := inst.GetErrors()
	if len(errors) > 0 {
		t.Errorf("Expected no errors, got %d: %v", len(errors), errors)
	}
}

// TestFormInstance_HandleIntent_ValidateField tests validating a specific field.
func TestFormInstance_HandleIntent_ValidateField(t *testing.T) {
	// Arrange
	props := rtui.Props{
		"key": "testForm",
	}
	inst := NewInstance(props)

	inst.SetValue("username", "testuser")

	// Act: Validate specific field
	validateIntent := FormValidateIntent{
		FormID: "testForm",
		Field:  "username",
	}
	inst.HandleIntent(validateIntent)

	// Assert: No error for this field
	_, hasError := inst.GetError("username")
	if hasError {
		t.Error("Expected no error for username field")
	}

	// Assert: Form should be valid
	if !inst.IsValid() {
		t.Error("Expected form to be valid")
	}
}

// TestFormInstance_HandleIntent_Submit tests FormSubmitIntent handling.
func TestFormInstance_HandleIntent_Submit(t *testing.T) {
	// Arrange
	props := rtui.Props{
		"key":   "paymentForm",
		"label": "Payment",
	}
	inst := NewInstance(props)

	// Set form data
	inst.SetValue("amount", "100")
	inst.SetValue("currency", "USD")

	// Act: Submit form
	submitIntent := FormSubmitIntent{
		FormID: "paymentForm",
		Data: map[string]interface{}{
			"amount":   "100",
			"currency": "USD",
		},
	}
	inst.HandleIntent(submitIntent)

	// Assert: Form should not be submitting (no onSubmit configured)
	if inst.IsSubmitting() {
		t.Error("Expected form not to be submitting")
	}
}

// TestFormInstance_HandleIntent_Reset tests FormResetIntent handling.
func TestFormInstance_HandleIntent_Reset(t *testing.T) {
	// Arrange: Create form with initial values
	props := rtui.Props{
		"key":    "testForm",
		"values": map[string]interface{}{"name": "Initial Name"},
	}
	inst := NewInstance(props)

	// Change values
	inst.SetValue("name", "Changed Name")
	inst.SetValue("email", "new@example.com")

	// Verify changed values before reset
	value, _ := inst.GetValue("name")
	if value != "Changed Name" {
		t.Fatalf("Expected name='Changed Name' before reset, got '%v'", value)
	}

	// Act: Reset form
	resetIntent := FormResetIntent{
		FormID: "testForm",
	}
	inst.HandleIntent(resetIntent)

	// Assert: Values should be reset to initial
	value, _ = inst.GetValue("name")
	if value != "Initial Name" {
		t.Errorf("Expected name='Initial Name' after reset, got '%v'", value)
	}

	// Assert: Email should not exist after reset (no initial value)
	_, exists := inst.GetValue("email")
	if exists {
		t.Error("Expected email to be removed after reset (no initial value)")
	}

	// Assert: All errors should be cleared
	errors := inst.GetErrors()
	if len(errors) > 0 {
		t.Errorf("Expected no errors after reset, got %d", len(errors))
	}

	// Assert: Form should be valid
	if !inst.IsValid() {
		t.Error("Expected form to be valid after reset")
	}
}

// TestFormInstance_HandleIntent_WrongForm tests that intents from different forms are ignored.
func TestFormInstance_HandleIntent_WrongForm(t *testing.T) {
	// Arrange: Create two forms with different keys
	props1 := rtui.Props{"key": "form1"}
	props2 := rtui.Props{"key": "form2"}

	_ = NewInstance(props1) // inst1 unused but needed for clarity
	inst2 := NewInstance(props2)

	// Act: Try to send form1 intent to form2
	changeIntent := FormFieldChangeIntent{
		FormID:  "form1", // Wrong form ID for inst2
		Field:   "username",
		Value:   "test",
		IsDirty: true,
	}
	handled := inst2.HandleIntent(changeIntent)

	// Assert: Should not be handled by form2
	if handled {
		t.Error("Expected form2 to ignore form1's intent")
	}

	// Assert: form2 should not have the value
	_, exists := inst2.GetValue("username")
	if exists {
		t.Error("Expected form2 not to have username value")
	}
}

// TestFormInstance_GetSetValues tests getting and setting values.
func TestFormInstance_GetSetValues(t *testing.T) {
	// Arrange
	props := rtui.Props{"key": "testForm"}
	inst := NewInstance(props)

	// Act & Assert: Set single value
	inst.SetValue("field1", "value1")
	value, exists := inst.GetValue("field1")
	if !exists || value != "value1" {
		t.Errorf("Expected field1='value1', got '%v', exists=%v", value, exists)
	}

	// Act & Assert: Set multiple values
	inst.SetValues(map[string]interface{}{
		"field2": "value2",
		"field3": "value3",
	})
	value, exists = inst.GetValue("field2")
	if !exists || value != "value2" {
		t.Errorf("Expected field2='value2', got '%v', exists=%v", value, exists)
	}

	value, exists = inst.GetValue("field3")
	if !exists || value != "value3" {
		t.Errorf("Expected field3='value3', got '%v', exists=%v", value, exists)
	}

	// Act & Assert: Get all values
	allValues := inst.GetValues()
	if len(allValues) != 3 {
		t.Errorf("Expected 3 values, got %d", len(allValues))
	}
	if allValues["field1"] != "value1" || allValues["field2"] != "value2" || allValues["field3"] != "value3" {
		t.Error("Values don't match expected")
	}
}

func TestFormInstance_SetPropsSyncsValues(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"key":    "testForm",
		"values": map[string]interface{}{"username": "initial"},
	})

	inst.HandleIntent(FormFieldChangeIntent{
		FormID:  "testForm",
		Field:   "username",
		Value:   "edited",
		IsDirty: true,
	})
	inst.HandleIntent(FormFieldBlurIntent{
		FormID: "testForm",
		Field:  "username",
		Value:  "edited",
	})

	inst.mu.Lock()
	inst.errors["username"] = "stale error"
	inst.isValid = false
	inst.mu.Unlock()

	changed := inst.SetProps(rtui.Props{
		"values": map[string]interface{}{
			"username": "server",
			"email":    "server@example.com",
		},
	})

	if !changed {
		t.Fatal("expected SetProps to report a values change")
	}

	values := inst.GetValues()
	if len(values) != 2 {
		t.Fatalf("expected 2 synced values, got %d", len(values))
	}
	if values["username"] != "server" || values["email"] != "server@example.com" {
		t.Fatalf("unexpected synced values: %#v", values)
	}
	if !inst.IsValid() {
		t.Fatal("expected prop-driven value sync to clear stale validation state")
	}
	if _, ok := inst.GetError("username"); ok {
		t.Fatal("expected prop-driven value sync to clear stale field errors")
	}
	if inst.IsDirty() {
		t.Fatal("expected prop-driven value sync to reset form dirty state")
	}
	if inst.IsFieldDirty("username") {
		t.Fatal("expected prop-driven value sync to reset field dirty state")
	}
	if inst.IsFieldTouched("username") {
		t.Fatal("expected prop-driven value sync to reset field touched state")
	}

	inst.HandleIntent(FormFieldChangeIntent{
		FormID:  "testForm",
		Field:   "username",
		Value:   "edited-again",
		IsDirty: true,
	})
	inst.HandleIntent(Reset("testForm"))

	value, ok := inst.GetValue("username")
	if !ok || value != "server" {
		t.Fatalf("expected reset to use the latest synced initial value, got %v (exists=%v)", value, ok)
	}
	if inst.IsFieldDirty("username") || inst.IsFieldTouched("username") {
		t.Fatal("expected reset to clear field state metadata")
	}
}

func TestFormInstance_FieldStateRegressionMatrix(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"key": "profileForm",
		"values": map[string]interface{}{
			"email": "",
		},
	})
	inst.AddValidator("email", validation.Required())
	inst.AddValidator("email", validation.Email())

	if inst.ShouldShowError("email") {
		t.Fatal("expected pristine field to hide errors")
	}
	if inst.IsFieldTouched("email") || inst.IsFieldDirty("email") || inst.IsFieldSubmitted("email") {
		t.Fatal("expected pristine field metadata to be empty")
	}

	inst.HandleIntent(FormFieldBlurIntent{
		FormID: "profileForm",
		Field:  "email",
		Value:  "",
	})

	if !inst.IsFieldTouched("email") {
		t.Fatal("expected blur to mark field touched")
	}
	if inst.IsFieldDirty("email") {
		t.Fatal("expected unchanged blur to keep field clean")
	}
	if !inst.ShouldShowError("email") {
		t.Fatal("expected touched invalid field to show errors")
	}
	if errText, ok := inst.GetError("email"); !ok || errText == "" {
		t.Fatal("expected blur validation to record required error")
	}

	inst.HandleIntent(FormFieldChangeIntent{
		FormID:  "profileForm",
		Field:   "email",
		Value:   "invalid-email",
		IsDirty: true,
	})

	if !inst.IsFieldTouched("email") {
		t.Fatal("expected change after blur to preserve touched state")
	}
	if !inst.IsFieldDirty("email") {
		t.Fatal("expected changed value to mark field dirty")
	}
	if _, ok := inst.GetError("email"); ok {
		t.Fatal("expected change to clear stale field error before next validation")
	}
	if !inst.ShouldShowError("email") {
		t.Fatal("expected touched field to keep error visibility gate open")
	}

	inst.HandleIntent(FormValidateIntent{
		FormID: "profileForm",
		Field:  "email",
	})

	if errText, ok := inst.GetError("email"); !ok || errText == "" {
		t.Fatal("expected explicit validate to restore field error")
	}
	if inst.IsFieldSubmitted("email") {
		t.Fatal("expected field-only validate to avoid submitted metadata")
	}

	inst.HandleIntent(FormSubmitIntent{
		FormID: "profileForm",
		Data: map[string]interface{}{
			"email": "invalid-email",
		},
	})

	if !inst.HasSubmitted() || inst.GetSubmitCount() != 1 {
		t.Fatalf("submit status = (%v,%d), want (true,1)", inst.HasSubmitted(), inst.GetSubmitCount())
	}
	if !inst.IsFieldSubmitted("email") {
		t.Fatal("expected submit to mark field submitted")
	}
	if !inst.ShouldShowError("email") {
		t.Fatal("expected submitted invalid field to keep showing errors")
	}

	inst.HandleIntent(FormResetIntent{FormID: "profileForm"})

	if inst.ShouldShowError("email") {
		t.Fatal("expected reset to hide field errors again")
	}
	if inst.IsFieldTouched("email") || inst.IsFieldDirty("email") || inst.IsFieldSubmitted("email") {
		t.Fatal("expected reset to clear field metadata")
	}
	if inst.HasSubmitted() || inst.GetSubmitCount() != 0 {
		t.Fatalf("expected reset to clear submit metadata, got (%v,%d)", inst.HasSubmitted(), inst.GetSubmitCount())
	}
	if _, ok := inst.GetError("email"); ok {
		t.Fatal("expected reset to clear field errors")
	}
}

func TestFormInstance_SetPropsRebasesAcceptedControlledValue(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"key": "profileForm",
		"values": map[string]interface{}{
			"email": "initial@example.com",
		},
	})

	inst.HandleIntent(FormFieldChangeIntent{
		FormID:  "profileForm",
		Field:   "email",
		Value:   "accepted@example.com",
		IsDirty: true,
	})
	inst.HandleIntent(FormFieldBlurIntent{
		FormID: "profileForm",
		Field:  "email",
		Value:  "accepted@example.com",
	})
	inst.HandleIntent(FormSubmitIntent{
		FormID: "profileForm",
		Data: map[string]interface{}{
			"email": "accepted@example.com",
		},
	})

	if !inst.IsFieldDirty("email") || !inst.IsFieldTouched("email") || !inst.IsFieldSubmitted("email") {
		t.Fatal("expected precondition: field should be dirty, touched and submitted before rebase")
	}

	changed := inst.SetProps(rtui.Props{
		"values": map[string]interface{}{
			"email": "accepted@example.com",
		},
	})

	if !changed {
		t.Fatal("expected SetProps to treat accepted controlled value as a new baseline")
	}
	if inst.IsDirty() || inst.IsFieldDirty("email") {
		t.Fatal("expected accepted controlled value to clear dirty state")
	}
	if inst.IsFieldTouched("email") {
		t.Fatal("expected accepted controlled value to clear touched state")
	}
	if inst.IsFieldSubmitted("email") || inst.HasSubmitted() || inst.GetSubmitCount() != 0 {
		t.Fatalf("expected accepted controlled value to clear submit state, got (%v,%v,%d)", inst.IsFieldSubmitted("email"), inst.HasSubmitted(), inst.GetSubmitCount())
	}
	if inst.ShouldShowError("email") {
		t.Fatal("expected accepted controlled value to close error visibility gate")
	}

	inst.HandleIntent(Reset("profileForm"))
	value, ok := inst.GetValue("email")
	if !ok || value != "accepted@example.com" {
		t.Fatalf("expected reset to use rebased controlled value, got %v (exists=%v)", value, ok)
	}
}

// TestFormInstance_GetSetErrors tests getting and setting errors.
func TestFormInstance_GetSetErrors(t *testing.T) {
	// Arrange
	props := rtui.Props{"key": "testForm"}
	inst := NewInstance(props)

	// Set some field errors (simulate validation)
	inst.mu.Lock()
	inst.errors["email"] = "Invalid email format"
	inst.errors["age"] = "Must be at least 18"
	inst.isValid = false
	inst.mu.Unlock()

	// Act & Assert: Get single error
	error, exists := inst.GetError("email")
	if !exists || error != "Invalid email format" {
		t.Errorf("Expected email error='Invalid email format', got '%v', exists=%v", error, exists)
	}

	// Act & Assert: Get all errors
	allErrors := inst.GetErrors()
	if len(allErrors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(allErrors))
	}
	if allErrors["email"] != "Invalid email format" || allErrors["age"] != "Must be at least 18" {
		t.Error("Errors don't match expected")
	}

	// Assert: Form should be invalid
	if inst.IsValid() {
		t.Error("Expected form to be invalid with errors")
	}
}

// TestFormInstance_InstanceTree_Methods tests Instance Tree methods.
func TestFormInstance_InstanceTree_Methods(t *testing.T) {
	// Arrange
	props := rtui.Props{"key": "testForm"}
	inst := NewInstance(props)

	// Act & Assert: Parent should be nil
	if inst.Parent() != nil {
		t.Error("Expected form Parent() to return nil")
	}

	// Act & Assert: AddChild should work
	childInst := NewInstance(rtui.Props{"key": "child"})
	inst.AddChild(childInst)

	children := inst.Children()
	if len(children) != 1 {
		t.Errorf("Expected 1 child, got %d", len(children))
	}
	if children[0] != childInst {
		t.Error("Child instance doesn't match")
	}

	// Act & Assert: RemoveChild should work
	inst.RemoveChild(childInst)

	children = inst.Children()
	if len(children) != 0 {
		t.Errorf("Expected 0 children after removal, got %d", len(children))
	}
}

// TestFormInstance_Measurable tests MeasurableInstance interface.
func TestFormInstance_Measurable(t *testing.T) {
	// Arrange
	inst := NewInstance(rtui.Props{"key": "testForm"})

	// Act: Measure with max dimensions
	constraints := layout.Constraints{
		MinWidth:  100,
		MinHeight: 50,
		MaxWidth:  200,
		MaxHeight: 100,
	}
	size := inst.Measure(constraints)

	// Assert
	if size.Width != constraints.MaxWidth {
		t.Errorf("Expected width=%v, got %v", constraints.MaxWidth, size.Width)
	}
	if size.Height != constraints.MinHeight {
		t.Errorf("Expected height=%v, got %v", constraints.MinHeight, size.Height)
	}
}

func TestFormInstance_MeasureDoesNotFillMaxHeight(t *testing.T) {
	inst := NewInstance(rtui.Props{"key": "testForm"})

	size := inst.Measure(layout.Constraints{MaxWidth: 80, MaxHeight: 20})

	if size.Height != 0 {
		t.Fatalf("Measure().Height = %d, want 0 for chrome-only form container", size.Height)
	}
	if size.Width != 80 {
		t.Fatalf("Measure().Width = %d, want 80", size.Width)
	}
}

func TestFormInstance_FlexStyleUsesNaturalVerticalLayout(t *testing.T) {
	inst := NewInstance(rtui.Props{"key": "testForm", "label": "Profile"})

	flexStyle := inst.GetFlexStyle()
	if flexStyle == nil {
		t.Fatal("expected form flex style")
	}
	if flexStyle.Direction != layout.FlexColumn {
		t.Fatalf("Direction = %v, want FlexColumn", flexStyle.Direction)
	}
	if flexStyle.Padding.Top != 1 {
		t.Fatalf("Padding.Top = %d, want 1 for label row", flexStyle.Padding.Top)
	}
}

// =============================================================================
// VNode Tests
// =============================================================================

// TestFormVNode_BuilderAPI tests VNode builder API.
func TestFormVNode_BuilderAPI(t *testing.T) {
	// Act: Build form using fluent API
	vnode := NewForm("signupForm").
		Label("User Registration").
		SetValue("username", "").
		SetValue("email", "").
		ValidateAll(true).
		WithStyle(style.New())

	// Assert
	if vnode.Key() != "signupForm" {
		t.Errorf("Expected key='signupForm', got '%s'", vnode.Key())
	}

	props := vnode.Props()
	if props["label"] != "User Registration" {
		t.Errorf("Expected label='User Registration', got '%v'", props["label"])
	}
	if v, ok := props["validateAll"].(bool); !ok || !v {
		t.Error("Expected validateAll=true")
	}
}

// =============================================================================
// Intent Creation Tests
// =============================================================================

// TestFormIntents_Creation tests intent creation helpers.
func TestFormIntents_Creation(t *testing.T) {
	// Test FieldChange
	intent1 := FieldChange("testForm", "username", "newvalue", true)
	if intent1.IntentType() != "Form:FieldChange" {
		t.Errorf("Expected intent type='Form:FieldChange', got '%s'", intent1.IntentType())
	}

	// Test FieldBlur
	intent2 := FieldBlur("testForm", "email", "test@example.com")
	if intent2.IntentType() != "Form:FieldBlur" {
		t.Errorf("Expected intent type='Form:FieldBlur', got '%s'", intent2.IntentType())
	}

	// Test Validate
	intent3 := Validate("testForm", "username")
	if intent3.IntentType() != "Form:Validate" {
		t.Errorf("Expected intent type='Form:Validate', got '%s'", intent3.IntentType())
	}

	// Test Submit
	intent4 := Submit("testForm", map[string]interface{}{"field": "value"})
	if intent4.IntentType() != "Form:Submit" {
		t.Errorf("Expected intent type='Form:Submit', got '%s'", intent4.IntentType())
	}

	// Test Reset
	intent5 := Reset("testForm")
	if intent5.IntentType() != "Form:Reset" {
		t.Errorf("Expected intent type='Form:Reset', got '%s'", intent5.IntentType())
	}
}

func TestFormBindings_Creation(t *testing.T) {
	fieldBinding := BindField("username")
	if fieldBinding.GetField() != "username" {
		t.Fatalf("expected field binding to target username, got %q", fieldBinding.GetField())
	}

	formBinding := BindForm("loginForm")
	if formBinding.GetFormID() != "loginForm" {
		t.Fatalf("expected form binding to target loginForm, got %q", formBinding.GetFormID())
	}
}

func TestFormInstance_HandleIntent_SubmitUsesIntentData(t *testing.T) {
	inst := NewInstance(rtui.Props{"key": "paymentForm"})
	inst.SetValue("amount", "100")
	inst.SetValue("currency", "USD")

	inst.HandleIntent(FormSubmitIntent{
		FormID: "paymentForm",
		Data: map[string]interface{}{
			"amount":   "250",
			"currency": "CNY",
			"memo":     "updated by submit",
		},
	})

	values := inst.GetValues()
	if values["amount"] != "250" || values["currency"] != "CNY" || values["memo"] != "updated by submit" {
		t.Fatalf("expected submit payload to be reflected in form values, got %#v", values)
	}
}
