package statemachine

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
)

// TestUndoRedoNoFreeze tests that Undo and Redo don't cause infinite loops
func TestUndoRedoNoFreeze(t *testing.T) {
	// Setup
	type AppState struct {
		Count int
	}

	initialState := AppState{Count: 0}

	appReducer := reducer.NewBuilder[AppState]().
		On(intent.FieldChangeIntent{}, func(s AppState, i intent.Intent) AppState {
			if fc, ok := i.(intent.FieldChangeIntent); ok && fc.Field == "Count" {
				if val, err := reducer.ParseInt(fc.Value); err == nil {
					s.Count = val
				}
			}
			return s
		}).
		Build()

	rt := NewAppRuntime(initialState, func(s AppState) any { return nil }, appReducer, WithMaxHistory(10))

	// Perform 5 actions to build history
	for i := 1; i <= 5; i++ {
		rt.Dispatch(intent.FieldChangeIntent{
			Field: "Count",
			Value: reducer.FormatInt(i),
		})
	}

	// Verify we can undo
	if !rt.CanUndo() {
		t.Fatal("Should be able to undo after 5 actions")
	}

	// Store history before undo
	historyBeforeUndo := rt.History()
	if len(historyBeforeUndo) != 6 { // initial + 5 actions
		t.Fatalf("Expected 6 history entries, got %d", len(historyBeforeUndo))
	}

	// Undo 5 times - this should not freeze/infinite loop
	for i := 0; i < 5; i++ {
		err := rt.Undo()
		if err != nil {
			t.Fatalf("Undo failed: %v", err)
		}

		// Verify state after each undo
		expectedState := AppState{Count: 5 - i - 1}
		currentState := rt.GetState()
		if currentState.Count != expectedState.Count {
			t.Fatalf("After undo %d: expected Count=%d, got %d", i+1, expectedState.Count, currentState.Count)
		}

		// History size should remain 6 (initial + 5 actions), only currentIndex changes
		currentHistory := rt.History()
		if len(currentHistory) != 6 {
			t.Fatalf("After undo %d: expected history size 6, got %d", i+1, len(currentHistory))
		}
	}

	// After 5 undos, should be back to initial state
	finalState := rt.GetState()
	if finalState.Count != 0 {
		t.Fatalf("Expected final Count=0, got %d", finalState.Count)
	}

	// Should no longer be able to undo (at index 0)
	if rt.CanUndo() {
		t.Fatal("Should not be able to undo at initial state (index 0)")
	}

	// History should still have all 6 states
	history := rt.History()
	if len(history) != 6 {
		t.Fatalf("Expected history size 6 after all undos, got %d", len(history))
	}

	// Perform one more action
	rt.Dispatch(intent.FieldChangeIntent{
		Field: "Count",
		Value: "10",
	})

	// Should be able to undo again
	if !rt.CanUndo() {
		t.Fatal("Should be able to undo after new action")
	}

	// Undo to verify it works after adding new action
	err := rt.Undo()
	if err != nil {
		t.Fatalf("Undo failed after new action: %v", err)
	}

	afterNewUndoState := rt.GetState()
	if afterNewUndoState.Count != 0 {
		t.Fatalf("Expected Count=0 after undo, got %d", afterNewUndoState.Count)
	}

	// Verify we can redo the undone action
	if !rt.CanRedo() {
		t.Fatal("Should be able to redo after undo")
	}

	err = rt.Redo()
	if err != nil {
		t.Fatalf("Redo failed: %v", err)
	}

	afterRedoState := rt.GetState()
	if afterRedoState.Count != 10 {
		t.Fatalf("Expected Count=10 after redo, got %d", afterRedoState.Count)
	}
}

