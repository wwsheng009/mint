// Time Travel Debug Demo - Demonstrates state history navigation
//
// This example shows:
// 1. Recording state changes with TimeTravelDebugger
// 2. Undo/Redo navigation
// 3. Jumping to specific points in history
// 4. Exporting/importing state snapshots
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/wwsheng009/mint/runtime/debug"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/ui"
)

// =============================================================================
// Application State
// =============================================================================

type AppState struct {
	Count   int
	Message string
	History []string
}

// =============================================================================
// Intent Types
// =============================================================================

type IncrementIntent struct{}

func (IncrementIntent) IntentType() string { return "Increment" }

type DecrementIntent struct{}

func (DecrementIntent) IntentType() string { return "Decrement" }

type SetMessageIntent struct {
	Message string
}

func (SetMessageIntent) IntentType() string { return "SetMessage" }

type ResetIntent struct{}

func (ResetIntent) IntentType() string { return "Reset" }

// =============================================================================
// Reducer
// =============================================================================

func appReducer(state AppState, i intent.Intent) AppState {
	now := time.Now().Format("15:04:05")
	entry := fmt.Sprintf("[%s] %s", now, i.IntentType())

	switch v := i.(type) {
	case IncrementIntent:
		state.Count++
		state.History = append(state.History, entry+": count++")
	case DecrementIntent:
		state.Count--
		state.History = append(state.History, entry+": count--")
	case SetMessageIntent:
		state.Message = v.Message
		state.History = append(state.History, entry+": msg='"+v.Message+"'")
	case ResetIntent:
		state.Count = 0
		state.Message = ""
		state.History = append(state.History, entry+": reset")
	}

	return state
}

// =============================================================================
// Main Demo
// =============================================================================

func main() {
	fmt.Println(`
╔════════════════════════════════════════════════════════════╗
║           Time Travel Debug Demo                           ║
╚════════════════════════════════════════════════════════════╝
`)

	// Create time travel debugger
	dbg := debug.NewTimeTravelDebugger[AppState](
		debug.WithMaxHistory(50),
		debug.WithApplyState(func(s AppState) {
			fmt.Printf("⏪ Applied state: Count=%d, Message='%s'\n", s.Count, s.Message)
		}),
		debug.WithOnRecord(func(s debug.Snapshot) {
			fmt.Printf("📝 Recorded [%d]: %s\n", s.Index, s.GetIntentType())
		}),
	)

	// Initial state
	state := AppState{
		Count:   0,
		Message: "Hello",
		History: []string{},
	}

	// Record initial state
	dbg.RecordWithIntent(state, nil, "initial")

	fmt.Println("=== Simulating State Changes ===\n")

	// Simulate state changes
	state = appReducer(state, IncrementIntent{})
	dbg.RecordWithIntent(state, IncrementIntent{}, "increment 1")

	state = appReducer(state, IncrementIntent{})
	dbg.RecordWithIntent(state, IncrementIntent{}, "increment 2")

	state = appReducer(state, SetMessageIntent{Message: "World"})
	dbg.RecordWithIntent(state, SetMessageIntent{}, "set message")

	state = appReducer(state, DecrementIntent{})
	dbg.RecordWithIntent(state, DecrementIntent{}, "decrement 1")

	state = appReducer(state, SetMessageIntent{Message: "Time Travel!"})
	dbg.RecordWithIntent(state, SetMessageIntent{}, "set message 2")

	state = appReducer(state, IncrementIntent{})
	dbg.RecordWithIntent(state, IncrementIntent{}, "increment 3")

	fmt.Println("\n=== Current State ===")
	fmt.Printf("Count: %d\n", state.Count)
	fmt.Printf("Message: %s\n", state.Message)

	fmt.Println("\n=== History Navigation ===\n")

	// Show history
	history := dbg.GetHistory()
	fmt.Printf("History (%d snapshots):\n", len(history))
	for i, s := range history {
		marker := " "
		if i == dbg.GetCurrentIndex() {
			marker = ">"
		}
		fmt.Printf("  %s [%d] %s - %s\n", marker, s.Index, s.GetIntentType(), s.Label)
	}

	fmt.Println("\n=== Undo Demo ===\n")

	// Undo 3 times
	for i := 0; i < 3; i++ {
		if dbg.CanUndo() {
			dbg.Undo()
		}
	}

	fmt.Println("\n=== Redo Demo ===\n")

	// Redo once
	if dbg.CanRedo() {
		dbg.Redo()
	}

	fmt.Println("\n=== Jump To Specific Point ===\n")

	// Jump to beginning
	dbg.JumpTo(0)
	currentState, _ := dbg.GetCurrentState()
	fmt.Printf("After jump to 0: Count=%d\n", currentState.Count)

	// Jump to end
	dbg.JumpTo(len(history) - 1)
	currentState, _ = dbg.GetCurrentState()
	fmt.Printf("After jump to end: Count=%d\n", currentState.Count)

	fmt.Println("\n=== Export/Import Demo ===\n")

	// Export history
	data, err := dbg.Export()
	if err != nil {
		fmt.Printf("Export error: %v\n", err)
	} else {
		fmt.Printf("Exported %d bytes of history\n", len(data))

		// Save to temp file
		tmpFile := "/tmp/mint_history.json"
		os.WriteFile(tmpFile, data, 0644)
		fmt.Printf("Saved to: %s\n", tmpFile)

		// Create new debugger and import
		dbg2 := debug.NewTimeTravelDebugger[AppState]()
		data2, _ := os.ReadFile(tmpFile)
		if err := dbg2.Import(data2); err != nil {
			fmt.Printf("Import error: %v\n", err)
		} else {
			fmt.Printf("Imported history with %d snapshots\n", len(dbg2.GetHistory()))
		}
	}

	fmt.Println("\n=== Debug Panel State ===\n")

	panelState := dbg.GetDebugPanelState()
	panelJSON, _ := json.MarshalIndent(panelState, "", "  ")
	fmt.Printf("%s\n", panelJSON)

	fmt.Println(`
╔════════════════════════════════════════════════════════════╗
║                    Demo Complete                           ║
║                                                            ║
║  Time Travel Debugger Features:                            ║
║  • Record all state changes                                ║
║  • Undo/Redo navigation                                    ║
║  • Jump to any point in history                            ║
║  • Export/Import state snapshots                           ║
║  • Integration with Store + Reducer pattern                ║
╚════════════════════════════════════════════════════════════╝
`)
}

// Note: This demo shows the API usage. In a real application:
// 1. TimeTravelDebugger would be connected to a Store
// 2. State changes would come from user interactions
// 3. A debug panel UI would be rendered for navigation
//
// Example integration:
//
//	store := store.NewStore(AppState{})
//	dbg := debug.NewTimeTravelDebugger[AppState]()
//	store.Subscribe(dbg.RecordFunc())
//
//	// In UI:
//	ui.On(UndoIntent{}, func(ctx *intent.ActionContext) {
//	    dbg.Undo()
//	})
