// Package intent provides type-safe Intent system for declarative UI actions.
package intent

// FieldIntent is an optional interface for intents that represent field state changes.
// This allows the framework to extract the field key from various intent types.
//
// Example:
//
//	type CustomFieldIntent struct {
//	    Field string
//	    Value interface{}
//	}
//
//	func (CustomFieldIntent) IntentType() string {
//	    return "CustomField"
//	}
//
//	func (c CustomFieldIntent) GetField() string {
//	    return c.Field
//	}
type FieldIntent interface {
	Intent

	// GetField returns the state key identifier for this field.
	GetField() string
}
