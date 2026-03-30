package mcp

import (
	"strings"
	"testing"
)

func TestValidateLocator(t *testing.T) {
	if err := validateLocator(""); err == nil {
		t.Error("expected error for empty locator")
	}
	if err := validateLocator("#my-button"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	long := strings.Repeat("a", maxLocatorLen+1)
	if err := validateLocator(long); err == nil {
		t.Error("expected error for too-long locator")
	}
}

func TestValidateSelector(t *testing.T) {
	if err := validateSelector(""); err == nil {
		t.Error("expected error for empty selector")
	}
	if err := validateSelector(".Button"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	long := strings.Repeat("a", maxSelectorLen+1)
	if err := validateSelector(long); err == nil {
		t.Error("expected error for too-long selector")
	}
}

func TestValidateDirection(t *testing.T) {
	for _, d := range validDirections {
		if err := validateDirection(d); err != nil {
			t.Errorf("direction %q: %v", d, err)
		}
	}
	if err := validateDirection(""); err == nil {
		t.Error("expected error for empty direction")
	}
	if err := validateDirection("invalid"); err == nil {
		t.Error("expected error for invalid direction")
	}
}

func TestValidateTreeKind(t *testing.T) {
	for _, k := range validTreeKinds {
		if err := validateTreeKind(k); err != nil {
			t.Errorf("kind %q: %v", k, err)
		}
	}
	if err := validateTreeKind(""); err == nil {
		t.Error("expected error for empty kind")
	}
	if err := validateTreeKind("invalid"); err == nil {
		t.Error("expected error for invalid kind")
	}
}

func TestValidateInputText(t *testing.T) {
	if err := validateInputText("hello"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	long := strings.Repeat("a", maxInputTextLen+1)
	if err := validateInputText(long); err == nil {
		t.Error("expected error for too-long input text")
	}
}

func TestValidatePropKey(t *testing.T) {
	if err := validatePropKey(""); err == nil {
		t.Error("expected error for empty key")
	}
	if err := validatePropKey("label"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	long := strings.Repeat("k", maxPropKeyLen+1)
	if err := validatePropKey(long); err == nil {
		t.Error("expected error for too-long key")
	}
}

func TestValidateFieldName(t *testing.T) {
	if err := validateFieldName(""); err == nil {
		t.Error("expected error for empty field")
	}
	if err := validateFieldName("email"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	long := strings.Repeat("f", maxFieldNameLen+1)
	if err := validateFieldName(long); err == nil {
		t.Error("expected error for too-long field")
	}
}

func TestValidateActionType(t *testing.T) {
	if err := validateActionType(""); err == nil {
		t.Error("expected error for empty action_type")
	}
	if err := validateActionType("click"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	long := strings.Repeat("a", maxActionTypeLen+1)
	if err := validateActionType(long); err == nil {
		t.Error("expected error for too-long action_type")
	}
}

func TestValidateBatchOperation(t *testing.T) {
	tests := []struct {
		name    string
		op      BatchOperation
		wantErr bool
	}{
		{"valid click", BatchOperation{Operation: "click", Locator: "#btn"}, false},
		{"valid navigate", BatchOperation{Operation: "navigate", Direction: "next"}, false},
		{"valid input", BatchOperation{Operation: "input", Locator: "#inp", Text: "hello"}, false},
		{"valid set_value", BatchOperation{Operation: "set_value", Locator: "#inp", Value: "x"}, false},
		{"valid set_prop", BatchOperation{Operation: "set_prop", Locator: "#inp", Key: "label", Value: "y"}, false},
		{"valid set_form_field", BatchOperation{Operation: "set_form_field", Locator: "#form", Field: "name", Value: "z"}, false},
		{"valid select", BatchOperation{Operation: "select", Locator: "#sel", Value: 0}, false},
		{"valid dispatch", BatchOperation{Operation: "dispatch", Locator: "#btn", ActionType: "click"}, false},
		{"unknown op", BatchOperation{Operation: "unknown"}, true},
		{"click missing locator", BatchOperation{Operation: "click"}, true},
		{"navigate invalid dir", BatchOperation{Operation: "navigate", Direction: "sideways"}, true},
		{"input missing locator", BatchOperation{Operation: "input", Text: "hi"}, true},
		{"set_prop missing key", BatchOperation{Operation: "set_prop", Locator: "#x"}, true},
		{"set_form_field missing field", BatchOperation{Operation: "set_form_field", Locator: "#x"}, true},
		{"dispatch missing action_type", BatchOperation{Operation: "dispatch", Locator: "#x"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateBatchOperation(tt.op)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateBatchOperation(%+v) = %v, wantErr %v", tt.op, err, tt.wantErr)
			}
		})
	}
}
