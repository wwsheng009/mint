// ui/test_integration_test.go - UI测试集成示例
package ui

import (
	"fmt"
	"testing"

	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/sandbox"
)

func TestTestRun(t *testing.T) {
	app := "test app"

	testApp, err := TestRun(app, TestWithSize(80, 24))
	if err != nil {
		t.Fatalf("TestRun() error = %v", err)
	}
	defer testApp.Close()

	sb := testApp.Sandbox()
	if sb == nil {
		t.Fatal("Sandbox() returned nil")
	}

	width, height := sb.Size()
	if width != 80 || height != 24 {
		t.Errorf("Size() = %dx%d, want 80x24", width, height)
	}
}

func TestTestRunWithConfig(t *testing.T) {
	app := "test app"
	config := sandbox.DefaultConfig()
	config.Width = 100
	config.Height = 30

	testApp, err := TestRunWithConfig(app, config)
	if err != nil {
		t.Fatalf("TestRunWithConfig() error = %v", err)
	}
	defer testApp.Close()

	sb := testApp.Sandbox()
	width, height := sb.Size()
	if width != 100 || height != 30 {
		t.Errorf("Size() = %dx%d, want 100x30", width, height)
	}
}

func TestTestHelperChain(t *testing.T) {
	app := "test app"
	testApp, err := TestRun(app, TestWithSize(80, 24))
	if err != nil {
		t.Fatalf("TestRun() error = %v", err)
	}
	defer testApp.Close()

	helper := testApp.Helper()

	// 链式测试
	result := helper.
		Type("hello").
		Tab().
		Enter().
		Process().
		Result()

	if !result.OK() {
		t.Errorf("Chain produced errors: %v", result.Errors)
	}
}

func TestTestWithWidth(t *testing.T) {
	config := &testConfig{}
	TestWithWidth(100)(config)

	if config.width != 100 {
		t.Errorf("TestWithWidth() width = %v, want 100", config.width)
	}
}

func TestTestWithHeight(t *testing.T) {
	config := &testConfig{}
	TestWithHeight(30)(config)

	if config.height != 30 {
		t.Errorf("TestWithHeight() height = %v, want 30", config.height)
	}
}

func TestTestWithSize(t *testing.T) {
	config := &testConfig{}
	TestWithSize(120, 40)(config)

	if config.width != 120 || config.height != 40 {
		t.Errorf("TestWithSize() size = %dx%d, want 120x40", config.width, config.height)
	}
}

// TestRunTestWithSandbox_VerifyIntegration 验证 SandboxEventSource 集成
func TestRunTestWithSandbox_VerifyIntegration(t *testing.T) {
	// 创建一个简单的计数器组件
	count := 0

	counter := func() VNode {
		return VStack(
			NewTextBuilder(fmt.Sprintf("Count: %d", count)).Build(),
			ButtonBuilder("+").OnClick(func() {
				count++
			}).Build(),
		)
	}

	// 使用 RunTestWithSandbox 创建测试应用
	testApp, err := RunTestWithSandbox(counter,
		WithWidth(40),
		WithHeight(12),
	)
	if err != nil {
		t.Fatalf("RunTestWithSandbox() error = %v", err)
	}
	defer testApp.Close()

	// 验证 MockSandbox 可用
	sb := testApp.GetSandbox()
	if sb == nil {
		t.Fatal("GetSandbox() returned nil")
	}

	// 验证可以注入事件
	if err := sb.InjectSpecialKey(platform.KeyTab); err != nil {
		t.Errorf("InjectSpecialKey() error = %v", err)
	}

	// 等待事件处理
	// (实际测试中需要更长的等待和状态验证)
}

// TestSandboxSourceVsDirectInjection 对比两种注入方式
func TestSandboxSourceVsDirectInjection(t *testing.T) {
	simpleApp := func() VNode {
		return NewTextBuilder("Hello, Sandbox!").Build()
	}

	// 方式1: RunTest (直接注入到 Pump)
	t.Run("DirectInjection", func(t *testing.T) {
		testApp, err := RunTest(simpleApp, WithWidth(30), WithHeight(10))
		if err != nil {
			t.Fatalf("RunTest() error = %v", err)
		}
		defer testApp.Close()

		// 直接注入应该工作
		if err := testApp.InjectKey('q'); err != nil {
			t.Errorf("InjectKey() error = %v", err)
		}
	})

	// 方式2: RunTestWithSandbox (通过 MockSandbox)
	t.Run("SandboxInjection", func(t *testing.T) {
		testApp, err := RunTestWithSandbox(simpleApp, WithWidth(30), WithHeight(10))
		if err != nil {
			t.Fatalf("RunTestWithSandbox() error = %v", err)
		}
		defer testApp.Close()

		// 通过 Sandbox 注入应该工作
		sb := testApp.GetSandbox()
		if err := sb.InjectKey('q'); err != nil {
			t.Errorf("Sandbox.InjectKey() error = %v", err)
		}

		// 直接注入也应该工作（因为 TestableApp 提供了统一的接口）
		if err := testApp.InjectKey('q'); err != nil {
			t.Errorf("InjectKey() error = %v", err)
		}
	})
}
