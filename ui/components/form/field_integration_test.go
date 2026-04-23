package form_test

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/intent"
	formcomp "github.com/wwsheng009/mint/ui/components/form"
	"github.com/wwsheng009/mint/ui/components/input"
)

// =============================================================================
// Input + Form Integration Tests (Phase 6)
// =============================================================================

// TestInputWithFormIntegration 测试 Input 组件与 Form 的 intent 集成。
func TestInputWithFormIntegration(t *testing.T) {
	// Arrange: Create form instance
	formInst := formcomp.NewInstance(map[string]interface{}{"key": "loginForm"})

	// Act: Simulate FormFieldChangeIntent from input (simulates intent bubble)
	changeIntent := formcomp.FormFieldChangeIntent{
		FormID: "loginForm",
		Field:  "username",
		Value:  "johndoe",
		IsDirty: true,
	}
	formInst.HandleIntent(changeIntent)

	// Assert: Form should have the updated value
	value, exists := formInst.GetValue("username")
	if !exists {
		t.Fatal("Expected username to exist in form")
	}
	if value != "johndoe" {
		t.Errorf("Expected username='johndoe', got '%v'", value)
	}
}

// TestMultipleInputsInForm 测试多个 Input 组件在同一个 Form 中的集成。
func TestMultipleInputsInForm(t *testing.T) {
	// Arrange: Create form instance
	formInst := formcomp.NewInstance(map[string]interface{}{"key": "signupForm"})

	// Act: Simulate FormFieldChangeIntent from multiple inputs
	formInst.HandleIntent(formcomp.FormFieldChangeIntent{
		FormID: "signupForm", Field: "username", Value: "testuser", IsDirty: true,
	})
	formInst.HandleIntent(formcomp.FormFieldChangeIntent{
		FormID: "signupForm", Field: "email", Value: "test@example.com", IsDirty: true,
	})
	formInst.HandleIntent(formcomp.FormFieldChangeIntent{
		FormID: "signupForm", Field: "password", Value: "securepass123", IsDirty: true,
	})

	// Assert: Form should have all values
	username, _ := formInst.GetValue("username")
	email, _ := formInst.GetValue("email")
	password, _ := formInst.GetValue("password")

	if username != "testuser" {
		t.Errorf("Expected username='testuser', got '%v'", username)
	}
	if email != "test@example.com" {
		t.Errorf("Expected email='test@example.com', got '%v'", email)
	}
	if password != "securepass123" {
		t.Errorf("Expected password='securepass123', got '%v'", password)
	}

	// Assert: Form should be dirty with all fields
	if !formInst.IsDirty() {
		t.Error("Expected form to be dirty with multiple field changes")
	}
}

// =============================================================================
// Form Validation Integration Tests (Phase 6)
// =============================================================================

// TestFormFieldValidation 测试 Form 验证与 Field Blur Intent 的集成。
func TestFormFieldValidation(t *testing.T) {
	// Arrange: Create form instance with validation
	formInst := formcomp.NewInstance(map[string]interface{}{
		"key":        "validationForm",
		"validateAll": true,
	})

	// Act: Simulate field change then blur
	formInst.HandleIntent(formcomp.FormFieldChangeIntent{
		FormID: "validationForm", Field: "email", Value: "invalid-email", IsDirty: true,
	})

	blurIntent := formcomp.FieldBlur("validationForm", "email", "invalid-email")
	formInst.HandleIntent(blurIntent)

	// Assert: Value should be in form
	value, exists := formInst.GetValue("email")
	if !exists {
		t.Fatal("Expected email to exist in form")
	}
	if value != "invalid-email" {
		t.Errorf("Expected email='invalid-email', got '%v'", value)
	}

	// Assert: Form should be valid（placeholder validation currently allows all values）
	if !formInst.IsValid() {
		t.Error("Expected form to be valid (placeholder validation)")
	}
}

// =============================================================================
// Form Isolation Tests (Phase 6)
// =============================================================================

