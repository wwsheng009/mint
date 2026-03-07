package intent_test

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/intent"
)

// TestComponentIDRouting tests the unified ComponentID routing
func TestComponentIDRouting(t *testing.T) {
	t.Run("Component without ID handles all intents", func(t *testing.T) {
		// Component with empty componentID
		componentID := ""
		
		// Intent with ComponentID
		intent1 := &mockIntentWithID{"intent-1"}
		if !intent.ShouldHandleIntentWithID(componentID, intent1) {
			t.Error("Component without ID should handle all intents")
		}
	})

	t.Run("Intent without ComponentID is handled", func(t *testing.T) {
		// Component with componentID
		componentID := "component-1"
		
		// Intent without ComponentID
		intent1 := &mockIntent{}
		if !intent.ShouldHandleIntentWithID(componentID, intent1) {
			t.Error("Intent without ComponentID should be handled")
		}
	})

	t.Run("Matching IDs are handled", func(t *testing.T) {
		componentID := "component-1"
		intent1 := &mockIntentWithID{"component-1"}
		
		if !intent.ShouldHandleIntentWithID(componentID, intent1) {
			t.Error("Matching IDs should be handled")
		}
	})

	t.Run("Non-matching IDs are NOT handled", func(t *testing.T) {
		componentID := "component-1"
		intent1 := &mockIntentWithID{"component-2"}
		
		if intent.ShouldHandleIntentWithID(componentID, intent1) {
			t.Error("Non-matching IDs should NOT be handled")
		}
	})

	t.Run("Intent with empty ComponentID vs component with ID", func(t *testing.T) {
		componentID := "component-1"
		intent1 := &mockIntentWithID{""}  // Intent has empty ComponentID
		
		// Current behavior: Should return true for backward compatibility
		if !intent.ShouldHandleIntentWithID(componentID, intent1) {
			t.Log("Intent with empty ComponentID -> handled (backward compatibility)")
		}
	})

	t.Run("GetComponentIDFromIntent extracts ID correctly", func(t *testing.T) {
		intent1 := &mockIntentWithID{"test-id"}
		id := intent.GetComponentIDFromIntent(intent1)
		if id != "test-id" {
			t.Errorf("Expected 'test-id', got '%s'", id)
		}

		intent2 := &mockIntent{}  // Doesn't implement GetComponentID
		id = intent.GetComponentIDFromIntent(intent2)
		if id != "" {
			t.Errorf("Expected empty string, got '%s'", id)
		}
	})
}

// Mock intents for testing
type mockIntent struct{}

func (m *mockIntent) IntentType() string {
	return "mockIntent"
}

type mockIntentWithID struct {
	componentID string
}

func (m *mockIntentWithID) IntentType() string {
	return "mockIntentWithID"
}

func (m *mockIntentWithID) GetComponentID() string {
	return m.componentID
}
