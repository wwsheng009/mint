package action

import (
	"testing"
	"time"
)

func TestThrottleMiddleware_DoesNotThrottleEditingActions(t *testing.T) {
	middleware := NewThrottleMiddleware(time.Hour)
	for _, actionType := range []ActionType{
		ActionInputText,
		ActionPaste,
		ActionBackspace,
		ActionDeleteChar,
		ActionCursorLeft,
		ActionCursorRight,
	} {
		t.Run(string(actionType), func(t *testing.T) {
			if got := middleware.Before(NewAction(actionType)); got == nil {
				t.Fatalf("first %s action was throttled", actionType)
			}
			if got := middleware.Before(NewAction(actionType)); got == nil {
				t.Fatalf("second %s action was throttled", actionType)
			}
		})
	}
}

func TestThrottleMiddleware_DoesNotThrottleNavigationActions(t *testing.T) {
	middleware := NewThrottleMiddleware(time.Hour)
	for _, actionType := range []ActionType{
		ActionNavigateUp,
		ActionNavigateDown,
		ActionNavigateLeft,
		ActionNavigateRight,
		ActionNavigatePageUp,
		ActionNavigatePageDown,
		ActionNavigateHome,
		ActionNavigateEnd,
	} {
		t.Run(string(actionType), func(t *testing.T) {
			if got := middleware.Before(NewAction(actionType)); got == nil {
				t.Fatalf("first %s action was throttled", actionType)
			}
			if got := middleware.Before(NewAction(actionType)); got == nil {
				t.Fatalf("second %s action was throttled", actionType)
			}
		})
	}
}

func TestThrottleMiddleware_StillThrottlesNonEditingActions(t *testing.T) {
	middleware := NewThrottleMiddleware(time.Hour)
	if got := middleware.Before(NewAction(ActionRefresh)); got == nil {
		t.Fatal("first refresh action was throttled")
	}
	if got := middleware.Before(NewAction(ActionRefresh)); got != nil {
		t.Fatal("second refresh action should be throttled")
	}
}
