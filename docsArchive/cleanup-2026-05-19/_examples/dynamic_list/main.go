// Package main demonstrates dynamic list with state preservation (Store 模式)
package main

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/ui"
)

// =============================================================================
// AppState - Todo 列表状态
// =============================================================================

type AppState struct {
	Todos []Todo
}

// Todo represents a todo item
type Todo struct {
	ID   int
	Text string
	Done bool
}

// TodoItemState - 每个 Todo 项的本地状态
type TodoItemState struct {
	Count int // 点击次数
}

// =============================================================================
// Intent Types
// =============================================================================

type IncrementTodoItemIntent struct {
	TodoID int
}

func (IncrementTodoItemIntent) IntentType() string { return "IncrementTodoItem" }
func (IncrementTodoItemIntent) StayPressed() bool  { return true }

// =============================================================================
// Store 初始化
// =============================================================================

// 主应用的 Store（存储 todos）
var todoListStore = store.NewStore(AppState{
	Todos: []Todo{
		{ID: 1, Text: "Buy groceries", Done: false},
		{ID: 2, Text: "Write code", Done: false},
		{ID: 3, Text: "Test the app", Done: false},
	},
})

// =============================================================================
// TodoItem 组件 - 使用独立 Store 管理局部状态
// =============================================================================

// TodoItemStoreKey - 创建或获取每个 Todo 项的独立 Store
func getOrCreateTodoItemStore(todoID int) *store.Store[TodoItemState] {
	key := fmt.Sprintf("todo-item-%d", todoID)

	// 这里简化处理 - 在实际应用中，您可以使用注册表或单例模式
	// 为演示目的，我们为每个项创建一个新的 Store
	// 注意：在生产环境中，应该有更好的 Store 生命周期管理

	if existingStore, ok := todoItemStores[key]; ok {
		return existingStore
	}

	// 创建新的 Store
	newStore := store.NewStore(TodoItemState{Count: 0})

	// 注册 Reducer
	reducer.NewBuilder[TodoItemState]().
		On(IncrementTodoItemIntent{}, func(s TodoItemState, i intent.Intent) TodoItemState {
			// 只处理匹配的 TodoID
			inc := i.(IncrementTodoItemIntent)
			if inc.TodoID == todoID {
				s.Count++
			}
			return s
		}).
		BuildAndRegister(intent.DefaultRegistry(), newStore)

	// 保存到注册表中
	if todoItemStores == nil {
		todoItemStores = make(map[string]*store.Store[TodoItemState])
	}
	todoItemStores[key] = newStore

	return newStore
}

// todoItemStores - Todo 项 Store 的注册表
var todoItemStores map[string]*store.Store[TodoItemState]

// TodoItem represents a single todo item component
func TodoItem(props ui.Props) ui.VNode {
	todo, ok := props["todo"].(Todo)
	if !ok {
		return ui.Element("text").Prop("content", "Invalid todo item").Build()
	}

	// ✅ 获取或创建该 Todo 项的独立 Store
	todoItemStore := getOrCreateTodoItemStore(todo.ID)

	// ✅ 订阅 count 状态
	count := ui.UseStoreSelector(todoItemStore, func(s TodoItemState) int { return s.Count })

	// 模拟一些本地状态，在重新渲染时保留
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
		ui.NewButtonBuilder(" +").
			// ✅ 发送带参数的 Intent
			OnPress(IncrementTodoItemIntent{TodoID: todo.ID}).
			Build(),
	)
}

// =============================================================================
// TodoList 主组件 - 使用独立 Store 管理列表
// =============================================================================

// TodoList is the main component with dynamic list
func TodoList() ui.VNode {
	// ✅ 订阅 todos 列表
	todos := ui.UseStoreSelector(todoListStore, func(s AppState) []Todo { return s.Todos })

	// 渲染 todo 列表，使用 keys 保留状态
	items := make([]ui.VNode, 0, len(todos)+5)
	items = append(items,
		ui.NewTextBuilder("Dynamic List Demo - State Preservation").
			FgColor("cyan").
			Bold(true).
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("Each item has local click count (Store 模式)").
			FgColor("yellow").
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("Click + on each item, then re-run to see").
			FgColor("bright-black").
			Build(),
		ui.Text(""),
	)

	// 渲染 todo 项，使用 proper keys
	for _, todo := range todos {
		// 重要: 使用 ID 作为 key 以保留状态
		// 没有 key，当列表变化时每个项都会重置 count
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

// =============================================================================
// Main Function
// =============================================================================

func main() {
	// 启用 key 警告以查看正确的 key 使用
	// TUI_DEBUG_KEYS=true go run examples/dynamic_list/main.go

	err := ui.Run(TodoList,
		ui.WithWidth(50),
		ui.WithHeight(16),
		ui.WithTitle("Dynamic List Test (Store 模式)"),
	)
	if err != nil {
		panic(err)
	}
}