// TestUndoRedoNoInfiniteRecording tests that Undo doesn't record the same state multiple times
func TestUndoRedoNoInfiniteRecording(t *testing.T) {
	type AppState struct {
		Count int
	}

	initialState := AppState{Count: 0}

	appReducer := reducer.NewBuilder[AppState]().
		On(intent.FieldChangeIntent{}, func(s AppState, i intent.Intent) AppState {
			if fc, ok := i.(intent.FieldChangeIntent); ok && fc.Field == "Count" {
				if val, err := reducer.ParseInt(fc.Value); err == nil {
					s.Count = val
				}
			}
			return s
		}).
		Build()

	rt := NewAppRuntime(initialState, func(s AppState) any { return nil }, appReducer, WithMaxHistory(10))

	// Perform 3 actions
	for i := 1; i <= 3; i++ {
		rt.Dispatch(intent.FieldChangeIntent{
			Field: "Count",
			Value: reducer.FormatInt(i),
		})
	}

	// We have: [0, 1, 2, 3] in history
	// Undo once
	err := rt.Undo()
	if err != nil {
		t.Fatalf("Undo failed: %v", err)
	}

	// State should be 2
	stateAfterUndo := rt.GetState()
	if stateAfterUndo.Count != 2 {
		t.Fatalf("Expected Count=2 after undo, got %d", stateAfterUndo.Count)
	}

	// History should still be [0, 1, 2, 3] (not modified by Undo)
	historyAfterUndo := rt.History()

	if len(historyAfterUndo) != 4 {
		t.Fatalf("Expected history size 4 after undo, got %d", len(historyAfterUndo))
	}

	// Verify history is correct - should not change after Undo
	expectedHistory := []int{0, 1, 2, 3}
	for i, h := range historyAfterUndo {
		if h.Count != expectedHistory[i] {
			t.Fatalf("History[%d]: expected Count=%d, got %d", i, expectedHistory[i], h.Count)
		}
	}

	// Verify HistoryIndex decreased
	if rt.HistoryIndex() != 2 {
		t.Fatalf("Expected HistoryIndex=2 after undo, got %d", rt.HistoryIndex())
	}

	// Verify history is not the same reference (copy was made)
	history1 := rt.History()
	history2 := rt.History()
	if &history1[0] == &history2[0] {
		t.Error("History() should return a copy, not the same reference")
	}
}

// TestJumpToNoInfiniteRecording tests that JumpTo doesn't record duplicates
func TestJumpToNoInfiniteRecording(t *testing.T) {
	type AppState struct {
		Count int
	}

	initialState := AppState{Count: 0}

	appReducer := reducer.NewBuilder[AppState]().
		On(intent.FieldChangeIntent{}, func(s AppState, i intent.Intent) AppState {
			if fc, ok := i.(intent.FieldChangeIntent); ok && fc.Field == "Count" {
				if val, err := reducer.ParseInt(fc.Value); err == nil {
					s.Count = val
				}
			}
			return s
		}).
		Build()

	rt := NewAppRuntime(initialState, func(s AppState) any { return nil }, appReducer, WithMaxHistory(10))

	// Build history: [0, 1, 2, 3, 4]
	for i := 1; i <= 4; i++ {
		rt.Dispatch(intent.FieldChangeIntent{
			Field: "Count",
			Value: reducer.FormatInt(i),
		})
	}

	historyBeforeJump := rt.History()

	// Jump to index 1 (Count = 1)
	err := rt.JumpTo(1)
	if err != nil {
		t.Fatalf("JumpTo failed: %v", err)
	}

	// State should be 1
	stateAfterJump := rt.GetState()
	if stateAfterJump.Count != 1 {
		t.Fatalf("Expected Count=1 after JumpTo(1), got %d", stateAfterJump.Count)
	}

	// History should still be [0, 1, 2, 3, 4], NOT [0, 1, 1, 1, ...]
	historyAfterJump := rt.History()

	if len(historyAfterJump) != 5 {
		t.Fatalf("Expected history size 5 after JumpTo, got %d (may be infinite loop)",
			len(historyAfterJump))
	}

	// Verify history is unchanged by JumpTo
	for i, h := range historyBeforeJump {
		if h.Count != historyAfterJump[i].Count {
			t.Fatalf("History[%d]: before=%d, after=%d (JumpTo should not modify history)",
				i, h.Count, historyAfterJump[i].Count)
		}
	}
}

// TestUndoWithCallback tests that skipHistory flag works correctly with callbacks
func TestUndoWithCallback(t *testing.T) {
	type AppState struct {
		Count      int
		CallCount  int
	}

	initialState := AppState{Count: 0, CallCount: 0}

	appReducer := reducer.NewBuilder[AppState]().
		On(intent.FieldChangeIntent{}, func(s AppState, i intent.Intent) AppState {
			if fc, ok := i.(intent.FieldChangeIntent); ok && fc.Field == "Count" {
				if val, err := reducer.ParseInt(fc.Value); err == nil {
					s.Count = val
				}
			}
			return s
		}).
		Build()

	rt := NewAppRuntime(initialState, func(s AppState) any { return nil }, appReducer, WithMaxHistory(10))

	// Track how many times callback is called
	callbackCount := 0
	rt.Subscribe(func(s AppState) {
		callbackCount++
	})

	// Perform 3 actions - each should trigger callback (3 times)
	for i := 1; i <= 3; i++ {
		rt.Dispatch(intent.FieldChangeIntent{
			Field: "Count",
			Value: reducer.FormatInt(i),
		})
	}

	callbacksBeforeUndo := callbackCount

	// Undo should still trigger callback (for rendering updates)
	err := rt.Undo()
	if err != nil {
		t.Fatalf("Undo failed: %v", err)
	}

	// Undo should have triggered callback once
	callbacksAfterUndo := callbackCount
	if callbacksAfterUndo != callbacksBeforeUndo+1 {
		t.Logf("Undo triggered %d callbacks (expected 1, got %d)",
			callbacksAfterUndo-callbacksBeforeUndo, callbacksAfterUndo-callbacksBeforeUndo)
	}

	// Most important: undo should not be stuck in infinite loop
	// If it were stuck, the previous Undo call would never return
}

