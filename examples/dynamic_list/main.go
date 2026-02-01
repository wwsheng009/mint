// Package main demonstrates dynamic list with state preservation
package main

import (
	"fmt"

	"github.com/wwsheng009/mint/ui"
)

// Todo represents a todo item
type Todo struct {
	ID   int
	Text string
	Done bool
}

// TodoItem represents a single todo item component
func TodoItem(props ui.Props) ui.VNode {
	todo := props["todo"].(Todo)

	// Local state for this item - preserved across re-renders
	count, setCount, _ := ui.UseStateInt(0)

	// This simulates some local state that should be preserved
	// when the item is re-rendered (due to parent update)
	displayText := todo.Text
	if count > 0 {
		displayText = fmt.Sprintf("%s (clicked: %d)", todo.Text, count)
	}

	status := "[ ]"
	if todo.Done {
		status = "[x]"
	}

	return ui.HStack(
		ui.NewTextBuilder(fmt.Sprintf("%s %s", status, displayText)).
			FgColor(func() string {
				if todo.Done {
					return "bright-black"
				}
				return "white"
			}()).
			Build(),
		ui.ButtonBuilder(" +").
			OnClick(func() {
				setCount(func(c int) int { return c + 1 })
			}).
			Build(),
	)
}

// TodoList is the main component with dynamic list
func TodoList() ui.VNode {
	// State for todos - use the generic useState for slice
	todos := []Todo{
		{ID: 1, Text: "Buy groceries", Done: false},
		{ID: 2, Text: "Write code", Done: false},
		{ID: 3, Text: "Test the app", Done: false},
	}

	// Render todo list with keys for state preservation
	items := make([]ui.VNode, 0, len(todos)+5)
	items = append(items,
		ui.NewTextBuilder("Dynamic List Demo - State Preservation").
			FgColor("cyan").
			Bold(true).
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("Each item has local click count").
			FgColor("yellow").
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("Click + on each item, then re-run to see").
			FgColor("bright-black").
			Build(),
		ui.Text(""),
	)

	// Render todo items with proper keys
	for _, todo := range todos {
		// IMPORTANT: Use ID as key to preserve state
		// Without key, each item would reset its count when list changes
		items = append(items, ui.ComponentWithProps("TodoItem", TodoItem).
			Key(fmt.Sprintf("todo-%d", todo.ID)). // Key from data ID!
			Props(map[string]interface{}{
				"todo": todo,
			}).
			Build())
	}

	items = append(items,
		ui.Text(""),
		ui.NewTextBuilder(fmt.Sprintf("Total: %d items", len(todos))).
			FgColor("bright-black").
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("Press 'q' to quit").
			FgColor("bright-black").
			Build(),
	)

	return ui.VStack(items...)
}

func main() {
	// Enable key warnings to see proper key usage
	// TUI_DEBUG_KEYS=true go run examples/dynamic_list/main.go

	err := ui.Run(TodoList,
		ui.WithWidth(50),
		ui.WithHeight(16),
		ui.WithTitle("Dynamic List Test"),
	)
	if err != nil {
		panic(err)
	}
}
