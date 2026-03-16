package form

import (
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestMain resets the global form registry before each test run for isolation.
func TestMain(m *testing.M) {
	ResetRegistry()
	m.Run()
}

// =============================================================================
// Form Context Tests (Phase 2)
// =============================================================================

// TestFormContext_Registration tests form instance registration and cleanup.
func TestFormContext_Registration(t *testing.T) {
	t.Run("Register and unregister form", func(t *testing.T) {
		props := rtui.Props{
			"key":   "test-form-reg",
			"label": "Test Form",
		}
		inst := NewInstance(props)

		// OnMount should register the form
		inst.OnMount()

		// Verify form is registered
		form := GetForm("test-form-reg")
		if form == nil {
			t.Error("Form should be registered after OnMount")
		}
		if form != inst {
			t.Error("Registered form should be the same instance")
		}

		// OnUnmount should unregister the form
		inst.OnUnmount()

		// Verify form is unregistered
		form = GetForm("test-form-reg")
		if form != nil {
			t.Error("Form should be unregistered after OnUnmount")
		}
	})

	t.Run("Multiple forms can coexist", func(t *testing.T) {
		form1Props := rtui.Props{
			"key":   "form-1",
			"label": "Form 1",
		}
		form2Props := rtui.Props{
			"key":   "form-2",
			"label": "Form 2",
		}

		inst1 := NewInstance(form1Props)
		inst2 := NewInstance(form2Props)

		// Mount both forms
		inst1.OnMount()
		inst2.OnMount()

		// Verify both are registered
		form1 := GetForm("form-1")
		form2 := GetForm("form-2")

		if form1 == nil || form1 != inst1 {
			t.Error("Form 1 should be registered")
		}
		if form2 == nil || form2 != inst2 {
			t.Error("Form 2 should be registered")
		}

		// Unmount one form
		inst1.OnUnmount()

		// Verify form1 is unregistered but form2 remains
		form1 = GetForm("form-1")
		form2 = GetForm("form-2")

		if form1 != nil {
			t.Error("Form 1 should be unregistered")
		}
		if form2 == nil || form2 != inst2 {
			t.Error("Form 2 should still be registered")
		}

		// Cleanup
		inst2.OnUnmount()
	})
}

// TestFormContext_GetFormContext tests the GetFormContext helper function.
func TestFormContext_GetFormContext(t *testing.T) {
	t.Run("GetFormContext returns valid context", func(t *testing.T) {
		props := rtui.Props{
			"key":   "test-form-ctx",
			"label": "Test Form",
			"values": map[string]interface{}{
				"username": "initial_user",
				"email":    "initial@email.com",
			},
		}
		inst := NewInstance(props)

		inst.OnMount()
		defer inst.OnUnmount()

		// Get FormContext
		ctx := GetFormContext("test-form-ctx")
		if ctx == nil {
			t.Fatal("GetFormContext should return non-nil context")
		}

		// Test GetValue
		value, exists := ctx.GetValue("username")
		if !exists {
			t.Error("GetValue should find username")
		}
		if value != "initial_user" {
			t.Errorf("GetValue should return initial_user, got %v", value)
		}

		// Test SetValue
		ctx.SetValue("username", "new_user")
		newValue, exists := ctx.GetValue("username")
		if !exists || newValue != "new_user" {
			t.Error("SetValue should update the value")
		}
		if !ctx.IsFieldDirty("username") {
			t.Error("IsFieldDirty should return true after field value changes")
		}
		if ctx.IsFieldTouched("username") {
			t.Error("IsFieldTouched should remain false until blur/visit is recorded")
		}

		// Test GetValues
		values := ctx.GetValues()
		if values == nil {
			t.Error("GetValues should return non-nil map")
		}
		if values["email"] != "initial@email.com" {
			t.Error("GetValues should return all initial values")
		}
	})

	t.Run("GetFormContext with non-existent form", func(t *testing.T) {
		ctx := GetFormContext("non-existent-form")
		if ctx != nil {
			t.Error("GetFormContext should return nil for non-existent form")
		}
	})

	t.Run("GetFormContext after form unmount", func(t *testing.T) {
		props := rtui.Props{
			"key": "test-form-unmount",
		}
		inst := NewInstance(props)

		inst.OnMount()
		inst.OnUnmount()

		ctx := GetFormContext("test-form-unmount")
		if ctx != nil {
			t.Error("GetFormContext should return nil after form is unmounted")
		}
	})
}

func TestFormContext_FieldState(t *testing.T) {
	props := rtui.Props{
		"key": "test-form-field-state",
		"values": map[string]interface{}{
			"email": "initial@example.com",
		},
	}
	inst := NewInstance(props)

	inst.OnMount()
	defer inst.OnUnmount()

	ctx := GetFormContext("test-form-field-state")
	if ctx == nil {
		t.Fatal("expected non-nil form context")
	}

	if ctx.IsFieldTouched("email") {
		t.Fatal("expected field to start untouched")
	}
	if ctx.IsFieldDirty("email") {
		t.Fatal("expected field to start clean")
	}

	inst.HandleIntent(FormFieldBlurIntent{
		FormID: "test-form-field-state",
		Field:  "email",
		Value:  "initial@example.com",
	})
	if !ctx.IsFieldTouched("email") {
		t.Fatal("expected blur to mark field touched")
	}
	if ctx.IsFieldDirty("email") {
		t.Fatal("expected unchanged blur value to keep field clean")
	}

	inst.HandleIntent(FormFieldChangeIntent{
		FormID:  "test-form-field-state",
		Field:   "email",
		Value:   "changed@example.com",
		IsDirty: true,
	})
	if !ctx.IsFieldDirty("email") {
		t.Fatal("expected changed value to mark field dirty")
	}
}

// TestFormContext_SetValue tests SetValue through context.
func TestFormContext_SetValue(t *testing.T) {
	t.Run("SetValue updates form values", func(t *testing.T) {
		props := rtui.Props{
			"key":   "test-form-setvalue",
			"label": "Test Form",
			"values": map[string]interface{}{
				"field1": "value1",
			},
		}
		inst := NewInstance(props)

		inst.OnMount()
		defer inst.OnUnmount()

		ctx := GetFormContext("test-form-setvalue")

		// Set new value
		ctx.SetValue("field1", "updated_value")
		ctx.SetValue("field2", "new_field_value")

		// Verify
		value, exists := ctx.GetValue("field1")
		if !exists || value != "updated_value" {
			t.Error("SetValue should update existing field")
		}

		value, exists = ctx.GetValue("field2")
		if !exists || value != "new_field_value" {
			t.Error("SetValue should add new field")
		}
	})

	t.Run("SetValues updates multiple values", func(t *testing.T) {
		props := rtui.Props{
			"key":   "test-form-setvalues",
			"label": "Test Form",
			"values": map[string]interface{}{
				"a": 1,
				"b": 2,
			},
		}
		inst := NewInstance(props)

		inst.OnMount()
		defer inst.OnUnmount()

		ctx := GetFormContext("test-form-setvalues")

		// Set multiple values
		ctx.SetValues(map[string]interface{}{
			"a": 10,
			"b": 20,
			"c": 30,
		})

		// Verify
		a, _ := ctx.GetValue("a")
		b, _ := ctx.GetValue("b")
		c, _ := ctx.GetValue("c")

		if a != 10 || b != 20 || c != 30 {
			t.Errorf("SetValues should update all fields: a=%v b=%v c=%v", a, b, c)
		}
	})
}

// TestFormContext_Errors tests error management through context.
func TestFormContext_Errors(t *testing.T) {
	t.Run("GetError and GetErrors", func(t *testing.T) {
		props := rtui.Props{
			"key":   "test-form-errors",
			"label": "Test Form",
		}
		inst := NewInstance(props)

		inst.OnMount()
		defer inst.OnUnmount()

		ctx := GetFormContext("test-form-errors")

		// Set some errors
		inst.mu.Lock()
		inst.errors["username"] = "Username is required"
		inst.errors["email"] = "Invalid email format"
		inst.isValid = false
		inst.mu.Unlock()

		// Test GetError
		err, exists := ctx.GetError("username")
		if !exists {
			t.Error("GetError should find username error")
		}
		if err != "Username is required" {
			t.Errorf("GetError should return correct error, got %s", err)
		}

		// Test GetError for non-existent field
		_, exists = ctx.GetError("nonexistent")
		if exists {
			t.Error("GetError should return false for non-existent field")
		}

		// Test GetErrors
		errors := ctx.GetErrors()
		if len(errors) != 2 {
			t.Errorf("GetErrors should return 2 errors, got %d", len(errors))
		}

		// Test IsValid
		if ctx.IsValid() {
			t.Error("IsValid should return false when there are errors")
		}

		// Clear errors
		inst.mu.Lock()
		inst.errors = make(map[string]string)
		inst.isValid = true
		inst.mu.Unlock()

		if !ctx.IsValid() {
			t.Error("IsValid should return true when there are no errors")
		}
	})
}