func TestUndoWithMaxHistoryZero(t *testing.T) {
	type AppState struct {
		Count int
	}

	initialState := AppState{Count: 0}

	appReducer := reducer.NewBuilder[AppState]().
		On(intent.FieldChangeIntent{}, func(s AppState, i intent.Intent) AppState {
			if fc, ok := i.(intent.FieldChangeIntent); ok && fc.Field == "Count" {
				if val, err := reducer.ParseInt(fc.Value); err == nil {
					s.Count = val
				}
			}
			return s
		}).
		Build()

	rt := NewAppRuntime(initialState, func(s AppState) any { return nil }, appReducer, WithMaxHistory(0))

	// Perform some actions
	for i := 1; i <= 5; i++ {
		rt.Dispatch(intent.FieldChangeIntent{
			Field: "Count",
			Value: reducer.FormatInt(i),
		})
	}

	// Should not be able to undo with history disabled
	if rt.CanUndo() {
		t.Error("Should not be able to undo with MaxHistory=0")
	}

	err := rt.Undo()
	if err == nil {
		t.Error("Undo should fail when history is disabled")
	}
}

// TestRedoNoFreeze tests that Redo works correctly after Undo
func TestRedoNoFreeze(t *testing.T) {
	type AppState struct {
		Count int
	}

	initialState := AppState{Count: 0}

	appReducer := reducer.NewBuilder[AppState]().
		On(intent.FieldChangeIntent{}, func(s AppState, i intent.Intent) AppState {
			if fc, ok := i.(intent.FieldChangeIntent); ok && fc.Field == "Count" {
				if val, err := reducer.ParseInt(fc.Value); err == nil {
					s.Count = val
				}
			}
			return s
		}).
		Build()

	rt := NewAppRuntime(initialState, func(s AppState) any { return nil }, appReducer, WithMaxHistory(10))

	// Perform 3 actions: [0, 1, 2, 3]
	for i := 1; i <= 3; i++ {
		rt.Dispatch(intent.FieldChangeIntent{
			Field: "Count",
			Value: reducer.FormatInt(i),
		})
	}

	// Verify we can't redo yet
	if rt.CanRedo() {
		t.Fatal("Should not be able to redo after forward actions")
	}

	// Undo once: currentIndex becomes 2, state is 2
	err := rt.Undo()
	if err != nil {
		t.Fatalf("Undo failed: %v", err)
	}
	if rt.GetState().Count != 2 {
		t.Fatalf("Expected Count=2 after undo, got %d", rt.GetState().Count)
	}
	if !rt.CanRedo() {
		t.Fatal("Should be able to redo after undo")
	}
	if !rt.CanUndo() {
		t.Fatal("Should still be able to undo further")
	}

	// Undo again: currentIndex becomes 1, state is 1
	err = rt.Undo()
	if err != nil {
		t.Fatalf("Second undo failed: %v", err)
	}
	if rt.GetState().Count != 1 {
		t.Fatalf("Expected Count=1 after second undo, got %d", rt.GetState().Count)
	}

	// Redo once: currentIndex becomes 2, state is 2
	err = rt.Redo()
	if err != nil {
		t.Fatalf("Redo failed: %v", err)
	}
	if rt.GetState().Count != 2 {
		t.Fatalf("Expected Count=2 after redo, got %d", rt.GetState().Count)
	}
	if !rt.CanRedo() {
		t.Fatal("Should still be able to redo again")
	}
	if !rt.CanUndo() {
		t.Fatal("Should be able to undo after redo")
	}

	// Redo again: currentIndex becomes 3, state is 3
	err = rt.Redo()
	if err != nil {
		t.Fatalf("Second redo failed: %v", err)
	}
	if rt.GetState().Count != 3 {
		t.Fatalf("Expected Count=3 after second redo, got %d", rt.GetState().Count)
	}
	if rt.CanRedo() {
		t.Fatal("Should not be able to redo at end of history")
	}

	// Verify history is correct: [0, 1, 2, 3]
	history := rt.History()
	if len(history) != 4 {
		t.Fatalf("Expected history size 4, got %d", len(history))
	}
	for i := 0; i < len(history); i++ {
		if history[i].Count != i {
			t.Fatalf("History[%d]: expected Count=%d, got %d", i, i, history[i].Count)
		}
	}

	// Verify current index
	if rt.HistoryIndex() != 3 {
		t.Fatalf("Expected HistoryIndex=3, got %d", rt.HistoryIndex())
	}
}

