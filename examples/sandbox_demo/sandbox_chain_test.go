package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/sandbox/mock"
	"github.com/wwsheng009/mint/ui"
)

// ChainAPIExamples 演示链式测试 API 的各种用法
func ChainAPIExamples(t *testing.T) {
	t.Run("BasicChain", testBasicChain)
	t.Run("ConditionalChain", testConditionalChain)
	t.Run("ErrorHandling", testErrorHandling)
	t.Run("ComplexScenario", testComplexScenario)
}

// testBasicChain 基础链式调用
func testBasicChain(t *testing.T) {
	testApp, err := ui.TestRun(Counter, ui.TestWithSize(40, 18))
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	helper := testApp.Helper()

	// 最简单的链式调用
	result := helper.
		Process().
		AssertRender("Count: 0").
		Result()

	if !result.OK() {
		t.Error(result.Error())
	}
}

// testConditionalChain 条件链式调用
func testConditionalChain(t *testing.T) {
	testApp, err := ui.TestRun(Counter, ui.TestWithSize(40, 18))
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	helper := testApp.Helper()

	// 先检查初始状态
	result := helper.Process().AssertRender("Count: 0").Result()
	if !result.OK() {
		t.Fatalf("Initial state check failed: %v", result.Error())
	}

	// 如果初始状态正确，继续测试
	result = helper.
		Tab().Tab().
		Enter().
		Process().
		AssertRender("Count: 1").
		Result()

	if !result.OK() {
		t.Errorf("Increment failed: %v", result.Error())
	}
}

// testErrorHandling 错误处理
func testErrorHandling(t *testing.T) {
	testApp, err := ui.TestRun(Counter, ui.TestWithSize(40, 18))
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	helper := testApp.Helper()

	// 故意制造一个错误（断言不存在的文本）
	result := helper.
		Process().
		AssertRender("This text does not exist").
		Result()

	if result.OK() {
		t.Error("Expected error but got success")
	}

	// 获取错误信息
	if err := result.Error(); err != nil {
		t.Logf("Got expected error: %v", err)
	}

	// 获取所有错误
	for i, e := range result.Errors {
		t.Logf("Error %d: %v", i, e)
	}
}

// testComplexScenario 复杂场景测试
func testComplexScenario(t *testing.T) {
	testApp, err := ui.TestRun(Counter, ui.TestWithSize(40, 18))
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	helper := testApp.Helper()

	// 完整的用户流程
	result := helper.
		// 1. 验证初始状态
		Process().
		AssertRender("Hello, Guest").
		AssertRender("Count: 0").

		// 2. 修改名字
		Tab().Tab().Tab().
		Type("Alice").
		Process().
		AssertRender("Hello, Alice").

		// 3. 增加计数 3 次
		Tab().Tab().Tab().
		Enter().
		Process().
		AssertRender("Count: 1").

		Tab().Tab().Tab().
		Enter().
		Process().
		AssertRender("Count: 2").

		Tab().Tab().Tab().
		Enter().
		Process().
		AssertRender("Count: 3").

		// 4. 减少计数 1 次
		Tab().Tab().Tab().Tab().
		Enter().
		Process().
		AssertRender("Count: 2").

		// 5. 再次修改名字
		Tab().Tab().Tab().Tab().
		Type("Bob").
		Process().
		AssertRender("Hello, Bob").

		Result()

	if !result.OK() {
		t.Errorf("Complex scenario failed: %v", result.Error())
	}
}

// testWithClearErrors 使用 ClearErrors
func testWithClearErrors(t *testing.T) {
	testApp, err := ui.TestRun(Counter, ui.TestWithSize(40, 18))
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	helper := testApp.Helper()

	// 第一个测试（可能有错误）
	helper.
		Process().
		AssertRender("Hello, Guest").
		Result()

	if helper.HasErrors() {
		t.Log("Test 1 had errors (expected or not)")
	}

	// 清除错误，继续测试
	helper.ClearErrors()

	// 第二个测试
	result := helper.
		Tab().Tab().
		Enter().
		Process().
		AssertRender("Count: 1").
		Result()

	if !result.OK() {
		t.Errorf("Test 2 failed: %v", result.Error())
	}
}

// testRepeatedActions 重复操作模式
func testRepeatedActions(t *testing.T) {
	testApp, err := ui.TestRun(Counter, ui.TestWithSize(40, 18))
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	helper := testApp.Helper()

	// 重复增加计数 10 次
	for i := 0; i < 10; i++ {
		// 导航到 "+" 按钮并点击
		result := helper.
			Tab().Tab().
			Enter().
			Process().
			AssertRender(fmt.Sprintf("Count: %d", i+1)).
			Result()

		if !result.OK() {
			t.Errorf("Increment %d failed: %v", i+1, result.Error())
			break
		}
	}

	// 最终验证
	result := helper.AssertRender("Count: 10").Result()
	if !result.OK() {
		t.Errorf("Final count check failed: %v", result.Error())
	}
}

// testNavigationPattern 导航模式
func testNavigationPattern(t *testing.T) {
	testApp, err := ui.TestRun(Counter, ui.TestWithSize(40, 18))
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	helper := testApp.Helper()

	// 测试导航到不同的元素
	elements := []struct {
		name        string
		tabCount    int
		expected    string
	}{
		{"Minus button", 2, "[ - ]"},
		{"Plus button", 3, "[ + ]"},
		{"Input field", 4, "Enter name"},
	}

	for _, elem := range elements {
		// 先回到开始（Escape 清除焦点）
		result := helper.
			Escape().
			Process().
			Result()

		if !result.OK() {
			t.Errorf("Reset failed for %s: %v", elem.name, result.Error())
			continue
		}

		// 导航到元素
		for i := 0; i < elem.tabCount; i++ {
			helper.Tab()
		}

		result = helper.Process().AssertRender(elem.expected).Result()
		if !result.OK() {
			t.Errorf("Navigation to %s failed: %v", elem.name, result.Error())
		}
	}
}

