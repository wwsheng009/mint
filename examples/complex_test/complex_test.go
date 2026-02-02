package main

import (
	"fmt"
	"os"
	"testing"
	"time"

	ui "github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/sandbox/mock"
)

// TestComplexAppWithSandbox 使用 sandbox 测试复杂应用
func TestComplexAppWithSandbox(t *testing.T) {
	os.Setenv("MINT_USE_FIBER", "true")

	sb := mock.New(100, 40)
	_, err := ui.RunTestWithSandbox(RootApp,
		ui.WithWidth(100),
		ui.WithHeight(40),
	)
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}

	// 等待初始渲染
	time.Sleep(100 * time.Millisecond)

	// 测试 Tab 切换
	t.Log("=== Testing Tab Navigation ===")
	testTabSwitch(t, sb)

	// 测试表单
	t.Log("=== Testing Form ===")
	testForm(t, sb)

	// 测试列表操作
	t.Log("=== Testing List ===")
	testList(t, sb)

	// 测试进度条
	t.Log("=== Testing Progress ===")
	testProgress(t, sb)

	// 打印最终状态
	t.Log("=== Test Complete ===")
}

func testTabSwitch(t *testing.T, sb *mock.MockSandbox) {
	tabs := []string{"form", "list", "modal", "progress"}
	for _, tab := range tabs {
		// 切换标签（通过修改全局状态）
		switchTab(tab)
		t.Logf("Switched to tab: %s", tab)

		// 注入事件触发重绘
		sb.InjectKey(' ')
		time.Sleep(50 * time.Millisecond)
	}
}

func testForm(t *testing.T, sb *mock.MockSandbox) {
	// 切换到表单标签
	switchTab("form")

	// 测试 checkbox 切换
	toggleCheckbox()
	t.Logf("Toggled checkbox: %v", getFormAgree())

	// 填写表单
	setFormName("Test User")
	setFormEmail("test@example.com")
	setFormAgree(true)

	sb.InjectKey(' ')
	time.Sleep(50 * time.Millisecond)
	t.Logf("Form filled: Name=%s, Email=%s, Agree=%v",
		getFormName(), getFormEmail(), getFormAgree())
}

func testList(t *testing.T, sb *mock.MockSandbox) {
	// 切换到列表标签
	switchTab("list")

	// 测试添加项目
	initialCount := getListCount()
	t.Logf("Initial item count: %d", initialCount)

	// 添加3个项目
	for i := 0; i < 3; i++ {
		addListItem(fmt.Sprintf("New Item %d", i+1))
	}

	newCount := getListCount()
	t.Logf("After adding 3 items: %d", newCount)

	if newCount != initialCount+3 {
		t.Errorf("Expected %d items, got %d", initialCount+3, newCount)
	}

	// 移除一个项目
	if getListCount() > 0 {
		removeLastListItem()
	}
	t.Logf("After removing 1 item: %d", getListCount())
}

func testProgress(t *testing.T, sb *mock.MockSandbox) {
	// 切换到进度条标签
	switchTab("progress")

	// 测试设置不同的进度值
	testValues := []int{0, 25, 50, 75, 100}
	for _, val := range testValues {
		setProgress(val)
		t.Logf("Progress set to: %d%%", val)
		time.Sleep(20 * time.Millisecond)
	}

	// 测试自动增加
	setProgress(0)
	for i := 0; i < 10; i++ {
		incrementProgress()
		t.Logf("Auto increment: %d%%", getProgress())
	}
}

// BenchmarkComplexApp 性能基准测试
func BenchmarkComplexApp(b *testing.B) {
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("TUI_DEBUG_UI", "false")

	for i := 0; i < b.N; i++ {
		// 每次迭代创建一个新的应用实例
		sb := mock.New(100, 40)
		_, err := ui.RunTestWithSandbox(RootApp,
			ui.WithWidth(100),
			ui.WithHeight(40),
		)
		if err != nil {
			b.Fatalf("Failed to create test app: %v", err)
		}

		// 模拟一些事件
		switchTab("form")
		sb.InjectKey(' ')

		switchTab("list")
		sb.InjectKey(' ')
	}
}

// ============================================================================
// State Accessors for Testing
// These functions access the global state from main.go
// ============================================================================

func switchTab(tab string) {
	currentTab = tab
}

func getFormName() string {
	return formName
}

func setFormName(name string) {
	formName = name
}

func getFormEmail() string {
	return formEmail
}

func setFormEmail(email string) {
	formEmail = email
}

func getFormAgree() bool {
	return formAgree
}

func setFormAgree(agree bool) {
	formAgree = agree
}

func toggleCheckbox() {
	formAgree = !formAgree
}

func getListCount() int {
	return len(listItems)
}

func addListItem(item string) {
	listItems = append(listItems, item)
}

func removeLastListItem() {
	if len(listItems) > 0 {
		listItems = listItems[:len(listItems)-1]
	}
}

func clearList() {
	listItems = []string{}
}

func getProgress() int {
	return progressValue
}

func setProgress(value int) {
	progressValue = value
}

func incrementProgress() {
	progressValue += 10
	if progressValue > 100 {
		progressValue = 0
	}
}