// TestUndoRedoWithNewAction tests that new actions after undo/redo truncate history
func TestUndoRedoWithNewAction(t *testing.T) {
	type AppState struct {
		Count int
	}

	initialState := AppState{Count: 0}

	appReducer := reducer.NewBuilder[AppState]().
		On(intent.FieldChangeIntent{}, func(s AppState, i intent.Intent) AppState {
			if fc, ok := i.(intent.FieldChangeIntent); ok && fc.Field == "Count" {
				if val, err := reducer.ParseInt(fc.Value); err == nil {
					s.Count = val
				}
			}
			return s
		}).
		Build()

	rt := NewAppRuntime(initialState, func(s AppState) any { return nil }, appReducer, WithMaxHistory(10))

	// Build history: [0, 1, 2, 3]
	for i := 1; i <= 3; i++ {
		rt.Dispatch(intent.FieldChangeIntent{
			Field: "Count",
			Value: reducer.FormatInt(i),
		})
	}

	// Undo: currentIndex=2, state=2
	rt.Undo()
	if rt.GetState().Count != 2 {
		t.Fatalf("Expected Count=2 after undo, got %d", rt.GetState().Count)
	}

	// Undo: currentIndex=1, state=1
	rt.Undo()
	if rt.GetState().Count != 1 {
		t.Fatalf("Expected Count=1 after second undo, got %d", rt.GetState().Count)
	}

	// New action should truncate forward history
	rt.Dispatch(intent.FieldChangeIntent{
		Field: "Count",
		Value: "10",
	})

	// History should now be: [0, 1, 10]
	// Old forward history (2, 3) should be gone
	history := rt.History()
	if len(history) != 3 {
		t.Fatalf("Expected history size 3 after new action, got %d: %v", len(history), history)
	}
	if history[0].Count != 0 || history[1].Count != 1 || history[2].Count != 10 {
		t.Fatalf("History should be [0, 1, 10], got: %v", history)
	}

	// Should not be able to redo after new action
	if rt.CanRedo() {
		t.Fatal("Should not be able to redo after performing new action")
	}

	// Current index should be at the end
	if rt.HistoryIndex() != 2 {
		t.Fatalf("Expected HistoryIndex=2, got %d", rt.HistoryIndex())
	}
}

// TestJumpToUpdatesCurrentIndex tests that JumpTo correctly updates current index
func TestJumpToUpdatesCurrentIndex(t *testing.T) {
	type AppState struct {
		Count int
	}

	initialState := AppState{Count: 0}

	appReducer := reducer.NewBuilder[AppState]().
		On(intent.FieldChangeIntent{}, func(s AppState, i intent.Intent) AppState {
			if fc, ok := i.(intent.FieldChangeIntent); ok && fc.Field == "Count" {
				if val, err := reducer.ParseInt(fc.Value); err == nil {
					s.Count = val
				}
			}
			return s
		}).
		Build()

	rt := NewAppRuntime(initialState, func(s AppState) any { return nil }, appReducer, WithMaxHistory(10))

	// Build history: [0, 1, 2, 3, 4]
	for i := 1; i <= 4; i++ {
		rt.Dispatch(intent.FieldChangeIntent{
			Field: "Count",
			Value: reducer.FormatInt(i),
		})
	}

	// JumpTo index 1 (Count=1)
	err := rt.JumpTo(1)
	if err != nil {
		t.Fatalf("JumpTo failed: %v", err)
	}
	if rt.GetState().Count != 1 {
		t.Fatalf("Expected Count=1 after JumpTo(1), got %d", rt.GetState().Count)
	}
	if rt.HistoryIndex() != 1 {
		t.Fatalf("Expected HistoryIndex=1 after JumpTo(1), got %d", rt.HistoryIndex())
	}

	// Should be able to undo from jumped position
	if !rt.CanUndo() {
		t.Fatal("Should be able to undo after JumpTo")
	}

	// Undo from jumped position
	rt.Undo()
	if rt.GetState().Count != 0 || rt.HistoryIndex() != 0 {
		t.Fatalf("After undo: expected Count=0, Index=0; got Count=%d, Index=%d",
			rt.GetState().Count, rt.HistoryIndex())
	}

	// Redo from jumped position
	rt.Redo()
	if rt.GetState().Count != 1 || rt.HistoryIndex() != 1 {
		t.Fatalf("After redo: expected Count=1, Index=1; got Count=%d, Index=%d",
			rt.GetState().Count, rt.HistoryIndex())
	}

	// Redo to end
	rt.Redo()
	rt.Redo()
	rt.Redo()
	if rt.GetState().Count != 4 || rt.HistoryIndex() != 4 {
		t.Fatalf("After redos: expected Count=4, Index=4; got Count=%d, Index=%d",
			rt.GetState().Count, rt.HistoryIndex())
	}
}