// testWaitExample 等待示例
func testWaitExample(t *testing.T) {
	testApp, err := ui.TestRun(Counter, ui.TestWithSize(40, 18))
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	helper := testApp.Helper()

	// 在操作之间等待（模拟用户思考时间）
	result := helper.
		Process().
		Wait(100 * time.Millisecond).  // 等待 100ms
		Tab().Tab().
		Wait(50 * time.Millisecond).   // 等待 50ms
		Enter().
		Process().
		AssertRender("Count: 1").
		Result()

	if !result.OK() {
		t.Errorf("Test with wait failed: %v", result.Error())
	}
}

// testKeyboardShortcuts 键盘快捷键测试
func testKeyboardShortcuts(t *testing.T) {
	testApp, err := ui.TestRun(Counter, ui.TestWithSize(40, 18))
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	helper := testApp.Helper()

	// 测试各种键盘快捷键
	shortcuts := []struct {
		name     string
		key      rune
		special  platform.SpecialKey
		expected string
	}{
		{"Tab key", 0, platform.KeyTab, ""},
		{"Enter key", 0, platform.KeyEnter, ""},
		{"Escape key", 0, platform.KeyEscape, ""},
		{"Character 'a'", 'a', 0, "a"},
		{"Character 'Z'", 'Z', 0, "Z"},
	}

	for _, sc := range shortcuts {
		result := helper.
			Process().
			Wait(50 * time.Millisecond).
			Result()

		if !result.OK() {
			t.Logf("Before %s: %v", sc.name, result.Error())
			continue
		}

		// 发送快捷键
		if sc.special != 0 {
			helper.Press(sc.special)
		} else {
			helper.PressKey(sc.key)
		}

		result = helper.Process().Result()
		if !result.OK() {
			t.Logf("After %s: %v", sc.name, result.Error())
		}
	}
}

// testMouseActions 鼠标操作测试
func testMouseActions(t *testing.T) {
	testApp, err := ui.TestRun(Counter, ui.TestWithSize(40, 18))
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	helper := testApp.Helper()

	// 测试鼠标点击
	// 注意：需要知道按钮的位置
	// 这里演示语法，实际坐标需要根据应用布局确定

	result := helper.
		Process().
		Click(10, 8).  // 点击 "+" 按钮位置（假设）
		Process().
		AssertRender("Count: 1").
		Result()

	if !result.OK() {
		t.Errorf("Mouse click test failed: %v", result.Error())
	}
}

// testCombinationActions 组合操作测试
func testCombinationActions(t *testing.T) {
	testApp, err := ui.TestRun(Counter, ui.TestWithSize(40, 18))
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	helper := testApp.Helper()

	// 组合使用键盘和鼠标
	result := helper.
		// 键盘操作
		Tab().Tab().Tab().
		Type("Test").
		Process().
		Wait(50 * time.Millisecond).

		// 鼠标操作（假设知道坐标）
		Click(20, 10).
		Process().
		Wait(50 * time.Millisecond).

		// 再次键盘操作
		Tab().Tab().
		Enter().
		Process().
		AssertRender("Count: 1").
		AssertRender("Test").
		Result()

	if !result.OK() {
		t.Errorf("Combination test failed: %v", result.Error())
	}
}

// testStateValidation 状态验证测试
func testStateValidation(t *testing.T) {
	testApp, err := ui.TestRun(Counter, ui.TestWithSize(40, 18))
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	helper := testApp.Helper()

	// 验证多个状态点
	states := []struct {
		name     string
		actions  func(*mock.TestHelper) *mock.TestHelper
		expected []string
	}{
		{
			name: "Initial state",
			actions: func(h *mock.TestHelper) *mock.TestHelper {
				return h.Process()
			},
			expected: []string{"Count: 0", "Hello, Guest"},
		},
		{
			name: "After increment",
			actions: func(h *mock.TestHelper) *mock.TestHelper {
				return h.Tab().Tab().Enter().Process()
			},
			expected: []string{"Count: 1"},
		},
		{
			name: "After name change",
			actions: func(h *mock.TestHelper) *mock.TestHelper {
				return h.Tab().Tab().Tab().Type("Bob").Process()
			},
			expected: []string{"Hello, Bob"},
		},
	}

	for _, state := range states {
		t.Run(state.name, func(t *testing.T) {
			result := state.actions(helper).Result()

			for _, exp := range state.expected {
				result = helper.AssertRender(exp).Result()
				if !result.OK() {
					t.Errorf("State '%s' missing expected '%s': %v",
						state.name, exp, result.Error())
				}
			}
		})
	}
}

// testPerformanceTest 性能测试
func testPerformanceTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	testApp, err := ui.TestRun(Counter, ui.TestWithSize(40, 18))
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	helper := testApp.Helper()

	// 性能测试：快速重复操作
	start := time.Now()

	// 执行 100 次增加操作
	for i := 0; i < 100; i++ {
		result := helper.
			Tab().Tab().
			Enter().
			Process().
			AssertRender(fmt.Sprintf("Count: %d", i+1)).
			Result()

		if !result.OK() {
			t.Errorf("Iteration %d failed: %v", i+1, result.Error())
			break
		}
	}

	elapsed := time.Since(start)
	t.Logf("100 iterations took: %v", elapsed)
	t.Logf("Average per iteration: %v", elapsed/time.Duration(100))
}
