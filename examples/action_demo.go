package main

import (
	"fmt"

	"github.com/wwsheng009/mint/framework/action"
)

// MockActionTarget 模拟的 ActionTarget 实现
type MockActionTarget struct {
	name string
}

func (m *MockActionTarget) HandleAction(act *action.Action) bool {
	fmt.Printf("  [%s] Handled: %s\n", m.name, act.Type)
	return true
}

func (m *MockActionTarget) GetSupportedActions() []action.ActionType {
	return []action.ActionType{
		action.ActionClick,
		action.ActionEnter,
	}
}

func (m *MockActionTarget) CanHandleAction(act *action.Action) bool {
	return act.Type == action.ActionClick || act.Type == action.ActionEnter
}

func main() {
	fmt.Println("=== Action System Demo ===\n")

	// 1. 测试 Action 创建
	fmt.Println("1. Creating Actions:")
	clickAction := action.NewAction(action.ActionClick)
	enterAction := action.NewAction(action.ActionEnter)
	fmt.Printf("  Click Action: %s\n", clickAction.Type)
	fmt.Printf("  Enter Action: %s\n", enterAction.Type)

	// 2. 测试 Router
	fmt.Println("\n2. Setting up Router:")
	router := action.NewRouter(nil)

	target1 := &MockActionTarget{name: "Button"}
	router.RegisterTarget(1, target1)
	fmt.Println("  Registered Button with ID=1")

	// 3. 测试中间件
	fmt.Println("\n3. Setting up Middleware:")

	loggingMW := action.NewLoggingMiddleware()
	loggingMW.SetEnabled(true)
	fmt.Printf("  ✓ Logging Middleware\n")

	throttleMW := action.NewThrottleMiddleware(0) // 0 = 不节流，方便测试
	fmt.Printf("  ✓ Throttle Middleware (disabled for demo)\n")

	metricsMW := action.NewMetricsMiddleware()
	fmt.Printf("  ✓ Metrics Middleware\n")

	chain := action.NewMiddlewareChain(loggingMW, throttleMW, metricsMW)
	router.SetMiddleware(chain)
	fmt.Printf("  ✓ Middleware chain configured\n")

	// 4. 测试 Action 分发
	fmt.Println("\n4. Dispatching Actions:")

	clickAction.TargetID = 1
	result := router.Dispatch(clickAction)
	fmt.Printf("  Result: Handled=%v\n", result.Handled)

	// 5. 测试多次分发
	fmt.Println("\n5. Multiple Dispatches:")
	for i := 0; i < 3; i++ {
		act := action.NewAction(action.ActionEnter)
		act.TargetID = 1
		router.Dispatch(act)
	}

	// 6. 测试指标
	fmt.Println("\n6. Metrics:")
	allCounts := metricsMW.GetAllActionCounts()
	for actionType, count := range allCounts {
		if count > 0 {
			fmt.Printf("  %s: %d times\n", actionType, count)
		}
	}

	// 7. 测试预配置中间件链
	fmt.Println("\n7. Built-in Middleware Chains:")

	defaultChain := action.DefaultMiddlewareChain()
	fmt.Printf("  Default: %d middlewares\n", len(defaultChain.Middlewares()))

	debugChain := action.DebugMiddlewareChain()
	fmt.Printf("  Debug: %d middlewares\n", len(debugChain.Middlewares()))

	prodChain := action.ProductionMiddlewareChain()
	fmt.Printf("  Production: %d middlewares\n", len(prodChain.Middlewares()))

	fmt.Println("\n=== Demo Complete ===")
	fmt.Println("\nEnvironment Variables:")
	fmt.Println("  ACTION_DEBUG=true  Enable debug logging")
	fmt.Println("  ACTION_PROD=true    Use production middleware")
}