// TestFieldIntentIsolation 测试不同 Form 之间的 Intent 隔离。
func TestFieldIntentIsolation(t *testing.T) {
	// Arrange: Create two separate forms
	form1 := formcomp.NewInstance(map[string]interface{}{"key": "form1"})
	form2 := formcomp.NewInstance(map[string]interface{}{"key": "form2"})

	// Act: Send form1's intent
	changeIntent := formcomp.FormFieldChangeIntent{
		FormID: "form1", Field: "field", Value: "value_from_form1", IsDirty: true,
	}
	form1.HandleIntent(changeIntent)
	form2.HandleIntent(changeIntent) // Try to send to both forms

	// Assert: form1 should have the value
	value1, exists1 := form1.GetValue("field")
	if !exists1 || value1 != "value_from_form1" {
		t.Errorf("form1: Expected field='value_from_form1', got '%v', exists=%v", value1, exists1)
	}

	// Assert: form2 should NOT have the value (intent isolation by formID)
	_, exists2 := form2.GetValue("field")
	if exists2 {
		t.Error("form2: Expected field to NOT exist (intent isolation failed)")
	}
}

// =============================================================================
// Form Reset Integration Tests (Phase 6)
// =============================================================================

// TestFormResetWithFields 测试 Form Reset 与 Field 组件的集成。
func TestFormResetWithFields(t *testing.T) {
	// Arrange: Create form with initial values
	formInst := formcomp.NewInstance(map[string]interface{}{
		"key":    "resetForm",
		"values": map[string]interface{}{"username": "initial"},
	})

	// Act: Change value then reset
	formInst.HandleIntent(formcomp.FormFieldChangeIntent{
		FormID: "resetForm", Field: "username", Value: "modified", IsDirty: true,
	})
	formInst.HandleIntent(formcomp.Reset("resetForm"))

	// Assert: Value should be reset to initial
	value, _ := formInst.GetValue("username")
	if value != "initial" {
		t.Errorf("Expected username='initial' after reset, got '%v'", value)
	}

	// Assert: Form should be valid after reset
	if !formInst.IsValid() {
		t.Error("Expected form to be valid after reset")
	}
}

// =============================================================================
// Form Submit Integration Tests (Phase 6)
// =============================================================================

// TestFormSubmitWithFields 测试 Form Submit 与 Field 组件的集成。
func TestFormSubmitWithFields(t *testing.T) {
	// Arrange: Create form
	formInst := formcomp.NewInstance(map[string]interface{}{"key": "submitForm"})

	// Act: Fill form and submit
	formInst.HandleIntent(formcomp.FormFieldChangeIntent{
		FormID: "submitForm", Field: "field1", Value: "value1", IsDirty: true,
	})
	formInst.HandleIntent(formcomp.FormFieldChangeIntent{
		FormID: "submitForm", Field: "field2", Value: "value2", IsDirty: true,
	})

	formData := map[string]interface{}{
		"field1": "value1",
		"field2": "value2",
	}
	formInst.HandleIntent(formcomp.Submit("submitForm", formData))

	// Assert: Form should have the data
	allValues := formInst.GetValues()
	if len(allValues) != 2 {
		t.Errorf("Expected 2 values, got %d", len(allValues))
	}
	if allValues["field1"] != "value1" || allValues["field2"] != "value2" {
		t.Error("Values don't match expected")
	}
}

// =============================================================================
// Form Field Change Intent Helper Tests (Phase 6)
// =============================================================================

// TestFormFieldChangeIntentCreation 测试 FormFieldChangeIntent 的创建和使用。
func TestFormFieldChangeIntentCreation(t *testing.T) {
	intent := formcomp.FieldChange("testForm", "username", "newvalue", true)
	if intent.IntentType() != "Form:FieldChange" {
		t.Errorf("Expected intent type='Form:FieldChange', got '%s'", intent.IntentType())
	}
	if intent.FormID != "testForm" {
		t.Errorf("Expected formID='testForm', got '%s'", intent.FormID)
	}
	if intent.Field != "username" {
		t.Errorf("Expected field='username', got '%s'", intent.Field)
	}
	if intent.Value != "newvalue" {
		t.Errorf("Expected value='newvalue', got '%v'", intent.Value)
	}
	if !intent.IsDirty {
		t.Error("Expected isDirty=true")
	}
}

// TestFormFieldBlurIntentCreation 测试 FormFieldBlurIntent 的创建和使用。
func TestFormFieldBlurIntentCreation(t *testing.T) {
	intent := formcomp.FieldBlur("testForm", "email", "test@example.com")
	if intent.IntentType() != "Form:FieldBlur" {
		t.Errorf("Expected intent type='Form:FieldBlur', got '%s'", intent.IntentType())
	}
	if intent.FormID != "testForm" {
		t.Errorf("Expected formID='testForm', got '%s'", intent.FormID)
	}
	if intent.Field != "email" {
		t.Errorf("Expected field='email', got '%s'", intent.Field)
	}
	if intent.Value != "test@example.com" {
		t.Errorf("Expected value='test@example.com', got '%v'", intent.Value)
	}
}

// =============================================================================
// Field + Form Builder Integration Tests (Phase 6)
// =============================================================================

// TestInputBuilderWithForm 测试 Input Builder 与 Form binding 的集成。
func TestInputBuilderWithForm(t *testing.T) {
	// Act: Build input with form binding
	inputVNode := input.NewBuilder().
		ForField(intent.BindField("username")).
		ForForm(intent.BindForm("loginForm")).
		Build()

	props := inputVNode.Props()

	// Assert: Props should contain formID
	formID, ok := props["formID"].(string)
	if !ok {
		t.Fatal("Expected formID to be a string in props")
	}
	if formID != "loginForm" {
		t.Errorf("Expected formID='loginForm', got '%s'", formID)
	}
}

// TestInputBuilderWithoutForm 测试 Input Builder 不使用 Form binding。
func TestInputBuilderWithoutForm(t *testing.T) {
	// Act: Build input without form binding
	inputVNode := input.NewBuilder().
		ForField(intent.BindField("username")).
		Build()

	props := inputVNode.Props()

	// Assert: Props should NOT contain formID (or should be empty)
	formID, ok := props["formID"].(string)
	if ok && formID != "" {
		t.Errorf("Expected formID to be empty or not set, got '%s'", formID)
	}
}

// =============================================================================
// Form Dirty State Management Tests (Phase 6)
// =============================================================================

// TestFormDirtyStateWithFieldChanges 测试 Form 的 dirty 状态管理。
func TestFormDirtyStateWithFieldChanges(t *testing.T) {
	formInst := formcomp.NewInstance(map[string]interface{}{"key": "testForm"})

	// Initially form should not be dirty
	if formInst.IsDirty() {
		t.Error("Expected form to be clean initially")
	}

	// Change a field - form should become dirty
	formInst.HandleIntent(formcomp.FormFieldChangeIntent{
		FormID: "testForm", Field: "field1", Value: "value1", IsDirty: true,
	})
	if !formInst.IsDirty() {
		t.Error("Expected form to be dirty after field change")
	}

	// Reset form - form should be clean
	formInst.HandleIntent(formcomp.Reset("testForm"))
	if formInst.IsDirty() {
		t.Error("Expected form to be clean after reset")
	}
}

// =============================================================================
// Multi-Form Integration Tests (Phase 6)
// =============================================================================

// TestMultipleFormsIndependent 测试多个 Form 独立工作。
func TestMultipleFormsIndependent(t *testing.T) {
	form1 := formcomp.NewInstance(map[string]interface{}{"key": "form1"})
	form2 := formcomp.NewInstance(map[string]interface{}{"key": "form2"})

	// Change form1
	form1.HandleIntent(formcomp.FormFieldChangeIntent{
		FormID: "form1", Field: "field", Value: "form1_value", IsDirty: true,
	})

	// Change form2
	form2.HandleIntent(formcomp.FormFieldChangeIntent{
		FormID: "form2", Field: "field", Value: "form2_value", IsDirty: true,
	})

	// Both forms should have independent values
	val1, _ := form1.GetValue("field")
	val2, _ := form2.GetValue("field")

	if val1 != "form1_value" {
		t.Errorf("form1: Expected 'form1_value', got '%v'", val1)
	}
	if val2 != "form2_value" {
		t.Errorf("form2: Expected 'form2_value', got '%v'", val2)
	}
}
